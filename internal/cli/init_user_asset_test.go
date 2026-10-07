// init_user_asset_test.go — the init first-install trigger
// (SPEC-USER-ASSET-INSTALL-001 REQ-024; AC-001's init-level arms).
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// TestInitEnsuresUserAssets covers AC-001's core arm end to end through the
// init command: after `moai init` on a fresh temp HOME, L0 skill dirs exist
// under $HOME/.claude/skills and the manifest records their hashes; the
// Codex roots carry the same set (AC-002 presence half).
func TestInitEnsuresUserAssets(t *testing.T) {
	_, homeDir := runInitForAutonomy(t, nil, map[string]string{"llm": "claude"})

	// L0 dispatcher + workflow skills under the Claude user root.
	for _, skill := range []string{"moai", "moai-workflow-tdd", "moai-workflow-spec"} {
		if _, err := os.Stat(filepath.Join(homeDir, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf("L0 skill %s missing under the user folder after init: %v", skill, err)
		}
	}
	// The Codex skill root carries the same L0 set (the dispatcher mirror
	// included — $HOME/.agents/skills/moai/SKILL.md, round-5 F2).
	if _, err := os.Stat(filepath.Join(homeDir, ".agents", "skills", "moai", "SKILL.md")); err != nil {
		t.Errorf("user-side dispatcher mirror missing: %v", err)
	}
	// The five D-Q1 agents under both agent roots.
	for _, agent := range []string{"manager-spec", "manager-develop", "manager-docs", "plan-auditor", "sync-auditor"} {
		if _, err := os.Stat(filepath.Join(homeDir, ".claude", "agents", agent+".md")); err != nil {
			t.Errorf("L0 agent %s missing under ~/.claude/agents: %v", agent, err)
		}
		if _, err := os.Stat(filepath.Join(homeDir, ".codex", "agents", agent+".toml")); err != nil {
			t.Errorf("L0 agent %s missing under ~/.codex/agents: %v", agent, err)
		}
	}
	// Manifest records the installed set.
	m, err := userassets.Load(userassets.ManifestPath(homeDir))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(m.Files) == 0 {
		t.Fatal("manifest empty after init")
	}
	fe, ok := m.Files["claude-skills/moai/SKILL.md"]
	if !ok || fe.SHA256 == "" || fe.Bundle != "core" || fe.MoaiVersion == "" {
		t.Errorf("dispatcher record incomplete: %+v ok=%v", fe, ok)
	}
}

// TestInitBundlesFlagRecordsSelection covers the REQ-004 selection surface:
// `moai init --bundles <name,...>` installs the named bundles' entries and
// records the selection in the manifest.
func TestInitBundlesFlagRecordsSelection(t *testing.T) {
	_, homeDir := runInitForAutonomy(t, nil, map[string]string{"llm": "claude", "bundles": "ops-tools"})

	if _, err := os.Stat(filepath.Join(homeDir, ".claude", "skills", "moai-workflow-loop", "SKILL.md")); err != nil {
		t.Errorf("opted-in bundle skill missing after init: %v", err)
	}
	m, err := userassets.Load(userassets.ManifestPath(homeDir))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range m.Bundles {
		if b == "ops-tools" {
			found = true
		}
	}
	if !found {
		t.Errorf("manifest bundles = %v, want ops-tools", m.Bundles)
	}
}
