package runtime

// audit_ceiling.go — the ceiling-policy outcome engine
// (SPEC-AUDIT-CEILING-001 REQ-ACE-003..007, REQ-ACE-013). Evaluated at a
// LIVE run-entry admission seam before the verdict reaches the admitting
// consumer, for a SPEC at a ceiling state — the effective ceiling reached
// (tier ceiling + auto_delta_rounds), or the tier ceiling reached on a final
// hit with no eligible delta round. The ladder evaluates pass-through FIRST
// (an admission-clean verdict admits at any ceiling state, REQ-ACE-013);
// for a verdict that fails admission the rungs are debt-admit (label-only
// failure, REQ-ACE-004), split (anchored blocking findings, REQ-ACE-005),
// and hold (everything else, REQ-ACE-006). No rung asks a question (C3);
// every outcome is recorded to the SPEC's progress.md §G Override and
// Refusal Record and the machine-local audit trail (REQ-ACE-007/012).

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/auditverdict"
	"github.com/modu-ai/moai-adk/internal/config"
)

// Outcome vocabulary (design.md §2).
const (
	OutcomePassThrough = "pass-through"
	OutcomeDebtAdmit   = "debt-admit"
	OutcomeSplit       = "split"
	OutcomeHold        = "hold"
)

// VerdictCeilingOutcome is the structured record of one ceiling evaluation. Refusal
// is observable via Blocked; a pass-through carries Blocked == false and
// records that the ceiling was reached (REQ-ACE-013).
type VerdictCeilingOutcome struct {
	Outcome  string              `json:"outcome"`
	Reasons  []string            `json:"reasons"`
	Evidence []string            `json:"evidence"`
	Blocked  bool                `json:"blocked"`
	Debts    []auditverdict.Debt `json:"debts,omitempty"`
}

// VerdictCeilingInput names the SPEC a ceiling evaluation runs for. SpecDir is the
// SPEC directory (tier, threshold, and the progress.md §G record target);
// ProjectRoot is the tree whose harness config and report directories the
// engine reads.
type VerdictCeilingInput struct {
	SpecID      string
	SpecDir     string
	ProjectRoot string
	CardID      string
}

var (
	fixScopeLine = regexp.MustCompile(`(?im)^\s*fix_scope:\s*(\S.*)$`)
	stopSignal   = regexp.MustCompile(`(?im)^\s*STOP\b`)
	reqACID      = regexp.MustCompile(`\b(?:REQ|AC)-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`)
)

// EvaluateCeiling decides the ceiling-policy outcome for one presented
// verdict at a run-entry admission seam. fields is the parsed verdict, hashOK
// the seam's plan-artifact hash binding, required the tree's resolved
// required-backend set. Returns nil (no ceiling state — the seam proceeds
// with its normal admission), or the outcome plus override: override is true
// only for a debt-admit outcome, where the seam substitutes the outcome's
// PASS-WITH-DEBT for the raw verdict label in the label check alone (D20).
// A counter or evidence-read error is returned for the seam to refuse on —
// an unverifiable count must not silently read as below the ceiling.
//
// @MX:ANCHOR: [AUTO] the ceiling-policy outcome ladder every LIVE run-entry admission seam evaluates before admitting a verdict
// @MX:REASON: fan_in 3 measured — kickoff planAuditCheck, homestate admitVerdictFile, and GateConfig.attachCeiling all resolve through it; a second copy of the ladder is the drift this SPEC exists to close
// @MX:SPEC: SPEC-AUDIT-CEILING-001
func EvaluateCeiling(in VerdictCeilingInput, fields auditverdict.Fields, hashOK bool, required []string) (*VerdictCeilingOutcome, bool, error) {
	tier := auditverdict.SpecTier(in.SpecDir)
	threshold := auditverdict.PlanThreshold(in.SpecDir)
	tierCeiling, deltaRounds, policyNamed, cfgErr := loadCeilings(in.ProjectRoot, tier)
	if cfgErr != nil {
		return nil, false, cfgErr
	}

	ev, err := CountAuditRounds(in.SpecID, RoundReportDirs(in.ProjectRoot, in.SpecID, in.CardID))
	if err != nil {
		return nil, false, fmt.Errorf("count audit rounds: %w", err)
	}
	if ev.Count < tierCeiling {
		return nil, false, nil // below the tier ceiling: no ceiling state
	}

	// Delta eligibility (REQ-ACE-003, AC-ACE-021): an eligible delta round
	// proceeds under the prose policy; an ineligible delta or a STOP signal
	// is a final hit, adjudicated by the ladder below.
	latestRaw := readEvidenceRaw(ev.LatestPath)
	anchors, stop := deltaMarkers(latestRaw)
	prevSHA := previousAuditedSHA(in, ev)
	diffOK, reqACOK := false, false
	if audited := auditedSHAOf(ev.LatestPath); prevSHA != "" && audited != "" {
		diffOK = diffInsideAnchors(in.ProjectRoot, prevSHA, audited, anchors)
		reqACOK = reqACSetsUnchanged(in.ProjectRoot, in.SpecID, prevSHA, audited)
	}
	deltaOK := DeltaEligible(anchors, stop, diffOK, reqACOK)
	if ev.Count < tierCeiling+deltaRounds && deltaOK {
		return nil, false, nil
	}

	// Rung 0 — pass-through (REQ-ACE-013, D31): an admission-clean verdict
	// admits at any ceiling state, without a question.
	if ok, _ := auditverdict.AdmitWithRequired(fields, auditverdict.PhasePlan, threshold, hashOK, required); ok {
		oc := &VerdictCeilingOutcome{
			Outcome:  OutcomePassThrough,
			Blocked:  false,
			Evidence: ev.Sources,
			Reasons: []string{fmt.Sprintf(
				"plan-audit ceiling reached (round count %d >= tier ceiling %d + %d delta rounds%s); the verdict is admission-clean and admits without a question (REQ-ACE-013)",
				ev.Count, tierCeiling, deltaRounds, policyNote(policyNamed))},
		}
		persistOutcome(in, "ceiling-outcome", oc)
		return oc, false, nil
	}

	// Rung 1 — debt-admit (REQ-ACE-004, Q5's SPEC-embedded default): the
	// verdict fails admission on the label alone. Eligibility is the
	// converted-label run of the real predicate — if Admit passes with the
	// label replaced, the label was the only refusing check (score, must-
	// pass, blocking, hash, duplicate keys, and the receipt checks all
	// evaluated the raw verdict and passed). The enumerated debts are the
	// verdict's own machine debt lines; the auditor's verdict file is never
	// rewritten.
	if fields.Label != auditverdict.LabelPassWithDebt {
		converted := fields
		converted.Label = auditverdict.LabelPassWithDebt
		if ok, _ := auditverdict.AdmitWithRequired(converted, auditverdict.PhasePlan, threshold, hashOK, required); ok {
			oc := &VerdictCeilingOutcome{
				Outcome:  OutcomeDebtAdmit,
				Blocked:  false,
				Debts:    fields.Debts,
				Evidence: ev.Sources,
				Reasons: []string{fmt.Sprintf(
					"plan-audit ceiling reached (round count %d >= tier ceiling %d%s); the verdict fails admission on the label alone and debt-admits — findings recorded as PASS-WITH-DEBT debts, entry admitted without a question (REQ-ACE-004)",
					ev.Count, tierCeiling, policyNote(policyNamed))},
			}
			persistOutcome(in, "ceiling-outcome", oc)
			return oc, true, nil
		}
	}

	// Rung 2 — split (REQ-ACE-005): blocking findings each carrying a scoped
	// fix anchor produce a hold record plus a split proposal naming the
	// anchored scope.
	if fields.BlockingKnown && fields.BlockingCount > 0 && len(anchors) > 0 {
		oc := &VerdictCeilingOutcome{
			Outcome:  OutcomeSplit,
			Blocked:  true,
			Evidence: ev.Sources,
			Reasons: []string{fmt.Sprintf(
				"plan-audit ceiling reached with %d blocking finding(s) each carrying scoped fix anchors (%s): hold record plus split proposal naming the anchored scope, entry blocked (REQ-ACE-005)",
				fields.BlockingCount, strings.Join(anchors, " "))},
		}
		persistOutcome(in, "ceiling-refusal", oc)
		return oc, false, nil
	}

	// Rung 3 — hold (REQ-ACE-006): everything else (a stale hash, no
	// findings, unanchored blocking findings, missing fields). The hold
	// names its release path so it is never a silent dead end (D30).
	oc := &VerdictCeilingOutcome{
		Outcome:  OutcomeHold,
		Blocked:  true,
		Evidence: ev.Sources,
		Reasons: []string{fmt.Sprintf(
			"plan-audit ceiling reached (round count %d >= tier ceiling %d%s); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G",
			ev.Count, tierCeiling, policyNote(policyNamed))},
	}
	persistOutcome(in, "ceiling-refusal", oc)
	return oc, false, nil
}

