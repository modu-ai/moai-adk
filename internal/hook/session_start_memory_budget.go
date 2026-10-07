package hook

// session_start_memory_budget.go — the MEMORY.md size advisory for the
// SessionStart hook (SPEC-MEMORY-FOLD-BUDGET-001 follow-up card, REQ-MFB-011
// / REQ-MFB-012).
//
// One line, prefixed "[moai:memory-budget]", joined to the session's
// additionalContext and to no other output channel, whenever the store
// resolved for the session holds a MEMORY.md at or above the warn percentage
// on bytes or lines. Below the threshold, with no store, on any read error,
// or with MOAI_MEMORY_AUDIT=0, it adds nothing. The check reads one file,
// never blocks or fails session start, and never reads a path outside the
// stores derived for the session.
//
// The advisory runs on EVERY session start — startup, resume, clear and
// compact alike — because the wiring sits in Handle past every source
// condition, next to the guard-liveness advisory (REQ-MFB-012).

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// memoryBudgetPrefix starts the advisory line. The bracketed token is the
// greppable marker the orchestrator and the operator recognize.
const memoryBudgetPrefix = "[moai:memory-budget]"

// memoryBudgetJoinBound caps how long session start waits for the MEMORY.md
// budget read before giving up on the advisory for this session
// (REQ-MFB-011; the bound is a configuration constant per spec C-3, and it
// carries its own value with a ceiling strictly below the 5s per-event hook
// timeout — audit finding D23's join-bound half; the test asserts both by
// name). The var, not the constant, is what the code reads so the
// join-bound test can override it.
var memoryBudgetJoinBound = config.DefaultMemoryBudgetJoinBound

// memoryBudgetReadFile is the read seam every candidate MEMORY.md goes
// through. Production is os.ReadFile; the package TestMain installs a
// path-recording recorder over it (audit finding D5) that returns
// not-exist for — and never opens — any path outside the test home sandbox,
// and a post-run containment assertion fails the package naming every
// recorded path outside it. The seam is the containment boundary: a read
// the advisory attempts is a read the recorder sees.
var memoryBudgetReadFile = os.ReadFile

// memoryBudgetSample is the one read the advisory bounded-waits for: the
// store directory the sample came from and the MEMORY.md bytes.
type memoryBudgetSample struct {
	store string
	data  []byte
}

// memoryBudgetAdvisory returns the one-line budget warning for the store
// resolved for this session, or the empty string when there is nothing to
// say. Empty covers every non-warning outcome, distinguished nowhere on
// purpose: below the threshold, no store, unreadable MEMORY.md, a read
// error, the kill switch, and a read that overruns the join bound are all
// silence — a session start is not a diagnostic report (same posture as
// binaryLagAdvisory).
//
// The async parameter is the CALLER's answer (the test-binary seam and the
// per-handler option, h.asyncDeferredScans()), for the same
// goroutine-outliving-the-test reason the deferred-scan and binary-lag
// advisories use it.
//
// @MX:WARN @MX:REASON bounded-background-goroutine join — a read that
// overruns the bound is abandoned for this session; safe because the
// advisory is read-only and re-derives identically at the next session
// start.
// @MX:SPEC: SPEC-MEMORY-FOLD-BUDGET-001
func memoryBudgetAdvisory(ctx context.Context, dir string, async bool) string {
	if dir == "" {
		return ""
	}
	if os.Getenv(config.EnvMemoryAudit) == "0" {
		return ""
	}

	// Capture the read dependency before starting a goroutine: an abandoned
	// scan must not observe a later caller replacing the package-level seam.
	readFile := memoryBudgetReadFile

	// read walks the candidate stores until one yields a MEMORY.md. It checks
	// ctx between candidates, so an abandoned scan (the bound fired, the
	// caller cancelled) stops issuing reads instead of continuing through the
	// remaining candidates — the bound caps the advisory's WORK, not only the
	// caller's wait for it.
	read := func(ctx context.Context) memoryBudgetSample {
		for _, store := range memoryBudgetStoreCandidates(dir) {
			if ctx.Err() != nil {
				return memoryBudgetSample{}
			}
			data, err := readFile(filepath.Join(store, "MEMORY.md"))
			if err != nil {
				continue
			}
			return memoryBudgetSample{store: store, data: data}
		}
		return memoryBudgetSample{}
	}

	if !async {
		// Test path (TestMain sets deferredScansAsync=false, or the caller
		// passed WithSynchronousDeferredScans): run inline so no goroutine
		// outlives the test boundary, matching binaryLagAdvisory.
		return renderMemoryBudgetLine(read(ctx))
	}

	// The scan runs under a child context the join cancels on abandonment,
	// so a read still in flight when the bound fires is the LAST work this
	// advisory does for the session.
	scanCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Buffered so a read that finishes after the bound has elapsed can still
	// send and exit rather than leaking blocked on an abandoned channel.
	result := make(chan memoryBudgetSample, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Debug("session start: memory budget read panicked (non-blocking)", "recover", r)
			}
		}()
		result <- read(scanCtx)
	}()

	timer := time.NewTimer(memoryBudgetJoinBound)
	defer timer.Stop()
	select {
	case sample := <-result:
		return renderMemoryBudgetLine(sample)
	case <-timer.C:
		cancel()
		slog.Debug("session start: memory budget read exceeded join bound (non-blocking)",
			"bound", memoryBudgetJoinBound.String())
		return ""
	case <-ctx.Done():
		return ""
	}
}

