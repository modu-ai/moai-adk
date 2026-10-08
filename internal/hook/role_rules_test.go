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
		rel := rule.Rel
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil && strings.HasPrefix(rel, ".claude/rules/moai/") {
			// Codex-only deployment shape: the rules install under
			// .moai/policies. The helper projects the same way the deployer
			// does so the expected block set stays derived in that layout.
			alt := ".moai/policies/" + strings.TrimPrefix(rel, ".claude/rules/moai/")
			data, err = os.ReadFile(filepath.Join(root, filepath.FromSlash(alt)))
			rel = alt
		}
		if err != nil {
			t.Fatalf("read deployed rule %s: %v", rel, err)
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
				inj := roleRuleInjectionFor(root, source, "", langEnglish)
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
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		if inj.Context != "" || inj.OperatorNotice != "" {
			t.Errorf("unmarked startup injected: context=%dB notice=%q", len(inj.Context), inj.OperatorNotice)
		}
	})
	t.Run("marked_resume", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		inj := roleRuleInjectionFor(root, "resume", "", langEnglish)
		if inj.Context != "" || inj.OperatorNotice != "" {
			t.Errorf("resume injected: context=%dB notice=%q", len(inj.Context), inj.OperatorNotice)
		}
	})
	t.Run("unknown_source_fork", func(t *testing.T) {
		// The source set grows (fork arrived in 2.1.214); gate on the values
		// wanted — an unknown source is not an injecting source.
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-3")
		inj := roleRuleInjectionFor(root, "fork", "", langEnglish)
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
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		assertFailVisible(t, inj, "absent")
	})
	t.Run("unreadable_file", func(t *testing.T) {
		// Platform-independent read-failure fixture: os.ReadFile on a
		// directory fails on every platform (EISDIR), so the fixture binds
		// on Windows CI too — a chmod 0o000 fixture does not remove read
		// permission there.
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		path := filepath.Join(root, filepath.FromSlash(dispatch.Rel))
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove: %v", err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		assertFailVisible(t, inj, "unreadable")
	})
	t.Run("empty_file", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch, "")
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		assertFailVisible(t, inj, "empty")
	})
	t.Run("unmarked_file", func(t *testing.T) {
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch, "# Rule\n\nbody without any role-core markers\n")
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		assertFailVisible(t, inj, "unmarked")
	})
	t.Run("unclosed_region", func(t *testing.T) {
		// A start marker without its closing pair must fail visible, not
		// pass as a legitimate empty core (the required rules would vanish
		// from the session without any warning).
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		root := build(t)
		writeRoleRuleFixture(t, root, dispatch, "<!-- moai:role-core-start -->\nrequired core body with the end marker missing\n")
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
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
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		assertFailVisible(t, inj, "end_before_start")
	})
}

// TestSessionStartRoleRulesRootFollowsSessionCWD pins the role-rules ROOT
// resolution: the injection serves the role rules of the tree the session
// runs in (the hook input's CWD), not the tree CLAUDE_PROJECT_DIR points at
// — in a worktree session CWD is the worktree while ProjectDir stays the
// primary checkout, and the ProjectDir-first resolution delivered the
// primary tree's rules. Two trees carry DIFFERENT marked rule content; the
// injected context must carry the CWD tree's text only.
func TestSessionStartRoleRulesRootFollowsSessionCWD(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv("ANTHROPIC_BASE_URL", "")

	primary := t.TempDir()
	worktree := t.TempDir()
	for _, d := range []string{primary, worktree} {
		if err := os.MkdirAll(filepath.Join(d, ".moai", "state"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	mark := func(root, text string) {
		t.Helper()
		writeRoleRuleFixture(t, root, roleRuleFiles[0], smallMarkedRule(text))
		writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))
	}
	mark(primary, "PRIMARY-TREE-CORE")
	mark(worktree, "WORKTREE-TREE-CORE")

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:  "role-rules-cwd-root",
		ProjectDir: primary,
		CWD:        worktree,
		Source:     "startup",
	}
	out, err := h.Handle(t.Context(), input)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	ctx := ""
	if out.HookSpecificOutput != nil {
		ctx = out.HookSpecificOutput.AdditionalContext
	}
	if !strings.Contains(ctx, "WORKTREE-TREE-CORE") {
		t.Fatalf("injection does not carry the session CWD (worktree) tree's role core:\n%s", ctx)
	}
	if strings.Contains(ctx, "PRIMARY-TREE-CORE") {
		t.Fatalf("injection carried the project (primary) tree's role core:\n%s", ctx)
	}
	t.Logf("PASS: injection serves the session CWD tree's role core (CWD %q), primary %q excluded", worktree, primary)
}

