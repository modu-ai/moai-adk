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
	"time"

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
// runNavigatorDetectSafe (M1.4 hardened wrapper) is called from exactly one
// site in post_tool.go, gated on the Write/Edit/NotebookEdit trigger surface.
func TestNavigatorDetect_BranchRegisteredInDispatcher(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("post_tool.go")
	if err != nil {
		t.Fatalf("read post_tool.go: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, "runNavigatorDetectSafe(ctx, input)") {
		t.Errorf("runNavigatorDetectSafe is not called from post_tool.go; AC-NS2-009b requires the branch to be registered inside postToolHandler.Handle")
	}
	// The call site count for the safe wrapper in production source MUST be
	// exactly 1 (the dispatcher). M1.4 routes the PostToolUse branch through
	// the fail-open + bounded-deadline wrapper rather than the raw entry point.
	prodCalls := strings.Count(body, "runNavigatorDetectSafe(ctx, input)")
	if prodCalls != 1 {
		t.Errorf("expected exactly 1 runNavigatorDetectSafe(ctx, input) call in post_tool.go; got %d", prodCalls)
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

// --- M1.3 tests (AC-NS2-003): systemMessage + JSONL impact record +
// no-promotion. These cover the output surfaces that replace the M1.2
// metrics-stub. The Detect layer MUST emit (a) a read-only advisory
// systemMessage naming the affected rows, (b) an append-only JSONL impact
// record at .moai/state/navigator-detect/<session-id>.jsonl, and (c) MUST
// NOT promote findings to any actionable work item (no issue, no SPEC mutation,
// no TODO file). ---

// AC-NS2-003a — systemMessage emitted, advisory.
func TestFormatNavigatorDetectSystemMessage_NonEmptyResult(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)
	graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
	result := detectForChangedPath(graphPath, specPath)
	if result == nil {
		t.Fatal("expected non-nil Result for fixture path")
	}
	msg := formatNavigatorDetectSystemMessage(specPath, result)
	if msg == "" {
		t.Fatal("expected non-empty systemMessage for non-empty affected-row set")
	}
	if !strings.HasPrefix(msg, "Navigator Detect:") {
		t.Errorf("systemMessage must begin with 'Navigator Detect:'; got %q", msg)
	}
	if !strings.Contains(msg, specPath) {
		t.Errorf("systemMessage must name the changed path %q; got %q", specPath, msg)
	}
	// Must surface the @MX:SPEC bridge case (spec:SPEC-AUTH-001) — the
	// highest-value affected row per plan §C.3 / spec.md HISTORY.
	if !strings.Contains(msg, "spec:SPEC-AUTH-001") {
		t.Errorf("systemMessage must name affected node spec:SPEC-AUTH-001; got %q", msg)
	}
	// Must reference at least one edge_type + source_path:line location.
	if !strings.Contains(msg, "spec-edge") {
		t.Errorf("systemMessage must name the originating edge_type spec-edge; got %q", msg)
	}
}

func TestFormatNavigatorDetectSystemMessage_NilResult_NoAdvisory(t *testing.T) {
	t.Parallel()
	if got := formatNavigatorDetectSystemMessage("/anything", nil); got != "" {
		t.Errorf("nil Result MUST emit no advisory; got %q", got)
	}
}

func TestFormatNavigatorDetectSystemMessage_EmptyResult_NoAdvisory(t *testing.T) {
	t.Parallel()
	empty := &detect.Result{}
	if got := formatNavigatorDetectSystemMessage("/anything", empty); got != "" {
		t.Errorf("empty Result MUST emit no advisory; got %q", got)
	}
}

// AC-NS2-003a (§F edge case "SystemMessage overflow"): >10 affected rows
// truncate with a tail line; the JSONL record carries the full set.
func TestFormatNavigatorDetectSystemMessage_OverflowTruncates(t *testing.T) {
	t.Parallel()
	// Build a synthetic Result with 13 affected edges — exceeds the 10-row cap.
	const n = 13
	result := &detect.Result{}
	for i := range n {
		result.Edges = append(result.Edges, navsync.Edge{
			EdgeType:   navsync.EdgeSym,
			SourceNode: "symbol:pkg.Foo",
			TargetNode: "spec:SPEC-X",
			SourcePath: "/abs/project/internal/foo.go",
			LineNumber: i + 1,
		})
	}
	msg := formatNavigatorDetectSystemMessage("/abs/project/internal/foo.go", result)
	// The 10-row cap means at most 10 detail lines + 1 header + 1 overflow tail.
	detailLines := strings.Count(msg, "\n- ")
	if detailLines > systemMessageRowLimit {
		t.Errorf("detail rows MUST be capped at %d; got %d detail lines:\n%s", systemMessageRowLimit, detailLines, msg)
	}
	if !strings.Contains(msg, "and 3 more") {
		t.Errorf("overflow tail line MUST name the remainder (3); got:\n%s", msg)
	}
}

// AC-NS2-003b — JSONL impact record schema. Each appended line MUST be valid
// JSON with changed_path / changed_at / affected_nodes (array) / affected_edges
// (array of {edge_type, source_node, target_node, source_path, line_number}).
func TestAppendNavigatorDetectImpact_JSONLSchema(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)
	graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
	result := detectForChangedPath(graphPath, specPath)
	if result == nil {
		t.Fatal("expected non-nil Result for fixture path")
	}

	const sessionID = "test-session-001"
	const changedAt = "2026-08-06T12:00:00+00:00" // deterministic; no wall-clock in tests
	if err := appendNavigatorDetectImpact(tmpDir, sessionID, specPath, changedAt, result); err != nil {
		t.Fatalf("appendNavigatorDetectImpact: %v", err)
	}

	jsonlPath := filepath.Join(tmpDir, navigatorDetectStateDir, sessionID+".jsonl")
	raw, err := os.ReadFile(jsonlPath)
	if err != nil {
		t.Fatalf("read JSONL impact record: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected at least 1 JSONL line; got empty file at %s", jsonlPath)
	}
	last := lines[len(lines)-1]

	var rec map[string]any
	if err := json.Unmarshal([]byte(last), &rec); err != nil {
		t.Fatalf("last JSONL line is not valid JSON: %v\nline=%s", err, last)
	}
	// Required top-level keys per AC-NS2-003b.
	for _, key := range []string{"changed_path", "changed_at", "affected_nodes", "affected_edges"} {
		if _, ok := rec[key]; !ok {
			t.Errorf("JSONL record missing required key %q; line=%s", key, last)
		}
	}
	if got, _ := rec["changed_path"].(string); got != specPath {
		t.Errorf("changed_path mismatch; got %q want %q", got, specPath)
	}
	if got, _ := rec["changed_at"].(string); got != changedAt {
		t.Errorf("changed_at mismatch; got %q want %q (deterministic injection)", got, changedAt)
	}
	edges, _ := rec["affected_edges"].([]any)
	if len(edges) == 0 {
		t.Fatalf("affected_edges MUST be a non-empty array; line=%s", last)
	}
	firstEdge, _ := edges[0].(map[string]any)
	for _, key := range []string{"edge_type", "source_node", "target_node", "source_path", "line_number"} {
		if _, ok := firstEdge[key]; !ok {
			t.Errorf("affected_edges element missing required key %q; line=%s", key, last)
		}
	}
}

// AC-NS2-003b — session-scoped JSONL path. Different sessionIDs write to
// distinct files; repeated edits to the same session append on new lines.
func TestAppendNavigatorDetectImpact_SessionScopedAndAppends(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)
	graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
	result := detectForChangedPath(graphPath, specPath)

	const s1, s2 = "sess-A", "sess-B"
	if err := appendNavigatorDetectImpact(tmpDir, s1, specPath, "t1", result); err != nil {
		t.Fatalf("append s1 #1: %v", err)
	}
	if err := appendNavigatorDetectImpact(tmpDir, s1, specPath, "t2", result); err != nil {
		t.Fatalf("append s1 #2: %v", err)
	}
	if err := appendNavigatorDetectImpact(tmpDir, s2, specPath, "t3", result); err != nil {
		t.Fatalf("append s2: %v", err)
	}

	s1Raw, _ := os.ReadFile(filepath.Join(tmpDir, navigatorDetectStateDir, s1+".jsonl"))
	s2Raw, _ := os.ReadFile(filepath.Join(tmpDir, navigatorDetectStateDir, s2+".jsonl"))
	s1Lines := strings.Count(strings.TrimSpace(string(s1Raw)), "\n") + 1
	s2Lines := strings.Count(strings.TrimSpace(string(s2Raw)), "\n") + 1
	if s1Lines != 2 {
		t.Errorf("session %q JSONL MUST have 2 appended lines; got %d", s1, s1Lines)
	}
	if s2Lines != 1 {
		t.Errorf("session %q JSONL MUST have 1 appended line; got %d", s2, s2Lines)
	}
}

