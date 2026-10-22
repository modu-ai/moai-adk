---
id: SPEC-FACTORY-BOOT-MINIMAL-001
title: "팩토리 시작 메시지 최소화 — 실측 생산자 기반 조립 예산 9,146 회계, AGENTS.md 바이트 여유 확보"
version: "0.2.0"
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
related_specs: [SPEC-ROLE-INJECTION-BUDGET-001, SPEC-FACTORY-BOOTSTRAP-001, SPEC-SESSION-START-GUIDE-I18N-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-AUTONOMY-BATCH-GATE-001]
---

# SPEC-FACTORY-BOOT-MINIMAL-001

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-10-23 | 0.1.0 | 최초 작성. 카드 t1923 (운영자 지시 10-21 · Class C · P0 · 3.2 필수 차단; 10-22 운영자 정정으로 byte-budget 부분 필수 차단 승격). 기준 트리 `59959629908cda723c37e26c7c15718194d6b38a` (= origin/main, 브랜치 WT-factory-lane-internal). plan 단계 실측 원문은 `research` 절(plan.md §B)과 `probe/` 증거 파일. | manager-spec |
| 2026-10-23 | 0.2.0 | plan-audit 1회차 FAIL(0.69) 수정: 리더 공지는 완료 SPEC이 고정한 원소(SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017 배치 게이트 문장, t85 디스패치 규율)를 그대로 두고 자동 모드 설명만 옮긴다(D1); 재배치 R1~R5와 목적지를 확정하고 REQ-FBM-003을 그 표에 묶음(D4); 고정분 상한을 결합자 2를 넣어 2,845로 재도출(D6); AC-FBM-001·002·004를 변이 차단형으로 강화하고 AC-FBM-010~012 추가(D2·D3·D5). | manager-spec |

## §A 배경과 실측 사실

### §A.1 카드가 관측한 23,145자 넘침의 실제 원인

카드는 「factory-lane 세션 시작 조립 맥락 23,145자가 10,000자 전달 한도(`internal/hook/role_rules.go:52` `roleRulesContextLimit`)를 넘었다」고 관측했다. plan 단계에서 그 세션의 훅 출력을 직접 분해했다(plan.md §B.3, 명령과 출력 원문 포함). 결과:

- 23,145자 넘침(2026-10-21T05:41Z, lane 6개 동시 기동)은 **역할 블록 18,367자**가 원인이었다. 이것은 SPEC-ROLE-INJECTION-BUDGET-001(t1617)이 2026-10-10에 4,0xx자로 압축하기 **전의** role core다. 세션의 작업 디렉터리는 primary checkout이었고, 그 시점의 primary checkout은 10-21 사고(primary가 local `main`보다 559커밋 뒤에 detach된 상태)로 낡은 규칙 파일을 서빙했다.
- 2026-10-21T06:14Z 이후 관측된 role 세션 5건은 전부 넘침 없음, 최대 합본 9,071자(lane·ko·startup), 역할 블록 4,293자다.

따라서 「role core 압축」은 이미 완료된 일이고, 남은 문제는 **역할 블록 밖 고정분**이다.

### §A.2 지금 트리에서 아직 위험한 칸 — 영어 레인

기준 트리에서 저장소 내 실조립 테스트 `TestCompositeBudgetRealAssembly`(role_composite_e2e_test.go)는 `lane_en_glm` 합본 **9,946자(여유 54자)**를 보고하고, 실측 비-팩토리 생산자 허용분을 쓴 plan 단계 탐침은 lane·en·startup 합본 **10,006자**(10,000 초과)를 보고한다. 원인은 영어 lane 고정분 3,751자(부팅 공지 1,036 + 결합자 2 + 레인 규칙 2,713)다. 현재 예산 테스트 `TestRoleInjectionAssemblyBudget`는 생산자를 실제 렌더가 아니라 합성 상수 4,797자로 모델링하므로 이 위험을 보지 못하고 PASS한다(composite 9,100/9,102 ≤ 9,146).

리더 칸은 위험하지 않다. 리더 영어 고정분은 2,573으로 고정분 상한 2,845 아래이고(plan.md §B.2), 예산 때문에 리더 공지에서 무엇을 빼야 할 이유는 없다. 리더 공지에서 빠지는 것은 카드가 이름으로 지목한 자동 모드 2플래그 해설 하나다(§B.3 R1).

