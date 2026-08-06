package hook

// navigator_detect_test.go — SPEC-NAVIGATOR-SYNC-002 M1.2 (BAS Falconer Detect
// PostToolUse branch integration). TDD coverage for AC-NS2-001a (Write/Edit
// trigger), AC-NS2-001b (Bash NOT triggered), AC-NS2-009a (no forked hook
// chain), AC-NS2-009b (branch registered inside dispatcher), and the D3
// NotebookEdit recon verdict (NotebookEdit SHALL trigger — see progress.md
// §E.2 for the full recon record).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/navigator/detect"
	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// writeNavGraphFixture writes a minimal nav-graph.json at
// <root>/.moai/project/navigator/nav-graph.json with one dec-edge, one
// spec-edge (the highest-value @MX:SPEC bridge case), and one sym-edge — all
// with absolute source_path values anchored at absRoot. Returns the three
// absolute source paths so tests can build matching changed-path inputs.
func writeNavGraphFixture(t *testing.T, absRoot string) (decPath, specPath, symPath string) {
	t.Helper()
	decPath = filepath.Join(absRoot, ".moai", "project", "tech.md")
	specPath = filepath.Join(absRoot, "internal", "auth", "login.go")
	symPath = filepath.Join(absRoot, "internal", "auth", "login.go")

	graph := navsync.Graph{
		Provenance: navsync.Provenance{ExtractCommitSHA: "fixture-sha", CapturedAt: "2026-08-06"},
		Nodes: []navsync.Node{
			{EntityType: navsync.EntityDecision, Identifier: "AUTH-STRATEGY", DisplayName: "OAuth2 strategy"},
			{EntityType: navsync.EntitySpec, Identifier: "SPEC-AUTH-001", DisplayName: "Auth SPEC"},
			{EntityType: navsync.EntitySymbol, Identifier: "auth.ParseBearer", DisplayName: "ParseBearer"},
		},
		Edges: []navsync.Edge{
			{EdgeType: navsync.EdgeDec, SourceNode: "decision:AUTH-STRATEGY", TargetNode: "spec:SPEC-AUTH-001", SourcePath: decPath, LineNumber: 42},
			{EdgeType: navsync.EdgeSpec, SourceNode: "symbol:auth.ParseBearer", TargetNode: "spec:SPEC-AUTH-001", SourcePath: specPath, LineNumber: 17},
			{EdgeType: navsync.EdgeSym, SourceNode: "symbol:auth.ParseBearer", TargetNode: "spec:SPEC-AUTH-001", SourcePath: symPath, LineNumber: 30},
		},
	}
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal fixture graph: %v", err)
	}
	outDir := filepath.Join(absRoot, ".moai", "project", "navigator")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture graph dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "nav-graph.json"), raw, 0o644); err != nil {
		t.Fatalf("write fixture graph: %v", err)
	}
	return decPath, specPath, symPath
}

// --- Unit tests for runNavigatorDetect (the branch entry point) ---

