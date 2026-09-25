package cmd

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/neticdk-k8s/ic/internal/apiclient"
	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/oidc"
	"github.com/neticdk-k8s/ic/internal/usecases/authentication"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_GetClusterVulnerabilitiesCommand(t *testing.T) {
	got := new(bytes.Buffer)
	ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
	ec.Stderr = got
	ec.Stdout = got
	ui.SetDefaultOutput(got)
	ac := ic.NewContext()
	ac.EC = ec
	mockAuthenticator := authentication.NewMockAuthenticator(t)
	mockAuthenticator.EXPECT().
		SetLogger(mock.Anything).
		Run(func(_ *slog.Logger) {}).
		Return()
	mockAuthenticator.EXPECT().
		Login(mock.Anything, mock.Anything).
		Run(func(_ context.Context, in authentication.LoginInput) {}).
		Return(&oidc.TokenSet{
			AccessToken:  "YOUR_ACCESS_TOKEN",
			IDToken:      "YOUR_ID_TOKEN",
			RefreshToken: "YOUR_REFRESH_TOKEN",
		}, nil)
	ac.Authenticator = mockAuthenticator

	id := "CVE-2024-45337"
	severity := "CRITICAL"
	pkgName := "golang.org/x/crypto"
	pkgVersion := "v0.28.0"
	fixVersions := []string{"0.31.0"}
	ref1 := "registry.k8s.io/ingress-nginx/controller:v1.11.3"
	ref2 := "registry.k8s.io/ingress-nginx/controller:v1.11.2"
	categories := []string{"customer", "platform"}
	vulnerabilities := []apiclient.Vulnerability{
		{
			Id:         &id,
			Severity:   &severity,
			Package:    &apiclient.VulnerabilityPackage{Name: &pkgName, Version: &pkgVersion, FixVersions: &fixVersions},
			Images:     &[]apiclient.VulnerabilityImage{{Ref: &ref1}, {Ref: &ref2}},
			Categories: &categories,
		},
	}
	mockClientWithResponsesInterface := apiclient.NewMockClientWithResponsesInterface(t)
	mockClientWithResponsesInterface.EXPECT().
		ListVulnerabilitiesWithResponse(mock.Anything, "my-cluster.my-provider").
		Return(
			&apiclient.ListVulnerabilitiesResponse{
				HTTPResponse: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
				},
				ApplicationldJSONDefault: &apiclient.Vulnerabilities{
					Vulnerabilities: &vulnerabilities,
				},
			}, nil)
	notFound := "Cluster not found"
	mockClientWithResponsesInterface.EXPECT().
		ListVulnerabilitiesWithResponse(mock.Anything, "missing.my-provider").
		Return(
			&apiclient.ListVulnerabilitiesResponse{
				HTTPResponse: &http.Response{
					Status:     "404 Not Found",
					StatusCode: 404,
				},
				ApplicationproblemJSON404: &apiclient.Problem{Title: &notFound},
			}, nil)
	ac.APIClient = mockClientWithResponsesInterface

	cmd := newRootCmd(ac)

	t.Run("csv", func(t *testing.T) {
		got.Reset()
		cmd.SetArgs([]string{"get", "cluster-vulnerabilities", "my-cluster.my-provider", "-o", "csv"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), "Cluster,CVE,Severity,Package Name,Package Version,Fix Versions,Image,Categories\n")
		assert.Contains(t, got.String(), `my-cluster.my-provider,CVE-2024-45337,CRITICAL,golang.org/x/crypto,v0.28.0,0.31.0,`+ref1+`,"customer,platform"`)
		assert.Contains(t, got.String(), ref2)
	})

	t.Run("json", func(t *testing.T) {
		got.Reset()
		cmd.SetArgs([]string{"get", "cluster-vulnerabilities", "my-cluster.my-provider", "-o", "json"})
		err := cmd.ExecuteContext(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, got.String(), `"cluster_id": "my-cluster.my-provider"`)
		assert.Contains(t, got.String(), `"id": "CVE-2024-45337"`)
	})

	t.Run("output dir skips failing cluster", func(t *testing.T) {
		got.Reset()
		dir := t.TempDir()
		cmd.SetArgs([]string{"get", "cluster-vulnerabilities", "missing.my-provider", "my-cluster.my-provider", "-o", "csv", "--output-dir", dir})
		err := cmd.ExecuteContext(context.Background())
		assert.Error(t, err)
		assert.Contains(t, got.String(), "Cluster not found")

		data, err := os.ReadFile(filepath.Join(dir, "cve_my-cluster.my-provider.csv"))
		assert.NoError(t, err)
		assert.Contains(t, string(data), "CVE-2024-45337")
		assert.NoFileExists(t, filepath.Join(dir, "cve_missing.my-provider.csv"))
	})
}
