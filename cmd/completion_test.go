package cmd

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// helper func to read from stdout, captures completion output
func readCompletionOutput(err error, args ...string) string {
	oldStdout := os.Stdout // keep backup of the real stdout
	read, write, _ := os.Pipe()
	os.Stdout = write
	completion(err, args...) // completion(...) output to be captured
	write.Close()
	out, _ := ioutil.ReadAll(read)
	os.Stdout = oldStdout
	got := strings.TrimSpace(string(out))
	return got
}

func Test_completion(t *testing.T) {
	t.Run("err is not nill", func(t *testing.T) {
		args := []string{"/etc/bash_completion"} // args triggers err in completion
		error, _ := CaptureStdout(Completion, myCmd, args)
		assert.Contains(t, error, "no such file or directory")
	})
	t.Run("err is nil", func(t *testing.T) {
		var err error
		got := readCompletionOutput(err, "/etc/bash_completion")
		want := "Bash completion file for " + RootCmd.Use + " saved to " + completionTarget
		if got != want {
			t.Errorf("have '%s' but want '%s'", got, want)
		}
	})

}
