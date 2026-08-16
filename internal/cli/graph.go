package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// newGraphCmd creates the 'moai graph' parent command — persisted codebase
// graph artifacts. Registered from this file's own init() so the command
// lands without editing any existing CLI registration surface.
//
// @MX:NOTE: [AUTO] moai graph build — top-level command (not `moai mx export`) because the edge aggregate spans codemaps/@MX/SPEC layers, exceeding the @MX sidecar scope
func newGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "graph",
		Short:   "Codebase graph artifact tools",
		Long:    `Build persisted graph artifacts that aggregate existing graph-producing layers.`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newGraphBuildCmd())
	return cmd
}

// newGraphBuildCmd 'moai graph build' aggregates the three existing edge
// layers (codemaps import adjacency, @MX:SPEC sub-lines, SPEC depends_on
// frontmatter) into one git-diffable JSONL artifact.
func newGraphBuildCmd() *cobra.Command {
	var outPath string
	var rootArg string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Aggregate codemaps/@MX/SPEC edges into edges.jsonl",
		Long: `Aggregate the three existing graph-producing layers into edges.jsonl (one JSON edge per line, sorted, git-diffable):

  import        package -> package   from .moai/project/codemaps/dependencies.md
  mx-spec       file -> SPEC         from @MX:SPEC sub-lines (mx scanner)
  spec-depends  SPEC -> SPEC         from spec.md frontmatter depends_on

Fails open per layer: an absent layer contributes zero edges.

Examples:
  moai graph build                                  # -> .moai/project/graph/edges.jsonl
  moai graph build --out my-edges.jsonl             # custom output path
  moai graph build --root /path/to/project          # explicit project root`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveGraphRoot(rootArg)
			if err != nil {
				return err
			}

			edges, err := graph.Build(projectRoot)
			if err != nil {
				return fmt.Errorf("build graph: %w", err)
			}

			target := outPath
			if target == "" {
				target = filepath.Join(projectRoot, ".moai", "project", "graph", "edges.jsonl")
			} else {
				abs, absErr := filepath.Abs(target)
				if absErr != nil {
					return fmt.Errorf("resolve --out: %w", absErr)
				}
				target = abs
			}
			if err := graph.WriteJSONL(target, edges); err != nil {
				return fmt.Errorf("write edges: %w", err)
			}

			out := cmd.OutOrStdout()
			counts := map[string]int{}
			for _, e := range edges {
				counts[e.Kind]++
			}
			_, _ = fmt.Fprintf(out, "OK: wrote %d edges to %s\n", len(edges), target)
			for _, kind := range []string{graph.KindImport, graph.KindMXSpec, graph.KindSpecDepends} {
				if c := counts[kind]; c > 0 {
					_, _ = fmt.Fprintf(out, "  %s: %d\n", kind, c)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&outPath, "out", "", "output path (defaults to <root>/.moai/project/graph/edges.jsonl)")
	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")

	return cmd
}

// resolveGraphRoot resolves the project root: an explicit --root wins (made
// absolute), otherwise the shared findProjectRootFn auto-detection.
func resolveGraphRoot(rootArg string) (string, error) {
	if rootArg == "" {
		root, err := findProjectRootFn()
		if err != nil {
			return "", fmt.Errorf("failed to find project root: %w", err)
		}
		return root, nil
	}
	abs, err := filepath.Abs(rootArg)
	if err != nil {
		return "", fmt.Errorf("resolve --root: %w", err)
	}
	return abs, nil
}

func init() {
	rootCmd.AddCommand(newGraphCmd())
}
