package cmd

import (
	"reflect"
	"testing"

	"github.com/damianoneill/nc-hammer/result"
)

var ts1 = result.NetconfResult{0, 0, "10.0.0.1", "get", 0, "error", 0}
var ts2 = result.NetconfResult{0, 0, "10.0.0.2", "get", 0, "", 0}
var ts3 = result.NetconfResult{0, 0, "10.0.0.3 ", "get", 0, "", 0}
var ts4 = result.NetconfResult{0, 0, "10.0.0.3 ", "edit-config", 0, "", 0}

func TestSortResults(t *testing.T) {

	testSort := func(t *testing.T, testArray []result.NetconfResult, want []result.NetconfResult) {
		t.Helper()

		sortResults(testArray)
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
		testArray := []result.NetconfResult{ts3, ts4, ts2}
		want := []result.NetconfResult{ts2, ts4, ts3}

		testSort(t, testArray, want)
	})

}

// Note*  How can I check if latencies has been populated? Unsure as
// to whether the latencies map is being copied or referenced in func.
// If it's not returned here, how can I check it's validity?

func TestOrderAndExcludeErrValues(t *testing.T) {
	testResults := []result.NetconfResult{ts1, ts2, ts3}
	testLatencies := make(map[string]map[string][]float64)

	got := 1
	want := orderAndExcludeErrValues(testResults, testLatencies)

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAnalyseResults(t *testing.T) {
	// func analyseResults(cmd *cobra.Command, ts *suite.TestSuite, results []result.NetconfResult)

	got := "" //analyseResults
	want := ""

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
