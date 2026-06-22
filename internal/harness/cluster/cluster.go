// Package cluster — 실패 시그니처 클러스터링 엔진 (Epic Dive-into-CC N4).
//
// 이 패키지는 harness-learning 루프(trace → eval → cluster → policy → repair)에서
// 유일하게 비어 있던 `cluster` 단계를 채운다. usage-log.jsonl 에 이미 기록된
// apply_outcome 이벤트를 읽어 들여(read-only), 결정론적 시그니처 키로
// 실패/롤백 이벤트를 그룹화한다.
//
// HONEST FRAMING (spec.md §A.3): 이 패키지는 read-only 관측(observability) 표면이다.
//   - 입력(usage-log.jsonl / manifest.jsonl / tier-promotions.jsonl)은 READ-ONLY로 취급한다.
//   - 출력은 오직 learning-history/ 아래의 자체 리포트 아티팩트 하나뿐이다.
//   - proposal/apply 경로에 절대 write-back 하지 않으며 Apply 결정을 바꾸지 않는다.
//   - autoApply 기본값(false)을 절대 읽거나 쓰거나 뒤집지 않는다.
//
// 결정론(determinism)이 검증의 핵심 지렛대다(REQ-OBL-007): 머신러닝/난수를 쓰지 않고,
// 시그니처 키는 정렬된 outcome_regressed 차원 집합 + outcome_verdict + outcome_decision
// 에서만 파생되며(pattern_key는 apply_outcome 이벤트에 존재하지 않음 — REQ-OBL-005),
// 클러스터 출력은 안정 키로 정렬되어 동일 입력에 대해 바이트 동일 출력을 보장한다.
package cluster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

const (
	// verdictRolledBack 은 실패(롤백) 결과 verdict 값이다. 오직 이 verdict 의
	// 이벤트만 실패 클러스터에 포함된다(REQ-OBL-008).
	verdictRolledBack = "rolled-back"

	// verdictKept 은 비-실패(유지) 결과 verdict 값이다. kept 이벤트는 어떤 실패
	// 클러스터에도 포함되지 않는다(REQ-OBL-008, AC-OBL-003).
	verdictKept = "kept"

	// sigDimDelimiter 는 정렬된 regressed 차원 집합을 하나의 키 조각으로 합칠 때
	// 쓰는 안정 구분자다. JSONL 필드 값에 등장하지 않는 문자를 선택한다.
	sigDimDelimiter = "|"

	// sigFieldDelimiter 는 (차원집합, verdict, decision) 세 조각을 합칠 때 쓰는
	// 안정 구분자다.
	sigFieldDelimiter = "::"

	// emptyDimToken 은 regressed 차원 집합이 비어 있을 때 키에 쓰는 안정 토큰이다.
	// (롤백이지만 regressed 목록이 omitempty 로 빠진 경우 등.)
	emptyDimToken = "(none)"

	// scanBufferMax 는 bufio.Scanner 의 최대 토큰 크기다. usage-log.jsonl 의 한 줄이
	// 기본 64KB 버퍼를 넘는 경우(긴 prompt_content 등)를 대비해 넉넉히 잡는다.
	scanBufferMax = 4 * 1024 * 1024 // 4 MiB
)

// EventRef 는 한 클러스터에 속한 멤버 이벤트의 경량 참조다. 전체 Event 를 복사하지
// 않고 시그니처 파생에 쓰인 식별 정보와 타임스탬프만 보존한다.
type EventRef struct {
	// ProposalID 는 apply_outcome 이벤트의 outcome_proposal_id 상관 키다.
	ProposalID string `json:"proposal_id"`

	// Timestamp 는 이벤트 발생 시각(입력에 고정된 값 — time.Now() 가 아님)이다.
	Timestamp time.Time `json:"timestamp"`
}

