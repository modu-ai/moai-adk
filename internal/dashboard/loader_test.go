package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ─── 테스트 헬퍼 ────────────────────────────────────────────────────────────

// writeJSONL은 TaskMetric 슬라이스를 JSONL 형식으로 지정 경로에 쓴다.
func writeJSONL(t *testing.T, path string, metrics []TaskMetric) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("파일 생성 실패 %s: %v", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, m := range metrics {
		if err := enc.Encode(m); err != nil {
			t.Fatalf("JSON 인코딩 실패: %v", err)
		}
	}
}

// sampleMetrics는 테스트용 기본 메트릭 슬라이스를 반환한다.
func sampleMetrics(n int) []TaskMetric {
	metrics := make([]TaskMetric, n)
	for i := range metrics {
		metrics[i] = TaskMetric{
			Timestamp:    time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC),
			SessionID:    "ses-test",
			SpecID:       "SPEC-001",
			AgentName:    "expert-backend",
			Phase:        "run",
			Action:       "implement",
			InputTokens:  int64(100 * (i + 1)),
			OutputTokens: int64(50 * (i + 1)),
			DurationMS:   int64(500 * (i + 1)),
			Success:      i%2 == 0,
		}
	}
	return metrics
}

// ─── Loader 생성 테스트 ──────────────────────────────────────────────────────

func TestNewLoader(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	l := NewLoader(dir)
	if l == nil {
		t.Fatal("NewLoader nil 반환")
	}
}

// ─── LoadMetrics: 빈 디렉토리 ────────────────────────────────────────────────

func TestLoadMetrics_EmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	l := NewLoader(dir)

	metrics, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("빈 디렉토리: 에러 없이 빈 슬라이스 기대, 실제 에러: %v", err)
	}
	// 파일 없음 → 빈 슬라이스 반환
	if len(metrics) != 0 {
		t.Errorf("빈 디렉토리: 길이 0 기대, 실제 %d", len(metrics))
	}
}

// ─── LoadMetrics: 단일 파일 (task-metrics.jsonl) ────────────────────────────

func TestLoadMetrics_SingleFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	want := sampleMetrics(3)
	writeJSONL(t, filepath.Join(dir, "task-metrics.jsonl"), want)

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("단일 파일 로드 에러: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("메트릭 수 %d 기대, 실제 %d", len(want), len(got))
	}
}

// ─── LoadMetrics: glob 패턴 (task-metrics*.jsonl) ──────────────────────────

func TestLoadMetrics_GlobPattern(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// task-metrics-2025-01.jsonl, task-metrics-2025-02.jsonl 두 파일 생성
	writeJSONL(t, filepath.Join(dir, "task-metrics-2025-01.jsonl"), sampleMetrics(2))
	writeJSONL(t, filepath.Join(dir, "task-metrics-2025-02.jsonl"), sampleMetrics(3))

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("glob 로드 에러: %v", err)
	}
	// 두 파일 합계: 2 + 3 = 5
	if len(got) != 5 {
		t.Errorf("메트릭 수 5 기대, 실제 %d", len(got))
	}
}

// ─── LoadMetrics: 비glob 단일 파일 (task-metrics.jsonl) 우선순위 확인 ────────

func TestLoadMetrics_SingleFileFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// glob과 단일 파일이 공존하지 않는 상황 — glob 미매칭, 단일 파일만 존재
	writeJSONL(t, filepath.Join(dir, "task-metrics.jsonl"), sampleMetrics(4))

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("단일 파일 폴백 에러: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("메트릭 수 4 기대, 실제 %d", len(got))
	}
}

// ─── LoadMetrics: 비관련 파일 무시 ──────────────────────────────────────────

func TestLoadMetrics_UnrelatedFilesIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// glob에 매칭되지 않는 파일들
	writeJSONL(t, filepath.Join(dir, "other-metrics.jsonl"), sampleMetrics(5))
	writeJSONL(t, filepath.Join(dir, "metrics.json"), sampleMetrics(2))

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("에러: %v", err)
	}
	// task-metrics*.jsonl 도, task-metrics.jsonl 도 없으므로 0개
	if len(got) != 0 {
		t.Errorf("비관련 파일 무시: 0개 기대, 실제 %d", len(got))
	}
}

