package cmd

import (
	"errors"

	"github.com/orangekame3/arq/internal/arxiv"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/spf13/cobra"
)

var hasCmd = &cobra.Command{
	Use:   "has <id>",
	Short: "Check if a paper exists locally (exit 0 if yes, 1 if no)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := arxiv.NormalizeID(args[0])
		if err != nil {
			return err
		}
		if _, err := paper.FindByID(id); err != nil {
			return errors.New(id)
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}
