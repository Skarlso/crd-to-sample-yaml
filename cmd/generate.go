package cmd

import (
	"github.com/spf13/cobra"
)

type rootArgs struct {
	fileLocation   string
	folderLocation string
	url            string
	username       string
	password       string
	token          string
	output         string
	format         string
	stdOut         bool
	comments       bool
	minimal        bool
	skipRandom     bool
}

var (
	// generateCmd is root for various `generate ...` commands.
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Simply generate a CRD output.",
	}

	args = &rootArgs{}
)

func init() {
	rootCmd.AddCommand(generateCmd)
	// using persistent flags so all flags will be available for all sub commands.
	f := generateCmd.PersistentFlags()
	f.StringVarP(&args.fileLocation, "crd", "c", "", "The CRD file to generate a yaml from.")
	f.StringVarP(&args.folderLocation, "folder", "r", "", "A folder from which to parse a series of CRDs.")
	f.StringVarP(&args.url, "url", "u", "", "If provided, will use this URL to fetch CRD YAML content from.")
	f.StringVar(&args.username, "username", "", "Optional username to authenticate a URL.")
	f.StringVar(&args.password, "password", "", "Optional password to authenticate a URL.")
	f.StringVar(&args.token, "token", "", "A bearer token to authenticate a URL.")
	f.StringVarP(&args.output, "output", "o", "", "The location of the output file. Default is next to the CRD.")
	f.StringVarP(&args.format, "format", "f", FormatYAML, "The format in which to output. Default is YAML. Options are: yaml, html.")
	f.BoolVarP(&args.stdOut, "stdout", "s", false, "If set, it will output the generated content to stdout.")
	f.BoolVarP(&args.comments, "comments", "m", false, "If set, it will add descriptions as comments to each line where available.")
	f.BoolVarP(&args.minimal, "minimal", "l", false, "If set, only the minimal required example yaml is generated.")
	f.BoolVar(&args.skipRandom, "no-random", false, "Skip generating random values that satisfy the property patterns.")
}
