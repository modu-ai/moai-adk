package cli

// codex_stop_chain.go — the Codex Stop chain (SPEC-DUAL-HARNESS-HOOK-PARITY-001
// M2d; design §D2 option B, §D3.3–§D3.8). Codex registers one Stop handler,
// `moai hook stop --harness codex`; it runs the eight Claude Stop members in
// Claude order inside that one handler and merges their decisions with one
// rule MoAI owns. Each member is the Claude member's own Go entry (or its Go
// port, for the sync gate), evaluated so that the same input yields the same
// normalized decision as on Claude (REQ-HPR-002). Members whose work cannot fit
// the handler timeout read a receipt instead (REQ-HPR-018, REQ-HPR-019).
//
// The code lives in internal/cli, not internal/hook
// (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/hook/security"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// Continuation reason classes (design §D3.4). A plain allow carries none.
const (
	reasonUnmet               = "unmet"
	reasonUnmeasured          = "unmeasured"
	reasonGateFailed          = "gate_failed"
	reasonUnverified          = "unverified"
	reasonFactoryContinuation = "factory_continuation"
)

// Record statuses for members that end without a reason class. A member's
// status is its reason class when it has one; no status below reads as a
// pass except stopStatusPass, and only an evaluated gate or goal earns it.
const (
	stopStatusPass          = "pass"           // gate or goal evaluated and passed
	stopStatusNotApplicable = "not-applicable" // disabled, or a self-gate did not hold
	stopStatusOK            = "ok"             // advisory member ran without a failure
	stopStatusFailed        = "failed"         // advisory member failed or was cut off
	stopStatusFailOpen      = "fail-open"      // reviewer or result missing: allowed, recorded
	stopStatusAdvisoryFail  = "advisory-fail"  // gate failed under an advisory mode
	stopStatusTerminated    = "terminated"     // goal loop ended by its budget (ceiling, wall clock, stagnation) or proved unsatisfiable
	stopStatusCancelled     = "cancelled"      // goal cancelled by the user (Codex Interrupt)
)

// Gate ids for the §D3.8 cap counter.
const (
	stopCapGateSync   = "sync-gate"
	stopCapGateReview = "codex-review"
)

// stopChainRecordDir holds the per-session verdict record of the last run.
const stopChainRecordDir = ".moai/state/codex-stop-chain"

// errAdvisorySkipped marks an advisory member whose own gate did not hold.
var errAdvisorySkipped = errors.New("advisory member not applicable")

// stopMemberOutcome is one member's result on one Codex Stop.
type stopMemberOutcome struct {
	Number      int                   `json:"number"`
	Name        string                `json:"name"`
	Decision    codexadapter.Decision `json:"decision"`
	Class       string                `json:"class,omitempty"`
	Status      string                `json:"status"`
	Reason      string                `json:"reason,omitempty"`
	ReceiptRead bool                  `json:"receipt_read,omitempty"`
	ElapsedMS   int64                 `json:"elapsed_ms"`
	Err         string                `json:"error,omitempty"`
	// Advisory is text for the output's systemMessage; it never decides.
	Advisory string `json:"-"`
	// Discards are written to the adapter's sink by run, so an allow that
	// skipped a check is visible rather than silent.
	Discards []codexadapter.Discard `json:"-"`
}

// codexStopChainResult is one run of the chain.
type codexStopChainResult struct {
	Members []stopMemberOutcome
	// Output is the merged hook output. With no member contributing beyond
	// member 1 it is member 1's own output, unchanged.
	Output *hook.HookOutput
	// Fault is member 1's dispatch error. The caller answers it fail-closed
	// (M2c, REQ-HPR-009) exactly as it answered a dispatch error before.
	Fault error
}

// codexStopChain runs the members for one Stop payload.
type codexStopChain struct {
	root  string
	input *hook.HookInput
	// member1 is `moai hook stop` (the registry's Stop handlers).
	member1 func(context.Context, *hook.HookInput) (*hook.HookOutput, error)
	// advisory holds members 4, 5 and 8; a member returns its message.
	advisory map[int]func(context.Context) (string, error)
	// budgetFor is a member's internal deadline: the declared budget table
	// (codexwiring.StopChainMembers). The decision goldens widen it so a
	// loaded machine cannot turn a decision test into a timing test; the
	// AC-HPR-016 timing leg measures the declared budgets.
	budgetFor func(n int) time.Duration

	keyMu  sync.Mutex
	key    string
	keyErr error
	keyOK  bool

	capMu sync.Mutex
}

