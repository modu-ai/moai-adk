// ci_cost_gating_test.go: guards for SPEC-RELEASE-PIPELINE-001 M2 (card t1536,
// REQ-RPC-002 / REQ-RPC-003 / REQ-RPC-004).
//
// ci.yml's always-on cost jobs (lint, build, constitution-check, test-browser)
// are gated on shared-filter input classes for pull_request events only
// (decision-index Q8), fail open when the classification did not conclude
// success, and publish their retained required names (`Lint`, the five
// `Build (...)` names; decision-index Q4) through skip-markers plus a
// classification-verdict step so that "the classification failed" never yields
// a passing required name. Three guards:
//
//   - TestCICostJobsInputGated        (AC-RPC-002) the gate itself
//   - TestCIGatedFamilyVerdictSemantics (AC-RPC-003) what each required name concludes
//   - TestCIRequiredNamesSatisfied    (AC-RPC-004) every required-name set is
//     still reported on the event class it governs
//
// Each guard is a pure check function over a parsed workflow; the test runs it
// on the real tree and then on in-memory mutants. A mutant that does not turn
// the check red is a vacuous control, so every control first asserts its
// baseline is clean and that the mutation actually changed the workflow.
//
// Sentinels on failure: CI_COST_GATE_VIOLATION, CI_VERDICT_SEMANTICS_VIOLATION,
// CI_REQUIRED_NAME_VIOLATION.
package template_test

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	rpcCIWorkflowRel       = ".github/workflows/ci.yml"
	rpcCodeQLWorkflowRel   = ".github/workflows/codeql.yml"
	rpcReleaseWorkflowRel  = ".github/workflows/release-pr-multi-os.yml"
	rpcRequiredChecksRel   = ".github/required-checks.yml"
	rpcBranchProtectionRel = ".github/branch-protection.json.gtmpl"
	rpcBranchProtectionMir = "internal/template/templates/.github/branch-protection.json.gtmpl"
)

// rpcCostJobClasses maps each gated cost job to the filter classes whose
// positive match runs it (the design decision the filter file's evidence
// comments document).
var rpcCostJobClasses = map[string][]string{
	"lint":               {"binary_inputs", "harness_inputs"},
	"build":              {"binary_inputs"},
	"constitution-check": {"binary_inputs", "harness_inputs"},
	"test-browser":       {"browser_inputs"},
}

// rpcGatedRealJobs are the cost jobs that publish retained required names.
var rpcGatedRealJobs = []string{"lint", "build"}

func rpcSortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func rpcContains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// ---- AC-RPC-002: the gate ---------------------------------------------------

func rpcCheckCostJobsGated(t *testing.T, wf *rpcWorkflow) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CI_COST_GATE_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	detect := wf.Jobs["detect"]
	if detect == nil {
		return []string{"CI_COST_GATE_VIOLATION: ci.yml has no detect job"}
	}
	for _, k := range rpcClassKeys[1:] {
		if !strings.Contains(detect.Outputs[k], "steps.f.outputs."+k) {
			add("detect does not export the %s class from the shared filter step (output = %q)", k, detect.Outputs[k])
		}
	}
	for _, id := range rpcSortedKeys(rpcCostJobClasses) {
		classes := rpcCostJobClasses[id]
		j := wf.Jobs[id]
		if j == nil {
			add("job %s is missing", id)
			continue
		}
		if !rpcContains(j.needsList(), "detect") {
			add("%s: does not need detect, so it cannot read the input classification", id)
		}
		if strings.Contains(j.If, "always()") {
			add("%s: its condition is built on always(), which starts the job after the run is cancelled", id)
		}
		if !strings.Contains(j.If, "!cancelled()") {
			add("%s: its condition is not !cancelled()-guarded", id)
		}
		runs := func(c rpcCell) bool { return rpcJobRuns(t, j, c.ctx()) }
		if runs(rpcCell{event: "pull_request", detect: "success", classes: rpcOutcome("docs-only")}) {
			add("%s: runs on a docs-only pull_request (no input class matched)", id)
		}
		for _, k := range rpcClassKeys[1:] {
			want := rpcContains(classes, k)
			if got := runs(rpcCell{event: "pull_request", detect: "success", classes: rpcOnly(k)}); got != want {
				add("%s: with only %s matched it runs=%v, want %v", id, k, got, want)
			}
		}
		for _, d := range []string{"failure", "cancelled"} {
			if !runs(rpcCell{event: "pull_request", detect: d}) {
				add("%s: does not run when the classification concluded %s (fail-open)", id, d)
			}
		}
		for _, d := range []string{"success", "failure"} {
			if runs(rpcCell{event: "pull_request", detect: d, classes: rpcOutcome("go-code"), cancelled: true}) {
				add("%s: starts although the run itself is cancelled (detect=%s)", id, d)
			}
		}
		for _, ev := range []string{"push", "workflow_dispatch"} {
			for _, d := range []string{"success", "failure"} {
				if !runs(rpcCell{event: ev, detect: d, classes: rpcOutcome("docs-only")}) {
					add("%s: does not run on %s with detect=%s (Q8-exempt event class)", id, ev, d)
				}
			}
		}
	}
	return v
}

