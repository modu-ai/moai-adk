//go:build ignore

// m7_rename.go is the re-runnable M7 rename of SPEC-LAUNCHER-ENTRY-FLAGS-001 (card t1399): Go identifiers
// and Go file names outside internal/web. It is the mechanical half of milestone M7, written as a program so
// that the lane can absorb the integration branch and run it again to convert references other lanes added
// meanwhile (a new use of an old identifier, a new file carrying the old word in its name).
//
//	go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m7_rename.go -root <git tree> [-dry-run]
//
// What it does, in this order (rules adapted from probe/probe.go stageM7, re-measured on the live tree):
//
//  1. deletes the field KanbanEnabled of the entry-parse struct (kanbanEntryParse or launcherEntryParse);
//  2. collapses the two launch-facts functions: where a package declares exportKanbanLaunchFacts, the wrapper
//     exportFactoryLaunchFacts is deleted, and the rename below turns the survivor into exportFactoryLaunchFacts;
//  3. renames every Go identifier that carries the word (any letter case), except the bare package name
//     `kanban`, package clauses (M8), the four test names that the acceptance commands pin (preserve), and
//     KanbanEnabled (reported if still referenced); the six marker constants take the explicit map of
//     design.md section 4.7 (their STRING VALUES are never touched: only identifiers and the comments and
//     string literals that spell an identifier are rewritten), and kanbanEntryParse becomes launcherEntryParse;
//  4. in comments and string literals, rewrites tokens that spell a renamed identifier (camel case with the
//     capital-K form, or kanbanXxx), so a message or comment naming a renamed function stays true; in
//     comments only, it also follows the file renames of step 6 (a comment that says kanban.go says
//     factory_launch_helpers.go afterwards);
//  5. applies the literal rewrites of design.md section 4.7 "Names that are strings or texts" that M7 owns:
//     the transient settings prefix moai-kanban to moai-factory (not the skill id moai-kanban-foreman), the error
//     text prefixes of the five landing and backlog files, the timing lap kanban_record, and the CLI/MCP help
//     sentences;
//  6. renames the Go files whose NAME carries the word (kanban.go to factory_launch_helpers.go, every other to
//     the same name with the word replaced), with git mv for tracked files.
//
// It never touches: internal/web (milestone M9), non-Go files, testdata, node_modules, vendor, .git, .moai,
// .claude, internal/template/templates; the package path and the package directory (M8); rule, skill, template
// and documentation text (M10, M11); comments that merely say the word (reworded by hand, listed in progress.md).
//
// Safety: it refuses (exit 2) when the tree has tracked modifications; it refuses (exit 1, nothing written) when
// a rename target already exists, when the rewritten source does not parse, or when an identifier rename would
// land on a name already declared at package level in the same directory. It is idempotent: a second run on an
// already-renamed tree finds nothing to change, prints zeros, and exits 0.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// preserve names the identifiers that stay: the test names the acceptance commands select by name.
var preserve = map[string]bool{
	"TestKanbanEntryRefused":                  true, // AC-011
	"TestSessionStartEmitsNoKanbanNotice":     true, // AC-014
	"TestPreexistingKanbanArtifactsTolerated": true, // AC-017
	"TestLegacyKanbanRouteRedirects":          true, // AC-019 (internal/web, M9)
}

// explicit is the design.md section 4.7 map; every other identifier takes the word replacement.
var explicit = map[string]string{
	"EnvMoaiKanbanID":               "EnvFactoryRunID",
	"EnvMoaiKanbanLeadAddr":         "EnvFactoryLeadAddr",
	"EnvMoaiKanbanLeadName":         "EnvFactoryLeadName",
	"EnvMoaiKanbanSettingsInjected": "EnvFactorySettingsInjected",
	"EnvMoaiKanbanBackend":          "EnvFactoryBackend",
	"EnvMoaiKanbanCard":             "EnvFactoryCard",
	"kanbanEntryParse":              "launcherEntryParse",
}

var replacer = strings.NewReplacer("Kanban", "Factory", "kanban", "factory", "KANBAN", "FACTORY")

var tokenRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// textPairs records every old to new token pair rewritten in a comment or string literal, for the report.
var textPairs = map[string]string{}

