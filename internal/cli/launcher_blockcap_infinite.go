package cli

// launcher_blockcap_infinite.go implements the SPEC-INFINITE-GOAL-001 REQ-2
// (OQ-3) launcher inject path: when an armed goal at --max-turns 0 (infinite)
// exists for the resolving session at launch time, the launcher raises the
// runtime consecutive-Stop-hook-block cap so the infinite loop actually
// persists instead of being silently terminated at the default 8.
//
// The cap is a Claude Code RUNTIME env (not MoAI-owned); this is the MoAI-side
// convenience that makes "arm and walk away" work. The doctrine guide
// (goal-directive.md / goal.md) remains the load-bearing deliverable; this
// inject is the convenience.

import (
	"context"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/goal"
)

// DefaultRaisedStopHookBlockCap is the raised cap injected when an infinite
// goal is armed. It comfortably exceeds the goal engine's DefaultMaxTurns (30)
// and a long unattended run's expected turn count, so the runtime cap no longer
// pre-empts the goal loop before a real bound (wall-clock / cost / stagnation)
// fires. Tunable in principle; held as a named constant per CLAUDE.local.md §14.
const DefaultRaisedStopHookBlockCap = 200

// injectStopHookBlockCapForGoal returns base with CLAUDE_CODE_STOP_HOOK_BLOCK_CAP
// set to DefaultRaisedStopHookBlockCap when an armed MaxTurns==0 goal exists for
// sessionID under projectRoot; otherwise it returns base unchanged (backward
// compat — a finite or absent goal never triggers the inject). A pre-existing
// cap entry in base is replaced (the raised value wins for the infinite case);
// when no infinite goal is armed, any pre-existing user-set cap is preserved.
//
// The goal read is best-effort + fail-open: a read error degrades to no inject
// (the launch must never block on a goal-state read).
func injectStopHookBlockCapForGoal(ctx context.Context, base []string, projectRoot, sessionID string) []string {
	_ = ctx
	// SPEC-FACTORY-MODE-001 REQ-FM-023: the factory branch is UNCONDITIONAL and
	// sits ahead of the goal read. The goal-conditional branch below reads goal
	// state at launch time; a factory chain arms its goal mid-session, so that
	// predicate is structurally unable to see it and the chain would otherwise
	// stay capped at the runtime default of 8.
	//
	// A COMPANION of a factory run has the same problem — it arms its own goal
	// mid-session too — so it takes the same raise. It is signalled by the label
	// variable rather than the factory one because it must not be seeded with
	// the chain, which only the lead drives.
	if os.Getenv(config.EnvMoaiFactory) != "" || os.Getenv(config.EnvMoaiFactoryLabel) != "" {
		return setStopHookBlockCap(base, DefaultRaisedStopHookBlockCap)
	}
	if projectRoot == "" || sessionID == "" {
		return base
	}
	g, err := goal.LoadGoal(projectRoot, sessionID)
	if err != nil || g == nil {
		return base // absent or unreadable → no inject (fail-open)
	}
	if g.Status != goal.StatusArmed || g.Ceiling.MaxTurns != 0 {
		return base // not an armed infinite goal → backward compat
	}
	return setStopHookBlockCap(base, DefaultRaisedStopHookBlockCap)
}

// setStopHookBlockCap returns base with the runtime block-cap entry set to cap,
// replacing any pre-existing entry in place rather than appending a duplicate
// key (which a child process would resolve ambiguously).
func setStopHookBlockCap(base []string, cap int) []string {
	key := config.EnvClaudeCodeStopHookBlockCap
	entry := key + "=" + strconvItoa(cap)
	result := make([]string, 0, len(base)+1)
	replaced := false
	for _, e := range base {
		if strings.HasPrefix(e, key+"=") {
			result = append(result, entry)
			replaced = true
		} else {
			result = append(result, e)
		}
	}
	if !replaced {
		result = append(result, entry)
	}
	return result
}

// strconvItoa is a thin local wrapper to avoid importing strconv in this file
// (keeps the inject path dependency-light). Equivalent to strconv.Itoa.
func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// resolveLaunchSessionID resolves the session id the launcher keys on for the
// block-cap inject. It mirrors the goal CLI's resolution: --session override
// wins; otherwise the side-channel current-session id. Returns "" when none
// resolves (the inject then no-ops). Best-effort: never blocks the launch.
//
// NOTE: this is intentionally a thin read of the SAME side channel the goal arm
// path uses, so the launcher and the arm verb key identically. It does NOT
// re-derive the writer_pid fallback (an armed pid-keyed goal is unreachable by
// the hook anyway, so injecting for it would be noise).
func resolveLaunchSessionID(sessionOverride string) string {
	if sessionOverride != "" {
		return sessionOverride
	}
	if id, _, ok := resolveCurrentSessionID(); ok {
		return id
	}
	return ""
}

// launchProjectRoot resolves the project root for the goal-state read. It reuses
// the goal CLI's resolver so the inject reads the SAME .moai/state/goal/ tree.
func launchProjectRoot() string {
	return resolveProjectDir()
}