// FailureCluster 는 동일 시그니처 키를 공유하는 실패/롤백 이벤트의 그룹이다
// (REQ-OBL-006).
type FailureCluster struct {
	// Signature 는 이 클러스터의 결정론적 시그니처 키다(REQ-OBL-005).
	Signature string `json:"signature"`

	// Count 는 멤버 이벤트 수다(최소 크기 임계값 없음 — count==1 도 유효, EC-5).
	Count int `json:"count"`

	// Members 는 멤버 이벤트의 경량 참조 목록이다(시간 순 정렬).
	Members []EventRef `json:"members"`

	// RepresentativeDimensions 는 대표 regressed 차원 집합(정렬됨)이다.
	RepresentativeDimensions []string `json:"representative_dimensions"`

	// Verdict 는 이 클러스터를 구성한 verdict(현재는 항상 "rolled-back")다.
	Verdict string `json:"verdict"`

	// Decision 은 이 클러스터를 구성한 transition decision 이다.
	Decision string `json:"decision"`

	// FirstSeen / LastSeen 는 멤버 이벤트 타임스탬프의 최소/최대값이다.
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// signatureKey 는 apply_outcome 이벤트에 실제로 존재하는 필드(types.go:151-176)에서만
// 결정론적 시그니처 키를 파생한다(REQ-OBL-005): 정렬된 OutcomeRegressed 차원 집합 +
// OutcomeVerdict + OutcomeDecision. pattern_key 에는 절대 의존하지 않는다 —
// pattern_key 는 apply_outcome 이벤트에 없으며 OutcomeProposalID 로 단방향
// sha256 해시되어(mapper.go:97,102-107) 복원 불가능하기 때문이다. ML/난수 없음.
func signatureKey(evt harness.Event) string {
	dims := normalizeDimensions(evt.OutcomeRegressed)
	dimPart := emptyDimToken
	if len(dims) > 0 {
		dimPart = strings.Join(dims, sigDimDelimiter)
	}
	return strings.Join([]string{dimPart, evt.OutcomeVerdict, evt.OutcomeDecision}, sigFieldDelimiter)
}

// normalizeDimensions 는 regressed 차원 슬라이스를 결정론적 표현으로 정규화한다:
// 빈 항목 제거 후 안정 정렬한다. 입력 슬라이스는 변경하지 않는다(복사본 정렬).
func normalizeDimensions(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, d := range in {
		if d = strings.TrimSpace(d); d != "" {
			out = append(out, d)
		}
	}
	sort.Strings(out)
	return out
}

// LoadEvents 는 logPath(usage-log.jsonl)에서 apply_outcome 이벤트만 읽어 들인다
// (REQ-OBL-001). fail-open 수집:
//   - 파싱 실패/비-apply_outcome 줄은 건너뛰고 계속 진행한다(REQ-OBL-003, EC-2).
//   - 파일이 없거나 비어 있으면 0개 이벤트 + 성공(에러 아님)을 반환한다(REQ-OBL-004, EC-1).
//
// manifest.jsonl 은 선택적 보충 입력이다(REQ-OBL-002): 본 read-only 클러스터링은
// usage-log.jsonl 단독으로 동작하며, manifest 부재는 에러가 아니다(EC-3).
// (lineage 상관은 향후 SPEC의 몫 — §C 범위 외.)
func LoadEvents(logPath string) ([]harness.Event, error) {
	f, err := os.Open(logPath) //nolint:gosec // logPath는 resolveProjectRoot 기반 내부 경로
	if err != nil {
		if os.IsNotExist(err) {
			// 파일 부재는 정상 상태 — 빈 슬라이스 + 성공 반환(REQ-OBL-004).
			return nil, nil
		}
		return nil, fmt.Errorf("cluster: usage-log 열기 실패 %s: %w", logPath, err)
	}
	defer func() { _ = f.Close() }()

	var events []harness.Event
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), scanBufferMax)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var evt harness.Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			// 손상된 줄은 건너뛴다(fail-open — 한 줄 때문에 중단하지 않음, REQ-OBL-003).
			continue
		}
		if evt.EventType != harness.EventTypeApplyOutcome {
			// 비-apply_outcome 이벤트는 클러스터링 대상이 아니다 — 건너뛴다.
			continue
		}
		events = append(events, evt)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cluster: usage-log 스캔 오류: %w", err)
	}

	return events, nil
}

