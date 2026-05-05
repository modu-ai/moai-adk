// SPEC-DOCS-SITE-001 Phase 5, REQ-DS-17/18
//
// At release time, copies docs-site/content/{locale}/ content of the previous version
// to content/{locale}/v<previous-minor>/ folder
//
//
//	docs-version-snapshot \
//	  --current-version v2.13.0 \
//	  --previous-version v2.12.4 \
//	  --content-dir docs-site/content \
//	  [--dry-run]
//
//   - Patch release (v2.12.1 → v2.12.2): no snapshot, exit 0
//   - Minor release (v2.12.x → v2.13.0): content/{locale}/v2.12/ creation
//   - Major release (v2.x.y → v3.0.0): content/{locale}/v2/ creation

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// locales supported by the docs site
var locales = []string{"ko", "en", "ja", "zh"}

// semver holds a parsed semantic version
type semver struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

// parseSemver parses a version string like "v2.13.0" or "2.13.0"
func parseSemver(raw string) (semver, error) {
	s := strings.TrimPrefix(raw, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("invalid semver %q: expected MAJOR.MINOR.PATCH", raw)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, fmt.Errorf("invalid major in %q: %w", raw, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, fmt.Errorf("invalid minor in %q: %w", raw, err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, fmt.Errorf("invalid patch in %q: %w", raw, err)
	}
	return semver{Major: major, Minor: minor, Patch: patch, Raw: raw}, nil
}

// releaseType represents the kind of version bump
type releaseType int

const (
	patchRelease releaseType = iota
	minorRelease
	majorRelease
)

// classifyRelease compares two versions and returns the release type
//
// @MX:ANCHOR: [AUTO] Version classification — called by main + all unit tests
// @MX:REASON: [AUTO] Core decision gate; wrong classification silently skips or incorrectly creates snapshots
func classifyRelease(prev, curr semver) (releaseType, error) {
	if curr.Major > prev.Major {
		return majorRelease, nil
	}
	if curr.Major < prev.Major {
		return patchRelease, fmt.Errorf("current version %s is older than previous %s", curr.Raw, prev.Raw)
	}
	// Same major
	if curr.Minor > prev.Minor {
		return minorRelease, nil
	}
	if curr.Minor < prev.Minor {
		return patchRelease, fmt.Errorf("current version %s is older than previous %s (minor)", curr.Raw, prev.Raw)
	}
	// Same major + minor -> patch
	return patchRelease, nil
}

// snapshotDirName returns the version folder name to create
// Minor release:  "v2.12"  (from previous v2.12.x)
// Major release:  "v2.12"  (from previous v2.12.0 — last minor of the old major)
//
// Both Minor and Major releases snapshot the previous Minor version (MAJOR.MINOR),
// because that is the most precise stable identifier for the docs snapshot
// Per AC-G3-04: v2.12.0 -> v3.0.0 must produce content/{locale}/v2.12/
func snapshotDirName(prev semver, rt releaseType) string {
	switch rt {
	case minorRelease, majorRelease:
		return fmt.Sprintf("v%d.%d", prev.Major, prev.Minor)
	default:
		return ""
	}
}

// Config holds the parsed CLI flags
type Config struct {
	CurrentVersion  string
	PreviousVersion string
	ContentDir      string
	DryRun          bool
	Force           bool
}

// @MX:WARN: [AUTO] Calls os/exec (git archive) — subprocess execution
// @MX:REASON: [AUTO] git archive extracts previous tag content; failure must be detected early before any copy starts

// extractTagContent uses "git show <tag>:<path>" to read file content from a git tag
// Returns ErrTagNotFound if the tag does not exist in the repository
var errTagNotFound = errors.New("git tag not found")

func extractTagContent(tag, gitPath string) ([]byte, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", tag, gitPath))
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("%w: %s", errTagNotFound, tag)
		}
		return nil, fmt.Errorf("git show %s:%s: %w", tag, gitPath, err)
	}
	return out, nil
}

// tagExists checks whether a given tag is present in the repo
func tagExists(tag string) bool {
	cmd := exec.Command("git", "tag", "-l", tag)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == tag
}

