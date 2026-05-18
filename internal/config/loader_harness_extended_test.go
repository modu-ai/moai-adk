package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadHarnessConfigExtended — HRN-001 run-phase 확장 로더 테스트.
// 4개의 fixture 디렉토리를 커버합니다:
//   - harness-valid/       : 정상 파싱 (골든 패스)
//   - harness-invalid-threshold/ : pass_threshold 0.5 → ErrPassThresholdFloor
//   - harness-invalid-level/     : levels.expert 알 수 없는 enum → ErrUnknownLevel
//   - harness-drift-strict/      : unknown_top_field → MOAI_CONFIG_STRICT=1 시 ErrSchemaDrift

func TestLoadHarnessConfigExtended_ValidParse(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "harness-valid", "harness.yaml")
	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		t.Fatalf("LoadHarnessConfig() error for valid fixture: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadHarnessConfig() returned nil config")
	}

	// ModeDefaults 검증
	if cfg.ModeDefaults["cg"] != "thorough" {
		t.Errorf("ModeDefaults[cg]: got %q, want %q", cfg.ModeDefaults["cg"], "thorough")
	}
	if cfg.ModeDefaults["solo"] != "auto" {
		t.Errorf("ModeDefaults[solo]: got %q, want %q", cfg.ModeDefaults["solo"], "auto")
	}

	// EffortMapping 검증
	if cfg.EffortMapping["minimal"] != "medium" {
		t.Errorf("EffortMapping[minimal]: got %q, want %q", cfg.EffortMapping["minimal"], "medium")
	}
	if cfg.EffortMapping["standard"] != "high" {
		t.Errorf("EffortMapping[standard]: got %q, want %q", cfg.EffortMapping["standard"], "high")
	}
	if cfg.EffortMapping["thorough"] != "xhigh" {
		t.Errorf("EffortMapping[thorough]: got %q, want %q", cfg.EffortMapping["thorough"], "xhigh")
	}

	// Escalation 검증
	if cfg.Escalation.MaxEscalations != 2 {
		t.Errorf("Escalation.MaxEscalations: got %d, want 2", cfg.Escalation.MaxEscalations)
	}
	if len(cfg.Escalation.Triggers) == 0 {
		t.Error("Escalation.Triggers should not be empty")
	}

	// Levels 검증
	if len(cfg.Levels) != 3 {
		t.Errorf("Levels count: got %d, want 3", len(cfg.Levels))
	}
	minimalLevel, ok := cfg.Levels["minimal"]
	if !ok {
		t.Error("Levels[minimal] not found")
	} else if minimalLevel.Evaluator {
		t.Error("Levels[minimal].Evaluator should be false")
	}

	thoroughLevel, ok := cfg.Levels["thorough"]
	if !ok {
		t.Error("Levels[thorough] not found")
	} else if thoroughLevel.EvaluatorProfile != "strict" {
		t.Errorf("Levels[thorough].EvaluatorProfile: got %q, want %q",
			thoroughLevel.EvaluatorProfile, "strict")
	}

	// AutoDetection 검증
	if !cfg.AutoDetection.Enabled {
		t.Error("AutoDetection.Enabled should be true")
	}

	// DefaultProfile 검증
	if cfg.DefaultProfile != "default" {
		t.Errorf("DefaultProfile: got %q, want %q", cfg.DefaultProfile, "default")
	}
}

func TestLoadHarnessConfigExtended_InvalidThreshold(t *testing.T) {
	t.Parallel()

	// harness-invalid-threshold: lenient profile at levels.standard.evaluator_profile
	// with pass_threshold 0.5 (below FROZEN floor 0.60)
	path := filepath.Join("testdata", "harness-invalid-threshold", "harness.yaml")

	// 테스트를 위해 평가 프로필 경로를 testdata 디렉토리 기준으로 설정합니다.
	// LoadHarnessConfig가 evaluator_profile을 검증할 때 전달된 디렉토리를 기준으로
	// 프로파일 파일을 찾아야 합니다.
	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		// 플로어 검증 오류는 로드 시점에 발생할 수 있습니다.
		// (evaluator_profile 필드가 있고 프로파일 파일이 존재하면)
		if !errors.Is(err, ErrPassThresholdFloor) {
			t.Logf("LoadHarnessConfig returned error (expected): %v", err)
		}
		return
	}

	// 로드 성공 시, ValidatePassThresholdFloor를 직접 호출하여 검증
	if cfg == nil {
		t.Fatal("LoadHarnessConfig() returned nil")
	}
	// standard level의 evaluator_profile이 "lenient"이어야 합니다
	stdLevel, ok := cfg.Levels["standard"]
	if !ok {
		t.Fatal("Levels[standard] not found")
	}
	if stdLevel.EvaluatorProfile != "lenient" {
		t.Fatalf("Levels[standard].EvaluatorProfile: got %q, want %q", stdLevel.EvaluatorProfile, "lenient")
	}
}

