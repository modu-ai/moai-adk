package cli

// SPEC-AUDIT-BUILD-IDENTITY-001 — audit verdicts carry the producing binary's
// build commit (and, on an ancestor build, the lag advisory) as flat sibling
// fields on ReviewOutput and ConvergenceResult.
//
// Every test here is judged against tree-built test binaries only (acceptance
// §D.2): backends are stubbed through the existing seams (codexReviewRPC /
// codexLookPath, glmKeyLoader + glmHTTPClient, backendCall), the lag comparison
// is stubbed through binlag.Comparer, no network is touched, and all fixtures
// live under t.TempDir().

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/binlag"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// The two JSON key names, held as string constants so every assertion below
// reads the SAME names for all three entry points — a rename in one handler
// and not the others shows up as a diff here (AC-ABI-001 shape clause).
const (
	keyBuildCommit = "build_commit"
	keyBuildLag    = "build_lag"
)

// Fixture identities. The version is HELD CONSTANT across the pair used by
// the anti-mutant test: M2 (a field named build_commit carrying the version
// string) survives a version-differing pair and dies only when the two builds
// share a version and differ in commit alone (acceptance AC-ABI-003 [HARD]).
const (
	fxVersion  = "v3.1.3"
	fxCommitA  = "abc123def456789"
	fxCommitB  = "999fedcba012345"
	fxAncestor = "aaa111111111111"
	fxHead     = "bbb222222222222"
)

// ─── shared fixtures ───

// withVersionIdentity pins pkg/version for the test and restores it on cleanup.
func withVersionIdentity(t *testing.T, v, commit string) {
	t.Helper()
	prevV, prevC := version.Version, version.Commit
	version.Version, version.Commit = v, commit
	t.Cleanup(func() { version.Version, version.Commit = prevV, prevC })
}

// binlagStub replaces binlag.Comparer with a canned verdict, counting calls and
// capturing the request so a test can observe the comparison actually running
// through the single Evaluate seam (AC-ABI-007 observation 1) and which
// directory it compared against (AC-ABI-006 empty-project_root clause).
type binlagStub struct {
	mu      sync.Mutex
	calls   int
	lastReq binlag.Request
	verdict binlag.Verdict
}

func (s *binlagStub) compare(_ context.Context, req binlag.Request) binlag.Verdict {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.lastReq = req
	return s.verdict
}

func (s *binlagStub) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func withBinlagStub(t *testing.T, v binlag.Verdict) *binlagStub {
	t.Helper()
	stub := &binlagStub{verdict: v}
	prev := binlag.Comparer
	binlag.Comparer = stub.compare
	t.Cleanup(func() { binlag.Comparer = prev })
	return stub
}

// behindStubVerdict is the StatusBehind verdict whose advisory names both
// commits; freshStubVerdict is the silent control.
func behindStubVerdict() binlag.Verdict {
	return binlag.Verdict{Status: binlag.StatusBehind, BinaryCommit: fxAncestor, SourceHead: fxHead}
}

func freshStubVerdict() binlag.Verdict {
	return binlag.Verdict{Status: binlag.StatusFresh, BinaryCommit: fxCommitA, SourceHead: fxHead}
}

// withCodexReviewStub replaces the codex RPC seam with a canned pass verdict.
func withCodexReviewStub(t *testing.T) {
	t.Helper()
	prevRPC, prevLook := codexReviewRPC, codexLookPath
	codexReviewRPC = func(context.Context, string, string, map[string]any) (ReviewOutput, error) {
		return ReviewOutput{Verdict: "pass", Summary: "stub pass", Findings: []Finding{}, NextSteps: []string{"ok"}}, nil
	}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	t.Cleanup(func() { codexReviewRPC, codexLookPath = prevRPC, prevLook })
}

// withBackendCallStub replaces the audit_multi backend seam with a canned pass.
func withBackendCallStub(t *testing.T) {
	t.Helper()
	prev := backendCall
	backendCall = func(context.Context, string, string, string, string) ReviewOutput {
		return ReviewOutput{Verdict: "pass", Summary: "stub pass", Findings: []Finding{}, NextSteps: []string{"ok"}}
	}
	t.Cleanup(func() { backendCall = prev })
}

