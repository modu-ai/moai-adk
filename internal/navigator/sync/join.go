package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"

	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// Options configures a Run.
type Options struct {
	ProjectRoot string
	// CapabilityMapPath defaults to <root>/.moai/project/navigator/capability-map.md.
	CapabilityMapPath string
	// CapabilitySymbolsPath defaults to <root>/.moai/project/codemaps/capability-symbols.json.
	CapabilitySymbolsPath string
	// AuditReportPath defaults to <root>/.moai/project/navigator/audit-report.json.
	AuditReportPath string
	// OutPath defaults to <root>/.moai/project/navigator/nav-graph.json.
	OutPath string
	// LogPath defaults to <root>/.moai/logs/navigator-sync.log.
	LogPath string
}

// resolve fills empty path fields with the standard defaults.
func (o *Options) resolve() {
	root := o.ProjectRoot
	if o.CapabilityMapPath == "" {
		o.CapabilityMapPath = filepath.Join(root, ".moai", "project", "navigator", "capability-map.md")
	}
	if o.CapabilitySymbolsPath == "" {
		o.CapabilitySymbolsPath = filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.json")
	}
	if o.AuditReportPath == "" {
		o.AuditReportPath = filepath.Join(root, ".moai", "project", "navigator", "audit-report.json")
	}
	if o.OutPath == "" {
		o.OutPath = filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")
	}
	if o.LogPath == "" {
		o.LogPath = filepath.Join(root, logPathName)
	}
}

// CapabilityGateError is returned when the capability gate (REQ-NS-011)
// trips: capability-map.md absent. Callers MUST treat this as fail-open
// (exit code 0, no output) — the Run function does this internally; this
// sentinel only surfaces for callers that want to distinguish the gate trip
// from a normal completion.
type CapabilityGateError struct {
	ProjectRoot string
}

func (e *CapabilityGateError) Error() string {
	return "navigator-sync: capability-map absent at " + e.ProjectRoot
}

// Run is the fail-open core. It joins the three chain outputs with the
// binding-token scanner records and writes nav-graph.json atomically.
// On the capability gate (capability-map.md absent), Run writes an info log,
// emits NO output, and returns nil (REQ-NS-011). All other errors are logged
// and swallowed; Run never aborts the caller (carries forward 003's fail-open
// contract).
func Run(opts Options) error {
	opts.resolve()

	// Capability gate (REQ-NS-011).
	if _, err := os.Stat(opts.CapabilityMapPath); err != nil {
		appendLog(opts.LogPath,
			fmt.Sprintf("navigator-sync: capability-map absent at %s; skipping join", opts.CapabilityMapPath))
		slog.Debug("sync: capability gate tripped", "path", opts.CapabilityMapPath)
		return nil
	}

	prov := CurrentProvenance(opts.ProjectRoot)
	logPath := opts.LogPath

	// 1. capability-map.md → spec nodes (header-driven parse, reusing 003).
	specFromMap := loadSpecsFromCapabilityMap(opts.CapabilityMapPath, logPath)

	// 2. capability-symbols.json → symbol nodes (003 output, advisory).
	symsFrom003 := loadCapabilitySymbols(opts.CapabilitySymbolsPath, logPath)

	// 3. audit-report.json → advisory only (002 output); no required edges.
	_ = loadAuditReport(opts.AuditReportPath, logPath)

	// 4. @NAV:DEC records → decision nodes + dec-edges.
	decRecs, _, err := ScanDec(opts.ProjectRoot, prov.ExtractCommitSHA, logPath)
	if err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: scan-dec error: %v", err))
	}
	// 5. @NAV:SYM records → sym-edges (+ new symbol nodes per D2 rule).
	symRecs, _, err := ScanSym(opts.ProjectRoot, prov.ExtractCommitSHA, logPath)
	if err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: scan-sym error: %v", err))
	}
	// 6. @MX:SPEC associations → spec-edges (consumed, not re-scanned; REQ-NS-005).
	mxBridge, err := BridgeMxAssociations(opts.ProjectRoot)
	if err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: mx-bridge error: %v", err))
	}

	g := buildGraph(prov, specFromMap, symsFrom003, decRecs, symRecs, mxBridge)

	if err := os.MkdirAll(filepath.Dir(opts.OutPath), 0o755); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: mkdir error: %v", err))
		return nil
	}
	if err := WriteGraph(opts.OutPath, g); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: write error: %v", err))
		return nil
	}
	return nil
}

