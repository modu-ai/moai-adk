package codexwiring

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func wireAt(t *testing.T, root string) {
	t.Helper()
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire: %v (%s)", err, warn.String())
	}
}

func unwireAt(t *testing.T, root string) UnwireResult {
	t.Helper()
	var warn bytes.Buffer
	res, err := Unwire(root, &warn)
	if err != nil {
		t.Fatalf("Unwire: %v (%s)", err, warn.String())
	}
	return res
}

// kept reports whether res left rel's part untouched for reason.
func kept(res UnwireResult, rel, part string, reason RefusalReason) bool {
	for _, k := range res.Kept {
		if k.Path == rel && k.Part == part && k.Reason == reason {
			return true
		}
	}
	return false
}

func removed(res UnwireResult, rel, part string) bool {
	for _, r := range res.Removed {
		if r.Path == rel && r.Part == part {
			return true
		}
	}
	return false
}

func mustAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("%s still present (err=%v)", path, err)
	}
}

// moaiHandlerIn returns the first MoAI handler registered under event in a
// rendered hooks.json, as a raw JSON object.
func moaiHandlerIn(t *testing.T, doc []byte, event string) json.RawMessage {
	t.Helper()
	var top struct {
		Hooks map[string][]struct {
			Hooks []json.RawMessage `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(doc, &top); err != nil {
		t.Fatal(err)
	}
	for _, e := range top.Hooks[event] {
		for _, h := range e.Hooks {
			var c struct {
				Command string `json:"command"`
			}
			_ = json.Unmarshal(h, &c)
			if strings.HasPrefix(c.Command, moaiHandlerPrefix) {
				return h
			}
		}
	}
	t.Fatalf("no MoAI handler under %s", event)
	return nil
}

// withoutMoAI is the structure unwire must leave behind: the parsed pre-unwire
// document minus every MoAI handler and minus the MoAI-added description, with
// entries and events left empty by that removal dropped. Array order is kept.
func withoutMoAI(t *testing.T, pre []byte) any {
	t.Helper()
	var top map[string]any
	if err := json.Unmarshal(pre, &top); err != nil {
		t.Fatal(err)
	}
	delete(top, "description")
	events := top["hooks"].(map[string]any)
	for ev, rawEntries := range events {
		var keptEntries []any
		for _, re := range rawEntries.([]any) {
			entry := re.(map[string]any)
			var hs []any
			for _, h := range entry["hooks"].([]any) {
				if strings.HasPrefix(h.(map[string]any)["command"].(string), moaiHandlerPrefix) {
					continue
				}
				hs = append(hs, h)
			}
			if len(hs) == 0 {
				continue
			}
			entry["hooks"] = hs
			keptEntries = append(keptEntries, entry)
		}
		if len(keptEntries) == 0 {
			delete(events, ev)
			continue
		}
		events[ev] = keptEntries
	}
	return top
}

// TestCodexUnwireRemovesOnlyOwnedParts covers AC-DHR-003 (REQ-DHR-001,
// REQ-DHR-005): unwire removes exactly the parts MoAI created, deletes a file
// only when MoAI created it whole and it is unchanged, preserves a re-serialized
// hooks.json by structure and config.toml by bytes.
func TestCodexUnwireRemovesOnlyOwnedParts(t *testing.T) {
	t.Run("created_file_deleted", func(t *testing.T) {
		root := t.TempDir()
		wireAt(t, root)
		res := unwireAt(t, root)
		mustAbsent(t, filepath.Join(root, HooksRelPath))
		mustAbsent(t, filepath.Join(root, ConfigRelPath))
		if !removed(res, HooksRelPath, "") || !removed(res, ConfigRelPath, "") {
			t.Fatalf("whole-file removals not reported: %+v", res.Removed)
		}
		if _, ok := manifestEntry(t, root, HooksRelPath); ok {
			t.Fatal("hooks.json record survived its file")
		}
	})

	t.Run("mixed_hooks_structure", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, HooksRelPath), `{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"./user-check.sh","timeout":5}]}]}}`)
		wireAt(t, root)
		wired := readFile(t, filepath.Join(root, HooksRelPath))
		moai := moaiHandlerIn(t, wired, "PreToolUse")
		// After wiring, the user adds a top-level key of their own and puts a
		// MoAI handler into their own entry, beside their handler.
		var top map[string]any
		if err := json.Unmarshal(wired, &top); err != nil {
			t.Fatal(err)
		}
		var moaiVal any
		_ = json.Unmarshal(moai, &moaiVal)
		top["x_user_note"] = map[string]any{"owner": "me", "list": []any{3, 1, 2}}
		pre := top["hooks"].(map[string]any)["PreToolUse"].([]any)
		userEntry := pre[0].(map[string]any)
		userEntry["hooks"] = []any{userEntry["hooks"].([]any)[0], moaiVal, map[string]any{"type": "command", "command": "./user-second.sh", "timeout": 7}}
		edited, err := json.MarshalIndent(top, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, HooksRelPath), string(edited)+"\n")

		res := unwireAt(t, root)
		var got any
		if err := json.Unmarshal(readFile(t, filepath.Join(root, HooksRelPath)), &got); err != nil {
			t.Fatal(err)
		}
		if want := withoutMoAI(t, edited); !reflect.DeepEqual(got, want) {
			t.Fatalf("structure not preserved:\n got=%v\nwant=%v", got, want)
		}
		if !removed(res, HooksRelPath, PartKeyDescription) {
			t.Fatalf("MoAI-added description not reported removed: %+v", res.Removed)
		}
	})

	t.Run("config_appended_tables_bytes", func(t *testing.T) {
		root := t.TempDir()
		const user = "model = \"gpt-5\"\n\n[mcp_servers.other]\ncommand = \"other\"\n"
		writeFile(t, filepath.Join(root, ConfigRelPath), user)
		wireAt(t, root)
		// A user table added after MoAI's appended ones puts the regions in
		// the middle of the file.
		const later = "\n[profiles.fast]\nmodel = \"gpt-5-mini\"\n"
		cfg := filepath.Join(root, ConfigRelPath)
		writeFile(t, cfg, string(readFile(t, cfg))+later)
		unwireAt(t, root)
		if got := readFile(t, cfg); !bytes.Equal(got, []byte(user+later)) {
			t.Fatalf("config bytes:\n got=%q\nwant=%q", got, user+later)
		}
	})

	t.Run("config_status_line_key_bytes", func(t *testing.T) {
		root := t.TempDir()
		const user = "[tui]\ntheme = \"dark\"\n\n[history]\npersistence = \"none\"\n"
		writeFile(t, filepath.Join(root, ConfigRelPath), user)
		wireAt(t, root)
		cfg := filepath.Join(root, ConfigRelPath)
		if !strings.Contains(string(readFile(t, cfg)), "[tui]\n"+statusLineDefaultTOML+"\ntheme") {
			t.Fatalf("fixture: status_line not inserted under the user [tui]:\n%s", readFile(t, cfg))
		}
		res := unwireAt(t, root)
		if got := readFile(t, cfg); !bytes.Equal(got, []byte(user)) {
			t.Fatalf("config bytes:\n got=%q\nwant=%q", got, user)
		}
		if !removed(res, ConfigRelPath, PartKeyStatusLine) || !removed(res, ConfigRelPath, PartKeyMCPTable) {
			t.Fatalf("removals: %+v", res.Removed)
		}
	})

	t.Run("preexisting_table_kept", func(t *testing.T) {
		root := t.TempDir()
		const user = "[mcp_servers.moai]\ncommand = \"moai\"\nargs = [\"mcp-server\"]\n"
		writeFile(t, filepath.Join(root, ConfigRelPath), user)
		wireAt(t, root)
		res := unwireAt(t, root)
		cfg := filepath.Join(root, ConfigRelPath)
		if got := readFile(t, cfg); !bytes.Equal(got, []byte(user)) {
			t.Fatalf("config bytes:\n got=%q\nwant=%q", got, user)
		}
		if !kept(res, ConfigRelPath, PartKeyMCPTable, ReasonUserOwned) {
			t.Fatalf("pre-existing table not reported user-owned: %+v", res.Kept)
		}
	})
}

// TestCodexUnwireRefusesUnownedModifiedSymlink covers AC-DHR-004
// (REQ-DHR-006, REQ-DHR-001): a matching hash is never enough to remove a
// file or part; unrecorded, unknown-origin, modified, and symlinked targets
// stay byte-identical and are reported with their reason.
func TestCodexUnwireRefusesUnownedModifiedSymlink(t *testing.T) {
	t.Run("no_provenance_canonical_content", func(t *testing.T) {
		root := t.TempDir()
		canonical, err := RenderHooks(nil)
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, HooksRelPath), string(canonical))
		before := snapshotTree(t, root)
		res := unwireAt(t, root)
		if after := snapshotTree(t, root); !sameMap(before, after) {
			t.Fatalf("tree changed:\nbefore=%v\nafter =%v", before, after)
		}
		if !kept(res, HooksRelPath, "", ReasonNoProvenance) {
			t.Fatalf("not reported no-provenance: %+v", res.Kept)
		}
	})

	t.Run("old_install_unknown_origin", func(t *testing.T) {
		root := t.TempDir()
		// An older binary wrote the config tables and the sidecar, but no
		// part records.
		cfgBody := string(EnsureStatusLine(EnsureMCPTable([]byte("model = \"gpt-5\"\n"))))
		writeFile(t, filepath.Join(root, ConfigRelPath), cfgBody)
		writeFile(t, filepath.Join(root, SidecarPath), "{}\n")
		cfg := filepath.Join(root, ConfigRelPath)
		res := unwireAt(t, root)
		if got := string(readFile(t, cfg)); got != cfgBody {
			t.Fatalf("old install config changed:\n%q", got)
		}
		if !kept(res, ConfigRelPath, "", ReasonUnknownOrigin) {
			t.Fatalf("old install not reported unknown-origin: %+v", res.Kept)
		}
		// Re-wiring records the parts as unknown; they stay unremovable on the
		// first re-wiring and on every later one.
		for round := 1; round <= 2; round++ {
			wireAt(t, root)
			res = unwireAt(t, root)
			if got := string(readFile(t, cfg)); got != cfgBody {
				t.Fatalf("round %d: re-wired old install config changed:\n%q", round, got)
			}
			for _, key := range []string{PartKeyMCPTable, PartKeyStatusLine} {
				if !kept(res, ConfigRelPath, key, ReasonUnknownOrigin) {
					t.Fatalf("round %d: re-wired %s not reported unknown-origin: %+v", round, key, res.Kept)
				}
			}
		}
	})

	t.Run("owned_but_modified", func(t *testing.T) {
		root := t.TempDir()
		wireAt(t, root)
		cfg := filepath.Join(root, ConfigRelPath)
		body := strings.Replace(string(readFile(t, cfg)), "command = \"moai\"", "command = \"/opt/moai\"", 1)
		body = strings.Replace(body, statusLineDefaultTOML, "status_line = [\"model\"]", 1)
		writeFile(t, cfg, body)
		res := unwireAt(t, root)
		if got := string(readFile(t, cfg)); got != body {
			t.Fatalf("modified config changed:\n%q", got)
		}
		for _, key := range []string{PartKeyMCPTable, PartKeyTUITable} {
			if !kept(res, ConfigRelPath, key, ReasonModified) {
				t.Fatalf("%s not reported modified: %+v", key, res.Kept)
			}
		}
	})

	t.Run("symlink_outside_project", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		wireAt(t, root)
		hooks := filepath.Join(root, HooksRelPath)
		real := filepath.Join(outside, "hooks.json")
		writeFile(t, real, string(readFile(t, hooks)))
		if err := os.Remove(hooks); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(real, hooks); err != nil {
			t.Skipf("symlink unsupported: %v", err)
		}
		beforeIn, beforeOut := snapshotTree(t, root), snapshotTree(t, outside)
		res := unwireAt(t, root)
		if !kept(res, HooksRelPath, "", ReasonSymlinkBoundary) {
			t.Fatalf("symlink not reported symlink-boundary: %+v", res.Kept)
		}
		afterIn := snapshotTree(t, root)
		if afterIn[filepath.FromSlash(HooksRelPath)] != beforeIn[filepath.FromSlash(HooksRelPath)] {
			t.Fatal("the link itself changed")
		}
		if afterOut := snapshotTree(t, outside); !sameMap(beforeOut, afterOut) {
			t.Fatalf("link destination changed:\nbefore=%v\nafter =%v", beforeOut, afterOut)
		}
	})
}

// TestCodexUnwireIsNotUndoneByRefresh: once `moai tool disable codex` ran, the
// existence-gated update refresh must not wire the project again just because
// a user-owned config.toml survived; `moai tool enable codex` re-enables.
func TestCodexUnwireIsNotUndoneByRefresh(t *testing.T) {
	root := t.TempDir()
	const user = "model = \"gpt-5\"\n"
	cfg := filepath.Join(root, ConfigRelPath)
	writeFile(t, cfg, user)
	wireAt(t, root)
	unwireAt(t, root)
	var out, warn bytes.Buffer
	if _, err := RefreshWiring(root, &out, &warn); err != nil {
		t.Fatalf("RefreshWiring: %v", err)
	}
	if got := string(readFile(t, cfg)); got != user {
		t.Fatalf("refresh re-wired a disabled project:\n%q", got)
	}
	mustAbsent(t, filepath.Join(root, HooksRelPath))
	if files := OwnedWiringFiles(root); len(files) != 0 {
		t.Fatalf("disabled project still lists owned wiring: %v", files)
	}
	wireAt(t, root)
	if !strings.Contains(string(readFile(t, cfg)), "[mcp_servers.moai]") {
		t.Fatal("enable after disable did not wire again")
	}
	e, _ := manifestEntry(t, root, ConfigRelPath)
	if p, ok := partByKey(e.Parts, PartKeyMCPTable); !ok || p.Origin != manifest.OriginCreated {
		t.Fatalf("re-enabled table recorded %+v", p)
	}
}

// TestOwnedWiringFiles pins which files the update path names as needing
// `moai tool disable codex`: files holding a created part, every present
// wiring file of an older install (sidecar, no records), and nothing for a
// project whose wiring is entirely user-owned.
func TestOwnedWiringFiles(t *testing.T) {
	t.Run("wired", func(t *testing.T) {
		root := t.TempDir()
		wireAt(t, root)
		if got := OwnedWiringFiles(root); !reflect.DeepEqual(got, []string{HooksRelPath, ConfigRelPath}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("older_install", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ConfigRelPath), string(EnsureMCPTable(nil)))
		writeFile(t, filepath.Join(root, SidecarPath), "{}\n")
		if got := OwnedWiringFiles(root); !reflect.DeepEqual(got, []string{ConfigRelPath}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("user_only", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ConfigRelPath), "model = \"gpt-5\"\n")
		if got := OwnedWiringFiles(root); len(got) != 0 {
			t.Fatalf("got %v", got)
		}
	})
}

// TestCodexUnwireHooksModifiedAndUnparseable: a MoAI handler or description
// the user edited stays, reported modified, while untouched MoAI handlers
// beside it go; an unparseable hooks.json with a record is left untouched.
func TestCodexUnwireHooksModifiedAndUnparseable(t *testing.T) {
	t.Run("modified_handler_and_description", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, HooksRelPath), `{"hooks":{}}`)
		wireAt(t, root)
		hooks := filepath.Join(root, HooksRelPath)
		body := strings.Replace(string(readFile(t, hooks)), `"timeout": 3`, `"timeout": 2`, 1)
		writeFile(t, hooks, body)
		res := unwireAt(t, root)
		got := string(readFile(t, hooks))
		if !strings.Contains(got, `"timeout": 2`) {
			t.Fatalf("edited handler was removed:\n%s", got)
		}
		if strings.Contains(got, "moai hook pre-tool --harness codex") {
			t.Fatalf("unchanged MoAI handler survived:\n%s", got)
		}
		modified := false
		for _, k := range res.Kept {
			if k.Path == HooksRelPath && k.Reason == ReasonModified && strings.Contains(k.Part, "session-end") {
				modified = true
			}
		}
		if !modified {
			t.Fatalf("edited handler not reported modified: %+v", res.Kept)
		}
	})
	t.Run("modified_description", func(t *testing.T) {
		root := t.TempDir()
		wireAt(t, root)
		hooks := filepath.Join(root, HooksRelPath)
		writeFile(t, hooks, strings.Replace(string(readFile(t, hooks)), "MoAI-managed hook layer.", "My hook layer.", 1))
		res := unwireAt(t, root)
		if !kept(res, HooksRelPath, PartKeyDescription, ReasonModified) {
			t.Fatalf("edited description not reported modified: %+v", res.Kept)
		}
		if !strings.Contains(string(readFile(t, hooks)), "My hook layer.") {
			t.Fatal("edited description was removed")
		}
	})
	t.Run("unparseable_with_record", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, HooksRelPath), `{"hooks":{}}`)
		wireAt(t, root)
		hooks := filepath.Join(root, HooksRelPath)
		writeFile(t, hooks, "{ broken")
		res := unwireAt(t, root)
		if got := string(readFile(t, hooks)); got != "{ broken" {
			t.Fatalf("unparseable file changed: %q", got)
		}
		if !kept(res, HooksRelPath, "", ReasonUnparseable) {
			t.Fatalf("not reported unparseable: %+v", res.Kept)
		}
	})
}
