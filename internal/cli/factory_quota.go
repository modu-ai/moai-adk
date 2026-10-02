package cli

import (
	"fmt"
	"io"
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
// without writing records. Its type is the single-directory aggregator's. The
// production value reads the quota records of the primary checkout and of every
// linked worktree (SPEC-QUOTA-RECORD-WORKTREES-001 REQ-QWR-006, -007): it
// derives the repository root from the state-directory path, which the one
// caller builds as <root>/.moai/state, enumerates through factoryQuotaStateDirs
// with the configured bound, and aggregates across the directories. A state
// directory of any other shape is read alone, as before (fail open).
var factoryQuotaAggregate = func(stateDir string, now time.Time, maxAge time.Duration) statusline.QuotaAggregate {
	clean := filepath.Clean(stateDir)
	root := filepath.Dir(filepath.Dir(clean))
	if filepath.Join(root, ".moai", "state") != clean {
		return statusline.AggregateQuota(stateDir, now, maxAge)
	}
	dirs := factoryQuotaStateDirs(root, config.LoadQuotaScanBound(root))
	return statusline.AggregateQuotaDirs(dirs, now, maxAge)
}

// factoryQuotaStateDirs is the worktree-enumeration seam (SPEC-QUOTA-RECORD-WORKTREES-001
// REQ-QWR-006): the one place this package references the enumerator, so a test
// can count its calls and a static check can pin that no code path reaches it
// before the gate check. The production value is the statusline enumerator.
var factoryQuotaStateDirs = statusline.QuotaStateDirs

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
// @MX:ANCHOR: [AUTO] The one pressure evaluation — the lane gate, the integration-window warning, the status block, and the --auto steering line all funnel through it.
// @MX:REASON: 4 non-test call sites (factory_quota.go x3, factory_quota_lanes.go); it takes no caller input, and the lane gate and the acquire warning apply the Claude-caller predicate to its result, so a surface that re-derived pressure or fed it a backend would let two surfaces disagree (REQ-QAS-017).
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
//
// @MX:ANCHOR: [AUTO] The one formatter of a held window's segment — shared by the hold line, the acquire warning, and the steering line.
// @MX:REASON: 3 non-test call sites (factory_quota.go x2, factory_quota_lanes.go); changing the format changes three user-visible lines at once (REQ-QAS-010/-014).
// @MX:SPEC: SPEC-QUOTA-AWARE-SCHEDULING-001
func factoryQuotaHoldSegment(w factoryQuotaWindowState) string {
	return fmt.Sprintf("%s used=%.1f%% resets_at=%s",
		w.Name, w.Reading.UsedPercentage, time.Unix(w.Reading.ResetsAt, 0).UTC().Format(time.RFC3339))
}

// SPEC-QUOTA-AWARE-SCHEDULING-001 M6 (REQ-QAS-014, DO-7 final): the
// integration-window warning.

// integrationQuotaWarningPrefix begins the warning line.
const integrationQuotaWarningPrefix = "quota warning: "

// integrationQuotaWarningTail ends it; the tail states that the window is still
// taken, so a reader cannot take the line for a refusal (DO-7: warn-only).
const integrationQuotaWarningTail = " (warn-only; the integration window is still taken)"

// integrationQuotaWarning is the line `moai integration acquire` prints on the
// error stream when a Claude lane takes the window under quota pressure, naming
// each held window and its reset time with the lane gate's own segment; it is
// empty when there is nothing to say — a non-Claude caller, a disabled gate, no
// data, or pressure off — and so is every failure of the quota read, which
// reads as no pressure (fail open). It calls the one shared evaluation and adds
// only the caller rule (a Claude lane); it never blocks, refuses, or delays.
func integrationQuotaWarning(root string) string {
	if !factoryQuotaClaudeLane() {
		return ""
	}
	held := factoryQuotaEvaluate(root).HeldWindows()
	if len(held) == 0 {
		return ""
	}
	segments := make([]string, 0, len(held))
	for _, w := range held {
		segments = append(segments, factoryQuotaHoldSegment(w))
	}
	return integrationQuotaWarningPrefix + strings.Join(segments, "; ") + integrationQuotaWarningTail
}

// SPEC-QUOTA-AWARE-SCHEDULING-001 M5 (REQ-QAS-013, -019, -020): the read-only
// quota block of `moai factory status`.

// factoryQuotaWindowView is one window of the status block (REQ-QAS-013): its
// state, the reading's percentage (fresh only), reset instant, source capture
// time, whether it would hold a Claude lane, and the first-observed-exhausted
// time where the source record carries one (DO-13). An unknown window carries
// only its name, state, and the held flag (false).
type factoryQuotaWindowView struct {
	Window          string   `json:"window"`
	State           string   `json:"state"`
	UsedPercentage  *float64 `json:"used_percentage,omitempty"`
	ResetsAt        string   `json:"resets_at,omitempty"`
	CapturedAt      string   `json:"captured_at,omitempty"`
	HoldsClaudeLane bool     `json:"holds_claude_lane"`
	ExhaustedAt     string   `json:"exhausted_at,omitempty"`
}

// factoryQuotaSteeringView is the steering part of the block, present only while
// quota pressure is on (REQ-QAS-019, -020): the candidate lanes, the count of
// live lanes of unknown backend, and the warning marker when there is no
// candidate. Its fields are embedded in the block's JSON object.
type factoryQuotaSteeringView struct {
	Candidates   []factoryQuotaLane `json:"candidates"`
	UnknownLanes int                `json:"unknown_lanes"`
	Warning      string             `json:"warning,omitempty"`
}

// factoryQuotaBlock is the `quota` key of the status report and the block under
// the text report. line is the steering line, set only while pressure is on.
type factoryQuotaBlock struct {
	Pressure bool                     `json:"pressure"`
	Windows  []factoryQuotaWindowView `json:"windows"`
	*factoryQuotaSteeringView
	line string
}

// factoryQuotaStatusBlock builds the status block, or nil when there is nothing
// to report: the gate is disabled (no record is read at all) or no window has
// data — a window reads as data when it is fresh or reset, so a directory of
// stale or window-less records adds no block. It takes the one shared pressure
// evaluation and, only when pressure is on, reads the lane inventory.
func factoryQuotaStatusBlock(root string) *factoryQuotaBlock {
	ev := factoryQuotaEvaluate(root)
	if !ev.Enabled {
		return nil
	}
	data := false
	views := make([]factoryQuotaWindowView, 0, len(ev.Windows))
	for _, w := range ev.Windows {
		if w.Reading.State != statusline.QuotaUnknown {
			data = true
		}
		views = append(views, factoryQuotaWindowViewOf(w))
	}
	if !data {
		return nil
	}
	block := &factoryQuotaBlock{Pressure: ev.Pressure(), Windows: views}
	if block.Pressure {
		inv := factoryQuotaReadLanes(root)
		steer := &factoryQuotaSteeringView{Candidates: inv.Candidates, UnknownLanes: inv.Unknown}
		if steer.Candidates == nil {
			steer.Candidates = []factoryQuotaLane{}
		}
		if len(steer.Candidates) == 0 {
			steer.Warning = factoryQuotaWarningNoLane
		}
		block.factoryQuotaSteeringView = steer
		block.line = factoryQuotaSteeringLine(ev.HeldWindows(), inv)
	}
	return block
}

// factoryQuotaWindowViewOf renders one window's reading.
func factoryQuotaWindowViewOf(w factoryQuotaWindowState) factoryQuotaWindowView {
	r := w.Reading
	v := factoryQuotaWindowView{Window: w.Name, State: string(r.State), HoldsClaudeLane: w.atOrAboveHold()}
	if r.State == statusline.QuotaUnknown {
		return v
	}
	if r.State == statusline.QuotaFresh {
		used := r.UsedPercentage
		v.UsedPercentage = &used
	}
	v.ResetsAt = time.Unix(r.ResetsAt, 0).UTC().Format(time.RFC3339)
	if !r.CapturedAt.IsZero() {
		v.CapturedAt = r.CapturedAt.UTC().Format(time.RFC3339)
	}
	if !r.ExhaustedAt.IsZero() {
		v.ExhaustedAt = r.ExhaustedAt.UTC().Format(time.RFC3339)
	}
	return v
}

// writeFactoryQuotaText prints the block: one `quota <window>:` line per window,
// then the steering line while pressure is on. The informational lines never
// start with `quota pressure:` or `quota hold:`, so a pressure-off block carries
// no steering or hold line.
func writeFactoryQuotaText(w io.Writer, b *factoryQuotaBlock) {
	for _, v := range b.Windows {
		cells := []string{"state=" + v.State}
		if v.UsedPercentage != nil {
			cells = append(cells, fmt.Sprintf("used=%.1f%%", *v.UsedPercentage))
		}
		if v.State != string(statusline.QuotaUnknown) {
			cells = append(cells, "resets_at="+v.ResetsAt)
			if v.CapturedAt != "" {
				cells = append(cells, "captured_at="+v.CapturedAt)
			}
			holds := "no"
			if v.HoldsClaudeLane {
				holds = "yes"
			}
			cells = append(cells, "holds_claude_lane="+holds)
			if v.ExhaustedAt != "" {
				cells = append(cells, "exhausted_at="+v.ExhaustedAt)
			}
		}
		_, _ = fmt.Fprintf(w, "quota %s: %s\n", v.Window, strings.Join(cells, " "))
	}
	if b.line != "" {
		_, _ = fmt.Fprintln(w, b.line)
	}
}
