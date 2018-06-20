package main

import (
	"fmt"
	"log"

	//"github.com/pkg/profile"

	"github.com/damianoneill/nc-hammer/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	// https://flaviocopes.com/golang-profiling/

	// CPU profiling by default
	//defer profile.Start().Stop()

	// Memory profiling
	//defer profile.Start(profile.MemProfile).Stop()

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	cmd.Execute(fmt.Sprintf("%v, commit %v, built at %v", version, commit, date))
}
