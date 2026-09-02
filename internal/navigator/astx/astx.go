// Package astx extracts source-tree symbols via tree-sitter for the
// /moai codemaps AST-enrichment step (SPEC-PROJECT-NAVIGATOR-003).
//
// It is a sibling package to internal/hook/mx/complexity (the prior-art
// tree-sitter consumer); the two share no import edge and serve different
// consumers (complexity scoring vs codemaps symbol enrichment).
//
// Build constraints (mirrors internal/hook/mx/complexity):
//   - CGO enabled: full tree-sitter implementation (measure_cgo.go)
//   - CGO disabled: stub returning Supported: false (measure_nocgo.go)
//
// Language registration is data-driven: the 14 working grammars are seeded
// with per-language .scm query files (queries/*.scm); r and flutter are
// registered but scaffolded (Supported: false) pending upstream grammar
// availability. Adding a language is a registration-table row + a .scm file
// (no per-language Go logic) per REQ-NT-006.
package astx

import "strings"

// Symbol is a single named definition extracted from a source file.
type Symbol struct {
	// Name is the identifier text of the symbol.
	Name string
	// Kind is the symbol category: "function", "method", "type", etc.
	Kind string
	// File is the source file path the symbol was extracted from.
	File string
	// Line is the 1-indexed line where the symbol name appears.
	Line int
}

// SymbolSet is the result of extracting symbols from one source file.
type SymbolSet struct {
	// Supported is false when the language has no seeded grammar (scaffolded
	// languages r/flutter, or any CGO-disabled build).
	Supported bool
	// Symbols groups extracted symbols by kind ("function", "type", ...).
	Symbols map[string][]Symbol
	// SourceBytes is the byte length of the parsed source.
	SourceBytes int64
	// Error carries a non-fatal parse/read error when present; callers
	// treat it as advisory (the extractor never aborts on a single file).
	Error error
}

// maxFileSizeBytes caps single-file parse input (1 MiB) to bound tree-sitter
// memory, mirroring internal/hook/mx/complexity.
const maxFileSizeBytes = 1 << 20

// langMeta is the build-tag-independent metadata for one registered language.
type langMeta struct {
	Name       string
	Extensions []string
}

// supportedLanguages is the single registration list for all known languages.
// The 14 working languages are seeded with grammars in measure_cgo.go;
// r and flutter are scaffolded (Supported: false) pending upstream grammars.
// Order is stable so SupportedLanguages() output is deterministic.
var supportedLanguages = []langMeta{
	{Name: "go", Extensions: []string{".go"}},
	{Name: "python", Extensions: []string{".py"}},
	{Name: "typescript", Extensions: []string{".ts", ".tsx"}},
	{Name: "javascript", Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}},
	{Name: "rust", Extensions: []string{".rs"}},
	{Name: "java", Extensions: []string{".java"}},
	{Name: "kotlin", Extensions: []string{".kt", ".kts"}},
	{Name: "csharp", Extensions: []string{".cs"}},
	{Name: "ruby", Extensions: []string{".rb"}},
	{Name: "php", Extensions: []string{".php"}},
	{Name: "elixir", Extensions: []string{".ex", ".exs"}},
	{Name: "cpp", Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hh"}},
	{Name: "scala", Extensions: []string{".scala", ".sc"}},
	{Name: "swift", Extensions: []string{".swift"}},
	{Name: "r", Extensions: []string{".r", ".R"}},
	{Name: "flutter", Extensions: []string{".dart"}},
}

// scaffoldedLanguages are registered but have no tree-sitter grammar
// available in github.com/smacker/go-tree-sitter; Extract returns
// Supported: false for them without attempting a parse.
var scaffoldedLanguages = map[string]bool{
	"r":       true,
	"flutter": true,
}

// SupportedLanguages returns the names of all registered languages (14 working
// + 2 scaffolded). The 14 working languages are seeded with grammars under a
// CGO-enabled build; r and flutter are always Supported: false.
//
// @MX:NOTE: [AUTO] registration order is stable for deterministic output
// @MX:SPEC:SPEC-PROJECT-NAVIGATOR-003
func SupportedLanguages() []string {
	out := make([]string, 0, len(supportedLanguages))
	for _, m := range supportedLanguages {
		out = append(out, m.Name)
	}
	return out
}

// DetectLanguage resolves a language name from a filename by extension using
// the registration table. It returns "" when no registered extension matches.
func DetectLanguage(filename string) string {
	ext := fileExt(filename)
	for _, m := range supportedLanguages {
		for _, e := range m.Extensions {
			if e == ext {
				return m.Name
			}
		}
	}
	return ""
}

