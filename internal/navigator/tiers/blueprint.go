package tiers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"log/slog"
)

// moduleTreeRelPath is the Tier 1 authored module-tree path (REQ-NS3-004).
const moduleTreeRelPath = ".moai/project/blueprint/module_tree.json"

// codemapsDepRelPath is the /moai codemaps dep-graph output (read-only seed).
const codemapsDepRelPath = ".moai/project/codemaps/dependencies.md"

// overviewProvenanceCommit is the placeholder last_updated_commit value
// written into overview.md when the engine is called WITHOUT an explicit
// commit. The CLI layer (M4.6) supplies the real git SHA.
const overviewProvenanceCommit = "<pending>"

// rawModuleTree is the JSON shape of the authored module_tree.json.
type rawModuleTree struct {
	Modules []rawModule `json:"modules"`
}

// rawModule is one authored module entry.
type rawModule struct {
	PackagePath    string   `json:"package_path"`
	DisplayName    string   `json:"display_name"`
	Layer          string   `json:"layer"`
	Responsibility string   `json:"responsibility"`
	DependsOn      []string `json:"depends_on"`
	OverviewPath   string   `json:"overview_path"`
}

// ensureModuleTreeScaffold implements the scaffold-then-refine loop
// (REQ-NS3-004 / design.md §1.D3). When module_tree.json is absent OR
// rescaffold=true, a draft is scaffolded from /moai codemaps dependencies.md.
// When module_tree.json exists and rescaffold=false, the file is left
// byte-identical (authored, not auto-replaced).
//
// @MX:ANCHOR: [AUTO] scaffold-then-refine gate; load-bearing for the blueprint-first stance (REQ-NS3-006)
// @MX:REASON: a plain run MUST NOT overwrite a human-edited module_tree.json — this gate is the operational seam between authored and generated
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func ensureModuleTreeScaffold(projectRoot string, rescaffold bool) error {
	path := filepath.Join(projectRoot, moduleTreeRelPath)
	if !rescaffold {
		if _, err := os.Stat(path); err == nil {
			// File exists and plain run → leave authored content alone.
			return nil
		}
	}
	// Either absent or rescaffold=true → write the deterministic draft.
	draft := scaffoldModuleTree(projectRoot)
	body, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return fmt.Errorf("tiers: marshal module_tree scaffold: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("tiers: mkdir module_tree dir: %w", err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o644); err != nil {
		return fmt.Errorf("tiers: write module_tree scaffold: %w", err)
	}
	return nil
}

// scaffoldModuleTree builds the deterministic draft by parsing the codemaps
// dependencies.md. Absent/unparseable dep-graph → empty modules list (fail-open).
func scaffoldModuleTree(projectRoot string) rawModuleTree {
	depPath := filepath.Join(projectRoot, codemapsDepRelPath)
	content, err := os.ReadFile(depPath)
	if err != nil {
		slog.Debug("tiers: codemaps dep-graph absent, scaffold degrades to empty", "path", depPath)
		return rawModuleTree{Modules: []rawModule{}}
	}
	mods := parseDependenciesMarkdown(string(content))
	if mods == nil {
		mods = []rawModule{}
	}
	return rawModuleTree{Modules: mods}
}

// parseDependenciesMarkdown extracts module-dependency pairs from the
// /moai codemaps dependencies.md body. It recognizes lines of the shape:
//
//	"<pkg> depends on <pkg>" / "<pkg> depends on <pkg>, <pkg>"
//	"<pkg> → <pkg>" (mermaid edge)
//
// Best-effort — unrecognized lines are skipped. The output is sorted by
// PackagePath for byte-stable scaffolds.
func parseDependenciesMarkdown(content string) []rawModule {
	byPath := map[string]*rawModule{}
	ensure := func(pkg string) *rawModule {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			return nil
		}
		// Skip mermaid subgraph labels / noise tokens.
		if isNoiseToken(pkg) {
			return nil
		}
		if m, ok := byPath[pkg]; ok {
			return m
		}
		m := &rawModule{PackagePath: pkg, DependsOn: []string{}}
		byPath[pkg] = m
		return m
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip markdown list/edge decorators.
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		// Strip mermaid edge decorators.
		if i := strings.Index(line, "["); i >= 0 {
			line = line[:i] + " " + strings.TrimSpace(line[strings.Index(line, "]")+1:])
		}
		// Recognize "<pkg> depends on <rest>".
		if idx := strings.Index(line, " depends on "); idx > 0 {
			src := ensure(strings.TrimSpace(line[:idx]))
			if src == nil {
				continue
			}
			rhs := strings.TrimSpace(line[idx+len(" depends on "):])
			for _, dep := range splitTopLevel(rhs) {
				if d := ensure(dep); d != nil {
					src.DependsOn = append(src.DependsOn, d.PackagePath)
				}
			}
			continue
		}
		// Recognize "<pkg> --> <pkg>" / "<pkg> -> <pkg>" / "<pkg> → <pkg>" /
		// "<pkg> -.-> <pkg>" (mermaid dotted arrow). "-.->" MUST be probed
		// before "->" — it contains "->" as a suffix, so a later probe would
		// split mid-arrow and leave a "pkg -." source token.
		for _, sep := range []string{"-.->", "-->", "->", "→"} {
			if i := strings.Index(line, sep); i > 0 {
				src := ensure(strings.TrimSpace(line[:i]))
				dst := ensure(strings.TrimSpace(line[i+len(sep):]))
				if src != nil && dst != nil {
					src.DependsOn = append(src.DependsOn, dst.PackagePath)
				}
				break
			}
		}
	}

	// Dedupe + sort depends_on per module, then sort modules.
	out := make([]rawModule, 0, len(byPath))
	for _, m := range byPath {
		m.DependsOn = dedupeSorted(m.DependsOn)
		m.Layer = string(inferLayer(m.PackagePath))
		if m.DisplayName == "" {
			m.DisplayName = m.PackagePath
		}
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PackagePath < out[j].PackagePath })
	return out
}

