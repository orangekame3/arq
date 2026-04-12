package query

import (
	"fmt"
	"strings"

	"github.com/orangekame3/arq/internal/paper"
)

// Resolve finds a paper matching the given query.
// It checks: exact ID match, partial ID match, title substring match.
// Returns error if 0 or >1 matches.
func Resolve(q string) (*paper.Paper, error) {
	papers, err := paper.List()
	if err != nil {
		return nil, fmt.Errorf("list papers: %w", err)
	}

	q = strings.TrimSpace(q)
	qLower := strings.ToLower(q)

	// Exact ID match
	for _, p := range papers {
		if p.ID == q {
			return p, nil
		}
	}

	// Partial ID or title substring match
	var matches []*paper.Paper
	for _, p := range papers {
		if strings.Contains(p.ID, q) || strings.Contains(strings.ToLower(p.Title), qLower) {
			matches = append(matches, p)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no paper found for query: %s", q)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("multiple papers match query %q (%d results), use 'arq select' to choose", q, len(matches))
	}
}
