package cli

// SPEC-V3R6-MOAI-CLEAN-HOME-001: the ~/.moai cleanup core shared by
// `moai clean --home` (M2 deletion path) and the doctor Home Disk Usage check
// (cleanable-bytes estimate). The scanner is allowlist-only (REQ-MCH-004) and
// the carve-out predicate below is the single guard both consumers share
// (REQ-MCH-005: the guard predicate and the scanner share one implementation).

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// carveOutDirNames are directory-segment names that are never deletable, at
// any depth — root level and per-profile under claude-profiles/<p>/ alike
// (REQ-MCH-005). plugins/ is listed here as well as in the doctor
// duplicate-cluster detection: reported, never touched (REQ-MCH-006).
var carveOutDirNames = map[string]bool{
	"projects":  true,
	"config":    true,
	"state":     true,
	"worktrees": true,
	"mcp":       true,
	"bin":       true,
	"search":    true,
	"studio":    true,
	"plugins":   true,
}

// carveOutFileNames are exact file names that are never deletable.
var carveOutFileNames = map[string]bool{
	"launch.yaml":      true,
	"preferences.yaml": true,
}

// isCarvedOut reports whether relPath — slash-separated, relative to the
// ~/.moai root — touches ANY carved-out segment at ANY depth (root level and
// per-profile under claude-profiles/<p>/ alike). The carve-out WINS inside
// allowlisted containers: a credentials*-named file inside an aged
// backups/removed-* directory is never deleted, even under --force
// (REQ-MCH-005).
//
// Any-segment matching may over-match unexpected segments — a debug log
// directory that happens to be named "bin" is protected too. That is
// deliberately conservative: over-protection, never over-deletion
// (audit residual R4).
//
// @MX:WARN: [AUTO] deletion guard — every deletion site under clean --home MUST consult this predicate in the same scan iteration that produced the candidate
// @MX:REASON: [AUTO] home-level --force is the highest-blast-radius change in this SPEC; a deletion that bypasses the guard can destroy credentials or state (SPEC-V3R6-MOAI-CLEAN-HOME-001 REQ-MCH-005)
func isCarvedOut(relPath string) bool {
	relPath = filepath.ToSlash(relPath)
	if relPath == "" || relPath == "." {
		return true // the root itself is never a deletion candidate
	}
	for _, seg := range strings.Split(relPath, "/") {
		if seg == "" {
			continue
		}
		if carveOutDirNames[seg] || carveOutFileNames[seg] {
			return true
		}
		if strings.HasPrefix(seg, "credentials") {
			return true
		}
	}
	return false
}

// homeStateYAML is the home-tier ~/.moai/config/sections/state.yaml shape read
// by the --home path (REQ-MCH-007, audit D3). It is deliberately a separate
// dedicated surface from the project-scope stateYAMLWrapper in clean.go,
// which stays byte-untouched: cwd-dependent retention reads would let two
// projects clean one home with different windows.
type homeStateYAML struct {
	State struct {
		// HomeRetentionDays is a pointer so an explicit 0 ("disable") is
		// distinguishable from an absent key ("use the compiled default").
		HomeRetentionDays *int `yaml:"home_retention_days"`
	} `yaml:"state"`
}

// loadHomeRetentionDays reads state.home_retention_days from the HOME tier —
// ~/.moai/config/sections/state.yaml resolved via paths.UserConfigSectionsDir().
// An absent file or key yields DefaultHomeCleanRetentionDays; an explicit 0
// disables cleaning (mirroring the existing retention semantics). Read and
// parse failures are returned as errors so callers choose their degradation.
func loadHomeRetentionDays() (int, error) {
	dir, err := paths.UserConfigSectionsDir()
	if err != nil {
		return 0, fmt.Errorf("resolve home config dir: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return config.DefaultHomeCleanRetentionDays, nil
		}
		return 0, fmt.Errorf("read home state.yaml: %w", err)
	}
	var wrapper homeStateYAML
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return 0, fmt.Errorf("parse home state.yaml: %w", err)
	}
	if wrapper.State.HomeRetentionDays == nil {
		return config.DefaultHomeCleanRetentionDays, nil
	}
	return *wrapper.State.HomeRetentionDays, nil
}

// homeCleanCandidate is one allowlisted deletion candidate under the home root.
type homeCleanCandidate struct {
	RelPath  string // root-relative, slash-separated
	AbsPath  string
	Size     int64
	Category string // debug | releases | logs | backups
}

