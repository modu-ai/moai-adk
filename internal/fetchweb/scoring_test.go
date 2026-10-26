package fetchweb

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// scoring_test.go — SPEC-FETCH-WEB-001 M2 (AC-FW-002, AC-FW-003). Two-stage
// Jev scoring and selection rules against a fake answerer pinned per rule,
// plus the passthrough short-circuit with an answerer that FAILS the test on
// any call.

// fakeAnswerer scripts probabilities by question id (stage-1 ids are "b<i>",
// stage-2 ids "c<i>") and counts calls. An id absent from the script yields
// 0. When failOnCall is set, any Ask fails the test — the passthrough probe's
// zero-call assertion.
type fakeAnswerer struct {
	script   map[string]float64
	calls    int
	failTest *testing.T
}

func (f *fakeAnswerer) Ask(_ context.Context, req jev.Request) jev.Result {
	f.calls++
	if f.failTest != nil {
		f.failTest.Fatalf("answerer was called but the pipeline promised zero scoring calls; state=%q", req.State)
	}
	answers := make([]jev.Answer, 0, len(req.Questions))
	for _, q := range req.Questions {
		answers = append(answers, jev.Answer{
			QuestionID:  q.ID,
			Kind:        jev.KindNoul,
			Probability: f.script[q.ID],
			Model:       jev.VerifiedInterpretedModel,
		})
	}
	return jev.Result{Availability: jev.Available, Answers: answers}
}

// mkBlocks builds n paragraph blocks whose texts are distinct. tokens each
// sets the per-block approximate token count (bytes = tokens*4), so callers
// can size a document past the ~2,000-token passthrough bound.
func mkBlocks(n, tokens int) []Block {
	blocks := make([]Block, n)
	for i := range blocks {
		head := fmt.Sprintf("block number %02d alpha beta gamma", i)
		if pad := tokens*4 - len(head); pad > 0 {
			head += strings.Repeat("p", pad)
		}
		blocks[i] = Block{
			Kind:        KindParagraph,
			HeadingPath: []string{"Doc"},
			Text:        head,
		}
	}
	return blocks
}

// REQ-FW-004: a block whose relevance probability is >= 0.6 becomes a
// candidate; below 0.6 it does not.
func TestStage1_CandidateThreshold(t *testing.T) {
	blocks := mkBlocks(3, 10)
	fake := &fakeAnswerer{script: map[string]float64{"b0": 0.7, "b1": 0.59, "b2": 0.6}}
	ident := []int{0, 1, 2}
	scores, err := stage1(t.Context(), blocks, ident, "q", fake, &warningSink{})
	if err != nil {
		t.Fatal(err)
	}
	want := []bool{true, false, true}
	for i, w := range want {
		if got := scores[i] >= Stage1Threshold; got != w {
			t.Errorf("block %d: candidate = %v (p=%.2f), want %v", i, got, scores[i], w)
		}
	}
	if fake.calls == 0 {
		t.Fatal("stage 1 made no scoring call")
	}
}

// REQ-FW-004: stage-1 windows hold approximately 3,000 tokens. A 12-block
// document of ~9-token blocks is one window (108 tokens); a document over
// the window bound splits so that no window exceeds it (approximate counts).
func TestStage1_WindowsBoundTokens(t *testing.T) {
	var blocks []Block
	// 700 blocks x ~9 tokens ≈ 6,300 tokens -> must split into >1 window.
	for i := 0; i < 700; i++ {
		blocks = append(blocks, Block{Kind: KindParagraph, Text: fmt.Sprintf("block number %03d alpha beta gamma", i)})
	}
	windows := buildWindows(blocks, Stage1WindowTokens)
	if len(windows) < 2 {
		t.Fatalf("6300-token doc produced %d windows, want >= 2", len(windows))
	}
	for wi, w := range windows {
		total := 0
		for _, i := range w {
			total += blocks[i].Tokens()
		}
		if total > Stage1WindowTokens && len(w) > 1 {
			t.Errorf("window %d holds %d tokens over the ~%d bound with %d blocks", wi, total, Stage1WindowTokens, len(w))
		}
	}
	// Window membership covers every block exactly once, in document order.
	seen := 0
	for _, w := range windows {
		seen += len(w)
	}
	if seen != len(blocks) {
		t.Errorf("windows cover %d blocks, want %d", seen, len(blocks))
	}
}

