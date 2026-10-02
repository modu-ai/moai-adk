//go:build ignore

// probe.go replays the milestone sequence of SPEC-LAUNCHER-ENTRY-FLAGS-001 on a SCRATCH COPY of the Go
// tree and compiles after every stage. It never touches the tree it copies from (rsync reads -src only).
//
//	go run .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe/probe.go \
//	    -src <repo checkout> -work <scratch dir> -data .moai/specs/SPEC-LAUNCHER-ENTRY-FLAGS-001/probe \
//	    [-from M0] [-to M11] [-vet=false]
//
// The checkout's Go sources must equal tree a6d3e6fd4f21f9c04fbcb7ca7507e87571f9b2c2 (the SPEC directory
// itself is not copied): the stage data is a diff against that tree. Needs rsync, patch, and go on PATH,
// plus the templ module in the module cache for stage M9.
//
// Stages, in order: M0 and M1 and M11 change no Go source; M2, M3, M4, M5a, M5b, M6a, M6b are recorded
// data under <data>/patches (<stage>.rm lists removed paths, <stage>.patch is a unified diff of modified
// and added files); M7, M8, M9, M10 are programmatic renames below. M6a and M6b are the two commits inside
// milestone M6 (the lock re-home, then the deletions). After every stage four checks print a labeled block
// followed by its exit code: go build -gcflags=-e ./... and go vet ./... on the host OS, and the same two with
// GOOS=windows GOARCH=amd64. An empty block with exit 0 is a clean stage; the windows vet block is read
// against the one failure that predates the SPEC (internal/cli/worktree/sweep_test.go:1687, parseLsofCWDs).
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

var stageOrder = []string{"M0", "M1", "M2", "M3", "M4", "M5a", "M5b", "M6a", "M6b", "M7", "M8", "M9", "M10", "M11"}

var (
	src   = flag.String("src", "", "repository checkout to copy the Go tree from (read only)")
	work  = flag.String("work", "", "scratch directory (created; removed first if present)")
	from  = flag.String("from", "M0", "first stage to apply")
	to    = flag.String("to", "M11", "last stage to apply")
	vet   = flag.Bool("vet", true, "also run go vet ./... on both OSes")
	data  = flag.String("data", "", "directory holding patches/ (default: directory of this file)")
	tree  string
	patch string
)

func main() {
	flag.Parse()
	if *src == "" || *work == "" {
		fmt.Fprintln(os.Stderr, "need -src and -work")
		os.Exit(2)
	}
	if *data == "" {
		*data = "."
	}
	patch = filepath.Join(*data, "patches")
	tree = filepath.Join(*work, "tree")
	must(os.RemoveAll(*work))
	must(os.MkdirAll(*work, 0o755))
	run(*src, nil, "rsync", "-a",
		*src+"/go.mod", *src+"/go.sum", *src+"/cmd", *src+"/e2e", *src+"/internal", *src+"/pkg", *src+"/scripts", *src+"/test", tree+"/")
	started := false
	for _, st := range stageOrder {
		if st == *from {
			started = true
		}
		if !started {
			continue
		}
		fmt.Printf("##### stage %s\n", st)
		apply(st)
		check(st)
		if st == *to {
			break
		}
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

// run executes a command, returning combined output and exit code.
func run(dir string, env []string, name string, args ...string) (string, int) {
	cmd := exec.Command(name, args...)
	if dir != "" && name != "rsync" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 127
		}
	}
	return string(out), code
}

func apply(st string) {
	switch st {
	case "M0", "M1", "M11":
		fmt.Printf("(no Go change: %s)\n", map[string]string{
			"M0":  "no code, baseline only",
			"M1":  "adds test files only; the additions are not authored by this probe",
			"M11": "documentation only",
		}[st])
	case "M7":
		stageM7()
	case "M8":
		stageM8()
	case "M9":
		stageM9()
	case "M10":
		stageM10()
	default:
		dataStage(st)
	}
}

// dataStage applies a recorded stage: removals first, then the unified diff.
func dataStage(st string) {
	if b, err := os.ReadFile(filepath.Join(patch, st+".rm")); err == nil {
		for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			if l != "" {
				must(os.Remove(filepath.Join(tree, l)))
			}
		}
	}
	p := filepath.Join(patch, st+".patch")
	abs, _ := filepath.Abs(p)
	out, code := run(tree, nil, "patch", "-p1", "-s", "-i", abs)
	if code != 0 {
		fmt.Printf("PATCH FAILED %s (exit %d)\n%s", st, code, out)
		os.Exit(1)
	}
	_ = filepath.WalkDir(tree, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(p, ".orig") || strings.HasSuffix(p, ".rej")) {
			os.Remove(p)
		}
		return nil
	})
}

