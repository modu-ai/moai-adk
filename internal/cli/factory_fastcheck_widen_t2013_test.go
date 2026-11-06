package cli

import (
	"reflect"
	"testing"
)

// card-review P2 #1: a changed .go path no package directory claims (a whole
// package directory deleted, with no Go ancestor up to a Go-less module root)
// widens the scope to the whole module even when another changed path selects
// a package; otherwise a test elsewhere that still references the deleted
// package would never run.
func TestT2013FastCheckPackagesUnownedGoPathWidensDespiteOtherPackage(t *testing.T) {
	base := map[string]string{
		"go.mod":              "module example.com/m\n",
		"a/a.go":              "package a\n",
		"gone/pkg/x.go":       "package pkg\n",
		"gone/pkg/x_test.go":  "package pkg\n",
		"other/user_test.go":  "package other\n",
		"other/user_other.go": "package other\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"a/a.go": "package a\n// edit\n"}, []string{"gone/pkg"})
	if got, want := factoryFastCheckPackages(dir, "main"), []string{"./..."}; !reflect.DeepEqual(got, want) {
		t.Errorf("packages = %q, want %q", got, want)
	}
}
