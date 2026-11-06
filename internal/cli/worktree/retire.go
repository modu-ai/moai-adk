package worktree

// retire.go — SPEC-WORKTREE-HOME-ROOT-001 M7: the retirement of the empty
// MoAI legacy directory (REQ-WHR-020). After the migration has moved the
// last moai-made tree out of <primary>/.moai/worktrees, the empty directory
// is removed by a non-recursive removal only. The Claude-owned directory
// (<primary>/.claude/worktrees) is never removed, renamed or pruned by any
// code path.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RetireEmptyLegacyDir removes the empty MoAI legacy directory
// (<root>/.moai/worktrees) by non-recursive removal, returning a notice
// when the directory is not empty (it still holds an unmigrated tree or
// any file). The Claude-owned directory (<root>/.claude/worktrees) is
// never touched by this function.
//
// @MX:NOTE: [AUTO] retirement — the M7 non-recursive removal of the empty MoAI legacy dir
// @MX:SPEC: SPEC-WORKTREE-HOME-ROOT-001
func RetireEmptyLegacyDir(root string) error {
	legacyDir := filepath.Join(root, ".moai", "worktrees")
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already retired
		}
		return fmt.Errorf("retire: read %s: %w", legacyDir, err)
	}
	if len(entries) > 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return fmt.Errorf("retire: %s is not empty (contains %s); the directory is left in place", legacyDir, strings.Join(names, ", "))
	}
	// Non-recursive removal: the directory is empty, so os.Remove succeeds
	// without a recursive flag. os.RemoveAll is deliberately not used — the
	// retirement removes only the empty directory, never its content.
	return os.Remove(legacyDir)
}
