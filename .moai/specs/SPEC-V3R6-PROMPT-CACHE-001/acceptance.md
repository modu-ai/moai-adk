# SPEC-V3R6-PROMPT-CACHE-001 — Acceptance Criteria

## Overview

10 binary acceptance criteria (AC-PC-001..010). 모든 AC는 `bash` 명령 또는 코드 실행으로 검증 가능한 binary outcome (PASS / FAIL). 100% traceability with `spec.md` § 4 REQs (7 REQs ↔ 10 ACs — 모든 normative `shall`이 ≥1 AC로 매핑됨. REQ-PC-007은 AC-PC-010으로 binary-testable이며, AC-PC-008/009는 cross-cutting 품질 게이트).

## AC-PC-001 — cache_control 주입 호출부 식별

**REQ 매핑**: REQ-PC-001, REQ-PC-002
**Severity**: Blocking

**Verification command**:
```bash
grep -rn 'cache_control' internal/cli/ internal/runtime/ | grep -v _test.go | wc -l
```

**Expected**: 결과 ≥ `2` (최소 1회 1h session, 1회 5m SPEC body).

**PASS criterion**: stdout 값 (정수)이 2 이상.

**Rationale**: REQ-PC-001 + REQ-PC-002 두 주입 지점이 코드베이스 내 실재해야 한다.

## AC-PC-002 — cache.yaml 스키마 키 검증

**REQ 매핑**: REQ-PC-005
**Severity**: Blocking

**Verification command**:
```bash
# cache.yaml 또는 확장된 runtime.yaml 중 하나에 존재
grep -E '^\s*(cacheStrategy|enabled|session_ttl|spec_ttl):' \
  .moai/config/sections/cache.yaml .moai/config/sections/runtime.yaml 2>/dev/null \
  | grep -cE '(cacheStrategy|enabled|session_ttl|spec_ttl)'
```

**Expected**: 결과 ≥ `4` (cacheStrategy + 3 하위 키).

**PASS criterion**: stdout 값 ≥ 4.

**Rationale**: M2 milestone 산출물 검증.

## AC-PC-003 — cache_control schema integration test

**REQ 매핑**: REQ-PC-001, REQ-PC-003
**Severity**: Blocking

**Verification command**:
```bash
go test -run 'TestCacheControl_SessionStart_OneHour|TestCacheControl_AnthropicPayloadSchema' \
  ./internal/cli/... ./internal/runtime/... -count=1 -v 2>&1 | tail -20
```

**Expected**: 모든 매칭 테스트 `PASS`, exit 0.

**PASS criterion**:
1. Test output에 `--- PASS:` 항목 ≥ 1
2. `--- FAIL:` 항목 0
3. exit code 0

**Rationale**: 합성 Anthropic API request payload가 `cache_control: {type: "ephemeral", ttl: "1h"}` 필드를 system prompt LAST item에 포함해야 한다.

## AC-PC-004 — GLM 백엔드 시 cache_control omit

**REQ 매핑**: REQ-PC-003
**Severity**: Blocking

**Verification command**:
```bash
go test -run 'TestCacheControl_GLMMode_NoInjection' \
  ./internal/cli/... ./internal/runtime/... -count=1 -v 2>&1 | tail -10
```

**Expected**: `TestCacheControl_GLMMode_NoInjection PASS`.

**PASS criterion**:
1. `--- PASS: TestCacheControl_GLMMode_NoInjection` 출현
2. `llm.mode = "glm"` 설정 시 outgoing payload에 `cache_control` 필드 부재

**Rationale**: Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001 분리 boundary 준수.

## AC-PC-005 — PostToolUse hook JSONL append

**REQ 매핑**: REQ-PC-004
**Severity**: Blocking

**Verification command**:
```bash
go test -run 'TestPostToolUseCache_JSONLAppend' \
  ./internal/hook/... -count=1 -v 2>&1 | tail -10
```

**Expected**: `TestPostToolUseCache_JSONLAppend PASS`.

