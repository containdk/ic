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

func createClusterCmd(ac *ic.Context) *cobra.Command {
	o := &createClusterOptions{}
	c := cmd.NewSubCommand("cluster", o, ac).
		WithShortDesc("Create a cluster").
		WithGroupID(groupCluster).
		Build()

	return c
}

type createClusterOptions struct {
	Name                     string
	Description              string
	ProviderName             string
	EnvironmentName          string
	Partition                string
	Region                   string
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

func (o *createClusterOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.StringVar(&o.Name, "name", "", "Cluster name")
	f.StringVar(&o.ProviderName, "provider", "", "Provider Name")
	f.StringVar(&o.Description, "description", "", "Cluster Description")
	f.StringVar(&o.EnvironmentName, "environment", "", "Environment Name")
	f.StringVar(&o.Partition, "partition", "netic", fmt.Sprintf("Partition. One of (%s)", strings.Join(types.AllPartitionsString(), "|")))
	f.StringVar(&o.Region, "region", "dk-north", "Region. Depends on the partition.")
	f.StringVar(&o.SubscriptionID, "subscription", "", "Subscription ID")
	f.StringVar(&o.InfrastructureProvider, "infrastructure-provider", "netic", fmt.Sprintf("Infrastructure Provider. One of (%s)", strings.Join(types.AllInfrastructureProvidersString(), "|")))
	f.StringVar(&o.ResilienceZone, "resilience-zone", "netic", fmt.Sprintf("Resilience Zone. Should be one of (%s)", strings.Join(types.AllResilienceZonesString(), "|")))
	f.BoolVar(&o.HasTechnicalOperations, "has-to", true, "Technical Operations")
	f.BoolVar(&o.HasTechnicalManagement, "has-tm", true, "Technical Management")
	f.BoolVar(&o.HasApplicationOperations, "has-ao", false, "Application Operations")
	f.BoolVar(&o.HasApplicationManagement, "has-am", false, "Application Management")
	f.BoolVar(&o.HasCustomOperations, "has-co", false, "Custom Operations")
	f.StringVar(&o.CustomOperationsURL, "co-url", "", "Custom Operations URL")

	c.Flags().SortFlags = false
	c.MarkFlagRequired("name")            //nolint:errcheck
	c.MarkFlagRequired("provider")        //nolint:errcheck
	c.MarkFlagRequired("environment")     //nolint:errcheck
	c.MarkFlagRequired("subscription")    //nolint:errcheck
	c.MarkFlagRequired("resilience-zone") //nolint:errcheck
	c.MarkFlagsRequiredTogether("has-co", "co-url")

	return nil
}

func (o *createClusterOptions) Complete(_ context.Context, _ *ic.Context) error {
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

func (o *createClusterOptions) Validate(ctx context.Context, ac *ic.Context) error {
	p, ok := types.ParsePartition(o.Partition)
	if !ok {
		return &cmd.InvalidArgumentError{
			Flag:  "partition",
			Val:   o.Partition,
			OneOf: types.AllPartitionsString(),
		}
	}
	r, ok := types.ParseRegion(o.Region)
	if !ok {
		return &cmd.InvalidArgumentError{
			Flag:     "region",
			Val:      o.Region,
			SeeOther: "get regions",
		}
	}
	if !types.HasRegion(p, r) {
		return &cmd.InvalidArgumentError{
			Flag:     "region",
			Val:      o.Region,
			SeeOther: fmt.Sprintf("get regions --partition %s", o.Partition),
		}
	}
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
			Flag: "name",
			Val:  o.Name,
		},
		{
			Flag: "provider",
			Val:  o.ProviderName,
		},
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
		if !validation.IsDNSRFC1035Label(v.Val) {
			return &cmd.InvalidArgumentError{
				Flag:    v.Flag,
				Val:     v.Val,
				Context: "must be an RFC1035 DNS label",
			}
		}
	}
	if !slices.Contains(types.AllInfrastructureProvidersString(), o.InfrastructureProvider) {
		return &cmd.InvalidArgumentError{
			Flag:  "infrastructure-provider",
			Val:   o.InfrastructureProvider,
			OneOf: types.AllInfrastructureProvidersString(),
		}
	}
	if !slices.Contains(types.AllResilienceZonesString(), o.ResilienceZone) {
		ac.EC.Logger.WarnContext(ctx, fmt.Sprintf("Non-standard resilience zone used: %s", o.ResilienceZone))
	}
	if !validation.IsPrintableASCII(o.SubscriptionID) || len(o.SubscriptionID) < 5 {
		return &cmd.InvalidArgumentError{
			Flag:    "subscription",
			Val:     o.SubscriptionID,
			Context: "must be an ASCII string of minimum 5 characters length",
		}
	}
	return nil
}

func (o *createClusterOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("Clusters")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	var result *cluster.CreateClusterResult
	spinnerText := fmt.Sprintf("Creating cluster %s", o.Name)
	if err := ui.Spin(ac.EC.Spinner, spinnerText, func(_ ui.Spinner) error {
		in := cluster.CreateClusterInput{
			Logger:                   logger,
			APIClient:                ac.APIClient,
			Name:                     o.Name,
			Description:              o.Description,
			EnvironmentName:          o.EnvironmentName,
			Provider:                 o.ProviderName,
			Partition:                o.Partition,
			Region:                   o.Region,
			ResilienceZone:           o.ResilienceZone,
			SubscriptionID:           o.SubscriptionID,
			InfrastructureProvider:   o.InfrastructureProvider,
			HasTechnicalOperations:   o.HasTechnicalOperations,
			HasTechnicalManagement:   o.HasTechnicalManagement,
			HasApplicationOperations: o.HasApplicationOperations,
			HasApplicationManagement: o.HasApplicationManagement,
			HasCustomOperations:      o.HasCustomOperations,
			CustomOperationsURL:      o.CustomOperationsURL,
		}
		result, err = cluster.CreateCluster(ctx, in)
		return err
	}); err != nil {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Creating cluster",
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
