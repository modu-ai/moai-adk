package statusline

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-STATUSLINE-OFFICIAL-USAGE-001 M2 — the local context-usage snapshot
// fallback. While the stdin payload carries no rate_limits, the builder reads
// the account-level quota aggregate from the context-usage records of the
// primary checkout and every linked worktree (REQ-SOU-003) through the
// existing, PRESERVE-listed quota reader (quota.go AggregateQuota /
// quota_dirs.go AggregateQuotaDirs). Pure reads: no network, no process spawn,
// no write. Every failure degrades to "no snapshot" — the segments render n/a,
// never wrong data (REQ-SOU-004).

// quotaSnapshotMaxAge feeds the aggregator's max-age parameter (plan D4). It
// is the workflow.quota_gate.max_age default that
// internal/config/loader_quota_gate.go parses and validates — the config
// package's own constant, parsed once here, not a statusline-local inline
// literal. The loader's validated floor (twice QuotaHeartbeatInterval) is the
// fallback if the constant ever stops parsing.
var quotaSnapshotMaxAge = func() time.Duration {
	d, err := time.ParseDuration(config.DefaultQuotaGateMaxAge)
	if err != nil || d <= 0 {
		return 2 * config.QuotaHeartbeatInterval
	}
	return d
}()

// quotaScanMaxDirs bounds the worktree-metadata enumeration handed to
// QuotaStateDirs. A statusline render needs the recent records of the
// operator's own trees, not an unbounded enumeration.
const quotaScanMaxDirs = 64

// worktreeGitdirSeparator splits a worktree gitdir pointer into the primary
// checkout root and the per-worktree tail: <primary>/.git/worktrees/<name>.
const worktreeGitdirSeparator = "/.git/worktrees/"

// primaryRepoRoot resolves the repository root whose records a render anchored
// at projDir should read (plan D7). When the session is anchored at the
// primary checkout, projDir/.git is a directory and the answer is projDir —
// the helper is the identity over QuotaStateDirs's existing behavior. When
// projDir/.git is a regular file (a linked worktree), its first line names the
// primary checkout's per-worktree metadata (<primary>/.git/worktrees/<name>),
// so the primary root is that path minus the suffix. A RELATIVE pointer — git
// accepts relative worktree anchors — is resolved against projDir with
// filepath.Clean + filepath.Join before the suffix strip.
//
// Fail-open: an unreadable, malformed, or non-worktree-shaped pointer returns
// projDir unchanged, degrading the scan to the single-root behavior (worst
// case: n/a, never another tree's data — plan §F R5).
func primaryRepoRoot(projDir string) string {
	gitPath := filepath.Join(projDir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil || info.IsDir() {
		return projDir // primary checkout, or no git metadata: scan projDir itself
	}
	data, err := os.ReadFile(gitPath)
	if err != nil {
		return projDir
	}
	first := strings.TrimSpace(strings.SplitN(string(data), "\n", 2)[0])
	first = strings.TrimSuffix(first, "\r")
	ptr, ok := strings.CutPrefix(first, "gitdir:")
	if !ok {
		return projDir // not a worktree anchor
	}
	ptr = strings.TrimSpace(ptr)
	if !filepath.IsAbs(ptr) {
		ptr = filepath.Clean(filepath.Join(projDir, ptr))
	}
	ptr = filepath.ToSlash(ptr)
	idx := strings.LastIndex(ptr, worktreeGitdirSeparator)
	if idx <= 0 {
		return projDir // malformed or non-worktree pointer: fail open
	}
	return filepath.Clean(filepath.FromSlash(ptr[:idx]))
}

// quotaScanStateDirs names the state directories the snapshot fallback reads
// for a render anchored at projDir: the resolved primary checkout's state
// directory plus every linked worktree's (the current one included), per
// QuotaStateDirs (plan D7).
func quotaScanStateDirs(projDir string, maxDirs int) []string {
	return QuotaStateDirs(primaryRepoRoot(projDir), maxDirs)
}

// collectQuotaSnapshot reads the account-level quota aggregate for the render
// anchored at projDir (plan D8). The read phase runs under the render context:
// cancellation is honored on the read path itself — a select over aggregate
// delivery and ctx.Done(), not only an entry check — so an over-deadline read
// abandons the fallback (nil → the segments degrade to n/a) instead of
// stalling the render (REQ-SOU-006, AC-SOU-010). The abandoned os.ReadFile
// finishes in its own goroutine; that residual is named and accepted in
// REQ-SOU-006 (records are tool-written small JSON, not attacker-controlled).
// Returns nil when there is nothing to scan.
func collectQuotaSnapshot(ctx context.Context, projDir string, now time.Time) *QuotaAggregate {
	dirs := quotaScanStateDirs(projDir, quotaScanMaxDirs)
	if len(dirs) == 0 {
		return nil
	}
	type quotaResult struct{ agg QuotaAggregate }
	ch := make(chan quotaResult, 1)
	go func() {
		ch <- quotaResult{AggregateQuotaDirs(dirs, now, quotaSnapshotMaxAge)}
	}()
	select {
	case r := <-ch:
		return &r.agg
	case <-ctx.Done():
		slog.Debug("quota snapshot read abandoned", "reason", ctx.Err())
		return nil
	}
}

// quotaWindowReading returns one window's reading from the aggregate, unknown
// when the aggregate is absent.
func quotaWindowReading(snap *QuotaAggregate, sevenDay bool) QuotaReading {
	if snap == nil {
		return QuotaReading{State: QuotaUnknown}
	}
	if sevenDay {
		return snap.SevenDay
	}
	return snap.FiveHour
}
