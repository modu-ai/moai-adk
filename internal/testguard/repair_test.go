package testguard

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestKeys_BackslashPathEqualsSlashForm: seed keys use slash paths, so a path
// computed with OS separators must produce the same key. Proven without Windows
// by feeding the backslash spelling directly to the key function.
func TestKeys_BackslashPathEqualsSlashForm(t *testing.T) {
	slash := Keys([]Violation{{File: "internal/cli/a_test.go", Line: 3, Kind: KindUserHomeDir, Text: "x := os.UserHomeDir()"}})
	back := Keys([]Violation{{File: `internal\cli\a_test.go`, Line: 3, Kind: KindUserHomeDir, Text: "x := os.UserHomeDir()"}})
	if len(slash) != 1 || len(back) != 1 || slash[0] != back[0] {
		t.Fatalf("backslash key %q != slash key %q", back, slash)
	}
}

// fakeHomeWithTemp builds a fake real home that contains the temp dir, the
// Windows layout (%USERPROFILE%\AppData\Local\Temp), and points TMPDIR at it.
func fakeHomeWithTemp(t *testing.T) (home, tmp string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home = filepath.Join(base, "realhome")
	tmp = filepath.Join(home, "AppData", "Temp")
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", tmp)
	t.Setenv("TMP", tmp)
	t.Setenv("TEMP", tmp)
	return home, tmp
}

// TestRedirectHome_AllowsSandboxUnderTempInsideRealHome simulates Windows: the
// real home contains the temp dir, so the sandbox root is inside the real home.
func TestRedirectHome_AllowsSandboxUnderTempInsideRealHome(t *testing.T) {
	home, tmp := fakeHomeWithTemp(t)
	withRealHome(t, home)
	root := filepath.Join(tmp, "sandbox-root")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	fake := &fakeTB{TB: t, tmp: root}
	redirectHomeAt(fake, root)
	if len(fake.fatals) != 0 {
		t.Fatalf("redirect refused a sandbox under TempDir while the real home contains TempDir: %q", fake.fatals)
	}
	if len(fake.setenvs) != len(sandboxTargets(root)) {
		t.Fatalf("set %d variables; want %d", len(fake.setenvs), len(sandboxTargets(root)))
	}
}

// TestJudgeSandbox_StillRefusesSeamsInTheRealHome: the exemption covers only
// seams inside the sandbox root. A seam directly in the real home, one equal to
// it, one above it, an empty one and a relative one are refused.
func TestJudgeSandbox_StillRefusesSeamsInTheRealHome(t *testing.T) {
	home, tmp := fakeHomeWithTemp(t)
	root := filepath.Join(tmp, "sandbox-root")
	inside := filepath.Join(root, "home")

	cases := map[string][]string{
		"seam_in_real_home":    {inside, filepath.Join(home, ".codex")},
		"seam_equal_real_home": {home},
		"seam_above_real_home": {filepath.Dir(home)},
		"empty_seam":           {""},
		"relative_seam":        {filepath.Join("a", "b")},
	}
	for name, seams := range cases {
		t.Run(name, func(t *testing.T) {
			if err := judgeSandbox(home, root, seams); err == nil {
				t.Fatalf("judgeSandbox(%q) = nil; want a refusal", seams)
			}
		})
	}
	if err := judgeSandbox("", root, []string{inside}); err == nil {
		t.Fatal("an unknown real home must be refused")
	}
	if err := judgeSandbox(home, root, []string{inside}); err != nil {
		t.Fatalf("a seam inside the sandbox root under TempDir must be allowed: %v", err)
	}
}

// TestFakeHomeWithTemp_SetsEveryTempVariable: os.TempDir reads TMPDIR on Unix
// but TMP / TEMP on Windows, so the simulation must set all three.
func TestFakeHomeWithTemp_SetsEveryTempVariable(t *testing.T) {
	_, tmp := fakeHomeWithTemp(t)
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		if got := os.Getenv(name); got != tmp {
			t.Errorf("%s = %q; want the fake temp dir %q", name, got, tmp)
		}
	}
}

// TestCheckSeamHomes_RefusesDotDotThroughSymlink: filepath.Clean collapses ".."
// before symlinks resolve, so outside/link/../victim with link -> <home>/child
// looks outside the home while the kernel writes <home>/victim.
func TestCheckSeamHomes_RefusesDotDotThroughSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on Windows")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "realhome")
	outside := filepath.Join(base, "outside")
	if err := os.MkdirAll(filepath.Join(home, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "child"), filepath.Join(outside, "link")); err != nil {
		t.Fatal(err)
	}
	seam := outside + "/link/../victim"
	err = CheckSeamHomes(home, []string{seam})
	if err == nil {
		t.Fatalf("CheckSeamHomes(%q) = nil; the kernel resolves it to %s", seam, filepath.Join(home, "victim"))
	}
	if !strings.Contains(err.Error(), `".."`) {
		t.Fatalf("error = %v; want it to name the \"..\" element", err)
	}
}

