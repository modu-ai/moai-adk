// Package worktreeroot resolves the home-level worktree root for a
// repository (SPEC-WORKTREE-HOME-ROOT-001 M1).
//
// The path form is literally
//
//	<MoaiHome>/worktrees/projects/<project-key>/<name>
//
// where <MoaiHome> comes from paths.MoaiHome (MOAI_HOME when absolute,
// otherwise ~/.moai) and <project-key> is assigned once per repository and
// never changes (REQ-WHR-002). Every accessor accepts any directory of the
// repository — the primary checkout or one of its linked worktrees: the key
// is derived from the primary root, never from the invoking worktree.
//
// Assignment is recorded in two places that travel with their owner: the
// owner marker <MoaiHome>/worktrees/projects/<key>/.moai-project.json under
// the home root, and a record file in the repository's git common directory,
// which moves with the checkout and survives a rename of its directory.
// Resolution reads the record first and only derives a key when none is
// recorded, so re-running resolution is idempotent and a later same-name
// arrival cannot displace the first (design.md, Stability).
//
// The package keeps the config Loader out of its dependency graph, so
// config consumers can share these paths without an import cycle.
package worktreeroot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/paths"
	"gopkg.in/yaml.v3"
)

// RootKind names which root a candidate tree was found under.
type RootKind string

// The three roots Candidates searches. No root is preferred: the caller
// reports ambiguity when it receives two or more (REQ-WHR-022).
const (
	KindHome         RootKind = "home"          // <MoaiHome>/worktrees/projects/<key>
	KindLegacyMoai   RootKind = "legacy-moai"   // <primary>/.moai/worktrees
	KindLegacyClaude RootKind = "legacy-claude" // <primary>/.claude/worktrees
)

// Candidate is one tree of a given name found under one root.
type Candidate struct {
	Path string
	Kind RootKind
}

// Resolution is the full result of resolving a repository: the primary root
// the key was derived from, the assigned key, the project folder under the
// home root, and a notice for an operator pin that differs from the recorded
// key (REQ-WHR-004).
type Resolution struct {
	PrimaryRoot string
	Key         string
	Root        string
	Notice      string
}

// Package-level path pieces under the home root and the primary checkout.
var (
	projectsSubdir  = filepath.Join("worktrees", "projects")                            // under MoaiHome
	gitStrategyYAML = filepath.Join(".moai", "config", "sections", "git-strategy.yaml") // under the primary root
)

const (
	markerName = ".moai-project.json"
	recordName = "moai-project-key"
)

// ownerMarker is the JSON document at <projects>/<key>/.moai-project.json:
// schema version, owner identity, primary root last seen and creation time
// (AC-WHR-002a).
type ownerMarker struct {
	Schema      int    `json:"schema"`
	Owner       string `json:"owner"`
	PrimaryRoot string `json:"primary_root"`
	CreatedAt   string `json:"created_at"`
}

// @MX:ANCHOR: [AUTO] resolver entry SSOT — every creation path and reader resolves the home root here
// @MX:REASON: [AUTO] M2 materializer, M3 readers and M4 list all consume it; a second key style would split one repository across two project folders (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-002)
func Resolve(repoDir string) (Resolution, error) {
	return resolve(repoDir)
}

// Key returns the repository's assigned project key, assigning it on first
// use (REQ-WHR-002).
func Key(repoDir string) (string, error) {
	res, err := resolve(repoDir)
	if err != nil {
		return "", err
	}
	return res.Key, nil
}

// Root returns the repository's project folder under the home root,
// assigning the key on first use. The directory itself is created by the
// assignment; the per-tree subdirectories are not (REQ-WHR-001).
func Root(repoDir string) (string, error) {
	res, err := resolve(repoDir)
	if err != nil {
		return "", err
	}
	return res.Root, nil
}

