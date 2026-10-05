package verify

import (
	"os"
	"strings"
	"testing"
	"time"
)

// receiptFixture is a receipt written after a passing out-of-hook check, and
// the state the Stop chain observes when nothing has changed since.
func receiptFixture() (Receipt, ReceiptState, time.Time) {
	recorded := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	r := Receipt{
		CheckID:      "sync-gate",
		Head:         "2222222222222222222222222222222222222222",
		TreeDigest:   "0123456789abcdef",
		ConfigDigest: "c0nf16d16e57c0de",
		Command:      "moai gate sync --record",
		ToolVersion:  "go version go1.26.8 darwin/arm64",
		ExitCode:     0,
		Verdict:      "pass",
		RecordedAt:   recorded,
	}
	state := ReceiptState{
		Head:         r.Head,
		TreeDigest:   r.TreeDigest,
		ConfigDigest: r.ConfigDigest,
		Command:      r.Command,
		ToolVersion:  r.ToolVersion,
	}
	return r, state, recorded.Add(time.Minute)
}

// TestCheckReceiptAcceptsUnchangedReceipt is the positive leg of AC-HPR-017:
// a receipt whose five fields all equal the current values is accepted.
func TestCheckReceiptAcceptsUnchangedReceipt(t *testing.T) {
	t.Parallel()
	r, state, now := receiptFixture()

	got := CheckReceipt(&r, state, now, 0)
	if !got.Run {
		t.Fatalf("unchanged receipt must be accepted, got not-run: %s", got.Reason)
	}
	if got.Receipt == nil || got.Receipt.Verdict != "pass" || got.Receipt.ExitCode != 0 {
		t.Fatalf("accepted receipt must carry the recorded outcome, got %+v", got.Receipt)
	}
}

// TestCheckReceiptFieldMismatchIsNotRun is AC-HPR-017's per-field leg: each of
// the five bound fields changed on its own makes the check read as not run,
// and the reason names that field.
func TestCheckReceiptFieldMismatchIsNotRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		field  string
		mutate func(*ReceiptState)
	}{
		{"head", func(s *ReceiptState) { s.Head = "3333333333333333333333333333333333333333" }},
		{"tree_digest", func(s *ReceiptState) { s.TreeDigest = "fedcba9876543210" }},
		{"config_digest", func(s *ReceiptState) { s.ConfigDigest = "0000000000000000" }},
		{"command", func(s *ReceiptState) { s.Command = "moai gate sync --record " }},
		{"tool_version", func(s *ReceiptState) { s.ToolVersion = "go version go1.26.9 darwin/arm64" }},
	}
	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			t.Parallel()
			r, state, now := receiptFixture()
			tc.mutate(&state)

			got := CheckReceipt(&r, state, now, 0)
			if got.Run {
				t.Fatalf("%s differs but the receipt was accepted", tc.field)
			}
			if !strings.Contains(got.Reason, tc.field) {
				t.Errorf("reason must name %s, got %q", tc.field, got.Reason)
			}
		})
	}
}

// TestCheckReceiptStoredFieldMismatchIsNotRun mutates the stored side instead
// of the current side, so a comparison that only checks one direction (or only
// checks that the current field is non-empty) cannot pass.
func TestCheckReceiptStoredFieldMismatchIsNotRun(t *testing.T) {
	t.Parallel()

	mutations := map[string]func(*Receipt){
		"head":          func(r *Receipt) { r.Head = "4444444444444444444444444444444444444444" },
		"tree_digest":   func(r *Receipt) { r.TreeDigest = "aaaaaaaaaaaaaaaa" },
		"config_digest": func(r *Receipt) { r.ConfigDigest = "bbbbbbbbbbbbbbbb" },
		"command":       func(r *Receipt) { r.Command = "moai gate sync" },
		"tool_version":  func(r *Receipt) { r.ToolVersion = "codex-cli 0.155.1" },
	}
	for field, mutate := range mutations {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			r, state, now := receiptFixture()
			mutate(&r)
			if got := CheckReceipt(&r, state, now, 0); got.Run || !strings.Contains(got.Reason, field) {
				t.Fatalf("stored %s differs: got run=%v reason=%q", field, got.Run, got.Reason)
			}
		})
	}
}

// TestCheckReceiptEmptyFieldIsNotRun: a receipt that does not bind a field (an
// entry written before the fields existed, or a producer that could not read
// the tool version) is not evidence, even when the current value is also empty.
func TestCheckReceiptEmptyFieldIsNotRun(t *testing.T) {
	t.Parallel()

	for _, field := range []string{"config_digest", "tool_version"} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			r, state, now := receiptFixture()
			switch field {
			case "config_digest":
				r.ConfigDigest, state.ConfigDigest = "", ""
			case "tool_version":
				r.ToolVersion, state.ToolVersion = "", ""
			}
			if got := CheckReceipt(&r, state, now, 0); got.Run || !strings.Contains(got.Reason, field) {
				t.Fatalf("unbound %s must read as not run: run=%v reason=%q", field, got.Run, got.Reason)
			}
		})
	}
}

