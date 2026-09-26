package hook

// SPEC-HOOK-MATCHER-POWERSHELL-001 (card t1224) — source guards.
//
//   - REQ-HMP-005/013(i): a shell-tool decision goes through IsShellTool. A
//     string literal equal to "bash" (any letter case) in a non-test Go file of
//     this package or in internal/cli/hook.go, outside the predicate file, is a
//     Bash-only branch that PowerShell calls fall through.
//   - REQ-HMP-013(ii): a tool_name match naming Bash in a pre-tool wrapper must
//     also name PowerShell, unless it is on the exclusion list with a reason.
//   - REQ-HMP-014 / AC-HMP-011: every TestHMP* function isolates the MoAI home
//     and does not run in parallel.
//
// Every list is closed-world: an entry that no longer matches a hit fails.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// hmpPredicateFile is the one Go file allowed to carry the shell-tool names.
const hmpPredicateFile = "internal/hook/shell_tool.go"

// hmpGoLiteralExclusions lists "<file>:<literal>" hits allowed outside the
// predicate file, with the reason. Empty: every Bash branch uses IsShellTool.
var hmpGoLiteralExclusions = map[string]string{}

// hmpWrapperExclusions lists wrapper lines that match tool_name "Bash" without
// PowerShell on purpose, keyed "<file>|<line substring>", with the reason.
var hmpWrapperExclusions = map[string]string{
	"internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl|" + `grep -q '"tool_name"[[:space:]]*:[[:space:]]*"Bash"'`: "D6: the Risk-Amplifier subcommand count is Bash-only; backtick and $( mean something else in PowerShell",
	".claude/hooks/moai/handle-pre-tool.sh|" + `grep -q '"tool_name"[[:space:]]*:[[:space:]]*"Bash"'`:                                  "D6: the Risk-Amplifier subcommand count is Bash-only; backtick and $( mean something else in PowerShell",
}

// hmpIsolateHome points the MoAI home and the project dir at per-test
// directories (REQ-HMP-014). It never calls t.Parallel.
func hmpIsolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
}

// hmpRepoRoot walks up to the directory containing go.mod.
func hmpRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found")
		}
		dir = parent
	}
}

// hmpBashLiterals returns every string literal in src equal to "bash" in any
// letter case, as "<name>:<line>:<literal>". Comments are not literals, so they
// are excluded by construction.
func hmpBashLiterals(t *testing.T, name string, src []byte) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	var hits []string
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		v, err := strconv.Unquote(lit.Value)
		if err != nil || !strings.EqualFold(v, "bash") {
			return true
		}
		hits = append(hits, name+":"+strconv.Itoa(fset.Position(lit.Pos()).Line)+":"+v)
		return true
	})
	return hits
}

// hmpGoSources lists the non-test Go files the guard scans, repo-relative.
func hmpGoSources(t *testing.T, root string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, "internal", "hook", "*.go"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	var out []string
	for _, m := range matches {
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		rel, _ := filepath.Rel(root, m)
		out = append(out, filepath.ToSlash(rel))
	}
	out = append(out, "internal/cli/hook.go")
	sort.Strings(out)
	return out
}

// hmpGoViolations applies the predicate-file rule and the exclusion list to
// the hits and returns one message per violation, plus stale exclusions.
func hmpGoViolations(hits []string, exclusions map[string]string) []string {
	used := map[string]bool{}
	var out []string
	for _, h := range hits {
		parts := strings.SplitN(h, ":", 3)
		file, lit := parts[0], parts[2]
		if file == hmpPredicateFile {
			continue
		}
		key := file + ":" + lit
		if _, ok := exclusions[key]; ok {
			used[key] = true
			continue
		}
		out = append(out, "shell-tool literal outside "+hmpPredicateFile+": "+h)
	}
	for key := range exclusions {
		if !used[key] {
			out = append(out, "stale Go literal exclusion: "+key)
		}
	}
	sort.Strings(out)
	return out
}

