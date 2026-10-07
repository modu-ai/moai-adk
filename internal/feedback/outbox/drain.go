// Package outbox is the local pipeline of the opt-in participation flow
// (SPEC-FEEDBACK-PARTICIPATION-001 design.md section 6): the drain that
// consumes the user-scoped capture spool, the dedupe/caps ledger, the scrub
// tripwire, the one render function the preview and the publication create
// path share, the withdrawal discard, and the append-only outbox log.
//
// The package imports internal/bugreport, internal/feedback, and
// internal/config — and neither os/exec nor net/http (REQ-ANON-025): this
// half of the pipeline is network-free by construction. Publication happens
// in internal/feedback/publish, which the CLI's flush call drives.
package outbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// Store file names under <moai home>/state/bugreport/ (design.md section 6,
// D36 — every bugreport store is user-scoped). The spool's file name and
// the store directory live in the SSOT bugreport package (spoolFileName /
// BugreportStoreDir) — no local copies.
const (
	QueueFileName  = "queue.json"
	LedgerFileName = "ledger.json"
	OutboxFileName = "outbox.log"
	outboxFilePerm = 0o600
	storeDirPerms  = 0o700
	titleKeyPrefix = "[auto-report] "
	schemaMarkerV1 = "v1"
	issueMarkerV1  = "<!-- moai-bugreport:v1 "
)

// OutboxRow is one outbox log line. Payload-bearing rows (queued, sent,
// withheld) carry the exact title and body; decision rows (deduped, capped,
// dropped, discarded, ambiguous) carry the reason only — never payload text
// for non-queued outcomes (design.md section 6).
type OutboxRow struct {
	At       string `json:"at"`
	Outcome  string `json:"outcome"`
	Reason   string `json:"reason,omitempty"`
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	Fingerpr string `json:"fingerprint,omitempty"`
	Masked   bool   `json:"masked,omitempty"`
}

// storeDir returns the user-scoped store directory, creating it 0700.
func storeDir() (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	if err := os.MkdirAll(dir, storeDirPerms); err != nil {
		return "", err
	}
	return dir, nil
}

// StorePath returns one store file's path under the user-scoped directory.
func StorePath(name string) (string, error) {
	dir, err := storeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// AppendOutbox appends one row to the append-only outbox log (0600). The
// log is the pipeline's durable trail: withdrawal keeps it, purge removes
// it.
func AppendOutbox(row OutboxRow) error {
	path, err := StorePath(OutboxFileName)
	if err != nil {
		return err
	}
	if row.At == "" {
		row.At = clock().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(row)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, outboxFilePerm)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(append(line, '\n'))
	return err
}

// clock is the pipeline's time seam (tests pin it for the dedupe window and
// the rolling caps).
var clock = time.Now

// SetClockForTest pins the drain's clock; nil restores time.Now.
func SetClockForTest(fn func() time.Time) {
	if fn == nil {
		clock = time.Now
		return
	}
	clock = fn
}

// buildIdentityForTest is the build-identity seam the caps test uses to
// vary the fingerprint per signal; production reads the compiled identity.
var buildIdentityForTest func() (version, commit string)

// RenderReport renders a validated payload into the exact title and body
// the queue carries, the preview prints, and the publication create path
// hands to gh — the ONE render function (REQ-ANON-014, AC-020).
//
// The title is the duplicate-lookup key; the body opens with the
// machine-readable marker block the issue contract defines (REQ-ANON-019).
// The rendering takes the validated payload only — no error text, no free
// strings — and carries no timestamp.
func RenderReport(p bugreport.Payload) (title, body string) {
	title = IssueTitle(p)
	var b strings.Builder
	b.WriteString(IssueMarker(p))
	b.WriteString("\n\n")
	b.WriteString("A moai-adk user opted into automatic improvement participation and hit this defect.\n")
	b.WriteString("This report is machine-generated from closed fixed fields: no error text, no paths,\n")
	b.WriteString("no project or session data. The fingerprint family key is in the title.\n")
	body = b.String()
	return title, body
}

// RenderReportWithSummary renders the report with the model summary's
// section slotted AHEAD of the deterministic template text (design section
// 8; D38's M6 half — the preview keeps printing the queued pre-summary
// render). An empty summary renders BYTE-IDENTICAL to RenderReport — the
// template path never drifts from the one render.
func RenderReportWithSummary(p bugreport.Payload, summary string) (title, body string) {
	title = IssueTitle(p)
	if strings.TrimSpace(summary) == "" {
		return RenderReport(p)
	}
	var b strings.Builder
	b.WriteString(IssueMarker(p))
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(summary))
	b.WriteString("\n\n")
	b.WriteString("A moai-adk user opted into automatic improvement participation and hit this defect.\n")
	b.WriteString("This report is machine-generated from closed fixed fields: no error text, no paths,\n")
	b.WriteString("no project or session data. The fingerprint family key is in the title.\n")
	return title, b.String()
}