// rpcClassReaders inverts rpcCostJobClasses: class -> job ids reading it.
func rpcClassReaders() map[string][]string {
	out := map[string][]string{}
	for _, id := range rpcSortedKeys(rpcCostJobClasses) {
		for _, c := range rpcCostJobClasses[id] {
			out[c] = append(out[c], id)
		}
	}
	return out
}

// rpcCheckFilterClasses checks the shared filter file defines the three input
// classes, each with a non-empty root list and an evidence comment naming the
// jobs whose verdict reads it.
func rpcCheckFilterClasses(text string) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CI_COST_GATE_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return []string{"CI_COST_GATE_VIOLATION: " + rpcFilterRelPath() + " does not parse as a mapping"}
	}
	m := doc.Content[0]
	readers := rpcClassReaders()
	for _, class := range rpcClassKeys[1:] {
		var key, val *yaml.Node
		for i := 0; i+1 < len(m.Content); i += 2 {
			if m.Content[i].Value == class {
				key, val = m.Content[i], m.Content[i+1]
			}
		}
		if key == nil || val.Kind != yaml.SequenceNode || len(val.Content) == 0 {
			add("%s: class %s missing or empty", rpcFilterRelPath(), class)
			continue
		}
		for _, job := range readers[class] {
			if !strings.Contains(key.HeadComment, job) {
				add("%s: class %s has no evidence comment naming the %s job that reads it", rpcFilterRelPath(), class, job)
			}
		}
		first := val.Content[0]
		if first.HeadComment == "" && first.LineComment == "" {
			add("%s: class %s: first root %q carries no evidence comment", rpcFilterRelPath(), class, first.Value)
		}
	}
	return v
}

func rpcFilterRelPath() string { return ".github/test-input-filters.yml" }

// rpcStripKeyHeadComment removes the contiguous comment block directly above
// `key:` in filter text (negative-control helper).
func rpcStripKeyHeadComment(t *testing.T, text, key string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		if strings.HasPrefix(ln, key+":") {
			start := i
			for start > 0 && strings.HasPrefix(lines[start-1], "#") {
				start--
			}
			if start == i {
				t.Fatalf("control did not apply: %s has no head comment to strip", key)
			}
			return strings.Join(append(append([]string{}, lines[:start]...), lines[i:]...), "\n")
		}
	}
	t.Fatalf("control did not apply: key %s not found", key)
	return ""
}

