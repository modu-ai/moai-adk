// Package harness — execute verb 블랙박스(T2) 테스트.
//
// design.md §F.1 [D4] 테스트 seam 결정에 따른 T2 (black-box, internal/cli/harness)
// 테스트. exported SafetyEvaluator 인터페이스(applier.go:194)에 stub evaluator를
// 주입하여 verb-shape + error→exit-code 매핑을 non-vacuous하게 검증한다.
// 실제 measurer는 절대 활성화하지 않는다 (stub evaluator가 결정을 좌우하거나,
// error-branch의 경우 Apply가 gate 도달 전 반환).
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
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

// ── M2: Apply 파이프라인 배선 + autoApply contract + canonical paths ──────────

// TestExecute_ProductionPipeline_UsesAutoApplyTrue — [D3 분리, AC-AEX-007]
// production 배선이 safety.NewPipeline에 전달하는 PipelineConfig가 AutoApply: true임을
// 직접 관측한다 (non-vacuous). "AutoApply=true 하 Pending 안 냄" tautology가 아니라,
// RunExecute가 실제 production 구성 시 AutoApply: true를 전달하는지를 검증한다.
func TestExecute_ProductionPipeline_UsesAutoApplyTrue(t *testing.T) {
	t.Parallel()

	paths := resolveExecutePaths(t.TempDir())
	cfg := buildExecutePipelineConfig(paths)
	if !cfg.AutoApply {
		t.Error("production execute pipeline config must set AutoApply: true (autoApply contract, REQ-AEX-005)")
	}
	// ViolationLogPath / RateLimitPath도 canonical 경로로 채워져야 한다 (L1/L4 배선).
	if cfg.ViolationLogPath == "" {
		t.Error("PipelineConfig.ViolationLogPath must be wired (L1 frozen-guard violation log)")
	}
	if cfg.RateLimitPath == "" {
		t.Error("PipelineConfig.RateLimitPath must be wired (L4 rate limiter state)")
	}
}

// TestExecute_ResolvesCanonicalHarnessPaths — [AC-AEX-009] 4개 canonical 경로가
// project root 상대로 join되는지 table-driven으로 검증한다 (design.md §B Wiring Recipe).
func TestExecute_ResolvesCanonicalHarnessPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	paths := resolveExecutePaths(root)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"snapshotBase", paths.snapshotBase, filepath.Join(root, ".moai", "harness", "learning-history", "snapshots")},
		{"manifestPath", paths.manifestPath, filepath.Join(root, ".moai", "harness", "learning-history", "manifest.jsonl")},
		{"baselinePath", paths.baselinePath, filepath.Join(root, ".moai", "harness", "measurements-baseline.yaml")},
		{"usageLogPath", paths.usageLogPath, filepath.Join(root, ".moai", "harness", "usage-log.jsonl")},
		{"proposalDir", paths.proposalDir, filepath.Join(root, ".moai", "harness", "proposals")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestExecute_RelativeProjectRoot_AbsolutizedByAbs — 절대경로 규칙(AC-AEX-009):
// user-supplied 상대 --project-root는 filepath.Abs로 절대화되어야 한다. 빈 ID로
// 호출해 proposal 단계 전 root 처리 경로만 exercise한다 (userError exit 1).
func TestExecute_EmptyID_UserError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := RunExecute(ExecuteOptions{ID: "", ProjectRoot: root})
	if err == nil {
		t.Fatal("RunExecute must reject empty proposal ID")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("empty ID exit code = %d, want 1 (user error)", got)
	}
}

// TestNewExecuteCmd_FlagWiring — NewExecuteCmd 팩토리가 --id(required) + --project-root
// flag을 노출하고 Use="execute"임을 검증한다 (verb-shape, AC-AEX-001 보조).
func TestNewExecuteCmd_FlagWiring(t *testing.T) {
	t.Parallel()

	cmd := NewExecuteCmd()
	if cmd.Use != "execute" {
		t.Errorf("cmd.Use = %q, want \"execute\"", cmd.Use)
	}
	if cmd.Flags().Lookup("id") == nil {
		t.Error("execute cmd must expose --id flag")
	}
	if cmd.Flags().Lookup("project-root") == nil {
		t.Error("execute cmd must expose --project-root flag")
	}
	// --id는 required로 표시되어야 한다.
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil || idFlag.Annotations[cobraRequiredAnnotation] == nil {
		t.Error("execute cmd --id flag must be marked required")
	}
}

// cobraRequiredAnnotation은 cobra가 MarkFlagRequired 시 부여하는 annotation key다.
const cobraRequiredAnnotation = "cobra_annotation_bash_completion_one_required_flag"

