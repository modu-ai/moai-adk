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

# SPEC-WF-AUDIT-GATE-001 Acceptance Criteria

> Given-When-Then 인수 조건 + 추적성 + 증거 매핑
> 최종 갱신: 2026-04-25
> AC 개수: **11** (REQ §3 의 7개 1:1 + 추론된 엣지 케이스 4개)
> 출처: spec.md §3 EARS Requirements + 본 문서에서 파생된 AC-08 ~ AC-11

---

## 읽는 방법

- 각 AC는 고유 ID (`AC-WAG-NN`), REQ 추적성, Given-When-Then-Evidence 4단 구조를 따른다.
- "Evidence"는 통합 테스트 함수, 로그 라인 패턴, 또는 파일 산출물 경로 중 하나 이상을 명시한다.
- `Source`: REQ 직접 매핑(SPEC §3 명시) 또는 `INFERRED`(엣지 케이스, 본 문서에서 파생).
- 모든 AC는 객관적으로 검증 가능해야 하며, 주관 평가("적절히", "잘") 표현을 금지한다.

---

## Part A: REQ 직접 매핑 AC (7건, REQ-WAG-001 ~ 007 1:1)

### `AC-WAG-01` — Plan-Auditor 자동 호출 (게이트 진입)
- **Source**: REQ-WAG-001 (Ubiquitous)
- **Traceability**: `REQ-WAG-001`
- **Given**: `.moai/specs/SPEC-DUMMY-001/{spec,plan,acceptance}.md`가 존재하고, `--skip-audit` 플래그 미전달, `MOAI_SKIP_PLAN_AUDIT` 환경변수 미설정 상태.
- **When**: 사용자가 `/moai run SPEC-DUMMY-001` 을 호출한다.
- **Then**: 어떠한 implementation phase 작업(파일 작성, 테스트 작성, agent 위임)이 시작되기 전에 `plan-auditor` 서브에이전트가 정확히 1회 호출되어야 하며, 그 호출은 run.md Phase 0 직후, Phase 1 진입 직전에 위치해야 한다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_integration_test.go`
  - 테스트 함수: `TestRunInvokesPlanAuditorBeforeImplementation`
  - 로그 라인 패턴: `[plan-audit] invoking plan-auditor for SPEC-DUMMY-001 (gate=mandatory)`
  - 파일 산출물: `.moai/reports/plan-audit/SPEC-DUMMY-001-<YYYY-MM-DD>.md` 생성 확인
  - 음성 검증: 로그에 `[implementation]` 라인이 `[plan-audit]` 라인보다 먼저 등장하지 않음

### `AC-WAG-02` — FAIL 결과 시 Run 차단
- **Source**: REQ-WAG-002 (Event-driven)
- **Traceability**: `REQ-WAG-002`
- **Given**: `SPEC-FAIL-001`이 plan-auditor의 must-pass 4개 중 1개 이상을 위반한다(예: spec.md에 EARS 요구사항 0건). grace window는 종료 상태(차단 모드 활성).
- **When**: 사용자가 `/moai run SPEC-FAIL-001` 을 호출한다.
- **Then**: 시스템은 (a) Phase 1 implementation을 시작하지 않고, (b) audit report 경로를 stdout에 출력하며, (c) AskUserQuestion으로 사용자 의사결정(revise / override / abort)을 수집한다. 사용자 명시 응답 전에는 어떤 자동 진행도 발생하지 않는다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_integration_test.go`
  - 테스트 함수: `TestRunBlockedOnAuditFail`
  - 로그 라인 패턴: `[plan-audit] verdict=FAIL, blocking Run phase, report=<path>`
  - 파일 산출물: `.moai/reports/plan-audit/SPEC-FAIL-001-<YYYY-MM-DD>.md` (verdict: FAIL 명시)
  - 사용자 결정 missing 시 exit code: 비제로 또는 명시적 wait 상태

