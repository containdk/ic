package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk-k8s/ic/internal/validation"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
)

// New creates a new "update cluster" command
func updateClusterCmd(ac *ic.Context) *cobra.Command {
	o := &updateClusterOptions{}
	c := cmd.NewSubCommand("cluster", o, ac).
		WithShortDesc("Update a cluster's metadata").
		WithGroupID(groupCluster).
		WithExactArgs(1).
		Build()
	c.Use = "cluster CLUSTER-ID" //nolint:goconst

	return c
}

type updateClusterOptions struct {
	clusterID                string
	Description              string
	EnvironmentName          string
	SubscriptionID           string
	InfrastructureProvider   string
	ResilienceZone           string
	HasTechnicalOperations   bool
	HasTechnicalManagement   bool
	HasApplicationOperations bool
	HasApplicationManagement bool
	HasCustomOperations      bool
	CustomOperationsURL      string
}

func (o *updateClusterOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.StringVar(&o.Description, "description", "", "Cluster Description")
	f.StringVar(&o.EnvironmentName, "environment", "", "Environment Name")
	f.StringVar(&o.SubscriptionID, "subscription", "", "Subscription ID")
	f.StringVar(&o.InfrastructureProvider, "infrastructure-provider", "netic", fmt.Sprintf("Infrastructure Provider. One of (%s)", strings.Join(types.AllInfrastructureProvidersString(), "|")))
	f.StringVar(&o.ResilienceZone, "resilience-zone", "netic", fmt.Sprintf("Resilience Zone. Should be one of (%s)", strings.Join(types.AllResilienceZonesString(), "|")))
	f.BoolVar(&o.HasTechnicalOperations, "has-to", true, "Technical Operations")
	f.BoolVar(&o.HasTechnicalManagement, "has-tm", true, "Technical Management")
	f.BoolVar(&o.HasApplicationOperations, "has-ao", false, "Application Operations")
	f.BoolVar(&o.HasApplicationManagement, "has-am", false, "Application Management")
	f.BoolVar(&o.HasCustomOperations, "has-co", false, "Custom Operations")
	f.StringVar(&o.CustomOperationsURL, "co-url", "", "Custom Operations URL")

	f.SortFlags = false
	c.MarkFlagsRequiredTogether("has-co", "co-url")
	c.MarkFlagsOneRequired("description", "environment", "subscription", "infrastructure-provider", "resilience-zone", "has-to", "has-tm", "has-ao", "has-am", "has-co")
	return nil
}

func (o *updateClusterOptions) Complete(_ context.Context, ac *ic.Context) error {
	o.clusterID = ac.EC.CommandArgs[0]

	if o.HasCustomOperations {
		o.HasTechnicalOperations = false
		o.HasTechnicalManagement = false
		o.HasApplicationOperations = false
		o.HasApplicationManagement = false
	}
	if o.HasTechnicalManagement {
		o.HasTechnicalOperations = true
	}
	if o.HasApplicationOperations {
		o.HasTechnicalManagement = true
	}
	if o.HasApplicationManagement {
		o.HasApplicationOperations = true
	}

	return nil
}

func (o *updateClusterOptions) Validate(ctx context.Context, ac *ic.Context) error {
	command := ac.EC.Command

	if o.HasCustomOperations && !validation.IsWebURL(o.CustomOperationsURL) {
		return &cmd.InvalidArgumentError{
			Flag:    "co-url",
			Val:     o.CustomOperationsURL,
			Context: "must be a URL using a http(s) scheme",
		}
	}
	rfc1035FieldFlags := []struct {
		Flag string
		Val  string
	}{
		{
			Flag: "resilience-zone",
			Val:  o.ResilienceZone,
		},
		{
			Flag: "environment",
			Val:  o.EnvironmentName,
		},
	}
	for _, v := range rfc1035FieldFlags {
		if command.Flags().Changed(v.Flag) {
			if !validation.IsDNSRFC1035Label(v.Val) {
				return &cmd.InvalidArgumentError{
					Flag:    v.Flag,
					Val:     v.Val,
					Context: "must be an RFC1035 DNS label",
				}
			}
		}
	}
	if command.Flags().Changed("infrastructure-provider") {
		if !slices.Contains(types.AllInfrastructureProvidersString(), o.InfrastructureProvider) {
			return &cmd.InvalidArgumentError{
				Flag:  "infrastructure-provider",
				Val:   o.InfrastructureProvider,
				OneOf: types.AllInfrastructureProvidersString(),
			}
		}
	}
	if command.Flags().Changed("resilience-zone") {
		if !slices.Contains(types.AllResilienceZonesString(), o.ResilienceZone) {
			ac.EC.Logger.WarnContext(ctx, fmt.Sprintf("Non-standard resilience zone used: %s", o.ResilienceZone))
		}
	}
	if command.Flags().Changed("subscription") {
		if !validation.IsPrintableASCII(o.SubscriptionID) || len(o.SubscriptionID) < 5 {
			return &cmd.InvalidArgumentError{
				Flag:    "subscription",
				Val:     o.SubscriptionID,
				Context: "must be an ASCII string of minimum 5 characters length",
			}
		}
	}
	return nil
}

func (o *updateClusterOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Clusters")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *cluster.UpdateClusterResult
	spinnerText := fmt.Sprintf("Updating cluster metadata for %q", o.clusterID)
	if err := ui.Spin(ac.EC.Spinner, spinnerText, func(s ui.Spinner) error {
		in := cluster.UpdateClusterInput{
			Logger:    logger,
			APIClient: ac.APIClient,
		}
		if ac.EC.Command.Flags().Changed("description") {
			in.Description = &o.Description
		}
		if ac.EC.Command.Flags().Changed("environment") {
			in.EnvironmentName = &o.EnvironmentName
		}
		if ac.EC.Command.Flags().Changed("resilience-zone") {
			in.ResilienceZone = &o.ResilienceZone
		}
		if ac.EC.Command.Flags().Changed("subscription") {
			in.SubscriptionID = &o.SubscriptionID
		}
		if ac.EC.Command.Flags().Changed("infrastructure-provider") {
			in.InfrastructureProvider = &o.InfrastructureProvider
		}
		if ac.EC.Command.Flags().Changed("has-to") ||
			ac.EC.Command.Flags().Changed("has-tm") ||
			ac.EC.Command.Flags().Changed("has-ao") ||
			ac.EC.Command.Flags().Changed("has-am") ||
			ac.EC.Command.Flags().Changed("has-co") {
			in.HasTechnicalOperations = &o.HasTechnicalOperations
			in.HasTechnicalManagement = &o.HasTechnicalManagement
			in.HasApplicationOperations = &o.HasApplicationOperations
			in.HasApplicationManagement = &o.HasApplicationManagement
			in.HasCustomOperations = &o.HasCustomOperations
		}
		if ac.EC.Command.Flags().Changed("co-url") {
			in.CustomOperationsURL = &o.CustomOperationsURL
		}
		result, err = cluster.UpdateCluster(ctx, o.clusterID, in)
		if err == nil {
			ui.UpdateSpinnerText(s, "Cluster metadata updated")
		}
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Updating cluster metadata",
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

	r := cluster.NewClusterRenderer(result.ClusterResponse, result.JSONResponse, ac.EC.Stdout)
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
