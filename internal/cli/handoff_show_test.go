package cli

// SPEC-HANDOFF-NEUTRAL-001 M1.2 — `moai handoff show` tests (AC-HN-001..004,
// 004b). The show verb is the harness-neutral P3 path: it re-prints the saved
// 6-block body verbatim and never mutates row state (REQ-HN-002).
//
// Fixture contract (acceptance.md): every pending/consumed fixture is created
// via handoff.SavePending or the claim/consume state machine against a
// t.TempDir() project. The ONLY file fixture is the explicit legacy-compat
// case (AC-HN-004b), which exists to exercise ReadPending's read-only
// legacy branch.

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

const handoffShowTestBody = "✂──── 여기부터 복사 ────✂\nultrathink. SPEC-X run 진입.\nRun: /moai run SPEC-X\n✂──── 여기까지 복사 ────✂"

// saveShowFixture writes one pending resume row through SavePending (the
// fixture contract's sanctioned producer).
func saveShowFixture(t *testing.T, pd, body, spec, phase, lang string) {
	t.Helper()
	rec := &handoff.PendingRecord{
		SchemaVersion:        handoff.PendingSchemaVersion,
		SpecID:               spec,
		Phase:                phase,
		SavedAt:              time.Now(),
		ConversationLanguage: lang,
		Body:                 body,
	}
	if err := handoff.SavePending(pd, rec); err != nil {
		t.Fatalf("SavePending fixture: %v", err)
	}
}

// consumeShowFixture drives the full state machine pending → claimed →
// consumed so the show fallback source exists.
func consumeShowFixture(t *testing.T, pd string) {
	t.Helper()
	rec, id, ok, err := handoff.ClaimPending(pd, "show-test-token")
	if err != nil || !ok {
		t.Fatalf("ClaimPending fixture: ok=%v err=%v", ok, err)
	}
	if rec == nil {
		t.Fatal("ClaimPending fixture: nil record")
	}
	if err := handoff.FinishClaim(pd, id, true, "consumed by show test", "show-test-token"); err != nil {
		t.Fatalf("FinishClaim fixture: %v", err)
	}
}

func TestHandoffShow_PendingSource(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	saveShowFixture(t, pd, handoffShowTestBody, "SPEC-X", "run", "ko")

	out, err := runHandoff(t, "", "show", "--project-dir", pd)
	if err != nil {
		t.Fatalf("handoff show: %v (out: %s)", err, out)
	}
	if !strings.Contains(out, handoffShowTestBody) {
		t.Errorf("show must reprint the body verbatim; got %q", out)
	}
	if !strings.Contains(out, "pending") {
		t.Errorf("header must name the pending source; got %q", out)
	}
}

func TestHandoffShow_ConsumedFallback(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	saveShowFixture(t, pd, "consumed body 6-block", "SPEC-Y", "sync", "en")
	consumeShowFixture(t, pd)

	// The pending branch must be empty — the fallback is what is under test.
	if _, present, err := handoff.ReadPending(pd); err != nil || present {
		t.Fatalf("fixture must leave no pending row: present=%v err=%v", present, err)
	}

	out, err := runHandoff(t, "", "show", "--project-dir", pd)
	if err != nil {
		t.Fatalf("handoff show: %v (out: %s)", err, out)
	}
	if !strings.Contains(out, "consumed body 6-block") {
		t.Errorf("show must reprint the latest consumed body verbatim; got %q", out)
	}
	if !strings.Contains(out, "consumed") {
		t.Errorf("header must name the consumed source; got %q", out)
	}
}

