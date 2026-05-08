package worktree

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// @MX:ANCHOR: Snapshot/Diff primitives are the public API contract for orchestrator
// Bash invocation. Changing these signatures breaks moai worktree CLI subcommands.
// @MX:REASON: fan_in >= 3 (snapshot_io.go, divergence_log.go, internal/cli/worktree/guard.go).
// @MX:SPEC: SPEC-V3R3-CI-AUTONOMY-001 Wave 5

// Public constants exposed for CLI and other consumers.
// All paths are project-root relative; absolute resolution happens at CLI entry.
const (
	// SchemaVersion identifies the JSON snapshot schema version. Increment on
	// breaking schema changes; consumers must support graceful forward migration.
	SchemaVersion = "1.0.0"

	// StateDirRel is the relative directory for runtime state artifacts.
	StateDirRel = ".moai/state"

	// SnapshotPathTemplate formats snapshot file names. Use with fmt.Sprintf and
	// a snapshot ID; the result is project-root-relative.
	SnapshotPathTemplate = ".moai/state/worktree-snapshot-%s.json"

	// DivergenceReportDir is the relative directory where divergence markdown
	// reports are appended (per-day rolling).
	DivergenceReportDir = ".moai/reports/worktree-guard"

	// SuspectFlagTemplate formats suspect flag file names.
	SuspectFlagTemplate = ".moai/state/worktree-suspect-%s.flag"

	// UntrackedScope limits untracked-file enumeration to the SPEC area
	// (per OQ3 in strategy-wave5.md §7).
	UntrackedScope = ".moai/specs/"

	// DefaultGitTimeout caps each git subprocess invocation.
	DefaultGitTimeout = 30 * time.Second
)

// DefaultExclusions are paths that should never appear in untracked detection
// (these are runtime artifacts owned by MoAI tooling). Listed for documentation
// and future-proofing; .gitignore should already exclude them.
var DefaultExclusions = []string{
	".moai/reports/",
	".moai/cache/",
	".moai/logs/",
	".moai/state/",
}

// Snapshot captures working tree state at a point in time.
type Snapshot struct {
	SchemaVersion  string    `json:"schema_version"`
	CapturedAt     time.Time `json:"captured_at"`
	SnapshotID     string    `json:"snapshot_id"`
	HeadSHA        string    `json:"head_sha"`
	Branch         string    `json:"branch"`
	PorcelainLines []string  `json:"porcelain_lines"`
	UntrackedSpecs []string  `json:"untracked_specs"`
}

// CaptureOptions configures the snapshot capture behavior.
type CaptureOptions struct {
	// RepoDir is the working directory for git invocations. Empty = process CWD.
	RepoDir string
	// Timeout caps each git invocation. Zero = DefaultGitTimeout.
	Timeout time.Duration
	// SnapshotID overrides the auto-generated ID. Empty = auto-generate.
	SnapshotID string
}

// Capture takes a Snapshot of the current working tree state.
//
// On empty repository or detached HEAD the snapshot's HeadSHA / Branch fields
// degrade gracefully to the empty string; subsequent Diff still works.
func Capture(ctx context.Context, opts CaptureOptions) (*Snapshot, error) {
	if opts.Timeout == 0 {
		opts.Timeout = DefaultGitTimeout
	}
	cctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	snap := &Snapshot{
		SchemaVersion: SchemaVersion,
		CapturedAt:    time.Now().UTC(),
		SnapshotID:    opts.SnapshotID,
	}
	if snap.SnapshotID == "" {
		snap.SnapshotID = generateSnapshotID(snap.CapturedAt)
	}

	headSHA, err := runGit(cctx, opts.RepoDir, "rev-parse", "HEAD")
	if err != nil {
		// empty repo / unborn HEAD — degrade gracefully
		snap.HeadSHA = ""
	} else {
		snap.HeadSHA = strings.TrimSpace(headSHA)
	}

	branch, err := runGit(cctx, opts.RepoDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		snap.Branch = ""
	} else {
		snap.Branch = strings.TrimSpace(branch)
	}

	porcelain, err := runGit(cctx, opts.RepoDir, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git status --porcelain: %w", err)
	}
	// Filter out lines whose path is under DefaultExclusions (.moai/state/, .moai/reports/, ...).
	// Otherwise the snapshot file we are about to write would itself appear as untracked
	// in the next snapshot, producing spurious divergence (self-modification noise).
	snap.PorcelainLines = filterExcludedPorcelain(sortedNonEmptyLines(porcelain), DefaultExclusions)

	// ls-files --others --exclude-standard returns untracked files NOT ignored
	// by .gitignore, scoped to .moai/specs/.
	untracked, err := runGit(cctx, opts.RepoDir, "ls-files", "--others", "--exclude-standard", UntrackedScope)
	if err != nil {
		// scope path may not exist — empty result is correct
		snap.UntrackedSpecs = []string{}
	} else {
		snap.UntrackedSpecs = sortedNonEmptyLines(untracked)
	}

	return snap, nil
}

