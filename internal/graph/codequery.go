package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/mx"
	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// maxTraceDepth bounds TraceCalls traversal (shared by the MCP tool
// description — one named policy value, not three drifting literals;
// CR round-2 3855002055).
const maxTraceDepth = 8

// ─── graph_file_api (REQ-GF-017) ───

// FileAPISymbol is one exported declaration with its signature.
type FileAPISymbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Line      int    `json:"line"`
	Signature string `json:"signature"`
	Exported  bool   `json:"exported"`
}

// FileAPIResult answers "what does this file expose" without source bodies.
type FileAPIResult struct {
	File       string          `json:"file"`
	Symbols    []FileAPISymbol `json:"symbols"`
	Provenance string          `json:"provenance"`
}

// FileAPI lists a file's exported declarations with signatures — no function
// bodies (REQ-GF-017/019). Exportedness is a language convention; Go filters
// on the capitalization rule, other languages list their declaration set
// (the kind field distinguishes) since their export notions differ.
func FileAPI(projectRoot, relPath string) (FileAPIResult, error) {
	abs := filepath.Join(projectRoot, filepath.FromSlash(relPath))
	// Trust boundary: relPath arrives from an LLM-facing MCP tool parameter.
	// A `..` component would escape the project root after Join — reject any
	// path that does not resolve to inside projectRoot (audit F1).
	if rel, err := filepath.Rel(projectRoot, abs); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return FileAPIResult{}, fmt.Errorf("graph: file path escapes the project root: %s", relPath)
	}
	// Regular-file guard (CR round-2 3855001937): a FIFO/socket at the named
	// path would hang the read; refuse non-regular files explicitly.
	if info, err := os.Stat(abs); err != nil || !info.Mode().IsRegular() {
		return FileAPIResult{}, fmt.Errorf("graph: not a regular file: %s", relPath)
	}
	lang := astx.DetectLanguage(abs)
	res := FileAPIResult{File: filepath.ToSlash(relPath)}
	if lang == "" {
		return res, fmt.Errorf("graph: unsupported file type: %s", relPath)
	}

	set, err := astx.Extract(lang, abs)
	if err != nil {
		// Preserve the cause (CR round-2 3855002024): the MCP client sees
		// this message and callers may errors.Is/As it.
		return res, fmt.Errorf("graph: extract %s: %w", relPath, err)
	}
	if !set.Supported {
		return res, fmt.Errorf("graph: extraction unsupported for %s (grammar unavailable or CGO disabled)", relPath)
	}

	lines := readLines(abs)
	for _, kind := range sortedKinds(set.Symbols) {
		for _, sym := range set.Symbols[kind] {
			exported := isExported(lang, sym.Name)
			if lang == "go" && !exported {
				continue // Go's export rule is mechanical; apply it strictly
			}
			res.Symbols = append(res.Symbols, FileAPISymbol{
				Name:      sym.Name,
				Kind:      kind,
				Line:      sym.Line,
				Signature: signatureAt(lines, sym.Line, lang),
				Exported:  exported,
			})
		}
	}
	res.Provenance = quickProvenance(projectRoot)
	return res, nil
}

// ─── graph_find_code (REQ-GF-018) ───

// CodeMatch is one find_code hit over the code-derived layer.
type CodeMatch struct {
	Symbol string `json:"Symbol"`
	File   string `json:"File"`
	Line   int    `json:"Line"`
	Grade  string `json:"Grade"`
	Via    string `json:"Via"` // how the layer observed the symbol
}

