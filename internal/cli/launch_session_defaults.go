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
//
// An explicit operator --model / --effort wins.

import (
	"fmt"
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

// warnSessionRejections writes one launcher warning per rejected llm.yaml
// session cell of the given field ("" = every field): the key, the rejected
// value, and the built-in default the launch uses instead (or, for a model
// normalized by trimming, the trimmed value it uses). The launch proceeds —
// an invalid cell never blocks it and never reaches argv or env.
func warnSessionRejections(rejected []config.SessionDefaultRejection, field string) {
	for _, r := range rejected {
		if field != "" && r.Field != field {
			continue
		}
		if r.Normalized {
			_, _ = fmt.Fprintf(launcherStderr, "Warning: %s %q has surrounding whitespace or control characters; using %q\n",
				r.Key, r.Value, r.Fallback)
			continue
		}
		_, _ = fmt.Fprintf(launcherStderr, "Warning: %s %q is not a valid session %s; using the built-in default %q\n",
			r.Key, r.Value, r.Field, r.Fallback)
	}
}

// factorySessionDefaultChecked resolves the session default for a
// Claude-hosted factory launch (cc or glm backend), plus the llm.yaml cells
// the resolver rejected and replaced with the built-in default. ok is false
// outside a factory session.
func factorySessionDefaultChecked(env []string, root string) (backend string, me config.ModelEffort, rejected []config.SessionDefaultRejection, ok bool) {
	role := factorySessionRole(env)
	if role == "" {
		return "", config.ModelEffort{}, nil, false
	}
	switch launchEnvValue(env, config.EnvFactoryBackend) {
	case factory.BackendGLM:
		backend = config.SessionBackendGLM
	case factory.BackendClaude, "":
		backend = config.SessionBackendCC
	default:
		return "", config.ModelEffort{}, nil, false
	}
	me, rejected, ok = sessionLLMConfig(root).ResolveSessionDefaultChecked(backend, role)
	return backend, me, rejected, ok
}

// factorySessionModel returns the launch model: an explicit --model wins;
// inside a factory session the session default replaces the stored/profile
// model; otherwise current is kept. A rejected llm.yaml model is warned about
// here (the model half's single consumer) and replaced by the built-in default.
func factorySessionModel(env []string, root string, explicit bool, current string) string {
	if explicit {
		return current
	}
	if _, me, rejected, ok := factorySessionDefaultChecked(env, root); ok && me.Model != "" {
		warnSessionRejections(rejected, "model")
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
	backend, me, rejected, ok := factorySessionDefaultChecked(env, root)
	if !ok || backend != config.SessionBackendGLM || me.Effort == "" {
		return "", false
	}
	warnSessionRejections(rejected, "effort")
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
	if backend, me, rejected, ok := factorySessionDefaultChecked(env, root); ok && backend == config.SessionBackendCC && me.Effort != "" {
		warnSessionRejections(rejected, "effort")
		return me.Effort, true
	}
	return "", false
}
