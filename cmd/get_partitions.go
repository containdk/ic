package cmd

import (
	"context"

	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/partition"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/spf13/cobra"
)

func getPartitionsCmd(ac *ic.Context) *cobra.Command {
	o := &getPartitionsOptions{}
	c := cmd.NewSubCommand("partitions", o, ac).
		WithShortDesc("List partitions").
		WithGroupID(groupOther).
		Build()

	return c
}

type getPartitionsOptions struct{}

func (o *getPartitionsOptions) SetupFlags(_ context.Context, _ *ic.Context) error { return nil }
func (o *getPartitionsOptions) Complete(_ context.Context, _ *ic.Context) error   { return nil }
func (o *getPartitionsOptions) Validate(_ context.Context, _ *ic.Context) error   { return nil }

func (o *getPartitionsOptions) Run(_ context.Context, ac *ic.Context) error {
	partitions := partition.ListPartitions()
	r := partition.NewPartitionsRenderer(partitions, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
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
