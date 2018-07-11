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

func TestTable(t *testing.T) {
	tests := []struct {
		mockCmd     string
		mockResults string
		mockTs      bool
	}{
		// TODO: Add correcy test cases to test TestAnalyseResults
		{"valid get-config", "<get-config><source><running/></source></get-config>", false},
	}
	fmt.Println(tests)
}

func TestAnalyseResults(t *testing.T) {

	var mockCmd *cobra.Command
	var mockResults = []result.NetconfResult{ts1}
	var mockTs = TestSuite{}
	mockTs.File = "testdata/emptytestsuite.yml"

	/*
		Capture StdErr and StdOut
	*/

	//StdErr
	var lOut = new(bytes.Buffer)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamps
	log.SetOutput(lOut)

	// StdOut
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
	cOut := <-out

	/*
		Build test strings
	*/

	var tbuf bytes.Buffer

	tbuf.WriteString("Testsuite executed at " + strings.Split(mockTs.File, string(filepath.Separator))[1] + " Suite defined the following hosts: ")
	tbuf.WriteString("[")
	for _, config := range mockTs.Configs {

		tbuf.WriteString(config.Hostname + " ")
	}
	tbuf.WriteString("] ")

	latencies := make(map[string]map[string][]float64)
	errCount := OrderAndExcludeErrValues(mockResults, latencies)

	var when float64
	for _, result := range mockResults {
		if result.When > when {
			when = result.When
		}
	}
	executionTime := time.Duration(when) * time.Millisecond

	tbuf.WriteString(strconv.Itoa(mockTs.Clients) + " client(s) started, " + strconv.Itoa(mockTs.Iterations) + " iterations per client, " + strconv.Itoa(mockTs.Rampup) + " seconds wait between starting each client ")
	tbuf.WriteString(" Total execution time: " + executionTime.String() + ", Suite execution contained " + strconv.Itoa(errCount) + " errors")

	// Format logString
	re := regexp.MustCompile(`\r?\n`)
	got := strings.Trim(re.ReplaceAllString(lOut.String(), " "), " ")

	assert.Equal(t, got, tbuf.String())

	op := ""
	hostname := ""

	tbuf1 := new(bytes.Buffer)
	tbuf1.WriteString("HOST OPERATION REUSE CONNECTION REQUESTS TPS MEAN VARIANCE STD DEVIATION ")
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
			//data = append(data, []string{host, operation, strconv.FormatBool(mockTs.Configs.IsReuseConnection(host)), strconv.Itoa(len(latencies)), fmt.Sprintf("%.2f", tps), fmt.Sprintf("%.2f", mean), fmt.Sprintf("%.2f", variance), fmt.Sprintf("%.2f", stddev)})
			tbuf1.WriteString(host + " " + operation + " " + strconv.FormatBool(mockTs.Configs.IsReuseConnection(host)) + " " + strconv.Itoa(len(latencies)) + " " + fmt.Sprintf("%.2f", tps) + " " + fmt.Sprintf("%.2f", mean) + " " + fmt.Sprintf("%.2f", variance) + " " + fmt.Sprintf("%.2f", stddev) + " ")
		}
	}
	re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	final := re_leadclose_whtsp.ReplaceAllString(cOut, "")
	final = re_inside_whtsp.ReplaceAllString(final, " ")

	got = strings.Trim(tbuf1.String(), " ")
	assert.Equal(t, final, got)
}

func TestAnalyseArgs(t *testing.T) {

	var testCmd = AnalyseCmd
	var cmd = &cobra.Command{}

	testStruct := func(t *testing.T, args []string, got error) {
		t.Helper()

		want := testCmd.Args(cmd, args) // args > 1 or = 1

		assert.Equal(t, got, want)
	}

	t.Run("args = 1", func(t *testing.T) {
		var a = []string{"a"}

		testStruct(t, a, nil)
	})

	t.Run("args != 1", func(t *testing.T) {
		b := []string{"a", "b", "c"}
		rstr := errors.New("analyse command requires a test results directory as an argument")

		testStruct(t, b, rstr)
	})
}

func Crasher() {
	fmt.Println("Going down in flames!")
	os.Exit(1)
}

func TestAnalyseRun(t *testing.T) {

	tests := []struct {
		name     string
		testCmd  *cobra.Command
		testArgs []string
	}{
		{name: "single valid yaml file", testArgs: []string{"/Users/pconcannon/Documents/go/src/github.com/nc-hammer/results/2018-07-0-15-35-11/"}}, //&cobra.Command{},
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
		if e, ok := err.(*exec.ExitError); ok && !e.Success() { // check to see if the program ran successfully or not -- if you got an error and it didn't run successfully
			t.Fatalf("Program failed to load file -- os.Exit(1)")
		}
		return

	}
}
