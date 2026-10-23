package cli

// t1619_init_resume_coverage_test.go — card t1619, bundle B1: the unexecuted lines
// of init_resume.go. Each test drives one failure branch (or one absent/empty
// branch of a helper) and asserts the observed outcome. The direct tests call the
// unexported functions with a real temp home and project; the flow tests seed an
// interrupted init with resumeSeedRun, plant one obstacle, and resume through
// runInit — the route a production resume takes.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// t1619ResumeHome points HOME and the user-home seam at a fresh temp home.
func t1619ResumeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	orig := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = orig })
	userHomeDirFn = func() (string, error) { return home, nil }
	return home
}

// t1619ResumeCall runs resumeInitializedProject directly, with no wizard answer.
func t1619ResumeCall(t *testing.T, projectRoot string) error {
	t.Helper()
	cmd := newInitTestCmd()
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	opts := &project.InitOptions{ProjectRoot: projectRoot}
	return resumeInitializedProject(cmd, opts, false, nil, profile.ProfilePreferences{})
}

// t1619ResumeRunCapture runs init on projectDir like resumeRun and also returns the
// text written to the command's error stream, where the resume's warnings go. The
// test command defaults to --no-hooks; withHooks clears it so the git hook steps run.
func t1619ResumeRunCapture(t *testing.T, projectDir string, withHooks bool) (string, error) {
	t.Helper()
	cmd := newInitTestCmd()
	if withHooks {
		if err := cmd.Flags().Set("no-hooks", "false"); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := runInit(cmd, []string{projectDir})
	return errOut.String(), err
}

// t1619WizardAnswers makes the resume's wizard return the given answers.
func t1619WizardAnswers(t *testing.T, res wizard.WizardResult) {
	t.Helper()
	prev := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		answer := res
		return &answer, nil
	}
	t.Cleanup(func() { runWizardFn = prev })
}

// t1619SkipWithoutFilePerms skips the permission-based tests where the process cannot
// be refused a write: root bypasses file modes, and Windows does not apply them the same way.
func t1619SkipWithoutFilePerms(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the write refusal is a unix file-mode condition")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file modes, so the write refusal cannot be produced")
	}
}

// t1619MustMkdir creates dir and its parents, failing the test on error.
func t1619MustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

// t1619MustWrite writes a file, failing the test on error.
func t1619MustWrite(t *testing.T, path, content string) {
	t.Helper()
	t1619MustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// t1619MarkerPath is the resume checkpoint marker of a project.
func t1619MarkerPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "resume-pending")
}

// TestInitResumeHomeProbeReadsUnresolvableHomeAsEmpty — an unresolvable user home
// reads as the empty home, the "no checkpoint" answer the ordinary re-run relies on.
func TestInitResumeHomeProbeReadsUnresolvableHomeAsEmpty(t *testing.T) {
	orig := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = orig })
	userHomeDirFn = func() (string, error) { return "", errors.New("no home directory") }
	if got := userHomeDirOrEmpty(); got != "" {
		t.Fatalf("an unresolvable home must read as the empty home, got %q", got)
	}
}

// TestInitResumeHomeFailureIsReturnedBeforeAnyWrite — a resume that cannot resolve the
// user home fails at its first step and leaves no checkpoint behind.
func TestInitResumeHomeFailureIsReturnedBeforeAnyWrite(t *testing.T) {
	orig := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = orig })
	userHomeDirFn = func() (string, error) { return "", errors.New("no home directory") }
	projectRoot := t.TempDir()
	err := t1619ResumeCall(t, projectRoot)
	if err == nil || !strings.Contains(err.Error(), "resolve user home") {
		t.Fatalf("an unresolvable home must fail the resume at the home step, got %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectRoot)); !os.IsNotExist(statErr) {
		t.Fatalf("the failed home step left a resume marker behind (stat err = %v)", statErr)
	}
}

