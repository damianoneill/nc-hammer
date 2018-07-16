package cmd

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/damianoneill/nc-hammer/suite"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init <directory>",
	Short: "Scaffold a Test Suite and snippets directory",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("init command requires a directory as an argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		_, err := os.Stat(args[0])
		if !os.IsNotExist(err) {
			log.Fatalf("%v already exists, remove it before running init or use a different directory", args[0])
		}

		// create new directory and snippets directory
		path := filepath.Join(args[0], "snippets")
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// write out TestSuite Scaffold
		bytes, err := yaml.Marshal(BuildTestSuite(path))
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(args[0], "test-suite.yml"), bytes, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// write out edit-config.xml Scaffold
		bytes = []byte("<interface><name>Ethernet0/0</name><mtu>1500</mtu></interface>")
		err = ioutil.WriteFile(filepath.Join(path, "edit-config.xml"), bytes, 0644)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// BuildTestSuite Initialises a TestSuite struct with default values and returns a pointer to it.
func BuildTestSuite(path string) *suite.TestSuite {
	var ts suite.TestSuite
	ts.Iterations = 5
	ts.Clients = 2
	ts.Rampup = 0
	ts.Configs = suite.Configs{suite.Sshconfig{Hostname: "10.0.0.1", Port: 830, Username: "user", Password: "pass", Reuseconnection: false}}

	initBlock := suite.Block{Type: "init", Actions: []suite.Action{}}
	config := "file:" + filepath.Join("snippets", "edit-config.xml")
	editAction := suite.Action{Netconf: &suite.Netconf{Hostname: "10.0.0.1", Operation: "edit-config", Source: nil, Target: nil, Filter: nil, Config: &config}, Sleep: nil}
	initBlock.Actions = []suite.Action{editAction}

	concurrentBlock := suite.Block{Type: "concurrent", Actions: []suite.Action{}}
	namespace := "http://example.com/schema/1.2/config"
	filter := suite.Filter{Type: "subtree", Ns: &namespace, Select: "<users/>"}
	getConfigAction := suite.Action{Netconf: &suite.Netconf{Hostname: "10.0.0.1", Operation: "get-config", Source: nil, Target: nil, Filter: &filter, Config: nil}, Sleep: nil}
	concurrentBlock.Actions = []suite.Action{getConfigAction}
	ts.Blocks = []suite.Block{initBlock, concurrentBlock}

	return &ts
}

func init() {
	RootCmd.AddCommand(initCmd)
}
