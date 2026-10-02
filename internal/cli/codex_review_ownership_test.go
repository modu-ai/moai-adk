package cli

// SPEC-CODEX-REVIEW-OWNERSHIP-001 M1/M2 — the tree_scope policy of the codex
// review gate (REQ-CRO-001..006, AC-001..006).
//
// What the policy decides: a session whose scope class is TREE and which carries
// no WT- branch evidence either reviews the whole uncommitted tree (tree_scope
// review, the default) or is let through without a review (tree_scope skip).
// The policy reads exactly two inputs — the resolver's result and the key value
// (REQ-CRO-005) — and runs on both automatic paths: the Claude Stop hook
// (HandleCodexReviewGate) and Codex Stop-chain member 6 (codexReviewMember). The
// explicit producer `moai verify codex-review` ignores it (REQ-CRO-006).
//
// Observation discipline (acceptance.md §A): every assertion here reads what the
// gate DID — reviewer lookups, detector calls, the wire requests of the faked
// codex session, the policy log row, the chain outcome — never a verdict value
// alone.
//
// Class of each test (acceptance.md §C): the reader, skip, env-matrix, root and
// log tests are RED before the policy exists; the "NonSkip", "WTSessions" and
// "ExplicitProducer" tests are PRESERVE lines that are GREEN before and after
// (a skip implementation that swallows them turns them red — the mutant probe).

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// --- fixtures -------------------------------------------------------------

// ownershipWorkflow renders a workflow.yaml enabling the codex gate with the
// given tree_scope value text ("" = the key is absent).
func ownershipWorkflow(treeScope string) string {
	body := "workflow:\n  codex:\n    review_gate:\n      enabled: true\n"
	if treeScope != "" {
		body += "      tree_scope: " + treeScope + "\n"
	}
	return body
}

// writeOwnershipConfig writes the gate config into root's workflow.yaml.
func writeOwnershipConfig(t *testing.T, root, treeScope string) {
	t.Helper()
	writeCardFile(t, root, filepath.Join(".moai", "config", "sections", "workflow.yaml"), ownershipWorkflow(treeScope))
}

// captureTreeScopeSkips swaps the policy's skip-log seam and returns the values
// logged (one entry per skip row).
func captureTreeScopeSkips(t *testing.T) *[]reviewScope {
	t.Helper()
	var rows []reviewScope
	prev := treeScopeSkipLogger
	treeScopeSkipLogger = func(s reviewScope, _ string) { rows = append(rows, s) }
	t.Cleanup(func() { treeScopeSkipLogger = prev })
	return &rows
}

// recordTreeScopeReads swaps the policy's key reader with a recorder that
// delegates to the real reader and notes every root it was asked to read.
func recordTreeScopeReads(t *testing.T) *[]string {
	t.Helper()
	var roots []string
	prev := reviewGateTreeScopeReader
	reviewGateTreeScopeReader = func(root string) string {
		roots = append(roots, root)
		return prev(root)
	}
	t.Cleanup(func() { reviewGateTreeScopeReader = prev })
	return &roots
}

// newTreeSession builds a TREE-scope session directory of the named variant
// with one reviewable uncommitted file, a fixed shape for every variant except
// "non-git" (nothing to review there, which is the point of that variant).
func newTreeSession(t *testing.T, variant string) string {
	t.Helper()
	if variant == "non-git" {
		return t.TempDir()
	}
	f := newStopFixture(t) // branch main, .moai/ git-ignored
	switch variant {
	case "main":
	case "develop":
		f.git(t, "checkout", "-q", "-b", "develop")
	case "detached":
		f.git(t, "checkout", "-q", "--detach")
	default:
		t.Fatalf("unknown tree-session variant %q", variant)
	}
	f.dirty(t, "reviewable")
	return f.root
}

// gatePath drives the Claude Stop-hook handler for a session sitting in dir
// whose config root is rooted at projectDir.
func gatePath(t *testing.T, dir, projectDir string) *hook.HookOutput {
	t.Helper()
	out, err := HandleCodexReviewGate(&hook.HookInput{SessionID: "own", CWD: dir}, true, projectDir)
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	return out
}

