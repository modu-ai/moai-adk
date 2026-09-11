package merge

// settings_snapshot_base_test.go pins the base selection of the per-file
// 3-way merge for .claude/settings.json (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001,
// card t656): a valid canonical snapshot of the previously deployed render is
// the base, and anything else falls back to the derived base unchanged.
//
// The snapshot paths are spelled out literally here, not taken from the
// production constants, so the tests pin the on-disk location the SPEC names
// (REQ-USB-002) rather than whatever the constant happens to say.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

const (
	settingsRel          = ".claude/settings.json"
	canonicalSnapshotRel = ".moai/cache/template-snapshot/claude/settings.json"
	conflictLine         = ".claude/settings.json merged with conflicts (user version preferred)"
)

// newSnapshotProject creates a t.TempDir() project carrying an empty manifest,
// which MergeUserFiles requires.
func newSnapshotProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeRel(t, root, ".moai/manifest.json", `{"version":"1","files":{}}`)
	return root
}

func writeRel(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func readRel(t *testing.T, root, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return data
}

func decodeObject(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}
	return doc
}

// mergeSettings deploys render at .claude/settings.json and merges user into it.
func mergeSettings(t *testing.T, root, user, render string) (map[string]any, string) {
	t.Helper()
	writeRel(t, root, settingsRel, render)
	var out strings.Builder
	if err := MergeUserFiles(root, []FileBackup{{Path: settingsRel, Data: []byte(user)}}, &out); err != nil {
		t.Fatalf("MergeUserFiles: %v", err)
	}
	return decodeObject(t, readRel(t, root, settingsRel)), out.String()
}

func leaf(doc map[string]any, path ...string) any {
	var cur any = doc
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[key]
	}
	return cur
}

const (
	priorRender = `{"statusLine":{"command":"old"},"env":{"PATH":"/old"},"permissions":{"deny":["A"]}}`
	newRender   = `{"statusLine":{"command":"new"},"env":{"PATH":"/new"},"permissions":{"deny":["A","B"]}}`
)

// AC-USB-001 — a template value change on a shared leaf the user never touched
// reaches the project when the canonical snapshot is the base.
func TestMergeUserFiles_SnapshotBaseDeliversTemplateValueChange(t *testing.T) {
	t.Run("with_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		writeRel(t, root, canonicalSnapshotRel, priorRender)
		doc, _ := mergeSettings(t, root, priorRender, newRender)

		if got := leaf(doc, "statusLine", "command"); got != "new" {
			t.Errorf("statusLine.command = %v, want new", got)
		}
		if got := leaf(doc, "env", "PATH"); got != "/new" {
			t.Errorf("env.PATH = %v, want /new", got)
		}
		if got := leaf(doc, "permissions", "deny"); !reflect.DeepEqual(got, []any{"A", "B"}) {
			t.Errorf("permissions.deny = %v, want [A B]", got)
		}
	})
	// Control: without a canonical snapshot the derived base cannot see the
	// template's value change, so the user's copy stands (today's behaviour).
	t.Run("control_no_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		doc, _ := mergeSettings(t, root, priorRender, newRender)

		if got := leaf(doc, "statusLine", "command"); got != "old" {
			t.Errorf("statusLine.command = %v, want old", got)
		}
		if got := leaf(doc, "env", "PATH"); got != "/old" {
			t.Errorf("env.PATH = %v, want /old", got)
		}
		if got := leaf(doc, "permissions", "deny"); !reflect.DeepEqual(got, []any{"A"}) {
			t.Errorf("permissions.deny = %v, want [A]", got)
		}
	})
}

// AC-USB-002 — a key the user edited and the template did not change keeps the
// user's value without a conflict.
func TestMergeUserFiles_SnapshotBaseKeepsUserEdit(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, canonicalSnapshotRel, `{"model":"sonnet"}`)
	doc, out := mergeSettings(t, root, `{"model":"opus"}`, `{"model":"sonnet"}`)

	if got := doc["model"]; got != "opus" {
		t.Errorf("model = %v, want opus", got)
	}
	if strings.Contains(out, "merged with conflicts") {
		t.Errorf("unexpected conflict report:\n%s", out)
	}
}

// AC-USB-003 — both sides changed the same leaf: the user's value stays and the
// file is reported as merged with conflicts, exactly once.
func TestMergeUserFiles_SnapshotBaseBothChangedReportsConflict(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, canonicalSnapshotRel, `{"model":"sonnet"}`)
	doc, out := mergeSettings(t, root, `{"model":"opus"}`, `{"model":"haiku"}`)

	if got := doc["model"]; got != "opus" {
		t.Errorf("model = %v, want opus", got)
	}
	if n := strings.Count(out, conflictLine); n != 1 {
		t.Errorf("conflict line count = %d, want 1:\n%s", n, out)
	}
}

