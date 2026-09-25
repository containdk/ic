package cmd

import (
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

// New creates a new update command
func updateCmd(ac *ic.Context) *cobra.Command {
	o := &cmd.NoopRunner[*ic.Context]{}
	c := cmd.NewSubCommand("update", o, ac).
		WithShortDesc("Update a resource").
		WithExample(updateCmdExample()).
		WithGroupID(cmd.GroupBase).
		WithNoArgs().
		Build()
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}

	c.AddCommand(
		updateClusterCmd(ac),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: groupClusterTitle,
		},
	)
	return c
}

func updateCmdExample() string {
	b := strings.Builder{}

	b.WriteString("  # Update a cluster, enabling Application Operations\n")
	b.WriteString("  ic update cluster mycluster.my-provider --has-ao\n\n")

	b.WriteString("  # Update a cluster, settings a new description and resilience zone\n")
	b.WriteString("  ic update cluster mycluster.my-provider\n")
	b.WriteString("	--description \"new description\"\n")
	b.WriteString("	--resilience-zone new-resilience-zone\n")
	b.WriteString("\n")

	return b.String()
}