// chainPath drives Codex Stop-chain member 6 for the session tree root.
func chainPath(t *testing.T, root string) stopMemberOutcome {
	t.Helper()
	fakeCodexVersion(t, "codex-cli 0.0.0-ownership")
	return newCodexStopChain(root, stopInput("own", false)).codexReviewMember(context.Background())
}

// chainSkipped reports whether the member outcome is the policy's skip: allowed,
// not-applicable, naming tree_scope, and without having read a receipt.
func chainSkipped(o stopMemberOutcome) bool {
	return o.Decision == codexadapter.DecisionAllow && o.Status == stopStatusNotApplicable &&
		strings.Contains(o.Reason, "tree_scope") && !o.ReceiptRead
}

// requireGateSkipped asserts the strongest skip observation on the Claude path:
// ALLOW, and not one reviewer lookup, detector call or wire request happened.
func requireGateSkipped(t *testing.T, label string, out *hook.HookOutput, p *ownershipProbe, skips *[]reviewScope) {
	t.Helper()
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("%s: a skipped turn must ALLOW, got %+v", label, out)
	}
	if p.lookups != 0 || p.detects != 0 || p.reviewed() {
		t.Errorf("%s: a skip must precede the detector, the lookup and the review (lookups=%d detects=%d reviewed=%v)",
			label, p.lookups, p.detects, p.reviewed())
	}
	if len(*skips) != 1 {
		t.Errorf("%s: exactly one policy skip row expected, got %d", label, len(*skips))
	}
}

// --- AC-001: the reader ---------------------------------------------------

// nestedTreeScope renders the deployed (workflow-rooted) shape with a raw
// tree_scope value text.
func nestedTreeScope(value string) string {
	return "workflow:\n  codex:\n    review_gate:\n      tree_scope: " + value + "\n"
}

// treeScopeFixtures is the AC-001 truth table, shared by the reader test and the
// loader-agreement pin. skip=true means the key must read as skip.
var treeScopeFixtures = []struct {
	name string
	body string
	skip bool
}{
	// the skip family: case and surrounding whitespace are ignored
	{"skip", nestedTreeScope("skip"), true},
	{"padded and capitalised", nestedTreeScope(`" Skip "`), true},
	{"upper case", nestedTreeScope("SKIP"), true},
	{"inline comment", nestedTreeScope("skip  # note"), true},
	{"double quoted", nestedTreeScope(`"skip"`), true},
	{"single quoted", nestedTreeScope(`'skip'`), true},
	// everything else reads review
	{"review", nestedTreeScope("review"), false},
	{"empty value", nestedTreeScope(""), false},
	{"unknown value", nestedTreeScope("never"), false},
	{"prefix variant", nestedTreeScope("skipx"), false},
	{"suffix variant", nestedTreeScope("no-skip"), false},
	{"key absent", "workflow:\n  codex:\n    review_gate:\n      enabled: true\n", false},
	{"malformed yaml", "workflow:\n\tcodex: [oops\n", false},
	{"file absent", "", false},
	// misplaced keys: only the nested workflow.codex.review_gate path counts
	{"flat without workflow root", "codex:\n  review_gate:\n    tree_scope: skip\n", false},
	{"commented out", "workflow:\n  codex:\n    review_gate:\n      # tree_scope: skip\n      enabled: true\n", false},
	{"under the multi gate", "workflow:\n  multi:\n    review_gate:\n      tree_scope: skip\n", false},
	{"under codex task", "workflow:\n  codex:\n    task:\n      tree_scope: skip\n", false},
}

