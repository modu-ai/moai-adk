// todo_jev_finding.go — SPEC-JEV-CONSUMERS-001 M4, Consumer C: the
// near-duplicate card marking a TypeSafe System One answer produces.
//
// What this file is for, stated as the thing it must NOT become: the answer is
// a RECORD a person reads, never a decision anything acts on (REQ-JEVC-011 /
// REQ-JEVC-012). Nothing here folds, reorders, drops, or edits a card; nothing
// here refuses an admission; nothing here reaches a completion verdict, a merge
// approval, a queue mutation, or an operator gate. The only write is one
// BacklogFinding appended beside the ones the text analyser already writes, and
// the finding contract's own doc comment records why that write cannot become
// anything else by accident.
//
// THE SHIPPING GATE IS NOT MET. REQ-JEVO-009 forbids shipping a consumer whose
// measured accuracy does not beat its own constant-answer baseline, and that
// measurement has not been run — no labelled set exists, and its size is open
// question Q3 in SPEC-JEV-CONSUMERS-001 §E. The code below is therefore built
// behind the default-off `workflow.jev.enabled` gate and is NOT shippable on
// the strength of existing here. The sibling premise-death task is the reason
// the bar is stated that way: it measured 58.9% where a constant answer scored
// 75.0%, and a consumer that looks plausible is exactly the one that gets
// shipped unmeasured.
//
// Three properties are load-bearing and each is the reason a simpler shape was
// rejected:
//
//   - **Admission only** (REQ-JEVN-001). The `moai todo analyze` re-sweep is a
//     distinct entry point and does NOT consult this seam. The re-sweep walks
//     pairs that already carry findings, so a consumer there changes the
//     finding population on a path this SPEC never measured.
//
//   - **Precedence half (a) is expressed, not inherited** (REQ-JEVN-006).
//     AppendFindingOnce's key includes Source, so it would never suppress a
//     Jev finding on a pair a measurement already names. The suppression goes
//     through HasFindingForPairAnySource instead. An un-suppressed append looks
//     exactly like a correct one, so inheriting the existing dedup would ship a
//     rule that silently suppresses nothing.
//
//   - **A missing answer is not a negative answer.** internal/jev returns a
//     typed unavailable result rather than an error for every absence —
//     disabled, no credential, 401, 429, 529, unreachable, oversize. All of
//     them land on the same branch here: record nothing, emit at most one
//     notice line, and let the admission proceed exactly as it would with the
//     capability off (REQ-JEVC-005..007).
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// jevAdmissionTimeout bounds the admission-path call. The probe runs inside
// the queue's locked write, so an unbounded call would hold the cross-process
// lock for as long as the network chose to take — the bound is what keeps a
// degraded endpoint from turning `todo add` into a hang for every session on
// the machine. A timeout lands on Unreachable, which is a recorded absence
// rather than a failure.
const jevAdmissionTimeout = 5 * time.Second

// jevNearDuplicateCandidateLimit bounds how many queued cards enter the state.
// Accuracy falls as irrelevant state grows, and the request's size bounds
// refuse whole rather than truncating, so the bound is applied here where the
// selection is visible rather than at the wire.
const jevNearDuplicateCandidateLimit = 40

// jevNoNearDuplicateChoice is the no-match option every Choice carries. A
// forced choice over an option set that excludes the true answer returns a
// confident wrong answer — for this question the true answer is usually
// "none of them", so omitting the option would manufacture a finding for
// every admitted card.
const jevNoNearDuplicateChoice = "none"

// jevNoticeWriter is where the at-most-one notice line goes. It is a variable
// so a test can read the line without a process-wide stderr capture.
var jevNoticeWriter io.Writer = os.Stderr

// jevNearDuplicateJudgment is one Consumer C answer: the card the model reads
// as a near-duplicate of the newly admitted one, and the probability it
// carried. It is deliberately not a BacklogFinding — the judgment is what the
// model said, and the finding is what the queue records about it.
type jevNearDuplicateJudgment struct {
	RelatedID   string
	Probability float64
}

// jevNearDuplicateProbe is the Consumer C seam. Production wiring resolves the
// capability gate and calls internal/jev; tests replace it with a stub whose
// call count is the positive control behind every absence this consumer
// asserts.
var jevNearDuplicateProbe = liveJevNearDuplicateProbe

// appendJevNearDuplicateFinding records the Jev judgment for the newly
// admitted item, reporting whether a finding was appended.
//
// It is called from the admission path and nowhere else (REQ-JEVN-001). The
// return value exists for tests and for a future caller that wants to report
// what happened; no caller acts on it, and none may.
func appendJevNearDuplicateFinding(rec *kanban.BacklogRecord, item kanban.BacklogItem) bool {
	judgment, ok := jevNearDuplicateProbe(rec, item)
	if !ok {
		return false
	}
	if judgment.RelatedID == "" || judgment.RelatedID == item.ID {
		return false
	}
	f := kanban.BacklogFinding{
		SubjectID: item.ID,
		RelatedID: judgment.RelatedID,
		Relation:  kanban.BacklogRelationNearDuplicate,
		// REQ-JEVN-002 / REQ-JEVN-003: the third constant, never `agent`.
		Source: kanban.BacklogSourceJev,
		Score:  judgment.Probability,
		At:     item.AddedAt,
	}
	// REQ-JEVN-006 half (a): suppressed when a finding of ANY source already
	// names this unordered pair with this relation. Half (b) — a later
	// mechanical or agent finding landing alongside a Jev one — is
	// AppendFindingOnce's unchanged default and is deliberately not touched
	// here.
	if rec.HasFindingForPairAnySource(f) {
		return false
	}
	return rec.AppendFindingOnce(f)
}

