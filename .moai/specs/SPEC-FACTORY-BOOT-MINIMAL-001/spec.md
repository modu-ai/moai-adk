---
id: SPEC-FACTORY-BOOT-MINIMAL-001
title: "팩토리 시작 메시지 최소화 — 실측 생산자 기반 조립 예산 9,146 회계, AGENTS.md 바이트 여유 확보"
version: "0.1.0"
status: draft
created: 2026-10-23
updated: 2026-10-23
author: manager-spec
priority: P0
phase: "v3.2.0"
module: "internal/hook, AGENTS.md"
lifecycle: spec-anchored
tags: "factory, session-start, bootstrap-notice, role-injection, budget, i18n, agents-md, codex-byte-ceiling"
tier: M
related_specs: [SPEC-ROLE-INJECTION-BUDGET-001, SPEC-FACTORY-BOOTSTRAP-001, SPEC-SESSION-START-GUIDE-I18N-001, SPEC-FACTORY-SELF-DISPATCH-001]
---

# SPEC-FACTORY-BOOT-MINIMAL-001

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-10-23 | 0.1.0 | 최초 작성. 카드 t1923 (운영자 지시 10-21 · Class C · P0 · 3.2 필수 차단; 10-22 운영자 정정으로 byte-budget 부분 필수 차단 승격). 기준 트리 `59959629908cda723c37e26c7c15718194d6b38a` (= origin/main, 브랜치 WT-factory-lane-internal). plan 단계 실측 원문은 `research` 절(plan.md §B)과 `probe/` 증거 파일. | manager-spec |

## §A 배경과 실측 사실

### §A.1 카드가 관측한 23,145자 넘침의 실제 원인

카드는 「factory-lane 세션 시작 조립 맥락 23,145자가 10,000자 전달 한도(`internal/hook/role_rules.go:52` `roleRulesContextLimit`)를 넘었다」고 관측했다. plan 단계에서 그 세션의 훅 출력을 직접 분해했다(plan.md §B.3, 명령과 출력 원문 포함). 결과:

- 23,145자 넘침(2026-10-21T05:41Z, lane 6개 동시 기동)은 **역할 블록 18,367자**가 원인이었다. 이것은 SPEC-ROLE-INJECTION-BUDGET-001(t1617)이 2026-10-10에 4,0xx자로 압축하기 **전의** role core다. 세션의 작업 디렉터리는 primary checkout이었고, 그 시점의 primary checkout은 10-21 사고(primary가 local `main`보다 559커밋 뒤에 detach된 상태)로 낡은 규칙 파일을 서빙했다.
- 2026-10-21T06:14Z 이후 관측된 role 세션 5건은 전부 넘침 없음, 최대 합본 9,071자(lane·ko·startup), 역할 블록 4,293자다.

따라서 「role core 압축」은 이미 완료된 일이고, 남은 문제는 **역할 블록 밖 고정분**이다.

### §A.2 지금 트리에서 아직 위험한 칸 — 영어 레인

기준 트리에서 저장소 내 실조립 테스트 `TestCompositeBudgetRealAssembly`(role_composite_e2e_test.go)는 `lane_en_glm` 합본 **9,946자(여유 54자)**를 보고하고, 실측 비-팩토리 생산자 허용분을 쓴 plan 단계 탐침은 lane·en·startup 합본 **10,006자**(10,000 초과)를 보고한다. 원인은 영어 lane 고정분 3,751자(부팅 공지 1,036 + 레인 규칙 2,713)다. 현재 예산 테스트 `TestRoleInjectionAssemblyBudget`는 생산자를 실제 렌더가 아니라 합성 상수 4,797자로 모델링하므로 이 위험을 보지 못하고 PASS한다(composite 9,100/9,102 ≤ 9,146).

### §A.3 AGENTS.md 바이트 상한 — 이미 녹색, 여유 75바이트

