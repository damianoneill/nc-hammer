package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// analyseErrorCmd represents the analyseError command
var analyseErrorCmd = &cobra.Command{
	Use:   "error",
	Short: "Analyse the errors of a Test Suite run",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("error command requires a test results directory as an argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if results, ts, err := result.UnarchiveResults(args[0]); err != nil {
			log.Fatalf("Problem with loading result information: %v ", err)
		} else {
			analyseErrors(cmd, ts, results)
		}
	},
}

func analyseErrors(cmd *cobra.Command, ts *suite.TestSuite, results []result.NetconfResult) {
	log.Println("")
	log.Printf("Testsuite executed at %v\n", strings.Split(ts.File, string(filepath.Separator))[1])

	sortResults(results)

	var errors [][]string
	for idx := range results {
		if results[idx].Err != "" {
			errors = append(errors, []string{results[idx].Hostname, results[idx].Operation, results[idx].Err})
		}
	}

	log.Printf("Total Number of Errors for suite: %d\n", len(errors))

	var table = tablewriter.NewWriter(os.Stdout)
	table.SetReflowDuringAutoWrap(true)
	table.SetColWidth(80)
	renderTable(table, []string{"Hostname", "Operation", "Error"}, &errors)

	table.Render()
}

func init() {
	analyseCmd.AddCommand(analyseErrorCmd)
}
