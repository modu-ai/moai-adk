package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factory_run_adopt.go — the launcher-side half of the leader-start run
// adoption. A leader restart is a new process, so the previous run's owner
// stamp is dead while its cards, worker registrations and lane bindings are
// alive (card t1707). The launcher therefore adopts the prior run instead of
// minting a fresh id over it, and — since card t1963, SPEC-FACTORY-COMPLETION-
// RECOVERY-001 REQ-FCR-017/018 — it does so without a resume argument and
// folds every OTHER dead-owner run into the adopted one, so a leader start
// leaves one active run instead of one more zombie.

// leaderResumeAdoptionBasis is the basis the adoption records in the run's
// run.resumed event, so the audit trail names why the run id survived.
const leaderResumeAdoptionBasis = "leader-resume-adoption"

// leaderAdoptOptions resolves the reconcile options the leader-entry adoption
// classifies under. A seam so a test can pin the boot proof's host-boot time;
// production is the shared option set every retirement path passes.
var leaderAdoptOptions = factorymsg.ReconcileOptionsFor

// leaderAdoptBeforeAdopt is a test-only seam between the classification and
// the adoption compare-and-set: nil in production, a test uses it to change an
// owner stamp (a second leader winning the race) or to break the database
// handle in that window.
var leaderAdoptBeforeAdopt func(db *homestate.FactoryDB)

// leaderRunIDWasMinted reports whether the run id the leader entry published
// was freshly minted — the condition under which adoption may replace it. An
// id the operator supplied explicitly (a pasted `leader-<run-id>` name, a
// still-set environment marker, or `--factory-run`) is deliberate and is never
// overridden; only the NewRunID() fallback step of leaderRunID mints.
func leaderRunIDWasMinted(leaderLabel, priorEnvRunID string) bool {
	if priorEnvRunID != "" {
		return false
	}
	if suffix, ok := factory.SplitLeaderLabel(leaderLabel); ok && suffix != "" && !allDigits(suffix) {
		return false
	}
	return true
}

// adoptResumedLeaderRun adopts the newest dead-owner run on a leader start and
// consolidates the others into it. shouldAdopt is the call site's gate: the
// published id was freshly minted AND no explicit --factory-run was given.
// On adoption the launcher republishes the adopted id AND its conventional
// address into the environment BEFORE the SessionStart notice and the
// messaging store open read them — enterFactoryLeaderMode stamped this
// launch's minted id into the address variable, and the notice must advertise
// the adopted run (sync-audit t1707 F4).
//
// Nothing is adopted — the minted id stands and one warn line per candidate
// names the run, its owner classification and the deliberate --factory-run
// command — when no run holds live records, when any active run classifies
// live or indeterminate (it may be a live leader), when no candidate is dead on
// its own stamp, and when the newest candidate is led by codex or holds a live
// legacy peer record: the newest is the only run considered, so which run a
// start resumes never depends on which one happened to be rejected.
//
// The adoption is the owner-stamp compare-and-set. A leader that loses it
// does not fail: it keeps the minted id, writes one warn line naming the
// contested run, and consolidates nothing. Any other adoption error is a
// database failure and propagates, matching recordFactoryRunStart's own
// failure contract at the same call site.
//
// @MX:WARN: [AUTO] decides which run a restarted leader owns and which dead runs are retired behind it
// @MX:REASON: adopting beside a live or unknown leader would overwrite its stamp and move its cards (t1707 F2); every guard here fails toward adopting nothing
func adoptResumedLeaderRun(root string, shouldAdopt bool, warn io.Writer) (adopted string, err error) {
	if !shouldAdopt {
		return "", nil
	}
	db, openErr := homestate.OpenFactory(root)
	if openErr != nil {
		return "", openErr
	}
	defer closeFactoryInto(&err, db, "factory state")
	ctx := context.Background()
	opts := leaderAdoptOptions(root)
	plan, err := db.PlanLeaderEntry(ctx, opts)
	if err != nil {
		return "", err
	}
	if plan.Blocked || len(plan.Stamped) == 0 {
		warnLeaderCandidates(warn, plan.Candidates, "")
		return "", nil
	}
	adopt := plan.Stamped[0]
	refusal, err := leaderAdoptionRefusal(root, adopt.RunID)
	if err != nil {
		return "", err
	}
	if refusal != "" {
		_, _ = fmt.Fprintln(warn, refusal)
		warnLeaderCandidates(warn, plan.Candidates, adopt.RunID)
		return "", nil
	}
	if leaderAdoptBeforeAdopt != nil {
		leaderAdoptBeforeAdopt(db)
	}
	start := homestate.CurrentProcessFingerprint()
	if err := db.AdoptRun(ctx, adopt, os.Getpid(), start, leaderResumeAdoptionBasis); err != nil {
		if errors.Is(err, homestate.ErrAdoptStampChanged) {
			_, _ = fmt.Fprintf(warn, "factory leader resume: run %s was adopted by another leader first — this leader starts a fresh run and consolidates nothing\n", adopt.RunID)
			return "", nil
		}
		return "", err
	}
	_ = os.Setenv(config.EnvFactoryRunID, adopt.RunID)
	_ = os.Setenv(config.EnvFactoryLeadAddr, factory.FactoryLeaderSocketPath(adopt.RunID))
	target := homestate.RunOwner{
		RunID: adopt.RunID, Status: "active", LeadPID: os.Getpid(), LeadProcessStart: start,
		Classification: homestate.OwnerLive, Basis: homestate.BasisStamp,
	}
	consolidateDeadRuns(ctx, db, root, plan, target, opts, warn)
	return adopt.RunID, nil
}