// IssueTitle is the contract title: `[auto-report] <kind> <fingerprint>` —
// the exact title key the duplicate lookup matches (REQ-ANON-019).
func IssueTitle(p bugreport.Payload) string {
	return titleKeyPrefix + string(p.Kind) + " " + p.Fingerprint
}

// IssueMarker is the body's opening machine-readable block: the fields
// REQ-ANON-019 names, no more. A value that itself carries a space — the
// hook detail's canonical token — is quoted so the marker's field
// tokenizer keeps it as one value (the marker round-trip, outbox/marker.go).
func IssueMarker(p bugreport.Payload) string {
	detail := ""
	if p.Detail != nil {
		detail = " detail=" + QuoteMarkerValue(p.Detail.Token())
	}
	return fmt.Sprintf("%sschema=%s fingerprint=%s kind=%s version=%s commit=%s os_arch=%s/%s frames=%s%s -->",
		issueMarkerV1, p.Schema, p.Fingerprint, string(p.Kind), p.Version, p.Commit, p.OS, p.Arch,
		strings.Join(p.Frames, ","), detail)
}

// entryIdentity is the build identity a spool line carries: the
// CAPTURE-TIME version and commit stamped by the binary that observed the
// defect (review-gate P2). Entries written before that field existed fall
// back to the flushing binary's identity — the old behavior — rather than
// dropping the signal.
func entryIdentity(entry bugreport.SpoolEntry) (string, string) {
	if entry.Version != "" && entry.Commit != "" {
		return entry.Version, entry.Commit
	}
	return buildIdentity()
}

// fingerprintOf derives the dedupe key from the spool entry: the same
// canonical inputs the payload's fingerprint uses, computed from the spool
// line (design section 4). Frames carry the moai-internal names; the
// version/commit inputs are the entry's CAPTURE-TIME identity so two
// signals from one build collide and one defect yields one issue — and a
// signal flushed by a different binary still keys on the build that
// observed it.
func fingerprintOf(entry bugreport.SpoolEntry) string {
	version, commit := entryIdentity(entry)
	lines := []string{
		bugreport.SchemaV1,
		version,
		commit,
		runtime.GOOS + "/" + runtime.GOARCH,
		string(entry.Kind),
	}
	lines = append(lines, entry.Frames...)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])[:16]
}

// buildIdentity is the build identity the fingerprint uses. The seam exists
// for the caps test, which varies the identity per signal so each drain sees
// a fresh fingerprint; production reads the compiled build identity.
var buildIdentity = func() (ver, commit string) {
	if buildIdentityForTest != nil {
		return buildIdentityForTest()
	}
	return version.GetVersion(), version.GetCommit()
}

// Drain is DrainContext under the caller's background context.
func Drain() error {
	return DrainContext(context.Background())
}

