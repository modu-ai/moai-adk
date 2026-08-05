//go:build cgo

package astx

import (
	"context"
	_ "embed"
	"log/slog"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Embedded per-language query files. Each capture is named @symbol.<kind>
// so the extractor can group captures by kind generically.

//go:embed queries/go.scm
var queryGo []byte

//go:embed queries/python.scm
var queryPython []byte

//go:embed queries/typescript.scm
var queryTypeScript []byte

//go:embed queries/javascript.scm
var queryJavaScript []byte

//go:embed queries/rust.scm
var queryRust []byte

//go:embed queries/java.scm
var queryJava []byte

//go:embed queries/kotlin.scm
var queryKotlin []byte

//go:embed queries/csharp.scm
var queryCSharp []byte

//go:embed queries/ruby.scm
var queryRuby []byte

//go:embed queries/php.scm
var queryPHP []byte

//go:embed queries/elixir.scm
var queryElixir []byte

//go:embed queries/cpp.scm
var queryCPP []byte

//go:embed queries/scala.scm
var queryScala []byte

//go:embed queries/swift.scm
var querySwift []byte

// seedEntry binds a tree-sitter grammar to its embedded .scm query bytes.
type seedEntry struct {
	grammar *sitter.Language
	query   []byte
}

// seededGrammars is the CGO-enabled grammar map. A language present here with
// a compiled query is Supported: true; absent or scaffolded entries return
// Supported: false (see scaffoldedLanguages in astx.go).
//
// Extending to a new language is a row here + a queries/<lang>.scm file — no
// per-language Go logic (REQ-NT-006 data-file extension property).
var seededGrammars = map[string]seedEntry{
	"go":         {grammar: golang.GetLanguage(), query: queryGo},
	"python":     {grammar: python.GetLanguage(), query: queryPython},
	"typescript": {grammar: typescript.GetLanguage(), query: queryTypeScript},
	"javascript": {grammar: javascript.GetLanguage(), query: queryJavaScript},
	"rust":       {grammar: rust.GetLanguage(), query: queryRust},
	"java":       {grammar: java.GetLanguage(), query: queryJava},
	"kotlin":     {grammar: kotlin.GetLanguage(), query: queryKotlin},
	"csharp":     {grammar: csharp.GetLanguage(), query: queryCSharp},
	"ruby":       {grammar: ruby.GetLanguage(), query: queryRuby},
	"php":        {grammar: php.GetLanguage(), query: queryPHP},
	"elixir":     {grammar: elixir.GetLanguage(), query: queryElixir},
	"cpp":        {grammar: cpp.GetLanguage(), query: queryCPP},
	"scala":      {grammar: scala.GetLanguage(), query: queryScala},
	"swift":      {grammar: swift.GetLanguage(), query: querySwift},
}

// extractImpl is the CGO-enabled tree-sitter implementation of Extract.
func extractImpl(language string, sourcePath string) (SymbolSet, error) {
	// Scaffolded languages have no upstream grammar.
	if scaffoldedLanguages[language] {
		return SymbolSet{Supported: false}, nil
	}
	entry, ok := seededGrammars[language]
	if !ok {
		return SymbolSet{Supported: false}, nil
	}

	// Read the source file (fail-open on read error).
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		slog.Debug("astx: read error", "lang", language, "path", sourcePath, "error", err)
		return SymbolSet{Supported: false}, nil
	}

	// File-size guard (mirrors internal/hook/mx/complexity).
	if len(content) > maxFileSizeBytes {
		slog.Debug("astx: file exceeds size cap", "lang", language, "path", sourcePath, "bytes", len(content))
		return SymbolSet{Supported: false}, nil
	}

	// Parse with tree-sitter (tree-sitter recovers from syntax errors and
	// returns a partial tree, so malformed-but-readable files still yield
	// captures — REQ-NT-008 malformed-source tolerance).
	parser := sitter.NewParser()
	parser.SetLanguage(entry.grammar)
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil || tree == nil {
		if err != nil {
			slog.Debug("astx: parse error", "lang", language, "path", sourcePath, "error", err)
		}
		return SymbolSet{Supported: false}, nil
	}
	defer tree.Close()
	root := tree.RootNode()

	// Compile the query (fail-open on compile error — logged, not fatal).
	q, err := sitter.NewQuery(entry.query, entry.grammar)
	if err != nil {
		slog.Debug("astx: query compile error", "lang", language, "error", err)
		return SymbolSet{Supported: false}, nil
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(q, root)

	symbols := map[string][]Symbol{}
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}
		for _, cap := range match.Captures {
			kind := q.CaptureNameForId(cap.Index)
			// Only honor @symbol.<kind> captures.
			if len(kind) <= len("symbol.") || kind[:len("symbol.")] != "symbol." {
				continue
			}
			symKind := kind[len("symbol."):]
			name := string(content[cap.Node.StartByte():cap.Node.EndByte()])
			symbols[symKind] = append(symbols[symKind], Symbol{
				Name: name,
				Kind: symKind,
				File: sourcePath,
				Line: int(cap.Node.StartPoint().Row) + 1,
			})
		}
	}

	return SymbolSet{
		Supported:   true,
		Symbols:     symbols,
		SourceBytes: int64(len(content)),
	}, nil
}
