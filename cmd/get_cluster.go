package cmd

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

// New creates a new "get cluster" command
func getClusterCmd(ac *ic.Context) *cobra.Command {
	o := &getClusterOptions{}
	c := cmd.NewSubCommand("cluster", o, ac).
		WithShortDesc("Get a cluster").
		WithGroupID(groupCluster).
		WithExactArgs(1).
		Build()
	c.Use = "cluster CLUSTER-ID" //nolint:goconst
	return c
}

type getClusterOptions struct {
	clusterID string
}

func (o *getClusterOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *getClusterOptions) Complete(_ context.Context, ac *ic.Context) error {
	o.clusterID = ac.EC.CommandArgs[0]
	return nil
}

func (o *getClusterOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getClusterOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Clusters")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *cluster.GetClusterResult
	spinnerText := fmt.Sprintf("Getting cluster %q", o.clusterID)
	if err := ui.Spin(ac.EC.Spinner, spinnerText, func(_ ui.Spinner) error {
		in := cluster.GetClusterInput{
			Logger:    logger,
			APIClient: ac.APIClient,
		}
		result, err = cluster.GetCluster(ctx, o.clusterID, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Getting cluster",
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

	r := cluster.NewClusterRenderer(result.ClusterResponse, result.JSONResponse, ac.EC.Stdout)
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
