package paper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/orangekame3/arq/internal/config"
)

// Root returns the arq root directory.
// Priority: $ARQ_ROOT > config file > ~/papers
func Root() string {
	if v := os.Getenv("ARQ_ROOT"); v != "" {
		return v
	}
	if c := config.Load(); c.Root != "" {
		return c.Root
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "papers")
}

// Dir returns the directory for a given paper (arxiv.org/<category>/<id>).
func Dir(p *Paper) string {
	return filepath.Join(Root(), "arxiv.org", p.Category, p.ID)
}

// PDFPath returns the PDF path for a given paper.
func PDFPath(p *Paper) string {
	return filepath.Join(Dir(p), "paper.pdf")
}

// MetaPath returns the meta.json path for a given paper.
func MetaPath(p *Paper) string {
	return filepath.Join(Dir(p), "meta.json")
}

// ThumbnailPath returns the absolute path of the thumbnail if set.
func ThumbnailPath(p *Paper) string {
	if p.Thumbnail == "" {
		return ""
	}
	return filepath.Join(Dir(p), p.Thumbnail)
}

// SummaryPath returns the summary.md path for a given paper.
func SummaryPath(p *Paper) string {
	return filepath.Join(Dir(p), "summary.md")
}

// NotePath returns the note.md path for a given paper.
func NotePath(p *Paper) string {
	return filepath.Join(Dir(p), "note.md")
}

// AssetsDir returns the assets directory for a given paper.
func AssetsDir(p *Paper) string {
	return filepath.Join(Dir(p), "assets")
}

// Save writes the paper metadata and ensures the directory exists.
func Save(p *Paper) error {
	dir := Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	return os.WriteFile(MetaPath(p), data, 0o644)
}

// Load reads a paper from a meta.json at the given path.
func Load(metaPath string) (*Paper, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var p Paper
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// List returns all papers sorted by added_at (newest first).
// Walks arxiv.org/<category>/<id>/meta.json.
func List() ([]*Paper, error) {
	baseDir := filepath.Join(Root(), "arxiv.org")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, nil
	}

	var papers []*Paper
	// Walk category dirs
	categories, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	for _, cat := range categories {
		if !cat.IsDir() {
			continue
		}
		catDir := filepath.Join(baseDir, cat.Name())
		ids, err := os.ReadDir(catDir)
		if err != nil {
			continue
		}
		for _, idEntry := range ids {
			if !idEntry.IsDir() {
				continue
			}
			metaPath := filepath.Join(catDir, idEntry.Name(), "meta.json")
			p, err := Load(metaPath)
			if err != nil {
				continue
			}
			papers = append(papers, p)
		}
	}

	sort.Slice(papers, func(i, j int) bool {
		return papers[i].AddedAt > papers[j].AddedAt
	})
	return papers, nil
}

// FindByID searches for a paper by ID across all categories.
func FindByID(id string) (*Paper, error) {
	baseDir := filepath.Join(Root(), "arxiv.org")
	categories, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("paper not found: %s", id)
		}
		return nil, err
	}
	for _, cat := range categories {
		if !cat.IsDir() {
			continue
		}
		metaPath := filepath.Join(baseDir, cat.Name(), id, "meta.json")
		if p, err := Load(metaPath); err == nil {
			return p, nil
		}
	}
	return nil, fmt.Errorf("paper not found: %s", id)
}
