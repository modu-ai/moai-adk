// factory_card_rebind.go — card t1930 defect ④: the formal evidence re-bind
// verb `moai factory rebind-evidence <card> --run <run> --reason "<text>"`.
//
// The recorded evidence_sha is written only by the T5/T11 entry guards, which
// require the newest phase verdict's audited_sha to match the recorded
// evidence commit. The audit loop's normal conclusion — repair commits land
// AFTER the verdict — diverges the two permanently, and the guards then
// refuse forever; lanes resorted to direct SQL against the factory database
// (forbidden; five reproductions: t1832, t1536, t1929, t1828, and t1828's
// re-issue workaround). This verb is the formalization of that
// operator-approved exception path:
//
//   - the new value comes ONLY from the card's newest audit verdict file
//     (newest-selection rule of readAuditVerdict; the phase is inferred from
//     the card's stage) and must satisfy the T10 commit conditions in the
//     card worktree — a caller cannot pass an arbitrary SHA, so an audit
//     claim cannot be forged through this verb. The one other source (card
//     t1962) is the landing commit the OPERATOR recorded with `todo landed
//     --sha`, consulted only when the verdict path refuses (a squash-landed
//     PR leaves the audited head outside the worktree's ancestry) and only
//     when that commit is reachable from the landed ref; the verdict wins
//     whenever it resolves;
//   - the record write is version-fenced exactly like set-spec's record
//     layer (internal/homestate/card_metadata.go);
//   - exactly one card_metadata_events audit row (kind rebind-evidence,
//     field evidence_sha, old → new, reason) is appended AFTER the store
//     commit — never inside the mutation;
//   - the verb is lane-callable by design (no operator flag gate, no lane
//     refusal): the SQL exceptions it replaces were lane-executed under
//     operator approval, and the verdict binding plus the mandatory audit
//     row are the anti-abuse boundary.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The rebind verb's audit vocabulary (the card_metadata_events family).
const (
	metadataKindRebindEvidence = "rebind-evidence"
	metadataFieldEvidenceSHA   = "evidence_sha"
)

// rebindAuditPhase infers which phase's verdict file binds the card's
// evidence (card t1930 defect ④): plan stages bind the plan-audit verdict,
// everything from run onward binds the sync-audit verdict — the same split
// the T6/T7 and T13 guards read. An empty stage falls back to the state.
func rebindAuditPhase(c homestate.Card) (string, error) {
	s := c.Stage
	if s == "" {
		s = c.State
	}
	switch s {
	case homestate.CardPlan, homestate.CardPlanAudit:
		return "plan-audit", nil
	case homestate.CardRun, homestate.CardSync, homestate.CardSyncAudit,
		homestate.CardMergeReady, homestate.CardMerging, homestate.CardPROpen, homestate.CardMergedPR:
		return "sync-audit", nil
	}
	return "", fmt.Errorf("cannot infer the audit phase from stage %q (state %s)", c.Stage, c.State)
}

