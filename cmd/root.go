package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arq",
	Short: "Manage arXiv papers in your terminal",
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
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(thumbnailCmd)
	rootCmd.AddCommand(summarizeCmd)
	rootCmd.AddCommand(translateCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upgradeCmd)
}
