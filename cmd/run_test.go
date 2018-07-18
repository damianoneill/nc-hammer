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
	start := time.Now()
	myTs, err := suite.NewTestSuite("../suite/testdata/test-suite.yml")

	if err != nil {
		t.Errorf("Problem loading testdata/testsuite.yml: %v", err)
	}

	var buff bytes.Buffer
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)
	block := myTs.GetInitBlock()
	runTestSuite(myTs)
	got := strings.TrimSpace(buff.String())
	want := "Testsuite " + myTs.File + " started at " + start.Format("Mon Jan _2 15:04:05 2006") +
		"\n > " + strconv.Itoa(myTs.Clients) + " client(s), " +
		strconv.Itoa(myTs.Iterations) + " iterations per client, " +
		strconv.Itoa(myTs.Rampup) + " seconds wait between starting each client\n"

	if block != nil {
		want += " > Init Block defined, executing " + strconv.Itoa(len(block.Actions)) + " init actions sequentially up front"
		strconv.Itoa(len(block.Actions))
	}
	want += "\n\nTestsuite completed in "
	want = strings.TrimSpace(want)
	assert.True(t, strings.Contains(got, want))

}
