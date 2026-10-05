package backup

// settings_snapshot_test.go pins the staging/canonical data model of the
// .claude/settings.json base snapshot (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001,
// card t656, plan.md M2): where the two copies live, when a copy is valid,
// that recording only happens when the deploy actually wrote the file, and
// that every failure warns once without blocking.

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

const (
	canonicalRel = ".moai/cache/template-snapshot/claude/settings.json"
	pendingRel   = ".moai/cache/template-snapshot/claude/settings.json.pending"
	liveRel      = ".claude/settings.json"
)

func readFileRel(t *testing.T, root, rel string) ([]byte, bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		t.Fatalf("read %s: %v", rel, err)
	}
	return data, true
}

func assertFileRel(t *testing.T, root, rel, want string) {
	t.Helper()
	got, ok := readFileRel(t, root, rel)
	if !ok {
		t.Errorf("%s is absent, want %s", rel, want)
		return
	}
	if !bytes.Equal(got, []byte(want)) {
		t.Errorf("%s = %s, want %s", rel, got, want)
	}
}

func assertAbsentRel(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
		t.Errorf("%s exists, want absent", rel)
	}
}

// deployRender simulates a deploy that wrote render: the file lands at the live
// path and the manifest records it as template-managed with the render hash,
// exactly as template.deployer does after a real write.
func deployRender(t *testing.T, root string, m manifest.Manager, render string) {
	t.Helper()
	writeFileRel(t, root, liveRel, render)
	if err := m.Track(liveRel, manifest.TemplateManaged, manifest.HashBytes([]byte(render))); err != nil {
		t.Fatalf("track: %v", err)
	}
}

func loadedManager(t *testing.T, root string) manifest.Manager {
	t.Helper()
	m := manifest.NewManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	return m
}

// REQ-USB-002 — the two copies sit under the shared cache root, in their own
// sub-path, never under sections/.
func TestSettingsSnapshotPaths(t *testing.T) {
	root := t.TempDir()
	if got, want := SettingsSnapshotPath(root), filepath.Join(root, filepath.FromSlash(canonicalRel)); got != want {
		t.Errorf("SettingsSnapshotPath = %s, want %s", got, want)
	}
	if got, want := SettingsSnapshotPendingPath(root), filepath.Join(root, filepath.FromSlash(pendingRel)); got != want {
		t.Errorf("SettingsSnapshotPendingPath = %s, want %s", got, want)
	}
}

// REQ-USB-004 / REQ-USB-009 — a canonical copy is usable only when it exists,
// reads, and decodes as a JSON object.
func TestLoadSettingsSnapshot_Validity(t *testing.T) {
	cases := []struct {
		name  string
		plant func(t *testing.T, root string)
		want  bool
	}{
		{"object", func(t *testing.T, root string) { writeFileRel(t, root, canonicalRel, `{"a":1}`) }, true},
		{"absent", func(*testing.T, string) {}, false},
		{"directory", func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(canonicalRel)), 0o755); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"invalid_json", func(t *testing.T, root string) { writeFileRel(t, root, canonicalRel, `{"a":`) }, false},
		{"array", func(t *testing.T, root string) { writeFileRel(t, root, canonicalRel, `[1]`) }, false},
		{"null", func(t *testing.T, root string) { writeFileRel(t, root, canonicalRel, `null`) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.plant(t, root)
			data, ok := LoadSettingsSnapshot(root)
			if ok != tc.want {
				t.Fatalf("LoadSettingsSnapshot ok = %v, want %v", ok, tc.want)
			}
			if ok && string(data) != `{"a":1}` {
				t.Errorf("LoadSettingsSnapshot data = %s", data)
			}
		})
	}
}

