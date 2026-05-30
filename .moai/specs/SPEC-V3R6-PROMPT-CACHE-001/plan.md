# SPEC-V3R6-PROMPT-CACHE-001 — Implementation Plan

## Section A — Baseline (Estimated / Illustrative — 검증된 출처 없음)

> **[출처 정정 2026-05-30, v0.1.1]** 초안(v0.1.0)은 본 Section A를 `.moai/research/v3.0-design-2026-05-22.md` 및 `.moai/research/moai-adk-current-state-2026-05-22.md` 의 "verbatim 인용"으로 기술했으나, 두 파일은 **현재 디스크에 존재하지 않는다**(committed 흔적 없음). 따라서 아래 손익분기 수치는 verbatim 인용이 아니라 **공개된 Anthropic Prompt Caching 가격 정책 구조에 기반한 estimated / illustrative 산술 예시**로 재명시한다. 본 SPEC의 동기(motivation)일 뿐 acceptance 기준이 아니며, 실제 수치는 머지 후 K1-K5 KPI로 실측한다(spec.md § 1 동기).

본 SPEC는 Anthropic Prompt Caching을 MoAI-ADK 워크플로우에 도입하는 enablement layer다. 설계 골격(estimated):

- **1h cache write at session start**: CLAUDE.md + always-loaded rules + output style + MCP initial context을 단일 breakpoint로 묶음
- **5min cache for SPEC body**: `/moai run SPEC-XXX` 진입 시 spec.md + acceptance.md + plan.md 묶음

**가격 레버 (Anthropic 공개 정책 구조 — verify 대상)**:
- Cache Read: 입력 단가 대비 약 90% off
- Cache Write 5min: +25%
- Cache Write 1h: +100%

**illustrative 손익분기 (Sonnet 단가 가정, 일일 ~15 turn 가정 — 측정값 아님)**:
- No cache: `15 turn × turn당 단가`
- With 1h cache: `1회 write (+100%) + 14회 read (−90%)`
- 위 구조에서 일일 비용이 약 50% 안팎 절감될 수 있다는 추정. 구체 금액은 검증된 baseline이 없으므로 머지 후 실측한다.

cache_control 주입 위치는 `internal/cli/cc.go` SDK wrapper 진입점(또는 신규 `internal/runtime/cache_control.go`)이다. 현재 baseline은 cache_control 미사용, 0% hit rate.

본 SPEC은 `SPEC-V3R6-PROMPT-CACHE-001 Tier M`로 분류된다.

## Section B — Goal & KPIs

### Goal

cache hit rate ≥ 80% (7-day rolling)을 달성하여 종량제 Agent SDK 풀에서도 지속 가능한 토큰 이코노미를 확보한다. cache hit rate가 **유일한 측정 가능 1차 성공 지표(K1)**이며, 평균 turn 비용 절감은 estimated 방향성 목표일 뿐 hard acceptance 기준이 아니다(검증된 baseline 출처 없음 — Section A 참조).

### KPIs (post-merge 측정)

| KPI | Target | Measurement |
|-----|--------|-------------|
| **K1** (1차 지표): cache hit rate (7-day rolling) | ≥ 80% | `sum(cache_read_input_tokens) / (sum(cache_read_input_tokens) + sum(cache_creation_input_tokens))` from `.moai/state/cache-usage.jsonl` |
| **K2**: cache_creation log entries per session | ≥ 1 | grep `cache_creation_input_tokens` in JSONL, count per session_id |
| **K3**: cache_read log entries per session | ≥ N-1 (N = turn count) | grep `cache_read_input_tokens` in JSONL, count per session_id |
| **K4** (estimated 방향성 목표): 평균 turn 비용 절감률 | 약 50% 안팎 (illustrative, baseline 출처 없음 → 절대 금액 hard target 아님) | derived from K1 hit rate × Anthropic 공개 단가 — 머지 후 실측 |
| **K5**: 단일-turn 세션 비율 (cache penalty risk) | ≤ 10% | REQ-PC-007 warning log count / total sessions |

### Non-Goals

- GLM cache 호환성 확보 (Out of Scope, Sprint 3로 이연)
- Per-message cache breakpoint 최적화 (Out of Scope, 별도 SPEC)
- Background cache warming (Out of Scope, 후속 SPEC)

## Section C — Requirements & Traceability

