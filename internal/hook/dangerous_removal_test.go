package hook

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/security"
	"github.com/modu-ai/moai-adk/internal/template"
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

// --- Deployed-policy regression surface (SPEC-REMOVAL-GUARD-EXTRAS-001) ---
//
// The four TestDangerousRemoval_* functions above compose the handler with
// DefaultSecurityPolicy only, so they never exercise the extras path that
// internal/cli/deps.go wires into every live session. That gap is what let a
// superseded extras regex survive the structural-check repair. Every test below
// therefore judges through the DEPLOYED policy: defaults plus the extras merged
// from the security.yaml template exactly as it ships in the binary.

// structuralDenyPrefix is the reason prefix produced by the structural
// target check in checkBashCommand. The pattern-scan path formats its reason as
// the pattern text itself, so asserting this prefix pins WHICH component judged
// the command — the reason a deny-direction test fails the moment the structural
// check is bypassed or a pattern reclaims the decision.
const structuralDenyPrefix = "Dangerous command blocked: removal of protected path"

// deployedPolicy composes the policy the way internal/cli/deps.go does for a
// live session: DefaultSecurityPolicy plus MergeExtraPatterns over the extra
// patterns loaded from the deployed security.yaml. The template is read from
// the embedded filesystem — the exact bytes compiled into the binary — and
// seeded into a temp project directory so LoadExtraSecurityConfig loads it
// through the same path production uses.
func deployedPolicy(t *testing.T) *SecurityPolicy {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("open embedded templates: %v", err)
	}
	data, err := fs.ReadFile(embedded, filepath.Join(".moai", "config", "sections", "security.yaml"))
	if err != nil {
		t.Fatalf("read embedded security.yaml: %v", err)
	}
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("seed config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "security.yaml"), data, 0o644); err != nil {
		t.Fatalf("seed security.yaml: %v", err)
	}
	policy := DefaultSecurityPolicy()
	policy.MergeExtraPatterns(security.LoadExtraSecurityConfig(dir))
	return policy
}

// decideBashDeployedReason runs a raw Bash command through the PreToolUse Bash
// guard composed with the deployed policy and returns both the decision and the
// reason, so a test can tell the structural check from the pattern scan.
func decideBashDeployedReason(t *testing.T, command string) (string, string) {
	t.Helper()
	h := &preToolHandler{policy: deployedPolicy(t)}
	raw, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}
	return h.checkBashCommand(json.RawMessage(raw))
}

// decideBashDeployed is decideBashDeployedReason with the reason discarded,
// mirroring decideBash for the deployed policy.
func decideBashDeployed(t *testing.T, command string) string {
	t.Helper()
	decision, _ := decideBashDeployedReason(t, command)
	return decision
}

// TestDangerousRemovalDeployed_AllowsDeepPathHeredocDataMention covers the
// heredoc direction of the deployed-policy defect: a heredoc body that only
// MENTIONS a deep-path removal as example data was refused by the superseded
// extras pattern even though the structural check classifies every line
// correctly (cat and echo are not removals). Reproduces the live R1
// observation; the body carries the real-space example form only.
func TestDangerousRemovalDeployed_AllowsDeepPathHeredocDataMention(t *testing.T) {
	for _, command := range []string{
		"cat > /tmp/t511-deployed-allow.md <<'EOF'\nexample form: rm -rf /tmp/moai-spec-repro-123\nEOF\necho WROTE_OK",
	} {
		if got := decideBashDeployed(t, command); got != "" {
			t.Errorf("command %q: decision = %q, want allow", command, got)
		}
	}
}

// TestDangerousRemovalDeployed_AllowsUnquotedDataMention covers the unquoted
// direction: a bare echo whose argument merely names a deep-path removal form
// has no quoted span to fold, so the superseded extras pattern matched the text
// as-is. Reproduces the live R2 observation.
func TestDangerousRemovalDeployed_AllowsUnquotedDataMention(t *testing.T) {
	for _, command := range []string{
		"echo example: rm -rf /tmp/moai-spec-repro-456",
	} {
		if got := decideBashDeployed(t, command); got != "" {
			t.Errorf("command %q: decision = %q, want allow", command, got)
		}
	}
}

