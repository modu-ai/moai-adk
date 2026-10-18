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

// factory_run_adopt.go — the launcher-side half of the leader-resume run
// re-adoption (card t1707): when a RESUME-shaped leader start would mint a
// fresh run id because the id's only carrier (the launch environment) is gone,
// the launcher re-adopts the prior run instead, so every card record, worker
// registration, and lane binding the run carries survives the leader process
// change.

// leaderResumeAdoptionBasis is the basis the adoption records in the run's
// run.resumed event, so the audit trail names why the run id survived.
const leaderResumeAdoptionBasis = "leader-resume-adoption"

// leaderStartAdoptionBasis is the basis the DELIBERATE --factory-run adopt
// records in the run's run.resumed event — distinct from the resume
// adoption's leaderResumeAdoptionBasis, so the audit trail tells the two
// entries apart (SPEC-FACTORY-LEADER-ADOPT-CORE-001 plan A4).
const leaderStartAdoptionBasis = "leader-start-adoption"

// classifyNamedRun resolves the ONE named run's owner classification for the
// deliberate adopt. Production reads it through homestate.FindRunOwner with
// the production reconcile options; the owner-change test overrides it to
// restamp the row between classification and write (AC-FLAC-004(c), the
// competing writer inside the classify-to-write window).
var classifyNamedRun = func(ctx context.Context, root, runID string) (homestate.RunOwner, error) {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return homestate.RunOwner{}, err
	}
	defer closeFactoryInto(&err, db, "factory state")
	return db.FindRunOwner(ctx, factorymsg.ReconcileOptionsFor(root), runID)
}

// leaderRunIDWasMinted reports whether the run id the leader entry published
// was freshly minted — the condition under which resume re-adoption may
// replace it. An id the operator supplied explicitly (a pasted
// `leader-<run-id>` name, a still-set environment marker, or `--factory-run`)
// is deliberate and is never overridden; only the NewRunID() fallback step of
// leaderRunID mints.
func leaderRunIDWasMinted(leaderLabel, priorEnvRunID string) bool {
	if priorEnvRunID != "" {
		return false
	}
	if suffix, ok := factory.SplitLeaderLabel(leaderLabel); ok && suffix != "" && !allDigits(suffix) {
		return false
	}
	return true
}

