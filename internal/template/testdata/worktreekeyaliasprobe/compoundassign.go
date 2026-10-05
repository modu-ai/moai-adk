package worktreekeyaliasprobe

import "github.com/modu-ai/moai-adk/internal/config"

// CompoundAppend consumes the session-name pattern via += — a read under
// REQ-002 (AC-005c): compound assignments consume the old value. The += is
// this file's ONLY tracked-field access (acceptance §D.5's only-as shape).
func CompoundAppend(cfg *config.Config) {
	w := cfg.Workflow.Worktree
	w.SessionNamePattern += "-suffix"
}
