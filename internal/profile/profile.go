package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	//
	// WRITE-ONLY on this binary: the resolution path no longer reads this key.
	// The global read was the source of cross-project profile bleed — a project
	// with no projects[] entry inherited whichever named profile was last
	// -p-launched in ANY project. The write is retained so an older moai binary
	// on the same machine (which still reads last_profile for its global
	// fallback) keeps working; this binary simply no longer reads it.
	lastProfileKey = "last_profile"

	// projectsKey is the YAML key holding the per-project profile memory: a
	// mapping from a normalized absolute project path to the named profile
	// last launched from that project. It is the live source this binary reads
	// at resolution time; lastProfileKey sits alongside it as a write-only
	// legacy/downgrade-compat key.
	projectsKey = "projects"

	// claudeConfigStateFile is the Claude Code account-state file inside a
	// profile directory. Its presence is the sole signal HasClaudeConfig reads;
	// no platform credential store is consulted.
	claudeConfigStateFile = ".claude.json"
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
//
// When CLAUDE_CONFIG_DIR is unset (the common `moai web` case — the cc/glm/cg
// launchers set it only when spawning Claude Code), the function consults the
// last-used-profile ledger (launch.yaml under GetBaseDir()) via
// ResolveLaunchProfile and returns the recorded named profile if its directory
// still exists. This keeps the web console's displayed profile in sync with the
// profile a bare `moai cc` would actually launch. The ledger fallback honors
// MOAI_NO_PROFILE_FALLBACK=1 (disabled) and the stale-record guard (directory
// absent → "" → "default").
func GetCurrentName() string {
	return GetCurrentNameForProject("")
}

