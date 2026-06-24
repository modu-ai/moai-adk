// Package harness — runExecuteCommand success 경로 + NewExecuteCmd RunE success 경로
// 커버리지 테스트.
//
// SPEC-HARNESS-CLI-COVERAGE-001 M3: 기존 execute_test.go는 runExecuteCommand의
// error 경로(stderr diagnostic emit + missing-proposal inherit)만 커버한다. 본 파일은
//   (1) runExecuteCommand success stdout emit 블록(execute.go:366-369),
//   (2) runExecuteCommand InheritedFlags success 분기(execute.go:358-360),
//   (3) NewExecuteCmd().Execute() RunE 클로저 success 경로(execute.go:327-328 entry +
//       335 return nil) — D3 debt: runExecuteCommand 직접 호출은 RunE 클로저를 통과하지
//       않으므로, cmd.Execute()를 valid fixture로 구동해야 RunE 진입 + return-nil이 도달된다.
//
// gate-active Apply는 corrected project root에서 `go test ./...`를 실제 실행하므로
// (SPEC-HARNESS-EXECUTE-E2E-001 measurementRoot threading), success 경로 fixture는
// real 최소 Go 모듈(writeE2EFixtureModule) + valid proposal(writeProposalFixture)이어야
// 한다 (EC-4).
//
// 격리(REQ-HCC-016): 모든 write는 t.TempDir() 내부. HARD subagent boundary(C-HRA-008):
// AskUserQuestion 미호출 — 패키지 가드 TestPropose_NoAskUserQuestion이 스캔한다.
package harness

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunExecuteCommand_Success_EmitsTelemetryNotice — runExecuteCommand가 success
// 경로에서 stdout telemetry 알림 블록(execute.go:366-369)을 emit함을 검증한다
// (REQ-HCC-009). real Go 모듈 fixture + valid proposal로 gate-active Apply를 성공시킨다.
func TestRunExecuteCommand_Success_EmitsTelemetryNotice(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-M3-SUCCESS-001"
	targetAbs := writeE2EFixtureModule(t, root)
	writeProposalFixture(t, root, id, targetAbs, "description", "m3 enriched note")

	cmd := NewExecuteCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	// runExecuteCommand 직접 호출: RunE 클로저를 우회하여 내부 success-emit 블록만 exercise.
	err := runExecuteCommand(cmd, id, root)
	if err != nil {
		t.Fatalf("runExecuteCommand must succeed against the e2e fixture; got: %v\nstderr: %s",
			err, errBuf.String())
	}
	// 366-369: success stdout emit ("apply-outcome telemetry recorded").
	if !strings.Contains(outBuf.String(), "apply-outcome telemetry recorded") {
		t.Errorf("runExecuteCommand did not emit the telemetry-recorded notice; got stdout: %q",
			outBuf.String())
	}
	if !strings.Contains(outBuf.String(), id) {
		t.Errorf("telemetry notice should reference the proposal ID %q; got: %q", id, outBuf.String())
	}
}

// TestRunExecuteCommand_InheritsProjectRootFromParent_Success — runExecuteCommand가
// 빈 projectRoot에서 부모 persistent --project-root flag를 상속하고(execute.go:358-360
// `if f != nil` body), 상속된 root가 valid e2e fixture이므로 success 경로로 진행함을
// 검증한다 (REQ-HCC-010).
//
// 구조적 주의: NewExecuteCmd는 own --project-root local flag를 가지므로, NewExecuteCmd를
// 자식으로 둔 트리에서는 cobra가 local flag로 부모 persistent flag를 shadow하여
// InheritedFlags().Lookup("project-root")이 nil을 반환한다 (기존
// TestRunExecuteCommand_EmptyProjectRoot_InheritsFromParent는 그래서 cwd fallback
// missing-proposal error만 커버하고 359 body는 미도달). local shadow가 없는 자식
// (bare cobra.Command)로 트리를 구성해야 InheritedFlags lookup이 non-nil이 되어
// 359 body(projectRoot = f.Value.String())가 도달된다. runExecuteCommand는 임의의
// *cobra.Command를 받으므로 이 caller shape는 합법적이다.
func TestRunExecuteCommand_InheritsProjectRootFromParent_Success(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-M3-INHERIT-001"
	targetAbs := writeE2EFixtureModule(t, root)
	writeProposalFixture(t, root, id, targetAbs, "description", "m3 inherit note")

	// 부모 persistent --project-root flag를 e2e fixture root로 세팅.
	parent := &cobra.Command{Use: "harness"}
	parent.PersistentFlags().String("project-root", root, "")
	// local --project-root shadow가 없는 자식 → InheritedFlags lookup이 부모 flag를 반환.
	child := &cobra.Command{Use: "execute-inherit-probe"}
	parent.AddCommand(child)

	var outBuf, errBuf bytes.Buffer
	child.SetOut(&outBuf)
	child.SetErr(&errBuf)

	// projectRoot=""로 호출 → InheritedFlags()에서 부모 값(root) 상속(359 body) 후 success.
	err := runExecuteCommand(child, id, "")
	if err != nil {
		t.Fatalf("runExecuteCommand must succeed after inheriting --project-root; got: %v\nstderr: %s",
			err, errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "apply-outcome telemetry recorded") {
		t.Errorf("inherited-root success must emit telemetry notice; got stdout: %q", outBuf.String())
	}
}

// TestNewExecuteCmd_Execute_Success — [D3 핵심] NewExecuteCmd().Execute()를 valid
// fixture로 구동하여 RunE 클로저의 success 경로(execute.go:327 entry + 335 return nil)를
// 도달시킨다. runExecuteCommand가 nil error를 반환하면 RunE는 os.Exit를 건너뛰고
// return nil로 정상 종료한다 (success 경로에서는 프로세스 exit이 발생하지 않는다).
// 이것이 cmd.Execute() 직접 호출이 안전한 이유다 (실패 경로만 os.Exit, EX-2 residual).
func TestNewExecuteCmd_Execute_Success(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-M3-RUNE-001"
	targetAbs := writeE2EFixtureModule(t, root)
	writeProposalFixture(t, root, id, targetAbs, "description", "m3 rune note")

	cmd := NewExecuteCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--id", id, "--project-root", root})

	// cmd.Execute()가 RunE 클로저를 구동한다 — runExecuteCommand 성공 → os.Exit 미발생
	// → RunE return nil(335). Execute()는 nil을 반환한다.
	if err := cmd.Execute(); err != nil {
		t.Fatalf("NewExecuteCmd().Execute() must succeed against the e2e fixture; got: %v\nstderr: %s",
			err, errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "apply-outcome telemetry recorded") {
		t.Errorf("Execute() success must emit telemetry notice via RunE; got stdout: %q",
			outBuf.String())
	}

	// usage-log telemetry가 실제로 기록되었는지 보조 확인(success 경로 진정성).
	usageLogPath := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
	if got := countApplyOutcomeKept(t, usageLogPath); got != 1 {
		t.Errorf("apply_outcome(kept) line count = %d, want 1 (RunE success telemetry)", got)
	}
}
