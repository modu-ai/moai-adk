package spec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// archive.go — SPEC auto-archive (SPEC-SESSIONSTART-PERF-001 M2).
//
// M1 made drift detection asymptotically independent of the SPEC count. Archiving
// attacks the same problem from the other side: it BOUNDS the dataset, moving
// finished SPECs out of .moai/specs/ so every lifecycle scan that walks that
// directory has less to walk. The two are complementary — M1 makes the scan cheap
// per SPEC, M2 keeps the SPEC count from growing without limit.
//
// Destination: .moai/archive/specs/<year>/. That path sits OUTSIDE the
// os.ReadDir(.moai/specs) scan set by construction, so no scanner needs to learn
// to skip it. Archiving is a MOVE, never a delete: the directory stays git-tracked
// and grep-discoverable (REQ-SSP-007).

// Terminal statuses eligible for archival.
//
// This set deliberately DIVERGES from drift.go's isTerminalStatus, which excludes
// "completed". The two answer different questions: drift asks "can git positively
// infer this status?" (it can infer a close, so completed is not frontmatter-
// authoritative there), while archive asks "is this SPEC finished?" (a completed
// SPEC certainly is — and it is the largest slice of a mature corpus). Collapsing
// the two predicates would silently make archiving a no-op for most of the tree.
var archiveTerminalStatuses = []string{"completed", "superseded", "archived", "rejected"}

// Activity-source labels reported on each candidate.
const (
	// ActivitySourceGit means the last-activity date came from git history.
	ActivitySourceGit = "git"
	// ActivitySourceFrontmatter means git had no record of the SPEC, so the
	// frontmatter `updated:` date stood in.
	ActivitySourceFrontmatter = "frontmatter"
)

// archiveDirName segments — .moai/archive/specs/<year>/<SPEC-ID>.
const (
	moaiDirName        = ".moai"
	specsDirName       = "specs"
	archiveDirName     = "archive"
	archiveSpecsSubdir = "specs"
)

// IsArchiveTerminalStatus reports whether a frontmatter status is terminal for
// archive purposes (one of completed / superseded / archived / rejected).
func IsArchiveTerminalStatus(status string) bool {
	return slices.Contains(archiveTerminalStatuses, status)
}

// ArchiveCandidate is one archive-eligible SPEC.
type ArchiveCandidate struct {
	SPECID string `json:"spec_id"`
	Status string `json:"status"`

	// LastActivity is the newest commit touching the SPEC directory, or — when git
	// has no record of it — the frontmatter `updated:` date.
	LastActivity time.Time `json:"last_activity"`

	// ActivitySource records which of those two produced LastActivity.
	ActivitySource string `json:"activity_source"`

	// EraFinal reports whether era classification marks this SPEC as
	// grandfather-protected. It is REPORTED for operator review and NEVER gates the
	// decision — see the eligibility contract on planArchive.
	EraFinal bool `json:"era_final"`

	// SourceDir and DestDir are project-root-relative.
	SourceDir string `json:"source_dir"`
	DestDir   string `json:"dest_dir"`
}

// ArchiveOptions parameterises an archive scan.
type ArchiveOptions struct {
	// GraceDays is the window a terminal SPEC stays put after its last activity.
	// Non-positive means "unset" and resolves to config.DefaultArchiveGraceDays.
	GraceDays int

	// Now is the reference instant. The zero value means time.Now() — tests pin it.
	Now time.Time
}

// ArchivePlan is the result of an archive scan.
type ArchivePlan struct {
	Candidates []ArchiveCandidate `json:"candidates"`
	// Scanned is the number of SPEC directories examined (not just eligible ones).
	Scanned   int       `json:"scanned"`
	GraceDays int       `json:"grace_days"`
	Cutoff    time.Time `json:"cutoff"`
}

// archiveDeps is the injectable seam. lastActivity is the git-facing half (one
// subprocess, whatever N is); move is the filesystem-facing half.
type archiveDeps struct {
	lastActivity func(baseDir string) map[string]time.Time
	move         func(baseDir, src, dst string) error
}

func realArchiveDeps() archiveDeps {
	return archiveDeps{
		lastActivity: gitLastActivity,
		move:         gitMoveOrRename,
	}
}

// PlanArchive computes the archive-eligible set WITHOUT moving anything.
//
// This is the dry-run path and it is observation-only by construction: it never
// writes, so an operator can always preview a plan with no risk.
//
// @MX:ANCHOR: [AUTO] PlanArchive — the archive eligibility contract.
// @MX:REASON: SPEC-SESSIONSTART-PERF-001 REQ-SSP-010 [HARD]. Archival is
//
//	destructive-adjacent (it relocates directories), and the eligibility predicate
//	is the only thing standing between an operator and a mass relocation. See
//	.claude/rules/moai/core/verification-claim-integrity.md §5: a text-pattern
//	inference over frontmatter once nearly batch-touched 29 grandfather-protected
//	SPECs. Any change here must preserve the status+date-only predicate.
func PlanArchive(baseDir string, opts ArchiveOptions) (*ArchivePlan, error) {
	return planArchive(baseDir, opts, realArchiveDeps())
}