// REQ-FW-005: a candidate with stage-2 p >= 0.8 is kept; below 0.8 it is
// kept ONLY when a same-heading neighbor scores >= 0.6 at stage 1.
func TestSelection_Stage2KeepAndRescue(t *testing.T) {
	// Four blocks of ~550 approximate tokens each (total ~2,200 > the
	// passthrough bound, so the document is scored; the kept set of three
	// blocks sits far above the 300-token backfill floor, so backfill cannot
	// re-add the dropped block). b0 candidate 0.9 (stage2 0.85 -> keep).
	// b1 candidate 0.7 (stage2 0.5, same-heading neighbor b0 at 0.9 ->
	// rescue keep). b2 candidate 0.8 (stage2 0.3 -> rescue keep via b0).
	// b3 under a DIFFERENT heading with no candidate sibling: stage2 0.3,
	// no rescue -> dropped.
	blocks := []Block{
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "first block under heading A " + strings.Repeat("a", 2180)},
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "second block also under A " + strings.Repeat("b", 2170)},
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "third block under A too " + strings.Repeat("c", 2175)},
		{Kind: KindParagraph, HeadingPath: []string{"B"}, Text: "lonely block under heading B " + strings.Repeat("d", 2170)},
	}
	fake := &fakeAnswerer{script: map[string]float64{
		"b0": 0.9, "b1": 0.7, "b2": 0.8, "b3": 0.65,
		"c0": 0.85, "c1": 0.5, "c2": 0.3, "c3": 0.3,
	}}
	res := Process(context.Background(), blocks, "q", fake)
	if res.Status != StatusFiltered {
		t.Fatalf("status = %q, want filtered", res.Status)
	}
	if !keptHas(res, blocks[0].Text) || !keptHas(res, blocks[1].Text) {
		t.Errorf("stage-2 keep or same-heading rescue failed: %d blocks", len(res.Blocks))
	}
	if !keptHas(res, blocks[2].Text) {
		t.Errorf("same-heading rescue (neighbor b0 >= 0.6) failed to keep b2: %d blocks", len(res.Blocks))
	}
	if keptHas(res, blocks[3].Text) {
		t.Errorf("b3 (stage2 0.3, no same-heading candidate neighbor) must be dropped: %d blocks", len(res.Blocks))
	}
}

// A block under a heading with NO candidate sibling scoring >= 0.6 is not
// rescued (the neighbor rule needs a qualifying neighbor).
func TestSelection_RescueNeedsQualifyingNeighbor(t *testing.T) {
	blocks := []Block{
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "only block under A"},
	}
	fake := &fakeAnswerer{script: map[string]float64{"b0": 0.61, "c0": 0.2}}
	res := Process(context.Background(), blocks, "q", fake)
	if !keptHas(res, "only block under A") {
		t.Fatalf("the document's first block must be kept once a candidate exists: %v", texts(res.Blocks))
	}
}

// REQ-FW-006 decision order: zero stage-1 candidates -> no_match, and the
// first-block keep does NOT apply.
func TestSelection_NoMatchWhenZeroCandidates(t *testing.T) {
	blocks := mkBlocks(4, 550) // ~2,200 tokens: past the passthrough bound
	fake := &fakeAnswerer{script: map[string]float64{"b0": 0.5, "b1": 0.1, "b2": 0.59, "b3": 0.0}}
	res := Process(context.Background(), blocks, "q", fake)
	if res.Status != StatusNoMatch {
		t.Fatalf("status = %q, want no_match (zero stage-1 candidates)", res.Status)
	}
	if len(res.Blocks) != 0 {
		t.Errorf("no_match must emit no blocks, got %v", texts(res.Blocks))
	}
}