// TestTreeScopeReader_TruthTable pins REQ-CRO-001: exactly the nested path's
// skip family reads skip; every other shape, value and failure reads review.
func TestTreeScopeReader_TruthTable(t *testing.T) {
	for _, tc := range treeScopeFixtures {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeGateWorkflowYAML(t, tc.body)
			want := config.CodexReviewGateTreeScopeReview
			if tc.skip {
				want = config.CodexReviewGateTreeScopeSkip
			}
			if got := readCodexReviewGateTreeScope(dir); got != want {
				t.Errorf("readCodexReviewGateTreeScope = %q, want %q", got, want)
			}
		})
	}
	if got := readCodexReviewGateTreeScope(""); got != config.CodexReviewGateTreeScopeReview {
		t.Errorf("empty root must read review (the pre-policy direction), got %q", got)
	}
}

// TestTreeScopeReader_AgreesWithConfigLoader is the drift guard (the sibling of
// TestReviewGateReaders_AgreeWithConfigLoader): on every fixture the hand-rolled
// reader and config.Loader — through the single normaliser — return the same
// policy. A loader error is the review direction, as for the reader.
func TestTreeScopeReader_AgreesWithConfigLoader(t *testing.T) {
	for _, tc := range treeScopeFixtures {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeGateWorkflowYAML(t, tc.body)
			want := config.CodexReviewGateTreeScopeReview
			if cfg, err := config.NewLoader().Load(filepath.Join(dir, ".moai")); err == nil {
				want = config.NormalizeCodexReviewGateTreeScope(cfg.Workflow.Codex.ReviewGate.TreeScope)
			}
			if got := readCodexReviewGateTreeScope(dir); got != want {
				t.Errorf("reader = %q, config loader = %q (schema drift)", got, want)
			}
		})
	}
}

// --- AC-002: the skip, on both automatic paths ----------------------------

// TestCodexReviewGate_TreeScopeSkip pins REQ-CRO-002 on the Claude path: with
// tree_scope skip, a tree-scope session without WT- evidence is allowed before
// the detector, the lookup and the review — one policy row — while the same
// session under tree_scope review reaches the self-gate as before.
func TestCodexReviewGate_TreeScopeSkip(t *testing.T) {
	for _, variant := range []string{"main", "develop", "detached", "non-git"} {
		t.Run(variant, func(t *testing.T) {
			root := newTreeSession(t, variant)

			writeOwnershipConfig(t, root, "skip")
			p := newOwnershipProbe(t)
			skips := captureTreeScopeSkips(t)
			out := gatePath(t, root, root)
			requireGateSkipped(t, variant+"/skip", out, p, skips)
			if len(p.scopes) != 1 || p.scopes[0].Class != reviewScopeTree {
				t.Errorf("%s: the skip row rides after the scope row (tree class), scopes=%+v", variant, p.scopes)
			}

			// control: the same session under review is NOT skipped
			writeOwnershipConfig(t, root, "review")
			c := newOwnershipProbe(t)
			cskips := captureTreeScopeSkips(t)
			gatePath(t, root, root)
			if len(*cskips) != 0 {
				t.Errorf("%s/review: no skip row expected, got %d", variant, len(*cskips))
			}
			if c.detects != 1 {
				t.Errorf("%s/review: the self-gate must run once, got %d detector calls", variant, c.detects)
			}
			if variant != "non-git" {
				cwd, target := c.request(t)
				if got, _ := target["type"].(string); got != codexTargetUncommitted || cwd != root {
					t.Errorf("%s/review: request = {%v, %q}, want {%s, %q}", variant, target["type"], cwd, codexTargetUncommitted, root)
				}
			}
		})
	}
}

