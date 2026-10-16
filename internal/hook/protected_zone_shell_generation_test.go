package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProtectedZoneShellCdGenerationMixDeny pins the codex generation-mix example
// as an accepted residual (operator decision d-20261011T051116Z-8893, card t1584).
// The guard denies both shapes on purpose. The rm target's pre-4.2 reading keeps
// the NUL escape as literal text, and the lexical tail rule cleans the missing
// component's ".." onto the zone file, so the codex example is a precision false
// block. The rule must stay: at check time the guard cannot know whether the
// component is created later in the same command, so blocking is the safe
// direction. The mkdir-then-rm shape is the one real bash deletes; both were
// measured on bash 3.2.57. Gap: modern bash (5.x) is not observed on this host,
// so its NUL-termination behavior is not part of this pin. It lives outside the
// parsing matrix, next to the allow dispositions of REQ-ZSCD-005.
func TestProtectedZoneShellCdGenerationMixDeny(t *testing.T) {
	bs := string(rune(92))
	cases := []struct {
		name string
		cmd  string
	}{
		{name: "codex_example", cmd: "cd -- $'sub'; rm $'" + bs + "u0000/../a.log'"},
		{name: "mkdir_then_rm", cmd: "mkdir -- 'sub/" + bs + "u0000'; rm -- 'sub/" + bs + "u0000/../a.log'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"sub/a.log\"]\n"), "")
			if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "sub", "a.log"), []byte("protected"), 0o644); err != nil {
				t.Fatal(err)
			}
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.cmd})
			t.Logf("command=%q decision=%q reason=%q", tc.cmd, d, r)
			if d != DecisionDeny {
				t.Errorf("%q: decision=%q reason=%q, want deny", tc.cmd, d, r)
			}
		})
	}
}
