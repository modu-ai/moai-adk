package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// clustersTestProject 는 .moai/harness/ 디렉터리를 갖춘 임시 프로젝트 루트를 만든다.
func clustersTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	harnessDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("harness 디렉터리 생성 실패: %v", err)
	}
	return dir
}

// writeClusterUsageLog 은 apply_outcome JSONL 줄들을 usage-log.jsonl 에 기록한다.
func writeClusterUsageLog(t *testing.T, dir string, lines []string) {
	t.Helper()
	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if err := os.WriteFile(logPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("usage-log.jsonl 기록 실패: %v", err)
	}
}

const (
	rolledBackCovLine = `{"timestamp":"2026-06-01T00:00:00Z","event_type":"apply_outcome","subject":"apply:PROPOSAL-A","outcome_verdict":"rolled-back","outcome_decision":"regression-blocked","outcome_proposal_id":"PROPOSAL-A","outcome_regressed":["coverage"],"schema_version":"v2.1"}`
	rolledBackCov2Line = `{"timestamp":"2026-06-02T00:00:00Z","event_type":"apply_outcome","subject":"apply:PROPOSAL-B","outcome_verdict":"rolled-back","outcome_decision":"regression-blocked","outcome_proposal_id":"PROPOSAL-B","outcome_regressed":["coverage"],"schema_version":"v2.1"}`
	keptLine           = `{"timestamp":"2026-06-03T00:00:00Z","event_type":"apply_outcome","subject":"apply:PROPOSAL-K","outcome_verdict":"kept","outcome_decision":"approved","outcome_proposal_id":"PROPOSAL-K","schema_version":"v2.1"}`
)

// ── AC-OBL-004: CLI read 표면 (REQ-OBL-010/011/012) ──

// TestHarnessClustersCmdHelp: `moai harness clusters --help` 가 exit 0 (live-tree wiring 증명).
// newHarnessRouterCmd 트리에서 clusters 가 발견되어야 한다(deprecation-marker 트리 아님).
func TestHarnessClustersCmdHelp(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"clusters", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("`harness clusters --help` 는 exit 0 이어야 함 (live-tree wiring): %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "clusters") {
		t.Errorf("--help 출력에 clusters 가 없음:\n%s", output)
	}
}

// TestHarnessClustersCmd: 텍스트 출력이 클러스터를 나열하고, --json 이 유효한 JSON 을
// stdout 으로 방출한다.
func TestHarnessClustersCmd(t *testing.T) {
	t.Parallel()

	dir := clustersTestProject(t)
	writeClusterUsageLog(t, dir, []string{rolledBackCovLine, rolledBackCov2Line, keptLine})

	// 텍스트 출력
	var textBuf bytes.Buffer
	textCmd := newHarnessRouterCmd()
	textCmd.SetOut(&textBuf)
	textCmd.SetArgs([]string{"clusters", "--project-root", dir})
	if err := textCmd.Execute(); err != nil {
		t.Fatalf("`harness clusters` (텍스트) 실패: %v", err)
	}
	text := textBuf.String()
	if !strings.Contains(text, "Failure-Signature Clusters") {
		t.Errorf("텍스트 출력에 클러스터 헤더가 없음:\n%s", text)
	}
	if !strings.Contains(text, "coverage") {
		t.Errorf("텍스트 출력에 coverage 차원이 없음:\n%s", text)
	}

	// --json 출력
	var jsonBuf bytes.Buffer
	jsonCmd := newHarnessRouterCmd()
	jsonCmd.SetOut(&jsonBuf)
	jsonCmd.SetArgs([]string{"clusters", "--project-root", dir, "--json"})
	if err := jsonCmd.Execute(); err != nil {
		t.Fatalf("`harness clusters --json` 실패: %v", err)
	}

	// stdout 이 유효한 JSON 이어야 한다.
	var report struct {
		SchemaVersion string `json:"schema_version"`
		ClusterCount  int    `json:"cluster_count"`
		Clusters      []struct {
			Signature string `json:"signature"`
			Count     int    `json:"count"`
		} `json:"clusters"`
	}
	if err := json.Unmarshal(jsonBuf.Bytes(), &report); err != nil {
		t.Fatalf("--json 출력이 유효한 JSON 이 아님: %v\noutput:\n%s", err, jsonBuf.String())
	}
	// rolled-back 2개(coverage)는 한 클러스터(count==2), kept 는 제외.
	if report.ClusterCount != 1 {
		t.Errorf("cluster_count = %d, want 1 (kept 제외)", report.ClusterCount)
	}
	if len(report.Clusters) != 1 || report.Clusters[0].Count != 2 {
		t.Errorf("클러스터 = %+v, want 1 cluster count=2", report.Clusters)
	}
}

