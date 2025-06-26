package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/resiliencezone"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

func getResilienceZonesCmd(ac *ic.Context) *cobra.Command {
	o := &getResilienceZonesOptions{}
	c := cmd.NewSubCommand("resilience-zones", o, ac).
		WithShortDesc("List resilience zones").
		WithGroupID(groupOther).
		Build()

	return c
}

type getResilienceZonesOptions struct{}

func (o *getResilienceZonesOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *getResilienceZonesOptions) Complete(_ context.Context, _ *ic.Context) error   { return nil }
func (o *getResilienceZonesOptions) Validate(_ context.Context, _ *ic.Context) error   { return nil }

func (o *getResilienceZonesOptions) Run(_ context.Context, ac *ic.Context) error {
	rzs := resiliencezone.ListResilienceZones()
	r := resiliencezone.NewResilienceZonesRenderer(rzs, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
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
