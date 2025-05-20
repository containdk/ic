package authcode

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/int128/oauth2cli/oauth2params"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/reader"
	"golang.org/x/oauth2"
)

const keyboardPrompt = "Enter code: "

// KeyboardLoginInput is the input given to Login
type KeyboardLoginInput struct {
	// RedirectURI is the URI used for redirection after login
	RedirectURI string
}

// Keyboard represents a keyboard based login
type Keyboard struct {
	// Reader is used to read input from stdin
	Reader reader.Reader
	// Logger holds a logging instance
	Logger *slog.Logger
}

// Login performs keyboard based autocode flow, i.e.:
// 1. Getting the auth code URL from the OIDC issuer
// 2. Printing the URL and requesting the code to be entered
// 3. Validating the code and token against the OIDC issuer
func (k *Keyboard) Login(ctx context.Context, in *KeyboardLoginInput, oidcClient oidc.Client) (*oidc.TokenSet, error) {
	state, err := oauth2params.NewState()
	if err != nil {
		return nil, fmt.Errorf("could not generate a state: %w", err)
	}

	nonce, err := oauth2params.NewState()
	if err != nil {
		return nil, fmt.Errorf("could not generate a nonce: %w", err)
	}

	pkceVerifier := oauth2.GenerateVerifier()

	authCodeURL, err := oidcClient.GetAuthCodeURL(ctx, oidc.GetAuthCodeURLInput{
		State:        state,
		Nonce:        nonce,
		PKCEVerifier: pkceVerifier,
		RedirectURI:  in.RedirectURI,
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Please visit the following URL in your browser: %s\n", authCodeURL)
	code, err := k.Reader.ReadString(keyboardPrompt)
	if err != nil {
		return nil, fmt.Errorf("reading authorization code: %w", err)
	}

	k.Logger.DebugContext(ctx, "Exchanging code and token")
	tokenSet, err := oidcClient.ExchangeAuthCode(ctx, oidc.ExchangeAuthCodeInput{
		Code:         code,
		PKCEVerifier: pkceVerifier,
		Nonce:        nonce,
		RedirectURI:  in.RedirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("exchanging authorization code: %w", err)
	}

	return tokenSet, nil
}
