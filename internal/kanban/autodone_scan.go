// autodone_scan.go — the auto-done landing scan's policy substrate
// (SPEC-TODO-LAND-AUTO-DONE-001 M1).
//
// Three things live here and nowhere else, so the whole close policy is
// unit-testable before any CLI surface exists and every reader of a skip
// reason reads the same closed vocabulary:
//
//  1. The scan-decision pure function (AutoDoneDecide): evidence facts in,
//     one close-or-skip decision out. It encodes the AC-pinned precedence —
//     the sync gate (guard M2) outranks both evidence forms; the recorded
//     delivering SHA (form 1) closes even through an id collision (guard M1),
//     because it is the one form that names THIS card's commit; the collision
//     gate blocks subject evidence alone; and an unanswerable query is a
//     skip, never a close (the three-valued-answer asymmetry holds at the
//     scan layer).
//
//  2. The id-collision counter (AutoDoneDistinctTexts): the reissued-id
//     shape observed on t654/t656/t657, counted across the live queue and the
//     archive.
//
//  3. The one-query subject stream (ScanLandedSubjects + LandedAttributions):
//     the scan evaluates every live card against ONE `git log` — a SHA-keyed
//     subject stream, the same subject-only discipline the landed predicate
//     is built on, so a body mention never becomes landing evidence here
//     either.
//
// SUBAGENT BOUNDARY: nothing here prompts; every function is a pure
// computation over its inputs.
package kanban

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// The skip-reason vocabulary is CLOSED at exactly four tokens (REQ-AD-010).
// They are constants, and AutoDoneSkipReasons is the enumeration every reader
// and the `--help` body render from — a new reason is a one-place change that
// makes the closed-set test go red.
const (
	// AutoDoneSkipAmbiguousID — guard M1: the id has carried more than one
	// distinct card text, and the evidence on offer is subject attribution,
	// which cannot tell the predecessor's landed work from this card's.
	AutoDoneSkipAmbiguousID = "ambiguous-id"
	// AutoDoneSkipSpecNotCompleted — guard M2: the card's SPEC frontmatter
	// status is anything other than `completed` (an unreadable status
	// included — unknown is not a pass), so a run commit landing does not
	// license the close while sync is unfinished.
	AutoDoneSkipSpecNotCompleted = "spec-not-completed"
	// AutoDoneSkipNotLanded — the query ran and found no landing evidence,
	// including a subject carrying an explicit non-landing declaration.
	AutoDoneSkipNotLanded = "not-landed"
	// AutoDoneSkipQueryInconclusive — the question could not be asked (no
	// git, an unresolvable ref, a malformed stream). LandingUnknown is not
	// evidence of landed, so the card skips.
	AutoDoneSkipQueryInconclusive = "query-inconclusive"
)

// AutoDoneSkipReasons returns the closed skip-reason vocabulary in its
// canonical order. The scan's `--help` body and the closed-set test render
// from this one enumeration.
//
// @MX:NOTE: [AUTO] the vocabulary is CLOSED at four tokens (SPEC-TODO-LAND-AUTO-DONE-001
// REQ-AD-010); a new skip reason is a one-place change here plus the
// closed-set test — never a new literal at a call site.
func AutoDoneSkipReasons() []string {
	return []string{
		AutoDoneSkipAmbiguousID,
		AutoDoneSkipSpecNotCompleted,
		AutoDoneSkipNotLanded,
		AutoDoneSkipQueryInconclusive,
	}
}

// The two evidence forms a close may carry (REQ-AD-004). A close on anything
// else is a policy violation the decision function cannot express.
const (
	// AutoDoneFormSHA — the card's recorded delivering SHA (operator-asserted
	// via `todo landed --sha`) resolved and reachable from the landed ref.
	AutoDoneFormSHA = "sha-recorded"
	// AutoDoneFormSubject — the landed ref's subject stream attributes the
	// card through one of the positional shapes.
	AutoDoneFormSubject = "subject-attribution"
)

// AutoDoneTri is the three-valued answer the decision function's gates run
// on. It exists so an UNANSWERABLE question cannot masquerade as a negative:
// AutoDoneUnknown on the SHA-reachability axis skips as query-inconclusive,
// and AutoDoneUnknown on the sync gate skips as spec-not-completed — never
// the opposite.
type AutoDoneTri int

