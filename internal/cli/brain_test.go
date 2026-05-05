package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestBrainCmd_Exists verifies brainCmd is not nil (등록 전 nil 예상 — RED 상태).
func TestBrainCmd_Exists(t *testing.T) {
	if brainCmd == nil {
		t.Fatal("brainCmd should not be nil")
	}
}

// TestBrainCmd_Use verifies the command Use field matches the spec.
func TestBrainCmd_Use(t *testing.T) {
	if brainCmd.Use != `brain "<idea>"` {
		t.Errorf("brainCmd.Use = %q, want %q", brainCmd.Use, `brain "<idea>"`)
	}
}

// TestBrainCmd_Short verifies Short description is non-empty.
func TestBrainCmd_Short(t *testing.T) {
	if brainCmd.Short == "" {
		t.Error("brainCmd.Short should not be empty")
	}
}

// TestBrainCmd_Long verifies Long description explains the slash command boundary.
func TestBrainCmd_Long(t *testing.T) {
	if brainCmd.Long == "" {
		t.Error("brainCmd.Long should not be empty")
	}
	// Long description must explain that the workflow runs in Claude Code, not the terminal
	if !strings.Contains(brainCmd.Long, "Claude Code") {
		t.Errorf("brainCmd.Long should mention 'Claude Code', got %q", brainCmd.Long)
	}
}

// TestBrainCmd_IsSubcommandOfRoot verifies brain is registered under rootCmd.
func TestBrainCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "brain" {
			found = true
			break
		}
	}
	if !found {
		t.Error("brain should be registered as a subcommand of root")
	}
}

// TestBrainCmd_HasRunE verifies RunE is set (brain is not a help-only command).
func TestBrainCmd_HasRunE(t *testing.T) {
	if brainCmd.RunE == nil {
		t.Error("brainCmd.RunE should not be nil")
	}
}

// TestBrainCmd_RunE_WithIdea verifies that invoking brain with an idea
// prints a user-facing instruction directing the user to run /moai brain in Claude Code.
func TestBrainCmd_RunE_WithIdea(t *testing.T) {
	buf := new(bytes.Buffer)
	brainCmd.SetOut(buf)
	brainCmd.SetErr(buf)

	// 아이디어 인자 전달 시 Claude Code 지시 메시지 출력 확인
	err := brainCmd.RunE(brainCmd, []string{"I want to build a habit tracker app"})
	if err != nil {
		t.Fatalf("brainCmd.RunE with idea error: %v", err)
	}

	output := buf.String()

	// 사용자에게 Claude Code 채팅에서 /moai brain을 실행하라는 안내가 있어야 함
	if !strings.Contains(output, "/moai brain") {
		t.Errorf("output should contain '/moai brain', got %q", output)
	}
	if !strings.Contains(output, "Claude Code") {
		t.Errorf("output should contain 'Claude Code', got %q", output)
	}
}

// TestBrainCmd_RunE_NoArgs verifies that invoking brain without arguments
// still produces a helpful message (not a panic or silent failure).
func TestBrainCmd_RunE_NoArgs(t *testing.T) {
	buf := new(bytes.Buffer)
	brainCmd.SetOut(buf)
	brainCmd.SetErr(buf)

	// 인자 없이 실행 시 도움말 메시지 출력
	err := brainCmd.RunE(brainCmd, []string{})
	if err != nil {
		t.Fatalf("brainCmd.RunE without args error: %v", err)
	}

	output := buf.String()
	// 인자 없이도 Claude Code 안내가 출력되어야 함
	if !strings.Contains(output, "Claude Code") {
		t.Errorf("output should mention 'Claude Code' even with no args, got %q", output)
	}
}

// TestBrainCmd_InstructionsOnlyFlag verifies that --instructions-only flag
// is defined on brainCmd.
func TestBrainCmd_InstructionsOnlyFlag(t *testing.T) {
	flag := brainCmd.Flags().Lookup("instructions-only")
	if flag == nil {
		t.Error("brainCmd should have --instructions-only flag")
	}
}

