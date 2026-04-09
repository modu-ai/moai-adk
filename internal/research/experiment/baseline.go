package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// BaselineManager는 타겟별 베이스라인 평가 결과를 파일 시스템에 저장/로드한다.
type BaselineManager struct {
	storeDir string
}

// NewBaselineManager는 지정된 디렉토리에 베이스라인을 관리하는 매니저를 생성한다.
func NewBaselineManager(storeDir string) *BaselineManager {
	return &BaselineManager{storeDir: storeDir}
}

// Save는 타겟의 베이스라인 평가 결과를 JSON 파일로 저장한다.
func (m *BaselineManager) Save(target string, result *eval.EvalResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("베이스라인 직렬화 실패 (target=%s): %w", target, err)
	}

	filePath := m.filePath(target)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("베이스라인 디렉토리 생성 실패: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("베이스라인 파일 쓰기 실패 (target=%s): %w", target, err)
	}

	return nil
}

// Load는 타겟의 베이스라인 평가 결과를 파일에서 읽어온다.
func (m *BaselineManager) Load(target string) (*eval.EvalResult, error) {
	data, err := os.ReadFile(m.filePath(target))
	if err != nil {
		return nil, fmt.Errorf("베이스라인 파일 읽기 실패 (target=%s): %w", target, err)
	}

	var result eval.EvalResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("베이스라인 역직렬화 실패 (target=%s): %w", target, err)
	}

	return &result, nil
}

// Exists는 타겟의 베이스라인 파일이 존재하는지 확인한다.
func (m *BaselineManager) Exists(target string) bool {
	_, err := os.Stat(m.filePath(target))
	return err == nil
}

// filePath는 타겟에 대한 베이스라인 파일 경로를 반환한다.
func (m *BaselineManager) filePath(target string) string {
	return filepath.Join(m.storeDir, sanitizeTarget(target)+".baseline.json")
}

// sanitizeTarget은 타겟 문자열을 파일명에 안전한 형태로 변환한다.
// '/'를 '_'로 대체하고, '.'을 제거한다.
func sanitizeTarget(target string) string {
	s := strings.ReplaceAll(target, "/", "_")
	s = strings.ReplaceAll(s, ".", "")
	return s
}
