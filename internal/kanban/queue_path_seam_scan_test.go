package kanban

// queue_path_seam_scan_test.go — SPEC-TODO-QUEUE-HOME-CANON-001 AC-003b:
// the repo-scope seam guard. The legacy `backlog.json` queue path is built in
// exactly one production place — this package's `backlogFileName` seam const —
// plus one deliberate boundary exception: the one-off t657-merge utility
// (card t835's merge-core boundary; the merge utility's census logic and its
// retirement are that card's, not this SPEC's). This test makes "exactly one
// seam" mechanically checkable rather than a claim in a comment, widening the
// existing internal/web read-seam assertion (todo_queue_read_test.go) to the
// whole repository.
//
// The scan matches the exact quoted literal `"backlog.json"` so the
// lock-file constant (`"backlog.json.lock"`, backlog_store.go) and unquoted
// prose mentions never trip it — the guard watches queue-PATH constructions,
// not naming (dispositions ledger row B5).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// queuePathLiteral is the exact quoted literal a queue-path construction
// carries. The lock-file name "backlog.json.lock" does NOT contain this
// sequence (its closing quote sits later), which is the discrimination the
// acceptance edge case demands.
const queuePathLiteral = "\"backlog.json\""

// queuePathSeamExemptions are the only files allowed to carry the literal,
// with the reason each exemption stands.
var queuePathSeamExemptions = map[string]string{
	filepath.Join("internal", "kanban", "state_dir.go"): "the seam const every queue path enters through (backlogFileName)",
	filepath.Join("cmd", "t657-merge", "main.go"):       "one-off merge utility boundary exception, owned by card t835",
}

func TestBacklogJSONLiteralStaysSeamScoped(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	scanned := 0
	hits := map[string]int{}
	for _, dir := range []string{"internal", "pkg", "cmd"} {
		base := filepath.Join(repoRoot, dir)
		if _, err := os.Stat(base); err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
		err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if name == "vendor" || name == "testdata" || name == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			name := d.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			scanned++
			if strings.Contains(string(src), queuePathLiteral) {
				rel, err := filepath.Rel(repoRoot, path)
				if err != nil {
					return err
				}
				hits[filepath.Clean(rel)]++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	// An empty sweep asserts nothing (verification-completeness §1.1): this
	// repository carries hundreds of non-test Go files, so a walk that saw
	// only a handful means the scope broke, not that the code is clean.
	if scanned < 100 {
		t.Fatalf("the scan swept only %d non-test Go files — the walked scope is broken", scanned)
	}
	for rel, n := range hits {
		reason, ok := queuePathSeamExemptions[rel]
		if !ok {
			t.Errorf("%s constructs the legacy queue-path literal %s (%d occurrence(s)); the queue path is built only through the internal/kanban seam", rel, queuePathLiteral, n)
		} else {
			t.Logf("exempt: %s — %s", rel, reason)
		}
	}
	for rel, reason := range queuePathSeamExemptions {
		if hits[rel] == 0 {
			t.Errorf("exemption %s no longer carries the literal — the seam moved; re-point this guard (%s)", rel, reason)
		}
	}
	t.Logf("swept %d non-test Go files under internal/, pkg/, cmd/; %d file(s) carry the literal", scanned, len(hits))
}
