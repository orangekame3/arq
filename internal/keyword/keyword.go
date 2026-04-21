// Package keyword extracts bilingual keywords from paper metadata using an LLM.
package keyword

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/llm"
)

const prompt = `Extract 5-10 search keywords from this paper's title and abstract.
Return ONLY a JSON object with two fields:
- "en": array of English keywords (lowercase, single words or short phrases)
- "ja": array of Japanese keywords (corresponding translations)

Example:
{"en":["quantum error correction","surface code","decoding"],"ja":["量子誤り訂正","表面符号","復号"]}

Title: %s

Abstract: %s`

type result struct {
	EN []string `json:"en"`
	JA []string `json:"ja"`
}

// Extract returns English and Japanese keywords for the given title and abstract.
func Extract(title, abstract string) (en, ja []string, err error) {
	provider, model, apiKey := resolveConfig()

	p := fmt.Sprintf(prompt, title, abstract)
	raw, err := llm.Generate(context.Background(), provider, model, apiKey, p, 1024)
	if err != nil {
		return nil, nil, fmt.Errorf("keyword extraction: %w", err)
	}

	// Strip markdown code fences if present
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var r result
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, nil, fmt.Errorf("parse keywords: %w (raw: %s)", err, raw)
	}

	return r.EN, r.JA, nil
}

func resolveConfig() (provider, model, apiKey string) {
	cfg := config.Load()
	sc := cfg.Summarize

	apiKey = sc.APIKey
	if apiKey == "" {
		apiKey = cfg.Translate.APIKey
	}

	provider = sc.Provider
	if provider == "" {
		provider = cfg.Translate.Provider
	}
	provider = llm.ResolveProvider(provider, apiKey)
	if provider == "" {
		provider = "openai"
	}

	model = sc.Model
	if model == "" {
		model = cfg.Translate.Model
	}
	if model == "" {
		model = llm.DefaultModel(provider)
	}

	return provider, model, apiKey
}
