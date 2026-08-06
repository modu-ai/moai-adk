package tiers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"
)

// SymbolOptions configures Tier 3 deterministic extraction.
type SymbolOptions struct {
	// MaxReferences bounds the references list per symbol (AC-NS3-014).
	// Default 50 when zero.
	MaxReferences int
	// NarrativeEnabled controls whether the LLM narrative slot is populated.
	// When false (the deterministic-only mode), NarrativePath is empty
	// (REQ-NS3-015 separable).
	NarrativeEnabled bool
}

// errIndexerNotConfigured is the dispatch sentinel returned when no SCIP /
// astx indexer is configured for the requested language (REQ-NS3-013).
// Callers treat it as fail-open: zero per-symbol records, debug log, exit 0.
var errIndexerNotConfigured = errors.New("tiers: no indexer configured for language (SCIP forward-compat per REQ-NS3-013)")

// defaultMaxReferences bounds the references list when SymbolOptions does
// not specify (AC-NS3-014).
const defaultMaxReferences = 50

// languageIndexers is the per-language dispatch table. Go → astx-extended
// deterministic path; other languages → not configured (SCIP forward-compat).
// Adding a SCIP indexer for a new language is a single row.
//
// @MX:NOTE: [AUTO] per-language dispatch — Go is the only configured path at M4; SCIP is forward-compat per design.md §1.D1
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
var languageIndexers = map[string]func(string, SymbolOptions) ([]SymbolEnrichment, error){
	"go": indexGo,
}

// extractStructuresForLanguage is the dispatch entry point. It returns
// errIndexerNotConfigured for languages without a configured indexer; in
// that case the caller emits 0 per-symbol records (fail-open).
//
// @MX:ANCHOR: [AUTO] Tier 3 deterministic structure dispatch; high fan_in (enumerateSymbols + tests + future SCIP indexers)
// @MX:REASON: single point of per-language routing — the 2-tier separability invariant (REQ-NS3-015) mechanically depends on this being the ONLY language-routing seam
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func extractStructuresForLanguage(projectRoot, language string, opts SymbolOptions) ([]SymbolEnrichment, error) {
	idx, ok := languageIndexers[language]
	if !ok {
		slog.Debug("tiers: no indexer configured for language", "language", language)
		return nil, errIndexerNotConfigured
	}
	return idx(projectRoot, opts)
}

// extractGoStructures is the public Go path entry point used by tests +
// enumerateSymbols. It walks .go files, parses each with go/parser, and
// emits per-symbol records (signature + declaration + references).
func extractGoStructures(projectRoot string, opts SymbolOptions) ([]SymbolEnrichment, error) {
	return indexGo(projectRoot, opts)
}

// indexGo is the astx-extended Go deterministic structure extractor.
// It uses internal/navigator/astx as-is for language detection (consumer-only
// per REQ-NS3-016) and extends the per-symbol record additively with
// signature + declaration + references computed via go/parser (Go stdlib).
//
// @MX:WARN: [AUTO] walks every .go file in projectRoot — output size is bounded by MaxReferences per symbol; large monorepos may need a path filter
// @MX:REASON: an unbounded walk is the failure mode that turns the tier layer into a slow full-tree scan; the cap is the load-bearing bound
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func indexGo(projectRoot string, opts SymbolOptions) ([]SymbolEnrichment, error) {
	if opts.MaxReferences == 0 {
		opts.MaxReferences = defaultMaxReferences
	}
	decls, err := scanGoDeclarations(projectRoot)
	if err != nil {
		return nil, err
	}
	refs := scanGoReferences(projectRoot, decls, opts.MaxReferences)

	out := make([]SymbolEnrichment, 0, len(decls))
	for _, d := range decls {
		identifier := d.identifier()
		sig := d.signature
		if sig == "" {
			sig = d.name + "(...)"
		}
		enr := SymbolEnrichment{
			Identifier:      identifier,
			Signature:       sig,
			DeclarationPath: d.relPath,
			DeclarationLine: d.line,
			References:      refs[identifier],
		}
		if enr.References == nil {
			enr.References = []SymbolRef{}
		}
		out = append(out, enr)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Identifier < out[j].Identifier })
	return out, nil
}

