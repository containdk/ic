package cmd

import (
	"context"
	"fmt"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

// New creates a new api-token command
func apiTokenCmd(ac *ic.Context) *cobra.Command {
	o := &apiTokenOptions{}
	c := cmd.NewSubCommand("api-token", o, ac).
		WithShortDesc("Get access token for the API").
		WithLongDesc("Get access token for the API. It uses the cached token if it is valid and performs login otherwise.").
		WithGroupID(groupAuth).
		Build()

	c.Flags().SortFlags = false
	return c
}

type apiTokenOptions struct{}

func (o *apiTokenOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *apiTokenOptions) Complete(_ context.Context, _ *ic.Context) error   { return nil }
func (o *apiTokenOptions) Validate(_ context.Context, _ *ic.Context) error   { return nil }
func (o *apiTokenOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Login")
	ac.Authenticator.SetLogger(logger)

	tokenSet, err := doLogin(ctx, ac)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	fmt.Println(tokenSet.AccessToken)
	return nil
}
