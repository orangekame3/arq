package summarize

import (
	"context"
	"strings"

	"github.com/orangekame3/arq/internal/ar5iv"
	"github.com/orangekame3/arq/internal/config"
	"github.com/orangekame3/arq/internal/llm"
)

const maxContentChars = 30000

// DefaultPrompt is the default instruction prompt for full-text summarization.
// Users can override this via [summarize].prompt in config.toml.
// The placeholder {{lang}} is replaced with the configured target language at runtime.
const DefaultPrompt = `You are an academic paper summarizer. Summarize the following paper in {{lang}}.

Output ONLY a well-structured markdown document (no wrapping code fences).
Include these sections (write all headers and content in {{lang}}):

1. Overview (2-3 sentence summary)
2. Background & Motivation
3. Method
4. Key Results
5. Key Figures — For EACH figure listed below, embed it using markdown image syntax and add a 1-2 sentence explanation of what it shows. Format:

![caption](assets/filename.png)

Explanation of what this figure shows and why it matters.

6. Significance & Future Work

IMPORTANT: You MUST include a "Key Figures" section with embedded images. Use the exact filenames from the figure list below.`

// DefaultAbstractPrompt is the default instruction prompt for abstract-only summarization.
const DefaultAbstractPrompt = `You are an academic paper summarizer. Based on the title and abstract below, create a structured summary in {{lang}}.

Output ONLY a well-structured markdown document (no wrapping code fences).
Include these sections (write all headers and content in {{lang}}):

1. Overview (2-3 sentence summary)
2. Background & Motivation (inferred from abstract)
3. Approach (inferred from abstract)
4. Key Points`

// expandLang replaces {{lang}} placeholders in the prompt with the actual language.
func expandLang(prompt, lang string) string {
	return strings.ReplaceAll(prompt, "{{lang}}", lang)
}

// Summarize generates a markdown summary from ar5iv content using an LLM.
func Summarize(content *ar5iv.Content) (string, error) {
	provider, model, apiKey, lang := resolveConfig()

	instruction := config.Load().Summarize.Prompt
	if instruction == "" {
		instruction = DefaultPrompt
	}
	instruction = expandLang(instruction, lang)

	sections, figures := ar5iv.FormatForPrompt(content, maxContentChars)
	prompt := instruction + "\n\n---\nPaper content:\n" + sections + "\nFigures:\n" + figures

	// Truncate prompt if too long to avoid API payload errors
	if len(prompt) > 60000 {
		prompt = prompt[:60000] + "\n\n[truncated]"
	}

	return llm.Generate(context.Background(), provider, model, apiKey, prompt, 8192)
}

// SummarizeAbstract generates a markdown summary from just title and abstract.
// Used as fallback when ar5iv is not available.
func SummarizeAbstract(title, abstract string) (string, error) {
	provider, model, apiKey, lang := resolveConfig()

	instruction := config.Load().Summarize.Prompt
	if instruction == "" {
		instruction = DefaultAbstractPrompt
	}
	instruction = expandLang(instruction, lang)

	prompt := instruction + "\n\nTitle: " + title + "\n\nAbstract: " + abstract

	return llm.Generate(context.Background(), provider, model, apiKey, prompt, 8192)
}

func resolveConfig() (provider, model, apiKey, lang string) {
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
		provider = "openai" // fallback
	}

	model = sc.Model
	if model == "" {
		model = cfg.Translate.Model
	}
	if model == "" {
		model = llm.DefaultModel(provider)
	}

	lang = sc.Lang
	if lang == "" {
		lang = cfg.Translate.Lang
	}
	if lang == "" {
		lang = "Japanese"
	}

	return provider, model, apiKey, lang
}

// ResolveProviderForDisplay returns the effective provider name for display purposes.
func ResolveProviderForDisplay() string {
	provider, _, _, _ := resolveConfig()
	return provider
}
