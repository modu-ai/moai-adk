//go:build ignore

// m9_rename.go is the re-runnable M9 rename of SPEC-LAUNCHER-ENTRY-FLAGS-001 (card t1399): the web console
// (internal/web) stops carrying the retired mode word. It is the mechanical half of milestone M9, written as a
// program so that the lane can absorb the integration branch and run it again to convert what other lanes added
// meanwhile (a new test that calls an old helper, a new i18n key, a new GET of the old route).
//
//	go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m9_rename.go -root <git tree> [-dry-run] [-v]
//
// What it does, in this order (rules adapted from probe/probe.go stageM9, re-measured on the live tree):
//
//  1. Go identifiers (syntax-tree based): every identifier of a non-generated Go file of internal/web that
//     carries the word in any letter case is renamed by replacing the word with "factory" (KanbanVM becomes
//     FactoryVM, buildKanban becomes buildFactory), except the two test names that acceptance commands select
//     by name. A new name that is already declared at package level is a collision and nothing is written.
//  2. String literals and comments of those files, and the text of the .templ sources and the three assets
//     (assets/app.js, assets/i18n.js, assets/console.css): every standalone spelling of the word (a token that is exactly kanban or
//     Kanban, or a camel-case token that contains it, not joined to its neighbours by a hyphen or an underscore)
//     becomes factory/Factory. That covers the route /kanban, the data-live area, the SSE event key, the
//     "kanban.*" i18n key family, the icon id, the shell area id, and "GET /kanban" in messages. In comments
//     the Korean spelling is rewritten too. A hyphenated compound (kanban-board, kanban-dispatch.md,
//     SPEC-KANBAN-...) is left; a literal "kanban" that follows a "state" argument of a path join (the legacy
//     state directory, a frozen data name) is left. The legacy-route file, its test, and the pre-existing
//     artifacts test keep the word in their strings and comments: their subject is the retired name.
//  3. The two nav/screen i18n title values are replaced per locale (Kanban/factory, the Korean, Japanese, and
//     Chinese names for the board become the loanwords the same files already use for "factory"): "nav.kanban"
//     and "screen.kanban" become "nav.factory" and "screen.factory". The two Chinese strings of the SPEC
//     board page that use the Chinese word for board (board.title, board.subtitle; the English values are
//     Board and dashboard) take the Chinese words for panel and dashboard, because the word sweep matches them.
//  4. The chain session board: the panel block of screens.templ, the i18n keys only it used (chain.title,
//     chain.open, chain.stopped, chain.sessionNotStarted, kanban.viewA, kanban.card, kanban.noSession,
//     kanban.noStart, kanban.model), the view model only it read (ChainVM, RoleVM, ChainRoles, buildChain,
//     chainRoleRecords, chainCardID, the Chain field of OverviewVM, the CardID, IdleRole, and Roles fields of
//     the screen model, the chain parameter of buildAttention and the chain-stopped attention row, decision
//     Q24 written to its smallest-footprint reading), and every test declaration that references them (one
//     function, TestBuildAttentionOrderAndCap, is rewritten without the chain row instead of removed because
//     it also pins the ordering and the cap of the must-fix rows). Declarations whose only readers went with
//     them (roleOf, readTelemetry, the two legacy-leader constants, the telemetry test helper, the Role field
//     of AttentionVM) go too, when nothing in the package reads them any more. A Go file left without
//     declarations is removed; imports the removals leave unused are dropped.
//  5. A file of internal/web whose name carries the word is renamed (git mv).
//  6. The legacy route: internal/web/legacy_routes.go (created when absent) holds the only remaining spelling
//     of the old route, a GET redirect to /factory and nothing else, and app.go registers it.
//  7. The generated *_templ.go files are regenerated the way `make templ-generate` does:
//     go run github.com/a-h/templ/cmd/templ generate -path ./internal/web
//
// It never touches: files outside internal/web, the six marker string values (MOAI_KANBAN_* tokens contain
// underscores and are left), the legacy state-directory name, hyphenated compounds, the generated files by
// text (they are regenerated), testdata, or the gitignored directories.
//
// Safety: it refuses (exit 2) when the tree has tracked modifications; it refuses (exit 1, nothing written)
// when a new identifier is already declared, when a source does not parse, when a rewrite does not gofmt, or
// when a structure it removes is not in the shape it expects. It is idempotent: a second run on a converted
// tree changes nothing, prints zeros, and exits 0. After the summary it prints, informationally, what still
// carries the word (the hand-edit list: prose the rules cannot judge, a citation of another SPEC's id).
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const webDir = "internal/web"

// pinnedTests are the test names that acceptance commands select by name; they keep the word.
var pinnedTests = map[string]bool{
	"TestPreexistingKanbanArtifactsTolerated": true,
	"TestLegacyKanbanRouteRedirects":          true,
}

// legacyFiles keep the word in their strings and comments: their subject is the retired name.
var legacyFiles = map[string]bool{
	"legacy_routes.go":                         true,
	"legacy_routes_test.go":                    true,
	"preexisting_kanban_artifacts_m1_test.go":  true,
	"preexisting_factory_artifacts_m1_test.go": true,
}

// removedNames are the chain view-model declarations the panel's removal takes with it.
var removedNames = []string{"ChainVM", "RoleVM", "ChainRoles", "buildChain", "chainRoleRecords", "chainCardID"}

