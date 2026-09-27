package hook

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// gitio helpers the fixture builders use.
func gitioHeadOf(dir string) string {
	sha, err := gitio.Head(dir)
	if err != nil {
		return ""
	}
	return sha
}

func gitioIsAncestor(dir, a, b string) (bool, error) { return gitio.IsAncestor(dir, a, b) }

func gitioNonMerge(dir, from, to string) ([]closure.CommitPaths, error) {
	return gitio.NonMergeCommits(dir, from, to)
}

func gitioHeadSeam(root string) (string, error) { return gitio.Head(root) }

// ─── harness ───

// closureCfgProvider is a ConfigProvider over one configuration.
type closureCfgProvider struct{ c *config.Config }

func (p closureCfgProvider) Get() *config.Config { return p.c }

// autonomyProvider builds a provider whose effective autonomy mode is mode
// ("contract", "guided", "" absent, "bogus").
func autonomyProvider(mode string) closureCfgProvider {
	wf := config.NewDefaultWorkflowConfig()
	wf.Autonomy.Mode = mode
	wf.Autonomy.Contract.SecondReview = "required"
	wf.Autonomy.Contract.PushDevelop = true
	return closureCfgProvider{c: &config.Config{Workflow: wf}}
}

// bashInput is a Bash tool call whose command runs in cwd.
func closureBashInput(command, cwd string) *HookInput {
	return &HookInput{
		ToolName:  "Bash",
		ToolInput: []byte(`{"command":` + quoteJSON(command) + `}`),
		CWD:       cwd,
		SessionID: "closure-test",
	}
}

func quoteJSON(s string) string {
	data, _ := json.Marshal(s)
	return string(data)
}

// writeGitFlowConfig writes the fixture's git-strategy.yaml so
// LoadGitFlowDevelopBranch resolves `develop` (the hook reads the tree's
// configuration, exactly as production does).
func writeGitFlowConfig(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	f.Write(filepath.Join(f.Root, ".moai", "config", "sections", "git-strategy.yaml"),
		"git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: develop\n")
}

// queueC1 installs the fixture's c1 card mapping.
func queueC1(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(f.Root))
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		spec := closuretest.SpecID
		rec.Items = append(rec.Items, kanban.BacklogItem{
			ID: closuretest.Card, Text: "fixture", AddedAt: "2026-09-27T00:00:00Z",
			SpecID: &spec, State: kanban.BacklogStatePicked,
		})
		return nil
	}); err != nil {
		t.Fatalf("queue: %v", err)
	}
}

