# MoAI-ADK v2.22.0 Release Notes — Embedding-Cluster Classifier (Tier-2 Pattern Aggregation Upgrade)

> Target release: v2.22.0
> Authoritative SPEC: `SPEC-V3R4-HARNESS-003`
> Prerequisite: `SPEC-V3R4-HARNESS-002` (Multi-Event Observer Expansion, v2.21.0)

---

## TL;DR

`usage-log.jsonl` 패턴 집계 레이어에 **Stage-2 SimHash 클러스터 분류기**를 추가합니다. 유사한 관측 이벤트를 Union-Find 알고리즘으로 자동 클러스터링하여 Tier-2(Heuristic) 진급에 필요한 신호를 더 빠르게 수렴시킵니다. `Stage2Enabled` 기본값은 `false` — 기존 Stage-1 byte-identical 경로는 완전 보존됩니다 (opt-in 업그레이드).

---

## 1. What's New

### SimHash 64-bit 지문 엔진 (Wave B)

- `SimHash64(features []string) uint64` — Charikar SimHash 알고리즘으로 feature 슬라이스를 64-bit 지문으로 변환합니다. FNV-1a 64-bit hash 기반, stdlib-only.
- `Hamming(a, b uint64) int` — `bits.OnesCount64(a ^ b)` O(1) Hamming 거리.
- `buildFeatureString(evt Event) []string` — 닫힌 허용 필드 목록: `Subject`, `PromptPreview`, `PromptLang`, `AgentName`, `AgentType`. **`PromptContent`는 PII 가드에 의해 명시적으로 제외됩니다** (REQ-HRN-CLS-014).
- `tokenize(s string)` — Unicode word-boundary split, lowercase, 스테밍 없음.

### Union-Find 클러스터링 (Wave C)

- `clusterSingletons(patterns, events, cfg, auditLogPath)` — Hamming threshold 내의 패턴을 Union-Find(path compression + union-by-rank)로 병합합니다.
- **EventType 파티셔닝**: 다른 EventType의 패턴은 절대 병합되지 않습니다.
- `merged_key` 형식: `"event_type:lex-min-subject"` (2-field, context_hash 제거).
- **append-only JSONL 감사 로그** (`cluster-merges.jsonl`): 각 병합 이벤트마다 ts / member_keys / member_counts / hamming_distances / hamming_pair_count / truncated / merged_key / merged_count / confidence 필드를 기록합니다.
- `hamming_distances` CAP=20: 100개 멤버 클러스터 시 C(100,2)=4950 쌍 중 첫 20개만 저장, `truncated=true`, `hamming_pair_count=4950` 기록 (EDGE-005).

### Config 로더 통합 (Wave D)

- `internal/config.LearningConfig` + `internal/config.ClassifierConfig` — `internal/harness.ClassifierConfig`와 동일한 YAML 태그. import cycle 방지를 위해 config 패키지에 독립 정의.
- `LoadHarnessConfig` — `harness.yaml`의 `learning.classifier` 블록을 파싱. `yaml.TypeError` 발생 시 `ClassifierConfig{}.WithDefaults()`로 fallback (REQ-HRN-CLS-018).
- `.moai/config/sections/harness.yaml`에 `learning.classifier` 블록 추가 (기본값: `stage_2_enabled: false`, simhash/3/3).

---

## 2. Performance

| 시나리오 | ns/op (Apple M4 Max) | 결과 |
|---------|----------------------|------|
| 1k events, Stage-2 ON | ~6,614,656 ns (~6.6ms) | REQ-HRN-CLS-015 1초 이내 PASS |
| 1k events, Stage-2 OFF | ~47,262 ns (~47µs) | baseline |

Stage-2 On/Off 비율: ~140x. Union-Find 알고리즘 상한은 O(n · α(n)) ≈ O(n) (역-Ackermann 함수).

---

## 3. Coverage & Quality

| 패키지 | 커버리지 | 비고 |
|--------|---------|------|
| `internal/harness` | 88.3% | 목표 85% 초과 |
| `internal/harness/safety` | 94.3% | |
| `internal/config` | 전체 PASS | |

- golangci-lint: 0 issues (모든 4개 Wave)
- 외부 의존성 추가: 0건 (stdlib-only)
- 고루틴/채널: 0건 (`[HARD]` 제약 준수)

---

## 4. PII Guard

`PromptContent` JSON 필드 (`prompt_content`)는 다음 두 레이어에서 보호됩니다:

1. **빌드 타임**: `buildFeatureString`의 닫힌 허용 목록에 포함되지 않습니다.
2. **런타임**: `TestPIIGuard_ClusterAuditLogNoPromptContent`가 `cluster-merges.jsonl`에 `PromptContent` 값이 나타나지 않는지 end-to-end 검증합니다 (AC-HRN-CLS-009).
3. **코드 레벨**: `classifier_*.go` 프로덕션 파일에 `PromptContent` 리터럴이 없습니다 (`TestPIIGuard_PromptContentExcluded` grep 검증).

---

## 5. Breaking Changes

없음. `Stage2Enabled` 기본값 `false` — 기존 동작 완전 보존.

## 6. Migration Guide

