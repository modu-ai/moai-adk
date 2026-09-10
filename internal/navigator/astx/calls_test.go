//go:build cgo

package astx

import (
	"os"
	"path/filepath"
	"testing"
)

// writeCallFixture writes a Go source exercising call/import captures:
// two functions, a method, plain and selector calls, stdlib + internal imports.
func writeCallFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := `package demo

import (
	"fmt"
	"strings"
)

func Alpha() {
	fmt.Println("hi")
	Beta()
}

func Beta() string {
	return strings.ToUpper("x")
}

type T struct{}

func (t T) Render() string {
	return Beta()
}
`
	path := filepath.Join(dir, "demo.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// REQ-GF-013 substrate — call/import extraction on Go: callee names (plain
// and selector), function ranges with caller names, import modules.
func TestExtractCalls_Go(t *testing.T) {
	path := writeCallFixture(t)
	set, err := ExtractCalls("go", path)
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	if !set.Supported {
		t.Fatal("ExtractCalls(go) Supported=false, want true (grammar + call captures seeded)")
	}

	callees := map[string]bool{}
	for _, c := range set.Calls {
		callees[c.Callee] = true
	}
	for _, want := range []string{"Println", "Beta", "ToUpper"} {
		if !callees[want] {
			t.Errorf("callee %q not captured; calls=%v", want, set.Calls)
		}
	}

	fns := map[string]FuncRange{}
	for _, f := range set.Functions {
		fns[f.Name] = f
	}
	for _, want := range []string{"Alpha", "Beta", "Render"} {
		if _, ok := fns[want]; !ok {
			t.Errorf("function range %q missing; functions=%v", want, set.Functions)
		}
	}
	// Ranges must bracket their bodies: Alpha starts at its declaration line
	// and ends after its calls (enough for caller joins by containment).
	if f, ok := fns["Alpha"]; !ok || f.EndLine <= f.StartLine {
		t.Errorf("Alpha range invalid: %+v", f)
	}

	imports := map[string]bool{}
	for _, imp := range set.Imports {
		imports[imp.Module] = true
	}
	if !imports["fmt"] || !imports["strings"] {
		t.Errorf("imports fmt/strings not captured: %v", set.Imports)
	}
}

// TestExtractCalls_GoRawStringImport pins the raw-string import form (CR
// 3855002146): Go accepts raw string literals as import paths, and the
// @code.import capture must include them exactly as interpreted strings are.
func TestExtractCalls_GoRawStringImport(t *testing.T) {
	dir := t.TempDir()
	src := "package demo\n\nimport (\n\t\"fmt\"\n\t`strings`\n)\n\nfunc A() { fmt.Println(1) }\n"
	path := filepath.Join(dir, "raw.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := ExtractCalls("go", path)
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	if !set.Supported {
		t.Fatal("ExtractCalls(go) Supported=false, want true")
	}
	modules := map[string]bool{}
	for _, imp := range set.Imports {
		modules[imp.Module] = true
	}
	// fmt (interpreted) doubles as the extraction-ran control; strings is the
	// raw-string form under test.
	if !modules["fmt"] {
		t.Errorf("interpreted import fmt missing: %v", set.Imports)
	}
	if !modules["strings"] {
		t.Errorf("raw-string import strings not captured: %v", set.Imports)
	}
}

// Python capture parity (name-based grade).
func TestExtractCalls_Python(t *testing.T) {
	dir := t.TempDir()
	src := "import os\nfrom sys import path\n\ndef run():\n    os.getcwd()\n    helper()\n\ndef helper():\n    return 1\n"
	path := filepath.Join(dir, "m.py")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := ExtractCalls("python", path)
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	if !set.Supported {
		t.Fatal("ExtractCalls(python) Supported=false, want true")
	}
	callees := map[string]bool{}
	for _, c := range set.Calls {
		callees[c.Callee] = true
	}
	if !callees["getcwd"] || !callees["helper"] {
		t.Errorf("python callees missing: %v", set.Calls)
	}
	imports := map[string]bool{}
	for _, imp := range set.Imports {
		imports[imp.Module] = true
	}
	if !imports["os"] {
		t.Errorf("python import os missing: %v", set.Imports)
	}
	fns := map[string]bool{}
	for _, f := range set.Functions {
		fns[f.Name] = true
	}
	if !fns["run"] || !fns["helper"] {
		t.Errorf("python function ranges missing: %v", set.Functions)
	}
}

// Unseeded call captures grade "none": Supported=false or empty capture set —
// the grade matrix records honesty, the extractor never fabricates.
func TestExtractCalls_NoneGradeLanguages(t *testing.T) {
	// ruby has a grammar but (this milestone) no call captures seeded.
	dir := t.TempDir()
	src := "def run\n  helper\nend\n"
	path := filepath.Join(dir, "m.rb")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := ExtractCalls("ruby", path)
	if err != nil {
		t.Fatalf("ExtractCalls(ruby): %v", err)
	}
	// Either unsupported for calls, or supported with zero captures — both
	// are honest "none" grades; fabricating call sites is the failure mode.
	if set.Supported && len(set.Calls) > 0 {
		t.Errorf("ruby call captures unexpectedly present: %v", set.Calls)
	}
}

// Scaffolded languages stay unsupported for calls too.
func TestExtractCalls_Scaffolded(t *testing.T) {
	set, err := ExtractCalls("flutter", "/nonexistent.dart")
	if err != nil {
		t.Fatalf("ExtractCalls(flutter): %v", err)
	}
	if set.Supported {
		t.Error("flutter must be Supported=false (scaffolded)")
	}
}

// callPolyglotCases seeds one minimal source per name-based-graded language;
// each row must compile the extended query (Supported=true) and capture at
// least the named callee — the .scm syntax is exercised against the real
// grammar, not just string-checked.
var callPolyglotCases = []struct {
	lang   string
	src    string
	callee string
}{
	{"go", "package m\n\nimport \"fmt\"\n\nfunc A() { fmt.Println(1) }\n", "Println"},
	{"python", "import os\n\ndef a():\n    os.getcwd()\n", "getcwd"},
	{"javascript", "import x from 'y';\nfunction a() { x.go(); }\n", "go"},
	{"typescript", "import x from 'y';\nfunction a(): void { x.go(); }\n", "go"},
	{"java", "import java.util.List;\nclass C { void a() { System.exit(0); } }\n", "exit"},
	{"rust", "use std::io;\nfn a() { io::flush(); }\n", "flush"},
}

func TestExtractCalls_NameBasedLanguagesCompile(t *testing.T) {
	for _, tc := range callPolyglotCases {
		t.Run(tc.lang, func(t *testing.T) {
			dir := t.TempDir()
			ext := map[string]string{
				"go": ".go", "python": ".py", "javascript": ".js",
				"typescript": ".ts", "java": ".java", "rust": ".rs",
			}[tc.lang]
			path := filepath.Join(dir, "m"+ext)
			if err := os.WriteFile(path, []byte(tc.src), 0o644); err != nil {
				t.Fatal(err)
			}
			set, err := ExtractCalls(tc.lang, path)
			if err != nil {
				t.Fatalf("ExtractCalls(%s): %v", tc.lang, err)
			}
			if !set.Supported {
				t.Fatalf("ExtractCalls(%s) Supported=false — extended .scm failed to compile or parse", tc.lang)
			}
			found := false
			for _, c := range set.Calls {
				if c.Callee == tc.callee {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: callee %q not captured; calls=%v", tc.lang, tc.callee, set.Calls)
			}
		})
	}
}

// GradeMatrix covers all 16 registered languages with no empty cells
// (AC-GF-019 precondition at the astx layer): every language carries
// full | name-based | none.
func TestGradeMatrix_NoEmptyCells(t *testing.T) {
	langs := SupportedLanguages()
	if len(langs) != 16 {
		t.Fatalf("registered languages = %d, want 16", len(langs))
	}
	for _, lang := range langs {
		g := GradeFor(lang)
		switch g {
		case GradeFull, GradeNameBased, GradeNone:
		default:
			t.Errorf("language %s grade = %q — empty/invalid cell", lang, g)
		}
	}
	// Grounded expectations: seeded call captures are name-based this
	// milestone; scaffolded languages are none; no language claims full yet.
	for _, lang := range []string{"go", "python", "javascript", "typescript", "java", "rust"} {
		if g := GradeFor(lang); g != GradeNameBased {
			t.Errorf("%s grade = %q, want name-based (captures seeded, scope-aware not claimed)", lang, g)
		}
	}
	for _, lang := range []string{"r", "flutter", "ruby", "kotlin"} {
		if g := GradeFor(lang); g != GradeNone {
			t.Errorf("%s grade = %q, want none (no call captures seeded)", lang, g)
		}
	}
}
