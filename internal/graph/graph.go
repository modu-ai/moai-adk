// Package graph persists the codebase edge list as a git-diffable JSONL
// artifact (edges.jsonl). It is a WRITER only: the three graph-producing
// layers it aggregates are existing extractors this package reuses.
//
//	kind "import"       package → package   /moai codemaps dependencies.md
//	                                          (via navigator/tiers.ParseDependenciesMarkdown)
//	kind "mx-spec"      file:line → SPEC    @MX:SPEC sub-lines
//	                                          (via mx.Scanner SpecRef capture)
//	kind "spec-depends" SPEC → SPEC         spec.md frontmatter depends_on
//	                                          (via mx.LoadSpecDependencies)
//	kind "report-milestone" report → milestone  .moai/reports/*.md Card
//	                                          Cross-Check sections (report.go)
//	kind "milestone-card"   milestone → card    same section, card column
//
// Output contract: one line = one edge, edges sorted by (kind, source,
// target, line), no timestamps — two runs on the same tree produce
// byte-identical output so the artifact diffs cleanly in git.
//
// @MX:NOTE: [AUTO] edges.jsonl writer — JSONL over a server/DB for operational cost: git-diffable, zero cgo, zero background services
package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/mx"
	"github.com/modu-ai/moai-adk/internal/navigator/astx"
	"github.com/modu-ai/moai-adk/internal/navigator/tiers"
)

// Edge kinds emitted by Build.
const (
	// KindImport is a package→package dependency edge from the /moai codemaps
	// dependency-graph artifact.
	KindImport = "import"
	// KindMXSpec is a code-location→SPEC edge from an @MX:SPEC sub-line.
	KindMXSpec = "mx-spec"
	// KindSpecDepends is a SPEC→SPEC edge from spec.md frontmatter depends_on.
	KindSpecDepends = "spec-depends"
)

// Tag-kind edge kinds (REQ-MTE-001): one kind per standalone @MX tag kind,
// named by lowercasing the tag kind and prefixing mx-. Uniform over the
// closed mx.TagKind domain — a new scanner tag kind extends this table, it
// does not carve out a special case. mx-* edges carry occurrence + endpoints
// only (kind, source, target, line): tag content, rot state, provenance, and
// wall-clock stay scanner-side (REQ-MTE-007), where the mx sidecar remains
// the single source of truth.
const (
	// KindMXNote is a standalone @MX:NOTE tag occurrence.
	KindMXNote = "mx-note"
	// KindMXWarn is a standalone @MX:WARN tag occurrence.
	KindMXWarn = "mx-warn"
	// KindMXAnchor is a standalone @MX:ANCHOR tag occurrence.
	KindMXAnchor = "mx-anchor"
	// KindMXTodo is a standalone @MX:TODO tag occurrence.
	KindMXTodo = "mx-todo"
	// KindMXLegacy is a standalone @MX:LEGACY tag occurrence.
	KindMXLegacy = "mx-legacy"
	// KindMXDebt is a standalone @MX:DEBT tag occurrence.
	KindMXDebt = "mx-debt"
)

// mxTagEdgeKind maps a scanner tag kind to its edge kind ("" for kinds
// outside the standalone domain — sub-lines never reach the artifact).
func mxTagEdgeKind(k mx.TagKind) string {
	switch k {
	case mx.MXNote:
		return KindMXNote
	case mx.MXWarn:
		return KindMXWarn
	case mx.MXAnchor:
		return KindMXAnchor
	case mx.MXTodo:
		return KindMXTodo
	case mx.MXLegacy:
		return KindMXLegacy
	case mx.MXDebt:
		return KindMXDebt
	default:
		return ""
	}
}

// scanDirFn is the mx scanner seam: one project scan feeds BOTH the mx-spec
// layer and the tag layer (REQ-MTE-006 — no second project walk). Tests
// replace it to count passes.
var scanDirFn = func(projectRoot string) ([]mx.Tag, error) {
	s := mx.NewScanner()
	s.SetIgnorePatterns(mx.DefaultScanIgnore)
	return s.ScanDir(projectRoot)
}

