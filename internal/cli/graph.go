package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/mx"
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
		Long:    `Build and query persisted codebase graph artifacts (edges.jsonl).`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newGraphBuildCmd())
	cmd.AddCommand(newGraphQueryCmd())
	return cmd
}

// unreferencedSpecCaveat is printed with every --specs-no-code result: a
// SPEC absent from edges.jsonl is NOT an unimplemented SPEC — most SPECs
// deliver docs/rules/harness with no code by design.
const unreferencedSpecCaveat = "NOTE: unreferenced != unimplemented (미연결 ≠ 미구현) — most SPECs deliver docs/rules/harness, so\nhaving no code reference is normal; treat this list as a coverage map, not a defect list."

// newGraphQueryCmd 'moai graph query' answers reverse-dependency questions
// from the persisted edges.jsonl without re-running extraction.
//
// @MX:NOTE: [AUTO] moai graph query — read-only consumer of edges.jsonl; agents use it instead of re-running go list -deps per question
func newGraphQueryCmd() *cobra.Command {
	var callersNode, blastNode, edgesPath, rootArg string
	var fanin, specsNoCode bool
	var limit int

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query edges.jsonl: callers / blast radius / import fan-in / SPECs with no code reference",
		Long: `Answer reverse-dependency questions from the persisted edges.jsonl (run 'moai graph build' first). Exactly one selector per invocation:

  --callers <node>    direct reverse neighbors: importers of a package, SPECs
                      depending on a SPEC, code files tagged @MX:SPEC
  --blast <node>      transitive blast radius of a change at <node> (BFS over
                      reverse edges; mx-spec edges propagate both ways, so a
                      code file reaches the SPECs it implements)
  --fanin             import fan-in ranking (stand-in for an @MX:DEBT fan-in
                      query — edges.jsonl carries no tag-kind edges yet)
  --specs-no-code     SPEC ids (spec.md frontmatter universe) with zero
                      mx-spec edges in the artifact

Examples:
  moai graph query --callers SPEC-FOO-001
  moai graph query --blast internal/config
  moai graph query --fanin --limit 20
  moai graph query --specs-no-code`,
		RunE: func(cmd *cobra.Command, args []string) error {
			selectors := 0
			for _, on := range []bool{callersNode != "", blastNode != "", fanin, specsNoCode} {
				if on {
					selectors++
				}
			}
			if selectors != 1 {
				return fmt.Errorf("exactly one of --callers, --blast, --fanin, --specs-no-code is required")
			}

			projectRoot, err := resolveGraphRoot(rootArg)
			if err != nil {
				return err
			}

			edgesFile := edgesPath
			if edgesFile == "" {
				edgesFile = filepath.Join(projectRoot, ".moai", "project", "graph", "edges.jsonl")
			} else if abs, absErr := filepath.Abs(edgesFile); absErr == nil {
				edgesFile = abs
			}
			edges, err := graph.LoadJSONL(edgesFile)
			// errors.Is (not os.IsNotExist): LoadJSONL wraps the open error
			// with %w, and os.IsNotExist does not follow wrap chains.
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("edges artifact not found at %s — run 'moai graph build' first", edgesFile)
			}
			if err != nil {
				return fmt.Errorf("load edges: %w", err)
			}

			out := cmd.OutOrStdout()
			switch {
			case callersNode != "":
				res := graph.FindCallers(edges, callersNode)
				_, _ = fmt.Fprintf(out, "callers of %s: %d\n", callersNode, len(res))
				for _, n := range res {
					_, _ = fmt.Fprintln(out, n)
				}
			case blastNode != "":
				res := graph.BlastRadius(edges, blastNode)
				_, _ = fmt.Fprintf(out, "blast radius of %s: %d\n", blastNode, len(res))
				for _, n := range res {
					_, _ = fmt.Fprintln(out, n)
				}
			case fanin:
				rows := graph.ImportFanIn(edges, limit)
				_, _ = fmt.Fprintf(out, "import fan-in: top %d\n", len(rows))
				for _, r := range rows {
					_, _ = fmt.Fprintf(out, "%d\t%s\n", r.FanIn, r.Package)
				}
			default: // --specs-no-code
				deps, err := mx.LoadSpecDependencies(projectRoot)
				if err != nil {
					return fmt.Errorf("load spec universe: %w", err)
				}
				universe := make([]string, 0, len(deps))
				for id := range deps {
					universe = append(universe, id)
				}
				sort.Strings(universe)

				res := graph.UnreferencedSpecs(edges, universe)
				_, _ = fmt.Fprintf(out, "SPECs with no code reference: %d of %d\n", len(res), len(universe))
				for _, id := range res {
					_, _ = fmt.Fprintln(out, id)
				}
				_, _ = fmt.Fprintln(out, unreferencedSpecCaveat)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&callersNode, "callers", "", "direct reverse neighbors of this node (package path, file path, or SPEC id)")
	cmd.Flags().StringVar(&blastNode, "blast", "", "transitive blast radius of a change at this node")
	cmd.Flags().BoolVar(&fanin, "fanin", false, "import fan-in ranking (default top 10, --limit 0 for all)")
	cmd.Flags().BoolVar(&specsNoCode, "specs-no-code", false, "SPEC ids with zero mx-spec edges in the artifact")
	cmd.Flags().IntVar(&limit, "limit", 10, "rows to show for --fanin (0 = all)")
	cmd.Flags().StringVar(&edgesPath, "edges", "", "edges.jsonl path (defaults to <root>/.moai/project/graph/edges.jsonl)")
	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")

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
