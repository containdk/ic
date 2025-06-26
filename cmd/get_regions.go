package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/region"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/types"
	"github.com/spf13/cobra"
)

func getRegionsCmd(ac *ic.Context) *cobra.Command {
	o := &getRegionsOptions{}
	c := cmd.NewSubCommand("regions", o, ac).
		WithShortDesc("List regions").
		WithGroupID(groupOther).
		Build()

	return c
}

type getRegionsOptions struct {
	Partition string
}

func (o *getRegionsOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	c := ac.EC.Command
	f := c.Flags()

	f.StringVar(&o.Partition, "partition", "", fmt.Sprintf("Partition. One of (%s)", strings.Join(types.AllPartitionsString(), "|")))
	return nil
}

func (o *getRegionsOptions) Complete(_ context.Context, _ *ic.Context) error { return nil }
func (o *getRegionsOptions) Validate(_ context.Context, _ *ic.Context) error { return nil }

func (o *getRegionsOptions) Run(_ context.Context, ac *ic.Context) error {
	var (
		regions []string
		err     error
	)
	if o.Partition != "" {
		regions, err = region.ListRegionsForPartition(o.Partition)
		if err != nil {
			return ac.EC.ErrorHandler.NewGeneralError(
				"Getting regions",
				"See details for more information",
				err,
				0,
			)
		}
	} else {
		regions = region.ListRegions()
	}

	r := region.NewRegionsRenderer(regions, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
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
