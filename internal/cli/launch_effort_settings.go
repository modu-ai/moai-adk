package cli

// launch_effort_settings.go — the profile effort → injected-settings
// `effortLevel` translation shared by every launcher injection point.
//
// Why the effort no longer travels in CLAUDE_CODE_EFFORT_LEVEL: that variable
// is an OVERRIDE in Claude Code's precedence order, not a default. It beats
// every settings scope AND refuses an in-session change — selecting a new level
// in `/effort` (or through `/model`) answers with
//
//	CLAUDE_CODE_EFFORT_LEVEL=<value> overrides this session — clear it and <new> takes over
//
// so a launcher that pinned the profile's effort there froze the level for the
// whole session. The transient settings file the launcher already injects
// carries the same value as a launch DEFAULT, which an in-session change is
// free to replace — measured on Claude Code v2.1.269 (card t595).
//
// An INHERITED CLAUDE_CODE_EFFORT_LEVEL is deliberately left alone. A user who
// writes `CLAUDE_CODE_EFFORT_LEVEL=max moai cc` is using the documented
// per-session override (.claude/rules/moai/core/settings-management.md), and
// the launcher no longer creates one of its own, so no new pin is manufactured.
//
// Scope: the Claude backend, including every gateway binding (claude / gpt /
// glm gateway modes host Claude Code too — card t668). Under a GLM backend without a gateway z.ai honors
// ANTHROPIC_REASONING_EFFORT and treats Claude's 5-step vocabulary as inert —
// buildEnvForGLMLaunch owns that path and is untouched here.

import "github.com/modu-ai/moai-adk/internal/profile"

// effortSettingsKey is the Claude Code settings key carrying the session's
// launch effort level.
const effortSettingsKey = "effortLevel"

// launchEffortPrefsFn reads the profile preferences the effort resolution uses.
// Tests override it so they never read the host's real profile.
var launchEffortPrefsFn = profile.ReadPreferences

// applyLaunchEffort merges the profile's resolved effort level into an injected
// settings payload and returns it. The resolution is resolveLaunchEffort's, so
// the two profile levers behave exactly as they did on the env path: an
// explicit `effort_level` wins, `model_policy` supplies the fallback, and both
// empty resolves to "" — which injects nothing, leaving the payload (and so the
// launch) byte-identical to a launcher that never knew about effort.
//
// Fail-open: an unreadable profile contributes no effort rather than blocking
// the launch, matching crossSessionSettingsPayload's stance on a bad config.
func applyLaunchEffort(payload map[string]any, profileName string) map[string]any {
	prefs, err := launchEffortPrefsFn(profileName)
	if err != nil {
		return payload
	}
	effort := resolveLaunchEffort(prefs.EffortLevel, prefs.ModelPolicy)
	if effort == "" {
		return payload
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload[effortSettingsKey] = effort
	return payload
}

// buildEnvForClaudeLaunch returns the environment the Claude backend launches
// with. It returns base unchanged — and that is the INVARIANT this seam exists
// to state, not an accident of the current implementation. The profile's effort
// travels in the injected --settings payload, so nothing on this path may add,
// replace, or strip CLAUDE_CODE_EFFORT_LEVEL: adding one restores the override
// that froze the session's effort, and stripping one discards the user's own
// documented per-session override.
//
// The seam is named rather than inlined so the invariant is assertable. A
// wrapper that pins or scrubs the variable turns
// TestClaudeLaunchEnvPreservesInheritedEffort red the moment it lands here —
// which a bare os.Environ() call site could not do.
func buildEnvForClaudeLaunch(base []string) []string {
	return base
}