var (
	identReplacer = strings.NewReplacer("Kanban", "Factory", "kanban", "factory", "KANBAN", "FACTORY")
	tokReplacer   = strings.NewReplacer("Kanban", "Factory", "kanban", "factory")

	i18nRemove      = regexp.MustCompile(`^\s*"(chain\.(title|open|stopped|sessionNotStarted)|kanban\.(viewA|card|noSession|noStart|model))":`)
	i18nTitleKey    = regexp.MustCompile(`^(\s*"(?:nav|screen)\.kanban":\s*")([^"]*)(".*)$`)
	zhBoardTitle    = regexp.MustCompile(`^(\s*"board\.title":\s*")看板(".*)$`)
	zhBoardSubtitle = regexp.MustCompile(`^(\s*"board\.subtitle":\s*"[^"]*?)看板([^"]*".*)$`)
	// the board's name per locale becomes the word the same locale's file already uses for "factory"
	titleValue = map[string]string{"Kanban": "Factory", "칸반": "팩토리", "かんばん": "ファクトリー", "カンバン": "ファクトリー", "看板": "工厂"}
)

const legacyRoutesSource = `package web

import "net/http"

// registerLegacyRoutes keeps the retired screen path as a redirect and nothing else: an old bookmark or a
// browser history entry lands on the screen that replaced it. The path has no handler logic and renders no
// page; the redirect is the whole of what survives of the old name.
func registerLegacyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /kanban", func(w http.ResponseWriter, r *http.Request) {
		target := "/factory"
		if q := r.URL.RawQuery; q != "" {
			target += "?" + q
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}
`

const attentionTestSource = `// TestBuildAttentionOrderAndCap pins the attention list: MUST-FIX findings lead with their spec-targeted
// links, non-must severities are skipped, and the list caps so a catastrophic audit cannot drown the screen.
func TestBuildAttentionOrderAndCap(t *testing.T) {
	rows := []SpecRowVM{{ID: "SPEC-A-001"}, {ID: "SPEC-B-002"}}
	findings := map[string][]FindingVM{}
	for i := 0; i < 12; i++ {
		findings["SPEC-A-001"] = append(findings["SPEC-A-001"], FindingVM{Severity: "MUST-FIX", Message: "fix", File: "type"})
	}
	findings["SPEC-B-002"] = []FindingVM{{Severity: "SHOULD-FIX", Message: "minor", File: "type"}}

	got := buildAttention(rows, findings)
	if len(got) != maxOverviewRows {
		t.Fatalf("attention rows = %d, want capped at %d", len(got), maxOverviewRows)
	}
	if got[0].Href != "/specs?id=SPEC-A-001" {
		t.Errorf("must-fix rows lost their spec target: %+v", got[0])
	}
	for _, a := range got {
		if a.Text == "minor" {
			t.Errorf("a SHOULD-FIX finding reached the attention list: %+v", a)
		}
	}

	quiet := buildAttention(rows, map[string][]FindingVM{})
	if len(quiet) != 0 {
		t.Errorf("a clean audit produced attention rows: %+v", quiet)
	}
}
`

type edit struct {
	a, b int
	s    string
}

type stats struct {
	idents, literals, comments, templ, js, i18nKeys, panels, vmDecls, testDecls, importsDropped int
	identNames                                                                                  map[string]string
	filesRemoved, filesRenamed                                                                  []string
	legacy                                                                                      string
}

var (
	verbose bool
	st      = &stats{identNames: map[string]string{}}
)

func fatal(code int, format string, a ...any) {
	fmt.Fprintf(os.Stderr, "m9_rename: "+format+"\n", a...)
	os.Exit(code)
}

func note(kind, rel string, line int, what string) {
	if verbose {
		fmt.Printf("  %s %s:%d: %s\n", kind, rel, line, what)
	}
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func isTokByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// rewriteText replaces every standalone or camel-case spelling of the word in s. A token joined to a
// neighbour by a hyphen (kanban-board, SPEC-KANBAN-x) or containing an underscore is left.
func rewriteText(s string) (string, int) {
	if !strings.Contains(s, "kanban") && !strings.Contains(s, "Kanban") {
		return s, 0
	}
	var sb strings.Builder
	n := 0
	for i := 0; i < len(s); {
		if !isTokByte(s[i]) {
			sb.WriteByte(s[i])
			i++
			continue
		}
		j := i
		for j < len(s) && isTokByte(s[j]) {
			j++
		}
		tok := s[i:j]
		if strings.Contains(tok, "kanban") || strings.Contains(tok, "Kanban") {
			hyphen := (i > 0 && s[i-1] == '-') || (j < len(s) && s[j] == '-')
			if !hyphen && !strings.Contains(tok, "_") {
				nt := tokReplacer.Replace(tok)
				if nt != tok {
					n++
				}
				tok = nt
			}
		}
		sb.WriteString(tok)
		i = j
	}
	return sb.String(), n
}

// rewriteProse is rewriteText plus the Korean spelling, for comments.
func rewriteProse(s string) (string, int) {
	s, n := rewriteText(s)
	if strings.Contains(s, "칸반") {
		n += strings.Count(s, "칸반")
		s = strings.ReplaceAll(s, "칸반", "팩토리")
	}
	return s, n
}

func applyEdits(src []byte, es []edit) []byte {
	sort.SliceStable(es, func(i, j int) bool { return es[i].a > es[j].a })
	out := append([]byte(nil), src...)
	last := -1
	for _, e := range es {
		if e.a == last && e.b == e.a {
			continue
		}
		last = e.a
		out = append(out[:e.a], append([]byte(e.s), out[e.b:]...)...)
	}
	return out
}

func parseGo(rel string, src []byte) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, rel, src, parser.ParseComments)
	if err != nil {
		fatal(1, "parse %s: %v; nothing written", rel, err)
	}
	return fset, f
}

