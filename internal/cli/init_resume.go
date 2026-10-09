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

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/homestate"
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
// autonomy tier → Jev record → ApplyHarness → deploy mode → git hooks →
// MCP entry → Codex wiring; gate round 35-6 included). Gate round 47-4:
// the wizard's Jev answer rides in — a resume that skipped it left
// jev.enabled false against an explicit wizard selection.
func resumeInitializedProject(cmd *cobra.Command, opts *project.InitOptions, wiring agentWiring, wizardRan bool, wizardResult *wizard.WizardResult) error {
	homeDir, homeErr := userHomeDirFn()
	if homeErr != nil {
		return fmt.Errorf("resolve user home: %w", homeErr)
	}
	lock, lockErr := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if lockErr != nil {
		return fmt.Errorf("user-asset resume lock: %w", lockErr)
	}
	defer func() { _ = lock.Release() }()

	// Gate round 51-3: the deployed LAYOUT is the interrupted record of the
	// original --llm selection. A re-run without options resolves the claude
	// default; applying it over a deployed GPT layout records harness:claude
	// while the Codex wiring is skipped — config and on-disk layout stay
	// permanently mismatched. The codex-only deployer HIDES the claude-only
	// surfaces (CLAUDE.md among them) and projects AGENTS.md, so
	// AGENTS.md-present-without-CLAUDE.md IS the gpt record.
	if wiring == agentWiringClaude && deployedWiringIsGPT(opts.ProjectRoot) {
		wiring = agentWiringGPT
	}

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
	wireCodexUnlessClaude(cmd, wiring, opts.ProjectRoot)
	// Gate round 51-2: every post-step completed — the checkpoint marker
	// retires (the early-return paths above leave it in place, keeping the
	// next run's resume route alive).
	_ = os.Remove(markerPath)
	return nil
}

// deployedWiringIsGPT reads the deployed layout as the interrupted --llm
// record (gate round 51-3): the codex-only deployer hides CLAUDE.md and
// projects AGENTS.md, so their presence pattern names the deployer that
// ran.
func deployedWiringIsGPT(projectRoot string) bool {
	if _, err := os.Stat(filepath.Join(projectRoot, "CLAUDE.md")); err == nil {
		return false // the claude deployer ran
	}
	_, err := os.Stat(filepath.Join(projectRoot, "AGENTS.md"))
	return err == nil
}

// initResumeCheckpoint reports whether an explicit interruption
// checkpoint justifies the init resume (gate round 35-3):
//   - a PENDING-INSTALL JOURNAL (an attempt interrupted mid-install), or
//   - a CORRUPT user manifest (an attempt that failed before/at the
//     ensure),
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