// ExecuteArchive computes the plan and relocates every eligible SPEC.
func ExecuteArchive(baseDir string, opts ArchiveOptions) (*ArchivePlan, error) {
	return executeArchive(baseDir, opts, realArchiveDeps())
}

// planArchive is the seam-injectable core of PlanArchive.
//
// Eligibility — a SPEC is archive-eligible iff BOTH hold:
//
//  1. its frontmatter status is terminal (IsArchiveTerminalStatus), AND
//  2. its last activity predates now − graceDays (strictly before the cutoff).
//
// Grandfather (era-final) status is ORTHOGONAL: it is reported on the candidate
// but is not consulted as a gate. It therefore neither FORCES archival (an
// era-final SPEC still in draft, or still inside the grace window, stays put) nor
// FORBIDS it (an era-final SPEC that independently satisfies both criteria is
// eligible like any other). That is exactly REQ-SSP-010.
func planArchive(baseDir string, opts ArchiveOptions, deps archiveDeps) (*ArchivePlan, error) {
	graceDays := opts.GraceDays
	if graceDays <= 0 {
		graceDays = config.DefaultArchiveGraceDays
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	cutoff := now.AddDate(0, 0, -graceDays)

	plan := &ArchivePlan{
		Candidates: []ArchiveCandidate{},
		GraceDays:  graceDays,
		Cutoff:     cutoff,
	}

	specsDir := filepath.Join(baseDir, moaiDirName, specsDirName)
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No SPECs to archive is a clean empty plan, not an error.
			return plan, nil
		}
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	// ONE git pass for the whole corpus — the archive scan must not reintroduce the
	// per-SPEC subprocess fan-out that M1 removed from the drift path.
	activity := deps.lastActivity(baseDir)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specID := entry.Name()
		specDir := filepath.Join(specsDir, specID)

		status, err := ParseStatus(specDir)
		if err != nil {
			// Unparseable: we cannot prove it is terminal, so we never touch it.
			continue
		}
		plan.Scanned++

		// Gate 1 — terminal status.
		if !IsArchiveTerminalStatus(status) {
			continue
		}

		// Gate 2 — past the grace window.
		last, source := resolveLastActivity(specDir, specID, activity)
		if last.IsZero() || !last.Before(cutoff) {
			// A SPEC with no discoverable date is treated as recently active
			// (fail-safe: absence of evidence is never evidence of staleness).
			continue
		}

		plan.Candidates = append(plan.Candidates, ArchiveCandidate{
			SPECID:         specID,
			Status:         status,
			LastActivity:   last,
			ActivitySource: source,
			EraFinal:       isEraFinal(specDir), // reported, never a gate
			SourceDir:      filepath.Join(moaiDirName, specsDirName, specID),
			DestDir:        archiveDestDir(specID, last),
		})
	}

	sort.Slice(plan.Candidates, func(i, j int) bool {
		return plan.Candidates[i].SPECID < plan.Candidates[j].SPECID
	})

	return plan, nil
}

// executeArchive plans, then relocates each eligible SPEC.
func executeArchive(baseDir string, opts ArchiveOptions, deps archiveDeps) (*ArchivePlan, error) {
	plan, err := planArchive(baseDir, opts, deps)
	if err != nil {
		return nil, err
	}

	for _, c := range plan.Candidates {
		if err := deps.move(baseDir, c.SourceDir, c.DestDir); err != nil {
			return nil, fmt.Errorf("failed to archive %s: %w", c.SPECID, err)
		}
	}

	return plan, nil
}

// archiveDestDir builds the year-partitioned destination, project-root-relative.
func archiveDestDir(specID string, last time.Time) string {
	year := strconv.Itoa(last.Year())
	return filepath.Join(moaiDirName, archiveDirName, archiveSpecsSubdir, year, specID)
}

// resolveLastActivity prefers git history and falls back to frontmatter.
//
// Git is authoritative because the frontmatter `updated:` field is hand-maintained
// and routinely goes stale — trusting it alone would archive SPECs that git shows
// were touched yesterday.
func resolveLastActivity(specDir, specID string, activity map[string]time.Time) (time.Time, string) {
	if t, ok := activity[specID]; ok && !t.IsZero() {
		return t, ActivitySourceGit
	}
	if t, ok := frontmatterUpdated(specDir); ok {
		return t, ActivitySourceFrontmatter
	}
	return time.Time{}, ""
}

