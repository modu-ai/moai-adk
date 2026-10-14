package cli

import (
	"context"
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
// explicit --factory-run was given. On adoption the launcher's run-id
// environment is rewritten to the adopted id BEFORE the SessionStart notice
// and the messaging store open read it, which rebinds leader messaging to the
// adopted run (the third clause of the card fix).
//
// Zero candidates adopts nothing (a genuinely fresh start mints as before);
// several candidates adopts nothing and writes one warn line per candidate to
// warn naming the run, its owner classification, and the deliberate
// re-adoption command — the card's "refuse or warn and offer re-adoption".
// adopt "" with a nil error means "mint as before"; an error is a factory
// database failure and propagates, matching recordFactoryRunStart's own
// failure contract at the same call site.
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
	switch {
	case len(candidates) == 1:
		if err := db.AdoptRun(context.Background(), adopt.RunID, os.Getpid(), homestate.CurrentProcessFingerprint(), leaderResumeAdoptionBasis); err != nil {
			return "", err
		}
		_ = os.Setenv(config.EnvFactoryRunID, adopt.RunID)
		return adopt.RunID, nil
	case len(candidates) > 1:
		for _, c := range candidates {
			fmt.Fprintf(warn, "factory leader resume: run %s (owner %s) still holds live records — adopting nothing; to adopt one deliberately, relaunch with `moai cc -f --factory-run %s`\n", c.RunID, c.Classification, c.RunID)
		}
		return "", nil
	default:
		return "", nil
	}
}