// REQ-USB-001 / REQ-USB-005 — staging records the deployed render and leaves
// the canonical base alone, so the same flow's merge still reads the prior one.
func TestStageDeployedSettingsSnapshot_RecordsRenderWithoutTouchingCanonical(t *testing.T) {
	root := t.TempDir()
	writeFileRel(t, root, canonicalRel, `{"marker":"prior"}`)
	m := loadedManager(t, root)
	deployRender(t, root, m, `{"a":2}`)

	var warn strings.Builder
	StageDeployedSettingsSnapshot(root, m, &warn)

	assertFileRel(t, root, pendingRel, `{"a":2}`)
	assertFileRel(t, root, canonicalRel, `{"marker":"prior"}`)
	if warn.Len() != 0 {
		t.Errorf("unexpected warning: %s", warn.String())
	}
}

// REQ-USB-016 / D7 — the manifest must prove this deploy wrote the render: a
// template-managed entry whose hash no longer matches the file (a stale entry,
// or a later rewrite) records nothing.
func TestStageDeployedSettingsSnapshot_StaleManifestHashRecordsNothing(t *testing.T) {
	root := t.TempDir()
	m := loadedManager(t, root)
	deployRender(t, root, m, `{"a":2}`)
	writeFileRel(t, root, liveRel, `{"a":2,"rewritten":true}`)

	StageDeployedSettingsSnapshot(root, m, &strings.Builder{})

	assertAbsentRel(t, root, pendingRel)
}

// AC-USB-014 — driven through the real non-force deployer (the init deployer's
// rules, internal/template/deployer.go): a fresh directory records the render;
// an existing untracked or user-modified settings.json is skipped by the
// deploy, and the skip records nothing and leaves the user's file untouched.
func TestInitSettingsSnapshot_SkippedDeployRecordsNothing(t *testing.T) {
	const render = `{"a":2,"K":1}`
	const userFile = `{"mine":true}`
	fsys := fstest.MapFS{liveRel: &fstest.MapFile{Data: []byte(render)}}

	cases := []struct {
		name          string
		seed          func(t *testing.T, root string)
		wantPending   bool
		wantLiveBytes string
	}{
		{"fresh_dir", func(*testing.T, string) {}, true, render},
		{"untracked_existing", func(t *testing.T, root string) {
			writeFileRel(t, root, liveRel, userFile)
		}, false, userFile},
		{"user_modified_existing", func(t *testing.T, root string) {
			writeFileRel(t, root, liveRel, userFile)
			writeFileRel(t, root, ".moai/manifest.json",
				`{"version":"1","files":{".claude/settings.json":{"provenance":"user_modified","template_hash":"sha256:0","deployed_hash":"sha256:0","current_hash":"sha256:0"}}}`)
		}, false, userFile},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.seed(t, root)
			m := loadedManager(t, root)
			deployer := template.NewDeployer(fsys)
			if err := deployer.Deploy(context.Background(), root, m, nil); err != nil {
				t.Fatalf("deploy: %v", err)
			}

			var warn strings.Builder
			StageDeployedSettingsSnapshot(root, m, &warn)
			SettleSettingsSnapshot(root, false, &warn) // init: no merge, flow end

			assertFileRel(t, root, liveRel, tc.wantLiveBytes)
			if tc.wantPending {
				assertFileRel(t, root, canonicalRel, render)
			} else {
				assertAbsentRel(t, root, pendingRel)
				assertAbsentRel(t, root, canonicalRel)
			}
		})
	}
}

// AC-USB-008 `helper` — a staging write failure warns exactly once with its own
// prefix, never with the sections-snapshot wording, and does not panic or block.
func TestStageDeployedSettingsSnapshot_WriteFailureDoesNotBlock(t *testing.T) {
	root := t.TempDir()
	// A regular file where the claude/ directory must go makes the write fail
	// while leaving the sections snapshot path untouched.
	writeFileRel(t, root, ".moai/cache/template-snapshot/claude", "not a directory")
	m := loadedManager(t, root)
	deployRender(t, root, m, `{"a":2}`)

	var warn strings.Builder
	StageDeployedSettingsSnapshot(root, m, &warn)

	out := warn.String()
	if n := countLinePrefix(out, SettingsSnapshotWriteFailedPrefix); n != 1 {
		t.Errorf("%q lines = %d, want 1:\n%s", SettingsSnapshotWriteFailedPrefix, n, out)
	}
	if strings.Contains(out, "Warning: template snapshot write failed:") {
		t.Errorf("settings staging failure used the sections wording:\n%s", out)
	}
	if SettingsSnapshotWriteFailedPrefix != "settings-snapshot-write-failed:" {
		t.Errorf("prefix drifted from the SPEC wording: %q", SettingsSnapshotWriteFailedPrefix)
	}
}

