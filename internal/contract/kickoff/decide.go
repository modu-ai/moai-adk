package kickoff

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

// Decide reason codes (event line and --json only; never in the receipt).
const (
	ReasonDeciderHuman          = "decider-human"
	ReasonAuthorConflict        = "author-decider-conflict"
	ReasonAuthorUnmeasured      = "author-check-unmeasured"
	ReasonCrossCheckDisagree    = "cross-check-disagree"
	ReasonJevDoctrineNotAmended = "jev-doctrine-not-amended"
)

// Judgement is the deciding LLM's answer (--judgement). Agent and Model are
// the decider's self-declared identity; moai cannot verify them.
type Judgement struct {
	Agent      string   `json:"agent"`
	Model      string   `json:"model"`
	Answer     string   `json:"answer"`
	Confidence *float64 `json:"confidence"`
	Reason     string   `json:"reason"`
	ReasonRefs []string `json:"reason_refs"`
}

var (
	answers         = []string{"approve", "reject", "escalate"}
	reasonRefRe     = regexp.MustCompile(`^contract\.yaml:([1-9][0-9]*)$`)
	planAuditNameRe = regexp.MustCompile(`^plan-audit-(?:iter)?([0-9]+)\.md$`)
	// authoredByAgentLine reads the `Authored-By-Agent:` trailer anywhere in
	// a commit body, line by line (the same rule as the SPEC lint's ownership
	// reader); git's trailer parser misses it when a signature line follows.
	authoredByAgentLine = regexp.MustCompile(`(?mi)^\s*Authored-By-Agent:\s*(\S+)\s*$`)
	tierThreshold       = map[string]float64{"S": 0.75, "M": 0.80, "L": 0.85}
)

