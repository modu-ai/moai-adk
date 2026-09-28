package codexwiring

// seed.go — SPEC-HANDOFF-NEUTRAL-001 M1.3 (REQ-HN-005/006/007).
//
// Worktrees do not carry the project layer's .codex/hooks.json: it is an
// untracked runtime artifact, so `git worktree` never copies it and a Codex
// session inside a worktree runs without any moai hook. SeedHooksIfMissing is
// the idempotent seeding entry the worktree materializer and the launcher
// entry path call: it creates the file with the MoAI-owned table ONLY when no
// hooks.json exists, and never touches an existing file's bytes (user-owned
// entries preserved). Generation is RenderHooks — the same renderer Wire uses,
// with a nil existing document — so there is no second hooks.json builder.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
)

// SeedHooksIfMissing writes .codex/hooks.json into projectRoot when (and only
// when) it does not exist yet. The rendered content is the MoAI-owned hook
// table (RenderHooks(nil)) gated by the same whitelist as Wire (REQ-CW-003):
// refusing bytes never reach disk.
//
// Idempotency contract: an existing file — user-owned, moai-owned, or mixed —
// is left byte-identical. Callers that need refresh semantics use Wire.
//
// Returns whether the file was seeded. An error is a caller-visible failure;
// the caller decides the failure posture (every current caller is fail-open:
// stderr diagnostic + proceed, REQ-HN-007).
func SeedHooksIfMissing(projectRoot string) (bool, error) {
	hooksPath := filepath.Join(projectRoot, filepath.FromSlash(HooksRelPath))
	if _, err := os.Stat(hooksPath); err == nil {
		return false, nil // existence gate — existing bytes are never rewritten
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("stat %s: %w", HooksRelPath, err)
	}

	rendered, err := RenderHooks(nil)
	if err != nil {
		return false, fmt.Errorf("render %s: %w", HooksRelPath, err)
	}
	violations, verr := codexadapter.ValidateConfig(rendered)
	if verr != nil {
		return false, fmt.Errorf("%w: %v", ErrValidationRefused, verr)
	}
	if len(violations) > 0 {
		names := make([]string, len(violations))
		for i, v := range violations {
			names[i] = v.Error()
		}
		return false, fmt.Errorf("%w: %s", ErrValidationRefused, strings.Join(names, "; "))
	}
	if err := writeAtomic(hooksPath, rendered); err != nil {
		return false, fmt.Errorf("write %s: %w", HooksRelPath, err)
	}
	return true, nil
}