// newFactoryRebindEvidenceCommand builds `moai factory rebind-evidence <card>
// --run <run> --reason "<text>"`.
func newFactoryRebindEvidenceCommand() *cobra.Command {
	var run, reason string
	cmd := &cobra.Command{
		Use:   "rebind-evidence <card>",
		Short: "Re-bind a card's evidence_sha to its newest audit verdict's audited commit (audit-trailed; card t1930)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runFactoryRebindEvidence(cmd, args[0], reason, run); err != nil {
				return fmt.Errorf("factory rebind-evidence: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "",
		"why the evidence is being re-bound (required, 1-500 runes; recorded verbatim in the audit row)")
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// resolveRecordedLanding reads the card's operator-recorded landing commit
// (`moai todo landed --sha`) and verifies it against the ref that record
// names (the configured landed ref when it names none) in the queue root's
// repository, where the record was made. An empty
// commit means the fallback did not apply; note is then the suffix for the
// verdict refusal — empty when the card carries no landing record, so an
// unlanded card keeps the verdict's own refusal.
//
// The record is stored per card id, not per run, so a reissued card still
// carries the previous issuance's landing. A record observed before this run
// first engaged the card is refused; when the run holds no event for the card
// the engagement start is unknown and the record is judged on its commit alone
// (refusal needs positive evidence).
func resolveRecordedLanding(ctx context.Context, root string, db *homestate.FactoryDB, runID string, card homestate.Card) (full, note string) {
	rec, err := newTodoStore().LoadPure()
	if err != nil {
		return "", fmt.Sprintf("; landing record unreadable: %v", err)
	}
	var sha, ref, observed string
	for i := range rec.Items {
		if rec.Items[i].ID == card.CardID && rec.Items[i].Landing != nil {
			sha = strings.TrimSpace(rec.Items[i].Landing.SHA)
			ref = strings.TrimSpace(rec.Items[i].Landing.Ref)
			observed = strings.TrimSpace(rec.Items[i].Landing.ObservedAt)
		}
	}
	if sha == "" {
		return "", ""
	}
	start, known, err := db.CardEngagementStart(ctx, runID, card.CardID)
	if err != nil {
		return "", fmt.Sprintf("; run engagement start unreadable: %v", err)
	}
	if known {
		at, perr := time.Parse(time.RFC3339Nano, observed)
		if perr != nil {
			return "", fmt.Sprintf("; recorded landing refused: unreadable observation time %q", observed)
		}
		// `todo landed` records ObservedAt at second precision, the engagement
		// start carries fractions: compare at the record's precision, so only a
		// strictly earlier second is "before" the engagement.
		if at.Before(start.Truncate(time.Second)) {
			return "", fmt.Sprintf("; recorded landing refused: observed %s, before this run engaged the card (%s) — it belongs to an earlier issuance",
				at.UTC().Format(time.RFC3339), start.UTC().Format(time.RFC3339))
		}
	}
	// The record names the ref the operator observed the landing against
	// (`todo landed --ref`); only a record without one falls back to the
	// configured landed ref.
	if ref == "" {
		ref = factory.LandedRefFor(root)
	}
	// The record was made in the queue root's repository (`todo landed`
	// resolves its ref there), so it is verified there: a per-worktree ref
	// such as HEAD means the root's HEAD, not the card worktree's.
	full, err = homestate.ResolveLandedEvidence(ctx, root, sha, ref)
	if err != nil {
		return "", fmt.Sprintf("; recorded landing commit refused: %v", err)
	}
	return full, ""
}

func runFactoryRebindEvidence(cmd *cobra.Command, cardID, reason, run string) error {
	if !homestate.ValidCardID(cardID) {
		return fmt.Errorf("%q is not a card id", cardID)
	}
	// The reason gate (the same bound the metadata verbs carry): the audit
	// row records why, so an empty reason is refused before anything opens.
	if err := validateMetadataReason(reason); err != nil {
		return err
	}
	root := factoryCardRoot()
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// The record must exist before the store opens (the set-spec convention:
	// OpenFactory migrates on open, so a refused path must not materialize a
	// store as a side effect of the read that found nothing).
	fdbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return err
	}
	if _, serr := os.Stat(fdbPath); errors.Is(serr, fs.ErrNotExist) {
		return fmt.Errorf("card %s has no factory record — rebind-evidence re-binds an existing record; it creates none", cardID)
	}
	runID, err := resolveSetSpecRun(ctx, fdbPath, run)
	if err != nil {
		return err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(ctx, runID, cardID)
	if errors.Is(err, homestate.ErrCardNotFound) {
		return fmt.Errorf("card %s has no factory record under run %s — rebind-evidence re-binds an existing record; it creates none", cardID, runID)
	}
	if err != nil {
		return err
	}
	phase, err := rebindAuditPhase(card)
	if err != nil {
		return err
	}
	// Binding safety: the new value is READ from the card's newest verdict
	// file of the inferred phase and verified in the worktree — never taken
	// from the caller. A missing or unreadable verdict, or an audited commit
	// that does not resolve as an ancestor of the worktree HEAD, falls to the
	// operator's recorded landing commit (see the header) and refuses without
	// writing anything when that does not apply either.
	full, err := homestate.ResolveRebindEvidence(ctx, card, phase)
	source := phase + " verdict's audited commit"
	evidenceSource := homestate.EvidenceSourceVerdict
	if err != nil {
		// A PR that landed as a squash commit leaves the audited PR head
		// outside the worktree's ancestry (card t1962); the operator's
		// landing record names the commit that did land.
		landed, note := resolveRecordedLanding(ctx, root, db, runID, card)
		if landed == "" {
			return fmt.Errorf("card %s: no %s verdict to re-bind onto: %w%s", cardID, phase, err, note)
		}
		full, source, evidenceSource = landed, "recorded landing commit", homestate.EvidenceSourceLanding
	}
	actor, session := metadataActor()
	now := factoryCardNow()
	out := cmd.OutOrStdout()
	old := strings.TrimSpace(card.EvidenceSHA)
	if full == old {
		// The record already carries the verdict's sha (card-review r1, F2):
		// a converged-and-audited state is a duplicate, but a converged
		// record whose audit row is missing is the crashed run's recovery —
		// the catch-up append lands the owed attestation without re-writing
		// the record (version unchanged), so a failed append can always be
		// healed by a re-run.
		latest, audited, lerr := db.LatestCardMetadataNewValue(ctx, cardID, metadataKindRebindEvidence, metadataFieldEvidenceSHA)
		if lerr != nil {
			return lerr
		}
		if audited && latest == full {
			return fmt.Errorf("card %s already records evidence %s and the audit surface's latest rebind-evidence row already records it (converged-and-audited duplicate)", cardID, full)
		}
		auditOld := ""
		if audited {
			auditOld = latest
		}
		for attempt := 0; attempt < 2; attempt++ {
			if err := db.RecordCardMetadataCatchUp(ctx, homestate.CardMetadataEvent{
				CardID: cardID, Kind: metadataKindRebindEvidence, Actor: actor, Session: session,
				Field: metadataFieldEvidenceSHA, Old: auditOld, New: full, Reason: reason, Source: evidenceSource,
			}, now); err == nil {
				_, _ = fmt.Fprintf(out, "%s evidence_sha %s (attested; the record already carried it, no store re-mutated)\n", cardID, shortSHA(full))
				return nil
			}
		}
		return fmt.Errorf("%s on %s (rebind-evidence catch-up → %s): the record already carried the value and was not re-mutated, but the audit append failed after one retry; the missing attestation is surfaced — re-run rebind-evidence to land it",
			metadataAuditLossSentinel, cardID, full)
	}
	// The record write: version-fenced (a stale read refuses with
	// ErrStaleVersion and nothing is written).
	if _, err := db.SetCardEvidenceSHA(ctx, runID, cardID, full, card.Version, now); err != nil {
		return err
	}
	// The audit append only after the store commit — one retry, then the
	// audit-loss sentinel with the completed write left in place (the same
	// discipline the metadata verbs carry); the catch-up branch above heals it.
	for attempt := 0; attempt < 2; attempt++ {
		if err := db.RecordCardMetadataEvent(ctx, homestate.CardMetadataEvent{
			CardID: cardID, Kind: metadataKindRebindEvidence, Actor: actor, Session: session,
			Field: metadataFieldEvidenceSHA, Old: old, New: full, Reason: reason, Source: evidenceSource,
		}, now); err == nil {
			_, _ = fmt.Fprintf(out, "%s evidence_sha %s (was %s) — rebound to the newest %s\n",
				cardID, shortSHA(full), dash(shortSHA(old)), source)
			return nil
		}
	}
	return fmt.Errorf("%s on %s (rebind-evidence → %s): the record committed but the audit append failed after one retry; the missing row is surfaced — re-run rebind-evidence to land the catch-up attestation",
		metadataAuditLossSentinel, cardID, full)
}
