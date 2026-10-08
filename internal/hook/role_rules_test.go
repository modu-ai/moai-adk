package hook

// SessionStart role-rules injection tests of SPEC-ALWAYS-LOADED-BUDGET-001
// (REQ-ALB-007..011, AC-ALB-009..013).
//
// Fixture rule files are copies of the deployed template rule files (the
// deployment origin), written into a t.TempDir() project root — the builder
// reads the deployed tree only (REQ-ALB-023). Role markers ride t.Setenv
// (never raw os.Setenv) so parallel tests cannot cross-contaminate.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeDeployedRoleRules copies the deployed role-gated rule files from the
// template source into root (the deployed layout), returning the rule paths.
func writeDeployedRoleRules(t *testing.T, root string) {
	t.Helper()
	for _, rule := range roleRuleFiles {
		src := filepath.Join("..", "..", "internal", "template", "templates", filepath.FromSlash(rule.Rel))
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read deployed template rule %s: %v", rule.Rel, err)
		}
		dst := filepath.Join(root, filepath.FromSlash(rule.Rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rule.Rel, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatalf("write deployed rule %s: %v", rule.Rel, err)
		}
	}
}

// writeRoleRuleFixture writes ONE role rule file with arbitrary content.
func writeRoleRuleFixture(t *testing.T, root string, rule roleRuleFile, content string) {
	t.Helper()
	dst := filepath.Join(root, filepath.FromSlash(rule.Rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rule.Rel, err)
	}
	if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", rule.Rel, err)
	}
}

// removeRoleRule deletes one role rule file (the absent fixture).
func removeRoleRule(t *testing.T, root string, rule roleRuleFile) {
	t.Helper()
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(rule.Rel))); err != nil {
		t.Fatalf("remove %s: %v", rule.Rel, err)
	}
}

// smallMarkedRule builds a rule file whose role core is one small block —
// a core that fits the delivery cap with room to spare.
func smallMarkedRule(marker string) string {
	return "# Fixture rule\n\n" +
		config.RoleCoreMarkerStart + "\nSmall binding block for " + marker + ": the session must wait for evidence before advancing a card.\n" +
		config.RoleCoreMarkerEnd + "\n\nplain tail\n"
}

// roleCoreBlocksFromDeployed derives the expected role-core block set from
// the DEPLOYED rule files under root (REQ-ALB-011: the expectation is
// derived, never hand-written).
func roleCoreBlocksFromDeployed(t *testing.T, root string) []string {
	t.Helper()
	var blocks []string
	for _, rule := range roleRuleFiles {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rule.Rel)))
		if err != nil {
			t.Fatalf("read deployed rule %s: %v", rule.Rel, err)
		}
		regions, marked := config.ExtractRoleCoreRegions(string(data))
		if !marked {
			t.Fatalf("deployed rule %s carries no role-core markers", rule.Rel)
		}
		for _, r := range regions {
			if strings.TrimSpace(r) != "" {
				blocks = append(blocks, r)
			}
		}
	}
	if len(blocks) == 0 {
		t.Fatal("expected role-core block set is empty — a derived empty set asserts nothing")
	}
	return blocks
}

// blockLabel names a role-core block by its first non-empty line (regions
// begin with a newline — the markers sit on their own lines — so a raw
// firstLine would return "" and Contains(x, "") is vacuously true).
func blockLabel(b string) string {
	return firstLine(strings.TrimSpace(b))
}

// injectionMissingBlocks returns the expected blocks absent from an injected
// context (empty = the guard holds). firstLine is the package helper
// (pre_tool.go).
func injectionMissingBlocks(ctx string, blocks []string) []string {
	var missing []string
	for _, b := range blocks {
		if !strings.Contains(ctx, b) {
			missing = append(missing, blockLabel(b))
		}
	}
	return missing
}

// TestSessionStartRoleRulesInjectionPerMarkerAndSource is the AC-ALB-009
// guard: for every registry marker × every injecting source, the assembled
// injection carries EVERY role-core block of the deployed rule files. Each
// named PASS line is one marker × source cell.
func TestSessionStartRoleRulesInjectionPerMarkerAndSource(t *testing.T) {
	clearFactoryEnv(t)
	root := t.TempDir()
	writeDeployedRoleRules(t, root)
	blocks := roleCoreBlocksFromDeployed(t, root)

	for _, marker := range roleMarkerRegistry() {
		for _, source := range []string{"startup", "clear", "compact"} {
			t.Run(marker.Name+"_"+source, func(t *testing.T) {
				t.Setenv(marker.EnvKey, "1")
				inj := roleRuleInjectionFor(root, source, "")
				if inj.Context == "" {
					t.Fatalf("no injection for role %s on source %s", marker.Name, source)
				}
				if missing := injectionMissingBlocks(inj.Context, blocks); len(missing) > 0 {
					t.Fatalf("injection for %s/%s misses %d of %d role-core blocks; first missing: %q",
						marker.Name, source, len(missing), len(blocks), missing[0])
				}
				t.Logf("PASS %s x %s: injection carries all %d derived role-core blocks (%d chars)",
					marker.Name, source, len(blocks), utf16Len(inj.Context))
			})
		}
	}
}

