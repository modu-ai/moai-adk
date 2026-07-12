package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

const profilesDir = ".moai/claude-profiles"

// Named constants for the last-used-profile ledger. These avoid magic strings
// (CLAUDE.local.md §14) and document the single source of truth for the
// launch.yaml contract.
const (
	// launchLedgerFile is the filename of the last-used-profile ledger under
	// GetBaseDir(). The file is read-modify-written so legacy keys (e.g.
	// model:/bypass:) are preserved alongside last_profile.
	launchLedgerFile = "launch.yaml"

	// lastProfileKey is the YAML key holding the most recently -p-launched
	// named profile. Only named profiles are recorded; "" and "default" are
	// refused by RecordLastUsedProfile.
	lastProfileKey = "last_profile"
)

// BaseDirOverride allows tests to inject a custom base directory.
// When non-empty, GetBaseDir returns this value instead of ~/.moai/claude-profiles.
var BaseDirOverride string

// GetBaseDir returns ~/.moai/claude-profiles/.
func GetBaseDir() string {
	if BaseDirOverride != "" {
		return BaseDirOverride
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
		return "."
	}
	return filepath.Join(home, profilesDir)
}

// GetCurrentName returns the current profile name based on CLAUDE_CONFIG_DIR.
func GetCurrentName() string {
	configDir := os.Getenv(config.EnvClaudeConfigDir)
	if configDir == "" {
		return "default"
	}

	baseDir := GetBaseDir()

	rel, err := filepath.Rel(baseDir, configDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return configDir
	}

	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	return parts[0]
}

// ProfileEntry represents a single profile in the list.
type ProfileEntry struct {
	Name    string
	Current bool
}

// List returns all profile names with an indicator for the current one.
func List() []ProfileEntry {
	currentProfile := GetCurrentName()
	baseDir := GetBaseDir()

	var entries []ProfileEntry
	entries = append(entries, ProfileEntry{
		Name:    "default",
		Current: currentProfile == "default",
	})

	dirEntries, err := os.ReadDir(baseDir)
	if err != nil {
		return entries
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		entries = append(entries, ProfileEntry{
			Name:    name,
			Current: name == currentProfile,
		})
	}

	return entries
}

// Delete removes a profile directory. Returns error if it's the default profile
// or doesn't exist.
func Delete(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}

	profileDir := filepath.Join(GetBaseDir(), name)

	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile %q does not exist", name)
	}

	currentProfile := GetCurrentName()
	if currentProfile == name {
		fmt.Fprintf(os.Stderr, "Warning: %q is the currently active profile\n", name)
		fmt.Fprintf(os.Stderr, "Run: moai cc (without -p) to use default\n")
	}

	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Deleted profile: %s\n", name)
	return nil
}

// GetProfileDir returns the directory path for a named profile without creating it.
// Returns an empty string for invalid profile names.
//
// REQ-SEC-001 (SPEC-INTERNAL-SECURITY-001): central traversal guard — a name
// with path separators, traversal sequences, or that is absolute returns "" so
// no caller can construct a directory path outside the profile base.
func GetProfileDir(name string) string {
	if name == "" || name == "default" {
		return ""
	}
	if !isValidProfileName(name) {
		return ""
	}
	return filepath.Join(GetBaseDir(), name)
}

// isValidProfileName checks that a profile name does not contain path traversal
// characters. Names must not contain slashes, backslashes, or start with a dot.
// The two special names "" (base store) and "default" are valid — they resolve to
// the base preferences.yaml, not a named subdirectory, so they never traverse.
func isValidProfileName(name string) bool {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") ||
		strings.HasPrefix(name, ".") || filepath.IsAbs(name) {
		return false
	}
	return true
}

// IsValidProfileName is the exported guard over isValidProfileName. Callers
// outside the profile package (e.g. the web console write boundary,
// SPEC-WEB-CONSOLE-011 REQ-WC11-031) reuse this predicate rather than
// re-declaring the traversal rules.
func IsValidProfileName(name string) bool {
	return isValidProfileName(name)
}