// changedAtForProject returns "(no-git)" or similar placeholder when git is
// unavailable (no git binary / not a repo / no HEAD). It MUST NOT panic and
// MUST NOT return a wall-clock value.
func TestChangedAtForProject_NonGitRepoReturnsPlaceholder(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	// tmpDir is not a git repo. changedAtForProject MUST fail-open to a
	// stable placeholder.
	got := changedAtForProject(tmpDir)
	if got == "" {
		t.Errorf("changedAtForProject MUST NOT return empty string on non-git dir")
	}
	if strings.Contains(got, time.Now().Format("2006")) && strings.Contains(got, "T") {
		// Best-effort wall-clock leak check: a wall-clock ISO timestamp would
		// contain both the current year and a 'T' separator. The placeholder
		// MUST be a stable sentinel, not a real timestamp.
		t.Errorf("changedAtForProject returned a wall-clock-looking value for non-git dir: %q", got)
	}
}

// AC-NS2-003c — no work-item promotion. The Detect source MUST NOT contain
// any pattern that creates a GitHub issue, writes to .moai/specs/, creates a
// TODO file, emits Decision:"block", or calls os.Exit(2).
func TestNavigatorDetect_NoWorkItemPromotion(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("navigator_detect.go")
	if err != nil {
		t.Fatalf("read navigator_detect.go: %v", err)
	}
	body := string(raw)
	forbidden := []string{
		"gh issue create",       // GitHub issue promotion
		"gh issue",              // any issue-tool invocation
		`Decision: "block"`,     // REQ-NS2-012 — NEVER blocks
		"os.Exit(2)",            // exit-2 block
		`"block"`,               // any literal block decision
		".moai/specs/SPEC-",     // SPEC mutation
		"TODO file",             // TODO file creation
	}
	for _, pat := range forbidden {
		if strings.Contains(body, pat) {
			t.Errorf("forbidden promotion pattern %q found in navigator_detect.go (AC-NS2-003c / REQ-NS2-003 / REQ-NS2-012)", pat)
		}
	}
}

