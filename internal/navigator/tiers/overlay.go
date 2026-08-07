package tiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"
)

// tiersOutRelPath is the overlay output path (REQ-NS3-018 write-surface #5).
const tiersOutRelPath = ".moai/project/navigator/tiers.json"

// navGraphRelPath is the M0 producer path (READ-ONLY per REQ-NS3-016).
const navGraphRelPath = ".moai/project/navigator/nav-graph.json"

// navigatorLogRelPath is the fail-open diagnostic log destination (REQ-NS3-020).
const navigatorLogRelPath = ".moai/logs/navigator-sync.log"

// Enrich emits the tiers.json overlay into projectRoot's
// .moai/project/navigator/ directory.
//
// REQ-NS3-016 (consumer-only): the implementation MUST NOT modify M0/M1
// producer paths. REQ-NS3-018 (write-surface isolation): the implementation
// writes ONLY to the 6 named surfaces and NEVER overwrites nav-graph.json.
// REQ-NS3-020 (fail-open): every error mode logs to navigator-sync.log and
// yields a successful return — the calling /moai project step never aborts.
//
// @MX:ANCHOR: [AUTO] 4-tier overlay emission entry point; high fan_in (called by the M4.6 CLI + tests + future /moai project wiring)
// @MX:REASON: the single seam between the 4 tier engines and the on-disk overlay — provenance, atomic-write, fail-open, and the overlay-not-overwrite invariant all hinge on this call site
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func Enrich(projectRoot string) error {
	// (1) Provenance — git baseline, NO wall-clock (REQ-NS3-019).
	prov := currentProvenance(projectRoot)

	// (2) Tier 0 — contract nodes (REQ-NS3-001/002/003). Drift checks run
	// but never block emission (graph fail-open).
	contracts, _ := enumerateContracts(projectRoot)
	runDriftChecks(projectRoot, contracts)

	// (3) Tier 1 — blueprint nodes + module-edges (REQ-NS3-004/005/006/007).
	// Scaffold-then-refine: a plain run does NOT overwrite an authored tree.
	_ = ensureModuleTreeScaffold(projectRoot, false)
	blueprints, moduleEdges, _ := enumerateBlueprints(projectRoot)
	commitSHA := prov.ExtractCommitSHA
	for _, b := range blueprints {
		_ = instantiateOverview(projectRoot, b, commitSHA)
	}

	// (4) Tier 2 — decision enrichment (REQ-NS3-008/009/010). Decisions are
	// sourced from M0 nav-graph.json (READ-ONLY consumer); absent/unparseable
	// graph → empty identifiers (fail-open).
	decisionIDs := readM0DecisionIDs(projectRoot)
	decisions := enumerateDecisions(projectRoot, decisionIDs)
	decisionEdges := buildSupersededByEdges(decisions)

	// (5) Tier 3 — symbol enrichment (REQ-NS3-011..015). Deterministic-only
	// by default (NarrativeEnabled=false → 2-tier separable).
	symbols, _, _ := enumerateSymbols(projectRoot, SymbolOptions{NarrativeEnabled: false})
	ownsEdges := buildOwnsEdges(blueprints, symbols)

	// (6) Assemble the overlay.
	overlay := TiersOverlay{
		Provenance:      prov,
		Tier0Contracts:  contracts,
		Tier1Blueprints: blueprints,
		Tier2Decisions:  decisions,
		Tier3Symbols:    symbols,
		TierEdges:       mergeTierEdges(moduleEdges, ownsEdges, decisionEdges),
	}
	sortOverlay(&overlay)

	// (7) Emit — atomic write (tmp + rename).
	if err := writeOverlay(projectRoot, overlay); err != nil {
		logNavigator(projectRoot, fmt.Sprintf("tiers: overlay write failed (fail-open): %v", err))
		return nil
	}
	return nil
}

// currentProvenance returns the git baseline provenance (no wall-clock).
// Fail-open: returns "<unknown>" on any git error.
func currentProvenance(projectRoot string) Provenance {
	sha := gitTrimmed(projectRoot, "rev-parse", "HEAD")
	if sha == "" {
		sha = "<unknown>"
	}
	captured := gitTrimmed(projectRoot, "log", "-1", "--format=%cI", sha)
	if captured == "" {
		captured = "<unknown>"
	}
	return Provenance{ExtractCommitSHA: sha, CapturedAt: captured}
}

// gitTrimmed runs a git command in projectRoot and returns trimmed stdout.
func gitTrimmed(projectRoot string, args ...string) string {
	cmd := exec.Command("git", args...)
	if projectRoot != "" {
		cmd.Dir = projectRoot
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Debug("tiers: git command failed", "args", args, "dir", projectRoot, "error", err)
		return ""
	}
	return strings.TrimSpace(out.String())
}

// m0NavGraph is the minimal M0 nav-graph.json shape this layer consumes.
// Only the `nodes` array is read; the join key is (entity_type, identifier).
type m0NavGraph struct {
	Nodes []m0Node `json:"nodes"`
}

