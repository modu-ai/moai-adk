// m1b_schema_preserve_test.go — SPEC-USERASSET-DEPLOY-GUARD-001, gate round
// 14 bucket B3 (REQ-JRN-004 preserve-in-place): an unsupported-schema
// journal must be refused WITHOUT the corrupt-sidecar rename. The corrupt
// path hides the file under a sidecar name no compatible binary will find,
// and the second run then succeeds WITHOUT the journal — silently dropping
// the interrupted install's recorded selection (measured defect: extras
// choice lost on the second run).
package userassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJournalUnsupportedSchemaPreservedInPlace(t *testing.T) {
	f := newFixture(t)
	moai := filepath.Join(f.home, ".moai")
	raw := `{
  "schema_version": 99,
  "bundles_selection": ["extras"],
  "entries": [{"path": "claude-skills/moai-alpha/SKILL.md", "expected_sha256": "ab"}],
  "started_at": "2026-10-08T00:00:00Z"
}
`
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(JournalPath(f.home), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run 1: refused, and the journal stays AT ITS ORIGINAL PATH.
	if _, err := f.installer(t).Install(nil); err == nil {
		t.Fatal("an unsupported-schema journal was silently decoded — REQ-JRN-004 regressed")
	} else if !strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("the refusal does not carry the schema diagnosis: %v", err)
	}
	got, err := os.ReadFile(JournalPath(f.home))
	if err != nil {
		t.Fatalf("the journal left its original path after the refused run: %v", err)
	}
	if !strings.Contains(string(got), "\"schema_version\": 99") {
		t.Fatalf("the journal at the original path was rewritten: %s", got)
	}
	for _, e := range moaiEntries(t, moai) {
		if strings.HasPrefix(e, "user-assets-journal.json.corrupt-") {
			t.Fatalf("the unsupported-schema journal was routed to the corrupt sidecar %s — a compatible binary can no longer find it", e)
		}
	}

	// Run 2: refuses AGAIN (keeps rejecting until a compatible binary
	// recovers it) — never succeeds-without-the-journal.
	if _, err := f.installer(t).Install(nil); err == nil {
		t.Fatal("the second run succeeded while the unsupported-schema journal was still in place — the recovery data was silently dropped")
	}
	if _, err := os.ReadFile(JournalPath(f.home)); err != nil {
		t.Fatalf("the journal did not survive the second refused run: %v", err)
	}
}

// TestJournalFutureSchemaTypeConflictPreservedInPlace — gate round 17: the
// version header is validated BEFORE the body decode. A future schema may
// change body field TYPES; the body decode would fail such a journal into
// the corrupt sidecar and hide it from the binary that could recover it.
// The header peek routes it to the preserve-in-place refusal instead.
func TestJournalFutureSchemaTypeConflictPreservedInPlace(t *testing.T) {
	f := newFixture(t)
	moai := filepath.Join(f.home, ".moai")
	raw := `{
  "schema_version": 99,
  "bundles_selection": ["extras"],
  "entries": "not-an-array-in-a-future-schema",
  "started_at": "2026-10-08T00:00:00Z"
}
`
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(JournalPath(f.home), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := f.installer(t).Install(nil)
	if err == nil {
		t.Fatal("a future-schema journal was silently decoded")
	}
	var schemaErr *JournalSchemaError
	if !AsJournalSchemaError(err, &schemaErr) {
		t.Fatalf("the type-conflicting future schema routed to the corrupt path instead of the in-place schema refusal: %v", err)
	}
	if schemaErr.Found != 99 {
		t.Fatalf("schema diagnosis carries the wrong version: %d", schemaErr.Found)
	}
	if _, readErr := os.ReadFile(JournalPath(f.home)); readErr != nil {
		t.Fatalf("the journal left its original path: %v", readErr)
	}
	for _, e := range moaiEntries(t, moai) {
		if strings.HasPrefix(e, "user-assets-journal.json.corrupt-") {
			t.Fatalf("the future-schema journal was routed to the corrupt sidecar %s", e)
		}
	}
}

// moaiEntries lists the .moai directory's entry names.
func moaiEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
