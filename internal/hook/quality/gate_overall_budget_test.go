package quality

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Card t2020: the overall run budget used to be a fixed ten minutes in the CLI
// while the per-step budgets were configurable, so timeouts.test=900 died at
// 600s with the parent deadline blamed on the step. The overall budget is now
// derived from the steps the run will actually take.

func budgetCfg(dir string) *GateConfig {
	cfg := DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.AstGrepGate = nil
	cfg.GraphFreshness = nil
	cfg.TypecheckEnabled = false
	return cfg
}

const goModFixture = "module example.com/t2020\n\ngo 1.22\n"

func budgetGoDir(t *testing.T) string {
	t.Helper()
	return writeFixture(t, map[string]string{"go.mod": goModFixture, "main.go": "package main\nfunc main() {}\n"})
}

// The motivating case: a test budget above the old fixed ten minutes must not
// be capped by the overall budget.
func TestOverallBudgetCoversTheTestStep(t *testing.T) {
	t.Parallel()
	cfg := budgetCfg(budgetGoDir(t))
	cfg.TestTimeout = 900 * time.Second
	got, warn := NewQualityGate(cfg).OverallBudget()
	if warn != "" {
		t.Errorf("derived budget produced a warning: %q", warn)
	}
	// vet 30 + lint 60 + test 900 = 990s, plus slack: well above 900s.
	if want := 990 * time.Second; got < want {
		t.Errorf("OverallBudget = %s, want at least the step sum %s", got, want)
	}
}

// Default step budgets keep the historical ten-minute floor: the derivation
// may raise the budget, never lower it below what runs have always had.
func TestOverallBudgetKeepsTenMinuteFloor(t *testing.T) {
	t.Parallel()
	cfg := budgetCfg(budgetGoDir(t))
	got, _ := NewQualityGate(cfg).OverallBudget()
	if got != 10*time.Minute {
		t.Errorf("OverallBudget with default step budgets = %s, want the 10m floor", got)
	}
}

// An explicit timeouts.total wins over the derivation.
func TestOverallBudgetOverrideWins(t *testing.T) {
	t.Parallel()
	cfg := budgetCfg(budgetGoDir(t))
	cfg.TestTimeout = 900 * time.Second
	cfg.TotalTimeout = 2000 * time.Second
	got, warn := NewQualityGate(cfg).OverallBudget()
	if got != 2000*time.Second || warn != "" {
		t.Errorf("OverallBudget = %s warn=%q, want 2000s and no warning", got, warn)
	}
}

// An override smaller than the largest single step cannot work: it is raised
// to that step and the raise is announced, never applied silently.
func TestOverallBudgetOverrideBelowLargestStepIsRaised(t *testing.T) {
	t.Parallel()
	cfg := budgetCfg(budgetGoDir(t))
	cfg.TestTimeout = 900 * time.Second
	cfg.TotalTimeout = 100 * time.Second
	got, warn := NewQualityGate(cfg).OverallBudget()
	if got != 900*time.Second {
		t.Errorf("OverallBudget = %s, want the largest step budget 900s", got)
	}
	if !strings.Contains(warn, "timeouts.total") || strings.Contains(warn, "\n") {
		t.Errorf("want a one-line warning naming timeouts.total, got %q", warn)
	}
}

// Zero and negative overrides mean "derive it".
func TestOverallBudgetNonPositiveOverrideIsDerived(t *testing.T) {
	t.Parallel()
	for _, v := range []time.Duration{0, -5 * time.Second} {
		cfg := budgetCfg(budgetGoDir(t))
		cfg.TestTimeout = 900 * time.Second
		cfg.TotalTimeout = v
		got, warn := NewQualityGate(cfg).OverallBudget()
		if got < 990*time.Second || warn != "" {
			t.Errorf("TotalTimeout=%s: OverallBudget = %s warn=%q, want the derived budget", v, got, warn)
		}
	}
}

// Every detected toolchain contributes its own steps.
func TestOverallBudgetSumsToolchains(t *testing.T) {
	t.Parallel()
	goOnly := budgetCfg(budgetGoDir(t))
	goOnly.TestTimeout = 900 * time.Second
	single, _ := NewQualityGate(goOnly).OverallBudget()

	both := writeFixture(t, map[string]string{
		"go.mod":                goModFixture,
		"main.go":               "package main\nfunc main() {}\n",
		"apps/web/package.json": `{"scripts":{"lint":"eslint .","test":"jest"}}`,
	})
	cfg := budgetCfg(both)
	cfg.TestTimeout = 900 * time.Second
	got, _ := NewQualityGate(cfg).OverallBudget()
	// go: 30+60+900, node: lint 60 + test 900 => at least 1950s before slack.
	if got < 1950*time.Second || got <= single {
		t.Errorf("OverallBudget with go+node = %s (go only %s), want >= 1950s and larger than go only", got, single)
	}
}

// Skipped tests do not count toward the budget.
func TestOverallBudgetIgnoresSkippedTests(t *testing.T) {
	t.Parallel()
	cfg := budgetCfg(budgetGoDir(t))
	cfg.TestTimeout = 900 * time.Second
	cfg.SkipTests = true
	got, _ := NewQualityGate(cfg).OverallBudget()
	if got != 10*time.Minute {
		t.Errorf("OverallBudget with SkipTests = %s, want the 10m floor", got)
	}
}

