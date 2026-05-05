package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ErrRateLimitExceeded is the error returned when a rate limit is exceeded.
var ErrRateLimitExceeded = fmt.Errorf("research: rate limit exceeded")

// RateLimiter enforces experiment frequency limits.
type RateLimiter struct {
	storePath string // Path to the JSONL file storing action records
}

// NewRateLimiter creates a limiter that stores records at the given path.
// @MX:ANCHOR: [AUTO] NewRateLimiter is a factory called from 3+ packages
// @MX:REASON: [AUTO] Multi-package dependency — used by research, CLI, and auto-update
func NewRateLimiter(storePath string) *RateLimiter {
	return &RateLimiter{
		storePath: storePath,
	}
}

// CheckSessionLimit checks whether the session experiment count is within the limit.
func (l *RateLimiter) CheckSessionLimit(config RateLimitConfig, sessionActions int) error {
	if sessionActions >= config.MaxExperimentsPerSession {
		return fmt.Errorf("research/safety: session experiment limit exceeded (%d/%d): %w",
			sessionActions, config.MaxExperimentsPerSession, ErrRateLimitExceeded)
	}
	return nil
}

// CheckWeeklyLimit reads the action log and checks the weekly limit.
func (l *RateLimiter) CheckWeeklyLimit(config RateLimitConfig) error {
	records, err := l.loadRecords()
	if err != nil {
		return fmt.Errorf("research/safety: failed to read action log: %w", err)
	}

	// Count only auto_research records within the last 7 days
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	count := 0
	for _, r := range records {
		if r.Type == "auto_research" && r.Timestamp.After(weekAgo) {
			count++
		}
	}

	if count >= config.MaxAutoResearchPerWeek {
		return fmt.Errorf("research/safety: weekly auto-research limit exceeded (%d/%d): %w",
			count, config.MaxAutoResearchPerWeek, ErrRateLimitExceeded)
	}
	return nil
}

// RecordAction appends an action to the JSONL log.
func (l *RateLimiter) RecordAction(actionType string) error {
	f, err := os.OpenFile(l.storePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("research/safety: failed to open action log file: %w", err)
	}

	record := ActionRecord{
		Type:      actionType,
		Timestamp: time.Now(),
	}

	if err := json.NewEncoder(f).Encode(record); err != nil {
		_ = f.Close()
		return fmt.Errorf("research/safety: failed to write action record: %w", err)
	}
	return f.Close()
}

// loadRecords reads all action records from the JSONL file.
func (l *RateLimiter) loadRecords() ([]ActionRecord, error) {
	f, err := os.Open(l.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Return empty list if file does not exist
		}
		return nil, fmt.Errorf("research/safety: failed to open action log file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var records []ActionRecord
	dec := json.NewDecoder(f)
	for dec.More() {
		var r ActionRecord
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("research/safety: failed to decode action record: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}
