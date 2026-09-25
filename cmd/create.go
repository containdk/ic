package cmd

import (
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

// New creates a new create command
func createCmd(ac *ic.Context) *cobra.Command {
	o := &cmd.NoopRunner[*ic.Context]{}
	c := cmd.NewSubCommand("create", o, ac).
		WithShortDesc("Create a resource").
		WithExample(createCmdExample()).
		WithGroupID(cmd.GroupBase).
		WithNoArgs().
		Build()
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}

	c.AddCommand(
		createClusterCmd(ac),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: groupClusterTitle,
		},
	)
	return c
}

func createCmdExample() string {
	b := strings.Builder{}

	b.WriteString("  # Create a new Netic on-prem cluster with default SLA (TO, TM)\n")
	b.WriteString("  ic create cluster\n")
	b.WriteString("	--name my-cluster\n")
	b.WriteString("	--provider my-provider\n")
	b.WriteString("	--environment myenv\n")
	b.WriteString("	--partition netic\n")
	b.WriteString("	--region dk-north\n")
	b.WriteString("	--subscription 123456\n")
	b.WriteString("	--infrastructure-provider netic\n\n")

	b.WriteString("  # Create a new Azure cluster with Application Operations\n")
	b.WriteString("  ic create cluster\n")
	b.WriteString("	--name my-azure-cluster\n")
	b.WriteString("	--provider my-azure-provider\n")
	b.WriteString("	--environment myenv\n")
	b.WriteString("	--partition azure\n")
	b.WriteString("	--region swedencentral\n")
	b.WriteString("	--subscription 654321\n")
	b.WriteString("	--infrastructure-provider azure\n")
	b.WriteString("	--has-application-operations\n")
	b.WriteString("\n")

	return b.String()
}
