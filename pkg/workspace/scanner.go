// Package workspace provides workspace scanning for Tekton YAML files.
package workspace

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vdemeester/tekton-lsp-go/pkg/cache"
)

// Scan walks a workspace root directory and indexes all YAML files into the cache.
// The rootURI should be a file:// URI. Returns the number of files indexed.
func Scan(rootURI string, c *cache.Cache) (int, error) {
	root := strings.TrimPrefix(rootURI, "file://")

	count := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read.
		}
		if d.IsDir() {
			// Skip hidden directories.
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read.
		}

		uri := "file://" + path
		c.Insert(uri, "yaml", 0, string(content))
		count++
		return nil
	})

	return count, err
}
