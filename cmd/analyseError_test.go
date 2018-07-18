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

	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"

	"github.com/damianoneill/nc-hammer/result"
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

	var testArray = []result.NetconfResult{
		{5, 4, "172.26.138.91", "edit-config", 55282, "", 288},
		{6, 859, "172.26.138.92", "get-config", 55943, "kill-session is not a supported operation", 176},
		{4, 2322, "172.26.138.91", "get", 56967, "kill-session is not a supported operation", 420},
		{4, 601, "172.26.138.93", "get", 59840, "session closed by remote side", 3320},
		{4, 2322, "172.26.138.91", "get", 56967, "kill-session is not a supported operation", 420},
	}

	err := [][]string{
		{"172.26.138.92", "get-config", "kill-session is not a supported operation"},
		{"172.26.138.93", "get", "session closed by remote side"},
		{"172.26.138.91", "get", "kill-session is not a supported operation"},
		{"172.26.138.92", "kill-session", "session closed by remote side"},
	}

	myTs := suite.TestSuite{}
	myTs.File = "testdata/testsuite.yml"

	var buff bytes.Buffer

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)

	analyseErrors(myCmd, &myTs, testArray)

	got := strings.TrimSpace(buff.String())
	errLen := strconv.Itoa(len(err))
	want := strings.TrimSpace("Testsuite executed at " + strings.Split(myTs.File, string(filepath.Separator))[1] +
		"\n" + "Total Number of Errors for suite: " + errLen)
	assert.Equal(t, got, want, "failed not equal")

}
