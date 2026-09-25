package tokencache

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/pkg/errors"
)

const (
	newDirPermissions  = 0o700
	newFilePermissions = 0o600
)

type fsCache struct {
	CacheDir string
}

type cachedToken struct {
	AccessToken  string `json:"access_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// NewFSCache creates a new filesystem backed cache
func NewFSCache(cacheDir string) (*fsCache, error) {
	cache := &fsCache{
		CacheDir: cacheDir,
	}
	return cache, nil
}

// Lookup retrieves a cached token
func (c *fsCache) Lookup(key Key) (*oidc.TokenSet, error) {
	filename, err := computeFilename(key)
	if err != nil {
		return nil, fmt.Errorf("could not compute the key: %w", err)
	}
	p := filepath.Join(c.CacheDir, filename)
	f, err := os.OpenInRoot(c.CacheDir, filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &CacheMissError{}
		}
		return nil, fmt.Errorf("could not open file %s: %w", p, err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var e cachedToken
	if err := d.Decode(&e); err != nil {
		return nil, fmt.Errorf("invalid json file %s: %w", p, err)
	}
	return &oidc.TokenSet{
		AccessToken:  e.AccessToken,
		IDToken:      e.IDToken,
		RefreshToken: e.RefreshToken,
	}, nil
}

// Save stores a cached token
func (c *fsCache) Save(key Key, tokenSet oidc.TokenSet) error {
	if err := os.MkdirAll(c.CacheDir, newDirPermissions); err != nil {
		return fmt.Errorf("could not create directory %s: %w", c.CacheDir, err)
	}
	filename, err := computeFilename(key)
	if err != nil {
		return fmt.Errorf("could not compute the key: %w", err)
	}
	p := filepath.Join(c.CacheDir, filename)
	root, err := os.OpenRoot(c.CacheDir)
	if err != nil {
		return fmt.Errorf("could not open cache directory %s: %w", c.CacheDir, err)
	}
	defer root.Close()
	f, err := root.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, newFilePermissions)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", p, err)
	}
	defer f.Close()
	if err := f.Chmod(newFilePermissions); err != nil {
		return fmt.Errorf("could not set permissions on cache file %s: %w", p, err)
	}
	e := cachedToken{
		AccessToken:  tokenSet.AccessToken,
		IDToken:      tokenSet.IDToken,
		RefreshToken: tokenSet.RefreshToken,
	}
	// #nosec G117 -- Token serialization is intentional in this owner-only credential cache.
	if err := json.NewEncoder(f).Encode(&e); err != nil {
		return fmt.Errorf("json encode error: %w", err)
	}
	return nil
}

// Invalidate deletes a cached token
func (c *fsCache) Invalidate(key Key) error {
	filename, err := computeFilename(key)
	if err != nil {
		return fmt.Errorf("could not compute the key: %w", err)
	}
	p := filepath.Join(c.CacheDir, filename)

	if err := os.Remove(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &CacheMissError{}
		}
		return fmt.Errorf("could not remove file %s: %w", p, err)
	}
	return nil
}

func computeFilename(key Key) (string, error) {
	s := sha256.New()
	e := gob.NewEncoder(s)
	if err := e.Encode(&key); err != nil {
		return "", fmt.Errorf("could not encode the key: %w", err)
	}
	h := hex.EncodeToString(s.Sum(nil))
	return h, nil
}
