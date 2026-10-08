// m0_frozen_guard_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, hook
// family: AC-009 (frozen guard denies a user-path delete).
//
// M0.1 repair (gate round 13): the test now drives the REAL hook entry
// (Handle) — the former decideBash helper skipped Handle's
// checkProtectedZoneShell branch, so the observed verdict was a partial
// one. Two arms, one Given:
//   - CONTROL: deleting the PROJECT-side covered twin is denied by Handle
//     at HEAD — proves the guard and the loaded config are live;
//   - DEFECT ARM: deleting the USER-side twin (the same managed file under
//     the user-install root, userassets paths.go:38's four roots) is
//     ALLOWED at HEAD — the user roots are absent from the loaded
//     protection set (원장 9a). RED here is the M4 fix target.
package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFrozenGuardDeniesUserPathDelete(t *testing.T) {
	root := t.TempDir() // the project the hook guards
	home := t.TempDir() // the user install root's base

	// The loaded protection set: the project overlay declares the managed
	// asset roots — BOTH faces (the shipped manifest's shape for the
	// project side, and the M4 user-root kind for the user side).
	overlayDir := filepath.Join(root, ".moai", "project")
	if err := os.MkdirAll(overlayDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overlay := `version: 1
categories:
  managed_assets:
    runtime_paths:
      - .claude/agents/moai/
      - .claude/skills/
      - .codex/agents/moai/
      - user-root:claude-agents/moai/
      - user-root:claude-skills/
      - user-root:codex-agents/moai/
`
	if err := os.WriteFile(filepath.Join(overlayDir, "protected-zone.yaml"), []byte(overlay), 0o644); err != nil {
		t.Fatal(err)
	}
	// The managed files exist on BOTH faces — the project twins (covered)
	// and the user install twins (the AC's Given).
	for _, base := range []string{root, home} {
		for _, rel := range []string{
			filepath.Join(".claude", "agents", "moai", "plan-auditor.md"),
			filepath.Join(".claude", "skills", "moai-foundation-core", "SKILL.md"),
		} {
			abs := filepath.Join(base, rel)
			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(abs, []byte("---\nname: managed\n---\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	// An UNTRACKED user file in a managed directory: REQ-GRD-002 limits the
	// protection to moai-managed assets — this one must stay editable.
	untrackedAbs := filepath.Join(home, ".claude", "agents", "moai", "user-own-note.md")
	if err := os.WriteFile(untrackedAbs, []byte("the user's own note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The MANAGED status is evidenced by the user manifest tracking the
	// targets (REQ-GRD-002 — the record is what proves a target qualifies).
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	userManifest := `{
  "schema_version": 1,
  "bundles": ["core"],
  "files": {
    "claude-agents/moai/plan-auditor.md": {"sha256": "ab", "bundle": "core"},
    "claude-skills/moai-foundation-core/SKILL.md": {"sha256": "cd", "bundle": "core"}
  }
}
`
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte(userManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	// The zone resolver resolves relative command paths against the hook
	// process cwd — pin it to the guarded project for the control arm. The
	// defect arm targets the USER twin by its ABSOLUTE path — the real
	// shape the user-root resolution (design §5) must judge.
	prevGetwd := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prevGetwd })
	prevHome := zoneHomeFn
	zoneHomeFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { zoneHomeFn = prevHome })

	decide := func(command string) string {
		t.Helper()
		h := &preToolHandler{policy: DefaultSecurityPolicy(), projectDir: root}
		raw, err := json.Marshal(map[string]string{"command": command})
		if err != nil {
			t.Fatalf("marshal command: %v", err)
		}
		out, err := h.Handle(context.Background(), &HookInput{
			ToolName:  "Bash",
			ToolInput: raw,
			AgentType: harnessLearnerIdentity,
		})
		if err != nil {
			t.Fatalf("Handle(%q): %v", command, err)
		}
		if out.HookSpecificOutput == nil {
			return ""
		}
		return out.HookSpecificOutput.PermissionDecision
	}

	// CONTROL: the project-side covered twin is denied — the guard and its
	// config are live, so a user-path allow below would be a SCOPE defect,
	// not a dead guard.
	if got := decide("rm -f .claude/agents/moai/plan-auditor.md"); got != DecisionDeny {
		t.Fatalf("control arm broke: deleting the PROJECT-side covered twin was not denied (decision %q) — the guard/config setup is invalid, the defect arm below would be meaningless", got)
	}

	// AC-009 flip (M4): deleting a TRACKED user-installed managed asset is
	// denied — by absolute path and by the ~ alias both.
	for _, command := range []string{
		"rm -f \"" + filepath.Join(home, ".claude", "agents", "moai", "plan-auditor.md") + "\"",
		"rm -f ~/.claude/skills/moai-foundation-core/SKILL.md",
	} {
		if got := decide(command); got != DecisionDeny {
			t.Errorf("AC-009 regression: delete of a TRACKED user-installed managed asset was allowed (decision %q, want %q) — command: %s", got, DecisionDeny, command)
		}
	}

	// REQ-GRD-002 over-protection guard: an UNTRACKED user file in a
	// managed directory stays editable — the protection is scoped to
	// moai-managed assets, never to the user's own files.
	if got := decide("rm -f \"" + untrackedAbs + "\""); got == DecisionDeny {
		t.Errorf("over-protection: an UNTRACKED user file was denied (%q) — the protection must stay scoped to manifest-tracked assets", untrackedAbs)
	}
}
