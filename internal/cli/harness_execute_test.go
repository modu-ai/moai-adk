// Package cli — `moai harness apply --execute` 위임 + execute verb 등록 테스트.
// SPEC-HARNESS-APPLY-EXECUTE-001 REQ-AEX-002 / REQ-AEX-003 (AC-AEX-002, AC-AEX-003).
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeApplyProposalFixture는 root에 pending proposal JSON 1건을 작성한다.
func writeApplyProposalFixture(t *testing.T, root, id, targetPath, fieldKey, newValue string) {
	t.Helper()
	propDir := filepath.Join(root, ".moai", "harness", "proposals")
	if err := os.MkdirAll(propDir, 0o755); err != nil {
		t.Fatalf("mkdir proposals: %v", err)
	}
	prop := map[string]any{
		"id":          id,
		"target_path": targetPath,
		"field_key":   fieldKey,
		"new_value":   newValue,
		"pattern_key": "test-pattern",
	}
	data, err := json.Marshal(prop)
	if err != nil {
		t.Fatalf("marshal proposal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(propDir, id+".json"), data, 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
}

// TestExecuteCmd_RegisteredInRouter — [AC-AEX-002] `moai harness execute`가
// newHarnessRouterCmd()에 등록되어 있음을 검증한다.
func TestExecuteCmd_RegisteredInRouter(t *testing.T) {
	t.Parallel()

	router := newHarnessRouterCmd()
	var found bool
	for _, sub := range router.Commands() {
		if sub.Name() == "execute" {
			found = true
			break
		}
	}
	if !found {
		t.Error("`execute` verb must be registered in newHarnessRouterCmd() (REQ-AEX-002, AC-AEX-002)")
	}
}

// TestApply_PayloadOnly_PreservedWithoutExecuteFlag — [AC-AEX-003] --execute 부재
// 시 기존 payload-only 동작(JSON stdout 출력, Applier.Apply() 미호출)이 보존됨을
// 검증한다 (회귀 테스트, C4).
func TestApply_PayloadOnly_PreservedWithoutExecuteFlag(t *testing.T) {
	t.Parallel()

	dir := harnessTestProject(t)
	writeApplyProposalFixture(t, dir, "prop-payload-001", "docs/sample.md", "description", "x")

	var buf bytes.Buffer
	// --project-root is a persistent flag on the parent harness command; invoke
	// through the parent so the apply subcommand inherits it (mirrors TestCmdApply).
	cmd := newHarnessCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"apply", "--project-root", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("apply (payload-only) failed: %v", err)
	}
	out := buf.String()
	// payload-only 동작의 시그니처 출력: "next Proposal" 헤더 + proposal JSON.
	if !strings.Contains(out, "next Proposal") {
		t.Errorf("payload-only apply must print the proposal payload header; got:\n%s", out)
	}
	if !strings.Contains(out, "prop-payload-001") {
		t.Errorf("payload-only apply must print the proposal JSON; got:\n%s", out)
	}
	// Go 경로 위임 시그니처가 나오면 안 됨 (Applier.Apply 미호출 보장).
	if strings.Contains(out, "applied via Go pipeline") {
		t.Errorf("payload-only apply must NOT delegate to the Go pipeline; got:\n%s", out)
	}
}

// TestApply_DelegatesToGoPath_WithExecuteFlag — [AC-AEX-003] --execute --id 존재
// 시 Go execute 경로(harnesscli.RunExecute)로 위임됨을 검증한다.
//
// gate-active Applier(production)가 실제 go test를 재귀 실행하는 것을 회피하기 위해,
// proposal 파일을 부재시켜 RunExecute가 Apply 도달 전 loader 단계(exit 1)에서
// 반환하도록 한다 — 위임 자체(Go 경로 호출)는 stderr 진단 메시지로 실증된다.
func TestApply_DelegatesToGoPath_WithExecuteFlag(t *testing.T) {
	t.Parallel()

	dir := harnessTestProject(t)
	// proposal 부재 → RunExecute는 "proposal not found"로 반환 (Apply/measurer 미도달).
	var buf bytes.Buffer
	cmd := newHarnessCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"apply", "--project-root", dir, "--execute", "--id", "SPEC-NONEXISTENT-001"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("apply --execute with a missing proposal must return an error (delegation reached the Go path)")
	}
	out := buf.String()
	// 위임이 Go 경로(execute.go)에 도달했음을 진단 메시지로 실증한다.
	if !strings.Contains(out, "harness apply --execute") && !strings.Contains(err.Error(), "harness execute") {
		t.Errorf("apply --execute must delegate to the Go execute path (harness execute diagnostic); got out=%q err=%v", out, err)
	}
	// payload-only 시그니처는 나오면 안 됨 (Go 경로로 분기했으므로).
	if strings.Contains(out, "next Proposal") {
		t.Errorf("apply --execute must NOT run the payload-only path; got:\n%s", out)
	}
}
