package template_test

// SPEC-CI-CODEX-REVIEW-001 repair pass (card t1868, sync-audit-1 findings F1, F2, F3, F6).
//
// These tests pin four repairs that change no gate decision: a pull request that is not open ends the
// job before any guard runs (F1); no step interpolates a `${{ }}` expression into its script (F2);
// every outcome of the merge step leaves one explicit line in the job summary (F3); the Lint job
// runs shellcheck over the gate script (F6). Each one was seen red on the unrepaired workflow first;
// the evidence is recorded in the SPEC's progress.md.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// amRepairClean is the fixture on which the whole chain merges: two benign files bound to the head, a
// clean native state comment bound to the head, passing checks, the head stable.
func amRepairClean(t *testing.T) *amRig {
	t.Helper()
	r := newAMRig(t, "real")
	r.e.setPages("files", cgBenignFiles(2, cgHead))
	r.e.setPages("issues", cgNativeIssueComments(t, cgHead))
	r.setPulls(amPullWithIssueComments(cgHead, false, 2, 0, 3))
	return r
}

// amHasTerm reports whether the parsed condition carries steps.<id>.outputs.<name> == '<want>'.
func amHasTerm(t *testing.T, cond, id, name, want string) bool {
	t.Helper()
	terms, err := amParseIf(cond)
	if err != nil {
		t.Fatal(err)
	}
	for _, tm := range terms {
		if tm.id == id && tm.name == name && tm.want == want {
			return true
		}
	}
	return false
}

// F1 — a pull request that is not open ends the job before any guard runs.
//
// The workflow_run payload's pull_requests[] element carries no `state` or `merged` field, so an empty
// value must NOT be read as "not open": the controls open and state_absent_payload must still merge.
// Mutants: the active computation reverted to always true; the condition dropped from the head, author,
// checks, files, review or merge step (each surfaces as a step that runs for a closed pull request, or
// as the static pin below).
func TestAutoMergeSkipsPullRequestThatIsNotOpen(t *testing.T) {
	t.Parallel()
	cgTools(t)
	wf, _ := amLoad(t)

	t.Run("later_steps_carry_the_open_condition", func(t *testing.T) {
		t.Parallel()
		swept := 0
		for _, s := range wf.steps() {
			guarded := s.ID == "head" || s.ID == "author" || s.ID == "checks" || s.ID == "files" || s.ID == "review" ||
				s.Name == amMergeName || strings.HasPrefix(s.Uses, "actions/checkout@")
			if !guarded {
				continue
			}
			swept++
			if !amHasTerm(t, s.If, "pr", "active", "true") {
				t.Errorf("step %q does not require steps.pr.outputs.active == 'true': %q", s.Name, s.If)
			}
		}
		if swept != 7 {
			t.Fatalf("swept %d guarded steps, want 7 (checkout, head, author, checks, files, review, merge)", swept)
		}
	})

	type tc struct {
		name   string
		event  map[string]string
		merges int
		ran    bool
	}
	for _, c := range []tc{
		{"closed_pull_request", map[string]string{"state": "closed", "merged": "false"}, 0, false},
		{"merged_pull_request", map[string]string{"state": "closed", "merged": "true"}, 0, false},
		{"merged_flag_alone", map[string]string{"state": "", "merged": "true"}, 0, false},
		{"no_pull_request_number", map[string]string{"number": ""}, 0, false},
		{"open", map[string]string{"state": "open", "merged": "false"}, 1, true},
		{"state_absent_payload", map[string]string{"state": "", "merged": ""}, 1, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := amRepairClean(t)
			for k, v := range c.event {
				r.ctx.prEvent[k] = v
			}
			if res, _ := r.runStep(r.step("pr")); res.code != 0 {
				t.Fatalf("pr step exit %d\n%s%s", res.code, res.stdout, res.stderr)
			}
			ch := r.runChain()
			if c.ran && (len(ch.ran) == 0 || len(r.mergeCalls()) != c.merges) {
				t.Fatalf("an eligible pull request must still merge: ran=%+v merges=%v", ch.ran, r.mergeCalls())
			}
			if !c.ran {
				if len(ch.ran) != 0 {
					t.Fatalf("steps ran for a pull request that is not open: %+v", ch.ran)
				}
				if calls := r.e.calls(); len(calls) != 0 {
					t.Fatalf("a pull request that is not open made calls: %v", calls)
				}
			}
		})
	}
}

