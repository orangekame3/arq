package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/orangekame3/arq/internal/ar5iv"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/orangekame3/arq/internal/summarize"
	"github.com/spf13/cobra"
)

var summarizeForce bool

var summarizeCmd = &cobra.Command{
	Use:     "summarize <query>",
	Aliases: []string{"sum"},
	Short:   "Generate a markdown summary of a paper",
	Long: `Generate a markdown summary of a paper using an LLM.

Fetches the paper's HTML from ar5iv for full-text analysis.
Falls back to abstract-based summary if ar5iv is unavailable.
Downloads figures to the paper's assets/ directory.

Configure the LLM provider in config:
  [summarize]
  provider = "anthropic"   # or "openai" (falls back to [translate] settings)
  model = "claude-sonnet-4-5-20241022"
  lang = "Japanese"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(cmd.ErrOrStderr())

		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}

		summaryPath := paper.SummaryPath(p)
		if !summarizeForce {
			if _, err := os.Stat(summaryPath); err == nil {
				logger.Warn("summary already exists", "path", summaryPath, "hint", "use --force to regenerate")
				return nil
			}
		}

		// Try ar5iv first, fall back to abstract
		var markdown string
		var figures []ar5iv.Figure

		logger.Info("fetching ar5iv HTML", "id", p.ID)
		content, err := ar5iv.Fetch(p.ID)
		if err != nil {
			logger.Warn("ar5iv unavailable, using abstract", "error", err)
			logger.Info("summarizing from abstract", "id", p.ID)
			markdown, err = summarize.SummarizeAbstract(p.Title, p.Abstract)
			if err != nil {
				return fmt.Errorf("summarize: %w", err)
			}
		} else {
			figures = content.Figures
			logger.Info("summarizing", "id", p.ID, "sections", len(content.Sections), "figures", len(figures))
			markdown, err = summarize.Summarize(content)
			if err != nil {
				return fmt.Errorf("summarize: %w", err)
			}
		}

		// Build final summary.md with metadata header
		var sb strings.Builder
		fmt.Fprintf(&sb, "# %s\n\n", p.Title)
		fmt.Fprintf(&sb, "- **Authors**: %s\n", strings.Join(p.Authors, ", "))
		fmt.Fprintf(&sb, "- **Published**: %s\n", p.Published)
		fmt.Fprintf(&sb, "- **Category**: %s\n", p.Category)
		fmt.Fprintf(&sb, "- **arXiv**: https://arxiv.org/abs/%s\n", p.ID)
		fmt.Fprintf(&sb, "- **PDF**: %s\n", p.PDFURL)
		sb.WriteString("\n---\n\n")
		sb.WriteString(markdown)
		sb.WriteString("\n")

		if err := os.WriteFile(summaryPath, []byte(sb.String()), 0o644); err != nil {
			return fmt.Errorf("write summary: %w", err)
		}
		logger.Info("saved summary", "path", summaryPath)

		// Download figures
		if len(figures) > 0 {
			downloaded := downloadFigures(logger, p, figures)
			if downloaded > 0 {
				logger.Info("downloaded figures", "count", downloaded)
			}
		}

		return nil
	},
}

func downloadFigures(logger *log.Logger, p *paper.Paper, figures []ar5iv.Figure) int {
	assetsDir := paper.AssetsDir(p)
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		logger.Warn("failed to create assets dir", "error", err)
		return 0
	}

	downloaded := 0
	for _, fig := range figures {
		if fig.URL == "" {
			continue
		}

		dest := filepath.Join(assetsDir, fig.Filename)

		resp, err := http.Get(fig.URL)
		if err != nil {
			logger.Warn("failed to download figure", "url", fig.URL, "error", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			continue
		}

		f, err := os.Create(dest)
		if err != nil {
			_ = resp.Body.Close()
			logger.Warn("failed to create file", "path", dest, "error", err)
			continue
		}

		_, err = io.Copy(f, resp.Body)
		_ = resp.Body.Close()
		_ = f.Close()
		if err != nil {
			logger.Warn("failed to write figure", "path", dest, "error", err)
			continue
		}

		downloaded++
	}
	return downloaded
}

func init() {
	summarizeCmd.Flags().BoolVar(&summarizeForce, "force", false, "Regenerate summary even if it already exists")
}