// frontmatterUpdated reads the `updated:` date (YYYY-MM-DD) from spec.md.
func frontmatterUpdated(specDir string) (time.Time, bool) {
	content, err := os.ReadFile(filepath.Join(specDir, "spec.md"))
	if err != nil {
		return time.Time{}, false
	}

	fm, _, err := extractFrontmatter(string(content))
	if err != nil {
		return time.Time{}, false
	}

	t, err := time.Parse("2006-01-02", strings.TrimSpace(fm.Updated))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// isEraFinal reports grandfather protection. Read-only reuse of the era classifier —
// surfaced on the candidate for operator review, never used as an eligibility gate.
func isEraFinal(specDir string) bool {
	signals, err := LoadEraSignalsFromDir(specDir)
	if err != nil {
		return false
	}
	era, _ := ClassifyEra(signals)
	return era.EraFinal()
}

// gitActivityFormat emits a record separator + committer date per commit; the
// --name-only paths follow on their own lines.
const gitActivityFormat = "--format=" + gitLogRecordSep + "%cI"

// gitLastActivity resolves the newest commit date per SPEC directory in ONE git
// subprocess.
//
// The alternative — `git log -1 -- <specDir>` per SPEC — is O(n) subprocesses, the
// exact pattern M1 removed from the drift path. Reintroducing it here would make
// `moai spec archive` scale as badly as session-start used to.
//
// A git failure (no repository, no history) yields an empty map, and every SPEC
// then falls back to its frontmatter date.
func gitLastActivity(baseDir string) map[string]time.Time {
	specsPath := filepath.Join(moaiDirName, specsDirName)

	cmd := exec.Command("git", "log", "--no-merges", gitActivityFormat, "--name-only", "--", specsPath)
	cmd.Dir = baseDir

	output, err := cmd.Output()
	if err != nil {
		return map[string]time.Time{}
	}

	return parseGitActivity(string(output))
}

// parseGitActivity maps each SPEC ID to the newest commit that touched it.
//
// Input is git's newest-first stream, one record per commit: a date line followed
// by the paths that commit touched. The FIRST time a SPEC ID appears is therefore
// its newest touch — later appearances are older and are ignored.
func parseGitActivity(output string) map[string]time.Time {
	activity := make(map[string]time.Time)
	prefix := filepath.Join(moaiDirName, specsDirName) + string(filepath.Separator)

	for _, record := range strings.Split(output, gitLogRecordSep) {
		record = strings.TrimLeft(record, "\r\n")
		if record == "" {
			continue
		}

		lines := strings.Split(record, "\n")
		when, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[0]))
		if err != nil {
			continue
		}

		for _, path := range lines[1:] {
			path = strings.TrimSpace(path)
			if !strings.HasPrefix(path, prefix) {
				continue
			}

			// .moai/specs/<SPEC-ID>/<file> → <SPEC-ID>
			rest := strings.TrimPrefix(path, prefix)
			specID, _, found := strings.Cut(rest, string(filepath.Separator))
			if !found || specID == "" {
				continue
			}

			// Newest-first: the first sighting wins.
			if _, seen := activity[specID]; !seen {
				activity[specID] = when
			}
		}
	}

	return activity
}

// gitMoveOrRename relocates a SPEC directory, preferring `git mv` so the move is
// recorded as a tracked rename (history and grep-discoverability survive).
//
// It falls back to os.Rename when git mv cannot apply — an untracked SPEC, or no
// repository at all (the t.TempDir() case). The archived SPEC then still lands in
// the right place; only the rename-tracking is lost, which is exactly the
// degradation you want rather than a hard failure.
func gitMoveOrRename(baseDir, src, dst string) error {
	absDst := filepath.Join(baseDir, dst)
	if err := os.MkdirAll(filepath.Dir(absDst), 0o755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	if _, err := os.Stat(absDst); err == nil {
		return fmt.Errorf("archive destination already exists: %s", dst)
	}

	cmd := exec.Command("git", "mv", src, dst)
	cmd.Dir = baseDir
	if err := cmd.Run(); err == nil {
		return nil
	}

	if err := os.Rename(filepath.Join(baseDir, src), absDst); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
	}
	return nil
}

// osRenameMove is the git-free mover used by tests.
func osRenameMove(baseDir, src, dst string) error {
	absDst := filepath.Join(baseDir, dst)
	if err := os.MkdirAll(filepath.Dir(absDst), 0o755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	if err := os.Rename(filepath.Join(baseDir, src), absDst); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
	}
	return nil
}
