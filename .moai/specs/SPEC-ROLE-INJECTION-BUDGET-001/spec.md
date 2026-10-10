---
id: SPEC-ROLE-INJECTION-BUDGET-001
title: "역할 규칙 주입 예산 — 조립 합본 9,000자 회계, 역할 core 재작성, *-core 스텁 게이트, 버전 불일치 안내"
version: "0.1.0"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec
priority: P0
phase: "v3.2.1"
module: "internal/hook, internal/template/templates/.claude/rules/moai/workflow, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "role-injection, session-start, budget, role-core, relocation-audit, version-skew, template"
tier: M
related_specs: [SPEC-ALWAYS-LOADED-BUDGET-001, SPEC-ALWAYS-LOADED-DIET-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-ALWAYS-LOADED-HEADROOM-001, SPEC-INSTRUCTIONS-BUDGET-001]
---

# SPEC-ROLE-INJECTION-BUDGET-001

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-10-10 | 0.1.0 | 최초 작성. 카드 t1617 (3.2-1-5, P0, Class C — 리더 처분 dd1e, 운영자 지시 「이 문제 부터 해결하자」, no-new-dispatch 규칙 fde3의 예외). 기준 트리 카드 워크트리 HEAD `2aab5f797` (브랜치 WT-3-2-1). 상위 SPEC `SPEC-ALWAYS-LOADED-BUDGET-001` 의 후속 수리다. plan 단계 전수 재측정 기록은 `.moai/reports/t1617/measurements.md`. | manager-spec |

---

## §A. 배경

### §A.1 증상 A — 조립 합본 한도 초과 (이 SPEC 의 본제)

라이브 lane-15 세션의 시작 공지가 이 카드를 발행했다:

> 역할 규칙 주입 초과(factory-lane 세션): 조립된 맥락(23166자)이 세션 시작 전달 한도 10000자를 넘습니다

`internal/hook/role_rules.go` 는 역할 세션(팩토리 리더·레인)의 SessionStart(startup/clear/compact)에 두 역할 한정 규칙(`factory-dispatch.md`, `cross-session-messaging.md`)의 역할 core 를 주입한다(상위 SPEC 의 REQ-ALB-007). 전달 한도는 10,000 UTF-16 코드 단위(`roleRulesContextLimit`, role_rules.go:42, decision-index Q4 — 올릴 수 없음)이고, 크기 게이트(`roleRuleSizeGate`, role_rules.go:414)는 최종 조립 합본을 dispatch-finalize 시점(`FinalizeSessionStartOutput`, session_start.go:70)에 잰다. 한도 초과 시 REQ-ALB-010 의 오버플로 파일 전달 사다리가 작동한다 — 메커니즘은 올바르고, 이 SPEC 의 목표는 그것이 **발화하지 않게** 하는 것이다.

현재 역할 core 는 예산을 4배 이상 넘는다. 이 트리에서 잰 값(§C.2 재측정 명령, `.moai/reports/t1617/measurements.md`):

| 측정 대상 | 값 (UTF-16) |
|---|---|
| `roleRulesContextLimit` (올릴 수 없는 상한) | 10,000 |
| lane-15 세션 조립 합본 (게이트 자체 측정, 카드 발행 근거) | 23,166 |
| lane-15 세션 역할 블록 이전 생산자 (dispatch 분해 — §A.3 참조) | 4,376 |
| 배포본 `factory-dispatch.md` 역할 core (36개 영역 join) | 18,114 |
| 템플릿본 `factory-dispatch.md` 역할 core (같은 36개 영역) | 17,793 |
| 역할 블록 헤더 (lane 123 / leader 125) + 포인터 128 + 결합자 | 253–255 |
| `cross-session-messaging.md` 역할 core | 0 (빈 영역 1개 — 포인터만 전달) |

### §A.2 증상 B — 버전 불일치 설치의 실패 경로 (mo.ai.kr)

규칙 파일이 구버전(rc.28, 역할 core 표지 없음)이고 바이너리가 신버전(rc.29)인 설치는 `buildRoleCore` 의 "carries no markers" 오류 → REQ-ALB-009 InjectionFailed 경로로 떨어진다. 오늘 그 경로의 운영자 경고 4개 로캘(role_rules.go:263/274/285/296)과 에이전트용 지시가 원인 설명만 하고 수리 명령(`moai update`)을 이름 대지 않는다. 배포본 `.moai/config/sections/system.yaml:45` 의 `template_version`(템플릿 원본 `system.yaml.tmpl:9` 의 `{{.Version}}` 렌더)이 "어느 바이너리 버전이 이 파일들을 마지막에 배포했는가"의 판독 가능한 스탬프다 — 이 리포 자체가 v3.1.3 스탬프에 rc.29 바이너리인, 증상 B 가 실제로 살아 있는 상태다.