// DrainContext consumes the user-scoped spool and runs each recorded signal through
// the local pipeline (design.md section 6):
//
//   - consent off → every unsent queue item and the spool are discarded,
//     each as a `discarded` row; the sent history (the outbox log) stays;
//   - an ambiguous verdict → one `ambiguous` retention row, stopped (DEC-7);
//   - a moai verdict → fingerprint → dedupe window → rolling caps → build
//     and validate the payload → render → scrub tripwire → enqueue to the
//     user-scoped queue through feedback.QueueStore → one outbox row.
//
// Each stage that stops a signal appends exactly one row naming the reason.
// Drain is network-free by construction (this package cannot import the
// network) and makes no model call (there is no model seam here). The
// context bounds the whole run — the CLI's flush time box: the per-item
// loop stops on cancellation and the queue-lock acquisition selects on it
// (review-gate P2: lock waits used to accumulate per item, so a
// 50ms-deadline drain ran thirteen seconds against a live lock holder).
// Consumption is prefix-scoped: only the entries handled to a terminal
// state leave the spool; a cancelled or retryably-failed item and
// everything after it stay for the next drain (whose ledger dedupe window
// makes reprocessing the tail safe).
func DrainContext(ctx context.Context) error {
	// Withdrawal first: consent off discards everything unsent and stops.
	if !config.ReadUserParticipation().Enabled {
		return discardAll(ctx)
	}

	entries, consumed, err := bugreport.ReadSpoolConsumable()
	if err != nil {
		return fmt.Errorf("outbox: read spool: %w", err)
	}
	if len(entries) == 0 && len(consumed) == 0 {
		return nil
	}

	// Test seam: a capture landing between the drain's read and its
	// consume — the interleaving the batch-clear contract must survive.
	if spoolAfterReadForTest != nil {
		spoolAfterReadForTest()
	}

	// processed counts the entries the drain handled to a terminal state —
	// queued, or decided and logged. ONLY that prefix leaves the spool
	// (review-gate finding: a retryable failure on any item used to consume
	// the whole batch, losing the failed item and everything after it — a
	// lost report).
	processed := 0
	for i, entry := range entries {
		if err := ctx.Err(); err != nil {
			// The deadline or cancellation arrived: the decided prefix
			// leaves the spool, the rest stays for the next drain.
			return consumeProcessed(consumed, len(entries), processed)
		}
		switch bugreport.Verdict(entry.Verdict) {
		case bugreport.VerdictAmbiguous:
			_ = AppendOutbox(OutboxRow{
				Outcome: "ambiguous",
				Reason:  "retained locally (DEC-7): no model call, nothing queued or sent",
			})
			processed = i + 1
			continue
		case bugreport.VerdictMoai:
			// continue below
		default:
			// user/environment verdicts never reach the spool (capture
			// drops them); a hostile line saying so is inert and consumed
			// with the prefix.
			processed = i + 1
			continue
		}

		outcome, retry := drainMoai(ctx, entry)
		if retry {
			// A retryable failure — context expiry, a lost lock budget, an
			// unreadable ledger: the failed item and everything after it
			// stay in the spool for the next drain (the ledger's dedupe
			// window makes reprocessing the tail safe).
			if outcome != nil {
				_ = AppendOutbox(OutboxRow{Outcome: outcome.outcome, Reason: outcome.reason, Fingerpr: outcome.fp})
			}
			return consumeProcessed(consumed, len(entries), i)
		}
		if outcome != nil {
			_ = AppendOutbox(OutboxRow{Outcome: outcome.outcome, Reason: outcome.reason, Fingerpr: outcome.fp})
			processed = i + 1
			continue
		}
		processed = i + 1
	}

	// Only the processed prefix is removed (bugreport.ConsumeSpoolPrefix,
	// under the spool's own cross-process section): entries captured after
	// the drain's read survive for the next drain. The predecessor
	// whole-file ClearSpool deleted every capture that landed mid-drain —
	// a lost report.
	return consumeProcessed(consumed, len(entries), processed)
}

// consumeProcessed removes exactly the spool prefix covering the first
// processed valid entries — the successfully handled part of the batch the
// drain read.
func consumeProcessed(consumed []byte, entries, processed int) error {
	prefix := consumed
	if processed < entries {
		prefix = consumed[:bugreport.PrefixLenForEntries(consumed, processed)]
	}
	if err := bugreport.ConsumeSpoolPrefix(prefix); err != nil {
		return fmt.Errorf("outbox: consume spool: %w", err)
	}
	return nil
}