func newCodexStopChain(root string, input *hook.HookInput) *codexStopChain {
	if input == nil {
		input = &hook.HookInput{HookEventName: string(hook.EventStop)}
	}
	c := &codexStopChain{root: root, input: input, budgetFor: func(n int) time.Duration { return stopMemberSpec(n).Budget }}
	c.member1 = func(ctx context.Context, in *hook.HookInput) (*hook.HookOutput, error) {
		if deps == nil || deps.HookRegistry == nil {
			return nil, errors.New("hook system not initialized")
		}
		return deps.HookRegistry.Dispatch(ctx, hook.EventStop, in)
	}
	c.advisory = map[int]func(context.Context) (string, error){
		4: func(context.Context) (string, error) {
			return c.runGuardian(security.HandleSecurityTurn)
		},
		5: func(context.Context) (string, error) {
			return c.runGuardian(security.HandleSecurityCommit)
		},
		8: func(context.Context) (string, error) {
			if !isHookOptInEnabled(c.root) || !isHarnessLearningEnabled(c.root) {
				return "", errAdvisorySkipped
			}
			var errOut bytes.Buffer
			harnessObserveStop(c.root, c.input, &errOut)
			_, _ = os.Stderr.Write(errOut.Bytes())
			if strings.Contains(errOut.String(), "failed") {
				return "", errors.New(strings.TrimSpace(errOut.String()))
			}
			return "", nil
		},
	}
	return c
}

// resolveCodexStopRoot picks the project root the chain acts on: the
// dispatcher's own resolution (CLAUDE_PROJECT_DIR) first, then the payload's
// cwd, then the process working directory.
func resolveCodexStopRoot(input *hook.HookInput) string {
	if root := os.Getenv(config.EnvClaudeProjectDir); root != "" {
		return root
	}
	if input != nil && input.CWD != "" {
		return input.CWD
	}
	return resolveHookProjectRoot()
}

func stopMemberSpec(n int) codexwiring.StopMember {
	for _, m := range codexwiring.StopChainMembers {
		if m.Number == n {
			return m
		}
	}
	return codexwiring.StopMember{Number: n}
}

// stopChainDeadline is the runner deadline, T_stop − chain_overhead, which
// the budget table declares equal to the member sum (design §D3.5).
func (c *codexStopChain) stopChainDeadline() time.Duration {
	var d time.Duration
	for _, m := range codexwiring.StopChainMembers {
		d += c.budgetFor(m.Number) + m.UncutBudget
	}
	return d
}

// currentKey computes verify.Key once per run; every member binds to the same
// tree state.
func (c *codexStopChain) currentKey(ctx context.Context) (string, error) {
	c.keyMu.Lock()
	defer c.keyMu.Unlock()
	if !c.keyOK {
		c.key, c.keyErr = verify.Key(ctx, c.root)
		c.keyOK = c.keyErr == nil
	}
	return c.key, c.keyErr
}

func (c *codexStopChain) sessionID() string {
	s := filepath.Base(strings.TrimSpace(c.input.SessionID))
	if s == "" || s == "." || s == string(filepath.Separator) {
		return "no-session"
	}
	return s
}

// @MX:ANCHOR: [AUTO] Codex Stop chain runner — the one place the Codex Stop decision is made (design §D2 merge rule)
// @MX:REASON: runHookEvent, the AC-HPR-002/003/005 goldens, and the AC-HPR-016 timing leg call it; a member that allows where Claude continues, or a merge that drops a deny, loosens every Codex turn end