func formatGo(rel string, out []byte) []byte {
	fm, err := format.Source(out)
	if err != nil {
		dump := filepath.Join(os.TempDir(), "m9_unformatted_"+filepath.Base(rel))
		_ = os.WriteFile(dump, out, 0o644)
		fatal(1, "gofmt %s: %v; nothing written (the unformatted result is in %s)", rel, err, dump)
	}
	return fm
}

// lineSpan widens [a,b) to whole lines when only blanks surround it on its first and last line.
func lineSpan(src []byte, a, b int) (int, int) {
	ls := bytes.LastIndexByte(src[:a], '\n') + 1
	if len(bytes.TrimSpace(src[ls:a])) == 0 {
		a = ls
	}
	le := bytes.IndexByte(src[b:], '\n')
	if le < 0 {
		le = len(src) - b
	} else {
		le++
	}
	if len(bytes.TrimSpace(src[b:b+le])) == 0 {
		b += le
	}
	return a, b
}

// extendUpComments widens a deletion start over the comment lines that sit directly above it.
func extendUpComments(fset *token.FileSet, f *ast.File, src []byte, a int) int {
	for {
		line := fset.Position(token.Pos(fset.File(f.Pos()).Base() + a)).Line
		moved := false
		for _, cg := range f.Comments {
			if fset.Position(cg.End()).Line == line-1 {
				ca := fset.Position(cg.Pos()).Offset
				if len(bytes.TrimSpace(src[bytes.LastIndexByte(src[:ca], '\n')+1:ca])) == 0 {
					a = bytes.LastIndexByte(src[:ca], '\n') + 1
					moved = true
					break
				}
			}
		}
		if !moved {
			return a
		}
	}
}

func identsOf(n ast.Node) map[string]bool {
	m := map[string]bool{}
	ast.Inspect(n, func(x ast.Node) bool {
		if id, ok := x.(*ast.Ident); ok {
			m[id.Name] = true
		}
		return true
	})
	return m
}

func anyOf(m map[string]bool, names ...string) bool {
	for _, n := range names {
		if m[n] {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------- step 1 and 2: Go sources

func rewriteGo(rel string, src []byte) []byte {
	if !bytes.Contains(bytes.ToLower(src), []byte("kanban")) && !bytes.Contains(src, []byte("칸반")) {
		return src
	}
	fset, f := parseGo(rel, src)
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	exempt := legacyFiles[filepath.Base(rel)]
	var es []edit

	keep := map[token.Pos]bool{} // a literal "kanban" after a "state" argument: the legacy state directory
	imports := map[token.Pos]bool{}
	for _, im := range f.Imports {
		imports[im.Path.Pos()] = true
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			for i := 1; i < len(x.Args); i++ {
				l1, ok1 := x.Args[i-1].(*ast.BasicLit)
				l2, ok2 := x.Args[i].(*ast.BasicLit)
				if ok1 && ok2 && l1.Value == `"state"` && l2.Value == `"kanban"` {
					keep[l2.Pos()] = true
				}
			}
		case *ast.Ident:
			if strings.Contains(strings.ToLower(x.Name), "kanban") && !pinnedTests[x.Name] {
				nn := identReplacer.Replace(x.Name)
				if nn != x.Name {
					a := off(x.Pos())
					es = append(es, edit{a, a + len(x.Name), nn})
					st.idents++
					st.identNames[x.Name] = nn
				}
			}
		case *ast.BasicLit:
			if x.Kind == token.STRING && !exempt && !imports[x.Pos()] && !keep[x.Pos()] {
				if nv, n := rewriteText(x.Value); n > 0 {
					a := off(x.Pos())
					es = append(es, edit{a, a + len(x.Value), nv})
					st.literals += n
					note("literal", rel, fset.Position(x.Pos()).Line, strings.TrimSpace(x.Value)+" -> "+strings.TrimSpace(nv))
				}
			}
		}
		return true
	})
	if !exempt {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if nv, n := rewriteProse(c.Text); n > 0 {
					a := off(c.Slash)
					es = append(es, edit{a, a + len(c.Text), nv})
					st.comments += n
					note("comment", rel, fset.Position(c.Slash).Line, strings.TrimSpace(c.Text)+" -> "+strings.TrimSpace(nv))
				}
			}
		}
	}
	if len(es) == 0 {
		return src
	}
	return formatGo(rel, applyEdits(src, es))
}

// declaredNames lists the package-level names of one file.
func declaredNames(src []byte, rel string) []string {
	_, f := parseGo(rel, src)
	var out []string
	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.FuncDecl:
			if x.Recv == nil {
				out = append(out, x.Name.Name)
			}
		case *ast.GenDecl:
			for _, sp := range x.Specs {
				switch t := sp.(type) {
				case *ast.TypeSpec:
					out = append(out, t.Name.Name)
				case *ast.ValueSpec:
					for _, n := range t.Names {
						out = append(out, n.Name)
					}
				}
			}
		}
	}
	return out
}

// ---------------------------------------------------------------- step 4: the chain view model and its tests

// orphanNames are declarations whose only readers were the chain view model and the tests removed with it.
// They go too, but only when no reader is left in the package after the removal (a late reader keeps one).
var orphanNames = []string{"roleOf", "readTelemetry", "legacyLeaderRole", "legacyLeaderLabel", "writeTelemetry"}

