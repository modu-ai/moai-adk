package template_test

// SPEC-CI-CODEX-REVIEW-001 M3 — the auto-merge consumer.
//
// The real .github/workflows/auto-merge.yml is parsed as YAML; its steps are rendered (a small,
// strict `${{ }}` renderer) and executed with bash against a stub `gh`, a no-op `sleep`, and
// either the real gate script or ONE subcommand-aware stub gate script (debt D41).
//
// Stub dispatch table (cgGateStub). The stub dispatches by subcommand and appends one line per
// invocation to CALL_LOG, the same ordered log the stub `gh` and the stub `sleep` write to, so
// the relative order of every gh read, gate call, and sleep is observable:
//
//	fetch-pages         exit code = next value of STUB_FETCH_EXITS (default 0); on 0 it creates the
//	                    --out file (empty), on non-zero it removes it and prints a refusal line
//	gate-files-changed  exit code = next value of STUB_FILES_EXITS (default 0)
//	read-native-review  exit code = next value of STUB_REVIEW_EXITS (default 0); the last value
//	                    repeats once the list is exhausted; 0 prints native-review-clean
//
// The restricted `if:` evaluator (amEvalIf) supports only
// `steps.<id>.outputs.<name> == '<literal>'` terms joined by `&&`; a condition outside that
// grammar is a test failure, so a `&& false` or any new shape cannot pass unnoticed. It is not
// to be widened to make a mutant pass.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const cgGateStub = `#!/bin/sh
if [ -n "${CALL_LOG:-}" ]; then echo "gate $*" >> "$CALL_LOG"; fi
sub="${1:-}"
[ $# -gt 0 ] && shift
next_exit() {
  n=0
  if [ -f "$2" ]; then n=$(cat "$2"); fi
  n=$((n + 1))
  echo "$n" > "$2"
  i=0
  last=0
  for v in $1; do
    i=$((i + 1))
    last=$v
    if [ "$i" -eq "$n" ]; then echo "$v"; return; fi
  done
  echo "$last"
}
case "$sub" in
  fetch-pages)
    out=""
    while [ $# -gt 0 ]; do
      if [ "$1" = "--out" ]; then out="$2"; fi
      shift
    done
    code="$(next_exit "${STUB_FETCH_EXITS:-0}" "$FIXTURES/gate-fetch.count")"
    if [ "$code" -eq 0 ]; then
      : > "$out"
      echo "fetched:0 count-proof:stub"
    else
      rm -f "$out"
      echo "CODEX_GATE_REFUSED: page-error:stub" >&2
    fi
    exit "$code"
    ;;
  gate-files-changed)
    code="$(next_exit "${STUB_FILES_EXITS:-0}" "$FIXTURES/gate-files.count")"
    if [ "$code" -ne 0 ]; then echo "CODEX_GATE_REFUSED: gate-path-changed:stub" >&2; else echo "files-clear: stub"; fi
    exit "$code"
    ;;
  read-native-review)
    code="$(next_exit "${STUB_REVIEW_EXITS:-0}" "$FIXTURES/gate-review.count")"
    case "$code" in
      0) echo "native-review-clean: stub" ;;
      10) echo "CODEX_GATE_REFUSED: pending:no-review" >&2 ;;
      11) echo "CODEX_GATE_REFUSED: not-run:stub" >&2 ;;
      *) echo "CODEX_GATE_REFUSED: stub-exit-$code" >&2 ;;
    esac
    exit "$code"
    ;;
  *) echo "stub gate: unknown subcommand $sub" >&2; exit 98 ;;
esac
`

// --- workflow model -----------------------------------------------------------------------

type amStep struct {
	Name string            `yaml:"name"`
	ID   string            `yaml:"id"`
	If   string            `yaml:"if"`
	Uses string            `yaml:"uses"`
	With map[string]any    `yaml:"with"`
	Env  map[string]string `yaml:"env"`
	Run  string            `yaml:"run"`

	ContinueOnError bool `yaml:"continue-on-error"`
}

type amWorkflow struct {
	On          map[string]any    `yaml:"on"`
	Permissions map[string]string `yaml:"permissions"`
	Jobs        map[string]struct {
		If    string   `yaml:"if"`
		Steps []amStep `yaml:"steps"`
	} `yaml:"jobs"`
}

const amWorkflowRel = ".github/workflows/auto-merge.yml"

func amLoad(t *testing.T) (amWorkflow, string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(amWorkflowRel)))
	if err != nil {
		t.Fatal(err)
	}
	var wf amWorkflow
	if err := yaml.Unmarshal(raw, &wf); err != nil {
		t.Fatalf("parse %s: %v", amWorkflowRel, err)
	}
	return wf, string(raw)
}

func (w amWorkflow) steps() []amStep {
	for _, j := range w.Jobs {
		return j.Steps
	}
	return nil
}

func (w amWorkflow) indexOf(pred func(amStep) bool) int {
	for i, s := range w.steps() {
		if pred(s) {
			return i
		}
	}
	return -1
}

func (w amWorkflow) byID(id string) (amStep, bool) {
	if i := w.indexOf(func(s amStep) bool { return s.ID == id }); i >= 0 {
		return w.steps()[i], true
	}
	return amStep{}, false
}

const amMergeName = "Auto merge PR"

func (w amWorkflow) merge() (amStep, bool) {
	if i := w.indexOf(func(s amStep) bool { return s.Name == amMergeName }); i >= 0 {
		return w.steps()[i], true
	}
	return amStep{}, false
}

var amTermRE = regexp.MustCompile(`^steps\.([A-Za-z0-9_-]+)\.outputs\.([A-Za-z0-9_-]+) == '([^']*)'$`)

type amTerm struct{ id, name, want string }

// amParseCond parses the only grammar the workflow uses: `steps.<id>.outputs.<name> == '<literal>'` terms joined by
// `&&`, optionally led by the status function `always()`, which the harness reads as the exemption from the implied
// `success()` (it exists for mutant MF alone, AC-CCR-019). Anything else is an error.
func amParseCond(cond string) (terms []amTerm, always bool, err error) {
	cond = strings.Join(strings.Fields(cond), " ")
	if cond == "" {
		return nil, false, nil
	}
	if rest, ok := strings.CutPrefix(cond, "always() &&"); ok {
		always, cond = true, strings.TrimSpace(rest)
	}
	for _, raw := range strings.Split(cond, "&&") {
		m := amTermRE.FindStringSubmatch(strings.TrimSpace(raw))
		if m == nil {
			return nil, false, fmt.Errorf("condition term %q is outside the grammar [always() &&] steps.<id>.outputs.<name> == '<literal>' joined by &&", strings.TrimSpace(raw))
		}
		terms = append(terms, amTerm{m[1], m[2], m[3]})
	}
	return terms, always, nil
}

// amParseIf is amParseCond without the status function flag.
func amParseIf(cond string) ([]amTerm, error) {
	terms, _, err := amParseCond(cond)
	return terms, err
}

// --- rig ----------------------------------------------------------------------------------

type amCtx struct {
	outputs map[string]map[string]string
	tested  string
	repo    string
}

type amRig struct {
	t        *testing.T
	e        *cgEnvT
	wf       amWorkflow
	work     string
	ctx      *amCtx
	extra    []string
	deadline string
	results  map[string]cgResult // the last result of each step run, by step id (merge: "merge")
	raw      map[string][]string // the non-empty lines of each step's output file, in order, by step id
	last     amChainResult       // the last chain run
}

const (
	amRepo = "modu-ai/moai-adk"
	amPR   = "7"
)

var amPassChecks = `[{"name":"Lint","state":"SUCCESS","bucket":"pass"}]`

func amPull(head string, draft bool, changed, reviewComments int) string {
	return fmt.Sprintf(`{"head":{"sha":%q,"repo":{"id":1}},"base":{"repo":{"id":1}},"user":{"login":"GoosLab","id":6458969,"type":"User"},"draft":%t,"changed_files":%d,"review_comments":%d,"comments":0}`, head, draft, changed, reviewComments)
}

func amPullWithIssueComments(head string, draft bool, changed, reviewComments, issueComments int) string {
	return strings.Replace(amPull(head, draft, changed, reviewComments), `"comments":0`, fmt.Sprintf(`"comments":%d`, issueComments), 1)
}

func newAMRig(t *testing.T, script string) *amRig {
	t.Helper()
	e := newCGEnv(t)
	wf, _ := amLoad(t)
	r := &amRig{t: t, e: e, wf: wf, work: filepath.Join(e.dir, "work"),
		ctx:      &amCtx{outputs: map[string]map[string]string{}, tested: cgHead, repo: amRepo},
		deadline: fmt.Sprint(time.Now().Unix() + 3600)}
	if err := os.MkdirAll(filepath.Join(r.work, "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	r.installScript(script)
	r.setPulls(amPull(cgHead, false, 2, 0))
	r.setViews(cgHead, cgHead)
	e.write(filepath.Join(e.fixtures, "checks.json"), amPassChecks, 0o600)
	r.ctx.outputs["pr"] = map[string]string{"number": amPR, "branch": "feature/x", "state": "open", "merged": "false"}
	return r
}

// installScript puts the gate script where the checkout would: "real", "stub", or "none".
func (r *amRig) installScript(mode string) {
	r.t.Helper()
	dst := filepath.Join(r.work, "scripts", "ci", "codex-review-gate.sh")
	switch mode {
	case "real":
		raw, err := os.ReadFile(cgScriptPath(r.t))
		if err != nil {
			r.t.Fatal(err)
		}
		r.e.write(dst, string(raw), 0o644)
	case "stub":
		r.e.write(dst, cgGateStub, 0o644)
	case "none":
	default:
		r.t.Fatalf("unknown script mode %q", mode)
	}
}

func (r *amRig) setPulls(entries ...string) {
	r.t.Helper()
	var seq []string
	for i, js := range entries {
		if js == "FAIL" {
			seq = append(seq, "FAIL")
			continue
		}
		name := fmt.Sprintf("pull-%d.json", i+1)
		r.e.write(filepath.Join(r.e.fixtures, name), js, 0o600)
		seq = append(seq, name)
	}
	r.e.write(filepath.Join(r.e.fixtures, "pull-seq"), strings.Join(seq, "\n")+"\n", 0o600)
	_ = os.Remove(filepath.Join(r.e.fixtures, "pull.count"))
}

func (r *amRig) setViews(shas ...string) {
	r.t.Helper()
	r.e.write(filepath.Join(r.e.fixtures, "view-seq"), strings.Join(shas, "\n")+"\n", 0o600)
	_ = os.Remove(filepath.Join(r.e.fixtures, "view.count"))
}

func (r *amRig) seed(id, name, val string) {
	if r.ctx.outputs[id] == nil {
		r.ctx.outputs[id] = map[string]string{}
	}
	r.ctx.outputs[id][name] = val
}

// seedGuards stands in for the head, checks, and files steps when only the review step runs.
func (r *amRig) seedGuards() {
	r.seed("head", "ok", "true")
	r.seed("head", "sha", cgHead)
	r.seed("checks", "should_merge", "true")
	r.seed("files", "ok", "true")
}

func (r *amRig) setEnv(kv ...string) { r.extra = append(r.extra, kv...) }

var amExprRE = regexp.MustCompile(`\$\{\{\s*([^}]*?)\s*\}\}`)

func (r *amRig) render(s string) string {
	return amExprRE.ReplaceAllStringFunc(s, func(m string) string {
		expr := strings.TrimSpace(amExprRE.FindStringSubmatch(m)[1])
		switch expr {
		case "secrets.GITHUB_TOKEN":
			return "test-token"
		case "github.repository":
			return r.ctx.repo
		case "github.event.workflow_run.head_sha":
			return r.ctx.tested
		}
		if parts := strings.Split(expr, "."); len(parts) == 4 && parts[0] == "steps" && parts[2] == "outputs" {
			return r.ctx.outputs[parts[1]][parts[3]]
		}
		r.t.Fatalf("workflow expression %q is not supported by the test renderer", expr)
		return ""
	})
}

func (r *amRig) cond(s amStep) bool {
	ok, _ := r.condStatus(s)
	return ok
}

// condStatus evaluates the output terms of a step's if and reports whether it leads with always().
func (r *amRig) condStatus(s amStep) (terms bool, always bool) {
	ts, always, err := amParseCond(s.If)
	if err != nil {
		r.t.Errorf("step %q: %v", s.Name, err)
		return false, false
	}
	for _, tm := range ts {
		if r.ctx.outputs[tm.id][tm.name] != tm.want {
			return false, always
		}
	}
	return true, always
}

func (r *amRig) step(id string) amStep {
	r.t.Helper()
	s, ok := r.wf.byID(id)
	if !ok {
		r.t.Fatalf("%s has no step with id %q", amWorkflowRel, id)
	}
	return s
}

func amParseOutputs(path string) map[string]string {
	out := map[string]string{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			out[k] = v // the last write wins, as on the runner
		}
	}
	return out
}

func (r *amRig) runStep(s amStep) (cgResult, map[string]string) {
	r.t.Helper()
	key := s.ID
	if key == "" {
		key = "merge"
	}
	scriptPath := filepath.Join(r.e.dir, "step-"+key+".sh")
	outFile := filepath.Join(r.e.dir, "out-"+key+".txt")
	r.e.write(scriptPath, r.render(s.Run), 0o600)
	r.e.write(outFile, "", 0o600)
	env := append(os.Environ(), r.e.env()...)
	for k, v := range s.Env {
		env = append(env, k+"="+r.render(v))
	}
	env = append(env, "GITHUB_OUTPUT="+outFile, "GITHUB_ENV="+filepath.Join(r.e.dir, "env.txt"),
		"RUNNER_TEMP="+filepath.Join(r.work, "tmp"), "merge_deadline="+r.deadline)
	env = append(env, r.extra...)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-e", scriptPath)
	cmd.Dir = r.work
	cmd.Env = env
	var so, se strings.Builder
	cmd.Stdout, cmd.Stderr = &so, &se
	res := cgResult{}
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			r.t.Fatalf("run step %s: %v", key, err)
		}
		res.code = ee.ExitCode()
	}
	res.stdout, res.stderr = so.String(), se.String()
	outputs := amParseOutputs(outFile)
	if s.ID != "" {
		r.ctx.outputs[s.ID] = outputs
	}
	if r.results == nil {
		r.results = map[string]cgResult{}
		r.raw = map[string][]string{}
	}
	r.results[key] = res
	r.raw[key] = nil
	if body, err := os.ReadFile(outFile); err == nil {
		for _, l := range strings.Split(string(body), "\n") {
			if strings.TrimSpace(l) != "" {
				r.raw[key] = append(r.raw[key], l)
			}
		}
	}
	return res, outputs
}

