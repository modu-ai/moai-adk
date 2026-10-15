package spec

import "testing"

// The No-Haiku policy (REQ-AA2-012) is superseded for the git/todo/Explore
// roles by the operator's per-role defaults (card t1808, 2026-10-15): an agent
// definition may declare `model: haiku` in its frontmatter. The exemption is
// the frontmatter `model:` field ONLY — a haiku mention anywhere else in the
// agent file (body prose, another frontmatter key) still yields a finding.

func TestHaikuResidualRule_FrontmatterModelFieldExempt(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeHaikuFixtureFile(t, root, ".claude/agents/moai/manager-git.md",
		"---\nname: manager-git\ntools: Read, Bash\nmodel: haiku\neffort: medium\n---\nGit workflow body.\n")
	writeHaikuFixtureFile(t, root, "internal/template/templates/.claude/agents/moai/manager-git.md",
		"---\nname: manager-git\ntools: Read, Bash\nmodel: haiku\neffort: medium\n---\nGit workflow body.\n")

	if findings := (&HaikuResidualRule{baseDir: root}).CheckAll(nil); len(findings) != 0 {
		t.Errorf("frontmatter `model: haiku` must be exempt, got %d findings: %+v", len(findings), findings)
	}
}

func TestHaikuResidualRule_ProseAndOtherKeysStillFlagged(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"body prose":             "---\nname: a\nmodel: haiku\n---\nRoute cheap work to haiku.\n",
		"other frontmatter key":  "---\nname: a\ndescription: a haiku-tier helper\nmodel: sonnet\n---\nbody\n",
		"model line in the body": "---\nname: a\nmodel: sonnet\n---\nmodel: haiku\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeHaikuFixtureFile(t, root, ".claude/agents/moai/a.md", content)
			if findings := (&HaikuResidualRule{baseDir: root}).CheckAll(nil); len(findings) != 1 {
				t.Errorf("%s: got %d findings, want 1: %+v", name, len(findings), findings)
			}
		})
	}
}
