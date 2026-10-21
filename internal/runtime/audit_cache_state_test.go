// audit_cache_state_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-020): the audit_cache store persists its entries to disk under the
// project's state root, requires the hash of an EXISTING verdict file as each
// entry's evidence, revalidates that file's existence and SHA-256 before a
// reuse, and refuses to mint the empty-input hash for a non-SPEC directory.
package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeVerdictFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write verdict file: %v", err)
	}
	return p
}

func passResult(reportPath string) *AuditResult {
	return &AuditResult{Verdict: VerdictPass, AuditAt: time.Now(), AuditorVersion: "plan-auditor", ReportPath: reportPath}
}

// TestComputeHash_RefusesNonSpecDirectory pins the empty-input refusal: a
// directory with NO plan artifact returns an error, never the empty-input
// hash — a hash of nothing must not read as a legitimate cache identity.
func TestComputeHash_RefusesNonSpecDirectory(t *testing.T) {
	dir := t.TempDir() // no artifact files at all
	c := NewInMemoryCache()
	if _, err := c.ComputeHash(dir); err == nil {
		t.Fatal("ComputeHash succeeded on a directory with no plan artifacts — the empty-input hash must be refused (REQ-MN-020)")
	}
}

// TestStateCache_StoreRequiresVerdictFile pins the evidence requirement: a
// store whose verdict file does not exist writes NO entry.
func TestStateCache_StoreRequiresVerdictFile(t *testing.T) {
	root := t.TempDir()
	c := NewStateAuditCache(root)
	if err := c.StoreVerdict("SPEC-X-001", "hash-1", passResult(filepath.Join(root, "missing-report.md"))); err == nil {
		t.Fatal("StoreVerdict accepted a missing verdict file — an entry requires the hash of an EXISTING file (REQ-MN-020)")
	}
	if _, ok := c.Lookup("SPEC-X-001", "hash-1", time.Now()); ok {
		t.Fatal("an entry without an existing verdict file was cached")
	}
}

// TestStateCache_PersistsAcrossInstances pins the disk persistence: an entry
// stored by one cache instance is a lookup hit for a FRESH instance — the
// store lives on disk under the project's state root, not in process memory.
func TestStateCache_PersistsAcrossInstances(t *testing.T) {
	root := t.TempDir()
	report := writeVerdictFile(t, root, "review-1.md", "# plan review: pass")
	c := NewStateAuditCache(root)
	if err := c.StoreVerdict("SPEC-X-002", "hash-2", passResult(report)); err != nil {
		t.Fatalf("StoreVerdict: %v", err)
	}

	fresh := NewStateAuditCache(root)
	entry, ok := fresh.Lookup("SPEC-X-002", "hash-2", time.Now())
	if !ok {
		t.Fatal("a freshly constructed cache missed an entry another instance stored — the entries are not on disk (REQ-MN-020)")
	}
	if entry.VerdictFileSHA == "" {
		t.Fatal("the cached entry carries no verdict-file SHA — the evidence hash is missing")
	}
	sum := sha256.Sum256([]byte("# plan review: pass"))
	if entry.VerdictFileSHA != hex.EncodeToString(sum[:]) {
		t.Errorf("verdict_file_sha = %q, want the report's SHA-256", entry.VerdictFileSHA)
	}
}

// TestStateCache_LookupRefusesChangedVerdictFile pins the reuse validation: a
// cached entry whose verdict file changed (or vanished) since the store is a
// MISS, never a hit.
func TestStateCache_LookupRefusesChangedVerdictFile(t *testing.T) {
	root := t.TempDir()
	report := writeVerdictFile(t, root, "review-2.md", "# plan review: pass")
	c := NewStateAuditCache(root)
	if err := c.StoreVerdict("SPEC-X-003", "hash-3", passResult(report)); err != nil {
		t.Fatalf("StoreVerdict: %v", err)
	}

	// The verdict file is rewritten after the store.
	if err := os.WriteFile(report, []byte("# plan review: REWRITTEN"), 0o644); err != nil {
		t.Fatalf("rewrite report: %v", err)
	}
	if _, ok := c.Lookup("SPEC-X-003", "hash-3", time.Now()); ok {
		t.Fatal("a cached entry whose verdict file changed since the store was reused (REQ-MN-020)")
	}
}

// TestStateCache_LookupRefusesSpecIDCollision pins the sync-audit r3 F4 repair:
// sanitizeStateKey folds SPEC IDs that differ only by a separator onto ONE
// file name ("SPEC-A/B" and "SPEC-A_B" both sanitize to SPEC-A_B.json), so a
// lookup must ALSO match the stored SpecID — with identical artifact hashes,
// one SPEC's PASS report must never be served to another.
func TestStateCache_LookupRefusesSpecIDCollision(t *testing.T) {
	root := t.TempDir()
	report := writeVerdictFile(t, root, "review-5.md", "# plan review: pass")
	c := NewStateAuditCache(root)
	if err := c.StoreVerdict("SPEC-A/B", "hash-5", passResult(report)); err != nil {
		t.Fatalf("StoreVerdict: %v", err)
	}
	if _, ok := c.Lookup("SPEC-A_B", "hash-5", time.Now()); ok {
		t.Fatal("a separator-collision lookup was served another SPEC's PASS entry — the stored SpecID is not compared on lookup (sync-audit r3 F4)")
	}
	if _, ok := c.Lookup("SPEC-A/B", "hash-5", time.Now()); !ok {
		t.Fatal("the storing SPEC's own lookup must still hit after the collision repair")
	}
}

// TestStateCache_RelativeReportPathAnchoredToRoot pins the P2-7 repair: a
// RELATIVE verdict report path resolves against the project root the cache is
// bound to — at store AND at lookup — never against whatever cwd the call
// happens to run from.
func TestStateCache_RelativeReportPathAnchoredToRoot(t *testing.T) {
	root := t.TempDir()
	writeVerdictFile(t, root, "review-3.md", "# plan review: pass")
	c := NewStateAuditCache(root)
	// A relative spelling, as a caller inside the project would hand over.
	if err := c.StoreVerdict("SPEC-X-004", "hash-4", passResult("review-3.md")); err != nil {
		t.Fatalf("StoreVerdict with a relative path: %v", err)
	}
	entry, ok := c.Lookup("SPEC-X-004", "hash-4", time.Now())
	if !ok {
		t.Fatal("a relative-path entry did not survive store+lookup — the path was not anchored to the project root (card-review P2-7)")
	}
	if !filepath.IsAbs(entry.ReportPath) {
		t.Errorf("entry.ReportPath = %q, want the absolute root-anchored path", entry.ReportPath)
	}
}