// listTagFiles lists files under <tag>:contentDir/{locale}/ using "git ls-tree"
func listTagFiles(tag, contentDir, locale string) ([]string, error) {
	treePath := fmt.Sprintf("%s/%s", contentDir, locale)
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", tag, treePath)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-tree %s %s: %w", tag, treePath, err)
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// copyFromTag copies all files under contentDir/{locale}/ at the given tag
// into destDir (which is contentDir/{locale}/vX.Y/)
func copyFromTag(tag, contentDir, locale, destDir string, dryRun bool) (int, error) {
	srcPrefix := fmt.Sprintf("%s/%s/", contentDir, locale)
	files, err := listTagFiles(tag, contentDir, locale)
	if err != nil {
		return 0, err
	}

	// Filter out existing v*/ sub-directories to avoid nested snapshots
	var filtered []string
	for _, f := range files {
		rel := strings.TrimPrefix(f, srcPrefix)
		// Skip files that start with "v" followed by a digit (versioned sub-dirs)
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) > 0 && len(parts[0]) >= 2 && parts[0][0] == 'v' && parts[0][1] >= '0' && parts[0][1] <= '9' {
			continue
		}
		filtered = append(filtered, f)
	}

	if len(filtered) == 0 {
		return 0, fmt.Errorf("no files found under %s/%s at tag %s", contentDir, locale, tag)
	}

	count := 0
	for _, gitPath := range filtered {
		rel := strings.TrimPrefix(gitPath, srcPrefix)
		destFile := filepath.Join(destDir, filepath.FromSlash(rel))

		if dryRun {
			fmt.Printf("[dry-run] copy %s -> %s\n", gitPath, destFile)
			count++
			continue
		}

		content, err := extractTagContent(tag, gitPath)
		if err != nil {
			return count, fmt.Errorf("read %s: %w", gitPath, err)
		}

		if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
			return count, fmt.Errorf("mkdir %s: %w", filepath.Dir(destFile), err)
		}
		if err := os.WriteFile(destFile, content, 0o644); err != nil {
			return count, fmt.Errorf("write %s: %w", destFile, err)
		}
		fmt.Printf("copied %s -> %s\n", gitPath, destFile)
		count++
	}
	return count, nil
}

// rollback removes all directories created during a failed snapshot
func rollback(dirs []string) {
	for _, d := range dirs {
		fmt.Fprintf(os.Stderr, "[rollback] removing %s\n", d)
		_ = os.RemoveAll(d)
	}
}

// run executes the snapshot logic. Returns an error if any step fails
func run(cfg Config) error {
	prev, err := parseSemver(cfg.PreviousVersion)
	if err != nil {
		return fmt.Errorf("parse previous-version: %w", err)
	}
	curr, err := parseSemver(cfg.CurrentVersion)
	if err != nil {
		return fmt.Errorf("parse current-version: %w", err)
	}

	rt, err := classifyRelease(prev, curr)
	if err != nil {
		return err
	}

	if rt == patchRelease {
		fmt.Printf("Patch release detected (%s -> %s). No snapshot needed.\n",
			cfg.PreviousVersion, cfg.CurrentVersion)
		return nil
	}

	snapName := snapshotDirName(prev, rt)
	fmt.Printf("%s release detected (%s -> %s). Creating snapshot: %s\n",
		map[releaseType]string{minorRelease: "Minor", majorRelease: "Major"}[rt],
		cfg.PreviousVersion, cfg.CurrentVersion, snapName)

	// Validate the previous tag exists in git
	if !cfg.DryRun {
		if !tagExists(cfg.PreviousVersion) {
			return fmt.Errorf("%w: %s", errTagNotFound, cfg.PreviousVersion)
		}
	}

	var createdDirs []string

	for _, locale := range locales {
		destDir := filepath.Join(cfg.ContentDir, locale, snapName)

		// Guard: do not overwrite existing snapshot
		if _, err := os.Stat(destDir); err == nil {
			if !cfg.Force {
				return fmt.Errorf("snapshot directory already exists: %s (use --force to overwrite)", destDir)
			}
			fmt.Printf("[force] removing existing %s\n", destDir)
			if !cfg.DryRun {
				if err := os.RemoveAll(destDir); err != nil {
					return fmt.Errorf("remove existing %s: %w", destDir, err)
				}
			}
		}

		if !cfg.DryRun {
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				rollback(createdDirs)
				return fmt.Errorf("mkdir %s: %w", destDir, err)
			}
			createdDirs = append(createdDirs, destDir)
		}

		count, err := copyFromTag(cfg.PreviousVersion, cfg.ContentDir, locale, destDir, cfg.DryRun)
		if err != nil {
			if !cfg.DryRun {
				rollback(createdDirs)
			}
			return fmt.Errorf("copy locale %s: %w", locale, err)
		}
		fmt.Printf("locale %s: %d files %s\n", locale, count,
			map[bool]string{true: "would be copied (dry-run)", false: "copied"}[cfg.DryRun])
	}

	fmt.Println("Snapshot complete.")
	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.CurrentVersion, "current-version", "", "Current release version (e.g. v2.13.0)")
	flag.StringVar(&cfg.PreviousVersion, "previous-version", "", "Previous release version (e.g. v2.12.4)")
	flag.StringVar(&cfg.ContentDir, "content-dir", "docs-site/content", "Path to docs-site/content directory")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Print actions without executing them")
	flag.BoolVar(&cfg.Force, "force", false, "Overwrite existing snapshot directory")
	flag.Parse()

	if cfg.CurrentVersion == "" || cfg.PreviousVersion == "" {
		fmt.Fprintln(os.Stderr, "Error: --current-version and --previous-version are required")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
