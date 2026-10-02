package cli

// SPEC-CODEX-REVIEW-OWNERSHIP-001 M3 — the on-demand self-review tools
// (codex_review / glm_review), REQ-CRO-007..010 and the M3 slice of
// REQ-CRO-014. Acceptance criteria covered here: AC-007 (surface), AC-008 (card
// scope = the gate's resolver), AC-009 (uncommitted scope, untracked list,
// truncation), AC-010 (advisory + metadata on every path), AC-011 (model),
// AC-016 (no receipts, no required-gate conversion), and the AC-015 slice that
// concerns the tool catalogue rows (plan-audit round 3, I3-7).
//
// Observation discipline (acceptance.md §A): the assertions read the codex wire
// requests, the exact material posted to z.ai, the registered schemas, the
// receipt directory before and after, and the result metadata by VALUE. A
// verdict alone proves nothing — the stub reviewers answer whatever they like.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/hook"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

var selfReviewTools = []string{"codex_review", "glm_review"}

// --- AC-007: the surface ---

func TestSelfReview_ToolSurface(t *testing.T) {
	tools := selfReviewListTools(t)
	catalog := map[string]bool{}
	for _, d := range mcpcat.MoaiMCPTools() {
		catalog[d.Name] = d.WriteCapable
	}
	for _, name := range selfReviewTools {
		tool, ok := tools[name]
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		in, _ := tool["inputSchema"].(map[string]any)
		props, _ := in["properties"].(map[string]any)
		scope, _ := props["scope"].(map[string]any)
		var enum []string
		for _, v := range scope["enum"].([]any) {
			enum = append(enum, v.(string))
		}
		sort.Strings(enum)
		if !reflect.DeepEqual(enum, []string{"card", "uncommitted"}) {
			t.Errorf("%s scope enum = %v, want exactly [card uncommitted]", name, enum)
		}
		required := map[string]bool{}
		for _, r := range in["required"].([]any) {
			required[r.(string)] = true
		}
		if !required["scope"] {
			t.Errorf("%s: scope must be REQUIRED (a default would let a lane silently review the wrong change); required=%v", name, in["required"])
		}
		if _, ok := scope["default"]; ok {
			t.Errorf("%s: scope must carry no default", name)
		}
		for _, p := range []string{"project_root", "model"} {
			if _, ok := props[p]; !ok {
				t.Errorf("%s: input property %q not declared", name, p)
			}
		}
		if _, ok := props["focus"]; ok {
			t.Errorf("%s: a focus input is out of scope and must not be declared", name)
		}
		// project_root rides the fallback contract the existing tools share
		// (I3-7 ii): the pass-through variant would declare a different text.
		pr, _ := props["project_root"].(map[string]any)
		if got, _ := pr["description"].(string); got != projectRootDesc {
			t.Errorf("%s: project_root declared with the wrong contract text: %q", name, got)
		}
		out, _ := tool["outputSchema"].(map[string]any)
		outProps, _ := out["properties"].(map[string]any)
		for _, f := range []string{"advisory", "scope", "base", "backend", "tree", "truncated", "excluded_untracked", "verdict", "summary", "findings", "next_steps"} {
			if _, ok := outProps[f]; !ok {
				t.Errorf("%s: output schema does not declare %q (the embedded review fields must be flattened)", name, f)
			}
		}
		ann, _ := tool["annotations"].(map[string]any)
		if ro, _ := ann["readOnlyHint"].(bool); !ro {
			t.Errorf("%s: read-only hint must be true", name)
		}
		if wc, present := catalog[name]; !present || wc {
			t.Errorf("%s: catalog entry present=%v WriteCapable=%v, want present and false", name, present, wc)
		}
		desc, _ := tool["description"].(string)
		if !strings.Contains(strings.ToLower(desc), "advisory") {
			t.Errorf("%s: description must say the result is advisory: %q", name, desc)
		}
	}
}

// TestSelfReview_AuditToolSchemasUnchanged — AC-007 (e)(f): the four audit
// tools keep their names and their input/output schemas, compared BY VALUE with
// the snapshot taken from the pre-change server (plan.md §I M1). The snapshot
// is local evidence (.moai/reports is gitignored), so a checkout without it
// cannot run the comparison: that is reported as a skip — an unobserved
// comparison, never a pass.
func TestSelfReview_AuditToolSchemasUnchanged(t *testing.T) {
	snapPath := "../../.moai/reports/t1422/red/audit-tools-schema-pre.json"
	raw, err := os.ReadFile(snapPath)
	if err != nil {
		t.Skipf("audit-tool schema snapshot absent (%v): the comparison is unobserved here", err)
	}
	var snap []map[string]any
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if len(snap) != 4 {
		t.Fatalf("snapshot holds %d tools, want the four audit tools", len(snap))
	}
	tools := selfReviewListTools(t)
	for _, want := range snap {
		name, _ := want["name"].(string)
		got, ok := tools[name]
		if !ok {
			t.Errorf("audit tool %q disappeared from tools/list", name)
			continue
		}
		for _, key := range []string{"inputSchema", "outputSchema"} {
			if !reflect.DeepEqual(got[key], want[key]) {
				t.Errorf("%s %s differs from the pre-change snapshot", name, key)
			}
		}
	}
}

// TestSelfReview_AuditToolHoldersUnchanged — the audit-tool holder set (P1):
// codex_audit / glm_audit stay listed by plan-auditor and sync-auditor only, in
// both the local and the distributed copies. A regression line, green before
// and after this milestone.
func TestSelfReview_AuditToolHoldersUnchanged(t *testing.T) {
	for _, dir := range []string{
		"../../.claude/agents/moai",
		"../../internal/template/templates/.claude/agents/moai",
	} {
		for _, tool := range []string{"mcp__moai__codex_audit", "mcp__moai__glm_audit"} {
			var holders []string
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				if strings.Contains(srReadFile(t, filepath.Join(dir, e.Name())), tool) {
					holders = append(holders, strings.TrimSuffix(e.Name(), ".md"))
				}
			}
			sort.Strings(holders)
			if !reflect.DeepEqual(holders, []string{"plan-auditor", "sync-auditor"}) {
				t.Errorf("%s holders in %s = %v, want exactly [plan-auditor sync-auditor]", tool, dir, holders)
			}
		}
	}
}