// adoptResumedLeaderRun re-adopts the prior run on a resume-shaped leader
// start (card t1707). shouldAdopt is the call site's combined gate: the launch
// carries a resume carrier AND the published id was freshly minted AND no
// explicit --factory-run was given. On adoption the launcher republishes the
// adopted id AND its conventional address into the environment BEFORE the
// SessionStart notice and the messaging store open read them —
// enterFactoryLeaderMode stamped this launch's minted id into the address
// variable, and the notice must advertise the adopted run (sync-audit t1707
// F4) — which rebinds leader messaging to the adopted run (the third clause of
// the card fix).
//
// Zero candidates adopts nothing (a genuinely fresh start mints as before);
// anything the finder did not return as adoptable — several candidates, or a
// single candidate whose owner is only indeterminate (sync-audit t1707 F2) —
// adopts nothing and writes one warn line per candidate to warn naming the
// run, its owner classification, and the deliberate re-adoption command, the
// card's "refuse or warn and offer re-adoption". adopt "" with a nil error
// means "mint as before"; an error is a factory database failure and
// propagates, matching recordFactoryRunStart's own failure contract at the
// same call site.
func adoptResumedLeaderRun(root string, shouldAdopt bool, backend string, warn io.Writer) (adopted string, err error) {
	if !shouldAdopt {
		return "", nil
	}
	db, openErr := homestate.OpenFactory(root)
	if openErr != nil {
		return "", openErr
	}
	defer closeFactoryInto(&err, db, "factory state")
	adopt, candidates, err := db.FindResumableRun(context.Background(), factorymsg.ReconcileOptionsFor(root))
	if err != nil {
		return "", err
	}
	if adopt.RunID == "" {
		for _, c := range candidates {
			fmt.Fprintf(warn, "factory leader resume: run %s (owner %s) still holds live records — adopting nothing; to adopt one deliberately, relaunch with `moai cc -f --factory-run %s`\n", c.RunID, c.Classification, c.RunID)
		}
		return "", nil
	}
	// The automatic path shares the explicit path's cross-backend guard
	// (card-review round 2): a codex-led run is never adopted by a claude
	// leader, deliberately or automatically — it is warned with the same
	// re-adoption/retire offer instead of being taken over silently.
	if err := refuseCodexLeaderRun(root, adopt.RunID); err != nil {
		if !errors.Is(err, ErrRunLedByCodex) {
			return "", err
		}
		fmt.Fprintf(warn, "factory leader resume: run %s is led by codex — a claude leader does not adopt it; retire it first or relaunch the codex leader deliberately\n", adopt.RunID)
		return "", nil
	}
	// REQ-RNC-022 parity with the explicit path (codex_audit round 2): a run
	// holding a LIVE legacy peer record — a worker registered by a binary
	// before the leader/lane rename — is not adoptable, because the messaging
	// layer refuses to address that peer alongside this leader. The same
	// refusal enterSelectedFactoryRun applies, warned here rather than
	// failing the launch.
	if value, live, err := factorymsg.LiveLegacyPeer(context.Background(), root, adopt.RunID); err != nil {
		return "", err
	} else if live {
		fmt.Fprintf(warn, "factory leader resume: run %s holds live legacy record %q from a binary before the leader/lane rename — adopting nothing; end its sessions and retire with 'moai factory runs --retire %s' first\n", adopt.RunID, value, adopt.RunID)
		return "", nil
	}
	// The resume path passes no peer guard: its window contract is the
	// warn-and-mint outcome E3 pins (a live legacy peer at the T-check warns
	// and adopts nothing), and aborting inside the write would change that
	// contract. Its backend axis is covered by AdoptRun's extended
	// compare-and-set regardless (sync-audit round 3 F1').
	if err := db.AdoptRun(context.Background(), adopt, os.Getpid(), homestate.CurrentProcessFingerprint(), backend, leaderResumeAdoptionBasis, nil); err != nil {
		return "", err
	}
	_ = os.Setenv(config.EnvFactoryRunID, adopt.RunID)
	_ = os.Setenv(config.EnvFactoryLeadAddr, factory.FactoryLeaderSocketPath(adopt.RunID))
	return adopt.RunID, nil
}

// adoptExplicitLeaderRun takes the run named by --factory-run over through
// the same adoption write the resume adoption uses
// (SPEC-FACTORY-LEADER-ADOPT-CORE-001 REQ-FLAC-003): a compare-and-set on the
// owner identity the launch itself classified (whatever that classification
// is — the deliberate adopt classifies the run it is pointed at), one
// run.resumed event carrying the leader-start-adoption basis, no second run
// row, no run.started event, the run's lane capacity untouched, and the
// adopted id and its conventional address republished exactly as the resume
// path republishes them. backend is the adopting launcher's own backend; the
// adoption write lands it on the runs row in the same transaction as the
// stamp (plan-audit D1).
//
// The codex-led and live-legacy-peer refusals are re-applied INSIDE the
// classify-to-write window (sync-audit F1): enterSelectedFactoryRun applies
// both before the adopt, but the window is wide — a backend flip to codex,
// or a legacy-peer registration, can land between the fresh classification
// and the write, and the stamp the compare-and-set checks did not change, so
// a bare CAS would adopt a run its guards refuse. REQ-FLAC-004's abort is
// any of these refusals (or the compare-and-set miss — the owner changed
// between classification and write) returning unchanged — nothing minted,
// nothing modified; the first write of the launch is still
// recordLeaderRunStart, which this precedes.
func adoptExplicitLeaderRun(root, runID, backend string) (adopted string, err error) {
	owner, err := classifyNamedRun(context.Background(), root, runID)
	if err != nil {
		return "", err
	}
	// The classified-backend refusal (sync-audit round 4 F1''): a run that
	// is codex-led AT CLASSIFICATION is refused outright. The window
	// re-checks below read the row's CURRENT state, so an ABA round-trip —
	// codex at classification, claude at the re-checks, codex again before
	// the write — would pass them and match the classified codex value in
	// the compare-and-set; refusing the classified value here makes every
	// commit-time state provably the classified non-codex state, with the
	// CAS as the atomic backstop for post-read flips.
	if owner.Backend == BackendCodex {
		return "", codexLedRunRefusal(runID)
	}
	// The window re-checks: immediately-checked preconditions after the
	// fresh classification, before the write.
	if err = refuseCodexLeaderRun(root, runID); err != nil {
		return "", err
	}
	if err = refuseLiveLegacyPeerRun(root, runID); err != nil {
		return "", err
	}
	db, openErr := homestate.OpenFactory(root)
	if openErr != nil {
		return "", openErr
	}
	defer closeFactoryInto(&err, db, "factory state")
	// The peer guard closes the legacy axis inside the write's own
	// transaction (sync-audit round 3 F1'): a registration landing after the
	// re-checks above but before the commit rolls the adoption back. The
	// backend axis is atomic without it — AdoptRun's compare-and-set carries
	// the classified backend, so a flip landing after the re-checks fails the
	// write.
	if err = db.AdoptRun(context.Background(), owner, os.Getpid(), homestate.CurrentProcessFingerprint(), backend, leaderStartAdoptionBasis, func(ctx context.Context) error {
		return refuseLiveLegacyPeerRun(root, runID)
	}); err != nil {
		return "", err
	}
	_ = os.Setenv(config.EnvFactoryRunID, runID)
	_ = os.Setenv(config.EnvFactoryLeadAddr, factory.FactoryLeaderSocketPath(runID))
	return runID, nil
}

