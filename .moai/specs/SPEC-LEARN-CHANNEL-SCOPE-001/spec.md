---
id: SPEC-LEARN-CHANNEL-SCOPE-001
title: "학습 채널 범위의 정직화 — lessons-inbox 유용성 범위의 경계 선언 + 인간 매개 루프의 학습 채널 인정"
version: "0.1.2"
status: in-progress
created: 2026-09-01
updated: 2026-09-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".moai/docs, .claude/rules/moai/core, .claude/skills/hns-lsel-curator, CLAUDE.local.md"
lifecycle: spec-anchored
tags: "learning-channel, lessons-inbox, scope-honesty, human-loop, feedback-memory, lsel, docs-only"
era: V3R6
tier: S
related_specs: [SPEC-LSEL-DRAIN-STALL-001, SPEC-LSEL-LOCAL-EVOLUTION-001]
---

# SPEC-LEARN-CHANNEL-SCOPE-001 — 학습 채널 범위의 정직화

## HISTORY

- 2026-09-01 (plan-phase, v0.1.0) 최초 작성. 카드 t260 (운영자 지정, Tier S). 카드 전문이 범위 권위다: "lessons-inbox가 담지 못하는 결함 계열 — 자동 수집 채널의 범위 갭". 카드의 경고 문장이 본 SPEC의 설계 축이다: **"이 카드는 기전 추가가 목적이 아니라 무엇이 실제로 학습되는지를 정직하게 정하는 것이 목적이다."** 따라서 본 SPEC은 새 수집·추출 기전을 만들지 않으며, 기존 채널들의 실제 역할을 측정에 귀속해 문서로 고정한다.
- 2026-09-01 (plan-phase, v0.1.1) — plan-audit iter-1 반영 (FAIL 0.63, Tier S 문턱 0.75; 감사 보고서 `.moai/reports/t260/plan-audit-iter1.md`). **D1** (critical): writer 단일성 주장("tool_failure 유일 배출")은 구조적으로 거짓 — `internal/hook/failure_observer.go`의 stub 생산자는 정확히 2개다(`:77` `tool_failure:<tool>:<sig>`, `:111` `test_fail:<pkg>:` — 후자는 `internal/hook/evidence_writer.go:583-591`의 `rec.IsTestFail` 경로로 배선, 도입 `e70c77576` SPEC-HARNESS-RATCHET-REWIRE-001 M1). 경계 주장을 **구성 기반 + 능력 정직형**으로 재프레임하고 REQ-LCS-001/§B.1/§B.2/§G/§H 동시 수정. v0.1.0의 "미구현 서술" 가설(§B.2) 철회 — constitution-detail L53의 "test failures"는 능력 기준으로 정확했다. **D2**: AC-LCS-001 판정식을 `{tool_failure, test_fail}` 패밀리-집합형으로. **D3**: AC-LCS-004 범위를 REQ-LCS-004 트리거("describes")에 정렬 + 스윕 목록의 plan.md 명시 열거(§A.5 신설 — navigator 배제 선언·apply_test.sh 등 비-claim 표면 배제). **D4**: curator SKILL.md:34 스테일 카운트(624)를 M2 편집 목록에 추가. **D5**: §B.1에 문서-수준 tree SHA 핀 추가. **D6**: 미러 배제 목록에 card ids 추가. **D7**: 하중 AC에 RED-now 셀(4요소, 본 라운드에서 실관측) + flipping 마일스톤 태그.
- 2026-09-02 (plan-phase, v0.1.2) — plan-audit iter-2 반영 (PASS-WITH-DEBT 0.86, D1-D7 7/7 폐쇄 판정; `.moai/reports/t260/plan-audit-iter2.md`). **N1** (major, kickoff-blocker): AC-LCS-002/003의 RED 프로브가 한국어 마커 `'인간 매개 루프'`를 영어 표면에 grep했다 — constitution-detail은 한글 0행 순수 영어, SKILL.md는 한글 2행(인용 발췌)이 전부(본 라운드 재판독)라 올바른 구현으로도 뒤집힐 수 없는 impossible 방향. claim marker를 영어 토큰 `human-mediated loop`로 정의하고 RED 표 두 행 재키잉(재관측: 양 표면 오늘 `0`/exit 1 — RED 유효, 영어 구현 시 뒤집히는 키). AC-007(한국어 표면 CLAUDE.local.md)·AC-004(파일명 마커)는 언어 정합이라 미접촉. **N2** (minor): AC-LCS-003을 제약 7 pointer-only 기조에 정렬 — "요지 문장 + anchor 포인터, 수치는 anchor doc에만".

