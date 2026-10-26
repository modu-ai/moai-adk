package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factory_run_select.go — the interactive run chooser of the lane-join gate
// (SPEC-LANE-JOIN-RUN-SELECT-001). When two or more active factory runs
// survive reconciliation and the session is interactive, the gate shows each
// run's distinguishing facts and lets the operator join one, clean one up and
// join the remainder, or cancel. Off a terminal the gate keeps its
// fail-closed refusal and this file is not consulted.

// Korean catalogue (native written register). Tests assert on these
// constants, so a wording change is one edit.
const (
	laneMsgHeaderMany         = "활성 factory run이 2개 이상입니다. 합류할 run을 고르세요."
	laneMsgHeaderOne          = "활성 factory run이 1개 남았습니다. 합류할 run을 고르거나 취소하세요."
	laneMsgLegendMany         = "번호: 해당 run에 합류 / r번호: 정리한 뒤 남은 run에 합류 / q 또는 빈 줄: 취소"
	laneMsgLegendOne          = "번호: 해당 run에 합류 / q 또는 빈 줄: 취소"
	laneMsgIneligibleFmt      = "선택 r%d: 정리할 수 없습니다. 리더가 종료되고 진행 중 카드가 없는 run만 정리할 수 있습니다."
	laneMsgInvalid            = "입력을 이해하지 못했습니다. 번호, r번호, q 중 하나를 입력하세요."
	laneMsgLegendMerge        = "m번호: 진행 중 카드를 살아 있는 run으로 옮겨 합치고 정리한 뒤 합류"
	laneMsgIneligibleMergeFmt = "선택 m%d: 합칠 수 없습니다. 리더가 종료되고 진행 중 카드가 있으며, 리더가 살아 있는 활성 run이 정확히 1개일 때만 합칠 수 있습니다."
	laneMsgMergeFailFmt       = "run %s 합치기에 실패했습니다 (%s). 저장소에서 다시 읽은 최신 상태로 목록을 보여 드립니다."
	laneMsgMergedFmt          = "run %s의 진행 중 카드를 run %s로 옮기고 정리했습니다. 옮긴 카드: %s / 포기 처리한 카드: %s"
	laneMsgCleanupFailFmt     = "run %s 정리에 실패했습니다 (%s). 저장소에서 다시 읽은 최신 상태로 목록을 보여 드립니다."
	laneMsgCancelled          = "합류를 취소했습니다. 이 선택으로 추가된 변경은 없습니다."
	laneLivenessLive          = "live (생존)"
	laneLivenessDead          = "dead (종료)"
	laneLivenessIndetermin    = "indeterminate (판정 불가)"
	laneEmptyCell             = "-"
	laneColSep                = " | "
	lanePrompt                = "선택> "
)

// laneColumnLabels are the table's column labels, one per rendered cell.
func laneColumnLabels() []string {
	return []string{"번호", "run id", "리더 생존", "리더 이름", "시작 (KST)", "마지막 레인 활동 (KST)", "진행 중 카드"}
}

// runChoiceKind is what the operator chose. The zero value is cancel, so a
// choice nobody filled in can never join or retire anything.
type runChoiceKind int

const (
	choiceCancel runChoiceKind = iota
	choiceJoin
	choiceCleanupThenJoin
	choiceMergeThenJoin
)

type runChoice struct {
	Kind  runChoiceKind
	RunID string
}

// laneRunRow is one listed run as DISPLAY CELLS; the offered choices ride
// beside them. Times are FormatKST text, "-" when empty.
type laneRunRow struct {
	RunID, Liveness, LeaderName, Start, LastActivity string
	InProgress                                       int
	CanCleanup                                       bool
}

// buildLaneRows is the one pure function that turns read-path facts into rows
// and decides which rows may offer cleanup. Cleanup needs a dead owner and no
// non-terminal card, and at least one other run to join afterwards — a single
// listed run has no remainder, so a one-row table never offers it
// (REQ-LJR-006). RetireRunIfDead re-enforces owner liveness whatever a
// chooser returns.
func buildLaneRows(facts []homestate.RunFacts, names map[int]string) []laneRunRow {
	rows := make([]laneRunRow, 0, len(facts))
	for _, f := range facts {
		rows = append(rows, laneRunRow{
			RunID:        f.RunID,
			Liveness:     laneLiveness(f.Classification),
			LeaderName:   cellOrDash(names[f.LeadPID]),
			Start:        cellOrDash(factorymsg.FormatKST(f.CreatedAt)),
			LastActivity: cellOrDash(factorymsg.FormatKST(f.LastActivity)),
			InProgress:   f.InProgress,
			CanCleanup:   f.Classification == homestate.OwnerDead && f.InProgress == 0 && len(facts) >= 2,
		})
	}
	return rows
}

