//go:build darwin

package session

import (
	"os"
	"path/filepath"
	"strings"
)

// realpathPlatform returns dir's on-disk spelling by walking components and
// taking each directory entry's stored name: darwin APFS is case-insensitive
// and case-preserving (the t1290 F1 environment — lanes registered
// /Users/goos/moai/... while the entries spell /Users/goos/MoAI/...), so the
// entry name IS the canonical case. A component that matches only exactly,
// or any read failure, aborts with "" — canonicalCWD fails open.
func realpathPlatform(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.Clean(abs), string(filepath.Separator))
	cur := string(filepath.Separator)
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			cur = filepath.Dir(cur)
			continue
		}
		entries, err := os.ReadDir(cur)
		if err != nil {
			return ""
		}
		found := ""
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), part) {
				found = entry.Name()
				break
			}
		}
		if found == "" {
			return ""
		}
		cur = filepath.Join(cur, found)
	}
	return cur
}
