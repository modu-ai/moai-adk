package dashboard

import (
	"math"
	"testing"
	"time"
)

// ─── 테스트 헬퍼 ────────────────────────────────────────────────────────────

// newMetric은 반복 작성을 줄이기 위한 TaskMetric 팩토리 함수다.
func newMetric(specID, agent, phase string, in, out, durMS int64, success bool) TaskMetric {
	return TaskMetric{
		Timestamp:    time.Now(),
		SessionID:    "ses-1",
		SpecID:       specID,
		AgentName:    agent,
		Phase:        phase,
		Action:       "test",
		InputTokens:  in,
		OutputTokens: out,
		DurationMS:   durMS,
		Success:      success,
	}
}

// almostEqual은 부동소수점 비교에서 허용 오차를 적용한다.
func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

// ─── ComputeSPECCosts 테스트 ─────────────────────────────────────────────────

func TestComputeSPECCosts_EmptyInput(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	result := agg.ComputeSPECCosts(nil)

	// 빈 입력이면 빈 슬라이스를 반환해야 한다.
	if len(result) != 0 {
		t.Errorf("빈 입력: 길이 0 기대, 실제 %d", len(result))
	}
}

func TestComputeSPECCosts_SpecIDEmptySkipped(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "agent-a", "run", 100, 50, 200, true), // specID 없음 → 무시
	}
	result := agg.ComputeSPECCosts(metrics)

	// specID 빈 메트릭은 집계 대상에서 제외된다.
	if len(result) != 0 {
		t.Errorf("specID 빈 항목 무시 기대, 실제 %d개", len(result))
	}
}

func TestComputeSPECCosts_MultipleSpecs(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("SPEC-001", "agent-a", "plan", 100, 50, 500, true),
		newMetric("SPEC-001", "agent-b", "run", 200, 100, 1000, true),
		newMetric("SPEC-002", "agent-a", "run", 300, 150, 800, false),
	}
	result := agg.ComputeSPECCosts(metrics)

	// 두 SPEC이 모두 집계되어야 한다.
	if len(result) != 2 {
		t.Fatalf("SPEC 수 2 기대, 실제 %d", len(result))
	}

	// 결과는 TotalTokens 내림차순으로 정렬된다.
	if result[0].TotalTokens < result[1].TotalTokens {
		t.Error("TotalTokens 내림차순 정렬 실패")
	}
}

func TestComputeSPECCosts_SingleSpecMultiplePhases(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("SPEC-001", "agent-a", "plan", 100, 50, 500, true),
		newMetric("SPEC-001", "agent-b", "run", 200, 80, 1000, true),
		newMetric("SPEC-001", "agent-c", "run", 50, 20, 300, false),
	}
	result := agg.ComputeSPECCosts(metrics)

	if len(result) != 1 {
		t.Fatalf("단일 SPEC 기대, 실제 %d개", len(result))
	}
	sc := result[0]

	// 토큰 합계 검증
	wantTotal := int64((100 + 50) + (200 + 80) + (50 + 20))
	if sc.TotalTokens != wantTotal {
		t.Errorf("TotalTokens: %d 기대, 실제 %d", wantTotal, sc.TotalTokens)
	}

	// AgentCalls 검증
	if sc.AgentCalls != 3 {
		t.Errorf("AgentCalls 3 기대, 실제 %d", sc.AgentCalls)
	}

	// 페이즈 수 검증: plan(1) + run(2 → 하나의 run 페이즈로 집계)
	if len(sc.Phases) != 2 {
		t.Errorf("페이즈 수 2 기대, 실제 %d", len(sc.Phases))
	}

	// 페이즈별 AgentCalls 검증
	phaseMap := make(map[string]int)
	for _, p := range sc.Phases {
		phaseMap[p.Phase] = p.AgentCalls
	}
	if phaseMap["plan"] != 1 {
		t.Errorf("plan 페이즈 AgentCalls 1 기대, 실제 %d", phaseMap["plan"])
	}
	if phaseMap["run"] != 2 {
		t.Errorf("run 페이즈 AgentCalls 2 기대, 실제 %d", phaseMap["run"])
	}
}

