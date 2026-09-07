// skill_mirror_repair.go — SPEC-UPDATE-MIRROR-HEAL-001, the deploy-less
// mirror repair pass.
//
// Both .agents/skills producers live inside Deploy: the Path A symlink mirror
// (mirrorSkills, called as Deploy's last step) and the Path B published
// SKILL.md files (ordinary template writes). A version-matched `moai update`
// returns before Deploy, so a deleted mirror is permanent for that project.
// This file supplies the one seam that restores both WITHOUT a deploy.
//
// Seam shape, decided rather than discovered (plan.md M3 run-phase note 2):
// this is a package-level function, NOT a DeployerOption. A deployer option
// would also be live on the deploy-time path, where Deploy already runs the
// producer — so choosing that shape would change deploy behavior as a side
// effect of a repair feature. The blast radius of the shape below is the
// repair pass alone; Deploy is untouched.
//
// The pass reuses mirrorOneSkill for the per-entry semantics, so the
// non-symlink-occupancy skip and the idempotent already-correct branch come
// for free rather than being reimplemented (and cannot drift from the
// producer's).
package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MirrorIntroducedVersion is the first released version whose deploy creates
// the .agents/skills mirror. It is the existence gate for the repair pass: a
// project whose recorded template_version is below it was last deployed by a
// binary with no mirror step, so restoring a mirror there would be CREATION
// for a population the feature never targeted, not repair.
//
// The value is fixed from the release record, not guessed:
//   - CHANGELOG.md, section "[3.1.3] - 2026-08-24" — "A skill mirror at
//     .agents/skills for codex-cli, derived from the run rather than
//     hand-listed".
//   - skill_mirror.go, the producer this constant describes, was added in
//     commit 9c94c6b7a (2026-08-22), i.e. before that release and after
//     3.1.2.
//
// The constant lives here, next to the producer whose behavior it describes,
// rather than next to the gate that reads it.
const MirrorIntroducedVersion = "3.1.3"

// MirrorRepairResult is the observable outcome of one repair pass. Like the
// producer it wraps, the pass never returns an error: everything it could not
// do lands here as a warning and the update continues (REQ-UMH-006 / C-5).
type MirrorRepairResult struct {
	// SkillMirrors holds one entry per Path A candidate the pass attempted.
	SkillMirrors []SkillMirrorEntry
	// PathACreated names the skills whose mirror entry was ABSENT before the
	// pass ran — the subset the pass actually created. Distinguishing this
	// from SkillMirrors is what keeps a healthy project's run silent.
	PathACreated []string
	// PublishedRestored holds the deploy-relative paths of Path B artifacts
	// this pass wrote back.
	PublishedRestored []string
	// Warnings carries every failure, in encounter order.
	Warnings []string
}

// Changed reports whether the pass altered the project.
func (r *MirrorRepairResult) Changed() bool {
	if r == nil {
		return false
	}
	return len(r.PathACreated) > 0 || len(r.PublishedRestored) > 0
}