// TestCodexStopChain_TreeScopeSkip pins REQ-CRO-002 on the Codex path: member 6
// returns allow / not-applicable naming tree_scope, reads no receipt and never
// touches the reviewer; under review the same state takes the receipt route.
func TestCodexStopChain_TreeScopeSkip(t *testing.T) {
	for _, variant := range []string{"main", "develop", "detached", "non-git"} {
		t.Run(variant, func(t *testing.T) {
			root := newTreeSession(t, variant)

			writeOwnershipConfig(t, root, "skip")
			p := newOwnershipProbe(t)
			skips := captureTreeScopeSkips(t)
			got := chainPath(t, root)
			if !chainSkipped(got) {
				t.Errorf("%s/skip: member 6 must skip (allow, not-applicable, tree_scope reason, no receipt read), got %+v", variant, got)
			}
			if p.lookups != 0 || p.detects != 0 || p.reviewed() || len(*skips) != 1 {
				t.Errorf("%s/skip: lookups=%d detects=%d reviewed=%v skipRows=%d, want 0/0/false/1",
					variant, p.lookups, p.detects, p.reviewed(), len(*skips))
			}

			writeOwnershipConfig(t, root, "review")
			c := newOwnershipProbe(t)
			cskips := captureTreeScopeSkips(t)
			got = chainPath(t, root)
			if chainSkipped(got) || len(*cskips) != 0 {
				t.Errorf("%s/review: member 6 must not skip, got %+v (skip rows %d)", variant, got, len(*cskips))
			}
			if variant != "non-git" && !got.ReceiptRead {
				t.Errorf("%s/review: the receipt route must be taken, got %+v", variant, got)
			}
			if c.detects != 1 {
				t.Errorf("%s/review: the self-gate must run once, got %d detector calls", variant, c.detects)
			}
		})
	}
}

// --- AC-003: non-skip values keep the pre-policy tree request -------------

// TestTreeScope_NonSkipValuesKeepTreeRequest pins REQ-CRO-003 (the REQ-CGS-003
// amendment): an absent key, review, an unknown value and a misplaced skip all
// leave the tree-scope request shape-identical to its form before the policy
// (REQ-CRT-006). PRESERVE line — GREEN before and after.
func TestTreeScope_NonSkipValuesKeepTreeRequest(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"key absent", ownershipWorkflow("")},
		{"review", ownershipWorkflow("review")},
		{"unknown value", ownershipWorkflow("never")},
		{"misplaced flat skip", "codex:\n  review_gate:\n    tree_scope: skip\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newStopFixture(t)
			f.dirty(t, "tree request regression")
			f.write(t, filepath.Join(".moai", "config", "sections", "workflow.yaml"), tc.body)
			p := newOwnershipProbe(t)

			out := gatePath(t, f.root, f.root)
			if out == nil || out.Decision != hook.DecisionBlock {
				t.Fatalf("the probe review fails, so the turn must BLOCK; got %+v", out)
			}
			cwd, target := p.request(t)
			if got, _ := target["type"].(string); got != codexTargetUncommitted {
				t.Errorf("target.type = %q, want %q (REQ-CRT-006 shape)", got, codexTargetUncommitted)
			}
			if cwd != f.root {
				t.Errorf("thread/start cwd = %q, want the resolved tree %q", cwd, f.root)
			}
		})
	}
}

// --- AC-004: WT- sessions are never swallowed by skip ---------------------

