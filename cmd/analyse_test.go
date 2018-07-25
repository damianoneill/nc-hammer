package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/damianoneill/nc-hammer/cmd"
	"github.com/damianoneill/nc-hammer/result"
	. "github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat"
)

// Mock testsuites used to populate []result.NetconfResult used in tests below.
var (
	mockCmd       *cobra.Command = &cobra.Command{}
	mockTestSuite                = TestSuite{File: "testdata/emptytestsuite.yml"}

	mts1 = result.NetconfResult{Client: 5, SessionID: 2318, Hostname: "10.0.0.1", Operation: "edit-config", When: 55282, Err: "", Latency: 288}
	mts2 = result.NetconfResult{Client: 6, SessionID: 859, Hostname: "10.0.0.2", Operation: "get-config", When: 55943, Err: "", Latency: 176}
	mts3 = result.NetconfResult{Client: 4, SessionID: 601, Hostname: "10.0.0.3", Operation: "get", When: 9840, Err: "", Latency: 3320}
	mts4 = result.NetconfResult{Client: 4, SessionID: 2322, Hostname: "10.0.0.1", Operation: "get", When: 56967, Err: "", Latency: 420}
	mts5 = result.NetconfResult{Client: 4, SessionID: 860, Hostname: "10.0.0.2", Operation: "kill-session", When: 0, Err: "kill-session is not a supported operation", Latency: 0}
	mts6 = result.NetconfResult{Client: 1, SessionID: 80, Hostname: "10.0.0.3", Operation: "close-session", When: 0, Err: "close-session is not a supported operation", Latency: 0}
	mts7 = result.NetconfResult{Client: 1, SessionID: 80, Hostname: "10.0.0.3", Operation: "-----", When: 0, Err: "----- ----- is not a supported operation", Latency: 0}
)

