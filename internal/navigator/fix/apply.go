package fix

// apply.go — M3.5 apply-on-approval engine (SPEC-NAVIGATOR-SYNC-005,
// AC-NS5-008a / AC-NS5-008c / AC-NS5-008d). The deterministic post-approval
// layer-3 consumer: given a draft-id whose draft was approved at the
// orchestrator's AskUserQuestion gate (layer 3a, design.md §B), it
// atomic-renames the approved + scope-conformant draft subtrees into their
// target live doc surfaces and records the applied.json ledger.
//
// The approval is a real artifact, not "someone typed a command" (plan.md
// §C.6): the orchestrator writes approval.json carrying a DETERMINISTIC token
// = SHA-256(draft-id + approval-option + request.json provenance) at gate
// time. The --apply CLI entry point validates this token before applying; a
// direct shell invocation without a valid token is REFUSED (exit non-zero, no
// live-doc mutation). This is the hard guard (REQ-NS5-008 sub-clause c4) —
// distinct from the fail-open degraded paths (REQ-NS5-009, M3.6) which exit 0.
//
// Idempotence (DBT-2, plan.md §C.6): the applied.json ledger records each
// applied subtree ID; a resume after a crash-mid-set skips already-applied IDs
// and does NOT re-rename or error.
//
// Scope-conformance (REQ-NS5-013): an out-of-scope draft subtree (over-produced
// by manager-develop) is excluded from the apply AND logged via
// LogScopeExclusion (reusing M3.4's ConformDraftToScope). The live doc surface
// it would have mutated is NOT touched beyond the in-scope patches.
//
// Provenance (REQ-NS5-004, no wall-clock): ApprovalTimestamp is
// git-committer-date (`git log -1 --format=%cI`), NOT time.Now() — so two
// apply runs on the same HEAD produce byte-identical applied.json.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Approval is the approval.json record (REQ-NS5-008 sub-clause c4). Written by
// the orchestrator at gate time; validated by Apply before any live-doc
// mutation. The Token is a DETERMINISTIC hash (ComputeApprovalToken), NOT a
// random nonce — plan.md §C.6.
type Approval struct {
	DraftID           string `json:"draft_id"`
	ApprovalOption    string `json:"approval_option"` // "a" (approve+apply) | "b" (approve selected) | "c" (edit then apply)
	Approver          string `json:"approver"`
	ApprovalTimestamp string `json:"approval_timestamp"` // git-committer-date (`git log -1 --format=%cI`)
	Token             string `json:"token"`
}

// ApplyOptions configures an Apply invocation.
type ApplyOptions struct {
	ProjectRoot string
	DraftID     string
}

// ApplyResult records the apply outcome (returned to the CLI / orchestrator).
// Applied is false on the token-refusal path (Apply returns a non-nil error
// alongside); true on a successful apply (including the idempotent no-op
// resume case where every subtree was already applied).
type ApplyResult struct {
	Applied           bool
	AppliedSubtreeIDs []string
	LiveDocSHAs       map[string]string // docSurface → resulting content SHA
	SkippedAlready    []string          // subtree IDs skipped (DBT-2 idempotence)
	ExcludedOutOfScope []string         // subtree IDs excluded (REQ-NS5-013)
}

// Sentinel errors for the token-refusal path (AC-NS5-008d). The CLI caller
// translates either to exit-non-zero + a message naming the missing/invalid
// approval.json token. The two MUST NOT share an exit-code contract with the
// fail-open degraded paths (REQ-NS5-009, M3.6) which exit 0.
var (
	ErrApprovalTokenMissing = errors.New("approval.json token missing — orchestrator approval required before apply (AC-NS5-008d)")
	ErrApprovalTokenInvalid = errors.New("approval.json token invalid — token does not match draft-id + option + request.json provenance (AC-NS5-008d)")
)

// approvalJSONName is the approval record filename at the draft staging dir.
const approvalJSONName = "approval.json"

// appliedJSONName is the apply-ledger filename at the draft staging dir.
const appliedJSONName = "applied.json"

// draftSubdir is the layer-2 AI-drafted output directory under the staging dir.
const draftSubdir = "draft"

