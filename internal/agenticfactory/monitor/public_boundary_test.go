package monitor_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// The two helpers the web record rows need, used from outside the package: one field masked the way the
// monitor masks its own rows, and the backlog classification the monitor shows.
func TestMaskFieldIsTheMonitorsOwnMaskingForOneValue(t *testing.T) {
	const secret = "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	got, redacted := monitor.MaskField("rejected token " + secret)
	if !redacted || strings.Contains(got, secret) || !strings.Contains(got, "[REDACTED]") {
		t.Errorf("a known credential shape: %q redacted=%v", got, redacted)
	}
	if got, redacted := monitor.MaskField("build failed on head"); got != "build failed on head" || redacted {
		t.Errorf("plain text must come back unchanged: %q redacted=%v", got, redacted)
	}
	// text that still looks like an unknown credential is dropped altogether, not shown partly
	if got, redacted := monitor.MaskField("see Zk3fQ9xLmN2vB7cD4hJ8sT1wY5pR6aE0uG"); got != "" || !redacted {
		t.Errorf("an opaque run: %q redacted=%v, want it dropped", got, redacted)
	}
	if got, redacted := monitor.MaskField(strings.Repeat("a", 40)); got != strings.Repeat("a", 40) || redacted {
		t.Errorf("a full commit SHA is not a credential: %q redacted=%v", got, redacted)
	}
}

func TestBacklogStateOfClassifiesTheCountAgainstTheLimit(t *testing.T) {
	for _, c := range []struct {
		open, limit int
		want        monitor.BacklogState
	}{
		{3, 0, monitor.BacklogNoLimit}, {3, -1, monitor.BacklogNoLimit},
		{2, 3, monitor.BacklogWithin}, {3, 3, monitor.BacklogAtLimit}, {4, 3, monitor.BacklogOver},
	} {
		if got := monitor.BacklogStateOf(c.open, c.limit); got != c.want {
			t.Errorf("open=%d limit=%d: %q, want %q", c.open, c.limit, got, c.want)
		}
	}
}
