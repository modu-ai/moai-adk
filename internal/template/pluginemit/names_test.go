// names_test.go — AC-004 (b): the generator holds no component name.
//
// The scan is defined so that correct code passes it: it reads the string
// literals (interpreted values, found with the Go parser) of the non-test
// sources of this package, splits each on '/' and '\', and reports a token that
// equals — exactly and case-sensitively — a skill, agent or command name of the
// real template tree. Comments, identifiers and directives are not scanned.
// The plugin identifier "moai" is exempt: the generator must write it, and it
// is also the name of the catalog's first skill.
package pluginemit_test

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

	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/template/pluginemit"
)

// componentNames is the set N: every catalog entry name (skills and agents,
// every tier) and every command stem of the template tree, less the exempt
// plugin identifier.
func componentNames(t *testing.T) map[string]bool {
	t.Helper()
	cat, err := template.LoadCatalog(os.DirFS(rawTemplateDir))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, e := range cat.AllEntries() {
		names[e.Name] = true
	}
	cmds, err := os.ReadDir(filepath.Join(rawTemplateDir, "templates", ".claude", "commands", "moai"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cmds {
		names[strings.TrimSuffix(strings.TrimSuffix(c.Name(), ".tmpl"), ".md")] = true
	}
	delete(names, pluginemit.PluginName)
	return names
}

// literalsOf returns the interpreted string literals of one parsed file.
func literalsOf(f *ast.File) []string {
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if v, err := strconv.Unquote(lit.Value); err == nil {
			out = append(out, v)
		}
		return true
	})
	return out
}

// nameHits returns the literals that contain a whole token equal to a name.
func nameHits(literals []string, names map[string]bool) []string {
	var hits []string
	for _, lit := range literals {
		tokens := strings.FieldsFunc(lit, func(r rune) bool { return r == '/' || r == '\\' })
		for _, tok := range tokens {
			if names[tok] {
				hits = append(hits, tok)
			}
		}
	}
	return hits
}

func parseSource(t *testing.T, src string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "control.go", src, 0)
	if err != nil {
		t.Fatalf("control source does not parse: %v", err)
	}
	return f
}

func TestGeneratorHoldsNoComponentNames(t *testing.T) {
	names := componentNames(t)
	if len(names) == 0 {
		t.Fatal("the set of component names is empty; an empty sweep asserts nothing")
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)

	// Positive control: the scan finds a name in a path-segment literal and in a
	// lone literal, and a hard-coded mutant of the generator turns it red.
	pathForm := parseSource(t, "package x\nvar p = \"skills/"+sorted[0]+"/SKILL.md\"\n")
	if got := nameHits(literalsOf(pathForm), names); len(got) != 1 {
		t.Fatalf("positive control (path-segment form) found %d hits, want 1", len(got))
	}
	loneForm := parseSource(t, "package x\nvar n = \""+sorted[len(sorted)-1]+"\"\n")
	if got := nameHits(literalsOf(loneForm), names); len(got) != 1 {
		t.Fatalf("positive control (lone literal) found %d hits, want 1", len(got))
	}
	sentence := parseSource(t, "package x\nvar e = \"cannot "+sorted[0]+" the file\"\n")
	if got := nameHits(literalsOf(sentence), names); len(got) != 0 {
		t.Fatalf("a sentence containing a name is not a hit, scan reported %v", got)
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files, literals := 0, 0
	var hits []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", e.Name(), err)
		}
		files++
		lits := literalsOf(f)
		literals += len(lits)
		for _, h := range nameHits(lits, names) {
			hits = append(hits, e.Name()+": "+strconv.Quote(h))
		}
	}
	if files == 0 || literals == 0 {
		t.Fatalf("swept %d files and %d literals; an empty sweep asserts nothing", files, literals)
	}
	if len(hits) > 0 {
		t.Errorf("the generator holds %d component-name literal(s):\n%s", len(hits), strings.Join(hits, "\n"))
	}
}
