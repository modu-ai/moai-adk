---
id: SPEC-WF-AUDIT-GATE-001
version: "1.0.0"
status: draft
created_at: 2026-04-25
updated_at: 2026-04-25
author: GOOS
priority: High
labels: [workflow, plan-audit, gate, governance, dogfood]
issue_number: null
depends_on: []
related_specs: []
---

# SPEC-WF-AUDIT-GATE-001 Implementation Plan

> Plan→Run 전이 의무 감사 게이트 구현 로드맵
> 최종 갱신: 2026-04-25
> 상태: ready (자기 자신을 dogfood 대상으로 검증 예정)
> 종속 SPEC: 없음. 본 SPEC은 후속 game-keeper SPEC들이 의존할 수 있는 워크플로우 protocol foundation.

---

## 0. 계획 개요 (Executive Summary)

본 계획은 SPEC-WF-AUDIT-GATE-001을 6개 Phase로 분해한다. 핵심 산출물은 (a) `Phase 0.5: Plan Audit Gate` 워크플로우 skill 단락 신설(solo + team 모드 동시), (b) `--skip-audit` / `MOAI_SKIP_PLAN_AUDIT` 우회 경로 + 기록 보장, (c) `.moai/reports/plan-audit/` 보고서 디렉터리 + daily append, (d) 7일 grace window warn-only 모드, (e) plan-auditor 실패 시 INCONCLUSIVE fall-back, (f) Template-First 규율에 따른 `internal/template/templates/` 트윈 동기화 및 `make build`.

본 SPEC은 **워크플로우 protocol 변경**이며, **plan-auditor agent 내부 채점 로직은 변경하지 않는다** — 호출 지점과 차단 의사결정 로직만 신설한다. 본 SPEC은 dogfood 원칙에 따라 자기 자신이 첫 번째 게이트 통과 사례가 되어야 한다.

- Phase 수: **6** (A: 기반 디렉터리/skill skeleton → B: solo run.md → C: team run.md → D: skip-audit/INCONCLUSIVE → E: 통합 테스트 → F: 템플릿 동기화 + dogfood)
- 영향 받는 워크플로우 skill 파일: 3개 (`run.md`, `team/run.md`, `plan.md`)
- 영향 받는 rule 파일: 1개 (`spec-workflow.md`)
- 신설 디렉터리: 1개 (`.moai/reports/plan-audit/`)
- Template-First 트윈 파일: 위 4개 모두에 대해 `internal/template/templates/` 트윈 동기화 의무
- 핵심 의존: 기존 `plan-auditor` agent (`.claude/agents/moai/plan-auditor.md` — 변경 없음)

---

## 1. 구현 전략 (Implementation Strategy)

### 1.1 워크플로우 변경의 위치

게이트는 정확히 다음 지점에 삽입된다:

```
[/moai run SPEC-XXX 호출]
  ↓
Phase 0: Pre-flight checks (existing — SPEC 파일 존재 확인 등)
  ↓
[Phase 0.5: Plan Audit Gate]   ← NEW (본 SPEC)
  ├── (1) plan 산출물 hash 계산 (캐시 키)
  ├── (2) 24h cache 확인 → hit 시 verdict 채택, miss 시 (3)
  ├── (3) plan-auditor subagent 호출 (single invocation, main session)
  ├── (4) verdict 분기:
  │       ├── PASS → progress.md persist → Phase 1 진행
  │       ├── FAIL → (grace window 활성? → warn / 비활성? → block + AskUserQuestion)
  │       ├── BYPASSED (--skip-audit) → 보고서 BYPASSED 기록 → Phase 1 진행
  │       └── INCONCLUSIVE (timeout/error) → AskUserQuestion (retry/proceed/abort)
  └── (5) `.moai/reports/plan-audit/<SPEC>-<DATE>.md` append
  ↓
Phase 1: Implementation (existing — agent 위임 등)
  ↓
...
```

게이트는 **단일 진입점**이며 solo/team 모드 모두 동일 로직을 공유한다. team 모드에서는 main session에서 1회만 호출되고 verdict를 모든 teammate spawn 결정에 적용한다(REQ-WAG-005).

### 1.2 핵심 의사결정 원칙