// loadCeilings reads the SPEC's tier ceiling and the delta-round count from
// the tree's harness config. A genuinely ABSENT config reads fail-closed on
// the template defaults (the same posture the loader's absence path takes);
// a config that exists but cannot be read or parsed is an ERROR the seam
// refuses on, never a silent fallback to the defaults (card-review F1). An
// on_final_hit NAME the engine does not implement passes the load without
// an error (SPEC-AUDIT-CEILING-002 matrix M2/M12 pass-through) and reads
// fail-closed here: policyNamed=false, deltaRounds forced to zero — the
// delta ladder is unreachable for a policy the CLI cannot enforce.
func loadCeilings(projectRoot, tier string) (tierCeiling int, deltaRounds int, policyNamed bool, err error) {
	def := config.PlanAuditTierCeilingsConfig{}.Defaults()
	tierCeiling = def.L
	switch tier {
	case "S":
		tierCeiling = def.S
	case "M":
		tierCeiling = def.M
	}
	cfg, err := config.LoadHarnessConfig(filepath.Join(projectRoot, ".moai", "config", "sections", "harness.yaml"))
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			return tierCeiling, 0, false, nil
		}
		return 0, 0, false, fmt.Errorf("read harness config: %w", err)
	}
	switch tier {
	case "S":
		if cfg.PlanAuditTierCeilings["S"] > 0 {
			tierCeiling = cfg.PlanAuditTierCeilings["S"]
		}
	case "M":
		if cfg.PlanAuditTierCeilings["M"] > 0 {
			tierCeiling = cfg.PlanAuditTierCeilings["M"]
		}
	default:
		if cfg.PlanAuditTierCeilings["L"] > 0 {
			tierCeiling = cfg.PlanAuditTierCeilings["L"]
		}
	}
	if cfg.PlanAuditCeilingPolicy.AutoDeltaRounds > 0 {
		deltaRounds = cfg.PlanAuditCeilingPolicy.AutoDeltaRounds
	}
	policyNamed = cfg.PlanAuditCeilingPolicy.OnFinalHit == "hold-and-split"
	if !policyNamed {
		return tierCeiling, 0, false, nil // an unnamed policy: no delta round is granted
	}
	return tierCeiling, deltaRounds, true, nil
}

func policyNote(named bool) string {
	if named {
		return ""
	}
	return ", on_final_hit policy unnamed — fail-closed"
}

// DeltaEligible is the delta-round eligibility predicate (REQ-ACE-003,
// AC-ACE-021): anchors present, no STOP signal, the diff between the two
// audited SHAs inside the anchors, and the REQ/AC id sets unchanged. Any
// other combination is a final hit.
func DeltaEligible(anchors []string, stop bool, diffInAnchors bool, reqACUnchanged bool) bool {
	if stop || len(anchors) == 0 {
		return false
	}
	return diffInAnchors && reqACUnchanged
}

// deltaMarkers extracts the machine-observable delta evidence from the
// latest verdict: its fix_scope anchors and its STOP signal.
func deltaMarkers(latestRaw []byte) (anchors []string, stop bool) {
	if len(latestRaw) == 0 {
		return nil, false
	}
	if m := fixScopeLine.FindStringSubmatch(string(latestRaw)); m != nil {
		anchors = strings.Fields(m[1])
	}
	stop = stopSignal.Match(latestRaw)
	return anchors, stop
}

