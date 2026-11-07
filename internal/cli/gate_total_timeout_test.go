package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// Card t2020: gate.yaml timeouts.total must reach the quality gate as its
// overall-budget override; unset or non-positive must leave it at zero so the
// budget is derived from the steps.
func TestMapConfigGateToQualityCarriesTotalTimeout(t *testing.T) {
	t.Parallel()
	cases := []struct {
		total int
		want  time.Duration
	}{
		{0, 0},
		{-30, 0},
		{1500, 1500 * time.Second},
	}
	for _, c := range cases {
		g := config.GateConfig{Timeouts: config.GateTimeouts{Total: c.total}}
		if got := mapConfigGateToQuality(g).TotalTimeout; got != c.want {
			t.Errorf("timeouts.total=%d: TotalTimeout = %s, want %s", c.total, got, c.want)
		}
	}
}

func gateRunFixtureCfg(t *testing.T) *quality.GateConfig {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{
		"go.mod":  "module example.com/t2020cli\n\ngo 1.22\n",
		"main.go": "package main\nfunc main() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := quality.DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.AstGrepGate = nil
	cfg.GraphFreshness = nil
	cfg.TypecheckEnabled = false
	return cfg
}

// The deadline the CLI applies to a whole gate run is the derived budget: a
// 900s test step must not be capped at the historical ten minutes.
func TestGateRunContextAppliesTheDerivedBudget(t *testing.T) {
	t.Parallel()
	cfg := gateRunFixtureCfg(t)
	cfg.TestTimeout = 900 * time.Second
	var errOut bytes.Buffer
	ctx, cancel := gateRunContext(quality.NewQualityGate(cfg), &errOut)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("the gate run context has no deadline (unbounded)")
	}
	remaining := time.Until(deadline)
	if remaining <= 900*time.Second || remaining > 1300*time.Second {
		t.Errorf("deadline in %s, want the derived budget above the 900s test step (and finite, <= 1300s)", remaining)
	}
	if errOut.Len() != 0 {
		t.Errorf("a derived budget wrote to stderr: %q", errOut.String())
	}
}

// An override below the largest single step is raised to that step, and the
// correction is announced on stderr in one line, never applied silently.
func TestGateRunContextAnnouncesARaisedOverride(t *testing.T) {
	t.Parallel()
	cfg := gateRunFixtureCfg(t)
	cfg.TestTimeout = 900 * time.Second
	cfg.TotalTimeout = 100 * time.Second
	var errOut bytes.Buffer
	ctx, cancel := gateRunContext(quality.NewQualityGate(cfg), &errOut)
	defer cancel()

	deadline, _ := ctx.Deadline()
	if remaining := time.Until(deadline); remaining <= 800*time.Second || remaining > 900*time.Second {
		t.Errorf("deadline in %s, want the raised budget of the 900s step", remaining)
	}
	got := strings.TrimSuffix(errOut.String(), "\n")
	if !strings.Contains(got, "timeouts.total") || strings.Contains(got, "\n") {
		t.Errorf("want one stderr line naming timeouts.total, got %q", errOut.String())
	}
}

// An explicit override at or above the largest step is used as given.
func TestGateRunContextHonoursAnOverride(t *testing.T) {
	t.Parallel()
	cfg := gateRunFixtureCfg(t)
	cfg.TestTimeout = 900 * time.Second
	cfg.TotalTimeout = 2000 * time.Second
	var errOut bytes.Buffer
	ctx, cancel := gateRunContext(quality.NewQualityGate(cfg), &errOut)
	defer cancel()

	deadline, _ := ctx.Deadline()
	if remaining := time.Until(deadline); remaining <= 1900*time.Second || remaining > 2000*time.Second {
		t.Errorf("deadline in %s, want about the 2000s override", remaining)
	}
	if errOut.Len() != 0 {
		t.Errorf("an honoured override wrote to stderr: %q", errOut.String())
	}
}
