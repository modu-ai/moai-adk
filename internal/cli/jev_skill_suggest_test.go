// jev_skill_suggest_test.go — SPEC-JEV-CONSUMERS-001 M6 (Consumer B): the
// skill-suggestion protocol's verification surface.
//
// Consumer B's host is the /moai intent router — a prose-level skill surface
// the Go quality gate does not reach — so the verification surface follows the
// foreman precedent (foreman_queue_watch_test.go): take the protocol prose
// VERBATIM from BOTH copies of the skill (the local dogfood copy and the
// template mirror), assert the two agree, validate the block against the
// protocol contract the requirements state, and execute the mechanical anchor
// the block names (the flow, entry-path, and command tests below).
//
// The consumer is in the gate-unrun state (REQ-JEVN-016): present in the tree,
// unreachable at the shipped default, its measurement gate not yet run. Every
// absence asserted here carries a positive control, and the prose contract
// carries a falsification control — mutated blocks are pinned and asserted
// REJECTED, so a validator that rejected nothing could not pass this file.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// suggestionSkillPaths are the two copies of the /moai skill whose
// suggestion-protocol block must agree. Paths are relative to this package
// directory. The copies differ elsewhere by design (the skill-dir token the
// local copy carries); only THIS block is required to be byte-identical.
var suggestionSkillPaths = map[string]string{
	"local":    "../../.claude/skills/moai/SKILL.md",
	"template": "../../internal/template/templates/.claude/skills/moai/SKILL.md",
}

// suggestionProtocolHeading is the block's heading, and the extractor's
// marker. The block runs from the heading to the horizontal rule that closes
// the Intent Router section.
const suggestionProtocolHeading = "### Skill Suggestion (gated — default off)"

// extractSuggestionProtocol returns the suggestion-protocol block from a /moai
// SKILL.md: the `### Skill Suggestion` heading through the section-closing
// horizontal rule, verbatim.
func extractSuggestionProtocol(t *testing.T, skillPath string) string {
	t.Helper()
	raw, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read %s: %v", skillPath, err)
	}
	lines := strings.Split(string(raw), "\n")
	start := -1
	for i, ln := range lines {
		if strings.TrimSpace(ln) == suggestionProtocolHeading {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s: no %q section found — the suggestion-protocol block is absent", skillPath, suggestionProtocolHeading)
	}
	var body []string
	for i := start; i < len(lines); i++ {
		if i > start && strings.TrimSpace(lines[i]) == "---" {
			break
		}
		body = append(body, lines[i])
	}
	return strings.Join(body, "\n")
}

// suggestionProtocolTokens are the clauses the protocol contract requires.
// Each names one requirement the block must carry: the gate and its shipped
// default, the not-available framing, the anchor command and its two-request
// shape, the ranked-signal presentation, the router's retained authority, and
// the suppression rule. A block missing any of them does not state the
// protocol the requirements define.
var suggestionProtocolTokens = []string{
	"workflow.jev.enabled",          // the capability gate, by its config key
	"not an available feature",      // gate-unrun condition (iii): never presented as available
	"Priorities 1-4",                // inert when off; authority unchanged when on
	"moai jev-suggest",              // the mechanical anchor the block names
	"exactly two requests",          // REQ shape: wide rank + Noul, then top-3 rerank
	"needs a skill",                 // the load-bearing Noul
	"rerank of the top three",       // the second request
	"rank position",                 // ranked-signal presentation
	"keeps its selection authority", // the router is never overridden
	"Suppress the ranked list",      // the suppression rule (REQ-JEVN-014)
}

// validateSuggestionProtocol reports whether block carries every clause the
// protocol contract requires. It is a plain token check over the extracted
// prose — the same prose an orchestrator reads — because the block's content
// IS the protocol; nothing else carries it.
func validateSuggestionProtocol(block string) error {
	for _, token := range suggestionProtocolTokens {
		if !strings.Contains(block, token) {
			return errSuggestionProtocol("block is missing the required clause: " + token)
		}
	}
	return nil
}

// errSuggestionProtocol is a tiny error type so validateSuggestionProtocol can
// return a plain error without importing errors in a file whose other tests
// do not need it.
type errSuggestionProtocol string

func (e errSuggestionProtocol) Error() string { return string(e) }

// TestSuggestionProtocol_CopiesAgreeAndCarryContract — the two skill copies
// carry the SAME block, byte-identical, the block sits inside the Intent
// Router section, and it states the full protocol contract.
func TestSuggestionProtocol_CopiesAgreeAndCarryContract(t *testing.T) {
	extracted := map[string]string{}
	for name, path := range suggestionSkillPaths {
		extracted[name] = extractSuggestionProtocol(t, path)

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		body := string(raw)
		router := strings.Index(body, "## Intent Router")
		heading := strings.Index(body, suggestionProtocolHeading)
		workflows := strings.Index(body, "## Workflow Quick Reference")
		if router < 0 || heading < 0 || workflows < 0 || router >= heading || heading >= workflows {
			t.Errorf("%s: the suggestion block does not sit inside the Intent Router section (router=%d, heading=%d, workflows=%d)",
				path, router, heading, workflows)
		}
		if err := validateSuggestionProtocol(extracted[name]); err != nil {
			t.Errorf("%s: suggestion-protocol block fails the contract: %v", name, err)
		}
	}
	if extracted["local"] != extracted["template"] {
		t.Errorf("the local and template suggestion-protocol blocks differ:\n--- local ---\n%s\n--- template ---\n%s",
			extracted["local"], extracted["template"])
	}
}

