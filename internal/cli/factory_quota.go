package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/statusline"
)

// SPEC-QUOTA-AWARE-SCHEDULING-001 M3 (REQ-QAS-009..012, -017): the lane gate.
//
// A Claude lane near its account quota limit leases no NEW card. This file owns
// the one pressure evaluation every surface shares, the Claude-lane predicate,
// the in-wait latch, and the hold line. It reads the quota settings and the
// session telemetry records the statusline writes; it opens no network
// connection, spawns no process, and writes nothing (REQ-QAS-007). Every
// failure — an unreadable configuration, an absent or unreadable record, a
// non-Claude lane — reads as "not held", so a gate that cannot decide leaves
// the existing lease path exactly as it was (fail open).

// The window names the hold line and the evaluation carry.
const (
	factoryQuotaWindowFiveHour = "five_hour"
	factoryQuotaWindowSevenDay = "seven_day"
)

// factoryQuotaAggregate is the reading seam: tests replace it to feed readings
// without writing records. The production value is the statusline aggregator.
var factoryQuotaAggregate = statusline.AggregateQuota

// factoryQuotaWindowState is one rate-limit window as the evaluation sees it:
// its aggregate reading and the hold percentage the configuration sets for it.
type factoryQuotaWindowState struct {
	Name    string
	Reading statusline.QuotaReading
	HoldPct int
}

// atOrAboveHold reports whether the window is fresh and at or above its own
// hold percentage. A reset or unknown window never is (REQ-QAS-005/-006).
func (w factoryQuotaWindowState) atOrAboveHold() bool {
	return w.Reading.State == statusline.QuotaFresh && w.Reading.UsedPercentage >= float64(w.HoldPct)
}

// factoryQuotaEvaluation is the result of the one pressure evaluation
// (REQ-QAS-017). It is the same for every caller; a surface applies its own
// caller rule to it (the lane gate: a Claude lane).
type factoryQuotaEvaluation struct {
	Enabled   bool
	MarginPct int
	Windows   []factoryQuotaWindowState // five_hour then seven_day; empty while the gate is disabled
}

// HeldWindows returns the windows at or above their hold percentage.
func (e factoryQuotaEvaluation) HeldWindows() []factoryQuotaWindowState {
	if !e.Enabled {
		return nil
	}
	var held []factoryQuotaWindowState
	for _, w := range e.Windows {
		if w.atOrAboveHold() {
			held = append(held, w)
		}
	}
	return held
}

// Pressure reports quota pressure: the gate is enabled and some fresh window is
// at or above its own hold percentage.
func (e factoryQuotaEvaluation) Pressure() bool { return len(e.HeldWindows()) > 0 }

// factoryQuotaEvaluate is the single pressure evaluation (REQ-QAS-017). It takes
// no caller input — the project root names a tree, never a backend — so the
// lane gate, the status block, the --auto recommendation, and the
// integration-window warning cannot disagree about whether pressure is on. A
// disabled gate reads no record at all.
//
// @MX:NOTE: [AUTO] The one pressure evaluation; the lane gate calls it today and the status block, the --auto recommendation, and the integration-window warning adopt it in M5/M6 — a surface that re-derived pressure would let two surfaces disagree.
// @MX:SPEC: SPEC-QUOTA-AWARE-SCHEDULING-001
func factoryQuotaEvaluate(root string) factoryQuotaEvaluation {
	gate := config.LoadQuotaGate(root)
	ev := factoryQuotaEvaluation{Enabled: gate.Enabled, MarginPct: gate.ReleaseMarginPct}
	if !gate.Enabled {
		return ev
	}
	agg := factoryQuotaAggregate(filepath.Join(root, ".moai", "state"), factoryCardNow(), gate.MaxAge)
	ev.Windows = []factoryQuotaWindowState{
		{Name: factoryQuotaWindowFiveHour, Reading: agg.FiveHour, HoldPct: gate.FiveHourHoldPct},
		{Name: factoryQuotaWindowSevenDay, Reading: agg.SevenDay, HoldPct: gate.SevenDayHoldPct},
	}
	return ev
}

// factoryQuotaClaudeLane reports whether this session is a Claude lane: its
// launch provider names Claude, falling back to the kanban backend variable
// only when no launch provider is set (the pattern of
// internal/hook/session_start_factory.go factoryLaunchEntry; the launch-provider
// variable does not depend on the kanban launcher). A glm, gpt, empty, or
// unrecognised value is not a Claude lane and is never held (REQ-QAS-012).
func factoryQuotaClaudeLane() bool {
	provider := os.Getenv(config.EnvMoaiLaunchProvider)
	if provider == "" {
		provider = os.Getenv(config.EnvMoaiKanbanBackend)
	}
	return provider == kanban.BackendClaude
}

// factoryQuotaLatch carries the hold state across the re-checks of one
// `moai factory next --wait` invocation (REQ-QAS-011). Nothing is persisted: a
// new invocation starts with an empty latch, so its first evaluation holds at
// the hold percentage itself.
type factoryQuotaLatch struct {
	held map[string]bool
}

// evaluate decides whether this lane is held now and, when it is, returns the
// hold line. A window stays held once latched until its reading falls below
// hold minus the release margin (a reading exactly at that value stays held),
// its reset time passes, or the reading is no longer fresh — the last because a
// waiting lane renders no statusline, so its record is never refreshed, and an
// aged-out reading is unknown, which never holds. The lane is held while ANY
// window is latched, so with both windows held it is released only when every
// held window has released.
func (l *factoryQuotaLatch) evaluate(root string) (bool, string) {
	if !factoryQuotaClaudeLane() {
		l.held = nil
		return false, ""
	}
	ev := factoryQuotaEvaluate(root)
	next := map[string]bool{}
	var segments []string
	if ev.Enabled {
		for _, w := range ev.Windows {
			if w.Reading.State != statusline.QuotaFresh {
				continue // reset or unknown: released
			}
			used := w.Reading.UsedPercentage
			if w.atOrAboveHold() || (l.held[w.Name] && used >= float64(w.HoldPct-ev.MarginPct)) {
				next[w.Name] = true
				segments = append(segments, factoryQuotaHoldSegment(w))
			}
		}
	}
	l.held = next
	if len(segments) == 0 {
		return false, ""
	}
	return true, factoryQuotaHoldPrefix + strings.Join(segments, "; ")
}

// factoryQuotaHoldPrefix begins the hold line — the only output that
// distinguishes a hold from an empty queue (REQ-QAS-010).
const factoryQuotaHoldPrefix = "quota hold: "

// factoryQuotaHoldSegment renders one held window: its name, its used
// percentage, and its reset instant as RFC 3339 UTC.
func factoryQuotaHoldSegment(w factoryQuotaWindowState) string {
	return fmt.Sprintf("%s used=%.1f%% resets_at=%s",
		w.Name, w.Reading.UsedPercentage, time.Unix(w.Reading.ResetsAt, 0).UTC().Format(time.RFC3339))
}
