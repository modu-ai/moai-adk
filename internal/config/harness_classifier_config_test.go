// Package config — harness_classifier_config_test.go
// Wave D T-D1: harness.yaml learning.classifier 블록 로딩 검증 테스트.
// REQ-HRN-CLS-016: learning.classifier 설정을 harness.yaml에서 읽어 ClassifierConfig로 반환한다.
// REQ-HRN-CLS-018: yaml.TypeError 포함 파싱 오류 시 Stage-1 fallback (WithDefaults 값 반환).
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// makeHarnessYAMLFromString은 주어진 content로 임시 harness.yaml 파일을 생성하고 경로를 반환한다.
func makeHarnessYAMLFromString(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "harness.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("harness.yaml 생성 실패: %v", err)
	}
	return path
}

// TestLoadHarnessConfig_LearningClassifierBlock은 harness.yaml에 learning.classifier 블록이 있을 때
// ClassifierConfig 필드가 올바르게 파싱되는지 검증한다 (REQ-HRN-CLS-016).
func TestLoadHarnessConfig_LearningClassifierBlock(t *testing.T) {
	t.Parallel()

	content := `harness:
  evaluator:
    memory_scope: per_iteration
  learning:
    classifier:
      stage_2_enabled: true
      similarity_algorithm: simhash
      hamming_threshold: 5
      cluster_min_size: 4
`
	path := makeHarnessYAMLFromString(t, content)

	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		t.Fatalf("LoadHarnessConfig 오류: %v", err)
	}

	got := cfg.Learning.Classifier
	if !got.Stage2Enabled {
		t.Error("Stage2Enabled = false, want true")
	}
	if got.SimilarityAlgorithm != "simhash" {
		t.Errorf("SimilarityAlgorithm = %q, want simhash", got.SimilarityAlgorithm)
	}
	if got.HammingThreshold != 5 {
		t.Errorf("HammingThreshold = %d, want 5", got.HammingThreshold)
	}
	if got.ClusterMinSize != 4 {
		t.Errorf("ClusterMinSize = %d, want 4", got.ClusterMinSize)
	}
}

// TestLoadHarnessConfig_MissingLearningBlock은 learning 블록 없을 때
// ClassifierConfig가 제로값(Stage-1 fallback)으로 반환되는지 검증한다.
func TestLoadHarnessConfig_MissingLearningBlock(t *testing.T) {
	t.Parallel()

	content := `harness:
  evaluator:
    memory_scope: per_iteration
`
	path := makeHarnessYAMLFromString(t, content)

	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		t.Fatalf("LoadHarnessConfig 오류: %v", err)
	}

	// 제로값 = ClassifierConfig{} = Stage2Enabled false
	got := cfg.Learning.Classifier
	if got.Stage2Enabled {
		t.Error("Stage2Enabled = true (제로값이어야 함)")
	}
	if got.SimilarityAlgorithm != "" {
		t.Errorf("SimilarityAlgorithm = %q (제로값 \"\"이어야 함)", got.SimilarityAlgorithm)
	}
}

// TestLoadHarnessConfig_TypeMismatchFallsBackToDefaults는 learning.classifier 필드 타입 오류 시
// (yaml.TypeError) Stage-1 fallback으로 WithDefaults() 값을 반환하는지 검증한다 (REQ-HRN-CLS-018).
func TestLoadHarnessConfig_TypeMismatchFallsBackToDefaults(t *testing.T) {
	t.Parallel()

	// hamming_threshold에 문자열 대입 → yaml.TypeError
	content := `harness:
  evaluator:
    memory_scope: per_iteration
  learning:
    classifier:
      stage_2_enabled: true
      hamming_threshold: "not-a-number"
`
	path := makeHarnessYAMLFromString(t, content)

	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		t.Fatalf("LoadHarnessConfig 오류: %v (yaml.TypeError는 fallback이어야 함)", err)
	}

	// TypeError fallback → WithDefaults() 값 (simhash / 3 / 3)
	got := cfg.Learning.Classifier
	const wantAlgo = "simhash"
	const wantHamming = 3
	const wantCluster = 3
	if got.SimilarityAlgorithm != wantAlgo {
		t.Errorf("fallback SimilarityAlgorithm = %q, want %q", got.SimilarityAlgorithm, wantAlgo)
	}
	if got.HammingThreshold != wantHamming {
		t.Errorf("fallback HammingThreshold = %d, want %d", got.HammingThreshold, wantHamming)
	}
	if got.ClusterMinSize != wantCluster {
		t.Errorf("fallback ClusterMinSize = %d, want %d", got.ClusterMinSize, wantCluster)
	}
}

// TestLoadHarnessConfig_Stage2DisabledByDefault는 stage_2_enabled 미설정 시
// Stage2Enabled가 false(제로값)인지 검증한다 (backward compatibility).
func TestLoadHarnessConfig_Stage2DisabledByDefault(t *testing.T) {
	t.Parallel()

	content := `harness:
  evaluator:
    memory_scope: per_iteration
  learning:
    classifier:
      similarity_algorithm: simhash
      hamming_threshold: 3
      cluster_min_size: 3
`
	path := makeHarnessYAMLFromString(t, content)

	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		t.Fatalf("LoadHarnessConfig 오류: %v", err)
	}

	if cfg.Learning.Classifier.Stage2Enabled {
		t.Error("Stage2Enabled = true (명시적 미설정 시 false여야 함)")
	}
}
