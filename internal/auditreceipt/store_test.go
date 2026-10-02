package auditreceipt

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func requiredTree(t *testing.T, gate string) string {
	t.Helper()
	root := t.TempDir()
	if gate != "" {
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		body := "workflow:\n  audit:\n    gates:\n      codex: " + gate + "\n"
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	return canonical
}

// The gate is read raw and matched case-sensitively after trimming surrounding
// whitespace (the resolver's reading, SPEC-AUDIT-MODEL-CONVERGE-001 EC-1): only
// the literal `required` opts in, and a padded value reads like the trimmed one
// on every audit surface.
func TestCodexGateRequired_ExactMatchOnly(t *testing.T) {
	for _, tc := range []struct {
		gate string
		want bool
	}{
		{"required", true},
		{"advisory", false},
		{"off", false},
		{"", false},
		{"\"required \"", true},
		{"REQUIRED", false},
	} {
		if got := CodexGateRequired(requiredTree(t, tc.gate)); got != tc.want {
			t.Errorf("CodexGateRequired(gate=%q) = %v, want %v", tc.gate, got, tc.want)
		}
	}
}

// workflowYAMLTree writes body as the tree's workflow.yaml (an empty body writes
// no file) and returns the canonical root.
func workflowYAMLTree(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if body != "" {
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	return canonical
}

// The predicate follows the audit plan (SPEC-AUDIT-MODEL-CONVERGE-001 AC-ACV-010):
// an audit.model token that requires codex opts the tree in, an explicit gate
// overrides the token, and the default profile and an invalid value do not.
func TestCodexGateRequired_FromModelToken(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want bool
	}{
		{"multi", "workflow:\n  audit:\n    model: multi\n", true},
		{"codex", "workflow:\n  audit:\n    model: codex\n", true},
		{"padded token", "workflow:\n  audit:\n    model: \" multi \"\n", true},
		{"explicit gate required", "workflow:\n  audit:\n    gates:\n      codex: required\n", true},
		{"claude token leaves codex on the default", "workflow:\n  audit:\n    model: claude\n", false},
		{"glm token turns codex off", "workflow:\n  audit:\n    model: glm\n", false},
		{"empty token", "workflow:\n  audit:\n    model: \"\"\n", false},
		{"pins only", "workflow:\n  audit:\n    codex:\n      model: gpt-6.1-sol\n      effort: high\n", false},
		{"no file", "", false},
		{"multi with advisory override", "workflow:\n  audit:\n    model: multi\n    gates:\n      codex: advisory\n", false},
		{"multi with off override", "workflow:\n  audit:\n    model: multi\n    gates:\n      codex: \"off\"\n", false},
		{"unknown token reads as not required", "workflow:\n  audit:\n    model: grok\n", false},
		{"unknown gate reads as not required", "workflow:\n  audit:\n    model: multi\n    gates:\n      codex: requird\n", false},
		{"unparseable file", "workflow: [unterminated\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CodexGateRequired(workflowYAMLTree(t, tc.yaml)); got != tc.want {
				t.Errorf("CodexGateRequired(%q) = %v, want %v", tc.yaml, got, tc.want)
			}
		})
	}
}

// A receipt round-trips through the store and carries an opaque rcpt- id.
func TestWriteReceipt_RoundTrip(t *testing.T) {
	root := requiredTree(t, "required")
	r := Receipt{Tool: ToolCodexAudit, TreeRoot: root, RootSource: RootSourceArgument, CodexVerdict: "fail"}
	id, err := WriteReceipt(root, &r)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	if !receiptIDPattern.MatchString(id) {
		t.Fatalf("receipt id %q does not match the opaque rcpt- token shape", id)
	}
	got, err := ReadReceipt(root, id)
	if err != nil {
		t.Fatalf("ReadReceipt: %v", err)
	}
	if got.Tool != ToolCodexAudit || got.TreeRoot != root || got.RootSource != RootSourceArgument {
		t.Errorf("round-tripped receipt = %+v", got)
	}
	if got.CreatedAt.IsZero() {
		t.Error("created_at is zero — the store stamps the creation time")
	}
}

