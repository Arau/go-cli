package get

import (
	"io"

	"github.com/spf13/cobra"

	"code.storageos.net/storageos/c2-cli/pkg/id"
)

type nodeCommand struct {
	client  GetClient
	display GetDisplayer

	writer io.Writer
}

func (c *nodeCommand) run(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 1:
		return c.getNode(cmd, args)
	default:
		return c.listNodes(cmd, args)
	}
}

func (c *nodeCommand) getNode(_ *cobra.Command, args []string) error {
	uid := id.Node(args[0])

	node, err := c.client.GetNode(uid)
	if err != nil {
		return err
	}

	return c.display.WriteGetNode(c.writer, node)
}

func (c *nodeCommand) listNodes(_ *cobra.Command, args []string) error {
	uids := make([]id.Node, len(args))
	for i, a := range args {
		uids[i] = id.Node(a)
	}

	nodes, err := c.client.GetListNodes(uids...)
	if err != nil {
		return err
	}

	return c.display.WriteGetNodeList(c.writer, nodes)
}

func newNode(w io.Writer, client GetClient, display GetDisplayer) *cobra.Command {
	c := &nodeCommand{
		client:  client,
		display: display,

		writer: w,
	}
	cobraCommand := &cobra.Command{
		Aliases: []string{"nodes"},
		Use:     "node [node ids...]",
		Short:   "node retrieves basic information about StorageOS nodes",
		Example: `
$ storageos get node banana
`,

		RunE: c.run,
	}

	return cobraCommand
}
