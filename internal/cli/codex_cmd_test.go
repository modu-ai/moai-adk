package cli

import "testing"

// card t1909 step ⑤ (t1910 §3.3) — after the general-execution removal the
// codexCmd parent keeps its registration and its audit/role-audit subcommands
// but carries NO RunE: the general run path is REMOVED, not deactivated. The
// audit user surface (`moai codex audit` / `moai codex role-audit`) must stay
// byte-for-byte reachable.
func TestCodexCmd_AuditOnlySurface(t *testing.T) {
	if codexCmd.RunE != nil {
		t.Error("codexCmd.RunE is non-nil — the general run path must be removed, not kept (card t1909 removal acceptance)")
	}
	if codexCmd.Parent() != rootCmd {
		t.Error("codexCmd.Parent() != rootCmd — parent registration lost in the removal")
	}
	found := map[string]bool{}
	for _, c := range codexCmd.Commands() {
		found[c.Name()] = true
	}
	for _, want := range []string{"audit", "role-audit"} {
		if !found[want] {
			t.Errorf("codexCmd subcommand %q missing — the audit user surface must stay unchanged (t1910 §3.3)", want)
		}
	}
}