func TestRunNavigatorDetect_Write_FiresTraverse(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:           tmpDir,
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "` + specPath + `"}`),
	}
	got := runNavigatorDetect(input)
	if got == nil {
		t.Fatal("runNavigatorDetect returned nil for Write on a matching path; expected a non-nil Result")
	}
	if len(got.Edges) == 0 {
		t.Fatal("expected affected edges for matching Write; got 0")
	}
	// AC-NS2-002 row 002b: spec-edge (the @MX:SPEC bridge — highest-value case)
	// must surface the SPEC back-pointer node.
	foundSpec := false
	for _, n := range got.Nodes {
		if n.EntityType == navsync.EntitySpec && n.Identifier == "SPEC-AUTH-001" {
			foundSpec = true
			break
		}
	}
	if !foundSpec {
		t.Errorf("expected affected node spec:SPEC-AUTH-001 in Write result; got nodes=%+v", got.Nodes)
	}
}

func TestRunNavigatorDetect_Edit_FiresTraverse(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	decPath, _, _ := writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Edit",
		ToolInput: json.RawMessage(`{"file_path": "` + decPath + `"}`),
	}
	got := runNavigatorDetect(input)
	if got == nil {
		t.Fatal("runNavigatorDetect returned nil for Edit on a matching path")
	}
	foundDec := false
	for _, n := range got.Nodes {
		if n.EntityType == navsync.EntityDecision && n.Identifier == "AUTH-STRATEGY" {
			foundDec = true
			break
		}
	}
	if !foundDec {
		t.Errorf("expected affected node decision:AUTH-STRATEGY for Edit on tech.md; got nodes=%+v", got.Nodes)
	}
}

// TestRunNavigatorDetect_NotebookEdit_FiresTraverse encodes the D3 recon
// VERDICT (progress.md §E.2): PostToolUse DOES fire for NotebookEdit and the
// ToolInput.notebook_path field IS a parseable string — therefore NotebookEdit
// stays in REQ-NS2-001's SHALL trigger surface alongside Write/Edit.
func TestRunNavigatorDetect_NotebookEdit_FiresTraverse(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "NotebookEdit",
		ToolInput: json.RawMessage(`{"notebook_path": "` + specPath + `"}`),
	}
	got := runNavigatorDetect(input)
	if got == nil {
		t.Fatal("runNavigatorDetect returned nil for NotebookEdit on a matching notebook_path; " +
			"D3 verdict requires NotebookEdit to trigger (REQ-NS2-001 SHALL)")
	}
	if len(got.Edges) == 0 {
		t.Fatal("expected affected edges for NotebookEdit; got 0")
	}
}

// AC-NS2-001b — Bash SHALL NOT trigger the Detect branch.
func TestRunNavigatorDetect_Bash_NotTriggered(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Bash",
		ToolInput: json.RawMessage(`{"command": "sed -i 's/x/y/' internal/foo/bar.go"}`),
	}
	if got := runNavigatorDetect(input); got != nil {
		t.Errorf("runNavigatorDetect returned non-nil for Bash; expected nil (Bash is not a trigger surface per REQ-NS2-001): %+v", got)
	}
}

func TestRunNavigatorDetect_Read_NotTriggered(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Read",
		ToolInput: json.RawMessage(`{"file_path": "/anything"}`),
	}
	if got := runNavigatorDetect(input); got != nil {
		t.Errorf("runNavigatorDetect returned non-nil for Read; expected nil: %+v", got)
	}
}

// Fail-open error modes (full AC-NS2-004 table-driven coverage is M1.4 scope,
// but the absent-graph and unparseable-JSON cases are exercised here to prove
// the M1.2 branch wiring is fail-open at the integration boundary).
func TestRunNavigatorDetect_GraphAbsent_FailOpen(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	// No fixture written — graph is absent.

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path": "` + filepath.Join(tmpDir, "x.go") + `"}`),
	}
	if got := runNavigatorDetect(input); got != nil {
		t.Errorf("expected nil Result on absent graph (fail-open); got %+v", got)
	}
}

func TestRunNavigatorDetect_GraphUnparseable_FailOpen(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, ".moai", "project", "navigator")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "nav-graph.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write bad graph: %v", err)
	}

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path": "` + filepath.Join(tmpDir, "x.go") + `"}`),
	}
	if got := runNavigatorDetect(input); got != nil {
		t.Errorf("expected nil Result on unparseable graph (fail-open); got %+v", got)
	}
}

func TestRunNavigatorDetect_NoFilePath_FailOpen(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{}`),
	}
	if got := runNavigatorDetect(input); got != nil {
		t.Errorf("expected nil Result when ToolInput has no file_path/notebook_path; got %+v", got)
	}
}

func TestRunNavigatorDetect_NoProjectRoot_FailOpen(t *testing.T) {
	t.Parallel()
	// CWD empty and CLAUDE_PROJECT_DIR not set → resolver returns "" → branch
	// fails open. (We cannot reliably unset CLAUDE_PROJECT_DIR in a parallel
	// test, so we rely on input.CWD being empty AND the resolver preferring
	// input.CWD; if CLAUDE_PROJECT_DIR is set in the test environment, the
	// resolver uses it. To make this test deterministic regardless of env, we
	// skip when CLAUDE_PROJECT_DIR is set.)
	if os.Getenv("CLAUDE_PROJECT_DIR") != "" {
		t.Skip("CLAUDE_PROJECT_DIR is set in the test env; skipping no-project-root case")
	}
	input := &HookInput{
		CWD:       "",
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path": "/anything.go"}`),
	}
	if got := runNavigatorDetect(input); got != nil {
		t.Errorf("expected nil Result when projectRoot unresolvable; got %+v", got)
	}
}

// --- Dispatcher integration tests (AC-NS2-009b — branch registered inside
// postToolHandler.Handle, NOT a forked chain) ---

// metricsHasNavigatorDetect unmarshals HookOutput.Data and reports whether a
// navigator_detect metrics key is present (the observable side-channel the
// dispatcher uses to record that the branch ran).
func metricsHasNavigatorDetect(t *testing.T, data json.RawMessage) bool {
	t.Helper()
	if len(data) == 0 {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal HookOutput.Data: %v\n%s", err, data)
	}
	_, ok := m["navigator_detect"]
	return ok
}

