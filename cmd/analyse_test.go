package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/damianoneill/nc-hammer/result"
	. "github.com/damianoneill/nc-hammer/suite"
	. "github.com/nc-hammer/cmd"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat"
)

var (
	ts1 = result.NetconfResult{5, 2318, "172.26.138.91", "edit-config", 55282, "", 288}
	ts2 = result.NetconfResult{6, 859, "172.26.138.92", "get-config", 55943, "", 176}
	ts3 = result.NetconfResult{4, 601, "172.26.138.93", "get", 59840, "", 3320}
	ts4 = result.NetconfResult{4, 2322, "172.26.138.91", "get", 56967, "", 420}
	ts5 = result.NetconfResult{4, 860, "172.26.138.92", "kill-session", 0, "kill-session is not a supported operation", 0}
)

func TestSortResults(t *testing.T) {

	testSort := func(t *testing.T, testArray []result.NetconfResult, want []result.NetconfResult) {
		t.Helper()

		SortResults(testArray)
		got := testArray

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	t.Run("sort by Hostname", func(t *testing.T) {
		testArray := []result.NetconfResult{ts3, ts1, ts2}
		want := []result.NetconfResult{ts1, ts2, ts3}

		testSort(t, testArray, want)
	})

	t.Run("sort by Operation", func(t *testing.T) {
		testArray := []result.NetconfResult{ts3, ts4, ts2, ts5, ts1}
		want := []result.NetconfResult{ts1, ts4, ts2, ts5, ts3}

		testSort(t, testArray, want)
	})

}

func TestOrderAndExcludeErrValues(t *testing.T) {
	testResults := []result.NetconfResult{ts1, ts2, ts3, ts4, ts5}
	testLatencies := make(map[string]map[string][]float64)

	got := OrderAndExcludeErrValues(testResults, testLatencies)

	if got != 1 {
		t.Errorf("got %v, want %v", 1, got)
	}
}

func TestAnalyseResults(t *testing.T) {

	// capture log printout
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamps
	var logOutput = new(bytes.Buffer)
	log.SetOutput(logOutput)

	var mockCmd *cobra.Command
	var testArray = []result.NetconfResult{ts1, ts2, ts3}
	mockTs := TestSuite{}
	mockTs.File = "testdata/emptytestsuite.yml"

	// capture console printout
	old := os.Stdout // keep backup of the real stdout
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout = w
	outC := make(chan string)

	go func() { // copy the output in a separate goroutine so printing can't block indefinitely
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	AnalyseResults(mockCmd, &mockTs, testArray)

	// back to normal state
	w.Close()
	os.Stdout = old
	consoleOutput := <-outC

	// fmt.Println("\nLOG\n-------------")
	// fmt.Print(logOutput)
	// fmt.Println("\nCONSOLE\n-------------")
	// fmt.Println(consoleOutput)
	//
	//

	// Create log test string
	var testLogOutput bytes.Buffer

	testLogOutput.WriteString("\nTestsuite executed at ")
	filePath := mockTs.File
	filePath = strings.Split(filePath, string(filepath.Separator))[1]
	testLogOutput.WriteString(filePath + "\n")

	var hosts bytes.Buffer
	hosts.WriteString("[")
	for _, config := range mockTs.Configs {
		hosts.WriteString(config.Hostname + " ")
	}
	hosts.WriteString("]")
	testLogOutput.WriteString("Suite defined the following hosts: ")
	testLogOutput.WriteString(hosts.String() + "\n")

	latencies := make(map[string]map[string][]float64)
	errCount := OrderAndExcludeErrValues(testArray, latencies)

	var when float64
	for _, result := range testArray {
		if result.When > when {
			when = result.When
		}
	}
	executionTime := time.Duration(when) * time.Millisecond

	testLogOutput.WriteString(strconv.Itoa(mockTs.Clients) + " client(s) started, " + strconv.Itoa(mockTs.Iterations) +
		" iterations per client, " + strconv.Itoa(mockTs.Rampup) + " seconds wait between starting each client\n")
	testLogOutput.WriteString("\nTotal execution time: " + (executionTime).String() + ", Suite execution contained " +
		strconv.Itoa(errCount) + " errors\n\n")

	//fmt.Println("\nTEST\n-------------")
	//fmt.Print(testLogOutput.String())

	assert.Equal(t, testLogOutput.String(), logOutput.String(), "LOG: the two outputs do not match")

	//
	//

	// Create console test string
	op, hostname := "", ""

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
			data = append(data, []string{host, operation, strconv.FormatBool(mockTs.Configs.IsReuseConnection(host)), strconv.Itoa(len(latencies)), fmt.Sprintf("%.2f", tps), fmt.Sprintf("%.2f", mean), fmt.Sprintf("%.2f", variance), fmt.Sprintf("%.2f", stddev)})
		}
	}

	// remove formatting from table in console output
	re := regexp.MustCompile(`\s+`)
	consoleOutput = re.ReplaceAllString(consoleOutput, " ")

	var testConsoleOutput = bytes.Buffer{}
	testConsoleOutput.WriteString(" HOST OPERATION REUSE CONNECTION REQUESTS TPS MEAN VARIANCE STD DEVIATION ")
	for i, row := range data {
		for j, _ := range row {
			testConsoleOutput.WriteString(data[i][j] + " ")
		}
	}
	assert.Equal(t, testConsoleOutput.String(), consoleOutput, "CONSOLE: the two outputs do not match")

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