// assertFailVisible asserts the REQ-ALB-009 pair: operator warning AND agent
// read directive together, and no silent partial core.
func assertFailVisible(t *testing.T, inj roleRuleInjection, fixture string) {
	t.Helper()
	if inj.OperatorNotice == "" {
		t.Errorf("%s fixture: no operator-visible warning (silent failure)", fixture)
	}
	if !strings.Contains(slashNorm(inj.RecoveryHead), roleRuleFiles[0].Rel) || !strings.Contains(slashNorm(inj.RecoveryHead), roleRuleFiles[1].Rel) {
		t.Errorf("%s fixture: read directive does not name both rule files: %q", fixture, inj.RecoveryHead)
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
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
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
		inj := roleRuleInjectionFor(root, "startup", existing, langEnglish)
		total := utf16Len(existing) + utf16Len("\n\n"+inj.Context)
		if total <= roleRulesContextLimit {
			t.Fatalf("fixture not in the overflow regime: total=%d cap=%d", total, roleRulesContextLimit)
		}
		if missing := injectionMissingBlocks(inj.Context, blocks); len(missing) > 0 {
			t.Fatalf("over-cap emission truncated units: %d of %d blocks missing; first: %q",
				len(missing), len(blocks), missing[0])
		}
		if !strings.Contains(inj.RecoveryHead, "preview of the first 2,000 characters") {
			t.Errorf("over-cap emission lacks the overflow directive as the recovery head")
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

		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
		if inj.OperatorNotice == "" {
			t.Errorf("fallback emitted no operator warning")
		}
		if !strings.Contains(slashNorm(inj.RecoveryHead), roleRuleFiles[0].Rel) {
			t.Errorf("fallback recovery head is not the read directive: %q", inj.RecoveryHead)
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
			inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
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
		inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
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

// TestSessionStartRoleRulesOperatorLocales observes the operator-facing
// warnings rendering in the settings conversation language (ko/ja/zh/en
// table): each locale's failure and overflow warnings are non-empty,
// locale-distinct, and carry the deciding figures (role name, total, cap).
func TestSessionStartRoleRulesOperatorLocales(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	overCapRoot := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		writeDeployedRoleRules(t, root) // real files: the core exceeds the cap
		return root
	}

	for _, lang := range []string{"ko", "ja", "zh", "en"} {
		t.Run("overflow_"+lang, func(t *testing.T) {
			inj := roleRuleInjectionFor(overCapRoot(t), "startup", "", lang)
			if inj.OperatorNotice == "" {
				t.Fatalf("locale %s: overflow warning empty", lang)
			}
			if !strings.Contains(inj.OperatorNotice, "factory-leader") {
				t.Errorf("locale %s: warning missing the role name: %q", lang, inj.OperatorNotice)
			}
			t.Logf("PASS overflow %s: %q", lang, inj.OperatorNotice)
		})
		t.Run("failure_"+lang, func(t *testing.T) {
			root := t.TempDir()
			inj := roleRuleInjectionFor(root, "startup", "", lang)
			if inj.OperatorNotice == "" {
				t.Fatalf("locale %s: failure warning empty", lang)
			}
			if !strings.Contains(inj.OperatorNotice, "factory-leader") {
				t.Errorf("locale %s: warning missing the role name: %q", lang, inj.OperatorNotice)
			}
			t.Logf("PASS failure %s: %q", lang, inj.OperatorNotice)
		})
	}

	t.Run("unknown_locale_falls_back_to_english", func(t *testing.T) {
		inj := roleRuleInjectionFor(t.TempDir(), "startup", "", "zz")
		if !strings.Contains(inj.OperatorNotice, "Role-rule injection failed") {
			t.Errorf("unknown locale did not fall back to English: %q", inj.OperatorNotice)
		}
	})
}

// TestSessionStartRoleRulesRootStopsAtProjectBoundary pins the walk
// boundary: a lane worktree (its own git-dir pointer file) nested UNDER a
// tree that carries the rules must NOT inherit the parent tree's rules —
// the walk stops at the worktree root, the injection takes the REQ-ALB-009
// missing path, and no parent-tree block reaches the session.
func TestSessionStartRoleRulesRootStopsAtProjectBoundary(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	parent := t.TempDir()
	writeDeployedRoleRules(t, parent)
	lane := filepath.Join(parent, ".claude", "worktrees", "lane-2")
	if err := os.MkdirAll(lane, 0o755); err != nil {
		t.Fatalf("mkdir lane: %v", err)
	}
	// The worktree's repository pointer is a FILE (the git worktree shape).
	if err := os.WriteFile(filepath.Join(lane, ".git"), []byte("gitdir: /somewhere/else\n"), 0o644); err != nil {
		t.Fatalf("write repository pointer: %v", err)
	}

	resolved := roleRulesRootFromCWD(filepath.Join(lane, "internal", "deep"))
	if resolved != lane {
		t.Fatalf("lane cwd resolved to %q, want the lane tree itself %q (no parent-tree reach)", resolved, lane)
	}
	inj := roleRuleInjectionFor(resolved, "startup", "", langEnglish)
	if inj.OperatorNotice == "" {
		t.Fatal("lane tree without the rules injected silently — the REQ-ALB-009 missing path did not fire")
	}
	if !strings.Contains(inj.OperatorNotice, "absent") {
		t.Errorf("missing warning does not name the absent rule: %q", inj.OperatorNotice)
	}
	for _, b := range roleCoreBlocksFromDeployed(t, parent) {
		if strings.Contains(inj.Context, blockLabel(b)) {
			t.Fatalf("parent-tree block leaked into the lane session: %q", blockLabel(b))
		}
	}
	t.Logf("PASS lane worktree without rules: root=%q, missing warning fired, 0 parent blocks delivered", resolved)
}

// TestSessionStartRoleRulesRootStopsAtNestedMoAIProject is the nested-
// project boundary: a nested MoAI project carries its own `.moai/` but no
// `.git` — the walk must treat that directory as a project root too. A
// child whose rules are missing takes the REQ-ALB-009 missing path; the
// parent project's rules (one level up) are never injected silently.
func TestSessionStartRoleRulesRootStopsAtNestedMoAIProject(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	parent := t.TempDir()
	writeDeployedRoleRules(t, parent)
	child := filepath.Join(parent, "examples", "nested-moai-project")
	if err := os.MkdirAll(filepath.Join(child, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir child .moai: %v", err)
	}
	// No .git in the child — the boundary is the .moai directory itself.

	resolved := roleRulesRootFromCWD(filepath.Join(child, "docs"))
	if resolved != child {
		t.Fatalf("nested-project cwd resolved to %q, want the child project root %q (no parent-tree reach)", resolved, child)
	}
	inj := roleRuleInjectionFor(resolved, "startup", "", langEnglish)
	if inj.OperatorNotice == "" {
		t.Fatal("nested project without the rules injected silently — the REQ-ALB-009 missing path did not fire")
	}
	for _, b := range roleCoreBlocksFromDeployed(t, parent) {
		if strings.Contains(inj.Context, blockLabel(b)) {
			t.Fatalf("parent-tree block leaked into the nested project: %q", blockLabel(b))
		}
	}
	t.Logf("PASS nested MoAI project without rules: root=%q, missing warning fired, 0 parent blocks delivered", resolved)
}

// TestSessionStartRoleRulesMarkerSequence is the full sequence validation: a
// START END END START file passes the count balance AND the first-marker
// checks, yet its second start region is never closed — the injected core
// would silently drop it. The sequence must fail visible.
func TestSessionStartRoleRulesMarkerSequence(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0],
		config.RoleCoreMarkerStart+"\nclosed block\n"+config.RoleCoreMarkerEnd+"\n"+config.RoleCoreMarkerEnd+"\n"+config.RoleCoreMarkerStart+"\nunclosed tail block\n")
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))
	inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
	assertFailVisible(t, inj, "marker_sequence")
}

