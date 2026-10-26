// Package jev is the single implementation of the TypeSafe System One call
// path (REQ-JEVC-001): request construction, credential read, HTTP transport,
// response decoding, and usage accounting.
//
// @MX:ANCHOR: [AUTO] TypeSafe System One call-path SSOT — the only client implementation
// @MX:REASON: a second client would double the size bounds, the fail-open policy, the secret screening, and the request alias with its attribution, and the four would drift independently (REQ-JEVC-002)
//
// Three properties shape the whole package, and each is the reason a more
// obvious design was rejected:
//
//   - **An unavailable result is a VALUE, not an error.** An error return
//     propagates, and a propagated error becomes a non-zero exit somewhere —
//     precisely the failure REQ-JEVC-007 forbids. Returning a value makes
//     degradation the default path: a caller that ignores Availability gets
//     "no answer", which every consumer must already handle because a disabled
//     capability produces it. The symmetry every later consumer leans on:
//     absence of an answer is NOT a negative answer. It is no signal at all.
//
//   - **The gate sits inside the package, below every caller.** REQ-JEVC-017
//     requires a disabled capability to construct no request at all; a gate at
//     the call sites is N places to forget, and the "no request constructed"
//     property would then need testing N times instead of once.
//
//   - **The capability is display-only.** A Jev answer is a labelled,
//     model-produced signal a person reads (REQ-JEVC-013). Nothing in this
//     package writes a file, mutates a queue, or touches git — it has no
//     dependency that could (see display_only_test.go, which asserts the
//     package's entire import set is standard library). Two in-process
//     consumers outside this package use an answer as more than a display:
//     the Kickoff cross-check, and the todo --auto cycle's own candidate
//     ranking, the auto-scoped ranking exception. That ranking runs behind
//     the default-off workflow.jev.enabled gate, and an answer there sets
//     selection order only: the candidate set is fixed by mechanical filters
//     before any answer is read, and the answer is never the basis of a
//     completion verdict, a merge approval or a queue mutation. Neither
//     consumer reaches the MCP tool, and no ordering accuracy is claimed.
//
// The package depends only on the standard library so it cannot participate in
// an import cycle and can be imported from both internal/cli and internal/web.
// Configuration reaches it as a plain bool through New; it does not import
// internal/config.
package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ModelID is the request alias every request carries: the model value the
// package sends. It names the vendor's alias, not a version, and it is never a
// citation of the model that answered. The answering model is the interpreted
// model id the response reports, which ValidInterpretedModel checks and which
// every result, usage record, and answer label carries.
//
// It is a compiled constant rather than operator-visible configuration, per the
// repository rule that model names, endpoint URLs, and API headers live in Go
// constants. The alias is what the request asks for. The attribution comes from
// the response, not from this constant, so a vendor move of the version behind
// the alias shows up in the record instead of being silently re-attributed. Its
// value is asserted by TestRequestModel_IsTheAliasNotAVersion.
const ModelID = "jev-latest"

// VerifiedInterpretedModel is the interpreted model id the measurement
// regression set was last run against. Measured 2026-10-15: the request alias
// resolved to this version at that time. It changes only by a code change made
// after the regression set has been re-run against the new id. An answered call
// reporting any other id earns the client one change notice.
const VerifiedInterpretedModel = "jev-1.13.0"

// interpretedModelShape is the vendor's versioned model id: the family name and
// three dot-separated numbers, as in jev-1.13.0. The bare family name (jev-1.13)
// and the request alias (jev-latest) do not match, because neither names the
// model that answered.
var interpretedModelShape = regexp.MustCompile(`^jev-[0-9]+\.[0-9]+\.[0-9]+$`)

// ValidInterpretedModel reports whether s is an interpreted model id in the
// versioned shape. A response whose model is not valid yields no attributable
// answer: its result is Unattributed (REQ-JRM-003).
func ValidInterpretedModel(s string) bool {
	return interpretedModelShape.MatchString(s)
}

// EndpointURL is the System One endpoint. Compiled, for the same reason the
// model id is: no configuration path may aim a request somewhere else.
const EndpointURL = "https://api.typesafe.ai/v1/systemone"

// Request size bounds (REQ-JEVC-009). The vendor limits are 64k tokens for the
// whole request and 32k for the state plus the longest question. A request
// exceeding either is REFUSED rather than truncated: a truncated state is a
// silently different question, and the answer to a different question is worse
// than no answer.
const (
	// MaxStateAndQuestionTokens bounds the state plus the single longest question.
	MaxStateAndQuestionTokens = 32_000
	// MaxRequestTokens bounds the whole request — state plus every question.
	MaxRequestTokens = 64_000
)