// TestInitResumeLockFailureIsReturnedBeforeAnyMarker — a user lock that cannot be
// taken fails the resume before the marker is written. A regular file where the lock's
// directory must be makes the acquisition fail at once, without the two-minute wait.
func TestInitResumeLockFailureIsReturnedBeforeAnyMarker(t *testing.T) {
	home := t1619ResumeHome(t)
	t1619MustWrite(t, filepath.Join(home, ".moai"), "not a directory\n")
	projectRoot := t.TempDir()
	err := t1619ResumeCall(t, projectRoot)
	if err == nil || !strings.Contains(err.Error(), "user-asset resume lock") {
		t.Fatalf("a lock that cannot be taken must fail the resume at the lock step, got %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectRoot)); !os.IsNotExist(statErr) {
		t.Fatalf("the failed lock step left a resume marker behind (stat err = %v)", statErr)
	}
}

// TestInitResumeMarkerWriteFailureIsReturnedAndReleasesLock — a marker that cannot be
// written fails the resume, and the user lock taken before it is released on that path.
func TestInitResumeMarkerWriteFailureIsReturnedAndReleasesLock(t *testing.T) {
	home := t1619ResumeHome(t)
	projectRoot := t.TempDir()
	// A directory where the marker file must be makes the marker write fail.
	t1619MustMkdir(t, t1619MarkerPath(projectRoot))
	err := t1619ResumeCall(t, projectRoot)
	if err == nil || !strings.Contains(err.Error(), "write resume marker") {
		t.Fatalf("an unwritable marker must fail the resume at the marker step, got %v", err)
	}
	if _, statErr := os.Stat(userassets.LockPath(home)); !os.IsNotExist(statErr) {
		t.Fatalf("the user lock was not released on the marker failure (stat err = %v)", statErr)
	}
}

// TestInitResumeSelectionRecordFailureIsReturned — a bundle-selection record that cannot
// be persisted fails the resume after the marker is in place.
func TestInitResumeSelectionRecordFailureIsReturned(t *testing.T) {
	t1619ResumeHome(t)
	projectRoot := t.TempDir()
	// The marker write succeeds inside .moai; a regular file where .moai/state must be
	// makes the selection record fail.
	t1619MustWrite(t, filepath.Join(projectRoot, ".moai", "state"), "not a directory\n")
	err := t1619ResumeCall(t, projectRoot)
	if err == nil || !strings.Contains(err.Error(), "record bundle selection") {
		t.Fatalf("an unwritable selection record must fail the resume, got %v", err)
	}
}

// TestInitResumeManifestReadFailureIsReturnedAndKeepsMarker — a user manifest that cannot
// be read fails the ensure step; the checkpoint marker written before it stays in place,
// so the next run still has a resume route.
func TestInitResumeManifestReadFailureIsReturnedAndKeepsMarker(t *testing.T) {
	home := t1619ResumeHome(t)
	// A directory at the manifest path: the quarantine leaves a non-file alone, and the
	// ensure's manifest read fails.
	t1619MustMkdir(t, userassets.ManifestPath(home))
	projectRoot := t.TempDir()
	// A real init has created the project's .moai directory before the resume runs.
	t1619MustMkdir(t, filepath.Join(projectRoot, ".moai"))
	err := t1619ResumeCall(t, projectRoot)
	if err == nil || !strings.Contains(err.Error(), "user-asset ensure") {
		t.Fatalf("an unreadable user manifest must fail the ensure step, got %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectRoot)); statErr != nil {
		t.Fatalf("the checkpoint marker did not stay in place after the ensure failure: %v", statErr)
	}
}

// TestInitResumeInstallerBuildsFromEmbeddedCatalog — evidence for the unreachable
// error return after newUserAssetInstaller: its only failure sources are the catalog
// and templates compiled into the binary, and they load without error.
func TestInitResumeInstallerBuildsFromEmbeddedCatalog(t *testing.T) {
	inst, err := newUserAssetInstaller(t.TempDir())
	if err != nil || inst == nil {
		t.Fatalf("the installer built from the embedded catalog must not fail (err = %v)", err)
	}
}

// TestInitResumeAutonomyTierFailureKeepsMarker — the autonomy-tier step's failure is
// returned, and the marker stays so the next run resumes the step.
func TestInitResumeAutonomyTierFailureKeepsMarker(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "autonomy-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	orig := applyAutonomyTierBundleFn
	t.Cleanup(func() { applyAutonomyTierBundleFn = orig })
	applyAutonomyTierBundleFn = func(_, _, _, _ string) error {
		return errors.New("simulated autonomy tier write failure")
	}
	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "apply autonomy tier") {
		t.Fatalf("the autonomy-tier failure did not surface as the resume's error: %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the autonomy-tier failure: %v", statErr)
	}
}

// TestInitResumeJevWriteFailureKeepsMarker — a Jev answer whose section write is refused
// (a malformed workflow section) fails the resume, and the marker stays.
func TestInitResumeJevWriteFailureKeepsMarker(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "jev-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	// The patch refuses malformed YAML, so an enabling write fails here.
	t1619MustWrite(t, filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n  jev: [unclosed\n")
	t1619WizardAnswers(t, wizard.WizardResult{JevEnabled: true, AutonomyTier: config.AutonomyTierAutomatic})
	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "apply jev selection") {
		t.Fatalf("the refused Jev write did not surface as the resume's error: %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the Jev failure: %v", statErr)
	}
}

// TestInitResumeParticipationWriteFailureKeepsMarker — a participation answer whose
// consent file cannot be patched (malformed YAML) fails the resume, and the marker stays.
func TestInitResumeParticipationWriteFailureKeepsMarker(t *testing.T) {
	resumeHome(t)
	t.Setenv("CI", "")
	t.Setenv("MOAI_HOME", "")
	projectDir := filepath.Join(t.TempDir(), "participation-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	consentPath, err := config.UserParticipationFilePath()
	if err != nil {
		t.Fatal(err)
	}
	t1619MustWrite(t, consentPath, "participation: [unclosed\n")
	t1619WizardAnswers(t, wizard.WizardResult{ParticipationEnabled: true, AutonomyTier: config.AutonomyTierAutomatic})
	err = resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "apply participation selection") {
		t.Fatalf("the refused participation write did not surface as the resume's error: %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the participation failure: %v", statErr)
	}
}

// TestInitResumeHarnessWriteFailureKeepsMarker — a harness record that must change but
// whose llm.yaml is read-only fails the resume with the harness step's error. Needs a
// non-root process: root writes through file modes.
func TestInitResumeHarnessWriteFailureKeepsMarker(t *testing.T) {
	t1619SkipWithoutFilePerms(t)
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "harness-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	llmPath := filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml")
	// The recorded harness is claude; a llm.yaml naming codex forces the harness write.
	t1619MustWrite(t, llmPath, "llm:\n  harness: codex\n")
	if err := os.Chmod(llmPath, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(llmPath, 0o644) })

	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "apply harness") {
		t.Fatalf("the refused harness write did not surface as the resume's error: %v", err)
	}
	if got := readResumeLLMYAML(t, projectDir); !strings.Contains(got, "harness: codex") {
		t.Fatalf("the refused harness write changed llm.yaml anyway:\n%s", got)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the harness failure: %v", statErr)
	}
}

