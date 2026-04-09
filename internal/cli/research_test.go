package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- TDD RED: research 커맨드 구조 테스트 ---

func TestNewResearchCmd_Use(t *testing.T) {
	cmd := newResearchCmd()
	if cmd.Use != "research" {
		t.Errorf("Use = %q, want %q", cmd.Use, "research")
	}
}

func TestNewResearchCmd_HasThreeSubcommands(t *testing.T) {
	cmd := newResearchCmd()
	subs := cmd.Commands()
	if len(subs) != 3 {
		t.Errorf("subcommand count = %d, want 3", len(subs))
	}

	// 서브커맨드 이름 확인
	names := make(map[string]bool)
	for _, s := range subs {
		names[s.Use] = true
	}
	for _, want := range []string{"status", "baseline [target]", "list"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestNewResearchCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "research" {
			found = true
			break
		}
	}
	if !found {
		t.Error("research should be registered as a subcommand of root")
	}
}

// --- TDD RED: research status 테스트 ---

func TestRunResearchStatus_NoDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	buf := new(bytes.Buffer)

	err := runResearchStatus(buf, tmpDir)
	if err != nil {
		t.Fatalf("runResearchStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No research data found") {
		t.Errorf("output should contain 'No research data found', got %q", output)
	}
}

func TestRunResearchStatus_WithDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	// .moai/research/ 디렉터리 생성
	researchDir := filepath.Join(tmpDir, ".moai", "research")
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err := runResearchStatus(buf, tmpDir)
	if err != nil {
		t.Fatalf("runResearchStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Research Status") {
		t.Errorf("output should contain 'Research Status', got %q", output)
	}
}

// --- TDD RED: research baseline 테스트 ---

func TestRunResearchBaseline_ComingSoon(t *testing.T) {
	cmd := newResearchBaselineCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("RunE error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "not yet implemented") {
		t.Errorf("output should mention 'not yet implemented', got %q", output)
	}
}

// --- TDD RED: research list 테스트 ---

func TestRunResearchList_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	buf := new(bytes.Buffer)

	err := runResearchList(buf, tmpDir)
	if err != nil {
		t.Fatalf("runResearchList error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No eval suites found") {
		t.Errorf("output should contain 'No eval suites found', got %q", output)
	}
}

func TestRunResearchList_WithEvalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	evalsDir := filepath.Join(tmpDir, ".moai", "research", "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// eval 파일 생성
	files := []string{
		"hook-perf.eval.yaml",
		"template-quality.eval.yaml",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(evalsDir, f), []byte("name: test\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	buf := new(bytes.Buffer)
	err := runResearchList(buf, tmpDir)
	if err != nil {
		t.Fatalf("runResearchList error: %v", err)
	}

	output := buf.String()
	for _, f := range files {
		if !strings.Contains(output, f) {
			t.Errorf("output should contain %q, got %q", f, output)
		}
	}
}

func TestRunResearchList_WithNestedEvalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, ".moai", "research", "evals", "hooks")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(nestedDir, "latency.eval.yaml"), []byte("name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err := runResearchList(buf, tmpDir)
	if err != nil {
		t.Fatalf("runResearchList error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "latency.eval.yaml") {
		t.Errorf("output should contain nested eval file, got %q", output)
	}
}

// --- TDD RED: research status 서브커맨드 필드 테스트 ---

func TestNewResearchStatusCmd_Use(t *testing.T) {
	cmd := newResearchStatusCmd()
	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}
}

func TestNewResearchBaselineCmd_Use(t *testing.T) {
	cmd := newResearchBaselineCmd()
	if cmd.Use != "baseline [target]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "baseline [target]")
	}
}

func TestNewResearchListCmd_Use(t *testing.T) {
	cmd := newResearchListCmd()
	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}
}