## §A. User Story

**As a** MoAI-ADK 유지자(GOOS) whose 저장소의 가장 비싼 결함 계열들 — 아무것도 관측하지 않으면서 통과하는 검사(공허 초록), 판정 전 `t.Skip`, 0행 AC를 통과 조건으로 쓰는 관문, 둘만 고치고 셋째를 놓치는 수리, 움직이는 ref에 고정된 불변 단언, 현재값처럼 인용되는 스테일 값 — 이 어느 자동 수집 채널에도 기록되지 않는다는 사실을 문서가 말해주지 않아,
**I want** (1) 인박스(`.moai/lessons-inbox.jsonl`)가 실제로 담는 것의 경계가 능력(배선된 패밀리)과 구성(측정값)을 구분해 선언되고, (2) 인간 매개 루프(레인 발견 → 리드 판정 → auto-memory `feedback_*.md` 기록)가 그 결함 계열의 실제 학습 채널로 문서상 인정되며, (3) 기전은 0개 추가되기를,
**so that** "인박스가 학습을 담당한다"는 과신이 남지 않고, 도구 실패·테스트 실패 어느 쪽으로도 나타나지 않는 결함이 "수집되지 않은 것"이 아니라 "인간 루프로 학습되는 것"으로 정확히 분류된다.

**결과 가설:**

- 어느 문서 표면에도 인박스의 포착 범위를 과대 주장하는 문구가 남지 않는다.
- 인간 루프가 채널로 명명되고, 그 상류 증거원(카드 진행 기록, 판정 보고서)이 문서로 명시된다 — 이미 일어나는 일의 문서화일 뿐 새 흐름이 아니다.
- run-phase diff가 문서 전용임이 diff 자체로 증명된다.

## §B. Context and Background

> **본 §B 전체의 측정 baseline (D5)**: worktree `WT-learn-channel-gap` HEAD `d7ce6c6bd`에서 수행, 2026-09-01 (v0.1.1 재측정 포함). 인박스와 메모리 저장소는 primary 체크아웃의 live runtime state다 — 워크트리에 존재하지 않고, 복사해 오지도 않았다.

### §B.1 실측 — 채널별 유입 구성 (귀속 명시)

**자동 채널 (`.moai/lessons-inbox.jsonl`, primary live state):**