카드의 「main@3f81377dc9 ubuntu CI에서 `TestCodexContractByteCeiling`·`TestCodexNestedTemplateDiscoveryBudget` FAIL」은 AGENTS.md 24,807바이트 시점의 관측이다. 이후 `a1762cabe1`(#1869)이 AGENTS.md를 24,501바이트로 줄였고, 기준 트리에서 두 테스트는 PASS다(상한 24,576, 여유 75). 운영자 제안 패치는 24,807바이트 판을 기준으로 만들어져 기준 트리에 그대로 적용되지 않는다(`git apply --check` exit 1). 같은 문단 교체를 기준 트리에 옮기면 24,218바이트(여유 358)가 된다. 패치 검토 결과 의미 손실은 한 군데다: 「WT worktrees live **only** under」의 배타성(`only`)이 「belong under」로 약해진다. 사고 기록 괄호는 근거 문장이라 제거 가능하다(AGENTS.md §0 「압축은 근거와 사고 기록만 제거하고 의무는 제거하지 않았다」 규약).

### §A.4 Tier 판단

Tier **M**. Go 소스 2개(`session_start_factory.go`, `session_start_factory_i18n.go`) + 테스트 4~6개 + 재배치 목적지 문서(템플릿·배포 쌍) + AGENTS.md 1개로 5~15파일, 300~1,000 LOC 범위다. 단일 서브시스템(SessionStart 주입 경로)이고 헌법급 변경이 아니므로 L이 아니다. Go 코드·테스트·상시 로드 파일 편집이 함께 있어 S도 아니다.

## §B 범위와 예산 회계

### §B.1 예산 사슬 — 0.6.0 수치는 바꾸지 않는다

SPEC-ROLE-INJECTION-BUDGET-001 0.6.0 계약(합본 9,146 = 생산자 4,797 + 오버헤드 257 + core 4,092)을 **수치 그대로 유지**한다. 바뀌는 것은 생산자 4,797의 회계 방식이다:

| 항목 | 0.6.0 (현재) | 본 SPEC |
|---|---|---|
| 생산자 상한 | 합성 상수 4,797 (렌더와 무관) | 비-팩토리 허용분 **1,950** + 팩토리 고정분 상한 **2,847** = 4,797 |
| 팩토리 고정분 | 측정 안 함 | 부팅 공지 + 레인 규칙을 **실제 렌더러로 측정** (역할 × 소스 × 언어 × 디스패치 변형) |
| 오버헤드 | 257 | 257 (변경 없음) |
| core 상한 | 4,092 | 4,092 (변경 없음, 실측 4,048) |
| 합본 상한 | 9,146 | 9,146 (변경 없음) |
| 전달 한도 | 10,000 | 10,000 (변경 없음, 상향 불가) |

비-팩토리 허용분 1,950은 라이브 세션 실측(attribution 509 + messaging bound 111 + GLM 359 + binary-lag 421 + guard-liveness 118 + 버전 불일치 안내 411 + 결합자 12 = 1,941)을 올림한 값이다(plan.md §B.4).

### §B.2 경계 소스 봉투

SPEC-ROLE-INJECTION-BUDGET-001 §A.5의 봉투를 그대로 쓴다: startup · clear(대기 중 핸드오프 없음) · compact(armed goal 없음). 핸드오프 본문 재주입·goal 재주입·계보 배너는 봉투 밖이며 기존 넘침 파일 폴백과 운영자 가시 경고가 그대로 백스톱이다.

## §C 요구사항 (GEARS)

- **REQ-FBM-001** (Event-driven) — When a factory leader session starts with source startup, the leader bootstrap notice shall carry only these elements on both delivery channels (additionalContext and systemMessage): the run id, the leader session label, the lane-start sentence naming the `-l` entry commands, the leader socket line where the launcher captured one, one pointer line naming the on-disk location of the relocated guidance, and the online docs URL `https://adk.mo.ai.kr`.
- **REQ-FBM-002** (Event-driven) — When a factory lane session starts with source startup, the lane bootstrap notice shall carry only these elements: the join line naming the lane label, the standing spawn-authority sentence (its pinned markers and the Status Transition Ownership Matrix path unchanged), one pointer line naming the on-disk location of the relocated guidance, and the online docs URL.
- **REQ-FBM-003** (Ubiquitous) — Every element removed from the bootstrap notice or the lane rule (at minimum the two-flag auto-mode explanation, card routing, stagger, batch gate summary, operational-status caveat, entry-point guide, fan-out cap, and lane-rule detail prose) shall land, before removal, in an existing on-disk guidance location outside the two role-gated rule files, and the pointer line of REQ-FBM-001/002 shall name that location.
- **REQ-FBM-004** (Ubiquitous) — The factory fixed part — the bootstrap notice plus the lane rule rendered by the production renderers — shall stay at or under 2,847 UTF-16 code units for every combination of role (leader, lane), start source (startup, clear, compact), conversation language (en, ko, ja, zh), and lane dispatch variant (auto, manual).
- **REQ-FBM-005** (Ubiquitous) — The assembled composite of non-factory allowance (1,950 UTF-16) + joiner + factory fixed part + role block shall stay at or under 9,146 UTF-16 code units for every combination named in REQ-FBM-004, so that no session inside the bounded-source envelope (§B.2) reaches the 10,000-unit delivery cap or emits the overflow notice.
- **REQ-FBM-006** (Ubiquitous) — The lane rule shall keep the protocol tokens other SPECs pin: the six MCP tool names with their CLI equivalents (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019), the `moai factory next` / `moai factory complete` verbs, and the stall-recovery tokens (`CronCreate`, `CronList`, `CronDelete`, `moai-lane-watchdog`, the rendered cron expression).
- **REQ-FBM-007** (Ubiquitous) — The role core shall stay lossless: this SPEC's change set shall not modify `.claude/rules/moai/workflow/factory-dispatch.md` or `.claude/rules/moai/workflow/cross-session-messaging.md` in either the deployed tree or the template tree, and the role-core ceiling 4,092 shall remain unchanged.
- **REQ-FBM-008** (Event-driven) — When the hook package's tests run, the in-tree budget tests (`TestRoleInjectionAssemblyBudget` and `TestCompositeBudgetRealAssembly`) shall measure the factory fixed part from the production renderers for every combination in REQ-FBM-004 (both deployment faces), shall fail on any breach of REQ-FBM-004/005 naming the breakdown, and shall record in the test header that the 0.6.0 producer constant 4,797 is now split into allowance 1,950 + factory fixed ceiling 2,847.
- **REQ-FBM-009** (Unwanted) — The change set shall not delete, skip, or weaken a budget or ceiling check to reach green: no `t.Skip` is added to the budget, guide, or Codex-ceiling tests, `CodexContractByteCeiling` (24,576) and `roleRulesContextLimit` (10,000) are not raised, and a guide test whose old contract is superseded by REQ-FBM-001/002 is rewritten to assert the new contract rather than removed.
- **REQ-FBM-010** (Ubiquitous) — The root `AGENTS.md` shall carry at least 300 bytes of headroom under the Codex contract ceiling (at or under 24,276 bytes), reached by applying the operator's reviewed rewording of the two worktree paragraphs onto the current tree with the exclusivity word "only" preserved, with every obligation of those paragraphs kept.
- **REQ-FBM-011** (Event-driven) — When the landed binary is installed and six lanes plus one leader start, each session's SessionStart composite shall be at or under 10,000 UTF-16 code units with zero overflow notices, as observed from the session transcripts.

수용 기준은 acceptance.md(AC-FBM-001 ~ AC-FBM-009)에 둔다.

## §D 설계 결정 요약

- 공지 최소화의 기준 원소 목록은 카드 원문(「run id·슬롯 라벨·스폰 권한 한 줄·채널 포인터·docs URL」)을 따르되, 리더의 레인 시작 문장(`-l` 명령)은 운영자가 레인 터미널을 여는 유일한 안내라 유지한다(decision-index Q2). 「채널 포인터」는 리더 소켓 줄(이미 있는 채널 주소)과 재배치 안내 위치 포인터 한 줄로 해석한다(Q6).
- 스폰 권한 문장은 손대지 않는다(Q3). t224 사고 이후 고정된 문장이고 테스트 6개 표지가 묶여 있다.
- 재배치 목적지는 역할 게이트 규칙 2개가 아닌 상세 동반 문서다(Q5). 이렇게 하면 role core 무손실(REQ-FBM-007)을 `git diff`로 기계 확인할 수 있고, t1637 (d)의 role-core 행 복원 작업과 겹치지 않는다.
- 0.6.0 예산 수치는 유지하고 생산자 회계만 실측으로 바꾼다(Q4).

## §E 범위 밖

### Out of Scope — t1637 (d)와의 분리

- role core 안으로 t1617 OUT 행을 되돌리는 작업(t1637 (d) 후반부)은 하지 않는다. 본 SPEC이 고정분을 줄여 생기는 여유(실측)는 t1637이 core 상한 개정으로 청구할 수 있다.
- t1637이 이어받은 t1617 부채(`detectRoleRuleVersionSkew` 문자열 절단, `os.ReadFile` FIFO, `TestCompositeBudgetMeasurement` 추정 셀의 OVER CAP 로그, binding ledger 본문 행 유실 탐지)는 다루지 않는다.

### Out of Scope — t1780 및 훅 실행 방식

- 훅의 동기/비동기 전환(t1780)은 다루지 않는다. 본 SPEC은 SessionStart 출력 텍스트의 크기만 다룬다.

### Out of Scope — 봉투 밖 생산자와 낡은 배포본

- 핸드오프 본문 재주입, goal 재주입, 계보 배너의 크기 상한은 다루지 않는다(기존 폴백 유지).
- 낡은 배포 규칙 파일(예: 10-21 primary checkout)이 큰 core를 서빙하는 경우의 자동 교체는 다루지 않는다. t1729 감사가 버전 문자열로 빌드 순서를 증명할 수 없다는 이유로 기각한 설계이며, 넘침 파일 폴백과 운영자 경고가 그 경우의 백스톱이다.

### Out of Scope — 상한 상수와 템플릿 AGENTS.md

- `CodexContractByteCeiling`, `roleRulesContextLimit`, core 상한 4,092의 상향.
- `internal/template/templates/AGENTS.md.tmpl` 편집. 교체 대상 두 문단은 루트 AGENTS.md에만 있다(템플릿 측 grep 0건, plan.md §B.5).
