package get

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"code.storageos.net/storageos/c2-cli/cluster"
	"code.storageos.net/storageos/c2-cli/namespace"
	"code.storageos.net/storageos/c2-cli/node"
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
// reasonably implement the "get" verb commands.
type Client interface {
	GetCluster(ctx context.Context) (*cluster.Resource, error)

	GetNode(ctx context.Context, uid id.Node) (*node.Resource, error)
	GetNodeByName(ctx context.Context, name string) (*node.Resource, error)
	GetListNodes(ctx context.Context, uids ...id.Node) ([]*node.Resource, error)
	GetListNodesByName(ctx context.Context, names ...string) ([]*node.Resource, error)

	GetVolume(ctx context.Context, namespaceID id.Namespace, uid id.Volume) (*volume.Resource, error)
	GetVolumeByName(ctx context.Context, namespaceID id.Namespace, name string) (*volume.Resource, error)
	GetAllVolumes(ctx context.Context) ([]*volume.Resource, error)
	GetNamespaceVolumes(ctx context.Context, namespaceID id.Namespace, uids ...id.Volume) ([]*volume.Resource, error)
	GetNamespaceVolumesByName(ctx context.Context, namespaceID id.Namespace, names ...string) ([]*volume.Resource, error)

	GetNamespace(ctx context.Context, uid id.Namespace) (*namespace.Resource, error)
	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)
	GetListNamespaces(ctx context.Context, uids ...id.Namespace) ([]*namespace.Resource, error)
	GetListNamespacesByName(ctx context.Context, name ...string) ([]*namespace.Resource, error)
	GetAllNamespaces(ctx context.Context) ([]*namespace.Resource, error)
}

// Displayer defines the functionality required by the CLI application to
// display the results gathered by the "get" verb commands.
type Displayer interface {
	GetCluster(ctx context.Context, w io.Writer, resource *cluster.Resource) error
	GetNode(ctx context.Context, w io.Writer, resource *node.Resource) error
	GetListNodes(ctx context.Context, w io.Writer, resources []*node.Resource) error
	GetNamespace(ctx context.Context, w io.Writer, resource *namespace.Resource) error
	GetListNamespaces(ctx context.Context, w io.Writer, resources []*namespace.Resource) error
	GetVolume(ctx context.Context, w io.Writer, volume *output.Volume) error
	GetListVolumes(ctx context.Context, w io.Writer, volumes []*output.Volume) error
}

// NewCommand configures the set of commands which are grouped by the "get" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "get",
		Short: "fetch basic details for resources",
	}

	command.AddCommand(
		newCluster(os.Stdout, client, config),
		newNode(os.Stdout, client, config),
		newNamespace(os.Stdout, client, config),
		newVolume(os.Stdout, client, config),
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