// TestInitResumeDeployModeWriteFailureKeepsMarker — the harness already matches, so its
// step is a no-op; the deploy-mode record must still be written to a read-only llm.yaml,
// and that refusal fails the resume with the deploy-mode step's error.
func TestInitResumeDeployModeWriteFailureKeepsMarker(t *testing.T) {
	t1619SkipWithoutFilePerms(t)
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "deploy-mode-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	llmPath := filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml")
	before := readResumeLLMYAML(t, projectDir)
	if strings.Contains(before, "deployment_mode:") {
		t.Fatalf("precondition: the seed run already recorded the deploy mode:\n%s", before)
	}
	if !hasHarness(before, "claude") {
		t.Fatalf("precondition: the seed run did not record the claude harness:\n%s", before)
	}
	if err := os.Chmod(llmPath, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(llmPath, 0o644) })

	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "apply deploy mode") {
		t.Fatalf("the refused deploy-mode write did not surface as the resume's error: %v", err)
	}
	if got := readResumeLLMYAML(t, projectDir); strings.Contains(got, "deployment_mode:") {
		t.Fatalf("the refused deploy-mode write changed llm.yaml anyway:\n%s", got)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the deploy-mode failure: %v", statErr)
	}
}

// TestInitResumeMalformedGlobalSettingsFailsAtAutonomyStep — a global settings file that
// is not JSON stops the resume at the autonomy-tier step, which reads and rewrites the same
// file before the global settings cleanup runs. The marker stays, so the next run resumes.
func TestInitResumeMalformedGlobalSettingsFailsAtAutonomyStep(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "global-settings-proj")
	resumeSeedRun(t, projectDir, "claude")

	t1619MustWrite(t, filepath.Join(os.Getenv("HOME"), ".claude", "settings.json"), "{ not json\n")
	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "apply autonomy tier") {
		t.Fatalf("the malformed global settings did not stop the resume at the autonomy step: %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the global settings failure: %v", statErr)
	}
}

