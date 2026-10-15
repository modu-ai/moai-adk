// Package testguard is the test-only home-directory guard shared by the test
// packages of this repository. It has two halves that share no dependency on
// the packages they guard:
//
//   - RedirectHome sandboxes the nine environment variables that resolve a home
//     directory or git's global config under t.TempDir, and refuses to continue when a redirected path
//     would resolve under the operator's real home.
//   - ScanTree is a light source scanner. It fails on a test file that reads the
//     home directory through os.UserHomeDir or expands a "~/" literal, unless the
//     line carries a `home-guard:allow <reason>` comment.
//
// Only the standard library is imported, so the scanner never compiles the
// packages it scans. Import this package only from _test.go files.
package testguard

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// RealHomeResolver resolves the operator's real home directory. It is the
// package-level seam of this guard: production code never reads it, and a test
// injects a fake real home (a t.TempDir path) to prove that the guard refuses a
// redirect reaching it. The default reads the account database and never HOME,
// because HOME is the variable a test redirects.
//
// The seam is mutable package state by design, mirroring codexUserHomeDir. A
// test that replaces it must restore it with t.Cleanup and must not call
// t.Parallel.
var RealHomeResolver = accountHomeDir

// RealHome returns the operator's real home directory through RealHomeResolver.
func RealHome() (string, error) {
	return RealHomeResolver()
}

// accountHomeDir reads the home directory from the account database. On Unix
// it looks up the process UID with user.LookupId, which fails rather than
// falling back to HOME when the account is unknown. On Windows it uses
// user.Current.
func accountHomeDir() (string, error) {
	var (
		u   *user.User
		err error
	)
	if runtime.GOOS == "windows" {
		u, err = user.Current()
	} else {
		u, err = user.LookupId(strconv.Itoa(os.Getuid()))
	}
	if err != nil {
		return "", fmt.Errorf("look up the current account: %w", err)
	}
	if u.HomeDir == "" {
		return "", errors.New("the account database names no home directory for this user")
	}
	return u.HomeDir, nil
}

// envTarget is one redirected environment variable and the sandbox path it takes.
type envTarget struct {
	name  string
	value string
}

// sandboxTargets lays the nine sandbox paths out under root. Each variable that
// defaults to a directory under HOME points at the same directory under the
// sandbox HOME, so a test that redirects HOME alone and a test that calls
// RedirectHome observe the same layout. HOME comes first because RedirectHome
// returns it.
func sandboxTargets(root string) []envTarget {
	home := filepath.Join(root, "home")
	return []envTarget{
		{"HOME", home},
		{"CODEX_HOME", filepath.Join(home, ".codex")},
		{"XDG_CONFIG_HOME", filepath.Join(home, ".config")},
		{"XDG_DATA_HOME", filepath.Join(home, ".local", "share")},
		{"XDG_STATE_HOME", filepath.Join(home, ".local", "state")},
		{"XDG_CACHE_HOME", filepath.Join(home, ".cache")},
		{"CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude")},
		{"MOAI_HOME", filepath.Join(home, ".moai")},
		// A file, not a directory: git reads and writes it. redirectHomeAt creates
		// it empty, so a `git config --global` write from a test, or from any
		// child git process, lands in the sandbox and never in ~/.gitconfig.
		{"GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig")},
	}
}

// RedirectHome forces the nine home-resolving variables under a fresh
// t.TempDir and returns the sandbox HOME. Call it before the test touches any
// file. t.Setenv restores every variable when the test ends, so the call must
// not be combined with t.Parallel.
//
// The helper fails the test, before it sets any variable or creates any
// directory, when the real home cannot be determined or when a redirected path
// would resolve under the real home.
func RedirectHome(t testing.TB) string {
	t.Helper()
	return redirectHomeAt(t, t.TempDir())
}

// redirectHomeAt is RedirectHome with the sandbox root supplied by the caller,
// so that the failure path can be driven with a root the test chose. The judgment
// runs before any variable is set and before any directory is created.
func redirectHomeAt(t testing.TB, root string) string {
	t.Helper()
	targets := sandboxTargets(root)

	realHome, err := RealHome()
	if err != nil {
		t.Fatalf("testguard: resolve the real home from the account database: %v", err)
		return ""
	}
	seams := make([]string, 0, len(targets))
	for _, tg := range targets {
		seams = append(seams, tg.value)
	}
	if err := CheckSeamHomes(realHome, seams); err != nil {
		t.Fatalf("testguard: %v; refusing to redirect the home directory", err)
		return ""
	}

	for _, tg := range targets {
		t.Setenv(tg.name, tg.value)
	}
	home := targets[0].value
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("testguard: create the sandbox HOME %s: %v", home, err)
		return ""
	}
	for _, tg := range targets {
		if tg.name != gitGlobalConfigEnv {
			continue
		}
		if err := os.WriteFile(tg.value, nil, 0o644); err != nil {
			t.Fatalf("testguard: create the empty sandbox git config %s: %v", tg.value, err)
			return ""
		}
	}
	return home
}

// gitGlobalConfigEnv names the variable that moves git's global config.
const gitGlobalConfigEnv = "GIT_CONFIG_GLOBAL"

// CheckSeamHomes is the pure guard decision. It returns an error when realHome
// is unknown, or when any seam path is empty, relative, equal to realHome, or
// located under realHome. A path is judged both as written and after its
// symlinks are resolved, so a link into the real home is refused too. A path on
// another volume is treated as outside.
func CheckSeamHomes(realHome string, seamPaths []string) error {
	if strings.TrimSpace(realHome) == "" {
		return errors.New("home guard: real home is empty; refusing to judge seam paths")
	}
	home := filepath.Clean(realHome)
	resolvedHome := resolveForGuard(home)
	for _, p := range seamPaths {
		if strings.TrimSpace(p) == "" {
			return errors.New("home guard: a seam path is empty and would fall back to the real home")
		}
		clean := filepath.Clean(p)
		if !filepath.IsAbs(clean) {
			return fmt.Errorf("home guard: seam path %q is relative and cannot be judged", p)
		}
		if within(clean, home) || within(resolveForGuard(clean), resolvedHome) {
			return fmt.Errorf("home guard: seam path %q resolves inside the real home %q", p, realHome)
		}
	}
	return nil
}

// within reports whether p equals home or lies beneath it. Both arguments are
// clean absolute paths. A path on another volume has no relative form and is
// outside.
func within(p, home string) bool {
	rel, err := filepath.Rel(home, p)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// resolveForGuard returns p with its symlinks resolved. A path that does not
// exist yet resolves through its nearest existing ancestor, so the judgment
// still follows any link above it.
func resolveForGuard(p string) string {
	clean := filepath.Clean(p)
	cur, rest := clean, ""
	for {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(resolved, rest)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return clean
		}
		if rest == "" {
			rest = filepath.Base(cur)
		} else {
			rest = filepath.Join(filepath.Base(cur), rest)
		}
		cur = parent
	}
}
