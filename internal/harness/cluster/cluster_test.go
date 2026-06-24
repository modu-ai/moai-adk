package cluster

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// ── AC-OBL-001: 수집 + 결정론적 그룹화 (REQ-OBL-001/005/006) ──

// TestClusterEvents: 동일 시그니처 키(정렬된 outcome_regressed + verdict + decision)를
// 공유하는 rolled-back 이벤트가 하나의 FailureCluster(count==2)로 묶이고, 다른 차원의
// 이벤트는 별도 클러스터로 분리되는지 검증한다. 시그니처 파생은 pattern_key 를 절대
// 읽지 않는다(픽스처 이벤트에 pattern_key 가 없음).
func TestClusterEvents(t *testing.T) {
	events := []harness.Event{
		// 두 개의 동일 시그니처(coverage / rolled-back / regression-blocked)
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-B", []string{"coverage"}, ts(2)),
		// 별도 차원(lint) → 별도 클러스터
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-C", []string{"lint"}, ts(3)),
	}

	clusters := ClusterEvents(events)

	if len(clusters) != 2 {
		t.Fatalf("클러스터 수 = %d, want 2", len(clusters))
	}

	// 시그니처로 정렬되므로 "coverage|..." < "lint|..." (사전순).
	cov := clusters[0]
	lint := clusters[1]

	if cov.Count != 2 {
		t.Errorf("coverage 클러스터 count = %d, want 2", cov.Count)
	}
	if !reflect.DeepEqual(cov.RepresentativeDimensions, []string{"coverage"}) {
		t.Errorf("coverage 대표 차원 = %v, want [coverage]", cov.RepresentativeDimensions)
	}
	if lint.Count != 1 {
		t.Errorf("lint 클러스터 count = %d, want 1", lint.Count)
	}
	if !reflect.DeepEqual(lint.RepresentativeDimensions, []string{"lint"}) {
		t.Errorf("lint 대표 차원 = %v, want [lint]", lint.RepresentativeDimensions)
	}

	// FirstSeen/LastSeen 가 멤버 타임스탬프 범위와 일치하는지(coverage: ts(1)~ts(2)).
	if !cov.FirstSeen.Equal(ts(1)) {
		t.Errorf("coverage FirstSeen = %v, want %v", cov.FirstSeen, ts(1))
	}
	if !cov.LastSeen.Equal(ts(2)) {
		t.Errorf("coverage LastSeen = %v, want %v", cov.LastSeen, ts(2))
	}

	// 멤버 참조에 ProposalID 가 보존되는지.
	if len(cov.Members) != 2 {
		t.Fatalf("coverage 멤버 수 = %d, want 2", len(cov.Members))
	}
}

// TestSignatureKeyOrderInvariant: outcome_regressed 차원의 입력 순서가 달라도
// 동일 시그니처 키를 만든다(정렬된 차원 집합 — REQ-OBL-005). 순서가 다른 두
// rolled-back 이벤트가 하나의 클러스터로 묶여야 한다.
func TestSignatureKeyOrderInvariant(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage", "lint"}, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-B", []string{"lint", "coverage"}, ts(2)),
	}
	clusters := ClusterEvents(events)
	if len(clusters) != 1 {
		t.Fatalf("차원 순서가 달라도 한 클러스터여야 함, got %d", len(clusters))
	}
	if clusters[0].Count != 2 {
		t.Errorf("count = %d, want 2", clusters[0].Count)
	}
	// 대표 차원은 정렬되어야 한다.
	if !reflect.DeepEqual(clusters[0].RepresentativeDimensions, []string{"coverage", "lint"}) {
		t.Errorf("대표 차원 = %v, want [coverage lint] (정렬)", clusters[0].RepresentativeDimensions)
	}
}

// TestSignatureKeyNoPatternKey: 시그니처 키 파생이 pattern_key 에 의존하지 않음을
// 검증한다(REQ-OBL-005). apply_outcome 이벤트에는 pattern_key 필드 자체가 없으며,
// OutcomeProposalID 가 서로 달라도(동일 시그니처라면) 한 클러스터로 묶여야 한다 —
// 즉 ProposalID 는 시그니처에 들어가지 않는다.
func TestSignatureKeyNoPatternKey(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-DIFFERENT-1", []string{"coverage"}, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-DIFFERENT-2", []string{"coverage"}, ts(2)),
	}
	clusters := ClusterEvents(events)
	if len(clusters) != 1 {
		t.Fatalf("ProposalID 가 달라도 동일 시그니처는 한 클러스터여야 함(ProposalID 는 시그니처 미포함), got %d", len(clusters))
	}
	// 시그니처 키 문자열에 ProposalID 가 들어가면 안 된다(직접 검증).
	sig := clusters[0].Signature
	if containsAny(sig, "PROPOSAL-DIFFERENT-1", "PROPOSAL-DIFFERENT-2") {
		t.Errorf("시그니처 키에 ProposalID 가 누수됨: %q", sig)
	}
	// 시그니처 키는 (차원|verdict|decision) 구조여야 한다.
	wantSig := "coverage" + sigFieldDelimiter + verdictRolledBack + sigFieldDelimiter + "regression-blocked"
	if sig != wantSig {
		t.Errorf("시그니처 키 = %q, want %q", sig, wantSig)
	}
}

