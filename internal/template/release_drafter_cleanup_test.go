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
//   - the cleanup had no concurrency group and Release Drafter's own was
//     ref-keyed with cancel-in-progress: true; the v0.7.x repair joined both in
//     one shared group, which lets a newer pending drafter run replace the
//     pending weekly cleanup, so since v0.8.0 the two carry different literal
//     groups and the cleanup's re-query plus idempotent delete carry the safety;
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
  # A marker file makes the delete of that tag fail the way gh does: a draft that
  # is already gone answers HTTP 404, any other failure answers something else.
  if [ -f "$state/delete_notfound/$3" ]; then
    echo "HTTP 404: Not Found (https://api.github.com/repos/o/r/releases/1)" >&2
    exit 1
  fi
  if [ -f "$state/delete_error/$3" ]; then
    echo "HTTP 500: Server Error (https://api.github.com/repos/o/r/releases/1)" >&2
    exit 1
  fi
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

// mark drops a marker file the mock gh reads (see rdcMockGh) for one tag.
func (h *rdcHarness) mark(dir, tag string) {
	h.t.Helper()
	d := filepath.Join(h.mock, dir)
	if err := os.MkdirAll(d, 0o755); err != nil {
		h.t.Fatalf("mkdir %s: %v", d, err)
	}
	if err := os.WriteFile(filepath.Join(d, tag), []byte("1"), 0o644); err != nil {
		h.t.Fatalf("write marker %s/%s: %v", dir, tag, err)
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
	// A run that is not a workflow_dispatch has no inputs: inputs.dry_run reads
	// false there, which is a modelled value, not an unmodelled context.
	c.bools["inputs.dry_run"] = h.dispatchDryRun != nil && *h.dispatchDryRun
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
//
// Amended v0.8.0: the drafter and the cleanup carry DIFFERENT ref-independent
// literal groups (a shared group lets a newer pending drafter run replace a
// pending weekly cleanup), and safety between them rests on the re-query before
// every delete plus an idempotent delete.

// rdcYAMLNode parses a YAML fragment into its content node (a control helper).
func rdcYAMLNode(t *testing.T, text string) yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil || len(doc.Content) == 0 {
		t.Fatalf("control text does not parse: %v", err)
	}
	return *doc.Content[0]
}

// rdcGroupKey normalizes a concurrency group name for comparison.
func rdcGroupKey(g string) string { return strings.ToLower(strings.TrimSpace(g)) }

// rdcCheckConcurrency is the static part of the amended AC-RPC-012: (a) literal,
// expression-free, ref-invariant groups, (b) the two groups differ, (c)
// cancel-in-progress false on both, (d) the re-query precedes the delete.
func rdcCheckConcurrency(cleanup, drafter *rpcWorkflow) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_CONCURRENCY_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	// Every concurrency group a workflow declares, at workflow level AND on any job
	// (a job-level `concurrency:` joins the same repository-wide group namespace,
	// so a cleanup JOB carrying the drafter's group collides exactly like a shared
	// workflow-level group would).
	type declared struct {
		where, group string
		cancel       any
	}
	byPath := map[string][]declared{}
	for _, w := range []*rpcWorkflow{cleanup, drafter} {
		if w.Concurrency.Group == "" {
			add("%s declares no concurrency group", w.Path)
		} else {
			byPath[w.Path] = append(byPath[w.Path], declared{w.Path, w.Concurrency.Group, w.Concurrency.CancelInProgress})
		}
		for _, id := range w.jobIDs() {
			// An alias (`concurrency: *group`) is read as the node it stands for; any
			// other node kind is refused rather than skipped.
			n := rpcResolveAlias(&w.Jobs[id].Concurrency)
			switch n.Kind {
			case 0:
				// no job-level concurrency block
			case yaml.ScalarNode:
				byPath[w.Path] = append(byPath[w.Path], declared{fmt.Sprintf("%s job %q", w.Path, id), n.Value, false})
			case yaml.MappingNode:
				var jc struct {
					Group            string `yaml:"group"`
					CancelInProgress any    `yaml:"cancel-in-progress"`
				}
				if err := n.Decode(&jc); err != nil {
					add("%s job %q: its concurrency block cannot be read: %v", w.Path, id, err)
					continue
				}
				byPath[w.Path] = append(byPath[w.Path], declared{fmt.Sprintf("%s job %q", w.Path, id), jc.Group, jc.CancelInProgress})
			default:
				add("%s job %q: its concurrency block has a YAML node kind (%d) this guard does not read", w.Path, id, n.Kind)
			}
		}
		for _, d := range byPath[w.Path] {
			g := d.group
			if strings.Contains(g, "${{") {
				add("%s: the group key %q contains an expression; it must be a plain literal, because any ref- or run-dependent value (github.ref, github.run_id, ...) splits the group per ref or per run", d.where, g)
			}
			var seen []string
			for _, ref := range []string{"refs/heads/main", "refs/heads/release/v3.2.0"} {
				seen = append(seen, strings.ReplaceAll(g, "${{ github.ref }}", ref))
			}
			if seen[0] != seen[1] {
				add("%s: the group key varies with github.ref (%q vs %q), so a dispatched run on another ref escapes the group", d.where, seen[0], seen[1])
			}
			if c, ok := d.cancel.(bool); !ok || c {
				add("%s: cancel-in-progress is %v, want false on BOTH sides", d.where, d.cancel)
			}
		}
	}
	// Group names are compared as GitHub compares them: case-insensitively, with
	// surrounding whitespace dropped (sync-audit iter4 B4: `Release-Drafter-X` vs
	// `release-drafter-x` is ONE group to GitHub and two to a byte comparison).
	for _, c := range byPath[cleanup.Path] {
		for _, d := range byPath[drafter.Path] {
			if rdcGroupKey(c.group) == rdcGroupKey(d.group) {
				add("%s and %s share the group %q (compared case-insensitively, as GitHub does): GitHub keeps one running and one pending run per group and a newer pending run replaces the older pending one, so a drafter run can replace the pending weekly cleanup; they must use different groups", c.where, d.where, c.group)
			}
		}
	}
	del := rdcStepByName(cleanup, "Delete stale drafts")
	if del == nil {
		add("the cleanup workflow has no Delete stale drafts step")
		return v
	}
	text := rdcJoinContinuations(del.Run)
	iView, iDel := strings.Index(text, "gh release view"), strings.Index(text, "gh release delete")
	if iView < 0 || iDel < 0 || iView > iDel {
		add("the delete step does not re-query (gh release view) before gh release delete (view at %d, delete at %d)", iView, iDel)
	}
	return v
}

