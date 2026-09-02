package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

)

// ---------------------------------------------------------------------------
// SPEC-INBOX-DRAIN-GAP-001 M3 — the `moai inbox` lifecycle surface
// (AC-IBX-004/005/006). Tests drive the cobra command tree with t.Chdir'd
// fixture project roots; findProjectRoot resolves the fixture.
// ---------------------------------------------------------------------------

// runInboxCmd executes the inbox command tree with the given args against the
// current working directory, returning stdout, stderr, and Execute's error
// (the exit-0 / exit-non-zero boundary at this level).
func runInboxCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newInboxCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

// seedFixtureInbox writes a small non-empty lessons inbox under root.
func seedFixtureInbox(t *testing.T, root string) []byte {
	t.Helper()
	moai := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	data := []byte("{\"timestamp\":\"t\",\"event_key\":\"k\",\"summary\":\"s\",\"source\":\"src\",\"v\":1}\n" +
		"{\"timestamp\":\"t\",\"event_key\":\"k2\",\"summary\":\"s2\",\"source\":\"src\",\"v\":1}\n")
	if err := os.WriteFile(filepath.Join(moai, "lessons-inbox.jsonl"), data, 0o600); err != nil {
		t.Fatalf("seed inbox: %v", err)
	}
	return data
}

// seedFixtureMarker creates the LSEL drain-ownership marker directory.
func seedFixtureMarker(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state", "lsel"), 0o755); err != nil {
		t.Fatalf("mkdir marker: %v", err)
	}
}

// TestInboxStatus_ReportsAllFields covers AC-IBX-004 (REQ-IBX-005): status
// exits 0 and its output carries the live size, line count, cap distance,
// archive-generation count, and the ownership-regime token.
func TestInboxStatus_ReportsAllFields(t *testing.T) {
	root := t.TempDir()
	seedFixtureInbox(t, root)
	t.Chdir(root)

	out, _, err := runInboxCmd(t, "status")
	if err != nil {
		t.Fatalf("inbox status exited non-zero: %v", err)
	}
	for _, field := range []string{"size_bytes:", "lines:", "cap_distance_bytes:", "archive_generations:"} {
		if !strings.Contains(out, field) {
			t.Errorf("status output missing %q; output = %q", field, out)
		}
	}
	if !strings.Contains(out, "ownership: cap-managed") {
		t.Errorf("status output missing the cap-managed regime token; output = %q", out)
	}
	if !strings.Contains(out, "lines: 2") {
		t.Errorf("status output wrong line count; output = %q", out)
	}
}

// TestInboxStatus_CuratorRegime covers the curator branch of the regime token:
// marker present -> `curator` (REQ-IBX-005).
func TestInboxStatus_CuratorRegime(t *testing.T) {
	root := t.TempDir()
	seedFixtureInbox(t, root)
	seedFixtureMarker(t, root)
	t.Chdir(root)

	out, _, err := runInboxCmd(t, "status")
	if err != nil {
		t.Fatalf("inbox status exited non-zero: %v", err)
	}
	if !strings.Contains(out, "ownership: curator") {
		t.Errorf("status output missing the curator regime token; output = %q", out)
	}
}

// TestInboxCmdRegisteredOnRoot guards the root-command wiring: `moai inbox`
// must appear on the real root command tree (mirrors TestMCPCmdRegisteredOnRoot).
func TestInboxCmdRegisteredOnRoot(t *testing.T) {
	t.Parallel()
	for _, c := range rootCmd.Commands() {
		if c.Name() == "inbox" {
			return
		}
	}
	t.Fatal("`moai inbox` is not registered on the root command")
}

// TestInboxDrain_StandardInstallRotates covers AC-IBX-005 (REQ-IBX-006):
// marker absent + non-empty inbox -> exit 0, the live file rotated into an
// archive, and rotation statistics printed.
func TestInboxDrain_StandardInstallRotates(t *testing.T) {
	root := t.TempDir()
	seed := seedFixtureInbox(t, root)
	t.Chdir(root)

	out, _, err := runInboxCmd(t, "drain")
	if err != nil {
		t.Fatalf("inbox drain exited non-zero: %v", err)
	}
	gen1, readErr := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl.1"))
	if readErr != nil {
		t.Fatalf("drain did not rotate: %v", readErr)
	}
	if !bytes.Equal(gen1, seed) {
		t.Errorf("archive .1 does not carry the rotated content")
	}
	// A one-shot drain moves live into .1 and does not fabricate a fresh live
	// file — the collector's next append recreates it (O_CREATE path).
	if _, statErr := os.Stat(filepath.Join(root, ".moai", "lessons-inbox.jsonl")); !os.IsNotExist(statErr) {
		t.Errorf("live inbox still present after drain (rotation did not move it)")
	}
	if !strings.Contains(out, "rotated") || !strings.Contains(out, "archive_generations: 1") {
		t.Errorf("drain statistics missing from output; output = %q", out)
	}
}

// TestInboxDrain_EmptyInboxNothingToRotate covers the acceptance §B edge:
// empty/absent inbox + drain on a standard install -> exit 0, "nothing to
// rotate", no archive created.
func TestInboxDrain_EmptyInboxNothingToRotate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Chdir(root)

	out, errOut, err := runInboxCmd(t, "drain")
	if err != nil {
		t.Fatalf("drain on an absent inbox must exit 0: %v", err)
	}
	if !strings.Contains(out+errOut, "nothing to rotate") {
		t.Errorf("expected a nothing-to-rotate report; out = %q, errOut = %q", out, errOut)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".moai", "lessons-inbox.jsonl.1")); !os.IsNotExist(statErr) {
		t.Errorf("drain created an archive for an empty inbox")
	}
}

// TestInboxDrain_CuratorRefusal covers AC-IBX-006 (REQ-IBX-007): marker
// present -> exit non-zero, the notice names the wrapper-mediated drain
// (session_drain.sh), and the inbox bytes are unchanged.
func TestInboxDrain_CuratorRefusal(t *testing.T) {
	root := t.TempDir()
	seed := seedFixtureInbox(t, root)
	seedFixtureMarker(t, root)
	t.Chdir(root)

	sum := sha256.Sum256(seed)
	before := hex.EncodeToString(sum[:])

	out, errOut, err := runInboxCmd(t, "drain")
	if err == nil {
		t.Fatal("drain on a curator machine must exit non-zero")
	}
	combined := out + errOut + err.Error()
	if !strings.Contains(combined, "session_drain.sh") {
		t.Errorf("refusal notice does not name session_drain.sh; combined = %q", combined)
	}
	afterData, readErr := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if readErr != nil {
		t.Fatalf("live inbox missing after refusal: %v", readErr)
	}
	sumAfter := sha256.Sum256(afterData)
	if hex.EncodeToString(sumAfter[:]) != before {
		t.Errorf("inbox bytes changed under refusal (before %s, after %s)", before, hex.EncodeToString(sumAfter[:]))
	}
	if _, statErr := os.Stat(filepath.Join(root, ".moai", "lessons-inbox.jsonl.1")); !os.IsNotExist(statErr) {
		t.Errorf("refusal created an archive (zero mutation required)")
	}
}
