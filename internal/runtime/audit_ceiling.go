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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

// CeilingOutcome is the structured record of one ceiling evaluation. Refusal
// is observable via Blocked; a pass-through carries Blocked == false and
// records that the ceiling was reached (REQ-ACE-013).
type CeilingOutcome struct {
	Outcome  string              `json:"outcome"`
	Reasons  []string            `json:"reasons"`
	Evidence []string            `json:"evidence"`
	Blocked  bool                `json:"blocked"`
	Debts    []auditverdict.Debt `json:"debts,omitempty"`
}

// CeilingInput names the SPEC a ceiling evaluation runs for. SpecDir is the
// SPEC directory (tier, threshold, and the progress.md §G record target);
// ProjectRoot is the tree whose harness config and report directories the
// engine reads.
type CeilingInput struct {
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
func EvaluateCeiling(in CeilingInput, fields auditverdict.Fields, hashOK bool, required []string) (*CeilingOutcome, bool, error) {
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
	if ok, _ := auditverdict.Admit(fields, auditverdict.PhasePlan, threshold, hashOK, required); ok {
		oc := &CeilingOutcome{
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
		if ok, _ := auditverdict.Admit(converted, auditverdict.PhasePlan, threshold, hashOK, required); ok {
			oc := &CeilingOutcome{
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
		oc := &CeilingOutcome{
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
	oc := &CeilingOutcome{
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
// a config that exists but cannot be read, parsed, or validated — an invalid
// on_final_hit included — is an ERROR the seam refuses on, never a silent
// fallback to the defaults (card-review F1: a policy name the CLI cannot
// enforce must not be swallowed).
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
		if cfg.PlanAuditTierCeilings.S > 0 {
			tierCeiling = cfg.PlanAuditTierCeilings.S
		}
	case "M":
		if cfg.PlanAuditTierCeilings.M > 0 {
			tierCeiling = cfg.PlanAuditTierCeilings.M
		}
	default:
		if cfg.PlanAuditTierCeilings.L > 0 {
			tierCeiling = cfg.PlanAuditTierCeilings.L
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
// largest iteration strictly below the latest, deduped across both families
// and including the legacy stream (card-review F4), "" when no prior round
// exists.
func previousAuditedSHA(in CeilingInput, ev RoundEvidence) string {
	if ev.LatestPath == "" {
		return ""
	}
	latestN := iterationOf(filepath.Base(ev.LatestPath), in.SpecID)
	if latestN <= 0 {
		return ""
	}
	legacyFile := regexp.MustCompile(`^` + regexp.QuoteMeta(in.SpecID) + `-review-([0-9]+)\.md$`)
	best := -1
	prev := ""
	for _, src := range ev.Sources {
		if src == ev.LatestPath {
			continue
		}
		name := filepath.Base(src)
		n := iterationOf(name, in.SpecID)
		if n <= 0 {
			if m := legacyFile.FindStringSubmatch(name); m != nil {
				if v, perr := strconv.Atoi(m[1]); perr == nil {
					n = v
				}
			}
		}
		if n <= 0 || n >= latestN {
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

// iterationOf returns the iteration number a convention-family file name
// carries (1 for the bare plan-audit.md shape), 0 when the name is not a
// convention shape.
func iterationOf(name, _ string) int {
	m := conventionFile.FindStringSubmatch(name)
	if m == nil {
		return 0
	}
	if m[1] == "" {
		return 1
	}
	if v, err := strconv.Atoi(m[1]); err == nil {
		return v
	}
	return 0
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

// reqACSetsUnchanged reports whether the REQ/AC id set of the SPEC's spec.md
// is identical at both audited SHAs. A git failure is fail-closed.
func reqACSetsUnchanged(projectRoot, specID, fromSHA, toSHA string) bool {
	specRel := filepath.ToSlash(filepath.Join(".moai", "specs", specID, "spec.md"))
	from, err1 := auditreceipt.RunScrubbedGit(projectRoot, "show", fromSHA+":"+specRel)
	to, err2 := auditreceipt.RunScrubbedGit(projectRoot, "show", toSHA+":"+specRel)
	if err1 != nil || err2 != nil {
		return false
	}
	return equalStringSets(reqACSet(from), reqACSet(to))
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
// audit trail (REQ-ACE-007/012). Persistence is best-effort: a record-write
// failure warns on stderr but never flips the admission decision the outcome
// already decided.
func persistOutcome(in CeilingInput, kind string, oc *CeilingOutcome) {
	line := fmt.Sprintf("- %s %s %s outcome=%s reasons=%q evidence=%s",
		time.Now().UTC().Format(time.RFC3339), in.SpecID, kind, oc.Outcome,
		strings.Join(oc.Reasons, "; "), strings.Join(oc.Evidence, ","))
	if err := appendProgressRecord(in.SpecDir, line); err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: progress.md §G record: %v\n", err)
	}
	if err := appendAuditTrail(in, kind, oc.Outcome, strings.Join(oc.Reasons, "; ")); err != nil {
		fmt.Fprintf(os.Stderr, "[audit-ceiling] warning: audit trail: %v\n", err)
	}
}

// progressSectionHeading is the §G section the engine and the override path
// append records under — outside the plan-artifact hash subject set (D10).
const progressSectionHeading = "## §G Override and Refusal Record"

// appendProgressRecord appends one record line under the §G heading of the
// SPEC's progress.md, creating the file and the heading when absent.
func appendProgressRecord(specDir, line string) error {
	path := filepath.Join(specDir, "progress.md")
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(raw)
	if !strings.Contains(content, progressSectionHeading) {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + progressSectionHeading + "\n\n"
	} else if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += line + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

// RecordRequiredBackendRefusal persists a required-backend refusal
// (REQ-ACE-009/010) to the SPEC's progress.md §G record and the audit trail
// — the durable carrier REQ-ACE-007/012 requires for EVERY required-backend
// refusal, including the below-ceiling ones the ceiling ladder never sees
// (card-review F7). Recording is best-effort (the same stderr-warning
// posture persistOutcome takes): a record-write failure never changes the
// refusal the seam already returned.
func RecordRequiredBackendRefusal(in CeilingInput, reason string) {
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
// <root>/.moai/state/audit-enforcement.log (design.md §5).
func appendAuditTrail(in CeilingInput, kind, outcome, reason string) error {
	dir := filepath.Join(in.ProjectRoot, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	line := fmt.Sprintf("%s spec=%s kind=%s outcome=%s reason=%q\n",
		time.Now().UTC().Format(time.RFC3339), in.SpecID, kind, outcome, reason)
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
func AcknowledgeRequiredBackend(in CeilingInput, backend, note string) error {
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
