package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/Skarlso/crd-to-sample-yaml/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	// serveCmd is root for various `generate ...` commands
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Create a webserver from a specific location.",
		RunE:  runServe,
	}
	address string
)

func init() {
	rootCmd.AddCommand(serveCmd)

	f := serveCmd.PersistentFlags()
	f.StringVarP(&address, "address", "a", ":9998", "The address under which the server is running")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Run service & server
	log.Println("starting to serve under: ", address)
	sv, err := pkg.NewServer(address)
	if err != nil {
		return err
	}
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return sv.Run()
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}
	return nil
}