func TestSelfReview_ScopeIsRequiredAndValidated(t *testing.T) {
	f := newSelfReviewFixture(t)
	for _, tool := range selfReviewTools {
		for name, args := range map[string]map[string]any{
			"missing":   {"project_root": f.canon},
			"invalid":   {"project_root": f.canon, "scope": "branch"},
			"empty":     {"project_root": f.canon, "scope": ""},
			"baseBr":    {"project_root": f.canon, "scope": "baseBranch"},
			"badRoot":   {"project_root": filepath.Join(f.canon, "does-not-exist"), "scope": "uncommitted"},
			"noMoaiDir": {"project_root": t.TempDir(), "scope": "uncommitted"},
		} {
			t.Run(tool+"/"+name, func(t *testing.T) {
				rig := newSelfReviewRig(t, strings.TrimSuffix(tool, "_review"), "pass")
				res, _ := selfReviewCall(t, tool, args)
				if !res.IsError {
					t.Errorf("a call with %v must be rejected as a tool error, got %+v", args, res)
				}
				if rig.calls() != 0 {
					t.Errorf("a rejected call reached the reviewer %d time(s)", rig.calls())
				}
			})
		}
	}
}

// --- AC-008: card scope is the gate's resolver ---

func TestSelfReview_CodexCardScopeRequestIsTheGateRequest(t *testing.T) {
	f := newSelfReviewFixture(t)
	// The server's own environment names the PRIMARY checkout: the tool must
	// follow project_root, never that default.
	t.Setenv("CLAUDE_PROJECT_DIR", f.primary)
	sess := withCodexSession(t, codexSessionScript(realCleanReview))

	_, m := selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": f.link})
	thread, target := codexWire(t, sess)

	if cwd, _ := thread["cwd"].(string); cwd != f.canon {
		t.Errorf("thread/start cwd = %q, want the canonical card worktree %q", cwd, f.canon)
	}
	if typ, _ := target["type"].(string); typ != codexTargetBaseBranch {
		t.Errorf("target.type = %q, want %q", typ, codexTargetBaseBranch)
	}
	if br, _ := target["branch"].(string); br != f.base || br == f.devTip {
		t.Errorf("target.branch = %q, want the recomputed merge base %q (not the develop tip %q)", br, f.base, f.devTip)
	}
	want, _ := reviewRequestParams(reviewScopeResolver(f.canon))["target"].(map[string]any)
	if !reflect.DeepEqual(target, want) {
		t.Errorf("review target %v != the gate's request for the same tree %v", target, want)
	}
	if raw := strings.Join(sess.sent, ""); strings.Contains(raw, f.primary) {
		t.Errorf("the request references the primary-role tree %q: %s", f.primary, raw)
	}
	if got, _ := m["base"].(string); got != f.base {
		t.Errorf("result base = %q, want %q", got, f.base)
	}
}

func TestSelfReview_GLMCardMaterialIsTheCardDiff(t *testing.T) {
	f := newSelfReviewFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", f.primary)
	rig := newSelfReviewRig(t, "glm", "pass")

	_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": f.link})

	want := selfReviewIndepDiff(t, f.card, f.base)
	for _, must := range []string{"card_a.go", "card_b.go", "card_g.go"} {
		if !strings.Contains(want, must) {
			t.Fatalf("fixture error: the independent card diff lacks %s", must)
		}
	}
	sent := rig.glm.sent(t)
	if got := sent.Messages[0].Content; got != glmAuditUserPrompt("", truncateDiff(want)) {
		t.Errorf("the material posted to z.ai is not the card diff.\n got: %q\nwant: %q", got, glmAuditUserPrompt("", truncateDiff(want)))
	}
	for _, bad := range []string{"foreign WIP", "untracked_c", "fixture-tracked.json", ".moai/state"} {
		if strings.Contains(sent.Messages[0].Content, bad) {
			t.Errorf("the material must not contain %q (files F, C, D belong outside the card diff)", bad)
		}
	}
	if got, _ := m["base"].(string); got != f.base {
		t.Errorf("result base = %q, want %q", got, f.base)
	}
}

// TestSelfReview_CardBaseMovesWhenTheCardAbsorbsDevelop — the merge base is
// recomputed on every call, never pinned (REQ-CRO-008, gitflow-lane-protocol
// §8): after the card absorbs develop's new commit both backends measure from
// the NEW base.
func TestSelfReview_CardBaseMovesWhenTheCardAbsorbsDevelop(t *testing.T) {
	f := newSelfReviewFixture(t)
	codex := newSelfReviewRig(t, "codex", "pass")
	glm := newSelfReviewRig(t, "glm", "pass")

	call := func() (codexBranch, glmContent string) {
		codex.codex.sent = nil
		selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": f.canon})
		_, target := codexWire(t, codex.codex)
		selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": f.canon})
		codexBranch, _ = target["branch"].(string)
		return codexBranch, glm.glm.sent(t).Messages[0].Content
	}
	before, beforeMaterial := call()
	if before != f.base {
		t.Fatalf("before absorption branch = %q, want %q", before, f.base)
	}

	cardScopeGit(t, f.card, "add", "-A")
	cardScopeGit(t, f.card, "commit", "-q", "-m", "wip: card state before absorbing develop")
	cardScopeGit(t, f.card, "merge", "-q", "--no-edit", "develop")
	newBase := cardScopeGit(t, f.card, "merge-base", "develop", "HEAD")
	if newBase == f.base {
		t.Fatalf("fixture error: absorbing develop did not move the merge base (%s)", newBase)
	}

	after, afterMaterial := call()
	if after != newBase {
		t.Errorf("after absorption the codex branch = %q, want the NEW merge base %q (the old value %q must not linger)", after, newBase, f.base)
	}
	wantMaterial := glmAuditUserPrompt("", truncateDiff(selfReviewIndepDiff(t, f.card, newBase)))
	if afterMaterial != wantMaterial {
		t.Errorf("after absorption the GLM material is not measured from the new base %s", newBase)
	}
	if afterMaterial == beforeMaterial {
		t.Error("the GLM material did not change when the base moved")
	}
}

