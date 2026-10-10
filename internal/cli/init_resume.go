package cli

// init_resume.go — the init resume path (SPEC-USERASSET-DEPLOY-GUARD-001
// M6, REQ-SRF-007 + audit DEBT R4): a prior init whose user-asset ensure
// failed leaves the project INITIALIZED but incomplete — the executor's
// project half is on disk while the ensure and every setup step after it
// never ran. Re-running init hit the "already initialized" refusal, so
// the shortfall had no path forward. The resume completes the ensure
// shortfall and the not-yet-run setup steps (ApplyHarness, the MCP entry,
// the Codex wiring) instead of refusing; the template relocation the
// failed attempt already finished is skipped (never re-run).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
	"github.com/spf13/cobra"
)

// userHomeDirOrEmpty resolves the user home for the resume-checkpoint
// probe; an unresolvable home reads as "no checkpoint" (the ordinary
// re-run keeps the update redirect).
func userHomeDirOrEmpty() string {
	home, err := userHomeDirFn()
	if err != nil {
		return ""
	}
	return home
}

// resumeInitializedProject completes an init whose prior attempt failed at
// the user-asset ensure: the ENTIRE re-check → quarantine → install span
// runs INSIDE the user lock scope (gate rounds 37-2/38-3 — a quarantine
// outside the lock could interleave with another run's healthy manifest
// store and move the HEALTHY file), then the setup steps the failed
// attempt never reached run in the production order (project layout →
// autonomy tier → ApplyHarness → MCP entry → Codex wiring; gate round
// 35-6 included).
// resumeInitializedProject completes an init whose prior attempt failed at
// the user-asset ensure: the ENTIRE re-check → quarantine → install span
// runs INSIDE the user lock scope (gate rounds 37-2/38-3 — a quarantine
// outside the lock could interleave with another run's healthy manifest
// store and move the HEALTHY file), then the setup steps the failed
// attempt never reached run in the production order (project layout →
// autonomy tier → profile sync → Jev record → ApplyHarness → deploy mode → git hooks →
// MCP entry → Codex wiring → section retrack; gate round 35-6 included). Gate round 47-4:
// the wizard's Jev answer rides in — a resume that skipped it left
// jev.enabled false against an explicit wizard selection.
func resumeInitializedProject(cmd *cobra.Command, opts *project.InitOptions, wiring agentWiring, wizardRan bool, wizardResult *wizard.WizardResult, prefs profile.ProfilePreferences) error {
	homeDir, homeErr := userHomeDirFn()
	if homeErr != nil {
		return fmt.Errorf("resolve user home: %w", homeErr)
	}

	// Finding 1 (card t1591): the harness the FAILING run selected is the
	// record of the interrupted --llm, and a resume with no explicit selection
	// of its own reads it back. The file layout does not decide it: neither
	// claude nor both writes CLAUDE.md any more, and both deploy the same
	// .claude/ surface, so the layout cannot tell them apart.
	if recorded, ok := resumeWiringFromRecord(cmd, wizardResult, opts.ProjectRoot); ok {
		wiring = recorded
	} else if legacyLayoutSpeaks(cmd, opts.ProjectRoot, wiring) {
		// Gate round 51-3 (legacy fallback, applies only without a record): the
		// deployed LAYOUT is the interrupted record of the original --llm
		// selection. A re-run without options resolves the claude default;
		// applying it over a deployed GPT layout records harness:claude while
		// the Codex wiring is skipped — config and on-disk layout stay
		// permanently mismatched.
		wiring = agentWiringGPT
	}

	// Item C (card t1591): a harness change is refused, not half-applied. The check
	// runs after the wiring decision and before any side effect; the user lock below
	// creates a guard file under the user home, so the refusal precedes it as well.
	if err := refuseResumeHarnessChange(opts.ProjectRoot, wiring); err != nil {
		return err
	}
	lock, lockErr := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if lockErr != nil {
		return fmt.Errorf("user-asset resume lock: %w", lockErr)
	}
	defer func() { _ = lock.Release() }()

	// Gate round 51-2: the checkpoint OUTLIVES the asset install — a
	// transient failure in a LATER post-step (autonomy tier, harness, MCP)
	// must leave the next run a resume route instead of the "already
	// initialized" refusal. The marker is written at entry and removed only
	// when EVERY post-step completed; initResumeCheckpoint reads it as a
	// checkpoint.
	markerPath := filepath.Join(opts.ProjectRoot, ".moai", "resume-pending")
	if err := os.WriteFile(markerPath, []byte("resume in progress\n"), 0o644); err != nil {
		return fmt.Errorf("write resume marker: %w", err)
	}

	// The ensure shortfall: a CORRUPT user manifest is quarantined inside
	// the lock scope, then the installer runs from scratch (the doctor's
	// sanctioned rebuild-from-fresh-init recovery for that state).
	quarantineCorruptUserManifest(homeDir)
	inst, err := newUserAssetInstaller(homeDir)
	if err != nil {
		return err
	}
	res, err := inst.InstallPreserveSelection(parseBundleSelection(getStringFlag(cmd, "bundles")))
	if err != nil {
		return fmt.Errorf("user-asset ensure: %w", err)
	}
	writeInstallSummary(cmd.OutOrStdout(), "user-asset resume", res)
	_ = lock.Release()

	// The project layout the failed attempt never reached.
	if err := homestate.EnsureProjectLayout(opts.ProjectRoot); err != nil {
		return fmt.Errorf("project layout: %w", err)
	}
	// Gate round 35-6: the AUTONOMY TIER this run answered is applied too —
	// the failed attempt never reached the bundle, and a resume that skips
	// it leaves permissions.defaultMode empty (the gate repro).
	projectSettingsPath := filepath.Join(opts.ProjectRoot, ".claude", "settings.json")
	if wiring == agentWiringGPT {
		projectSettingsPath = ""
	}
	if err := applyAutonomyTierBundleFn(
		opts.ProjectRoot,
		filepath.Join(homeDir, ".claude", "settings.json"),
		projectSettingsPath,
		opts.AutonomyTier,
	); err != nil {
		return fmt.Errorf("apply autonomy tier: %w", err)
	}
	// Item D (card t1591): the profile synchronization a healthy init runs after its
	// template deploy (init.go, the healthy tail). The same helper and inputs, so a
	// resumed run carries the current profile preferences and the wizard's user name
	// and conversation language into the project config. A failure is a post-step
	// failure: the marker stays in place and the next run resumes.
	if err := syncInitProfile(opts.ProjectRoot, prefs, wizardResult); err != nil {
		return fmt.Errorf("profile sync: %w", err)
	}
	// Gate round 47-4: the wizard's Jev answer is part of the same post-path
	// (healthy init applies it before the harness record — init.go's
	// applyJevFromWizard) — a resume without it silently flipped an explicit
	// wizard selection back to false.
	if err := applyJevFromWizard(wizardRan, wizardResult, opts.ProjectRoot); err != nil {
		return fmt.Errorf("apply jev selection: %w", err)
	}
	// Gate round 50-3: the participation consent completes the same family —
	// a resume that skipped it never wrote the Asked record, so an explicit
	// wizard DECLINE stayed invisible and the consent read as enabled.
	if err := applyParticipationFromWizard(wizardRan, wizardResult, opts.ProjectRoot); err != nil {
		return fmt.Errorf("apply participation selection: %w", err)
	}
	// DEBT R4: the resume is judged by CONTENT downstream — the harness
	// config, the MCP entry, and the Codex wiring the failed attempt never
	// wrote are all written here, in the production order.
	if err := template.ApplyHarness(opts.ProjectRoot, string(wiring)); err != nil {
		return fmt.Errorf("apply harness: %w", err)
	}
	// Gate round 44-4: the deploy-mode record is part of the same post-path —
	// a resume without it leaves deployment_mode empty, and the next update
	// misroutes to the migration path (healthy init records "local" here,
	// SPEC-INIT-SHRINK-001 REQ-009).
	if err := template.ApplyDeployMode(opts.ProjectRoot, opts.DeployMode); err != nil {
		return fmt.Errorf("apply deploy mode: %w", err)
	}
	// The global settings cleanup the healthy init runs right after the deploy-mode
	// record (init.go): it removes the legacy global MoAI hooks directory, which the
	// failed attempt never reached. Non-fatal exactly as in healthy init, and it does
	// not gate the marker: a malformed global settings file must not keep the resume
	// route alive on every run.
	if globalErr := ensureGlobalSettingsEnv(); globalErr != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: global settings env update failed: %v\n", globalErr)
	}
	// The git hooks the failed attempt never reached (queued with gate
	// round 44-4): pre-push (REQ-CIAUT-002) and pre-commit (REQ-PC-001),
	// non-fatal exactly as in healthy init — an install failure must not
	// abort a resume that already repaired the manifest store.
	if pushErr := installPrePushHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); pushErr != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Pre-push hook installation failed: %v\n", pushErr)
	}
	if commitErr := installPreCommitHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); commitErr != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Pre-commit hook installation failed: %v\n", commitErr)
	}
	mcpDeclined := !opts.MCPProvision
	switch wiring {
	case agentWiringGPT:
		mcpDeclined = true
	case agentWiringBoth:
		mcpDeclined = false
	}
	if err := provisionMCPEntryUnlessDeclined(cmd.OutOrStdout(), opts.ProjectRoot, mcpDeclined); err != nil {
		return fmt.Errorf("provision MCP entry: %w", err)
	}
	// Codex wiring (card t1591, repair round): on the resume path a wiring failure
	// is a failed post-step, not a warning. Returning here skips the snapshot
	// settle, the section retrack, and the marker and record retirement below, so
	// the next run resumes from the marker and the selection record.
	if err := wireCodex(cmd, wiring, opts.ProjectRoot); err != nil {
		return fmt.Errorf("codex wiring: %w", err)
	}
	// Finding 2 (card t1591): the resume is this flow's end, so it settles the
	// snapshot lifecycle the way a healthy init does (init.go): the settings.json
	// and .mcp.json staging copies are promoted to the canonical base the next
	// update merges against. Settled here, after every post-step succeeded — a
	// half-finished resume never promotes. Preserved is false: init and resume
	// run no settings merge, so the staged render is the base.
	backup.SettleSettingsSnapshot(opts.ProjectRoot, false, cmd.ErrOrStderr())
	backup.SettleMCPSnapshot(opts.ProjectRoot, false, cmd.ErrOrStderr())
	// Item E (card t1591): the resume rewrote the section family above (llm.yaml
	// through ApplyHarness and ApplyDeployMode, user.yaml and language.yaml through the
	// profile sync), and it re-records their on-disk hashes as a healthy init does
	// (init.go, card t1277). Without the retrack the manifest keeps the pre-write hashes,
	// and the next init --force or update reads the rewritten sections as user edits.
	// It runs after the last section writer; template_managed-only filtering keeps
	// user-owned entries untouched. A failure is non-fatal: it is reported as a warning,
	// the resume does not fail, and the marker stays in place (below).
	retrackErr := retrackSectionFiles(opts.ProjectRoot, cmd.ErrOrStderr())
	if retrackErr != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: manifest retrack (config sections) failed: %v\n", retrackErr)
	}
	// Gate round 51-2: every post-step completed — the checkpoint marker
	// retires (the early-return paths above leave it in place, keeping the
	// next run's resume route alive). Finding 1: the selection record retires
	// with the marker, and only once the marker is gone — a marker that
	// survives still needs its selection.
	// Item E: a retrack failure is not a completed post-step, so the marker stays and
	// the next run resumes the bookkeeping. The resume itself still returns nil.
	if retrackErr == nil {
		// Finding B (card t1591): the resumed init's ensure window closes with its post-steps.
		clearInitPending(opts.ProjectRoot)
		if err := os.Remove(markerPath); err == nil || os.IsNotExist(err) {
			clearResumeWiring(opts.ProjectRoot)
		}
	}
	return nil
}

