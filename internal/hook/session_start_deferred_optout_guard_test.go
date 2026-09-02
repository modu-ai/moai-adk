// Package hook deferred-scan opt-out boundary guard.
//
// This test mirrors the structural pattern of TestPropose_NoAskUserQuestion in
// internal/cli/harness/propose_boundary_test.go: a static source scan that
// asserts a boundary contract no type checker can express, plus a non-empty
// population control so the guard cannot pass vacuously.
//
// The contract it guards: a test OUTSIDE this package that calls
// NewSessionStartHandler must pass WithSynchronousDeferredScans.
//
// Why the outside is where the hazard lives. Handle's deferred steps run in a
// background goroutine joined with a bound; the goroutine writes under the
// project directory. A test that hands Handle a t.TempDir owns that directory
// only until its body returns, so a write that outruns the join races the
// cleanup and the test fails on removal rather than on an assertion
// ("unlinkat ... .moai/state: directory not empty").
//
// Inside this package the hazard is already closed for every test at once:
// TestMain flips deferredScansAsync to false for the whole binary, so the
// ~50 Handle-calling tests here need no per-call opt-out. That switch is an
// unexported package variable, so it does not reach a caller in another
// package — which is why the exported WithSynchronousDeferredScans option
// exists, and why nothing but this guard makes a caller remember it.
package hook

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	// handlerCtor is the constructor whose cross-package test call sites carry
	// the obligation.
	handlerCtor = "NewSessionStartHandler"
	// syncOption is the exported opt-out that discharges it.
	syncOption = "WithSynchronousDeferredScans"
)

// moduleRoot walks up from the working directory to the directory holding
// go.mod.
//
// Resolved from the filesystem rather than from `git rev-parse` so the scan
// still runs where git is absent or the checkout is shallow. A guard that
// skips itself under those conditions reports a pass it did not measure.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod found walking up from the working directory; the scan has no root")
		}
		dir = parent
	}
}

// skippedDir reports whether a directory is outside the scan.
//
// internal/hook is excluded because TestMain already covers it (see the file
// comment). The rest are directories that hold no first-party test callers:
// vendored code, git metadata, and the embedded template tree, whose files are
// distributed content rather than this module's compiled sources.
func skippedDir(rel string) bool {
	switch rel {
	case "vendor", ".git", "testdata",
		filepath.Join("internal", "hook"),
		filepath.Join("internal", "template", "templates"):
		return true
	}
	return false
}

// callsSyncOption reports whether a NewSessionStartHandler call passes
// WithSynchronousDeferredScans among its option arguments.
//
// Matched on the selector's final name, so both a qualified call
// (hook.WithSynchronousDeferredScans()) and a same-package one count.
func callsSyncOption(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		optCall, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		if calleeName(optCall.Fun) == syncOption {
			return true
		}
	}
	return false
}

// calleeName returns the final identifier of a call's function expression:
// "F" for both F(...) and pkg.F(...). Anything else yields "".
func calleeName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// TestDeferredScanOptOut_CrossPackageTestCallersOptOut asserts that every
// NewSessionStartHandler call site in a test file outside this package passes
// WithSynchronousDeferredScans.
//
// Production call sites are deliberately out of scope: production WANTS the
// async path, which is the whole point of the deferred step. The obligation
// attaches to _test.go files, where a t.TempDir is torn down at test exit.
func TestDeferredScanOptOut_CrossPackageTestCallersOptOut(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	fset := token.NewFileSet()
	callSites := 0

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil && rel != "." && skippedDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if !strings.Contains(string(src), handlerCtor) {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, src, 0)
		if parseErr != nil {
			t.Errorf("parse %s: %v", path, parseErr)
			return nil
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || calleeName(call.Fun) != handlerCtor {
				return true
			}
			callSites++
			if callsSyncOption(call) {
				return true
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s:%d calls %s without %s.\n"+
				"A test outside internal/hook owns the directory it hands Handle, and "+
				"t.TempDir removes it the moment the test body returns. Without the "+
				"option a deferred step can outrun the join bound and race that removal, "+
				"failing the test on cleanup rather than on an assertion "+
				"(\"unlinkat ... .moai/state: directory not empty\"). The unexported "+
				"deferredScansAsync switch TestMain uses does not reach this package.",
				rel, fset.Position(call.Pos()).Line, handlerCtor, syncOption)
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}

	// Non-empty control. The population this guard watches is small and every
	// member currently complies, so a scan that found nothing would report the
	// same green as a scan that found everything correct. Failing here says the
	// guard stopped reaching its subjects — a moved package, a renamed
	// constructor, a broken walk — rather than that the subjects are clean.
	if callSites == 0 {
		t.Fatalf("scanned 0 cross-package %s test call sites; the guard is not reaching "+
			"its subjects and its pass asserts nothing", handlerCtor)
	}
}