// amExec is one executed step of a chain: its key, its exit status, and whether it is continue-on-error.
type amExec struct {
	key  string
	code int
	coe  bool
}

// amChainResult is what a chain run observed: the steps that ran, in order, and the derived job result.
type amChainResult struct {
	ran []amExec
	job string
}

func (c amChainResult) executed(key string) (amExec, bool) {
	for _, e := range c.ran {
		if e.key == key {
			return e, true
		}
	}
	return amExec{}, false
}

// amDeriveJob is the harness's job-result rule: failure when a step that is not continue-on-error exited non-zero,
// success otherwise. What GitHub reports on a runner is a separate, unobserved fact (spec section 4, research U11).
func amDeriveJob(ran []amExec) string {
	for _, e := range ran {
		if e.code != 0 && !e.coe {
			return "failure"
		}
	}
	return "success"
}

// runChain runs head, checks, files, review, merge in workflow order. A step is skipped when its if is false, or when
// an earlier step failed the job and its if has no status function (the implied success()).
func (r *amRig) runChain() amChainResult {
	r.t.Helper()
	var res amChainResult
	failed := false
	for _, s := range r.wf.steps() {
		inChain := s.ID == "head" || s.ID == "author" || s.ID == "checks" || s.ID == "files" || s.ID == "review" || s.Name == amMergeName
		if !inChain {
			continue
		}
		terms, always := r.condStatus(s)
		if !terms || (failed && !always) {
			continue
		}
		out, _ := r.runStep(s)
		key := s.ID
		if key == "" {
			key = "merge"
		}
		res.ran = append(res.ran, amExec{key: key, code: out.code, coe: s.ContinueOnError})
		if out.code != 0 && !s.ContinueOnError {
			failed = true
		}
	}
	res.job = amDeriveJob(res.ran)
	r.last = res
	return res
}

func (r *amRig) mergeCalls() []string {
	var out []string
	for _, c := range r.e.calls() {
		if strings.HasPrefix(c, "gh pr merge") {
			out = append(out, c)
		}
	}
	return out
}

func (r *amRig) gateCalls(sub string) []string {
	var out []string
	for _, c := range r.e.calls() {
		if strings.HasPrefix(c, "gate "+sub+" ") || c == "gate "+sub {
			out = append(out, c)
		}
	}
	return out
}

func (r *amRig) sleeps() int {
	n := 0
	for _, c := range r.e.calls() {
		if strings.HasPrefix(c, "sleep ") {
			n++
		}
	}
	return n
}

func amFlagValue(call, flag string) string {
	f := strings.Fields(call)
	for i := 0; i+1 < len(f); i++ {
		if f[i] == flag {
			return f[i+1]
		}
	}
	return ""
}

func amRepeat(v string, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = v
	}
	return strings.Join(parts, " ")
}

// --- static scans -------------------------------------------------------------------------

// amGhRE matches `gh` only in command position (line start, after $( or {, after a control operator,
// or after if/then/elif/while/do), so prose such as "(gh exit $rc)" inside an echo string is not a call.
var amGhRE = regexp.MustCompile(`(?:^\s*|\$\(\s*|\{\s*|(?:&&|\|\||;|\|)\s*|\b(?:if|then|elif|while|do)\s+(?:!\s*)?)gh\s+(\S+)(?:\s+(\S+))?(.*)`)

type amGh struct{ group, sub, rest, line string }

// amGhInvocations returns every gh command word in the text outside comment-only lines;
// a backslash continuation joins its lines.
func amGhInvocations(text string) []amGh {
	var out []amGh
	joined := strings.ReplaceAll(text, "\\\n", " ")
	for _, line := range strings.Split(joined, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if m := amGhRE.FindStringSubmatch(line); m != nil {
			out = append(out, amGh{group: m[1], sub: m[2], rest: m[3], line: strings.TrimSpace(line)})
		}
	}
	return out
}

func amReadRepoFile(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// AC-CCR-006 (workflow rows, REQ-CCR-006) — the gh calls are an exact allow-list; no comment, status,
// check-run, or label is read; native issue-comment artifacts are explicitly allowed.
// Mutants: a comment-parse fallback, a status read, a `gh pr comment`, a `gh api` of any other path;
// a scan that matches nothing.
func TestAutoMergeReadsOnlyAllowedReviewArtifacts(t *testing.T) {
	t.Parallel()
	const pullObject = `"repos/$REPO/pulls/$PR_NUMBER"`
	scriptApi := regexp.MustCompile(`^"\$\{?endpoint\}?\?per_page=(100|\$\{?PER_PAGE\}?)&page=\$\{?page\}?"$`)
	forbidden := []string{"/status", "statuses", "check-runs", "check_runs", "labels", "CodeRabbit", "Merge Risk", "AUTOMERGE_MAX_MERGE_RISK"}
	sub := func(name string, fn func(t *testing.T, wf string, gate string)) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			fn(t, amReadRepoFile(t, amWorkflowRel), amReadRepoFile(t, cgScriptRel))
		})
	}
	sub("gh_invocations_match_allow_list_exactly", func(t *testing.T, wf, gate string) {
		for _, inv := range amGhInvocations(wf) {
			switch {
			case inv.group == "pr" && inv.sub == "view":
				if !strings.Contains(inv.rest, "--json headRefOid") {
					t.Errorf("workflow gh pr view is not the headRefOid form: %s", inv.line)
				}
			case inv.group == "pr" && inv.sub == "checks":
				if !strings.Contains(inv.rest, "--required --json name,state,bucket") {
					t.Errorf("workflow gh pr checks is not the required-checks form: %s", inv.line)
				}
			case inv.group == "pr" && inv.sub == "merge":
				if !strings.Contains(inv.line, "--match-head-commit") || strings.Contains(inv.line, "--delete-branch") {
					t.Errorf("workflow gh pr merge is not the head-pinned form: %s", inv.line)
				}
			case inv.group == "api" && inv.sub == pullObject:
				// the pull request object, with no further path segment
			default:
				t.Errorf("workflow gh invocation outside the allow-list: %s", inv.line)
			}
		}
		invs := amGhInvocations(gate)
		if len(invs) != 1 {
			t.Errorf("the gate script makes %d gh invocations, want exactly 1: %+v", len(invs), invs)
		}
		for _, inv := range invs {
			if inv.group != "api" || !scriptApi.MatchString(inv.sub) {
				t.Errorf("gate script gh invocation outside the allow-list: %s", inv.line)
			}
		}
	})
	sub("gh_api_paths_in_workflow_are_pull_object_only", func(t *testing.T, wf, _ string) {
		n := 0
		for _, inv := range amGhInvocations(wf) {
			if inv.group != "api" {
				continue
			}
			n++
			if inv.sub != pullObject {
				t.Errorf("workflow gh api path is not the pull request object: %s", inv.line)
			}
		}
		if n == 0 {
			t.Error("no gh api invocation swept in the workflow")
		}
	})
	sub("retained_gh_pr_forms_view_checks_merge_only", func(t *testing.T, wf, _ string) {
		seen := map[string]int{}
		for _, inv := range amGhInvocations(wf) {
			if inv.group == "pr" {
				seen[inv.sub]++
				if inv.sub != "view" && inv.sub != "checks" && inv.sub != "merge" {
					t.Errorf("gh pr %s is outside {view, checks, merge}: %s", inv.sub, inv.line)
				}
			}
			if inv.group == "issue" || inv.group == "run" || inv.group == "workflow" || inv.group == "label" {
				t.Errorf("gh %s is not allowed: %s", inv.group, inv.line)
			}
		}
		for _, want := range []string{"view", "checks", "merge"} {
			if seen[want] == 0 {
				t.Errorf("gh pr %s not swept", want)
			}
		}
	})
	sub("no_forbidden_token", func(t *testing.T, wf, gate string) {
		for name, text := range map[string]string{"auto-merge.yml": wf, "codex-review-gate.sh": gate} {
			for _, tok := range forbidden {
				if strings.Contains(text, tok) {
					t.Errorf("%s contains the forbidden token %q", name, tok)
				}
			}
		}
	})
	sub("no_coderabbit_or_merge_risk", func(t *testing.T, wf, gate string) {
		for name, text := range map[string]string{"auto-merge.yml": wf, "codex-review-gate.sh": gate} {
			low := strings.ToLower(text)
			for _, tok := range []string{"coderabbit", "merge risk", "merge_risk"} {
				if strings.Contains(low, tok) {
					t.Errorf("%s mentions %q", name, tok)
				}
			}
		}
	})
	sub("no_risk_ceiling_variable", func(t *testing.T, wf, gate string) {
		for name, text := range map[string]string{"auto-merge.yml": wf, "codex-review-gate.sh": gate} {
			for _, tok := range []string{"AUTOMERGE_MAX_MERGE_RISK", "MAX_RISK", "vars."} {
				if strings.Contains(text, tok) {
					t.Errorf("%s still references %q", name, tok)
				}
			}
		}
	})
	sub("no_label_read", func(t *testing.T, wf, gate string) {
		for name, text := range map[string]string{"auto-merge.yml": wf, "codex-review-gate.sh": gate} {
			for _, tok := range []string{"labels", "--label", "gh label", ".labels"} {
				if strings.Contains(text, tok) {
					t.Errorf("%s reads a label: %q", name, tok)
				}
			}
		}
	})
	sub("scan_swept_at_least_6_gh_invocations", func(t *testing.T, wf, _ string) {
		pr, api := 0, 0
		for _, inv := range amGhInvocations(wf) {
			switch inv.group {
			case "pr":
				pr++
			case "api":
				api++
			}
		}
		if pr < 4 || api < 2 || pr+api < 6 {
			t.Fatalf("scan swept %d gh pr and %d gh api invocations; want at least 4 and 2", pr, api)
		}
	})
}

