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
	"gopkg.in/yaml.v3"
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
func resumeInitializedProject(cmd *cobra.Command, opts *project.InitOptions, wizardRan bool, wizardResult *wizard.WizardResult, prefs profile.ProfilePreferences) error {
	homeDir, homeErr := userHomeDirFn()
	if homeErr != nil {
		return fmt.Errorf("resolve user home: %w", homeErr)
	}

	// The harness selection is claude-only (card t1909). A legacy record or
	// layout from a pre-removal gpt/both init can no longer be restored — it
	// is refused instead, so the resume never mis-wires a deployed tree.
	if err := refuseResumeHarnessChange(opts.ProjectRoot); err != nil {
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

	// Finding 1 (card t1591): the ensure installs the effective selection — the explicit
	// --bundles, or the one the failing run recorded before its ensure. The init marker
	// keeps the selection the ensure runs with, so a later resume restores the most
	// recent one.
	selection := resumeBundleSelection(cmd, opts.ProjectRoot)
	if err := persistInitPending(opts.ProjectRoot, selection); err != nil {
		return fmt.Errorf("record bundle selection: %w", err)
	}
	// The ensure shortfall: a CORRUPT user manifest is quarantined inside
	// the lock scope, then the installer runs from scratch (the doctor's
	// sanctioned rebuild-from-fresh-init recovery for that state).
	quarantineCorruptUserManifest(homeDir)
	inst, err := newUserAssetInstaller(homeDir)
	if err != nil {
		return err
	}
	res, err := inst.InstallPreserveSelection(selection)
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
	if err := applyAutonomyTierBundleFn(
		opts.ProjectRoot,
		filepath.Join(homeDir, ".claude", "settings.json"),
		filepath.Join(opts.ProjectRoot, ".claude", "settings.json"),
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
	// config and the MCP entry the failed attempt never wrote are written
	// here, in the production order. The harness is claude-only (card t1909).
	if err := template.ApplyHarness(opts.ProjectRoot, string(agentWiringClaude)); err != nil {
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
	if err := provisionMCPEntryUnlessDeclined(cmd.OutOrStdout(), opts.ProjectRoot, !opts.MCPProvision); err != nil {
		return fmt.Errorf("provision MCP entry: %w", err)
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
// The codex-only deployer writes no .claude/ surface and projects the
// instruction file, so .claude/ absent with the instruction file present —
// CLAUDE.md (canonical) OR a legacy AGENTS.md the pre-migration deployer
// wrote (card-review R1) — names the gpt layout. The probe is .claude/ first:
// a claude or both deployment always carries .claude/, whichever instruction
// name its era wrote, so a legacy claude tree never reads as gpt.
func deployedWiringIsGPT(projectRoot string) bool {
	if _, err := os.Stat(filepath.Join(projectRoot, ".claude")); err == nil {
		return false // the claude-side deployer ran (claude and both alike)
	}
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		if _, err := os.Stat(filepath.Join(projectRoot, name)); err == nil {
			return true
		}
	}
	return false
}

// refuseResumeHarnessChange refuses a resume whose deployed harness is not the
// claude harness init now deploys (card t1591, Item C, narrowed by card t1909:
// the claude harness is the only deployment, so nothing can be restored INTO a
// legacy gpt/both tree). The resume restores the content the failed run never
// wrote and deploys no template files, so resuming a gpt/both tree as claude
// would rewrite llm.yaml over the old harness's files and still report success,
// without the claude harness's own files — a mismatched tree. The deployed
// harness is read in order: the selection record the failing run left; with no
// record, the GPT layout, .claude/ absent with CLAUDE.md present. With none of
// this evidence, nothing is refused.
func refuseResumeHarnessChange(projectRoot string) error {
	if recorded, ok := readResumeWiring(projectRoot); ok {
		if recorded == agentWiringClaude {
			return nil
		}
		return resumeHarnessRefusal(recorded, "the selection the failing run recorded", agentWiringClaude,
			"--force (moai init deploys the claude harness only; the recorded "+string(recorded)+" deployment cannot be resumed)")
	}
	if deployedWiringIsGPT(projectRoot) {
		return resumeHarnessRefusal(agentWiringGPT, "the GPT layout; no selection record", agentWiringClaude,
			"--force (moai init deploys the claude harness only; a GPT-layout deployment cannot be resumed)")
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
		"Resume with %s to reinitialize from scratch",
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

// clearResumeWiring retires the selection record once the resume completed.
func clearResumeWiring(projectRoot string) {
	_ = os.Remove(resumeWiringPath(projectRoot))
}

// initPendingPath is the project-scoped marker a fresh init leaves across its
// user-asset ensure window (card t1591, finding B). It sits beside the selection
// record under .moai/state/, which is gitignored, and it is a file of its own, so
// the resume-pending marker's checkpoint semantics do not change.
// @MX:ANCHOR: [AUTO] initPendingPath fixes the location of the init marker for its writer, its retirements, and the checkpoint readers
// @MX:REASON: [AUTO] fan_in=4 non-test call sites: persistInitPending, clearInitPending, initResumeCheckpoint, initPendingBundles
func initPendingPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", "init-pending")
}

// persistInitPending opens the ensure window: it writes the init marker before the
// user-asset ensure runs, so an ensure that fails leaves the rerun a resume route. The
// marker also carries the requested bundle selection, one name per line under its
// header, so a resume with no --bundles of its own restores it (card t1591, finding 1).
func persistInitPending(projectRoot string, selection []string) error {
	path := initPendingPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lines := append([]string{"init in progress"}, selection...)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// initPendingBundles returns the bundle selection the init marker recorded: every line
// after the header that names a bundle. A marker with no names, and a marker written
// before the selection was recorded, read as no selection.
func initPendingBundles(projectRoot string) []string {
	data, ok := readManifestRecord(initPendingPath(projectRoot))
	if !ok {
		return nil
	}
	var names []string
	for _, line := range strings.Split(string(data), "\n")[1:] {
		if name := strings.TrimSpace(line); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// resumeBundleSelection returns the bundle selection the resumed ensure installs. An
// explicit --bundles is the current invocation's answer and wins; with none, the
// selection the failing run recorded is restored. The precedence mirrors
// resumeWiringFromRecord.
func resumeBundleSelection(cmd *cobra.Command, projectRoot string) []string {
	if raw := getStringFlag(cmd, "bundles"); strings.TrimSpace(raw) != "" {
		return parseBundleSelection(raw)
	}
	return initPendingBundles(projectRoot)
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
	initMarker := false
	if _, err := os.Stat(initPendingPath(projectRoot)); err == nil {
		interrupted = true
		initMarker = true
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
	//
	// Finding 1 (card t1591): the init marker outranks the deployment_mode read.
	// An init that stopped after ApplyDeployMode wrote deployment_mode, but before
	// the MCP and Codex wiring finished, still carries the marker: the project is
	// deployed and its init unfinished, so the rerun resumes. The llm.yaml presence
	// check stays first, so a project the executor never deployed keeps the original
	// redirect even with the marker standing. The journal and the corrupt manifest
	// are user-global evidence, so they keep the completeness read above.
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	llmData, ok := readManifestRecord(filepath.Join(sections, "llm.yaml"))
	if !ok {
		return false
	}
	if initMarker {
		return true
	}
	return !llmDeployModeRecorded(llmData)
}

// llmDeployModeRecorded reports whether llm.yaml records a deployment_mode key. The key is read
// from the parsed document as a mapping key at any depth: ApplyDeployMode matches the line at any
// indentation, so the record is wherever that writer put it. A comment that names the key does not
// record it, and a spaced colon does (card t1619, codex review round 3).
func llmDeployModeRecorded(llm []byte) bool {
	var doc yaml.Node
	if err := yaml.Unmarshal(llm, &doc); err != nil {
		return false
	}
	return yamlHasMappingKey(&doc, "deployment_mode")
}

// yamlHasMappingKey reports whether any mapping in the parsed tree has a key named name.
func yamlHasMappingKey(n *yaml.Node, name string) bool {
	if n == nil {
		return false
	}
	if n.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == name {
				return true
			}
		}
	}
	for _, child := range n.Content {
		if yamlHasMappingKey(child, name) {
			return true
		}
	}
	return false
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