// ComputeApprovalToken returns the deterministic approval token: SHA-256 hex
// of (draft-id + approval-option + request.json provenance) (plan.md §C.6).
// NOT a random nonce. The orchestrator calls this when writing approval.json;
// Apply re-derives it for validation. Two calls on identical inputs produce
// the identical token (deterministic — verification-claim-integrity §2).
//
// The hash inputs are length-prefixed with a NUL separator so concatenation
// ambiguity (e.g. "ab"+"c" == "a"+"bc") cannot produce a token collision.
//
// @MX:ANCHOR: [AUTO] deterministic approval-token derivation (REQ-NS5-008 c4)
// @MX:REASON: the token is the load-binding contract that makes the approval a real artifact, not "someone typed a command" — Apply validates it before any live-doc mutation; a non-deterministic or random token would collapse the guard
func ComputeApprovalToken(draftID, option string, prov Provenance) string {
	h := sha256.New()
	h.Write([]byte(draftID))
	h.Write([]byte{0})
	h.Write([]byte(option))
	h.Write([]byte{0})
	h.Write([]byte(prov.FixCommitSHA))
	h.Write([]byte{0})
	h.Write([]byte(prov.BaselineCommitSHA))
	h.Write([]byte{0})
	h.Write([]byte(prov.CapturedAt))
	return hex.EncodeToString(h.Sum(nil))
}

