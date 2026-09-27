package closure

import (
	"errors"
	"slices"
	"testing"
)

// fakeFacts builds GitFacts from explicit maps: ancestors maps a descendant
// to the set of its ancestors (a commit is its own ancestor), commits maps
// a from..to range to the non-merge commits oldest-first.
func fakeFacts(ancestors map[string][]string, commits map[string][]CommitPaths) GitFacts {
	isAnc := func(a, b string) (bool, error) {
		for _, anc := range ancestors[b] {
			if anc == a {
				return true, nil
			}
		}
		return false, nil
	}
	list := func(from, to string) ([]CommitPaths, error) {
		if cs, ok := commits[from+".."+to]; ok {
			return cs, nil
		}
		return nil, nil
	}
	return GitFacts{IsAncestor: isAnc, NonMergeCommits: list}
}

// errFacts builds GitFacts whose every call fails (git broken).
func errFacts() GitFacts {
	return GitFacts{
		IsAncestor:      func(a, b string) (bool, error) { return false, errors.New("git broken") },
		NonMergeCommits: func(from, to string) ([]CommitPaths, error) { return nil, errors.New("git broken") },
	}
}

// partialErrFacts builds GitFacts where ancestry answers but range listings
// fail — the shape that reaches the currency rule with a performed record.
func partialErrFacts() GitFacts {
	return GitFacts{
		IsAncestor:      func(a, b string) (bool, error) { return a == b, nil },
		NonMergeCommits: func(from, to string) ([]CommitPaths, error) { return nil, errors.New("git broken") },
	}
}

// fixtureGlobs mirrors the acceptance fixture contract: ownership.write
// [src/**, .moai/specs/SPEC-FIXTURE-001/**].
var fixtureGlobs = []string{"src/**", ".moai/specs/SPEC-FIXTURE-001/**"}

// rec builds a second-review record with the given overrides.
type recOpt func(*SecondReviewRecord)

func recAt(head, recordedAt string, opts ...recOpt) SecondReviewRecord {
	r := SecondReviewRecord{
		SchemaVersion:  1,
		Card:           "c1",
		ContractCard:   "c1",
		SpecID:         "SPEC-FIXTURE-001",
		ContractSHA256: "digestD",
		HeadSHA:        head,
		Target:         "baseBranch",
		Scope: SecondReviewScope{
			BaseBranch:   "main",
			BaseSHA:      "base01",
			HeadSHA:      head,
			ChangedFiles: 3,
			DiffSHA256:   "diff01",
		},
		Backends: []SecondReviewBackend{
			{Backend: "claude", Gate: "required", Verdict: "pass"},
			{Backend: "codex", Gate: "required", Verdict: "pass"},
			{Backend: "glm", Gate: "advisory", Verdict: "inconclusive"},
		},
		ParticipantCount: 2,
		RecordedAt:       recordedAt,
	}
	for _, o := range opts {
		o(&r)
	}
	return r
}

func withCard(card string) recOpt   { return func(r *SecondReviewRecord) { r.Card = card } }
func withDigest(d string) recOpt    { return func(r *SecondReviewRecord) { r.ContractSHA256 = d } }
func withTarget(t string) recOpt    { return func(r *SecondReviewRecord) { r.Target = t } }
func withChangedFiles(n int) recOpt { return func(r *SecondReviewRecord) { r.Scope.ChangedFiles = n } }
func withBackends(bs ...SecondReviewBackend) recOpt {
	return func(r *SecondReviewRecord) { r.Backends = bs }
}
func recordedAt(at string) recOpt { return func(r *SecondReviewRecord) { r.RecordedAt = at } }

