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
	"time"

	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
	"github.com/spf13/cobra"
)

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
