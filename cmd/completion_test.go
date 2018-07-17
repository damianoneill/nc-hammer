package cmd

import (
	"bytes"
	"errors"
	"testing"
)

func Test_completion(t *testing.T) {
	buffer := bytes.Buffer{}

	t.Run("error is nil", func(t *testing.T) {
		var err = errors.New("my error ")
		completion(&buffer, err, "arg1", "arg2")
		have := buffer.String()
		want := err.Error()
		if have != want {
			t.Errorf("have '%s' but want '%s'", have, want)
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