| REQ | Pattern | Statement | AC Mapping |
|-----|---------|-----------|------------|
| REQ-PC-001 | Ubiquitous | Session start `ttl: "1h"` cache_control 주입 | AC-PC-001, AC-PC-003 |
| REQ-PC-002 | When | `/moai run SPEC-XXX` SPEC body `ttl: "5m"` cache_control 주입 | AC-PC-001 |
| REQ-PC-003 | Where | GLM 백엔드 시 cache_control omit | AC-PC-004 |
| REQ-PC-004 | Ubiquitous | PostToolUse hook이 cache token 필드 추출 + JSONL append | AC-PC-005, AC-PC-006 |
| REQ-PC-005 | When | `session_ttl: "off"` 시 session-level breakpoint 미주입 | AC-PC-002 |
| REQ-PC-006 | Ubiquitous | `moai doctor`가 7-day cache hit rate 표시 | AC-PC-007 |
| REQ-PC-007 | While | 단일-turn 세션 종료 시 cache penalty 경고 로그 | AC-PC-010 |

추가 ACs:
- AC-PC-008: Race-safe test suite green (`-race -count=1`)
- AC-PC-009: docs-site 4-locale 손익분기 문서화
- AC-PC-010: 단일-turn fixture에서 REQ-PC-007 warning 라인 + `session_ttl: "off"` 권고 문자열 검증 (M3 unit test)

100% traceability: 7 REQs → 10 ACs (모든 normative `shall`이 ≥1 AC로 매핑됨 — REQ-PC-007도 AC-PC-010으로 binary-testable; AC-PC-008/009는 cross-cutting 품질 게이트). D3 해소: REQ-PC-007은 `shall log <specific string> with <specific recommendation>` 형태이므로 합성 단일-turn fixture로 binary 검증 가능 (이전 "observational only" 라벨 철회).

## Section D — Milestones (Tier M, no Rounds)

Tier M SPEC 표준 lifecycle 5 milestones. Sequencing 순서 의무. 각 M은 manager-develop cycle_type=ddd Section A-E MANDATORY로 위임 대상.

### M1 — cache_control injection at SDK wrapper

**파일**:
- `internal/runtime/cache_control.go` (신규) 또는 `internal/cli/cc.go` (기존 확장 — M1 진입 시 결정)
- `internal/runtime/cache_control_test.go`

**동작**:
1. Claude Code SDK 호출 직전 outgoing request payload 검사
2. `cacheStrategy.enabled == true` AND `llm.mode != "glm"` 시
3. **threshold-agnostic fallback (R4 decoupled)**: cacheable 후보 payload의 토큰 수가 `cacheStrategy.min_cacheable_tokens` (default 2048) 미만이면 cache_control을 주입하지 않는다. 이 임계값은 `cache.yaml` 설정 상수이며 per-agent model 배정(SPEC-V3R6-AGENT-MODEL-ROUTING-001)에 의존하지 않는다 (spec.md § 6 R4 NOTE).
4. system prompt array의 LAST item에 `cache_control: {type: "ephemeral", ttl: "1h"}` 주입
5. messages array에서 SPEC body marker (예: `<spec-body>...</spec-body>` 또는 `/moai run` 시 자동 삽입된 SPEC bundle) 직후에 `ttl: "5m"` 주입

**REQ 매핑**: REQ-PC-001, REQ-PC-002, REQ-PC-003
**AC 매핑**: AC-PC-001 (grep ≥ 2), AC-PC-003 (integration test), AC-PC-004 (GLM omit)

**리스크**: Anthropic SDK 버전 의존성 (R3) — SDK pin + schema test로 완화. model-specific minimum cacheable tokens (R4) — `min_cacheable_tokens` 설정 상수로 self-contained 처리.

### M2 — cache.yaml config schema

**파일**:
- `.moai/config/sections/cache.yaml` (신규)
- `internal/config/cache_config.go` (신규) — 로더 + validator
- `internal/template/templates/.moai/config/sections/cache.yaml` (mirror)

**스키마**:
```yaml
cacheStrategy:
  enabled: true
  session_ttl: "1h"            # enum: "1h" | "5m" | "off"
  spec_ttl: "5m"               # enum: "5m" | "off"
  min_cacheable_tokens: 2048   # int; payload가 이 값 미만이면 cache_control 미주입 (R4 self-contained fallback)
```

**Validation**:
- `enabled: bool` (required)
- `session_ttl: enum["1h", "5m", "off"]` (default "1h")
- `spec_ttl: enum["5m", "off"]` (default "5m")
- `min_cacheable_tokens: int` (default 2048 — haiku 임계값 기준 보수 안전 상한; R4 decoupling — SPEC-V3R6-AGENT-MODEL-ROUTING-001 의존 제거. spec.md § 6 R4 NOTE 참조)
- Missing/invalid → log warning, fall back to `enabled: false` (safe default)