// deployedWiringIsGPT reads the deployed layout as the interrupted --llm
// record when no selection record exists (gate round 51-3, legacy fallback).
// The codex-only deployer writes no .claude/ surface and projects AGENTS.md,
// so .claude/ absent with AGENTS.md present names the gpt layout. The probe is
// .claude/, not CLAUDE.md: no harness writes CLAUDE.md any more, so a
// CLAUDE.md probe read every claude and both project as gpt.
func deployedWiringIsGPT(projectRoot string) bool {
	if _, err := os.Stat(filepath.Join(projectRoot, ".claude")); err == nil {
		return false // the claude-side deployer ran (claude and both alike)
	}
	_, err := os.Stat(filepath.Join(projectRoot, "AGENTS.md"))
	return err == nil
}

// legacyLayoutSpeaks reports whether the deployed layout decides the resume's
// harness through the legacy fallback. It speaks only for the claude default, and
// only when the resume has no record and no explicit --llm (card t1591, Item B).
// An explicit --llm is the current invocation's answer. A record that exists is
// the interrupted run's own selection, and a changed wizard answer must win over
// the layout while it stands. The --llm check is the one resumeWiringFromRecord
// applies.
func legacyLayoutSpeaks(cmd *cobra.Command, projectRoot string, wiring agentWiring) bool {
	if wiring != agentWiringClaude || !deployedWiringIsGPT(projectRoot) {
		return false
	}
	if cmd.Flags().Changed("llm") && getStringFlag(cmd, "llm") != "" {
		return false
	}
	_, recorded := readResumeWiring(projectRoot)
	return !recorded
}

