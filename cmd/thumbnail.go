package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/orangekame3/arq/internal/paper"
	"github.com/orangekame3/arq/internal/query"
	"github.com/spf13/cobra"
)

var thumbnailCmd = &cobra.Command{
	Use:   "thumbnail",
	Short: "Manage paper thumbnails",
}

var thumbnailSetCmd = &cobra.Command{
	Use:   "set <query> <image-path>",
	Short: "Set a thumbnail image for a paper",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}

		srcPath := args[1]
		ext := filepath.Ext(srcPath)
		if ext == "" {
			ext = ".png"
		}
		filename := "thumbnail" + ext

		src, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("open image: %w", err)
		}
		defer func() { _ = src.Close() }()

		dstPath := filepath.Join(paper.Dir(p), filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("create thumbnail: %w", err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			_ = dst.Close()
			return fmt.Errorf("copy image: %w", err)
		}
		if err := dst.Close(); err != nil {
			return fmt.Errorf("close thumbnail: %w", err)
		}

		p.Thumbnail = filename
		if err := paper.Save(p); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✔ thumbnail set for %s\n", p.ID)
		return nil
	},
}

var thumbnailPathCmd = &cobra.Command{
	Use:   "path <query>",
	Short: "Print the thumbnail path for a paper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := query.Resolve(args[0])
		if err != nil {
			return err
		}
		path := paper.ThumbnailPath(p)
		if path == "" {
			return fmt.Errorf("no thumbnail for %s", p.ID)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
		return nil
	},
}

func init() {
	thumbnailCmd.AddCommand(thumbnailSetCmd)
	thumbnailCmd.AddCommand(thumbnailPathCmd)
}