// Edge is one persisted graph edge. Line carries the 1-based line of the
// @MX tag owning the SPEC sub-line (mx-spec edges only; the scanner records
// the owning tag's line, not the sub-line's own position).
type Edge struct {
	Kind   string `json:"kind"`
	Source string `json:"source"`
	Target string `json:"target"`
	Line   int    `json:"line,omitempty"`

	// Grade is the resolution grade used to derive a code-derived edge
	// (full | name-based). Empty on doc-derived edges.
	Grade string `json:"grade,omitempty"`
	// DisagreesWith marks a doc/code layer disagreement on the same
	// relationship (REQ-GF-015): the value names the other layer's claim.
	// Both edges stay in the artifact — never a silent pick.
	DisagreesWith string `json:"disagrees_with,omitempty"`

	// Resolution is how strongly a code-call edge's callee resolution is
	// believed (REQ-GEC-001): how the edge was matched, an EVIDENCE axis
	// orthogonal to Grade's capability axis (REQ-GEC-007). Populated on
	// code-call edges only; omitempty keeps doc-derived and code-import
	// serialization byte-identical to the pre-field artifact (REQ-GEC-006).
	// The appended-after-DisagreesWith position preserves every existing
	// edge's key order.
	Resolution string `json:"resolution,omitempty"`
	// Confidence is the numeric confidence for Resolution, defined as a
	// pure function of it at exactly one point (ConfidenceFor). All three
	// domain values are non-zero, so omitempty never drops a populated
	// confidence; absent on old artifacts (REQ-GEC-009: loaded as unknown).
	Confidence float64 `json:"confidence,omitempty"`
}

// Resolution labels for code-call edges (REQ-GEC-001): the closed 3-value
// domain. extracted = same-file declaration or Go-module import evidence;
// intra-package = same-directory declaration (by-construction package
// proximity, no import evidence); inferred = name-only fallback.
const (
	ResolutionExtracted    = "extracted"
	ResolutionIntraPackage = "intra-package"
	ResolutionInferred     = "inferred"
)

// ConfidenceFor maps a resolution label to its numeric confidence — the
// SINGLE definition of that map (REQ-GEC-001), so the two fields cannot
// drift. Total over the closed 3-value domain; any other label (including
// the empty label of pre-upgrade artifacts) maps to 0, which serializes as
// absent.
//
// @MX:NOTE: [AUTO] single-definition resolution→confidence map — stamp via this, never inline the numbers (SPEC-GRAPH-EDGE-CONFIDENCE-001)
func ConfidenceFor(resolution string) float64 {
	switch resolution {
	case ResolutionExtracted:
		return 1.0
	case ResolutionIntraPackage:
		return 0.95
	case ResolutionInferred:
		return 0.85
	default:
		return 0
	}
}

// codemapsDepRelPath is the /moai codemaps dep-graph output (read-only seed;
// same path contract as internal/navigator/tiers.blueprint).
const codemapsDepRelPath = ".moai/project/codemaps/dependencies.md"

// Build aggregates the edge layers under projectRoot and returns the sorted
// edge list. Every layer fails open: an absent codemaps artifact, an empty
// @MX scan, or a missing .moai/specs/ tree yields zero edges of that kind,
// never an error — the artifact reflects exactly what exists on disk.
//
// Without range data the tag layer emits the self-edge form for every tag
// (REQ-MTE-015); BuildWithCodeLayersMode supplies the extractor's retained
// ranges so body-anchored tags join to their enclosing symbol (REQ-MTE-002).
func Build(projectRoot string) ([]Edge, error) {
	return buildDocLayers(projectRoot, nil)
}