// TestSelfReview_ResolverIsTheOnlyDiscriminator — a resolver stub returning a
// sentinel MergeBase (a real ancestor that is NOT the merge base) must reach
// both backends and the result: the tools carry no discriminator and no base
// computation of their own.
func TestSelfReview_ResolverIsTheOnlyDiscriminator(t *testing.T) {
	f := newSelfReviewFixture(t)
	old := reviewScopeResolver
	reviewScopeResolver = func(dir string) reviewScope {
		return reviewScope{Class: reviewScopeCard, Basis: "stub", Dir: dir, Branch: "WT-stub", MergeBase: f.c0}
	}
	t.Cleanup(func() { reviewScopeResolver = old })
	if f.c0 == f.base {
		t.Fatal("fixture error: the sentinel must differ from the real merge base")
	}

	codex := newSelfReviewRig(t, "codex", "pass")
	_, cm := selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": f.canon})
	_, target := codexWire(t, codex.codex)
	if br, _ := target["branch"].(string); br != f.c0 {
		t.Errorf("codex branch = %q, want the resolver's sentinel %q", br, f.c0)
	}
	if got, _ := cm["base"].(string); got != f.c0 {
		t.Errorf("codex result base = %q, want the sentinel %q", got, f.c0)
	}

	glm := newSelfReviewRig(t, "glm", "pass")
	_, gm := selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": f.canon})
	want := glmAuditUserPrompt("", truncateDiff(selfReviewIndepDiff(t, f.card, f.c0)))
	if got := glm.glm.sent(t).Messages[0].Content; got != want {
		t.Error("the GLM material is not `git diff <sentinel>`: the tool computed its own base")
	}
	if got, _ := gm["base"].(string); got != f.c0 {
		t.Errorf("glm result base = %q, want the sentinel %q", got, f.c0)
	}
}

// TestSelfReview_NonCardTreeReturnsInconclusiveWithoutReviewing — AC-008 (v).
func TestSelfReview_NonCardTreeReturnsInconclusiveWithoutReviewing(t *testing.T) {
	nobase := t.TempDir()
	cardScopeGit(t, nobase, "init", "-q", "-b", "WT-fixture-nobase")
	writeCardFile(t, nobase, "x.go", "package x\n")
	writeCardFile(t, nobase, ".moai/state/x.json", "{}\n")
	cardScopeGit(t, nobase, "add", "-A")
	cardScopeGit(t, nobase, "commit", "-q", "-m", "init")
	writeCardFile(t, nobase, "x.go", "package x\n\n// dirty\n")

	trees := map[string]struct{ dir, cause string }{
		"develop":  {newSelfReviewPlainTree(t, "develop", true), "no card branch"},
		"detached": {newSelfReviewDetachedTree(t), "detached"},
		"nongit":   {newSelfReviewNonGitTree(t), "unreadable"},
		"nobase":   {nobase, "merge base"},
	}
	for name, tc := range trees {
		for _, backend := range []string{"codex", "glm"} {
			t.Run(name+"/"+backend, func(t *testing.T) {
				rig := newSelfReviewRig(t, backend, "fail") // would block if it were reached
				_, m := selfReviewCall(t, rig.tool, map[string]any{"scope": "card", "project_root": tc.dir})
				if v, _ := m["verdict"].(string); v != VerdictInconclusive {
					t.Errorf("verdict = %q, want inconclusive", v)
				}
				summary, _ := m["summary"].(string)
				if !strings.Contains(summary, "card worktree") || !strings.Contains(summary, tc.cause) {
					t.Errorf("summary %q must name the cause (a card worktree is required; %q)", summary, tc.cause)
				}
				if rig.calls() != 0 {
					t.Errorf("the reviewer was reached %d time(s); a non-card tree must not be reviewed in the card's place", rig.calls())
				}
			})
		}
	}
}

// --- AC-009: uncommitted scope, untracked list, truncation ---

func TestSelfReview_CodexUncommittedRequestShapeIsTheGatesTreeRequest(t *testing.T) {
	f := newSelfReviewFixture(t)
	plain := newSelfReviewPlainTree(t, "develop", true)
	plainCanon, err := filepath.EvalSymlinks(plain)
	if err != nil {
		t.Fatal(err)
	}

	// The gate's own request for the same plain tree is the reference.
	gateSess := withCodexSession(t, codexSessionScript(realCleanReview))
	if _, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "ref", CWD: plainCanon}, true, plainCanon); err != nil {
		t.Fatalf("gate: %v", err)
	}
	if len(gateSess.sent) < 3 {
		t.Fatalf("fixture error: the gate sent %d requests", len(gateSess.sent))
	}
	gateThread, gateReview := gateSess.sent[1], gateSess.sent[2]

	toolSess := withCodexSession(t, codexSessionScript(realCleanReview))
	selfReviewCall(t, "codex_review", map[string]any{"scope": "uncommitted", "project_root": plain})
	if len(toolSess.sent) < 3 || toolSess.sent[1] != gateThread || toolSess.sent[2] != gateReview {
		t.Errorf("the tool's tree request is not shape-identical to the gate's.\n tool: %v\n gate: %v", toolSess.sent, []string{gateThread, gateReview})
	}

	// On a WT- card tree the uncommitted request is STILL the tree request, not
	// a baseBranch request.
	cardSess := withCodexSession(t, codexSessionScript(realCleanReview))
	selfReviewCall(t, "codex_review", map[string]any{"scope": "uncommitted", "project_root": f.link})
	thread, target := codexWire(t, cardSess)
	if typ, _ := target["type"].(string); typ != codexTargetUncommitted {
		t.Errorf("on a card tree scope=uncommitted sent target %v, want %q", target, codexTargetUncommitted)
	}
	if _, has := target["branch"]; has {
		t.Errorf("an uncommitted request must carry no branch: %v", target)
	}
	if cwd, _ := thread["cwd"].(string); cwd != f.canon {
		t.Errorf("cwd = %q, want %q", cwd, f.canon)
	}
}