### Stage-2 활성화 (opt-in)

`.moai/config/sections/harness.yaml`에서 다음을 변경하세요:

```yaml
# 변경 전 (Stage-1 only)
learning:
  classifier:
    stage_2_enabled: false

# 변경 후 (Stage-2 활성)
learning:
  classifier:
    stage_2_enabled: true
    similarity_algorithm: simhash
    hamming_threshold: 3  # 낮을수록 더 엄격 (0-64)
    cluster_min_size: 3   # 최소 클러스터 크기 (>= 2)
```

변경 후 재시작하면 다음 `AggregatePatterns` 호출부터 Stage-2 경로가 활성화됩니다.

### cluster-merges.jsonl 감사 로그

Stage-2 활성 시 `usage-log.jsonl`과 동일한 디렉토리에 `cluster-merges.jsonl`이 자동 생성됩니다. 이 파일을 삭제해도 데이터 손실 없습니다 — 다음 집계 시 재생성됩니다.

---

## 7. Acceptance Criteria Status

| AC ID | 설명 | 상태 |
|-------|------|------|
| AC-HRN-CLS-001 | Stage-2 비활성 시 Stage-1 경로 그대로 | PASS |
| AC-HRN-CLS-002 | 유사 10개 singleton → 1개 merged 패턴 | PASS |
| AC-HRN-CLS-003 | 비유사 10개 singleton → 10개 분리 패턴 | PASS |
| AC-HRN-CLS-004 | merged 패턴 confidence = mean(멤버 confidence) | PASS |
| AC-HRN-CLS-005 | schema_version 필드 라운드트립 보존 | PASS |
| AC-HRN-CLS-006 | confidence < 0.70 → TierObservation 강제 | PASS |
| AC-HRN-CLS-007 | confidence < 0.70 → tier 진급 차단 | PASS |
| AC-HRN-CLS-008 | cluster-merges.jsonl 감사 로그 스키마 검증 | PASS |
| AC-HRN-CLS-009 | cluster-merges.jsonl에 PromptContent 미포함 | PASS |
| AC-HRN-CLS-010 | EventType 파티셔닝 — 다른 타입 병합 불가 | PASS (Union-Find partitioning 구현) |
| AC-HRN-CLS-011 | hamming_distances CAP=20, truncated 필드 | PASS |
| AC-HRN-CLS-012 | AggregatePatterns 시그니처 동결 | PASS |
| AC-HRN-CLS-013 | TierThresholds config override | PASS |
| AC-HRN-CLS-016 | harness.yaml learning.classifier 로딩 | PASS |
| AC-HRN-CLS-018 | yaml.TypeError → WithDefaults() fallback | PASS |

---

## 8. Files Changed

### New Files

| 파일 | 설명 |
|------|------|
| `internal/harness/classifier_simhash.go` | SimHash64, Hamming, tokenize, buildFeatureString |
| `internal/harness/classifier_simhash_test.go` | Wave B 단위 테스트 |
| `internal/harness/classifier_pii_test.go` | PII guard + AC-HRN-CLS-009 end-to-end |
| `internal/harness/classifier_cluster_test.go` | Wave C 통합 테스트 + T-D5 fallback |
| `internal/harness/classifier_cluster_audit_test.go` | 감사 로그 스키마 + EDGE-005 |
| `internal/harness/classifier_cluster_bench_test.go` | BenchmarkClusterSingletons1k |
| `internal/harness/classifier_schema_regression_test.go` | 스키마 회귀 가드 |
| `internal/harness/classifier_frozen_guard_regression_test.go` | EC-A5/EC-A6 회귀 가드 |
| `internal/harness/classifier_rate_limit_test.go` | AC-HRN-CLS-006 rate limit 가드 |
| `internal/config/harness_classifier_config_test.go` | config 로더 테스트 (REQ-HRN-CLS-016/018) |
| `internal/harness/testdata/stage2_similar_10.jsonl` | Wave C 유사 이벤트 fixture |
| `internal/harness/testdata/stage2_dissimilar_10.jsonl` | Wave C 비유사 이벤트 fixture |
| `internal/harness/testdata/stage2_perf_1k.jsonl` | Wave D 1k 성능 fixture |
| `internal/harness/testdata/legacy_promotions.jsonl` | Wave D legacy 호환성 fixture |

### Modified Files

| 파일 | 변경 내용 |
|------|----------|
| `internal/harness/classifier_cluster.go` | Wave A stub → Wave C 전체 구현 (Union-Find + 감사 로그) |
| `internal/harness/learner.go` | collectedEvents 수집 + clusterSingletons 호출 업데이트 |
| `internal/config/types.go` | LearningConfig, ClassifierConfig, WithDefaults() 추가 |
| `internal/config/loader.go` | yaml.TypeError fallback + isYAMLTypeError |
| `.moai/config/sections/harness.yaml` | learning.classifier 블록 추가 |

---

Version: v2.22.0
SPEC: SPEC-V3R4-HARNESS-003
Delivery: 4-wave TDD (Wave A plan `d32bb674a` + Wave B `b358554de` + Wave C `1c1ac039c` + Wave D `ed45b7c30` + Wave E)
