package testguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Violation kinds reported by ScanSource.
const (
	// KindUserHomeDir marks a selector expression os.UserHomeDir (a call or a
	// function value), which reads the real home directory.
	KindUserHomeDir = "os.UserHomeDir"
	// KindTildeLiteral marks a string literal that contains "~/". The standard
	// library never expands it, so a test using it either creates a directory
	// named "~" under the working directory or relies on a home-expansion seam.
	KindTildeLiteral = "~/ literal"
	// KindIndirectGlobalGitConfig marks a top-level test function whose call
	// chain reaches an `exec.Command("git", "config", "--global", ...)` sink
	// without swapping the seam on that path, without redirecting the global git
	// config, and without an allowlist marker.
	KindIndirectGlobalGitConfig = "indirect git config --global"
)

// allowComment matches the allowlist marker `home-guard:allow <reason>` inside a
// line comment. A marker without a reason does not match, so it exempts nothing.
var allowComment = regexp.MustCompile(`//\s*home-guard:allow\s+\S`)

// Violation is one home-directory access in a test file that carries no
// allowlist comment on its own line or on the line directly above it.
type Violation struct {
	File string // path as scanned; relative to the scan root for ScanTree
	Line int    // 1-based line of the offending expression or literal
	Kind string // KindUserHomeDir, KindTildeLiteral or KindIndirectGlobalGitConfig
	Text string // the offending source line, trimmed
}

// Result is the outcome of ScanTree.
type Result struct {
	Files      int // test files parsed
	Violations []Violation
}

// Keys returns one stable key per violation, in the order of vs (ScanTree order:
// file, then line). A key is "<file>\t<kind>\t<normalized line text>\t<n>", where
// the line text has its whitespace runs collapsed to single spaces and n is the
// 0-based occurrence index of that (file, kind, text) triple within the file.
// The key holds no line number, so unrelated edits that shift lines leave it
// unchanged; only a changed or added access line changes the key set.
func Keys(vs []Violation) []string {
	seen := map[string]int{}
	keys := make([]string, 0, len(vs))
	for _, v := range vs {
		base := strings.ReplaceAll(v.File, `\`, "/") + "\t" + v.Kind + "\t" + strings.Join(strings.Fields(v.Text), " ")
		keys = append(keys, fmt.Sprintf("%s\t%d", base, seen[base]))
		seen[base]++
	}
	return keys
}

// ScanSource scans one Go source file for home-directory access. It returns an
// error only when src does not parse. The caller treats that as a failure, never
// as a clean file.
//
// The scan is AST-based, so a mention of os.UserHomeDir in a comment is not a
// violation, and a "~/" inside any string literal is. Only the package os
// reached through a local name bound to the os import is matched. A local value
// that shadows the name is not.
func ScanSource(filename string, src []byte) ([]Violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}

	// Only real comment tokens carry the marker; a string literal that contains
	// the same text is not a comment.
	markerLines := map[int]bool{}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if allowComment.MatchString(c.Text) {
				markerLines[fset.Position(c.Slash).Line] = true
			}
		}
	}

	lines := strings.Split(string(src), "\n")
	osNames := osImportNames(file)
	var out []Violation
	seen := map[string]bool{}
	report := func(line int, kind string) {
		key := fmt.Sprintf("%d:%s", line, kind)
		if seen[key] || markerLines[line] || markerLines[line-1] {
			return
		}
		seen[key] = true
		out = append(out, Violation{
			File: filename,
			Line: line,
			Kind: kind,
			Text: strings.TrimSpace(lineAt(lines, line)),
		})
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			id, ok := x.X.(*ast.Ident)
			if !ok || x.Sel.Name != "UserHomeDir" || !osNames[id.Name] || id.Obj != nil {
				return true
			}
			report(fset.Position(x.Pos()).Line, KindUserHomeDir)
		case *ast.BasicLit:
			if x.Kind != token.STRING {
				return true
			}
			// Report the line where "~/" sits inside the literal, not the line
			// where the literal starts, so a multi-line raw string is judged on
			// the line a reader would edit.
			start := fset.Position(x.Pos()).Line
			for from := 0; ; {
				i := strings.Index(x.Value[from:], "~/")
				if i < 0 {
					break
				}
				at := from + i
				report(start+strings.Count(x.Value[:at], "\n"), KindTildeLiteral)
				from = at + len("~/")
			}
		}
		return true
	})
	return out, nil
}

// osImportNames returns the local names under which file imports package os. A
// blank or dot import contributes no qualifier.
func osImportNames(file *ast.File) map[string]bool {
	names := map[string]bool{}
	for _, spec := range file.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil || p != "os" {
			continue
		}
		name := "os"
		if spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				continue
			}
			name = spec.Name.Name
		}
		names[name] = true
	}
	return names
}

// lineAt returns the 1-based line n of lines, or "" when n is out of range.
func lineAt(lines []string, n int) string {
	if n < 1 || n > len(lines) {
		return ""
	}
	return lines[n-1]
}

// ScanTree scans every *_test.go file below root/<dir> for each dir. The go tool
// skips testdata and vendor, and so does this walk. Symbolic links are not
// followed. Violations carry paths relative to root and are sorted for stable
// output. A file that does not parse fails the whole scan.
func ScanTree(root string, dirs []string) (Result, error) {
	var res Result
	// pkgs groups every .go file by its directory for the indirect check.
	pkgs := map[string]map[string][]byte{}
	for _, dir := range dirs {
		base := filepath.Join(root, dir)
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != base && (d.Name() == "testdata" || d.Name() == "vendor") {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() || !strings.HasSuffix(d.Name(), ".go") {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			pkg := pkgs[filepath.ToSlash(filepath.Dir(rel))]
			if pkg == nil {
				pkg = map[string][]byte{}
				pkgs[filepath.ToSlash(filepath.Dir(rel))] = pkg
			}
			pkg[rel] = src
			if !strings.HasSuffix(d.Name(), "_test.go") {
				return nil
			}
			vs, err := ScanSource(rel, src)
			if err != nil {
				return err
			}
			res.Files++
			res.Violations = append(res.Violations, vs...)
			return nil
		})
		if err != nil {
			return Result{}, fmt.Errorf("scan %s: %w", base, err)
		}
	}
	for dir, files := range pkgs {
		if !mentionsGlobalFlag(files) {
			continue
		}
		vs, err := scanIndirect(files)
		if err != nil {
			return Result{}, fmt.Errorf("scan %s: %w", dir, err)
		}
		res.Violations = append(res.Violations, vs...)
	}
	sort.SliceStable(res.Violations, func(i, j int) bool {
		a, b := res.Violations[i], res.Violations[j]
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	return res, nil
}
