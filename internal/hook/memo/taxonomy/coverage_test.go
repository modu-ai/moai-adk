// coverage_test.go contains supplemental tests to reach the 90% coverage target
// for the taxonomy sub-package (SPEC-V3R2-EXT-001 acceptance.md DoD §5.2).
package taxonomy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// TestDetectStale_DefaultThreshold verifies that thresholdHours <= 0 falls back
// to the config default (24h) and still correctly classifies files.
func TestDetectStale_DefaultThreshold(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeMemoryFile(t, filepath.Join(dir, "stale.md"), 25*time.Hour)
	writeMemoryFile(t, filepath.Join(dir, "fresh.md"), 1*time.Hour)

	reports, err := taxonomy.DetectStale(dir, 0, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale(0 threshold) error = %v", err)
	}
	if len(reports) != 1 {
		t.Errorf("len(reports) = %d, want 1 (only stale.md should be flagged)", len(reports))
	}
}

// TestDetectStale_NonMdFilesSkipped verifies that non-.md files are ignored.
func TestDetectStale_NonMdFilesSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Write a very old .txt file (should be skipped)
	old := filepath.Join(dir, "old.txt")
	if err := os.WriteFile(old, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	modTime := fixedNow.Add(-48 * time.Hour)
	if err := os.Chtimes(old, modTime, modTime); err != nil {
		t.Fatal(err)
	}

	reports, err := taxonomy.DetectStale(dir, 24, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale error = %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("len(reports) = %d, want 0 (non-.md files must be skipped)", len(reports))
	}
}

// TestDetectStale_DirectorySkipped verifies that subdirectories are skipped.
func TestDetectStale_DirectorySkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create a subdirectory (should be skipped)
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	reports, err := taxonomy.DetectStale(dir, 24, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale error = %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("len(reports) = %d, want 0 (directories must be skipped)", len(reports))
	}
}

// TestAuditIndex_DefaultCap verifies that lineCap <= 0 falls back to config default.
func TestAuditIndex_DefaultCap(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "MEMORY.md")
	// 201 lines triggers overflow at default cap (200)
	content := strings.Repeat("- line\n", 201)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditIndex(path, 0)
	if err != nil {
		t.Fatalf("AuditIndex error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnIndexOverflow) {
		t.Error("want MEMORY_INDEX_OVERFLOW when using default cap; got no finding")
	}
}

// TestAuditIndex_NonExistentFile verifies that a missing file returns no error
// and no findings.
func TestAuditIndex_NonExistentFile(t *testing.T) {
	t.Parallel()
	findings, err := taxonomy.AuditIndex("/nonexistent/path/MEMORY.md", 200)
	if err != nil {
		t.Fatalf("AuditIndex(non-existent) error = %v, want nil", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none", findings)
	}
}

// TestAuditDuplicates_EmptyDir verifies that an empty dir returns no error.
func TestAuditDuplicates_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	findings, err := taxonomy.AuditDuplicates(dir)
	if err != nil {
		t.Fatalf("AuditDuplicates(empty dir) error = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none", findings)
	}
}

// TestAuditDuplicates_NonExistentDir verifies graceful handling of missing dir.
func TestAuditDuplicates_NonExistentDir(t *testing.T) {
	t.Parallel()
	findings, err := taxonomy.AuditDuplicates("/nonexistent/path")
	if err != nil {
		t.Fatalf("AuditDuplicates(non-existent) error = %v, want nil", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %v, want none", findings)
	}
}

// TestAuditDuplicates_SkipsMEMORYmd verifies MEMORY.md is excluded from duplicate scan.
func TestAuditDuplicates_SkipsMEMORYmd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sharedDesc := "same description"
	// MEMORY.md with matching description — should NOT be counted.
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"),
		[]byte("---\nname: index\ndescription: "+sharedDesc+"\ntype: user\n---\n"),
		0o644); err != nil {
		t.Fatal(err)
	}
	writeMemFile(t, dir, "a.md", "user", sharedDesc)

	findings, err := taxonomy.AuditDuplicates(dir)
	if err != nil {
		t.Fatalf("AuditDuplicates error = %v", err)
	}
	if hasCode(findings, taxonomy.WarnDuplicate) {
		t.Error("MEMORY.md must be excluded from duplicate detection")
	}
}

// TestAuditDuplicates_EmptyDescription verifies files with empty descriptions are skipped.
func TestAuditDuplicates_EmptyDescription(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Two files, both with empty description — no duplicate warning
	for _, name := range []string{"a.md", "b.md"} {
		content := "---\nname: " + name + "\ntype: user\n---\nbody\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	findings, err := taxonomy.AuditDuplicates(dir)
	if err != nil {
		t.Fatalf("AuditDuplicates error = %v", err)
	}
	if hasCode(findings, taxonomy.WarnDuplicate) {
		t.Error("empty descriptions must not trigger MEMORY_DUPLICATE")
	}
}

// TestAuditFile_ValidProjectBodyStructure verifies project-type files are checked
// for **Why:** and **How to apply:** markers.
func TestAuditFile_ValidProjectBodyStructure(t *testing.T) {
	t.Parallel()
	findings, err := taxonomy.AuditFile(fixturesPath("project_migration.md"))
	if err != nil {
		t.Fatalf("AuditFile(project_migration.md) error = %v", err)
	}
	if hasCode(findings, taxonomy.WarnBodyStructureMissing) {
		t.Error("want no MEMORY_BODY_STRUCTURE_MISSING for well-formed project memory")
	}
}

// TestAuditFile_ProjectMissingBodyStructure verifies that a project-type file
// without Why/How markers triggers MEMORY_BODY_STRUCTURE_MISSING.
func TestAuditFile_ProjectMissingBodyStructure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "project.md")
	content := "---\nname: proj\ndescription: d\ntype: project\n---\nJust a plain project note.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditFile(path)
	if err != nil {
		t.Fatalf("AuditFile error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnBodyStructureMissing) {
		t.Error("want MEMORY_BODY_STRUCTURE_MISSING for project file without markers")
	}
}

