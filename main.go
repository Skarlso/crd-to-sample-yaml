package main

import (
	"log"

	"github.com/Skarlso/crd-to-sample-yaml/cmd"
)

// version is set during build time by GoReleaser.
var version string

func main() {
	if version != "" {
		cmd.Version = version
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to run command: ", err)
	}
}