// AC-CCR-007 (loop rows, REQ-CCR-007) — the review step's wait is bounded and never ends in a pass.
// Mutants: expiry that authorizes a merge; a refusal read as pending; a stale review waited on for the
// whole window; an unbounded loop; a last pass that does not pass --wait-closed true; an unreadable
// list at expiry reported as NOT RUN.
func TestAutoMergeWaitLoopIsBounded(t *testing.T) {
	t.Parallel()
	type sub struct {
		name string
		run  func(t *testing.T)
	}
	reviewRig := func(t *testing.T, script string, env ...string) (*amRig, map[string]string, cgResult) {
		t.Helper()
		r := newAMRig(t, script)
		r.seedGuards()
		r.setEnv(env...)
		res, outs := r.runStep(r.step("review"))
		return r, outs, res
	}
	subs := []sub{
		{"loop_bound_at_most_60", func(t *testing.T) {
			r, outs, _ := reviewRig(t, "stub", "STUB_REVIEW_EXITS=10")
			if n := len(r.gateCalls("read-native-review")); n != 60 {
				t.Fatalf("read-native-review ran %d times, want exactly 60", n)
			}
			if outs["ok"] == "true" {
				t.Fatal("a window that never saw a review ended in ok=true")
			}
		}},
		{"pending_keeps_waiting", func(t *testing.T) {
			r, outs, res := reviewRig(t, "stub", "STUB_REVIEW_EXITS=10 10 10 0")
			if res.code != 0 {
				t.Fatalf("step exit %d, want 0 on the clean pass", res.code)
			}
			// Four passes plus the one pre-merge recheck of the clean pass (ruling B).
			if n := len(r.gateCalls("read-native-review")); n != 5 {
				t.Fatalf("read-native-review ran %d times, want 5 (4 passes + 1 pre-merge recheck)", n)
			}
			if outs["ok"] != "true" || outs["state"] != "pass" || r.sleeps() != 3 {
				t.Fatalf("ok=%q state=%q sleeps=%d, want true/pass/3", outs["ok"], outs["state"], r.sleeps())
			}
		}},
		{"last_pass_passes_wait_closed_true", func(t *testing.T) {
			r, _, _ := reviewRig(t, "stub", "STUB_REVIEW_EXITS=10")
			calls := r.gateCalls("read-native-review")
			for i, c := range calls {
				want := "false"
				if i == len(calls)-1 {
					want = "true"
				}
				if got := amFlagValue(c, "--wait-closed"); got != want {
					t.Fatalf("pass %d passed --wait-closed %q, want %q: %s", i+1, got, want, c)
				}
			}
		}},
		{"no_sleep_after_last_pass", func(t *testing.T) {
			r, _, _ := reviewRig(t, "stub", "STUB_REVIEW_EXITS=10")
			if n := r.sleeps(); n != 59 {
				t.Fatalf("slept %d times across 60 passes, want 59", n)
			}
			calls := r.e.calls()
			if last := calls[len(calls)-1]; !strings.HasPrefix(last, "gate read-native-review") {
				t.Fatalf("the last call is %q, a sleep must not follow the last pass", last)
			}
		}},
		{"exit_one_withholds_at_once", func(t *testing.T) {
			r, outs, res := reviewRig(t, "stub", "STUB_REVIEW_EXITS=1")
			if n := len(r.gateCalls("read-native-review")); n != 1 || r.sleeps() != 0 {
				t.Fatalf("reads=%d sleeps=%d, want 1 and 0", n, r.sleeps())
			}
			if res.code != 1 {
				t.Fatalf("step exit %d, want 1 (REQ-CCR-018: a refusal ends the step non-zero)", res.code)
			}
			if outs["ok"] == "true" || outs["state"] != "fail" {
				t.Fatalf("ok=%q state=%q, want not-true/fail", outs["ok"], outs["state"])
			}
		}},
		{"stale_withholds_at_once", func(t *testing.T) {
			r := newAMRig(t, "real")
			r.seedGuards()
			r.e.setPages("reviews", []any{cgReview(10, cgOther, "APPROVED")})
			r.setPulls(amPull(cgHead, false, 2, 0))
			res, outs := r.runStep(r.step("review"))
			if outs["ok"] == "true" || outs["state"] != "fail" || r.sleeps() != 0 {
				t.Fatalf("ok=%q state=%q sleeps=%d, want not-true/fail/0\n%s%s", outs["ok"], outs["state"], r.sleeps(), res.stdout, res.stderr)
			}
			if !strings.Contains(res.stdout+res.stderr, "stale-review") {
				t.Fatalf("the step does not name the stale review:\n%s%s", res.stdout, res.stderr)
			}
		}},
		{"fetch_failure_counts_pending", func(t *testing.T) {
			r, outs, _ := reviewRig(t, "stub", "STUB_FETCH_EXITS=1 1 0 0", "STUB_REVIEW_EXITS=0")
			if outs["ok"] != "true" || outs["state"] != "pass" {
				t.Fatalf("ok=%q state=%q after two unreadable passes, want true/pass", outs["ok"], outs["state"])
			}
			// One clean pass plus the one pre-merge recheck (ruling B); the two unreadable passes read nothing.
			if n := len(r.gateCalls("read-native-review")); n != 2 || r.sleeps() != 2 {
				t.Fatalf("reads=%d sleeps=%d, want 2 (1 pass + 1 pre-merge recheck) and 2", n, r.sleeps())
			}
		}},
		{"unreadable_to_end_of_window_fails_review_unreadable", func(t *testing.T) {
			r, outs, res := reviewRig(t, "stub", "STUB_FETCH_EXITS=1")
			all := res.stdout + res.stderr
			if res.code != 1 {
				t.Fatalf("step exit %d, want 1 (a list unreadable to the end of the window fails the step)", res.code)
			}
			if outs["ok"] == "true" || outs["state"] != "fail" {
				t.Fatalf("ok=%q state=%q, want not-true/fail", outs["ok"], outs["state"])
			}
			if !strings.Contains(all, "review-unreadable") || strings.Contains(all, "NOT RUN") {
				t.Fatalf("an unreadable window must report review-unreadable and never NOT RUN:\n%s", all)
			}
			if n := len(r.gateCalls("read-native-review")); n != 0 {
				t.Fatalf("read-native-review ran %d times on lists that were never readable", n)
			}
		}},
		{"review_arriving_after_window_not_read", func(t *testing.T) {
			// The 61st value would be a clean review; the window ends first and it is never read.
			exits := amRepeat("10", 59) + " 11 0"
			r, outs, res := reviewRig(t, "stub", "STUB_REVIEW_EXITS="+exits)
			if n := len(r.gateCalls("read-native-review")); n != 60 {
				t.Fatalf("read-native-review ran %d times, want 60", n)
			}
			if outs["ok"] == "true" || strings.Contains(res.stdout, "native-review-clean") {
				t.Fatalf("a review arriving after the window was read: ok=%q\n%s", outs["ok"], res.stdout)
			}
		}},
		{"expiry_never_merges", func(t *testing.T) {
			for _, exits := range []string{"10", amRepeat("10", 59) + " 11"} {
				r, outs, _ := reviewRig(t, "stub", "STUB_REVIEW_EXITS="+exits)
				if outs["ok"] == "true" || outs["state"] == "pass" {
					t.Fatalf("exits %q: the end of the window ended ok=%q state=%q", exits, outs["ok"], outs["state"])
				}
				merge, ok := r.wf.merge()
				if !ok {
					t.Fatal("merge step missing")
				}
				if r.cond(merge) {
					t.Fatalf("exits %q: the merge condition is true at the end of the window", exits)
				}
				if n := len(r.mergeCalls()); n != 0 {
					t.Fatalf("exits %q: gh pr merge ran %d times", exits, n)
				}
			}
		}},
	}
	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			s.run(t)
		})
	}
}

// AC-CCR-008 (step rows, REQ-CCR-008) — NOT RUN is one named result, withheld, with its own state and notice.
// Mutants: a loop that proceeds on 11; a state output of fail for NOT RUN; a message that claims a review result.
func TestAutoMergeNotRunWithholds(t *testing.T) {
	t.Parallel()
	type sub struct {
		name string
		run  func(t *testing.T)
	}
	notRun := func(t *testing.T, draft bool, script string) (*amRig, map[string]string, cgResult) {
		t.Helper()
		r := newAMRig(t, script)
		r.seedGuards()
		r.setPulls(amPull(cgHead, draft, 2, 0))
		r.setEnv("STUB_REVIEW_EXITS=11")
		res, outs := r.runStep(r.step("review"))
		return r, outs, res
	}
	subs := []sub{
		// D39: this row, not review_step_unknown_exit_not_ok, kills the exit-11 half of the AC-CCR-012
		// mutant "treating exit 11 or an unknown exit as pass". Revision 6: on a non-draft pull request exit 11 is
		// the failure not-run:review-absent (T-CHAIN row 4): step exit 1, state fail.
		{"exit_11_withholds_merge", func(t *testing.T) {
			r, outs, res := notRun(t, false, "stub")
			if outs["ok"] == "true" || outs["state"] != "fail" || res.code != 1 {
				t.Fatalf("ok=%q state=%q exit=%d, want not-true/fail/1", outs["ok"], outs["state"], res.code)
			}
			merge, _ := r.wf.merge()
			if r.cond(merge) {
				t.Fatal("the merge condition is true after exit 11")
			}
		}},
		{"not_run_message_draft_names_cause", func(t *testing.T) {
			_, _, res := notRun(t, true, "real")
			if !strings.Contains(res.stdout+res.stderr, "NOT RUN: draft") {
				t.Fatalf("the notice does not name the cause draft:\n%s%s", res.stdout, res.stderr)
			}
		}},
		{"not_run_message_review_absent_names_cause", func(t *testing.T) {
			_, _, res := notRun(t, false, "stub")
			if !strings.Contains(res.stdout+res.stderr, "::error::native review gate: not-run:review-absent") {
				t.Fatalf("the annotation does not name the cause not-run:review-absent as an error:\n%s%s", res.stdout, res.stderr)
			}
		}},
		{"not_run_message_never_a_pass", func(t *testing.T) {
			_, _, res := notRun(t, true, "real")
			if !strings.Contains(res.stdout+res.stderr, "never a pass") {
				t.Fatalf("the draft notice does not say it is never a pass:\n%s%s", res.stdout, res.stderr)
			}
		}},
		{"not_run_message_gate_path_only", func(t *testing.T) {
			_, _, res := notRun(t, true, "real")
			all := res.stdout + res.stderr
			if !strings.Contains(all, "not a review result") {
				t.Fatalf("the draft notice does not say it is not a review result:\n%s", all)
			}
		}},
		{"not_run_state_output_distinct_from_fail", func(t *testing.T) {
			_, outs, _ := notRun(t, true, "real")
			if outs["state"] != "not-run:draft" || outs["state"] == "fail" || outs["state"] == "pass" {
				t.Fatalf("state=%q, want not-run:draft", outs["state"])
			}
		}},
		{"exit_11_never_reaches_merge_condition", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_REVIEW_EXITS=11")
			r.runChain()
			if n := len(r.mergeCalls()); n != 0 {
				t.Fatalf("gh pr merge ran %d times after exit 11", n)
			}
			if r.ctx.outputs["review"]["ok"] == "true" {
				t.Fatal("exit 11 produced ok=true")
			}
			if n := strings.Count(r.step("review").Run, "conclude true "); n != 1 {
				t.Fatalf("the review step concludes ok=true in %d places, want exactly one (the clean pass)", n)
			}
		}},
		{"draft_read_from_pull_request_object", func(t *testing.T) {
			for _, draft := range []bool{true, false} {
				r := newAMRig(t, "stub")
				r.seedGuards()
				r.setPulls(amPull(cgHead, draft, 2, 0))
				r.setEnv("STUB_REVIEW_EXITS=11")
				r.runStep(r.step("review"))
				calls := r.gateCalls("read-native-review")
				if len(calls) == 0 {
					t.Fatalf("draft=%t: read-native-review never ran", draft)
				}
				if got := amFlagValue(calls[0], "--draft"); got != fmt.Sprint(draft) {
					t.Fatalf("draft=%t: the step passed --draft %q: %s", draft, got, calls[0])
				}
			}
		}},
		{"draft_skips_list_fetch_and_is_not_run", func(t *testing.T) {
			r, outs, _ := notRun(t, true, "stub")
			if n := len(r.gateCalls("fetch-pages")); n != 0 {
				t.Fatalf("a draft fetched %d lists, want none", n)
			}
			if outs["state"] != "not-run:draft" {
				t.Fatalf("state=%q, want not-run:draft", outs["state"])
			}
		}},
	}
	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			s.run(t)
		})
	}
}

// AC-CCR-011 (REQ-CCR-011, regression-guard) — the checkout, head re-verification, required-check
// observation, deadline re-evaluation, and head-pinned merge are untouched.
// Mutants: `ref: ${{ github.event.workflow_run.head_sha }}` on the checkout; removing --match-head-commit.
func TestAutoMergeCheckoutAndGuardsUnchanged(t *testing.T) {
	t.Parallel()
	wf, _ := amLoad(t)
	t.Run("checkout_has_no_ref", func(t *testing.T) {
		i := wf.indexOf(func(s amStep) bool { return strings.HasPrefix(s.Uses, "actions/checkout@") })
		if i < 0 {
			t.Fatal("no checkout step")
		}
		if _, has := wf.steps()[i].With["ref"]; has {
			t.Fatal("the checkout step carries a ref")
		}
	})
	t.Run("head_reverify_present", func(t *testing.T) {
		s, ok := wf.byID("head")
		if !ok || !strings.Contains(s.Run, "gh pr view") || !strings.Contains(s.Run, `"$TESTED_SHA" != "$CURRENT_SHA"`) {
			t.Fatal("the head re-verification step is missing or changed")
		}
	})
	t.Run("required_checks_step_name_present", func(t *testing.T) {
		if wf.indexOf(func(s amStep) bool { return s.Name == "Check all required CI checks passed" }) < 0 {
			t.Fatal("the required-checks step lost its name")
		}
	})
	t.Run("match_head_commit_present", func(t *testing.T) {
		m, ok := wf.merge()
		if !ok || !strings.Contains(m.Run, `--match-head-commit "$TESTED_SHA"`) {
			t.Fatal("the merge call is not pinned with --match-head-commit \"$TESTED_SHA\"")
		}
	})
	t.Run("deadline_reeval_present", func(t *testing.T) {
		m, _ := wf.merge()
		if n := strings.Count(m.Run, `"${merge_deadline:-9999999999}"`); n < 2 {
			t.Fatalf("the merge step re-evaluates the deadline %d times, want at least 2", n)
		}
	})
}

