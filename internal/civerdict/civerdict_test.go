package civerdict_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/civerdict"
)

// fixture returns a valid record (REQ-CV-004 five-field schema).
func fixture() civerdict.Record {
	return civerdict.Record{
		HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
		Conclusion: civerdict.ConclusionFailure,
		RunID:      "run-123",
		ObservedAt: time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Producer:   "moai ci-verdict",
	}
}

// AC-CV-001 (REQ-CV-004): Save writes exactly one file named for the head
// under .moai/state/ci-verdicts/, and Load returns the same five fields.
func TestSaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	rec := fixture()
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "state", "ci-verdicts"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != rec.HeadSHA+".json" {
		t.Fatalf("record files = %v, want exactly %s.json", entries, rec.HeadSHA)
	}
	got, err := civerdict.Load(root, rec.HeadSHA)
	if err != nil || got == nil {
		t.Fatalf("Load: %v, %v", got, err)
	}
	if *got != rec {
		t.Errorf("round trip = %+v, want %+v", *got, rec)
	}
}

// AC-CV-001 (REQ-CV-004): re-recording the same head with the same content is
// a byte-identical rewrite (idempotent), still exactly one file.
func TestSaveIdempotentRewrite(t *testing.T) {
	root := t.TempDir()
	rec := fixture()
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	p := civerdict.Path(root, rec.HeadSHA)
	first, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	second, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Errorf("rewrite is not byte-identical:\nfirst  %q\nsecond %q", first, second)
	}
	entries, _ := os.ReadDir(filepath.Dir(p))
	if len(entries) != 1 {
		t.Errorf("files after rewrite = %d, want 1", len(entries))
	}
}

// AC-CV-001 (REQ-CV-004): a re-run with changed content overwrites
// (last-writer-wins).
func TestSaveLastWriterWins(t *testing.T) {
	root := t.TempDir()
	rec := fixture()
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	rec.Conclusion = civerdict.ConclusionSuccess
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	got, err := civerdict.Load(root, rec.HeadSHA)
	if err != nil || got == nil {
		t.Fatalf("Load: %v, %v", got, err)
	}
	if got.Conclusion != civerdict.ConclusionSuccess {
		t.Errorf("conclusion = %q, want last writer's %q", got.Conclusion, civerdict.ConclusionSuccess)
	}
}

// REQ-CV-004: conclusion membership, head presence, and RFC 3339 observed_at
// are validated; an invalid record is never written.
func TestValidateRejects(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*civerdict.Record)
	}{
		{"empty-head", func(r *civerdict.Record) { r.HeadSHA = "" }},
		{"bad-conclusion", func(r *civerdict.Record) { r.Conclusion = "cancelled" }},
		{"empty-conclusion", func(r *civerdict.Record) { r.Conclusion = "" }},
		{"bad-observed-at", func(r *civerdict.Record) { r.ObservedAt = "yesterday" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := fixture()
			tc.mut(&rec)
			if err := rec.Validate(); err == nil {
				t.Fatalf("Validate accepted %+v", rec)
			}
			root := t.TempDir()
			if err := civerdict.Save(root, rec); err == nil {
				t.Errorf("Save wrote an invalid record")
			}
		})
	}
}

// REQ-CV-002: an offline input file naming head, conclusion, and run id
// parses into a record whose bytes are indistinguishable in schema from a
// fetched verdict (both go through Save).
func TestParseInput(t *testing.T) {
	root := t.TempDir()
	in := `{"head_sha":"0123456789abcdef0123456789abcdef01234567","conclusion":"failure","run_id":"run-9","observed_at":"2026-09-26T12:00:00Z","producer":"test"}`
	rec, err := civerdict.ParseInput([]byte(in))
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := civerdict.Load(root, rec.HeadSHA)
	if err != nil || got == nil {
		t.Fatalf("Load: %v, %v", got, err)
	}
	if *got != rec {
		t.Errorf("recorded = %+v, want %+v", *got, rec)
	}
	if _, err := civerdict.ParseInput([]byte("not json")); err == nil {
		t.Errorf("ParseInput accepted unparseable input")
	}
}

// Missing file and a stored-head mismatch (filename-collision defense) are
// both absent, never errors.
func TestLoadAbsent(t *testing.T) {
	root := t.TempDir()
	rec := fixture()
	if err := civerdict.Save(root, rec); err != nil {
		t.Fatal(err)
	}
	if got, err := civerdict.Load(root, "ffffffffffffaaaaaaaaaaaaaaaaaaaaaaaaffff"); got != nil || err != nil {
		t.Errorf("missing head = %v, %v; want nil, nil", got, err)
	}
	// A record file whose stored head differs from its filename is absent.
	other := fixture()
	other.HeadSHA = "ffffffffffffaaaaaaaaaaaaaaaaaaaaaaaaffff"
	if err := os.WriteFile(civerdict.Path(root, rec.HeadSHA), mustJSON(t, other), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := civerdict.Load(root, rec.HeadSHA); got != nil || err != nil {
		t.Errorf("mismatched head = %v, %v; want nil, nil", got, err)
	}
}

// LoadAll skips an unreadable file and reports it by name, so a caller can
// list it not-observed (convergenceFile precedent).
func TestLoadAllUnreadable(t *testing.T) {
	root := t.TempDir()
	good := fixture()
	if err := civerdict.Save(root, good); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(root, ".moai", "state", "ci-verdicts", "zzz.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, unreadable := civerdict.LoadAll(root)
	if len(loaded) != 1 || loaded[0].Record != good {
		t.Errorf("loaded = %+v, want the one good record", loaded)
	}
	if len(unreadable) != 1 || !strings.HasSuffix(unreadable[0], "zzz.json") {
		t.Errorf("unreadable = %v, want zzz.json named", unreadable)
	}
}

func mustJSON(t *testing.T, r civerdict.Record) []byte {
	t.Helper()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
