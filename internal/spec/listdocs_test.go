package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// SPEC-WEB-CONSOLE-011 M5 — exported read-only SPEC catalog lister (REQ-WC11-041)
// + optional Tier frontmatter field (REQ-WC11-042).
//
// ListDocs is the exported wrapper around the unexported discoverSPECs/parseSPECDoc
// helpers. It powers the READ-ONLY web SPEC board — it MUST NOT mutate any file.

// writeListDocsFixture creates baseDir/.moai/specs/<id>/spec.md with the given
// frontmatter body (named to avoid collision with the drift-test writeSpecFixture).
func writeListDocsFixture(t *testing.T, baseDir, id, frontmatter string) {
	t.Helper()
	dir := filepath.Join(baseDir, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	content := "---\n" + frontmatter + "---\n\n# " + id + "\n\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
}

func TestListDocs_DiscoversAndParsesFrontmatter(t *testing.T) {
	root := t.TempDir()
	writeListDocsFixture(t, root, "SPEC-ALPHA-001",
		"id: SPEC-ALPHA-001\ntitle: \"Alpha feature\"\nstatus: implemented\nupdated: 2026-07-01\ntier: L\n")
	writeListDocsFixture(t, root, "SPEC-BETA-002",
		"id: SPEC-BETA-002\ntitle: \"Beta feature\"\nstatus: completed\nupdated: 2026-07-02\n")

	records, err := ListDocs(root)
	if err != nil {
		t.Fatalf("ListDocs error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Deterministic order (sorted by path → SPEC-ALPHA before SPEC-BETA).
	if records[0].Frontmatter.ID != "SPEC-ALPHA-001" {
		t.Errorf("record[0] ID = %q, want SPEC-ALPHA-001", records[0].Frontmatter.ID)
	}
	if records[0].Frontmatter.Status != "implemented" {
		t.Errorf("record[0] Status = %q, want implemented", records[0].Frontmatter.Status)
	}
	if records[0].Frontmatter.Tier != "L" {
		t.Errorf("record[0] Tier = %q, want L (REQ-WC11-042)", records[0].Frontmatter.Tier)
	}
	if records[0].ParseError != nil {
		t.Errorf("record[0] ParseError = %v, want nil", records[0].ParseError)
	}

	// REQ-WC11-042: Tier is OPTIONAL — absent → empty string, not an error.
	if records[1].Frontmatter.Tier != "" {
		t.Errorf("record[1] Tier = %q, want empty (tier absent)", records[1].Frontmatter.Tier)
	}
	if records[1].Frontmatter.Status != "completed" {
		t.Errorf("record[1] Status = %q, want completed", records[1].Frontmatter.Status)
	}
}

func TestListDocs_MissingSpecsDir_ReturnsEmpty(t *testing.T) {
	root := t.TempDir() // no .moai/specs subtree
	records, err := ListDocs(root)
	if err != nil {
		t.Fatalf("ListDocs on missing specs dir should not error, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records for missing specs dir, got %d", len(records))
	}
}

func TestListDocs_EmptyBaseDirDefaultsToCwd(t *testing.T) {
	// Empty baseDir must default to "." without panicking. In the test's working
	// directory there is no ./.moai/specs, so an empty (or populated) slice with
	// no error is the contract.
	if _, err := ListDocs(""); err != nil {
		t.Fatalf("ListDocs(\"\") should not error, got %v", err)
	}
}

func TestListDocs_ParseErrorSurfacedNotFatal(t *testing.T) {
	root := t.TempDir()
	// A spec.md with no frontmatter delimiter → parse error, surfaced per-record.
	dir := filepath.Join(root, ".moai", "specs", "SPEC-BROKEN-003")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("no frontmatter here\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// A valid sibling so the lister does not abort the whole scan.
	writeListDocsFixture(t, root, "SPEC-OK-004",
		"id: SPEC-OK-004\ntitle: \"ok\"\nstatus: draft\nupdated: 2026-07-03\n")

	records, err := ListDocs(root)
	if err != nil {
		t.Fatalf("ListDocs error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records (broken + ok), got %d", len(records))
	}
	var broken *DocRecord
	for i := range records {
		if filepath.Base(filepath.Dir(records[i].Path)) == "SPEC-BROKEN-003" {
			broken = &records[i]
		}
	}
	if broken == nil {
		t.Fatal("SPEC-BROKEN-003 record missing")
	}
	if broken.ParseError == nil {
		t.Error("expected ParseError on the malformed spec.md, got nil")
	}
}
