# acceptance.md — SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 수용 기준

> 각 AC는 Given-When-Then 시나리오 + 관측 가능 증거 요구사항. 심각도 분류 (S1 Blocker / S2 Critical / S3 Major / S4 Minor). TRACE = 요구사항 매핑. Tier M — ≥2 AC per REQ (블록커는 ≥3).

---

## §D. AC 매트릭스

### AC-ADM-001 (TRACE: REQ-ADM-001) — upsert 멱등성 [S1 Blocker]

**Given** 선호 메모리 계층이 초기화되어 있고
**When** 동일 (domain, decision_key)에 대해 2회 연속 upsert가 발생하면
**Then** 메모리에는 1개의 엔트리만 존재하며, 두 번째 upsert가 첫 번째를 교체한다 (append 아님).

**관측 증거**: `moai preference list --domain=<D> --key=<K>` 출력이 단일 엔트리; `user_decisions/recall.jsonl` grep 결과 1건; upsert 전후 엔트리 카운트 불변.

**심각도 근거**: append-only는 토큰 비용 선형 증가 + 검색 노이즈 — 통합 원칙 위반은 SPEC 목적 자체를 무력화.

### AC-ADM-002 (TRACE: REQ-ADM-002) — 네임스페이스 분리 [S1 Blocker]

**Given** 기존 기술 교훈 메모리(`memory/feedback_*.md`, `MEMORY.md`)가 존재하고
**When** 선호 메모리 계층이 활성화되면
**Then** 사용자 의사결정 사실은 `memory/user_decisions/`에, 기술 교훈은 `memory/feedback_*.md`에 각각 독립 저장되며, 어느 한쪽의 쿼리가 다른 쪽을 오염시키지 않는다.

**관측 증거**: `grep -r "<decision_fact>" memory/feedback_*.md` → 0건; `grep -r "<engineering_lesson>" memory/user_decisions/` → 0건; 두 경로의 독립적 쿼리 가능.

### AC-ADM-003 (TRACE: REQ-ADM-003) — 엔트리 스키마 필수 필드 [S2 Critical]

**Given** 캡처 또는 추론 경로가 선호 엔트리를 기록하고
**When** 엔트리가 저장되면
**Then** `fact`, `source_citation`, `valid_time`, `last_used`, `scope` (stable|transient), `domain`, `confidence` (observed|inferred) 7개 필드가 모두 존재한다.

**관측 증거**: 임의 엔트리 `cat memory/user_decisions/recall.jsonl | jq '.[0] | keys'` 출력이 7개 필드 포함; 누락 필드 시 스키마 검증 테스트 실패.

### AC-ADM-004 (TRACE: REQ-ADM-004) — 3계층 검색 계단식 [S2 Critical]

**Given** core(≤4KB), recall(최근 N 세션), archival(전체) 3계층이 존재하고
**When** orchestrator가 선호 회수를 요청하면
**Then** core 먼저 검색 → 미스 시 recall → 미스 시 archival 순으로 계단식 fallthrough하며, core 히트 시 recall/archival 미접근.

**관측 증거**: core에 존재하는 엔트리 회수 시 recall/archival 파일 read 시도 0건 (파일 접근 로그 또는 strace); recall에만 존재 시 core miss 후 recall hit.

### AC-ADM-005 (TRACE: REQ-ADM-005) — 정보이익 정렬 발화 [S2 Critical]

**Given** orchestrator가 다가오는 결정에 대해 불확실성 p를 추정하고
**When** p ≈ 0.5 (Fisher 정보 I=p(1−p) 최대)이면
**Then** AskUserQuestion으로 해당 질문을 발화한다; p ≈ 0 또는 1이면 자동 처리하고 질문을 생략한다.

**관측 증거**: 결정 로그에 추정 p값 + 발화/생략 결정 기록; p≈0.5 결정의 발화율 100%, p>0.8 결정의 생략율 ≥ 임계값.

