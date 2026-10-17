package cli

// integration_base_freshness.go — the thin CLI adapters of the pre-integration
// base-freshness gate (SPEC-INTEGRATION-STALE-BASE-001, card t1834). The
// decision lives in factory.EvaluateBaseFreshness; this file only reads the
// configuration, maps a refusal onto the process exit code, and writes the one
// notice line a determinate-or-uncertain verdict may carry. Keeping the two
// verbs' edits to one call each (acquire, complete) is deliberate: both verbs
// are shared with neighbouring landing-surface work.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// baseFreshnessSeams is the gate's test seam; production leaves it zero-valued
// so the gate runs a real git in the candidate tree.
var baseFreshnessSeams factory.BaseFreshnessSeams

// baseFreshnessSettings is what the gate reads from workflow.base_freshness_gate.
// Err is set when the configuration could not be loaded: the gate then has no
// trustworthy K and admits with a notice.
type baseFreshnessSettings struct {
	Enabled bool
	K       int
	Err     error
}

// baseFreshnessGateConfig reads the gate's two keys from the project at root.
func baseFreshnessGateConfig(root string) baseFreshnessSettings {
	cfg, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil {
		return baseFreshnessSettings{Err: err}
	}
	if cfg == nil {
		return baseFreshnessSettings{Err: errors.New("the configuration did not load")}
	}
	gate := cfg.Workflow.BaseFreshnessGate
	return baseFreshnessSettings{Enabled: gate.Enabled, K: gate.MaxBehindLandings}
}

// checkBaseFreshness judges the candidate tree in dir for card against the
// local integration branch and renders the outcome: the one notice line (an
// uncertainty cause or the gate-off advisory) goes to notices, and an enabled
// refusal comes back as the exit-16 error whose text starts with the sentinel.
func checkBaseFreshness(notices io.Writer, root, dir, cardID, integrationBranch string) error {
	settings := baseFreshnessGateConfig(root)
	res := factory.EvaluateBaseFreshness(factory.BaseFreshnessInput{
		Root:              root,
		Dir:               dir,
		CardID:            cardID,
		IntegrationBranch: integrationBranch,
		K:                 settings.K,
		Enabled:           settings.Enabled,
		ConfigErr:         settings.Err,
	}, baseFreshnessSeams)
	if res.Notice != "" {
		_, _ = fmt.Fprintln(notices, res.Notice)
	}
	if !res.Admit {
		return &exitCodeError{code: factory.BaseFreshnessExitRefused, msg: res.Refusal}
	}
	return nil
}

// acquireBaseFreshnessPrecondition is `moai integration acquire`'s call: the
// candidate is the caller's working directory, as the remeasure verb keys it.
// An invocation without a card id is not the gate's business and stays silent
// (a leader acquiring for a batch).
func acquireBaseFreshnessPrecondition(cmd *cobra.Command, root, cardID, integrationBranch string) error {
	if cardID == "" {
		return nil
	}
	dir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "base-freshness: not measured — cannot resolve the working directory: %v\n", err)
		return nil
	}
	return checkBaseFreshness(cmd.ErrOrStderr(), root, dir, cardID, integrationBranch)
}

// completeBaseFreshnessPrecondition is `moai factory complete`'s call, made
// only where complete is about to acquire the window itself. A tip that is an
// adoption (REQ-MWQ-019 step 2) is exempt: the card is already merged through
// the merge verb, and its own merge commit is a first-parent landing that would
// otherwise read as staleness. The predicate is the one the adoption path uses,
// extracted into factoryAdoptedTip so the two cannot drift.
func completeBaseFreshnessPrecondition(out io.Writer, lockRoot, integTree, branch, cardTree, cardBranch, cardID string) error {
	if factoryAdoptedTip(lockRoot, integTree, branch, cardTree, cardBranch) != "" {
		return nil
	}
	return checkBaseFreshness(out, lockRoot, cardTree, cardID, branch)
}