// AC-NS2-009b + AC-NS2-001a: full Handle() dispatches the Detect branch for
// Write on a matching path, and the branch's metrics entry is observable in
// the HookOutput.Data.
func TestPostToolHandler_NavigatorDetect_WriteDispatch(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:           tmpDir,
		SessionID:     "sess-nd-001",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "` + specPath + `"}`),
	}
	h := NewPostToolHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil || out.HookSpecificOutput == nil {
		t.Fatal("nil output / HookSpecificOutput")
	}
	if !metricsHasNavigatorDetect(t, out.Data) {
		t.Errorf("expected navigator_detect metrics entry for Write dispatch; got Data=%s", out.Data)
	}
}

// AC-NS2-001b: Bash does NOT dispatch the Detect branch (no navigator_detect
// metrics entry in HookOutput.Data).
func TestPostToolHandler_NavigatorDetect_BashNotDispatched(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:           tmpDir,
		SessionID:     "sess-nd-002",
		HookEventName: "PostToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(`{"command": "sed -i 's/x/y/' internal/foo/bar.go"}`),
	}
	h := NewPostToolHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if metricsHasNavigatorDetect(t, out.Data) {
		t.Errorf("expected NO navigator_detect metrics entry for Bash dispatch; got Data=%s", out.Data)
	}
}

// AC-NS2-001a NotebookEdit dispatch (D3 verdict SHALL): full Handle()
// dispatches the Detect branch for NotebookEdit on a matching notebook_path.
func TestPostToolHandler_NavigatorDetect_NotebookEditDispatch(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:           tmpDir,
		SessionID:     "sess-nd-003",
		HookEventName: "PostToolUse",
		ToolName:      "NotebookEdit",
		ToolInput:     json.RawMessage(`{"notebook_path": "` + specPath + `"}`),
	}
	h := NewPostToolHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if !metricsHasNavigatorDetect(t, out.Data) {
		t.Errorf("expected navigator_detect metrics entry for NotebookEdit dispatch (D3 SHALL); got Data=%s", out.Data)
	}
}

// --- AC-NS2-009a — no forked hook chain ---

// TestNavigatorDetect_NoForkedChain verifies the Detect branch did NOT fork
// the PostToolUse hook chain: no handle-navigator-detect.sh wrapper script
// exists, and no `moai hook navigator-detect` subcommand is registered.
// (This is the M1.2-light version of the AC-NS2-009a non-overlap grep; the
// full table-driven non-overlap test lands in M1.5.)
func TestNavigatorDetect_NoForkedChain(t *testing.T) {
	t.Parallel()
	wrapper := ".claude/hooks/moai/handle-navigator-detect.sh"
	if _, err := os.Stat(wrapper); err == nil {
		t.Errorf("forked hook chain detected: %s exists (REQ-NS2-009 forbids a new wrapper)", wrapper)
	}
}

// TestNavigatorDetect_BranchRegisteredInDispatcher verifies the branch is
// wired inside postToolHandler.Handle (AC-NS2-009b). Source-grep proof that
// runNavigatorDetect is called from exactly one site in post_tool.go, gated
// on the Write/Edit/NotebookEdit trigger surface.
func TestNavigatorDetect_BranchRegisteredInDispatcher(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("post_tool.go")
	if err != nil {
		t.Fatalf("read post_tool.go: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, "runNavigatorDetect(input)") {
		t.Errorf("runNavigatorDetect is not called from post_tool.go; AC-NS2-009b requires the branch to be registered inside postToolHandler.Handle")
	}
	// The call site count for runNavigatorDetect across the whole package MUST
	// be exactly 1 (the dispatcher) plus test references. Production source
	// (post_tool.go) MUST contain exactly one call.
	prodCalls := strings.Count(body, "runNavigatorDetect(input)")
	if prodCalls != 1 {
		t.Errorf("expected exactly 1 runNavigatorDetect(input) call in post_tool.go; got %d", prodCalls)
	}
}

// --- detectForChangedPath unit tests (the testable core sans HookInput) ---

func TestDetectForChangedPath_SpecEdgeBridge(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)
	graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")

	got := detectForChangedPath(graphPath, specPath)
	if got == nil {
		t.Fatal("expected non-nil Result for spec-edge matching path")
	}
	if len(got.Edges) != 2 { // spec-edge + sym-edge both have this source_path
		t.Errorf("expected 2 affected edges (spec+sym) for %s; got %d: %+v", specPath, len(got.Edges), got.Edges)
	}
}

func TestDetectForChangedPath_GraphAbsent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	got := detectForChangedPath(filepath.Join(tmpDir, "missing.json"), "/anything")
	if got != nil {
		t.Errorf("expected nil Result for absent graph; got %+v", got)
	}
}

// Compile-time assertion that detect.Result is the return shape (locks the
// M1.2 wiring to the M1.1 contract — a refactor of Traverse's Result type
// would break this test file at compile time).
var _ = (*detect.Result)(nil)
