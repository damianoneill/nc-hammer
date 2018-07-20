package cmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/stretchr/testify/assert"
)

var myCmd = analyseErrorCmd

func Test_AnalyseErrorCmdArgs(t *testing.T) {
	t.Run("test that a directory is not passed to the command", func(t *testing.T) {
		args := []string{}
		errLen := strconv.Itoa(len(args))
		analyseErrorCmd.Args(myCmd, args)
		assert.Equal(t, analyseErrorCmd.Args(myCmd, args), errors.New("error command requires a test results directory as an argument"), "failed"+errLen)
	})
	t.Run("test that a directory is passed to the command", func(t *testing.T) {
		args := []string{"arg1"}
		errLen := strconv.Itoa(len(args))
		analyseErrorCmd.Args(myCmd, args)
		assert.Equal(t, analyseErrorCmd.Args(myCmd, args), nil, "failed"+errLen)
	})
}
func Test_AnalyseErrorCmdRun(t *testing.T) {
	// pathArgs used to give an error
	if os.Getenv("BE_CRASHER") == "1" {
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
		pathArgs := []string{"error/2018-07-18-19-56-01/"}
		analyseErrorCmd.Run(myCmd, pathArgs)
		return
	}

	// Start the actual test in a different subprocess
	cmd := exec.Command(os.Args[0], "-test.run=Test_AnalyseErrorCmdRun")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	stdout, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Check that the log fatal message is what we expected
	gotBytes, _ := ioutil.ReadAll(stdout)
	got := string(gotBytes)
	_, _, errReturned := result.UnarchiveResults("error/2018-07-18-19-56-01/") // get err that triggered fatalf
	expected := "Problem with loading result information: " + errReturned.Error() + " "
	if !strings.HasSuffix(got[:len(got)-1], expected) {
		t.Fatalf("Unexpected log message. Got '%s' but should contain '%s'", got[:len(got)-1], expected)
	}

	// Check that the program exited
	err := cmd.Wait()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Process ran with err %v, want exit status 1", err)
	}
}
func Test_AnalyseErrors(t *testing.T) {
	results, myTs, err := result.UnarchiveResults("../suite/testdata/results_test/2018-07-18-19-56-01/")
	if err != nil {
		t.Error(err)
	}
	var errors [][]string
	for i := range results {
		if results[i].Err != "" {
			errors = append(errors, []string{results[i].Hostname, results[i].Operation, results[i].Err})
		}
	}
	var buff bytes.Buffer
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)

	analyseErrors(myCmd, myTs, results)

	got := strings.TrimSpace(buff.String())
	errLen := strconv.Itoa(len(errors))
	want := strings.TrimSpace("Testsuite executed at " + strings.Split(myTs.File, string(filepath.Separator))[1] +
		"\n" + "Total Number of Errors for suite: " + errLen)
	if got != want {
		t.Errorf("wanted, '%s', but got '%s'", want, got)
	}
}
