package quality

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// affected_scope_env_test.go — unit coverage of the two adapter helpers whose
// branches the real-adapter tests reach only in a child process or not at all.

// The changed-set query keeps ONLY the hook-provided index location: it is made
// absolute against the hook's working directory, and every other
// repository-location variable stays scrubbed (card t560).
func TestCommitQueryEnv(t *testing.T) {
	t.Run("a relative index location becomes absolute and the repository dir stays scrubbed", func(t *testing.T) {
		t.Setenv("GIT_INDEX_FILE", ".git/index")
		t.Setenv("GIT_DIR", "/elsewhere/.git")
		env := commitQueryEnv()
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if want := "GIT_INDEX_FILE=" + filepath.Join(cwd, ".git", "index"); !slices.Contains(env, want) {
			t.Errorf("env lacks %q", want)
		}
		if slices.ContainsFunc(env, func(kv string) bool { return strings.HasPrefix(kv, "GIT_DIR=") }) {
			t.Error("GIT_DIR leaked into the changed-set query")
		}
	})
	t.Run("an unset index location adds nothing", func(t *testing.T) {
		t.Setenv("GIT_INDEX_FILE", "")
		if slices.ContainsFunc(commitQueryEnv(), func(kv string) bool { return strings.HasPrefix(kv, "GIT_INDEX_FILE=") }) {
			t.Error("an empty GIT_INDEX_FILE was forwarded")
		}
	})
}

// go list -e exits 0 for errored and incomplete packages; problem() is what
// notices them.
func TestGoListPackageProblem(t *testing.T) {
	for _, tc := range []struct {
		name string
		pkg  goListPackage
		want string
	}{
		{"sound", goListPackage{ImportPath: "p"}, ""},
		{"errored", goListPackage{Error: &struct{ Err string }{Err: "no Go files\nsecond line"}}, "errored: no Go files"},
		{"incomplete with a dependency error", goListPackage{Incomplete: true, DepsErrors: []struct{ Err string }{{Err: "cannot find package x"}}}, "incomplete: cannot find package x"},
		{"incomplete without detail", goListPackage{Incomplete: true}, "incomplete"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pkg.problem(); got != tc.want {
				t.Errorf("problem() = %q, want %q", got, tc.want)
			}
		})
	}
}