const (
	// AutoDoneNo — the question ran and answered no.
	AutoDoneNo AutoDoneTri = iota
	// AutoDoneYes — the question ran and answered yes. For the sync gate,
	// also the value when no gate applies (a card with no spec id).
	AutoDoneYes
	// AutoDoneUnknown — the question could not be asked.
	AutoDoneUnknown
)

// AutoDoneFacts is everything the scan knows about one live card. The CLI
// layer gathers these from git and the queue record; the decision is a pure
// function of them.
type AutoDoneFacts struct {
	// RecordedSHA is the card's stored delivering SHA, or "" when the card
	// carries none.
	RecordedSHA string
	// SHAReachable is the reachability of RecordedSHA from the landed ref.
	SHAReachable AutoDoneTri
	// SubjectHit is the commit whose subject attributes the card, or nil when
	// none does.
	SubjectHit *LandedCommit
	// SubjectKnown reports whether the subject query itself ran and answered;
	// a false value with a nil SubjectHit is INCONCLUSIVE, not not-landed.
	SubjectKnown bool
	// DistinctTexts is how many distinct texts the id has carried across the
	// live queue and the archive. More than one is the reissued-id collision
	// guard M1 fires on.
	DistinctTexts int
	// SpecSyncGate answers "may this card close on the sync axis": AutoDoneYes
	// when the card carries no spec id (the gate does not apply) or its SPEC
	// reads `completed`; AutoDoneNo when the SPEC reads anything else or could
	// not be read; AutoDoneUnknown when the read itself was impossible.
	SpecSyncGate AutoDoneTri
}

// AutoDoneDecision is one card's scan outcome: a close carrying its evidence
// form, or a skip carrying its closed-vocabulary reason.
type AutoDoneDecision struct {
	Close  bool
	Form   string // AutoDoneFormSHA | AutoDoneFormSubject; set only when Close
	Reason string // the skip reason; set only when !Close
}

// AutoDoneDecide decides one card. The precedence is the SPEC's, in order:
//
//  1. Guard M2 — the sync gate. A card whose SPEC has not completed skips
//     before any evidence is weighed (AC-AD-006).
//  2. Evidence form 1 — the recorded delivering SHA. It closes when
//     reachable — even through an id collision, because the SHA names THIS
//     card's commit (AC-AD-005). An unanswerable reachability check is
//     inconclusive; a checked-and-unreachable SHA is not evidence, and
//     evaluation falls through to form 2.
//  3. Guard M1 — the collision gate, applied to subject evidence alone: an
//     id carried by more than one distinct text skips ambiguous-id (AC-AD-004).
//  4. Evidence form 2 — an attributed subject closes; a query that could not
//     run skips query-inconclusive (AC-AD-008); an answered query that found
//     nothing skips not-landed.
func AutoDoneDecide(f AutoDoneFacts) AutoDoneDecision {
	// Guard M2 — the sync gate outranks evidence. An unknown read is not a
	// pass: the gate skips the card rather than closing on an unreadable SPEC.
	if f.SpecSyncGate != AutoDoneYes {
		return AutoDoneDecision{Reason: AutoDoneSkipSpecNotCompleted}
	}

	// Evidence form 1 — the recorded delivering SHA.
	if strings.TrimSpace(f.RecordedSHA) != "" {
		switch f.SHAReachable {
		case AutoDoneYes:
			return AutoDoneDecision{Close: true, Form: AutoDoneFormSHA}
		case AutoDoneUnknown:
			return AutoDoneDecision{Reason: AutoDoneSkipQueryInconclusive}
		case AutoDoneNo:
			// Checked and not reachable: the record names a commit that is
			// not on the landed ref. It is not evidence; keep evaluating.
		}
	}

	// Evidence form 2 — subject attribution.
	if !f.SubjectKnown {
		return AutoDoneDecision{Reason: AutoDoneSkipQueryInconclusive}
	}
	if f.SubjectHit != nil {
		if f.DistinctTexts > 1 {
			return AutoDoneDecision{Reason: AutoDoneSkipAmbiguousID}
		}
		return AutoDoneDecision{Close: true, Form: AutoDoneFormSubject}
	}
	return AutoDoneDecision{Reason: AutoDoneSkipNotLanded}
}

