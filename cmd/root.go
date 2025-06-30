package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	defaultConfigFilename = "ic"
	envPrefix             = "IC"
	oobRedirectURI        = "urn:ietf:wg:oauth:2.0:oob"

	groupBase      = "group-base"
	groupAuth      = "group-auth"
	groupCluster   = "group-cluster"
	groupComponent = "group-component"
	groupOther     = "group-other"
)

const (
	AppName   = "ic"
	ShortDesc = "Inventory CLI"
	LongDesc  = `ic is a tool to manage das Inventar`
)

func newRootCmd(ac *ic.Context) *cobra.Command {
	pf := pflag.NewFlagSet("", pflag.ContinueOnError)
	pf.StringVarP(&ac.APIServer, "api-server", "s", "https://api.k8s.netic.dk", "URL for the inventory server.")
	pf.StringVar(&ac.OIDC.IssuerURL, "oidc-issuer-url", "https://keycloak.netic.dk/auth/realms/mcs", "Issuer URL for the OIDC Provider")
	pf.StringVar(&ac.OIDC.ClientID, "oidc-client-id", "inventory-cli", "OIDC client ID")
	pf.StringVar(&ac.OIDC.GrantType, "oidc-grant-type", "authcode-browser", "OIDC authorization grant type. One of (authcode-browser|authcode-keyboard)")
	pf.StringVar(&ac.OIDC.RedirectURLHostname, "oidc-redirect-url-hostname", "localhost", "[authcode-browser] Hostname of the redirect URL")
	pf.StringVar(&ac.OIDC.AuthBindAddr, "oidc-auth-bind-addr", "localhost:18000", "[authcode-browser] Bind address and port for local server used for OIDC redirect")
	pf.StringVar(&ac.OIDC.RedirectURIAuthCodeKeyboard, "oidc-redirect-uri-authcode-keyboard", oobRedirectURI, "[authcode-keyboard] Redirect URI when using authcode keyboard")
	pf.StringVar(&ac.OIDC.TokenCacheDir, "oidc-token-cache-dir", getDefaultTokenCacheDir(), "Directory used to store cached tokens")
	pf.BoolVarP(&ac.Quiet, "quiet", "q", false, "Suppress informational UI output")

	c := cmd.NewRootCommand(ac.EC).
		WithInitFunc(func(_ *cobra.Command, _ []string) error {
			if ac.Quiet {
				var buf bytes.Buffer
				ui.SetDefaultOutput(&buf) // Mute informational UI output
			}
			ac.SetupDefaultAuthenticator()
			ac.SetupDefaultOIDCProvider()
			if err := ac.SetupDefaultTokenCache(); err != nil {
				return fmt.Errorf("settings up execution context: %w", err)
			}
			return nil
		}).
		WithPersistentFlags(pf).
		Build()

	c.AddGroup(
		&cobra.Group{
			ID:    groupBase,
			Title: "Basic Commands:",
		},
		&cobra.Group{
			ID:    groupAuth,
			Title: "Authentication Commands:",
		},
		&cobra.Group{
			ID:    groupOther,
			Title: "Other Commands:",
		},
	)

	c.AddCommand(
		loginCmd(ac),
		logoutCmd(ac),
		apiTokenCmd(ac),
		getCmd(ac),
		createCmd(ac),
		deleteCmd(ac),
		updateCmd(ac),
		filtersHelpCmd(ac),
	)

	return c
}

func getDefaultTokenCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "ic", "oidc-login")
	}
	return filepath.Join(cacheDir, "ic", "oidc-login")
}

// Execute runs the root command and returns the exit code
func Execute(version string) int {
	ec := cmd.NewExecutionContext(AppName, ShortDesc, version)
	ec.PFlags.NoInputEnabled = true
	ec.PFlags.ForceEnabled = true
	ec.PFlags.NoHeadersEnabled = true
	ac := ic.NewContext()
	ac.EC = ec
	ui.SetDefaultOutput(ac.EC.Stderr) // Redirect informational UI output (e.g., spinners) to stderr
	ec.LongDescription = LongDesc
	rootCmd := newRootCmd(ac)
	err := rootCmd.Execute()
	_ = ec.Spinner.Stop()
	if err != nil {
		ec.ErrorHandler.HandleError(err)
		return 1
	}
	return 0
}
