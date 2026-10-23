// sets_test.go — SPEC-AGENT-CORE4-001 REQ-AC4-001/008/009: the agent source set
// is one-to-one with the rendered definitions, the read-only check fails on a
// hand edit without touching a file, and a passing check reports its summary.
package agentemit_test

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// fixtureRoot builds a temporary repository root holding the given agent
// sources under src/ and returns it with a four-set layout over out/ and proj/.
func fixtureRoot(t *testing.T, sources map[string]string) (string, agentemit.Layout) {
	t.Helper()
	root := t.TempDir()
	for name, content := range sources {
		writeFile(t, filepath.Join(root, "src", name), []byte(content))
	}
	return root, agentemit.Layout{
		SourceDir: "src",
		Targets: []agentemit.Target{
			{Name: "claude_template", Dir: "out/c2"},
			{Name: "codex_template", Dir: "out/c3", Codex: true},
			{Name: "claude_project", Dir: "proj/c1", Only: []string{"alpha"}},
			{Name: "codex_project", Dir: "proj/c3p", Codex: true, Only: []string{"alpha"}},
		},
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func twoSources() map[string]string {
	return map[string]string{
		"alpha.md": fixtureMD("alpha", "Alpha.\n", "Read, Grep", "inherit", "low", nil, "\n# alpha\n"),
		"beta.md":  fixtureMD("beta", "Beta.\n", "Read, Grep", "inherit", "low", nil, "\n# beta\n"),
	}
}

// listNames returns the sorted file names directly inside dir with the suffix.
func listNames(t *testing.T, dir, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			out = append(out, strings.TrimSuffix(e.Name(), suffix))
		}
	}
	sort.Strings(out)
	return out
}

func targetDir(t *testing.T, lay agentemit.Layout, name string) string {
	t.Helper()
	for _, tg := range lay.Targets {
		if tg.Name == name {
			return filepath.Join(repoRoot, filepath.FromSlash(tg.Dir))
		}
	}
	t.Fatalf("layout has no target %q", name)
	return ""
}

func TestSourceSetIsOneToOne(t *testing.T) {
	t.Run("twelve_sources_twelve_definitions", func(t *testing.T) {
		lay := agentemit.DefaultLayout()
		sources := listNames(t, filepath.Join(repoRoot, filepath.FromSlash(lay.SourceDir)), ".md")
		c2 := listNames(t, targetDir(t, lay, "claude_template"), ".md")
		c3 := listNames(t, targetDir(t, lay, "codex_template"), ".toml")
		if len(sources) == 0 {
			t.Fatal("agent source directory holds no sources")
		}
		if strings.Join(sources, ",") != strings.Join(c2, ",") {
			t.Errorf("sources %v and C2 definitions %v are not one-to-one", sources, c2)
		}
		if strings.Join(sources, ",") != strings.Join(c3, ",") {
			t.Errorf("sources %v and C3 definitions %v are not one-to-one", sources, c3)
		}
	})
	t.Run("rendered_file_without_source_fails", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		man := mustManifest(t)
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if _, err := agentemit.Check(root, lay, man); err != nil {
			t.Fatalf("Check on a fresh tree: %v", err)
		}
		writeFile(t, filepath.Join(root, "out/c2/orphan.md"), []byte("orphan"))
		sum, err := agentemit.Check(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "orphan.md") {
			t.Errorf("a rendered file without a source must fail naming orphan.md, got %v", err)
		}
		if len(sum.Diffs) != 1 {
			t.Errorf("want exactly the orphan reported, got %v", sum.Diffs)
		}
	})
}

// snapshotTree records the content hash and modification time of every file
// under root, so a check can be shown to have written nothing.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		out[p] = fmt.Sprintf("%x %d", sha256.Sum256(data), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

