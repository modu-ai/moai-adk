package quality

// affected_scope_gate.go — wires the affected-set core into the Go test step
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-001/-008/-010).

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// goTestStepName is the Go toolchain's test step, the only step the affected
// scope narrows.
const goTestStepName = "go test"

// goAllPackagesPattern is the scope the Go test step has without the affected
// scope: every package of the module.
const goAllPackagesPattern = "./..."

// scopeQueries are the two queries the affected scope depends on; tests
// substitute them to stay deterministic. The zero value uses the real adapters.
type scopeQueries struct {
	changes func(ctx context.Context, dir string) (commitChanges, error)
	graph   func(ctx context.Context, dir, tags string) (moduleGraph, error)
}

// stepRunFunc is the boundary executeStep hands a prepared command to. Production
// uses QualityGate.runStep; the budget test substitutes a recorder that captures
// the budget each step receives.
type stepRunFunc func(ctx context.Context, stepName, dir string, timeout time.Duration, name string, args ...string) (bool, string)

// queries returns the substituted queries, or the real adapters.
func (g *QualityGate) queries() scopeQueries {
	q := scopeQueries{changes: queryCommitChanges, graph: listModuleGraph}
	if g.scopeQ != nil {
		if g.scopeQ.changes != nil {
			q.changes = g.scopeQ.changes
		}
		if g.scopeQ.graph != nil {
			q.graph = g.scopeQ.graph
		}
	}
	return q
}

// affectedScopeApplies reports whether the Go test step runs the affected
// scope: only the Go toolchain's test step, only in the pre-commit context, and
// not when gate.pre_commit.test_scope.mode is "full" (REQ-PAS-010/-011).
func (g *QualityGate) affectedScopeApplies(step gateStep) bool {
	return g.config.PreCommitContext &&
		g.config.PreCommitTestScope != config.GateTestScopeFull &&
		step.name == goTestStepName && step.binary == "go"
}

// scopeGoTestStep narrows the Go test step to the packages the commit can
// affect. It returns the step to execute, a notice block for the gate summary,
// and a non-empty skip reason when nothing is left to run. Any failure of the
// computation returns the step unchanged — the full module — with a notice, and
// never a pass derived from the failed computation (REQ-PAS-008).
//
// @MX:NOTE: [AUTO] the only place the pre-commit context changes what the gate tests; every other path leaves the step untouched
// @MX:SPEC: SPEC-PRECOMMIT-AFFECTED-SCOPE-001
func (g *QualityGate) scopeGoTestStep(ctx context.Context, step gateStep, dir string) (gateStep, string, string) {
	if !g.affectedScopeApplies(step) {
		return step, "", ""
	}
	// A step the configuration turned off keeps its "disabled" row: executeStep
	// reports it, so it must not be recorded skipped here.
	if disabled, ok := g.config.DisabledSteps[step.name]; ok && !disabled {
		return step, "", ""
	}

	// Both queries share one budget, the vet budget: they are analysis, not a
	// test run, and no new number is introduced.
	if g.config.VetTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.config.VetTimeout)
		defer cancel()
	}
	fallback := func(what string, err error) (gateStep, string, string) {
		return step, fmt.Sprintf("%s%s failed (%v); falling back to the full module (go test %s)",
			scopeNotePrefix, what, err, goAllPackagesPattern), ""
	}

	q := g.queries()
	changes, err := q.changes(ctx, dir)
	if err != nil {
		return fallback("changed-file query", err)
	}
	// A dependency-graph file widens the scope without needing the graph.
	if _, widened := widenTrigger(changes.Files); widened {
		res := selectAffected(affectedInput{Changed: changes.Files, ModuleRel: changes.ModulePrefix})
		return step, strings.Join(res.Notes, "\n"), ""
	}
	// The graph comes from the working tree and the changed set from the index.
	// A Go file that differs between them could make the graph omit an edge the
	// committed code still has, so the scope is not trusted: run the full module.
	if len(changes.UnstagedGo) > 0 {
		return step, fmt.Sprintf("%sworking tree differs from the index in %s; the package graph would not describe the commit — falling back to the full module (go test %s)",
			scopeNotePrefix, listNames(changes.UnstagedGo, 5), goAllPackagesPattern), ""
	}
	graph, err := q.graph(ctx, dir, g.config.GoBuildTags)
	if err != nil {
		return fallback("package-graph query", err)
	}

	res := selectAffected(affectedInput{Changed: changes.Files, ModuleRel: changes.ModulePrefix, Graph: graph})
	notes := strings.Join(res.Notes, "\n")
	if res.SkipReason != "" {
		return step, notes, res.SkipReason
	}
	return withScopedPackages(step, res.Run), notes, ""
}

// listNames renders up to limit names, with a count of the rest.
func listNames(names []string, limit int) string {
	if len(names) <= limit {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(names[:limit], ", "), len(names)-limit)
}

// withScopedPackages replaces the all-packages pattern of a Go test step with
// the selected module-root-relative directory patterns, keeping every other
// argument (the subcommand, build tags) as it was.
func withScopedPackages(step gateStep, patterns []string) gateStep {
	out := step
	out.args = slices.DeleteFunc(slices.Clone(step.args), func(a string) bool { return a == goAllPackagesPattern })
	out.args = append(out.args, patterns...)
	return out
}