// run evaluates every member in Claude order and merges the decisions.
// Merge rule (design §D2): a cancelled goal already yields no block inside the
// goal member, so cancellation is first by construction; then any deny wins
// (a continuation is a Stop deny), carrying every denying member's reason in
// member order; otherwise the stop is allowed. Advisory text and the notes of
// an allow that skipped a check ride systemMessage and the discard sink.
func (c *codexStopChain) run(ctx context.Context) codexStopChainResult {
	ctx, cancel := context.WithTimeout(ctx, c.stopChainDeadline())
	defer cancel()

	var res codexStopChainResult
	start := time.Now()
	out1, err := c.member1(ctx, c.input)
	m1 := stopMemberOutcome{Number: 1, Name: stopMemberSpec(1).Name, Decision: codexadapter.DecisionAllow,
		Status: stopStatusOK, ElapsedMS: time.Since(start).Milliseconds()}
	if err != nil {
		res.Fault = err
		m1.Status, m1.Err = stopStatusFailed, err.Error()
		res.Members = append(res.Members, m1)
		c.writeRecord(res.Members)
		return res
	}
	if out1 != nil && out1.Decision == hook.DecisionBlock {
		m1.Decision, m1.Class, m1.Status, m1.Reason = codexadapter.DecisionDeny, reasonFactoryContinuation, reasonFactoryContinuation, out1.Reason
	}
	res.Members = append(res.Members, m1)

	res.Members = append(res.Members,
		c.budgeted(ctx, 2, c.syncGateMember, c.cutOffUnmeasured(2, stopCapGateSync, codexwiring.SyncGateReceiptCommand)),
		c.budgeted(ctx, 3, c.goalMember, c.cutOffUnmeasured(3, "", codexwiring.GoalReceiptCommand)),
		c.advisoryMember(ctx, 4),
		c.advisoryMember(ctx, 5),
		c.budgeted(ctx, 6, c.codexReviewMember, c.cutOffUnmeasured(6, stopCapGateReview, codexwiring.CodexReviewReceiptCommand)),
		c.budgeted(ctx, 7, c.multiReviewMember, c.multiCutOff),
		c.advisoryMember(ctx, 8),
	)

	res.Output = c.merge(out1, res.Members)
	var discards []codexadapter.Discard
	for _, m := range res.Members {
		discards = append(discards, m.Discards...)
	}
	if rerr := codexadapter.RecordDiscards(c.root, discards, false, os.Stderr); rerr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: record discards: %v\n", rerr)
	}
	c.writeRecord(res.Members)
	return res
}

// merge folds the member outcomes into one hook output.
func (c *codexStopChain) merge(out1 *hook.HookOutput, members []stopMemberOutcome) *hook.HookOutput {
	var reasons, notes []string
	for _, m := range members[1:] {
		if m.Decision == codexadapter.DecisionDeny {
			reasons = append(reasons, m.Reason)
		}
		if m.Advisory != "" {
			notes = append(notes, m.Advisory)
		}
	}
	if len(reasons) == 0 && len(notes) == 0 {
		return out1
	}
	out := &hook.HookOutput{}
	if out1 != nil {
		cp := *out1
		out = &cp
	}
	if len(reasons) > 0 {
		if out.Decision == hook.DecisionBlock && out.Reason != "" {
			reasons = append([]string{out.Reason}, reasons...)
		}
		out.Decision = hook.DecisionBlock
		out.Reason = strings.Join(reasons, "\n\n")
	}
	if len(notes) > 0 {
		if out.SystemMessage != "" {
			notes = append([]string{out.SystemMessage}, notes...)
		}
		out.SystemMessage = strings.Join(notes, "\n")
	}
	return out
}

// budgeted runs a member under its declared internal budget. A member that
// does not return in time is replaced by cutOff's outcome and keeps running
// in the background; the handler process ends shortly after, which bounds it.
func (c *codexStopChain) budgeted(ctx context.Context, n int, fn func(context.Context) stopMemberOutcome, cutOff func(context.Context) stopMemberOutcome) stopMemberOutcome {
	spec := stopMemberSpec(n)
	budget := c.budgetFor(n)
	mctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	start := time.Now()
	done := make(chan stopMemberOutcome, 1)
	go func() { done <- fn(mctx) }()
	var o stopMemberOutcome
	select {
	case o = <-done:
	case <-mctx.Done():
		o = cutOff(ctx)
		o.Err = "cut off at its internal budget " + budget.String()
	}
	o.Number, o.Name = n, spec.Name
	o.ElapsedMS = time.Since(start).Milliseconds()
	if o.Status == "" {
		o.Status = o.Class
	}
	return o
}