// TestSuggestionProtocol_FalsificationControl — the falsifiability condition,
// demonstrated in-suite: pinned MUTATED blocks are each REJECTED by the same
// validation the real block passes. A validator that rejected nothing would
// fail here, which is what makes the green above worth reading. Each mutant
// removes or inverts one load-bearing clause.
func TestSuggestionProtocol_FalsificationControl(t *testing.T) {
	real := extractSuggestionProtocol(t, suggestionSkillPaths["local"])
	if err := validateSuggestionProtocol(real); err != nil {
		t.Fatalf("the real block must validate before the mutants below mean anything: %v", err)
	}

	swap := func(from, to string) string {
		if !strings.Contains(real, from) {
			t.Fatalf("pinned mutant source clause not found in the real block: %q", from)
		}
		return strings.Replace(real, from, to, 1)
	}
	cut := func(marker string) string {
		idx := strings.Index(real, marker)
		if idx < 0 {
			t.Fatalf("pinned mutant cut marker not found in the real block: %q", marker)
		}
		return real[:idx]
	}

	mutants := map[string]string{
		// The gate framing removed: the block no longer names the gate key or
		// admits it is not an available feature — it presents unconditionally.
		"gate-framing-removed": cut("Only when"),
		// The authority clause inverted: the suggestion drives selection — the
		// exact display-becomes-verdict defect the consumer is bounded against.
		"authority-inverted": swap("keeps its selection authority unchanged", "drives the router's selection"),
		// The suppression rule removed: a best-of-nothing would be presented.
		"suppression-removed": cut("Suppress the ranked list"),
		// The anchor command removed: the block names no executable anchor, so
		// the protocol it describes cannot be executed.
		"anchor-command-removed": swap("moai jev-suggest", "the suggestion anchor"),
	}

	for name, block := range mutants {
		if err := validateSuggestionProtocol(block); err == nil {
			t.Errorf("mutant %q PASSED validation — the contract check rejects nothing, so the real block's green asserts nothing", name)
		}
	}
}

// ---------------------------------------------------------------------------
// The mechanical anchor the prose block names. The transport stub is the
// recording instrument behind AC-JEVN-009 and behind every AC-JEVN-016
// absence assertion: a zero it records is a measured zero only because the
// positive controls prove it counts.
// ---------------------------------------------------------------------------

// jevSuggestWireRequest is the decoded on-the-wire shape the recording
// transport captures for each request the flow issues.
type jevSuggestWireRequest struct {
	State     string         `json:"state"`
	Questions []jev.Question `json:"questions"`
}

// jevRecordingTransport counts wire requests, captures each request's decoded
// shape, and answers from a per-call script. It IS the "recording transport
// stub" AC-JEVN-009's method names.
type jevRecordingTransport struct {
	mu       sync.Mutex
	requests []jevSuggestWireRequest
	scripts  [][]jev.Answer
}

func (tr *jevRecordingTransport) count() int {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return len(tr.requests)
}

func (tr *jevRecordingTransport) Do(r *http.Request) (*http.Response, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var wire jevSuggestWireRequest
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, err
	}
	tr.requests = append(tr.requests, wire)
	answers := []jev.Answer{}
	if i := len(tr.requests) - 1; i < len(tr.scripts) {
		answers = tr.scripts[i]
	}
	payload, err := json.Marshal(struct {
		Answers []jev.Answer `json:"answers"`
	}{Answers: answers})
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(payload)),
		Header:     make(http.Header),
	}, nil
}

// newFlowClient builds a gate-on client over the recording transport with a
// test-only credential. The credential value must not appear in any scripted
// state, or the secret screen would refuse the request unsent.
func newFlowClient(tr *jevRecordingTransport) *jev.Client {
	client := jev.New(true)
	client.HTTP = tr
	client.LoadCredential = func() string { return "test-only-credential" }
	return client
}

// jevSuggestFixtureCandidates is the five-skill candidate set the flow tests
// run over. The Full fields are the "fuller text" the rerank request must
// carry and the wide-rank request must not.
func jevSuggestFixtureCandidates() []jevSkillCandidate {
	return []jevSkillCandidate{
		{Name: "moai-workflow-tdd", Description: "short: test-first development", Full: "FULL-tdd red green refactor"},
		{Name: "moai-domain-frontend", Description: "short: react components", Full: "FULL-frontend component patterns"},
		{Name: "moai-ref-owasp-checklist", Description: "short: security baseline", Full: "FULL-owasp top ten"},
		{Name: "moai-workflow-project", Description: "short: project documentation", Full: "FULL-project docs"},
		{Name: "moai-meta-harness", Description: "short: legacy redirect", Full: "FULL-harness redirect"},
	}
}

// jevFixtureWideScores rank the fixture set: tdd, frontend, owasp are the top
// three; project and harness fill ranks four and five.
func jevFixtureWideScores() map[string]float64 {
	return map[string]float64{
		"moai-workflow-tdd":         0.90,
		"moai-domain-frontend":      0.80,
		"moai-ref-owasp-checklist":  0.70,
		"moai-workflow-project":     0.60,
		"moai-meta-harness":         0.50,
	}
}