// rollbackLedgerRecord removes the ledger reservation a FAILED queue
// mutation left behind: the fingerprint's stamp and exactly the QueuedAt
// entry this attempt appended (matched by its exact timestamp). The removal
// is OWNERSHIP-VERIFIED (review-gate finding 4, P2): a mutation that failed
// at lock acquisition ran no callback and wrote no reservation, so the
// recorded stamp must EQUAL this attempt's stamp before anything is
// touched — a prior success's record (a different stamp) survives, or the
// next drain would re-open the dedupe window of an already-queued report
// and publish it twice. Best-effort under its own SHORT queue-lock section
// — bounded by the caller's context and a quarter-second of its own, so the
// rollback never turns one failed wait into another full budget; a failure
// here leaves the orphan shape, which the dedupe check's recovery path
// re-queues.
func rollbackLedgerRecord(ctx context.Context, fp, stamp string) {
	rctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	store := BugreportQueueStore()
	_ = store.MutateContext(rctx, func(rec *feedback.QueueRecord) error {
		ledger, lerr := loadLedger()
		if lerr != nil {
			return nil // unreadable: leave it; the orphan recovery handles the retry
		}
		if cur, ok := ledger.FingerprintSeen[fp]; !ok || cur != stamp {
			// The record on disk is not this attempt's — either a prior
			// success's (leave it: it carries the dedupe window) or the
			// failed save never landed (nothing to roll back). (A
			// same-second collision with a prior record's stamp is below
			// RFC3339's resolution and is accepted, matching the QueuedAt
			// exact-timestamp match.)
			return nil
		}
		delete(ledger.FingerprintSeen, fp)
		for i, ts := range ledger.QueuedAt {
			if ts == stamp {
				ledger.QueuedAt = append(ledger.QueuedAt[:i], ledger.QueuedAt[i+1:]...)
				break
			}
		}
		return saveLedger(ledger)
	})
}

// queueHasFingerprint reports whether a live queue item already carries the
// fingerprint.
func queueHasFingerprint(items []feedback.QueueItem, fp string) bool {
	for i := range items {
		if items[i].Fingerprint == fp {
			return true
		}
	}
	return false
}

// SentHistoryHasFingerprintWithin reports whether the outbox log carries a
// sent row for the fingerprint INSIDE the duplicate-suppression window —
// the send-completion record, scoped like every other dedupe signal (the
// DEC-3 window). The drain's orphan recovery discriminates with it, and
// the sender reconciles a surviving queue item whose send already landed
// (review-gate hardening round: a send whose queue cleanup failed must not
// publish again) — and an expired-window recurrence must NOT be suppressed
// by an old row: past the window it is a new report.
func SentHistoryHasFingerprintWithin(fp string, now time.Time, windowDays int) bool {
	return sentHistoryHasFingerprint(fp, now, windowDays)
}

// sentHistoryHasFingerprint reports whether the outbox log carries a sent
// row for the fingerprint inside the window — the discriminator between a
// legitimately handled record (sent: the queue no longer holds it by
// design) and an orphaned one (recorded, never queued, never sent).
func sentHistoryHasFingerprint(fp string, now time.Time, windowDays int) bool {
	path, err := StorePath(OutboxFileName)
	if err != nil {
		return false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if line == "" {
			continue
		}
		var row OutboxRow
		if json.Unmarshal([]byte(line), &row) != nil {
			continue
		}
		if row.Outcome != "sent" || row.Fingerpr != fp {
			continue
		}
		at, perr := time.Parse(time.RFC3339, row.At)
		if perr != nil {
			continue // an unreadable stamp suppresses nothing
		}
		if now.Sub(at) < time.Duration(windowDays)*24*time.Hour {
			return true
		}
	}
	return false
}

// spoolAfterReadForTest runs between the drain's read and its consume (nil
// in production); the batch-clear test interposes a capture there.
var spoolAfterReadForTest func()

// drainOutcome is the reason a moai signal stopped before the queue.
type drainOutcome struct {
	outcome string
	reason  string
	fp      string
}

// queueWriteBlockForTest, when set and returning true for an entry, makes
// drainMoai report a queue-write failure for that entry without touching
// the lock — the deterministic driver for the partial-consumption tests (a
// real cross-process contention would be timing-dependent).
var queueWriteBlockForTest func(entry bugreport.SpoolEntry) bool

