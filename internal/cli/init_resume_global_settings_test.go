package cli

// init_resume_global_settings_test.go — the resume runs the global settings
// cleanup the healthy init runs (card t1591, final-gate finding P2-2).
//
// The healthy init calls ensureGlobalSettingsEnv after the deploy-mode record
// (init.go), and that call removes the legacy global MoAI hooks directory. The
// failed attempt never reached it, and the resume returned without making the
// call, so an interrupted install that resumed kept those hooks alongside the
// project hooks.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitResumeRunsGlobalSettingsCleanup seeds the failing first run, plants a
// legacy global hooks directory under the test HOME, and resumes. The resume
// must remove that directory exactly as the healthy init does.
func TestInitResumeRunsGlobalSettingsCleanup(t *testing.T) {
	resumeHome(t)
	projectDir := t.TempDir()
	resumeSeedRun(t, projectDir, "claude")

	legacyHooks := filepath.Join(os.Getenv("HOME"), ".claude", "hooks", "moai")
	if err := os.MkdirAll(legacyHooks, 0o755); err != nil {
		t.Fatalf("seed the legacy global hooks directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyHooks, "legacy-hook.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("seed the legacy hook script: %v", err)
	}

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if _, err := os.Stat(legacyHooks); !os.IsNotExist(err) {
		t.Errorf("the resume left the legacy global hooks directory in place (stat err = %v); the healthy init removes it through ensureGlobalSettingsEnv, and a resumed install must do the same", err)
	}
}
