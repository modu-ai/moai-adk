// monitor_source.go — connects the monitor package to the web panels
// (Issue #1948). MonitorRowsFrom turns a read-only monitor.Collector into the
// monitorRows source: it collects the View and encodes it with the monitor's
// own MarshalView, so the web decodes exactly what the monitor emits. Nothing
// here is stored; every refresh reads again.
package web

import (
	"context"
	"errors"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// MonitorRowsFrom builds a MonitorRows source. issues lists the Issues to show
// on each read. A nil collector or issue list yields a source that fails, which
// the panels render as "not provided".
func MonitorRowsFrom(c *monitor.Collector, issues func(context.Context) []int) func(context.Context) ([]byte, error) {
	return func(ctx context.Context) ([]byte, error) {
		if c == nil || issues == nil {
			return nil, errors.New("monitor source is not configured")
		}
		return monitor.MarshalView(c.View(ctx, issues(ctx)))
	}
}
