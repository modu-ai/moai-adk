package cli

import "testing"

// TestVerifySubcommandsRegisteredOnRoot guards the init-order defect (card
// t1489): every verify verb must be a real subcommand of `verify` in the root
// command tree that main executes, regardless of which file defines it.
func TestVerifySubcommandsRegisteredOnRoot(t *testing.T) {
	for _, verb := range []string{"record", "check", "run", "sync-gate", "codex-review", "audit-plan"} {
		t.Run(verb, func(t *testing.T) {
			cmd, rest, err := rootCmd.Find([]string{"verify", verb})
			if err != nil || cmd == nil || cmd.Name() != verb || len(rest) != 0 {
				t.Fatalf("rootCmd.Find(verify %s) = %v, rest %v, err %v; want the %s command", verb, cmd, rest, err, verb)
			}
		})
	}
}
