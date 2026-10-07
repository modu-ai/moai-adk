package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// deadlineHandler returns exactly the sentinel the review-gate finding
// names: a handler whose work hits its own deadline surfaces
// context.DeadlineExceeded as its error.
type deadlineHandler struct{ event EventType }

func (h *deadlineHandler) EventType() EventType { return h.event }
func (h *deadlineHandler) Handle(_ context.Context, _ *HookInput) (*HookOutput, error) {
	return nil, context.DeadlineExceeded
}

// TestHandlerTimeoutCapturedAsAmbiguous pins the end-to-end shape of the
// review-gate P2 finding: a handler returning a timeout sentinel must
// produce an AMBIGUOUS, locally-retained spool entry under the
// hook_timeout kind — never a discarded environment classification. Before
// the fix the registry routed the kind correctly but Attribute's A2
// net.Error row outranked it (context.DeadlineExceeded implements
// net.Error), Capture classified environment, and the signal was dropped
// without a trace — the SPEC's hook_timeout retention (DEC-7) never
// happened.
func TestHandlerTimeoutCapturedAsAmbiguous(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	consent := "participation:\n  enabled: true\n  asked: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(consent), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistryWithTimeout(cfg, 50*time.Millisecond)
	reg.Register(&deadlineHandler{event: EventSessionStart})

	_, _ = reg.Dispatch(context.Background(), EventSessionStart, &HookInput{})

	entries, err := bugreport.ReadSpool()
	if err != nil {
		t.Fatalf("read spool: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("spool holds %d entrie(s), want exactly the retained timeout signal: %v", len(entries), entries)
	}
	if entries[0].Kind != bugreport.KindHookTimeout {
		t.Fatalf("spooled kind = %s, want hook_timeout", entries[0].Kind)
	}
	if entries[0].Verdict != bugreport.VerdictAmbiguous {
		t.Fatalf("spooled verdict = %s, want ambiguous (locally retained, never a discarded environment)", entries[0].Verdict)
	}
}
