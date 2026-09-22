package jevcred

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// withTempHome swaps the package home-dir seam to a temp dir and restores it on
// cleanup. Mirrors internal/glmcred's helper: t.Setenv("HOME", ...) is
// deliberately NOT used — it pollutes parallel tests (CLAUDE.local.md §13).
func withTempHome(t *testing.T) string {
	t.Helper()
	orig := HomeDirFn
	dir := t.TempDir()
	HomeDirFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { HomeDirFn = orig })
	return dir
}

func TestPath_UnderMoaiHome(t *testing.T) {
	dir := withTempHome(t)
	got := Path()
	want := filepath.Join(dir, ".moai", ".env.typesafe")
	if got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	withTempHome(t)
	const key = "NOT-A-REAL-KEY-typesafe-acceptance-sentinel"
	if err := Save(key); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := Load(); got != key {
		t.Fatalf("Load() = %q, want %q", got, key)
	}
}

// AC-JEVC-010: a fresh credential write lands at 0600.
func TestSave_FileMode0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce POSIX permission bits; the 0600 assertion holds on unix only")
	}
	withTempHome(t)
	if err := Save("mode-test-key"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file mode = %o, want 0600", perm)
	}
}

// AC-JEVC-011: a pre-existing 0644 credential file is TIGHTENED to 0600 on the
// next write. os.WriteFile's perm argument applies at creation only, so a naive
// Save leaves it at 0644 — the latent defect internal/glmcred closes with an
// explicit Chmod and this package inherits if it forgets.
func TestSave_NarrowsExisting0644to0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows cannot create the wide-mode pre-existing file, so the narrowing is unobservable there")
	}
	dir := withTempHome(t)
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".moai", ".env.typesafe")
	if err := os.WriteFile(envPath, []byte(`TYPESAFE_API_KEY="stale"`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save("new-key-after-0644"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file mode = %o, want 0600 after rewrite of pre-existing wide-mode file", perm)
	}
	if got := Load(); got != "new-key-after-0644" {
		t.Fatalf("Load() = %q, want new-key-after-0644", got)
	}
}

func TestSave_DotenvQuotedForm(t *testing.T) {
	withTempHome(t)
	if err := Save("plain-key"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	body, err := os.ReadFile(Path())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(body), `TYPESAFE_API_KEY="plain-key"`) {
		t.Fatalf("credential file missing quoted TYPESAFE_API_KEY dotenv form; got:\n%s", body)
	}
}

func TestEscapeUnescape_RoundTrip(t *testing.T) {
	for _, raw := range []string{`a"b`, `a\b`, `a$b`, `plain`, `a\"$b`} {
		if got := UnescapeValue(EscapeValue(raw)); got != raw {
			t.Errorf("round trip of %q = %q", raw, got)
		}
	}
}

func TestSaveLoad_RoundTripSpecialCharacters(t *testing.T) {
	withTempHome(t)
	const key = `we"ird\key$value`
	if err := Save(key); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := Load(); got != key {
		t.Fatalf("Load() = %q, want %q", got, key)
	}
}

func TestLoad_TestEnvOverrideHonoured(t *testing.T) {
	withTempHome(t)
	if err := Save("on-disk-key"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Setenv(EnvTestTypeSafeKey, "override-from-env")
	if got := Load(); got != "override-from-env" {
		t.Fatalf("Load() = %q, want override-from-env (EnvTestTypeSafeKey not honoured)", got)
	}
}

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	withTempHome(t)
	if got := Load(); got != "" {
		t.Fatalf("Load() with no credential file = %q, want empty", got)
	}
}

func TestLoad_UnreadableFileReturnsEmpty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce the 0000 read denial this case depends on")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root: a 0000 file stays readable, so the denial cannot be observed")
	}
	dir := withTempHome(t)
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".moai", ".env.typesafe")
	if err := os.WriteFile(envPath, []byte(`TYPESAFE_API_KEY="x"`), 0o000); err != nil {
		t.Fatal(err)
	}
	if got := Load(); got != "" {
		t.Fatalf("Load() on unreadable file = %q, want empty (fail-open, REQ-JEVC-005)", got)
	}
}

func TestLoad_EmptyValueReturnsEmpty(t *testing.T) {
	dir := withTempHome(t)
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".moai", ".env.typesafe"),
		[]byte("# comment\n\nTYPESAFE_API_KEY=\"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := Load(); got != "" {
		t.Fatalf("Load() on empty value = %q, want empty", got)
	}
}

func TestPath_EmptyWhenHomeUnresolvable(t *testing.T) {
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { HomeDirFn = orig })
	t.Setenv("MOAI_HOME", "")
	if got := Path(); got != "" {
		t.Fatalf("Path() with unresolvable home = %q, want empty", got)
	}
}

func TestSave_ErrorsWhenHomeUnresolvable(t *testing.T) {
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { HomeDirFn = orig })
	t.Setenv("MOAI_HOME", "")
	if err := Save("k"); err == nil {
		t.Fatal("Save with unresolvable home = nil error, want error")
	}
}

func TestPath_MoaiHomeOverride(t *testing.T) {
	withTempHome(t)
	override := t.TempDir()
	t.Setenv("MOAI_HOME", override)
	if got, want := Path(), filepath.Join(override, ".env.typesafe"); got != want {
		t.Fatalf("Path() under MOAI_HOME = %q, want %q", got, want)
	}
}

// AC-JEVC-013 / AC-JEVC-014: the bounded disclosure. A key longer than four
// characters discloses exactly its final four; a key of four characters or
// fewer discloses NOTHING but Configured. A naive "last four or the whole key"
// fallback would disclose a short key entirely — the exact inverse of
// REQ-JEVC-020.
func TestView_DisclosureFloor(t *testing.T) {
	cases := []struct {
		name       string
		stored     string
		configured bool
		hint       string
	}{
		{"absent", "", false, ""},
		{"one char", "a", true, ""},
		{"exactly four", "abcd", true, ""},
		{"five chars", "abcde", true, "bcde"},
		{"long fixture", "NOT-A-REAL-KEY-0123456789wxyz", true, "wxyz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withTempHome(t)
			if tc.stored != "" {
				if err := Save(tc.stored); err != nil {
					t.Fatalf("Save: %v", err)
				}
			}
			got := View()
			if got.Configured != tc.configured {
				t.Errorf("Configured = %v, want %v", got.Configured, tc.configured)
			}
			if got.Hint != tc.hint {
				t.Errorf("Hint = %q, want %q", got.Hint, tc.hint)
			}
			if got.Hint != "" && len(got.Hint) != 4 {
				t.Errorf("Hint %q is not exactly four characters", got.Hint)
			}
			// The credential itself must never appear in the view model.
			if tc.stored != "" && got.Hint == tc.stored {
				t.Errorf("View() disclosed the whole credential %q", tc.stored)
			}
		})
	}
}

// Configured is the presence predicate the doctor check reads. It must agree
// with View().Configured on every input.
func TestConfigured_AgreesWithView(t *testing.T) {
	withTempHome(t)
	if Configured() {
		t.Error("Configured() = true with no stored credential")
	}
	if err := Save("abc"); err != nil {
		t.Fatal(err)
	}
	if !Configured() || !View().Configured {
		t.Error("Configured() = false with a stored credential")
	}
}
