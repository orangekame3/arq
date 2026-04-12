package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/spf13/cobra"
)

var (
	listTSV  bool
	listJSON bool
	listID   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all papers",
	RunE: func(cmd *cobra.Command, args []string) error {
		papers, err := paper.List()
		if err != nil {
			return err
		}

		if listJSON {
			return printJSON(cmd, papers)
		}

		if listID {
			for _, p := range papers {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), p.ID)
			}
			return nil
		}

		for _, p := range papers {
			sep := "\t"
			if !listTSV {
				sep = "  "
			}
			fields := []string{
				p.ID,
				p.Title,
				p.AuthorShort(),
				p.PublishedShort(),
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.Join(fields, sep))
		}
		return nil
	},
}

type listEntry struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Authors   []string `json:"authors"`
	Published string   `json:"published"`
}

func printJSON(cmd *cobra.Command, papers []*paper.Paper) error {
	entries := make([]listEntry, len(papers))
	for i, p := range papers {
		entries[i] = listEntry{
			ID:        p.ID,
			Title:     p.Title,
			Authors:   p.Authors,
			Published: p.Published,
		}
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func init() {
	listCmd.Flags().BoolVar(&listTSV, "tsv", false, "Output in TSV format")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	listCmd.Flags().BoolVar(&listID, "id", false, "Output IDs only")
}
