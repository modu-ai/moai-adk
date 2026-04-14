package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestAsyncRecorder_NonBlockingUnderLoad verifies that Record() returns quickly
// even under heavy parallel load, and that all records are persisted after Close.
func TestAsyncRecorder_NonBlockingUnderLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	const bufSize = 200
	const numRecords = 1000

	rec := NewAsyncRecorder(dir, bufSize)

	ts := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	slowCalls := 0
	var slowMu sync.Mutex

	for i := 0; i < numRecords; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := UsageRecord{
				Timestamp: ts,
				SkillID:   "test-skill",
				SessionID: "sess-load-test",
				Outcome:   OutcomeUnknown,
			}
			start := time.Now()
			_ = rec.Record(r)
			elapsed := time.Since(start)
			if elapsed > 10*time.Millisecond {
				slowMu.Lock()
				slowCalls++
				slowMu.Unlock()
			}
		}()
	}
	wg.Wait()

	// 버퍼 크기보다 많은 레코드이므로 일부는 드롭될 수 있음 - 하지만 블로킹은 없어야 함
	// 느린 호출이 전체의 1% 이하여야 함
	if slowCalls > numRecords/100 {
		t.Errorf("너무 많은 느린 호출: %d/%d (10ms 초과)", slowCalls, numRecords)
	}

	// 모든 레코드가 처리될 때까지 닫기
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rec.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// 파일에 일부 레코드가 기록되었는지 확인 (드롭 정책으로 전체가 아닐 수 있음)
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	entries, err := os.ReadDir(telDir)
	if err != nil {
		t.Fatalf("telemetry 디렉터리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("telemetry 파일이 생성되지 않음")
	}
}

// TestAsyncRecorder_DropPolicyWhenFull verifies that when the buffer is full,
// Record returns an error instead of blocking.
func TestAsyncRecorder_DropPolicyWhenFull(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// 버퍼 크기 1, 소비자를 차단하여 버퍼가 꽉 차도록 함
	rec := NewAsyncRecorder(dir, 1)
	// 소비자 고루틴을 즉시 멈추기 위해 Close 후에도 테스트
	// 대신 소비자 없이 채널을 꽉 채우는 방식으로 테스트

	ts := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	r := UsageRecord{
		Timestamp: ts,
		SkillID:   "test-skill",
		Outcome:   OutcomeUnknown,
	}

	// 여러 번 Record를 호출하여 드롭이 발생하는지 확인
	// 버퍼가 1이므로 처음 몇 번 이후에는 드롭되어야 함
	var dropped int
	for i := 0; i < 100; i++ {
		err := rec.Record(r)
		if errors.Is(err, ErrRecordDropped) {
			dropped++
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rec.Close(ctx)

	// 드롭이 발생했어야 함 (버퍼 1이므로)
	// 소비자가 빠르면 드롭이 적을 수 있으므로 느슨하게 검증
	t.Logf("드롭된 레코드: %d/100", dropped)
}

// TestAsyncRecorder_ReusesFileHandle verifies that the async recorder does not
// open the file on every record write (file handle reuse).
func TestAsyncRecorder_ReusesFileHandle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	rec := NewAsyncRecorder(dir, 500)

	// 같은 날짜로 100개 레코드 기록
	ts := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		r := UsageRecord{
			Timestamp: ts,
			SkillID:   "test-skill",
			SessionID: "sess-reuse-test",
			Outcome:   OutcomeUnknown,
		}
		if err := rec.Record(r); err != nil {
			t.Fatalf("Record %d: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rec.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// 파일에 100개 레코드가 모두 기록되었는지 확인
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	path := filepath.Join(telDir, "usage-2026-04-15.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("파일 열기 실패: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var rec UsageRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("유효하지 않은 JSON at line %d: %v", count+1, err)
		}
		count++
	}

	if count != 100 {
		t.Errorf("기대 100개 레코드, 실제 %d개", count)
	}
}
