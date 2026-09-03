package fix

// request_test.go — M3.2 draft-request emission tests (SPEC-NAVIGATOR-SYNC-005,
// REQ-NS5-001 / REQ-NS5-002 / REQ-NS5-004 / REQ-NS5-009). These tests drive the
// I/O wrapper (Run) end-to-end against fixture inputs in a tempdir: they assert
// the request.json schema, the §A.4 stdout handoff signal, byte-identical
// idempotence, provenance attribution (no wall-clock), and fail-open behavior
// across the absent-input / empty-diff-scope / baseline-unresolvable modes.
//
// Test isolation: every temp tree lives under t.TempDir() (auto-cleaned). No
// test writes to the real project root or to a real .moai/project/ directory.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- Fixture builders -------------------------------------------------------

// fixWrite creates a file under root with the given content, mkdir-ing parents.
func fixWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// navGraphFixture is a minimal M0 graph that binds one symbol subtree:
// the edge source_path "src/x/a.go" binds symbol pkg.ParseHeader to
// SPEC-X-001. Its provenance.extract_commit_sha is the default baseline
// (REQ-NS5-003 priority 2). Changing the SHA changes the resolved baseline.
const navGraphFixture = `{
  "provenance": {"extract_commit_sha": "baseline123", "captured_at": "2026-01-01T00:00:00+00:00"},
  "nodes": [
    {"entity_type": "symbol", "identifier": "pkg.ParseHeader", "display_name": "ParseHeader"},
    {"entity_type": "spec", "identifier": "SPEC-X-001", "display_name": "X"}
  ],
  "edges": [
    {"edge_type": "sym-edge", "source_node": "symbol:pkg.ParseHeader", "target_node": "spec:SPEC-X-001", "source_path": "src/x/a.go", "line_number": 5}
  ]
}`

// workItemsFixture is a minimal M2 work-items.json: one detect-sourced work
// item whose owner_path is graph-bound (src/x/a.go).
const workItemsFixture = `{
  "provenance": {"route_commit_sha": "route456", "captured_at": "2026-01-01T00:00:00+00:00"},
  "work_items": [
    {"source_kind": "detect", "owner_path": "src/x/a.go", "action": "verify the affected doc rows still hold after this edit"}
  ]
}`

// detectJSONLFixture is a minimal M1 detect JSONL row touching the graph-bound
// path src/x/a.go.
const detectJSONLFixture = `{"changed_path": "src/x/a.go", "changed_at": "2026-01-02T00:00:00Z", "affected_nodes": []}
`

// populatedFixRoot writes nav-graph.json + work-items.json + a detect JSONL row
// under root. Returns root. Used by the happy-path + idempotence + schema tests.
func populatedFixRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
	return root
}

// --- Happy path + schema ----------------------------------------------------

