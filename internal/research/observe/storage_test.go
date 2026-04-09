package observe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 단일 관찰 추가 후 파일에 1줄 존재 확인
func TestStorage_Append_Single(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	obs := &Observation{
		Type:      ObsCorrection,
		Agent:     "expert-backend",
		Target:    "error-handling",
		Detail:    "에러 래핑 누락",
		Timestamp: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}

	if err := s.Append(obs); err != nil {
		t.Fatalf("Append 실패: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "observations.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Errorf("줄 수 = %d, want 1", len(lines))
	}
}

// 여러 관찰 추가 후 파일에 N줄 존재 확인
func TestStorage_Append_Multiple(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	for i := 0; i < 5; i++ {
		obs := &Observation{
			Type:      ObsSuccess,
			Agent:     "expert-testing",
			Target:    "coverage",
			Detail:    "테스트 통과",
			Timestamp: time.Date(2026, 4, 9, 12, i, 0, 0, time.UTC),
		}
		if err := s.Append(obs); err != nil {
			t.Fatalf("Append[%d] 실패: %v", i, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, "observations.jsonl"))
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 5 {
		t.Errorf("줄 수 = %d, want 5", len(lines))
	}
}

// LoadAll이 모든 관찰을 순서대로 반환하는지 확인
func TestStorage_LoadAll(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	agents := []string{"agent-a", "agent-b", "agent-c"}
	for i, agent := range agents {
		obs := &Observation{
			Type:      ObsCorrection,
			Agent:     agent,
			Target:    "target",
			Detail:    "상세 정보",
			Timestamp: time.Date(2026, 4, 9, 12, i, 0, 0, time.UTC),
		}
		if err := s.Append(obs); err != nil {
			t.Fatalf("Append 실패: %v", err)
		}
	}

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll 실패: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("len = %d, want 3", len(all))
	}
	for i, agent := range agents {
		if all[i].Agent != agent {
			t.Errorf("all[%d].Agent = %q, want %q", i, all[i].Agent, agent)
		}
	}
}

// LoadSince가 지정 시간 이후 관찰만 필터링하는지 확인
func TestStorage_LoadSince(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		obs := &Observation{
			Type:      ObsSuccess,
			Agent:     "agent",
			Target:    "target",
			Detail:    "상세",
			Timestamp: base.Add(time.Duration(i) * time.Hour),
		}
		if err := s.Append(obs); err != nil {
			t.Fatalf("Append 실패: %v", err)
		}
	}

	// base+2h 이후 → 인덱스 2,3,4 (3개)
	since := base.Add(2 * time.Hour)
	filtered, err := s.LoadSince(since)
	if err != nil {
		t.Fatalf("LoadSince 실패: %v", err)
	}

	if len(filtered) != 3 {
		t.Errorf("len = %d, want 3", len(filtered))
	}
}

// 존재하지 않는 파일에서 LoadAll → 빈 슬라이스, 에러 없음
func TestStorage_LoadAll_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll 에러 발생: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("len = %d, want 0", len(all))
	}
}

// 빈 파일에서 LoadAll → 빈 슬라이스, 에러 없음
func TestStorage_LoadAll_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	// 빈 파일 생성
	if err := os.WriteFile(filepath.Join(dir, "observations.jsonl"), []byte(""), 0o644); err != nil {
		t.Fatalf("빈 파일 생성 실패: %v", err)
	}

	s := NewStorage(dir)
	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll 에러 발생: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("len = %d, want 0", len(all))
	}
}

// 손상된 줄은 건너뛰고 나머지를 정상 로드
func TestStorage_LoadAll_CorruptedLine(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	// 정상 관찰 1개 추가
	obs := &Observation{
		Type:      ObsCorrection,
		Agent:     "agent",
		Target:    "target",
		Detail:    "상세",
		Timestamp: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
	if err := s.Append(obs); err != nil {
		t.Fatalf("Append 실패: %v", err)
	}

	// 손상된 줄 추가
	filePath := filepath.Join(dir, "observations.jsonl")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("파일 열기 실패: %v", err)
	}
	if _, err := f.WriteString("이것은 유효하지 않은 JSON입니다\n"); err != nil {
		_ = f.Close()
		t.Fatalf("쓰기 실패: %v", err)
	}
	_ = f.Close()

	// 정상 관찰 1개 더 추가
	obs2 := &Observation{
		Type:      ObsSuccess,
		Agent:     "agent-2",
		Target:    "target-2",
		Detail:    "상세 2",
		Timestamp: time.Date(2026, 4, 9, 13, 0, 0, 0, time.UTC),
	}
	if err := s.Append(obs2); err != nil {
		t.Fatalf("Append 실패: %v", err)
	}

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll 에러 발생: %v", err)
	}

	// 손상된 줄은 건너뛰고 2개만 반환
	if len(all) != 2 {
		t.Errorf("len = %d, want 2", len(all))
	}
}
