package cmd

import (
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

// New creates a new get command
func getCmd(ac *ic.Context) *cobra.Command {
	o := &cmd.NoopRunner[*ic.Context]{}
	c := cmd.NewSubCommand("get", o, ac).
		WithShortDesc("Add one or many resources").
		WithExample(getCmdExample()).
		WithGroupID(cmd.GroupBase).
		WithNoArgs().
		Build()
	c.RunE = func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	}

	c.AddCommand(
		getClustersCmd(ac),
		getClusterCmd(ac),
		getClusterNodesCmd(ac),
		getClusterNodeCmd(ac),
		getClusterKubeconfigCmd(ac),
		getClusterVulnerabilitiesCmd(ac),
		getPartitionsCmd(ac),
		getRegionsCmd(ac),
		getResilienceZonesCmd(ac),
		getComponentsCmd(ac),
		getComponentCmd(ac),
	)

	c.AddGroup(
		&cobra.Group{
			ID:    groupCluster,
			Title: "Cluster Commands:",
		},
		&cobra.Group{
			ID:    groupComponent,
			Title: "Component Commands:",
		},
		&cobra.Group{
			ID:    groupOther,
			Title: "Other Commands:",
		},
	)
	return c
}

func getCmdExample() string {
	b := strings.Builder{}

	b.WriteString("  # Get a list of clusters\n")
	b.WriteString("  ic get clusters\n\n")

	b.WriteString("  # Get information about a single cluster\n")
	b.WriteString("  ic get cluster mycluster.myprovider\n\n")

	b.WriteString("  # Get information about a single cluster in json format\n")
	b.WriteString("  ic -o json get cluster mycluster.myprovider\n")
	b.WriteString("\n")

	return b.String()
}
