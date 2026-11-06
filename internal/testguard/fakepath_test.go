package testguard

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeRoot is the root of the invented paths: the platform separator, preceded
// by a drive on Windows, where a rooted path without a volume is not absolute.
func fakeRoot() string {
	if runtime.GOOS == "windows" {
		return `C:\`
	}
	return string(filepath.Separator)
}

// fakePath joins elements under fakeRoot with the platform separator.
func fakePath(elem ...string) string {
	return filepath.Join(append([]string{fakeRoot()}, elem...)...)
}

func TestFakePathIsAbsoluteOnThisPlatform(t *testing.T) {
	for _, p := range []string{fakeRoot(), fakePath("fake", "users", "me"), fakePath("a", "b", "c")} {
		if !filepath.IsAbs(p) {
			t.Fatalf("fake path %q is not absolute on %s", p, runtime.GOOS)
		}
	}
	if runtime.GOOS == "windows" && filepath.VolumeName(fakeRoot()) == "" {
		t.Fatal("the fake root carries no volume on windows")
	}
}

// On Windows the decision functions get the real drive-letter shape as a plain
// string; elsewhere those backslash paths are not absolute, so the case is skipped.
func TestWindowsShapedPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("drive-letter paths are absolute only on windows")
	}
	home := `C:\Users\me`
	tmp := `C:\Users\me\AppData\Local\Temp`
	if !tempExempt(home, tmp, tmp+`\x\gitconfig`) {
		t.Fatal("temp under home was not exempt")
	}
	if err := checkGlobalConfigPin(home, tmp, tmp+`\x\gitconfig`); err != nil {
		t.Fatalf("temp file under home was refused: %v", err)
	}
	if err := checkGlobalConfigPin(home, tmp, home+`\.gitconfig`); err == nil {
		t.Fatal("the real global config was accepted")
	}
}

// TestTempExemptTestsBuildFakePathsPerPlatform pins the audit iter6 N1 defect
// class: a rooted literal such as a leading-slash fake path has no volume, so
// filepath.IsAbs is false on Windows and every accepted-path case fails there.
// The exemption tests must build their fake absolute paths per platform.
func TestTempExemptTestsBuildFakePathsPerPlatform(t *testing.T) {
	src, err := os.ReadFile("tempexempt_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{`"/` + `fake`, `"/` + `elsewhere`} {
		if n := strings.Count(string(src), needle); n != 0 {
			t.Errorf("tempexempt_test.go spells %d rooted literal(s) starting %s; build them with fakePath", n, needle)
		}
	}
}