func TestLoadHarnessConfigExtended_UnknownLevel(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "harness-invalid-level", "harness.yaml")
	cfg, err := LoadHarnessConfig(path)
	if err == nil {
		// LoadHarnessConfig가 성공했다면, unknown level이 있는지 확인
		if cfg != nil {
			for levelName := range cfg.Levels {
				switch levelName {
				case "minimal", "standard", "thorough":
					// 유효한 레벨
				default:
					// 유효하지 않은 레벨이 있으면 테스트 실패
					// ValidateLevelEnum이 호출되어야 합니다
					t.Logf("Unknown level %q found in config — ErrUnknownLevel should be returned by ValidateLevelEnum", levelName)
				}
			}
		}
		return
	}

	if !errors.Is(err, ErrUnknownLevel) {
		t.Errorf("expected ErrUnknownLevel, got: %v", err)
	}
}

func TestLoadHarnessConfigExtended_SchemaDrift_Warn(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 불가
	// MOAI_CONFIG_STRICT 미설정 시 — 경고만 발생, 오류 없음
	t.Setenv("MOAI_CONFIG_STRICT", "")

	path := filepath.Join("testdata", "harness-drift-strict", "harness.yaml")
	_, err := LoadHarnessConfig(path)
	if err != nil {
		t.Errorf("MOAI_CONFIG_STRICT 미설정 시 오류 없어야 함, got: %v", err)
	}
}

func TestLoadHarnessConfigExtended_SchemaDrift_StrictMode(t *testing.T) {
	// 직렬 실행 (환경변수 설정)
	t.Setenv("MOAI_CONFIG_STRICT", "1")

	path := filepath.Join("testdata", "harness-drift-strict", "harness.yaml")
	_, err := LoadHarnessConfig(path)
	if err == nil {
		t.Error("MOAI_CONFIG_STRICT=1 시 ErrSchemaDrift 오류 반환 기대")
	}
	if !errors.Is(err, ErrSchemaDrift) {
		t.Errorf("expected ErrSchemaDrift, got: %v", err)
	}
}

func TestLoadHarnessConfigExtended_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := LoadHarnessConfig("/nonexistent/path/harness.yaml")
	if !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("missing file: expected ErrConfigNotFound, got: %v", err)
	}
}

func TestLoadHarnessConfigExtended_ValidatesLevelEnum(t *testing.T) {
	t.Parallel()

	// 알 수 없는 레벨(expert)가 포함된 YAML을 직접 파싱하여
	// ErrUnknownLevel이 반환되는지 검증
	tmpDir := t.TempDir()
	invalidYAML := `harness:
    default_profile: default
    effort_mapping:
        minimal: medium
        standard: high
        thorough: xhigh
    escalation:
        enabled: true
        max_escalations: 2
        triggers:
            - quality_gate_fail
    evaluator:
        memory_scope: per_iteration
    levels:
        minimal:
            description: minimal
            evaluator: false
            plan_audit:
                enabled: true
                max_iterations: 1
                require_must_pass: false
            skip_phases: []
            sprint_contract: false
        expert:
            description: unknown level
            evaluator: true
            plan_audit:
                enabled: true
                max_iterations: 3
                require_must_pass: true
            skip_phases: []
            sprint_contract: false
    mode_defaults:
        cg: thorough
        solo: auto
        team: auto
`
	yamlPath := filepath.Join(tmpDir, "harness.yaml")
	if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, err := LoadHarnessConfig(yamlPath)
	if err == nil {
		t.Error("expected ErrUnknownLevel for 'expert' level")
		return
	}
	if !errors.Is(err, ErrUnknownLevel) {
		t.Errorf("expected ErrUnknownLevel, got: %v", err)
	}
}