| 항목 | 값 | 측정 (귀속) |
|---|---|---|
| 전체 구성 | **5,942행 전부 `tool_failure`** — `{tool_failure, test_fail}` 집합 밖 행 0 | 본 라운드 재측정(v0.1.1), 2026-09-01: `jq -r '.event_key // "NO_EVENT_KEY"' <primary>/.moai/lessons-inbox.jsonl \| cut -d: -f1-2 …` 전수 — 버킷 전부 `tool_failure:*`. 측정 사슬: 5,916(오케스트레이터) → 5,919(v0.1.0) → 5,932(감사 iter-1) → 5,942(본 라운드) — append-only 성장, 구성 결론은 4중 일치 |
| `test_fail` 행 | **0행** (능력은 배선돼 있으나 이 파일에서 미관측) | `grep -c 'test_fail:' <primary>/.moai/lessons-inbox.jsonl` → 원문 출력 `0`, exit 1 (본 라운드 실행) |
| **writer 구조 (능력)** | stub 생산자는 **정확히 2개** — `internal/hook/failure_observer.go:77` (`tool_failure:<tool>:<sig>`)와 `:111` (`test_fail:<pkg>:`), 둘 다 같은 `appendLessonsInboxStub`(:129)으로 같은 인박스에 append. 제2 패밀리는 살아있는 호출 경로를 가진다: `internal/hook/evidence_writer.go:583-591`의 `logEvidence` PostToolUse에서 `rec.IsTestFail` → `extractTestPackage` → `recordTestFailEvent`. 도입 커밋 `e70c77576` (SPEC-HARNESS-RATCHET-REWIRE-001 M1) | 본 라운드 Read 직접 판독 (failure_observer.go:60-134, evidence_writer.go:578-592) — 감사 iter-1 판독과 일치 |
| 구조 vs 구성의 구분 | **배선(능력)과 구성(측정)은 다른 진술이다.** 구조적 사실: 인박스는 실패 이벤트 스텁(`tool_failure` + `test_fail` 2패밀리)만 받도록 돼 있다. 경험적 사실: 오늘 시점 구성은 100% `tool_failure`다. v0.1.0이 저지른 결함(D1)은 전자 없이 후자만으로 "단일 패밀리 독점"이라는 구조 주장을 만든 것 — 본 SPEC이 고치려는 문서 정직성 결함과 같은 형태를 본 SPEC 자신이 저지른 셈이었다 | — |
| writer 낡은 주석 (관측) | `failure_observer.go:80`의 `recordTestFailEvent` 함수 헤더 주석은 "records … to usage-log.jsonl"만 말하고 같은 함수가 :109-111에서 수행하는 인박스 스텁 append를 빠뜨린다 — 이 파일의 문서-코드 드리프트 실례 | 본 라운드 Read 판독. 정정은 kickoff 조건부 제안으로 plan.md §F M2에 올린다 (본문 결정 아님) |

**인간 채널 (auto-memory store, `/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/`):**

| 항목 | 값 | 측정 (귀속) |
|---|---|---|
| `feedback_*.md` 교훈 파일 | **164개** — 그중 2026-08-25 이후 생성 **145개** (1주) | v0.1.0 plan-phase 측정(2026-09-01): `find <store> -maxdepth 1 -name 'feedback_*.md' \| wc -l` / `-newermt 2026-08-25` 변형 — 오케스트레이터·감사 iter-1 측정(164/145)과 3중 일치 |
| 카드가 지목한 고비용 계열의 존재 | 공허 초록·뮤턴트·빈 스윕·형제 누수·이동 ref 고정·스테일 귀속 계열이 전부 기록된 교훈으로 존재 (예: `feedback_a_guard_can_be_silent_while_its_subject_degrades`, `feedback_citation_sweep_matches_shape_not_prefix`, `feedback_proving_vacuity_is_not_choosing_the_mutant_that_shows_it`, `feedback_invariant_assertions_pin_a_sha` — MEMORY.md 색인 판독) | v0.1.0 plan-phase, MEMORY.md 색인 판독 |

두 baseline이 같은 결론을 가리킨다: **도구 실패·테스트 실패 어느 쪽으로도 나타나지 않는 결함 계열은 인박스에 0행이고, 인간 루프는 1주 145건을 생산하고 있다.** 인간 루프가 "실제 학습 채널"이라는 진술은 은유가 아니라 측정값이다.

### §B.2 기존 문서 claim 목록 (grep 전수, 본 워크트리 기준)

`lessons-inbox`를 참조하는 산문 표면 전수 (본 plan-phase `grep -rln`; 전체 24파일의 성격별 분류와 배제 사유는 plan.md §A.5 — claim 표면만 여기에 둔다):

