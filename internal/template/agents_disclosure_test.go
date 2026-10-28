package template

// SPEC-INIT-HARNESS-001 M3 — CLAUDE.md disclosure completeness (REQ-IH-008,
// AC-IH-007): the deployed CLAUDE.md explicitly discloses every capability in
// the claude-only runtime freeze list. The template file is the deployment
// source, so the check reads it directly; the byte ceiling guard rides along.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// templateSourceRoot is the on-disk template tree that //go:embed compiles in
// (same constant as embed_skill_guard_test.go).
const templateSourceRootClaude = "templates"

// disclosureMatrix maps each freeze-list capability to the marker the
// CLAUDE.md must carry. Table-row capabilities are asserted by their row id;
// prose-covered capabilities by their section marker.
var disclosureMatrix = []struct {
	id     string // freeze-list capability id (test diagnostics)
	marker string // text that must appear in the deployed CLAUDE.md
}{
	{"question-channel", "question-channel"},
	{"task-list", "task-list"},
	{"design-sync", "design-sync"},
	{"agent-spawning", "agent-spawning"},
	{"output-style", "output-style"},
	{"slash-commands", "slash-commands"},
	{"workflow-scripts", "workflow-scripts"},
	// The skill loader IS disclosed as reachable by every harness (the prose
	// pointer), but the codex-only spec requires stating its NON-EQUIVALENCE
	// (deferred) explicitly — that is the missing piece this row pins.
	{"skill-loader-non-equivalence", "deferred"},
	{"hook-event-coverage", "Hook Event Coverage"},
}

// TestAgentsDisclosureCompleteness asserts every freeze-list capability is
// disclosed in the shipped CLAUDE.md (AC-IH-007). Failures report exactly
// which disclosures are missing.
func TestAgentsDisclosureCompleteness(t *testing.T) {
	// The template source ships as `CLAUDE.md.tmpl`; the deployer strips the
	// suffix so a user project still receives `CLAUDE.md`. The suffix keeps the
	// mirror out of Codex's filename-keyed discovery inside THIS repo, where it
	// would otherwise merge with the root contract and truncate the tail
	// silently (card t925).
	data, err := os.ReadFile(filepath.Join(templateSourceRootClaude, "CLAUDE.md.tmpl"))
	if err != nil {
		t.Fatalf("read template CLAUDE.md: %v", err)
	}
	body := string(data)

	var missing []string
	for _, d := range disclosureMatrix {
		if !strings.Contains(body, d.marker) {
			missing = append(missing, d.id)
		}
	}
	if len(missing) > 0 {
		t.Errorf("missing disclosures: %v", missing)
	}

	// Byte ceiling (REQ-IH-008): the codex contract must stay within
	// config.CodexContractByteCeiling so the CLAUDE.md disclosure hardening
	// cannot crowd out the contract itself.
	if ceiling := config.CodexContractByteCeiling; ceiling > 0 && len(data) > ceiling {
		t.Errorf("CLAUDE.md is %d bytes, exceeds CodexContractByteCeiling %d", len(data), ceiling)
	}
}