// TestCheckSeamHomes_DotDotRuleIsExactElementMatch: any ".." element is refused
// with the dedicated error even when the cleaned path is outside the real home,
// while names that merely contain two dots keep their old judgment.
func TestCheckSeamHomes_DotDotRuleIsExactElementMatch(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "realhome")
	outside := filepath.Join(base, "outside")

	err := CheckSeamHomes(home, []string{outside + "/x/../y"})
	if err == nil || !strings.Contains(err.Error(), `".."`) {
		t.Fatalf("outside/x/../y: error = %v; want the \"..\" element refusal", err)
	}
	if err := CheckSeamHomes(home, []string{outside + `\x\..\y`}); runtime.GOOS == "windows" && err == nil {
		t.Fatal("a backslash-separated .. element was accepted on Windows")
	}
	for _, name := range []string{"..moai", "a..b", "x.."} {
		if err := CheckSeamHomes(home, []string{filepath.Join(outside, name, "y")}); err != nil {
			t.Errorf("%q is not a \"..\" element but was refused: %v", name, err)
		}
	}
	// The old judgment for a dotted child of the real home is unchanged.
	if err := CheckSeamHomes(home, []string{filepath.Join(home, "..moai")}); err == nil || strings.Contains(err.Error(), `".."`) {
		t.Fatalf("<home>/..moai must be refused as inside the home, not by the \"..\" rule: %v", err)
	}
}

// TestCheckSeamHomes_RefusesCaseAliasOfRealHome: on a case-insensitive
// filesystem REALHOME/x is the real home's directory; the string comparison
// alone misses it, so file identity is compared too.
func TestCheckSeamHomes_RefusesCaseAliasOfRealHome(t *testing.T) {
	home := filepath.Join(t.TempDir(), "RealHome")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := strings.ToUpper(home)
	a, errA := os.Stat(home)
	b, errB := os.Stat(alias)
	if alias == home || errA != nil || errB != nil || !os.SameFile(a, b) {
		t.Skipf("filesystem is case-sensitive here (%q does not alias %q); the case-alias path cannot be exercised", alias, home)
	}
	if err := CheckSeamHomes(home, []string{filepath.Join(alias, "x")}); err == nil {
		t.Fatalf("a case-altered spelling of %s/x was accepted", home)
	}
	if err := CheckSeamHomes(home, []string{alias}); err == nil {
		t.Fatalf("a case-altered spelling of the home itself (%s) was accepted", alias)
	}
	if err := CheckSeamHomes(filepath.Join(alias, "x"), []string{home}); err == nil {
		t.Fatalf("a seam above a case-altered real home was accepted")
	}
}

// TestCheckSeamHomes_MissingSeamIsJudgedByItsNearestExistingAncestor pins how a
// seam that does not exist yet is judged: by file identity of its nearest
// existing ancestor, plus the string checks.
func TestCheckSeamHomes_MissingSeamIsJudgedByItsNearestExistingAncestor(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "RealHome")
	other := filepath.Join(base, "other")
	for _, d := range []string{home, other} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("b_missing_seam_under_real_home_is_refused", func(t *testing.T) {
		seam := filepath.Join(home, "not-yet", "x")
		if _, err := os.Stat(seam); err == nil {
			t.Fatal("premise: the seam must not exist")
		}
		if err := CheckSeamHomes(home, []string{seam}); err == nil {
			t.Fatalf("a missing seam %q under the real home was accepted", seam)
		}
	})

	t.Run("b_missing_seam_under_case_alias_is_refused", func(t *testing.T) {
		alias := strings.ToUpper(home)
		a, errA := os.Stat(home)
		b, errB := os.Stat(alias)
		if alias == home || errA != nil || errB != nil || !os.SameFile(a, b) {
			t.Skipf("filesystem is case-sensitive here (%q does not alias %q)", alias, home)
		}
		seam := filepath.Join(alias, "not-yet", "x")
		if err := CheckSeamHomes(home, []string{seam}); err == nil {
			t.Fatalf("a missing seam %q under a case alias of the real home was accepted", seam)
		}
	})

	t.Run("c_missing_seam_outside_real_home_passes", func(t *testing.T) {
		seam := filepath.Join(other, "not-yet", "x")
		if err := CheckSeamHomes(home, []string{seam}); err != nil {
			t.Fatalf("a missing seam %q outside the real home was refused: %v", seam, err)
		}
	})

	t.Run("d_unstatable_ancestors_do_not_panic", func(t *testing.T) {
		// A real home that does not exist gives identity nothing to compare, so
		// only the string checks decide.
		ghost := filepath.Join(base, "ghost-home")
		if err := CheckSeamHomes(ghost, []string{filepath.Join(other, "a", "b")}); err != nil {
			t.Fatalf("string checks refused an unrelated seam: %v", err)
		}
		if err := CheckSeamHomes(ghost, []string{filepath.Join(ghost, "a")}); err == nil {
			t.Fatal("string checks accepted a seam under a non-existent real home")
		}
	})
}

