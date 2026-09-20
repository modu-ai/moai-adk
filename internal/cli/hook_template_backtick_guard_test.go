package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The git-hook templates under internal/template/templates/.git_hooks are held
// verbatim in Go raw string literals (prePushHookContent,
// preCommitHookContent), and a Go raw string literal cannot contain a backtick.
// The project therefore carries a convention: those template files contain no
// backtick at all.
//
// The existing byte-identity pair tests (TestPrePushTemplateMatchesConstant,
// TestPreCommitTemplateMatchesConstant) do NOT enforce that convention. They
// compare the constant's VALUE against the file, so a backtick introduced into
// a template can be "repaired" by rewriting the constant as a concatenation
//
//	const c = `head ` + "`" + ` tail`
//
// which compiles, produces the same value, and leaves the pair tests green.
// That workaround is not hypothetical — shellSyntaxChars in goal_runnable.go
// uses exactly this form. The pair tests guard the PAIR; these guards guard the
// CONVENTION. Different axes.
//
// What the guards buy is therefore not earlier detection — a backtick in a
// template already breaks the pair test — but two things the pair test cannot
// give: a failure that names its own cause (the pair test reports a byte-length
// divergence, which does not mention backticks), and enforcement of the
// convention itself against the concatenation workaround.
//
// Keeping the ban rather than adopting the concatenation form is a lead
// judgement, not a measurement: the constants hold whole shell scripts, and a
// concatenation makes the constant a value that must be evaluated to be read,
// which breaks the reader's ability to compare template and constant by eye.
// Recorded here so a later reader can revisit the trade-off rather than assume
// it was forced. Provenance: cards t1002, t1003 §11, t1005 §2.2 / §6, t1009.

// backtick is written as a double-quoted string on purpose: the character
// cannot appear inside a raw string literal, which is the very constraint these
// guards enforce.
const backtick = "`"

// backtickConstraintAnchor is the stable phrase the in-file constraint comment
// opens with in every guarded template. The guard asserts the anchor rather
// than the whole comment so rewording stays possible; changing the anchor is a
// deliberate contract change that updates this constant too.
const backtickConstraintAnchor = "No backticks anywhere in this file"

// guardedHookTemplate pairs a hook template with the Go constant that must hold
// it verbatim as a single raw string literal.
type guardedHookTemplate struct {
	name         string
	templatePath []string // path segments relative to the project root
	goFile       string   // path relative to this package directory
	constName    string
}

func guardedHookTemplates() []guardedHookTemplate {
	return []guardedHookTemplate{
		{
			name:         "pre-push",
			templatePath: []string{"internal", "template", "templates", ".git_hooks", "pre-push"},
			goFile:       "hook_install.go",
			constName:    "prePushHookContent",
		},
		{
			name:         "pre-commit",
			templatePath: []string{"internal", "template", "templates", ".git_hooks", "pre-commit"},
			goFile:       "hook_install_precommit.go",
			constName:    "preCommitHookContent",
		},
	}
}

// projectRootForGuard walks up from the working directory to the go.mod marker.
// A tree without go.mod is not a source checkout (tarball test environments),
// and the guards skip there. A tree WITH go.mod but without the template is a
// different matter and fails below — in a real source checkout a missing
// template is itself the defect, and a guard that skips over it reports green
// for a tree it never inspected.
func projectRootForGuard(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := wd
	for {
		if _, statErr := os.Stat(filepath.Join(root, "go.mod")); statErr == nil {
			return root
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Skip("project root not found (go.mod missing) — skipping in isolated test environments")
		}
		root = parent
	}
}

func readGuardedTemplate(t *testing.T, root string, tmpl guardedHookTemplate) string {
	t.Helper()

	path := filepath.Join(append([]string{root}, tmpl.templatePath...)...)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: cannot read the guarded template at %s: %v\n"+
			"This tree has a go.mod, so it is a source checkout and the template is expected to exist.",
			tmpl.name, path, err)
	}
	return string(body)
}