// ─── the three entry points, driven uniformly ───

type auditEntryPoint struct {
	name string
	// run executes one audit through the handler. root names the tree to pass
	// as project_root; "" means the argument is OMITTED (the audit_multi
	// documented normal call).
	run func(t *testing.T, root string) (*mcp.CallToolResult, error)
}

func auditEntryPoints() []auditEntryPoint {
	return []auditEntryPoint{
		{"codex_audit", runCodexAuditEntry},
		{"glm_audit", runGLMAuditEntry},
		{"audit_multi", runAuditMultiEntry},
	}
}

func runCodexAuditEntry(t *testing.T, root string) (*mcp.CallToolResult, error) {
	t.Helper()
	withCodexReviewStub(t)
	args := map[string]any{}
	if root != "" {
		args[projectRootArg] = root
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = "codex_audit"
	req.Params.Arguments = args
	return handleCodexAudit(context.Background(), req)
}

func runGLMAuditEntry(t *testing.T, root string) (*mcp.CallToolResult, error) {
	t.Helper()
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "stub pass", Findings: []Finding{}, NextSteps: []string{"ok"}})}
	withGLMSeams(t, "test-glm-key", stub)
	if root == "" {
		t.Fatal("glm_audit entry needs a named root: the diff collector has nothing to carry otherwise")
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = "glm_audit"
	req.Params.Arguments = map[string]any{projectRootArg: root}
	return handleGLMAudit(context.Background(), req)
}

func runAuditMultiEntry(t *testing.T, root string) (*mcp.CallToolResult, error) {
	t.Helper()
	withBackendCallStub(t)
	args := map[string]any{
		"claude_verdict": map[string]any{
			"verdict": "pass", "summary": "claude pass", "findings": []any{}, "next_steps": []any{},
		},
	}
	if root != "" {
		args[projectRootArg] = root
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = "audit_multi"
	req.Params.Arguments = args
	return handleAuditMulti(context.Background(), req)
}

// toolResultJSON parses the structured JSON payload out of a tool result.
func toolResultJSON(t *testing.T, res *mcp.CallToolResult) map[string]any {
	t.Helper()
	if res == nil {
		t.Fatal("nil tool result")
	}
	if res.IsError {
		t.Fatalf("unexpected IsError tool result: %+v", res)
	}
	if len(res.Content) == 0 {
		t.Fatal("no content blocks")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] type = %T, want TextContent", res.Content[0])
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(tc.Text), &m); err != nil {
		t.Fatalf("result text is not a JSON object: %v (text=%q)", err, tc.Text)
	}
	return m
}

// reviewTreeFixture is a valid MoAI project root (passes validateProjectRoot)
// holding exactly one uncommitted change, so the glm_audit diff collector has
// material to carry.
func reviewTreeFixture(t *testing.T) string {
	t.Helper()
	return newGLMReviewTree(t, true)
}

// noGitTreeFixture is a valid MoAI project root that is NOT a git working tree.
func noGitTreeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	return root
}

// assertBuildCommitPresent is AC-ABI-001 clause 1+2: the key exists, is
// non-empty, and carries the exact build commit — evaluated BEFORE any shape
// comparison so an all-empty omitempty result cannot pass vacuously.
func assertBuildCommitPresent(t *testing.T, m map[string]any, wantCommit, label string) {
	t.Helper()
	raw, ok := m[keyBuildCommit]
	if !ok {
		t.Fatalf("%s: %q key ABSENT — the verdict carries no build identity", label, keyBuildCommit)
	}
	got, ok := raw.(string)
	if !ok {
		t.Fatalf("%s: %q is %T, want string", label, keyBuildCommit, raw)
	}
	if got == "" {
		t.Fatalf("%s: %q is EMPTY — an omitted identity asserts nothing", label, keyBuildCommit)
	}
	if got != wantCommit {
		t.Fatalf("%s: %q = %q, want %q", label, keyBuildCommit, got, wantCommit)
	}
}