// GetCurrentNameForProject is the project-aware form of GetCurrentName.
//
// CLAUDE_CONFIG_DIR still wins: when it is set the ledger is not consulted at
// all, so a `moai web` launched from inside a `moai cc -p X` session keeps
// reporting X. Only when it is unset does the ledger decide, and the decision
// is project-scoped: the projects[projectRoot] entry alone (REQ-PM-024). A
// project with no projects[] entry resolves to "" → "default". The global
// last_profile key is write-only on this binary — it no longer participates
// in resolution, so one project's profile cannot bleed into another.
//
// projectRoot == "" skips the project-scoped lookup entirely (so the wrapper
// resolves "default" when CLAUDE_CONFIG_DIR is unset), which is what makes
// GetCurrentName the project-less wrapper.
func GetCurrentNameForProject(projectRoot string) string {
	configDir := os.Getenv(config.EnvClaudeConfigDir)
	if configDir == "" {
		// Ledger-aware fallback: a bare `moai web` should reflect the last-used
		// named profile, not blindly report "default". Resolution returns ""
		// when the ledger is absent, stale, opted out, or corrupt — in all
		// those cases we degrade to "default" (the original behavior).
		if name := ResolveLaunchProfileForProject(projectRoot, ""); name != "" {
			return name
		}
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
// `moai cc`/`glm`/`cg`), resolution is project-scoped: this wrapper forwards
// to ResolveLaunchProfileForProject with an empty projectRoot, which skips the
// project-scoped lookup and returns "" (default semantics). The global
// last_profile key is write-only on this binary and no longer participates in
// resolution — one project's profile cannot bleed into another via a shared
// global default.
//
// The fallback is disabled by setting MOAI_NO_PROFILE_FALLBACK=1.
//
// A missing or corrupt launch.yaml, a stale project entry (directory absent),
// or a traversal-shaped name all yield "" (default semantics). The function
// never errors — callers receive a profile name, possibly "".
//
// @MX:NOTE: [AUTO] Last-used-profile fallback resolver — read-only, no error return.
func ResolveLaunchProfile(profileName string) string {
	return ResolveLaunchProfileForProject("", profileName)
}

// normalizeProjectKey turns a project root into the canonical ledger key.
//
// The same normalization MUST be applied on both the write and the read side:
// on macOS t.TempDir() and /var paths are symlinks into /private/var, so a key
// written unresolved would never match a lookup done resolved (and the reverse).
// EvalSymlinks is best-effort — a path that no longer exists falls back to a
// lexical Clean. The two branches produce different strings, so a project whose
// directory disappears and reappears can occupy two keys; that is harmless
// (both resolve) and is accepted as limitation L-002.
func normalizeProjectKey(projectRoot string) string {
	if projectRoot == "" {
		return ""
	}
	cleaned := filepath.Clean(projectRoot)
	if resolved, err := filepath.EvalSymlinks(cleaned); err == nil {
		return resolved
	}
	return cleaned
}

// lookupProjectKey finds the ledger key that names the SAME directory as key,
// returning the stored key and whether one was found.
//
// An exact string match wins. When it misses, the map is scanned and each
// candidate is compared with os.SameFile — the filesystem itself decides
// whether two spellings name one directory, rather than this code guessing a
// per-OS case rule.
//
// The case-variant miss is the failure this exists for. On macOS and Windows
// the filesystem is case-insensitive, so `cd /Users/x/moai/repo` and
// `cd /Users/x/MoAI/repo` open the same directory, but os.Getwd reports back
// whichever spelling the shell used and filepath.EvalSymlinks does NOT
// canonicalize case. The two spellings therefore hash to different ledger keys,
// the lookup misses, and the launch silently falls back to the default profile
// — which stores its transcripts elsewhere, so `--continue` / `--resume` find
// no prior session. os.SameFile keeps the lookup correct on case-sensitive
// filesystems too, where two same-spelled-but-differently-cased directories can
// genuinely both exist and must NOT be conflated.
func lookupProjectKey(projects map[string]any, key string) (string, bool) {
	if key == "" || projects == nil {
		return "", false
	}
	if _, ok := projects[key]; ok {
		return key, true
	}

	info, err := os.Stat(key)
	if err != nil {
		return "", false
	}

	// Sorted, not map order. A ledger written before this fix can already hold
	// two spellings of one directory, and Go randomizes map iteration — scanning
	// unordered would hand back either entry from run to run, so a project with
	// duplicates would resolve to a different profile (and a different
	// transcript store) on each launch. Sorting makes the winner the same every
	// time, and the write side collapses the duplicate on the next launch.
	candidates := make([]string, 0, len(projects))
	for candidate := range projects {
		candidates = append(candidates, candidate)
	}
	sort.Strings(candidates)

	var matched string
	var conflict string
	for _, candidate := range candidates {
		candidateInfo, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if !os.SameFile(info, candidateInfo) {
			continue
		}
		if matched == "" {
			matched = candidate
			continue
		}
		// A second entry for the same directory. Sorting already fixed WHICH one
		// wins, but when the two name different profiles the winner is stable and
		// still possibly not the one the user meant — and nothing in the ledger
		// records which was written last. Only the user can say, so say so once
		// rather than routing them somewhere silently.
		if asString(projects[candidate]) != asString(projects[matched]) && conflict == "" {
			conflict = candidate
		}
	}

	if conflict != "" {
		fmt.Fprintf(os.Stderr,
			"Warning: %s records %q twice for one directory, under different profiles (%q vs %q).\n"+
				"Using %q. Remove the stale entry from %s to silence this.\n",
			launchLedgerFile, key,
			asString(projects[matched]), asString(projects[conflict]),
			asString(projects[matched]), filepath.Join(GetBaseDir(), launchLedgerFile))
	}

	return matched, matched != ""
}

// asString reads a ledger value as a string, yielding "" for a non-string entry
// rather than panicking on a hand-edited or downgrade-written ledger.
func asString(v any) string {
	s, _ := v.(string)
	return s
}

// lookupSubtreeProjectKey resolves a directory that has no ledger entry of its
// own to the profile of its deepest REGISTERED ancestor
// (SPEC-STATUSLINE-PROFILE-RESPECT-001 REQ-006..008, kickoff decision D1: the
// walk lives here in the READ path, only after lookupProjectKey's exact/alias
// scan has missed — lookupProjectKey itself is untouched).
//
// The chain of filepath.Dir steps makes the match structural rather than
// lexical: "/proj" is an ancestor of "/proj/worktrees/x" but never of
// "/proj-other", because no Dir step ever produces the latter from the former.
// Walking outward from the session directory, the FIRST registered ancestor is
// by construction the deepest (REQ-007). Each ancestor must exist on disk —
// a registration for a directory that is gone (a deleted worktree) is skipped
// rather than matched, and the walk continues outward to the enclosing live
// project.
//
// A hit whose profile directory has vanished (launchCandidateIsUsable false)
// is likewise skipped, mirroring the exact path's fall-through to "" rather
// than dead-ending on a stale entry.
//
// @MX:NOTE: [AUTO] subtree resolution walk — the read-side twin of
// registeredAncestorKey (t297, REQ-009): the write side now folds subtree
// launches into the registered ancestor this walk resolves to, so the walk
// stays the resolver for subtree directories and its stale-entry skip remains
// the rare fallback (dead rows are reclaimed by PruneStaleProjectEntries).
func lookupSubtreeProjectKey(projects map[string]any, dir, baseDir string) (string, bool) {
	if dir == "" || len(projects) == 0 {
		return "", false
	}
	for cur := filepath.Dir(dir); ; cur = filepath.Dir(cur) {
		if _, ok := projects[cur]; ok {
			if info, err := os.Stat(cur); err == nil && info.IsDir() {
				if name := asString(projects[cur]); launchCandidateIsUsable(baseDir, name) {
					return name, true
				}
			}
		}
		if parent := filepath.Dir(cur); parent == cur {
			return "", false // reached the volume root: no registered ancestor
		}
	}
}

// loadLaunchLedger reads and decodes the ledger as a generic map so unknown and
// legacy keys are tolerated rather than rejected. A missing or corrupt file
// yields an error the callers translate into "no fallback".
func loadLaunchLedger(baseDir string) (map[string]any, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, launchLedgerFile))
	if err != nil {
		return nil, err
	}
	var ledger map[string]any
	if err := yaml.Unmarshal(data, &ledger); err != nil {
		return nil, err
	}
	return ledger, nil
}

