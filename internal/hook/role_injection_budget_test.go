// role_injection_budget_test.go — SPEC-ROLE-INJECTION-BUDGET-001 AC-RIB-001
// (REQ-RIB-001/002): the per-role assembled SessionStart composite stays at
// or under 9,000 UTF-16 code units for every registry role, within the
// bounded-source envelope of spec.md §A.5 (startup · clear without a pending
// handoff · compact without an armed goal), with the 10,000-unit delivery cap
// unchanged (decision-index Q4).
//
// The test measures BOTH deployment faces of the template SSOT: the template
// tree (internal/template/templates/.claude/rules/...) and this repository's
// own deployed copy (.claude/rules/...). Assembly uses the PRODUCTION path —
// buildRoleCore + roleRuleInjectionFor + assembleInjectionComposite — so the
// formula is never duplicated (plan.md M1 item 1: package hook for direct
// buildRoleCore access, assembly-formula reuse).
//
// Budget chain (spec.md §B, code-cited): composite = producers + "\n\n" +
// header + core + "\n\n" + pointer (assembleInjectionComposite role_rules.go
// RecoveryHead-less branch + roleRuleInjectionFor :373-380 — the header's
// trailing "\n\n" is inside the format string, core follows directly).
// Overhead = joiner 2 + header 125 (leader, worst) + joiner 2 + pointer 128 =
// 257. Producer ceiling 4,797 = gate-measured composite 23,166 − this-tree
// role block 18,369 (= 2 + 123 + 18,114 + 2 + 128) — the gate-implied bound
// derived in the plan-phase audit (F1 repair); startup is the structurally
// heaviest bounded source (factoryBootstrapNoticeForSource is startup-only,
// session_start.go:541-580 — spec.md REQ-RIB-002). Core ceiling 3,946 =
// 9,000 − 4,797 − 257 (leader-binding; lane 3,948). Design target 3,800.
package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// budgetProducerCeiling is the gate-implied producer bound (UTF-16): the
	// conservative worst case inside the bounded-source envelope. Derived at
	// plan time: gate-measured composite 23,166 − role block 18,369. The 421
	// gap against the direct lane-15 startup decomposition (4,376) is
	// absorbed by this larger side (spec.md §B) — a live-session asset that
	// cannot be re-measured.
	budgetProducerCeiling = 4797
	// budgetOverhead is the leader-worst role-block overhead: joiner 2 +
	// header 125 (incl. its trailing "\n\n") + joiner 2 + pointer 128.
	budgetOverhead = 257
	// budgetCompositeLimit is the assembled-composite budget (REQ-RIB-001):
	// the assembled reading of 「역할별 조립 ≤9,000자」 (spec.md §A.4).
	budgetCompositeLimit = 9000
	// budgetCoreCeiling is the derived role-core ceiling: 9,000 − 4,797 − 257
	// (leader-binding; the lane header is 2 units smaller → 3,948).
	budgetCoreCeiling = 3946
	// budgetCoreTarget is the design target (budget − 146 headroom); a
	// diagnostic log line, not an assertion.
	budgetCoreTarget = 3800
)

// budgetModuleRoot resolves the repository root from this test file's
// location (runtime.Caller — the SessionStart hook may run from any cwd).
func budgetModuleRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate the module root")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// budgetRuleTreeRoot resolves the project root the deployed rule files
// resolve under for one deployment face: the module root itself (the
// deployed copy) or the template tree root (internal/template/templates).
func budgetRuleTreeRoot(t *testing.T, root string, face string) string {
	t.Helper()
	if face == "template" {
		return filepath.Join(root, "internal", "template", "templates")
	}
	return root
}

// budgetCompositeError renders the auditable breakdown a failure names
// (REQ-RIB-002: producers, joiner, header, core, pointer, total).
func budgetCompositeError(face, role string, producers, joiner, header, core, pointer, total int) error {
	return fmt.Errorf(
		"%s tree: assembled composite for role %s exceeds the %d-unit budget — breakdown (UTF-16): producers=%d joiner=%d header=%d core=%d pointer=%d total=%d (core ceiling %d, design target %d, producer ceiling %d, overhead %d)",
		face, role, budgetCompositeLimit, producers, joiner, header, core, pointer, total, budgetCoreCeiling, budgetCoreTarget, budgetProducerCeiling, budgetOverhead)
}