// queueCommitFailForTest makes the mutation callback fail AFTER the ledger
// record lands — the exact state a queue-replace failure produces (ledger
// recorded, queue unchanged): the orphaned-record shape the recovery tests
// seed and repair.
var queueCommitFailForTest bool

// drainMoai runs one moai verdict through fingerprint → payload → tripwire
// → ONE queue-lock critical section covering dedupe → caps → append →
// bound → ledger-record. A nil outcome means the signal was queued. The
// second result separates RETRYABLE failures (queue-write boundary: the
// item and the spool tail stay for the next drain) from DECIDED outcomes
// (deterministic refusals — withheld, deduped, capped, build refused —
// which are consumed, since re-deciding them changes nothing).
//
// The critical section is the review-gate serialization finding: the dedupe
// check and the ledger update used to bracket the queue mutation as
// separate steps, so two concurrent drains both read an empty ledger, both
// passed the dedupe window and the caps, and both enqueued the same
// fingerprint — the queue lock alone did not cover the ledger. The ledger
// now lives INSIDE the queue lock: its load, the checks, the record, and
// its save are one read-modify-write under the same cross-process lock as
// the append, and a ledger save failure aborts the whole mutation (the
// queue file stays unchanged — a signal whose ledger commit failed is
// never queued).
func drainMoai(ctx context.Context, entry bugreport.SpoolEntry) (*drainOutcome, bool) {
	fp := fingerprintOf(entry)

	// Build the validated payload from the spool's closed fields — pure,
	// lock-free work. An absent detail is absent — only a non-empty token
	// goes through the read-back validator (kinds whose register rows carry
	// no closed set accept no detail at all, which ParseDetail would
	// rightly refuse).
	var detail bugreport.Detail
	if entry.Detail != "" {
		var derr error
		detail, derr = bugreport.ParseDetail(entry.Kind, entry.Detail)
		if derr != nil {
			return &drainOutcome{outcome: "withheld", reason: "detail failed read-back validation: " + derr.Error(), fp: fp}, false
		}
	}
	versionID, commitID := entryIdentity(entry)
	payload, err := bugreport.Build(entry.Kind, entry.Frames, detail, bugreport.BuildIdentity{
		Version: versionID,
		Commit:  commitID,
	})
	if err != nil {
		return &drainOutcome{outcome: "dropped", reason: "payload build refused: " + err.Error(), fp: fp}, false
	}

	// The scrub tripwire: the rendered title and body must pass the existing
	// classifier and scrubber with zero findings — any finding means a
	// validator gap, and the payload is WITHHELD, never masked and sent
	// (REQ-ANON-012). The path-traversal token is always withheld.
	if reason, withheld := tripwire(payload, entry); withheld {
		return &drainOutcome{outcome: "withheld", reason: reason, fp: fp}, false
	}

	// The deterministic driver for the partial-consumption tests.
	if queueWriteBlockForTest != nil && queueWriteBlockForTest(entry) {
		return &drainOutcome{outcome: "dropped", reason: "queue write failed: (test contention)", fp: fp}, true
	}

	// ONE cross-process critical section: dedupe check → rolling caps →
	// append → queue bound → ledger record, all under the queue lock.
	store := BugreportQueueStore()
	title, body := RenderReport(payload)
	at := clock()
	stamp := at.UTC().Format(time.RFC3339)
	var outcome *drainOutcome
	var ledgerBroken bool
	var droppedIDs []string
	var queued feedback.QueueItem
	var orphanStamp string // set when this attempt ADOPTS an orphaned reservation (findings 5/9)
	err = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		ledger, lerr := loadLedger()
		if lerr != nil {
			outcome = &drainOutcome{outcome: "dropped", reason: "ledger unreadable: " + lerr.Error()}
			ledgerBroken = true
			return nil
		}

		// Per-fingerprint window (design section 10): no re-queue inside it.
		if !ledger.FingerprintAllowed(fp, clock(), config.DefaultBugreportFingerprintWindowDays) {
			// Orphan repair (review-gate residual, P2): the ledger record
			// and the queue replacement are two writes inside this one
			// section, and the record lands first — a crash or write
			// failure in between leaves a record whose report never
			// entered the queue, and the next drain used to judge the
			// report deduped and consume it. A record whose report lives
			// in NEITHER the queue nor the sent history is that orphan:
			// fall through and queue it (RecordQueued re-stamps below).
			if queueHasFingerprint(rec.Items, fp) || sentHistoryHasFingerprint(fp, clock(), config.DefaultBugreportFingerprintWindowDays) {
				outcome = &drainOutcome{outcome: "deduped", reason: "fingerprint already queued or sent inside the window"}
				return nil
			}
			// Terminal discard (review-gate finding 8, P2): a record whose
			// reservation ended at the send-attempt limit is FINISHED work
			// — without this check the record read exactly like the crash
			// orphan above and the recovery re-enrolled the discarded
			// report, resetting its attempts and re-spending the model
			// budget. Recovery re-queues only genuinely unfinished
			// reservations.
			if ledger.TerminallyDiscardedWithin(fp, clock(), config.DefaultBugreportFingerprintWindowDays) {
				outcome = &drainOutcome{outcome: "deduped", reason: "fingerprint was terminally discarded (send attempt limit)"}
				return nil
			}
			// The genuine orphan: ADOPT its reservation (review-gate
			// findings 5 and 9, P2) — the slot it already occupies belongs
			// to the report being re-queued. The caps judgment below
			// excludes the orphan's own entry (a report must not be
			// cap-judged against its own slot and consumed), and the record
			// below REPLACES the orphan's QueuedAt entry instead of
			// appending a second reservation for one report.
			orphanStamp = ledger.FingerprintSeen[fp]
		}

		// Rolling global caps. The adopted orphan slot counts for its
		// re-queued report, not against it.
		if capped, why := ledger.GlobalCapsAllowedExcluding(clock(), orphanStamp); !capped {
			outcome = &drainOutcome{outcome: "capped", reason: why}
			return nil
		}

		rec.LastSeq++
		queued = feedback.QueueItem{
			ID:          fmt.Sprintf("f%d", rec.LastSeq),
			Title:       title,
			Body:        body,
			QueuedAt:    stamp,
			Fingerprint: payload.Fingerprint,
			Kind:        string(payload.Kind),
		}
		rec.Items = append(rec.Items, queued)

		// The queue bound shares this mutation (review-gate finding:
		// EnforceQueueBound existed but nothing called it — repeated runs
		// grew the queue past the configured cap). Oldest dropped first,
		// each recorded once the mutation commits.
		for len(rec.Items) > config.DefaultBugreportQueueBound {
			oldest := rec.Items[0]
			rec.Items = rec.Items[1:]
			droppedIDs = append(droppedIDs, oldest.ID)
		}

		if orphanStamp != "" {
			ledger.RecordQueuedAdopting(payload.Fingerprint, at, orphanStamp)
		} else {
			ledger.RecordQueued(payload.Fingerprint, at)
		}
		if serr := saveLedger(ledger); serr != nil {
			// Aborting the callback leaves the queue file unchanged: a
			// signal whose ledger commit failed is never queued.
			return serr
		}
		if queueCommitFailForTest {
			return errors.New("outbox: (test) queue replace failed after the ledger record")
		}
		return nil
	})
	if err != nil {
		// EVERY queue-write failure is retryable — context expiry, a lock
		// budget lost to contention, a failed save: the item stays in the
		// spool for the next drain rather than being consumed as lost
		// (review-gate finding: the failed item used to leave with the
		// batch — a lost report).
		//
		// The ledger reservation the failed mutation recorded is rolled
		// back: a save that did not land must not count into the rolling
		// caps, or N retries reach the daily cap with an empty queue and
		// the next run classifies the report `capped` — and consumes it
		// (review-gate residual, P2). A crash between record and rollback
		// leaves the ORPHAN shape, which the dedupe check's recovery path
		// re-queues. A cancelled context skips the rollback — the orphan
		// recovery is the net for that shape too, and the rollback must
		// never extend the drain past its own deadline.
		if ctx.Err() == nil {
			rollbackLedgerRecord(ctx, fp, stamp)
		}
		return &drainOutcome{outcome: "dropped", reason: "queue write failed: " + err.Error(), fp: fp}, true
	}
	if outcome != nil {
		outcome.fp = fp
		return outcome, ledgerBroken
	}
	for _, id := range droppedIDs {
		_ = AppendOutbox(OutboxRow{
			Outcome: "dropped",
			Reason:  fmt.Sprintf("queue bound of %d reached; oldest item %s dropped", config.DefaultBugreportQueueBound, id),
		})
	}
	_ = AppendOutbox(OutboxRow{
		Outcome:  "queued",
		Title:    title,
		Body:     body,
		Fingerpr: payload.Fingerprint,
	})
	return nil, false
}

