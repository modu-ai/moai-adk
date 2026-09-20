package merge

// mcp_snapshot_base_test.go pins the base selection of the per-file 3-way
// merge for .mcp.json (card t1029, generalising SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001).
//
// Every assertion here runs through the PRODUCTION path (MergeUserFiles /
// MergeUserFilesAndSettleSnapshot), never by feeding the engine a base
// directly: the design measurement that motivated this card reached only the
// engine, and MergeUserFilesWithOutcome has earlier branches (templateManaged,
// derived) that could short-circuit before the snapshot is ever consulted.
//
// The assertions are split BY DELIVERY KIND rather than per file, because a
// partial success here is a symptomless failure — K1 landing while K2 does not
// makes the update look successful while the value stays stale:
//
//	K1 key addition   — the template introduces mcpServers.moai
//	K2 value change   — the template rewrites context7.command the user never touched
//	K3 key removal    — the template retires mcpServers.gone
//	K4 user key       — the user's own mcpServers.mine survives
//
// The snapshot paths are spelled out literally, not taken from the production
// constants, so the tests pin the on-disk location rather than whatever the
// constant happens to say.

import (
	"bytes"
	"strings"
	"testing"
)

const (
	mcpRel                  = ".mcp.json"
	mcpCanonicalSnapshotRel = ".moai/cache/template-snapshot/mcp.json"
)

// The three renders model the real precedent f770c0211, which changed
// context7's launcher from /bin/bash to npx. The user never touched that key,
// carries the prior render verbatim, and adds one server of their own.
const (
	mcpPriorRender = `{"mcpServers":{` +
		`"context7":{"command":"/bin/bash","args":["-l","-c","exec npx -y @upstash/context7-mcp@latest"]},` +
		`"gone":{"command":"retired"}}}`

	mcpNewRender = `{"mcpServers":{` +
		`"context7":{"command":"npx","args":["-y","@upstash/context7-mcp@latest"]},` +
		`"moai":{"command":"moai","args":["mcp-server"]}}}`

	mcpUserCopy = `{"mcpServers":{` +
		`"context7":{"command":"/bin/bash","args":["-l","-c","exec npx -y @upstash/context7-mcp@latest"]},` +
		`"gone":{"command":"retired"},` +
		`"mine":{"command":"my-own-server"}}}`
)

// mergeMCPJSON deploys render at .mcp.json and merges user into it through the
// production path, returning the decoded result and the merge log.
func mergeMCPJSON(t *testing.T, root, user, render string) (map[string]any, string) {
	t.Helper()
	writeRel(t, root, mcpRel, render)
	var out strings.Builder
	if err := MergeUserFiles(root, []FileBackup{{Path: mcpRel, Data: []byte(user)}}, &out); err != nil {
		t.Fatalf("MergeUserFiles: %v", err)
	}
	return decodeObject(t, readRel(t, root, mcpRel)), out.String()
}

// server returns the named entry under mcpServers, or nil.
func server(doc map[string]any, name string) map[string]any {
	entry, _ := leaf(doc, "mcpServers", name).(map[string]any)
	return entry
}

// --- K1 key addition ---------------------------------------------------------

// A key the new render introduces reaches the project. This already worked
// before the snapshot base existed; it is asserted separately so a regression
// here cannot hide behind K2 passing.
func TestMergeUserFiles_MCPSnapshotBase_K1KeyAdditionDelivered(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, mcpCanonicalSnapshotRel, mcpPriorRender)
	doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

	if server(doc, "moai") == nil {
		t.Errorf("mcpServers.moai absent — template key addition not delivered: %v", doc)
	}
}

// --- K2 value change (the defect this card closes) ---------------------------

// A template value change to a key the user never touched reaches the project
// when the canonical snapshot is the base. This is the cell that fails today.
func TestMergeUserFiles_MCPSnapshotBase_K2ValueChangeDelivered(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, mcpCanonicalSnapshotRel, mcpPriorRender)
	doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

	if got := leaf(doc, "mcpServers", "context7", "command"); got != "npx" {
		t.Errorf("context7.command = %v, want npx — template value change not delivered", got)
	}
}

// Control for K2: with no canonical snapshot the derived base cannot see the
// template's value change, so the user's value stands. This is the pre-snapshot
// behaviour and it must stay intact (REQ-USB-009 analogue).
func TestMergeUserFiles_MCPSnapshotBase_K2ControlNoSnapshotKeepsUserValue(t *testing.T) {
	root := newSnapshotProject(t)
	doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

	if got := leaf(doc, "mcpServers", "context7", "command"); got != "/bin/bash" {
		t.Errorf("context7.command = %v, want /bin/bash (derived-base fallback)", got)
	}
}