// previousAuditedSHA returns the audited SHA of the previous round — the
// largest round strictly below the latest, deduped across both families and
// including the legacy stream (card-review F4), with the bare base report
// ordering earliest as round 0 (REQ-ACR-009), "" when no prior round exists.
// Both the latest round number and every source round number resolve through
// the same dual-family, base-aware parse: the counter counts BOTH families
// and a legacy-family file can be the selected LatestPath, so resolving the
// latest through the convention family alone would lose the diff baseline
// whenever the stream tip is a legacy verdict (REQ-ACR-001).
func previousAuditedSHA(in VerdictCeilingInput, ev RoundEvidence) string {
	if ev.LatestPath == "" {
		return ""
	}
	latestN, ok := evidenceRoundOf(filepath.Base(ev.LatestPath), in.SpecID)
	if !ok || latestN <= 0 {
		// An unparseable or non-positive latest round number stays
		// fail-closed: no guessed, defaulted, or sibling-derived baseline
		// (REQ-ACR-002).
		return ""
	}
	best := -1
	prev := ""
	for _, src := range ev.Sources {
		if src == ev.LatestPath {
			continue
		}
		n, ok := evidenceRoundOf(filepath.Base(src), in.SpecID)
		if !ok || n >= latestN {
			continue
		}
		if n > best {
			best = n
			prev = src
		}
	}
	if prev == "" {
		return ""
	}
	return auditedSHAOf(prev)
}

// evidenceRoundOf is the scan-local dual-family, base-aware parse of one
// evidence file name (plan.md §G sanctioned exception): the convention
// family, with the bare base report ranking as round 0 — the engine's own
// plan-round convention (planAuditRoundFile) — and an explicitly
// 0-numbered convention file equally valid, and the legacy family
// (<SpecID>-review-<N>.md) with identical round-0 validity and ordering:
// renaming the same history between families must not change the
// previous-SHA selection (round-3 repair 2). ok is false when the name
// parses under neither family or its number does not parse: a name the
// caller cannot order must not yield a baseline.
func evidenceRoundOf(name, specID string) (n int, ok bool) {
	if m := conventionFile.FindStringSubmatch(name); m != nil {
		if m[1] == "" {
			return 0, true // the bare base report — round 0, earliest
		}
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, false // a convention shape whose number does not parse
		}
		return v, true // a parsed 0 is a valid round 0 — family parity
	}
	legacyFile := regexp.MustCompile(`^` + regexp.QuoteMeta(specID) + `-review-([0-9]+)\.md$`)
	if m := legacyFile.FindStringSubmatch(name); m != nil {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, false
		}
		return v, true
	}
	return 0, false
}

// diffInsideAnchors reports whether every change git names between the two
// audited SHAs stays inside the fix_scope anchors — the anchor's #token is
// load-bearing: a hunk in an anchored file verifies only when its context
// references one of that file's anchors (card-review F3 — a filename-only
// check admitted out-of-anchor edits), while edits to progress.md and the
// report tree stay always-allowed (the prose policy's own composition, D6).
// A git failure is fail-closed: the delta is not verified.
func diffInsideAnchors(projectRoot, fromSHA, toSHA string, anchors []string) bool {
	if len(anchors) == 0 {
		return false
	}
	// anchorTokens maps a changed file to the anchor tokens scoped to it.
	anchorTokens := map[string][]string{}
	for _, a := range anchors {
		file, token, _ := strings.Cut(a, "#")
		file = strings.TrimSpace(file)
		anchorTokens[file] = append(anchorTokens[file], strings.TrimSpace(token))
	}
	out, err := auditreceipt.RunScrubbedGit(projectRoot, "diff", "--name-only", fromSHA+".."+toSHA)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		p := strings.TrimSpace(line)
		if p == "" {
			continue
		}
		if p == "progress.md" || strings.HasSuffix(p, "/progress.md") || strings.HasPrefix(p, ".moai/reports/") {
			continue
		}
		tokens, anchored := anchorTokens[p]
		if !anchored {
			return false
		}
		hunks, herr := diffHunkBodies(projectRoot, fromSHA, toSHA, p)
		if herr != nil || len(hunks) == 0 {
			return false
		}
		for _, hunk := range hunks {
			matched := false
			for _, tok := range tokens {
				if tok != "" && strings.Contains(hunk, tok) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
	}
	return true
}

// diffHunkBodies returns the text of each diff hunk for one file between two
// SHAs. The @@ hunk header line is EXCLUDED from the body: git appends the
// nearest enclosing funcname line after the @@ range, and that heuristic can
// carry an anchor-section line from far outside the hunk (observed: a tail-
// section edit whose @@ header read "REQ-ACE-001 detail") — only the hunk's
// own context and changed lines decide (card-review F3). A git failure is an
// error — the caller treats it as fail-closed.
func diffHunkBodies(projectRoot, fromSHA, toSHA, path string) ([]string, error) {
	out, err := auditreceipt.RunScrubbedGit(projectRoot, "diff", "-U3", fromSHA+".."+toSHA, "--", path)
	if err != nil {
		return nil, err
	}
	var hunks []string
	cur := ""
	inHunk := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "@@") {
			if inHunk && cur != "" {
				hunks = append(hunks, cur)
			}
			cur = ""      // the @@ header and its trailing funcname context do not count
			inHunk = true // the diff preamble ("--- a/...") must not become a hunk
			continue
		}
		if inHunk {
			cur += line + "\n"
		}
	}
	if inHunk && cur != "" {
		hunks = append(hunks, cur)
	}
	return hunks, nil
}