// refuseResumeHarnessChange refuses a resume whose effective wiring contradicts the
// deployed harness (card t1591, Item C). The resume restores the content the failed
// run never wrote and deploys no template files, so a harness switch would rewrite
// llm.yaml over the old harness's files and still report success, without the new
// harness's files. The deployed harness is read in order: the selection record the
// failing run left (an exact match); with no record, the GPT layout, .claude/ absent
// (an exact match with gpt); with no record, the claude-side layout, .claude/ present
// (claude and both deploy the same surface and the layout cannot tell them apart, so
// only a gpt request contradicts it). With none of this evidence, nothing is refused.
func refuseResumeHarnessChange(projectRoot string, wiring agentWiring) error {
	if recorded, ok := readResumeWiring(projectRoot); ok {
		if wiring == recorded {
			return nil
		}
		return resumeHarnessRefusal(recorded, "the selection the failing run recorded", wiring,
			"--llm "+string(recorded)+" (or accept the recorded default in the wizard)")
	}
	if deployedWiringIsGPT(projectRoot) {
		if wiring == agentWiringGPT {
			return nil
		}
		return resumeHarnessRefusal(agentWiringGPT, "the GPT layout; no selection record", wiring, "--llm gpt")
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".claude")); err == nil && wiring == agentWiringGPT {
		return resumeHarnessRefusal(agentWiringClaude,
			"the claude-side layout; no selection record, and claude and both cannot be told apart on disk", wiring,
			"--llm claude (or --llm both, if that was the deployment)")
	}
	return nil
}

