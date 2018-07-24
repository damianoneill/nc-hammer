package suite

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/xml"
	yaml "gopkg.in/yaml.v2"
)

// Sshconfig defines a definition for the parameters required to connect to a NETCONF Agent via SSH
type Sshconfig struct {
	Hostname        string `json:"hostname" yaml:"hostname"`
	Port            int    `json:"port" yaml:"port"`
	Username        string `json:"username" yaml:"username"`
	Password        string `json:"password" yaml:"password"`
	Reuseconnection bool   `json:"reuseconnection" yaml:"reuseconnection"`
}

// Filter defines the parameters required to generate a subtree or xpath filter within a NETCONF Request
type Filter struct {
	Type   string  `json:"type" yaml:"type"`
	Ns     *string `json:"ns,omitempty" yaml:"ns,omitempty"`
	Select string  `json:"select" yaml:"select"`
}

// Netconf struct contains information required to construct a valid NETCONF Operation.
// Addresses are used to indicate optional content
type Netconf struct {
	Hostname  string  `json:"hostname" yaml:"hostname"`
	Message   *string `json:"message,omitempty" yaml:"message,omitempty"`
	Method    *string `json:"method,omitempty" yaml:"method,omitempty"`
	Operation *string `json:"operation,omitempty" yaml:"operation,omitempty"`
	Source    *string `json:"source,omitempty" yaml:"source,omitempty"`
	Target    *string `json:"target,omitempty" yaml:"target,omitempty"`
	Filter    *Filter `json:"filter,omitempty" yaml:"filter,omitempty"`
	Config    *string `json:"config,omitempty" yaml:"config,omitempty"`
	Expected  *string `json:"expected,omitempty" yaml:"expected,omitempty"`
}

// Sleep is an action instructing the client to sleep for the period defined in duration
type Sleep struct {
	Duration int `json:"duration" yaml:"duration"` // seconds
}

// Action is a wrapper for the different actions types (netconf, sleep)
type Action struct {
	Netconf *Netconf `json:"netconf,omitempty" yaml:"netconf,omitempty"`
	Sleep   *Sleep   `json:"sleep,omitempty" yaml:"sleep,omitempty"`
}

// Block describes a list of actions and how these should treated; as an init block, sequentially or concurrently
type Block struct {
	Type    string   `json:"type" yaml:"type"`
	Actions []Action `json:"actions" yaml:"actions"`
}

// Configs rebinds the slice of Sshconfig so that methods can be constructed against it
type Configs []Sshconfig

// IsReuseConnection iterates through the Config slice and matches on host returning whether the connection should be reused or not
func (c Configs) IsReuseConnection(hostname string) bool {
	for idx := range c {
		if c[idx].Hostname == hostname {
			return c[idx].Reuseconnection
		}
	}
	return false
}

// TestSuite is the top level struct for the yaml document definition
type TestSuite struct {
	File       string  `json:"-" yaml:"-"`
	Iterations int     `json:"iterations" yaml:"iterations"`
	Clients    int     `json:"clients" yaml:"clients"`
	Rampup     int     `json:"rampup" yaml:"rampup"`
	Configs    Configs `json:"configs" yaml:"configs"`
	Blocks     []Block `json:"blocks" yaml:"blocks"`
}

// NewTestSuite returns an TestSuite initialized from a yaml file
func NewTestSuite(file string) (*TestSuite, error) {
	yamlFile, err := ioutil.ReadFile(file) // #nosec
	if err != nil {
		return nil, err
	}

	var ts TestSuite
	err = yaml.Unmarshal(yamlFile, &ts)
	if err == nil {
		err = validateTestSuite(&ts)
		if err != nil {
			return nil, err
		}
	}
	// inline any embedded xml
	err = InlineXML(&ts)
	if err != nil {
		return nil, err
	}

	ts.File = file
	return &ts, err
}

var snippets map[string]*string

// InlineXML iterates over a testsuite looking for inline file tag, on finding
// it will attempt to load and replace with inline xml
func InlineXML(ts *TestSuite) error {
	snippets = make(map[string]*string)
	m := minify.New()
	m.AddFunc("text/xml", xml.Minify)
	for _, block := range ts.Blocks {
		for _, action := range block.Actions {
			if action.Netconf != nil {
				if *action.Netconf.Operation == "edit-config" {
					if action.Netconf.Config != nil {
						if strings.HasPrefix(*action.Netconf.Config, "file:") {
							if _, ok := snippets[*action.Netconf.Config]; !ok {
								// first time reading file and store in map
								b, err := readXMLSnippet(strings.SplitAfter(*action.Netconf.Config, "file:")[1])
								if err != nil {
									return err
								}
								inline, err := m.String("text/xml", string(b))
								if err != nil {
									return err
								}
								snippets[*action.Netconf.Config] = &inline

							}
							action.Netconf.Config = snippets[*action.Netconf.Config]
						}
					}
				}
			}
		}
	}
	return nil
}