// rdcCheckIdempotentDelete is clause (e): under the mock gh, a draft that is
// already gone (absent at its re-query, or reported not found by the delete)
// leaves the step at exit 0 with an already-gone line and no further delete,
// while any other delete failure still fails the step. Every delete is also
// preceded by a re-query of the same tag in the recorded gh calls.
func rdcCheckIdempotentDelete(t *testing.T, wf *rpcWorkflow, fixture string) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CLEANUP_CONCURRENCY_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	del := rdcStepByName(wf, "Delete stale drafts")
	if del == nil {
		return []string{"CLEANUP_CONCURRENCY_VIOLATION: the cleanup workflow has no Delete stale drafts step"}
	}
	scenarios := []struct {
		name        string
		mutate      func(h *rdcHarness)
		wantFail    bool
		wantGone    bool
		wantDeletes int
	}{
		{"draft absent at its re-query", func(h *rdcHarness) {
			if err := os.Remove(filepath.Join(h.mock, "view", "stale-draft.json")); err != nil {
				h.t.Fatalf("remove view payload: %v", err)
			}
		}, false, true, 0},
		{"delete reports the draft not found", func(h *rdcHarness) { h.mark("delete_notfound", "stale-draft") }, false, true, 1},
		{"delete fails for another reason", func(h *rdcHarness) { h.mark("delete_error", "stale-draft") }, true, false, 1},
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
		if sc.wantFail && code == 0 {
			add("%s: the delete step exits 0, want a failure (only a draft that is already gone is forgiven): %s", sc.name, strings.TrimSpace(out))
		}
		if !sc.wantFail && code != 0 {
			add("%s: the delete step exits %d, want 0 (an already-gone draft must not fail the run): %s", sc.name, code, strings.TrimSpace(out))
		}
		gone := strings.Contains(strings.ToLower(out), "already gone")
		if sc.wantGone && !gone {
			add("%s: the step output has no already-gone line (%q)", sc.name, strings.TrimSpace(out))
		}
		if !sc.wantGone && gone {
			add("%s: the step reports the draft already gone for a failure that is not a not-found (%q)", sc.name, strings.TrimSpace(out))
		}
		deletes := 0
		calls := h.readLines("calls.log")
		for i, c := range calls {
			tag, ok := strings.CutPrefix(c, "release delete ")
			if !ok {
				continue
			}
			deletes++
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
		if deletes != sc.wantDeletes {
			add("%s: %d delete call(s), want %d", sc.name, deletes, sc.wantDeletes)
		}
		if got := h.readLines("delete.log"); len(got) != 0 {
			add("%s: a delete succeeded (%v) although the draft was gone or the delete was made to fail", sc.name, got)
		}
	}
	return v
}

// rdcReplaceLast replaces the last occurrence of old in s, reporting whether
// one was found.
func rdcReplaceLast(s, old, repl string) (string, bool) {
	i := strings.LastIndex(s, old)
	if i < 0 {
		return s, false
	}
	return s[:i] + repl + s[i+len(old):], true
}

const rdcDeleteBeforeRequeryLoop = `dry_run="${DRY_RUN}"
while IFS= read -r tag; do
  [ -z "$tag" ] && continue
  gh release delete "$tag" --yes
  gh release view "$tag" --json tagName,isDraft,createdAt,targetCommitish > /dev/null
done < "$RUNNER_TEMP/stale_tags.txt"
`