// tripwireInputForTest overrides what the tripwire screens (the withheld
// tests inject a credential-shaped body to kill the never-scrubs mutant);
// production leaves it nil and the real render is screened.
var tripwireInputForTest func() (title, body string)

// RequiresHold reports whether a payload's kind/detail combination carries
// a REQUIRED HOLD — a report the pipeline may never publish (REQ-ANON-012):
// the path-traversal sentinel report, whose publication could disclose the
// probed path shape. The drain's tripwire applies this against the spool
// entry, and the SENDER applies it against the revalidated payload — a
// tampered queue re-carrying the token must not publish what the pipeline
// is required to hold. The rule lives here so the two checkpoints cannot
// fork.
func RequiresHold(p bugreport.Payload) bool {
	if d, ok := p.Detail.(bugreport.TemplateToken); ok && d == bugreport.TokenPathTraversal {
		return true
	}
	return false
}

// tripwire runs the existing scrubber and classifier over the rendered
// report: a blocked verdict or ANY masking finding withholds the payload.
func tripwire(payload bugreport.Payload, entry bugreport.SpoolEntry) (reason string, withheld bool) {
	// The path-traversal token is always withheld (REQ-ANON-012): the report
	// itself derives from a path-traversal sentinel, and publishing it could
	// disclose the probed path shape. Held on the spool entry's reason AND
	// on the payload's own detail token (RequiresHold) — either carrier of
	// the sentinel triggers the hold.
	if entry.Reason == string(bugreport.ReasonPathTraversal) || RequiresHold(payload) {
		return "path_traversal reports are always withheld", true
	}

	title, body := RenderReport(payload)
	if tripwireInputForTest != nil {
		title, body = tripwireInputForTest()
	}

	res, err := feedback.Scrub(feedback.Input{Title: title, Body: body}, feedback.Options{
		// No ProjectRoot: the automatic path writes no project-tier mask log
		// — every bugreport store is user-scoped (D36).
	})
	if err != nil {
		return "scrub failed: " + err.Error(), true
	}
	if res.Verdict == feedback.VerdictBlocked {
		return "classifier blocked the report", true
	}
	if len(res.Findings) > 0 {
		return "masking findings — a validator gap; withheld, not masked", true
	}
	return "", false
}

