package closure

import (
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/contract"
)

// Readiness codes (REQ-CLOSURE-016). The set is closed: exactly these nine,
// no more. ReadinessCodes returns the sorted list and is the evaluator's
// declared code list (AC-CLOSURE-016).
const (
	CodeContractInvalid           = "contract_invalid"
	CodeClosureReportMissing      = "closure_report_missing"
	CodeClosureReportStale        = "closure_report_stale"
	CodeHumanVerdictReject        = "human_verdict_reject"
	CodeHumanVerdictAmendContract = "human_verdict_amend_contract"
	CodeSecondReviewNotPerformed  = "second_review_not_performed"
	CodeSecondReviewStale         = "second_review_stale"
	CodeSecondReviewFailed        = "second_review_failed"
	CodePushCheckUndetermined     = "push_check_undetermined"
)

// ReadinessCodes returns the closed nine-code set, sorted.
func ReadinessCodes() []string {
	return []string{
		CodeClosureReportMissing,
		CodeClosureReportStale,
		CodeContractInvalid,
		CodeHumanVerdictAmendContract,
		CodeHumanVerdictReject,
		CodePushCheckUndetermined,
		CodeSecondReviewFailed,
		CodeSecondReviewNotPerformed,
		CodeSecondReviewStale,
	}
}

// Review targets (design.md §A.1). Only a baseBranch review covers a scope.
const TargetBaseBranch = "baseBranch"

// Second-review policies (A1 configuration).
const (
	PolicyRequired = "required"
	PolicyAdvisory = "advisory"
	PolicyOff      = "off"
)

// CommitPaths is one non-merge commit and the repo-relative paths it
// changed. It is gitio.CommitPaths: the pure evaluator and the subprocess
// layer share one type so no adapter can drop a field.
type CommitPaths = gitio.CommitPaths

// GitFacts is what the in-history filter and the currency rule need from git,
// injected so the evaluator stays pure. The production implementation lives in
// internal/closure/gitio (M3); tests inject maps or a real repository.
type GitFacts struct {
	// IsAncestor reports whether a is b or an ancestor of b.
	IsAncestor func(a, b string) (bool, error)
	// NonMergeCommits lists the non-merge commits in from..to (from
	// exclusive, to inclusive), oldest first, with each commit's paths.
	NonMergeCommits func(from, to string) ([]CommitPaths, error)
}

// SecondReviewInput is everything the §D selection needs.
type SecondReviewInput struct {
	// Records are the decodable lines of second-review.jsonl in file order.
	Records []SecondReviewRecord
	// ContractCard is the contract's signed card ("" when absent/unsigned).
	ContractCard string
	// ContractDigest is the signed contract digest ("" when absent/unsigned).
	ContractDigest string
	// ContractSecondModel is contract review.second_model (display only).
	ContractSecondModel string
	// EvalCommit is the evaluation commit P: the report HEAD, or the pushed
	// source commit for push readiness.
	EvalCommit string
	// Facts is the injected git access.
	Facts GitFacts
	// WriteGlobs and SpecID define the governed paths of the currency rule
	// (spec.md §C.7): matched by WriteGlobs, excluding .moai/specs/SpecID/.
	WriteGlobs []string
	SpecID     string
}

// SecondReviewState is the §D selection result (REQ-CLOSURE-013).
type SecondReviewState struct {
	State               string                    // performed | not-performed | stale | not-required
	Cause               string                    // not-performed cause
	Verdict             string                    // pass | fail (performed)
	Backends            []SecondReviewBackendView // the counted backends
	SubstituteBackend   bool                      // performing backend ≠ contract second_model
	ContractSecondModel string
	SupersedingCommit   string // stale: the first governed-path commit after the review
	// CurrencyUndetermined states that the currency rule could not run (a git
	// failure), NOT that the review is current or stale. The performed verdict
	// is still what was recorded; the push evaluator maps this flag to
	// push_check_undetermined (REQ-CLOSURE-017), never to a silent pass.
	CurrencyUndetermined bool
}