// REQ-FW-006: once >= 1 candidate exists the first block is always kept.
func TestSelection_FirstBlockAlwaysKept(t *testing.T) {
	blocks := mkBlocks(5, 550)
	// Only the LAST block is a candidate with a high stage-2; the first
	// block is not a candidate at all and must still be kept.
	fake := &fakeAnswerer{script: map[string]float64{
		"b0": 0.1, "b1": 0.2, "b2": 0.3, "b3": 0.1, "b4": 0.95,
		"c4": 0.9,
	}}
	res := Process(context.Background(), blocks, "q", fake)
	if res.Status != StatusFiltered {
		t.Fatalf("status = %q, want filtered", res.Status)
	}
	if len(res.Blocks) == 0 || res.Blocks[0].Text != blocks[0].Text {
		t.Errorf("first block of the document must lead the kept set, got %v", texts(res.Blocks))
	}
}

// REQ-FW-006: a kept set under 300 tokens is backfilled in descending
// stage-1 score order until the threshold is met or blocks run out.
func TestSelection_BackfillUnder300Tokens(t *testing.T) {
	// ~65-token blocks. Candidates: b0 (0.95), b2 (0.9), b4 (0.8), b6 (0.7);
	// stage-2 keeps only b0 (0.9) and drops the rest. Kept = {b0} = 65 tokens
	// << 300. Backfill must add candidates by stage-1 score desc: b2 (0.9),
	// b4 (0.8), b6 (0.7), then non-candidates by score until ~300 tokens.
	n := 34 // 34 blocks x ~65 tokens ≈ 2,210: past passthrough
	blocks := mkBlocks(n, 65)
	script := map[string]float64{"b0": 0.95, "b2": 0.9, "b4": 0.8, "b6": 0.7, "c0": 0.9}
	for i := 0; i < n; i++ {
		if _, ok := script[fmt.Sprintf("b%d", i)]; !ok {
			script[fmt.Sprintf("b%d", i)] = 0.05
		}
	}
	fake := &fakeAnswerer{script: script}
	res := Process(context.Background(), blocks, "q", fake)
	// Backfilled set must be in document order in the output, but the ORDER
	// OF ADDITION is by score: verify b2 present (added first after b0's 9
	// tokens).
	if !keptHas(res, blocks[2].Text) {
		t.Errorf("backfill must add highest-scored non-kept candidate b2 first, got %v", texts(res.Blocks))
	}
	if tokensOf(res.Blocks) < BackfillTokens && len(res.Blocks) < n {
		t.Errorf("backfill stopped before both the %d-token threshold and exhaustion: %d blocks, %d tokens",
			BackfillTokens, len(res.Blocks), tokensOf(res.Blocks))
	}
}

// Backfill prefers by stage-1 score: two candidate blocks both dropped at
// stage 2, the higher-scored one is added first (observable via a document
// that can only fit one more before the threshold).
func TestSelection_BackfillScoreOrder(t *testing.T) {
	// 40 + 1,120 + 1,160 approximate tokens = 2,320: scored, not passed
	// through.
	blocks := []Block{
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: strings.Repeat("x", 160)},
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: strings.Repeat("y", 4480)},
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: strings.Repeat("z", 4640)},
	}
	// b0 first block kept; b1, b2 candidates dropped at stage 2; kept = {b0}
	// = 40 tokens < 300 -> backfill picks b1 (0.9) over b2 (0.5); after b1
	// the set is 1,160 -> stop, b2 not added.
	fake := &fakeAnswerer{script: map[string]float64{"b0": 0.62, "b1": 0.9, "b2": 0.5, "c0": 0.1, "c1": 0.1, "c2": 0.1}}
	res := Process(context.Background(), blocks, "q", fake)
	if !keptHas(res, strings.Repeat("y", 4480)) {
		t.Errorf("backfill must add the higher-scored b1 first, got %d blocks", len(res.Blocks))
	}
	if keptHas(res, strings.Repeat("z", 4640)) {
		t.Errorf("backfill must stop once >= 300 tokens; b0(40)+b1(1120) is past it, b2 must not be added")
	}
}

