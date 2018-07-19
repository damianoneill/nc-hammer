package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// helper func to read from stdout
func readMe(err error, args ...string) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	completion(err, args...)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	got := strings.TrimSpace(string(out[:]))
	return got

}
func Test_completion(t *testing.T) {
	var err error
	t.Run("err is nil", func(t *testing.T) {
		got := readMe(err, "arg1", "arg2")
		want := "Bash completion file for " + RootCmd.Use + " saved to " + completionTarget
		if got != want {
			t.Errorf("have '%s' but want '%s'", got, want)
		}

	})
	t.Run("error is not nil", func(t *testing.T) {
		var err = errors.New("my error")
		got := readMe(err, "arg1", "arg2")
		want := err.Error()
		if got != want {
			t.Errorf("have '%s' but want '%s'", got, want)
		}
	})
}
func Test_Completion(t *testing.T) {
	var err error
	args := []string{""}
	Completion(completionCmd, args)
	if err != nil {
		t.Error()
	}
}