// cutOffUnmeasured is the cut-off outcome of a goal or required gate:
// unmeasured, never an allow (design §D3.5). For a capped gate the cut-off
// counts toward the §D3.8 cap like any other unmeasured continuation.
func (c *codexStopChain) cutOffUnmeasured(n int, capGate, command string) func(context.Context) stopMemberOutcome {
	return func(ctx context.Context) stopMemberOutcome {
		why := "the member did not finish within its internal budget"
		if capGate != "" {
			// Without a tree key the cut-offs still count, under one stable
			// placeholder key, so a member that is always slow is bounded too.
			key, ok := c.cachedKey()
			if !ok {
				key = "tree-unread"
			}
			return c.unmeasured(ctx, n, capGate, key, command, why)
		}
		return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonUnmeasured,
			Reason: fmt.Sprintf("%s: not measured on this tree (%s). Run `%s`, then end the turn again.", stopMemberSpec(n).Name, why, command)}
	}
}

// multiCutOff: member 7 cut off reads as result-missing — it allows, with the
// discard record (design §D3.5).
func (c *codexStopChain) multiCutOff(context.Context) stopMemberOutcome {
	return c.multiMissing("the multi review gate did not finish within its internal budget")
}

// cachedKey reads the key without waiting: the cut-off path must not block on
// a key computation the abandoned member still holds.
func (c *codexStopChain) cachedKey() (string, bool) {
	if !c.keyMu.TryLock() {
		return "", false
	}
	defer c.keyMu.Unlock()
	return c.key, c.keyOK
}

// advisoryMember runs an advisory member under its budget. A failure or a
// cut-off is recorded as failed and never as passed; the decision is not
// affected (REQ-HPR-004).
func (c *codexStopChain) advisoryMember(ctx context.Context, n int) stopMemberOutcome {
	spec := stopMemberSpec(n)
	fn := c.advisory[n]
	o := stopMemberOutcome{Number: n, Name: spec.Name, Decision: codexadapter.DecisionAllow}
	if fn == nil {
		o.Status = stopStatusNotApplicable
		return o
	}
	budget := c.budgetFor(n)
	mctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	type result struct {
		msg string
		err error
	}
	start := time.Now()
	done := make(chan result, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- result{err: fmt.Errorf("panic: %v", r)}
			}
		}()
		msg, err := fn(mctx)
		done <- result{msg, err}
	}()
	select {
	case r := <-done:
		switch {
		case errors.Is(r.err, errAdvisorySkipped):
			o.Status = stopStatusNotApplicable
		case r.err != nil:
			o.Status, o.Err = stopStatusFailed, r.err.Error()
		default:
			o.Status, o.Advisory = stopStatusOK, r.msg
		}
	case <-mctx.Done():
		o.Status, o.Err = stopStatusFailed, "cut off at its internal budget "+budget.String()
	}
	o.ElapsedMS = time.Since(start).Milliseconds()
	if o.Status == stopStatusFailed {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: advisory member %d (%s) failed: %s\n", n, spec.Name, o.Err)
	}
	return o
}

// runGuardian runs one security guardian Stop handler and returns its
// advisory message. The guardian's opt-in block rides hookSpecificOutput,
// which a Stop decision does not read, so on both harnesses it is advisory.
func (c *codexStopChain) runGuardian(handler func([]string, io.Reader, io.Writer, string) error) (string, error) {
	payload, _ := json.Marshal(c.input)
	var out bytes.Buffer
	if err := handler(nil, bytes.NewReader(payload), &out, c.root); err != nil {
		return "", err
	}
	var parsed struct {
		SystemMessage string `json:"systemMessage"`
	}
	if out.Len() > 0 {
		if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
			return "", fmt.Errorf("unparseable guardian output: %w", err)
		}
	}
	return parsed.SystemMessage, nil
}

// ── member 3: goal (lookup-only) ────────────────────────────────────────

// lookupOnlyRunner is the Codex goal runner: it never executes a condition.
// A snapshot miss reaches it and is reported as not measured (design §D3.3).
type lookupOnlyRunner struct {
	mu     sync.Mutex
	missed []string
}

var errGoalNotMeasured = errors.New("not measured on this tree: no fresh receipt, and the Codex Stop chain does not execute conditions")

func (r *lookupOnlyRunner) Run(_ context.Context, cmd string) (int, string, error) {
	r.mu.Lock()
	r.missed = append(r.missed, cmd)
	r.mu.Unlock()
	return -1, "", errGoalNotMeasured
}

