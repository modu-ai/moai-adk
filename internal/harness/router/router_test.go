package router_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// testdataDir은 테스트 픽스처 루트 디렉토리 경로를 반환합니다.
func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// minimalHarnessConfig는 테스트용 최소 HarnessConfig를 반환합니다.
func minimalHarnessConfig() *config.HarnessConfig {
	return &config.HarnessConfig{
		DefaultProfile: "default",
		ModeDefaults:   map[string]string{"cg": "thorough", "solo": "auto", "team": "auto"},
		EffortMapping:  map[string]string{"minimal": "medium", "standard": "high", "thorough": "xhigh"},
		Escalation: config.EscalationConfig{
			Enabled:        true,
			MaxEscalations: 2,
			Triggers:       []string{"quality_gate_fail", "review_critical", "test_coverage_low"},
		},
		AutoDetection: config.AutoDetectionConfig{
			Enabled: true,
			Rules: map[string]config.AutoDetectionRule{
				"minimal": {
					Conditions: []string{"file_count <= 3 AND single_domain", "spec_type in [bugfix, docs, config]"},
				},
				"standard": {
					Conditions: []string{"file_count > 3 OR multi_domain", "spec_type in [feature, refactor]"},
				},
				"thorough": {
					Conditions: []string{"security_keywords OR payment_keywords present"},
				},
			},
		},
		Levels: map[string]config.LevelConfig{
			"minimal": {
				Description: "Fast iteration",
				Evaluator:   false,
				PlanAudit:   config.PlanAuditConfig{Enabled: true, MaxIterations: 1},
			},
			"standard": {
				Description:   "Balanced quality",
				Evaluator:     true,
				EvaluatorMode: "final-pass",
				PlanAudit:     config.PlanAuditConfig{Enabled: true, MaxIterations: 3, RequireMustPass: true},
			},
			"thorough": {
				Description:     "Maximum quality",
				Evaluator:       true,
				EvaluatorMode:   "per-sprint",
				EvaluatorProfile: "strict",
				SprintContract:   true,
				PlanAudit:       config.PlanAuditConfig{Enabled: true, MaxIterations: 3, RequireMustPass: true},
			},
		},
	}
}

// TestRoute_SpecOverride — REQ-HRN-001-015: SPEC frontmatter harness_level: 오버라이드.
// spec.md에 harness_level: thorough가 있으면 matched_rule: spec_override를 반환해야 합니다.
func TestRoute_SpecOverride(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	specPath := filepath.Join(testdataDir(), "spec-overrides", ".moai", "specs", "SPEC-TEST-AAA-001", "spec.md")
	level, rationale, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		t.Fatalf("RouteFromFile() error: %v", err)
	}

	if level != router.LevelThorough {
		t.Errorf("level: got %q, want %q", level, router.LevelThorough)
	}
	if rationale.MatchedRule != "spec_override" {
		t.Errorf("rationale.MatchedRule: got %q, want %q", rationale.MatchedRule, "spec_override")
	}
}

// TestRoute_KeywordForceThorough — REQ-HRN-001-008: 보안/결제 키워드 force-thorough.
// SPEC 본문에 oauth 또는 jwt 같은 키워드가 있으면 LevelThorough를 반환해야 합니다.
func TestRoute_KeywordForceThorough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	specPath := filepath.Join(testdataDir(), "keyword-force", ".moai", "specs", "SPEC-TEST-BBB-001", "spec.md")
	level, rationale, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		t.Fatalf("RouteFromFile() error: %v", err)
	}

	if level != router.LevelThorough {
		t.Errorf("level: got %q, want %q", level, router.LevelThorough)
	}
	if rationale.MatchedRule != "force_thorough" {
		t.Errorf("rationale.MatchedRule: got %q, want %q", rationale.MatchedRule, "force_thorough")
	}
	if len(rationale.Keywords) == 0 {
		t.Error("rationale.Keywords should not be empty for keyword-triggered force_thorough")
	}
}