// loadSpecsFromCapabilityMap reads 001's capability-map.md and returns the
// sorted-unique set of SPEC IDs from its spec-id column. It delegates to the
// lightweight `astx.SpecIDsFromCapabilityMap` helper, which reuses the shared
// `parseCapabilityMap` parser WITHOUT walking implementation paths or running
// tree-sitter extraction (REQ-NS-012 consumer-only). The previous
// implementation called `astx.EnrichRows`, which runs the full enrichment
// pipeline (per-row `filepath.WalkDir` + tree-sitter `Extract()` on every
// file) just to read spec-ids already present in the table.
func loadSpecsFromCapabilityMap(capMapPath, logPath string) []string {
	ids, err := astx.SpecIDsFromCapabilityMap(capMapPath)
	if err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: capability-map parse error: %v", err))
		return nil
	}
	return ids
}

// capabilitySymbolsDoc is the minimal shape of 003's capability-symbols.json
// needed by the join (PrimarySymbols[].Name only — REQ-NS-007).
type capabilitySymbolsDoc struct {
	Rows []struct {
		SpecID         string `json:"spec_id"`
		PrimarySymbols []struct {
			Name string `json:"name"`
		} `json:"primary_symbols"`
	} `json:"rows"`
}

// loadCapabilitySymbols reads 003's capability-symbols.json (advisory). A
// missing or malformed file returns nil — the symbol node set degrades to
// @NAV:SYM-only (REQ-NS-016 graceful degradation).
func loadCapabilitySymbols(path, logPath string) capabilitySymbolsDoc {
	b, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("sync: capability-symbols absent", "path", path, "error", err)
		return capabilitySymbolsDoc{}
	}
	var doc capabilitySymbolsDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: capability-symbols parse error: %v", err))
		return capabilitySymbolsDoc{}
	}
	return doc
}

// loadAuditReport reads 002's audit-report.json (advisory). A missing or
// malformed file returns nil — the join proceeds without audit-derived edges.
func loadAuditReport(path, logPath string) json.RawMessage {
	b, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("sync: audit-report absent", "path", path, "error", err)
		return nil
	}
	// Validate JSON without enforcing a shape (advisory only).
	var raw json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		appendLog(logPath, fmt.Sprintf("navigator-sync: audit-report parse error: %v", err))
		return nil
	}
	return raw
}

// buildGraph assembles the Graph: nodes and edges from the three chain inputs
// plus the three binding-record sets. All keys / elements are sorted for
// byte-identical re-runs (REQ-NS-009).
func buildGraph(
	prov Provenance,
	specsFromMap []string,
	syms capabilitySymbolsDoc,
	decRecs []BindingRecord,
	symRecs []BindingRecord,
	mxBridge []MxBridgeSpec,
) Graph {
	nodes := newNodeSet()
	edges := []Edge{}

	// (a) Spec nodes from capability-map.md.
	for _, specID := range specsFromMap {
		nodes.add(EntitySpec, specID, specID)
	}

	// (b) Symbol nodes from 003's capability-symbols.json (D2 source-of-truth).
	for _, row := range syms.Rows {
		for _, s := range row.PrimarySymbols {
			if s.Name == "" {
				continue
			}
			nodes.add(EntitySymbol, s.Name, lastSegment(s.Name))
		}
	}

	// (c) @NAV:DEC records → decision nodes + dec-edges (Decision↔Spec when a
	// spec-id-named decision matches a known spec; Decision node either way).
	for _, r := range decRecs {
		nodes.add(EntityDecision, r.Identifier, r.Identifier)
		targetKind, targetID := decideTarget(nodes, r.Identifier, specsFromMap)
		if targetKind == "" {
			continue
		}
		edges = append(edges, Edge{
			EdgeType:   EdgeDec,
			SourceNode: nodeKey(EntityDecision, r.Identifier),
			TargetNode: nodeKey(targetKind, targetID),
			SourcePath: r.SourcePath,
			LineNumber: r.LineNumber,
		})
	}

	// (d) @NAV:SYM records → sym-edges + new symbol nodes per D2 rule.
	for _, r := range symRecs {
		canonicalID := resolveSymID(nodes, r.Identifier)
		nodes.add(EntitySymbol, canonicalID, lastSegment(canonicalID))
		// Self-edge (Symbol↔Symbol) per REQ-NS-008 (Symbol↔Symbol or Symbol↔Doc).
		// M0 binds the binding occurrence as a self-loop; future milestones may
		// re-target to a Doc node.
		edges = append(edges, Edge{
			EdgeType:   EdgeSym,
			SourceNode: nodeKey(EntitySymbol, canonicalID),
			TargetNode: nodeKey(EntitySymbol, canonicalID),
			SourcePath: r.SourcePath,
			LineNumber: r.LineNumber,
		})
	}

	// (e) @MX:SPEC associations → spec-edges (Code↔Spec).
	for _, m := range mxBridge {
		// Materialize a per-occurrence symbol node keyed by source-path + line.
		symID := "mx:" + m.SourcePath + ":" + itoa(m.LineNumber)
		nodes.add(EntitySymbol, symID, filepath.Base(m.SourcePath))
		// Ensure the spec node exists even when capability-map missed it.
		nodes.add(EntitySpec, m.SpecID, m.SpecID)
		edges = append(edges, Edge{
			EdgeType:   EdgeSpec,
			SourceNode: nodeKey(EntitySymbol, symID),
			TargetNode: nodeKey(EntitySpec, m.SpecID),
			SourcePath: m.SourcePath,
			LineNumber: m.LineNumber,
		})
	}

	g := Graph{
		Provenance: prov,
		Nodes:      nodes.sorted(),
		Edges:      edges,
	}
	sortEdges(g.Edges)
	return g
}