// TestDangerousRemovalDeployed_AllowsScratchCleanupExecution covers the
// execution direction: an actual removal aimed at an unprotected deep scratch
// path must follow the structural check's verdict, not be refused by a pattern.
// Reproduces the live R3 observation and re-runs an ordinary-cleanup case from
// the builtin-only suite under the deployed policy.
func TestDangerousRemovalDeployed_AllowsScratchCleanupExecution(t *testing.T) {
	for _, command := range []string{
		"rm -rf /tmp/t511-scratch-nonexistent-dir",
		"rm -rf /var/folders/x9/abc/T/build",
		"rm -rf ~/go/pkg/mod/cache/tmp",
	} {
		if got := decideBashDeployed(t, command); got != "" {
			t.Errorf("command %q: decision = %q, want allow", command, got)
		}
	}
}

// TestDangerousRemovalDeployed_AllowsQuotedDataMention re-runs quoted
// data-mention cases from the builtin-only suite under the deployed policy: a
// quoted span is folded before the pattern scan, so printing or storing a
// dangerous form stays allowed when extras are merged too.
func TestDangerousRemovalDeployed_AllowsQuotedDataMention(t *testing.T) {
	for _, command := range []string{
		`echo "rm -rf /"`,
		`git commit -m "fix(hook): block the removal form regardless of flag order"`,
	} {
		if got := decideBashDeployed(t, command); got != "" {
			t.Errorf("command %q: decision = %q, want allow", command, got)
		}
	}
}

// TestDangerousRemovalDeployed_DeniesProtectedTargetsAllOrders is the deny
// direction under the deployed policy: a removal aimed at the filesystem root,
// a top-level directory, a home alias, or an irreproducible-work directory is
// refused regardless of flag order, flag clustering, quoting, or compound
// command position. The reason is asserted to carry the structural-check
// prefix so the test fails if the structural check stops judging (its
// regression axis) or a pattern reclaims a protected-target decision.
func TestDangerousRemovalDeployed_DeniesProtectedTargetsAllOrders(t *testing.T) {
	for _, command := range []string{
		"rm -fr /",
		"rm -r -f /",
		"rm --recursive --force /",
		`rm -rf "/"`,
		`rm -rf "$HOME"`,
		"rm -fr ~",
		"rm -fr $HOME",
		"rm -fr .git",
		"rm -fr node_modules",
		"rm -rf /usr",
		"echo x && rm -fr /",
	} {
		decision, reason := decideBashDeployedReason(t, command)
		if decision != DecisionDeny {
			t.Errorf("command %q: decision = %q, want %q", command, decision, DecisionDeny)
			continue
		}
		if !strings.HasPrefix(reason, structuralDenyPrefix) {
			t.Errorf("command %q: reason = %q, want prefix %q (structural check must be the judger)", command, reason, structuralDenyPrefix)
		}
	}
}

// TestDangerousRemovalDeployed_DeniesHeredocBodyProtectedTarget characterizes
// current behavior shared with the documented limitation: a heredoc body line
// naming a PROTECTED target is refused, because newline splitting reviews body
// lines as command segments. Correct for an executing consumer, deliberately
// over-blocking for a data consumer — the documented ceiling, not a defect.
func TestDangerousRemovalDeployed_DeniesHeredocBodyProtectedTarget(t *testing.T) {
	for _, command := range []string{
		"cat > /tmp/t511-protected-note.md <<'EOF'\nrm -rf /\nEOF\necho WROTE_OK",
	} {
		decision, reason := decideBashDeployedReason(t, command)
		if decision != DecisionDeny {
			t.Errorf("command %q: decision = %q, want %q", command, decision, DecisionDeny)
			continue
		}
		if !strings.HasPrefix(reason, structuralDenyPrefix) {
			t.Errorf("command %q: reason = %q, want prefix %q (structural check must be the judger)", command, reason, structuralDenyPrefix)
		}
	}
}
