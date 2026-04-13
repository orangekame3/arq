package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var viewTarget string

var viewCmd = &cobra.Command{
	Use:   "view <query>",
	Short: "Open a paper's summary in mo (Markdown viewer)",
	Long: `Open a paper's summary.md in mo for browser-based viewing.

Requires mo (https://github.com/k1LoW/mo) to be installed.
If a mo server is already running, the file is added to the existing session.

Use --target to organize papers into named groups:
  arq view 2303.12345 --target reads`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}

		summaryPath := paper.SummaryPath(p)
		if _, err := os.Stat(summaryPath); err != nil {
			return fmt.Errorf("summary not found: run 'arq summarize %s' first", p.ID)
		}

		if _, err := exec.LookPath("mo"); err != nil {
			return fmt.Errorf("mo is not installed. Install it with: brew install k1LoW/tap/mo")
		}

		moArgs := []string{summaryPath, "--open"}
		if viewTarget != "" {
			moArgs = append(moArgs, "--target", viewTarget)
		}

		moCmd := exec.Command("mo", moArgs...)
		moCmd.Stdout = cmd.OutOrStdout()
		moCmd.Stderr = cmd.ErrOrStderr()
		return moCmd.Run()
	},
}

func init() {
	viewCmd.Flags().StringVarP(&viewTarget, "target", "t", "arq", "mo group name")
}