func TestSelfReview_GLMUncommittedMaterialIncludesStagedOnlyChanges(t *testing.T) {
	f := newSelfReviewFixture(t)
	rig := newSelfReviewRig(t, "glm", "pass")
	_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "uncommitted", "project_root": f.link})

	want := selfReviewIndepDiff(t, f.card, "HEAD")
	if !strings.Contains(want, "staged only") {
		t.Fatalf("fixture error: the independent diff HEAD lacks the staged-only change G")
	}
	got := rig.glm.sent(t).Messages[0].Content
	if got != glmAuditUserPrompt("", truncateDiff(want)) {
		t.Errorf("the material is not `git diff HEAD` (staged-only change G must be included).\n got: %q", got)
	}
	if strings.Contains(got, "card_a.go") {
		t.Error("committed card work is not an uncommitted change and must not be in the material")
	}
	if base, present := m["base"]; !present || base != "" {
		t.Errorf("uncommitted base = %v (present=%v), want the empty string", base, present)
	}
}

func TestSelfReview_GLMExcludedUntrackedForBothScopes(t *testing.T) {
	f := newSelfReviewFixture(t)
	for _, scope := range []string{"card", "uncommitted"} {
		t.Run(scope, func(t *testing.T) {
			newSelfReviewRig(t, "glm", "pass")
			_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": scope, "project_root": f.canon})
			var got []string
			for _, v := range m["excluded_untracked"].([]any) {
				got = append(got, v.(string))
			}
			want := []string{"untracked_c.go", "untracked_c2.go"} // the runtime-prefix untracked file is NOT listed
			if !reflect.DeepEqual(got, want) {
				t.Errorf("excluded_untracked = %v, want %v", got, want)
			}
		})
	}
}

func TestSelfReview_GLMTruncatedBoundary(t *testing.T) {
	dir := newSelfReviewPlainTree(t, "develop", false)
	writeCardFile(t, dir, "big.txt", "seed\n")
	cardScopeGit(t, dir, "add", "big.txt")
	cardScopeGit(t, dir, "commit", "-q", "-m", "big")
	diffAt := func(n int) string {
		writeCardFile(t, dir, "big.txt", "seed\n"+strings.Repeat("x", n)+"\n")
		return selfReviewIndepDiff(t, dir, "HEAD")
	}
	n := 1000 + (reviewDiffMaxBytes - len(diffAt(1000)))

	for _, tc := range []struct {
		name      string
		delta     int
		truncated bool
	}{
		{"one-under-the-cap", -1, false},
		{"exactly-the-cap", 0, false},
		{"one-over-the-cap", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := diffAt(n + tc.delta)
			if wantLen := reviewDiffMaxBytes + tc.delta; len(want) != wantLen {
				t.Fatalf("fixture error: diff length %d, want %d", len(want), wantLen)
			}
			rig := newSelfReviewRig(t, "glm", "pass")
			_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "uncommitted", "project_root": dir})
			if got, _ := m["truncated"].(bool); got != tc.truncated {
				t.Errorf("truncated = %v, want %v for a %d-byte diff", got, tc.truncated, len(want))
			}
			content := rig.glm.sent(t).Messages[0].Content
			if content != glmAuditUserPrompt("", truncateDiff(want)) {
				t.Error("the posted material is not the (possibly truncated) diff")
			}
			if tc.truncated && !strings.Contains(content, "[diff truncated") {
				t.Error("a truncated material must carry the visible truncation marker")
			}
		})
	}
}

func TestSelfReview_CodexNeverTruncatesAndListsNoUntracked(t *testing.T) {
	f := newSelfReviewFixture(t)
	for _, scope := range []string{"card", "uncommitted"} {
		newSelfReviewRig(t, "codex", "pass")
		_, m := selfReviewCall(t, "codex_review", map[string]any{"scope": scope, "project_root": f.canon})
		if tr, ok := m["truncated"].(bool); !ok || tr {
			t.Errorf("%s: codex truncated = %v (ok=%v), want false", scope, m["truncated"], ok)
		}
		if ex, ok := m["excluded_untracked"].([]any); !ok || len(ex) != 0 {
			t.Errorf("%s: codex excluded_untracked = %v, want an empty array", scope, m["excluded_untracked"])
		}
	}
}

// --- AC-010: advisory + metadata by value on every path ---

type selfReviewMetaWant struct{ scope, backend, base, tree string }