**REQ 매핑**: REQ-PC-005
**AC 매핑**: AC-PC-002 (config 키 검증 — `min_cacheable_tokens` 포함)

**R4 decoupling NOTE**: `min_cacheable_tokens`는 Anthropic 공개 정책상 모델별 최소 캐시 가능 토큰을 자체 처리하기 위한 self-contained 설정이다. per-agent model 배정(SPEC-V3R6-AGENT-MODEL-ROUTING-001)에 의존하지 않으므로, 해당 SPEC의 머지 여부와 무관하게 M1 fallback이 threshold-agnostic으로 동작한다.

**리스크**: Template mirror 동기화 누락 — `make build` + `commands_audit_test.go` 패턴으로 회귀 차단.

### M3 — PostToolUse telemetry hook

**파일**:
- `internal/hook/posttooluse_cache.go` (신규)
- `internal/hook/posttooluse_cache_test.go`
- `internal/state/cache_usage_log.go` (신규 JSONL writer)

**동작**:
1. PostToolUse 후크가 Anthropic API response에서 `usage.cache_creation_input_tokens` + `usage.cache_read_input_tokens` 추출
2. JSONL entry append to `.moai/state/cache-usage.jsonl`:
   ```json
   {"timestamp":"2026-05-23T10:15:30Z","session_id":"...","turn":3,"cache_creation":12450,"cache_read":48200,"model":"claude-sonnet-4-6"}
   ```
3. 단일-turn 세션 검출 시 (session_id 첫 등장 + turn=1만 존재 + wall-time < 5min) `WARN: single-turn cache write penalty risk` 추가 로그

**REQ 매핑**: REQ-PC-004, REQ-PC-007
**AC 매핑**: AC-PC-005 (JSONL append 검증), AC-PC-006 (2-turn 시 turn 2 cache_read 비zero), AC-PC-010 (단일-turn fixture warning 라인 + `session_ttl: "off"` 권고 검증 — D3 해소)

**리스크**: PostToolUse 후크 contract 변경 (HOOK-ASYNC-EXPAND-001과 충돌 가능) — 본 SPEC는 PostToolUse만 사용 (HOOK-OBSERVE-OPT-IN-001은 독립).

### M4 — moai doctor metric + REQ-PC-007 warning surfacing

**파일**:
- `internal/cli/doctor.go` (기존 확장)
- `internal/cli/doctor_cache_test.go` (신규)

**동작**:
1. `moai doctor` 출력에 `Cache hit rate (last 7 days): NN%` 라인 추가 (cacheStrategy.enabled 시)
2. 단일-turn 세션 비율 (`K5`) > 10% 시 `WARN: consider setting session_ttl: "off"` 출력
3. cache-usage.jsonl 파싱 (7-day window) + 집계 함수 단위 테스트 포함

**REQ 매핑**: REQ-PC-006
**AC 매핑**: AC-PC-007 (doctor 출력 grep)

**리스크**: jsonl 파싱 성능 (대량 entry 시) — M4 진입 시 streaming reader 적용.

### M5 — docs-site 4-locale mirror

**파일**:
- `docs-site/content/{en,ko,ja,zh}/cost-optimization/prompt-caching.md` (신규 4개)
- 또는 기존 cost-optimization 페이지 확장

**내용**:
1. 손익분기 룰 명문화: "1h cache는 세션당 2+ turn 발생 시에만 권장"
2. cache_control 메커니즘 설명 (Anthropic 공식 docs 인용 + WebFetch verify)
3. `session_ttl: "off"` opt-out 가이드
4. `moai doctor` cache 메트릭 해석 가이드

**REQ 매핑**: (전체 KPI 사용자 가시화)
**AC 매핑**: AC-PC-009 (4-locale parity ratio ≤ 1.20)

**리스크**: 4-locale 동기화 누락 — `.moai/docs/docs-site-i18n-rules.md` discipline 적용.

### Sequencing & Dependencies

```
M1 (cache_control inject)
  ↓ requires
M2 (cache.yaml config) — provides toggle for M1
  ↓ enables observability
M3 (PostToolUse telemetry)
  ↓ surfaces data
M4 (moai doctor metric)
  ↓ user-facing
M5 (docs-site 4-locale)
```

M1 + M2는 양방향 의존 (M1 동작이 M2 config 키 참조, M2는 M1 진입 시 동시 머지 가능). 실용적으로 M1 + M2를 single PR로 묶어 머지하는 것을 권장하나, manager-develop 위임 시 별도 milestone으로 분리하여 코드 리뷰 가시성 확보.

## Section E — Risks & Out of Scope

### Risks (전 spec.md § 6 동기)