// TestAC_CLOSURE_013 — second-review selection and causes.
func TestAC_CLOSURE_013(t *testing.T) {
	const P = "cardhead"
	const D = "digestD"

	selectFor := func(records []SecondReviewRecord, facts GitFacts) SecondReviewState {
		return SelectSecondReview(SecondReviewInput{
			Records:        records,
			ContractCard:   "c1",
			ContractDigest: D,
			EvalCommit:     P,
			Facts:          facts,
			WriteGlobs:     fixtureGlobs,
			SpecID:         "SPEC-FIXTURE-001",
		})
	}

	t.Run("no file", func(t *testing.T) {
		st := selectFor(nil, fakeFacts(nil, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseNoRecord {
			t.Fatalf("state=%q cause=%q, want not-performed/no-record", st.State, st.Cause)
		}
	})

	t.Run("unbound card", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z", withCard("c2"))},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseUnbound {
			t.Fatalf("state=%q cause=%q, want not-performed/unbound", st.State, st.Cause)
		}
	})

	t.Run("contract changed", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z", withDigest("other"))},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseContractChanged {
			t.Fatalf("state=%q cause=%q, want not-performed/contract-changed", st.State, st.Cause)
		}
	})

	t.Run("scope not covered by target", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z", withTarget("uncommittedChanges"))},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseScopeNotCovered {
			t.Fatalf("state=%q cause=%q, want not-performed/scope-not-covered", st.State, st.Cause)
		}
	})

	t.Run("scope not covered by zero changed files", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z", withChangedFiles(0))},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseScopeNotCovered {
			t.Fatalf("state=%q cause=%q, want not-performed/scope-not-covered", st.State, st.Cause)
		}
	})

	t.Run("no second model", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z",
			withBackends(SecondReviewBackend{Backend: "claude", Gate: "required", Verdict: "pass"}))},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseNoSecondModel {
			t.Fatalf("state=%q cause=%q, want not-performed/no-second-model", st.State, st.Cause)
		}
	})

	t.Run("not in history", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt("future01", "2026-09-26T09:00:00Z")},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewNotPerformed || st.Cause != SecondReviewCauseNotInHistory {
			t.Fatalf("state=%q cause=%q, want not-performed/not-in-history", st.State, st.Cause)
		}
	})

	t.Run("performed pass with glm inconclusive", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z")},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewPerformed || st.Verdict != "pass" {
			t.Fatalf("state=%q verdict=%q, want performed/pass", st.State, st.Verdict)
		}
		if len(st.Backends) != 1 || st.Backends[0].Backend != "codex" {
			t.Fatalf("counted backends = %+v, want codex only (glm inconclusive is not counted)", st.Backends)
		}
	})

	t.Run("performed fail with glm fail", func(t *testing.T) {
		st := selectFor([]SecondReviewRecord{recAt(P, "2026-09-26T09:00:00Z",
			withBackends(
				SecondReviewBackend{Backend: "codex", Gate: "required", Verdict: "pass"},
				SecondReviewBackend{Backend: "glm", Gate: "advisory", Verdict: "fail"},
			))},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewPerformed || st.Verdict != "fail" {
			t.Fatalf("state=%q verdict=%q, want performed/fail", st.State, st.Verdict)
		}
	})

	t.Run("stale by governed change", func(t *testing.T) {
		commits := map[string][]CommitPaths{
			"review01.." + P: {{SHA: "later01", Paths: []string{"src/a.go"}}},
		}
		st := selectFor([]SecondReviewRecord{recAt("review01", "2026-09-26T09:00:00Z")},
			fakeFacts(map[string][]string{P: {P, "review01"}, "review01": {"review01"}}, commits))
		if st.State != SecondReviewStale {
			t.Fatalf("state=%q, want stale", st.State)
		}
		if st.SupersedingCommit != "later01" {
			t.Fatalf("superseding = %q, want later01", st.SupersedingCommit)
		}
	})

	t.Run("spec-directory sync commit does not stale", func(t *testing.T) {
		commits := map[string][]CommitPaths{
			"review01.." + P: {{SHA: "sync01", Paths: []string{
				".moai/specs/SPEC-FIXTURE-001/progress.md",
				".moai/specs/SPEC-FIXTURE-001/spec.md",
			}}},
		}
		st := selectFor([]SecondReviewRecord{recAt("review01", "2026-09-26T09:00:00Z")},
			fakeFacts(map[string][]string{P: {P, "review01"}, "review01": {"review01"}}, commits))
		if st.State != SecondReviewPerformed {
			t.Fatalf("state=%q, want performed (sync commit and SHA backfill do not stale)", st.State)
		}
	})

	t.Run("older valid line survives a newer scope-not-covered line", func(t *testing.T) {
		older := recAt(P, "2026-09-26T08:00:00Z")
		newer := recAt(P, "2026-09-26T10:00:00Z", withChangedFiles(0))
		st := selectFor([]SecondReviewRecord{older, newer},
			fakeFacts(map[string][]string{P: {P}}, nil))
		if st.State != SecondReviewPerformed {
			t.Fatalf("state=%q, want performed", st.State)
		}
	})

	t.Run("latest is by recorded_at", func(t *testing.T) {
		a := recAt(P, "2026-09-26T08:00:00Z", recordedAt("2026-09-26T08:00:00Z"))
		// The later-recorded record wins; give it a failing verdict so the
		// pick is observable.
		b := recAt(P, "2026-09-26T09:00:00Z", recordedAt("2026-09-26T09:00:00Z"),
			withBackends(SecondReviewBackend{Backend: "codex", Gate: "required", Verdict: "fail"}))
		st := selectFor([]SecondReviewRecord{a, b}, fakeFacts(map[string][]string{P: {P}}, nil))
		if st.Verdict != "fail" {
			t.Fatalf("verdict=%q, want fail (the later-recorded record wins)", st.Verdict)
		}
	})
}