// TestHookTemplatesCarryNoBacktick is the readable-failure half of the guard.
// When a backtick reaches a template, the pair test also breaks — but it breaks
// on a byte-length divergence that says nothing about backticks, and the
// obvious repair (copying the bytes into the constant) then breaks compilation
// in a different file. This failure names the cause once.
func TestHookTemplatesCarryNoBacktick(t *testing.T) {
	t.Parallel()

	root := projectRootForGuard(t)

	for _, tmpl := range guardedHookTemplates() {
		t.Run(tmpl.name, func(t *testing.T) {
			t.Parallel()

			body := readGuardedTemplate(t, root, tmpl)
			if !strings.Contains(body, backtick) {
				return
			}

			var lines []string
			for i, line := range strings.Split(body, "\n") {
				if strings.Contains(line, backtick) {
					lines = append(lines, strings.TrimSpace(line)+"  (line "+itoa(i+1)+")")
				}
			}
			t.Fatalf(
				"%s: the template contains a backtick, which it must never do.\n"+
					"  %s in %s holds this file verbatim as a Go raw string literal,\n"+
					"  and a raw string literal cannot contain a backtick.\n"+
					"  offending lines:\n    %s\n"+
					"  Remove the backtick. Do NOT repair this by rewriting the constant as a\n"+
					"  concatenation — that would satisfy the byte-identity pair test while\n"+
					"  silently retiring this convention, which is what %s also guards.",
				tmpl.name, tmpl.constName, tmpl.goFile,
				strings.Join(lines, "\n    "),
				"TestHookContentConstantsAreSingleRawStringLiterals",
			)
		})
	}
}

// TestHookContentConstantsAreSingleRawStringLiterals is the convention-
// enforcement half, and the reason this guard exists at all. The byte-identity
// pair tests compare the constant's VALUE, so they stay green under the
// concatenation workaround; only the declaration's SHAPE distinguishes the two,
// and only an AST read can see it.
func TestHookContentConstantsAreSingleRawStringLiterals(t *testing.T) {
	t.Parallel()

	for _, tmpl := range guardedHookTemplates() {
		t.Run(tmpl.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tmpl.goFile, nil, 0)
			if err != nil {
				t.Fatalf("%s: parse %s: %v", tmpl.name, tmpl.goFile, err)
			}

			value, found := constValueExpr(file, tmpl.constName)
			if !found {
				t.Fatalf("%s: constant %s not found in %s — the guard names a constant that no longer exists",
					tmpl.name, tmpl.constName, tmpl.goFile)
			}

			lit, ok := value.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || !strings.HasPrefix(lit.Value, backtick) {
				t.Fatalf(
					"%s: %s in %s must be declared as a SINGLE raw string literal.\n"+
						"  found instead: %T at %s\n"+
						"  A concatenation such as  const c = `head ` + \"`\" + ` tail`  compiles and\n"+
						"  produces the same value, so the byte-identity pair test stays green — but it\n"+
						"  retires the no-backtick convention for the matching template, silently.\n"+
						"  Keeping the single raw literal is what makes template and constant\n"+
						"  comparable by eye. If the ban is genuinely to be lifted, that is a\n"+
						"  deliberate decision that changes this guard, not a repair that routes\n"+
						"  around it.",
					tmpl.name, tmpl.constName, tmpl.goFile, value, fset.Position(value.Pos()),
				)
			}
		})
	}
}

// TestHookTemplatesCarryBacktickConstraintComment keeps the in-file explanation
// alive. The comment is what an editor sees at edit time, before CI runs; the
// guards above only speak once a build happens. Deleting the comment from both
// copies leaves them byte-identical, so no other check notices.
func TestHookTemplatesCarryBacktickConstraintComment(t *testing.T) {
	t.Parallel()

	root := projectRootForGuard(t)

	for _, tmpl := range guardedHookTemplates() {
		t.Run(tmpl.name, func(t *testing.T) {
			t.Parallel()

			body := readGuardedTemplate(t, root, tmpl)
			if strings.Contains(body, backtickConstraintAnchor) {
				return
			}
			t.Fatalf(
				"%s: the template no longer carries its backtick-constraint comment.\n"+
					"  expected a comment opening with: %q\n"+
					"  That comment is the edit-time signal; the guards in this file only speak\n"+
					"  at build time, and deleting the comment from both the template and its\n"+
					"  twin leaves them byte-identical, so no other check notices.\n"+
					"  Restore it, or change %s deliberately if the wording is being revised.",
				tmpl.name, backtickConstraintAnchor, "backtickConstraintAnchor",
			)
		})
	}
}

// constValueExpr returns the single value expression of a top-level constant.
func constValueExpr(file *ast.File, name string) (ast.Expr, bool) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
				continue
			}
			if vs.Names[0].Name == name {
				return vs.Values[0], true
			}
		}
	}
	return nil, false
}
