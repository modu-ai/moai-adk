//go:build cgo

package complexity

import (
	_ "embed"
	"log/slog"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Embedded query files for each seeded language.
//
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

// langEntry holds the tree-sitter language and decision-node query for one language.
type langEntry struct {
	language     *sitter.Language
	decisionSCM  []byte
	ifBranchSCM  []byte
	funcNodeType []string // node types that represent function definitions
}

// seededLanguages maps language identifiers to their tree-sitter configuration.
var seededLanguages = map[string]*langEntry{
	"go": {
		language:    golang.GetLanguage(),
		decisionSCM: queryGo,
		funcNodeType: []string{"function_declaration", "method_declaration"},
	},
	"python": {
		language:    python.GetLanguage(),
		decisionSCM: queryPython,
		funcNodeType: []string{"function_definition"},
	},
	"typescript": {
		language:    typescript.GetLanguage(),
		decisionSCM: queryTypeScript,
		funcNodeType: []string{"function_declaration", "method_definition", "function"},
	},
	"javascript": {
		language:    javascript.GetLanguage(),
		decisionSCM: queryJavaScript,
		funcNodeType: []string{"function_declaration", "method_definition", "function"},
	},
	"rust": {
		language:    rust.GetLanguage(),
		decisionSCM: queryRust,
		funcNodeType: []string{"function_item"},
	},
}

// scaffoldedLanguages are registered but not yet seeded with tree-sitter grammars.
// They return Result{Supported: false} without attempting any parse.
var scaffoldedLanguages = map[string]bool{
	"java":    true,
	"kotlin":  true,
	"csharp":  true,
	"ruby":    true,
	"php":     true,
	"elixir":  true,
	"cpp":     true,
	"scala":   true,
	"r":       true,
	"flutter": true,
	"swift":   true,
}

// measure is the CGO implementation using tree-sitter.
func measure(lang string, content []byte, funcName string, startLine int) (Result, error) {
	// File size guard: refuse inputs > 1 MiB (REQ-UTIL-001-032)
	if len(content) > maxFileSizeBytes {
		return Result{Supported: false}, nil
	}

	// Scaffolded language: return stub immediately.
	if scaffoldedLanguages[lang] {
		return Result{Supported: false}, nil
	}

	entry, ok := seededLanguages[lang]
	if !ok {
		return Result{Supported: false}, nil
	}

	// Parse the file.
	parser := sitter.NewParser()
	parser.SetLanguage(entry.language)
	tree := parser.Parse(nil, content)
	if tree == nil {
		return Result{Supported: false}, nil
	}

	root := tree.RootNode()

	// Find the function node matching funcName (and approximately startLine).
	funcNode := findFunctionNode(root, content, funcName, entry.funcNodeType, startLine)
	if funcNode == nil {
		slog.Debug("complexity: function not found in AST", "lang", lang, "func", funcName)
		return Result{Supported: false}, nil
	}

	// Count decision nodes within the function's byte range.
	startByte := funcNode.StartByte()
	endByte := funcNode.EndByte()

	// Parse the decision query (includes both @decision and @if_branch captures).
	decisionQuery, err := sitter.NewQuery(entry.decisionSCM, entry.language)
	if err != nil {
		slog.Debug("complexity: query compile error", "lang", lang, "error", err)
		return Result{Supported: false}, nil
	}

	// Use SetPointRange to constrain the cursor to the function's source range.
	cursor := sitter.NewQueryCursor()
	cursor.SetPointRange(funcNode.StartPoint(), funcNode.EndPoint())
	cursor.Exec(decisionQuery, root)

	var decisionCount, ifBranchCount int
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}
		for _, capture := range match.Captures {
			// Extra guard: ensure the captured node is within the function byte range.
			if capture.Node.StartByte() < startByte || capture.Node.EndByte() > endByte {
				continue
			}
			captureName := decisionQuery.CaptureNameForId(capture.Index)
			switch captureName {
			case "decision":
				decisionCount++
			case "if_branch":
				ifBranchCount++
			}
		}
	}

	return Result{
		Supported:  true,
		Cyclomatic: decisionCount + 1, // McCabe = decision nodes + 1
		IfBranches: ifBranchCount,
	}, nil
}

// findFunctionNode walks the AST to locate a function/method node by name.
// nodeTypes is the list of AST node types that represent function definitions for this language.
// startLine is a 1-indexed hint for disambiguation; 0 means pick the first match.
func findFunctionNode(node *sitter.Node, content []byte, funcName string, nodeTypes []string, startLine int) *sitter.Node {
	nodeType := node.Type()

	for _, ft := range nodeTypes {
		if nodeType == ft {
			if nameNode := nameChildOf(node, nodeType); nameNode != nil {
				name := string(content[nameNode.StartByte():nameNode.EndByte()])
				if name == funcName {
					if startLine == 0 {
						return node
					}
					// Check that function starts within ±5 lines of startLine.
					nodeLine := int(node.StartPoint().Row) + 1 // convert to 1-indexed
					if abs(nodeLine-startLine) <= 5 {
						return node
					}
				}
			}
		}
	}

	// Recurse into children.
	for i := 0; i < int(node.ChildCount()); i++ {
		if found := findFunctionNode(node.Child(i), content, funcName, nodeTypes, startLine); found != nil {
			return found
		}
	}
	return nil
}

// nameChildOf returns the name child node for a function declaration node.
// The field name for the function name differs across grammars.
func nameChildOf(node *sitter.Node, nodeType string) *sitter.Node {
	// Most languages use "name" as the field name for function identifier.
	if n := node.ChildByFieldName("name"); n != nil {
		return n
	}
	// Fallback: walk children looking for identifier-like node types.
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier", "field_identifier", "property_identifier":
			return child
		}
	}
	return nil
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
