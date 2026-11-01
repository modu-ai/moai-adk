package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// Card t2008 / RUN-10. The run workflow was renumbered to phases 11-20, but a
// "proceeding to Phase N" hand-off and the run router's phase labels kept the
// old numbers, so a reader following the text is sent to a phase that does not
// exist. The check is deliberately narrow: hand-offs must name an existing
// phase heading, and the router must not carry the retired labels.

const (
	taskDecompositionPath = ".claude/skills/moai/workflows/run/task-decomposition.md"
	runRouterPath         = ".claude/skills/moai/workflows/run.md"
)

var (
	phaseHeadingRE = regexp.MustCompile(`(?m)^## Phase (\d+):`)
	phaseHandoffRE = regexp.MustCompile(`proceeding to Phase (\d+)`)
)

// phaseReferenceViolations lists dangling hand-offs in the task decomposition
// body and retired phase labels in the router body.
func phaseReferenceViolations(taskDecomp, router string) []string {
	var v []string
	exists := map[string]bool{}
	for _, m := range phaseHeadingRE.FindAllStringSubmatch(taskDecomp, -1) {
		exists[m[1]] = true
	}
	for _, m := range phaseHandoffRE.FindAllStringSubmatch(taskDecomp, -1) {
		if !exists[m[1]] {
			v = append(v, "task-decomposition: 'proceeding to Phase "+m[1]+"' names no phase heading")
		}
	}
	for _, stale := range []string{"Phase 11~4", "2.8"} {
		if strings.Contains(router, stale) {
			v = append(v, "run router: retired phase label "+stale)
		}
	}
	return v
}

// TestRunPhaseReferences_Deployed checks the deployed template bodies.
func TestRunPhaseReferences_Deployed(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	td, err := fs.ReadFile(fsys, taskDecompositionPath)
	if err != nil {
		t.Fatalf("read %s: %v", taskDecompositionPath, err)
	}
	rt, err := fs.ReadFile(fsys, runRouterPath)
	if err != nil {
		t.Fatalf("read %s: %v", runRouterPath, err)
	}
	if v := phaseReferenceViolations(string(td), string(rt)); len(v) > 0 {
		t.Errorf("stale run phase references:\n  %s", strings.Join(v, "\n  "))
	}
}

// TestRunPhaseReferences_Mutants proves the checker can fail in each direction
// and that a correct hand-off passes.
func TestRunPhaseReferences_Mutants(t *testing.T) {
	t.Parallel()

	const headings = "## Phase 18: MX\n## Phase 19: Git\n## Phase 20: Done\n"
	if v := phaseReferenceViolations(headings+"before proceeding to Phase 3.", "ok"); len(v) != 1 {
		t.Errorf("dangling hand-off must be one violation, got %v", v)
	}
	if v := phaseReferenceViolations(headings+"before proceeding to Phase 19.", "ok"); len(v) != 0 {
		t.Errorf("valid hand-off must pass, got %v", v)
	}
	if v := phaseReferenceViolations(headings, "| Phase 11~4: Implementation |"); len(v) != 1 {
		t.Errorf("retired router label must be one violation, got %v", v)
	}
	if v := phaseReferenceViolations(headings, "quality validation (Phase 13/2.8)"); len(v) != 1 {
		t.Errorf("retired 2.8 label must be one violation, got %v", v)
	}
	if v := phaseReferenceViolations(headings, "| Phase 11-20: Implementation | Phase 13 |"); len(v) != 0 {
		t.Errorf("current labels must pass, got %v", v)
	}
}