| ID | Severity/Likelihood | Risk | Mitigation Milestone |
|----|---------------------|------|---------------------|
| R1 | Medium/High | 단일-turn 세션 1h cache_write +100% 페널티 | M3 (REQ-PC-007 warning) + M2 (`session_ttl: "off"` opt-out) |
| R2 | Medium/Medium | Cache prefix exact-match — rule churn 시 무효화 | Sprint 1 머지 완료로 사전 완화 (always-loaded baseline 안정) |
| R3 | Low/Medium | Anthropic SDK schema 변경 | M1 (SDK 버전 pin + AC-PC-003 schema test) |
| R4 | Low/Medium | model-specific minimum cacheable token (모델별 상이) — **decoupled from SPEC-V3R6-AGENT-MODEL-ROUTING-001** | M2 `cache.yaml` `min_cacheable_tokens` (보수 default 2048) self-contained 설정 + M1 threshold-agnostic fallback. AGENT-MODEL-ROUTING-001에 **의존하지 않음** (spec.md § 6 R4 NOTE). |
| R5 | Low/Low | cache-usage.jsonl 무한 성장 | 본 SPEC 범위 외 — 후속 telemetry rotation SPEC로 deferred |

### Out of Scope (3 h3 sections)

#### Out of Scope: Per-message cache breakpoint optimization

Auto-placing additional cache breakpoints inside long messages (예: per-paragraph)은 본 SPEC 범위 밖이다. 본 SPEC는 2개 고정 breakpoint (session start + SPEC body)만 설정한다. 향후 SPEC-V3R6-CACHE-GRANULAR-001로 분리.

#### Out of Scope: GLM cache compatibility

GLM (Z.AI) cache 메커니즘은 다른 control 필드를 사용한다. 본 SPEC는 Anthropic Claude SDK에만 적용된다. GLM cache 지원은 Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001에서 다룬다.

#### Out of Scope: Pre-emptive cache warming

사용자가 세션 진입 전 background process로 cache를 미리 워밍업하는 기능 (예: cron job)은 본 SPEC 범위 밖이다. 본 SPEC는 첫 user-triggered API call 시점에만 cache를 활성화한다.

## Section F — Validation Strategy

### Unit Tests

- `internal/runtime/cache_control_test.go`: cache_control 주입 위치 검증 (system LAST + SPEC body 직후) + `min_cacheable_tokens` 미만 payload omit 검증 (R4 threshold-agnostic fallback)
- `internal/config/cache_config_test.go`: cache.yaml 스키마 validation (`min_cacheable_tokens` 키 포함)
- `internal/hook/posttooluse_cache_test.go`: JSONL append + 단일-turn 검출 + AC-PC-010 (warning 라인 + `session_ttl: "off"` 권고 문자열 검증)
- `internal/cli/doctor_cache_test.go`: 7-day window 집계 + warning surfacing

### Integration Tests

- AC-PC-003: 합성 Anthropic API call payload 검증 (mock SDK)
- AC-PC-004: `llm.mode == "glm"` 시 cache_control 부재 검증
- AC-PC-006: 2-turn 합성 세션에서 turn 2 cache_read 비zero 검증

### CI Gates

- `go test ./internal/cli/... ./internal/runtime/... ./internal/hook/... -race -count=1` exit 0 (AC-PC-008)
- `golangci-lint run` zero new issues
- `make build` 후 template mirror 동기화 확인

### Manual Verification (post-merge)

- 본인 환경에서 `/moai run SPEC-V3R6-PROMPT-CACHE-001` 실제 실행 → cache-usage.jsonl 첫 entry 확인 (cache_creation > 0, turn 2부터 cache_read > 0)
- `moai doctor` 출력에 cache hit rate 라인 확인
- 7일 사용 후 hit rate ≥ 80% (KPI K1) 달성 검증

## Section G — Cross-References

- spec.md § 9 Cross-References (Sprint 2 sibling SPECs + Sprint 1 dependencies)
- `.claude/rules/moai/development/sprint-round-naming.md` (Sprint = multi-SPEC, Round = within-SPEC; Tier M은 Milestones M1-M5로 구성, Round 분할 불필요. "Wave" terminology는 AP-SRN-004로 retired)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical 12 fields)
- **출처 정정**: 초안이 인용한 `.moai/research/v3.0-design-2026-05-22.md` / `.moai/research/moai-adk-current-state-2026-05-22.md` 두 파일은 현재 디스크에 부재 (Section A + spec.md § 8 참조). 손익분기는 estimated/illustrative로 재명시됨.
- `CLAUDE.local.md` § 22 (settings.json local intent)
- PR #1046 머지 `134a43fac` (GEARS notation 의무)
