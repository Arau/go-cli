package textformat

import (
	"context"
	"fmt"
	"io"

	"github.com/dustin/go-humanize"
	"github.com/gosuri/uitable"

	"code.storageos.net/storageos/c2-cli/output"
	"code.storageos.net/storageos/c2-cli/pkg/capacity"
	"code.storageos.net/storageos/c2-cli/pkg/health"
)

// DescribeCluster encodes a cluster as JSON, writing the result to w.
func (d *Displayer) DescribeCluster(ctx context.Context, w io.Writer, c *output.Cluster) error {
	table, write := createTable(nil)

	capacity := fmt.Sprintf("%s (%d)", humanize.IBytes(c.Licence.ClusterCapacityBytes), c.Licence.ClusterCapacityBytes)

	table.AddRow("ID:", c.ID)
	table.AddRow("Licence:", "")
	table.AddRow("  expiration:", d.timeToHuman(c.Licence.ExpiresAt))
	table.AddRow("  capacity:", capacity)
	table.AddRow("  kind:", c.Licence.Kind)
	table.AddRow("  customer name:", c.Licence.CustomerName)
	table.AddRow("Version:", c.Version)
	table.AddRow("Created at:", d.timeToHuman(c.CreatedAt))
	table.AddRow("Updated at:", d.timeToHuman(c.UpdatedAt))
	table.AddRow("Telemetry:", disableToHuman(c.DisableTelemetry))
	table.AddRow("Crash Reporting:", disableToHuman(c.DisableCrashReporting))
	table.AddRow("Version Check:", disableToHuman(c.DisableVersionCheck))
	table.AddRow("Log Level:", c.LogLevel.String())
	table.AddRow("Log Format:", c.LogFormat.String())
	table.AddRow("Nodes:", "")
	for i, n := range c.Nodes {
		if i > 0 {
			table.AddRow("", "")
		}
		table.AddRow("  ID:", n.ID.String())
		table.AddRow("  Name:", n.Name)
		table.AddRow("  Health:", n.Health)
		table.AddRow("  Address:", n.IOAddr)
	}

	return write(w)
}

// DescribeNode prints all the detailed information about a node
func (d *Displayer) DescribeNode(ctx context.Context, w io.Writer, node *output.NodeDescription) error {
	return d.describeNode(ctx, w, node)
}

// DescribeListNodes prints all the detailed information about a list of nodes
func (d *Displayer) DescribeListNodes(ctx context.Context, w io.Writer, nodes []*output.NodeDescription) error {
	for i, node := range nodes {
		if i != 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}

		if err := d.describeNode(ctx, w, node); err != nil {
			return err
		}
	}

	return nil
}

func (d *Displayer) describeNode(ctx context.Context, w io.Writer, node *output.NodeDescription) error {
	table := uitable.New()
	table.MaxColWidth = 80
	table.Separator = "  "

	table.AddRow("ID", node.ID.String())
	table.AddRow("Name", node.Name)
	table.AddRow("Health", node.Health.String())
	// Addresses
	table.AddRow("Addresses:")
	table.AddRow("  Data Transfer address", node.IOAddr)
	table.AddRow("  Gossip address", node.GossipAddr)
	table.AddRow("  Supervisor address", node.SupervisorAddr)
	table.AddRow("  Clustering address", node.ClusteringAddr)

	table.AddRow("Labels", node.Labels.String())
	table.AddRow("Created at", d.timeToHuman(node.CreatedAt))
	table.AddRow("Updated at", d.timeToHuman(node.UpdatedAt))
	table.AddRow("Version", node.Version.String())

	capacityStats := "n/a"
	if node.Capacity != (capacity.Stats{}) {
		capacityStats = fmt.Sprintf(
			"%s/%s (%s in use)",
			humanize.IBytes(node.Capacity.Available),
			humanize.IBytes(node.Capacity.Total),
			humanize.IBytes(node.Capacity.Total-node.Capacity.Free),
		)
	}
	table.AddRow("Available capacity", capacityStats)

	// Volumes
	if len(node.HostedVolumes) > 0 {
		table.AddRow()
		table.AddRow("Local volume deployments:")
		table.AddRow("  DEPLOYMENT ID", "VOLUME", "NAMESPACE", "HEALTH", "TYPE", "SIZE")
		for _, vol := range node.HostedVolumes {
			table.AddRow(
				"  "+vol.LocalDeployment.ID,
				vol.Name,
				vol.NamespaceName,
				vol.LocalDeployment.Health,
				vol.LocalDeployment.Kind,
				humanize.IBytes(vol.SizeBytes),
			)
		}
	}

	_, err := fmt.Fprintln(w, table)
	return err
}

// DescribeVolume prints in the output writer a tabular representation, in a key
// value shape, of all details about a volume and its master and replicas.
func (d *Displayer) DescribeVolume(ctx context.Context, w io.Writer, volume *output.Volume) error {
	return d.describeVolume(ctx, w, volume)
}

