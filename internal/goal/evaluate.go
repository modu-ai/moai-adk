package goal

import (
	"context"
	"fmt"
	"strings"
)

// CmdRunner abstracts shell-command execution so the evaluator is unit-testable
// without spawning real processes. The production hook wires a real runner
// (internal/cli/hook_stop_goal.go); tests inject a fake.
type CmdRunner interface {
	Run(ctx context.Context, cmd string) (exitCode int, output string, err error)
}

// SnapshotSource resolves a fresh recorded result for a Tier-1 mechanical
// condition by EXACT byte-string command match (no normalization — a variant
// differing by one byte never matches). ok=false on miss, stale snapshot,
// key-computation error, or time-box exceed; the caller then executes the
// command exactly as today. The production source is internal/verify.Source
// (structural match — this package does not import it); tests inject a fake.
type SnapshotSource interface {
	Lookup(ctx context.Context, cmd string) (exitCode int, attribution string, ok bool)
}

// FailedCond records a failed mechanical condition plus its output tail so the
// orchestrator's confirm AskUserQuestion can surface WHY the goal is not
// satisfied (REQ-GLE-010 ↔ REQ-GLE-028 reconciliation — the failed-condition
// + tail detail rides the plain block `reason` in autonomous mode and the
// checkpoint's `failed_conditions` array in semi-autonomous mode).
type FailedCond struct {
	Cmd  string `json:"cmd"`
	Exit int    `json:"exit"`
	Tail string `json:"tail"`
}

// CeilingVerdict is the 5-section evidence-bearing report emitted when the turn
// ceiling is reached (REQ-GLE-013) or stagnation halts the loop (REQ-GLE-017).
// The five section names are load-bearing — AC-GLE-013 greps for them.
type CeilingVerdict struct {
	Claim               string `json:"Claim"`
	Evidence            string `json:"Evidence"`
	BaselineAttribution string `json:"Baseline-attribution"`
	Gaps                string `json:"Gaps"`
	ResidualRisk        string `json:"Residual-risk"`
}

// Verdict is the Stop-hook stdout payload. A non-empty Decision ("block")
// continues the turn per Claude Code hook semantics; an empty Decision lets the
// turn end.
type Verdict struct {
	Decision         string          `json:"decision,omitempty"`
	Reason           string          `json:"reason,omitempty"`
	Mode             string          `json:"mode,omitempty"`
	Turn             int             `json:"turn,omitempty"`
	Ceiling          int             `json:"ceiling,omitempty"`
	LastProgress     string          `json:"last_progress,omitempty"`
	FailedConditions []FailedCond    `json:"failed_conditions,omitempty"`
	CeilingExit      bool            `json:"ceiling_exit,omitempty"`
	Stagnation       bool            `json:"stagnation,omitempty"`
	Verdict          *CeilingVerdict `json:"verdict,omitempty"`
	Yielded          bool            `json:"yielded,omitempty"`
	// SnapshotAttribution records, per reused Tier-1 condition, the snapshot
	// citation (path + key + command + recorded exit) — the evidence-attribution
	// trail for results served from the shared diagnostic snapshot instead of
	// re-execution.
	SnapshotAttribution []string `json:"snapshot_attribution,omitempty"`
}

// Eval carries the evaluator's injectable dependencies.
type Eval struct {
	Runner              CmdRunner
	StagnationThreshold int  // default DefaultStagnationThreshold when zero
	NativeGoalActive    bool // set when the runtime signals an active native /goal (REQ-GLE-016)
	// Snapshot, when non-nil, is consulted before executing a Tier-1 mechanical
	// condition: an exact-match fresh entry reuses the recorded exit code
	// without a CmdRunner call. nil preserves the pre-existing behavior
	// unchanged (strictly additive contract).
	Snapshot SnapshotSource
}

// DefaultStagnationThreshold is the number of consecutive identical progress
// notes that trigger the stagnation guard (REQ-GLE-017).
const DefaultStagnationThreshold = 3

// outputTailLen is the max number of bytes of command output retained in a
// failed-condition record (keeps the block JSON compact).
const outputTailLen = 200