// --- K3 key removal (characterization: STILL NOT DELIVERED) ------------------

// A key the template retired stays in the user's file — with AND without the
// snapshot base. This is NOT fixed by this card and must not be claimed as
// fixed: the same cell fails identically on the already-shipped
// .claude/settings.json snapshot path (t1023 E7), which makes it a pre-existing
// defect of its own, tracked as card D3.
//
// The test is a characterization test: it pins today's wrong-but-known
// behaviour so a future change to it is visible rather than silent.
func TestMergeUserFiles_MCPSnapshotBase_K3KeyRemovalStillNotDelivered(t *testing.T) {
	t.Run("with_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		writeRel(t, root, mcpCanonicalSnapshotRel, mcpPriorRender)
		doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

		if server(doc, "gone") == nil {
			t.Errorf("mcpServers.gone was removed — key removal now delivered; card D3 assumption changed: %v", doc)
		}
	})
	t.Run("control_no_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

		if server(doc, "gone") == nil {
			t.Errorf("mcpServers.gone was removed without a snapshot: %v", doc)
		}
	})
}

// --- K4 user key preservation ------------------------------------------------

// The user's own server survives the merge. This is the invariant the whole
// mechanism exists to protect, so it is asserted on both base paths.
func TestMergeUserFiles_MCPSnapshotBase_K4UserKeyPreserved(t *testing.T) {
	t.Run("with_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		writeRel(t, root, mcpCanonicalSnapshotRel, mcpPriorRender)
		doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

		if server(doc, "mine") == nil {
			t.Errorf("mcpServers.mine lost: %v", doc)
		}
	})
	t.Run("control_no_canonical", func(t *testing.T) {
		root := newSnapshotProject(t)
		doc, _ := mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)

		if server(doc, "mine") == nil {
			t.Errorf("mcpServers.mine lost: %v", doc)
		}
	})
}

// --- fallback byte-identity ---------------------------------------------------

// With no usable canonical snapshot the result is byte-identical to today's,
// on every shape of unusable copy. This is the REQ-USB-009 property t656
// preserved for settings.json, held for .mcp.json.
func TestMergeUserFiles_MCPSnapshotFallbackMatchesDerivedBase(t *testing.T) {
	reference := func(t *testing.T) []byte {
		root := newSnapshotProject(t)
		mergeMCPJSON(t, root, mcpUserCopy, mcpNewRender)
		return readRel(t, root, mcpRel)
	}(t)

	cases := []struct {
		name  string
		plant func(t *testing.T, root string)
	}{
		{"absent", func(*testing.T, string) {}},
		{"invalid_json", func(t *testing.T, root string) { writeRel(t, root, mcpCanonicalSnapshotRel, `{"a":`) }},
		{"json_array", func(t *testing.T, root string) { writeRel(t, root, mcpCanonicalSnapshotRel, `[1]`) }},
		{"json_null", func(t *testing.T, root string) { writeRel(t, root, mcpCanonicalSnapshotRel, `null`) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newSnapshotProject(t)
			tc.plant(t, root)
			writeRel(t, root, mcpRel, mcpNewRender)
			var out strings.Builder
			if err := MergeUserFiles(root, []FileBackup{{Path: mcpRel, Data: []byte(mcpUserCopy)}}, &out); err != nil {
				t.Fatalf("MergeUserFiles returned %v, want nil", err)
			}
			if got := readRel(t, root, mcpRel); !bytes.Equal(got, reference) {
				t.Errorf("fallback result differs from the no-snapshot result:\ngot:  %s\nwant: %s", got, reference)
			}
		})
	}
}

// --- cross-file scoping -------------------------------------------------------

// The .mcp.json snapshot must not be read as the settings.json base, nor the
// reverse. Planting each file's snapshot with the OTHER file's content would
// deliver a visibly wrong value if the lookup were keyed loosely.
func TestMergeUserFiles_MCPSnapshotDoesNotLeakIntoSettings(t *testing.T) {
	root := newSnapshotProject(t)
	// Only the .mcp.json snapshot exists; settings.json must fall back to the
	// derived base and keep the user's value.
	writeRel(t, root, mcpCanonicalSnapshotRel, priorRender)
	doc, _ := mergeSettings(t, root, priorRender, newRender)

	if got := leaf(doc, "statusLine", "command"); got != "old" {
		t.Errorf("statusLine.command = %v, want old — the .mcp.json snapshot leaked into settings.json", got)
	}
}
