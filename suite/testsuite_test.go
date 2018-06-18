package suite_test

import (
	"reflect"
	"testing"

	"github.com/damianoneill/nc-hammer/suite"
	"github.com/stretchr/testify/assert"
)

// func init() {
// 	suite := ts.TestSuite{}
// 	suite.Clients = 2
// 	suite.Iterations = 10
// 	suite.Rampup = 0
// 	suite.Configs = []ts.Sshconfig{
// 		ts.Sshconfig{Hostname: "10.0.0.1", Port: 830, Username: "uname", Password: "pass", Reuseconnection: false},
// 		ts.Sshconfig{Hostname: "10.0.0.2", Port: 830, Username: "uname", Password: "pass", Reuseconnection: true},
// 	}
// 	config := "<top xmlns=\"http://example.com/schema/1.2/config\"><protocols><ospf><area><name>0.0.0.0</name><interfaces><interface xc:operation=\"delete\"><name>192.0.2.4</name></interface></interfaces></area></ospf></protocols></top>"
// 	target := "running"
// 	editConfig := ts.Netconf{Hostname: "10.0.0.1", Operation: "edit-config", Source: nil, Target: &target, Filter: nil, Config: &config}
// 	netconfAction := ts.Action{Netconf: &editConfig, Sleep: nil}
// 	initBlock := ts.Block{Type: "init", Actions: []ts.Action{netconfAction}}

// 	get := ts.Netconf{Hostname: "10.0.0.1", Operation: "get", Source: nil, Target: nil, Filter: nil, Config: nil}
// 	getWithFilter := ts.Netconf{Hostname: "10.0.0.2", Operation: "get", Source: nil, Target: nil, Filter: &ts.Filter{Type: "subtree", Ns: nil, Select: "<users/>"}, Config: nil}
// 	getAction := ts.Action{Netconf: &get, Sleep: nil}
// 	getWithFilterAction := ts.Action{Netconf: &getWithFilter, Sleep: nil}
// 	concurrentBlock := ts.Block{Type: "concurrent", Actions: []ts.Action{getAction, getWithFilterAction}}

// 	getConfig := ts.Netconf{Hostname: "10.0.0.1", Operation: "get-config", Source: nil, Target: nil, Filter: nil, Config: nil}
// 	source := "running"
// 	getConfigSource := ts.Netconf{Hostname: "10.0.0.1", Operation: "get-config", Source: &source, Target: nil, Filter: nil, Config: nil}
// 	getConfigAction := ts.Action{Netconf: &getConfig, Sleep: nil}
// 	getConfigWithSourceAction := ts.Action{Netconf: &getConfigSource, Sleep: nil}
// 	sleep := ts.Sleep{Duration: 5}
// 	sleepAction := ts.Action{Netconf: nil, Sleep: &sleep}
// 	sequentialBlock := ts.Block{Type: "sequential", Actions: []ts.Action{getConfigAction, sleepAction, getConfigWithSourceAction}}
// 	suite.Blocks = []ts.Block{initBlock, concurrentBlock, sequentialBlock}

// 	data, err := json.Marshal(suite)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("\n%s\n\n", data)
// 	data, err = yaml.Marshal(suite)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("\n%s\n\n", data)
// }

func TestNetconf_ToXMLString(t *testing.T) {
	candidate := "candidate"
	editOperation := "<top xmlns=\"http://example.com/schema/1.2/config\"><interface><name>Ethernet0/0</name><mtu>1500</mtu></interface>"
	ns := "urn:ietf:params:xml:ns:netconf:base:1.0"
	filter := suite.Filter{Type: "type", Ns: nil, Select: "<select/>"}
	filterWithNs := suite.Filter{Type: "type", Ns: &ns, Select: "<select/>"}
	type fields struct {
		Hostname  string
		Operation string
		Source    *string
		Target    *string
		Filter    *suite.Filter
		Config    *string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"valid get-config", fields{"hostname", "get-config", nil, nil, nil, nil}, "<get-config><source><running/></source></get-config>", false},
		{"valid get-config candidate source", fields{"hostname", "get-config", &candidate, nil, nil, nil}, "<get-config><source><candidate/></source></get-config>", false},
		{"valid get-config filter", fields{"hostname", "get-config", nil, nil, &filter, nil}, "<get-config><source><running/></source><filter type=\"type\"><top><select/></top></filter></get-config>", false},
		{"valid get-config filter with ns", fields{"hostname", "get-config", nil, nil, &filterWithNs, nil}, "<get-config><source><running/></source><filter type=\"type\"><top xmlns=\"urn:ietf:params:xml:ns:netconf:base:1.0\"><select/></top></filter></get-config>", false},
		{"not supported kill-session", fields{"hostname", "kill-session", nil, nil, nil, nil}, "", true},
		{"valid get", fields{"hostname", "get", nil, nil, nil, nil}, "<get/>", false},
		{"valid edit-config", fields{"hostname1", "edit-config", nil, nil, nil, nil}, "<edit-config><target><running/></target><config/></edit-config>", false},
		{"valid edit-config2", fields{"hostname2", "edit-config", nil, &candidate, nil, &editOperation}, "<edit-config><target><candidate/></target><config><top xmlns=\"http://example.com/schema/1.2/config\"><interface><name>Ethernet0/0</name><mtu>1500</mtu></interface></top></config></edit-config>", false},
		{"valid get with filter", fields{"hostname", "get", nil, nil, &filter, nil}, "<get><filter type=\"type\"><top><select/></top></filter></get>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &suite.Netconf{
				Hostname:  tt.fields.Hostname,
				Operation: tt.fields.Operation,
				Source:    tt.fields.Source,
				Target:    tt.fields.Target,
				Filter:    tt.fields.Filter,
				Config:    tt.fields.Config,
			}
			got, err := n.ToXMLString()
			if (err != nil) != tt.wantErr {
				t.Errorf("Netconf.ToXMLString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Netconf.ToXMLString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTestSuite(t *testing.T) {
	emptyTs := suite.TestSuite{}
	emptyTs.File = "testdata/emptytestsuite.yml"
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    *suite.TestSuite
		wantErr bool
	}{
		// TODO: Add test cases.
		{"file not present", args{"doesnt-exist.txt"}, nil, true},
		{"file present, no content", args{"testdata/emptytestsuite.yml"}, &emptyTs, true},
		//{"valid test suite", args{"testdata/testsuite.yml"}, &emptyTs, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := suite.NewTestSuite(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTestSuite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTestSuite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTestSuite_GetConfig(t *testing.T) {
	got, err := suite.NewTestSuite("testdata/testsuite.yml")
	if err != nil {
		t.Errorf("Problem loading testdata/testsuite.yml: %v", err)
	}
	if got.GetConfig("10.0.0.1") == nil {
		t.Errorf("Should have returned a valid configuration for hostname: %v", "10.0.0.1")
	}
}

func TestTestSuite_GetInitBlock(t *testing.T) {
	got, err := suite.NewTestSuite("testdata/testsuite.yml")
	if err != nil {
		t.Errorf("Problem loading testdata/testsuite.yml: %v", err)
	}
	block := got.GetInitBlock()
	assert.NotNil(t, block)
	assert.Equal(t, block.Actions[0].Netconf.Operation, "edit-config", "they should be equal")
}
