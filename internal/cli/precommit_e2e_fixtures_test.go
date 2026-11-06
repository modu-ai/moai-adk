package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// precommit_e2e_fixtures_test.go — cross-platform plumbing shared by the
// pre-commit e2e tests (ported from card t2134's tree; card t2141).
//
// The Windows CI legs (release-pr-multi-os run 42147372729, windows cli-3)
// failed the pre-commit e2e tests whose fixtures replaced the commit
// subprocess PATH with the colon-separated POSIX string
// "/usr/bin:/bin:/usr/sbin:/sbin" — on Windows that is one nonexistent
// directory entry, so git.exe could not resolve its bundled sh to spawn the
// hook at all. The helper here builds the PATH from real directories with
// os.PathListSeparator.

// hookE2EPath builds the PATH for the commit subprocess: extraDir entries
// first (the fixture tool dirs, e.g. the directory holding gofmt), then the
// directories of the tools the REAL hook shells out to (git, sh, grep, sed,
// sort, tr, dirname), resolved from this process's PATH, then the classic
// POSIX directories where they exist. Deliberately NOT the ambient PATH: the
// tests must control whether moai is reachable, so only named tool
// directories enter the list.
func hookE2EPath(t *testing.T, extraDirs ...string) string {
	t.Helper()
	var dirs []string
	add := func(d string) {
		if d == "" {
			return
		}
		for _, x := range dirs {
			if x == d {
				return
			}
		}
		dirs = append(dirs, d)
	}
	for _, d := range extraDirs {
		add(d)
	}
	for _, tool := range []string{"git", "sh", "grep", "sed", "sort", "tr", "dirname"} {
		if p, err := exec.LookPath(tool); err == nil {
			add(filepath.Dir(p))
		}
	}
	// POSIX fallbacks for macOS/Linux — included only where they exist; a
	// hardcoded colon-separated list does not exist on Windows and leaves
	// the hook unspawnable there.
	for _, d := range []string{"/usr/bin", "/bin", "/usr/sbin", "/sbin"} {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			add(d)
		}
	}
	return strings.Join(dirs, string(os.PathListSeparator))
}