func readXMLSnippet(filename string) ([]byte, error) {
	xmlFile, err := os.Open(filename) // #nosec
	if err != nil {
		return nil, err
	}
	// nolint
	defer xmlFile.Close()

	b, err := ioutil.ReadAll(xmlFile)
	return b, err
}

// ToXMLString generates a XML representation of the information provided in the Netconf section of the TestSuite
func (n *Netconf) ToXMLString() (string, error) {
	doc := etree.NewDocument()
	operation := doc.CreateElement(*n.Operation)
	switch *n.Operation {
	case "get-config":
		source := operation.CreateElement("source")
		if n.Source != nil {
			source.CreateElement(*n.Source)
		} else {
			source.CreateElement("running")
		}
		addFilterIfPresent(n, operation)
	case "get":
		addFilterIfPresent(n, operation)
	case "edit-config":
		source := operation.CreateElement("target")
		if n.Target != nil {
			source.CreateElement(*n.Target)
		} else {
			source.CreateElement("running")
		}
		config := operation.CreateElement("config")
		if n.Config != nil {
			inner := etree.NewDocument()
			err := inner.ReadFromString(*n.Config)
			if err != nil {
				log.Println("Config data is not valid xml")
			}
			config.AddChild(inner.Root().Copy())
		}
	default:
		return "", errors.New(*n.Operation + " is not a supported operation")

	}
	return doc.WriteToString()
}

func addFilterIfPresent(n *Netconf, operation *etree.Element) {
	if n.Filter != nil {
		filter := operation.CreateElement("filter")
		filter.CreateAttr("type", n.Filter.Type)
		//  https://github.com/beevik/etree/issues/49
		inner := etree.NewDocument()
		err := inner.ReadFromString(n.Filter.Select)
		if err != nil {
			log.Println("Filter Select is not valid xml")
		}
		if n.Filter.Ns != nil {
			top := filter.CreateElement("top")
			top.CreateAttr("xmlns", *n.Filter.Ns)
			top.AddChild(inner.Root().Copy())
		} else {
			filter.AddChild(inner.Root().Copy())
		}
	}
}

// GetConfig returns the connection information for a specific host
func (ts *TestSuite) GetConfig(hostname string) *Sshconfig {
	for idx := range ts.Configs {
		if ts.Configs[idx].Hostname == hostname {
			return &ts.Configs[idx]
		}
	}
	return nil
}

// GetInitBlock returns an init block if defined in the TestSuite
func (ts *TestSuite) GetInitBlock() *Block {
	for _, block := range ts.Blocks {
		if block.Type == "init" {
			return &block
		}
	}
	return nil
}

func validateTestSuite(ts *TestSuite) error {
	if len(ts.Configs) == 0 {
		return errors.New("Testsuite should contain at least one SSH Config section")
	}

	hosts, err := validateSSHConfig(ts)
	if err != nil {
		return err
	}

	for _, block := range ts.Blocks {
		for _, action := range block.Actions {
			if action.Netconf != nil {
				if *action.Netconf.Operation == "" {
					return errors.New("netconf: operation cannot be empty")
				}
				if !StringInSlice(action.Netconf.Hostname, hosts) {
					return errors.New("netconf: operation has to use a host defined in the configs section")
				}
			}

		}
	}
	return nil
}

func validateSSHConfig(ts *TestSuite) ([]string, error) {
	var hosts []string
	for idx := range ts.Configs {
		if ts.Configs[idx].Hostname == "" {
			return nil, errors.New("ssh config: hostname cannot be empty")
		}
		if ts.Configs[idx].Username == "" {
			return nil, errors.New("ssh config: username cannot be empty")
		}
		if ts.Configs[idx].Password == "" {
			return nil, errors.New("ssh config: password cannot be empty")
		}
		hosts = append(hosts, ts.Configs[idx].Hostname)
	}
	return hosts, nil
}

// StringInSlice helper function to test if a slice contains a value
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
