// bundle.go — the `moai bundle add|remove` command (SPEC-USER-ASSET-INSTALL-001
// REQ-004). Bundle membership is declared in the catalog; the user's opt-in
// SELECTION is the manifest's bundles: list, adjusted here and honored by
// `moai update` (M3).
//
// add installs exactly the named bundle's entries (REQ-010 collision and
// REQ-023 divergence semantics ride the shared installer judgment).
// remove applies the COMPLEMENT rule (in-round extension E3): the removal
// target is the entries of the removed bundle NOT in (L0 ∪ the remaining
// opted-in selections) — entries the removed bundle SHARES with L0 or a
// still-opted bundle survive with a report note. An entry that is a declared
// dependency of a PRESERVED asset has its deletion DEFERRED (kept + reported,
// R-f-②) — the preserved assets' consumers are not orphaned. The dispatcher's
// own dep list is excluded from the deferral rule: its bundle-row edges are
// the derivation matrix's conditional class, their absence is handled by the
// remediation/refusal pattern, and including them would defeat the D28
// selection-based prune (AC-018's shipped-but-deselected arm requires the
// prune to function).
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"

	"github.com/spf13/cobra"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Manage opt-in common-asset bundles (user-folder install)",
	Long: `Add or remove opt-in bundles of common skills and agents.

Bundles install into your user folders (~/.claude/skills, ~/.claude/agents,
~/.agents/skills, ~/.codex/agents) and are recorded in
~/.moai/user-assets.json. "moai update" honors the recorded selection.`,
}

var bundleAddCmd = &cobra.Command{
	Use:   "add <bundle-name>",
	Short: "Install an opt-in bundle's entries into your user folders",
	Args:  cobra.ExactArgs(1),
	RunE:  runBundleAdd,
}

var bundleRemoveCmd = &cobra.Command{
	Use:   "remove <bundle-name>",
	Short: "Remove a bundle's entries (entries shared with L0 or other selections survive)",
	Args:  cobra.ExactArgs(1),
	RunE:  runBundleRemove,
}

func init() {
	bundleCmd.AddCommand(bundleAddCmd)
	bundleCmd.AddCommand(bundleRemoveCmd)
	rootCmd.AddCommand(bundleCmd)
}

func newBundleTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "bundle-test"}
	cmd.SetOut(os.Stdout)
	return cmd
}

func runBundleAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	homeDir, err := userHomeDirFn()
	if err != nil {
		return fmt.Errorf("resolve home: %w", err)
	}
	lock, err := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if err != nil {
		return fmt.Errorf("acquire user-asset lock: %w", err)
	}
	defer func() { _ = lock.Release() }()

	manifest, err := userassets.Load(userassets.ManifestPath(homeDir))
	if err != nil {
		return err
	}
	if bundleExists(name) {
		manifest.Bundles = addSelection(manifest.Bundles, name)
	} else {
		return fmt.Errorf("unknown bundle %q — valid bundles: %s", name, strings.Join(bundleNames(), ", "))
	}

	inst, err := newUserAssetInstaller(homeDir)
	if err != nil {
		return err
	}
	res, err := inst.Install(manifest.Bundles)
	if err != nil {
		return err
	}
	writeInstallSummary(cmd.OutOrStdout(), fmt.Sprintf("bundle add %s", name), res)
	return nil
}

func runBundleRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	homeDir, err := userHomeDirFn()
	if err != nil {
		return fmt.Errorf("resolve home: %w", err)
	}
	lock, err := userassets.AcquireUserLock(homeDir, userLockWaitWindow)
	if err != nil {
		return fmt.Errorf("acquire user-asset lock: %w", err)
	}
	defer func() { _ = lock.Release() }()

	manifestPath := userassets.ManifestPath(homeDir)
	manifest, err := userassets.Load(manifestPath)
	if err != nil {
		return err
	}
	if !bundleExists(name) {
		return fmt.Errorf("unknown bundle %q — valid bundles: %s", name, strings.Join(bundleNames(), ", "))
	}

	remaining := dropSelection(manifest.Bundles, name)
	inst, err := newUserAssetInstaller(homeDir)
	if err != nil {
		return err
	}

	// RemoveBundle applies the E3 complement + R-f-② deferral + the one
	// REQ-009 removal rule, then the caller saves the manifest with the
	// remaining selection (the save is why the lock spans this whole body).
	res, err := inst.RemoveBundle(manifest, name, remaining)
	if err != nil {
		return err
	}
	manifest.Bundles = remaining
	if err := manifest.Save(manifestPath); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}
	writeInstallSummary(cmd.OutOrStdout(), fmt.Sprintf("bundle remove %s", name), res)
	return nil
}

func addSelection(sel []string, name string) []string {
	for _, s := range sel {
		if s == name {
			return sel
		}
	}
	return append(sel, name)
}

func dropSelection(sel []string, name string) []string {
	out := make([]string, 0, len(sel))
	for _, s := range sel {
		if s != name {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func bundleExists(name string) bool {
	_, ok := loadCatalogForBundle().Catalog.OptionalPacks[name]
	return ok
}

func bundleNames() []string {
	packs := loadCatalogForBundle().Catalog.OptionalPacks
	names := make([]string, 0, len(packs))
	for n := range packs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func loadCatalogForBundle() *template.Catalog {
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		// The command paths already surface catalog failures through
		// newUserAssetInstaller; this lookup keeps the name check harmless
		// when the embed is broken.
		return &template.Catalog{Catalog: template.CatalogSections{
			OptionalPacks: map[string]*template.Pack{},
		}}
	}
	return cat
}
