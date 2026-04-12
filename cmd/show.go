package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var showJSON bool

var showCmd = &cobra.Command{
	Use:   "show <query>",
	Short: "Show paper details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}

		if showJSON {
			type showEntry struct {
				ID         string   `json:"id"`
				Title      string   `json:"title"`
				TitleJA    string   `json:"title_ja,omitempty"`
				Authors    []string `json:"authors"`
				Abstract   string   `json:"abstract"`
				AbstractJA string   `json:"abstract_ja,omitempty"`
				Published  string   `json:"published"`
				Category   string   `json:"category"`
				PDFURL     string   `json:"pdf_url"`
				Path       string   `json:"path"`
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(showEntry{
				ID:         p.ID,
				Title:      p.Title,
				TitleJA:    p.TitleJA,
				Authors:    p.Authors,
				Abstract:   p.Abstract,
				AbstractJA: p.AbstractJA,
				Published:  p.Published,
				Category:   p.Category,
				PDFURL:     p.PDFURL,
				Path:       paper.PDFPath(p),
			})
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Title:     %s\n", p.Title)
		if p.TitleJA != "" {
			_, _ = fmt.Fprintf(w, "Title(ja): %s\n", p.TitleJA)
		}
		_, _ = fmt.Fprintf(w, "Authors:   %s\n", p.AuthorShort())
		_, _ = fmt.Fprintf(w, "Published: %s\n", p.Published)
		_, _ = fmt.Fprintf(w, "Category:  %s\n", p.Category)
		_, _ = fmt.Fprintf(w, "Path:      %s\n", paper.PDFPath(p))
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "--- Abstract ---")
		_, _ = fmt.Fprintln(w, p.Abstract)
		if p.AbstractJA != "" {
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "--- Abstract (ja) ---")
			_, _ = fmt.Fprintln(w, p.AbstractJA)
		}
		return nil
	},
}

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output in JSON format")
}