// dropOrphans removes the orphanNames that nothing in the package reads any more. It returns the files that
// were left without declarations.
func dropOrphans(out map[string][]byte, rels []string) []string {
	total := map[string]int{}
	for _, rel := range rels {
		if out[rel] == nil {
			continue
		}
		_, f := parseGo(rel, out[rel])
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				total[id.Name]++
			}
			return true
		})
	}
	var emptied []string
	for _, rel := range rels {
		src := out[rel]
		if src == nil {
			continue
		}
		fset, f := parseGo(rel, src)
		declared := map[string]bool{}
		for _, n := range declaredNames(src, rel) {
			declared[n] = true
		}
		orphan := map[string]bool{}
		for _, n := range orphanNames {
			if declared[n] && total[n] == 1 {
				orphan[n] = true
			}
		}
		if len(orphan) == 0 {
			continue
		}
		es := removeDecls(fset, f, src, orphan, rel)
		res := dropUnusedImports(rel, src, applyEdits(src, es))
		var empty bool
		res, empty = formatOrEmpty(rel, res)
		out[rel] = res
		if empty {
			emptied = append(emptied, rel)
		}
	}
	return emptied
}

// removeDecls deletes top-level declarations by name (functions without a receiver, types, const and var
// specs; a block whose specs are all deleted goes whole). It returns the edits and the spans it deletes.
func removeDecls(fset *token.FileSet, f *ast.File, src []byte, names map[string]bool, rel string) []edit {
	var es []edit
	del := func(a, b int, what string, line int) {
		a, b = lineSpan(src, a, b)
		es = append(es, edit{a, b, ""})
		st.vmDecls++
		note("remove", rel, line, what)
	}
	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.FuncDecl:
			if x.Recv == nil && names[x.Name.Name] {
				a := fset.Position(x.Pos()).Offset
				if x.Doc != nil {
					a = fset.Position(x.Doc.Pos()).Offset
				}
				del(a, fset.Position(x.End()).Offset, "func "+x.Name.Name, fset.Position(x.Pos()).Line)
			}
		case *ast.GenDecl:
			var hit []ast.Spec
			for _, sp := range x.Specs {
				switch t := sp.(type) {
				case *ast.TypeSpec:
					if names[t.Name.Name] {
						hit = append(hit, sp)
					}
				case *ast.ValueSpec:
					for _, n := range t.Names {
						if names[n.Name] {
							hit = append(hit, sp)
							break
						}
					}
				}
			}
			if len(hit) == 0 {
				continue
			}
			if len(hit) == len(x.Specs) {
				a := fset.Position(x.Pos()).Offset
				if x.Doc != nil {
					a = fset.Position(x.Doc.Pos()).Offset
				}
				del(a, fset.Position(x.End()).Offset, "decl block", fset.Position(x.Pos()).Line)
				continue
			}
			for _, sp := range hit {
				a, b := fset.Position(sp.Pos()).Offset, fset.Position(sp.End()).Offset
				switch t := sp.(type) {
				case *ast.TypeSpec:
					if t.Doc != nil {
						a = fset.Position(t.Doc.Pos()).Offset
					}
				case *ast.ValueSpec:
					if t.Doc != nil {
						a = fset.Position(t.Doc.Pos()).Offset
					}
				}
				del(a, b, "spec", fset.Position(sp.Pos()).Line)
			}
		}
	}
	return es
}