// laneMergeEligible reports whether row i may be merged (card t1963,
// REQ-FCR-021): its owner is dead, it still holds in-progress cards, and
// exactly one listed run has a live owner — that run is the merge target, so
// with none or several live runs there is no single place to move the cards.
// It reads the display cells the rows already carry, so the row type and the
// existing `r<n>` eligibility are untouched.
func laneMergeEligible(rows []laneRunRow, i int) bool {
	if i < 0 || i >= len(rows) {
		return false
	}
	if rows[i].Liveness != laneLivenessDead || rows[i].InProgress == 0 {
		return false
	}
	live := 0
	for _, r := range rows {
		if r.Liveness == laneLivenessLive {
			live++
		}
	}
	return live == 1
}

func laneLiveness(c homestate.OwnerClassification) string {
	switch c {
	case homestate.OwnerLive:
		return laneLivenessLive
	case homestate.OwnerDead:
		return laneLivenessDead
	case homestate.OwnerIndeterminate:
		return laneLivenessIndetermin
	}
	return string(c)
}

func cellOrDash(s string) string {
	if s == "" {
		return laneEmptyCell
	}
	return s
}

// renderLaneTable emits one line per run: seven cells (the 1-based number,
// run id, liveness, leader name, start, last activity, count) joined by
// laneColSep. A test splits a line on the separator and compares each cell by
// equality, so a dropped column changes the cell count.
func renderLaneTable(rows []laneRunRow) string {
	var b strings.Builder
	for i, r := range rows {
		cells := []string{strconv.Itoa(i + 1), r.RunID, r.Liveness, r.LeaderName, r.Start, r.LastActivity, strconv.Itoa(r.InProgress)}
		b.WriteString(strings.Join(cells, laneColSep))
		b.WriteByte('\n')
	}
	return b.String()
}

// runChooser is the operator-interaction seam: tests inject a scripted double,
// production uses textRunChooser.
type runChooser interface {
	ChooseLaneRun(rows []laneRunRow) (runChoice, error)
}

// textRunChooser prompts on out and reads answers from in. Grammar: <n> joins
// listed run n; r<n> cleans run n up and joins the remainder (accepted only
// when row n offers it); q, an empty line, or input that ends before a
// complete line cancels. There is no default: an empty line is cancel.
type textRunChooser struct {
	in  io.Reader
	out io.Writer
}

func (c textRunChooser) ChooseLaneRun(rows []laneRunRow) (runChoice, error) {
	if err := c.render(rows); err != nil {
		return runChoice{}, err
	}
	for {
		if _, err := io.WriteString(c.out, lanePrompt); err != nil {
			return runChoice{}, err
		}
		line, complete, err := readLineBytewise(c.in)
		if err != nil {
			return runChoice{}, err
		}
		if !complete {
			// Input ended before a valid choice: cancel, not a hang.
			return runChoice{Kind: choiceCancel}, nil
		}
		choice, reason := parseLaneChoice(line, rows)
		if reason == "" {
			return choice, nil
		}
		if _, err := fmt.Fprintln(c.out, reason); err != nil {
			return runChoice{}, err
		}
	}
}

func (c textRunChooser) render(rows []laneRunRow) error {
	header, legend := laneMsgHeaderMany, laneMsgLegendMany
	if len(rows) == 1 {
		header, legend = laneMsgHeaderOne, laneMsgLegendOne
	}
	for i := range rows {
		if laneMergeEligible(rows, i) {
			legend += "\n" + laneMsgLegendMerge
			break
		}
	}
	_, err := fmt.Fprintf(c.out, "%s\n%s\n%s%s\n", header, strings.Join(laneColumnLabels(), laneColSep), renderLaneTable(rows), legend)
	return err
}

