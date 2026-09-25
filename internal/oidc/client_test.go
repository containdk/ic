package oidc

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/neticdk/go-stdlib/assert"
	"github.com/neticdk/go-stdlib/require"
	"golang.org/x/oauth2"
)

func TestGetTokenByAuthCodeRedirectURL(t *testing.T) {
	for _, tt := range []struct {
		name     string
		hostname string
		wantHost string
	}{
		{name: "default hostname", wantHost: "localhost"},
		{name: "localhost with automatic port", hostname: "localhost", wantHost: "localhost"},
		{name: "custom hostname", hostname: "login.example.com", wantHost: "login.example.com"},
		{name: "IPv4 hostname", hostname: "127.0.0.1", wantHost: "127.0.0.1"},
		{name: "IPv6 hostname", hostname: "::1", wantHost: "::1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			bindAddress := "127.0.0.1:0"
			if tt.hostname != "" && tt.hostname != "localhost" {
				var lc net.ListenConfig
				listener, err := lc.Listen(ctx, "tcp", bindAddress)
				require.NoError(t, err)
				bindAddress = listener.Addr().String()
				require.NoError(t, listener.Close())
			}
			c := &client{
				oauth2config: oauth2.Config{
					ClientID: "test-client",
					Endpoint: oauth2.Endpoint{AuthURL: "https://issuer.example.com/authorize"},
				},
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			ready := make(chan string, 1)
			result := make(chan error, 1)
			go func() {
				_, err := c.GetTokenByAuthCode(ctx, GetTokenByAuthCodeInput{
					BindAddress:         bindAddress,
					RedirectURLHostname: tt.hostname,
					PKCEVerifier:        oauth2.GenerateVerifier(),
				}, ready)
				result <- err
			}()
			var readyURL string
			select {
			case readyURL = <-ready:
			case err := <-result:
				t.Fatalf("starting OIDC callback server: %v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			t.Cleanup(func() {
				cancel()
				select {
				case err := <-result:
					require.ErrorIs(t, err, context.Canceled)
				case <-time.After(10 * time.Second):
					t.Error("OIDC callback server did not stop")
				}
			})

			localURL, err := url.Parse(readyURL)
			require.NoError(t, err)
			assert.Equal(t, localURL.Hostname(), tt.wantHost)
			assert.NotEqual(t, localURL.Port(), "0")
			// Connect to the listener without resolving the configured redirect hostname.
			localURL.Host = net.JoinHostPort("127.0.0.1", localURL.Port())
			httpClient := &http.Client{
				CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			defer httpClient.CloseIdleConnections()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, localURL.String(), nil)
			require.NoError(t, err)
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, res.StatusCode, http.StatusFound)
			authURL, err := res.Location()
			require.NoError(t, err)
			assert.Equal(t, authURL.Query().Get("redirect_uri"), strings.TrimSuffix(readyURL, "/"))
			assert.Empty(t, c.oauth2config.RedirectURL)
		})
	}
}

func TestGetTokenByAuthCodeInvalidCustomRedirect(t *testing.T) {
	for _, tt := range []struct {
		bindAddress string
		wantError   string
	}{
		{bindAddress: "127.0.0.1", wantError: "parsing OIDC bind address"},
		{bindAddress: "127.0.0.1:0", wantError: "requires a fixed bind port"},
		{bindAddress: "127.0.0.1:", wantError: "requires a fixed bind port"},
	} {
		t.Run(tt.bindAddress, func(t *testing.T) {
			c := &client{}
			_, err := c.GetTokenByAuthCode(t.Context(), GetTokenByAuthCodeInput{
				BindAddress:         tt.bindAddress,
				RedirectURLHostname: "login.example.com",
			}, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}
