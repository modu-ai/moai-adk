// release_drafter_cleanup_test.go: guards for SPEC-RELEASE-PIPELINE-001 M5
// (card t1536, REQ-RPC-010 .. REQ-RPC-013): the Release Drafter cleanup
// workflow must be able to run and must delete only what it should.
//
// The defects, reproduced RED before the repair (acceptance.md EV-3/7/8/10):
//
//   - the fetch asked `gh release list` for `targetCommitish`, a field that
//     command does not support (gh 2.92.0: "Unknown JSON field", exit 1), so the
//     step exited before the filter ever ran;
//   - the filter's `now | floor as $now |` rebinds the iteration input to a
//     number, so `.[]` iterated a number (jq exit 5);
//   - the delete loop deleted straight from the listing;
//   - nothing serialized the cleanup with Release Drafter, whose own group was
//     ref-keyed with cancel-in-progress: true;
//   - the dry-run path was unreachable by dispatch and, once wired, would have
//     compared a stringified boolean against the literal "1".
//
// Guards execute the workflow's own shell and jq against a mock `gh` that
// enforces the measured gh 2.92.0 field sets, so a requested-but-unsupported
// field fails here the way it fails on the runner. They need bash and jq; a
// missing tool fails (plan.md B5) except on windows, where the bash-driven
// guards are not exercised (the same precedent as release_fixture_test.go).
//
// Sentinels on failure: CLEANUP_FETCH_VIOLATION, CLEANUP_FILTER_VIOLATION,
// CLEANUP_REQUERY_VIOLATION, CLEANUP_CONCURRENCY_VIOLATION,
// CLEANUP_DRYRUN_VIOLATION.
package template_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	rdcCleanupRel = ".github/workflows/release-drafter-cleanup.yml"
	rdcDrafterRel = ".github/workflows/release-drafter.yml"
	rdcFixtureRel = "internal/template/testdata/release-drafter-cleanup.json"
)

// Field sets measured on gh 2.92.0 (acceptance.md EV-10 for list, and the
// `gh release view --json bogusField` listing for view, 2026-10-23).
var (
	rdcListFields = []string{"createdAt", "isDraft", "isImmutable", "isLatest", "isPrerelease", "name", "publishedAt", "tagName"}
	rdcViewFields = []string{"apiUrl", "assets", "author", "body", "createdAt", "databaseId", "id", "isDraft", "isImmutable", "isPrerelease", "name", "publishedAt", "tagName", "tarballUrl", "targetCommitish", "uploadUrl", "url", "zipballUrl"}
	// rdcFilterReads are the fields the filter reads, by name.
	rdcFilterReads = []string{"isDraft", "createdAt", "tagName", "targetCommitish"}
)

func rdcRequireTools(t *testing.T, needBash bool) {
	t.Helper()
	if needBash && runtime.GOOS == "windows" {
		t.Skip("the cleanup guards drive the workflow's bash through a mock gh; not exercised on windows (the verdict is platform-independent and the ubuntu legs run it)")
	}
	tools := []string{"jq"}
	if needBash {
		tools = append(tools, "bash")
	}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("%s is required by the cleanup guard and is not on PATH (plan.md B5: absence fails, it does not skip)", tool)
		}
	}
}

func rdcCleanupJob(wf *rpcWorkflow) *rpcJob { return wf.Jobs["cleanup"] }

// rdcFilterProgram returns the single stale-draft jq program the workflow
// carries (job env STALE_FILTER, shared by the fetch step and the delete loop).
func rdcFilterProgram(wf *rpcWorkflow) string {
	j := rdcCleanupJob(wf)
	if j == nil {
		return ""
	}
	if v, ok := j.Env["STALE_FILTER"].(string); ok {
		return v
	}
	return ""
}

func rdcAllRun(wf *rpcWorkflow) string {
	j := rdcCleanupJob(wf)
	if j == nil {
		return ""
	}
	var b strings.Builder
	for _, s := range j.Steps {
		b.WriteString(s.Run)
		b.WriteString("\n")
	}
	return b.String()
}

// rdcStepByName finds a step of the cleanup job by name.
func rdcStepByName(wf *rpcWorkflow, name string) *rpcStep {
	j := rdcCleanupJob(wf)
	if j == nil {
		return nil
	}
	for i := range j.Steps {
		if j.Steps[i].Name == name {
			return &j.Steps[i]
		}
	}
	return nil
}