func TestDrafterCleanupConcurrency(t *testing.T) {
	cleanupText := rpcReadRepoFile(t, rdcCleanupRel)
	drafterText := rpcReadRepoFile(t, rdcDrafterRel)
	fixture := rpcReadRepoFile(t, rdcFixtureRel)
	// The mock-gh scenarios (clause e) drive the workflow's bash; windows runs
	// the static clauses only (the same precedent as the other cleanup guards).
	harness := runtime.GOOS != "windows"
	rdcRequireTools(t, false)

	check := func(c, d *rpcWorkflow) []string {
		v := rdcCheckConcurrency(c, d)
		if harness {
			v = append(v, rdcCheckIdempotentDelete(t, c, fixture)...)
		}
		return v
	}
	cleanup := rpcParseWorkflow(t, rdcCleanupRel, cleanupText)
	drafter := rpcParseWorkflow(t, rdcDrafterRel, drafterText)
	for _, msg := range check(cleanup, drafter) {
		t.Error(msg)
	}

	baselineClean := len(check(cleanup, drafter)) == 0
	mutants := []struct {
		name     string
		needBash bool
		mutate   func(t *testing.T, c, d *rpcWorkflow)
	}{
		{"ref-keyed drafter group (the original)", false, func(t *testing.T, c, d *rpcWorkflow) { d.Concurrency.Group = "release-drafter-${{ github.ref }}" }},
		{"drafter cancels in progress", false, func(t *testing.T, c, d *rpcWorkflow) { d.Concurrency.CancelInProgress = true }},
		{"cleanup cancels in progress (separate groups)", false, func(t *testing.T, c, d *rpcWorkflow) { c.Concurrency.CancelInProgress = true }},
		{"cleanup has no group", false, func(t *testing.T, c, d *rpcWorkflow) { c.Concurrency.Group = "" }},
		{"groups equal again (the v0.7.x shared group)", false, func(t *testing.T, c, d *rpcWorkflow) { c.Concurrency.Group = d.Concurrency.Group }},
		// sync-audit iter4 B4: GitHub compares group names case-insensitively.
		{"groups equal apart from letter case (cleanup upper-cased)", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Concurrency.Group = strings.ToUpper(d.Concurrency.Group)
		}},
		{"groups equal apart from letter case (drafter capitalised)", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Concurrency.Group = d.Concurrency.Group
			d.Concurrency.Group = strings.ToUpper(d.Concurrency.Group[:1]) + d.Concurrency.Group[1:]
		}},
		{"groups equal apart from surrounding whitespace", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Concurrency.Group = "  " + d.Concurrency.Group + " "
		}},
		// A job-level group joins the same repository-wide namespace.
		{"cleanup JOB carries the drafter's group", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Jobs["cleanup"].Concurrency = rdcYAMLNode(t, "group: "+d.Concurrency.Group+"\ncancel-in-progress: false\n")
		}},
		{"cleanup JOB carries the drafter's group in another case", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Jobs["cleanup"].Concurrency = rdcYAMLNode(t, "group: "+strings.ToUpper(d.Concurrency.Group)+"\ncancel-in-progress: false\n")
		}},
		{"cleanup JOB carries the drafter's group as a scalar", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Jobs["cleanup"].Concurrency = rdcYAMLNode(t, d.Concurrency.Group+"\n")
		}},
		// sync-audit iter5 F-4: a job-level group written through a YAML alias is read
		// as the group it stands for; a node kind the guard cannot read is refused.
		{"iter5 D1 failure: cleanup JOB group through a YAML alias of the drafter's group in another case", false, func(t *testing.T, c, d *rpcWorkflow) {
			target := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: strings.ToUpper(d.Concurrency.Group)}
			c.Jobs["cleanup"].Concurrency = yaml.Node{Kind: yaml.AliasNode, Alias: target}
		}},
		{"iter5 D1 failure: cleanup JOB concurrency block is a sequence the guard cannot read", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Jobs["cleanup"].Concurrency = rdcYAMLNode(t, "- not-a-group\n")
		}},
		{"cleanup JOB group is ref-keyed", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Jobs["cleanup"].Concurrency = rdcYAMLNode(t, "group: x-${{ github.ref }}\ncancel-in-progress: false\n")
		}},
		{"cleanup JOB cancels in progress", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Jobs["cleanup"].Concurrency = rdcYAMLNode(t, "group: cleanup-job-only\ncancel-in-progress: true\n")
		}},
		{"cleanup group carries github.ref", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Concurrency.Group = "release-drafter-cleanup-${{ github.ref }}"
		}},
		{"cleanup group carries github.run_id", false, func(t *testing.T, c, d *rpcWorkflow) {
			c.Concurrency.Group = "release-drafter-cleanup-${{ github.run_id }}"
		}},
		{"delete before the re-query", false, func(t *testing.T, c, d *rpcWorkflow) {
			rdcStepByName(c, "Delete stale drafts").Run = rdcDeleteBeforeRequeryLoop
		}},
		{"delete never re-queries (pre-fix unguarded loop)", false, func(t *testing.T, c, d *rpcWorkflow) {
			rdcStepByName(c, "Delete stale drafts").Run = rdcUnguardedDeleteLoop
		}},
		{"fatal delete of a draft that is already gone", true, func(t *testing.T, c, d *rpcWorkflow) {
			s := rdcStepByName(c, "Delete stale drafts")
			run, ok := rdcReplaceLast(s.Run, "grep -qiE 'not found|HTTP 404'", "false")
			if !ok {
				t.Fatalf("control did not apply: the not-found test of the delete is not where the control expects it")
			}
			s.Run = run
		}},
		{"every delete failure forgiven", true, func(t *testing.T, c, d *rpcWorkflow) {
			s := rdcStepByName(c, "Delete stale drafts")
			run, ok := rdcReplaceLast(s.Run, "grep -qiE 'not found|HTTP 404'", "grep -qiE '.'")
			if !ok {
				t.Fatalf("control did not apply: the not-found test of the delete is not where the control expects it")
			}
			s.Run = run
		}},
	}
	for _, m := range mutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if m.needBash && !harness {
				t.Skip("the mock-gh scenarios are not exercised on windows")
			}
			if !baselineClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			c := rpcParseWorkflow(t, rdcCleanupRel, cleanupText)
			d := rpcParseWorkflow(t, rdcDrafterRel, drafterText)
			m.mutate(t, c, d)
			got := check(c, d)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			if !strings.Contains(got[0], "CLEANUP_CONCURRENCY_VIOLATION") {
				t.Fatalf("mutant %q was rejected without the CLEANUP_CONCURRENCY_VIOLATION sentinel: %s", m.name, got[0])
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
	// Positive control for the alias handling (iter5 F-4): a job-level group written
	// through a YAML alias of a value that does NOT collide with the drafter's
	// group is read and accepted, so refusing aliases outright is not what the
	// failure rows above prove.
	t.Run("control/iter5 D1 positive: cleanup JOB group through a YAML alias of a distinct literal stays clean", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
		}
		c := rpcParseWorkflow(t, rdcCleanupRel, cleanupText)
		d := rpcParseWorkflow(t, rdcDrafterRel, drafterText)
		target := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "release-drafter-cleanup"}
		c.Jobs["cleanup"].Concurrency = yaml.Node{Kind: yaml.AliasNode, Alias: target}
		if got := check(c, d); len(got) != 0 {
			t.Fatalf("an aliased job group that does not collide was refused: %s", got[0])
		}
	})
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

// ---- run-script interpolation (sync-audit iter1 F1) ----------------------------
//
// A `${{ ... }}` expression inside a `run:` body is substituted into the script
// TEXT before the shell parses it, so a value an outsider can influence (a draft
// tag name routed through a step output, an event payload field, a dispatch
// input) becomes shell syntax in a job that may hold contents: write. The
// repair is to pass the value through the step's `env:` and reference "$VAR".
//
// The check is an ALLOWLIST (sync-audit iter2 G1: the iter1 denylist matched only
// the lower-case dotted spelling, so bracket indexing, upper case, a toJSON /
// fromJSON / format / join wrapper, the env context and an expression spread over
// several lines all slipped through). Each `${{ ... }}` is matched across lines,
// whitespace-collapsed and lower-cased (Actions contexts are case-insensitive),
// and must then equal one of exactly these, every other form being refused:
//
//   - needs.<id>.result        runner-defined status enum (success/failure/...)
//   - matrix.<key>             workflow-authored constants; TestWorkflowMatrixValuesLiteral
//     holds every matrix value in the four workflows to a literal scalar or list
//     (no ${{ }}, no fromJSON), which is what makes this entry safe
//   - github.event_name        runner-defined event-kind enum
//   - github.run_id / run_number / run_attempt / sha
//     runner-defined numeric ids and a 40-hex commit id
//   - runner.os / runner.arch  runner-defined platform enums
//   - runner.temp              runner-defined scratch directory path
//   - job.status               runner-defined status enum
//
// Deliberately NOT allowed although some are harmless: secrets.* and vars.*
// (the real workflows use secrets only in env:/with:, never in a run body, so
// allowing them would widen the guard for nothing), github.repository,
// github.ref_name, github.actor, github.head_ref, github.event.*, inputs.*,
// steps.*, env.*, needs.<id>.outputs.*, any function call and any bracket index.
//
// An expression is not matched against the allowlist as one fixed string: it is
// read as operands joined by `==` `!=` `&&` `||` `!` and parentheses, with string
// literals, and is accepted when EVERY operand is on the allowlist and nothing
// else (no function call, no bracket index, no other character) appears. So
// `runner.os == 'Windows' && 'error' || 'warn'` is accepted and any form that
// can yield an outsider-influenced value is refused.
//
// Swept surfaces (sync-audit iter3 G1-R2, iter4 B2): the `run:` body; a step's
// `shell:`; `defaults.run.shell` at workflow and job level; EVERY value and key of
// a step's `with:` map at any depth, whatever the input is called, because what an
// action does with an input (azure/powershell `inlineScript`, nick-fields/retry
// `on_retry_command`, github-script `script`, a docker `args`) cannot be read from
// the workflow, so no input name is assumed to be data; and a job's `container:` /
// `services:` blocks (their `env` and `credentials` data maps aside). A `with:`
// value, unlike a run body, may also reference `secrets.<NAME>` (rdcWithOnlyExpr):
// the real tree passes a token there, and a secret is not an outsider-controlled
// string.
//
// NOT swept, because each is data the runner reads rather than text a shell or
// interpreter evaluates: `env:` (the safe channel), `if:` (an expression, never
// substituted into text), `name:`, `working-directory:`, `runs-on:`, `outputs:`,
// `concurrency:`, `permissions:`, `timeout-minutes:`, `continue-on-error:`, and
// `strategy:` (its own guard, TestWorkflowMatrixValuesLiteral). Residual risk: an
// action that evaluates an `env:` value as code, and any surface a future
// workflow-syntax version adds.

