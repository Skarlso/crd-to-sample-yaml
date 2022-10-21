package main

import (
	"fmt"

	"github.com/Skarlso/crd-to-sample-yaml/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("failed to run command: ", err)
	}
}
