package testguard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// fakeTB records the failure path of redirectHomeAt. It implements only the
// methods the helper calls. Any other testing.TB method panics through the nil
// embedded interface, so an unexpected call stays visible.
type fakeTB struct {
	testing.TB
	tmp     string
	fatals  []string
	setenvs []string
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.fatals = append(f.fatals, fmt.Sprintf(format, args...))
}

func (f *fakeTB) Setenv(key, value string) {
	f.setenvs = append(f.setenvs, key+"="+value)
}

func (f *fakeTB) TempDir() string {
	return f.tmp
}

// withRealHome points RealHomeResolver at home for the rest of the test.
func withRealHome(t *testing.T, home string) {
	t.Helper()
	prev := RealHomeResolver
	RealHomeResolver = func() (string, error) { return home, nil }
	t.Cleanup(func() { RealHomeResolver = prev })
}

// valueOf returns the sandbox path that targets assigns to name.
func valueOf(t *testing.T, targets []envTarget, name string) string {
	t.Helper()
	for _, tg := range targets {
		if tg.name == name {
			return tg.value
		}
	}
	t.Fatalf("no sandbox target named %s", name)
	return ""
}

// assertRefusedBeforeWriting checks the refusal contract: the helper reported
// one fatal failure, set no variable, and created no sandbox directory.
func assertRefusedBeforeWriting(t *testing.T, fake *fakeTB, root, wantSubstring string) {
	t.Helper()
	if len(fake.fatals) != 1 {
		t.Fatalf("redirectHomeAt reported %d fatal failure(s); want exactly 1: %q", len(fake.fatals), fake.fatals)
	}
	if !strings.Contains(fake.fatals[0], wantSubstring) {
		t.Fatalf("fatal message = %q; want it to contain %q", fake.fatals[0], wantSubstring)
	}
	if len(fake.setenvs) != 0 {
		t.Fatalf("redirectHomeAt set %d variable(s) before refusing: %q; want none", len(fake.setenvs), fake.setenvs)
	}
	for _, tg := range sandboxTargets(root) {
		if _, err := os.Stat(tg.value); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("sandbox path %s exists after the refusal (stat err %v); the helper wrote before judging", tg.value, err)
		}
	}
}

// TestRedirectHome_RefusesRealHomeAsWriteTarget is the RED case from the card:
// the injected real home equals a redirected write target. The helper must
// t.Fatal before it sets any variable or creates any directory.
func TestRedirectHome_RefusesRealHomeAsWriteTarget(t *testing.T) {
	root := t.TempDir()
	codexHome := valueOf(t, sandboxTargets(root), "CODEX_HOME")
	withRealHome(t, codexHome)

	fake := &fakeTB{TB: t, tmp: root}
	redirectHomeAt(fake, root)

	// The sandbox HOME is the parent of the injected real home, so the seam list
	// is refused as "above the real home"; a path equal to or inside it would be
	// refused as "inside". Either way nothing is written.
	assertRefusedBeforeWriting(t, fake, root, "the real home")
	t.Logf("captured t.Fatalf output: %s", fake.fatals[0])
}

// TestRedirectHome_RefusesRealHomeAboveTheSandbox: a real home that contains
// every sandbox root is refused the same way.
func TestRedirectHome_RefusesRealHomeAboveTheSandbox(t *testing.T) {
	root := t.TempDir()
	withRealHome(t, filepath.Dir(root))

	fake := &fakeTB{TB: t, tmp: root}
	redirectHomeAt(fake, root)

	assertRefusedBeforeWriting(t, fake, root, "inside the real home")
}

// TestRedirectHome_RefusesWhenRealHomeIsUnknown: an unresolvable real home is
// a refusal, never a guess.
func TestRedirectHome_RefusesWhenRealHomeIsUnknown(t *testing.T) {
	root := t.TempDir()
	prev := RealHomeResolver
	RealHomeResolver = func() (string, error) { return "", errors.New("no account") }
	t.Cleanup(func() { RealHomeResolver = prev })

	fake := &fakeTB{TB: t, tmp: root}
	redirectHomeAt(fake, root)

	assertRefusedBeforeWriting(t, fake, root, "resolve the real home")
}

// TestRedirectHome_RefusesEmptyRealHome: an empty real home cannot be judged.
func TestRedirectHome_RefusesEmptyRealHome(t *testing.T) {
	root := t.TempDir()
	withRealHome(t, "")

	fake := &fakeTB{TB: t, tmp: root}
	redirectHomeAt(fake, root)

	assertRefusedBeforeWriting(t, fake, root, "real home is empty")
}

