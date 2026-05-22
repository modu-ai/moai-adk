package cli

// TestReservedAck_* exercise the reserved-filename acknowledgment ledger
// introduced by SPEC-V3R6-UPDATE-NOISE-001 (REQ-UN-001 through REQ-UN-005 and
// REQ-UN-011). Tests use t.TempDir() exclusively per CLAUDE.local.md §6 test
// isolation rule and never touch the real project tree.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withVerboseMode is a small helper that flips the package-level
// updateVerboseMode flag for the duration of the test body.  Each Test*
// function calls it with t.Cleanup so the previous value is restored even on
// failure, preventing cross-test contamination.
func withVerboseMode(t *testing.T, v bool) {
	t.Helper()
	prev := updateVerboseMode
	updateVerboseMode = v
	t.Cleanup(func() { updateVerboseMode = prev })
}

// seedReservedFile creates projectRoot/.moai/design/<name> with the given
// content and returns the absolute path.
func seedReservedFile(t *testing.T, projectRoot, name string, content []byte) string {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "design")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir design: %v", err)
	}
	target := filepath.Join(dir, name)
	if err := os.WriteFile(target, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return target
}

// TestReservedAck_FirstOccurrenceEmitsWarning covers AC-UN-001/002 first half +
// REQ-UN-001/002/003. The first update with a reserved file present must emit
// the legacy warning AND seed the ack ledger with the file's SHA-256.
func TestReservedAck_FirstOccurrenceEmitsWarning(t *testing.T) {
	withVerboseMode(t, false)

	projectRoot := t.TempDir()
	seedReservedFile(t, projectRoot, "tokens.json", []byte(`{"x":1}`))

	var buf bytes.Buffer
	if err := checkReservedCollision(projectRoot, &buf, false); err != nil {
		t.Fatalf("checkReservedCollision: %v", err)
	}

	if !strings.Contains(buf.String(), "warning: reserved filename: tokens.json") {
		t.Errorf("expected legacy warning, got: %q", buf.String())
	}

	// REQ-UN-001: ledger file must exist with schema {path: {sha256, acknowledged_at}}.
	ledgerPath := filepath.Join(projectRoot, ".moai", "state", "reserved-acknowledged.json")
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("ledger missing: %v", err)
	}
	ledger := map[string]reservedAckEntry{}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatalf("ledger not valid JSON: %v", err)
	}
	entry, ok := ledger["tokens.json"]
	if !ok {
		t.Fatalf("ledger missing tokens.json entry: %+v", ledger)
	}
	if len(entry.SHA256) != 64 {
		t.Errorf("sha256 must be 64-hex chars, got %q (len=%d)", entry.SHA256, len(entry.SHA256))
	}
	if entry.AcknowledgedAt == "" {
		t.Error("acknowledged_at must be populated")
	}
}

// TestReservedAck_SecondCallSilent covers AC-UN-002 + REQ-UN-002. After a first
// occurrence has stamped the ledger, a second checkReservedCollision call on
// the same unchanged file MUST NOT emit any warning.
func TestReservedAck_SecondCallSilent(t *testing.T) {
	withVerboseMode(t, false)

	projectRoot := t.TempDir()
	seedReservedFile(t, projectRoot, "tokens.json", []byte(`{"x":1}`))

	// First call — primes the ack ledger.
	var first bytes.Buffer
	if err := checkReservedCollision(projectRoot, &first, false); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if !strings.Contains(first.String(), "warning: reserved filename") {
		t.Fatalf("first call must emit warning, got: %q", first.String())
	}

	// Second call — same file, same content, should be silent.
	var second bytes.Buffer
	if err := checkReservedCollision(projectRoot, &second, false); err != nil {
		t.Fatalf("second call: %v", err)
	}
	if strings.Contains(second.String(), "warning: reserved filename") {
		t.Errorf("second call must be silent, got: %q", second.String())
	}
}

