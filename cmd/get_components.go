package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/component"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

func getComponentsCmd(ac *ic.Context) *cobra.Command {
	o := &getComponentsOptions{}
	c := cmd.NewSubCommand("components", o, ac).
		WithShortDesc("Get list of components").
		WithGroupID(groupComponent).
		Build()

	return c
}

type getComponentsOptions struct{}

func (o *getComponentsOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *getComponentsOptions) Complete(_ context.Context, _ *ic.Context) error   { return nil }
func (o *getComponentsOptions) Validate(_ context.Context, _ *ic.Context) error   { return nil }

func (o *getComponentsOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Components")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *component.ListComponentResults
	if err := ui.Spin(ac.EC.Spinner, "Getting components", func(_ ui.Spinner) error {
		in := component.ListComponentsInput{
			Logger:    logger,
			APIClient: ac.APIClient,
		}
		result, err = component.ListComponents(ctx, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Listing components",
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

	r := component.NewComponentsRenderer(result.ComponentListResponse, result.JSONResponse, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
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
