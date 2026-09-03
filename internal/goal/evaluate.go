package goal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	Decision         string       `json:"decision,omitempty"`
	Reason           string       `json:"reason,omitempty"`
	Mode             string       `json:"mode,omitempty"`
	Turn             int          `json:"turn,omitempty"`
	Ceiling          int          `json:"ceiling,omitempty"`
	LastProgress     string       `json:"last_progress,omitempty"`
	FailedConditions []FailedCond `json:"failed_conditions,omitempty"`
	CeilingExit      bool         `json:"ceiling_exit,omitempty"`
	// WallClockExit is set when the wall-clock bound (Ceiling.MaxDuration) fires
	// (SPEC-INFINITE-GOAL-001 REQ-4 / OQ-2). The emitted 5-section Verdict is
	// indistinguishable in shape from a MaxTurns-ceiling verdict; this flag lets
	// callers/tests distinguish the cause.
	WallClockExit bool            `json:"wall_clock_exit,omitempty"`
	Stagnation    bool            `json:"stagnation,omitempty"`
	// Unsatisfiable is set when a mechanical condition proved unrunnable — the
	// shell reported exit 127 ("command not found") for a condition that did not
	// declare 127 as its expected status. Such a condition can never pass, so
	// the evaluator stops blocking instead of spending every remaining turn on
	// it; the shape distinguishes this from an ordinary not-yet-converged turn.
	Unsatisfiable bool `json:"unsatisfiable,omitempty"`
	Verdict       *CeilingVerdict `json:"verdict,omitempty"`
	Yielded       bool            `json:"yielded,omitempty"`
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
	// ProjectRoot, when non-empty, enables the bounded file-set SHA axis of the
	// strengthened stagnation guard (SPEC-INFINITE-GOAL-001 REQ-4 / D7). Empty
	// → the file-set axis is best-effort/no-op (constant), and stagnation keys
	// on per-condition exit + output only. The hook (hook_stop_goal.go) supplies
	// the real project root; unit tests may leave it empty.
	ProjectRoot string
	// Snapshot, when non-nil, is consulted before executing a Tier-1 mechanical
	// condition: an exact-match fresh entry reuses the recorded exit code
	// without a CmdRunner call. nil preserves the pre-existing behavior
	// unchanged (strictly additive contract).
	Snapshot SnapshotSource
}

// DefaultStagnationThreshold is the number of consecutive identical progress
// notes that trigger the stagnation guard (REQ-GLE-017).
const DefaultStagnationThreshold = 3

// shellCommandNotFound is the POSIX shell's exit status for an unresolvable
// command name. It is the language-independent signal that a condition's first
// word names nothing runnable.
const shellCommandNotFound = 127

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
// identical mechanical-condition signatures (N = StagnationThreshold).
//
// SPEC-INFINITE-GOAL-001 REQ-4 strengthens the guard: the comparison key is the
// per-turn Fingerprint (mechanical-condition fingerprint: exit + output-hash +
// bounded file-set SHA) when present; it falls back to the legacy Note when
// Fingerprint is empty (backward compat with pre-M4 / ceiling / satisfied
// entries). N consecutive identical signatures → stagnation.
func (e *Eval) isStagnant(g *Goal) bool {
	n := e.StagnationThreshold
	if n == 0 {
		n = DefaultStagnationThreshold
	}
	if len(g.Progress) < n {
		return false
	}
	last := stagnationKey(g.Progress[len(g.Progress)-1])
	for i := len(g.Progress) - 1; i >= len(g.Progress)-n; i-- {
		if stagnationKey(g.Progress[i]) != last {
			return false
		}
		if last == "" {
			// Empty signature (e.g. all-empty entries) must not collapse into a
			// vacuous "N identical empties" stagnation.
			return false
		}
	}
	return true
}

// stagnationKey returns the per-turn signature used by isStagnant: the
// Fingerprint when non-empty, else the legacy Note.
func stagnationKey(p ProgressEntry) string {
	if p.Fingerprint != "" {
		return p.Fingerprint
	}
	return p.Note
}

// condResult captures one mechanical condition's evaluation result for the
// fingerprint computation (SPEC-INFINITE-GOAL-001 REQ-4).
type condResult struct {
	cmd  string
	exit int
	out  string
}

