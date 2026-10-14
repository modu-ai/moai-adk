package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

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
func adoptResumedLeaderRun(root string, shouldAdopt bool, warn io.Writer) (adopted string, err error) {
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
	if err := db.AdoptRun(context.Background(), adopt, os.Getpid(), homestate.CurrentProcessFingerprint(), leaderResumeAdoptionBasis); err != nil {
		return "", err
	}
	_ = os.Setenv(config.EnvFactoryRunID, adopt.RunID)
	_ = os.Setenv(config.EnvFactoryLeadAddr, factory.FactoryLeaderSocketPath(adopt.RunID))
	return adopt.RunID, nil
}

// leaderEntryRunAdoption is the leader branch's resume re-adoption entry
// (card t1707), shared by the cc and glm leader branches — glm mirrors cc
// exactly except for the backend constant. It computes the call site's
// combined gate — a resume carrier at guard polarity AND a freshly minted id
// AND no explicit --factory-run — and re-adopts the prior run when it holds.
func leaderEntryRunAdoption(root string, args []string, leaderLabel, priorRunID, explicitRun string, warn io.Writer) (string, error) {
	resumeSeen, _ := scanResumeArgs(args, true)
	return adoptResumedLeaderRun(root, resumeSeen && explicitRun == "" && leaderRunIDWasMinted(leaderLabel, priorRunID), warn)
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
