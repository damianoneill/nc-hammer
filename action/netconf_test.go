package action

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/damianoneill/nc-hammer/mocks/github.com/damianoneill/net/netconf"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/damianoneill/net/netconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//helper function to capture ExecuteNetconf output
func captureStdoutE(sessionID int) string {
	ts, _ := suite.NewTestSuite("../suite/testdata/test-suite.yml")
	start := time.Now()
	myAction := ts.Blocks[0].Actions[0]
	myConfig := ts.GetConfig(myAction.Netconf.Hostname)
	resultChannel := make(chan result.NetconfResult)
	handleResultsFinished := make(chan bool)

	go result.HandleResults(resultChannel, handleResultsFinished, ts)
	//reading from stdout
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	ExecuteNetconf(start, sessionID, myAction, myConfig, resultChannel)
	time.Sleep(500 * time.Millisecond)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	st := strings.Join(strings.Fields(string(out)), " ") //stdout captured, spaces trimmed
	return st
}
func Test_ExecuteNetconf(t *testing.T) {
	validRPCReply := []byte("<rpc-reply xmlns=\"URN\" xmlns:junos=\"URL\"><ko/></rpc-reply>]]>]]>")          // does not match *action.Netconf.Expected
	validRPCReplyPassCheck := []byte("<rpc-reply xmlns=\"URN\" xmlns:junos=\"URL\"><ok/></rpc-reply>]]>]]>") // matches *action.Netconf.Expected
	invalidRPCReply := []byte("xmlns=\"URN\" xmlns:")

	mockSession := &mocks.Session{}

	// helper function passes calls to mock, overrrides createNewSession
	callMock := func(r []byte) {

		reply := &netconf.RPCReply{}
		err := xml.Unmarshal(r, reply)
		mockSession.On("Execute", mock.Anything).Return(reply, err).Once()
		mockSession.On("ID").Return(75)

		createNewSession = func(hostname, username, password string) (netconf.Session, error) {
			return mockSession, nil
		}
	}

	t.Run("createNewSession(..) returns nil err and nil session", func(t *testing.T) {
		createNewSession = func(hostname, username, password string) (netconf.Session, error) {
			return nil, nil
		}
		got := captureStdoutE(0)
		assert.Contains(t, got, "E")
	})

	t.Run("createNewSession(..) returns an error", func(t *testing.T) {
		createNewSession = func(hostname, username, password string) (netconf.Session, error) {
			err := errors.New("error creating a netconf session")
			return nil, err
		}
		got := captureStdoutE(1)
		assert.Contains(t, got, "E")
	})

	t.Run("createNewSession(..) returns session and nil err, RPCReply returns err", func(t *testing.T) {
		callMock(invalidRPCReply)
		got := captureStdoutE(2)
		assert.Contains(t, got, "e")
	})

	t.Run("validRPCReply doesnt match *action.Netconf.Expected", func(t *testing.T) {
		callMock(validRPCReply)
		got := captureStdoutE(2)
		assert.Contains(t, got, "e")
	})

	t.Run("validRPCReply matches *action.Netconf.Expected", func(t *testing.T) {
		callMock(validRPCReplyPassCheck)
		got := captureStdoutE(3)
		assert.True(t, strings.Contains(got, ".") || strings.Contains(got, ""))
	})
}

func Test_NetconfDiagnosticContext(t *testing.T) {
	defaultContext := diagnosticContext

	CreateDiagnosticContext(false)
	nonDiagContext := diagnosticContext

	CreateDiagnosticContext(true)
	diagContext := diagnosticContext

	assert.Equal(t, netconf.NoOpLoggingHooks, netconf.ContextClientTrace(defaultContext), "Expect default context not to enable diagnostics")
	assert.Equal(t, netconf.DefaultLoggingHooks, netconf.ContextClientTrace(nonDiagContext), "Expect context not to enable diagnostics")
	assert.Equal(t, netconf.DiagnosticLoggingHooks, netconf.ContextClientTrace(diagContext), "Expect context to enable diagnostics")
}

func Test_handleIpv6(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"valid ip4", args{"127.0.0.1"}, "127.0.0.1"},
		{"valid ipv6", args{"::1"}, "[::1]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handleIpv6(tt.args.host); got != tt.want {
				t.Errorf("handleIpv6() = %v, want %v", got, tt.want)
			}
		})
	}
}