// TestHarnessClustersEmpty: 빈/부재 로그 → 빈 결과, exit 0 (에러 아님) (REQ-OBL-012).
func TestHarnessClustersEmpty(t *testing.T) {
	t.Parallel()

	dir := clustersTestProject(t)
	// usage-log.jsonl 을 의도적으로 만들지 않음(부재 케이스).

	// 텍스트
	var textBuf bytes.Buffer
	textCmd := newHarnessRouterCmd()
	textCmd.SetOut(&textBuf)
	textCmd.SetArgs([]string{"clusters", "--project-root", dir})
	if err := textCmd.Execute(); err != nil {
		t.Fatalf("부재 로그는 에러가 아니어야 함 (exit 0): %v", err)
	}
	if !strings.Contains(textBuf.String(), "No failure-signature clusters") {
		t.Errorf("부재 로그 텍스트 출력에 빈 결과 메시지가 없음:\n%s", textBuf.String())
	}

	// --json (빈 결과도 유효한 JSON)
	var jsonBuf bytes.Buffer
	jsonCmd := newHarnessRouterCmd()
	jsonCmd.SetOut(&jsonBuf)
	jsonCmd.SetArgs([]string{"clusters", "--project-root", dir, "--json"})
	if err := jsonCmd.Execute(); err != nil {
		t.Fatalf("부재 로그 --json 은 에러가 아니어야 함: %v", err)
	}
	if !strings.Contains(jsonBuf.String(), `"cluster_count": 0`) {
		t.Errorf("부재 로그 --json 에 cluster_count: 0 가 없음:\n%s", jsonBuf.String())
	}
}

// TestHarnessClustersEmptyLogFile: usage-log.jsonl 이 존재하지만 비어 있는 경우도
// 빈 결과 + exit 0 (EC-1).
func TestHarnessClustersEmptyLogFile(t *testing.T) {
	t.Parallel()

	dir := clustersTestProject(t)
	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if err := os.WriteFile(logPath, []byte(""), 0o600); err != nil {
		t.Fatalf("빈 로그 생성 실패: %v", err)
	}

	var buf bytes.Buffer
	cmd := newHarnessRouterCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"clusters", "--project-root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("빈 로그는 에러가 아니어야 함: %v", err)
	}
	if !strings.Contains(buf.String(), "No failure-signature clusters") {
		t.Errorf("빈 로그 출력에 빈 결과 메시지가 없음:\n%s", buf.String())
	}
}

// TestHarnessClustersFactoryNotNil: 팩토리가 nil 이 아니고 Use 가 "clusters" 다.
func TestHarnessClustersFactoryNotNil(t *testing.T) {
	t.Parallel()
	cmd := newHarnessClustersCmd()
	if cmd == nil {
		t.Fatal("newHarnessClustersCmd() 가 nil 반환")
	}
	if cmd.Use != "clusters" {
		t.Errorf("Use = %q, want clusters", cmd.Use)
	}
	if cmd.Flags().Lookup("json") == nil {
		t.Error("--json 플래그가 등록되지 않음")
	}
}

// TestHarnessClustersRegisteredInLiveTree: clusters 가 LIVE 트리(newHarnessRouterCmd)에
// 등록되어 있는지 직접 검증한다(AC-OBL-004 — deprecation-marker 트리 아님).
func TestHarnessClustersRegisteredInLiveTree(t *testing.T) {
	t.Parallel()
	router := newHarnessRouterCmd()
	found := false
	for _, sub := range router.Commands() {
		if sub.Name() == "clusters" {
			found = true
			break
		}
	}
	if !found {
		t.Error("clusters 가 newHarnessRouterCmd (LIVE 트리)에 등록되지 않음")
	}
}