// launchCandidateIsUsable reports whether a name recorded in the ledger can be
// launched: it must be traversal-safe and its directory must still exist. A
// candidate that fails either check is skipped so resolution can fall through
// to the next source rather than dead-ending on a stale entry (REQ-PM-008).
func launchCandidateIsUsable(baseDir, name string) bool {
	if name == "" || !isValidProfileName(name) {
		return false
	}
	info, err := os.Stat(filepath.Join(baseDir, name))
	return err == nil && info.IsDir()
}

// ResolveLaunchProfileForProject is the project-aware form of
// ResolveLaunchProfile. Resolution order:
//
//  1. explicit profileName (the user's -p) — always wins
//  2. MOAI_NO_PROFILE_FALLBACK=1 — disables the lookup below
//  3. projects[normalizeProjectKey(projectRoot)] — this project's memory
//  4. "" — default semantics
//
// Step 3 verifies the candidate's directory still exists; a stale project
// entry falls through to "" (default) rather than shadowing it. projectRoot
// == "" skips step 3, which is what makes ResolveLaunchProfile the
// project-less wrapper. The function never errors.
//
// The global last_profile key is WRITE-ONLY on this binary: an older moai
// binary still reads it for downgrade compat, but this binary's resolution
// path no longer does — that global read was the source of cross-project
// profile bleed (one project's profile leaking into another via a shared
// global default).
func ResolveLaunchProfileForProject(projectRoot, profileName string) string {
	// Explicit -p wins.
	if profileName != "" {
		return profileName
	}

	// Opt-out env disables the fallback entirely — project scope included.
	if os.Getenv(config.EnvNoProfileFallback) == "1" {
		return ""
	}

	baseDir := GetBaseDir()
	ledger, err := loadLaunchLedger(baseDir)
	if err != nil {
		return "" // missing or corrupt ledger — no fallback
	}

	if key := normalizeProjectKey(projectRoot); key != "" {
		if projects, ok := ledger[projectsKey].(map[string]any); ok {
			if stored, found := lookupProjectKey(projects, key); found {
				if name, ok := projects[stored].(string); ok && launchCandidateIsUsable(baseDir, name) {
					return name
				}
			} else if name, found := lookupSubtreeProjectKey(projects, key, baseDir); found {
				// The directory has no entry of its own but sits inside a
				// registered project's subtree — a fresh worktree, typically.
				// The walk reuses the already-normalized key so both sides of
				// the comparison went through the same EvalSymlinks spelling.
				return name
			}
		}
	}

	return ""
}

// HasClaudeConfig reports whether the named profile directory already carries
// Claude Code account state.
//
// The decision rests solely on the presence of claudeConfigStateFile inside the
// profile directory; no platform credential store is consulted (REQ-PM-018).
// That is deliberate: the credential carrier differs by platform (macOS keeps
// the token in the OS-level credential store, where it survives a
// CLAUDE_CONFIG_DIR change, while Linux/WSL2 keeps it in a file), but the
// account state that decides whether Claude Code shows the login/onboarding
// screen is per-config-dir on every platform.
//
// Pure predicate: writes nothing, prints nothing. Unnamed and invalid profiles
// report false.
func HasClaudeConfig(name string) bool {
	dir := GetProfileDir(name)
	if dir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, claudeConfigStateFile))
	return err == nil
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
	return RecordLastUsedProfileForProject("", name)
}

