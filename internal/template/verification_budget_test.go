// verification_budget_test.go: guards for SPEC-RELEASE-PIPELINE-001 M4 (card
// t1536, REQ-RPC-007 / REQ-RPC-008 / REQ-RPC-009): stage budgets, the browser
// phase split, and retry evidence.
//
//   - TestStageCeilingBelowJobBudget (AC-RPC-007): every long stage carries a
//     ceiling strictly below its job's, and the evidence steps that follow it
//     keep their share of the budget. Browser half: both phases of the
//     test-browser job. Release half: the race step of the release matrix, on
//     the measured budget model (acceptance.md EV-15), not on the per-binary
//     `-timeout 35m`.
//   - TestBrowserPhaseSplit (AC-RPC-008): the green and mutation phases are
//     separately bounded, a maximum-duration green phase leaves the mutation
//     phase its declared budget, and the SPEC-APPJS-FIRE-GUARD-001 invariants
//     (0 -> 1 -> 0 order, restore assertion, pinned Chrome) are unchanged.
//   - TestRetryPolicyPreservesFirstFailure (AC-RPC-009): retries go through the
//     registry wrapper only, the wrapper preserves the first failure before the
//     re-run, and every retry-wrapped job uploads that record.
//
// Budget numbers are test constants citing their measurement, never read back
// from the workflow comment that declares them (a mutant that edits its own
// comment cannot widen the bound). Sentinels on failure:
// STAGE_BUDGET_VIOLATION, BROWSER_SPLIT_VIOLATION, RETRY_EVIDENCE_VIOLATION.
package template_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const retryWrapperRel = "scripts/ci/retry-flaky.sh"

// Release leg budget model (acceptance.md EV-15: failing run 39729541531,
// n=1, windows, non-main branch; a FAILING-run measurement that bounds nothing
// about green legs, so the test asserts no lower bound for a ceiling).
const (
	vbReleaseJobBudget     = 40 // job timeout-minutes, not raised (SPEC-CI-RELEASE-VERIFY-SHARD-001)
	vbReleasePreStepMin    = 5  // pre-step allowance: 4m21s measured (cli-2, go vet), rounded up
	vbReleaseEvidenceMin   = 1  // evidence allowance: 6-8s measured, one minute held
	vbReleasePerBinaryFlag = "35m"
)

// Browser job budget model. Measured on four green main runs (38991715666,
// 38697937293, 38673824546, 37861953651; `gh run view --json jobs`, 2026-10-23):
// green phase 13m05s / 12m17s / 13m43s / 12m44s, mutation phase 1m11s / 1m09s /
// 1m16s / 59s, steps before the green phase about 45s. The allowances below keep
// a margin over those maxima; the browser job's own timeout-minutes stays 20.
const (
	vbBrowserJobBudget   = 20
	vbBrowserSetupMin    = 1 // checkout, Go, Chrome provisioning, manifest lint (about 45s measured)
	vbBrowserEvidenceMin = 1 // the always() evidence upload and post steps after the mutation phase
)

func vbNodeInt(n yaml.Node) (int, bool) {
	if n.Kind != yaml.ScalarNode {
		return 0, false
	}
	v, err := strconv.Atoi(n.Value)
	return v, err == nil
}

func vbStepIndex(j *rpcJob, prefix string) int {
	for i, s := range j.Steps {
		if strings.HasPrefix(s.Name, prefix) {
			return i
		}
	}
	return -1
}

var vbGoTimeoutRe = regexp.MustCompile(`-timeout\s+(\d+)m\b`)

// ---- AC-RPC-007 --------------------------------------------------------------

