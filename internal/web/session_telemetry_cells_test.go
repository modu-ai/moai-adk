package web

// session_telemetry_cells_test.go — SPEC-WEB-CONSOLE-015 M2: the three RoleVM
// cells the chain view left blank.
//
// The values arrive through the ONE reader SPEC-SESSION-TELEMETRY-001 exports
// (statusline.ReadSessionTelemetry, addressed by statusline.SessionTelemetryPath).
// This package declares no second reader and no copy of that record's schema —
// two declarations of one on-disk format is how a format forks.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/statusline"
)

// writeTelemetry writes one session's telemetry record beneath root, through
// the sibling SPEC's own path helper so the location is never reconstructed here.
func writeTelemetry(t *testing.T, root string, rec statusline.SessionTelemetryRecord) {
	t.Helper()
	stateDir := filepath.Join(root, ".moai", "state")
	path := statusline.SessionTelemetryPath(stateDir, rec.SessionID)
	if path == "" {
		t.Fatalf("SessionTelemetryPath refused key %q", rec.SessionID)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir telemetry dir: %v", err)
	}
	body, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal telemetry: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write telemetry: %v", err)
	}
}

// roleByName finds the chain row for a role, failing when it is absent.
func roleByName(t *testing.T, roles []RoleVM, role string) RoleVM {
	t.Helper()
	for _, r := range roles {
		if r.Role == role {
			return r
		}
	}
	t.Fatalf("role %q absent from %d rows", role, len(roles))
	return RoleVM{}
}

// TestChainCellsCarryTelemetryValues — AC-WC15-021b: a role bound to a session
// with a telemetry record shows that record's model, effort and context
// percentage, where the pre-change tree hard-coded "", "" and -1.
func TestChainCellsCarryTelemetryValues(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-run", pid)})
	writeKanbanRecord(t, root, kanban.Record{
		SessionID: "sess-run", Role: "run", Backend: kanban.BackendClaude,
	})
	writeTelemetry(t, root, statusline.SessionTelemetryRecord{
		SchemaVersion: 2, SessionID: "sess-run", WriterPID: pid,
		ContextWindowSize: 1000000, TokensUsed: 420000, RawPct: 42,
		Model: "claude-opus-5", Effort: "xhigh",
	})

	_, byID := loadSessions(root, time.Now())
	chain := buildChain(root, loadKanbanRecords(root), byID, "")

	row := roleByName(t, chain.Roles, "run")
	if row.Model != "claude-opus-5" {
		t.Errorf("Model = %q, want claude-opus-5", row.Model)
	}
	if row.Effort != "xhigh" {
		t.Errorf("Effort = %q, want xhigh", row.Effort)
	}
	if row.ContextPct != 42 {
		t.Errorf("ContextPct = %d, want 42", row.ContextPct)
	}
}

// TestChainCellsDoNotBorrowAnotherSessionsValues — AC-WC15-023: two sessions,
// only one with a record. The other shows the not-recorded sentinels — never
// the first session's values. This is the direct regression test for the
// single-slot last-writer-wins race the per-session split removed.
func TestChainCellsDoNotBorrowAnotherSessionsValues(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeActiveSessions(t, root, []session.Entry{
		liveEntry("sess-a", pid),
		liveEntry("sess-b", pid+100000),
	})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-a", Role: "plan", Backend: kanban.BackendClaude})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-b", Role: "run", Backend: kanban.BackendClaude})
	writeTelemetry(t, root, statusline.SessionTelemetryRecord{
		SchemaVersion: 2, SessionID: "sess-a",
		ContextWindowSize: 200000, TokensUsed: 110000, RawPct: 55,
		Model: "claude-sonnet-5", Effort: "high",
	})

	_, byID := loadSessions(root, time.Now())
	chain := buildChain(root, loadKanbanRecords(root), byID, "")

	a := roleByName(t, chain.Roles, "plan")
	if a.Model != "claude-sonnet-5" || a.Effort != "high" || a.ContextPct != 55 {
		t.Fatalf("recorded session lost its own values: %+v", a)
	}
	b := roleByName(t, chain.Roles, "run")
	if b.Model != "" || b.Effort != "" || b.ContextPct != -1 {
		t.Errorf("unrecorded session borrowed values: Model=%q Effort=%q ContextPct=%d",
			b.Model, b.Effort, b.ContextPct)
	}
}

