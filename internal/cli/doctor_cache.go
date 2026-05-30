package cli

// doctor_cache.go — SPEC-V3R6-PROMPT-CACHE-001 M4.
//
// `moai doctor` cache hit-rate metric. Reports the 7-day rolling cache hit rate
// from .moai/state/cache-usage.jsonl when cacheStrategy.enabled == true, and
// warns when the single-turn session ratio (K5) exceeds 10%.
//
// REQ-PC-006: report cache hit rate (last 7 days) when cacheStrategy.enabled.
// K5: single-turn ratio > 10% → WARN to consider session_ttl: "off".

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/state"
)

// cacheHitRateWindow is the 7-day rolling window for the doctor cache metric
// (REQ-PC-006, KPI K1 7-day rolling).
const cacheHitRateWindow = 7 * 24 * time.Hour

// singleTurnRatioThreshold is the K5 threshold: above this single-turn-session
// ratio, the doctor warns to consider session_ttl: "off".
const singleTurnRatioThreshold = 0.10

// checkCacheHitRate reports the 7-day cache hit rate from the cache-usage JSONL
// telemetry. The check has three outcomes:
//   - cacheStrategy.enabled == false (or no cache.yaml): reports OK with a
//     "caching disabled" message and NO "Cache hit rate" line (AC-PC-007 §3).
//   - enabled, single-turn ratio <= 10%: OK with the "Cache hit rate (last 7
//     days): NN%" line (AC-PC-007).
//   - enabled, single-turn ratio > 10%: WARN with a session_ttl: "off"
//     recommendation in Detail (K5).
//
// @MX:ANCHOR: [AUTO] checkCacheHitRate — sole moai doctor cache metric surfacing (hit rate + single-turn warning)
// @MX:REASON: fan_in >= 3 expected — runGroupedChecks workspace registration, AC-PC-007 doctor test, and the K5 single-turn-warning test all depend on this single check; the "Cache hit rate (last 7 days): NN%" output literal is grep-verified by AC-PC-007, and the enabled-gated absence (no line when disabled) is a contractual KPI surfacing behavior.
func checkCacheHitRate(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Cache Hit Rate"}

	cachePath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "cache.yaml")
	// LoadCacheConfig never errors: file-not-found / malformed → safe default
	// (enabled: false). The doctor metric only surfaces when caching is on.
	cfg, _ := config.LoadCacheConfig(cachePath)
	if cfg == nil || !cfg.Enabled {
		check.Status = CheckOK
		check.Message = "prompt caching disabled (cacheStrategy.enabled: false)"
		if verbose {
			check.Detail = "set .moai/config/sections/cache.yaml `cacheStrategy.enabled: true` to enable + surface hit rate"
		}
		return check
	}

	entries, err := state.ReadCacheUsage(projectRoot)
	if err != nil {
		// Telemetry read failure degrades to OK — never a hard failure.
		check.Status = CheckOK
		check.Message = "prompt caching enabled — no telemetry yet (cache-usage.jsonl unreadable)"
		if verbose {
			check.Detail = err.Error()
		}
		return check
	}

	stats := state.AggregateCacheUsage(entries, time.Now().UTC(), cacheHitRateWindow)
	if stats.EntryCount == 0 {
		check.Status = CheckOK
		check.Message = "Cache hit rate (last 7 days): n/a (no telemetry in window)"
		return check
	}

	pct := int(stats.HitRate*100 + 0.5) // round to nearest percent
	check.Message = fmt.Sprintf("Cache hit rate (last 7 days): %d%%", pct)

	ratio := stats.SingleTurnRatio()
	if ratio > singleTurnRatioThreshold {
		check.Status = CheckWarn
		check.Detail = fmt.Sprintf(
			"WARN: consider setting session_ttl: \"off\" — single-turn sessions are %.0f%% of the window (> %.0f%%), incurring 1h cache-write penalty",
			ratio*100, singleTurnRatioThreshold*100,
		)
		return check
	}

	check.Status = CheckOK
	if verbose {
		check.Detail = fmt.Sprintf(
			"reads %d / creation %d tokens over %d entries (%d sessions, %d single-turn)",
			stats.TotalCacheRead, stats.TotalCacheCreation, stats.EntryCount,
			stats.TotalSessions, stats.SingleTurnSessions,
		)
	}
	return check
}
