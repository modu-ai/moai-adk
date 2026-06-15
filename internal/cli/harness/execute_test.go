// Package harness — execute verb 블랙박스(T2) 테스트.
//
// design.md §F.1 [D4] 테스트 seam 결정에 따른 T2 (black-box, internal/cli/harness)
// 테스트. exported SafetyEvaluator 인터페이스(applier.go:194)에 stub evaluator를
// 주입하여 verb-shape + error→exit-code 매핑을 non-vacuous하게 검증한다.
// 실제 measurer는 절대 활성화하지 않는다 (stub evaluator가 결정을 좌우하거나,
// error-branch의 경우 Apply가 gate 도달 전 반환).
package harness

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// writeProposalFixture는 t.TempDir() 프로젝트 root에 pending proposal JSON 1건을
// 작성하고 proposal ID를 반환한다. targetPath는 비-FROZEN 경로여야 L1을 통과한다.
func writeProposalFixture(t *testing.T, root, id, targetPath, fieldKey, newValue string) {
	t.Helper()
	propDir := filepath.Join(root, ".moai", "harness", "proposals")
	if err := os.MkdirAll(propDir, 0o755); err != nil {
		t.Fatalf("mkdir proposals: %v", err)
	}
	prop := harness.Proposal{
		ID:         id,
		TargetPath: targetPath,
		FieldKey:   fieldKey,
		NewValue:   newValue,
		PatternKey: "test-pattern",
	}
	data, err := json.Marshal(prop)
	if err != nil {
		t.Fatalf("marshal proposal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(propDir, id+".json"), data, 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
}

// TestExecute_MissingProposal_ExitsUserError — proposal 파일 부재 시 user error(exit 1).
// AC-AEX-004 (proposal 로드) + AC-AEX-012 (부재 → exit 1).
func TestExecute_MissingProposal_ExitsUserError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// proposals 디렉터리 자체가 부재 (EC-3) — "no proposal" user error.
	err := RunExecute(ExecuteOptions{ID: "SPEC-DOES-NOT-EXIST-001", ProjectRoot: root})
	if err == nil {
		t.Fatal("RunExecute must return an error for a missing proposal")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("missing proposal exit code = %d, want 1 (user error); err=%v", got, err)
	}
}

// TestExecute_LoadsProposalByID — 정상 proposal ID가 .moai/harness/proposals/<id>.json
// 에서 harness.Proposal로 로드됨 (AC-AEX-004). loader만 검증.
func TestExecute_LoadsProposalByID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-LOAD-TEST-001"
	writeProposalFixture(t, root, id, "docs/sample.md", "description", "enriched note")

	prop, err := loadProposalByID(filepath.Join(root, ".moai", "harness", "proposals", id+".json"))
	if err != nil {
		t.Fatalf("loadProposalByID error: %v", err)
	}
	if prop.ID != id {
		t.Errorf("loaded proposal ID = %q, want %q", prop.ID, id)
	}
	if prop.TargetPath != "docs/sample.md" {
		t.Errorf("loaded TargetPath = %q, want docs/sample.md", prop.TargetPath)
	}
	if prop.FieldKey != "description" {
		t.Errorf("loaded FieldKey = %q, want description", prop.FieldKey)
	}
}

// TestExecute_PathTraversalProposalID_Rejected — proposal ID에 경로 traversal(../)
// 시도 시 base 디렉터리 밖 접근을 거부 (EC-1, 절대경로 규칙). exit 1.
func TestExecute_PathTraversalProposalID_Rejected(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := RunExecute(ExecuteOptions{ID: "../../../etc/passwd", ProjectRoot: root})
	if err == nil {
		t.Fatal("RunExecute must reject a path-traversal proposal ID")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("path-traversal exit code = %d, want 1 (user error); err=%v", got, err)
	}
}

// ── M3: error → exit code 매핑 (stub SafetyEvaluator 주입) ─────────────────

// stubEvaluator는 cross-package(T2)에서 주입 가능한 SafetyEvaluator stub이다.
// applier.go:194 SafetyEvaluator 인터페이스가 exported이므로 주입이 가능하다.
type stubEvaluator struct {
	decision harness.Decision
	err      error
}

func (s stubEvaluator) Evaluate(_ harness.Proposal, _ []harness.Session) (harness.Decision, error) {
	return s.decision, s.err
}

