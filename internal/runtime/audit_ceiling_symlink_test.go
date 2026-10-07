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