// TestTreeScopeSkip_WTSessionsStillReviewed pins REQ-CRO-004 under tree_scope
// skip on both paths: a card session (merge base available) reviews its card
// diff, and a WT- session without a develop base reviews the whole tree and
// keeps its "merge base unavailable" scope row. A skip implementation keyed on
// the scope CLASS alone would swallow the second case — that is the mutant this
// test kills (see the M2 mutant probe in progress.md §E.2).
func TestTreeScopeSkip_WTSessionsStillReviewed(t *testing.T) {
	t.Run("card session with a merge base", func(t *testing.T) {
		f := newCardScopeFixture(t)
		writeOwnershipConfig(t, f.card, "skip")

		p := newOwnershipProbe(t)
		skips := captureTreeScopeSkips(t)
		out := gatePath(t, f.card, f.card)
		if out == nil || out.Decision != hook.DecisionBlock {
			t.Fatalf("a card session under skip must still be reviewed and BLOCK on the probe finding; got %+v", out)
		}
		requireCardRequest(t, p.sess.sent, f)
		if len(*skips) != 0 {
			t.Errorf("a card session must log no policy skip, got %d rows", len(*skips))
		}

		got := chainPath(t, f.card)
		if chainSkipped(got) || !got.ReceiptRead {
			t.Errorf("member 6 must take the receipt route for a card session, got %+v", got)
		}
	})
	t.Run("WT- session without a develop base", func(t *testing.T) {
		tree := newWTNoBaseTree(t)
		writeOwnershipConfig(t, tree, "skip")

		p := newOwnershipProbe(t)
		skips := captureTreeScopeSkips(t)
		out := gatePath(t, tree, tree)
		if out == nil || out.Decision != hook.DecisionBlock {
			t.Fatalf("a WT- session without a base under skip must still be reviewed; got %+v", out)
		}
		cwd, target := p.request(t)
		if got, _ := target["type"].(string); got != codexTargetUncommitted || cwd != tree {
			t.Errorf("request = {%v, %q}, want the whole-tree request {%s, %q}", target["type"], cwd, codexTargetUncommitted, tree)
		}
		if len(p.scopes) != 1 || !strings.Contains(p.scopes[0].Basis, "merge base unavailable") {
			t.Errorf("the scope row must keep the unavailable-merge-base basis, got %+v", p.scopes)
		}
		if len(*skips) != 0 {
			t.Errorf("a WT- session must log no policy skip, got %d rows", len(*skips))
		}

		got := chainPath(t, tree)
		if chainSkipped(got) || !got.ReceiptRead {
			t.Errorf("member 6 must take the receipt route for a WT- session, got %+v", got)
		}
	})
}

// --- AC-005: the decision reads no environment ----------------------------

// ownershipEnvKeys derives the launcher-env key list from internal/config's
// envkeys.go: every MOAI_KANBAN* and MOAI_FACTORY* constant. Deriving it (rather
// than listing it) keeps the matrix honest when a key is added or deleted.
func ownershipEnvKeys(t *testing.T) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "config", "envkeys.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse envkeys.go: %v", err)
	}
	var keys []string
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		v := strings.Trim(lit.Value, `"`)
		if strings.HasPrefix(v, "MOAI_KANBAN") || strings.HasPrefix(v, "MOAI_FACTORY") {
			keys = append(keys, v)
		}
		return true
	})
	if len(keys) < 8 {
		t.Fatalf("env key derivation found only %d keys (%v) — the scan is blind", len(keys), keys)
	}
	return keys
}

// setOwnershipEnv clears every launcher env key (restoring on cleanup) and then
// sets the row's values; a value of "" in row means set-but-empty.
func setOwnershipEnv(t *testing.T, all []string, row map[string]string) {
	t.Helper()
	for _, k := range all {
		old, had := os.LookupEnv(k)
		if err := os.Unsetenv(k); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if had {
				_ = os.Setenv(k, old)
			}
		})
	}
	for k, v := range row {
		t.Setenv(k, v)
	}
}

