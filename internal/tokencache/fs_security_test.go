package tokencache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk/go-stdlib/assert"
	"github.com/neticdk/go-stdlib/require"
)

func TestFSCacheRejectsEscapingSymlinks(t *testing.T) {
	for _, relative := range []bool{false, true} {
		for _, operation := range []string{"lookup", "save"} {
			name := "absolute/" + operation
			if relative {
				name = "relative/" + operation
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				cacheDir := filepath.Join(dir, "cache")
				require.NoError(t, os.Mkdir(cacheDir, 0o700))
				target := filepath.Join(dir, "outside.json")
				original := []byte(`{"access_token":"outside"}`)
				require.NoError(t, os.WriteFile(target, original, 0o600))
				key := Key{IssuerURL: "issuer", ClientID: "client"}
				filename, err := computeFilename(key)
				require.NoError(t, err)
				linkTarget := target
				if relative {
					linkTarget = filepath.Join("..", "outside.json")
				}
				require.NoError(t, os.Symlink(linkTarget, filepath.Join(cacheDir, filename)))
				cache, err := NewFSCache(cacheDir)
				require.NoError(t, err)

				if operation == "lookup" {
					tokenSet, err := cache.Lookup(key)
					assert.Error(t, err)
					assert.Nil(t, tokenSet)
				} else {
					assert.Error(t, cache.Save(key, oidc.TokenSet{AccessToken: "replacement"}))
				}
				data, err := os.ReadFile(target)
				require.NoError(t, err)
				assert.Equal(t, data, original)
			})
		}
	}
}

func TestFSCacheLookupMissing(t *testing.T) {
	for _, existingDir := range []bool{false, true} {
		name := "missing directory"
		if existingDir {
			name = "missing file"
		}
		t.Run(name, func(t *testing.T) {
			cacheDir := filepath.Join(t.TempDir(), "cache")
			if existingDir {
				require.NoError(t, os.Mkdir(cacheDir, 0o700))
			}
			cache, err := NewFSCache(cacheDir)
			require.NoError(t, err)
			tokenSet, err := cache.Lookup(Key{IssuerURL: "issuer", ClientID: "client"})
			var miss *CacheMissError
			assert.ErrorAs(t, err, &miss)
			assert.Nil(t, tokenSet)
		})
	}
}

func TestFSCacheSavePermissions(t *testing.T) {
	for _, existingFile := range []bool{false, true} {
		name := "new file"
		if existingFile {
			name = "existing file with loose permissions"
		}
		t.Run(name, func(t *testing.T) {
			cacheDir := filepath.Join(t.TempDir(), "cache")
			key := Key{IssuerURL: "issuer", ClientID: "client"}
			filename, err := computeFilename(key)
			require.NoError(t, err)
			path := filepath.Join(cacheDir, filename)
			if existingFile {
				require.NoError(t, os.Mkdir(cacheDir, 0o700))
				require.NoError(t, os.WriteFile(path, []byte("old contents"), 0o600))
				require.NoError(t, os.Chmod(path, 0o644))
			}
			cache, err := NewFSCache(cacheDir)
			require.NoError(t, err)
			tokenSet := oidc.TokenSet{AccessToken: "access", IDToken: "id", RefreshToken: "refresh"}
			require.NoError(t, cache.Save(key, tokenSet))

			info, err := os.Stat(cacheDir)
			require.NoError(t, err)
			assert.Equal(t, info.Mode().Perm(), os.FileMode(0o700))
			info, err = os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, info.Mode().Perm(), os.FileMode(0o600))
			got, err := cache.Lookup(key)
			require.NoError(t, err)
			assert.Equal(t, got, &tokenSet)
		})
	}
}