// TestSessionStartRoleRulesNoInjection is AC-ALB-010 / REQ-ALB-008: an
// unmarked environment injects nothing, and a marked session on resume
// injects nothing.
func TestSessionStartRoleRulesNoInjection(t *testing.T) {
	clearFactoryEnv(t)
	root := t.TempDir()
	writeDeployedRoleRules(t, root)

	t.Run("unmarked_startup", func(t *testing.T) {
		inj := roleRuleInjectionFor(root, "startup", "")
		if inj.Context != "" || inj.OperatorNotice != "" {
			t.Errorf("unmarked startup injected: context=%dB notice=%q", len(inj.Context), inj.OperatorNotice)
		}
	})
	t.Run("marked_resume", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		inj := roleRuleInjectionFor(root, "resume", "")
		if inj.Context != "" || inj.OperatorNotice != "" {
			t.Errorf("resume injected: context=%dB notice=%q", len(inj.Context), inj.OperatorNotice)
		}
	})
	t.Run("unknown_source_fork", func(t *testing.T) {
		// The source set grows (fork arrived in 2.1.214); gate on the values
		// wanted — an unknown source is not an injecting source.
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-3")
		inj := roleRuleInjectionFor(root, "fork", "")
		if inj.Context != "" || inj.OperatorNotice != "" {
			t.Errorf("fork injected: context=%dB notice=%q", len(inj.Context), inj.OperatorNotice)
		}
	})
}

// TestSessionStartRoleRulesFailVisible is AC-ALB-011 / REQ-ALB-009: for the
// absent, unreadable, empty, and unmarked fixtures the hook emits the
// operator-visible warning AND the agent read directive TOGETHER — never a
// silent start.
func TestSessionStartRoleRulesFailVisible(t *testing.T) {
	clearFactoryEnv(t)
	dispatch := roleRuleFiles[0]
	messaging := roleRuleFiles[1]

	build := func(t *testing.T) string {
		root := t.TempDir()
		// Both rules start present (small cores keep the failure path clear
		// of the size gate); each fixture then breaks exactly one of them.
		writeRoleRuleFixture(t, root, dispatch, smallMarkedRule("dispatch"))
		writeRoleRuleFixture(t, root, messaging, smallMarkedRule("messaging"))
		return root
	}

	t.Run("absent_file", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		removeRoleRule(t, root, dispatch)
		inj := roleRuleInjectionFor(root, "startup", "")
		assertFailVisible(t, inj, "absent")
	})
	t.Run("unreadable_file", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root: chmod-based unreadable fixture cannot bind")
		}
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		path := filepath.Join(root, filepath.FromSlash(dispatch.Rel))
		if err := os.Chmod(path, 0o000); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
		inj := roleRuleInjectionFor(root, "startup", "")
		assertFailVisible(t, inj, "unreadable")
	})
	t.Run("empty_file", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch, "")
		inj := roleRuleInjectionFor(root, "startup", "")
		assertFailVisible(t, inj, "empty")
	})
	t.Run("unmarked_file", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch, "# Rule\n\nbody without any role-core markers\n")
		inj := roleRuleInjectionFor(root, "startup", "")
		assertFailVisible(t, inj, "unmarked")
	})
	t.Run("unclosed_region", func(t *testing.T) {
		// A start marker without its closing pair must fail visible, not
		// pass as a legitimate empty core (the required rules would vanish
		// from the session without any warning).
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch, "<!-- moai:role-core-start -->\nrequired core body with the end marker missing\n")
		inj := roleRuleInjectionFor(root, "startup", "")
		assertFailVisible(t, inj, "unclosed_region")
	})
	t.Run("end_before_start", func(t *testing.T) {
		// Balanced counts alone still pass an END-before-START file: the
		// extractor scans forward from the FIRST start marker, so this file
		// yielded an empty core with err=nil and the caller read it as the
		// legitimate empty-pair state — the required rule vanished without a
		// warning. Marker order must fail visible.
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch,
			"<!-- moai:role-core-end -->\n\ntext before any start marker\n\n<!-- moai:role-core-start -->\nrequired core body with no end marker after it\n")
		inj := roleRuleInjectionFor(root, "startup", "")
		assertFailVisible(t, inj, "end_before_start")
	})
}

