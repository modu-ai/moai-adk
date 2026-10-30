package ledger

import (
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
		{"the path is cleaned", []string{"./internal/a//x.go", "internal/a/../a/y.go"}, []string{"pkg:internal/a"}},
		{"nothing is nothing", nil, []string{}},
		{"blank paths are skipped", []string{"", "  "}, []string{}},
	}
	for _, c := range cases {
		if got := KeysFor(c.paths); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
