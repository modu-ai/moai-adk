package fetchweb

import (
	"context"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// scoring.go — SPEC-FETCH-WEB-001 M2. The two-stage Jev scoring pipeline and
// the selection rules (REQ-FW-004..REQ-FW-006).

// Scoring thresholds. Every token figure below is APPROXIMATE: token counts
// come from jev.EstimateTokens (bytes/4, conservative), so windows split and
// the passthrough/backfill bounds trigger at slightly different points than
// an exact tokenizer would produce (REQ-FW-002). Character bounds (the
// neighbor context) are characters, not tokens.
const (
	// Stage1WindowTokens is the approximate stage-1 window size (REQ-FW-004).
	Stage1WindowTokens = 3000
	// Stage1Threshold is the stage-1 candidate cut (REQ-FW-004).
	Stage1Threshold = 0.6
	// Stage2Threshold is the stage-2 keep cut (REQ-FW-005).
	Stage2Threshold = 0.8
	// NeighborRescueThreshold is the stage-1 score a SAME-HEADING neighbor
	// needs for a below-cut candidate to stay kept (REQ-FW-005).
	NeighborRescueThreshold = 0.6
	// BackfillTokens is the approximate kept-set floor the selection stage
	// backfills to (REQ-FW-006).
	BackfillTokens = 300
	// PassthroughTokens is the approximate whole-document ceiling under
	// which no scoring happens at all (REQ-FW-003).
	PassthroughTokens = 2000
	// NeighborContextChars is the stage-2 neighbor context size, in
	// CHARACTERS (not tokens — upstream's figure is a character count).
	NeighborContextChars = 400
)

// Answerer is the Jev call seam. *jev.Client satisfies it; tests inject a
// fake. The interface exists so this package never constructs a transport
// and no test can reach the vendor endpoint (REQ-FW-010).
type Answerer interface {
	Ask(ctx context.Context, req jev.Request) jev.Result
}

// Result is the pipeline outcome the CLI and MCP surfaces render.
type Result struct {
	// Status is one of filtered / passthrough / no_match / empty /
	// fetch_error (REQ-FW-007). fetch_error is produced here only for an
	// unavailable scoring call or one whose answer list is missing a
	// requested id — "could not judge" is an error, not a no_match (F2,
	// t1968 sync-audit) — and by the page fetcher for transport failures.
	Status Status
	// Blocks are the kept blocks in document order, text verbatim.
	Blocks []Block
	// Condition carries the failure reason for fetch_error.
	Condition string
	// Warnings collects the non-fatal notices the scoring calls raised —
	// today the stale-cache warning (REQ-FW-013 keep-and-warn) — so the CLI
	// and MCP surfaces can carry them to the reader. Warnings never change
	// the status or the exit code.
	Warnings []string
	// TotalTokens / KeptTokens are approximate token counts (see the package
	// note) of the split document and the kept set.
	TotalTokens int
	KeptTokens  int
}

// Status is the pipeline's terminal condition (REQ-FW-007).
type Status string

const (
	StatusFiltered    Status = "filtered"
	StatusPassthrough Status = "passthrough"
	StatusNoMatch     Status = "no_match"
	StatusEmpty       Status = "empty"
	StatusFetchError  Status = "fetch_error"
)

// ExitCode maps statuses onto the port's exit-code contract: 0 for filtered
// and passthrough, 1 for no_match and empty, 2 for fetch_error (REQ-FW-007,
// plan B2).
func (s Status) ExitCode() int {
	switch s {
	case StatusFiltered, StatusPassthrough:
		return 0
	case StatusNoMatch, StatusEmpty:
		return 1
	default:
		return 2
	}
}

// Process runs the pipeline over an already-split document (REQ-FW-003..
// REQ-FW-006): empty first, then the ≤ 2,000-token passthrough short-circuit
// (no scoring call at all — approximate bound), then two-stage scoring and
// selection.
func Process(ctx context.Context, blocks []Block, question string, ask Answerer) Result {
	if len(blocks) == 0 {
		return Result{Status: StatusEmpty}
	}
	total := DocTokens(blocks)
	if total <= PassthroughTokens { // approximate count (package note)
		return Result{Status: StatusPassthrough, Blocks: blocks, TotalTokens: total, KeptTokens: total}
	}

	scores, warnings, err := scoreTwoStage(ctx, blocks, question, ask)
	if err != nil {
		return Result{Status: StatusFetchError, Condition: err.Error(), TotalTokens: total, Warnings: warnings}
	}
	kept := selectBlocks(blocks, scores)
	if len(kept) == 0 {
		return Result{Status: StatusNoMatch, TotalTokens: total, Warnings: warnings}
	}
	keptTokens := 0
	for _, i := range kept {
		keptTokens += blocks[i].Tokens()
	}
	out := make([]Block, len(kept))
	for j, i := range kept {
		out[j] = blocks[i]
	}
	return Result{Status: StatusFiltered, Blocks: out, TotalTokens: total, KeptTokens: keptTokens, Warnings: warnings}
}

// warningSink collects the warnings the scoring calls raise, deduplicated —
// a windowed or batched stage asks many questions per document, and one
// stale cache serve would otherwise repeat its warning per call. The sink
// keeps REQ-FW-013's keep-and-warn contract: a served answer's warning must
// reach the pipeline's reader, not die inside the stage that observed it.
type warningSink struct {
	seen map[string]bool
	list []string
}

func (w *warningSink) add(s string) {
	if s == "" || w.seen[s] {
		return
	}
	if w.seen == nil {
		w.seen = map[string]bool{}
	}
	w.seen[s] = true
	w.list = append(w.list, s)
}

// blockScore is one block's two-stage outcome.
type blockScore struct {
	stage1 float64
	stage2 float64 // defined only for candidates
}

// scoreTwoStage runs stage 1 over windows and stage 2 over the candidates,
// scoring the main flow and the aside stream separately (REQ-FW-001). The
// warnings the scoring calls raise are collected for the pipeline result.
func scoreTwoStage(ctx context.Context, blocks []Block, question string, ask Answerer) ([]blockScore, []string, error) {
	scores := make([]blockScore, len(blocks))
	var warnings warningSink
	for _, pass := range [2]bool{false, true} { // main flow, then aside
		idx := filterIndices(blocks, func(b Block) bool { return b.Aside == pass })
		if len(idx) == 0 {
			continue
		}
		stream := pick(blocks, idx)
		streamPos := make(map[int]int, len(idx)) // global index -> stream position
		for k, i := range idx {
			streamPos[i] = k
		}
		s1, err := stage1(ctx, stream, idx, question, ask, &warnings)
		if err != nil {
			return nil, warnings.list, err
		}
		for k, i := range idx {
			scores[i].stage1 = s1[k]
		}
		candidates := candidateIndices(s1)
		if len(candidates) == 0 {
			continue
		}
		globalCand := make([]int, len(candidates))
		for k, ci := range candidates {
			globalCand[k] = idx[ci]
		}
		s2, err := stage2(ctx, stream, streamPos, globalCand, question, ask, &warnings)
		if err != nil {
			return nil, warnings.list, err
		}
		for k, gi := range globalCand {
			scores[gi].stage2 = s2[k]
		}
	}
	return scores, warnings.list, nil
}

// stage1Question builds one block's stage-1 question text. Extracted so the
// window sizer and the ask loop build the identical string and cannot drift
// apart on the framing a request actually carries.
func stage1Question(question, passage string) string {
	return fmt.Sprintf("Is this passage relevant to the question %q? Passage: %q", question, passage)
}

// stage1 asks one noul question per block over windows whose BODY stays
// within approximately Stage1WindowTokens tokens (REQ-FW-004) and whose
// assembled REQUEST stays inside jev's size bounds (see stage1Windows). One
// request per window: state is the window's joined block text, and each
// question quotes its own block so the per-block answer is self-contained.
// global maps stream positions to whole-document indices; question ids carry
// the GLOBAL index (b<i>) so ids stay unique across the main and aside
// streams.
func stage1(ctx context.Context, stream []Block, global []int, question string, ask Answerer, sink *warningSink) ([]float64, error) {
	out := make([]float64, len(stream))
	for _, window := range stage1Windows(stream, question) { // approximate bounds
		texts := make([]string, len(window))
		qs := make([]jev.Question, len(window))
		for k, i := range window {
			texts[k] = stream[i].Text
			qs[k] = jev.Question{
				ID:   fmt.Sprintf("b%d", global[i]),
				Text: stage1Question(question, stream[i].Text),
			}
		}
		res := ask.Ask(ctx, jev.Request{
			State:      strings.Join(texts, "\n\n"),
			Questions:  qs,
			Repeatable: true, // a pure read; no side effects to repeat
		})
		if !res.OK() {
			return nil, fmt.Errorf("stage-1 scoring unavailable (%s): %s", res.Availability, res.Condition)
		}
		sink.add(res.StaleCacheWarning)
		probs := answersByID(res)
		for _, i := range window {
			id := fmt.Sprintf("b%d", global[i])
			p, ok := probs[id]
			if !ok {
				// F2 (t1968 sync-audit): an OK response whose answer list
				// lacks a requested id is a judgment failure — "could not
				// judge" — not a probability-0 data verdict. Read as a zero
				// it would surface as no_match (exit 1), hiding the failure.
				return nil, fmt.Errorf("stage-1 scoring incomplete: no answer for question %s (could not judge, not a no-match)", id)
			}
			out[i] = p
		}
	}
	return out, nil
}

// stage2 re-examines each candidate standalone plus a neighbor context of
// approximately NeighborContextChars CHARACTERS (REQ-FW-005). candidates are
// GLOBAL indices (also positions into stream's own numbering is avoided: the
// caller passes stream-local context but global ids); neighbor context is
// drawn from the candidate's own stream. The questions are BATCHED so every
// request stays inside jev's size bounds: a candidate set whose combined
// questions exceed MaxRequestTokens would be refused unsent after stage 1
// already ran. Each batch is scored independently; the selection over the
// collected scores is unchanged.
func stage2(ctx context.Context, stream []Block, streamPos map[int]int, candidates []int, question string, ask Answerer, sink *warningSink) ([]float64, error) {
	qs := make([]jev.Question, len(candidates))
	posByID := make(map[string]int, len(candidates))
	for k, gi := range candidates {
		contextText := neighborContext(stream, streamPos[gi], NeighborContextChars)
		qs[k] = jev.Question{
			ID:   fmt.Sprintf("c%d", gi),
			Text: fmt.Sprintf("Passage: %q Context: %q — does the passage answer the question %q?", stream[streamPos[gi]].Text, contextText, question),
		}
		posByID[qs[k].ID] = k
	}
	out := make([]float64, len(candidates))
	for _, batch := range batchQuestions(question, qs) {
		res := ask.Ask(ctx, jev.Request{
			State:      question,
			Questions:  batch,
			Repeatable: true,
		})
		if !res.OK() {
			return nil, fmt.Errorf("stage-2 scoring unavailable (%s): %s", res.Availability, res.Condition)
		}
		sink.add(res.StaleCacheWarning)
		probs := answersByID(res)
		for _, q := range batch {
			p, ok := probs[q.ID]
			if !ok {
				// F2 (t1968 sync-audit), stage-2 leg: a missing answer here
				// read as 0 would drop the candidate into a misleading
				// verdict — the same could-not-judge error path as stage 1.
				return nil, fmt.Errorf("stage-2 scoring incomplete: no answer for question %s (could not judge, not a no-match)", q.ID)
			}
			out[posByID[q.ID]] = p
		}
	}
	return out, nil
}

// batchQuestions greedily groups questions so each request stays within
// jev's size bounds: state + every question <= MaxRequestTokens and state +
// the longest question <= MaxStateAndQuestionTokens — the same accounting
// jev's checkBounds applies, so no batch the builder emits is refused unsent.
// A single question that alone exceeds the per-question bound occupies its
// own batch: jev refuses it, which is the bound's contract, not a batching
// failure.
func batchQuestions(state string, qs []jev.Question) [][]jev.Question {
	stateTokens := jev.EstimateTokens(state)
	var batches [][]jev.Question
	var cur []jev.Question
	curTokens := stateTokens
	longest := 0
	for _, q := range qs {
		n := jev.QuestionTokens(q)
		if len(cur) > 0 && (curTokens+n > jev.MaxRequestTokens || stateTokens+max(longest, n) > jev.MaxStateAndQuestionTokens) {
			batches = append(batches, cur)
			cur, curTokens, longest = nil, stateTokens, 0
		}
		cur = append(cur, q)
		curTokens += n
		longest = max(longest, n)
	}
	if len(cur) > 0 {
		batches = append(batches, cur)
	}
	return batches
}

// selectBlocks applies REQ-FW-006's decision order:
//
//  1. no_match first: zero stage-1 candidates -> no selection guarantees at
//     all, not even the first-block keep;
//  2. with >= 1 candidate: keep stage-2 survivors and same-heading rescues,
//     always keep the document's first block, backfill to ~300 tokens in
//     descending stage-1 score order, and accompany every kept table row
//     with its header row;
//  3. emit kept indices in document order.
func selectBlocks(blocks []Block, scores []blockScore) []int {
	candidate := make([]bool, len(blocks))
	anyCandidate := false
	for i, s := range scores {
		if s.stage1 >= Stage1Threshold {
			candidate[i] = true
			anyCandidate = true
		}
	}
	if !anyCandidate {
		return nil // no_match decided first (REQ-FW-006)
	}

	kept := map[int]bool{}
	for i, s := range scores {
		if !candidate[i] {
			continue
		}
		if s.stage2 >= Stage2Threshold {
			kept[i] = true
			continue
		}
		if sameHeadingNeighborQualifies(blocks, scores, i) {
			kept[i] = true
		}
	}
	// First block of the document is always kept once a candidate exists.
	kept[0] = true

	// Backfill in descending stage-1 score order until ~BackfillTokens
	// tokens (approximate) or exhaustion.
	if totalKeptTokens(blocks, kept) < BackfillTokens {
		order := make([]int, len(blocks))
		for i := range order {
			order[i] = i
		}
		sortByScoreDesc(order, scores)
		for _, i := range order {
			if totalKeptTokens(blocks, kept) >= BackfillTokens {
				break
			}
			kept[i] = true
		}
	}

	// Table-header accompaniment: a kept row pulls in its header row.
	for i, b := range blocks {
		if kept[i] && b.Kind == KindTableRow && b.HeaderIndex >= 0 && b.HeaderIndex < len(blocks) {
			kept[b.HeaderIndex] = true
		}
	}

	out := make([]int, 0, len(kept))
	for i := range blocks {
		if kept[i] {
			out = append(out, i)
		}
	}
	return out
}

// sameHeadingNeighborQualifies reports whether any block in the SAME scoring
// stream (same Aside flag — REQ-FW-001 separate scoring) with the SAME
// heading path (a sibling under the same heading, excluding the block
// itself) has a stage-1 score at or above the rescue threshold.
func sameHeadingNeighborQualifies(blocks []Block, scores []blockScore, i int) bool {
	for j, b := range blocks {
		if j == i || b.Aside != blocks[i].Aside || !samePath(b.HeadingPath, blocks[i].HeadingPath) {
			continue
		}
		if scores[j].stage1 >= NeighborRescueThreshold {
			return true
		}
	}
	return false
}

func samePath(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}

// neighborContext draws approximately maxChars CHARACTERS (runes, not bytes
// — CJK text is 3 bytes per character, and a byte cut both under-fills the
// window and can split a rune mid-character) of context from the adjacent
// blocks of stream[i], previous side first, then next. The joining space
// counts against the budget.
func neighborContext(stream []Block, i, maxChars int) string {
	var b strings.Builder
	used := 0
	add := func(text string) {
		if used >= maxChars {
			return
		}
		room := maxChars - used
		if b.Len() > 0 {
			room-- // reserve the joining space
			if room <= 0 {
				return
			}
		}
		r := []rune(text)
		if len(r) > room {
			r = r[:room]
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
			used++
		}
		b.WriteString(string(r))
		used += len(r)
	}
	if i > 0 {
		add(stream[i-1].Text)
	}
	if i+1 < len(stream) {
		add(stream[i+1].Text)
	}
	return b.String()
}

// stage1Windows groups a stream into consecutive windows (as index slices
// into the stream) for stage-1 scoring. Two bounds govern a window:
//
//   - the BODY budget: the window's joined block text stays within
//     Stage1WindowTokens (the approximate window semantics, REQ-FW-004);
//   - the REQUEST bound (card t1968 review N1): the request a window
//     assembles — its state (the joined body) plus every block's question,
//     each quoting the user question and its own passage — stays within
//     jev's size bounds, the same accounting jev's checkBounds applies
//     (state + all questions <= MaxRequestTokens; state + the longest
//     question <= MaxStateAndQuestionTokens). Body tokens alone miss the
//     per-request overhead: many short blocks or a long question push the
//     assembled request past the bound while the body is still under the
//     window size, and jev refuses the request unsent after stage 1 already
//     committed the window — surfacing as fetch_error on ordinary pages.
//
// The separator between joined texts is counted per part and each part's
// tokens round up individually, so the tracked state slightly OVER-estimates
// the estimate of the joined string — the safe direction: the bound refuses
// earlier, never later. An oversized single block keeps its own window: jev
// refusing it is the bound's contract, not a windowing failure.
func stage1Windows(stream []Block, question string) [][]int {
	var windows [][]int
	var cur []int
	bodyTokens, stateTokens, qTokens, longest := 0, 0, 0, 0
	for i, b := range stream {
		bt := b.Tokens()
		n := jev.QuestionTokens(jev.Question{Text: stage1Question(question, b.Text)})
		sep := 0
		if len(cur) > 0 {
			sep = jev.EstimateTokens("\n\n")
		}
		nextState, nextQ, nextLongest := stateTokens+sep+bt, qTokens+n, max(longest, n)
		if len(cur) > 0 && (bodyTokens+bt > Stage1WindowTokens ||
			nextState+nextQ > jev.MaxRequestTokens ||
			nextState+nextLongest > jev.MaxStateAndQuestionTokens) {
			windows = append(windows, cur)
			cur, bodyTokens = nil, 0
			sep, nextState, nextQ, nextLongest = 0, bt, n, n
		}
		cur = append(cur, i)
		bodyTokens += bt
		stateTokens, qTokens, longest = nextState, nextQ, nextLongest
	}
	if len(cur) > 0 {
		windows = append(windows, cur)
	}
	return windows
}

// answersByID flattens a result's answers into a probability map.
func answersByID(res jev.Result) map[string]float64 {
	m := make(map[string]float64, len(res.Answers))
	for _, a := range res.Answers {
		m[a.QuestionID] = a.Probability
	}
	return m
}

// filterIndices lists the indices of blocks satisfying keep, in order.
func filterIndices(blocks []Block, keep func(Block) bool) []int {
	var out []int
	for i, b := range blocks {
		if keep(b) {
			out = append(out, i)
		}
	}
	return out
}

// pick returns the blocks at the given indices.
func pick(blocks []Block, idx []int) []Block {
	out := make([]Block, len(idx))
	for k, i := range idx {
		out[k] = blocks[i]
	}
	return out
}

func candidateIndices(scores []float64) []int {
	var out []int
	for i, s := range scores {
		if s >= Stage1Threshold {
			out = append(out, i)
		}
	}
	return out
}

func totalKeptTokens(blocks []Block, kept map[int]bool) int {
	total := 0
	for i := range kept {
		if i < len(blocks) {
			total += blocks[i].Tokens()
		}
	}
	return total
}

// sortByScoreDesc sorts order by descending stage-1 score, ties broken by
// document order (stable).
func sortByScoreDesc(order []int, scores []blockScore) {
	for i := 1; i < len(order); i++ {
		for j := i; j > 0; j-- {
			if scores[order[j]].stage1 > scores[order[j-1]].stage1 ||
				(scores[order[j]].stage1 == scores[order[j-1]].stage1 && order[j] < order[j-1]) {
				order[j], order[j-1] = order[j-1], order[j]
			} else {
				break
			}
		}
	}
}