func (c *codexStopChain) goalMember(ctx context.Context) stopMemberOutcome {
	runner := &lookupOnlyRunner{}
	src := &verify.Source{ProjectRoot: c.root, KeyFunc: func(ctx context.Context, _ string) (string, error) {
		return c.currentKey(ctx)
	}}
	verdict, block, found := evaluateStopGoal(ctx, c.root, c.input.SessionID, runner, src, os.Stderr)
	switch {
	case !found:
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable}
	case !block:
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: c.goalAllowStatus()}
	}
	reason := verdict.Reason
	if reason == "" {
		reason = "goal: conditions not yet satisfied"
	}
	runner.mu.Lock()
	missed := append([]string(nil), runner.missed...)
	runner.mu.Unlock()
	if len(missed) == 0 {
		return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonUnmet, Reason: reason}
	}
	var steps []string
	for _, cmd := range missed {
		steps = append(steps, fmt.Sprintf("run `%s`, then `%s --check-id goal --command '%s' --exit <its exit code>`",
			cmd, codexwiring.GoalReceiptCommand, strings.ReplaceAll(cmd, "'", `'\''`)))
	}
	return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonUnmeasured,
		Reason: "goal: condition not measured on this tree. " + strings.Join(steps, "; ") + "; then end the turn again.\n" + reason}
}

// goalAllowStatus records why the goal member allowed, read from the goal
// state the evaluation just persisted. Only a satisfied goal records a pass: a
// cancelled goal, one ended by its budget, and one whose status is outside the
// vocabulary must never read as success in the verdict record
// (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2f, REQ-HPR-015/016/017).
func (c *codexStopChain) goalAllowStatus() string {
	g, err := goal.LoadGoal(c.root, c.input.SessionID)
	if err != nil || g == nil {
		return stopStatusNotApplicable
	}
	switch g.Status {
	case goal.StatusSatisfied:
		return stopStatusPass
	case goal.StatusCancelled:
		return stopStatusCancelled
	case goal.StatusCeilingExit, goal.StatusUnsatisfiable:
		return stopStatusTerminated
	default:
		return stopStatusNotApplicable
	}
}

// ── member 2: sync-phase quality gate (receipt, behind its self-gates) ──

func (c *codexStopChain) syncGateMember(ctx context.Context) stopMemberOutcome {
	applies, langs, why := syncGateApplies(ctx, c.root)
	if !applies {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable, Reason: why}
	}
	key, err := c.currentKey(ctx)
	if err != nil {
		return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonUnmeasured,
			Reason: fmt.Sprintf("sync-phase quality gate: the tree state could not be read (%v). Run `%s`, then end the turn again.", err, codexwiring.SyncGateReceiptCommand)}
	}
	state, checks := syncGateReceiptState(key, langs)
	chk := verify.CheckReceipt(verify.LoadReceipt(c.root, state), state, time.Now(), 0)
	if !chk.Run {
		o := c.unmeasured(ctx, 2, stopCapGateSync, key, codexwiring.SyncGateReceiptCommand, chk.Reason)
		o.ReceiptRead = true
		return o
	}
	c.capReset(ctx, stopCapGateSync)
	r := chk.Receipt
	if r.Verdict != "fail" {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusPass, ReceiptRead: true}
	}
	label := syncGateFailedLabel(checks, r.ExitCode)
	if syncGateMode(r.ExitCode&syncGateC1Failed != 0, r.ExitCode&syncGateC2Failed != 0) != "blocking" {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusAdvisoryFail, ReceiptRead: true,
			Advisory: "sync-phase quality gate WARNING (advisory, not blocking): " + label}
	}
	if c.input.StopHookActive {
		// Claude does not re-deliver a stored block on a flagged turn
		// (sync-phase-quality-gate.sh:484–486); the receipt is unchanged and
		// the next unflagged Stop blocks again.
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable, ReceiptRead: true,
			Reason: "stored failure not re-delivered on a stop_hook_active turn"}
	}
	return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonGateFailed, ReceiptRead: true,
		Reason: fmt.Sprintf("sync-phase quality gate BLOCKED: %s (receipt from `%s` for this tree). Fix it, run `%s` again, then end the turn.",
			label, codexwiring.SyncGateReceiptCommand, codexwiring.SyncGateReceiptCommand)}
}

// ── member 6: codex review gate (receipt, behind its self-gates) ─────────