// TestSessionStartRoleRulesCodexOnlyDeploymentLayout is the Codex-only
// deployment shape: the rules install under .moai/policies (the
// harness-neutral projection of .claude/rules/moai) and no .claude/rules
// tree exists. A session there must receive the normal core — not the
// REQ-ALB-009 absent warning plus a read directive naming a path that
// does not exist in its tree.
func TestSessionStartRoleRulesCodexOnlyDeploymentLayout(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	root := t.TempDir()
	for _, rule := range roleRuleFiles {
		alt := rule.Rel
		if strings.HasPrefix(alt, ".claude/rules/moai/") {
			alt = ".moai/policies/" + strings.TrimPrefix(alt, ".claude/rules/moai/")
		}
		writeRoleRuleFixture(t, root, roleRuleFile{Rel: alt, Name: rule.Name}, smallMarkedRule("dispatch"))
	}
	blocks := roleCoreBlocksFromDeployed(t, root)

	inj := roleRuleInjectionFor(root, "startup", "", langEnglish)
	if inj.OperatorNotice != "" {
		t.Fatalf("Codex-only layout raised an operator notice instead of delivering the core: %q", inj.OperatorNotice)
	}
	if missing := injectionMissingBlocks(inj.Context, blocks); len(missing) > 0 {
		t.Fatalf("Codex-only layout injection misses %d of %d blocks; first: %q", len(missing), len(blocks), missing[0])
	}
	// The recovery directive names the paths the rules are actually deployed
	// at — in this tree, the .moai/policies projections.
	directive := roleRulesReadDirective(root, "")
	if !strings.Contains(slashNorm(directive), ".moai/policies/workflow/factory-dispatch.md") ||
		!strings.Contains(slashNorm(directive), ".moai/policies/workflow/cross-session-messaging.md") {
		t.Errorf("Codex-only layout recovery directive does not name the deployed policies paths: %q", directive)
	}
	t.Logf("PASS Codex-only deployment: all %d blocks delivered from .moai/policies, recovery directive names the deployed paths", len(blocks))
}

