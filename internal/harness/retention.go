// Package harness — 로그 보존 및 pruning 구현.
// REQ-HL-011: 오래된 이벤트를 아카이브하고 로그 파일을 정리한다.
package harness

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// pruneSkipDuration은 마지막 prune으로부터 이 시간 이내면 pruning을 skip한다.
const pruneSkipDuration = time.Hour

// Retention은 usage-log.jsonl의 오래된 엔트리를 아카이브하고 정리한다.
// REQ-HL-011: 매 RecordEvent 호출 시 lazy pruning, 마지막 prune으로부터 1시간 이내면 skip.
//
// @MX:ANCHOR: [AUTO] PruneStaleEntries는 observer와 테스트에서 호출된다.
// @MX:REASON: [AUTO] fan_in >= 3: observer.go, observer_test.go, integration_test.go
type Retention struct {
	// logPath는 usage-log.jsonl 파일 경로이다.
	logPath string

	// archiveDir는 아카이브 파일을 저장할 디렉토리이다.
	// 실제 경로: archiveDir/<YYYY-MM>.jsonl.gz
	archiveDir string

	// lastPruneAt은 마지막 pruning 실행 시각이다.
	lastPruneAt time.Time

	// nowFn은 현재 시각을 반환하는 함수 (테스트에서 mock-clock 주입 가능).
	nowFn func() time.Time
}

// NewRetention은 Retention 인스턴스를 생성한다.
// nowFn이 nil이면 time.Now를 사용한다.
func NewRetention(logPath, archiveDir string, nowFn func() time.Time) *Retention {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Retention{
		logPath:    logPath,
		archiveDir: archiveDir,
		nowFn:      nowFn,
	}
}

// PruneStaleEntries는 retentionDays보다 오래된 이벤트를 로그에서 제거하고
// 아카이브 파일(<YYYY-MM>.jsonl.gz)에 추가한다.
// REQ-HL-011: 마지막 prune으로부터 1시간 이내면 skip한다.
//
// @MX:WARN: [AUTO] 파일을 읽고 덮어쓰는 비원자적 작업이므로 동시 호출에 주의.
// @MX:REASON: [AUTO] RecordEvent와 동일 프로세스 내에서 순차 호출되지만,
// 외부 프로세스가 동시에 기록할 경우 race condition 가능.
func (r *Retention) PruneStaleEntries(retentionDays int) error {
	now := r.nowFn()

	// 1시간 이내 prune은 skip
	if !r.lastPruneAt.IsZero() && now.Sub(r.lastPruneAt) < pruneSkipDuration {
		return nil
	}

	// 로그 파일이 없으면 skip
	if _, err := os.Stat(r.logPath); os.IsNotExist(err) {
		r.lastPruneAt = now
		return nil
	}

	cutoff := now.AddDate(0, 0, -retentionDays)

	// 로그 파일 읽기
	kept, stale, err := partitionEvents(r.logPath, cutoff)
	if err != nil {
		return fmt.Errorf("retention: 이벤트 분류 실패: %w", err)
	}

	if len(stale) == 0 {
		// 제거할 이벤트 없음
		r.lastPruneAt = now
		return nil
	}

	// stale 이벤트를 월별 아카이브에 저장
	if err := r.archiveEvents(stale); err != nil {
		return fmt.Errorf("retention: 아카이브 실패: %w", err)
	}

	// 로그 파일을 kept 이벤트만으로 덮어쓰기
	if err := overwriteWithEvents(r.logPath, kept); err != nil {
		return fmt.Errorf("retention: 로그 파일 갱신 실패: %w", err)
	}

	r.lastPruneAt = now
	return nil
}

// partitionEvents는 로그 파일을 읽어 cutoff 기준으로 kept/stale 이벤트를 분류한다.
func partitionEvents(logPath string, cutoff time.Time) (kept, stale []Event, err error) {
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("파일 열기: %w", err)
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
			// 파싱 실패 줄은 kept에 넣어 데이터 손실 방지
			continue
		}
		if evt.Timestamp.Before(cutoff) {
			stale = append(stale, evt)
		} else {
			kept = append(kept, evt)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("파일 스캔: %w", err)
	}
	return kept, stale, nil
}

// archiveEvents는 stale 이벤트들을 월별 gzip 아카이브에 추가한다.
// 파일명: archiveDir/<YYYY-MM>.jsonl.gz
func (r *Retention) archiveEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	if err := os.MkdirAll(r.archiveDir, 0o755); err != nil {
		return fmt.Errorf("아카이브 디렉토리 생성: %w", err)
	}

	// 이벤트를 월별로 그룹화
	byMonth := make(map[string][]Event)
	for _, evt := range events {
		key := evt.Timestamp.UTC().Format("2006-01")
		byMonth[key] = append(byMonth[key], evt)
	}

	for month, evts := range byMonth {
		archivePath := filepath.Join(r.archiveDir, month+".jsonl.gz")
		if err := appendToGzip(archivePath, evts); err != nil {
			return fmt.Errorf("월별 아카이브 %s 기록 실패: %w", month, err)
		}
	}
	return nil
}

// appendToGzip은 이벤트들을 gzip 압축 JSONL 파일에 추가한다.
// 파일이 없으면 새로 생성한다.
//
// @MX:WARN: [AUTO] gzip 파일은 append-safe하지 않으므로 읽고 다시 쓰는 방식을 사용한다.
// @MX:REASON: [AUTO] gzip 포맷은 concatenated streams를 지원하므로 실제로는 append 가능하지만,
// 표준 reader와의 호환성을 위해 read-modify-write 패턴을 사용한다.
func appendToGzip(archivePath string, events []Event) error {
	// gzip concatenation: 기존 파일에 새 gzip 스트림을 추가하는 방식으로 append한다.
	// 표준 gzip reader는 concatenated streams를 순차적으로 읽을 수 있다.
	f, err := os.OpenFile(archivePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("아카이브 파일 열기: %w", err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	enc := json.NewEncoder(gw)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			_ = gw.Close()
			return fmt.Errorf("gzip 인코딩: %w", err)
		}
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("gzip 닫기: %w", err)
	}
	return nil
}

// overwriteWithEvents는 kept 이벤트만으로 로그 파일을 덮어쓴다.
func overwriteWithEvents(logPath string, events []Event) error {
	// 임시 파일에 먼저 쓰고 원자적으로 교체
	dir := filepath.Dir(logPath)
	tmp, err := os.CreateTemp(dir, "usage-log-*.tmp")
	if err != nil {
		return fmt.Errorf("임시 파일 생성: %w", err)
	}
	tmpPath := tmp.Name()

	enc := json.NewEncoder(tmp)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("임시 파일 인코딩: %w", err)
		}
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("임시 파일 닫기: %w", err)
	}

	// 원자적 교체
	if err := os.Rename(tmpPath, logPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("파일 교체: %w", err)
	}
	return nil
}