// AutoDoneDistinctTexts counts the distinct card texts the id has carried
// across the live queue and the archive — the input guard M1 fires on. The
// same text recorded in both places is ONE text, not a collision: a reissue
// is a DIFFERENT card under a reused name.
func AutoDoneDistinctTexts(rec *BacklogRecord, id string) int {
	seen := map[string]bool{}
	for i := range rec.Items {
		if rec.Items[i].ID == id {
			seen[rec.Items[i].Text] = true
		}
	}
	for i := range rec.Archived {
		if rec.Archived[i].Item.ID == id {
			seen[rec.Archived[i].Item.Text] = true
		}
	}
	return len(seen)
}

// LandedCommit is one commit of the landed ref's subject stream: its full
// SHA, its subject line, and its committer time (unix seconds, from %ct).
// The scan's log rows carry all three when a close rides on the subject's
// attribution; the committer time is what the generation boundary
// (AutoDoneSubjectFresh) judges.
type LandedCommit struct {
	SHA        string
	Subject    string
	CommitTime int64
}

// LandedScanArgs builds the argv for the scan's ONE subject-stream query
// against ref. Like LandedSubjectArgs, it is a SUBJECT stream (%s), never a
// whole-message stream (%B); the %H prefix keys each subject to its commit so
// a form-2 close can name the commit that attributed the card, the %ct field
// carries the committer time the generation boundary reads, and the %x00
// separator keeps the three fields separable for any subject content.
func LandedScanArgs(ref string) []string {
	if strings.TrimSpace(ref) == "" {
		ref = DefaultLandedRef
	}
	return []string{"log", ref, "--format=%H%x00%ct%x00%s"}
}

// ScanLandedSubjects runs the scan's one query and returns the landed ref's
// subject stream, newest first (git log order).
//
// An unanswerable query is an ERROR, never an empty result: an empty stream
// reads as "nothing landed", which collapses the inconclusive answer into
// not-landed — the exact collapse the three-valued answer exists to prevent.
// A line that does not carry the SHA separator is reported too, for the same
// reason a silently-empty stream would be.
func ScanLandedSubjects(run CommandRunner, ref string) ([]LandedCommit, error) {
	if run == nil {
		return nil, fmt.Errorf("kanban: landed scan has no command runner")
	}
	out, err := run("git", LandedScanArgs(ref)...)
	if err != nil {
		return nil, fmt.Errorf("kanban: git log %s: %w", ref, err)
	}
	var commits []LandedCommit
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		sha, rest, ok := strings.Cut(line, "\x00")
		if !ok {
			return nil, fmt.Errorf("kanban: git log %s: malformed scan line %q — no SHA separator", ref, line)
		}
		timeStr, subject, ok := strings.Cut(rest, "\x00")
		if !ok {
			return nil, fmt.Errorf("kanban: git log %s: malformed scan line %q — no time separator", ref, line)
		}
		ct, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("kanban: git log %s: malformed scan line %q — committer time %q is not unix seconds", ref, line, timeStr)
		}
		commits = append(commits, LandedCommit{SHA: sha, Subject: subject, CommitTime: ct})
	}
	return commits, nil
}

// LandedAttributions maps each card id the subject stream attributes to the
// FIRST commit that attributes it — newest first, so a card named by several
// commits closes on the one that delivered its latest work. The attribution
// itself is subjectAttribution, the same predicate the landed check answers
// through; the scan introduces no second matcher.
func LandedAttributions(commits []LandedCommit, landedBranch string) map[string]LandedCommit {
	out := map[string]LandedCommit{}
	for _, c := range commits {
		id := subjectAttribution(c.Subject, landedBranch)
		if id == "" {
			continue
		}
		if _, seen := out[id]; !seen {
			out[id] = c
		}
	}
	return out
}

// AutoDoneSubjectFresh answers whether an attributed subject hit may close
// THIS card: the attributing commit's committer time must not precede the
// card's creation time (added_at). This is the t684 generation boundary —
// the subject attribution reads the landed ref's WHOLE history, so a
// reissued id would otherwise inherit its old generation's landing; the
// store cannot tell the generations apart (a cross-store reissue counts one
// distinct text, so guard M1 stays silent), but the clock can: the old
// generation's commits are older than the reissued card.
//
// A card whose added_at cannot be parsed (an empty or corrupt field) fails
// CLOSED — no subject close without a decidable boundary. Same-second
// granularity counts as fresh: the boundary separates generations, it is
// not second-level forensics.
func AutoDoneSubjectFresh(hit LandedCommit, addedAt string) bool {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(addedAt))
	if err != nil {
		return false
	}
	return hit.CommitTime >= t.Unix()
}