// newGateInactiveApplier는 gate-active가 아닌(measurer/baselineStore nil) Applier를
// 반환한다. error-branch 테스트는 Apply가 evaluator 단계(Step1)에서 분기를 반환하므로
// gate에 도달하지 않는다 (rejection/pending는 즉시 반환).
func newGateInactiveApplier() *harness.Applier {
	return harness.NewApplier()
}

// TestExecute_PendingErrorUnderAutoApply_ClassifiedAsInvariantExit2 — [D3 핵심]
// stub evaluator가 DecisionPendingApproval을 반환하여 *ApplyPendingError를 강제
// 발생시키고, RunExecute의 error→exit-code 매핑이 이를 INVARIANT VIOLATION(exit 2)으로
// 분류함을 실증한다 (AC-AEX-014). tautology "AutoApply=true는 Pending 안 냄"이 아님.
func TestExecute_PendingErrorUnderAutoApply_ClassifiedAsInvariantExit2(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-PENDING-001"
	writeProposalFixture(t, root, id, "docs/sample.md", "description", "x")

	eval := stubEvaluator{decision: harness.Decision{Kind: harness.DecisionPendingApproval}}
	err := runExecuteWith(ExecuteOptions{ID: id, ProjectRoot: root}, eval, newGateInactiveApplier())
	if err == nil {
		t.Fatal("runExecuteWith must return an error when evaluator returns PendingApproval")
	}
	// errjoin walk: *ApplyPendingError가 errors.As로 검출되어야 한다.
	var pendingErr *harness.ApplyPendingError
	if !errors.As(err, &pendingErr) {
		t.Errorf("error must classify as *ApplyPendingError via errors.As; got %T: %v", err, err)
	}
	if got := ExitCodeForError(err); got != 2 {
		t.Errorf("PendingApproval-under-AutoApply exit code = %d, want 2 (INVARIANT VIOLATION); err=%v", got, err)
	}
}

// TestExecute_RejectionError_ExitsUserError — L1~L4 rejection → exit 1 (AC-AEX-012).
func TestExecute_RejectionError_ExitsUserError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-REJECT-001"
	writeProposalFixture(t, root, id, "docs/sample.md", "description", "x")

	eval := stubEvaluator{decision: harness.Decision{Kind: harness.DecisionRejected, RejectedBy: 1}}
	err := runExecuteWith(ExecuteOptions{ID: id, ProjectRoot: root}, eval, newGateInactiveApplier())
	if err == nil {
		t.Fatal("runExecuteWith must return an error when evaluator rejects")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("rejection exit code = %d, want 1 (user error); err=%v", got, err)
	}
}

// TestExecute_RegressionError_ExitsUserError — *ApplyRegressionError → exit 1 (AC-AEX-013).
// SafetyEvaluator만으로는 regression을 발생시킬 수 없으므로, error-classification
// 분기만 단위로 검증한다 (errors.Join walk 포함).
func TestExecute_RegressionError_ExitsUserError(t *testing.T) {
	t.Parallel()

	regErr := &harness.ApplyRegressionError{Regressed: []string{"coverage"}}
	if got := ExitCodeForError(regErr); got != 1 {
		t.Errorf("ApplyRegressionError exit code = %d, want 1 (user-actionable)", got)
	}
	// errors.Join으로 감싸여도 walk 가능해야 한다 (errjoin 선례).
	wrapped := errors.Join(regErr, errors.New("outcome record failed"))
	if got := ExitCodeForError(wrapped); got != 1 {
		t.Errorf("joined ApplyRegressionError exit code = %d, want 1", got)
	}
}

// TestExecute_MeasurementExecError_ExitsSystemError — measurement-exec wrapped error
// → exit 2 (system error, AC-AEX-015). 알려진 error 타입이 아닌 system error는 exit 2.
func TestExecute_MeasurementExecError_ExitsSystemError(t *testing.T) {
	t.Parallel()

	measErr := errors.New("applier: regression gate baseline measurement failed (fail-closed): build error")
	if got := ExitCodeForError(measErr); got != 2 {
		t.Errorf("measurement-exec error exit code = %d, want 2 (system error)", got)
	}
}

// TestExecute_NilError_ExitZero — nil error → exit 0 (정상 경계).
func TestExecute_NilError_ExitZero(t *testing.T) {
	t.Parallel()

	if got := ExitCodeForError(nil); got != 0 {
		t.Errorf("nil error exit code = %d, want 0", got)
	}
}