// TestAC_CLOSURE_016 — the closed nine-code readiness set.
func TestAC_CLOSURE_016(t *testing.T) {
	if got, want := ReadinessCodes(), []string{
		CodeClosureReportMissing,
		CodeClosureReportStale,
		CodeContractInvalid,
		CodeHumanVerdictAmendContract,
		CodeHumanVerdictReject,
		CodePushCheckUndetermined,
		CodeSecondReviewFailed,
		CodeSecondReviewNotPerformed,
		CodeSecondReviewStale,
	}; !slices.Equal(got, want) {
		t.Fatalf("ReadinessCodes() = %v, want %v", got, want)
	}

	ready := ContractReadiness{
		SpecID:             "SPEC-FIXTURE-001",
		Card:               "c1",
		VerifyState:        "signed-valid",
		ContractDigest:     "digestD",
		WriteGlobs:         fixtureGlobs,
		SecondReviewPolicy: "required",
	}
	history := fakeFacts(map[string][]string{
		"cardhead": {"cardhead"}, "report01": {"report01", "cardhead"},
	}, nil)

	// producedByContains asserts the fixture produces the named code (other
	// codes may ride along: the AC is "each code is produced by its fixture").
	producedByContains := func(name string, code string, mutate func(*ContractReadiness, *EvidenceInput, *GitFacts)) {
		t.Run(name, func(t *testing.T) {
			c, ev, facts := ready, EvidenceInput{}, history
			mutate(&c, &ev, &facts)
			got := EvaluateReadiness(c, ev, "cardhead", facts)
			if !slices.Contains(got, code) {
				t.Fatalf("codes = %v, want it to contain %s", got, code)
			}
		})
	}
	// readyEvidence is the fully ready evidence set (current report,
	// performed pass review at the evaluation commit).
	readyEvidence := func() EvidenceInput {
		return EvidenceInput{
			ClosureReportFound:   true,
			ClosureReportHeadSHA: "cardhead",
			SecondReviews:        []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z")},
		}
	}

	producedByContains("contract invalid", CodeContractInvalid, func(c *ContractReadiness, _ *EvidenceInput, _ *GitFacts) {
		c.VerifyState = "unsigned"
	})

	producedByContains("closure report missing", CodeClosureReportMissing, func(_ *ContractReadiness, ev *EvidenceInput, _ *GitFacts) {
		*ev = EvidenceInput{} // no report, no records
	})

	producedByContains("closure report stale", CodeClosureReportStale, func(_ *ContractReadiness, ev *EvidenceInput, facts *GitFacts) {
		ev.ClosureReportFound = true
		ev.ClosureReportHeadSHA = "report01"
		ev.SecondReviews = []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z")}
		*facts = fakeFacts(map[string][]string{
			"cardhead": {"cardhead"}, "report01": {"report01", "cardhead"},
		}, map[string][]CommitPaths{
			"report01..cardhead": {{SHA: "mid01", Paths: []string{"src/x.go"}}},
		})
	})

	producedByContains("human verdict reject", CodeHumanVerdictReject, func(_ *ContractReadiness, ev *EvidenceInput, _ *GitFacts) {
		*ev = readyEvidence()
		ev.Verdicts = []VerdictRecord{{Verdict: "reject"}}
	})

	producedByContains("human verdict amend-contract", CodeHumanVerdictAmendContract, func(_ *ContractReadiness, ev *EvidenceInput, _ *GitFacts) {
		*ev = readyEvidence()
		ev.Verdicts = []VerdictRecord{{Verdict: "amend-contract"}}
	})

	producedByContains("second review not performed", CodeSecondReviewNotPerformed, func(_ *ContractReadiness, ev *EvidenceInput, _ *GitFacts) {
		ev.ClosureReportFound = true
		ev.ClosureReportHeadSHA = "cardhead"
		// no second-review records
	})

	producedByContains("second review stale", CodeSecondReviewStale, func(_ *ContractReadiness, ev *EvidenceInput, facts *GitFacts) {
		ev.ClosureReportFound = true
		ev.ClosureReportHeadSHA = "cardhead"
		ev.SecondReviews = []SecondReviewRecord{recAt("review01", "2026-09-26T09:00:00Z")}
		*facts = fakeFacts(map[string][]string{
			// descendant -> its ancestors (a commit is its own ancestor).
			"cardhead": {"cardhead", "review01"}, "review01": {"review01"},
		}, map[string][]CommitPaths{
			"review01..cardhead": {{SHA: "mid01", Paths: []string{"src/a.go"}}},
		})
	})

	producedByContains("second review failed", CodeSecondReviewFailed, func(_ *ContractReadiness, ev *EvidenceInput, _ *GitFacts) {
		*ev = readyEvidence()
		ev.SecondReviews = []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z",
			withBackends(SecondReviewBackend{Backend: "codex", Gate: "required", Verdict: "fail"}))}
	})

	producedByContains("git broken is undetermined", CodePushCheckUndetermined, func(_ *ContractReadiness, ev *EvidenceInput, facts *GitFacts) {
		*ev = readyEvidence()
		*facts = errFacts()
	})

	t.Run("failed review does not also report not-performed", func(t *testing.T) {
		ev := readyEvidence()
		ev.SecondReviews = []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z",
			withBackends(SecondReviewBackend{Backend: "codex", Gate: "required", Verdict: "fail"}))}
		got := EvaluateReadiness(ready, ev, "cardhead", history)
		if !slices.Equal(got, []string{CodeSecondReviewFailed}) {
			t.Fatalf("codes = %v, want only second_review_failed", got)
		}
	})

	t.Run("accept verdict and no verdict are ready", func(t *testing.T) {
		for name, verdicts := range map[string][]VerdictRecord{
			"accept":     {{Verdict: "accept"}},
			"no verdict": nil,
		} {
			ev := readyEvidence()
			ev.Verdicts = verdicts
			if got := EvaluateReadiness(ready, ev, "cardhead", history); len(got) != 0 {
				t.Fatalf("%s: codes = %v, want ready (empty)", name, got)
			}
		}
	})

	t.Run("two codes sorted", func(t *testing.T) {
		ev := readyEvidence()
		ev.SecondReviews = []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z",
			withBackends(SecondReviewBackend{Backend: "codex", Gate: "required", Verdict: "fail"}))}
		ev.Verdicts = []VerdictRecord{{Verdict: "reject"}}
		got := EvaluateReadiness(ready, ev, "cardhead", history)
		want := []string{CodeHumanVerdictReject, CodeSecondReviewFailed}
		if !slices.Equal(got, want) {
			t.Fatalf("codes = %v, want %v", got, want)
		}
	})
}