func vbCheckBrowserBudget(ci *rpcWorkflow) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "STAGE_BUDGET_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	j := ci.Jobs["test-browser"]
	if j == nil {
		return []string{"STAGE_BUDGET_VIOLATION: ci.yml has no test-browser job"}
	}
	if got, ok := vbNodeInt(j.TimeoutMinutes); !ok || got != vbBrowserJobBudget {
		add("test-browser timeout-minutes is %v, want the unchanged %d (a stage bound replaces nothing by raising the job)", j.TimeoutMinutes.Value, vbBrowserJobBudget)
	}
	green, red := vbStepIndex(j, "Green phase"), vbStepIndex(j, "Red phase")
	if green < 0 || red < 0 {
		return append(v, "STAGE_BUDGET_VIOLATION: the green or the mutation phase step is missing")
	}
	total := vbBrowserSetupMin + vbBrowserEvidenceMin
	for _, idx := range []int{green, red} {
		s := j.Steps[idx]
		got, ok := vbNodeInt(s.TimeoutMinutes)
		if !ok {
			add("%q has no step-level timeout-minutes", s.Name)
			continue
		}
		if got >= vbBrowserJobBudget {
			add("%q ceiling %d is not strictly below the job's %d", s.Name, got, vbBrowserJobBudget)
		}
		total += got
	}
	if total > vbBrowserJobBudget {
		add("setup %d + green + mutation ceilings + evidence %d = %d exceeds the job budget %d: a stage at its ceiling starves what follows", vbBrowserSetupMin, vbBrowserEvidenceMin, total, vbBrowserJobBudget)
	}
	// Go's own timeout must fire first so a hung driver test leaves a goroutine dump.
	if g, ok := vbNodeInt(j.Steps[green].TimeoutMinutes); ok {
		if m := vbGoTimeoutRe.FindStringSubmatch(j.Steps[green].Run); m != nil {
			if goT, _ := strconv.Atoi(m[1]); goT >= g {
				add("the green phase's go test -timeout %dm is not below its step ceiling %d: the step would be killed before Go can report", goT, g)
			}
		}
	}
	// The evidence upload must outlive a failing or ceiling-hit phase.
	found := false
	for i := red + 1; i < len(j.Steps); i++ {
		s := j.Steps[i]
		if strings.HasPrefix(s.Uses, "actions/upload-artifact@") && strings.Contains(s.If, "always()") {
			found = true
		}
	}
	if !found {
		add("no always() evidence upload follows the mutation phase, so a failing phase leaves no trail")
	}
	return v
}

func vbCheckReleaseBudget(rel *rpcWorkflow) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "STAGE_BUDGET_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	j := rel.Jobs["full-matrix-test"]
	if j == nil {
		return []string{"STAGE_BUDGET_VIOLATION: release-pr-multi-os.yml has no full-matrix-test job"}
	}
	if got, ok := vbNodeInt(j.TimeoutMinutes); !ok || got != vbReleaseJobBudget {
		add("full-matrix-test timeout-minutes is %q, want the unchanged %d", j.TimeoutMinutes.Value, vbReleaseJobBudget)
	}
	race := vbStepIndex(j, "Run tests with race detector")
	compress, upload := vbStepIndex(j, "Compress test event stream"), vbStepIndex(j, "Upload test event stream")
	if race < 0 || compress < 0 || upload < 0 {
		return append(v, "STAGE_BUDGET_VIOLATION: the race, compress or upload step is missing")
	}
	if race >= compress || compress >= upload {
		add("the evidence steps must follow the race step (race %d, compress %d, upload %d)", race, compress, upload)
	}
	for _, i := range []int{compress, upload} {
		if !strings.Contains(j.Steps[i].If, "always()") {
			add("%q is not an `if: always()` step, so a failing or ceiling-hit race step leaves no evidence", j.Steps[i].Name)
		}
	}
	if strings.Contains(j.Steps[race].Run, "test-stream.json.gz") || strings.Contains(j.Steps[race].Run, "gzip") {
		add("the race step folds the evidence collection into itself")
	}
	raceRun := j.Steps[race].Run
	if n := strings.Count(raceRun, "-timeout "+vbReleasePerBinaryFlag); n != 2 {
		add("per-binary `-timeout %s` appears %d times in the race step, want the unchanged 2 (:300/:302)", vbReleasePerBinaryFlag, n)
	}
	if regexp.MustCompile(`-timeout (3[6-9]|[4-9][0-9])m`).MatchString(raceRun) {
		add("a go test -timeout above %s is present: the per-shard ceiling is not raised", vbReleasePerBinaryFlag)
	}
	// If the race step carries a ceiling T, T + pre-step + evidence must fit.
	if t, ok := vbNodeInt(j.Steps[race].TimeoutMinutes); ok {
		if t+vbReleasePreStepMin+vbReleaseEvidenceMin > vbReleaseJobBudget {
			add("race-step ceiling %d + pre-step %d + evidence %d = %d exceeds the job budget %d", t, vbReleasePreStepMin, vbReleaseEvidenceMin, t+vbReleasePreStepMin+vbReleaseEvidenceMin, vbReleaseJobBudget)
		}
	} else if j.Steps[race].TimeoutMinutes.Kind != 0 {
		add("the race step's timeout-minutes is not a plain integer, so the budget cannot be checked")
	}
	return v
}

