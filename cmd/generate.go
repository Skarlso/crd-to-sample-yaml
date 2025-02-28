package cmd

import (
	"github.com/spf13/cobra"
)

type rootArgs struct {
	fileLocation       string
	folderLocation     string
	configFileLocation string
	url                string
	username           string
	password           string
	token              string
	tag                string
	caBundle           string
	privSSHKey         string
	useSSHAgent        bool
	gitURL             string
	stdin              bool
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
	f.BoolVarP(&args.stdin, "stdin", "i", false, "Take CRD content from stdin.")
	f.StringVarP(&args.fileLocation, "crd", "c", "", "The CRD file to generate a yaml from.")
	f.StringVarP(&args.folderLocation, "folder", "r", "", "A folder from which to parse a series of CRDs.")
	f.StringVarP(&args.url, "url", "u", "", "If provided, will use this URL to fetch CRD YAML content from.")
	f.StringVarP(&args.gitURL, "git-url", "g", "", "If provided, CRDs will be discovered using a git repository.")
	f.StringVar(&args.username, "username", "", "Optional username to authenticate a URL.")
	f.StringVar(&args.password, "password", "", "Optional password to authenticate a URL.")
	f.StringVar(&args.token, "token", "", "A bearer token to authenticate a URL.")
	f.StringVar(&args.configFileLocation, "config", "", "An optional configuration file that can define grouping data for various rendered crds.")
	f.StringVar(&args.tag, "tag", "", "The ref to check out. Default is head.")
	f.StringVar(&args.caBundle, "ca-bundle-file", "", "Additional certificate bundle to load. Should the name of the file.")
	f.StringVar(&args.privSSHKey, "private-ssh-key-file", "", "Private key to use for cloning. Should the name of the file.")
	f.BoolVar(&args.useSSHAgent, "ssh-agent", false, "If set, the configured SSH agent will be used to clone the repository..")
}
