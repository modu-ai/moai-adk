package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// seedLiveSpec gives tree a non-closed SPEC directory, so a rejection naming
// id still counts as live for the spawn guard.
func seedLiveSpec(t *testing.T, tree, id string) {
	t.Helper()
	dir := filepath.Join(tree, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nid: " + id + "\nstatus: in-progress\n---\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAuditReceiptSpawn_StaleRejections(t *testing.T) {
	type rej struct {
		spec   string
		status string // "" = no spec dir; "-" = dir without spec.md status
	}
	tests := []struct {
		name      string
		rejs      []rej
		wantDeny  bool
		wantIn    []string
		wantNotIn []string
	}{
		{"spec dir absent", []rej{{"SPEC-GONE-001", ""}}, false, nil, nil},
		{"spec completed", []rej{{"SPEC-DONE-001", "completed"}}, false, nil, nil},
		{"spec archived", []rej{{"SPEC-ARCH-001", "archived"}}, false, nil, nil},
		{"spec in-progress", []rej{{"SPEC-LIVE-001", "in-progress"}}, true, []string{"SPEC-LIVE-001"}, nil},
		{"unknown spec", []rej{{auditreceipt.UnknownSpec, ""}}, true, []string{auditreceipt.UnknownSpec}, nil},
		{"live plus stale", []rej{{"SPEC-LIVE-001", "in-progress"}, {"SPEC-GONE-001", ""}, {"SPEC-DONE-001", "completed"}}, true,
			[]string{"SPEC-LIVE-001"}, []string{"SPEC-GONE-001", "SPEC-DONE-001"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := newGateTree(t, "required")
			for _, r := range tc.rejs {
				if r.status != "" {
					dir := filepath.Join(root, ".moai", "specs", r.spec)
					if err := os.MkdirAll(dir, 0o755); err != nil {
						t.Fatal(err)
					}
					body := "---\nid: " + r.spec + "\nstatus: " + r.status + "\n---\n# body\n"
					if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				rj := auditreceipt.Rejection{AgentType: auditreceipt.AgentPlanAuditor, SpecID: r.spec, Cause: auditreceipt.CauseNoReceiptCited}
				if err := auditreceipt.WriteRejection(root, &rj); err != nil {
					t.Fatalf("WriteRejection: %v", err)
				}
			}
			isDenied, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent")))
			if isDenied != tc.wantDeny {
				t.Fatalf("denied = %v (%s), want %v", isDenied, reason, tc.wantDeny)
			}
			for _, w := range tc.wantIn {
				if !strings.Contains(reason, w) {
					t.Errorf("reason %q lacks %q", reason, w)
				}
			}
			for _, w := range tc.wantNotIn {
				if strings.Contains(reason, w) {
					t.Errorf("reason %q names stale %q", reason, w)
				}
			}
		})
	}
}