func TestStageCeilingBelowJobBudget(t *testing.T) {
	ciText := rpcReadRepoFile(t, rpcCIWorkflowRel)
	ci := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
	relText := rpcReadRepoFile(t, rpcReleaseWorkflowRel)
	rel := rpcParseWorkflow(t, rpcReleaseWorkflowRel, relText)
	for _, msg := range vbCheckBrowserBudget(ci) {
		t.Error(msg)
	}
	for _, msg := range vbCheckReleaseBudget(rel) {
		t.Error(msg)
	}

	relClean := len(vbCheckReleaseBudget(rel)) == 0
	browserClean := len(vbCheckBrowserBudget(ci)) == 0
	setTimeout := func(s *rpcStep, minutes int) {
		s.TimeoutMinutes = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(minutes)}
	}
	relMutants := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"race step ceiling 38 (the withdrawn v0.4.0 window)", func(w *rpcWorkflow) {
			j := w.Jobs["full-matrix-test"]
			setTimeout(&j.Steps[vbStepIndex(j, "Run tests with race detector")], 38)
		}},
		{"upload step without if: always()", func(w *rpcWorkflow) {
			j := w.Jobs["full-matrix-test"]
			j.Steps[vbStepIndex(j, "Upload test event stream")].If = ""
		}},
		{"compress step without if: always()", func(w *rpcWorkflow) {
			j := w.Jobs["full-matrix-test"]
			j.Steps[vbStepIndex(j, "Compress test event stream")].If = ""
		}},
		{"per-shard ceiling raised", func(w *rpcWorkflow) {
			j := w.Jobs["full-matrix-test"]
			s := &j.Steps[vbStepIndex(j, "Run tests with race detector")]
			s.Run = strings.ReplaceAll(s.Run, "-timeout 35m", "-timeout 45m")
		}},
		{"job budget raised", func(w *rpcWorkflow) {
			w.Jobs["full-matrix-test"].TimeoutMinutes = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "60"}
		}},
	}
	for _, m := range relMutants {
		t.Run("control/release/"+m.name, func(t *testing.T) {
			if !relClean {
				t.Fatalf("control precondition: the real release workflow is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rpcReleaseWorkflowRel, relText)
			m.mutate(mut)
			got := vbCheckReleaseBudget(mut)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
	browserMutants := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"green phase ceiling equals the job budget", func(w *rpcWorkflow) {
			j := w.Jobs["test-browser"]
			setTimeout(&j.Steps[vbStepIndex(j, "Green phase")], 20)
		}},
		{"mutation phase has no ceiling", func(w *rpcWorkflow) {
			j := w.Jobs["test-browser"]
			j.Steps[vbStepIndex(j, "Red phase")].TimeoutMinutes = yaml.Node{}
		}},
		{"stage ceilings consume the whole job", func(w *rpcWorkflow) {
			j := w.Jobs["test-browser"]
			setTimeout(&j.Steps[vbStepIndex(j, "Green phase")], 19)
			setTimeout(&j.Steps[vbStepIndex(j, "Red phase")], 19)
		}},
		{"no always() evidence upload", func(w *rpcWorkflow) {
			j := w.Jobs["test-browser"]
			for i := range j.Steps {
				if strings.HasPrefix(j.Steps[i].Uses, "actions/upload-artifact@") {
					j.Steps[i].If = ""
				}
			}
		}},
	}
	for _, m := range browserMutants {
		t.Run("control/browser/"+m.name, func(t *testing.T) {
			if !browserClean {
				t.Fatalf("control precondition: the real browser job is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
			m.mutate(mut)
			got := vbCheckBrowserBudget(mut)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
}

// ---- AC-RPC-008 --------------------------------------------------------------

func vbCheckBrowserSplit(ci *rpcWorkflow) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "BROWSER_SPLIT_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	j := ci.Jobs["test-browser"]
	if j == nil {
		return []string{"BROWSER_SPLIT_VIOLATION: ci.yml has no test-browser job"}
	}
	green, red := vbStepIndex(j, "Green phase"), vbStepIndex(j, "Red phase")
	if green < 0 || red < 0 || green >= red {
		return append(v, "BROWSER_SPLIT_VIOLATION: the green phase must be a separate step before the mutation phase")
	}
	g, gok := vbNodeInt(j.Steps[green].TimeoutMinutes)
	m, mok := vbNodeInt(j.Steps[red].TimeoutMinutes)
	if !gok || !mok {
		add("the phases do not both carry independent step-level bounds (green set=%v, mutation set=%v)", gok, mok)
	} else if rest := vbBrowserJobBudget - vbBrowserSetupMin - g - vbBrowserEvidenceMin; rest < m {
		add("a maximum-duration green phase (%d) leaves %d minutes, below the mutation phase's declared budget %d", g, rest, m)
	}
	// SPEC-APPJS-FIRE-GUARD-001: 0 -> 1 -> 0 order and the restore assertion.
	run := j.Steps[red].Run
	last := -1
	for _, marker := range []string{"t1060-ci-green", "appjs_fire_mutation.py", "t1060-ci-red", "RESTORED_BYTE_IDENTICAL", "t1060-ci-post-restore"} {
		i := strings.Index(run, marker)
		if i < 0 {
			add("the mutation phase lost %q", marker)
			continue
		}
		if i < last {
			add("the mutation phase runs %q out of order (0 -> 1 -> 0 broken)", marker)
		}
		last = i
	}
	for _, need := range []string{"git diff --exit-code -- internal/web/assets/app.js", `cmp internal/web/assets/app.js`} {
		if !strings.Contains(run, need) {
			add("the restore assertion %q is gone", need)
		}
	}
	all := ""
	for _, s := range j.Steps {
		all += s.Name + "\n" + s.Run + "\n"
	}
	for _, need := range []string{
		`CFT_SHA256="e66f66d4802a46d4a022667e668aa950e277cadbfbed4b3777915b47413a0ef9"`,
		"sha256sum --check --strict",
		"apparmor_parser -r /etc/apparmor.d/moai-cft-chrome",
		"--lint-manifest",
	} {
		if !strings.Contains(all, need) {
			add("the SPEC-APPJS-FIRE-GUARD-001 invariant %q is gone", need)
		}
	}
	// A failure of the mutation phase keeps its own diagnosable trail.
	trail := false
	for i := red + 1; i < len(j.Steps); i++ {
		s := j.Steps[i]
		if strings.HasPrefix(s.Uses, "actions/upload-artifact@") && strings.Contains(s.If, "always()") {
			p := fmt.Sprint(s.With["path"])
			if strings.Contains(p, "fire-guard") && strings.Contains(p, ".log") && strings.Contains(p, ".json") {
				trail = true
			}
		}
	}
	if !trail {
		add("no always() upload carries the mutation phase's server/chrome logs and red report")
	}
	return v
}

func TestBrowserPhaseSplit(t *testing.T) {
	ciText := rpcReadRepoFile(t, rpcCIWorkflowRel)
	ci := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
	for _, msg := range vbCheckBrowserSplit(ci) {
		t.Error(msg)
	}
	clean := len(vbCheckBrowserSplit(ci)) == 0
	setTimeout := func(s *rpcStep, minutes int) {
		s.TimeoutMinutes = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(minutes)}
	}
	mutants := []struct {
		name   string
		mutate func(j *rpcJob)
	}{
		{"one shared budget (no mutation bound)", func(j *rpcJob) { j.Steps[vbStepIndex(j, "Red phase")].TimeoutMinutes = yaml.Node{} }},
		{"green phase leaves the mutation phase nothing", func(j *rpcJob) {
			setTimeout(&j.Steps[vbStepIndex(j, "Green phase")], 18)
			setTimeout(&j.Steps[vbStepIndex(j, "Red phase")], 2)
		}},
		{"restore assertion removed", func(j *rpcJob) {
			s := &j.Steps[vbStepIndex(j, "Red phase")]
			s.Run = strings.Replace(s.Run, `echo "RESTORED_BYTE_IDENTICAL"`, "true", 1)
		}},
		{"phases swapped", func(j *rpcJob) {
			g, r := vbStepIndex(j, "Green phase"), vbStepIndex(j, "Red phase")
			j.Steps[g], j.Steps[r] = j.Steps[r], j.Steps[g]
		}},
		{"pinned Chrome checksum dropped", func(j *rpcJob) {
			for i := range j.Steps {
				j.Steps[i].Run = strings.ReplaceAll(j.Steps[i].Run, "sha256sum --check --strict", "true")
			}
		}},
		{"mutation evidence upload removed", func(j *rpcJob) {
			var keep []rpcStep
			for _, s := range j.Steps {
				if !strings.HasPrefix(s.Uses, "actions/upload-artifact@") {
					keep = append(keep, s)
				}
			}
			j.Steps = keep
		}},
	}
	for _, m := range mutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if !clean {
				t.Fatalf("control precondition: the real browser job is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
			m.mutate(mut.Jobs["test-browser"])
			got := vbCheckBrowserSplit(mut)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
}

// ---- AC-RPC-009 --------------------------------------------------------------

var (
	vbWrapperInvokeRe = regexp.MustCompile(`bash scripts/ci/retry-flaky\.sh --registry scripts/ci/flaky-registry\.txt -- go test`)
	vbAdhocRetryRe    = regexp.MustCompile(`(?m)(\|\|\s*go test|\b(for|while|until)\b[^\n]*\b(attempt|retry)\b)`)
)

// vbAttempt1Path derives the first-failure record path from the wrapper: the
// retry log default with its .log stem replaced by .attempt1.json.
func vbAttempt1Path(script string) (string, bool) {
	m := regexp.MustCompile(`log_file="\$\{FLAKY_RETRY_LOG:-([^}]+\.log)\}"`).FindStringSubmatch(script)
	if m == nil {
		return "", false
	}
	return strings.TrimSuffix(m[1], ".log") + ".attempt1.json", true
}

func vbCheckRetryScript(script string) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "RETRY_EVIDENCE_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	lines := strings.Split(script, "\n")
	cp, retry := -1, -1
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, `cp "$work/attempt1.json" "${log_file%.log}.attempt1.json"`) && cp < 0 {
			cp = i
		}
		if t == `"$@"` {
			retry = i
		}
	}
	switch {
	case cp < 0:
		add("the wrapper never preserves the first attempt's record")
	case retry < 0:
		add("the wrapper has no re-run line")
	case cp > retry:
		add("the wrapper preserves the first failure (line %d) AFTER the re-run (line %d)", cp+1, retry+1)
	}
	if _, ok := vbAttempt1Path(script); !ok {
		add("the wrapper's retry-log default cannot be read, so the record path is unknown")
	}
	return v
}

func vbCheckRetryWorkflow(ci, rel *rpcWorkflow, attempt1 string) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "RETRY_EVIDENCE_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	wrapped := 0
	for _, id := range ci.jobIDs() {
		j := ci.Jobs[id]
		invocations := 0
		for _, s := range j.Steps {
			invocations += len(vbWrapperInvokeRe.FindAllString(s.Run, -1))
			if strings.Contains(s.Run, "retry-flaky.sh") && !vbWrapperInvokeRe.MatchString(s.Run) {
				add("job %s invokes the retry wrapper outside the registry form", id)
			}
			if strings.Contains(strings.ToLower(s.Uses), "retry") {
				add("job %s retries through a third-party action %q, not the registry wrapper", id, s.Uses)
			}
			if vbAdhocRetryRe.MatchString(s.Run) {
				add("job %s carries an ad-hoc retry (%q), bypassing the registry", id, vbAdhocRetryRe.FindString(s.Run))
			}
			if _, ok := s.Env["FLAKY_RETRY_LOG"]; ok {
				add("job %s redirects the wrapper's log, so the upload path no longer matches the record", id)
			}
		}
		if invocations == 0 {
			continue
		}
		wrapped += invocations
		uploaded := false
		for _, s := range j.Steps {
			if !strings.HasPrefix(s.Uses, "actions/upload-artifact@") {
				continue
			}
			if strings.Contains(fmt.Sprint(s.With["path"]), attempt1) {
				uploaded = true
				if !strings.Contains(s.If, "always()") {
					add("job %s: the attempt1 upload is not if: always()", id)
				}
				if fmt.Sprint(s.With["if-no-files-found"]) != "warn" {
					add("job %s: the attempt1 upload does not warn on absence (a run with no retry has no record)", id)
				}
				if fmt.Sprint(s.With["include-hidden-files"]) != "true" {
					add("job %s: the attempt1 record lives under a hidden directory, so the upload needs include-hidden-files: true", id)
				}
			}
		}
		if !uploaded {
			add("job %s runs the retry wrapper but no upload step carries %s", id, attempt1)
		}
	}
	if wrapped != 7 {
		add("the retry wrapper is invoked %d times in ci.yml, want the seven SPEC-CANDIDATE-CI-001 sites", wrapped)
	}
	for _, id := range rel.jobIDs() {
		for _, s := range rel.Jobs[id].Steps {
			if strings.Contains(s.Run, "retry-flaky") {
				add("release job %s uses the retry wrapper: the release legs stay wrapper-free (decision-index Q6)", id)
			}
		}
	}
	return v
}

func TestRetryPolicyPreservesFirstFailure(t *testing.T) {
	script := rpcReadRepoFile(t, retryWrapperRel)
	ciText := rpcReadRepoFile(t, rpcCIWorkflowRel)
	ci := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
	relText := rpcReadRepoFile(t, rpcReleaseWorkflowRel)
	rel := rpcParseWorkflow(t, rpcReleaseWorkflowRel, relText)
	attempt1, ok := vbAttempt1Path(script)
	if !ok {
		t.Fatalf("cannot derive the first-failure record path from %s", retryWrapperRel)
	}
	for _, msg := range vbCheckRetryScript(script) {
		t.Error(msg)
	}
	for _, msg := range vbCheckRetryWorkflow(ci, rel, attempt1) {
		t.Error(msg)
	}

	scriptClean := len(vbCheckRetryScript(script)) == 0
	wfClean := len(vbCheckRetryWorkflow(ci, rel, attempt1)) == 0
	t.Run("control/wrapper preserves the first failure after the re-run", func(t *testing.T) {
		if !scriptClean {
			t.Fatalf("control precondition: the real wrapper is not clean for this guard, so the control proves nothing")
		}
		cpLine := `cp "$work/attempt1.json" "${log_file%.log}.attempt1.json"`
		if !strings.Contains(script, cpLine) {
			t.Fatalf("control did not apply")
		}
		mut := strings.Replace(script, cpLine, ":", 1) + "\n" + cpLine + "\n"
		got := vbCheckRetryScript(mut)
		if len(got) == 0 {
			t.Fatalf("a wrapper that preserves after the re-run passes the guard")
		}
		t.Logf("control observed red: %s", got[0])
	})
	wfMutants := []struct {
		name   string
		mutate func(w, r *rpcWorkflow)
	}{
		{"one retry-wrapped job uploads no attempt1 record", func(w, _ *rpcWorkflow) {
			for i := range w.Jobs["test-race-cli-1"].Steps {
				s := &w.Jobs["test-race-cli-1"].Steps[i]
				if strings.HasPrefix(s.Uses, "actions/upload-artifact@") {
					s.With["path"] = "test-stream-race-cli-1.json.gz"
				}
			}
		}},
		{"attempt1 upload without if: always()", func(w, _ *rpcWorkflow) {
			for i := range w.Jobs["test"].Steps {
				s := &w.Jobs["test"].Steps[i]
				if strings.HasPrefix(s.Uses, "actions/upload-artifact@") && strings.Contains(fmt.Sprint(s.With["path"]), attempt1) {
					s.If = ""
				}
			}
		}},
		{"ad-hoc retry loop beside the wrapper", func(w, _ *rpcWorkflow) {
			w.Jobs["lint"].Steps[0].Run = "for attempt in 1 2; do go test ./... && break; done"
		}},
		{"third-party retry action", func(w, _ *rpcWorkflow) {
			w.Jobs["lint"].Steps[0].Uses = "nick-fields/retry@v3"
		}},
		{"retry wrapper added to a release leg", func(_, r *rpcWorkflow) {
			r.Jobs["full-matrix-test"].Steps[0].Run = "bash scripts/ci/retry-flaky.sh -- go test ./..."
		}},
	}
	for _, m := range wfMutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if !wfClean {
				t.Fatalf("control precondition: the real workflows are not clean for this guard, so the control proves nothing")
			}
			c := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
			r := rpcParseWorkflow(t, rpcReleaseWorkflowRel, relText)
			m.mutate(c, r)
			got := vbCheckRetryWorkflow(c, r, attempt1)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard", m.name)
			}
			t.Logf("mutant observed red: %s", got[0])
		})
	}
}
