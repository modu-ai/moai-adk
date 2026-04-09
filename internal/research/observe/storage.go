package observe

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// observationsFile은 관찰 데이터를 저장하는 JSONL 파일명이다.
const observationsFile = "observations.jsonl"

// Storage는 관찰 레코드의 파일 기반 저장소를 관리한다.
type Storage struct {
	baseDir string
}

// NewStorage는 지정된 디렉토리에 관찰 저장소를 생성한다.
func NewStorage(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

// filePath는 관찰 파일의 전체 경로를 반환한다.
func (s *Storage) filePath() string {
	return filepath.Join(s.baseDir, observationsFile)
}

// Append는 단일 관찰을 JSONL 파일에 추가한다.
// 파일이 존재하지 않으면 새로 생성한다.
func (s *Storage) Append(obs *Observation) error {
	f, err := os.OpenFile(s.filePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("observe: 파일 열기 실패: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(obs)
	if err != nil {
		return fmt.Errorf("observe: JSON 직렬화 실패: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("observe: 파일 쓰기 실패: %w", err)
	}

	return nil
}

// LoadAll은 저장된 모든 관찰을 순서대로 반환한다.
// 파일이 존재하지 않거나 비어있으면 빈 슬라이스를 반환한다.
// 손상된 줄은 건너뛴다.
func (s *Storage) LoadAll() ([]*Observation, error) {
	f, err := os.Open(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("observe: 파일 열기 실패: %w", err)
	}
	defer f.Close()

	var result []*Observation
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var obs Observation
		if err := json.Unmarshal([]byte(line), &obs); err != nil {
			// 손상된 줄은 건너뛴다
			continue
		}
		result = append(result, &obs)
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("observe: 파일 스캔 실패: %w", err)
	}

	return result, nil
}

// LoadSince는 지정된 시간 이후의 관찰만 필터링하여 반환한다.
// since와 같거나 이후의 Timestamp를 가진 관찰이 포함된다.
func (s *Storage) LoadSince(since time.Time) ([]*Observation, error) {
	all, err := s.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("observe: LoadSince 실패: %w", err)
	}

	var filtered []*Observation
	for _, obs := range all {
		if !obs.Timestamp.Before(since) {
			filtered = append(filtered, obs)
		}
	}

	return filtered, nil
}