// jevWideScript answers request 1: the needs-a-skill Noul plus one score per
// candidate.
func jevWideScript(cands []jevSkillCandidate, needsSkill bool, noulP float64, scores map[string]float64) []jev.Answer {
	answers := []jev.Answer{{
		QuestionID:  "needs-skill",
		Kind:        jev.KindNoul,
		Noul:        needsSkill,
		Probability: noulP,
	}}
	for _, c := range cands {
		answers = append(answers, jev.Answer{
			QuestionID:  "skill-score:" + c.Name,
			Kind:        jev.KindScore,
			Score:       scores[c.Name],
			Probability: scores[c.Name],
		})
	}
	return answers
}

// jevRerankScript answers request 2 over the top three: frontend reorders to
// first, tdd drops to second, owasp falls below both.
func jevRerankScript() []jev.Answer {
	return []jev.Answer{
		{QuestionID: "rerank:moai-domain-frontend", Kind: jev.KindScore, Score: 0.95, Probability: 0.95},
		{QuestionID: "rerank:moai-workflow-tdd", Kind: jev.KindScore, Score: 0.85, Probability: 0.85},
		{QuestionID: "rerank:moai-ref-owasp-checklist", Kind: jev.KindScore, Score: 0.40, Probability: 0.40},
	}
}

func hasQuestion(qs []jev.Question, id string, kind jev.AnswerKind) bool {
	for _, q := range qs {
		if q.ID == id && q.Kind == kind {
			return true
		}
	}
	return false
}

// isZeroSignal reports whether sig is the zero signal — the value the
// disabled path must produce, byte-for-byte the pre-Jev absence. A helper
// because the struct carries a slice and cannot use ==.
func isZeroSignal(sig jevSkillRankSignal) bool {
	return len(sig.Candidates) == 0 && !sig.NeedsSkill && !sig.Suppressed
}

// TestJevSkillSuggest_TwoRequests — AC-JEVN-009 (REQ-JEVN-012): suggestion
// issues EXACTLY two requests — the wide rank over all candidates batched with
// the needs-a-skill Noul, then the top-3 rerank under fuller text.
func TestJevSkillSuggest_TwoRequests(t *testing.T) {
	cands := jevSuggestFixtureCandidates()
	tr := &jevRecordingTransport{scripts: [][]jev.Answer{
		jevWideScript(cands, true, 0.90, jevFixtureWideScores()),
		jevRerankScript(),
	}}

	sig, ok := jevSkillSuggestionFlow(t.Context(), newFlowClient(tr), "route this turn", cands, nil)
	if !ok {
		t.Fatalf("the flow reported no signal against a scripted available endpoint (requests recorded: %d)", tr.count())
	}
	if got := tr.count(); got != 2 {
		t.Fatalf("wire requests = %d, want EXACTLY 2 (AC-JEVN-009)", got)
	}

	// Request 1: the Noul plus one score question per candidate, over the
	// short descriptions only.
	first := tr.requests[0]
	if !hasQuestion(first.Questions, "needs-skill", jev.KindNoul) {
		t.Errorf("request 1 carries no needs-a-skill Noul: %+v", first.Questions)
	}
	for _, c := range cands {
		if !hasQuestion(first.Questions, "skill-score:"+c.Name, jev.KindScore) {
			t.Errorf("request 1 carries no wide-rank question for %s — the rank is not wide", c.Name)
		}
	}
	if len(first.Questions) != len(cands)+1 {
		t.Errorf("request 1 question count = %d, want %d (one Noul + one score per candidate)", len(first.Questions), len(cands)+1)
	}
	if strings.Contains(first.State, "FULL-") {
		t.Errorf("request 1 state carries fuller text — the wide rank must run over short descriptions only:\n%s", first.State)
	}
	for _, c := range cands {
		if !strings.Contains(first.State, c.Description) {
			t.Errorf("request 1 state omits the short description of %s", c.Name)
		}
	}

	// Request 2: exactly the top three, under the fuller text.
	second := tr.requests[1]
	top3 := []string{"moai-workflow-tdd", "moai-domain-frontend", "moai-ref-owasp-checklist"}
	if len(second.Questions) != 3 {
		t.Fatalf("request 2 question count = %d, want 3 (the top-3 rerank)", len(second.Questions))
	}
	for _, name := range top3 {
		if !hasQuestion(second.Questions, "rerank:"+name, jev.KindScore) {
			t.Errorf("request 2 carries no rerank question for %s: %+v", name, second.Questions)
		}
		if !strings.Contains(second.State, "FULL-") {
			t.Fatalf("request 2 state carries no fuller text at all:\n%s", second.State)
		}
	}
	for _, c := range cands {
		isTop3 := false
		for _, name := range top3 {
			if c.Name == name {
				isTop3 = true
			}
		}
		if isTop3 {
			if !strings.Contains(second.State, c.Full) {
				t.Errorf("request 2 state omits the fuller text of top-3 candidate %s", c.Name)
			}
			continue
		}
		if strings.Contains(second.State, c.Full) {
			t.Errorf("request 2 state carries fuller text for %s, which is not in the top three", c.Name)
		}
	}

	// The signal: rerank reorders the top three, the rest keep their wide
	// positions, every rank is present and consecutive.
	want := []jevRankedSkill{
		{Skill: "moai-domain-frontend", Rank: 1, Probability: 0.95},
		{Skill: "moai-workflow-tdd", Rank: 2, Probability: 0.85},
		{Skill: "moai-ref-owasp-checklist", Rank: 3, Probability: 0.40},
		{Skill: "moai-workflow-project", Rank: 4, Probability: 0.60},
		{Skill: "moai-meta-harness", Rank: 5, Probability: 0.50},
	}
	if len(sig.Candidates) != len(want) {
		t.Fatalf("signal candidates = %d, want %d", len(sig.Candidates), len(want))
	}
	for i, w := range want {
		if sig.Candidates[i] != w {
			t.Errorf("candidate %d = %+v, want %+v", i, sig.Candidates[i], w)
		}
	}
	if !sig.NeedsSkill {
		t.Errorf("signal NeedsSkill = false, want true (the scripted Noul answered true)")
	}
}