// Decide evaluates the kickoff preconditions and outcome rules for one SPEC
// and card and records the decision in the contract store. A deciding path
// (rules R1-R4) appends an issued receipt to receipts.jsonl, a decide event to
// events.jsonl, and writes the same receipt body to
// .moai/specs/<SPEC-ID>/kickoff-receipt.json. A non-deciding path (decider
// human, a failed precondition, author exclusion) appends only the decide
// event. It never signs, never edits the contract, SPEC documents, the queue,
// or git state, and calls no LLM: the judgement arrives as input.
//
// Errors write nothing: ErrIntegrity (exit 1) for a broken store chain;
// ErrUsage (exit 2) for usage, a card mismatch, decider jev (R5), a malformed
// judgement, or I/O.
//
// @MX:ANCHOR: [AUTO] The only issuer of kickoff receipts.
// @MX:REASON: the decide CLI, the kickoff-check receipt lookup, and the A1
// receipt signer all consume what it writes; rule order R1-R5 and the
// deciding/non-deciding split are the published contract.
func Decide(in DecideInput) (DecideResult, error) {
	res := DecideResult{Preconditions: []receipt.Precondition{}}
	if !contract.ValidSpecID(in.SpecID) || in.Card == "" || in.Root == "" {
		return res, fmt.Errorf("%w: need a SPEC ID, a card id, and a project root", ErrUsage)
	}
	now := in.Now
	if now == nil {
		now = time.Now
	}
	st := in.Store
	if st == nil {
		var err error
		if st, err = receipt.Open(in.Root); err != nil {
			return res, fmt.Errorf("%w: %v", ErrUsage, err)
		}
	}
	if err := st.Verify(); err != nil {
		return res, err
	}

	dir, err := contract.ResolveSpecDir(in.Root, in.SpecID)
	if err != nil {
		return res, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	cin, err := contract.LoadDir(dir)
	if err != nil {
		return res, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	c, err := contract.Decode(cin.Contract)
	if err != nil {
		return res, fmt.Errorf("%w: contract.yaml does not decode: %v", ErrUsage, err)
	}
	if c.Card != in.Card {
		return res, fmt.Errorf("%w: %w: the contract's card is %q, not %q", ErrUsage, ErrCardMismatch, c.Card, in.Card)
	}
	if in.Config.DeciderJevSole || in.Config.Decider == "jev" {
		return res, fmt.Errorf("%w: %w: Jev is never a sole decider (R5)", ErrUsage, ErrDeciderJevRefused)
	}

	ev := receipt.DecideEvent{Spec: in.SpecID, Card: in.Card, RequestedDecider: in.Config.Decider,
		AuthorTrailers: []string{}, Preconditions: []receipt.Precondition{}}
	res.RequestedDecider = in.Config.Decider

	if in.Config.Decider != contract.DeciderLLM && in.Config.Decider != contract.DeciderLLMJev {
		return nonDeciding(st, &res, ev, "", ReasonDeciderHuman)
	}

	cin.Policy = in.Policy
	cin.RegistryRuleIDs, cin.RegistryFrozenFiles = in.RegistryRuleIDs, in.RegistryFrozenFiles
	rep := contract.Verify(cin)
	pre, planAudit := preconditions(in, dir, c, rep)
	res.Preconditions, ev.Preconditions = pre, pre
	for _, p := range pre {
		if !p.OK {
			return nonDeciding(st, &res, ev, "", "precondition:"+p.Name)
		}
	}

	j, err := parseJudgement(in.Judgement, contract.ContractLineCount(cin.Contract))
	if err != nil {
		return res, err
	}
	ev.DeciderAgent, ev.DeciderModel = j.Agent, j.Model

	bodies, err := commitBodies(in, dir)
	if err != nil {
		return res, fmt.Errorf("%w: read SPEC commit history: %v", ErrUsage, err)
	}
	authors := trailerAuthors(bodies)
	ev.AuthorTrailers = authors
	switch {
	case strings.EqualFold(j.Agent, "manager-spec") || slices.Contains(authors, strings.ToLower(j.Agent)):
		return nonDeciding(st, &res, ev, "", ReasonAuthorConflict)
	case len(authors) == 0:
		return nonDeciding(st, &res, ev, "", ReasonAuthorUnmeasured)
	}

	llm := contract.ReceiptAnswer{Answer: j.Answer, Confidence: j.Confidence, Reason: j.Reason, ReasonRefs: j.ReasonRefs}
	r := contract.KickoffReceipt{
		ReceiptVersion:   1,
		SpecID:           in.SpecID,
		RequestedDecider: in.Config.Decider,
		EffectiveDecider: contract.DeciderLLM,
		Fallback:         &contract.ReceiptFallback{Applied: false},
		IssuedAt:         now().UTC().Format(time.RFC3339),
		Inputs: &contract.ReceiptInputs{
			ContractSHA256:   rep.SignableContractSHA256,
			AcceptanceSHA256: rep.Acceptance.MeasuredSHA256,
			PlanAuditReport:  planAudit,
		},
		LLMAnswer: &llm,
	}
	var rd receipt.ReceiptData
	rule, reason := "R4", ""
	if in.Config.Decider == contract.DeciderLLMJev {
		req, jres := askJev(in, c, j)
		rd.JevRequest, _ = json.Marshal(req)
		rd.JevResponse, _ = json.Marshal(jres)
		ja, fallback := jevVerdict(in, jres)
		if fallback != "" {
			rule = "R2"
			r.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &fallback}
			res.FallbackReason = fallback
			r.Outcome = llmOutcome(j.Answer)
		} else {
			r.EffectiveDecider = contract.DeciderLLMJev
			r.JevAnswer = &contract.ReceiptJevAnswer{
				ReceiptAnswer: contract.ReceiptAnswer{Answer: ja.Choice, Confidence: &ja.Probability,
					Reason: "Jev cross-check signal (" + jev.ModelID + ").", ReasonRefs: j.ReasonRefs},
				RequestSHA256: receipt.SHA256Hex(rd.JevRequest),
				RawResponse:   string(rd.JevResponse),
			}
			switch {
			case !in.DoctrineAmended:
				rule, reason, r.Outcome = "R3", ReasonJevDoctrineNotAmended, "human"
			case j.Answer == "approve" && ja.Choice == "approve":
				rule, r.Outcome = "R1", "approve"
			case j.Answer == "reject" && ja.Choice == "reject":
				rule, r.Outcome = "R1", "reject"
			default:
				rule, reason, r.Outcome = "R1", ReasonCrossCheckDisagree, "human"
			}
		}
	} else {
		r.Outcome = llmOutcome(j.Answer)
	}
	return deciding(in, st, &res, ev, r, rd, rule, reason)
}

func llmOutcome(answer string) string {
	switch answer {
	case "approve":
		return "approve"
	case "reject":
		return "reject"
	}
	return "human"
}

// nonDeciding appends the decide event of a path that decided nothing.
func nonDeciding(st *receipt.Store, res *DecideResult, ev receipt.DecideEvent, rule, reason string) (DecideResult, error) {
	ev.Rule, ev.Outcome, ev.Reason = rule, "human", reason
	if _, err := st.AppendEvent(receipt.KindDecide, ev); err != nil {
		return *res, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	res.Outcome, res.Reason, res.Rule = "human", reason, rule
	return *res, nil
}

// deciding appends the issued receipt and its decide event, then writes the
// receipt file.
func deciding(in DecideInput, st *receipt.Store, res *DecideResult, ev receipt.DecideEvent,
	r contract.KickoffReceipt, rd receipt.ReceiptData, rule, reason string) (DecideResult, error) {
	body, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return *res, fmt.Errorf("%w: encode receipt: %v", ErrUsage, err)
	}
	body = append(body, '\n')
	rd.Spec, rd.Card, rd.Body, rd.BodySHA256 = in.SpecID, in.Card, string(body), receipt.SHA256Hex(body)
	line, err := st.AppendReceipt(rd)
	if err != nil {
		return *res, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	ev.ReceiptLine, ev.Rule, ev.Outcome, ev.Reason = line.Hash, rule, r.Outcome, reason
	ev.EffectiveDecider = r.EffectiveDecider
	if _, err := st.AppendEvent(receipt.KindDecide, ev); err != nil {
		return *res, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	path := filepath.Join(in.Root, ".moai", "specs", in.SpecID, contract.ReceiptFile)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return *res, fmt.Errorf("%w: write %s: %v", ErrUsage, path, err)
	}
	res.Outcome, res.Reason, res.Rule = r.Outcome, reason, rule
	res.EffectiveDecider = r.EffectiveDecider
	res.ReceiptPath, res.ReceiptSHA256 = path, rd.BodySHA256
	return *res, nil
}

// parseJudgement strictly decodes and validates the judgement.
func parseJudgement(raw []byte, lines int) (Judgement, error) {
	var j Judgement
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&j); err != nil {
		return j, fmt.Errorf("%w: judgement: %v", ErrUsage, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return j, fmt.Errorf("%w: judgement: trailing data", ErrUsage)
	}
	bad := strings.TrimSpace(j.Agent) == "" || !slices.Contains(answers, j.Answer) ||
		j.Confidence == nil || *j.Confidence < 0 || *j.Confidence > 1 ||
		strings.TrimSpace(j.Reason) == "" || len(j.ReasonRefs) == 0
	for _, ref := range j.ReasonRefs {
		m := reasonRefRe.FindStringSubmatch(ref)
		if m == nil {
			bad = true
			continue
		}
		if n, err := strconv.Atoi(m[1]); err != nil || n > lines {
			bad = true
		}
	}
	if bad {
		return j, fmt.Errorf("%w: judgement needs agent, answer (approve|reject|escalate), confidence in [0,1], reason, and contract.yaml:<line> reason_refs", ErrUsage)
	}
	return j, nil
}

// preconditions evaluates (a)-(f) in order. The plan-audit reference is
// returned for the receipt inputs.
func preconditions(in DecideInput, dir string, c *contract.Contract, rep contract.Report) ([]receipt.Precondition, *contract.ReceiptFileRef) {
	var out []receipt.Precondition
	add := func(name string, ok bool, detail string) {
		out = append(out, receipt.Precondition{Name: name, OK: ok, Detail: detail})
	}

	ref, detail := planAuditCheck(in, dir)
	add("a", ref != nil && detail == "", detail)

	markers := 0
	for _, name := range []string{"plan.md", "research.md"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			markers += strings.Count(string(data), "[NEEDS CLARIFICATION")
		}
	}
	add("b", markers == 0, fmt.Sprintf("%d open clarification markers", markers))

	var other []string
	forbidden := false
	for _, r := range rep.Reasons {
		switch r {
		case contract.ReasonUnsigned:
		case contract.ReasonForbiddenAction:
			forbidden = true
		default:
			other = append(other, r)
		}
	}
	add("c", len(other) == 0, strings.Join(other, ","))
	add("d", !forbidden, "")

	eOK, eDetail := true, ""
	if nd, err := escalation.NeedsDecision(in.Root, in.Card); err != nil || nd {
		eOK, eDetail = false, "open escalation record"
	}
	if c.Signature != nil {
		if blocked, err := revokeReader(in.Root, in.Card, in.SpecID, c.Signature.Seal); err != nil || blocked {
			eOK, eDetail = false, "revocation of the current signature"
		}
	}
	add("e", eOK, eDetail)

	var clash []string
	if c.Ownership != nil {
		for _, w := range c.Ownership.Write {
			for _, fz := range rep.FrozenFiles {
				if w == fz || contract.MatchGlob(w, fz) || contract.MatchGlob(fz, w) {
					clash = append(clash, w)
					break
				}
			}
		}
	}
	add("f", len(clash) == 0, strings.Join(clash, ","))
	return out, ref
}

// planAuditCheck picks the highest-N plan-audit report in the card evidence
// path and checks its verdict, score threshold, and binding to the current
// plan artifacts. detail is "" when (a) holds.
func planAuditCheck(in DecideInput, dir string) (*contract.ReceiptFileRef, string) {
	reportDir := filepath.Join(in.Root, ".moai", "reports", in.Card)
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		return nil, "no plan-audit report"
	}
	best, bestN := "", -1
	for _, e := range entries {
		m := planAuditNameRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, _ := strconv.Atoi(m[1]); n > bestN {
			best, bestN = e.Name(), n
		}
	}
	if best == "" {
		return nil, "no plan-audit report"
	}
	path := filepath.Join(reportDir, best)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "unreadable plan-audit report"
	}
	rel := filepath.ToSlash(filepath.Join(".moai", "reports", in.Card, best))
	ref := &contract.ReceiptFileRef{Path: rel, SHA256: receipt.SHA256Hex(data)}

	verdict, score, hash, scoreOK := "", 0.0, "", false
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		key, value, ok := strings.Cut(strings.TrimSpace(sc.Text()), ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), " ", "_"))
		value = strings.TrimSpace(strings.Trim(strings.TrimSpace(value), "`"))
		switch key {
		case "verdict":
			verdict = value
		case "overall_score", "score":
			f, err := strconv.ParseFloat(value, 64)
			score, scoreOK = f, err == nil
		case "plan_artifact_hash", "plan-artifact-hash":
			hash = value
		}
	}
	if verdict != "PASS" && verdict != "PASS-WITH-DEBT" {
		return ref, "verdict " + verdict
	}
	if !scoreOK {
		return ref, "unparseable Overall Score"
	}
	if score < tierThreshold[specTier(dir)] {
		return ref, fmt.Sprintf("score %.3f below the tier threshold", score)
	}
	cur, err := runtime.NewInMemoryCache().ComputeHash(dir)
	if err != nil || hash == "" || hash != cur {
		return ref, "plan_artifact_hash does not bind the current plan artifacts"
	}
	return ref, ""
}