func TestComputeSPECCosts_DurationAccumulation(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("SPEC-X", "ag", "run", 10, 10, 1000, true),
		newMetric("SPEC-X", "ag", "run", 10, 10, 2000, true),
	}
	result := agg.ComputeSPECCosts(metrics)
	if len(result) == 0 {
		t.Fatal("결과 없음")
	}
	wantDur := 3 * time.Second
	if result[0].Duration != wantDur {
		t.Errorf("Duration %s 기대, 실제 %s", wantDur, result[0].Duration)
	}
}

// ─── ComputeAgentStats 테스트 ────────────────────────────────────────────────

func TestComputeAgentStats_EmptyInput(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	result := agg.ComputeAgentStats(nil)

	if len(result) != 0 {
		t.Errorf("빈 입력: 길이 0 기대, 실제 %d", len(result))
	}
}

func TestComputeAgentStats_EmptyAgentNameSkipped(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("SPEC-1", "", "run", 100, 50, 200, true), // AgentName 빈 문자열 → 무시
	}
	result := agg.ComputeAgentStats(metrics)

	if len(result) != 0 {
		t.Errorf("빈 AgentName 항목 무시 기대, 실제 %d개", len(result))
	}
}

func TestComputeAgentStats_SuccessFailureRate(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "expert-backend", "run", 100, 50, 500, true),
		newMetric("", "expert-backend", "run", 200, 100, 600, true),
		newMetric("", "expert-backend", "run", 150, 75, 400, false),
	}
	result := agg.ComputeAgentStats(metrics)

	if len(result) != 1 {
		t.Fatalf("에이전트 1개 기대, 실제 %d", len(result))
	}
	s := result[0]

	if s.TotalCalls != 3 {
		t.Errorf("TotalCalls 3 기대, 실제 %d", s.TotalCalls)
	}
	if s.SuccessCount != 2 {
		t.Errorf("SuccessCount 2 기대, 실제 %d", s.SuccessCount)
	}
	if s.FailureCount != 1 {
		t.Errorf("FailureCount 1 기대, 실제 %d", s.FailureCount)
	}
	// SuccessRate: 2/3 * 100 ≈ 66.67
	if !almostEqual(s.SuccessRate, 66.667, 0.1) {
		t.Errorf("SuccessRate ~66.67 기대, 실제 %.3f", s.SuccessRate)
	}
}

func TestComputeAgentStats_ZeroDivisionSafety(t *testing.T) {
	t.Parallel()

	// totalCalls == 0 인 경우는 발생하지 않지만(집계 로직상),
	// AgentName이 빈 문자열이면 맵에 추가되지 않으므로 결과가 비어 있어야 한다.
	agg := NewAggregator()
	result := agg.ComputeAgentStats([]TaskMetric{})

	// 나누기 0 패닉 없이 빈 슬라이스 반환
	if result == nil {
		result = []AgentStats{}
	}
	if len(result) != 0 {
		t.Errorf("빈 입력: 0개 기대, 실제 %d", len(result))
	}
}

func TestComputeAgentStats_AvgCalculations(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "manager-spec", "plan", 1000, 500, 2000, true),
		newMetric("", "manager-spec", "plan", 2000, 1000, 4000, true),
	}
	result := agg.ComputeAgentStats(metrics)

	if len(result) == 0 {
		t.Fatal("결과 없음")
	}
	s := result[0]

	// AvgTokens: (1500 + 3000) / 2 = 2250
	wantAvgTokens := int64(2250)
	if s.AvgTokens != wantAvgTokens {
		t.Errorf("AvgTokens %d 기대, 실제 %d", wantAvgTokens, s.AvgTokens)
	}

	// AvgDuration: (2000 + 4000) / 2 / 1000 = 3.0 seconds
	if !almostEqual(s.AvgDuration, 3.0, 0.001) {
		t.Errorf("AvgDuration 3.0s 기대, 실제 %.3f", s.AvgDuration)
	}
}

func TestComputeAgentStats_SortedByTotalCallsDesc(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "agent-a", "run", 100, 50, 200, true),
		newMetric("", "agent-b", "run", 100, 50, 200, true),
		newMetric("", "agent-b", "run", 100, 50, 200, true),
	}
	result := agg.ComputeAgentStats(metrics)

	if len(result) < 2 {
		t.Fatal("에이전트 2개 기대")
	}
	// agent-b가 2호출 → 첫 번째여야 한다.
	if result[0].AgentName != "agent-b" {
		t.Errorf("정렬 오류: 첫 번째 에이전트 'agent-b' 기대, 실제 '%s'", result[0].AgentName)
	}
}

