package cli

// SPEC-MCP-WORKTREE-ROOT-001 — the MCP tools whose subject is the caller's
// working tree accept an optional project_root, because resolveProjectDir()
// prefers CLAUDE_PROJECT_DIR and Claude Code sets that to the PRIMARY checkout
// even for a worktree session. Without the parameter, an audit issued from a
// worktree audits the primary and a worktree-only SPEC is silently absent.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// projectRootTools are the in-scope tools of REQ-1 that are registered on the
// MCP server surface. codex_audit and audit_multi are asserted separately (M2),
// because their assertion targets a params map rather than a catalogue result.
var projectRootSpecTools = []string{"spec_audit", "spec_drift", "spec_progress"}

// newProbeProject writes a minimal project tree carrying one SPEC, so a
// catalogue scan rooted there has something distinctive to find.
func newProbeProject(t *testing.T, specID string) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "---\nid: " + specID + "\ntitle: \"probe\"\nversion: \"0.1.0\"\n" +
		"status: draft\ncreated: 2026-08-22\nupdated: 2026-08-22\nauthor: t\n" +
		"priority: P3\nphase: \"v3.1.3\"\nmodule: \"internal/cli\"\n" +
		"lifecycle: exploratory\ntags: \"probe\"\n---\n\n# " + specID + "\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Return the CANONICAL path. validateProjectRoot resolves symlinks, and
	// t.TempDir() hands back a symlinked path on macOS (/var → /private/var), so
	// a test comparing a handler's answer against this value would otherwise be
	// comparing two spellings of one directory and failing on the spelling.
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", root, err)
	}
	return canonical
}

// AC-REQ-1a (schema half): each in-scope SPEC tool declares the optional input.
func TestSpecTools_DeclareProjectRootInput(t *testing.T) {
	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	seen := map[string]bool{}
	for _, tool := range res.Tools {
		for _, want := range projectRootSpecTools {
			if tool.Name != want {
				continue
			}
			seen[want] = true
			b, _ := json.Marshal(tool.InputSchema)
			if !strings.Contains(string(b), "project_root") {
				t.Errorf("%s inputSchema does not declare project_root; schema=%s", want, b)
			}
		}
	}
	for _, want := range projectRootSpecTools {
		if !seen[want] {
			t.Errorf("tool %q not registered — cannot assert its schema", want)
		}
	}
}

// AC-REQ-1a (behaviour half, PRESENT direction): a catalogue scan rooted at a
// named tree finds a SPEC that exists only there.
func TestSpecAudit_ProjectRootRedirectsTheCatalogue(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBEONE-901")
	got := callSpecAuditJSON(t, map[string]any{"project_root": root})
	if !strings.Contains(got, "SPEC-PROBEONE-901") && !strings.Contains(got, "\"total_specs\":1") {
		t.Errorf("audit rooted at the named tree did not see its only SPEC; got=%s", got)
	}
}

// AC-REQ-1a (behaviour half, ABSENT direction). The same probe SPEC is NOT
// visible when the parameter is omitted, because resolution falls back to the
// default root. A test that only asserted the present direction would pass just
// as well if the parameter were ignored and the default tree happened to
// contain the SPEC — which is exactly how this class of defect hides.
func TestSpecAudit_WithoutProjectRootDoesNotSeeTheProbe(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBETWO-902")
	_ = root // deliberately NOT passed
	got := callSpecAuditJSON(t, map[string]any{})
	if strings.Contains(got, "SPEC-PROBETWO-902") {
		t.Errorf("audit without project_root saw a SPEC that exists only in the probe tree; got=%s", got)
	}
}

// AC-REQ-2: an absent or empty parameter resolves exactly as before.
func TestResolveToolProjectRoot_AbsentAndEmptyMatchTheDefault(t *testing.T) {
	want := resolveProjectDir()
	for _, args := range []map[string]any{{}, {"project_root": ""}, {"project_root": "   "}} {
		got, err := resolveToolProjectRoot(newToolRequest(args))
		if err != nil {
			t.Fatalf("args=%v: unexpected error: %v", args, err)
		}
		if got != want {
			t.Errorf("args=%v: root=%q, want the default %q", args, got, want)
		}
	}
}

// AC-REQ-3: a named root that cannot be a project is REJECTED, never silently
// replaced by the default. A fallback would return a caller who mistyped its
// own worktree path to auditing the primary — this SPEC's defect, arriving
// through the mechanism meant to fix it, and reporting success while it does.
func TestResolveToolProjectRoot_RejectsAnUnusableRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir")

	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	noMoai := t.TempDir()

	for name, path := range map[string]string{
		"nonexistent":      missing,
		"file-not-dir":     file,
		"dir-without-moai": noMoai,
	} {
		got, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": path}))
		if err == nil {
			t.Errorf("%s: expected rejection, got root=%q", name, got)
			continue
		}
		if !strings.Contains(err.Error(), quotedPath(path)) {
			t.Errorf("%s: error does not name the offending path %q: %v", name, path, err)
		}
		if got != "" {
			t.Errorf("%s: rejected root must be empty, got %q", name, got)
		}
	}
}