// liveJevNearDuplicateProbe is the production seam body: resolve the gate, ask
// one Choice over the queued cards, and return the judgment.
//
// Every absence returns (zero, false). None returns an error, because an error
// here would propagate into the admission's own error return and turn a
// missing model answer into a refused card — the display-becomes-verdict
// failure in its most direct form.
func liveJevNearDuplicateProbe(rec *kanban.BacklogRecord, item kanban.BacklogItem) (jevNearDuplicateJudgment, bool) {
	enabled, err := jevEnabled(resolveProjectDir())
	if err != nil || !enabled {
		// A config that cannot be read is treated exactly as a disabled
		// capability, and silently: the shipped default is off, so a notice
		// on this branch would fire for every `todo add` in every project
		// that never opted in.
		return jevNearDuplicateJudgment{}, false
	}

	candidates := jevNearDuplicateCandidates(rec, item)
	if len(candidates) == 0 {
		return jevNearDuplicateJudgment{}, false
	}

	client := jev.New(true)
	client.LoadCredential = jevcred.Load

	ctx, cancel := context.WithTimeout(context.Background(), jevAdmissionTimeout)
	defer cancel()

	res := client.Ask(ctx, jev.Request{
		State:     jevNearDuplicateState(item, candidates),
		Questions: []jev.Question{jevNearDuplicateQuestion(candidates)},
	})
	if !res.OK() {
		jevNotice(res.NoticeLine())
		return jevNearDuplicateJudgment{}, false
	}
	if len(res.Answers) == 0 {
		return jevNearDuplicateJudgment{}, false
	}
	answer := res.Answers[0]
	if answer.Kind != jev.KindChoice || answer.Choice == "" || answer.Choice == jevNoNearDuplicateChoice {
		return jevNearDuplicateJudgment{}, false
	}
	// The answer is only honoured when it names a card that was actually
	// offered. A choice outside the option set is a malformed answer, not a
	// finding about a card nobody asked about.
	for _, c := range candidates {
		if c.ID == answer.Choice {
			return jevNearDuplicateJudgment{RelatedID: c.ID, Probability: answer.Probability}, true
		}
	}
	return jevNearDuplicateJudgment{}, false
}

// jevNearDuplicateCandidates returns the cards the question is asked about:
// every non-dropped card other than the newly admitted one, most recent first,
// bounded.
func jevNearDuplicateCandidates(rec *kanban.BacklogRecord, item kanban.BacklogItem) []kanban.BacklogItem {
	out := make([]kanban.BacklogItem, 0, len(rec.Items))
	for i := len(rec.Items) - 1; i >= 0; i-- {
		c := rec.Items[i]
		if c.ID == item.ID || c.State == kanban.BacklogStateDropped {
			continue
		}
		out = append(out, c)
		if len(out) == jevNearDuplicateCandidateLimit {
			break
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].ID < out[b].ID })
	return out
}

// jevNearDuplicateState renders the one state the question is asked over.
func jevNearDuplicateState(item kanban.BacklogItem, candidates []kanban.BacklogItem) string {
	var b strings.Builder
	b.WriteString("A new backlog card has just been admitted to a work queue.\n\n")
	b.WriteString("NEW CARD\n")
	fmt.Fprintf(&b, "%s: %s\n\n", item.ID, item.Text)
	b.WriteString("EXISTING CARDS\n")
	for _, c := range candidates {
		fmt.Fprintf(&b, "%s: %s\n", c.ID, c.Text)
	}
	return b.String()
}

// jevNearDuplicateQuestion builds the single Choice, with the no-match option
// appended last.
func jevNearDuplicateQuestion(candidates []kanban.BacklogItem) jev.Question {
	choices := make([]string, 0, len(candidates)+1)
	for _, c := range candidates {
		choices = append(choices, c.ID)
	}
	choices = append(choices, jevNoNearDuplicateChoice)
	return jev.Question{
		ID:   "near-duplicate",
		Text: "Which existing card, if any, describes the same work as the new card? Answer " + jevNoNearDuplicateChoice + " if none does.",
		Kind: jev.KindChoice,
		// The no-match option is not decoration: without it the model must
		// name a card, and the usual true answer is that no card matches.
		Choices: choices,
	}
}

// jevFindingSignalFragment renders the probability fragment `todoFindingLine`
// prints for a Jev finding (REQ-JEVN-005).
//
// It lives here rather than in the render function so the internal/jev
// dependency stays inside the one file that owns this consumer — the render
// surface needs the LABEL, not the client. The label is the package constant
// rather than a literal, so a reader's recognition of the marker cannot drift
// from what internal/jev emits elsewhere, and the word "score" is deliberately
// absent: a calibrated model confidence must not read as a measured
// similarity.
func jevFindingSignalFragment(f kanban.BacklogFinding) string {
	return fmt.Sprintf(", %s p=%.2f", jev.SignalLabel, f.Score)
}

// jevNotice writes the at-most-one notice line for an unavailable call
// (REQ-JEVC-007). Silent when there is nothing to notice.
func jevNotice(line string) {
	if line == "" || jevNoticeWriter == nil {
		return
	}
	_, _ = fmt.Fprintln(jevNoticeWriter, line)
}