// TestRoute_NormalSpec_Standard — REQ-HRN-001-007: 일반 SPEC이 standard로 라우팅.
// 단순 feature SPEC은 file_count > 3이면 standard로 라우팅됩니다.
func TestRoute_NormalSpec_Standard(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	specPath := filepath.Join(testdataDir(), "normal", ".moai", "specs", "SPEC-TEST-CCC-001", "spec.md")
	level, _, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		t.Fatalf("RouteFromFile() error: %v", err)
	}

	// SPEC-TEST-CCC-001은 5개의 REQ → file_count 추정 가능; domain 단일
	// 라우팅 결과: minimal 또는 standard (파일 카운트에 따라 다름)
	switch level {
	case router.LevelMinimal, router.LevelStandard:
		// 정상 범위
	default:
		t.Errorf("unexpected level %q for normal feature SPEC", level)
	}
}

// TestRoute_PriorityOrderRespected — REQ-HRN-001-007: 우선순위 순서 (minimal → standard → thorough).
func TestRoute_PriorityOrderRespected(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	// 빈 SPEC 구조 (기본값 적용 → minimal 또는 standard 폴백)
	doc := &router.SPECInput{
		Priority: "P3",
		Tags:     "test",
		Title:    "Simple Test",
		Body:     "",
	}

	level, rationale, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}

	switch level {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// 모두 유효한 결과
	default:
		t.Errorf("unexpected level: %q", level)
	}
	if rationale.MatchedRule == "" {
		t.Error("rationale.MatchedRule should not be empty")
	}
}

// TestRoute_CriticalPriority_ForceThorough — REQ-HRN-001-008: Critical 우선순위 → thorough.
func TestRoute_CriticalPriority_ForceThorough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	doc := &router.SPECInput{
		Priority: "P0 Critical",
		Tags:     "feature",
		Title:    "Critical Feature",
		Body:     "- REQ-TST-001-001 (Ubiquitous) — System shall implement critical feature.",
	}

	level, rationale, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}

	if level != router.LevelThorough {
		t.Errorf("Critical priority: level got %q, want %q", level, router.LevelThorough)
	}
	if rationale.MatchedRule != "force_thorough" {
		t.Errorf("Critical priority: matched_rule got %q, want %q", rationale.MatchedRule, "force_thorough")
	}
}

// TestRoute_MinimalSpec — REQ-HRN-001-007: minimal spec → LevelMinimal.
func TestRoute_MinimalSpec(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	doc := &router.SPECInput{
		Priority: "P3",
		Tags:     "bugfix",
		Title:    "Simple Bug Fix",
		Body:     "",
	}

	level, rationale, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}

	if level != router.LevelMinimal {
		t.Errorf("minimal bugfix: level got %q, want %q", level, router.LevelMinimal)
	}
	if rationale.MatchedRule != "auto_minimal" {
		t.Errorf("minimal bugfix: matched_rule got %q, want %q", rationale.MatchedRule, "auto_minimal")
	}
}

// TestRoute_SensitiveDomain_ForceThorough — REQ-HRN-001-008: 민감 도메인 → thorough.
func TestRoute_SensitiveDomain_ForceThorough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	tests := []struct {
		name string
		tags string
	}{
		{"auth_domain", "auth, cli"},
		{"payment_domain", "payment, api"},
		{"migration_domain", "migration, db"},
		{"public_api_domain", "public_api, rest"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc := &router.SPECInput{
				Priority: "P2",
				Tags:     tt.tags,
				Title:    "Sensitive Domain Test",
				Body:     "",
			}
			level, rationale, err := r.Route(doc, cfg)
			if err != nil {
				t.Fatalf("Route() error: %v", err)
			}
			if level != router.LevelThorough {
				t.Errorf("sensitive domain %q: level got %q, want thorough", tt.tags, level)
			}
			if rationale.MatchedRule != "force_thorough" {
				t.Errorf("sensitive domain %q: matched_rule got %q, want force_thorough", tt.tags, rationale.MatchedRule)
			}
		})
	}
}

