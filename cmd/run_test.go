package cmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"
)

var cmdTest = runCmd

func Test_RunCmdArgs(t *testing.T) {
	t.Run("no args passed to RunCmd test", func(t *testing.T) {
		args := []string{}
		errLen := strconv.Itoa(len(args))
		runCmd.Args(cmdTest, args)
		assert.Equal(t, runCmd.Args(cmdTest, args), errors.New("run command requires a test suite file as an argument"), "failed"+errLen)
	})
	t.Run("arg/path passed to RunCmd test ", func(t *testing.T) {
		args := []string{"arg1"}
		errLen := strconv.Itoa(len(args))
		runCmd.Args(cmdTest, args)
		assert.Equal(t, runCmd.Args(cmdTest, args), nil, "failed"+errLen)
	})
}

func Test_runCmdRun(t *testing.T) {
	// pathArgs used to give an error
	if os.Getenv("BE_CRASHER") == "1" {
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
		pathArgs := []string{"error/test-suite.yml"}
		runCmd.Run(cmdTest, pathArgs)
		return
	}

	// Start the actual test in a different subprocess
	cmd := exec.Command(os.Args[0], "-test.run=Test_runCmdRun")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	stdout, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Check that the log fatal message is what we expected
	gotBytes, _ := ioutil.ReadAll(stdout)
	got := string(gotBytes)
	_, errReturned := suite.NewTestSuite("error/test-suite.yml") // get err that triggered fatalf
	expected := "Problem with YAML file: " + errReturned.Error() + " "
	if !strings.HasSuffix(got[:len(got)-1], expected) {
		t.Fatalf("Unexpected log message. Got '%s' but should contain '%s'", got[:len(got)-1], expected)
	}

	// Check that the program exited
	err := cmd.Wait()
	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Process ran with err %v, want exit status 1", err)
	}
}

func Test_runTestSuite(t *testing.T) {
	start := time.Now()
	ts, _ := suite.NewTestSuite("../suite/testdata/test-suite.yml")
	block := ts.GetInitBlock()
	var buff bytes.Buffer
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)

	runTestSuite(ts)

	got := strings.Replace(buff.String(), "\n", "", -1)
	want := "Testsuite " + ts.File + " started at " + start.Format("Mon Jan _2 15:04:05 2006") +
		"\n > " + strconv.Itoa(ts.Clients) + " client(s), " +
		strconv.Itoa(ts.Iterations) + " iterations per client, " +
		strconv.Itoa(ts.Rampup) + " seconds wait between starting each client"
	if block != nil {
		want += " > Init Block defined, executing " + strconv.Itoa(len(block.Actions)) + " init actions sequentially up front"
		strconv.Itoa(len(block.Actions))
	}
	want += "Testsuite completed in "
	want = strings.Replace(want, "\n", "", -1)
	// testing runTestSuite output
	assert.True(t, strings.Contains(got, want))
}