// A rejected root must surface as a tool error rather than an audit of the
// default tree — the same guarantee as above, asserted at the handler boundary
// where a caller actually observes it.
func TestSpecAudit_RejectedProjectRootDoesNotAuditTheDefault(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent")
	res, err := handleSpecAudit(context.Background(), newToolRequest(map[string]any{"project_root": missing}))
	if err != nil {
		t.Fatalf("handleSpecAudit returned a hard error: %v", err)
	}
	got := resultTextOf(res)
	if !strings.Contains(got, quotedPath(missing)) {
		t.Errorf("handler did not surface the offending path; got=%s", got)
	}
	if strings.Contains(got, "total_specs") {
		t.Errorf("handler audited something despite an unusable project_root; got=%s", got)
	}
}

// The other two SPEC tools, exercised rather than schema-asserted.
//
// The sync audit found spec_progress and spec_drift resting on read-the-diff
// assurance: their declaration was checked, their parameter-PRESENT behaviour
// never was — not by a test and not by the live probe, which called only
// spec_audit. AC-1a names spec_audit, so this was not an AC miss; it was two of
// five in-scope tools nobody had watched work. The risk was low precisely
// because all three share resolveToolProjectRoot — but "it shares a covered
// helper" is an argument, not an observation.
func TestSpecProgress_ProjectRootRedirectsTheScanner(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBESCAN-905")

	withParam := callSpecToolJSON(t, handleSpecProgress, map[string]any{"project_root": root})
	if !strings.Contains(withParam, "SPEC-PROBESCAN-905") {
		t.Errorf("spec_progress rooted at the named tree did not list its only SPEC; got=%s", withParam)
	}

	// The absent direction, for the same reason AC-1a demands it: a
	// present-only assertion passes just as well if the parameter is ignored
	// and the default tree happens to hold the SPEC.
	without := callSpecToolJSON(t, handleSpecProgress, map[string]any{})
	if strings.Contains(without, "SPEC-PROBESCAN-905") {
		t.Errorf("spec_progress without project_root listed a SPEC that exists only in the probe tree; got=%s", without)
	}
}

func TestSpecDrift_ProjectRootRedirectsTheCatalogue(t *testing.T) {
	// Two SPECs rather than one, so the count that proves the redirect is a
	// number the ambient tree is unlikely to match by coincidence.
	root := newProbeProject(t, "SPEC-PROBEDRIFT-906")
	addProbeSpec(t, root, "SPEC-PROBEDRIFT-907")

	withParam := callSpecToolJSON(t, handleSpecDrift, map[string]any{"project_root": root})
	if !strings.Contains(withParam, `"total_specs":2`) {
		t.Errorf("spec_drift rooted at the named tree did not audit its two SPECs; got=%s", withParam)
	}

	without := callSpecToolJSON(t, handleSpecDrift, map[string]any{})
	if strings.Contains(without, `"total_specs":2`) {
		t.Errorf("spec_drift without project_root reported the probe tree's count, so the parameter was not what redirected it; got=%s", without)
	}
}

// --- helpers -------------------------------------------------------------

// resultTextOf renders a tool result's TEXT content as one string.
//
// An assertion about a rejected path must match the message a caller reads, not
// its JSON transport encoding. The two differ on Windows: validateProjectRoot
// formats the offending path with %q, which escapes each separator once, and
// json.Marshal escapes them again — so a needle built from the raw path is
// absent from both spellings even when the message names the path correctly.
func resultTextOf(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}

// quotedPath is the needle a %q-formatted message actually contains on every
// platform: on POSIX the path in quotes, on Windows the same with each
// backslash escaped. Comparing against the raw path is a POSIX-only assumption
// that happens to hold because POSIX paths carry nothing %q escapes.
func quotedPath(p string) string { return strconv.Quote(p) }

// addProbeSpec writes one more minimal SPEC into an existing probe tree.
func addProbeSpec(t *testing.T, root, specID string) {
	t.Helper()
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "---\nid: " + specID + "\ntitle: \"probe\"\nversion: \"0.1.0\"\n" +
		"status: draft\ncreated: 2026-08-23\nupdated: 2026-08-23\nauthor: t\n" +
		"priority: P3\nphase: \"v3.1.3\"\nmodule: \"internal/cli\"\n" +
		"lifecycle: exploratory\ntags: \"probe\"\n---\n\n# " + specID + "\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// callSpecToolJSON invokes any SPEC-tool handler directly and returns its
// rendered content as one string.
func callSpecToolJSON(
	t *testing.T,
	handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error),
	args map[string]any,
) string {
	t.Helper()
	res, err := handler(context.Background(), newToolRequest(args))
	if err != nil {
		t.Fatalf("handler returned a hard error: %v", err)
	}
	b, _ := json.Marshal(res)
	return string(b)
}

