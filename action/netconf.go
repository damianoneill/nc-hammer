package action

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/damianoneill/nc-hammer/result"
	"github.com/damianoneill/nc-hammer/suite"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/xml"
	"golang.org/x/crypto/ssh"
)

var gSessions map[string]*netconf.Session

func init() {
	gSessions = make(map[string]*netconf.Session)
}

// CloseAllSessions is called on exit to gracefully close the sockets
func CloseAllSessions() {
	// nolint
	for _, session := range gSessions {
		session.Close()
	}
}

// ExecuteNetconf invoked when a NETCONF Action is identified
func ExecuteNetconf(tsStart time.Time, cID int, action suite.Action, config *suite.Sshconfig, resultChannel chan result.NetconfResult) {

	var result result.NetconfResult
	result.Client = cID
	result.Hostname = action.Netconf.Hostname
	result.Operation = action.Netconf.Operation

	session, err := getSession(cID, config.Hostname+":"+strconv.Itoa(config.Port), config.Username, config.Password, config.Reuseconnection)
	if err != nil {
		fmt.Printf("E")
		result.Err = err.Error()
		resultChannel <- result
		return
	}

	// not reusing the connection, then explicitly close it
	if !config.Reuseconnection {
		// nolint
		defer session.Close()
	}

	if session != nil {
		result.SessionID = session.SessionID
	} else {
		fmt.Printf("E")
		result.Err = "session has expired"
		resultChannel <- result
		return
	}

	xmlNetconf, err := action.Netconf.ToXMLString()
	if err != nil {
		fmt.Printf("E")
		result.Err = err.Error()
		resultChannel <- result
		return
	}

	raw := netconf.RawMethod(xmlNetconf)
	result.Request = raw.MarshalMethod()
	start := time.Now()
	response, err := session.Exec(raw)
	if err != nil {
		if err.Error() == "WaitForFunc failed" {
			delete(gSessions, strconv.Itoa(cID)+config.Hostname+":"+strconv.Itoa(config.Port))
		}
		fmt.Printf("e")
		result.Err = "session closed by remote side"
		resultChannel <- result
		return
	}
	elapsed := time.Since(start)
	result.When = float64(time.Since(tsStart).Nanoseconds() / int64(time.Millisecond))
	result.Latency = float64(elapsed.Nanoseconds() / int64(time.Millisecond))

	m := minify.New()
	m.AddFunc("text/xml", xml.Minify)
	minified, err := m.String("text/xml", response.RawReply)
	if err != nil {
		fmt.Printf("E")
		result.Err = err.Error()
		resultChannel <- result
		return
	}
	result.Response = minified

	resultChannel <- result
}

// getSession returns a NETCONF Session, either a new one or a pre existing one if resuseConnection is valid for client/host
func getSession(client int, hostname string, username string, password string, reuseConnection bool) (*netconf.Session, error) {
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
		return session, nil
	}
	return createNewSession(hostname, username, password)
}

func createNewSession(hostname string, username string, password string) (*netconf.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return netconf.DialSSHTimeout(hostname, sshConfig, 30*time.Second)
}