// ─── LoadMetrics: 잘못된 JSON 라인 무시 ────────────────────────────────────

func TestLoadMetrics_MalformedJSONSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "task-metrics.jsonl")

	// 유효한 줄 1개 + 잘못된 JSON 1개 + 유효한 줄 1개 혼합
	valid := sampleMetrics(1)[0]
	validJSON, _ := json.Marshal(valid)

	lines := []string{
		string(validJSON),
		`{"broken json`,
		string(validJSON),
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("에러: %v", err)
	}
	// 잘못된 JSON 줄은 건너뛰고 유효한 2개만 반환
	if len(got) != 2 {
		t.Errorf("유효 메트릭 수 2 기대, 실제 %d", len(got))
	}
}

// ─── LoadMetrics: 완전히 잘못된 파일 무시 ──────────────────────────────────

func TestLoadMetrics_EntirelyMalformedFileSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// 전체가 잘못된 파일과 정상 파일 공존
	if err := os.WriteFile(
		filepath.Join(dir, "task-metrics-bad.jsonl"),
		[]byte("not json at all\nstill not json"),
		0644,
	); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}
	writeJSONL(t, filepath.Join(dir, "task-metrics-good.jsonl"), sampleMetrics(2))

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("에러: %v", err)
	}
	// 잘못된 파일의 줄은 모두 건너뛰고 정상 파일 2개만 반환
	if len(got) != 2 {
		t.Errorf("정상 파일 메트릭 수 2 기대, 실제 %d", len(got))
	}
}

// ─── LoadMetrics: 빈 파일 ────────────────────────────────────────────────────

func TestLoadMetrics_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// 빈 파일 생성
	if err := os.WriteFile(
		filepath.Join(dir, "task-metrics.jsonl"),
		[]byte{},
		0644,
	); err != nil {
		t.Fatalf("파일 쓰기 실패: %v", err)
	}

	l := NewLoader(dir)
	got, err := l.LoadMetrics()
	if err != nil {
		t.Fatalf("빈 파일 로드 에러: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("빈 파일: 0개 기대, 실제 %d", len(got))
	}
}

// ─── loadFile 내부 테스트 ────────────────────────────────────────────────────

func TestLoadFile_DirectRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.jsonl")
	want := sampleMetrics(5)
	writeJSONL(t, path, want)

	l := NewLoader(dir)
	got, err := l.loadFile(path)
	if err != nil {
		t.Fatalf("loadFile 에러: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("메트릭 수 %d 기대, 실제 %d", len(want), len(got))
	}
}

func TestLoadFile_NonExistentPath(t *testing.T) {
	t.Parallel()

	l := NewLoader(t.TempDir())
	_, err := l.loadFile("/nonexistent/path/metrics.jsonl")

	// 파일 없으면 에러 반환 기대
	if err == nil {
		t.Error("존재하지 않는 파일: 에러 기대, 실제 nil")
	}
}

// ─── Renderer: 출력 형식 검증 ────────────────────────────────────────────────

func TestRenderSummary_ContainsKeyFields(t *testing.T) {
	t.Parallel()

	r := NewRenderer()
	summary := DashboardSummary{
		ActiveSPECs:     3,
		CompletedSPECs:  7,
		TotalTokensUsed: 150_000,
		Budget: BudgetStatus{
			TotalBudget:  200_000,
			UsedTokens:   150_000,
			UsagePercent: 75.0,
		},
	}
	output := r.RenderSummary(summary)

	checks := []string{
		"MoAI Dashboard",
		"Active SPECs",
		"Completed SPECs",
		"Budget",
	}
	for _, c := range checks {
		if !strings.Contains(output, c) {
			t.Errorf("출력에 '%s' 포함 기대", c)
		}
	}
}

