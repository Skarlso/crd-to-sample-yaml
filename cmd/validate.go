package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
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

type validateArgs struct {
	fromVersion    string
	toVersion      string
	outputFormat   string
	failOnBreaking bool
}

var valArgs = &validateArgs{}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.AddCommand(schemaValidateCmd)

	// Inherit persistent flags from generateCmd to access CRD sources
	validateCmd.PersistentFlags().AddFlagSet(generateCmd.PersistentFlags())

	f := schemaValidateCmd.Flags()
	f.StringVar(&valArgs.fromVersion, "from", "", "Source version to compare from (e.g., v1alpha1)")
	f.StringVar(&valArgs.toVersion, "to", "", "Target version to compare to (e.g., v1beta1)")
	f.StringVarP(&valArgs.outputFormat, "output", "o", "text", "Output format: text, json, yaml")
	f.BoolVar(&valArgs.failOnBreaking, "fail-on-breaking", false, "Exit with non-zero code if breaking changes detected")
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
