package cli

// update_codex_wiring.go — SPEC-CODEX-WIRING-001 M2, the update-path wiring
// refresh (REQ-CW-009). File existence is the user's standing opt-in: an
// update refreshes the Codex wiring ONLY in projects that already carry a
// wiring file (.codex/hooks.json or .codex/config.toml) and creates nothing
// in `--llm claude` / flag-absent projects.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// refreshCodexWiringBestEffortAt refreshes the Codex wiring of the project at
// projectRoot, existence-gated (REQ-CW-009). Best-effort (spec §F): a failure
// — including the REQ-CW-003 validation refusal — warns to errOut and the
// update continues; the hard part of the refusal (no violating bytes on
// disk) is already guaranteed by the codexwiring package.
//
// A wiring lock held by another live owner is its own outcome (REQ-DHR-002):
// nothing was changed, and the update says the wiring was not refreshed
// rather than that it failed.
//
// Wiring the configured harness profile no longer uses is not refreshed: it
// is reported with the command that removes it and left exactly as it is
// (REQ-DHR-007). Removal belongs to `moai tool disable codex` alone.
func refreshCodexWiringBestEffortAt(projectRoot string, out, errOut io.Writer) {
	if orphaned := orphanedCodexWiring(projectRoot); len(orphaned) > 0 {
		harness := config.ReadHarness(projectRoot)
		for _, rel := range orphaned {
			_, _ = fmt.Fprintf(errOut, "warning: %s is Codex wiring the %q harness profile does not use; left in place — run `%s` to remove it\n", rel, harness, codexwiring.DisableCommand)
		}
		return
	}
	if _, err := codexwiring.RefreshWiring(projectRoot, out, errOut); err != nil {
		if errOut == nil {
			return
		}
		if errors.Is(err, codexwiring.ErrWiringLockHeld) {
			_, _ = fmt.Fprintf(errOut, "warning: Codex wiring not refreshed: %v (%s); rerun `moai update` or `%s` once it is released\n", err, codexwiring.WiringLockRelPath, codexwiring.RecoverCommand)
			return
		}
		_, _ = fmt.Fprintf(errOut, "warning: Codex wiring refresh failed: %v\n", err)
	}
}

// refreshCodexWiringBestEffort is the runUpdate call-site form: runUpdate
// operates on the current working directory (".").
func refreshCodexWiringBestEffort(out, errOut io.Writer) {
	refreshCodexWiringBestEffortAt(".", out, errOut)
}

// addCodexWiringAt adds the Codex wiring of the project at projectRoot by
// calling the UNGATED codexwiring.Wire. `moai tool enable codex` creates
// wiring, so it intentionally bypasses the existence gate used by refresh.
//
// Error posture (plan §D1): a REQ-CW-003 validation refusal
// (ErrValidationRefused) propagates as a hard error and the tool command exits
// non-zero. The sibling refresh wrapper's warn-and-continue model is deliberately
// NOT followed here: best-effort is allowed only for IO errors, which warn and
// the update continues (spec §F posture). A held wiring lock and a wiring file
// that changed under the write (REQ-DHR-002/003) are outcomes the operator
// must see, so they exit non-zero too.
func addCodexWiringAt(projectRoot string, out, errOut io.Writer) error {
	if _, err := codexwiring.Wire(projectRoot, out, errOut); err != nil {
		if errors.Is(err, codexwiring.ErrValidationRefused) || errors.Is(err, codexwiring.ErrWiringLockHeld) || errors.Is(err, codexwiring.ErrWiringConflict) {
			return err
		}
		if errOut != nil {
			_, _ = fmt.Fprintf(errOut, "warning: Codex wiring failed: %v\n", err)
		}
	}
	return nil
}

// emitCodexWiringDryRunPreview prints the additive wiring actions for the
// named invocation. It writes nothing: the writer is its only parameter.
func emitCodexWiringDryRunPreview(out io.Writer, invocation string) {
	_, _ = fmt.Fprintf(out, "Dry-run %s wiring plan (nothing written):\n", invocation)
	_, _ = fmt.Fprintf(out, "  - create-or-refresh %s (merged hook render, whitelist-gated)\n", codexwiring.HooksRelPath)
	_, _ = fmt.Fprintf(out, "  - create-or-refresh %s ([mcp_servers.moai] + [tui].status_line, create-if-absent merge)\n", codexwiring.ConfigRelPath)
	_, _ = fmt.Fprintf(out, "  - create-or-refresh %s (trust sidecar, sha256 of the generated content)\n", codexwiring.SidecarPath)
	_, _ = fmt.Fprintln(out, "  - run without --dry-run to apply")
}

// codexHarness reports whether a harness profile deploys the Codex surfaces.
func codexHarness(harness string) bool {
	return harness == "both" || harness == "gpt"
}

// orphanedCodexWiring lists the wiring files MoAI owns in a project whose
// configured harness profile no longer uses Codex, after a profile that did
// deployed the Codex templates (the manifest records them). A project wired
// by `moai tool enable codex` on the claude profile never had those
// templates deployed, so its wiring is not orphaned. It reads only.
func orphanedCodexWiring(projectRoot string) []string {
	if codexHarness(config.ReadHarness(projectRoot)) {
		return nil
	}
	files, ok := readManifestFilesReadOnly(projectRoot)
	if !ok {
		return nil
	}
	recorded := false
	for path, e := range files {
		if strings.HasPrefix(path, ".codex/") && e.Provenance != manifest.GeneratedManaged {
			recorded = true
			break
		}
	}
	if !recorded {
		return nil
	}
	return codexwiring.OwnedWiringFiles(projectRoot)
}

// readManifestFilesReadOnly parses the manifest without Load's corrupt-file
// rename. ok is false when it is missing or unparseable.
func readManifestFilesReadOnly(projectRoot string) (map[string]manifest.FileEntry, bool) {
	raw, err := os.ReadFile(filepath.Join(projectRoot, defs.MoAIDir, defs.ManifestJSON))
	if err != nil {
		return nil, false
	}
	var mf manifest.Manifest
	if err := json.Unmarshal(raw, &mf); err != nil {
		return nil, false
	}
	return mf.Files, true
}

// undeployedCodexTemplates lists the .codex/ template paths the manifest
// records, that this deployment no longer ships, and that are still on disk.
// Generated wiring files are not template deployments and are excluded.
func undeployedCodexTemplates(projectRoot string, files map[string]manifest.FileEntry, deployed map[string]bool) []string {
	var out []string
	for path, e := range files {
		if !strings.HasPrefix(path, ".codex/") || e.Provenance == manifest.GeneratedManaged || deployed[path] {
			continue
		}
		if _, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(path))); err == nil {
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out
}

// reportUndeployedCodexTemplates reports each .codex/ template path the
// deployment no longer ships. It never deletes one (REQ-DHR-007).
func reportUndeployedCodexTemplates(w io.Writer, projectRoot string, files map[string]manifest.FileEntry, deployed map[string]bool) {
	harness := config.ReadHarness(projectRoot)
	for _, rel := range undeployedCodexTemplates(projectRoot, files, deployed) {
		_, _ = fmt.Fprintf(w, "warning: %s is no longer deployed by this update (harness profile %q); left in place — delete it by hand if you no longer need it\n", rel, harness)
	}
}