// TestJevSkillSuggest_RankSignalShape — AC-JEVN-014 (the presentation half of
// REQ-JEVN-013): the presented object is a ranked signal carrying each
// candidate and its consecutive rank position, and the SAME inspection rejects
// a value carrying a selection field — the display-only boundary, checked.
func TestJevSkillSuggest_RankSignalShape(t *testing.T) {
	sig := jevSkillRankSignal{
		Candidates: []jevRankedSkill{
			{Skill: "moai-workflow-tdd", Rank: 1, Probability: 0.95},
			{Skill: "moai-domain-frontend", Rank: 2, Probability: 0.80},
		},
		NeedsSkill: true,
	}
	presented, err := presentJevSkillRankSignal(sig)
	if err != nil {
		t.Fatalf("present a well-formed signal: %v", err)
	}
	if err := inspectJevRankSignal(presented); err != nil {
		t.Errorf("the presented signal fails its own contract: %v", err)
	}
	// The rank positions are visible as data, not implied by ordering.
	cands, _ := presented["candidates"].([]any)
	for i, entry := range cands {
		cand := entry.(map[string]any)
		if got := cand["rank"].(int); got != i+1 {
			t.Errorf("candidate %d presents rank %d, want %d", i, got, i+1)
		}
	}

	// Positive control: the same inspection REJECTS a value carrying a
	// selection field. Without this control, an inspection that checked
	// nothing would produce the same green.
	withSelection := map[string]any{
		"candidates": presented["candidates"],
		"needs_skill": true,
		"suppressed":  false,
		"selection":   "moai-workflow-tdd",
	}
	if err := inspectJevRankSignal(withSelection); err == nil {
		t.Error("a signal carrying a selection field PASSED inspection — the display-only boundary check rejects nothing")
	}

	withDispatch := map[string]any{
		"candidates": []any{
			map[string]any{"skill": "moai-workflow-tdd", "rank": 1, "probability": 0.95, "dispatch": true},
		},
		"needs_skill": true,
		"suppressed":  false,
	}
	if err := inspectJevRankSignal(withDispatch); err == nil {
		t.Error("a candidate carrying a dispatch field PASSED inspection")
	}

	brokenRanks := map[string]any{
		"candidates": []any{
			map[string]any{"skill": "moai-workflow-tdd", "rank": 1, "probability": 0.95},
			map[string]any{"skill": "moai-domain-frontend", "rank": 1, "probability": 0.80},
		},
		"needs_skill": true,
		"suppressed":  false,
	}
	if err := inspectJevRankSignal(brokenRanks); err == nil {
		t.Error("a signal with duplicate rank positions PASSED inspection")
	}
}