const rdcRunInterpSentinel = "RUN_INTERPOLATION_VIOLATION"

var (
	rdcExprRe = regexp.MustCompile(`(?s)\$\{\{(.*?)\}\}`)
	// rdcAllowedExpr is the operand allowlist: one whole context reference.
	rdcAllowedExpr = regexp.MustCompile(`^(needs\.[a-z0-9_-]+\.result|matrix\.[a-z0-9_-]+|github\.(event_name|run_id|run_number|run_attempt|sha)|runner\.(os|arch|temp)|job\.status)$`)
	// rdcWithOnlyExpr are the operands a `with:` value may add: a secret is
	// configured, not outsider-controlled, and the real tree passes a token there.
	rdcWithOnlyExpr = regexp.MustCompile(`^secrets\.[a-z0-9_]+$`)
	// rdcRunInterpWorkflows are the workflows SPEC-RELEASE-PIPELINE-001 changed.
	rdcRunInterpWorkflows = []string{
		".github/workflows/ci.yml",
		".github/workflows/release-drafter-cleanup.yml",
		".github/workflows/release-drafter.yml",
		".github/workflows/release-pr-multi-os.yml",
	}
)

// rdcSurface is one piece of text a workflow hands to a shell, a script engine or
// an action; withInput marks a `with:` value (which may also name a secret).
type rdcSurface struct {
	label, text string
	withInput   bool
}

// rdcFlatten returns every scalar (map keys included) at or below x as text, so a
// nested `with:` value is swept like a flat one.
func rdcFlatten(x any) []string {
	switch val := x.(type) {
	case nil:
		return nil
	case string:
		return []string{val}
	case map[string]any:
		var out []string
		for _, k := range rpcSortedKeys(val) {
			out = append(out, k)
			out = append(out, rdcFlatten(val[k])...)
		}
		return out
	case []any:
		var out []string
		for _, e := range val {
			out = append(out, rdcFlatten(e)...)
		}
		return out
	default:
		return []string{fmt.Sprint(val)}
	}
}

// rdcNodeScalars returns the scalar text under a yaml node. YAML aliases are
// followed to the node they stand for, so a value written through `*anchor` is
// swept as the text the runner sees. The data maps named in skip are skipped only
// where a container property can sit: when the mapping nesting depth equals
// skipDepth (1 = the keys of the node's own mapping, as in `container:`; 2 = the
// keys of each service in `services:`). A key of the same name anywhere else, a
// service called `env` among them, is swept like any other.
func rdcNodeScalars(n *yaml.Node, skipDepth int, skip ...string) []string {
	var out []string
	var walk func(x *yaml.Node, depth, hops int)
	walk = func(x *yaml.Node, depth, hops int) {
		if hops > 64 { // a self-referencing alias chain: stop rather than loop
			return
		}
		switch x.Kind {
		case yaml.AliasNode:
			if x.Alias != nil {
				walk(x.Alias, depth, hops+1)
			}
		case yaml.ScalarNode:
			out = append(out, x.Value)
		case yaml.MappingNode:
			for i := 0; i+1 < len(x.Content); i += 2 {
				if depth+1 == skipDepth && rpcContains(skip, rpcResolveAlias(x.Content[i]).Value) {
					continue
				}
				walk(x.Content[i], depth+1, hops)
				walk(x.Content[i+1], depth+1, hops)
			}
		default:
			for _, c := range x.Content {
				walk(c, depth, hops)
			}
		}
	}
	walk(n, 0, 0)
	return out
}

// rdcStepSurfaces lists the code-bearing text of a step: its run body, its shell
// and every value of its `with:` map, whatever the input is called.
func rdcStepSurfaces(s rpcStep) []rdcSurface {
	var out []rdcSurface
	if s.Run != "" {
		out = append(out, rdcSurface{label: "run script", text: s.Run})
	}
	if s.Shell != "" {
		out = append(out, rdcSurface{label: "shell", text: s.Shell})
	}
	for _, k := range rpcSortedKeys(s.With) {
		for _, txt := range rdcFlatten(map[string]any{k: s.With[k]}) {
			out = append(out, rdcSurface{label: fmt.Sprintf("with.%s input (%s may evaluate it as code)", k, s.Uses), text: txt, withInput: true})
		}
	}
	return out
}

// rdcJobSurfaces lists the code-bearing text a job carries outside its steps.
func rdcJobSurfaces(j *rpcJob) []rdcSurface {
	var out []rdcSurface
	if j.Defaults.Run.Shell != "" {
		out = append(out, rdcSurface{label: "defaults.run.shell", text: j.Defaults.Run.Shell})
	}
	for _, blk := range []struct {
		name      string
		node      *yaml.Node
		skipDepth int
	}{{"container", &j.Container, 1}, {"services", &j.Services, 2}} {
		for _, txt := range rdcNodeScalars(blk.node, blk.skipDepth, "env", "credentials") {
			out = append(out, rdcSurface{label: blk.name + " block", text: txt})
		}
	}
	// A job that calls a reusable workflow hands its job-level `with:` inputs and
	// `secrets:` to whatever that workflow does with them, which cannot be read
	// from here, so every value and key is swept like a step's `with:`.
	if j.Uses != "" {
		for _, blk := range []struct {
			name string
			node *yaml.Node
		}{{"with", &j.With}, {"secrets", &j.Secrets}} {
			for _, txt := range rdcNodeScalars(blk.node, 0) {
				out = append(out, rdcSurface{label: fmt.Sprintf("job-level %s input (%s may evaluate it as code)", blk.name, j.Uses), text: txt, withInput: true})
			}
		}
	}
	return out
}

