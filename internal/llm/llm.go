// Package llm provides a unified LLM client using charmbracelet/fantasy.
// It supports OpenAI, Anthropic, and OpenRouter providers.
package llm

import (
	"context"
	"fmt"
	"os"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openrouter"
)

// Generate sends a prompt to the configured LLM and returns the text response.
func Generate(ctx context.Context, provider, model, apiKey, prompt string, maxTokens int64) (string, error) {
	lm, err := resolveModel(ctx, provider, model, apiKey)
	if err != nil {
		return "", err
	}

	resp, err := lm.Generate(ctx, fantasy.Call{
		Prompt:          fantasy.Prompt{fantasy.NewUserMessage(prompt)},
		MaxOutputTokens: &maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("llm generate: %w", err)
	}

	text := resp.Content.Text()
	if text == "" {
		return "", fmt.Errorf("llm: empty response")
	}
	return text, nil
}

// GenerateWithSystem sends a system + user prompt and returns the text response.
func GenerateWithSystem(ctx context.Context, provider, model, apiKey, system, userPrompt string, maxTokens int64) (string, error) {
	lm, err := resolveModel(ctx, provider, model, apiKey)
	if err != nil {
		return "", err
	}

	resp, err := lm.Generate(ctx, fantasy.Call{
		Prompt: fantasy.Prompt{
			fantasy.NewSystemMessage(system),
			fantasy.NewUserMessage(userPrompt),
		},
		MaxOutputTokens: &maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("llm generate: %w", err)
	}

	text := resp.Content.Text()
	if text == "" {
		return "", fmt.Errorf("llm: empty response")
	}
	return text, nil
}

func resolveModel(ctx context.Context, providerName, modelID, apiKey string) (fantasy.LanguageModel, error) {
	p, err := newProvider(providerName, apiKey)
	if err != nil {
		return nil, err
	}
	lm, err := p.LanguageModel(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("llm model %q: %w", modelID, err)
	}
	return lm, nil
}

func newProvider(name, apiKey string) (fantasy.Provider, error) {
	switch name {
	case "openai":
		key := resolveAPIKey(apiKey, "OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("no API key for openai: set OPENAI_API_KEY or configure api_key")
		}
		return openai.New(openai.WithAPIKey(key))

	case "anthropic":
		key := resolveAPIKey(apiKey, "ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("no API key for anthropic: set ANTHROPIC_API_KEY or configure api_key")
		}
		return anthropic.New(anthropic.WithAPIKey(key))

	case "openrouter":
		key := resolveAPIKey(apiKey, "OPENROUTER_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("no API key for openrouter: set OPENROUTER_API_KEY or configure api_key")
		}
		return openrouter.New(openrouter.WithAPIKey(key))

	default:
		return nil, fmt.Errorf("unknown provider %q: use \"openai\", \"anthropic\", or \"openrouter\"", name)
	}
}

func resolveAPIKey(configKey, envVar string) string {
	if configKey != "" {
		return configKey
	}
	return os.Getenv(envVar)
}

// ResolveProvider auto-detects the provider from available API keys if not configured.
func ResolveProvider(provider, apiKey string) string {
	if provider != "" {
		return provider
	}
	if apiKey != "" || os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "anthropic"
	}
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		return "openrouter"
	}
	return ""
}

// Model represents a selectable LLM model.
type Model struct {
	ID   string // model ID to pass to the API
	Name string // human-readable display name
}

// Models returns the list of supported models for a given provider.
func Models(provider string) []Model {
	switch provider {
	case "openai":
		return []Model{
			{ID: "gpt-5.4-nano", Name: "GPT-5.4 Nano (fastest, cheapest)"},
			{ID: "gpt-5.4-mini", Name: "GPT-5.4 Mini (balanced)"},
			{ID: "gpt-5.4", Name: "GPT-5.4 (latest flagship)"},
			{ID: "gpt-5.2-pro", Name: "GPT-5.2 Pro"},
			{ID: "gpt-5.2", Name: "GPT-5.2"},
			{ID: "gpt-5.1", Name: "GPT-5.1"},
			{ID: "gpt-5.1-mini", Name: "GPT-5.1 Mini"},
			{ID: "o4-mini", Name: "o4-mini (reasoning)"},
			{ID: "o3", Name: "o3 (advanced reasoning)"},
			{ID: "gpt-4.1-nano", Name: "GPT-4.1 Nano"},
			{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini"},
			{ID: "gpt-4.1", Name: "GPT-4.1"},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini (legacy)"},
			{ID: "gpt-4o", Name: "GPT-4o (legacy)"},
		}
	case "anthropic":
		return []Model{
			{ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5 (fastest, cheapest)"},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (fast, latest)"},
			{ID: "claude-opus-4-6", Name: "Claude Opus 4.6 (most capable)"},
			{ID: "claude-sonnet-4-5-20250929", Name: "Claude Sonnet 4.5 (legacy)"},
			{ID: "claude-opus-4-5-20251101", Name: "Claude Opus 4.5 (legacy)"},
		}
	case "openrouter":
		return []Model{
			{ID: "openai/gpt-5.4-mini", Name: "GPT-5.4 Mini"},
			{ID: "openai/gpt-5.4", Name: "GPT-5.4"},
			{ID: "openai/gpt-4.1-mini", Name: "GPT-4.1 Mini"},
			{ID: "openai/gpt-4.1", Name: "GPT-4.1"},
			{ID: "openai/o4-mini", Name: "o4-mini (reasoning)"},
			{ID: "anthropic/claude-sonnet-4-6", Name: "Claude Sonnet 4.6"},
			{ID: "anthropic/claude-opus-4-6", Name: "Claude Opus 4.6"},
			{ID: "anthropic/claude-haiku-4-5", Name: "Claude Haiku 4.5"},
			{ID: "google/gemini-2.5-flash", Name: "Gemini 2.5 Flash"},
			{ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro"},
		}
	default:
		return nil
	}
}

// DefaultModel returns the default model for a given provider.
func DefaultModel(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-haiku-4-5-20251001"
	case "openrouter":
		return "openai/gpt-5.4-mini"
	default:
		return "gpt-5.4-mini"
	}
}
