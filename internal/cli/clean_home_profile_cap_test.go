package cli

// SPEC-CAPACITY-RECLAIM-001 AC-CAP-011: the profile cap warns and deletes
// nothing on card A's paths. The loader reads the HOME-tier
// state.profile_max_bytes with NO compiled fallback (an absent key yields no
// warning — the compiled DefaultProfileMaxBytes is no longer a candidate
// source in scanHomeCleanable either); the advisory messages name the
// profile and the key; and neither the scan nor the advisory mutates the
// fixture home.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/hygiene"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// snapshotHomeFiles returns rel-path -> size for every file under root, for
// the no-deletion comparison.
func snapshotHomeFiles(t *testing.T, root string) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		out[filepath.ToSlash(rel)] = info.Size()
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

// seedProfileFixture writes a fixture HOME whose single profile carries one
// projects/ entry of sizeBytes (default mtime = now, so no age-based
// projects candidate can exist) and, when withKey, a HOME-tier state.yaml
// carrying state.profile_max_bytes: 1048576.
func seedProfileFixture(t *testing.T, withKey bool, sizeBytes int64) string {
	t.Helper()
	home := t.TempDir()
	moai := filepath.Join(home, ".moai")
	projects := filepath.Join(moai, "claude-profiles", "prof-a", "projects", "session-fixture")
	if err := os.MkdirAll(projects, 0o755); err != nil {
		t.Fatalf("mkdir fixture projects: %v", err)
	}
	blob := make([]byte, sizeBytes)
	for i := range blob {
		blob[i] = 'p'
	}
	if err := os.WriteFile(filepath.Join(projects, "chunk.bin"), blob, 0o644); err != nil {
		t.Fatalf("seed profile payload: %v", err)
	}
	if withKey {
		sections := filepath.Join(moai, "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatalf("mkdir sections: %v", err)
		}
		stateYAML := "state:\n    profile_max_bytes: 1048576\n"
		if err := os.WriteFile(filepath.Join(sections, "state.yaml"), []byte(stateYAML), 0o644); err != nil {
			t.Fatalf("seed state.yaml: %v", err)
		}
	} else {
		// state.yaml present but the key absent — the sharper key-absent
		// case: the loader must parse the file and find the key missing
		// (a compiled fallback would surface exactly here).
		sections := filepath.Join(moai, "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatalf("mkdir sections: %v", err)
		}
		stateYAML := "state:\n    home_retention_days: 30\n"
		if err := os.WriteFile(filepath.Join(sections, "state.yaml"), []byte(stateYAML), 0o644); err != nil {
			t.Fatalf("seed state.yaml without the key: %v", err)
		}
	}
	return home
}

