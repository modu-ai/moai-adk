package eval

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EvalSuite는 YAML에서 로드하는 전체 평가 설정을 정의한다.
type EvalSuite struct {
	Target   TargetSpec      `yaml:"target"`
	Inputs   []TestInput     `yaml:"test_inputs"`
	Criteria []EvalCriterion `yaml:"evals"`
	Settings EvalSettings    `yaml:"settings"`
}

// LoadSuite는 YAML 파일에서 평가 스위트를 읽고 파싱한다.
// 파일이 존재하지 않거나, YAML이 유효하지 않거나, 검증에 실패하면 에러를 반환한다.
func LoadSuite(path string) (*EvalSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("eval suite 파일 읽기 실패: %w", err)
	}

	var suite EvalSuite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("eval suite YAML 파싱 실패: %w", err)
	}

	if err := suite.Validate(); err != nil {
		return nil, fmt.Errorf("eval suite 검증 실패: %w", err)
	}

	return &suite, nil
}

// Validate는 스위트가 올바르게 구성되었는지 검사한다.
// 최소 1개의 기준, 유효한 weight, 최소 1개의 테스트 입력이 필요하다.
func (s *EvalSuite) Validate() error {
	if len(s.Criteria) == 0 {
		return fmt.Errorf("평가 기준(evals)이 최소 1개 이상 필요합니다")
	}

	if len(s.Inputs) == 0 {
		return fmt.Errorf("테스트 입력(test_inputs)이 최소 1개 이상 필요합니다")
	}

	for i, c := range s.Criteria {
		if !c.Weight.IsValid() {
			return fmt.Errorf("기준[%d] %q의 weight %q이(가) 유효하지 않습니다 (must_pass 또는 nice_to_have만 허용)", i, c.Name, c.Weight)
		}
	}

	return nil
}
