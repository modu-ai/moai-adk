package testguard

import (
	"os"
	"path/filepath"
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