// bytesPerTokenEstimate is the divisor behind EstimateTokens. Four bytes per
// token is the usual English-text rule of thumb; applying it to UTF-8 BYTES
// rather than runes makes the estimate conservative for multi-byte scripts (a
// Korean or Japanese state estimates HIGH, so the bound refuses earlier). The
// error direction matters: over-estimating costs a refusal the caller can see,
// while under-estimating ships a request the vendor rejects.
const bytesPerTokenEstimate = 4

// SignalLabel is the marker every surfaced answer carries, so a reader can tell
// a model-produced signal from a mechanical measurement and from a human or
// agent judgement (REQ-JEVC-013).
const SignalLabel = "model signal"

// Availability names the observed condition of a call. It is deliberately an
// enumeration rather than a boolean: the doctor check (REQ-JEVC-022) must
// distinguish "disabled" from "unreachable", and a consumer reporting a notice
// line has to say WHICH absence it observed.
type Availability string

const (
	// Available — the call returned answers.
	Available Availability = "available"
	// Disabled — workflow.jev.enabled is false. No request was constructed.
	Disabled Availability = "disabled"
	// NoCredential — the credential file is absent, unreadable, or empty.
	NoCredential Availability = "no-credential"
	// Unauthorized — the transport returned HTTP 401.
	Unauthorized Availability = "unauthorized"
	// RateLimited — the transport returned HTTP 429.
	RateLimited Availability = "rate-limited"
	// Overloaded — the transport returned HTTP 529.
	Overloaded Availability = "overloaded"
	// Unreachable — the network failed, or the response could not be read.
	Unreachable Availability = "unreachable"
	// Oversize — the request exceeded a size bound and was refused unsent.
	Oversize Availability = "oversize"
	// SecretDetected — the payload carried a credential-shaped token and was
	// refused unsent (REQ-JEVC-021).
	SecretDetected Availability = "secret-detected"
	// Malformed — the request carried no question to ask.
	Malformed Availability = "malformed"
	// Unattributed — the endpoint answered OK, but its response names no
	// interpreted model id in the versioned shape. No answer is returned: an
	// answer that cannot be attributed to a version is not an answer (REQ-JRM-003).
	Unattributed Availability = "unattributed"
)

// AnswerKind is the typed shape of an answer. System One answers a typed
// question; it generates no text.
type AnswerKind string

const (
	// KindChoice — one option from a supplied set.
	KindChoice AnswerKind = "choice"
	// KindNoul — a yes/no judgement.
	KindNoul AnswerKind = "noul"
	// KindScore — a bounded numeric score.
	KindScore AnswerKind = "score"
)

// UserAgent identifies this client on every request. It is a compiled constant
// for the same reason the endpoint is. The vendor's edge refuses some default
// client User-Agents (an HTTP 403 carrying "error code: 1010" was observed for
// Python's urllib default; Go's default was observed to pass), so the client
// names itself rather than depending on whatever default its transport sends.
const UserAgent = "moai-adk-jev/1 (+https://github.com/modu-ai/moai-adk)"

// Question is one typed question asked over the request's state.
//
// This is the caller-facing shape. The wire shape differs — the vendor takes
// questions as a map keyed by ID, with Text sent as `instructions` and the
// options or levels sent as `criteria` — and the translation happens inside
// this package (see toWire), so no caller depends on the wire layout.
type Question struct {
	ID   string     `json:"question_id"`
	Text string     `json:"question"`
	Kind AnswerKind `json:"kind,omitempty"` // empty means KindNoul
	// Choices are the options of a KindChoice question (required for it).
	Choices []string `json:"choices,omitempty"`
	// Levels are the ordered score levels of a KindScore question, lowest
	// first (required for it).
	Levels []string `json:"levels,omitempty"`
}

// Request carries ONE state and a LIST of questions (REQ-JEVC-010). The list is
// in the type rather than left to caller discipline: input tokens are charged
// and output tokens are not, so cost tracks state size rather than answer
// count — and, independently of cost, N separate requests over the same state
// could return mutually inconsistent judgements over identical input.
type Request struct {
	State     string     `json:"state"`
	Questions []Question `json:"questions"`

	// Repeatable declares that repeating this exact call is provably free of
	// side effects (REQ-JEVC-008). It defaults to FALSE, which is the
	// fail-safe direction: an undeclared request is never retried. Setting it
	// true permits a bounded retry, and even then ONLY when the failure is
	// provably pre-delivery — an ambiguous failure is never retried whatever
	// this field says.
	Repeatable bool `json:"-"`
}

