package config

import "strings"

// session_defaults.go — launch defaults for factory sessions (card t1808,
// operator per-role defaults).
//
// Each launcher's leader (-f) and lane (-l) session starts from a model and
// effort pair. In `default` mode the shipped table below applies; in `custom`
// mode the llm.defaults.sessions cells in llm.yaml apply (an empty cell falls
// back to the shipped value), from the next launch on. The values travel only
// through channels that leave in-session changes free: the Claude launch flag
// and transient settings file, Codex `-m` / `-c`, and the GLM slot alias plus
// the z.ai reasoning mapping — never CLAUDE_CODE_EFFORT_LEVEL.

// llm.defaults.sessions.mode values.
const (
	SessionDefaultsModeDefault = "default"
	SessionDefaultsModeCustom  = "custom"
)

// Launcher backends and session roles a session default is keyed by.
const (
	SessionBackendCC    = "cc"
	SessionBackendGLM   = "glm"
	SessionBackendCodex = "codex"

	SessionRoleLeader = "leader"
	SessionRoleLane   = "lane"
)

// LLMDefaults groups the llm.defaults block.
type LLMDefaults struct {
	Sessions SessionDefaults `yaml:"sessions"`
}

// SessionDefaults is llm.defaults.sessions: the mode switch plus one
// leader/lane pair per launcher.
type SessionDefaults struct {
	Mode  string          `yaml:"mode"`
	CC    SessionRolePair `yaml:"cc"`
	GLM   SessionRolePair `yaml:"glm"`
	Codex SessionRolePair `yaml:"codex"`
}

// SessionRolePair is the leader and lane launch default of one launcher.
type SessionRolePair struct {
	Leader ModelEffort `yaml:"leader"`
	Lane   ModelEffort `yaml:"lane"`
}

// defaultSessionDefaults is the shipped table. The GLM model is the Claude
// slot alias (opus serves llm.glm.models.high, sonnet medium); its effort is a
// z.ai reasoning state. The glm leader's max is a launch mapping, not a
// shipped Claude/Codex default effort (TestNoShippedDefaultEffortIsMaxOrUltra).
func defaultSessionDefaults() SessionDefaults {
	return SessionDefaults{
		Mode: SessionDefaultsModeDefault,
		CC: SessionRolePair{
			Leader: ModelEffort{Model: "opus", Effort: "high"},
			Lane:   ModelEffort{Model: "sonnet", Effort: "medium"},
		},
		GLM: SessionRolePair{
			Leader: ModelEffort{Model: "opus", Effort: "max"},
			Lane:   ModelEffort{Model: "sonnet", Effort: "high"},
		},
		Codex: SessionRolePair{
			Leader: ModelEffort{Model: CodexModelSol, Effort: "high"},
			Lane:   ModelEffort{Model: CodexModelSol, Effort: "medium"},
		},
	}
}

func (s SessionDefaults) pair(backend string) (SessionRolePair, bool) {
	switch backend {
	case SessionBackendCC:
		return s.CC, true
	case SessionBackendGLM:
		return s.GLM, true
	case SessionBackendCodex:
		return s.Codex, true
	}
	return SessionRolePair{}, false
}

func (p SessionRolePair) role(role string) (ModelEffort, bool) {
	switch role {
	case SessionRoleLeader:
		return p.Leader, true
	case SessionRoleLane:
		return p.Lane, true
	}
	return ModelEffort{}, false
}

// ResolveSessionDefault returns the launch model and effort for a launcher
// backend (cc, glm, codex) and session role (leader, lane). ok is false for an
// unknown backend or role. In custom mode each non-empty llm.yaml cell
// replaces the shipped value field by field; any other mode yields the
// shipped table.
//
// @MX:ANCHOR: [AUTO] single resolver of factory session launch defaults
// @MX:REASON: the cc/glm settings injection, the Claude launch model, and the codex launcher argv all read it; a divergent copy would let the launchers disagree
func (c LLMConfig) ResolveSessionDefault(backend, role string) (ModelEffort, bool) {
	defPair, ok := defaultSessionDefaults().pair(backend)
	if !ok {
		return ModelEffort{}, false
	}
	out, ok := defPair.role(role)
	if !ok {
		return ModelEffort{}, false
	}
	if strings.TrimSpace(c.Defaults.Sessions.Mode) != SessionDefaultsModeCustom {
		return out, true
	}
	cfgPair, _ := c.Defaults.Sessions.pair(backend)
	cell, _ := cfgPair.role(role)
	if m := strings.TrimSpace(cell.Model); m != "" {
		out.Model = m
	}
	if e := strings.TrimSpace(cell.Effort); e != "" {
		out.Effort = e
	}
	return out, true
}

var (
	sessionClaudeModels  = map[string]bool{"opus": true, "sonnet": true, "haiku": true, "fable": true}
	sessionClaudeEfforts = map[string]bool{"low": true, "medium": true, "high": true, "xhigh": true}
	sessionGLMEfforts    = map[string]bool{"low": true, "high": true, "max": true}
)

// validateSessionDefaults checks llm.defaults.sessions: the mode closed set,
// and per cell a model and effort in the launcher's vocabulary. cc and codex
// efforts exclude max and ultra (never a default); the glm effort is a z.ai
// reasoning state {low, high, max}. Empty cells are valid (they fall back).
func validateSessionDefaults(cfg *Config) []ValidationError {
	s := cfg.LLM.Defaults.Sessions
	var errs []ValidationError
	add := func(field, msg, value string) {
		errs = append(errs, ValidationError{Field: field, Message: msg, Value: value, Wrapped: ErrInvalidConfig})
	}
	switch strings.TrimSpace(s.Mode) {
	case "", SessionDefaultsModeDefault, SessionDefaultsModeCustom:
	default:
		add("llm.defaults.sessions.mode", "mode "+s.Mode+" is not in the closed set {default, custom}", s.Mode)
	}
	for _, backend := range []string{SessionBackendCC, SessionBackendGLM, SessionBackendCodex} {
		pair, _ := s.pair(backend)
		for _, role := range []string{SessionRoleLeader, SessionRoleLane} {
			cell, _ := pair.role(role)
			prefix := "llm.defaults.sessions." + backend + "." + role
			m, e := strings.TrimSpace(cell.Model), strings.TrimSpace(cell.Effort)
			switch backend {
			case SessionBackendCodex:
				if m != "" && strings.ContainsAny(m, " \t\"'\\") {
					add(prefix+".model", "model "+m+" is not a Codex model id", m)
				}
				if e != "" && !sessionClaudeEfforts[e] {
					add(prefix+".effort", "effort "+e+" is not in {low, medium, high, xhigh} (max/ultra are never a default)", e)
				}
			case SessionBackendGLM:
				if m != "" && !sessionClaudeModels[m] {
					add(prefix+".model", "model "+m+" is not a GLM slot alias {opus, sonnet, haiku, fable}", m)
				}
				if e != "" && !sessionGLMEfforts[e] {
					add(prefix+".effort", "effort "+e+" is not a z.ai reasoning state {low, high, max}", e)
				}
			default:
				if m != "" && !sessionClaudeModels[m] {
					add(prefix+".model", "model "+m+" is not in {opus, sonnet, haiku, fable}", m)
				}
				if e != "" && !sessionClaudeEfforts[e] {
					add(prefix+".effort", "effort "+e+" is not in {low, medium, high, xhigh} (max/ultra are never a default)", e)
				}
			}
		}
	}
	return errs
}
