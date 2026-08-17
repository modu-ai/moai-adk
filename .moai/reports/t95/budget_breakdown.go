package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func hasPaths(data []byte) bool {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t\r") != "---" {
		return false
	}
	for _, l := range lines[1:] {
		if strings.TrimRight(l, " \t\r") == "---" {
			return false
		}
		if strings.HasPrefix(l, "paths:") {
			return true
		}
	}
	return false
}

func main() {
	root := os.Args[1]
	type ent struct {
		path  string
		bytes int
	}
	var ents []ent
	filepath.WalkDir(filepath.Join(root, ".claude", "rules", "moai"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil || hasPaths(data) {
			return nil
		}
		ents = append(ents, ent{path, len(data)})
		return nil
	})
	fixed := []string{
		filepath.Join(root, "CLAUDE.md"),
		filepath.Join(root, ".claude", "output-styles", "moai", "moai.md"),
	}
	for _, p := range fixed {
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			continue
		}
		ents = append(ents, ent{p, len(data)})
	}
	sort.Slice(ents, func(i, j int) bool { return ents[i].bytes > ents[j].bytes })
	total := 0
	for _, e := range ents {
		total += e.bytes / 4
		fmt.Printf("%6d  %6d  %s\n", e.bytes/4, e.bytes, e.path)
	}
	fmt.Printf("TOTAL tokens=%d (budget 76000, delta %d)\n", total, total-76000)
}
