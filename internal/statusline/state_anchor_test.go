package statusline

// state_anchor_test.go — SPEC-STATE-ANCHOR-001 (card t510, GH #1694).
//
// Committed RED-first tests for the state-anchor repair. C0 (this file's
// first test) reproduces the verdict §B-2 probe as a permanent regression
// guard: the session telemetry record must anchor to the project root, never
// to a directory the session merely visited — one stray .moai per visited
// directory was the reported defect (GH #1694: 226 stray .moai dirs).

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestContextUsageAnchorsToProjectDir — AC-SA-001 (B1, the #1694 primary
// offender). Given a render whose stdin carries BOTH workspace.current_dir
// (the directory the session cd'd to) and workspace.project_dir (the project
// root), the telemetry record must land under the project anchor and the
// visited directory must carry no .moai at all (REQ-SA-002 chain step 1,
// REQ-SA-007).
//
// RED at base 0b1e27877: resolveProjectDir anchored to current_dir, so the
// record landed in the visited directory while the project dir stayed clean —
// the exact inverse of what this test asserts (verdict .moai/reports/t510/
// verdict.md §B-2 probe observed the same shape).
func TestContextUsageAnchorsToProjectDir(t *testing.T) {
	visited := t.TempDir() // workspace.current_dir — where the session cd'd to
	project := t.TempDir() // workspace.project_dir — the project root

	// Isolate from ambient GLM env (same rationale as
	// TestBuild_WritesContextUsageWithSessionID): dev machines running
	// `moai glm` export ANTHROPIC_DEFAULT_*_MODEL, which would override the
	// synthetic 256K window below.
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	in := StdinData{
		SessionID: "sess-anchor-001",
		Workspace: &WorkspaceInfo{CurrentDir: visited, ProjectDir: project},
		ContextWindow: &ContextWindowInfo{
			ContextWindowSize: 256000,
			UsedPercentage:    new(90.0),
		},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	// Hermetic builder — same shape as TestBuild_WritesContextUsageWithSessionID:
	// no git/usage providers, so no network or repository is touched.
	b := &defaultBuilder{
		renderer: NewRenderer("default", true, nil),
		mode:     ModeDefault,
	}
	if _, err := b.Build(context.Background(), bytes.NewReader(raw)); err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	recordPath := filepath.Join(project, ".moai", "state", "context-usage", "sess-anchor-001.json")
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("telemetry record not anchored to project_dir (want %s): %v", recordPath, err)
	}
	var rec SessionTelemetryRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("unparseable record at project anchor: %v", err)
	}
	if rec.SessionID != "sess-anchor-001" {
		t.Errorf("session_id = %q, want sess-anchor-001", rec.SessionID)
	}

	if _, err := os.Stat(filepath.Join(visited, ".moai")); !os.IsNotExist(err) {
		t.Errorf("visited dir polluted with .moai (the GH #1694 mechanism): stat err = %v", err)
	}
}

// TestNoProjectNoState — AC-SA-005 (REQ-SA-003). Given a render with no
// project_dir, no original_cwd, and a current_dir outside any git
// repository, no state directory is created anywhere and the render
// completes normally — no error, no panic, and (mutant 4) no log noise or
// render failure from the skip path. "No project, no state."
func TestNoProjectNoState(t *testing.T) {
	visited := t.TempDir() // non-git temporary directory, no project anywhere

	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	in := StdinData{
		SessionID: "sess-nostate",
		Workspace: &WorkspaceInfo{CurrentDir: visited},
		ContextWindow: &ContextWindowInfo{
			ContextWindowSize: 256000,
			UsedPercentage:    new(90.0),
		},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	b := &defaultBuilder{
		renderer: NewRenderer("default", true, nil),
		mode:     ModeDefault,
	}
	out, err := b.Build(context.Background(), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("render must complete normally without a project, got error: %v", err)
	}
	if out == "" {
		t.Errorf("render produced no output — the skip path must not break the render")
	}

	if _, err := os.Stat(filepath.Join(visited, ".moai")); !os.IsNotExist(err) {
		t.Errorf("no-project render created .moai under the visited dir (must skip, REQ-SA-003): stat err = %v", err)
	}
}

// TestDisplaySegmentUnchangedByAnchorRepair — AC-SA-006 (REQ-SA-004, plan D2).
// The display derivation is INVARIANT: extractProjectDirectory keeps its
// project_dir-first priority and the repair must not touch it. The golden
// input is the DIVERGENCE input — project_dir(B) ≠ current_dir(A) — the only
// input where the display source and the state anchor actually split, so a
// repair that (wrongly) rewires display to the state anchor fails here.
func TestDisplaySegmentUnchangedByAnchorRepair(t *testing.T) {
	visited := t.TempDir() // workspace.current_dir (A)
	project := t.TempDir() // workspace.project_dir (B)

	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	in := StdinData{
		SessionID: "sess-display-golden",
		Workspace: &WorkspaceInfo{CurrentDir: visited, ProjectDir: project},
		ContextWindow: &ContextWindowInfo{
			ContextWindowSize: 256000,
			UsedPercentage:    new(90.0),
		},
	}

	// The extraction point collectAll feeds the renderer: project_dir stays
	// priority 1 (builder.go:237 → extractProjectDirectory).
	b := &defaultBuilder{
		renderer: NewRenderer("default", true, nil),
		mode:     ModeDefault,
	}
	data := b.collectAll(context.Background(), &in)
	if want := filepath.Base(project); data.Directory != want {
		t.Errorf("display derivation changed: Directory = %q, want %q (project_dir basename)", data.Directory, want)
	}

	// The rendered directory segment carries the project basename, never the
	// visited one (renderer.go: 📁 <Directory>).
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := b.Build(context.Background(), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if !strings.Contains(out, "📁 "+filepath.Base(project)) {
		t.Errorf("rendered directory segment must show project basename %q, got:\n%s", filepath.Base(project), out)
	}
	if strings.Contains(out, "📁 "+filepath.Base(visited)) {
		t.Errorf("rendered directory segment shows the visited basename %q — display must not follow the session's current dir:\n%s", filepath.Base(visited), out)
	}
}
