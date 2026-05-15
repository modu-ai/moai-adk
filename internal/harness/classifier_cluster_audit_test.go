// Package harness — classifier_cluster_audit_test.go
// Stage-2 클러스터 감사 로그 스키마 검증 테스트.
// AC-HRN-CLS-008: JSONL 감사 로그 필드 구조 검증.
// EDGE-005: 100개 멤버 클러스터의 hamming_distances CAP=20 검증.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// clusterAuditEntry는 cluster-merges.jsonl 한 줄의 파싱 스키마다.
type clusterAuditEntry struct {
	Ts               string  `json:"ts"`
	MemberKeys       []string `json:"member_keys"`
	MemberCounts     []int   `json:"member_counts"`
	HammingDistances []int   `json:"hamming_distances"`
	HammingPairCount int     `json:"hamming_pair_count"`
	Truncated        bool    `json:"truncated"`
	MergedKey        string  `json:"merged_key"`
	MergedCount      int     `json:"merged_count"`
	Confidence       float64 `json:"confidence"`
}

// TestClusterAuditLog_4MemberSchemaShape는 4개 멤버 클러스터가 C(4,2)=6개의 hamming_distances를 가지는지 검증한다
// (AC-HRN-CLS-008).
func TestClusterAuditLog_4MemberSchemaShape(t *testing.T) {
	t.Parallel()

	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 4 {
		subject := fmt.Sprintf("audit-4-member-%03d", i)
		preview := "moai run audit log four member cluster shape test"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, fmt.Sprintf("a4%d", i), preview, defaultConfidence)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	_, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("감사 로그 읽기 실패: %v", err)
	}

	lines := splitNonEmpty(string(data))
	if len(lines) != 1 {
		t.Fatalf("감사 로그 라인 수 = %d, want 1", len(lines))
	}

	var entry clusterAuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("감사 로그 파싱 실패: %v", err)
	}

	// 4멤버: C(4,2) = 6개 hamming_distances
	wantDistLen := 6 // C(4,2)
	if len(entry.HammingDistances) != wantDistLen {
		t.Errorf("hamming_distances 길이 = %d, want %d", len(entry.HammingDistances), wantDistLen)
	}
	if entry.HammingPairCount != wantDistLen {
		t.Errorf("hamming_pair_count = %d, want %d", entry.HammingPairCount, wantDistLen)
	}
	if entry.Truncated {
		t.Error("4멤버 클러스터: truncated = true이면 안 됨")
	}

	// 필수 필드 검증
	if entry.Ts == "" {
		t.Error("ts 필드 비어있음")
	}
	if len(entry.MemberKeys) != 4 {
		t.Errorf("member_keys 길이 = %d, want 4", len(entry.MemberKeys))
	}
	if len(entry.MemberCounts) != 4 {
		t.Errorf("member_counts 길이 = %d, want 4", len(entry.MemberCounts))
	}
	if entry.MergedKey == "" {
		t.Error("merged_key 비어있음")
	}
	if entry.MergedCount != 4 {
		t.Errorf("merged_count = %d, want 4", entry.MergedCount)
	}
}

// TestClusterAuditLog_100MemberTruncation는 100개 멤버 클러스터가 hamming_distances를 20으로 자르고
// truncated=true, hamming_pair_count=4950을 설정하는지 검증한다 (EDGE-005).
func TestClusterAuditLog_100MemberTruncation(t *testing.T) {
	t.Parallel()

	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 100 {
		subject := fmt.Sprintf("edge005-100member-%03d", i)
		preview := "moai run edge case one hundred member truncation test"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, fmt.Sprintf("e5%d", i), preview, defaultConfidence)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	_, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("감사 로그 읽기 실패: %v", err)
	}

	lines := splitNonEmpty(string(data))
	if len(lines) == 0 {
		t.Fatal("감사 로그 비어있음")
	}

	// 첫 번째 (유일한) 라인 검증
	var entry clusterAuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("감사 로그 파싱 실패: %v", err)
	}

	// 100멤버: C(100,2) = 4950 → 20으로 CAP
	wantPairCount := 4950 // 100*99/2
	if entry.HammingPairCount != wantPairCount {
		t.Errorf("hamming_pair_count = %d, want %d", entry.HammingPairCount, wantPairCount)
	}
	if len(entry.HammingDistances) != 20 {
		t.Errorf("hamming_distances 길이 = %d, want 20 (CAP=20)", len(entry.HammingDistances))
	}
	if !entry.Truncated {
		t.Error("100멤버 클러스터: truncated = false이면 안 됨")
	}
}

// TestClusterAuditLog_AllFieldsParseable는 모든 JSONL 필드가 json.Unmarshal로 파싱 가능한지 검증한다.
func TestClusterAuditLog_AllFieldsParseable(t *testing.T) {
	t.Parallel()

	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 5 {
		subject := fmt.Sprintf("parseable-check-%03d", i)
		preview := "moai run parseable check all fields test schema"
		p, evt := makePatternAndEvent(EventTypeAgentInvocation, subject, fmt.Sprintf("pc%d", i), preview, 0.85)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	_, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("감사 로그 읽기 실패: %v", err)
	}

	for i, line := range splitNonEmpty(string(data)) {
		var entry clusterAuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("라인 %d: json.Unmarshal 실패: %v", i, err)
		}
	}
}

// splitNonEmpty는 문자열을 줄 단위로 분할하고 빈 줄은 제거한다.
func splitNonEmpty(s string) []string {
	var result []string
	for _, line := range splitLines(s) {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}
