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

	"github.com/damianoneill/nc-hammer/result"
	. "github.com/damianoneill/nc-hammer/suite"
	. "github.com/nc-hammer/cmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat"
)

// Test variables used to populate []result.NetconfResult used throughout
var (
	ts1 = result.NetconfResult{5, 2318, "172.26.138.91", "edit-config", 55282, "", 288}
	ts2 = result.NetconfResult{6, 859, "172.26.138.92", "get-config", 55943, "", 176}
	ts3 = result.NetconfResult{4, 601, "172.26.138.93", "get", 59840, "", 3320}
	ts4 = result.NetconfResult{4, 2322, "172.26.138.91", "get", 56967, "", 420}
	ts5 = result.NetconfResult{4, 860, "172.26.138.92", "kill-session", 0, "kill-session is not a supported operation", 0}
)

func TestSortResults(t *testing.T) {

	testSort := func(t *testing.T, unsortedSlice []result.NetconfResult, want []result.NetconfResult) {
		t.Helper()

		SortResults(unsortedSlice)
		got := unsortedSlice

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	t.Run("Sort by Hostname", func(t *testing.T) {
		unsortedSlice := []result.NetconfResult{ts3, ts1, ts2}
		want := []result.NetconfResult{ts1, ts2, ts3}

		testSort(t, unsortedSlice, want)
	})

	t.Run("Sort by Operation", func(t *testing.T) {
		unsortedSlice := []result.NetconfResult{ts3, ts4, ts2, ts5, ts1}
		want := []result.NetconfResult{ts1, ts4, ts2, ts5, ts3}

		testSort(t, unsortedSlice, want)
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

/*
	NOTE: In the test function below I am only testing one NetconfResult struct this is due to the problem
	I mentioned I had in the meeting earlier on today regarding the latencies hashMap. I will try using the
	SortResults func after this commit to correct the problem.

	TODO:
	Add test cases to capture op and hostname test cases
		if op != "" && op != operation {
			continue
		}
		if hostname != "" && hostname != host {
			continue
		}
*/
func TestAnalyseResults(t *testing.T) {

	var mockCmd *cobra.Command
	var mockResults = []result.NetconfResult{ts1}
	var mockTs = TestSuite{}
	mockTs.File = "testdata/emptytestsuite.yml"

	// Capture StdErr
	var lOut = new(bytes.Buffer)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamps
	log.SetOutput(lOut)                                  // log output

	// Capture StdOut
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AnalyseResults(mockCmd, &mockTs, mockResults)

	out := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		defer r.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		out <- buf.String()
		buf.Reset()
	}()

	w.Close()
	os.Stdout = old
	cOut := <-out // console output

	// Build log test string

	var logBuffer bytes.Buffer

	logBuffer.WriteString("Testsuite executed at " + strings.Split(mockTs.File, string(filepath.Separator))[1] + " Suite defined the following hosts: ")
	logBuffer.WriteString("[")
	for _, config := range mockTs.Configs {
		logBuffer.WriteString(config.Hostname + " ")
	}
	logBuffer.WriteString("] ")

	latencies := make(map[string]map[string][]float64)
	errCount := OrderAndExcludeErrValues(mockResults, latencies)

	var when float64
	for _, result := range mockResults {
		if result.When > when {
			when = result.When
		}
	}
	executionTime := time.Duration(when) * time.Millisecond

	logBuffer.WriteString(strconv.Itoa(mockTs.Clients) + " client(s) started, " + strconv.Itoa(mockTs.Iterations) + " iterations per client, " + strconv.Itoa(mockTs.Rampup) + " seconds wait between starting each client ")
	logBuffer.WriteString(" Total execution time: " + executionTime.String() + ", Suite execution contained " + strconv.Itoa(errCount) + " errors")

	// Format logString

	re := regexp.MustCompile(`\r?\n`)
	got := strings.Trim(re.ReplaceAllString(lOut.String(), " "), " ")

	assert.Equal(t, got, logBuffer.String()) // test

	op := ""
	hostname := ""

	// Build console test string

	consoleBuffer := new(bytes.Buffer)
	consoleBuffer.WriteString("HOST OPERATION REUSE CONNECTION REQUESTS TPS MEAN VARIANCE STD DEVIATION ")
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
			consoleBuffer.WriteString(host + " " + operation + " " + strconv.FormatBool(mockTs.Configs.IsReuseConnection(host)) + " " + strconv.Itoa(len(latencies)) + " " + fmt.Sprintf("%.2f", tps) + " " + fmt.Sprintf("%.2f", mean) + " " + fmt.Sprintf("%.2f", variance) + " " + fmt.Sprintf("%.2f", stddev) + " ")
		}
	}
	removeWhtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	want := removeWhtsp.ReplaceAllString(cOut, "")
	removeWhtsp = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	want = removeWhtsp.ReplaceAllString(want, " ")

	got = strings.Trim(consoleBuffer.String(), " ")
	assert.Equal(t, want, got) //test
}

func TestAnalyseArgs(t *testing.T) {

	var testCmd = AnalyseCmd
	var cmd = &cobra.Command{}

	testStruct := func(t *testing.T, args []string, got error) {
		t.Helper()

		want := testCmd.Args(cmd, args) // args = 1 or != 1
		assert.Equal(t, got, want)
	}

	t.Run("args == 1", func(t *testing.T) {
		var a = []string{"a"}

		testStruct(t, a, nil)
	})

	t.Run("args != 1", func(t *testing.T) {
		b := []string{"a", "b", "c"}
		rstr := errors.New("analyse command requires a test results directory as an argument")

		testStruct(t, b, rstr)
	})
}

/*
NOTE:
Below is the test func which I discussed in the meeting earlier today.
To test the func I decided I would base my test around which exit status was returned: 0 for success;
non-0 for failure. Golang doesn’t provide a straightforward way of capturing this, so I had to look online for a solution.

Andrew Gerrand provides a solution to this problem here https://talks.golang.org/2014/testing.slide#23 ,
and discusses its implementation here https://www.youtube.com/watch?v=ndmB0bj7eyw&feature=youtu.be&t=47m09s

This solution uses a sub-process to examine how the program exits, and then allows you to look at it’s exit status.
The only downside to this is that as it uses a sub-process to examine how the main test function exits, Go
doesn’t recognise this subprocess and hence doesn’t count it when checking code coverage. This post explains
it further: https://stackoverflow.com/questions/40615641/testing-os-exit-scenarios-in-go-with-coverage-information-coveralls-io-goverall
*/

// TODO: Add further test cases
func TestAnalyseRun(t *testing.T) {

	tests := []struct {
		name     string
		testCmd  *cobra.Command
		testArgs []string
	}{
		{name: "single valid yaml file", testArgs: []string{"/Users/pconcannon/Documents/go/src/github.com/nc-hammer/results/2018-07-05-15-35-11/"}},
		//	{"single valid csv file", AnalyseCmd, []string{"results/2018-07-05-15-35-11/"}},
		//	{"single invalid file provided", AnalyseCmd, []string{"testdata/nonExisting.file/"}},
		//	{"no file provided", AnalyseCmd, []string{""}},
		//	{"multiple files provided", AnalyseCmd, []string{"testdata/test-suite.yml", "results/2018-07-05-15-35-11/", "testdata/nonExisting.file/"}},
	}

	for _, tt := range tests {
		//t.Run(tt.name, func(t *testing.T) {

		if os.Getenv("RUN_SUBPROCESS") == "1" {
			AnalyseCmd.Run(tt.testCmd, tt.testArgs)
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestAnalyseRun")
		cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			t.Fatalf("Program failed to load file -- os.Exit(1)")
		}
		return

	}
}
