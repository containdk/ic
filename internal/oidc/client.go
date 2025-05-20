package oidc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/int128/oauth2cli"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	logoutRetryMinWait = time.Duration(2) * time.Second
	logoutRetryMaxWait = time.Duration(30) * time.Second
)

// Client represents an OIDC Client
type Client interface {
	// Refresh creates an updated TokenSet by means of refreshing an oauth2 token
	Refresh(ctx context.Context, refreshToken string) (*TokenSet, error)
	// Logout deletes the session from the OIDC provider
	Logout(idToken string) error
	// GetTokenByAuthCode performs the Authorization Code Grant Flow and returns
	GetTokenByAuthCode(ctx context.Context, in GetTokenByAuthCodeInput, localServerReadyChan chan<- string) (*TokenSet, error)
	// GetAuthCodeURL returns a URL to OAuth 2.0 provider's consent page
	GetAuthCodeURL(ctx context.Context, in GetAuthCodeURLInput) (string, error)
	// ExchangeAuthCode converts an authorization code into a TokenSet
	ExchangeAuthCode(ctx context.Context, in ExchangeAuthCodeInput) (*TokenSet, error)
}

type client struct {
	provider          *gooidc.Provider
	oauth2config      oauth2.Config
	providerLogoutURL string
	logger            *slog.Logger
}

// Refresh creates an updated TokenSet by means of refreshing an oauth2 token
func (c *client) Refresh(ctx context.Context, refreshToken string) (*TokenSet, error) {
	currentToken := &oauth2.Token{
		Expiry:       time.Now(),
		RefreshToken: refreshToken,
	}
	source := c.oauth2config.TokenSource(ctx, currentToken)
	token, err := source.Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	return c.verifyToken(token, "")
}

// Logout deletes the session from the OIDC provider
func (c *client) Logout(idToken string) error {
	if c.providerLogoutURL == "" {
		return errors.New("logout URL not set")
	}

	logoutURL, err := url.Parse(c.providerLogoutURL)
	if err != nil {
		return fmt.Errorf("parsing logout URL: %w", err)
	}

	query := logoutURL.Query()
	if idToken != "" {
		query.Set("id_token_hint", idToken)
	}
	logoutURL.RawQuery = query.Encode()

	res, err := c.logoutWithRetries(logoutURL.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status code: %d", res.StatusCode)
	}

	return nil
}

func (c *client) logoutWithRetries(logoutURL string) (*http.Response, error) {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Duration(2) * time.Second
	client.Logger = c.logger
	client.RetryWaitMin = logoutRetryMinWait
	client.RetryWaitMax = logoutRetryMaxWait
	client.RetryMax = 5

	defer client.HTTPClient.CloseIdleConnections()
	return client.Get(logoutURL)
}

// GetTokenByAuthCodeInput is the input given to GetTokenByAuthCode
type GetTokenByAuthCodeInput struct {
	// BindAddress is the IP-address and port used by the redirect url server
	BindAddress string
	// RedirectURLHostname is the hostname of the redirect URL. You can set this
	// if your provider does not accept localhost.
	RedirectURLHostname string
	// PKCEVerifier is the verifier used to verify the token
	PKCEVerifier string
	// OAuth 2.0 state
	State string
	// OIDC Nonce
	Nonce string
}

// GetTokenByAuthCode performs the Authorization Code Grant Flow and returns
// a token received from the provider.
//
// It does this by creating a local http server used for serving the RedirectURL
// and opening a browser where the user logs in
func (c *client) GetTokenByAuthCode(ctx context.Context, in GetTokenByAuthCodeInput, localServerReadyChan chan<- string) (*TokenSet, error) {
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(in.PKCEVerifier),
		oauth2.AccessTypeOffline,
		gooidc.Nonce(in.Nonce),
	}

	cfg := oauth2cli.Config{
		OAuth2Config:           c.oauth2config,
		State:                  in.State,
		AuthCodeOptions:        authCodeOptions,
		TokenRequestOptions:    []oauth2.AuthCodeOption{oauth2.VerifierOption(in.PKCEVerifier)},
		LocalServerReadyChan:   localServerReadyChan,
		RedirectURLHostname:    in.RedirectURLHostname,
		LocalServerBindAddress: []string{in.BindAddress},
		Logf: func(format string, args ...any) {
			c.logger.DebugContext(ctx, fmt.Sprintf(format, args...))
		},
	}

	token, err := oauth2cli.GetToken(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("oauth2 error: %w", err)
	}

	return c.verifyToken(token, in.Nonce)
}

// GetGetAuthCodeURL is the input given to GetAuthCodeURL
type GetAuthCodeURLInput struct {
	// RedirectURI is the redirect url
	// This is typicalle an URN such as urn:ietf:wg:oauth:2.0:oob
	RedirectURI string
	// PKCEVerifier is the verifier used to verify the token
	PKCEVerifier string
	// OAuth 2.0 state
	State string
	// OIDC Nonce
	Nonce string
}

// GetAuthCodeURL returns a URL to OAuth 2.0 provider's consent page
func (c *client) GetAuthCodeURL(_ context.Context, in GetAuthCodeURLInput) (string, error) {
	cfg := c.oauth2config
	cfg.RedirectURL = in.RedirectURI

	requestOptions := []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(in.PKCEVerifier),
		oauth2.AccessTypeOffline,
		gooidc.Nonce(in.Nonce),
	}

	return cfg.AuthCodeURL(in.State, requestOptions...), nil
}

// ExchangeAuthCodeInput holds the input parameters for ExchangeAuthCode()
type ExchangeAuthCodeInput struct {
	Code string
	// PKCEVerifier is the verifier used to verify the token
	PKCEVerifier string
	// OIDC Nonce
	Nonce string
	// RedirectURI is the redirect url
	RedirectURI string
}

// ExchangeAuthCode converts an authorization code into a TokenSet
func (c *client) ExchangeAuthCode(ctx context.Context, in ExchangeAuthCodeInput) (*TokenSet, error) {
	cfg := c.oauth2config
	cfg.RedirectURL = in.RedirectURI

	token, err := cfg.Exchange(ctx, in.Code, []oauth2.AuthCodeOption{oauth2.VerifierOption(in.PKCEVerifier)}...)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	return c.verifyToken(token, in.Nonce)
}

func (c *client) verifyToken(token *oauth2.Token, nonce string) (*TokenSet, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("getting id token")
	}

	verifier := c.provider.Verifier(&gooidc.Config{ClientID: c.oauth2config.ClientID})
	verifiedIDToken, err := verifier.Verify(context.Background(), idToken)
	if err != nil {
		return nil, fmt.Errorf("verifying id token: %w", err)
	}

	if err = verifiedIDToken.VerifyAccessToken(token.AccessToken); err != nil {
		return nil, fmt.Errorf("verifying access token: %w", err)
	}

	if nonce != "" && nonce != verifiedIDToken.Nonce {
		return nil, fmt.Errorf("verifying nonce (wants %s but got %s)", nonce, verifiedIDToken.Nonce)
	}

	return &TokenSet{
		AccessToken:  token.AccessToken,
		IDToken:      idToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
