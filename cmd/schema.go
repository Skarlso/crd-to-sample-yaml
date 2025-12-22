package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
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

type KindVersionGroup struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
	Group   string `json:"group"`
}

type Schema struct {
	*v1beta1.JSONSchemaProps `json:",inline"`

	KubernetesGroupVersionKindList []KindVersionGroup `json:"x-kubernetes-group-version-kind"`
}

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
		for _, v := range crd.Versions {
			if v.Schema.ID == "" {
				v.Schema.ID = "https://crdtoyaml.com/" + crd.Kind + "." + crd.Group + "." + v.Name + ".schema.json"
			}

			if v.Schema.Schema == "" {
				v.Schema.Schema = "https://json-schema.org/draft/2020-12/schema"
			}

			schema := Schema{
				JSONSchemaProps: v.Schema,
				KubernetesGroupVersionKindList: []KindVersionGroup{
					{
						Kind:    crd.Kind,
						Group:   crd.Group,
						Version: v.Name,
					},
				},
			}

			content, err := json.Marshal(schema)
			if err != nil {
				return fmt.Errorf("failed to marshal schema: %w", err)
			}

			const perm = 0o600
			if err := os.WriteFile(filepath.Join(schemaArgs.outputFolder, crd.Kind+"."+crd.Group+"."+v.Name+".schema.json"), content, perm); err != nil {
				return fmt.Errorf("failed to write schema: %w", err)
			}
		}
	}

	return nil
}