func check(st string) {
	t := tree
	show := func(label, out string, code int) {
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) > 40 {
			lines = append(lines[:40], fmt.Sprintf("... (%d more lines)", len(lines)-40))
		}
		fmt.Printf("=== [%s] %s\n%s\n--- exit %d\n", st, label, strings.Join(lines, "\n"), code)
	}
	out, c := run(t, nil, "go", "build", "-gcflags=-e", "./...")
	show("go build -gcflags=-e ./...  (darwin/host)", out, c)
	out, c = run(t, []string{"GOOS=windows", "GOARCH=amd64"}, "go", "build", "-gcflags=-e", "./...")
	show("GOOS=windows GOARCH=amd64 go build -gcflags=-e ./...", out, c)
	if *vet {
		out, c = run(t, nil, "go", "vet", "./...")
		show("go vet ./...  (darwin/host; typechecks every test file)", out, c)
		out, c = run(t, []string{"GOOS=windows", "GOARCH=amd64"}, "go", "vet", "./...")
		show("GOOS=windows GOARCH=amd64 go vet ./...", out, c)
	}
}

// ---------------------------------------------------------------- go source helpers

type span struct{ a, b int }

func goFiles() []string {
	var out []string
	_ = filepath.WalkDir(tree, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".go") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out
}

// renameIdents rewrites identifier tokens: explicit holds exact old->new; generic maps any identifier that
// contains "kanban" (any case), other than the bare package name, by replacing the word.
var lastPairs = map[string]string{}

func renameIdents(explicit map[string]string, generic bool, skipFile func(string) bool) int {
	n := 0
	lastPairs = map[string]string{}
	for _, p := range goFiles() {
		if skipFile != nil && skipFile(p) {
			continue
		}
		srcb, err := os.ReadFile(p)
		must(err)
		if !bytes.Contains(bytes.ToLower(srcb), []byte("kanban")) && len(explicit) == 0 {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, p, srcb, parser.ParseComments)
		if err != nil {
			fmt.Printf("parse error %s: %v\n", p, err)
			continue
		}
		type rep struct {
			span
			s string
		}
		var reps []rep
		ast.Inspect(f, func(x ast.Node) bool {
			id, ok := x.(*ast.Ident)
			if !ok {
				return true
			}
			nn, hit := explicit[id.Name]
			if !hit && generic && id.Name != "kanban" && strings.Contains(strings.ToLower(id.Name), "kanban") {
				nn = strings.NewReplacer("Kanban", "Factory", "kanban", "factory", "KANBAN", "FACTORY").Replace(id.Name)
				hit = true
			}
			if hit && nn != id.Name {
				lastPairs[id.Name] = nn
				a := fset.Position(id.Pos()).Offset
				reps = append(reps, rep{span{a, a + len(id.Name)}, nn})
			}
			return true
		})
		if len(reps) == 0 {
			continue
		}
		sort.Slice(reps, func(i, j int) bool { return reps[i].a > reps[j].a })
		out := append([]byte(nil), srcb...)
		for _, r := range reps {
			out = append(out[:r.a], append([]byte(r.s), out[r.b:]...)...)
		}
		if fm, err := format.Source(out); err == nil {
			out = fm
		}
		must(os.WriteFile(p, out, 0o644))
		n += len(reps)
	}
	return n
}

// deleteDecls removes top-level declarations by name from one file: funcs (no receiver), types,
// and const/var specs.
func deleteDecls(rel string, names ...string) { cutDecls(rel, names, nil) }

// deleteFuncsReferencing removes every top-level func in rel whose body mentions any given identifier.
func deleteFuncsReferencing(rel string, idents ...string) { cutDecls(rel, nil, idents) }