// A corrupt receipt file is reported as unreadable, never as absent.
func TestReadReceipt_CorruptIsUnreadable(t *testing.T) {
	root := requiredTree(t, "required")
	if err := os.MkdirAll(filepath.Join(root, stateRel, receiptsRel), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	id := "rcpt-" + "0123456789abcdef0123"
	if err := os.WriteFile(filepath.Join(root, stateRel, receiptsRel, id+".json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := ReadReceipt(root, id); err == nil || os.IsNotExist(err) {
		t.Errorf("ReadReceipt on a corrupt file = %v, want a non-NotExist error", err)
	}
}

// The verdict line grammar (plan.md §B.6).
func TestParseVerdictLine(t *testing.T) {
	valid := "AUDIT-VERDICT: PASS spec=SPEC-CODEX-AUDIT-GATE-AXES-001 receipts=rcpt-0123456789abcdef0123,rcpt-abcdefabcdefabcdefab"
	for name, tc := range map[string]struct {
		msg      string
		ok       bool
		verdict  string
		spec     string
		receipts int
	}{
		"plain":            {valid, true, "PASS", "SPEC-CODEX-AUDIT-GATE-AXES-001", 2},
		"trailing-blank":   {valid + "\n\n  \n", true, "PASS", "SPEC-CODEX-AUDIT-GATE-AXES-001", 2},
		"trailing-cr":      {valid + " \r", true, "PASS", "SPEC-CODEX-AUDIT-GATE-AXES-001", 2},
		"leading-space":    {"   " + valid, true, "PASS", "SPEC-CODEX-AUDIT-GATE-AXES-001", 2},
		"pass-with-debt":   {"AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-X-001 receipts=none", true, "PASS-WITH-DEBT", "SPEC-X-001", 0},
		"fail":             {"AUDIT-VERDICT: FAIL spec=SPEC-X-001 receipts=none", true, "FAIL", "SPEC-X-001", 0},
		"not-last-line":    {valid + "\nsome trailing prose", false, "", "", 0},
		"empty":            {"   \n\n", false, "", "", 0},
		"malformed-spec":   {"AUDIT-VERDICT: PASS spec=nope receipts=none", false, "", "", 0},
		"malformed-prefix": {"VERDICT: PASS spec=SPEC-X-001 receipts=none", false, "", "", 0},
	} {
		t.Run(name, func(t *testing.T) {
			got, ok := ParseVerdictLine(tc.msg)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (line %+v)", ok, tc.ok, got)
			}
			if !ok {
				return
			}
			if got.Verdict != tc.verdict || got.SpecID != tc.spec || len(got.Receipts) != tc.receipts {
				t.Errorf("parsed = %+v, want verdict %q spec %q receipts %d", got, tc.verdict, tc.spec, tc.receipts)
			}
			if tc.verdict != "FAIL" && !got.IsPass() {
				t.Error("IsPass() = false for a PASS-class verdict")
			}
		})
	}
}

// CheckCitedReceipts reports the FIRST failing condition in the documented order.
func TestCheckCitedReceipts_Causes(t *testing.T) {
	root := requiredTree(t, "required")
	other := requiredTree(t, "required")
	start := StartMarker{AgentID: "a1", AgentType: AgentPlanAuditor, TreeRoot: root, StartedAt: time.Now().UTC()}

	after := Receipt{Tool: ToolCodexAudit, TreeRoot: root, CreatedAt: start.StartedAt.Add(time.Second)}
	idAfter, err := WriteReceipt(root, &after)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	wrongTree := Receipt{Tool: ToolCodexAudit, TreeRoot: other, CreatedAt: start.StartedAt.Add(time.Second)}
	idWrongTree, err := WriteReceipt(root, &wrongTree)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	stale := Receipt{Tool: ToolCodexAudit, TreeRoot: root, CreatedAt: start.StartedAt.Add(-time.Minute)}
	idStale, err := WriteReceipt(root, &stale)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}

	for name, tc := range map[string]struct {
		start     *StartMarker
		receipts  []string
		wantOK    bool
		wantCause string
	}{
		"valid":            {&start, []string{idAfter}, true, ""},
		"one-of-many":      {&start, []string{idStale, idAfter}, true, ""},
		"start-missing":    {nil, []string{idAfter}, false, CauseStartMarkerMissing},
		"no-citation":      {&start, nil, false, CauseNoReceiptCited},
		"unknown-receipt":  {&start, []string{"rcpt-ffffffffffffffffffff"}, false, CauseReceiptUnknown},
		"different-tree":   {&start, []string{idWrongTree}, false, CauseReceiptOtherTree},
		"before-the-start": {&start, []string{idStale}, false, CauseReceiptBeforeStart},
	} {
		t.Run(name, func(t *testing.T) {
			ok, cause := CheckCitedReceipts(root, tc.start, tc.receipts)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v (cause %q), want %v", ok, cause, tc.wantOK)
			}
			if !ok && cause != tc.wantCause {
				t.Errorf("cause = %q, want %q", cause, tc.wantCause)
			}
		})
	}
}

