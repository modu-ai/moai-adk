package effect

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

var writeNames = map[string]bool{`"POST"`: true, `"PATCH"`: true, `"PUT"`: true, `"DELETE"`: true, `"-X"`: true}

// writeFile is what the scanner learned about one non-test source: whether it names a write method
// (or gh's "-X" flag) and whether it calls CheckRequest.
type writeFile struct {
	path         string
	names, calls bool
}

// scanWrites reads every non-test .go source under root (testdata included only when root is inside it).
func scanWrites(root string) ([]writeFile, error) {
	var out []writeFile
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" && p != root && filepath.Base(root) != "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		af, perr := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if perr != nil {
			return perr
		}
		f := writeFile{path: p}
		ast.Inspect(af, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.BasicLit:
				if x.Kind == token.STRING && writeNames[x.Value] {
					f.names = true
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "MethodPost" || x.Sel.Name == "MethodPatch" || x.Sel.Name == "MethodPut" || x.Sel.Name == "MethodDelete" {
					f.names = true
				}
			case *ast.Ident:
				if x.Name == "CheckRequest" {
					f.calls = true
				}
			}
			return true
		})
		out = append(out, f)
		return nil
	})
	return out, err
}

// Any non-test source under effect/ that names a write method (or gh's "-X" flag) must go through
// CheckRequest. A new write path added without the guard fails here, so the refusal stays in one place.
func TestEveryWritePathUnderEffectUsesTheGuard(t *testing.T) {
	files, err := scanWrites(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f.names && !f.calls {
			t.Errorf("%s names a write method but never calls CheckRequest", f.path)
		}
	}
}

// Positive control: the scanner must catch a source that names a write method and skips the guard.
// testdata/unguarded_sender.go is such a source (the go tool ignores testdata, so it never builds);
// if the scanner stops seeing write paths, this fails where the assertion above would pass vacuously.
func TestWritePathScannerCatchesAnUnguardedSender(t *testing.T) {
	files, err := scanWrites("testdata")
	if err != nil {
		t.Fatal(err)
	}
	var caught int
	for _, f := range files {
		if f.names && !f.calls {
			caught++
		}
	}
	if caught != 1 {
		t.Fatalf("scanner flagged %d unguarded senders in testdata, want exactly 1", caught)
	}
}
