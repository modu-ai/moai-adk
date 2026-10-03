package template

// apply_deploy_mode_test.go — SPEC-INIT-SHRINK-001 REQ-009 (OD-5 settled
// (a)): the deployment_mode writer beside ApplyHarness, same llm.yaml
// section file, same line-patch write shape (byte-identity of everything
// else).

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

const deployModeFixtureLLM = "llm:\n  harness: claude\n  conversation_language: ko\n"

func writeDeployModeFixture(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(sectionPath(root, ""), 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(sectionPath(root, "llm.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
	return root
}

func readLLMYaml(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(sectionPath(root, "llm.yaml"))
	if err != nil {
		t.Fatalf("read back llm.yaml: %v", err)
	}
	return string(data)
}

// TestApplyDeployModeWritesKeyAndPreservesSiblings: the value lands on its
// own deployment_mode line and every other byte of the file survives.
func TestApplyDeployModeWritesKeyAndPreservesSiblings(t *testing.T) {
	root := writeDeployModeFixture(t, deployModeFixtureLLM)
	if err := ApplyDeployMode(root, "plugin"); err != nil {
		t.Fatalf("ApplyDeployMode: %v", err)
	}
	got := readLLMYaml(t, root)
	if !strings.Contains(got, "deployment_mode: plugin") {
		t.Fatalf("deployment_mode key not written:\n%s", got)
	}
	if !strings.Contains(got, "harness: claude") {
		t.Errorf("harness line damaged:\n%s", got)
	}
	if !strings.Contains(got, "conversation_language: ko") {
		t.Errorf("sibling line dropped:\n%s", got)
	}
	if got := config.ReadDeployMode(root); got != "plugin" {
		t.Errorf("round-trip read = %q, want plugin", got)
	}
}

// TestApplyDeployModeIdempotentAndStable: a second call with the same value
// leaves the file byte-identical, and switching values replaces the line.
func TestApplyDeployModeIdempotentAndStable(t *testing.T) {
	root := writeDeployModeFixture(t, deployModeFixtureLLM)
	if err := ApplyDeployMode(root, "local"); err != nil {
		t.Fatalf("first ApplyDeployMode: %v", err)
	}
	once := readLLMYaml(t, root)
	if err := ApplyDeployMode(root, "local"); err != nil {
		t.Fatalf("second ApplyDeployMode: %v", err)
	}
	if twice := readLLMYaml(t, root); twice != once {
		t.Fatalf("idempotent call changed the file:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if err := ApplyDeployMode(root, "plugin"); err != nil {
		t.Fatalf("switch ApplyDeployMode: %v", err)
	}
	if got := readLLMYaml(t, root); !strings.Contains(got, "deployment_mode: plugin") {
		t.Fatalf("value switch not applied:\n%s", got)
	}
}

// TestApplyDeployModeInvalidValueIsErrorNotWrite: an out-of-set value is an
// error and writes nothing (the same contract ApplyHarness carries).
func TestApplyDeployModeInvalidValueIsErrorNotWrite(t *testing.T) {
	root := writeDeployModeFixture(t, deployModeFixtureLLM)
	if err := ApplyDeployMode(root, "banana"); err == nil {
		t.Fatal("ApplyDeployMode(banana) = nil, want error")
	}
	if got := readLLMYaml(t, root); strings.Contains(got, "banana") || strings.Contains(got, "deployment_mode") {
		t.Fatalf("rejected value touched the file:\n%s", got)
	}
}

// TestApplyDeployModeAbsentFileIsGracefulNoOp mirrors ApplyHarness: an
// absent llm.yaml is a nil no-op (the caller deploys the section file first).
func TestApplyDeployModeAbsentFileIsGracefulNoOp(t *testing.T) {
	root := t.TempDir()
	if err := ApplyDeployMode(root, "plugin"); err != nil {
		t.Fatalf("absent file: err = %v, want nil", err)
	}
}

// TestApplyDeployModeInsertsUnderLLMRoot: a legacy llm.yaml without the key
// gains one under the llm: root.
func TestApplyDeployModeInsertsUnderLLMRoot(t *testing.T) {
	root := writeDeployModeFixture(t, "llm:\n  harness: gpt\n")
	if err := ApplyDeployMode(root, "local"); err != nil {
		t.Fatalf("ApplyDeployMode: %v", err)
	}
	got := readLLMYaml(t, root)
	if !strings.Contains(got, "deployment_mode: local") {
		t.Fatalf("key not inserted:\n%s", got)
	}
	// Inserted under the llm: root at two-space indent — the same
	// insert-under-root position the ApplyHarness pattern uses.
	if !strings.Contains(got, "llm:\n  deployment_mode: local") && !strings.Contains(got, "llm:\n  harness: gpt\n  deployment_mode: local") {
		t.Fatalf("key inserted at the wrong place:\n%s", got)
	}
}
