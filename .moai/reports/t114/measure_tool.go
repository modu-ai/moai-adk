//go:build ignore

// Command t114_measure replicates internal/config alwaysLoadedSurface +
// measureAlwaysLoaded per-file attribution for card t114 evidence.
// Build-tagged ignore so it never enters the module's package graph.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func frontmatterHasPaths(data []byte) bool {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t\r") != "---" {
		return false
	}
	for _, line := range lines[1:] {
		if strings.TrimRight(line, " \t\r") == "---" {
			return false
		}
		if strings.HasPrefix(line, "paths:") {
			return true
		}
	}
	return false
}

func main() {
	root := os.Args[1]
	var files []string
	filepath.WalkDir(filepath.Join(root, ".claude", "rules", "moai"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, _ := os.ReadFile(path)
		if !frontmatterHasPaths(data) {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	files = append(files, filepath.Join(root, "CLAUDE.md"), filepath.Join(root, ".claude", "output-styles", "moai", "moai.md"), filepath.Join(root, "MEMORY.md"))
	total := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		n := 0
		b := 0
		if err == nil {
			b = len(data)
			n = len(data) / 4
		}
		total += n
		rel, _ := filepath.Rel(root, f)
		fmt.Printf("%7d tokens %9d bytes  %s\n", n, b, rel)
	}
	fmt.Printf("TOTAL: %d tokens, %d entries\n", total, len(files))
}