// rename returns the new name of an identifier (text=false) or of a token spelled in a comment or string
// (text=true, which only accepts camel-case Go spellings, never MOAI_KANBAN_* or kanban_snake file names).
func rename(id string, text bool) (string, bool) {
	if preserve[id] || id == "KanbanEnabled" || id == "kanban" {
		return id, false
	}
	if n, ok := explicit[id]; ok {
		return n, true
	}
	if !strings.Contains(strings.ToLower(id), "kanban") {
		return id, false
	}
	if text && !isCamelKanban(id) {
		return id, false
	}
	n := replacer.Replace(id)
	if text && n != id {
		textPairs[id] = n
	}
	return n, n != id
}

func isCamelKanban(tok string) bool {
	if tok == "Kanban" {
		return false
	}
	if strings.Contains(tok, "Kanban") {
		return true
	}
	if len(tok) > 6 && strings.HasPrefix(tok, "kanban") {
		c := tok[6]
		return (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
	}
	return false
}

// renamedFiles maps the base names of the files this program renames (its own moves, listed here so a re-run
// still converts a comment another lane wrote after the move) to their new base names.
var renamedFiles = func() map[string]string {
	m := map[string]string{}
	for _, old := range []string{
		"kanban.go", "kanban_settings.go", "kanban_dispatch_test.go", "kanban_autonomy_test.go",
		"kanban_bootstrap_test.go", "kanban_launch_facts_test.go", "kanban_lead_name_test.go",
		"kanban_settings_test.go", "kanban_helper_test.go", "session_start_no_kanban_notice_test.go",
		"preexisting_kanban_artifacts_m1_test.go",
	} {
		m[old] = fileName("internal/cli", old)
	}
	return m
}()

var oldFileRe = func() *regexp.Regexp {
	names := make([]string, 0, len(renamedFiles))
	for k := range renamedFiles {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	for i, n := range names {
		names[i] = regexp.QuoteMeta(n)
	}
	return regexp.MustCompile(strings.Join(names, "|"))
}()

func isWordByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// replaceFileNames rewrites whole-name mentions of the renamed files in a comment.
func replaceFileNames(s string) (string, int) {
	var sb strings.Builder
	last, n := 0, 0
	for _, m := range oldFileRe.FindAllStringIndex(s, -1) {
		if m[0] > 0 && isWordByte(s[m[0]-1]) || m[1] < len(s) && isWordByte(s[m[1]]) {
			continue
		}
		sb.WriteString(s[last:m[0]])
		sb.WriteString(renamedFiles[s[m[0]:m[1]]])
		last = m[1]
		n++
	}
	if n == 0 {
		return s, 0
	}
	sb.WriteString(s[last:])
	return sb.String(), n
}

type edit struct {
	a, b int
	s    string
}

func applyEdits(src []byte, es []edit) []byte {
	sort.Slice(es, func(i, j int) bool { return es[i].a > es[j].a })
	out := append([]byte(nil), src...)
	for _, e := range es {
		out = append(out[:e.a], append([]byte(e.s), out[e.b:]...)...)
	}
	return out
}

type stats struct {
	idents, textTokens, literals, fieldDel, wrapperDel, fileRenames int
	distinct                                                        map[string]string
	perPkg                                                          map[string]int
	files                                                           map[string]bool
	leftoverEnabled                                                 []string
}

func fatal(code int, format string, a ...any) {
	fmt.Fprintf(os.Stderr, "m7_rename: "+format+"\n", a...)
	os.Exit(code)
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var skipDirs = map[string]bool{".git": true, ".moai": true, ".claude": true, "node_modules": true, "vendor": true, "testdata": true}

func skipRel(rel string) bool {
	return rel == "internal/web" || rel == "internal/template/templates"
}

func walkGo(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if skipDirs[d.Name()] || skipRel(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") {
			out = append(out, rel)
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func main() {
	rootFlag := flag.String("root", "", "git tree to rewrite")
	dry := flag.Bool("dry-run", false, "report what would change and write nothing")
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
		fatal(2, "the tree has tracked modifications; commit or revert them first:\n%s", out)
	}

	st := &stats{distinct: map[string]string{}, perPkg: map[string]int{}, files: map[string]bool{}}
	rels := walkGo(root)

	// load the files that carry the word (case-insensitive); everything else is read on demand for the
	// package-level collision check
	src := map[string][]byte{}
	for _, rel := range rels {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			fatal(1, "read %s: %v", rel, err)
		}
		if bytes.Contains(bytes.ToLower(b), []byte("kanban")) {
			src[rel] = b
		}
	}

	// steps 1 and 2: structural edits
	collapseDirs := map[string]bool{}
	for rel, b := range src {
		if bytes.Contains(b, []byte("func exportKanbanLaunchFacts(")) {
			collapseDirs[filepath.Dir(rel)] = true
		}
	}
	changed := map[string][]byte{}
	var relsLoaded []string
	for rel := range src {
		relsLoaded = append(relsLoaded, rel)
	}
	sort.Strings(relsLoaded)
	for _, rel := range relsLoaded {
		b := src[rel]
		nb, fd, wd := structural(rel, b, collapseDirs[filepath.Dir(rel)])
		st.fieldDel += fd
		st.wrapperDel += wd
		if fd+wd > 0 {
			changed[rel] = nb
		}
	}
	cur := func(rel string) []byte {
		if b, ok := changed[rel]; ok {
			return b
		}
		return src[rel]
	}

	// declared package-level names per directory, for the collision check
	declared := map[string]map[string]bool{}
	for _, rel := range rels {
		d := filepath.Dir(rel)
		b := cur(rel)
		if b == nil {
			var err error
			b, err = os.ReadFile(filepath.Join(root, rel))
			if err != nil {
				continue
			}
		}
		if declared[d] == nil {
			declared[d] = map[string]bool{}
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, rel, b, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		for _, name := range topLevelNames(f) {
			declared[d][name] = true
		}
	}

	// steps 3 to 5: identifiers, comment and string tokens, literal rewrites
	type collision struct{ dir, old, new string }
	var collisions []collision
	for _, rel := range relsLoaded {
		b := cur(rel)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, rel, b, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			fatal(1, "parse %s: %v", rel, err)
		}
		var es []edit
		off := func(p token.Pos) int { return fset.Position(p).Offset }
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.File:
				return true
			case *ast.Ident:
				if x == f.Name {
					return false
				}
				if x.Name == "KanbanEnabled" {
					st.leftoverEnabled = append(st.leftoverEnabled, fmt.Sprintf("%s:%d", rel, fset.Position(x.Pos()).Line))
					return false
				}
				if nn, ok := rename(x.Name, false); ok {
					a := off(x.Pos())
					es = append(es, edit{a, a + len(x.Name), nn})
					st.idents++
					st.distinct[x.Name] = nn
					st.perPkg[filepath.Dir(rel)]++
					st.files[rel] = true
				}
			case *ast.BasicLit:
				if x.Kind != token.STRING {
					return true
				}
				nv, nl, nt := rewriteLiteral(rel, x.Value)
				if nv != x.Value {
					a := off(x.Pos())
					es = append(es, edit{a, a + len(x.Value), nv})
					st.literals += nl
					st.textTokens += nt
					st.files[rel] = true
				}
			}
			return true
		})
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				nt := 0
				nv := tokenRe.ReplaceAllStringFunc(c.Text, func(tok string) string {
					if nn, ok := rename(tok, true); ok {
						nt++
						return nn
					}
					return tok
				})
				// a comment that names one of the files this program renames follows the file
				var nf int
				nv, nf = replaceFileNames(nv)
				nt += nf
				if nv != c.Text {
					a := off(c.Slash)
					es = append(es, edit{a, a + len(c.Text), nv})
					st.textTokens += nt
					st.files[rel] = true
				}
			}
		}
		if len(es) == 0 {
			continue
		}
		changed[rel] = applyEdits(b, es)
	}
	// collisions: an old package-level name whose new name is already declared in the same directory
	for rel := range src {
		d := filepath.Dir(rel)
		for old, nn := range st.distinct {
			if declared[d][old] && declared[d][nn] {
				collisions = append(collisions, collision{d, old, nn})
			}
		}
	}
	if len(collisions) > 0 {
		sort.Slice(collisions, func(i, j int) bool {
			return collisions[i].dir+collisions[i].old < collisions[j].dir+collisions[j].old
		})
		seen := map[string]bool{}
		for _, c := range collisions {
			k := c.dir + c.old
			if !seen[k] {
				seen[k] = true
				fmt.Fprintf(os.Stderr, "COLLISION: %s: %s -> %s (target already declared)\n", c.dir, c.old, c.new)
			}
		}
		fatal(1, "%d rename collision(s); nothing written", len(seen))
	}

	// gofmt every changed file and verify it parses
	var changedRels []string
	for rel := range changed {
		changedRels = append(changedRels, rel)
	}
	sort.Strings(changedRels)
	for _, rel := range changedRels {
		out, err := format.Source(changed[rel])
		if err != nil {
			fatal(1, "gofmt %s: %v; nothing written", rel, err)
		}
		changed[rel] = out
	}

	// step 6: file renames
	type mv struct{ from, to string }
	var moves []mv
	exists := map[string]bool{}
	for _, rel := range rels {
		exists[rel] = true
	}
	for _, rel := range rels {
		base := filepath.Base(rel)
		if !strings.Contains(strings.ToLower(base), "kanban") {
			continue
		}
		dir := filepath.Dir(rel)
		nb := fileName(dir, base)
		to := filepath.ToSlash(filepath.Join(dir, nb))
		if exists[to] {
			fatal(1, "file rename collision: %s already exists (would replace %s); nothing written", to, rel)
		}
		moves = append(moves, mv{rel, to})
	}
	st.fileRenames = len(moves)

	if !*dry {
		for _, rel := range changedRels {
			if err := os.WriteFile(filepath.Join(root, rel), changed[rel], 0o644); err != nil {
				fatal(1, "write %s: %v", rel, err)
			}
		}
		for _, m := range moves {
			if out, _ := git(root, "ls-files", "--", m.from); strings.TrimSpace(out) != "" {
				if o, err := git(root, "mv", m.from, m.to); err != nil {
					fatal(1, "git mv %s %s: %v\n%s", m.from, m.to, err, o)
				}
			} else if err := os.Rename(filepath.Join(root, m.from), filepath.Join(root, m.to)); err != nil {
				fatal(1, "rename %s: %v", m.from, err)
			}
		}
	}

	// report
	pk := make([]string, 0, len(st.perPkg))
	for k := range st.perPkg {
		pk = append(pk, k)
	}
	sort.Strings(pk)
	for _, k := range pk {
		fmt.Printf("  identifiers in %s: %d\n", k, st.perPkg[k])
	}
	tp := make([]string, 0, len(textPairs))
	for k := range textPairs {
		tp = append(tp, k)
	}
	sort.Strings(tp)
	for _, k := range tp {
		fmt.Printf("  text token (comment or string): %s -> %s\n", k, textPairs[k])
	}
	for _, m := range moves {
		fmt.Printf("  file: %s -> %s\n", m.from, m.to)
	}
	for _, l := range st.leftoverEnabled {
		fmt.Printf("  LEFTOVER (not renamed, delete by hand): KanbanEnabled at %s\n", l)
	}
	mode := ""
	if *dry {
		mode = " (dry run, nothing written)"
	}
	fmt.Printf("m7_rename: identifiers renamed: %d (%d distinct names, %d files); comment/string tokens rewritten: %d; literal rewrites: %d; KanbanEnabled fields deleted: %d; launch-facts wrappers deleted: %d; files renamed: %d%s\n",
		st.idents, len(st.distinct), len(st.files), st.textTokens, st.literals, st.fieldDel, st.wrapperDel, st.fileRenames, mode)
}