**주의**: p 추정은 휴리스틱(초기: 관측된 동일 결정의 다수 선택 비율); "complete" tier에서 Bayes 정제 이월.

### AC-ADM-006 (TRACE: REQ-ADM-006) — 질문 순서 정보이익 내림차순 [S3 Major]

**Given** 하나의 AskUserQuestion 호출에 여러 질문이 배치되고
**When** orchestrator가 각 질문의 정보이익을 추정하면
**Then** 질문은 정보이익 내림차순으로 정렬된다.

**관측 증거**: AskUserQuestion 호출의 questions 배열 순서 = 추정 정보이익 내림차순; 로그에 순서 결정 근거.

### AC-ADM-007 (TRACE: REQ-ADM-007) — 통계적 다수 합리적 기본값 추천 [S1 Blocker]

**Given** 선호 메모리에 동일 도메인의 N개 이상 관측된 의사결정이 존재하고
**When** orchestrator가 추천 옵션을 배치하면
**Then** `(권장)` 라벨이 붙은 첫 옵션은 관측된 통계적 다수 선택이며, 시스템이 밀고 싶은 정책 기본값이 아니다.

**관측 증거**: 추천 배치 로그에 "recommended=<majority_observed>, basis=<N_observations>, not system_default"; 시스템 기본값과 다를 시 그 사실 명시.

**cold-start 공개**: 관측 < N일 시 추천 description에 "based on static default, N observations needed for personalization" 포함.

### AC-ADM-008 (TRACE: REQ-ADM-008) — 전제조건 서술 [S2 Critical]

**Given** 추천 옵션이 배치되고
**When** 사용자가 옵션 description을 읽으면
**Then** 추천이 성립하는 전제조건이 서술되어 있어, 전제 위반 시 사용자가 즉시 거부할 수 있다.

**관측 증거**: 각 추천 옵션 description에 "Recommended when <precondition>" 또는 동등 문구 포함; 전제 서술 누락 시 lint/감사 실패.

### AC-ADM-009 (TRACE: REQ-ADM-009) — advisory/fail-open exit 0 [S1 Blocker]

**Given** PostToolUse 캡처 훅이 AskUserQuestion tool_result를 처리 중
**When** 다음 중 하나라도 발생하면 — (a) stdin JSON 파싱 실패, (b) 선호 메모리 upsert 실패, (c) 디스크 가득 참, (d) 권한 거부, (e) 내부 패닉
**Then** 훅은 exit 0으로 종료하고 (워크플로 계속 진행), `.moai/logs/hook-stderr.log`에 warning을 기록한다.

**관측 증거**: 5가지 오류 시나리오 각각에 대한 단위/통합 테스트에서 exit code 0; 로그 파일에 warning 라인; 워크플로 중단 없음.

**심각도 근거**: advisory 훅이 exit 2로 차단하면 워크플로 전체가 중단 — SPEC의 fail-open 약속 핵심.

### AC-ADM-010 (TRACE: REQ-ADM-010) — Recovery-Signal Carve-Out [S3 Major]

**Given** REQ-ADM-010은 SHOULD (doctrine-honest)이며 — 본 SPEC은 회복 턴에서의 *행동*(exit 0 + 캡처 미실행)만 정의하고, 회복 턴 *탐지 메커니즘*(stopReason 파싱)은 future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`로 이연됨 (`runtime-recovery-doctrine.md §4` + AP-RR-006 준거)
**When** advisory 캡처 훅의 소스 코드를 감사하면
**Then** 훅 소스는 이 이연을 문서화하고 있어야 하며 (탐지 메커니즘이 future SPEC에 위임됨을 명시), 오늘 날짜로 회복 턴을 기계적으로 탐지한다고 주장해서는 안 된다.

**관측 증거**: advisory 훅 소스(`internal/hook/post_tool.go` user_decision_capture 서브파이프라인)에 stopReason 파싱 불가 + future SPEC 위임 명시 주석 존재; `runtime-recovery-doctrine.md §4` + AP-RR-006("current hooks do NOT parse stopReason; no mechanical enforcement is possible") 인용; 회복 턴을 기계적으로 탐지한다는 허위 주장 grep → 0건.

**VERIFICATION GAP**: 회복 턴의 기계적 탐지(stopReason 파싱)는 본 SPEC scope 외이며 future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`에서 구현된다. 본 AC는 "탐지 불가를 정직하게 문서화했는가"를 검증하며, "회복 턴을 실제로 탐지했는가"를 검증하지 않는다.

