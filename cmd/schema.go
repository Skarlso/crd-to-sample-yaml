package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/json"
)

// schemaCmd is a command that can generate json schemas.
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Simply generate a JSON schema from the CRD.",
	RunE:  runGenerateSchema,
}

type schemaCmdArgs struct {
	outputFolder string
}

var schemaArgs = &schemaCmdArgs{}

func init() {
	generateCmd.AddCommand(schemaCmd)
	f := schemaCmd.PersistentFlags()
	f.StringVarP(&schemaArgs.outputFolder, "output", "o", ".", "output location of the generated schema files")
}

func runGenerateSchema(_ *cobra.Command, _ []string) error {
	crdHandler, err := constructHandler(args)
	if err != nil {
		return err
	}

	// determine location of output
	if schemaArgs.outputFolder == "" {
		loc, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable location: %w", err)
		}

		schemaArgs.outputFolder = filepath.Dir(loc)
	}

	crds, err := crdHandler.CRDs()
	if err != nil {
		return fmt.Errorf("failed to load CRDs: %w", err)
	}

	for _, crd := range crds {
		for _, v := range crd.Spec.Versions {
			if v.Schema.OpenAPIV3Schema.ID == "" {
				v.Schema.OpenAPIV3Schema.ID = "https://crdtoyaml.com/" + crd.Spec.Names.Kind + "." + crd.Spec.Group + "." + v.Name + ".schema.json"
			}
			if v.Schema.OpenAPIV3Schema.Schema == "" {
				v.Schema.OpenAPIV3Schema.Schema = "https://json-schema.org/draft/2020-12/schema"
			}
			content, err := json.Marshal(v.Schema.OpenAPIV3Schema)
			if err != nil {
				return fmt.Errorf("failed to marshal schema: %w", err)
			}

			const perm = 0o600
			if err := os.WriteFile(filepath.Join(schemaArgs.outputFolder, crd.Spec.Names.Kind+"."+crd.Spec.Group+"."+v.Name+".schema.json"), content, perm); err != nil {
				return fmt.Errorf("failed to write schema: %w", err)
			}
		}
	}

	return nil
}