// ─── ComputeTrends 테스트 ────────────────────────────────────────────────────

func TestComputeTrends_EmptyMetrics(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	report := agg.ComputeTrends(nil, "daily")

	if report == nil {
		t.Fatal("nil 반환 — TrendReport 기대")
	}
	if report.Period != "daily" {
		t.Errorf("Period 'daily' 기대, 실제 '%s'", report.Period)
	}
	// 빈 입력이면 나머지 필드는 zero value여야 한다.
	if report.SPECsCreated != 0 {
		t.Errorf("SPECsCreated 0 기대, 실제 %d", report.SPECsCreated)
	}
}

func TestComputeTrends_SingleEntry(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	metrics := []TaskMetric{
		{Timestamp: ts, SpecID: "SPEC-1", AgentName: "ag", Success: true},
	}
	report := agg.ComputeTrends(metrics, "weekly")

	if report.SPECsCreated != 1 {
		t.Errorf("SPECsCreated 1 기대, 실제 %d", report.SPECsCreated)
	}
	if !almostEqual(report.AvgQuality, 100.0, 0.01) {
		t.Errorf("AvgQuality 100.0 기대, 실제 %.2f", report.AvgQuality)
	}
}

func TestComputeTrends_ImprovingTrend(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// 전반부: 성공 0/4, 후반부: 성공 4/4 → improving
	metrics := make([]TaskMetric, 0, 8)
	for i := 0; i < 4; i++ {
		metrics = append(metrics, TaskMetric{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			AgentName: "ag",
			Success:   false,
		})
	}
	for i := 4; i < 8; i++ {
		metrics = append(metrics, TaskMetric{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			AgentName: "ag",
			Success:   true,
		})
	}

	report := agg.ComputeTrends(metrics, "daily")
	if report.QualityTrend != "improving" {
		t.Errorf("QualityTrend 'improving' 기대, 실제 '%s'", report.QualityTrend)
	}
}

func TestComputeTrends_DecliningTrend(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// 전반부: 성공 4/4, 후반부: 성공 0/4 → declining
	metrics := make([]TaskMetric, 0, 8)
	for i := 0; i < 4; i++ {
		metrics = append(metrics, TaskMetric{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			AgentName: "ag",
			Success:   true,
		})
	}
	for i := 4; i < 8; i++ {
		metrics = append(metrics, TaskMetric{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			AgentName: "ag",
			Success:   false,
		})
	}

	report := agg.ComputeTrends(metrics, "daily")
	if report.QualityTrend != "declining" {
		t.Errorf("QualityTrend 'declining' 기대, 실제 '%s'", report.QualityTrend)
	}
}

func TestComputeTrends_StableTrend(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// 전반부 성공률 ≈ 후반부 성공률 → stable (차이 < 0.05)
	metrics := make([]TaskMetric, 0, 8)
	for i := 0; i < 8; i++ {
		metrics = append(metrics, TaskMetric{
			Timestamp: base.Add(time.Duration(i) * time.Hour),
			AgentName: "ag",
			Success:   true, // 둘 다 100% → 차이 0
		})
	}

	report := agg.ComputeTrends(metrics, "monthly")
	if report.QualityTrend != "stable" {
		t.Errorf("QualityTrend 'stable' 기대, 실제 '%s'", report.QualityTrend)
	}
}

func TestComputeTrends_StartEndDate(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	metrics := []TaskMetric{
		{Timestamp: t2, AgentName: "ag", Success: true},
		{Timestamp: t1, AgentName: "ag", Success: true},
		{Timestamp: t3, AgentName: "ag", Success: true},
	}
	report := agg.ComputeTrends(metrics, "daily")

	if !report.StartDate.Equal(t1) {
		t.Errorf("StartDate %v 기대, 실제 %v", t1, report.StartDate)
	}
	if !report.EndDate.Equal(t2) {
		t.Errorf("EndDate %v 기대, 실제 %v", t2, report.EndDate)
	}
}

// ─── ComputeBudgetStatus 테스트 ──────────────────────────────────────────────

