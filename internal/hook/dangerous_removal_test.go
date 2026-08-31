package hook

import (
	"encoding/json"
	"testing"
)

// decideBash runs a raw Bash command string through the PreToolUse Bash guard
// and returns the decision it produced ("deny", "ask", or "" for allow).
func decideBash(t *testing.T, command string) string {
	t.Helper()
	h := &preToolHandler{policy: DefaultSecurityPolicy()}
	raw, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}
	decision, _ := h.checkBashCommand(json.RawMessage(raw))
	return decision
}

// TestDangerousRemoval_FlagOrderCannotBypass covers the "too narrow" direction of
// issue #1658: the guard matched the flag cluster literally, so reversing two
// characters walked straight past it. Every form below is shell-equivalent to
// the one form the old pattern caught.
func TestDangerousRemoval_FlagOrderCannotBypass(t *testing.T) {
	for _, command := range []string{
		"rm -rf /",
		"rm -fr /",
		"rm -r -f /",
		"rm -f -r /",
		"rm -r /",
		"rm -f /",
		"rm --recursive --force /",
		"rm --force --recursive /",
		"rm -Rf /",
		"rm -rf /usr",
		"rm -fr ~",
		"rm -fr $HOME",
		"rm -fr .git",
		"rm -fr node_modules",
		"rm -fr *",
		`rm -rf "/"`,
		"rm -rf '/'",
		"go build ./... && rm -fr /",
		"echo start; rm -fr /",
	} {
		if got := decideBash(t, command); got != DecisionDeny {
			t.Errorf("command %q: decision = %q, want %q", command, got, DecisionDeny)
		}
	}
}

// TestDangerousRemoval_QuotedDataIsNotACommand covers the "too wide" direction of
// issue #1658: the guard scanned the whole command text, so a command that only
// PRINTS or STORES the dangerous form was refused. That is what made it
// impossible to document, test, or report the guard's own defect.
func TestDangerousRemoval_QuotedDataIsNotACommand(t *testing.T) {
	for _, command := range []string{
		`echo "rm -rf /"`,
		`echo 'rm -rf /'`,
		`moai todo add "guard misses the removal form when flags are reversed"`,
		`git commit -m "fix(hook): block the removal form regardless of flag order"`,
		`printf '%s' "DROP DATABASE prod"`,
	} {
		if got := decideBash(t, command); got != "" {
			t.Errorf("command %q: decision = %q, want allow", command, got)
		}
	}
}

// TestDangerousRemoval_OrdinaryCleanupIsNotProtected covers the second half of the
// "too wide" direction: the old pattern ended at a bare slash, so EVERY absolute
// path was in range and routine scratch cleanup was refused.
func TestDangerousRemoval_OrdinaryCleanupIsNotProtected(t *testing.T) {
	for _, command := range []string{
		"rm -rf /tmp/moai-test-123",
		"rm -rf /var/folders/x9/abc/T/build",
		"rm -rf ./dist",
		"rm -rf build/cache",
		"rm -f coverage.out",
		"rm -rf ~/go/pkg/mod/cache/tmp",
	} {
		if got := decideBash(t, command); got != "" {
			t.Errorf("command %q: decision = %q, want allow", command, got)
		}
	}
}

// TestDangerousRemoval_StillBlocksAfterQuoteFolding is the safety-axis guard the
// relaxation owes: quote folding widens what is allowed, so this test states what
// the relaxed guard MUST still refuse. Quoting a protected target is never a path
// to executing its removal.
func TestDangerousRemoval_StillBlocksAfterQuoteFolding(t *testing.T) {
	for _, command := range []string{
		`rm -rf "$HOME"`,
		`rm -rf '~'`,
		`rm "-rf" /`,
		`echo "cleaning" && rm -rf /`,
		`echo 'rm -rf /tmp/x' && rm -fr ~`,
	} {
		if got := decideBash(t, command); got != DecisionDeny {
			t.Errorf("command %q: decision = %q, want %q", command, got, DecisionDeny)
		}
	}
}