// RecordLastUsedProfileForProject is the project-aware form of
// RecordLastUsedProfile. It writes the name to BOTH the project-scoped entry
// under projects: and the global last_profile key.
//
// The projects: entry is the live source this binary reads at resolution time
// (ResolveLaunchProfileForProject step 3). The last_profile write is
// LEGACY/DOWNGRADE COMPAT only: an older moai binary on the same machine still
// reads last_profile for its global fallback, so writing it keeps that older
// binary's "most recent launch wins" behavior intact. This binary never reads
// last_profile for resolution — keeping the write is pure interop hygiene, not
// a live signal for this code.
//
// Beyond the name-shape checks the original carried, the profile DIRECTORY must
// already exist (REQ-PM-011). Without that guard a name that never resolves to
// anything persists in the ledger forever with no signal to the user — the read
// side silently skips it. The launcher therefore records only after it has
// created the directory (see unifiedLaunchDefault step 4.5 → step 5).
//
// The projects: entry is normalized at write time (REQ-009, card t297): a
// launch recorded from inside a registered project's subtree — a worktree,
// typically — updates the registered root's row instead of adding one, so
// creating worktrees no longer grows the ledger. See resolveWriteProjectKey
// for the exact precedence.
//
// projectRoot == "" leaves the projects map untouched and writes last_profile
// only, which is what makes RecordLastUsedProfile a behavior-preserving wrapper.
func RecordLastUsedProfileForProject(projectRoot, name string) error {
	if name == "" || name == "default" {
		return fmt.Errorf("cannot record %q as last-used profile: only named profiles are recorded", name)
	}
	if !isValidProfileName(name) {
		return fmt.Errorf("invalid profile name %q: must not contain path separators or start with '.'", name)
	}

	baseDir := GetBaseDir()
	ledgerPath := filepath.Join(baseDir, launchLedgerFile)

	// Directory-existence guard (REQ-PM-011): refuse ghost names outright
	// rather than letting them accumulate in the ledger.
	if info, err := os.Stat(filepath.Join(baseDir, name)); err != nil || !info.IsDir() {
		return fmt.Errorf("cannot record profile %q: its directory does not exist under %s", name, baseDir)
	}

	// Read-modify-write: preserve legacy/unknown keys.
	existing := make(map[string]any)
	if data, err := os.ReadFile(ledgerPath); err == nil {
		_ = yaml.Unmarshal(data, &existing) // corrupt file → start fresh
	}
	existing[lastProfileKey] = name

	if key := normalizeProjectKey(projectRoot); key != "" {
		projects, ok := existing[projectsKey].(map[string]any)
		if !ok {
			projects = make(map[string]any)
		}
		projects[resolveWriteProjectKey(projects, key)] = name
		existing[projectsKey] = projects
	}

	return saveLaunchLedger(baseDir, existing)
}

// resolveWriteProjectKey returns the ledger key a launch recorded from the
// normalized directory key should be written under (REQ-009, card t297).
//
// Precedence mirrors the read side so a write and a later lookup agree:
//
//  1. An exact or alias entry for this directory wins — the path is a
//     registered project in its own right, and its row is updated in place.
//     This is what keeps nested INDEPENDENT projects (/mono and /mono/lib,
//     both registered) from folding into one another, and what keeps a legacy
//     row written by a pre-normalization binary self-consistent.
//  2. Otherwise, the deepest registered ancestor wins (registeredAncestorKey):
//     the launch came from a subtree of a registered project — a worktree —
//     and folding it into the root's row is what stops the ledger growing one
//     row per worktree.
//  3. Otherwise the normalized key itself: with nothing registered above it,
//     the row lives at the path it names. It stays live while the directory
//     does and is reclaimed with it by PruneStaleProjectEntries.
//
// @MX:NOTE: [AUTO] write-key precedence for subtree launches — exact/alias,
// then registered ancestor (fold), then the path itself (t297, REQ-009).
func resolveWriteProjectKey(projects map[string]any, key string) string {
	if key == "" {
		return ""
	}
	if stored, found := lookupProjectKey(projects, key); found {
		return stored
	}
	if ancestor, found := registeredAncestorKey(projects, key); found {
		return ancestor
	}
	return key
}