// TestCheckReceiptAbsentIsNotRun: no receipt reads as not run, never passed.
func TestCheckReceiptAbsentIsNotRun(t *testing.T) {
	t.Parallel()
	_, state, now := receiptFixture()

	got := CheckReceipt(nil, state, now, 0)
	if got.Run || got.Receipt != nil {
		t.Fatalf("absent receipt must read as not run, got %+v", got)
	}
	if !strings.Contains(got.Reason, "absent") {
		t.Errorf("reason must say the receipt is absent, got %q", got.Reason)
	}
}

// TestCheckReceiptTTLIsExtraBound: the wall-clock TTL stays as an extra
// staleness bound on top of the field comparison (design §D3.6).
func TestCheckReceiptTTLIsExtraBound(t *testing.T) {
	t.Parallel()
	r, state, _ := receiptFixture()

	late := r.RecordedAt.Add(DefaultTTL + time.Second)
	if got := CheckReceipt(&r, state, late, 0); got.Run {
		t.Fatal("a receipt older than the TTL must read as not run")
	}
}

// TestCheckReceiptStoreRoundTrip: a receipt recorded to the snapshot store is
// read back for the unchanged state and accepted; for a changed tree digest
// the lookup finds nothing, which reads as not run.
func TestCheckReceiptStoreRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	r, state, now := receiptFixture()

	if err := RecordReceipt(root, r); err != nil {
		t.Fatalf("RecordReceipt: %v", err)
	}
	loaded := LoadReceipt(root, state)
	if got := CheckReceipt(loaded, state, now, 0); !got.Run {
		t.Fatalf("round-tripped receipt must be accepted: %s", got.Reason)
	}

	moved := state
	moved.TreeDigest = "9999999999999999"
	if got := CheckReceipt(LoadReceipt(root, moved), moved, now, 0); got.Run {
		t.Fatal("a receipt for another tree must not be found for the current tree")
	}
}

// TestCheckReceiptTruncatedIsAbsent: a receipt file cut off mid-write reads
// as absent, not as a valid or failing receipt (acceptance.md §E).
func TestCheckReceiptTruncatedIsAbsent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	r, state, now := receiptFixture()

	if err := RecordReceipt(root, r); err != nil {
		t.Fatalf("RecordReceipt: %v", err)
	}
	path := SnapshotPath(root, state.Head+":"+state.TreeDigest)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data[:len(data)/2], 0o644); err != nil {
		t.Fatal(err)
	}

	loaded := LoadReceipt(root, state)
	if loaded != nil {
		t.Fatalf("a truncated receipt must load as absent, got %+v", loaded)
	}
	if got := CheckReceipt(loaded, state, now, 0); got.Run {
		t.Fatal("a truncated receipt must read as not run")
	}
}

// TestCheckReceiptLegacyEntryIsNotRun: a snapshot entry written by the
// pre-receipt `moai verify record` (no config digest, no tool version) is still
// usable by Source.Lookup but is never accepted as a receipt.
func TestCheckReceiptLegacyEntryIsNotRun(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, state, now := receiptFixture()

	key := state.Head + ":" + state.TreeDigest
	if _, err := RecordCheck(root, key, CheckEntry{
		CheckID:    "sync-gate",
		Command:    state.Command,
		ExitCode:   0,
		RecordedAt: now.Add(-time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
	if got := CheckReceipt(LoadReceipt(root, state), state, now, 0); got.Run {
		t.Fatal("a legacy entry without the receipt fields must not be accepted")
	}
}

// TestConfigDigestIsCanonical: the digest the producer writes and the digest
// the Stop chain computes must agree regardless of input order, and must move
// when any named input's value moves.
func TestConfigDigestIsCanonical(t *testing.T) {
	t.Parallel()

	a := ConfigDigest(map[string]string{"quality.yaml": "x: 1", "MOAI_SYNC_GATE_BLOCKING": "1"})
	b := ConfigDigest(map[string]string{"MOAI_SYNC_GATE_BLOCKING": "1", "quality.yaml": "x: 1"})
	if a == "" || a != b {
		t.Fatalf("digest must be non-empty and order-independent: %q vs %q", a, b)
	}
	if c := ConfigDigest(map[string]string{"quality.yaml": "x: 2", "MOAI_SYNC_GATE_BLOCKING": "1"}); c == a {
		t.Fatal("digest must change when an input value changes")
	}
	// Name/value boundary: moving bytes between name and value is a change.
	if ConfigDigest(map[string]string{"ab": "c"}) == ConfigDigest(map[string]string{"a": "bc"}) {
		t.Fatal("digest must separate names from values")
	}
}