// strictResumeCarrierSeen reports whether the args carry an EXPLICIT resume
// option — the auto-adoption adjudication, deliberately narrower than
// scanResumeArgs's refusal-side detection (codex_audit t1707 round 3): the
// guard's isShortClusterCarryingR fires on an `r` anywhere inside a short
// cluster (`-derror`), which is the safe side for a REFUSAL but the wrong
// side for adoption — a false fire would hand a brand-new session an
// existing run's ownership and messaging. Adoption therefore judges only the
// unambiguous spellings of the same vocabulary: the standalone
// `--resume`/`-r` options, the equals form, and the attached `-r<value>`
// cluster whose first option IS the resume flag. A missed adoption is
// fail-safe since the live-records guard: the minted run simply joins the
// prior one in the AMBIGUOUS_FACTORY refusal instead of taking it over.
func strictResumeCarrierSeen(args []string) bool {
	separators := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == argSeparator {
			separators++
			if separators >= 2 {
				return false // prompt text: nothing after the second separator is an option
			}
			continue
		}
		if arg == resumeFlag || arg == resumeFlagShort {
			return true
		}
		if _, isEquals := resumeEqualsValue(arg); isEquals {
			return true
		}
		if strings.HasPrefix(arg, "-r") && arg != resumeFlagShort {
			return true // the attached-value short form `-r<session-id>`
		}
		// A required-value option consumes the next token, so a value that
		// merely looks like a resume option is never judged as one.
		if segmentClass(arg, separators == 0) == claudeOptionRequiredValue {
			if i+1 < len(args) && args[i+1] != argSeparator {
				i++
			}
		}
	}
	return false
}

// leaderEntryRunAdoption is the leader branch's resume re-adoption entry
// (card t1707), shared by the cc and glm leader branches — glm mirrors cc
// exactly except for the backend constant. It computes the call site's
// combined gate — an EXPLICIT resume carrier (strictResumeCarrierSeen) AND a
// freshly minted id AND no explicit --factory-run — and re-adopts the prior
// run when it holds. backend is the resuming launcher's backend, which the
// adoption write lands on the runs row (the row already names this leader as
// its owner — a same-direction truth).
func leaderEntryRunAdoption(root string, args []string, leaderLabel, priorRunID, explicitRun, backend string, warn io.Writer) (string, error) {
	resumeSeen := strictResumeCarrierSeen(args)
	return adoptResumedLeaderRun(root, resumeSeen && explicitRun == "" && leaderRunIDWasMinted(leaderLabel, priorRunID), backend, warn)
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