// fileName maps a base file name to its new name.
func fileName(dir, base string) string {
	if dir == "internal/cli" {
		switch base {
		case "kanban.go":
			return "factory_launch_helpers.go"
		case "kanban_dispatch_test.go":
			return "factory_launch_dispatch_test.go"
		}
	}
	i := strings.Index(strings.ToLower(base), "kanban")
	return base[:i] + "factory" + base[i+len("kanban"):]
}

func topLevelNames(f *ast.File) []string {
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

// structural deletes the KanbanEnabled field and, in a collapse directory, the exportFactoryLaunchFacts wrapper.
func structural(rel string, b []byte, collapse bool) ([]byte, int, int) {
	needField := bytes.Contains(b, []byte("KanbanEnabled"))
	needWrapper := collapse && bytes.Contains(b, []byte("func exportFactoryLaunchFacts("))
	if !needField && !needWrapper {
		return b, 0, 0
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, rel, b, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		fatal(1, "parse %s: %v", rel, err)
	}
	var es []edit
	fd, wd := 0, 0
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	lineStart := func(o int) int {
		for o > 0 && b[o-1] != '\n' {
			o--
		}
		return o
	}
	lineEnd := func(o int) int {
		for o < len(b) && b[o] != '\n' {
			o++
		}
		if o < len(b) {
			o++
		}
		return o
	}
	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.GenDecl:
			if !needField {
				continue
			}
			for _, sp := range x.Specs {
				ts, ok := sp.(*ast.TypeSpec)
				if !ok || (ts.Name.Name != "kanbanEntryParse" && ts.Name.Name != "launcherEntryParse") {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, fl := range st.Fields.List {
					for _, n := range fl.Names {
						if n.Name == "KanbanEnabled" {
							a := lineStart(off(fl.Pos()))
							if fl.Doc != nil {
								a = lineStart(off(fl.Doc.Pos()))
							}
							es = append(es, edit{a, lineEnd(off(fl.End())), ""})
							fd++
						}
					}
				}
			}
		case *ast.FuncDecl:
			if needWrapper && x.Recv == nil && x.Name.Name == "exportFactoryLaunchFacts" {
				a := off(x.Pos())
				if x.Doc != nil {
					a = off(x.Doc.Pos())
				}
				es = append(es, edit{a, off(x.End()), ""})
				wd++
			}
		}
	}
	if len(es) == 0 {
		return b, 0, 0
	}
	return applyEdits(b, es), fd, wd
}