// ---- jq execution -----------------------------------------------------------

type rdcJQResult struct {
	selected []string
	exit     int
	stderr   string
}

func rdcRunJQ(t *testing.T, program, input string) rdcJQResult {
	t.Helper()
	cmd := exec.Command("jq", "-r", "--arg", "threshold_days", "30", program)
	cmd.Stdin = strings.NewReader(input)
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	res := rdcJQResult{stderr: strings.TrimSpace(errb.String())}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.exit = ee.ExitCode()
		} else {
			t.Fatalf("run jq: %v", err)
		}
	}
	for _, ln := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if ln != "" {
			res.selected = append(res.selected, ln)
		}
	}
	return res
}

// rdcCheckFilter executes the program against the committed fixture and the
// unresolved-target variants.
func rdcCheckFilter(t *testing.T, program, fixture string) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_FILTER_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	if strings.TrimSpace(program) == "" {
		return []string{"CLEANUP_FILTER_VIOLATION: the workflow carries no stale-draft jq program (job env STALE_FILTER)"}
	}
	res := rdcRunJQ(t, program, fixture)
	if res.exit != 0 {
		add("the program exits %d on the fixture (%s)", res.exit, res.stderr)
	} else if len(res.selected) != 1 || res.selected[0] != "stale-draft" {
		add("selected %v on the fixture, want exactly [stale-draft]", res.selected)
	}
	// E7 / REQ-RPC-013: a draft whose target branch is unresolvable is not selected.
	var entries []map[string]any
	if err := json.Unmarshal([]byte(fixture), &entries); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	for _, target := range []any{"", nil} {
		variant := append([]map[string]any{}, entries...)
		variant = append(variant, map[string]any{"tagName": "unresolved-draft", "isDraft": true, "createdAt": "2020-01-01T00:00:00Z", "targetCommitish": target})
		raw, _ := json.Marshal(variant)
		r := rdcRunJQ(t, program, string(raw))
		for _, s := range r.selected {
			if s == "unresolved-draft" {
				add("a draft with targetCommitish=%v is selected for deletion (the release/* exemption cannot be evaluated)", target)
			}
		}
	}
	if !strings.Contains(program, `startswith("release/")`) {
		add("the program no longer carries the release/* exemption")
	}
	return v
}

// ---- fetch fields -----------------------------------------------------------

var rdcJSONCmdRe = regexp.MustCompile(`gh\s+release\s+(list|view)\b[^\n|;&)]*?--json\s+([A-Za-z,]+)`)

func rdcJoinContinuations(s string) string {
	return regexp.MustCompile(`\\\n\s*`).ReplaceAllString(s, " ")
}

func rdcContainsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func rdcCheckFetchFields(t *testing.T, wf *rpcWorkflow, fixture string) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_FETCH_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	text := rdcJoinContinuations(rdcAllRun(wf))
	matches := rdcJSONCmdRe.FindAllStringSubmatch(text, -1)
	suppliers := 0
	for _, m := range matches {
		sub, fields := m[1], strings.Split(m[2], ",")
		supported := rdcListFields
		if sub == "view" {
			supported = rdcViewFields
		}
		for _, f := range fields {
			if !rdcContainsStr(supported, f) {
				add("gh release %s requests --json field %q, which that subcommand does not support (gh 2.92.0: \"Unknown JSON field\", exit 1)", sub, f)
			}
		}
		all := true
		for _, need := range rdcFilterReads {
			if !rdcContainsStr(fields, need) {
				all = false
			}
		}
		if all {
			suppliers++
		}
	}
	usesAPI := strings.Contains(text, "gh api")
	if usesAPI {
		for _, key := range []string{"tag_name", "draft", "created_at", "target_commitish"} {
			if !strings.Contains(text, key) {
				add("the workflow reads the REST payload but never maps %s to its camelCase name (the filter would read null and select nothing)", key)
			}
		}
	}
	if suppliers == 0 && !usesAPI {
		add("no fetch supplies %v under those names to the filter", rdcFilterReads)
	}
	for _, msg := range rdcCheckFilter(t, rdcFilterProgram(wf), fixture) {
		if strings.Contains(msg, "exemption") || strings.Contains(msg, "unresolved") || strings.Contains(msg, "unresolvable") || strings.Contains(msg, "no stale-draft") {
			v = append(v, strings.Replace(msg, "CLEANUP_FILTER_VIOLATION", "CLEANUP_FETCH_VIOLATION", 1))
		}
	}
	return v
}