// specTier reads spec.md's tier (absent → L, the backward-compatible default).
func specTier(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "spec.md"))
	if err != nil {
		return "L"
	}
	for _, l := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(l), "tier:"); ok {
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if _, known := tierThreshold[v]; known {
				return v
			}
		}
	}
	return "L"
}

// commitBodies returns the bodies of the commits that touched the SPEC
// directory.
func commitBodies(in DecideInput, dir string) ([]string, error) {
	rel, err := filepath.Rel(in.Root, dir)
	if err != nil {
		return nil, err
	}
	rel = filepath.ToSlash(rel)
	if in.CommitBodies != nil {
		return in.CommitBodies(in.Root, rel)
	}
	cmd := exec.Command("git", "-C", in.Root, "log", "--format=%B%x00", "--", rel)
	cmd.Env = gitenv.Env()
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var bodies []string
	for _, b := range strings.Split(string(out), "\x00") {
		if strings.TrimSpace(b) != "" {
			bodies = append(bodies, b)
		}
	}
	return bodies, nil
}

// trailerAuthors is the sorted set of Authored-By-Agent values.
func trailerAuthors(bodies []string) []string {
	set := []string{}
	for _, b := range bodies {
		for _, m := range authoredByAgentLine.FindAllStringSubmatch(b, -1) {
			v := strings.ToLower(strings.TrimSpace(m[1]))
			if !slices.Contains(set, v) {
				set = append(set, v)
			}
		}
	}
	slices.Sort(set)
	return set
}

