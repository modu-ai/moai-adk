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
	// asset roots (the shipped manifest's shape, overlay syntax — no
	// required categories).
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
`
	if err := os.WriteFile(filepath.Join(overlayDir, "protected-zone.yaml"), []byte(overlay), 0o644); err != nil {
		t.Fatal(err)
	}
	// The managed file exists on BOTH faces — the project twin (covered) and
	// the user install twin (the AC's Given).
	twin := filepath.Join(".claude", "agents", "moai", "plan-auditor.md")
	for _, base := range []string{root, home} {
		abs := filepath.Join(base, twin)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("---\nname: plan-auditor\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Gate round 18: the user twin's MANAGED status is evidenced by the user
	// manifest tracking it (REQ-GRD-002 limits the protection to moai-
	// managed files — the record is what proves this target qualifies).
	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	userManifest := `{
  "schema_version": 1,
  "bundles": ["core"],
  "files": {
    "claude-agents/moai/plan-auditor.md": {
      "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "bundle": "core",
      "installed_at": "2026-10-08T00:00:00Z",
      "moai_version": "test"
    }
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

	// CONTROL: the project-side covered twin is denied at HEAD — the guard
	// and its config are live, so a user-path allow below is a SCOPE defect,
	// not a dead guard.
	if got := decide("rm -f .claude/agents/moai/plan-auditor.md"); got != DecisionDeny {
		t.Fatalf("control arm broke: deleting the PROJECT-side covered twin was not denied (decision %q) — the guard/config setup is invalid, the defect arm below would be meaningless", got)
	}

	// DEFECT ARM: the user-side twin of the same managed file, named by its
	// absolute path under the real user home (and its tilde alias — both
	// spellings a real deletion uses).
	for _, command := range []string{
		"rm -f \"" + filepath.Join(home, ".claude", "agents", "moai", "plan-auditor.md") + "\"",
		"rm -rf \"" + filepath.Join(home, ".claude", "skills", "moai-foundation-core") + "\"",
	} {
		if got := decide(command); got != DecisionDeny {
			t.Errorf("RED (intended): delete of a user-installed managed asset was allowed (decision %q, want %q) — command: %s; the user install roots are not in the loaded protection set", got, DecisionDeny, command)
		}
	}
}