func countLinePrefix(out, prefix string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			n++
		}
	}
	return n
}

// REQ-USB-005 (normal end) — a flow that did not take the preserve path
// promotes its staging copy; one that did discards it and keeps the prior base.
func TestSettleSettingsSnapshot(t *testing.T) {
	t.Run("not_preserved_promotes", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, canonicalRel, `{"a":1}`)
		writeFileRel(t, root, pendingRel, `{"a":2}`)
		SettleSettingsSnapshot(root, false, &strings.Builder{})
		assertFileRel(t, root, canonicalRel, `{"a":2}`)
		assertAbsentRel(t, root, pendingRel)
	})
	t.Run("preserved_discards", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, canonicalRel, `{"a":1}`)
		writeFileRel(t, root, pendingRel, `{"a":2}`)
		SettleSettingsSnapshot(root, true, &strings.Builder{})
		assertFileRel(t, root, canonicalRel, `{"a":1}`)
		assertAbsentRel(t, root, pendingRel)
	})
	t.Run("no_pending_is_noop", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, canonicalRel, `{"a":1}`)
		SettleSettingsSnapshot(root, false, &strings.Builder{})
		assertFileRel(t, root, canonicalRel, `{"a":1}`)
	})
}

// REQ-USB-005 (leftover) — a staging copy an interrupted flow left behind is
// promoted only when the live file is still byte-identical to it.
func TestJudgeLeftoverSettingsSnapshot(t *testing.T) {
	cases := []struct {
		name          string
		live          *string
		wantCanonical string
	}{
		{"live_equals_leftover_promotes", strPtr(`{"a":2}`), `{"a":2}`},
		{"live_differs_discards", strPtr(`{"a":1}`), `{"a":1}`},
		{"live_absent_discards", nil, `{"a":1}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFileRel(t, root, canonicalRel, `{"a":1}`)
			writeFileRel(t, root, pendingRel, `{"a":2}`)
			if tc.live != nil {
				writeFileRel(t, root, liveRel, *tc.live)
			}
			JudgeLeftoverSettingsSnapshot(root, &strings.Builder{})
			assertFileRel(t, root, canonicalRel, tc.wantCanonical)
			assertAbsentRel(t, root, pendingRel)
		})
	}
	t.Run("no_leftover_is_noop", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, canonicalRel, `{"a":1}`)
		writeFileRel(t, root, liveRel, `{"a":9}`)
		JudgeLeftoverSettingsSnapshot(root, &strings.Builder{})
		assertFileRel(t, root, canonicalRel, `{"a":1}`)
	})
}

func strPtr(s string) *string { return &s }

// AC-USB-008 `promote_failure` / N3-06 — when the staging copy cannot replace
// the canonical base (a non-empty directory squats the path), the flow is not
// blocked: exactly one promote-failed line, no write-failed line, and the
// squatting directory is left as it was. Both promotion entry points share
// the failure path, so both are driven.
func TestSettingsSnapshotPromoteFailureDoesNotBlock(t *testing.T) {
	entries := map[string]func(root string, warn *strings.Builder){
		"settle_at_flow_end": func(root string, warn *strings.Builder) { SettleSettingsSnapshot(root, false, warn) },
		"leftover_judgement": func(root string, warn *strings.Builder) { JudgeLeftoverSettingsSnapshot(root, warn) },
	}
	for name, run := range entries {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writeFileRel(t, root, canonicalRel+"/squatter", "x")
			writeFileRel(t, root, pendingRel, `{"a":2}`)
			writeFileRel(t, root, liveRel, `{"a":2}`)

			var warn strings.Builder
			run(root, &warn)

			out := warn.String()
			if n := countLinePrefix(out, SettingsSnapshotPromoteFailedPrefix); n != 1 {
				t.Errorf("%q lines = %d, want 1:\n%s", SettingsSnapshotPromoteFailedPrefix, n, out)
			}
			if n := countLinePrefix(out, SettingsSnapshotWriteFailedPrefix); n != 0 {
				t.Errorf("%q lines = %d, want 0:\n%s", SettingsSnapshotWriteFailedPrefix, n, out)
			}
			if _, ok := readFileRel(t, root, canonicalRel+"/squatter"); !ok {
				t.Errorf("the squatting directory at the canonical path was disturbed")
			}
		})
	}
}

// AC-USB-009 — the settings sub-path leaves the sections snapshot alone.
//
//	cell A: identical sections input, one project also carrying both settings
//	        copies → identical sections/ trees and HasSnapshot results.
//	cell B: settings copy only, no sections snapshot → HasSnapshot is false and
//	        SaveTemplateBase falls back to SaveTemplateDefaults byte-for-byte.
func TestSectionsSnapshot_UnaffectedBySettingsSubpath(t *testing.T) {
	t.Run("cell_A_same_sections_tree", func(t *testing.T) {
		build := func(withSettings bool) (string, map[string]string, bool) {
			root := t.TempDir()
			writeFileRel(t, root, ".moai/config/sections/user.yaml", "user:\n  name: a\n")
			writeFileRel(t, root, ".moai/config/sections/nested/x.yml", "x: 1\n")
			if withSettings {
				writeFileRel(t, root, canonicalRel, `{"a":1}`)
				writeFileRel(t, root, pendingRel, `{"a":2}`)
			}
			if err := WriteSnapshot(root); err != nil {
				t.Fatalf("WriteSnapshot: %v", err)
			}
			has := HasSnapshot(root)
			dest := t.TempDir()
			if err := SaveTemplateBase(dest, root); err != nil {
				t.Fatalf("SaveTemplateBase: %v", err)
			}
			return root, treeBytes(t, filepath.Join(SnapshotDir(root), "sections")), has
		}
		_, plainTree, plainHas := build(false)
		_, withTree, withHas := build(true)
		if plainHas != withHas {
			t.Errorf("HasSnapshot differs: plain=%v with-settings=%v", plainHas, withHas)
		}
		if len(plainTree) == 0 {
			t.Fatal("sections snapshot tree is empty; the comparison would be vacuous")
		}
		if !equalTrees(plainTree, withTree) {
			t.Errorf("sections/ tree differs:\nplain: %v\nwith:  %v", plainTree, withTree)
		}
	})
	t.Run("cell_B_settings_only_is_not_a_sections_snapshot", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, canonicalRel, `{"a":1}`)
		if HasSnapshot(root) {
			t.Fatal("HasSnapshot = true with only the settings copy present")
		}
		dest, other := t.TempDir(), t.TempDir()
		if err := SaveTemplateBase(dest, root); err != nil {
			t.Fatalf("SaveTemplateBase: %v", err)
		}
		if err := SaveTemplateDefaults(other); err != nil {
			t.Fatalf("SaveTemplateDefaults: %v", err)
		}
		got, want := treeBytes(t, filepath.Join(dest, "sections")), treeBytes(t, filepath.Join(other, "sections"))
		if len(want) == 0 {
			t.Fatal("SaveTemplateDefaults produced no sections; the comparison would be vacuous")
		}
		if !equalTrees(got, want) {
			t.Errorf("SaveTemplateBase did not fall back to SaveTemplateDefaults (%d vs %d files)", len(got), len(want))
		}
	})
}

func treeBytes(t *testing.T, dir string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		tree[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return tree
}

func equalTrees(a, b map[string]string) bool {
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
