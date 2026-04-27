package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/search"
	"github.com/spf13/cobra"
)

var (
	searchField string
	searchJSON  bool
	searchID    bool
)

type searchResult struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Authors   []string `json:"authors"`
	Published string   `json:"published"`
	MatchedIn []string `json:"matched_in"`
}

var searchCmd = &cobra.Command{
	Use:   "search <keyword> [keyword...]",
	Short: "Search locally stored papers",
	Long: `Search across titles, abstracts, and summaries of locally stored papers.

All keywords must match (AND logic, case-insensitive).
Use --field to limit search to a specific field.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		papers, err := paper.List()
		if err != nil {
			return err
		}

		keywords := make([]string, len(args))
		for i, a := range args {
			keywords[i] = strings.ToLower(a)
		}

		var results []searchResult
		for _, p := range papers {
			matched, fields := search.Match(p, keywords, searchField)
			if matched {
				results = append(results, searchResult{
					ID:        p.ID,
					Title:     p.Title,
					Authors:   p.Authors,
					Published: p.Published,
					MatchedIn: fields,
				})
			}
		}

		if len(results) == 0 {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No papers found.")
			return nil
		}

		if searchJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(results)
		}

		if searchID {
			for _, r := range results {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), r.ID)
			}
			return nil
		}

		return printSearchTable(cmd, results)
	},
}

func printSearchTable(cmd *cobra.Command, results []searchResult) error {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	rows := make([][]string, len(results))
	for i, r := range results {
		rows[i] = []string{
			r.ID,
			truncate(r.Title, 45),
			strings.Join(r.MatchedIn, ","),
			r.Published,
		}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers("ID", "Title", "Matched In", "Published").
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

func init() {
	searchCmd.Flags().StringVar(&searchField, "field", "all", "Search field: title, abstract, keywords, summary, all")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output in JSON format")
	searchCmd.Flags().BoolVar(&searchID, "id", false, "Output IDs only")
}