1. **단일 호출 원칙**: plan-auditor는 `/moai run` 호출당 정확히 0~1회만 호출된다(0회 = 캐시 hit). team 모드여도 teammate 별 중복 호출 금지.
2. **자동 PASS 금지**: `REQ-WAG-007` — auditor 실패는 PASS와 다르다. INCONCLUSIVE는 사용자 결정으로만 해소.
3. **기록 의무**: 모든 게이트 호출(PASS/FAIL/BYPASSED/INCONCLUSIVE)이 일자 단위 보고서에 append된다.
4. **Template-First 규율**: skill 단락 변경은 `.claude/skills/`와 `internal/template/templates/.claude/skills/`에 동시 반영, `make build` 후 embedded 재생성 확인.
5. **Backward compatibility**: 기존 SPEC들은 7일 grace window 동안 warn-only 모드로 동작 → 사용자가 점진 보강 가능.

### 1.3 game-keeper 자세 (방어적 설계)

- 게이트 자체가 implementation을 차단하므로, 게이트의 버그가 전체 워크플로우를 마비시킬 수 있다 → INCONCLUSIVE fall-back은 **항상** 사용자에게 의사결정권을 이양한다.
- `--skip-audit`가 형해화 위험이 있으므로 모든 우회는 timestamped + user-tagged 보고서로 기록된다(`AC-WAG-06`).
- grace window는 **시간 의존 동작**이므로 시간 주입 가능한 설계(`MOAI_AUDIT_GATE_T0` env 또는 `time.Now` 추상화)로 테스트 가능성을 확보한다.

---

## 2. 영향 받는 파일 (Template-First per CLAUDE.local.md §2)

### 2.1 직접 수정 (실 작업 + 트윈 동시)