| 표면 | 현 claim | 정밀도 판정 |
|---|---|---|
| `.claude/rules/moai/core/moai-constitution-detail.md` § Lessons Protocol (템플릿 미러 존재) | "tool failures **and test failures** append structured stubs … Drain actor: the MoAI orchestrator … converting each recurring `event_key` cluster into one candidate `feedback_*.md` topic file" | **능력 기준으로 정확하다** — writer는 2패밀리(`tool_failure` + `test_fail`)를 배선돼 있고(D1 판독), 이 문구는 그 능력을 과장하지 않았다. v0.1.0의 "미구현 서술" 가설은 철회한다. **실제 갭은 다른 곳에 있다:** (1) 능력과 구성을 구분하지 않아 라이브 구성(100% `tool_failure`)이 문서로 전달되지 않고, (2) 어느 패밀리도 도구·테스트 실패로 나타나지 않는 계열을 담지 못한다는 경계 선언이 없으며, (3) 인간 루프를 학습 채널로 명명하지 않는다. |
| `.claude/skills/hns-lsel-curator/SKILL.md` (dev-only, 미러 없음) | "accumulates tool-failure stubs … (624 stubs at M1 start, re-measured — a moving target)" | 스텁 성격 서술은 대체로 정확하나(테스트 실패 경로 미언급) **:34의 라이브 카운트가 스테일**(실측 5,942 — 본 SPEC 제약 7 교리의 기존 위반 상태, D4). 경계 선언 없음. |
| `CLAUDE.local.md` §28 (로컬 전용) | 드레인 운영 절차만 서술 | 범위 선언 없음. |
| `internal/hook/failure_observer.go` | 구조적으로 2패밀리 배출(:77, :111); :80 헤더 주석은 인박스 스텁 누락(§B.1 관측 행) | 코드 — 본 SPEC 기본 무수정 (kickoff 조건부 1행 주석 제안만, §G) |

**결론(범위 갭의 실체):** 어느 산문 표면도 (1) 능력(배선된 2패밀리)과 구성(측정 100% `tool_failure`)을 구분해 말하지 않고, (2) 인박스가 담지 못하는 것 — 도구·테스트 실패 어느 쪽도 아닌 계열 — 을 말하지 않으며, (3) 인간 매개 루프를 학습 채널로 명명하지 않는다. 카드가 지목한 갭은 기능 결함이 아니라 **문서가 실제를 과소 기술하는 정직성 결함**이다 — 반대 방향(기전 추가)이 아니라 문서 정밀화로 닫힌다.

### §B.3 결정 축 요약 (상세는 plan.md §A.3)

카드가 제시한 3축에 대한 권고:

1. **(a) 인간 루프 인정 — YES, 문서 선언으로.** 측정이 이미 결론을 준다(§B.1). 기전 없음: 채널 명명 + 기록 대상은 기존 `feedback_*.md` + `MEMORY.md` 관례 그대로.
2. **(b) 카드/판정서 계열 추출 — 기전으로 하지 않는다.** 이미 일어나는 일(리드가 카드 진행 기록·판정 보고서를 읽고 교훈을 쓴다)을 문서로 명명할 뿐이다. 자동 추출 파이프라인은 카드 경고에 정면 반하는 기전 추가다.
3. **(c) 인박스 유용성 범위 문서화 — YES.** 경계 주장("실패 이벤트 스텁 2패밀리만 기록; 측정 구성 100% `tool_failure`; 도구·테스트 실패 어느 쪽도 아닌 계열은 인간 루프로 흐른다") + 능력·구성 구분 + 측정 귀속(dated baseline + 재검증 가능한 명령)을 모든 인박스 서술 표면에 정렬한다.

## §C. Requirements (GEARS)

> 요구 레이어는 GEARS. 검증 레이어(Given-When-Then)는 §I가 소유한다 — 여기에 재술하지 않는다.