func TestCICostJobsInputGated(t *testing.T) {
	ciText := rpcReadRepoFile(t, rpcCIWorkflowRel)
	wf := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
	for _, msg := range rpcCheckCostJobsGated(t, wf) {
		t.Error(msg)
	}
	filterText := rpcReadRepoFile(t, rpcFilterRelPath())
	for _, msg := range rpcCheckFilterClasses(filterText) {
		t.Error(msg)
	}

	// Negative controls: each mutant must turn the check red. The baseline must
	// be clean first, or a control proves nothing.
	baselineClean := len(rpcCheckCostJobsGated(t, wf)) == 0
	snap := func(w *rpcWorkflow) string {
		return w.Jobs["lint"].If + "|" + strconv.Itoa(int(w.Jobs["build"].Needs.Kind)) + "|" + w.Jobs["test-browser"].If
	}
	mutants := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"nothing-matching gate (!cancelled() && false)", func(w *rpcWorkflow) { w.Jobs["lint"].If = "${{ !cancelled() && false }}" }},
		{"no fail-open clause", func(w *rpcWorkflow) {
			w.Jobs["lint"].If = "${{ !cancelled() && (github.event_name != 'pull_request' || needs.detect.outputs.binary_inputs == 'true' || needs.detect.outputs.harness_inputs == 'true') }}"
		}},
		{"always() instead of !cancelled()", func(w *rpcWorkflow) {
			w.Jobs["lint"].If = strings.ReplaceAll(w.Jobs["lint"].If, "!cancelled()", "always()")
		}},
		{"gate without needs: detect", func(w *rpcWorkflow) { w.Jobs["build"].Needs = yaml.Node{} }},
		{"gates push events too", func(w *rpcWorkflow) {
			w.Jobs["test-browser"].If = "${{ !cancelled() && (needs.detect.result != 'success' || needs.detect.outputs.browser_inputs == 'true') }}"
		}},
	}
	for _, m := range mutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real tree is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
			before := snap(mut)
			m.mutate(mut)
			if snap(mut) == before {
				t.Fatalf("control did not apply: the mutation changed nothing")
			}
			got := rpcCheckCostJobsGated(t, mut)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard: the guard is vacuous in that direction", m.name)
			}
			t.Logf("mutant %q observed red: %s", m.name, got[0])
		})
	}
	t.Run("control/filter class without evidence comment", func(t *testing.T) {
		if len(rpcCheckFilterClasses(filterText)) != 0 {
			t.Fatalf("control precondition: the real filter file is not clean")
		}
		got := rpcCheckFilterClasses(rpcStripKeyHeadComment(t, filterText, "binary_inputs"))
		if len(got) == 0 {
			t.Fatalf("a class with no evidence comment passes the guard")
		}
		t.Logf("mutant observed red: %s", got[0])
	})
}

// ---- AC-RPC-003: verdict semantics -----------------------------------------

func rpcGatedNames(rc rpcRequiredChecks) []string {
	var out []string
	for _, c := range rc.Branches["main"].Contexts {
		if c == "Lint" || strings.HasPrefix(c, "Build (") {
			out = append(out, c)
		}
	}
	return out
}

func rpcCheckVerdictSemantics(t *testing.T, wf *rpcWorkflow, names []string) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CI_VERDICT_SEMANTICS_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	if len(names) == 0 {
		return []string{"CI_VERDICT_SEMANTICS_VIOLATION: no gated required names to check (empty sweep)"}
	}
	matched, nonMatch := rpcOutcome("go-code"), rpcOutcome("docs-only")
	for _, name := range names {
		conc := func(c rpcCell, mode rpcStepMode) []string {
			return rpcNonSkipped(rpcConclusions(t, wf, name, c.ctx(), mode))
		}
		// (i) the input-relevant job concludes failure -> the name fails.
		pr := rpcCell{event: "pull_request", detect: "success", classes: matched}
		if got := conc(pr, rpcVerificationFails); !rpcAllEqual(got, "failure", 0) {
			add("%s: (i) class matched + verification fails: published %v, want only failure", name, got)
		}
		if got := conc(pr, rpcAllStepsPass); !rpcAllEqual(got, "success", 0) {
			add("%s: (i) class matched + verification passes: published %v, want only success", name, got)
		}
		// (ii) positively observed non-match -> the marker passes, the real job is skipped.
		if got := conc(rpcCell{event: "pull_request", detect: "success", classes: nonMatch}, rpcAllStepsPass); !rpcAllEqual(got, "success", 1) {
			add("%s: (ii) positive non-match: published %v, want exactly one success", name, got)
		}
		// (iii) the classification did not conclude success, run not cancelled:
		// the real job AND the marker run and both fail (the counter-case is
		// pass mode: classification failure x verification success).
		for _, d := range []string{"failure", "cancelled"} {
			for _, mode := range []rpcStepMode{rpcAllStepsPass, rpcVerificationFails} {
				got := conc(rpcCell{event: "pull_request", detect: d}, mode)
				if !rpcAllEqual(got, "failure", 2) {
					add("%s: (iii) detect=%s mode=%d: published %v, want the real job and the marker both failure", name, d, mode, got)
				}
			}
		}
		// (iv) the run itself is cancelled: nothing may start.
		for _, d := range []string{"success", "failure", "cancelled"} {
			if got := conc(rpcCell{event: "pull_request", detect: d, classes: matched, cancelled: true}, rpcAllStepsPass); len(got) != 0 {
				add("%s: (iv) run cancelled (detect=%s): %v started after the cancellation", name, d, got)
			}
		}
		// Q8-exempt event classes: no marker, the name is the real job's own outcome.
		for _, ev := range []string{"push", "workflow_dispatch"} {
			for _, d := range []string{"success", "failure"} {
				if got := conc(rpcCell{event: ev, detect: d, classes: nonMatch}, rpcAllStepsPass); !rpcAllEqual(got, "success", 1) {
					add("%s: %s detect=%s, verification passes: published %v, want exactly the real job's success", name, ev, d, got)
				}
				if got := conc(rpcCell{event: ev, detect: d, classes: nonMatch}, rpcVerificationFails); !rpcAllEqual(got, "failure", 1) {
					add("%s: %s detect=%s, verification fails: published %v, want exactly the real job's failure", name, ev, d, got)
				}
			}
		}
	}
	// No cost job or marker condition may be built on always().
	for _, id := range wf.jobIDs() {
		j := wf.Jobs[id]
		publishes := false
		for _, n := range j.displayNames() {
			if rpcContains(names, n) {
				publishes = true
			}
		}
		if (publishes || rpcCostJobClasses[id] != nil) && strings.Contains(j.If, "always()") {
			add("job %s: its condition uses always(), so it starts after the run is cancelled", id)
		}
	}
	return v
}