// assertFailVisible asserts the REQ-ALB-009 pair: operator warning AND agent
// read directive together, and no silent partial core.
func assertFailVisible(t *testing.T, inj roleRuleInjection, fixture string) {
	t.Helper()
	if inj.OperatorNotice == "" {
		t.Errorf("%s fixture: no operator-visible warning (silent failure)", fixture)
	}
	if !strings.Contains(inj.Context, roleRuleFiles[0].Rel) || !strings.Contains(inj.Context, roleRuleFiles[1].Rel) {
		t.Errorf("%s fixture: read directive does not name both rule files: %q", fixture, inj.Context)
	}
	if !strings.Contains(inj.OperatorNotice, roleRuleFiles[0].Name) {
		t.Errorf("%s fixture: warning does not name the failing rule: %q", fixture, inj.OperatorNotice)
	}
}

// TestSessionStartRoleRulesSizeGate is AC-ALB-012: (a) an assembled context
// at or under the cap delivers the core directly; (b) over the cap the core
// goes out INTACT (zero truncated units) with the operator warning and the
// overflow directive; (c) overflow delivery unavailable retreats to the
// REQ-ALB-009 warning + read directive with no core.
func TestSessionStartRoleRulesSizeGate(t *testing.T) {
	clearFactoryEnv(t)
	messaging := roleRuleFiles[1]

	smallRoot := func(t *testing.T) string {
		root := t.TempDir()
		writeRoleRuleFixture(t, root, roleRuleFiles[0], smallMarkedRule("dispatch"))
		writeRoleRuleFixture(t, root, messaging, smallMarkedRule("messaging"))
		return root
	}

	t.Run("a_under_cap_delivers_core_directly", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := smallRoot(t)
		inj := roleRuleInjectionFor(root, "startup", "")
		if inj.OperatorNotice != "" {
			t.Errorf("under-cap injection raised an operator notice: %q", inj.OperatorNotice)
		}
		if !strings.Contains(inj.Context, "Small binding block for dispatch") {
			t.Errorf("under-cap injection lost the dispatch core: %q", inj.Context)
		}
		if strings.Contains(inj.Context, "NOTE: the output above exceeds") {
			t.Errorf("under-cap injection carries the overflow directive: %q", inj.Context)
		}
	})

	t.Run("b_over_cap_core_intact", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := t.TempDir()
		writeDeployedRoleRules(t, root) // real files: the dispatch core alone exceeds the cap
		blocks := roleCoreBlocksFromDeployed(t, root)

		existing := "session attribution line\n"
		inj := roleRuleInjectionFor(root, "startup", existing)
		total := utf16Len(existing) + utf16Len("\n\n"+inj.Context)
		if total <= roleRulesContextLimit {
			t.Fatalf("fixture not in the overflow regime: total=%d cap=%d", total, roleRulesContextLimit)
		}
		if missing := injectionMissingBlocks(inj.Context, blocks); len(missing) > 0 {
			t.Fatalf("over-cap emission truncated units: %d of %d blocks missing; first: %q",
				len(missing), len(blocks), missing[0])
		}
		if !strings.Contains(inj.Context, "preview of the first 2,000 characters") {
			t.Errorf("over-cap emission lacks the overflow directive")
		}
		if inj.OperatorNotice == "" || !strings.Contains(inj.OperatorNotice, "intact") {
			t.Errorf("over-cap emission lacks the operator overflow warning: %q", inj.OperatorNotice)
		}
		t.Logf("PASS over-cap intact: %d blocks x full text in a %d-unit output (cap %d)",
			len(blocks), total, roleRulesContextLimit)
	})

	t.Run("c_overflow_unavailable_falls_back", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		orig := roleRulesOverflowDelivery
		roleRulesOverflowDelivery = func() bool { return false }
		t.Cleanup(func() { roleRulesOverflowDelivery = orig })

		root := t.TempDir()
		writeDeployedRoleRules(t, root)
		blocks := roleCoreBlocksFromDeployed(t, root)

		inj := roleRuleInjectionFor(root, "startup", "")
		if inj.OperatorNotice == "" {
			t.Errorf("fallback emitted no operator warning")
		}
		if !strings.Contains(inj.Context, roleRuleFiles[0].Rel) {
			t.Errorf("fallback context is not the read directive: %q", inj.Context)
		}
		for i, b := range blocks {
			if strings.Contains(inj.Context, blockLabel(b)) {
				t.Errorf("fallback carried core block %d despite unavailable delivery", i)
				break
			}
		}
		t.Logf("PASS overflow-unavailable: warning + read directive, 0 core blocks of %d emitted", len(blocks))
	})
}