**심각도 근거 (S1→S3 downgrade)**: REQ-ADM-010은 SHOULD로 downgrade됨 (doctrine-honest). 현재 advisory 훅 layer에서는 stopReason을 파싱할 수 없으므로 (`runtime-recovery-doctrine.md §4` + AP-RR-006), 회복 턴 exit-0 동작의 기계적 강제는 불가능하다. S1 Blocker로 남기면 vacuous-pass 위험 (구현자가 가짜 신호나 대체 proxy를 조용히 대입). S3 Major는 정직한 문서화를 요구하면서 over-claim을 방지한다.

### AC-ADM-011 (TRACE: REQ-ADM-011) — stable/transient 분리 감쇠 [S1 Blocker]

**Given** stable-scope 엔트리(예: "사용자는 Go 백엔드 선호")와 transient-scope 엔트리(예: "이번 세션는 verbose 로그 선호")가 존재하고
**When** 28일이 경과하면
**Then** transient 엔트리는 soft-delete되고, stable 엔트리는 pure time-decay에 면제되어 잔존한다.

**관측 증거**: 28일 경과 시뮬레이션 후 stable 엔트리 존재 + transient 엔트리 archival 이동; stable 엔트리의 confidence weight 불변.

**심각도 근거**: 순진 time-decay가 stable 선호를 잃으면 Koren의 "지속 신호 상실" 재현 — 사용자 핵심 선호 반복 질문.

### AC-ADM-012 (TRACE: REQ-ADM-012) — 28일 TTL + 사용시 리셋 [S2 Critical]

**Given** transient-scope 엔트리의 last_used가 28일 초과이고
**When** 일일 감쇠 스캔이 실행되면
**Then** 해당 엔트리는 soft-delete된다; 엔트리가 28일 이내에 재사용되면 age 카운터가 0으로 리셋되고 confidence weight가 부양된다.

**관측 증거**: 28일 + 1일 엔트리의 soft-delete; 27일 재사용 엔트리의 age 리셋 + weight 부양 테스트.

### AC-ADM-013 (TRACE: REQ-ADM-013) — 세션 단위 개인화 비활성화 [S2 Critical]

**Given** 사용자가 `/moai preference toggle`을 실행하거나 "이번 세션 개인화 비활성화"를 명시하고
**When** orchestrator가 후속 AskUserQuestion을 발화하면
**Then** 해당 세션 동안 선호 기반 추천 배치와 불확실성 기반 질문 생략이 모두 억제된다; 다음 세션에서 자동 재활성화된다.

**관측 증거**: 토글 후 세션에서 preference-based `(권장)` 배치 0건 + 자동 처리된 결정 0건 (모두 질문); `.moai/state/session-preference-disabled` 센티넬 존재; 신규 세션에서 센티넬 부재.

### AC-ADM-014 (TRACE: REQ-ADM-014) — 민감 도메인 강도 저하 [S2 Critical]

**Given** orchestrator가 민감 도메인(security review, 일회성 탐색 쿼리, cold-start)에서 작동 중
**When** 추천 배치를 평가하면
**Then** 추천 강도가 neutral로 저하되고 (inferred preference 기반 `(권장)` 배치 없음), 저하 사실이 공개된다.

**관측 증거**: security review 도메인에서 inferred preference 기반 추천 0건; "personalization reduced for sensitive domain" 공개 로그.

