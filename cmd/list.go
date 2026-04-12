package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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

		if listTSV {
			for _, p := range papers {
				fields := []string{
					p.ID,
					p.Title,
					p.AuthorShort(),
					p.PublishedShort(),
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.Join(fields, "\t"))
			}
			return nil
		}

		return printTable(cmd, papers)
	},
}

func truncate(s string, max int) string {
	if max <= 3 || len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func printTable(cmd *cobra.Command, papers []*paper.Paper) error {
	if len(papers) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No papers found.")
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	rows := make([][]string, len(papers))
	for i, p := range papers {
		rows[i] = []string{p.ID, truncate(p.Title, 50), truncate(p.AuthorShort(), 25), p.PublishedShort()}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("ID", "Title", "Authors", "Published").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			switch col {
			case 0:
				return idStyle
			case 3:
				return dimStyle
			default:
				return lipgloss.NewStyle()
			}
		})

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), t)
	return nil
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
