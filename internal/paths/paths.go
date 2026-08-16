// Package paths owns the single resolution point for the ~/.moai directory
// tree (SPEC-V3R6-MOAI-HOME-PATHS-001).
//
// MoaiHome honours the MOAI_HOME environment variable as a full replacement
// of the ~/.moai root when it holds a non-empty absolute path; empty values
// are treated as unset and relative values are disregarded (XDG semantics,
// REQ-MHP-001). The fallback resolves the user's home HOME-first —
// os.Getenv("HOME") when non-empty, else os.UserHomeDir() — the contract
// previously owned by internal/cli/homedir.go (REQ-MHP-002), so
// t.Setenv("HOME", ...) overrides work on every platform.
//
// The package imports the standard library only (REQ-MHP-003) so that
// internal/glmcred can adopt it without participating in an import cycle.
// Every accessor resolves on each call (REQ-MHP-005): env changes made by
// later calls in the same process must be observed, which rules out any
// package-level result holder.
package paths

import (
	"os"
	"path/filepath"
)

// EnvHome mirrors config.EnvHome (internal/config/envkeys.go — the canonical
// owner). Declared locally because importing internal/config from this
// package would create an import cycle, the same alias pattern glmcred uses
// for EnvTestGLMKey (REQ-MHP-004). Exported so sibling stdlib-only packages
// can reference the name without re-declaring the literal, which AC-MHP-006
// forbids outside envkeys.go and this package.
const EnvHome = "MOAI_HOME"

// Segment constants mirroring internal/defs/dirs.go, the canonical owner of
// the MoAI directory names. Local aliases keep this package importable with
// standard-library dependencies only (REQ-MHP-007 + REQ-MHP-003): importing
// internal/defs would add a non-stdlib entry to go list -deps.
const (
	moaiDir          = ".moai"           // defs.MoAIDir
	stateSubdir      = "state"           // defs.StateSubdir
	cacheSubdir      = "cache"           // defs.CacheSubdir
	releasesSubdir   = "releases"        // defs.ReleasesSubdir
	worktreesSubdir  = "worktrees"       // defs.WorktreesSubdir
	profilesSubdir   = "claude-profiles" // defs.ClaudeProfilesSubdir
	sectionsSubdir   = "config/sections" // defs.SectionsSubdir
	glmEnvName       = ".env.glm"        // defs.GlmEnvFileName
	userSettingsName = "settings.json"   // defs.UserSettingsFileName
)

// Home returns the user's home directory, HOME-first: os.Getenv("HOME")
// when non-empty, else os.UserHomeDir() (REQ-MHP-002).
//
// @MX:ANCHOR: [AUTO] home-resolution SSOT — HOME-first contract for every ~/.moai consumer
// @MX:REASON: [AUTO] fan_in spans 7 packages; a second home style would drift the security whitelist and the state tree apart (SPEC-V3R6-MOAI-HOME-PATHS-001 REQ-MHP-002/010)
func Home() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	return os.UserHomeDir()
}

// MoaiHome returns the ~/.moai root (REQ-MHP-001/002). A non-empty absolute
// MOAI_HOME is returned verbatim as the root; empty equals unset and relative
// values are disregarded, falling back to Home() joined with the .moai
// segment. On resolution failure it returns ("", err) — never a "."-style
// relative fallback (REQ-MHP-006).
//
// @MX:ANCHOR: [AUTO] ~/.moai root SSOT — the only MOAI_HOME-aware entry point
// @MX:REASON: [AUTO] fan_in across all migrated sites; an override honored at one site but not another splits state across two roots (SPEC-V3R6-MOAI-HOME-PATHS-001 REQ-MHP-001)
func MoaiHome() (string, error) {
	if v := os.Getenv(EnvHome); v != "" && filepath.IsAbs(v) {
		return v, nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, moaiDir), nil
}

// StateDir returns ~/.moai/state.
func StateDir() (string, error) { return joinUnderMoaiHome(stateSubdir) }

// CacheDir returns ~/.moai/cache.
func CacheDir() (string, error) { return joinUnderMoaiHome(cacheSubdir) }

// ReleasesDir returns ~/.moai/releases.
func ReleasesDir() (string, error) { return joinUnderMoaiHome(releasesSubdir) }

// WorktreesDir returns ~/.moai/worktrees.
func WorktreesDir() (string, error) { return joinUnderMoaiHome(worktreesSubdir) }

// ProfilesDir returns ~/.moai/claude-profiles.
func ProfilesDir() (string, error) { return joinUnderMoaiHome(profilesSubdir) }

// GlmEnvFile returns ~/.moai/.env.glm.
func GlmEnvFile() (string, error) { return joinUnderMoaiHome(glmEnvName) }

// UserSettingsFile returns the user-tier ~/.moai/settings.json.
func UserSettingsFile() (string, error) { return joinUnderMoaiHome(userSettingsName) }

// UserConfigSectionsDir returns the user-tier ~/.moai/config/sections.
func UserConfigSectionsDir() (string, error) { return joinUnderMoaiHome(sectionsSubdir) }

// joinUnderMoaiHome joins segment onto the MoaiHome root, propagating the
// root-resolution error unchanged so callers keep their own degradation
// behavior (REQ-MHP-006).
//
// @MX:NOTE: [AUTO] deeper segments (e.g. "state/loop") join at the call site via the accessor, not here (plan.md §4 resolved accessor granularity)
func joinUnderMoaiHome(segment string) (string, error) {
	root, err := MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, segment), nil
}