// PublishedSkillNames returns the published command-skill directory names in
// sorted order. It is the production set — callers that need to enumerate the
// Path B artifacts derive them from here rather than hand-listing them.
func PublishedSkillNames() []string {
	names := make([]string, 0, len(publishedSkillNames))
	for name := range publishedSkillNames {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// @MX:NOTE: [AUTO] The only deploy-less writer into .agents/; every branch is bound by a SPEC-UPDATE-MIRROR-HEAL-001 acceptance criterion.
//
// RepairSkillMirror restores the .agents/skills mirror of the project at
// projectRoot from the template FS, without running a deploy.
//
// It restores BOTH producers (REQ-UMH-004):
//
//   - Path A: one symlink entry per TEMPLATE-SHIPPED skill name that is
//     actually present under .claude/skills. The candidate names come from
//     the template FS, never from a listing of .claude/skills — a listing
//     would readmit locally-authored skills no deploy produced (REQ-UMH-007).
//     The candidates are then filtered by canonical-target existence, because
//     os.Symlink succeeds against a missing target and an entry that does not
//     resolve is damage rather than repair — it is a `moai doctor` finding in
//     its own right (REQ-UMH-010).
//
//   - Path B: the published .agents/skills/moai-<command>/SKILL.md files,
//     restore-missing-only. A file already at a published-skill path is left
//     exactly as it is: update mode refreshes template-managed content, but
//     it must never overwrite a file it does not own at these paths (the same
//     rule the deploy walk encodes via ProtectedSkips).
//
// The caller decides what the user sees; this function never prints.
//
// The existence gate (MirrorIntroducedVersion) is NOT applied here — it is
// the caller's, because the recorded project version is a CLI-layer concern.
func RepairSkillMirror(fsys fs.FS, projectRoot string) *MirrorRepairResult {
	res := &MirrorRepairResult{}

	canonical, published, err := mirrorRepairCandidates(fsys)
	if err != nil {
		res.Warnings = append(res.Warnings,
			fmt.Sprintf("skill mirror repair: cannot enumerate template skills: %v", err))
		return res
	}

	res.repairPathA(projectRoot, canonical)
	res.repairPathB(fsys, projectRoot, published)
	return res
}

// repairPathA restores the symlink mirror for the candidates whose canonical
// directory exists.
func (r *MirrorRepairResult) repairPathA(projectRoot string, candidates []string) {
	mirrorDir := filepath.Join(projectRoot, mirrorSkillsRelDir)

	var present []string
	for _, name := range candidates {
		// REQ-UMH-010: the filter sits HERE, above the shared producer, not
		// inside it — the producer is also the deploy-time path, which has a
		// deploy's own guarantees about what it just wrote.
		info, err := os.Stat(filepath.Join(projectRoot, ".claude", "skills", name))
		if err != nil || !info.IsDir() {
			continue
		}
		present = append(present, name)
		if _, lerr := os.Lstat(filepath.Join(mirrorDir, name)); lerr != nil {
			r.PathACreated = append(r.PathACreated, name)
		}
	}
	if len(present) == 0 {
		return
	}

	d := &deployer{}
	r.SkillMirrors = d.mirrorSkills(projectRoot, present)
	for _, e := range r.SkillMirrors {
		if e.Warning != "" {
			r.Warnings = append(r.Warnings, e.Warning)
		}
	}
}

// repairPathB rewrites the published SKILL.md artifacts that are missing.
func (r *MirrorRepairResult) repairPathB(fsys fs.FS, projectRoot string, published []string) {
	for _, rel := range published {
		dest := filepath.Join(projectRoot, filepath.FromSlash(rel))
		if _, err := os.Stat(dest); err == nil {
			// Occupied — restore-missing-only, never overwrite.
			continue
		}
		content, err := fs.ReadFile(fsys, rel)
		if err != nil {
			r.Warnings = append(r.Warnings,
				fmt.Sprintf("published skill %s: cannot read template: %v", rel, err))
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			r.Warnings = append(r.Warnings,
				fmt.Sprintf("published skill %s: cannot create directory: %v", rel, err))
			continue
		}
		if err := atomicWriteFile(dest, content, 0o644); err != nil {
			r.Warnings = append(r.Warnings,
				fmt.Sprintf("published skill %s: cannot write: %v", rel, err))
			continue
		}
		r.PublishedRestored = append(r.PublishedRestored, rel)
	}
}

// mirrorRepairCandidates walks the template FS and returns the canonical skill
// names (Path A candidates) and the published artifact paths (Path B), both in
// sorted order. It mirrors the deploy walk's own classification: the .tmpl
// suffix is trimmed first, so a path is classified by where it LANDS.
func mirrorRepairCandidates(fsys fs.FS) (canonical, published []string, err error) {
	seen := map[string]struct{}{}
	walkErr := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		destRel := strings.TrimSuffix(p, ".tmpl")
		if skill, ok := skillNameFromDeployPath(destRel); ok {
			if _, dup := seen[skill]; !dup {
				seen[skill] = struct{}{}
				canonical = append(canonical, skill)
			}
			return nil
		}
		// p == destRel restricts Path B to non-.tmpl sources, which is what
		// the published artifacts are: the repair pass carries no Renderer,
		// so a rendered published file would have to be read from a path
		// that does not exist in the FS.
		if p == destRel && isPublishedSkillPath(destRel) {
			published = append(published, destRel)
		}
		return nil
	})
	if walkErr != nil {
		return nil, nil, walkErr
	}
	sort.Strings(canonical)
	sort.Strings(published)
	return canonical, published, nil
}