// ---- mock gh + step harness -------------------------------------------------

const rdcMockGh = `#!/usr/bin/env bash
# Mock gh for the cleanup guard tests: enforces the measured gh 2.92.0 --json
# field sets, so an unsupported field fails here as it does on the runner.
set -u
state="${MOCK_GH_DIR:?}"
echo "$*" >> "$state/calls.log"
list_fields="createdAt isDraft isImmutable isLatest isPrerelease name publishedAt tagName"
view_fields="apiUrl assets author body createdAt databaseId id isDraft isImmutable isPrerelease name publishedAt tagName tarballUrl targetCommitish uploadUrl url zipballUrl"
check_fields() {
  IFS=',' read -ra want <<< "$2"
  for f in "${want[@]}"; do
    case " $1 " in *" $f "*) ;; *) echo "Unknown JSON field: \"$f\"" >&2; exit 1 ;; esac
  done
}
json_flag() {
  while [ $# -gt 0 ]; do
    if [ "$1" = "--json" ]; then echo "$2"; return; fi
    shift
  done
}
if [ "$1" = "release" ] && [ "$2" = "list" ]; then
  fields="$(json_flag "$@")"
  check_fields "$list_fields" "$fields"
  jq -c --arg f "$fields" 'map(. as $o | ($f | split(",")) | map({(.): $o[.]}) | add)' "$state/list.json"
  exit 0
fi
if [ "$1" = "release" ] && [ "$2" = "view" ]; then
  tag="$3"
  fields="$(json_flag "$@")"
  check_fields "$view_fields" "$fields"
  f="$state/view/$tag.json"
  [ -f "$f" ] || { echo "release not found: $tag" >&2; exit 1; }
  jq -c --arg f "$fields" '. as $o | ($f | split(",")) | map({(.): $o[.]}) | add' "$f"
  exit 0
fi
if [ "$1" = "release" ] && [ "$2" = "delete" ]; then
  echo "$3" >> "$state/delete.log"
  exit 0
fi
echo "mock gh: unsupported invocation: $*" >&2
exit 2
`

type rdcHarness struct {
	t       *testing.T
	dir     string
	mock    string
	wf      *rpcWorkflow
	outputs map[string]map[string]string
	// dispatchDryRun is nil for a scheduled run (inputs.dry_run is null).
	dispatchDryRun *bool
	seq            int
}

func rdcNewHarness(t *testing.T, wf *rpcWorkflow, fixture string, dispatchDryRun *bool) *rdcHarness {
	t.Helper()
	rdcRequireTools(t, true)
	dir := t.TempDir()
	h := &rdcHarness{t: t, dir: dir, mock: filepath.Join(dir, "mock"), wf: wf, outputs: map[string]map[string]string{}, dispatchDryRun: dispatchDryRun}
	for _, d := range []string{filepath.Join(h.mock, "view"), filepath.Join(dir, "bin"), filepath.Join(dir, "runner-temp")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "gh"), []byte(rdcMockGh), 0o755); err != nil {
		t.Fatalf("write mock gh: %v", err)
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(fixture), &entries); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	// The listing carries what `gh release list` can supply; each release also
	// has a view payload with targetCommitish.
	raw, _ := json.Marshal(entries)
	if err := os.WriteFile(filepath.Join(h.mock, "list.json"), raw, 0o644); err != nil {
		t.Fatalf("write list: %v", err)
	}
	for _, e := range entries {
		h.setView(e["tagName"].(string), e)
	}
	return h
}

func (h *rdcHarness) setView(tag string, payload map[string]any) {
	h.t.Helper()
	raw, _ := json.Marshal(payload)
	if err := os.WriteFile(filepath.Join(h.mock, "view", tag+".json"), raw, 0o644); err != nil {
		h.t.Fatalf("write view %s: %v", tag, err)
	}
}