// TestExecute_DiagnosticForError_Branches — diagnosticForError가 에러 타입별로
// 올바른 진단 메시지를 구성함을 검증한다 (design.md §E 메시지 열).
func TestExecute_DiagnosticForError_Branches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{"pending invariant", &harness.ApplyPendingError{}, "INVARIANT VIOLATION"},
		{"regression", &harness.ApplyRegressionError{Regressed: []string{"coverage"}}, "regression gate rolled back"},
		{"generic", errors.New("some system failure"), "some system failure"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diagnosticForError(tt.err)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("diagnosticForError = %q, want substring %q", got, tt.contains)
			}
		})
	}
}

// TestExecute_UserError_Message — userError.Error()가 메시지를 그대로 반환함을 검증.
func TestExecute_UserError_Message(t *testing.T) {
	t.Parallel()

	e := &userError{msg: "harness execute: proposal not found: X"}
	if e.Error() != "harness execute: proposal not found: X" {
		t.Errorf("userError.Error() = %q", e.Error())
	}
}

// TestExecute_MalformedProposalJSON_UserError — loadProposalByID가 깨진 JSON에
// 대해 parse-error user error(exit 1)를 반환함을 검증한다 (loadProposalByID 분기).
func TestExecute_MalformedProposalJSON_UserError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-MALFORMED-001"
	propDir := filepath.Join(root, ".moai", "harness", "proposals")
	if err := os.MkdirAll(propDir, 0o755); err != nil {
		t.Fatalf("mkdir proposals: %v", err)
	}
	if err := os.WriteFile(filepath.Join(propDir, id+".json"), []byte("{ not valid json"), 0o644); err != nil {
		t.Fatalf("write malformed proposal: %v", err)
	}

	err := RunExecute(ExecuteOptions{ID: id, ProjectRoot: root})
	if err == nil {
		t.Fatal("RunExecute must return an error for malformed proposal JSON")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("malformed proposal exit code = %d, want 1 (user error); err=%v", got, err)
	}
	if !strings.Contains(err.Error(), "parse proposal") {
		t.Errorf("error must indicate a parse failure; got: %v", err)
	}
}

// TestExecute_ProposalPathIsDirectory_UserError — proposal 경로가 디렉터리일 때
// loadProposalByID의 non-IsNotExist read-error 분기(user error, exit 1)를 검증한다.
func TestExecute_ProposalPathIsDirectory_UserError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-ISDIR-001"
	propDir := filepath.Join(root, ".moai", "harness", "proposals")
	// proposal 파일 경로 위치에 디렉터리를 만든다 → os.ReadFile은 IsNotExist가 아닌
	// "is a directory" 에러를 반환한다.
	if err := os.MkdirAll(filepath.Join(propDir, id+".json"), 0o755); err != nil {
		t.Fatalf("mkdir proposal-as-dir: %v", err)
	}

	err := RunExecute(ExecuteOptions{ID: id, ProjectRoot: root})
	if err == nil {
		t.Fatal("RunExecute must return an error when proposal path is a directory")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("proposal-is-dir exit code = %d, want 1 (user error); err=%v", got, err)
	}
}

// TestExecute_RunExecuteWith_EmptyRoot — runExecuteWith가 빈 ProjectRoot에서
// os.Getwd()로 fallback함을 검증한다 (runExecuteWith empty-root 분기). cwd에
// proposal이 없으므로 loader 단계에서 user error로 반환된다.
func TestExecute_RunExecuteWith_EmptyRoot(t *testing.T) {
	t.Parallel()

	eval := stubEvaluator{decision: harness.Decision{Kind: harness.DecisionApproved}}
	err := runExecuteWith(ExecuteOptions{ID: "SPEC-RXW-EMPTY-001"}, eval, newGateInactiveApplier())
	if err == nil {
		t.Fatal("runExecuteWith with empty ProjectRoot + missing proposal must return an error")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("empty-root runExecuteWith exit code = %d, want 1 (user error)", got)
	}
}

// TestExecute_EmptyProjectRoot_UsesCwd — ProjectRoot 미지정 시 현재 디렉터리로
// fallback함을 검증한다 (RunExecute empty-root 분기). cwd에 proposal이 없으므로
// loader 단계에서 user error(exit 1)로 반환되어 cwd fallback 경로를 exercise한다.
func TestExecute_EmptyProjectRoot_UsesCwd(t *testing.T) {
	t.Parallel()

	// 현재 cwd(테스트 패키지 디렉터리)에는 .moai/harness/proposals/<id>.json이 없다.
	err := RunExecute(ExecuteOptions{ID: "SPEC-CWD-FALLBACK-001"})
	if err == nil {
		t.Fatal("RunExecute with empty ProjectRoot + missing proposal must return an error")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("empty-root missing-proposal exit code = %d, want 1", got)
	}
}