// TestTreeScopePolicy_EnvMatrix pins REQ-CRO-005 by behaviour: the policy's
// decision is identical across every launcher env shape. T1 (a develop tree with
// foreign work) must skip and T2 (a WT- card tree) must review on BOTH paths,
// whatever the environment — including the value shapes the launchers really
// set (a leader carries no label, a factory lane carries the lane-label marker)
// and set-but-empty. A decision read
// from a label or worker value, or from a helper that reads one outside the
// policy file (which the static guard below cannot see), flips a row.
func TestTreeScopePolicy_EnvMatrix(t *testing.T) {
	keys := ownershipEnvKeys(t)
	rows := []struct {
		name string
		env  map[string]string
	}{
		{"marked leader (no label)", map[string]string{
			config.EnvFactoryRunID:    "kb-1",
			config.EnvFactoryLeadAddr: "/tmp/lead.sock", config.EnvFactorySettingsInjected: "1"}},
		{"marked factory leader", map[string]string{
			config.EnvFactoryRunID:    "kb-1",
			config.EnvFactoryLeadAddr: "/tmp/lead.sock", config.EnvFactorySettingsInjected: "1",
			config.EnvMoaiFactoryWorkers: "8"}},
		{"factory leader", map[string]string{config.EnvMoaiFactoryWorkers: "8"}},
		{"factory worker", map[string]string{config.EnvMoaiFactoryWorker: "worker-1", config.EnvMoaiFactoryWorkers: "8"}},
		{"factory lane", map[string]string{config.EnvMoaiFactoryWorker: "lane-1", config.EnvMoaiFactoryWorkers: "8"}},
		{"set but empty", func() map[string]string {
			m := map[string]string{}
			for _, k := range keys {
				m[k] = ""
			}
			return m
		}()},
		{"absent", map[string]string{}},
	}

	f := newCardScopeFixture(t)
	writeOwnershipConfig(t, f.primary, "skip")
	writeOwnershipConfig(t, f.card, "skip")
	trees := []struct {
		name     string
		dir      string
		wantSkip bool
	}{{"T1 develop tree", f.primary, true}, {"T2 WT- card tree", f.card, false}}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			setOwnershipEnv(t, keys, row.env)
			for _, tr := range trees {
				p := newOwnershipProbe(t)
				skips := captureTreeScopeSkips(t)
				gatePath(t, tr.dir, tr.dir)
				if gateSkipped := len(*skips) == 1 && !p.reviewed(); gateSkipped != tr.wantSkip {
					t.Errorf("%s gate path: skipped=%v, want %v (skip rows %d, reviewed %v)",
						tr.name, gateSkipped, tr.wantSkip, len(*skips), p.reviewed())
				}
				cskips := captureTreeScopeSkips(t)
				got := chainPath(t, tr.dir)
				if chainSkipped(got) != tr.wantSkip || (len(*cskips) == 1) != tr.wantSkip {
					t.Errorf("%s chain path: skipped=%v, want %v (outcome %+v)", tr.name, chainSkipped(got), tr.wantSkip, got)
				}
			}
		})
	}
}

// policyEnvReferences lists the environment reads (os.Getenv / LookupEnv /
// Environ / ExpandEnv and config.Env* constants) in one Go source file, found
// by syntax rather than by text, so a comment naming them is not a read. A file
// that cannot be parsed — including a missing one — fails the calling test.
func policyEnvReferences(t *testing.T, path string) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	var refs []string
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch name := sel.Sel.Name; {
		case pkg.Name == "os" && (name == "Getenv" || name == "LookupEnv" || name == "Environ" || name == "ExpandEnv"):
			refs = append(refs, "os."+name)
		case pkg.Name == "config" && strings.HasPrefix(name, "Env"):
			refs = append(refs, "config."+name)
		}
		return true
	})
	return refs
}

// TestTreeScopePolicy_SourceReadsNoEnvironment is the static half of REQ-CRO-005:
// the policy file reads no environment. It is only half — a helper elsewhere
// that reads the environment and is called from the policy file escapes this
// scan, which is why the behavioural matrix above exists as its counterpart.
// The positive control proves the scanner sees what it claims to: the resolver
// file does read the environment (observed 2 references at HEAD 984d64957).
func TestTreeScopePolicy_SourceReadsNoEnvironment(t *testing.T) {
	if refs := policyEnvReferences(t, "codex_review_tree_scope.go"); len(refs) != 0 {
		t.Errorf("the policy file must read no environment, found %v", refs)
	}
	if refs := policyEnvReferences(t, "codex_review_scope.go"); len(refs) == 0 {
		t.Error("positive control failed: the scanner found no environment read in codex_review_scope.go, so a zero elsewhere proves nothing")
	}
}

// --- AC-006: roots and the explicit producer ------------------------------

