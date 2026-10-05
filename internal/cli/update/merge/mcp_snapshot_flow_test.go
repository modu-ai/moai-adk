package merge

// mcp_snapshot_flow_test.go measures the .mcp.json snapshot across TWO
// consecutive simulated update flows (card t1029), rather than planting a
// canonical snapshot by hand.
//
// Planting a snapshot proves the merge reads one; it does not prove any flow
// ever produces one. This file drives the real sequence a flow runs —
// deploy + manifest record → backup.StageDeployedMCPSnapshot → merge and
// settle via MergeUserFilesAndSettleSnapshot — so the staging and promotion
// halves are measured, not assumed.
//
// What it does NOT reach: the cobra command bodies (runUpdate,
// runCleanReinstall, runInit) and the real embedded-template deploy. Those
// render from the embedded FS, whose values cannot be changed at test time,
// which is exactly why a template VALUE change cannot be exercised through
// them. The seam measured here is the highest one where the two renders are
// inputs rather than embedded constants.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// runFlow simulates one update flow over .mcp.json: the deploy writes render
// and records it in the manifest, the flow stages it, and the merge settles it.
// It returns the file as it stands after the flow.
func runFlow(t *testing.T, root, userCopy, render string) string {
	t.Helper()

	// Deploy: the render lands at the live path and the manifest records the
	// bytes the deployer wrote, which is what gates the staging.
	writeRel(t, root, mcpRel, render)
	m := manifest.NewManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := m.Track(mcpRel, manifest.TemplateManaged, manifest.HashBytes([]byte(render))); err != nil {
		t.Fatalf("track: %v", err)
	}
	backup.StageDeployedMCPSnapshot(root, m, nil)

	var out, warn strings.Builder
	if err := MergeUserFilesAndSettleSnapshot(
		root,
		[]FileBackup{{Path: mcpRel, Data: []byte(userCopy)}},
		&out, &warn,
	); err != nil {
		t.Fatalf("MergeUserFilesAndSettleSnapshot: %v", err)
	}
	if warn.Len() > 0 {
		t.Logf("flow warnings: %s", warn.String())
	}
	return string(readRel(t, root, mcpRel))
}

// Two consecutive flows: the first produces the canonical base, the second
// reads it and delivers the template's value change. Split by delivery kind,
// because K1 landing while K2 does not is the symptomless failure this card
// exists to close.
func TestMCPSnapshot_TwoConsecutiveFlowsDeliverValueChange(t *testing.T) {
	root := newSnapshotProject(t)

	// Flow N — the project is already carrying the prior render plus the user's
	// own server; the deploy writes that same prior render.
	afterN := runFlow(t, root, mcpUserCopy, mcpPriorRender)
	if !strings.Contains(afterN, `"mine"`) {
		t.Fatalf("flow N lost the user's own server: %s", afterN)
	}

	// The flow must have produced a canonical base; without it flow N+1 is
	// merely the hand-planted case again.
	if _, ok := backup.LoadMCPSnapshot(root); !ok {
		t.Fatalf("flow N produced no canonical .mcp.json snapshot")
	}

	// Flow N+1 — the template now ships the npx launcher and a new moai server.
	afterNPlus1 := runFlow(t, root, afterN, mcpNewRender)
	doc := decodeObject(t, []byte(afterNPlus1))

	if got := leaf(doc, "mcpServers", "context7", "command"); got != "npx" {
		t.Errorf("K2 context7.command = %v, want npx", got)
	}
	if server(doc, "moai") == nil {
		t.Errorf("K1 mcpServers.moai absent: %s", afterNPlus1)
	}
	if server(doc, "mine") == nil {
		t.Errorf("K4 mcpServers.mine lost: %s", afterNPlus1)
	}
	// K3 stays undelivered — pinned here too so the end-to-end path cannot
	// quietly diverge from the per-kind matrix. Card D3 owns the fix.
	if server(doc, "gone") == nil {
		t.Errorf("K3 mcpServers.gone was removed — card D3 assumption changed: %s", afterNPlus1)
	}
}

// A flow whose merge preserved the user's file must NOT promote its staging
// copy: the live file does not reflect the render, so promoting it would make
// the render's own keys read as user deletions on the next update.
func TestMCPSnapshot_PreservedFlowDiscardsStagingCopy(t *testing.T) {
	root := newSnapshotProject(t)
	writeRel(t, root, mcpCanonicalSnapshotRel, `{"mcpServers":{"a":{"command":"old"}}}`)

	// A user file that does not parse as JSON forces the preserve path: no base
	// can be derived, so the pre-flow content is written back wholesale.
	writeRel(t, root, mcpRel, mcpNewRender)
	m := manifest.NewManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := m.Track(mcpRel, manifest.TemplateManaged, manifest.HashBytes([]byte(mcpNewRender))); err != nil {
		t.Fatalf("track: %v", err)
	}
	backup.StageDeployedMCPSnapshot(root, m, nil)

	var out, warn strings.Builder
	if err := MergeUserFilesAndSettleSnapshot(
		root,
		[]FileBackup{{Path: mcpRel, Data: []byte("not json at all")}},
		&out, &warn,
	); err != nil {
		t.Fatalf("MergeUserFilesAndSettleSnapshot: %v", err)
	}
	if !strings.Contains(out.String(), "user content preserved") {
		t.Fatalf("expected the preserve path, got merge log:\n%s", out.String())
	}

	snapshot, ok := backup.LoadMCPSnapshot(root)
	if !ok {
		t.Fatalf("canonical snapshot vanished")
	}
	if strings.Contains(string(snapshot), "context7") {
		t.Errorf("preserved flow promoted its staging copy: %s", snapshot)
	}
}

// The settle covers both files independently: a settings.json preserve must not
// discard the .mcp.json staging copy, nor the reverse.
func TestMCPSnapshot_SettleIsPerFile(t *testing.T) {
	root := newSnapshotProject(t)

	// .mcp.json merges cleanly; .claude/settings.json takes the preserve path.
	writeRel(t, root, mcpRel, mcpNewRender)
	writeRel(t, root, settingsRel, newRender)
	m := manifest.NewManager()
	if _, err := m.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for rel, render := range map[string]string{mcpRel: mcpNewRender, settingsRel: newRender} {
		if err := m.Track(rel, manifest.TemplateManaged, manifest.HashBytes([]byte(render))); err != nil {
			t.Fatalf("track %s: %v", rel, err)
		}
	}
	backup.StageDeployedMCPSnapshot(root, m, nil)
	backup.StageDeployedSettingsSnapshot(root, m, nil)

	var out, warn strings.Builder
	if err := MergeUserFilesAndSettleSnapshot(root, []FileBackup{
		{Path: mcpRel, Data: []byte(mcpUserCopy)},
		{Path: settingsRel, Data: []byte("not json at all")},
	}, &out, &warn); err != nil {
		t.Fatalf("MergeUserFilesAndSettleSnapshot: %v", err)
	}

	if _, ok := backup.LoadMCPSnapshot(root); !ok {
		t.Errorf("the settings.json preserve discarded the .mcp.json staging copy:\n%s", out.String())
	}
	if _, ok := backup.LoadSettingsSnapshot(root); ok {
		t.Errorf("the preserved settings.json staging copy was promoted anyway")
	}
}