func sameSnapshot(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// TestEmitCheckFailsOnHandEdit generates every set of the shipped layout from
// the shipped sources into a scratch root, hand-edits one file per set, and
// requires the read-only check to fail naming that file and to leave every
// file untouched.
func TestEmitCheckFailsOnHandEdit(t *testing.T) {
	man := mustManifest(t)
	lay := agentemit.DefaultLayout()
	root := t.TempDir()
	srcDir := filepath.Join(repoRoot, filepath.FromSlash(lay.SourceDir))
	for _, name := range listNames(t, srcDir, ".md") {
		data, err := os.ReadFile(filepath.Join(srcDir, name+".md"))
		if err != nil {
			t.Fatalf("read source %s: %v", name, err)
		}
		writeFile(t, filepath.Join(root, filepath.FromSlash(lay.SourceDir), name+".md"), data)
	}
	if _, err := agentemit.Update(root, lay, man); err != nil {
		t.Fatalf("Update over the shipped sources: %v", err)
	}

	before := snapshotTree(t, root)
	sum, err := agentemit.Check(root, lay, man)
	if err != nil || len(sum.Diffs) != 0 {
		t.Fatalf("Check on a freshly generated tree: %+v, %v", sum, err)
	}
	t.Run("files_unmodified_after_pass", func(t *testing.T) {
		if !sameSnapshot(before, snapshotTree(t, root)) {
			t.Error("a passing check modified the tree")
		}
	})

	for _, tg := range lay.Targets {
		t.Run(tg.Name, func(t *testing.T) {
			dir := filepath.Join(root, filepath.FromSlash(tg.Dir))
			ext := ".md"
			if tg.Codex {
				ext = ".toml"
			}
			names := listNames(t, dir, ext)
			if len(names) == 0 {
				t.Fatalf("set %s holds no files", tg.Name)
			}
			victim := filepath.Join(dir, names[0]+ext)
			orig, err := os.ReadFile(victim)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(victim, append(append([]byte{}, orig...), []byte("hand edit\n")...), 0o644); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.WriteFile(victim, orig, 0o644); err != nil {
					t.Fatal(err)
				}
			}()
			edited := snapshotTree(t, root)
			_, err = agentemit.Check(root, lay, man)
			if err == nil || !strings.Contains(err.Error(), names[0]+ext) {
				t.Errorf("a hand edit of %s must fail the check naming the file, got %v", victim, err)
			}
			if !sameSnapshot(edited, snapshotTree(t, root)) {
				t.Error("a failing check modified the tree")
			}
		})
	}
	t.Run("files_unmodified_after_fail", func(t *testing.T) {
		// A failing check must not repair the hand edit it reports: edit a file,
		// run the check, and require the edited tree to be byte- and
		// time-identical afterwards.
		dir := filepath.Join(root, filepath.FromSlash(lay.Targets[0].Dir))
		victim := filepath.Join(dir, listNames(t, dir, ".md")[0]+".md")
		if err := os.WriteFile(victim, []byte("hand edit, not repaired by the check"), 0o644); err != nil {
			t.Fatal(err)
		}
		edited := snapshotTree(t, root)
		if _, err := agentemit.Check(root, lay, man); err == nil {
			t.Fatal("the check must fail on a hand-edited file")
		}
		if !sameSnapshot(edited, snapshotTree(t, root)) {
			t.Error("a failing check modified the tree")
		}
	})
}

func TestUpdateThenCheckRoundTrip(t *testing.T) {
	root, lay := fixtureRoot(t, twoSources())
	man := mustManifest(t)
	sum, err := agentemit.Update(root, lay, man)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if sum.Specs != 2 || sum.Copies != 4 {
		t.Errorf("Update summary specs=%d copies=%d, want 2 and 4", sum.Specs, sum.Copies)
	}
	// Only-sets cover just the named agent; the full sets cover both.
	if got := listNames(t, filepath.Join(root, "out/c2"), ".md"); strings.Join(got, ",") != "alpha,beta" {
		t.Errorf("C2 set = %v", got)
	}
	if got := listNames(t, filepath.Join(root, "proj/c1"), ".md"); strings.Join(got, ",") != "alpha" {
		t.Errorf("project Claude set = %v, want only alpha", got)
	}
	if got := listNames(t, filepath.Join(root, "proj/c3p"), ".toml"); strings.Join(got, ",") != "alpha" {
		t.Errorf("project Codex set = %v, want only alpha", got)
	}
	chk, err := agentemit.Check(root, lay, man)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if chk.String() != "specs=2 copies=4 diff=0" {
		t.Errorf("summary line = %q, want %q", chk.String(), "specs=2 copies=4 diff=0")
	}
}