// TestInitResumeGitfileProjectWarnsOnHookInstalls — a project whose .git is a file (a
// linked worktree) cannot hold .git/hooks, so both git hook installs are reported as
// warnings; the hook step never fails the resume.
func TestInitResumeGitfileProjectWarnsOnHookInstalls(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "gitfile-proj")
	resumeSeedRun(t, projectDir, "claude")

	t1619MustWrite(t, filepath.Join(projectDir, ".git"), "gitdir: /nonexistent/linked-worktree\n")
	errOut, err := t1619ResumeRunCapture(t, projectDir, true)
	if err != nil {
		t.Fatalf("the hook warnings must not fail the resume: %v", err)
	}
	for _, want := range []string{"Pre-push hook installation failed", "Pre-commit hook installation failed"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("the resume did not report %q on the error stream:\n%s", want, errOut)
		}
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); !os.IsNotExist(statErr) {
		t.Fatalf("a completed resume left its marker behind (stat err = %v)", statErr)
	}
}

// TestInitResumeMCPEntryFailureKeepsMarker — a .mcp.json that cannot be provisioned
// (malformed JSON) fails the resume at the MCP step, and the marker stays.
func TestInitResumeMCPEntryFailureKeepsMarker(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "mcp-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	t1619MustWrite(t, filepath.Join(projectDir, ".mcp.json"), "{ not json\n")
	err := resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "provision MCP entry") {
		t.Fatalf("the unprovisionable MCP entry did not surface as the resume's error: %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the MCP failure: %v", statErr)
	}
}

// TestInitResumeProjectLayoutFailureKeepsMarker — a project directory whose todo entry
// is a regular file makes the layout step fail, and the marker stays.
func TestInitResumeProjectLayoutFailureKeepsMarker(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "layout-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	dir, err := homestate.ProjectDir(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	t1619MustMkdir(t, dir)
	if err := os.RemoveAll(filepath.Join(dir, "todo")); err != nil {
		t.Fatal(err)
	}
	t1619MustWrite(t, filepath.Join(dir, "todo"), "not a directory\n")

	err = resumeRun(t, projectDir, "")
	if err == nil || !strings.Contains(err.Error(), "project layout") {
		t.Fatalf("the blocked project layout did not surface as the resume's error: %v", err)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the resume marker did not stay in place after the layout failure: %v", statErr)
	}
}

// TestInitResumeRetrackFailureWarnsAndKeepsMarker — a manifest that cannot be saved after
// the sections are retracked is reported as a warning; the resume returns nil and the
// marker stays, so the next run resumes the bookkeeping. The save needs the manifest's
// directory writable, so the marker is planted first (an existing file is truncated
// without a new directory entry) and only then is the directory made read-only.
func TestInitResumeRetrackFailureWarnsAndKeepsMarker(t *testing.T) {
	t1619SkipWithoutFilePerms(t)
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "retrack-failure-proj")
	resumeSeedRun(t, projectDir, "claude")

	writeResumeMarker(t, projectDir)
	moaiDir := filepath.Join(projectDir, ".moai")
	if err := os.Chmod(moaiDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(moaiDir, 0o755) })

	errOut, err := t1619ResumeRunCapture(t, projectDir, false)
	if err != nil {
		t.Fatalf("a retrack warning must not fail the resume: %v", err)
	}
	if !strings.Contains(errOut, "Warning: manifest retrack (config sections) failed") {
		t.Fatalf("the retrack failure was not reported on the error stream:\n%s", errOut)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); statErr != nil {
		t.Fatalf("the marker was retired although the retrack failed: %v", statErr)
	}
}

// TestInitResumeWiringRecordWriteFailureIsReturned — a selection record whose directory
// cannot be created reports the failure to its writer.
func TestInitResumeWiringRecordWriteFailureIsReturned(t *testing.T) {
	projectRoot := t.TempDir()
	t1619MustWrite(t, filepath.Join(projectRoot, ".moai", "state"), "not a directory\n")
	if err := persistResumeWiring(projectRoot, agentWiringClaude); err == nil {
		t.Fatal("a selection record under a blocked state directory must report the failure")
	}
}

// TestInitResumeWiringRecordWithUnknownHarnessReadsAsNoRecord — a record naming a harness
// outside the closed set reads as no record, so it can never steer a resume.
func TestInitResumeWiringRecordWithUnknownHarnessReadsAsNoRecord(t *testing.T) {
	projectRoot := t.TempDir()
	t1619MustWrite(t, resumeWiringPath(projectRoot), "not-a-harness\n")
	got, ok := readResumeWiring(projectRoot)
	if ok || got != "" {
		t.Fatalf("a record outside the closed harness set must read as no record, got (%q, %v)", got, ok)
	}
}