// TestSessionStartRoleRulesDirectivePathsResolveFromDisk pins the recovery
// directive's path form: a session whose cwd sits in a subdirectory receives
// paths that EXIST on disk — the directive joins the names to the resolved
// root instead of emitting cwd-relative fragments that resolve to nothing.
func TestSessionStartRoleRulesDirectivePathsResolveFromDisk(t *testing.T) {
	clearFactoryEnv(t)
	root := t.TempDir()
	writeDeployedRoleRules(t, root)

	directive := roleRulesReadDirective(root, "")
	paths := directiveBacktickedPaths(directive)
	if len(paths) < 2 {
		t.Fatalf("directive names %d paths, want both rule files: %q", len(paths), directive)
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("directive path %q does not exist on disk (cwd-relative fragment?): %v", p, err)
		}
	}
	t.Logf("PASS recovery directive: %d named paths all resolve on disk (first: %q)", len(paths), paths[0])
}

// directiveBacktickedPaths extracts the backticked path tokens of the
// directive text.
func directiveBacktickedPaths(s string) []string {
	var out []string
	for {
		i := strings.Index(s, "`")
		if i < 0 {
			return out
		}
		s = s[i+1:]
		j := strings.Index(s, "`")
		if j < 0 {
			return out
		}
		out = append(out, s[:j])
		s = s[j+1:]
	}
}

// slashNorm normalizes Windows backslash separators to forward slashes so
// path-shape assertions hold on every platform: the recovery directive names
// paths via filepath.Join, whose separator is platform detail — the
// assertion cares about which FILE is named, not the separator.
func slashNorm(s string) string {
	return strings.ReplaceAll(s, "\\", "/")
}

