package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/orangekame3/arq/internal/arxiv"
	"github.com/orangekame3/arq/internal/paper"
	"github.com/spf13/cobra"
)

var hasCmd = &cobra.Command{
	Use:   "has <id> [...]",
	Short: "Check if papers exist locally (exit 0 if all found, 1 if any missing)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := collectHasIDs(args)
		if err != nil {
			return err
		}

		// Single ID mode: preserve original behavior (no output on success, error on not found)
		if len(ids) == 1 {
			if _, err := paper.FindByID(ids[0]); err != nil {
				return errors.New(ids[0])
			}
			return nil
		}

		// Batch mode: print found IDs, exit 1 if any missing
		missing := false
		for _, id := range ids {
			if _, err := paper.FindByID(id); err != nil {
				missing = true
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
			}
		}
		if missing {
			return errors.New("some papers not found")
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func collectHasIDs(args []string) ([]string, error) {
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