// assertNoVersionString is the anti-mutant scan (AC-ABI-003): no key in the
// result may carry the version string, and no key may be version-named —
// a commit field holding the version (M2) or a version field beside the
// commit (M1) both die here.
func assertNoVersionString(t *testing.T, m map[string]any, label string) {
	t.Helper()
	for k, v := range m {
		if strings.Contains(strings.ToLower(k), "version") {
			t.Errorf("%s: result carries a version-named key %q (REQ-ABI-003 — a version beside the commit invites reading the wrong one)", label, k)
		}
		if s, ok := v.(string); ok && s == version.GetVersion() {
			t.Errorf("%s: key %q carries the version string %q — build identity keyed on version, not commit (M2 mutant)", label, k, s)
		}
	}
}

// ─── AC-ABI-001 — verdicts carry a non-empty build_commit, one shape ───

func TestAuditVerdictCarriesBuildCommit(t *testing.T) {
	withVersionIdentity(t, fxVersion, fxCommitA)
	withBinlagStub(t, behindStubVerdict()) // build_lag present ⇒ shape comparison is non-vacuous

	commits := map[string]string{}
	for _, ep := range auditEntryPoints() {
		root := reviewTreeFixture(t)
		res, err := ep.run(t, root)
		if err != nil {
			t.Fatalf("%s: handler error: %v", ep.name, err)
		}
		m := toolResultJSON(t, res)

		// Clause 1+2 FIRST: presence + non-emptiness (the anti-vacuous guard).
		assertBuildCommitPresent(t, m, fxCommitA, ep.name)

		// Clause 3: shape — both keys, both strings, in this result.
		lag, ok := m[keyBuildLag]
		if !ok {
			t.Errorf("%s: %q key absent while the comparison returned StatusBehind", ep.name, keyBuildLag)
		} else if _, isStr := lag.(string); !isStr {
			t.Errorf("%s: %q is %T, want string", ep.name, keyBuildLag, lag)
		}
		commits[ep.name] = m[keyBuildCommit].(string)
	}

	// Clause 3 (cross-entry): the same commit value on all three.
	for _, ep := range auditEntryPoints() {
		if commits[ep.name] != fxCommitA {
			t.Errorf("%s: build_commit %q differs across entry points (want %q everywhere)", ep.name, commits[ep.name], fxCommitA)
		}
	}
}

// ─── AC-ABI-002 — the persisted record carries the same commit ───

func TestPersistedConvergenceCarriesBuildCommit(t *testing.T) {
	withVersionIdentity(t, fxVersion, fxCommitA)
	withBinlagStub(t, freshStubVerdict())
	// The secondary backends are stubbed out: this card's judgment runs on
	// tree-built binaries with NO network and NO external codex/glm processes
	// (acceptance §D.2). An unstubbed backendCall here shells into the real
	// codex and waits out its fail-open timeout.
	withBackendCallStub(t)

	root := t.TempDir()
	sid := "sess-build-identity"
	res := runMultiAudit(context.Background(), claudeReview("pass"), "", "", MultiAuditConfig{
		Gates:       config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateAdvisory},
		SessionID:   sid,
		ProjectRoot: root,
	}, nil)

	returned, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal returned result: %v", err)
	}
	var returnedMap map[string]any
	if err := json.Unmarshal(returned, &returnedMap); err != nil {
		t.Fatalf("unmarshal returned result: %v", err)
	}
	assertBuildCommitPresent(t, returnedMap, fxCommitA, "returned ConvergenceResult")

	raw, err := os.ReadFile(filepath.Join(root, ".moai", "state", "audit-multi", sid+".json"))
	if err != nil {
		t.Fatalf("read persisted record: %v", err)
	}
	var persisted map[string]any
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatalf("unmarshal persisted record: %v", err)
	}
	assertBuildCommitPresent(t, persisted, fxCommitA, "persisted record")

	if persisted[keyBuildCommit] != returnedMap[keyBuildCommit] {
		t.Errorf("persisted %q = %v, returned %q = %v — the file must carry the same commit the verdict carried",
			keyBuildCommit, persisted[keyBuildCommit], keyBuildCommit, returnedMap[keyBuildCommit])
	}
}

