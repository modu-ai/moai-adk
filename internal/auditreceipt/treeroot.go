package auditreceipt

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// treeRootTimeout bounds the git call: a hook runs under a 5-second budget, so
// a hanging git must not consume it.
const treeRootTimeout = 2 * time.Second

// TreeRootFromCWD resolves the canonical tree root a hook is acting in, from
// the hook input's own cwd: the git toplevel when git can answer, otherwise the
// cwd itself, canonicalized in both cases.
//
// CLAUDE_PROJECT_DIR is deliberately NOT consulted. In a worktree session it
// names the primary checkout, so a guard keyed on it would compare a worktree
// auditor's receipts against another tree's store and reject everything.
func TreeRootFromCWD(cwd string) string {
	dir := strings.TrimSpace(cwd)
	if dir == "" {
		return ""
	}
	if top := gitToplevel(dir); top != "" {
		dir = top
	}
	return Canonical(dir)
}

// Canonical resolves symlinks, falling back to the cleaned path when the
// resolution fails — a path that cannot be canonicalized still names a tree,
// and the comparison it feeds is byte equality against another value produced
// the same way.
func Canonical(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return filepath.Clean(dir)
}

func gitToplevel(dir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), treeRootTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
