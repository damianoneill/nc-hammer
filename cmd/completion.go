package cmd

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
)

var completionTarget string

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion script for " + RootCmd.Use,
	Long: `Generates a shell completion script for ` + RootCmd.Use + `.
	NOTE: The current version supports Bash only.
		  This should work for *nix systems with Bash installed.

	By default, the file is written directly to /etc/bash_completion.d
	for convenience, and the command may need superuser rights, e.g.:

		$ sudo ` + RootCmd.Use + ` completion
	
	Add ` + "`--completionfile=/path/to/file`" + ` flag to set alternative
	file-path and name.

	For e.g. on OSX with bash completion installed with brew you should 

	$ ` + RootCmd.Use + ` completion --completionfile $(brew --prefix)/etc/bash_completion.d/` + RootCmd.Use + `.sh

	Logout and in again to reload the completion scripts,
	or just source them directly:

		$ . /etc/bash_completion
		
	or using if using brew
	
		$ . $(brew --prefix)/etc/bash_completion`,

	Run: Completion,
}

// Completion is a helper function to allow passing arguments to
// other functions (so that they can be unit tested)
func Completion(cmd *cobra.Command, args []string) {
	buffer := bytes.Buffer{}
	err := cmd.Root().GenBashCompletionFile(completionTarget)
	completion(&buffer, err, args...)
}

func completion(writer *bytes.Buffer, err error, args ...string) {
	if err != nil {
		fmt.Fprintf(writer, err.Error())
		return
	}
	fmt.Fprintf(writer, "Bash completion file for "+RootCmd.Use+" saved to "+completionTarget)
}

func init() {
	RootCmd.AddCommand(completionCmd)

	completionCmd.PersistentFlags().StringVarP(&completionTarget, "completionfile", "", "/etc/bash_completion.d/"+RootCmd.Use+".sh", "completion file")
	// Required for bash-completion
	_ = completionCmd.PersistentFlags().SetAnnotation("completionfile", cobra.BashCompFilenameExt, []string{})
}
