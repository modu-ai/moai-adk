package backup

// retention_probe_helpers_test.go — fixtures shared by the pruner-level probe
// tests of SPEC-UPDATE-BACKUP-RETENTION-001 (AC-UBR-01, AC-UBR-02, AC-UBR-03).
// The probes drive CleanupOldBackups through the real directory layout, so each
// criterion is measured against the pruner rather than a model of it.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// t1733WriteFile writes body at p, creating its parents.
func t1733WriteFile(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// t1733DirNames returns the sorted names of the directories directly under root.
// A missing root yields none.
func t1733DirNames(root string) []string {
	ents, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// t1733NamespaceRun creates one namespace backup run (update-2026-10-NNT00-00-00Z)
// holding the file .claude/agents/x.md with content version, and stamps the run
// directory with the time its name encodes so runs order by day n. It returns the
// run's directory name.
func t1733NamespaceRun(t *testing.T, projectRoot string, n int, version string) string {
	t.Helper()
	name := fmt.Sprintf("update-2026-10-%02dT00-00-00Z", n)
	dir := filepath.Join(projectRoot, ".moai", "backups", name)
	t1733WriteFile(t, filepath.Join(dir, ".claude", "agents", "x.md"), version+"\n")
	ts := time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(dir, ts, ts); err != nil {
		t.Fatal(err)
	}
	return name
}

// t1733Contains reports whether names holds name.
func t1733Contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