func tail(s string) string {
	if len(s) <= outputTailLen {
		return s
	}
	return s[len(s)-outputTailLen:]
}

func lastProgress(g *Goal) string {
	if len(g.Progress) == 0 {
		return ""
	}
	return g.Progress[len(g.Progress)-1].Note
}

// isStagnant reports whether the goal's progress log shows N consecutive
// identical notes (N = StagnationThreshold).
func (e *Eval) isStagnant(g *Goal) bool {
	n := e.StagnationThreshold
	if n == 0 {
		n = DefaultStagnationThreshold
	}
	if len(g.Progress) < n {
		return false
	}
	last := g.Progress[len(g.Progress)-1].Note
	for i := len(g.Progress) - 1; i >= len(g.Progress)-n; i-- {
		if g.Progress[i].Note != last {
			return false
		}
	}
	return true
}

// Evaluate runs the 2-tier evaluation cycle (plan §B.3 steps 1-8). It MUTATES
// the goal (increments TurnsUsed, sets Status, appends a Progress entry) and
// returns the Verdict to emit + whether the turn should be blocked.
//
// The goal is persisted by the caller after Evaluate returns; Evaluate itself
// performs no I/O (the CmdRunner handles command execution).
func (e *Eval) Evaluate(ctx context.Context, g *Goal) (Verdict, bool) {
	// Step 1: inactive goal → no block.
	if g == nil || g.Status == StatusCleared || g.Status == StatusSatisfied {
		return Verdict{}, false
	}

	// Step 4 (checked before ceiling so a native /goal always wins): yield.
	if e.NativeGoalActive {
		return Verdict{Yielded: true, Reason: "native /goal active — stop-goal yields (no double-block)"}, false
	}

	// Step 2: increment turns; ceiling → 5-section verdict, no block.
	g.TurnsUsed++
	if g.Ceiling.MaxTurns > 0 && g.TurnsUsed >= g.Ceiling.MaxTurns {
		g.Status = StatusCeilingExit
		v := Verdict{
			CeilingExit: true,
			Verdict:     e.ceilingReport(g, "turn ceiling reached"),
		}
		e.appendProgress(g, "ceiling-exit")
		return v, false
	}

	// Step 3: stagnation → stop + E1/E3 escalation note, no block.
	if e.isStagnant(g) {
		g.Status = StatusCeilingExit
		v := Verdict{
			Stagnation: true,
			Verdict:    e.ceilingReport(g, "stagnation: no progress for N iterations (escalate E1/E3)"),
		}
		e.appendProgress(g, "stagnation-exit")
		return v, false
	}

	// Step 5: Tier 1 — evaluate mechanical conditions. A fresh snapshot entry
	// exactly matching the condition command (byte-string equality) reuses the
	// recorded exit code without executing; any miss/stale falls back to the
	// existing CmdRunner execution path unchanged.
	var failed []FailedCond
	var attributions []string
	hasMechanical := false
	for _, c := range g.Conditions {
		if c.Type != ConditionMechanical {
			continue
		}
		hasMechanical = true
		var exit int
		var out string
		var err error
		reused := false
		if e.Snapshot != nil {
			if recordedExit, attr, ok := e.Snapshot.Lookup(ctx, c.Cmd); ok {
				exit, reused = recordedExit, true
				out = "(reused from " + attr + ")"
				attributions = append(attributions, attr)
			}
		}
		if !reused {
			exit, out, err = e.Runner.Run(ctx, c.Cmd)
		}
		expect := c.ExpectExit
		condFailed := err != nil || exit != expect
		if condFailed {
			fc := FailedCond{Cmd: c.Cmd, Exit: exit, Tail: tail(out)}
			if err != nil {
				fc.Tail = tail(out + " (" + err.Error() + ")")
			}
			failed = append(failed, fc)
		}
	}

	blockNeeded := len(failed) > 0

	// Step 6: Tier 2 — only when all mechanical pass AND ≥1 model condition.
	// The hook surfaces the model claim in the block reason for orchestrator
	// self-evaluation (settled Option B); stop-goal runs no model call.
	if !blockNeeded && g.HasModelConditions() {
		blockNeeded = true // model claim pending orchestrator evaluation
	}

	// Step 7: all conditions satisfied → no block, status satisfied.
	if !blockNeeded {
		// All mechanical conditions passed and no model conditions exist.
		g.Status = StatusSatisfied
		e.appendProgress(g, "all conditions satisfied")
		return Verdict{SnapshotAttribution: attributions}, false
	}

	// A block is needed. Build the failed-condition detail (rides the plain
	// block `reason` in autonomous mode; rides `failed_conditions` in the
	// semi-autonomous checkpoint — REQ-GLE-010 ↔ REQ-GLE-028 reconciliation).
	detail := e.blockReason(g, failed, hasMechanical)

	// Step 8 (D8): semi-autonomous checkpoint branch. The checkpoint `reason`
	// is the checkpoint prefix; the mechanical detail rides `failed_conditions`
	// so the orchestrator's confirm AskUserQuestion surfaces WHY the goal is not
	// satisfied without polluting the reason label.
	if g.ProgressionMode == ProgressionSemiAutonomous {
		e.appendProgress(g, "semi-autonomous checkpoint")
		v := Verdict{
			Decision:            "block",
			Reason:              fmt.Sprintf("semi-autonomous checkpoint: orchestrator to confirm continuation (turn %d of %d)", g.TurnsUsed, g.Ceiling.MaxTurns),
			Mode:                string(ProgressionSemiAutonomous),
			Turn:                g.TurnsUsed,
			Ceiling:             g.Ceiling.MaxTurns,
			LastProgress:        lastProgress(g),
			SnapshotAttribution: attributions,
		}
		if len(failed) > 0 {
			v.FailedConditions = failed
		}
		return v, true
	}

	// Autonomous mode: plain block — the detail IS the reason (REQ-GLE-010).
	e.appendProgress(g, detail)
	return Verdict{Decision: "block", Reason: detail, SnapshotAttribution: attributions}, true
}

