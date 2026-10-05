package hook

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// TestSessionStartHygieneBestEffort — AC-HYG-013 (L-013). With the hygiene
// engine stubbed to fail, the hook still exits allow, the failure appears
// only as a logged warning, and the remaining SessionStart steps complete.
func TestSessionStartHygieneBestEffort(t *testing.T) {
	// Not parallel: the seam swap is package-level (same pattern as the
	// drift cache seam tests).
	t.Setenv("ANTHROPIC_BASE_URL", "")
	clearFactoryEnv(t)

	orig := hygieneRunFn
	hygieneRunFn = func(string) error { return errors.New("injected hygiene failure") }
	t.Cleanup(func() { hygieneRunFn = orig })

	c := newTestConfig()
	c.Project = models.ProjectConfig{
		Name:     "moai-adk-go",
		Type:     models.ProjectTypeCLI,
		Language: "go",
	}
	cfg := &mockConfigProvider{cfg: c}
	h := NewSessionStartHandler(cfg)

	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-hygiene-1",
		CWD:           t.TempDir(),
		ProjectDir:    t.TempDir(),
		HookEventName: "SessionStart",
	})
	if err != nil {
		t.Fatalf("Handle returned an error on hygiene failure: %v", err)
	}
	if out == nil {
		t.Fatalf("Handle returned no output")
	}
	if out.Decision != "" && out.Decision != DecisionAllow {
		t.Fatalf("decision = %q, want allow (hygiene must never block a launch)", out.Decision)
	}
	if out.Continue != nil && !*out.Continue && out.StopReason != "" {
		t.Fatalf("launch blocked by hygiene failure: %s", out.StopReason)
	}
	// The remaining SessionStart steps completed: the session id and status
	// landed in the Data payload.
	blob, mErr := json.Marshal(out)
	if mErr != nil {
		t.Fatalf("marshal output: %v", mErr)
	}
	if !strings.Contains(string(blob), "sess-hygiene-1") {
		t.Fatalf("session id missing from the output — remaining steps did not complete: %s", blob)
	}
}