// TestRun_EmitsRequestJSON_HappyPath exercises AC-NS5-001a: with the four
// inputs present, Run produces fix-drafts/<id>/request.json with a non-empty
// diff_scope[], returns a Result with Status "ready", Written true, and a
// stdout signal carrying the draft_request_path.
func TestRun_EmitsRequestJSON_HappyPath(t *testing.T) {
	root := populatedFixRoot(t)

	res := Run(Options{ProjectRoot: root})

	if !res.Written {
		t.Fatalf("Expected Written=true, got %+v", res)
	}
	if res.Status != "ready" {
		t.Errorf("Status = %q, want \"ready\"", res.Status)
	}
	if res.DraftID == "" {
		t.Errorf("DraftID empty")
	}
	if res.DiffScopeCount == 0 {
		t.Errorf("DiffScopeCount = 0, want >0 (graph-bound path seeded)")
	}

	// request.json exists at the contract path.
	reqPath := filepath.Join(root, ".moai", "project", "navigator", "fix-drafts", res.DraftID, "request.json")
	if res.DraftRequestPath != reqPath {
		t.Errorf("DraftRequestPath = %q, want %q", res.DraftRequestPath, reqPath)
	}
	raw, err := os.ReadFile(reqPath)
	if err != nil {
		t.Fatalf("read request.json: %v", err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal request.json: %v\n%s", err, raw)
	}
	// Required top-level keys (plan.md §C.3).
	for _, key := range []string{"provenance", "diff_scope", "work_item_refs", "draft_instructions"} {
		if _, ok := doc[key]; !ok {
			t.Errorf("request.json missing key %q", key)
		}
	}
	// diff_scope is non-empty.
	var scope []map[string]json.RawMessage
	if err := json.Unmarshal(doc["diff_scope"], &scope); err != nil {
		t.Fatalf("unmarshal diff_scope: %v", err)
	}
	if len(scope) == 0 {
		t.Errorf("diff_scope empty on happy path")
	}
}

// TestRun_StdoutSignal_ReadyJSON exercises the design.md §A.4 handoff contract:
// the Result signal is a single JSON line with draft_request_path, status
// "ready", and draft_id.
func TestRun_StdoutSignal_ReadyJSON(t *testing.T) {
	root := populatedFixRoot(t)
	res := Run(Options{ProjectRoot: root})

	line, err := res.SignalJSON()
	if err != nil {
		t.Fatalf("SignalJSON: %v", err)
	}
	var sig map[string]string
	if err := json.Unmarshal(line, &sig); err != nil {
		t.Fatalf("signal not a JSON object: %v\n%s", err, line)
	}
	for _, key := range []string{"draft_request_path", "status", "draft_id"} {
		if v, ok := sig[key]; !ok || v == "" {
			t.Errorf("signal missing/non-empty %q: %v", key, sig)
		}
	}
	if sig["status"] != "ready" {
		t.Errorf("signal status = %q, want \"ready\"", sig["status"])
	}
	// The signal is a single line (no embedded newline other than optional trailing).
	if strings.Count(strings.TrimRight(string(line), "\n"), "\n") != 0 {
		t.Errorf("signal must be a single JSON line, got:\n%s", line)
	}
}

// --- Idempotence ------------------------------------------------------------

// TestRun_Idempotent_ByteIdenticalReRun exercises AC-NS5-004b: two runs on the
// same inputs (same HEAD-equivalent + baseline + inputs) produce the same
// draft-id and byte-identical request.json.
func TestRun_Idempotent_ByteIdenticalReRun(t *testing.T) {
	root := populatedFixRoot(t)

	first := Run(Options{ProjectRoot: root})
	if !first.Written {
		t.Fatalf("first run did not write: %+v", first)
	}
	firstBytes, err := os.ReadFile(first.DraftRequestPath)
	if err != nil {
		t.Fatalf("read first request.json: %v", err)
	}

	second := Run(Options{ProjectRoot: root})
	secondBytes, err := os.ReadFile(second.DraftRequestPath)
	if err != nil {
		t.Fatalf("read second request.json: %v", err)
	}

	if first.DraftID != second.DraftID {
		t.Errorf("draft-id not stable: %q vs %q", first.DraftID, second.DraftID)
	}
	if string(firstBytes) != string(secondBytes) {
		t.Errorf("byte-identical re-run broken:\n--- first ---\n%s\n--- second ---\n%s", firstBytes, secondBytes)
	}
}

// TestRun_DraftIDDependsOnScopeAndBaseline exercises the draft-id formula
// (plan.md §C.3): SHA-256 of sorted diff-scope + baseline SHA. Changing the
// baseline (via nav-graph provenance) changes the draft-id; identical baseline
// + scope yields the identical id.
func TestRun_DraftIDDependsOnScopeAndBaseline(t *testing.T) {
	root := populatedFixRoot(t)
	base := Run(Options{ProjectRoot: root})

	// Different baseline (mutate the nav-graph provenance SHA) → different id.
	changed := strings.Replace(navGraphFixture, "baseline123", "baseline999", 1)
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), changed)
	other := Run(Options{ProjectRoot: root})
	if other.DraftID == base.DraftID {
		t.Errorf("draft-id should change when baseline changes; both = %q", base.DraftID)
	}

	// Restore → identical id.
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
	restored := Run(Options{ProjectRoot: root})
	if restored.DraftID != base.DraftID {
		t.Errorf("draft-id should match after restore: %q vs %q", restored.DraftID, base.DraftID)
	}
}

