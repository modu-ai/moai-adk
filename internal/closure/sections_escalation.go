package closure

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/escalation"
)

// ─── Invariants (REQ-CLOSURE-005) ───

func buildInvariants(r *Report, in BuildInput) {
	sec := InvariantsSection{Rows: []InvariantRow{}}
	if in.Verify.Contract == nil {
		r.Invariants = sec
		return
	}
	for _, inv := range in.Verify.Contract.Invariants {
		row := InvariantRow{Invariant: inv, Kind: invariantKind(inv), Result: InvariantNotObserved}
		if vr := violationRecord(in.Records, inv); vr != nil {
			if vr.Status == escalation.StatusResolved {
				row.Result = InvariantViolationResolved
			} else {
				row.Result = InvariantViolationOpen
			}
		} else if row.Kind == InvariantKindFrozenFiles && in.DetectorArmed {
			// A frozen-files invariant without a violation record was watched
			// by the armed detector: no trip is an observed "no violation".
			row.Result = InvariantNoViolation
		} else {
			r.AddNotPerformed(NotPerformedInvariantNotObserved,
				fmt.Sprintf("%s: no file records that it ran and passed", inv))
		}
		sec.Rows = append(sec.Rows, row)
	}
	r.Invariants = sec
}

// invariantKind classifies a contract invariant string.
func invariantKind(inv string) string {
	switch {
	case strings.HasPrefix(inv, "constitution:"):
		return InvariantKindConstitution
	case inv == "frozen-files":
		return InvariantKindFrozenFiles
	default:
		return InvariantKindCommand
	}
}

// violationRecord returns the invariant-violation record naming inv, open
// first.
func violationRecord(records []escalation.Record, inv string) *escalation.Record {
	var resolved *escalation.Record
	for i := range records {
		rec := &records[i]
		if rec.Class == escalation.ClassInvariantViolation && rec.EscalateOn == inv {
			if rec.Status == escalation.StatusOpen {
				return rec
			}
			resolved = rec
		}
	}
	return resolved
}

// ─── Ownership (REQ-CLOSURE-006) ───

func buildOwnership(r *Report, in BuildInput) {
	sec := OwnershipSection{Records: []EscalationRecordView{}}
	for i := range in.Records {
		rec := &in.Records[i]
		if rec.Class != escalation.ClassOwnershipMove {
			continue
		}
		sec.Records = append(sec.Records, recordView(*rec))
	}
	if in.DetectorArmed {
		sec.Count = CountValue{Observed: true, N: len(sec.Records)}
	} else {
		sec.Count = NotObservedCount()
	}
	r.Ownership = sec
}

func recordView(rec escalation.Record) EscalationRecordView {
	return EscalationRecordView{
		Kind:        rec.Kind,
		Class:       rec.Class,
		Status:      rec.Status,
		Occurrences: rec.Occurrences,
		ContractRef: rec.ContractRef,
		Decider:     rec.Decider,
	}
}

// ─── New APIs (REQ-CLOSURE-007) ───

func buildNewAPIs(r *Report, in BuildInput) {
	sec := NewAPIsSection{Additions: []AdditionView{}}
	// Additions named by the card's class-4 records (A2's own observations).
	for i := range in.Records {
		rec := &in.Records[i]
		if rec.Class != escalation.ClassNewArchitectureOrAPI {
			continue
		}
		if a, ok := additionFromObservation(rec.Observation); ok {
			sec.Additions = append(sec.Additions, a)
		}
	}
	// The read-only comparison at report time. nil seam or an error renders
	// "not observed" (R2: A2's comparison is unexported; a follow-up card
	// exports it — the escalation directory is never written).
	if in.NewAPICompare == nil {
		sec.Observed = false
		r.AddNotPerformed(NotPerformedNewAPIComparisonUnavailable, "the class-4 comparison is not available in this build")
	} else if adds, err := in.NewAPICompare(); err != nil {
		sec.Observed = false
		r.AddNotPerformed(NotPerformedNewAPIComparisonUnavailable, "the class-4 comparison could not run: "+err.Error())
	} else {
		sec.Observed = true
		for _, a := range adds {
			sec.Additions = append(sec.Additions, AdditionView{Kind: a.Kind, Name: a.Name, Path: a.Path})
		}
	}
	r.NewAPIs = sec
}

// additionFromObservation reads the `kind:/name:/path:` body lines A2's
// class-4 writer puts in every class-4 record Observation.
func additionFromObservation(observation string) (AdditionView, bool) {
	var a AdditionView
	for _, line := range strings.Split(observation, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "- kind: "); ok {
			a.Kind = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "- name: "); ok {
			a.Name = strings.TrimSpace(v)
		} else if v, ok := strings.CutPrefix(line, "- path: "); ok {
			a.Path = strings.TrimSpace(v)
		}
	}
	if a.Kind == "" || a.Name == "" {
		return AdditionView{}, false
	}
	return a, true
}

// ─── Escalations (REQ-CLOSURE-008) ───

func buildEscalations(r *Report, in BuildInput) {
	sec := EscalationsSection{
		Records:    []EscalationRecordView{},
		Unreadable: nonNil(in.UnreadableRecords),
	}
	for i := range in.Records {
		sec.Records = append(sec.Records, recordView(in.Records[i]))
	}
	// File order is the append order of the record directory; keep it
	// deterministic for the Markdown form.
	sort.SliceStable(sec.Records, func(i, j int) bool {
		return recordLess(sec.Records[i], sec.Records[j])
	})
	for _, name := range in.UnreadableRecords {
		r.AddNotPerformed(NotPerformedEscalationUnreadable, "record file did not parse: "+name)
	}
	if !in.DetectorArmed {
		sec.NeedsDecision = "not observed"
		r.AddNotPerformed(NotPerformedEscalationDetectorNotArmed, "the escalation detector was not armed for the contract in force")
		r.Escalations = sec
		return
	}
	nd, err := escalation.NeedsDecision(in.Home, in.Card)
	switch {
	case err == nil && nd:
		sec.NeedsDecision = "yes"
	case err == nil:
		sec.NeedsDecision = "no"
	case anyOpenDecision(in.Records):
		sec.NeedsDecision = "yes"
	default:
		sec.NeedsDecision = "not observed"
	}
	r.Escalations = sec
}

func recordLess(a, b EscalationRecordView) bool {
	if a.Class != b.Class {
		return a.Class < b.Class
	}
	return a.Status < b.Status
}

func anyOpenDecision(records []escalation.Record) bool {
	for _, rec := range records {
		if rec.Status == escalation.StatusOpen &&
			(rec.Kind == escalation.KindContract || rec.Kind == escalation.KindOperational) {
			return true
		}
	}
	return false
}
