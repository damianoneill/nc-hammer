package result_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

func TestHandleResults(t *testing.T) {

	var mock_NetConfResults = []result.NetconfResult{
		result.NetconfResult{Client: 5, SessionID: 2318, Hostname: "10.0.0.1", Operation: "edit-config", When: 55282, Err: "", Latency: 288},
		result.NetconfResult{Client: 6, SessionID: 859, Hostname: "10.0.0.2", Operation: "get-config", When: 55943, Err: "", Latency: 176},
		result.NetconfResult{Client: 4, SessionID: 601, Hostname: "10.0.0.3", Operation: "get", When: 9840, Err: "", Latency: 3320},
		result.NetconfResult{Client: 4, SessionID: 2322, Hostname: "10.0.0.1", Operation: "get", When: 56967, Err: "", Latency: 420},
		result.NetconfResult{Client: 4, SessionID: 860, Hostname: "10.0.0.2", Operation: "kill-session", When: 0, Err: "kill-session is not a supported operation", Latency: 0},
	}

	var mockTestsuite = &suite.TestSuite{}
	var mockResultChan = make(chan result.NetconfResult)
	var mockResultsHandler = make(chan bool)

	go result.HandleResults(mockResultChan, mockResultsHandler, mockTestsuite) // run channels

	// feed mock data into result.HandleResults() via mockResultChan channel
	foundResults := []result.NetconfResult{}
	for _, r := range mock_NetConfResults {
		mockResultChan <- r
		foundResults = append(foundResults, r)
	}
	close(mockResultChan)
	<-mockResultsHandler // Finish

	// test to see if HandleResults() has recorded Results correctly
	if !reflect.DeepEqual(foundResults, mock_NetConfResults) {
		t.Errorf("got %v want %v", foundResults, mock_NetConfResults)
	}
}

func TestUnarchiveResults(t *testing.T) {

	mockResultPath := "../suite/testdata/results_test/2018-07-18-19-56-01/"

	var mockNCR = []result.NetconfResult{}
	var mockTestsuite, mockErr = suite.NewTestSuite(filepath.Join(mockResultPath, "test-suite.yml"))

	results, _ := os.OpenFile(filepath.Join(mockResultPath, "results.csv"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	gocsv.UnmarshalFile(results, &mockNCR)

	nr, ts, e := result.UnarchiveResults(mockResultPath)

	assert.Equal(t, mockNCR, nr)
	assert.Equal(t, mockTestsuite, ts)
	assert.Equal(t, mockErr, e)

}
