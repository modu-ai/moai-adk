package ledger

import (
	"errors"
	"reflect"
	"testing"
)

func TestKeysForGivesTheNearestPackageDirectoryAndTheSharedKeys(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  []string
	}{
		{"a file is its directory", []string{"internal/agenticfactory/claim/claim.go"}, []string{"pkg:internal/agenticfactory/claim"}},
		{"a directory with a slash is itself", []string{"internal/agenticfactory/claim/"}, []string{"pkg:internal/agenticfactory/claim"}},
		{"two files of one package are one key", []string{"internal/a/x.go", "internal/a/y_test.go"}, []string{"pkg:internal/a"}},
		{"two packages are two keys, sorted", []string{"internal/b/x.go", "internal/a/x.go"}, []string{"pkg:internal/a", "pkg:internal/b"}},
		{"go.mod and go.sum share a key", []string{"go.mod", "go.sum"}, []string{"shared:go.mod"}},
		{"templates", []string{"internal/template/templates/.claude/rules/x.md"}, []string{"shared:templates"}},
		{"build files", []string{"Makefile", ".github/workflows/ci.yml"}, []string{"shared:build"}},
		{"rules", []string{".claude/rules/moai/workflow/x.md"}, []string{"shared:rules"}},
		{"a shared path adds no package key", []string{"go.mod", "internal/a/x.go"}, []string{"pkg:internal/a", "shared:go.mod"}},
		{"a root file", []string{"README.md"}, []string{"pkg:."}},
		{"the path is cleaned", []string{"./internal/a//x.go", "internal/a/./y.go"}, []string{"pkg:internal/a"}},
		{"testdata belongs to its package", []string{"internal/a/testdata/golden/x.json"}, []string{"pkg:internal/a"}},
		{"fixtures belongs to its package", []string{"internal/a/b/fixtures/"}, []string{"pkg:internal/a/b"}},
		{"a package and its testdata are one key", []string{"internal/a/x.go", "internal/a/testdata/y"}, []string{"pkg:internal/a"}},
		{"shared areas win over testdata", []string{"internal/template/templates/x/testdata/y"}, []string{"shared:templates"}},
		{"testdata at the root", []string{"testdata/x"}, []string{"pkg:."}},
		{"nothing is nothing", nil, []string{}},
		{"blank paths are skipped", []string{"", "  "}, []string{}},
	}
	for _, c := range cases {
		got, err := KeysFor(c.paths)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v %v, want %v", c.name, got, err, c.want)
		}
	}
}

// A path that leaves the repository, or is not written the way the repository's paths
// are, gives no key at all: the whole call is ErrInvalid.
func TestKeysForRefusesPathsThatLeaveTheRepository(t *testing.T) {
	for _, p := range []string{"/etc/passwd", "/abs/internal/a/x.go", "../x.go", "internal/../../x.go", "internal/a/../b/x.go", `internal\a\x.go`, "a/b\\c", ".."} {
		if got, err := KeysFor([]string{"internal/ok/x.go", p}); !errors.Is(err, ErrInvalid) || got != nil {
			t.Errorf("%q: %v %v", p, got, err)
		}
	}
}
