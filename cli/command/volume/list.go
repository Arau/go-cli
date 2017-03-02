package volume

import (
	"sort"

	"github.com/dnephin/cobra"
	"github.com/storageos/go-api/types"
	"github.com/storageos/go-cli/cli"
	"github.com/storageos/go-cli/cli/command"
	"github.com/storageos/go-cli/cli/command/formatter"
	"github.com/storageos/go-cli/cli/opts"
)

type byVolumeName []*types.Volume

func (r byVolumeName) Len() int      { return len(r) }
func (r byVolumeName) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r byVolumeName) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

type listOptions struct {
	quiet     bool
	format    string
	filter    opts.FilterOpt
	namespace string
}

func newListCommand(storageosCli *command.StorageOSCli) *cobra.Command {
	opts := listOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List volumes",
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(storageosCli, opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Only display volume names")
	flags.StringVar(&opts.format, "format", "", "Pretty-print volumes using a Go template")
	flags.VarP(&opts.filter, "filter", "f", "Provide filter values (e.g. 'dangling=true')")
	flags.StringVarP(&opts.namespace, "namespace", "n", "", "Namespace scope")

	return cmd
}

func runList(storageosCli *command.StorageOSCli, opts listOptions) error {
	client := storageosCli.Client()

	params := types.ListOptions{
		// LabelSelector: opts.filter.Value(),
		Namespace: opts.namespace,
	}

	// volumes, err := client.VolumeList(context.Background(), opts.filter.Value())
	volumes, err := client.VolumeList(params)
	if err != nil {
		return err
	}

	format := opts.format
	if len(format) == 0 {
		if len(storageosCli.ConfigFile().VolumesFormat) > 0 && !opts.quiet {
			format = storageosCli.ConfigFile().VolumesFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	sort.Sort(byVolumeName(volumes))

	volumeCtx := formatter.Context{
		Output: storageosCli.Out(),
		Format: formatter.NewVolumeFormat(format, opts.quiet),
	}
	return formatter.VolumeWrite(volumeCtx, volumes)
}
