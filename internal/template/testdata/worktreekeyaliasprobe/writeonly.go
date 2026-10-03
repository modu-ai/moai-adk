package worktreekeyaliasprobe

import "github.com/modu-ai/moai-adk/internal/config"

// PlainWriteOnly touches the auto-cleanup flag only as a plain `=`
// assignment — the sole non-read shape (AC-005b; REQ-002's exclusion).
// This file must NOT be classified as a reader.
func PlainWriteOnly(cfg *config.Config) {
	w := cfg.Workflow.Worktree
	w.AutoCleanup = false
}
