// Package harness — `moai harness install` CLI surface.
//
// SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001 (REQ-HAW-001..005): this command is
// the live call path that wires the previously-orphaned installers
// internal/harness.InjectMarker (layer3, CLAUDE.md marker block) and
// internal/harness.ScaffoldHarnessDir (layer5, emits .moai/harness/main.md)
// into the post-generation flow. Before this command existed both functions
// had 0 non-test callers; the meta-harness Phase 7 (5-Layer Activation) now
// invokes `moai harness install` so a generated harness actually auto-triggers.
//
// HARD subagent boundary (REQ-HAW-003): no source file in this package may
// invoke AskUserQuestion — the orchestrator owns user interaction. This command
// takes positional flag inputs (--spec-id, --domain, --project-root) and emits
// structured errors; it never prompts. The boundary is enforced by the
// package-level guard TestPropose_NoAskUserQuestion (propose_boundary_test.go).
//
// This command does NOT rewrite the installer algorithms (EX-6 / D-4) — it is a
// thin caller that gives the wiring a test anchor so the dead-code path cannot
// silently recur.
package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// InstallOptions carries the inputs for the harness activation wiring.
type InstallOptions struct {
	// ProjectRoot is the absolute path to the project root. CLAUDE.md and
	// .moai/harness/ are resolved relative to this.
	ProjectRoot string
	// SpecID is the generating SPEC identifier (e.g. "SPEC-PROJ-INIT-001")
	// recorded in the CLAUDE.md marker block and the scaffolded main.md.
	SpecID string
	// Domain is the project domain (e.g. "ios-mobile") captured during the
	// Phase 5 Socratic interview.
	Domain string
	// IncludeDesignExtension mirrors ScaffoldOpts.IncludeDesignExtension —
	// when true an 8th design-extension.md file is scaffolded (Q13=Advanced).
	IncludeDesignExtension bool
}

// harnessMainImportPath is the CLAUDE.md @import target the marker block points
// at — the scaffolded entry point that the auto-trigger chain follows.
const harnessMainImportPath = ".moai/harness/main.md"

// RunInstall scaffolds the .moai/harness/ directory (REQ-HAW-005, emitting
// main.md) and injects the CLAUDE.md harness marker block (REQ-HAW-001/002),
// wiring the two existing installers into a single call path.
//
// Order: scaffold first (so main.md exists before the marker @import references
// it), then inject the marker. A failure at either step returns a wrapped
// structured error and reports NO success (REQ-HAW-004).
func RunInstall(opts InstallOptions) error {
	if opts.ProjectRoot == "" {
		return errors.New("harness install: empty project root")
	}

	// REQ-HAW-005: ensure .moai/harness/main.md exists via the existing
	// ScaffoldHarnessDir installer (layer5). This is the wiring fix — the
	// scaffolder was never invoked from a live flow before.
	harnessDir := filepath.Join(opts.ProjectRoot, ".moai", "harness")
	scaffoldOpts := harness.ScaffoldOpts{
		Domain:                 opts.Domain,
		SpecID:                 opts.SpecID,
		IncludeDesignExtension: opts.IncludeDesignExtension,
	}
	if err := harness.ScaffoldHarnessDir(harnessDir, scaffoldOpts); err != nil {
		return fmt.Errorf("harness install: scaffold .moai/harness/: %w", err)
	}

	// REQ-HAW-001/002: inject the CLAUDE.md routing marker block via the
	// existing InjectMarker installer (layer3). Idempotent — re-running
	// replaces the existing block rather than duplicating it. The import path
	// points at the just-scaffolded main.md so the auto-trigger chain has its
	// entry point. REQ-HAW-004: a CLAUDE.md read/write failure surfaces here as
	// a wrapped error and does NOT report success.
	claudeMdPath := filepath.Join(opts.ProjectRoot, "CLAUDE.md")
	if err := harness.InjectMarker(claudeMdPath, opts.SpecID, opts.Domain,
		[]string{harnessMainImportPath}); err != nil {
		return fmt.Errorf("harness install: inject CLAUDE.md marker: %w", err)
	}

	return nil
}

// NewInstallCmd is the `moai harness install` cobra factory.
//
// The factory is exported so newHarnessRouterCmd() in
// internal/cli/harness_route.go (package cli) can register the subcommand under
// the single `moai harness` parent, mirroring how NewProposeCmd() is wired.
func NewInstallCmd() *cobra.Command {
	var (
		specID      string
		domain      string
		projectRoot string
		designExt   bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Wire a generated harness: scaffold .moai/harness/main.md + install CLAUDE.md markers",
		Long: `Activate a generated project harness by wiring the two installers that the
meta-harness generation flow relies on:

  1. Scaffold .moai/harness/ (emitting main.md — the CLAUDE.md @import entry
     point and task-shape router).
  2. Inject the CLAUDE.md '<!-- moai:harness-start -->' / '<!-- moai:harness-end -->'
     routing marker block (idempotent — a re-install replaces the existing block).

This subcommand never invokes AskUserQuestion. It takes positional flags
(--spec-id, --domain, --project-root) and emits structured errors. User
interaction is owned exclusively by the orchestrator per the subagent boundary
HARD contract.

Examples:
  moai harness install --spec-id SPEC-PROJ-INIT-001 --domain ios-mobile
  moai harness install --spec-id SPEC-X --domain web --project-root /path/to/proj`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := projectRoot
			if root == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("harness install: resolve project root: %w", err)
				}
				root = wd
			} else {
				// Resolve user-supplied paths absolutely (internal/cli/CLAUDE.md
				// absolute-path rule — never filepath.Join(cwd, userPath)).
				abs, err := filepath.Abs(root)
				if err != nil {
					return fmt.Errorf("harness install: resolve --project-root: %w", err)
				}
				root = abs
			}

			opts := InstallOptions{
				ProjectRoot:            root,
				SpecID:                 specID,
				Domain:                 domain,
				IncludeDesignExtension: designExt,
			}
			if err := RunInstall(opts); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(),
				"harness activation wired: scaffolded %s + installed CLAUDE.md markers (spec=%s domain=%s)\n",
				filepath.Join(root, ".moai", "harness"), specID, domain)
			return nil
		},
	}

	cmd.Flags().StringVar(&specID, "spec-id", "",
		"Generating SPEC ID recorded in the marker block + main.md (required)")
	cmd.Flags().StringVar(&domain, "domain", "",
		"Project domain captured during the Phase 5 interview (e.g. ios-mobile)")
	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"Project root path (default: current directory)")
	cmd.Flags().BoolVar(&designExt, "design-extension", false,
		"Also scaffold design-extension.md (Q13=Advanced branch)")

	if err := cmd.MarkFlagRequired("spec-id"); err != nil {
		panic(fmt.Sprintf("harness install: MarkFlagRequired: %v", err))
	}

	return cmd
}