// Apply is the apply-on-approval entry point (AC-NS5-008c / AC-NS5-008d).
//
// Flow:
//  1. Load request.json (the diff_scope + provenance).
//  2. Load + validate approval.json (AC-NS5-008d — refusal on missing/invalid
//     token; returns ErrApprovalTokenMissing / ErrApprovalTokenInvalid).
//  3. Scope-conformance (REQ-NS5-013): partition draft subtree IDs vs
//     diff_scope; exclude out-of-scope + log.
//  4. DBT-2 idempotence: load existing applied.json; skip already-applied IDs.
//  5. For each in-scope, not-already-applied doc surface, atomic-rename the
//     draft content into the live doc surface (AC-NS5-008c).
//  6. Write/update applied.json ledger (approver + git-committer-date +
//     applied IDs + resulting live-doc SHA).
//
// Returns (ApplyResult, nil) on a successful apply OR an idempotent no-op
// resume. Returns (zero, ErrApprovalToken*) on the token-refusal path. The
// only other non-nil errors are for unreadable request.json (a malformed
// staging dir); those are NOT token-refusal and the CLI MAY render them as
// exit-non-zero with the verbatim message.
//
// @MX:ANCHOR: [AUTO] apply-on-approval engine (AC-NS5-008c/008d)
// @MX:REASON: this is the sole layer-3 write surface that mutates live doc surfaces — every other Fix-layer write targets the staging dir; the approval-token guard + scope-conformance + idempotence are the three invariants that make the apply safe
func Apply(opts ApplyOptions) (ApplyResult, error) {
	root := resolveRoot(opts.ProjectRoot)
	draftDir := filepath.Join(root, fixDraftsRelDir, opts.DraftID)

	// 1. Load request.json — the diff_scope + provenance. A missing/unparseable
	//    request.json is a malformed staging dir (not token-refusal); surface
	//    as a plain error so the CLI can render it.
	req, err := loadRequestFile(draftDir)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("apply: load request.json: %w", err)
	}

	// 2. Load + validate approval.json (AC-NS5-008d).
	apv, err := loadApprovalFile(draftDir)
	if err != nil {
		logFix(root, fmt.Sprintf("navigator-fix apply: refused — approval.json missing/unreadable for draft %s: %v", opts.DraftID, err))
		return ApplyResult{}, fmt.Errorf("%w: draft %s: %v", ErrApprovalTokenMissing, opts.DraftID, err)
	}
	if apv.DraftID != opts.DraftID {
		logFix(root, fmt.Sprintf("navigator-fix apply: refused — approval.json draft_id %q != apply draft %s", apv.DraftID, opts.DraftID))
		return ApplyResult{}, fmt.Errorf("%w: approval.json draft_id %q does not match apply draft %s", ErrApprovalTokenInvalid, apv.DraftID, opts.DraftID)
	}
	expectedToken := ComputeApprovalToken(apv.DraftID, apv.ApprovalOption, req.Provenance)
	if apv.Token == "" || apv.Token != expectedToken {
		logFix(root, fmt.Sprintf("navigator-fix apply: refused — approval token invalid for draft %s", opts.DraftID))
		return ApplyResult{}, ErrApprovalTokenInvalid
	}

	// 3. Scope-conformance (REQ-NS5-013). The draft subtree IDs come from the
	//    draft.json manifest (if present) or default to the diff_scope subtree
	//    IDs (the in-scope set is the contract).
	draftSubtreeIDs := loadDraftSubtreeIDs(draftDir, req.DiffScope)
	inScope, excluded := ConformDraftToScope(draftSubtreeIDs, req.DiffScope)
	LogScopeExclusion(root, excluded, req.DiffScope)

	// 4. DBT-2 idempotence — load the existing ledger (if any) and skip
	//    already-applied IDs.
	existing := loadAppliedFile(draftDir)
	already := make(map[string]bool, len(existing.AppliedSubtreeIDs))
	for _, id := range existing.AppliedSubtreeIDs {
		already[id] = true
	}

	// 5. Determine the to-apply set: in-scope AND not-already-applied.
	toApply := make([]string, 0, len(inScope))
	skipped := make([]string, 0)
	for _, id := range inScope {
		if already[id] {
			skipped = append(skipped, id)
			continue
		}
		toApply = append(toApply, id)
	}
	sort.Strings(toApply)
	sort.Strings(skipped)

	// Group to-apply IDs by their doc surface (one write per surface).
	surfaceToIDs := make(map[string][]string)
	for _, id := range toApply {
		surface := surfaceForSubtree(id, req.DiffScope)
		if surface == "" {
			// Defensive: an in-scope ID with no matching diff_scope entry should
			// not happen (ConformDraftToScope guarantees membership); skip it
			// rather than fail the whole apply.
			logFix(root, fmt.Sprintf("navigator-fix apply: subtree %s has no doc-surface mapping in diff_scope, skipping", id))
			continue
		}
		surfaceToIDs[surface] = append(surfaceToIDs[surface], id)
	}

	// 6. Atomic-rename each doc surface's draft content into the live surface.
	appliedIDs := make([]string, 0)
	liveSHAs := make(map[string]string)
	for surface, ids := range surfaceToIDs {
		draftPath := filepath.Join(draftDir, draftSubdir, surface)
		livePath := filepath.Join(root, liveDocRelPath(surface))
		content, err := os.ReadFile(draftPath)
		if err != nil {
			logFix(root, fmt.Sprintf("navigator-fix apply: draft content unreadable for surface %s, skipping: %v", surface, err))
			continue
		}
		// atomicWriteFile (the M0/M2 pattern carried forward in request.go)
		// writes <live>.tmp then renames — a reader never sees a partial file.
		if err := atomicWriteFile(livePath, content); err != nil {
			return ApplyResult{}, fmt.Errorf("apply: atomic-write live doc %s: %w", livePath, err)
		}
		appliedIDs = append(appliedIDs, ids...)
		sum := sha256.Sum256(content)
		liveSHAs[surface] = hex.EncodeToString(sum[:])
	}
	sort.Strings(appliedIDs)

	// Merge with already-applied for the final ledger (idempotent accumulation).
	allApplied := append([]string{}, existing.AppliedSubtreeIDs...)
	allApplied = append(allApplied, appliedIDs...)
	allApplied = dedupSorted(allApplied)

	// Resulting live-doc SHA: composite hash across all in-scope doc surfaces
	// (sorted for byte-stability). This is the post-apply live-doc fingerprint
	// the ledger records (AC-NS5-008c "resulting live-doc SHA").
	resultingSHA := compositeLiveDocSHA(req.DiffScope, func(surface string) string {
		if sha, ok := liveSHAs[surface]; ok {
			return sha
		}
		// Surface not re-applied this round (already-applied) — read its
		// current content hash so the composite reflects the true post-apply
		// state, not just this round's deltas.
		data, err := os.ReadFile(filepath.Join(root, liveDocRelPath(surface)))
		if err != nil {
			return ""
		}
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	})

	// 7. Write/update the applied.json ledger. ApprovalTimestamp is
	//    git-committer-date (NOT wall-clock) — REQ-NS5-004 no-wall-clock.
	approvalTS := apv.ApprovalTimestamp
	if approvalTS == "" {
		approvalTS = gitLine(root, "log", "-1", "--format=%cI")
	}
	ledger := AppliedLedger{
		Approver:            apv.Approver,
		ApprovalTimestamp:   approvalTS,
		AppliedSubtreeIDs:   allApplied,
		ResultingLiveDocSHA: resultingSHA,
	}
	if err := writeAppliedFile(draftDir, ledger); err != nil {
		return ApplyResult{}, fmt.Errorf("apply: write applied.json: %w", err)
	}

	return ApplyResult{
		Applied:            true,
		AppliedSubtreeIDs:  appliedIDs,
		LiveDocSHAs:        liveSHAs,
		SkippedAlready:     skipped,
		ExcludedOutOfScope: excluded,
	}, nil
}

