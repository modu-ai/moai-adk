// jev_skill_suggest.go — SPEC-JEV-CONSUMERS-001 M6, Consumer B: the
// skill-suggestion signal the /moai intent router MAY consult when the
// `workflow.jev.enabled` gate is on.
//
// What this file is for, stated as the thing it must NOT become: the answer is
// a RANKED SIGNAL a reader consults, never a selection anything acts on. The
// intent router keeps its selection authority unchanged; nothing here
// dispatches, answers, closes, or selects. The consumer is bounded by the same
// display-only chain Consumer C is (the "letting the display become the
// verdict" defect is what the whole chain is bounded against).
//
// THE SHIPPING GATE IS NOT MET — the consumer is in the gate-unrun state
// (REQ-JEVN-016): the measurement gate is runnable and has NOT been run, so
// the code is present but unreachable at the shipped default. The sibling
// premise-death task is the reason the bar is stated that way: it measured
// 58.9% where a constant answer scored 75.0%, and a consumer that looks
// plausible is exactly the one that gets shipped unmeasured. Until the gate
// runs, the conditions below are what keep a present-but-off consumer
// equivalent to an absent one:
//
//   - **Off by default, everywhere.** The compiled default and every shipped
//     template keep `workflow.jev.enabled` false; while off, the entry path
//     resolves the gate FIRST and constructs no client, so no request is
//     armed (REQ-JEVC-017). The `jev-suggest` subcommand is Hidden so no
//     user-facing surface presents the consumer as available.
//
//   - **Behaviourally absent.** With the gate off the command emits exactly
//     one notice line and nothing else; the byte-compare against the absent
//     call site is AC-JEVN-016 condition 3's subject.
//
//   - **The suppression threshold is nil on every production path.** No
//     fitted threshold exists (owned by the measurement SPEC), so REQ-JEVN-014
//     is unsatisfiable and the anchor presents unsuppressed. The suppression
//     branch is mechanism-tested with a synthetic threshold only; passing a
//     fitted value here before the measurement chain produces one is out of
//     scope by SPEC exclusion.
//
// The two-request shape is deliberate and unconditional (REQ-JEVN-012): the
// wide rank needs every skill's short description and would blow the state
// budget with full text, while the rerank needs fuller text on three
// candidates and can afford it. The needs-a-skill Noul is the load-bearing
// half, not the ranking — a ranker over a fixed set always returns a best,
// including for turns that need none, and suppressing the list when the Noul
// answers negatively above threshold is what stops the consumer manufacturing
// suggestions out of turns that had no use for one.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/spf13/cobra"
)

// jevSuggestTimeout bounds the whole two-request flow. The command is an
// explicit invocation, not a locked hot path, so the bound is generous — but
// an unbounded call would still hang the invoking session on a degraded
// endpoint, and a timeout lands on Unreachable, a recorded absence rather
// than a failure.
const jevSuggestTimeout = 15 * time.Second

// jevSkillSuggestCandidateLimit bounds the candidate list. Accuracy falls as
// irrelevant state grows, the request size bounds refuse whole rather than
// truncating, and a skills catalogue is the one input here a caller controls.
const jevSkillSuggestCandidateLimit = 64

// Question-ID scheme. The IDs are part of the wire contract the verification
// surface asserts (AC-JEVN-009's "captures each request's shape").
const (
	jevNeedsSkillQuestionID = "needs-skill"
	jevWideScorePrefix      = "skill-score:"
	jevRerankScorePrefix    = "rerank:"
)

// jevSuggestDisabledNotice is the single line the gate-off entry path emits —
// the "at most one notice line" AC-JEVN-016 condition 3 byte-compares against.
const jevSuggestDisabledNotice = "jev unavailable (disabled): workflow.jev.enabled is false; no request constructed"

// jevSuggestRerankCount is how many candidates the second request reranks
// under fuller text.
const jevSuggestRerankCount = 3

// jevSkillCandidate is one entry of the caller-supplied skills list. Full is
// the fuller text only the rerank request carries; when absent the short
// description stands in.
type jevSkillCandidate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Full        string `json:"full,omitempty"`
}

// jevRankedSkill is one presented candidate: the skill, its 1-based rank
// position, and the model's probability. There is no field here a router
// could act on — the type IS the display-only contract.
type jevRankedSkill struct {
	Skill       string  `json:"skill"`
	Rank        int     `json:"rank"`
	Probability float64 `json:"probability"`
}

