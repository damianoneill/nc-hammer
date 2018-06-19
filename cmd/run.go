package cmd

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/damianoneill/nc-hammer/action"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <test suite file>",
	Short: "Execute a Test Suite",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("run command requires a test suite file as an argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if ts, err := suite.NewTestSuite(args[0]); err != nil {
			log.Fatalf("Problem with YAML file: %v ", err)
		} else {
			runTestSuite(ts)
		}
	},
}

func runTestSuite(ts *suite.TestSuite) {
	start := time.Now()
	log.Printf("Testsuite %v started at %v\n", ts.File, start.Format("Mon Jan _2 15:04:05 2006"))
	log.Printf(" > %d client(s), %d iterations per client, %d seconds wait between starting each client\n", ts.Clients, ts.Iterations, ts.Rampup)

	// handle results in separate goroutine
	resultChannel := make(chan result.NetconfResult)
	handleResultsFinished := make(chan bool)
	go result.HandleResults(resultChannel, handleResultsFinished, ts.File)

	// check first for an init block, this runs at the start, actions are sequential, it only runs once
	// if the tester has specified more than one init block, these are ignored
	if block := ts.GetInitBlock(); block != nil {
		log.Printf(" > Init Block defined, executing %d init actions sequentially up front", len(block.Actions))
		for _, a := range block.Actions {
			action.Execute(start, 0, ts, a, resultChannel)
		}
	}
	// create concurrent sessions for each of the defined clients
	clientWg := sync.WaitGroup{}
	for cID := 0; cID < ts.Clients; cID++ {
		clientWg.Add(1)
		go handleBlocks(start, ts, cID, &clientWg, resultChannel)
		// handle rampup for each client
		var waitDuration = float32(ts.Rampup) / float32(ts.Clients)
		time.Sleep(time.Duration(int(1000*waitDuration)) * time.Millisecond)
	}
	clientWg.Wait()

	// close the results channel and wait for the results goroutine to finish
	close(resultChannel)
	<-handleResultsFinished

	// close any cached sessions
	action.CloseAllSessions()

	log.Printf("\nTestsuite completed in %v\n", time.Since(start))
}

// handleBlocks determines the block type and processes the actions appropriately
func handleBlocks(start time.Time, ts *suite.TestSuite, cID int, clientWg *sync.WaitGroup, resultChannel chan result.NetconfResult) {
	for i := 0; i < ts.Iterations; i++ {
		for _, block := range ts.Blocks {
			// block sections are executed sequentially, individual blocks may execute actions sequentially or councurrently
			switch block.Type {
			case "sequential":
				for _, a := range block.Actions {
					action.Execute(start, cID, ts, a, resultChannel)
				}
			case "concurrent":
				blockWg := sync.WaitGroup{}
				for _, a := range block.Actions {
					// do concurrently
					blockWg.Add(1)
					go func(a suite.Action) {
						defer blockWg.Done()
						action.Execute(start, cID, ts, a, resultChannel)
					}(a)
				}
				blockWg.Wait()
			case "init":
				// do nothing
			}
		}
	}
	clientWg.Done()
}

func init() {
	RootCmd.AddCommand(runCmd)
}