// Answer is one typed answer.
//
// Probability is the answer's headline number, and its meaning depends on the
// kind: for a KindNoul it IS the answer — the model's probability that the
// statement holds (the vendor's `noul` value, in [0, 1]); there is no boolean,
// and any yes/no cut is the caller's threshold to choose. For a KindChoice or
// KindScore it is the vendor's `confidence` in the chosen option or score.
type Answer struct {
	QuestionID  string     `json:"question_id"`
	Kind        AnswerKind `json:"kind"`
	Choice      string     `json:"choice,omitempty"`
	Score       float64    `json:"score,omitempty"`
	Probability float64    `json:"probability"`
	// Model is the interpreted model id the response reported for the call that
	// produced this answer: the version that answered. Ask sets it on every answer
	// it returns, and it is empty only for an answer built by hand.
	Model string `json:"model,omitempty"`
	// Probabilities is the per-option (choice) or per-level-index (score)
	// distribution the vendor returned. Empty for a noul.
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	// Legend maps a score's level index to its description. Score only.
	Legend map[string]string `json:"legend,omitempty"`
}

// Label renders the answer as a labelled model signal (REQ-JEVC-013). It names
// the marker and the interpreted model that answered, and deliberately avoids
// any word that reads as a mechanical measurement — a reader must be able to
// tell this apart from something that was measured. An answer with no
// interpreted model renders as "unattributed", never as the request alias.
func (a Answer) Label() string {
	model := a.Model
	if model == "" {
		model = "unattributed"
	}
	prefix := fmt.Sprintf("[%s · %s] %s", SignalLabel, model, a.QuestionID)
	switch a.Kind {
	case KindChoice:
		return fmt.Sprintf("%s = %s (confidence=%.2f)", prefix, a.Choice, a.Probability)
	case KindScore:
		return fmt.Sprintf("%s = %s (confidence=%.2f)", prefix, strconv.FormatFloat(a.Score, 'f', -1, 64), a.Probability)
	default:
		return fmt.Sprintf("%s: p(true)=%.2f", prefix, a.Probability)
	}
}

// Usage is the per-call record REQ-JEVC-004 requires: the input-token count the
// call was charged for, the interpreted model id that answered, and the request
// alias that was sent (REQ-JRM-004). The two model fields are held apart, so a
// later cost or provenance question is answerable from the record and a reader
// cannot take the alias for the version that answered.
type Usage struct {
	InputTokens int `json:"input_tokens"`
	// Model is the interpreted model id the response reported: the version that
	// answered. It is empty when the call was not answered (refused before the
	// send, unavailable, or unattributed).
	Model string `json:"model"`
	// RequestModel is the request alias this call sent. It is recorded in its own
	// field, separate from Model.
	RequestModel string `json:"request_model"`
	// FromCache marks a record whose answer was served from the disk cache
	// (REQ-FW-013): no network request was made. The new-vs-cache separation
	// lives in this flag so a cost question is answerable from the log alone.
	FromCache bool `json:"from_cache"`
}

// Result is what every call returns. There is no error return: see the package
// doc — an error would propagate into someone's exit status.
type Result struct {
	Availability Availability
	// Condition names what was observed, in prose, for the notice line and the
	// doctor check. It is empty only when Availability is Available.
	Condition string
	Answers   []Answer
	Usage     Usage
	// Notice is set only on the change notice: the first answered call of a
	// client whose interpreted model id differs from VerifiedInterpretedModel.
	Notice string
	// StaleCacheWarning is set when a cache serve came from a prior-model
	// entry after the alias moved past it (REQ-FW-013, keep-and-warn): the
	// answer is still the model's own answer, and the entry is kept, but the
	// reader is told which version produced it.
	StaleCacheWarning string
}

// OK reports whether the call produced answers.
func (r Result) OK() bool { return r.Availability == Available }

// NoticeLine renders the at-most-one notice line a consumer may emit
// (REQ-JEVC-007). It returns the change notice when one is set, then the
// cache staleness warning, then the unavailability condition. An available
// result with neither notice nor warning returns empty.
func (r Result) NoticeLine() string {
	if r.Notice != "" {
		return r.Notice
	}
	if r.StaleCacheWarning != "" {
		return r.StaleCacheWarning
	}
	if r.Availability == Available {
		return ""
	}
	cond := strings.ReplaceAll(r.Condition, "\n", " ")
	if cond == "" {
		return fmt.Sprintf("jev unavailable (%s)", r.Availability)
	}
	return fmt.Sprintf("jev unavailable (%s): %s", r.Availability, cond)
}