func cutDecls(rel string, names, refs []string) {
	p := filepath.Join(tree, rel)
	srcb, err := os.ReadFile(p)
	must(err)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, p, srcb, parser.ParseComments)
	must(err)
	del := map[string]bool{}
	for _, n := range names {
		del[n] = true
	}
	rf := map[string]bool{}
	for _, n := range refs {
		rf[n] = true
	}
	mentions := func(n ast.Node) bool {
		hit := false
		ast.Inspect(n, func(x ast.Node) bool {
			if id, ok := x.(*ast.Ident); ok && rf[id.Name] {
				hit = true
			}
			return !hit
		})
		return hit
	}
	var spans []span
	cut := func(n ast.Node, doc *ast.CommentGroup) {
		a := fset.Position(n.Pos()).Offset
		if doc != nil {
			a = fset.Position(doc.Pos()).Offset
		}
		spans = append(spans, span{a, fset.Position(n.End()).Offset})
	}
	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.FuncDecl:
			if (x.Recv == nil && del[x.Name.Name]) || (len(rf) > 0 && mentions(x)) {
				cut(x, x.Doc)
			}
		case *ast.GenDecl:
			if x.Tok == token.IMPORT {
				continue
			}
			var gone, keep []ast.Spec
			for _, sp := range x.Specs {
				m := false
				switch t := sp.(type) {
				case *ast.TypeSpec:
					m = del[t.Name.Name]
				case *ast.ValueSpec:
					for _, nm := range t.Names {
						if del[nm.Name] {
							m = true
						}
					}
				}
				if m {
					gone = append(gone, sp)
				} else {
					keep = append(keep, sp)
				}
			}
			if len(gone) == 0 {
				continue
			}
			if len(keep) == 0 {
				cut(x, x.Doc)
				continue
			}
			for _, sp := range gone {
				cut(sp, nil)
			}
		}
	}
	if len(spans) == 0 {
		return
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].a > spans[j].a })
	out := append([]byte(nil), srcb...)
	for _, sp := range spans {
		out = append(out[:sp.a], out[sp.b:]...)
	}
	if fm, err := format.Source(out); err == nil {
		out = fm
	}
	must(os.WriteFile(p, out, 0o644))
	pruneImports(p)
}

// pruneImports drops import specs whose local name is no longer referenced as a selector qualifier
// (a stand-in for goimports: the name is the alias, else the last path element).
func pruneImports(p string) {
	b, err := os.ReadFile(p)
	must(err)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, p, b, parser.ParseComments)
	if err != nil {
		return
	}
	used := map[string]bool{}
	ast.Inspect(f, func(x ast.Node) bool {
		if se, ok := x.(*ast.SelectorExpr); ok {
			if id, ok := se.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	var spans []span
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.IMPORT {
			continue
		}
		unused := 0
		for _, sp := range gd.Specs {
			if !importUsed(sp.(*ast.ImportSpec), used) {
				unused++
			}
		}
		if unused == len(gd.Specs) && unused > 0 {
			spans = append(spans, span{fset.Position(gd.Pos()).Offset, fset.Position(gd.End()).Offset})
			for _, sp := range gd.Specs {
				used["\x00drop:"+sp.(*ast.ImportSpec).Path.Value] = true
			}
		}
	}
	for _, im := range f.Imports {
		if used["\x00drop:"+im.Path.Value] {
			continue
		}
		path := strings.Trim(im.Path.Value, `"`)
		name := ""
		if im.Name != nil {
			name = im.Name.Name
		} else {
			el := strings.Split(path, "/")
			name = el[len(el)-1]
			if regexp.MustCompile(`^v[0-9]+$`).MatchString(name) && len(el) > 1 {
				name = el[len(el)-2]
			}
		}
		if name == "_" || name == "." || used[name] {
			continue
		}
		a := fset.Position(im.Pos()).Offset
		spans = append(spans, span{a, fset.Position(im.End()).Offset})
	}
	if len(spans) == 0 {
		return
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].a > spans[j].a })
	out := append([]byte(nil), b...)
	for _, sp := range spans {
		out = append(out[:sp.a], out[sp.b:]...)
	}
	if fm, err := format.Source(out); err == nil {
		out = fm
	}
	must(os.WriteFile(p, out, 0o644))
}

func importUsed(im *ast.ImportSpec, used map[string]bool) bool {
	path := strings.Trim(im.Path.Value, `"`)
	name := ""
	if im.Name != nil {
		name = im.Name.Name
	} else {
		el := strings.Split(path, "/")
		name = el[len(el)-1]
		if regexp.MustCompile(`^v[0-9]+$`).MatchString(name) && len(el) > 1 {
			name = el[len(el)-2]
		}
	}
	return name == "_" || name == "." || used[name]
}