// --- helpers -----------------------------------------------------------------

// loadRequestFile reads + unmarshals request.json from the draft staging dir.
func loadRequestFile(draftDir string) (DraftRequest, error) {
	var req DraftRequest
	data, err := os.ReadFile(filepath.Join(draftDir, "request.json"))
	if err != nil {
		return req, err
	}
	if err := json.Unmarshal(data, &req); err != nil {
		return req, fmt.Errorf("unmarshal request.json: %w", err)
	}
	return req, nil
}

// loadApprovalFile reads + unmarshals approval.json from the draft staging dir.
func loadApprovalFile(draftDir string) (Approval, error) {
	var apv Approval
	data, err := os.ReadFile(filepath.Join(draftDir, approvalJSONName))
	if err != nil {
		return apv, err
	}
	if err := json.Unmarshal(data, &apv); err != nil {
		return apv, fmt.Errorf("unmarshal approval.json: %w", err)
	}
	return apv, nil
}

// loadAppliedFile reads the existing applied.json ledger (DBT-2 idempotence).
// Returns the zero value when absent (first apply) — never errors.
func loadAppliedFile(draftDir string) AppliedLedger {
	var ledger AppliedLedger
	data, err := os.ReadFile(filepath.Join(draftDir, appliedJSONName))
	if err != nil {
		return ledger
	}
	_ = json.Unmarshal(data, &ledger)
	return ledger
}

// writeAppliedFile marshals + atomic-writes the applied.json ledger.
func writeAppliedFile(draftDir string, ledger AppliedLedger) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ledger); err != nil {
		return err
	}
	data := bytes.TrimRight(buf.Bytes(), "\n")
	return atomicWriteFile(filepath.Join(draftDir, appliedJSONName), data)
}

// loadDraftSubtreeIDs returns the draft's subtree-ID set, for scope-conformance
// validation. The IDs come from the layer-2 draft.json manifest when present;
// otherwise they default to the diff_scope subtree IDs (the in-scope contract —
// the draft is assumed to cover exactly diff_scope when no manifest exists).
func loadDraftSubtreeIDs(draftDir string, diffScope []DiffScopeEntry) []string {
	manifestPath := filepath.Join(draftDir, draftSubdir, "draft.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var m struct {
			Subtrees []struct {
				SubtreeID  string `json:"subtree_id"`
				DocSurface string `json:"doc_surface"`
			} `json:"subtrees"`
		}
		if json.Unmarshal(data, &m) == nil && len(m.Subtrees) > 0 {
			ids := make([]string, 0, len(m.Subtrees))
			for _, s := range m.Subtrees {
				if s.SubtreeID != "" {
					ids = append(ids, s.SubtreeID)
				}
			}
			return ids
		}
	}
	// Default: the draft covers the diff_scope IDs.
	ids := make([]string, 0, len(diffScope))
	for _, e := range diffScope {
		ids = append(ids, e.SubtreeID)
	}
	return ids
}

// surfaceForSubtree returns the doc surface for a subtree ID, looked up from
// the diff_scope. Returns "" when the ID is not in diff_scope.
func surfaceForSubtree(subtreeID string, diffScope []DiffScopeEntry) string {
	for _, e := range diffScope {
		if e.SubtreeID == subtreeID {
			return e.DocSurface
		}
	}
	return ""
}

// liveDocRelPath returns the project-root-relative path of a live doc surface
// under the navigator output directory (.moai/project/navigator/<surface>).
func liveDocRelPath(docSurface string) string {
	return filepath.Join(".moai", "project", "navigator", docSurface)
}

// compositeLiveDocSHA returns a composite SHA-256 over the per-surface content
// SHAs of every distinct doc surface in diff_scope, sorted by surface name for
// byte-stability. The shaOf surface callback returns each surface's content
// SHA (empty surfaces contribute an empty string).
func compositeLiveDocSHA(diffScope []DiffScopeEntry, shaOf func(surface string) string) string {
	seen := make(map[string]bool)
	for _, e := range diffScope {
		seen[e.DocSurface] = true
	}
	surfaces := make([]string, 0, len(seen))
	for s := range seen {
		surfaces = append(surfaces, s)
	}
	sort.Strings(surfaces)
	h := sha256.New()
	for _, s := range surfaces {
		_, _ = h.Write([]byte(s))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(shaOf(s)))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// dedupSorted returns a sorted, deduplicated copy of the input slice.
func dedupSorted(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
