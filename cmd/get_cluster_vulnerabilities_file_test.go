package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-stdlib/assert"
	"github.com/neticdk/go-stdlib/require"
)

func TestClusterVulnerabilitiesOutputFiles(t *testing.T) {
	for _, tc := range []struct {
		format string
		ext    string
	}{
		{cluster.FormatPlain, "txt"},
		{cluster.FormatTable, "txt"},
		{cluster.FormatJson, "json"},
		{cluster.FormatCSV, "csv"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			o := &getClusterVulnerabilitiesOptions{outputDir: filepath.Join(t.TempDir(), "exports")}
			ac := ic.NewContext()
			ac.EC = cmd.NewExecutionContext(AppName, ShortDesc, "test")
			ac.EC.PFlags.OutputFormat = tc.format
			require.NoError(t, o.Validate(t.Context(), ac))
			require.NoError(t, o.writeFile(&cluster.ClusterVulnerabilities{ClusterID: "test.provider"}, tc.format, false))

			info, err := os.Stat(o.outputDir)
			require.NoError(t, err)
			assert.Equal(t, info.Mode().Perm()&^os.FileMode(0o750), os.FileMode(0))
			info, err = os.Stat(filepath.Join(o.outputDir, "cve_test.provider."+tc.ext))
			require.NoError(t, err)
			assert.Equal(t, info.Mode().Perm()&^os.FileMode(0o640), os.FileMode(0))
		})
	}
}

func TestClusterVulnerabilitiesRejectsOutputEscape(t *testing.T) {
	for _, attack := range []string{"traversal", "absolute symlink", "relative symlink"} {
		t.Run(attack, func(t *testing.T) {
			dir := t.TempDir()
			outputDir := filepath.Join(dir, "exports")
			require.NoError(t, os.Mkdir(outputDir, 0o750))
			target := filepath.Join(dir, "outside.json")
			original := []byte("original contents")
			require.NoError(t, os.WriteFile(target, original, 0o600))
			clusterID := "test.provider"
			switch attack {
			case "traversal":
				clusterID = "nested/../../outside"
				require.NoError(t, os.Mkdir(filepath.Join(outputDir, "cve_nested"), 0o750))
			case "absolute symlink":
				require.NoError(t, os.Symlink(target, filepath.Join(outputDir, "cve_test.provider.json")))
			case "relative symlink":
				require.NoError(t, os.Symlink(filepath.Join("..", "outside.json"), filepath.Join(outputDir, "cve_test.provider.json")))
			}

			o := &getClusterVulnerabilitiesOptions{outputDir: outputDir}
			err := o.writeFile(&cluster.ClusterVulnerabilities{ClusterID: clusterID}, cluster.FormatJson, false)
			assert.Error(t, err)
			data, err := os.ReadFile(target)
			require.NoError(t, err)
			assert.Equal(t, data, original)
		})
	}
}