// buildDocLayers is Build over an optional repo-relative-file → declaration
// range index. tags are scanned ONCE and feed both the mx-spec and the mx-*
// layers (REQ-MTE-006).
func buildDocLayers(projectRoot string, rangesByFile map[string][]astx.FuncRange) ([]Edge, error) {
	imports, err := importEdges(projectRoot)
	if err != nil {
		return nil, err
	}
	tags, err := scanDirFn(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("graph: scan @MX tags: %w", err)
	}
	specLinks := mxSpecEdgesFromTags(tags, projectRoot)
	tagLinks := tagEdgesFromTags(tags, projectRoot, rangesByFile)
	depends, err := specDependsEdges(projectRoot)
	if err != nil {
		return nil, err
	}
	reportLinks, err := reportEdges(projectRoot)
	if err != nil {
		return nil, err
	}

	edges := make([]Edge, 0, len(imports)+len(specLinks)+len(tagLinks)+len(depends)+len(reportLinks))
	edges = append(edges, imports...)
	edges = append(edges, specLinks...)
	edges = append(edges, tagLinks...)
	edges = append(edges, depends...)
	edges = append(edges, reportLinks...)
	sort.Slice(edges, func(i, j int) bool { return EdgeLess(edges[i], edges[j]) })
	return edges, nil
}

// repoRel renders an absolute path as a repo-relative slash path when it sits
// under projectRoot; the absolute form is the fallback (absolute paths are
// machine-specific and would make the artifact non-diffable).
func repoRel(projectRoot, abs string) string {
	if rel, err := filepath.Rel(projectRoot, abs); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return abs
}

// EdgeLess is the canonical edge ordering: kind, then source, then target,
// then line.
func EdgeLess(a, b Edge) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	if a.Source != b.Source {
		return a.Source < b.Source
	}
	if a.Target != b.Target {
		return a.Target < b.Target
	}
	return a.Line < b.Line
}

