package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/orangekame3/arq/internal/config"
)

// Result holds the translated title and abstract.
type Result struct {
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
}

const prompt = `Translate the following academic paper title and abstract into Japanese.
Output ONLY a JSON object with "title" and "abstract" keys. No markdown, no explanation.

Title: %s

Abstract: %s`

// Translate translates a title and abstract to Japanese using the configured LLM.
// Provider detection: config file > environment variable auto-detection.
func Translate(title, abstract string) (*Result, error) {
	cfg := config.Load()
	provider := cfg.Translate.Provider
	model := cfg.Translate.Model
	apiKey := cfg.Translate.APIKey

	// Auto-detect provider from env vars if not configured
	if provider == "" {
		if apiKey != "" || os.Getenv("ANTHROPIC_API_KEY") != "" {
			provider = "anthropic"
		} else if os.Getenv("OPENAI_API_KEY") != "" {
			provider = "openai"
		} else {
			return nil, fmt.Errorf("no API key found: run 'arq config setup' or set ANTHROPIC_API_KEY / OPENAI_API_KEY")
		}
	}

	// Resolve API key: config > env var
	if apiKey == "" {
		switch provider {
		case "anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no API key for %s: run 'arq config setup' or set env var", provider)
	}

	userPrompt := fmt.Sprintf(prompt, title, abstract)

	switch provider {
	case "anthropic":
		return callAnthropic(userPrompt, model, apiKey)
	case "openai":
		return callOpenAI(userPrompt, model, apiKey)
	default:
		return nil, fmt.Errorf("unknown provider: %s (use \"anthropic\" or \"openai\")", provider)
	}
}

func callAnthropic(userPrompt, model, apiKey string) (*Result, error) {
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	reqBody, _ := json.Marshal(struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Messages  []msg  `json:"messages"`
	}{
		Model:     model,
		MaxTokens: 4096,
		Messages:  []msg{{Role: "user", Content: userPrompt}},
	})

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	return doRequest(req, func(body []byte) (string, error) {
		var resp struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return "", err
		}
		if len(resp.Content) == 0 {
			return "", fmt.Errorf("empty response")
		}
		return resp.Content[0].Text, nil
	})
}

func callOpenAI(userPrompt, model, apiKey string) (*Result, error) {
	if model == "" {
		model = "gpt-4o-mini"
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	reqBody, _ := json.Marshal(struct {
		Model    string `json:"model"`
		Messages []msg  `json:"messages"`
	}{
		Model:    model,
		Messages: []msg{{Role: "user", Content: userPrompt}},
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	return doRequest(req, func(body []byte) (string, error) {
		var resp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("empty response")
		}
		return resp.Choices[0].Message.Content, nil
	})
}

func doRequest(req *http.Request, extractText func([]byte) (string, error)) (*Result, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	text, err := extractText(body)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var result Result
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse translation: %w", err)
	}
	return &result, nil
}
