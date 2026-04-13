package ar5iv

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"golang.org/x/net/html"
)

const baseURL = "https://ar5iv.labs.arxiv.org"

// Section represents a section of the paper with its heading and body text.
type Section struct {
	Heading string
	Body    string
}

// Figure represents a figure extracted from the paper.
type Figure struct {
	URL      string // absolute URL to download from
	Filename string // local filename (e.g. "x1.png")
	Caption  string
}

// Content represents the structured content extracted from an ar5iv page.
type Content struct {
	Sections []Section
	Figures  []Figure
}

// Fetch retrieves and parses the ar5iv HTML for the given arXiv ID.
func Fetch(id string) (*Content, error) {
	url := fmt.Sprintf("%s/html/%s", baseURL, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch ar5iv: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ar5iv returned status %d (paper may not be available)", resp.StatusCode)
	}

	return parse(resp.Body)
}

func parse(r io.Reader) (*Content, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	content := &Content{}
	w := &walker{content: content}
	w.walk(doc)

	// Trim whitespace from section bodies
	for i := range content.Sections {
		content.Sections[i].Body = strings.TrimSpace(content.Sections[i].Body)
	}

	return content, nil
}

type walker struct {
	content *Content
}

func (w *walker) walk(n *html.Node) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "nav", "footer":
			return
		case "h1", "h2", "h3", "h4", "h5", "h6":
			heading := strings.TrimSpace(textContent(n))
			if heading != "" {
				w.content.Sections = append(w.content.Sections, Section{
					Heading: heading,
				})
			}
			return
		case "figure":
			fig := extractFigure(n)
			if fig != nil {
				w.content.Figures = append(w.content.Figures, *fig)
			}
			return
		}
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" && len(w.content.Sections) > 0 {
			last := &w.content.Sections[len(w.content.Sections)-1]
			last.Body += text + " "
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		w.walk(c)
	}
}

// textContent recursively extracts all text from a node.
func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textContent(c))
	}
	return sb.String()
}

// extractFigure extracts the first image URL and caption from a figure element.
func extractFigure(n *html.Node) *Figure {
	var fig Figure
	findFigureContent(n, &fig)
	if fig.URL == "" && fig.Caption == "" {
		return nil
	}
	// Make URL absolute
	if fig.URL != "" && strings.HasPrefix(fig.URL, "/") {
		fig.URL = baseURL + fig.URL
	}
	// Extract filename from URL
	if fig.URL != "" {
		fig.Filename = path.Base(fig.URL)
	}
	return &fig
}

func findFigureContent(n *html.Node, fig *Figure) {
	if n.Type == html.ElementNode {
		if n.Data == "img" && fig.URL == "" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					fig.URL = a.Val
					break
				}
			}
		} else if n.Data == "figcaption" {
			fig.Caption = strings.TrimSpace(textContent(n))
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findFigureContent(c, fig)
	}
}

// FormatForPrompt formats the content as text suitable for an LLM prompt.
// maxChars limits the total output length to control token usage.
func FormatForPrompt(c *Content, maxChars int) (sections string, figures string) {
	var sb strings.Builder
	for _, s := range c.Sections {
		if sb.Len() > maxChars {
			break
		}
		sb.WriteString("### ")
		sb.WriteString(s.Heading)
		sb.WriteString("\n")
		sb.WriteString(s.Body)
		sb.WriteString("\n\n")
	}
	sections = sb.String()

	var fb strings.Builder
	for _, f := range c.Figures {
		if f.Caption != "" {
			fb.WriteString("- assets/")
			fb.WriteString(f.Filename)
			fb.WriteString(": ")
			fb.WriteString(f.Caption)
			fb.WriteString("\n")
		}
	}
	figures = fb.String()
	if figures == "" {
		figures = "(none)"
	}

	return sections, figures
}
