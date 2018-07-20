package action

import (
	"log"
	"time"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
)

// Execute used to determine type of Action and call the appropriate function
func Execute(tsStart time.Time, cID int, ts *suite.TestSuite, action suite.Action, resultChannel chan result.NetconfResult) {
	switch {
	case action.Netconf != nil:
		ExecuteNetconf(tsStart, cID, action, ts.GetConfig(action.Netconf.Hostname), resultChannel)
	case action.Sleep != nil:
		ExecuteSleep(action)
	default:
		log.Printf("\n ** Problem with your Testsuite, an action in a block section has incorrect YAML indentation for its body, ensure that anything after netconf or sleep is properly indented **\n\n")
	}
}
