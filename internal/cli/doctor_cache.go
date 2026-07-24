package cli

// doctor_cache.go — `moai doctor` cache hit-rate metric.
//
// Reports the 7-day rolling cache hit rate from .moai/state/cache-usage.jsonl,
// which the PostToolUse hook records on every turn, and warns when the
// single-turn session ratio exceeds 10%.

import (
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/state"
)

// cacheHitRateWindow is the 7-day rolling window for the doctor cache metric
// (REQ-PC-006, KPI K1 7-day rolling).
const cacheHitRateWindow = 7 * 24 * time.Hour

// singleTurnRatioThreshold is the threshold above which the doctor warns that
// one-shot sessions dominate the window.
const singleTurnRatioThreshold = 0.10

// checkCacheHitRate reports the 7-day cache hit rate from the cache-usage JSONL
// telemetry. The check has three outcomes:
//   - no readable telemetry in the window: OK with an "n/a" message.
//   - single-turn ratio <= 10%: OK with the "Cache hit rate (last 7 days): NN%"
//     line.
//   - single-turn ratio > 10%: WARN naming the cache-write cost of one-shot
//     sessions.
//
// The metric is deliberately NOT gated on cacheStrategy.enabled. Prompt caching
// is performed by Claude Code itself, not by moai — see
// .claude/rules/moai/workflow/cache-aware-execution.md, which states the
// orchestrator cannot place cache_control markers. The flag therefore never
// controlled whether caching happened, and gating this metric on it hid real
// PostToolUse telemetry behind a toggle that did nothing.
//
// @MX:ANCHOR: [AUTO] checkCacheHitRate — sole moai doctor cache metric surfacing (hit rate + single-turn warning)
// @MX:REASON: fan_in >= 3 — runGroupedChecks workspace registration, the doctor golden snapshots, and the single-turn-warning test all depend on this single check; the "Cache hit rate (last 7 days): NN%" output literal is grep-verified by the doctor tests.
func checkCacheHitRate(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Cache Hit Rate"}

	entries, err := state.ReadCacheUsage(projectRoot)
	if err != nil {
		// Telemetry read failure degrades to OK — never a hard failure.
		check.Status = uikit.CheckOK
		check.Message = "Cache hit rate (last 7 days): n/a (telemetry unreadable)"
		if verbose {
			check.Detail = err.Error()
		}
		return check
	}

	stats := state.AggregateCacheUsage(entries, time.Now().UTC(), cacheHitRateWindow)
	if stats.EntryCount == 0 {
		check.Status = uikit.CheckOK
		check.Message = "Cache hit rate (last 7 days): n/a (no telemetry in window)"
		return check
	}

	pct := int(stats.HitRate*100 + 0.5) // round to nearest percent
	check.Message = fmt.Sprintf("Cache hit rate (last 7 days): %d%%", pct)

	ratio := stats.SingleTurnRatio()
	if ratio > singleTurnRatioThreshold {
		check.Status = uikit.CheckWarn
		check.Detail = fmt.Sprintf(
			"WARN: %.0f%% of sessions in the window are single-turn (> %.0f%%). Each pays a cache-write cost with no later turn to read it back; batching work into fewer, longer sessions recovers that cost",
			ratio*100, singleTurnRatioThreshold*100,
		)
		return check
	}

	check.Status = uikit.CheckOK
	if verbose {
		check.Detail = fmt.Sprintf(
			"reads %d / creation %d tokens over %d entries (%d sessions, %d single-turn)",
			stats.TotalCacheRead, stats.TotalCacheCreation, stats.EntryCount,
			stats.TotalSessions, stats.SingleTurnSessions,
		)
	}
	return check
}