func (h *rdcHarness) readLines(name string) []string {
	raw, err := os.ReadFile(filepath.Join(h.mock, name))
	if err != nil {
		return nil
	}
	var out []string
	for _, ln := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

func (h *rdcHarness) resetCalls() {
	_ = os.Remove(filepath.Join(h.mock, "calls.log"))
}

func (h *rdcHarness) ctx() rpcExprCtx {
	vars := map[string]string{"secrets.GITHUB_TOKEN": ""}
	for id, kv := range h.outputs {
		for k, v := range kv {
			vars["steps."+id+".outputs."+k] = v
		}
	}
	c := rpcExprCtx{vars: vars, bools: map[string]bool{}}
	if h.dispatchDryRun != nil {
		c.bools["inputs.dry_run"] = *h.dispatchDryRun
	}
	return c
}

// runStep executes one step's script the way the runner would: expressions
// interpolated, shell flags per the step's shell, step and job env applied.
func (h *rdcHarness) runStep(step *rpcStep) (string, int) {
	h.t.Helper()
	h.seq++
	ctx := h.ctx()
	script := rpcInterpolate(h.t, step.Run, ctx)
	scriptPath := filepath.Join(h.dir, fmt.Sprintf("step-%d.sh", h.seq))
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		h.t.Fatalf("write step script: %v", err)
	}
	outPath := filepath.Join(h.dir, fmt.Sprintf("out-%d", h.seq))
	sumPath := filepath.Join(h.dir, fmt.Sprintf("summary-%d", h.seq))
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "MOAI_") || strings.HasPrefix(kv, "GH_") || strings.HasPrefix(kv, "GITHUB_") || strings.HasPrefix(kv, "INPUT_") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"PATH="+filepath.Join(h.dir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"RUNNER_TEMP="+filepath.Join(h.dir, "runner-temp"),
		"GITHUB_OUTPUT="+outPath,
		"GITHUB_STEP_SUMMARY="+sumPath,
		"MOCK_GH_DIR="+h.mock,
	)
	for k, v := range rdcCleanupJob(h.wf).Env {
		env = append(env, k+"="+rpcInterpolate(h.t, fmt.Sprint(v), ctx))
	}
	for k, v := range step.Env {
		env = append(env, k+"="+rpcInterpolate(h.t, fmt.Sprint(v), ctx))
	}
	args := []string{"-e", scriptPath}
	if step.Shell == "bash" {
		args = []string{"--noprofile", "--norc", "-eo", "pipefail", scriptPath}
	}
	cmd := exec.Command("bash", args...)
	cmd.Env = env
	cmd.Dir = h.dir
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			h.t.Fatalf("run step %q: %v", step.Name, err)
		}
		code = ee.ExitCode()
	}
	if step.ID != "" {
		kv := map[string]string{}
		if raw, rerr := os.ReadFile(outPath); rerr == nil {
			for _, ln := range strings.Split(string(raw), "\n") {
				if k, v, ok := strings.Cut(ln, "="); ok {
					kv[k] = v
				}
			}
		}
		h.outputs[step.ID] = kv
	}
	return string(out), code
}

// runUpToFetch runs the env and fetch steps; it returns the fetch step's exit.
func (h *rdcHarness) runUpToFetch() (string, int) {
	h.t.Helper()
	envStep := rdcStepByName(h.wf, "Set cleanup environment")
	fetch := rdcStepByName(h.wf, "Fetch stale drafts")
	if envStep == nil || fetch == nil {
		h.t.Fatalf("cleanup workflow lacks the env or fetch step")
	}
	if out, code := h.runStep(envStep); code != 0 {
		return out, code
	}
	return h.runStep(fetch)
}

// ---- AC-RPC-013 --------------------------------------------------------------

func rdcLoad(t *testing.T) (*rpcWorkflow, string, string) {
	t.Helper()
	text := rpcReadRepoFile(t, rdcCleanupRel)
	return rpcParseWorkflow(t, rdcCleanupRel, text), text, rpcReadRepoFile(t, rdcFixtureRel)
}

