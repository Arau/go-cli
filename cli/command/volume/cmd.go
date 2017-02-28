package volume

import (
	"github.com/dnephin/cobra"

	"github.com/storageos/go-cli/cli"
	"github.com/storageos/go-cli/cli/command"
)

// NewVolumeCommand returns a cobra command for `volume` subcommands
func NewVolumeCommand(storageosCli *command.StorageOSCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes",
		Args:  cli.NoArgs,
		RunE:  storageosCli.ShowHelp,
	}
	cmd.AddCommand(
		// newCreateCommand(dockerCli),
		// newInspectCommand(dockerCli),
		// newPsCommand(dockerCli),
		newListCommand(storageosCli),
	// newRemoveCommand(dockerCli),
	// newScaleCommand(dockerCli),
	// newUpdateCommand(dockerCli),
	// newLogsCommand(dockerCli),
	)
	return cmd
}
