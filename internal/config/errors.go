// Package config provides configuration management for MoAI-ADK Go Edition.
// It loads YAML section files, applies defaults, validates, and provides
// thread-safe access to configuration values.
package config

// @MX:NOTE: [AUTO] Sentinel errors for configuration operations including validation errors with field context

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for configuration operations.
var (
	// ErrConfigNotFound indicates the configuration directory was not found.
	ErrConfigNotFound = errors.New("config: configuration directory not found")

	// ErrInvalidConfig indicates the configuration is invalid.
	ErrInvalidConfig = errors.New("config: invalid configuration")

	// ErrSectionNotFound indicates the requested section does not exist.
	ErrSectionNotFound = errors.New("config: section not found")

	// ErrInvalidDevelopmentMode indicates an invalid development mode value.
	ErrInvalidDevelopmentMode = errors.New("config: invalid development_mode, must be one of: ddd, tdd")

	// ErrNotInitialized indicates the ConfigManager has not been initialized via Load().
	ErrNotInitialized = errors.New("config: manager not initialized, call Load() first")

	// ErrSectionTypeMismatch indicates the section value type does not match expected type.
	ErrSectionTypeMismatch = errors.New("config: section type mismatch")

	// ErrDynamicToken indicates an unexpanded dynamic token was detected in a config value.
	ErrDynamicToken = errors.New("config: unexpanded dynamic token detected")

	// ErrInvalidYAML indicates invalid YAML syntax in a configuration file.
	ErrInvalidYAML = errors.New("config: invalid YAML syntax")

	// ErrEvalMemoryFrozen은 evaluator.memory_scope가 per_iteration이 아닌 값으로
	// 설정된 경우 로더 검증에서 반환됩니다.
	// HRN-002 run-phase minimal substrate — design-constitution §11.4.1에 의거
	// per_iteration은 FROZEN 값이며 다른 값은 허용되지 않습니다.
	ErrEvalMemoryFrozen = errors.New("HRN_EVAL_MEMORY_FROZEN: evaluator.memory_scope must be 'per_iteration' (FROZEN per design-constitution §11.4.1)")

	// HRN-003 run-phase: 4-dimension × sub-criteria hierarchical scoring sentinels.

	// ErrUnknownDimension은 프로필이 canonical 집합
	// {Functionality, Security, Craft, Consistency} 외의 차원을 선언할 때 반환됩니다.
	// REQ-HRN-003-019, AC-HRN-003-08.
	ErrUnknownDimension = errors.New("HRN_UNKNOWN_DIMENSION: profile declares dimension outside canonical set {Functionality, Security, Craft, Consistency}")

	// ErrRubricCitationMissing은 sub-criterion 점수에 rubric_anchor 필드가 없거나
	// canonical anchor (0.25/0.50/0.75/1.00) 외의 값을 가질 때 반환됩니다.
	// REQ-HRN-003-009, AC-HRN-003-05, design-constitution §12 Mechanism 1.
	ErrRubricCitationMissing = errors.New("HRN_RUBRIC_CITATION_MISSING: sub-criterion score missing rubric_anchor field or non-canonical anchor (per design-constitution §12 Mechanism 1)")

	// ErrFlatScoreCardProhibited은 GAN loop runner가 비계층(flat) ScoreCard를 반환하려 할 때
	// CI 통합 테스트에서 발생합니다.
	// REQ-HRN-003-017, AC-HRN-003-02.
	ErrFlatScoreCardProhibited = errors.New("HRN_FLAT_SCORECARD_PROHIBITED: ScoreCard must be hierarchical; flat shape rejected per SPEC-V3R2-HRN-003 REQ-017")

	// ErrMustPassBypassProhibited은 프로필이 design-constitution §12 Mechanism 3에서
	// must-pass로 지정된 차원 (FROZEN Security + Functionality)에 대해 must_pass를 비활성화하려 할 때
	// 로더 검증에서 반환됩니다.
	// REQ-HRN-003-018, AC-HRN-003-11.
	ErrMustPassBypassProhibited = errors.New("HRN_MUSTPASS_BYPASS_PROHIBITED: profile attempts to narrow must-pass set below floor [Security] (per design-constitution §12 Mechanism 3)")

	// HRN-001 run-phase: harness routing + loader sentinels.

	// ErrUnknownLevel은 harness.yaml의 levels 맵에 FROZEN enum
	// {minimal, standard, thorough} 외의 레벨 이름이 포함될 때 반환됩니다.
	// REQ-HRN-001-017, AC-HRN-001-07.
	ErrUnknownLevel = errors.New("HRN_UNKNOWN_LEVEL: harness.yaml contains unknown level outside {minimal, standard, thorough} (FROZEN per SPEC-V3R2-HRN-001 REQ-017)")

	// ErrPassThresholdFloor은 evaluator 프로필의 pass_threshold가 FROZEN floor 0.60 미만일 때
	// LoadHarnessConfig() 검증에서 반환됩니다.
	// REQ-HRN-001-012, AC-HRN-001-05.
	ErrPassThresholdFloor = errors.New("HRN_PASS_THRESHOLD_FLOOR: evaluator profile pass_threshold is below FROZEN floor 0.60 (design-constitution §5, SPEC-V3R2-CON-001)")

	// ErrSchemaDrift은 harness.yaml에 HarnessConfig 구조체에 없는 알 수 없는 키가 있을 때
	// MOAI_CONFIG_STRICT=1 환경에서 반환됩니다.
	// REQ-HRN-001-019, AC-HRN-001-07.
	ErrSchemaDrift = errors.New("HRN_SCHEMA_DRIFT: harness.yaml contains unknown keys not present in HarnessConfig struct (set MOAI_CONFIG_STRICT=1 to make this an error)")

	// ErrEscalationCapExceeded는 EscalationManager가 MaxEscalations 한도에 도달한 후
	// 추가 에스컬레이션 시도 시 반환됩니다.
	// REQ-HRN-001-013, REQ-HRN-001-018, AC-HRN-001-08.
	ErrEscalationCapExceeded = errors.New("HRN_ESCALATION_CAP_REACHED: maximum escalation count exceeded (per SPEC-V3R2-HRN-001 REQ-013/018)")
)

// ValidationError represents a single validation error with field context.
type ValidationError struct {
	Field   string
	Message string
	Value   any
	Wrapped error // underlying sentinel error for errors.Is support
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation error: field %q: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation error: field %q: %s", e.Field, e.Message)
}

// Unwrap returns the underlying sentinel error.
func (e *ValidationError) Unwrap() error {
	return e.Wrapped
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors struct {
	Errors []ValidationError
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation: no errors"
	}
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("validation failed with %d error(s): %s", len(e.Errors), strings.Join(msgs, "; "))
}

// Is supports errors.Is by checking contained validation errors against the target.
func (e *ValidationErrors) Is(target error) bool {
	if target == ErrInvalidConfig {
		return true
	}
	for _, ve := range e.Errors {
		if ve.Wrapped != nil && errors.Is(ve.Wrapped, target) {
			return true
		}
	}
	return false
}