func TestRenderSummary_AlertMarker(t *testing.T) {
	t.Parallel()

	r := NewRenderer()
	summary := DashboardSummary{
		Budget: BudgetStatus{
			TotalBudget:    1000,
			UsedTokens:     900,
			UsagePercent:   90.0,
			AlertThreshold: 80.0,
			IsAlert:        true,
		},
	}
	output := r.RenderSummary(summary)

	if !strings.Contains(output, "ALERT") {
		t.Error("예산 초과 경고 'ALERT' 포함 기대")
	}
}

func TestRenderBudget_ProgressBar(t *testing.T) {
	t.Parallel()

	r := NewRenderer()
	status := BudgetStatus{
		TotalBudget:    1000,
		UsedTokens:     500,
		UsagePercent:   50.0,
		AlertThreshold: 80.0,
		IsAlert:        false,
	}
	output := r.RenderBudget(status)

	// 프로그레스 바 문자 포함 여부 확인
	if !strings.Contains(output, "█") {
		t.Error("프로그레스 바 '█' 포함 기대")
	}
	if !strings.Contains(output, "Budget") {
		t.Error("'Budget' 헤더 포함 기대")
	}
}

func TestRenderTrends_AllTrendIndicators(t *testing.T) {
	t.Parallel()

	r := NewRenderer()
	tests := []struct {
		trend    string
		contains string
	}{
		{"improving", "Improving"},
		{"declining", "Declining"},
		{"stable", "Stable"},
		{"unknown-value", "Unknown"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.trend, func(t *testing.T) {
			t.Parallel()
			report := &TrendReport{
				Period:       "daily",
				QualityTrend: tc.trend,
			}
			output := r.RenderTrends(report)
			if !strings.Contains(output, tc.contains) {
				t.Errorf("트렌드 '%s': '%s' 포함 기대\n출력: %s", tc.trend, tc.contains, output)
			}
		})
	}
}

func TestRenderSPECCosts_TableHeader(t *testing.T) {
	t.Parallel()

	r := NewRenderer()
	costs := []SPECCost{
		{SpecID: "SPEC-001", TotalTokens: 5000, AgentCalls: 3},
		{SpecID: "SPEC-002", TotalTokens: 2000, AgentCalls: 1},
	}
	output := r.RenderSPECCosts(costs)

	if !strings.Contains(output, "SPEC Cost Analysis") {
		t.Error("'SPEC Cost Analysis' 헤더 포함 기대")
	}
	if !strings.Contains(output, "SPEC-001") {
		t.Error("'SPEC-001' 포함 기대")
	}
}

func TestRenderAgentStats_TableHeader(t *testing.T) {
	t.Parallel()

	r := NewRenderer()
	stats := []AgentStats{
		{AgentName: "expert-backend", TotalCalls: 10, SuccessCount: 9, SuccessRate: 90.0},
	}
	output := r.RenderAgentStats(stats)

	if !strings.Contains(output, "Agent Performance") {
		t.Error("'Agent Performance' 헤더 포함 기대")
	}
	if !strings.Contains(output, "expert-backend") {
		t.Error("'expert-backend' 포함 기대")
	}
}

// ─── formatTokens 단위 테스트 ────────────────────────────────────────────────

func TestFormatTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tokens int64
		want   string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{999_999, "1000.0K"},
		{1_000_000, "1.0M"},
		{1_500_000, "1.5M"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := formatTokens(tc.tokens)
			if got != tc.want {
				t.Errorf("formatTokens(%d): '%s' 기대, 실제 '%s'", tc.tokens, tc.want, got)
			}
		})
	}
}

// ─── formatDuration 단위 테스트 ─────────────────────────────────────────────

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "0.5s"},
		{30 * time.Second, "30.0s"},
		{90 * time.Second, "1.5m"},
		{2 * time.Hour, "2.0h"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := formatDuration(tc.d)
			if got != tc.want {
				t.Errorf("formatDuration(%v): '%s' 기대, 실제 '%s'", tc.d, tc.want, got)
			}
		})
	}
}