// resumeHarnessRefusal is the error a refused harness change returns. It names the
// deployed and the requested harness and the way out, and it states why the resume
// cannot switch the harness, so the operator does not retry the same request.
func resumeHarnessRefusal(deployed agentWiring, basis string, requested agentWiring, resumeWith string) error {
	return fmt.Errorf("resume refused: harness %q is deployed (%s), and this run requests harness %q. "+
		"The resume restores the content the failed run never wrote and deploys no template files, "+
		"so it cannot switch the harness: the requested harness's files would never be deployed. "+
		"Resume with %s, or re-run with --force to reinitialize from scratch",
		deployed, basis, requested, resumeWith)
}

// resumeWiringPath is the project-scoped record of the harness selection the
// failing init run made. It sits under .moai/state/ (gitignored) and is a file
// separate from the resume-pending marker, so the marker's checkpoint
// semantics (initResumeCheckpoint) do not change.
// @MX:ANCHOR: [AUTO] resumeWiringPath fixes the location of the resume-wiring record for its writer and both readers
// @MX:REASON: [AUTO] fan_in=3 non-test call sites: persistResumeWiring, readResumeWiring, clearResumeWiring
func resumeWiringPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", "resume-wiring")
}

// persistResumeWiring records the harness selection an init run made before its
// user-asset ensure, so an interrupted run's resume can read it back.
func persistResumeWiring(projectRoot string, wiring agentWiring) error {
	path := resumeWiringPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(string(wiring)+"\n"), 0o644)
}

