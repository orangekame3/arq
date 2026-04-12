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
	Long: `Show or edit configuration.

Without subcommands, prints the current configuration.

Available keys for 'config set':
  root                  Paper storage root directory
  translate.enabled     Auto-translate on get ("true" or "false")
  translate.provider    LLM provider ("anthropic" or "openai")
  translate.model       Model name (e.g. "gpt-4o-mini", "claude-haiku-4-5-20251001")
  translate.lang        Target language (default: "Japanese")
  translate.api_key     API key (or use OPENAI_API_KEY / ANTHROPIC_API_KEY env var)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := config.Load()
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Config file: %s\n\n", config.Path())
		_, _ = fmt.Fprintf(w, "root                 = %q\n", c.Root)
		_, _ = fmt.Fprintf(w, "translate.enabled    = %v\n", c.Translate.Enabled)
		_, _ = fmt.Fprintf(w, "translate.provider   = %q\n", c.Translate.Provider)
		_, _ = fmt.Fprintf(w, "translate.model      = %q\n", c.Translate.Model)
		_, _ = fmt.Fprintf(w, "translate.lang       = %q\n", defaultStr(c.Translate.Lang, "Japanese"))
		_, _ = fmt.Fprintf(w, "translate.api_key    = %s\n", maskKey(c.Translate.APIKey))

		_, _ = fmt.Fprintf(w, "\nEffective API key: ")
		if key := effectiveAPIKey(c); key != "" {
			_, _ = fmt.Fprintf(w, "%s\n", maskKey(key))
		} else {
			_, _ = fmt.Fprintf(w, "(not set)\n")
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		c := config.Load()

		switch key {
		case "root":
			c.Root = value
		case "translate.enabled":
			if value != "true" && value != "false" {
				return fmt.Errorf("translate.enabled must be \"true\" or \"false\"")
			}
			c.Translate.Enabled = value == "true"
		case "translate.provider":
			if value != "anthropic" && value != "openai" {
				return fmt.Errorf("provider must be \"anthropic\" or \"openai\"")
			}
			c.Translate.Provider = value
		case "translate.model":
			c.Translate.Model = value
		case "translate.lang":
			c.Translate.Lang = value
		case "translate.api_key":
			c.Translate.APIKey = value
		default:
			return fmt.Errorf("unknown key: %s\n\nAvailable keys: root, translate.enabled, translate.provider, translate.model, translate.lang, translate.api_key", key)
		}

		if err := config.Save(c); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✔ %s = %q\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		c := config.Load()

		var value string
		switch key {
		case "root":
			value = c.Root
		case "translate.enabled":
			value = fmt.Sprintf("%v", c.Translate.Enabled)
		case "translate.provider":
			value = c.Translate.Provider
		case "translate.model":
			value = c.Translate.Model
		case "translate.lang":
			value = defaultStr(c.Translate.Lang, "Japanese")
		case "translate.api_key":
			value = c.Translate.APIKey
		default:
			return fmt.Errorf("unknown key: %s", key)
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), value)
		return nil
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup for translation",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := config.Load()
		reader := bufio.NewReader(os.Stdin)

		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Provider (openai/anthropic) [%s]: ", defaultStr(c.Translate.Provider, "openai"))
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			c.Translate.Provider = line
		} else if c.Translate.Provider == "" {
			c.Translate.Provider = "openai"
		}

		defaultModel := "gpt-4o-mini"
		if c.Translate.Provider == "anthropic" {
			defaultModel = "claude-haiku-4-5-20251001"
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Model [%s]: ", defaultStr(c.Translate.Model, defaultModel))
		line, _ = reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			c.Translate.Model = line
		} else if c.Translate.Model == "" {
			c.Translate.Model = defaultModel
		}

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

func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func defaultStr(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetupCmd)
}