### `AC-WAG-03` — PASS 결과 시 진행 + persist
- **Source**: REQ-WAG-003 (Event-driven)
- **Traceability**: `REQ-WAG-003`
- **Given**: `SPEC-PASS-001`이 모든 must-pass와 scored 임계점을 통과한다.
- **When**: 사용자가 `/moai run SPEC-PASS-001` 을 호출한다.
- **Then**: 시스템은 (a) `.moai/specs/SPEC-PASS-001/progress.md`에 `audit_verdict: PASS`, `audit_report: <path>`, `audit_at: <ISO-8601>`, `auditor_version: <plan-auditor identifier>` 4 필드를 append하고, (b) Phase 1 implementation을 진행한다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_integration_test.go`
  - 테스트 함수: `TestRunProceedsOnAuditPassAndPersistsVerdict`
  - 파일 산출물: `progress.md`에 4개 필드 모두 기록 (grep 검증)
  - 로그 라인 패턴: `[plan-audit] verdict=PASS, persisted to progress.md, proceeding to Phase 1`

### `AC-WAG-04` — 일자 단위 보고서 보존
- **Source**: REQ-WAG-004 (Ubiquitous)
- **Traceability**: `REQ-WAG-004`
- **Given**: 동일 일자(`2026-04-25`)에 동일 SPEC(`SPEC-MULTI-001`)에 대해 게이트가 3회 호출된다(예: 1회 PASS, 1회 FAIL, 1회 BYPASSED).
- **When**: 3회 호출이 모두 완료된 시점에 `.moai/reports/plan-audit/SPEC-MULTI-001-2026-04-25.md`를 읽는다.
- **Then**: 단일 파일에 3개 audit run이 시간순(ISO-8601 timestamp 순)으로 append되어 있으며, 각 run은 verdict, report path, timestamp, run trigger(`automatic`/`manual`/`bypassed`/`inconclusive`)를 포함한다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_integration_test.go`
  - 테스트 함수: `TestDailyAuditReportAccumulatesMultipleRuns`
  - 파일 산출물 시그니처: `## Audit Run 1 of 3` ... `## Audit Run 3 of 3` 헤더 3개 모두 존재
  - 부가 검증: `wc -l` 결과 단일 run보다 큼 (append 확인)

### `AC-WAG-05` — Team 모드 게이트 동등 적용
- **Source**: REQ-WAG-005 (State-driven)
- **Traceability**: `REQ-WAG-005`
- **Given**: `workflow.team.enabled: true`이고 `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`인 환경에서 `SPEC-TEAM-001`(plan-auditor must-pass FAIL 케이스)이 준비됨.
- **When**: 사용자가 `/moai run SPEC-TEAM-001 --team` 을 호출한다.
- **Then**: (a) main session에서 plan-auditor가 1회 호출되고, (b) `TeamCreate` 또는 `Agent(subagent_type: "general-purpose")` 호출이 audit verdict=FAIL 시점에 차단되며, (c) 어떤 teammate(implementer/tester/designer/reviewer)도 spawn되지 않는다.
- **Evidence**:
  - 통합 테스트: `internal/cli/team_run_audit_gate_test.go`
  - 테스트 함수: `TestTeamRunBlockedBeforeTeammateSpawn`
  - 로그 라인 패턴: `[plan-audit] team mode detected, gate applies before TeamCreate`
  - 음성 검증: `tmux list-panes` 결과에 신규 teammate pane 미생성