func requireSelfReviewMeta(t *testing.T, m map[string]any, w selfReviewMetaWant) {
	t.Helper()
	if m == nil {
		t.Fatal("the result carries no structured content")
	}
	if adv, ok := m["advisory"].(bool); !ok || !adv {
		t.Errorf("advisory = %v, want true", m["advisory"])
	}
	for key, want := range map[string]string{"scope": w.scope, "backend": w.backend, "base": w.base, "tree": w.tree} {
		got, present := m[key]
		if !present {
			t.Errorf("%s is absent from the result", key)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if _, ok := m["truncated"].(bool); !ok {
		t.Errorf("truncated must be present as a boolean, got %v", m["truncated"])
	}
	if _, ok := m["excluded_untracked"].([]any); !ok {
		t.Errorf("excluded_untracked must be present as an array, got %v", m["excluded_untracked"])
	}
	if _, has := m["gate_unmet"]; has {
		t.Errorf("a self-review result never carries gate_unmet: %v", m["gate_unmet"])
	}
}

func TestSelfReview_AdvisoryMetadataOnEveryPath(t *testing.T) {
	f := newSelfReviewFixture(t)
	emptyCard := f.newEmptyCard(t)
	dirty := newSelfReviewPlainTree(t, "develop", true)
	clean := newSelfReviewPlainTree(t, "develop", false)
	canon := func(p string) string {
		c, err := filepath.EvalSymlinks(p)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	link := func(p string) string {
		l := filepath.Join(t.TempDir(), "lnk")
		if err := os.Symlink(p, l); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		return l
	}

	type path struct {
		name, scope, root, wantTree, wantBase string
		reviewed                              bool // the reviewer is reached
	}
	paths := []path{
		{"p1-card", "card", f.link, f.canon, f.base, true},
		{"p2-card-uncommitted", "uncommitted", f.link, f.canon, "", true},
		{"p3-develop-uncommitted", "uncommitted", link(dirty), canon(dirty), "", true},
		{"p4-noncard-early-return", "card", link(dirty), canon(dirty), "", false},
		{"p5a-empty-develop", "uncommitted", link(clean), canon(clean), "", false},
		{"p5b-empty-card", "card", link(emptyCard), canon(emptyCard), "", false},
	}
	for _, backend := range []string{"codex", "glm"} {
		for _, verdict := range []string{"pass", "fail", "inconclusive"} {
			for _, p := range paths {
				t.Run(backend+"/"+verdict+"/"+p.name, func(t *testing.T) {
					rig := newSelfReviewRig(t, backend, verdict)
					_, m := selfReviewCall(t, rig.tool, map[string]any{"scope": p.scope, "project_root": p.root})
					requireSelfReviewMeta(t, m, selfReviewMetaWant{p.scope, backend, p.wantBase, p.wantTree})
					wantVerdict := verdict
					if !p.reviewed {
						wantVerdict = VerdictInconclusive
						if rig.calls() != 0 {
							t.Errorf("the reviewer was reached %d time(s) on a path that must not reach it", rig.calls())
						}
					}
					if got, _ := m["verdict"].(string); got != wantVerdict {
						t.Errorf("verdict = %q, want %q", got, wantVerdict)
					}
				})
			}
		}
	}
}

// TestSelfReview_EmptyMaterial — Q14 and plan-audit I3-2: an empty card diff or
// an empty uncommitted diff calls neither backend. GLM has no filesystem, so
// untracked-only trees are empty material for it (and the files are still
// REPORTED); codex reads the tree itself and is therefore still asked.
func TestSelfReview_EmptyMaterial(t *testing.T) {
	f := newSelfReviewFixture(t)
	emptyCard := f.newEmptyCard(t)

	t.Run("glm/clean", func(t *testing.T) {
		rig := newSelfReviewRig(t, "glm", "fail")
		_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": emptyCard})
		if v, _ := m["verdict"].(string); v != VerdictInconclusive || rig.calls() != 0 {
			t.Errorf("verdict=%v calls=%d, want inconclusive and zero HTTP calls", m["verdict"], rig.calls())
		}
		if s, _ := m["summary"].(string); !strings.Contains(s, "no change") {
			t.Errorf("summary %q must say there is no change to review", s)
		}
	})
	t.Run("codex/clean", func(t *testing.T) {
		rig := newSelfReviewRig(t, "codex", "fail")
		_, m := selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": emptyCard})
		if v, _ := m["verdict"].(string); v != VerdictInconclusive || rig.calls() != 0 {
			t.Errorf("verdict=%v calls=%d, want inconclusive and zero RPC requests", m["verdict"], rig.calls())
		}
	})

	writeCardFile(t, emptyCard, "only_untracked.go", "package main\n\n// the only change is untracked\n")
	t.Run("glm/untracked-only", func(t *testing.T) {
		rig := newSelfReviewRig(t, "glm", "fail")
		_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": emptyCard})
		if v, _ := m["verdict"].(string); v != VerdictInconclusive || rig.calls() != 0 {
			t.Errorf("verdict=%v calls=%d: GLM cannot see untracked files, so an untracked-only tree is empty material (zero HTTP calls)", m["verdict"], rig.calls())
		}
		var got []string
		for _, v := range m["excluded_untracked"].([]any) {
			got = append(got, v.(string))
		}
		if !reflect.DeepEqual(got, []string{"only_untracked.go"}) {
			t.Errorf("excluded_untracked = %v, want the untracked file still reported", got)
		}
		// Sync-audit F2: the summary must not tell a reader of `summary` alone that
		// the tree is clean — changes exist, GLM just cannot be sent them.
		s, _ := m["summary"].(string)
		if strings.Contains(s, "no change to review") {
			t.Errorf("summary %q claims there is no change to review, but untracked changes exist", s)
		}
		if !strings.Contains(s, "untracked") || !strings.Contains(s, "excluded_untracked") {
			t.Errorf("summary %q must say the only changes are untracked files and name excluded_untracked", s)
		}
	})
	t.Run("codex/untracked-only", func(t *testing.T) {
		rig := newSelfReviewRig(t, "codex", "fail")
		_, m := selfReviewCall(t, "codex_review", map[string]any{"scope": "uncommitted", "project_root": emptyCard})
		if rig.calls() == 0 {
			t.Errorf("codex reads untracked files itself, so an untracked-only tree must still be reviewed (verdict=%v)", m["verdict"])
		}
	})
}

// TestSelfReview_UntrackedListIsNULSeparatedSoNonASCIIPathsSurvive — sync-audit
// F1: with git's default core.quotepath=true a plain `ls-files` prints non-ASCII
// names octal-quoted, which would reach excluded_untracked unreadable and slip a
// runtime-managed non-ASCII path past the prefix filter. The tree's own config
// sets quotepath explicitly so the test does not depend on the machine's global
// setting.
func TestSelfReview_UntrackedListIsNULSeparatedSoNonASCIIPathsSurvive(t *testing.T) {
	quotedTree := func(t *testing.T) string {
		t.Helper()
		dir := newSelfReviewPlainTree(t, "develop", false)
		cardScopeGit(t, dir, "config", "core.quotepath", "true")
		return dir
	}

	t.Run("excluded_untracked-names-are-readable", func(t *testing.T) {
		dir := quotedTree(t)
		writeCardFile(t, dir, "새파일.txt", "x\n")
		writeCardFile(t, dir, "plain.txt", "x\n")
		writeCardFile(t, dir, ".moai/reports/한글/보고.md", "x\n")
		// Premise: the non-NUL listing really is octal-quoted under this config.
		if raw := cardScopeGit(t, dir, "ls-files", "--others", "--exclude-standard"); !strings.Contains(raw, `\355`) {
			t.Fatalf("fixture premise broken: expected an octal-quoted path in %q", raw)
		}
		newSelfReviewRig(t, "glm", "pass")
		_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "uncommitted", "project_root": dir})
		var got []string
		for _, v := range m["excluded_untracked"].([]any) {
			got = append(got, v.(string))
		}
		sort.Strings(got)
		if want := []string{"plain.txt", "새파일.txt"}; !reflect.DeepEqual(got, want) {
			t.Errorf("excluded_untracked = %q, want %q (real names, runtime-managed Korean path filtered)", got, want)
		}
	})

	t.Run("runtime-managed-non-ASCII-path-is-filtered-for-codex-too", func(t *testing.T) {
		dir := quotedTree(t)
		writeCardFile(t, dir, ".moai/reports/한글/보고.md", "x\n")
		rig := newSelfReviewRig(t, "codex", "fail")
		_, m := selfReviewCall(t, "codex_review", map[string]any{"scope": "uncommitted", "project_root": dir})
		if rig.calls() != 0 {
			t.Errorf("a runtime-managed path is not a change to review; codex was reached %d time(s) (verdict=%v)", rig.calls(), m["verdict"])
		}
	})
}

// TestSelfReview_MaterialIsAPlainUnifiedDiff — sync-audit F3: a developer's
// diff.external, diff textconv or forced colour must not change the bytes sent
// to the reviewer. Each variant is configured in the tree's own config.
func TestSelfReview_MaterialIsAPlainUnifiedDiff(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the external-diff and textconv doubles are shell commands")
	}
	cases := []struct {
		name   string
		setup  func(t *testing.T, dir string)
		banned []string
	}{
		{"external-diff", func(t *testing.T, dir string) {
			cardScopeGit(t, dir, "config", "diff.external", "echo EXTERNAL-DIFF-MARKER")
		}, []string{"EXTERNAL-DIFF-MARKER"}},
		{"colour", func(t *testing.T, dir string) {
			cardScopeGit(t, dir, "config", "color.ui", "always")
		}, []string{"\x1b"}},
		{"textconv", func(t *testing.T, dir string) {
			writeCardFile(t, dir, ".gitattributes", "*.go diff=sr\n")
			cardScopeGit(t, dir, "config", "diff.sr.textconv", "echo TEXTCONV-MARKER")
		}, []string{"TEXTCONV-MARKER"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := newSelfReviewPlainTree(t, "develop", true)
			tc.setup(t, dir)
			rig := newSelfReviewRig(t, "glm", "pass")
			selfReviewCall(t, "glm_review", map[string]any{"scope": "uncommitted", "project_root": dir})
			if rig.calls() != 1 {
				t.Fatalf("the reviewer was reached %d time(s), want 1 (the real change must survive the config)", rig.calls())
			}
			content := rig.glm.sent(t).Messages[0].Content
			if !strings.Contains(content, "diff --git") || !strings.Contains(content, "+// uncommitted change") {
				t.Errorf("the posted material is not the plain unified diff of the change:\n%s", content)
			}
			for _, b := range tc.banned {
				if strings.Contains(content, b) {
					t.Errorf("the posted material carries %q — it is not a plain unified diff", b)
				}
			}
		})
	}
}