// FindCode searches the code-derived layer for a symbol name: callee sites
// (where the symbol is called) and caller observations. Returns matches plus
// the answer provenance.
func FindCode(projectRoot, query string) ([]CodeMatch, string, error) {
	edges, err := loadCodeEdges(projectRoot)
	if err != nil {
		return nil, "", err
	}
	var out []CodeMatch
	seen := map[string]bool{}
	for _, e := range edges {
		if e.Kind != KindCodeCall {
			continue
		}
		file, fn := splitCodeNode(e.Source)
		if e.Target == query {
			key := "callee:" + e.Source + ":" + e.Target + ":" + fmt.Sprint(e.Line)
			if !seen[key] {
				seen[key] = true
				out = append(out, CodeMatch{Symbol: e.Target, File: file, Line: e.Line, Grade: e.Grade, Via: "callee (called at)"})
			}
		}
		if fn == query {
			// Key includes the target (CR round-2 3855002033): one line may
			// invoke two symbols (a(b())) — dropping the target would keep
			// only the first edge of the two.
			key := "caller:" + file + ":" + e.Target + ":" + fmt.Sprint(e.Line)
			if !seen[key] {
				seen[key] = true
				out = append(out, CodeMatch{Symbol: fn, File: file, Line: e.Line, Grade: e.Grade, Via: "caller (calls " + e.Target + ")"})
			}
		}
	}
	return out, quickProvenance(projectRoot), nil
}

// ─── graph_trace_calls (REQ-GF-018) ───

// CallTraceEdge is one traversed code-call edge.
type CallTraceEdge struct {
	From  string `json:"from"` // file:func (caller) or bare callee name
	To    string `json:"to"`
	Line  int    `json:"line"`
	Grade string `json:"grade"`
}

// TraceCalls traverses code-call edges from a symbol: callers = who reaches
// the symbol (reverse edges), callees = what the symbol calls (forward
// edges), each up to depth hops. Depth is bounded defensively at
// maxTraceDepth.
func TraceCalls(projectRoot, symbol string, depth int) (callers, callees []CallTraceEdge, err error) {
	if depth <= 0 {
		depth = 1
	}
	if depth > maxTraceDepth {
		depth = maxTraceDepth
	}
	edges, err := loadCodeEdges(projectRoot)
	if err != nil {
		return nil, nil, err
	}

	// Index once (CR round-2 3855002059): byTarget serves the reverse
	// traversal, byCaller the forward — the per-frontier full-slice rescan
	// was depth × |frontier| × |edges| on an interactive tool's hot path.
	byTarget := map[string][]Edge{}
	byCaller := map[string][]Edge{}
	for _, e := range edges {
		if e.Kind != KindCodeCall {
			continue
		}
		byTarget[e.Target] = append(byTarget[e.Target], e)
		if _, fn := splitCodeNode(e.Source); fn != "" {
			byCaller[fn] = append(byCaller[fn], e)
		}
	}

	revSeen := map[string]bool{symbol: true}
	frontier := []string{symbol}
	for hop := 0; hop < depth; hop++ {
		var next []string
		for _, node := range frontier {
			for _, e := range byTarget[node] {
				callers = append(callers, CallTraceEdge{From: e.Source, To: e.Target, Line: e.Line, Grade: e.Grade})
				if _, fn := splitCodeNode(e.Source); fn != "" && !revSeen[fn] {
					revSeen[fn] = true
					next = append(next, fn)
				}
			}
		}
		frontier = next
	}

	fwdSeen := map[string]bool{symbol: true}
	frontier = []string{symbol}
	for hop := 0; hop < depth; hop++ {
		var next []string
		for _, node := range frontier {
			for _, e := range byCaller[node] {
				callees = append(callees, CallTraceEdge{From: e.Source, To: e.Target, Line: e.Line, Grade: e.Grade})
				if !fwdSeen[e.Target] {
					fwdSeen[e.Target] = true
					next = append(next, e.Target)
				}
			}
		}
		frontier = next
	}
	return callers, callees, nil
}

// ─── shared helpers ───

// loadCodeEdges loads the edges artifact, mapping an ABSENT artifact to a
// distinct actionable error (CR round-2 3855002040): the untracked derived
// layer is missing — the remedy is a build, not a bug report.
func loadCodeEdges(projectRoot string) ([]Edge, error) {
	edges, err := LoadJSONL(edgesArtifactPath(projectRoot))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("graph layer absent: %s not found — run 'moai graph build' to create it", edgesArtifactPath(projectRoot))
		}
		return nil, fmt.Errorf("graph: load edges: %w", err)
	}
	return edges, nil
}

