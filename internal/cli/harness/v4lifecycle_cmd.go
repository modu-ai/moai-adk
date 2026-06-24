// Package harness — v4 lifecycle cobra command wrappers (SPEC-V3R6-HARNESS-V4-001 M4).
//
// NewHarnessV4ListCmd / NewHarnessV4EditCmd / NewHarnessV4RemoveCmd are the
// cobra command factories for the three v4 lifecycle verbs. They are registered
// under the existing `moai harness` parent (newHarnessRouterCmd in package cli)
// alongside the V3R2/V3R5 verbs. Each factory is a thin wrapper that resolves
// --project-root and delegates to the pure functions in v4lifecycle.go.
//
// The factories live in the boundary-guarded internal/cli/harness/ package so
// the C-HRA-008 static guard (TestPropose_NoAskUserQuestion) scans them. They
// MUST NOT call AskUserQuestion — the CLI surfaces structured output and
// stderr errors; the orchestrator owns user interaction.
package harness

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// resolveProjectRootV4 returns the --project-root flag value or the current
// working directory. Mirrors the resolveProjectRoot helper in package cli but
// is local to this package to keep the lifecycle handlers self-contained.
func resolveProjectRootV4(cmd *cobra.Command) (string, error) {
	root, _ := cmd.Flags().GetString("project-root")
	if root == "" {
		// Inherited --project-root from the parent harness command.
		if f := cmd.InheritedFlags().Lookup("project-root"); f != nil {
			root = f.Value.String()
		}
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("v4lifecycle: resolve project root: %w", err)
		}
	}
	return root, nil
}

// NewHarnessV4ListCmd is the `moai harness list` factory (AC-HV4-011a).
// Enumerates every harness under .claude/commands/harness/ joined with its
// manifest. Supports --json for machine-readable output.
func NewHarnessV4ListCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all v4 harnesses",
		Long: `List every harness-v4 entry under .claude/commands/harness/.

Each harness is shown with its name, domain (from manifest.json), and entry
command. A command file whose manifest is missing is still listed with a
manifest_missing flag so partial state is visible.

Use --json for machine-readable output.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRootV4(cmd)
			if err != nil {
				return err
			}
			entries, err := ListHarnesses(root)
			if err != nil {
				return err
			}
			if jsonOutput {
				data, mErr := json.MarshalIndent(entries, "", "  ")
				if mErr != nil {
					return fmt.Errorf("harness list: json marshal: %w", mErr)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}
			w := cmd.OutOrStdout()
			if len(entries) == 0 {
				_, _ = fmt.Fprintln(w, "No harnesses found under .claude/commands/harness/.")
				return nil
			}
			_, _ = fmt.Fprintf(w, "%-16s %-40s %s\n", "NAME", "DOMAIN", "ENTRY")
			for _, e := range entries {
				domain := e.Domain
				if domain == "" {
					domain = "(manifest missing)"
				}
				entry := e.EntryCommand
				if entry == "" {
					entry = "/harness:" + e.Name
				}
				_, _ = fmt.Fprintf(w, "%-16s %-40s %s\n", e.Name, domain, entry)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

// NewHarnessV4EditCmd is the `moai harness edit <name>` factory (design §B.3).
// Locates the manifest + specialist + skill files for the named harness and
// prints their paths so the user (or orchestrator) can open them for editing.
func NewHarnessV4EditCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Show paths to edit a v4 harness manifest + specialists",
		Long: `Show the file paths to edit for a harness-v4 entry.

The manifest is the single source of truth — editing it propagates to Runner
behavior on the next invocation. Specialist agent files and companion skill
directories are also listed so role definitions can be revised alongside.

Use --json for machine-readable output.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveProjectRootV4(cmd)
			if err != nil {
				return err
			}
			paths, err := EditHarness(root, args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				data, mErr := json.MarshalIndent(paths, "", "  ")
				if mErr != nil {
					return fmt.Errorf("harness edit: json marshal: %w", mErr)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Harness: %s\n", paths.Name)
			_, _ = fmt.Fprintf(w, "Manifest (SSOT): %s\n", paths.ManifestPath)
			if len(paths.SpecialistPaths) > 0 {
				_, _ = fmt.Fprintln(w, "Specialists:")
				for _, p := range paths.SpecialistPaths {
					_, _ = fmt.Fprintf(w, "  - %s\n", p)
				}
			}
			if len(paths.SkillPaths) > 0 {
				_, _ = fmt.Fprintln(w, "Skills:")
				for _, p := range paths.SkillPaths {
					_, _ = fmt.Fprintf(w, "  - %s\n", p)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

// NewHarnessV4RemoveCmd is the `moai harness remove <name>` factory
// (AC-HV4-011b/c). Atomically removes command + workflow + specialists +
// skills + manifest. Fails closed if any referenced artifact is missing.
func NewHarnessV4RemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Atomically remove a v4 harness (command + workflow + specialists + skills + manifest)",
		Long: `Remove a harness-v4 entry and all its artifacts.

Removes ALL of the following atomically:
  - .claude/commands/harness/<name>.md (thin-wrapper command)
  - .claude/commands/harness/<name>/manifest.json (manifest SSOT)
  - .claude/workflows/harness-<name>-run.js (Runner Workflow)
  - .claude/agents/harness/harness-<name>*-specialist.md (specialists)
  - .claude/skills/harness-<name>*/ (companion skills)

Fail-closed: if any referenced artifact is missing (orphan state), the remove
refuses to proceed and names the missing artifact. No partial state is left.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveProjectRootV4(cmd)
			if err != nil {
				return err
			}
			if err := RemoveHarness(root, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "harness %q removed (command + workflow + specialists + skills + manifest).\n", args[0])
			return nil
		},
	}
	return cmd
}