// TestSelfReview_ReviewBudgetAndCallerCancellationAreHonoured — sync-audit F4:
// both legs return promptly and fail open when the review budget elapses and
// when the caller's context is cancelled. The reviewer doubles block until
// their context ends, so a leg that drops the bound (or, for GLM, the
// cancellation forwarding) cannot return.
func TestSelfReview_ReviewBudgetAndCallerCancellationAreHonoured(t *testing.T) {
	dir := newSelfReviewPlainTree(t, "develop", true)
	for _, backend := range []string{"codex", "glm"} {
		for _, mode := range []string{"budget-elapses", "caller-cancels"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				b := newSelfReviewBlockingReviewer(t)
				handler := handleCodexReview
				if backend == "codex" {
					withCodexSession(t, nil)
					codexSession = selfReviewBlockingCodexSession{b}
				} else {
					handler = handleGLMReview
					withGLMSeams(t, "stub-key", selfReviewBlockingGLMDoer{b})
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if mode == "budget-elapses" {
					withReviewBudget(t, 150*time.Millisecond)
				} else {
					go func() {
						<-b.started
						cancel()
					}()
				}

				req := mcp.CallToolRequest{}
				req.Params.Name = backend + "_review"
				req.Params.Arguments = map[string]any{"scope": "uncommitted", "project_root": dir}
				done := make(chan *mcp.CallToolResult, 1)
				start := time.Now()
				go func() {
					res, _ := handler(ctx, req)
					done <- res
				}()
				var res *mcp.CallToolResult
				select {
				case res = <-done:
				case <-time.After(10 * time.Second):
					t.Fatalf("%s did not return within 10s of a %s — the review is unbounded", req.Params.Name, mode)
				}
				if took := time.Since(start); took > 5*time.Second {
					t.Errorf("returned after %s, want a prompt return", took)
				}
				if !b.byCtx.Load() {
					t.Error("the reviewer double was not released by its context")
				}
				raw, err := json.Marshal(res.StructuredContent)
				if err != nil {
					t.Fatal(err)
				}
				var m map[string]any
				if err := json.Unmarshal(raw, &m); err != nil {
					t.Fatalf("decode %s: %v", raw, err)
				}
				if v, _ := m["verdict"].(string); v != VerdictInconclusive {
					t.Errorf("verdict = %q, want inconclusive (fail-open)", v)
				}
				if adv, _ := m["advisory"].(bool); !adv {
					t.Errorf("advisory = %v, want true", m["advisory"])
				}
			})
		}
	}
}

