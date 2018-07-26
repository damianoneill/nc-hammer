package cmd

import (
	"testing"
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
	args := []string{"v1.2", "v1.3"}
	got, _ := CaptureStdout(Version, myCmd, args)
	want := RootCmd.Use + " version"
	if want != got {
		t.Errorf("wanted '%s', but got '%s'", want, got)
	}
}
