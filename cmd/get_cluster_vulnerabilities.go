package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neticdk-k8s/ic/internal/errors"
	"github.com/neticdk-k8s/ic/internal/ic"
	"github.com/neticdk-k8s/ic/internal/usecases/cluster"
	"github.com/neticdk/go-common/pkg/cli/cmd"
	"github.com/neticdk/go-common/pkg/cli/ui"
	"github.com/spf13/cobra"
)

const (
	outputDirPermissions  = 0o750
	outputFilePermissions = 0o640
)

const getClusterVulnerabilitiesLongDesc = `Get vulnerabilities detected in images running in one or more clusters.

Output formats: plain, table, json, csv

The csv format writes one row per vulnerable image with the columns:
Cluster, CVE, Severity, Package Name, Package Version, Fix Versions, Image, Categories

With --output-dir one file per cluster is written as cve_<cluster-id>.<format>
instead of writing all clusters to stdout.`

const getClusterVulnerabilitiesExample = `
# get vulnerabilities for a cluster
ic get cluster-vulnerabilities my-cluster.my-provider

# write all vulnerabilities for two clusters to a single csv file
ic -o csv get cluster-vulnerabilities my-cluster.my-provider other-cluster.my-provider > cves.csv

# write one csv file per cluster for all clusters
ic get clusters --no-headers | awk '{print $2}' | xargs ic -o csv get cluster-vulnerabilities --output-dir ./cves`

func getClusterVulnerabilitiesCmd(ac *ic.Context) *cobra.Command {
	o := &getClusterVulnerabilitiesOptions{}
	c := cmd.NewSubCommand("cluster-vulnerabilities", o, ac).
		WithShortDesc("Get vulnerabilities for one or more clusters").
		WithLongDesc(getClusterVulnerabilitiesLongDesc).
		WithExample(getClusterVulnerabilitiesExample).
		WithGroupID(groupCluster).
		WithMinArgs(1).
		Build()
	c.Use = "cluster-vulnerabilities CLUSTER-ID..."
	c.Aliases = []string{"vulnerabilities", "cves"}

	return c
}

type getClusterVulnerabilitiesOptions struct {
	clusterIDs []string
	outputDir  string
}

func (o *getClusterVulnerabilitiesOptions) SetupFlags(_ context.Context, ac *ic.Context) error {
	f := ac.EC.Command.Flags()
	f.StringVar(&o.outputDir, "output-dir", "", "Write one file per cluster to this directory instead of stdout")
	return nil
}

func (o *getClusterVulnerabilitiesOptions) Complete(_ context.Context, ac *ic.Context) error {
	o.clusterIDs = ac.EC.CommandArgs
	return nil
}

func (o *getClusterVulnerabilitiesOptions) Validate(_ context.Context, ac *ic.Context) error {
	switch ac.EC.PFlags.OutputFormat {
	case cluster.FormatPlain, cluster.FormatTable, cluster.FormatJson, cluster.FormatCSV:
	default:
		return fmt.Errorf("unsupported output format: %s", ac.EC.PFlags.OutputFormat)
	}
	if o.outputDir != "" {
		if err := os.MkdirAll(o.outputDir, outputDirPermissions); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}
	return nil
}

func (o *getClusterVulnerabilitiesOptions) Run(ctx context.Context, ac *ic.Context) error {
	logger := ac.EC.Logger.WithGroup("ClusterVulnerabilities")
	ac.Authenticator.SetLogger(logger)

	_, err := doLogin(ctx, ac)
	if err != nil {
		return err
	}

	// A failing cluster is reported and skipped so one missing cluster does not abort a long run.
	var results []*cluster.ClusterVulnerabilities
	var failed []string
	for _, clusterID := range o.clusterIDs {
		var result *cluster.ListClusterVulnerabilitiesResult
		spinnerText := fmt.Sprintf("Getting vulnerabilities for cluster %q", clusterID)
		if err := ui.Spin(ac.EC.Spinner, spinnerText, func(_ ui.Spinner) error {
			in := cluster.ListClusterVulnerabilitiesInput{
				Logger:    logger,
				APIClient: ac.APIClient,
				ClusterID: clusterID,
			}
			result, err = cluster.ListClusterVulnerabilities(ctx, in)
			return err
		}); err != nil {
			ui.Warning.Printfln("%s: %v", clusterID, err)
			failed = append(failed, clusterID)
			continue
		}
		if result.Problem != nil {
			ui.Warning.Printfln("%v", &errors.ProblemError{Title: clusterID, Problem: result.Problem})
			failed = append(failed, clusterID)
			continue
		}

		if o.outputDir != "" {
			if err := o.writeFile(result.Vulnerabilities, ac.EC.PFlags.OutputFormat, ac.EC.PFlags.NoHeaders); err != nil {
				return ac.EC.ErrorHandler.NewGeneralError(
					"Failed to write output",
					"See details for more information",
					err,
					0,
				)
			}
			continue
		}
		results = append(results, result.Vulnerabilities)
	}

	if o.outputDir == "" {
		r := cluster.NewClusterVulnerabilitiesRenderer(results, ac.EC.Stdout, ac.EC.PFlags.NoHeaders)
		if err := r.Render(ac.EC.PFlags.OutputFormat); err != nil {
			return ac.EC.ErrorHandler.NewGeneralError(
				"Failed to render output",
				"See details for more information",
				err,
				0,
			)
		}
	}

	if len(failed) > 0 {
		return ac.EC.ErrorHandler.NewGeneralError(
			"Getting cluster vulnerabilities",
			fmt.Sprintf("Failed for %d of %d clusters: %v", len(failed), len(o.clusterIDs), failed),
			nil,
			0,
		)
	}

	return nil
}

func (o *getClusterVulnerabilitiesOptions) writeFile(v *cluster.ClusterVulnerabilities, format string, noHeaders bool) error {
	ext := format
	if format == cluster.FormatPlain || format == cluster.FormatTable {
		ext = "txt"
	}
	root, err := os.OpenRoot(o.outputDir)
	if err != nil {
		return fmt.Errorf("opening output directory: %w", err)
	}
	defer root.Close()
	filename := fmt.Sprintf("cve_%s.%s", v.ClusterID, ext)
	f, err := root.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, outputFilePermissions)
	if err != nil {
		return err
	}

	r := cluster.NewClusterVulnerabilitiesRenderer([]*cluster.ClusterVulnerabilities{v}, f, noHeaders)
	err = r.Render(format)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		ui.Info.Printfln("Wrote %s", filepath.Join(o.outputDir, filename))
	}
	return err
}