// AC-CCR-012 (REQ-CCR-012) — the real workflow wires the files step and the review step, in order, ahead of
// the merge, and each treats every exit it does not recognise as a withheld merge.
// Mutants (each applied to the committed workflow): the files step deleted; the guard's exit code discarded;
// a files step that never re-reads the head; a review step that reads the head only before its list reads;
// a review step that ignores the exit code; exit 11 or an unknown exit treated as pass (exit 11 is killed by
// AC-CCR-008 exit_11_withholds_merge, an unknown exit by review_step_unknown_exit_not_ok); a merge if without
// the review term or without the files term (pinned structurally here); the merge step before the gate steps;
// the count arguments or the head SHA omitted.
func TestAutoMergeWiresGuardAndGate(t *testing.T) {
	t.Parallel()
	wf, _ := amLoad(t)
	filesStep := func(t *testing.T) amStep {
		s, ok := wf.byID("files")
		if !ok {
			t.Fatal("the files step is absent")
		}
		return s
	}
	reviewStep := func(t *testing.T) amStep {
		s, ok := wf.byID("review")
		if !ok {
			t.Fatal("the review step is absent")
		}
		return s
	}
	hasTerm := func(t *testing.T, cond string, id, name, want string) {
		t.Helper()
		terms, err := amParseIf(cond)
		if err != nil {
			t.Fatal(err)
		}
		for _, tm := range terms {
			if tm.id == id && tm.name == name && tm.want == want {
				return
			}
		}
		t.Fatalf("condition %q lacks steps.%s.outputs.%s == '%s'", cond, id, name, want)
	}
	runFiles := func(t *testing.T, env ...string) (*amRig, map[string]string) {
		t.Helper()
		r := newAMRig(t, "stub")
		r.seedGuards()
		r.seed("files", "ok", "")
		r.setEnv(env...)
		_, outs := r.runStep(r.step("files"))
		return r, outs
	}
	exitOf := func(r *amRig, id string) int { return r.results[id].code }
	type sub struct {
		name string
		run  func(t *testing.T)
	}
	subs := []sub{
		{"files_step_present", func(t *testing.T) {
			if s := filesStep(t); s.Name != "Collect and check the changed-file list" {
				t.Fatalf("files step name %q", s.Name)
			}
		}},
		{"review_step_present", func(t *testing.T) {
			if s := reviewStep(t); s.Name != "Read the native Codex review" {
				t.Fatalf("review step name %q", s.Name)
			}
		}},
		{"order_checks_then_files_then_review_then_merge", func(t *testing.T) {
			idx := func(pred func(amStep) bool) int { return wf.indexOf(pred) }
			head := idx(func(s amStep) bool { return s.ID == "head" })
			checks := idx(func(s amStep) bool { return s.ID == "checks" })
			files := idx(func(s amStep) bool { return s.ID == "files" })
			review := idx(func(s amStep) bool { return s.ID == "review" })
			merge := idx(func(s amStep) bool { return s.Name == amMergeName })
			if head < 0 || head >= checks || checks >= files || files >= review || review >= merge || merge != len(wf.steps())-1 {
				t.Fatalf("step order head=%d checks=%d files=%d review=%d merge=%d (of %d)", head, checks, files, review, merge, len(wf.steps()))
			}
		}},
		{"files_step_depends_on_head_and_checks_ok", func(t *testing.T) {
			s := filesStep(t)
			hasTerm(t, s.If, "head", "ok", "true")
			hasTerm(t, s.If, "checks", "should_merge", "true")
		}},
		{"review_step_depends_on_files_ok", func(t *testing.T) {
			s := reviewStep(t)
			hasTerm(t, s.If, "files", "ok", "true")
			hasTerm(t, s.If, "head", "ok", "true")
			hasTerm(t, s.If, "checks", "should_merge", "true")
		}},
		{"merge_if_requires_files_and_review_ok", func(t *testing.T) {
			m, ok := wf.merge()
			if !ok {
				t.Fatal("merge step missing")
			}
			hasTerm(t, m.If, "files", "ok", "true")
			hasTerm(t, m.If, "review", "ok", "true")
			hasTerm(t, m.If, "head", "ok", "true")
			hasTerm(t, m.If, "checks", "should_merge", "true")
		}},
		{"steps_run_script_from_default_checkout", func(t *testing.T) {
			co := wf.indexOf(func(s amStep) bool { return strings.HasPrefix(s.Uses, "actions/checkout@") })
			for _, id := range []string{"files", "review"} {
				s, ok := wf.byID(id)
				if !ok || !strings.Contains(s.Run, "scripts/ci/codex-review-gate.sh") {
					t.Fatalf("step %s does not run scripts/ci/codex-review-gate.sh", id)
				}
				if wf.indexOf(func(x amStep) bool { return x.ID == id }) < co {
					t.Fatalf("step %s runs before the checkout", id)
				}
			}
			if _, has := wf.steps()[co].With["ref"]; has {
				t.Fatal("the checkout carries a ref")
			}
		}},
		{"files_step_passes_changed_files_count", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgHead, false, 7, 3))
			r.runStep(r.step("files"))
			calls := r.gateCalls("fetch-pages")
			if len(calls) != 1 || !strings.Contains(calls[0], "--endpoint repos/modu-ai/moai-adk/pulls/7/files") || amFlagValue(calls[0], "--expect-count") != "7" {
				t.Fatalf("fetch-pages calls %v, want one files fetch with --expect-count 7", calls)
			}
		}},
		{"review_step_passes_review_comments_count", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgHead, false, 7, 5))
			r.runStep(r.step("review"))
			var reviews, comments string
			for _, c := range r.gateCalls("fetch-pages") {
				switch {
				case strings.Contains(c, "/pulls/7/reviews"):
					reviews = c
				case strings.Contains(c, "/pulls/7/comments"):
					comments = c
				}
			}
			if reviews == "" || strings.Contains(reviews, "--expect-count") {
				t.Fatalf("reviews fetch %q: want present and without a count", reviews)
			}
			if amFlagValue(comments, "--expect-count") != "5" {
				t.Fatalf("comments fetch %q: want --expect-count 5", comments)
			}
		}},
		{"files_step_passes_verified_head_sha", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.runStep(r.step("files"))
			calls := r.gateCalls("gate-files-changed")
			if len(calls) != 1 || amFlagValue(calls[0], "--head-sha") != cgHead {
				t.Fatalf("gate-files-changed calls %v, want one with --head-sha %s", calls, cgHead)
			}
		}},
		{"permissions_unchanged", func(t *testing.T) {
			if len(wf.Permissions) != 2 || wf.Permissions["contents"] != "write" || wf.Permissions["pull-requests"] != "write" {
				t.Fatalf("permissions %v, want exactly contents: write and pull-requests: write", wf.Permissions)
			}
		}},
		{"files_step_workflow_edit_not_ok", func(t *testing.T) {
			r, outs := runFiles(t, "STUB_FILES_EXITS=1")
			if outs["ok"] == "true" || outs["state"] != "fail" || exitOf(r, "files") != 1 {
				t.Fatalf("ok=%q state=%q exit=%d, want not-true/fail/1", outs["ok"], outs["state"], exitOf(r, "files"))
			}
		}},
		{"files_step_clear_list_ok_true", func(t *testing.T) {
			r, outs := runFiles(t)
			if outs["ok"] != "true" || outs["state"] != "pass" || exitOf(r, "files") != 0 {
				t.Fatalf("ok=%q state=%q exit=%d, want true/pass/0", outs["ok"], outs["state"], exitOf(r, "files"))
			}
		}},
		{"files_step_page_error_not_ok", func(t *testing.T) {
			r, outs := runFiles(t, "STUB_FETCH_EXITS=1")
			if outs["ok"] == "true" || len(r.gateCalls("gate-files-changed")) != 0 || exitOf(r, "files") != 1 {
				t.Fatalf("ok=%q exit=%d, gate calls %v: a failed page read must stop the step with exit 1", outs["ok"], exitOf(r, "files"), r.gateCalls("gate-files-changed"))
			}
		}},
		{"files_step_missing_script_not_ok", func(t *testing.T) {
			r := newAMRig(t, "none")
			r.seedGuards()
			res, outs := r.runStep(r.step("files"))
			if res.code != 1 || outs["ok"] == "true" || outs["state"] != "fail" {
				t.Fatalf("exit=%d ok=%q state=%q, want 1, not-true, fail: a missing script is a failure, never a pass", res.code, outs["ok"], outs["state"])
			}
		}},
		{"files_step_head_moved_before_read_not_ok", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgOther, false, 2, 0))
			res, outs := r.runStep(r.step("files"))
			if outs["ok"] == "true" || len(r.gateCalls("fetch-pages")) != 0 || res.code != 1 {
				t.Fatalf("ok=%q exit=%d fetches=%v: a moved head must stop the step with exit 1 before any list read", outs["ok"], res.code, r.gateCalls("fetch-pages"))
			}
		}},
		{"files_step_head_moved_during_read_not_ok", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgHead, false, 2, 0), amPull(cgOther, false, 2, 0))
			res, outs := r.runStep(r.step("files"))
			if outs["ok"] == "true" || len(r.gateCalls("gate-files-changed")) != 0 || res.code != 1 {
				t.Fatalf("ok=%q exit=%d verdict calls=%v: a head that moved during the list read must never be judged", outs["ok"], res.code, r.gateCalls("gate-files-changed"))
			}
		}},
		{"files_step_count_moved_during_read_not_ok", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgHead, false, 3, 0), amPull(cgHead, false, 4, 0))
			res, outs := r.runStep(r.step("files"))
			if outs["ok"] == "true" || len(r.gateCalls("gate-files-changed")) != 0 || res.code != 1 {
				t.Fatalf("ok=%q exit=%d verdict calls=%v: a changed count during the read must stop the step", outs["ok"], res.code, r.gateCalls("gate-files-changed"))
			}
		}},
		{"review_step_clean_ok_true_state_pass", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			res, outs := r.runStep(r.step("review"))
			if outs["ok"] != "true" || outs["state"] != "pass" || res.code != 0 {
				t.Fatalf("ok=%q state=%q exit=%d, want true/pass/0", outs["ok"], outs["state"], res.code)
			}
		}},
		{"review_step_exit_1_not_ok_state_fail", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setEnv("STUB_REVIEW_EXITS=1")
			res, outs := r.runStep(r.step("review"))
			if outs["ok"] == "true" || outs["state"] != "fail" || res.code != 1 {
				t.Fatalf("ok=%q state=%q exit=%d, want not-true/fail/1", outs["ok"], outs["state"], res.code)
			}
		}},
		{"review_step_unknown_exit_not_ok", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setEnv("STUB_REVIEW_EXITS=7")
			res, outs := r.runStep(r.step("review"))
			if outs["ok"] == "true" || outs["state"] != "fail" || res.code != 1 || !strings.Contains(res.stdout+res.stderr, "::error::native review gate: gate-exit-7") {
				t.Fatalf("ok=%q state=%q exit=%d, want not-true/fail/1 with an ::error:: naming gate-exit-7 for an unrecognised exit\n%s%s", outs["ok"], outs["state"], res.code, res.stdout, res.stderr)
			}
		}},
		{"review_step_head_moved_not_ok", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgOther, false, 2, 0))
			res, outs := r.runStep(r.step("review"))
			if outs["ok"] == "true" || outs["state"] != "fail" || len(r.gateCalls("fetch-pages")) != 0 || res.code != 1 {
				t.Fatalf("ok=%q state=%q exit=%d fetches=%v, want not-true/fail/1 with no list read", outs["ok"], outs["state"], res.code, r.gateCalls("fetch-pages"))
			}
		}},
		{"review_step_head_moved_during_read_not_ok", func(t *testing.T) {
			// The stub gh answers the pull request object with the verified head on the first read of a pass
			// and with another SHA on the second, beside lists that are the clean fixture for the verified
			// head. The stub gate records each subcommand; fetch-pages before the second read is expected,
			// a read-native-review (verdict) invocation after it is the defect.
			r := newAMRig(t, "stub")
			r.seedGuards()
			r.setPulls(amPull(cgHead, false, 2, 0), amPull(cgOther, false, 2, 0))
			res, outs := r.runStep(r.step("review"))
			if outs["ok"] == "true" || outs["state"] != "fail" || res.code != 1 {
				t.Fatalf("ok=%q state=%q exit=%d, want not-true/fail/1", outs["ok"], outs["state"], res.code)
			}
			calls := r.e.calls()
			pullReads, secondRead, fetches := 0, -1, 0
			for i, c := range calls {
				switch {
				case c == "gh api repos/modu-ai/moai-adk/pulls/7":
					pullReads++
					if pullReads == 2 {
						secondRead = i
					}
				case strings.HasPrefix(c, "gate fetch-pages"):
					fetches++
					if pullReads < 1 || pullReads >= 2 {
						t.Fatalf("fetch-pages at call %d is not between the two head reads: %v", i, calls)
					}
				case strings.HasPrefix(c, "gate read-native-review"):
					t.Fatalf("read-native-review ran (call %d) although the head moved during the list read: %v", i, calls)
				}
			}
			if pullReads != 2 || secondRead < 0 || fetches != 3 {
				t.Fatalf("pull reads=%d fetches=%d, want 2 and 3: %v", pullReads, fetches, calls)
			}
		}},
		{"review_step_forged_review_never_ok", func(t *testing.T) {
			r := newAMRig(t, "real")
			r.seedGuards()
			// Another author's bound APPROVED review is never the identity's.
			r.e.setPages("reviews", []any{cgReview(10, cgHead, "APPROVED", cgHuman, cgBody("approved"))})
			_, outs := r.runStep(r.step("review"))
			if outs["ok"] == "true" {
				t.Fatalf("a forged review produced ok=true (state=%q)", outs["state"])
			}
		}},
	}
	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			s.run(t)
		})
	}
}

