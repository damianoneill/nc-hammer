package cmd_test

import (
	"reflect"
	"testing"

	"github.com/damianoneill/nc-hammer/result"
	. "github.com/damianoneill/nc-hammer/suite"
	. "github.com/nc-hammer/cmd"
	"github.com/spf13/cobra"
	//. "github.com/nc-hammer/suite"
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

	//var mockBlock = []Block{Block{"test", []Action{}}}
	//var mockConfig Configs
	var mockCmd *cobra.Command

	var testArray = []result.NetconfResult{ts1, ts2, ts3}

	//var mockTest, _ = NewTestSuite("cmd/test.yml")
	//var mockTest = TestSuite{"/suite/testdata", 10, 7, 2, mockConfig, mockBlock}

	emptyTs := TestSuite{}
	emptyTs.File = "testdata/emptytestsuite.yml"

	AnalyseResults(mockCmd, &emptyTs, testArray)
}