// discardAll is the withdrawal branch (REQ-ANON-021): every unsent queue
// item and the spool are discarded, each recorded, the sent history kept.
// No network request and no model call — this package has neither. A spool
// removal that FAILS is returned, not swallowed (review gate finding, P2):
// withdrawal reporting success while untransmitted capture data survives on
// disk is a false success.
func discardAll(ctx context.Context) error {
	spoolHadContent := spoolFileNonEmpty()

	store := BugreportQueueStore()
	discarded := 0
	err := store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		discarded = len(rec.Items)
		rec.Items = []feedback.QueueItem{}
		return nil
	})
	if err != nil {
		return fmt.Errorf("outbox: discard queue: %w", err)
	}
	for i := 0; i < discarded; i++ {
		_ = AppendOutbox(OutboxRow{Outcome: "discarded", Reason: "participation off: unsent item discarded"})
	}
	if err := bugreport.ClearSpool(); err != nil {
		return fmt.Errorf("outbox: discard spool: %w", err)
	}
	if spoolHadContent {
		_ = AppendOutbox(OutboxRow{Outcome: "discarded", Reason: "participation off: capture spool discarded"})
	}
	return nil
}

// spoolFileNonEmpty reports whether the spool file still exists with
// content — read BEFORE ClearSpool erases it.
func spoolFileNonEmpty() bool {
	path, err := bugreport.SpoolPath()
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Size() > 0
}
