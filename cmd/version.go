package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show " + RootCmd.Use + " version",
	Run:   Version,
}

// Version is a helper function to allow passing arguments to
// other functions (so that they can be unit tested)
func Version(cmd *cobra.Command, args []string) {
	fmt.Println(version(RootCmd.Use, args...))
}

func version(command string, args ...string) string {
	return command + " version " + VERSION
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