// TestRoleInjectionAssemblyBudget assembles every registry role's composite
// through the production injection path against both deployment faces and
// asserts composite ≤ 9,000 with a diagnostic sub-assertion on the core
// ceiling (AC-RIB-001; E1 = template tree, E2 = deployed tree).
func TestRoleInjectionAssemblyBudget(t *testing.T) {
	moduleRoot := budgetModuleRoot(t)

	faces := []struct {
		name string
		root string
	}{
		{"template", budgetRuleTreeRoot(t, moduleRoot, "template")},
		{"deployed", budgetRuleTreeRoot(t, moduleRoot, "deployed")},
	}

	for _, face := range faces {
		for _, marker := range config.RoleMarkerRegistry() {
			t.Run(face.name+"/"+marker.Name, func(t *testing.T) {
				// The leader header is the binding-worst (125 vs lane 123);
				// the assertion scope covers every registry role either way.
				t.Setenv(marker.EnvKey, marker.Name)

				// Startup is the structurally heaviest bounded source
				// (Factory Mode join line — REQ-RIB-002); clear/compact are
				// strictly smaller inside the envelope.
				inj := roleRuleInjectionFor(face.root, "startup", "en")
				if inj.Context == "" {
					t.Fatalf("%s tree: role %s — injection produced no context (role detection or source gate failed)", face.name, marker.Name)
				}

				// Production composite assembly (assembleInjectionComposite —
				// RecoveryHead-less branch): producers + "\n\n" + role block.
				producers := strings.Repeat("a", budgetProducerCeiling) // ASCII: 1 UTF-16 unit per rune
				composite := assembleInjectionComposite(producers, inj)
				total := utf16Len(composite)

				// Breakdown, per the code-cited formula: the context is
				// header + core + ("\n\n" + pointer). The header format lives
				// at roleRuleInjectionFor (:374); the pointer renders only
				// when a rule's core is empty (cross-session-messaging.md).
				core, coreErr := buildRoleCore(face.root, roleRuleFiles[0])
				if coreErr != nil {
					t.Fatalf("%s tree: buildRoleCore failed: %v", face.name, coreErr)
				}
				coreLen := utf16Len(core)
				pointer := ""
				pointerLen := 0
				pointerPrefix := "Role rule (no role-core region"
				if idx := strings.LastIndex(inj.Context, pointerPrefix); idx >= 0 {
					pointer = inj.Context[idx:]
					pointerLen = utf16Len(pointer)
				}
				header := fmt.Sprintf("Role-gated rules (factory %s session — injected at session start; the always-loaded surface carries the stubs):\n\n", marker.Name)
				headerLen := utf16Len(header)
				joiners := 2 + pointerLenIf(pointerLen)

				if total > budgetCompositeLimit {
					t.Error(budgetCompositeError(face.name, marker.Name, budgetProducerCeiling, joiners, headerLen, coreLen, pointerLen, total))
				}

				// Diagnostic sub-assertion (REQ-RIB-002): the core ceiling is
				// where the budget is actually met — a composite pass with an
				// over-ceiling core means the constants drifted.
				if coreLen > budgetCoreCeiling {
					t.Errorf("%s tree: role core %d UTF-16 exceeds the derived ceiling %d (design target %d; producer ceiling %d + overhead %d + ceiling = %d)",
						face.name, coreLen, budgetCoreCeiling, budgetCoreTarget, budgetProducerCeiling, budgetOverhead, budgetProducerCeiling+budgetOverhead+budgetCoreCeiling)
				}
				if total <= budgetCompositeLimit && coreLen <= budgetCoreCeiling {
					t.Logf("%s tree: role %s composite=%d (budget %d) core=%d (ceiling %d, target %d)",
						face.name, marker.Name, total, budgetCompositeLimit, coreLen, budgetCoreCeiling, budgetCoreTarget)
				}
			})
		}
	}
}

// pointerLenIf keeps the joiner arithmetic explicit: the pointer joiner
// exists only when a pointer was rendered.
func pointerLenIf(pointerLen int) int {
	if pointerLen > 0 {
		return 2
	}
	return 0
}

// TestRoleInjectionBudgetConstantsExist pins the test constants' provenance
// so a reader of a failure can re-derive them (spec.md §B 재측정 규율): the
// numbers are re-measurable with the E1/E2 commands in
// acceptance.md §C against the committed tree.
func TestRoleInjectionBudgetConstantsExist(t *testing.T) {
	if budgetProducerCeiling+budgetOverhead+budgetCoreCeiling != budgetCompositeLimit {
		t.Fatalf("budget chain broken: %d + %d + %d != %d", budgetProducerCeiling, budgetOverhead, budgetCoreCeiling, budgetCompositeLimit)
	}
	// The 10,000-unit delivery cap is unchanged (Q4, unraisable) and the
	// composite budget must stay under it with headroom for the envelope's
	// bounded producer variance.
	if budgetCompositeLimit >= roleRulesContextLimit {
		t.Fatalf("composite budget %d must stay under the delivery cap %d", budgetCompositeLimit, roleRulesContextLimit)
	}
}

// TestRoleInjectionBudgetDeployedRuleTreesReadable guards the measurement
// surface itself: both deployment faces must carry a readable, marked
// dispatch rule, or the budget test above would sweep nothing (an empty
// sweep is a failure, not a pass — verification-completeness §1.1).
func TestRoleInjectionBudgetDeployedRuleTreesReadable(t *testing.T) {
	moduleRoot := budgetModuleRoot(t)
	paths := []string{
		filepath.Join(moduleRoot, filepath.FromSlash(roleRuleFiles[0].Rel)),
		filepath.Join(moduleRoot, "internal", "template", "templates", filepath.FromSlash(roleRuleFiles[0].Rel)),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("budget measurement surface unreadable: %s: %v", p, err)
		}
		if _, marked := config.ExtractRoleCoreRegions(string(data)); !marked {
			t.Fatalf("budget measurement surface unmarked: %s carries no %s", p, config.RoleCoreMarkerStart)
		}
	}
}
