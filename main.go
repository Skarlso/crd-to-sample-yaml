package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: [yaml-from-crd] <path-to-crd>")
		os.Exit(1)
	}
	path := args[1]
	if _, err := os.Stat(path); os.IsNotExist(err) {
		printAndQuit("file under '%s' does not exist\n", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		printAndQuit("failed to read file: ", err)
	}

	crd := &v1beta1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(content, crd); err != nil {
		printAndQuit("failed to unmarshal into custom resource definition")
	}

	// TODO: Each version should have its own file.
	dir := filepath.Dir(path)
	outputLocation := filepath.Join(dir, "output.yaml")
	outputFile, err := os.Create(outputLocation)
	if err != nil {
		printAndQuit("failed to create file at: '%s': %v", outputLocation, err)
	}
	//
	//fmt.Println(crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties)

	// parseHeader -- parses apiVersion, kind, metadata
	parseProperties(crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties, outputFile, 0)
	outputFile.Close()
}

// TODO: somehow sort the output otherwise it's constantly changing.
func parseProperties(properties map[string]v1beta1.JSONSchemaProps, file *os.File, indent int) {
	for k, v := range properties {
		if _, err := file.WriteString(fmt.Sprintf("%s%s:\n", strings.Repeat(" ", indent), k)); err != nil {
			file.Close()
			printAndQuit("failed to write k to file '%s'", k)
		}
		if len(v.Properties) > 0 {
			parseProperties(v.Properties, file, indent+2)
		}
	}
}

func printAndQuit(msg string, args ...any) {
	fmt.Printf(msg, args...)
	os.Exit(1)
}
