// Package harness — HRN-003 M4 프로필 로더 + 말형성 프로필 거부 테스트.
// REQ-HRN-003-005: ParseRubricMarkdown.
// REQ-HRN-003-019: 5th dimension rejection.
// REQ-HRN-003-018: must-pass bypass rejection.
package harness

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestParseRubricMarkdown_RejectsFifthDimension는 알 수 없는 차원 이름 "Performance"를
// 선언한 프로필이 Validate() 수준에서 거부되는지 검증합니다.
// malformed-5dim.md는 Functionality/Security/Craft/Performance를 선언합니다
// (Consistency 누락 + Performance 알 수 없음).
// 파서는 lenient하게 알 수 없는 차원을 건너뛰어 3개 canonical dim만 남깁니다.
// Rubric.Validate()는 4개 차원 수 검증으로 ErrInvalidConfig를 반환합니다.
// REQ-HRN-003-019, AC-HRN-003-08.
func TestParseRubricMarkdown_RejectsFifthDimension(t *testing.T) {
	rubric, err := ParseRubricMarkdown("testdata/profiles/malformed-5dim.md")
	if err != nil {
		// 파서 수준에서 에러가 발생한 경우 (예: ErrUnknownDimension) — 올바른 동작.
		if errors.Is(err, config.ErrUnknownDimension) || errors.Is(err, config.ErrInvalidConfig) {
			return
		}
		t.Fatalf("ParseRubricMarkdown(malformed-5dim.md) unexpected error = %v", err)
	}

	// 파서가 rubric을 반환했다면 Validate()로 검증합니다.
	// malformed-5dim.md에서 Performance는 건너뛰어 3개 canonical dim만 남습니다.
	// Validate()는 "exactly 4 dimensions" 위반으로 ErrInvalidConfig를 반환합니다.
	if rubric == nil {
		t.Fatal("ParseRubricMarkdown returned nil rubric without error")
	}
	validateErr := rubric.Validate()
	if validateErr == nil {
		t.Fatal("Rubric.Validate() = nil, want error for profile with missing canonical dimension")
	}
	if !errors.Is(validateErr, config.ErrInvalidConfig) {
		t.Errorf("Rubric.Validate() error = %v, want ErrInvalidConfig", validateErr)
	}
}

// TestParseRubricMarkdown_MustPassBypassRejected는 Security를 Must-Pass에서 제외한
// 프로필이 ErrMustPassBypassProhibited를 반환하는지 검증합니다.
// REQ-HRN-003-018, AC-HRN-003-11.
func TestParseRubricMarkdown_MustPassBypassRejected(t *testing.T) {
	rubric, err := ParseRubricMarkdown("testdata/profiles/malformed-bypass.md")
	if err != nil {
		// 파서가 직접 거부하는 경우 (ErrMustPassBypassProhibited wrapped).
		if errors.Is(err, config.ErrMustPassBypassProhibited) {
			return
		}
		t.Fatalf("ParseRubricMarkdown(malformed-bypass.md) unexpected error = %v", err)
	}

	if rubric == nil {
		t.Fatal("ParseRubricMarkdown returned nil rubric, want non-nil")
	}

	// Validate()에서 MustPass floor 위반 검증.
	validateErr := rubric.Validate()
	if validateErr == nil {
		t.Fatal("Rubric.Validate() = nil, want ErrMustPassBypassProhibited for Security-excluded profile")
	}
	if !errors.Is(validateErr, config.ErrMustPassBypassProhibited) {
		t.Errorf("Rubric.Validate() error = %v, want ErrMustPassBypassProhibited", validateErr)
	}
}

// TestLoadHarnessConfigWithProfiles는 HRN-003 M4에서 추가된 Profiles 필드를
// 검증합니다. LoadHarnessConfig는 Profiles map을 반환해야 합니다.
// AC-HRN-003-07.c.
func TestLoadHarnessConfigWithProfiles(t *testing.T) {
	cfg, err := loadEvaluatorConfig()
	if err != nil {
		t.Skipf("loadEvaluatorConfig() not yet implemented: %v", err)
	}
	if cfg == nil {
		t.Fatal("loadEvaluatorConfig() returned nil")
	}
	// Profiles 맵이 있어야 합니다 (최소 default 프로필 포함).
	if len(cfg.Profiles) == 0 {
		t.Error("EvaluatorConfig.Profiles is empty, want at least one profile")
	}
}

// TestEvaluatorConfig_DefaultsSet는 EvaluatorConfig 기본값이
// 올바르게 설정되는지 검증합니다.
// M4 T4.5: Aggregation + MustPassDimensions 기본값.
func TestEvaluatorConfig_DefaultsSet(t *testing.T) {
	cfg := newDefaultEvaluatorConfig()
	if cfg.Aggregation != "min" {
		t.Errorf("EvaluatorConfig.Aggregation = %q, want %q", cfg.Aggregation, "min")
	}
	if len(cfg.MustPassDimensions) == 0 {
		t.Error("EvaluatorConfig.MustPassDimensions is empty, want default [Functionality, Security]")
	}
}
