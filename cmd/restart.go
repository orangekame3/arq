package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/orangekame3/arq/internal/server"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the running arq view server",
	Long:  `Restart the running arq view server with the new binary. Run "brew upgrade arq" before this command to apply updates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := os.ReadFile(server.ServerFilePath())
		if err != nil {
			return fmt.Errorf("no running arq view server found (missing %s)", server.ServerFilePath())
		}

		url := fmt.Sprintf("http://%s/api/restart", string(addr))
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(url, "", nil)
		if err != nil {
			return fmt.Errorf("failed to reach server at %s: %w", string(addr), err)
		}
		_ = resp.Body.Close()

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Server is restarting...")
		return nil
	},
}