// Doer is the transport seam. Tests inject a fake or an httptest server client;
// no test in this package contacts the real endpoint.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client owns the single call path.
type Client struct {
	// Enabled mirrors workflow.jev.enabled. It reaches the package as a plain
	// bool so the package stays standard-library-only; the caller resolves the
	// config value (one bridge, in the doctor check, for this milestone).
	Enabled bool
	// HTTP is the transport. Nil means http.DefaultClient.
	HTTP Doer
	// LoadCredential reads the stored credential. It is a function rather than
	// a string so a client constructed once re-reads a credential written
	// later, and so tests never touch a real home directory.
	LoadCredential func() string
	// Endpoint defaults to EndpointURL.
	Endpoint string
	// MaxRetries bounds retries of a request that is BOTH declared repeatable
	// and failed provably before delivery. Every other failure is never
	// retried, whatever this holds. The Retry-After backoff consumes the SAME
	// budget (REQ-FW-012 composing with REQ-JEVC-008, not overriding it).
	MaxRetries int

	// CacheDir enables the disk cache when non-empty (REQ-FW-011): responses
	// cached under a body-SHA + resolved-model-id key, served without a
	// network request. Production callers set it from DefaultCacheDir()
	// (~/.moai/cache/jev); tests ALWAYS override it with a temp dir.
	CacheDir string
	// ModelOverride is the explicit per-request interpretation model id for
	// the cache key — step 1 of the lookup sequence (plan B5). Empty means
	// alias resolution from the persisted map. An override's cache miss goes
	// to the network — never another model's entry (the N5 applicability
	// condition) — and the entry a network answer writes is attributed to
	// the model id the response reported, which is the override's own
	// exactly when the endpoint answers with it.
	ModelOverride string
	// MaxConcurrency caps concurrently in-flight requests (REQ-FW-012's
	// semaphore limb). Zero means no cap; New sets a default.
	MaxConcurrency int

	mu    sync.Mutex
	usage []Usage
	// changeNoticed records that this client already issued the change notice.
	// Guarded by mu.
	changeNoticed bool
	// sem is the lazily-built concurrency semaphore. Guarded by mu.
	sem chan struct{}
}

// New returns a client for the given gate state, wired to the compiled endpoint.
// The credential reader is left nil deliberately: this package does not import
// internal/jevcred, so the caller supplies it (jevcred.Load) and the two
// packages stay independently testable.
func New(enabled bool) *Client {
	return &Client{
		Enabled:        enabled,
		Endpoint:       EndpointURL,
		MaxRetries:     1,
		MaxConcurrency: defaultMaxConcurrency,
	}
}

// UsageLog returns a copy of the per-call usage records.
func (c *Client) UsageLog() []Usage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Usage, len(c.usage))
	copy(out, c.usage)
	return out
}

func (c *Client) record(u Usage) {
	c.mu.Lock()
	c.usage = append(c.usage, u)
	c.mu.Unlock()
}

// EstimateTokens returns a conservative token estimate for s. It is an
// ESTIMATE, not a tokenizer: it divides UTF-8 byte length, which over-estimates
// multi-byte scripts, so the size bounds refuse early rather than late.
func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + bytesPerTokenEstimate - 1) / bytesPerTokenEstimate
}

// wireRequest is the on-the-wire payload. The request alias is written here and
// nowhere else, so every request this package constructs carries the same one.
//
// The layout follows the vendor's documented schema: `questions` is an OBJECT
// keyed by the caller's question id (the id is not a field), and each entry
// carries `type`, `instructions`, and — for choice and score — `criteria`. A
// list-shaped `questions` is rejected by the endpoint with HTTP 422.
type wireRequest struct {
	Model     string                  `json:"model"`
	State     string                  `json:"state"`
	Questions map[string]wireQuestion `json:"questions"`
}

// wireQuestion is one documented question entry. Criteria is an object
// {option: description} for a choice, an ordered level list for a score, and
// absent for a noul (whose text travels as instructions).
type wireQuestion struct {
	Type         AnswerKind `json:"type"`
	Instructions string     `json:"instructions"`
	Criteria     any        `json:"criteria,omitempty"`
}

// wireResponse is the documented response: `answers` is an object keyed by the
// question id the caller sent, and `model` names the version that answered. The
// model is read here and attributed in Ask; it is never assumed.
type wireResponse struct {
	Model   string                `json:"model"`
	Answers map[string]wireAnswer `json:"answers"`
	Usage   struct {
		InputTokens int `json:"input_tokens"`
	} `json:"usage"`
}

// wireAnswer is one documented answer. Noul is a probability, not a boolean.
type wireAnswer struct {
	Type          AnswerKind         `json:"type"`
	Noul          float64            `json:"noul"`
	Choice        string             `json:"choice"`
	Score         float64            `json:"score"`
	Probabilities map[string]float64 `json:"probabilities"`
	Legend        map[string]string  `json:"legend"`
	Confidence    float64            `json:"confidence"`
}

// kindOf resolves a question's kind, defaulting an unset one to noul.
func kindOf(q Question) AnswerKind {
	if q.Kind == "" {
		return KindNoul
	}
	return q.Kind
}

