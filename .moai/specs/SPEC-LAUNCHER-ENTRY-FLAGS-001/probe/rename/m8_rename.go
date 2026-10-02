//go:build ignore

// m8_rename.go is the re-runnable M8 rename of SPEC-LAUNCHER-ENTRY-FLAGS-001 (card t1399): the Go package
// internal/kanban becomes internal/factory. It is the mechanical half of milestone M8, written as a program so
// that the lane can absorb the integration branch and run it again to convert references other lanes added
// meanwhile (a new import of the old path, a new test file inside the old directory).
//
//	go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/rename/m8_rename.go -root <git tree> [-dry-run]
//
// What it does, in this order (rules adapted from probe/probe.go stageM8, re-measured on the live tree):
//
//  1. moves every file of internal/kanban to internal/factory (git mv for tracked files, a plain rename for
//     untracked ones; a target that already exists is a collision and nothing is moved) and removes the old
//     directory when it is empty;
//  2. in every Go file that is inside the moved package or imports the old path (syntax-tree based, so a string,
//     a comment, or a shadowing local is handled by position and not by text):
//     a. rewrites the package clause `kanban` / `kanban_test` to `factory` / `factory_test`;
//     b. rewrites the import path "github.com/modu-ai/moai-adk/internal/kanban" to ".../internal/factory"
//     (an aliased import keeps its alias);
//     c. rewrites every qualifier `kanban.X` of an unaliased import to `factory.X` (an identifier named kanban
//     that resolves to a local declaration is left alone);
//     d. renames a local variable, constant, or parameter named `factory` to `factoryRun` in a file that imports
//     the old path, because left alone it shadows the renamed qualifier and fails the build;
//     e. rewrites string literals that spell the package path (internal/kanban as a path, "./internal/kanban"
//     run-time package lists, the sibling-relative spelling "../kanban" a test of a neighbouring package uses,
//     the sample test-output line of evidence_writer.go), the two adjacent literals "internal", "kanban" of a
//     path join, and the home-state coverage key "kanban" of internal/cli/home_state_coverage.go;
//     f. rewrites comments that name the package path (internal/kanban, ../kanban) and the package doc line
//     "Package kanban";
//  3. gofmts every file it rewrote (go/format, which also re-sorts the import specs of a block).
//
// It never touches: non-Go files, testdata, node_modules, vendor, .git, .moai, .claude,
// internal/template/templates; the six marker string values and the legacy state-directory name
// ".moai/state/kanban" (data, not the package; they are not path literals this program rewrites); the
// codemaps record test internal/graph/codemaps_fold_guard_test.go, whose string literals name unit paths of the
// stale generated codemaps record (a data file the program does not own); a comment that names a path preceded
// by a colon (a git revision spelled rev:path); the other prose that says the word (qualifier mentions in
// comments such as kanban.BacklogStore are left: M7 residue, not owned by M8).
//
// Safety: it refuses (exit 2) when the tree has tracked modifications; it refuses (exit 1, nothing written)
// when a move target already exists, when a rewritten source does not parse, when a file that imports the old
// path also imports another package named factory, when a package-level declaration named factory would clash
// with the new import name, or when a shadowing rename target already exists in the file. It is idempotent: a
// second run on an already-renamed tree changes nothing, prints zeros, and exits 0.
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
	"sort"
	"strings"
)

const (
	oldImport = "github.com/modu-ai/moai-adk/internal/kanban"
	newImport = "github.com/modu-ai/moai-adk/internal/factory"
	oldDir    = "internal/kanban"
	newDir    = "internal/factory"
)

// literalDeny lists the files whose string literals and comments name paths of data this program does not own.
var literalDeny = map[string]bool{
	"internal/graph/codemaps_fold_guard_test.go": true, // unit paths of the stale generated codemaps record
}

type edit struct {
	a, b int
	s    string
}

type stats struct {
	moved, clauses, importLines, qualifiers, shadows, literals, comments int
	files                                                                map[string]bool
}

// verbose (-v) lists every string-literal and comment rewrite with its file and line.
var verbose bool

func note(kind, rel string, line int, old, nw string) {
	if verbose {
		fmt.Printf("  %s %s:%d: %s -> %s\n", kind, rel, line, strings.TrimSpace(old), strings.TrimSpace(nw))
	}
}

