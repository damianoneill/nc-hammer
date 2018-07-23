package action

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"
)

func Test_Execute(t *testing.T) {
	var buff bytes.Buffer
	tsValid, _ := suite.NewTestSuite("../suite/testdata/test-suite.yml")          // testsuite with netconf/sleep actions
	tsInvalid, _ := suite.NewTestSuite("../suite/testdata/testsuite-invalid.yml") // testsuite with no netconf or sleep actions
	myTests := []*suite.TestSuite{tsValid, tsInvalid}
	start := time.Now()
	resultChannel := make(chan result.NetconfResult)
	handleResultsFinished := make(chan bool)
	go result.HandleResults(resultChannel, handleResultsFinished, tsValid)
	for _, testsuite := range myTests {
		for _, b := range testsuite.Blocks {
			for _, a := range b.Actions {
				log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
				log.SetOutput(&buff)
				if testsuite == tsValid {
					Execute(start, 0, tsValid, a, resultChannel)
					assert.True(t, (a.Sleep != nil) || (a.Netconf != nil)) // checks for netconf or sleep actions
				} else {
					Execute(start, 0, tsInvalid, a, resultChannel)
					got := buff.String()
					want := "Problem"
					assert.Contains(t, got, want)
					buff.Reset()
				}
			}
		}
	}
}
