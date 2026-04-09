package safety

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRateLimiter_CheckSessionLimit은 세션 실험 횟수 제한을 검증한다.
func TestRateLimiter_CheckSessionLimit(t *testing.T) {
	tests := []struct {
		name           string
		config         RateLimitConfig
		sessionActions int
		wantErr        bool
	}{
		{
			name:           "세션 제한 미초과 → nil",
			config:         RateLimitConfig{MaxExperimentsPerSession: 10},
			sessionActions: 5,
			wantErr:        false,
		},
		{
			name:           "세션 제한 정확히 도달 → 에러",
			config:         RateLimitConfig{MaxExperimentsPerSession: 10},
			sessionActions: 10,
			wantErr:        true,
		},
		{
			name:           "세션 제한 초과 → 에러",
			config:         RateLimitConfig{MaxExperimentsPerSession: 5},
			sessionActions: 8,
			wantErr:        true,
		},
		{
			name:           "세션 액션 0 → nil",
			config:         RateLimitConfig{MaxExperimentsPerSession: 10},
			sessionActions: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			l := NewRateLimiter(filepath.Join(tmpDir, "actions.jsonl"))
			err := l.CheckSessionLimit(tt.config, tt.sessionActions)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSessionLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrRateLimitExceeded) {
				t.Errorf("CheckSessionLimit() error = %v, want ErrRateLimitExceeded", err)
			}
		})
	}
}

// TestRateLimiter_CheckWeeklyLimit은 주간 자동 리서치 제한을 검증한다.
func TestRateLimiter_CheckWeeklyLimit(t *testing.T) {
	t.Run("기록 없음 → nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		l := NewRateLimiter(filepath.Join(tmpDir, "actions.jsonl"))
		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}

		err := l.CheckWeeklyLimit(cfg)
		if err != nil {
			t.Errorf("CheckWeeklyLimit() error = %v, want nil", err)
		}
	})

	t.Run("주간 제한 미초과 → nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		// 최근 기록 2개 작성
		writeActionRecords(t, storePath, []ActionRecord{
			{Type: "auto_research", Timestamp: time.Now().Add(-1 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-2 * time.Hour)},
		})

		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}
		err := l.CheckWeeklyLimit(cfg)
		if err != nil {
			t.Errorf("CheckWeeklyLimit() error = %v, want nil", err)
		}
	})

	t.Run("주간 제한 초과 → ErrRateLimitExceeded", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		// 최근 기록 5개 작성 (제한: 5)
		records := make([]ActionRecord, 5)
		for i := range records {
			records[i] = ActionRecord{
				Type:      "auto_research",
				Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			}
		}
		writeActionRecords(t, storePath, records)

		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}
		err := l.CheckWeeklyLimit(cfg)
		if err == nil {
			t.Error("CheckWeeklyLimit() error = nil, want ErrRateLimitExceeded")
		}
		if !errors.Is(err, ErrRateLimitExceeded) {
			t.Errorf("CheckWeeklyLimit() error = %v, want ErrRateLimitExceeded", err)
		}
	})

	t.Run("7일 이전 기록은 주간 제한에 포함되지 않음", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		// 오래된 기록 10개 + 최근 기록 2개
		records := []ActionRecord{
			{Type: "auto_research", Timestamp: time.Now().Add(-10 * 24 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-9 * 24 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-8 * 24 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-1 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-2 * time.Hour)},
		}
		writeActionRecords(t, storePath, records)

		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}
		err := l.CheckWeeklyLimit(cfg)
		if err != nil {
			t.Errorf("CheckWeeklyLimit() error = %v, want nil (오래된 기록은 제외해야 함)", err)
		}
	})
}

// TestRateLimiter_RecordAction은 JSONL 기록 저장을 검증한다.
func TestRateLimiter_RecordAction(t *testing.T) {
	t.Run("액션 기록이 JSONL 파일에 저장됨", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		err := l.RecordAction("experiment")
		if err != nil {
			t.Fatalf("RecordAction() error = %v", err)
		}

		err = l.RecordAction("auto_research")
		if err != nil {
			t.Fatalf("RecordAction() error = %v", err)
		}

		// JSONL 파일 읽기 및 검증
		records := readActionRecords(t, storePath)
		if len(records) != 2 {
			t.Fatalf("기록 수 = %d, want 2", len(records))
		}
		if records[0].Type != "experiment" {
			t.Errorf("records[0].Type = %q, want %q", records[0].Type, "experiment")
		}
		if records[1].Type != "auto_research" {
			t.Errorf("records[1].Type = %q, want %q", records[1].Type, "auto_research")
		}
	})

	t.Run("존재하지 않는 디렉토리에 기록 시 에러", func(t *testing.T) {
		l := NewRateLimiter("/nonexistent/dir/actions.jsonl")
		err := l.RecordAction("experiment")
		if err == nil {
			t.Error("RecordAction() error = nil, want error for nonexistent directory")
		}
	})
}

// writeActionRecords는 테스트용 JSONL 파일에 액션 기록을 작성한다.
func writeActionRecords(t *testing.T, path string, records []ActionRecord) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("기록 인코딩 실패: %v", err)
		}
	}
}

// readActionRecords는 테스트용 JSONL 파일에서 액션 기록을 읽는다.
func readActionRecords(t *testing.T, path string) []ActionRecord {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("파일 열기 실패: %v", err)
	}
	defer f.Close()

	var records []ActionRecord
	dec := json.NewDecoder(f)
	for dec.More() {
		var r ActionRecord
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("기록 디코딩 실패: %v", err)
		}
		records = append(records, r)
	}
	return records
}