### §A.3 잔여 정비 항목 (카드 본문 keep-items — 전부 범위 안)

1. **경고 문구 정리** — `roleRuleSizeGate` 의 doc 주석이 같은 사다리를 두 번 말한다(role_rules.go:393–401 글머리 목록, :403–413 반복 문단). 카드가 지목한 role_rules.go:540 부근의 NOTE 지시문은 "10000" 천단위 없는 표기와 운영자 경고·로캘 경고 간 중복 서술이 있다.
2. **훅 출력 한도 문서 대조** — `.claude/rules/moai/core/hooks-system.md:131` 은 "Hook stdout over 50K characters is saved to disk…"만 말하고, 관측된 `additionalContext` 문자열당 10,000자 전달 한도(decision-index Q4, lane-15 공지가 실측)는 문서에 없다. 두 한도는 다른 메커니즘이다(총 stdout→디스크 저장 vs 문자열당 전달 한도) — 실측과 함께 구분 기재한다. 이 파일은 `paths:` 한정이라 상시 로드 비용 의무가 없다.
3. **gitflow 로컬 규칙 포인터 정합** — `.claude/rules/local/gitflow-lane-protocol.md` §1 이 세션 이동 금지의 정본으로 `.claude/rules/moai/workflow/factory-dispatch.md` 의 "Isolation 절"을 지목한다. 그 절의 본문은 상시 로드 다이어트에서 `factory-dispatch-mechanics.md` § Isolation 으로 이동했고, `factory-dispatch.md` 에는 절과 [HARD] 한 줄 요지(역할 core 영역 안)만 남았다. 재작성이 이 영역을 다루므로, 분배 파일의 문장과 로컬 규칙의 지목이 계속 서로를 가리키게 유지한다. 배포본과 템플릿본의 유일한 갈림(마지막 영역 +321 UTF-16 — `moai worktree sweep` 처분 꼬리 문장)도 이 영역 안에 있다.
4. **역할별 조립 예산 테스트** — 완료 기준 2의 그것. REQ-RIB-002.

### §A.4 해석 결정 — 「역할별 조립 ≤9,000자」의 두 판독

카드 본문의 "역할 카드 ≤9,000자 재작성"은 두 가지로 읽힌다. **축자 판독**(역할 카드 자체 ≤9,000)은 자기 모순이다: 9,000 역할 카드 + 4,376 생산자 = 13,376 > 10,000 한도 → 오버플로가 살아남아 완료 기준 1(오버플로 파일 0)을 위반한다. **조립 판독**(역할별 조립 합본 ≤9,000)만 자기 일관적이고 완료 기준 두 개가 함축하는 바와 같다. 이 SPEC 은 조립 판독으로 확정한다(REQ-RIB-001).

## §B. 용어

