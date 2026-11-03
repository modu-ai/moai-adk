package runner

import (
	"os"
	"path/filepath"
	"testing"
)

var wtRoot string

// worktree paths in the tests are real directories: a bind resolves them.
func TestMain(m *testing.M) {
	d, err := os.MkdirTemp("", "af-wt-")
	if err != nil {
		panic(err)
	}
	if wtRoot, err = filepath.EvalSymlinks(d); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(d)
	os.Exit(code)
}

// wtDir returns the canonical path of a worktree directory of that name,
// creating it on first use.
func wtDir(name string) string {
	p := filepath.Join(wtRoot, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		panic(err)
	}
	return p
}
