package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestEveryRegisteredHookHandlerHasBugreportName is AC-007's wiring guard:
// every handler the production wiring registers (internal/cli/deps.go) has
// its type name in bugreport's registered table, and the table carries no
// name the wiring does not register. A handler wired without a registration
// would be silently skipped by the dispatch-time capture (fail-closed in the
// safe direction); this guard makes the omission loud instead.
func TestEveryRegisteredHookHandlerHasBugreportName(t *testing.T) {
	names := bugreport.RegisteredHandlerNames()
	table := map[string]bool{}
	for _, n := range names {
		table[n] = true
	}

	wired := wiredHandlerTypeNames(t)
	for name := range wired {
		if !table[name] {
			t.Errorf("wired handler type %q is not registered with bugreport — add it to internal/hook's registration list", name)
		}
	}
	for name := range table {
		if !wired[name] {
			t.Errorf("registered handler type %q is not wired in deps.go — drop it from the registration list", name)
		}
	}

	// The registered event set equals the EventType constants declared in
	// internal/hook/types.go (the parser guard design.md section 4 names).
	events := bugreport.RegisteredEventNames()
	eventSet := map[string]bool{}
	for _, e := range events {
		eventSet[e] = true
	}
	declared := declaredHookEventConstants(t)
	for _, c := range declared {
		if !eventSet[c] {
			t.Errorf("EventType constant %q is not registered with bugreport", c)
		}
	}
	if len(events) != len(declared) {
		t.Errorf("registered event set carries %d names, want %d (the declared constants)", len(events), len(declared))
	}
}

// wiredHandlerTypeNames parses internal/cli/deps.go and normalizes every
// hook handler constructor's name to the handler's type name via the repo's
// uniform convention: strip a leading "New"/"build", cut after the first
// "Handler", lowercase the first letter. WithEscalationConfig returns the
// inner handler unchanged, so its inner constructor's name is the identity.
func wiredHandlerTypeNames(t *testing.T) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "deps.go", nil, 0)
	if err != nil {
		t.Fatalf("parse deps.go: %v", err)
	}

	constructorRe := regexp.MustCompile(`\b(?:New[A-Za-z0-9]*Handler[A-Za-z0-9]*|buildSessionEndHandler)\b`)
	wired := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Register" {
			return true
		}
		// The production call is deps.HookRegistry.Register(...): the
		// receiver chain must name deps and HookRegistry, so no other
		// .Register( in the file joins the wiring set.
		var recv strings.Builder
		ast.Inspect(sel.X, func(m ast.Node) bool {
			if id, ok := m.(*ast.Ident); ok {
				recv.WriteString(id.Name)
				recv.WriteString(" ")
			}
			return true
		})
		if !strings.Contains(recv.String(), "deps") || !strings.Contains(recv.String(), "HookRegistry") {
			return true
		}
		// Collect every handler-constructor identifier inside the call's
		// argument expressions (WithEscalationConfig's inner constructor
		// included — it returns the inner handler unchanged).
		var expr strings.Builder
		for _, arg := range ce.Args {
			ast.Inspect(arg, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					expr.WriteString(id.Name)
					expr.WriteString(" ")
				}
				return true
			})
		}
		for _, m := range constructorRe.FindAllString(expr.String(), -1) {
			wired[normalizeHandlerTypeName(m)] = true
		}
		return true
	})
	if len(wired) == 0 {
		t.Fatal("no Register(...) calls found in deps.go — the walk swept nothing, so its verdict proves nothing")
	}
	return wired
}

// normalizeHandlerTypeName maps a constructor identifier to the handler type
// name its %T spelling produces (internal/hook.BugreportHandlerName's form).
func normalizeHandlerTypeName(constructor string) string {
	name := strings.TrimPrefix(constructor, "New")
	name = strings.TrimPrefix(name, "build")
	if i := strings.Index(name, "Handler"); i >= 0 {
		name = name[:i+len("Handler")]
	}
	if name == "" {
		return ""
	}
	return strings.ToLower(name[:1]) + name[1:]
}

// declaredHookEventConstants parses internal/hook/types.go and returns every
// EventType constant value declared there.
func declaredHookEventConstants(t *testing.T) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../hook/types.go", nil, 0)
	if err != nil {
		t.Fatalf("parse ../hook/types.go: %v", err)
	}
	var out []string
	ast.Inspect(file, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok || vs.Type == nil {
			return true
		}
		if id, ok := vs.Type.(*ast.Ident); !ok || id.Name != "EventType" {
			return true
		}
		for _, v := range vs.Values {
			bl, ok := v.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				continue
			}
			out = append(out, strings.Trim(bl.Value, "`\""))
		}
		return true
	})
	if len(out) == 0 {
		t.Fatal("no EventType constants found in internal/hook/types.go — the sweep swept nothing")
	}
	return out
}
