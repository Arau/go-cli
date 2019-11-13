package get

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"code.storageos.net/storageos/c2-cli/output/jsonformat"
	"code.storageos.net/storageos/c2-cli/pkg/id"
	"code.storageos.net/storageos/c2-cli/volume"
)

type volumeCommand struct {
	config  ConfigProvider
	client  GetClient
	display GetDisplayer

	namespaceID string

	writer io.Writer
}

func (c *volumeCommand) run(cmd *cobra.Command, args []string) error {
	timeout, err := c.config.DialTimeout()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch len(args) {
	case 1:
		if c.namespaceID != "" {
			return c.getVolume(ctx, cmd, args)
		}
		fallthrough
	default:
		return c.listVolumes(ctx, cmd, args)
	}
}

func (c *volumeCommand) getVolume(ctx context.Context, _ *cobra.Command, args []string) error {
	uid := id.Volume(args[0])

	volume, err := c.client.GetVolume(
		ctx,
		id.Namespace(c.namespaceID),
		uid,
	)
	if err != nil {
		return err
	}

	return c.display.GetVolume(c.writer, volume)
}

func (c *volumeCommand) listVolumes(ctx context.Context, _ *cobra.Command, args []string) error {
	var volumes []*volume.Resource
	var err error

	uids := make([]id.Volume, len(args))
	for i, a := range args {
		uids[i] = id.Volume(a)
	}

	if c.namespaceID != "" {
		volumes, err = c.client.GetNamespaceVolumes(
			ctx,
			id.Namespace(c.namespaceID),
			uids...,
		)
	} else {
		volumes, err = c.client.GetAllVolumes(ctx)
	}

	if err != nil {
		return err
	}

	return c.display.GetVolumeList(c.writer, volumes)
}

func newVolume(w io.Writer, client GetClient, config ConfigProvider) *cobra.Command {
	c := &volumeCommand{
		config: config,
		client: client,
		display: jsonformat.NewDisplayer(
			jsonformat.DefaultEncodingIndent,
		),

		writer: w,
	}

	cobraCommand := &cobra.Command{
		Aliases: []string{"volumes"},
		Use:     "volume [volume ids...]",
		Short:   "volume retrieves basic information about StorageOS volumes",
		Example: `
$ storageos get volume banana
`,

		RunE: c.run,
	}

	cobraCommand.Flags().StringVarP(&c.namespaceID, "namespace", "n", "", "the id of the namespace to retrieve the volume resources from. if not set all namespaces are included")

	return cobraCommand
}