// readLineBytewise reads one line a byte at a time and stops at the newline,
// so nothing past it is consumed: the launcher hands the terminal to the
// child session next, and bytes a buffered reader over-read would be that
// session's first keystrokes. complete is false when the input ended before a
// newline (a partial line is not a choice).
func readLineBytewise(r io.Reader) (line string, complete bool, err error) {
	var sb strings.Builder
	buf := make([]byte, 1)
	for {
		n, rerr := r.Read(buf)
		if n == 1 {
			if buf[0] == '\n' {
				return sb.String(), true, nil
			}
			sb.WriteByte(buf[0])
		}
		if rerr == io.EOF {
			return sb.String(), false, nil
		}
		if rerr != nil {
			return "", false, rerr
		}
	}
}

// parseLaneChoice maps one entered line to a choice. A non-empty reason means
// the entry was rejected and the prompt asks again.
func parseLaneChoice(line string, rows []laneRunRow) (runChoice, string) {
	entry := strings.ToLower(strings.TrimSpace(line))
	if entry == "" || entry == "q" {
		return runChoice{Kind: choiceCancel}, ""
	}
	kind, digits := choiceJoin, entry
	switch {
	case strings.HasPrefix(entry, "r"):
		kind, digits = choiceCleanupThenJoin, strings.TrimPrefix(entry, "r")
	case strings.HasPrefix(entry, "m"):
		kind, digits = choiceMergeThenJoin, strings.TrimPrefix(entry, "m")
	}
	n, ok := parseRowNumber(digits, len(rows))
	if !ok {
		return runChoice{}, laneMsgInvalid
	}
	row := rows[n-1]
	switch kind {
	case choiceCleanupThenJoin:
		if !row.CanCleanup {
			return runChoice{}, fmt.Sprintf(laneMsgIneligibleFmt, n)
		}
	case choiceMergeThenJoin:
		if !laneMergeEligible(rows, n-1) {
			return runChoice{}, fmt.Sprintf(laneMsgIneligibleMergeFmt, n)
		}
	}
	return runChoice{Kind: kind, RunID: row.RunID}, ""
}

// parseRowNumber accepts ASCII digits only (no sign, no padding) naming a
// listed row.
func parseRowNumber(s string, rows int) (int, bool) {
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > rows {
		return 0, false
	}
	return n, true
}

// runSelectionInteractive reports whether the session can answer the chooser:
// stdin is a terminal and the prompt stream is a terminal-backed file.
// TestMain pins it false so no test ever reads a real terminal.
var runSelectionInteractive = func() bool { return isInteractiveStdin() && writerIsTerminal(os.Stderr) }

// activeRunChooser is the chooser the gate uses; tests swap it for a scripted
// double.
var activeRunChooser runChooser = textRunChooser{in: os.Stdin, out: os.Stderr}

// laneRunNotices receives the loop's own one-line notices (a refused cleanup,
// the cancel confirmation); tests capture it.
var laneRunNotices io.Writer = os.Stderr

