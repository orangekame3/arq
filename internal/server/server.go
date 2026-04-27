package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/search"
)

//go:embed static/*
var staticFS embed.FS

type paperEntry struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	TitleJA    string   `json:"title_ja,omitempty"`
	Authors    []string `json:"authors"`
	Published  string   `json:"published"`
	Category   string   `json:"category"`
	HasSummary bool     `json:"has_summary"`
	HasNote    bool     `json:"has_note"`
	HasPDFJa   bool     `json:"has_pdf_ja"`
	Keywords   []string `json:"keywords,omitempty"`
	KeywordsJA []string `json:"keywords_ja,omitempty"`
}

type paperDetail struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	TitleJA    string   `json:"title_ja,omitempty"`
	Authors    []string `json:"authors"`
	Abstract   string   `json:"abstract"`
	AbstractJA string   `json:"abstract_ja,omitempty"`
	Published  string   `json:"published"`
	Category   string   `json:"category"`
	PDFURL     string   `json:"pdf_url"`
	Keywords   []string `json:"keywords,omitempty"`
	KeywordsJA []string `json:"keywords_ja,omitempty"`
	AddedAt    string   `json:"added_at"`
	HasSummary bool     `json:"has_summary"`
	HasNote    bool     `json:"has_note"`
	HasPDFJa   bool     `json:"has_pdf_ja"`
}

type categoryEntry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type listResponse struct {
	Papers     []paperEntry    `json:"papers"`
	Categories []categoryEntry `json:"categories"`
}

// Start launches the browser-based paper viewer.
func Start(ctx context.Context, initialPaperID string) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/papers", handleListPapers)
	mux.HandleFunc("GET /api/papers/{id}", handlePaperDetail)
	mux.HandleFunc("GET /api/papers/{id}/pdf", handlePDF)
	mux.HandleFunc("GET /api/papers/{id}/pdf/ja", handlePDFJa)
	mux.HandleFunc("GET /api/papers/{id}/summary", handleSummary)
	mux.HandleFunc("GET /api/papers/{id}/note", handleNote)
	mux.HandleFunc("GET /api/papers/{id}/assets/", handleAssets)

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(staticSub))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// SPA entry point
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, _ := staticFS.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	if initialPaperID != "" {
		url += "#paper=" + initialPaperID
	}

	srv := &http.Server{Handler: mux}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	fmt.Fprintf(os.Stderr, "arq view: %s\n", url)
	openBrowser(url)

	if err := srv.Serve(ln); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func handleListPapers(w http.ResponseWriter, r *http.Request) {
	papers, err := paper.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	q := r.URL.Query().Get("q")
	cat := r.URL.Query().Get("category")

	var entries []paperEntry
	catCounts := map[string]int{}

	for _, p := range papers {
		catCounts[p.Category]++
	}

	for _, p := range papers {
		if cat != "" && p.Category != cat {
			continue
		}
		if q != "" {
			keywords := strings.Fields(strings.ToLower(q))
			matched, _ := search.Match(p, keywords, "all")
			if !matched {
				continue
			}
		}
		entries = append(entries, toPaperEntry(p))
	}

	var categories []categoryEntry
	for name, count := range catCounts {
		categories = append(categories, categoryEntry{Name: name, Count: count})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Count > categories[j].Count
	})

	writeJSON(w, listResponse{Papers: entries, Categories: categories})
}

func handlePaperDetail(w http.ResponseWriter, r *http.Request) {
	p, err := paper.FindByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, toPaperDetail(p))
}

func handlePDF(w http.ResponseWriter, r *http.Request) {
	p, err := paper.FindByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, paper.PDFPath(p))
}

func handlePDFJa(w http.ResponseWriter, r *http.Request) {
	p, err := paper.FindByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	path := paper.PDFJaPath(p)
	if _, err := os.Stat(path); err != nil {
		http.Error(w, "Japanese PDF not available", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, path)
}

func handleSummary(w http.ResponseWriter, r *http.Request) {
	p, err := paper.FindByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	serveTextFile(w, paper.SummaryPath(p))
}

func handleNote(w http.ResponseWriter, r *http.Request) {
	p, err := paper.FindByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	serveTextFile(w, paper.NotePath(p))
}

func handleAssets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := paper.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// Extract the filename after /assets/
	prefix := fmt.Sprintf("/api/papers/%s/assets/", id)
	filename := strings.TrimPrefix(r.URL.Path, prefix)
	if filename == "" || strings.Contains(filename, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(paper.AssetsDir(p), filename))
}

func toPaperEntry(p *paper.Paper) paperEntry {
	return paperEntry{
		ID:         p.ID,
		Title:      p.Title,
		TitleJA:    p.TitleJA,
		Authors:    p.Authors,
		Published:  p.Published,
		Category:   p.Category,
		HasSummary: fileExists(paper.SummaryPath(p)),
		HasNote:    fileExists(paper.NotePath(p)),
		HasPDFJa:   fileExists(paper.PDFJaPath(p)),
		Keywords:   p.Keywords,
		KeywordsJA: p.KeywordsJA,
	}
}

func toPaperDetail(p *paper.Paper) paperDetail {
	return paperDetail{
		ID:         p.ID,
		Title:      p.Title,
		TitleJA:    p.TitleJA,
		Authors:    p.Authors,
		Abstract:   p.Abstract,
		AbstractJA: p.AbstractJA,
		Published:  p.Published,
		Category:   p.Category,
		PDFURL:     p.PDFURL,
		Keywords:   p.Keywords,
		KeywordsJA: p.KeywordsJA,
		AddedAt:    p.AddedAt,
		HasSummary: fileExists(paper.SummaryPath(p)),
		HasNote:    fileExists(paper.NotePath(p)),
		HasPDFJa:   fileExists(paper.PDFJaPath(p)),
	}
}

func serveTextFile(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write(data)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	case "windows":
		cmd = "start"
	default:
		return
	}
	exec.Command(cmd, url).Start()
}