// renderMemoryBudgetLine turns the one read into the advisory line, or ""
// when the measurements sit below the warn percentage on both axes.
//
// The trigger is the spec's own wording — "at or above the warn percentage
// on bytes or lines" — evaluated with the doctor's integer test
// (value*100 >= warnPercent*cap) over the same measurements
// (taxonomy.MeasureIndex, line axis on OverflowLines) so the two surfaces
// cannot disagree about a percentage. One deliberate divergence from
// taxonomy.AuditIndexBudget: the doctor suppresses its line-axis warning
// above the line cap because MEMORY_INDEX_OVERFLOW owns that range, while
// this advisory has one line and no second finding to hand the signal to,
// so an over-cap index still counts as at-or-above the warn percentage and
// fires.
func renderMemoryBudgetLine(s memoryBudgetSample) string {
	if s.store == "" || len(s.data) == 0 {
		return ""
	}
	byteCap := config.DefaultMemoryIndexByteCap
	lineCap := config.DefaultMemoryIndexLineCap
	warn := config.DefaultMemoryIndexWarnPercent
	if byteCap <= 0 || lineCap <= 0 || warn <= 0 || warn > 100 {
		return ""
	}

	m := taxonomy.MeasureIndex(s.data)
	overBytes := m.Bytes*100 >= warn*byteCap
	overLines := m.OverflowLines*100 >= warn*lineCap
	if !overBytes && !overLines {
		return ""
	}

	// The larger of the two percentages and the measure it is keyed on
	// (REQ-MFB-011); the percentages come from the same integer division
	// the doctor's findings print.
	pct, axis, axisCap := m.Bytes*100/byteCap, "bytes", byteCap
	if linePct := m.OverflowLines * 100 / lineCap; linePct > pct {
		pct, axis, axisCap = linePct, "lines", lineCap
	}
	return fmt.Sprintf("%s %s: MEMORY.md at %d%% of the %d-%s cap — run `moai memory doctor`",
		memoryBudgetPrefix, s.store, pct, axisCap, axis)
}

// memoryBudgetStoreCandidates returns the memory stores this session could
// be using, in the order a session resolves them: for the session's working
// directory — and, inside a linked worktree, the repository's primary
// checkout beside it — the active profile's config dir key first, then the
// default ~/.claude key.
//
// This is the hook-side half of OD-9's no-shared-resolver decision: sharing
// internal/cli's memoryCandidateStores would cost migrating it into a
// neutral package (spec.md §4 records the corrected import fact), so the
// hook derives the SAME two keys in the SAME order instead, and the
// independent-expectation test guards the pair against drift — it computes
// the expected path with its own literal slug mapping, so a future edit to
// either site that changes key order or spelling breaks that test loudly.
func memoryBudgetStoreCandidates(projectRoot string) []string {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil
	}
	home, homeErr := os.UserHomeDir()

	var candidates []string
	add := func(dir string) {
		if dir != "" && !slices.Contains(candidates, dir) {
			candidates = append(candidates, dir)
		}
	}

	roots := []string{abs}
	if primary := memoryBudgetPrimaryCheckout(abs); primary != "" {
		roots = append(roots, primary)
	}
	for _, r := range roots {
		slug := projectSlug(r)
		if cfg := os.Getenv(config.EnvClaudeConfigDir); cfg != "" {
			add(filepath.Join(cfg, "projects", slug, "memory"))
		}
		if homeErr == nil {
			add(filepath.Join(home, ".claude", "projects", slug, "memory"))
		}
	}
	return candidates
}

// memoryBudgetPrimaryCheckout returns the primary checkout of the repository
// abs belongs to, or "" when abs is not inside a repository, git cannot
// answer, or abs already IS the primary checkout. It mirrors internal/cli's
// memoryPrimaryCheckout through the same gitcore.ResolveGitDirs mechanism;
// the mirror is deliberate (OD-9: no shared resolver), and both
// normalizations measured there — git's symlink-resolved answer, and the
// EvalSymlinks comparison that keeps a primary checkout reached through a
// symlinked parent from becoming a spurious second root — are reproduced
// here for the same reasons.
func memoryBudgetPrimaryCheckout(abs string) string {
	dirs, err := gitcore.ResolveGitDirs(abs)
	if err != nil || dirs.CommonDir == "" {
		return ""
	}
	primary := filepath.Dir(dirs.CommonDir)
	if resolved, err := filepath.EvalSymlinks(abs); err == nil && resolved == primary {
		return ""
	}
	return primary
}