// Rejection records: write, list, role-wide clear, corrupt reporting.
func TestRejections_LifecycleAndCorruption(t *testing.T) {
	root := requiredTree(t, "required")
	for _, rj := range []Rejection{
		{AgentType: AgentPlanAuditor, SpecID: "SPEC-X-001", Cause: CauseNoReceiptCited},
		{AgentType: AgentPlanAuditor, SpecID: UnknownSpec, Cause: CauseVerdictLineMissing},
		{AgentType: AgentSyncAuditor, SpecID: "SPEC-X-001", Cause: CauseNoReceiptCited},
	} {
		rj := rj
		if err := WriteRejection(root, &rj); err != nil {
			t.Fatalf("WriteRejection: %v", err)
		}
	}
	got, err := ListRejections(root)
	if err != nil {
		t.Fatalf("ListRejections: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ListRejections = %d records, want 3", len(got))
	}

	if err := ClearRejectionsForRole(root, AgentPlanAuditor); err != nil {
		t.Fatalf("ClearRejectionsForRole: %v", err)
	}
	got, err = ListRejections(root)
	if err != nil {
		t.Fatalf("ListRejections after clear: %v", err)
	}
	if len(got) != 1 || got[0].AgentType != AgentSyncAuditor {
		t.Errorf("after the plan-auditor clear, remaining = %+v, want only the sync-auditor record", got)
	}

	// A corrupt record is an error naming the file — never a silent "no rejection".
	bad := filepath.Join(root, stateRel, rejectionsRel, "plan-auditor--SPEC-Y-001.json")
	if err := os.WriteFile(bad, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := ListRejections(root); err == nil {
		t.Error("ListRejections on a corrupt record returned no error")
	}

	// An absent rejections directory is zero outstanding rejections, not an error.
	empty := requiredTree(t, "required")
	got, err = ListRejections(empty)
	if err != nil || len(got) != 0 {
		t.Errorf("ListRejections on a tree with no rejections dir = (%v, %v), want (empty, nil)", got, err)
	}
}

// Start markers round-trip and are removable.
func TestStartMarker_RoundTripAndRemove(t *testing.T) {
	root := requiredTree(t, "required")
	m := StartMarker{AgentID: "agent-1", AgentType: AgentSyncAuditor, SessionID: "s1", TreeRoot: root, StartedAt: time.Now().UTC()}
	if err := WriteStartMarker(root, &m); err != nil {
		t.Fatalf("WriteStartMarker: %v", err)
	}
	got, err := ReadStartMarker(root, "agent-1")
	if err != nil {
		t.Fatalf("ReadStartMarker: %v", err)
	}
	if got.AgentType != AgentSyncAuditor || got.TreeRoot != root {
		t.Errorf("round-tripped marker = %+v", got)
	}
	if err := RemoveStartMarker(root, "agent-1"); err != nil {
		t.Fatalf("RemoveStartMarker: %v", err)
	}
	if _, err := ReadStartMarker(root, "agent-1"); !os.IsNotExist(err) {
		t.Errorf("ReadStartMarker after removal = %v, want a NotExist error", err)
	}
}