// ─── AC-ABI-003 — the anti-mutant: version alone never passes ───

func TestBuildIdentityVersionAloneIsRejected(t *testing.T) {
	withBinlagStub(t, freshStubVerdict())

	for _, ep := range auditEntryPoints() {
		root := reviewTreeFixture(t)

		withVersionIdentity(t, fxVersion, fxCommitA)
		resA, err := ep.run(t, root)
		if err != nil {
			t.Fatalf("%s: handler error (build A): %v", ep.name, err)
		}
		withVersionIdentity(t, fxVersion, fxCommitB)
		resB, err := ep.run(t, root)
		if err != nil {
			t.Fatalf("%s: handler error (build B): %v", ep.name, err)
		}

		ma, mb := toolResultJSON(t, resA), toolResultJSON(t, resB)

		// 2. Both non-empty FIRST (an empty commit asserts nothing) — then
		// 1. same version, different commit ⇒ recorded commits differ.
		assertBuildCommitPresent(t, ma, fxCommitA, ep.name+" build A")
		assertBuildCommitPresent(t, mb, fxCommitB, ep.name+" build B")
		if ma[keyBuildCommit] == mb[keyBuildCommit] {
			t.Errorf("%s: two builds differing only in commit recorded the SAME build_commit %v — identity is not keyed on the commit (M2 mutant survives)", ep.name, ma[keyBuildCommit])
		}

		// 3. No version string anywhere in the result (kills M1 and M2).
		assertNoVersionString(t, ma, ep.name+" build A")
		assertNoVersionString(t, mb, ep.name+" build B")
	}
}

// ─── AC-ABI-004 — absent identity leaves the JSON key set unchanged ───

func TestBuildIdentityOmittedWhenAbsent(t *testing.T) {
	// Explicit expected key sets (acceptance AC-ABI-004: named in the test, not
	// read off the pre-change tree). synthesis_note / gate_unmet stay absent
	// while empty, per their own omitempty precedent.
	review := ReviewOutput{
		Verdict:   "pass",
		Summary:   "s",
		Findings:  []Finding{},
		NextSteps: []string{"n"},
	}
	reviewJSON, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("marshal ReviewOutput: %v", err)
	}
	var rm map[string]any
	if err := json.Unmarshal(reviewJSON, &rm); err != nil {
		t.Fatalf("unmarshal ReviewOutput: %v", err)
	}
	wantReview := []string{"verdict", "summary", "findings", "next_steps"}
	gotReview := sortedMapKeys(rm)
	sort.Strings(wantReview)
	if strings.Join(gotReview, ",") != strings.Join(wantReview, ",") {
		t.Errorf("ReviewOutput empty-identity key set = %v, want %v (build_commit/build_lag must vanish via omitempty)", gotReview, wantReview)
	}

	result := ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{},
		OverallVerdict:     "pass",
		DisagreementFlag:   boolPtr(false),
		ResidualRiskNote:   "",
		FailOpenBackends:   []string{},
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal ConvergenceResult: %v", err)
	}
	var cm map[string]any
	if err := json.Unmarshal(resultBytes, &cm); err != nil {
		t.Fatalf("unmarshal ConvergenceResult: %v", err)
	}
	// participant_count joined the always-present key set with
	// SPEC-AUDIT-PARTICIPANT-COUNT-001 (deliberately NOT omitempty — a
	// visible 0 is REQ-APC-001 row a); the AC-ABI-004 subject remains the
	// omitempty'd build_commit/build_lag, which stay absent here.
	wantResult := []string{"per_backend_verdicts", "overall_verdict", "disagreement_flag", "participant_count", "residual_risk_note", "fail_open_backends"}
	gotResult := sortedMapKeys(cm)
	sort.Strings(wantResult)
	if strings.Join(gotResult, ",") != strings.Join(wantResult, ",") {
		t.Errorf("ConvergenceResult empty-identity key set = %v, want %v", gotResult, wantResult)
	}
}

func sortedMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ─── AC-ABI-005 — fail-open: no identity, audit still completes ───

func TestAuditCompletesWithoutBuildIdentity(t *testing.T) {
	// States 1-2 (+ the D-8 extension "unknown"): the binary carries no commit
	// metadata. Every one of the three states is judged on its own — a run
	// that reuses one state's result for the others measures nothing.
	for _, commit := range []string{"none", "", "unknown"} {
		t.Run("commit="+fmt.Sprintf("%q", commit), func(t *testing.T) {
			withVersionIdentity(t, fxVersion, commit)
			// A StatusBehind stub verdict must NOT leak a lag advisory when the
			// binary itself carries no commit metadata (REQ-ABI-005).
			withBinlagStub(t, behindStubVerdict())

			for _, ep := range auditEntryPoints() {
				root := ""
				if ep.name != "audit_multi" {
					root = reviewTreeFixture(t)
				}
				res, err := ep.run(t, root) // audit_multi: documented normal call, no project_root
				if err != nil {
					t.Fatalf("%s (commit=%q): handler returned Go error — fail-open broken: %v", ep.name, commit, err)
				}
				m := toolResultJSON(t, res)
				if _, present := m[keyBuildCommit]; present {
					t.Errorf("%s (commit=%q): %q present — an unusable commit must be OMITTED, not carried", ep.name, commit, keyBuildCommit)
				}
				if _, present := m[keyBuildLag]; present {
					t.Errorf("%s (commit=%q): %q present — no commit metadata, no identity-derived advisory", ep.name, commit, keyBuildLag)
				}
				// The verdict itself is untouched by the identity work.
				if ep.name == "audit_multi" {
					if m["overall_verdict"] != "pass" {
						t.Errorf("%s (commit=%q): overall_verdict = %v, want pass — identity work changed the verdict", ep.name, commit, m["overall_verdict"])
					}
				} else if m["verdict"] != "pass" {
					t.Errorf("%s (commit=%q): verdict = %v, want pass — identity work changed the verdict", ep.name, commit, m["verdict"])
				}
			}
		})
	}

	// State 3: neither the named tree nor the process working directory is a
	// git working tree — the real comparison runs (no stub) and stays silent.
	t.Run("no git tree anywhere", func(t *testing.T) {
		withVersionIdentity(t, fxVersion, fxCommitA)
		t.Chdir(t.TempDir()) // cwd is now non-git; t.Chdir restores it and forbids parallel tests

		for _, ep := range auditEntryPoints() {
			root := ""
			if ep.name != "audit_multi" {
				root = noGitTreeFixture(t)
			}
			res, err := ep.run(t, root) // audit_multi: empty root ⇒ cwd fallback ⇒ non-git
			if err != nil {
				t.Fatalf("%s (non-git): handler returned Go error — fail-open broken: %v", ep.name, err)
			}
			m := toolResultJSON(t, res)
			if _, present := m[keyBuildLag]; present {
				t.Errorf("%s (non-git): %q present — a non-git tree must not produce a lag advisory", ep.name, keyBuildLag)
			}
		}
	})
}

// ─── AC-ABI-006 — the lag advisory names both commits, on all three entry points ───