// validateQuestions refuses what the documented schema cannot express, so it
// is refused unsent rather than rejected (and charged) by the endpoint. Ids
// become map keys, so an empty or repeated id would silently merge questions.
func validateQuestions(qs []Question) (string, bool) {
	seen := make(map[string]bool, len(qs))
	for i, q := range qs {
		if q.ID == "" {
			return fmt.Sprintf("question %d has an empty id", i), false
		}
		if seen[q.ID] {
			return fmt.Sprintf("question id %q is repeated; ids key the request", q.ID), false
		}
		seen[q.ID] = true
		switch kindOf(q) {
		case KindNoul:
		case KindChoice:
			if len(q.Choices) == 0 {
				return fmt.Sprintf("choice question %q carries no options", q.ID), false
			}
		case KindScore:
			if len(q.Levels) == 0 {
				return fmt.Sprintf("score question %q carries no levels", q.ID), false
			}
		default:
			return fmt.Sprintf("question %q has unknown kind %q", q.ID, q.Kind), false
		}
	}
	return "", true
}

// toWire translates the caller-facing request into the documented wire shape.
// A choice option's description is the option text itself: the caller-facing
// type carries no separate descriptions.
func toWire(req Request) wireRequest {
	qs := make(map[string]wireQuestion, len(req.Questions))
	for _, q := range req.Questions {
		wq := wireQuestion{Type: kindOf(q), Instructions: q.Text}
		switch wq.Type {
		case KindChoice:
			criteria := make(map[string]string, len(q.Choices))
			for _, c := range q.Choices {
				criteria[c] = c
			}
			wq.Criteria = criteria
		case KindScore:
			wq.Criteria = q.Levels
		}
		qs[q.ID] = wq
	}
	return wireRequest{Model: ModelID, State: req.State, Questions: qs}
}

// fromWire flattens the answer map into the caller-facing list, in the
// request's question order; answers for ids the request did not carry follow,
// sorted by id, so the result is deterministic. Every answer carries the
// interpreted model id the response reported.
func fromWire(questions []Question, answers map[string]wireAnswer, model string) []Answer {
	out := make([]Answer, 0, len(answers))
	convert := func(id string, w wireAnswer) Answer {
		a := Answer{QuestionID: id, Kind: w.Type, Model: model, Probabilities: w.Probabilities}
		switch w.Type {
		case KindNoul:
			a.Probability = w.Noul
			a.Probabilities = nil
		case KindChoice:
			a.Choice, a.Probability = w.Choice, w.Confidence
		case KindScore:
			a.Score, a.Probability, a.Legend = w.Score, w.Confidence, w.Legend
		}
		return a
	}
	used := make(map[string]bool, len(questions))
	for _, q := range questions {
		if w, ok := answers[q.ID]; ok {
			out = append(out, convert(q.ID, w))
			used[q.ID] = true
		}
	}
	extra := make([]string, 0)
	for id := range answers {
		if !used[id] {
			extra = append(extra, id)
		}
	}
	sort.Strings(extra)
	for _, id := range extra {
		out = append(out, convert(id, answers[id]))
	}
	return out
}

// answersComplete reports whether every requested question id carries an
// answer (N1, t1968 sync-audit iter6). An OK response that answered only
// part of the request is a transient endpoint fault, not a verdict — it is
// neither stored nor admitted as a cache hit, so the fault cannot persist
// past the outage. Extra answers for ids the request did not carry are not
// incompleteness.
func answersComplete(questions []Question, answers []Answer) bool {
	have := make(map[string]bool, len(answers))
	for _, a := range answers {
		have[a.QuestionID] = true
	}
	for _, q := range questions {
		if !have[q.ID] {
			return false
		}
	}
	return true
}

