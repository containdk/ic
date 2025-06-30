package cmd

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/component"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

func getComponentCmd(ac *ic.Context) *cobra.Command {
	o := &getComponentOptions{}
	c := cmd.NewSubCommand("component", o, ac).
		WithShortDesc("Get a component").
		WithGroupID(groupComponent).
		WithExactArgs(2).
		Build()
	c.Use = "component NAMESPACE-NAME COMPONENT-NAME"

	return c
}

type getComponentOptions struct {
	namespace string
	component string
}

func (o *getComponentOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *getComponentOptions) Complete(_ context.Context, ac *ic.Context) error {
	o.namespace = ac.EC.CommandArgs[0]
	o.component = ac.EC.CommandArgs[1]
	return nil
}

func (o *getComponentOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getComponentOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Components")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *component.GetComponentResult
	spinnerText := fmt.Sprintf("Getting component %q/%q", o.namespace, o.component)
	if err := ui.Spin(ac.EC.Spinner, spinnerText, func(_ ui.Spinner) error {
		in := component.GetComponentInput{
			Logger:    logger,
			APIClient: ac.APIClient,
		}
		result, err = component.GetComponent(ctx, o.namespace, o.component, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Getting component",
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

	r := component.NewComponentRenderer(result.ComponentResponse, result.JSONResponse, ac.EC.Stdout)
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