func fatal(code int, format string, a ...any) {
	fmt.Fprintf(os.Stderr, "m8_rename: "+format+"\n", a...)
	os.Exit(code)
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var skipDirs = map[string]bool{".git": true, ".moai": true, ".claude": true, "node_modules": true, "vendor": true, "testdata": true}

// walkGo lists the repository-relative paths of the Go files under internal and cmd.
func walkGo(root string) []string {
	var out []string
	for _, top := range []string{"internal", "cmd"} {
		_ = filepath.WalkDir(filepath.Join(root, top), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				if skipDirs[d.Name()] || rel == "internal/template/templates" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(p, ".go") {
				out = append(out, rel)
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// mapRel gives the repository-relative path a file has after the directory move.
func mapRel(rel string) string {
	if strings.HasPrefix(rel, oldDir+"/") {
		return newDir + "/" + strings.TrimPrefix(rel, oldDir+"/")
	}
	return rel
}

func isWordByte(c byte) bool {
	return c == '_' || c == '-' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// pathSpellings are the two spellings of the package directory this program rewrites: the module-relative
// path and the sibling-relative path a test of a neighbouring package uses to reach the directory.
var pathSpellings = [][2]string{{oldDir, newDir}, {"../kanban", "../factory"}}

// rewritePath rewrites each spelling of the package path in s (internal/kanban to internal/factory, ../kanban
// to ../factory). A spelling counts when the character before it is not a word character or a colon (a colon
// marks a git revision spelled rev:path) and the character after it is not a word character.
func rewritePath(s string) (string, int) {
	total := 0
	for _, sp := range pathSpellings {
		var sb strings.Builder
		n, last := 0, 0
		for i := 0; ; {
			j := strings.Index(s[i:], sp[0])
			if j < 0 {
				break
			}
			a := i + j
			b := a + len(sp[0])
			i = b
			if a > 0 && (isWordByte(s[a-1]) || s[a-1] == ':') {
				continue
			}
			if b < len(s) && isWordByte(s[b]) {
				continue
			}
			sb.WriteString(s[last:a])
			sb.WriteString(sp[1])
			last = b
			n++
		}
		if n > 0 {
			sb.WriteString(s[last:])
			s = sb.String()
			total += n
		}
	}
	return s, total
}

// rewriteQualified rewrites each spelling of a package-qualified exported name, kanban.Name, in s to
// factory.Name: the text a source-scanning test searches for, a reflect type name, a message that names a
// symbol. The identifier after the dot must start with an upper-case letter (an i18n key such as
// kanban.noSession and a file name such as kanban.go stay) and the character before kanban must not be a
// word character.
func rewriteQualified(s string) (string, int) {
	var sb strings.Builder
	n, last := 0, 0
	for i := 0; ; {
		j := strings.Index(s[i:], "kanban.")
		if j < 0 {
			break
		}
		a := i + j
		b := a + len("kanban.")
		i = b
		if a > 0 && (isWordByte(s[a-1]) || s[a-1] == '.') {
			continue
		}
		if b >= len(s) || s[b] < 'A' || s[b] > 'Z' {
			continue
		}
		sb.WriteString(s[last:a])
		sb.WriteString("factory.")
		last = b
		n++
	}
	if n == 0 {
		return s, 0
	}
	sb.WriteString(s[last:])
	return sb.String(), n
}

func applyEdits(src []byte, es []edit) []byte {
	sort.Slice(es, func(i, j int) bool { return es[i].a > es[j].a })
	out := append([]byte(nil), src...)
	last := -1
	for _, e := range es {
		if e.a == last { // an identifier visited twice yields one edit
			continue
		}
		last = e.a
		out = append(out[:e.a], append([]byte(e.s), out[e.b:]...)...)
	}
	return out
}

func main() {
	rootFlag := flag.String("root", "", "git tree to rewrite")
	dry := flag.Bool("dry-run", false, "report what would change and write nothing")
	flag.BoolVar(&verbose, "v", false, "list every string-literal and comment rewrite")
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

	st := &stats{files: map[string]bool{}}
	rels := walkGo(root) // pre-move paths

	// step 1: the directory move (planned first so a collision refuses before anything is written)
	type mv struct {
		from, to string
		tracked  bool
	}
	var moves []mv
	if _, err := os.Stat(filepath.Join(root, oldDir)); err == nil {
		tracked := map[string]bool{}
		if out, err := git(root, "ls-files", "--", oldDir); err == nil {
			for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
				if l != "" {
					tracked[l] = true
				}
			}
		}
		_ = filepath.WalkDir(filepath.Join(root, oldDir), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			moves = append(moves, mv{rel, mapRel(rel), tracked[rel]})
			return nil
		})
		for _, m := range moves {
			if _, err := os.Stat(filepath.Join(root, m.to)); err == nil {
				fatal(1, "move collision: %s already exists (would replace %s); nothing written", m.to, m.from)
			}
		}
	}
	st.moved = len(moves)
	if !*dry {
		for _, m := range moves {
			if err := os.MkdirAll(filepath.Dir(filepath.Join(root, m.to)), 0o755); err != nil {
				fatal(1, "mkdir for %s: %v", m.to, err)
			}
			if m.tracked {
				if o, err := git(root, "mv", m.from, m.to); err != nil {
					fatal(1, "git mv %s %s: %v\n%s", m.from, m.to, err, o)
				}
			} else if err := os.Rename(filepath.Join(root, m.from), filepath.Join(root, m.to)); err != nil {
				fatal(1, "rename %s: %v", m.from, err)
			}
		}
		if len(moves) > 0 {
			_ = os.Remove(filepath.Join(root, oldDir)) // only if empty
		}
	}

	// step 2: rewrite
	type result struct {
		rel string
		out []byte
	}
	var results []result
	importerDirs := map[string]bool{}
	for _, rel := range rels {
		newRel := mapRel(rel)
		readRel := newRel
		if *dry {
			readRel = rel
		}
		src, err := os.ReadFile(filepath.Join(root, readRel))
		if err != nil {
			fatal(1, "read %s: %v", readRel, err)
		}
		if !bytes.Contains(src, []byte("kanban")) {
			continue
		}
		out, changed := rewriteFile(newRel, src, st)
		if changed {
			results = append(results, result{newRel, out})
			st.files[newRel] = true
		}
		if bytes.Contains(src, []byte(`"`+oldImport+`"`)) {
			importerDirs[filepath.Dir(newRel)] = true
		}
	}

	// the new import name must not clash with a package-level declaration of an importing package
	for dir := range importerDirs {
		ents, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		for _, e := range ents {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			p := filepath.Join(root, filepath.FromSlash(dir), e.Name())
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			f, err := parser.ParseFile(token.NewFileSet(), p, b, parser.SkipObjectResolution)
			if err != nil {
				continue
			}
			for _, d := range f.Decls {
				for _, n := range declNames(d) {
					if n == "factory" {
						fatal(1, "%s declares a package-level name factory that would clash with the new import name; nothing written", filepath.ToSlash(filepath.Join(dir, e.Name())))
					}
				}
			}
		}
	}

	if !*dry {
		for _, r := range results {
			if err := os.WriteFile(filepath.Join(root, r.rel), r.out, 0o644); err != nil {
				fatal(1, "write %s: %v", r.rel, err)
			}
		}
	}
	mode := ""
	if *dry {
		mode = " (dry run, nothing written)"
	}
	fmt.Printf("m8_rename: package clauses: %d; import lines: %d; qualifiers: %d; shadowing locals renamed: %d; string literals: %d; comment mentions: %d; files rewritten: %d; files moved: %d%s\n",
		st.clauses, st.importLines, st.qualifiers, st.shadows, st.literals, st.comments, len(st.files), st.moved, mode)
}

func declNames(d ast.Decl) []string {
	var out []string
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
	return out
}

// rewriteFile applies the step 2 rules to one file and returns the gofmt-ed result and whether it changed.
func rewriteFile(rel string, src []byte, st *stats) ([]byte, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, rel, src, parser.ParseComments)
	if err != nil {
		fatal(1, "parse %s: %v; nothing written", rel, err)
	}
	off := func(p token.Pos) int { return fset.Position(p).Offset }
	var es []edit
	inFactory := strings.HasPrefix(rel, newDir+"/")

	// a. package clause
	if inFactory && (f.Name.Name == "kanban" || f.Name.Name == "kanban_test") {
		a := off(f.Name.Pos())
		es = append(es, edit{a, a + len(f.Name.Name), strings.Replace(f.Name.Name, "kanban", "factory", 1)})
		st.clauses++
	}

	// b. import path, and whether the file refers to the package by its default name
	importsOld := false
	importSpecLits := map[token.Pos]bool{}
	for _, im := range f.Imports {
		importSpecLits[im.Path.Pos()] = true
		if strings.Trim(im.Path.Value, "`\"") != oldImport {
			continue
		}
		a := off(im.Path.Pos())
		es = append(es, edit{a, a + len(im.Path.Value), `"` + newImport + `"`})
		st.importLines++
		if im.Name == nil {
			importsOld = true
		}
	}
	if importsOld {
		for _, im := range f.Imports {
			p := strings.Trim(im.Path.Value, "`\"")
			name := filepath.Base(p)
			if im.Name != nil {
				name = im.Name.Name
			}
			if p != oldImport && name == "factory" {
				fatal(1, "%s imports another package named factory next to the old path; nothing written", rel)
			}
		}
	}

	// the parameters and results of function types are the declarations a shadowing rename may take
	paramFields := map[*ast.Field]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if ft, ok := n.(*ast.FuncType); ok {
			for _, fl := range []*ast.FieldList{ft.TypeParams, ft.Params, ft.Results} {
				if fl != nil {
					for _, fld := range fl.List {
						paramFields[fld] = true
					}
				}
			}
		}
		return true
	})
	existing := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			existing[id.Name] = true
		}
		return true
	})

	denied := literalDeny[rel]
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			// c. qualifiers of an unaliased import
			if id, ok := x.X.(*ast.Ident); ok && importsOld && id.Name == "kanban" && id.Obj == nil {
				a := off(id.Pos())
				es = append(es, edit{a, a + len("kanban"), "factory"})
				st.qualifiers++
			}
		case *ast.Ident:
			// d. a local named factory shadows the renamed qualifier
			if importsOld && x.Name == "factory" && x.Obj != nil {
				if f.Scope != nil && f.Scope.Objects["factory"] == x.Obj {
					fatal(1, "%s declares a package-level factory next to the old import; nothing written", rel)
				}
				local := false
				switch d := x.Obj.Decl.(type) {
				case *ast.AssignStmt, *ast.ValueSpec, *ast.RangeStmt, *ast.TypeSpec:
					local = true
				case *ast.Field:
					local = paramFields[d]
				}
				if local {
					if existing["factoryRun"] {
						fatal(1, "%s already uses the name factoryRun, the rename target of a shadowing local; nothing written", rel)
					}
					a := off(x.Pos())
					es = append(es, edit{a, a + len("factory"), "factoryRun"})
					st.shadows++
				}
			}
		case *ast.CallExpr:
			// e. two adjacent literals of a path join: "internal", "kanban"
			if denied {
				return true
			}
			for i := 0; i+1 < len(x.Args); i++ {
				l1, ok1 := x.Args[i].(*ast.BasicLit)
				l2, ok2 := x.Args[i+1].(*ast.BasicLit)
				if ok1 && ok2 && l1.Kind == token.STRING && l2.Kind == token.STRING && l1.Value == `"internal"` && l2.Value == `"kanban"` {
					a := off(l2.Pos())
					es = append(es, edit{a, a + len(l2.Value), `"factory"`})
					st.literals++
					note("literal", rel, fset.Position(l2.Pos()).Line, l2.Value, `"factory"`)
				}
			}
		case *ast.BasicLit:
			// e. string literals that spell the package path, and the home-state coverage key
			if x.Kind != token.STRING || importSpecLits[x.Pos()] || denied {
				return true
			}
			a := off(x.Pos())
			if rel == "internal/cli/home_state_coverage.go" && x.Value == `"kanban"` {
				es = append(es, edit{a, a + len(x.Value), `"factory"`})
				st.literals++
				note("literal", rel, fset.Position(x.Pos()).Line, x.Value, `"factory"`)
				return true
			}
			nv, n1 := rewritePath(x.Value)
			nv, n2 := rewriteQualified(nv)
			if n1+n2 > 0 {
				es = append(es, edit{a, a + len(x.Value), nv})
				st.literals++
				note("literal", rel, fset.Position(x.Pos()).Line, x.Value, nv)
			}
		}
		return true
	})

	// f. comments that name the package path, and the package doc line
	if !denied {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				nv, n := rewritePath(c.Text)
				nv, n2 := rewriteQualified(nv)
				n += n2
				if inFactory && cg == f.Doc && strings.HasPrefix(nv, "// Package kanban ") {
					nv = "// Package factory " + strings.TrimPrefix(nv, "// Package kanban ")
					n++
				}
				if nv != c.Text {
					a := off(c.Slash)
					es = append(es, edit{a, a + len(c.Text), nv})
					st.comments += n
					note("comment", rel, fset.Position(c.Slash).Line, c.Text, nv)
				}
			}
		}
	}

	if len(es) == 0 {
		return src, false
	}
	out := applyEdits(src, es)
	fm, err := format.Source(out)
	if err != nil {
		fatal(1, "gofmt %s: %v; nothing written", rel, err)
	}
	return fm, !bytes.Equal(fm, src)
}
