package cli

// retired_word_identifiers_m7_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M7 (card
// t1399), the identifier half of AC-018 (REQ-017): outside internal/web, no Go
// identifier carries the retired mode word in any letter case. The four test
// names that acceptance commands select by name are the only identifiers that
// keep it (the bare package name was an allowed identifier until M8 renamed the
// package; a bare use of it now counts like any other). internal/web moves as
// one unit at M9 and is not scanned here.
//
// The scan is syntactic (go/parser identifiers), so a comment, a string literal,
// or a marker value (`MOAI_KANBAN_ID`, frozen by AC-016) never trips it. The
// re-runnable program probe/rename/m7_rename.go converts an offender.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// pinnedRetiredWordTestNames are the test names that acceptance commands select
// by name (AC-011, AC-014, AC-017, AC-019): they stay, with the word, so the
// commands keep selecting exactly the tests they were written against.
var pinnedRetiredWordTestNames = map[string]bool{
	"TestKanbanEntryRefused":                  true,
	"TestSessionStartEmitsNoKanbanNotice":     true,
	"TestPreexistingKanbanArtifactsTolerated": true,
	"TestLegacyKanbanRouteRedirects":          true,
}

// retiredWordIdentifiers returns the identifiers of one Go source that carry the
// word, other than the package clause and the pinned test names.
func retiredWordIdentifiers(t *testing.T, filename string, src any) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id == f.Name {
			return true
		}
		if strings.Contains(strings.ToLower(id.Name), "kanban") && !pinnedRetiredWordTestNames[id.Name] {
			out = append(out, filename+":"+strconv.Itoa(fset.Position(id.Pos()).Line)+" "+id.Name)
		}
		return true
	})
	return out
}

// TestRetiredWordIdentifierScanHasTeeth is the positive control: the scan finds
// an old-style identifier and a stale bare package qualifier, and lets the
// renamed qualifier and the pinned test name through, so a zero result on the
// tree below means something. The old-style spellings are assembled from parts
// so the rename programs, which rewrite spelled identifiers in strings, leave
// the fixture alone.
func TestRetiredWordIdentifierScanHasTeeth(t *testing.T) {
	oldName := "prepare" + "Kan" + "ban" + "Settings"
	src := "package x\n\n" +
		"import \"github.com/modu-ai/moai-adk/internal/factory\"\n\n" +
		"func " + oldName + "() {}\n\n" +
		"var _ = factory.Record{}\n\n" +
		"var _ = kan" + "ban.Record{}\n\n" +
		"func Test" + "Kan" + "banEntryRefused() {}\n"
	got := retiredWordIdentifiers(t, "control.go", src)
	if len(got) != 2 || !strings.HasSuffix(got[0], " "+oldName) || !strings.HasSuffix(got[1], " kan"+"ban") {
		t.Fatalf("the scan must report exactly %s and the stale bare qualifier (and let the renamed qualifier and the pinned test name through), got %v", oldName, got)
	}
}

func TestNoRetiredWordIdentifiersOutsideWeb(t *testing.T) {
	var offenders []string
	for _, root := range []string{"../../internal", "../../cmd"} {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				switch {
				case d.Name() == "testdata", d.Name() == "node_modules":
					return filepath.SkipDir
				case filepath.ToSlash(p) == "../../internal/web", filepath.ToSlash(p) == "../../internal/template/templates":
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".go") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil || !strings.Contains(strings.ToLower(string(b)), "kanban") {
				return nil
			}
			offenders = append(offenders, retiredWordIdentifiers(t, p, b)...)
			return nil
		})
	}
	sort.Strings(offenders)
	if len(offenders) > 0 {
		shown := offenders
		if len(shown) > 25 {
			shown = shown[:25]
		}
		t.Fatalf("%d Go identifier(s) outside internal/web still carry the retired mode word (run probe/rename/m7_rename.go):\n%s", len(offenders), strings.Join(shown, "\n"))
	}
}