func regexSub(rel, pat, repl string) {
	p := filepath.Join(tree, rel)
	b, err := os.ReadFile(p)
	must(err)
	re := regexp.MustCompile("(?s)" + pat)
	if !re.Match(b) {
		fmt.Printf("WARN regexSub: %s does not match %q\n", rel, pat)
		return
	}
	must(os.WriteFile(p, re.ReplaceAll(b, []byte(repl)), 0o644))
}

func literalSub(rel, old, nw string) {
	p := filepath.Join(tree, rel)
	b, err := os.ReadFile(p)
	must(err)
	if !bytes.Contains(b, []byte(old)) {
		fmt.Printf("WARN literalSub: %s does not contain %q\n", rel, old)
		return
	}
	must(os.WriteFile(p, bytes.ReplaceAll(b, []byte(old), []byte(nw)), 0o644))
}

func move(oldRel, newRel string) {
	must(os.MkdirAll(filepath.Dir(filepath.Join(tree, newRel)), 0o755))
	must(os.Rename(filepath.Join(tree, oldRel), filepath.Join(tree, newRel)))
}

var _ = regexp.MustCompile

func exists(rel string) bool {
	_, err := os.Stat(filepath.Join(tree, rel))
	return err == nil
}

// moveChecked refuses to overwrite.
func moveChecked(oldRel, newRel string) {
	if exists(newRel) {
		fmt.Printf("COLLISION: %s already exists; not moving %s\n", newRel, oldRel)
		return
	}
	move(oldRel, newRel)
}

// ---------------------------------------------------------------- M7: identifiers and files inside packages

func stageM7() {
	// the entry-parse type loses the kanban field; assertions on it are re-pinned
	literalSub("internal/cli/kanban.go", "KanbanEnabled  bool   // -k present (any shape)\n", "")
	literalSub("internal/cli/factory_test.go", " || !p.KanbanEnabled", "")
	literalSub("internal/cli/factory_test.go", " || p.KanbanEnabled", "")
	// the two launch-facts functions collapse into one
	deleteDecls("internal/cli/kanban.go", "exportFactoryLaunchFacts")
	explicit := map[string]string{
		"EnvMoaiKanbanID":               "EnvFactoryRunID",
		"EnvMoaiKanbanLeadAddr":         "EnvFactoryLeadAddr",
		"EnvMoaiKanbanLeadName":         "EnvFactoryLeadName",
		"EnvMoaiKanbanSettingsInjected": "EnvFactorySettingsInjected",
		"EnvMoaiKanbanBackend":          "EnvFactoryBackend",
		"EnvMoaiKanbanCard":             "EnvFactoryCard",
		"kanbanEntryParse":              "launcherEntryParse",
	}
	n := renameIdents(explicit, true, func(p string) bool { return strings.Contains(p, "/internal/web/") })
	fmt.Printf("identifiers renamed: %d\n", n)
	// files carrying the word in their name (inside internal/cli and internal/kanban)
	moveChecked("internal/cli/kanban.go", "internal/cli/factory_launch_helpers.go")
	moveChecked("internal/cli/kanban_dispatch_test.go", "internal/cli/factory_launch_dispatch_test.go")
	for _, dir := range []string{"internal/cli", "internal/kanban"} {
		ents, _ := os.ReadDir(filepath.Join(tree, dir))
		for _, e := range ents {
			if !e.IsDir() && strings.Contains(e.Name(), "kanban") && strings.HasSuffix(e.Name(), ".go") {
				moveChecked(filepath.Join(dir, e.Name()), filepath.Join(dir, strings.Replace(e.Name(), "kanban", "factory", 1)))
			}
		}
	}
}

// ---------------------------------------------------------------- M8: package rename

const oldImport = "github.com/modu-ai/moai-adk/internal/kanban"
const newImport = "github.com/modu-ai/moai-adk/internal/factory"

