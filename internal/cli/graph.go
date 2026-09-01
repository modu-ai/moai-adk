package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/kanban"
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
	cmd.AddCommand(newGraphCheckCmd())
	cmd.AddCommand(newGraphStampCmd())
	return cmd
}

// unreferencedSpecCaveat is printed with every --specs-no-code result: a
// SPEC absent from edges.jsonl is NOT an unimplemented SPEC — most SPECs
// deliver docs/rules/harness with no code by design.
const unreferencedSpecCaveat = "NOTE: unreferenced != unimplemented (미연결 ≠ 미구현) — most SPECs deliver docs/rules/harness, so\nhaving no code reference is normal; treat this list as a coverage map, not a defect list."

// milestoneNoCardCaveat is printed with every --milestones-no-card result:
// the queue drops rows on done, so "not in live queue" spans completed and
// never-issued alike — the flag forces the question, git history answers it.
// The third line warns that a zero-hit grep does not mean the work never
// happened: a card may have been re-issued under a new id, so the lineage is
// checked before a new card is requested.
const milestoneNoCardCaveat = "NOTE: 'not in live queue' covers completed AND never-issued cards — done removes queue rows.\nResolve each flag with: git log --oneline --grep 'merge: tNN' (완결이면 통과, 미발급이면 새 카드).\ngrep 0건 ≠ 작업 미완 — 카드가 재발행되었을 수 있으니 새 카드 발급 전에 계보를 확인."

// todoQueueRootFn is the seam the --milestones-no-card tests use to point
// the queue cross-check at a fixture backlog (same injection pattern as
// userHomeDirFn in todo.go).
var todoQueueRootFn = resolveTodoQueueRoot

// liveQueueCards returns the card ids currently live in the backlog queue
// (queued or picked — dropped and done cards are not live). ok=false means
// the queue file was unreadable, so the caller reports the card-vs-queue
// comparison as skipped instead of silently passing every claim.
func liveQueueCards() (live map[string]bool, ok bool) {
	rec, err := kanban.NewBacklogStore(todoBacklogPath(todoQueueRootFn())).Load()
	if err != nil {
		return nil, false
	}
	live = map[string]bool{}
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStateQueued || it.State == kanban.BacklogStatePicked {
			live[it.ID] = true
		}
	}
	return live, true
}

