package result

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/damianoneill/nc-hammer/suite"
	"github.com/gocarina/gocsv"
	yaml "gopkg.in/yaml.v2"
)

// NetconfResult used to store all data related to a NETCONF requests response
type NetconfResult struct {
	Client    int
	SessionID int
	Hostname  string
	Operation string
	Request   string
	Response  string
	When      float64
	Err       string
	Latency   float64
}

// HandleResults processes results as they occur
func HandleResults(resultChannel chan NetconfResult, handleResultsFinished chan bool, ts *suite.TestSuite) {
	// sit here collecting results until the channel is closed by the main go routine
	results := []NetconfResult{}
	for result := range resultChannel {
		results = append(results, result)
		if result.Err == "" {
			fmt.Printf(".")
		}
	}

	// store results for future processing
	err := ArchiveResults(results, ts)
	if err != nil {
		panic(err)
	}

	handleResultsFinished <- true
}

// ArchiveResults stores results for future processing
func ArchiveResults(results []NetconfResult, ts *suite.TestSuite) error {
	// create the output directory based on current timestamp
	now := time.Now().Format("2006-01-02-15-04-05")
	path := filepath.Join("./results", now)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	// open a file for writing
	resultsFile, err := os.OpenFile(filepath.Join(path, "results.csv"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	// nolint
	defer resultsFile.Close()

	// write the results as a csv file
	err = gocsv.MarshalFile(results, resultsFile)
	if err != nil {
		return err
	}

	// write the TestSuite, which included any xml files inlined.
	bytes, err := yaml.Marshal(&ts)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(path, "test-suite.yml"), bytes, 0644)
	return err
}

// UnarchiveResults loads a test suite results from the filesystem
func UnarchiveResults(resultsPath string) ([]NetconfResult, *suite.TestSuite, error) {
	var results []NetconfResult
	var s *suite.TestSuite

	resultFile, err := os.OpenFile(filepath.Join(resultsPath, "results.csv"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}
	// nolint
	defer resultFile.Close()

	if err = gocsv.UnmarshalFile(resultFile, &results); err != nil { // Load clients from file
		return nil, nil, err
	}

	s, err = suite.NewTestSuite(filepath.Join(resultsPath, "test-suite.yml"))

	return results, s, err
}
