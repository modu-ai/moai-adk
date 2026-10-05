package closure

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// ─── Summary ───

func buildSummary(r *Report, in BuildInput) {
	terminal := in.Verify.Terminal
	r.Summary = SummarySection{
		VerifyState:   orNotObserved(in.Verify.State),
		VerifyReasons: nonNil(in.Verify.Reasons),
		Terminal:      &terminal,
	}
}

// ─── Kickoff (REQ-CLOSURE-010) ───

func buildKickoff(r *Report, in BuildInput) {
	k := KickoffSection{
		Method:      NotRecorded,
		SignerKind:  NotRecorded,
		LLMAnswer:   NotRecorded,
		JevAnswer:   NotRecorded,
		VerifyState: orNotObserved(in.Verify.State),
	}
	if in.Verify.Contract == nil || in.Verify.Contract.Signature == nil {
		r.Kickoff = k
		return
	}
	sig := in.Verify.Contract.Signature
	k.Method = orNotRecorded(sig.Method)
	k.SignerKind = orNotRecorded(sig.SignerKind)
	if sig.Receipt == nil {
		r.Kickoff = k
		return
	}
	if !in.ReceiptPresent || len(in.Receipt) == 0 {
		// Signature method receipt and the receipt file is absent.
		k.ReceiptPresent = false
		r.AddNotPerformed(NotPerformedReceiptMissing, "signature method is receipt but the receipt file is absent")
		r.Kickoff = k
		return
	}
	view, unknown, err := DecodeKickoffReceipt(in.Receipt)
	if err != nil {
		r.AddNotPerformed(NotPerformedReceiptMissing, "receipt file unreadable: "+err.Error())
		r.Kickoff = k
		return
	}
	k.ReceiptPresent = true
	k.RequestedDecider = orNotRecorded(view.RequestedDecider)
	k.EffectiveDecider = orNotRecorded(view.EffectiveDecider)
	k.FallbackApplied = view.FallbackApplied
	k.FallbackReason = view.FallbackReason
	k.LLMAnswer = orNotRecorded(view.LLMAnswer)
	k.LLMConfidence = view.LLMConfidence
	k.JevAnswer = orNotRecorded(view.JevAnswer)
	k.JevConfidence = view.JevConfidence
	k.Outcome = orNotRecorded(view.Outcome)
	for _, f := range unknown {
		r.AddNotPerformed(NotPerformedReceiptFieldUnrecognized, "receipt field not recognized: "+f)
	}
	r.Kickoff = k
}

// ─── Reconciliation (REQ-CLOSURE-004) ───

func buildReconciliation(r *Report, in BuildInput) {
	sec := ReconciliationSection{
		Rows:          []ReconciliationRow{},
		VerifyState:   orNotObserved(in.Verify.State),
		VerifyReasons: nonNil(in.Verify.Reasons),
	}

	// Recorded versus measured acceptance binding, from A1's verify report.
	if in.Verify.Acceptance.SHA256 != nil {
		sec.RecordedAcceptanceSHA = *in.Verify.Acceptance.SHA256
	} else {
		sec.RecordedAcceptanceSHA = NotRecorded
	}
	sec.MeasuredAcceptanceSHA = in.Verify.Acceptance.MeasuredSHA256
	if in.Verify.Acceptance.ACCount != nil {
		sec.RecordedACCount = strconv.Itoa(*in.Verify.Acceptance.ACCount)
	} else {
		sec.RecordedACCount = NotRecorded
	}
	sec.MeasuredACCount = in.Verify.Acceptance.MeasuredACCount

	// Live AC IDs (A1 counter semantics, guarded against A1's count).
	liveIDs, idsOK := liveACIDs(in)
	if idsOK {
		rows := map[string]ProgressRow{}
		var unknown []ProgressRow
		if in.ProgressMD != nil {
			prows, _ := ParseProgressRows(in.ProgressMD)
			for _, pr := range prows {
				if _, isLive := liveIDs[pr.ID]; isLive {
					rows[pr.ID] = pr
				} else {
					unknown = append(unknown, pr)
				}
			}
		}
		for _, id := range sortedKeys(liveIDs) {
			if pr, ok := rows[id]; ok {
				sec.Rows = append(sec.Rows, ReconciliationRow{ID: id, Status: pr.Status, Evidence: pr.Evidence})
				continue
			}
			sec.Rows = append(sec.Rows, ReconciliationRow{ID: id, Status: "not reported", Evidence: ""})
			r.AddNotPerformed(NotPerformedACNotReported, id+" has no §E.2 matrix row")
		}
		for _, pr := range unknown {
			sec.Rows = append(sec.Rows, ReconciliationRow{ID: pr.ID, Status: "unknown", Evidence: pr.Evidence})
		}
	}
	r.Reconciliation = sec
}

// liveACIDs extracts the live acceptance-criterion IDs with A1's counter
// grammar (REQ-CLOSURE-004 "A1 counter semantics"). contract.CountAC returns
// only counts, so the IDs are extracted here and GUARDED against A1's live
// count: on disagreement the IDs are discarded (never rendered wrong) and
// the section carries the count only.
func liveACIDs(in BuildInput) (map[string]bool, bool) {
	if in.AcceptanceMD == nil || !in.ACAvailable {
		return nil, false
	}
	normalized := contract.NormalizeAcceptance(in.AcceptanceMD)
	ids, err := extractLiveACIDs(normalized)
	if err != nil {
		return nil, false
	}
	if len(ids) != in.MeasuredAC.Live || len(in.MeasuredAC.Ambiguous) > 0 {
		return nil, false
	}
	return ids, true
}