// @MX:ANCHOR: [AUTO] readResumeWiring is the only reader of the resume-wiring record; a change to its read contract changes every resume decision
// @MX:REASON: [AUTO] fan_in=4 non-test call sites: legacyLayoutSpeaks, refuseResumeHarnessChange, resumeWiringFromRecord, resumeAgentWiringDefault
// readResumeWiring returns the recorded selection. Only a member of the closed
// set counts; anything else reads as no record. The read is the non-blocking
// regular-file read every resume record uses, so a FIFO at the path reads as
// no record instead of hanging the resume.
func readResumeWiring(projectRoot string) (agentWiring, bool) {
	data, ok := readManifestRecord(resumeWiringPath(projectRoot))
	if !ok {
		return "", false
	}
	switch wiring := agentWiring(strings.TrimSpace(string(data))); wiring {
	case agentWiringClaude, agentWiringGPT, agentWiringBoth:
		return wiring, true
	}
	return "", false
}

// resumeWiringFromRecord returns the recorded selection when the resume was
// given no explicit selection of its own. An explicit --llm, or a wizard
// agent_wiring answer, is the current invocation's answer and wins over the
// record — the same precedence resolveAgentWiringWithWizard applies. That answer is
// then checked against the deployed harness: a mismatch is refused, never applied
// over the deployed files (refuseResumeHarnessChange).
func resumeWiringFromRecord(cmd *cobra.Command, wizardResult *wizard.WizardResult, projectRoot string) (agentWiring, bool) {
	if cmd.Flags().Changed("llm") && getStringFlag(cmd, "llm") != "" {
		return "", false
	}
	if wizardResult != nil && wizardResult.AgentWiring != "" {
		return "", false
	}
	return readResumeWiring(projectRoot)
}

// resumeAgentWiringDefault returns the harness the init wizard's agent_wiring
// question pre-fills for a resume: the selection the interrupted run recorded.
// The record is read only when the resume checkpoint applies (card t1591). A
// record without a checkpoint is stale and pre-fills nothing, and so does an
// unresolvable home, which would otherwise read CWD-relative .moai paths. "" keeps
// the shipped claude default. The read runs before the wizard builds its form,
// because the form takes its default at build time; the operator's answer
// overrides it on submit.
func resumeAgentWiringDefault(homeDir, projectRoot string) string {
	if homeDir == "" || !initResumeCheckpoint(homeDir, projectRoot) {
		return ""
	}
	wiring, ok := readResumeWiring(projectRoot)
	if !ok {
		return ""
	}
	return string(wiring)
}

// clearResumeWiring retires the selection record once the resume completed.
func clearResumeWiring(projectRoot string) {
	_ = os.Remove(resumeWiringPath(projectRoot))
}

// initPendingPath is the project-scoped marker a fresh init leaves across its
// user-asset ensure window (card t1591, finding B). It sits beside the selection
// record under .moai/state/, which is gitignored, and it is a file of its own, so
// the resume-pending marker's checkpoint semantics do not change.
// @MX:ANCHOR: [AUTO] initPendingPath fixes the location of the init marker for its writer, its retirements, and the checkpoint reader
// @MX:REASON: [AUTO] fan_in=3 non-test call sites: persistInitPending, clearInitPending, initResumeCheckpoint
func initPendingPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", "init-pending")
}

// persistInitPending opens the ensure window: it writes the init marker before the
// user-asset ensure runs, so an ensure that fails leaves the rerun a resume route.
func persistInitPending(projectRoot string) error {
	path := initPendingPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("init in progress\n"), 0o644)
}

// clearInitPending closes the ensure window once the init's post-steps completed.
func clearInitPending(projectRoot string) {
	_ = os.Remove(initPendingPath(projectRoot))
}

