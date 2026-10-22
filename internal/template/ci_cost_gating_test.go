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
	"regexp"
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

// ---- F4: the classification itself -------------------------------------------
//
// rpcCheckFilterClasses proves each class is non-empty and documented, which a
// filter whose every root was replaced by a path nothing changes still satisfies.
// The checks below make the roots load-bearing: representative changed files
// (test-owned facts, each tied to the evidence comment of its root in the filter
// file) must be classified into the classes the cost-job table says read them,
// docs files must be classified into none, and the classification must compose
// with the real gate expressions into the documented run decision.

// rpcFilterMatchFixtures: class -> changed files the class must match.
var rpcFilterMatchFixtures = map[string][]string{
	"go_code": {
		"cmd/moai/main.go", "go.mod", "go.sum", "Makefile", ".github/workflows/ci.yml",
		".moai/config/sections/quality.yaml", "internal/template/templates/x.md",
		"internal/template/catalog.yaml", ".claude/agents/moai/manager-develop.md",
		".claude/rules/moai/core/moai-constitution.md", ".claude/settings.json",
		"scripts/ci/retry.sh", "scripts/release.sh",
	},
	"binary_inputs": {
		"cmd/moai/main.go", "internal/cli/doctor.go", "internal/template/templates/x.txt",
		"web/page.templ", "go.mod", "go.sum", "Makefile", ".golangci.yml", ".github/workflows/ci.yml",
	},
	"browser_inputs": {
		"cmd/moai/main.go", "internal/web/static/app.js", "go.mod", "go.sum", ".github/workflows/ci.yml",
	},
	"harness_inputs": {
		".claude/rules/moai/core/zone-registry.md", ".claude/agents/moai/manager-develop.md",
	},
}

// rpcFilterNoMatchFixtures: changed files no class may match (a docs-only diff).
var rpcFilterNoMatchFixtures = []string{"README.md", "CHANGELOG.md", "LICENSE", "docs-site/content/en/index.md"}

// rpcDiffScenarios: a changed-file set, the input classes it belongs to by the
// cost-job design decision, and (derived) which cost jobs a pull_request runs.
var rpcDiffScenarios = []struct {
	label   string
	files   []string
	classes []string
}{
	{"docs-only diff", []string{"README.md", "docs-site/content/en/index.md"}, nil},
	{"Go source diff", []string{"cmd/moai/main.go"}, []string{"binary_inputs", "browser_inputs"}},
	{"harness-only diff", []string{".claude/rules/moai/core/zone-registry.md"}, []string{"harness_inputs"}},
}

