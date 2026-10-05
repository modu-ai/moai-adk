package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

// readReceiptDir returns every receipt stored under a tree.
func readReceiptDir(t *testing.T, root string) []auditreceipt.Receipt {
	t.Helper()
	dir := filepath.Join(auditreceipt.StateDir(root), "receipts")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	out := make([]auditreceipt.Receipt, 0, len(entries))
	for _, e := range entries {
		r, err := auditreceipt.ReadReceipt(root, e.Name()[:len(e.Name())-len(".json")])
		if err != nil {
			t.Fatalf("ReadReceipt %s: %v", e.Name(), err)
		}
		out = append(out, r)
	}
	return out
}

// AC-CAG-008 (codex_audit half): a receipt is recorded for every call, and the
// id is handed back only in a tree that declared the gate required.
func TestCodexAudit_RecordsReceiptAndExposesItOnlyWhenRequired(t *testing.T) {
	for name, tc := range map[string]struct {
		gate      string
		wantField bool
	}{
		"required": {config.AuditGateRequired, true},
		"advisory": {config.AuditGateAdvisory, false},
	} {
		t.Run(name, func(t *testing.T) {
			root := newProbeProject(t, "SPEC-CAGRCPT-008")
			writeCodexAuditGate(t, root, tc.gate)
			withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

			got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
			receipts := readReceiptDir(t, root)
			if len(receipts) != 1 {
				t.Fatalf("receipts recorded = %d, want exactly 1 per call", len(receipts))
			}
			r := receipts[0]
			if r.Tool != auditreceipt.ToolCodexAudit || r.TreeRoot != root || r.RootSource != auditreceipt.RootSourceArgument {
				t.Errorf("receipt = %+v, want a codex_audit receipt for the named tree", r)
			}
			field, _ := got["audit_receipt"].(string)
			if tc.wantField && field != r.ReceiptID {
				t.Errorf("audit_receipt = %q, want the recorded id %q", field, r.ReceiptID)
			}
			if !tc.wantField && field != "" {
				t.Errorf("audit_receipt = %q, want no receipt field outside a required tree", field)
			}
		})
	}
}

// AC-CAG-008 (audit_multi half): the same contract on the convergence result,
// and only when codex actually participated.
func TestAuditMulti_RecordsReceiptWhenCodexParticipates(t *testing.T) {
	for name, tc := range map[string]struct {
		gate      string
		codexGate string
		wantField bool
		wantCount int
	}{
		"required":        {config.AuditGateRequired, "required", true, 1},
		"advisory":        {config.AuditGateAdvisory, "advisory", false, 1},
		"codex-gated-off": {config.AuditGateRequired, "off", false, 0},
	} {
		t.Run(name, func(t *testing.T) {
			root := newProbeProject(t, "SPEC-CAGRCPT-009")
			writeCodexAuditGate(t, root, tc.gate)
			withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

			result := runMultiAudit(context.Background(),
				ReviewOutput{Verdict: "pass", Summary: "anchor"},
				codexTargetUncommitted, "",
				MultiAuditConfig{
					ProjectRoot:    root,
					OriginProvider: BackendClaude,
					Gates:          config.AuditGates{Claude: "advisory", Codex: tc.codexGate, GLM: "off"},
				}, nil)

			receipts := readReceiptDir(t, root)
			if len(receipts) != tc.wantCount {
				t.Fatalf("receipts recorded = %d, want %d", len(receipts), tc.wantCount)
			}
			if tc.wantCount == 1 && receipts[0].Tool != auditreceipt.ToolAuditMulti {
				t.Errorf("receipt tool = %q, want audit_multi", receipts[0].Tool)
			}
			if tc.wantField && result.AuditReceipt != receipts[0].ReceiptID {
				t.Errorf("audit_receipt = %q, want the recorded id", result.AuditReceipt)
			}
			if !tc.wantField && result.AuditReceipt != "" {
				t.Errorf("audit_receipt = %q, want empty", result.AuditReceipt)
			}
		})
	}
}

// A call that names no tree records its receipt against the fallback root and
// marks it as such, so a worktree auditor can refuse it as another tree's.
func TestCodexAudit_FallbackRootedReceiptIsMarked(t *testing.T) {
	fallback := newProbeProject(t, "SPEC-CAGRCPT-010")
	writeCodexAuditGate(t, fallback, config.AuditGateRequired)
	prev := auditReceiptFallbackRoot
	auditReceiptFallbackRoot = func() string { return fallback }
	t.Cleanup(func() { auditReceiptFallbackRoot = prev })
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

	callToolCodexAudit(t, map[string]any{})

	receipts := readReceiptDir(t, fallback)
	if len(receipts) != 1 || receipts[0].RootSource != auditreceipt.RootSourceFallback {
		t.Fatalf("receipts = %+v, want exactly one fallback-rooted receipt", receipts)
	}
}
