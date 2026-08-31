package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/fetcher"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate CRD schemas for breaking changes and compatibility issues.",
}

var schemaValidateCmd = &cobra.Command{
	Use:   "schema",
	Short: "Validate schema compatibility between CRD versions.",
	RunE:  runSchemaValidation,
}

var sampleValidateCmd = &cobra.Command{
	Use:          "sample",
	Short:        "Validate a sample YAML against the schema of a CRD.",
	RunE:         runSampleValidation,
	SilenceUsage: true,
}

type validateArgs struct {
	fromVersion    string
	toVersion      string
	outputFormat   string
	failOnBreaking bool
	sample         string
	ignoreErrors   []string
}

var valArgs = &validateArgs{}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.AddCommand(schemaValidateCmd)
	validateCmd.AddCommand(sampleValidateCmd)

	// Inherit persistent flags from generateCmd to access CRD sources
	validateCmd.PersistentFlags().AddFlagSet(generateCmd.PersistentFlags())

	f := schemaValidateCmd.Flags()
	f.StringVar(&valArgs.fromVersion, "from", "", "Source version to compare from (e.g., v1alpha1)")
	f.StringVar(&valArgs.toVersion, "to", "", "Target version to compare to (e.g., v1beta1)")
	f.StringVarP(&valArgs.outputFormat, "output", "o", "text", "Output format: text, json, yaml")
	f.BoolVar(&valArgs.failOnBreaking, "fail-on-breaking", false, "Exit with non-zero code if breaking changes detected")

	sf := sampleValidateCmd.Flags()
	sf.StringVarP(&valArgs.sample, "sample", "s", "", "The sample YAML file to validate against the CRD.")
	sf.StringSliceVar(&valArgs.ignoreErrors, "ignore-errors", nil, "Validation errors containing any of these substrings are ignored.")

	if err := sampleValidateCmd.MarkFlagRequired("sample"); err != nil {
		panic(err)
	}
}

// crdContent returns the raw CRD document for the configured input source.
func crdContent(args *rootArgs) ([]byte, error) {
	switch {
	case args.fileLocation != "":
		content, err := os.ReadFile(filepath.Clean(args.fileLocation))
		if err != nil {
			return nil, fmt.Errorf("failed to read CRD file: %w", err)
		}

		return content, nil
	case args.url != "":
		client := http.DefaultClient
		client.Timeout = timeout * time.Second

		content, err := fetcher.NewFetcher(client, args.username, args.password, args.token).Fetch(args.url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch CRD content: %w", err)
		}

		return content, nil
	default:
		return nil, errors.New("one of the flags (crd, url) must be set to validate a sample against")
	}
}

func runSampleValidation(cmd *cobra.Command, _ []string) error {
	crd, err := crdContent(args)
	if err != nil {
		return err
	}

	sample, err := os.ReadFile(filepath.Clean(valArgs.sample))
	if err != nil {
		return fmt.Errorf("failed to read sample file: %w", err)
	}

	if err := matches.Validate(crd, sample, valArgs.ignoreErrors); err != nil {
		return fmt.Errorf("sample %s is not valid: %w", valArgs.sample, err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "sample %s is valid\n", valArgs.sample)

	return nil
}

func runSchemaValidation(cmd *cobra.Command, _ []string) error {
	handler, err := constructHandler(args)
	if err != nil {
		return fmt.Errorf("failed to get handler: %w", err)
	}

	crds, err := handler.CRDs()
	if err != nil {
		return fmt.Errorf("failed to get CRDs: %w", err)
	}

	if len(crds) == 0 {
		return errors.New("no CRDs found")
	}

	validator := pkg.NewSchemaValidator()

	for _, crd := range crds {
		report, err := validator.ValidateVersions(crd, valArgs.fromVersion, valArgs.toVersion)
		if err != nil {
			return fmt.Errorf("failed to validate CRD %s: %w", crd.Kind, err)
		}

		if err := outputValidationReport(report, valArgs.outputFormat); err != nil {
			return fmt.Errorf("failed to output validation report: %w", err)
		}

		if valArgs.failOnBreaking && report.HasBreakingChanges() {
			os.Exit(1)
		}
	}

	return nil
}

func outputValidationReport(report *pkg.ValidationReport, format string) error {
	switch format {
	case "json":
		return report.OutputJSON(os.Stdout)
	case "yaml":
		return report.OutputYAML(os.Stdout)
	default:
		return report.OutputText(os.Stdout)
	}
}