// TestUpdateRefusesSymlinkTargets pins that Update never writes through a
// symlink: a link at a target file path, or at a target directory, is refused
// with an error naming the path, and the file it points at stays unchanged.
func TestUpdateRefusesSymlinkTargets(t *testing.T) {
	man := mustManifest(t)
	const victimBytes = "outside file: must stay untouched\n"

	t.Run("symlink_target_file_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		victim := filepath.Join(t.TempDir(), "victim.md")
		writeFile(t, victim, []byte(victimBytes))
		link := filepath.Join(root, "out/c2/alpha.md")
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Symlink(victim, link); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "out/c2/alpha.md") || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("Update must refuse the symlink target naming its path, got %v", err)
		}
		got, rerr := os.ReadFile(victim)
		if rerr != nil || string(got) != victimBytes {
			t.Errorf("the file behind the symlink changed: %q, %v", got, rerr)
		}
		if fi, lerr := os.Lstat(link); lerr != nil || fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("the refused symlink must stay in place, got %v, %v", fi, lerr)
		}
	})
	t.Run("symlink_target_directory_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		outside := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "out"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "out/c2")); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "out/c2") || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("Update must refuse the symlink directory naming its path, got %v", err)
		}
		entries, rerr := os.ReadDir(outside)
		if rerr != nil || len(entries) != 0 {
			t.Errorf("a file was written through the directory symlink: %v, %v", entries, rerr)
		}
	})
	t.Run("non_regular_target_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		if err := os.MkdirAll(filepath.Join(root, "out/c2/alpha.md"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "out/c2/alpha.md") || !strings.Contains(err.Error(), "non-regular") {
			t.Fatalf("Update must refuse a directory at a target file path naming it, got %v", err)
		}
	})
	t.Run("file_in_place_of_directory_refused", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		writeFile(t, filepath.Join(root, "out/c2"), []byte("not a directory\n"))
		_, err := agentemit.Update(root, lay, man)
		if err == nil || !strings.Contains(err.Error(), "out/c2") || !strings.Contains(err.Error(), "not a directory") {
			t.Fatalf("Update must refuse a file where a directory is needed naming it, got %v", err)
		}
	})
	t.Run("refusal_precedes_every_write", func(t *testing.T) {
		// The refused path sorts after a clean one: nothing is written first.
		root, lay := fixtureRoot(t, twoSources())
		lay.Targets = lay.Targets[:1]
		if err := os.MkdirAll(filepath.Join(root, "out/c2/beta.md"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if _, err := agentemit.Update(root, lay, man); err == nil {
			t.Fatal("Update must refuse the directory at out/c2/beta.md")
		}
		if _, err := os.Stat(filepath.Join(root, "out/c2/alpha.md")); !os.IsNotExist(err) {
			t.Errorf("alpha.md was written before the refusal (stat err %v)", err)
		}
	})
	t.Run("unwritable_directory_fails_without_leftovers", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		lay.Targets = lay.Targets[:1]
		dir := filepath.Join(root, "out/c2")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		if probe, err := os.CreateTemp(dir, "probe-*"); err == nil {
			_ = probe.Close()
			t.Skip("directory permissions are not enforced for this user")
		}
		if _, err := agentemit.Update(root, lay, man); err == nil || !strings.Contains(err.Error(), "agentemit:") {
			t.Fatalf("Update must fail when the target directory is not writable, got %v", err)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Errorf("a failed write left files behind: %v", entries)
		}
	})
	t.Run("unreadable_directory_fails_the_link_check", func(t *testing.T) {
		root, lay := fixtureRoot(t, twoSources())
		lay.Targets = lay.Targets[:1]
		dir := filepath.Join(root, "out")
		if err := os.MkdirAll(filepath.Join(dir, "c2"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Chmod(dir, 0o000); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		if _, err := os.Lstat(filepath.Join(dir, "c2")); err == nil {
			t.Skip("directory permissions are not enforced for this user")
		}
		if _, err := agentemit.Update(root, lay, man); err == nil || !strings.Contains(err.Error(), "agentemit:") {
			t.Fatalf("Update must fail when a path component cannot be inspected, got %v", err)
		}
	})
	t.Run("regular_target_is_replaced_without_leftovers", func(t *testing.T) {
		// Positive control: a plain stale file is overwritten by the temp-file
		// plus rename path, and no temp file stays behind.
		root, lay := fixtureRoot(t, twoSources())
		stale := filepath.Join(root, "out/c2/alpha.md")
		writeFile(t, stale, []byte("stale\n"))
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if _, err := agentemit.Check(root, lay, man); err != nil {
			t.Errorf("Check after Update: %v", err)
		}
		if got := listNames(t, filepath.Join(root, "out/c2"), ""); strings.Join(got, ",") != "alpha.md,beta.md" {
			t.Errorf("C2 directory = %v, want only the two rendered files", got)
		}
	})
}

// TestUpdatePreservesFileMode pins that replacing an existing definition keeps
// its permission bits instead of resetting them, even under a restrictive
// process umask (which would otherwise strip them from the new file). The
// umask is process-global, so this test must not run in parallel.
func TestUpdatePreservesFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on Windows")
	}
	if !restrictUmask(t) {
		t.Skip("the platform has no process umask")
	}
	man := mustManifest(t)
	root, lay := fixtureRoot(t, twoSources())
	stale := filepath.Join(root, "out/c2/alpha.md")
	writeFile(t, stale, []byte("stale\n"))
	if err := os.Chmod(stale, 0o640); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if _, err := agentemit.Update(root, lay, man); err != nil {
		t.Fatalf("Update: %v", err)
	}
	fi, err := os.Stat(stale)
	if err != nil || fi.Mode().Perm() != 0o640 {
		t.Errorf("replaced file mode = %v, %v; want -rw-r----- (0640) preserved under umask 077", fi.Mode().Perm(), err)
	}
	// A file created fresh gets the default 0644, not a umask-reduced mode.
	fresh, err := os.Stat(filepath.Join(root, "out/c2/beta.md"))
	if err != nil || fresh.Mode().Perm() != 0o644 {
		t.Errorf("new file mode = %v, %v; want 0644", fresh.Mode().Perm(), err)
	}
	if _, err := agentemit.Check(root, lay, man); err != nil {
		t.Errorf("Check after Update: %v", err)
	}
}

