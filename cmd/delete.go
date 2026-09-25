package cmd

import (
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

// New creates a new delete command
func deleteCmd(ac *ic.Context) *cobra.Command {
	o := &cmd.NoopRunner[*ic.Context]{}
	c := cmd.NewSubCommand("delete", o, ac).
		WithShortDesc("Delete a resource").
		WithExample(deleteCmdExample()).
		WithGroupID(cmd.GroupBase).
		WithNoArgs().
		Build()
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}

	c.AddCommand(
		deleteClusterCmd(ac),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: groupClusterTitle,
		},
	)
	return c
}

func deleteCmdExample() string {
	b := strings.Builder{}

	b.WriteString("  # Delete a cluster\n")
	b.WriteString("  ic delete cluster my-cluster.my-provider\n\n")

	return b.String()
}
