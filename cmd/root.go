package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arq",
	Short: "Local arXiv paper index for fzf-driven exploration",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(rootPathCmd)
	rootCmd.AddCommand(hasCmd)
	rootCmd.AddCommand(versionCmd)
}