- **REQ-LCS-001 (ubiquitous, 경계 주장 — 능력+구성 정직형)** — The harness documentation shall state the lessons-inbox's capture scope in capability-and-composition form: the channel records failure-event stubs only — the two wired families `tool_failure:<tool>:<sig>` and `test_fail:<pkg>:`, produced by the stub appenders at `internal/hook/failure_observer.go:77` and `:111` — its measured composition is recorded as a dated baseline (measured 2026-09-01: 100% `tool_failure`, `test_fail` 0 rows), and it shall not be presented as capturing defect families that produce neither a tool failure nor a test failure.
- **REQ-LCS-002 (ubiquitous, 인간 루프 인정)** — The harness documentation shall name the human-mediated loop (lane discovery → lead judgment → auto-memory `feedback_*.md` + `MEMORY.md` index record) as the learning channel that carries defect families invisible to both wired families — vacuous green, skip-before-verdict, empty-result-set pass conditions, sibling repair misses, moving-ref invariant assertions, and stale-value quotations.
- **REQ-LCS-003 (ubiquitous, 측정 귀속)** — The bounded scope claim shall carry its measurement basis in a durable anchor document: the live inbox composition tally (a re-runnable command + count + measurement date + tree SHA as a dated baseline), the `test_fail` 0-row observation, and the human-channel yield count, so a later reader re-verifies the claim instead of trusting it.
- **REQ-LCS-004 (capability gate, 표면 일관성)** — **Where** a prose surface describes the lessons-inbox's capture scope (constitution-detail § Lessons Protocol, `hns-lsel-curator/SKILL.md`, `CLAUDE.local.md` LSEL section — the enumerated list in plan.md §A.5), that surface shall agree with the canonical bounded claim or point to the anchor document; the surfaces shall not carry divergent scope claims.
- **REQ-LCS-005 (unwanted, 기전 금지)** — The learning architecture shall not gain a new runtime mechanism in this SPEC's scope: no stub family beyond the two already wired, no new hook or writer, no automated extraction pipeline from verdict reports or cards, and no new record format. The existing auto-memory convention (`feedback_*.md` + `MEMORY.md` index) remains the recording target, and card progress records and verdict reports remain what they already are — the upstream evidence a human-loop lesson cites.
- **REQ-LCS-006 (state-driven, 드레인 무변경)** — **While** the LSEL drain pipeline runs (wrapper → `drain.sh` → `clusters.json` staging), it shall remain behaviorally unchanged: this SPEC adds documentation only, and the inbox stays append-only.
- **REQ-LCS-007 (ubiquitous, 미러 정직성)** — Where a template-managed surface carries the recognition or bounded-claim line (`moai-constitution-detail.md`), its template mirror under `internal/template/templates/` shall carry the same principle-level line, and the mirror shall omit dev-local measurement content (dates, counts, SPEC IDs, card ids, machine-local paths) per template neutrality (C1-C8).

## §D. Constraints (HARD)

1. **문서 전용** — Go 소스 0줄. `internal/**` 무수정 (`internal/graph` 포함 — 리드의 직렬 관리 경고). 단일 예외: plan.md §F M2의 kickoff 조건부 제안(§B.1 관측 행의 :80 주석 1행 정정)이 운영자에게 채택되는 경우에 한해 `internal/hook/failure_observer.go` 주석 1행 — 미채택 시 0줄 유지.
2. **기전 추가 금지** — 세 번째 stub 패밀리, 새 훅, 새 스크립트, 새 `.moai/config/sections/` 파일, 새 기록 포맷 전부 금지 (REQ-LCS-005의 HARD화). 제2 패밀리(`test_fail`)는 이미 배선돼 있다 — "없는 기전을 추가한다"는 서술을 금지한다.
3. **미러 쌍 수정** — `moai-constitution-detail.md` 수정 시 local + `internal/template/templates/` 미러를 함께 고치고 `make build`로 임베드 재생성 (Template-First, CLAUDE.local.md §2). 미러 측 neutrality(C1-C8) 준수.
4. **LSEL 파이프라인 무변경** — `drain.sh` / `session_drain.sh` / `backlog_check.sh` / `hns-lsel-applier` 무수정. 인박스 append-only 불변 (선행 SPEC 2건의 M1 불변식 계승).
5. **항상 로드 표면 무수정** — `moai-constitution.md` 본문은 토큰 다이어트(`rule-authoring.md`)로 그대로 둔다. 인정 문장은 detail companion + anchor doc에 둔다 (kickoff에서 운영자가 뒤집으면 그때 본문 편집을 검토).
6. **레코드 포맷 불변** — `feedback_*.md` + `MEMORY.md` 관례와 `moai-memory.md` 택소노미 유지. 새 포맷·새 스키마 없음.
7. **라이브 수치의 미러 유입 금지** — dated baseline(카운트·날짜·tree SHA)은 anchor doc에만 둔다. 미러 템플릿은 원칙 문장만 (neutrality C-클래스). 산문 표면의 기존 라이브 수치(SPILL: curator SKILL.md:34의 스테일 카운트)는 M2에서 anchor 포인터로 교체한다 (D4).

