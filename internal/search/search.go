package search

import (
	"os"
	"strings"

	"github.com/orangekame3/arq/internal/paper"
)

// Match checks whether a paper matches all keywords (AND logic, case-insensitive).
// field restricts to a specific field ("title", "abstract", "keywords", "summary", "all").
// Returns (matched bool, matchedFields []string).
func Match(p *paper.Paper, keywords []string, field string) (bool, []string) {
	fieldTexts := map[string]string{
		"title":    strings.ToLower(p.Title + " " + p.TitleJA),
		"abstract": strings.ToLower(p.Abstract + " " + p.AbstractJA),
		"keywords": strings.ToLower(strings.Join(p.Keywords, " ") + " " + strings.Join(p.KeywordsJA, " ")),
	}

	if field == "all" || field == "summary" {
		if summaryPath := paper.SummaryPath(p); summaryPath != "" {
			if data, err := os.ReadFile(summaryPath); err == nil {
				fieldTexts["summary"] = strings.ToLower(string(data))
			}
		}
	}

	for _, kw := range keywords {
		found := false
		for f, text := range fieldTexts {
			if field != "all" && field != f {
				continue
			}
			if strings.Contains(text, kw) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	var matchedFields []string
	for f, text := range fieldTexts {
		if field != "all" && field != f {
			continue
		}
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				matchedFields = append(matchedFields, f)
				break
			}
		}
	}

	return true, matchedFields
}