// edgesArtifactPath is the default edges.jsonl location.
func edgesArtifactPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "project", "graph", "edges.jsonl")
}

// AnswerProvenance names the tree root + commit an answer was computed from
// (REQ-GF-019): the edges meta sidecar when present, else a fresh anchor.
// Exported for the MCP handlers, which stamp every response with it.
func AnswerProvenance(projectRoot string) string {
	if pv, ok := ReadEdgesMeta(filepath.Join(filepath.Dir(edgesArtifactPath(projectRoot)), MetaFileName)); ok {
		return pv.Describe()
	}
	return (&mx.Provenance{TreeRoot: projectRoot, CommitSHA: mx.GitHead(projectRoot)}).Describe()
}

// quickProvenance is AnswerProvenance's internal alias.
func quickProvenance(projectRoot string) string { return AnswerProvenance(projectRoot) }

// splitCodeNode splits a code-call source "path/File.go:Func" into its parts.
func splitCodeNode(node string) (file, fn string) {
	if i := strings.LastIndex(node, ":"); i > 0 {
		return node[:i], node[i+1:]
	}
	return node, ""
}

// sortedKinds yields deterministic kind iteration.
func sortedKinds(m map[string][]astx.Symbol) []string {
	kinds := make([]string, 0, len(m))
	for k := range m {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}

// isExported applies the language's export convention. Go's rule is a
// first-rune Unicode upper-case letter (CR round-2 3855002067: ASCII-only
// misclassifies exported identifiers like Éxporté).
func isExported(lang, name string) bool {
	if lang == "go" && name != "" {
		r, _ := utf8.DecodeRuneInString(name)
		return unicode.IsUpper(r)
	}
	return true
}

// readLines reads a file into lines (nil on error — signature falls back).
func readLines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return strings.Split(string(data), "\n")
}

// signatureAt reconstructs a declaration signature from source lines: the
// symbol's line through (excluding) the line where the body opens — the
// signature text without the body. Falls back to the symbol line alone.
func signatureAt(lines []string, startLine int, lang string) string {
	if startLine < 1 || startLine > len(lines) {
		return ""
	}
	var b strings.Builder
	for i := startLine - 1; i < len(lines) && i < startLine-1+8; i++ {
		line := lines[i]
		if idx := bodyOpenIndex(line, lang); idx >= 0 {
			b.WriteString(strings.TrimSpace(line[:idx]))
			return strings.TrimSpace(b.String())
		}
		if i > startLine-1 {
			b.WriteString(" ")
		}
		b.WriteString(strings.TrimSpace(line))
	}
	return strings.TrimSpace(b.String())
}

// bodyOpenIndex finds where a declaration's body opens on this line, at
// bracket depth zero. Delimiters inside parameter types — Go's
// `map[string]struct{...}`, Python's `x: int`, Ruby's `a: 1` — sit inside
// brackets or annotations and must not truncate the signature (CR round-2
// 3855002078).
func bodyOpenIndex(line, lang string) int {
	depth := 0
	switch lang {
	case "python", "ruby", "elixir":
		// Body opens at the first depth-zero `:` — a colon inside (...) or
		// [...] is an annotation/default, not the body.
		for i, r := range line {
			switch r {
			case '(', '[':
				depth++
			case ')', ']':
				depth--
			case ':':
				if depth == 0 && i > 0 {
					return i
				}
			}
		}
		return -1
	default:
		// Body opens at the first depth-zero `{` — a brace inside (...) is
		// part of a parameter type.
		for i, r := range line {
			switch r {
			case '(', '[':
				depth++
			case ')', ']':
				depth--
			case '{':
				if depth == 0 && i > 0 {
					return i
				}
			}
		}
		return -1
	}
}
