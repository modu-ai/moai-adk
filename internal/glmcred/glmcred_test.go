package glmcred

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// withTempHome swaps the package home-dir seam to a temp dir and restores it
// on cleanup. Mirrors the userHomeDirFn override pattern used in internal/cli.
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
	want := filepath.Join(dir, ".moai", ".env.glm")
	if got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	withTempHome(t)
	const key = "NOT-A-REAL-KEY-glm-acceptance-sentinel"
	if err := Save(key); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := Load(); got != key {
		t.Fatalf("Load() = %q, want %q", got, key)
	}
}

func TestSave_FileMode0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce POSIX permission bits: os.Stat reports a synthesized mode, so the 0600 assertion cannot hold there; the credential-file tightening stays covered on unix")
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

func TestSave_NarrowsExisting0644to0600(t *testing.T) {
	// REQ-GKI-006-002: a pre-existing wide-mode credential file MUST be tightened
	// to 0600 on the next save, on BOTH the console save path and the moai glm
	// interactive save path (D-3 — they share this one writer).
	// os.WriteFile's perm argument applies at creation only, so a naive Save
	// leaves a pre-existing 0644 file at 0644. This is the §A.3 latent defect.
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not enforce POSIX permission bits: the pre-existing 0644 file cannot be created with a wide mode there, so the narrowing cannot be observed; the behavior stays covered on unix")
	}
	dir := withTempHome(t)
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".moai", ".env.glm")
	if err := os.WriteFile(envPath, []byte(`GLM_API_KEY="stale"`), 0o644); err != nil {
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
	// REQ-GKI-002-001: the credential file holds the quoted GLM_API_KEY dotenv form.
	withTempHome(t)
	if err := Save("plain-key"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	body, err := os.ReadFile(Path())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(body), `GLM_API_KEY="plain-key"`) {
		t.Fatalf("credential file missing quoted GLM_API_KEY dotenv form; got:\n%s", body)
	}
}

func TestLoad_TestEnvOverrideHonoured(t *testing.T) {
	// C-4: the MOAI_TEST_GLM_KEY override that loadGLMKey honours is a test seam
	// and must keep working after any extraction.
	withTempHome(t)
	if err := Save("on-disk-key"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Setenv(EnvTestGLMKey, "override-from-env")
	if got := Load(); got != "override-from-env" {
		t.Fatalf("Load() = %q, want override-from-env (EnvTestGLMKey not honoured)", got)
	}
}

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	withTempHome(t)
	if got := Load(); got != "" {
		t.Fatalf("Load() on missing file = %q, want empty", got)
	}
}

func TestSave_ReplaceInPlace(t *testing.T) {
	// REQ-GKI-006-001: replacing the key rewrites the same single credential file
	// rather than creating a second location.
	dir := withTempHome(t)
	if err := Save("first-key"); err != nil {
		t.Fatal(err)
	}
	if err := Save("second-key"); err != nil {
		t.Fatal(err)
	}
	// Exactly one credential file under the home tree.
	count := 0
	err := filepath.Walk(filepath.Join(dir, ".moai"), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(p, ".env.glm") {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if count != 1 {
		t.Fatalf("found %d .env.glm files, want exactly 1", count)
	}
	if got := Load(); got != "second-key" {
		t.Fatalf("Load() = %q, want second-key", got)
	}
}

func TestEscapeRoundTrip(t *testing.T) {
	cases := []string{
		`simple`,
		`with "quotes"`,
		`back\slash`,
		`dollar$sign`,
		`mix "\"\$ together`,
	}
	for _, in := range cases {
		escaped := EscapeValue(in)
		got := UnescapeValue(escaped)
		if got != in {
			t.Errorf("round-trip %q: escaped=%q unescaped=%q", in, escaped, got)
		}
	}
}