// editViewModel applies the chain removal to viewmodel_ops.go: declarations, struct fields, composite-literal
// keys, the chain parameter and row of buildAttention, and the builder statements that feed the chain.
func editViewModel(rel string, src []byte) []byte {
	if !bytes.Contains(src, []byte("ChainVM")) && !bytes.Contains(src, []byte("buildChain")) {
		return src
	}
	fset, f := parseGo(rel, src)
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	names := map[string]bool{}
	for _, n := range removedNames {
		names[n] = true
	}
	es := removeDecls(fset, f, src, names, rel)
	deleted := func(a int) bool {
		for _, e := range es {
			if e.s == "" && a >= e.a && a < e.b {
				return true
			}
		}
		return false
	}
	cut := func(a, b int, what string, line int) {
		if deleted(a) {
			return
		}
		a, b = lineSpan(src, a, b)
		es = append(es, edit{a, b, ""})
		st.vmDecls++
		note("remove", rel, line, what)
	}

	fieldsOf := func(typeName string) *ast.StructType {
		for _, d := range f.Decls {
			if g, ok := d.(*ast.GenDecl); ok {
				for _, sp := range g.Specs {
					if t, ok := sp.(*ast.TypeSpec); ok && t.Name.Name == typeName {
						if s, ok := t.Type.(*ast.StructType); ok {
							return s
						}
					}
				}
			}
		}
		return nil
	}
	dropFields := func(typeName string, fields ...string) {
		s := fieldsOf(typeName)
		if s == nil {
			return
		}
		for _, fl := range s.Fields.List {
			for _, n := range fl.Names {
				for _, want := range fields {
					if n.Name == want {
						a, b := off(fl.Pos()), off(fl.End())
						if fl.Doc != nil {
							a = off(fl.Doc.Pos())
						}
						if fl.Comment != nil {
							b = off(fl.Comment.End())
						}
						cut(a, b, "field "+typeName+"."+want, fset.Position(fl.Pos()).Line)
					}
				}
			}
		}
	}
	dropFields("OverviewVM", "Chain")
	dropFields("FactoryVM", "CardID", "IdleRole", "Roles")
	dropFields("AttentionVM", "Role") // set only by the removed chain-stopped row; its comment documents that row

	litKeys := map[string][]string{"OverviewVM": {"Chain"}, "FactoryVM": {"CardID", "IdleRole", "Roles"}}
	for _, fd := range f.Decls {
		fn, ok := fd.(*ast.FuncDecl)
		if !ok || fn.Body == nil || deleted(off(fn.Pos())) {
			continue
		}
		ast.Inspect(fn, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CompositeLit:
				id, ok := x.Type.(*ast.Ident)
				if !ok {
					return true
				}
				for _, e := range x.Elts {
					kv, ok := e.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					k, ok := kv.Key.(*ast.Ident)
					if !ok {
						continue
					}
					for _, want := range litKeys[id.Name] {
						if k.Name == want {
							b := off(kv.End())
							if b < len(src) && src[b] == ',' {
								b++ // the separator belongs to the element
							}
							cut(off(kv.Pos()), b, "literal key "+id.Name+"."+want, fset.Position(kv.Pos()).Line)
						}
					}
				}
			case *ast.CallExpr:
				if fid, ok := x.Fun.(*ast.Ident); ok && fid.Name == "buildAttention" && len(x.Args) == 3 {
					es = append(es, edit{off(x.Args[1].End()), off(x.Args[2].End()), ""})
					st.vmDecls++
					note("remove", rel, fset.Position(x.Pos()).Line, "buildAttention chain argument")
				}
			}
			return true
		})
		switch fn.Name.Name {
		case "buildAttention":
			ps := fn.Type.Params.List
			if len(ps) == 3 {
				es = append(es, edit{off(ps[1].End()), off(ps[2].End()), ""})
				st.vmDecls++
				note("remove", rel, fset.Position(fn.Pos()).Line, "buildAttention chain parameter")
			}
			for _, s := range fn.Body.List {
				if is, ok := s.(*ast.IfStmt); ok && identsOf(is.Cond)["chain"] {
					cut(off(is.Pos()), off(is.End()), "chain-stopped attention row", fset.Position(is.Pos()).Line)
				}
			}
		case "buildFactory":
			for _, s := range fn.Body.List {
				as, ok := s.(*ast.AssignStmt)
				if !ok {
					continue
				}
				if anyOf(identsOf(as), "chainRoleRecords", "buildChain") {
					a := extendUpComments(fset, f, src, off(as.Pos()))
					if deleted(a) {
						continue
					}
					_, b := lineSpan(src, off(as.Pos()), off(as.End()))
					a, _ = lineSpan(src, a, off(as.Pos()))
					es = append(es, edit{a, b, ""})
					st.vmDecls++
					note("remove", rel, fset.Position(as.Pos()).Line, "chain builder statement")
				}
			}
		}
	}
	if len(es) == 0 {
		return src
	}
	out := formatGo(rel, applyEdits(src, es))
	return dropUnusedLocals(rel, out, "buildOverview", "buildFactory")
}

// dropUnusedLocals removes, in the named functions, the top-level := locals that the chain removal left
// unused: a lone local is deleted with its statement, one of several becomes the blank identifier.
func dropUnusedLocals(rel string, src []byte, funcs ...string) []byte {
	fset, f := parseGo(rel, src)
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	var es []edit
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		wanted := false
		for _, n := range funcs {
			wanted = wanted || fn.Name.Name == n
		}
		if !wanted {
			continue
		}
		count := map[string]int{}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				count[id.Name]++
			}
			return true
		})
		for _, s := range fn.Body.List {
			as, ok := s.(*ast.AssignStmt)
			if !ok || as.Tok != token.DEFINE {
				continue
			}
			var unused []*ast.Ident
			for _, l := range as.Lhs {
				if id, ok := l.(*ast.Ident); ok && id.Name != "_" && count[id.Name] == 1 {
					unused = append(unused, id)
				}
			}
			switch {
			case len(unused) == 0:
			case len(unused) == len(as.Lhs) && len(as.Lhs) == 1:
				a, b := lineSpan(src, off(as.Pos()), off(as.End()))
				es = append(es, edit{a, b, ""})
				st.vmDecls++
				note("remove", rel, fset.Position(as.Pos()).Line, "unused local "+unused[0].Name)
			default:
				for _, id := range unused {
					es = append(es, edit{off(id.Pos()), off(id.End()), "_"})
					st.vmDecls++
					note("remove", rel, fset.Position(as.Pos()).Line, "unused local "+id.Name+" becomes _")
				}
			}
		}
	}
	if len(es) == 0 {
		return src
	}
	return formatGo(rel, applyEdits(src, es))
}

