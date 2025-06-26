package cmd

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

// New creates a new login command
func loginCmd(ac *ic.Context) *cobra.Command {
	o := &loginOptions{}
	c := cmd.NewSubCommand("login", o, ac).
		WithShortDesc("Login to Inventory Server").
		WithGroupID(groupAuth).
		Build()

	c.Flags().SortFlags = false

	return c
}

type loginOptions struct{}

func (o *loginOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *loginOptions) Complete(_ context.Context, _ *ic.Context) error   { return nil }
func (o *loginOptions) Validate(_ context.Context, _ *ic.Context) error   { return nil }
func (o *loginOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Login")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	ui.Success.Println("Logged in")

	return nil
}

func doLogin(ctx context.Context, ac *ic.Context) (*oidc.TokenSet, error) {
	var (
		tokenSet *oidc.TokenSet
		err      error
	)

	loginInput := authentication.LoginInput{
		Provider:    *ac.OIDCProvider,
		TokenCache:  ac.TokenCache,
		AuthOptions: authentication.AuthOptions{},
	}
	switch ac.OIDC.GrantType {
	case "authcode-browser":
		loginInput.AuthOptions.AuthCodeBrowser = &authcode.BrowserLoginInput{
			BindAddress:         ac.OIDC.AuthBindAddr,
			RedirectURLHostname: ac.OIDC.RedirectURLHostname,
		}
		if err := ui.Spin(ac.EC.Spinner, "Logging in", func(_ ui.Spinner) error {
			tokenSet, err = ac.Authenticator.Login(ctx, loginInput)
			return err
		}); err != nil {
			return nil, ac.EC.ErrorHandler.NewGeneralError(
				"Logging in",
				"See details for more information",
				err,
				0,
			)
		}
	case "authcode-keyboard":
		loginInput.AuthOptions.AuthCodeKeyboard = &authcode.KeyboardLoginInput{
			RedirectURI: ac.OIDC.RedirectURIAuthCodeKeyboard,
		}
		tokenSet, err = ac.Authenticator.Login(ctx, loginInput)
		if err != nil {
			return nil, ac.EC.ErrorHandler.NewGeneralError(
				"Logging in",
				"See details for more information",
				err,
				0,
			)
		}
	}

	if err := ac.SetupDefaultAPIClient(tokenSet.AccessToken); err != nil {
		return nil, ac.EC.ErrorHandler.NewGeneralError(
			"Setup API client in",
			"See details for more information",
			err,
			0,
		)
	}

	return tokenSet, nil
}