**PASS criterion**:
1. `--- PASS:` 출현
2. Test fixture로 합성 API response 처리 후, 출력 JSONL entry에 `cache_creation_input_tokens` AND `cache_read_input_tokens` 두 키 모두 존재

**Rationale**: REQ-PC-004 telemetry hook의 양쪽 필드 추출 검증.

## AC-PC-006 — 2-turn 세션 cache hit 검증

**REQ 매핑**: REQ-PC-004
**Severity**: Blocking

**Verification command**:
```bash
go test -run 'TestCacheUsage_TwoTurnSession_Turn2HitsCache' \
  ./internal/hook/... ./internal/state/... -count=1 -v 2>&1 | tail -10
```

**Expected**: `TestCacheUsage_TwoTurnSession_Turn2HitsCache PASS`.

**PASS criterion**:
1. 합성 2-turn 세션 fixture 실행
2. Turn 1 JSONL entry: `cache_creation_input_tokens > 0`, `cache_read_input_tokens == 0`
3. Turn 2 JSONL entry: `cache_read_input_tokens > 0` (cache hit verified)

**Rationale**: 실제 cache hit semantic 검증. 단순 schema가 아닌 동작 검증.

## AC-PC-007 — moai doctor cache hit rate 출력

**REQ 매핑**: REQ-PC-006
**Severity**: Blocking

**Verification command**:
```bash
# 합성 7-day window fixture 준비 후
moai doctor 2>&1 | grep -E 'Cache hit rate.*[0-9]+%'
```

**Expected**: `Cache hit rate (last 7 days): NN%` 패턴 매칭 결과 1행 이상.

**PASS criterion**:
1. grep 매칭 결과 line 수 ≥ 1
2. 매칭 라인에 정수 % 값 포함 (0-100)
3. `cacheStrategy.enabled: true` 일 때만 표시 (false 시 라인 부재가 정상)

**Rationale**: M4 사용자 가시화 검증.

## AC-PC-008 — race-safe full test suite

**REQ 매핑**: 전체 (cross-cutting quality gate)
**Severity**: Blocking

**Verification command**:
```bash
go test ./internal/cli/... ./internal/runtime/... ./internal/hook/... ./internal/state/... ./internal/config/... \
  -race -count=1 2>&1 | tail -20
```

**Expected**: exit 0, 전체 PASS, race detector clean.

**PASS criterion**:
1. exit code 0
2. `ok` 라인만 출력 (`FAIL` 없음)
3. `DATA RACE` 출현 0회
4. `--count=1`로 캐시 무효화 (Flaky 회피)

**Rationale**: 동시성 안전성 + 회귀 차단. CLAUDE.local.md § 6 [HARD] 규정 준수.

## AC-PC-009 — docs-site 4-locale 손익분기 문서화

**REQ 매핑**: 전체 (사용자 가시화)
**Severity**: Should-fix

**Verification command**:
```bash
# 4-locale 모두 손익분기 문서 존재 + parity ratio 검증
for locale in en ko ja zh; do
  file="docs-site/content/${locale}/cost-optimization/prompt-caching.md"
  if [ ! -f "$file" ]; then
    echo "MISSING: $file"
    exit 1
  fi
  wc -w "$file"
done
```

**Expected**: 4개 파일 모두 존재. wc -w 결과 4개 locale 워드카운트 출력. parity ratio (max/min) ≤ 1.20.

**PASS criterion**:
1. 4개 locale 파일 모두 존재 (MISSING 출력 부재)
2. 각 파일 워드카운트 > 100 (실질 내용 보유)
3. max(wordcount) / min(wordcount) ≤ 1.20 (4-locale 균형)

**Rationale**: `.moai/docs/docs-site-i18n-rules.md` 4-locale discipline. KPI 사용자 이해를 위한 손익분기 가이드 필수.

## AC-PC-010 — 단일-turn cache penalty 경고 로그 검증

**REQ 매핑**: REQ-PC-007
**Severity**: Blocking

**Verification command**:
```bash
go test -run 'TestPostToolUseCache_SingleTurnSession_PenaltyWarning' \
  ./internal/hook/... -count=1 -v 2>&1 | tail -15
```