var (
	acPrefixDeclRe  = regexp.MustCompile(`(?m)^<!-- *moai-ac-prefix: *(.*?) *-->.*$`)
	acIDReCache     = map[string]*regexp.Regexp{}
	acMarkerAfterRe = regexp.MustCompile(`^[ \t]*(\[RETIRED\]|\[REF\])`)
)

// extractLiveACIDs mirrors contract.CountAC's grammar (declaration-aware
// prefix, same identifier shape, same [RETIRED]/[REF] marker test) to list
// the live IDs in first-seen order. The liveACIDs guard against CountAC's
// count keeps any drift loud instead of silent.
func extractLiveACIDs(normalized []byte) (map[string]bool, error) {
	prefix := "AC"
	if m := acPrefixDeclRe.FindStringSubmatch(string(normalized)); m != nil {
		if d := strings.ReplaceAll(m[1], " ", ""); d != "" {
			prefix = d
		}
	}
	re, ok := acIDReCache[prefix]
	if !ok {
		var err error
		re, err = regexp.Compile("(" + prefix + ")-([A-Z0-9]+-)*[0-9]+[a-z]?")
		if err != nil {
			return nil, err
		}
		acIDReCache[prefix] = re
	}
	marked := map[string]bool{}
	unmarked := map[string]bool{}
	for _, line := range strings.Split(string(normalized), "\n") {
		rest := line
		for {
			loc := re.FindStringIndex(rest)
			if loc == nil {
				break
			}
			id := rest[loc[0]:loc[1]]
			rest = rest[loc[1]:]
			if acMarkerAfterRe.MatchString(rest) {
				marked[id] = true
			} else {
				unmarked[id] = true
			}
		}
	}
	ids := map[string]bool{}
	for id := range unmarked {
		if !marked[id] {
			ids[id] = true
		}
	}
	return ids, nil
}

// ─── First Verdict (REQ-CLOSURE-009) ───

func buildFirstVerdict(r *Report, in BuildInput) {
	fv := FirstVerdictSection{
		RunStatus:    NotRecorded,
		ACPassCount:  NotRecorded,
		ACFailCount:  NotRecorded,
		RunCommitSHA: NotRecorded,
	}
	if in.ProgressMD == nil {
		r.AddNotPerformed(NotPerformedProgressMissing, "no progress.md")
		r.FirstVerdict = fv
		return
	}
	pv := ParseFirstVerdict(in.ProgressMD)
	if !pv.Present {
		r.AddNotPerformed(NotPerformedFirstVerdictMissing, "no §E.3 block")
		r.FirstVerdict = fv
		return
	}
	fv.RunStatus = orNotRecorded(pv.RunStatus)
	fv.ACPassCount = orNotRecorded(pv.ACPassCount)
	fv.ACFailCount = orNotRecorded(pv.ACFailCount)
	fv.RunCommitSHA = orNotRecorded(pv.RunCommitSHA)
	if fv.RunStatus == NotRecorded || fv.ACPassCount == NotRecorded || fv.ACFailCount == NotRecorded {
		r.AddNotPerformed(NotPerformedFirstVerdictMissing, "a §E.3 field is absent")
	}
	// Count mismatch flag: reported pass+fail versus the measured AC count.
	if p, perr := strconv.Atoi(pv.ACPassCount); perr == nil {
		if f, ferr := strconv.Atoi(pv.ACFailCount); ferr == nil && in.ACAvailable {
			fv.CountMismatch = p+f != in.MeasuredAC.Live
		}
	}
	r.FirstVerdict = fv
}

// ─── shared helpers ───

func orNotObserved(s string) string {
	if s == "" {
		return "not observed"
	}
	return s
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// loadContractSideFiles loads everything Build needs from the SPEC directory
// and the card evidence home. Kept next to BuildInput so the CLI wiring
// cannot forget a field: policy values, registry values, and the spec status
// are the caller's (same convention as contract.LoadDir).
func LoadBuildSideFiles(root, specID string, autonomy config.AutonomySettings, ruleIDs, frozenFiles []string) (contract.Report, []byte, []byte, contract.ACCountResult, bool, error) {
	dir, err := contract.ResolveSpecDir(root, specID)
	if err != nil {
		return contract.Report{}, nil, nil, contract.ACCountResult{}, false, err
	}
	inputs, err := contract.LoadDir(dir)
	if err != nil {
		return contract.Report{}, nil, nil, contract.ACCountResult{}, false, err
	}
	inputs.Policy = contract.Policy{
		SecondReview: autonomy.SecondReview,
		PushDevelop:  autonomy.PushDevelop,
		Mode:         autonomy.Mode,
	}
	inputs.RegistryRuleIDs = ruleIDs
	inputs.RegistryFrozenFiles = frozenFiles
	if status, serr := spec.ParseStatus(dir); serr == nil {
		inputs.SpecStatus = status
	}
	rep := contract.Verify(inputs)

	var progress, acceptance []byte
	progress, _, _ = readFileOrEmpty(dir + "/progress.md")
	acceptance = inputs.Acceptance
	var ac contract.ACCountResult
	acOK := false
	if len(acceptance) > 0 {
		if a, err := contract.CountAC(contract.NormalizeAcceptance(acceptance)); err == nil {
			ac, acOK = a, true
		}
	}
	return rep, progress, acceptance, ac, acOK, nil
}
