// cross_model_audit_wiring_test.go: CI guard for the cross-model audit
// consumer link.
//
// `moai-ref-cross-model-audit` describes itself as "the single skill both
// audit entry points load", but neither auditor referenced it and neither
// declared the `audit_multi` MCP tool — so the skill was unreachable and the
// tool it documents was uncallable from the agents that exist to call it.
//
// This guard asserts the two halves of that link in both trees:
//
//  1. each auditor's Conditional Skill Loading section names the skill, so the
//     agent knows to load it on demand;
//  2. each auditor declares the MCP tool in `tools:`, so the runtime actually
//     exposes it (an explicit tools list is a whitelist — an undeclared
//     `mcp__` tool is not inherited).
//
// It also pins plan-auditor's independence invariant: the skill is loaded on
// demand only, never as a static `skills:` preload.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	crossModelAuditSkill = "moai-ref-cross-model-audit"
	auditMultiMCPTool    = "mcp__moai__audit_multi"
)

var crossModelAuditAgents = []string{
	".claude/agents/moai/plan-auditor.md",
	".claude/agents/moai/sync-auditor.md",
}

func TestCrossModelAuditConsumerLink(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for _, rel := range crossModelAuditAgents {
		for tree, base := range map[string]string{
			"source":   projectRoot,
			"template": filepath.Join(projectRoot, "internal", "template", "templates"),
		} {
			rel, tree, base := rel, tree, base
			t.Run(filepath.Base(rel)+"/"+tree, func(t *testing.T) {
				t.Parallel()

				path := filepath.Join(base, rel)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				body := string(data)

				if !strings.Contains(body, `Skill("`+crossModelAuditSkill+`")`) {
					t.Errorf("%s does not instruct loading Skill(%q); the skill claims both audit "+
						"entry points load it, so an unreferenced skill is an unreachable one", path, crossModelAuditSkill)
				}
				if !strings.Contains(body, auditMultiMCPTool) {
					t.Errorf("%s does not declare the %q tool; an explicit tools list is a whitelist, "+
						"so the agent cannot call the tool the skill documents", path, auditMultiMCPTool)
				}
			})
		}
	}
}

// TestReviewWorkflowCrossModelConvergenceWiring asserts the /moai review
// workflow document wires the cross-model convergence step: it names the
// audit_multi tool, instructs loading the usage-SSOT skill on the reviewer
// spawn, and documents the fail-open fallback that keeps the single-model
// path when backends are absent. Checked in both the source tree and the
// template mirror.
func TestReviewWorkflowCrossModelConvergenceWiring(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for tree, base := range map[string]string{
		"source":   projectRoot,
		"template": filepath.Join(projectRoot, "internal", "template", "templates"),
	} {
		tree, base := tree, base
		t.Run(tree, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(base, ".claude/skills/moai/workflows/review.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			body := string(data)

			if !strings.Contains(body, auditMultiMCPTool) {
				t.Errorf("%s does not name the %q tool; the review verdict step cannot "+
					"call the convergence tool the skill documents", path, auditMultiMCPTool)
			}
			if !strings.Contains(body, `Skill("`+crossModelAuditSkill+`")`) {
				t.Errorf("%s does not instruct loading Skill(%q); the convergence phase "+
					"cross-references the usage-SSOT skill for its call contract", path, crossModelAuditSkill)
			}
			if !strings.Contains(body, "fail-open") {
				t.Errorf("%s does not document the fail-open fallback; absent backends "+
					"must keep the single-model path, not block the review", path)
			}
		})
	}
}

// TestPlanAuditorHasNoStaticSkillPreload pins the independence invariant the
// plan-auditor body states: no static `skills:` preload. The cross-model audit
// skill is wired as an on-demand load precisely so this stays true.
func TestPlanAuditorHasNoStaticSkillPreload(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for tree, base := range map[string]string{
		"source":   projectRoot,
		"template": filepath.Join(projectRoot, "internal", "template", "templates"),
	} {
		tree, base := tree, base
		t.Run(tree, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(base, ".claude/agents/moai/plan-auditor.md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			// Frontmatter is the leading --- delimited block.
			parts := strings.SplitN(string(data), "---", 3)
			if len(parts) < 3 {
				t.Fatalf("%s has no frontmatter block", path)
			}
			for _, line := range strings.Split(parts[1], "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "skills:") {
					t.Errorf("%s declares a static `skills:` preload (%q); auditor independence "+
						"requires on-demand loading only", path, strings.TrimSpace(line))
				}
			}
		})
	}
}
