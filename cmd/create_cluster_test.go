package cmd

import (
	"bytes"
	"context"
	goerr "errors"
	"log/slog"
	"net/http"
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

func Test_CreateClusterCommand(t *testing.T) {
	ec, got := newMockedCreateClusterEC(t)
	cmd := newRootCmd(ec)

	cmd.SetArgs([]string{"create", "cluster", "--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123456", "--infrastructure-provider", "netic", "--resilience-zone", "platform"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, got.String(), "Logging in")
	assert.Contains(t, got.String(), "Creating cluster")
	assert.Contains(t, got.String(), "my-cluster")
	assert.Contains(t, got.String(), "my-provider")
}

func Test_CreateClusterCommandWithJSONOutput(t *testing.T) {
	ec, got := newMockedCreateClusterEC(t)
	cmd := newRootCmd(ec)

	cmd.SetArgs([]string{"create", "cluster", "--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123456", "--infrastructure-provider", "netic", "--resilience-zone", "platform", "-o", "json"})
	err := cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, got.String(), "\"name\": \"my-cluster\"")
	assert.Contains(t, got.String(), "\"provider_name\": \"my-provider\"")
}

func Test_CreateClusterCommandRequiredParameters(t *testing.T) {
	testCases := []struct {
		testName string
		args     []string
	}{
		{
			testName: "no name",
			args:     []string{"--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform"},
		},
		{
			testName: "no provider",
			args:     []string{"--name", "my-cluster", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform"},
		},
		{
			testName: "no environment",
			args:     []string{"--name", "my-cluster", "--provider", "my-provider", "--subscription", "123446", "--resilience-zone", "platform"},
		},
		{
			testName: "no subscription",
			args:     []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--resilience-zone", "platform"},
		},
		{
			testName: "no resilience zone",
			args:     []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123456"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			got := new(bytes.Buffer)
			ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
			ec.Stderr = got
			ec.Stdout = got
			ui.SetDefaultOutput(got)
			ac := ic.NewContext()
			ac.EC = ec
			cmd := newRootCmd(ac)
			args := append([]string{"create", "cluster"}, tc.args...)
			cmd.SetArgs(args)
			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "required")
		})
	}
}

func Test_CreateClusterCommandInvalidParameters(t *testing.T) {
	testCases := []struct {
		testName     string
		args         []string
		expErrString string
	}{
		{
			testName:     "invalid partition",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "invalid"},
			expErrString: "is not valid",
		},
		{
			testName:     "invalid region",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "invalid"},
			expErrString: "is not valid",
		},
		{
			testName:     "invalid region in partition",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "invalid"},
			expErrString: "is not valid",
		},
		{
			testName:     "custom operations without valid url",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "dk-north", "--has-co", "--co-url", "invalid://host"},
			expErrString: "must be a URL",
		},
		{
			testName:     "name is invalid rfc1035 label",
			args:         []string{"--name", "my cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "dk-north"},
			expErrString: "must be an RFC1035",
		},
		{
			testName:     "provider is invalid rfc1035 label",
			args:         []string{"--name", "mycluster", "--provider", "my provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "dk-north"},
			expErrString: "must be an RFC1035",
		},
		{
			testName:     "resilience zone is invalid rfc1035 label",
			args:         []string{"--name", "mycluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "my platform", "--partition", "netic", "--region", "dk-north"},
			expErrString: "must be an RFC1035",
		},
		{
			testName:     "environment is invalid rfc1035 label",
			args:         []string{"--name", "mycluster", "--provider", "my-provider", "--environment", "invalid environment", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "dk-north"},
			expErrString: "must be an RFC1035",
		},
		{
			testName:     "invalid infrastructure provider",
			args:         []string{"--name", "mycluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123446", "--resilience-zone", "platform", "--partition", "netic", "--region", "dk-north", "--infrastructure-provider", "invalid"},
			expErrString: "invalid",
		},
		{
			testName:     "invalid subscription length",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "446", "--resilience-zone", "platform", "--partition", "netic", "--region", "dk-north"},
			expErrString: "minimum 5 characters",
		},
		{
			testName:     "invalid subscription",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "ΩΩΩΩΩ", "--resilience-zone", "platform"},
			expErrString: "must be an ASCII string",
		},
		{
			testName:     "has-co required with co-url",
			args:         []string{"--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123456", "--resilience-zone", "platform", "--has-co"},
			expErrString: "they must all be set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			got := new(bytes.Buffer)
			ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
			ec.Stderr = got
			ec.Stdout = got
			ui.SetDefaultOutput(got)
			ac := ic.NewContext()
			ac.EC = ec
			command := newRootCmd(ac)
			args := append([]string{"create", "cluster"}, tc.args...)
			command.SetArgs(args)
			err := command.Execute()
			if err != nil {
				var invalidArgErr *cmd.InvalidArgumentError
				if goerr.As(err, &invalidArgErr) {
					assert.Contains(t, err.(cmd.ErrorWithHelp).Help(), tc.expErrString)
				} else {
					assert.Contains(t, err.Error(), tc.expErrString)
				}
			}
		})
	}
}