// rewriteLiteral applies the section 4.7 literal rewrites and the identifier-spelling rewrite to one string
// literal (its source text, quotes included). It returns the new text, the number of literal rewrites, and the
// number of identifier tokens rewritten.
func rewriteLiteral(rel, lit string) (string, int, int) {
	out := lit
	nl, nt := 0, 0
	base := filepath.Base(rel)
	dir := filepath.Dir(rel)
	// the transient settings prefix (not the skill id moai-kanban-foreman)
	if dir == "internal/cli" && strings.Contains(out, "moai-kanban") {
		var sb strings.Builder
		rest := out
		for {
			i := strings.Index(rest, "moai-kanban")
			if i < 0 {
				sb.WriteString(rest)
				break
			}
			after := rest[i+len("moai-kanban"):]
			sb.WriteString(rest[:i])
			if strings.HasPrefix(after, "-foreman") {
				sb.WriteString("moai-kanban")
			} else {
				sb.WriteString("moai-factory")
				nl++
			}
			rest = after
		}
		out = sb.String()
	}
	// error text prefixes of the five landing and backlog files
	switch base {
	case "backlog_store.go", "backlog_sqlite.go", "landing_verdict.go", "landing_evidence.go", "autodone_scan.go":
		if dir == "internal/kanban" || dir == "internal/factory" {
			if len(out) > 1 {
				q, body := out[:1], out[1:]
				if strings.HasPrefix(body, "kanban backlog ") {
					out = q + "todo queue " + strings.TrimPrefix(body, "kanban backlog ")
					nl++
				} else if strings.HasPrefix(body, "kanban: ") {
					out = q + "factory: " + strings.TrimPrefix(body, "kanban: ")
					nl++
				}
			}
		}
	}
	// the timing lap
	if dir == "internal/hook" && base == "session_start.go" && out == `"kanban_record"` {
		out = `"factory_record"`
		nl++
	}
	// CLI and MCP help sentences
	if dir == "internal/cli" {
		switch base {
		case "todo.go", "gtd.go", "mcp_todo.go":
			if strings.Contains(out, "kanban backlog queue") {
				out = strings.ReplaceAll(out, "kanban backlog queue", "backlog queue")
				nl++
			}
		case "tokens.go":
			if strings.Contains(out, "Kanban card id label") {
				out = strings.ReplaceAll(out, "Kanban card id label", "Card id label")
				nl++
			}
		}
	}
	// identifier spellings inside the literal
	out = tokenRe.ReplaceAllStringFunc(out, func(tok string) string {
		if nn, ok := rename(tok, true); ok {
			nt++
			return nn
		}
		return tok
	})
	return out, nl, nt
}