// newGraphQueryCmd 'moai graph query' answers reverse-dependency questions
// from the persisted edges.jsonl without re-running extraction.
//
// @MX:NOTE: [AUTO] moai graph query — read-only consumer of edges.jsonl; agents use it instead of re-running go list -deps per question
func newGraphQueryCmd() *cobra.Command {
	var callersNode, blastNode, edgesPath, rootArg string
	var fanin, debtFanin, specsNoCode, milestonesNoCard bool
	var limit int

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query edges.jsonl: callers / blast radius / fan-in / SPECs with no code / milestones with no card",
		Long: `Answer reverse-dependency questions from the persisted edges.jsonl (run 'moai graph build' first). Exactly one selector per invocation:

  --callers <node>    direct reverse neighbors: importers of a package, SPECs
                      depending on a SPEC, code files tagged @MX:SPEC
  --blast <node>      transitive blast radius of a change at <node> (BFS over
                      reverse edges; mx-spec edges propagate both ways, so a
                      code file reaches the SPECs it implements)
  --fanin             import fan-in ranking (package-level dependency
                      questions; for @MX:DEBT fan-in use --debt-fanin)
  --debt-fanin        rank @MX:DEBT tag targets by graph fan-in
                      (evidence-backed distinct caller files, descending,
                      ties by target; file-scope DEBT ranks 0 and is listed
                      with a (self) marker)
  --specs-no-code     SPEC ids (spec.md frontmatter universe) with zero
                      mx-spec edges in the artifact
  --milestones-no-card
                      report milestones whose Card Cross-Check row claims no
                      card, or whose claimed card is absent from the live
                      backlog queue (queued/picked)

Examples:
  moai graph query --callers SPEC-FOO-001
  moai graph query --blast internal/config
  moai graph query --fanin --limit 20
  moai graph query --debt-fanin --limit 20
  moai graph query --specs-no-code
  moai graph query --milestones-no-card`,
		RunE: func(cmd *cobra.Command, args []string) error {
			selectors := 0
			for _, on := range []bool{callersNode != "", blastNode != "", fanin, debtFanin, specsNoCode, milestonesNoCard} {
				if on {
					selectors++
				}
			}
			if selectors != 1 {
				return fmt.Errorf("exactly one of --callers, --blast, --fanin, --debt-fanin, --specs-no-code, --milestones-no-card is required")
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

			// REQ-GF-007: refresh the mechanical input layers before
			// answering. edges.jsonl is derived: it is stale exactly when a
			// source fingerprint moved OR its mx-index source layer itself
			// drifted (the index FILE hash cannot see scan-source edits until
			// the index is rewritten — the inventory drift probe can).
			// The curated codemaps layer is NEVER auto-rewritten here — its
			// staleness is the M1 gate's signal (spec.md §B.2). The decision
			// evaluates the SELECTED --edges artifact's provenance (CR
			// round-2 3855149254), with the gate-calibrated drift red line.
			if refreshNeeded := edgesRefreshNeeded(projectRoot, edgesFile, graph.DefaultThresholds().MXIndexChangedFiles); refreshNeeded {
				if stats, rErr := refreshEdgesArtifact(projectRoot, edgesFile); rErr != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph refresh failed (answering from the existing artifact): %v\n", rErr)
				} else if over := graphRefreshOverrun(projectRoot, stats.duration); over > 0 {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
						"graph refresh cost %s exceeded the %dms update budget by %.0fms (warning only, answer follows)\n",
						stats.duration.Round(time.Millisecond), graphRefreshBudgetMS(projectRoot), over.Seconds()*1000)
				}
				edges, err = graph.LoadJSONL(edgesFile)
				if err != nil {
					return fmt.Errorf("reload refreshed edges: %w", err)
				}
			}
			// REQ-GF-008: every query answer names the tree root and commit
			// (or dirty fingerprint) it was computed from.
			if pv, ok := graph.ReadEdgesMeta(filepath.Join(filepath.Dir(edgesFile), graph.MetaFileName)); ok {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), pv.Describe())
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
			case debtFanin:
				rows := graph.DebtFanIn(edges, limit)
				_, _ = fmt.Fprintf(out, "debt fan-in (evidence-backed caller files): top %d\n", len(rows))
				for _, r := range rows {
					file := r.File
					if r.Self {
						file = "(self)"
					}
					_, _ = fmt.Fprintf(out, "%d\t%s\t%s\n", r.FanIn, r.Target, file)
				}
			case specsNoCode:
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
			case milestonesNoCard:
				claims := graph.MilestoneClaims(edges)
				live, queueOK := liveQueueCards()

				flagged := 0
				for _, c := range claims {
					if len(c.Cards) == 0 {
						flagged++
						_, _ = fmt.Fprintf(out, "%s  no card claimed ([new card needed])\n", c.Milestone)
						continue
					}
					var missing []string
					for _, card := range c.Cards {
						if queueOK && !live[card] {
							missing = append(missing, card)
						}
					}
					if len(missing) > 0 {
						flagged++
						_, _ = fmt.Fprintf(out, "%s  claimed %s — not in live queue: %s (done or never issued)\n",
							c.Milestone, strings.Join(c.Cards, ","), strings.Join(missing, ","))
					}
				}
				_, _ = fmt.Fprintf(out, "milestones without a live card: %d of %d\n", flagged, len(claims))
				if len(claims) == 0 {
					_, _ = fmt.Fprintln(out, "NOTE: no Card Cross-Check sections found — reports without the section are invisible to this query")
				}
				if !queueOK {
					_, _ = fmt.Fprintln(out, "NOTE: backlog queue unreadable — card-vs-queue comparison skipped (only no-card milestones flagged)")
				}
				_, _ = fmt.Fprintln(out, milestoneNoCardCaveat)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&callersNode, "callers", "", "direct reverse neighbors of this node (package path, file path, or SPEC id)")
	cmd.Flags().StringVar(&blastNode, "blast", "", "transitive blast radius of a change at this node")
	cmd.Flags().BoolVar(&fanin, "fanin", false, "import fan-in ranking (default top 10, --limit 0 for all)")
	cmd.Flags().BoolVar(&debtFanin, "debt-fanin", false, "rank @MX:DEBT targets by evidence-backed graph fan-in (default top 10, --limit 0 for all)")
	cmd.Flags().BoolVar(&specsNoCode, "specs-no-code", false, "SPEC ids with zero mx-spec edges in the artifact")
	cmd.Flags().BoolVar(&milestonesNoCard, "milestones-no-card", false, "milestones claiming no card, or whose claimed card is absent from the live backlog queue")
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
	var allDisagreements bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Aggregate codemaps/@MX/SPEC/report edges into edges.jsonl",
		Long: `Aggregate the existing graph-producing layers into edges.jsonl (one JSON edge per line, sorted, git-diffable):

  import            package -> package     from .moai/project/codemaps/dependencies.md
  mx-spec           file -> SPEC           from @MX:SPEC sub-lines (mx scanner)
  spec-depends      SPEC -> SPEC           from spec.md frontmatter depends_on
  report-milestone  report -> milestone    from .moai/reports/*.md "## Card Cross-Check" sections
  milestone-card    milestone -> card      from the same section's card column (tNN ids)

The report layer reads the mandatory cross-check table a milestone-bearing
report carries: one row per milestone, a column headed "card" holding the
delivering card id (tNN) or an explicit new-card marker. Query the result
with 'moai graph query --milestones-no-card'.

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

			mode := graph.DisagreementRefuteOnly
			if allDisagreements {
				mode = graph.DisagreementAll
			}
			edges, matrix, err := graph.BuildWithCodeLayersMode(projectRoot, mode)
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

			// REQ-GF-003: stamp the source-set fingerprints next to the
			// artifact so its staleness is judgeable without a rebuild.
			metaPath := filepath.Join(filepath.Dir(target), graph.MetaFileName)
			if err := graph.WriteEdgesMeta(metaPath, projectRoot, graph.SourceFingerprintsForEdges(projectRoot), len(edges)); err != nil {
				return fmt.Errorf("write edges meta: %w", err)
			}

			out := cmd.OutOrStdout()
			counts := map[string]int{}
			for _, e := range edges {
				counts[e.Kind]++
			}
			_, _ = fmt.Fprintf(out, "OK: wrote %d edges to %s\n", len(edges), target)
			for _, kind := range []string{graph.KindImport, graph.KindMXSpec, graph.KindSpecDepends, graph.KindReportMilestone, graph.KindMilestoneCard, graph.KindCodeCall, graph.KindCodeImport} {
				if c := counts[kind]; c > 0 {
					_, _ = fmt.Fprintf(out, "  %s: %d\n", kind, c)
				}
			}
			// REQ-GF-016: grade-matrix defect verdict — reported, never silent.
			for _, defect := range graph.ValidateGradeMatrix(matrix) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", defect)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&outPath, "out", "", "output path (defaults to <root>/.moai/project/graph/edges.jsonl)")
	cmd.Flags().StringVar(&rootArg, "root", "", "project root (defaults to the auto-detected project root)")
	cmd.Flags().BoolVar(&allDisagreements, "all-disagreements", false,
		"also mark the suppressed direction: local code-import dependencies the doc layer does not record (revival path for the default's code-found/doc-silent suppression)")

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
