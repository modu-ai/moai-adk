package eval

import "fmt"

// EvalEngine은 평가 시스템의 메인 인터페이스이다.
type EvalEngine interface {
	// LoadSuite는 YAML 파일에서 평가 스위트를 로드한다.
	LoadSuite(path string) (*EvalSuite, error)

	// Evaluate는 스위트와 결과 맵을 받아 집계된 평가 결과를 반환한다.
	Evaluate(suite *EvalSuite, results map[string]bool) (*EvalResult, error)
}

// engine은 기본 EvalEngine 구현체이다.
type engine struct{}

// NewEngine은 새로운 EvalEngine을 생성한다.
func NewEngine() EvalEngine {
	return &engine{}
}

// LoadSuite는 YAML 파일에서 평가 스위트를 로드한다.
func (e *engine) LoadSuite(path string) (*EvalSuite, error) {
	return LoadSuite(path)
}

// Evaluate는 스위트와 결과 맵을 받아 집계된 평가 결과를 반환한다.
// suite가 nil이면 에러를 반환한다.
func (e *engine) Evaluate(suite *EvalSuite, results map[string]bool) (*EvalResult, error) {
	if suite == nil {
		return nil, fmt.Errorf("eval suite가 nil입니다")
	}

	return ComputeResult(suite.Criteria, results), nil
}