// AC-USB-004 — an absent or unusable canonical snapshot falls back to the
// derived base and yields byte-for-byte the pre-SPEC result without failing.
func TestMergeUserFiles_SnapshotFallbackMatchesDerivedBase(t *testing.T) {
	reference := func(t *testing.T) []byte {
		root := newSnapshotProject(t)
		mergeSettings(t, root, priorRender, newRender)
		return readRel(t, root, settingsRel)
	}(t)

	cases := []struct {
		name  string
		plant func(t *testing.T, root string)
	}{
		{"absent", func(*testing.T, string) {}},
		{"unreadable_dir", func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(canonicalSnapshotRel)), 0o755); err != nil {
				t.Fatalf("mkdir canonical-as-dir: %v", err)
			}
		}},
		{"invalid_json", func(t *testing.T, root string) { writeRel(t, root, canonicalSnapshotRel, `{"a":`) }},
		{"json_array", func(t *testing.T, root string) { writeRel(t, root, canonicalSnapshotRel, `[1]`) }},
		{"json_null", func(t *testing.T, root string) { writeRel(t, root, canonicalSnapshotRel, `null`) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newSnapshotProject(t)
			tc.plant(t, root)
			writeRel(t, root, settingsRel, newRender)
			var out strings.Builder
			if err := MergeUserFiles(root, []FileBackup{{Path: settingsRel, Data: []byte(priorRender)}}, &out); err != nil {
				t.Fatalf("MergeUserFiles returned %v, want nil", err)
			}
			if got := readRel(t, root, settingsRel); !bytes.Equal(got, reference) {
				t.Errorf("fallback result differs from the no-snapshot result:\ngot:  %s\nwant: %s", got, reference)
			}
		})
	}
}

// AC-USB-010 (new cell) — a key only the new render carries is still added
// when the canonical snapshot is the base.
func TestMergeUserFiles_SnapshotBaseAddsNewTemplateKey(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, canonicalSnapshotRel, `{"a":1}`)
	doc, _ := mergeSettings(t, root, `{"a":1}`, `{"a":1,"hooks":{"New":[1]}}`)

	if got := leaf(doc, "hooks", "New"); !reflect.DeepEqual(got, []any{float64(1)}) {
		t.Errorf("hooks.New = %v, want [1]", got)
	}
}

// AC-USB-011 — the canonical snapshot is the base for .claude/settings.json
// only; .mcp.json keeps the derived base even when a look-alike snapshot sits
// in the same cache root.
func TestMergeUserFiles_SnapshotBaseScopedToSettingsJSON(t *testing.T) {
	mergeMCP := func(t *testing.T, plant bool) []byte {
		root := newSnapshotProject(t)
		if plant {
			writeRel(t, root, canonicalSnapshotRel, priorRender)
			writeRel(t, root, ".moai/cache/template-snapshot/claude/mcp.json", priorRender)
			writeRel(t, root, ".moai/cache/template-snapshot/.mcp.json", priorRender)
		}
		writeRel(t, root, ".mcp.json", newRender)
		var out strings.Builder
		if err := MergeUserFiles(root, []FileBackup{{Path: ".mcp.json", Data: []byte(priorRender)}}, &out); err != nil {
			t.Fatalf("MergeUserFiles: %v", err)
		}
		return readRel(t, root, ".mcp.json")
	}
	with, without := mergeMCP(t, true), mergeMCP(t, false)
	if !bytes.Equal(with, without) {
		t.Errorf(".mcp.json result changed by snapshots:\nwith:    %s\nwithout: %s", with, without)
	}
	if got := leaf(decodeObject(t, with), "statusLine", "command"); got != "old" {
		t.Errorf(".mcp.json statusLine.command = %v, want old (derived base)", got)
	}
}

// AC-USB-012 — with a canonical snapshot, a template key the user deleted
// stays deleted.
func TestMergeUserFiles_SnapshotBaseHonorsUserKeyDeletion(t *testing.T) {
	const render = `{"statusLine":{"command":"x"},"model":"sonnet"}`
	t.Run("with_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		writeRel(t, root, canonicalSnapshotRel, render)
		doc, _ := mergeSettings(t, root, `{"model":"sonnet"}`, render)
		if _, ok := doc["statusLine"]; ok {
			t.Errorf("user-deleted statusLine came back: %v", doc)
		}
	})
	t.Run("control_no_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		doc, _ := mergeSettings(t, root, `{"model":"sonnet"}`, render)
		if _, ok := doc["statusLine"]; !ok {
			t.Errorf("derived base no longer re-adds statusLine: %v", doc)
		}
	})
}

// AC-USB-013 — arrays compare as one leaf (known limitation B1): a user
// addition plus a template addition to the same array is a conflict and the
// user's array stands.
func TestMergeUserFiles_SnapshotBaseArrayIsWholeLeaf(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, canonicalSnapshotRel, `{"permissions":{"allow":["Read"]}}`)
	doc, out := mergeSettings(t, root,
		`{"permissions":{"allow":["Read","Bash(ls)"]}}`,
		`{"permissions":{"allow":["Read","Grep"]}}`)

	if got := leaf(doc, "permissions", "allow"); !reflect.DeepEqual(got, []any{"Read", "Bash(ls)"}) {
		t.Errorf("permissions.allow = %v, want [Read Bash(ls)]", got)
	}
	if n := strings.Count(out, "merged with conflicts"); n != 1 {
		t.Errorf("conflict line count = %d, want 1:\n%s", n, out)
	}
}
