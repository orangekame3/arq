package cmd

import (
	"fmt"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <query>",
	Short: "Show paper details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		fmt.Fprintf(w, "Title:     %s\n", p.Title)
		fmt.Fprintf(w, "Authors:   %s\n", p.AuthorShort())
		fmt.Fprintf(w, "Published: %s\n", p.Published)
		fmt.Fprintf(w, "Category:  %s\n", p.Category)
		fmt.Fprintf(w, "Abstract:\n%s\n", p.Abstract)
		fmt.Fprintf(w, "Path:      %s\n", paper.PDFPath(p))
		return nil
	},
}