// jevSkillRankSignal is the Consumer B answer: a ranked list (empty when the
// needs-a-skill Noul suppressed it), the Noul's own verdict, and whether the
// list was suppressed.
type jevSkillRankSignal struct {
	Candidates []jevRankedSkill `json:"candidates"`
	NeedsSkill bool             `json:"needs_skill"`
	Suppressed bool             `json:"suppressed"`
}

// presentJevSkillRankSignal renders the signal as the object handed to the
// orchestrator and validates it against the rank-signal contract before
// returning it. The anchor's own output passes the same inspection the
// positive control exercises — a presentation that could not fail inspection
// would make the contract check decorative.
func presentJevSkillRankSignal(sig jevSkillRankSignal) (map[string]any, error) {
	cands := make([]any, 0, len(sig.Candidates))
	for _, c := range sig.Candidates {
		cands = append(cands, map[string]any{
			"skill":       c.Skill,
			"rank":        c.Rank,
			"probability": c.Probability,
		})
	}
	presented := map[string]any{
		"candidates":  cands,
		"needs_skill": sig.NeedsSkill,
		"suppressed":  sig.Suppressed,
	}
	if err := inspectJevRankSignal(presented); err != nil {
		return nil, err
	}
	return presented, nil
}

// inspectJevRankSignal validates the presented object against the rank-signal
// contract (AC-JEVN-014): only the contract keys at both levels, each
// candidate carrying its consecutive 1-based rank, and — the clause the
// positive control exercises — no selection, dispatch, or action field of any
// kind. A value carrying one is not a display-only signal, and this
// inspection is what refuses it.
func inspectJevRankSignal(presented map[string]any) error {
	for key := range presented {
		switch key {
		case "candidates", "needs_skill", "suppressed":
		default:
			return fmt.Errorf("rank-signal contract: unexpected top-level field %q", key)
		}
	}
	raw, ok := presented["candidates"]
	if !ok {
		return fmt.Errorf("rank-signal contract: missing %q", "candidates")
	}
	cands, ok := raw.([]any)
	if !ok {
		return fmt.Errorf("rank-signal contract: %q is %T, want a list", "candidates", raw)
	}
	for i, entry := range cands {
		cand, ok := entry.(map[string]any)
		if !ok {
			return fmt.Errorf("rank-signal contract: candidate %d is %T, want an object", i, entry)
		}
		for key := range cand {
			switch key {
			case "skill", "rank", "probability":
			default:
				return fmt.Errorf("rank-signal contract: candidate %d carries unexpected field %q", i, key)
			}
		}
		rank, ok := cand["rank"].(int)
		if !ok {
			return fmt.Errorf("rank-signal contract: candidate %d rank is %T, want int", i, cand["rank"])
		}
		if rank != i+1 {
			return fmt.Errorf("rank-signal contract: candidate %d carries rank %d, want %d (consecutive from 1)", i, rank, i+1)
		}
		if name, _ := cand["skill"].(string); name == "" {
			return fmt.Errorf("rank-signal contract: candidate %d carries no skill name", i)
		}
	}
	return nil
}