// REQ-FW-006: every kept table row is accompanied by its table's header row.
func TestSelection_TableHeaderAccompaniment(t *testing.T) {
	pad := strings.Repeat("t", 2380) // ~600 tokens per block; 4 blocks ≈ 2,400 > bound
	blocks := []Block{
		{Kind: KindParagraph, HeadingPath: []string{"Flags"}, Text: "paragraph before the table " + pad},
		{Kind: KindTableRow, HeadingPath: []string{"Flags"}, Text: "Flag | Meaning " + pad, HeaderIndex: 1},
		{Kind: KindTableRow, HeadingPath: []string{"Flags"}, Text: "--json | Emit JSON " + pad, HeaderIndex: 1},
		{Kind: KindTableRow, HeadingPath: []string{"Flags"}, Text: "--quiet | Silence " + pad, HeaderIndex: 1},
	}
	// Only the --json row survives scoring; its header must accompany it.
	fake := &fakeAnswerer{script: map[string]float64{
		"b0": 0.1, "b1": 0.2, "b2": 0.9, "b3": 0.1, "c2": 0.85,
	}}
	res := Process(context.Background(), blocks, "q", fake)
	if !keptHas(res, blocks[2].Text) {
		t.Fatalf("the scored table row must be kept, got %d blocks", len(res.Blocks))
	}
	if !keptHas(res, blocks[1].Text) {
		t.Errorf("kept table row must be accompanied by its header row, got %d blocks", len(res.Blocks))
	}
	if keptHas(res, blocks[3].Text) {
		t.Errorf("the unscored row must stay dropped (kept set is past the backfill floor)")
	}
}

// REQ-FW-006: kept blocks are emitted in original document order.
func TestSelection_DocumentOrderEmission(t *testing.T) {
	blocks := mkBlocks(4, 550)
	fake := &fakeAnswerer{script: map[string]float64{
		"b0": 0.1, "b1": 0.9, "b2": 0.1, "b3": 0.9, "c1": 0.9, "c3": 0.9,
	}}
	res := Process(context.Background(), blocks, "q", fake)
	if len(res.Blocks) < 3 {
		t.Fatalf("want first block + both candidates, got %v", texts(res.Blocks))
	}
	// b0 (first) precedes b1 which precedes b3.
	pos := map[string]int{}
	for i, b := range res.Blocks {
		pos[b.Text] = i
	}
	if pos[blocks[0].Text] >= pos[blocks[1].Text] || pos[blocks[1].Text] >= pos[blocks[3].Text] {
		t.Errorf("emission not in document order: %v", texts(res.Blocks))
	}
}

// REQ-FW-003 + AC-FW-003: a non-empty document at or under 2,000 tokens
// passes through whole with ZERO scoring calls.
func TestProcess_PassthroughUnder2000TokensZeroCalls(t *testing.T) {
	blocks := mkBlocks(10, 10) // ~100 tokens
	fake := &fakeAnswerer{failTest: t, script: map[string]float64{}}
	res := Process(context.Background(), blocks, "q", fake)
	if res.Status != StatusPassthrough {
		t.Fatalf("status = %q, want passthrough", res.Status)
	}
	if len(res.Blocks) != len(blocks) {
		t.Errorf("passthrough must keep every block, got %d of %d", len(res.Blocks), len(blocks))
	}
	for i := range blocks {
		if res.Blocks[i].Text != blocks[i].Text {
			t.Errorf("passthrough block %d not verbatim", i)
		}
	}
}

// A document just OVER the bound is scored, not passed through.
func TestProcess_Over2000TokensIsScored(t *testing.T) {
	var blocks []Block
	for len(blocks) == 0 || DocTokens(blocks) <= PassthroughTokens {
		blocks = append(blocks, Block{Kind: KindParagraph, Text: strings.Repeat("w", 400)}) // 100 tokens each
	}
	fake := &fakeAnswerer{script: map[string]float64{}}
	res := Process(context.Background(), blocks, "q", fake)
	if res.Status != StatusNoMatch && res.Status != StatusFiltered {
		t.Errorf("over-bound document must be scored, got %q", res.Status)
	}
	if fake.calls == 0 {
		t.Error("over-bound document made no scoring call")
	}
}