// Ask runs one call. The order of checks is load-bearing:
//
//  1. the GATE, because a disabled capability must construct no request at all
//     and must report "disabled" rather than whatever else is also missing;
//  2. the question list, because an empty one has nothing to ask and one the
//     documented schema cannot express would only be rejected by the endpoint;
//  3. the CREDENTIAL, because an absent one is an absence rather than a failure;
//  4. the SIZE bounds, before the send, so an oversize payload is refused whole;
//  5. the secret SCREEN, also before the send, so a credential-bearing payload
//     never leaves the process.
//
// Every one of the five returns a typed unavailable result. None returns an
// error, and none panics.
//
// Once the request is sent, the result is decided on the response. An OK
// response whose interpreted model id is not in the versioned shape is
// Unattributed and returns no answers. Every sent call is recorded in the usage
// log, whatever its outcome, with the request alias and, only when the call was
// answered, the interpreted model id.
func (c *Client) Ask(ctx context.Context, req Request) Result {
	if !c.Enabled {
		return Result{
			Availability: Disabled,
			Condition:    "workflow.jev.enabled is false; no request constructed",
		}
	}
	if len(req.Questions) == 0 {
		return Result{Availability: Malformed, Condition: "request carries no question"}
	}
	if reason, ok := validateQuestions(req.Questions); !ok {
		return Result{Availability: Malformed, Condition: reason}
	}

	credential := ""
	if c.LoadCredential != nil {
		credential = c.LoadCredential()
	}
	if credential == "" {
		return Result{
			Availability: NoCredential,
			Condition:    "no TypeSafe credential is stored; the capability degrades to no signal",
		}
	}

	if reason, over := checkBounds(req); over {
		return Result{Availability: Oversize, Condition: reason}
	}
	if reason, hit := ScreenPayload(req, credential); hit {
		return Result{Availability: SecretDetected, Condition: reason}
	}

	body, err := json.Marshal(toWire(req))
	if err != nil {
		return Result{Availability: Malformed, Condition: "request could not be encoded: " + err.Error()}
	}

	// Disk cache (REQ-FW-011): a hit is served with no network request at
	// all. The lookup key is the body SHA combined with the RESOLVED model
	// id — never the request alias. Cache errors are fail-open: a damaged or
	// unreadable cache degrades to a network call, never to a failed one.
	sha := bodySHA(body)
	if c.CacheDir != "" {
		if keyModel, aliasResolved, resolvable := c.resolveModel(); resolvable {
			// N1 serve leg: an entry that does not answer every requested
			// question (the shape a pre-repair client could have written) is
			// not admitted as a hit — the lookup misses and the call goes to
			// the network for a complete answer.
			if entry, hit, stale, _ := c.cacheLookup(sha, keyModel, aliasResolved); hit && answersComplete(req.Questions, entry.Answers) {
				served := serveCacheHit(entry, stale)
				// N2: the change notice applies to cache serves too, against
				// the ENTRY's model — consumers of a served answer learn the
				// model-versus-verified divergence the same way the network
				// path reports it.
				served.Notice = c.changeNotice(entry.Model)
				u := entry.Usage
				u.RequestModel = ModelID
				u.FromCache = true
				c.record(u)
				served.Usage = u
				return served
			}
		}
	}

	res, answers, model := c.send(ctx, body, req.Repeatable)
	// Attribution is decided on the response, never assumed. An OK response that
	// names no valid interpreted model has no attributable answer, so none is
	// returned (REQ-JRM-003).
	if res.OK() {
		if ValidInterpretedModel(model) {
			res.Answers = fromWire(req.Questions, answers, model)
			res.Notice = c.changeNotice(model)
			if c.CacheDir != "" {
				// Stored under the model id the RESPONSE reported — the
				// only authoritative attribution. The wire carries the
				// request alias even under a pinned override, so the
				// response id is what the call actually resolved to;
				// keying a mismatched answer under the override would
				// serve later override lookups a wrong-model answer as a
				// clean hit (card t1968 review N2). The entry is the
				// override's own exactly when the endpoint answers with
				// the override id; the alias map still moves only on
				// alias-resolved observations.
				//
				// N1 (t1968 sync-audit iter6): only COMPLETE answers are
				// stored. An OK response missing some requested ids is a
				// transient endpoint fault; caching it would persist the
				// fault — every later identical request would hit the
				// incomplete entry from_cache=true. The alias map still
				// records the reported model: the id is authoritative even
				// when the answer set is not.
				if answersComplete(req.Questions, res.Answers) {
					entry := cacheEntry{
						BodySHA:   sha,
						Model:     model,
						Answers:   res.Answers,
						Usage:     Usage{InputTokens: requestTokens(req), Model: model, RequestModel: ModelID},
						CreatedAt: time.Now(),
					}
					_ = c.cacheStore(entry)
				}
				if c.ModelOverride == "" {
					_ = c.updateAliasMap(model)
				}
			}
		} else {
			res = Result{
				Availability: Unattributed,
				Condition:    "endpoint response names no interpreted model id in the shape jev-<major>.<minor>.<patch>; no answer is attributable",
			}
		}
	}

	// Every call that reached the send is charged, whatever the outcome: a
	// sent-but-refused call still consumed input tokens. The record is written
	// once, after the send returns, and it names the interpreted model only for an
	// answered call. The request alias is always recorded in its own field.
	usage := Usage{InputTokens: requestTokens(req), RequestModel: ModelID}
	if res.OK() {
		usage.Model = model
	}
	c.record(usage)
	res.Usage = usage
	return res
}

