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

// Checks the creation of the skeleton testsuite file, and it's default values
func TestBuildTestSuite(t *testing.T) {

	mockPath := ""
	actual := BuildTestSuite(mockPath)

	var expected suite.TestSuite

	if actual.File != mockPath {
		t.Errorf("Filename: actual %v, expected %v", actual.File, mockPath)
	}
	if actual.Iterations != 5 {
		t.Errorf("Iterations: actual %v, expected %v", actual.Iterations, 5)
	}
	if actual.Clients != 2 {
		t.Errorf("Clients: actual %v, expected %v", actual.Clients, 2)
	}
	if actual.Rampup != 0 {
		t.Errorf("Rampup: actual %v, expected %v", actual.Rampup, 0)
	}
	if reflect.ValueOf(actual.Configs).Type() != reflect.TypeOf(expected.Configs) {
		t.Error("Testsuite.Configs is not of type Configs")
	}
	if reflect.ValueOf(actual.Blocks).Type() != reflect.TypeOf(expected.Blocks) {
		t.Error("Testsuite.Configs is not of type Blocks")
	}
}

// Test to check handling of arguments in InitCmd.Args
func TestInitCmd(t *testing.T) {
	var mockCmd = InitCmd
	var tempCmd = &cobra.Command{}

	testArgs := func(t *testing.T, args []string, expected error) { // args = 1 or != 1
		t.Helper()

		actual := mockCmd.Args(tempCmd, args)
		assert.Equal(t, actual, expected)
	}

	t.Run("args == 1", func(t *testing.T) {
		var mockArgs = []string{"run"}
		testArgs(t, mockArgs, nil)
	})

	t.Run("args != 1", func(t *testing.T) {
		var mockArgs = []string{"run", "error", "test"}
		testArgs(t, mockArgs, errors.New("init command requires a directory as an argument"))
	})
}

// Test checks to see if initial test files and directory are created correctly
func TestInitRun(t *testing.T) {
	mockDirPath := "temp_testDir" // create temp directory for files to be outputted to
	mockYAMLpath := filepath.Join(mockDirPath, "test-suite.yml")
	mockSnippetsPath := filepath.Join(mockDirPath, "snippets/")
	mockXMLpath := filepath.Join(mockDirPath, "/snippets/edit-config.xml")

	var mockCmd = InitCmd
	var tempCmd = &cobra.Command{}

	args := []string{mockDirPath}
	mockCmd.Run(tempCmd, args)

	testInit := func(t *testing.T) {
		t.Helper()
	}

	t.Run("check validity of initial load file", func(t *testing.T) {
		testInit(t)
		if _, err := os.Stat(mockDirPath); os.IsNotExist(err) {
			t.Errorf("\n - File '%v' not found", mockDirPath)
		}
	})

	t.Run("check permissions correctly set on init files", func(t *testing.T) {
		testInit(t)
		filesToCheck := []string{mockDirPath, mockYAMLpath, mockSnippetsPath, mockXMLpath}

		for _, c := range filesToCheck {
			// check if files exist
			if _, err := os.Stat(c); os.IsNotExist(err) {
				t.Errorf("\n - File '%v' not found", c)
			}
			// check if init files have correct permissions
			f, err := os.OpenFile(c, os.O_RDWR, 0666)
			if err != nil {
				if os.IsPermission(err) {
					t.Errorf("\n - %v", err)
				}
			}
			f.Close()
		}
	})

	t.Run("YAML scaffold check", func(t *testing.T) {
		testInit(t)
		// create mock YAML using test functions
		mockTestSuite := BuildTestSuite(mockYAMLpath)
		expectedYAML, _ := yaml.Marshal(mockTestSuite)

		// read init YAML
		actualYAML, _ := ioutil.ReadFile(mockYAMLpath)

		c := bytes.Compare(actualYAML, expectedYAML) // change for assert.Equal
		if c != 0 {
			t.Error("YAML files not equal")
		}
	})

	t.Run("XML scaffold check", func(t *testing.T) {
		testInit(t)
		expectedXML := []byte("<interface><name>Ethernet0/0</name><mtu>1500</mtu></interface>")

		// read init XML
		actualXML, _ := ioutil.ReadFile(mockXMLpath)

		c := bytes.Compare(actualXML, expectedXML) // change for assert.Equal
		if c != 0 {
			t.Error("XML files not equal")
		}
	})
	// clean up test dir and files
	os.RemoveAll(mockDirPath)
}

// Test checks if Init.run runs successfully or not
func TestInitRunCompletion(t *testing.T) {
	var mockCmd = InitCmd
	var tempCmd = &cobra.Command{}

	mockDirPath := "temp_testDir"
	args := []string{mockDirPath}

	if os.Getenv("RUN_SUBPROCESS") == "1" {
		mockCmd.Run(tempCmd, args)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInitRunCompletion") // create new process to run test
	cmd.Env = append(os.Environ(), "RUN_SUBPROCESS=1")                 // set environmental variable
	err := cmd.Run()                                                   // run
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {            // check exit status of test subprocess
		t.Errorf("\n - Exit Status 1 returned\n - File '%v' already exists", args[0])
	}
	// clean up test dir and files
	os.RemoveAll(mockDirPath)
}
