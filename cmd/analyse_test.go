package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/damianoneill/nc-hammer/result"
	. "github.com/damianoneill/nc-hammer/suite"
	. "github.com/nc-hammer/cmd"
	"github.com/spf13/cobra"
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

	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	AnalyseResults(mockCmd, &mockTs, testArray)

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
		buf.Reset()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	// reading our temp stdout
	fmt.Println("previous output:")
	fmt.Print(out)

	//fmt.Println("\nLOG\n-------------")
	//fmt.Print(logOutput)
	//fmt.Println("\nCONSOLE\n-------------")
	//fmt.Println(b)
}
