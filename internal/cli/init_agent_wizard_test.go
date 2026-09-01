package cli

// SPEC-INIT-HARNESS-PROMPT-001 — the wizard's agent-harness answer reaches
// BOTH consumers of the harness selection, and the --agent flag still beats it.
//
// The centre of gravity is the resolution seam, not the question: before this
// change resolveAgentWiring read only the cobra flag, so no wizard answer could
// reach either the Codex wiring call or the .mcp.json provisioning precedence
// switch (spec.md §4 D1).
//
// The MCP-side observable throughout is the stdout announcement, NOT the
// existence of .mcp.json: that file is template-deployed and present in every
// branch, so a file-existence assertion would pass vacuously (acceptance.md
// §A.1).
//
// @MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001

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

// TestRunInit_WizardCodexReachesBothConsumers asserts AC-IHP-003: with --agent
// absent and the wizard answering codex, the selection reaches BOTH consumers
// in one test — (a) the Codex wiring call wrote all three artifacts, and (b)
// the MCP precedence switch declined provisioning, so the announcement is
// absent.
//
// mcp_provision is pinned to YES so leg (b) cannot pass by way of the decline
// default: without the harness selection reaching the switch, a yes answer
// provisions and the announcement appears.
func TestRunInit_WizardCodexReachesBothConsumers(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "codex", MCPProvision: true}
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	projectDir, stdout := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

	// Leg (a): the Codex-wiring consumer saw codex.
	assertCodexArtifacts(t, projectDir, true)

	// Leg (b): the MCP-precedence consumer saw codex.
	if strings.Contains(stdout, mcpProvisionAnnouncement) {
		t.Errorf("stdout contains %q — the MCP precedence consumer did not see the wizard's codex answer (AC-IHP-003 leg b)\nstdout:\n%s", mcpProvisionAnnouncement, stdout)
	}
}

// TestRunInit_WizardCodexDeclinesMCPProvisioning asserts AC-IHP-009: the
// wizard × wizard combination — harness codex, mcp_provision yes — resolves in
// the harness selection's favour (spec.md §4 D2), and the moai MCP server is
// still registered for the user, through .codex/config.toml rather than
// .mcp.json.
func TestRunInit_WizardCodexDeclinesMCPProvisioning(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "codex", MCPProvision: true}
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	projectDir, stdout := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

	if strings.Contains(stdout, mcpProvisionAnnouncement) {
		t.Errorf("harness codex must decline .mcp.json provisioning even with mcp_provision=yes (REQ-IHP-009); stdout:\n%s", stdout)
	}

	cfg, err := os.ReadFile(filepath.Join(projectDir, codexwiring.ConfigRelPath))
	if err != nil {
		t.Fatalf("read %s: %v", codexwiring.ConfigRelPath, err)
	}
	if !strings.Contains(string(cfg), "mcp_servers.moai") {
		t.Errorf("%s does not carry the mcp_servers.moai entry — declining .mcp.json under codex would withdraw the MCP surface; got:\n%s", codexwiring.ConfigRelPath, cfg)
	}
}

// TestRunInit_FlagClaudeBeatsWizardCodex asserts the AC-IHP-004 claude row:
// an explicit --agent claude discards a wizard answer of codex, so no Codex
// artifact is written and provisioning follows the mcp_provision answer.
//
// The outcome alone is vacuous (the wizard answer was discarded before this
// change too), which is why the precedence RULE is asserted directly in
// TestResolveAgentWiringWithWizard_PrecedenceTable. This row is the end-to-end
// companion, not the binding evidence.
func TestRunInit_FlagClaudeBeatsWizardCodex(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "codex", MCPProvision: true}
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	projectDir, stdout := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, map[string]string{"agent": "claude"})

	assertCodexArtifacts(t, projectDir, false)
	if !strings.Contains(stdout, mcpProvisionAnnouncement) {
		t.Errorf("--agent claude must leave the mcp_provision answer (yes) intact, so the announcement is expected; stdout:\n%s", stdout)
	}
}

// TestRunInit_FlagBothBeatsWizardCodexAndForcesProvisioning asserts the
// AC-IHP-004 both row: --agent both beats a wizard answer of codex AND forces
// provisioning on over an explicit mcp_provision decline.
//
// mcp_provision is pinned to NO deliberately: with a yes the announcement
// would be emitted anyway and the row would assert nothing.
func TestRunInit_FlagBothBeatsWizardCodexAndForcesProvisioning(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "codex", MCPProvision: false}
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	projectDir, stdout := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, map[string]string{"agent": "both"})

	assertCodexArtifacts(t, projectDir, true)
	if !strings.Contains(stdout, mcpProvisionAnnouncement) {
		t.Errorf("--agent both must force provisioning on over an explicit mcp_provision decline (REQ-IHP-009); stdout:\n%s", stdout)
	}
}

// TestRunInit_WizardBothForcesProvisioningOverDecline asserts the both rule
// holds when it arrives from the WIZARD rather than the flag — the half of
// REQ-IHP-009 that says the rule is about the harness selection's origin being
// irrelevant.
func TestRunInit_WizardBothForcesProvisioningOverDecline(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "both", MCPProvision: false}
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	projectDir, stdout := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

	assertCodexArtifacts(t, projectDir, true)
	if !strings.Contains(stdout, mcpProvisionAnnouncement) {
		t.Errorf("a wizard answer of both must force provisioning on (REQ-IHP-009); stdout:\n%s", stdout)
	}
}

// TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence asserts AC-IHP-006a
// — the narrowed SPEC-CODEX-WIRING-001 REQ-CW-001 clause (spec.md §2 C3).
//
// With no --agent flag and no wizard, NONE of the three Codex artifacts is
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