// F2 — no `${{ }}` expression is interpolated into a run script; the merge step takes the pull request
// number through its env like every other step.
// Mutant: `PR_NUMBER="${{ steps.pr.outputs.number }}"` restored in the script (kills all three).
func TestAutoMergeNeverInterpolatesExpressionsIntoScripts(t *testing.T) {
	t.Parallel()
	cgTools(t)
	wf, _ := amLoad(t)
	merge, ok := wf.merge()
	if !ok {
		t.Fatal("merge step missing")
	}

	t.Run("no_run_script_carries_an_expression", func(t *testing.T) {
		t.Parallel()
		swept := 0
		for _, s := range wf.steps() {
			if s.Run == "" {
				continue
			}
			swept++
			if strings.Contains(s.Run, "${{") {
				t.Errorf("step %q interpolates an expression into its script", s.Name)
			}
		}
		if swept < 6 {
			t.Fatalf("swept %d run scripts, want at least 6", swept)
		}
	})
	t.Run("merge_step_env_carries_the_number", func(t *testing.T) {
		t.Parallel()
		if got := merge.Env["PR_NUMBER"]; got != "${{ steps.pr.outputs.number }}" {
			t.Fatalf("merge step env PR_NUMBER = %q", got)
		}
	})
	t.Run("hostile_number_is_data_not_code", func(t *testing.T) {
		t.Parallel()
		r := newAMRig(t, "stub")
		r.seedGuards()
		r.seed("review", "ok", "true")
		marker := filepath.Join(r.e.dir, "injected")
		r.seed("pr", "number", `7"; : > `+marker+`; X="`)
		if _, err := os.Stat(marker); err == nil {
			t.Fatal("marker exists before the step ran")
		}
		r.runStep(merge)
		if _, err := os.Stat(marker); err == nil {
			t.Fatal("the pull request number was executed as shell code")
		}
	})
}

