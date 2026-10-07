//go:build darwin || linux

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAppendProgressRecordWritesThroughSymlink (sync-audit-2 F4) — a
// progress.md that is itself a symlink (a dotfiles-managed file, for
// example) stays a symlink: the record lands in the link's TARGET, the
// path is still a symlink after the append, and no regular file
// materializes inside the tracked SPEC directory.
func TestAppendProgressRecordWritesThroughSymlink(t *testing.T) {
	specDir := t.TempDir()
	targetDir := t.TempDir() // deliberately outside the SPEC directory
	target := filepath.Join(targetDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(target, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(specDir, "progress.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the progress.md symlink was replaced by a regular file (mode %v)", info.Mode())
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if !strings.Contains(content, "- old record") || !strings.Contains(content, "- new record") {
		t.Fatalf("the target does not carry both records:\n%s", content)
	}
	if !strings.HasSuffix(content, "- new record\n") {
		t.Fatalf("the record did not append at the target's end:\n%s", content)
	}
}

// TestAppendProgressRecordWritesThroughDanglingSymlink (round-4 edge 4) —
// a symlink whose TARGET does not exist yet (its parent directory does):
// the record writes through to the newly created target, exactly as the
// pre-repair os.WriteFile followed the link and created it, and the link
// survives as a symlink.
func TestAppendProgressRecordWritesThroughDanglingSymlink(t *testing.T) {
	specDir := t.TempDir()
	targetDir := t.TempDir()
	target := filepath.Join(targetDir, "progress.md") // deliberately absent
	link := filepath.Join(specDir, "progress.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	if err := appendProgressRecord(specDir, "- first record"); err != nil {
		t.Fatalf("a dangling-symlink progress.md refused the record the pre-repair write-through would have created: %v", err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the progress.md symlink was replaced by a regular file (mode %v)", info.Mode())
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("the target was not created through the link: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, progressSectionHeading) || !strings.Contains(content, "- first record") {
		t.Fatalf("the created target does not carry the record:\n%s", content)
	}
}

// TestAppendProgressRecordWritesThroughSymlinkChain (gate finding 7) — a
// symlink CHAIN (progress.md → alias.md → target) resolves to the FINAL
// referent: the record lands in the target created through the chain, and
// every intermediate link survives as a symlink — resolving one hop would
// replace the midlink with a regular file and strand the record.
func TestAppendProgressRecordWritesThroughSymlinkChain(t *testing.T) {
	specDir := t.TempDir()
	targetDir := t.TempDir()
	target := filepath.Join(targetDir, "real.md") // deliberately absent
	alias := filepath.Join(specDir, "alias.md")
	if err := os.Symlink(target, alias); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	link := filepath.Join(specDir, "progress.md")
	if err := os.Symlink(alias, link); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	if err := appendProgressRecord(specDir, "- first record"); err != nil {
		t.Fatalf("the symlink chain refused the record: %v", err)
	}
	for _, p := range []string{link, alias} {
		info, err := os.Lstat(p)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("%s was replaced by a regular file — the chain was not followed to the end", p)
		}
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("the final target was not created through the chain: %v", err)
	}
	if !strings.Contains(string(raw), progressSectionHeading) || !strings.Contains(string(raw), "- first record") {
		t.Fatalf("the final target does not carry the record:\n%s", raw)
	}
}

// TestAppendProgressRecordResolvesDotDotThroughSymlink (consolidated item
// 3, gate round-38) — a referent carrying `..` must apply filesystem
// order: resolve `hop` (a symlink into another directory) FIRST, then
// apply `..` to the RESOLVED location, then the final name. A resolver
// that pre-cleans `hop/../actual.md` against the link's own directory
// selects the wrong file and materializes a stray.
func TestAppendProgressRecordResolvesDotDotThroughSymlink(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "spec")
	otherDir := filepath.Join(root, "other")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The REAL actual.md: `hop` points into otherDir, so `hop/../actual.md`
	// means PARENT-of-otherDir/actual.md — applying `..` to the RESOLVED
	// hop lands here.
	real := filepath.Join(root, "actual.md")
	if err := os.WriteFile(real, []byte("# progress\n\n## §G Override and Refusal Record\n\n- old record\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hop := filepath.Join(specDir, "hop")
	if err := os.Symlink(otherDir, hop); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	link := filepath.Join(specDir, "progress.md")
	// The referent is the RAW string "hop/../actual.md" — filepath.Join
	// would pre-clean the `..` away before the kernel ever stored it, and
	// the defect (pre-cleaning inside the resolver) would never fire.
	if err := os.Symlink("hop/../actual.md", link); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatalf("the hop/../ referent refused the record: %v", err)
	}
	// The record landed in the REAL actual.md.
	raw, err := os.ReadFile(real)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "- new record") {
		t.Fatalf("the real actual.md does not carry the record:\n%s", raw)
	}
	// Nothing stray appeared in the spec directory.
	stray := filepath.Join(specDir, "actual.md")
	if _, serr := os.Stat(stray); !os.IsNotExist(serr) {
		t.Fatalf("a stray actual.md was materialized in the spec directory (pre-cleaned resolution)")
	}
	for _, p := range []string{hop, link} {
		info, lerr := os.Lstat(p)
		if lerr != nil {
			t.Fatal(lerr)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("%s was replaced by a regular file", p)
		}
	}
}