### AC-ADM-015 (TRACE: REQ-ADM-015) — 데이터 신선도 공개 [S3 Major]

**Given** 추천이 선호 메모리 데이터를 기반하고
**When** 추천 옵션이 배치되면
**Then** 옵션 description에 "based on N-day-old data" 형태의 신선도 공개가 포함된다.

**관측 증거**: 추천 옵션 description grep "based on .*-day-old data" 또는 동등; 신선도 누락 시 감사 실패.

### AC-ADM-016 (TRACE: REQ-ADM-016) — 정정 루프 [S2 Critical]

**Given** orchestrator가 추론된 선호를 공개하고
**When** 사용자가 "이 추론이 틀리다"고 정정하면
**Then** 정정된 사실이 `confidence: observed`로 즉시 upsert되고, 기존 inferred 엔트리의 weight가 감소한다.

**관측 증거**: 정정 전후 메모리 상태 비교 — inferred 엔트리 weight 감소 + 새 observed 엔트리 존재; 정정 채널 미제공 시 감사 실패.

### AC-ADM-017 (TRACE: REQ-ADM-017) — 숙련도 기반 적응형 강도 [S2 Critical]

**Given** orchestrator가 사용자 숙련도를 추정하고 (세션 카운트 ≥ 임계값, 의사결정 일관성, 또는 명시적 자가 평가)
**When** 고숙련도(전문가)로 추정되면
**Then** 약 추천 강도(info-centric, `(권장)` 라벨 override 없이 inferred preference 공개); 저숙련도(일반)로 추정되면 강 추천 강도(`(권장)` 라벨 + 투명한 이유).

**관측 증거**: 숙련도 추정 로그; 전문가 세션에서 `(권장)` override 0건; 일반 사용자 세션에서 `(권장)` + reason 포함.

### AC-ADM-018 (TRACE: REQ-ADM-018) — verification-claim-integrity 준수 [S1 Blocker]

**Given** orchestrator가 추천을 발화하고
**When** 추천이 "과거 사용자 의사결정을 반영한다"고 주장하면
**Then** 해당 추천은 관측된 선호 메모리 엔트리(`confidence: observed` 또는 `inferred` with 공개된 근거)에 매핑되며, 관측되지 않은 가정에 기반한 추천은 발화되지 않는다.

**관측 증거**: 각 추천의 메모리 엔트리 백링크 (`memory_entry_id`); 백링크 없는 추천 0건 (grep `recommendation without memory_entry` → 0); `verification-claim-integrity.md §1.1 surface 3` 준수 감사 통과.

**심각도 근거**: 관측되지 않은 추론 주장은 verification-claim-integrity 위반 — "증거 부재 ≠ 성공 증거" 원칙 위반.

### AC-ADM-NFR-001 (TRACE: NFR-ADM-001) — 캡처 훅 지연 [S3 Major]

**Given** PostToolUse 캡처 훱이 정상 실행되고
**When** 지연을 측정하면
**Then** p95 지연이 50ms 이하이다.

**관측 증거**: 벤치마크 테스트 — 1000회 캡처 p95 ≤ 50ms.

### AC-ADM-NFR-002 (TRACE: NFR-ADM-002) — core 계층 크기 [S3 Major]

**Given** core 계층이 항상 로드되고
**When** core 크기를 측정하면
**Then** 4KB 이하이다; 초과 시 recall로 강등된다.

**관측 증거**: `wc -c memory/user_decisions/core.yaml` ≤ 4096; 초과 시 강등 로그.

### AC-ADM-NFR-003 (TRACE: NFR-ADM-003) — 추천 배치 결정 지연 [S3 Major]

**Given** orchestrator가 AskUserQuestion 발화 직전에 추천 배치 결정을 평가하고
**When** 결정 지연을 측정하면
**Then** p95 지연이 10ms 이하이다 (인라인 결정 — 발화 직전에 평가되므로 워크플로 병목이 되어서는 안 됨).

