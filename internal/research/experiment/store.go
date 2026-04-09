package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResultStore는 타겟별 실험 결과와 변경 이력을 파일 시스템에 관리한다.
type ResultStore struct {
	baseDir string
}

// NewResultStore는 지정된 디렉토리에 실험 결과를 관리하는 저장소를 생성한다.
func NewResultStore(baseDir string) *ResultStore {
	return &ResultStore{baseDir: baseDir}
}

// SaveExperiment는 실험 결과를 JSON 파일로 저장한다.
// 파일명은 exp-{NNN}.json 형식으로 0-패딩 3자리 순번을 사용한다.
func (s *ResultStore) SaveExperiment(target string, exp *Experiment) error {
	targetDir := s.targetDir(target)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("실험 디렉토리 생성 실패 (target=%s): %w", target, err)
	}

	count := s.ExperimentCount(target)
	filename := fmt.Sprintf("exp-%03d.json", count+1)

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return fmt.Errorf("실험 직렬화 실패 (id=%s): %w", exp.ID, err)
	}

	filePath := filepath.Join(targetDir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("실험 파일 쓰기 실패 (%s): %w", filePath, err)
	}

	return nil
}

// LoadExperiments는 타겟의 모든 실험 결과를 파일명 순서대로 로드한다.
// 디렉토리가 존재하지 않으면 빈 슬라이스를 반환한다.
func (s *ResultStore) LoadExperiments(target string) ([]*Experiment, error) {
	targetDir := s.targetDir(target)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Experiment{}, nil
		}
		return nil, fmt.Errorf("실험 디렉토리 읽기 실패 (target=%s): %w", target, err)
	}

	// exp-*.json 파일만 필터링하여 정렬
	var jsonFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "exp-") && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, e.Name())
		}
	}
	sort.Strings(jsonFiles)

	experiments := make([]*Experiment, 0, len(jsonFiles))
	for _, name := range jsonFiles {
		data, err := os.ReadFile(filepath.Join(targetDir, name))
		if err != nil {
			return nil, fmt.Errorf("실험 파일 읽기 실패 (%s): %w", name, err)
		}

		var exp Experiment
		if err := json.Unmarshal(data, &exp); err != nil {
			return nil, fmt.Errorf("실험 역직렬화 실패 (%s): %w", name, err)
		}
		experiments = append(experiments, &exp)
	}

	return experiments, nil
}

// AppendChangelog는 타겟의 changelog.md에 마크다운 항목을 추가한다.
func (s *ResultStore) AppendChangelog(target string, entry ChangelogEntry) error {
	targetDir := s.targetDir(target)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("changelog 디렉토리 생성 실패: %w", err)
	}

	changelogPath := filepath.Join(targetDir, "changelog.md")

	// 마크다운 항목 생성
	md := fmt.Sprintf(
		"### %s (score: %.2f, decision: %s)\n\n- **변경**: %s\n- **이유**: %s\n\n",
		entry.ExperimentID,
		entry.Score,
		entry.Decision,
		entry.Change,
		entry.Reasoning,
	)

	f, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("changelog 파일 열기 실패: %w", err)
	}

	if _, err := f.WriteString(md); err != nil {
		_ = f.Close()
		return fmt.Errorf("changelog 쓰기 실패: %w", err)
	}

	return f.Close()
}

// ExperimentCount는 타겟 디렉토리의 exp-*.json 파일 수를 반환한다.
func (s *ResultStore) ExperimentCount(target string) int {
	targetDir := s.targetDir(target)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "exp-") && strings.HasSuffix(e.Name(), ".json") {
			count++
		}
	}

	return count
}

// targetDir는 타겟에 대한 디렉토리 경로를 반환한다.
func (s *ResultStore) targetDir(target string) string {
	return filepath.Join(s.baseDir, sanitizeTarget(target))
}