func stageM8() {
	move("internal/kanban", "internal/factory")
	files := goFiles()
	qual, shadow, pkgDecl := 0, 0, 0
	for _, p := range files {
		b, err := os.ReadFile(p)
		must(err)
		inPkg := strings.HasPrefix(p, filepath.Join(tree, "internal", "factory")+string(os.PathSeparator)) &&
			filepath.Dir(p) == filepath.Join(tree, "internal", "factory")
		hasImport := bytes.Contains(b, []byte(`"`+oldImport+`"`))
		if !inPkg && !hasImport {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, p, b, parser.ParseComments)
		if err != nil {
			fmt.Printf("parse error %s: %v\n", p, err)
			continue
		}
		type rep struct {
			span
			s string
		}
		var reps []rep
		add := func(n ast.Node, old, nw string) {
			a := fset.Position(n.Pos()).Offset
			reps = append(reps, rep{span{a, a + len(old)}, nw})
		}
		if f.Name.Name == "kanban" || f.Name.Name == "kanban_test" {
			add(f.Name, f.Name.Name, strings.Replace(f.Name.Name, "kanban", "factory", 1))
			pkgDecl++
		}
		importsOld := false
		for _, im := range f.Imports {
			if strings.Trim(im.Path.Value, `"`) == oldImport && im.Name == nil {
				importsOld = true
				a := fset.Position(im.Path.Pos()).Offset
				reps = append(reps, rep{span{a, a + len(im.Path.Value)}, `"` + newImport + `"`})
			}
			if strings.Trim(im.Path.Value, `"`) == oldImport && im.Name != nil {
				a := fset.Position(im.Path.Pos()).Offset
				reps = append(reps, rep{span{a, a + len(im.Path.Value)}, `"` + newImport + `"`})
			}
		}
		if importsOld {
			ast.Inspect(f, func(x ast.Node) bool {
				switch n := x.(type) {
				case *ast.SelectorExpr:
					if id, ok := n.X.(*ast.Ident); ok && id.Name == "kanban" && id.Obj == nil {
						add(id, "kanban", "factory")
						qual++
					}
				case *ast.Ident:
					// a local declaration named factory would shadow the renamed package qualifier
					if n.Name == "factory" && n.Obj != nil {
						add(n, "factory", "factoryRun")
						shadow++
					}
				}
				return true
			})
		}
		if len(reps) == 0 {
			continue
		}
		sort.Slice(reps, func(i, j int) bool { return reps[i].a > reps[j].a })
		// de-duplicate identical spans (an ident may be visited as selector X and as Ident)
		out := append([]byte(nil), b...)
		last := -1
		for _, r := range reps {
			if r.a == last {
				continue
			}
			last = r.a
			out = append(out[:r.a], append([]byte(r.s), out[r.b:]...)...)
		}
		if fm, err := format.Source(out); err == nil {
			out = fm
		}
		must(os.WriteFile(p, out, 0o644))
	}
	fmt.Printf("package clauses renamed: %d, qualifiers renamed: %d, shadowing locals renamed: %d\n", pkgDecl, qual, shadow)
}
// ---------------------------------------------------------------- M9: web console

