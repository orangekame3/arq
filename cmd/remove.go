package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <query>",
	Aliases: []string{"rm"},
	Short:   "Remove a paper from local storage",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(cmd.ErrOrStderr())

		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}

		dir := paper.Dir(p)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove %s: %w", dir, err)
		}

		logger.Info("removed", "id", p.ID, "title", p.Title)
		return nil
	},
}
