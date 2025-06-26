package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

const getClusterNodeExample = `
# get node information for my-cluster.my-provider
ic get cluster-node --cluster-name my-cluster.my-provider --node-name my-node`

func getClusterNodeCmd(ac *ic.Context) *cobra.Command {
	o := &getClusterNodeOptions{}
	c := cmd.NewSubCommand("cluster-node", o, ac).
		WithShortDesc("Get a cluster node").
		WithGroupID(groupCluster).
		WithExample(getClusterNodeExample).
		Build()
	c.Aliases = []string{"node"}

	return c
}

type getClusterNodeOptions struct {
	clusterName string
	nodeName    string
}

func (o *getClusterNodeOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.StringVar(&o.clusterName, "cluster-name", "", "The name of the cluster")
	f.StringVar(&o.nodeName, "node-name", "", "The name of the node")

	c.MarkFlagRequired("cluster-name") //nolint:errcheck
	c.MarkFlagRequired("node-name")    //nolint:errcheck

	return nil
}

func (o *getClusterNodeOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *getClusterNodeOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getClusterNodeOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("ClusterNode")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *cluster.GetClusterNodeResult
	if err := ui.Spin(ac.EC.Spinner, "Getting cluster node", func(_ ui.Spinner) error {
		in := cluster.GetClusterNodeInput{
			Logger:      logger,
			APIClient:   ac.APIClient,
			ClusterName: o.clusterName,
			NodeName:    o.nodeName,
		}
		result, err = cluster.GetClusterNode(ctx, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Getting cluster node",
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

	r := cluster.NewClusterNodeRenderer(result.ClusterNodeResponse, result.JSONResponse, ac.EC.Stdout)
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
