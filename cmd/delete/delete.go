package delete

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"code.storageos.net/storageos/c2-cli/apiclient"
	"code.storageos.net/storageos/c2-cli/namespace"
	"code.storageos.net/storageos/c2-cli/output"
	"code.storageos.net/storageos/c2-cli/output/jsonformat"
	"code.storageos.net/storageos/c2-cli/output/textformat"
	"code.storageos.net/storageos/c2-cli/output/yamlformat"
	"code.storageos.net/storageos/c2-cli/pkg/id"
	"code.storageos.net/storageos/c2-cli/volume"
)

// ConfigProvider specifies the configuration settings which commands require
// access to.
type ConfigProvider interface {
	CommandTimeout() (time.Duration, error)
	UseIDs() (bool, error)
	Namespace() (string, error)
	OutputFormat() (output.Format, error)
}

// Client defines the functionality required by the CLI application to
// reasonably implement the "delete" verb commands.
type Client interface {
	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)
	GetVolumeByName(ctx context.Context, namespaceID id.Namespace, name string) (*volume.Resource, error)

	DeleteVolume(ctx context.Context, namespaceID id.Namespace, volumeID id.Volume, params *apiclient.DeleteVolumeRequestParams) error
	DeleteNamespace(ctx context.Context, uid id.Namespace, params *apiclient.DeleteNamespaceRequestParams) error
}

// Displayer defines the functionality required by the CLI application to
// display the results gathered by the "delete" verb commands.
type Displayer interface {
	DeleteVolume(ctx context.Context, w io.Writer, confirmation output.VolumeDeletion) error
	DeleteVolumeAsync(ctx context.Context, w io.Writer) error
	DeleteNamespace(ctx context.Context, w io.Writer, confirmation output.NamespaceDeletion) error
}

// NewCommand configures the set of commands which are grouped by the "delete" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources in the cluster",
	}

	command.AddCommand(
		newVolume(os.Stdout, client, config),
		newNamespace(os.Stdout, client, config),
	)

	return command
}

// SelectDisplayer returns the right command displayer specified in the
// config provider.
func SelectDisplayer(cp ConfigProvider) Displayer {
	out, err := cp.OutputFormat()
	if err != nil {
		return textformat.NewDisplayer(textformat.NewTimeFormatter())
	}

	switch out {
	case output.JSON:
		return jsonformat.NewDisplayer(jsonformat.DefaultEncodingIndent)
	case output.YAML:
		return yamlformat.NewDisplayer("")
	case output.Text:
		fallthrough
	default:
		return textformat.NewDisplayer(textformat.NewTimeFormatter())
	}
}
