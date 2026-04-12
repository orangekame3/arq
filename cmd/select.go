package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactively select a paper with fzf",
	RunE: func(cmd *cobra.Command, args []string) error {
		papers, err := paper.List()
		if err != nil {
			return err
		}
		if len(papers) == 0 {
			return fmt.Errorf("no papers found")
		}

		// Build TSV input for fzf
		var input strings.Builder
		for _, p := range papers {
			fmt.Fprintf(&input, "%s\t%s\t%s\t%s\n",
				p.ID, p.Title, p.AuthorShort(), p.PublishedShort())
		}

		// Run fzf
		fzf := exec.Command("fzf", "--with-nth=2..")
		fzf.Stdin = strings.NewReader(input.String())
		fzf.Stderr = os.Stderr

		var out bytes.Buffer
		fzf.Stdout = &out

		if err := fzf.Run(); err != nil {
			return fmt.Errorf("fzf: %w", err)
		}

		// Extract ID (first field)
		line := strings.TrimSpace(out.String())
		if line == "" {
			return nil
		}
		id, _, _ := strings.Cut(line, "\t")
		fmt.Fprintln(cmd.OutOrStdout(), id)
		return nil
	},
}
