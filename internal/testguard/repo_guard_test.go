package testguard

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// seedFile is the checked-in allowlist seed: the home-access sites that existed
// when the guard landed. One key per line, in the format documented on Keys.
// Migrating a site to the helper removes its line; no line is ever added by hand.
//
//go:embed home_access_seed.txt
var seedFile string

// writeSeedEnv regenerates the seed from the current tree. It exists only so
// the follow-up migration card can shrink the seed; the guard never sets it.
const writeSeedEnv = "TESTGUARD_WRITE_SEED"

// scanRepo scans internal/ and cmd/ below the module root.
func scanRepo(t *testing.T) (root string, res Result) {
	t.Helper()
	root, err := findModuleRoot(".")
	if err != nil {
		t.Fatalf("locate module root: %v", err)
	}
	res, err = ScanTree(root, []string{"internal", "cmd"})
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	// Premise: the walk actually saw the repository. The tree holds well over a
	// thousand test files; a count this low means the scan found almost nothing
	// and an empty difference below would pass vacuously.
	if res.Files < 1000 {
		t.Fatalf("scanned %d test files; want >= 1000 — the scan no longer sees the repository", res.Files)
	}
	return root, res
}

// seedKeys returns the seed as a set.
func seedKeys() map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(seedFile, "\n") {
		if strings.TrimSpace(line) != "" {
			set[line] = true
		}
	}
	return set
}

// TestHomeGuard_RepositoryHasNoNewHomeAccess is the repository-wide static
// guard. A home read (os.UserHomeDir) or a "~/" literal in a test file under
// internal/ or cmd/ passes only when its key is in the checked-in seed or the
// line (or the line above) carries `home-guard:allow <reason>`. Sites present
// when the guard landed are seeded; every NEW site fails.
func TestHomeGuard_RepositoryHasNoNewHomeAccess(t *testing.T) {
	root, res := scanRepo(t)
	keys := Keys(res.Violations)

	if os.Getenv(writeSeedEnv) == "1" {
		out := strings.Join(keys, "\n") + "\n"
		if err := os.WriteFile(filepath.Join(root, "internal", "testguard", "home_access_seed.txt"), []byte(out), 0o644); err != nil {
			t.Fatalf("write seed: %v", err)
		}
		t.Logf("seed regenerated with %d entries", len(keys))
		return
	}

	seed := seedKeys()
	if len(seed) == 0 {
		t.Fatal("the seed is empty; every existing site would read as new, or the embed is broken")
	}
	current := map[string]bool{}
	newSites := 0
	for i, k := range keys {
		current[k] = true
		if !seed[k] {
			newSites++
			v := res.Violations[i]
			hint := "Use testguard.RedirectHome(t) or testguard.RealHome(), or add `// home-guard:allow <reason>` on the line or the line above."
			if v.Kind == KindIndirectGlobalGitConfig {
				hint = "This test reaches an exec of `git config --global`. Swap the seams on that path (internal/cli: isolateGitGlobalSeams(t) or swapM7Seams), call testguard.RedirectHome(t), or add `// home-guard:allow <reason>` on the line above the func."
			}
			t.Errorf("NEW home access %s:%d: %s: %s\n  %s", v.File, v.Line, v.Kind, v.Text, hint)
		}
	}
	stale := 0
	for k := range seed {
		if !current[k] {
			stale++
		}
	}
	t.Logf("scanned %d test files: %d violations, seed %d entries, %d new, %d stale seed entries (stale entries do not fail; regenerate with %s=1)",
		res.Files, len(keys), len(seed), newSites, stale, writeSeedEnv)
}

