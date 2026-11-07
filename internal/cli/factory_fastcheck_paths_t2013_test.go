package cli

import (
	"reflect"
	"testing"
)

// Sync audit 6 finding C4: the changed-file list must survive a path git would
// quote (non-ASCII under the default core.quotePath) or a path with a space, so
// a card that changes only such a fixture still plans its owning package.
func TestT2013FastCheckPackagesSurviveUnusualPathNames(t *testing.T) {
	base := map[string]string{
		"go.mod":                    "module example.com/m\n",
		"internal/pkg/p.go":         "package pkg\n",
		"internal/my pkg/q.go":      "package mypkg\n",
		"internal/uni/u.go":         "package uni\n",
		"internal/pkg/testdata/a":   "a\n",
		"internal/sp ace/data/x.md": "x\n",
	}
	cases := []struct {
		name   string
		writes map[string]string
		want   []string
	}{
		{"non-ASCII testdata file maps to its package", map[string]string{"internal/pkg/testdata/한글.txt": "v\n"}, []string{"./internal/pkg"}},
		{"non-ASCII go file maps to its package", map[string]string{"internal/uni/한글.go": "package uni\n"}, []string{"./internal/uni"}},
		{"directory with a space maps to its package", map[string]string{"internal/my pkg/q.go": "package mypkg\n// edit\n"}, []string{"./internal/my pkg"}},
		{"file with a space maps to its package", map[string]string{"internal/pkg/testdata/a b.txt": "v\n"}, []string{"./internal/pkg"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := fastCheckRepo(t, base, tc.writes, nil)
			if got := factoryChangedGoPackages(dir, "main"); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("packages = %q, want %q", got, tc.want)
			}
		})
	}
}
