// drift_test.go — AC-009 (b): the drift check reports a byte difference, a mode
// difference, a missing file and an extra file in the committed payload, and
// never writes.
package pluginemit_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/pluginemit"
)

// driftPublication is a small emission: a manifest, a script and a command.
func driftPublication() *pluginemit.Publication {
	files := map[string][]byte{
		pluginemit.ClaudePluginPath:           []byte("{\"name\": \"x\"}\n"),
		pluginRoot + "/skills/s/scripts/t.sh": []byte("#!/bin/sh\n"),
		pluginRoot + "/skills/s/SKILL.md":     []byte("skill\n"),
		pluginRoot + "/commands/c.md":         []byte("command\n"),
	}
	modes := map[string]fs.FileMode{}
	for p := range files {
		modes[p] = 0o644
		if strings.HasSuffix(p, ".sh") {
			modes[p] = 0o755
		}
	}
	return &pluginemit.Publication{Files: files, Modes: modes}
}

// materialise writes pub into a fresh directory with plain file operations and
// returns it. It does not use pluginemit.Write, so a defect there cannot hide a
// defect in Drift (TestWriteMaterialisesTree judges Write on its own).
func materialise(t *testing.T, pub *pluginemit.Publication) string {
	t.Helper()
	root := t.TempDir()
	for p, data := range pub.Files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, data, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(full, pub.Modes[p]); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestWriteMaterialisesTree: Write creates every file at its mode, repairs the
// mode of a file that already exists, and removes a file the emission does not
// carry inside the generated roots — while leaving unrelated files alone.
func TestWriteMaterialisesTree(t *testing.T) {
	pub := driftPublication()
	root := t.TempDir()

	stale := filepath.Join(root, filepath.FromSlash(pluginRoot+"/commands/stale.md"))
	script := filepath.Join(root, filepath.FromSlash(pluginRoot+"/skills/s/scripts/t.sh"))
	unrelated := filepath.Join(root, "docs", "keep.md")
	for _, p := range []string{stale, script, unrelated} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("old\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := pluginemit.Write(pub, root); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if diffs := driftOf(t, pub, root); len(diffs) != 0 {
		t.Errorf("a tree just written still drifts: %v", diffs)
	}
	for p, want := range pub.Files {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil || string(got) != string(want) {
			t.Errorf("%s: read back %q (%v), want %q", p, got, err, want)
		}
		if runtime.GOOS == "windows" {
			continue
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			t.Errorf("%s: %v", p, err)
			continue
		}
		if info.Mode().Perm() != pub.Modes[p] {
			t.Errorf("%s: mode %v, want %v", p, info.Mode().Perm(), pub.Modes[p])
		}
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("a stale file inside the generated root survived a regeneration")
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Errorf("a file outside the generated roots was touched: %v", err)
	}
}

func driftOf(t *testing.T, pub *pluginemit.Publication, root string) []pluginemit.Difference {
	t.Helper()
	diffs, err := pluginemit.Drift(pub, root)
	if err != nil {
		t.Fatalf("Drift: %v", err)
	}
	return diffs
}

// wantOne asserts that diffs holds exactly the one expected difference.
func wantOne(t *testing.T, diffs []pluginemit.Difference, kind, path string) {
	t.Helper()
	if len(diffs) != 1 || diffs[0].Kind != kind || diffs[0].Path != path {
		t.Errorf("Drift = %v, want exactly {%s %s}", diffs, kind, path)
	}
}

// snapshot records the bytes and mode of every file under root.
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = info.Mode().String() + "|" + string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDriftDetectsMutatedArtifact(t *testing.T) {
	scriptPath := pluginRoot + "/skills/s/scripts/t.sh"
	manifestPath := pluginemit.ClaudePluginPath

	t.Run("clean-tree-has-no-drift", func(t *testing.T) {
		pub := driftPublication()
		if diffs := driftOf(t, pub, materialise(t, pub)); len(diffs) != 0 {
			t.Errorf("a freshly written tree drifts: %v", diffs)
		}
	})

	t.Run("flipped-byte", func(t *testing.T) {
		pub := driftPublication()
		root := materialise(t, pub)
		target := filepath.Join(root, filepath.FromSlash(pluginRoot+"/skills/s/SKILL.md"))
		if err := os.WriteFile(target, []byte("skill!\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		wantOne(t, driftOf(t, pub, root), pluginemit.DriftBytes, pluginRoot+"/skills/s/SKILL.md")
	})

	t.Run("mode-flipped", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows has no execute bit; the mode comparison is not made there")
		}
		pub := driftPublication()
		root := materialise(t, pub)
		if err := os.Chmod(filepath.Join(root, filepath.FromSlash(scriptPath)), 0o644); err != nil {
			t.Fatal(err)
		}
		wantOne(t, driftOf(t, pub, root), pluginemit.DriftMode, scriptPath)

		root = materialise(t, pub)
		if err := os.Chmod(filepath.Join(root, filepath.FromSlash(manifestPath)), 0o755); err != nil {
			t.Fatal(err)
		}
		wantOne(t, driftOf(t, pub, root), pluginemit.DriftMode, manifestPath)
	})

	t.Run("deleted-file", func(t *testing.T) {
		pub := driftPublication()
		root := materialise(t, pub)
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(pluginRoot+"/commands/c.md"))); err != nil {
			t.Fatal(err)
		}
		wantOne(t, driftOf(t, pub, root), pluginemit.DriftMissing, pluginRoot+"/commands/c.md")
	})

	t.Run("extra-file", func(t *testing.T) {
		pub := driftPublication()
		root := materialise(t, pub)
		extra := pluginRoot + "/commands/nested/extra.md"
		full := filepath.Join(root, filepath.FromSlash(extra))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		wantOne(t, driftOf(t, pub, root), pluginemit.DriftExtra, extra)
	})

	t.Run("committed-set-unchanged", func(t *testing.T) {
		pub := driftPublication()
		root := materialise(t, pub)
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(pluginRoot+"/skills/s/SKILL.md")), []byte("hand edit\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(pluginRoot+"/commands/c.md"))); err != nil {
			t.Fatal(err)
		}
		extra := filepath.Join(root, filepath.FromSlash(pluginRoot+"/commands/extra.md"))
		if err := os.WriteFile(extra, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		before := snapshot(t, root)
		diffs := driftOf(t, pub, root)
		if len(diffs) < 3 {
			t.Fatalf("expected the three mutations to be reported, got %v", diffs)
		}
		after := snapshot(t, root)
		var keys []string
		for k := range before {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if len(before) != len(after) {
			t.Errorf("the check changed the file count: %d -> %d", len(before), len(after))
		}
		for _, k := range keys {
			if before[k] != after[k] {
				t.Errorf("the check changed %s; a read-only check must not regenerate", k)
			}
		}
	})
}
