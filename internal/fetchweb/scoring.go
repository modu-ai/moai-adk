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
	// unavailable scoring call — "could not judge" is an error, not a
	// no_match — and by the page fetcher for transport failures.
	Status Status
	// Blocks are the kept blocks in document order, text verbatim.
	Blocks []Block
	// Condition carries the failure reason for fetch_error.
	Condition string
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

	scores, err := scoreTwoStage(ctx, blocks, question, ask)
	if err != nil {
		return Result{Status: StatusFetchError, Condition: err.Error(), TotalTokens: total}
	}
	kept := selectBlocks(blocks, scores)
	if len(kept) == 0 {
		return Result{Status: StatusNoMatch, TotalTokens: total}
	}
	keptTokens := 0
	for _, i := range kept {
		keptTokens += blocks[i].Tokens()
	}
	out := make([]Block, len(kept))
	for j, i := range kept {
		out[j] = blocks[i]
	}
	return Result{Status: StatusFiltered, Blocks: out, TotalTokens: total, KeptTokens: keptTokens}
}

// blockScore is one block's two-stage outcome.
type blockScore struct {
	stage1 float64
	stage2 float64 // defined only for candidates
}

// scoreTwoStage runs stage 1 over windows and stage 2 over the candidates,
// scoring the main flow and the aside stream separately (REQ-FW-001).
func scoreTwoStage(ctx context.Context, blocks []Block, question string, ask Answerer) ([]blockScore, error) {
	scores := make([]blockScore, len(blocks))
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
		s1, err := stage1(ctx, stream, idx, question, ask)
		if err != nil {
			return nil, err
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
		s2, err := stage2(ctx, stream, streamPos, globalCand, question, ask)
		if err != nil {
			return nil, err
		}
		for k, gi := range globalCand {
			scores[gi].stage2 = s2[k]
		}
	}
	return scores, nil
}

// stage1 asks one noul question per block over windows of approximately
// Stage1WindowTokens tokens (REQ-FW-004). One request per window: state is
// the window's joined block text, and each question quotes its own block so
// the per-block answer is self-contained. global maps stream positions to
// whole-document indices; question ids carry the GLOBAL index (b<i>) so ids
// stay unique across the main and aside streams.
func stage1(ctx context.Context, stream []Block, global []int, question string, ask Answerer) ([]float64, error) {
	out := make([]float64, len(stream))
	for _, window := range buildWindows(stream, Stage1WindowTokens) { // approximate bound
		texts := make([]string, len(window))
		qs := make([]jev.Question, len(window))
		for k, i := range window {
			texts[k] = stream[i].Text
			qs[k] = jev.Question{
				ID:   fmt.Sprintf("b%d", global[i]),
				Text: fmt.Sprintf("Is this passage relevant to the question %q? Passage: %q", question, stream[i].Text),
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
		probs := answersByID(res)
		for _, i := range window {
			out[i] = probs[fmt.Sprintf("b%d", global[i])]
		}
	}
	return out, nil
}

// stage2 re-examines each candidate standalone plus a neighbor context of
// approximately NeighborContextChars CHARACTERS (REQ-FW-005). candidates are
// GLOBAL indices (also positions into stream's own numbering is avoided: the
// caller passes stream-local context but global ids); neighbor context is
// drawn from the candidate's own stream.
func stage2(ctx context.Context, stream []Block, streamPos map[int]int, candidates []int, question string, ask Answerer) ([]float64, error) {
	qs := make([]jev.Question, len(candidates))
	for k, gi := range candidates {
		contextText := neighborContext(stream, streamPos[gi], NeighborContextChars)
		qs[k] = jev.Question{
			ID:   fmt.Sprintf("c%d", gi),
			Text: fmt.Sprintf("Passage: %q Context: %q — does the passage answer the question %q?", stream[streamPos[gi]].Text, contextText, question),
		}
	}
	res := ask.Ask(ctx, jev.Request{
		State:      question,
		Questions:  qs,
		Repeatable: true,
	})
	if !res.OK() {
		return nil, fmt.Errorf("stage-2 scoring unavailable (%s): %s", res.Availability, res.Condition)
	}
	probs := answersByID(res)
	out := make([]float64, len(candidates))
	for k, gi := range candidates {
		out[k] = probs[fmt.Sprintf("c%d", gi)]
	}
	return out, nil
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

// neighborContext draws approximately maxChars CHARACTERS of context from
// the adjacent blocks of stream[i] (previous side first, then next).
func neighborContext(stream []Block, i, maxChars int) string {
	var b strings.Builder
	add := func(text string) {
		if b.Len() >= maxChars {
			return
		}
		room := maxChars - b.Len()
		if len(text) > room {
			text = text[:room]
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(text)
	}
	if i > 0 {
		add(stream[i-1].Text)
	}
	if i+1 < len(stream) {
		add(stream[i+1].Text)
	}
	return b.String()
}

// buildWindows groups a stream into consecutive windows (as index slices
// into the stream) whose approximate token totals stay within bound. An
// oversized single block occupies its own window rather than being split
// (its text is one indivisible unit).
func buildWindows(stream []Block, boundTokens int) [][]int {
	var windows [][]int
	var cur []int
	curTokens := 0
	for i, b := range stream {
		bt := b.Tokens()
		if len(cur) > 0 && curTokens+bt > boundTokens {
			windows = append(windows, cur)
			cur, curTokens = nil, 0
		}
		cur = append(cur, i)
		curTokens += bt
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
