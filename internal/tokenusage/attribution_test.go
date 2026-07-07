package tokenusage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// notAvailableFallback은 session-handoff.md § Block 2 environment-fallback 패턴의
// 정준 문자열이다 (source_session_id가 사용자 실측 UUID가 아닌 fallback 토큰인 경우).
const notAvailableFallback = "<not-available — environment-fallback, next session will backfill via /moai session register on activation>"

// mustMkdirAll은 dir 생성 실패 시 테스트를 즉시 실패시킨다 (errcheck 경고 방지).
func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", dir, err)
	}
}

// writeProgressFixture는 specsDir/specDirName/progress.md를 생성하고 그 경로를
// 반환한다. sessionLines의 각 항목은 "source_session_id: <항목>" 라인으로
// 기록된다. 실제 progress.md를 건드리지 않고 t.TempDir() 내에서 fixture를
// 구성하기 위한 테스트 전용 헬퍼이다.
func writeProgressFixture(t *testing.T, specsDir, specDirName string, sessionLines []string) string {
	t.Helper()
	specDir := filepath.Join(specsDir, specDirName)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", specDir, err)
	}
	path := filepath.Join(specDir, "progress.md")
	var buf strings.Builder
	buf.WriteString("# Progress — ")
	buf.WriteString(specDirName)
	buf.WriteString("\n\n## §E.2 Run-phase Evidence\n\n")
	for _, line := range sessionLines {
		buf.WriteString("source_session_id: ")
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(buf.String()), 0o644); err != nil {
		t.Fatalf("writeProgressFixture: %v", err)
	}
	return path
}

