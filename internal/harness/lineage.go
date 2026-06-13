// Package harness — M6 auditable lineage manifest writer + loader.
// SPEC-HARNESS-LOOP-CLOSURE-001 REQ-HLC-001/002.
//
// 매 apply 전환마다 manifest.jsonl에 LineageEntry를 append한다 (accept 1개, reject 1개).
// learner.go:145 WritePromotion의 append-JSONL 관용구를 mirror한다:
// MkdirAll 부모 디렉토리 → Timestamp default → json.Marshal → '\n' append →
// O_APPEND|O_CREATE|O_WRONLY 0o644 open → write.
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

// WriteLineageEntry는 manifestPath에 단일 LineageEntry를 append한다 (부모 디렉토리 자동 생성).
// REQ-HLC-001/002: manifest는 append-only이며, Timestamp가 zero면 time.Now().UTC()로 채운다.
//
// manifestPath는 PARAMETER이다 (하드코딩 금지) — 테스트는 t.TempDir()을 주입하고,
// 프로덕션 caller는 <learning-history-dir>/manifest.jsonl을 전달한다.
//
// @MX:ANCHOR: [AUTO] WriteLineageEntry는 lineage 기록의 단일 진입점.
// @MX:REASON: [AUTO] fan_in >= 3: applier.go(Apply accept+reject), lineage_test.go, harness CLI(future)
func WriteLineageEntry(manifestPath string, entry LineageEntry) error {
	// 부모 디렉토리 자동 생성 (learner.go:147 WritePromotion mirror).
	dir := filepath.Dir(manifestPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("lineage: manifest 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	// Timestamp가 zero면 현재 시각(UTC)으로 채운다.
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("lineage: entry 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	f, err := os.OpenFile(manifestPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("lineage: manifest 파일 열기 실패 %s: %w", manifestPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("lineage: manifest 파일 쓰기 실패: %w", err)
	}

	return nil
}

// LoadManifest는 manifestPath의 모든 LineageEntry를 write 순서대로 읽는다.
// REQ-HLC-005: 파일이 없으면 ([]LineageEntry{}, nil)을 반환한다 (backward compat — 에러 아님).
// 빈 줄은 건너뛰고, 파싱 실패 줄은 데이터 손실 방지를 위해 건너뛴다.
//
// @MX:ANCHOR: [AUTO] LoadManifest는 lineage 조회의 단일 진입점.
// @MX:REASON: [AUTO] fan_in >= 3: lineage_test.go(여러 AC), harness CLI status(future), Phase 5 IT
func LoadManifest(manifestPath string) ([]LineageEntry, error) {
	entries := []LineageEntry{}

	f, err := os.Open(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일 부재는 정상 상태 — 빈 슬라이스 반환 (backward compat).
			return entries, nil
		}
		return nil, fmt.Errorf("lineage: manifest 파일 열기 실패 %s: %w", manifestPath, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry LineageEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// 파싱 실패 줄은 건너뛴다 (데이터 손실 방지, learner.go AggregatePatterns mirror).
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("lineage: manifest 파일 스캔 오류: %w", err)
	}

	return entries, nil
}
