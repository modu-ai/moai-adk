// Package worktreekeyaliasprobe is the SPEC-TPL-AST-GUARD-001 alias
// characterization fixture (AC-005). It is loaded by explicit pattern from
// TestWorkflowWorktreeKeyHonestyAliasFixture and stays invisible to ./...
// builds because it lives under testdata/.
//
// This file reads the auto-cleanup flag through a local alias copy — the
// shape the retained t682 text scan cannot see (AC-005a). The full accessor
// chain is deliberately never spelled out in this file's source, which is
// itself the AC-005d characterization: the legacy text accessor string has
// zero matches here.
package worktreekeyaliasprobe

import "github.com/modu-ai/moai-adk/internal/config"

// AliasRead copies the worktree config into a local alias and reads the
// auto-cleanup flag through the copy.
func AliasRead(cfg *config.Config) bool {
	w := cfg.Workflow.Worktree
	return w.AutoCleanup
}