// stripMid was a parser helper retained during scaffolding; the parser now
// inlines its index lookup. The symbol is intentionally absent (deleted).

// ModuleDependency is one module→module adjacency entry (who imports whom)
// extracted from a /moai codemaps dependencies.md artifact.
//
// @MX:NOTE: [AUTO] ModuleDependency — exported adjacency seam so the graph writer (internal/graph) reuses the scaffold parser without a second markdown extractor
type ModuleDependency struct {
	PackagePath string
	DependsOn   []string
}

// ParseDependenciesMarkdown extracts the module dependency adjacency from
// /moai codemaps dependencies.md content. It is the exported seam over
// parseDependenciesMarkdown (same parsing, same fail-open behavior:
// unrecognized lines are skipped) so cross-package consumers — the graph
// writer (internal/graph) — reuse one extractor instead of forking it.
// Output ordering is deterministic (sorted by PackagePath, DependsOn sorted).
func ParseDependenciesMarkdown(content string) []ModuleDependency {
	mods := parseDependenciesMarkdown(content)
	out := make([]ModuleDependency, 0, len(mods))
	for _, m := range mods {
		out = append(out, ModuleDependency{
			PackagePath: m.PackagePath,
			DependsOn:   m.DependsOn,
		})
	}
	return out
}

// splitTopLevel splits a comma-separated dependency list, tolerant of "and".
func splitTopLevel(rhs string) []string {
	rhs = strings.ReplaceAll(rhs, " and ", ",")
	parts := strings.Split(rhs, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// Strip trailing prose like "for X" / "(comment)".
		if i := strings.IndexAny(p, " \t"); i > 0 {
			// Keep only the first whitespace-delimited token if it looks like a path.
			first := p[:i]
			if looksLikePath(first) {
				p = first
			}
		}
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// looksLikePath reports whether s looks like a package path (contains a slash
// or dot, no internal whitespace).
func looksLikePath(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t") {
		return false
	}
	return strings.Contains(s, "/") || strings.Contains(s, ".")
}

// isNoiseToken filters mermaid subgraph labels and prose tokens that are not
// package paths.
func isNoiseToken(s string) bool {
	if looksLikePath(s) {
		return false
	}
	// Single-word noise tokens that the markdown may surface.
	switch s {
	case "subgraph", "end", "graph", "TD", "LR", "P", "B":
		return true
	}
	// Strip surrounding quotes.
	s = strings.Trim(s, "\"")
	if s == "" {
		return true
	}
	// All-non-ASCII (Korean) prose → noise.
	for _, r := range s {
		if r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

// dedupeSorted returns a sorted, deduped copy of in.
func dedupeSorted(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

// inferLayer assigns a Layer (REQ-NS3-004) heuristically from the package
// path. The user/agent refines this in the authored tree; the scaffold's
// inference is just a starting point.
func inferLayer(pkg string) Layer {
	switch {
	case strings.Contains(pkg, "internal/cli"), strings.Contains(pkg, "internal/web"),
		strings.Contains(pkg, "internal/tui"), strings.Contains(pkg, "internal/statusline"),
		strings.Contains(pkg, "cmd/"), strings.Contains(pkg, "pkg/version"):
		return LayerPresentation
	case strings.Contains(pkg, "internal/spec"), strings.Contains(pkg, "internal/workflow"),
		strings.Contains(pkg, "internal/loop"), strings.Contains(pkg, "internal/harness"),
		strings.Contains(pkg, "internal/foundation"):
		return LayerDomain
	case strings.Contains(pkg, "internal/hook"), strings.Contains(pkg, "internal/config"),
		strings.Contains(pkg, "internal/navigator"), strings.Contains(pkg, "internal/mx"),
		strings.Contains(pkg, "internal/harness"):
		return LayerInfrastructure
	case strings.Contains(pkg, "internal/quality"), strings.Contains(pkg, "internal/audit"),
		strings.Contains(pkg, "internal/doctor"):
		return LayerMeasurement
	}
	return LayerDomain
}

// enumerateBlueprints loads the authored module_tree.json and emits
// BlueprintNode records + module-edges (REQ-NS3-007). Each BlueprintNode's
// OverviewPath defaults to `.moai/project/blueprint/<module>/overview.md`
// when the authored entry omits it. Output is sorted by Identifier for
// byte-stable emission (REQ-NS3-019).
func enumerateBlueprints(projectRoot string) ([]BlueprintNode, []TierEdge, error) {
	path := filepath.Join(projectRoot, moduleTreeRelPath)
	content, err := os.ReadFile(path)
	if err != nil {
		// Absent tree → 0 nodes (fail-open).
		return nil, nil, nil
	}
	var raw rawModuleTree
	if err := json.Unmarshal(content, &raw); err != nil {
		slog.Debug("tiers: module_tree.json unparseable, emitting 0 blueprint nodes", "error", err)
		return nil, nil, nil
	}

	nodes := make([]BlueprintNode, 0, len(raw.Modules))
	edges := []TierEdge{}
	for _, m := range raw.Modules {
		if m.PackagePath == "" {
			continue
		}
		overviewPath := m.OverviewPath
		if overviewPath == "" {
			overviewPath = filepath.Join(".moai/project/blueprint", m.PackagePath, "overview.md")
		}
		nodes = append(nodes, BlueprintNode{
			Identifier:     m.PackagePath,
			DisplayName:    m.DisplayName,
			Layer:          Layer(m.Layer),
			Responsibility: m.Responsibility,
			DependsOn:      m.DependsOn,
			OverviewPath:   overviewPath,
		})
		for _, dep := range m.DependsOn {
			edges = append(edges, TierEdge{
				EdgeType:   EdgeModule,
				SourceNode: "blueprint:" + m.PackagePath,
				TargetNode: "blueprint:" + dep,
			})
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Identifier < nodes[j].Identifier })
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].SourceNode != edges[j].SourceNode {
			return edges[i].SourceNode < edges[j].SourceNode
		}
		return edges[i].TargetNode < edges[j].TargetNode
	})
	return nodes, edges, nil
}

// instantiateOverview writes the Kiro 7-section overview.md template for one
// module IF it does not already exist (authored not overwritten). The
// provenance block carries last_updated_commit = commitSHA (a git SHA, never
// wall-clock per REQ-NS3-019).
func instantiateOverview(projectRoot string, node BlueprintNode, commitSHA string) error {
	if commitSHA == "" {
		commitSHA = overviewProvenanceCommit
	}
	overviewPath := node.OverviewPath
	if overviewPath == "" {
		overviewPath = filepath.Join(".moai/project/blueprint", node.Identifier, "overview.md")
	}
	abs := overviewPath
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(projectRoot, overviewPath)
	}
	if _, err := os.Stat(abs); err == nil {
		// Authored overview exists → do not overwrite.
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("tiers: mkdir overview dir: %w", err)
	}
	body := buildOverviewTemplate(node, commitSHA)
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		return fmt.Errorf("tiers: write overview.md: %w", err)
	}
	return nil
}

// buildOverviewTemplate returns the Kiro 7-section overview.md body with a
// provenance block.
func buildOverviewTemplate(node BlueprintNode, commitSHA string) string {
	return fmt.Sprintf(`# %s — Overview

> Module: %s
> Layer: %s
> Responsibility: %s

## Component Architecture

<Describe the module's internal components and their boundaries.>

## Data Flow

<Trace how data moves through this module.>

## Data Model

<List the primary types and their relationships.>

## Error Handling

<How errors are surfaced, wrapped, and recovered.>

## Test Strategy

<What is tested at unit / integration / e2e level.>

## Implementation Approach

<Key design decisions and patterns.>

## Migration

<Migration notes for callers when this module's contract changes.>

---

<!-- provenance
last_updated_commit: %s
-->
`, node.DisplayName, node.Identifier, node.Layer, node.Responsibility, commitSHA)
}
