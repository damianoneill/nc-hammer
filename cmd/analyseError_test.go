package cmd

import (
	"bytes"
	"errors"
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

func Test_Args(t *testing.T) {
	t.Run("len(arg) != 1", func(t *testing.T) {
		args := []string{}
		errLen := strconv.Itoa(len(args))
		analyseErrorCmd.Args(myCmd, args)

		assert.Equal(t, analyseErrorCmd.Args(myCmd, args), errors.New("error command requires a test results directory as an argument"), "failed"+errLen)
	})
	t.Run("len(arg) == 1", func(t *testing.T) {
		args := []string{"arg1"}
		errLen := strconv.Itoa(len(args))
		analyseErrorCmd.Args(myCmd, args)
		assert.Equal(t, analyseErrorCmd.Args(myCmd, args), nil, "failed"+errLen)
	})
}

func Test_Run(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		// Run the crashing code when FLAG is set
		if os.Getenv("FLAG") == "1" {
			args := []string{"arg1"}
			analyseErrorCmd.Run(myCmd, args)
			return
		}
		// Run the test in a subprocess
		cmd := exec.Command(os.Args[0], "-test.run=Test_Run")
		cmd.Env = append(os.Environ(), "FLAG=1")
		err := cmd.Run()

		// Cast the error as *exec.ExitError and compare the result
		e, ok := err.(*exec.ExitError)
		expectedErrorString := "exit status 1"
		assert.Equal(t, true, ok)
		assert.Equal(t, expectedErrorString, e.Error())
	})
	t.Run("err is nil", func(t *testing.T) {

		args := []string{"../suite/testdata/"}
		analyseErrorCmd.Run(myCmd, args)
	})
}
func Test_anlyseErrors(t *testing.T) {
	results, myTs, err := result.UnarchiveResults("../suite/testdata/")
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
	assert.Equal(t, got, want, "failed not equal")
}