### §A.3 AGENTS.md 바이트 상한 — 이미 녹색, 여유 75바이트

카드의 「main@3f81377dc9 ubuntu CI에서 `TestCodexContractByteCeiling`·`TestCodexNestedTemplateDiscoveryBudget` FAIL」은 AGENTS.md 24,807바이트 시점의 관측이다. 이후 `a1762cabe1`(#1869)이 AGENTS.md를 24,501바이트로 줄였고, 기준 트리에서 두 테스트는 PASS다(상한 24,576, 여유 75). 운영자 제안 패치는 24,807바이트 판을 기준으로 만들어져 기준 트리에 그대로 적용되지 않는다(`git apply --check` exit 1). 같은 문단 교체를 기준 트리에 옮기고 「live only under」의 `only`를 되살리면 24,221바이트(여유 355)가 된다(plan.md §B.5, `probe/research/agents-mutants-5995962990.txt`). 패치 검토 결과 의미 손실은 한 군데다: 「WT worktrees live **only** under」의 배타성이 「belong under」로 약해진다. 사고 기록 괄호는 근거 문장이라 제거 가능하다(AGENTS.md §0 「압축은 근거와 사고 기록만 제거하고 의무는 제거하지 않았다」 규약).

### §A.4 Tier 판단

Tier **M**. Go 소스 2개(`session_start_factory.go`, `session_start_factory_i18n.go`) + 테스트 4개 + AGENTS.md 1개로 5~15파일, 300~1,000 LOC 범위다. 재배치 목적지 문서는 이미 있어 새로 쓰지 않는다(§B.3). 단일 서브시스템(SessionStart 주입 경로)이고 헌법급 변경이 아니므로 L이 아니다. Go 코드·테스트·상시 로드 파일 편집이 함께 있어 S도 아니다.

## §B 범위와 예산 회계

### §B.1 예산 사슬 — 0.6.0 수치는 바꾸지 않는다

SPEC-ROLE-INJECTION-BUDGET-001 0.6.0 계약(합본 9,146 = 생산자 4,797 + 오버헤드 257 + core 4,092)을 **수치 그대로 유지**한다. 바뀌는 것은 생산자 4,797의 회계 방식이다. 조립 순서는 「허용분 + `\n\n` + 팩토리 고정분 + `\n\n` + 역할 블록(core + 래퍼 255)」이므로 생산자 안에도 결합자 2가 하나 들어간다:

| 항목 | 0.6.0 (현재) | 본 SPEC |
|---|---|---|
| 생산자 상한 | 합성 상수 4,797 (렌더와 무관) | 비-팩토리 허용분 **1,950** + 결합자 **2** + 팩토리 고정분 상한 **2,845** = 4,797 |
| 팩토리 고정분 | 측정 안 함 | 부팅 공지 + 레인 규칙을 **실제 렌더러로 측정** (역할 × 소스 × 언어 × 디스패치 변형 × 배포 면) |
| 오버헤드 | 257 | 257 = 결합자 2 + 역할 블록 래퍼 255 (변경 없음) |
| core 상한 | 4,092 | 4,092 (변경 없음, 실측 4,048) |
| 합본 상한 | 9,146 | 9,146 (변경 없음) — 1,950 + 2 + 2,845 + 2 + (4,092 + 255) = 9,146 |
| 전달 한도 | 10,000 | 10,000 (변경 없음, 상향 불가) |

비-팩토리 허용분 1,950은 라이브 세션 실측(attribution 509 + messaging bound 111 + GLM 359 + binary-lag 421 + guard-liveness 118 + 버전 불일치 안내 411 + 결합자 12 = 1,941)을 올림한 값이다(plan.md §B.3).

### §B.2 경계 소스 봉투

SPEC-ROLE-INJECTION-BUDGET-001 §A.5의 봉투를 그대로 쓴다: startup · clear(대기 중 핸드오프 없음) · compact(armed goal 없음). 핸드오프 본문 재주입·goal 재주입·계보 배너는 봉투 밖이며 기존 넘침 파일 폴백과 운영자 가시 경고가 그대로 백스톱이다.

### §B.3 재배치 목록 R1~R5 — 이 SPEC이 빼는 것의 전부

빼는 원소는 아래 다섯이 전부다. 목적지는 모두 기준 트리에 **이미 같은 내용으로 있다**(배포·템플릿 두 면, `git grep` 실측은 acceptance.md E10). 새 목적지 문안을 쓰지 않으므로 「제거 전에 목적지에 있다」는 순서는 기준 커밋이 제거 커밋의 조상이라는 커밋 그래프로 증명된다.

| ID | 빼는 원소 | 빠지는 곳 | 목적지(두 면) | 목적지 표지 | 빠진 뒤 사라져야 할 표지 |
|---|---|---|---|---|---|
| R1 | 자동 모드 2플래그 해설 | 리더 공지, 레인 공지 | `.claude/skills/moai/workflows/gtd.md` (「Two explicit mode flags」 문단) | `--auto-leader`, `--auto-lane`, `stays for compatibility` | `--auto-leader`, `--auto-lane`, `shipped — t1814` |
| R2 | 착지 절차 상세(「Landing is yours …」) | 레인 자동 규칙 | `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § The lane's standard landing | `## The lane's standard landing`, `codex_review`, `Merge Risk`, `moai worktree sweep` | `codex_review`, `Merge Risk`, `moai worktree sweep` |
| R3 | 리더 잔여 역할 목록(「The leader keeps only …」) | 레인 자동 규칙 | 같은 파일 § The leader's remaining role | `## The leader's remaining role`, `Keep-set gate relay` | `keep-set` |
| R4 | 판단 사다리 상세(「Judgments are yours … `jev_ask` …」) | 레인 자동 규칙 | `.claude/rules/moai/workflow/auto-semantics.md` § 6. The decision ladder | `## 6. The decision ladder`, `jev_ask` | `jev_ask` |
| R5 | 배정 카드 임대 형식(`moai factory next --card <id>` 절) | 레인 자동 규칙 | `.claude/skills/moai/workflows/gtd.md` | `moai factory next --card <id>` | `--card` |

리더 공지의 나머지 원소(카드 라우팅, stagger, 배치 게이트 요약, 운영 상태, 진입점 안내, 팬아웃 상한, 설정 안내)는 빼지 않는다. 그중 배치 게이트 문장은 완료 SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017(운영자 결정 Q5)이, 카드 라우팅·stagger 문구는 t85 성문화 테스트가 고정하고 있고, 리더 칸은 예산 압력이 없다(§A.2). decision-index Q8.

## §C 요구사항 (GEARS)

- **REQ-FBM-001** (Event-driven) — When a factory leader session starts with source startup, the leader bootstrap notice shall carry, on both delivery channels (the agent-facing additionalContext and the operator-facing systemMessage copy), exactly these elements and nothing else: every element it carries at the base tree except the two-flag auto-mode explanation — the run-id header, the leader identity sentence, the lane-start sentence naming the `-l` entry commands, the entry-point guide, the per-lane fan-out line, the leader socket line where the launcher captured one, the card-routing discipline, the stagger discipline, the batch gate summary sentence of SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017, the operational-status caveat, the injected-settings line, and the docs sentence carrying `https://adk.mo.ai.kr`, each verbatim as at the base tree — plus one pointer line naming the repository path of the relocated auto-mode guidance.
- **REQ-FBM-002** (Event-driven) — When a factory lane session starts with source startup, the lane bootstrap notice shall carry exactly these elements and nothing else: the join line naming the lane label, the standing spawn-authority sentence, and the docs sentence, each verbatim as at the base tree, plus one pointer line naming the repository path of the relocated auto-mode guidance.
- **REQ-FBM-003** (Ubiquitous) — Each element this SPEC removes from a bootstrap notice or from the lane rule shall be exactly one of the relocations R1-R5 of §B.3, shall be present in its destination document on both the deployed and the template face — a destination outside the two role-gated rule files, carrying the element before the commit that removes it — and the startup text of the surface it was removed from shall name that destination by its repository path.
- **REQ-FBM-004** (Ubiquitous) — The factory fixed part — the bootstrap notice plus the lane rule rendered by the production renderers — shall stay at or under 2,845 UTF-16 code units for every combination of role (leader, lane), start source (startup, clear, compact), conversation language (en, ko, ja, zh), lane dispatch variant (auto, manual), and deployment face (deployed, template).
- **REQ-FBM-005** (Ubiquitous) — The assembled composite of non-factory allowance (1,950 UTF-16) + joiner + factory fixed part + joiner + role block shall stay at or under 9,146 UTF-16 code units for every combination named in REQ-FBM-004, so that no session inside the bounded-source envelope (§B.2) reaches the 10,000-unit delivery cap or emits the overflow notice.
- **REQ-FBM-006** (Ubiquitous) — The lane rule shall keep the protocol tokens and phrases other SPECs pin, in every locale: the six MCP tool names with their CLI equivalents and the per-locale operator-authorization phrase (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019), the `TaskCreate` / `TaskUpdate` discipline (SPEC-WORKFLOW-TASKS-001), the card-takeover recheck trigger, `ExitWorktree(keep)` and the relaunch exception (SPEC-FACTORY-LANE-NO-CLEAR-001), the `moai factory next` / `moai factory complete` verbs, and the stall-recovery tokens (`CronCreate`, `CronList`, `CronDelete`, `moai-lane-watchdog`, the rendered cron expression).
- **REQ-FBM-007** (Ubiquitous) — The role core shall stay lossless: this SPEC's change set shall not modify `.claude/rules/moai/workflow/factory-dispatch.md` or `.claude/rules/moai/workflow/cross-session-messaging.md` in either the deployed tree or the template tree, and the role-core ceiling 4,092 shall remain unchanged.
- **REQ-FBM-008** (Event-driven) — When the hook package's tests run, the in-tree budget tests (`TestRoleInjectionAssemblyBudget` and `TestCompositeBudgetRealAssembly`) shall measure the factory fixed part from the production renderers for every combination in REQ-FBM-004, shall report each measured combination and the producer split (allowance 1,950 + joiner 2 + factory fixed ceiling 2,845 = 4,797) in their verbose output, and shall fail on any breach of REQ-FBM-004/005, naming the combination and the measured value.
- **REQ-FBM-009** (Unwanted) — The change set shall not delete, skip, or weaken a budget, ceiling, or notice-contract check to reach green: no `t.Skip` is added, `CodexContractByteCeiling` (24,576) and `roleRulesContextLimit` (10,000) are not raised, the only tests rewritten are the seven auto-mode guide tests whose contract REQ-FBM-001/002 supersede (`TestRedAutoModeGuidancePresent`, `TestFactoryGuideNewFieldsCompleteInEveryLocale`, `TestFactoryGuideAutoModeMarkerBlocksEveryLocale`, `TestFactoryGuideFallbackCompleteNotice`, `TestFactoryGuideNewFieldsLayoutInvariants`, `TestFactoryGuideAutoLeaderDescribesInSessionWorker`, `TestFactoryGuideLaneAudienceLeaderModeNoSelfReference`) plus the two budget tests of REQ-FBM-008, each rewritten to assert the new contract rather than removed, and every other test that pins notice or lane-rule text passes unmodified.
- **REQ-FBM-010** (Ubiquitous) — The root `AGENTS.md` shall carry at least 300 bytes of headroom under the Codex contract ceiling (at or under 24,276 bytes), reached only by applying the operator's reviewed rewording of the two worktree paragraphs onto the current tree with the exclusivity word "only" preserved, with every obligation of those paragraphs kept and every byte outside those two paragraphs unchanged.
- **REQ-FBM-011** (Event-driven) — When the landed binary is installed and six lanes plus one leader start, each session's SessionStart composite shall be at or under 10,000 UTF-16 code units with zero overflow notices, as observed from the session transcripts.

수용 기준은 acceptance.md(AC-FBM-001 ~ AC-FBM-012)에 둔다. REQ 11개, AC 12개(Tier M 상한 16/16 이내).

## §D 설계 결정 요약

- 공지 최소화는 카드가 이름으로 지목한 자동 모드 2플래그 해설(R1)과, 영어 레인 고정분을 상한 아래로 내리는 데 필요한 레인 규칙 상세(R2~R5)로 한정한다. 리더 공지의 나머지 원소는 완료 SPEC 계약과 t85 테스트가 고정하고 있고 예산 압력도 없으므로 둔다(decision-index Q8). 이로써 이 SPEC은 다른 완료 SPEC의 요구사항을 하나도 대체하지 않는다. 자동 모드 해설 제거는 SESSION-START-GUIDE-I18N-001의 테스트 7개가 고정한 「설명이 있다」는 기대를 뒤집지만, 그 SPEC의 REQ-004/005는 「설명을 할 경우」의 조건부 계약이라 요구사항 문장 자체와는 충돌하지 않는다(Q1).
- 카드 원소 목록의 리더·레인 비대칭(레인 공지에 run id 없음, 리더 공지에 스폰 권한 없음)은 현재 동작을 유지한다(Q9). 리더의 레인 시작 문장은 운영자가 레인 터미널을 여는 유일한 안내라 유지한다(Q2). 「채널 포인터」는 리더 소켓 줄과 재배치 위치 포인터 한 줄로 해석한다(Q6).
- 스폰 권한 문장은 손대지 않는다(Q3). t224 사고 이후 고정된 문장이고 테스트 표지가 묶여 있다.
- 재배치 목적지는 역할 게이트 규칙 2개가 아닌, 이미 같은 내용을 가진 기존 문서다(Q5, §B.3). 이렇게 하면 role core 무손실(REQ-FBM-007)을 `git diff`로 기계 확인할 수 있고, t1637 (d)의 role-core 행 복원 작업과 겹치지 않는다.
- 0.6.0 예산 수치는 유지하고 생산자 회계만 실측으로 바꾼다(Q4). 레인 규칙에서 무엇을 뺄지는 R2~R5로 고정한다(Q10).

## §E 범위 밖

### Out of Scope — t1637 (d)와의 분리

- role core 안으로 t1617 OUT 행을 되돌리는 작업(t1637 (d) 후반부)은 하지 않는다. 본 SPEC이 고정분을 줄여 생기는 여유(실측)는 t1637이 core 상한 개정으로 청구할 수 있다.
- t1637이 이어받은 t1617 부채(`detectRoleRuleVersionSkew` 문자열 절단, `os.ReadFile` FIFO, `TestCompositeBudgetMeasurement` 추정 셀의 OVER CAP 로그, binding ledger 본문 행 유실 탐지)는 다루지 않는다.

### Out of Scope — 다른 완료 SPEC의 계약

- SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017(리더 공지의 배치 게이트 문장)과 그 테스트 `TestLeaderNoticeBatchGatePointer`는 바꾸지 않는다.
- t85 디스패치 규율(카드 라우팅·stagger)과 그 테스트 `TestFactoryLeadNoticeCarriesDispatchDiscipline`(한국어 변형 포함), 진입점·팬아웃·소켓 줄과 `TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide`는 바꾸지 않는다.
- R1~R5 밖의 원소 재배치. 실행 단계에서 더 빼야 한다고 판명되면 SPEC 개정(blocker report → manager-spec) 대상이다.

### Out of Scope — t1780 및 훅 실행 방식

- 훅의 동기/비동기 전환(t1780)은 다루지 않는다. 본 SPEC은 SessionStart 출력 텍스트의 크기만 다룬다.

### Out of Scope — 봉투 밖 생산자와 낡은 배포본

- 핸드오프 본문 재주입, goal 재주입, 계보 배너의 크기 상한은 다루지 않는다(기존 폴백 유지).
- 낡은 배포 규칙 파일(예: 10-21 primary checkout)이 큰 core를 서빙하는 경우의 자동 교체는 다루지 않는다. t1729 감사가 버전 문자열로 빌드 순서를 증명할 수 없다는 이유로 기각한 설계이며, 넘침 파일 폴백과 운영자 경고가 그 경우의 백스톱이다.

### Out of Scope — 상한 상수와 템플릿 AGENTS.md

- `CodexContractByteCeiling`, `roleRulesContextLimit`, core 상한 4,092의 상향.
- `internal/template/templates/AGENTS.md.tmpl` 편집. 교체 대상 두 문단은 루트 AGENTS.md에만 있다(템플릿 측 grep 0건, plan.md §B.5).
