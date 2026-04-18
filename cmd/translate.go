package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/orangekame3/arq/internal/translate"
	"github.com/spf13/cobra"
)

var (
	translateForce bool
	translateAll   bool
)

var translateCmd = &cobra.Command{
	Use:   "translate [query]",
	Short: "Translate a paper's title and abstract",
	Long: `Translate a paper's title and abstract using an LLM.

Use --all to translate all papers that have not been translated yet.
Use --force to re-translate even if already translated.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(cmd.ErrOrStderr())

		if translateAll && len(args) > 0 {
			return fmt.Errorf("--all and positional arguments are mutually exclusive")
		}
		if !translateAll && len(args) == 0 {
			return fmt.Errorf("requires a query argument or --all flag")
		}

		if translateAll {
			return translateAllPapers(logger)
		}

		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}
		return translatePaper(logger, p)
	},
}

func translatePaper(logger *log.Logger, p *paper.Paper) error {
	if !translateForce && p.TitleJA != "" {
		logger.Warn("already translated", "id", p.ID, "hint", "use --force to re-translate")
		return nil
	}

	logger.Info("translating", "id", p.ID)
	result, err := translate.Translate(p.Title, p.Abstract)
	if err != nil {
		return fmt.Errorf("translate: %w", err)
	}

	p.TitleJA = result.Title
	p.AbstractJA = result.Abstract
	if err := paper.Save(p); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	logger.Info("translated", "id", p.ID, "title_ja", p.TitleJA)
	return nil
}

func translateAllPapers(logger *log.Logger) error {
	papers, err := paper.List()
	if err != nil {
		return err
	}

	var translated, skipped, failed int
	total := len(papers)

	for i, p := range papers {
		if !translateForce && p.TitleJA != "" {
			skipped++
			continue
		}

		logger.Info("translating", "progress", fmt.Sprintf("%d/%d", i+1, total), "id", p.ID)
		if err := translatePaper(logger, p); err != nil {
			logger.Warn("failed", "id", p.ID, "error", err)
			failed++
		} else {
			translated++
		}
	}

	logger.Info("done", "translated", translated, "skipped", skipped, "failed", failed)
	return nil
}

func init() {
	translateCmd.Flags().BoolVar(&translateForce, "force", false, "Re-translate even if already translated")
	translateCmd.Flags().BoolVar(&translateAll, "all", false, "Translate all papers without a translation")
}
