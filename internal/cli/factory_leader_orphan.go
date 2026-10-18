package cli

// factory_leader_orphan.go — SPEC-FACTORY-LEADER-ADOPT-CORE-001 M3: the
// leader-start orphan refusal gate. When a `moai cc -f` / `moai glm -f`
// launch carries no resume option and no --factory-run, would publish a
// freshly minted run id, and the orphan finder returns at least one orphan
// run, the launcher refuses the launch: it writes one guidance block to
// stderr naming each orphan run and the deliberate adopt command of the
// launcher's own entry, and exits non-zero WITHOUT writing any factory run
// record or starting a session (REQ-FLAC-002, REQ-FLAC-005). The gate reads
// no stream, probes no terminal, and renders no prompt — the interactive
// chooser is a follow-up card (DP-3).

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Documentation anchors (AC-FLAC-010): the distributed guidance for the
// factory leader start and lane join — the factory skill and its local twin,
// the four docs-site factory-mode pages, and the four README dead-leader
// paragraphs — carries each of these two backticked spans exactly once, and
// M4's TestLeaderAdoptGuidanceDocsMatchAnchors builds its expectation from
// these same constants so the docs and the behavior cannot drift apart.
const (
	// LeaderStartOrphanAnchor states the refusal the orphan gate performs
	// (REQ-FLAC-002/005) and the deliberate adopt that gets past it.
	LeaderStartOrphanAnchor = "leader-start/orphan: refused; no run created; adopt with --factory-run <run-id>"
	// LaneJoinDeadOwnerAnchor states the lane-join preservation the retire
	// guard already ships (SPEC-FACTORY-RUN-RETIRE-001): a dead-owner run
	// that still holds in-progress cards or workers is NOT retired by a lane
	// join.
	LaneJoinDeadOwnerAnchor = "lane-join/dead-owner-with-cards-or-workers: kept"
)

// errLeaderStartOrphan is the refusal error the gate returns; the guidance
// block already carries the orphan facts and the adopt command on stderr, so
// the error stays one line.
var errLeaderStartOrphan = errors.New("factory leader start refused over an orphaned factory run")

// leaderStartOrphanHeader is the guidance block's header line; the
// acceptance criteria count its occurrences to prove exactly one block was
// rendered.
const leaderStartOrphanHeader = "factory leader start refused — orphaned factory run(s) still hold in-progress work:"

// kstZone is the render zone of REQ-FLAC-009: every run start time the
// guidance block renders is KST (UTC+9) regardless of the host's zone.
var kstZone = time.FixedZone("KST", 9*3600)

// kstTimeLayout is the clock layout of REQ-FLAC-009; the rendered value
// appends the " KST" suffix.
const kstTimeLayout = "2006-01-02 15:04:05"

// renderRunStartKST renders one run start time in KST as
// `YYYY-MM-DD HH:MM:SS KST`; a value that does not parse as an RFC 3339 time
// is rendered unchanged (REQ-FLAC-009).
func renderRunStartKST(value string) string {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return t.In(kstZone).Format(kstTimeLayout) + " KST"
}

// renderLeaderStartOrphanGuidance writes the one guidance block of
// REQ-FLAC-005: the header line, one fact line per orphan run (run id, owner
// classification, lead pid, start time in KST, non-terminal card count) with
// its deliberate adopt command naming the launcher's OWN entry command, and
// the closing line stating no run was created and no session was started.
// The block offers no retirement: no path this SPEC adds offers to retire a
// run (REQ-FLAC-007).
func renderLeaderStartOrphanGuidance(w io.Writer, entryCmd string, orphans []homestate.OrphanRun) {
	// The renderer's writes go to the caller's stderr buffer; a short write
	// cannot be recovered here, so the error is explicitly discarded.
	_, _ = fmt.Fprintln(w, leaderStartOrphanHeader)
	for _, o := range orphans {
		_, _ = fmt.Fprintf(w, "  run %s: owner %s, lead pid %d, started %s, %d in-progress card(s)\n",
			o.RunID, o.Classification, o.LeadPID, renderRunStartKST(o.CreatedAt), o.NonTerminalCards)
		_, _ = fmt.Fprintf(w, "  adopt this run deliberately: %s --factory-run %s\n", entryCmd, o.RunID)
	}
	_, _ = fmt.Fprintln(w, "no run was created and no session was started.")
}

// refuseLeaderStartOverOrphans is the leader branch's orphan gate
// (REQ-FLAC-002), shared by the cc and glm leader branches — the launcher's
// entry command is the only per-launcher argument (REQ-FLAC-008). The gate
// condition is the complement of every exemption of REQ-FLAC-006: a launch
// carrying an EXPLICIT resume option (strictResumeCarrierSeen — never the
// broader scanResumeArgs vocabulary), an explicit --factory-run, or a run id
// that was not freshly minted is exempt, and so is a launch the orphan
// finder answers with no orphan. Everything else — a fresh-mint leader start
// over an orphan — renders the guidance block to stderr and returns the
// refusal error, which the call sites return unwrapped for a non-zero exit
// BEFORE recordLeaderRunStart can write any run record (the refusal writes
// nothing and starts no session).
//
// @MX:ANCHOR: [AUTO] the one orphan-refusal gate both leader entries share; its condition is the exemption complement of REQ-FLAC-006
// @MX:REASON: widening or narrowing the condition changes which launches are refused over stranded card records — the over/under-refusal pair the acceptance criteria name as mutants (AC-FLAC-006)
func refuseLeaderStartOverOrphans(root, entryCmd string, args []string, leaderLabel, priorRunID, explicitRun string, stderr io.Writer) error {
	if strictResumeCarrierSeen(args) || explicitRun != "" || !leaderRunIDWasMinted(leaderLabel, priorRunID) {
		return nil
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	orphans, findErr := db.FindOrphanRuns(context.Background(), factorymsg.ReconcileOptionsFor(root))
	closeErr := db.Close()
	if findErr != nil {
		return findErr
	}
	if closeErr != nil {
		return closeErr
	}
	if len(orphans) == 0 {
		return nil
	}
	renderLeaderStartOrphanGuidance(stderr, entryCmd, orphans)
	return errLeaderStartOrphan
}