// TestAuditFile_GitHistoryExcluded verifies the git_history excluded category.
func TestAuditFile_GitHistoryExcluded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "git.md")
	content := "---\nname: git\ndescription: d\ntype: reference\n---\n\ngit log --oneline HEAD~5\ngit blame src/main.go\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditFile(path)
	if err != nil {
		t.Fatalf("AuditFile error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnExcludedCategory) {
		t.Errorf("want MEMORY_EXCLUDED_CATEGORY for git_history content; got %v", findings)
	}
}

// TestAuditFile_EphemeralStateExcluded verifies ephemeral_state excluded category.
func TestAuditFile_EphemeralStateExcluded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ephem.md")
	content := "---\nname: e\ndescription: d\ntype: project\n---\n\ncurrently working on the auth refactor.\n\n**Why:** test\n\n**How to apply:** test\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditFile(path)
	if err != nil {
		t.Fatalf("AuditFile error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnExcludedCategory) {
		t.Errorf("want MEMORY_EXCLUDED_CATEGORY for ephemeral_state content; got %v", findings)
	}
}

// TestParseFile_UnclosedFrontmatter verifies that frontmatter with no closing "---"
// returns ErrNoFrontmatter.
func TestParseFile_UnclosedFrontmatter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "unclosed.md")
	content := "---\nname: test\ntype: user\nno closing delimiter here\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := taxonomy.ParseFile(path)
	if err == nil {
		t.Error("ParseFile(unclosed frontmatter) = nil error, want ErrNoFrontmatter")
	}
}

// TestAggregateWarning_Zero verifies empty input returns empty string.
func TestAggregateWarning_Zero(t *testing.T) {
	t.Parallel()
	msg := taxonomy.AggregateWarning(nil)
	if msg != "" {
		t.Errorf("AggregateWarning(nil) = %q, want empty", msg)
	}
	msg = taxonomy.AggregateWarning([]taxonomy.StaleReport{})
	if msg != "" {
		t.Errorf("AggregateWarning([]) = %q, want empty", msg)
	}
}

// TestAuditFile_NoFrontmatterFile verifies that a file without frontmatter
// triggers MEMORY_MISSING_TYPE through the ErrNoFrontmatter path in AuditFile.
func TestAuditFile_NoFrontmatterFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "nofm.md")
	if err := os.WriteFile(path, []byte("# Just a heading\n\nbody only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditFile(path)
	if err != nil {
		t.Fatalf("AuditFile(no frontmatter) error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnMissingType) {
		t.Errorf("want MEMORY_MISSING_TYPE for file without frontmatter; got %v", findings)
	}
}

// TestAuditFile_SymlinkSkipped verifies that symlink audit returns no error and no findings.
func TestAuditFile_SymlinkSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	real := filepath.Join(dir, "real.md")
	if err := os.WriteFile(real, []byte("---\nname: t\ndescription: d\ntype: user\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.md")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("symlink not supported on this OS")
	}
	findings, err := taxonomy.AuditFile(link)
	if err != nil {
		t.Fatalf("AuditFile(symlink) error = %v, want nil (symlink is silently skipped)", err)
	}
	if len(findings) != 0 {
		t.Errorf("AuditFile(symlink) = %v, want no findings (symlink skipped)", findings)
	}
}

// TestParseFile_LstatError verifies that ParseFile propagates lstat errors properly.
func TestParseFile_LstatError(t *testing.T) {
	t.Parallel()
	// Passing a path inside a non-existent directory forces lstat to fail.
	_, _, err := taxonomy.ParseFile("/nonexistent/dir/file.md")
	if err == nil {
		t.Error("ParseFile(non-existent path) = nil error, want error")
	}
}

// TestDetectStale_MultipleFiles verifies multiple stale files are all reported.
func TestDetectStale_MultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		name := filepath.Join(dir, fmt.Sprintf("mem%d.md", i))
		writeMemoryFile(t, name, 25*time.Hour)
	}
	// One fresh file
	writeMemoryFile(t, filepath.Join(dir, "fresh.md"), 1*time.Hour)

	reports, err := taxonomy.DetectStale(dir, 24, fixedNow)
	if err != nil {
		t.Fatalf("DetectStale error = %v", err)
	}
	if len(reports) != 3 {
		t.Errorf("len(reports) = %d, want 3", len(reports))
	}
}
