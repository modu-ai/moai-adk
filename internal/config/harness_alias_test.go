package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadHarnessFromLegacyCodexAliasesGPT pins the legacy alias: an llm.harness
// value written before the codex→gpt rename reads as gpt, not as the claude
// fallback. The claude and unknown rows are controls — a canonical value reads
// as itself and an out-of-set value still falls back.
func TestReadHarnessFromLegacyCodexAliasesGPT(t *testing.T) {
	t.Parallel()

	cases := []struct {
		stored string
		want   string
	}{
		{stored: "codex", want: "gpt"},
		{stored: "gpt", want: "gpt"},
		{stored: "claude", want: "claude"},
		{stored: "gemini", want: DefaultHarness},
	}
	for _, tc := range cases {
		t.Run(tc.stored, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			body := "llm:\n  harness: " + tc.stored + "\n"
			if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			if got := ReadHarnessFrom(dir); got != tc.want {
				t.Errorf("ReadHarnessFrom(harness: %s) = %q, want %q", tc.stored, got, tc.want)
			}
		})
	}
}