func rpcLoadRequiredChecks(t *testing.T) rpcRequiredChecks {
	t.Helper()
	return rpcParseRequiredChecks(t, rpcReadRepoFile(t, rpcRequiredChecksRel))
}

// rpcRequiredChecks is the parsed .github/required-checks.yml.
type rpcRequiredChecks struct {
	Branches map[string]struct {
		Contexts []string `yaml:"contexts"`
	} `yaml:"branches"`
	ReleaseGate []string `yaml:"release_gate"`
}

func rpcParseRequiredChecks(t *testing.T, text string) rpcRequiredChecks {
	t.Helper()
	var rc rpcRequiredChecks
	if err := yaml.Unmarshal([]byte(text), &rc); err != nil {
		t.Fatalf("parse %s: %v", rpcRequiredChecksRel, err)
	}
	return rc
}

func TestCIGatedFamilyVerdictSemantics(t *testing.T) {
	ciText := rpcReadRepoFile(t, rpcCIWorkflowRel)
	wf := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
	rc := rpcLoadRequiredChecks(t)
	names := rpcGatedNames(rc)
	for _, msg := range rpcCheckVerdictSemantics(t, wf, names) {
		t.Error(msg)
	}

	baselineClean := len(rpcCheckVerdictSemantics(t, wf, names)) == 0
	markerIDs := func(w *rpcWorkflow, real string) []string {
		var ids []string
		for _, id := range w.jobIDs() {
			if id == real {
				continue
			}
			for _, n := range w.Jobs[id].displayNames() {
				if (n == "Lint" && real == "lint") || (strings.HasPrefix(n, "Build (") && real == "build") {
					ids = append(ids, id)
					break
				}
			}
		}
		return ids
	}
	dropVerdict := func(w *rpcWorkflow) {
		for _, id := range rpcGatedRealJobs {
			var keep []rpcStep
			for _, s := range w.Jobs[id].Steps {
				if !rpcStepHardFails(s) || !strings.Contains(s.If, "always()") {
					keep = append(keep, s)
				}
			}
			w.Jobs[id].Steps = keep
		}
	}
	mutants := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"no verdict step (run on failure, publish its success)", dropVerdict},
		{"marker passes on classification failure", func(w *rpcWorkflow) {
			for _, id := range markerIDs(w, "lint") {
				for i := range w.Jobs[id].Steps {
					if rpcStepHardFails(w.Jobs[id].Steps[i]) {
						w.Jobs[id].Steps[i].If = "${{ false }}"
					}
				}
			}
		}},
		{"marker skipped on classification failure", func(w *rpcWorkflow) {
			for _, id := range markerIDs(w, "lint") {
				w.Jobs[id].If = "${{ !cancelled() && github.event_name == 'pull_request' && needs.detect.result == 'success' && needs.detect.outputs.binary_inputs != 'true' }}"
			}
		}},
		{"always() job condition", func(w *rpcWorkflow) {
			w.Jobs["lint"].If = strings.Replace(w.Jobs["lint"].If, "!cancelled()", "always()", 1)
		}},
		{"verdict step on a job still skipped when detect fails", func(w *rpcWorkflow) {
			w.Jobs["lint"].If = "${{ github.event_name != 'pull_request' || needs.detect.outputs.binary_inputs == 'true' }}"
		}},
	}
	for _, m := range mutants {
		t.Run("control/"+m.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real tree is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
			m.mutate(mut)
			got := rpcCheckVerdictSemantics(t, mut, names)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the guard: the guard is vacuous in that direction", m.name)
			}
			t.Logf("mutant %q observed red: %s", m.name, got[0])
		})
	}
}