// SelectSecondReview applies the five filters of design.md §D in order and
// returns the resulting state. When a filter empties the set the cause is
// that filter's cause; with no records at all it is no-record.
func SelectSecondReview(in SecondReviewInput) SecondReviewState {
	st := SecondReviewState{
		ContractSecondModel: orNotRecorded(in.ContractSecondModel),
		Backends:            []SecondReviewBackendView{},
	}
	if len(in.Records) == 0 {
		st.State, st.Cause = SecondReviewNotPerformed, SecondReviewCauseNoRecord
		return st
	}

	records := in.Records
	apply := func(keep func(SecondReviewRecord) bool, cause string) {
		if st.State != "" {
			return
		}
		kept := make([]SecondReviewRecord, 0, len(records))
		for _, r := range records {
			if keep(r) {
				kept = append(kept, r)
			}
		}
		if len(kept) == 0 {
			st.State, st.Cause = SecondReviewNotPerformed, cause
			return
		}
		records = kept
	}

	// 1. bound: card argument equals the contract card; SPEC ID and digest
	// non-empty.
	apply(func(r SecondReviewRecord) bool {
		return in.ContractCard != "" && r.Card == in.ContractCard &&
			r.SpecID != "" && r.ContractSHA256 != ""
	}, SecondReviewCauseUnbound)

	// 2. same contract: digest equals the current signed digest.
	apply(func(r SecondReviewRecord) bool {
		return in.ContractDigest != "" && r.ContractSHA256 == in.ContractDigest
	}, SecondReviewCauseContractChanged)

	// 3. scope covered: baseBranch target, scope head equals the audited
	// commit, at least one file changed.
	apply(func(r SecondReviewRecord) bool {
		return r.Target == TargetBaseBranch &&
			r.Scope.HeadSHA == r.HeadSHA &&
			r.Scope.ChangedFiles >= 1
	}, SecondReviewCauseScopeNotCovered)

	// 4. second model: at least one codex or GLM verdict of pass or fail.
	apply(func(r SecondReviewRecord) bool {
		return len(countedBackends(r)) > 0
	}, SecondReviewCauseNoSecondModel)

	// 5. in history: audited commit is the evaluation commit or its ancestor.
	// An ancestry error excludes the record (fail-closed); the push evaluator
	// still stops through other codes when git is broken.
	apply(func(r SecondReviewRecord) bool {
		anc, err := in.Facts.IsAncestor(r.HeadSHA, in.EvalCommit)
		return err == nil && anc
	}, SecondReviewCauseNotInHistory)

	if st.State != "" {
		return st
	}

	// Latest by recorded_at (unparsable sorts oldest). File order breaks ties,
	// so the last equally-old append wins.
	r := records[0]
	best := parseRFC3339(r.RecordedAt)
	for _, cand := range records[1:] {
		if !candAtAfter(cand.RecordedAt, best) {
			continue
		}
		r = cand
		best = parseRFC3339(cand.RecordedAt)
	}

	backends := countedBackends(r)
	for _, b := range backends {
		st.Backends = append(st.Backends, SecondReviewBackendView{Backend: b.Backend, Verdict: b.Verdict})
	}
	st.Verdict = "pass"
	for _, b := range backends {
		if b.Verdict == "fail" {
			st.Verdict = "fail"
			break
		}
	}
	st.State = SecondReviewPerformed
	st.SubstituteBackend = in.ContractSecondModel != "" && !backendMatches(in.ContractSecondModel, backends)

	// Currency rule (spec.md §C.7): R = r.HeadSHA, P = EvalCommit.
	current, superseding, ok := currency(in.Facts, in.WriteGlobs, in.SpecID, r.HeadSHA, in.EvalCommit)
	if !ok {
		st.CurrencyUndetermined = true
		return st
	}
	if !current {
		st.State = SecondReviewStale
		st.SupersedingCommit = superseding
	}
	return st
}

// countedBackends returns the record's codex/GLM backends whose verdict is
// pass or fail — the ones REQ-CLOSURE-013 counts.
func countedBackends(r SecondReviewRecord) []SecondReviewBackend {
	var out []SecondReviewBackend
	for _, b := range r.Backends {
		if (b.Backend == "codex" || b.Backend == "glm") && (b.Verdict == "pass" || b.Verdict == "fail") {
			out = append(out, b)
		}
	}
	return out
}

// backendMatches reports whether the contract's named second model performed:
// a codex contract is satisfied by codex; anything else the contract names
// that no counted backend carried makes it a substitute.
func backendMatches(secondModel string, counted []SecondReviewBackend) bool {
	for _, b := range counted {
		if b.Backend == secondModel {
			return true
		}
	}
	return false
}