// TestCheckReportsNestedOrphans pins that the orphan scan covers a whole target
// directory tree: a definition file with no agent source is reported wherever
// it sits below a full set's directory, symlinks included, and a legitimate
// generated set still reads diff=0.
func TestCheckReportsNestedOrphans(t *testing.T) {
	man := mustManifest(t)
	generated := func(t *testing.T) (string, agentemit.Layout) {
		t.Helper()
		root, lay := fixtureRoot(t, twoSources())
		if _, err := agentemit.Update(root, lay, man); err != nil {
			t.Fatalf("Update: %v", err)
		}
		return root, lay
	}
	wantOrphan := func(t *testing.T, root string, lay agentemit.Layout, wantPath string) {
		t.Helper()
		sum, err := agentemit.Check(root, lay, man)
		if err == nil || len(sum.Diffs) != 1 || !strings.HasPrefix(sum.Diffs[0], wantPath+":") ||
			!strings.Contains(sum.Diffs[0], "no agent source") {
			t.Fatalf("want exactly one orphan diff for %s, got err=%v diffs=%v", wantPath, err, sum.Diffs)
		}
	}

	t.Run("generated_set_still_clean", func(t *testing.T) {
		root, lay := generated(t)
		if err := os.MkdirAll(filepath.Join(root, "out/c2/empty/dir"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		writeFile(t, filepath.Join(root, "out/c2/notes.txt"), []byte("not a definition\n"))
		sum, err := agentemit.Check(root, lay, man)
		if err != nil || sum.String() != "specs=2 copies=4 diff=0" {
			t.Errorf("clean generated set: %q, %v", sum.String(), err)
		}
	})
	t.Run("top_level_rogue_file_reported", func(t *testing.T) {
		root, lay := generated(t)
		writeFile(t, filepath.Join(root, "out/c2/rogue.md"), []byte("rogue\n"))
		wantOrphan(t, root, lay, "out/c2/rogue.md")
	})
	t.Run("nested_rogue_file_reported", func(t *testing.T) {
		root, lay := generated(t)
		writeFile(t, filepath.Join(root, "out/c2/nested/rogue.md"), []byte("rogue\n"))
		wantOrphan(t, root, lay, "out/c2/nested/rogue.md")
	})
	t.Run("deeply_nested_rogue_toml_reported", func(t *testing.T) {
		root, lay := generated(t)
		writeFile(t, filepath.Join(root, "out/c3/a/b/rogue.toml"), []byte("x = 1\n"))
		wantOrphan(t, root, lay, "out/c3/a/b/rogue.toml")
	})
	t.Run("nested_file_with_rendered_name_reported", func(t *testing.T) {
		// A nested copy of a legitimate name is still an extra shipped file.
		root, lay := generated(t)
		writeFile(t, filepath.Join(root, "out/c2/sub/alpha.md"), []byte("copy\n"))
		wantOrphan(t, root, lay, "out/c2/sub/alpha.md")
	})
	t.Run("symlink_entry_reported", func(t *testing.T) {
		root, lay := generated(t)
		if err := os.Symlink(t.TempDir(), filepath.Join(root, "out/c2/linkdir")); err != nil {
			t.Skipf("symlinks are not available: %v", err)
		}
		wantOrphan(t, root, lay, "out/c2/linkdir")
	})
	t.Run("project_set_ignores_other_agents", func(t *testing.T) {
		root, lay := generated(t)
		writeFile(t, filepath.Join(root, "proj/c1/hand-managed.md"), []byte("hand\n"))
		writeFile(t, filepath.Join(root, "proj/c1/nested/hand-managed.md"), []byte("hand\n"))
		if sum, err := agentemit.Check(root, lay, man); err != nil {
			t.Errorf("a project set sits beside hand-managed files: %q, %v", sum.String(), err)
		}
	})
	t.Run("unreadable_nested_directory_reported", func(t *testing.T) {
		root, lay := generated(t)
		dir := filepath.Join(root, "out/c2/locked")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Chmod(dir, 0o000); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		if _, err := os.ReadDir(dir); err == nil {
			t.Skip("directory permissions are not enforced for this user")
		}
		sum, err := agentemit.Check(root, lay, man)
		if err == nil || len(sum.Diffs) != 1 || !strings.HasPrefix(sum.Diffs[0], "out/c2/locked:") ||
			!strings.Contains(sum.Diffs[0], "unreadable") {
			t.Errorf("want one unreadable diff for out/c2/locked, got err=%v diffs=%v", err, sum.Diffs)
		}
	})
}

func TestCheckReportsMissingAndDifferingFiles(t *testing.T) {
	root, lay := fixtureRoot(t, twoSources())
	man := mustManifest(t)
	if _, err := agentemit.Update(root, lay, man); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "out/c3/beta.toml")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "proj/c1/alpha.md"), []byte("changed"))
	// A project set holds hand-managed files of other agents; they are not extras.
	writeFile(t, filepath.Join(root, "proj/c1/other.md"), []byte("hand managed"))
	sum, err := agentemit.Check(root, lay, man)
	if err == nil {
		t.Fatal("Check must fail")
	}
	if len(sum.Diffs) != 2 {
		t.Fatalf("want 2 differences (missing beta.toml, differing alpha.md), got %v", sum.Diffs)
	}
	joined := strings.Join(sum.Diffs, "\n")
	if !strings.Contains(joined, "beta.toml") || !strings.Contains(joined, "alpha.md") || strings.Contains(joined, "other.md") {
		t.Errorf("differences = %q", joined)
	}
	if !strings.Contains(err.Error(), sum.Diffs[0]) {
		t.Errorf("error %q must name the first difference %q", err, sum.Diffs[0])
	}
}

func TestLayoutFailsClosed(t *testing.T) {
	man := mustManifest(t)
	if _, err := agentemit.Check(t.TempDir(), agentemit.Layout{SourceDir: "absent"}, man); err == nil {
		t.Error("a missing source directory must fail the check")
	}
	root, lay := fixtureRoot(t, twoSources())
	lay.Targets[2].Only = []string{"nosuchagent"}
	if _, err := agentemit.Check(root, lay, man); err == nil || !strings.Contains(err.Error(), "nosuchagent") {
		t.Errorf("a set naming an agent without a source must fail naming it, got %v", err)
	}
}

// TestDefaultLayoutRealTreeHasNoDrift is the in-test face of
// `make agents-emit-check`: the committed copies equal a fresh render of the
// committed sources.
func TestDefaultLayoutRealTreeHasNoDrift(t *testing.T) {
	sum, err := agentemit.Check(repoRoot, agentemit.DefaultLayout(), mustManifest(t))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if sum.Specs == 0 || sum.Copies == 0 || len(sum.Diffs) != 0 {
		t.Errorf("summary %q over the committed tree", sum.String())
	}
}