// TestRedirectHome_AcceptsUnrelatedRealHome: a real home beside the sandbox,
// not above it, lets the redirect proceed and sets all ten variables.
func TestRedirectHome_AcceptsUnrelatedRealHome(t *testing.T) {
	root := t.TempDir()
	withRealHome(t, t.TempDir())

	fake := &fakeTB{TB: t, tmp: root}
	home := redirectHomeAt(fake, root)

	if len(fake.fatals) != 0 {
		t.Fatalf("redirectHomeAt refused an unrelated real home: %q", fake.fatals)
	}
	if len(fake.setenvs) != len(sandboxTargets(root)) {
		t.Fatalf("redirectHomeAt set %d variable(s); want %d", len(fake.setenvs), len(sandboxTargets(root)))
	}
	if want := filepath.Join(root, "home"); home != want {
		t.Fatalf("redirectHomeAt returned %q; want the sandbox HOME %q", home, want)
	}
}

// TestRedirectHome_SetsTenVariablesUnderTempDir runs the real helper and
// checks the ten variables name paths under one sandbox root, with HOME
// first and returned.
func TestRedirectHome_SetsTenVariablesUnderTempDir(t *testing.T) {
	home := RedirectHome(t)
	root := filepath.Dir(home)

	if n := len(sandboxTargets(root)); n != 10 {
		t.Fatalf("sandbox layout has %d variables; the guard covers exactly ten", n)
	}
	for _, tg := range sandboxTargets(root) {
		got := os.Getenv(tg.name)
		if got != tg.value {
			t.Errorf("%s = %q; want the sandbox path %q", tg.name, got, tg.value)
		}
		if !strings.HasPrefix(got, root+string(filepath.Separator)) {
			t.Errorf("%s = %q; want a path under the sandbox root %q", tg.name, got, root)
		}
	}
	if got := os.Getenv("USERPROFILE"); got != home {
		t.Fatalf("USERPROFILE = %q; want the sandbox HOME %q (os.UserHomeDir reads it on Windows)", got, home)
	}
	// home-guard:allow reads the sandbox HOME that RedirectHome just set, to prove os.UserHomeDir resolves inside it
	if got, err := os.UserHomeDir(); err != nil || got != home {
		t.Fatalf("os.UserHomeDir() = %q, %v; want the sandbox HOME %q", got, err, home)
	}
	if os.Getenv("HOME") != home {
		t.Fatalf("HOME = %q; RedirectHome returned %q", os.Getenv("HOME"), home)
	}
}

// TestAccountHomeDir_IgnoresHOME pins the premise of the seam: the default
// resolver reads the account database, so a redirected HOME does not change
// its answer.
func TestAccountHomeDir_IgnoresHOME(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the independent account lookup below is Unix-only; Windows is not verified (see the card's Gaps)")
	}
	decoy := t.TempDir()
	t.Setenv("HOME", decoy)

	got, err := accountHomeDir()
	if err != nil {
		t.Fatalf("accountHomeDir: %v", err)
	}
	if got == decoy {
		t.Fatalf("accountHomeDir() = the HOME override %q; the resolver must read the account database", got)
	}
	u, err := user.LookupId(strconv.Itoa(os.Getuid()))
	if err != nil {
		t.Fatalf("independent account lookup: %v", err)
	}
	if got != u.HomeDir {
		t.Fatalf("accountHomeDir() = %q; the account database names %q", got, u.HomeDir)
	}
}

// childEnv marks the subprocess that runs the real t.Fatal path.
const childEnv = "MOAI_TESTGUARD_CHILD"

// TestRedirectHome_RealFatalStopsTheTest exercises the real testing.T failure.
// The parent re-executes this test binary with childEnv set. In the child, the
// real home is the parent of every sandbox root, so RedirectHome must fail the
// child test. The parent then requires a non-zero exit and the refusal message.
func TestRedirectHome_RealFatalStopsTheTest(t *testing.T) {
	if os.Getenv(childEnv) == "1" {
		// Every t.TempDir call returns a new directory under one per-test temp
		// directory, and that directory is the parent of every sandbox root. The
		// resolver runs inside RedirectHome, after its own TempDir call, so the
		// real home it reports contains the sandbox root.
		RealHomeResolver = func() (string, error) { return filepath.Dir(t.TempDir()), nil }
		RedirectHome(t)
		// Reached only if the helper did not refuse. The parent treats a passing
		// child as a failure.
		os.Exit(0)
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("locate the test binary: %v", err)
	}
	cmd := exec.Command(exe, "-test.run=^TestRedirectHome_RealFatalStopsTheTest$", "-test.v")
	cmd.Env = append(os.Environ(), childEnv+"=1")
	out, err := cmd.CombinedOutput()
	t.Logf("child output:\n%s", out)
	if err == nil {
		t.Fatal("the child test passed; RedirectHome did not refuse a real home that contains its sandbox")
	}
	if !strings.Contains(string(out), "inside the real home") {
		t.Fatalf("child output lacks the refusal message:\n%s", out)
	}
}

