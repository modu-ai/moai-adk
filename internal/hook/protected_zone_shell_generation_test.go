package hook

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProtectedZoneShellCdGenerationMixAllow pins the codex P2 example of the
// delta audit (finding N1, operator decision via 영실이, scope limited to the
// generation mix at protected_zone_shell.go:989). The first operand of a
// multi-word cd and the rm target must be read in the same bash generation.
// Real bash does not delete sub/a.log in this shape, so the guard must allow.
// It lives outside the parsing matrix: REQ-ZSCD-005 keeps allow dispositions
// for cd shapes outside the matrix.
func TestProtectedZoneShellCdGenerationMixAllow(t *testing.T) {
	bs := string(rune(92))
	cmd := "cd -- $'sub'; rm $'" + bs + "u0000/../a.log'"
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"sub/a.log\"]\n"), "")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "a.log"), []byte("protected"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
	t.Logf("command=%q decision=%q reason=%q", cmd, d, r)
	if d == DecisionDeny {
		t.Errorf("%q: decision=%q reason=%q, want allow", cmd, d, r)
	}
}
