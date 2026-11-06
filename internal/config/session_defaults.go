package config

import (
	"strings"
	"unicode"
)

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

// defaultSessionDefaults is the shipped table. Every effort is Claude
// vocabulary. The GLM model is the Claude slot alias (opus serves
// llm.glm.models.high, sonnet medium) and its effort is translated to a z.ai
// reasoning state at launch through the llm.glm.effort table (high -> max,
// medium -> high), so `glm -f` reaches z.ai at max and `glm -l` at high.
func defaultSessionDefaults() SessionDefaults {
	return SessionDefaults{
		Mode: SessionDefaultsModeDefault,
		CC: SessionRolePair{
			Leader: ModelEffort{Model: "opus", Effort: "high"},
			Lane:   ModelEffort{Model: "sonnet", Effort: "medium"},
		},
		GLM: SessionRolePair{
			Leader: ModelEffort{Model: "opus", Effort: "high"},
			Lane:   ModelEffort{Model: "sonnet", Effort: "medium"},
		},
		Codex: SessionRolePair{
			// Card t1980 scope expansion (+ audit repair P1): the shipped codex
			// session pair follows the template llm.yaml cells — terra at the
			// shared effort default, both roles — so `mode: default` resolves
			// to the same values the distributed file restates.
			Leader: ModelEffort{Model: CodexModelTerra, Effort: DefaultCodexAuditEffort},
			Lane:   ModelEffort{Model: CodexModelTerra, Effort: DefaultCodexAuditEffort},
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
	out, _, ok := c.ResolveSessionDefaultChecked(backend, role)
	return out, ok
}

// SessionDefaultRejection names one custom llm.yaml session cell the resolver
// did not use verbatim: either refused because it fails the
// validateSessionDefaults rules (the shipped value is used in its place), or,
// with Normalized set, a model accepted after stripping surrounding whitespace
// and control characters (the trimmed value is used).
type SessionDefaultRejection struct {
	Key        string // llm.defaults.sessions.<backend>.<role>.<field>
	Field      string // "model" or "effort"
	Value      string // the raw configured value
	Fallback   string // the value used instead: the shipped value, or the trimmed one when Normalized
	Normalized bool   // true when the value was accepted after trimming
}

// ResolveSessionDefaultChecked is ResolveSessionDefault plus the cells it
// rejected. A custom cell that config.Validate would refuse never reaches a
// launch: the shipped value for that field is used and the rejection is
// reported so the launcher can warn (fail-open — a launch never blocks on it).
// A model with surrounding whitespace or control characters is used trimmed
// and reported with Normalized set, so the launcher warns about it too.
func (c LLMConfig) ResolveSessionDefaultChecked(backend, role string) (ModelEffort, []SessionDefaultRejection, bool) {
	defPair, ok := defaultSessionDefaults().pair(backend)
	if !ok {
		return ModelEffort{}, nil, false
	}
	out, ok := defPair.role(role)
	if !ok {
		return ModelEffort{}, nil, false
	}
	if strings.TrimSpace(c.Defaults.Sessions.Mode) != SessionDefaultsModeCustom {
		return out, nil, true
	}
	cfgPair, _ := c.Defaults.Sessions.pair(backend)
	cell, _ := cfgPair.role(role)
	prefix := "llm.defaults.sessions." + backend + "." + role
	var rejected []SessionDefaultRejection
	if m := trimSessionModel(cell.Model); m != "" {
		switch {
		case !sessionModelValid(backend, m):
			rejected = append(rejected, SessionDefaultRejection{Key: prefix + ".model", Field: "model", Value: cell.Model, Fallback: out.Model})
		case m != cell.Model:
			out.Model = m
			rejected = append(rejected, SessionDefaultRejection{Key: prefix + ".model", Field: "model", Value: cell.Model, Fallback: m, Normalized: true})
		default:
			out.Model = m
		}
	}
	if e := strings.TrimSpace(cell.Effort); e != "" {
		if sessionClaudeEfforts[e] {
			out.Effort = e
		} else {
			rejected = append(rejected, SessionDefaultRejection{Key: prefix + ".effort", Field: "effort", Value: e, Fallback: out.Effort})
		}
	}
	return out, rejected, true
}

var (
	sessionClaudeModels  = map[string]bool{"opus": true, "sonnet": true, "haiku": true, "fable": true}
	sessionClaudeEfforts = map[string]bool{"low": true, "medium": true, "high": true, "xhigh": true}
)

// trimSessionModel strips leading and trailing whitespace and control
// characters from a session model. config.Validate and the resolver both judge
// the trimmed value, so they accept and refuse the same cells; interior
// whitespace or control characters survive the trim and are refused.
func trimSessionModel(m string) string {
	return strings.TrimFunc(m, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) })
}

// sessionModelValid reports whether m is an acceptable session model for
// backend: a Claude alias (cc) or slot alias (glm) from the closed set, or a
// Codex model id carrying no whitespace, control character, quote, or
// backslash (the value travels as one argv element and must not cross a
// TOML value boundary).
func sessionModelValid(backend, m string) bool {
	if backend != SessionBackendCodex {
		return sessionClaudeModels[m]
	}
	return !strings.ContainsAny(m, "\"'\\") && strings.IndexFunc(m, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) < 0
}

// validateSessionDefaults checks llm.defaults.sessions: the mode closed set,
// and per cell a model and effort. Every effort is Claude vocabulary
// {low, medium, high, xhigh} — max and ultra are never a default; the glm
// effort is translated to a z.ai reasoning state at launch. Empty cells are
// valid (they fall back).
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
			m, e := trimSessionModel(cell.Model), strings.TrimSpace(cell.Effort)
			if m != "" && !sessionModelValid(backend, m) {
				switch backend {
				case SessionBackendCodex:
					add(prefix+".model", "model "+m+" is not a Codex model id", m)
				case SessionBackendGLM:
					add(prefix+".model", "model "+m+" is not a GLM slot alias {opus, sonnet, haiku, fable}", m)
				default:
					add(prefix+".model", "model "+m+" is not in {opus, sonnet, haiku, fable}", m)
				}
			}
			if e != "" && !sessionClaudeEfforts[e] {
				add(prefix+".effort", "effort "+e+" is not in {low, medium, high, xhigh} (max/ultra are never a default)", e)
			}
		}
	}
	return errs
}