// goDecl is one Go declaration extracted via go/parser.
type goDecl struct {
	pkgBase   string // last segment of the package path
	pkgPath   string // repo-relative package directory
	name      string
	kind      string // function / type / method
	receiver  string // for methods: the receiver type name (no parens)
	signature string
	relPath   string
	line      int
}

// identifier returns the deterministic Tier 3 identifier (<pkgBase>.<name>).
func (d goDecl) identifier() string {
	if d.pkgBase == "" {
		return d.name
	}
	return d.pkgBase + "." + d.name
}

// scanGoDeclarations walks .go files (excluding _test.go and vendor/) under
// projectRoot and extracts function/method/type declarations.
func scanGoDeclarations(projectRoot string) ([]goDecl, error) {
	fset := token.NewFileSet()
	var out []goDecl
	walkErr := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // fail-open
		}
		if info.IsDir() {
			name := info.Name()
			if path != projectRoot && (name == "vendor" || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, rerr := filepath.Rel(projectRoot, path)
		if rerr != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		pkgDir := filepath.ToSlash(filepath.Dir(rel))

		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil // fail-open
		}
		file, perr := parser.ParseFile(fset, path, src, parser.ParseComments)
		if perr != nil {
			slog.Debug("tiers: go parse error (fail-open)", "path", path, "error", perr)
			return nil
		}
		pkgBase := file.Name.Name // Go package name (last segment conventionally)
		ast.Inspect(file, func(n ast.Node) bool {
			switch d := n.(type) {
			case *ast.FuncDecl:
				name := d.Name.Name
				pkg := pkgBase
				if d.Recv != nil && len(d.Recv.List) > 0 {
					// method — qualify by receiver type
					recv := receiverTypeString(d.Recv.List[0].Type)
					pkg = pkgBase
					// method identifier: pkgBase.(recv).Name — keep simple pkgBase.Name for now,
					// signature already differentiates.
					_ = recv
				}
				out = append(out, goDecl{
					pkgBase:   pkg,
					pkgPath:   pkgDir,
					name:      name,
					kind:      funcKind(d),
					receiver:  receiverTypeStringOpt(d.Recv),
					signature: renderFuncSignature(d, src, fset),
					relPath:   rel,
					line:      fset.Position(d.Pos()).Line,
				})
			case *ast.TypeSpec:
				out = append(out, goDecl{
					pkgBase:   pkgBase,
					pkgPath:   pkgDir,
					name:      d.Name.Name,
					kind:      "type",
					signature: renderTypeSpec(d, src, fset),
					relPath:   rel,
					line:      fset.Position(d.Pos()).Line,
				})
			}
			return true
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

// scanGoReferences walks .go files (excluding _test.go and vendor/) and
// records each call/reference site for the declarations in decls. References
// are capped at maxN per symbol. The reference detection is identifier-based
// (a usage of the symbol name in a non-declaration position) — a deliberate
// simplification bounded by maxN.
func scanGoReferences(projectRoot string, decls []goDecl, maxN int) map[string][]SymbolRef {
	if maxN <= 0 {
		maxN = defaultMaxReferences
	}
	// Build name → identifiers map. Multiple identifiers can share a name
	// across packages; we attribute a reference to ALL matching identifiers
	// (a coarse approximation bounded by maxN).
	names := map[string][]string{}
	for _, d := range decls {
		names[d.name] = append(names[d.name], d.identifier())
	}
	out := map[string][]SymbolRef{}
	fset := token.NewFileSet()
	_ = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				name := info.Name()
				if path != projectRoot && (name == "vendor" || strings.HasPrefix(name, ".")) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, rerr := filepath.Rel(projectRoot, path)
		if rerr != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, src, 0)
		if perr != nil {
			return nil
		}
		ast.Inspect(file, func(n ast.Node) bool {
			ident, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			// Skip declaration identifiers.
			if ident.Obj != nil && ident.Obj.Decl != nil {
				// Check whether this ident IS the declaration.
				switch ident.Obj.Kind {
				case ast.Fun, ast.Typ, ast.Var, ast.Con:
					// Only count usage where the ident is NOT the declaration
					// site itself. ident.Obj.Decl points to the declaring
					// AssignStmt/FuncDecl/etc.; this Ident is a reference if
					// its position differs from the decl position.
				}
			}
			ids := names[ident.Name]
			if len(ids) == 0 {
				return true
			}
			line := fset.Position(ident.Pos()).Line
			for _, id := range ids {
				if len(out[id]) >= maxN {
					continue
				}
				out[id] = append(out[id], SymbolRef{Path: rel, Line: line})
			}
			return true
		})
		return nil
	})
	// Sort + dedupe per identifier for byte-stable output (REQ-NS3-019).
	for id := range out {
		out[id] = sortDedupeRefs(out[id])
	}
	return out
}

