package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/orangekame3/arq/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or edit configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := config.Load()
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Config file: %s\n\n", config.Path())
		_, _ = fmt.Fprintf(w, "root       = %q\n", c.Root)
		_, _ = fmt.Fprintf(w, "[translate]\n")
		_, _ = fmt.Fprintf(w, "provider   = %q\n", c.Translate.Provider)
		_, _ = fmt.Fprintf(w, "model      = %q\n", c.Translate.Model)
		_, _ = fmt.Fprintf(w, "api_key    = %q\n", c.Translate.APIKey)

		// Show effective state
		_, _ = fmt.Fprintf(w, "\nEffective API key: ")
		if key := effectiveAPIKey(c); key != "" {
			_, _ = fmt.Fprintf(w, "%s...%s\n", key[:4], key[len(key)-4:])
		} else {
			_, _ = fmt.Fprintf(w, "(not set)\n")
		}
		return nil
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup for translation",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := config.Load()
		reader := bufio.NewReader(os.Stdin)

		// Provider
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Provider (anthropic/openai) [%s]: ", defaultStr(c.Translate.Provider, "anthropic"))
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			c.Translate.Provider = line
		} else if c.Translate.Provider == "" {
			c.Translate.Provider = "anthropic"
		}

		// Model
		defaultModel := "claude-haiku-4-5-20251001"
		if c.Translate.Provider == "openai" {
			defaultModel = "gpt-4o-mini"
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Model [%s]: ", defaultStr(c.Translate.Model, defaultModel))
		line, _ = reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			c.Translate.Model = line
		} else if c.Translate.Model == "" {
			c.Translate.Model = defaultModel
		}

		// API Key
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "API Key (leave empty to use env var): ")
		line, _ = reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			c.Translate.APIKey = line
		}

		if err := config.Save(c); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✔ saved to %s\n", config.Path())
		return nil
	},
}

func effectiveAPIKey(c config.Config) string {
	if c.Translate.APIKey != "" {
		return c.Translate.APIKey
	}
	switch c.Translate.Provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	default:
		return os.Getenv("ANTHROPIC_API_KEY")
	}
}

func defaultStr(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}

func init() {
	configCmd.AddCommand(configSetupCmd)
}