**Expected**: `TestPostToolUseCache_SingleTurnSession_PenaltyWarning PASS`.

**PASS criterion**:
1. `--- PASS: TestPostToolUseCache_SingleTurnSession_PenaltyWarning` 출현, `--- FAIL:` 0, exit 0
2. 합성 단일-turn 세션 fixture (turn=1 only AND elapsed wall-time < 5min) 처리 시, 로그 출력에 `single-turn cache write penalty risk` 경고 문자열이 존재
3. 동일 경고 라인 또는 인접 권고에 `session_ttl: "off"` 권고 문자열이 존재 (REQ-PC-007의 "concrete recommendation" 검증)
4. (negative case) 2-turn 이상 fixture 처리 시 해당 경고 라인 **부재** (false-positive 회피)

**Rationale**: REQ-PC-007은 `shall log <specific string> with <specific recommendation>` 형태의 normative 요구이므로 binary 검증 가능하다 (이전 "observational only" 라벨 철회 — D3 해소). 검출 로직은 plan.md M3 §3에 명시되어 있어 unit test가 feasible하다.

## Traceability Matrix

| AC | REQ | Milestone | Severity |
|----|-----|-----------|----------|
| AC-PC-001 | REQ-PC-001, REQ-PC-002 | M1 | Blocking |
| AC-PC-002 | REQ-PC-005 | M2 | Blocking |
| AC-PC-003 | REQ-PC-001, REQ-PC-003 | M1 | Blocking |
| AC-PC-004 | REQ-PC-003 | M1 | Blocking |
| AC-PC-005 | REQ-PC-004 | M3 | Blocking |
| AC-PC-006 | REQ-PC-004 | M3 | Blocking |
| AC-PC-007 | REQ-PC-006 | M4 | Blocking |
| AC-PC-008 | (cross-cutting) | M1-M4 | Blocking |
| AC-PC-009 | (cross-cutting) | M5 | Should-fix |
| AC-PC-010 | REQ-PC-007 | M3 | Blocking |

모든 normative `shall`(REQ-PC-001..007)이 ≥1 binary AC로 매핑되었다 (D3 해소 — REQ-PC-007의 이전 "observational only, AC 미설정" 라벨은 철회되었고 AC-PC-010으로 대체). AC-PC-010은 합성 단일-turn fixture 기반 unit test로 경고 문자열 + `session_ttl: "off"` 권고를 binary 검증한다. M4 doctor 출력의 단일-turn 비율 surfacing(K5)은 AC-PC-007과 동일 doctor unit test에서 추가 확인한다.

## Definition of Done

본 SPEC가 "implemented" 상태로 전환되려면:

1. **AC-PC-001 ~ AC-PC-008 + AC-PC-010 모두 PASS** (Blocking 9개)
2. **AC-PC-009 PASS** 또는 **Should-fix exception 문서화** (i18n discipline)
3. **manager-develop progress.md** 작성 (M1~M5 evidence + AC PASS 증거 보존)
4. **plan-auditor 통과** (Tier M threshold 0.80)
5. **CI green**: `go test ./... -race`, `golangci-lint run`, `make build` 모두 exit 0
6. **머지 후 7일 K1 측정**: 실제 hit rate ≥ 80% 확인 (실패 시 follow-up SPEC)

## Out of Scope (Acceptance level)

### Out of Scope: Per-locale 손익분기 수치 차등

각 locale별 (ko vs en) 손익분기 수치를 다르게 표시하는 것은 본 SPEC 범위 밖이다. AC-PC-009는 4-locale 균형 (parity ratio)만 검증한다.

### Out of Scope: 실시간 cache hit rate dashboard

웹 dashboard, Prometheus exporter, OpenTelemetry metric 송출은 본 SPEC 범위 밖이다. JSONL append + `moai doctor` CLI만 본 SPEC 범위에 포함된다.

### Out of Scope: cache_control 동적 TTL 결정

세션 길이를 예측해 1h / 5m / off를 동적으로 선택하는 ML 기반 결정 로직은 본 SPEC 범위 밖이다. 본 SPEC는 config 기반 정적 결정만 다룬다.