// TestAttributionConfidence는 귀속 신뢰도 한정자의 정준 인수 테스트이다
// (acceptance.md §D.1 Scenario 3 — AC-TA-006 + AC-TA-007 + edge §D.2 shared-UUID→low).
func TestAttributionConfidence(t *testing.T) {
	// --- AC-TA-006: session-set 합산 + high 신뢰도 ---
	t.Run("session_set_high_confidence", func(t *testing.T) {
		root := t.TempDir()
		specsDir := filepath.Join(root, "specs")
		transcriptDir := filepath.Join(root, "transcripts")

		uuidA := "11111111-aaaa-1111-aaaa-111111111111"
		uuidB := "22222222-bbbb-2222-bbbb-222222222222"

		progressPath := writeProgressFixture(t, specsDir, "SPEC-TEST-001", []string{uuidA, uuidB})
		mustMkdirAll(t, transcriptDir)

		writeTranscript(t, transcriptDir, uuidA+".jsonl", []string{assistantLine(100, 20, 0, 500)})
		writeTranscript(t, transcriptDir, uuidB+".jsonl", []string{assistantLine(200, 40, 0, 1000)})

		attr, err := Attribute(progressPath, transcriptDir, specsDir, "")
		if err != nil {
			t.Fatalf("Attribute error: %v", err)
		}
		if attr.AttributionMethod != AttributionSessionSet {
			t.Errorf("method = %q, want %q", attr.AttributionMethod, AttributionSessionSet)
		}
		if attr.Confidence != ConfidenceHigh {
			t.Errorf("confidence = %q, want %q", attr.Confidence, ConfidenceHigh)
		}
		if attr.SessionCount != 2 {
			t.Errorf("session_count = %d, want 2", attr.SessionCount)
		}
		// 합산 검증: input=300, output=60, cache_read=1500, spent=1860
		if attr.TokensInput != 300 {
			t.Errorf("TokensInput = %d, want 300", attr.TokensInput)
		}
		if attr.TokensOutput != 60 {
			t.Errorf("TokensOutput = %d, want 60", attr.TokensOutput)
		}
		if attr.TokensCacheRead != 1500 {
			t.Errorf("TokensCacheRead = %d, want 1500", attr.TokensCacheRead)
		}
		if attr.TokensSpent != 1860 {
			t.Errorf("TokensSpent = %d, want 1860", attr.TokensSpent)
		}
		// 기여 세션 추적성
		if len(attr.ContributingSessions) != 2 {
			t.Errorf("ContributingSessions len = %d, want 2", len(attr.ContributingSessions))
		}
	})

	// --- AC-TA-007: lineage 부재 → low 폴백 ---
	t.Run("lineage_absent_low_fallback", func(t *testing.T) {
		root := t.TempDir()
		specsDir := filepath.Join(root, "specs")
		transcriptDir := filepath.Join(root, "transcripts")

		activeUUID := "33333333-cccc-3333-cccc-333333333333"

		progressPath := writeProgressFixture(t, specsDir, "SPEC-TEST-001", []string{notAvailableFallback})
		mustMkdirAll(t, transcriptDir)

		writeTranscript(t, transcriptDir, activeUUID+".jsonl", []string{assistantLine(50, 10, 0, 100)})

		attr, err := Attribute(progressPath, transcriptDir, specsDir, activeUUID)
		if err != nil {
			t.Fatalf("Attribute error: %v", err)
		}
		if attr.Confidence != ConfidenceLow {
			t.Errorf("confidence = %q, want %q (lineage unavailable → low)", attr.Confidence, ConfidenceLow)
		}
		if attr.AttributionMethod != AttributionCurrentSession {
			t.Errorf("method = %q, want %q (fallback)", attr.AttributionMethod, AttributionCurrentSession)
		}
		if attr.SessionCount != 1 {
			t.Errorf("session_count = %d, want 1 (active session only)", attr.SessionCount)
		}
		if attr.TokensInput != 50 {
			t.Errorf("TokensInput = %d, want 50", attr.TokensInput)
		}
	})

	// --- Edge §D.2: 동일 UUID가 2 SPEC lineage에 등장 → low 강등 ---
	t.Run("shared_uuid_degrades_to_low", func(t *testing.T) {
		root := t.TempDir()
		specsDir := filepath.Join(root, "specs")
		transcriptDir := filepath.Join(root, "transcripts")

		sharedUUID := "44444444-dddd-4444-dddd-444444444444"
		dedicatedUUID := "55555555-eeee-5555-eeee-555555555555"

		// SPEC-TEST-001은 shared UUID + dedicated UUID를 참조
		progress1 := writeProgressFixture(t, specsDir, "SPEC-TEST-001", []string{sharedUUID, dedicatedUUID})
		// SPEC-TEST-002도 shared UUID를 참조 → 세션 공유
		writeProgressFixture(t, specsDir, "SPEC-TEST-002", []string{sharedUUID})
		mustMkdirAll(t, transcriptDir)

		writeTranscript(t, transcriptDir, sharedUUID+".jsonl", []string{assistantLine(100, 20, 0, 500)})
		writeTranscript(t, transcriptDir, dedicatedUUID+".jsonl", []string{assistantLine(100, 20, 0, 500)})

		attr, err := Attribute(progress1, transcriptDir, specsDir, "")
		if err != nil {
			t.Fatalf("Attribute error: %v", err)
		}
		if attr.Confidence != ConfidenceLow {
			t.Errorf("confidence = %q, want %q (shared UUID → low)", attr.Confidence, ConfidenceLow)
		}
		// lineage는 가용하므로 method는 여전히 session-set
		if attr.AttributionMethod != AttributionSessionSet {
			t.Errorf("method = %q, want %q", attr.AttributionMethod, AttributionSessionSet)
		}
	})
}

