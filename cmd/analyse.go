package cmd

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/olekukonko/tablewriter"
	"gonum.org/v1/gonum/stat"

	"github.com/spf13/cobra"
)

// AnalyseCmd represents the analyse command
var AnalyseCmd = &cobra.Command{
	Use:   "analyse <results file>",
	Short: "Analyse the output of a Test Suite run",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("analyse command requires a test results directory as an argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if results, ts, err := result.UnarchiveResults(args[0]); err != nil {
			log.Fatalf("Problem with loading result information: %v ", err)
		} else {
			AnalyseResults(cmd, ts, results)
		}
	},
}

func AnalyseResults(cmd *cobra.Command, ts *suite.TestSuite, results []result.NetconfResult) {
	log.Println("")
	log.Printf("Testsuite executed at %v\n", strings.Split(ts.File, string(filepath.Separator))[1])
	var hosts []string
	for _, config := range ts.Configs {
		hosts = append(hosts, config.Hostname)
	}
	log.Printf("Suite defined the following hosts: %v\n", hosts)

	latencies := make(map[string]map[string][]float64)
	errCount := OrderAndExcludeErrValues(results, latencies)

	// get the largest when time from the results, this is the last action to run
	var when float64
	for _, result := range results {
		if result.When > when {
			when = result.When
		}
	}
	executionTime := time.Duration(when) * time.Millisecond

	log.Printf("%d client(s) started, %d iterations per client, %d seconds wait between starting each client\n", ts.Clients, ts.Iterations, ts.Rampup)
	log.Printf("\nTotal execution time: %v, Suite execution contained %v errors", executionTime, errCount)

	log.Println("")

	/*
		//nolint
		op, _ := cmd.Flags().GetString("operation")
		//nolint
		hostname, _ := cmd.Flags().GetString("hostname")
	*/
	op := ""
	hostname := ""

	data := [][]string{}
	for host, operations := range latencies {
		for operation, latencies := range operations {
			if op != "" && op != operation {
				continue
			}
			if hostname != "" && hostname != host {
				continue
			}
			mean := stat.Mean(latencies, nil)
			tps := 1000 / mean
			variance := stat.Variance(latencies, nil)
			stddev := math.Sqrt(variance)
			data = append(data, []string{host, operation, strconv.FormatBool(ts.Configs.IsReuseConnection(host)), strconv.Itoa(len(latencies)), fmt.Sprintf("%.2f", tps), fmt.Sprintf("%.2f", mean), fmt.Sprintf("%.2f", variance), fmt.Sprintf("%.2f", stddev)})
		}
	}

	var table = tablewriter.NewWriter(os.Stdout)
	renderTable(table, []string{"Host", "Operation", "Reuse Connection", "Requests", "TPS", "Mean", "Variance", "Std Deviation"}, &data)
	table.Render()
}

func OrderAndExcludeErrValues(results []result.NetconfResult, latencies map[string]map[string][]float64) int {
	SortResults(results)

	var errCount int
	for _, result := range results {
		if latencies[result.Hostname] == nil {
			latencies[result.Hostname] = make(map[string][]float64)
		}
		// only add latency if its not in error
		if result.Err != "" {
			errCount++
		} else {
			latencies[result.Hostname][result.Operation] = append(latencies[result.Hostname][result.Operation], result.Latency)
		}
	}
	return errCount
}

func SortResults(results []result.NetconfResult) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Hostname != results[j].Hostname {
			return results[i].Hostname < results[j].Hostname
		}
		return results[i].Operation < results[j].Operation
	})
}

func renderTable(table *tablewriter.Table, header []string, data *[][]string) {
	table.SetHeader(header)
	table.SetRowLine(true)

	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	table.AppendBulk(*data)
}

func init() {
	RootCmd.AddCommand(AnalyseCmd)
	AnalyseCmd.Flags().StringP("operation", "o", "", "filter based on operation type; get, get-config or edit-config")
	AnalyseCmd.Flags().StringP("hostname", "", "", "filter based on host name or ip")
}
