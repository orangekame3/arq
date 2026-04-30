package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/query"
	"github.com/orangekame3/arq/internal/server"
	"github.com/spf13/cobra"
)

var (
	viewListen string
	viewNoOpen bool
)

var viewCmd = &cobra.Command{
	Use:   "view [query]",
	Short: "Open paper library in browser",
	Long: `Launch a browser-based paper viewer.

Without arguments, opens the full library.
With a query, navigates to the matching paper.

Use --listen to bind to a specific address (e.g. 0.0.0.0:8080) for remote access.
Use --no-open to suppress automatic browser opening (useful for headless servers).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var initialPaperID string
		if len(args) > 0 {
			p, err := query.Resolve(args[0])
			if err != nil {
				return err
			}
			initialPaperID = p.ID
		}

		// Use config listen address if --listen flag was not explicitly set
		if !cmd.Flags().Changed("listen") {
			if c := config.Load(); c.View.Listen != "" {
				viewListen = c.View.Listen
			}
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		return server.Start(ctx, initialPaperID, server.Options{
			ListenAddr: viewListen,
			NoOpen:     viewNoOpen,
			Version:    Version,
		})
	},
}

func init() {
	viewCmd.Flags().StringVar(&viewListen, "listen", "", "address to listen on (default \"127.0.0.1:0\")")
	viewCmd.Flags().BoolVar(&viewNoOpen, "no-open", false, "do not open browser automatically")
}
