package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ErrRateLimitExceeded는 속도 제한 초과 시 반환되는 에러이다.
var ErrRateLimitExceeded = fmt.Errorf("research: rate limit exceeded")

// RateLimiter는 실험 빈도 제한을 적용한다.
type RateLimiter struct {
	storePath string // 액션 기록을 저장하는 JSONL 파일 경로
}

// NewRateLimiter는 주어진 경로에 기록을 저장하는 리미터를 생성한다.
func NewRateLimiter(storePath string) *RateLimiter {
	return &RateLimiter{
		storePath: storePath,
	}
}

// CheckSessionLimit은 세션 실험 횟수가 제한 내에 있는지 확인한다.
func (l *RateLimiter) CheckSessionLimit(config RateLimitConfig, sessionActions int) error {
	if sessionActions >= config.MaxExperimentsPerSession {
		return fmt.Errorf("research/safety: 세션 실험 제한 초과 (%d/%d): %w",
			sessionActions, config.MaxExperimentsPerSession, ErrRateLimitExceeded)
	}
	return nil
}

// CheckWeeklyLimit은 액션 로그를 읽고 주간 제한을 확인한다.
func (l *RateLimiter) CheckWeeklyLimit(config RateLimitConfig) error {
	records, err := l.loadRecords()
	if err != nil {
		return fmt.Errorf("research/safety: 액션 로그 읽기 실패: %w", err)
	}

	// 최근 7일 이내의 auto_research 기록만 카운트
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	count := 0
	for _, r := range records {
		if r.Type == "auto_research" && r.Timestamp.After(weekAgo) {
			count++
		}
	}

	if count >= config.MaxAutoResearchPerWeek {
		return fmt.Errorf("research/safety: 주간 자동 리서치 제한 초과 (%d/%d): %w",
			count, config.MaxAutoResearchPerWeek, ErrRateLimitExceeded)
	}
	return nil
}

// RecordAction은 JSONL 로그에 액션을 추가한다.
func (l *RateLimiter) RecordAction(actionType string) error {
	f, err := os.OpenFile(l.storePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("research/safety: 액션 로그 파일 열기 실패: %w", err)
	}
	defer f.Close()

	record := ActionRecord{
		Type:      actionType,
		Timestamp: time.Now(),
	}

	if err := json.NewEncoder(f).Encode(record); err != nil {
		return fmt.Errorf("research/safety: 액션 기록 쓰기 실패: %w", err)
	}
	return nil
}

// loadRecords는 JSONL 파일에서 모든 액션 기록을 읽는다.
func (l *RateLimiter) loadRecords() ([]ActionRecord, error) {
	f, err := os.Open(l.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 파일이 없으면 빈 목록 반환
		}
		return nil, fmt.Errorf("research/safety: 액션 로그 파일 열기 실패: %w", err)
	}
	defer f.Close()

	var records []ActionRecord
	dec := json.NewDecoder(f)
	for dec.More() {
		var r ActionRecord
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("research/safety: 액션 기록 디코딩 실패: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}