// secondReviewAt records a performed review with the given verdict at commit
// head.
func secondReviewAt(t *testing.T, f *closuretest.Fixture, head, verdict string) {
	t.Helper()
	digest := closureDigestOf(t, f)
	disagree := false
	line, err := json.Marshal(map[string]any{
		"schema_version": 1, "card": closuretest.Card, "contract_card": closuretest.Card,
		"spec_id": closuretest.SpecID, "contract_sha256": digest, "head_sha": head,
		"target": "baseBranch",
		"scope": map[string]any{
			"base_branch": "develop", "base_sha": "base01", "head_sha": head,
			"changed_files": 3, "diff_sha256": "diff01",
		},
		"backends": []map[string]string{
			{"backend": "codex", "gate": "required", "verdict": verdict},
		},
		"participant_count": 1, "disagreement_flag": &disagree,
		"recorded_at": "2026-09-27T09:00:00Z",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	f.Write(filepath.Join(f.EvidenceDir, closure.SecondReviewFile), string(line)+"\n")
}

func closureDigestOf(t *testing.T, f *closuretest.Fixture) string {
	t.Helper()
	rep, _, _, _, _, err := closure.LoadBuildSideFiles(f.CardDir, closuretest.SpecID,
		closuretest.Autonomy(), nil, nil)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return rep.RecordedContractSHA256
}

// closureReportPair builds the report pair at the card evidence home.
func closureReportPair(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	rep, progress, acceptance, ac, acOK, err := closure.LoadBuildSideFiles(
		f.CardDir, closuretest.SpecID, closuretest.Autonomy(), nil, nil)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ev := closure.EvidenceFor(f.CardDir, closuretest.Card)
	reviews, _, _ := closure.LoadSecondReviews(ev.SecondReview)
	verdicts, _, _ := closure.LoadVerdictRecords(ev.ClosureVerdict)
	r, err := closure.Build(closure.BuildInput{
		Card: closuretest.Card, SpecID: closuretest.SpecID, Home: f.CardDir,
		Mode: "contract", SecondReviewPolicy: "required",
		Verify: rep, ProgressMD: progress, AcceptanceMD: acceptance,
		MeasuredAC: ac, ACAvailable: acOK,
		SecondReviews: reviews, Verdicts: verdicts,
		Facts: closure.GitFacts{
			IsAncestor:      func(a, b string) (bool, error) { return gitioIsAncestor(f.CardDir, a, b) },
			NonMergeCommits: func(from, to string) ([]closure.CommitPaths, error) { return gitioNonMerge(f.CardDir, from, to) },
		},
		EvalCommit: gitioHeadOf(f.CardDir),
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if _, err := closure.WriteReportPair(ev, r); err != nil {
		t.Fatalf("write pair: %v", err)
	}
}

// runClosurePush runs the push check directly against a fixture.
func runClosurePush(t *testing.T, provider closureCfgProvider, command, cwd string) closurePushEvalResult {
	t.Helper()
	return checkClosurePush(provider, closureBashInput(command, cwd))
}

// expectClosureStop asserts a deny whose reason begins with the push-stop
// sentinel.
func expectClosureStop(t *testing.T, r closurePushEvalResult, wantSubstrings ...string) {
	t.Helper()
	if r.decision != DecisionDeny {
		t.Fatalf("decision = %q reason %q, want deny", r.decision, r.reason)
	}
	if !strings.HasPrefix(r.reason, ClosurePushStopPrefix) {
		t.Fatalf("reason = %q, want prefix %s", r.reason, ClosurePushStopPrefix)
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(r.reason, want) {
			t.Fatalf("reason = %q, want it to contain %q", r.reason, want)
		}
	}
}

// ─── AC-CLOSURE_015 — push stop evaluates the pushed ref ───

func TestAC_CLOSURE_015(t *testing.T) {
	f := closuretest.New(t)
	queueC1(t, f)
	writeGitFlowConfig(t, f)
	provider := autonomyProvider("contract")

	// A tree whose checked-out branch is main (the push names the branch, so
	// the evaluation must not read the current branch).
	f.Git(f.Root, "branch", "main", "HEAD")
	f.Git(f.Root, "worktree", "add", filepath.Join(f.Parent, "main-tree"), "-b", "wt-main-tree", "main")
	mainTree := filepath.Join(f.Parent, "main-tree")

	// Merge the card branch into local develop so the pushed source commit
	// carries the SPEC (the not-ready state: report but no review).
	f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
	closureReportPair(t, f)

	forms := []struct {
		name    string
		command string
		cwd     string
	}{
		{"branch push from a main tree", "git push origin develop", mainTree},
		{"HEAD refspec from the develop worktree", "git push origin HEAD:develop", f.Root},
		{"forced branch push", "git push origin +develop", mainTree},
		{"full-ref destination", "git push origin develop:refs/heads/develop", mainTree},
		{"-C into the develop worktree", "git -C " + f.Root + " push origin develop", mainTree},
	}
	for _, tc := range forms {
		t.Run(tc.name, func(t *testing.T) {
			r := runClosurePush(t, provider, tc.command, tc.cwd)
			expectClosureStop(t, r, closuretest.SpecID+"=", "second_review_not_performed")
		})
	}

	t.Run("performed pass review clears the stop", func(t *testing.T) {
		secondReviewAt(t, f, gitioHeadOf(f.Root), "pass")
		closureReportPair(t, f)
		r := runClosurePush(t, provider, "git push origin develop", mainTree)
		if r.decision != "" {
			t.Fatalf("decision = %q reason %q, want no A4 deny", r.decision, r.reason)
		}
	})

	t.Run("completed status in the pushed commit still stops", func(t *testing.T) {
		// The sync-commit landing: the SPEC goes completed before the push.
		// Back to the not-ready state first (no second review), then merge —
		// the SPEC-directory change in the range keeps the closed card a
		// candidate (REQ-CLOSURE-015's own-card branch).
		if err := os.Remove(filepath.Join(f.EvidenceDir, closure.SecondReviewFile)); err != nil {
			t.Fatalf("remove review: %v", err)
		}
		f.Write(filepath.Join(f.SpecDir, "spec.md"),
			"---\nid: "+closuretest.SpecID+"\nstatus: completed\n---\n")
		f.Git(f.CardDir, "add", ".moai")
		f.Git(f.CardDir, "commit", "-m", "sync")
		f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
		// The report is regenerated at the new head so the only missing
		// readiness is the second review.
		closureReportPair(t, f)
		r := runClosurePush(t, provider, "git push origin develop", mainTree)
		expectClosureStop(t, r, closuretest.SpecID+"=", "second_review_not_performed")
	})

	t.Run("non-integration push is not evaluated", func(t *testing.T) {
		before := closurePushEvalInvocations
		r := runClosurePush(t, provider, "git push origin WT-feature", mainTree)
		if r.decision != "" {
			t.Fatalf("decision = %q, want no evaluation", r.decision)
		}
		if closurePushEvalInvocations != before {
			t.Fatalf("evaluations %d -> %d, want none", before, closurePushEvalInvocations)
		}
	})

	t.Run("out-of-range contract is not a candidate", func(t *testing.T) {
		// SPEC-FIXTURE-002 (card c0, completed, src/**, signed) already on
		// origin/develop; the range's src/ change is c1's src/b.go.
		f.Write(filepath.Join(f.Root, ".moai", "specs", "SPEC-FIXTURE-002", "spec.md"),
			"---\nid: SPEC-FIXTURE-002\nstatus: completed\n---\n")
		f.Write(filepath.Join(f.Root, ".moai", "specs", "SPEC-FIXTURE-002", "acceptance.md"), closuretest.Acceptance)
		f.Write(filepath.Join(f.Root, ".moai", "specs", "SPEC-FIXTURE-002", "contract.yaml"),
			closuretest.DraftContract("c0", "SPEC-FIXTURE-002",
				[]string{"frozen-files"}, []string{"src/**", ".moai/specs/SPEC-FIXTURE-002/**"}))
		signFixtureOnRoot(t, f, "SPEC-FIXTURE-002")
		f.Git(f.Root, "add", ".moai")
		f.Git(f.Root, "commit", "-m", "spec 002")
		f.Git(f.Root, "push", "-q", "origin", "develop")
		// c1's src/b.go rides the card branch; merge it.
		f.Write(filepath.Join(f.CardDir, "src", "b.go"), "package b\n")
		f.Git(f.CardDir, "add", "src")
		f.Git(f.CardDir, "commit", "-m", "b")
		f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
		// c1 ready at the new head.
		secondReviewAt(t, f, gitioHeadOf(f.Root), "pass")
		closureReportPair(t, f)

		r := runClosurePush(t, provider, "git push origin develop", mainTree)
		if r.decision != "" {
			t.Fatalf("decision = %q reason %q, want no A4 deny", r.decision, r.reason)
		}
	})

	t.Run("another card's SPEC-directory edit re-admits the closed card", func(t *testing.T) {
		// One more non-merge commit in the range edits SPEC-FIXTURE-002's own
		// directory: the closed contract becomes a candidate again and its
		// missing evidence stops the push (spec.md §H accepted residual).
		f.Write(filepath.Join(f.Root, ".moai", "specs", "SPEC-FIXTURE-002", "spec.md"),
			"---\nid: SPEC-FIXTURE-002\nstatus: completed\nsuperseded-by: none # one supersession line\n---\n")
		f.Git(f.Root, "add", ".moai")
		f.Git(f.Root, "commit", "-m", "supersession line")
		r := runClosurePush(t, provider, "git push origin develop", mainTree)
		expectClosureStop(t, r, "SPEC-FIXTURE-002=", "closure_report_missing", "second_review_not_performed")
	})
}

// signFixtureOnRoot signs one SPEC directly in the fixture root.
func signFixtureOnRoot(t *testing.T, f *closuretest.Fixture, specID string) {
	t.Helper()
	o := f.SignOptions()
	o.ProjectRoot = f.Root
	o.SpecIDs = []string{specID}
	seams, _ := signtest.Seams(true, map[string]string{}, specID)
	seams.GitHead = gitioHeadSeam
	res, err := sign.Sign(o, seams)
	if err != nil {
		t.Fatalf("sign %s: %v", specID, err)
	}
	if res.Refusal != "" {
		t.Fatalf("sign %s refused: %s", specID, res.Refusal)
	}
}

// ─── AC-CLOSURE_017 — undetermined denies ───

func TestAC_CLOSURE_017(t *testing.T) {
	f := closuretest.New(t)
	queueC1(t, f)
	writeGitFlowConfig(t, f)
	provider := autonomyProvider("contract")
	f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
	secondReviewAt(t, f, gitioHeadOf(f.Root), "pass")
	closureReportPair(t, f)

	forms := []string{
		"git push --all origin",
		"git push --mirror origin",
		`sh -c "git push origin develop"`,
		`git push origin "$BR"`,
		"git push origin $(git branch --show-current)",
	}
	for _, command := range forms {
		t.Run(command, func(t *testing.T) {
			r := runClosurePush(t, provider, command, f.Root)
			expectClosureStop(t, r, "push_check_undetermined")
		})
	}

	t.Run("bare push without upstream", func(t *testing.T) {
		r := runClosurePush(t, provider, "git push", f.CardDir)
		expectClosureStop(t, r, "push_check_undetermined")
	})

	t.Run("missing remote ref", func(t *testing.T) {
		f.Git(f.Root, "branch", "-dr", "origin/develop")
		r := runClosurePush(t, provider, "git push origin develop", f.Root)
		expectClosureStop(t, r, "push_check_undetermined")
	})
}

// ─── AC-CLOSURE_021 — the verdict command is reserved to humans ───

func TestAC_CLOSURE_021(t *testing.T) {
	f := closuretest.New(t)
	for _, mode := range []string{"guided", "contract"} {
		decision, reason := checkContractVerdict(closureBashInput("moai contract verdict c1 accept", f.Root))
		if decision != DecisionDeny || !strings.HasPrefix(reason, ClosureVerdictDenyPrefix) {
			t.Fatalf("mode %s: decision=%q reason=%q, want deny with the sentinel", mode, decision, reason)
		}
	}

	// No writer under this package touches closure-verdict.jsonl: the report
	// pair write and the second-review record never create the verdict file,
	// and the cli surface's verdict-file writes live only in
	// contract_verdict.go (the human path).
	closureReportPair(t, f)
	if _, err := os.Stat(filepath.Join(f.EvidenceDir, closure.ClosureVerdictFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("closure-verdict.jsonl exists after report writes: %v", err)
	}
	hits := grepGuard(t, "../cli", `ClosureVerdictFile|closure-verdict\.jsonl`,
		map[string]bool{"contract_verdict.go": true, "contract_verdict_test.go": true,
			"mcp_audit_multi_record_test.go": true, "contract_closure_test.go": true})
	if len(hits) > 0 {
		t.Fatalf("closure-verdict writes outside the human path:\n%s", strings.Join(hits, "\n"))
	}
}

// grepGuard scans the named directory's .go files for needle, skipping the
// exempt names (the repo's static-guard convention).
func grepGuard(t *testing.T, dir, needle string, exempt map[string]bool) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var hits []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || exempt[e.Name()] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, needle) && !strings.HasPrefix(strings.TrimSpace(line), "//") {
				hits = append(hits, filepath.Join(dir, e.Name())+":"+itoa(i+1)+": "+strings.TrimSpace(line))
			}
		}
	}
	return hits
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// ─── AC-CLOSURE_023 — guided mode is a no-op ───

func TestAC_CLOSURE_023(t *testing.T) {
	f := closuretest.New(t)
	queueC1(t, f)
	// The evaluation counter is package-global and earlier AC tests ran
	// evaluations; zero it for this test's no-op assertions.
	savedEvalCount := closurePushEvalInvocations
	closurePushEvalInvocations = 0
	t.Cleanup(func() { closurePushEvalInvocations = savedEvalCount })
	corpus := []string{
		"git push origin develop",
		"git status",
		"go test ./...",
	}
	for _, mode := range []string{"guided", "", "bogus"} {
		provider := autonomyProvider(mode)
		for _, command := range corpus {
			before := closurePushEvalInvocations
			r := runClosurePush(t, provider, command, f.Root)
			if r.decision != "" {
				t.Fatalf("mode %q command %q: decision %q, want no A4 decision", mode, command, r.decision)
			}
			if closurePushEvalInvocations != before {
				t.Fatalf("mode %q: A4 evaluations ran (%d -> %d)", mode, before, closurePushEvalInvocations)
			}
		}
	}

	// Byte-identical Handle output: the checks add no decision, so the
	// marshaled HookOutput equals the plain safe-default the unchanged path
	// returns for these inputs.
	handler := NewPreToolHandler(autonomyProvider("guided"), &SecurityPolicy{})
	for _, command := range corpus {
		input := closureBashInput(command, f.Root)
		got, err := handler.Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("handle: %v", err)
		}
		want, err := json.Marshal(NewSafeDefaultOutput(permissionModeOf(input)))
		if err != nil {
			t.Fatalf("marshal baseline: %v", err)
		}
		gave, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal got: %v", err)
		}
		if string(gave) != string(want) {
			t.Fatalf("output %s != baseline %s", gave, want)
		}
	}
	if closurePushEvalInvocations != 0 {
		t.Fatalf("A4 evaluations = %d, want 0 under guided", closurePushEvalInvocations)
	}
}
