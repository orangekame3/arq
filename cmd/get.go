package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/orangekame3/arq/internal/arxiv"
	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/translate"
	"github.com/spf13/cobra"
)

var (
	getTranslate   bool
	getNoTranslate bool
)

var getCmd = &cobra.Command{
	Use:   "get <id-or-url> [...]",
	Short: "Fetch papers from arXiv",
	Long: `Fetch papers from arXiv.

Accepts one or more arXiv IDs or URLs as arguments.
Use "-" to read IDs from stdin (one per line).
Use --translate to translate title and abstract to Japanese.`,
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
	if _, err := paper.FindByID(id); err == nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "already exists: %s\n", id)
		return nil
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "fetching %s...\n", id)

	p, err := arxiv.Fetch(id)
	if err != nil {
		return err
	}

	shouldTranslate := getTranslate || (config.Load().Translate.Enabled && !getNoTranslate)
	if shouldTranslate {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "translating...\n")
		result, err := translate.Translate(p.Title, p.Abstract)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⚠ translation failed: %s\n", err)
		} else {
			p.TitleJA = result.Title
			p.AbstractJA = result.Abstract
		}
	}

	if err := paper.Save(p); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "downloading PDF...\n")
	if err := arxiv.DownloadPDF(p); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✔ added %s  %s\n", p.ID, p.Title)
	return nil
}

func init() {
	getCmd.Flags().BoolVar(&getTranslate, "translate", false, "Translate title and abstract")
	getCmd.Flags().BoolVar(&getNoTranslate, "no-translate", false, "Skip translation even if enabled in config")
}