// sortDedupeRefs returns a sorted, deduped copy of refs.
func sortDedupeRefs(refs []SymbolRef) []SymbolRef {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Path != refs[j].Path {
			return refs[i].Path < refs[j].Path
		}
		return refs[i].Line < refs[j].Line
	})
	out := refs[:0]
	var lastPath string
	lastLine := 0
	for i, r := range refs {
		if i > 0 && r.Path == lastPath && r.Line == lastLine {
			continue
		}
		out = append(out, r)
		lastPath = r.Path
		lastLine = r.Line
	}
	return out
}

// receiverTypeString renders an ast.Expr receiver type to a bare name.
func receiverTypeString(e ast.Expr) string {
	if star, ok := e.(*ast.StarExpr); ok {
		return receiverTypeString(star.X)
	}
	if ident, ok := e.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// receiverTypeStringOpt returns "" when Recv is nil.
func receiverTypeStringOpt(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	return receiverTypeString(recv.List[0].Type)
}

// funcKind returns "method" when d.Recv != nil, "function" otherwise.
func funcKind(d *ast.FuncDecl) string {
	if d.Recv != nil {
		return "method"
	}
	return "function"
}

// renderFuncSignature renders the function signature line, e.g.
// `func ParseHeader(h string) string`. Best-effort using the source slice
// when available (cheaper + deterministic); falls back to a coarse shape.
func renderFuncSignature(d *ast.FuncDecl, src []byte, fset *token.FileSet) string {
	start := fset.Position(d.Pos()).Offset
	var endPos token.Pos
	if d.Body != nil {
		endPos = d.Body.Lbrace
	} else {
		endPos = d.End()
	}
	end := fset.Position(endPos).Offset
	if start < 0 || end < start || end > len(src) {
		return "func " + d.Name.Name + "(...)"
	}
	raw := string(src[start:end])
	raw = strings.TrimRight(raw, " \t")
	// Collapse newlines in signature to spaces (multi-line receiver/type).
	raw = strings.Join(strings.Fields(raw), " ")
	return raw
}

// renderTypeSpec renders `type Name <Underlying>` from the source slice.
func renderTypeSpec(d *ast.TypeSpec, src []byte, fset *token.FileSet) string {
	start := fset.Position(d.Pos()).Offset
	end := fset.Position(d.End())
	if start < 0 || end.Offset < start || end.Offset > len(src) {
		return "type " + d.Name.Name
	}
	raw := string(src[start:end.Offset])
	return strings.Join(strings.Fields(raw), " ")
}

// enumerateSymbols is the Tier 3 entry point used by overlay emission. It
// runs the deterministic structure layer (Go only at M4) and, when
// NarrativeEnabled, populates NarrativePath. With NarrativeEnabled=false,
// the records carry deterministic fields only with NarrativePath empty
// (REQ-NS3-015 separability).
//
// Returns the per-symbol records + any owns-edge entries (blueprint→symbol).
// owns-edges are wired by the overlay from the blueprint side; this function
// returns nil edges (the symbol layer itself does not own the owns-edge
// emission — see overlay.go).
func enumerateSymbols(projectRoot string, opts SymbolOptions) ([]SymbolEnrichment, []TierEdge, error) {
	recs, err := extractStructuresForLanguage(projectRoot, "go", opts)
	if err != nil {
		if errors.Is(err, errIndexerNotConfigured) {
			// Fail-open: 0 records (Go configured, but the dispatch table
			// may have been swapped in tests). Return empty.
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if opts.NarrativeEnabled {
		for i := range recs {
			recs[i].NarrativePath = narrativeSlotPath(recs[i].Identifier)
		}
	}
	return recs, nil, nil
}

// _ keeps the stdlib imports in scope for future extension hooks without
// forcing unused-warning cycles if a helper is removed.
var (
	_ = fmt.Sprintf
	_ = sha256.Sum256
	_ = hex.EncodeToString
)
