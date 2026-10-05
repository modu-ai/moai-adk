// Package jevmeasure is the measurement harness for Jev consumers, and the
// request builder they construct their questions through
// (SPEC-JEV-OPTIN-MEASURE-001 §C.3 and §C.4).
//
// It implements an APPARATUS, not a result. No consumer lives here: this
// package decides whether a consumer is ALLOWED to exist, by measuring it
// against its own constant-answer baseline in two language arms under a pinned
// model id. The consumers themselves are a later SPEC's work.
//
// Two things make that gate meaningful rather than ceremonial:
//
//   - The baseline, not raw accuracy, is the number to beat. Raw accuracy is
//     unreadable without the base rate, and this repository has already been
//     misled by one: a headline that moved sixteen points and read as success
//     while the constant baseline sat above it the whole time, unbeaten.
//   - An absence carries a positive control. A zero-hit and a broken measuring
//     apparatus are indistinguishable on their own, and the zero direction is
//     the one that does not prompt a re-measurement.
//
// [HARD] No code in this package contacts the vendor endpoint. Callers supply
// an Answerer; the live one is *jev.Client, and every test supplies a stub.
package jevmeasure

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// NoMatchChoice is the option every Choice question must carry.
//
// A forced choice over an option set that excludes the true answer produces a
// confident wrong answer. That failure mode is not model-specific — this
// repository has recorded it from its own dispatch practice, where a question
// offering two branches got a confident answer and the true answer was a third
// thing.
const NoMatchChoice = "none of these"

// State is the named-field payload a request carries. It is deliberately not a
// bare map: the field ORDER is preserved so the serialized state is stable
// across runs, which is what makes two measurements comparable.
//
// Every value here is computed in Go. The model is never asked to count, to do
// arithmetic, to compare numeric proximity, to order dates, or to compare
// SHAs — those reach it as finished values under names (REQ-JEVO-016).
type State struct {
	order  []string
	fields map[string]any
}

// NewState returns an empty state.
func NewState() *State {
	return &State{fields: map[string]any{}}
}

// Set records a named field. Re-setting a name overwrites the value and keeps
// the original position, so a builder that conditionally overrides a field does
// not reorder the payload.
func (s *State) Set(name string, value any) *State {
	if _, exists := s.fields[name]; !exists {
		s.order = append(s.order, name)
	}
	s.fields[name] = value
	return s
}

// Names returns the field names in insertion order.
func (s *State) Names() []string {
	out := make([]string, len(s.order))
	copy(out, s.order)
	return out
}