// reqACSetsUnchanged reports whether the REQ/AC id sets of the SPEC's
// definition files — spec.md AND acceptance.md — are identical at both
// audited SHAs (REQ-ACR-010): an identifier change confined to acceptance.md
// refuses the delta exactly as a spec.md change does. A git failure on
// spec.md is fail-closed; acceptance.md absent at BOTH ends (a Tier S SPEC)
// reads as an unchanged empty set, while absent at one end only fails
// closed.
func reqACSetsUnchanged(projectRoot, specID, fromSHA, toSHA string) bool {
	specRel := filepath.ToSlash(filepath.Join(".moai", "specs", specID, "spec.md"))
	from, err1 := auditreceipt.RunScrubbedGit(projectRoot, "show", fromSHA+":"+specRel)
	to, err2 := auditreceipt.RunScrubbedGit(projectRoot, "show", toSHA+":"+specRel)
	if err1 != nil || err2 != nil {
		return false
	}
	if !equalStringSets(reqACSet(from), reqACSet(to)) {
		return false
	}
	accRel := filepath.ToSlash(filepath.Join(".moai", "specs", specID, "acceptance.md"))
	fromAcc, errA := auditreceipt.RunScrubbedGit(projectRoot, "show", fromSHA+":"+accRel)
	toAcc, errB := auditreceipt.RunScrubbedGit(projectRoot, "show", toSHA+":"+accRel)
	if errA == nil && errB == nil {
		return equalStringSets(reqACSet(fromAcc), reqACSet(toAcc))
	}
	// At least one end failed to read. Only the CLEAN absence shape may
	// read as "unchanged empty at both ends" — a corrupt or unreadable
	// object also fails git-show, and equating it with absence admitted the
	// delta fail-open (sync-audit-2 F5). git ls-tree separates the shapes
	// without matching git's error prose: exit 0 with output when the path
	// is in the tree, exit 0 with empty output when it is not, non-zero
	// when the object cannot be read at all — only the middle shape is
	// absence; anything else fails closed.
	absentA, errLA := gitPathAbsent(projectRoot, fromSHA, accRel)
	absentB, errLB := gitPathAbsent(projectRoot, toSHA, accRel)
	if errLA != nil || errLB != nil {
		return false // absence undecidable — fail closed
	}
	if absentA && absentB {
		return true // no acceptance.md at either end — nothing to compare
	}
	return false // appeared, disappeared, or an unreadable object
}

// gitPathAbsent reports whether rel is absent from sha's tree — and only
// that. A non-zero ls-tree (corrupt or unreadable object) is an error, not
// absence: the caller fails closed on it.
func gitPathAbsent(projectRoot, sha, rel string) (bool, error) {
	out, err := auditreceipt.RunScrubbedGit(projectRoot, "ls-tree", sha, "--", rel)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

func reqACSet(raw string) map[string]bool {
	set := map[string]bool{}
	for _, id := range reqACID.FindAllString(raw, -1) {
		set[id] = true
	}
	return set
}

func equalStringSets(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func readEvidenceRaw(path string) []byte {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// persistOutcome records one ceiling outcome to the SPEC's progress.md §G
// Override and Refusal Record (the durable carrier) and the machine-local
// audit trail (REQ-ACE-007/012). A non-empty debt inventory rides both
// records as the additive `debts=` field (REQ-ACR-004). Persistence is
// best-effort: a record-write or serialization failure warns on stderr but
// never flips the admission decision the outcome already decided
// (REQ-ACR-007).
func persistOutcome(in VerdictCeilingInput, kind string, oc *VerdictCeilingOutcome) {
	line := fmt.Sprintf("- %s %s %s outcome=%s reasons=%q evidence=%s",
		time.Now().UTC().Format(time.RFC3339), in.SpecID, kind, oc.Outcome,
		strings.Join(oc.Reasons, "; "), strings.Join(oc.Evidence, ","))
	debtsField := debtsRecordField(oc.Debts)
	if debtsField != "" {
		line += " " + debtsField
	}
	if err := appendProgressRecord(in.SpecDir, line); err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: progress.md §G record: %v\n", err)
	}
	if err := appendAuditTrail(in, kind, oc.Outcome, strings.Join(oc.Reasons, "; "), debtsField); err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: audit trail: %v\n", err)
	}
}

// debtRecord is the persisted form of one admitted debt in the debts= field
// (decision-index Q2): a JSON object with snake_case keys — the verdict's
// own Debt struct stays untouched.
type debtRecord struct {
	ID          string `json:"id"`
	DisposeIn   string `json:"dispose_in"`
	Description string `json:"description"`
}

// debtsRecordField serializes the admitted debt inventory as the record's
// additive `debts=` field (decision-index Q2): one JSON array of
// {id, dispose_in, description} objects — a single line by construction and
// fully escaped by the stdlib encoder, so quotes, separators, and newlines
// in a debt's free text never inject record lines and every field stays
// recoverable (REQ-ACR-004/005). Empty when the inventory is empty — no
// debt-free record changes shape (REQ-ACR-006). A serialization failure
// degrades to the best-effort warning posture, never to an admission change
// (REQ-ACR-007).
func debtsRecordField(debts []auditverdict.Debt) string {
	if len(debts) == 0 {
		return ""
	}
	records := make([]debtRecord, len(debts))
	for i, d := range debts {
		records[i] = debtRecord{ID: d.ID, DisposeIn: d.DisposeIn, Description: d.Description}
	}
	b, err := json.Marshal(records)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: debt inventory serialization: %v\n", err)
		return ""
	}
	return "debts=" + string(b)
}

// progressSectionHeading is the §G section the engine and the override path
// append records under — outside the plan-artifact hash subject set (D10).
const progressSectionHeading = "## §G Override and Refusal Record"

// progressRecordMu serializes the §G read-modify-write in-process, so
// concurrent persistOutcome calls on one SPEC each read progress.md after
// the previous write lands: every record survives exactly once and no
// pre-existing content is overwritten (REQ-ACR-008 — the measured defect
// was 24 parallel record calls surviving as 7 with the heading lost).
var progressRecordMu sync.Mutex

