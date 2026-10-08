// version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time with ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)


var versionCmd = &cobra.Command{
	Use: "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("yaml2hcl %s (commit: %s, built: %s)\n", version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}