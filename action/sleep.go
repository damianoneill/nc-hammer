package action

import (
	"time"

	"github.com/damianoneill/nc-hammer/suite"
)

// ExecuteSleep invoked when a Sleep Action is identified
func ExecuteSleep(action suite.Action) {
	time.Sleep(time.Duration(action.Sleep.Duration) * time.Second)
}
