package arxiv

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/orangekame3/arq/internal/paper"
)

var idPattern = regexp.MustCompile(`^(\d{4}\.\d{4,5})(v\d+)?$`)

// NormalizeID extracts an arXiv ID from a raw string (ID or URL).
func NormalizeID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	// Strip URL prefix
	for _, prefix := range []string{
		"https://arxiv.org/abs/",
		"http://arxiv.org/abs/",
		"https://arxiv.org/pdf/",
		"http://arxiv.org/pdf/",
	} {
		if strings.HasPrefix(raw, prefix) {
			raw = strings.TrimPrefix(raw, prefix)
			raw = strings.TrimSuffix(raw, ".pdf")
			break
		}
	}

	// Remove version suffix
	if m := idPattern.FindStringSubmatch(raw); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("invalid arXiv ID: %s", raw)
}

// Fetch retrieves metadata and PDF for the given arXiv ID.
func Fetch(id string) (*paper.Paper, error) {
	apiURL := fmt.Sprintf("https://export.arxiv.org/api/query?id_list=%s", id)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	p, err := ParseAtomEntry(body)
	if err != nil {
		return nil, err
	}

	p.AddedAt = time.Now().UTC().Format(time.RFC3339)

	return p, nil
}

// DownloadPDF downloads the PDF for the given paper.
func DownloadPDF(p *paper.Paper) error {
	resp, err := http.Get(p.PDFURL)
	if err != nil {
		return fmt.Errorf("download pdf: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download pdf: status %d", resp.StatusCode)
	}

	pdfPath := paper.PDFPath(p)
	f, err := os.Create(pdfPath)
	if err != nil {
		return fmt.Errorf("create pdf file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write pdf: %w", err)
	}
	return nil
}