// nodeSet is a deterministic node collection keyed by (entity_type, identifier).
type nodeSet struct {
	m map[EntityType]map[string]string // entity → identifier → display_name
}

func newNodeSet() nodeSet {
	return nodeSet{m: map[EntityType]map[string]string{}}
}

func (n *nodeSet) add(t EntityType, id, display string) {
	if id == "" {
		return
	}
	inner, ok := n.m[t]
	if !ok {
		inner = map[string]string{}
		n.m[t] = inner
	}
	if _, exists := inner[id]; !exists {
		inner[id] = display
	}
}

func (n *nodeSet) has(t EntityType, id string) bool {
	if inner, ok := n.m[t]; ok {
		_, found := inner[id]
		return found
	}
	return false
}

func (n *nodeSet) sorted() []Node {
	var types []EntityType
	for t := range n.m {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })
	var out []Node
	for _, t := range types {
		inner := n.m[t]
		ids := make([]string, 0, len(inner))
		for id := range inner {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			out = append(out, Node{
				EntityType:  t,
				Identifier:  id,
				DisplayName: inner[id],
			})
		}
	}
	return out
}

// decideTarget picks the dec-edge target for a decision token. If the
// decision identifier matches a known SPEC ID (e.g. `@NAV:DEC-SPEC-NAV-001`),
// target the Spec node; otherwise target the Decision node itself (self-loop).
func decideTarget(nodes nodeSet, decID string, specs []string) (EntityType, string) {
	for _, specID := range specs {
		if decID == specID {
			return EntitySpec, specID
		}
	}
	return EntityDecision, decID
}

// resolveSymID implements the D2 resolution rule: exact match against an
// existing symbol node → reuse; suffix match (last `.`-segment) → reuse;
// else the token-authored form is the new node identifier.
func resolveSymID(nodes nodeSet, token string) string {
	if nodes.has(EntitySymbol, token) {
		return token
	}
	if dot := strings.LastIndex(token, "."); dot >= 0 {
		suffix := token[dot+1:]
		if suffix != "" && nodes.has(EntitySymbol, suffix) {
			return suffix
		}
	}
	return token
}

// lastSegment returns the substring after the final `.` or `/` (the bare
// identifier for display purposes).
func lastSegment(s string) string {
	if i := strings.LastIndexAny(s, "./"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// sortEdges orders edges deterministically.
func sortEdges(e []Edge) {
	sort.SliceStable(e, func(i, j int) bool {
		if e[i].EdgeType != e[j].EdgeType {
			return e[i].EdgeType < e[j].EdgeType
		}
		if e[i].SourceNode != e[j].SourceNode {
			return e[i].SourceNode < e[j].SourceNode
		}
		if e[i].TargetNode != e[j].TargetNode {
			return e[i].TargetNode < e[j].TargetNode
		}
		if e[i].SourcePath != e[j].SourcePath {
			return e[i].SourcePath < e[j].SourcePath
		}
		return e[i].LineNumber < e[j].LineNumber
	})
}
