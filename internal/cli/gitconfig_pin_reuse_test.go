package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// The reuse branch of sandboxGitConfigGlobal trusts an inherited pin only when
// it is the very file the marker names (card t1801, audit iter6 N2). None of
// these tests call t.Parallel: some swap process environment variables.

// TestSandboxGitConfigGlobal_MarkerMismatchGetsAFreshPin reproduces the audit
// probe: marker names one file, GIT_CONFIG_GLOBAL another existing file. The
// second file must not be trusted.
func TestSandboxGitConfigGlobal_MarkerMismatchGetsAFreshPin(t *testing.T) {
	standIn := filepath.Join(t.TempDir(), "stand-in-for-a-real-gitconfig")
	if err := os.WriteFile(standIn, []byte("[user]\n\tname = stand-in\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", standIn)
	t.Setenv(gitConfigSandboxEnv, filepath.Join(t.TempDir(), "some-other-pinned-file"))

	cleanup, err := sandboxGitConfigGlobal()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	if got := os.Getenv("GIT_CONFIG_GLOBAL"); got == standIn {
		t.Fatalf("GIT_CONFIG_GLOBAL still points at the unvalidated inherited file %q", got)
	}
}

func TestReusablePin(t *testing.T) {
	// A real home that is not under the temp dir, so a temp file is outside it.
	outsideHome := filepath.Join(filepath.VolumeName(os.TempDir())+string(filepath.Separator), "moai-no-such-home")
	regular := func(dir string) string {
		p := filepath.Join(dir, "gitconfig")
		if err := os.WriteFile(p, []byte("[user]\n\tname = x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	tmpFile := regular(t.TempDir())
	otherFile := regular(t.TempDir())
	fakeHomeDir := t.TempDir()
	inHome := regular(fakeHomeDir)
	dir := t.TempDir()
	missing := filepath.Join(t.TempDir(), "gone")

	cases := []struct {
		name                 string
		marker, global, home string
		want                 bool
	}{
		{"marker equals global, regular file outside any home: reused", tmpFile, tmpFile, outsideHome, true},
		{"marker differs from global: not reused", otherFile, tmpFile, outsideHome, false},
		{"empty marker: not reused", "", tmpFile, outsideHome, false},
		{"global inside the real home: not reused", inHome, inHome, fakeHomeDir, false},
		{"global is a directory: not reused", dir, dir, outsideHome, false},
		{"global is missing: not reused", missing, missing, outsideHome, false},
		{"global is the null device: not reused", os.DevNull, os.DevNull, outsideHome, false},
		{"real home unknown: not reused", tmpFile, tmpFile, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := reusablePin(tc.marker, tc.global, tc.home); got != tc.want {
				t.Fatalf("reusablePin(%q, %q, %q) = %v, want %v", tc.marker, tc.global, tc.home, got, tc.want)
			}
		})
	}
}