// TestJevSkillSuggest_RouterAuthorityUntouched — AC-JEVN-011 (the authority
// half of REQ-JEVN-013), at the anchor level: paired enabled/disabled runs
// over the same input, with the stub-logged-at-least-one precondition, and a
// presented value that is data only — identical across runs, contract-clean,
// and carrying nothing the router could act on. The prose half of the
// authority claim is the "keeps its selection authority" clause the protocol
// block is validated against above.
func TestJevSkillSuggest_RouterAuthorityUntouched(t *testing.T) {
	cands := jevSuggestFixtureCandidates()
	run := func(enabled bool) (*jevRecordingTransport, jevSkillRankSignal, bool) {
		tr := &jevRecordingTransport{scripts: [][]jev.Answer{
			jevWideScript(cands, true, 0.90, jevFixtureWideScores()),
			jevRerankScript(),
		}}
		client := jev.New(enabled)
		client.HTTP = tr
		client.LoadCredential = func() string { return "test-only-credential" }
		sig, ok := jevSkillSuggestionFlow(t.Context(), client, "route this turn", cands, nil)
		return tr, sig, ok
	}

	trOn, sigOn, okOn := run(true)
	if !okOn || trOn.count() == 0 {
		t.Fatalf("precondition failed: the enabled run produced no logged suggestion (ok=%v, stub logged %d) — "+
			"without it the enabled/disabled comparison is satisfied by the consumer's absence", okOn, trOn.count())
	}
	trOff, sigOff, okOff := run(false)
	if okOff {
		t.Error("the disabled run produced a signal — the gate did not hold")
	}
	if trOff.count() != 0 {
		t.Errorf("the disabled run issued %d wire requests, want 0 (REQ-JEVC-017)", trOff.count())
	}
	if !isZeroSignal(sigOff) {
		t.Errorf("the disabled run's signal = %+v, want the zero value — disabled behaviour must be the pre-Jev absence", sigOff)
	}

	// The enabled run's presented value is a pure value: byte-identical across
	// two runs over the same input, and clean against the rank-signal
	// contract — no selection, no dispatch, no instruction to act.
	present, err := presentJevSkillRankSignal(sigOn)
	if err != nil {
		t.Fatalf("present the enabled run's signal: %v", err)
	}
	if err := inspectJevRankSignal(present); err != nil {
		t.Fatalf("the presented signal violates the rank-signal contract: %v", err)
	}
	first, err := json.Marshal(present)
	if err != nil {
		t.Fatalf("marshal first presentation: %v", err)
	}
	_, sigAgain, _ := run(true)
	present2, err := presentJevSkillRankSignal(sigAgain)
	if err != nil {
		t.Fatalf("present the repeat run's signal: %v", err)
	}
	second, err := json.Marshal(present2)
	if err != nil {
		t.Fatalf("marshal second presentation: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("two enabled runs over the same input presented different values — the anchor is not a pure signal source:\n%s\n%s", first, second)
	}
}

// TestJevSkillSuggest_SuppressionMechanism — the REQ-JEVN-014 branch,
// mechanism-tested with a SYNTHETIC threshold. The criterion itself
// (AC-JEVN-010) is blocked-on-threshold: no fitted threshold exists (owned by
// the measurement SPEC), so this test exercises the branch's mechanics only
// and fits nothing.
func TestJevSkillSuggest_SuppressionMechanism(t *testing.T) {
	cands := jevSuggestFixtureCandidates()
	threshold := 0.75
	run := func(threshold *float64, needsSkill bool, noulP float64) (*jevRecordingTransport, jevSkillRankSignal, bool) {
		tr := &jevRecordingTransport{scripts: [][]jev.Answer{
			jevWideScript(cands, needsSkill, noulP, jevFixtureWideScores()),
			jevRerankScript(),
		}}
		sig, ok := jevSkillSuggestionFlow(t.Context(), newFlowClient(tr), "route this turn", cands, threshold)
		return tr, sig, ok
	}

	// Negative Noul above the threshold: the ranked list is suppressed —
	// presented as no list — while the two-request shape is unchanged.
	tr, sig, ok := run(&threshold, false, 0.90)
	if !ok {
		t.Fatal("the suppressed run reported no result at all — suppression is a presentation state, not an absence")
	}
	if tr.count() != 2 {
		t.Errorf("the suppressed run issued %d requests, want 2 — REQ-JEVN-012's shape is unconditional", tr.count())
	}
	if !sig.Suppressed {
		t.Errorf("a negative Noul above threshold did not suppress the list: %+v", sig)
	}
	if len(sig.Candidates) != 0 {
		t.Errorf("the suppressed signal still carries %d candidates — a best-of-nothing would be presented", len(sig.Candidates))
	}
	presented, err := presentJevSkillRankSignal(sig)
	if err != nil {
		t.Fatalf("present the suppressed signal: %v", err)
	}
	if err := inspectJevRankSignal(presented); err != nil {
		t.Errorf("the suppressed presentation violates the rank-signal contract: %v", err)
	}

	// Negative Noul BELOW the threshold: the list presents.
	if _, sig, _ = run(&threshold, false, 0.50); sig.Suppressed || len(sig.Candidates) == 0 {
		t.Errorf("a negative Noul below threshold suppressed the list: %+v", sig)
	}
	// A positive Noul never suppresses, whatever its confidence.
	if _, sig, _ = run(&threshold, true, 0.99); sig.Suppressed || len(sig.Candidates) == 0 {
		t.Errorf("a positive Noul suppressed the list: %+v", sig)
	}
	// No threshold (the shipped state): the suppression trigger is
	// unsatisfiable and the anchor presents unsuppressed — even on a
	// confidently negative Noul, because there is no threshold to be above.
	tr, sig, ok = run(nil, false, 0.99)
	if !ok || tr.count() != 2 || sig.Suppressed || len(sig.Candidates) == 0 {
		t.Errorf("the nil-threshold run misbehaved: ok=%v requests=%d suppressed=%v candidates=%d",
			ok, tr.count(), sig.Suppressed, len(sig.Candidates))
	}
}

// installSuggestionClient replaces the client-construction seam for one test.
// The count answers "was the client even constructed" — a stronger absence
// than "no request was sent".
func installSuggestionClient(t *testing.T, client *jev.Client) *int {
	t.Helper()
	calls := 0
	prev := newJevSkillSuggestClient
	newJevSkillSuggestClient = func() *jev.Client {
		calls++
		return client
	}
	t.Cleanup(func() { newJevSkillSuggestClient = prev })
	return &calls
}

// captureJevNotice redirects the at-most-one notice writer for one test.
func captureJevNotice(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := jevNoticeWriter
	jevNoticeWriter = buf
	t.Cleanup(func() { jevNoticeWriter = prev })
	return buf
}

// TestJevSkillSuggestEntry_GateOffConstructsNothing — AC-JEVN-016 condition 2
// at the real entry path: with the gate off (the shipped default, resolved
// from the fixture's config), no client is even constructed and the recording
// transport records zero. The positive control re-resolves the SAME entry
// path with the gate on and records at least one — proving the zero above is
// the gate's doing and the transport counts.
func TestJevSkillSuggestEntry_GateOffConstructsNothing(t *testing.T) {
	root := jevProject(t, false)
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cands := jevSuggestFixtureCandidates()

	clientCalls := installSuggestionClient(t, nil)
	notice := captureJevNotice(t)

	sig, ok := liveJevSkillSuggestProbe("route this turn", cands)
	if ok {
		t.Fatal("the gate-off entry path produced a signal")
	}
	if !isZeroSignal(sig) {
		t.Errorf("gate-off signal = %+v, want the zero value", sig)
	}
	if *clientCalls != 0 {
		t.Errorf("the jev client was constructed %d times with the gate off — no request may even be armed (REQ-JEVC-017)", *clientCalls)
	}
	if got := notice.String(); got != jevSuggestDisabledNotice+"\n" {
		t.Errorf("gate-off notice = %q, want exactly the single disabled notice line", got)
	}

	// Positive control: the same entry path, gate ON, through the SAME
	// constructor seam now returning a client over the recording transport.
	if err := os.WriteFile(
		filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"),
		[]byte("workflow:\n    jev:\n        enabled: true\n"), 0o600); err != nil {
		t.Fatalf("flip the fixture gate on: %v", err)
	}
	tr := &jevRecordingTransport{scripts: [][]jev.Answer{
		jevWideScript(cands, true, 0.90, jevFixtureWideScores()),
		jevRerankScript(),
	}}
	onClient := jev.New(true)
	onClient.HTTP = tr
	onClient.LoadCredential = func() string { return "test-only-credential" }
	*clientCalls = 0
	clientCalls = installSuggestionClient(t, onClient)

	_, ok = liveJevSkillSuggestProbe("route this turn", cands)
	if !ok {
		t.Fatal("positive control failed: the gate-on entry path produced no signal")
	}
	if *clientCalls == 0 {
		t.Fatal("positive control failed: the gate-on entry path never constructed the client")
	}
	if tr.count() == 0 {
		t.Fatal("positive control failed: the recording transport recorded nothing, so the zero above is unattributable")
	}
}

// TestJevSuggestCmd_GateOffOutputIdenticalApartFromNotice — AC-JEVN-016
// condition 3: with the gate off, the command's output is byte-identical to
// the same run with the consumer's call site absent, apart from exactly one
// notice line.
func TestJevSuggestCmd_GateOffOutputIdenticalApartFromNotice(t *testing.T) {
	root := jevProject(t, false)
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	skillsPath := filepath.Join(t.TempDir(), "skills.json")
	raw, err := json.Marshal(jevSuggestFixtureCandidates())
	if err != nil {
		t.Fatalf("marshal fixture candidates: %v", err)
	}
	if err := os.WriteFile(skillsPath, raw, 0o600); err != nil {
		t.Fatalf("write fixture skills file: %v", err)
	}

	run := func(probe func(string, []jevSkillCandidate) (jevSkillRankSignal, bool)) (string, string, error) {
		t.Helper()
		prev := jevSkillSuggestProbe
		jevSkillSuggestProbe = probe
		t.Cleanup(func() { jevSkillSuggestProbe = prev })
		notice := captureJevNotice(t)
		cmd := newJevSuggestCmd()
		var out, errBuf bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errBuf)
		cmd.SetArgs([]string{"--intent", "route this turn", "--skills", skillsPath})
		execErr := cmd.Execute()
		// The notice goes to the package notice writer, not the cobra error
		// buffer; combine both so nothing the command emitted can hide.
		return out.String(), notice.String() + errBuf.String(), execErr
	}

	liveOut, liveAll, liveErr := run(liveJevSkillSuggestProbe)
	absentOut, absentAll, absentErr := run(func(string, []jevSkillCandidate) (jevSkillRankSignal, bool) {
		// The consumer's call site absent: consulted by nothing, emitting nothing.
		return jevSkillRankSignal{}, false
	})
	if liveErr != nil || absentErr != nil {
		t.Fatalf("exit errors: live=%v absent=%v — an advisory capability must not fail the path it runs on", liveErr, absentErr)
	}
	if liveOut != absentOut {
		t.Errorf("stdout differs between the gate-off run and the absent-call-site run:\n%q\n%q", liveOut, absentOut)
	}
	if liveAll != jevSuggestDisabledNotice+"\n" {
		t.Errorf("gate-off output = %q, want EXACTLY the single disabled notice line (at most one notice, REQ-JEVC-014)", liveAll)
	}
	if absentAll != "" {
		t.Errorf("the absent-call-site run emitted %q, want silence", absentAll)
	}
}