// DescribeListVolumes writes a detailed, yet human-friendly table
// representation to w for each item in volumes.
func (d *Displayer) DescribeListVolumes(ctx context.Context, w io.Writer, volumes []*output.Volume) error {
	for i, v := range volumes {
		if i > 0 {
			_, err := w.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}

		err := d.describeVolume(ctx, w, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Displayer) describeVolume(ctx context.Context, w io.Writer, volume *output.Volume) error {
	table := uitable.New()
	table.Separator = "  "

	table.AddRow("ID", volume.ID.String())
	table.AddRow("Name", volume.Name)
	table.AddRow("Description", volume.Description)

	var attachedOnString string
	if volume.AttachedOn != "" {
		attachedOnString = fmt.Sprintf("%s (%s)", volume.AttachedOnName, volume.AttachedOn)
	}

	table.AddRow("AttachedOn", attachedOnString)
	table.AddRow("Namespace", fmt.Sprintf("%s (%s)", volume.NamespaceName, volume.Namespace))
	table.AddRow("Labels", volume.Labels.String())
	table.AddRow("Filesystem", volume.Filesystem.String())
	table.AddRow("Size", fmt.Sprintf("%v (%v bytes)", humanize.IBytes(volume.SizeBytes), volume.SizeBytes))

	table.AddRow("Version", volume.Version)
	table.AddRow("Created at", d.timeToHuman(volume.CreatedAt))
	table.AddRow("Updated at", d.timeToHuman(volume.UpdatedAt))

	table.AddRow("", "")
	table.AddRow("Master:")
	d.describeMaster(table, volume.Master)

	if len(volume.Replicas) > 0 {
		table.AddRow("", "")
		table.AddRow("Replicas:")
		for i, rep := range volume.Replicas {
			if i > 0 {
				table.AddRow("", "")
			}
			d.describeReplica(table, volume.SizeBytes, rep)
		}
	}

	_, err := fmt.Fprintln(w, table)
	return err
}

func (d *Displayer) describeMaster(table *uitable.Table, master *output.Deployment) {
	table.AddRow("  ID", master.ID.String())
	table.AddRow("  Node", fmt.Sprintf("%s (%s)", master.NodeName, master.Node))
	table.AddRow("  Health", master.Health.String())
}

func (d *Displayer) describeReplica(table *uitable.Table, size uint64, replica *output.Deployment) {
	table.AddRow("  ID", replica.ID.String())
	table.AddRow("  Node", fmt.Sprintf("%s (%s)", replica.NodeName, replica.Node))
	table.AddRow("  Health", replica.Health.String())
	table.AddRow("  Promotable", replica.Promotable)
	if replica.Health == health.ReplicaSyncing {
		d.describeSyncProgress(table, size, replica.SyncProgress)
	}
}

func (d *Displayer) describeSyncProgress(table *uitable.Table, size uint64, progress *output.SyncProgress) {
	if progress == nil {
		table.AddRow("  Sync Progress", "n/a")
		return
	}

	barStr, err := syncProgressBarString(
		progress.BytesRemaining,
		size,
		progress.EstimatedSecondsRemaining,
	)
	if err != nil {
		recap := fmt.Sprintf("%d/%d", size-progress.BytesRemaining, size)
		table.AddRow("  Sync Progress", recap)
	} else {
		table.AddRow("  Sync Progress", barStr)
	}
}

// DescribePolicyGroup prints in the output writer a tabular representation, in
// a key value shape, of all details about a policy group.
func (d *Displayer) DescribePolicyGroup(ctx context.Context, w io.Writer, group *output.PolicyGroup) error {
	return d.describePolicyGroup(ctx, w, group)
}

// DescribeListPolicyGroups writes a detailed, yet human-friendly table
// representation to w for each item in groups.
func (d *Displayer) DescribeListPolicyGroups(ctx context.Context, w io.Writer, groups []*output.PolicyGroup) error {
	for i, g := range groups {
		if i > 0 {
			_, err := w.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}

		err := d.describePolicyGroup(ctx, w, g)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Displayer) describePolicyGroup(ctx context.Context, w io.Writer, group *output.PolicyGroup) error {
	table := uitable.New()
	table.Separator = "  "

	table.AddRow("ID", group.ID.String())
	table.AddRow("Name", group.Name)

	// Specs
	if len(group.Specs) == 0 {
		table.AddRow("Specs:", "[]")
	} else {
		table.AddRow("Specs:", "")
		for _, s := range group.Specs {
			var rw string
			if s.ReadOnly {
				rw = "read"
			} else {
				rw = "write"
			}
			specString := fmt.Sprintf("%5s %6s on %s", rw, s.ResourceType, s.NamespaceName)

			table.AddRow("", specString)
		}
	}

	// Members
	if len(group.Users) == 0 {
		table.AddRow("Members:", "[]")
	} else {
		table.AddRow("Members:", "")
		for _, u := range group.Users {
			table.AddRow("", u.Username)
		}
	}

	table.AddRow("Created at", d.timeToHuman(group.CreatedAt))
	table.AddRow("Updated at", d.timeToHuman(group.UpdatedAt))
	table.AddRow("Version", group.Version)

	_, err := fmt.Fprintln(w, table)
	return err
}