// TestReservedAck_HashDriftReemits covers AC-UN-003 + REQ-UN-004. When the
// reserved file's content changes between updates, the warning MUST re-appear
// and the ledger sha256 MUST be replaced with the new hash.
func TestReservedAck_HashDriftReemits(t *testing.T) {
	withVerboseMode(t, false)

	projectRoot := t.TempDir()
	target := seedReservedFile(t, projectRoot, "tokens.json", []byte(`{"x":1}`))

	// Prime ledger.
	var prime bytes.Buffer
	_ = checkReservedCollision(projectRoot, &prime, false)

	// Capture the pre-drift hash.
	ledgerPath := filepath.Join(projectRoot, ".moai", "state", "reserved-acknowledged.json")
	beforeData, _ := os.ReadFile(ledgerPath)
	beforeLedger := map[string]reservedAckEntry{}
	_ = json.Unmarshal(beforeData, &beforeLedger)
	beforeHash := beforeLedger["tokens.json"].SHA256

	// Modify file content — induces hash drift.
	if err := os.WriteFile(target, []byte(`{"x":2,"y":3}`), 0o644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	// Drift call — must re-emit warning.
	var drift bytes.Buffer
	if err := checkReservedCollision(projectRoot, &drift, false); err != nil {
		t.Fatalf("drift call: %v", err)
	}
	if !strings.Contains(drift.String(), "warning: reserved filename: tokens.json") {
		t.Errorf("drift call must emit warning, got: %q", drift.String())
	}

	// Verify ledger sha256 updated.
	afterData, _ := os.ReadFile(ledgerPath)
	afterLedger := map[string]reservedAckEntry{}
	_ = json.Unmarshal(afterData, &afterLedger)
	afterHash := afterLedger["tokens.json"].SHA256
	if afterHash == beforeHash {
		t.Errorf("hash drift not recorded: before=%s after=%s", beforeHash, afterHash)
	}
}

// TestReservedAck_VerboseBypass covers AC-UN-007 (reserved half) + REQ-UN-005.
// With updateVerboseMode=true the ack lookup is bypassed and EVERY collision
// emits the warning, even for previously acknowledged files.
func TestReservedAck_VerboseBypass(t *testing.T) {
	projectRoot := t.TempDir()
	seedReservedFile(t, projectRoot, "tokens.json", []byte(`{"x":1}`))

	// Prime ledger with verbose=false so the entry is recorded.
	withVerboseMode(t, false)
	var prime bytes.Buffer
	_ = checkReservedCollision(projectRoot, &prime, false)

	// Flip to verbose=true; subsequent call MUST emit warning despite ack record.
	withVerboseMode(t, true)
	var verbose bytes.Buffer
	if err := checkReservedCollision(projectRoot, &verbose, false); err != nil {
		t.Fatalf("verbose call: %v", err)
	}
	if !strings.Contains(verbose.String(), "warning: reserved filename: tokens.json") {
		t.Errorf("verbose mode must re-emit warning, got: %q", verbose.String())
	}
}

// TestReservedAck_CorruptedLedgerRecovers covers AC-UN-009 + REQ-UN-011. A
// truncated / malformed ledger MUST be silently re-initialized to an empty map
// (no crash, no error surfaced) and the next call MUST emit the warning as if
// the ledger were absent.
func TestReservedAck_CorruptedLedgerRecovers(t *testing.T) {
	withVerboseMode(t, false)

	projectRoot := t.TempDir()
	seedReservedFile(t, projectRoot, "tokens.json", []byte(`{"x":1}`))

	// Plant a corrupted ledger file.
	ledgerDir := filepath.Join(projectRoot, ".moai", "state")
	if err := os.MkdirAll(ledgerDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	ledgerPath := filepath.Join(ledgerDir, "reserved-acknowledged.json")
	if err := os.WriteFile(ledgerPath, []byte("this is not valid JSON {{{"), 0o644); err != nil {
		t.Fatalf("plant corrupt ledger: %v", err)
	}

	var buf bytes.Buffer
	if err := checkReservedCollision(projectRoot, &buf, false); err != nil {
		t.Fatalf("checkReservedCollision must not error on corrupt ledger: %v", err)
	}
	if !strings.Contains(buf.String(), "warning: reserved filename: tokens.json") {
		t.Errorf("corrupt ledger must trigger warning (fail-safe path), got: %q", buf.String())
	}

	// Verify ledger was rewritten to valid JSON.
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("ledger missing after recovery: %v", err)
	}
	ledger := map[string]reservedAckEntry{}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Errorf("ledger not re-initialized to valid JSON: %v (raw=%q)", err, data)
	}
	if _, ok := ledger["tokens.json"]; !ok {
		t.Errorf("ledger must contain new ack for tokens.json after recovery, got: %+v", ledger)
	}
}