// ── AC-OBL-002: 결정론 (바이트 동일 반복 출력) (REQ-OBL-007) ──

// TestClusterDeterministic: 동일 입력에 대해 ClusterEvents 를 두 번 호출하면
// 바이트 동일 출력(직렬화 후)을 만든다. 맵-순회-순서/난수/time.Now() 누수 없음.
func TestClusterDeterministic(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-Z", []string{"tests"}, ts(5)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-M", []string{"lint", "coverage"}, ts(3)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-B", []string{"coverage"}, ts(2)),
	}

	first := ClusterEvents(events)
	second := ClusterEvents(events)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("반복 실행 결과가 DeepEqual 이 아님(결정론 위반)")
	}

	// 직렬화 바이트 동일성도 검증(AC-OBL-002 의 byte-equal 표현).
	b1, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("first 직렬화 실패: %v", err)
	}
	b2, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("second 직렬화 실패: %v", err)
	}
	if string(b1) != string(b2) {
		t.Errorf("반복 실행 직렬화 바이트가 동일하지 않음(결정론 위반)")
	}

	// 클러스터가 시그니처 사전순으로 정렬되어 있는지 검증(안정 정렬).
	for i := 1; i < len(first); i++ {
		if first[i-1].Signature >= first[i].Signature {
			t.Errorf("클러스터가 시그니처 사전순 정렬이 아님: %q >= %q", first[i-1].Signature, first[i].Signature)
		}
	}
}

// TestClusterMembersDeterministicOrder: 한 클러스터 내 멤버가 (Timestamp, ProposalID)
// 안정 순서로 정렬되는지 검증한다(입력 순서 무관, REQ-OBL-007).
func TestClusterMembersDeterministicOrder(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-C", []string{"coverage"}, ts(3)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-B", []string{"coverage"}, ts(2)),
	}
	clusters := ClusterEvents(events)
	if len(clusters) != 1 {
		t.Fatalf("한 클러스터여야 함, got %d", len(clusters))
	}
	members := clusters[0].Members
	for i := 1; i < len(members); i++ {
		if members[i-1].Timestamp.After(members[i].Timestamp) {
			t.Errorf("멤버가 시간 순 정렬이 아님: %v after %v", members[i-1].Timestamp, members[i].Timestamp)
		}
	}
	// 첫 멤버는 ts(1)/PROPOSAL-A 여야 한다.
	if members[0].ProposalID != "PROPOSAL-A" {
		t.Errorf("첫 멤버 ProposalID = %q, want PROPOSAL-A", members[0].ProposalID)
	}
}

// ── AC-OBL-003: kept 결과 제외 (REQ-OBL-008) ──

// TestClusterExcludesKept: kept verdict 이벤트는 어떤 실패 클러스터에도 포함되지
// 않는다. 모든 클러스터의 멤버 합계 == rolled-back 이벤트 수.
func TestClusterExcludesKept(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictKept, "approved", "PROPOSAL-K1", nil, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-R1", []string{"coverage"}, ts(2)),
		mkEvent(verdictKept, "approved", "PROPOSAL-K2", nil, ts(3)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-R2", []string{"lint"}, ts(4)),
		mkEvent(verdictKept, "approved", "PROPOSAL-K3", nil, ts(5)),
	}
	clusters := ClusterEvents(events)

	total := 0
	for _, c := range clusters {
		total += c.Count
		if c.Verdict != verdictRolledBack {
			t.Errorf("클러스터 verdict = %q, want %q (kept 가 클러스터를 구성하면 안 됨)", c.Verdict, verdictRolledBack)
		}
	}
	// rolled-back 2개만 클러스터링되어야 한다(kept 3개는 0 기여).
	if total != 2 {
		t.Errorf("총 멤버 수 = %d, want 2 (rolled-back 만)", total)
	}
}