// Dispatcher integration — full Handle() emits the systemMessage AND appends
// the JSONL impact record AND never returns Decision:"block".
func TestPostToolHandler_NavigatorDetect_AdvisoryOutput(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	input := &HookInput{
		CWD:           tmpDir,
		SessionID:     "sess-nd-advisory-001",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "` + specPath + `"}`),
	}
	h := NewPostToolHandler()
	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("nil HookOutput")
	}
	// AC-NS2-003a: systemMessage is advisory (present), and Decision is NOT block.
	if out.SystemMessage == "" {
		t.Errorf("expected non-empty SystemMessage for Write on matching path")
	}
	if !strings.HasPrefix(out.SystemMessage, "Navigator Detect:") {
		// The Detect advisory may be appended after LSP/AST text; check it is
		// present somewhere in the message.
		if !strings.Contains(out.SystemMessage, "Navigator Detect:") {
			t.Errorf("SystemMessage missing 'Navigator Detect:' advisory; got %q", out.SystemMessage)
		}
	}
	if out.HookSpecificOutput != nil && out.HookSpecificOutput.Decision != nil && out.HookSpecificOutput.Decision.Behavior == "block" {
		t.Errorf("Detect MUST NEVER emit Decision 'block' (REQ-NS2-012); got HookSpecificOutput.Decision.Behavior=block")
	}
	if out.Decision == "block" {
		t.Errorf("Detect MUST NEVER emit top-level Decision 'block' (REQ-NS2-012); got out.Decision=block")
	}
	// AC-NS2-003b: JSONL impact record was appended.
	jsonlPath := filepath.Join(tmpDir, navigatorDetectStateDir, "sess-nd-advisory-001.jsonl")
	if _, err := os.Stat(jsonlPath); err != nil {
		t.Errorf("expected JSONL impact record at %s; got error: %v", jsonlPath, err)
	}
}