// editTests removes the declarations of a test file that reference the removed names, rewrites the one test
// that also pins the must-fix rows, and drops the imports the removals leave unused. It reports whether
// nothing but imports remains.
func editTests(rel string, src []byte) ([]byte, bool) {
	hasAny := false
	for _, n := range removedNames {
		if bytes.Contains(src, []byte(n)) {
			hasAny = true
		}
	}
	if !hasAny {
		return src, false
	}
	fset, f := parseGo(rel, src)
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	var es []edit
	removed := 0
	for _, d := range f.Decls {
		gd, isGen := d.(*ast.GenDecl)
		if isGen && gd.Tok == token.IMPORT {
			continue
		}
		if !anyOf(identsOf(d), removedNames...) {
			continue
		}
		a, b := off(d.Pos()), off(d.End())
		var doc *ast.CommentGroup
		var name string
		switch x := d.(type) {
		case *ast.FuncDecl:
			doc, name = x.Doc, x.Name.Name
		case *ast.GenDecl:
			doc, name = x.Doc, "declaration"
		}
		if doc != nil {
			a = off(doc.Pos())
		}
		if name == "TestBuildAttentionOrderAndCap" {
			es = append(es, edit{a, b, strings.TrimSuffix(attentionTestSource, "\n")})
			st.testDecls++
			note("rewrite", rel, fset.Position(d.Pos()).Line, name+" (chain row dropped, order and cap kept)")
			continue
		}
		a, b = lineSpan(src, a, b)
		es = append(es, edit{a, b, ""})
		removed++
		st.testDecls++
		note("remove", rel, fset.Position(d.Pos()).Line, "test declaration "+name)
	}
	if len(es) == 0 {
		return src, false
	}
	out := applyEdits(src, es)
	out = dropUnusedImports(rel, src, out)
	return formatOrEmpty(rel, out)
}

func formatOrEmpty(rel string, out []byte) ([]byte, bool) {
	out = formatGo(rel, out)
	_, f := parseGo(rel, out)
	for _, d := range f.Decls {
		if g, ok := d.(*ast.GenDecl); !ok || g.Tok != token.IMPORT {
			return out, false
		}
	}
	return out, true
}

// importName is the identifier the file uses for an import: its alias, else the last path element.
func importName(im *ast.ImportSpec) string {
	if im.Name != nil {
		return im.Name.Name
	}
	p := strings.Trim(im.Path.Value, "\"`")
	base := p[strings.LastIndex(p, "/")+1:]
	if m := regexp.MustCompile(`^v[0-9]+$`).MatchString(base); m {
		rest := strings.TrimSuffix(p, "/"+base)
		base = rest[strings.LastIndex(rest, "/")+1:]
	}
	return base
}

func usedQualifiers(f *ast.File) map[string]bool {
	m := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok {
				m[id.Name] = true
			}
		}
		return true
	})
	return m
}

// dropUnusedImports removes the imports that were used before the edit and are not after it.
func dropUnusedImports(rel string, before, after []byte) []byte {
	_, fb := parseGo(rel, before)
	usedBefore := usedQualifiers(fb)
	fset, fa := parseGo(rel, after)
	usedAfter := usedQualifiers(fa)
	var es []edit
	for _, d := range fa.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.IMPORT {
			continue
		}
		var drop []*ast.ImportSpec
		for _, sp := range g.Specs {
			im := sp.(*ast.ImportSpec)
			name := importName(im)
			if name == "_" || name == "." {
				continue
			}
			if usedBefore[name] && !usedAfter[name] {
				drop = append(drop, im)
			}
		}
		if len(drop) == 0 {
			continue
		}
		if len(drop) == len(g.Specs) {
			a, b := lineSpan(after, fset.Position(g.Pos()).Offset, fset.Position(g.End()).Offset)
			es = append(es, edit{a, b, ""})
		} else {
			for _, im := range drop {
				a, b := lineSpan(after, fset.Position(im.Pos()).Offset, fset.Position(im.End()).Offset)
				es = append(es, edit{a, b, ""})
			}
		}
		st.importsDropped += len(drop)
	}
	if len(es) == 0 {
		return after
	}
	return applyEdits(after, es)
}

// ---------------------------------------------------------------- templ, scripts

// removePanel cuts the chain session board block out of screens.templ: from the panel head that carries the
// title to the line before the note banner that follows the role cards.
func removePanel(rel string, src []byte) []byte {
	t := string(src)
	i := strings.Index(t, "Chain session board")
	if i < 0 {
		return src
	}
	hs := strings.LastIndex(t[:i], `<div class="panel__head"`)
	nb := strings.Index(t[i:], "@noteBanner(")
	if hs < 0 || nb < 0 {
		fatal(1, "%s: the chain session board is not in the expected shape (panel head or note banner not found); nothing written", rel)
	}
	start := strings.LastIndex(t[:hs], "\n") + 1
	end := strings.LastIndex(t[:i+nb], "\n") + 1
	blk := t[start:end]
	if end <= start || !strings.Contains(blk, "roles--chain") || strings.Contains(blk, ".lanes\"") {
		fatal(1, "%s: the chain session board block is not the expected one; nothing written", rel)
	}
	st.panels++
	note("remove", rel, strings.Count(t[:start], "\n")+1, fmt.Sprintf("chain session board panel (%d lines)", strings.Count(blk, "\n")))
	return []byte(t[:start] + t[end:])
}

func rewriteTemplText(rel string, src []byte) []byte {
	out, n := rewriteProse(string(src))
	st.templ += n
	if n > 0 {
		note("templ", rel, 0, fmt.Sprintf("%d spelling(s)", n))
	}
	return []byte(out)
}