**관측 증거**: 벤치마크 테스트 — 1000회 추천 배치 결정 p95 ≤ 10ms (core 계층 조회 + 통계적 다수 계산 포함).

### AC-ADM-NFR-004 (TRACE: NFR-ADM-004) — 감쇠 스캔 주기 [S3 Major]

**Given** 감쇠 정책이 백그라운드에서 실행되고
**When** 일일 감쇠 스캔 실행 빈도를 측정하면
**Then** 1일 1회 실행된다 (매 턴 감쇠 계산은 비용 과다이므로 백그라운드 batch).

**관측 증거**: `moai preference decay-scan` 실행 로그 (또는 SessionStart 훅 분기) — 24시간 window 내 정확히 1회 실행; 스캔 타임스탬프 간격 23~25시간.

### AC-ADM-NFR-005 (TRACE: NFR-ADM-005) — 회복 제어 토글 세션 단위 (비영구) [S3 Major]

**Given** 사용자가 `/moai preference toggle`로 개인화를 비활성화하고
**When** 신규 세션을 시작하면
**Then** 토글 상태는 세션 단위로만 유지되며 (영구 아님), 신규 세션에서 개인화가 자동 재활성화된다 (Loughrey 자율성 — 세션마다 재활성화 가능).

**관측 증거**: 세션 A에서 토글 ON → `.moai/state/session-preference-disabled` 센티넬 존재; 세션 종료 후 신규 세션 B 시작 시 센티넬 부재 + 개인화 자동 활성화 (추천 배치 정상 동작). 영구 설정 파일(`.moai/config/`)에 토글 상태 누출 0건.

### AC-ADM-NFR-006 (TRACE: NFR-ADM-006) — 템플릿 중립성 경계 [S1 Blocker]

**Given** template-shipped 산출물(`internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` + `internal/template/templates/.claude/output-styles/moai/moai.md`)이 작성되고
**When** 내부 SPEC ID / REQ 토큰 / 내부 날짜 / commit SHA / archive-memory 경로를 grep하면
**Then** 0건이 검출된다.

**관측 증거**: `grep -rE 'SPEC-V3R6|REQ-ADM|AC-ADM|2026-06-24|[0-9a-f]{40}|\.moai/(specs|reports)' internal/template/templates/.claude/rules/moai/core/askuser-protocol.md internal/template/templates/.claude/output-styles/moai/moai.md` → 0건; CI guard `internal/template/internal_content_leak_test.go` 통과.

---

## §D.1 심각도 분류 요약

| 심각도 | AC | 의미 |
|--------|----|----|
| S1 Blocker | 001, 002, 007, 009, 011, 018, NFR-006 | 미충족 시 SPEC 목적 무력화 또는 워크플로 중단 |
| S2 Critical | 003, 004, 005, 008, 012, 013, 014, 016, 017 | 미충족 시 핵심 기능 저하 |
| S3 Major | 006, 010, 015, NFR-001, NFR-002, NFR-003, NFR-004, NFR-005 | 미충족 시 사용자 경험 저하 (010은 SHOULD doctrine-honest — 탐지 메커니즘 future SPEC 이연) |
| S4 Minor | (없음 — Tier M에서 S4는 추적 제외) | — |

---

## §D.2 간접 검증 (direct evidence 불가 시)

| AC | 간접 증거 | 근거 |
|----|----------|------|
| AC-ADM-005 (p 추정) | 결정 로그의 추정 p값 + 발화/생략 결정 기록 | p 추정 자체가 휴리스틱이므로 "정확한 p"가 아닌 "결정의 일관성"을 검증 |
| AC-ADM-017 (숙련도 추정) | 숙련도 추정 방법 + 추정값 + 강도 분기 로그 | 숙련도 "참값"이 없으므로 추정 절차의 일관성을 검증 |

---

## §D.3 폐쇄 게이트 (Definition of Done)

