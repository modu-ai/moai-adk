package factorylane

import "fmt"

// Classified-pickup consumption (SPEC-FACTORY-LANE-AUTONOMY-001 fragment 2,
// design.md D2). This package is the CONSUMER ONLY: the classification field
// schema — names, types, storage format, defaults — remains t1332's producer
// concern (spec.md §F Out of Scope). What lives here is the DECLARED MINIMAL
// CONSUMPTION INTERFACE: an execution axis and, for sequential work, a group
// name, read per card through the Classifier seam. Absent or unrecognized
// metadata is a tolerated value that routes the pickup to the fallback
// classification (REQ-FLA-007) — never an error (REQ-FLA-008).

// Axis is the execution axis the consumer understands. The values are the
// consumer's own interface vocabulary from REQ-FLA-006; a producer token the
// consumer does not recognize normalizes to AxisFallback (Known=false), not
// to an error.
type Axis string

const (
	// AxisSequential marks single-holder work: one lane at a time per group.
	AxisSequential Axis = "sequential"
	// AxisParallel marks concurrently pickable work: many lanes may hold it.
	AxisParallel Axis = "parallel"
	// AxisFallback is the consumer's normalization of absent or unrecognized
	// metadata — the REQ-FLA-007 single-dispatch default.
	AxisFallback Axis = "fallback"
)

// Classification is what the pickup consumer reads per card. Known=false
// marks metadata that is absent or carries an unrecognized axis value: the
// decision then follows the fallback classification and the tolerated-unknown
// condition is logged (REQ-FLA-008). Group scopes sequential exclusivity —
// one lane at a time per group — and is meaningless for parallel work.
// Priority joins this interface only when a rule consumes it; nothing in the
// M2 rules does.
type Classification struct {
	Axis  Axis
	Group string
	Known bool
}

// Classifier is the consumption seam t1332's future producer wiring will
// implement. Its contract is tolerant by construction: a card with no
// metadata — or metadata the reader cannot map — reads back as
// Classification{Axis: AxisFallback, Known: false} with a nil error. A real
// read failure (I/O, corruption) is still an error; absence is not.
//
// @MX:NOTE: [AUTO] declared minimal consumption interface (design.md D2) — the producer schema stays t1332's; when t1332 lands, its wiring implements this seam and nothing here changes shape.
// @MX:SPEC: SPEC-FACTORY-LANE-AUTONOMY-001
type Classifier interface {
	Classify(cardID string) (Classification, error)
}

// StaticClassifier is a fixed-card classification lookup: the M2 wiring point
// (no producer exists yet) and the test seam. A card absent from the map —
// every card, until t1332 lands — reads back unknown and tolerated.
type StaticClassifier map[string]Classification

// Classify implements Classifier over the fixed map.
func (m StaticClassifier) Classify(cardID string) (Classification, error) {
	if cls, ok := m[cardID]; ok {
		return cls, nil
	}
	return Classification{Axis: AxisFallback}, nil
}

// NormalizeClassification maps one raw producer axis token onto the
// consumer's Classification — the single normalization point of the declared
// minimal interface (REQ-FLA-008). The two recognized tokens read back
// known; an empty axis field or any unrecognized token reads back
// Known=false with AxisFallback. It never returns an error: tolerance is
// the contract, and t1332's future reader wiring calls this to build what
// its Classifier serves.
func NormalizeClassification(rawAxis, group string) Classification {
	switch Axis(rawAxis) {
	case AxisSequential:
		return Classification{Axis: AxisSequential, Group: group, Known: true}
	case AxisParallel:
		return Classification{Axis: AxisParallel, Known: true}
	default:
		return Classification{Axis: AxisFallback}
	}
}

// Hold is one lane's active hold on a card, read from the F1 lease model
// (card records in a lease-holding state; the holder is the lane label).
// Sequential exclusivity is decided against this list — no new
// serialization mechanism is introduced (REQ-FLA-006; design.md D3 keeps
// the integration window out of pickup entirely).
type Hold struct {
	Lane string
	Card string
}

// PickupDecision is the classified-pickup verdict for one candidate card.
// A decision is a report, not an error: even a denied sequential pickup
// returns a nil error with Allowed=false and WaitOn naming the holder —
// the CLI exit stays 0 (the probe's unavailable-verdict precedent).
type PickupDecision struct {
	Card             string         `json:"card"`
	Lane             string         `json:"lane"`
	Classification   Classification `json:"classification"`
	Allowed          bool           `json:"allowed"`
	MultiPick        bool           `json:"multi_pick"`
	Fallback         bool           `json:"fallback"`
	ToleratedUnknown bool           `json:"tolerated_unknown"`
	WaitOn           string         `json:"wait_on,omitempty"`
	Reason           string         `json:"reason"`
}

// PlanPickup answers one classified-pickup question: may lane pick card now?
// The rules are REQ-FLA-006/007/008: a sequential-classified card is pickable
// only while no other lane holds a card of the same sequential group; a
// parallel-classified card is pickable by multiple lanes concurrently; a card
// with absent or unrecognized metadata follows the fallback classification —
// the operator-picked single-dispatch behavior, never autonomously
// multi-picked.
func PlanPickup(lane, card string, holds []Hold, reader Classifier) (PickupDecision, error) {
	if reader == nil {
		reader = StaticClassifier{}
	}
	cls, err := reader.Classify(card)
	if err != nil {
		return PickupDecision{}, err
	}
	decision := PickupDecision{Card: card, Lane: lane, Classification: cls}
	// REQ-FLA-007/008: absent metadata or an axis outside the enumeration
	// routes the pickup to the fallback classification — the operator-picked
	// single-dispatch behavior, never an autonomous multi-pick. This branch
	// is the M2 default too (plan.md §F): until t1332's producer exists,
	// every card evaluates through it.
	if !cls.Known || (cls.Axis != AxisSequential && cls.Axis != AxisParallel) {
		decision.Allowed = true
		decision.Fallback = true
		decision.ToleratedUnknown = true
		decision.Reason = "no usable classification metadata — tolerated; single-dispatch fallback behavior, no autonomous multi-pick (REQ-FLA-007)"
		return decision, nil
	}
	if cls.Axis == AxisParallel {
		decision.Allowed = true
		decision.MultiPick = true
		decision.Reason = "parallel-classified card — multiple lanes may hold it concurrently (REQ-FLA-006)"
		return decision, nil
	}
	// AxisSequential: pickable only while no other lane holds a card of the
	// same sequential group. The holds are the F1 lease model's records —
	// the existing serialization, never a new lock (design.md D3).
	for _, h := range holds {
		if h.Lane == lane {
			continue
		}
		held, err := reader.Classify(h.Card)
		if err != nil {
			return PickupDecision{}, err
		}
		if held.Known && held.Axis == AxisSequential && held.Group == cls.Group {
			decision.WaitOn = h.Lane
			decision.Reason = fmt.Sprintf("sequential group %q held by %s (card %s) — pickable only while no other lane holds the group (REQ-FLA-006)", cls.Group, h.Lane, h.Card)
			return decision, nil
		}
	}
	decision.Allowed = true
	decision.Reason = fmt.Sprintf("sequential group %q free — exclusive pickup granted (REQ-FLA-006)", cls.Group)
	return decision, nil
}
