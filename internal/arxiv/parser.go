package arxiv

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/orangekame3/arq/internal/paper"
)

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID              string          `xml:"id"`
	Title           string          `xml:"title"`
	Summary         string          `xml:"summary"`
	Published       string          `xml:"published"`
	Authors         []atomName      `xml:"author"`
	Links           []atomLink      `xml:"link"`
	PrimaryCategory atomCategory `xml:"http://arxiv.org/schemas/atom primary_category"`
	Categories      []atomCategory `xml:"category"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

type atomName struct {
	Name string `xml:"name"`
}

type atomLink struct {
	Href  string `xml:"href,attr"`
	Title string `xml:"title,attr"`
	Type  string `xml:"type,attr"`
	Rel   string `xml:"rel,attr"`
}

// ParseAtomEntry parses an arXiv API Atom response and returns a Paper.
func ParseAtomEntry(data []byte) (*paper.Paper, error) {
	var feed atomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("parse xml: %w", err)
	}
	if len(feed.Entries) == 0 {
		return nil, fmt.Errorf("no entries found")
	}

	entry := feed.Entries[0]

	// Extract ID from URL like http://arxiv.org/abs/2303.12345v1
	rawID := entry.ID
	if idx := strings.LastIndex(rawID, "/"); idx >= 0 {
		rawID = rawID[idx+1:]
	}
	id, err := NormalizeID(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse entry id: %w", err)
	}

	// Check for error response (arXiv returns entry with "Error" in id)
	title := strings.TrimSpace(entry.Title)
	if title == "Error" {
		return nil, fmt.Errorf("paper not found: %s", id)
	}

	authors := make([]string, len(entry.Authors))
	for i, a := range entry.Authors {
		authors[i] = strings.TrimSpace(a.Name)
	}

	// Normalize whitespace in title and abstract
	title = strings.Join(strings.Fields(title), " ")
	abstract := strings.TrimSpace(entry.Summary)
	abstract = strings.Join(strings.Fields(abstract), " ")

	published := entry.Published
	if len(published) >= 10 {
		published = published[:10]
	}

	pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", id)
	for _, link := range entry.Links {
		if link.Title == "pdf" {
			pdfURL = link.Href
			break
		}
	}

	category := entry.PrimaryCategory.Term
	if category == "" {
		category = "unknown"
	}

	return &paper.Paper{
		ID:        id,
		Title:     title,
		Authors:   authors,
		Abstract:  abstract,
		Published: published,
		Category:  category,
		PDFURL:    pdfURL,
	}, nil
}