// @MX:NOTE: gate-unrun Consumer B anchor (SPEC-JEV-CONSUMERS-001 REQ-JEVN-016) —
// present in the tree, unreachable at the shipped default (workflow.jev.enabled
// is false), and owed its measurement gate (fitted threshold Q3, owned by
// SPEC-JEV-OPTIN-MEASURE-001) before any consumer may ship. Not an error path.
// jevSkillSuggestionFlow is the mechanical anchor the protocol block names:
// the two-request flow through the existing CORE client. Request 1 is the
// wide rank over every candidate batched with the needs-a-skill Noul (one
// state, one request); request 2 reranks the top three under fuller text.
//
// suppressThreshold is the fitted threshold REQ-JEVN-014 consumes. Every
// production caller passes nil — no fitted threshold exists — which makes the
// suppression trigger unsatisfiable and the anchor present unsuppressed. The
// parameter exists so the branch is real rather than described, and so the
// measurement chain can supply the value without editing this call path.
//
// The returned bool reports whether a signal was produced. Every absence —
// gate, malformed, credential, size, secret, transport — lands on false with
// at most one notice line, never an error: an error here would propagate into
// the caller's exit status and turn a missing model answer into a failure.
func jevSkillSuggestionFlow(ctx context.Context, client *jev.Client, intent string, cands []jevSkillCandidate, suppressThreshold *float64) (jevSkillRankSignal, bool) {
	if client == nil || strings.TrimSpace(intent) == "" || len(cands) == 0 {
		return jevSkillRankSignal{}, false
	}
	ctx, cancel := context.WithTimeout(ctx, jevSuggestTimeout)
	defer cancel()

	wide := jev.Request{
		State:     jevSkillWideState(intent, cands),
		Questions: append(jevSkillScoreQuestions(jevWideScorePrefix, cands), jevNeedsSkillQuestion()),
	}
	res := client.Ask(ctx, wide)
	if !res.OK() {
		jevNotice(res.NoticeLine())
		return jevSkillRankSignal{}, false
	}
	wideScores, needsSkill, noulProbability := jevSkillWideAnswers(res.Answers)
	if needsSkill == nil {
		// The Noul is the load-bearing half. Without it there is no answer
		// about whether this turn needs a skill, and presenting a ranking
		// anyway would manufacture a suggestion out of an unanswered question.
		return jevSkillRankSignal{}, false
	}

	allNames := make([]string, 0, len(cands))
	for _, c := range cands {
		allNames = append(allNames, c.Name)
	}
	wideOrder := jevSkillOrder(allNames, wideScores)
	if len(wideOrder) > jevSuggestRerankCount {
		wideOrder = wideOrder[:jevSuggestRerankCount]
	}
	byName := make(map[string]jevSkillCandidate, len(cands))
	for _, c := range cands {
		byName[c.Name] = c
	}
	top3 := make([]jevSkillCandidate, 0, len(wideOrder))
	for _, name := range wideOrder {
		top3 = append(top3, byName[name])
	}

	rerank := jev.Request{
		State:     jevSkillRerankState(intent, top3),
		Questions: jevSkillScoreQuestions(jevRerankScorePrefix, top3),
	}
	res2 := client.Ask(ctx, rerank)
	if !res2.OK() {
		jevNotice(res2.NoticeLine())
		return jevSkillRankSignal{}, false
	}
	rerankScores := jevSkillScoreAnswers(res2.Answers, jevRerankScorePrefix)

	// Suppression is presentation-only: the two-request shape is unconditional
	// (REQ-JEVN-012), and what REQ-JEVN-014 governs is the list the
	// orchestrator is shown, never the requests behind it.
	signal := jevSkillAssembleSignal(cands, wideScores, wideOrder, rerankScores)
	if suppressThreshold != nil && !*needsSkill && noulProbability > *suppressThreshold {
		return jevSkillRankSignal{Suppressed: true}, true
	}
	return signal, true
}

// jevNeedsSkillQuestion is the load-bearing Noul: whether the turn needs a
// skill at all. Without it a ranker over a fixed set always returns a best —
// including for turns that need none.
func jevNeedsSkillQuestion() jev.Question {
	return jev.Question{
		ID:   jevNeedsSkillQuestionID,
		Text: "Does this turn need a skill at all? Answer false when the orchestrator can proceed well without one.",
		Kind: jev.KindNoul,
	}
}

// jevSkillScoreQuestions builds one score question per candidate under the
// given ID prefix ("skill-score:" for the wide rank, "rerank:" for the rerank).
func jevSkillScoreQuestions(prefix string, cands []jevSkillCandidate) []jev.Question {
	qs := make([]jev.Question, 0, len(cands))
	for _, c := range cands {
		qs = append(qs, jev.Question{
			ID:   prefix + c.Name,
			Text: fmt.Sprintf("How relevant is the skill %q to this turn? Score from 0 (irrelevant) to 1 (directly on task).", c.Name),
			Kind: jev.KindScore,
		})
	}
	return qs
}

// jevSkillWideState renders request 1's state: the intent plus every
// candidate's SHORT description — the wide rank must not pay for full text.
func jevSkillWideState(intent string, cands []jevSkillCandidate) string {
	var b strings.Builder
	b.WriteString("The orchestrator is routing one turn and may consult a skill-suggestion signal.\n\nTURN INTENT\n")
	b.WriteString(intent)
	b.WriteString("\n\nCANDIDATE SKILLS (name: short description)\n")
	for _, c := range cands {
		fmt.Fprintf(&b, "%s: %s\n", c.Name, c.Description)
	}
	return b.String()
}

