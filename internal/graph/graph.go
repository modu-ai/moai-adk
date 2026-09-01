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

// Build aggregates the three edge layers under projectRoot and returns the
// sorted edge list. Every layer fails open: an absent codemaps artifact, an
// empty @MX scan, or a missing .moai/specs/ tree yields zero edges of that
// kind, never an error — the artifact reflects exactly what exists on disk.
func Build(projectRoot string) ([]Edge, error) {
	imports, err := importEdges(projectRoot)
	if err != nil {
		return nil, err
	}
	specLinks, err := mxSpecEdges(projectRoot)
	if err != nil {
		return nil, err
	}
	depends, err := specDependsEdges(projectRoot)
	if err != nil {
		return nil, err
	}
	reportLinks, err := reportEdges(projectRoot)
	if err != nil {
		return nil, err
	}

	edges := make([]Edge, 0, len(imports)+len(specLinks)+len(depends)+len(reportLinks))
	edges = append(edges, imports...)
	edges = append(edges, specLinks...)
	edges = append(edges, depends...)
	edges = append(edges, reportLinks...)
	sort.Slice(edges, func(i, j int) bool { return EdgeLess(edges[i], edges[j]) })
	return edges, nil
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

// mxSpecEdges extracts code-location→SPEC edges from @MX:SPEC sub-lines via
// the existing mx scanner. Source paths are repo-relative (the scanner
// reports absolute paths; absolute paths are machine-specific and would make
// the artifact non-diffable).
func mxSpecEdges(projectRoot string) ([]Edge, error) {
	s := mx.NewScanner()
	s.SetIgnorePatterns(mx.DefaultScanIgnore)
	tags, err := s.ScanDir(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("graph: scan @MX tags: %w", err)
	}

	var edges []Edge
	for _, tag := range tags {
		if tag.SpecRef == "" {
			continue
		}
		source := tag.File
		if rel, relErr := filepath.Rel(projectRoot, tag.File); relErr == nil && !strings.HasPrefix(rel, "..") {
			source = filepath.ToSlash(rel)
		}
		edges = append(edges, Edge{
			Kind:   KindMXSpec,
			Source: source,
			Target: tag.SpecRef,
			Line:   tag.Line,
		})
	}
	return edges, nil
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
