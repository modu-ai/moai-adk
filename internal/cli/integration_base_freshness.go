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
	"gopkg.in/yaml.v3"

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
	dotMoai := filepath.Join(root, ".moai")
	loader := config.NewLoader()
	cfg, err := loader.Load(dotMoai)
	if err != nil {
		// The loader's text can quote a rejected value; name the file and kind.
		return baseFreshnessSettings{Err: errors.New("workflow.yaml was rejected by validation")}
	}
	if cfg == nil {
		return baseFreshnessSettings{Err: errors.New("the configuration did not load")}
	}
	// The loader turns a workflow.yaml read or parse failure into a warning plus
	// defaults and returns no error, so a broken file would read as "gate off".
	// A verdict on defaults nobody chose is no verdict: report the failure.
	if cause := workflowConfigFailure(dotMoai, loader.LoadedSections()); cause != nil {
		return baseFreshnessSettings{Err: cause}
	}
	gate := cfg.Workflow.BaseFreshnessGate
	return baseFreshnessSettings{Enabled: gate.Enabled, K: gate.MaxBehindLandings}
}

// workflowConfigFailure reports why workflow.yaml, present on disk, did not
// load: nil when the file is absent (defaults are the intended state) or the
// loader marked the workflow section loaded.
func workflowConfigFailure(dotMoai string, loaded map[string]bool) error {
	data, err := os.ReadFile(filepath.Join(dotMoai, "config", "sections", "workflow.yaml"))
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("read workflow.yaml: %w", err)
	case loaded["workflow"]:
		return nil
	}
	// Fixed, input-free causes only: a parser error quotes the offending value.
	var probe map[string]any
	if yaml.Unmarshal(data, &probe) != nil {
		return errors.New("parse workflow.yaml: invalid YAML syntax")
	}
	return errors.New("parse workflow.yaml: a value has the wrong type")
}

// checkBaseFreshness judges the candidate tree in dir for card against the
// local integration branch and renders the outcome: the one notice line (an
// uncertainty cause or the gate-off advisory) goes to notices, and an enabled
// refusal comes back as the exit-16 error whose text starts with the sentinel.
func checkBaseFreshness(notices io.Writer, root, dir, cardID, integrationBranch string, explicitBranch bool) error {
	settings := baseFreshnessGateConfig(root)
	res := factory.EvaluateBaseFreshness(factory.BaseFreshnessInput{
		Root:              root,
		Dir:               dir,
		CardID:            cardID,
		IntegrationBranch: integrationBranch,
		ExplicitBranch:    explicitBranch,
		K:                 settings.K,
		Enabled:           settings.Enabled,
		ConfigErr:         settings.Err,
	}, baseFreshnessSeams)
	if res.Notice != "" {
		_, _ = fmt.Fprintln(notices, res.Notice)
	}
	if !res.Admit {
		// The root error handler renders nothing for an exit-coded error, so the
		// verb writes the sentinel line itself (the idiom auditPlanFail uses);
		// without it the refusal exits 16 with no output at all.
		_, _ = fmt.Fprintln(notices, res.Refusal)
		return &exitCodeError{code: factory.BaseFreshnessExitRefused, msg: res.Refusal}
	}
	return nil
}

// acquireBaseFreshnessPrecondition is `moai integration acquire`'s call: the
// candidate is the caller's working directory, as the remeasure verb keys it.
// An invocation without a card id is not the gate's business and stays silent
// (a leader acquiring for a batch). explicitBranch is true when the target came
// from --branch, so the refusal's remedy names remeasure --branch.
func acquireBaseFreshnessPrecondition(cmd *cobra.Command, root, cardID, integrationBranch string, explicitBranch bool) error {
	if cardID == "" {
		return nil
	}
	dir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "base-freshness: not measured — cannot resolve the working directory")
		return nil
	}
	return checkBaseFreshness(cmd.ErrOrStderr(), root, dir, cardID, integrationBranch, explicitBranch)
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
	return checkBaseFreshness(out, lockRoot, cardTree, cardID, branch, false)
}