// ---- AC-RPC-004: required names per event class ------------------------------

type rpcEventClass struct {
	label  string
	event  string
	branch string
	set    func(rc rpcRequiredChecks) []string
}

var rpcEventClasses = []rpcEventClass{
	{"pull_request to main", "pull_request", "main", func(rc rpcRequiredChecks) []string { return rc.Branches["main"].Contexts }},
	{"push to ci/**", "push", "ci/**", func(rc rpcRequiredChecks) []string { return rc.Branches["ci/**"].Contexts }},
	{"push to main", "push", "main", func(rc rpcRequiredChecks) []string { return rc.ReleaseGate }},
}

// rpcFailClosedByAbsence names required checks that are ALREADY unreported when
// the classification step fails, before this SPEC: the Go test family
// (`test` / `test-skip-marker`) and nothing else is gated on go_code through
// `needs: detect` with no failure arm (PRESERVE item 4). M2 does not change
// them, so the classification-failure outcome does not assert them.
var rpcFailClosedByAbsence = map[string]bool{"Test (ubuntu-latest)": true}

func rpcTriggerMatches(wf *rpcWorkflow, event, branch string) bool {
	present, branches := wf.triggerBranches(event)
	if !present {
		return false
	}
	return len(branches) == 0 || rpcContains(branches, branch)
}

func rpcCheckRequiredNames(t *testing.T, ci *rpcWorkflow, others []*rpcWorkflow, rc rpcRequiredChecks) []string {
	t.Helper()
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CI_REQUIRED_NAME_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	outcomes := []struct {
		label  string
		detect string
		kind   string
	}{
		{"docs-only", "success", "docs-only"},
		{"go_code", "success", "go-code"},
		{"harness-only", "success", "harness-only"},
		{"classification failure", "failure", ""},
	}
	for _, ec := range rpcEventClasses {
		set := ec.set(rc)
		if len(set) == 0 {
			add("%s: the governing required-name set is empty (empty sweep)", ec.label)
		}
		for _, o := range outcomes {
			cell := rpcCell{event: ec.event, detect: o.detect}
			if o.kind != "" {
				cell.classes = rpcOutcome(o.kind)
			}
			for _, name := range set {
				reported := false
				if rpcTriggerMatches(ci, ec.event, ec.branch) &&
					len(rpcNonSkipped(rpcConclusions(t, ci, name, cell.ctx(), rpcAllStepsPass))) > 0 {
					reported = true
				}
				for _, w := range others {
					if !rpcTriggerMatches(w, ec.event, ec.branch) {
						continue
					}
					for _, id := range w.jobIDs() {
						if rpcContains(w.Jobs[id].displayNames(), name) {
							reported = true
						}
					}
				}
				if !reported && (o.kind != "" || !rpcFailClosedByAbsence[name]) {
					add("%s x %s: required name %q is not reported by any job", ec.label, o.label, name)
				}
			}
		}
	}
	// release_gate must stay a subset of main's contexts (scripts/release.sh reads it).
	for _, n := range rc.ReleaseGate {
		if !rpcContains(rc.Branches["main"].Contexts, n) {
			add("release_gate name %q is not in branches.main.contexts", n)
		}
	}
	return v
}