| 파일 | 수정 내용 | 트윈 경로 |
|------|----------|----------|
| `.claude/skills/moai/workflows/run.md` | `Phase 0.5: Plan Audit Gate` 단락 신설 (Phase 0과 Phase 1 사이) | `internal/template/templates/.claude/skills/moai/workflows/run.md` |
| `.claude/skills/moai/team/run.md` | 동일 게이트 단락 신설 (team 모드 parity, REQ-WAG-005) | `internal/template/templates/.claude/skills/moai/team/run.md` |
| `.claude/skills/moai/workflows/plan.md` | 종료 단락에 audit-ready 선언 출력 추가 (progress.md `plan_complete_at` 기록) | `internal/template/templates/.claude/skills/moai/workflows/plan.md` |
| `.claude/rules/moai/workflow/spec-workflow.md` | `## Phase Transitions` 의 "Plan to Run" 항목 보강 (Phase 0.5 게이트 명시), `## Phase 0.5: Plan Audit Gate` 단락 신설 | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` |

### 2.2 신설

| 파일/디렉터리 | 목적 |
|-------------|-----|
| `.moai/reports/plan-audit/.gitkeep` | 보고서 디렉터리 (REQ-WAG-004) |
| `internal/template/templates/.moai/reports/plan-audit/.gitkeep` | 트윈 |
| `.gitignore` (보강) | `.moai/reports/plan-audit/*.md` 추가 (보고서는 로컬 산출물; .gitkeep만 추적) |
| `internal/template/templates/.gitignore` (보강) | 트윈 |

### 2.3 명시적 비변경 (지킬 것)

- `.claude/agents/moai/plan-auditor.md` — plan-auditor 정의 자체는 변경 없음 (호출 지점만 신설)
- 기존 SPEC 35개 — retroactive audit 자동 실행 금지 (out-of-scope)
- `/moai plan` UI/CLI prompt — annotation cycle 그대로 (out-of-scope)

---

## 3. Phase 분해 (A → F)

### Phase A: 기반 — 디렉터리 + skill 단락 skeleton + 트윈 동기 + spec-workflow.md 보강

**목표**: 후속 Phase에서 채울 수 있는 빈 단락(skeleton)과 보고서 저장 디렉터리를 사전에 마련한다.

**구현 단계**:

1. `.moai/reports/plan-audit/.gitkeep` 생성 + 트윈
2. `.gitignore` 보강: `.moai/reports/plan-audit/*.md` 무시 + `.moai/reports/plan-audit/.gitkeep` 추적 유지
3. `.claude/skills/moai/workflows/run.md`에 "Phase 0.5: Plan Audit Gate (TBD)" placeholder 단락 추가 (Phase 0과 Phase 1 사이)
4. `.claude/skills/moai/team/run.md`에 동일 placeholder 추가
5. `.claude/skills/moai/workflows/plan.md` 종료에 `progress.md plan_complete_at` 기록 instruction 추가
6. `.claude/rules/moai/workflow/spec-workflow.md` "Phase Transitions" → "Plan to Run" 항목 보강 (게이트 추가 명시)
7. 트윈 4개 동시 변경
8. `make build && make install` 후 embedded 재생성 확인

**테스트 전략**: skill 파일은 정적 텍스트이므로 테스트 직접 검증 어려움 → `internal/template/commands_audit_test.go` 와 유사한 skill audit 테스트 신설(또는 기존 테스트의 path 패턴 추가).

**커밋 boundary**: A 완료 시점에서 단독 커밋 가능 (기능적 효과 0, 사후 Phase의 사전조건만 마련).

**롤백**: 마크다운 4개 파일 + .gitkeep + .gitignore 변경 revert.

---

### Phase B: solo `run.md` 게이트 본문 작성

**목표**: solo 모드 `/moai run`에서 Phase 0.5 게이트 단락의 실제 instruction을 작성한다.

**구현 단계**:

1. `.claude/skills/moai/workflows/run.md` Phase 0.5 단락 본문 작성:
   - sub-step (1) plan 산출물 hash 계산 (sha256 결합)
   - sub-step (2) 24h cache 확인 (`.moai/reports/plan-audit/<SPEC>-<DATE>.md` 존재 + verdict=PASS + age < 24h)
   - sub-step (3) plan-auditor subagent 호출 syntax 명시 ("Use the plan-auditor subagent to audit `.moai/specs/<SPEC>/...`")
   - sub-step (4) verdict 분기 매트릭스 (PASS/FAIL/BYPASSED/INCONCLUSIVE 4-way)
   - sub-step (5) progress.md persist + 보고서 append
2. grace window 활성 시 분기 명시: "If today < merge_date + 7days, treat FAIL as warning instead of blocking"
3. 사용자 결정 수집은 AskUserQuestion으로 위임 (orchestrator 책임 — agent에서 호출 금지)
4. 트윈 동기 + `make build`

**테스트 전략**: 통합 테스트 단계(Phase E)에서 실제 동작 검증. Phase B는 마크다운 작성 단계.

**롤백**: run.md 단일 파일 revert.

---

### Phase C: team `team/run.md` 게이트 본문 작성

**목표**: team 모드에서도 동일 게이트가 적용됨을 보장 (REQ-WAG-005, AC-WAG-05).

**구현 단계**:

1. `.claude/skills/moai/team/run.md` Phase 0.5 단락 추가 (Phase B와 동등 내용)
2. team-specific 차이 명시:
   - 게이트는 main session에서 1회만 호출 (TeamCreate 이전)
   - verdict=PASS 시에만 `Agent(subagent_type: "general-purpose")` 또는 `TeamCreate` 진행
   - verdict=FAIL → 어떤 teammate spawn도 차단
   - 캐시 hit 시에도 main session에서 단 1회 검사 (teammate들은 spawn 시 cache hit으로 확인)
3. 기존 team workflow의 "Phase 1 — Task Decomposition" 섹션 직전에 위치
4. 트윈 동기 + `make build`

**테스트 전략**: `internal/cli/team_run_audit_gate_test.go` (Phase E에서 작성).

**롤백**: team/run.md 단일 파일 revert.

---

### Phase D: `--skip-audit` 플래그 + INCONCLUSIVE fall-back 명세

**목표**: 우회 경로(REQ-WAG-006)와 실패 fall-back(REQ-WAG-007)을 워크플로우 skill 차원에서 정의한다.

**구현 단계**:

1. `.claude/skills/moai/workflows/run.md` 및 team 동등 파일에 다음 절 추가:
   - `### When --skip-audit Flag Is Provided` — bypass 동작, 보고서 기록 형식, 비대화형 환경 처리
   - `### When Plan-Auditor Fails or Times Out` — INCONCLUSIVE 분류, AskUserQuestion 3-way 옵션
2. 사용자 식별자 추출 경로 명시: `.moai/config/sections/user.yaml#user.name`
3. rationale 수집 방법 명시: AskUserQuestion (대화형) / `non-interactive` 자동 기록 (비대화형)
4. CLI 플래그 binding은 본 SPEC 범위 외 — 향후 SPEC에서 cobra flag 정의 예정 (현 단계는 워크플로우 skill 차원 합의)
5. 트윈 동기 + `make build`

**테스트 전략**: `internal/cli/run_audit_gate_integration_test.go` 의 다음 함수가 Phase E에서 작성됨:
- `TestSkipAuditFlagRecordsBypassWithUserRationale`
- `TestEnvVarSkipAuditEquivalentToFlag`
- `TestPlanAuditorFailureClassifiesAsInconclusive`

**롤백**: skill 파일 단락 revert.

---

### Phase E: 통합 테스트 작성 (TDD)

**목표**: AC-WAG-01 ~ 11 모두를 검증하는 통합 테스트 5개 파일을 RED → GREEN → REFACTOR 순으로 작성한다.

**구현 단계**:

1. **RED Phase** — 각 AC당 실패 테스트 1개 작성 (구현 미완 상태에서 fail 확인):
   - `internal/cli/run_audit_gate_integration_test.go` — AC-WAG-01, 02, 03, 06, 07
   - `internal/cli/run_audit_gate_grace_test.go` — AC-WAG-08
   - `internal/cli/run_audit_gate_cache_test.go` — AC-WAG-09
   - `internal/cli/run_audit_gate_filesystem_test.go` — AC-WAG-10
   - `internal/cli/team_run_audit_gate_test.go` — AC-WAG-05
   - `internal/cli/dogfood_self_audit_test.go` — AC-WAG-11
   - 각 테스트는 plan-auditor를 mock 또는 real call로 호출 가능 (build tag로 분리)

2. **GREEN Phase** — Phase B/C/D에서 작성한 워크플로우 skill 단락이 plan-auditor를 호출하도록 하는 minimal Go-side 통합 코드 작성:
   - `internal/runtime/audit_gate.go` — gate orchestration logic
   - `internal/runtime/audit_cache.go` — 24h cache + plan artifact hash
   - `internal/runtime/audit_report.go` — daily report append
   - 기존 `/moai run` 핸들러에서 audit_gate.Invoke() 호출

3. **REFACTOR Phase** — 중복 제거, gofmt, golangci-lint 통과
4. 통합 테스트 build tag: `-tags=integration`로 분리하여 일반 `go test ./...` 시 무거운 테스트 회피

**테스트 전략**:
- 단위 테스트: hash 계산, cache validity check, verdict 분기 로직
- 통합 테스트: plan-auditor mock harness 사용 (실제 agent 호출 없음, deterministic verdict 주입)
- 시간 의존 테스트: `MOAI_AUDIT_GATE_T0` 환경변수로 grace window 시점 주입

**롤백**: Go 코드 + 테스트 파일 단위 revert.

---

### Phase F: 템플릿 최종 동기화 + dogfood self-audit + grace window 개시

**목표**: Template-First 규율 최종 검증 + 본 SPEC 자체를 첫 게이트 통과 사례로 등록.

**구현 단계**:

1. 모든 영향 파일에 대해 `.claude/`와 `internal/template/templates/.claude/` byte-level 일치 확인
2. `make build && go test ./internal/template/...` — embedded 동기 검증
3. dogfood: `moai-adk-go` 자체에서 `/moai run SPEC-WF-AUDIT-GATE-001` 호출 (또는 plan-auditor를 mock harness로 본 SPEC에 대해 직접 호출)
4. verdict=PASS 확인 → `.moai/reports/plan-audit/SPEC-WF-AUDIT-GATE-001-2026-04-25.md` 생성
5. grace window 시작 timestamp 기록: `.moai/state/audit-gate-merge-at.txt` 또는 git tag (T0 기준점)
6. CHANGELOG.md 업데이트
7. spec.md status: `draft → implemented` 전환

**테스트 전략**: `dogfood_self_audit_test.go::TestSelfAuditPassesOnOwnSpec` 이 PASS 반환.

**롤백**: 본 SPEC 자체 status 복구 + CHANGELOG 라인 revert. 워크플로우 skill 변경은 유지.

---

## 4. 단계 간 의존성 그래프

```
Phase A (skeleton)
   ↓
Phase B (solo)  ──┐
Phase C (team)  ──┼──→ Phase D (skip/INCONCLUSIVE)
                  │       ↓
                  └──→ Phase E (TDD: tests + Go runtime)
                              ↓
                         Phase F (sync + dogfood)
```

- Phase B와 C는 영향 파일이 다르므로 병렬 가능 (그러나 Phase D 진입 전 둘 다 완료 필수).
- Phase E는 Go-side 구현을 동반하므로 가장 무거운 Phase. tasks.md에서 sub-task로 분해.
- Phase F는 Template-First + dogfood이므로 모든 선행 Phase 완료 후 실행.

---

## 5. 테스트 전략 (TDD)

### 5.1 단위 테스트

| 모듈 | 테스트 함수 (예시) |
|------|-------------------|
| `internal/runtime/audit_cache.go` | `TestPlanArtifactHashStableAcrossWhitespace`, `TestCacheTTLBoundary24Hours`, `TestCacheInvalidateOnHashChange` |
| `internal/runtime/audit_report.go` | `TestReportFilePathFormat`, `TestAppendMultipleAuditRunsSameDay`, `TestReportSecurityFilepathClean` |
| `internal/runtime/audit_gate.go` | `TestVerdictRouting4Way`, `TestGraceWindowBoundary7Days`, `TestSkipAuditFlagRecording` |

### 5.2 통합 테스트

| 파일 | AC 매핑 |
|------|---------|
| `run_audit_gate_integration_test.go` | AC-WAG-01, 02, 03, 06, 07 |
| `run_audit_gate_grace_test.go` | AC-WAG-08 |
| `run_audit_gate_cache_test.go` | AC-WAG-09 |
| `run_audit_gate_filesystem_test.go` | AC-WAG-10 |
| `team_run_audit_gate_test.go` | AC-WAG-05 |
| `dogfood_self_audit_test.go` | AC-WAG-11 |

빌드 태그: 모든 통합 테스트는 `//go:build integration` 으로 격리. CI에서 별도 job으로 실행.

### 5.3 plan-auditor 호출 추상화

테스트 가능성을 위해 plan-auditor 호출을 인터페이스로 추상화:

```
type PlanAuditor interface {
    Audit(ctx context.Context, specDir string) (Verdict, error)
}
```

Production 구현: `Agent(subagent_type: "plan-auditor")` 래퍼.
Test 구현: deterministic mock (verdict 주입 가능).

본 인터페이스 정의는 Phase E에서 단위 테스트와 함께 추가됨.

### 5.4 시간 의존 테스트

`time.Now`를 직접 사용하지 않고 `runtime/clock.go`의 `Clock` 인터페이스 경유. 테스트에서는 `FakeClock`으로 grace window 종료 시점, 캐시 만료 시점을 주입.

---

## 6. 마이그레이션 전략 (기존 SPEC 보호)

### 6.1 7일 grace window

본 SPEC merge 시점을 `T0`로 정의 (예: `2026-04-25 12:00 KST`). `T0 + 7일` 시점까지 게이트는 다음과 같이 동작:

- verdict=PASS → 정상 진행
- verdict=FAIL → **차단하지 않고** stdout 경고 출력 + progress.md `audit_verdict: FAIL_WARNED` 기록 + 일자 보고서에 FAIL 기록
- 매 호출 출력에 grace 잔여 일수 카운트다운 표기: `[grace-window] D-N (auto-block at T0+7)`

`T0 + 7일` 이후 자동으로 차단 모드로 전환. 사용자 추가 작업 불필요.

### 6.2 기존 35개 SPEC 처리

- retroactive 자동 audit 미실행 (out-of-scope)
- 사용자가 각 SPEC을 `/moai run` 호출하는 시점에 게이트 적용
- 기존 SPEC이 FAIL 시 사용자가 (a) SPEC 보강 후 재실행, (b) `--skip-audit` 우회, (c) abort 중 선택
- grace window 동안은 어떤 SPEC도 차단되지 않으므로 점진 보강 가능

### 6.3 retroactive audit 수동 트리거 (선택)

`moai constitution audit-all` (또는 동등) 커맨드는 본 SPEC 범위 외이며, 향후 별도 SPEC에서 정의. 본 SPEC merge 직후 사용자가 일괄 점검을 원할 경우 수동으로 각 SPEC에 대해 `/moai run --dry-run --audit-only`(미정의)를 호출하는 대신, 첫 진짜 `/moai run` 호출 시점에 자연스럽게 적용.

---

## 7. 백워드 호환성 (Backward Compatibility)

### 7.1 호환 보장

- 기존 SPEC 디렉터리 구조(`spec.md` + `plan.md` + `acceptance.md` + 선택 `tasks.md`) 그대로 사용 — 변경 없음
- plan-auditor agent 정의 변경 없음 — 기존 카탈로그 그대로
- `/moai run` CLI invocation 시그니처 변경 없음 (`--skip-audit` 추가만, 기존 미사용)
- `progress.md` 기존 필드 보존, 신규 필드(`audit_verdict`, `audit_report`, `audit_at`, `auditor_version`, `audit_cache_hit`, `cached_audit_at`, `inconclusive_acknowledged_by`) 추가만

### 7.2 호환 비보장 (사용자 영향)

- 본 SPEC merge 후 `/moai run` 평균 latency 증가 (audit 호출 시간; 캐시 hit 시 ≤ 2초, miss 시 ≤ 30초 추가)
- grace window 종료 후 audit FAIL SPEC은 명시적 사용자 결정 없이는 진행 불가 — 의도된 변경

### 7.3 환경변수 / 플래그 신설

- `--skip-audit` (CLI flag)
- `MOAI_SKIP_PLAN_AUDIT=1` (env)
- `MOAI_AUDIT_GATE_T0` (env, 테스트용 — production에서는 git tag 또는 `.moai/state/audit-gate-merge-at.txt` 사용)

---

## 8. 위험 및 완화 (재요약, spec.md §5 보강)

| Risk ID | Phase | 완화 작업 |
|---------|-------|----------|
| R-WAG-1 (대량 차단) | Phase A, F | 7일 grace window warn-only 모드 |
| R-WAG-2 (token 비용) | Phase E | 24h cache + plan artifact hash invalidation, 단일 호출 원칙 |
| R-WAG-3 (dogfood paradox) | Phase F | 본 SPEC을 첫 게이트 입력으로 자가 검증 (`AC-WAG-11`) |
| R-WAG-4 (사용자 거부권) | N/A | annotation cycle 그대로 — 게이트는 PASS 보장만 제공 |
| R-WAG-5 (warn 무시) | Phase A, B | grace 종료 카운트다운 매 호출 표기 |
| R-WAG-6 (`--skip-audit` 남용) | Phase D | 일자 보고서에 누적 기록 (post-MVP에서 30일 누적 경고) |
| R-WAG-7 (INCONCLUSIVE 일률 proceed) | Phase D | progress.md `inconclusive_acknowledged_by` 강제 기록 |

---

## 9. OPEN QUESTIONS

| ID | 질문 | 영향 Phase | 해소 시점 |
|----|------|-----------|----------|
| Q1 | plan 산출물 hash 알고리즘: 4 파일 SHA-256 결합 vs 정규화 후 결합 (whitespace insensitive)? | Phase E | `audit_cache.go` 구현 직전 |
| Q2 | grace window T0 기준점 저장 방식: git tag vs `.moai/state/` 파일 vs `merge_pr` frontmatter? | Phase F | dogfood 직전 |
| Q3 | INCONCLUSIVE에서 retry 옵션 선택 시 plan-auditor 재호출 횟수 제한 (3회?)? | Phase D | Phase D 작성 중 |
| Q4 | 비대화형 환경 자동 감지 방법 (`stdin` 닫힘 vs 환경변수 vs CLI flag `--non-interactive`)? | Phase D | Phase D 작성 중 |
| Q5 | team 모드에서 게이트 호출이 main session인지, leader teammate인지 명확화 (현 plan: main session 단독)? | Phase C | Phase C 작성 직전 |

---

## 10. 성공 지표 (Success Metrics)

1. dogfood: 본 SPEC을 자기 게이트에 입력 → PASS (`AC-WAG-11`)
2. v3R2 35 SPEC 사용자 점검 시: SPEC-V3R2-CON-001 (broken master-v3 anchor), SPEC-V3R2-WF-001 (48→24 arithmetic) 같은 결함이 게이트 호출 시 FAIL 또는 must-pass 위반으로 자동 검출됨을 sample 확인
3. token 예산: `/moai run` 캐시 hit 시 latency 증가 ≤ 2초
4. grace window 동작: T0+3일 시점 호출에서 FAIL 경고만, T0+8일 시점 호출에서 차단 동작
5. dogfood self-audit 보고서가 `.moai/reports/plan-audit/SPEC-WF-AUDIT-GATE-001-2026-04-25.md`로 정상 생성

---

## 11. 비목표 명시 (Non-Goals)

본 SPEC이 **하지 않는** 작업:

- plan-auditor 채점 알고리즘 변경
- 새 agent 정의
- `/moai plan` 호출 plot 변경
- 과거 SPEC 일괄 retroactive audit
- 게이트 우회 정책 자동 lock-out (남용 감지 후 자동 차단)
- audit 결과 시각화 UI/대시보드

이러한 항목은 별도 SPEC에서 다룬다.
