package main

import (
	"fmt"
	"log"

	"github.com/damianoneill/nc-hammer/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	cmd.Execute(fmt.Sprintf("%v, commit %v, built at %v", version, commit, date))
}