| 용어 | 뜻 |
|---|---|
| 계수 단위 | UTF-16 코드 단위(상위 SPEC §B 와 같다). 한글·영문 1, 보충 평면 2 |
| 조립 합본 | `FinalizeSessionStartOutput` 이 크기 게이트에 넘기는 최종 `additionalContext` — 이전 생산자 전체 + 결합자(`\n\n`) + 역할 블록. 게이트가 재는 대상 |
| 역할 블록 | 역할 core 주입이 만드는 부분 — 헤더 1줄 + 역할 core + (역할 core 가 빈 규칙의) 포인터 |
| 생산자 기준선 | 역할 블록 이전 생산자들의 UTF-16 합계. **4,376으로 고정** — lane-15 세션(카드 발행 세션, 2026-10-09T17:0xZ)에서 리더가 직접 분해해 잰 직접 측정치(세션 귀속, 종속 내용: 세션 귀속 행·팩토리 메시징 바인드·GLM 백엔드 라우팅 공지·Factory Mode 합류 줄·상립 스폰 권한). plan 단계 재측정은 불가(라이브 세션 자산) — 합본 교차 검산: 4,376 + 본 트리 역할 블록 18,690 = 23,066 vs 게이트 실측 23,166, 차 100(±100 노이즈, §D 설계 목표 4,000이 흡수) |
| core 예산 | 조립 판독에서 역할할 수 있는 역할 core 상한 = 9,000 − 4,376 − 2 − 125(leader 헤더) − 2 − 128(포인터) = **4,367** (leader 가 더 큰 헤더로 구속 역할). lane 은 4,369 |
| 설계 목표 | 4,000 (core 예산 − 367 여유: 생산자 기준선 ±100 노이즈와 재측정 오차 흡수) |
| 재배치 감사 | 역할 core 에서 잘려나가는 모든 조각에 대해 (a) 목적지(이미 그 본문을 가진 companion 절, 또는 먼저 이관한 곳)를 이름 대고 (b) 구속 절은 의미 보존 압축 재작성으로 core/스텁에 남기는 감사. `relocation-ledger.md` 로 기록 |
| 스텁 | `*-core.md` — 상시 표면에 남는 역할별 짧은 파일. 현재 2개(`factory-dispatch-core.md`, `cross-session-messaging-core.md`, 배포·템플릿 쌍) |
| 버전 불일치 | 배포 트리의 규칙 파일 세대와 실행 바이너리 세대의 어긋남. `system.yaml` `template_version` vs 바이너리 버전으로 판독 |
| 상위 원장 | `internal/template/testdata/binding_ledger.json` — 상위 SPEC 의 구속 원장(테스트 고정물). 이 파일들에 대해 36개 `role-core:` 행을 이미 가지며, 어떤 규칙 편집 후에도 현재 배포 트리와의 정합을 REQ-ALB-015 원장 테스트가 검증한다(상시 유지보수 계약 — 완결 SPEC 의 의미론 수정이 아니라 그 계약의 이행) |

## §C. 요구사항 (GEARS)

### C.1 조립 예산