// TestAttributionAbsentTranscript는 REQ-TA-013 견고성을 검증한다: 세션 집합 내
// UUID의 transcript 파일이 부재하면 해당 세션은 0 기여로 skip-and-continue되며
// panic이나 중단 없이 진행한다.
func TestAttributionAbsentTranscript(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs")
	transcriptDir := filepath.Join(root, "transcripts")

	uuidPresent := "66666666-ffff-6666-ffff-666666666666"
	uuidAbsent := "77777777-abcd-7777-abcd-777777777777"

	progressPath := writeProgressFixture(t, specsDir, "SPEC-TEST-001", []string{uuidPresent, uuidAbsent})
	mustMkdirAll(t, transcriptDir)

	// uuidPresent의 transcript만 존재; uuidAbsent는 파일 부재
	writeTranscript(t, transcriptDir, uuidPresent+".jsonl", []string{assistantLine(100, 20, 0, 500)})

	attr, err := Attribute(progressPath, transcriptDir, specsDir, "")
	if err != nil {
		t.Fatalf("Attribute error on absent transcript: %v", err)
	}
	// 부재 transcript는 0 기여 → present 세션만 반영
	if attr.TokensSpent != 620 {
		t.Errorf("TokensSpent = %d, want 620 (only present transcript counts)", attr.TokensSpent)
	}
	// 세션 집합 크기는 2 (부재 세션도 집합에 포함, 0 기여)
	if attr.SessionCount != 2 {
		t.Errorf("SessionCount = %d, want 2 (absent session still in set)", attr.SessionCount)
	}
}

// TestAttributionNoLineageNoActive는 lineage도 없고 활성 세션도 없는 퇴화 케이스를
// 검증한다: panic 없이 빈 귀속 결과를 반환한다.
func TestAttributionNoLineageNoActive(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs")
	transcriptDir := filepath.Join(root, "transcripts")

	progressPath := writeProgressFixture(t, specsDir, "SPEC-TEST-001", []string{notAvailableFallback})

	attr, err := Attribute(progressPath, transcriptDir, specsDir, "")
	if err != nil {
		t.Fatalf("Attribute error: %v", err)
	}
	if attr.SessionCount != 0 {
		t.Errorf("SessionCount = %d, want 0", attr.SessionCount)
	}
	if attr.TokensSpent != 0 {
		t.Errorf("TokensSpent = %d, want 0", attr.TokensSpent)
	}
	if attr.Confidence != ConfidenceLow {
		t.Errorf("confidence = %q, want %q", attr.Confidence, ConfidenceLow)
	}
}

// TestAttributionProgressNotFound는 progress.md 자체가 부재한 경우를 검증한다.
func TestAttributionProgressNotFound(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs")
	transcriptDir := filepath.Join(root, "transcripts")
	missingPath := filepath.Join(specsDir, "SPEC-NOPE-001", "progress.md")

	_, err := Attribute(missingPath, transcriptDir, specsDir, "")
	if err == nil {
		t.Fatalf("want error for missing progress.md, got nil")
	}
}

// TestExtractSessionUUIDs는 lineage UUID 파서를 직접 단위 테스트한다.
func TestExtractSessionUUIDs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "two real UUIDs",
			content: "source_session_id: 11111111-aaaa-1111-aaaa-111111111111\n" +
				"source_session_id: 22222222-bbbb-2222-bbbb-222222222222\n",
			want: []string{
				"11111111-aaaa-1111-aaaa-111111111111",
				"22222222-bbbb-2222-bbbb-222222222222",
			},
		},
		{
			name: "environment-fallback line skipped",
			content: "source_session_id: <not-available — environment-fallback>\n" +
				"source_session_id: 33333333-cccc-3333-cccc-333333333333\n",
			want: []string{"33333333-cccc-3333-cccc-333333333333"},
		},
		{
			name:    "only not-available → empty",
			content: "source_session_id: <not-available — environment-fallback>\n",
			want:    nil,
		},
		{
			name: "duplicate UUID deduplicated",
			content: "source_session_id: 11111111-aaaa-1111-aaaa-111111111111\n" +
				"source_session_id: 11111111-aaaa-1111-aaaa-111111111111\n",
			want: []string{"11111111-aaaa-1111-aaaa-111111111111"},
		},
		{
			name:    "no source_session_id lines",
			content: "some other content\n## §E.2\n",
			want:    nil,
		},
		{
			name: "UUID in non-source_session_id line ignored",
			content: "a random note about 11111111-aaaa-1111-aaaa-111111111111 here\n" +
				"source_session_id: 22222222-bbbb-2222-bbbb-222222222222\n",
			want: []string{"22222222-bbbb-2222-bbbb-222222222222"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSessionUUIDs(tt.content)
			if len(got) != len(tt.want) {
				t.Fatalf("extractSessionUUIDs got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
