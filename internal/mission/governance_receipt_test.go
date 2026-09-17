package mission

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadGovernanceReceiptsBindsDecisionAndIndependentAudit(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 9, 15, 1, 2, 3, 0, time.UTC)
	expect := GovernanceExpectation{
		MissionID: "018f4f4a-7b7c-7a11-8f4d-111111111111", ContractHash: strings.Repeat("a", 64),
		SnapshotHash: strings.Repeat("b", 64), Action: ActionDispatch, Targets: []string{"gtd:gtd-1111111111111111"},
		HeadSHA: strings.Repeat("c", 40), Now: now,
	}
	dir := filepath.Join(root, ".moai", "state", "mission", "governance")
	governorPath := filepath.Join(dir, "decision.json")
	auditPath := filepath.Join(dir, "audit.json")
	decision := GovernanceReceipt{Version: 1, Kind: GovernanceDecision, MissionID: expect.MissionID, ContractHash: expect.ContractHash, SnapshotHash: expect.SnapshotHash, Action: expect.Action, Targets: expect.Targets, ExpiresAt: now.Add(time.Minute), Issuer: "mission-governor", HeadSHA: expect.HeadSHA, Status: GovernanceRecommended}
	audit := GovernanceReceipt{Version: 1, Kind: GovernanceAudit, MissionID: expect.MissionID, ContractHash: expect.ContractHash, SnapshotHash: expect.SnapshotHash, Action: expect.Action, Targets: expect.Targets, ExpiresAt: now.Add(time.Minute), Issuer: "sync-auditor", HeadSHA: expect.HeadSHA, Status: GovernancePassed}
	if err := WriteGovernanceReceipt(root, governorPath, decision); err != nil {
		t.Fatal(err)
	}
	if err := WriteGovernanceReceipt(root, auditPath, audit); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, governorPath, auditPath, expect); err != nil {
		t.Fatalf("valid governance receipts rejected: %v", err)
	}

	mutate := func(path string, fn func(*GovernanceReceipt)) {
		t.Helper()
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		r, err := ParseGovernanceReceipt(raw)
		if err != nil {
			t.Fatal(err)
		}
		fn(&r)
		// Deliberately retain the old integrity field: this is the tamper mutant.
		encoded, err := marshalGovernanceReceipt(r)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, encoded, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mutate(governorPath, func(r *GovernanceReceipt) { r.Targets = []string{"repo:outside"} })
	if _, err := LoadGovernanceReceipts(root, governorPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), "receipt_integrity") {
		t.Fatalf("tampered receipt err=%v", err)
	}
}

func TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	expect := GovernanceExpectation{MissionID: "018f4f4a-7b7c-7a11-8f4d-222222222222", ContractHash: strings.Repeat("a", 64), SnapshotHash: strings.Repeat("b", 64), Action: ActionPick, Targets: []string{"gtd:gtd-1111111111111111"}, HeadSHA: strings.Repeat("c", 40), Now: now}
	dir := filepath.Join(root, ".moai", "state", "mission", "governance")
	decisionPath, auditPath := filepath.Join(dir, "decision.json"), filepath.Join(dir, "audit.json")
	decision := GovernanceReceipt{Version: 1, Kind: GovernanceDecision, MissionID: expect.MissionID, ContractHash: expect.ContractHash, SnapshotHash: expect.SnapshotHash, Action: expect.Action, Targets: expect.Targets, ExpiresAt: now.Add(time.Minute), Issuer: "mission-governor", HeadSHA: expect.HeadSHA, Status: GovernanceRecommended}
	audit := GovernanceReceipt{Version: 1, Kind: GovernanceAudit, MissionID: expect.MissionID, ContractHash: expect.ContractHash, SnapshotHash: expect.SnapshotHash, Action: expect.Action, Targets: expect.Targets, ExpiresAt: now.Add(time.Minute), Issuer: "sync-auditor", HeadSHA: expect.HeadSHA, Status: GovernancePassed}
	if err := WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, decisionPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing audit err=%v", err)
	}
	if err := WriteGovernanceReceipt(root, auditPath, audit); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(decisionPath, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, decisionPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("weak mode err=%v", err)
	}
	if err := os.Chmod(decisionPath, 0o600); err != nil {
		t.Fatal(err)
	}
	outsideRoot := t.TempDir()
	outside := filepath.Join(outsideRoot, ".moai", "state", "mission", "governance", "decision.json")
	if err := WriteGovernanceReceipt(outsideRoot, outside, decision); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, outside, auditPath, expect); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("outside err=%v", err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(decisionPath, link); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, link, auditPath, expect); err == nil || !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("symlink err=%v", err)
	}
	audit.Status = GovernanceFailed
	if err := WriteGovernanceReceipt(root, auditPath, audit); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, decisionPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), "audit_failed") {
		t.Fatalf("failed audit err=%v", err)
	}
	audit.Status, audit.ExpiresAt = GovernancePassed, now.Add(-time.Second)
	if err := WriteGovernanceReceipt(root, auditPath, audit); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadGovernanceReceipts(root, decisionPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale err=%v", err)
	}
}