// PathFor returns <project root>/<name> without creating anything; the tree
// directory is left to the creation paths (REQ-WHR-001).
func PathFor(repoDir, name string) (string, error) {
	res, err := resolve(repoDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(res.Root, name), nil
}

// Plan answers the landing path of name for repoDir WITHOUT writing
// anything: no project folder, no owner marker, no repository-side record
// (card t1640 review P2i — the migration's dry run resolves destinations
// for its plan and must leave the repository and MOAI_HOME byte-identical,
// AC-WHR-015). The recorded key wins exactly as resolve() does; a
// first-time assignment computes the key the writing resolver WOULD claim —
// the folded candidate, or its AC-WHR-003a collision suffix — without
// claiming it, so a plan naming a suffixed key is a plan, not a promise:
// the real claim may suffix differently if the home root changed between
// the plan and the apply. An existing folder owned by another repository
// is an error here, the same answer the writing claim gives.
func Plan(repoDir, name string) (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	common, err := gitCommonDir(repoDir)
	if err != nil {
		return "", err
	}
	primary := filepath.Dir(common)
	canonical, err := filepath.EvalSymlinks(primary)
	if err != nil {
		return "", fmt.Errorf("resolving primary root %s: %w", primary, err)
	}
	projects := filepath.Join(home, projectsSubdir)

	// The recorded key wins over everything (design.md, Stability) — read
	// only: ensureProjectDir is resolve()'s write and is skipped here.
	if k := readRecord(common); k != "" {
		return filepath.Join(projectDir(home, k), name), nil
	}

	pin, err := readPin(primary)
	if err != nil {
		return "", err
	}
	candidate := sanitize(filepath.Base(canonical))
	if pin != "" {
		candidate = pin
	}
	if folder := foldFind(projects, candidate); folder == "" {
		return filepath.Join(projectDir(home, candidate), name), nil
	} else if owner, oerr := readMarkerOwner(filepath.Join(projects, folder, markerName)); oerr != nil {
		return "", fmt.Errorf("project folder %q already exists and its owner marker %s cannot be read: %w", folder, filepath.Join(projects, folder, markerName), oerr)
	} else if strings.EqualFold(owner, canonical) {
		return filepath.Join(projectDir(home, folder), name), nil
	} else if pin != "" {
		return "", fmt.Errorf("git_strategy.worktree_project_key pin %q names project folder %q owned by another repository (owner marker %s); refusing without suffixing", candidate, folder, filepath.Join(projects, folder, markerName))
	} else {
		suffixed := candidate + "-" + suffixHash(canonical)
		if folder2 := foldFind(projects, suffixed); folder2 == "" {
			return filepath.Join(projectDir(home, suffixed), name), nil
		} else if owner2, oerr := readMarkerOwner(filepath.Join(projects, folder2, markerName)); oerr != nil {
			return "", fmt.Errorf("project folder %q already exists and its owner marker %s cannot be read: %w", folder2, filepath.Join(projects, folder2, markerName), oerr)
		} else if !strings.EqualFold(owner2, canonical) {
			return "", fmt.Errorf("project folder %q is already owned by another repository (owner marker %s); refusing", folder2, filepath.Join(projects, folder2, markerName))
		} else {
			return filepath.Join(projectDir(home, folder2), name), nil
		}
	}
}

// RecordedKey returns the repository-side key record of repoDir — a
// read-only observation that writes nothing and reports "" when the
// repository has no recorded assignment (card t1640 review P1-5: a reader
// that scans the whole home root cannot tell this repository's trees from
// another repository's; the recorded key scopes the scan).
func RecordedKey(repoDir string) string {
	common, err := gitCommonDir(repoDir)
	if err != nil {
		return ""
	}
	return readRecord(common)
}

// resolve runs the full resolution: primary-root discovery, pin read and
// validation, record-first key lookup, first-time assignment with the owner
// marker and the repository-side record, and the collision suffix.
func resolve(repoDir string) (Resolution, error) {
	common, err := gitCommonDir(repoDir)
	if err != nil {
		return Resolution{}, err
	}
	primary := filepath.Dir(common)
	canonical, err := filepath.EvalSymlinks(primary)
	if err != nil {
		return Resolution{}, fmt.Errorf("resolving primary root %s: %w", primary, err)
	}
	home, err := paths.MoaiHome()
	if err != nil {
		return Resolution{}, err
	}

	pin, err := readPin(primary)
	if err != nil {
		return Resolution{}, err
	}
	if pin != "" {
		if err := validatePin(pin); err != nil {
			return Resolution{}, err
		}
	}

	// The recorded key wins over everything (design.md, Stability); a
	// differing valid pin is never applied, only noticed (REQ-WHR-004). The
	// project folder under the CURRENT home root is ensured here too: a
	// relocated MOAI_HOME must still hold the owner marker for the recorded
	// key (REQ-WHR-002a), while an existing marker is left byte-untouched.
	if k := readRecord(common); k != "" {
		if err := ensureProjectDir(home, k, canonical); err != nil {
			return Resolution{}, err
		}
		res := Resolution{PrimaryRoot: canonical, Key: k, Root: projectDir(home, k)}
		if pin != "" && pin != k {
			res.Notice = fmt.Sprintf("git_strategy.worktree_project_key pin %q ignored; the recorded key %q is kept", pin, k)
		}
		return res, nil
	}

	base := sanitize(filepath.Base(canonical))
	candidate := base
	if pin != "" {
		candidate = pin
	}
	key, err := claim(home, candidate, canonical, pin != "", base)
	if err != nil {
		return Resolution{}, err
	}
	if err := writeRecord(common, key); err != nil {
		return Resolution{}, err
	}
	return Resolution{PrimaryRoot: canonical, Key: key, Root: projectDir(home, key)}, nil
}

// gitCommonDir returns the repository's git common directory as an absolute
// path — <primary>/.git — derived from any directory of the repository,
// linked worktree included (design.md D-11 uses the same call).
func gitCommonDir(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolving the repository of %s: %w", repoDir, err)
	}
	common := strings.TrimSpace(string(out))
	if common == "" {
		return "", fmt.Errorf("git rev-parse --git-common-dir printed nothing for %s", repoDir)
	}
	return common, nil
}

