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
}