func TestCleanupFetchFieldsSupported(t *testing.T) {
	rdcRequireTools(t, false)
	wf, text, fixture := rdcLoad(t)
	for _, msg := range rdcCheckFetchFields(t, wf, fixture) {
		t.Error(msg)
	}

	baselineClean := len(rdcCheckFetchFields(t, wf, fixture)) == 0
	mutants := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"targetCommitish requested from gh release list (the original defect)", func(w *rpcWorkflow) {
			s := rdcStepByName(w, "Fetch stale drafts")
			// The listing is a backslash-continued command: the --json flag sits
			// on its own line directly after `gh release list \`.
			s.Run = strings.Replace(s.Run, "--json tagName,isDraft,createdAt \\", "--json tagName,isDraft,createdAt,targetCommitish \\", 1)
		}},
		{"drop the field without re-sourcing it", func(w *rpcWorkflow) {
			for i := range rdcCleanupJob(w).Steps {
				s := &rdcCleanupJob(w).Steps[i]
				s.Run = strings.ReplaceAll(s.Run, ",targetCommitish", "")
			}
		}},
		{"REST source without the snake_case to camelCase mapping", func(w *rpcWorkflow) {
			s := rdcStepByName(w, "Fetch stale drafts")
			s.Run = "gh api repos/${GITHUB_REPOSITORY}/releases > \"$RUNNER_TEMP/releases.json\"\n"
		}},
		{"release/* exemption dropped to make the fetch succeed", func(w *rpcWorkflow) {
			rdcCleanupJob(w).Env["STALE_FILTER"] = strings.ReplaceAll(rdcFilterProgram(w), `(.targetCommitish | startswith("release/") | not)`, "true")
		}},
	}
	for _, m := range mutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real workflow is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rdcCleanupRel, text)
			before := rdcAllRun(mut) + rdcFilterProgram(mut)
			m.mutate(mut)
			if rdcAllRun(mut)+rdcFilterProgram(mut) == before {
				t.Fatalf("control did not apply")
			}
			got := rdcCheckFetchFields(t, mut, fixture)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
}

// ---- AC-RPC-010 --------------------------------------------------------------

func TestCleanupFilterSelectsStaleDrafts(t *testing.T) {
	rdcRequireTools(t, false)
	wf, _, fixture := rdcLoad(t)
	program := rdcFilterProgram(wf)
	for _, msg := range rdcCheckFilter(t, program, fixture) {
		t.Error(msg)
	}

	baselineClean := len(rdcCheckFilter(t, program, fixture)) == 0
	// EV-3 vs EV-4 divergence on the same fixture: the pre-fix binding form
	// errors (jq exit 5), the fixed form selects the stale draft.
	brokenForm := strings.Replace(program, "(now|floor) as $now |", "now | floor as $now |", 1)
	snake := strings.NewReplacer(`"tagName"`, `"tag_name"`, `"isDraft"`, `"draft"`, `"createdAt"`, `"created_at"`, `"targetCommitish"`, `"target_commitish"`).Replace(fixture)
	controls := []struct {
		name    string
		program string
		fixture string
	}{
		{"pre-fix `now | floor as $now` binding (EV-3)", brokenForm, fixture},
		{"release/* exemption removed", strings.ReplaceAll(program, `(.targetCommitish | startswith("release/") | not)`, "true"), fixture},
		{"threshold comparison inverted", strings.Replace(program, "> $threshold_secs", "< $threshold_secs", 1), fixture},
		{"draft flag not required", strings.Replace(program, ".isDraft == true and", "", 1), fixture},
		{"snake_case payload under the camelCase filter", program, snake},
	}
	for _, c := range controls {
		t.Run("control/"+c.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real program is not clean for this guard, so the control proves nothing")
			}
			if c.program == program && c.fixture == fixture {
				t.Fatalf("control did not apply")
			}
			got := rdcCheckFilter(t, c.program, c.fixture)
			if len(got) == 0 {
				t.Fatalf("control %q passes the guard", c.name)
			}
			t.Logf("control observed red: %s", got[0])
		})
	}
	if baselineClean && brokenForm != program {
		r := rdcRunJQ(t, brokenForm, fixture)
		t.Logf("EV-3 re-observed: pre-fix binding exits %d (%s)", r.exit, r.stderr)
		if r.exit != 5 {
			t.Errorf("the pre-fix binding form exited %d, want 5 (Cannot iterate over number)", r.exit)
		}
	}
}

// ---- AC-RPC-011 --------------------------------------------------------------

