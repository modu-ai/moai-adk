package cli

// update_mirror_heal.go — SPEC-UPDATE-MIRROR-HEAL-001, the update-path skill
// mirror repair (REQ-UMH-001).
//
// A version-matched `moai update` returns before Deploy
// (update_template_sync.go, version-match branch), and both .agents/skills
// producers live inside Deploy — so a deleted mirror is permanent for that
// project. This pass runs BESIDE that early return, at the same call site as
// refreshCodexWiringBestEffort: the optimization is preserved unchanged
// (C-2), and the repair no longer depends on a template redeploy.
//
// The existence gate is the project's RECORDED DEPLOY VERSION, not the
// presence of .agents/ (gone exactly when repair is needed) and not the
// presence of .codex/ wiring (absent by default in every `--llm claude`
// project, whose mirror is created unconditionally). See spec.md §3 for both
// rejections, each measured.

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/template"
)

// mirrorRepairGateOpen reports whether a project stamped with the given
// template_version is one whose deploy creates a skill mirror.
//
// It reuses compareVersionLoose, so a `v` prefix is stripped and a
// pre-release suffix is truncated before the numeric comparison; a stamp that
// parses to nothing (an unset field, an unparseable value) degrades to
// [0,0,0] and closes the gate. Closed is the safe default: it makes the pass
// inert outside a MoAI project and in the pre-feature population.
func mirrorRepairGateOpen(templateVersion string) bool {
	return compareVersionLoose(templateVersion, template.MirrorIntroducedVersion) >= 0
}

// repairSkillMirrorBestEffortAt repairs the .agents/skills mirror of the
// project at projectRoot. Best-effort by contract (C-5): every failure warns
// to errOut and the update continues — losing the update over a
// Codex-visibility convenience would be a wildly disproportionate reaction,
// which is the same stance skill_mirror.go takes at deploy time.
func repairSkillMirrorBestEffortAt(projectRoot string, out, errOut io.Writer) {
	stamp, err := plan.GetProjectConfigVersion(projectRoot)
	if err != nil {
		warnMirrorRepair(errOut, fmt.Sprintf("cannot read the project's template_version: %v", err))
		return
	}
	if !mirrorRepairGateOpen(stamp) {
		// Pre-feature project (or not a MoAI project at all): creating a
		// mirror here would be provisioning, not repair (REQ-UMH-002 / C-1).
		return
	}

	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		warnMirrorRepair(errOut, fmt.Sprintf("cannot load embedded templates: %v", err))
		return
	}

	res := template.RepairSkillMirror(fsys, projectRoot)
	for _, w := range res.Warnings {
		warnMirrorRepair(errOut, w)
	}
	if res.Changed() && out != nil {
		// Write error intentionally ignored (nolint:errcheck rationale): a
		// failed write to the progress stream must not turn a successful
		// repair into an update failure (C-5). golangci-lint runs errcheck
		// with check-blank:false, so this blank-assign form is the project's
		// required idiom for a deliberately-unchecked console write.
		_, _ = fmt.Fprintf(out, "Restored .agents/skills: %d mirror %s, %d published %s\n",
			len(res.PathACreated), pluralMirrorEntries(len(res.PathACreated)),
			len(res.PublishedRestored), pluralMirrorFiles(len(res.PublishedRestored)))
	}
}

// repairSkillMirrorBestEffort is the runUpdate call-site form: runUpdate
// operates on the current working directory (".").
func repairSkillMirrorBestEffort(out, errOut io.Writer) {
	repairSkillMirrorBestEffortAt(".", out, errOut)
}

func warnMirrorRepair(errOut io.Writer, msg string) {
	if errOut == nil {
		return
	}
	// Write error intentionally ignored (nolint:errcheck rationale): this is
	// the fail-open warning path itself, so a failure to report a failure has
	// nowhere left to escalate.
	_, _ = fmt.Fprintf(errOut, "warning: skill mirror repair: %s\n", msg)
}

func pluralMirrorEntries(n int) string {
	if n == 1 {
		return "entry"
	}
	return "entries"
}

func pluralMirrorFiles(n int) string {
	if n == 1 {
		return "file"
	}
	return "files"
}
