package eval

import (
	"os"
	"path/filepath"
	"testing"
)

// validSuiteYAML 테스트용 올바른 YAML 스위트 문자열.
const validSuiteYAML = `
target:
  path: ".claude/skills/moai-lang-go/SKILL.md"
  type: skill
test_inputs:
  - name: "기본 테스트"
    prompt: "Go 코드를 작성해주세요"
evals:
  - name: "정확성"
    question: "응답이 정확한가?"
    pass: "정확한 Go 코드를 포함"
    fail: "Go 코드가 없거나 부정확"
    weight: must_pass
  - name: "가독성"
    question: "코드가 읽기 쉬운가?"
    pass: "명확한 변수명과 구조"
    fail: "난해한 코드"
    weight: nice_to_have
settings:
  runs_per_experiment: 3
  max_experiments: 10
  pass_threshold: 0.75
  target_score: 0.90
  budget_cap_tokens: 50000
`

// TestLoadSuite LoadSuite 함수에 대한 테이블 기반 테스트.
func TestLoadSuite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string // 파일에 쓸 내용 (빈 문자열이면 파일 미생성)
		setup     func(t *testing.T) string
		wantErr   bool
		checkFunc func(t *testing.T, s *EvalSuite)
	}{
		{
			name:    "유효한 YAML 파일 → 성공적 파싱",
			content: validSuiteYAML,
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "suite.yaml")
				if err := os.WriteFile(p, []byte(validSuiteYAML), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: false,
			checkFunc: func(t *testing.T, s *EvalSuite) {
				t.Helper()
				if s.Target.Path != ".claude/skills/moai-lang-go/SKILL.md" {
					t.Errorf("Target.Path = %q", s.Target.Path)
				}
				if s.Target.Type != "skill" {
					t.Errorf("Target.Type = %q", s.Target.Type)
				}
				if len(s.Inputs) != 1 {
					t.Errorf("Inputs 개수 = %d, want 1", len(s.Inputs))
				}
				if len(s.Criteria) != 2 {
					t.Errorf("Criteria 개수 = %d, want 2", len(s.Criteria))
				}
				if s.Settings.RunsPerExperiment != 3 {
					t.Errorf("RunsPerExperiment = %d, want 3", s.Settings.RunsPerExperiment)
				}
				if s.Settings.PassThreshold != 0.75 {
					t.Errorf("PassThreshold = %f, want 0.75", s.Settings.PassThreshold)
				}
				if s.Criteria[0].Weight != MustPass {
					t.Errorf("Criteria[0].Weight = %q, want %q", s.Criteria[0].Weight, MustPass)
				}
				if s.Criteria[1].Weight != NiceToHave {
					t.Errorf("Criteria[1].Weight = %q, want %q", s.Criteria[1].Weight, NiceToHave)
				}
			},
		},
		{
			name: "파일 없음 → 에러",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "nonexistent.yaml")
			},
			wantErr: true,
		},
		{
			name: "잘못된 YAML → 에러",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "bad.yaml")
				if err := os.WriteFile(p, []byte("{{invalid yaml"), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: true,
		},
		{
			name: "기준 없음 → 검증 에러",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "no_criteria.yaml")
				content := `
target:
  path: "test.md"
  type: skill
test_inputs:
  - name: "테스트"
    prompt: "질문"
evals: []
settings:
  runs_per_experiment: 1
`
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: true,
		},
		{
			name: "잘못된 weight 값 → 검증 에러",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "bad_weight.yaml")
				content := `
target:
  path: "test.md"
  type: skill
test_inputs:
  - name: "테스트"
    prompt: "질문"
evals:
  - name: "테스트 기준"
    question: "질문?"
    pass: "통과"
    fail: "실패"
    weight: invalid_weight
settings:
  runs_per_experiment: 1
`
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: true,
		},
		{
			name: "빈 test_inputs → 검증 에러",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				p := filepath.Join(dir, "no_inputs.yaml")
				content := `
target:
  path: "test.md"
  type: skill
test_inputs: []
evals:
  - name: "기준"
    question: "질문?"
    pass: "통과"
    fail: "실패"
    weight: must_pass
settings:
  runs_per_experiment: 1
`
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.setup(t)

			suite, err := LoadSuite(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadSuite() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, suite)
			}
		})
	}
}

// TestEvalSuite_Validate Validate 메서드를 직접 테스트한다.
func TestEvalSuite_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		suite   EvalSuite
		wantErr bool
	}{
		{
			name: "유효한 스위트",
			suite: EvalSuite{
				Inputs:   []TestInput{{Name: "t", Prompt: "p"}},
				Criteria: []EvalCriterion{{Name: "c", Weight: MustPass}},
			},
			wantErr: false,
		},
		{
			name: "기준 없음",
			suite: EvalSuite{
				Inputs:   []TestInput{{Name: "t", Prompt: "p"}},
				Criteria: nil,
			},
			wantErr: true,
		},
		{
			name: "입력 없음",
			suite: EvalSuite{
				Inputs:   nil,
				Criteria: []EvalCriterion{{Name: "c", Weight: MustPass}},
			},
			wantErr: true,
		},
		{
			name: "잘못된 weight",
			suite: EvalSuite{
				Inputs:   []TestInput{{Name: "t", Prompt: "p"}},
				Criteria: []EvalCriterion{{Name: "c", Weight: "bad"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.suite.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
