package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var upgradeDryRun bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade arq to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(cmd.ErrOrStderr())

		latest, downloadURL, err := fetchLatestRelease()
		if err != nil {
			return fmt.Errorf("fetch latest release: %w", err)
		}

		if latest == Version {
			logger.Info("already up to date", "version", Version)
			return nil
		}

		logger.Info("new version available", "current", Version, "latest", latest)

		if upgradeDryRun {
			return nil
		}

		if isHomebrew() {
			logger.Warn("installed via Homebrew — run: brew upgrade orangekame3/tap/arq")
			return nil
		}

		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("detect executable path: %w", err)
		}

		logger.Info("downloading", "url", downloadURL)

		bin, err := downloadBinary(downloadURL)
		if err != nil {
			return fmt.Errorf("download: %w", err)
		}

		if err := replaceBinary(execPath, bin); err != nil {
			return fmt.Errorf("replace binary: %w", err)
		}

		logger.Info("upgraded", "version", latest, "path", execPath)
		return nil
	},
}

type ghRelease struct {
	TagName string `json:"tag_name"`
}

func fetchLatestRelease() (version string, downloadURL string, err error) {
	resp, err := http.Get("https://api.github.com/repos/orangekame3/arq/releases/latest")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", err
	}

	version = strings.TrimPrefix(rel.TagName, "v")

	osName := archiveOS()
	archName := archiveArch()
	archive := fmt.Sprintf("arq_%s_%s.tar.gz", osName, archName)
	downloadURL = fmt.Sprintf("https://github.com/orangekame3/arq/releases/download/%s/%s", rel.TagName, archive)

	return version, downloadURL, nil
}

func archiveOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin"
	case "linux":
		return "Linux"
	default:
		return strings.Title(runtime.GOOS) //nolint:staticcheck
	}
}

func archiveArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "386":
		return "i386"
	default:
		return runtime.GOARCH
	}
}

func isHomebrew() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(execPath, "Cellar") || strings.Contains(execPath, "homebrew")
}

func downloadBinary(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar: %w", err)
		}
		if hdr.Name == "arq" {
			return io.ReadAll(tr)
		}
	}

	return nil, fmt.Errorf("binary not found in archive")
}

func replaceBinary(path string, newBin []byte) error {
	// Write to a temp file next to the target, then atomic rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, newBin, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", false, "Check for updates without installing")
}