// TestKeys_StableUnderLineShiftsAndCountOccurrences pins the key format: the
// line number is not part of the key, whitespace is normalized, and repeated
// identical lines in one file get increasing occurrence indexes.
func TestKeys_StableUnderLineShiftsAndCountOccurrences(t *testing.T) {
	body := []string{"\thome, _ := os.UserHomeDir()", "\t_ = home", "\thome, _ = os.UserHomeDir()"}
	srcA := fixtureBody(body...)
	srcB := fixtureBody(append([]string{"\t_ = 1", "\t_ = 2"}, body...)...)

	scan := func(src string) []string {
		vs, err := ScanSource("x_test.go", []byte(src))
		if err != nil {
			t.Fatalf("ScanSource: %v", err)
		}
		return Keys(vs)
	}
	a, b := scan(srcA), scan(srcB)
	if len(a) != 2 || strings.Join(a, "|") != strings.Join(b, "|") {
		t.Fatalf("keys changed under a line shift:\n a=%q\n b=%q", a, b)
	}
	want0 := "x_test.go\tos.UserHomeDir\thome, _ := os.UserHomeDir()\t0"
	if a[0] != want0 {
		t.Fatalf("key[0] = %q; want %q", a[0], want0)
	}
	// The two source lines differ (":=" vs "="), so both have occurrence index 0.
	if !strings.HasSuffix(a[1], "\t0") {
		t.Fatalf("key[1] = %q; want occurrence index 0", a[1])
	}

	dup := fixtureBody("\thome, _ := os.UserHomeDir()", "\thome, _ := os.UserHomeDir()")
	d := scan(dup)
	if len(d) != 2 || !strings.HasSuffix(d[0], "\t0") || !strings.HasSuffix(d[1], "\t1") {
		t.Fatalf("identical lines must get occurrence indexes 0 and 1; got %q", d)
	}
}

// TestHomeGuard_NewSiteIsNotInSeed is the RED proof for the seed gate: a site
// that is not in the seed produces a key the seed does not contain.
func TestHomeGuard_NewSiteIsNotInSeed(t *testing.T) {
	vs, err := ScanSource("internal/cli/codex_lane_memory_guard_r2_test.go", []byte(incidentSource))
	if err != nil {
		t.Fatalf("ScanSource: %v", err)
	}
	keys := Keys(vs)
	if len(keys) != 1 {
		t.Fatalf("incident fixture produced %d keys; want 1", len(keys))
	}
	if seedKeys()[keys[0]] {
		t.Fatalf("the incident key %q is in the seed; the gate would not fail", keys[0])
	}
}

// TestRedirectHome_GitGlobalWriteLandsInSandbox is the write block: after
// RedirectHome, `git config --global` from the test process (and any child)
// reads and writes the sandbox file, not the operator's global git config. The
// write runs only after the variable is proven to point at the sandbox.
func TestRedirectHome_GitGlobalWriteLandsInSandbox(t *testing.T) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	home := RedirectHome(t)
	want := filepath.Join(home, ".gitconfig")
	if got := os.Getenv(gitGlobalConfigEnv); got != want {
		t.Fatalf("%s = %q; want the sandbox file %q — refusing to run git config --global", gitGlobalConfigEnv, got, want)
	}
	if info, err := os.Stat(want); err != nil || info.Size() != 0 {
		t.Fatalf("sandbox git config: stat err %v, info %v; want an existing empty file", err, info)
	}

	cmd := exec.Command(gitPath, "config", "--global", "--add", "safe.directory", "/testguard/probe")
	cmd.Dir = t.TempDir()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config --global --add: %v\n%s", err, out)
	}
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("read sandbox git config: %v", err)
	}
	if !strings.Contains(string(data), "/testguard/probe") {
		t.Fatalf("the global write did not land in the sandbox file; content:\n%s", data)
	}
}

// TestSessionWorktreeTests_AllUseRedirectHome: every top-level Test function in
// internal/cli/session_worktree*_test.go calls testguard.RedirectHome. These
// tests run git, and git writes the global config; the main checkout had no
// isolation there.
func TestSessionWorktreeTests_AllUseRedirectHome(t *testing.T) {
	root, err := findModuleRoot(".")
	if err != nil {
		t.Fatalf("locate module root: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(root, "internal", "cli", "session_worktree*_test.go"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) < 14 {
		t.Fatalf("found %d session_worktree test files; want >= 14 — the glob no longer sees them", len(files))
	}
	sort.Strings(files)
	tests := 0
	for _, f := range files {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, d := range file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || !strings.HasPrefix(fd.Name.Name, "Test") || fd.Body == nil {
				continue
			}
			tests++
			found := false
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "RedirectHome" {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "testguard" {
						found = true
					}
				}
				return !found
			})
			if !found {
				t.Errorf("%s: %s does not call testguard.RedirectHome", filepath.Base(f), fd.Name.Name)
			}
		}
	}
	if tests < 80 {
		t.Fatalf("checked %d Test functions; want >= 80 — the scan no longer sees them", tests)
	}
}