// Divergence categorizes how a post-state Snapshot differs from a pre-state Snapshot.
type Divergence struct {
	HeadChanged      bool     `json:"head_changed"`
	BranchChanged    bool     `json:"branch_changed"`
	UntrackedAdded   []string `json:"untracked_added"`
	UntrackedRemoved []string `json:"untracked_removed"`
	PorcelainDelta   []string `json:"porcelain_delta"`
	PreHeadSHA       string   `json:"pre_head_sha"`
	PostHeadSHA      string   `json:"post_head_sha"`
	PreBranch        string   `json:"pre_branch"`
	PostBranch       string   `json:"post_branch"`
}

// IsDivergent returns true when any dimension shows a difference.
// Per OQ4: binary detection (any difference = divergence).
func (d Divergence) IsDivergent() bool {
	return d.HeadChanged ||
		d.BranchChanged ||
		len(d.UntrackedAdded) > 0 ||
		len(d.UntrackedRemoved) > 0 ||
		len(d.PorcelainDelta) > 0
}

// Diff compares pre and post Snapshots and produces a Divergence report.
// Pre or post snapshots passed as nil cause a panic; callers are expected to
// have valid snapshots from Capture or LoadSnapshot.
func Diff(pre, post *Snapshot) Divergence {
	d := Divergence{
		PreHeadSHA:  pre.HeadSHA,
		PostHeadSHA: post.HeadSHA,
		PreBranch:   pre.Branch,
		PostBranch:  post.Branch,
	}
	if pre.HeadSHA != post.HeadSHA {
		d.HeadChanged = true
	}
	if pre.Branch != post.Branch {
		d.BranchChanged = true
	}
	d.UntrackedAdded = stringSliceDiff(post.UntrackedSpecs, pre.UntrackedSpecs)
	d.UntrackedRemoved = stringSliceDiff(pre.UntrackedSpecs, post.UntrackedSpecs)
	d.PorcelainDelta = porcelainDiff(pre.PorcelainLines, post.PorcelainLines)
	return d
}

// runGit executes a git command and returns combined output (stdout).
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// filterExcludedPorcelain removes porcelain lines whose path lies under any of
// the given exclusion prefixes. Porcelain format is "XY <path>" (status + space + path);
// rename entries "R  old -> new" are filtered when EITHER side matches an exclusion.
func filterExcludedPorcelain(lines, exclusions []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) < 4 {
			out = append(out, line)
			continue
		}
		body := line[3:]
		paths := []string{body}
		if idx := strings.Index(body, " -> "); idx >= 0 {
			paths = []string{strings.TrimSpace(body[:idx]), strings.TrimSpace(body[idx+4:])}
		}
		excluded := false
		for _, p := range paths {
			for _, ex := range exclusions {
				if strings.HasPrefix(p, ex) {
					excluded = true
					break
				}
			}
			if excluded {
				break
			}
		}
		if !excluded {
			out = append(out, line)
		}
	}
	return out
}

// sortedNonEmptyLines splits stdout on newline, drops empty lines, sorts result.
func sortedNonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimRight(line, "\r")
		if line != "" {
			out = append(out, line)
		}
	}
	sort.Strings(out)
	return out
}

// stringSliceDiff returns elements in a but not in b. Inputs need not be sorted.
func stringSliceDiff(a, b []string) []string {
	bset := make(map[string]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make([]string, 0)
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

// porcelainDiff returns symmetric line differences with +/- prefixes (post-only / pre-only).
func porcelainDiff(pre, post []string) []string {
	preSet := make(map[string]struct{}, len(pre))
	for _, x := range pre {
		preSet[x] = struct{}{}
	}
	postSet := make(map[string]struct{}, len(post))
	for _, x := range post {
		postSet[x] = struct{}{}
	}
	out := make([]string, 0)
	for x := range preSet {
		if _, ok := postSet[x]; !ok {
			out = append(out, "- "+x)
		}
	}
	for x := range postSet {
		if _, ok := preSet[x]; !ok {
			out = append(out, "+ "+x)
		}
	}
	sort.Strings(out)
	return out
}

// generateSnapshotID returns a deterministic-ish ID based on capture timestamp.
func generateSnapshotID(t time.Time) string {
	return fmt.Sprintf("snap-%d", t.UnixNano())
}