// askJev asks the start question over the contract summary and the LLM's
// conclusion.
func askJev(in DecideInput, c *contract.Contract, j Judgement) (jev.Request, jev.Result) {
	state := fmt.Sprintf("SPEC contract %s (card %s).\nApproach: %s\nActions: %s\nLLM decider answer: %s — %s",
		c.SpecID, c.Card, c.Approach, strings.Join(c.Actions, ", "), j.Answer, j.Reason)
	req := jev.Request{State: state, Questions: []jev.Question{{
		ID: "start", Text: "Should the run start under this signed contract?",
		Kind: jev.KindChoice, Choices: answers,
	}}}
	ask := in.Jev
	if ask == nil {
		ask = defaultJev(in.Config.JevEnabled)
	}
	return req, ask(context.Background(), req)
}

// jevVerdict maps the Jev result to its answer, or to a fallback reason.
func jevVerdict(in DecideInput, r jev.Result) (jev.Answer, string) {
	if !in.Config.JevEnabled {
		return jev.Answer{}, "jev_disabled"
	}
	switch r.Availability {
	case jev.Available:
	case jev.Disabled:
		return jev.Answer{}, "jev_disabled"
	case jev.NoCredential, jev.Unauthorized:
		return jev.Answer{}, "jev_key_missing"
	case jev.Malformed:
		return jev.Answer{}, "jev_malformed_response"
	default:
		return jev.Answer{}, "jev_call_failed"
	}
	for _, a := range r.Answers {
		if a.QuestionID != "start" {
			continue
		}
		if !slices.Contains(answers, a.Choice) || a.Probability < 0 || a.Probability > 1 {
			return jev.Answer{}, "jev_malformed_response"
		}
		if a.Probability < in.Config.JevMinConfidence {
			return jev.Answer{}, "jev_low_confidence"
		}
		return a, ""
	}
	return jev.Answer{}, "jev_malformed_response"
}
