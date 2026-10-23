package cli

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// codexAuditLegEnv marks the codex child of an audit leg (t1788). The Stop chain
// of that child skips every member that asks the model to run a verification
// command: the model is already reviewing inside the leg, and a nested
// verification command is the defect that stalled the leg (t1754).
const codexAuditLegEnv = config.EnvCodexAuditLeg

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

// codexLaneLaunchEnvKeys are the lane launch keys an outer session may
// export. A plain Codex launch scrubs them. An explicit -f launch supplies a
// freshly selected factory identity after the scrub.
var codexLaneLaunchEnvKeys = []string{
	config.EnvFactoryRunID,
	config.EnvFactoryLeadAddr,
	config.EnvFactoryLeadName,
	config.EnvFactoryBackend,
	config.EnvFactoryCard,
	config.EnvFactorySettingsInjected,
	config.EnvMoaiFactoryWorker,
	config.EnvMoaiFactoryWorkers,
	config.EnvFactoryRole,
}

// codexChildEnv is the environment handed to the codex child: the parent's own
// environment with the RESOLVED CODEX_HOME appended. Appending (rather than
// replacing) keeps the rest of the parent environment intact, and last-wins
// makes the appended entry the value the child reads.
//
// Explicit rather than ambient: where the parent has no CODEX_HOME at all,
// there is nothing to inherit, and the child would otherwise resolve its own
// default independently of the value the readout reports.
//
// Removed from the inherited environment: the foreign attribution pair
// (CLAUDE_CODE_SESSION_ID, MOAI_SESSION_PID) and the lane launch keys,
// so a codex run inside a Claude lane never presents that lane's identity
// Posture keys such as MOAI_AUTONOMY_TIER carry no identity and pass through.
func codexChildEnv() []string {
	drop := map[string]bool{config.EnvClaudeCodeSessionID: true, config.EnvMoaiSessionPID: true}
	for _, key := range codexLaneLaunchEnvKeys {
		drop[key] = true
	}
	env := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if drop[key] {
			continue
		}
		env = append(env, entry)
	}
	// resolveCodexHomeDir's second result is the source label, not an error.
	if home, _ := resolveCodexHomeDir(); home != "" {
		env = append(env, codexHomeEnvVar+"="+home)
	}
	return env
}