func TestHandoffShow_NoHandoffErrors(t *testing.T) {
	t.Parallel()

	pd := t.TempDir() // empty project: neither a pending row nor consumed history
	_, err := runHandoff(t, "", "show", "--project-dir", pd)
	if err == nil {
		t.Fatal("show on an empty project must fail (exit 1), got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no saved handoff") {
		t.Errorf("error must carry the English guidance; got %q", msg)
	}
	if !strings.Contains(msg, "저장된 핸드오프") {
		t.Errorf("error must carry the Korean guidance; got %q", msg)
	}
}

func TestHandoffShow_DoesNotMutateState(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	saveShowFixture(t, pd, handoffShowTestBody, "SPEC-X", "run", "ko")
	pending, present, err := handoff.ReadPending(pd)
	if err != nil || !present || pending == nil {
		t.Fatalf("fixture pending row: present=%v err=%v", present, err)
	}

	snapshot := func() (string, string, string) {
		t.Helper()
		db, openErr := homestate.OpenFactory(pd)
		if openErr != nil {
			t.Fatalf("OpenFactory: %v", openErr)
		}
		defer func() { _ = db.Close() }()
		var status string
		var consumed sql.NullString
		var body string
		if err := db.DB.QueryRow(`SELECT status, consumed_at, body FROM resume_handoffs WHERE id=?`, pending.ID).Scan(&status, &consumed, &body); err != nil {
			t.Fatalf("snapshot row: %v", err)
		}
		consumedStr := ""
		if consumed.Valid {
			consumedStr = consumed.String
		}
		return status, consumedStr, body
	}

	bStatus, bConsumed, bBody := snapshot()
	out1, err1 := runHandoff(t, "", "show", "--project-dir", pd)
	out2, err2 := runHandoff(t, "", "show", "--project-dir", pd)
	if err1 != nil || err2 != nil {
		t.Fatalf("show runs: %v / %v", err1, err2)
	}
	if out1 != out2 {
		t.Error("repeated show must produce identical output (read-only contract)")
	}
	aStatus, aConsumed, aBody := snapshot()
	if bStatus != aStatus || bConsumed != aConsumed || bBody != aBody {
		t.Errorf("row mutated by show: status %q→%q consumed_at %q→%q body changed=%v",
			bStatus, aStatus, bConsumed, aConsumed, bBody != aBody)
	}
}

func TestHandoffShow_JSONOutput(t *testing.T) {
	t.Parallel()

	// Pending source.
	pdPending := t.TempDir()
	saveShowFixture(t, pdPending, "json pending body", "SPEC-J", "plan", "ko")
	out, err := runHandoff(t, "", "show", "--project-dir", pdPending, "--json")
	if err != nil {
		t.Fatalf("show --json (pending): %v (out: %s)", err, out)
	}
	var payload struct {
		Source string `json:"source"`
		Record struct {
			Body   string `json:"body"`
			SpecID string `json:"spec_id"`
			Phase  string `json:"phase"`
		} `json:"record"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("show --json must print one JSON object; got %q (%v)", out, err)
	}
	if payload.Source != "pending" {
		t.Errorf("source: got %q, want pending", payload.Source)
	}
	if payload.Record.Body != "json pending body" || payload.Record.SpecID != "SPEC-J" || payload.Record.Phase != "plan" {
		t.Errorf("record fields: got %+v", payload.Record)
	}

	// Consumed fallback source.
	pdConsumed := t.TempDir()
	saveShowFixture(t, pdConsumed, "json consumed body", "SPEC-J2", "run", "en")
	consumeShowFixture(t, pdConsumed)
	out, err = runHandoff(t, "", "show", "--project-dir", pdConsumed, "--json")
	if err != nil {
		t.Fatalf("show --json (consumed): %v (out: %s)", err, out)
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("show --json (consumed) parse: %v (%q)", err, out)
	}
	if payload.Source != "consumed" {
		t.Errorf("source: got %q, want consumed", payload.Source)
	}
	if payload.Record.Body != "json consumed body" {
		t.Errorf("record body: got %q", payload.Record.Body)
	}
}

func TestHandoffShow_LocaleHeader(t *testing.T) {
	t.Parallel()

	body := "ultrathink. SPEC-L run 진입.\n실행: /moai run SPEC-L"

	pdKo := t.TempDir()
	saveShowFixture(t, pdKo, body, "SPEC-L", "run", "ko")
	outKo, err := runHandoff(t, "", "show", "--project-dir", pdKo)
	if err != nil {
		t.Fatalf("show (ko): %v (out: %s)", err, outKo)
	}
	if !strings.Contains(outKo, "출처: pending") {
		t.Errorf("ko header must use the Korean label set; got %q", outKo)
	}
	if !strings.Contains(firstLine(outKo), "핸드오프") {
		t.Errorf("ko title must be Korean; got %q", firstLine(outKo))
	}
	if strings.Contains(outKo, "Source:") {
		t.Errorf("ko header must not carry English labels; got %q", outKo)
	}
	if !strings.Contains(outKo, body) {
		t.Errorf("ko body must be verbatim; got %q", outKo)
	}

	pdEn := t.TempDir()
	saveShowFixture(t, pdEn, body, "SPEC-L", "run", "en")
	outEn, err := runHandoff(t, "", "show", "--project-dir", pdEn)
	if err != nil {
		t.Fatalf("show (en): %v (out: %s)", err, outEn)
	}
	if !strings.Contains(outEn, "Source: pending") {
		t.Errorf("en header must use the English label set; got %q", outEn)
	}
	if !strings.Contains(firstLine(outEn), "handoff") {
		t.Errorf("en title must be English; got %q", firstLine(outEn))
	}
	if !strings.Contains(outEn, body) {
		t.Errorf("en body must be verbatim; got %q", outEn)
	}
}

func TestHandoffShow_LegacyCompatRead(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	// The one sanctioned file fixture: a pre-SQLite pending.json against an
	// otherwise empty factory.db exercises ReadPending's read-only legacy
	// branch (AC-HN-004b).
	rec := handoff.PendingRecord{
		SchemaVersion:        handoff.PendingSchemaVersion,
		SpecID:               "SPEC-LEGACY",
		Phase:                "run",
		SavedAt:              time.Now(),
		ConversationLanguage: "ko",
		Body:                 "legacy compat body 6-block",
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal legacy fixture: %v", err)
	}
	legacyDir := filepath.Join(pd, ".moai", "state", "handoff")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "pending.json"), raw, 0o600); err != nil {
		t.Fatalf("write legacy fixture: %v", err)
	}

	out, err := runHandoff(t, "", "show", "--project-dir", pd)
	if err != nil {
		t.Fatalf("handoff show (legacy): %v (out: %s)", err, out)
	}
	if !strings.Contains(out, "legacy compat body 6-block") {
		t.Errorf("legacy compat path must reprint the legacy body; got %q", out)
	}
}

// firstLine returns the first line of s (header assertions stay local).
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