// Empty precedes passthrough (REQ-FW-003): zero blocks -> empty.
func TestProcess_EmptyPrecedesPassthrough(t *testing.T) {
	fake := &fakeAnswerer{failTest: t, script: map[string]float64{}}
	res := Process(context.Background(), nil, "q", fake)
	if res.Status != StatusEmpty {
		t.Fatalf("status = %q, want empty (empty precedes passthrough)", res.Status)
	}
}

// An aside block is scored in its own stream: a main-flow sibling cannot
// rescue an aside candidate, and vice versa (REQ-FW-001 separate scoring).
func TestProcess_AsideScoredSeparately(t *testing.T) {
	// Two ~1,100-token blocks (total ~2,200: scored). The kept main block
	// alone is past the backfill floor, so backfill cannot re-add the aside.
	blocks := []Block{
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "main candidate block " + strings.Repeat("m", 4360)},
		{Kind: KindParagraph, HeadingPath: []string{"A"}, Text: "aside block " + strings.Repeat("n", 4350), Aside: true},
	}
	// Main block candidate 0.9/stage2 0.9 (kept); aside candidate 0.7,
	// stage-2 0.5, would be "rescued" if streams mixed (main sibling 0.9
	// under same heading). Separate scoring drops it.
	fake := &fakeAnswerer{script: map[string]float64{"b0": 0.9, "b1": 0.7, "c0": 0.9, "c1": 0.5}}
	res := Process(context.Background(), blocks, "q", fake)
	if !keptHas(res, blocks[0].Text) {
		t.Fatalf("main-flow keep failed: %d blocks", len(res.Blocks))
	}
	if keptHas(res, blocks[1].Text) {
		t.Errorf("aside block must not be rescued by a main-flow neighbor (separate scoring): %d blocks", len(res.Blocks))
	}
}

// A scoring call that comes back unavailable fails the pipeline into
// fetch_error, not no_match: "could not judge" is not "nothing matched".
func TestProcess_UnavailableScoringIsError(t *testing.T) {
	blocks := mkBlocks(3, 700) // ~2,100 tokens: scored
	fake := &unavailableAnswerer{}
	res := Process(context.Background(), blocks, "q", fake)
	if res.Status != StatusFetchError {
		t.Fatalf("status = %q, want fetch_error on unavailable scoring", res.Status)
	}
	if res.Condition == "" {
		t.Error("fetch_error must carry a condition")
	}
}

type unavailableAnswerer struct{}

func (unavailableAnswerer) Ask(_ context.Context, _ jev.Request) jev.Result {
	return jev.Result{Availability: jev.RateLimited, Condition: "endpoint returned HTTP 429"}
}

// Stage-2 neighbor context is ~400 CHARACTERS drawn from adjacent blocks.
func TestNeighborContextAbout400Chars(t *testing.T) {
	blocks := []Block{
		{Kind: KindParagraph, Text: strings.Repeat("a", 300)},
		{Kind: KindParagraph, Text: strings.Repeat("b", 80)},
		{Kind: KindParagraph, Text: strings.Repeat("c", 300)},
	}
	ctx := neighborContext(blocks, 1, NeighborContextChars)
	if l := len(ctx); l == 0 || l > NeighborContextChars+2 {
		t.Errorf("neighbor context length = %d, want ~(0, %d]", l, NeighborContextChars)
	}
	if !strings.Contains(ctx, strings.Repeat("a", 100)) && !strings.Contains(ctx, "aaa") {
		t.Errorf("context should draw from the previous block")
	}
	if !strings.Contains(ctx, strings.Repeat("c", 100)) && !strings.Contains(ctx, "ccc") {
		t.Errorf("context should draw from the next block")
	}
}

// recordingAnswerer scripts probabilities by id like fakeAnswerer and also
// records every request it serves, in order.
type recordingAnswerer struct {
	script map[string]float64
	reqs   []jev.Request
}