## §E. Local Deliverables (PR/커밋 미탑재 — 유지자 머신 상태)

| 항목 | 대상 | 이유 |
|---|---|---|
| LSEL 운영 섹션 범위 문구 | `CLAUDE.local.md` §28 | 로컬 전용 파일(템플릿 금지, CLAUDE.local.md §2 Local-Only 목록). tracked PR에 실리지 않는 운영자-facing 표면. |

적용 증거(grep 출력)는 progress.md §E.2에 기록한다. anchor doc·constitution-detail·SKILL.md·미러는 tracked 파일로 커밋된다.

## §F. Success Criteria

- **경계 선언**: 모든 claim 표면이 능력+구성 형태의 경계 주장 또는 anchor 포인터를 가진다 — 발산 claim 0 (AC-LCS-004).
- **채널 인정**: 인간 루프가 문서상 학습 채널로 명명되고, 그 상류 증거원(카드 진행 기록·판정 보고서)이 명시된다 (AC-LCS-002).
- **재검증 가능성**: anchor doc의 명령을 다시 돌리면 같은 결론 — `{tool_failure, test_fail}` 집합 밖 0행 — 이 나온다 (AC-LCS-001).
- **기전 영(zero)**: run-phase diff가 문서+미러 전용임이 diff 자체로 증명된다 (AC-LCS-005/006).

## §G. Out of Scope

### Out of Scope — 새 수집·추출 기전 (카드 경고의 직접 적용)

- 세 번째 stub 패밀리, 새 훅 배선, 판정 보고서/카드의 자동 계열 추출 파이프라인, 새 기록 포맷/스키마 — 전부 비목표. 카드 경고("기전 추가가 목적이 아니다")가 이 절의 근거다. 참고: 제2 패밀리 `test_fail:<pkg>:`는 이미 트리에 배선돼 있다(`e70c77576`, §B.1) — 본 SPEC은 그 존재를 문서로 정직히 반영할 뿐, 패밀리를 추가하거나 제거하지 않는다. 자동 추출의 필요성이 실측되면 후속 SPEC으로 — 본 SPEC은 anchor doc에 그 경계를 non-goal로 기록해 둔다.

### Out of Scope — internal/graph 및 Go 소스 전반 (kickoff 조건부 1행 예외 제외)

- 리드의 직렬 관리 경고에 따라 `internal/graph` 0줄. `internal/hook/failure_observer.go`의 :80 낡은 함수 헤더 주석(usage-log.jsonl만 언급, :109-111의 인박스 스텁 누락 — §B.1 관측 행)의 **1행 정정은 기본 비목표**이며, plan.md §F M2에 제안으로 올려 kickoff에서 운영자가 채택할 때만 예외로 허용된다(제약 1). 미채택 시 코드는 관측만 남고 넘어간다 — 문서 표면에서의 정밀화가 본 카드의 본질이다.

### Out of Scope — drain/cluster/applier 동작 변경

- `drain.sh` 심각도 필터, 클러스터링, `session_drain.sh` wrapper, `backlog_check.sh` 임계, frozen applier — 전부 무수정 (선행 SPEC: SPEC-LSEL-LOCAL-EVOLUTION-001, SPEC-LSEL-DRAIN-STALL-001의 불변식 계승).