- **REQ-RIB-001** (Ubiquitous) — The per-role assembled SessionStart composite — the pinned producer baseline (4,376 UTF-16, the lane-15 direct decomposition) plus the joiner plus the role block — shall stay at or under 9,000 UTF-16 code units for every role in the role-marker registry, with the 10,000-unit delivery cap unchanged (decision-index Q4, unraisable); the derived role-core ceiling is 4,367 UTF-16 (leader-binding) and the design target is 4,000. The SPEC's reading resolution is binding: 「역할별 조립 ≤9,000자」 is the assembled-composite reading (§A.4); the literal role-card-only reading is named unsatisfiable and rejected.
- **REQ-RIB-002** (Event-driven) — When the template tree's role-gated rule files change, a Go budget test shall build each registry role's role block from the deployed-tree files (template SSOT and this repository's own deployed copy both measured), assemble it with the pinned producer baseline constant, and fail when any role's assembled composite exceeds 9,000 UTF-16 — the failure message naming the breakdown (producers, joiner, header, core, pointer, total) so the arithmetic is auditable; it shall additionally assert the core ceiling (4,367) as a diagnostic sub-assertion.
- **REQ-RIB-003** (Ubiquitous) — The role-core rewrite shall pass a relocation audit recorded in `relocation-ledger.md` (SPEC dir): every chunk removed from the core names its destination — a companion section that already carries the text (verified: the destination heading holds the binding sentence) or a companion the chunk was moved into first — and every binding clause stays reachable through the injection path or the always-loaded stub as a meaning-preserving compressed rewrite whose after-text updates the parent binding-ledger rows (the standing maintenance contract; REQ-ALB-015's ledger test is the mechanical witness). No deletion without a named destination. `buildRoleCore`'s marker validation is untouched.
- **REQ-RIB-004** (Event-driven) — When the measured compression floor of the rewrite — the lowest core size reachable without dropping or weakening a binding obligation — exceeds 4,367 UTF-16, the run phase shall stop before the rule edit and report the measured floor to the leader (the parent SPEC's REQ-ALB-022 pattern); the budget shall not be met by dropping obligations.

### C.2 스텁 게이트

- **REQ-RIB-005** (Ubiquitous) — A template test shall enumerate every `*-core.md` file under the deployed and template rule trees mechanically (`*-core.md` glob, not a hand-written list — a new stub is covered on arrival) and shall fail when any enumerated stub (a) carries any `moai:role-core` region marker, or (b) exceeds 10,000 UTF-16 code units (current: `factory-dispatch-core.md` 8,583, `cross-session-messaging-core.md` 8,072; budgets leave ≥19% headroom).

### C.3 증상 B — 버전 불일치

- **REQ-RIB-006** (Event-driven) — When the role-core build fails with the missing-markers error and the deployed `system.yaml` `template_version` differs from the running binary's version, the failure detail shall name the skew (both versions) and the remediation command `moai update`; every locale's InjectionFailed operator warning shall carry the same update guidance.
- **REQ-RIB-007** (Ubiquitous) — The version-skew predicate shall be unit-tested with fixtures only — no live-repo self-assertion (this repository's own v3.1.3-stamp/r.c29-binary state is the motivating skew; a live assertion would be permanently red). The test family covers: the skew predicate's equal/unequal/unreadable cases, and the update guidance's presence in all four locales.

### C.4 잔여 정비

- **REQ-RIB-008** (Event-driven) — When `role_rules.go` is edited, the size-gate doc comment shall state the REQ-ALB-010 ladder exactly once (the :393–413 duplication removed), and the overflow NOTE directive / operator-locale wording shall be unified (thousands separators consistent with the locale warnings; the NOTE and the warning stop repeating each other's content).
- **REQ-RIB-009** (Ubiquitous) — `.claude/rules/moai/core/hooks-system.md` shall document both measured hook-output behaviors distinctly: total stdout over 50K characters saved to disk (existing line) and the 10,000-character per-`additionalContext`-string delivery cap (decision-index Q4, observed by the lane-15 notice), each with its measurement citation.
- **REQ-RIB-010** (Ubiquitous) — After the rewrite, the local rule's citation (`gitflow-lane-protocol.md` §1 → "factory-dispatch.md의 Isolation 절") shall resolve to a live section still carrying the mid-session-move prohibition, and the git-flow variant sentence (the deployed copy's :263 region) shall stay consistent with the local rule's disposition; the deployed copy's local sweep-tail sentence (the +321 divergence) is a relocation-audit row like any other chunk.

### C.5 Template-First 와 비용 기록

- **REQ-RIB-011** (Ubiquitous) — Every content change shall originate under `internal/template/templates/`, be embedded through `make build`, and be mirrored to this repository's `.claude/rules/` copy within the same run, with changed rule pairs registered in the rule-template mirror test; the run shall record the before/after UTF-16 byte deltas of every always-loaded file it touches (the two stubs chiefly) and state the rule-authoring cost justification for any growth over 1,000 bytes.

## §D. 제약

- **한도 인상 금지** — 10,000은 decision-index Q4 로 올릴 수 없다. 이 SPEC 이 만지는 것은 조립 합본의 크기뿐이다.
- **오버플로 경로 보존** — REQ-ALB-010 의 사다리(오버플로 파일 전달 → REQ-ALB-009 후퇴)는 메커니즘으로서 그대로다. 완료 기준 1의 "오버플로 파일 0"은 발화가 없다는 뜻이지 경로 삭제가 아니다.
- **역할 분할 기각** — leader/lane 전용 core 분할(buildRoleCore 역할 필터 추가)은 기각한다. 산술 이유: 완전한 역할 분할이어도 각 역할은 core의 절반 ~9,057을 보게 되고, 그것도 예산 4,367을 2배 이상 넘는다 — 역할 분할은 예산에 필요한 4.15배 압축을 대신하지 못한다. 하나의 공유 core 를 예산 안으로 재작성하는 (i)이 정답이고, `buildRoleCore` 의 표지 검증(균형·순서·열림 검사)은 전부 보존된다.
- **구속 절의 이동 반경** — 구속 절은 core(압축 재작성) 또는 상시 스텁에만 남는다. rationale·절차 중복(이미 companion 이 본문을 가진 것)만 core에서 떨어진다. 스텁으로의 대규모 재분류(상시 표면 성장)는 금지 — 상위 다이어트의 성과를 되돌리는 것이다.
- **로컬 갈림 처분** — 배포본 마지막 영역의 +321 꼬리 문장은 로컬 전용 내용이다. 재작성은 이를 재배치 감사 행으로 다룬다(본문이 `worktree-integration.md` § Hoist 에 이미 있음을 확인하고 core에서 압축·정리). 템플릿 미러는 로컬 꼬리를 갖지 않는 현 상태를 유지한다.
- **재측정 규율** — §B 의 고정 수치를 인용하는 AC 는 `verification-claim-integrity.md` §2/§4 에 따라 이 SPEC dir 의 `relocation-ledger.md` 와 `.moai/reports/t1617/measurements.md` 에 커밋 SHA와 함께 귀속된다. 생산자 기준선 4,376은 lane-15 직접 분해치로 귀속되며(라이브 세션 자산, 재측정 불가), 합본 교차 검산과 설계 목표 4,000이 ±100 노이즈를 흡수한다.
- Codex 하네스는 `.claude/rules/` 를 읽지 않으므로 이 SPEC 의 영향 밖이다.

## §E. 수용 기준

인수 조건 전체는 `acceptance.md` 에 있다(AC-RIB-001 ~ AC-RIB-012, 두 칸 채택 — RED-now + green 경로).

## §F. 범위 밖

### Out of Scope — 상수와 메커니즘

- `roleRulesContextLimit` 인상. decision-index Q4 — 올릴 수 없다.
- `roleRuleSizeGate` 사다리·`FinalizeSessionStartOutput` 구조 변경. 오버플로 경로는 보존된다(§D).
- `internal/hook/instructions_loaded.go` 의 `sessionCharBudget`(210,000). 상위 SPEC §F 가 이미 범위 밖으로 확정한 운영자 판정값.

### Out of Scope — 상위 SPEC 의미론

- 상위 SPEC `SPEC-ALWAYS-LOADED-BUDGET-001` 본문의 수정. 그 원장은 테스트 고정물이고 규칙 편집 후 원장 행 갱신은 REQ-ALB-015 의 상시 유지보수 계약 이행이다(§B 상위 원장) — 완결 SPEC 의 재계획·amendment 가 아니다. `companion:` 행 vocabulary 확장도 하지 않는다(§D: 구속 절은 core/스텁에만).
- `cross-session-messaging.md` 역할 core 의 내용 변경. 빈 영역(포인터 전용)이 예산에 가장 유리한 형태로 유지된다.

### Out of Scope — 로컬 규칙의 흐름 내용

- `gitflow-lane-protocol.md` 의 git-flow→GitHub Flow 전환 서술 정리(develop→main 등). 2026-10-05 GitHub Flow 전환(t1453 계열)의 몫이며 이 SPEC 은 **포인터 정합**(REQ-RIB-010)만 맡는다.
- `hooks-system.md` 의 50K/10K 외 다른 한도 서술 정리.

### Out of Scope — 스텁 재분류

- 역할 core 의 구속 절을 스텁으로 대량 이동. 상시 표면을 모든 세션이 다시 내는 일이 된다(§D). REQ-RIB-011 의 스텁 델타 기록은 소규모 변동까지만 예상한다.

## §G. 관련 관계

- `SPEC-ALWAYS-LOADED-BUDGET-001` — 상위. REQ-ALB-007~011, REQ-ALB-009/010(오버플로·실패 경로), REQ-ALB-023(배포 트리 전용 생성), REQ-ALB-015(원장 정합)를 **승계**하며, 그 REQ-ALB-010 이 허용한 오버플로 전달이 실제로 발화한 상태(23,166)를 예산 안으로 되돌리는 후속 수리다. 상위의 구속 원장 유지보수 계약 위에서 움직인다(§B).
- `SPEC-ALWAYS-LOADED-DIET-001/002`, `SPEC-ALWAYS-LOADED-HEADROOM-001` — 항상 로드 표면 축소 계보. 이 SPEC 은 상시 표면을 키우지 않는 방향(§D)으로 그 성과를 보존한다.
- `SPEC-INSTRUCTIONS-BUDGET-001` — 세션 문자 예산의 축. 이 SPEC 은 SessionStart 주입 경로만 다룬다.
- `rule-authoring.md` — 상시 로드 4슬롯 성장 의무. REQ-RIB-011 이 이행한다(스텁이 paths:-scoped companion 이 아니라 상시 파일이라는 점이 이 SPEC 의 유일한 걸리는 지점).

## §H. 교차 참조

- 카드 t1617 · 리더 처분 dd1e · no-new-dispatch 규칙 fde3 예외 · 런 tmnboq
- decision-index Q4 (10,000 비가반 상한, 상위 SPEC)
- `internal/hook/role_rules.go` / `internal/config/role_markers.go` / `internal/hook/session_start.go` — 주입·게이트·표지 코드
- `internal/template/testdata/binding_ledger.json` — 상위 원장(36개 role-core 행)
- `.moai/reports/t1617/measurements.md` — plan 단계 전수 재측정 원문
- `.moai/specs/SPEC-ROLE-INJECTION-BUDGET-001/relocation-ledger.md` — 재배치 감사 원장(run 단계에서 작성)