// TestProfileCapWarnsWithoutDeleting covers AC-CAP-011's two fixtures.
func TestProfileCapWarnsWithoutDeleting(t *testing.T) {
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	miB := int64(1024 * 1024)

	t.Run("cap set: warning names profile and key; no projects candidate; nothing deleted", func(t *testing.T) {
		home := seedProfileFixture(t, true, 6*miB)
		moai := filepath.Join(home, ".moai")
		t.Setenv("HOME", home)
		t.Setenv("MOAI_HOME", "")

		// The loader reads the fixture key with no compiled fallback. The
		// empty project dir exercises the HOME-tier-first ordering: the home
		// key wins before any project fallback is consulted.
		capBytes, ok, err := hygiene.LoadProfileCapMaxBytes("")
		if err != nil {
			t.Fatalf("LoadProfileCapMaxBytes: %v", err)
		}
		if !ok || capBytes != 1048576 {
			t.Fatalf("loaded cap = (%d, %v), want (1048576, true)", capBytes, ok)
		}

		// The advisory names the profile and the key.
		profilesDir, err := paths.ProfilesDir()
		if err != nil {
			t.Fatalf("ProfilesDir: %v", err)
		}
		msgs, err := hygiene.ProfileCapAdvisories(profilesDir, capBytes, 5*time.Second)
		if err != nil {
			t.Fatalf("ProfileCapAdvisories: %v", err)
		}
		if len(msgs) != 1 {
			t.Fatalf("advisories = %v, want exactly one", msgs)
		}
		if !strings.Contains(msgs[0], "prof-a") || !strings.Contains(msgs[0], "state.profile_max_bytes") {
			t.Fatalf("advisory does not name the profile and the key: %q", msgs[0])
		}

		// The scan produces no projects/ candidate on account of the cap
		// (the fixture entry is young, so any projects candidate would be
		// cap-driven), and the fixture home is byte-count-identical after.
		before := snapshotHomeFiles(t, moai)
		candidates := scanHomeCleanable(moai, 30, 3, "v9.9.9", now)
		for _, c := range candidates {
			if c.Category == "projects" {
				t.Fatalf("cap produced a projects deletion candidate: %+v", c)
			}
		}
		after := snapshotHomeFiles(t, moai)
		if len(after) != len(before) {
			t.Fatalf("fixture home changed across the scan: before %v after %v", before, after)
		}
		for rel, size := range before {
			if after[rel] != size {
				t.Fatalf("fixture file changed across the scan: %s %d -> %d", rel, size, after[rel])
			}
		}
	})

	t.Run("key absent: no warning and no projects candidate", func(t *testing.T) {
		home := seedProfileFixture(t, false, 6*miB)
		moai := filepath.Join(home, ".moai")
		t.Setenv("HOME", home)
		t.Setenv("MOAI_HOME", "")

		// No compiled fallback: an absent key yields no warning. The
		// emission gate is the caller's `if !ok { return }` (mirrored from
		// runProfileCapAdvisory) — the advisory function itself is only ever
		// invoked with a cap that was actually loaded. The end-to-end
		// absent-key-no-warning behavior is the hook test's clause-3
		// fixture-2 arm, which drives the real pass. The empty project dir
		// ends the fallback chain after the key-absent home tier.
		capBytes, ok, err := hygiene.LoadProfileCapMaxBytes("")
		if err != nil {
			t.Fatalf("LoadProfileCapMaxBytes: %v", err)
		}
		if ok {
			t.Fatalf("key-absent state.yaml resolved to (%d, true); the loader must carry no compiled fallback", capBytes)
		}

		// The file-absent case reads the same: no fallback either.
		homeNoFile := t.TempDir()
		t.Setenv("HOME", homeNoFile)
		t.Setenv("MOAI_HOME", "")
		fileAbsentBytes, fileAbsentOk, fileAbsentErr := hygiene.LoadProfileCapMaxBytes(t.TempDir())
		if fileAbsentErr != nil {
			t.Fatalf("LoadProfileCapMaxBytes (no state.yaml): %v", fileAbsentErr)
		}
		if fileAbsentOk {
			t.Fatalf("missing state.yaml resolved to (%d, true); the loader must carry no compiled fallback", fileAbsentBytes)
		}

		candidates := scanHomeCleanable(moai, 30, 3, "v9.9.9", now)
		for _, c := range candidates {
			if c.Category == "projects" {
				t.Fatalf("projects candidate without any cap in play: %+v", c)
			}
		}
	})

	t.Run("project-tier fallback: home key-absent state.yaml falls through to the project sections", func(t *testing.T) {
		home := seedProfileFixture(t, false, 6*miB) // home state.yaml present WITHOUT the key
		moai := filepath.Join(home, ".moai")
		project := t.TempDir()
		sections := filepath.Join(project, defs.MoAIDir, defs.SectionsSubdir)
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatalf("mkdir project sections: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sections, "state.yaml"),
			[]byte("state:\n    profile_max_bytes: 1048576\n"), 0o644); err != nil {
			t.Fatalf("seed project state.yaml: %v", err)
		}
		t.Setenv("HOME", home)
		t.Setenv("MOAI_HOME", "")

		// The project tier carries the shipped default the home tier omits.
		capBytes, ok, err := hygiene.LoadProfileCapMaxBytes(project)
		if err != nil {
			t.Fatalf("LoadProfileCapMaxBytes (project fallback): %v", err)
		}
		if !ok || capBytes != 1048576 {
			t.Fatalf("project fallback cap = (%d, %v), want (1048576, true)", capBytes, ok)
		}

		// The advisory still names the profile and the key.
		profilesDir, err := paths.ProfilesDir()
		if err != nil {
			t.Fatalf("ProfilesDir: %v", err)
		}
		msgs, err := hygiene.ProfileCapAdvisories(profilesDir, capBytes, 5*time.Second)
		if err != nil {
			t.Fatalf("ProfileCapAdvisories: %v", err)
		}
		if len(msgs) != 1 || !strings.Contains(msgs[0], "prof-a") || !strings.Contains(msgs[0], "state.profile_max_bytes") {
			t.Fatalf("project-fallback advisories = %v, want one naming prof-a and the key", msgs)
		}

		// And the scan still produces no cap-driven projects candidate.
		for _, c := range scanHomeCleanable(moai, 30, 3, "v9.9.9", now) {
			if c.Category == "projects" {
				t.Fatalf("cap produced a projects deletion candidate under the project-tier cap: %+v", c)
			}
		}
	})

	t.Run("unreadable subdirectory: over-cap profile still warns with partial-size semantics", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("directory permission bits are not enforced on Windows")
		}
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}
		home := seedProfileFixture(t, true, 6*miB) // 6 MiB profile above the 1 MiB cap
		moai := filepath.Join(home, ".moai")
		t.Setenv("HOME", home)
		t.Setenv("MOAI_HOME", "")

		// A chmod-000 directory inside the profile: the sync-audit P2-2
		// runtime reproduction. The walker must skip it (counted), not
		// abort the profile's measurement — the advisory still fires.
		locked := filepath.Join(moai, "claude-profiles", "prof-a", "projects", "locked-dir")
		if err := os.MkdirAll(locked, 0o755); err != nil {
			t.Fatalf("mkdir locked dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(locked, "secret.bin"), []byte("unreadable"), 0o600); err != nil {
			t.Fatalf("seed locked dir: %v", err)
		}
		if err := os.Chmod(locked, 0o000); err != nil {
			t.Fatalf("chmod locked dir: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

		capBytes, ok, err := hygiene.LoadProfileCapMaxBytes("")
		if err != nil {
			t.Fatalf("LoadProfileCapMaxBytes: %v", err)
		}
		if !ok {
			t.Fatalf("cap not loaded")
		}
		profilesDir, err := paths.ProfilesDir()
		if err != nil {
			t.Fatalf("ProfilesDir: %v", err)
		}
		msgs, err := hygiene.ProfileCapAdvisories(profilesDir, capBytes, 5*time.Second)
		if err != nil {
			t.Fatalf("ProfileCapAdvisories: %v", err)
		}
		if len(msgs) != 1 {
			t.Fatalf("advisories = %v, want exactly one despite the unreadable subdirectory", msgs)
		}
		for _, want := range []string{"prof-a", "state.profile_max_bytes", "measurement partial"} {
			if !strings.Contains(msgs[0], want) {
				t.Fatalf("advisory %q does not name %q", msgs[0], want)
			}
		}
	})
}