func TestLoadGovernanceReceiptsRejectsEverySignedLineageAndStatusMutant(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	expect := GovernanceExpectation{MissionID: "mission", ContractHash: "contract", SnapshotHash: "snapshot", Action: ActionDispatch, Targets: []string{"gtd:b", "gtd:a"}, HeadSHA: "head", Now: now}
	dir := filepath.Join(root, ".moai", "state", "mission", "governance")
	decisionPath, auditPath := filepath.Join(dir, "decision.json"), filepath.Join(dir, "audit.json")
	base := GovernanceReceipt{Version: 1, MissionID: expect.MissionID, ContractHash: expect.ContractHash, SnapshotHash: expect.SnapshotHash, Action: expect.Action, Targets: []string{"gtd:a", "gtd:b"}, ExpiresAt: now.Add(time.Minute), HeadSHA: expect.HeadSHA}
	validDecision := base
	validDecision.Kind, validDecision.Issuer, validDecision.Status = GovernanceDecision, "mission-governor", GovernanceRecommended
	validAudit := base
	validAudit.Kind, validAudit.Issuer, validAudit.Status = GovernanceAudit, "sync-auditor", GovernancePassed

	tests := []struct {
		name   string
		mutate func(*GovernanceReceipt)
		want   string
	}{
		{name: "version", mutate: func(r *GovernanceReceipt) { r.Version = 2 }, want: "receipt_status"},
		{name: "kind", mutate: func(r *GovernanceReceipt) { r.Kind = GovernanceAudit }, want: "receipt_status"},
		{name: "issuer", mutate: func(r *GovernanceReceipt) { r.Issuer = "super-advisor" }, want: "receipt_status"},
		{name: "status", mutate: func(r *GovernanceReceipt) { r.Status = GovernancePassed }, want: "receipt_status"},
		{name: "mission", mutate: func(r *GovernanceReceipt) { r.MissionID = "other" }, want: "receipt_lineage"},
		{name: "contract", mutate: func(r *GovernanceReceipt) { r.ContractHash = "other" }, want: "receipt_lineage"},
		{name: "snapshot", mutate: func(r *GovernanceReceipt) { r.SnapshotHash = "other" }, want: "receipt_lineage"},
		{name: "action", mutate: func(r *GovernanceReceipt) { r.Action = ActionPick }, want: "receipt_lineage"},
		{name: "targets length", mutate: func(r *GovernanceReceipt) { r.Targets = []string{"gtd:a"} }, want: "receipt_lineage"},
		{name: "targets value", mutate: func(r *GovernanceReceipt) { r.Targets = []string{"gtd:a", "gtd:c"} }, want: "receipt_lineage"},
		{name: "head", mutate: func(r *GovernanceReceipt) { r.HeadSHA = "other" }, want: "receipt_lineage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := validDecision
			tt.mutate(&decision)
			if err := WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
				t.Fatal(err)
			}
			if err := WriteGovernanceReceipt(root, auditPath, validAudit); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadGovernanceReceipts(root, decisionPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("mutant accepted or wrong error: %v", err)
			}
		})
	}

	if _, err := ParseGovernanceReceipt([]byte("{")); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if err := WriteGovernanceReceipt(root, "relative.json", validDecision); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("relative receipt err=%v", err)
	}
	unsafePath := filepath.Join(dir, "unsafe.json")
	if err := os.Symlink(decisionPath, unsafePath); err != nil {
		t.Fatal(err)
	}
	if err := WriteGovernanceReceipt(root, unsafePath, validDecision); err == nil || !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("write symlink err=%v", err)
	}
	invalidTime := validDecision
	invalidTime.ExpiresAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := WriteGovernanceReceipt(root, filepath.Join(dir, "invalid-time.json"), invalidTime); err == nil {
		t.Fatal("unencodable receipt accepted")
	}
	directoryPath := filepath.Join(dir, "directory.json")
	if err := os.Mkdir(directoryPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := WriteGovernanceReceipt(root, directoryPath, validDecision); err == nil {
		t.Fatal("directory destination accepted")
	}
	lockedDir := filepath.Join(dir, "locked")
	if err := os.Mkdir(lockedDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(lockedDir, 0o700) })
	if err := WriteGovernanceReceipt(root, filepath.Join(lockedDir, "receipt.json"), validDecision); err == nil {
		t.Fatal("unwritable receipt directory accepted")
	}
	if _, err := LoadGovernanceReceipts(filepath.Join(root, "missing-root"), decisionPath, auditPath, expect); err == nil || !strings.Contains(err.Error(), "invalid_root") {
		t.Fatalf("missing root err=%v", err)
	}
}

func TestGovernanceDirectoryRejectsMissingRelativeAndEscapedRoots(t *testing.T) {
	missing := t.TempDir()
	if _, err := governanceDir(missing, false); err == nil {
		t.Fatal("missing governance directory accepted")
	}
	if _, err := governanceDir(".", false); err == nil || !strings.Contains(err.Error(), "invalid_root") {
		t.Fatalf("relative root err=%v", err)
	}
	escaped := t.TempDir()
	outside := t.TempDir()
	parent := filepath.Join(escaped, ".moai", "state", "mission")
	if err := os.MkdirAll(parent, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(parent, "governance")); err != nil {
		t.Fatal(err)
	}
	if _, err := governanceDir(escaped, false); err == nil || !strings.Contains(err.Error(), "receipt_outside") {
		t.Fatalf("escaped governance directory err=%v", err)
	}
	blocked := t.TempDir()
	if err := os.WriteFile(filepath.Join(blocked, ".moai"), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governanceDir(blocked, true); err == nil {
		t.Fatal("mkdir failure ignored")
	}
}