// TestHMPSourceGuardGoLiterals pins REQ-HMP-005 / REQ-HMP-013(i).
func TestHMPSourceGuardGoLiterals(t *testing.T) {
	hmpIsolateHome(t)
	root := hmpRepoRoot(t)
	var hits []string
	for _, rel := range hmpGoSources(t, root) {
		src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		hits = append(hits, hmpBashLiterals(t, rel, src)...)
	}
	t.Logf("shell-tool literal hits: %d", len(hits))
	for _, h := range hits {
		t.Logf("  %s", h)
	}
	for _, v := range hmpGoViolations(hits, hmpGoLiteralExclusions) {
		t.Error(v)
	}
}

// TestHMPSourceGuardPositiveControls proves the scanner is not blind.
func TestHMPSourceGuardPositiveControls(t *testing.T) {
	hmpIsolateHome(t)
	fixtures := map[string]string{
		"const":     "package x\nconst x = \"Bash\"\nfunc f(n string) bool { return n == x }\n",
		"equalfold": "package x\nimport \"strings\"\nfunc f(name string) bool { return strings.EqualFold(name, \"bash\") }\n",
	}
	for name, src := range fixtures {
		t.Run(name, func(t *testing.T) {
			hits := hmpBashLiterals(t, "internal/hook/fixture.go", []byte(src))
			if got := hmpGoViolations(hits, nil); len(got) != 1 {
				t.Errorf("fixture %q: want exactly 1 violation, got %q", name, got)
			}
		})
	}
	t.Run("comment is not a literal", func(t *testing.T) {
		hits := hmpBashLiterals(t, "internal/hook/fixture.go", []byte("package x\n// tool_name \"Bash\"\n"))
		if len(hits) != 0 {
			t.Errorf("comment counted as a literal: %q", hits)
		}
	})
	t.Run("stale exclusion", func(t *testing.T) {
		got := hmpGoViolations(nil, map[string]string{"internal/hook/gone.go:Bash": "x"})
		if len(got) != 1 || !strings.Contains(got[0], "stale Go literal exclusion") {
			t.Errorf("stale exclusion not reported: %q", got)
		}
	})
}

// hmpWrapperFiles lists the pre-tool wrappers the guard scans.
var hmpWrapperFiles = []string{
	"internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl",
	".claude/hooks/moai/handle-pre-tool.sh",
}

// hmpWrapperViolations flags each line that matches tool_name and names Bash
// without PowerShell, applying the exclusion list and reporting stale entries.
func hmpWrapperViolations(files map[string]string, exclusions map[string]string) []string {
	used := map[string]bool{}
	var out []string
	for file, body := range files {
		for i, line := range strings.Split(body, "\n") {
			if !strings.Contains(line, "tool_name") || !strings.Contains(line, `"Bash"`) || strings.Contains(line, "PowerShell") {
				continue
			}
			excluded := false
			for key := range exclusions {
				f, sub, _ := strings.Cut(key, "|")
				if f == file && strings.Contains(line, sub) {
					used[key] = true
					excluded = true
				}
			}
			if !excluded {
				out = append(out, file+":"+strconv.Itoa(i+1)+": tool_name match names Bash without PowerShell")
			}
		}
	}
	for key := range exclusions {
		f, _, _ := strings.Cut(key, "|")
		if _, scanned := files[f]; scanned && !used[key] {
			out = append(out, "stale wrapper exclusion: "+key)
		}
	}
	sort.Strings(out)
	return out
}

// TestHMPSourceGuardWrappers pins REQ-HMP-013(ii).
func TestHMPSourceGuardWrappers(t *testing.T) {
	hmpIsolateHome(t)
	root := hmpRepoRoot(t)
	files := map[string]string{}
	for _, rel := range hmpWrapperFiles {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Logf("%s not present: %v", rel, err)
			continue
		}
		files[rel] = string(data)
	}
	if len(files) == 0 {
		t.Fatal("no wrapper scanned — the guard would pass vacuously")
	}
	for _, v := range hmpWrapperViolations(files, hmpWrapperExclusions) {
		t.Error(v)
	}
	t.Run("positive control", func(t *testing.T) {
		got := hmpWrapperViolations(map[string]string{"x.sh": `grep -q '"tool_name":"Bash"'`}, nil)
		if len(got) != 1 {
			t.Errorf("unexcluded Bash-only match not flagged: %q", got)
		}
		if got := hmpWrapperViolations(map[string]string{"x.sh": `grep -qE '"tool_name":"(Bash|PowerShell)"'`}, nil); len(got) != 0 {
			t.Errorf("a match naming both tools was flagged: %q", got)
		}
	})
}