// TestTreeScopePolicy_SameDecisionOnBothPaths pins the four-way table of
// REQ-CRO-006: for the same session state, in the same root, the two automatic
// paths decide alike — skip only for a tree session without WT- evidence under
// skip.
func TestTreeScopePolicy_SameDecisionOnBothPaths(t *testing.T) {
	cases := []struct {
		name      string
		tree      func(t *testing.T) string
		value     string
		wantSkips bool
	}{
		{"tree session, skip", func(t *testing.T) string { return newTreeSession(t, "develop") }, "skip", true},
		{"tree session, review", func(t *testing.T) string { return newTreeSession(t, "develop") }, "review", false},
		{"WT- session without base, skip", newWTNoBaseTree, "skip", false},
		{"WT- session without base, review", newWTNoBaseTree, "review", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.tree(t)
			writeOwnershipConfig(t, root, tc.value)

			p := newOwnershipProbe(t)
			skips := captureTreeScopeSkips(t)
			gatePath(t, root, root)
			gateSkipped := len(*skips) == 1 && !p.reviewed()

			cskips := captureTreeScopeSkips(t)
			chainSkippedNow := chainSkipped(chainPath(t, root))
			if gateSkipped != tc.wantSkips || chainSkippedNow != tc.wantSkips || (len(*cskips) == 1) != tc.wantSkips {
				t.Errorf("gate skipped=%v, chain skipped=%v, want both %v", gateSkipped, chainSkippedNow, tc.wantSkips)
			}
		})
	}
}

// TestTreeScopePolicy_EachPathReadsItsOwnEnabledRoot pins the root half of
// REQ-CRO-006: the Claude path reads tree_scope from the root it reads enabled
// from (reviewGateConfigRoot(projectDir)), the Codex path from the session tree
// (c.root), and neither reads the other's. The reader seam records the roots.
func TestTreeScopePolicy_EachPathReadsItsOwnEnabledRoot(t *testing.T) {
	session := newTreeSession(t, "develop")
	decoy := t.TempDir() // a config-only directory, the other path's root

	t.Run("Claude path reads reviewGateConfigRoot(projectDir)", func(t *testing.T) {
		for _, tc := range []struct {
			name, decoyValue, sessionValue string
			wantSkip                       bool
		}{
			{"skip only in the project root", "skip", "review", true},
			{"skip only in the session tree", "review", "skip", false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				writeOwnershipConfig(t, decoy, tc.decoyValue)
				writeOwnershipConfig(t, session, tc.sessionValue)
				p := newOwnershipProbe(t)
				skips := captureTreeScopeSkips(t)
				reads := recordTreeScopeReads(t)
				gatePath(t, session, decoy)
				if skipped := len(*skips) == 1 && !p.reviewed(); skipped != tc.wantSkip {
					t.Errorf("skipped=%v, want %v", skipped, tc.wantSkip)
				}
				if len(*reads) != 1 || (*reads)[0] != decoy {
					t.Errorf("the key must be read once, from %q; reads=%v", decoy, *reads)
				}
			})
		}
	})
	t.Run("Codex path reads the session tree", func(t *testing.T) {
		for _, tc := range []struct {
			name, decoyValue, sessionValue string
			wantSkip                       bool
		}{
			{"skip only in the session tree", "review", "skip", true},
			{"skip only in the other root", "skip", "review", false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				writeOwnershipConfig(t, decoy, tc.decoyValue)
				writeOwnershipConfig(t, session, tc.sessionValue)
				newOwnershipProbe(t)
				reads := recordTreeScopeReads(t)
				if skipped := chainSkipped(chainPath(t, session)); skipped != tc.wantSkip {
					t.Errorf("skipped=%v, want %v", skipped, tc.wantSkip)
				}
				if len(*reads) != 1 || (*reads)[0] != session {
					t.Errorf("the key must be read once, from %q; reads=%v", session, *reads)
				}
			})
		}
	})
}

