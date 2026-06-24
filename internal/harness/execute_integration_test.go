// Package harness — Apply-outcome 텔레메트리 통합(T1 white-box) 테스트.
// SPEC-HARNESS-APPLY-EXECUTE-001 REQ-AEX-010 (AC-AEX-010).
//
// design.md §F.1 [D4] 2-tier 테스트 seam 결정에 따른 T1(white-box, package harness)
// 테스트. same-package 접근으로 gate-active Applier를 stub measurer(고정 triple) +
// NewBaselineStore(t.TempDir())로 직접 구성하여, 정상 apply 후 usage-log.jsonl에
// apply_outcome 라인이 1건 기록됨을 실증한다. 실제 go test/go vet 재귀 실행 없음
// (stub measurer가 toolchain을 대체).
package harness

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// approvingEvaluator는 항상 DecisionApproved를 반환하는 same-package stub이다.
// production safety.Pipeline(AutoApply: true)의 L1~L5 통과 결과를 white-box 통합
// 테스트에서 직접 모사하여, gate-active Apply가 telemetry 기록 경로까지 도달하게 한다.
type approvingEvaluator struct{}

func (approvingEvaluator) Evaluate(_ Proposal, _ []Session) (Decision, error) {
	return Decision{Kind: DecisionApproved}, nil
}

// TestExecute_FirstApply_WritesApplyOutcomeTelemetry — [AC-AEX-010]
// first run(baseline 파일 부재) + 통과 proposal로 gate-active Apply를 수행하면,
// usage-log.jsonl에 apply_outcome event 라인이 1건 추가되고 verdict="kept"임을
// 검증한다 (첫 apply-outcome telemetry — Phase 5 분석 입력 substrate).
func TestExecute_FirstApply_WritesApplyOutcomeTelemetry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// 비-FROZEN 타겟 SKILL.md fixture (description fieldKey). applyFileModification이
	// EnrichDescription으로 실제 파일을 수정하므로 frontmatter 파일이 필요하다.
	targetPath := filepath.Join(root, "docs", "telemetry-sample.md")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("---\ndescription: original\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("write target fixture: %v", err)
	}

	usageLogPath := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
	manifestPath := filepath.Join(root, ".moai", "harness", "learning-history", "manifest.jsonl")
	snapshotBase := filepath.Join(root, ".moai", "harness", "learning-history", "snapshots")
	baselineDir := filepath.Join(root, ".moai", "harness")
	if err := os.MkdirAll(baselineDir, 0o755); err != nil {
		t.Fatalf("mkdir baseline dir: %v", err)
	}
	baselinePath := filepath.Join(baselineDir, "measurements-baseline.yaml")

	// gate-active Applier 직접 구성 (same-package): stub measurer로 고정 triple 반환
	// → 실제 toolchain 미실행. baselineStore는 t.TempDir() 하위 파일(부재=first run).
	applier := &Applier{
		manifestPath:    manifestPath,
		measurer:        stubMeasurer{triple: MetricTriple{TestsPassed: 10, Coverage: 80.0, LintCount: 0}},
		baselineStore:   NewBaselineStore(baselinePath),
		outcomeObserver: NewObserver(usageLogPath),
	}
	if !applier.gateActive() {
		t.Fatal("test setup error: applier must be gate-active (measurer + baselineStore wired)")
	}

	proposal := Proposal{
		ID:         "SPEC-TELEMETRY-001",
		TargetPath: targetPath,
		FieldKey:   "description",
		NewValue:   "telemetry enriched note",
		PatternKey: "test-pattern",
	}

	if err := applier.Apply(proposal, approvingEvaluator{}, snapshotBase, nil); err != nil {
		t.Fatalf("gate-active Apply (first run) must succeed; got: %v", err)
	}

	// usage-log.jsonl에 apply_outcome 라인 1건 + verdict="kept" 검증.
	applyOutcomes := readApplyOutcomeEvents(t, usageLogPath)
	if len(applyOutcomes) != 1 {
		t.Fatalf("expected exactly 1 apply_outcome telemetry line, got %d", len(applyOutcomes))
	}
	evt := applyOutcomes[0]
	if evt.EventType != EventTypeApplyOutcome {
		t.Errorf("event_type = %q, want %q", evt.EventType, EventTypeApplyOutcome)
	}
	if evt.OutcomeVerdict != "kept" {
		t.Errorf("outcome_verdict = %q, want \"kept\" (first-run baseline adopted)", evt.OutcomeVerdict)
	}
	if evt.OutcomeProposalID != proposal.ID {
		t.Errorf("outcome_proposal_id = %q, want %q", evt.OutcomeProposalID, proposal.ID)
	}
}

// readApplyOutcomeEvents reads usage-log.jsonl and returns only apply_outcome events.
func readApplyOutcomeEvents(t *testing.T, logPath string) []Event {
	t.Helper()
	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("open usage-log.jsonl: %v", err)
	}
	defer func() { _ = f.Close() }()

	var outcomes []Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var evt Event
		if err := json.Unmarshal(line, &evt); err != nil {
			t.Fatalf("unmarshal usage-log line: %v", err)
		}
		if evt.EventType == EventTypeApplyOutcome {
			outcomes = append(outcomes, evt)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan usage-log: %v", err)
	}
	return outcomes
}