// hmpIsolationHelpers are the helpers AC-HMP-011 accepts in place of a direct
// t.Setenv("MOAI_HOME", t.TempDir()) call.
var hmpIsolationHelpers = map[string]bool{"hmpIsolateHome": true}

// hmpIsolationViolations checks every TestHMP* function in src: it must call
// t.Setenv("MOAI_HOME", t.TempDir()) or a listed helper, and must not call
// t.Parallel. It returns the violations and how many functions it checked.
func hmpIsolationViolations(t *testing.T, name string, src []byte) (violations []string, checked int) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "TestHMP") {
			continue
		}
		checked++
		isolated, parallel := false, false
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				if hmpIsolationHelpers[fun.Name] {
					isolated = true
				}
			case *ast.SelectorExpr:
				switch fun.Sel.Name {
				case "Parallel":
					parallel = true
				case "Setenv":
					if len(call.Args) == 2 {
						key, _ := call.Args[0].(*ast.BasicLit)
						val, _ := call.Args[1].(*ast.CallExpr)
						if key != nil && key.Value == `"MOAI_HOME"` && val != nil {
							if sel, ok := val.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "TempDir" {
								isolated = true
							}
						}
					}
				}
			}
			return true
		})
		if !isolated {
			violations = append(violations, name+": "+fn.Name.Name+" does not isolate MOAI_HOME")
		}
		if parallel {
			violations = append(violations, name+": "+fn.Name.Name+" calls t.Parallel")
		}
	}
	return violations, checked
}

// hmpHelperBodies returns, per scanned file, whether each listed helper it
// defines sets MOAI_HOME to a t.TempDir() value.
func hmpHelperIsolates(t *testing.T, name string, src []byte) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	var bad []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || !hmpIsolationHelpers[fn.Name.Name] {
			continue
		}
		body := string(src[fset.Position(fn.Body.Pos()).Offset:fset.Position(fn.Body.End()).Offset])
		if !strings.Contains(body, `t.Setenv("MOAI_HOME", t.TempDir())`) {
			bad = append(bad, name+": helper "+fn.Name.Name+" does not set MOAI_HOME to t.TempDir()")
		}
	}
	return bad
}

// TestHMPTestIsolation pins AC-HMP-011 over every test file in the three
// packages this card touches.
func TestHMPTestIsolation(t *testing.T) {
	hmpIsolateHome(t)
	root := hmpRepoRoot(t)
	total := 0
	for _, dir := range []string{"internal/hook", "internal/cli", "internal/template"} {
		files, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(dir), "*_test.go"))
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		for _, file := range files {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			rel, _ := filepath.Rel(root, file)
			violations, checked := hmpIsolationViolations(t, filepath.ToSlash(rel), src)
			total += checked
			for _, v := range violations {
				t.Error(v)
			}
			for _, v := range hmpHelperIsolates(t, filepath.ToSlash(rel), src) {
				t.Error(v)
			}
		}
	}
	t.Logf("TestHMP functions checked: %d", total)
	if total == 0 {
		t.Fatal("no TestHMP function found — an empty sweep is a failure, not a pass")
	}
	t.Run("positive control", func(t *testing.T) {
		src := "package x\nimport \"testing\"\nfunc TestHMPBad(t *testing.T) { t.Parallel() }\n"
		got, checked := hmpIsolationViolations(t, "fixture_test.go", []byte(src))
		if checked != 1 || len(got) != 2 {
			t.Errorf("want 2 violations on 1 function, got %d on %d: %q", len(got), checked, got)
		}
	})
}