### Out of Scope — 항상 로드 표면(moai-constitution.md 본문) 편집

- 인정 문장을 본문에 넣으면 매 세션 비용이다(`rule-authoring.md`). detail companion(`moai-constitution-detail.md`)이 같은 § Lessons Protocol의 상세 표면이므로 거기 둔다. 운영자가 kickoff에서 본문 편집을 요구하면 그때 1문장 후보를 검토한다 — 기본은 무수정.

### Out of Scope — hns-lsel-* 16-언어 graduation

- `moai-lsel-*` 분산 graduation은 선행 SPEC 당시부터 별도 과제로 분리돼 있다 — 그대로 유지.

## §H. Cross-references

- 카드: t260 (운영자 지정, Tier S — "자동 수집 채널의 범위 갭")
- 감사: `.moai/reports/t260/plan-audit-iter1.md` (iter-1 FAIL 0.63 — D1-D7, 본 v0.1.1의 수정 경로)
- 선행: `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/` (drain 파이프라인·clusters.json 원천), `.moai/specs/SPEC-LSEL-DRAIN-STALL-001/` (내구 트리거·wrapper 규율)
- 제2 패밀리 원천: SPEC-HARNESS-RATCHET-REWIRE-001 M1 (도입 커밋 `e70c77576`), `internal/hook/failure_observer.go` (`recordToolFailureEvent` :77 / `recordTestFailEvent` :111 / `appendLessonsInboxStub` :129), `internal/hook/evidence_writer.go:583-591` (`rec.IsTestFail` 배선)
- 소관 룰: `.claude/rules/moai/workflow/moai-memory.md` (메모리 택소노미·영어 규칙), `.claude/rules/moai/development/rule-authoring.md` (항상 로드 비용), `.claude/rules/moai/development/verification-completeness.md` (§I RED-now 셀 형식), `.claude/rules/moai/development/spec-frontmatter-schema.md` (미러 편집 시 참조)
- neutrality 가드: `.github/workflows/template-neutrality-check.yaml`, `internal/template/internal_content_leak_test.go`
- 인간 루프 상류: 카드 진행 기록(`.moai/reports/<card-id>/`, `.moai/specs/<SPEC-ID>/progress.md`), 판정 보고서(plan-auditor/sync-auditor 산출물) — 기존 흐름의 증거원이지 본 SPEC이 만드는 것이 아니다

## §I. Acceptance Criteria (Tier S inline — Given-When-Then)

> Tier S: AC는 본 섹션에 인라인한다 (acceptance.md 없음). 검증 형식은 `verification-completeness.md` §2를 따른다 — 하중 AC(판정 대상: AC-LCS-001/002/004)는 release-blocking으로 RED-now 셀 4요소(명령/원문 출력/exit/tree SHA)를 가지며, 전 AC에 flipping 마일스톤을 태그한다. RED 관측은 본 라운드(v0.1.1, 2026-09-01, worktree `d7ce6c6bd`)에서 실제 실행해 수집했다. claim 마커는 표면 언어와 일치시킨다 — 영어 표면(moai-constitution-detail.md, SKILL.md)엔 `human-mediated loop`, 한국어 표면(CLAUDE.local.md)엔 `인간 매개 루프` (N1 재키잉, v0.1.2 — 재관측: 영어 마커 양 표면 오늘 `0`/exit 1).

**RED-now 셀 (본 라운드 실관측 — worktree `d7ce6c6bd`):**

