// Package harness — Pattern aggregator and tier classifier.
// REQ-HL-002: Reads usage-log.jsonl events, aggregates into patterns, and classifies tiers.
package harness

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// confidenceThreshold is the minimum confidence that allows tier promotion.
// REQ-HL-002: If below 0.70, forces TierObservation regardless of count.
const confidenceThreshold = 0.70

// Learner reads usage-log.jsonl, aggregates patterns, and classifies tiers.
// WritePromotion records promotion events to tier-promotions.jsonl.
//
// @MX:ANCHOR: [AUTO] AggregatePatterns, ClassifyTier, WritePromotion are learning pipeline entry points.
// @MX:REASON: [AUTO] fan_in >= 3: learner_test.go, safety.go(Phase 3), applier.go
type Learner struct {
	// promotionPath is the tier-promotions.jsonl file path.
	promotionPath string

	// nowFn is a function that returns current time (overridable in tests).
	nowFn func() time.Time
}

// NewLearner creates a Learner using the specified promotionPath.
func NewLearner(promotionPath string) *Learner {
	return &Learner{
		promotionPath: promotionPath,
		nowFn:         time.Now,
	}
}

// AggregatePatterns reads JSONL logs from logPath and returns a pattern map grouped by
// (event_type, subject, context_hash) combinations.
// REQ-HL-002: Returns empty map if file does not exist or is empty.
//
// @MX:ANCHOR: [AUTO] Used by multiple callers as the first step of the learning pipeline.
// @MX:REASON: [AUTO] fan_in >= 3: learner_test.go, CLI harness status, Phase 5 IT-01
func AggregatePatterns(logPath string) (map[string]*Pattern, error) {
	patterns := make(map[string]*Pattern)

	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found is normal state — return empty map
			return patterns, nil
		}
		return nil, fmt.Errorf("learner: 로그 파일 열기 실패 %s: %w", logPath, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			// Skip parsing failure lines (prevent data loss)
			continue
		}

		key := buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash)
		if p, ok := patterns[key]; ok {
			p.Count++
		} else {
			patterns[key] = &Pattern{
				Key:         key,
				EventType:   evt.EventType,
				Subject:     evt.Subject,
				ContextHash: evt.ContextHash,
				Count:       1,
				Confidence:  defaultConfidence,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("learner: 로그 파일 스캔 오류: %w", err)
	}

	return patterns, nil
}

// defaultConfidence is the initial confidence default value.
// Will be updated via canary check in Phase 3 safety layer.
const defaultConfidence = 1.0

// buildPatternKey generates a unique key from (event_type, subject, context_hash) combination.
func buildPatternKey(et EventType, subject, contextHash string) string {
	return fmt.Sprintf("%s:%s:%s", et, subject, contextHash)
}

// ClassifyTier returns Tier based on Pattern's count and confidence.
// REQ-HL-002:
//   - If confidence < 0.70, returns TierObservation regardless of count.
//   - If count >= thresholds[3] (10+), returns TierAutoUpdate.
//   - If count >= thresholds[2] (5+), returns TierRule.
//   - If count >= thresholds[1] (3+), returns TierHeuristic.
//   - If count >= thresholds[0] (1+), returns TierObservation.
//
// thresholds must be in [1, 3, 5, 10] format (plan.md §4.3).
// Returns TierObservation if thresholds is empty.
func ClassifyTier(p *Pattern, thresholds []int) Tier {
	// Insufficient confidence: force Observation regardless of count
	if p.Confidence < confidenceThreshold {
		return TierObservation
	}

	if len(thresholds) == 0 {
		return TierObservation
	}

	// thresholds = [t0, t1, t2, t3] = [1, 3, 5, 10]
	// count >= t3 → AutoUpdate
	// count >= t2 → Rule
	// count >= t1 → Heuristic
	// count >= t0 → Observation
	// count < t0  → Observation (before observation)
	if len(thresholds) >= 4 && p.Count >= thresholds[3] {
		return TierAutoUpdate
	}
	if len(thresholds) >= 3 && p.Count >= thresholds[2] {
		return TierRule
	}
	if len(thresholds) >= 2 && p.Count >= thresholds[1] {
		return TierHeuristic
	}
	// Both count >= thresholds[0] and count < thresholds[0] are Observation
	return TierObservation
}

// WritePromotion appends Promotion events to tier-promotions.jsonl.
// REQ-HL-002: Records according to plan.md §4.2 schema.
// Automatically creates parent directory if it does not exist.
func (l *Learner) WritePromotion(p Promotion) error {
	// Auto-create parent directory
	dir := filepath.Dir(l.promotionPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("learner: promotion 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// Set to current time if Ts is empty
	if p.Ts.IsZero() {
		p.Ts = l.nowFn().UTC()
	}

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("learner: promotion 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	f, err := os.OpenFile(l.promotionPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("learner: promotion 파일 열기 실패 %s: %w", l.promotionPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("learner: promotion 파일 쓰기 실패: %w", err)
	}

	return nil
}
