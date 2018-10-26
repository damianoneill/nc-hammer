package action

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/damianoneill/net/netconf"
	"golang.org/x/crypto/ssh"
)

var (
	diagnosticContext = context.Background()
)

const (
	colon = ":"
	left  = "["
	right = "]"
)

// CreateDiagnosticContext creates a context used for instantiating new Netconf sessions, with the option
// of enabling netconf client diagnostics (diagFlag == true).
func CreateDiagnosticContext(diagFlag bool) {
	var trace *netconf.ClientTrace
	if diagFlag {
		trace = netconf.DiagnosticLoggingHooks
	} else {
		trace = netconf.DefaultLoggingHooks
	}
	diagnosticContext = netconf.WithClientTrace(diagnosticContext, trace)
}

var gSessions map[string]netconf.Session

func init() {
	gSessions = make(map[string]netconf.Session)
}

// CloseAllSessions is called on exit to gracefully close the sockets
func CloseAllSessions() {
	// nolint
	for _, session := range gSessions {
		session.Close()
	}
}

func operationOrMessage(netconf *suite.Netconf) string {
	if netconf.Operation != nil {
		return *netconf.Operation
	}
	return *netconf.Message
}

// ExecuteNetconf invoked when a NETCONF Action is identified
func ExecuteNetconf(tsStart time.Time, cID int, action suite.Action, config *suite.Sshconfig, resultChannel chan result.NetconfResult) {

	var r result.NetconfResult
	r.Client = cID
	r.Hostname = action.Netconf.Hostname
	r.Operation = operationOrMessage(action.Netconf)

	session, err := getSession(cID, handleIpv6(config.Hostname)+":"+strconv.Itoa(config.Port), config.Username, config.Password, config.Reuseconnection)
	if err != nil {
		fmt.Printf("E")
		r.Err = err.Error()
		resultChannel <- r
		return
	}

	// not reusing the connection, then explicitly close it
	if !config.Reuseconnection {
		// nolint
		defer session.Close()
	}

	if session != nil {
		r.SessionID = session.ID()
	} else {
		fmt.Printf("E")
		r.Err = "session has expired"
		resultChannel <- r
		return
	}

	xml, err := action.Netconf.ToXMLString()
	if err != nil {
		fmt.Printf("E")
		r.Err = err.Error()
		resultChannel <- r
		return
	}

	raw := netconf.Request(xml)
	start := time.Now()
	rpcReply, err := session.Execute(raw)
	if err != nil {
		r.Err = err.Error()
		fmt.Printf("e")
		resultChannel <- r
		return
	}
	elapsed := time.Since(start)
	r.When = float64(time.Since(tsStart).Nanoseconds() / int64(time.Millisecond))
	r.Latency = float64(elapsed.Nanoseconds() / int64(time.Millisecond))

	r.MessageID = rpcReply.MessageID

	if action.Netconf.Expected != nil {
		match, err := regexp.MatchString(*action.Netconf.Expected, rpcReply.Data)
		if err != nil {
			fmt.Printf("E")
			r.Err = err.Error()
			resultChannel <- r
			return
		}
		if !match {
			fmt.Printf("e")
			r.Err = "expected response did not match, expected: " + *action.Netconf.Expected + " actual: " + rpcReply.Data
			resultChannel <- r
			return
		}
	}
	resultChannel <- r
}

// getSession returns a NETCONF Session, either a new one or a pre existing one if resuseConnection is valid for client/host
func getSession(client int, hostname, username, password string, reuseConnection bool) (netconf.Session, error) {
	// check if hostname should reuse connection
	if reuseConnection {
		// get Session from Map if present
		session, present := gSessions[strconv.Itoa(client)+hostname]
		if present {
			return session, nil
		}
		// not present in map, therefore first time its called, create a new session and store in map
		session, err := createNewSession(hostname, username, password)
		if err == nil {
			gSessions[strconv.Itoa(client)+hostname] = session
		}
		return session, err
	}
	return createNewSession(hostname, username, password)
}

var createNewSession = func(hostname, username, password string) (netconf.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return netconf.NewRPCSession(diagnosticContext, sshConfig, hostname)
}

func handleIpv6(host string) string {
	if strings.Contains(host, colon) {
		return left + host + right
	}
	return host
}