func rdcCheckRequery(t *testing.T, wf *rpcWorkflow, fixture string) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_REQUERY_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	del := rdcStepByName(wf, "Delete stale drafts")
	if del == nil {
		return []string{"CLEANUP_REQUERY_VIOLATION: the workflow has no Delete stale drafts step"}
	}
	text := rdcJoinContinuations(del.Run)
	iView, iDel := strings.Index(text, "gh release view"), strings.Index(text, "gh release delete")
	if iView < 0 || iDel < 0 || iView > iDel {
		add("the delete step does not re-query (gh release view) before gh release delete (view at %d, delete at %d)", iView, iDel)
	}

	scenarios := []struct {
		name       string
		mutate     func(h *rdcHarness)
		wantDelete []string
		wantSkip   bool
	}{
		{"unchanged draft is deleted", func(h *rdcHarness) {}, []string{"stale-draft"}, false},
		{"published between listing and deletion", func(h *rdcHarness) {
			h.setView("stale-draft", map[string]any{"tagName": "stale-draft", "isDraft": false, "createdAt": "2020-01-01T00:00:00Z", "targetCommitish": "main"})
		}, nil, true},
		{"retargeted at release/* between listing and deletion", func(h *rdcHarness) {
			h.setView("stale-draft", map[string]any{"tagName": "stale-draft", "isDraft": true, "createdAt": "2020-01-01T00:00:00Z", "targetCommitish": "release/v3.2.0"})
		}, nil, true},
		{"refreshed so it is younger than the threshold", func(h *rdcHarness) {
			h.setView("stale-draft", map[string]any{"tagName": "stale-draft", "isDraft": true, "createdAt": "2999-01-01T00:00:00Z", "targetCommitish": "main"})
		}, nil, true},
	}
	for _, sc := range scenarios {
		h := rdcNewHarness(t, wf, fixture, nil)
		if out, code := h.runUpToFetch(); code != 0 {
			add("%s: the env/fetch steps exit %d before the delete loop is reached: %s", sc.name, code, strings.TrimSpace(out))
			continue
		}
		sc.mutate(h)
		h.resetCalls()
		out, code := h.runStep(del)
		if code != 0 {
			add("%s: the delete step exits %d: %s", sc.name, code, strings.TrimSpace(out))
			continue
		}
		deleted := h.readLines("delete.log")
		if strings.Join(deleted, ",") != strings.Join(sc.wantDelete, ",") {
			add("%s: deleted %v, want %v", sc.name, deleted, sc.wantDelete)
		}
		if sc.wantSkip && !strings.Contains(out, "[SKIP] stale-draft") {
			add("%s: the skip is not recorded in the step output (%q)", sc.name, strings.TrimSpace(out))
		}
		// Every delete must be preceded by a view of the same tag within this step.
		calls := h.readLines("calls.log")
		for i, c := range calls {
			if tag, ok := strings.CutPrefix(c, "release delete "); ok {
				tag = strings.Fields(tag)[0]
				seen := false
				for _, prev := range calls[:i] {
					if strings.HasPrefix(prev, "release view "+tag+" ") {
						seen = true
					}
				}
				if !seen {
					add("%s: gh release delete %s was not preceded by a gh release view %s", sc.name, tag, tag)
				}
			}
		}
	}
	return v
}

const rdcUnguardedDeleteLoop = `dry_run="${DRY_RUN}"
while IFS= read -r tag; do
  [ -z "$tag" ] && continue
  if [ "$dry_run" = "1" ]; then
    echo "[DRY-RUN] Would delete: ${tag}"
  else
    echo "Deleting stale draft: ${tag}"
    gh release delete "$tag" --yes
  fi
done < "$RUNNER_TEMP/stale_tags.txt"
`