// JSON serializes the state as a JSON object. Key order follows insertion.
func (s *State) JSON() (string, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, name := range s.order {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := json.Marshal(name)
		if err != nil {
			return "", fmt.Errorf("jevmeasure: marshal state key %q: %w", name, err)
		}
		v, err := json.Marshal(s.fields[name])
		if err != nil {
			return "", fmt.Errorf("jevmeasure: marshal state field %q: %w", name, err)
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.String(), nil
}

// Question is one typed question plus the state fields it reads.
//
// Reads is not documentation. It is what lets BuildRequest verify that the
// state carries only fields some question actually consults — accuracy falls
// as irrelevant state grows, so an unread field is measurable harm rather than
// mere waste (REQ-JEVO-018).
type Question struct {
	ID      string
	Text    string
	Kind    jev.AnswerKind
	Choices []string
	Reads   []string
}

// computationMarkers are the phrasings that hand a computation to the model.
// The list is deliberately over-broad on the verbs that matter and silent on
// everything else: a false rejection costs the author a rewording, while a
// false acceptance ships a question the model is known to be unreliable at.
var computationMarkers = []string{
	"how many", "count ", "count the", "number of",
	"sum of", "total of", "add up", "subtract", "multiply", "divide",
	"closer to", "nearer to", "closest to", "nearest to",
	"greater than", "less than", "larger than", "smaller than",
	"earlier", "later than", "which date", "sort ", "order these", "order the",
	"compare the sha", "which sha", "same sha", "sha match",
}

// BuildRequest assembles a jev.Request from a state and its questions,
// enforcing the question-design constraints mechanically.
//
// The enforcement is the point. Each of these was written as a rule an author
// could follow; a rule only a reviewer enforces is a rule that holds until the
// first hurried change.
func BuildRequest(state *State, questions []Question) (jev.Request, error) {
	if state == nil {
		return jev.Request{}, fmt.Errorf("jevmeasure: nil state")
	}
	if len(questions) == 0 {
		return jev.Request{}, fmt.Errorf("jevmeasure: no questions")
	}

	read := map[string]bool{}
	out := make([]jev.Question, 0, len(questions))

	for _, q := range questions {
		if strings.TrimSpace(q.ID) == "" {
			return jev.Request{}, fmt.Errorf("jevmeasure: question with empty id")
		}
		if strings.TrimSpace(q.Text) == "" {
			return jev.Request{}, fmt.Errorf("jevmeasure: question %q has no text", q.ID)
		}
		// REQ-JEVO-016: no computation is handed to the model.
		lower := strings.ToLower(q.Text)
		for _, marker := range computationMarkers {
			if strings.Contains(lower, marker) {
				return jev.Request{}, fmt.Errorf(
					"jevmeasure: question %q asks the model to compute (%q) — compute it in Go and pass the result as a named state field",
					q.ID, strings.TrimSpace(marker))
			}
		}
		// REQ-JEVO-017: every Choice carries an explicit no-match option.
		if q.Kind == jev.KindChoice {
			if !hasNoMatch(q.Choices) {
				return jev.Request{}, fmt.Errorf(
					"jevmeasure: Choice question %q carries no no-match option (add jevmeasure.NoMatchChoice)", q.ID)
			}
		}
		// A question reading nothing is a question answered from context the
		// request does not carry.
		if len(q.Reads) == 0 {
			return jev.Request{}, fmt.Errorf("jevmeasure: question %q declares no state fields to read", q.ID)
		}
		for _, name := range q.Reads {
			if _, ok := state.fields[name]; !ok {
				return jev.Request{}, fmt.Errorf(
					"jevmeasure: question %q reads state field %q, which the state does not carry", q.ID, name)
			}
			read[name] = true
		}
		out = append(out, jev.Question{
			ID: q.ID, Text: q.Text, Kind: q.Kind, Choices: q.Choices,
		})
	}

	// REQ-JEVO-018: no state field goes unread.
	var unread []string
	for _, name := range state.order {
		if !read[name] {
			unread = append(unread, name)
		}
	}
	if len(unread) > 0 {
		sort.Strings(unread)
		return jev.Request{}, fmt.Errorf(
			"jevmeasure: state carries fields no question reads: %s — accuracy falls as irrelevant state grows",
			strings.Join(unread, ", "))
	}

	// REQ-JEVO-019: the state travels as a JSON object under the state key.
	// Nothing from it is spliced into any question's text, so its content is
	// data the model reads ABOUT, never instruction it reads AS.
	payload, err := state.JSON()
	if err != nil {
		return jev.Request{}, err
	}
	return jev.Request{State: payload, Questions: out}, nil
}

func hasNoMatch(choices []string) bool {
	for _, c := range choices {
		if strings.EqualFold(strings.TrimSpace(c), NoMatchChoice) {
			return true
		}
	}
	return false
}

// NoulPair carries both polarities of a yes/no judgment.
//
// REQ-JEVO-020: P(yes) is not guaranteed to equal 1 - P(no). Where both
// matter, both are asked and both are read; neither is derived from the other.
// This type exists so that "read both" is the shape of the API rather than a
// discipline the caller has to remember.
type NoulPair struct {
	Positive jev.Answer
	Negative jev.Answer
}

// ReadNoulPair returns both polarities, or an error naming the missing one.
// It never derives an absent probability from its counterpart.
func ReadNoulPair(answers []jev.Answer, positiveID, negativeID string) (NoulPair, error) {
	pos, okPos := findAnswer(answers, positiveID)
	neg, okNeg := findAnswer(answers, negativeID)
	switch {
	case !okPos && !okNeg:
		return NoulPair{}, fmt.Errorf("jevmeasure: neither %q nor %q was answered", positiveID, negativeID)
	case !okPos:
		return NoulPair{}, fmt.Errorf("jevmeasure: %q was not answered; it is NOT derivable from %q", positiveID, negativeID)
	case !okNeg:
		return NoulPair{}, fmt.Errorf("jevmeasure: %q was not answered; it is NOT derivable from %q", negativeID, positiveID)
	}
	return NoulPair{Positive: pos, Negative: neg}, nil
}

func findAnswer(answers []jev.Answer, id string) (jev.Answer, bool) {
	for _, a := range answers {
		if a.QuestionID == id {
			return a, true
		}
	}
	return jev.Answer{}, false
}
