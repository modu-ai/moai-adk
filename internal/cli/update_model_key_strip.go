package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// update_model_key_strip.go — the moai update strip step for the retired
// per-agent model/effort keys (llm.yaml profile/profiles/harness_agents/
// agent_overrides/performance_tier, workflow.yaml workflow_agents/
// model_routing/model_routing_profiles/agent_model_guard). Subagents inherit
// the main session's model and effort, so these keys have no reader; the 3-way
// merge would retain them forever (old-only keys are kept), hence an explicit
// strip. It runs at three hosts, always after a configuration backup exists:
// (a) the template-sync "Restore Settings" step, (b) a version-matched update
// that skipped the sync, (c) the clean reinstall.

// confirmViaPreviewFn is the merge-confirmation seam runTemplateSyncWithProgress
// calls. Tests replace it to answer "cancel" without a TTY, which is the only
// way to reach the user-cancel return that host (b) must leave alone.
var confirmViaPreviewFn = confirmViaPreview

// shippedRetiredModelKeysFn names the retired keys the embedded template still
// ships; the strip leaves those alone (template.ShippedRetiredModelKeys). Tests
// replace it to exercise the state after the template drops the keys.
var shippedRetiredModelKeysFn = template.ShippedRetiredModelKeys

// retiredModelKeyReason is the one-line explanation printed with each removal.
const retiredModelKeyReason = "subagents now inherit the main session's model and effort"

// stripRetiredModelConfig removes the retired keys from projectRoot and prints
// one line per removed key, naming it as removed. backupPath is where the
// original values are kept; it is named on a line whose key carried user
// values. A strip error is returned without having printed a partial report
// line for the failed file.
func stripRetiredModelConfig(out io.Writer, projectRoot, backupPath string) ([]template.RetiredModelKey, error) {
	shipped, err := shippedRetiredModelKeysFn()
	if err != nil {
		return nil, err
	}
	removed, err := template.StripRetiredModelKeys(projectRoot, shipped)
	for _, k := range removed {
		line := fmt.Sprintf("  %s Removed %s from %s (%s)", uikit.SymSuccess(), k.Key, k.Section, retiredModelKeyReason)
		if k.UserValues {
			line += fmt.Sprintf(" — it carried user values; the original is kept in %s", backupPath)
		}
		_, _ = fmt.Fprintln(out, line)
	}
	return removed, err
}

// withoutStrippedKeys drops from refs every retained-key ref that names a
// removed key or a key beneath it, so the retained-key advisory never reports
// a stripped key as retained.
func withoutStrippedKeys(refs []backup.RetainedKeyRef, removed []template.RetiredModelKey) []backup.RetainedKeyRef {
	if len(removed) == 0 {
		return refs
	}
	out := make([]backup.RetainedKeyRef, 0, len(refs))
	for _, ref := range refs {
		stripped := false
		for _, k := range removed {
			if ref.Section == k.Section && (ref.Key == k.Key || strings.HasPrefix(ref.Key, k.Key+".")) {
				stripped = true
				break
			}
		}
		if !stripped {
			out = append(out, ref)
		}
	}
	return out
}

// updateSkippedOnVersionMatch re-evaluates the version-match predicate
// runTemplateSyncWithProgress uses, with the same inputs. It is how host (b)
// tells a version-matched skip from a user-cancelled merge: both return
// (true, nil), and that return shape is pinned (REQ-UMH-003).
func updateSkippedOnVersionMatch(cmd *cobra.Command, projectRoot string) bool {
	forceUpdate := getBoolFlag(cmd, "force")
	packageVersion := version.GetVersion()
	projectVersion, verr := plan.GetProjectConfigVersion(projectRoot)
	return verr == nil && packageVersion == projectVersion && !forceUpdate
}

// stripRetiredModelConfigOnVersionMatch is host (b): a version-matched update
// runs no sync, no backup, and no merge, so it takes its own configuration
// backup — only when a retired key is actually present, so a clean project
// gains no backup directory — and then strips. It does nothing after a
// user-cancelled merge.
func stripRetiredModelConfigOnVersionMatch(cmd *cobra.Command, out io.Writer, projectRoot string) error {
	if !updateSkippedOnVersionMatch(cmd, projectRoot) {
		return nil
	}
	shipped, err := shippedRetiredModelKeysFn()
	if err != nil {
		return err
	}
	present, err := template.RetiredModelKeysPresent(projectRoot, shipped)
	if err != nil || !present {
		return err
	}
	backupPath, err := backup.BackupMoaiConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("back up .moai/config before removing retired model keys: %w", err)
	}
	if _, err := stripRetiredModelConfig(out, projectRoot, backupPath); err != nil {
		return fmt.Errorf("remove retired model keys: %w", err)
	}
	return nil
}

// writeRetainedKeyAdvisoryLines prints the legacy per-key retained-key
// advisory for refs (KeptOverDefault refs have no legacy line). The text is
// the one backup.RestoreMoaiConfig prints; the clean reinstall keeps that
// format while filtering out the keys the strip removed.
func writeRetainedKeyAdvisoryLines(w io.Writer, refs []backup.RetainedKeyRef) {
	for _, ref := range refs {
		if ref.KeptOverDefault {
			continue
		}
		_, _ = fmt.Fprintf(w, "advisory: retained key %q absent from new template (preserved from user config)\n", ref.Key)
	}
}