- [ ] 모든 S1 Blocker AC PASS (7개 — 010은 S3으로 downgrade됨: SHOULD doctrine-honest)
- [ ] 모든 S2 Critical AC PASS (9개)
- [ ] 모든 S3 Major AC PASS (8개 — 010, NFR-003/004/005 포함)
- [ ] `moai spec lint` PASS
- [ ] `moai spec audit` drift 0
- [ ] `internal/template/internal_content_leak_test.go` PASS (템플릿 중립성)
- [ ] 기존 exit-2 훅과 cohabitation 테스트 PASS
- [ ] plan-auditor 독립 감사 PASS (편향 방지)
- [ ] Implementation Kickoff Approval (사용자 명시적 run-phase 진입 승인)
- [ ] V3R6 3-phase close: plan → run → sync, frontmatter `status: implemented`, `progress.md §E.4 sync_commit_sha` 백fill

---

## §D.4 Forward-looking checks (run-phase 발견 시 plan-phase에 회신)

- 캡처 훅이 새로운 CC 버전의 PostToolUse stdin JSON 스키마 변경에 어떻게 대응하는가 → advisory/fail-open으로 미인식 필드 warn-and-skip
- 숙련도 추정이 세션 카운트만으로 초기에 부정확할 때 → cold-start 게이트(REQ-ADM-014)가 neutral 강도로 보호
- 멱법칙 α=0.5 고정값이 도메인별로 부적합할 때 → "complete" tier에서 동적 학습 이월 (spec.md §E Out of Scope)

---

## §D.5 Edge cases

| Edge | 기대 동작 | AC |
|------|----------|----|
| 캡처 중 프로세스 강제 종료 (SIGKILL) | 부분 upsert 없음 (원자적 쓰기 또는 임시 파일 + rename) | AC-ADM-001 |
| 두 세션이 동시에 동일 결정 캡처 | 마지막 writer wins (원자적 upsert); race는 advisory이므로 허용 | AC-ADM-001, AC-ADM-009 |
| 사용자가 정정 후 즉시 동일 추천 재발화 | 정정된 observed 엔트리 우선; inferred weight 감소 후 추천 변경 | AC-ADM-016 |
| core 크기 4KB 초과 시 | 가장 오래된/낮은 weight 엔트리 recall로 강등 | AC-ADM-NFR-002 |
| 민감 도메인 + 전문가 사용자 동시 | 민감 도메인 게이트가 우선 (neutral 강도); 숙련도 분기는 비민감 도메인에서만 | AC-ADM-014, AC-ADM-017 |
| 회복 턴 중 AskUserQuestion 발화 | 캡처 훅은 exit 0 + 미실행 (SHOULD); orchestrator의 질문 자체는 정상 (캡처만 생략). 단, 탐지 메커니즘은 future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001로 이연되어 현재 layer에서 기계적 강제 불가 (AP-RR-006) | AC-ADM-010 |

---

## §D.6 품질 게이트 기준

| 게이트 | 기준 | 담당 |
|--------|------|------|
| plan-auditor | 독립 회의적 감사 PASS (편향 방지) | plan-auditor |
| LSP (run-phase) | 0 errors, 0 type-errors, 0 lint-errors | manager-develop |
| coverage | `internal/cli/preference/` ≥ 85%, `internal/hook/post_tool.go` 확장 분기 ≥ 85% | manager-develop |
| sync-auditor | 4-dimension scoring (Functionality/Security/Craft/Consistency) ≥ 임계값 | sync-auditor |
| 템플릿 중립성 CI | `internal/template/internal_content_leak_test.go` + `template-neutrality-check.yaml` PASS | CI |

---

## §D.7 Verification 인용 준거 (verification-claim-integrity 준수)

본 acceptance.md의 모든 AC는 `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 2(manager-agent §E self-verification)를 준수한다. 각 AC의 PASS 주장은 관측된 명령 출력(grep, 테스트, audit)에 귀속되며, 미관측 항목은 Gap으로 명시된다.