// TestClusterAllKeptZeroClusters: kept 만 있는 로그 → 0개 실패 클러스터 (EC-4).
func TestClusterAllKeptZeroClusters(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictKept, "approved", "PROPOSAL-K1", nil, ts(1)),
		mkEvent(verdictKept, "approved", "PROPOSAL-K2", nil, ts(2)),
	}
	clusters := ClusterEvents(events)
	if len(clusters) != 0 {
		t.Errorf("all-kept 로그는 0개 클러스터여야 함(EC-4), got %d", len(clusters))
	}
}

// TestClusterSingleMember: 고유 시그니처의 단일 멤버도 유효한 클러스터(count==1)를
// 형성한다 — 최소 크기 임계값 없음(EC-5).
func TestClusterSingleMember(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-SOLO", []string{"coverage"}, ts(1)),
	}
	clusters := ClusterEvents(events)
	if len(clusters) != 1 {
		t.Fatalf("단일 멤버도 클러스터를 형성해야 함(EC-5), got %d", len(clusters))
	}
	if clusters[0].Count != 1 {
		t.Errorf("count = %d, want 1", clusters[0].Count)
	}
}

// TestClusterEmptyInput: 빈 입력 → 빈(nil 아닌 0-length) 클러스터 슬라이스.
func TestClusterEmptyInput(t *testing.T) {
	clusters := ClusterEvents(nil)
	if len(clusters) != 0 {
		t.Errorf("빈 입력은 0개 클러스터여야 함, got %d", len(clusters))
	}
}

// TestClusterRolledBackEmptyDimensions: rolled-back 이지만 regressed 목록이 비어
// 있는 경우(omitempty 로 빠진 경우), emptyDimToken 으로 안정 클러스터링된다.
func TestClusterRolledBackEmptyDimensions(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", nil, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-B", []string{}, ts(2)),
	}
	clusters := ClusterEvents(events)
	if len(clusters) != 1 {
		t.Fatalf("빈 차원 두 이벤트는 한 클러스터(emptyDimToken)여야 함, got %d", len(clusters))
	}
	wantSig := emptyDimToken + sigFieldDelimiter + verdictRolledBack + sigFieldDelimiter + "regression-blocked"
	if clusters[0].Signature != wantSig {
		t.Errorf("빈 차원 시그니처 = %q, want %q", clusters[0].Signature, wantSig)
	}
}

// TestClusterConvenienceEntryPoint: Cluster(logPath) 편의 진입점이 LoadEvents →
// ClusterEvents 결합을 올바르게 수행한다.
func TestClusterConvenienceEntryPoint(t *testing.T) {
	path := writeLog(t, []string{
		applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
		applyOutcomeLine(t, verdictKept, "approved", "PROPOSAL-B", nil, ts(2)),
	})
	clusters, err := Cluster(path)
	if err != nil {
		t.Fatalf("Cluster() err = %v", err)
	}
	if len(clusters) != 1 {
		t.Errorf("Cluster() = %d 클러스터, want 1 (kept 제외)", len(clusters))
	}
}

// TestClusterAbsentLogZeroClusters: 부재 로그 → 0 클러스터 + 성공 (REQ-OBL-004).
func TestClusterAbsentLogZeroClusters(t *testing.T) {
	clusters, err := Cluster("/nonexistent/path/usage-log.jsonl")
	if err != nil {
		t.Fatalf("부재 로그는 에러가 아니어야 함: %v", err)
	}
	if len(clusters) != 0 {
		t.Errorf("부재 로그는 0 클러스터여야 함, got %d", len(clusters))
	}
}

// TestDefaultLogPath: projectRoot 기준 기본 usage-log 경로를 올바르게 결합한다.
func TestDefaultLogPath(t *testing.T) {
	got := DefaultLogPath("/proj")
	want := "/proj/.moai/harness/usage-log.jsonl"
	if got != want {
		t.Errorf("DefaultLogPath() = %q, want %q", got, want)
	}
}

// ── 테스트 헬퍼 ──────────────────────────────────────────────

// mkEvent 은 apply_outcome 이벤트 값을 만든다(in-memory ClusterEvents 테스트용).
func mkEvent(verdict, decision, proposalID string, regressed []string, when time.Time) harness.Event {
	return harness.Event{
		Timestamp:         when,
		EventType:         harness.EventTypeApplyOutcome,
		Subject:           "apply:" + proposalID,
		SchemaVersion:     harness.LogSchemaVersion,
		OutcomeVerdict:    verdict,
		OutcomeDecision:   decision,
		OutcomeProposalID: proposalID,
		OutcomeRegressed:  regressed,
	}
}

// containsAny 는 s 가 substrs 중 하나라도 포함하는지 검사한다.
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if sub != "" && strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
