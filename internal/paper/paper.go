package paper

import (
	"time"
)

// Paper represents an arXiv paper with its metadata.
type Paper struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	TitleJA    string   `json:"title_ja,omitempty"`
	Authors    []string `json:"authors"`
	Abstract   string   `json:"abstract"`
	AbstractJA string   `json:"abstract_ja,omitempty"`
	Published  string   `json:"published"`
	Category   string   `json:"category"`
	PDFURL     string   `json:"pdf_url"`
	Thumbnail  string   `json:"thumbnail,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
	KeywordsJA []string `json:"keywords_ja,omitempty"`
	AddedAt    string   `json:"added_at"`
}

// AuthorShort returns a short author string like "Smith et al." or "Smith, Tanaka".
func (p *Paper) AuthorShort() string {
	switch len(p.Authors) {
	case 0:
		return ""
	case 1:
		return p.Authors[0]
	case 2:
		return p.Authors[0] + ", " + p.Authors[1]
	default:
		return p.Authors[0] + " et al."
	}
}

// PublishedShort returns YYYY-MM format.
func (p *Paper) PublishedShort() string {
	t, err := time.Parse("2006-01-02", p.Published)
	if err != nil {
		return p.Published
	}
	return t.Format("2006-01")
}