// TestChainCellsTolerateSchemaV1Record — AC-WC15-051 / REQ-WC15-051: a record
// written before the dependency landed carries no model and no effort. Those
// cells read as not-recorded while the context percentage it DOES carry still
// renders — and the row renders rather than failing.
func TestChainCellsTolerateSchemaV1Record(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeActiveSessions(t, root, []session.Entry{
		liveEntry("sess-old", pid),
		liveEntry("sess-new", pid+100001),
	})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-old", Role: "lead", Backend: kanban.BackendClaude})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-new", Role: "sync", Backend: kanban.BackendClaude})
	// Pre-dependency shape: schema 1, no model, no effort.
	writeTelemetry(t, root, statusline.SessionTelemetryRecord{
		SchemaVersion: 1, SessionID: "sess-old",
		ContextWindowSize: 200000, TokensUsed: 40000, RawPct: 20,
	})
	writeTelemetry(t, root, statusline.SessionTelemetryRecord{
		SchemaVersion: 2, SessionID: "sess-new",
		ContextWindowSize: 1000000, TokensUsed: 330000, RawPct: 33,
		Model: "glm-5.3", Effort: "medium",
	})

	_, byID := loadSessions(root, time.Now())
	chain := buildChain(root, loadKanbanRecords(root), byID, "")

	old := roleByName(t, chain.Roles, "lead")
	if old.Model != "" || old.Effort != "" {
		t.Errorf("pre-dependency record invented model/effort: %+v", old)
	}
	if old.ContextPct != 20 {
		t.Errorf("pre-dependency ContextPct = %d, want 20", old.ContextPct)
	}
	fresh := roleByName(t, chain.Roles, "sync")
	if fresh.Model != "glm-5.3" || fresh.Effort != "medium" || fresh.ContextPct != 33 {
		t.Errorf("post-dependency row did not carry its values: %+v", fresh)
	}
}

// TestChainCellsUnreadableTelemetryStaysBlank — REQ-WC15-023: an unparseable
// record is "no readable record", not a zero. The cells stay at the
// not-recorded sentinels and nothing fails.
func TestChainCellsUnreadableTelemetryStaysBlank(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-bad", pid)})
	writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-bad", Role: "run", Backend: kanban.BackendClaude})

	path := statusline.SessionTelemetryPath(filepath.Join(root, ".moai", "state"), "sess-bad")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, byID := loadSessions(root, time.Now())
	row := roleByName(t, buildChain(root, loadKanbanRecords(root), byID, "").Roles, "run")
	if row.Model != "" || row.Effort != "" || row.ContextPct != -1 {
		t.Errorf("unparseable record produced values: %+v", row)
	}
}

// TestChainCellsClampContextPercentage — the gauge draws a width from this
// value, so a record carrying an out-of-range percentage must not escape the
// meter's own range. A NEGATIVE percentage is not a recorded reading at all,
// so it takes the console's existing not-recorded sentinel (-1) rather than
// being asserted as 0% — clampPct (render_helpers.go) already owns that
// vocabulary and is reused rather than restated.
func TestChainCellsClampContextPercentage(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  float64
		want int
	}{
		{"over", 140, 100},
		{"nonsense negative reads as not recorded", -5, -1},
		{"rounds down", 42.9, 42},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			pid := os.Getpid()
			writeActiveSessions(t, root, []session.Entry{liveEntry("sess-c", pid)})
			writeKanbanRecord(t, root, kanban.Record{SessionID: "sess-c", Role: "run", Backend: kanban.BackendClaude})
			writeTelemetry(t, root, statusline.SessionTelemetryRecord{
				SchemaVersion: 2, SessionID: "sess-c",
				ContextWindowSize: 200000, TokensUsed: 1, RawPct: tc.raw,
			})
			_, byID := loadSessions(root, time.Now())
			row := roleByName(t, buildChain(root, loadKanbanRecords(root), byID, "").Roles, "run")
			if row.ContextPct != tc.want {
				t.Errorf("ContextPct = %d, want %d", row.ContextPct, tc.want)
			}
		})
	}
}