// scanJevEnabledBlocks returns every `enabled:` value found inside a `jev:`
// block in the supplied YAML text. It is the scanner behind the template
// sweep of AC-JEVN-016 condition 1: it keys on the block KEY line (a comment
// naming jev is not a block) and reads only keys indented deeper than it.
func scanJevEnabledBlocks(data string) []string {
	var values []string
	lines := strings.Split(data, "\n")
	for i, ln := range lines {
		if strings.TrimSpace(ln) != "jev:" {
			continue
		}
		keyIndent := len(ln) - len(strings.TrimLeft(ln, " \t"))
		for j := i + 1; j < len(lines); j++ {
			next := lines[j]
			trimmed := strings.TrimSpace(next)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			nextIndent := len(next) - len(strings.TrimLeft(next, " \t"))
			if nextIndent <= keyIndent {
				break
			}
			if strings.HasPrefix(trimmed, "enabled:") {
				values = append(values, strings.TrimSpace(strings.TrimPrefix(trimmed, "enabled:")))
			}
		}
	}
	return values
}

// TestJevGateUnrun_DefaultOffEverywhere — AC-JEVN-016 condition 1: the
// compiled default is false, and no shipped template under
// internal/template/templates/ sets `enabled: true` under a `jev:` key.
//
// Positive-control adaptation, recorded: the criterion's written control (the
// same pattern over the LOCAL workflow.yaml) cannot fire — this tree's local
// config carries no jev block, and adding one to make a control fire would be
// editing operator-owned state to suit a test. The control is adapted to
// fixture configs instead: the scanner is shown resolving a real `enabled:
// false` block AND reading an `enabled: true` value, so a violation in the
// template sweep above would be detected rather than silently unread.
func TestJevGateUnrun_DefaultOffEverywhere(t *testing.T) {
	// (a) The compiled default, read at its source of truth.
	if config.NewDefaultWorkflowConfig().Jev.Enabled {
		t.Error("compiled default Workflow.Jev.Enabled = true, want false (REQ-JEVC-016)")
	}

	// (b) The template sweep — non-empty by assertion, every value false.
	found := 0
	templates := filepath.Join("..", "..", "internal", "template", "templates")
	walkErr := filepath.WalkDir(templates, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		values := scanJevEnabledBlocks(string(data))
		found += len(values)
		for i, v := range values {
			if v != "false" {
				t.Errorf("%s: jev block #%d ships enabled: %s, want false", path, i+1, v)
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk the template tree: %v", walkErr)
	}
	if found == 0 {
		t.Fatal("the template sweep found no jev: block at all — an empty sweep asserts nothing")
	}

	// (c) The adapted positive control, both directions.
	if got := scanJevEnabledBlocks("workflow:\n    jev:\n        enabled: false\n"); len(got) != 1 || got[0] != "false" {
		t.Errorf("the scanner cannot resolve a real enabled: false block: %v — the sweep above is unattributable", got)
	}
	if got := scanJevEnabledBlocks("workflow:\n    jev:\n        enabled: true\n"); len(got) != 1 || got[0] != "true" {
		t.Errorf("the scanner cannot read an enabled: true value: %v — a template violation would go undetected", got)
	}
}

// TestJevSuggestCmd_ArgumentValidationAndHappyPath — the command surface's
// argument errors fire before the consumer is ever consulted, and the gate-on
// happy path emits the ranked-signal JSON on stdout, contract-clean: only the
// contract keys, no selection or dispatch field, at the wire level.
func TestJevSuggestCmd_ArgumentValidationAndHappyPath(t *testing.T) {
	run := func(args ...string) (string, error) {
		t.Helper()
		cmd := newJevSuggestCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs(args)
		// Capture the output AFTER Execute, never in the return expression —
		// out.String() there would be evaluated before the command runs.
		execErr := cmd.Execute()
		return out.String(), execErr
	}

	skillsPath := filepath.Join(t.TempDir(), "skills.json")
	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(skillsPath, []byte(content), 0o600); err != nil {
			t.Fatalf("write skills fixture: %v", err)
		}
	}
	// Every argument error below fires BEFORE the probe: no config read, no
	// client, no request.
	if _, err := run(); err == nil {
		t.Error("no --intent and no --skills must be an argument error")
	}
	if _, err := run("--intent", "route this turn"); err == nil {
		t.Error("a missing --skills must be an argument error")
	}
	if _, err := run("--intent", "route this turn", "--skills", filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("an unreadable skills file must be an error")
	}
	write("[]")
	if _, err := run("--intent", "route this turn", "--skills", skillsPath); err == nil {
		t.Error("an empty candidate list must be an error")
	}
	write(`[{"name":"a"},{"name":"a"}]`)
	if _, err := run("--intent", "route this turn", "--skills", skillsPath); err == nil {
		t.Error("duplicate candidate names must be an error")
	}
	write(`[{"description":"no name here"}]`)
	if _, err := run("--intent", "route this turn", "--skills", skillsPath); err == nil {
		t.Error("an unnamed candidate must be an error")
	}
	over := make([]jevSkillCandidate, 0, jevSkillSuggestCandidateLimit+1)
	for i := 0; i <= jevSkillSuggestCandidateLimit; i++ {
		over = append(over, jevSkillCandidate{Name: fmt.Sprintf("skill-%02d", i)})
	}
	rawOver, err := json.Marshal(over)
	if err != nil {
		t.Fatalf("marshal over-limit fixture: %v", err)
	}
	write(string(rawOver))
	if _, err := run("--intent", "route this turn", "--skills", skillsPath); err == nil {
		t.Error("an over-limit candidate list must be an error — refused whole, never truncated")
	}

	// Happy path: gate ON, the signal JSON on stdout, nothing else. The rank
	// contract is asserted at the wire level (after a JSON round trip, ranks
	// are numbers, not ints) — what matters here is that ONLY the contract
	// keys reach the command's stdout.
	root := jevProject(t, true)
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cands := jevSuggestFixtureCandidates()
	tr := &jevRecordingTransport{scripts: [][]jev.Answer{
		jevWideScript(cands, true, 0.90, jevFixtureWideScores()),
		jevRerankScript(),
	}}
	onClient := jev.New(true)
	onClient.HTTP = tr
	onClient.LoadCredential = func() string { return "test-only-credential" }
	installSuggestionClient(t, onClient)

	raw, err := json.Marshal(cands)
	if err != nil {
		t.Fatalf("marshal fixture candidates: %v", err)
	}
	write(string(raw))
	out, err := run("--intent", "route this turn", "--skills", skillsPath)
	if err != nil {
		t.Fatalf("the gate-on happy path failed: %v", err)
	}
	if tr.count() != 2 {
		t.Errorf("the happy path issued %d wire requests, want 2", tr.count())
	}
	var presented map[string]any
	if err := json.Unmarshal([]byte(out), &presented); err != nil {
		t.Fatalf("stdout is not the signal JSON: %v (%q)", err, out)
	}
	for key := range presented {
		switch key {
		case "candidates", "needs_skill", "suppressed":
		default:
			t.Errorf("stdout carries a non-contract field %q — nothing but the ranked signal may reach the orchestrator", key)
		}
	}
	candsOut, ok := presented["candidates"].([]any)
	if !ok || len(candsOut) != len(cands) {
		t.Fatalf("stdout candidates = %T len-mismatch, want the full ranked list", presented["candidates"])
	}
	for i, entry := range candsOut {
		cand, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("candidate %d is %T, want an object", i, entry)
		}
		for key := range cand {
			switch key {
			case "skill", "rank", "probability":
			default:
				t.Errorf("candidate %d carries a non-contract field %q at the wire level", i, key)
			}
		}
	}
}

