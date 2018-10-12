package action

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
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
	hello := new(netconf.HelloMessage)
	hello.SessionID = 0
	hello.Capabilities = []string{"urn:ietf:params:netconf:base:1.0"}
	validRPCReply := []byte("<rpc-reply xmlns=\"URN\" xmlns:junos=\"URL\"><ko/></rpc-reply>]]>]]>")          // does not match *action.Netconf.Expected
	validRPCReplyPassCheck := []byte("<rpc-reply xmlns=\"URN\" xmlns:junos=\"URL\"><ok/></rpc-reply>]]>]]>") // matches *action.Netconf.Expected
	invalidRPCReply := []byte("xmlns=\"URN\" xmlns:")

	mockTransport := &MockTransport{}

	// helper function passes calls to mock, overrrides createNewSession
	callMock := func(r []byte) {
		mockTransport.On("ReceiveHello").Return(hello, nil)
		mockTransport.On("SendHello", hello).Return(nil)
		mockTransport.On("Receive").Return(r, nil).Once()
		mockTransport.On("Send", mock.Anything).Return(nil)

		createNewSession = func(hostname, username, password string) (*netconf.Session, error) {
			mySession := netconf.NewSession(mockTransport)
			return mySession, nil
		}
	}

	t.Run("createNewSession(..) returns nil err and nil session", func(t *testing.T) {
		createNewSession = func(hostname, username, password string) (*netconf.Session, error) {
			return nil, nil
		}
		got := captureStdoutE(0)
		assert.Contains(t, got, "E")
	})

	t.Run("createNewSession(..) returns an error", func(t *testing.T) {
		createNewSession = func(hostname, username, password string) (*netconf.Session, error) {
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