// jevSkillRerankState renders request 2's state: the intent plus the top
// candidates' FULLER text — the rerank can afford what the wide rank cannot.
func jevSkillRerankState(intent string, top []jevSkillCandidate) string {
	var b strings.Builder
	b.WriteString("The orchestrator is routing one turn; the candidates below scored highest on the wide rank.\n\nTURN INTENT\n")
	b.WriteString(intent)
	b.WriteString("\n\nTOP CANDIDATES (name: full description)\n")
	for _, c := range top {
		full := c.Full
		if strings.TrimSpace(full) == "" {
			full = c.Description
		}
		fmt.Fprintf(&b, "%s: %s\n", c.Name, full)
	}
	return b.String()
}

// jevSkillWideAnswers reads request 1's answers: a score per candidate plus
// the Noul. The Noul is returned as a pointer so a MISSING answer is
// distinguishable from a false one — the former is no signal at all, the
// latter is a judgment.
func jevSkillWideAnswers(answers []jev.Answer) (map[string]float64, *bool, float64) {
	scores := jevSkillScoreAnswers(answers, jevWideScorePrefix)
	var needs *bool
	var probability float64
	for _, a := range answers {
		if a.QuestionID == jevNeedsSkillQuestionID && a.Kind == jev.KindNoul {
			v := a.Noul
			needs = &v
			probability = a.Probability
		}
	}
	return scores, needs, probability
}

// jevSkillScoreAnswers collects score answers under the given ID prefix,
// keyed by the candidate name that follows it.
func jevSkillScoreAnswers(answers []jev.Answer, prefix string) map[string]float64 {
	scores := make(map[string]float64)
	for _, a := range answers {
		if a.Kind != jev.KindScore || !strings.HasPrefix(a.QuestionID, prefix) {
			continue
		}
		scores[strings.TrimPrefix(a.QuestionID, prefix)] = a.Score
	}
	return scores
}

// jevSkillOrder orders names by score, highest first, ties broken by name so
// the presented ranking is deterministic across identical runs.
func jevSkillOrder(names []string, scores map[string]float64) []string {
	out := append([]string(nil), names...)
	sort.SliceStable(out, func(a, b int) bool {
		if scores[out[a]] != scores[out[b]] {
			return scores[out[a]] > scores[out[b]]
		}
		return out[a] < out[b]
	})
	return out
}

// jevSkillAssembleSignal builds the final ranking: the reranked top three in
// rerank order (ranks 1-3, carrying their rerank probabilities), then the
// remainder in wide-rank order (ranks 4..N, carrying their wide scores).
func jevSkillAssembleSignal(cands []jevSkillCandidate, wideScores map[string]float64, wideOrder []string, rerankScores map[string]float64) jevSkillRankSignal {
	rerankSet := make(map[string]bool, len(wideOrder))
	for _, name := range wideOrder {
		rerankSet[name] = true
	}
	rerankOrd := jevSkillOrder(wideOrder, rerankScores)

	allNames := make([]string, 0, len(cands))
	for _, c := range cands {
		allNames = append(allNames, c.Name)
	}
	rest := make([]string, 0, len(allNames))
	for _, name := range jevSkillOrder(allNames, wideScores) {
		if !rerankSet[name] {
			rest = append(rest, name)
		}
	}

	final := append(append([]string(nil), rerankOrd...), rest...)
	out := make([]jevRankedSkill, 0, len(final))
	for i, name := range final {
		probability := wideScores[name]
		if rerankSet[name] {
			probability = rerankScores[name]
		}
		out = append(out, jevRankedSkill{Skill: name, Rank: i + 1, Probability: probability})
	}
	return jevSkillRankSignal{Candidates: out, NeedsSkill: true}
}

// jevSkillSuggestProbe is the Consumer B seam — the consumer's call site, the
// one surface the gate-off byte-compare stubs out (AC-JEVN-016 condition 3)
// and the command invokes. Tests replace it the way installJevProbe replaces
// Consumer C's.
var jevSkillSuggestProbe = liveJevSkillSuggestProbe

