package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// fastCheckRepo builds a temp git repo whose `main` holds base, then a `card`
// branch whose single commit writes the given files and deletes the given
// paths. It returns the repo dir; the release tip is `main`.
func fastCheckRepo(t *testing.T, base map[string]string, writes map[string]string, deletes []string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	put := func(files map[string]string) {
		for p, body := range files {
			full := filepath.Join(dir, filepath.FromSlash(p))
			if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	fcGit(t, dir, "init", "-q", "-b", "main")
	put(base)
	fcGit(t, dir, "add", "-A")
	fcGit(t, dir, "commit", "-q", "-m", "base")
	fcGit(t, dir, "checkout", "-q", "-b", "card")
	put(writes)
	for _, p := range deletes {
		if err := os.RemoveAll(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
			t.Fatal(err)
		}
	}
	fcGit(t, dir, "add", "-A")
	fcGit(t, dir, "commit", "-q", "-m", "card change")
	return dir
}

// The fast check selects the Go packages whose tests can be affected by a
// change: a testdata fixture or a go:embed asset belongs to the package that
// owns the directory, so changing only it still runs that package's tests.
func TestT2013FastCheckPackagesMapNonGoChangesToOwningPackage(t *testing.T) {
	goBase := map[string]string{
		"go.mod":                                  "module example.com/m\n",
		"internal/pkg/p.go":                       "package pkg\n",
		"internal/pkg/p_test.go":                  "package pkg\n",
		"internal/pkg/testdata/x.json":            "{}\n",
		"internal/pkg/sub/dir/f.txt":              "f\n",
		"internal/template/t.go":                  "package template\n",
		"internal/template/templates/a/b.md":      "b\n",
		"internal/template/templates/old/c.md":    "c\n",
		"internal/only/o_test.go":                 "package only\n",
		"internal/only/testdata/golden/case.txt":  "g\n",
		"docs/guide.md":                           "g\n",
		"internal/nogo/data/readme.md":            "r\n",
		"internal/nogo/data/nested/deeper/doc.md": "d\n",
	}
	cases := []struct {
		name    string
		writes  map[string]string
		deletes []string
		want    []string
	}{
		{"testdata fixture maps to its package", map[string]string{"internal/pkg/testdata/x.json": "{\"v\":1}\n"}, nil, []string{"./internal/pkg"}},
		{"embedded asset maps to the package that embeds it", map[string]string{"internal/template/templates/a/b.md": "changed\n"}, nil, []string{"./internal/template"}},
		{"new file under a nested asset dir maps to the owning package", map[string]string{"internal/template/templates/new/dir/n.md": "n\n"}, nil, []string{"./internal/template"}},
		{"changed go file maps to its package as before", map[string]string{"internal/pkg/p.go": "package pkg\n// edit\n"}, nil, []string{"./internal/pkg"}},
		{"test-only package keeps its own directory", map[string]string{"internal/only/o_test.go": "package only\n// edit\n"}, nil, []string{"./internal/only"}},
		{"testdata under a test-only package maps to that package", map[string]string{"internal/only/testdata/golden/case.txt": "g2\n"}, nil, []string{"./internal/only"}},
		{"several changes in one package de-duplicate", map[string]string{"internal/pkg/p.go": "package pkg\n// e\n", "internal/pkg/testdata/x.json": "{}\n// e\n"}, nil, []string{"./internal/pkg"}},
		{"changes in two packages are sorted", map[string]string{"internal/template/templates/a/b.md": "e\n", "internal/pkg/testdata/x.json": "e\n"}, nil, []string{"./internal/pkg", "./internal/template"}},
		{"deleted file whose directory vanished maps to the nearest ancestor with Go files", nil, []string{"internal/pkg/sub"}, []string{"./internal/pkg"}},
		{"deleted asset directory maps to the embedding package", nil, []string{"internal/template/templates/old"}, []string{"./internal/template"}},
		{"change with no Go ancestor below a Go-less root selects nothing", map[string]string{"internal/nogo/data/nested/deeper/doc.md": "e\n"}, nil, nil},
		{"top-level doc with a Go-less root selects nothing", map[string]string{"docs/guide.md": "e\n"}, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := fastCheckRepo(t, goBase, tc.writes, tc.deletes)
			if got := factoryChangedGoPackages(dir, "main"); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("packages = %q, want %q", got, tc.want)
			}
		})
	}
}

// When the module root itself holds Go files it is the nearest Go ancestor of a
// top-level change, so it is selected as `.`.
func TestT2013FastCheckPackagesRootWithGoFilesIsSelected(t *testing.T) {
	base := map[string]string{
		"go.mod":        "module example.com/m\n",
		"main.go":       "package main\n",
		"docs/guide.md": "g\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"docs/guide.md": "e\n"}, nil)
	if got, want := factoryChangedGoPackages(dir, "main"), []string{"."}; !reflect.DeepEqual(got, want) {
		t.Errorf("packages = %q, want %q", got, want)
	}
}
