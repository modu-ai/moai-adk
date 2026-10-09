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
	"time"

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
func resumeInitializedProject(cmd *cobra.Command, opts *project.InitOptions, wiring agentWiring) error {
	homeDir, homeErr := userHomeDirFn()
	if homeErr != nil {
		return fmt.Errorf("resolve user home: %w", homeErr)
	}
	lock, lockErr := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if lockErr != nil {
		return fmt.Errorf("user-asset resume lock: %w", lockErr)
	}
	defer func() { _ = lock.Release() }()

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
	// DEBT R4: the resume is judged by CONTENT downstream — the harness
	// config, the MCP entry, and the Codex wiring the failed attempt never
	// wrote are all written here, in the production order.
	if err := template.ApplyHarness(opts.ProjectRoot, string(wiring)); err != nil {
		return fmt.Errorf("apply harness: %w", err)
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
	return nil
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
	sections := filepath.Join(projectRoot, ".moai", "config", "sections")
	if _, err := os.Stat(filepath.Join(sections, "llm.yaml")); err != nil {
		return false
	}
	return true
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
