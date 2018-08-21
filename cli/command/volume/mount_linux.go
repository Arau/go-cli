// +build linux

package volume

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/dnephin/cobra"
	api "github.com/storageos/go-api"
	"github.com/storageos/go-cli/cli"
	"github.com/storageos/go-cli/cli/command"
	cliconfig "github.com/storageos/go-cli/cli/config"
	"github.com/storageos/go-cli/pkg/host"
	"github.com/storageos/go-cli/pkg/mount"
	"github.com/storageos/go-cli/pkg/system"
	"github.com/storageos/go-cli/pkg/validation"

	"github.com/storageos/go-api/types"

	log "github.com/sirupsen/logrus"
)

type mountOptions struct {
	ref        string
	timeout    time.Duration
	mountpoint string // mountpoint
}

func newMountCommand(storageosCli *command.StorageOSCli) *cobra.Command {
	var opt mountOptions

	cmd := &cobra.Command{
		Use:   "mount [OPTIONS] VOLUME MOUNTPOINT",
		Short: "Mount specified volume",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.ref = args[0]
			opt.mountpoint = args[1]
			return runMount(storageosCli, opt)
		},
	}

	flags := cmd.Flags()
	flags.DurationVarP(&opt.timeout, "timeout", "t", 20*time.Second, "Retryable mount timeout period in seconds")

	return cmd
}

func runMount(storageosCli *command.StorageOSCli, opt mountOptions) error {
	// Get current hostname.
	hostname, err := host.Get()
	if err != nil {
		hostname = "unknown"
	}

	client := storageosCli.Client()

	// Check whether we are on storageos node.
	node, err := client.Node(hostname)
	if err != nil {
		if err == api.ErrNoSuchNode {
			return fmt.Errorf("cannot mount volume: current host is not a registered storageos cluster node")
		}
		return fmt.Errorf("failed to check if this host is a storageos cluster node: %#v", err)
	}

	// Check whether device dir exists.
	_, err = system.Stat(node.DeviceDir)
	if err != nil {
		return fmt.Errorf("device root path %q not found, check whether StorageOS is running", cliconfig.DeviceRootPath)
	}

	// must be root
	if euid := syscall.Geteuid(); euid != 0 {
		return fmt.Errorf("volume mount requires root permission - try prefixing command with `sudo -E`")
	}

	namespace, name, err := validation.ParseRefWithDefault(opt.ref)
	if err != nil {
		return err
	}

	vol, err := client.Volume(namespace, name)
	if err != nil {
		return err
	}

	// checking readiness
	if err := isVolumeReady(vol); err != nil {
		return fmt.Errorf("cannot mount volume: %v", err)
	}

	err = client.VolumeMount(types.VolumeMountOptions{
		ID: vol.ID, Namespace: namespace,
		Client:     hostname,
		Mountpoint: opt.mountpoint,
		FsType:     vol.FSType,
	})
	if err != nil {
		return err
	}

	fst, err := mount.ParseFSType(vol.FSType)
	if err != nil {
		return err
	}

	err = retryableMount(vol, node.DeviceDir, opt, fst)
	if err != nil {
		log.WithFields(log.Fields{
			"namespace":  namespace,
			"volumeName": name,
			"error":      err,
		}).Error("error while mounting volume")
		// should unmount volume in the CP if we failed here
		newErr := client.VolumeUnmount(types.VolumeUnmountOptions{ID: vol.ID, Namespace: namespace})
		if newErr != nil {
			log.WithFields(log.Fields{
				"volumeId": vol.ID,
				"err":      newErr,
			}).Error("failed to unmount volume")
		}

		return fmt.Errorf("Failed to mount: %v", err)
	}

	fmt.Printf("volume %s mounted: %s\n", vol.Name, opt.mountpoint)

	return nil
}

func retryableMount(volume *types.Volume, deviceRootDir string, opts mountOptions, fsType mount.FSType) error {
	var deadlineExceeded bool
	driver := mount.New(deviceRootDir)
	// Limit the time which can be spent retrying
	timer := time.NewTimer(opts.timeout)
	reqTimeout := time.Second

RETRY:
	// Perform the mount
	ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
	defer cancel()

	go func() {
		for {
			select {
			case <-timer.C:
				deadlineExceeded = true
				cancel() // Cancel any ongoing mount attempt
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	err := driver.MountVolume(ctx, volume.ID, opts.mountpoint, fsType, volume.MkfsDoneAt.IsZero() && !volume.MkfsDone)
	if err != nil {
		log.WithFields(log.Fields{
			"volume_id":  volume.ID,
			"mountpoint": opts.mountpoint,
			"err":        err.Error(),
		}).Error(" failed to mount volume")

		// If this is a permanent error, stop retrying
		if mountErr, ok := err.(*mount.MountError); ok && mountErr.Fatal {
			return err
		}

		if !deadlineExceeded {
			fmt.Printf("error mounting, retrying")
			time.Sleep(250 * time.Millisecond)
			reqTimeout *= 2 // Increase the request time out
			goto RETRY
		}

		return err
	}

	if deadlineExceeded {
		return fmt.Errorf("exceeded mount retry duration")
	}

	return nil
}

// isVolumeReady - mount only unmounted and active volume
func isVolumeReady(vol *types.Volume) error {

	if vol.Status != "active" {
		return fmt.Errorf("can only mount active volumes, current status: '%s'", vol.Status)
	}

	if vol.Mounted {
		return errors.New("volume is mounted, unmount it before mounting it again")
	}

	return nil
}