// chooseLaneRun is the interactive answer to a lane join that two or more
// active runs make ambiguous (SPEC-LANE-JOIN-RUN-SELECT-001). It is a loop of
// presentations, and it never joins by side effect:
//
//   - every presentation reads the store through the read path
//     (readLaneRunFacts): no reconciler, no write, and neither ResolveActiveRun
//     nor enterSelectedFactoryRun is called to obtain facts, so the
//     factory-run variable is unset at every presentation entry;
//   - enterSelectedFactoryRun is called at exactly two places, both with an
//     explicit run id: the operator's join choice, and the join half of a
//     successful cleanup-then-join once exactly one run remains.
//
// @MX:WARN: [AUTO] this loop decides which run a lane joins and whether a run is retired; the factory-run variable must stay unset until the operator names the run
// @MX:REASON: re-resolving with an empty selector (the predecessor's defect N2) joins the lone survivor by side effect, so a refused cleanup would silently join a run the operator never chose
//
// The loop ends at the operator's cancel, at a join, at a successful cleanup
// that leaves one run, or when no run is active: every pass after the first
// needs a fresh operator entry (end of input cancels), a successful
// retirement strictly shrinks the active set, and a refused attempt writes
// nothing.
func chooseLaneRun(root string, timing *factoryLaunchTiming) (func(), error) {
	afterCleanup := false
	for {
		facts, names, err := readLaneRunFacts(root)
		if err != nil {
			return func() {}, err
		}
		switch {
		case len(facts) == 0:
			return func() {}, errors.New(noActiveFactorySentinel)
		case len(facts) == 1 && afterCleanup:
			// The join half of the operator's cleanup-then-join choice.
			return enterSelectedFactoryRun(root, facts[0].RunID, true, timing)
		}
		choice, err := activeRunChooser.ChooseLaneRun(buildLaneRows(facts, names))
		if err != nil {
			return func() {}, err
		}
		switch choice.Kind {
		case choiceJoin:
			return enterSelectedFactoryRun(root, choice.RunID, true, timing)
		case choiceCleanupThenJoin:
			// Any non-nil error is a refusal: print it and present again from
			// facts re-read, even when one run remains (REQ-LJR-006).
			rerr := retireChosenRun(root, choice.RunID)
			afterCleanup = rerr == nil
			if rerr != nil {
				laneNotice(laneMsgCleanupFailFmt, choice.RunID, rerr)
			}
		case choiceMergeThenJoin:
			// Same shape as the cleanup arm: a refusal is printed and the
			// facts re-read; a completed merge leaves the live target, and
			// the join half fires once exactly one run remains.
			res, target, merr := mergeChosenRun(root, choice.RunID)
			afterCleanup = merr == nil
			if merr != nil {
				laneNotice(laneMsgMergeFailFmt, choice.RunID, merr)
			} else {
				laneNotice(laneMsgMergedFmt, choice.RunID, target, joinOrNone(res.Moved), joinOrNone(res.Abandoned))
			}
		default:
			laneNotice("%s", laneMsgCancelled)
			return func() {}, errors.New(joinCancelledSentinel)
		}
	}
}

// laneNotice prints one line to the loop's notice stream. A notice that cannot
// be written changes nothing the operator chose, so the failure is dropped.
func laneNotice(format string, args ...any) {
	if _, err := fmt.Fprintf(laneRunNotices, format+"\n", args...); err != nil {
		return
	}
}

// readLaneRunFacts reads the active runs through the read path of
// REQ-LJR-001 together with the leader-name registry, and closes the factory
// handle before the chooser runs.
func readLaneRunFacts(root string) (_ []homestate.RunFacts, _ map[int]string, err error) {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return nil, nil, err
	}
	defer closeFactoryInto(&err, db, "factory state")
	facts, err := db.ActiveRunFacts(context.Background(), factorymsg.ReconcileOptionsFor(root))
	if err != nil {
		return nil, nil, err
	}
	return facts, leaderNamesByPID(root), nil
}

// mergeChosenRun is the chooser's merge call: it moves the dead run's cards
// under the one active run with a live owner and retires it, through the
// consolidation operation. The eligibility rule is re-checked here against the
// store, not trusted from the rows the operator was shown — a source that is
// not dead, or a listing with no single live target, is refused and changes
// nothing, whatever a chooser returned.
func mergeChosenRun(root, sourceID string) (res homestate.RebindResult, targetID string, err error) {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return res, "", err
	}
	defer closeFactoryInto(&err, db, "factory state")
	opts := factorymsg.ReconcileOptionsFor(root)
	owners, err := db.ClassifyRuns(context.Background(), opts)
	if err != nil {
		return res, "", err
	}
	var src, tgt *homestate.RunOwner
	liveRuns := 0
	for i := range owners {
		if owners[i].RunID == sourceID {
			src = &owners[i]
		}
		if owners[i].Status == "active" && owners[i].Classification == homestate.OwnerLive {
			liveRuns++
			tgt = &owners[i]
		}
	}
	if src == nil {
		return res, "", fmt.Errorf("factory run %s not found", sourceID)
	}
	if liveRuns != 1 {
		return res, "", fmt.Errorf("%d active runs have a live owner; a merge needs exactly one target", liveRuns)
	}
	res, err = db.RebindRun(context.Background(), homestate.RebindRequest{Source: *src, Target: *tgt}, opts)
	return res, tgt.RunID, err
}

// retireChosenRun is the chooser's ONE retirement call. It goes through
// RetireRunIfDead with the production options — no injected classifier — so
// `retirable` stays the only retirement predicate and a live or
// indeterminate owner is refused whatever a chooser returned.
func retireChosenRun(root, runID string) (err error) {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	if _, rerr := db.RetireRunIfDead(context.Background(), runID, factorymsg.ReconcileOptionsFor(root)); rerr != nil {
		return rerr
	}
	return nil
}