// TestSessionStartRoleRulesRootFromSubdirectoryCWD is the subdirectory-CWD
// root resolution: a session whose cwd sits DEEP inside the tree (e.g.
// internal/deep) still receives the core — the root is resolved by walking
// the cwd's ancestors outward until the deployed rule files appear, not
// taken as the raw cwd. Both a root-level and a nested cwd deliver the same
// blocks.
func TestSessionStartRoleRulesRootFromSubdirectoryCWD(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	root := t.TempDir()
	writeDeployedRoleRules(t, root)
	blocks := roleCoreBlocksFromDeployed(t, root)
	nested := filepath.Join(root, "internal", "deep", "worker")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	resolved := roleRulesRootFromCWD(nested)
	if resolved != root {
		t.Fatalf("nested cwd %q resolved to root %q, want %q", nested, resolved, root)
	}
	inj := roleRuleInjectionFor(resolved, "startup", "", langEnglish)
	if missing := injectionMissingBlocks(inj.Context, blocks); len(missing) > 0 {
		t.Fatalf("nested-cwd injection misses %d of %d blocks; first: %q", len(missing), len(blocks), missing[0])
	}
	t.Logf("PASS nested cwd %q resolved to %q; all %d blocks delivered", nested, resolved, len(blocks))
}

// TestSessionStartRoleRulesOverflowDirectiveSurvivesPrefixCut is the
// save-failure scenario at the COMPOSITE level: when the runtime's
// oversized-output save fails it delivers only the FIRST 10,000 characters
// of the final additionalContext — so with a full 10,000-character prior
// context, an appended directive would start at position 10,000 and be cut.
// The recovery directive opens the COMPOSITE (position 0) — observed here.
func TestSessionStartRoleRulesOverflowDirectiveSurvivesPrefixCut(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	root := t.TempDir()
	writeDeployedRoleRules(t, root)

	const existingFull = "existing producer context that alone fills the whole delivery window."
	existing := existingFull[:1] + strings.Repeat("producer context line — fills the delivery window. ", 220) // ≥ 10,000 chars
	if utf16Len(existing) < roleRulesContextLimit {
		t.Fatalf("fixture existing context %d < cap %d — not in the cut regime", utf16Len(existing), roleRulesContextLimit)
	}

	inj := roleRuleInjectionFor(root, "startup", existing, langEnglish)
	if inj.RecoveryHead == "" {
		t.Fatal("over-cap injection carries no recovery head")
	}
	composite := assembleInjectionComposite(existing, inj)
	const prefix = 10000
	head := composite
	if len(head) > prefix {
		head = head[:prefix]
	}
	if !strings.HasPrefix(composite, "NOTE: the output above exceeds") {
		t.Fatalf("composite does not OPEN with the recovery directive (starts %.120q) — a save failure with a full prior context delivers no directive", composite[:min(120, len(composite))])
	}
	if !strings.Contains(head, "NOTE: the output above exceeds") {
		t.Fatalf("composite's first %d characters carry no read directive", prefix)
	}
	t.Logf("PASS save-failure prefix cut at composite level: composite %d chars, directive opens it, prior context %d chars", len(composite), utf16Len(existing))
}

// TestSessionStartRoleRulesCompositeNilGuard is the panic regression: a
// compact event WITH the role marker and NO prior producer context leaves
// out.HookSpecificOutput nil — the composite-head assignment dereferenced
// it and panicked (the warning and the read directive were lost with the
// session-start output). The handler must create the struct first; the
// fixture observes warning + directive delivered, no panic.
func TestSessionStartRoleRulesCompositeNilGuard(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv("ANTHROPIC_BASE_URL", "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID: "role-rules-nil-guard",
		CWD:       "", // no cwd: the root falls back to ProjectDir, also empty — fail-visible path
		Source:    "compact",
	}
	out, err := h.Handle(t.Context(), input)
	if err != nil {
		t.Fatalf("Handle panicked or errored: %v", err)
	}
	if out == nil || out.SystemMessage == "" {
		t.Fatal("nil-guard composite delivered no operator warning")
	}
	if out.HookSpecificOutput == nil || !strings.Contains(slashNorm(out.HookSpecificOutput.AdditionalContext), roleRuleFiles[0].Rel) {
		t.Fatalf("nil-guard composite delivered no read directive: %+v", out.HookSpecificOutput)
	}
	t.Logf("PASS nil guard: compact + role marker + no prior context delivered warning + directive, no panic")
}
