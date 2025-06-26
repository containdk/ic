package cmd

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

func deleteClusterCmd(ac *ic.Context) *cobra.Command {
	o := &deleteClusterOptions{}
	c := cmd.NewSubCommand("cluster", o, ac).
		WithShortDesc("Delete a cluster").
		WithGroupID(groupCluster).
		WithExactArgs(1).
		Build()
	c.Use = "cluster CLUSTER-ID" //nolint:goconst

	return c
}

type deleteClusterOptions struct {
	clusterID string
}

func (o *deleteClusterOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *deleteClusterOptions) Complete(_ context.Context, ac *ic.Context) error {
	o.clusterID = ac.EC.CommandArgs[0]
	return nil
}

func (o *deleteClusterOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }
func (o *deleteClusterOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Clusters")
	ac.Authenticator.SetLogger(logger)

	if !ac.EC.PFlags.Force {
		if yes := ui.Confirm(fmt.Sprintf("Delete %q", o.clusterID)); !yes {
			ui.Info.Println("User aborted")
			return nil
		}
	}

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *cluster.DeleteClusterResult
	spinnerText := fmt.Sprintf("Deleting cluster %s", o.clusterID)
	if err := ui.Spin(ac.EC.Spinner, spinnerText, func(s ui.Spinner) error {
		in := cluster.DeleteClusterInput{
			Logger:    logger,
			APIClient: ac.APIClient,
		}
		result, err = cluster.DeleteCluster(ctx, o.clusterID, in)
		if err == nil {
			ui.UpdateSpinnerText(s, "Cluster deleted")
		}
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Deleting cluster",
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

	if result.Problem != nil {
		return &errors.ProblemError{
			Title:   "deleting cluster",
			Problem: result.Problem,
		}
	}

	return nil
}
