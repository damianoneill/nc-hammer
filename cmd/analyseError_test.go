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
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var myCmd = &cobra.Command{}

type fn func(*cobra.Command, []string)

// helper function to redirect to stdout for xxCmdRun
func CaptureStdout(runFunction fn, command *cobra.Command, args []string) (string, string) {
	var buff bytes.Buffer
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)
	//reading from stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runFunction(command, args)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	st := strings.Join(strings.Fields(string(out)), " ") // stdout captured, spaces trimmed
	buf := strings.TrimSpace(buff.String())              // logs captured
	return st, buf
}

func Test_AnalyseErrorCmdArgs(t *testing.T) {
	t.Run("test that a directory is not passed to the command", func(t *testing.T) {
		args := []string{}
		analyseErrorCmd.Args(myCmd, args)
		assert.Equal(t, analyseErrorCmd.Args(myCmd, args), errors.New("error command requires a test results directory as an argument"), "failed")
	})
	t.Run("test that a directory is passed to the command", func(t *testing.T) {
		args := []string{"../results/2018-07-18-19-56-01/"}
		analyseErrorCmd.Args(myCmd, args)
		assert.Equal(t, analyseErrorCmd.Args(myCmd, args), nil, "failed")
	})
}
func Test_AnalyseErrorCmdRun(t *testing.T) {
	t.Run("test that a wrong path is passed as arg", func(t *testing.T) {
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
	})
	t.Run("test that a correct path is passed as arg", func(t *testing.T) {
		pathArgs := []string{"../suite/testdata/results_test/2018-07-18-19-56-01/"}
		_, _, err := result.UnarchiveResults(pathArgs[0])
		CaptureStdout(analyseErrorCmd.Run, myCmd, pathArgs)
		assert.Nil(t, err)
	})
}
func Test_analyseErrors(t *testing.T) {
	expectedResults := [][]string{
		{"172.26.138.91", "kill-session", "kill-session is not a supported operation"},
		{"172.26.138.92", "delete-config", "delete-config is not a supported operation"},
		{"172.26.138.93", "kill-session", "kill-session is not a supported operation"},
		{"172.26.138.94", "delete-config", "delete-config is not a supported operation"},
	}

	results, ts, err := result.UnarchiveResults("../suite/testdata/results_test/2018-07-18-19-56-01/")
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

	//reading from stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	analyseErrors(myCmd, ts, results)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	have := strings.Join(strings.Fields(string(out)), " ") // stdout captured and trim spaces

	assert.Contains(t, have, "HOSTNAME OPERATION MESSAGE ID ERROR")
	for _, expectedError := range expectedResults {
		errorsTostring := strings.Join(expectedError, " ")
		assert.Contains(t, have, errorsTostring) // check errors are printed to stdout
	}

	got := strings.TrimSpace(buff.String())
	errLen := strconv.Itoa(len(errors))
	want := strings.TrimSpace("Testsuite executed at " + strings.Split(ts.File, string(filepath.Separator))[1] +
		"\n" + "Total Number of Errors for suite: " + errLen)
	if got != want {
		t.Errorf("wanted, '%s', but got '%s'", want, got)
	}
}
