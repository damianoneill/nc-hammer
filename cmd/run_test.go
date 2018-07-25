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

func Test_runCmdArgs(t *testing.T) {
	t.Run("no args passed to RunCmd test", func(t *testing.T) {
		args := []string{}
		errLen := strconv.Itoa(len(args))
		runCmd.Args(cmdTest, args)
		assert.Equal(t, runCmd.Args(cmdTest, args), errors.New("run command requires a test suite file as an argument"), "failed"+errLen)
	})
	t.Run("arg/path passed to RunCmd test ", func(t *testing.T) {
		args := []string{"../suite/testdata/test-suite.yml"}
		errLen := strconv.Itoa(len(args))
		runCmd.Args(cmdTest, args)
		assert.Equal(t, runCmd.Args(cmdTest, args), nil, "failed"+errLen)
	})
}

func Test_runCmdRun(t *testing.T) {

	t.Run("test that an arg is of invalid path for yml file", func(t *testing.T) {
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
	})
	t.Run("test that a correct arg is passed", func(t *testing.T) {
		pathArgs := []string{"../suite/testdata/test-suite.yml"}
		_, err := suite.NewTestSuite(pathArgs[0])
		runCmd.Run(cmdTest, pathArgs)
		assert.Nil(t, err)
	})
}

func Test_runTestSuite(t *testing.T) {
	start := time.Now()
	ts, err := suite.NewTestSuite("../suite/testdata/test-suite.yml")
	if err != nil {
		t.Errorf("Problem loading YAML file: %v", err)
	}
	block := ts.GetInitBlock()
	//reading from buffer
	var buff bytes.Buffer
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(&buff)
	//reading from stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runTestSuite(ts)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	have := strings.TrimSpace(string(out))              // stdout captured
	got := strings.Replace(buff.String(), "\n", "", -1) // capturing logs
	want := "Testsuite " + ts.File + " started at " + start.Format("Mon Jan _2 15:04:05 2006") +
		"\n > " + strconv.Itoa(ts.Clients) + " client(s), " +
		strconv.Itoa(ts.Iterations) + " iterations per client, " +
		strconv.Itoa(ts.Rampup) + " seconds wait between starting each client"
	if block != nil {
		want += " > Init Block defined, executing " + strconv.Itoa(len(block.Actions)) + " init actions sequentially up front"
		strconv.Itoa(len(block.Actions))
	}
	want += "Testsuite completed in "
	want = strings.Replace(want, "\n", "", -1) // logs captured
	// testing runTestSuite output
	assert.True(t, (strings.Contains(have, "E")) || (strings.Contains(have, ".")))
	assert.True(t, strings.Contains(got, want))
}
