package profile

// SPEC-WEB-CONSOLE-011 M4 (REQ-WC11-034) — profile 패키지 경계의 방어 계층
// reproduction test. WritePreferences 는 GetPreferencesPath 로 경로를 만든 뒤
// os.MkdirAll 로 디렉터리를 암묵 생성하므로, path-traversal 이름(`../../x`)이
// profile store 밖 디렉터리를 만들 수 있다. 본 테스트는 WritePreferences 가
// 그런 이름을 거부하고 디렉터리를 생성하지 않음을 검증한다 (RED: 수정 전 FAIL).

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProfileNameTraversal verifies WritePreferences rejects path-traversal
// profile names without creating any directory outside the profile store. The
// base dir is nested two levels deep so a `../../` escape lands inside the test's
// t.TempDir() (auto-cleaned) yet outside the profile base.
func TestProfileNameTraversal(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "sub1", "sub2", "profiles")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	orig := BaseDirOverride
	BaseDirOverride = base
	t.Cleanup(func() { BaseDirOverride = orig })

	const sentinel = "moai-repro-escaped"
	escaped := filepath.Join(tmp, "sub1", sentinel)

	badNames := []string{
		"../../" + sentinel,
		"a/" + sentinel,
		"a\\" + sentinel,
		"..",
		"." + sentinel,
	}

	for _, name := range badNames {
		if err := WritePreferences(name, ProfilePreferences{UserName: "x"}); err == nil {
			t.Errorf("WritePreferences(%q) = nil, want error (invalid profile name)", name)
		}
	}

	if _, err := os.Stat(escaped); err == nil {
		t.Errorf("WritePreferences created a directory outside the profile store: %s", escaped)
	} else if !os.IsNotExist(err) {
		t.Errorf("unexpected stat error for %s: %v", escaped, err)
	}
}

// TestWritePreferencesAllowsValidNames guards against over-tightening: the two
// special names ("" and "default") plus an ordinary profile name must remain
// writable (no regression from the traversal guard).
func TestWritePreferencesAllowsValidNames(t *testing.T) {
	BaseDirOverrideCleanup(t)

	for _, name := range []string{"", "default", "myprofile"} {
		if err := WritePreferences(name, ProfilePreferences{UserName: "ok"}); err != nil {
			t.Errorf("WritePreferences(%q) = %v, want nil (valid name)", name, err)
		}
	}
}

// BaseDirOverrideCleanup points the profile store at a fresh temp dir for the
// duration of the test.
func BaseDirOverrideCleanup(t *testing.T) {
	t.Helper()
	orig := BaseDirOverride
	BaseDirOverride = t.TempDir()
	t.Cleanup(func() { BaseDirOverride = orig })
}
