package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"io"
	"net/http"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/orangekame3/arq/internal/ar5iv"
	"github.com/orangekame3/arq/internal/arxiv"
	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/keyword"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/summarize"
	"github.com/orangekame3/arq/internal/translate"
	"github.com/spf13/cobra"
)

var (
	getTranslate   bool
	getNoTranslate bool
	getSummarize   bool
	getNoSummarize bool
	getOpen        bool
	getForce       bool
)

var getCmd = &cobra.Command{
	Use:   "get <id-or-url> [...]",
	Short: "Fetch papers from arXiv",
	Long: `Fetch papers from arXiv.

Accepts one or more arXiv IDs or URLs as arguments.
Use "-" to read IDs from stdin (one per line).
Use --translate to translate title and abstract.
Use --summarize to generate a summary from ar5iv HTML.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := collectIDs(args)
		if err != nil {
			return err
		}

		for _, id := range ids {
			if err := fetchOne(cmd, id); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "✗ %s: %s\n", id, err)
			}
		}
		return nil
	},
}

func collectIDs(args []string) ([]string, error) {
	var ids []string
	for _, arg := range args {
		if arg == "-" {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				id, err := arxiv.NormalizeID(line)
				if err != nil {
					return nil, err
				}
				ids = append(ids, id)
			}
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("read stdin: %w", err)
			}
		} else {
			id, err := arxiv.NormalizeID(arg)
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func fetchOne(cmd *cobra.Command, id string) error {
	logger := log.New(cmd.ErrOrStderr())

	if existing, err := paper.FindByID(id); err == nil {
		if !getForce {
			logger.Warn("already exists", "id", id, "hint", "use --force to re-fetch")
			return nil
		}
		logger.Info("overwriting", "id", existing.ID)
	}

	logger.Info("fetching metadata", "id", id)

	p, err := arxiv.Fetch(id)
	if err != nil {
		return err
	}

	logger.Info("found", "title", p.Title, "category", p.Category)

	if err := paper.Save(p); err != nil {
		return err
	}

	logger.Info("downloading PDF", "id", id)
	if err := arxiv.DownloadPDF(p); err != nil {
		return err
	}

	shouldTranslate := getTranslate || (config.Load().Translate.Enabled && !getNoTranslate)
	if shouldTranslate {
		logger.Info("translating", "id", id)
		result, err := translate.Translate(p.Title, p.Abstract)
		if err != nil {
			logger.Warn("translation failed", "error", err)
		} else {
			p.TitleJA = result.Title
			p.AbstractJA = result.Abstract
			_ = paper.Save(p)
			logger.Info("translated", "title_ja", p.TitleJA)
		}
	}

	shouldSummarize := getSummarize || (config.Load().Summarize.Enabled && !getNoSummarize)
	if shouldSummarize {
		logger.Info("summarizing", "id", id)
		if err := summarizeOne(logger, p); err != nil {
			logger.Warn("summarization failed", "error", err)
		}
	}

	logger.Info("added", "id", p.ID, "path", paper.Dir(p))

	if getOpen {
		_ = openFile(paper.PDFPath(p))
	}
	return nil
}

func summarizeOne(logger *log.Logger, p *paper.Paper) error {
	content, err := ar5iv.Fetch(p.ID)
	if err != nil {
		logger.Warn("ar5iv unavailable, using abstract", "error", err)
		md, err := summarize.SummarizeAbstract(p.Title, p.Abstract)
		if err != nil {
			return err
		}
		return writeSummary(p, md)
	}

	logger.Info("ar5iv fetched", "sections", len(content.Sections), "figures", len(content.Figures))
	md, err := summarize.Summarize(content)
	if err != nil {
		return err
	}
	if err := writeSummary(p, md); err != nil {
		return err
	}

	// Download figures
	if len(content.Figures) > 0 {
		assetsDir := paper.AssetsDir(p)
		if err := os.MkdirAll(assetsDir, 0o755); err != nil {
			return nil
		}
		for _, fig := range content.Figures {
			if fig.URL == "" {
				continue
			}
			resp, err := http.Get(fig.URL)
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					_ = resp.Body.Close()
				}
				continue
			}
			f, err := os.Create(filepath.Join(assetsDir, fig.Filename))
			if err != nil {
				_ = resp.Body.Close()
				continue
			}
			_, _ = io.Copy(f, resp.Body)
			_ = resp.Body.Close()
			_ = f.Close()
		}
	}

	// Extract keywords
	if len(p.Keywords) == 0 {
		en, ja, err := keyword.Extract(p.Title, p.Abstract)
		if err != nil {
			logger.Warn("keyword extraction failed", "error", err)
		} else {
			p.Keywords = en
			p.KeywordsJA = ja
			_ = paper.Save(p)
		}
	}

	return nil
}

func writeSummary(p *paper.Paper, markdown string) error {
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
	return os.WriteFile(paper.SummaryPath(p), []byte(sb.String()), 0o644)
}

func init() {
	getCmd.Flags().BoolVar(&getTranslate, "translate", false, "Translate title and abstract")
	getCmd.Flags().BoolVar(&getNoTranslate, "no-translate", false, "Skip translation even if enabled in config")
	getCmd.Flags().BoolVar(&getSummarize, "summarize", false, "Generate summary from ar5iv HTML")
	getCmd.Flags().BoolVar(&getNoSummarize, "no-summarize", false, "Skip summarization even if enabled in config")
	getCmd.Flags().BoolVar(&getOpen, "open", false, "Open PDF after download")
	getCmd.Flags().BoolVar(&getForce, "force", false, "Re-fetch even if already exists")
}