func (c *codexStopChain) codexReviewMember(ctx context.Context) stopMemberOutcome {
	// Steps 1–3 in Claude's order (codex_review_gate.go:56–76).
	if !readCodexReviewGateEnabled(c.root) {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable}
	}
	if c.input.StopHookActive {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable, Reason: "stop_hook_active"}
	}
	if !reviewGateChangeDetector(c.root) {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable, Reason: "no reviewable change"}
	}
	// Step 4: no reviewer → allow, as on Claude, and never silently.
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		note := "codex review gate: the codex binary is not installed, so the stop was allowed without a codex review (fail-open, as on Claude)"
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusFailOpen, Reason: note, Advisory: note,
			Discards: []codexadapter.Discard{{Event: hook.EventStop, Key: "codex-stop-chain/codex-review/reviewer-missing", Reason: note}}}
	}
	key, err := c.currentKey(ctx)
	if err != nil {
		return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonUnmeasured,
			Reason: fmt.Sprintf("codex review gate: the tree state could not be read (%v). Run `%s`, then end the turn again.", err, codexwiring.CodexReviewReceiptCommand)}
	}
	state := codexReviewReceiptState(ctx, key, binaryPath)
	chk := verify.CheckReceipt(verify.LoadReceipt(c.root, state), state, time.Now(), 0)
	if !chk.Run {
		// Step 7: codex installed, receipt missing or stale → continue.
		o := c.unmeasured(ctx, 6, stopCapGateReview, key, codexwiring.CodexReviewReceiptCommand, chk.Reason)
		o.ReceiptRead = true
		return o
	}
	c.capReset(ctx, stopCapGateReview)
	if chk.Receipt.Verdict == codexReviewVerdictFail {
		return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonGateFailed, ReceiptRead: true,
			Reason: fmt.Sprintf("codex review gate: the codex review recorded for this tree failed. Run `%s` to see the findings, address them, and end the turn again.", codexwiring.CodexReviewReceiptCommand)}
	}
	return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusPass, ReceiptRead: true}
}

// ── member 7: multi review gate (fail-open-on-missing) ───────────────────

func (c *codexStopChain) multiReviewMember(context.Context) stopMemberOutcome {
	// The same decision order as HandleMultiReviewGate (multi_review_gate.go).
	if !readMultiReviewGateEnabled(c.root) || c.input.StopHookActive || !reviewGateChangeDetector(c.root) {
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusNotApplicable}
	}
	result, ok := loadConvergenceResult(c.root, c.input.SessionID)
	if !ok {
		return c.multiMissing("no multi-model review result is recorded for this session")
	}
	if result.OverallVerdict == overallVerdictFail {
		return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonGateFailed, Reason: blockReason(result)}
	}
	return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusPass}
}

func (c *codexStopChain) multiMissing(why string) stopMemberOutcome {
	note := "multi review gate: result missing (" + why + "); the stop was allowed (fail-open, as on Claude)"
	return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: stopStatusFailOpen, Reason: note, Advisory: note,
		Discards: []codexadapter.Discard{{Event: hook.EventStop, Key: "codex-stop-chain/multi-review/result-missing", Reason: note}}}
}

// ── §D3.8 consecutive-unmeasured cap (Codex-only) ─────────────────────────

type stopCapEntry struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type stopCapState struct {
	Gates map[string]stopCapEntry `json:"gates"`
}

func (c *codexStopChain) capPath() string {
	return filepath.Join(c.root, filepath.FromSlash(codexwiring.StopCapStateDir), c.sessionID()+".json")
}

func (c *codexStopChain) loadCap() stopCapState {
	var s stopCapState
	if b, err := os.ReadFile(c.capPath()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	if s.Gates == nil {
		s.Gates = map[string]stopCapEntry{}
	}
	return s
}

func (c *codexStopChain) saveCap(s stopCapState) {
	p := c.capPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: cap state: %v\n", err)
		return
	}
	b, _ := json.Marshal(s)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: cap state: %v\n", err)
		return
	}
	if err := os.Rename(tmp, p); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: cap state: %v\n", err)
	}
}