func (r *recordingAnswerer) Ask(_ context.Context, req jev.Request) jev.Result {
	r.reqs = append(r.reqs, req)
	answers := make([]jev.Answer, 0, len(req.Questions))
	for _, q := range req.Questions {
		answers = append(answers, jev.Answer{
			QuestionID:  q.ID,
			Kind:        jev.KindNoul,
			Probability: r.script[q.ID],
			Model:       jev.VerifiedInterpretedModel,
		})
	}
	return jev.Result{Availability: jev.Available, Answers: answers}
}

// R1 (card t1968 review): stage-2 candidates are batched so EVERY request the
// pipeline builds stays inside jev's size bounds — a candidate set whose
// combined questions exceed MaxRequestTokens is split across requests, each
// scored independently, with the selection unchanged.
func TestStage2_BatchesWithinRequestBounds(t *testing.T) {
	// 30 blocks of ~3,000 approximate tokens each (~90k total: scored, not
	// passed through). Every block a stage-1 candidate; each stage-2 question
	// carries its ~3,000-token passage plus context, so ONE request over all
	// 30 candidates would hold ~93k tokens — over the 64k bound.
	blocks := mkBlocks(30, 3000)
	script := map[string]float64{}
	for i := 0; i < 30; i++ {
		script[fmt.Sprintf("b%d", i)] = 0.9
		script[fmt.Sprintf("c%d", i)] = 0.9
	}
	rec := &recordingAnswerer{script: script}
	res := Process(context.Background(), blocks, "q", rec)
	if res.Status != StatusFiltered {
		t.Fatalf("status = %q, want filtered", res.Status)
	}
	var stage2Reqs int
	asked := map[string]bool{}
	for _, req := range rec.reqs {
		if len(req.Questions) == 0 || !strings.HasPrefix(req.Questions[0].ID, "c") {
			continue // stage-1 window request
		}
		stage2Reqs++
		total := jev.EstimateTokens(req.State)
		for _, q := range req.Questions {
			asked[q.ID] = true
			total += jev.EstimateTokens(q.Text)
		}
		if total > jev.MaxRequestTokens {
			t.Errorf("stage-2 request holds ~%d tokens, over the %d bound — batch it", total, jev.MaxRequestTokens)
		}
	}
	if stage2Reqs < 2 {
		t.Errorf("a ~93k-token candidate set must split into >= 2 stage-2 requests, got %d", stage2Reqs)
	}
	for i := 0; i < 30; i++ {
		if !asked[fmt.Sprintf("c%d", i)] {
			t.Errorf("candidate c%d was never asked", i)
		}
	}
}

// R5 (card t1968 review): the neighbor context is measured and cut in
// CHARACTERS (runes) — CJK text must fill the full window, not a third of it,
// and must never be cut mid-rune.
func TestNeighborContext_RuneCutCJK(t *testing.T) {
	cjk := strings.Repeat("가", 450) // 450 runes, 1,350 UTF-8 bytes
	blocks := []Block{
		{Kind: KindParagraph, Text: cjk},
		{Kind: KindParagraph, Text: "target"},
		{Kind: KindParagraph, Text: cjk},
	}
	ctx := neighborContext(blocks, 1, NeighborContextChars)
	if !utf8.ValidString(ctx) {
		t.Errorf("neighbor context split a rune mid-character: %q", ctx)
	}
	if got := len([]rune(ctx)); got == 0 || got > NeighborContextChars {
		t.Errorf("neighbor context = %d runes, want ~(0, %d]", got, NeighborContextChars)
	}
	if got := len([]rune(ctx)); got < NeighborContextChars-1 {
		t.Errorf("CJK context under-filled: %d runes, want ~%d (byte-based cutting yields about a third)", got, NeighborContextChars)
	}
}

// helpers

func keptHas(res Result, text string) bool {
	for _, b := range res.Blocks {
		if b.Text == text {
			return true
		}
	}
	return false
}

func texts(blocks []Block) []string {
	out := make([]string, len(blocks))
	for i, b := range blocks {
		out[i] = b.Text
	}
	return out
}

func tokensOf(blocks []Block) int { return DocTokens(blocks) }
