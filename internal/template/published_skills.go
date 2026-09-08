// published_skills.go — deploy-side protection for the published /moai
// command skills (SPEC-CODEX-COMMAND-SKILLS-001 R-011).
//
// The published skills are committed template files under
// .agents/skills/moai-<command>/SKILL.md. They deploy through the regular
// walk like any other template file, but their path namespace overlaps the
// .agents/skills root the skill mirror also writes into, and update mode
// (forceUpdate) skips the provenance checks everywhere else. A user-owned
// file sitting at a published-skill path must survive an update: the check
// below keeps the init-mode provenance behavior alive on these paths even
// in update mode, and the skip is reported instead of being silent.
//
// The scope is deliberately narrower than the §24.4 template-managed
// moai-* convention, which classifies .claude/skills/moai-* names; this
// guard governs the published .agents/skills/moai-<command>/SKILL.md paths
// only, so the .claude/skills classification is untouched.
package template

import "strings"

// publishedSkillsDir is the deploy-relative published-skills root.
const publishedSkillsDir = ".agents/skills/"

// publishedSkillNames is the exact set of published command-skill
// directory names. It must equal the emitter's publication set (the
// committed tree under templates/.agents/skills/ is golden-pinned to that
// output by the drift check); publishedSkillsNamesMatchTree is the guard
// that keeps this list from drifting from the committed set.
var publishedSkillNames = map[string]struct{}{
	"moai-clean": {}, "moai-codemaps": {}, "moai-e2e": {},
	"moai-feedback": {}, "moai-fix": {}, "moai-gate": {},
	"moai-goal": {}, "moai-harness": {}, "moai-loop": {},
	"moai-mx": {}, "moai-plan": {}, "moai-project": {},
	"moai-review": {}, "moai-run": {}, "moai-sync": {},
	"moai-todo": {},
}

// isPublishedSkillPath reports whether a deploy-relative path is a
// published command-skill artifact: .agents/skills/<published name>/SKILL.md.
// The exact-name match is what separates published entries from the
// mirror's canonical-skill entries (name sets disjoint by the emitter's
// collision guard).
func isPublishedSkillPath(deployRelPath string) bool {
	rest, ok := strings.CutPrefix(deployRelPath, publishedSkillsDir)
	if !ok {
		return false
	}
	name, remainder, found := strings.Cut(rest, "/")
	if !found || remainder != "SKILL.md" {
		return false
	}
	_, isPublished := publishedSkillNames[name]
	return isPublished
}

// recordProtectedSkip appends one skipped published-skill path to the run's
// result. The deployer reports; the caller decides what the user sees.
func (r *DeployResult) recordProtectedSkip(deployRelPath string) {
	if r == nil {
		return
	}
	r.ProtectedSkips = append(r.ProtectedSkips, deployRelPath)
}