// changeNotice returns the one-line change notice when model differs from
// VerifiedInterpretedModel and this client has not yet issued it, and marks the
// client as noticed. Otherwise it returns empty.
func (c *Client) changeNotice(model string) string {
	if model == VerifiedInterpretedModel {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.changeNoticed {
		return ""
	}
	c.changeNoticed = true
	return fmt.Sprintf("jev interpreted model changed: answered by %s, verified %s (internal/jev/jev.go:VerifiedInterpretedModel); "+
		"until the measurement regression set is re-run against the new id and the verified id is changed, "+
		"consumers that show or select on a Jev answer return to shadow", model, VerifiedInterpretedModel)
}

// send performs the round trip with the retry policy applied. It returns the
// result, the raw answer map, and the interpreted model id the response reported
// (empty when there was no decodable response).
func (c *Client) send(ctx context.Context, body []byte, repeatable bool) (Result, map[string]wireAnswer, string) {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = EndpointURL
	}
	doer := c.HTTP
	if doer == nil {
		doer = http.DefaultClient
	}

	attempts := 1
	if repeatable && c.MaxRetries > 0 {
		attempts += c.MaxRetries
	}

	var last Result
	for attempt := 0; attempt < attempts; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return Result{Availability: Unreachable, Condition: "request could not be built: " + err.Error()}, nil, ""
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", UserAgent)
		httpReq.Header.Set("Authorization", "Bearer "+c.credentialHeader())

		// The semaphore caps in-flight round trips (REQ-FW-012). The slot is
		// held until the response BODY has been read and closed — a round
		// trip is in flight while its body streams, so releasing at the
		// headers would let decode-time reads exceed MaxConcurrency. Waits
		// and retries still never hold it: both sleep limbs run only after
		// the slot has been returned.
		release, semErr := c.semaphore(ctx)
		if semErr != nil {
			return Result{Availability: Unreachable, Condition: "request canceled while waiting for a concurrency slot: " + semErr.Error()}, nil, ""
		}
		resp, err := doer.Do(httpReq)
		if err != nil {
			release()
			last = Result{Availability: Unreachable, Condition: "transport failed: " + err.Error()}
			// REQ-JEVC-008: retry ONLY a failure that is provably pre-delivery.
			// An ambiguous failure — a read error, a timeout, anything that is
			// not a dial failure — may have been delivered, so repeating it is
			// not provably side-effect-free and it is not retried.
			if preDelivery(err) {
				continue
			}
			return last, nil, ""
		}
		res, answers, model := decode(resp) // reads the body to EOF and closes it
		release()
		// Retry-After limb (REQ-FW-012): a REPEATABLE request facing 429, 529,
		// or another 5xx with a Retry-After header waits the parsed (and
		// bounded) delay and retries, consuming the same attempt budget as
		// the pre-delivery limb — one budget, not two. A non-repeatable
		// request is never retried whatever the header says, composing with
		// REQ-JEVC-008 rather than overriding it.
		if retryableStatus(resp.StatusCode) && repeatable && attempt+1 < attempts {
			if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
				if sleepErr := retryAfterSleep(ctx, d); sleepErr == nil {
					continue
				}
				return res, nil, "" // canceled while waiting
			}
		}
		return res, answers, model
	}
	return last, nil, ""
}

// credentialHeader re-reads the credential for the Authorization header. It is
// read per attempt rather than captured so a rotated credential is honoured.
func (c *Client) credentialHeader() string {
	if c.LoadCredential == nil {
		return ""
	}
	return c.LoadCredential()
}

// preDelivery reports whether err establishes that the request never reached
// the server. Only a dial failure does: the connection was never established,
// so nothing was sent. Everything else — including a timeout, which is the
// tempting case — leaves delivery unknown.
func preDelivery(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return opErr.Op == "dial"
	}
	return false
}

// decode maps the HTTP response to a Result, the raw answer map on success, and
// the interpreted model id the response reported. Ask attributes the answers
// (see Ask) and flattens the map in request order (fromWire).
func decode(resp *http.Response) (Result, map[string]wireAnswer, string) {
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return Result{Availability: Unauthorized, Condition: "endpoint returned HTTP 401"}, nil, ""
	case http.StatusTooManyRequests:
		return Result{Availability: RateLimited, Condition: "endpoint returned HTTP 429"}, nil, ""
	case 529:
		return Result{Availability: Overloaded, Condition: "endpoint returned HTTP 529"}, nil, ""
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{
			Availability: Unreachable,
			Condition:    fmt.Sprintf("endpoint returned HTTP %d", resp.StatusCode),
		}, nil, ""
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{Availability: Unreachable, Condition: "response could not be read: " + err.Error()}, nil, ""
	}
	var wire wireResponse
	if err := json.Unmarshal(raw, &wire); err != nil {
		return Result{Availability: Unreachable, Condition: "response could not be decoded: " + err.Error()}, nil, ""
	}
	return Result{Availability: Available}, wire.Answers, wire.Model
}