// appendProgressRecord appends one record line at the END of the §G block
// of the SPEC's progress.md — immediately before the next same-level
// heading when a section follows §G, end-of-file only when §G is last —
// creating the file and the heading when absent (REQ-ACR-008/AC-ACR-016).
// The read-modify-write is serialized in-process and written through an
// atomic mode-preserving same-directory replace, so concurrent records
// neither lose each other nor destroy pre-existing content and a reader
// never observes a partial write.
func appendProgressRecord(specDir, line string) error {
	progressRecordMu.Lock()
	defer progressRecordMu.Unlock()
	path := filepath.Join(specDir, "progress.md")
	// A progress.md that is itself a symlink (a dotfiles-managed file, for
	// example) must stay one: the atomic replace swaps the directory entry,
	// so replacing `path` would convert the link into a regular file and
	// materialize the target's content inside the tracked SPEC directory
	// (sync-audit-2 F4). The pre-repair os.WriteFile wrote THROUGH the
	// link — restore that posture by resolving the link once and reading +
	// replacing the resolved target; the link itself is never replaced. A
	// dangling link cannot be written through and fails closed — a
	// best-effort warning upstream, never an admission change.
	if info, lerr := os.Lstat(path); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
		resolved, rerr := filepath.EvalSymlinks(path)
		if rerr != nil {
			// A dangling link — the target does not exist yet. Resolve the
			// link's parent and append the target name, so the record
			// writes through to the target the next append will find,
			// exactly as the pre-repair os.WriteFile followed the link and
			// created it (round-4 edge 4); the link survives. A parent that
			// itself does not exist fails closed, the same ENOENT the
			// in-place write would have raised.
			ref, rlerr := os.Readlink(path)
			if rlerr != nil {
				return rlerr
			}
			if !filepath.IsAbs(ref) {
				ref = filepath.Join(filepath.Dir(path), ref)
			}
			parent, perr := filepath.EvalSymlinks(filepath.Dir(ref))
			if perr != nil {
				return perr
			}
			resolved = filepath.Join(parent, filepath.Base(ref))
		}
		path = resolved
	}
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := progressWithRecord(string(raw), line)
	// Mode posture (sync-audit-1 F2): an existing progress.md keeps its
	// pre-write permission bits — the atomic replace preserves the
	// destination's mode. A brand-new progress.md takes the mode the
	// pre-repair os.WriteFile gave it, 0644 with the process umask applied,
	// so a not-yet-existing file is pre-created empty through os.WriteFile
	// itself and the kernel applies the umask at create time; handing 0644
	// verbatim to an atomic writer's explicit chmod would bypass it.
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			return err
		}
	}
	// Write-denial check (sync-audit-4 F8): the replace must not bypass the
	// original's write restriction — the pre-repair os.WriteFile failed
	// with permission denied on a read-only progress.md, while a rename
	// never opens it. Open the original for writing first; a denial is a
	// clean error with the file untouched and no temp created.
	if f, oerr := os.OpenFile(path, os.O_WRONLY, 0); oerr != nil {
		return oerr
	} else {
		_ = f.Close()
	}
	// Atomic same-directory replace, seeded with the original's file
	// metadata (round-3 repair 3): rename(2) swaps the directory entry, so
	// the replacement carries the TEMP file's access-control entries — a
	// bare rename drops an ACL the pre-repair os.WriteFile preserved. The
	// seeder copies the original's metadata onto the temp before the new
	// content is written. A seeding failure ABORTS the replace — the temp
	// is removed and the original untouched; there is NO mode-only
	// fallback, which dropped the ACL and silently rewrote a
	// write-restricted file (the two F8 faces). The mode is applied
	// explicitly either way (the F2 posture: existing-file mode preserved,
	// new-file umask-adjusted 0644 — the pre-create above already put the
	// umask into the original's mode).
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".progress-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds
	if serr := seedFileMetadataFn(tmpName, path); serr != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("seed progress.md metadata: %w", serr)
	}
	if werr := os.WriteFile(tmpName, []byte(content), 0); werr != nil {
		return werr
	}
	mode := os.FileMode(0o644)
	if info, serr := os.Stat(path); serr == nil {
		mode = info.Mode().Perm()
	}
	if cerr := os.Chmod(tmpName, mode); cerr != nil {
		return cerr
	}
	return os.Rename(tmpName, path)
}

// seedFileMetadataFn is the metadata-seeding seam (a package var so the
// abort contract is testable by injection); it points at the platform
// implementation below.
var seedFileMetadataFn = seedFileMetadata

// seedFileMetadata copies the original's mode, access-control entries, and
// extended attributes onto the temp file via cp -p — the preservation a
// bare rename cannot provide (round-3 repair 3). A failure is an ERROR and
// the caller aborts the replace: there is no mode-only fallback, which
// dropped the ACL and bypassed a write restriction (sync-audit-4 F8). On
// Windows there is no cp and no explicit-file ACL axis to copy — the seeder
// is a documented no-op success (the temp file inherits the directory's ACL
// at creation); flagged for leader review.
func seedFileMetadata(tmp, original string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if out, err := exec.Command("cp", "-p", original, tmp).CombinedOutput(); err != nil {
		return fmt.Errorf("cp -p %s %s: %v (%s)", original, tmp, err, out)
	}
	return nil
}

// progressWithRecord returns content with one record line inserted at the
// end of the §G block. The heading-absent and §G-last shapes are
// byte-identical to the pre-repair append-at-EOF behavior; only the
// §G-followed-by-a-section shape changed, from absorbing the record into
// the later section to ending the §G block with it.
func progressWithRecord(content, line string) string {
	if content == "" {
		return "\n" + progressSectionHeading + "\n\n" + line + "\n"
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	heading := -1
	// The §G heading itself is looked up with the same fence state the
	// next-heading scan applies: a §G-heading-prefixed line inside an open
	// fenced code block (an example, a quoted template) is code, not the
	// section (round-3 repair 1).
	inFence := false
	var fenceChar byte
	fenceLen := 0
	for i, l := range lines {
		if inFence {
			if closesFence(l, fenceChar, fenceLen) {
				inFence = false
			}
			continue
		}
		if c, n, opened := opensFence(l); opened {
			inFence = true
			fenceChar, fenceLen = c, n
			continue
		}
		if strings.HasPrefix(l, progressSectionHeading) {
			heading = i
			break
		}
	}
	var out []string
	if heading < 0 {
		out = append(lines, "", progressSectionHeading, "", line)
	} else {
		at := len(lines) // §G is the last section: append at end-of-file
		// A `## `-prefixed line inside an open fenced code block is code,
		// not a section boundary — the scan tracks fence state so the
		// record lands at the §G block's end, after the fence closes
		// (round-2 leader-ruled fold).
		inFence := false
		var fenceChar byte
		fenceLen := 0
		for i := heading + 1; i < len(lines); i++ {
			l := lines[i]
			if inFence {
				if closesFence(l, fenceChar, fenceLen) {
					inFence = false
				}
				continue
			}
			if c, n, opened := opensFence(l); opened {
				inFence = true
				fenceChar, fenceLen = c, n
				continue
			}
			if strings.HasPrefix(l, "## ") {
				at = i // a section follows §G: the record ends the §G block
				break
			}
		}
		out = append(lines[:at:at], line)
		out = append(out, lines[at:]...)
	}
	return strings.Join(out, "\n") + "\n"
}

// opensFence reports the fence a line OPENS: a line whose leading run is at
// least three backticks or three tildes, indented 0-3 columns (round-4
// edge 5 — Markdown's fenced code blocks; a deeper indent makes the line an
// INDENTED CODE BLOCK, not a fence; the fence character and its run length
// decide which line can close it).
func opensFence(line string) (c byte, n int, ok bool) {
	indent, trimmed := fenceIndent(line)
	if indent > 3 || trimmed == "" {
		return 0, 0, false
	}
	c = trimmed[0]
	if c != '`' && c != '~' {
		return 0, 0, false
	}
	for n < len(trimmed) && trimmed[n] == c {
		n++
	}
	return c, n, n >= 3
}

// closesFence reports whether line closes a fence opened with n of the
// fence character c: indented 0-3 columns (the same Markdown rule), at
// least n of that character, then nothing but whitespace.
func closesFence(line string, c byte, n int) bool {
	indent, trimmed := fenceIndent(line)
	if indent > 3 {
		return false
	}
	i := 0
	for i < len(trimmed) && trimmed[i] == c {
		i++
	}
	return i >= n && strings.TrimSpace(trimmed[i:]) == ""
}

// fenceIndent counts the line's leading indentation in columns (a space
// counts one, a tab advances to the next multiple-of-four column — the
// CommonMark rule) and returns the line without it.
func fenceIndent(line string) (int, string) {
	col := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			col++
		case '\t':
			col = (col/4 + 1) * 4
		default:
			return col, line[i:]
		}
	}
	return col, ""
}