func projectDir(home, key string) string {
	return filepath.Join(home, projectsSubdir, key)
}

// sanitize maps a directory name onto one safe path segment: every character
// outside [A-Za-z0-9._-] becomes "-" (AC-WHR-002a: "my proj (v2)" assigns
// "my-proj--v2-").
func sanitize(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}

// validatePin rejects a pin that is not one safe path segment (AC-WHR-004).
func validatePin(pin string) error {
	if pin == "." || pin == ".." || strings.ContainsAny(pin, `/\`) {
		return fmt.Errorf("git_strategy.worktree_project_key value %q is not one safe path segment", pin)
	}
	return nil
}

// claim takes candidate as this repository's project folder: fresh folders
// are created with their owner marker; a folder owned by another repository
// sends a derived key to the hash-suffixed name (REQ-WHR-003a) and a pinned
// key to a refusal without suffixing (REQ-WHR-004); an unreadable owner
// marker refuses creation, naming the marker path (REQ-WHR-003b).
func claim(home, candidate, canonical string, pinned bool, base string) (_ string, err error) {
	release, err := lockProjectClaims(home)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, release()) }()
	projects := filepath.Join(home, projectsSubdir)

	folder := foldFind(projects, candidate)
	if folder == "" {
		if err := writeMarker(filepath.Join(projects, candidate), canonical); err != nil {
			return "", err
		}
		return candidate, nil
	}

	markerPath := filepath.Join(projects, folder, markerName)
	owner, err := readMarkerOwner(markerPath)
	if err != nil {
		// A folder with NO marker is unowned, not refused: claim it.
		// AC-WHR-003b refuses a marker that exists but cannot be read or
		// parsed, never an absent one.
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("project folder %q already exists and its owner marker %s cannot be read: %w", folder, markerPath, err)
		}
		if werr := writeMarker(filepath.Join(projects, folder), canonical); werr != nil {
			return "", werr
		}
		return folder, nil
	}
	if strings.EqualFold(owner, canonical) {
		return folder, nil
	}
	if pinned {
		return "", fmt.Errorf("git_strategy.worktree_project_key pin %q names project folder %q owned by another repository (owner marker %s); refusing without suffixing", candidate, folder, markerPath)
	}

	suffixed := base + "-" + suffixHash(canonical)
	folder2 := foldFind(projects, suffixed)
	if folder2 == "" {
		if err := writeMarker(filepath.Join(projects, suffixed), canonical); err != nil {
			return "", err
		}
		return suffixed, nil
	}
	markerPath2 := filepath.Join(projects, folder2, markerName)
	owner2, err := readMarkerOwner(markerPath2)
	if err != nil {
		return "", fmt.Errorf("project folder %q already exists and its owner marker %s cannot be read: %w", folder2, markerPath2, err)
	}
	if !strings.EqualFold(owner2, canonical) {
		return "", fmt.Errorf("project folder %q is already owned by another repository (owner marker %s); refusing", folder2, markerPath2)
	}
	return folder2, nil
}

// suffixHash is the 8-hex prefix of SHA-256 over the case-folded canonical
// primary root — the disambiguator of REQ-WHR-003a.
func suffixHash(canonical string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(canonical)))
	return hex.EncodeToString(sum[:])[:8]
}

// foldFind returns the name of the entry of dir that equals name
// case-insensitively (the folder-name test of AC-WHR-003a is
// case-insensitive on every volume), or "" when none matches.
func foldFind(dir, name string) string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return e.Name()
		}
	}
	return ""
}

func writeMarker(dir, canonical string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	m := ownerMarker{
		Schema:      1,
		Owner:       canonical,
		PrimaryRoot: canonical,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, markerName), append(body, '\n'), 0o644)
}

// readMarkerOwner parses an existing owner marker and returns its owner
// identity; a marker that exists but cannot be read or parsed is an error
// (the caller refuses creation, REQ-WHR-003b).
func readMarkerOwner(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return "", err
	}
	owner, _ := m["owner"].(string)
	return owner, nil
}

// readPin returns the git_strategy.worktree_project_key value configured in
// the primary checkout, or "" when the file, the section or the key is
// absent. The narrow read keeps the package import-light (the Loader-based
// reader for config-package consumers lives in internal/config); a missing
// or non-regular file yields the neutral "".
func readPin(primary string) (string, error) {
	path := filepath.Join(primary, gitStrategyYAML)
	fi, err := os.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return "", nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	var document struct {
		GitStrategy struct {
			WorktreeProjectKey string `yaml:"worktree_project_key"`
		} `yaml:"git_strategy"`
	}
	if err := yaml.Unmarshal(body, &document); err != nil {
		return "", fmt.Errorf("read worktree project pin in %s: %w", path, err)
	}
	return strings.TrimSpace(document.GitStrategy.WorktreeProjectKey), nil
}

// ensureProjectDir creates the project folder and its owner marker under the
// home root when they are absent, leaving an existing marker untouched.
func ensureProjectDir(home, key, canonical string) (err error) {
	release, err := lockProjectClaims(home)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, release()) }()
	dir := projectDir(home, key)
	markerPath := filepath.Join(dir, markerName)
	owner, err := readMarkerOwner(markerPath)
	if err == nil {
		if !strings.EqualFold(owner, canonical) && !Same(owner, canonical) {
			// The repository-side record travels with a renamed checkout;
			// its historical marker remains byte-untouched. An owner root
			// that still exists, however, is another repository's claim.
			if _, statErr := os.Stat(owner); !os.IsNotExist(statErr) {
				if statErr != nil {
					return fmt.Errorf("inspect recorded marker owner %s: %w", owner, statErr)
				}
				return fmt.Errorf("recorded project key %q names a folder owned by another repository (owner marker %s)", key, markerPath)
			}
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("read owner marker %s: %w", markerPath, err)
	}
	return writeMarker(dir, canonical)
}

// The case-folded folder search and marker creation form one claim across
// repositories and processes. The shared claim primitive never reclaims a
// live owner and keeps both acquisition and stale-owner inspection bounded.
func lockProjectClaims(home string) (func() error, error) {
	dir := filepath.Dir(filepath.Join(home, projectsSubdir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return atomicfile.ClaimSection(context.Background(), filepath.Join(dir, ".project-key.lock"), 0o644, 100, 10*time.Millisecond)
}

// readRecord returns the key recorded in the repository's git common
// directory, or "" when none is recorded.
func readRecord(common string) string {
	body, err := os.ReadFile(filepath.Join(common, recordName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}

func writeRecord(common, key string) error {
	return os.WriteFile(filepath.Join(common, recordName), []byte(key+"\n"), 0o644)
}

// Candidates returns every distinct tree of name found under the home
// project root and both legacy roots, with no root preferred and no
// ambiguity judgment (REQ-WHR-022). One directory reached under two
// spellings is one tree, not an ambiguity (REQ-WHR-023).
func Candidates(repoDir, name string) ([]Candidate, error) {
	res, err := resolve(repoDir)
	if err != nil {
		return nil, err
	}
	roots := []Candidate{
		{filepath.Join(res.Root, name), KindHome},
		{filepath.Join(res.PrimaryRoot, ".moai", "worktrees", name), KindLegacyMoai},
		{filepath.Join(res.PrimaryRoot, ".claude", "worktrees", name), KindLegacyClaude},
	}
	var out []Candidate
	var seen []string
	for _, c := range roots {
		fi, err := os.Stat(c.Path)
		if err != nil || !fi.IsDir() {
			continue
		}
		real, err := filepath.EvalSymlinks(c.Path)
		if err != nil {
			real = filepath.Clean(c.Path)
		}
		dup := false
		for _, s := range seen {
			if Same(real, s) {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		seen = append(seen, real)
		out = append(out, c)
	}
	return out, nil
}

// OwnerMarkerOf returns the owner identity recorded in projectDir's owner
// marker, with ok=false when the marker is absent or unreadable — the
// classifier contexts' data source for the home-root marker witness
// (SPEC-WORKTREE-HOME-ROOT-001 M3; the marker format itself stays owned by
// this package).
func OwnerMarkerOf(projectDir string) (string, bool) {
	owner, err := readMarkerOwner(filepath.Join(projectDir, markerName))
	if err != nil {
		return "", false
	}
	return owner, true
}

// Same reports whether two path spellings denote one directory:
// spelling-safe equality (REQ-WHR-023). Existing paths compare by resolved
// identity (symlinks, case-aliased directories on case-insensitive volumes);
// paths that do not exist compare by their cleaned spelling.
func Same(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA == nil && errB == nil {
		if ra == rb {
			return true
		}
		sa, errA2 := os.Stat(ra)
		sb, errB2 := os.Stat(rb)
		return errA2 == nil && errB2 == nil && os.SameFile(sa, sb)
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