// rdcExprRefusal reads one `${{ }}` body as operands joined by == != && || ! and
// parentheses, with string and number literals, and returns why it is refused ("" =
// accepted). Every operand must be on the allowlist; a function call, a bracket
// index or any other character is refused.
func rdcExprRefusal(expr string, withInput bool) string {
	rs := []rune(expr)
	for i := 0; i < len(rs); {
		c := rs[i]
		switch {
		case c == ' ':
			i++
		case c == '(' || c == ')':
			i++
		case c == '\'':
			i++
			for i < len(rs) {
				if rs[i] == '\'' {
					if i+1 < len(rs) && rs[i+1] == '\'' {
						i += 2
						continue
					}
					break
				}
				i++
			}
			if i >= len(rs) {
				return "an unterminated string literal"
			}
			i++
		case (c == '=' || c == '!') && i+1 < len(rs) && rs[i+1] == '=':
			i += 2
		case (c == '&' || c == '|') && i+1 < len(rs) && rs[i+1] == c:
			i += 2
		case c == '!':
			i++
		case c >= '0' && c <= '9':
			for i < len(rs) && rs[i] >= '0' && rs[i] <= '9' {
				i++
			}
		case c >= 'a' && c <= 'z' || c == '_':
			j := i
			for j < len(rs) && (rs[j] >= 'a' && rs[j] <= 'z' || rs[j] >= '0' && rs[j] <= '9' || rs[j] == '_' || rs[j] == '.' || rs[j] == '-') {
				j++
			}
			id := string(rs[i:j])
			k := j
			for k < len(rs) && rs[k] == ' ' {
				k++
			}
			switch {
			case k < len(rs) && rs[k] == '(':
				return fmt.Sprintf("the function call %s(...)", id)
			case id == "true" || id == "false" || id == "null":
			case rdcAllowedExpr.MatchString(id):
			case withInput && rdcWithOnlyExpr.MatchString(id):
			default:
				return fmt.Sprintf("the context reference %s", id)
			}
			i = j
		default:
			return fmt.Sprintf("the character %q", c)
		}
	}
	return ""
}

// rdcCheckRunInterpolation returns one violation per refused expression found
// in any swept surface (see the surface list above).
func rdcCheckRunInterpolation(wfs []*rpcWorkflow) []string {
	var v []string
	sweep := func(wf *rpcWorkflow, where string, sf rdcSurface) {
		matches := rdcExprRe.FindAllStringSubmatch(sf.text, -1)
		if strings.Count(sf.text, "${{") != len(matches) {
			v = append(v, fmt.Sprintf("%s: %s: %s has a ${{ with no closing }} in its %s; the text cannot be read as expressions", rdcRunInterpSentinel, wf.Path, where, sf.label))
		}
		for _, m := range matches {
			expr := strings.ToLower(strings.Join(strings.Fields(m[1]), " "))
			if why := rdcExprRefusal(expr, sf.withInput); why != "" {
				v = append(v, fmt.Sprintf("%s: %s: %s interpolates ${{ %s }} into its %s: %s is not on the allowlist; pass the value through the step env and reference a quoted shell variable (process.env in a script input)", rdcRunInterpSentinel, wf.Path, where, expr, sf.label, why))
			}
		}
	}
	for _, wf := range wfs {
		if wf.Defaults.Run.Shell != "" {
			sweep(wf, "the workflow", rdcSurface{label: "defaults.run.shell", text: wf.Defaults.Run.Shell})
		}
		for _, id := range wf.jobIDs() {
			j := wf.Jobs[id]
			for _, sf := range rdcJobSurfaces(j) {
				sweep(wf, fmt.Sprintf("job %q", id), sf)
			}
			for _, s := range j.Steps {
				for _, sf := range rdcStepSurfaces(s) {
					sweep(wf, fmt.Sprintf("job %q step %q", id, s.Name), sf)
				}
			}
		}
	}
	return v
}

// ---- matrix values are literals (sync-audit iter3 G1-R1) -----------------------
//
// The allowlist above admits `matrix.<key>` because a matrix value is authored in
// the workflow. That holds only while every value IS authored there: a leg value
// set from `${{ github.event.pull_request.title }}` or a whole matrix built with
// `fromJSON(needs.x.outputs.m)` smuggles an outsider-controlled string through an
// allowlisted spelling (mutant M20: it passes the allowlist guard and actionlint).
// So every matrix value, key and nested include / exclude entry in the four
// workflows must be a literal scalar or a literal list / mapping of literals.

const rdcMatrixSentinel = "MATRIX_LITERAL_VIOLATION"

var rdcMatrixExprRe = regexp.MustCompile(`(?i)\$\{\{|fromjson\s*\(`)

// rdcWalkMatrix appends a violation for each expression-bearing string at or
// below v (keys included).
func rdcWalkMatrix(v *[]string, where, path string, x any) {
	bad := func() {
		*v = append(*v, fmt.Sprintf("%s: %s: %s carries an expression or fromJSON (%q); a matrix value must be a literal", rdcMatrixSentinel, where, path, x))
	}
	switch val := x.(type) {
	case string:
		if rdcMatrixExprRe.MatchString(val) {
			bad()
		}
	case map[string]any:
		for _, k := range rpcSortedKeys(val) {
			if rdcMatrixExprRe.MatchString(k) {
				*v = append(*v, fmt.Sprintf("%s: %s: %s has the key %q, which carries an expression", rdcMatrixSentinel, where, path, k))
			}
			rdcWalkMatrix(v, where, path+"."+k, val[k])
		}
	case []any:
		for i, e := range val {
			rdcWalkMatrix(v, where, fmt.Sprintf("%s[%d]", path, i), e)
		}
	case nil, bool, int, int64, uint64, float64:
	default:
		if rdcMatrixExprRe.MatchString(fmt.Sprint(val)) {
			bad()
		}
	}
}

// rdcCheckMatrixLiteral returns one violation per non-literal matrix value, and
// one per matrix written as a scalar expression instead of a mapping.
func rdcCheckMatrixLiteral(wfs []*rpcWorkflow) []string {
	var v []string
	for _, wf := range wfs {
		for _, id := range wf.jobIDs() {
			m := wf.Jobs[id].Strategy.Matrix
			where := fmt.Sprintf("%s job %q strategy.matrix", wf.Path, id)
			for _, k := range rpcSortedKeys(m) {
				if k == rpcMatrixScalarKey {
					v = append(v, fmt.Sprintf("%s: %s is the scalar %q, not a mapping of literals", rdcMatrixSentinel, where, m[k]))
					continue
				}
				rdcWalkMatrix(&v, where, k, m[k])
			}
		}
	}
	return v
}