// m0Node is one M0 graph node.
type m0Node struct {
	EntityType string `json:"entity_type"`
	Identifier string `json:"identifier"`
}

// readM0DecisionIDs extracts the decision identifiers from nav-graph.json.
// Absent/unparseable graph → empty list (fail-open per REQ-NS3-020).
func readM0DecisionIDs(projectRoot string) []string {
	path := filepath.Join(projectRoot, navGraphRelPath)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var g m0NavGraph
	if err := json.Unmarshal(content, &g); err != nil {
		logNavigator(projectRoot, fmt.Sprintf("tiers: nav-graph.json unparseable, proceeding without decision enrichment: %v", err))
		return nil
	}
	out := []string{}
	for _, n := range g.Nodes {
		if n.EntityType == "decision" {
			out = append(out, n.Identifier)
		}
	}
	return out
}

// buildSupersededByEdges emits one superseded_by edge per decision whose
// SupersededBy is populated (REQ-NS3-009).
func buildSupersededByEdges(decisions []DecisionEnrichment) []TierEdge {
	out := []TierEdge{}
	for _, d := range decisions {
		if d.SupersededBy == "" {
			continue
		}
		out = append(out, TierEdge{
			EdgeType:   EdgeSupersededBy,
			SourceNode: "decision:" + d.Identifier,
			TargetNode: "decision:" + d.SupersededBy,
		})
	}
	return out
}

// buildOwnsEdges emits owns-edge entries joining each blueprint to the
// symbols extracted within it (REQ-NS3-007). The attribution is by
// declaration-path containment: a symbol whose DeclarationPath is rooted
// under the blueprint's package path is "owned" by that blueprint.
func buildOwnsEdges(blueprints []BlueprintNode, symbols []SymbolEnrichment) []TierEdge {
	out := []TierEdge{}
	for _, b := range blueprints {
		bp := b.Identifier + "/"
		for _, s := range symbols {
			if s.DeclarationPath == "" {
				continue
			}
			if s.DeclarationPath == b.Identifier || strings.HasPrefix(s.DeclarationPath, bp) {
				out = append(out, TierEdge{
					EdgeType:   EdgeOwns,
					SourceNode: "blueprint:" + b.Identifier,
					TargetNode: "symbol:" + s.Identifier,
				})
			}
		}
	}
	return out
}

// mergeTierEdges concatenates edge slices and sorts the result.
func mergeTierEdges(parts ...[]TierEdge) []TierEdge {
	var out []TierEdge
	for _, p := range parts {
		out = append(out, p...)
	}
	sortEdges(out)
	return out
}

// sortEdges sorts edges by (EdgeType, SourceNode, TargetNode).
func sortEdges(edges []TierEdge) {
	for i := 1; i < len(edges); i++ {
		for j := i; j > 0; j-- {
			a, b := edges[j-1], edges[j]
			if a.EdgeType > b.EdgeType {
				edges[j-1], edges[j] = edges[j], edges[j-1]
				continue
			}
			if a.EdgeType < b.EdgeType {
				break
			}
			if a.SourceNode > b.SourceNode {
				edges[j-1], edges[j] = edges[j], edges[j-1]
				continue
			}
			if a.SourceNode < b.SourceNode {
				break
			}
			if a.TargetNode > b.TargetNode {
				edges[j-1], edges[j] = edges[j], edges[j-1]
				continue
			}
			break
		}
	}
}

// sortOverlay ensures deterministic ordering of every overlay slice for
// byte-identical re-runs (REQ-NS3-019).
func sortOverlay(o *TiersOverlay) {
	sortContractNodes(o.Tier0Contracts)
	sort.SliceStable(o.Tier1Blueprints, func(i, j int) bool {
		return o.Tier1Blueprints[i].Identifier < o.Tier1Blueprints[j].Identifier
	})
	sort.SliceStable(o.Tier2Decisions, func(i, j int) bool {
		return o.Tier2Decisions[i].Identifier < o.Tier2Decisions[j].Identifier
	})
	sort.SliceStable(o.Tier3Symbols, func(i, j int) bool {
		return o.Tier3Symbols[i].Identifier < o.Tier3Symbols[j].Identifier
	})
	sortEdges(o.TierEdges)
}

// writeOverlay serializes the overlay to JSON and writes it atomically.
func writeOverlay(projectRoot string, o TiersOverlay) error {
	body, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	outPath := filepath.Join(projectRoot, tiersOutRelPath)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, append(body, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, outPath)
}

// logNavigator appends a single diagnostic line to .moai/logs/navigator-sync.log
// (REQ-NS3-020 fail-open). Errors during logging are themselves swallowed.
func logNavigator(projectRoot, msg string) {
	path := filepath.Join(projectRoot, navigatorLogRelPath)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	// No wall-clock in the message itself — the deterministic layer never
	// stamps wall time. The log is an operational trace, not a tier artifact.
	_, _ = f.WriteString(msg + "\n")
}