func TestCIRequiredNamesSatisfied(t *testing.T) {
	ciText := rpcReadRepoFile(t, rpcCIWorkflowRel)
	ci := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
	others := []*rpcWorkflow{
		rpcLoadWorkflow(t, rpcCodeQLWorkflowRel),
		rpcLoadWorkflow(t, rpcReleaseWorkflowRel),
	}
	rcText := rpcReadRepoFile(t, rpcRequiredChecksRel)
	rc := rpcParseRequiredChecks(t, rcText)
	for _, msg := range rpcCheckRequiredNames(t, ci, others, rc) {
		t.Error(msg)
	}

	// REQ-RPC-004: harness roots stay inside go_code, so release 3-OS coverage
	// (keyed to go_code) is not narrowed by input splitting.
	var filter map[string][]string
	if err := yaml.Unmarshal([]byte(rpcReadRepoFile(t, rpcFilterRelPath())), &filter); err != nil {
		t.Fatalf("parse %s: %v", rpcFilterRelPath(), err)
	}
	// Presence of the class is TestCICostJobsInputGated's obligation; this
	// assertion is the containment of whatever roots it declares, so the guard
	// is green before the class exists and stays a regression guard after.
	for _, root := range filter["harness_inputs"] {
		if !rpcContains(filter["go_code"], root) {
			t.Errorf("CI_REQUIRED_NAME_VIOLATION: harness root %q is outside go_code — release 3-OS coverage would narrow", root)
		}
	}

	// The branch-protection template and its distributed mirror are SHAPE-ONLY:
	// no required check name may appear in them (plan.md PRESERVE item 11).
	all := map[string]bool{}
	for _, c := range rc.Branches["main"].Contexts {
		all[c] = true
	}
	for _, c := range rc.Branches["ci/**"].Contexts {
		all[c] = true
	}
	for _, rel := range []string{rpcBranchProtectionRel, rpcBranchProtectionMir} {
		text := rpcReadRepoFile(t, rel)
		for _, name := range rpcSortedKeys(all) {
			if strings.Contains(text, `"`+name+`"`) {
				t.Errorf("CI_REQUIRED_NAME_VIOLATION: %s carries the check name %q — it must stay shape-only", rel, name)
			}
		}
	}

	// Negative controls (verification-completeness.md 1.1), both observed red.
	baselineClean := len(rpcCheckRequiredNames(t, ci, others, rc)) == 0
	t.Run("control/lint gated with no Lint marker", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real tree is not clean for this guard, so the control proves nothing")
		}
		mut := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
		mut.Jobs["lint"].If = "${{ github.event_name != 'pull_request' || needs.detect.outputs.binary_inputs == 'true' }}"
		mut.Jobs["lint"].Needs = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "detect"}
		for _, id := range mut.jobIDs() {
			if id != "lint" && rpcContains(mut.Jobs[id].displayNames(), "Lint") {
				delete(mut.Jobs, id)
			}
		}
		got := rpcCheckRequiredNames(t, mut, others, rc)
		found := false
		for _, g := range got {
			if strings.Contains(g, `pull_request to main x docs-only: required name "Lint"`) {
				found = true
			}
		}
		if !found {
			t.Fatalf("a gated lint with no Lint marker did not report Lint missing on the docs-only pull_request cell; got %v", got)
		}
		t.Logf("control (a) observed red: %s", got[0])
	})
	t.Run("control/release gate name in release_gate", func(t *testing.T) {
		if !baselineClean {
			t.Fatalf("control precondition: the real tree is not clean for this guard, so the control proves nothing")
		}
		mut := rpcParseRequiredChecks(t, rcText)
		mut.ReleaseGate = append(append([]string{}, mut.ReleaseGate...), "Release PR Multi-OS Gate")
		got := rpcCheckRequiredNames(t, ci, others, mut)
		found := false
		for _, g := range got {
			if strings.Contains(g, `push to main x docs-only: required name "Release PR Multi-OS Gate"`) {
				found = true
			}
		}
		if !found {
			t.Fatalf("adding Release PR Multi-OS Gate to release_gate did not report it unpublished on push to main; got %v", got)
		}
		t.Logf("control (b) observed red: %s", got[0])
	})
}
