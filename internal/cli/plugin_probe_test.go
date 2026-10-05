package cli

// plugin_probe_test.go — the post-install list-surface probe's arms
// (SPEC-INIT-SHRINK-001 design §2.4, plan M2). Every read goes through the
// injected fake runner; no test reaches a real tool (REQ-020).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// claudeListFake builds a fake whose `claude plugin list` reads answer per
// the given sequence (the pre-execution snapshot is the first list call,
// the post-execution read the second); every other call succeeds silently.
func claudeListFake(t *testing.T, reads []string) *fakePluginRunner {
	t.Helper()
	fake := &fakePluginRunner{}
	n := 0
	fake.fn = func(_ context.Context, call pluginCall, _ int) ([]byte, error) {
		if filepath.Base(call.bin) == "claude-pinned" && len(call.args) >= 2 &&
			call.args[0] == "plugin" && call.args[1] == "list" {
			if n >= len(reads) {
				n++
				return []byte("(no plugins installed)"), nil
			}
			out := reads[n]
			n++
			return []byte(out), nil
		}
		return nil, nil
	}
	return fake
}

// TestProbeOptedOutDecidedBeforeTheStep: the opt-out arm takes no snapshot
// (no list call reaches the runner) and returns opted-out.
func TestProbeOptedOutDecidedBeforeTheStep(t *testing.T) {
	setupPluginTools(t)
	fake := &fakePluginRunner{}
	withPluginRunner(t, fake)

	out := runInitPluginInstallProbed(&strings.Builder{}, agentWiringClaude, t.TempDir(), true)
	if out != probeOutcomeOptedOut {
		t.Fatalf("outcome = %q, want opted-out", out)
	}
	for _, v := range fake.vectors() {
		if strings.Contains(v, "plugin list") {
			t.Fatalf("opted-out arm probed the list surface: %v", fake.vectors())
		}
	}
}

// TestProbeEnvOptOutMatchesFlagOptOut: MOAI_SKIP_PLUGIN_INSTALL=1 behaves
// like --no-plugin (the t1435 OD-5 pin).
func TestProbeEnvOptOutMatchesFlagOptOut(t *testing.T) {
	setupPluginTools(t)
	t.Setenv(config.EnvSkipPluginInstall, "1")
	fake := &fakePluginRunner{}
	withPluginRunner(t, fake)

	out := runInitPluginInstallProbed(&strings.Builder{}, agentWiringClaude, t.TempDir(), false)
	if out != probeOutcomeOptedOut {
		t.Fatalf("outcome = %q, want opted-out", out)
	}
}

// TestProbeConfirmedRequiresAbsentThenPresent: the ref absent from the
// pre-snapshot and present post-run is the ONLY confirmed diff.
func TestProbeConfirmedRequiresAbsentThenPresent(t *testing.T) {
	setupPluginTools(t)
	withPluginRunner(t, claudeListFake(t, []string{"(none)", "moai@moai-adk 1.0.0"}))

	out := runInitPluginInstallProbed(&strings.Builder{}, agentWiringClaude, t.TempDir(), false)
	if out != probeOutcomeConfirmed {
		t.Fatalf("outcome = %q, want confirmed", out)
	}
}

// TestProbePreExistingPluginIsNotDemonstrated: present in BOTH reads — a
// pre-existing installation's diff never demonstrates THIS run's install.
func TestProbePreExistingPluginIsNotDemonstrated(t *testing.T) {
	setupPluginTools(t)
	withPluginRunner(t, claudeListFake(t, []string{"moai@moai-adk 0.9.0", "moai@moai-adk 1.0.0"}))

	out := runInitPluginInstallProbed(&strings.Builder{}, agentWiringClaude, t.TempDir(), false)
	if out != probeOutcomeNotDemonstrated {
		t.Fatalf("outcome = %q, want not-demonstrated", out)
	}
}

// TestProbeUnreadableSurfaceIsNotDemonstrated: every probe error — a failed
// pre-execution read here — resolves to not-demonstrated, even though the
// post read alone would look like a fresh install.
func TestProbeUnreadableSurfaceIsNotDemonstrated(t *testing.T) {
	setupPluginTools(t)
	fake := &fakePluginRunner{}
	lists := 0
	fake.fn = func(_ context.Context, call pluginCall, _ int) ([]byte, error) {
		if filepath.Base(call.bin) == "claude-pinned" && len(call.args) >= 2 &&
			call.args[0] == "plugin" && call.args[1] == "list" {
			lists++
			if lists == 1 {
				return nil, context.DeadlineExceeded // the pre-read timed out
			}
			return []byte("moai@moai-adk 1.0.0"), nil
		}
		return nil, nil
	}
	withPluginRunner(t, fake)

	out := runInitPluginInstallProbed(&strings.Builder{}, agentWiringClaude, t.TempDir(), false)
	if out != probeOutcomeNotDemonstrated {
		t.Fatalf("outcome = %q, want not-demonstrated", out)
	}
}

// TestProbeCodexArmReadsTheJSONSurface: the Codex surface read parses the
// doctor's installed[].pluginId JSON shape.
func TestProbeCodexArmReadsTheJSONSurface(t *testing.T) {
	e := setupPluginTools(t)
	if err := os.WriteFile(e.codexBin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write codex fake: %v", err)
	}
	fake := &fakePluginRunner{}
	lists := 0
	fake.fn = func(_ context.Context, call pluginCall, _ int) ([]byte, error) {
		if filepath.Base(call.bin) == "codex-fake" && len(call.args) >= 2 &&
			call.args[0] == "plugin" && call.args[1] == "list" {
			lists++
			if lists >= 2 {
				return []byte(`{"installed":[{"pluginId":"moai@moai-adk","version":"1.0.0"}]}`), nil
			}
			return []byte(`{"installed":[]}`), nil
		}
		return nil, nil
	}
	withPluginRunner(t, fake)

	out := runInitPluginInstallProbed(&strings.Builder{}, agentWiringGPT, t.TempDir(), false)
	if out != probeOutcomeConfirmed {
		t.Fatalf("outcome = %q, want confirmed (codex arm)", out)
	}
}
