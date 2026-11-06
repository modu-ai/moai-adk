package cli

// init_resume_settle_test.go — review finding 2 (card t1591): a completed
// resume must end the template-snapshot lifecycle the way a healthy init does.
//
// Healthy init settles the settings.json and .mcp.json staging copies
// (SettleSettingsSnapshot / SettleMCPSnapshot, init.go). The resume rewrites
// both files but never settled them, so a staging copy the failing run left
// behind was left to the next run's leftover judgement instead of being
// settled by the resume that completed the install.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// stagedSnapshots maps each staging copy that exists now to its canonical base.
func stagedSnapshots(projectDir string) map[string]string {
	pairs := map[string]string{
		backup.SettingsSnapshotPendingPath(projectDir): backup.SettingsSnapshotPath(projectDir),
		backup.MCPSnapshotPendingPath(projectDir):      backup.MCPSnapshotPath(projectDir),
	}
	for pending := range pairs {
		if _, err := os.Stat(pending); err != nil {
			delete(pairs, pending)
		}
	}
	return pairs
}

// TestInitResumeSettlesTemplateSnapshots — finding 2. The seed run staged the
// settings and MCP renders its deploy wrote. The resume itself (called directly,
// so no leftover judgement runs first) must settle both: no staging copy
// remains, and each canonical base holds the render that was staged.
func TestInitResumeSettlesTemplateSnapshots(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "settle-resume-proj")
	resumeSeedRun(t, projectDir, "claude")

	staged := stagedSnapshots(projectDir)
	if len(staged) != 2 {
		t.Fatalf("precondition: the seed run must stage both snapshot copies (settings and MCP), staged %d", len(staged))
	}
	renders := make(map[string][]byte, len(staged))
	for pending := range staged {
		data, err := os.ReadFile(pending)
		if err != nil {
			t.Fatalf("read staging copy %s: %v", pending, err)
		}
		renders[pending] = data
	}

	opts := project.InitOptions{
		ProjectRoot:  projectDir,
		DeployMode:   "local",
		AutonomyTier: config.AutonomyTierAutomatic,
		MCPProvision: true,
	}
	cmd := newInitTestCmd()
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := resumeInitializedProject(cmd, &opts, false, nil, profile.ProfilePreferences{}); err != nil {
		t.Fatalf("the resume failed: %v", err)
	}

	for pending, canonical := range staged {
		if _, err := os.Stat(pending); err == nil {
			t.Errorf("the staging copy %s outlived the completed resume: the resume never settled it", pending)
		}
		got, err := os.ReadFile(canonical)
		if err != nil {
			t.Errorf("the completed resume did not promote the staged render to %s: %v", canonical, err)
			continue
		}
		if string(got) != string(renders[pending]) {
			t.Errorf("the promoted base %s does not hold the render that was staged", canonical)
		}
	}
}

// TestInitResumeEndToEndLeavesNoStagingCopy — the end-to-end resume (the second
// runInit) leaves no staging copy behind. The leftover judgement that runs at
// the start of every init settles the copies the seed run staged, so this pins
// the observable outcome; TestInitResumeSettlesTemplateSnapshots pins the
// resume's own completion.
func TestInitResumeEndToEndLeavesNoStagingCopy(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "settle-e2e-proj")
	resumeSeedRun(t, projectDir, "claude")
	if n := len(stagedSnapshots(projectDir)); n != 2 {
		t.Fatalf("precondition: the seed run must stage both snapshot copies, staged %d", n)
	}

	if err := resumeRun(t, projectDir, ""); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	for _, pending := range []string{
		backup.SettingsSnapshotPendingPath(projectDir),
		backup.MCPSnapshotPendingPath(projectDir),
	} {
		if _, err := os.Stat(pending); err == nil {
			t.Errorf("the staging copy %s survived the end-to-end resume", pending)
		}
	}
}
