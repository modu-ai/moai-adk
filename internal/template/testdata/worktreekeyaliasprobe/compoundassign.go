package worktreekeyaliasprobe

import "github.com/modu-ai/moai-adk/internal/config"

// CompoundAppend consumes the session-name pattern via += — a read under
// REQ-002 (AC-005c): compound assignments consume the old value.
func CompoundAppend(cfg *config.Config) string {
	w := cfg.Workflow.Worktree
	w.SessionNamePattern += "-suffix"
	return w.SessionNamePattern
}
