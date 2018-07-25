package cmd_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/damianoneill/nc-hammer/cmd"
)

func Test_RootExecuteFail(t *testing.T) {
	mockRootCmd := cmd.RootCmd

	// set dud flag for root command to test failing-cause within root.Execute()
	mockRootCmd.SetArgs([]string{
		"_failing value_",
	})

	if os.Getenv("RUN_SUBPROCESS") == "1" {
		cmd.Execute("")
		return
	}
	// run function under test in parallel command process to check Exit code it returns
	cmd := exec.Command(os.Args[0], "-test.run=Test_RootExecute")
	cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return // if Exit code is expected return (passing)
	}
	t.Errorf("Error not caught; Execute() failed to exit correctly")
}
