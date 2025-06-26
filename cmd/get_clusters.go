package cmd

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/neticdk/go-common/pkg/qsparser"
	"github.com/spf13/cobra"
)

const PerPage = 50

var getClustersFilterNames = []string{
	"name", "description", "clusterID", "clusterType", "region", "environmentName",
	"providerName", "navisionSubscriptionNumber", "navisionCustomerNumber",
	"navisionCustomerName", "resilienceZone", "clientVersion", "kubernetesVersion",
}

const getClustersLongDesc = `Get list of clusters.

Supported field names for filters:

name, description, clusterID, clusterType, region, environmentName,
providerName, navisionSubscriptionNumber, navisionCustomerNumber,
navisionCustomerName, resilienceZone, clientVersion, kubernetesVersion
`

const getClustersExample = `
# get all cluster
ic get clusters

# get clusters in the resilience zone 'platform'
ic get clusters --filter resilienceZone=platform

use: 'ic help filters' for more information on using filters`

// New creates a new "get clusters" command
func getClustersCmd(ac *ic.Context) *cobra.Command {
	o := &getClustersOptions{}
	c := cmd.NewSubCommand("clusters", o, ac).
		WithShortDesc("Get list of clusters").
		WithLongDesc(getClustersLongDesc).
		WithExample(getClustersExample).
		WithGroupID(groupCluster).
		Build()

	return c
}

type getClustersOptions struct {
	// A filter has the form: fieldName operator value (e.g. name=Peter)
	//
	// Supported operators:
	// == (or =) - equals
	// != (or !) - not equals
	// >         - greater than
	// <         - less than
	// >=        - greater than or equals
	// <=        - less than or equals
	// =~ (or ~) - matches (case insensitive regular expression)
	// !~        - does not match (case insensitive expression)
	Filters []string
}

func (o *getClustersOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.StringArrayVar(&o.Filters, "filter", []string{}, "Filter output based on conditions")
	return nil
}

func (o *getClustersOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *getClustersOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getClustersOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Clusters")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	searchFields := make(map[string]*qsparser.SearchField)
	for _, f := range o.Filters {
		out, err := parseFilter(f)
		if err != nil {
			return err
		}
		searchFields[out.FieldName] = out.SearchField
	}

	var result *cluster.ListClusterResults

	if err := ui.Spin(ac.EC.Spinner, "Getting clusters", func(_ ui.Spinner) error {
		in := cluster.ListClustersInput{
			Logger:    logger,
			APIClient: ac.APIClient,
			PerPage:   PerPage,
			Filters:   searchFields,
		}
		result, err = cluster.ListClusters(ctx, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Listing clusters",
			"See details for more information",
			err,
			0,
		)
	}
	if result.Problem != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			*result.Problem.Title,
			*result.Problem.Detail,
			nil,
			0,
		)
	}

	r := cluster.NewClustersRenderer(result.ClusterListResponse, result.JSONResponse, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
	if err := r.Render(ac.EC.PFlags.OutputFormat); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Failed to render output",
			"See details for more information",
			err,
			0,
		)
	}

	return nil
}

type parseFilterOut struct {
	FieldName   string
	SearchField *qsparser.SearchField
}

func parseFilter(filterArg string) (*parseFilterOut, error) {
	r := regexp.MustCompile(`^([a-zA-Z0-9]+)(==|!=|>=|<=|=~|!~|=|!|<|>|~| (?i)in | (?i)notin )(.*)$`)
	m := r.FindStringSubmatch(filterArg)
	if m == nil {
		return nil, fmt.Errorf("syntax error in filter: %v", filterArg)
	}
	fieldName := m[1]
	if !slices.Contains(getClustersFilterNames, fieldName) {
		return nil, fmt.Errorf("unknown field name: %s in %s", fieldName, filterArg)
	}
	searchOp := m[2]
	searchVal := m[3]
	ops := map[string]string{
		"=":       "eq",
		"==":      "eq",
		"!=":      "ne",
		"!":       "ne",
		">":       "gt",
		"<":       "lt",
		">=":      "ge",
		"<=":      "le",
		"=~":      "ire",
		"~":       "ire",
		"!~":      "nire",
		" in ":    "in",
		" notin ": "notin",
	}
	field := &qsparser.SearchField{
		SearchVal: &searchVal,
	}
	op, ok := ops[strings.ToLower(searchOp)]
	if !ok {
		return nil, fmt.Errorf("unknown search operator: %s in %s", searchOp, filterArg)
	}
	field.SearchOp = &op

	return &parseFilterOut{FieldName: fieldName, SearchField: field}, nil
}