// TestInspectJevRankSignal_RejectsMalformedShapes — the inspection's error
// branches are each observed rejecting their shape, so the green paths above
// are read against an instrument known to fire.
func TestInspectJevRankSignal_RejectsMalformedShapes(t *testing.T) {
	cases := map[string]map[string]any{
		"missing-candidates":      {"needs_skill": true, "suppressed": false},
		"candidates-not-a-list":   {"candidates": "nope", "needs_skill": true, "suppressed": false},
		"candidate-not-an-object": {"candidates": []any{"nope"}, "needs_skill": true, "suppressed": false},
		"rank-not-an-int":         {"candidates": []any{map[string]any{"skill": "a", "rank": "1", "probability": 0.9}}, "needs_skill": true, "suppressed": false},
		"unnamed-candidate":       {"candidates": []any{map[string]any{"rank": 1, "probability": 0.9}}, "needs_skill": true, "suppressed": false},
	}
	for name, shape := range cases {
		if err := inspectJevRankSignal(shape); err == nil {
			t.Errorf("malformed shape %q PASSED inspection — the contract check rejects nothing", name)
		}
	}
}

// TestLiveJevSkillSuggestClient_WiresTheProductionClient — the production
// constructor is never reached by the seam-replacing tests above, so its one
// load-bearing property is asserted directly: it builds a gate-ON client with
// the credential reader wired, because the ENTRY path is what resolves the
// gate before constructing.
func TestLiveJevSkillSuggestClient_WiresTheProductionClient(t *testing.T) {
	client := liveJevSkillSuggestClient()
	if client == nil {
		t.Fatal("the production constructor returned nil")
	}
	if !client.Enabled {
		t.Error("the production constructor built a gate-OFF client — the entry path already resolves the gate; a second off here would be dead silence")
	}
	if client.LoadCredential == nil {
		t.Error("the production constructor left the credential reader unwired — every call would degrade to no-credential")
	}
}

// TestJevGateUnrun_NotPresentedAsAvailable — AC-JEVN-016 condition 4: no
// published surface (CHANGELOG, README) presents the consumer. The positive
// control names a shipped capability on the same surfaces, so a search that
// read nothing could not produce the zeros above.
func TestJevGateUnrun_NotPresentedAsAvailable(t *testing.T) {
	const consumer = "jev-suggest"
	surfaces := []string{"../../CHANGELOG.md", "../../README.md", "../../README.ko.md"}
	controlHits := 0
	surfacesRead := 0
	for _, path := range surfaces {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("read %s: %v", path, err)
		}
		surfacesRead++
		body := string(data)
		if strings.Contains(body, consumer) {
			t.Errorf("%s names %q — the consumer is in the gate-unrun state and no published surface may present it", path, consumer)
		}
		if strings.Contains(body, "moai doctor") {
			controlHits++
		}
	}
	if surfacesRead == 0 {
		t.Fatal("no published surface could be read — nothing was measured")
	}
	if controlHits == 0 {
		t.Fatal("positive control failed: no read surface names the shipped capability `moai doctor`, so the absences above are unmeasured")
	}
}
