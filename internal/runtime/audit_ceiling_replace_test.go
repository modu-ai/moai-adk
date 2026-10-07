//go:build darwin || linux

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// TestAppendProgressRecordWriteDeniedKeepsFile (sync-audit-4 F8) — a
// write-restricted progress.md is never silently rewritten: the replace
// verifies the original is writable first and a denial returns a clean
// error with the file unchanged and no temp left behind — the pre-repair
// os.WriteFile failure mode.
func TestAppendProgressRecordWriteDeniedKeepsFile(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	original := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(original), 0o444); err != nil {
		t.Fatal(err)
	}
	err := appendProgressRecord(specDir, "- new record")
	if err == nil {
		t.Fatal("a write-restricted progress.md was silently rewritten — the pre-repair os.WriteFile returned permission denied")
	}
	raw, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(raw) != original {
		t.Fatalf("the read-only original was modified:\n%s", raw)
	}
	entries, derr := os.ReadDir(specDir)
	if derr != nil {
		t.Fatal(derr)
	}
	if len(entries) != 1 || entries[0].Name() != "progress.md" {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

// TestAppendProgressRecordSeedFailureAborts (sync-audit-4 F8) — a metadata
// seeding failure ABORTS the replace: the original stays untouched, no
// temp survives, and the caller gets an error — never a mode-only fallback
// rewrite that drops the ACL.
func TestAppendProgressRecordSeedFailureAborts(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	original := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	restore := stubSeedFailure(t)
	defer restore()
	err := appendProgressRecord(specDir, "- new record")
	if err == nil {
		t.Fatal("a seeding failure fell back to a mode-only replace instead of aborting")
	}
	raw, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(raw) != original {
		t.Fatalf("the original was modified by an aborted replace:\n%s", raw)
	}
	entries, derr := os.ReadDir(specDir)
	if derr != nil {
		t.Fatal(derr)
	}
	if len(entries) != 1 || entries[0].Name() != "progress.md" {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

// stubSeedFailure swaps the metadata-seeding function for one that always
// fails, restoring it at test end.
func stubSeedFailure(t *testing.T) func() {
	t.Helper()
	orig := seedFileMetadataFn
	seedFileMetadataFn = func(tmp *os.File, tmpPath, original string) error {
		return errSeedInjected
	}
	return func() { seedFileMetadataFn = orig }
}

var errSeedInjected = errSeedFailure{}

type errSeedFailure struct{}

func (errSeedFailure) Error() string { return "injected seeding failure" }

var _ = strings.TrimSpace // keep strings linked for sibling tests in this file

// TestAppendProgressRecordPreservesOwnership (round-4 edge 6b) — a
// progress.md owned by a different GROUP than the process keeps its
// ownership through the replace: mode+xattr copy alone would re-own the
// file to the process and change who can access it; the seeder chowns the
// temp to the original's owner, and an impossible preservation aborts the
// replace (the F8 posture). darwin||linux: syscall.Stat_t ownership.
func TestAppendProgressRecordPreservesOwnership(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	// Move the file to a supplementary group of this process (a chown the
	// process is allowed to make), so the original's group differs from
	// the group a fresh temp file would get.
	gids, err := syscall.Getgroups()
	if err != nil || len(gids) == 0 {
		t.Skipf("no supplementary groups to test with: %v", err)
	}
	ogid := gidOf(t, path)
	target := ogid
	for _, g := range gids {
		if g != ogid {
			target = g
			break
		}
	}
	if target == ogid {
		t.Skip("process belongs to a single group — ownership preservation untestable here")
	}
	if err := os.Chown(path, -1, target); err != nil {
		t.Skipf("cannot chown the fixture: %v", err)
	}
	if gidOf(t, path) != target {
		t.Skip("the chown did not take effect")
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatalf("the ownership-preserving replace failed: %v", err)
	}
	if got := gidOf(t, path); got != target {
		t.Fatalf("progress.md gid %d, want %d — ownership changed on replace", got, target)
	}
}

func gidOf(t *testing.T, path string) int {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skip("no Stat_t ownership on this platform")
	}
	return int(st.Gid)
}

// TestAppendProgressRecordTempSwapFailsClosed (sync-audit-7 F13) — a
// directory writer that swaps the temp's NAME for a symlink to a
// project-external file must not redirect anything: the fd-held flow
// detects the inode mismatch at the post-seed verification and fails
// closed — the victim is untouched, progress.md keeps its original
// content, and the append reports the error.
func TestAppendProgressRecordTempSwapFailsClosed(t *testing.T) {
	specDir := t.TempDir()
	victimDir := t.TempDir()
	victim := filepath.Join(victimDir, "victim.md")
	victimData := "victim data — must stay intact\n"
	if err := os.WriteFile(victim, []byte(victimData), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := seedFileMetadataFn
	seedFileMetadataFn = func(tmp *os.File, tmpPath, original string) error {
		// The dir-writer swap: move our temp away, put a symlink to the
		// victim at the temp's name, and report success.
		if err := os.Rename(tmpPath, tmpPath+".attacker"); err != nil {
			return err
		}
		return os.Symlink(victim, tmpPath)
	}
	t.Cleanup(func() {
		seedFileMetadataFn = orig
		_ = os.Remove(path + ".attacker")
		_ = os.Remove(path)
	})

	if err := appendProgressRecord(specDir, "- new record"); err == nil {
		t.Fatal("the swapped temp silently completed the replace")
	}
	raw, rerr := os.ReadFile(victim)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(raw) != victimData {
		t.Fatalf("the victim was overwritten:\n%s", raw)
	}
	info, lerr := os.Lstat(path)
	if lerr != nil {
		t.Fatal(lerr)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("progress.md became a symlink to the victim")
	}
	pinned, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(pinned) != pre {
		t.Fatalf("progress.md was modified by the aborted replace:\n%s", pinned)
	}
}

// TestAppendProgressRecordSwapKeepsForeignFile (gate round-46 item 5) —
// when the inode check detects a swap, the swapped-in entry is EXCLUDED
// from cleanup: it is no longer ours to touch. A foreign REGULAR file
// swapped in by the attacker must survive the abort (the F8 posture
// already preserves the original progress.md).
func TestAppendProgressRecordSwapKeepsForeignFile(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(specDir, "foreign.md")
	foreignData := "foreign data — the attacker's file, not ours to delete\n"
	if err := os.WriteFile(foreign, []byte(foreignData), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := seedFileMetadataFn
	var swappedInPath string
	seedFileMetadataFn = func(tmp *os.File, tmpPath, original string) error {
		// The dir-writer swap: move our temp away, replace the name with a
		// REGULAR file of the attacker's own (not a symlink — deletion
		// would destroy it).
		if err := os.Rename(tmpPath, tmpPath+".attacker"); err != nil {
			return err
		}
		swappedInPath = tmpPath
		return os.WriteFile(tmpPath, []byte(foreignData), 0o644)
	}
	t.Cleanup(func() {
		seedFileMetadataFn = orig
		_ = os.Remove(path + ".attacker")
		_ = os.Remove(path)
	})

	if err := appendProgressRecord(specDir, "- new record"); err == nil {
		t.Fatal("the swapped temp silently completed the replace")
	}
	if swappedInPath == "" {
		t.Fatal("the swap never ran")
	}
	// The swapped-in foreign file must SURVIVE the abort: it is no longer
	// ours to touch (gate round-46 item 5).
	raw, rerr := os.ReadFile(swappedInPath)
	if rerr != nil {
		t.Fatalf("the swapped-in foreign regular file was deleted by the cleanup: %v", rerr)
	}
	if string(raw) != foreignData {
		t.Fatalf("the foreign file was modified:\n%s", raw)
	}
}

// TestAppendProgressRecordPreservesHardlink (round-4 edge 7c) — a
// hardlinked progress.md keeps the link relationship across the append: a
// rename would replace only this directory entry's inode and the other
// names would stop seeing records. The append writes IN PLACE through the
// shared inode (the pre-repair os.WriteFile semantics), so both paths show
// the record and nlink is unchanged.
func TestAppendProgressRecordPreservesHardlink(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	mirror := filepath.Join(specDir, "progress-mirror.md")
	if err := os.Link(path, mirror); err != nil {
		t.Skipf("hardlinks unavailable here: %v", err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{path, mirror} {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), "- new record") {
			t.Fatalf("%s does not see the record — the hardlink was broken:\n%s", p, raw)
		}
	}
	i1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	i2, err := os.Stat(mirror)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(i1, i2) {
		t.Fatal("the hardlink relationship was broken by the replace")
	}
}

// TestAppendProgressRecordLandsInUnwritableDir (consolidated item 5) — an
// existing writable progress.md in an UNWRITABLE directory still receives
// the record: the in-place write needs no directory write, exactly as the
// pre-repair os.WriteFile worked. Losing the record because CreateTemp
// cannot run in the directory is the defect.
func TestAppendProgressRecordLandsInUnwritableDir(t *testing.T) {
	specDir := t.TempDir()
	path := filepath.Join(specDir, "progress.md")
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(path, []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(specDir, 0o555); err != nil {
		t.Skipf("cannot make the directory unwritable here: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(specDir, 0o755) })
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatalf("the record was lost to an unwritable directory although the file itself is writable: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if !strings.Contains(content, "- old record") || !strings.Contains(content, "- new record") {
		t.Fatalf("the record did not land:\n%s", content)
	}
}