// TestInitResumeInitMarkerWriteFailureIsReturned — an init marker whose directory cannot
// be created reports the failure to the caller that opens the ensure window.
func TestInitResumeInitMarkerWriteFailureIsReturned(t *testing.T) {
	projectRoot := t.TempDir()
	t1619MustWrite(t, filepath.Join(projectRoot, ".moai", "state"), "not a directory\n")
	if err := persistInitPending(projectRoot, []string{"core"}); err == nil {
		t.Fatal("an init marker under a blocked state directory must report the failure")
	}
}

// TestInitResumeNoInitMarkerReadsAsNoBundles — with no init marker there is no recorded
// bundle selection to restore.
func TestInitResumeNoInitMarkerReadsAsNoBundles(t *testing.T) {
	if names := initPendingBundles(t.TempDir()); names != nil {
		t.Fatalf("a project with no init marker must restore no bundle selection, got %q", names)
	}
}

// TestInitResumeMarkerAloneIsACheckpoint — the resume's own marker is a checkpoint even
// when no other interruption evidence exists.
func TestInitResumeMarkerAloneIsACheckpoint(t *testing.T) {
	projectRoot := t.TempDir()
	t1619MustWrite(t, t1619MarkerPath(projectRoot), "resume in progress\n")
	if !initResumeCheckpoint(t.TempDir(), projectRoot) {
		t.Fatal("a resume marker alone must read as a resume checkpoint")
	}
}

// TestInitResumePendingJournalIsInterruptionEvidence — a pending-install journal is
// evidence that an install was interrupted, so a project whose deployment never finished
// resumes; with no journal the same project is not resumable.
func TestInitResumePendingJournalIsInterruptionEvidence(t *testing.T) {
	home := t.TempDir()
	projectRoot := t.TempDir()
	t1619MustWrite(t, filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml"),
		"llm:\n  harness: claude\n")
	if initResumeCheckpoint(home, projectRoot) {
		t.Fatal("with no interruption evidence the project must not read as resumable")
	}
	t1619MustWrite(t, userassets.JournalPath(home), "{}\n")
	if !initResumeCheckpoint(home, projectRoot) {
		t.Fatal("a pending-install journal must make the undeployed project resumable")
	}
}

// TestInitResumeHealthyUserManifestIsLeftInPlace — a user manifest that parses is not a
// corrupt one: the quarantine leaves it where it is and renames nothing aside.
func TestInitResumeHealthyUserManifestIsLeftInPlace(t *testing.T) {
	home := t.TempDir()
	t1619MustWrite(t, userassets.ManifestPath(home), "{}\n")
	quarantineCorruptUserManifest(home)
	if _, err := os.Stat(userassets.ManifestPath(home)); err != nil {
		t.Fatalf("a healthy manifest was moved away: %v", err)
	}
	matches, err := filepath.Glob(userassets.ManifestPath(home) + ".corrupt-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("a healthy manifest was quarantined as corrupt: %q", matches)
	}
}

// TestInitResumeTrailingCommaGlobalSettingsWarnsAndResumes — a global settings file with a
// trailing comma (a realistic hand edit) is not JSON. The autonomy-tier step rewrites the
// permissions region by splicing text, so it may accept the file; the global settings
// cleanup parses it as JSON and reports the failure as a warning, and the resume completes.
func TestInitResumeTrailingCommaGlobalSettingsWarnsAndResumes(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "trailing-comma-proj")
	resumeSeedRun(t, projectDir, "claude")

	t1619MustWrite(t, filepath.Join(os.Getenv("HOME"), ".claude", "settings.json"),
		"{\n  \"permissions\": {\n    \"defaultMode\": \"acceptEdits\"\n  },\n}\n")
	errOut, err := t1619ResumeRunCapture(t, projectDir, false)
	if err != nil {
		t.Fatalf("a global settings parse warning must not fail the resume: %v", err)
	}
	if !strings.Contains(errOut, "Warning: global settings env update failed") {
		t.Fatalf("the global settings parse failure was not reported on the error stream:\n%s", errOut)
	}
	if _, statErr := os.Stat(t1619MarkerPath(projectDir)); !os.IsNotExist(statErr) {
		t.Fatalf("a completed resume left its marker behind (stat err = %v)", statErr)
	}
}