func TestCleanupRequeriesBeforeDelete(t *testing.T) {
	rdcRequireTools(t, true)
	wf, text, fixture := rdcLoad(t)
	for _, msg := range rdcCheckRequery(t, wf, fixture) {
		t.Error(msg)
	}

	baselineClean := len(rdcCheckRequery(t, wf, fixture)) == 0
	t.Run("control/delete straight from the listing (pre-fix loop)", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real workflow is not clean for this guard, so the control proves nothing")
		}
		mut := rpcParseWorkflow(t, rdcCleanupRel, text)
		rdcStepByName(mut, "Delete stale drafts").Run = rdcUnguardedDeleteLoop
		got := rdcCheckRequery(t, mut, fixture)
		if len(got) == 0 {
			t.Fatalf("a delete loop with no re-query passes the guard")
		}
		t.Logf("control observed red: %s", got[0])
	})
	t.Run("control/re-query result ignored", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real workflow is not clean for this guard, so the control proves nothing")
		}
		mut := rpcParseWorkflow(t, rdcCleanupRel, text)
		s := rdcStepByName(mut, "Delete stale drafts")
		// Keep the gh release view call but drop the comparison that acts on it.
		re := regexp.MustCompile(`(?s)if \[ "\$still" != "\$tag" \]; then.*?fi\n`)
		if !re.MatchString(s.Run) {
			t.Fatalf("control did not apply: the still-stale comparison is not where the control expects it")
		}
		s.Run = re.ReplaceAllString(s.Run, "")
		got := rdcCheckRequery(t, mut, fixture)
		if len(got) == 0 {
			t.Fatalf("a re-query whose result is ignored passes the guard")
		}
		t.Logf("control observed red: %s", got[0])
	})
}

// ---- AC-RPC-012 --------------------------------------------------------------

func rdcCheckConcurrency(cleanup, drafter *rpcWorkflow) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_CONCURRENCY_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	groups := map[string]string{}
	for _, w := range []*rpcWorkflow{cleanup, drafter} {
		g := w.Concurrency.Group
		if g == "" {
			add("%s declares no concurrency group", w.Path)
			continue
		}
		var seen []string
		for _, ref := range []string{"refs/heads/main", "refs/heads/release/v3.2.0"} {
			resolved := strings.ReplaceAll(g, "${{ github.ref }}", ref)
			seen = append(seen, resolved)
			if strings.Contains(resolved, "${{") {
				add("%s: the group key %q is not a literal (unresolved expression after substituting github.ref)", w.Path, g)
			}
		}
		if len(seen) == 2 && seen[0] != seen[1] {
			add("%s: the group key varies with github.ref (%q vs %q), so a dispatched drafter on another ref escapes the group", w.Path, seen[0], seen[1])
		}
		groups[w.Path] = seen[0]
		if c, ok := w.Concurrency.CancelInProgress.(bool); !ok || c {
			add("%s: cancel-in-progress is %v, want false on BOTH sides (the incoming run's own flag governs a shared group)", w.Path, w.Concurrency.CancelInProgress)
		}
	}
	if len(groups) == 2 && groups[cleanup.Path] != groups[drafter.Path] {
		add("the two workflows use different groups (%q vs %q), so they do not serialize", groups[cleanup.Path], groups[drafter.Path])
	}
	return v
}

func TestDrafterCleanupConcurrency(t *testing.T) {
	cleanupText := rpcReadRepoFile(t, rdcCleanupRel)
	drafterText := rpcReadRepoFile(t, rdcDrafterRel)
	cleanup := rpcParseWorkflow(t, rdcCleanupRel, cleanupText)
	drafter := rpcParseWorkflow(t, rdcDrafterRel, drafterText)
	for _, msg := range rdcCheckConcurrency(cleanup, drafter) {
		t.Error(msg)
	}

	baselineClean := len(rdcCheckConcurrency(cleanup, drafter)) == 0
	mutants := []struct {
		name   string
		mutate func(c, d *rpcWorkflow)
	}{
		{"ref-keyed drafter group (the original)", func(c, d *rpcWorkflow) { d.Concurrency.Group = "release-drafter-${{ github.ref }}" }},
		{"drafter cancels in progress", func(c, d *rpcWorkflow) { d.Concurrency.CancelInProgress = true }},
		{"cleanup cancels in progress", func(c, d *rpcWorkflow) { c.Concurrency.CancelInProgress = true }},
		{"different literal groups", func(c, d *rpcWorkflow) { c.Concurrency.Group = "release-drafter-cleanup" }},
		{"cleanup has no group", func(c, d *rpcWorkflow) { c.Concurrency.Group = "" }},
	}
	for _, m := range mutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			c := rpcParseWorkflow(t, rdcCleanupRel, cleanupText)
			d := rpcParseWorkflow(t, rdcDrafterRel, drafterText)
			m.mutate(c, d)
			got := rdcCheckConcurrency(c, d)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
}

// ---- dry-run reachability and value contract ----------------------------------