// ClusterEvents 는 apply_outcome 이벤트 슬라이스를 결정론적 시그니처 키로 그룹화하여
// 정렬된 FailureCluster 슬라이스를 반환한다(REQ-OBL-005/006/007).
//
//   - kept(비-실패) 결과는 어떤 클러스터에도 포함되지 않는다(REQ-OBL-008, AC-OBL-003).
//   - 동일 입력에 대해 바이트 동일 출력을 보장한다: 클러스터는 Signature 로 안정
//     정렬되고, 각 클러스터의 멤버는 (Timestamp, ProposalID) 로 안정 정렬된다.
//     time.Now()/난수/맵-순회-순서가 출력에 누수되지 않는다(REQ-OBL-007).
func ClusterEvents(events []harness.Event) []FailureCluster {
	groups := make(map[string]*FailureCluster)

	for _, evt := range events {
		// 실패/롤백 verdict 만 클러스터링한다(REQ-OBL-008).
		if evt.OutcomeVerdict != verdictRolledBack {
			continue
		}

		key := signatureKey(evt)
		ref := EventRef{
			ProposalID: evt.OutcomeProposalID,
			Timestamp:  evt.Timestamp,
		}

		c, ok := groups[key]
		if !ok {
			c = &FailureCluster{
				Signature:                key,
				RepresentativeDimensions: normalizeDimensions(evt.OutcomeRegressed),
				Verdict:                  evt.OutcomeVerdict,
				Decision:                 evt.OutcomeDecision,
				FirstSeen:                evt.Timestamp,
				LastSeen:                 evt.Timestamp,
			}
			groups[key] = c
		}
		c.Count++
		c.Members = append(c.Members, ref)
		if evt.Timestamp.Before(c.FirstSeen) {
			c.FirstSeen = evt.Timestamp
		}
		if evt.Timestamp.After(c.LastSeen) {
			c.LastSeen = evt.Timestamp
		}
	}

	// 맵을 안정 정렬된 슬라이스로 변환한다(결정론적 순서, REQ-OBL-007).
	clusters := make([]FailureCluster, 0, len(groups))
	for _, c := range groups {
		sortMembers(c.Members)
		clusters = append(clusters, *c)
	}
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Signature < clusters[j].Signature
	})
	return clusters
}

// sortMembers 는 멤버 참조를 (Timestamp, ProposalID) 안정 순서로 정렬한다.
// 동일 타임스탬프 시 ProposalID 로 타이브레이크하여 결정론을 보장한다.
func sortMembers(members []EventRef) {
	sort.SliceStable(members, func(i, j int) bool {
		if !members[i].Timestamp.Equal(members[j].Timestamp) {
			return members[i].Timestamp.Before(members[j].Timestamp)
		}
		return members[i].ProposalID < members[j].ProposalID
	})
}

// Cluster 는 logPath 를 읽어 결정론적 클러스터 슬라이스를 반환하는 편의 진입점이다
// (LoadEvents → ClusterEvents 결합). CLI read 표면과 리포트 emitter 가 공유한다.
func Cluster(logPath string) ([]FailureCluster, error) {
	events, err := LoadEvents(logPath)
	if err != nil {
		return nil, err
	}
	return ClusterEvents(events), nil
}

// DefaultLogPath 는 projectRoot 기준 usage-log.jsonl 의 기본 상대 경로를 결합한다.
// (CLI 핸들러는 resolveProjectRoot(cmd) + 이 헬퍼로 입력 경로를 구성한다 — REQ-OBL-011.)
func DefaultLogPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "harness", "usage-log.jsonl")
}