// TestAC_CLOSURE_018 — advisory and off report no second-review code.
func TestAC_CLOSURE_018(t *testing.T) {
	for _, policy := range []string{"advisory", "off"} {
		t.Run(policy, func(t *testing.T) {
			c := ContractReadiness{
				SpecID: "SPEC-FIXTURE-001", Card: "c1", VerifyState: "signed-valid",
				WriteGlobs: fixtureGlobs, SecondReviewPolicy: policy,
			}
			ev := EvidenceInput{ClosureReportFound: true, ClosureReportHeadSHA: "cardhead"}
			got := EvaluateReadiness(c, ev, "cardhead",
				fakeFacts(map[string][]string{"cardhead": {"cardhead"}}, nil))
			if len(got) != 0 {
				t.Fatalf("codes = %v, want none under %s", got, policy)
			}
		})
	}
}

// TestSelectSecondReviewCurrencyUndetermined pins the honest undetermined
// shape: the performed verdict is stated, currency undetermined is flagged,
// and the push evaluator maps it to push_check_undetermined. The facts fail
// only at the range listing — with ancestry broken too, the in-history
// filter excludes the record first (fail-closed), which the AC-016
// fully-broken-git fixture covers.
func TestSelectSecondReviewCurrencyUndetermined(t *testing.T) {
	st := SelectSecondReview(SecondReviewInput{
		Records:        []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z")},
		ContractCard:   "c1",
		ContractDigest: "digestD",
		EvalCommit:     "cardhead",
		Facts:          partialErrFacts(),
		WriteGlobs:     fixtureGlobs,
		SpecID:         "SPEC-FIXTURE-001",
	})
	if st.State != SecondReviewPerformed || st.Verdict != "pass" {
		t.Fatalf("state=%q verdict=%q, want performed/pass kept", st.State, st.Verdict)
	}
	if !st.CurrencyUndetermined {
		t.Fatalf("CurrencyUndetermined = false, want true")
	}

	c := ContractReadiness{
		SpecID: "SPEC-FIXTURE-001", Card: "c1", VerifyState: "signed-valid",
		WriteGlobs: fixtureGlobs, SecondReviewPolicy: "required",
	}
	ev := EvidenceInput{
		ClosureReportFound:   true,
		ClosureReportHeadSHA: "cardhead",
		SecondReviews:        []SecondReviewRecord{recAt("cardhead", "2026-09-26T09:00:00Z")},
	}
	got := EvaluateReadiness(c, ev, "cardhead", errFacts())
	// Under fully broken git the in-history filter also excludes the record
	// (fail-closed), so second_review_not_performed rides along. Either way
	// the push is stopped, and push_check_undetermined is present.
	if !slices.Equal(got, []string{CodePushCheckUndetermined, CodeSecondReviewNotPerformed}) {
		t.Fatalf("codes = %v, want [push_check_undetermined second_review_not_performed]", got)
	}
}