// AC-CCR-013 (REQ-CCR-013, regression-guard) — workflow_run only, the workflow token only, no artifact or cache.
// Mutants: adding pull_request_target, ${{ secrets.X }}, an actions/cache step.
func TestAutoMergeAddsNoSecretTriggerOrArtifact(t *testing.T) {
	t.Parallel()
	wf, raw := amLoad(t)
	t.Run("trigger_is_workflow_run_only", func(t *testing.T) {
		if len(wf.On) != 1 {
			t.Fatalf("triggers %v, want exactly workflow_run", wf.On)
		}
		if _, ok := wf.On["workflow_run"]; !ok {
			t.Fatalf("triggers %v lack workflow_run", wf.On)
		}
	})
	t.Run("no_pull_request_target_family", func(t *testing.T) {
		if strings.Contains(raw, "pull_request_target") {
			t.Fatal("auto-merge.yml names pull_request_target")
		}
	})
	t.Run("secrets_only_workflow_token", func(t *testing.T) {
		refs := regexp.MustCompile(`secrets\.[A-Za-z0-9_]+`).FindAllString(raw, -1)
		if len(refs) == 0 {
			t.Fatal("no secret reference found; the scan swept nothing")
		}
		for _, ref := range refs {
			if ref != "secrets.GITHUB_TOKEN" {
				t.Errorf("secret reference %q is not the workflow token", ref)
			}
		}
	})
	t.Run("no_artifact_or_cache_step", func(t *testing.T) {
		for _, tok := range []string{"actions/upload-artifact", "actions/download-artifact", "actions/cache"} {
			if strings.Contains(raw, tok) {
				t.Errorf("auto-merge.yml uses %s", tok)
			}
		}
		for _, s := range wf.steps() {
			if _, has := s.With["cache"]; has {
				t.Errorf("step %q has a cache input", s.Name)
			}
		}
	})
}

func TestAutoMergeAuthorEligibility(t *testing.T) {
	t.Parallel()
	cgTools(t)
	raw, _ := amLoad(t)
	idx := func(id string) int { return raw.indexOf(func(s amStep) bool { return s.ID == id }) }
	merge, ok := raw.merge()
	if !ok || idx("author") < 0 || idx("author") >= idx("checks") || idx("checks") >= idx("files") || idx("files") >= idx("review") || idx("review") >= raw.indexOf(func(s amStep) bool { return s.Name == amMergeName }) {
		t.Fatalf("author/check/files/review/merge order is incomplete: author=%d checks=%d files=%d review=%d merge=%d", idx("author"), idx("checks"), idx("files"), idx("review"), raw.indexOf(func(s amStep) bool { return s.Name == amMergeName }))
	}
	for _, id := range []string{"checks", "files", "review"} {
		s, _ := raw.byID(id)
		if !strings.Contains(s.If, "steps.author.outputs.eligible == 'true'") {
			t.Errorf("%s step does not require eligible author: %s", id, s.If)
		}
	}
	if !strings.Contains(merge.If, "steps.author.outputs.eligible == 'true'") {
		t.Errorf("merge step does not require eligible author: %s", merge.If)
	}

	for _, tc := range []struct {
		name   string
		mutate func(string) string
		fail   bool
	}{
		{name: "eligible_same_repository_author"},
		{name: "wrong_login", mutate: func(s string) string { return strings.Replace(s, `"login":"GoosLab"`, `"login":"other"`, 1) }, fail: true},
		{name: "wrong_id", mutate: func(s string) string { return strings.Replace(s, `"id":6458969`, `"id":7`, 1) }, fail: true},
		{name: "bot_author", mutate: func(s string) string { return strings.Replace(s, `"type":"User"`, `"type":"Bot"`, 1) }, fail: true},
		{name: "fork_head", mutate: func(s string) string {
			return strings.Replace(s, `"base":{"repo":{"id":1}}`, `"base":{"repo":{"id":2}}`, 1)
		}, fail: true},
		{name: "unreadable_pull", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newAMRig(t, "stub")
			pull := amPull(cgHead, false, 2, 0)
			if tc.mutate != nil {
				pull = tc.mutate(pull)
			}
			if tc.name == "unreadable_pull" {
				r.setPulls("FAIL")
			} else {
				r.setPulls(pull)
			}
			result := r.runChain()
			got := r.ctx.outputs["author"]["eligible"]
			if tc.fail {
				if got != "false" || len(r.mergeCalls()) != 0 {
					t.Fatalf("ineligible pull: eligible=%q merges=%v", got, r.mergeCalls())
				}
				for _, step := range []string{"checks", "files", "review", "merge"} {
					if _, ran := result.executed(step); ran {
						t.Errorf("%s ran for ineligible pull", step)
					}
				}
				if !strings.Contains(r.results["author"].stdout+r.results["author"].stderr, "test confirmation and explicit approval") {
					t.Errorf("ineligible pull lacks operator notice: %s%s", r.results["author"].stdout, r.results["author"].stderr)
				}
				return
			}
			if got != "true" || len(r.mergeCalls()) != 1 {
				t.Fatalf("eligible pull: eligible=%q merges=%v, want exactly one", got, r.mergeCalls())
			}
		})
	}
}

// AC-CCR-019 (REQ-CCR-012) — the merge decision path: head, checks, files, review, merge condition, merge step.
// Mutants (each applied to the committed workflow): `&& false` in the merge if; an exit 0 before gh pr merge;
// merge_deadline assigned inside the merge step; the review term dropped from the merge if; the files
// dependency dropped from the review step and the merge if together; --match-head-commit dropped, or pinned
// to a constant other than the verified head SHA (D38: the recorded form of the audit's "another variable").
func TestAutoMergeMergeDecisionPath(t *testing.T) {
	t.Parallel()
	type sub struct {
		name string
		run  func(t *testing.T)
	}
	// clean builds the fixture on which the whole chain merges: two benign files bound to the head, a clean
	// native state comment bound to the head, no corroborating finding, passing checks, the head stable.
	clean := func(t *testing.T, script string) *amRig {
		t.Helper()
		r := newAMRig(t, script)
		r.e.setPages("files", cgBenignFiles(2, cgHead))
		r.e.setPages("issues", cgNativeIssueComments(t, cgHead))
		r.setPulls(amPullWithIssueComments(cgHead, false, 2, 0, 3))
		return r
	}
	noMerge := func(t *testing.T, r *amRig) {
		t.Helper()
		r.runChain()
		if n := len(r.mergeCalls()); n != 0 {
			t.Fatalf("gh pr merge was called %d times, want 0: %v", n, r.mergeCalls())
		}
	}
	subs := []sub{
		{"clean_fixture_merges_once_with_head_pin", func(t *testing.T) {
			r := clean(t, "real")
			r.runChain()
			calls := r.mergeCalls()
			if len(calls) != 1 {
				t.Fatalf("gh pr merge called %d times, want exactly 1: %v\n%v", len(calls), calls, r.e.calls())
			}
			c := calls[0]
			if amFlagValue(c, "--match-head-commit") != cgHead || strings.Contains(c, "--delete-branch") || !strings.Contains(c, "--squash") || strings.Contains(c, "--merge") {
				t.Fatalf("merge call %q: want --squash, no branch deletion, --match-head-commit %s", c, cgHead)
			}
		}},
		{"release_branch_clean_fixture_uses_merge_flag", func(t *testing.T) {
			r := clean(t, "real")
			r.seed("pr", "branch", "release/v3.2.0")
			r.runChain()
			calls := r.mergeCalls()
			if len(calls) != 1 || !strings.Contains(calls[0], "--merge") || strings.Contains(calls[0], "--squash") || amFlagValue(calls[0], "--match-head-commit") != cgHead {
				t.Fatalf("merge calls %v, want one --merge pinned to %s", calls, cgHead)
			}
		}},
		{"review_exit_1_zero_merge_calls", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_REVIEW_EXITS=1")
			noMerge(t, r)
		}},
		{"review_draft_exit_11_zero_merge_calls", func(t *testing.T) {
			r := clean(t, "real")
			r.setPulls(amPull(cgHead, true, 2, 0))
			noMerge(t, r)
			if got := r.ctx.outputs["review"]["state"]; got != "not-run:draft" {
				t.Fatalf("review state=%q, want not-run:draft", got)
			}
		}},
		{"review_absent_expiry_zero_merge_calls", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_REVIEW_EXITS=" + amRepeat("10", 59) + " 11")
			noMerge(t, r)
			if got := r.ctx.outputs["review"]["state"]; got != "fail" {
				t.Fatalf("review state=%q, want fail (a non-draft NOT RUN is a failure, T-CHAIN row 4)", got)
			}
		}},
		{"review_unknown_exit_zero_merge_calls", func(t *testing.T) {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_REVIEW_EXITS=7")
			noMerge(t, r)
		}},
		{"old_review_does_not_block_current_clean_receipt", func(t *testing.T) {
			r := clean(t, "real")
			r.e.setPages("reviews", []any{cgReview(10, cgOther, "APPROVED")})
			r.runChain()
			if n := len(r.mergeCalls()); n != 1 {
				t.Fatalf("gh pr merge was called %d times, want one current-head clean receipt: %v", n, r.mergeCalls())
			}
		}},
		{"head_moved_at_merge_step_zero_merge_calls", func(t *testing.T) {
			r := clean(t, "real")
			r.setViews(cgHead, cgOther)
			noMerge(t, r)
		}},
		{"deadline_expired_zero_merge_calls", func(t *testing.T) {
			r := clean(t, "real")
			r.deadline = "1"
			noMerge(t, r)
		}},
		{"gate_path_change_zero_merge_calls", func(t *testing.T) {
			// A gate path is changed beside a clean native review bound to the head.
			r := clean(t, "real")
			files := cgBenignFiles(1, cgHead)
			files = append(files, cgFile(".github/workflows/zz-injected.yml", "added", cgHead))
			r.e.setPages("files", files)
			r.setPulls(amPull(cgHead, false, 2, 0))
			noMerge(t, r)
		}},
		{"page_failure_zero_merge_calls", func(t *testing.T) {
			r := clean(t, "real")
			r.setEnv("FAIL_PAGE=files:1")
			noMerge(t, r)
		}},
		{"gate_script_missing_zero_merge_calls", func(t *testing.T) {
			// T-CHAIN row 11: the missing script fails the files step and the job; nothing after it runs.
			r := clean(t, "none")
			ch := r.runChain()
			if n := len(r.mergeCalls()); n != 0 || ch.job != "failure" {
				t.Fatalf("merge calls=%d job=%q, want 0 and failure: %+v", n, ch.job, ch.ran)
			}
		}},
		// Plan revision 6: every refusal fixture ends the job as a failure (the files step, the review step, or the
		// missing script exits non-zero and the later steps are skipped); the clean fixture is a success with one
		// merge; the two guards the merge step owns (the head differs at its own gh pr view, the deadline has passed)
		// keep a success job with zero merge calls (Out of Scope of the amendment, decision-index Q22).
		{"refusal_fixtures_job_result_failure_clean_fixture_success", func(t *testing.T) {
			type fx struct {
				name string
				job  string
				mk   func() *amRig
			}
			fixtures := []fx{
				{"clean", "success", func() *amRig { return clean(t, "real") }},
				{"review_exit_1", "failure", func() *amRig {
					r := newAMRig(t, "stub")
					r.setEnv("STUB_REVIEW_EXITS=1")
					return r
				}},
				{"draft", "failure", func() *amRig {
					r := clean(t, "real")
					r.setPulls(amChainPulls(amPDraft)...)
					return r
				}},
				{"review_absent_expiry", "failure", func() *amRig {
					r := newAMRig(t, "stub")
					r.setEnv("STUB_REVIEW_EXITS=" + amRepeat("10", 59) + " 11")
					return r
				}},
				{"unknown_exit", "failure", func() *amRig {
					r := newAMRig(t, "stub")
					r.setEnv("STUB_REVIEW_EXITS=7")
					return r
				}},
				{"stale_review", "success", func() *amRig {
					r := clean(t, "real")
					r.e.setPages("reviews", []any{cgReview(10, cgOther, "APPROVED")})
					return r
				}},
				{"gate_path_change", "failure", func() *amRig {
					r := clean(t, "real")
					r.e.setPages("files", append(cgBenignFiles(1, cgHead), cgFile(".github/workflows/zz-injected.yml", "added", cgHead)))
					return r
				}},
				{"page_failure", "failure", func() *amRig {
					r := clean(t, "real")
					r.setEnv("FAIL_PAGE=files:1")
					return r
				}},
				{"gate_script_missing", "failure", func() *amRig { return clean(t, "none") }},
				{"head_moved_at_merge_step", "success", func() *amRig {
					r := clean(t, "real")
					r.setViews(cgHead, cgOther)
					return r
				}},
				{"deadline_expired", "success", func() *amRig {
					r := clean(t, "real")
					r.deadline = "1"
					return r
				}},
			}
			for _, f := range fixtures {
				r := f.mk()
				ch := r.runChain()
				wantMerges := 0
				if f.name == "clean" || f.name == "stale_review" {
					wantMerges = 1
				}
				if ch.job != f.job || len(r.mergeCalls()) != wantMerges {
					t.Fatalf("%s: job=%q merges=%d, want %q and %d: %+v", f.name, ch.job, len(r.mergeCalls()), f.job, wantMerges, ch.ran)
				}
			}
		}},
		{"if_expression_outside_grammar_fails_test", func(t *testing.T) {
			for _, bad := range []string{
				"steps.a.outputs.ok == 'true' && false",
				"steps.a.outputs.ok == 'true' || steps.b.outputs.ok == 'true'",
				"!(steps.a.outputs.ok == 'true')",
				"steps.a.outputs.ok != 'true'",
				"always()",
				"always() && false",
				"failure() && steps.a.outputs.ok == 'true'",
				"success() && steps.a.outputs.ok == 'true'",
				"always() || steps.a.outputs.ok == 'true'",
			} {
				if _, _, err := amParseCond(bad); err == nil {
					t.Errorf("the evaluator accepted %q", bad)
				}
			}
			// The one admitted extension, for mutant MF alone: an optional leading always().
			if terms, always, err := amParseCond("always() && steps.a.outputs.ok == 'true' && steps.b.outputs.ok == 'true'"); err != nil || !always || len(terms) != 2 {
				t.Errorf("the evaluator did not read a leading always(): terms=%v always=%t err=%v", terms, always, err)
			}
			wf, _ := amLoad(t)
			conds := 0
			for _, s := range wf.steps() {
				if s.ID == "checks" || s.ID == "files" || s.ID == "review" || s.Name == amMergeName {
					conds++
					if _, always, err := amParseCond(s.If); err != nil || always {
						t.Errorf("step %q: always=%t err=%v (the committed workflow carries no always())", s.Name, always, err)
					}
				}
			}
			if conds < 4 {
				t.Errorf("swept %d guard conditions, want at least 4", conds)
			}
		}},
	}
	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			s.run(t)
		})
	}
}

