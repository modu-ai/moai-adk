package cli

// t1619_glm_coverage_test.go — card t1619, bundle B2: the unexecuted changed lines of glm.go
// (saveLLMSection's yamlpatch refusal and its config-directory failure). Family: TestSaveLLM.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestSaveLLMSectionMultilineValueIsRefusedByYAMLPatch — a value that spans lines has no
// yamlpatch form, so the save reports the refusal and writes nothing.
func TestSaveLLMSectionMultilineValueIsRefusedByYAMLPatch(t *testing.T) {
	sections := filepath.Join(t.TempDir(), "sections")
	llm := config.NewDefaultLLMConfig()
	llm.TeamMode = "glm\nsecond line"

	err := saveLLMSection(sections, llm)
	if err == nil || !strings.Contains(err.Error(), "cannot be written through the yamlpatch seam") {
		t.Fatalf("a multi-line value must be refused by the yamlpatch seam, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(sections, "llm.yaml")); statErr == nil {
		t.Fatal("a refused save must not create llm.yaml")
	}
}

// TestSaveLLMSectionUncreatableDirectoryIsReturned — a sections path that cannot be created
// (a dangling symlink sits there) reports the failure instead of writing.
func TestSaveLLMSectionUncreatableDirectoryIsReturned(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the dangling-symlink fixture is a unix filesystem shape")
	}
	base := t.TempDir()
	sections := filepath.Join(base, "sections")
	if err := os.Symlink(filepath.Join(base, "no-such-target"), sections); err != nil {
		t.Fatal(err)
	}

	err := saveLLMSection(sections, config.NewDefaultLLMConfig())
	if err == nil || !strings.Contains(err.Error(), "create config directory") {
		t.Fatalf("an uncreatable sections directory must be reported, got %v", err)
	}
}