func Test_CreateClusterCommandServiceLevels(t *testing.T) {
	baseArgs := []string{"create", "cluster", "--name", "my-cluster", "--provider", "my-provider", "--environment", "test", "--subscription", "123456", "--resilience-zone", "platform"}

	testCases := []struct {
		testName string
		args     []string
		expTO    bool
		expTM    bool
		expAO    bool
		expAM    bool
		expCO    bool
	}{
		{
			testName: "no service level specified",
			args:     baseArgs,
			expTO:    true,
			expTM:    true,
			expAO:    false,
			expAM:    false,
			expCO:    false,
		},
		{
			testName: "--has-co",
			args:     append(baseArgs, []string{"--has-co", "--co-url", "https://example.com"}...),
			expTO:    false,
			expTM:    false,
			expAO:    false,
			expAM:    false,
			expCO:    true,
		},
		{
			testName: "--has-tm",
			args:     append(baseArgs, []string{"--has-tm"}...),
			expTO:    true,
			expTM:    true,
			expAO:    false,
			expAM:    false,
			expCO:    false,
		},
		{
			testName: "--has-ao",
			args:     append(baseArgs, []string{"--has-ao"}...),
			expTO:    true,
			expTM:    true,
			expAO:    true,
			expAM:    false,
			expCO:    false,
		},
		{
			testName: "--has-am",
			args:     append(baseArgs, []string{"--has-am"}...),
			expTO:    true,
			expTM:    true,
			expAO:    true,
			expAM:    true,
			expCO:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ac, _ := newMockedCreateClusterEC(t)
			command := newRootCmd(ac)
			ac.EC.Command = command // Reassign, since the last subcommand is the command of the execution context
			o := &createClusterOptions{}
			err := o.SetupFlags(t.Context(), ac)
			if err != nil {
				t.Log(err)
			}
			assert.NoError(t, err)
			command.SetArgs(tc.args)
			err = command.ParseFlags(tc.args)
			if err != nil {
				t.Log(err)
			}
			assert.NoError(t, err)
			err = o.Complete(t.Context(), ac)
			if err != nil {
				t.Log(err)
			}
			assert.NoError(t, err)
			err = command.Execute()
			if err != nil {
				t.Log(err.(cmd.ErrorWithHelp).Unwrap())
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expCO, o.HasCustomOperations)
			assert.Equal(t, tc.expTO, o.HasTechnicalOperations)
			assert.Equal(t, tc.expTM, o.HasTechnicalManagement)
			assert.Equal(t, tc.expAO, o.HasApplicationOperations)
			assert.Equal(t, tc.expAM, o.HasApplicationManagement)
		})
	}
}

func newMockedCreateClusterEC(t *testing.T) (*ic.Context, *bytes.Buffer) {
	got := new(bytes.Buffer)
	ec := cmd.NewExecutionContext(AppName, ShortDesc, "test")
	ec.Stdin = nil
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
	included := []map[string]interface{}{
		{
			"@id":   "my-provider-id",
			"@type": "Provider",
			"name":  "my-provider",
		},
		{
			"@id":   "my-rz-id",
			"@type": "ResilienceZone",
			"name":  "my-resilience-zone",
		},
		{
			"@id":             "my-cluster-id",
			"@type":           "Cluster",
			"name":            "my-cluster",
			"clusterType":     "dedicated",
			"environmentName": "testing",
			"provider":        "my-provider-id",
			"resilienceZone":  "my-rz-id",
			"kubernetesVersion": map[string]interface{}{
				"version": "v1.2.3",
			},
		},
	}
	name := "my-cluster"
	providerId := "my-provider-id"
	mockClientWithResponsesInterface := apiclient.NewMockClientWithResponsesInterface(t)
	mockClientWithResponsesInterface.EXPECT().
		CreateClusterWithResponse(mock.Anything, mock.Anything).
		Return(
			&apiclient.CreateClusterResponse{
				Body: make([]byte, 0),
				HTTPResponse: &http.Response{
					Status:     "201 CREATED",
					StatusCode: 201,
				},
				ApplicationldJSON201: &apiclient.Cluster{
					Name:     &name,
					Provider: &providerId,
					Included: &included,
				},
			}, nil)
	apiClient := mockClientWithResponsesInterface
	ac.APIClient = apiClient

	return ac, got
}