// --- Provenance (no wall-clock) ---------------------------------------------

// TestRun_ProvenanceNoWallClock exercises AC-NS5-004a: provenance is git-sourced
// (rev-parse HEAD + committer date), never wall-clock. In a real git repo the
// captured_at equals `git log -1 --format=%cI`.
func TestRun_ProvenanceNoWallClock(t *testing.T) {
	root := initGitRepo(t, populatedFixRoot(t))
	// Commit the fixture tree so HEAD is defined, then advance one commit so
	// baseline (HEAD~1) resolves too.
	gitAddCommit(t, root, "baseline state")
	// Touch a graph-bound path so git-diff contributes to the scope.
	fixWrite(t, filepath.Join(root, "src", "x", "a.go"), "package x\n")
	gitAddCommit(t, root, "touch bound path")

	res := Run(Options{ProjectRoot: root})
	if !res.Written {
		t.Fatalf("expected write in git repo: %+v", res)
	}
	raw, err := os.ReadFile(res.DraftRequestPath)
	if err != nil {
		t.Fatalf("read request.json: %v", err)
	}
	var doc struct {
		Provenance struct {
			FixCommitSHA      string `json:"fix_commit_sha"`
			BaselineCommitSHA string `json:"baseline_commit_sha"`
			CapturedAt        string `json:"captured_at"`
		} `json:"provenance"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// captured_at must equal the committer date of HEAD (no wall-clock).
	want := gitLineT(t, root, "log", "-1", "--format=%cI")
	if want == "" {
		t.Fatalf("could not determine expected captured_at from git")
	}
	if doc.Provenance.CapturedAt != want {
		t.Errorf("captured_at = %q, want git committer date %q", doc.Provenance.CapturedAt, want)
	}
	if doc.Provenance.FixCommitSHA == "" || doc.Provenance.FixCommitSHA == "<unknown>" {
		t.Errorf("fix_commit_sha not resolved: %q", doc.Provenance.FixCommitSHA)
	}
}

// --- Fail-open (REQ-NS5-009) ------------------------------------------------

// TestRun_EmptyDiffScope_009g exercises AC-NS5-009 row 009g: when the inputs
// touch ZERO graph-bound subtrees (doc map already consistent), Run still
// writes request.json with diff_scope:[] and returns Status "consistent".
func TestRun_EmptyDiffScope_009g(t *testing.T) {
	root := t.TempDir()
	// Graph binds a path the inputs do NOT touch.
	const graph = `{
  "provenance": {"extract_commit_sha": "b1", "captured_at": "2026-01-01T00:00:00+00:00"},
  "nodes": [{"entity_type": "symbol", "identifier": "pkg.Unused", "display_name": "Unused"}],
  "edges": [{"edge_type": "sym-edge", "source_node": "symbol:pkg.Unused", "target_node": "symbol:pkg.Unused", "source_path": "src/unused.go", "line_number": 1}]
}`
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), graph)
	// work-items + detect touch a DIFFERENT, non-graph-bound path.
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"),
		`{"work_items":[{"source_kind":"detect","owner_path":"src/other.go","action":"verify"}]}`)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s.jsonl"),
		`{"changed_path":"src/other.go","changed_at":"2026-01-02T00:00:00Z"}`)

	res := Run(Options{ProjectRoot: root})
	if !res.Written {
		t.Fatalf("009g must still write request.json: %+v", res)
	}
	if res.Status != "consistent" {
		t.Errorf("Status = %q, want \"consistent\"", res.Status)
	}
	raw, err := os.ReadFile(res.DraftRequestPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc struct {
		DiffScope []json.RawMessage `json:"diff_scope"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(doc.DiffScope) != 0 {
		t.Errorf("diff_scope = %d entries, want 0", len(doc.DiffScope))
	}
	if !strings.Contains(res.Message, "stale") || !strings.Contains(res.Message, "consistent") {
		t.Errorf("Message should mention stale/consistent: %q", res.Message)
	}
}

// TestRun_WorkItemsAbsent_009a exercises AC-NS5-009 row 009a: work-items.json
// absent → degrade (diff-scope from M1 detect + git-diff only), work_item_refs
// empty, exit 0 (Run never errors).
func TestRun_WorkItemsAbsent_009a(t *testing.T) {
	root := t.TempDir()
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
	// No work-items.json.

	res := Run(Options{ProjectRoot: root})
	if !res.Written {
		t.Fatalf("009a must degrade + write: %+v", res)
	}
	if res.Status != "ready" {
		t.Errorf("Status = %q, want \"ready\" (degraded but non-empty scope)", res.Status)
	}
	raw, err := os.ReadFile(res.DraftRequestPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc struct {
		WorkItemRefs []json.RawMessage `json:"work_item_refs"`
		DiffScope    []json.RawMessage `json:"diff_scope"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(doc.WorkItemRefs) != 0 {
		t.Errorf("work_item_refs should be empty when work-items absent: %d", len(doc.WorkItemRefs))
	}
	// diff_scope still seeded by M1 detect (graph-bound path).
	if len(doc.DiffScope) == 0 {
		t.Errorf("diff_scope empty; M1 detect should still seed graph-bound subtrees")
	}
}

// TestRun_DetectAbsent_009b exercises AC-NS5-009 row 009b: detect JSONL absent
// → degrade (diff-scope from M2 owner_paths + git-diff only).
func TestRun_DetectAbsent_009b(t *testing.T) {
	root := t.TempDir()
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
	// No detect state dir.

	res := Run(Options{ProjectRoot: root})
	if !res.Written {
		t.Fatalf("009b must degrade + write: %+v", res)
	}
	raw, err := os.ReadFile(res.DraftRequestPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc struct {
		DiffScope []json.RawMessage `json:"diff_scope"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// M2 owner_path src/x/a.go is graph-bound → seeds one subtree.
	if len(doc.DiffScope) == 0 {
		t.Errorf("diff_scope empty; M2 owner_paths should seed graph-bound subtrees")
	}
}

// TestRun_NavGraphAbsent_009c exercises AC-NS5-009 row 009c: nav-graph.json
// absent → ComputeScope returns empty (no graph-bound paths), AND baseline
// degrades to HEAD~1 (or fails). In a non-git tempdir with no nav-graph and no
// compareTo, the baseline is unresolvable → this collapses to 009d (skipped).
// This test asserts the nav-graph-absent path degrades rather than crashes.
func TestRun_NavGraphAbsent_009c(t *testing.T) {
	root := initGitRepo(t, t.TempDir())
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
	gitAddCommit(t, root, "first")
	fixWrite(t, filepath.Join(root, "src", "x", "a.go"), "package x\n")
	gitAddCommit(t, root, "second")
	// compareTo unset, nav-graph absent → baseline = HEAD~1 (resolves in git repo).

	res := Run(Options{ProjectRoot: root})
	// No nav-graph → no graph-bound paths → empty diff-scope → 009g-style
	// consistent (baseline resolved via HEAD~1 in the git repo).
	if !res.Written {
		t.Fatalf("009c must degrade + write (baseline via HEAD~1): %+v", res)
	}
	if res.Status != "consistent" {
		t.Errorf("Status = %q, want \"consistent\" (empty scope, baseline resolved)", res.Status)
	}
}

// TestRun_BaselineUnresolvable_009d exercises AC-NS5-009 row 009d: no
// compareTo, no nav-graph provenance, AND HEAD~1 fails (non-git dir) → write
// NO request.json, Status "skipped".
func TestRun_BaselineUnresolvable_009d(t *testing.T) {
	root := t.TempDir()
	// Non-git dir, no nav-graph, no compareTo → baseline unresolvable.
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)

	res := Run(Options{ProjectRoot: root})
	if res.Written {
		t.Errorf("009d must NOT write request.json when baseline unresolvable: %+v", res)
	}
	if res.Status != "skipped" {
		t.Errorf("Status = %q, want \"skipped\"", res.Status)
	}
}

// TestRun_UnparseableJSON_009e exercises AC-NS5-009 row 009e: a malformed
// input is skipped and the run degrades using the well-formed inputs.
func TestRun_UnparseableJSON_009e(t *testing.T) {
	root := t.TempDir()
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
	// Malformed work-items.json.
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), `{not valid json`)
	fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)

	res := Run(Options{ProjectRoot: root})
	if !res.Written {
		t.Fatalf("009e must degrade + write (skip malformed, keep well-formed): %+v", res)
	}
	// Malformed work-items skipped → work_item_refs empty; detect still seeds scope.
	raw, err := os.ReadFile(res.DraftRequestPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc struct {
		WorkItemRefs []json.RawMessage `json:"work_item_refs"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(doc.WorkItemRefs) != 0 {
		t.Errorf("malformed work-items should yield empty work_item_refs: %d", len(doc.WorkItemRefs))
	}
}

// TestRun_CompareToOverridesBaseline exercises AC-NS5-003(b): the --compare-to
// flag wins over nav-graph provenance (priority 1).
func TestRun_CompareToOverridesBaseline(t *testing.T) {
	root := populatedFixRoot(t)
	noFlag := Run(Options{ProjectRoot: root})
	withFlag := Run(Options{ProjectRoot: root, CompareTo: "explicit-baseline-999"})
	if !noFlag.Written || !withFlag.Written {
		t.Fatalf("both should write: %+v / %+v", noFlag, withFlag)
	}
	if noFlag.DraftID == withFlag.DraftID {
		t.Errorf("compare-to should change baseline → change draft-id; both = %q", noFlag.DraftID)
	}
}

// --- git fixture helpers ----------------------------------------------------

// initGitRepo initializes a git repo at root (idempotent if already a repo).
func initGitRepo(t *testing.T, root string) string {
	t.Helper()
	gitMust(t, root, "init")
	gitMust(t, root, "config", "user.email", "t@example.com")
	gitMust(t, root, "config", "user.name", "test")
	return root
}

// gitAddCommit stages all changes and commits, allowing an empty start.
func gitAddCommit(t *testing.T, root, msg string) {
	t.Helper()
	gitMust(t, root, "add", "-A")
	// --allow-empty keeps the helper robust for the first commit.
	cmd := exec.Command("git", "commit", "-m", msg, "--allow-empty")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

// gitMust runs a git subcommand in root, failing the test on error.
func gitMust(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// gitLineT runs a read-only git command in root and returns trimmed stdout.
// (Named gitLineT to avoid colliding with the package's unexported gitLine
// helper, which the same-package test file can see.)
func gitLineT(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\n")
}

// --- Coverage: internal helpers (actionToStrategy / loadGraph / resolveRoot) ---

// TestActionToStrategy covers all three canonical strategy branches plus the
// doc-surface default fallback (AC-NS5-010: regenerate row / re-link symbol /
// draft SPEC stub).
func TestActionToStrategy(t *testing.T) {
	cases := []struct {
		action, surface, want string
	}{
		{"create a SPEC for this design feature or link existing code", "audit-report.json", "draft SPEC stub"},
		{"link this SPEC to a design feature or document its design rationale", "capability-map.md", "re-link symbol"},
		{"verify the affected doc rows still hold after this edit", "capability-symbols.json", "regenerate row"},
		{"some unrecognized action", "audit-report.json", "draft SPEC stub"},
		{"some unrecognized action", "capability-map.md", "regenerate row"},
	}
	for _, c := range cases {
		if got := actionToStrategy(c.action, c.surface); got != c.want {
			t.Errorf("actionToStrategy(%q, %q) = %q, want %q", c.action, c.surface, got, c.want)
		}
	}
}

// TestDefaultStrategyForSurface covers the no-work-item-ref default mapping.
func TestDefaultStrategyForSurface(t *testing.T) {
	if got := defaultStrategyForSurface("audit-report.json"); got != "draft SPEC stub" {
		t.Errorf("audit-report default = %q, want \"draft SPEC stub\"", got)
	}
	if got := defaultStrategyForSurface("capability-map.md"); got != "regenerate row" {
		t.Errorf("capability-map default = %q, want \"regenerate row\"", got)
	}
}

// TestLoadGraph_SchemaInvalid covers the schema-invalid path (nav-graph valid
// JSON but missing the edges array): returns a nil graph (degraded subtree
// resolution) while still surfacing the provenance baseline SHA.
func TestLoadGraph_SchemaInvalid(t *testing.T) {
	root := t.TempDir()
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"),
		`{"provenance":{"extract_commit_sha":"shaX"},"nodes":[{"entity_type":"symbol","identifier":"s","display_name":"s"}]}`)
	g, sha := loadGraph(root)
	if g != nil {
		t.Errorf("schema-invalid graph should return nil, got %+v", g)
	}
	if sha != "shaX" {
		t.Errorf("provenance SHA should still surface, got %q", sha)
	}
}

// TestLoadGraph_Unparseable covers the malformed-JSON path (009e).
func TestLoadGraph_Unparseable(t *testing.T) {
	root := t.TempDir()
	fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), `{bad json`)
	g, sha := loadGraph(root)
	if g != nil {
		t.Errorf("unparseable graph should return nil, got %+v", g)
	}
	if sha != "" {
		t.Errorf("unparseable graph should yield empty baseline SHA, got %q", sha)
	}
}

// TestLoadGraph_Absent covers the absent-file path (009c).
func TestLoadGraph_Absent(t *testing.T) {
	root := t.TempDir()
	g, sha := loadGraph(root)
	if g != nil || sha != "" {
		t.Errorf("absent graph should return (nil, \"\"), got (%+v, %q)", g, sha)
	}
}

// TestResolveRoot_Fallbacks covers the B7 path-resolution priority: explicit
// flag > $CLAUDE_PROJECT_DIR > CWD.
func TestResolveRoot_Fallbacks(t *testing.T) {
	// Explicit flag wins. Both paths are real absolute directories rather than
	// POSIX-shaped literals: resolveRoot runs its result through filepath.Abs,
	// which on Windows rewrites a rooted-but-driveless "/explicit/path" to
	// "C:\explicit\path" and would fail the comparison on the fixture rather
	// than on the behavior.
	explicit := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	if got := resolveRoot(explicit); got != explicit {
		t.Errorf("explicit flag should win: got %q, want %q", got, explicit)
	}
	// CLAUDE_PROJECT_DIR fallback when no flag.
	fallback := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", fallback)
	if got := resolveRoot(""); got != fallback {
		t.Errorf("CLAUDE_PROJECT_DIR fallback: got %q, want %q", got, fallback)
	}
}

// TestResult_SignalJSON_ConsistentHasMessage verifies the 009g signal carries
// a human-readable message field alongside the path/status/draft_id.
func TestResult_SignalJSON_ConsistentHasMessage(t *testing.T) {
	res := Result{
		DraftRequestPath: "/p/request.json",
		Status:           "consistent",
		DraftID:          "abc",
		Message:          "0 stale subtrees, doc map consistent",
	}
	line, err := res.SignalJSON()
	if err != nil {
		t.Fatalf("SignalJSON: %v", err)
	}
	var sig map[string]string
	if err := json.Unmarshal(line, &sig); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sig["message"] != "0 stale subtrees, doc map consistent" {
		t.Errorf("message field = %q", sig["message"])
	}
	if sig["status"] != "consistent" {
		t.Errorf("status = %q", sig["status"])
	}
}
