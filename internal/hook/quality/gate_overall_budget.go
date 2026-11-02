package quality

import (
	"fmt"
	"time"
)

// Overall-budget derivation (card t2020).
//
// The CLI used to bound a whole gate run with a fixed ten minutes while every
// step budget was configurable, so a project that raised timeouts.test to 900s
// died at 600s — with the parent deadline blamed on the step. The run budget
// is now derived from the steps the run will actually take.
const (
	// overallBudgetFloor is the budget a run has always had; the derivation
	// raises it, never lowers it.
	overallBudgetFloor = 10 * time.Minute
	// graphFreshnessAllowance covers the in-process graph-freshness step,
	// which carries no budget of its own.
	graphFreshnessAllowance = 30 * time.Second
	// overallSlackMin and overallSlackFraction give the sum headroom for
	// process start-up and between-step work: the larger of the two applies.
	overallSlackMin      = 60 * time.Second
	overallSlackFraction = 10 // percent of the step sum
)

// OverallBudget returns the budget for one whole gate run and, when the
// configured timeouts.total override had to be corrected, a one-line warning.
//
// Derived budget: for every detected toolchain, its vet steps x VetTimeout,
// plus TypecheckTimeout when the typecheck axis resolves a step, plus its lint
// steps x LintTimeout, plus TestTimeout when the test step runs; plus the
// graph-freshness and ast-grep allowances when those steps are enabled; plus
// slack of max(60s, 10%). It is never below the ten-minute floor and never
// below the largest single step budget; it is never unlimited.
//
// A positive TotalTimeout wins over the derivation, but a value below the
// largest single step cannot work, so it is raised to that step and the
// correction is announced in the returned warning. Zero or negative means
// "derive it".
func (g *QualityGate) OverallBudget() (time.Duration, string) {
	var sum, largest time.Duration
	note := func(d time.Duration, n int) {
		if n <= 0 {
			return
		}
		sum += time.Duration(n) * d
		largest = max(largest, d)
	}

	for _, dt := range g.detectToolchains() {
		dir := dt.root
		if dir == "" {
			dir = resolveQualityProjectDir(*g.config, "QualityGate.OverallBudget")
		}
		note(g.config.VetTimeout, len(dt.tc.vetSteps))
		if g.config.TypecheckEnabled {
			if _, _, ok := resolveTypecheckStep(dt.tc.typecheckStep, dir, g.config.TypecheckCommand); ok {
				note(g.config.TypecheckTimeout, 1)
			}
		}
		lintSteps, _, watchProne := resolveNodeLintSteps(dt.tc.lintSteps, dir)
		if !watchProne {
			note(g.config.LintTimeout, len(lintSteps))
		}
		if dt.tc.testStep != nil && !g.config.SkipTests {
			note(g.config.TestTimeout, 1)
		}
	}
	if gf := g.config.GraphFreshness; gf != nil && gf.Enabled {
		note(graphFreshnessAllowance, 1)
	}
	if ag := g.config.AstGrepGate; ag != nil && ag.Enabled {
		note(astGrepScanTimeout, 1)
	}

	derived := sum + max(overallSlackMin, sum*overallSlackFraction/100)
	derived = max(derived, overallBudgetFloor, largest)

	if g.config.TotalTimeout <= 0 {
		return derived, ""
	}
	if g.config.TotalTimeout < largest {
		return largest, fmt.Sprintf(
			"gate: timeouts.total=%s is below the largest single step budget %s; using %s",
			g.config.TotalTimeout, largest, largest)
	}
	return g.config.TotalTimeout, ""
}