func rpcParseFilterRoots(text string) (map[string][]string, error) {
	var m map[string][]string
	if err := yaml.Unmarshal([]byte(text), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func rpcAnyGlob(patterns []string, file string) (bool, error) {
	for _, p := range patterns {
		ok, err := rpcGlobMatch(p, file)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func rpcCheckFilterFixtures(text string) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CI_COST_GATE_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	roots, err := rpcParseFilterRoots(text)
	if err != nil {
		return []string{"CI_COST_GATE_VIOLATION: " + rpcFilterRelPath() + " does not parse as class -> root list: " + err.Error()}
	}
	for _, class := range rpcSortedKeys(rpcFilterMatchFixtures) {
		for _, f := range rpcFilterMatchFixtures[class] {
			ok, err := rpcAnyGlob(roots[class], f)
			if err != nil {
				add("%s: class %s: %v", rpcFilterRelPath(), class, err)
			} else if !ok {
				add("%s: class %s matches no root for %q, a file it must classify", rpcFilterRelPath(), class, f)
			}
		}
		for _, f := range rpcFilterNoMatchFixtures {
			if ok, _ := rpcAnyGlob(roots[class], f); ok {
				add("%s: class %s matches %q, a docs file no class may classify", rpcFilterRelPath(), class, f)
			}
		}
	}
	return v
}

// rpcCheckClassificationEndToEnd classifies each scenario's files with the real
// filter, feeds the resulting classes into the real gate expressions as a
// successful `detect`, and compares which cost jobs run with the design table.
func rpcCheckClassificationEndToEnd(t *testing.T, wf *rpcWorkflow, filterText string) []string {
	t.Helper()
	var v []string
	roots, err := rpcParseFilterRoots(filterText)
	if err != nil {
		return []string{"CI_COST_GATE_VIOLATION: " + rpcFilterRelPath() + " does not parse as class -> root list: " + err.Error()}
	}
	for _, sc := range rpcDiffScenarios {
		got := map[string]bool{}
		for class, patterns := range roots {
			for _, f := range sc.files {
				ok, err := rpcAnyGlob(patterns, f)
				if err != nil {
					v = append(v, fmt.Sprintf("CI_COST_GATE_VIOLATION: %s: class %s: %v", rpcFilterRelPath(), class, err))
				}
				if ok {
					got[class] = true
				}
			}
		}
		cell := rpcCell{event: "pull_request", detect: "success", classes: got}
		for _, id := range rpcSortedKeys(rpcCostJobClasses) {
			j := wf.Jobs[id]
			if j == nil {
				continue
			}
			want := false
			for _, c := range rpcCostJobClasses[id] {
				if rpcContains(sc.classes, c) {
					want = true
				}
			}
			if runs := rpcJobRuns(t, j, cell.ctx()); runs != want {
				v = append(v, fmt.Sprintf("CI_COST_GATE_VIOLATION: %s: on a %s the filter classifies %v and the gate gives runs=%v, want %v", id, sc.label, rpcSortedKeys(got), runs, want))
			}
		}
	}
	return v
}

// rpcCheckDetectWiring checks the classification step ci.yml's gate reads from:
// the step `f` of the detect job runs dorny/paths-filter over the shared filter
// file, and detect re-exports go_code from it (the class exports are checked in
// rpcCheckCostJobsGated).
func rpcCheckDetectWiring(wf *rpcWorkflow) []string {
	var v []string
	add := func(format string, a ...any) {
		v = append(v, "CI_COST_GATE_VIOLATION: "+fmt.Sprintf(format, a...))
	}
	detect := wf.Jobs["detect"]
	if detect == nil {
		return []string{"CI_COST_GATE_VIOLATION: ci.yml has no detect job"}
	}
	var f *rpcStep
	for i := range detect.Steps {
		if detect.Steps[i].ID == "f" {
			f = &detect.Steps[i]
		}
	}
	switch {
	case f == nil:
		add("detect has no classification step with id f, so steps.f.outputs.* reads nothing")
	case !strings.HasPrefix(f.Uses, "dorny/paths-filter@"):
		add("detect step f runs %q, not dorny/paths-filter", f.Uses)
	default:
		if got := fmt.Sprint(f.With["filters"]); got != rpcFilterRelPath() {
			add("detect step f reads filters %q, not the shared %s", got, rpcFilterRelPath())
		}
	}
	if !strings.Contains(detect.Outputs["go_code"], "steps.f.outputs.go_code") {
		add("detect does not export go_code from the shared filter step (output = %q)", detect.Outputs["go_code"])
	}
	return v
}

var (
	rpcFilterKeyRe  = regexp.MustCompile(`^([a-z_]+):\s*$`)
	rpcFilterItemRe = regexp.MustCompile(`^  - '(.*)'\s*$`)
)

// rpcEditFilterRoots rewrites the `  - '<root>'` items of one class of the filter
// text (every class when class is ""); edit returns the replacement root, or ""
// to drop the item. Negative-control helper.
func rpcEditFilterRoots(t *testing.T, text, class string, edit func(root string) string) string {
	t.Helper()
	var out []string
	cur := ""
	for _, ln := range strings.Split(text, "\n") {
		if m := rpcFilterKeyRe.FindStringSubmatch(ln); m != nil {
			cur = m[1]
		}
		if m := rpcFilterItemRe.FindStringSubmatch(ln); m != nil && (class == "" || cur == class) {
			if repl := edit(m[1]); repl != "" {
				out = append(out, "  - '"+repl+"'")
			}
			continue
		}
		out = append(out, ln)
	}
	res := strings.Join(out, "\n")
	if res == text {
		t.Fatalf("control did not apply: no root of class %q changed", class)
	}
	return res
}

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
	// F4: the classification roots and their composition with the gate.
	for _, msg := range rpcCheckFilterFixtures(filterText) {
		t.Error(msg)
	}
	for _, msg := range rpcCheckDetectWiring(wf) {
		t.Error(msg)
	}
	for _, msg := range rpcCheckClassificationEndToEnd(t, wf, filterText) {
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
	// F4 controls on the filter text: each rewrite leaves every class non-empty and
	// documented (rpcCheckFilterClasses stays green) yet breaks the classification.
	filterClean := len(rpcCheckFilterFixtures(filterText)) == 0 && len(rpcCheckClassificationEndToEnd(t, wf, filterText)) == 0
	keep := func(root string) string { return root }
	filterControls := []struct {
		name string
		edit func(t *testing.T) string
	}{
		{"every classification path replaced by a non-matching path", func(t *testing.T) string {
			return rpcEditFilterRoots(t, filterText, "", func(string) string { return "zz-nonmatching/**" })
		}},
		{"binary_inputs loses internal/**", func(t *testing.T) string {
			return rpcEditFilterRoots(t, filterText, "binary_inputs", func(r string) string {
				if r == "internal/**" {
					return ""
				}
				return keep(r)
			})
		}},
		{"harness_inputs retargeted at a path nothing changes", func(t *testing.T) string {
			return rpcEditFilterRoots(t, filterText, "harness_inputs", func(string) string { return "zz-nonmatching/**" })
		}},
		{"go_code loses .moai/**", func(t *testing.T) string {
			return rpcEditFilterRoots(t, filterText, "go_code", func(r string) string {
				if r == ".moai/**" {
					return ""
				}
				return keep(r)
			})
		}},
		{"a docs tree classified as a binary input", func(t *testing.T) string {
			return rpcEditFilterRoots(t, filterText, "binary_inputs", func(r string) string {
				if r == ".github/test-input-filters.yml" {
					return "**/*.md"
				}
				return keep(r)
			})
		}},
	}
	for _, c := range filterControls {
		t.Run("control/filter/"+c.name, func(t *testing.T) {
			if !filterClean {
				t.Fatalf("control precondition: the real filter is not clean for this guard, so the control proves nothing")
			}
			mut := c.edit(t)
			if len(rpcCheckFilterClasses(mut)) != 0 {
				t.Logf("note: the class-shape guard already reads this mutant red")
			}
			got := append(rpcCheckFilterFixtures(mut), rpcCheckClassificationEndToEnd(t, wf, mut)...)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the classification guards", c.name)
			}
			t.Logf("mutant %q observed red: %s", c.name, got[0])
		})
	}
	wiringControls := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"classifier reads another filter file", func(w *rpcWorkflow) {
			for i := range w.Jobs["detect"].Steps {
				if w.Jobs["detect"].Steps[i].ID == "f" {
					w.Jobs["detect"].Steps[i].With["filters"] = ".github/other-filters.yml"
				}
			}
		}},
		{"classifier step id renamed", func(w *rpcWorkflow) {
			for i := range w.Jobs["detect"].Steps {
				if w.Jobs["detect"].Steps[i].ID == "f" {
					w.Jobs["detect"].Steps[i].ID = "g"
				}
			}
		}},
		{"classifier replaced by another action", func(w *rpcWorkflow) {
			for i := range w.Jobs["detect"].Steps {
				if w.Jobs["detect"].Steps[i].ID == "f" {
					w.Jobs["detect"].Steps[i].Uses = "actions/checkout@v7"
				}
			}
		}},
		{"go_code no longer exported from the classifier", func(w *rpcWorkflow) {
			w.Jobs["detect"].Outputs["go_code"] = "true"
		}},
	}
	for _, c := range wiringControls {
		t.Run("control/wiring/"+c.name, func(t *testing.T) {
			if len(rpcCheckDetectWiring(wf)) != 0 {
				t.Fatalf("control precondition: the real detect job is not clean for this guard, so the control proves nothing")
			}
			mut := rpcParseWorkflow(t, rpcCIWorkflowRel, ciText)
			c.mutate(mut)
			got := rpcCheckDetectWiring(mut)
			if len(got) == 0 {
				t.Fatalf("mutant %q passes the detect-wiring guard", c.name)
			}
			t.Logf("mutant %q observed red: %s", c.name, got[0])
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
	// rewriteVerdict replaces the script of every hard-failing step of the given
	// jobs (the verdict step of a real job, or the failing step of a marker).
	rewriteVerdict := func(ids func(w *rpcWorkflow) []string, script string) func(w *rpcWorkflow) {
		return func(w *rpcWorkflow) {
			for _, id := range ids(w) {
				for i := range w.Jobs[id].Steps {
					if rpcStepHardFails(w.Jobs[id].Steps[i]) {
						w.Jobs[id].Steps[i].Run = script
					}
				}
			}
		}
	}
	realJobs := func(*rpcWorkflow) []string { return rpcGatedRealJobs }
	lintMarkers := func(w *rpcWorkflow) []string { return markerIDs(w, "lint") }
	mutants := []struct {
		name   string
		mutate func(w *rpcWorkflow)
	}{
		{"no verdict step (run on failure, publish its success)", dropVerdict},
		// F6: the failing branch must be reachable, not merely present as text.
		{"verdict exit 1 inside an `if false` body", rewriteVerdict(realJobs, "echo \"verdict\"\nif false; then\n  exit 1\nfi\n")},
		{"verdict exit 1 behind an earlier exit 0", rewriteVerdict(realJobs, "echo \"verdict\"\nexit 0\nexit 1\n")},
		{"verdict exit 1 chained behind a failing test", rewriteVerdict(realJobs, "echo \"verdict\"\n[ -f /nonexistent ] && exit 1\n")},
		{"marker exit 1 inside an `if false` body", rewriteVerdict(lintMarkers, "echo \"verdict\"\nif false; then\n  exit 1\nfi\n")},
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

// TestScriptHardFailReachability pins what "this step hard-fails" means to the
// verdict guards (F6): an unconditional, reachable top-level `exit 1`. Each row
// is a script and whether it counts; the false rows are the unreachable or
// conditional spellings a text match would have credited.
func TestScriptHardFailReachability(t *testing.T) {
	cases := []struct {
		name   string
		script string
		want   bool
	}{
		{"bare exit 1", "exit 1", true},
		{"echo then exit 1", "echo \"::error::x\"\nexit 1\n", true},
		{"exit 1 with trailing comment", "echo x\nexit 1 # fail the check\n", true},
		{"indented, extra spaces", "\t  exit   1   \n", true},
		{"exit 1 then more text", "exit 1; echo after\n", true},
		{"exit 1 after a comment mentioning exit 0", "# exit 0 here would end it\nexit 1\n", true},
		{"exit 1 after a quoted exit 0 argument", "echo \"exit 0\"\nexit 1\n", true},

		{"empty script", "", false},
		{"exit 0 then exit 1", "echo x\nexit 0\nexit 1\n", false},
		{"exit 0; exit 1 on one line", "exit 0; exit 1\n", false},
		{"bare exit then exit 1", "exit\nexit 1\n", false},
		{"exit with a computed code then exit 1", "exit $rc\nexit 1\n", false},
		{"exit 2", "exit 2\n", false},
		{"if false body, multi-line", "if false; then\n  exit 1\nfi\n", false},
		{"if false body, one line", "if false; then exit 1; fi\n", false},
		{"if true body (a condition is never credited)", "if true; then\n  exit 1\nfi\n", false},
		{"elif branch", "if [ -z x ]; then\n  echo a\nelif [ -z y ]; then\n  exit 1\nfi\n", false},
		{"while body", "while true; do\n  exit 1\ndone\n", false},
		{"for body", "for x in a b; do exit 1; done\n", false},
		{"case arm", "case \"$X\" in\n  a) exit 1 ;;\nesac\n", false},
		{"&& chained", "[ -f /nonexistent ] && exit 1\n", false},
		{"|| chained", "test -n \"$x\" || { echo bad; exit 1; }\n", false},
		{"|| continued onto the next line", "test -n \"$x\" ||\n  exit 1\n", false},
		{"brace group", "{\n  exit 1\n}\n", false},
		{"subshell", "(\n  exit 1\n)\n", false},
		{"exit 1 line inside a quoted multi-line string", "echo \"first\nexit 1\nlast\"\n", false},
		{"exit 1 line inside a heredoc", "cat <<EOF\nexit 1\nEOF\n", false},
		{"exit 1 line inside a quoted heredoc", "cat <<'EOF'\nexit 1\nEOF\n", false},
		{"exit 1 only in a comment", "echo x\n# exit 1\n", false},
		{"exit 1 only as an echo argument", "echo exit 1\n", false},
		{"exit 1 inside command substitution", "x=$(echo a; exit 1)\n", false},
		{"exit 1 after a closed if", "if [ -z x ]; then\n  echo a\nfi\nexit 1\n", true},
		{"exit 1 after a loop that ends", "for x in a; do\n  echo $x\ndone\nexit 1\n", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := rpcScriptHardFails(c.script); got != c.want {
				t.Fatalf("rpcScriptHardFails(%q) = %v, want %v", c.script, got, c.want)
			}
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

// rpcOutcomeDiffFile is the one changed file a representative pull_request / push
// of each classification outcome carries; a workflow's `paths:` / `paths-ignore:`
// filter is evaluated against it. A classification failure says nothing about the
// diff, so it takes the docs-only file — the one a code-path filter skips.
var rpcOutcomeDiffFile = map[string]string{
	"docs-only":    "README.md",
	"go-code":      "cmd/moai/main.go",
	"harness-only": ".claude/rules/moai/core/zone-registry.md",
	"":             "README.md",
}

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
			diff := rpcOutcomeDiffFile[o.kind]
			for _, name := range set {
				reported := false
				var skippedBy []string
				for _, w := range append([]*rpcWorkflow{ci}, others...) {
					if !rpcTriggerMatches(w, ec.event, ec.branch) {
						continue
					}
					supplies := false
					if w == ci {
						supplies = len(rpcNonSkipped(rpcConclusions(t, ci, name, cell.ctx(), rpcAllStepsPass))) > 0
					} else {
						for _, id := range w.jobIDs() {
							if rpcContains(w.Jobs[id].displayNames(), name) {
								supplies = true
							}
						}
					}
					if !supplies {
						continue
					}
					// F5: a workflow that supplies the name still publishes nothing
					// when its own paths / paths-ignore filter skips this diff.
					_, paths, ignore := w.triggerPathFilters(ec.event)
					admitted, err := rpcPathsAdmit(paths, ignore, diff)
					if err != nil {
						add("%s: %s: %v", ec.label, w.Path, err)
					}
					if err != nil || !admitted {
						skippedBy = append(skippedBy, w.Path)
						continue
					}
					reported = true
				}
				if !reported && (o.kind != "" || !rpcFailClosedByAbsence[name]) {
					why := ""
					if len(skippedBy) > 0 {
						why = fmt.Sprintf(" (supplied only by %v, whose %s paths / paths-ignore filter skips the diff {%s})", skippedBy, ec.event, diff)
					}
					add("%s x %s: required name %q is not reported by any job%s", ec.label, o.label, name, why)
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
	// F5: a workflow that supplies a required name stops supplying it on a diff its
	// own `paths:` / `paths-ignore:` filter skips. Each control edits the real
	// workflow text (adds a pull_request path filter) and re-runs the guard.
	cqText := rpcReadRepoFile(t, rpcCodeQLWorkflowRel)
	relText := rpcReadRepoFile(t, rpcReleaseWorkflowRel)
	withFilter := func(t *testing.T, text, filter string) string {
		t.Helper()
		out := strings.Replace(text, "  pull_request:\n", "  pull_request:\n"+filter, 1)
		if out == text {
			t.Fatalf("control did not apply: no `  pull_request:` trigger to attach a filter to")
		}
		return out
	}
	pathsControls := []struct {
		name     string
		target   string // ci | codeql | release
		filter   string
		wantCell string // substring of a violation; empty = the guard must stay clean
	}{
		{"codeql paths exclude docs", "codeql", "    paths:\n      - 'internal/**'\n", `pull_request to main x docs-only: required name "Analyze (Go) (go)"`},
		{"codeql paths-ignore markdown", "codeql", "    paths-ignore:\n      - '**.md'\n", `pull_request to main x docs-only: required name "Analyze (Go) (go)"`},
		{"codeql go-only paths skip a harness diff", "codeql", "    paths:\n      - '**/*.go'\n", `pull_request to main x harness-only: required name "Analyze (Go) (go)"`},
		{"release gate behind a go-only paths filter", "release", "    paths:\n      - '**/*.go'\n", `pull_request to main x docs-only: required name "Release PR Multi-OS Gate"`},
		{"ci behind a code-only paths filter", "ci", "    paths:\n      - 'internal/**'\n", `pull_request to main x docs-only: required name "Lint"`},
		{"codeql paths admitting every file stays clean", "codeql", "    paths:\n      - '**'\n", ""},
		{"codeql paths-ignore of an unrelated tree stays clean", "codeql", "    paths-ignore:\n      - 'docs-site/**'\n", ""},
	}
	for _, c := range pathsControls {
		t.Run("control/paths filter/"+c.name, func(t *testing.T) {
			if !baselineClean {
				t.Fatalf("control precondition: the real tree is not clean for this guard, so the control proves nothing")
			}
			ciT, cqT, relT := ciText, cqText, relText
			switch c.target {
			case "ci":
				ciT = withFilter(t, ciText, c.filter)
			case "codeql":
				cqT = withFilter(t, cqText, c.filter)
			case "release":
				relT = withFilter(t, relText, c.filter)
			}
			got := rpcCheckRequiredNames(t, rpcParseWorkflow(t, rpcCIWorkflowRel, ciT), []*rpcWorkflow{
				rpcParseWorkflow(t, rpcCodeQLWorkflowRel, cqT),
				rpcParseWorkflow(t, rpcReleaseWorkflowRel, relT),
			}, rc)
			if c.wantCell == "" {
				if len(got) != 0 {
					t.Fatalf("a path filter that admits the diff was flagged: %s", got[0])
				}
				return
			}
			for _, g := range got {
				if strings.Contains(g, c.wantCell) {
					t.Logf("control observed red: %s", g)
					return
				}
			}
			t.Fatalf("path filter mutant not reported at %q; got %v", c.wantCell, got)
		})
	}
	t.Run("control/glob matcher", func(t *testing.T) {
		cases := []struct {
			pattern, file string
			want          bool
		}{
			{"**", "README.md", true},
			{"**.md", "docs/a/b.md", true},
			{"*.md", "docs/a.md", false},
			{"*.md", "README.md", true},
			{"internal/**", "internal/cli/x.go", true},
			{"internal/**", "README.md", false},
			{"**/*.go", "main.go", true},
			{"**/*.go", "cmd/moai/main.go", true},
			{"**/*.go", "README.md", false},
			{".claude/rules/moai/**", ".claude/rules/moai/core/a.md", true},
			{"docs/?.md", "docs/a.md", true},
			{"docs/?.md", "docs/ab.md", false},
		}
		for _, c := range cases {
			got, err := rpcGlobMatch(c.pattern, c.file)
			if err != nil || got != c.want {
				t.Errorf("rpcGlobMatch(%q, %q) = %v, %v; want %v", c.pattern, c.file, got, err, c.want)
			}
		}
		if _, err := rpcGlobMatch("!docs/**", "a"); err == nil {
			t.Errorf("a negated pattern must be refused, not guessed at")
		}
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