// TestSubstituteBackend pins REQ-CLOSURE-014: the performing backend
// differing from the contract's second_model is flagged.
func TestSubstituteBackend(t *testing.T) {
	glmOnly := recAt("cardhead", "2026-09-26T09:00:00Z", withBackends(
		SecondReviewBackend{Backend: "glm", Gate: "required", Verdict: "pass"}))
	st := SelectSecondReview(SecondReviewInput{
		Records:             []SecondReviewRecord{glmOnly},
		ContractCard:        "c1",
		ContractDigest:      "digestD",
		ContractSecondModel: "codex",
		EvalCommit:          "cardhead",
		Facts:               fakeFacts(map[string][]string{"cardhead": {"cardhead"}}, nil),
		WriteGlobs:          fixtureGlobs,
		SpecID:              "SPEC-FIXTURE-001",
	})
	if !st.SubstituteBackend {
		t.Fatalf("SubstituteBackend = false, want true (glm performed, contract names codex)")
	}
	if st.ContractSecondModel != "codex" {
		t.Fatalf("ContractSecondModel = %q, want codex", st.ContractSecondModel)
	}

	codex := recAt("cardhead", "2026-09-26T09:00:00Z")
	st = SelectSecondReview(SecondReviewInput{
		Records:             []SecondReviewRecord{codex},
		ContractCard:        "c1",
		ContractDigest:      "digestD",
		ContractSecondModel: "codex",
		EvalCommit:          "cardhead",
		Facts:               fakeFacts(map[string][]string{"cardhead": {"cardhead"}}, nil),
		WriteGlobs:          fixtureGlobs,
		SpecID:              "SPEC-FIXTURE-001",
	})
	if st.SubstituteBackend {
		t.Fatalf("SubstituteBackend = true, want false (codex performed)")
	}
}

// TestReportStaleNaming pins the superseding commit is the first (oldest)
// non-merge commit that touched a governed path.
func TestReportStaleNaming(t *testing.T) {
	facts := fakeFacts(
		map[string][]string{"p": {"p", "r"}, "r": {"r"}},
		map[string][]CommitPaths{"r..p": {
			{SHA: "first01", Paths: []string{"src/a.go"}},
			{SHA: "then01", Paths: []string{"src/b.go"}},
		}},
	)
	st := SelectSecondReview(SecondReviewInput{
		Records:        []SecondReviewRecord{recAt("r", "2026-09-26T09:00:00Z")},
		ContractCard:   "c1",
		ContractDigest: "digestD",
		EvalCommit:     "p",
		Facts:          facts,
		WriteGlobs:     fixtureGlobs,
		SpecID:         "SPEC-FIXTURE-001",
	})
	if st.State != SecondReviewStale || st.SupersedingCommit != "first01" {
		t.Fatalf("state=%q superseding=%q, want stale/first01", st.State, st.SupersedingCommit)
	}
}
