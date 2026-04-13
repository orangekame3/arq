package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/orangekame3/arq/internal/pager"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var (
	showJSON    bool
	showSummary bool
)

var (
	titleStyle     = lipgloss.NewStyle().Bold(true)
	titleJAStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	labelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sectionStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	sectionJAStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	dividerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
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
				Thumbnail  string   `json:"thumbnail,omitempty"`
				Summary    string   `json:"summary,omitempty"`
			}
			entry := showEntry{
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
				Thumbnail:  paper.ThumbnailPath(p),
			}
			if summaryPath := paper.SummaryPath(p); summaryPath != "" {
				if data, err := os.ReadFile(summaryPath); err == nil {
					entry.Summary = string(data)
				}
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(entry)
		}

		// Summary-only mode: render markdown with pager
		if showSummary {
			summaryPath := paper.SummaryPath(p)
			data, err := os.ReadFile(summaryPath)
			if err != nil {
				return fmt.Errorf("summary not found: run 'arq summarize %s' first", p.ID)
			}
			if isatty.IsTerminal(os.Stdout.Fd()) {
				return pager.RunMarkdown(string(data))
			}
			// Non-TTY: render with glamour and write directly
			var buf bytes.Buffer
			if err := renderSummary(p, &buf); err != nil {
				return err
			}
			_, writeErr := cmd.OutOrStdout().Write(buf.Bytes())
			return writeErr
		}

		// Buffer output for pager support
		var buf bytes.Buffer
		w := io.Writer(&buf)

		// Show thumbnail if TTY and chafa is available
		if thumb := paper.ThumbnailPath(p); thumb != "" && isatty.IsTerminal(os.Stdout.Fd()) {
			if chafa, err := exec.LookPath("chafa"); err == nil {
				c := exec.Command(chafa, "--format=kitty", "--size=40x15", thumb)
				c.Stdout = cmd.OutOrStdout()
				_ = c.Run()
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}
		}

		divider := dividerStyle.Render(strings.Repeat("─", 50))

		_, _ = fmt.Fprintln(w, titleStyle.Render(p.Title))
		if p.TitleJA != "" {
			_, _ = fmt.Fprintln(w, titleJAStyle.Render(p.TitleJA))
		}
		_, _ = fmt.Fprintln(w)

		_, _ = fmt.Fprintf(w, "%s  %s\n", labelStyle.Render("Authors  "), p.AuthorShort())
		_, _ = fmt.Fprintf(w, "%s  %s\n", labelStyle.Render("Published"), p.Published)
		_, _ = fmt.Fprintf(w, "%s  %s\n", labelStyle.Render("Category "), p.Category)
		_, _ = fmt.Fprintf(w, "%s  %s\n", labelStyle.Render("Path     "), paper.PDFPath(p))
		if thumb := paper.ThumbnailPath(p); thumb != "" {
			_, _ = fmt.Fprintf(w, "%s  %s\n", labelStyle.Render("Thumbnail"), thumb)
		}
		if _, err := os.Stat(paper.SummaryPath(p)); err == nil {
			_, _ = fmt.Fprintf(w, "%s  %s\n", labelStyle.Render("Summary  "), paper.SummaryPath(p))
		}

		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, divider)
		_, _ = fmt.Fprintln(w, sectionStyle.Render("Abstract"))
		_, _ = fmt.Fprintln(w, divider)
		_, _ = fmt.Fprintln(w, p.Abstract)

		if p.AbstractJA != "" {
			_, _ = fmt.Fprintln(w)
			_, _ = fmt.Fprintln(w, divider)
			_, _ = fmt.Fprintln(w, sectionJAStyle.Render("要旨"))
			_, _ = fmt.Fprintln(w, divider)
			_, _ = fmt.Fprintln(w, p.AbstractJA)
		}

		// Append rendered summary if it exists
		if _, err := os.Stat(paper.SummaryPath(p)); err == nil {
			_, _ = fmt.Fprintln(w)
			_, _ = fmt.Fprintln(w, divider)
			_, _ = fmt.Fprintln(w, sectionStyle.Render("Summary"))
			_, _ = fmt.Fprintln(w, divider)
			if err := renderSummary(p, w); err != nil {
				_, _ = fmt.Fprintf(w, "(failed to render summary: %s)\n", err)
			}
		}

		// TTY: use viewport pager, non-TTY: write directly
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return pager.Run(buf.String())
		}
		_, writeErr := cmd.OutOrStdout().Write(buf.Bytes())
		return writeErr
	},
}

func renderSummary(p *paper.Paper, w io.Writer) error {
	summaryPath := paper.SummaryPath(p)
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return fmt.Errorf("summary not found: run 'arq summarize %s' first", p.ID)
	}

	var styleOpt glamour.TermRendererOption
	if isatty.IsTerminal(os.Stdout.Fd()) {
		styleOpt = glamour.WithAutoStyle()
	} else {
		styleOpt = glamour.WithStandardStyle("notty")
	}

	renderer, err := glamour.NewTermRenderer(
		styleOpt,
		glamour.WithWordWrap(80),
	)
	if err != nil {
		_, _ = w.Write(data)
		return nil
	}

	rendered, err := renderer.Render(string(data))
	if err != nil {
		_, _ = w.Write(data)
		return nil
	}

	_, _ = fmt.Fprint(w, rendered)
	return nil
}

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output in JSON format")
	showCmd.Flags().BoolVar(&showSummary, "summary", false, "Show only the rendered summary")
}