// initResumeCheckpoint reports whether an explicit interruption
// checkpoint justifies the init resume (gate round 35-3):
//   - a PENDING-INSTALL JOURNAL (an attempt interrupted mid-install), or
//   - a CORRUPT user manifest (an attempt that failed before/at the
//     ensure), or
//   - the INIT MARKER (a fresh attempt that died in or after its user-asset
//     ensure; card t1591, finding B),
//
// AND this project's own deployment is incomplete (.mcp.json absent — the
// failed attempt never reached MCP provisioning). Gate round 37-3: the
// checkpoint is PROJECT-SCOPED through the deployment-completeness arm —
// another project's interrupted journal (the journal is user-global) no
// longer routes a healthy project into resume. Gate rounds 38-3/30: both
// reads are NON-BLOCKING with the type check bound to the open handle.
func initResumeCheckpoint(homeDir, projectRoot string) bool {
	// Gate round 51-2: the resume's own pending marker IS a checkpoint —
	// the original interruption evidence (journal/corrupt manifest) is
	// consumed by the first resume, and a LATER-step failure must not wedge
	// the half-done resume behind "already initialized".
	if _, err := os.Stat(filepath.Join(projectRoot, ".moai", "resume-pending")); err == nil {
		return true
	}
	interrupted := false
	// Finding B (card t1591): a fresh init that died in or after its user-asset ensure
	// leaves the init marker. The ensure can fail before any journal exists (a lock
	// refusal), so the marker is the only interruption evidence a rerun can read.
	if _, err := os.Stat(initPendingPath(projectRoot)); err == nil {
		interrupted = true
	}
	if jData, ok := readManifestRecord(userassets.JournalPath(homeDir)); ok {
		var j userassets.PendingJournal
		if json.Unmarshal(jData, &j) == nil {
			interrupted = true
		}
	}
	manifestCorrupt := false
	if mData, ok := readManifestRecord(userassets.ManifestPath(homeDir)); ok {
		var m userassets.Manifest
		if json.Unmarshal(mData, &m) != nil {
			manifestCorrupt = true
		}
	}
	if !interrupted && !manifestCorrupt {
		return false
	}
	// Deployment completeness: the executor's config sections must exist —
	// a project the executor never deployed is NOT "initialized but
	// incomplete"; it keeps the original redirect.
	//
	// Gate round 46-3: deployment_mode is written ONLY by the post-deploy
	// ApplyDeployMode step — the shipped template carries no such key (the
	// template's llm.yaml and .mcp.json both already carry harness/moai
	// content, so mere presence distinguishes nothing). Its presence says
	// the post-deploy path RAN: the project is COMPLETE, and a stale
	// user-global journal (the journal carries no project identity) must
	// not hijack it into a resume consuming another project's pending
	// intent. Its absence is the M6 window — deployed, setup unfinished —
	// and stays resumable.
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	llmData, err := os.ReadFile(filepath.Join(sections, "llm.yaml"))
	if err != nil {
		return false
	}
	return !strings.Contains(string(llmData), "deployment_mode:")
}

// quarantineCorruptUserManifest renames a CORRUPT user manifest aside with
// a timestamped suffix (never deletes). Absent, healthy, and unreadable-
// for-other-reasons manifests are left untouched. Gate round 36: the read
// is NON-BLOCKING with the type check bound to the open handle — a FIFO at
// the manifest path surfaces as "no record" instead of hanging the resume.
func quarantineCorruptUserManifest(homeDir string) {
	manifestPath := userassets.ManifestPath(homeDir)
	data, ok := readManifestRecord(manifestPath)
	if !ok {
		return // absent or not a readable regular file — not ours to move
	}
	var probe json.RawMessage
	if json.Unmarshal(data, &probe) != nil {
		// not even valid JSON: a corrupt manifest — quarantine it
	} else {
		var m userassets.Manifest
		if json.Unmarshal(data, &m) == nil {
			return // parses as a manifest — healthy, leave it
		}
	}
	_ = os.Rename(manifestPath, manifestPath+".corrupt-"+time.Now().UTC().Format("20060102T150405"))
}