// blockReason builds the failed-condition + tail detail that rides the block.
// In autonomous mode this is the plain `reason`; in semi-autonomous mode the
// same detail also populates `failed_conditions` (REQ-GLE-010 ↔ REQ-GLE-028).
func (e *Eval) blockReason(g *Goal, failed []FailedCond, hasMechanical bool) string {
	if len(failed) > 0 {
		var parts []string
		for _, f := range failed {
			parts = append(parts, fmt.Sprintf("cmd %q exited %d (want 0): %s", f.Cmd, f.Exit, f.Tail))
		}
		return "mechanical condition failed: " + strings.Join(parts, "; ")
	}
	// No mechanical failure → the block is the Tier-2 model-claim gate.
	var claims []string
	for _, c := range g.Conditions {
		if c.Type == ConditionModel {
			claims = append(claims, c.Claim)
		}
	}
	return "model claim pending orchestrator evaluation: " + strings.Join(claims, "; ")
}

// ceilingReport assembles the 5-section verdict for a ceiling-exit or
// stagnation halt.
func (e *Eval) ceilingReport(g *Goal, cause string) *CeilingVerdict {
	return &CeilingVerdict{
		Claim:               fmt.Sprintf("goal %q did not converge within %d turns (%s)", g.Goal, g.Ceiling.MaxTurns, cause),
		Evidence:            fmt.Sprintf("turns_used=%d ceiling=%d conditions=%d last_progress=%q", g.TurnsUsed, g.Ceiling.MaxTurns, len(g.Conditions), lastProgress(g)),
		BaselineAttribution: "measured against this session's goal state (.moai/state/goal/<session>.json)",
		Gaps:                "the orchestrator did not surface evidence that all conditions hold; remaining conditions unverified",
		ResidualRisk:        "stagnation or ceiling may indicate a semantic failure requiring E1/E3 escalation rather than further turns",
	}
}

func (e *Eval) appendProgress(g *Goal, note string) {
	g.Progress = append(g.Progress, ProgressEntry{Turn: g.TurnsUsed, Note: note})
}
