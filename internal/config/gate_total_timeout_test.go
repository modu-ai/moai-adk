package config

import (
	"testing"
	"time"
)

// Card t2020: timeouts.total is the explicit overall-budget override.
func TestGateTotalTimeoutDuration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		secs int
		want time.Duration
	}{
		{0, 0},
		{-1, 0},
		{1500, 1500 * time.Second},
	}
	for _, c := range cases {
		g := GateConfig{Timeouts: GateTimeouts{Total: c.secs}}
		if got := g.TotalTimeoutDuration(); got != c.want {
			t.Errorf("Total=%d: TotalTimeoutDuration() = %s, want %s", c.secs, got, c.want)
		}
	}
}