// amSetGen writes the pages the stub gh serves for the gen-th collection of a list (the stub counts
// page-1 reads per list): the first collection is generation 1, the pre-merge recheck's is generation 2.
func amSetGen(t *testing.T, r *amRig, kind string, gen int, entries ...any) {
	t.Helper()
	raw, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	r.e.write(filepath.Join(r.e.fixtures, kind, fmt.Sprintf("g%d", gen), "page-1.json"), string(raw), 0o600)
}

// Sync-phase correction, ruling B (codex finding F1 of the sync audit; no new REQ or AC identifier). The review
// step used to compare only head.sha on its second pull request read and wrote ok=true on the first clean pass, so
// a review comment added, or an approval dismissed, after the list reads, with the head unchanged, still reached a
// merge call. Each row takes the REAL failure path: the real gate script, the real step, the whole chain, a stub gh
// that serves a different second read or a different second collection; the step must end not-ok and gh pr merge
// must run zero times. The positive control merges once.
//
// Mutants (each applied to the committed workflow, recorded in run-evidence): the step without any recheck (no
// second-read count compare and no pre-merge block) merges in every row; the step without the pre-merge block only
// merges in every row but the second-read count row; the step without the second-read count compare only is caught
// by the pre-merge block (defence in depth, the count row stays red-free).
func TestAutoMergeReviewRecheckBeforeMerge(t *testing.T) {
	t.Parallel()
	H := cgHead
	p0 := amPullWithIssueComments(H, false, 2, 0, 3)
	p1 := amPullWithIssueComments(H, false, 2, 1, 3)
	p2 := amPullWithIssueComments(H, false, 2, 0, 4)
	p2MissingIssueCount := strings.Replace(p2, `,"comments":4`, ``, 1)
	p2MalformedIssueCount := strings.Replace(p2, `"comments":4`, `"comments":"four"`, 1)
	// The author step and files step read the pull object three times before the review pass. The review
	// pass reads twice, followed by the pre-merge recheck's two reads.
	chainPulls := func(review ...string) []string { return append([]string{p0, p0, p0}, review...) }
	type sub struct {
		name   string
		setup  func(t *testing.T, r *amRig)
		reason string // empty: the control, which must merge once
	}
	subs := []sub{
		{"unchanged_lists_merge_once_after_recheck", func(t *testing.T, r *amRig) {}, ""},
		{"second_read_review_comments_count_differs_not_ok", func(t *testing.T, r *amRig) {
			r.setPulls(chainPulls(p0, p1)...)
		}, "review-comments-moved-during-read"},
		{"recheck_review_comments_count_differs_not_ok", func(t *testing.T, r *amRig) {
			r.setPulls(chainPulls(p0, p0, p1)...)
		}, "review-comments-moved-before-merge"},
		{"state_comment_updated_on_recheck_not_ok", func(t *testing.T, r *amRig) {
			issues := cgNativeIssueComments(t, H)
			issues[0].(map[string]any)["updated_at"] = "2026-10-19T23:00:00Z"
			amSetGen(t, r, "issues", 2, issues...)
		}, "review-changed-before-merge"},
		{"unrelated_issue_comment_added_before_recheck_does_not_block", func(t *testing.T, r *amRig) {
			issues := cgNativeIssueComments(t, H)
			issues = append(issues, cgEntry{
				"id":         9999999999,
				"user":       map[string]any{"id": 90000001, "login": "external-user", "type": "User"},
				"body":       "Unrelated question about this pull request.",
				"created_at": "2026-10-19T23:00:00Z",
				"updated_at": "2026-10-19T23:00:00Z",
			})
			amSetGen(t, r, "issues", 2, issues...)
			r.setPulls(chainPulls(p0, p0, p2)...)
		}, ""},
		{"missing_issue_comment_count_on_recheck_not_ok", func(t *testing.T, r *amRig) {
			r.setPulls(chainPulls(p0, p0, p2MissingIssueCount)...)
		}, "issue-comments-unreadable-before-merge"},
		{"malformed_issue_comment_count_on_recheck_not_ok", func(t *testing.T, r *amRig) {
			r.setPulls(chainPulls(p0, p0, p2MalformedIssueCount)...)
		}, "issue-comments-unreadable-before-merge"},
		{"issue_comment_count_moved_during_recheck_not_ok", func(t *testing.T, r *amRig) {
			issues := cgNativeIssueComments(t, H)
			issues = append(issues, cgEntry{
				"id":         9999999999,
				"user":       map[string]any{"id": 90000001, "login": "external-user", "type": "User"},
				"body":       "Unrelated question about this pull request.",
				"created_at": "2026-10-19T23:00:00Z",
				"updated_at": "2026-10-19T23:00:00Z",
			})
			amSetGen(t, r, "issues", 2, issues...)
			r.setPulls(chainPulls(p0, p0, p2, p0)...)
		}, "issue-comments-moved-before-merge"},
		{"later_changes_requested_on_recheck_not_ok", func(t *testing.T, r *amRig) {
			amSetGen(t, r, "reviews", 2, cgReview(10, H, "APPROVED"), cgReview(11, H, "CHANGES_REQUESTED"))
		}, "review-changed-before-merge:changes-requested"},
		{"new_finding_comment_on_recheck_not_ok", func(t *testing.T, r *amRig) {
			// One comment before and one after, so the count proof still agrees: a human comment is replaced
			// by a finding of the identity bound to the head.
			r.setPulls(chainPulls(p1, p1, p1, p1)...)
			r.e.setPages("comments", []any{cgComment(1, H, H, cgHuman)})
			amSetGen(t, r, "comments", 2, cgComment(2, H, H))
		}, "review-changed-before-merge:finding:2"},
		{"second_state_comment_on_recheck_not_ok", func(t *testing.T, r *amRig) {
			issues := cgNativeIssueComments(t, H)
			duplicate := issues[0].(map[string]any)
			copy := make(map[string]any, len(duplicate))
			for k, v := range duplicate {
				copy[k] = v
			}
			copy["id"] = 9999999999
			issues[1] = copy
			amSetGen(t, r, "issues", 2, issues...)
		}, "review-changed-before-merge:state-comment-duplicated"},
	}
	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			r := newAMRig(t, "real")
			r.e.setPages("files", cgBenignFiles(2, H))
			r.e.setPages("reviews", []any{})
			r.e.setPages("comments", []any{})
			r.e.setPages("issues", cgNativeIssueComments(t, H))
			r.setPulls(chainPulls(p0, p0, p0, p0)...)
			s.setup(t, r)
			r.runChain()
			merges := r.mergeCalls()
			out := r.ctx.outputs["review"]
			res := r.results["review"]
			all := res.stdout + res.stderr
			if s.reason == "" {
				if len(merges) != 1 || out["ok"] != "true" || out["state"] != "pass" {
					t.Fatalf("control: merges=%v ok=%q state=%q, want one merge and true/pass\n%s", merges, out["ok"], out["state"], all)
				}
				return
			}
			if out["ok"] == "true" || out["state"] != "fail" {
				t.Fatalf("ok=%q state=%q, want not-true/fail\n%s", out["ok"], out["state"], all)
			}
			if len(merges) != 0 {
				t.Fatalf("gh pr merge was called %d times after a changed review state: %v", len(merges), merges)
			}
			if !strings.Contains(all, s.reason) {
				t.Fatalf("the step does not name %q:\n%s", s.reason, all)
			}
		})
	}
}

func TestAutoMergeRetriesIssueCommentCountRaceDuringReviewWait(t *testing.T) {
	t.Parallel()
	cgTools(t)
	H := cgHead
	p2 := amPullWithIssueComments(H, false, 2, 0, 3)
	p3 := amPullWithIssueComments(H, false, 2, 0, 4)
	r := newAMRig(t, "real")
	r.e.setPages("files", cgBenignFiles(2, H))
	r.e.setPages("reviews", []any{})
	r.e.setPages("comments", []any{})
	initial := cgNativeIssueComments(t, H)
	withReceipt := append(cgNativeIssueComments(t, H), cgEntry{
		"id":         6269999999,
		"user":       map[string]any{"id": cgBotID, "login": cgBotLogin, "type": "Bot"},
		"body":       "markerless review receipt",
		"created_at": "2026-10-19T23:00:00Z",
		"updated_at": "2026-10-19T23:00:00Z",
	})
	amSetGen(t, r, "issues", 1, initial...)
	amSetGen(t, r, "issues", 2, withReceipt...)
	amSetGen(t, r, "issues", 3, withReceipt...)
	// author + two files reads, then the first review pass observes 2 -> 3 comments;
	// the second review pass and final recheck observe the stable 3-comment list.
	r.setPulls([]string{p2, p2, p2, p2, p3, p3, p3, p3, p3}...)
	result := r.runChain()
	if result.job != "success" || len(r.mergeCalls()) != 1 || r.sleeps() != 1 {
		review := r.results["review"]
		t.Fatalf("comment-count race should retry once then merge: job=%s merges=%v sleeps=%d review=%q/%q err=%s calls=%v", result.job, r.mergeCalls(), r.sleeps(), r.ctx.outputs["review"]["ok"], r.ctx.outputs["review"]["state"], review.stdout+review.stderr, r.e.calls())
	}
	if !strings.Contains(r.results["review"].stdout, "native-review-clean") {
		t.Fatalf("review step did not record the clean native result: %s%s", r.results["review"].stdout, r.results["review"].stderr)
	}
	issueReads := 0
	for _, call := range r.e.calls() {
		if strings.HasPrefix(call, "gh api repos/modu-ai/moai-adk/issues/7/comments?") {
			issueReads++
		}
	}
	if issueReads != 3 {
		t.Fatalf("issue-comment list collections=%d, want retry pass plus clean pass and recheck: %v", issueReads, r.e.calls())
	}
}

// ---------------------------------------------------------------------------------------------
// Plan revision 6, M5 (REQ-CCR-018, REQ-CCR-019; AC-CCR-020 to AC-CCR-025): a failure in the files step or the
// review step ends that step AND the job as not-success; a draft is the single state not-run:draft with exit 11;
// each output key carries one value. Every row below runs the real step body (the real gate script or the one
// subcommand-aware stub, the stub gh, the no-op sleep) through the shared chain harness and asserts THREE
// observables together, never the output file alone: the step's exit status, the job result the harness derives
// from the executed steps (amDeriveJob), and the step's output-file content (each key's values, in order).
// The derived job result is a rule over the workflow's structure; what GitHub reports on a runner is UNOBSERVED
// (spec section 4, research U11). The single table every row instantiates is design.md section 5, T-CHAIN.
// ---------------------------------------------------------------------------------------------

var (
	amP0     = amPullWithIssueComments(cgHead, false, 2, 0, 3)
	amP1     = amPullWithIssueComments(cgHead, false, 2, 1, 3)
	amPOther = amPull(cgOther, false, 2, 0)
	amPDraft = amPull(cgHead, true, 2, 0)
)

// amChainPulls is the pull-object sequence of a whole chain: the files step reads it twice before the review step.
func amChainPulls(review ...string) []string { return append([]string{amP0, amP0, amP0}, review...) }

// amCleanRig is the fixture on which the whole chain merges.
func amCleanRig(t *testing.T, script string) *amRig {
	t.Helper()
	r := newAMRig(t, script)
	r.e.setPages("files", cgBenignFiles(2, cgHead))
	r.e.setPages("reviews", []any{cgReview(10, cgHead, "APPROVED")})
	r.e.setPages("issues", cgNativeIssueComments(t, cgHead))
	r.setPulls(amP0)
	return r
}

// amObs is what the harness observes for one step: exit status, derived job result, and output values by key
// (the values of a key joined with "|", so a duplicated key is a different observation from a single one).
type amObs struct {
	exit    int
	job     string
	outputs map[string]string
}

func (r *amRig) obs(step, job string) amObs {
	out := map[string]string{}
	for _, l := range r.raw[step] {
		if k, v, ok := strings.Cut(l, "="); ok {
			if prev, has := out[k]; has {
				out[k] = prev + "|" + v
			} else {
				out[k] = v
			}
		}
	}
	return amObs{exit: r.results[step].code, job: job, outputs: out}
}

// amCompare compares all three observables; any difference is an error.
func amCompare(got, want amObs) error {
	if got.exit != want.exit {
		return fmt.Errorf("step exit status %d, want %d", got.exit, want.exit)
	}
	if got.job != want.job {
		return fmt.Errorf("job result %q, want %q", got.job, want.job)
	}
	if len(got.outputs) != len(want.outputs) {
		return fmt.Errorf("outputs %v, want %v", got.outputs, want.outputs)
	}
	for k, v := range want.outputs {
		if got.outputs[k] != v {
			return fmt.Errorf("output %s=%q, want %q (all outputs %v)", k, got.outputs[k], v, got.outputs)
		}
	}
	return nil
}

