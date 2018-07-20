package cmd

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_version(t *testing.T) {
	type args struct {
		command string
		args    []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"name", args{"skeleton", nil}, "skeleton version "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := version(tt.args.command, tt.args.args...); got != tt.want {
				t.Errorf("version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Version(t *testing.T) {
	myCmd := &cobra.Command{}
	args := []string{"arg1", "arg2"}
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Version(myCmd, args)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	got := strings.TrimSpace(string(out))
	want := RootCmd.Use + " version"
	if want != got {
		t.Errorf("wanted '%s', but got '%s'", want, got)
	}
}
