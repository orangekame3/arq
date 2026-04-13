package translate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/llm"
)

// Result holds the translated title and abstract.
type Result struct {
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
}

const promptTemplate = `Translate the following academic paper title and abstract into %s.
Output ONLY a JSON object with "title" and "abstract" keys. No markdown, no explanation.

Title: %s

Abstract: %s`

// Translate translates a title and abstract using the configured LLM.
func Translate(title, abstract string) (*Result, error) {
	cfg := config.Load()
	provider := llm.ResolveProvider(cfg.Translate.Provider, cfg.Translate.APIKey)
	if provider == "" {
		return nil, fmt.Errorf("no API key found: run 'arq config setup' or set OPENAI_API_KEY / ANTHROPIC_API_KEY / OPENROUTER_API_KEY")
	}

	model := cfg.Translate.Model
	if model == "" {
		model = llm.DefaultModel(provider)
	}

	lang := cfg.Translate.Lang
	if lang == "" {
		lang = "Japanese"
	}

	userPrompt := fmt.Sprintf(promptTemplate, lang, title, abstract)

	text, err := llm.Generate(context.Background(), provider, model, cfg.Translate.APIKey, userPrompt, 4096)
	if err != nil {
		return nil, err
	}

	var result Result
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse translation: %w", err)
	}
	return &result, nil
}
