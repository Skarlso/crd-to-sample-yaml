package main

import (
	"log"

	"github.com/Skarlso/crd-to-sample-yaml/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to run command: ", err)
	}
}