func rdcCheckDryRun(t *testing.T, wf *rpcWorkflow, fixture string) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_DRYRUN_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	if !strings.Contains(strings.Join(rdcDispatchInputs(wf), ","), "dry_run") {
		add("workflow_dispatch declares no dry_run input, so the dry-run path is unreachable by dispatch")
	}
	del := rdcStepByName(wf, "Delete stale drafts")
	if del == nil {
		return append(v, "CLEANUP_DRYRUN_VIOLATION: no Delete stale drafts step")
	}
	yes, no := true, false
	cases := []struct {
		name       string
		input      *bool
		wantDry    string
		wantDelete []string
	}{
		{"dispatched with dry_run=true", &yes, "1", nil},
		{"dispatched with dry_run=false", &no, "0", []string{"stale-draft"}},
		{"scheduled run (input null)", nil, "0", []string{"stale-draft"}},
	}
	for _, c := range cases {
		h := rdcNewHarness(t, wf, fixture, c.input)
		if out, code := h.runUpToFetch(); code != 0 {
			add("%s: the env/fetch steps exit %d: %s", c.name, code, strings.TrimSpace(out))
			continue
		}
		if got := h.outputs["env"]["dry_run"]; got != c.wantDry {
			add("%s: the env step resolved dry_run=%q, want the literal %q (the delete loop honors only \"1\")", c.name, got, c.wantDry)
		}
		out, code := h.runStep(del)
		if code != 0 {
			add("%s: the delete step exits %d: %s", c.name, code, strings.TrimSpace(out))
			continue
		}
		deleted := h.readLines("delete.log")
		if strings.Join(deleted, ",") != strings.Join(c.wantDelete, ",") {
			add("%s: gh release delete ran for %v, want %v", c.name, deleted, c.wantDelete)
		}
		if c.wantDry == "1" && !strings.Contains(out, "[DRY-RUN] Would delete: stale-draft") {
			add("%s: the dry run did not list the stale draft (%q)", c.name, strings.TrimSpace(out))
		}
	}
	return v
}

// rdcDispatchInputs lists the workflow_dispatch input names.
func rdcDispatchInputs(wf *rpcWorkflow) []string {
	var names []string
	if wf.On.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(wf.On.Content); i += 2 {
		if wf.On.Content[i].Value != "workflow_dispatch" {
			continue
		}
		wd := wf.On.Content[i+1]
		for k := 0; k+1 < len(wd.Content); k += 2 {
			if wd.Content[k].Value != "inputs" {
				continue
			}
			in := wd.Content[k+1]
			for n := 0; n+1 < len(in.Content); n += 2 {
				names = append(names, in.Content[n].Value)
			}
		}
	}
	sort.Strings(names)
	return names
}

func TestCleanupDryRunDispatchReachable(t *testing.T) {
	rdcRequireTools(t, true)
	wf, text, fixture := rdcLoad(t)
	for _, msg := range rdcCheckDryRun(t, wf, fixture) {
		t.Error(msg)
	}

	baselineClean := len(rdcCheckDryRun(t, wf, fixture)) == 0
	t.Run("control/boolean input passed through as its string form", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real workflow is not clean for this guard, so the control proves nothing")
		}
		mut := rpcParseWorkflow(t, rdcCleanupRel, text)
		s := rdcStepByName(mut, "Set cleanup environment")
		changed := false
		for k, v := range s.Env {
			if strings.Contains(fmt.Sprint(v), "inputs.dry_run") {
				s.Env[k] = "${{ inputs.dry_run }}"
				changed = true
			}
		}
		if !changed {
			t.Fatalf("control did not apply: the env step does not wire inputs.dry_run")
		}
		got := rdcCheckDryRun(t, mut, fixture)
		if len(got) == 0 {
			t.Fatalf("a stringified boolean (\"true\" != \"1\") passes the guard")
		}
		t.Logf("control observed red: %s", got[0])
	})
	t.Run("control/no dry_run input declared", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real workflow is not clean for this guard, so the control proves nothing")
		}
		mut := rpcParseWorkflow(t, rdcCleanupRel, strings.Replace(text, "      dry_run:", "      removed_input:", 1))
		got := rdcCheckDryRun(t, mut, fixture)
		if len(got) == 0 {
			t.Fatalf("a workflow with no dry_run dispatch input passes the guard")
		}
		t.Logf("control observed red: %s", got[0])
	})
}