// IsScaffolded reports whether the named language is registered-but-unsupported
// (r, flutter). Unknown languages return true (treated as not parseable).
func IsScaffolded(language string) bool {
	return scaffoldedLanguages[language]
}

// Extract parses sourcePath with the tree-sitter grammar for language and
// returns the captured symbols grouped by kind.
//
// Returns SymbolSet{Supported: false} when:
//   - the language is scaffolded or unregistered,
//   - the build is CGO-disabled (measure_nocgo.go),
//   - the file exceeds maxFileSizeBytes,
//   - or a parse/query error occurs (logged via the extractor, never propagated
//     as a fatal error — Error carries the detail for advisory use).
//
// Never panics.
//
// @MX:ANCHOR: [AUTO] per-file extraction entry point; high fan_in (called once per walked source file)
// @MX:REASON: public API boundary consumed by EnrichRows + the navigator-enrich CLI + future downstream tooling
// @MX:SPEC:SPEC-PROJECT-NAVIGATOR-003
func Extract(language string, sourcePath string) (SymbolSet, error) {
	return extractImpl(language, sourcePath)
}

// Resolution grades for the per-language matrix (REQ-GF-016). A grade is a
// claim about CALL-EDGE resolution capability: full means scope-aware target
// resolution, name-based means edges are name-matched without scope, none
// means no call/import capture is available for the language.
const (
	GradeFull      = "full"
	GradeNameBased = "name-based"
	GradeNone      = "none"
)

// callGradeLanguages carries seeded @code.caller/@func.node/@code.call/
// @code.import captures (name-based grade this milestone). Every other
// registered language grades none honestly — the matrix publishes what the
// queries actually capture, and IsScaffolded-style placeholders never pass
// silently.
var callGradeLanguages = map[string]bool{
	"go":         true,
	"python":     true,
	"javascript": true,
	"typescript": true,
	"java":       true,
	"rust":       true,
}

// GradeFor returns the language's call-resolution grade. Unregistered
// languages grade none.
func GradeFor(language string) string {
	if callGradeLanguages[language] {
		return GradeNameBased
	}
	return GradeNone
}

// CallSite is one extracted call: the callee name as written in source.
// Callers are resolved by containment against FuncRange (the consumer joins
// by line range — the extractor stays language-neutral).
type CallSite struct {
	// Callee is the callee identifier text (last segment for selectors:
	// x.Do → "Do").
	Callee string
	// File is the source file path.
	File string
	// Line is the 1-indexed line of the call.
	Line int
}

// FuncRange brackets one function/method declaration so consumers can join
// call sites to their enclosing caller by line containment.
type FuncRange struct {
	// Name is the declared function/method name.
	Name string
	// File is the source file path.
	File string
	// StartLine and EndLine bracket the declaration (1-indexed, inclusive).
	StartLine int
	EndLine   int
}

// ImportSite is one extracted import statement target.
type ImportSite struct {
	// Module is the imported module/package text as written (quotes stripped).
	Module string
	// File is the source file path.
	File string
	// Line is the 1-indexed line of the import.
	Line int
}

// CallSet is the result of extracting call/import sites from one file.
type CallSet struct {
	// Supported is false when the language has no call captures seeded, is
	// scaffolded, the build is CGO-disabled, or a read/parse/query error
	// occurred (fail-open, never fatal).
	Supported bool
	// Calls lists extracted call sites.
	Calls []CallSite
	// Functions lists declaration ranges for caller joins.
	Functions []FuncRange
	// Imports lists extracted import targets.
	Imports []ImportSite
	// SourceBytes is the byte length of the parsed source.
	SourceBytes int64
}

// ExtractCalls parses sourcePath and returns call/import sites plus function
// ranges. Independent of Extract (symbol extraction) so the two concerns —
// declaration inventory and relationship extraction — evolve separately.
// Never panics; per-file errors are fail-open (Supported: false).
//
// @MX:ANCHOR: [AUTO] call-edge extraction entry point; consumed by the graph builder's code-derived layers
// @MX:REASON: non-navigator consumer seam (REQ-GF-013) — internal/graph imports this without navigator-tier deps
// @MX:SPEC:SPEC-V3R6-GRAPH-FRESHNESS-001
func ExtractCalls(language string, sourcePath string) (CallSet, error) {
	return extractCallsImpl(language, sourcePath)
}

// extractImpl is the build-tag-selected implementation:
// measure_cgo.go (real tree-sitter) or measure_nocgo.go (stub).
// It is defined in those files, not here.

// fileExt returns the file extension of name including the leading dot,
// case-preserved (language detection compares against registered extensions
// which are stored case-sensitively; both forms are registered where needed).
func fileExt(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return ""
	}
	return name[i:]
}
