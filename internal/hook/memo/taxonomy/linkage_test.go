// linkage_test.go — index-linkage and topic-count audits.
//
// The pre-existing audits validate a memory file's own shape (frontmatter,
// type, body structure, duplicate descriptions) and the index's line count.
// None of them reads the index's links, so a topic file that was written but
// never indexed passes every check while being unreachable: only MEMORY.md is
// loaded into a session, so an unlinked file is stored and never recalled.
// These tests pin the two checks that close that gap.
package taxonomy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMemoryFixture lays out a memory directory: an index carrying one link
// line per name in linked, and a topic file per name in present.
func writeMemoryFixture(t *testing.T, linked, present []string) string {
	t.Helper()
	dir := t.TempDir()

	var idx strings.Builder
	idx.WriteString("# Memory Index\n\n")
	for _, name := range linked {
		idx.WriteString("- [Title](" + name + ") — hook\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(idx.String()), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	for _, name := range present {
		body := "---\nname: x\ndescription: d\nmetadata:\n  type: feedback\n---\n\nbody\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func codesOf(findings []AuditFinding) map[AuditCode]int {
	out := map[AuditCode]int{}
	for _, f := range findings {
		out[f.Code]++
	}
	return out
}

// TestAuditLinkageFindsOrphans is the case that actually bit this project:
// files on disk that no index line points at.
func TestAuditLinkageFindsOrphans(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_a.md"},
		[]string{"feedback_a.md", "feedback_b.md", "project_c.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnOrphanNotIndexed]; got != 2 {
		t.Errorf("orphan findings = %d, want 2 (feedback_b, project_c): %+v", got, findings)
	}
	for _, f := range findings {
		if f.Code == WarnOrphanNotIndexed && strings.Contains(f.Path, "feedback_a.md") {
			t.Error("indexed file reported as an orphan")
		}
	}
}

// TestAuditLinkageFindsDanglingLinks covers the other direction: an index
// line whose target was moved or removed.
func TestAuditLinkageFindsDanglingLinks(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_a.md", "feedback_gone.md"},
		[]string{"feedback_a.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if got := codesOf(findings)[WarnDanglingIndexLink]; got != 1 {
		t.Errorf("dangling findings = %d, want 1: %+v", got, findings)
	}
}

// TestAuditLinkageCleanIsSilent keeps the audit quiet when the index and the
// directory agree — a check that always fires teaches people to ignore it.
func TestAuditLinkageCleanIsSilent(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t,
		[]string{"feedback_a.md", "project_b.md"},
		[]string{"feedback_a.md", "project_b.md"})

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("clean fixture produced findings: %+v", findings)
	}
}

// TestAuditLinkageIgnoresArchive pins that archived files are out of scope:
// archiving is the remedy the cap prescribes, so counting archived files as
// orphans would make the remedy trip the alarm.
func TestAuditLinkageIgnoresArchive(t *testing.T) {
	t.Parallel()
	dir := writeMemoryFixture(t, []string{"feedback_a.md"}, []string{"feedback_a.md"})
	arch := filepath.Join(dir, "_archive")
	if err := os.MkdirAll(arch, 0o755); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(arch, "project_old.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatalf("write archived: %v", err)
	}

	findings, err := AuditLinkage(dir)
	if err != nil {
		t.Fatalf("AuditLinkage: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("archived file produced findings: %+v", findings)
	}
}

// TestAuditTopicCountOverCap pins the constitution's per-project ceiling,
// which had no checker at all.
func TestAuditTopicCountOverCap(t *testing.T) {
	t.Parallel()
	names := make([]string, 0, 4)
	for _, n := range []string{"feedback_a.md", "feedback_b.md", "feedback_c.md", "feedback_d.md"} {
		names = append(names, n)
	}
	dir := writeMemoryFixture(t, names, names)

	findings, err := AuditTopicCount(dir, 3)
	if err != nil {
		t.Fatalf("AuditTopicCount: %v", err)
	}
	if got := codesOf(findings)[WarnTopicCountOverCap]; got != 1 {
		t.Errorf("over-cap findings = %d, want 1: %+v", got, findings)
	}
	if len(findings) == 1 && !strings.Contains(findings[0].Detail, "4") {
		t.Errorf("detail should name the observed count: %q", findings[0].Detail)
	}

	under, err := AuditTopicCount(dir, 10)
	if err != nil {
		t.Fatalf("AuditTopicCount under cap: %v", err)
	}
	if len(under) != 0 {
		t.Errorf("under cap produced findings: %+v", under)
	}
}

// TestAuditMissingDirIsNotAnError keeps both audits fail-open: a project with
// no memory directory yet is not a defect.
func TestAuditMissingDirIsNotAnError(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "nope")

	if f, err := AuditLinkage(missing); err != nil || len(f) != 0 {
		t.Errorf("AuditLinkage(missing) = %+v, %v; want nil, nil", f, err)
	}
	if f, err := AuditTopicCount(missing, 50); err != nil || len(f) != 0 {
		t.Errorf("AuditTopicCount(missing) = %+v, %v; want nil, nil", f, err)
	}
}
