package authcode

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/int128/oauth2cli/oauth2params"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

// BrowserLoginInput is th input given to Login
type BrowserLoginInput struct {
	// BindAddress is the IP-address and port used by the redirect url server
	BindAddress string
	// RedirectURLHostname is the hostname of the redirect URL. You can set this
	// if your provider does not accept localhost.
	RedirectURLHostname string
}

// Browser represents a browser based login
type Browser struct {
	// Logger holds a logging instance
	Logger *slog.Logger
}

// Login performs browser based autocode flow, i.e.:
// 1. Starting a http server that response to the redirect url request
// 2. Opening a browser that performs login and redirects to the redirect url
// 3. Validates the token
func (b *Browser) Login(ctx context.Context, in *BrowserLoginInput, oidcClient oidc.Client) (*oidc.TokenSet, error) {
	state, err := oauth2params.NewState()
	if err != nil {
		return nil, fmt.Errorf("could not generate a state: %w", err)
	}

	nonce, err := oauth2params.NewState()
	if err != nil {
		return nil, fmt.Errorf("could not generate a nonce: %w", err)
	}

	pkceVerifier := oauth2.GenerateVerifier()

	authCodeInput := oidc.GetTokenByAuthCodeInput{
		BindAddress:         in.BindAddress,
		RedirectURLHostname: in.RedirectURLHostname,
		PKCEVerifier:        pkceVerifier,
		State:               state,
		Nonce:               nonce,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	ready := make(chan string, 1)
	var out *oidc.TokenSet
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		select {
		case url := <-ready:
			b.Logger.Debug("Open url", "url", url)
			if err := browser.OpenURL(url); err != nil {
				b.Logger.Error("could not open the browser", "err", err)
			}
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		defer close(ready)
		tokenSet, err := oidcClient.GetTokenByAuthCode(ctx, authCodeInput, ready)
		if err != nil {
			return errors.Wrap(err, "getting token")
		}
		out = tokenSet
		b.Logger.Debug("Got a valid token set")
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}

	return out, nil
}
