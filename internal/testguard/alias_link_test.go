package testguard

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// requireAlias skips unless alias names the same directory as dir at run time,
// i.e. the filesystem is case-insensitive for this pair.
func requireAlias(t *testing.T, dir, alias string) {
	t.Helper()
	a, errA := os.Stat(dir)
	b, errB := os.Stat(alias)
	if dir == alias || errA != nil || errB != nil || !os.SameFile(a, b) {
		t.Skipf("filesystem is case-sensitive here (%q does not alias %q); the case-alias path cannot be exercised", alias, dir)
	}
}

// aliasLinkLayout builds: base/RealHome/child (the fake real home), base/outside,
// and outside/link -> base/REALHOME/child (a case-altered spelling of a home
// child). It returns the home, the link and the outside directory.
func aliasLinkLayout(t *testing.T) (home, link, outside string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on Windows")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home = filepath.Join(base, "RealHome")
	outside = filepath.Join(base, "outside")
	for _, d := range []string{filepath.Join(home, "child"), outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	requireAlias(t, home, filepath.Join(base, "REALHOME"))
	link = filepath.Join(outside, "link")
	if err := os.Symlink(filepath.Join(base, "REALHOME", "child"), link); err != nil {
		t.Fatal(err)
	}
	return home, link, outside
}

// TestCaseAliasViaSymlink_ExistingSeamIsRefused (shape 1): the seam is a link
// whose target is a case alias of a home child. The written path's ancestors
// never stat to the home, so only identity on the RESOLVED path catches it.
func TestCaseAliasViaSymlink_ExistingSeamIsRefused(t *testing.T) {
	home, link, _ := aliasLinkLayout(t)
	if err := CheckSeamHomes(home, []string{link}); err == nil {
		t.Fatalf("link %s -> a case alias of a home child was accepted", link)
	}
}

// TestCaseAliasViaSymlink_MissingSeamBelowLinkIsRefused (shape 2).
func TestCaseAliasViaSymlink_MissingSeamBelowLinkIsRefused(t *testing.T) {
	home, link, _ := aliasLinkLayout(t)
	seam := filepath.Join(link, "not-yet", "x")
	if _, err := os.Stat(seam); err == nil {
		t.Fatal("premise: the seam must not exist")
	}
	if err := CheckSeamHomes(home, []string{seam}); err == nil {
		t.Fatalf("missing seam %s below a link into the home was accepted", seam)
	}
}

// TestCaseAliasViaSymlink_LinkAsSandboxRootIsRefused (shape 3): judgeSandbox
// with the link itself as the sandbox root.
func TestCaseAliasViaSymlink_LinkAsSandboxRootIsRefused(t *testing.T) {
	home, link, _ := aliasLinkLayout(t)
	if err := judgeSandbox(home, link, []string{filepath.Join(link, "home"), link}); err == nil {
		t.Fatalf("judgeSandbox accepted a sandbox root %s that links into the real home", link)
	}
}

// TestCaseAliasViaSymlink_LinkToAncestorAliasIsRefused (shape 4, the 'above'
// direction): a link whose target is a case alias of an ANCESTOR of the real
// home. os.Stat follows the link, so the existing above-check already sees it.
func TestCaseAliasViaSymlink_LinkToAncestorAliasIsRefused(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on Windows")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(base, "parentdir")
	home := filepath.Join(parent, "RealHome")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{home, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	requireAlias(t, parent, filepath.Join(base, "PARENTDIR"))
	link := filepath.Join(outside, "uplink")
	if err := os.Symlink(filepath.Join(base, "PARENTDIR"), link); err != nil {
		t.Fatal(err)
	}
	if err := CheckSeamHomes(home, []string{link}); err == nil {
		t.Fatalf("link %s -> a case alias of an ancestor of the real home was accepted", link)
	}
}

// TestJudgeSandbox_CaseAliasTempDirKeepsRefusal (F7): TMPDIR/TMP/TEMP name a
// case alias of the fake home's parent and the sandbox root sits under the
// aliased home. The real home is under the temp dir by identity, so the
// Windows-layout exemption must NOT apply; only the identity clause in
// judgeSandbox sees this.
func TestJudgeSandbox_CaseAliasTempDirKeepsRefusal(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(base, "parentdir")
	home := filepath.Join(parent, "RealHome")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	aliasParent := filepath.Join(base, "PARENTDIR")
	requireAlias(t, parent, aliasParent)
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, aliasParent)
	}
	root := filepath.Join(aliasParent, "RealHome", "sandbox-root")
	if err := judgeSandbox(home, root, []string{filepath.Join(root, "home")}); err == nil {
		t.Fatal("exemption applied although the real home is under the temp dir by identity")
	}
}