func TestWorkflowRunScriptsNoUntrustedInterpolation(t *testing.T) {
	texts := map[string]string{}
	load := func() []*rpcWorkflow {
		var wfs []*rpcWorkflow
		for _, rel := range rdcRunInterpWorkflows {
			if _, ok := texts[rel]; !ok {
				texts[rel] = rpcReadRepoFile(t, rel)
			}
			wfs = append(wfs, rpcParseWorkflow(t, rel, texts[rel]))
		}
		return wfs
	}
	base := load()
	swept := 0
	for _, wf := range base {
		for _, j := range wf.Jobs {
			for _, s := range j.Steps {
				if s.Run != "" {
					swept++
				}
			}
		}
	}
	if swept == 0 {
		t.Fatalf("%s: the sweep selected no run steps, so a green would assert nothing", rdcRunInterpSentinel)
	}
	t.Logf("swept %d run steps across %d workflows", swept, len(base))
	for _, msg := range rdcCheckRunInterpolation(base) {
		t.Error(msg)
	}
	baselineClean := len(rdcCheckRunInterpolation(base)) == 0

	// step names the run step to mutate; empty selects the first run step.
	controls := []struct{ name, wf, step, expr string }{
		{"drafts stale_tags step output (the audited sink)", ".github/workflows/release-drafter-cleanup.yml", "Show results", "steps.drafts.outputs.stale_tags"},
		{"step output with a hyphenated id", ".github/workflows/release-drafter-cleanup.yml", "", "steps.set-env.outputs.dry_run"},
		{"event payload field", ".github/workflows/ci.yml", "", "github.event.pull_request.title"},
		{"pull request head ref", ".github/workflows/release-pr-multi-os.yml", "", "github.head_ref"},
		{"needs output", ".github/workflows/release-pr-multi-os.yml", "Verify all OS legs passed", "needs.detect-release.outputs.go_code"},
		{"dispatch input", ".github/workflows/release-drafter-cleanup.yml", "", "inputs.dry_run"},
		{"expression inside a function call", ".github/workflows/ci.yml", "", "format('{0}', github.event.head_commit.message)"},
		// sync-audit iter2 G1: spellings the denylist did not recognise.
		{"bracket-index step output", ".github/workflows/release-drafter-cleanup.yml", "Show results", "steps['drafts'].outputs.stale_tags"},
		{"upper-case step output", ".github/workflows/release-drafter-cleanup.yml", "Show results", "STEPS.drafts.outputs.stale_tags"},
		{"toJSON of the whole event", ".github/workflows/release-drafter-cleanup.yml", "Show results", "toJSON(github.event)"},
		{"env context value", ".github/workflows/release-drafter-cleanup.yml", "Show results", "env.STALE_TAGS"},
		{"fromJSON(toJSON(steps)) round trip", ".github/workflows/release-drafter-cleanup.yml", "Show results", "fromJSON(toJSON(steps)).drafts.outputs.stale_tags"},
		{"expression spread over several lines", ".github/workflows/ci.yml", "", "format('{0}',\n  github.event.head_commit.message)"},
		{"multi-line bare payload field", ".github/workflows/ci.yml", "", "\n  github.event.pull_request.title\n"},
		{"attacker-influenceable ref name", ".github/workflows/release-pr-multi-os.yml", "", "github.ref_name"},
		{"attacker-influenceable actor", ".github/workflows/release-pr-multi-os.yml", "", "github.actor"},
		{"secrets context value", ".github/workflows/ci.yml", "", "secrets.CODECOV_TOKEN"},
		{"join wrapper over a matrix constant", ".github/workflows/ci.yml", "", "join(matrix.goos, ',')"},
		{"needs output via bracket index", ".github/workflows/release-pr-multi-os.yml", "Verify all OS legs passed", "needs['detect-release'].outputs.go_code"},
	}
	for _, c := range controls {
		t.Run("control/"+c.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			wfs := load()
			injected := false
			for _, wf := range wfs {
				if wf.Path != c.wf {
					continue
				}
				for _, id := range wf.jobIDs() {
					j := wf.Jobs[id]
					for i := range j.Steps {
						if j.Steps[i].Run != "" && !injected && (c.step == "" || j.Steps[i].Name == c.step) {
							j.Steps[i].Run += "\necho \"${{ " + c.expr + " }}\"\n"
							injected = true
						}
					}
				}
			}
			if !injected {
				t.Fatalf("control did not apply: %s has no matching run step (step %q)", c.wf, c.step)
			}
			got := rdcCheckRunInterpolation(wfs)
			if len(got) == 0 {
				t.Fatalf("mutant interpolating ${{ %s }} passes the guard", c.expr)
			}
			t.Logf("control observed red: %s", got[0])
		})
	}
	t.Run("control/every allowlisted form stays allowed", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
		}
		wfs := load()
		injected := false
		for _, id := range wfs[0].jobIDs() {
			j := wfs[0].Jobs[id]
			if len(j.Steps) > 0 && j.Steps[0].Run != "" {
				j.Steps[0].Run += "\necho \"${{ needs.detect.result }} ${{ needs.detect-release.result }} ${{ matrix.goos }} ${{ github.event_name }}\"\n" +
					"echo \"${{ github.run_id }} ${{ github.run_number }} ${{ github.run_attempt }} ${{ github.sha }}\"\n" +
					"echo \"${{ runner.os }} ${{ runner.arch }} ${{ job.status }}\"\n" +
					"echo \"${{ NEEDS.detect.RESULT }} ${{\n   matrix.goarch\n}}\"\n"
				injected = true
				break
			}
		}
		if !injected {
			t.Fatalf("control did not apply: no run step to extend")
		}
		if got := rdcCheckRunInterpolation(wfs); len(got) != 0 {
			t.Fatalf("a runner status enum / workflow constant was refused: %s", got[0])
		}
	})

	// G1-R2: code-valued action inputs are swept like run bodies. Each control
	// appends a `uses:` step to the first job of the named workflow.
	withControls := []struct {
		name, wf, uses string
		with           map[string]any
		wantRed        bool
	}{
		{"github-script script with a payload field (M15)", ".github/workflows/ci.yml", "actions/github-script@v7",
			map[string]any{"script": "const t = '${{ github.event.pull_request.title }}';\ncore.info(t);\n"}, true},
		{"github-script script with a step output", ".github/workflows/release-drafter-cleanup.yml", "actions/github-script@v7",
			map[string]any{"script": "core.info('${{ steps.drafts.outputs.stale_tags }}')"}, true},
		{"github-script script, expression spread over lines", ".github/workflows/release-pr-multi-os.yml", "actions/github-script@v7",
			map[string]any{"script": "core.info(`${{\n  github.head_ref\n}}`)"}, true},
		{"github-script input name in another case", ".github/workflows/ci.yml", "actions/github-script@v7",
			map[string]any{"Script": "core.info('${{ github.event.head_commit.message }}')"}, true},
		{"command input of another action", ".github/workflows/ci.yml", "example/shell-action@v1",
			map[string]any{"command": "echo ${{ github.head_ref }}"}, true},
		{"args input of another action", ".github/workflows/ci.yml", "example/runner-action@v1",
			map[string]any{"args": "--title ${{ github.event.pull_request.title }}"}, true},
		{"github-script script using only allowlisted forms", ".github/workflows/ci.yml", "actions/github-script@v7",
			map[string]any{"script": "core.info('${{ github.event_name }} ${{ needs.detect.result }} ${{ matrix.goos }}')"}, false},
		// iter4 B2: the sweep is a rule over every `with:` value, not a list of names.
		{"azure/powershell inlineScript (N1)", ".github/workflows/ci.yml", "azure/powershell@v2",
			map[string]any{"inlineScript": "Write-Output \"${{ github.event.pull_request.title }}\""}, true},
		{"nick-fields/retry on_retry_command (N17)", ".github/workflows/ci.yml", "nick-fields/retry@v3",
			map[string]any{"command": "true", "on_retry_command": "echo '${{ github.event.pull_request.title }}'"}, true},
		{"nick-fields/retry new_command_on_retry", ".github/workflows/ci.yml", "nick-fields/retry@v3",
			map[string]any{"command": "true", "new_command_on_retry": "echo ${{ github.head_ref }}"}, true},
		{"an input name no list has ever heard of", ".github/workflows/ci.yml", "example/anything@v1",
			map[string]any{"zz_unlisted_input": "${{ github.head_ref }}"}, true},
		{"expression nested in a map value", ".github/workflows/ci.yml", "example/anything@v1",
			map[string]any{"config": map[string]any{"hook": "${{ github.head_ref }}"}}, true},
		{"expression nested in a list value", ".github/workflows/ci.yml", "example/anything@v1",
			map[string]any{"steps": []any{"ok", "${{ github.head_ref }}"}}, true},
		{"expression with no closing braces", ".github/workflows/ci.yml", "example/anything@v1",
			map[string]any{"script": "echo ${{ github.event_name "}, true},
		{"a with value built with a function call", ".github/workflows/ci.yml", "example/anything@v1",
			map[string]any{"x": "${{ toJSON(matrix) }}"}, true},
		{"secret in a with value stays allowed (the real tree passes a token)", ".github/workflows/ci.yml", "codecov/codecov-action@v7",
			map[string]any{"token": "${{ secrets.CODECOV_TOKEN }}"}, false},
		{"runner enum comparison with literals stays allowed (the real tree's if-no-files-found)", ".github/workflows/ci.yml", "actions/upload-artifact@v7",
			map[string]any{"if-no-files-found": "${{ runner.os == 'Windows' && 'error' || 'warn' }}", "path": "${{ runner.temp }}/x"}, false},
		{"comparison that falls back to an outsider value", ".github/workflows/ci.yml", "actions/upload-artifact@v7",
			map[string]any{"name": "${{ runner.os == 'Windows' && 'error' || github.head_ref }}"}, true},
	}
	for _, c := range withControls {
		t.Run("control/with/"+c.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			wfs := load()
			injected := false
			for _, wf := range wfs {
				if wf.Path != c.wf {
					continue
				}
				id := wf.jobIDs()[0]
				wf.Jobs[id].Steps = append(wf.Jobs[id].Steps, rpcStep{Name: "Injected", Uses: c.uses, With: c.with})
				injected = true
			}
			if !injected {
				t.Fatalf("control did not apply: %s is not in the swept set", c.wf)
			}
			got := rdcCheckRunInterpolation(wfs)
			if !c.wantRed {
				if len(got) != 0 {
					t.Fatalf("an allowlisted form in a script input was refused: %s", got[0])
				}
				return
			}
			if len(got) == 0 {
				t.Fatalf("mutant %v with %v passes the guard", c.uses, c.with)
			}
			t.Logf("control observed red: %s", got[0])
		})
	}

	// iter4 B2: the shell a step runs under, and the blocks a job carries outside
	// its steps, are code surfaces too. Each control edits the first job of ci.yml.
	nodeOf := func(t *testing.T, text string) yaml.Node {
		t.Helper()
		var n yaml.Node
		if err := yaml.Unmarshal([]byte(text), &n); err != nil {
			t.Fatalf("control text does not parse: %v", err)
		}
		return n
	}
	surfaceControls := []struct {
		name    string
		mutate  func(t *testing.T, wf *rpcWorkflow, j *rpcJob)
		wantRed bool
	}{
		{"job defaults.run.shell carrying head_ref (N2)", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Defaults.Run.Shell = `bash -c "echo ${{ github.head_ref }}; bash -e {0}"`
		}, true},
		{"workflow defaults.run.shell carrying head_ref", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			wf.Defaults.Run.Shell = `bash -c "echo ${{ github.head_ref }}; bash -e {0}"`
		}, true},
		{"step shell carrying head_ref", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Steps = append(j.Steps, rpcStep{Name: "Injected", Run: "echo ok", Shell: `bash -c "echo ${{ github.head_ref }}; bash -e {0}"`})
		}, true},
		{"container options carrying head_ref", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Container = nodeOf(t, "image: alpine:3.20\noptions: --hostname ${{ github.head_ref }}\n")
		}, true},
		{"service options carrying a payload field", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Services = nodeOf(t, "db:\n  image: postgres:16\n  options: --health-cmd ${{ github.event.pull_request.title }}\n")
		}, true},
		{"secret in a run body stays refused (a with value may name one, a script may not)", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Steps = append(j.Steps, rpcStep{Name: "Injected", Run: "echo ${{ secrets.CODECOV_TOKEN }}"})
		}, true},
		// sync-audit iter5 F-3. Failure rows hide the expression where the iter4 sweep
		// did not look; positive rows put an expression only where it is data.
		{"iter5 B2a failure: container options through a YAML alias whose target sits outside the block", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			target := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "--hostname ${{ github.head_ref }}"}
			j.Container = yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "image"}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: "alpine:3.20"},
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "options"}, {Kind: yaml.AliasNode, Alias: target},
			}}
		}, true},
		{"iter5 B2a positive: container options through a YAML alias of a plain value", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			target := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "--cpus 1"}
			j.Container = yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "image"}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: "alpine:3.20"},
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "options"}, {Kind: yaml.AliasNode, Alias: target},
			}}
		}, false},
		{"iter5 B2b failure: a service NAMED env carrying an expression in its options", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Services = nodeOf(t, "env:\n  image: alpine:3.20\n  options: --health-cmd ${{ github.head_ref }}\n")
		}, true},
		{"iter5 B2b positive: a service's own env and credentials maps stay data", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Services = nodeOf(t, "db:\n  image: postgres:16\n  env:\n    X: ${{ github.head_ref }}\n  credentials:\n    password: ${{ secrets.X }}\n")
		}, false},
		{"iter5 B2c failure: job-level with: of a reusable-workflow call carrying a payload field", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Uses = "octo-org/tools/.github/workflows/run.yml@v1"
			j.With = nodeOf(t, "cmd: echo ${{ github.event.pull_request.title }}\n")
		}, true},
		{"iter5 B2c failure: job-level secrets: of a reusable-workflow call carrying head_ref", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Uses = "octo-org/tools/.github/workflows/run.yml@v1"
			j.Secrets = nodeOf(t, "token: ${{ github.head_ref }}\n")
		}, true},
		{"iter5 B2c positive: job-level with: of allowlisted forms and a secret, secrets: inherit", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			j.Uses = "octo-org/tools/.github/workflows/run.yml@v1"
			j.With = nodeOf(t, "cmd: ${{ github.event_name }}\ntoken: ${{ secrets.X }}\n")
			j.Secrets = nodeOf(t, "inherit\n")
		}, false},
		{"literal shell and a service without expressions stay allowed", func(t *testing.T, wf *rpcWorkflow, j *rpcJob) {
			wf.Defaults.Run.Shell = "bash"
			j.Defaults.Run.Shell = "bash"
			j.Container = nodeOf(t, "image: alpine:3.20\noptions: --cpus 1\nenv:\n  X: ${{ github.head_ref }}\ncredentials:\n  password: ${{ secrets.X }}\n")
			j.Steps = append(j.Steps, rpcStep{Name: "Injected", Run: "echo ok", Shell: "bash"})
		}, false},
	}
	for _, c := range surfaceControls {
		t.Run("control/surface/"+c.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			wfs := load()
			wf := wfs[0]
			c.mutate(t, wf, wf.Jobs[wf.jobIDs()[0]])
			got := rdcCheckRunInterpolation(wfs)
			if !c.wantRed {
				if len(got) != 0 {
					t.Fatalf("an allowed surface was refused: %s", got[0])
				}
				return
			}
			if len(got) > 0 && !strings.Contains(got[0], rdcRunInterpSentinel) {
				t.Fatalf("mutant %q was rejected without the %s sentinel: %s", c.name, rdcRunInterpSentinel, got[0])
			}
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", c.name)
			}
			t.Logf("control observed red: %s", got[0])
		})
	}
}