// checkBounds applies both size bounds (REQ-JEVC-009) and reports which one was
// crossed. The two are checked separately because they catch different shapes:
// one enormous state, and many individually-small questions whose sum is large.
func checkBounds(req Request) (string, bool) {
	state := EstimateTokens(req.State)

	longest := 0
	total := state
	for _, q := range req.Questions {
		n := QuestionTokens(q)
		total += n
		if n > longest {
			longest = n
		}
	}

	if state+longest > MaxStateAndQuestionTokens {
		return fmt.Sprintf("state plus longest question is ~%d tokens, over the %d bound; refused unsent rather than truncated",
			state+longest, MaxStateAndQuestionTokens), true
	}
	if total > MaxRequestTokens {
		return fmt.Sprintf("whole request is ~%d tokens, over the %d bound; refused unsent rather than truncated",
			total, MaxRequestTokens), true
	}
	return "", false
}

func requestTokens(req Request) int {
	total := EstimateTokens(req.State)
	for _, q := range req.Questions {
		for _, part := range questionParts(q) {
			total += EstimateTokens(part)
		}
	}
	return total
}

// questionParts lists every caller-supplied text a question sends — its text,
// its choice options, and its score levels — so the size bounds, the usage
// estimate, and the secret screen all cover the same payload.
func questionParts(q Question) []string {
	parts := make([]string, 0, 1+len(q.Choices)+len(q.Levels))
	parts = append(parts, q.Text)
	parts = append(parts, q.Choices...)
	return append(parts, q.Levels...)
}

// QuestionTokens estimates the tokens a question's caller-supplied text
// contributes to a request — the same accounting checkBounds applies.
// Exposed so a caller batching questions into multiple requests (fetchweb's
// stage-2 splitter) sizes each batch with the numbers the bound itself will
// use, rather than re-deriving an accounting that can drift.
func QuestionTokens(q Question) int {
	n := 0
	for _, part := range questionParts(q) {
		n += EstimateTokens(part)
	}
	return n
}

// credentialShapedPatterns are the credential shapes the screener refuses. The
// set is deliberately narrow: each entry matches a token whose PREFIX declares
// what it is, so a false positive requires text that already looks like a
// credential. A broad entropy heuristic was rejected — it would refuse ordinary
// payloads (commit hashes, base64 fixtures, UUID runs), and a screener that
// refuses everything is indistinguishable from one that refuses nothing.
var credentialShapedPatterns = []*regexp.Regexp{
	// Vendor-prefixed API keys (OpenAI-style, Anthropic-style, Stripe-style).
	regexp.MustCompile(`\bsk-[A-Za-z0-9_\-]{20,}`),
	// GitHub personal-access and app tokens.
	regexp.MustCompile(`\b(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{20,}`),
	// AWS access key id.
	regexp.MustCompile(`\b(AKIA|ASIA)[A-Z0-9]{16}\b`),
	// A dotenv-style assignment of a secret-named variable to a non-trivial value.
	regexp.MustCompile(`(?i)\b[A-Z0-9_]*(SECRET|PASSWORD|API_?KEY|ACCESS_?TOKEN|PRIVATE_?KEY)[A-Z0-9_]*\s*[=:]\s*["']?[A-Za-z0-9/+_\-]{12,}`),
	// PEM private-key header.
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
}

// minStoredCredentialMatchLength is the floor below which the stored credential
// is NOT searched for verbatim. A very short stored value would otherwise match
// ordinary prose and turn every payload into a refusal — the failure mode that
// makes a screener useless by making it useless to keep enabled.
const minStoredCredentialMatchLength = 8

// ScreenPayload inspects a request's text for credential-shaped tokens before
// it is sent (REQ-JEVC-021). It returns a human-readable reason and whether a
// hit was found.
//
// Two layers, and the second is the one a generic pattern set cannot supply:
// the STORED credential appearing verbatim in the payload is refused even when
// its shape matches no pattern, because a state assembled from a log excerpt or
// a pasted config is exactly how a credential reaches a payload by accident.
func ScreenPayload(req Request, storedCredential string) (string, bool) {
	var b strings.Builder
	b.WriteString(req.State)
	for _, q := range req.Questions {
		for _, part := range questionParts(q) {
			b.WriteByte('\n')
			b.WriteString(part)
		}
	}
	payload := b.String()

	if len(storedCredential) >= minStoredCredentialMatchLength && strings.Contains(payload, storedCredential) {
		return "payload carries the stored TypeSafe credential verbatim; refused unsent", true
	}
	for _, re := range credentialShapedPatterns {
		if loc := re.FindStringIndex(payload); loc != nil {
			return fmt.Sprintf("payload carries a credential-shaped token at byte offset %d; refused unsent", loc[0]), true
		}
	}
	return "", false
}