// RecordRequiredBackendRefusal persists a required-backend refusal
// (REQ-ACE-009/010) to the SPEC's progress.md §G record and the audit trail
// — the durable carrier REQ-ACE-007/012 requires for EVERY required-backend
// refusal, including the below-ceiling ones the ceiling ladder never sees
// (card-review F7). Recording is best-effort (the same stderr-warning
// posture persistOutcome takes): a record-write failure never changes the
// refusal the seam already returned.
func RecordRequiredBackendRefusal(in VerdictCeilingInput, reason string) {
	line := fmt.Sprintf("- %s %s required-backend-refusal outcome=refused reasons=%q",
		time.Now().UTC().Format(time.RFC3339), in.SpecID, reason)
	if err := appendProgressRecord(in.SpecDir, line); err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: progress.md §G record: %v\n", err)
	}
	if err := appendAuditTrail(in, "required-backend-refusal", "refused", reason); err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: audit trail: %v\n", err)
	}
}

// appendAuditTrail appends one machine-local trail line to
// <root>/.moai/state/audit-enforcement.log (design.md §5). Each non-empty
// extra value appends as an additive `key=value` trail field — the debt
// inventory rides it (REQ-ACR-004) — and absent extras leave the line
// byte-identical to the pre-repair grammar (REQ-ACR-006).
func appendAuditTrail(in VerdictCeilingInput, kind, outcome, reason string, extra ...string) error {
	dir := filepath.Join(in.ProjectRoot, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	line := fmt.Sprintf("%s spec=%s kind=%s outcome=%s reason=%q",
		time.Now().UTC().Format(time.RFC3339), in.SpecID, kind, outcome, reason)
	for _, e := range extra {
		if e != "" {
			line += " " + e
		}
	}
	line += "\n"
	f, err := os.OpenFile(filepath.Join(dir, "audit-enforcement.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(line)
	return err
}

// AcknowledgeRequiredBackend records an operator override of a required-
// backend refusal (REQ-ACE-011): the backend and a mandatory acknowledgement
// note, written to the SPEC's progress.md §G record and the audit trail.
// An empty note is refused (acceptance.md §C edge 8); there is no
// silent-skip path.
//
// @MX:NOTE: [AUTO] operator-override record path for a required-backend refusal — the acknowledgement note is mandatory, there is no silent-skip form
// @MX:SPEC: SPEC-AUDIT-CEILING-001
func AcknowledgeRequiredBackend(in VerdictCeilingInput, backend, note string) error {
	if strings.TrimSpace(note) == "" {
		return fmt.Errorf("override of required backend %s refused: the acknowledgement note is mandatory", backend)
	}
	if strings.TrimSpace(backend) == "" {
		return fmt.Errorf("override refused: the backend name is mandatory")
	}
	line := fmt.Sprintf("- %s %s override backend=%s note=%q",
		time.Now().UTC().Format(time.RFC3339), in.SpecID, backend, note)
	if err := appendProgressRecord(in.SpecDir, line); err != nil {
		return fmt.Errorf("progress.md §G override record: %w", err)
	}
	if err := appendAuditTrail(in, "override", "ack", fmt.Sprintf("backend=%s note=%q", backend, note)); err != nil {
		return fmt.Errorf("audit trail: %w", err)
	}
	return nil
}

// planAuditRoundFile is one recorded plan-audit file: plan-audit.md ranks as
// iteration 0 and each plan-audit-iter<N>.md carries its parsed N.
type planAuditRoundFile struct {
	path string
	n    int
}

// iterFileName matches the plan-audit-iter<N>.md family; the captured suffix
// must parse as a positive integer. The suffix may be empty (an
// iter-with-no-number file still carries the family shape, so it fails the
// count closed rather than skipping silently).
var iterFileName = regexp.MustCompile(`^plan-audit-iter(.*)\.md$`)

// evidenceDirInput is one deduplicated evidence-directory input carrying its
// directory class.
type evidenceDirInput struct {
	path   string
	listed bool
}

// canonicalEvidenceDir resolves one evidence-directory spelling to the
// absolute, symlink-resolved form the dedupe keys on (card-review r1
// CR-P2-2): identical directories compare equal regardless of spelling, so a
// relative `.moai/reports/<SPEC-ID>` listing and the auto-included absolute
// form are one directory, never two. A path that is already absolute passes
// through Abs untouched (Abs never joins the cwd onto an absolute path);
// EvalSymlinks resolves macOS /var-style links; a non-existent path keeps its
// Abs form (the absent-listed error still fires on the read, naming the
// resolved path).
func canonicalEvidenceDir(p string) string {
	if p == "" {
		return ""
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	if resolved, linkErr := filepath.EvalSymlinks(abs); linkErr == nil {
		return resolved
	}
	return abs
}

// evidenceDirInputs builds the deduplicated directory inputs over canonical
// spellings: listed directories register first so their error-on-absent class
// wins a dedupe collision against the SPEC-scoped class (fail-closed
// direction).
func evidenceDirInputs(specDir string, listedDirs []string) []evidenceDirInput {
	var inputs []evidenceDirInput
	seen := map[string]bool{}
	for _, d := range listedDirs {
		if d == "" {
			continue
		}
		clean := canonicalEvidenceDir(d)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		inputs = append(inputs, evidenceDirInput{path: clean, listed: true})
	}
	if specDir != "" {
		clean := canonicalEvidenceDir(specDir)
		if !seen[clean] {
			seen[clean] = true
			inputs = append(inputs, evidenceDirInput{path: clean, listed: false})
		}
	}
	return inputs
}

// evidenceDirList names the deduplicated evidence directories an evaluation
// covered — the record's evidence paths (REQ-ACR-003).
func evidenceDirList(specDir string, listedDirs []string) []string {
	inputs := evidenceDirInputs(specDir, listedDirs)
	paths := make([]string, 0, len(inputs))
	for _, in := range inputs {
		paths = append(paths, in.path)
	}
	return paths
}

// scanEvidenceDirs is the shared scanner behind CountPlanAuditRounds and
// SelectLatestVerdict: it walks the SPEC-scoped directory and the explicitly
// listed directories and returns the plan-audit round files each holds.
//
// Directory-class rules (REQ-ACR-001/013): an absent SPEC-scoped directory
// contributes 0 (an unaudited SPEC is not an error); an explicitly LISTED
// directory that does not exist is an error naming the missing path — a typo
// must never silently count 0. Identical listed paths are deduplicated before
// counting (R6).
func scanEvidenceDirs(specDir string, listedDirs []string) ([]planAuditRoundFile, error) {
	var out []planAuditRoundFile
	for _, in := range evidenceDirInputs(specDir, listedDirs) {
		entries, err := os.ReadDir(in.path)
		if err != nil {
			if in.listed || !os.IsNotExist(err) {
				return nil, fmt.Errorf("read evidence directory %q: %w", in.path, err)
			}
			continue // SPEC-scoped and absent → contributes 0
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == "plan-audit.md" {
				out = append(out, planAuditRoundFile{path: filepath.Join(in.path, name), n: 0})
				continue
			}
			m := iterFileName.FindStringSubmatch(name)
			if m == nil {
				continue // not in the plan-audit round family
			}
			n, perr := strconv.Atoi(m[1])
			if perr != nil || n <= 0 {
				return nil, fmt.Errorf("round file %q has an iteration suffix that does not parse as a positive integer", filepath.Join(in.path, name))
			}
			out = append(out, planAuditRoundFile{path: filepath.Join(in.path, name), n: n})
		}
	}
	return out, nil
}

// CountPlanAuditRounds computes a SPEC's plan-audit round count from durable
// iteration evidence on disk, one recorded file one round: plan-audit.md
// contributes one round and each plan-audit-iter<N>.md one round. specDir is
// the SPEC-scoped evidence directory (.moai/reports/<SPEC-ID>); listedDirs are
// the evidence directories explicitly listed in the invocation. The count is
// never derived from in-process memory or session state.
func CountPlanAuditRounds(specDir string, listedDirs []string) (int, error) {
	files, err := scanEvidenceDirs(specDir, listedDirs)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// SelectLatestVerdict returns the recorded file with the highest parsed
// iteration number — plan-audit.md ranking as iteration 0 — the verdict the
// disposition branches read. The same highest number appearing in two evidence
// directories is a selection error rather than a silent pick; no evidence at
// all selects nothing (empty path, no error).
func SelectLatestVerdict(specDir string, listedDirs []string) (string, error) {
	files, err := scanEvidenceDirs(specDir, listedDirs)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", nil
	}
	sort.Slice(files, func(i, j int) bool { return files[i].n < files[j].n })
	best := files[len(files)-1]
	var tied []string
	for _, f := range files {
		if f.n == best.n {
			tied = append(tied, f.path)
		}
	}
	if len(tied) > 1 {
		return "", fmt.Errorf("latest plan-audit verdict is ambiguous: iteration %d recorded in %d evidence files (%q and %q); list one evidence directory to disambiguate",
			best.n, len(tied), tied[0], tied[1])
	}
	return best.path, nil
}

// ResolvePlanAuditCeiling resolves the SPEC's ceiling from the configured
// plan_audit_tier_ceilings map. The tier resolves under the shared predicate's
// rule (auditverdict.SpecTier): an absent or unknown tier resolves to L. A
// resolved ceiling that is missing from the map or non-positive is a
// configuration error rather than a ceiling of zero (REQ-ACR-002).
func ResolvePlanAuditCeiling(tier string, ceilings map[string]int) (int, error) {
	resolved := tier
	if resolved != "S" && resolved != "M" && resolved != "L" {
		resolved = "L" // auditverdict.SpecTier's rule: absent or unknown → L
	}
	ceiling, ok := ceilings[resolved]
	if !ok || ceiling <= 0 {
		return 0, fmt.Errorf("plan_audit_tier_ceilings carries no positive ceiling for tier %q (resolved to %q); a mis-configured ceilings map must never resolve to a ceiling of zero", tier, resolved)
	}
	return ceiling, nil
}

// Ceiling disposition values — exactly the three operator-named tokens of
// REQ-ACR-003/D2.
const (
	CeilingDispositionDebtProceed = "debt-proceed"
	CeilingDispositionSplit       = "split"
	CeilingDispositionHold        = "hold"
)

// AuditCeilingStateDir is the machine-local directory the outcome records
// live under, relative to the project root (gitignored parent — the record is
// machine-local state, not a cross-machine durable carrier; SPEC
// §F F2).
const AuditCeilingStateDir = ".moai/state/audit-ceiling"

// CeilingOutcome is one ceiling policy outcome record. Every record carries
// the round count, the ceiling, the latest verdict label, and the evidence
// paths (REQ-ACR-003).
type CeilingOutcome struct {
	SpecID           string   `json:"spec_id"`
	Disposition      string   `json:"disposition"`
	Count            int      `json:"count"`
	Ceiling          int      `json:"ceiling"`
	VerdictLabel     string   `json:"verdict_label"`
	VerdictAdmitted  bool     `json:"verdict_admitted"`
	DebtIDs          []string `json:"debt_ids,omitempty"`
	EvidencePaths    []string `json:"evidence_paths"`
	SplitProposalRef string   `json:"split_proposal_ref,omitempty"`
}

// CeilingInput carries one ceiling evaluation's inputs.
type CeilingInput struct {
	// SpecID is the SPEC whose ceiling is evaluated (names the record file).
	SpecID string
	// SpecEvidenceDir is the SPEC-scoped evidence directory
	// (.moai/reports/<SPEC-ID>); absent contributes 0.
	SpecEvidenceDir string
	// ListedDirs are the evidence directories explicitly listed in the
	// invocation; an absent listing is an error.
	ListedDirs []string
	// Tier is the SPEC's tier; an absent or unknown value resolves to L.
	Tier string
	// Threshold is the plan PASS threshold the admission check uses (the
	// caller resolves it from the SPEC's tier).
	Threshold float64
	// Ceilings is the configured plan_audit_tier_ceilings map.
	Ceilings map[string]int
	// OnFinalHit is the configured plan_audit_ceiling_policy.on_final_hit
	// value; any other or unreadable value fails closed to hold.
	OnFinalHit string
}

// EvaluatePlanAuditCeiling is the single evaluation entry composing select →
// count → ceiling → disposition. The ceiling resolution error propagates
// BEFORE any count/ceiling comparison, and nothing is recorded here — the
// caller records the returned outcome through RecordCeilingOutcome, the one
// recording path (REQ-ACR-003). The returned bool reports whether the ceiling
// applies: a clean admitted PASS at the ceiling is not a ceiling outcome at
// all (REQ-ACR-004 — the ceiling caps repetition, not a healthy result), and
// rounds below the ceiling apply nothing.
func EvaluatePlanAuditCeiling(in CeilingInput) (CeilingOutcome, bool, error) {
	latest, err := SelectLatestVerdict(in.SpecEvidenceDir, in.ListedDirs)
	if err != nil {
		return CeilingOutcome{}, false, err
	}
	count, err := CountPlanAuditRounds(in.SpecEvidenceDir, in.ListedDirs)
	if err != nil {
		return CeilingOutcome{}, false, err
	}
	// The configuration error propagates before any comparison is made (R2's
	// Evaluate-level arm): a mis-configured ceilings map never fires the path.
	ceiling, err := ResolvePlanAuditCeiling(in.Tier, in.Ceilings)
	if err != nil {
		return CeilingOutcome{}, false, err
	}
	if count < ceiling {
		return CeilingOutcome{}, false, nil
	}

	// The hashOK argument is true here by design: the plan-artifact hash
	// binding belongs to the run-entry seams (the kickoff evaluator and the
	// card-transition guard) that run the shared predicate before any ceiling
	// evaluation; this path judges repetition against the verdict's recorded
	// fields, not the tree's current hash state.
	fields := auditverdict.Fields{}
	if latest != "" {
		raw, readErr := os.ReadFile(latest)
		if readErr != nil {
			return CeilingOutcome{}, false, fmt.Errorf("read latest verdict %q: %w", latest, readErr)
		}
		fields = auditverdict.Parse(raw)
	}
	admitted, _ := auditverdict.Admit(fields, auditverdict.PhasePlan, in.Threshold, true)

	if admitted && fields.Label == auditverdict.LabelPass {
		return CeilingOutcome{}, false, nil
	}

	out := CeilingOutcome{
		SpecID:          in.SpecID,
		Count:           count,
		Ceiling:         ceiling,
		VerdictLabel:    fields.Label,
		VerdictAdmitted: admitted,
		EvidencePaths:   evidenceDirList(in.SpecEvidenceDir, in.ListedDirs),
	}
	switch {
	case admitted:
		// The only admitted label left is PASS-WITH-DEBT: record debt-proceed
		// referencing the verdict's debt ids (REQ-ACR-003 arm 1).
		out.Disposition = CeilingDispositionDebtProceed
		for _, d := range fields.Debts {
			out.DebtIDs = append(out.DebtIDs, d.ID)
		}
	case in.OnFinalHit == config.PlanAuditCeilingOnFinalSplit:
		out.Disposition = CeilingDispositionSplit
	case in.OnFinalHit == config.PlanAuditCeilingOnFinalHoldAndSplit:
		// One record carrying both halves of the shipped value: disposition
		// hold plus the split-proposal reference (REQ-ACR-003 arm 2 / D2).
		out.Disposition = CeilingDispositionHold
		out.SplitProposalRef = splitProposalReference(in.SpecID)
	default:
		// Any other or unreadable policy value records hold (fail-closed;
		// REQ-ACR-003 arm 4).
		out.Disposition = CeilingDispositionHold
	}
	return out, true, nil
}

// splitProposalReference is the deterministic split-proposal reference a
// hold-and-split record carries: the SPEC must be split before further plan
// audits, and the reference names what the split applies to.
func splitProposalReference(specID string) string {
	return "split required by plan_audit_ceiling_policy.on_final_hit=hold-and-split: propose splitting " + specID
}

// RecordCeilingOutcome writes the outcome record to
// <projectRoot>/.moai/state/audit-ceiling/<SPEC-ID>.json (machine-local
// state). The record lands under the project whose ceiling was judged — the
// caller passes its root, so a CLI run from a different cwd records where the
// SPEC lives, not where the command ran (card-review r1 CR-P2-1). It is the
// ONE recording path: no other code writes a ceiling outcome record.
//
// @MX:ANCHOR: [AUTO] the ONE ceiling-outcome recording path — every recorded plan-audit ceiling outcome is written by this function, never elsewhere
// @MX:REASON: a second writer would fork the record format and let two audit-ceiling truth sources drift silently
// @MX:SPEC: SPEC-AUDIT-CEILING-002
func RecordCeilingOutcome(projectRoot, specID string, outcome CeilingOutcome) error {
	if specID == "" || specID == "." || specID == ".." || strings.ContainsAny(specID, `/\`) {
		return fmt.Errorf("RecordCeilingOutcome: invalid SPEC id %q", specID)
	}
	dir := filepath.Join(projectRoot, AuditCeilingStateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("RecordCeilingOutcome: create %s: %w", dir, err)
	}
	path := filepath.Join(dir, specID+".json")
	data, err := json.MarshalIndent(outcome, "", "  ")
	if err != nil {
		return fmt.Errorf("RecordCeilingOutcome: marshal: %w", err)
	}
	if err := atomicWrite(path, append(data, '\n')); err != nil {
		return fmt.Errorf("RecordCeilingOutcome: write %s: %w", path, err)
	}
	return nil
}