// warnLeaderCandidates writes the adopt-nothing warning, one line per
// candidate, skipping the run (if any) whose own refusal line was written.
func warnLeaderCandidates(warn io.Writer, candidates []homestate.RunOwner, skip string) {
	for _, c := range candidates {
		if c.RunID == skip {
			continue
		}
		_, _ = fmt.Fprintf(warn, "factory leader resume: run %s (owner %s) still holds live records — adopting nothing; to adopt one deliberately, relaunch with `moai cc -f --factory-run %s`\n", c.RunID, c.Classification, c.RunID)
	}
}

// leaderAdoptionRefusal reports why a claude leader may not take over the run:
// the automatic path shares the explicit path's cross-backend guard (card-review
// round 2) and the legacy-record guard (REQ-RNC-022). It returns the warn line,
// or "" when the run is adoptable; an error is a database failure.
func leaderAdoptionRefusal(root, runID string) (string, error) {
	if err := refuseCodexLeaderRun(root, runID); err != nil {
		if !errors.Is(err, ErrRunLedByCodex) {
			return "", err
		}
		return fmt.Sprintf("factory leader resume: run %s is led by codex — a claude leader does not adopt it; retire it first or relaunch the codex leader deliberately", runID), nil
	}
	value, live, err := factorymsg.LiveLegacyPeer(context.Background(), root, runID)
	if err != nil {
		return "", err
	}
	if live {
		return fmt.Sprintf("factory leader resume: run %s holds live legacy record %q from a binary before the leader/lane rename — adopting nothing; end its sessions and retire with 'moai factory runs --retire %s' first", runID, value, runID), nil
	}
	return "", nil
}

// consolidateDeadRuns folds every other dead-stamp run into the run just
// adopted, one consolidation per source, and names on stderr what moved. A
// source the filter excludes — led by codex, holding a live legacy peer, or
// proved dead only through its peer or the boot proof — or that the operation
// refuses stays active with its cards untouched, and one line says why. The
// adoption already happened, so a failure here is reported and the launch goes
// on: the next leader start finds the run still in the plan.
func consolidateDeadRuns(ctx context.Context, db *homestate.FactoryDB, root string, plan homestate.LeaderEntryPlan, target homestate.RunOwner, opts homestate.ReconcileOptions, warn io.Writer) {
	stamped := make(map[string]bool, len(plan.Stamped))
	for _, s := range plan.Stamped {
		stamped[s.RunID] = true
	}
	stays := func(runID, format string, args ...any) {
		_, _ = fmt.Fprintf(warn, "factory leader resume: run %s stays active — %s\n", runID, fmt.Sprintf(format, args...))
	}
	for _, src := range plan.Stamped {
		if src.RunID == target.RunID {
			continue
		}
		refusal, err := leaderAdoptionRefusal(root, src.RunID)
		if err != nil {
			stays(src.RunID, "%v", err)
			continue
		}
		if refusal != "" {
			stays(src.RunID, "%s", strings.TrimPrefix(refusal, "factory leader resume: "))
			continue
		}
		res, err := db.RebindRun(ctx, homestate.RebindRequest{Source: src, Target: target, RequireStampBasis: true}, opts)
		if err != nil {
			stays(src.RunID, "the consolidation was refused: %v", err)
			continue
		}
		_, _ = fmt.Fprintf(warn, "factory leader resume: consolidated run %s into %s — moved cards: %s; abandoned cards: %s\n",
			src.RunID, target.RunID, joinOrNone(res.Moved), joinOrNone(res.Abandoned))
	}
	for _, c := range plan.Candidates {
		if stamped[c.RunID] {
			continue
		}
		stays(c.RunID, "its owner was proved dead through the %s proof, not its own stamp; retire it deliberately with `moai factory runs --retire %s`", c.Basis, c.RunID)
	}
}

func joinOrNone(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	return strings.Join(ids, ", ")
}

// leaderEntryRunAdoption is the leader branch's adoption entry (card t1707,
// widened by card t1963), shared by the cc and glm leader branches — glm
// mirrors cc exactly except for the backend constant. The gate is a freshly
// minted id and no explicit --factory-run; the launch arguments no longer
// matter (decision Q2: no --resume is needed), and are accepted only so the
// call sites hand over the same launch view they always did.
func leaderEntryRunAdoption(root string, _ []string, leaderLabel, priorRunID, explicitRun string, warn io.Writer) (string, error) {
	return adoptResumedLeaderRun(root, explicitRun == "" && leaderRunIDWasMinted(leaderLabel, priorRunID), warn)
}

// recordLeaderRunStart records the leader entry's run start: on the adopted
// path only the runtime facts are re-recorded under the adopted id — the
// adopted run keeps its row and its run.resumed event, so no second
// run.started lands on it — otherwise the full recordFactoryRunStart runs and
// records the start under the published (minted) id.
func recordLeaderRunStart(root, adopted, backend, specID string, declaredLanes int) error {
	if adopted != "" {
		if err := factory.RecordFactoryRunStart(root, adopted, backend, specID); err != nil {
			return fmt.Errorf("record factory run: %w", err)
		}
		return nil
	}
	if err := recordFactoryRunStart(root, os.Getenv(config.EnvFactoryRunID), backend, specID, declaredLanes); err != nil {
		return fmt.Errorf("record factory run: %w", err)
	}
	return nil
}