// capStep counts one more unmeasured continuation for gate on key. A key
// change resets the count first. It returns the count after the step, which
// never exceeds the cap. A member already cut off (ctx done) no longer owns
// the counter: its cut-off outcome counted instead, so it writes nothing.
func (c *codexStopChain) capStep(ctx context.Context, gate, key string) int {
	c.capMu.Lock()
	defer c.capMu.Unlock()
	s := c.loadCap()
	if ctx.Err() != nil {
		return s.Gates[gate].Count
	}
	e := s.Gates[gate]
	if e.Key != key {
		e = stopCapEntry{Key: key}
	}
	if e.Count < codexwiring.StopUnmeasuredCap {
		e.Count++
	}
	s.Gates[gate] = e
	c.saveCap(s)
	return e.Count
}

// capReset clears gate's count: a fresh receipt was read.
func (c *codexStopChain) capReset(ctx context.Context, gate string) {
	c.capMu.Lock()
	defer c.capMu.Unlock()
	if ctx.Err() != nil {
		return
	}
	s := c.loadCap()
	if _, ok := s.Gates[gate]; !ok {
		return
	}
	delete(s.Gates, gate)
	c.saveCap(s)
}

// @MX:WARN: [AUTO] the Nth consecutive unmeasured Stop allows without the check having run — the only Codex allow on a required gate that is not a verdict
// @MX:REASON: [AUTO] design §D3.8 bounds the continuation loop; the allow must stay paired with the unverified record and status, or a capped gate reads as passed

// unmeasured is the continuation of a required gate whose receipt is missing
// or stale, subject to the §D3.8 cap.
func (c *codexStopChain) unmeasured(ctx context.Context, n int, gate, key, command, why string) stopMemberOutcome {
	count := c.capStep(ctx, gate, key)
	name := stopMemberSpec(n).Name
	if count >= codexwiring.StopUnmeasuredCap {
		note := fmt.Sprintf("%s: unverified — %d consecutive Stops found no valid receipt for this tree (%s), and `%s` was never run for it. The stop is allowed to bound the loop; the gate is NOT passed.",
			name, count, why, command)
		return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Class: reasonUnverified, Reason: note, Advisory: note,
			Discards: []codexadapter.Discard{{Event: hook.EventStop, Key: "codex-stop-chain/" + gate + "/unverified", Reason: note}}}
	}
	return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Class: reasonUnmeasured,
		Reason: fmt.Sprintf("%s: not measured on this tree (%s). Run `%s`, then end the turn again. (continuation %d of %d before the stop is allowed as unverified)",
			name, why, command, count, codexwiring.StopUnmeasuredCap)}
}

// ── verdict record ───────────────────────────────────────────────────────

// codexStopChainRecord is the per-session record of the last run: one status
// per member. A capped gate reads unverified, a failed advisory member reads
// failed; only an evaluated gate or goal reads pass.
type codexStopChainRecord struct {
	SessionID  string              `json:"session_id"`
	RecordedAt time.Time           `json:"recorded_at"`
	Members    []stopMemberOutcome `json:"members"`
}

func (r *codexStopChainRecord) status(n int) string {
	for _, m := range r.Members {
		if m.Number == n {
			return m.Status
		}
	}
	return ""
}

func (c *codexStopChain) writeRecord(members []stopMemberOutcome) {
	rec := codexStopChainRecord{SessionID: c.sessionID(), RecordedAt: time.Now().UTC(), Members: members}
	p := filepath.Join(c.root, filepath.FromSlash(stopChainRecordDir), c.sessionID()+".json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: record: %v\n", err)
		return
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex stop chain: record: %v\n", err)
	}
}

func readCodexStopChainRecord(root, sessionID string) (*codexStopChainRecord, error) {
	c := &codexStopChain{root: root, input: &hook.HookInput{SessionID: sessionID}}
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(stopChainRecordDir), c.sessionID()+".json"))
	if err != nil {
		return nil, fmt.Errorf("read codex stop chain record: %w", err)
	}
	var rec codexStopChainRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return nil, fmt.Errorf("parse codex stop chain record: %w", err)
	}
	return &rec, nil
}

// renderCodexStop renders the merged output as the Codex Stop handler writes
// it (the adapter's mapping).
func renderCodexStop(res codexStopChainResult) ([]byte, error) {
	out := res.Output
	if out == nil {
		out = &hook.HookOutput{}
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("render codex stop: %w", err)
	}
	mapped, _, err := codexadapter.MapOutput(hook.EventStop, raw)
	if err != nil {
		return nil, fmt.Errorf("render codex stop: %w", err)
	}
	return mapped, nil
}
