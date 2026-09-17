package auditreceipt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Only the two auditors are guarded roles.
func TestIsAuditorAgent(t *testing.T) {
	for agent, want := range map[string]bool{
		AgentPlanAuditor:  true,
		AgentSyncAuditor:  true,
		"manager-develop": false,
		"":                false,
	} {
		if got := IsAuditorAgent(agent); got != want {
			t.Errorf("IsAuditorAgent(%q) = %v, want %v", agent, got, want)
		}
	}
}

// A single rejection reads back by role and SPEC; an absent one is NotExist.
func TestReadRejection_PresentAndAbsent(t *testing.T) {
	root := requiredTree(t, "required")
	rj := Rejection{AgentType: AgentPlanAuditor, SpecID: "SPEC-R-001", Cause: CauseNoReceiptCited}
	if err := WriteRejection(root, &rj); err != nil {
		t.Fatalf("WriteRejection: %v", err)
	}
	got, err := ReadRejection(root, AgentPlanAuditor, "SPEC-R-001")
	if err != nil {
		t.Fatalf("ReadRejection: %v", err)
	}
	if got.Cause != CauseNoReceiptCited || got.RejectedAt.IsZero() {
		t.Errorf("rejection = %+v, want the cause and a stamped time", got)
	}
	if _, err := ReadRejection(root, AgentSyncAuditor, "SPEC-R-001"); !os.IsNotExist(err) {
		t.Errorf("absent rejection error = %v, want NotExist", err)
	}
}

// A receipt whose tool is neither codex_audit nor a codex-carrying audit_multi
// is not evidence that a codex audit ran.
func TestCheckCitedReceipts_NonCodexToolRejected(t *testing.T) {
	root := requiredTree(t, "required")
	start := StartMarker{AgentID: "a1", AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: Now()}
	r := Receipt{Tool: "glm_audit", TreeRoot: root, CreatedAt: start.StartedAt.Add(time.Second)}
	id, err := WriteReceipt(root, &r)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	ok, cause := CheckCitedReceipts(root, &start, []string{id})
	if ok || cause != CauseReceiptNotACodexAudit {
		t.Errorf("(ok, cause) = (%v, %q), want (false, %q)", ok, cause, CauseReceiptNotACodexAudit)
	}
}

// A malformed receipt id is refused before it reaches the filesystem, so a
// citation cannot be used to read an arbitrary path.
func TestReadReceipt_MalformedIDRejected(t *testing.T) {
	root := requiredTree(t, "required")
	for _, id := range []string{"", "../../etc/passwd", "rcpt-SHORT", "notrcpt-0123456789abcdef0123"} {
		if _, err := ReadReceipt(root, id); err == nil {
			t.Errorf("ReadReceipt(%q) returned no error", id)
		}
	}
}

// Runtime-supplied identifiers cannot escape the store directory.
func TestStartMarker_AgentIDIsSanitized(t *testing.T) {
	root := requiredTree(t, "required")
	m := StartMarker{AgentID: "../../escape", AgentType: AgentPlanAuditor, TreeRoot: root}
	if err := WriteStartMarker(root, &m); err != nil {
		t.Fatalf("WriteStartMarker: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(StateDir(root), startsRel))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("marker file names = %v, want exactly one", entries)
	}
	name := entries[0].Name()
	if strings.ContainsRune(name, os.PathSeparator) || name != filepath.Base(name) {
		t.Fatalf("marker file name %q is not a flat name inside the store", name)
	}
	if _, err := ReadStartMarker(root, "../../escape"); err != nil {
		t.Errorf("ReadStartMarker on the sanitized key: %v", err)
	}
}

// An empty SPEC id keys the unknown-spec record rather than an unnamed file.
func TestRejectionFileName_EmptySpecBecomesUnknownSpec(t *testing.T) {
	if got := rejectionFileName(AgentPlanAuditor, "  "); got != AgentPlanAuditor+"--"+UnknownSpec+".json" {
		t.Errorf("rejectionFileName with a blank spec = %q", got)
	}
	if got := sanitizeKey(".."); got != "_" {
		t.Errorf("sanitizeKey(\"..\") = %q, want the neutral placeholder", got)
	}
}

// A store path that cannot be created surfaces as an error rather than a silent
// success — the caller decides what an unrecordable receipt means.
func TestWriteReceipt_UnwritableStoreErrors(t *testing.T) {
	root := t.TempDir()
	// A regular file where the store directory must go.
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "state", "audit-receipts"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	r := Receipt{Tool: ToolCodexAudit, TreeRoot: root}
	if _, err := WriteReceipt(root, &r); err == nil {
		t.Error("WriteReceipt into an unwritable store returned no error")
	}
}

// Clearing a role in a tree with no store at all is a no-op, not an error.
func TestClearRejectionsForRole_NoStoreIsNoOp(t *testing.T) {
	if err := ClearRejectionsForRole(t.TempDir(), AgentPlanAuditor); err != nil {
		t.Errorf("ClearRejectionsForRole with no store: %v", err)
	}
}

// Canonical is total: an empty path answers empty, a non-existent one answers
// the cleaned path rather than failing.
func TestCanonical_EmptyAndMissingPaths(t *testing.T) {
	if got := Canonical("  "); got != "" {
		t.Errorf("Canonical(blank) = %q, want empty", got)
	}
	missing := filepath.Join(t.TempDir(), "nope", "deeper")
	if got := Canonical(missing); got != filepath.Clean(missing) {
		t.Errorf("Canonical(missing) = %q, want the cleaned path", got)
	}
}

// Receipt ids are unique per mint.
func TestNewReceiptID_Unique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		id, err := NewReceiptID()
		if err != nil {
			t.Fatalf("NewReceiptID: %v", err)
		}
		if !receiptIDPattern.MatchString(id) || seen[id] {
			t.Fatalf("id %q is malformed or repeated", id)
		}
		seen[id] = true
	}
}
