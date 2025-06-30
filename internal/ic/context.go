package ic

import (
	"fmt"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/reader"
	"github.com/neticdk-k8s/ic/internal/tokencache"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication/authcode"
	"github.com/neticdk/go-common/pkg/cli/cmd"
)

// OIDCConfig holds flag values for OIDC settings
type OIDCConfig struct {
	IssuerURL                   string
	ClientID                    string
	GrantType                   string
	RedirectURLHostname         string
	RedirectURIAuthCodeKeyboard string
	AuthBindAddr                string
	TokenCacheDir               string
}

type Context struct {
	EC *cmd.ExecutionContext

	// APIServer is the inventory api server endpoint
	APIServer string

	// APIClient is an inventory server api client
	APIClient apiclient.ClientWithResponsesInterface

	// OIDC is the OIDC settings
	OIDC OIDCConfig

	// OIDC is the OIDC provider settings
	OIDCProvider *oidc.Provider

	// Authenticator is the Authenticator
	Authenticator authentication.Authenticator

	// TokenCache is the token cache
	TokenCache tokencache.Cache

	// Quiet is a flag indicating whether to suppress CLI-output
	Quiet bool
}

func NewContext() *Context {
	return &Context{}
}

// SetupDefaultAPIClient sets up ec.APIClient from flags if it's not already set
func (ac *Context) SetupDefaultAPIClient(token string) (err error) {
	if ac.APIClient != nil {
		return
	}

	provider := apiclient.NewBearerTokenProvider(token)
	ac.APIClient, err = apiclient.NewClientWithResponses(
		ac.APIServer,
		apiclient.WithRequestEditorFn(provider.WithAuthHeader))
	return
}

// SetupDefaultAuthenticator sets up ec.Authenticator from flags if it's not already set
// It should be called from rootCmd.PersistentPreRunE
func (ac *Context) SetupDefaultAuthenticator() {
	if ac.Authenticator != nil {
		return
	}

	authn := authentication.NewAuthentication(
		ac.EC.Logger,
		nil,
		&authcode.Browser{Logger: ac.EC.Logger},
		&authcode.Keyboard{Reader: reader.NewReader(), Logger: ac.EC.Logger})
	ac.Authenticator = authentication.NewAuthenticator(ac.EC.Logger, authn)
}

// SetupDefaultOIDCProvider sets up ec.OIDCProvider from flags if it's not already set
// It should be called from rootCmd.PersistentPreRunE
func (ac *Context) SetupDefaultOIDCProvider() {
	if ac.OIDCProvider != nil {
		return
	}
	ac.OIDCProvider = &oidc.Provider{
		IssuerURL:   ac.OIDC.IssuerURL,
		ClientID:    ac.OIDC.ClientID,
		ExtraScopes: []string{"profile", "email", "roles", "offline_access"},
	}
}

// SetupDefaultTokenCache sets up ec.TokenCache from flags if it's not already set
// It should be called from rootCmd.PersistentPreRunE
func (ac *Context) SetupDefaultTokenCache() (err error) {
	if ac.TokenCache != nil {
		return
	}

	if ac.TokenCache, err = tokencache.NewFSCache(ac.OIDC.TokenCacheDir); err != nil {
		return fmt.Errorf("creating token cache: %w", err)
	}

	return
}