// TestHomeGuard_RejectsPathInsideRealHome is the negative case for the guard
// decision: a seam path equal to or under the real home is refused, and so is
// an unjudgeable input.
func TestHomeGuard_RejectsPathInsideRealHome(t *testing.T) {
	realHome := filepath.Join(t.TempDir(), "home")
	outside := t.TempDir()

	cases := []struct {
		name      string
		realHome  string
		seamPaths []string
	}{
		{name: "equal_to_real_home", realHome: realHome, seamPaths: []string{realHome}},
		{name: "equal_with_trailing_separator", realHome: realHome, seamPaths: []string{realHome + string(filepath.Separator)}},
		{name: "direct_child", realHome: realHome, seamPaths: []string{filepath.Join(realHome, ".moai")}},
		{name: "deep_child", realHome: realHome, seamPaths: []string{filepath.Join(realHome, ".moai", "claude-profiles")}},
		{name: "child_named_with_leading_dots", realHome: realHome, seamPaths: []string{filepath.Join(realHome, "..moai")}},
		{name: "one_inside_among_outside", realHome: realHome, seamPaths: []string{outside, filepath.Join(realHome, ".claude"), outside}},
		{name: "empty_seam_path", realHome: realHome, seamPaths: []string{""}},
		{name: "relative_seam_path", realHome: realHome, seamPaths: []string{filepath.Join(".moai", "home")}},
		{name: "empty_real_home", realHome: "", seamPaths: []string{outside}},
	}
	if home, err := RealHome(); err == nil && home != "" {
		cases = append(cases, struct {
			name      string
			realHome  string
			seamPaths []string
		}{name: "actual_real_home_child", realHome: home, seamPaths: []string{filepath.Join(home, ".moai")}})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckSeamHomes(tc.realHome, tc.seamPaths); err == nil {
				t.Errorf("CheckSeamHomes(%q, %q) = nil, want an error", tc.realHome, tc.seamPaths)
			}
		})
	}
}

// TestHomeGuard_AcceptsTempDir is the positive case: t.TempDir() paths are
// accepted against the actual real home, and a sibling that only shares a name
// prefix with the real home is not mistaken for a child.
func TestHomeGuard_AcceptsTempDir(t *testing.T) {
	t.Run("temp_dirs_against_actual_real_home", func(t *testing.T) {
		realHome, err := RealHome()
		if err != nil {
			t.Skipf("real home unavailable: %v", err)
		}
		fakeHome := t.TempDir()
		seamPaths := []string{
			fakeHome,
			filepath.Join(fakeHome, ".moai", "claude-profiles"),
			filepath.Join(fakeHome, ".moai"),
		}
		if err := judgeSandbox(realHome, fakeHome, seamPaths); err != nil {
			t.Errorf("bounded temp sandbox against actual real home = %v, want nil", err)
		}
		resolvedHome, homeErr := filepath.EvalSymlinks(realHome)
		resolvedTemp, tempErr := filepath.EvalSymlinks(fakeHome)
		if homeErr != nil || tempErr != nil {
			t.Fatalf("resolve actual guard paths: home=%v temp=%v", homeErr, tempErr)
		}
		rel, relErr := filepath.Rel(resolvedHome, resolvedTemp)
		underHome := relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
		pureErr := CheckSeamHomes(realHome, seamPaths)
		if underHome {
			if pureErr == nil {
				t.Error("the pure guard must still refuse real-home descendants; only the bounded temp sandbox receives the exemption")
			}
		} else if pureErr != nil {
			t.Errorf("CheckSeamHomes(real home, external temp dirs) = %v, want nil", pureErr)
		}
	})

	t.Run("sibling_sharing_name_prefix", func(t *testing.T) {
		base := t.TempDir()
		realHome := filepath.Join(base, "home")
		sibling := filepath.Join(base, "home2", ".moai")
		if err := CheckSeamHomes(realHome, []string{sibling}); err != nil {
			t.Errorf("CheckSeamHomes(%q, [%q]) = %v, want nil", realHome, sibling, err)
		}
	})
}

// TestHomeGuard_RefusesSymlinkIntoRealHome: a link outside the real home that
// points into it is refused, because the guard judges resolved paths too.
func TestHomeGuard_RefusesSymlinkIntoRealHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on Windows")
	}
	base := t.TempDir()
	realHome := filepath.Join(base, "home")
	if err := os.MkdirAll(realHome, 0o755); err != nil {
		t.Fatalf("mkdir real home: %v", err)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(realHome, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if err := CheckSeamHomes(realHome, []string{filepath.Join(link, ".moai")}); err == nil {
		t.Fatal("CheckSeamHomes accepted a path that resolves through a symlink into the real home")
	}
}