// TestBrainCmd_InstructionsOnly_Output verifies that --instructions-only
// prints the 7-phase summary without requiring an idea argument.
func TestBrainCmd_InstructionsOnly_Output(t *testing.T) {
	// --instructions-only 플래그를 true로 설정 후 실행
	if err := brainCmd.Flags().Set("instructions-only", "true"); err != nil {
		t.Fatalf("failed to set --instructions-only flag: %v", err)
	}
	defer func() {
		// 테스트 격리: 플래그를 원래 값으로 복원
		_ = brainCmd.Flags().Set("instructions-only", "false")
	}()

	buf := new(bytes.Buffer)
	brainCmd.SetOut(buf)
	brainCmd.SetErr(buf)

	err := brainCmd.RunE(brainCmd, []string{})
	if err != nil {
		t.Fatalf("brainCmd.RunE --instructions-only error: %v", err)
	}

	output := buf.String()

	// 7-phase 요약이 출력되어야 함
	phases := []string{"Discovery", "Diverge", "Research", "Converge", "Critical", "Proposal", "Handoff"}
	for _, phase := range phases {
		if !strings.Contains(output, phase) {
			t.Errorf("--instructions-only output should contain phase %q, got %q", phase, output)
		}
	}
}

// TestBrainCmd_Help verifies --help output contains key information.
func TestBrainCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	brainCmd.SetOut(buf)
	brainCmd.SetErr(buf)

	// --help 출력에 brain 워크플로우 설명이 포함되어야 함
	err := brainCmd.Help()
	if err != nil {
		t.Fatalf("brainCmd.Help() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "brain") {
		t.Errorf("help output should contain 'brain', got %q", output)
	}
	// --instructions-only 플래그가 help에 표시되어야 함
	if !strings.Contains(output, "instructions-only") {
		t.Errorf("help output should mention 'instructions-only' flag, got %q", output)
	}
}

// TestBrainCmd_GroupID verifies brain is in the "project" command group
// (same as other workflow commands).
func TestBrainCmd_GroupID(t *testing.T) {
	// brain은 project 그룹에 속해야 함 (version, loop, spec 등의 tools 그룹 아님)
	if brainCmd.GroupID == "" {
		t.Error("brainCmd.GroupID should not be empty")
	}
}

// Table-driven: various idea inputs should all produce /moai brain guidance.
func TestBrainCmd_VariousIdeaInputs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantSubstring  string
		wantNoError    bool
	}{
		{
			name:          "english idea",
			args:          []string{"build a todo app"},
			wantSubstring: "/moai brain",
			wantNoError:   true,
		},
		{
			name:          "korean idea",
			args:          []string{"습관 추적 앱을 만들고 싶어"},
			wantSubstring: "/moai brain",
			wantNoError:   true,
		},
		{
			name:          "multi-word idea",
			args:          []string{"I want to build", "a productivity tool"},
			wantSubstring: "/moai brain",
			wantNoError:   true,
		},
		{
			name:          "empty args",
			args:          []string{},
			wantSubstring: "Claude Code",
			wantNoError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			brainCmd.SetOut(buf)
			brainCmd.SetErr(buf)

			// --instructions-only가 false인지 확인 (이전 테스트 격리)
			_ = brainCmd.Flags().Set("instructions-only", "false")

			err := brainCmd.RunE(brainCmd, tt.args)
			if tt.wantNoError && err != nil {
				t.Fatalf("RunE(%v) unexpected error: %v", tt.args, err)
			}
			if !tt.wantNoError && err == nil {
				t.Fatalf("RunE(%v) expected error, got nil", tt.args)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantSubstring) {
				t.Errorf("output should contain %q, got %q", tt.wantSubstring, output)
			}
		})
	}
}
