package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/llm"
	"github.com/orangekame3/arq/internal/summarize"
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
  translate.api_key     API key (or use OPENAI_API_KEY / ANTHROPIC_API_KEY env var)
  summarize.provider    LLM provider for summarization (falls back to translate.provider)
  summarize.model       Model name for summarization (falls back to translate.model)
  summarize.lang        Target language for summarization (falls back to translate.lang)
  summarize.api_key     API key for summarization (falls back to translate.api_key)
  summarize.prompt      Custom instruction prompt (use {{lang}} as language placeholder)`,
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
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "summarize.enabled    = %v\n", c.Summarize.Enabled)
		_, _ = fmt.Fprintf(w, "summarize.provider   = %s\n", fallbackStr(c.Summarize.Provider, c.Translate.Provider, "(inherit from translate)"))
		_, _ = fmt.Fprintf(w, "summarize.model      = %s\n", fallbackStr(c.Summarize.Model, c.Translate.Model, "(inherit from translate)"))
		_, _ = fmt.Fprintf(w, "summarize.lang       = %s\n", fallbackStr(c.Summarize.Lang, c.Translate.Lang, "Japanese"))
		_, _ = fmt.Fprintf(w, "summarize.api_key    = %s\n", maskKey(c.Summarize.APIKey))
		if c.Summarize.Prompt != "" {
			// Show first line + truncation indicator
			lines := strings.SplitN(c.Summarize.Prompt, "\n", 2)
			if len(lines) > 1 {
				_, _ = fmt.Fprintf(w, "summarize.prompt     = %q... (%d chars)\n", lines[0], len(c.Summarize.Prompt))
			} else {
				_, _ = fmt.Fprintf(w, "summarize.prompt     = %q\n", c.Summarize.Prompt)
			}
		} else {
			_, _ = fmt.Fprintf(w, "summarize.prompt     = (default)\n")
		}

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
			if !config.IsValidProvider(value) {
				return fmt.Errorf("provider must be one of: %v", config.ValidProviders)
			}
			c.Translate.Provider = value
		case "translate.model":
			c.Translate.Model = value
		case "translate.lang":
			c.Translate.Lang = value
		case "translate.api_key":
			c.Translate.APIKey = value
		case "summarize.enabled":
			if value != "true" && value != "false" {
				return fmt.Errorf("summarize.enabled must be \"true\" or \"false\"")
			}
			c.Summarize.Enabled = value == "true"
		case "summarize.provider":
			if !config.IsValidProvider(value) {
				return fmt.Errorf("provider must be one of: %v", config.ValidProviders)
			}
			c.Summarize.Provider = value
		case "summarize.model":
			c.Summarize.Model = value
		case "summarize.lang":
			c.Summarize.Lang = value
		case "summarize.api_key":
			c.Summarize.APIKey = value
		case "summarize.prompt":
			c.Summarize.Prompt = value
		default:
			return fmt.Errorf("unknown key: %s\n\nRun 'arq config' to see available keys", key)
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
		case "summarize.enabled":
			value = fmt.Sprintf("%v", c.Summarize.Enabled)
		case "summarize.provider":
			value = c.Summarize.Provider
		case "summarize.model":
			value = c.Summarize.Model
		case "summarize.lang":
			value = defaultStr(c.Summarize.Lang, defaultStr(c.Translate.Lang, "Japanese"))
		case "summarize.api_key":
			value = c.Summarize.APIKey
		case "summarize.prompt":
			value = c.Summarize.Prompt
			if value == "" {
				value = summarize.DefaultPrompt
			}
		default:
			return fmt.Errorf("unknown key: %s\n\nRun 'arq config' to see available keys", key)
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), value)
		return nil
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := config.Load()

		// Section selection
		configureTranslate := true
		configureSummarize := true

		// Pre-fill translate values
		translateProvider := defaultStr(c.Translate.Provider, "openai")
		translateModel := c.Translate.Model
		translateLang := defaultStr(c.Translate.Lang, "Japanese")
		translateAPIKey := c.Translate.APIKey
		translateEnabled := c.Translate.Enabled

		// Pre-fill summarize values
		summarizeEnabled := c.Summarize.Enabled
		summarizeProvider := c.Summarize.Provider
		summarizeModel := c.Summarize.Model
		summarizePrompt := defaultStr(c.Summarize.Prompt, summarize.DefaultPrompt)

		// Pre-fill general
		root := c.Root

		// Build model options dynamically based on selected provider
		modelOptionsFor := func(provider string) []huh.Option[string] {
			models := llm.Models(provider)
			opts := make([]huh.Option[string], 0, len(models))
			for _, m := range models {
				opts = append(opts, huh.NewOption(m.Name, m.ID))
			}
			return opts
		}

		// Set default model if not configured
		if translateModel == "" {
			translateModel = llm.DefaultModel(translateProvider)
		}

		form := huh.NewForm(
			// General + section selection
			huh.NewGroup(
				huh.NewInput().
					Title("Paper storage root").
					Description("Leave empty for ~/papers").
					Placeholder("~/papers").
					Value(&root),
				huh.NewConfirm().
					Title("Configure translate settings?").
					Value(&configureTranslate).
					Affirmative("Yes").
					Negative("Skip"),
				huh.NewConfirm().
					Title("Configure summarize settings?").
					Value(&configureSummarize).
					Affirmative("Yes").
					Negative("Skip"),
			).Title("General"),

			// Translate
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Provider").
					Options(
						huh.NewOption("OpenAI", "openai"),
						huh.NewOption("Anthropic", "anthropic"),
						huh.NewOption("OpenRouter", "openrouter"),
					).
					Value(&translateProvider),
				huh.NewSelect[string]().
					Title("Model").
					OptionsFunc(func() []huh.Option[string] {
						return modelOptionsFor(translateProvider)
					}, &translateProvider).
					Value(&translateModel),
				huh.NewInput().
					Title("Language").
					Placeholder("Japanese").
					Value(&translateLang),
				huh.NewInput().
					Title("API Key").
					Description("Leave empty to use env var (OPENAI_API_KEY / ANTHROPIC_API_KEY / OPENROUTER_API_KEY)").
					EchoMode(huh.EchoModePassword).
					Value(&translateAPIKey),
				huh.NewConfirm().
					Title("Auto-translate on get?").
					Value(&translateEnabled),
			).Title("Translate").
				WithHideFunc(func() bool { return !configureTranslate }),

			// Summarize
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Provider").
					Options(
						huh.NewOption("Inherit from translate", ""),
						huh.NewOption("OpenAI", "openai"),
						huh.NewOption("Anthropic", "anthropic"),
						huh.NewOption("OpenRouter", "openrouter"),
					).
					Value(&summarizeProvider),
				huh.NewSelect[string]().
					Title("Model").
					OptionsFunc(func() []huh.Option[string] {
						p := summarizeProvider
						if p == "" {
							p = translateProvider
						}
						opts := []huh.Option[string]{huh.NewOption("Inherit from translate", "")}
						opts = append(opts, modelOptionsFor(p)...)
						return opts
					}, []*string{&summarizeProvider, &translateProvider}).
					Value(&summarizeModel),
				huh.NewText().
					Title("Prompt").
					Description("Use {{lang}} as language placeholder").
					Value(&summarizePrompt).
					Lines(8).
					CharLimit(5000),
				huh.NewConfirm().
					Title("Auto-summarize on get?").
					Value(&summarizeEnabled),
			).Title("Summarize").
				WithHideFunc(func() bool { return !configureSummarize }),
		)

		if err := form.Run(); err != nil {
			return err
		}

		// Apply values
		c.Root = root

		if configureTranslate {
			c.Translate.Provider = translateProvider
			c.Translate.Model = translateModel
			c.Translate.Lang = translateLang
			c.Translate.APIKey = translateAPIKey
			c.Translate.Enabled = translateEnabled
		}

		if configureSummarize {
			c.Summarize.Enabled = summarizeEnabled
			c.Summarize.Provider = summarizeProvider
			c.Summarize.Model = summarizeModel
			c.Summarize.Prompt = summarizePrompt
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

func fallbackStr(val, inherit, fallback string) string {
	if val != "" {
		return fmt.Sprintf("%q", val)
	}
	if inherit != "" {
		return fmt.Sprintf("%q (from translate)", inherit)
	}
	return fallback
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetupCmd)
}