func (r *amRig) keyValues(step, key string) []string {
	var out []string
	for _, l := range r.raw[step] {
		if k, v, ok := strings.Cut(l, "="); ok && k == key {
			out = append(out, v)
		}
	}
	return out
}

func amText(res cgResult) string { return res.stdout + res.stderr }

// amWantFailure asserts a T-CHAIN failure row for one step of a chain run: the step ran and exited 1, the derived
// job result is failure, ok=false and state=fail once each, the annotation is an ::error:: naming the reason, the
// steps after it did not run, and gh pr merge was never called.
func amWantFailure(t *testing.T, r *amRig, ch amChainResult, step, annotationPrefix, reason string) {
	t.Helper()
	if _, ran := ch.executed(step); !ran {
		t.Fatalf("the %s step did not run: %+v", step, ch.ran)
	}
	want := amObs{exit: 1, job: "failure", outputs: map[string]string{"ok": "false", "state": "fail"}}
	if err := amCompare(r.obs(step, ch.job), want); err != nil {
		t.Fatalf("%s step: %v\n%s", step, err, amText(r.results[step]))
	}
	text := amText(r.results[step])
	if !strings.Contains(text, "::error::"+annotationPrefix) || !strings.Contains(text, reason) {
		t.Fatalf("%s step annotation is not an ::error::%s naming %q:\n%s", step, annotationPrefix, reason, text)
	}
	if _, ran := ch.executed("merge"); ran {
		t.Fatalf("the merge step ran after the %s step failed: %+v", step, ch.ran)
	}
	if n := len(r.mergeCalls()); n != 0 {
		t.Fatalf("gh pr merge was called %d times: %v", n, r.mergeCalls())
	}
}

const (
	amReviewNote = "native review gate: "
	amFilesNote  = "changed-file check: "
)

// AC-CCR-020 (REQ-CCR-018) — the review step: every failure of T-CHAIN rows 2, 3, 6, 7, 8, 9 ends the step with exit 1
// and the job as a failure, with ok=false, state=fail, an ::error:: annotation, and no merge; the clean pass exits 0.
// Mutants (each applied to the committed workflow): a withhold that ends exit 0 (the pre-revision-6 step) turns every
// failure row red on the exit status although its outputs read ok=false state=fail; continue-on-error: true on the
// step turns the job-result assertion of every failure row and review_step_has_no_continue_on_error red; a step
// that exits 1 on the clean pass turns clean_pass_step_exit_0_job_success red.
func TestAutoMergeReviewFailureFailsStepAndJob(t *testing.T) {
	t.Parallel()
	type fail struct {
		name   string
		reason string
		mk     func(t *testing.T) *amRig
	}
	H := cgHead
	fails := []fail{
		{"changes_requested_step_exit_nonzero_job_failure", "changes-requested", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.e.setPages("reviews", []any{cgReview(10, H, "CHANGES_REQUESTED")})
			return r
		}},
		{"finding_step_exit_nonzero_job_failure", "finding:2", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setPulls(amChainPulls(amP1)...)
			r.e.setPages("comments", []any{cgComment(2, H, H)})
			return r
		}},
		{"stale_review_without_current_native_summary_fails_closed", "stale-review", func(t *testing.T) *amRig {
			r := newAMRig(t, "real")
			r.e.setPages("files", cgBenignFiles(2, cgHead))
			r.e.setPages("reviews", []any{cgReview(10, cgOther, "APPROVED")})
			r.e.setPages("issues", []any{})
			r.setPulls(amPullWithIssueComments(cgHead, false, 2, 0, 0))
			return r
		}},
		{"commented_not_conclusive_step_exit_nonzero", "review-not-conclusive", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.e.setPages("reviews", []any{cgReview(10, H, "COMMENTED")})
			return r
		}},
		{"unknown_gate_exit_step_exit_nonzero", "gate-exit-7", func(t *testing.T) *amRig {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_REVIEW_EXITS=7")
			return r
		}},
		{"head_moved_during_read_step_exit_nonzero", "head-moved-during-read", func(t *testing.T) *amRig {
			r := newAMRig(t, "stub")
			r.setPulls(amChainPulls(amP0, amPOther)...)
			return r
		}},
		{"review_comments_moved_during_read_step_exit_nonzero", "review-comments-moved-during-read", func(t *testing.T) *amRig {
			r := newAMRig(t, "stub")
			r.setPulls(amChainPulls(amP0, amP1)...)
			return r
		}},
		{"recheck_mismatch_step_exit_nonzero", "review-changed-before-merge", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			amSetGen(t, r, "reviews", 2, cgReview(10, H, "CHANGES_REQUESTED"))
			return r
		}},
		{"unreadable_to_end_of_window_step_exit_nonzero", "review-unreadable", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setEnv("FAIL_KIND=reviews")
			return r
		}},
	}
	for _, f := range fails {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			r := f.mk(t)
			ch := r.runChain()
			amWantFailure(t, r, ch, "review", amReviewNote, f.reason)
		})
	}
	t.Run("gate_script_missing_step_exit_nonzero", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := newAMRig(t, "none")
		r.seedGuards()
		res, _ := r.runStep(r.step("review"))
		job := amDeriveJob([]amExec{{key: "review", code: res.code}})
		want := amObs{exit: 1, job: "failure", outputs: map[string]string{"ok": "false", "state": "fail"}}
		if err := amCompare(r.obs("review", job), want); err != nil {
			t.Fatalf("%v\n%s", err, amText(res))
		}
		if !strings.Contains(amText(res), "::error::"+amReviewNote+"gate-script-missing") {
			t.Fatalf("no ::error:: naming gate-script-missing:\n%s", amText(res))
		}
	})
	t.Run("merge_step_skipped_after_review_failure", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "real")
		r.e.setPages("reviews", []any{cgReview(10, H, "CHANGES_REQUESTED")})
		ch := r.runChain()
		if _, ran := ch.executed("merge"); ran || len(r.mergeCalls()) != 0 || ch.job != "failure" {
			t.Fatalf("merge ran=%t merge calls=%v job=%q, want skipped, none, failure: %+v", ran, r.mergeCalls(), ch.job, ch.ran)
		}
		if e, _ := ch.executed("review"); e.code != 1 {
			t.Fatalf("review exit %d, want 1", e.code)
		}
	})
	t.Run("clean_pass_step_exit_0_job_success", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "real")
		ch := r.runChain()
		want := amObs{exit: 0, job: "success", outputs: map[string]string{"ok": "true", "state": "pass"}}
		if err := amCompare(r.obs("review", ch.job), want); err != nil {
			t.Fatalf("%v\n%s", err, amText(r.results["review"]))
		}
	})
	t.Run("review_step_has_no_continue_on_error", func(t *testing.T) {
		t.Parallel()
		wf, _ := amLoad(t)
		s, ok := wf.byID("review")
		if !ok || s.ContinueOnError {
			t.Fatalf("review step present=%t continue-on-error=%t, want present and not continue-on-error", ok, s.ContinueOnError)
		}
	})
}

// AC-CCR-021 (REQ-CCR-008) — a non-draft pull request with no review at the end of the window is a failure.
// Mutants: a step that ends review-absent with exit 0 and state=not-run (the pre-revision-6 behaviour) turns the
// first two rows red; one that reads review-absent as success turns review_absent_zero_merge_calls red.
func TestAutoMergeReviewAbsentIsFailure(t *testing.T) {
	t.Parallel()
	absent := func(t *testing.T) (*amRig, amChainResult) {
		t.Helper()
		cgTools(t)
		r := newAMRig(t, "stub")
		r.setEnv("STUB_REVIEW_EXITS=" + amRepeat("10", 59) + " 11")
		return r, r.runChain()
	}
	t.Run("review_absent_window_closed_step_exit_1_job_failure", func(t *testing.T) {
		t.Parallel()
		r, ch := absent(t)
		want := amObs{exit: 1, job: "failure", outputs: map[string]string{"ok": "false", "state": "fail"}}
		if err := amCompare(r.obs("review", ch.job), want); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("review_absent_state_fail_not_not_run", func(t *testing.T) {
		t.Parallel()
		r, _ := absent(t)
		if v := r.keyValues("review", "state"); len(v) != 1 || v[0] != "fail" {
			t.Fatalf("state values %v, want exactly [fail] (never not-run:draft, never a second key)", v)
		}
		if strings.Contains(strings.Join(r.raw["review"], "\n"), "not-run") {
			t.Fatalf("the output file mentions not-run: %v", r.raw["review"])
		}
	})
	t.Run("review_absent_annotation_error_names_cause", func(t *testing.T) {
		t.Parallel()
		r, _ := absent(t)
		if !strings.Contains(amText(r.results["review"]), "::error::"+amReviewNote+"not-run:review-absent") {
			t.Fatalf("no ::error:: naming not-run:review-absent:\n%s", amText(r.results["review"]))
		}
	})
	t.Run("review_absent_zero_merge_calls", func(t *testing.T) {
		t.Parallel()
		r, ch := absent(t)
		if _, ran := ch.executed("merge"); ran || len(r.mergeCalls()) != 0 {
			t.Fatalf("merge ran=%t calls=%v after a review absent at the end of the window", ran, r.mergeCalls())
		}
	})
	t.Run("review_absent_window_open_keeps_waiting_no_output", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := newAMRig(t, "stub")
		r.seedGuards()
		r.setEnv("STUB_REVIEW_EXITS=10 10 0")
		res, _ := r.runStep(r.step("review"))
		if res.code != 0 || !strings.Contains(amText(res), "not observed yet") {
			t.Fatalf("exit=%d, want 0 after two waiting passes and a notice that no review was observed yet:\n%s", res.code, amText(res))
		}
		if len(r.raw["review"]) != 2 {
			t.Fatalf("the output file holds %v: waiting passes must record nothing, only the terminating path records", r.raw["review"])
		}
	})
	t.Run("lists_unreadable_at_expiry_step_exit_1_review_unreadable", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "real")
		r.seedGuards()
		r.setEnv("FAIL_KIND=reviews")
		res, _ := r.runStep(r.step("review"))
		job := amDeriveJob([]amExec{{key: "review", code: res.code}})
		want := amObs{exit: 1, job: "failure", outputs: map[string]string{"ok": "false", "state": "fail"}}
		if err := amCompare(r.obs("review", job), want); err != nil {
			t.Fatalf("%v\n%s", err, amText(res))
		}
		if !strings.Contains(amText(res), "review-unreadable") || strings.Contains(amText(res), "not-run") {
			t.Fatalf("want a failure named review-unreadable, never not-run:\n%s", amText(res))
		}
	})
}

// AC-CCR-022 (REQ-CCR-008) — a draft is the single state not-run:draft, step exit 11, a job that is never success.
// Mutants: the pre-revision-6 step (draft ends exit 0, state=fail then state=not-run) turns the first, second, and
// last rows red; mapping a draft to state=pass or ok=true turns draft_ok_output_false red; continue-on-error turns the
// job-result assertion red.
func TestAutoMergeDraftIsNotRunNeverSuccess(t *testing.T) {
	t.Parallel()
	H := cgHead
	draft := func(t *testing.T, reviews ...any) (*amRig, amChainResult) {
		t.Helper()
		cgTools(t)
		r := amCleanRig(t, "real")
		switch {
		case reviews == nil:
		case len(reviews) == 0:
			r.e.setRawPage("reviews", 1, "[]")
		default:
			r.e.setPages("reviews", reviews)
		}
		r.setPulls(amChainPulls(amPDraft)...)
		return r, r.runChain()
	}
	t.Run("draft_step_exit_11_job_not_success", func(t *testing.T) {
		t.Parallel()
		for name, reviews := range map[string][]any{
			"no_review":         {},
			"approved":          {cgReview(10, H, "APPROVED")},
			"changes_requested": {cgReview(10, H, "CHANGES_REQUESTED")},
		} {
			r, ch := draft(t, reviews...)
			want := amObs{exit: 11, job: "failure", outputs: map[string]string{"ok": "false", "state": "not-run:draft"}}
			if err := amCompare(r.obs("review", ch.job), want); err != nil {
				t.Fatalf("%s: %v\n%s", name, err, amText(r.results["review"]))
			}
			if ch.job == "success" {
				t.Fatalf("%s: a draft ended the job as success", name)
			}
		}
	})
	t.Run("draft_state_output_is_single_value_not_run_draft", func(t *testing.T) {
		t.Parallel()
		r, _ := draft(t)
		if v := r.keyValues("review", "state"); len(v) != 1 || v[0] != "not-run:draft" {
			t.Fatalf("state values %v, want exactly [not-run:draft]", v)
		}
	})
	t.Run("draft_ok_output_false", func(t *testing.T) {
		t.Parallel()
		r, _ := draft(t)
		if v := r.keyValues("review", "ok"); len(v) != 1 || v[0] != "false" {
			t.Fatalf("ok values %v, want exactly [false]", v)
		}
	})
	t.Run("draft_skips_list_fetch", func(t *testing.T) {
		t.Parallel()
		r, _ := draft(t)
		for _, c := range r.e.calls() {
			if strings.Contains(c, "/reviews") || strings.Contains(c, "/comments") {
				t.Fatalf("a draft fetched a review list: %s", c)
			}
		}
	})
	t.Run("draft_with_clean_approved_review_still_not_success", func(t *testing.T) {
		t.Parallel()
		r, ch := draft(t, cgReview(10, H, "APPROVED"))
		if e, _ := ch.executed("review"); e.code != 11 || ch.job == "success" || r.ctx.outputs["review"]["state"] == "pass" {
			t.Fatalf("review exit=%d job=%q state=%q: a draft with a clean approval must not read as a pass", e.code, ch.job, r.ctx.outputs["review"]["state"])
		}
	})
	t.Run("draft_annotation_notice_names_cause_and_says_not_a_pass", func(t *testing.T) {
		t.Parallel()
		r, _ := draft(t)
		text := amText(r.results["review"])
		if !strings.Contains(text, "::notice::NOT RUN: draft") || !strings.Contains(text, "not a review result") ||
			!strings.Contains(text, "never a pass") || strings.Contains(text, "::error::") {
			t.Fatalf("the draft annotation must be a ::notice:: naming draft, not a review result, never a pass:\n%s", text)
		}
	})
	t.Run("draft_zero_merge_calls", func(t *testing.T) {
		t.Parallel()
		r, ch := draft(t, cgReview(10, H, "APPROVED"))
		if _, ran := ch.executed("merge"); ran || len(r.mergeCalls()) != 0 {
			t.Fatalf("merge ran=%t calls=%v on a draft", ran, r.mergeCalls())
		}
	})
	t.Run("draft_state_distinct_from_fail_and_pass", func(t *testing.T) {
		t.Parallel()
		r, _ := draft(t)
		if s := r.ctx.outputs["review"]["state"]; s == "fail" || s == "pass" || s != "not-run:draft" {
			t.Fatalf("state=%q, want not-run:draft, distinct from fail and pass", s)
		}
	})
}

