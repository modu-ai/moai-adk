package bugreport

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestKindEnumClosed(t *testing.T) {
	rows := KindRegister()
	if len(rows) != 6 {
		t.Fatalf("kind register carries %d rows, want exactly 6", len(rows))
	}
	seen := map[Kind]bool{}
	for _, row := range rows {
		if !row.Kind.Valid() {
			t.Errorf("register row kind %q is not one of the six", row.Kind)
		}
		if seen[row.Kind] {
			t.Errorf("kind %q registered twice", row.Kind)
		}
		seen[row.Kind] = true
		if !row.VerdictRule.Valid() {
			t.Errorf("kind %q carries no verdict rule", row.Kind)
		}
		if row.Derivation != DerivationDerived && row.Derivation != DerivationCallSite {
			t.Errorf("kind %q derivation %q is not one of the two modes", row.Kind, row.Derivation)
		}
		if len(row.Sites) == 0 {
			t.Errorf("kind %q has no registered emit site", row.Kind)
		}
		// Every token the row lists must be a member of the closed set and
		// valid for the kind (the two registers agree).
		for _, tok := range row.Tokens {
			if !tok.ValidForKind(row.Kind) {
				t.Errorf("kind %q lists token %q that is not valid for it", row.Kind, tok)
			}
			v, ok := row.TokenVerdicts[tok]
			if !ok || !v.Valid() {
				t.Errorf("kind %q token %q has no verdict in the register", row.Kind, tok)
			}
		}
	}

	// The harness_defect token set is exactly missing_key, unexpanded_token,
	// invalid_json.
	for _, row := range rows {
		if row.Kind != KindHarnessDefect {
			continue
		}
		if len(row.Tokens) != 3 {
			t.Fatalf("harness_defect carries %d tokens, want exactly 3", len(row.Tokens))
		}
		want := map[TemplateToken]bool{TokenMissingKey: true, TokenUnexpandedToken: true, TokenInvalidJSON: true}
		for _, tok := range row.Tokens {
			if !want[tok] {
				t.Errorf("harness_defect carries unexpected token %q", tok)
			}
			delete(want, tok)
		}
		if len(want) != 0 {
			t.Errorf("harness_defect is missing tokens %v", want)
		}
	}
}

// TestEmitSiteRegister pins that every registered emit-site lead names a real
// location in this tree: the file exists and carries the capture call or, for
// the hook registry, the capture branch. The go/parser walk in
// TestEveryRecoverSiteReportsOrIsAllowlisted is the AUTHORITATIVE inventory;
// this test keeps the register's leads honest so a stale lead is visible.
func TestEmitSiteRegister(t *testing.T) {
	for _, row := range KindRegister() {
		for _, site := range row.Sites {
			file := site
			if i := strings.IndexByte(file, ':'); i >= 0 {
				file = file[:i]
			}
			if _, err := os.Stat(filepath.Join("..", "..", file)); err != nil {
				t.Errorf("kind %q site lead %q names a missing file", row.Kind, site)
			}
		}
	}

	// Every registered event carries a well-formed name and every moai-kind
	// site row declares its derivation mode (the cli-side parser guard asserts
	// the registered event set equals internal/hook's declared constants).
	for _, row := range KindRegister() {
		if row.VerdictRule == VerdictMoai && row.Derivation == "" {
			t.Errorf("kind %q: a moai verdict without a declared derivation mode", row.Kind)
		}
	}
}

// TestRegisterRowsDeclareDerivationMode is AC-008's last arm: every register
// row states whether its verdict is derived inside bugreport or asserted by
// the call site.
func TestRegisterRowsDeclareDerivationMode(t *testing.T) {
	for _, row := range KindRegister() {
		switch row.Derivation {
		case DerivationDerived, DerivationCallSite:
			// declared
		default:
			t.Errorf("kind %q derivation %q undeclared", row.Kind, row.Derivation)
		}
	}
}