// scanHomeCleanable enumerates the REQ-MCH-004 allowlist under root:
//
//   - per-profile debug/ entries older than retentionDays,
//   - releases/ binaries beyond the current version + the keep newest
//     (sidecar .sha256 files follow their binary),
//   - root logs/ files older than retentionDays,
//   - backups/removed-* directories older than retentionDays.
//
// Everything not enumerated is invisible to the scanner. The isCarvedOut
// guard is applied to every candidate inside this function — the same
// iteration that produced it — and a carved-out file inside an aged
// backups/removed-* directory skips the WHOLE container (audit residual R1:
// whole-dir skip, never partial delete of a backup).
func scanHomeCleanable(root string, retentionDays, releaseKeep int, currentVersion string, now time.Time) []homeCleanCandidate {
	var candidates []homeCleanCandidate
	if retentionDays <= 0 {
		return candidates // cleaning disabled
	}
	cutoff := now.AddDate(0, 0, -retentionDays)

	add := func(abs, category string, size int64) {
		relPath, relErr := relFromHomeRoot(root, abs)
		if relErr != nil {
			return
		}
		if isCarvedOut(relPath) {
			return
		}
		candidates = append(candidates, homeCleanCandidate{
			RelPath:  relPath,
			AbsPath:  abs,
			Size:     size,
			Category: category,
		})
	}

	// Category 1 — per-profile debug/ entries older than retention.
	profilesDir := filepath.Join(root, defs.ClaudeProfilesSubdir)
	if profiles, err := os.ReadDir(profilesDir); err == nil {
		for _, p := range profiles {
			if !p.IsDir() {
				continue
			}
			debugDir := filepath.Join(profilesDir, p.Name(), "debug")
			entries, err := os.ReadDir(debugDir)
			if err != nil {
				continue // no debug/ for this profile
			}
			for _, e := range entries {
				info, err := e.Info()
				if err != nil || !info.ModTime().Before(cutoff) {
					continue
				}
				abs := filepath.Join(debugDir, e.Name())
				size := info.Size()
				if e.IsDir() {
					dirBytes, _, walkErr := walkHomeSize(abs)
					if walkErr != nil {
						continue
					}
					size = dirBytes
				}
				add(abs, "debug", size)
			}
		}
	}

	// Category 2 — releases/ beyond current + keep newest.
	releasesDir := filepath.Join(root, defs.ReleasesSubdir)
	candidates = append(candidates, scanReleaseCandidates(root, releasesDir, releaseKeep, currentVersion)...)

	// Category 3 — root logs/ files older than retention.
	logsDir := filepath.Join(root, "logs")
	if entries, err := os.ReadDir(logsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, err := e.Info()
			if err != nil || !info.ModTime().Before(cutoff) {
				continue
			}
			add(filepath.Join(logsDir, e.Name()), "logs", info.Size())
		}
	}

	// Category 4 — backups/removed-* directories older than retention.
	backupsDir := filepath.Join(root, "backups")
	if entries, err := os.ReadDir(backupsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() || !strings.HasPrefix(e.Name(), "removed-") {
				continue
			}
			info, err := e.Info()
			if err != nil || !info.ModTime().Before(cutoff) {
				continue
			}
			abs := filepath.Join(backupsDir, e.Name())
			if homeDirContainsCarvedOut(root, abs) {
				continue // R1: whole-container skip
			}
			size, _, walkErr := walkHomeSize(abs)
			if walkErr != nil {
				continue
			}
			add(abs, "backups", size)
		}
	}

	return candidates
}

// scanReleaseCandidates returns the deletable release binaries (plus their
// .sha256 sidecars) under releasesDir: every moai-v* binary EXCEPT those
// named for currentVersion and EXCEPT the releaseKeep newest of the rest.
// version.json and LATEST are never candidates.
func scanReleaseCandidates(root, releasesDir string, releaseKeep int, currentVersion string) []homeCleanCandidate {
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return nil
	}
	type binStat struct {
		name  string
		mtime time.Time
		size  int64
	}
	currentPrefix := "moai-" + currentVersion
	var rest []binStat
	protected := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "moai-") || strings.HasSuffix(name, ".sha256") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		// Current-version binaries (including rc builds) always survive.
		if strings.HasPrefix(name, currentPrefix) {
			protected[name] = true
			continue
		}
		rest = append(rest, binStat{name: name, mtime: info.ModTime(), size: info.Size()})
	}
	sort.Slice(rest, func(i, j int) bool { return rest[i].mtime.After(rest[j].mtime) })
	deletable := releaseKeep < len(rest)
	if !deletable {
		return nil
	}
	older := rest[releaseKeep:] // everything past the keep-newest window

	var candidates []homeCleanCandidate
	appendBin := func(name string, size int64) {
		abs := filepath.Join(releasesDir, name)
		relPath, relErr := relFromHomeRoot(root, abs)
		if relErr != nil || isCarvedOut(relPath) {
			return
		}
		candidates = append(candidates, homeCleanCandidate{
			RelPath:  relPath,
			AbsPath:  abs,
			Size:     size,
			Category: "releases",
		})
	}
	for _, b := range older {
		appendBin(b.name, b.size)
		if info, err := os.Stat(filepath.Join(releasesDir, b.name+".sha256")); err == nil && !info.IsDir() {
			appendBin(b.name+".sha256", info.Size())
		}
	}
	return candidates
}

// homeDirContainsCarvedOut reports whether any path under dir (root-relative
// to the home root) touches a carved-out segment. Used for the R1
// whole-container skip on aged backups/removed-* directories. Walk errors are
// treated as carved-out (conservative: unreadable means unverified, and
// unverified never deletes).
func homeDirContainsCarvedOut(root, dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			found = true
			return fs.SkipAll
		}
		relPath, relErr := relFromHomeRoot(root, p)
		if relErr == nil && isCarvedOut(relPath) {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// relFromHomeRoot converts abs into a slash-separated path relative to the
// home root, for isCarvedOut's any-segment matching.
func relFromHomeRoot(root, abs string) (string, error) {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// walkHomeSize returns the total logical size and regular-file count under
// path (path itself may be a file — then it reports that file).
func walkHomeSize(path string) (int64, int, error) {
	var total int64
	var files int
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			if info, infoErr := d.Info(); infoErr == nil {
				total += info.Size()
				files++
			}
		}
		return nil
	})
	return total, files, err
}
