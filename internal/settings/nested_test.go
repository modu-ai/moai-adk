package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeNestedFixture는 quality.yaml + git-convention.yaml 섹션 파일을 t.TempDir()
// 아래에 생성하여 config 매니저가 LoadRaw 할 수 있는 최소 프로젝트를 만든다.
func writeNestedFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	// quality.yaml 의 정규 top-level 키는 `constitution:` 다(LoadRaw 매핑 — 기존
	// internal/web/projectnested_test.go seedNestedProject 와 동일). `quality:` 가
	// 아님에 주의: LoadRaw 는 constitution 섹션을 cfg.Quality 로 매핑한다.
	quality := `constitution:
  development_mode: tdd
  test_coverage_target: 80
  enforce_quality: true
  tdd_settings:
    min_coverage_per_commit: 70
`
	gitConv := `git_convention:
  convention: auto
  auto_detection:
    enabled: true
    confidence_threshold: 0.6
    sample_size: 100
  validation:
    enforce_on_push: false
`
	if err := os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"), []byte(quality), 0o644); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "git-convention.yaml"), []byte(gitConv), 0o644); err != nil {
		t.Fatalf("write git-convention.yaml: %v", err)
	}
	return root
}

// TestSharedNestedSeamRoundTrip는 재배치된 seam(WriteProjectNestedConfig →
// ReadProjectNestedConfig)이 7개 중첩 필드를 모두 round-trip 함을 검증한다(M2).
func TestSharedNestedSeamRoundTrip(t *testing.T) {
	root := writeNestedFixture(t)

	form := NestedForm{
		CoverageTarget: 92, CoverageTargetSet: true,
		EnforceQuality: false, EnforceQualitySet: true,
		MinCoverage: 85, MinCoverageSet: true,
		Confidence: 0.8, ConfidenceSet: true,
		AutoEnabled: false, AutoEnabledSet: true,
		SampleSize: 250, SampleSizeSet: true,
		EnforceOnPush: true, EnforceOnPushSet: true,
	}
	if err := WriteProjectNestedConfig(root, form); err != nil {
		t.Fatalf("WriteProjectNestedConfig: %v", err)
	}

	got, err := ReadProjectNestedConfig(root)
	if err != nil {
		t.Fatalf("ReadProjectNestedConfig: %v", err)
	}

	if got.CoverageTarget != "92" {
		t.Errorf("CoverageTarget = %q, want 92", got.CoverageTarget)
	}
	if got.EnforceQuality != false {
		t.Errorf("EnforceQuality = %v, want false", got.EnforceQuality)
	}
	if got.MinCoverage != "85" {
		t.Errorf("MinCoverage = %q, want 85", got.MinCoverage)
	}
	if got.ConfidenceThreshold != "0.8" {
		t.Errorf("ConfidenceThreshold = %q, want 0.8", got.ConfidenceThreshold)
	}
	if got.AutoDetectionEnabled != false {
		t.Errorf("AutoDetectionEnabled = %v, want false", got.AutoDetectionEnabled)
	}
	if got.SampleSize != "250" {
		t.Errorf("SampleSize = %q, want 250", got.SampleSize)
	}
	if got.EnforceOnPush != true {
		t.Errorf("EnforceOnPush = %v, want true", got.EnforceOnPush)
	}
}

// TestSharedNestedSeamEmptyPreserve는 *Set 플래그가 꺼진 필드가 디스크 값을
// 유지함을 검증한다(empty=preserve, REQ-WC10-012). coverage_target 만 제출하고
// 나머지는 미제출 → 나머지는 fixture 값 그대로여야 한다.
func TestSharedNestedSeamEmptyPreserve(t *testing.T) {
	root := writeNestedFixture(t)

	form := NestedForm{CoverageTarget: 95, CoverageTargetSet: true}
	if err := WriteProjectNestedConfig(root, form); err != nil {
		t.Fatalf("WriteProjectNestedConfig: %v", err)
	}

	got, err := ReadProjectNestedConfig(root)
	if err != nil {
		t.Fatalf("ReadProjectNestedConfig: %v", err)
	}

	// 제출한 필드는 변경.
	if got.CoverageTarget != "95" {
		t.Errorf("CoverageTarget = %q, want 95", got.CoverageTarget)
	}
	// 미제출 필드는 fixture 값 보존.
	if got.MinCoverage != "70" {
		t.Errorf("MinCoverage = %q, want 70 (preserved)", got.MinCoverage)
	}
	if got.EnforceQuality != true {
		t.Errorf("EnforceQuality = %v, want true (preserved)", got.EnforceQuality)
	}
	if got.ConfidenceThreshold != "0.6" {
		t.Errorf("ConfidenceThreshold = %q, want 0.6 (preserved)", got.ConfidenceThreshold)
	}
	if got.SampleSize != "100" {
		t.Errorf("SampleSize = %q, want 100 (preserved)", got.SampleSize)
	}
}

// TestSharedNestedSeamWholeSectionPreserve는 whole-section-copy 의미를 검증한다:
// 중첩 필드 하나를 변경해도 같은 섹션의 비-UI 필드(development_mode, convention)는
// byte-identical 하게 유지된다(B4 / AP-2 핵심 불변식).
func TestSharedNestedSeamWholeSectionPreserve(t *testing.T) {
	root := writeNestedFixture(t)

	form := NestedForm{CoverageTarget: 99, CoverageTargetSet: true, ConfidenceSet: true, Confidence: 0.9}
	if err := WriteProjectNestedConfig(root, form); err != nil {
		t.Fatalf("WriteProjectNestedConfig: %v", err)
	}

	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	// 같은 quality 섹션의 development_mode 가 유지되어야 한다.
	if string(cfg.Quality.DevelopmentMode) != "tdd" {
		t.Errorf("Quality.DevelopmentMode = %q, want tdd (preserved through whole-section-copy)", cfg.Quality.DevelopmentMode)
	}
	// 같은 git_convention 섹션의 convention 이 유지되어야 한다.
	if cfg.GitConvention.Convention != "auto" {
		t.Errorf("GitConvention.Convention = %q, want auto (preserved)", cfg.GitConvention.Convention)
	}
}
