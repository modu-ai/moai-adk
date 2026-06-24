package cluster

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// ── 테스트 픽스처 헬퍼 ──────────────────────────────────────────────

// ts 는 결정론적 타임스탬프를 만든다(테스트 가독성용 헬퍼).
func ts(day int) time.Time {
	return time.Date(2026, time.June, day, 12, 0, 0, 0, time.UTC)
}

// applyOutcomeLine 은 apply_outcome 이벤트 하나를 JSONL 한 줄로 직렬화한다.
func applyOutcomeLine(t *testing.T, verdict, decision, proposalID string, regressed []string, when time.Time) string {
	t.Helper()
	evt := harness.Event{
		Timestamp:         when,
		EventType:         harness.EventTypeApplyOutcome,
		Subject:           "apply:" + proposalID,
		SchemaVersion:     harness.LogSchemaVersion,
		OutcomeVerdict:    verdict,
		OutcomeDecision:   decision,
		OutcomeProposalID: proposalID,
		OutcomeRegressed:  regressed,
	}
	b, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("apply_outcome 직렬화 실패: %v", err)
	}
	return string(b)
}

// nonOutcomeLine 은 비-apply_outcome 이벤트 한 줄을 만든다(건너뛰어져야 함).
func nonOutcomeLine(t *testing.T) string {
	t.Helper()
	evt := harness.Event{
		Timestamp:     ts(1),
		EventType:     harness.EventTypeMoaiSubcommand,
		Subject:       "/moai plan",
		SchemaVersion: harness.LogSchemaVersion,
	}
	b, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("non-outcome 직렬화 실패: %v", err)
	}
	return string(b)
}

// writeLog 은 줄 슬라이스를 t.TempDir() 안의 usage-log.jsonl 로 기록하고 경로를 반환한다.
func writeLog(t *testing.T, lines []string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "usage-log.jsonl")
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("로그 파일 기록 실패: %v", err)
	}
	return path
}

// ── AC-OBL-010 / EC-1·EC-2·EC-3: 수집 엣지 케이스 (REQ-OBL-002/003/004) ──

func TestLoadEvents(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string // nil 이면 파일 자체를 만들지 않음(부재 케이스)
		makeFile  bool     // false 이면 파일 부재 케이스
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid apply_outcome 줄들만 수집한다 (REQ-OBL-001)",
			lines: []string{
				applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
				applyOutcomeLine(t, verdictKept, "approved", "PROPOSAL-B", nil, ts(2)),
			},
			makeFile:  true,
			wantCount: 2, // LoadEvents 는 verdict 필터링을 하지 않음 — ClusterEvents 가 함
			wantErr:   false,
		},
		{
			name: "손상된 줄은 건너뛰고 나머지 유효 줄을 수집한다 (REQ-OBL-003 fail-open, EC-2)",
			lines: []string{
				applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
				`{this is not valid json at all`,
				applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-C", []string{"lint"}, ts(3)),
			},
			makeFile:  true,
			wantCount: 2, // 손상된 줄 1개는 건너뛰고 유효 2개만
			wantErr:   false,
		},
		{
			name: "비-apply_outcome 이벤트는 건너뛴다",
			lines: []string{
				nonOutcomeLine(t),
				applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
				nonOutcomeLine(t),
			},
			makeFile:  true,
			wantCount: 1, // apply_outcome 1개만
			wantErr:   false,
		},
		{
			name:      "빈 파일은 0개 이벤트 + 성공 (REQ-OBL-004, EC-1)",
			lines:     []string{},
			makeFile:  true,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "공백/빈 줄은 건너뛴다",
			lines:     []string{"", "   ", applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)), ""},
			makeFile:  true,
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.makeFile {
				path = writeLog(t, tt.lines)
			} else {
				path = filepath.Join(t.TempDir(), "usage-log.jsonl") // 존재하지 않는 경로
			}

			got, err := LoadEvents(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadEvents() err = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantCount {
				t.Errorf("LoadEvents() count = %d, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestLoadEventsAbsentFile: 파일 부재 → 0개 이벤트 + 성공(에러 아님) (REQ-OBL-004, EC-1).
func TestLoadEventsAbsentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.jsonl")
	got, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("부재 파일은 에러가 아니어야 함: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("부재 파일은 0개 이벤트여야 함, got %d", len(got))
	}
}

// TestLoadEventsManifestAbsentTolerated: usage-log.jsonl 단독으로 동작하며
// manifest.jsonl 부재는 에러가 아니다 (REQ-OBL-002, EC-3). LoadEvents 는
// usage-log.jsonl 만 읽고 manifest 를 요구하지 않으므로, manifest 가 없어도
// 정상 수집한다 — 그것이 EC-3 의 read-only 클러스터링 의미다.
func TestLoadEventsManifestAbsentTolerated(t *testing.T) {
	path := writeLog(t, []string{
		applyOutcomeLine(t, verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
	})
	// manifest.jsonl 은 의도적으로 만들지 않는다. LoadEvents 는 usage-log 단독으로 동작해야 한다.
	got, err := LoadEvents(path)
	if err != nil {
		t.Fatalf("manifest 부재는 에러가 아니어야 함: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("usage-log 단독 수집 count = %d, want 1", len(got))
	}
}
