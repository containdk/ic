package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ParseAccessToken parses an oauth2 access token into a JWT token
// It does not perform verification
func ParseAccessToken(accessToken string) (*jwt.Token, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	return token, err
}

// DecodeWithoutVerify decodes the JWT string and returns the claims.
// Note that this method does not verify the signature and always trust it.
func DecodeWithoutVerify(s string) (*Claims, error) {
	payload, err := DecodePayloadAsRawJSON(s)
	if err != nil {
		return nil, fmt.Errorf("could not decode the payload: %w", err)
	}
	var claims struct {
		Subject   string `json:"sub,omitempty"`
		ExpiresAt int64  `json:"exp,omitempty"`
	}
	if err := json.NewDecoder(bytes.NewReader(payload)).Decode(&claims); err != nil {
		return nil, fmt.Errorf("could not decode the json of token: %w", err)
	}
	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, payload, "", "  "); err != nil {
		return nil, fmt.Errorf("could not indent the json of token: %w", err)
	}
	return &Claims{
		Subject: claims.Subject,
		Expiry:  time.Unix(claims.ExpiresAt, 0),
		Pretty:  prettyJson.String(),
	}, nil
}

// DecodePayloadAsRawJSON extracts the payload and returns the raw JSON.
func DecodePayloadAsRawJSON(s string) ([]byte, error) {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("wants %d segments but got %d segments", 3, len(parts))
	}
	payloadJSON, err := decodePayload(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode the payload: %w", err)
	}
	return payloadJSON, nil
}

func decodePayload(payload string) ([]byte, error) {
	b, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}
	return b, nil
}