// currency applies the currency rule: R is current for P when R is P or an
// ancestor of P and no non-merge commit in R..P changes a governed path. The
// third return is false when git could not answer (undetermined, never
// silently current).
func currency(facts GitFacts, writeGlobs []string, specID, r, p string) (current bool, superseding string, ok bool) {
	if r == "" || p == "" {
		return false, "", true
	}
	anc, err := facts.IsAncestor(r, p)
	if err != nil {
		return false, "", false
	}
	if !anc {
		return false, "", true
	}
	commits, err := facts.NonMergeCommits(r, p)
	if err != nil {
		return false, "", false
	}
	for _, c := range commits {
		for _, path := range c.Paths {
			if governedPath(writeGlobs, specID, path) {
				return false, c.SHA, true
			}
		}
	}
	return true, "", true
}

// governedPath reports whether p is governed by the contract: matched by a
// write glob and not under the SPEC's own directory (spec.md §C.7).
func governedPath(writeGlobs []string, specID, p string) bool {
	if specID != "" {
		prefix := ".moai/specs/" + specID
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			return false
		}
	}
	for _, g := range writeGlobs {
		if contract.MatchGlob(g, p) {
			return true
		}
	}
	return false
}

// ContractReadiness is one in-push contract's facts for the readiness
// evaluation.
type ContractReadiness struct {
	SpecID      string
	Card        string // signed card
	VerifyState string // A1 verify state
	// ContractDigest is signature.contract_sha256 as the caller read it
	// ("" unsigned or unreadable — which correctly fails the same-contract
	// filter of the second-review selection).
	ContractDigest     string
	WriteGlobs         []string
	SecondReviewPolicy string // required | advisory | off (effective)
}

// EvidenceInput is the card evidence the readiness evaluator reads: file
// contents already loaded by the caller.
type EvidenceInput struct {
	ClosureReportFound   bool
	ClosureReportHeadSHA string
	SecondReviews        []SecondReviewRecord
	Verdicts             []VerdictRecord
}

// EvaluateReadiness returns the sorted, de-duplicated readiness codes of one
// in-push contract against the pushed source commit S (REQ-CLOSURE-016).
// Under an advisory or off policy no second-review code is ever reported
// (REQ-CLOSURE-018). A git failure is push_check_undetermined, never an
// allowance (REQ-CLOSURE-017).
func EvaluateReadiness(c ContractReadiness, ev EvidenceInput, sourceCommit string, facts GitFacts) []string {
	var codes []string
	add := func(code string) { codes = append(codes, code) }

	if c.VerifyState != contract.StateSignedValid {
		add(CodeContractInvalid)
	}
	if !ev.ClosureReportFound {
		add(CodeClosureReportMissing)
	} else if current, _, ok := currency(facts, c.WriteGlobs, c.SpecID, ev.ClosureReportHeadSHA, sourceCommit); !ok {
		add(CodePushCheckUndetermined)
	} else if !current {
		add(CodeClosureReportStale)
	}

	if latest := LatestVerdict(ev.Verdicts); latest != nil {
		switch latest.Verdict {
		case "reject":
			add(CodeHumanVerdictReject)
		case "amend-contract":
			add(CodeHumanVerdictAmendContract)
		}
	}

	if c.SecondReviewPolicy == PolicyRequired {
		st := SelectSecondReview(SecondReviewInput{
			Records:        ev.SecondReviews,
			ContractCard:   c.Card,
			ContractDigest: c.ContractDigest,
			EvalCommit:     sourceCommit,
			Facts:          facts,
			WriteGlobs:     c.WriteGlobs,
			SpecID:         c.SpecID,
		})
		switch {
		case st.CurrencyUndetermined:
			add(CodePushCheckUndetermined)
		case st.State == SecondReviewNotPerformed:
			add(CodeSecondReviewNotPerformed)
		case st.State == SecondReviewStale:
			add(CodeSecondReviewStale)
		case st.State == SecondReviewPerformed && st.Verdict == "fail":
			add(CodeSecondReviewFailed)
		}
	}

	sort.Strings(codes)
	return slices.Compact(codes)
}

// orNotRecorded renders an absent display string as "not recorded"
// (REQ-CLOSURE-014).
func orNotRecorded(s string) string {
	if s == "" {
		return NotRecorded
	}
	return s
}

// parseRFC3339 parses a record timestamp; an unparsable value sorts oldest.
func parseRFC3339(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// candAtAfter reports whether at is at-or-after the current best time, so a
// later append with an equal timestamp wins (the log is append-only).
func candAtAfter(at string, best time.Time) bool {
	return !parseRFC3339(at).Before(best)
}
