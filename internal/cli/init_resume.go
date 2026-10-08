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
	"errors"
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
// the user-asset ensure: the ensure shortfall runs (the bundle selection
// the re-run names applies), then the setup steps the failed attempt never
// reached, in the production order (ApplyHarness → MCP entry → Codex
// wiring).
func resumeInitializedProject(cmd *cobra.Command, opts *project.InitOptions, wiring agentWiring) error {
	homeDir, homeErr := userHomeDirFn()
	if homeErr != nil {
		return fmt.Errorf("resolve user home: %w", homeErr)
	}
	// A CORRUPT user manifest (a real ensure-failure shape: the write died
	// mid-record) is quarantined aside — the doctor's sanctioned
	// rebuild-from-fresh-init recovery for exactly this state — so the
	// resume ensure can complete from scratch. Healthy and absent
	// manifests are left alone.
	quarantineCorruptUserManifest(homeDir)

	if err := ensureUserAssetsLocked(homeDir, parseBundleSelection(getStringFlag(cmd, "bundles")), cmd.OutOrStdout()); err != nil {
		return fmt.Errorf("user-asset ensure: %w", err)
	}
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
// checkpoint justifies the init resume (gate round 35-3): a pending-
// install journal (an attempt interrupted mid-install) or a CORRUPT user
// manifest (an attempt that failed before/at the ensure). A healthy
// project's ordinary re-run has neither — it keeps the update redirect.
func initResumeCheckpoint(homeDir string) bool {
	if j, err := userassets.LoadJournal(userassets.JournalPath(homeDir)); err == nil && j != nil {
		return true
	}
	_, err := userassets.Load(userassets.ManifestPath(homeDir))
	var ce *userassets.CorruptError
	return errors.As(err, &ce)
}

// quarantineCorruptUserManifest renames a CORRUPT user manifest aside with
// a timestamped suffix (never deletes). Absent, healthy, and unreadable-
// for-other-reasons manifests are left untouched.
func quarantineCorruptUserManifest(homeDir string) {
	manifestPath := userassets.ManifestPath(homeDir)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return // absent or unreadable for other reasons — not ours to move
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
