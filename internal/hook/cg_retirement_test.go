package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyCGGuardIsRawDataNotLiveMode(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai/config/sections")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		raw  string
		want bool
	}{{"llm: {team_mode: cg}\n", true}, {"llm: {team_mode: claude} # team_mode: cg\n", false}, {"llm: {team_mode: cg, team_mode: claude}\n", true}} {
		if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(tc.raw), 0600); err != nil {
			t.Fatal(err)
		}
		if got := hasLegacyCGConfiguration(root); got != tc.want {
			t.Fatalf("guard %q=%v", tc.raw, got)
		}
	}
}
