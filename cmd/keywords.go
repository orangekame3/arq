package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/orangekame3/arq/internal/keyword"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var (
	keywordsForce bool
	keywordsAll   bool
)

var keywordsCmd = &cobra.Command{
	Use:   "keywords [query]",
	Short: "Extract search keywords from a paper",
	Long: `Extract bilingual (English/Japanese) keywords from a paper's title and abstract using an LLM.

Keywords are saved to meta.json and used for fzf fuzzy matching in arq select.

Use --all to extract keywords for all papers that don't have them yet.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(cmd.ErrOrStderr())

		if keywordsAll && len(args) > 0 {
			return fmt.Errorf("--all and positional arguments are mutually exclusive")
		}
		if !keywordsAll && len(args) == 0 {
			return fmt.Errorf("requires a query argument or --all flag")
		}

		if keywordsAll {
			return extractAllKeywords(logger)
		}

		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}
		return extractKeywords(logger, p)
	},
}

func extractKeywords(logger *log.Logger, p *paper.Paper) error {
	if len(p.Keywords) > 0 && !keywordsForce {
		logger.Warn("keywords already exist", "id", p.ID, "hint", "use --force to re-extract")
		return nil
	}

	logger.Info("extracting keywords", "id", p.ID)
	en, ja, err := keyword.Extract(p.Title, p.Abstract)
	if err != nil {
		return fmt.Errorf("extract keywords: %w", err)
	}

	p.Keywords = en
	p.KeywordsJA = ja
	if err := paper.Save(p); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	logger.Info("saved keywords", "id", p.ID, "en", en, "ja", ja)
	return nil
}

func extractAllKeywords(logger *log.Logger) error {
	papers, err := paper.List()
	if err != nil {
		return err
	}

	var extracted, skipped, failed int
	total := len(papers)

	for i, p := range papers {
		if len(p.Keywords) > 0 && !keywordsForce {
			skipped++
			continue
		}

		logger.Info("extracting", "progress", fmt.Sprintf("%d/%d", i+1, total), "id", p.ID)
		if err := extractKeywords(logger, p); err != nil {
			logger.Warn("failed", "id", p.ID, "error", err)
			failed++
		} else {
			extracted++
		}
	}

	logger.Info("done", "extracted", extracted, "skipped", skipped, "failed", failed)
	return nil
}

func init() {
	keywordsCmd.Flags().BoolVar(&keywordsForce, "force", false, "Re-extract keywords even if they already exist")
	keywordsCmd.Flags().BoolVar(&keywordsAll, "all", false, "Extract keywords for all papers without existing keywords")
}
