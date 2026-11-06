package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestRelaunchRefusesCodexProvider — card t1909: the general codex launcher
// was removed, so the relaunch verb must refuse --provider codex at the shared
// builder instead of emitting a launch line nothing executes. The refusal
// names the surviving recovery providers.
func TestRelaunchRefusesCodexProvider(t *testing.T) {
	_, err := buildRelaunchCommand(factory.RelaunchProviderCodex, "", "", "")
	if err == nil {
		t.Fatal("buildRelaunchCommand(codex) = nil error, want the removal refusal")
	}
	for _, want := range []string{"no longer available", factory.RelaunchProviderCC, factory.RelaunchProviderGLM} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name %q", err.Error(), want)
		}
	}
	// The surviving providers are untouched: they still build.
	for _, p := range []string{factory.RelaunchProviderCC, factory.RelaunchProviderGLM} {
		if _, err := buildRelaunchCommand(p, "", "", ""); err != nil {
			t.Errorf("buildRelaunchCommand(%s) = %v, want nil", p, err)
		}
	}
}