// TestTreeScopePolicy_ConfigOrphanedWorktree pins the one case where the two
// paths' roots differ: a linked worktree of a repository that keeps .moai
// untracked. The Claude path resolves the primary's key through StoreRoot; the
// Codex chain reads enabled from the worktree itself, finds the gate off and
// never reaches the policy. That asymmetry predates the policy and is
// INTENTIONAL (plan plan-audit I3-5: AC-006 (b) outranks the REQ-CRO-006
// "cannot disagree" sentence for this fixture) — each path uses the same root
// for tree_scope as for enabled, nothing more is asserted.
func TestTreeScopePolicy_ConfigOrphanedWorktree(t *testing.T) {
	fx := newUntrackedFixture(t, ownershipWorkflow("skip"))
	writeCardFile(t, fx.W, "orphan_work.go", "package main\n\n// reviewable work in the orphaned worktree\n")

	for _, tc := range []struct {
		value    string
		wantSkip bool
	}{{"skip", true}, {"review", false}} {
		t.Run("primary says "+tc.value, func(t *testing.T) {
			writeCardFile(t, fx.P, filepath.Join(".moai", "config", "sections", "workflow.yaml"), ownershipWorkflow(tc.value))
			p := newOwnershipProbe(t)
			skips := captureTreeScopeSkips(t)
			gatePath(t, fx.W, fx.W)
			if skipped := len(*skips) == 1 && !p.reviewed(); skipped != tc.wantSkip {
				t.Errorf("Claude path on the orphaned worktree: skipped=%v, want %v (the primary's key decides)", skipped, tc.wantSkip)
			}

			reads := recordTreeScopeReads(t)
			got := chainPath(t, fx.W)
			if got.Status != stopStatusNotApplicable || strings.Contains(got.Reason, "tree_scope") || len(*reads) != 0 {
				t.Errorf("Codex chain on the orphaned worktree must stop at its own enabled=false (outcome %+v, reads %v)", got, *reads)
			}
		})
	}
}

// TestTreeScopePolicy_ExplicitProducerIgnoresPolicy pins the last clause of
// REQ-CRO-006: `moai verify codex-review` is a review the user asked for, so a
// skip setting must not turn it off. PRESERVE line — GREEN before and after.
func TestTreeScopePolicy_ExplicitProducerIgnoresPolicy(t *testing.T) {
	ctx := context.Background()
	fakeCodexVersion(t, "codex-cli 0.0.0-ownership")
	f := newStopFixture(t)
	f.dirty(t, "explicit review")
	f.write(t, filepath.Join(".moai", "config", "sections", "workflow.yaml"), ownershipWorkflow("skip"))
	p := newOwnershipProbe(t)
	skips := captureTreeScopeSkips(t)

	r, err := produceCodexReviewReceipt(ctx, f.root)
	if err != nil {
		t.Fatalf("producer under tree_scope skip: %v", err)
	}
	if !p.reviewed() {
		t.Fatal("the explicit producer must still run the review under tree_scope skip")
	}
	if len(*skips) != 0 {
		t.Errorf("the producer must log no policy skip, got %d rows", len(*skips))
	}
	if r.Verdict != codexReviewVerdictFail {
		t.Errorf("producer verdict = %q, want %q for the probe finding", r.Verdict, codexReviewVerdictFail)
	}
	state, err := codexReviewReceiptStateForScope(ctx, reviewScopeResolver(f.root), "/fake/codex")
	if err != nil {
		t.Fatal(err)
	}
	if stored := verify.LoadReceipt(f.root, state); stored == nil || stored.Verdict != codexReviewVerdictFail {
		t.Errorf("the receipt must be recorded for the explicit review, got %+v", stored)
	}
}

// --- AC-002 (log row) -----------------------------------------------------

// TestTreeScopeSkipRow pins the skip log row's content (REQ-CRO-002): gate,
// class, key value and the resolver's basis — one structured row, distinct from
// the scope row.
func TestTreeScopeSkipRow(t *testing.T) {
	row := treeScopeSkipRow(reviewScope{Class: reviewScopeTree, Basis: "no card branch: develop", Branch: "develop"}, "skip")
	want := map[string]any{
		"gate": "codex-review-gate", "scope": reviewScopeTree, "tree_scope": "skip",
		"basis": "no card branch: develop",
	}
	for k, v := range want {
		if row[k] != v {
			t.Errorf("row[%q] = %v, want %v (row %v)", k, row[k], v, row)
		}
	}
	if len(row) != len(want) {
		t.Errorf("the skip row carries exactly gate, scope, tree_scope and basis, got %v", row)
	}
}