// newJevSkillSuggestClient is the client-construction seam. The count of its
// calls is what proves the gate-off entry path constructs NO client at all —
// a stronger absence than "no request was sent" (AC-JEVN-016 condition 2).
var newJevSkillSuggestClient = liveJevSkillSuggestClient

// liveJevSkillSuggestClient is the production construction: the gate is
// passed through the client too, so a request is refused unsent even if a
// future caller skips the entry-path resolution (the package-level gate, the
// one place to forget it).
func liveJevSkillSuggestClient() *jev.Client {
	client := jev.New(true)
	client.LoadCredential = jevcred.Load
	return client
}

// liveJevSkillSuggestProbe is the production seam body: resolve the gate for
// the invoking project, then run the two-request flow with no fitted
// threshold. Every absence returns the zero signal with at most one notice
// line.
func liveJevSkillSuggestProbe(intent string, cands []jevSkillCandidate) (jevSkillRankSignal, bool) {
	enabled, err := jevEnabled(resolveProjectDir())
	if err != nil || !enabled {
		// Unlike Consumer C's admission-path probe — deliberately silent
		// because it sits on the hot path — this seam sits behind an EXPLICIT
		// invocation, so the single disabled notice is the honest answer to a
		// command that was asked to run and could not. The line is a named
		// constant because AC-JEVN-016 condition 3 byte-compares against it.
		jevNotice(jevSuggestDisabledNotice)
		return jevSkillRankSignal{}, false
	}
	return jevSkillSuggestionFlow(context.Background(), newJevSkillSuggestClient(), intent, cands, nil)
}

// loadJevSkillCandidates reads and validates the caller-supplied skills file.
func loadJevSkillCandidates(path string) ([]jevSkillCandidate, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- caller-supplied path, resolved at invocation
	if err != nil {
		return nil, fmt.Errorf("read skills file: %w", err)
	}
	var cands []jevSkillCandidate
	if err := json.Unmarshal(raw, &cands); err != nil {
		return nil, fmt.Errorf("parse skills file: %w", err)
	}
	if len(cands) == 0 {
		return nil, fmt.Errorf("skills file carries no candidates")
	}
	if len(cands) > jevSkillSuggestCandidateLimit {
		return nil, fmt.Errorf("skills file carries %d candidates, over the %d bound; refused rather than truncated", len(cands), jevSkillSuggestCandidateLimit)
	}
	seen := make(map[string]bool, len(cands))
	for _, c := range cands {
		if strings.TrimSpace(c.Name) == "" {
			return nil, fmt.Errorf("skills file carries a candidate with no name")
		}
		if seen[c.Name] {
			return nil, fmt.Errorf("skills file repeats candidate %q", c.Name)
		}
		seen[c.Name] = true
	}
	return cands, nil
}

// newJevSuggestCmd is Consumer B's mechanical anchor as an invocable surface:
// `moai jev-suggest --intent <text> --skills <skills.json>`. Hidden — while
// the consumer sits in the gate-unrun state, no user-facing surface may
// present it (REQ-JEVN-016 condition iii); the /moai skill's gated protocol
// block is the only surface that names the command, and it names it behind
// the gate. Exit semantics follow the display-only chain: every absence is a
// notice and exit 0, never a failure.
func newJevSuggestCmd() *cobra.Command {
	var intent, skillsPath string
	cmd := &cobra.Command{
		Use:    "jev-suggest",
		Short:  "Model-produced skill-suggestion signal (gated; default off)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(intent) == "" {
				return fmt.Errorf("--intent is required: pass the turn's raw input text")
			}
			if skillsPath == "" {
				return fmt.Errorf("--skills is required: pass a JSON file of {name, description[, full]} candidates")
			}
			cands, err := loadJevSkillCandidates(skillsPath)
			if err != nil {
				return err
			}
			sig, ok := jevSkillSuggestProbe(intent, cands)
			if !ok {
				// The at-most-one notice line was already emitted by the probe.
				// Exit 0: an advisory capability never fails the path it runs on
				// (REQ-JEVC-007).
				return nil
			}
			presented, err := presentJevSkillRankSignal(sig)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(presented)
		},
	}
	cmd.Flags().StringVar(&intent, "intent", "", "the turn's raw intent text")
	cmd.Flags().StringVar(&skillsPath, "skills", "", "JSON file listing candidate skills ({name, description[, full]})")
	return cmd
}