func newToolRequest(args map[string]any) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Name = "spec_audit"
	req.Params.Arguments = args
	return req
}

// callSpecAuditJSON invokes the spec_audit handler directly and returns its
// rendered content as one string, so a test can assert on what a caller sees.
func callSpecAuditJSON(t *testing.T, args map[string]any) string {
	t.Helper()
	res, err := handleSpecAudit(context.Background(), newToolRequest(args))
	if err != nil {
		t.Fatalf("handleSpecAudit returned a hard error: %v", err)
	}
	b, _ := json.Marshal(res)
	return string(b)
}

// AC-t183 (boundary canonicalization): a project_root reached through a symlink
// resolves to the real directory, not to the symlink path.
//
// filepath.Abs does not resolve symlinks, so before this the helper returned the
// symlink path verbatim. Harmless while nothing compares the result against
// another path — and exactly the hole a future containment constraint (compare
// the resolved root against the repository's git common dir) would be walked
// through, because the two sides would be spelled differently for the same
// directory. Canonicalizing here is what keeps such a comparison meaningful
// rather than decorative.
func TestValidateProjectRoot_ResolvesSymlinkedRoot(t *testing.T) {
	real := newProbeProject(t, "SPEC-PROBELINK-906")
	realCanonical, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatalf("EvalSymlinks(real): %v", err)
	}

	link := filepath.Join(t.TempDir(), "linked-root")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}

	got, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": link}))
	if err != nil {
		t.Fatalf("unexpected rejection of a symlinked root: %v", err)
	}
	if got != realCanonical {
		t.Errorf("root=%q, want the canonical target %q (symlink was %q)", got, realCanonical, link)
	}
}

// A symlink whose target does not exist is rejected, and the error names the
// path the caller actually supplied — not the resolved one it never typed.
func TestValidateProjectRoot_RejectsBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "dangling")
	if err := os.Symlink(filepath.Join(dir, "no-such-target"), link); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}

	got, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": link}))
	if err == nil {
		t.Fatalf("expected rejection of a broken symlink, got root=%q", got)
	}
	if !strings.Contains(err.Error(), quotedPath(link)) {
		t.Errorf("error does not name the supplied path %q: %v", link, err)
	}
	if got != "" {
		t.Errorf("rejected root must be empty, got %q", got)
	}
}

// --- t236 / issue #1640 residuals (M2): fallback visibility + verify tools ---

// The fallback's silent misdirection (live gap L2) is only fixable if the
// response SAYS which tree it read. spec_progress must carry an "_root"
// provenance block: "param" (no warning) when the caller named a tree, and a
// non-empty warning when resolution fell back to spawn-frozen state.
func TestSpecProgress_ResponseCarriesRootProvenance(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBEPROV-908")

	withParam := callSpecToolJSON(t, handleSpecProgress, map[string]any{"project_root": root})
	if !strings.Contains(withParam, "SPEC-PROBEPROV-908") {
		t.Errorf("spec_progress rooted at the named tree did not list its only SPEC; got=%s", withParam)
	}
	if !strings.Contains(withParam, "_root") || !strings.Contains(withParam, `"source":"param"`) {
		t.Errorf("spec_progress response lacks param provenance; got=%s", withParam)
	}

	without := callSpecToolJSON(t, handleSpecProgress, map[string]any{})
	if strings.Contains(without, "SPEC-PROBEPROV-908") {
		t.Errorf("spec_progress without project_root listed a SPEC that exists only in the probe tree; got=%s", without)
	}
	if !strings.Contains(without, `"warning"`) {
		t.Errorf("spec_progress fallback resolution carried no warning; got=%s", without)
	}
}