func TestSelfReview_ReviewerErrorsFailOpen(t *testing.T) {
	f := newSelfReviewFixture(t)
	t.Run("codex-session-start-error", func(t *testing.T) {
		withCodexSession(t, codexSessionScript(realCleanReview))
		withCodexSessionStartErr(t, errors.New("spawn failed"))
		_, m := selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": f.canon})
		requireSelfReviewMeta(t, m, selfReviewMetaWant{"card", "codex", f.base, f.canon})
		if v, _ := m["verdict"].(string); v != VerdictInconclusive {
			t.Errorf("verdict = %q, want inconclusive", v)
		}
	})
	t.Run("glm-http-500", func(t *testing.T) {
		rig := newSelfReviewRig(t, "glm", "pass")
		rig.glm.status = 500
		_, m := selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": f.canon})
		requireSelfReviewMeta(t, m, selfReviewMetaWant{"card", "glm", f.base, f.canon})
		if v, _ := m["verdict"].(string); v != VerdictInconclusive {
			t.Errorf("verdict = %q, want inconclusive", v)
		}
	})
}

// TestSelfReview_ProjectRootOmittedFallsBackLikeTheOtherTools — I3-7 (ii): an
// absent project_root resolves through CLAUDE_PROJECT_DIR exactly as
// codex_audit / glm_audit do. The pass-through variant would hand the reviewer
// no root at all.
func TestSelfReview_ProjectRootOmittedFallsBackLikeTheOtherTools(t *testing.T) {
	f := newSelfReviewFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", f.canon)
	for _, backend := range []string{"codex", "glm"} {
		rig := newSelfReviewRig(t, backend, "pass")
		_, m := selfReviewCall(t, rig.tool, map[string]any{"scope": "uncommitted"})
		if got, _ := m["tree"].(string); got != f.canon {
			t.Errorf("%s: tree = %q with project_root omitted, want the CLAUDE_PROJECT_DIR fallback %q", backend, got, f.canon)
		}
		if rig.calls() == 0 {
			t.Errorf("%s: the fallback root was not reviewed", backend)
		}
		if rig.codex != nil {
			thread, _ := codexWire(t, rig.codex)
			if cwd, _ := thread["cwd"].(string); cwd != f.canon {
				t.Errorf("codex cwd = %q, want %q", cwd, f.canon)
			}
		}
	}
}

// --- AC-011: model ---

func TestSelfReview_ModelIsCallerInputOrBackendDefaultNeverAnAuditPin(t *testing.T) {
	f := newSelfReviewFixture(t)
	pins := "workflow:\n  audit:\n    codex:\n      model: gpt-pinned-audit\n      effort: high\n    glm:\n      model: glm-pinned-audit\n      effort: high\n"
	writeCardFile(t, f.card, ".moai/config/sections/workflow.yaml", pins)

	t.Run("no-model", func(t *testing.T) {
		codex := newSelfReviewRig(t, "codex", "pass")
		selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": f.canon})
		thread, _ := codexWire(t, codex.codex)
		if m, present := thread["model"]; present {
			t.Errorf("codex thread/start carries model %v; without a model input the pin-free gate driver sends none", m)
		}
		glm := newSelfReviewRig(t, "glm", "pass")
		selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": f.canon})
		if got := glm.glm.sent(t).Model; got != glmTaskDefaultModel || got == "glm-pinned-audit" {
			t.Errorf("glm model = %q, want the delegation default %q (the audit pin must not apply)", got, glmTaskDefaultModel)
		}
	})
	t.Run("explicit-model", func(t *testing.T) {
		codex := newSelfReviewRig(t, "codex", "pass")
		selfReviewCall(t, "codex_review", map[string]any{"scope": "card", "project_root": f.canon, "model": "gpt-X"})
		thread, _ := codexWire(t, codex.codex)
		if m, _ := thread["model"].(string); m != "gpt-X" {
			t.Errorf("codex model = %q, want gpt-X", m)
		}
		glm := newSelfReviewRig(t, "glm", "pass")
		selfReviewCall(t, "glm_review", map[string]any{"scope": "card", "project_root": f.canon, "model": "glm-X"})
		if got := glm.glm.sent(t).Model; got != "glm-X" {
			t.Errorf("glm model = %q, want glm-X", got)
		}
	})
}

// --- AC-016: no receipts, no required-gate conversion ---