// WriteJSONL writes edges to path as one JSON object per line, atomically
// (temp file + rename in the target directory, mirroring the mx sidecar
// write pattern). The parent directory is created when absent.
func WriteJSONL(path string, edges []Edge) error {
	var b strings.Builder
	for _, e := range edges {
		line, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("graph: marshal edge: %w", err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("graph: mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".edges-*.jsonl")
	if err != nil {
		return fmt.Errorf("graph: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.WriteString(b.String()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("graph: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("graph: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("graph: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("graph: rename into place: %w", err)
	}
	tmpName = "" // renamed; nothing to clean up
	return nil
}

// importEdges extracts package→package edges from the codemaps
// dependencies.md artifact. When the file carries a mermaid fence, only the
// fence is parsed — the surrounding prose sections ("X → Y" commentary)
// duplicate or pollute the edge set. Short mermaid node ids are resolved to
// full package paths via the fence's node labels.
func importEdges(projectRoot string) ([]Edge, error) {
	content, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(codemapsDepRelPath)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("graph: read codemaps dependencies: %w", err)
	}

	scope := string(content)
	fence := mermaidFence(scope)
	if fence != "" {
		scope = fence
	}
	labels := mermaidLabelPaths(fence)

	var edges []Edge
	seen := map[string]bool{}
	for _, mod := range tiers.ParseDependenciesMarkdown(scope) {
		src, ok := canonicalNode(mod.PackagePath, labels)
		if !ok {
			continue
		}
		for _, dep := range mod.DependsOn {
			dst, ok := canonicalNode(dep, labels)
			if !ok {
				continue
			}
			key := src + "\x00" + dst
			if seen[key] {
				continue
			}
			seen[key] = true
			edges = append(edges, Edge{Kind: KindImport, Source: src, Target: dst})
		}
	}
	return edges, nil
}

// mxSpecEdgesFromTags extracts code-location→SPEC edges from the tags of ONE
// scanner pass (the same pass the tag layer consumes — REQ-MTE-006).
// Source paths are repo-relative (the scanner reports absolute paths;
// absolute paths are machine-specific and would make the artifact
// non-diffable).
func mxSpecEdgesFromTags(tags []mx.Tag, projectRoot string) []Edge {
	var edges []Edge
	for _, tag := range tags {
		if tag.SpecRef == "" {
			continue
		}
		edges = append(edges, Edge{
			Kind:   KindMXSpec,
			Source: repoRel(projectRoot, tag.File),
			Target: tag.SpecRef,
			Line:   tag.Line,
		})
	}
	return edges
}

// tagEdgesFromTags maps the standalone tag occurrences of ONE scanner pass
// into the mx-* edge layer (REQ-MTE-001). Endpoints (REQ-MTE-002): source =
// repo-relative file, target = the innermost declared range containing the
// tag line, else the file itself (self-edge — a file-scope tag or missing
// range data is represented, never dropped). The mapping is a pure function
// of (File, Kind, Line) plus the tree content behind rangesByFile: no
// scanner-mutable metadata (Body, Reason, RotRisk, CreatedBy, LastSeenAt)
// reaches an edge (REQ-MTE-003, REQ-MTE-007).
func tagEdgesFromTags(tags []mx.Tag, projectRoot string, rangesByFile map[string][]astx.FuncRange) []Edge {
	var edges []Edge
	for _, tag := range tags {
		kind := mxTagEdgeKind(tag.Kind)
		if kind == "" {
			continue
		}
		source := repoRel(projectRoot, tag.File)
		target := source // self-edge fallback
		if ranges, ok := rangesByFile[source]; ok {
			if name := enclosingRangeName(ranges, tag.Line); name != "" {
				target = name
			}
		}
		edges = append(edges, Edge{Kind: kind, Source: source, Target: target, Line: tag.Line})
	}
	return edges
}

// enclosingRangeName returns the innermost range containing line — the same
// innermost-wins rule as the seam's enclosingFunction (acceptance §D.2: a
// tag inside a nested closure joins to the innermost declaration).
func enclosingRangeName(ranges []astx.FuncRange, line int) string {
	best := ""
	bestStart := -1
	for _, r := range ranges {
		if line >= r.StartLine && line <= r.EndLine && r.StartLine >= bestStart {
			bestStart = r.StartLine
			best = r.Name
		}
	}
	return best
}

// specDependsEdges extracts SPEC→SPEC edges from spec.md frontmatter
// depends_on via the existing mx spec loader.
func specDependsEdges(projectRoot string) ([]Edge, error) {
	deps, err := mx.LoadSpecDependencies(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("graph: load spec depends_on: %w", err)
	}

	ids := make([]string, 0, len(deps))
	for id := range deps {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var edges []Edge
	seen := map[string]bool{}
	for _, id := range ids {
		list := append([]string(nil), deps[id]...)
		sort.Strings(list)
		for _, dep := range list {
			key := id + "\x00" + dep
			if seen[key] {
				continue
			}
			seen[key] = true
			edges = append(edges, Edge{Kind: KindSpecDepends, Source: id, Target: dep})
		}
	}
	return edges, nil
}

// mermaidFence returns the body of the first ```mermaid code fence, or ""
// when the content has none.
func mermaidFence(content string) string {
	start := strings.Index(content, "```mermaid")
	if start < 0 {
		return ""
	}
	rest := content[start+len("```mermaid"):]
	if end := strings.Index(rest, "```"); end >= 0 {
		return rest[:end]
	}
	return rest
}

// mermaidLabelRe matches a mermaid node definition with a quoted label:
// id["label"]. Edge lines (A --> B) carry no ["..."] suffix and never match.
var mermaidLabelRe = regexp.MustCompile(`([A-Za-z0-9_.-]+)\["([^"]*)"\]`)

// mermaidLabelPaths builds a node-id → package-path table from mermaid node
// labels. A label qualifies only when its first token (before any HTML tag
// such as <br/>) looks like a package path (contains / or .); subgraph
// titles and prose labels never qualify.
func mermaidLabelPaths(fence string) map[string]string {
	out := map[string]string{}
	for _, m := range mermaidLabelRe.FindAllStringSubmatch(fence, -1) {
		id, label := m[1], m[2]
		path := label
		if i := strings.IndexAny(path, "<"); i >= 0 {
			path = path[:i]
		}
		path = strings.TrimSpace(path)
		if path == "" || !strings.ContainsAny(path, "/.") {
			continue
		}
		if _, exists := out[id]; !exists {
			out[id] = path
		}
	}
	return out
}

// canonicalNode resolves a parser token to a package path: the fence label
// table when the token is a declared node id, otherwise the token verbatim.
// Tokens carrying internal whitespace or markdown decoration are scaffold
// noise, not modules — the edge is dropped.
func canonicalNode(token string, labels map[string]string) (string, bool) {
	if p, ok := labels[token]; ok {
		return p, true
	}
	if token == "" || strings.ContainsAny(token, " \t`*") {
		return "", false
	}
	return token, true
}