// verify_snapshot ignored project_root entirely (live gap L3): a record made
// with a named root must land under THAT tree, and the fallback tree must stay
// clean. CLAUDE_PROJECT_DIR is pinned to a scratch dir so the pre-fix
// behaviour (fallback write) cannot leak into a real checkout.
func TestVerifySnapshot_HonorsProjectRoot(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBEVERIFY-909")
	fallback := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", fallback)

	key := "t236red00000000:verify-probe"
	cmd := "go test -run TestVerifySnapshot_HonorsProjectRoot"

	res, err := handleVerifySnapshot(context.Background(), newToolRequest(map[string]any{
		"project_root": root,
		"key":          key,
		"command":      cmd,
		"exit_code":    0,
	}))
	if err != nil {
		t.Fatalf("handleVerifySnapshot (record): %v", err)
	}
	recBlob, marshalErr := json.Marshal(res)
	if marshalErr != nil {
		t.Fatalf("marshal record result: %v", marshalErr)
	}
	if !strings.Contains(string(recBlob), `"action":"record"`) {
		t.Fatalf("expected record action; got=%s", recBlob)
	}

	snap, loadErr := verify.Load(root, key)
	if loadErr != nil {
		t.Fatalf("verify.Load(param root): %v", loadErr)
	}
	if snap == nil {
		t.Fatalf("snapshot missing under the named project_root — the parameter was ignored and the record landed on the fallback tree")
	}
	if snap.FindCommand(cmd) == nil {
		t.Fatalf("recorded snapshot under the named root lacks the check %q; snap=%+v", cmd, snap)
	}

	fallbackSnap, fallbackErr := verify.Load(fallback, key)
	if fallbackErr != nil {
		t.Fatalf("verify.Load(fallback): %v", fallbackErr)
	}
	if fallbackSnap != nil {
		t.Fatalf("record leaked into the fallback tree — project_root was not what redirected it")
	}

	// The read path must honor the parameter too, and both paths carry the
	// provenance block.
	loaded := callSpecToolJSON(t, handleVerifySnapshot, map[string]any{"project_root": root, "key": key})
	if !strings.Contains(loaded, `"action":"load"`) || !strings.Contains(loaded, cmd) {
		t.Errorf("verify_snapshot load from the named root lost the recorded check; got=%s", loaded)
	}
	if !strings.Contains(loaded, `"source":"param"`) {
		t.Errorf("verify_snapshot response lacks param provenance; got=%s", loaded)
	}
}

// verify_trend read the same frozen fallback (live gap L3): the trend of a
// key recorded under a named root is visible only when that root is passed.
func TestVerifyTrend_HonorsProjectRoot(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBETREND-910")
	fallback := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", fallback)

	key := "t236red00000000:trend-probe"
	cmd := "go test -run TestVerifyTrend_HonorsProjectRoot"

	if _, err := handleVerifySnapshot(context.Background(), newToolRequest(map[string]any{
		"project_root": root,
		"key":          key,
		"command":      cmd,
		"exit_code":    0,
	})); err != nil {
		t.Fatalf("handleVerifySnapshot (record): %v", err)
	}

	withParam := callSpecToolJSON(t, handleVerifyTrend, map[string]any{"project_root": root, "key": key})
	if !strings.Contains(withParam, cmd) {
		t.Errorf("verify_trend rooted at the named tree lost the recorded check; got=%s", withParam)
	}
	if !strings.Contains(withParam, `"source":"param"`) {
		t.Errorf("verify_trend response lacks param provenance; got=%s", withParam)
	}

	without := callSpecToolJSON(t, handleVerifyTrend, map[string]any{"key": key})
	if strings.Contains(without, cmd) {
		t.Errorf("verify_trend without project_root saw a check that exists only in the probe tree; got=%s", without)
	}
}

// The resolver reports which tier produced the root, so a response can say so:
// an explicit parameter resolves as source "param" with NO warning; an empty
// request resolves from a fallback tier and MUST carry the warning that makes
// the fallback visible instead of silent (live gap L2).
func TestResolveToolProjectRootWithSource_ParamAndFallback(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBESRC-911")

	withParam := newToolRequest(map[string]any{"project_root": root})
	gotParam, paramSource, paramErr := resolveToolProjectRootWithSource(withParam)
	if paramErr != nil {
		t.Fatalf("param resolution errored: %v", paramErr)
	}
	if gotParam != root {
		t.Fatalf("param root = %q, want %q", gotParam, root)
	}
	prov := rootProvenanceMap(gotParam, paramSource)
	if prov["source"] != "param" {
		t.Fatalf("param provenance = %+v", prov)
	}
	if _, has := prov["warning"]; has {
		t.Fatalf("param resolution must not warn: %+v", prov)
	}

	fallbackRoot, fallbackSource, fallbackErr := resolveToolProjectRootWithSource(newToolRequest(map[string]any{}))
	if fallbackErr != nil {
		t.Fatalf("fallback resolution errored: %v", fallbackErr)
	}
	if fallbackSource == "param" {
		t.Fatalf("empty request must not resolve as param: source=%q root=%q", fallbackSource, fallbackRoot)
	}
	provFallback := rootProvenanceMap(fallbackRoot, fallbackSource)
	if w, ok := provFallback["warning"].(string); !ok || w == "" {
		t.Fatalf("fallback must carry a non-empty warning: %+v", provFallback)
	}
}
