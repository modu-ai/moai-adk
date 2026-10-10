// user_asset_phase.go — the per-user common-asset phase driver
// (SPEC-USER-ASSET-INSTALL-001). One shared engine serves the three callers:
// `moai init` (the first-install trigger, REQ-024), `moai bundle add|remove`
// (REQ-004 — landed in M2 per the leader's mid-run sequencing decision), and
// `moai update`'s user-asset phase (M3).
//
// The user lock spans manifest read → asset changes → manifest save
// (round-5 F4). All calls resolve the user home through userHomeDirFn — the
// same seam update.go already carries — so tests run against temp HOMEs.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// ensureUserAssetsLocked runs the REQ-024 per-asset-state ensure under the
// user lock and reports the REQ-011-style summary to out. Systemic failures
// (lock, catalog, roots) return an error; per-file failures continue and are
// surfaced in the summary (REQ-013).
func ensureUserAssetsLocked(homeDir string, selection []string, out io.Writer) error {
	lock, err := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if err != nil {
		return fmt.Errorf("acquire user-asset lock: %w", err)
	}
	defer func() { _ = lock.Release() }()

	inst, err := newUserAssetInstaller(homeDir)
	if err != nil {
		return err
	}
	// RF9 (review fix): a second project's init passes the DEFAULT empty
	// selection — the recorded selection must be UNIONed, never wiped.
	res, err := inst.InstallPreserveSelection(selection)
	if err != nil {
		return err
	}
	writeInstallSummary(out, "user-asset ensure", res)
	return nil
}

// userLockWaitWindow bounds how long a run waits on the user-level lock.
// Generous: a cold-tree first install can legitimately take minutes.
const userLockWaitWindow = 2 * time.Minute

// newUserAssetInstaller builds the installer over the embedded catalog and
// tree.
func newUserAssetInstaller(homeDir string) (*userassets.Installer, error) {
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		return nil, fmt.Errorf("load catalog: %w", err)
	}
	src, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("embedded templates: %w", err)
	}
	return &userassets.Installer{
		Home:        homeDir,
		Catalog:     cat,
		Source:      src,
		MoaiVersion: version.GetVersion(),
	}, nil
}

// writeInstallSummary renders the REQ-011 count categories plus the
// actionable failure/collision/divergence rows (C4: path + reason + action).
func writeInstallSummary(out io.Writer, label string, res *userassets.Result) {
	if res == nil {
		return
	}
	_, _ = fmt.Fprintf(out, "%s: %d installed, %d refreshed, %d removed, %d collision-skipped, %d divergence-preserved\n",
		label, res.Installed, res.Refreshed, res.Removed, res.CollisionSkipped, res.DivergencePreserved)
	for _, c := range res.Collisions {
		_, _ = fmt.Fprintf(out, "  collision (skipped, your file kept): %s — rename it or remove it, then re-run to install\n", c)
	}
	for _, d := range res.Divergences {
		_, _ = fmt.Fprintf(out, "  divergence (preserved, shipped copy backed up under ~/.moai/backups/): %s\n", d)
	}
	for _, f := range res.Failures {
		_, _ = fmt.Fprintf(out, "  failure: %s: %s\n", f.Path, f.Reason)
	}
	for _, s := range res.SharedSurvivors {
		_, _ = fmt.Fprintf(out, "  kept (shared with L0 or another opted-in bundle): %s\n", s)
	}
	for _, d := range res.DeferredDeps {
		_, _ = fmt.Fprintf(out, "  kept (declared dependency of a preserved asset — re-evaluated at the next removal/update): %s\n", d)
	}
	for _, u := range res.Unconverted {
		_, _ = fmt.Fprintf(out, "  unconverted reference (no Codex-side mapping — report to the harness maintainers): %s\n", u)
	}
}

// runUserAssetUpdatePhase is `moai update`'s user-asset phase (SPEC-USER-
// ASSET-INSTALL-001 M3; design §2.4): refresh (REQ-008 — the installer's
// manifest-match arm), the selection-based prune (REQ-009 — the recorded
// selection is honored; a `moai bundle remove`d bundle's artifacts go),
// REQ-023 divergence handling and the REQ-011 summary. It runs BEFORE the
// project phase so the upgrade install precedes the migration removal
// (REQ-024's upgrade arm ordering). The recorded bundle selection is
// preserved — the phase passes the manifest's own list back as the selection.
func runUserAssetUpdatePhase(homeDir string, out io.Writer) error {
	lock, err := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if err != nil {
		return fmt.Errorf("acquire user-asset lock: %w", err)
	}
	defer func() { _ = lock.Release() }()

	inst, err := newUserAssetInstaller(homeDir)
	if err != nil {
		return err
	}
	manifestPath := userassets.ManifestPath(homeDir)
	manifest, err := userassets.Load(manifestPath)
	if err != nil {
		return err
	}

	// Refresh + missing-counterpart install over L0 ∪ the recorded selection.
	ensureRes, err := inst.Install(manifest.Bundles)
	if err != nil {
		return err
	}

	// The selection-based prune against the freshly saved manifest.
	saved, err := userassets.Load(manifestPath)
	if err != nil {
		return err
	}
	pruneRes, err := inst.PruneUnselected(saved)
	if err != nil {
		return err
	}
	if err := saved.Save(manifestPath); err != nil {
		return fmt.Errorf("save user manifest: %w", err)
	}

	writeInstallSummary(out, "user-asset update", ensureRes)
	if pruneRes != nil {
		_, _ = fmt.Fprintf(out, "user-asset prune: %d removed, %d divergence-preserved\n",
			pruneRes.Removed, pruneRes.DivergencePreserved)
		for _, c := range pruneRes.Collisions {
			_, _ = fmt.Fprintf(out, "  collision (kept): %s\n", c)
		}
		for _, d := range pruneRes.Divergences {
			_, _ = fmt.Fprintf(out, "  divergence (preserved, shipped copy backed up): %s\n", d)
		}
		for _, f := range pruneRes.Failures {
			_, _ = fmt.Fprintf(out, "  failure: %s: %s\n", f.Path, f.Reason)
		}
	}
	return nil
}

// userCodexInstallPresent reports whether the per-user Codex dispatcher
// install exists (SPEC-USER-ASSET-INSTALL-001: the user-folder primary).
func userCodexInstallPresent() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".agents", "skills", "moai", "SKILL.md"))
	return err == nil
}

// updatePluginOptedOut reads MOAI_SKIP_PLUGIN_INSTALL (the retired
// carrier's opt-out env, kept as a no-op-shaped input for the deploy-mode
// resolution until M7 removes the mode split entirely).
func updatePluginOptedOut() bool {
	v := strings.TrimSpace(os.Getenv("MOAI_SKIP_PLUGIN_INSTALL"))
	return v == "1" || strings.EqualFold(v, "true")
}

// setMigrationPreservedFiles replaces the migration-preserved set (test seam).
func setMigrationPreservedFiles(m map[string]bool) {
	migrationPreservedProjectFiles = m
}

// parseBundleSelection splits a comma-separated --bundles value.
func parseBundleSelection(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	var out []string
	for _, name := range strings.Split(v, ",") {
		if name = strings.TrimSpace(name); name != "" {
			out = append(out, name)
		}
	}
	return out
}
