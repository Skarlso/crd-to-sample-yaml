package cmd

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

var (
	// testCmd is root for various `test ...` commands.
	testCmd = &cobra.Command{
		Use:   "test",
		Short: "Run a set of tests to check CRD schema stability.",
		Run:   runTest,
	}

	testArgs struct {
		update bool
	}
)

func init() {
	rootCmd.AddCommand(testCmd)

	f := testCmd.PersistentFlags()
	f.BoolVarP(&testArgs.update, "update", "u", false, "Update any existing snapshots.")
}

func runTest(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "test needs an argument where the tests are located at")

		os.Exit(1)
	}

	path := args[0]
	runner := tests.NewSuiteRunner(path, testArgs.update)
	outcome, err := runner.Run(cmd.Context())
	if err != nil {
		os.Exit(1)
	}

	if err := displayWarnings(outcome); err != nil {
		os.Exit(1)
	}
}

func displayWarnings(warnings []tests.Outcome) error {
	errs := 0

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"It", "Matcher", "Error", "Template"})
	rows := make([]table.Row, 0, len(warnings))
	for _, w := range warnings {
		if w.Error != nil {
			errs++
		}

		rows = append(rows, table.Row{
			w.Name, w.Matcher, w.Error, w.Template,
		})
	}
	t.AppendRows(rows)
	t.AppendSeparator()
	t.Render()

	fmt.Fprintf(os.Stdout, "\nTests total: %d, failed: %d, passed: %d\n", len(warnings), errs, len(warnings)-errs)

	if errs > 0 {
		return fmt.Errorf("%d test(s) failed", errs)
	}

	return nil
}