// mechanicalFingerprint computes the per-turn mechanical-condition fingerprint
// (SPEC-INFINITE-GOAL-001 REQ-4 / D7). It keys on per-condition (exit,
// output-hash) plus a bounded file-set SHA derived from the condition commands'
// first path-like tokens (D7: "extract the first path-like token from each
// Condition.Cmd; empty set when none parses"). When no path-like token parses
// (the common shell-command case), the file-set axis is a constant — the load-
// bearing signal is (exit, output), which captures test-count + pass/fail
// tally implicitly (identical output ⟹ identical count/tally).
func mechanicalFingerprint(results []condResult, projectRoot string) string {
	var b strings.Builder
	for _, r := range results {
		b.WriteString(strconv.Itoa(r.exit))
		b.WriteByte(':')
		b.WriteString(shortHash(r.out))
		b.WriteByte('|')
	}
	b.WriteString("fs=")
	b.WriteString(boundedFileSetSHA(results, projectRoot))
	return b.String()
}

// boundedFileSetSHA derives the bounded file-set SHA (D7). For each condition
// command, extract the first path-like token (a whitespace-delimited token that
// contains a path separator '/' OR a file extension). Resolve each against
// projectRoot; hash the contents of the files that exist (sorted by path). When
// no path-like token resolves to an existing file (the common shell-command
// case), returns a constant "empty" — the file-set axis is best-effort/no-op.
//
// The set is BOUNDED by construction: it is at most one path per condition
// command (NOT the whole repo — AP-4). Empty when projectRoot is empty.
func boundedFileSetSHA(results []condResult, projectRoot string) string {
	if projectRoot == "" {
		return "empty"
	}
	var paths []string
	seen := map[string]bool{}
	for _, r := range results {
		tok := firstPathLikeToken(r.cmd)
		if tok == "" {
			continue
		}
		full := tok
		if !filepath.IsAbs(tok) {
			full = filepath.Join(projectRoot, tok)
		}
		if _, err := os.Stat(full); err != nil {
			continue // not an existing file → skip (best-effort)
		}
		if !seen[full] {
			seen[full] = true
			paths = append(paths, full)
		}
	}
	if len(paths) == 0 {
		return "empty"
	}
	for i := 1; i < len(paths); i++ {
		for j := i; j > 0 && paths[j-1] > paths[j]; j-- {
			paths[j-1], paths[j] = paths[j], paths[j-1]
		}
	}
	h := sha256.New()
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			h.Write([]byte(p + ":unreadable;"))
			continue
		}
		h.Write([]byte(p + ":"))
		h.Write(data)
		h.Write([]byte(";"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// firstPathLikeToken extracts the first whitespace-delimited token from cmd
// that looks like a file path: it contains a '/' separator OR matches a file
// extension pattern (dot followed by 1-4 alphanumerics at the token's end).
// Flags (leading '-') and the literal "./..." glob are excluded. Returns "" when
// no token parses.
func firstPathLikeToken(cmd string) string {
	for _, tok := range strings.Fields(cmd) {
		if tok == "" || strings.HasPrefix(tok, "-") {
			continue
		}
		if tok == "./..." || strings.HasSuffix(tok, "/...") {
			continue
		}
		if strings.Contains(tok, "/") {
			return tok
		}
		// extension pattern: name.ext(1-4)
		dot := strings.LastIndexByte(tok, '.')
		if dot > 0 && dot < len(tok)-1 && len(tok)-dot-1 <= 4 {
			return tok
		}
	}
	return ""
}

// shortHash returns the first 16 hex chars of the SHA-256 of s (compact,
// collision-safe for stagnation equality over small output strings).
func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
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

	// Step 2b (SPEC-INFINITE-GOAL-001 REQ-4 / OQ-2): wall-clock bound. When
	// Ceiling.MaxDuration > 0 and the elapsed wall-clock since CreatedAt exceeds
	// it, fire a 5-section verdict indistinguishable in shape from the MaxTurns
	// ceiling (WallClockExit flag distinguishes the cause). CreatedAt is RFC3339;
	// a missing/unparseable CreatedAt → no fire (fail-open, never blocks on a
	// clock parse error).
	if g.Ceiling.MaxDuration > 0 {
		if created, err := time.Parse(time.RFC3339, g.CreatedAt); err == nil {
			elapsed := int(time.Since(created).Seconds())
			if elapsed >= g.Ceiling.MaxDuration {
				g.Status = StatusCeilingExit
				v := Verdict{
					WallClockExit: true,
					Verdict: e.ceilingReport(g, fmt.Sprintf(
						"wall-clock bound reached: %ds elapsed >= %ds max-duration",
						elapsed, g.Ceiling.MaxDuration)),
				}
				e.appendProgress(g, "wallclock-exit")
				return v, false
			}
		}
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
	var unrunnable []FailedCond
	var attributions []string
	hasMechanical := false
	// SPEC-INFINITE-GOAL-001 REQ-4: collect per-condition (exit, output) so the
	// mechanical-condition fingerprint can be computed for the strengthened
	// stagnation guard (D7).
	var results []condResult
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
		results = append(results, condResult{cmd: c.Cmd, exit: exit, out: out})
		expect := c.ExpectExit
		condFailed := err != nil || exit != expect
		if condFailed {
			fc := FailedCond{Cmd: c.Cmd, Exit: exit, Tail: tail(out)}
			if err != nil {
				fc.Tail = tail(out + " (" + err.Error() + ")")
			}
			failed = append(failed, fc)
			// Exit 127 is the shell's "command not found": the condition's first
			// word names nothing runnable, so no number of further turns can make
			// it exit as expected. A condition that DECLARES 127 as its expected
			// status is a legitimate absence assertion and is excluded — it did
			// not reach this branch at all.
			if exit == shellCommandNotFound {
				unrunnable = append(unrunnable, fc)
			}
		}
	}
	// Unsatisfiable-by-construction: stop the loop rather than block on a
	// condition that can never pass. This is the backstop for goals armed before
	// the arm-time runnability gate landed, and for prose whose first word
	// happens to resolve at arm time.
	if len(unrunnable) > 0 {
		g.Status = StatusUnsatisfiable
		v := Verdict{
			Unsatisfiable:       true,
			FailedConditions:    unrunnable,
			Verdict:             e.unrunnableReport(g, unrunnable),
			SnapshotAttribution: attributions,
		}
		e.appendProgress(g, "unsatisfiable-exit")
		return v, false
	}
	// Compute the mechanical-condition fingerprint for this turn (D7). The
	// fingerprint keys on per-condition (exit, output-hash) + a bounded file-set
	// SHA derived from the condition commands' first path-like tokens. When no
	// path-like token parses (the common case for shell commands like `false`),
	// the file-set axis is a constant (best-effort/no-op per D7) — the load-
	// bearing signal is (exit, output), which captures test-count + pass/fail
	// tally implicitly (identical output ⟹ identical count/tally).
	currentFP := mechanicalFingerprint(results, e.ProjectRoot)

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
		e.appendProgressFP(g, "all conditions satisfied", currentFP)
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
		e.appendProgressFP(g, "semi-autonomous checkpoint", currentFP)
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
	e.appendProgressFP(g, detail, currentFP)
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

// unrunnableReport assembles the 5-section verdict for an unsatisfiable-by-
// construction halt. It names the likely cause rather than only the symptom:
// exit 127 on a condition that reads as prose almost always means a natural-
// language claim was classified mechanical, and the remedy is the `model:`
// declaration prefix.
func (e *Eval) unrunnableReport(g *Goal, unrunnable []FailedCond) *CeilingVerdict {
	var cmds []string
	for _, f := range unrunnable {
		cmds = append(cmds, fmt.Sprintf("%q", f.Cmd))
	}
	joined := strings.Join(cmds, ", ")
	return &CeilingVerdict{
		Claim: fmt.Sprintf(
			"goal %q is unsatisfiable as declared: %d mechanical condition(s) exited 127 (command not found)",
			g.Goal, len(unrunnable)),
		Evidence: fmt.Sprintf(
			"turn %d of %d; condition(s) %s were run as shell commands and the shell resolved no such command (exit 127)",
			g.TurnsUsed, g.Ceiling.MaxTurns, joined),
		BaselineAttribution: "measured against this session's goal state (.moai/state/goal/<session>.json) and the exit codes observed this turn",
		Gaps: "whether the condition was ever meant as a command was NOT determined — a condition reading as prose was probably a claim about the " +
			"conversation, misclassified into the mechanical tier. Re-arm it with the model: prefix (moai goal \"model: <claim>\") to declare the tier explicitly.",
		ResidualRisk: "a genuine command that is merely absent from this environment produces the same exit 127; if the command was intended, install it " +
			"or re-arm with the cmd: prefix once it resolves. No condition was evaluated on its merits this turn.",
	}
}

func (e *Eval) appendProgress(g *Goal, note string) {
	g.Progress = append(g.Progress, ProgressEntry{Turn: g.TurnsUsed, Note: note})
}

// appendProgressFP appends a Progress entry carrying both a note and the
// mechanical-condition fingerprint for the strengthened stagnation guard
// (SPEC-INFINITE-GOAL-001 REQ-4).
func (e *Eval) appendProgressFP(g *Goal, note, fingerprint string) {
	g.Progress = append(g.Progress, ProgressEntry{
		Turn:        g.TurnsUsed,
		Note:        note,
		Fingerprint: fingerprint,
	})
}
