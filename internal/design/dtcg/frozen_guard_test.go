package dtcg_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// TestIsFrozen_Violations: file paths under the FROZEN zone must always be blocked.
// REQ-DPL-011: design constitution §2, §3.1-3.3, §5, §11, §12 + brand context file protection.
func TestIsFrozen_Violations(t *testing.T) {
	t.Parallel()

	frozenPaths := []string{
		// The constitution file itself.
		".claude/rules/moai/design/constitution.md",
		// Brand context files (§3.1 Brand Context).
		".moai/project/brand/brand-voice.md",
		".moai/project/brand/visual-identity.md",
		".moai/project/brand/target-audience.md",
		// All files under the brand directory.
		".moai/project/brand/custom-file.md",
		// Paths corresponding to every FROZEN section in the constitution.
		".claude/rules/moai/design/constitution.md#2-frozen-vs-evolvable-zones",
		".claude/rules/moai/design/constitution.md#safety-architecture",
		".claude/rules/moai/design/constitution.md#gan-loop-contract",
		".claude/rules/moai/design/constitution.md#evaluator-leniency-prevention",
	}

	for _, path := range frozenPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if !dtcg.IsFrozen(path) {
				t.Errorf("IsFrozen(%q) = false; FROZEN 영역이어야 함", path)
			}
		})
	}
}

// TestIsFrozen_Allowed: file paths outside the FROZEN zone must be allowed.
func TestIsFrozen_Allowed(t *testing.T) {
	t.Parallel()

	allowedPaths := []string{
		// Implementation files (not FROZEN).
		"internal/design/dtcg/validator.go",
		"internal/design/dtcg/categories/color.go",
		".claude/skills/moai-workflow-design/SKILL.md",
		".moai/design/tokens.json",
		".moai/design/components.json",
		// EVOLVABLE sections (§6, §7, §8, §9, §10).
		".claude/rules/moai/design/CHANGELOG.md",
		// General rule files.
		".claude/rules/moai/core/moai-constitution.md",
		// User files.
		"CLAUDE.md",
		"internal/template/templates/CLAUDE.md",
	}

	for _, path := range allowedPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if dtcg.IsFrozen(path) {
				t.Errorf("IsFrozen(%q) = true; 허용돼야 함", path)
			}
		})
	}
}

// TestBlockWrite_FrozenPath: write attempts to FROZEN paths return an error.
func TestBlockWrite_FrozenPath(t *testing.T) {
	t.Parallel()

	frozenPaths := []string{
		".claude/rules/moai/design/constitution.md",
		".moai/project/brand/brand-voice.md",
		".moai/project/brand/visual-identity.md",
	}

	for _, path := range frozenPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			err := dtcg.BlockWrite(path, "test reason")
			if err == nil {
				t.Errorf("BlockWrite(%q) = nil; 오류 반환해야 함", path)
			}
			// Verify that the error message contains "frozen".
			if !strings.Contains(strings.ToLower(err.Error()), "frozen") {
				t.Errorf("BlockWrite(%q) 오류 메시지에 'frozen' 없음: %v", path, err)
			}
		})
	}
}

// TestBlockWrite_AllowedPath: BlockWrite returns nil for allowed paths.
func TestBlockWrite_AllowedPath(t *testing.T) {
	t.Parallel()

	allowedPaths := []string{
		"internal/design/dtcg/validator.go",
		".moai/design/tokens.json",
	}

	for _, path := range allowedPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			err := dtcg.BlockWrite(path, "test write")
			if err != nil {
				t.Errorf("BlockWrite(%q) = %v; nil 반환해야 함", path, err)
			}
		})
	}
}

// TestIsFrozen_ConfigBypass: config must not bypass the FROZEN zone.
// [HARD]: the FROZEN guard is hardcoded — config-bypass blocking test.
func TestIsFrozen_ConfigBypass(t *testing.T) {
	t.Parallel()

	// Bypass attempts on the FROZEN zone — cannot be altered by env vars or config.
	// constitution.md is always FROZEN.
	path := ".claude/rules/moai/design/constitution.md"
	if !dtcg.IsFrozen(path) {
		t.Errorf("IsFrozen(%q) = false; config bypass 불가 - 항상 FROZEN이어야 함", path)
	}

	// The brand directory is always FROZEN.
	brandPath := ".moai/project/brand/any-file.md"
	if !dtcg.IsFrozen(brandPath) {
		t.Errorf("IsFrozen(%q) = false; 브랜드 디렉토리는 항상 FROZEN이어야 함", brandPath)
	}
}
