// Package harness — NewInstallCmd RunE 클로저(install.go:125-155) 커버리지 테스트.
//
// SPEC-HARNESS-CLI-COVERAGE-001 M1: 기존 install_test.go의 TestRunInstall_* /
// TestNewInstallCmd_FlagsRegistered는 RunInstall 본문 + 팩토리 flag 배선만 커버한다.
// 본 파일은 NewInstallCmd().Execute()를 실제로 구동하여 RunE 클로저의 세 경로
// (1) --project-root 지정 success, (2) --project-root 생략 default-cwd(os.Getwd),
// (3) RunInstall 전파 error 를 도달시킨다.
//
// 격리(REQ-HCC-016): 모든 write는 t.TempDir() 내부. default-cwd 테스트는
// os.Chdir를 사용하므로 t.Parallel()을 호출하지 않고 t.Cleanup으로 cwd를 복원한다
// (REQ-HCC-018, AP-4).
//
// HARD subagent boundary (C-HRA-008): 이 파일은 AskUserQuestion을 호출하지 않는다.
// 패키지 가드 TestPropose_NoAskUserQuestion이 자동 스캔한다.
package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewInstallCmd_Execute_Success — RunE 클로저의 success 경로(install.go:125-155)를
// --project-root 지정으로 구동한다 (REQ-HCC-007). CLAUDE.md 픽스처가 존재해야
// InjectMarker가 성공한다. .moai/harness/main.md 스캐폴드 + CLAUDE.md marker 생성 +
// success stdout emit을 검증한다.
func TestNewInstallCmd_Execute_Success(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// InjectMarker는 CLAUDE.md를 먼저 읽으므로 픽스처가 필요하다.
	writeHarnessFile(t, filepath.Join(root, "CLAUDE.md"), "# Project\n")

	cmd := NewInstallCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"--spec-id", "SPEC-PROJ-INIT-001",
		"--domain", "ios-mobile",
		"--project-root", root,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v, want nil", err)
	}

	// REQ-HAW-005: main.md 스캐폴드.
	mainMDPath := filepath.Join(root, ".moai", "harness", "main.md")
	if _, err := os.Stat(mainMDPath); err != nil {
		t.Errorf("main.md not scaffolded: %v", err)
	}

	// CLAUDE.md marker block 주입.
	data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(data), "<!-- moai:harness-start") {
		t.Errorf("CLAUDE.md missing harness marker block after Execute")
	}

	// success stdout emit (152-154).
	if !strings.Contains(outBuf.String(), "harness activation wired") {
		t.Errorf("Execute did not emit success stdout; got: %q", outBuf.String())
	}
}

// TestNewInstallCmd_Execute_DefaultCwd — --project-root 생략 시 os.Getwd() 기본-cwd
// 분기(install.go:126-132)를 구동한다 (REQ-HCC-008). os.Chdir로 cwd를 t.TempDir()로
// 옮긴 뒤 실행하므로 non-parallel + t.Cleanup으로 원래 cwd 복원 (REQ-HCC-018, EC-1).
func TestNewInstallCmd_Execute_DefaultCwd(t *testing.T) {
	// 주의: os.Chdir는 프로세스 전역 상태이므로 t.Parallel() 금지 (AP-4).
	root := t.TempDir()
	writeHarnessFile(t, filepath.Join(root, "CLAUDE.md"), "# Project\n")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if cerr := os.Chdir(orig); cerr != nil {
			t.Errorf("restore cwd: %v", cerr)
		}
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to temp root: %v", err)
	}

	cmd := NewInstallCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	// --project-root 생략 → RunE 클로저가 os.Getwd() 분기로 cwd(=root)를 사용.
	cmd.SetArgs([]string{"--spec-id", "SPEC-CWD-001", "--domain", "web"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute (default cwd) error = %v, want nil", err)
	}

	// cwd(=root) 기준으로 main.md가 스캐폴드되어야 한다 (getwd 분기 도달 증거).
	if _, err := os.Stat(filepath.Join(root, ".moai", "harness", "main.md")); err != nil {
		t.Errorf("main.md not scaffolded under default cwd: %v", err)
	}
}

// TestNewInstallCmd_Execute_RunInstallError — CLAUDE.md 부재 root로 Execute 시
// RunInstall이 InjectMarker error를 반환하고, RunE 클로저의 error 전파 분기
// (install.go:149-150)가 도달됨을 검증한다. exit-error 반환 확인.
func TestNewInstallCmd_Execute_RunInstallError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// CLAUDE.md를 생성하지 않는다 → InjectMarker의 os.ReadFile 실패 → RunInstall error.

	cmd := NewInstallCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"--spec-id", "SPEC-ERR-001", "--domain", "d1", "--project-root", root})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute must return an error when CLAUDE.md is absent (RunInstall error propagation)")
	}
	if !strings.Contains(err.Error(), "CLAUDE.md") {
		t.Errorf("error should mention CLAUDE.md, got: %v", err)
	}
}