func TestComputeBudgetStatus_UnderBudget(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "ag", "run", 100, 50, 200, true), // 150 토큰
	}
	status := agg.ComputeBudgetStatus(metrics, 1000, 80.0)

	if status.UsedTokens != 150 {
		t.Errorf("UsedTokens 150 기대, 실제 %d", status.UsedTokens)
	}
	if status.RemainingTokens != 850 {
		t.Errorf("RemainingTokens 850 기대, 실제 %d", status.RemainingTokens)
	}
	if !almostEqual(status.UsagePercent, 15.0, 0.001) {
		t.Errorf("UsagePercent 15.0 기대, 실제 %.3f", status.UsagePercent)
	}
	if status.IsAlert {
		t.Error("IsAlert false 기대 (15% < 80%)")
	}
}

func TestComputeBudgetStatus_OverBudget(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "ag", "run", 800, 400, 200, true), // 1200 토큰 > 예산 1000
	}
	status := agg.ComputeBudgetStatus(metrics, 1000, 80.0)

	if status.RemainingTokens != 0 {
		t.Errorf("RemainingTokens 0 기대 (초과 시 음수 아님), 실제 %d", status.RemainingTokens)
	}
	if !almostEqual(status.UsagePercent, 120.0, 0.001) {
		t.Errorf("UsagePercent 120.0 기대, 실제 %.3f", status.UsagePercent)
	}
	if !status.IsAlert {
		t.Error("IsAlert true 기대 (120% > 80%)")
	}
}

func TestComputeBudgetStatus_ZeroBudget(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "ag", "run", 100, 50, 200, true),
	}
	// totalBudget == 0 → 나누기 0 방지, UsagePercent == 0
	status := agg.ComputeBudgetStatus(metrics, 0, 80.0)

	if status.UsagePercent != 0.0 {
		t.Errorf("ZeroBudget: UsagePercent 0 기대, 실제 %.3f", status.UsagePercent)
	}
	// 예산 0, 사용 150 → remaining 계산상 음수이지만 0으로 클램프
	if status.RemainingTokens != 0 {
		t.Errorf("RemainingTokens 0 기대 (음수 클램프), 실제 %d", status.RemainingTokens)
	}
}

func TestComputeBudgetStatus_AtAlertThreshold(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		newMetric("", "ag", "run", 800, 0, 200, true), // 800 토큰, 예산 1000
	}
	// usagePct == alertPct == 80 → IsAlert true (>= 조건)
	status := agg.ComputeBudgetStatus(metrics, 1000, 80.0)

	if !status.IsAlert {
		t.Errorf("IsAlert true 기대 (사용률 %v%% >= 임계값 80%%)", status.UsagePercent)
	}
}

func TestComputeBudgetStatus_EmptyMetrics(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	status := agg.ComputeBudgetStatus(nil, 10000, 90.0)

	if status.UsedTokens != 0 {
		t.Errorf("UsedTokens 0 기대, 실제 %d", status.UsedTokens)
	}
	if status.RemainingTokens != 10000 {
		t.Errorf("RemainingTokens 10000 기대, 실제 %d", status.RemainingTokens)
	}
	if status.IsAlert {
		t.Error("IsAlert false 기대")
	}
}

// ─── Aggregator 복합 테스트 ──────────────────────────────────────────────────

func TestNewAggregator(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	if agg == nil {
		t.Fatal("NewAggregator nil 반환")
	}
}

func TestComputeSPECCosts_TokenSplit(t *testing.T) {
	t.Parallel()

	agg := NewAggregator()
	metrics := []TaskMetric{
		{SpecID: "S1", InputTokens: 300, OutputTokens: 200},
	}
	result := agg.ComputeSPECCosts(metrics)
	if len(result) == 0 {
		t.Fatal("결과 없음")
	}
	sc := result[0]
	if sc.InputTokens != 300 {
		t.Errorf("InputTokens 300 기대, 실제 %d", sc.InputTokens)
	}
	if sc.OutputTokens != 200 {
		t.Errorf("OutputTokens 200 기대, 실제 %d", sc.OutputTokens)
	}
	if sc.TotalTokens != 500 {
		t.Errorf("TotalTokens 500 기대, 실제 %d", sc.TotalTokens)
	}
}