// TestJudgeSandbox_JudgesEverySeamInEitherOrder: a passing seam outside the
// sandbox root must not end the judgment; a bad seam later (or earlier) in the
// list is still refused.
func TestJudgeSandbox_JudgesEverySeamInEitherOrder(t *testing.T) {
	home, tmp := fakeHomeWithTemp(t)
	root := filepath.Join(tmp, "sandbox-root")
	inside := filepath.Join(root, "home")
	outsideOK := filepath.Join(filepath.Dir(home), "elsewhere")
	bad := filepath.Join(home, ".codex")

	if err := CheckSeamHomes(home, []string{outsideOK}); err != nil {
		t.Fatalf("premise: the outside seam must pass on its own: %v", err)
	}
	orders := map[string][]string{
		"bad_last":  {inside, outsideOK, bad},
		"bad_first": {bad, inside, outsideOK},
	}
	for name, seams := range orders {
		t.Run(name, func(t *testing.T) {
			if err := judgeSandbox(home, root, seams); err == nil {
				t.Fatalf("judgeSandbox(%q) = nil; want a refusal of %s", seams, bad)
			}
		})
	}
}

// TestJudgeSandbox_NoExemptionWhenRealHomeIsUnderTempDir: the existing refusal
// tests inject fake homes under t.TempDir; the exemption must not apply to them.
func TestJudgeSandbox_NoExemptionWhenRealHomeIsUnderTempDir(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	root := filepath.Join(home, "sandbox")
	if err := judgeSandbox(home, root, []string{filepath.Join(root, "home")}); err == nil {
		t.Fatal("a real home under TempDir got the exemption; the refusal tests would stop refusing")
	}
}

// TestScanSource_MarkerInsideStringLiteralDoesNotExempt: only a real comment
// token is an allow marker.
func TestScanSource_MarkerInsideStringLiteralDoesNotExempt(t *testing.T) {
	src := fixtureBody("\tmsg := \"// home-guard:allow looks like a marker\"; home, _ := os.UserHomeDir(); _, _ = msg, home")
	vs, err := ScanSource("lit_test.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 {
		t.Fatalf("ScanSource = %+v; a marker inside a string literal must not exempt", vs)
	}

	raw := "package cli\n\nimport \"os\"\n\nfunc f() {\n\t_ = `\n// home-guard:allow raw string line\n`\n\t_, _ = os.UserHomeDir()\n}\n"
	vs, err = ScanSource("raw_marker_test.go", []byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 {
		t.Fatalf("ScanSource = %+v; a marker line inside a raw string must not exempt the call below it", vs)
	}

	real := fixtureBody("\t// home-guard:allow reads the sandbox", "\t_, _ = os.UserHomeDir()")
	vs, err = ScanSource("real_marker_test.go", []byte(real))
	if err != nil || len(vs) != 0 {
		t.Fatalf("a real comment must exempt: %+v, %v", vs, err)
	}
}

// The init fixture uses the same bounded policy without changing HOME. Keep
// its public entry point subject to the same refusal cases as RedirectHome.
func TestCheckSandboxHomesKeepsTemporaryExemptionBounded(t *testing.T) {
	home, tmp := fakeHomeWithTemp(t)
	root := filepath.Join(tmp, "init-sandbox")
	inside := filepath.Join(root, "home")
	if err := CheckSandboxHomes(home, root, []string{inside}); err != nil {
		t.Fatal(err)
	}
	for _, seam := range []string{home, filepath.Join(home, ".codex"), "", filepath.Join(root, "..", "escape")} {
		if err := CheckSandboxHomes(home, root, []string{inside, seam}); err == nil {
			t.Fatalf("unsafe seam %q accepted", seam)
		}
	}
	if err := CheckSeamHomes(home, []string{inside}); err == nil {
		t.Fatal("strict home policy was weakened")
	}
}