### `AC-WAG-06` — `--skip-audit` 우회 기록
- **Source**: REQ-WAG-006 (Optional Feature)
- **Traceability**: `REQ-WAG-006`
- **Given**: 사용자가 `/moai run SPEC-BYPASS-001 --skip-audit` 을 대화형 환경에서 호출하고, AskUserQuestion 응답으로 우회 사유 "demo for ICSE 2026 deadline"을 제출한다.
- **When**: 명령이 완료된다.
- **Then**: `.moai/reports/plan-audit/SPEC-BYPASS-001-<YYYY-MM-DD>.md`에 `verdict: BYPASSED`, `bypass_at: <ISO-8601>`, `bypass_user: GOOS행님` (user.yaml#user.name 참조), `bypass_reason: "demo for ICSE 2026 deadline"` 4 필드가 모두 기록된다. 비대화형(`stdin` 닫힘) 환경에서는 `bypass_reason: "non-interactive"`가 자동 기록된다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_integration_test.go`
  - 테스트 함수 (대화형): `TestSkipAuditFlagRecordsBypassWithUserRationale`
  - 테스트 함수 (비대화형): `TestSkipAuditFlagInNonInteractiveRecordsAutoReason`
  - 파일 산출물 검증: 4 필드 모두 grep 매치
  - 환경변수 변종: `MOAI_SKIP_PLAN_AUDIT=1` 도 동일 동작 (테스트 함수 `TestEnvVarSkipAuditEquivalentToFlag`)

### `AC-WAG-07` — INCONCLUSIVE fall-back (auto-PASS 금지)
- **Source**: REQ-WAG-007 (Unwanted Behavior)
- **Traceability**: `REQ-WAG-007`
- **Given**: plan-auditor 호출이 (a) timeout(60s 초과), (b) malformed JSON 반환, (c) panic 중 하나로 실패하도록 mock 주입.
- **When**: 사용자가 `/moai run SPEC-INCONC-001` 을 호출한다.
- **Then**: 시스템은 (a) verdict를 PASS로 처리하지 **않으며**, (b) `verdict: INCONCLUSIVE`로 보고서에 기록하고, (c) AskUserQuestion으로 retry / proceed-with-acknowledgement / abort 3-way 결정을 사용자에게 위임한다. proceed 선택 시 progress.md에 `inconclusive_acknowledged_by: <user identifier>`가 강제 기록된다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_integration_test.go`
  - 테스트 함수: `TestPlanAuditorFailureClassifiesAsInconclusive` (3 sub-test: timeout / malformed / panic)
  - 로그 라인 패턴: `[plan-audit] verdict=INCONCLUSIVE, falling back to manual prompt`
  - 음성 검증: 로그에 `verdict=PASS` 라인이 등장하지 **않음**
  - 파일 산출물: `progress.md`에 `inconclusive_acknowledged_by` 필드 존재(proceed 선택 시)

---

## Part B: 추론된 엣지 케이스 AC (4건, INFERRED)

REQ에서 명시되지 않았으나 §4 in-scope 항목과 §5 위험 완화에 근거하여 검증해야 하는 케이스.

### `AC-WAG-08` — Grace window warn-only 모드
- **Source**: INFERRED (spec.md §4.1 in-scope: "7일 grace window")
- **Traceability**: `REQ-WAG-002` 의 시간 의존 변형, R-WAG-1 위험 완화
- **Given**: 본 SPEC merge timestamp가 `T0`, 현재 시각이 `T0 + 3일`(grace window 활성), `SPEC-GRACE-001`이 audit FAIL 상태.
- **When**: 사용자가 `/moai run SPEC-GRACE-001` 을 호출한다.
- **Then**: 시스템은 (a) FAIL을 stdout에 경고로 출력하되 Run 진행을 차단하지 **않으며**, (b) 출력에 grace window 잔여 일수 카운트다운 `[grace-window] D-4 (차단 모드 자동 전환)`을 포함하고, (c) progress.md에 `audit_verdict: FAIL_WARNED` 를 기록한다. `T0 + 8일` 시점 호출에서는 정상 차단 동작(`AC-WAG-02`와 동일)이 발생한다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_grace_test.go`
  - 테스트 함수: `TestGraceWindowWarnOnlyMode`, `TestGraceWindowExpiryRevertsToBlockingMode`
  - 시간 주입: `time.Now` mock 또는 `MOAI_AUDIT_GATE_T0` 환경변수로 grace 기준 시각 주입
  - 로그 라인 패턴: `[grace-window] D-N` 형태 정확히 매치

### `AC-WAG-09` — 24h audit cache hit
- **Source**: INFERRED (spec.md §5 R-WAG-2 mitigation: "audit report 24h validity 캐시")
- **Traceability**: `REQ-WAG-003` 의 비기능 요구
- **Given**: `SPEC-CACHE-001`에 대한 audit이 `T0`에 PASS로 기록됨. plan 산출물 hash(spec/plan/acceptance/tasks 4 파일의 SHA-256 결합)는 `H0`.
- **When**: `T0 + 1시간` 시점에 동일 SPEC에 대해 `/moai run SPEC-CACHE-001` 을 재호출한다(파일 변경 없음).
- **Then**: 시스템은 (a) plan-auditor를 재호출하지 **않고**, (b) 캐시된 verdict=PASS를 채택하며, (c) progress.md에 `audit_cache_hit: true`, `cached_audit_at: T0`를 기록한다. plan 산출물이 변경되어 hash가 `H1`이 되면 캐시는 무효화되어 plan-auditor가 재호출된다.
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_cache_test.go`
  - 테스트 함수: `TestAuditCacheHitWithin24Hours`, `TestAuditCacheInvalidatedOnPlanArtifactChange`
  - 로그 라인 패턴: `[plan-audit] cache hit (verdict=PASS, age=1h)`
  - 음성 검증: plan-auditor invocation count == 0(cache hit case)
  - 파일 산출물: `progress.md` cache 필드 2개 grep 검증

### `AC-WAG-10` — `.moai/reports/plan-audit/` 디렉터리 자동 생성
- **Source**: INFERRED (spec.md §4.1 in-scope: "디렉터리 생성")
- **Traceability**: `REQ-WAG-004` 의 사전 조건
- **Given**: `.moai/reports/plan-audit/` 디렉터리가 존재하지 않거나 `.gitkeep`만 있는 fresh 상태.
- **When**: `/moai run SPEC-FRESH-001` 이 처음 호출되어 audit이 실행된다.
- **Then**: 시스템은 (a) 디렉터리를 자동 생성하고(이미 있으면 노옵), (b) `SPEC-FRESH-001-<YYYY-MM-DD>.md`를 작성하며, (c) panic하거나 작업 디렉터리 외부에 파일을 생성하지 않는다. 권한 거부(읽기 전용 파일시스템) 시 명확한 에러 메시지 출력 후 INCONCLUSIVE 처리(`AC-WAG-07`로 수렴).
- **Evidence**:
  - 통합 테스트: `internal/cli/run_audit_gate_filesystem_test.go`
  - 테스트 함수: `TestAuditReportDirectoryAutoCreated`, `TestAuditReportReadOnlyFilesystemFallsBackToInconclusive`
  - 부가 검증: `os.Stat(".moai/reports/plan-audit")` IsDir 확인
  - 권한 테스트: `t.TempDir()` 후 `os.Chmod(0444)`로 readonly 시뮬레이션

### `AC-WAG-11` — Dogfood self-audit (본 SPEC 자체 PASS)
- **Source**: INFERRED (spec.md §1 마지막 단락 "dogfood 원칙", §5 R-WAG-3 mitigation)
- **Traceability**: 본 SPEC 자체의 self-validation
- **Given**: `SPEC-WF-AUDIT-GATE-001/{spec,plan,acceptance,tasks}.md` 4 파일이 모두 작성 완료된 상태.
- **When**: `plan-auditor` 서브에이전트가 본 SPEC을 단독 입력으로 호출된다.
- **Then**: verdict=PASS가 반환되며, must-pass 4개 항목(EARS compliance, acceptance criteria existence, exclusion section presence, frontmatter schema validity)이 모두 통과한다. scored 항목 4개의 평균 점수가 0.75 이상이다.
- **Evidence**:
  - 통합 테스트: `internal/cli/dogfood_self_audit_test.go`
  - 테스트 함수: `TestSelfAuditPassesOnOwnSpec`
  - 실행: `moai constitution audit .moai/specs/SPEC-WF-AUDIT-GATE-001/` (또는 plan-auditor를 mock harness로 직접 호출)
  - 보고서: `.moai/reports/plan-audit/SPEC-WF-AUDIT-GATE-001-<YYYY-MM-DD>.md` PASS
  - 부가 검증: `verdict: PASS` AND `must_pass_count_failed: 0`

---

## REQ → AC 추적성 매트릭스

| REQ ID | EARS Pattern | 1차 AC | 보강 AC |
|--------|--------------|--------|---------|
| `REQ-WAG-001` | Ubiquitous | `AC-WAG-01` | `AC-WAG-10` (디렉터리 사전조건) |
| `REQ-WAG-002` | Event-driven (FAIL) | `AC-WAG-02` | `AC-WAG-08` (grace window 변형) |
| `REQ-WAG-003` | Event-driven (PASS) | `AC-WAG-03` | `AC-WAG-09` (캐시 변형) |
| `REQ-WAG-004` | Ubiquitous (보고서) | `AC-WAG-04` | `AC-WAG-10` (디렉터리 자동 생성) |
| `REQ-WAG-005` | State-driven (--team) | `AC-WAG-05` | — |
| `REQ-WAG-006` | Optional Feature (--skip-audit) | `AC-WAG-06` | — |
| `REQ-WAG-007` | Unwanted Behavior (INCONCLUSIVE) | `AC-WAG-07` | `AC-WAG-10` (filesystem 실패 fall-back) |
| (Self) | dogfood | `AC-WAG-11` | — |

전체 REQ 7건 100% 커버, 추가 INFERRED AC 4건으로 엣지 케이스 보강. 단일 AC가 다중 REQ 보조 검증을 수행하는 cross-cut 구조.

---

## REQ → AC → Evidence 통합 표

| REQ | AC | Evidence (테스트 / 로그 / 산출물) |
|-----|-----|-----------------------------------|
| `REQ-WAG-001` | `AC-WAG-01` | `TestRunInvokesPlanAuditorBeforeImplementation` / `[plan-audit] invoking ...` / `<SPEC>-<DATE>.md` |
| `REQ-WAG-002` | `AC-WAG-02` | `TestRunBlockedOnAuditFail` / `[plan-audit] verdict=FAIL` / report verdict=FAIL |
| `REQ-WAG-002` (변형) | `AC-WAG-08` | `TestGraceWindowWarnOnlyMode` / `[grace-window] D-N` / progress.md FAIL_WARNED |
| `REQ-WAG-003` | `AC-WAG-03` | `TestRunProceedsOnAuditPassAndPersistsVerdict` / `[plan-audit] verdict=PASS` / progress.md 4 필드 |
| `REQ-WAG-003` (변형) | `AC-WAG-09` | `TestAuditCacheHitWithin24Hours` / `cache hit (age=1h)` / progress.md `audit_cache_hit: true` |
| `REQ-WAG-004` | `AC-WAG-04` | `TestDailyAuditReportAccumulatesMultipleRuns` / N/A / `## Audit Run N of M` |
| `REQ-WAG-004` (변형) | `AC-WAG-10` | `TestAuditReportDirectoryAutoCreated` / N/A / `os.Stat IsDir` |
| `REQ-WAG-005` | `AC-WAG-05` | `TestTeamRunBlockedBeforeTeammateSpawn` / `team mode detected` / tmux pane 미생성 |
| `REQ-WAG-006` | `AC-WAG-06` | `TestSkipAuditFlagRecordsBypassWithUserRationale` / N/A / report verdict=BYPASSED |
| `REQ-WAG-007` | `AC-WAG-07` | `TestPlanAuditorFailureClassifiesAsInconclusive` / `verdict=INCONCLUSIVE` / report INCONCLUSIVE |
| (Self) | `AC-WAG-11` | `TestSelfAuditPassesOnOwnSpec` / `verdict: PASS` / 본 SPEC report |

---

## Definition of Done

다음 체크리스트를 모두 만족할 때 본 SPEC 구현 완료:

### 기능 완성도

- [ ] AC-WAG-01 ~ AC-WAG-07 (REQ 직접 매핑) 모두 PASS
- [ ] AC-WAG-08 ~ AC-WAG-11 (INFERRED 엣지) 모두 PASS
- [ ] REQ 7건 → AC 11건 매핑 100% 검증

### 품질 게이트 (TRUST 5)

- [ ] **Tested**: 통합 테스트 5개 파일(`run_audit_gate_*_test.go`) 모두 PASS, 패키지 커버리지 ≥ 85%
- [ ] **Readable**: 신규 워크플로우 skill 단락에 `Phase 0.5: Plan Audit Gate` 헤더 명시, 모든 키 필드(verdict/path/timestamp/auditor)에 단어 정의 1줄 첨부
- [ ] **Unified**: `gofmt -l` empty, `golangci-lint run` zero error, 워크플로우 skill 단락이 기존 Phase 헤더 패턴 준수
- [ ] **Secured**: report 파일 경로는 `.moai/reports/plan-audit/` 내로 제한(filepath.Clean + projectDir scope), 사용자 입력 rationale은 마크다운 escape 후 기록
- [ ] **Trackable**: 모든 커밋 메시지에 `SPEC-WF-AUDIT-GATE-001` 레퍼런스, progress.md 4 필드(verdict/path/at/auditor) 기록

### 비기능 요구

- [ ] Audit gate 추가로 인한 `/moai run` 평균 latency 증가 ≤ 30초 (캐시 hit 시 ≤ 2초)
- [ ] Daily report 파일 크기 ≤ 200KB (3 run/day 시 < 100KB)
- [ ] Grace window 카운트다운 정확도: D-N 표기와 실제 잔여 일수 일치

### 프로세스 준수

- [ ] Template-First 규율: 6개 영향 파일 모두 `internal/template/templates/` 트윈과 byte-level 일치
- [ ] `make build` 후 embedded 파일 재생성 확인
- [ ] CHANGELOG.md 업데이트
- [ ] spec.md status 필드 `draft → implemented` 전환 (FROZEN 본문 변경 없음)

### 통합 검증 (수동)

- [ ] dogfood: 본 SPEC을 자기 자신의 게이트에 입력 → PASS 확인 (`AC-WAG-11`)
- [ ] grace window 종료 시점 자동 차단 모드 전환 확인 (시간 주입)
- [ ] team 모드 + skip-audit 조합 동작 확인 (cross-feature)

### Open Questions 해소

- [ ] plan.md §7 OPEN QUESTIONS 모두 결정 또는 후속 SPEC으로 이관 기록
- [ ] 특히 캐시 invalidation 정책(plan 산출물 hash 알고리즘) 문서화 완료

---

## 수동 검증 실행 순서 (최종 acceptance run)

QA 단계에서 다음 순서로 수동 실행:

1. `make build && make install` — 최신 바이너리
2. `moai --version` — 버전 출력 확인
3. 더미 SPEC 작성 (PASS/FAIL/BYPASS/INCONCLUSIVE 4가지) — `.moai/specs/SPEC-DUMMY-{PASS,FAIL,BYP,INC}-001/`
4. `/moai run SPEC-DUMMY-PASS-001` — `AC-WAG-01`, `AC-WAG-03`, `AC-WAG-10` 동시 검증
5. `/moai run SPEC-DUMMY-FAIL-001` (grace 종료 후) — `AC-WAG-02` 검증
6. `/moai run SPEC-DUMMY-FAIL-001 --skip-audit` — `AC-WAG-06` 검증
7. `/moai run SPEC-DUMMY-INC-001` (plan-auditor mock failure 주입) — `AC-WAG-07` 검증
8. `/moai run SPEC-DUMMY-PASS-001 --team` — `AC-WAG-05` 검증
9. `/moai run SPEC-DUMMY-PASS-001` 즉시 재호출 — `AC-WAG-09` 캐시 hit 검증
10. 4 더미 SPEC에 대해 동일 일자 다회 호출 → `AC-WAG-04` 누적 검증
11. `cat .moai/reports/plan-audit/*.md` — 모든 verdict 분포 (PASS/FAIL/BYPASSED/INCONCLUSIVE) 확인
12. `go test -race ./internal/cli/...` — 통합 테스트 전체 통과
13. `go test ./internal/cli/dogfood_self_audit_test.go` — `AC-WAG-11` 검증

모든 단계 통과 시 DoD 완료로 간주.

---

## 요약

- **AC 총 개수**: 11 (REQ 직접 매핑 7 + INFERRED 4)
- **REQ 커버리지**: 100% (REQ 7건 모두 ≥ 1 AC)
- **테스트 파일 수**: 5 (`run_audit_gate_integration_test.go`, `run_audit_gate_grace_test.go`, `run_audit_gate_cache_test.go`, `run_audit_gate_filesystem_test.go`, `team_run_audit_gate_test.go`, `dogfood_self_audit_test.go`)
- **TRUST 5 게이트**: 5/5 명시적 검증 항목 포함
- **Dogfood**: `AC-WAG-11`로 self-audit 명시