func TestAuditLagAdvisoryNamesBothCommits(t *testing.T) {
	withVersionIdentity(t, fxVersion, fxCommitA)
	withBinlagStub(t, behindStubVerdict())

	for _, ep := range auditEntryPoints() {
		res, err := ep.run(t, reviewTreeFixture(t))
		if err != nil {
			t.Fatalf("%s: handler error: %v", ep.name, err)
		}
		m := toolResultJSON(t, res)
		lag, ok := m[keyBuildLag].(string)
		if !ok || lag == "" {
			t.Fatalf("%s: %q empty/absent on a StatusBehind build — the lag advisory never fired", ep.name, keyBuildLag)
		}
		for _, want := range []string{binlag.Short(fxAncestor), binlag.Short(fxHead), binlag.RemedyCommand} {
			if !strings.Contains(lag, want) {
				t.Errorf("%s: build_lag %q does not name %q (must carry BOTH commits + the remedy)", ep.name, lag, want)
			}
		}
	}

	// Control: the same stub returning StatusFresh silences the advisory —
	// an implementation that always emits a string passes nothing here.
	withBinlagStub(t, freshStubVerdict())
	for _, ep := range auditEntryPoints() {
		res, err := ep.run(t, reviewTreeFixture(t))
		if err != nil {
			t.Fatalf("%s: handler error (control): %v", ep.name, err)
		}
		if m := toolResultJSON(t, res); m[keyBuildLag] != nil && m[keyBuildLag] != "" {
			t.Errorf("%s: build_lag = %v on a StatusFresh build — the advisory must be silent", ep.name, m[keyBuildLag])
		}
	}

	// D-3 clause: audit_multi with NO project_root still compares — via the
	// process working directory — and the fallback directory is used for the
	// COMPARISON ONLY (it must not reach the backend callers).
	stub := withBinlagStub(t, behindStubVerdict())
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	res, err := runAuditMultiEntry(t, "")
	if err != nil {
		t.Fatalf("audit_multi (no project_root): handler error: %v", err)
	}
	m := toolResultJSON(t, res)
	lag, ok := m[keyBuildLag].(string)
	if !ok || lag == "" {
		t.Fatalf("audit_multi (no project_root): %q empty — the empty-root path skipped the comparison (D-3)", keyBuildLag)
	}
	if stub.lastReq.Dir != cwd {
		t.Errorf("comparison ran against Dir %q, want the process working directory %q", stub.lastReq.Dir, cwd)
	}
}

// ─── AC-ABI-007 — exactly one comparison implementation, reached via the seam ───

func TestAuditLagUsesBinlagSeam(t *testing.T) {
	// Observation 1 — the stub: each entry point's single audit increments the
	// counter (the comparison is reached through binlag.Comparer, not skipped).
	withVersionIdentity(t, fxVersion, fxCommitA)
	for _, ep := range auditEntryPoints() {
		stub := withBinlagStub(t, freshStubVerdict())
		before := stub.count()
		root := ""
		if ep.name != "audit_multi" {
			root = reviewTreeFixture(t)
		}
		_, err := ep.run(t, root)
		if err != nil {
			t.Fatalf("%s: handler error: %v", ep.name, err)
		}
		if after := stub.count(); after <= before {
			t.Errorf("%s: one audit incremented the binlag counter %d → %d — the comparison never ran through the seam", ep.name, before, after)
		}
	}

	// Observation 2 — the exact-set source sweep: the hit set of the ancestry
	// primitives across internal/cli non-test sources equals the measured
	// baseline exactly (baseline 64bba61aa: 3 hits; additions AND removals
	// both fail — equal counts alone would not).
	want := map[string]bool{
		"graph_stamp.go:68":         true,
		"graph_stamp.go:131":        true,
		"mcp_review_material.go:95": true,
	}
	got := map[string]bool{}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("readdir package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for i, line := range strings.Split(string(b), "\n") {
			if strings.Contains(line, "merge-base") || strings.Contains(line, "is-ancestor") {
				got[fmt.Sprintf("%s:%d", name, i+1)] = true
			}
		}
	}
	for coord := range want {
		if !got[coord] {
			t.Errorf("sweep: baseline hit %s MISSING — the ancestry comparison moved or was reimplemented outside binlag", coord)
		}
	}
	for coord := range got {
		if !want[coord] {
			t.Errorf("sweep: NEW ancestry hit %s — a second comparison outside binlag.Evaluate (REQ-ABI-006 violation)", coord)
		}
	}
}