// AC-CCR-023 (REQ-CCR-019) — each output key is recorded once, on the terminating path, with an allowed value pair.
// Mutant: pre-writing ok=false / state=fail and appending the terminal value (the pre-revision-6 behaviour) turns every
// row red, the clean-pass row included (ok and state each appear twice).
func TestAutoMergeGuardStepsOutputOneValuePerKey(t *testing.T) {
	t.Parallel()
	H := cgHead
	type run struct {
		step  string
		build func(t *testing.T) *amRig
	}
	runs := map[string]run{
		"review_clean": {"review", func(t *testing.T) *amRig { return amCleanRig(t, "real") }},
		"review_changes_requested": {"review", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.e.setPages("reviews", []any{cgReview(10, H, "CHANGES_REQUESTED")})
			return r
		}},
		"review_draft": {"review", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setPulls(amChainPulls(amPDraft)...)
			return r
		}},
		"review_absent": {"review", func(t *testing.T) *amRig {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_REVIEW_EXITS=" + amRepeat("10", 59) + " 11")
			return r
		}},
		"files_clear": {"files", func(t *testing.T) *amRig { return amCleanRig(t, "real") }},
		"files_refusal": {"files", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.e.setPages("files", append(cgBenignFiles(1, H), cgFile(".github/workflows/zz.yml", "added", H)))
			return r
		}},
	}
	once := func(t *testing.T, name string) *amRig {
		t.Helper()
		cgTools(t)
		x := runs[name]
		r := x.build(t)
		r.runChain()
		if len(r.keyValues(x.step, "ok")) != 1 || len(r.keyValues(x.step, "state")) != 1 || len(r.raw[x.step]) != 2 {
			t.Fatalf("%s: the %s step output is %v, want exactly one ok line and one state line", name, x.step, r.raw[x.step])
		}
		return r
	}
	for _, c := range []struct{ row, run string }{
		{"review_step_clean_pass_each_key_once", "review_clean"},
		{"review_step_changes_requested_each_key_once", "review_changes_requested"},
		{"review_step_draft_each_key_once", "review_draft"},
		{"review_step_review_absent_each_key_once", "review_absent"},
		{"files_step_clear_each_key_once", "files_clear"},
		{"files_step_refusal_each_key_once", "files_refusal"},
	} {
		t.Run(c.row, func(t *testing.T) {
			t.Parallel()
			once(t, c.run)
		})
	}
	t.Run("output_pairs_within_allowed_set", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		allowed := map[string]bool{"true/pass": true, "false/fail": true, "false/not-run:draft": true}
		for name, x := range runs {
			r := once(t, name)
			pair := r.keyValues(x.step, "ok")[0] + "/" + r.keyValues(x.step, "state")[0]
			if !allowed[pair] || (x.step == "files" && pair == "false/not-run:draft") {
				t.Fatalf("%s: the %s step recorded the pair %s, outside the allowed set of T-OUTPUTS", name, x.step, pair)
			}
		}
	})
}

// AC-CCR-024 (REQ-CCR-018) — the harness itself: its comparison sees exit status and job result as well as outputs,
// its job-result rule is failure on any non-zero step that is not continue-on-error, and neither guard step carries
// continue-on-error or a status function. Mutants: a comparison on outputs only turns the two harness_comparison_rejects
// rows red; a derivation that ignores the exit status turns harness_job_result_failure_when_a_step_exits_nonzero red.
func TestAutoMergeStepHarnessObservesExitAndJobResult(t *testing.T) {
	t.Parallel()
	outs := map[string]string{"ok": "false", "state": "fail"}
	t.Run("harness_records_exit_status_and_outputs_per_step", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "real")
		ch := r.runChain()
		if len(ch.ran) != 6 {
			t.Fatalf("the clean chain ran %d steps, want 6 (head, author, checks, files, review, merge): %+v", len(ch.ran), ch.ran)
		}
		for _, e := range ch.ran {
			res, ok := r.results[e.key]
			if !ok || res.code != e.code {
				t.Fatalf("step %s: recorded result present=%t exit=%d, chain exit=%d", e.key, ok, res.code, e.code)
			}
		}
		for _, id := range []string{"files", "review"} {
			if len(r.raw[id]) != 2 {
				t.Fatalf("step %s: recorded outputs %v, want 2 lines", id, r.raw[id])
			}
		}
	})
	t.Run("harness_job_result_failure_when_a_step_exits_nonzero", func(t *testing.T) {
		t.Parallel()
		if got := amDeriveJob([]amExec{{key: "a", code: 0}, {key: "b", code: 1}}); got != "failure" {
			t.Fatalf("job %q, want failure", got)
		}
		if got := amDeriveJob([]amExec{{key: "a", code: 11}}); got != "failure" {
			t.Fatalf("job %q for exit 11, want failure", got)
		}
		if got := amDeriveJob([]amExec{{key: "a", code: 1, coe: true}}); got != "success" {
			t.Fatalf("job %q for a continue-on-error step, want success", got)
		}
	})
	t.Run("harness_job_result_success_only_when_all_steps_zero_or_skipped", func(t *testing.T) {
		t.Parallel()
		if amDeriveJob(nil) != "success" || amDeriveJob([]amExec{{key: "a"}, {key: "b"}}) != "success" {
			t.Fatal("zero executed steps, or all steps exiting 0, must derive success")
		}
		if amDeriveJob([]amExec{{key: "a"}, {key: "b", code: 2}}) == "success" {
			t.Fatal("a step that exited 2 derived success")
		}
	})
	t.Run("harness_comparison_rejects_exit_mismatch_with_matching_outputs", func(t *testing.T) {
		t.Parallel()
		if amCompare(amObs{exit: 0, job: "failure", outputs: outs}, amObs{exit: 1, job: "failure", outputs: outs}) == nil {
			t.Fatal("a result with the same outputs and job but another exit status passed the comparison")
		}
		if err := amCompare(amObs{exit: 1, job: "failure", outputs: outs}, amObs{exit: 1, job: "failure", outputs: outs}); err != nil {
			t.Fatalf("an identical observation was rejected: %v", err)
		}
	})
	t.Run("harness_comparison_rejects_job_result_mismatch_with_matching_exit", func(t *testing.T) {
		t.Parallel()
		if amCompare(amObs{exit: 1, job: "success", outputs: outs}, amObs{exit: 1, job: "failure", outputs: outs}) == nil {
			t.Fatal("a result with the same outputs and exit status but another job result passed the comparison")
		}
	})
	t.Run("guard_steps_have_no_continue_on_error_or_status_override", func(t *testing.T) {
		t.Parallel()
		wf, _ := amLoad(t)
		for _, id := range []string{"files", "review"} {
			s, ok := wf.byID(id)
			if !ok {
				t.Fatalf("step %s absent", id)
			}
			if s.ContinueOnError {
				t.Fatalf("step %s is continue-on-error", id)
			}
			for _, fn := range []string{"always()", "failure()", "cancelled()", "success()"} {
				if strings.Contains(s.If, fn) {
					t.Fatalf("step %s if carries the status function %s: %q", id, fn, s.If)
				}
			}
			if _, always, err := amParseCond(s.If); err != nil || always {
				t.Fatalf("step %s if %q: always=%t err=%v", id, s.If, always, err)
			}
		}
	})
}

// AC-CCR-025 (REQ-CCR-018) — the files step: every failure of T-CHAIN rows 11 to 13 ends the step with exit 1 and the
// job as a failure, skips the review and merge steps, and calls gh pr merge zero times; the clean chain exits 0
// everywhere, merges once, and the job is a success.
// Mutants: the pre-revision-6 files step (withhold ends exit 0; a missing script ends exit 0 with ok=false state=fail)
// turns every failure row red on the exit status and the job result; continue-on-error: true on files turns the
// job-result assertions red; a chain runner that asserts the output file only is rejected by AC-CCR-024.
func TestAutoMergeFilesStepFailureFailsStepAndJob(t *testing.T) {
	t.Parallel()
	type fail struct {
		name   string
		reason string
		mk     func(t *testing.T) *amRig
	}
	fails := []fail{
		{"gate_script_missing_files_step_exit_nonzero_job_failure", "gate-script-missing", func(t *testing.T) *amRig {
			return amCleanRig(t, "none")
		}},
		{"gate_path_changed_step_exit_nonzero_job_failure", "gate-path-changed:.github/workflows/zz.yml", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.e.setPages("files", append(cgBenignFiles(1, cgHead), cgFile(".github/workflows/zz.yml", "added", cgHead)))
			return r
		}},
		{"fetch_pages_error_step_exit_nonzero", "page-error", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setEnv("FAIL_PAGE=files:1")
			return r
		}},
		{"pr_unreadable_step_exit_nonzero", "files-unreadable", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setPulls(amP0, "FAIL")
			return r
		}},
		{"head_moved_before_read_step_exit_nonzero", "head-moved:", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setPulls(amPOther)
			return r
		}},
		{"head_moved_during_read_step_exit_nonzero", "head-moved-during-read", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setPulls(amP0, amP0, amPOther)
			return r
		}},
		{"count_moved_during_read_step_exit_nonzero", "count-moved-during-read", func(t *testing.T) *amRig {
			r := amCleanRig(t, "real")
			r.setPulls(amPull(cgHead, false, 2, 0), amPull(cgHead, false, 2, 0), amPull(cgHead, false, 3, 0))
			return r
		}},
		{"unknown_gate_exit_step_exit_nonzero", "gate-exit-7", func(t *testing.T) *amRig {
			r := newAMRig(t, "stub")
			r.setEnv("STUB_FILES_EXITS=7")
			return r
		}},
	}
	for _, f := range fails {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			r := f.mk(t)
			ch := r.runChain()
			amWantFailure(t, r, ch, "files", amFilesNote, f.reason)
			if _, ran := ch.executed("review"); ran {
				t.Fatalf("the review step ran after the files step failed: %+v", ch.ran)
			}
		})
	}
	t.Run("files_failure_skips_review_and_merge_steps", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "real")
		r.e.setPages("files", append(cgBenignFiles(1, cgHead), cgFile(".github/workflows/zz.yml", "added", cgHead)))
		ch := r.runChain()
		for _, key := range []string{"review", "merge"} {
			if _, ran := ch.executed(key); ran {
				t.Fatalf("the %s step ran after a files failure: %+v", key, ch.ran)
			}
		}
		for _, c := range r.e.calls() {
			if strings.Contains(c, "/reviews") || strings.Contains(c, "/comments") {
				t.Fatalf("a review list was read after a files failure: %s", c)
			}
		}
	})
	t.Run("chain_gate_script_missing_job_failure_zero_merges", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "none")
		ch := r.runChain()
		if ch.job != "failure" || len(r.mergeCalls()) != 0 {
			t.Fatalf("job=%q merge calls=%v, want failure and none: %+v", ch.job, r.mergeCalls(), ch.ran)
		}
		for _, key := range []string{"review", "merge"} {
			if _, ran := ch.executed(key); ran {
				t.Fatalf("the %s step ran with the gate script missing (a missing script must never read as a passing step)", key)
			}
		}
	})
	t.Run("files_step_has_no_continue_on_error", func(t *testing.T) {
		t.Parallel()
		wf, _ := amLoad(t)
		s, ok := wf.byID("files")
		if !ok || s.ContinueOnError {
			t.Fatalf("files step present=%t continue-on-error=%t, want present and not continue-on-error", ok, s.ContinueOnError)
		}
	})
	t.Run("clean_chain_step_exits_0_one_merge_job_success", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		r := amCleanRig(t, "real")
		ch := r.runChain()
		for _, e := range ch.ran {
			if e.code != 0 {
				t.Fatalf("step %s exited %d on the clean chain: %+v", e.key, e.code, ch.ran)
			}
		}
		if len(ch.ran) != 6 || ch.job != "success" || len(r.mergeCalls()) != 1 {
			t.Fatalf("steps=%d job=%q merges=%v, want 6, success, one merge: %+v", len(ch.ran), ch.job, r.mergeCalls(), ch.ran)
		}
	})
}
