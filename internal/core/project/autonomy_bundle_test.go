package project

// SPEC-AUTONOMY-TIERS-001 M9 — init applies the tier bundle (end-to-end wiring).
//
// Gap 1: applyAutonomyTierFromWizard (init.go) captures the effective tier into
// opts.AutonomyTier, but the init deployment path never CONSUMES it — no
// permission bundle (defaultMode + deny/ask) is written, so selecting a tier has
// NO effect on deployed permissions. M9 closes that gap: ApplyAutonomyTierBundle
// wires opts.AutonomyTier into the deployed settings at init time, reusing the
// existing core (EffectiveTierWithGates + TierDefaultMode + RenderTierPermissions).
//
// REQ-004 invariant (SPEC-AUT-PERMMODES-001, re-scoped from SPEC-AUTONOMY-TIERS-001
// REQ-007): semi-auto / unset writes ONLY the USER-scope defaultMode="acceptEdits"
// record — the PROJECT file and every other deployed file stay byte-identical.
// REQ-006 invariant: init MUST NOT default to fully-autonomous.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// seedTemplateProjectSettings writes a PROJECT-scope settings.json shaped like
// the deployed template: an allow/ask/deny block with NO defaultMode. The
// bounded-delta assertions compare against this seed verbatim.
func seedTemplateProjectSettings(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{
  "permissions": {
    "allow": ["Read", "Write"],
    "ask": ["Bash(rm:*)"],
    "deny": ["Bash(rm -rf /:*)"]
  }
}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readPermissionsDefaultMode(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var wrap struct {
		Permissions struct {
			DefaultMode string   `json:"defaultMode"`
			Allow       []string `json:"allow"`
			Ask         []string `json:"ask"`
			Deny        []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		t.Fatalf("unmarshal %s: %v\nbody: %s", path, err, body)
	}
	return wrap.Permissions.DefaultMode
}

// TestApplyAutonomyTierBundle_AutomaticWritesUserDefaultModeAuto: the ESSENTIAL
// end-to-end assertion. `moai init --autonomy-tier=automatic` MUST produce a
// USER-scope settings.json whose defaultMode reflects auto. Without M9 the
// defaultMode is never written, so the tier selection has no deployed effect.
func TestApplyAutonomyTierBundle_AutomaticWritesUserDefaultModeAuto(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	dir := t.TempDir()
	userPath := filepath.Join(dir, "home", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	seedTemplateProjectSettings(t, projectPath)

	if err := ApplyAutonomyTierBundle(dir, userPath, projectPath, config.AutonomyTierAutomatic); err != nil {
		t.Fatalf("ApplyAutonomyTierBundle: %v", err)
	}

	if got := readPermissionsDefaultMode(t, userPath); got != "auto" {
		t.Errorf("USER defaultMode = %q, want %q (automatic MUST deploy defaultMode=auto)", got, "auto")
	}
}

// readPermissionsBlock parses the permissions object at path into its raw
// JSON keys so the bounded-delta tests can assert EXACTLY which keys the
// bundle wrote (REQ-004: the sanctioned delta is one key in one file).
func readPermissionsBlock(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var wrap struct {
		Permissions map[string]json.RawMessage `json:"permissions"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		t.Fatalf("unmarshal %s: %v\nbody: %s", path, err, body)
	}
	return wrap.Permissions
}

// TestApplyAutonomyTierBundle_SemiAutoIsBoundedDelta: REQ-004 of
// SPEC-AUT-PERMMODES-001 (re-scoped) — semi-auto writes ONLY the USER-scope
// defaultMode="acceptEdits" record: the USER permissions block carries exactly
// that one key, and the PROJECT file is byte-identical to the template seed.
func TestApplyAutonomyTierBundle_SemiAutoIsBoundedDelta(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "home", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	seedTemplateProjectSettings(t, projectPath)
	projectBefore, _ := os.ReadFile(projectPath)

	if err := ApplyAutonomyTierBundle(dir, userPath, projectPath, config.AutonomyTierSemiAuto); err != nil {
		t.Fatalf("ApplyAutonomyTierBundle: %v", err)
	}

	// USER permissions block carries EXACTLY the acceptEdits defaultMode —
	// no allow/ask/deny keys may leak into USER scope (bounded delta).
	perms := readPermissionsBlock(t, userPath)
	if len(perms) != 1 {
		t.Errorf("semi-auto USER permissions must carry EXACTLY one key (bounded delta), got %d: %v", len(perms), perms)
	}
	if got := readPermissionsDefaultMode(t, userPath); got != "acceptEdits" {
		t.Errorf("semi-auto USER defaultMode = %q, want %q (SPEC-AUT-PERMMODES-001 REQ-004)", got, "acceptEdits")
	}
	// PROJECT file MUST be byte-identical to the seed.
	projectAfter, _ := os.ReadFile(projectPath)
	if string(projectBefore) != string(projectAfter) {
		t.Errorf("semi-auto MUST NOT modify PROJECT settings (bounded delta):\nbefore: %s\nafter:  %s",
			projectBefore, projectAfter)
	}
}

// TestApplyAutonomyTierBundle_EmptyIsBoundedDelta: unset tier (the init
// default) resolves to semi-auto and MUST pay the SAME bounded delta —
// USER defaultMode=acceptEdits only, PROJECT byte-identical.
func TestApplyAutonomyTierBundle_EmptyIsBoundedDelta(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "home", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	seedTemplateProjectSettings(t, projectPath)
	projectBefore, _ := os.ReadFile(projectPath)

	if err := ApplyAutonomyTierBundle(dir, userPath, projectPath, ""); err != nil {
		t.Fatalf("ApplyAutonomyTierBundle: %v", err)
	}
	perms := readPermissionsBlock(t, userPath)
	if len(perms) != 1 {
		t.Errorf("empty tier USER permissions must carry EXACTLY one key (bounded delta), got %d: %v", len(perms), perms)
	}
	if got := readPermissionsDefaultMode(t, userPath); got != "acceptEdits" {
		t.Errorf("empty tier USER defaultMode = %q, want %q (SPEC-AUT-PERMMODES-001 REQ-004)", got, "acceptEdits")
	}
	projectAfter, _ := os.ReadFile(projectPath)
	if string(projectBefore) != string(projectAfter) {
		t.Errorf("empty tier MUST NOT modify PROJECT settings (bounded delta):\nbefore: %s\nafter:  %s",
			projectBefore, projectAfter)
	}
}

// TestApplyAutonomyTierBundle_FullyAutonomousDowngradedWithoutProof: a flag
// `--autonomy-tier=fully-autonomous` with no sandbox proof is gated down to
// automatic (AC-005). The deployed defaultMode MUST be "auto" (NOT bypass).
func TestApplyAutonomyTierBundle_FullyAutonomousDowngradedWithoutProof(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	dir := t.TempDir()
	userPath := filepath.Join(dir, "home", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	seedTemplateProjectSettings(t, projectPath)

	if err := ApplyAutonomyTierBundle(dir, userPath, projectPath, config.AutonomyTierFullyAutonomous); err != nil {
		t.Fatalf("ApplyAutonomyTierBundle: %v", err)
	}
	if got := readPermissionsDefaultMode(t, userPath); got != "auto" {
		t.Errorf("fully-autonomous w/o proof MUST downgrade to defaultMode=auto, got %q", got)
	}
}

// TestApplyAutonomyTierBundle_FullyAutonomousWithProofDeploysBypass: when a
// sandbox proof IS present AND the kill-switch is off, fully-autonomous deploys
// defaultMode=bypassPermissions (REQ-006 — opt-in, never the default).
func TestApplyAutonomyTierBundle_FullyAutonomousWithProofDeploysBypass(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "docker")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	dir := t.TempDir()
	userPath := filepath.Join(dir, "home", ".claude", "settings.json")
	projectPath := filepath.Join(dir, "proj", ".claude", "settings.json")
	seedTemplateProjectSettings(t, projectPath)

	if err := ApplyAutonomyTierBundle(dir, userPath, projectPath, config.AutonomyTierFullyAutonomous); err != nil {
		t.Fatalf("ApplyAutonomyTierBundle: %v", err)
	}
	if got := readPermissionsDefaultMode(t, userPath); got != "bypassPermissions" {
		t.Errorf("fully-autonomous w/ proof + no kill-switch MUST deploy defaultMode=bypassPermissions, got %q", got)
	}
}

// TestApplyAutonomyTierBundle_PolicyDocFullBundle: when a tool-policy.yaml is
// available at the project, the renderer reuses RenderTierPermissions for the
// full bundle — defaultMode in USER scope, deny/ask regenerated in PROJECT scope.
func TestApplyAutonomyTierBundle_PolicyDocFullBundle(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	dir := t.TempDir()
	projectRoot := filepath.Join(dir, "proj")
	userPath := filepath.Join(dir, "home", ".claude", "settings.json")
	projectPath := filepath.Join(projectRoot, ".claude", "settings.json")
	seedTemplateProjectSettings(t, projectPath)

	// Deploy a minimal tool-policy.yaml so LoadFromProjectDir succeeds.
	policyDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(policyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	policyBody := "entries:\n" +
		"  - tool: Bash\n" +
		"    args_pattern: \"git push --force:*\"\n" +
		"    decision: deny\n" +
		"    risk_tier: irreversible\n" +
		"    owner_agent: manager-git\n" +
		"    audit: commit-log\n" +
		"  - tool: Bash\n" +
		"    args_pattern: \"git push:*\"\n" +
		"    decision: ask\n" +
		"    risk_tier: write\n" +
		"    owner_agent: manager-git\n" +
		"    audit: commit-log\n"
	if err := os.WriteFile(filepath.Join(policyDir, "tool-policy.yaml"), []byte(policyBody), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ApplyAutonomyTierBundle(projectRoot, userPath, projectPath, config.AutonomyTierAutomatic); err != nil {
		t.Fatalf("ApplyAutonomyTierBundle: %v", err)
	}
	// USER scope: defaultMode=auto, NO deny/ask (they are PROJECT-scoped).
	userBody, _ := os.ReadFile(userPath)
	var userPerms struct {
		Permissions struct {
			DefaultMode string   `json:"defaultMode"`
			Deny        []string `json:"deny"`
			Ask         []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(userBody, &userPerms); err != nil {
		t.Fatalf("unmarshal user: %v\nbody: %s", err, userBody)
	}
	if userPerms.Permissions.DefaultMode != "auto" {
		t.Errorf("USER defaultMode = %q, want auto", userPerms.Permissions.DefaultMode)
	}
	if len(userPerms.Permissions.Deny) != 0 || len(userPerms.Permissions.Ask) != 0 {
		t.Errorf("USER scope MUST NOT carry deny/ask (PROJECT-scoped): deny=%v ask=%v",
			userPerms.Permissions.Deny, userPerms.Permissions.Ask)
	}
	// PROJECT scope: deny/ask present (regenerated from the doc by the renderer).
	projectBody, _ := os.ReadFile(projectPath)
	if !contains(projectBody, "git push --force") || !contains(projectBody, "git push") {
		t.Errorf("PROJECT scope MUST carry deny/ask regenerated from the doc; body: %s", projectBody)
	}
}

func contains(haystack []byte, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(string(haystack), needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
