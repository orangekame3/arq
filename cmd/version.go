package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current version of arq.
// tagpr will update this value automatically.
var Version = "0.0.17"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "arq version "+Version)
	},
}