func redirectOutput(mockResults []result.NetconfResult) (one string, two string) {

	var fLOG, fCON string

	// Redirect StdErr to buffer
	var logOut = new(bytes.Buffer)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(logOut)

	// Redirect StdOut to buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AnalyseResults(mockCmd, &mockTestSuite, mockResults)

	// copy Stdout to buffer in a separate goroutine so printing can't block indefinitely
	out := make(chan string)
	go func() {
		defer r.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		out <- buf.String()
		buf.Reset()
	}()

	w.Close()
	os.Stdout = old // restore original
	consoleOut := <-out

	// have logOut and consoleOut
	re := regexp.MustCompile(`\r?\n`)
	fLOG = strings.Trim(re.ReplaceAllString(logOut.String(), " "), " ")

	// remove formating from table returned by AnalayseResults
	removeWhtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`) // remove whitespace outside required string
	fCON = removeWhtsp.ReplaceAllString(consoleOut, "")
	removeWhtsp = regexp.MustCompile(`[\s\p{Zs}]{2,}`) // remove whitespace inside required string
	fCON = removeWhtsp.ReplaceAllString(fCON, " ")

	return fCON, fLOG
}
func TestSortResults(t *testing.T) {

	testSort := func(t *testing.T, unsortedSlice, expected []result.NetconfResult) {
		t.Helper()

		SortResults(unsortedSlice)
		actual := unsortedSlice

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("actual %v, expected %v", actual, expected)
		}
	}

	t.Run("Sort by Hostname", func(t *testing.T) {
		unsortedSlice := []result.NetconfResult{mts3, mts1, mts2}
		expected := []result.NetconfResult{mts1, mts2, mts3}

		testSort(t, unsortedSlice, expected)
	})

	t.Run("Sort by Operation", func(t *testing.T) {
		unsortedSlice := []result.NetconfResult{mts3, mts4, mts2, mts5, mts1}
		expected := []result.NetconfResult{mts1, mts4, mts2, mts5, mts3}

		testSort(t, unsortedSlice, expected)
	})
}

func TestOrderAndExcludeErrValues(t *testing.T) {

	var mockLatencies = make(map[string]map[string][]float64)

	testOrderExclude := func(t *testing.T, mockResults []result.NetconfResult, expected int) {

		t.Helper()

		actual := OrderAndExcludeErrValues(mockResults, mockLatencies)

		if actual != expected {
			t.Errorf("actual %v, expected %v", expected, actual)
		}
	}

	t.Run("Exclude zero errors", func(t *testing.T) {
		mockResults := []result.NetconfResult{mts1, mts2, mts3, mts4}
		testOrderExclude(t, mockResults, 0)
	})

	t.Run("Exclude many errors", func(t *testing.T) {
		mockResults := []result.NetconfResult{mts5, mts6, mts7}
		testOrderExclude(t, mockResults, 3)
	})
}

// AnalyseResults prints to both Stdout and StdErr, so both must be captured in
// order to test for correct output.
func TestAnalyseResults(t *testing.T) {

	var mockResults = []result.NetconfResult{mts1, mts2, mts3}
	var mockLatencies = make(map[string]map[string][]float64)

	expectedStdout, expectedStderr := redirectOutput(mockResults)

	testAnalyse := func(t *testing.T, actual, expected string) {
		t.Helper()
		assert.Equal(t, actual, expected)
	}

	t.Run("Check correct output to Stderr", func(t *testing.T) {

		var logBuffer bytes.Buffer
		logBuffer.WriteString("Testsuite executed at " + strings.Split(mockTestSuite.File, string(filepath.Separator))[1] + " Suite defined the following hosts: ")
		logBuffer.WriteString("[")

		for _, config := range mockTestSuite.Configs {
			logBuffer.WriteString(config.Hostname + " ")
		}
		logBuffer.WriteString("] ")

		errCount := OrderAndExcludeErrValues(mockResults, mockLatencies)

		var when float64
		for _, result := range mockResults {
			if result.When > when {
				when = result.When
			}
		}
		executionTime := time.Duration(when) * time.Millisecond
		logBuffer.WriteString(strconv.Itoa(mockTestSuite.Clients) + " client(s) started, " + strconv.Itoa(mockTestSuite.Iterations) + " iterations per client, " + strconv.Itoa(mockTestSuite.Rampup) + " seconds wait between starting each client ")
		logBuffer.WriteString(" Total execution time: " + executionTime.String() + ", Suite execution contained " + strconv.Itoa(errCount) + " errors")

		actual := strings.Trim(logBuffer.String(), " ")

		testAnalyse(t, actual, expectedStderr)
	})

	t.Run("Check for correct output to Stdout -- flags not set", func(t *testing.T) {

		var consoleBuffer bytes.Buffer
		consoleBuffer.WriteString("HOST OPERATION REUSE CONNECTION REQUESTS TPS MEAN VARIANCE STD DEVIATION ")

		keys := SortLatencies(mockLatencies)
		for _, k := range keys {
			host := k
			operations := mockLatencies[k]
			for operation, mockLatencies := range operations {
				mean := stat.Mean(mockLatencies, nil)
				tps := 1000 / mean
				variance := stat.Variance(mockLatencies, nil)
				stddev := math.Sqrt(variance)
				consoleBuffer.WriteString(host + " " + operation + " " + strconv.FormatBool(mockTestSuite.Configs.IsReuseConnection(host)) + " " + strconv.Itoa(len(mockLatencies)) + " " + fmt.Sprintf("%.2f", tps) + " " + fmt.Sprintf("%.2f", mean) + " " + fmt.Sprintf("%.2f", variance) + " " + fmt.Sprintf("%.2f", stddev) + " ")
			}
		}

		actual := strings.Trim(consoleBuffer.String(), " ")

		testAnalyse(t, actual, expectedStdout)
	})
}

func TestAnalyseResultsWithFlags(t *testing.T) {

	var mockResults = []result.NetconfResult{mts1, mts2, mts3}

	testAnalyseWithFlag := func(t *testing.T) {
		t.Helper()

		redirectOutput(mockResults)

		// reset flags after each test
		mockCmd.ResetFlags()
	}

	t.Run("Operation flag set", func(t *testing.T) {
		var flag = "operation"
		mockCmd.Flags().StringVar(&flag, "operation", "edit-config", "")

		testAnalyseWithFlag(t)
	})

	t.Run("Hostname flag set", func(t *testing.T) {
		var flag = "hostname"
		mockCmd.Flags().StringVar(&flag, "hostname", "10.0.0.2", "")

		testAnalyseWithFlag(t)
	})
}

func TestAnalyseCmdArgs(t *testing.T) {

	var testCmd = AnalyseCmd
	var tempCmd = &cobra.Command{}

	testArgs := func(t *testing.T, args []string, expected error) {
		t.Helper()

		actual := testCmd.Args(tempCmd, args) // args = 1 or != 1
		assert.Equal(t, actual, expected)
	}

	t.Run("args == 1", func(t *testing.T) {
		var mockArgs = []string{"run"}

		testArgs(t, mockArgs, nil)
	})

	t.Run("args != 1", func(t *testing.T) {
		var mockArgs = []string{"run", "error", "kill"}
		expected := errors.New("analyse command requires a test results directory as an argument")

		testArgs(t, mockArgs, expected)
	})
}

func TestAnalyseCmdRun(t *testing.T) {

	tests := []struct {
		name     string
		testCmd  *cobra.Command
		testArgs []string
	}{
		// TODO: add more use cases
		{name: "single valid yaml file", testCmd: &cobra.Command{}, testArgs: []string{"../suite/testdata/results_test/2018-07-18-19-56-01/"}},
	}

	for _, tt := range tests {
		// Run test as subprocess when environment variable is set as 1
		if os.Getenv("RUN_SUBPROCESS") == "1" {
			AnalyseCmd.Run(tt.testCmd, tt.testArgs)
			return
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestAnalyseRun") // create new process to run test
		cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")          // set environmental variable
		err := cmd.Run()                                            // run
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {     // check exit status of test subprocess
			t.Fatalf("Program failed to load file -- os.Exit(1)")
		}
	}
}
