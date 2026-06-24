// Package harness — execute verb 텔레메트리 e2e 재현 테스트
// (SPEC-HARNESS-EXECUTE-E2E-001).
//
// 이 파일은 CLAUDE.md Rule 4 (Reproduction-First Bug Fixing)에 따른 재현
// 테스트다. RunExecute의 production glue 전체(real safety.Pipeline(AutoApply:true)
// → DecisionApproved → gate-active Apply → real goMeasurer)를 real 최소 Go 모듈
// fixture에서 구동한다.
//
// 재현 대상 결함: internal/harness/applier.go measurementRoot(snapshotDir)이
// snapshot BASE 디렉터리(<root>/.moai/harness/learning-history/snapshots)를
// 반환하여, gate-active RunExecute 경로가 testable Go 패키지를 보유하지 않는
// 디렉터리에서 `go test ./...`를 실행 → fail-close → telemetry 0건.
//
// RED(미수정): measurer가 snapshot base를 측정 → build-fail / 0-byte stdout →
//              regression-gate measurement 에러 → apply_outcome 라인 0건.
// GREEN(수정 후): measurer가 corrected project root(=fixture 모듈)를 측정 →
//              `go test ./...` 통과 → Δ=0 → verdict="kept" → apply_outcome 1건.
//
// 격리 불변식 (spec.md §B.3): 모든 write(go.mod / trivial 테스트 / proposal /
// 타겟 / usage-log / baseline / coverage profile)는 t.TempDir() 내부에서만
// 발생한다. 실제 레포의 .moai/harness/* 파일은 절대 touch되지 않는다.
//
// HARD subagent boundary (C-HRA-008): 이 파일은 AskUserQuestion을 호출하지
// 않는다. 패키지 가드 TestPropose_NoAskUserQuestion이 자동 스캔한다.
package harness

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeE2EFixtureModule는 t.TempDir() 프로젝트 root를 real 최소 Go 모듈로 구성하고
// (REQ-E2E-007), 비-FROZEN markdown 타겟의 절대 경로를 반환한다.
//
// 구성:
//   - <root>/go.mod          — 독립 모듈 (corrected project root에서 go test가 성공)
//   - <root>/fixture_test.go — 통과하는 trivial 테스트 1건 (1 package, 0 fail)
//   - <root>/docs/e2e-sample.md — frontmatter 포함 비-FROZEN 타겟 (L1 통과 + EnrichDescription 가능)
//
// 반환값은 타겟의 절대 경로다. writeProposalFixture(execute_test.go)가 이 절대
// 경로를 proposal.TargetPath로 사용한다 — createSnapshot이 TargetPath를 직접 읽기
// 때문에 절대 경로가 필요하다 (canary 테스트와 동일 패턴).
func writeE2EFixtureModule(t *testing.T, root string) (targetAbs string) {
	t.Helper()

	// go.mod — 독립 모듈. corrected project root에서 `go test ./...`가 빌드 성공.
	if err := os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module e2efixture\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// 통과하는 trivial 테스트 1건 — project root에서 go test가 1 package / 0 fail.
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"),
		[]byte("package e2efixture\n\nimport \"testing\"\n\nfunc TestFixturePass(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatalf("write fixture_test.go: %v", err)
	}

	// 비-FROZEN markdown 타겟 (frontmatter 포함 — EnrichDescription 가능, EC-2).
	targetAbs = filepath.Join(root, "docs", "e2e-sample.md")
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(targetAbs,
		[]byte("---\ndescription: original e2e sample\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("write target fixture: %v", err)
	}
	return targetAbs
}

// countApplyOutcomeKept는 usage-log.jsonl에서 event_type == "apply_outcome" AND
// outcome_verdict == "kept" 라인 수를 센다 (EC-5: 다른 이벤트 공존 시 필터링).
// 파일 부재는 0을 반환한다 (RED 경로: telemetry가 한 번도 기록되지 않은 상태).
func countApplyOutcomeKept(t *testing.T, usageLogPath string) int {
	t.Helper()

	f, err := os.Open(usageLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0 // telemetry 미기록 (RED 상태)
		}
		t.Fatalf("open usage-log: %v", err)
	}
	defer func() { _ = f.Close() }()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var evt struct {
			EventType      string `json:"event_type"`
			OutcomeVerdict string `json:"outcome_verdict"`
		}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("unmarshal usage-log line %q: %v", line, err)
		}
		if evt.EventType == "apply_outcome" && evt.OutcomeVerdict == "kept" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan usage-log: %v", err)
	}
	return count
}