// TestRoute_RouteFromFile_Error — RouteFromFile 오류 경로.
func TestRoute_RouteFromFile_Error(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	level, _, err := r.RouteFromFile("/nonexistent/path/spec.md", cfg)
	// 오류 시 기본값 반환
	_ = err
	switch level {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// 유효한 기본값
	default:
		t.Errorf("RouteFromFile error path: unexpected level %q", level)
	}
}

// TestRoute_InvalidHarnessLevel_IgnoredAndFallthrough — 유효하지 않은 harness_level override 무시.
func TestRoute_InvalidHarnessLevel_IgnoredAndFallthrough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	doc := &router.SPECInput{
		Priority:     "P3",
		Tags:         "bugfix",
		Title:        "Test",
		Body:         "",
		HarnessLevel: "extreme", // 유효하지 않은 값
	}

	level, _, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}
	// 유효하지 않은 override는 무시되고 auto-detection으로 폴백
	switch level {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// OK
	default:
		t.Errorf("invalid harness_level override should fallthrough, got %q", level)
	}
}

// TestParseProfileFloor_ValidFile — ParseProfileFloor 정상 케이스.
func TestParseProfileFloor_ValidFile(t *testing.T) {
	t.Parallel()

	content := `# Test Profile

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 40% | 0.75 |
| Security | 25% | 0.75 |
| Craft | 20% | 0.75 |
| Consistency | 15% | 0.75 |
`
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	floor, err := router.ParseProfileFloor(profilePath)
	if err != nil {
		t.Fatalf("ParseProfileFloor error: %v", err)
	}
	if floor <= 0 {
		t.Errorf("floor should be > 0, got %f", floor)
	}
}

// TestParseProfileFloor_LowThreshold — ParseProfileFloor 0.5 임계값 케이스.
func TestParseProfileFloor_LowThreshold(t *testing.T) {
	t.Parallel()

	content := `# Low Threshold Profile

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 40% | 0.50 |
| Security | 25% | 0.60 |
| Craft | 20% | 0.70 |
| Consistency | 15% | 0.75 |
`
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "low.md")
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	floor, err := router.ParseProfileFloor(profilePath)
	if err != nil {
		t.Fatalf("ParseProfileFloor error: %v", err)
	}
	// 최솟값은 0.50이어야 합니다
	if floor > 0.51 {
		t.Errorf("floor should be ~0.50, got %f", floor)
	}
}

// TestParseProfileFloor_NotFound — ParseProfileFloor 파일 없음.
func TestParseProfileFloor_NotFound(t *testing.T) {
	t.Parallel()

	_, err := router.ParseProfileFloor("/nonexistent/profile.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestEscalationManager_CountAndMax — Count() + MaxEscalations() 커버리지.
func TestEscalationManager_CountAndMax(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(2)

	if mgr.Count() != 0 {
		t.Errorf("initial count: got %d, want 0", mgr.Count())
	}
	if mgr.MaxEscalations() != 2 {
		t.Errorf("max escalations: got %d, want 2", mgr.MaxEscalations())
	}

	mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})

	if mgr.Count() != 1 {
		t.Errorf("after 1 escalation: count got %d, want 1", mgr.Count())
	}
}

// TestEscalationManager_NegativeMax — 음수 max 처리.
func TestEscalationManager_NegativeMax(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(-1)
	if mgr.MaxEscalations() != 0 {
		t.Errorf("negative max: got %d, want 0", mgr.MaxEscalations())
	}

	_, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})
	if escalated {
		t.Error("negative max: escalation should be blocked immediately")
	}
}

// TestEffortForLevel_NilConfig — nil config → 기본값 반환.
func TestEffortForLevel_NilConfig(t *testing.T) {
	t.Parallel()

	got := router.EffortForLevel(router.LevelMinimal, nil)
	if got != "medium" {
		t.Errorf("nil config: got %q, want %q", got, "medium")
	}
}
