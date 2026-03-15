package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	watermark "github.com/esteban-plaza/img_settings/lib"
)

// uriToPath converts a fyne.URI (file://...) to an OS filesystem path.
func uriToPath(u fyne.URI) string {
	s := u.String()
	s = strings.TrimPrefix(s, "file://")
	decoded, err := url.PathUnescape(s)
	if err != nil {
		decoded = s
	}
	return filepath.FromSlash(decoded)
}

// isDir returns true if path points to a directory.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// collectFiles gathers supported image files from the given paths.
// When recurse is false, sub-directories inside dropped folders are skipped.
func collectFiles(paths []string, recurse bool) []string {
	var files []string
	for _, p := range paths {
		if isDir(p) {
			_ = filepath.Walk(p, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fi.IsDir() {
					if !recurse && path != p {
						return filepath.SkipDir
					}
					return nil
				}
				if watermark.IsSupportedExt(path) {
					files = append(files, path)
				}
				return nil
			})
		} else {
			if watermark.IsSupportedExt(p) {
				files = append(files, p)
			}
		}
	}
	return files
}

// humanSize formats a byte count as a human-readable string.
func humanSize(path string) string {
	fi, err := os.Stat(path)
	if err != nil {
		return ""
	}
	mb := float64(fi.Size()) / 1e6
	return fmt.Sprintf("%.1f MB", mb)
}
