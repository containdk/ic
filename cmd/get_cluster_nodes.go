package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/neticdk/go-common/pkg/qsparser"
	"github.com/spf13/cobra"
)

const getClusterNodesLongDesc = `Get list of nodes for a nodes.

Supported field names for filters:

name, role, criName, criVersion, controlPlane, topologyRegion,
topologyZone, memoryAllocatableBytes, cpuAllocatableMillis,
memoryCapacityBytes, cpuCapacityMillis
`

const getClusterNodesExample = `
# get nodes for my-cluster.my-provider
ic get cluster-nodes --cluster-name my-cluster.my-provider

use: 'ic help filters' for more information on using filters`

func getClusterNodesCmd(ac *ic.Context) *cobra.Command {
	o := &getClusterNodesOptions{}
	c := cmd.NewSubCommand("cluster-nodes", o, ac).
		WithShortDesc("Get list of nodes in a cluster").
		WithLongDesc(getClusterNodesLongDesc).
		WithExample(getClusterNodesExample).
		WithGroupID(groupCluster).
		Build()
	c.Aliases = []string{"nodes"}

	return c
}

type getClusterNodesOptions struct {
	clusterName string
	// A filter has the form: fieldName operator value (e.g. name=Peter)
	//
	// Supported operators:
	// == (or =) - equals
	// != (or !) - not equals
	// >         - greater than
	// <         - less than
	// >=        - greater than or equals
	// <=        - less than or equals
	// =~ (or ~) - matches (case insensitive regular expression)
	// !~        - does not match (case insensitive expression)
	Filters []string
}

func (o *getClusterNodesOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.StringVar(&o.clusterName, "cluster-name", "", "The name of the cluster")
	f.StringArrayVar(&o.Filters, "filter", []string{}, "Filter output based on conditions")

	c.MarkFlagRequired("cluster-name") //nolint:errcheck

	return nil
}

func (o *getClusterNodesOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *getClusterNodesOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getClusterNodesOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("ClusterNodes")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	searchFields := make(map[string]*qsparser.SearchField)
	for _, f := range o.Filters {
		out, err := parseFilter(f)
		if err != nil {
			return err
		}
		searchFields[out.FieldName] = out.SearchField
	}

	var result *cluster.ListClusterNodesResults

	if err := ui.Spin(ac.EC.Spinner, "Getting cluster nodes", func(_ ui.Spinner) error {
		in := cluster.ListClusterNodesInput{
			Logger:      logger,
			APIClient:   ac.APIClient,
			PerPage:     PerPage,
			Filters:     searchFields,
			ClusterName: o.clusterName,
		}
		result, err = cluster.ListClusterNodes(ctx, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Listing cluster nodes",
			"See details for more information",
			err,
			0,
		)
	}
	if result.Problem != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			*result.Problem.Title,
			*result.Problem.Detail,
			nil,
			0,
		)
	}

	r := cluster.NewClusterNodesRenderer(result.ClusterNodeListResponse, result.JSONResponse, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
	if err := r.Render(ac.EC.PFlags.OutputFormat); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Failed to render output",
			"See details for more information",
			err,
			0,
		)
	}

	return nil
}