// F3 — every outcome of the merge step leaves one explicit line in the job summary naming the outcome and the
// head SHA. Revision 13 (REQ-CCR-024): "merge requested" ends exit 0; the four reached-and-not-merged outcomes end
// exit 1 (formerly exit 0).
// Mutants: the line of any one outcome deleted (each subtest below kills exactly its own outcome).
func TestAutoMergeSummarisesEveryMergeOutcome(t *testing.T) {
	t.Parallel()
	cgTools(t)
	wf, _ := amLoad(t)
	merge, ok := wf.merge()
	if !ok {
		t.Fatal("merge step missing")
	}
	run := func(t *testing.T, setup func(r *amRig)) (*amRig, cgResult, []string) {
		t.Helper()
		r := newAMRig(t, "stub")
		r.seedGuards()
		r.seed("review", "ok", "true")
		if setup != nil {
			setup(r)
		}
		res, _ := r.runStep(merge)
		raw, err := os.ReadFile(filepath.Join(r.e.dir, "summary-merge.md"))
		if err != nil {
			t.Fatal(err)
		}
		var lines []string
		for _, l := range strings.Split(string(raw), "\n") {
			if strings.TrimSpace(l) != "" {
				lines = append(lines, l)
			}
		}
		return r, res, lines
	}
	// Revision 13 (REQ-CCR-024): the exit status is part of the outcome. Only "merge requested" ends exit 0; the four
	// outcomes where the merge step was reached and did not merge (the deadline before the call, the deadline during
	// the head re-read, the head moved, the merge call failed) end exit 1, and each still leaves its one summary line.
	want := func(t *testing.T, r *amRig, res cgResult, lines []string, prefix string, merges int, shas ...string) {
		t.Helper()
		wantCode := 1
		if prefix == "merge requested" {
			wantCode = 0
		}
		if res.code != wantCode {
			t.Fatalf("exit %d, want %d for %q\n%s%s", res.code, wantCode, prefix, res.stdout, res.stderr)
		}
		if len(r.mergeCalls()) != merges {
			t.Fatalf("merge calls %v, want %d", r.mergeCalls(), merges)
		}
		if len(lines) != 1 {
			t.Fatalf("summary lines %q, want exactly one", lines)
		}
		if !strings.Contains(lines[0], prefix) {
			t.Fatalf("summary line %q lacks %q", lines[0], prefix)
		}
		for _, sha := range shas {
			if !strings.Contains(lines[0], sha) {
				t.Fatalf("summary line %q lacks the head SHA %s", lines[0], sha)
			}
		}
	}

	t.Run("merge_requested", func(t *testing.T) {
		t.Parallel()
		r, res, lines := run(t, nil)
		want(t, r, res, lines, "merge requested", 1, cgHead)
	})
	t.Run("review_window_elapsed_before_merge", func(t *testing.T) {
		t.Parallel()
		r, res, lines := run(t, func(r *amRig) { r.deadline = "1" })
		want(t, r, res, lines, "merge held: review window elapsed", 0, cgHead)
	})
	t.Run("review_window_elapsed_during_head_check", func(t *testing.T) {
		t.Parallel()
		r, res, lines := run(t, func(r *amRig) {
			// The head query takes two real seconds, so a deadline one second ahead passes between the
			// first deadline check and the second.
			realGh := filepath.Join(r.e.bin, "gh.real")
			if err := os.Rename(filepath.Join(r.e.bin, "gh"), realGh); err != nil {
				t.Fatal(err)
			}
			r.e.write(filepath.Join(r.e.bin, "gh"), "#!/bin/sh\nif [ \"${1:-}\" = pr ] && [ \"${2:-}\" = view ]; then /bin/sleep 2; fi\nexec \""+realGh+"\" \"$@\"\n", 0o755)
			r.deadline = fmt.Sprint(time.Now().Unix() + 1)
		})
		want(t, r, res, lines, "merge held: review window elapsed", 0, cgHead)
	})
	t.Run("head_moved", func(t *testing.T) {
		t.Parallel()
		r, res, lines := run(t, func(r *amRig) { r.setViews(cgOther) })
		want(t, r, res, lines, "merge not executed: head moved", 0, cgHead, cgOther)
	})
	t.Run("merge_failed", func(t *testing.T) {
		t.Parallel()
		r, res, lines := run(t, func(r *amRig) { r.setEnv("MERGE_EXIT=1") })
		want(t, r, res, lines, "merge failed:", 1, cgHead)
	})
}

// F6 — the Lint job runs shellcheck over the gate script, as a hard step (no `|| true`, no
// continue-on-error).
// Mutants: the step deleted; the script path changed; `|| true` appended; continue-on-error added.
func TestAutoMergeGateScriptIsShellcheckedInLint(t *testing.T) {
	t.Parallel()
	raw := amReadRepoFile(t, ".github/workflows/ci.yml")
	var wf struct {
		Jobs map[string]struct {
			Name  string   `yaml:"name"`
			Steps []amStep `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(raw), &wf); err != nil {
		t.Fatalf("parse ci.yml: %v", err)
	}
	lint, ok := wf.Jobs["lint"]
	if !ok || lint.Name != "Lint" {
		t.Fatalf("ci.yml has no job lint named Lint: %+v", lint)
	}
	if len(lint.Steps) == 0 {
		t.Fatal("the Lint job swept no steps")
	}
	if strings.TrimSpace(amReadRepoFile(t, "scripts/ci/codex-review-gate.sh")) == "" {
		t.Fatal("the gate script is empty")
	}
	found := false
	for _, s := range lint.Steps {
		if !strings.Contains(s.Run, "shellcheck") {
			continue
		}
		found = true
		if !strings.Contains(s.Run, "scripts/ci/codex-review-gate.sh") {
			t.Errorf("shellcheck step %q does not name the gate script: %q", s.Name, s.Run)
		}
		if strings.Contains(s.Run, "|| true") || strings.Contains(s.Run, "||true") || s.ContinueOnError {
			t.Errorf("shellcheck step %q is not a hard step: run=%q continue-on-error=%t", s.Name, s.Run, s.ContinueOnError)
		}
	}
	if !found {
		t.Fatal("the Lint job has no shellcheck step")
	}
}
