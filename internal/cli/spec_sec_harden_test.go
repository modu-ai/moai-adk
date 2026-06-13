// SPEC-SEC-HARDEN-002 M3 — spec view/status/close read-path traversal guard (A-F4).
//
// 재현 우선: 픽스 전 코드는 traversal SPEC-ID(`../../../../etc`)를 거부하지 않고
// .moai/specs/ 밖의 read-path를 구성한다(또는 spec.Close로 전달). 픽스 후
// validateSpecID 가드가 CLI args[0] 경계에서 거부한다. spec_view/spec_status/
// spec_close 세 핸들러에 동일 가드를 적용한다(spec_drift 제외 — positional
// SPEC-ID 없음). closer.go는 수정하지 않는다.
package cli

import (
	"strings"
	"testing"
)

// TestSpecView_TraversalRejected 은 AC-SEC2-M3-001 (RED→GREEN) 이다.
// `moai spec view '../../../../etc'` 는 read-path 구성 전에 거부되어야 한다.
func TestSpecView_TraversalRejected(t *testing.T) {
	tempDir := t.TempDir()
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tempDir, nil }
	defer func() { findProjectRootFn = origFn }()

	cmd := newSpecViewCmd()
	outBuf := &strings.Builder{}
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)

	err := viewAcceptanceCriteria(cmd, "../../../../etc", false)
	if err == nil {
		t.Fatalf("expected rejection of traversal SPEC-ID, got nil")
	}
	if !strings.Contains(err.Error(), "SPEC-ID") {
		t.Errorf("expected SPEC-ID validation error, got: %v", err)
	}
}

// TestSpecStatus_TraversalRejected 은 AC-SEC2-M3-002 (spec status 핸들러) 이다.
func TestSpecStatus_TraversalRejected(t *testing.T) {
	tempDir := t.TempDir()
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tempDir, nil }
	defer func() { findProjectRootFn = origFn }()

	cmd := newSpecStatusCmd()
	outBuf := &strings.Builder{}
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)

	err := updateSpecStatus(cmd, "../../../../etc", "completed", true /* dryRun */)
	if err == nil {
		t.Fatalf("expected rejection of traversal SPEC-ID, got nil")
	}
	if !strings.Contains(err.Error(), "SPEC-ID") {
		t.Errorf("expected SPEC-ID validation error, got: %v", err)
	}
}

// TestSpecClose_TraversalRejected 은 AC-SEC2-M3-002 (spec close 핸들러) 이다.
// 가드는 CLI args[0] 경계(spec_close.go)에서 spec.Close 호출 전에 발동한다.
// 내부 transitive sink(closer.go)는 수정하지 않는다.
func TestSpecClose_TraversalRejected(t *testing.T) {
	cmd := newSpecCloseCmd()
	cmd.SetArgs([]string{"../../../../etc"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	outBuf := &strings.Builder{}
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected rejection of traversal SPEC-ID, got nil")
	}
	if !strings.Contains(err.Error(), "SPEC-ID") {
		t.Errorf("expected SPEC-ID validation error, got: %v", err)
	}
}

// TestSpecSubcommands_LegitimateSpecIDNotRejected 은 AC-SEC2-M3-003 (NO-REG) 이다.
// 정상 SPEC-ID 는 가드를 통과해 기존 동작(파일 미존재 시 not-found 에러 등)으로
// 진행해야 한다 — SPEC-ID 검증 에러가 아니어야 한다.
func TestSpecSubcommands_LegitimateSpecIDNotRejected(t *testing.T) {
	tempDir := t.TempDir()
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tempDir, nil }
	defer func() { findProjectRootFn = origFn }()

	cmd := newSpecViewCmd()
	outBuf := &strings.Builder{}
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)

	// 정상 SPEC-ID: spec.md 가 없으므로 not-found 에러는 나지만,
	// validateSpecID 가 거부한 "SPEC-ID must not..." 에러는 아니어야 한다.
	err := viewAcceptanceCriteria(cmd, "SPEC-SEC-HARDEN-002", false)
	if err != nil && strings.Contains(err.Error(), "SPEC-ID must") {
		t.Errorf("legitimate SPEC-ID must not be rejected by validateSpecID, got: %v", err)
	}
}
