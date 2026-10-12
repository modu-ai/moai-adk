// role_core_stub_gate_test.go — SPEC-ROLE-INJECTION-BUDGET-001 AC-RIB-003
// (REQ-RIB-005): a mechanical gate over every *-core.md stub in the deployed
// and template rule trees. A stub (a) carries no moai:role-core region
// marker, and (b) stays at or under 10,000 UTF-16 code units (the binding
// figure is the 10,000 budget — REQ-RIB-005; today: factory-dispatch-core
// 8,583 with 14.2% headroom, cross-session-messaging-core 8,072 with 19.3%).
//
// Enumeration is a glob over the trees, never a hand-written list, so a new
// stub is covered on arrival. The motor subtests are the adoption evidence
// (verification-completeness §1.1 observed-failure completion): fixtures
// that violate each condition are run through the same gate function and
// observed failing inside the suite.
package template

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// roleCoreStubBudget is the binding per-stub budget in UTF-16 code units.
const roleCoreStubBudget = 10000

// stubGateModuleRoot resolves the repository root from this test file.
func stubGateModuleRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate the module root")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// checkRoleCoreStubContent is the gate predicate over one stub file's
// content: nil when the content is a legal stub (no role-core markers, at or
// under the UTF-16 budget), the violation otherwise. Test-local by design —
// the gate is the test's own instrument, exercised by the motor subtests.
func checkRoleCoreStubContent(content string) error {
	if strings.Contains(content, config.RoleCoreMarkerStart) || strings.Contains(content, config.RoleCoreMarkerEnd) {
		return errStubGate{reason: "carries a moai:role-core region marker"}
	}
	if n := utf16CodeUnits(content); n > roleCoreStubBudget {
		return errStubGate{reason: "exceeds the 10,000 UTF-16 budget", units: n}
	}
	return nil
}

// errStubGate is the gate's violation type; units is set only for size
// violations.
type errStubGate struct {
	reason string
	units  int
}

func (e errStubGate) Error() string {
	if e.units > 0 {
		return e.reason
	}
	return e.reason
}

// enumerateRoleCoreStubs globs every *-core.md under the rules tree at root
// (deployed layout) and returns their repo-relative paths.
func enumerateRoleCoreStubs(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	rulesRoot := filepath.Join(root, ".claude", "rules")
	err := filepath.WalkDir(rulesRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".md") || !strings.HasSuffix(name, "-core.md") {
			return nil //nolint:nilerr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil //nolint:nilerr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil //nolint:nilerr
	})
	if err != nil {
		t.Fatalf("stub glob over %s failed: %v", rulesRoot, err)
	}
	return out
}

// TestRoleCoreStubGate enumerates every *-core.md in both trees and asserts
// the two REQ-RIB-005 conditions per stub.
func TestRoleCoreStubGate(t *testing.T) {
	moduleRoot := stubGateModuleRoot(t)
	trees := map[string]string{
		"deployed": moduleRoot,
		"template": filepath.Join(moduleRoot, "internal", "template", "templates"),
	}
	total := 0
	for face, root := range trees {
		stubs := enumerateRoleCoreStubs(t, root)
		if len(stubs) == 0 {
			t.Fatalf("%s tree: the *-core.md glob matched nothing — an empty sweep is a failure, not a pass", face)
		}
		for _, rel := range stubs {
			total++
			t.Run(face+"/"+rel, func(t *testing.T) {
				data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
				if err != nil {
					t.Fatalf("read stub %s: %v", rel, err)
				}
				if err := checkRoleCoreStubContent(string(data)); err != nil {
					t.Errorf("stub gate violation in %s/%s: %v", face, rel, err)
				}
				t.Logf("%s/%s: %d UTF-16 (budget %d), role-core markers 0", face, rel, utf16CodeUnits(string(data)), roleCoreStubBudget)
			})
		}
	}
	t.Logf("[stub gate] swept %d stubs across 2 trees", total)
	if total == 0 {
		t.Fatal("stub gate swept 0 files across both trees — empty sweep is a failure, not a pass")
	}
}

// TestRoleCoreStubGateMotor_Marker is the self-motor observation (1/2): a
// stub fixture carrying a role-core marker makes the gate fail — the marker
// condition is live, not vacuous.
func TestRoleCoreStubGateMotor_Marker(t *testing.T) {
	mutant := "# Factory Dispatch Core\n\n<!-- moai:role-core-start -->\n[HARD] smuggled region\n<!-- moai:role-core-end -->\n"
	err := checkRoleCoreStubContent(mutant)
	if err == nil {
		t.Fatal("MOTOR: gate passed a marker-carrying stub — the marker condition is not live")
	}
	t.Logf("MOTOR observed: marker-carrying stub fails the gate: %v", err)
}

// TestRoleCoreStubGateMotor_Size is the self-motor observation (2/2): a stub
// fixture over the 10,000 UTF-16 budget makes the gate fail — the size
// condition is live, not vacuous.
func TestRoleCoreStubGateMotor_Size(t *testing.T) {
	mutant := strings.Repeat("a", roleCoreStubBudget+1) // 10,001 units, no markers
	err := checkRoleCoreStubContent(mutant)
	if err == nil {
		t.Fatal("MOTOR: gate passed an over-budget stub — the size condition is not live")
	}
	t.Logf("MOTOR observed: %d-unit stub fails the gate: %v", roleCoreStubBudget+1, err)
}

// TestRoleCoreStubGateMotor_Boundary pins the boundary the mutant probe
// names (acceptance.md §B): exactly 10,000 passes, 10,001 fails.
func TestRoleCoreStubGateMotor_Boundary(t *testing.T) {
	if err := checkRoleCoreStubContent(strings.Repeat("a", roleCoreStubBudget)); err != nil {
		t.Fatalf("exactly-at-budget stub must pass: %v", err)
	}
	if err := checkRoleCoreStubContent(strings.Repeat("a", roleCoreStubBudget+1)); err == nil {
		t.Fatal("one unit over budget must fail")
	}
}
