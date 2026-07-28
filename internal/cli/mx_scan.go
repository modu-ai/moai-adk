package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// defaultScanIgnore are directory base-names excluded from a full project scan.
// The scanner matches these against filepath.Base(path) for directories
// (scanner.go ScanDir), so these are plain names — no globbing across path
// separators is needed or supported there.
var defaultScanIgnore = []string{
	".git",
	"node_modules",
	"vendor",
	"worktrees", // harness worktrees (recursive copies of the repo)
	"main-fork",  // moai-adk local fork (not distributed source)
	"target",     // rust/cargo build output
	"build",      // java/gradle/kotlin build output
	"dist",       // generic build output
	"out",
	".next",    // next.js build cache
	".cache",
	".turbo",
	"coverage", // test coverage reports
	// Harness / state directories hold no user source code.
	".claude",
	".moai",
}

// newMxScanCmd 'moai mx scan' scans the project and builds the @MX tag sidecar
// index in one pass. It is the build entry-point that connects scanner.ScanDir
// to Manager.Write — without it, 'moai mx query' has no index to read.
func newMxScanCmd() *cobra.Command {
	var pathArg string
	var dryRun bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the project and build the @MX tag sidecar index",
		Long: `Scan source files for @MX tags and write the sidecar index that 'moai mx query' reads.

This builds .moai/state/mx-index.json in one pass. Run it once after checkout, or
any time you want a fresh full index. After it runs, 'moai mx query' returns results.

Examples:
  moai mx scan                 # full project scan -> .moai/state/mx-index.json
  moai mx scan --dry           # preview tag counts without writing the index
  moai mx scan --path internal # scan only one subtree`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("failed to find project root: %w", err)
			}

			scanRoot := projectRoot
			if pathArg != "" {
				abs, err := filepath.Abs(pathArg)
				if err != nil {
					return fmt.Errorf("resolve --path: %w", err)
				}
				scanRoot = abs
			}

			s := mx.NewScanner()
			s.SetIgnorePatterns(defaultScanIgnore)

			tags, err := s.ScanDir(scanRoot)
			if err != nil {
				return fmt.Errorf("scan %s: %w", scanRoot, err)
			}

			// Summary counts (cheap, single pass over the tag slice).
			counts := make(map[string]int)
			rotRisk := 0
			for _, t := range tags {
				counts[string(t.Kind)]++
				if t.RotRisk == "no-trigger" {
					rotRisk++
				}
			}

			out := cmd.OutOrStdout()
			errs := cmd.ErrOrStderr()

			if dryRun {
				_, _ = fmt.Fprintf(out, "DRY RUN: %d tags would be written (index not saved)\n", len(tags))
				printMxScanSummary(out, counts, rotRisk)
				return nil
			}

			stateDir := filepath.Join(projectRoot, ".moai", "state")
			mgr := mx.NewManager(stateDir)
			sidecar := &mx.Sidecar{
				SchemaVersion: mx.SchemaVersion,
				Tags:          tags,
				ScannedAt:     time.Now(),
			}
			if err := mgr.Write(sidecar); err != nil {
				return fmt.Errorf("write sidecar: %w", err)
			}

			_, _ = fmt.Fprintf(out, "OK: wrote %d tags to %s\n", len(tags), filepath.Join(stateDir, mx.SidecarFileName))
			if !quiet {
				printMxScanSummary(out, counts, rotRisk)
			}

			// Surface scanner-detected issues to stderr (advisory, never blocks).
			if warns := s.GetWarnings(); len(warns) > 0 && !quiet {
				_, _ = fmt.Fprintf(errs, "scanner warnings: %d\n", len(warns))
				for i, w := range warns {
					if i >= 20 {
						_, _ = fmt.Fprintf(errs, "  ... and %d more\n", len(warns)-20)
						break
					}
					_, _ = fmt.Fprintf(errs, "  - %s\n", w)
				}
			}
			if scanErrs := s.GetErrors(); len(scanErrs) > 0 && !quiet {
				_, _ = fmt.Fprintf(errs, "scan errors: %d (non-fatal, tags still written)\n", len(scanErrs))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&pathArg, "path", "", "scan only this subtree (absolute or project-relative path)")
	cmd.Flags().BoolVar(&dryRun, "dry", false, "preview tag counts without writing the index")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress the per-kind summary and scanner warnings")

	return cmd
}

// printMxScanSummary writes a stable per-kind breakdown to the given writer.
func printMxScanSummary(out io.Writer, counts map[string]int, rotRisk int) {
	order := []string{"NOTE", "ANCHOR", "WARN", "TODO", "DEBT", "LEGACY"}
	_, _ = fmt.Fprintln(out, "by kind:")
	for _, k := range order {
		if c := counts[k]; c > 0 {
			_, _ = fmt.Fprintf(out, "  %s: %d\n", k, c)
		}
	}
	if rotRisk > 0 {
		_, _ = fmt.Fprintf(out, "DEBT rotRisk (missing @MX:UPGRADE): %d\n", rotRisk)
	}
}
