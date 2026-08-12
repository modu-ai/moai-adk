package fix

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestDraftRequest_JSONSchema verifies the request.json schema contract
// (plan.md §C.3, AC-NS5-004a): the top-level keys are exactly provenance,
// diff_scope, work_item_refs, draft_instructions; the provenance block carries
// fix_commit_sha + baseline_commit_sha + captured_at; and NO wall-clock field
// (generated_at / created_at / timestamp) appears anywhere.
func TestDraftRequest_JSONSchema(t *testing.T) {
	req := DraftRequest{
		Provenance: Provenance{
			FixCommitSHA:      "abc1234",
			BaselineCommitSHA: "def5678",
			CapturedAt:        "2026-08-12T10:00:00+00:00",
		},
		DiffScope: []DiffScopeEntry{
			{
				DocSurface:  "capability-symbols.json",
				SubtreeID:   "pkg.ParseHeader",
				StaleReason: "git-diff",
			},
		},
		WorkItemRefs: []WorkItemRef{
			{SourceKind: "detect", OwnerPath: "/proj/internal/auth/parse.go", Action: "re-link @NAV:SYM symbol"},
		},
		DraftInstructions: DraftInstructions{
			PerSubtree: []SubtreeStrategy{
				{SubtreeID: "pkg.ParseHeader", Strategy: "re-link @NAV:SYM symbol to renamed declaration"},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal DraftRequest: %v", err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatalf("unmarshal top-level: %v", err)
	}

	for _, key := range []string{"provenance", "diff_scope", "work_item_refs", "draft_instructions"} {
		if _, ok := top[key]; !ok {
			t.Errorf("DraftRequest JSON missing top-level key %q (AC-NS5-004a schema)", key)
		}
	}

	// Provenance sub-keys (AC-NS5-004a).
	var prov struct {
		FixCommitSHA      string `json:"fix_commit_sha"`
		BaselineCommitSHA string `json:"baseline_commit_sha"`
		CapturedAt        string `json:"captured_at"`
	}
	if err := json.Unmarshal(top["provenance"], &prov); err != nil {
		t.Fatalf("unmarshal provenance: %v", err)
	}
	if prov.FixCommitSHA != "abc1234" {
		t.Errorf("provenance.fix_commit_sha = %q, want abc1234", prov.FixCommitSHA)
	}
	if prov.BaselineCommitSHA != "def5678" {
		t.Errorf("provenance.baseline_commit_sha = %q, want def5678", prov.BaselineCommitSHA)
	}
	if prov.CapturedAt == "" {
		t.Error("provenance.captured_at is empty")
	}

	// NO wall-clock field (AC-NS5-004a): the canonical field is captured_at
	// (sourced from git), never generated_at / created_at / timestamp.
	raw := string(data)
	for _, forbidden := range []string{`"generated_at"`, `"created_at"`, `"timestamp"`} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("DraftRequest JSON contains forbidden wall-clock field %s (AC-NS5-004a: no wall-clock)", forbidden)
		}
	}
}

// TestDiffScopeEntry_OmitEmptyWorkItemRef verifies that a DiffScopeEntry with
// no M2 work-item (git-diff-only or M1-only seed) omits the work_item_ref key
// entirely (omitempty), so the JSON stays clean for non-M2-seeded subtrees.
func TestDiffScopeEntry_OmitEmptyWorkItemRef(t *testing.T) {
	entry := DiffScopeEntry{
		DocSurface:  "audit-report.json",
		SubtreeID:   "SPEC-AUTH-001",
		StaleReason: "git-diff",
		// WorkItemRef intentionally nil — a git-diff-only seed has no M2 ref.
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), "work_item_ref") {
		t.Errorf("DiffScopeEntry with nil WorkItemRef should omit the key (omitempty); got %s", data)
	}

	// And WITH a ref, the key appears.
	entry.WorkItemRef = &WorkItemRef{SourceKind: "audit-orphan", OwnerPath: "/proj/x.go", Action: "draft SPEC stub"}
	data, err = json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal with ref: %v", err)
	}
	if !strings.Contains(string(data), "work_item_ref") {
		t.Errorf("DiffScopeEntry with WorkItemRef should include the key; got %s", data)
	}
}

// TestAppliedLedger_JSONSchema verifies the applied.json ledger schema
// (REQ-NS5-008c): approver + approval_timestamp + applied_subtree_ids +
// resulting_live_doc_sha. No wall-clock field.
func TestAppliedLedger_JSONSchema(t *testing.T) {
	ledger := AppliedLedger{
		Approver:            "orchestrator",
		ApprovalTimestamp:   "2026-08-12T10:00:00+00:00",
		AppliedSubtreeIDs:   []string{"pkg.ParseHeader"},
		ResultingLiveDocSHA: "deadbeef",
	}
	data, err := json.Marshal(ledger)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"approver", "approval_timestamp", "applied_subtree_ids", "resulting_live_doc_sha"} {
		if _, ok := m[key]; !ok {
			t.Errorf("AppliedLedger JSON missing key %q (REQ-NS5-008c schema)", key)
		}
	}
}