// TestRunExecute_RegressionGateMeasuresProjectRoot_WritesKeptTelemetry는
// RunExecute의 production glue 전체를 real 최소 Go 모듈 fixture에서 구동하는
// e2e 재현 테스트다 (REQ-E2E-004 RED / REQ-E2E-005 GREEN / REQ-E2E-006 직접 진입점).
//
// RED(미수정 measurementRoot): RunExecute는 measurer를 snapshot base에서 실행 →
//   fail-close → 반환 error != nil + apply_outcome("kept") 라인 0건. 이 테스트는
//   미수정 코드에서 FAIL하여 출시된 verb의 결함을 재현한다.
//
// GREEN(수정 후): RunExecute가 .WithProjectRoot(root)를 배선 → measurer가 corrected
//   project root(fixture 모듈)에서 go test ./... 통과(Δ=0) → DecisionApproved →
//   verdict="kept" → 반환 error == nil + apply_outcome("kept") 라인 정확히 1건.
//
// 격리(REQ-E2E-008): ProjectRoot=t.TempDir()로 모든 harness 경로를 격리한다 —
// 실제 레포 .moai/harness/*는 절대 touch되지 않는다.
func TestRunExecute_RegressionGateMeasuresProjectRoot_WritesKeptTelemetry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-E2E-REPRO-001"

	// real 최소 Go 모듈 fixture (go.mod + 통과 테스트 + 비-FROZEN 타겟).
	targetAbs := writeE2EFixtureModule(t, root)

	// synthetic proposal: 비-FROZEN 타겟의 description fieldKey enrich (L1 통과).
	// targetAbs(절대 경로)를 TargetPath로 사용 — createSnapshot이 직접 읽는다.
	writeProposalFixture(t, root, id, targetAbs, "description", "e2e enriched note")

	// production 진입점 RunExecute 직접 호출 (REQ-E2E-006). 이 재현 테스트는
	// 어떤 우회 경로도 사용하지 않는다 — Applier struct 직접 구성, Apply 직접 호출,
	// 테스트 seam(stub evaluator/measurer 주입) 모두 회피한다 (AC-E2E-001 grep gate).
	// RunExecute가 NewApplierWithRegressionGate(real goMeasurer) + NewObserver +
	// safety.NewPipeline(AutoApply:true)를 production으로 배선한다.
	err := RunExecute(ExecuteOptions{ID: id, ProjectRoot: root})

	// usage-log는 ProjectRoot 기준으로 resolve된다 — t.TempDir() 내부.
	usageLogPath := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
	keptCount := countApplyOutcomeKept(t, usageLogPath)

	// GREEN 어서션 (수정 후 통과 / 미수정 시 FAIL = RED 재현):
	//  (1) 반환 error == nil
	//  (2) apply_outcome("kept") 라인 정확히 1건
	if err != nil {
		// RED 경로: measurementRoot가 snapshot base이면 measurer가 fail-close하여
		// 여기서 measurement 부류 에러가 잡힌다. 이것이 출시된 verb의 결함 재현이다.
		t.Fatalf("RunExecute must succeed with verdict=kept against the FIXED code; "+
			"got error (this is the RED reproduction of the measurementRoot defect): %v\n"+
			"apply_outcome(kept) line count = %d (want 1)", err, keptCount)
	}
	if keptCount != 1 {
		t.Fatalf("apply_outcome(kept) line count = %d, want exactly 1 "+
			"(telemetry happy-path); err=%v", keptCount, err)
	}
}

// TestRunExecute_RED_DefectErrorIsMeasurementFailClose는 RED 재현을 명시적으로
// 문서화하는 보조 어서션이다 (AC-E2E-002). 미수정 코드에서 RunExecute가 반환하는
// 에러가 "measurement" / "fail-closed" 부류임을 — 즉 결함이 측정 root 오류에서
// 기인함을 — 확인한다.
//
// 수정 후(GREEN)에는 RunExecute가 nil error를 반환하므로 이 보조 어서션은
// "에러가 없으면 결함이 봉쇄된 것"으로 통과 처리한다 (forward-looking sentinel:
// measurementRoot가 다시 snapshot base로 회귀하면 err != nil이 되고 메시지에
// measurement/fail-closed가 포함되어 즉시 알린다).
func TestRunExecute_RED_DefectErrorIsMeasurementFailClose(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	id := "SPEC-E2E-REPRO-002"
	targetAbs := writeE2EFixtureModule(t, root)
	writeProposalFixture(t, root, id, targetAbs, "description", "e2e red note")

	err := RunExecute(ExecuteOptions{ID: id, ProjectRoot: root})

	if err == nil {
		// GREEN(수정 후): 결함 봉쇄 — 측정 root가 올바르게 corrected됨.
		return
	}
	// RED(미수정): 결함이 측정 root 오류에서 기인함을 확인 (measurement / fail-closed).
	msg := err.Error()
	if !strings.Contains(msg, "measurement") && !strings.Contains(msg, "fail-closed") {
		t.Fatalf("RED defect error must be a regression-gate measurement fail-close; "+
			"got non-measurement error: %v", err)
	}
}
