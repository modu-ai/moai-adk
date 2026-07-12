package routing

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestRecord_EmptySessionUsesNosessionKey exercises the degraded single-session
// key path (D5): an empty session id maps to the "nosession" pending file.
func TestRecord_EmptySessionUsesNosessionKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.Record(PendingRow{SessionID: "", MatchedSubcommand: "run"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, pendingPrefix+"nosession"+pendingSuffix)); err != nil {
		t.Fatalf("empty session should write routing-pending-nosession.json: %v", err)
	}
	if sessionKey("") != "nosession" {
		t.Fatal("sessionKey empty must map to nosession")
	}
}

// TestWriterAppend_MkdirError covers the Append error branch when the ledger
// parent directory cannot be created (a plain file occupies the parent path).
func TestWriterAppend_MkdirError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "sub")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Parent "sub" is a file, so MkdirAll(dir/sub) fails.
	w := NewWriter(filepath.Join(blocker, "ledger.jsonl"))
	if err := w.Append(Row{SchemaVersion: SchemaVersion}); err == nil {
		t.Fatal("expected append error when parent dir cannot be created")
	}
}

// TestWriterAppend_OpenError covers the Append error branch when the ledger path
// itself is a directory.
func TestWriterAppend_OpenError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ledgerAsDir := filepath.Join(dir, LedgerFileName)
	if err := os.Mkdir(ledgerAsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := NewWriter(ledgerAsDir).Append(Row{SchemaVersion: SchemaVersion}); err == nil {
		t.Fatal("expected append error when ledger path is a directory")
	}
}

// TestRerouteSelf_MalformedOwnPending: a corrupt own-pending file is dropped so
// the fresh record overwrites it (no reroute row, no error).
func TestRerouteSelf_MalformedOwnPending(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	own := filepath.Join(dir, pendingPrefix+"corrupt"+pendingSuffix)
	if err := os.WriteFile(own, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(recordRow("corrupt", "run")); err != nil {
		t.Fatalf("record over a corrupt own-pending should recover: %v", err)
	}
	if len(readLedger(t, dir)) != 0 {
		t.Fatal("a corrupt own-pending must not produce a reroute row")
	}
	if !pendingExists(dir, "corrupt") {
		t.Fatal("fresh pending must overwrite the corrupt file")
	}
}

// TestSweepStale_MalformedForeignSkipped: a corrupt FOREIGN pending file is left
// untouched (fail-open), never aborted.
func TestSweepStale_MalformedForeignSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	foreign := filepath.Join(dir, pendingPrefix+"garbage"+pendingSuffix)
	if err := os.WriteFile(foreign, []byte("<<not json>>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(recordRow("cur", "run")); err != nil {
		t.Fatal(err)
	}
	if len(readLedger(t, dir)) != 0 {
		t.Fatal("a corrupt foreign row must not be aborted")
	}
	if !pendingExists(dir, "garbage") {
		t.Fatal("a corrupt foreign row must be left in place (fail-open)")
	}
}

// TestFinalizeOnStop_MalformedPending: a corrupt own-pending file makes finalize
// a fail-open no-op — error to sink, nil returned; and silent with a nil sink.
func TestFinalizeOnStop_MalformedPending(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	own := filepath.Join(dir, pendingPrefix+"bad"+pendingSuffix)
	if err := os.WriteFile(own, []byte("{oops"), 0o644); err != nil {
		t.Fatal(err)
	}
	var sink bytes.Buffer
	if err := s.FinalizeOnStop("bad", &sink); err != nil {
		t.Fatalf("finalize over corrupt pending must be fail-open, got %v", err)
	}
	if sink.Len() == 0 {
		t.Fatal("corrupt-pending finalize must surface the parse error to the sink")
	}
	// nil sink -> silent fail-open (exercises logErr nil-sink branch).
	if err := s.FinalizeOnStop("bad", nil); err != nil {
		t.Fatalf("finalize with nil sink must not error, got %v", err)
	}
}

// TestAppendEvidence_MalformedReturnsError: appending to a corrupt pending file
// surfaces the parse error to the caller (CLI reports fail-open at its boundary).
func TestAppendEvidence_MalformedReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	own := filepath.Join(dir, pendingPrefix+"e"+pendingSuffix)
	if err := os.WriteFile(own, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvidence("e", EvidenceRef{Kind: KindGateExit, Value: "0"}); err == nil {
		t.Fatal("expected parse error appending evidence to a corrupt pending file")
	}
	if err := s.AppendDelegation("e", Delegation{Agent: "x", Outcome: "fail"}); err == nil {
		t.Fatal("expected parse error appending delegation to a corrupt pending file")
	}
}

// TestWritePending_Error: when the pending path is a non-empty directory, Record
// surfaces the writePending WriteFile error.
func TestWritePending_Error(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	// Occupy the pending path with a non-empty directory so WriteFile fails and
	// the pre-clear Remove of the directory also fails.
	pdir := filepath.Join(dir, pendingPrefix+"wp"+pendingSuffix)
	if err := os.Mkdir(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdir, "keep"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Record(recordRow("wp", "run")); err == nil {
		t.Fatal("expected writePending error when pending path is a non-empty directory")
	}
}