// TestSessionStartRoleRulesCoverageGuard is REQ-ALB-011 / AC-ALB-013: the
// guarded marker set derives from the registry (a fixture registry entry is
// covered by the same sweep), and a mutation that deletes one role-core
// block from the assembled injection demonstrably FAILS the guard.
func TestSessionStartRoleRulesCoverageGuard(t *testing.T) {
	clearFactoryEnv(t)
	root := t.TempDir()
	writeDeployedRoleRules(t, root)
	blocks := roleCoreBlocksFromDeployed(t, root)

	// Registry-derived sweep: production registry plus a fixture entry. The
	// subtest name shows the marker, so a new registry entry becomes a named
	// covered cell without touching this test.
	fixtureEntry := config.RoleMarker{Name: "factory-leader-nightly", EnvKey: "MOAI_TEST_FACTORY_ROLE"}
	orig := roleMarkerRegistry
	roleMarkerRegistry = func() []config.RoleMarker {
		return append(config.RoleMarkerRegistry(), fixtureEntry)
	}
	t.Cleanup(func() { roleMarkerRegistry = orig })

	for _, marker := range roleMarkerRegistry() {
		t.Run("guard_covers_"+marker.Name, func(t *testing.T) {
			t.Setenv(marker.EnvKey, "1")
			inj := roleRuleInjectionFor(root, "startup", "")
			if missing := injectionMissingBlocks(inj.Context, blocks); len(missing) > 0 {
				t.Fatalf("registry entry %s not covered: missing %q", marker.Name, missing[0])
			}
			t.Logf("PASS guard covers registry entry %s (all %d blocks)", marker.Name, len(blocks))
		})
	}

	// Mutation: delete one role-core block from the assembled injection —
	// the guard must FAIL naming that block (observed failure, not assumed).
	t.Run("mutation_delete_one_block_fails", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		inj := roleRuleInjectionFor(root, "startup", "")
		victim := blocks[len(blocks)/2]
		mutated := strings.Replace(inj.Context, victim, "", 1)
		if mutated == inj.Context {
			t.Fatal("mutation did not change the injected context")
		}
		missing := injectionMissingBlocks(mutated, blocks)
		if len(missing) == 0 {
			t.Fatal("mutation observed green: deleting one role-core block was NOT caught by the guard")
		}
		if !strings.Contains(missing[0], firstLine(victim)[:min(40, len(firstLine(victim)))]) {
			t.Errorf("guard failure does not name the deleted block: %q", missing[0])
		}
		t.Logf("mutation observed RED: deleted block reported missing (%d missing of %d)", len(missing), len(blocks))
	})
}

// TestSessionStartRoleRulesHandleIntegration runs the injection through the
// full SessionStart handler: leader + startup appends the core (and, over
// the cap, the operator warning); resume appends nothing.
func TestSessionStartRoleRulesHandleIntegration(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv("ANTHROPIC_BASE_URL", "")

	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeDeployedRoleRules(t, projectDir)
	blocks := roleCoreBlocksFromDeployed(t, projectDir)

	run := func(source string) (context, notice string) {
		h := NewSessionStartHandler(nil)
		input := &HookInput{
			SessionID:  "role-rules-integration-" + source,
			ProjectDir: projectDir,
			CWD:        projectDir,
			Source:     source,
		}
		out, err := h.Handle(t.Context(), input)
		if err != nil {
			t.Fatalf("Handle(%s): %v", source, err)
		}
		ctx := ""
		if out.HookSpecificOutput != nil {
			ctx = out.HookSpecificOutput.AdditionalContext
		}
		return ctx, out.SystemMessage
	}

	t.Run("leader_startup_injects_through_handle", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		ctx, notice := run("startup")
		if missing := injectionMissingBlocks(ctx, blocks); len(missing) > 0 {
			t.Fatalf("Handle injection misses role-core blocks; first: %q", missing[0])
		}
		if notice == "" {
			t.Errorf("over-cap Handle emission carried no operator warning")
		}
		t.Logf("PASS Handle integration (startup): all %d blocks in additionalContext, operator warned", len(blocks))
	})

	t.Run("leader_resume_no_injection", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		ctx, notice := run("resume")
		for i, b := range blocks {
			if strings.Contains(ctx, blockLabel(b)) {
				t.Fatalf("resume carried core block %d through Handle", i)
			}
		}
		if strings.Contains(notice, "Role-rule") {
			t.Errorf("resume raised a role-rule operator notice: %q", notice)
		}
	})
}
