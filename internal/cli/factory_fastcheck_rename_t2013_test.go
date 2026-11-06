package cli

// factory_fastcheck_rename_t2013_test.go — F2 of the rc.40 final-candidate
// audit: a card that moves a file out of one package into another is, under
// git's default rename detection, listed by `git diff --name-only` as the new
// path only. The package the file left was then never selected, so a test there
// that still used the moved symbol (and no longer compiles) escaped the fast
// check. The change list must carry both sides of a rename.

import (
	"reflect"
	"strings"
	"testing"
)

func TestT2013FastCheckPackagesRenameAcrossPackagesSelectsBothSides(t *testing.T) {
	helper := func(pkg string) string {
		return "package " + pkg + "\n\n// Helper is shared by the package tests.\nfunc Helper() int {\n\tvalue := 1\n\tvalue += 2\n\tvalue += 3\n\treturn value\n}\n"
	}
	base := map[string]string{
		"go.mod":         "module example.com/m\n",
		"a/a.go":         "package a\n",
		"a/a_test.go":    "package a\n\nimport \"testing\"\n\nfunc TestUsesHelper(t *testing.T) { _ = Helper() }\n",
		"a/helper.go":    helper("a"),
		"b/b.go":         "package b\n",
		"c/untouched.go": "package c\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"b/helper.go": helper("b")}, []string{"a/helper.go"})

	// Non-vacuity: git must really report this as a rename, or the case below
	// proves nothing about rename handling.
	if got := fcGit(t, dir, "diff", "--name-status", "main...HEAD"); !strings.HasPrefix(got, "R") && !strings.Contains(got, "\nR") {
		t.Fatalf("fixture is not detected as a rename by git:\n%s", got)
	}

	if got, want := factoryFastCheckPackages(dir, "main"), []string{"./a", "./b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("packages = %q, want %q (both sides of the rename)", got, want)
	}
	changed, known := factoryChangedPaths(dir, "main")
	if !known {
		t.Fatal("changed paths unknown")
	}
	if !reflect.DeepEqual(changed, []string{"a/helper.go", "b/helper.go"}) {
		t.Errorf("changed paths = %q, want both a/helper.go and b/helper.go", changed)
	}
}
