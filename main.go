package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	parseProperties(crd.Spec.Group, crd.Spec.Versions[0].Name, crd.Spec.Names.Kind, crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties, outputFile, 0)
	outputFile.Close()
}

// TODO: somehow sort the output otherwise it's constantly changing.
func parseProperties(group, version, kind string, properties map[string]v1beta1.JSONSchemaProps, file *os.File, indent int) {
	var sortedKeys []string
	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		writeOrFail(file, fmt.Sprintf("%s%s:", strings.Repeat(" ", indent), k))
		if len(properties[k].Properties) == 0 {
			if k == "apiVersion" {
				writeOrFail(file, fmt.Sprintf(" %s/%s\n", group, version))
				continue
			}
			if k == "kind" {
				writeOrFail(file, fmt.Sprintf(" %s\n", kind))
				continue
			}
			result := outputValueType(properties[k])
			writeOrFail(file, fmt.Sprintf(" %s\n", result))
		} else if len(properties[k].Properties) > 0 {
			writeOrFail(file, "\n")
			parseProperties(group, version, kind, properties[k].Properties, file, indent+2)
		}
	}
}

func writeOrFail(file *os.File, s string) {
	if _, err := file.WriteString(s); err != nil {
		file.Close()
		printAndQuit("failed to write '%s' to file '%s'", s, file.Name())
	}
}

func outputValueType(v v1beta1.JSONSchemaProps) string {
	switch v.Type {
	case "string":
		return "string"
	case "boolean":
		return "true"
	case "object":
		return "{}"
	case "array":
		return "[]"
	}
	return v.Type
}

func printAndQuit(msg string, args ...any) {
	fmt.Printf(msg, args...)
	os.Exit(1)
}