func receiptNames(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "state", "audit-receipts"))
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func TestSelfReview_NeverMintsReceiptsAndIgnoresTheRequiredGate(t *testing.T) {
	f := newSelfReviewFixture(t)
	writeCodexAuditGate(t, f.card, "required")

	// Positive control: the AUDIT tool under the same required gate still turns an
	// absent codex into fail + gate_unmet and files a receipt — so an unchanged
	// receipt directory below is a measurement, not an empty sweep.
	withCodexSession(t, codexSessionScript(realCleanReview))
	withCodexLookPath(t, func(string) (string, error) { return "", os.ErrNotExist })
	audit := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": f.canon}))
	if v, _ := audit["verdict"].(string); v != "fail" {
		t.Fatalf("control: codex_audit verdict = %q, want fail under a required gate", v)
	}
	if g, _ := audit["gate_unmet"].(string); g == "" {
		t.Fatal("control: codex_audit must carry a non-empty gate_unmet")
	}
	before := receiptNames(t, f.card)
	if len(before) == 0 {
		t.Fatal("control: codex_audit filed no receipt, so the receipt directory is not being observed")
	}

	for _, backend := range []string{"codex", "glm"} {
		for _, verdict := range []string{"pass", "fail", "inconclusive"} {
			for _, scope := range []string{"card", "uncommitted"} {
				rig := newSelfReviewRig(t, backend, verdict)
				_, m := selfReviewCall(t, rig.tool, map[string]any{"scope": scope, "project_root": f.canon})
				if _, has := m["gate_unmet"]; has {
					t.Errorf("%s/%s/%s: gate_unmet present: %v", backend, verdict, scope, m["gate_unmet"])
				}
				if v, _ := m["verdict"].(string); verdict == "inconclusive" && v != VerdictInconclusive {
					t.Errorf("%s/%s: an unavailable reviewer under a required gate must stay inconclusive, got %q", backend, scope, v)
				}
				if after := receiptNames(t, f.card); !reflect.DeepEqual(after, before) {
					t.Errorf("%s/%s/%s: the receipt directory changed: %v -> %v", backend, verdict, scope, before, after)
				}
			}
		}
	}
}

// TestSelfReview_SourceReferencesNoAuditMachinery — the static half of AC-016
// (c), read with its positive controls, plus the heartbeat-free decision
// (Q13): the self-review file calls neither the receipt writer, the
// required-gate conversion, nor the progress notifier.
func TestSelfReview_SourceReferencesNoAuditMachinery(t *testing.T) {
	src := srReadFile(t, "mcp_selfreview.go")
	codex := srReadFile(t, "mcp_codex.go")
	for _, tok := range []string{"recordAuditReceipt", "applyGateUnmet", "auditreceipt.", "notifyMCPProgress", "extractProgressToken"} {
		if n := strings.Count(src, tok); n != 0 {
			t.Errorf("mcp_selfreview.go references %q %d time(s)", tok, n)
		}
	}
	// Positive controls: the scanner is not blind (acceptance.md P2, P12).
	for tok, min := range map[string]int{"recordAuditReceipt": 2, "applyGateUnmet": 5, "auditreceipt.": 2} {
		if n := strings.Count(codex, tok); n < min {
			t.Errorf("control: mcp_codex.go references %q %d time(s), want >= %d", tok, n, min)
		}
	}
	// P11: the package-wide receipt-writing call sites are the same four lines
	// as before this milestone (one definition + three calls).
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	sites := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		for line := range strings.SplitSeq(srReadFile(t, e.Name()), "\n") {
			if strings.Contains(line, "recordAuditReceipt(") && !strings.HasPrefix(strings.TrimSpace(line), "//") {
				sites++
			}
		}
	}
	if sites != 4 {
		t.Errorf("recordAuditReceipt( appears on %d non-comment lines, want the same 4 as before this milestone", sites)
	}
}

// TestSelfReview_SendsNoProgressNotifications — Q13: the new tools do not
// heartbeat. The control call (glm_audit) proves notifications do reach this
// session, so the zero below is a measurement.
func TestSelfReview_SendsNoProgressNotifications(t *testing.T) {
	f := newSelfReviewFixture(t)

	newSelfReviewRig(t, "glm", "pass")
	control := selfReviewNotifications(t, "glm_audit", map[string]any{"project_root": f.canon})
	if control == 0 {
		t.Fatal("control: glm_audit produced no notification on this session, so absence below would be unobservable")
	}
	for _, tool := range selfReviewTools {
		rig := newSelfReviewRig(t, strings.TrimSuffix(tool, "_review"), "pass")
		if got := selfReviewNotifications(t, rig.tool, map[string]any{"scope": "card", "project_root": f.canon}); got != 0 {
			t.Errorf("%s sent %d notification(s); the self-review tools carry no heartbeat", tool, got)
		}
		if rig.calls() == 0 {
			t.Errorf("%s never reached its reviewer, so the zero above says nothing", tool)
		}
	}
}

// --- AC-015 (the catalogue-row slice, I3-7 i) ---

// TestSelfReview_CatalogueAndGuideCarryTheNewRows — the figure guards read
// numbers only; a doc that bumps 45 to 47 without ever naming the tools passes
// them. These rows are what makes the figures true.
func TestSelfReview_CatalogueAndGuideCarryTheNewRows(t *testing.T) {
	files := []string{
		"../../.claude/rules/moai/core/moai-mcp-tools-catalogue.md",
		"../../internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md",
	}
	for _, loc := range []string{"ko", "en", "ja", "zh"} {
		files = append(files, "../../docs-site/content/"+loc+"/guides/mcp-server.md")
	}
	for _, p := range files {
		body := srReadFile(t, p)
		for _, tool := range selfReviewTools {
			row := false
			for line := range strings.SplitSeq(body, "\n") {
				if strings.HasPrefix(line, "|") && strings.Contains(line, "`mcp__moai__"+tool+"`") {
					row = true
				}
			}
			if !row {
				t.Errorf("%s has no table row for `mcp__moai__%s`", p, tool)
			}
		}
	}
	for _, p := range projectRootDocFiles {
		body := srReadFile(t, p)
		for _, tool := range selfReviewTools {
			if !strings.Contains(body, "`"+tool+"`") {
				t.Errorf("%s does not name `%s` in the project_root inventory", p, tool)
			}
		}
		if !strings.Contains(body, "Four of the twenty-two") {
			t.Errorf("%s still carries the stale REQUIRE-count sentence (want \"Four of the twenty-two\")", p)
		}
	}
}
