package main

import (
	"log"

	"github.com/Skarlso/crd-to-sample-yaml/cmd"
	_ "github.com/Skarlso/crd-to-sample-yaml/pkg/matches/matchsnapshot"
	_ "github.com/Skarlso/crd-to-sample-yaml/pkg/matches/matchstring"
)

// version is set during build time by GoReleaser.
var version string

func main() {
	if version != "" {
		cmd.Version = version
	}

	err := cmd.Execute()
	if err != nil {
		log.Fatal("failed to run command: ", err)
	}
}