// registeredAncestorKey walks key's real path-segment ancestors outward and
// returns the first one that names a REGISTERED project whose directory still
// exists — the structural twin of the read-side lookupSubtreeProjectKey walk,
// so a key folded on write is exactly the key a later subtree lookup resolves
// to. The Dir-step chain makes the match structural (a path-segment
// boundary), never lexical.
//
// Deliberate asymmetry with the read side: the walk does NOT consult
// launchCandidateIsUsable. The read side must skip an ancestor whose recorded
// profile has vanished; the write side is about to OVERWRITE the value with a
// profile the recorder has just verified (the directory-existence guard in
// RecordLastUsedProfileForProject), so folding into such an ancestor repairs
// the row rather than preserving a stale value.
func registeredAncestorKey(projects map[string]any, key string) (string, bool) {
	if key == "" || len(projects) == 0 {
		return "", false
	}
	for cur := filepath.Dir(key); ; cur = filepath.Dir(cur) {
		if _, ok := projects[cur]; ok {
			if info, err := os.Stat(cur); err == nil && info.IsDir() {
				return cur, true
			}
		}
		if parent := filepath.Dir(cur); parent == cur {
			return "", false // reached the volume root: no registered ancestor
		}
	}
}

// saveLaunchLedger marshals and atomically writes the ledger (write-temp +
// os.Rename, the same pattern as settings.go) so a crash cannot leave a
// half-written file. The caller owns the map's contents; unknown and legacy
// keys survive because they were loaded into it.
func saveLaunchLedger(baseDir string, ledger map[string]any) error {
	out, err := yaml.Marshal(ledger)
	if err != nil {
		return fmt.Errorf("marshal launch ledger: %w", err)
	}

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

	return os.Rename(tmpName, filepath.Join(baseDir, launchLedgerFile))
}

// PruneStaleProjectEntries removes projects[] entries whose directory no
// longer exists on disk, and returns the removed keys in sorted order.
//
// A dead key is unrecoverable by every read path — lookupProjectKey requires
// os.Stat, the subtree walk requires Stat + IsDir — so removing it is
// semantics-preserving by construction, which is what makes the prune
// idempotent and safe to run at any worktree disposal: the second run finds
// nothing to remove. Keys naming live directories are never touched, and the
// non-projects keys (last_profile, model, bypass, ...) survive untouched.
// Only the key's own directory is judged: an entry whose PROFILE directory
// vanished is kept — the binding may be wanted again, and the read side
// already falls through it.
//
// This is the reclamation half of the ledger's lifecycle (card t297): the
// write side stopped creating dead rows (resolveWriteProjectKey), and this
// removes the ones that predate it and the cold-start rows whose worktree has
// since been disposed. Callers report the returned count; a missing ledger is
// a no-op (nil, nil), a corrupt one is an error.
//
// @MX:ANCHOR: [AUTO] dead-row reclamation for the launch ledger's projects map
// @MX:REASON: wired at every moai worktree disposal path (remove/done/clean
// sweeps); weakening the dead-directory predicate to also drop entries whose
// PROFILE is missing would silently erase wanted project bindings, and
// skipping the sorted delete-order changes which keys a partially-failed
// save leaves behind.
func PruneStaleProjectEntries() ([]string, error) {
	baseDir := GetBaseDir()
	ledger, err := loadLaunchLedger(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no ledger yet — nothing to reclaim
		}
		return nil, fmt.Errorf("load launch ledger: %w", err)
	}
	projects, ok := ledger[projectsKey].(map[string]any)
	if !ok || len(projects) == 0 {
		return nil, nil
	}

	var pruned []string
	for key := range projects {
		if info, err := os.Stat(key); err == nil && info.IsDir() {
			continue
		}
		pruned = append(pruned, key)
	}
	if len(pruned) == 0 {
		return nil, nil // nothing to write — the file is left byte-identical
	}
	sort.Strings(pruned)
	for _, key := range pruned {
		delete(projects, key)
	}
	if err := saveLaunchLedger(baseDir, ledger); err != nil {
		return nil, fmt.Errorf("save pruned launch ledger: %w", err)
	}
	return pruned, nil
}