| AC | 명령 (단일 호출) | 원문 출력 | exit | flipping |
|---|---|---|---|---|
| AC-LCS-001 | `test -f .moai/docs/learning-channel-scope.md` | (출력 없음) | 1 | M1 |
| AC-LCS-002 | `grep -c 'human-mediated loop' .claude/rules/moai/core/moai-constitution-detail.md` | `0` | 1 | M2 |
| AC-LCS-003 | `grep -c 'human-mediated loop' .claude/skills/hns-lsel-curator/SKILL.md` | `0` | 1 | M2 |
| AC-LCS-004 | `grep -c 'learning-channel-scope' .claude/skills/hns-lsel-curator/SKILL.md` | `0` | 1 | M3 |
| AC-LCS-007 | `grep -c '인간 매개 루프' CLAUDE.local.md` | `0` | 1 | M2 |
| (baseline) | `grep -c 'test_fail:' <primary>/.moai/lessons-inbox.jsonl` | `0` | 1 | — (dated baseline — 본 시점 경험값) |

> AC-LCS-005/006은 regression-guard다 — 판정 대상은 run-phase 완료 시점의 diff와 트리 상태이고, 사전 나무에서 "적색"이 성립하는 내용 주장이 없어 RED-now 셀을 갖지 않는다 (release-blocking 지정 없음).

- **AC-LCS-001 (재검증 가능성, release-blocking, M1)** — **Given** the anchor document `.moai/docs/learning-channel-scope.md`, **When** its recorded tally command is re-run against the primary checkout's live inbox, **Then** the fresh tally again shows zero rows outside the `{tool_failure, test_fail}` families, and the document's composition claim carries its measurement date and tree SHA as a dated baseline (the AC judges the claim's re-verifiability, not the count's immutability).
- **AC-LCS-002 (채널 인정 + 미러 패리티, release-blocking, M2)** — **Given** `.claude/rules/moai/core/moai-constitution-detail.md` § Lessons Protocol after run-phase, **When** both the local file and its template mirror (`internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md`) are searched for the claim marker `human-mediated loop` (English token — this surface is an English-language document, N1), **Then** both carry the bounded capture-scope statement naming the human loop as the channel for families invisible to both wired families, and the mirror omits dev-local measurement values (dates, counts, SPEC IDs, card ids, machine-local paths).
- **AC-LCS-003 (SKILL 표면, M2)** — **Given** `.claude/skills/hns-lsel-curator/SKILL.md`, **When** searched for the channel-scope note, **Then** a bounded scope statement exists in constraint-7 form — the bounded claim in brief plus the anchor pointer, numbers living only in the anchor doc (pointer-only) — naming the two wired families and carrying the claim marker `human-mediated loop` (English token — this skill body is English, N1), and the stale live count at line 34 ("624 stubs") is replaced by the anchor pointer (no live count in prose).
- **AC-LCS-004 (claim 표면 일관성, release-blocking, M3)** — **Given** the claim-surface list enumerated in plan.md §A.5 (the surfaces describing the inbox's capture scope, per REQ-LCS-004), **When** each listed surface is checked, **Then** none presents the inbox as capturing defect families invisible to both wired families — each states the bounded claim or points to the anchor document (surfaces outside the §A.5 list are out of this AC's scope).
- **AC-LCS-005 (문서 전용 diff, regression-guard, M3)** — **Given** the run-phase diff, **When** `git diff --stat` is examined, **Then** it touches only markdown surfaces plus the template mirror — `internal/hook/failure_observer.go` and `.claude/skills/hns-lsel-curator/{drain.sh,session_drain.sh,backlog_check.sh}` are byte-unchanged, and zero Go source lines changed beyond the constraint-1 kickoff-conditional comment exception (which, if adopted, shows exactly one comment line in `failure_observer.go`).
- **AC-LCS-006 (기전 영, regression-guard, M3)** — **Given** the tree after run-phase, **When** searched for a new record format, hook wiring, config section, or extraction script, **Then** none exists — zero new files under `.moai/config/sections/`, zero new hook entries, zero new scripts, zero stub families beyond the two wired ones; the recording target remains `feedback_*.md` + `MEMORY.md`.
- **AC-LCS-007 (로컬 운영 표면, M2)** — **Given** `CLAUDE.local.md` LSEL operating section, **When** read, **Then** it carries the bounded inbox-scope clause with the anchor pointer.