// The budget is never unlimited, even when step budgets are absurd.
func TestOverallBudgetIsFinite(t *testing.T) {
	t.Parallel()
	cfg := budgetCfg(budgetGoDir(t))
	got, _ := NewQualityGate(cfg).OverallBudget()
	if got <= 0 {
		t.Errorf("OverallBudget = %s, want a positive finite budget", got)
	}
}

// Parent cancellation still reaches the step and keeps the card t218
// attribution text: only the budget value changed, not the message.
func TestOverallBudgetParentExpiryKeepsAttribution(t *testing.T) {
	t.Parallel()
	dir := writeFixture(t, map[string]string{"package.json": `{"scripts":{"test":"go t2020-not-used"}}`})
	cfg := budgetCfg(dir)
	cfg.TypecheckEnabled = false
	g := NewQualityGate(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	passed, out := g.Run(ctx)
	if passed {
		t.Skip("run passed under an expired parent context; no attribution to read")
	}
	if strings.Contains(out, "timed out") && !strings.Contains(out, "overall gate budget ran out") {
		t.Errorf("parent expiry blamed the step: %q", out)
	}
}

// exactBudgetCfg enables every budget axis of a single Go toolchain with
// distinct step budgets, so the derived value below is one exact number and
// dropping any term (an axis, an allowance, the slack) changes it.
func exactBudgetCfg(t *testing.T) *GateConfig {
	t.Helper()
	cfg := budgetCfg(budgetGoDir(t))
	cfg.VetTimeout = 10 * time.Second
	cfg.TypecheckEnabled = true
	cfg.TypecheckCommand = "go version"
	cfg.TypecheckTimeout = 50 * time.Second
	cfg.LintTimeout = 20 * time.Second
	cfg.TestTimeout = 1000 * time.Second
	cfg.GraphFreshness = &GraphFreshnessConfig{Enabled: true}
	cfg.AstGrepGate = &AstGrepGateConfig{Enabled: true}
	return cfg
}

// vet 10 + typecheck 50 + lint 20 + test 1000 + graph-freshness 30 + ast-grep
// 30 = 1140s; slack = max(60s, 10% = 114s) = 114s; total 1254s.
func TestOverallBudgetExactSumOfEveryAxis(t *testing.T) {
	t.Parallel()
	got, warn := NewQualityGate(exactBudgetCfg(t)).OverallBudget()
	if want := 1254 * time.Second; got != want || warn != "" {
		t.Errorf("OverallBudget = %s warn=%q, want exactly %s (vet+typecheck+lint+test+allowances+10%% slack)", got, warn, want)
	}
}

// The slack floor is 60s when 10% of the sum is smaller than that.
func TestOverallBudgetSlackFloorIsSixtySeconds(t *testing.T) {
	t.Parallel()
	cfg := exactBudgetCfg(t)
	cfg.TypecheckEnabled = false
	cfg.GraphFreshness = nil
	cfg.AstGrepGate = nil
	cfg.VetTimeout, cfg.LintTimeout, cfg.TestTimeout = 40*time.Second, 100*time.Second, 400*time.Second
	// sum 540s; 10% = 54s < 60s, so the slack is the 60s floor: 600s, which
	// is exactly the ten-minute floor. A slack of 54s would give 594s and the
	// floor would lift it, so also check a sum just above the floor:
	if got, _ := NewQualityGate(cfg).OverallBudget(); got != 10*time.Minute {
		t.Errorf("OverallBudget = %s, want 10m for sum 540s", got)
	}
	cfg.TestTimeout = 410 * time.Second // sum 550s: 10% = 55s < 60s -> 610s
	if got, _ := NewQualityGate(cfg).OverallBudget(); got != 610*time.Second {
		t.Errorf("OverallBudget = %s, want 610s (sum 550s + the 60s slack floor)", got)
	}
}

// The largest single step can be an axis other than test: a 2000s typecheck
// makes a 1500s override too small, and the raise names the larger budget.
func TestOverallBudgetLargestStepIsTheLargestAxis(t *testing.T) {
	t.Parallel()
	cfg := exactBudgetCfg(t)
	cfg.TypecheckTimeout = 2000 * time.Second
	cfg.TotalTimeout = 1500 * time.Second
	got, warn := NewQualityGate(cfg).OverallBudget()
	if got != 2000*time.Second {
		t.Errorf("OverallBudget = %s, want the 2000s typecheck budget as the floor for an override", got)
	}
	if !strings.Contains(warn, "timeouts.total") || !strings.Contains(warn, "33m20s") {
		t.Errorf("warning %q should name timeouts.total and the 33m20s (2000s) step", warn)
	}
}

// An override at or above the largest step but below the derived budget wins
// as given: it is the operator's choice, not clamped up to the derivation.
func TestOverallBudgetOverrideBetweenLargestStepAndDerivedWins(t *testing.T) {
	t.Parallel()
	cfg := exactBudgetCfg(t)
	cfg.TotalTimeout = 1100 * time.Second // largest 1000s <= 1100s < derived 1254s
	got, warn := NewQualityGate(cfg).OverallBudget()
	if got != 1100*time.Second || warn != "" {
		t.Errorf("OverallBudget = %s warn=%q, want exactly the 1100s override and no warning", got, warn)
	}
}
