package config

import "fmt"

// Source represents a configuration tier in the 8-tier priority system.
// Higher priority sources override lower priority sources.
type Source int

const (
	// SrcPolicy is the highest priority tier - system-wide policy configuration.
	// Location: /etc/moai/settings.json (Linux), /Library/Application Support/moai/settings.json (macOS),
	// %ProgramData%\moai\settings.json (Windows).
	SrcPolicy Source = iota

	// SrcUser is user-specific configuration.
	// Location: ~/.moai/settings.json + ~/.moai/config/sections/*.yaml
	SrcUser

	// SrcProject is project-specific configuration.
	// Location: .moai/config/config.yaml + .moai/config/sections/*.yaml
	SrcProject

	// SrcLocal is local overrides (not committed to git).
	// Location: .claude/settings.local.json + .moai/config/local/*.yaml
	SrcLocal

	// SrcPlugin is plugin-contributed configuration (reserved for v3.1+).
	// No contributors in v3.0 - tier is always empty.
	SrcPlugin

	// SrcSkill is skill-specific configuration from frontmatter.
	// Location: .claude/skills/**/SKILL.md frontmatter `config:` block
	SrcSkill

	// SrcSession is session-scoped runtime configuration.
	// Populated by SPEC-V3R2-RT-004 checkpoint writes (short-lived).
	SrcSession

	// SrcBuiltin is compiled-in defaults.
	// Location: internal/config/defaults.go (lowest priority)
	SrcBuiltin
)

// String returns the human-readable name of the source.
func (s Source) String() string {
	switch s {
	case SrcPolicy:
		return "policy"
	case SrcUser:
		return "user"
	case SrcProject:
		return "project"
	case SrcLocal:
		return "local"
	case SrcPlugin:
		return "plugin"
	case SrcSkill:
		return "skill"
	case SrcSession:
		return "session"
	case SrcBuiltin:
		return "builtin"
	default:
		return "unknown"
	}
}

// Priority returns the numeric priority (lower = higher priority).
// This is useful for sorting and comparison.
func (s Source) Priority() int {
	return int(s)
}

// ParseSource converts a string name to a Source enum value.
// Returns an error if the name is not recognized.
func ParseSource(name string) (Source, error) {
	switch name {
	case "policy":
		return SrcPolicy, nil
	case "user":
		return SrcUser, nil
	case "project":
		return SrcProject, nil
	case "local":
		return SrcLocal, nil
	case "plugin":
		return SrcPlugin, nil
	case "skill":
		return SrcSkill, nil
	case "session":
		return SrcSession, nil
	case "builtin":
		return SrcBuiltin, nil
	default:
		return SrcBuiltin, fmt.Errorf("unknown source name: %s", name)
	}
}

// AllSources returns all sources in priority order (highest to lowest).
func AllSources() []Source {
	return []Source{
		SrcPolicy,
		SrcUser,
		SrcProject,
		SrcLocal,
		SrcPlugin,
		SrcSkill,
		SrcSession,
		SrcBuiltin,
	}
}