// EnsureDir creates the profile directory if it doesn't exist and sets
// CLAUDE_CONFIG_DIR in the current process.
func EnsureDir(name string) error {
	if name == "" || name == "default" {
		return nil
	}
	// Validate: no path traversal possible
	if !isValidProfileName(name) {
		return fmt.Errorf("invalid profile name %q: must not contain path separators or start with '.'", name)
	}
	profileDir := filepath.Join(GetBaseDir(), name)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}
	if err := os.Setenv("CLAUDE_CONFIG_DIR", profileDir); err != nil {
		return fmt.Errorf("set CLAUDE_CONFIG_DIR: %w", err)
	}
	return nil
}

// ResolveLaunchProfile determines the effective profile name for a launch.
//
// When profileName is non-empty (the user passed -p explicitly), it is returned
// as-is — explicit user intent always wins. When profileName is empty (bare
// `moai cc`/`glm`/`cg`), the function consults the last-used-profile ledger
// (launch.yaml under GetBaseDir()) and returns the recorded named profile if
// its directory still exists. This lets a bare launch fall back to the most
// recently -p-launched named profile when the default profile is unconfigured.
//
// The fallback is disabled by setting MOAI_NO_PROFILE_FALLBACK=1.
//
// A missing or corrupt launch.yaml, a stale last_profile entry (directory
// absent), or a traversal-shaped name all yield "" (default semantics). The
// function never errors — callers receive a profile name, possibly "".
//
// @MX:NOTE: [AUTO] Last-used-profile fallback resolver — read-only, no error return.
func ResolveLaunchProfile(profileName string) string {
	// Explicit -p wins.
	if profileName != "" {
		return profileName
	}

	// Opt-out env disables the fallback entirely.
	if os.Getenv(config.EnvNoProfileFallback) == "1" {
		return ""
	}

	baseDir := GetBaseDir()
	data, err := os.ReadFile(filepath.Join(baseDir, launchLedgerFile))
	if err != nil {
		return "" // missing ledger — no fallback
	}

	// Decode as a generic map so unknown/legacy keys are tolerated, not
	// rejected. A corrupt file returns "" silently.
	var ledger map[string]any
	if err := yaml.Unmarshal(data, &ledger); err != nil {
		return ""
	}

	last, ok := ledger[lastProfileKey].(string)
	if !ok || last == "" {
		return ""
	}

	// Traverse guard: reject names that could escape the profile base.
	if !isValidProfileName(last) {
		return ""
	}

	// Verify the named profile directory actually exists (stale-record guard).
	info, err := os.Stat(filepath.Join(baseDir, last))
	if err != nil || !info.IsDir() {
		return ""
	}

	return last
}

// RecordLastUsedProfile writes name to launch.yaml as the last_profile entry.
//
// Only NAMED profiles are recorded — "" and "default" are refused with an
// error (they resolve to the base preferences.yaml and carry no useful
// "last used" signal). The write is a read-modify-write that preserves any
// pre-existing keys (including legacy model:/bypass: entries from the
// orphaned launch.yaml) so the file is never destroyed by an update.
//
// The write is atomic (write-temp + os.Rename) so a crash cannot leave a
// half-written ledger. A write failure returns an error the caller logs but
// does not block the launch.
//
// @MX:NOTE: [AUTO] Last-used-profile ledger recorder — atomic read-modify-write.
func RecordLastUsedProfile(name string) error {
	if name == "" || name == "default" {
		return fmt.Errorf("cannot record %q as last-used profile: only named profiles are recorded", name)
	}
	if !isValidProfileName(name) {
		return fmt.Errorf("invalid profile name %q: must not contain path separators or start with '.'", name)
	}

	baseDir := GetBaseDir()
	ledgerPath := filepath.Join(baseDir, launchLedgerFile)

	// Read-modify-write: preserve legacy/unknown keys.
	existing := make(map[string]any)
	if data, err := os.ReadFile(ledgerPath); err == nil {
		_ = yaml.Unmarshal(data, &existing) // corrupt file → start fresh
	}
	existing[lastProfileKey] = name

	out, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshal launch ledger: %w", err)
	}

	// Atomic write: write-temp + os.Rename (same pattern as settings.go).
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("create base dir: %w", err)
	}
	tmp, err := os.CreateTemp(baseDir, ".launch-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(out); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	return os.Rename(tmpName, ledgerPath)
}