// TestRunExecuteCommand_ErrorEmitsDiagnostic — runExecuteCommand가 실패 시 stderr로
// 진단 메시지를 emit하고 에러를 반환함을 검증한다 (RunE 본문 분해 함수, os.Exit 미호출).
func TestRunExecuteCommand_ErrorEmitsDiagnostic(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	var outBuf, errBuf bytes.Buffer
	cmd := NewExecuteCmd()
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	err := runExecuteCommand(cmd, "SPEC-NONE-001", root)
	if err == nil {
		t.Fatal("runExecuteCommand must return an error for a missing proposal")
	}
	if !strings.Contains(errBuf.String(), "harness execute") {
		t.Errorf("runExecuteCommand must emit a stderr diagnostic; got: %q", errBuf.String())
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("missing proposal via runExecuteCommand exit code = %d, want 1", got)
	}
}

// TestRunExecuteCommand_EmptyProjectRoot_InheritsFromParent — runExecuteCommand가
// 빈 projectRoot에서 부모 persistent flag(--project-root)를 상속함을 검증한다.
// 부모 flag을 t.TempDir()로 세팅하고, 해당 root에 proposal이 없으므로 user error로
// 반환되어 상속 분기(projectRoot == "" → InheritedFlags lookup)를 exercise한다.
func TestRunExecuteCommand_EmptyProjectRoot_InheritsFromParent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// 부모 command에 persistent --project-root flag을 세팅한다 (router 구조 모사).
	parent := &cobra.Command{Use: "harness"}
	parent.PersistentFlags().String("project-root", root, "")
	child := NewExecuteCmd()
	parent.AddCommand(child)

	var errBuf bytes.Buffer
	child.SetErr(&errBuf)
	child.SetOut(&bytes.Buffer{})

	// projectRoot=""로 호출 → InheritedFlags()에서 부모 값(root) 상속.
	err := runExecuteCommand(child, "SPEC-INHERIT-001", "")
	if err == nil {
		t.Fatal("runExecuteCommand must return error (missing proposal in inherited root)")
	}
	if got := ExitCodeForError(err); got != 1 {
		t.Errorf("inherited-root missing-proposal exit code = %d, want 1", got)
	}
}

// TestExecute_NilSessions_CanaryDoesNotReject — [AC-AEX-011] nil/empty sessions로
// gate-active Applier.Apply를 호출할 때 L2 Canary가 reject하지 않음을 검증한다.
// production Pipeline(AutoApply: true)을 직접 구성하여 L1~L5를 실제로 통과시키고,
// gate-active Applier는 same-package stubMeasurer로 구성할 수 없으므로(cross-package)
// 여기서는 evaluator를 production Pipeline으로 두되 gate-inactive Applier를 주입해
// L2 통과 후 straight-line apply가 fieldKey 오류 없이 진행됨을 확인한다.
//
// 핵심: nil sessions가 L2 reject를 유발하지 않음을 production Pipeline.Evaluate로
// 실증한다 (canary.go nil-safe). meaningful proposal(TargetPath+NewValue 존재)을 사용.
func TestExecute_NilSessions_CanaryDoesNotReject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-CANARY-001"
	// 비-FROZEN 타겟 + description fieldKey. gate-inactive Applier(straight-line)는
	// applyFileModification에서 실제 파일을 EnrichDescription하므로 fixture 파일이 필요.
	targetRel := "docs/canary-sample.md"
	targetAbs := filepath.Join(root, targetRel)
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(targetAbs, []byte("---\ndescription: original\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("write target fixture: %v", err)
	}
	writeProposalFixture(t, root, id, targetAbs, "description", "canary enriched note")

	paths := resolveExecutePaths(root)
	pipeline := safety.NewPipeline(buildExecutePipelineConfig(paths))
	// gate-inactive Applier: L1~L5 통과(autoApply=true) 후 straight-line modify+lineage.
	applier := harness.NewApplier()

	err := runExecuteWith(ExecuteOptions{ID: id, ProjectRoot: root}, pipeline, applier)
	if err != nil {
		// L2 Canary reject라면 "rejected (L2" 메시지가 나온다 — 이는 실패.
		t.Fatalf("nil sessions must not trigger L2 canary rejection; got error: %v", err)
	}
}