// TestWorkflowMatrixValuesLiteral closes the trust the interpolation allowlist
// places in `matrix.<key>` (sync-audit iter3 G1-R1): every strategy.matrix value
// in the four workflows is a literal, with no expression and no fromJSON.
func TestWorkflowMatrixValuesLiteral(t *testing.T) {
	load := func() []*rpcWorkflow {
		var wfs []*rpcWorkflow
		for _, rel := range rdcRunInterpWorkflows {
			wfs = append(wfs, rpcLoadWorkflow(t, rel))
		}
		return wfs
	}
	base := load()
	matrices := 0
	for _, wf := range base {
		for _, j := range wf.Jobs {
			if len(j.Strategy.Matrix) > 0 {
				matrices++
			}
		}
	}
	if matrices == 0 {
		t.Fatalf("%s: the sweep selected no strategy.matrix, so a green would assert nothing", rdcMatrixSentinel)
	}
	t.Logf("swept %d matrices across %d workflows", matrices, len(base))
	for _, msg := range rdcCheckMatrixLiteral(base) {
		t.Error(msg)
	}
	baselineClean := len(rdcCheckMatrixLiteral(base)) == 0

	// include returns the idx-th include entry of a job's matrix.
	include := func(w *rpcWorkflow, job string, idx int) map[string]any {
		return w.Jobs[job].Strategy.Matrix["include"].([]any)[idx].(map[string]any)
	}
	byPath := func(wfs []*rpcWorkflow, rel string) *rpcWorkflow {
		for _, w := range wfs {
			if w.Path == rel {
				return w
			}
		}
		t.Fatalf("control did not apply: %s is not in the swept set", rel)
		return nil
	}
	const ciRel = ".github/workflows/ci.yml"
	const miniHead = "jobs:\n  m:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo ok\n    strategy:\n"
	controls := []struct {
		name   string
		mutate func(wfs []*rpcWorkflow) []*rpcWorkflow
	}{
		{"include entry set from the pull request title (M20)", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			include(byPath(wfs, ciRel), "build", 0)["goos"] = "${{ github.event.pull_request.title }}"
			return wfs
		}},
		{"axis list element carrying an expression", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			byPath(wfs, ciRel).Jobs["test"].Strategy.Matrix["os"] = []any{"ubuntu-latest", "${{ github.head_ref }}"}
			return wfs
		}},
		{"exclude entry carrying an expression", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			byPath(wfs, ciRel).Jobs["build"].Strategy.Matrix["exclude"] = []any{map[string]any{"goos": "${{ github.head_ref }}"}}
			return wfs
		}},
		{"axis built with fromJSON", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			byPath(wfs, ciRel).Jobs["test"].Strategy.Matrix["shard"] = "${{ fromJSON(needs.detect.outputs.shards) }}"
			return wfs
		}},
		{"include value nested in a list", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			include(byPath(wfs, ciRel), "build", 1)["tags"] = []any{"a", map[string]any{"x": "${{ github.event.pull_request.title }}"}}
			return wfs
		}},
		{"matrix key carrying an expression", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			include(byPath(wfs, ciRel), "build", 0)["${{ github.head_ref }}"] = "x"
			return wfs
		}},
		{"fromJSON call without braces", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			include(byPath(wfs, ciRel), "build", 0)["goarch"] = "fromJSON('[\"amd64\"]')"
			return wfs
		}},
		{"whole matrix written as a fromJSON scalar", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			return append(wfs, rpcParseWorkflow(t, "mini.yml", miniHead+"      matrix: ${{ fromJSON(needs.a.outputs.m) }}\n"))
		}},
		{"whole matrix written as an include scalar", func(wfs []*rpcWorkflow) []*rpcWorkflow {
			return append(wfs, rpcParseWorkflow(t, "mini.yml", miniHead+"      matrix:\n        include: ${{ fromJSON(needs.a.outputs.m) }}\n"))
		}},
	}
	for _, c := range controls {
		t.Run("control/"+c.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			got := rdcCheckMatrixLiteral(c.mutate(load()))
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", c.name)
			}
			t.Logf("control observed red: %s", got[0])
		})
	}
	t.Run("control/literal forms stay allowed", func(t *testing.T) {
		mini := rpcParseWorkflow(t, "mini.yml", miniHead+
			"      fail-fast: false\n      matrix:\n        os: [ubuntu-latest, macos-latest]\n        shard: [1, 2]\n        flag: [true, false]\n"+
			"        include:\n          - os: ubuntu-latest\n            tags: [a, b]\n            extra: ~\n"+
			"        exclude:\n          - os: macos-latest\n            shard: 2\n")
		if got := rdcCheckMatrixLiteral([]*rpcWorkflow{mini}); len(got) != 0 {
			t.Fatalf("a literal matrix was refused: %s", got[0])
		}
	})
}
