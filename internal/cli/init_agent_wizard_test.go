package cli

// The wizard's agent-harness answer surface went with the general codex
// execution surface removal (card t1909): the harness selection is
// claude-only, the wizard no longer asks, and no Codex wiring runs on init.
// What remains here are the shared observation helpers and the
// non-interactive-path guards.
//
// The MCP-side observable throughout is the stdout announcement, NOT the
// existence of .mcp.json: that file is template-deployed and present in every
// branch, so a file-existence assertion would pass vacuously.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// mcpProvisionAnnouncement is the sole user-visible signal that the .mcp.json
// ensure-entry call ran (internal/cli/init.go, provisionMCPEntryUnlessDeclined).
// Present => provisioned; absent => declined.
const mcpProvisionAnnouncement = "Provisioned the moai MCP server entry in .mcp.json (default-on)."

// codexWiringArtifacts is all THREE artifacts codexwiring.Wire writes. Absence
// is asserted per file, never on the .codex/ directory — the template ships
// .codex/agents/** regardless (REQ-IHP-007).
var codexWiringArtifacts = []string{
	codexwiring.HooksRelPath,
	codexwiring.ConfigRelPath,
	codexwiring.SidecarPath,
}

func assertCodexArtifacts(t *testing.T, projectDir string, want bool) {
	t.Helper()
	for _, rel := range codexWiringArtifacts {
		_, err := os.Stat(filepath.Join(projectDir, rel))
		switch {
		case want && err != nil:
			t.Errorf("%s missing, want present: %v", rel, err)
		case !want && err == nil:
			t.Errorf("%s present, want absent", rel)
		}
	}
}

// TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence asserts AC-IHP-006a
// — the narrowed SPEC-CODEX-WIRING-001 REQ-CW-001 clause (spec.md §2 C3).
//
// With no --llm flag and no wizard, NONE of the three Codex artifacts is
// created and the provisioning announcement is ABSENT. The announcement
// direction is pinned in the assertion rather than left to prose: on this path
// opts.MCPProvision keeps its zero value false (its sole writer runs inside the
// interactive block and there is no --mcp-provision flag), so mcpDeclined is
// true and the announcement is never reached.
//
// This does NOT conflict with SPEC-CODEX-WIRING-001 AC-CW-004, which asserts
// the .mcp.json ENTRY is present on this same path: that observes the
// template-deployed file's content, this observes whether the ensure-entry call
// ran. Both are true. Do not reconcile them by weakening either.
func TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	// wizResult nil => the wizard seams are not swapped, so runInit takes the
	// non-interactive path.
	projectDir, stdout := runInitForAutonomyAtHomeCapturingOut(t, homeDir, nil, nil)

	// All THREE paths, deliberately: the sibling guard
	// TestRunInit_AgentClaudeLeavesNoCodexFiles asserts only two, and a
	// preservation criterion weaker than the guard it preserves is worse than
	// none. (Strengthening that landed sibling is a separate card, not this one.)
	assertCodexArtifacts(t, projectDir, false)

	if strings.Contains(stdout, mcpProvisionAnnouncement) {
		t.Errorf("a flag-absent non-interactive init must not run the .mcp.json ensure-entry call (REQ-IHP-007); stdout:\n%s", stdout)
	}
}

// TestRunWizardFn_ZeroInvocationsNonInteractive asserts AC-IHP-005: the
// injectable wizard seam runWizardFn is invoked exactly ZERO times when the run
// is non-interactive — both via --non-interactive and via an absent TTY.
//
// This is a deliberate deviation from SPEC-CODEX-INIT-001 AC-CI-004's letter,
// which counts PROMPTS. The claim it rests on is narrow and measured:
// runWizardFn is the sole issuer of the HARNESS QUESTION (it renders only
// through Page3Questions -> InitQuestions -> runWizardFn), so zero invocations
// entails zero harness prompts. It is NOT the sole prompt issuer on the init
// path — a second, independent huh confirm (profile setup) reads
// isatty.IsTerminal directly and runs before the wizard gate — so this test
// makes no claim about prompt totals. Building a prompt-counting instrument is
// out of scope.
func TestRunWizardFn_ZeroInvocationsNonInteractive(t *testing.T) {
	cases := []struct {
		name        string
		flags       map[string]string
		interactive bool
	}{
		{"--non-interactive with a TTY present", map[string]string{"non-interactive": "true"}, true},
		{"absent TTY", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := t.TempDir()
			t.Setenv("HOME", homeDir)
			t.Setenv("MOAI_SANDBOX_PROOF", "")
			t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

			var calls int
			origWizard := runWizardFn
			runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
				calls++
				return &wizard.WizardResult{}, nil
			}
			t.Cleanup(func() { runWizardFn = origWizard })

			origInteractive := isInteractiveStdin
			isInteractiveStdin = func() bool { return tc.interactive }
			t.Cleanup(func() { isInteractiveStdin = origInteractive })

			origDeps := deps
			deps = nil
			t.Cleanup(func() { deps = origDeps })

			projectDir := filepath.Join(t.TempDir(), "noninteractive-proj")
			cmd := newInitTestCmd()
			for name, val := range tc.flags {
				if err := cmd.Flags().Set(name, val); err != nil {
					t.Fatalf("set --%s=%s: %v", name, val, err)
				}
			}
			if err := runInit(cmd, []string{projectDir}); err != nil {
				t.Fatalf("runInit: %v", err)
			}

			if calls != 0 {
				t.Errorf("runWizardFn was invoked %d times, want 0 (AC-IHP-005)", calls)
			}
		})
	}
}
