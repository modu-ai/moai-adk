package guardliveness

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// deliverableSources is this SPEC's own non-test source set, the scope every
// mechanical instrument in acceptance.md runs over. A package-wide sweep would
// count code this card did not write.
var deliverableSources = []string{
	"contract.go",
	"evaluator.go",
	"store.go",
	"advisory.go",
	"../hook/session_start_guard_liveness.go",
}

// mutatingImports are the packages an outbound call — a forge mutation, a
// shelled-out command, a repository write — would have to travel through.
// Naming the import rather than only the call is what keeps the count robust
// against an aliased selector.
var mutatingImports = map[string]bool{
	"os/exec":  true,
	"net":      true,
	"net/http": true,
	"net/url":  true,
	"github.com/modu-ai/moai-adk/internal/git": true,
}

// mutatingCalls are call names that reach a forge or a shell regardless of the
// package they are reached through.
var mutatingCalls = map[string]bool{
	"Command":               true,
	"CommandContext":        true,
	"Post":                  true,
	"PostForm":              true,
	"Do":                    true,
	"NewRequest":            true,
	"NewRequestWithContext": true,
}

// AC-GDL-008 (a) — the advisory path issues ZERO mutating forge calls, counted
// at the call layer over the deliverable's own non-test source.
//
// Clause (a) alone is satisfied by a renderer that writes its result cache into
// the repository working tree — no forge mutation, but a repository write that
// shows up as drift for the next reader. That direction is clause (b), asserted
// at the host surface where the real render runs.
func TestAdvisoryPathIssuesNoMutatingCall(t *testing.T) {
	fset := token.NewFileSet()
	var found []string

	for _, path := range deliverableSources {
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquote import in %s: %v", path, err)
			}
			if mutatingImports[imported] {
				found = append(found, path+": imports "+imported)
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall {
				return true
			}
			sel, isSel := call.Fun.(*ast.SelectorExpr)
			if !isSel {
				return true
			}
			if mutatingCalls[sel.Sel.Name] {
				pos := fset.Position(call.Pos())
				found = append(found, path+":"+strconv.Itoa(pos.Line)+": calls "+sel.Sel.Name)
			}
			return true
		})
	}

	if len(found) != 0 {
		t.Fatalf("the advisory path carries %d mutating call site(s), want 0:\n%s",
			len(found), strings.Join(found, "\n"))
	}
}
