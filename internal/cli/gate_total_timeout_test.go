package cli

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
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