func stageM9() {
	web := func(p string) bool { return !strings.Contains(p, "/internal/web/") }
	n := renameIdents(nil, true, web)
	pairs := lastPairs
	fmt.Printf("web identifiers renamed: %d (%d distinct)\n", n, len(pairs))
	// the .templ sources carry the same identifiers
	ents, _ := os.ReadDir(filepath.Join(tree, "internal", "web"))
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".templ") {
			continue
		}
		p := filepath.Join(tree, "internal", "web", e.Name())
		b, err := os.ReadFile(p)
		must(err)
		t := string(b)
		for old, nw := range pairs {
			t = regexp.MustCompile(`\b`+old+`\b`).ReplaceAllString(t, nw)
		}
		must(os.WriteFile(p, []byte(t), 0o644))
	}
	// the chain session board panel leaves the screen
	sp := filepath.Join(tree, "internal", "web", "screens.templ")
	b, err := os.ReadFile(sp)
	must(err)
	t := string(b)
	i := strings.Index(t, "Chain session board")
	if i < 0 {
		fmt.Println("WARN: chain panel marker not found")
	} else {
		start := strings.LastIndex(t[:i], "\t\t\t<div class=\"panel__head\"")
		e := strings.Index(t[i:], "\"kanban.note\")")
		if start < 0 || e < 0 {
			fmt.Println("WARN: chain panel bounds not found")
		} else {
			end := i + e + len("\"kanban.note\")")
			end += strings.Index(t[end:], "\n") + 1
			t = t[:start] + t[end:]
		}
	}
	must(os.WriteFile(sp, []byte(t), 0o644))
	// regenerate the *_templ.go files exactly as `make templ-generate` does
	out, code := run(tree, nil, "go", "run", "github.com/a-h/templ/cmd/templ", "generate", "-path", "./internal/web")
	fmt.Printf("templ generate exit %d\n%s", code, tail(out, 6))
	// the chain view model leaves with the panel: the types, builders, the Overview field and the
	// idle-role attention row that read it, and the tests that exercise them
	const vm = "internal/web/viewmodel_ops.go"
	deleteDecls(vm, "ChainVM", "RoleVM", "buildChain", "chainRoleRecords", "chainCardID", "ChainRoles")
	regexSub(vm, `\tChain\s+ChainVM\n`, ``)
	regexSub(vm, `\tCardID\s+string\n\tIdleRole string\n\tRoles\s+\[\]RoleVM\n`, ``)
	regexSub(vm, `\t\tChain:\s+buildChain\(root, records, byID, chainCardID\(records\)\),\n`, ``)
	regexSub(vm, `vm\.Attention = buildAttention\(rows, findings, vm\.Chain\)`, `vm.Attention = buildAttention(rows, findings)`)
	regexSub(vm, `func buildAttention\(rows \[\]SpecRowVM, findings map\[string\]\[\]FindingVM, chain ChainVM\) \[\]AttentionVM \{\n\tvar out \[\]AttentionVM\n\tif chain\.Present && chain\.IdleRole != "" \{.*?\n\t\}\n\tfor _, r := range rows \{`, "func buildAttention(rows []SpecRowVM, findings map[string][]FindingVM) []AttentionVM {\n\tvar out []AttentionVM\n\tfor _, r := range rows {")
	regexSub(vm, `\t// The chain is built from chain-role records only\..*?\tchain := buildChain\(root, chainRecords, byID, chainCardID\(chainRecords\)\)\n\n`, ``)
	regexSub(vm, `\t\tCardID:\s+chain\.CardID,\n\t\tIdleRole: chain\.IdleRole,\n\t\tRoles:\s+chain\.Roles,\n`, ``)
	regexSub(vm, `\tsessions, byID := loadSessions\(root, now\)\n\trecords := loadFactoryRecords\(root\)\n\n\tvar inProgress`, "\tsessions, _ := loadSessions(root, now)\n\n\tvar inProgress")
	for _, tf := range []string{"screen_states_test.go", "session_telemetry_cells_test.go", "factory_lane_identity_test.go",
		"role_naming_m3_legacy_leader_test.go", "viewmodel_ops_edges_test.go"} {
		deleteFuncsReferencing("internal/web/"+tf, "RoleVM", "buildChain", "chainRoleRecords", "ChainVM", "chainCardID", "ChainRoles")
	}
	// the redirect-only legacy route (the fourth file allowed to carry the word)
	must(os.WriteFile(filepath.Join(tree, "internal", "web", "legacy_routes.go"), []byte(`package web

import "net/http"

// registerLegacyRoutes keeps the retired screen path alive as a redirect only.
func registerLegacyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/kanban", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/factory", http.StatusMovedPermanently)
	})
}
`), 0o644))
	literalSub("internal/web/app.go", `mux.HandleFunc("/kanban", a.handleFactory)`, `mux.HandleFunc("/factory", a.handleFactory)
	registerLegacyRoutes(mux)`)
}

func tail(s string, n int) string {
	l := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(l) > n {
		l = l[len(l)-n:]
	}
	return strings.Join(l, "\n") + "\n"
}
// ---------------------------------------------------------------- M10: rules, skills, catalog

func stageM10() {
	rules := "internal/template/templates/.claude/rules/moai/workflow/"
	for _, suf := range []string{"", "-detail", "-mechanics"} {
		moveChecked(rules+"kanban-dispatch"+suf+".md", rules+"factory-dispatch"+suf+".md")
	}
	moveChecked("internal/template/templates/.claude/skills/moai-kanban-foreman", "internal/template/templates/.claude/skills/moai-factory-foreman")
	literalSub("internal/template/catalog.yaml", "moai-kanban-foreman", "moai-factory-foreman")
	literalSub("internal/cli/update_archive.go", "var legacySkillIDs = []string{\n", "var legacySkillIDs = []string{\n\t\"moai-kanban-foreman\",\n")
}