// TestEveryRecoverSiteReportsOrIsAllowlisted is the REQ-ANON-006 guard: a
// go/parser walk over every non-test file in internal/, cmd/, and pkg/ —
// every recover() call site either sits in a function that calls the capture
// entry point or appears on the reasoned allowlist. The walk is the
// authoritative inventory: an unlisted site surfaces here as a failure, never
// as a silent omission.
func TestEveryRecoverSiteReportsOrIsAllowlisted(t *testing.T) {
	allowlisted := map[string]bool{
		// Benign by construction: the recover defends a closed-channel send
		// race after Close() and drops one trace entry by design.
		"internal/hook/trace/writer.go:Write": true,
		// The capture entry point's own fail-open recover: a capture bug must
		// never take the host down, and capturing a capture is infinite.
		"internal/bugreport/capture.go:Capture": true,
	}

	fset := token.NewFileSet()
	found := 0
	// root, repoPrefix pairs: the walk reads the sibling trees of this
	// package — internal/ (..), and the repo root's cmd/ and pkg/ (../../).
	for _, tc := range []struct{ root, repoPrefix string }{
		{"..", "internal"},
		{"../../cmd", "cmd"},
		{"../../pkg", "pkg"},
	} {
		root, repoPrefix := tc.root, tc.repoPrefix
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			name := d.Name()
			if d.IsDir() {
				if name == "testdata" || name == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				return nil
			}
			// Repo-relative path: "internal/hook/x.go", "cmd/moai/main.go".
			trimmed := strings.TrimPrefix(filepath.ToSlash(path), "../")
			trimmed = strings.TrimPrefix(trimmed, "../")
			rel := repoPrefix + "/" + trimmed

			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Errorf("parse %s: %v", rel, err)
				return nil
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				hasRecover := false
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					ce, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					if ident, ok := ce.Fun.(*ast.Ident); ok && ident.Name == "recover" {
						hasRecover = true
					}
					return true
				})
				if !hasRecover {
					continue
				}
				found++
				if hasCaptureCall(fn) {
					return nil
				}
				if allowlisted[rel+":"+fn.Name.Name] {
					return nil
				}
				t.Errorf("recover site %s:%s neither calls the capture entry point nor is allowlisted", rel, fn.Name.Name)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if found == 0 {
		t.Fatal("the walk found no recover sites — it swept nothing, so its green proves nothing (check the roots)")
	}
}

// hasCaptureCall reports whether the function's body calls the capture entry
// point (bugreport.Capture in other packages).
func hasCaptureCall(fn *ast.FuncDecl) bool {
	captured := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := ce.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Capture" {
			captured = true
		}
		return true
	})
	return captured
}

// TestUserAndToolFailuresNeverCaptured is AC-007's discard arm: signals from
// a user configuration error and from a user test-run failure (the exec
// reason token standing in for the exec-typed causes bugreport cannot name)
// are offered to capture with consent ON and record nothing.
func TestUserAndToolFailuresNeverCaptured(t *testing.T) {
	path := spoolPathForTest(t)

	// Consent ON: a user file that enables capture (at the reader's real
	// path, <home>/config/participation.yaml), so the assertions below
	// measure the attribution drop, not the consent gate.
	configDir := filepath.Join(os.Getenv("MOAI_HOME"), "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "participation:\n  enabled: true\n  asked: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}

	Capture(KindHookHandlerFailure, config.ErrInvalidConfig, "", nil)                   // user-config error → user → dropped
	Capture(KindHookHandlerFailure, errString("user test run failed"), ReasonExec, nil) // user test-run failure → environment → dropped
	Capture(KindHookHandlerFailure, errString("tool failure event"), ReasonExec, nil)   // tool-failure-shaped exit → environment → dropped

	if got := spoolLineCount(t, path); got != 0 {
		t.Fatalf("spool carries %d line(s) after user/tool-failure signals, want none", got)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

// TestToolFailureHandlersDoNotImportCapture is AC-007's structural exclusion:
// the two Claude Code tool-failure handler files contain no reference to the
// capture package, so no signal can ever be offered from there.
func TestToolFailureHandlersDoNotImportCapture(t *testing.T) {
	for _, file := range []string{"../hook/post_tool_failure.go", "../hook/failure_observer.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(raw), "bugreport") {
			t.Errorf("%s references the capture package; tool-failure handlers must never capture", file)
		}
	}
}
