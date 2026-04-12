package cmd

import (
	"fmt"

	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/spf13/cobra"
)

var rootPathCmd = &cobra.Command{
	Use:   "root [path]",
	Short: "Show or set the root directory",
	Long: `Show or set the root directory for paper storage.

Without arguments, prints the current root.
With a path argument, saves it to the config file.

Priority: $ARQ_ROOT > config file > ~/papers`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), paper.Root())
			return nil
		}
		c := config.Load()
		c.Root = args[0]
		if err := config.Save(c); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "root set to %s\n", args[0])
		return nil
	},
}
