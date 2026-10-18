package cli

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// codexAuditLegEnv marks the codex child of an audit leg (t1788). The Stop chain
// of that child skips every member that asks the model to run a verification
// command: the model is already reviewing inside the leg, and a nested
// verification command is the defect that stalled the leg (t1754).
const codexAuditLegEnv = config.EnvCodexAuditLeg

// codexAuditLegSkipped names the Stop-chain members whose deny or continuation
// text asks the model to run a verification command: sync-gate (2), goal (3),
// and codex-review (6). Multi-review (7) reports only a residual-risk note.
var codexAuditLegSkipped = map[int]string{
	2: "sync-gate",
	3: "goal",
	6: "codex-review",
}

// stopStatusSkippedAuditLeg records a member skipped inside an audit leg. A skip
// is never a pass.
const stopStatusSkippedAuditLeg = "skipped-audit-leg"

// codexAuditLegCtxKey carries the leg marker on the spawner's context.
type codexAuditLegCtxKey struct{}

// withCodexAuditLeg marks ctx as an audit leg. Only the leg spawner reads the
// key; the child receives the marker through its environment.
func withCodexAuditLeg(ctx context.Context) context.Context {
	return context.WithValue(ctx, codexAuditLegCtxKey{}, true)
}

// codexLegChildEnv returns the environment for a codex child that an audit leg
// spawns, or nil when the caller is not a leg.
func codexLegChildEnv(ctx context.Context) []string {
	if ctx.Value(codexAuditLegCtxKey{}) == nil {
		return nil
	}
	return append(os.Environ(), codexAuditLegEnv+"=1")
}

// orcaPaneOwnedEnv names the Orca variables that bind a process to the pane that
// launched it (t1869). Exact keys only; the hook family is matched by prefix.
// Kept on purpose, one line each: ORCA_APP_VERSION and ORCA_CLI_BIN_DIR and
// ORCA_USER_DATA_PATH describe the install, not a pane; ORCA_IMAGE_PROTOCOL is a
// terminal capability; ORCA_OMP_*, ORCA_PI_SOURCE_AGENT_DIR and ORCA_OPENCODE_AGENT
// configure other agents' tooling and carry no pane identity.
var orcaPaneOwnedEnv = map[string]bool{
	config.EnvOrcaPaneKey:        true,
	config.EnvOrcaTabID:          true,
	config.EnvOrcaWorktreeID:     true,
	config.EnvOrcaAgentLaunchTok: true,
	config.EnvOrcaAgentPane:      true,
	config.EnvOrcaTerminalHandle: true,
}

// isOrcaPaneOwnedEnvKey reports whether key is pane-owned: an exact member of
// orcaPaneOwnedEnv or any key starting with ORCA_AGENT_HOOK_. The match is
// case-sensitive and anchored at the start, so ORCA_PANE_KEYS and XORCA_PANE_KEY
// are not pane-owned.
func isOrcaPaneOwnedEnvKey(key string) bool {
	return orcaPaneOwnedEnv[key] || strings.HasPrefix(key, config.EnvOrcaAgentHookPrefix)
}

// dropOrcaPaneEnv returns env without the pane-owned Orca entries. The input is
// not modified.
func dropOrcaPaneEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if isOrcaPaneOwnedEnvKey(key) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

// codexReviewChildEnv is the environment of a codex review/audit/task child:
// the parent's environment (plus the audit-leg marker for a leg) with the Orca
// pane-owned variables removed, so the child never reports into the parent pane.
func codexReviewChildEnv(ctx context.Context) []string {
	if env := codexLegChildEnv(ctx); env != nil {
		return dropOrcaPaneEnv(env)
	}
	return dropOrcaPaneEnv(os.Environ())
}

// codexAuditChildEnv composes the launcher-identity drop (codexChildEnv) with the
// Orca pane drop, for audit children. codexChildEnv itself is unchanged: the
// interactive lane launcher legitimately runs inside an Orca pane.
func codexAuditChildEnv() []string {
	return dropOrcaPaneEnv(codexChildEnv())
}

// codexChildCmd builds the codex app-server command from codexReviewChildEnv.
func codexChildCmd(ctx context.Context, binaryPath string, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Env = codexReviewChildEnv(ctx)
	return cmd
}

// inCodexAuditLeg reports whether this process runs inside an audit leg.
func inCodexAuditLeg() bool {
	return os.Getenv(codexAuditLegEnv) == "1"
}

// auditLegSkipsMember reports whether Stop-chain member n is skipped here.
func auditLegSkipsMember(n int) bool {
	if _, ok := codexAuditLegSkipped[n]; !ok {
		return false
	}
	return inCodexAuditLeg()
}

// auditLegSkipOutcome is a skipped member's result. It allows the turn and states
// the skip on the visible output, so the rollout shows it rather than the skip
// passing silently (t1788 brain ruling, condition b).
func auditLegSkipOutcome(n int) stopMemberOutcome {
	spec := stopMemberSpec(n)
	note := spec.Name + ": skipped: audit leg (verification is not run inside a codex audit leg)"
	return stopMemberOutcome{
		Number:   n,
		Name:     spec.Name,
		Decision: codexadapter.DecisionAllow,
		Status:   stopStatusSkippedAuditLeg,
		Advisory: note,
		Discards: []codexadapter.Discard{{
			Event:  hook.EventStop,
			Key:    "codex-stop-chain/" + spec.Name + "/audit-leg",
			Reason: note,
		}},
	}
}
