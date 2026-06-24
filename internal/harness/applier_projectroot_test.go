// Package harness — WithProjectRoot set/unset 단위 테스트
// (SPEC-HARNESS-EXECUTE-E2E-001 M3, AC-E2E-004).
//
// gate-active Applier가 측정 root로 무엇을 measurer에 전달하는지 capturing
// measurer로 관측한다:
//   - WithProjectRoot(root) set → measurer가 root를 수신.
//   - WithProjectRoot 미호출(unset) → measurer가 measurementRoot(snapshotDir)
//     = snapshot base 디렉터리를 수신 (기존 fallback 보존, 회귀 0).
//
// 이 파일은 ADD-ONLY다 — 기존 테스트 함수를 수정하지 않는다. 기존 stubMeasurer
// (regression_gate_test.go)는 root를 무시하므로 root 수신값을 관측할 수 없어,
// root를 기록하는 별도 capturing measurer를 여기서 정의한다.
package harness

import (
	"os"
	"path/filepath"
	"testing"
)

// rootCapturingMeasurer는 Measure가 받은 projectRoot를 기록하는 test Measurer다.
// 고정 triple을 반환하여 baseline==candidate(Δ=0)이 되게 하므로 gate는 항상 keep한다.
type rootCapturingMeasurer struct {
	gotRoots []string
	triple   MetricTriple
}

func (m *rootCapturingMeasurer) Measure(projectRoot string) (MetricTriple, error) {
	m.gotRoots = append(m.gotRoots, projectRoot)
	return m.triple, nil
}

// newGateActiveApplierForTest는 capturing measurer를 주입한 gate-active Applier를
// 구성한다 (in-package 직접 구성 — 단위 테스트 전용). 실제 goMeasurer를 쓰지 않으므로
// go test를 실행하지 않고 측정 root 전달만 빠르게 관측한다.
func newGateActiveApplierForTest(t *testing.T, m Measurer) (a *Applier, baselinePath string) {
	t.Helper()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.jsonl")
	baselinePath = filepath.Join(dir, "measurements-baseline.yaml")
	return &Applier{
		allowWrites:   enableTriggerInjectionWrites,
		manifestPath:  manifestPath,
		measurer:      m,
		baselineStore: NewBaselineStore(baselinePath),
	}, baselinePath
}

// writeApprovedProposalTarget는 비-FROZEN markdown 타겟(frontmatter 포함)을
// t.TempDir()에 작성하고 그 절대 경로를 담은 Proposal을 반환한다. createSnapshot이
// TargetPath를 직접 읽으므로 타겟 파일이 실재해야 한다.
func writeApprovedProposalTarget(t *testing.T, root, id string) Proposal {
	t.Helper()
	targetAbs := filepath.Join(root, "docs", "unit-sample.md")
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(targetAbs,
		[]byte("---\ndescription: unit original\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	return Proposal{
		ID:         id,
		TargetPath: targetAbs,
		FieldKey:   "description",
		NewValue:   "unit enriched note",
		PatternKey: "unit-pattern",
	}
}

// approveEvaluator는 항상 DecisionApproved를 반환하는 in-package stub evaluator다
// (gate-active Apply가 measurer 단계에 도달하도록).
type approveEvaluator struct{}

func (approveEvaluator) Evaluate(_ Proposal, _ []Session) (Decision, error) {
	return Decision{Kind: DecisionApproved}, nil
}

// TestWithProjectRoot_SetAndUnset는 WithProjectRoot set/unset 두 경로 모두에서
// gate가 measurer에 전달하는 측정 root를 capturing measurer로 관측한다 (AC-E2E-004).
func TestWithProjectRoot_SetAndUnset(t *testing.T) {
	t.Parallel()

	t.Run("set: measurer receives the wired project root", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		snapshotBase := filepath.Join(root, ".moai", "harness", "learning-history", "snapshots")
		prop := writeApprovedProposalTarget(t, root, "SPEC-UNIT-SET-001")

		capMeasurer := &rootCapturingMeasurer{triple: MetricTriple{TestsPassed: 10, Coverage: 80.0, LintCount: 0}}
		a, _ := newGateActiveApplierForTest(t, capMeasurer)
		a.WithProjectRoot("/x") // set

		if err := a.Apply(prop, approveEvaluator{}, snapshotBase, nil); err != nil {
			t.Fatalf("Apply error: %v", err)
		}
		if len(capMeasurer.gotRoots) == 0 {
			t.Fatal("measurer was never called — gate did not run")
		}
		for i, got := range capMeasurer.gotRoots {
			if got != "/x" {
				t.Errorf("Measure call %d received root %q, want the wired \"/x\"", i, got)
			}
		}
	})

	t.Run("unset: measurer receives measurementRoot(snapshotDir) fallback", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		snapshotBase := filepath.Join(root, ".moai", "harness", "learning-history", "snapshots")
		prop := writeApprovedProposalTarget(t, root, "SPEC-UNIT-UNSET-001")

		capMeasurer := &rootCapturingMeasurer{triple: MetricTriple{TestsPassed: 10, Coverage: 80.0, LintCount: 0}}
		a, _ := newGateActiveApplierForTest(t, capMeasurer)
		// WithProjectRoot 미호출 — fallback 경로.

		if err := a.Apply(prop, approveEvaluator{}, snapshotBase, nil); err != nil {
			t.Fatalf("Apply error: %v", err)
		}
		if len(capMeasurer.gotRoots) == 0 {
			t.Fatal("measurer was never called — gate did not run")
		}
		// fallback = measurementRoot(snapshotDir) = filepath.Dir(<snapshotBase>/<ISO-DATE>)
		// = snapshotBase. 수신 root가 snapshotBase와 같아야 한다 (기존 동작 보존).
		for i, got := range capMeasurer.gotRoots {
			if got != snapshotBase {
				t.Errorf("Measure call %d received root %q, want measurementRoot fallback %q (snapshotBase)",
					i, got, snapshotBase)
			}
			if got == "/x" {
				t.Errorf("Measure call %d received \"/x\" without WithProjectRoot — fallback not used", i)
			}
		}
	})
}