func rewriteScript(rel string, src []byte) []byte {
	if filepath.Base(rel) == "i18n.js" {
		var lines []string
		for _, l := range strings.Split(string(src), "\n") {
			if i18nRemove.MatchString(l) {
				st.i18nKeys++
				note("i18n", rel, len(lines)+1, "removed "+strings.TrimSpace(l))
				continue
			}
			if m := i18nTitleKey.FindStringSubmatch(l); m != nil {
				if nv, ok := titleValue[m[2]]; ok {
					l = m[1] + nv + m[3]
				}
			}
			// the Chinese strings of the SPEC board page use the board word for "board" and "dashboard"
			// (the English values are Board and dashboard), which the word sweep also matches; they take
			// the Chinese words for panel and dashboard.
			if m := zhBoardTitle.FindStringSubmatch(l); m != nil {
				l = m[1] + "面板" + m[2]
			} else if m := zhBoardSubtitle.FindStringSubmatch(l); m != nil {
				l = m[1] + "仪表板" + m[2]
			}
			lines = append(lines, l)
		}
		src = []byte(strings.Join(lines, "\n"))
	}
	out, n := rewriteProse(string(src))
	st.js += n
	if n > 0 {
		note("js", rel, 0, fmt.Sprintf("%d spelling(s)", n))
	}
	return []byte(out)
}

// ---------------------------------------------------------------- remaining-word report

var wordCJK = []string{"칸반", "かんばん", "カンバン", "看板"}
var markerTail = regexp.MustCompile(`^_(id|lead_addr|lead_name|settings_injected|backend|card)\b`)

// countWord counts the occurrences the AC-018 pattern would match in one text.
func countWord(s string) int {
	n := 0
	low := strings.ToLower(s)
	for i := 0; ; {
		j := strings.Index(low[i:], "kanban")
		if j < 0 {
			break
		}
		a := i + j
		i = a + len("kanban")
		if a >= 5 && low[a-5:a] == "moai_" && markerTail.MatchString(low[i:]) {
			continue // a frozen marker name (MOAI_KANBAN_ID and its five siblings)
		}
		n++
	}
	for _, w := range wordCJK {
		n += strings.Count(s, w)
	}
	return n
}

