package cli

import (
	"context"
	"os"
	"os/exec"

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
// spawns, or nil so the child inherits the parent environment unchanged.
func codexLegChildEnv(ctx context.Context) []string {
	if ctx.Value(codexAuditLegCtxKey{}) == nil {
		return nil
	}
	return append(os.Environ(), codexAuditLegEnv+"=1")
}

// codexChildCmd builds the codex app-server command. A leg's child gets the audit
// marker in its environment; every other child inherits the parent environment.
func codexChildCmd(ctx context.Context, binaryPath string, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	if env := codexLegChildEnv(ctx); env != nil {
		cmd.Env = env
	}
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
