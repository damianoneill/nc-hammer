package cmd

import (
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"
)

func Test_ArgsRun(t *testing.T) {
	t.Run("len(arg) != 1", func(t *testing.T) {
		args := []string{}
		errLen := strconv.Itoa(len(args))
		runCmd.Args(myCmd, args)

		assert.Equal(t, runCmd.Args(myCmd, args), errors.New("run command requires a test suite file as an argument"), "failed"+errLen)
	})
	t.Run("len(arg) == 1", func(t *testing.T) {
		args := []string{"arg1"}
		errLen := strconv.Itoa(len(args))
		runCmd.Args(myCmd, args)
		assert.Equal(t, runCmd.Args(myCmd, args), nil, "failed"+errLen)
	})
}

func Test_runTestSuite(t *testing.T) {
	myTs := suite.TestSuite{}
	myTs.File = "testdata/testsuite.yml"

	var buff bytes.Buffer
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)

	runTestSuite(&myTs)

	got := strings.TrimSpace(buff.String())
	wantPre := strings.TrimSpace(
		"Testsuite " + myTs.File + " started at " +
			time.Now().Format("Mon Jan _2 15:04:05 2006") + "\n > " + strconv.Itoa(myTs.Clients) + " client(s), " + strconv.Itoa(myTs.Iterations) + " iterations per client, " +
			strconv.Itoa(myTs.Rampup) + " seconds wait between starting each client\n")
	want := wantPre + "\n\nTestsuite completed in "

	assert.True(t, strings.Contains(got, want))
}