func main() {
	rootFlag := flag.String("root", "", "git tree to rewrite")
	dry := flag.Bool("dry-run", false, "report what would change and write nothing")
	flag.BoolVar(&verbose, "v", false, "list every rewrite with its file and line")
	flag.Parse()
	if *rootFlag == "" {
		fatal(2, "need -root <git tree>")
	}
	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		fatal(2, "%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		fatal(2, "%s has no go.mod", root)
	}
	if out, err := git(root, "status", "--porcelain", "--untracked-files=no"); err != nil {
		fatal(2, "git status failed: %v\n%s", err, out)
	} else if strings.TrimSpace(out) != "" {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		shown := lines
		if len(shown) > 10 {
			shown = shown[:10]
		}
		fatal(2, "the tree has %d tracked modification(s); commit or revert them first (first %d shown):\n%s", len(lines), len(shown), strings.Join(shown, "\n"))
	}

	// the files of the unit: non-generated Go, templ sources, the two scripts
	dir := filepath.Join(root, webDir)
	ents, err := os.ReadDir(dir)
	if err != nil {
		fatal(2, "%v", err)
	}
	var goRels, genRels, templRels []string
	for _, e := range ents {
		n := e.Name()
		switch {
		case e.IsDir():
		case strings.HasSuffix(n, "_templ.go"):
			genRels = append(genRels, webDir+"/"+n)
		case strings.HasSuffix(n, ".go"):
			goRels = append(goRels, webDir+"/"+n)
		case strings.HasSuffix(n, ".templ"):
			templRels = append(templRels, webDir+"/"+n)
		}
	}
	sort.Strings(goRels)
	read := func(rel string) []byte {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			fatal(1, "read %s: %v", rel, err)
		}
		return b
	}

	out := map[string][]byte{}  // rel -> new content
	orig := map[string][]byte{} // rel -> original
	var removeFiles []string

	// steps 1 and 2: Go sources
	for _, rel := range goRels {
		orig[rel] = read(rel)
		out[rel] = rewriteGo(rel, orig[rel])
	}
	// collision check: a new name may not be declared already
	declared := map[string]string{}
	for _, rel := range append(append([]string{}, goRels...), genRels...) {
		src := orig[rel]
		if src == nil {
			src = read(rel)
		}
		for _, n := range declaredNames(src, rel) {
			declared[n] = rel
		}
	}
	for old, nw := range st.identNames {
		if where, ok := declared[nw]; ok && declared[old] != "" {
			fatal(1, "collision: %s would be renamed %s, which %s already declares; nothing written", old, nw, where)
		}
	}

	// step 4: the chain view model and its tests
	for _, rel := range goRels {
		base := filepath.Base(rel)
		switch {
		case base == "viewmodel_ops.go":
			out[rel] = editViewModel(rel, out[rel])
		case strings.HasSuffix(base, "_test.go"):
			var empty bool
			out[rel], empty = editTests(rel, out[rel])
			if empty {
				removeFiles = append(removeFiles, rel)
			}
		}
	}
	removeFiles = append(removeFiles, dropOrphans(out, goRels)...)
	// the legacy route registration
	if app := webDir + "/app.go"; out[app] != nil {
		a := string(out[app])
		if !strings.Contains(a, "registerLegacyRoutes(mux)") {
			anchor := `mux.HandleFunc("/factory", a.handleFactory)`
			if !strings.Contains(a, anchor) {
				fatal(1, "%s: the /factory route registration is not in the expected shape; nothing written", app)
			}
			a = strings.Replace(a, anchor, anchor+"\n\tregisterLegacyRoutes(mux)", 1)
			out[app] = formatGo(app, []byte(a))
			st.legacy = "registered"
		}
	}
	legacyPath := webDir + "/legacy_routes.go"
	if _, err := os.Stat(filepath.Join(root, legacyPath)); err != nil {
		out[legacyPath] = []byte(legacyRoutesSource)
		if st.legacy == "" {
			st.legacy = "created"
		} else {
			st.legacy = "created and registered"
		}
	}

	// templ sources and scripts
	for _, rel := range templRels {
		orig[rel] = read(rel)
		b := rewriteTemplText(rel, removePanel(rel, orig[rel]))
		out[rel] = b
	}
	for _, rel := range []string{webDir + "/assets/app.js", webDir + "/assets/i18n.js", webDir + "/assets/console.css"} {
		orig[rel] = read(rel)
		out[rel] = rewriteScript(rel, orig[rel])
	}

	// step 5: file names
	type mv struct{ from, to string }
	var moves []mv
	for _, rel := range goRels {
		base := filepath.Base(rel)
		if strings.Contains(strings.ToLower(base), "kanban") {
			to := webDir + "/" + strings.NewReplacer("kanban", "factory", "Kanban", "Factory").Replace(base)
			if _, err := os.Stat(filepath.Join(root, to)); err == nil {
				fatal(1, "move collision: %s already exists; nothing written", to)
			}
			moves = append(moves, mv{rel, to})
		}
	}

	changed := 0
	removing := map[string]bool{}
	for _, rel := range removeFiles {
		removing[rel] = true
	}
	if !*dry {
		for rel, b := range out {
			if removing[rel] {
				continue
			}
			if o, ok := orig[rel]; ok && bytes.Equal(o, b) {
				continue
			}
			changed++
			if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), b, 0o644); err != nil {
				fatal(1, "write %s: %v", rel, err)
			}
		}
		for _, rel := range removeFiles {
			if o, _ := git(root, "ls-files", "--error-unmatch", "--", rel); strings.TrimSpace(o) != "" {
				if o, err := git(root, "rm", "-q", "-f", "--", rel); err != nil {
					fatal(1, "git rm %s: %v\n%s", rel, err, o)
				}
			} else {
				_ = os.Remove(filepath.Join(root, filepath.FromSlash(rel)))
			}
			delete(out, rel)
		}
		for _, m := range moves {
			if o, err := git(root, "mv", m.from, m.to); err != nil {
				fatal(1, "git mv %s %s: %v\n%s", m.from, m.to, err, o)
			}
			out[m.to] = out[m.from]
			delete(out, m.from)
		}
	} else {
		for rel, b := range out {
			if removing[rel] {
				continue
			}
			if o, ok := orig[rel]; !ok || !bytes.Equal(o, b) {
				changed++
			}
		}
		for _, rel := range removeFiles {
			delete(out, rel)
		}
		for _, m := range moves {
			out[m.to] = out[m.from]
			delete(out, m.from)
		}
	}
	st.filesRemoved = removeFiles
	for _, m := range moves {
		st.filesRenamed = append(st.filesRenamed, m.from+" -> "+m.to)
	}

	// step 7: regenerate the *_templ.go files
	gen := "skipped (dry run)"
	if !*dry {
		cmd := exec.Command("go", "run", "github.com/a-h/templ/cmd/templ", "generate", "-path", "./"+webDir)
		cmd.Dir = root
		o, err := cmd.CombinedOutput()
		if err != nil {
			fatal(1, "templ generate failed: %v\n%s", err, o)
		}
		gen = "exit 0"
	}

	distinct := len(st.identNames)
	mode := ""
	if *dry {
		mode = " (dry run, nothing written)"
	}
	fmt.Printf("m9_rename: identifiers: %d (%d distinct); string literals: %d; comment mentions: %d; templ spellings: %d; script spellings: %d; i18n keys removed: %d; panel blocks removed: %d; view-model removals: %d; test declarations removed or rewritten: %d; imports dropped: %d; files rewritten: %d; files removed: %d; files renamed: %d; legacy route: %s; templ generate: %s%s\n",
		st.idents, distinct, st.literals, st.comments, st.templ, st.js, st.i18nKeys, st.panels, st.vmDecls, st.testDecls, st.importsDropped, changed, len(st.filesRemoved), len(st.filesRenamed), orDash(st.legacy), gen, mode)
	for _, m := range st.filesRenamed {
		fmt.Println("  renamed", m)
	}
	for _, r := range st.filesRemoved {
		fmt.Println("  removed", r)
	}

	// informational: what still carries the word (the hand-edit list)
	type rem struct {
		rel string
		n   int
	}
	var left []rem
	if *dry {
		for rel, b := range out {
			if n := countWord(string(b)); n > 0 && filepath.Base(rel) != "legacy_routes.go" {
				left = append(left, rem{rel, n})
			}
		}
	} else {
		_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			if n := countWord(string(b)); n > 0 && filepath.Base(rel) != "legacy_routes.go" {
				left = append(left, rem{rel, n})
			}
			return nil
		})
	}
	sort.Slice(left, func(i, j int) bool { return left[i].rel < left[j].rel })
	nonTest, nonTestFiles, tests := 0, 0, 0
	for _, l := range left {
		if strings.HasSuffix(l.rel, "_test.go") {
			tests += l.n
		} else {
			nonTest += l.n
			nonTestFiles++
		}
	}
	fmt.Printf("m9_rename: still carrying the word outside legacy_routes.go (informational, hand-edit list): %d occurrence(s) in %d non-test file(s) and %d in test files\n", nonTest, nonTestFiles, tests)
	if verbose {
		for _, l := range left {
			fmt.Printf("  %s: %d\n", l.rel, l.n)
		}
	}
}

func orDash(s string) string {
	if s == "" {
		return "present"
	}
	return s
}
