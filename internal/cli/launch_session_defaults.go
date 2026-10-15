package cli

// launch_session_defaults.go — factory session launch defaults (card t1808).
//
// A factory leader (-f) and lane (-l) start from the llm.defaults.sessions
// model + effort pair (config.LLMConfig.ResolveSessionDefault). Every value
// travels as a launch DEFAULT that an in-session /model or /effort change can
// replace:
//
//   - Claude (cc): --model <alias> plus effortLevel in the transient
//     --settings payload (launch_effort_settings.go) — never
//     CLAUDE_CODE_EFFORT_LEVEL, which refuses an in-session change.
//   - GLM: the slot alias as --model (routing through the
//     ANTHROPIC_DEFAULT_*_MODEL slot env); the Claude-vocabulary effort is
//     translated through llm.glm.effort (high -> max, medium -> high) and the
//     resulting z.ai state is wired verbatim (buildEnvForGLMLaunchState).
//   - Codex: -m <model> and -c model_reasoning_effort="<effort>" on the lane
//     (`moai codex -l`) only; a plain `moai codex` keeps .codex/config.toml.
//
// An explicit operator --model / --effort / -m / -c effort wins.

import (
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// factorySessionRole reads the factory session role the launch env carries:
// a lane is stamped with the lane marker or worker label, a leader with the
// run id plus the lane count. Anything else is not a factory session.
func factorySessionRole(env []string) string {
	if launchEnvValue(env, config.EnvFactoryRole) == config.FactoryRoleLane ||
		strings.TrimSpace(launchEnvValue(env, config.EnvMoaiFactoryWorker)) != "" {
		return config.SessionRoleLane
	}
	if strings.TrimSpace(launchEnvValue(env, config.EnvFactoryRunID)) != "" &&
		strings.TrimSpace(launchEnvValue(env, config.EnvMoaiFactoryWorkers)) != "" {
		return config.SessionRoleLeader
	}
	return ""
}

// sessionLLMConfig loads the llm section from root, falling back to the
// compiled defaults on any read error (fail-open: a launch never blocks on it).
func sessionLLMConfig(root string) config.LLMConfig {
	if root != "" {
		if llm, err := loadLLMSectionOnly(filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)); err == nil {
			return llm
		}
	}
	return config.NewDefaultLLMConfig()
}

// factorySessionDefault resolves the session default for a Claude-hosted
// factory launch (cc or glm backend). ok is false outside a factory session.
func factorySessionDefault(env []string, root string) (backend string, me config.ModelEffort, ok bool) {
	role := factorySessionRole(env)
	if role == "" {
		return "", config.ModelEffort{}, false
	}
	switch launchEnvValue(env, config.EnvFactoryBackend) {
	case factory.BackendGLM:
		backend = config.SessionBackendGLM
	case factory.BackendClaude, "":
		backend = config.SessionBackendCC
	default:
		return "", config.ModelEffort{}, false
	}
	me, ok = sessionLLMConfig(root).ResolveSessionDefault(backend, role)
	return backend, me, ok
}

// factorySessionModel returns the launch model: an explicit --model wins;
// inside a factory session the session default replaces the stored/profile
// model; otherwise current is kept.
func factorySessionModel(env []string, root string, explicit bool, current string) string {
	if explicit {
		return current
	}
	if _, me, ok := factorySessionDefault(env, root); ok && me.Model != "" {
		return me.Model
	}
	return current
}

// factorySessionGLMState returns the z.ai reasoning state a glm factory
// launch wires: the session default's Claude-vocabulary effort translated
// through the llm.glm.effort tier table (low -> tier.low, medium ->
// tier.medium, high and xhigh -> tier.high), an empty tier cell falling back
// to the shipped config.DefaultGLMEffort* value. ok is false for any launch
// that is not a glm factory session.
func factorySessionGLMState(env []string, root string, tier config.GLMTierEffort) (string, bool) {
	backend, me, ok := factorySessionDefault(env, root)
	if !ok || backend != config.SessionBackendGLM || me.Effort == "" {
		return "", false
	}
	pick := func(stored, def string) string {
		if s := strings.TrimSpace(stored); s != "" {
			return s
		}
		return def
	}
	switch me.Effort {
	case "low":
		return pick(tier.Low, config.DefaultGLMEffortLow), true
	case "medium":
		return pick(tier.Medium, config.DefaultGLMEffortMedium), true
	default: // high, xhigh
		return pick(tier.High, config.DefaultGLMEffortHigh), true
	}
}

// factorySessionClaudeEffort returns the Claude effort a cc factory launch
// injects; ok is false when the profile effort applies instead.
func factorySessionClaudeEffort(env []string, root string) (string, bool) {
	if backend, me, ok := factorySessionDefault(env, root); ok && backend == config.SessionBackendCC && me.Effort != "" {
		return me.Effort, true
	}
	return "", false
}

// codexSessionDefaultArgsFn is the seam the codex launch paths call; the
// launch-mechanics tests that pin an exact argv neutralize it the way they
// neutralize codexMoaiHomeAddDirArgs, and TestCodexLaunchCarriesSessionDefaults
// restores the real body.
var codexSessionDefaultArgsFn = codexSessionDefaultArgs

// codexSessionDefaultArgs returns the -m / -c model_reasoning_effort argv a
// `moai codex` launch prepends for role (leader = the CLI session, lane =
// `moai codex -l`), skipping each half the operator already supplied in tail.
func codexSessionDefaultArgs(root, role string, tail []string) []string {
	me, ok := sessionLLMConfig(root).ResolveSessionDefault(config.SessionBackendCodex, role)
	if !ok {
		return nil
	}
	var out []string
	if me.Model != "" && !codexTailHasModel(tail) {
		out = append(out, "-m", me.Model)
	}
	if me.Effort != "" && !codexTailHasEffort(tail) {
		out = append(out, "-c", `model_reasoning_effort="`+me.Effort+`"`)
	}
	return out
}

func codexTailHasModel(tail []string) bool {
	for _, a := range tail {
		if a == "-m" || a == "--model" || strings.HasPrefix(a, "--model=") || strings.HasPrefix(a, "-m=") {
			return true
		}
	}
	return false
}

func codexTailHasEffort(tail []string) bool {
	for i, a := range tail {
		v := ""
		switch {
		case (a == "-c" || a == "--config") && i+1 < len(tail):
			v = tail[i+1]
		case strings.HasPrefix(a, "--config="):
			v = strings.TrimPrefix(a, "--config=")
		}
		if strings.HasPrefix(strings.TrimSpace(v), "model_reasoning_effort") {
			return true
		}
	}
	return false
}
