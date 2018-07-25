package cmd_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/damianoneill/nc-hammer/cmd"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestBuildTestSuite(t *testing.T) {

	mockPath := ""
	got := BuildTestSuite(mockPath)

	var want suite.TestSuite

	if got.File != mockPath {
		t.Errorf("Filename: got %v, want %v", got.File, mockPath)
	}
	if got.Iterations != 5 {
		t.Errorf("Iterations: got %v, want %v", got.Iterations, 5)
	}
	if got.Clients != 2 {
		t.Errorf("Clients: got %v, want %v", got.Clients, 2)
	}
	if got.Rampup != 0 {
		t.Errorf("Rampup: got %v, want %v", got.Rampup, 0)
	}
	if reflect.ValueOf(got.Configs).Type() != reflect.TypeOf(want.Configs) {
		t.Error("Testsuite.Configs is not of type Configs")
	}
	if reflect.ValueOf(got.Blocks).Type() != reflect.TypeOf(want.Blocks) {
		t.Error("Testsuite.Configs is not of type Blocks")
	}
}
func TestInitCmd(t *testing.T) { // check for correct return value
	var testCmd = InitCmd
	var cmd = &cobra.Command{}

	testFunc := func(t *testing.T, args []string, want error) {
		t.Helper()

		got := testCmd.Args(cmd, args)
		assert.Equal(t, got, want)
	}

	t.Run("args != 1", func(t *testing.T) {
		var a = []string{"x", "y"}
		testFunc(t, a, errors.New("init command requires a directory as an argument"))
	})

	t.Run("args == 1", func(t *testing.T) {
		var a = []string{"x"}
		testFunc(t, a, nil)
	})
}

func TestInitRun(t *testing.T) {
	mockPath := "x"
	YMLpath := filepath.Join(mockPath, "/test-suite.yml")
	snippetsPath := filepath.Join(mockPath, "/snippets")
	XMLpath := filepath.Join(mockPath, "/snippets/edit-config.xml")

	var testInitCmd = InitCmd
	var testCmd = &cobra.Command{}

	args := []string{mockPath}
	testInitCmd.Run(testCmd, args)

	testRun := func(t *testing.T) {
		t.Helper()
	}

	t.Run("initial load file is valid", func(t *testing.T) {
		testRun(t)
		if _, err := os.Stat(mockPath); os.IsNotExist(err) {
			t.Errorf("\n - File '%v' not found", mockPath)
		}
	})

	t.Run("init files created successfully with correct permissions", func(t *testing.T) {
		testRun(t)
		filesToCheck := []string{mockPath, YMLpath, snippetsPath, XMLpath}

		for _, n := range filesToCheck {
			// check if files exist
			if _, err := os.Stat(n); os.IsNotExist(err) {
				t.Errorf("\n - File '%v' not found", n)
			}
			// check if init files have correct permissions
			f, err := os.OpenFile(n, os.O_RDWR, 0666)
			if err != nil {
				if os.IsPermission(err) {
					t.Errorf("\n - %v", err)
				}
			}
			f.Close()
		}
	})

	t.Run("YML scaffold check", func(t *testing.T) {
		testRun(t)
		// create mock YAML using test functions
		mockTS := BuildTestSuite(YMLpath)
		mockYML, _ := yaml.Marshal(mockTS)

		// read init YAML
		initYML, _ := ioutil.ReadFile(YMLpath)

		c := bytes.Compare(mockYML, initYML) // change for assert.Equal or dirExist
		if c != 0 {
			t.Error("YML files not equal")
		}
	})

	t.Run("XML scaffold check", func(t *testing.T) {
		testRun(t)
		mockXML := []byte("<interface><name>Ethernet0/0</name><mtu>1500</mtu></interface>")

		// read init XML
		initXML, _ := ioutil.ReadFile(XMLpath)

		c := bytes.Compare(mockXML, initXML) // change for assert.Equal
		if c != 0 {
			t.Error("XML files not equal")
		}
	})
	// clean up test files
	os.RemoveAll(mockPath)
}

func TestRunSuccess(t *testing.T) {
	var testInitCmd = InitCmd
	var testCmd = &cobra.Command{}

	mockPath := "x"
	args := []string{mockPath}

	if os.Getenv("RUN_SUBPROCESS") == "1" {
		testInitCmd.Run(testCmd, args)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInitRun") // create new process to run test
	cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")       // set environmental variable
	err := cmd.Run()                                         // run
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {  // check exit status of test subprocess
		t.Errorf("\n - Exit Status 1 returned\n - File '%v' already exists", args[0])
	}
}
