package cmd

import (
	"fmt"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <query>",
	Short: "Print the PDF path for a paper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), paper.PDFPath(p))
		return nil
	},
}
