package cmd

import (
	"fmt"

	"github.com/orangekame3/arq/internal/arxiv"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id-or-url>",
	Short: "Fetch a paper from arXiv",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := arxiv.NormalizeID(args[0])
		if err != nil {
			return err
		}

		// Check if already downloaded
		if _, err := paper.FindByID(id); err == nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "already exists: %s\n", id)
			return nil
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "fetching %s...\n", id)

		p, err := arxiv.Fetch(id)
		if err != nil {
			return err
		}

		if err := paper.Save(p); err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "downloading PDF...\n")
		if err := arxiv.DownloadPDF(p); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✔ added %s  %s\n", p.ID, p.Title)
		return nil
	},
}
