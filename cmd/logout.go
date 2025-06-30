package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

func logoutCmd(ac *ic.Context) *cobra.Command {
	o := &logoutOptions{}
	c := cmd.NewSubCommand("logout", o, ac).
		WithShortDesc("Log out of Inventory Server").
		WithGroupID(groupAuth).
		Build()

	return c
}

type logoutOptions struct{}

func (o *logoutOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.SortFlags = false
	return nil
}
func (o *logoutOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *logoutOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *logoutOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Logout")
	ac.Authenticator.SetLogger(logger)

	logoutInput := authentication.LogoutInput{
		Provider:   *ac.OIDCProvider,
		TokenCache: ac.TokenCache,
	}

	if err := ui.Spin(ac.EC.Spinner, "Logging out", func(_ ui.Spinner) error {
		return ac.Authenticator.Logout(ctx, logoutInput)
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Logging out",
			"See details for more information",
			err,
			0,
		)
	}

	ui.Success.Println("Logged out")

	return nil
}
