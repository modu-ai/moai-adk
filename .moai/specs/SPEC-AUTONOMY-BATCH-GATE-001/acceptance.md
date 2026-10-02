---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "Acceptance criteria — 배치 게이트 요약"
version: "0.2.0"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
module: ".claude/rules/moai/workflow, .claude/skills/moai/workflows, internal/template/templates, internal/hook"
tier: M
---

# SPEC-AUTONOMY-BATCH-GATE-001 — Acceptance

## §D AC Matrix

검증 계층이다. 요구사항 계층(GEARS)은 `spec.md` §C가 소유하고, 여기서는 반복하지 않는다. 각 기준은 `Given … When … Then …` 형식이고 `AC-NNN`으로 라벨링했다. 분류는 두 가지다.

- **release-blocking**: RED-now 칸(명령 · 원문 stdout · 종료 코드 · 트리 SHA)을 §Evidence Ledger에 갖는다. 문서 수준 고정: **트리 `c50da9c2f`**(브랜치 `WT-batch-approval-gate`, 측정 시점 HEAD). 이 고정은 칸 자체에 고정이 없는 모든 기준에 적용된다.
- **regression-guard**: 현재 GREEN인 가드가 계속 GREEN이어야 하거나, 감사가 읽어 판정하는 시나리오. RED-now 칸 없음. 시나리오 기준은 기계 점검이 불가능하며 그 사실을 §D.1에 공시했다.

검증 명령은 단순 명령·리터럴 경로·`-run` 앵커 패턴만 쓴다(워크트리 가드 안전). 테스트 계열 기준의 통과는 `--- PASS` 줄과 *비어 있지 않은 하위 테스트 집합*을 함께 읽어야 한다(빈 selector의 `ok`는 통과가 아니다 — `verification-completeness.md` §1.1).

### AC-001 — 정본 절이 라이브와 미러에 존재한다 (release-blocking, RED: RED-1)

- **Given** 변경 전 트리에서 `auto-semantics.md`는 §9.1에서 끝나는 §9이다.
- **When** run-phase가 정본 절을 작성한다.
- **Then** 라이브와 템플릿 미러 양쪽에 제목 `### 9.2 The batch gate summary`가 정확히 한 번 있다.
- Verify: `grep -c "^### 9.2 The batch gate summary" .claude/rules/moai/workflow/auto-semantics.md` → `1`
- Verify: `grep -c "^### 9.2 The batch gate summary" internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` → `1`
- 얕은 기준이다: 제목만 있는 빈 절도 통과한다. 그래서 AC-002~007이 절 *안의* 내용을 점검한다.

### AC-002 — 요약 서식, 질문 하나, 승인 범위, pull-out (release-blocking, RED: RED-2)

- **Given** 같은 게이트 행에서 운영자 결정을 기다리는 카드가 둘 이상이다.
- **When** 정본 절의 서식 규칙(REQ-BGS-004·005·006)을 점검한다.
- **Then** 정본 절은 (a) 행 필드 전부(카드 id, 게이트 행, SPEC id, 감사 판정+점수+여유, 해시 불변, 기록 참조, `counter_refs=`)와 반대 증거 행 우선 정렬, (b) 질문이 정확히 하나이고 승인이 나열된 승인 가능 행에만 미친다는 문장, (c) 행별 pull-out 제공을 명시한다.
- Verify: `go test -count=1 -v -run '^TestBatchGateSummaryDoctrine$' ./internal/template/` — `--- PASS: TestBatchGateSummaryDoctrine/real_section ` (이름 뒤 공백 포함)과 불변식 하위 테스트가 모두 `--- PASS`.

### AC-003 — 반대 증거 필드, `none searched=`, 행별 결정 기록 (release-blocking, RED: RED-2)

- **Given** 어떤 행에도 반대 증거가 없거나 있다.
- **When** 정본 절의 REQ-BGS-007·008·009 규칙을 점검한다.
- **Then** 정본 절은 `counter_refs=` 필드와 원천 목록(감사 경고·부채, 여유, 미해결 decision-index 행, 감사 교차 불일치, 열린 차단·대기 기록, 행 간 경로 겹침), `counter_refs=none`에는 `searched=`가 따라야 한다는 문장, 승인된 행마다 자기 §10 결정 기록이 있어야 하고 기록 없는 행은 미승인이라는 문장을 담는다.
- Verify: AC-002와 같은 명령(하위 테스트 `counter_refs_*`, `record_per_row`).

### AC-004 — 예약 행과 contract 모드 제외 (release-blocking, RED: RED-2)

- **Given** 행이 keep-set 범주이거나 권한 게이트가 운영자·리더에게 남긴 결정이거나, 모드가 `contract`이다.
- **When** 정본 절의 REQ-BGS-010·012 규칙을 점검한다.
- **Then** 정본 절은 keep-set 세 범주(환경상 불가능, 운영자 보유, 외부 공유 시스템의 되돌릴 수 없는 조작)와 권한 예약 목록을 단일 승인에서 제외한다고 명시하고, contract 모드에서는 서명이 plan→run 게이트이며 요약 행이 생기지 않는다고 명시한다.
- Verify: AC-002와 같은 명령(하위 테스트 `reserved_rows`, `contract_mode`).

### AC-005 — 차단 행은 승인에서 빠진다 (release-blocking, RED: RED-2)

- **Given** 행의 감사 판정이 FAIL·INCONCLUSIVE·부재이거나, 판정 후 계획 산출물 해시가 바뀌었거나, 열린 차단이 있다.
- **When** 정본 절의 REQ-BGS-011 규칙을 점검한다.
- **Then** 정본 절은 그런 행을 차단 행으로 표시하고 승인에서 제외하며 §7 권한 게이트 불변식을 약화하지 않는다고 명시한다. 다섯 조건이 각각 나열돼 있다.
- Verify: AC-002와 같은 명령(하위 테스트 `blocked_rows` — 조건 다섯 개 앵커 각각).

### AC-006 — 구성원 규칙: 같은 게이트 행, 붙잡지 않기, 동일 판정 1회 (release-blocking, RED: RED-2)

- **Given** 카드가 서로 다른 게이트 행에서 기다리거나, 일부만 준비됐거나, 한 판정이 여러 카드를 막는다.
- **When** 정본 절의 REQ-BGS-001·002·003 규칙을 점검한다.
- **Then** 정본 절은 같은 게이트 행끼리만 묶고(레인은 교차 카드 배치를 만들지 않음), 준비된 카드를 기다리게 하지 않으며, 동일 판정 주체는 영향 카드를 전부 적어 한 번 묻는다고 명시한다.
- Verify: AC-002와 같은 명령(하위 테스트 `membership_rules`).

### AC-007 — 가드 테스트는 반증 가능하다: 변이 본문은 거부된다 (release-blocking, RED: RED-2)

- **Given** 정본 절의 불변식 앵커 표(AC-002~006의 각 불변식 1개)와 실제 절.
- **When** 앵커를 하나씩 뺀 변이 본문 각각을 같은 검사 함수에 넣는다.
- **Then** 실제 절은 위반 0이고, 변이마다 *해당* 불변식 하나가 위반으로 보고된다. 하위 테스트 `--- PASS`가 10개 이상 나열된다(실제 절 1 + 변이 여러 개).
- Verify: `go test -count=1 -v -run '^TestBatchGateSummaryDoctrine$' ./internal/template/` — 출력에 `--- PASS` 하위 테스트 줄이 10개 이상.
- 변이 탐침: 앵커 문구를 *모순 문장과 함께* 남긴 변이는 이 점검을 통과할 수 있다(어휘적 한계, G-4). 그 방향은 AC-011~013이 감사 읽기로 막는다.

### AC-008 — 기준선이 변경보다 앞선 별도 커밋으로 남는다 (release-blocking, RED: RED-3)

- **Given** run-phase가 시작되지 않았다.
- **When** M0가 기준선 산출물을 기록하고 이후 정본 절 커밋이 만들어진다.
- **Then** `.moai/reports/t1344/baseline-gate-rounds.md`가 존재하고(명령 원문, 관측 출력, 분류 방식, 한계 포함), 그 파일을 처음 추가한 커밋이 정본 절 커밋의 조상이다. 두 커밋은 서로 다르다.
- Verify: `ls .moai/reports/t1344/baseline-gate-rounds.md` → 경로 출력, exit 0
- Verify: `git log --format=%h --diff-filter=A -- .moai/reports/t1344/baseline-gate-rounds.md` → 커밋 하나(기준선 커밋). 이어서 `git merge-base --is-ancestor <그 SHA> <정본 절 커밋 SHA>` → exit 0. 이 조상 판정이 순서의 증인이다(VCI §2.3 — 커밋 메시지 주장이 아니라 그래프).

### AC-009 — Go 변경이 허용 목록 안에 있다 (release-blocking, RED: RED-4)

- **Given** 브랜치가 `develop`에서 분기한 상태이다.
- **When** run-phase가 끝난다.
- **Then** `develop`과의 merge-base 이후 바뀐 `.go` 파일은 정확히 일곱 개이고 전부 허용 목록 안에 있다. 제품 Go 넷: `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`, `internal/hook/session_start_kanban.go`, `internal/hook/session_start_kanban_i18n.go`. 테스트 Go 셋: `internal/hook/session_start_kanban_i18n_test.go`, `internal/hook/session_start_leader_gate_notice_test.go`, `internal/template/batch_gate_summary_doctrine_test.go`. 이 목록 밖의 `.go` 파일이 하나라도 있으면 실패다.
- Verify: `git diff --name-only develop...HEAD -- '*.go'` → 정확히 위 일곱 줄(경로 바이트 순서: 훅 다섯 줄 `session_start_factory.go`, `session_start_factory_i18n.go`, `session_start_kanban.go`, `session_start_kanban_i18n.go`, `session_start_kanban_i18n_test.go`, 이어서 `session_start_leader_gate_notice_test.go`, 마지막에 `internal/template/batch_gate_summary_doctrine_test.go`)
- 읽기 단계: 제품 Go 네 파일의 변경이 리더 공지 문장 필드와 그것을 기존 블록에 합치는 한 줄에 한정되는지는 `git diff --numstat develop...HEAD -- internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go` 출력을 sync 감사가 읽어 판정한다(기계 점검 없음, `spec.md` G-8).
- 평가는 병합 전에만 유효하다. 카드가 `develop`에 병합되면 merge-base가 카드 tip이 되어 범위가 비고 공허하게 통과한다(`gitflow-lane-protocol.md` §8의 한계). 병합 뒤 근거는 병합 트리와 카드 브랜치 트리의 동일성이다.

### AC-010 — 고정 가드와 예산이 GREEN을 유지한다 (regression-guard)

- **Given** 변경 전 트리에서 아래 가드가 전부 GREEN이다(§Evidence Ledger GREEN 기준선).
- **When** run-phase가 정본 절과 포인터를 라이브·미러에 반영한다.
- **Then** 가드가 전부 계속 `--- PASS`이고 `TestAlwaysLoadedTokenBudget`의 headroom이 0 이상이며, `kanban-dispatch.md` 순증가가 1,000바이트 미만이고 `run.md`는 199줄을 넘지 않는다. 리더 공지 쪽 기존 고정 테스트 63개도 그대로 `--- PASS`이고 `--- FAIL`은 없다.
- Verify: `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/`
- Verify: `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/`
- Verify: `go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/`
- Verify: `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/`
- Verify: `go test -count=1 -v -run '^(TestImplementationKickoffApprovalPreservedBeforeGoal|TestTemplateNoInternalContentLeak|TestContractModeEmitterSites)$' ./internal/template/`
- Verify: `wc -l .claude/skills/moai/workflows/run.md` → `199 …` 이하
- Verify: 리더 공지 고정 테스트 — `plan.md` "추가 측정" E17의 23개 이름 명령 한 줄 그대로. 읽는 법: `--- PASS` 63줄 이상, `--- FAIL` 0줄, `=== RUN`과 같은 수(비어 있지 않은 하위 테스트 집합). 새 테스트는 AC-015가 따로 다룬다.

### AC-011 — 시나리오: 혼합 배치 (regression-guard, 감사 읽기)

- **Given** 같은 Kickoff 게이트 행에서 카드 A(감사 PASS, 반대 증거 없음), 카드 B(PASS이고 기록된 부채 하나), 카드 C(감사 FAIL)가 대기한다.
- **When** 리더가 배치 게이트 요약을 만든다.
- **Then** 보고서에서 B 행이 A 행보다 먼저 나오고 B의 `counter_refs=`가 부채를 가리키며, A는 `counter_refs=none searched=<검색 내용>`이고, C는 차단 행으로 표시되어 승인에서 빠지고, 질문은 하나이며 승인은 A·B에만 미친다.
- Verify: sync 감사가 정본 절을 읽고 위 네 가지가 절의 규칙으로 도출됨을 확인한다(기계 점검 없음, §D.1).

### AC-012 — 시나리오: pull-out (regression-guard, 감사 읽기)

- **Given** A·B·C 세 행에 대한 단일 승인 질문이 나갔다.
- **When** 운영자가 B를 빼낸다.
- **Then** A·C는 각자 자기 §10 결정 기록과 `counter_refs=`를 남기며 승인되고, B는 개별 질문으로 남고, 기록 없는 행은 승인으로 치지 않는다.
- Verify: sync 감사가 REQ-BGS-006·009에서 도출되는지 읽는다.

### AC-013 — 시나리오: 붙잡지 않기와 동일 판정 1회 (regression-guard, 감사 읽기)

- **Given** 카드 D는 준비됐고 카드 E는 아직 감사 중이다. 별도로 하나의 게이트 설정이 카드 F·G·H를 막고 있다.
- **When** 리더가 운영자 질문을 구성한다.
- **Then** D는 E를 기다리지 않고 제시되며 E는 다음 요약이나 개별 질문으로 간다. F·G·H는 같은 판정 질문 하나에 세 카드를 모두 적어 한 번 묻는다.
- Verify: sync 감사가 REQ-BGS-002·003에서 도출되는지 읽는다.

### AC-014 — 낡은 표현 정합 (release-blocking, 무조건 — Q4 범위 안, RED: RED-5)

- **Given** `spec-assembly.md:202-208`에 "stays MANDATORY and score-independent"와 "does NOT substitute for the gate"가, `moai.md:144,240`에 카드별 필수 운영자 Kickoff를 말하는 "Score-independent … never bypasses it"와 "score-independent"가 있고, 두 문서 어디도 `auto-semantics.md` §9.1·§9.2를 인용하지 않는다(RED-5).
- **When** run-phase M4가 그 줄들을 §9.1·§9.2에 맞춰 다시 쓴다.
- **Then** 라이브와 미러 두 사본 모두에서 위 낡은 문구가 없고, `§9.1`과 `§9.2`가 각각 한 번 이상 인용되며, 보존 문구(`[HARD] The Implementation Kickoff Approval`, `moai plan render-html`, `Fail-open`)와 줄 수(`spec-assembly.md` 597, `moai.md` 라이브 284 / 미러 282)와 라이브↔미러 사전 차이 헌크 넷이 그대로다.
- Verify (아래 grep 다섯 줄은 각각 라이브와 미러 경로에 같은 명령으로 두 번씩 돌린다):
  - `grep -c "stays MANDATORY" .claude/skills/moai/workflows/plan/spec-assembly.md` → `0`
  - `grep -c "does NOT substitute for the gate" .claude/skills/moai/workflows/plan/spec-assembly.md` → `0`
  - `grep -c "never bypasses it" .claude/skills/moai/workflows/moai.md` → `0`
  - `grep -ci "score-independent" .claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/plan/spec-assembly.md` → 두 파일 모두 `0`
  - `grep -c "§9.2" .claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/plan/spec-assembly.md` → 두 파일 모두 1 이상(같은 방식으로 `§9.1`도 1 이상)
- Verify: `wc -l .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md` → `597`, `284`, `282`
- Verify: `diff .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md` → exit 1이고 헌크 헤더가 정확히 `215c215`, `245c245`, `253d252`, `283,284c282` 넷
- Verify: `go test -count=1 -v -run '^(TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/` → 두 `--- PASS`
- 변이 탐침: 낡은 구절 하나만 지우고 같은 주장을 말만 바꿔 유지하는 변이(예: "always mandatory")는 위 grep 일부를 통과할 수 있다(어휘적 한계). `§9.1`·`§9.2` 인용 요구가 포인터 누락 변이를 막고, 사본 한쪽만 고친 변이는 `diff` 헌크 헤더와 두 경로 grep이 막는다. 말만 바꾼 변이는 sync 감사가 다시 쓴 문구를 §9.1에 대조해 읽어 막는다(기계 점검 없음).

### AC-015 — 리더 공지 문장이 두 리더 공지의 네 로케일에 있다 (release-blocking, RED: RED-6)

- **Given** 팩토리 리더 공지와 칸반 리더 공지는 지금 배치 게이트 요약을 언급하지 않는다(`plan.md` E8, RED-6).
- **When** run-phase M5가 두 리더 공지에 문장 하나를 en·ko·ja·zh로 더한다.
- **Then** (a) 두 리더 공지 × 네 로케일 여덟 조합 모두에서 렌더된 공지가 포인터 토큰 `.claude/rules/moai/workflow/auto-semantics.md` §9.2를 담는다. (b) 그 문장은 질문 도구 이름, `SPEC-`, 카드 id 꼴(`t` 뒤 숫자 3–5자리), ISO 날짜, `moai todo`, 단어 `lead`, `리드`, `epic`을 담지 않는다. (c) 레인·동반 공지에는 포인터가 없다. (d) 알림 소스 네 파일의 주석 제외 `AskUserQuestion` 토큰 수가 0이다. (e) 기존 고정 테스트 63개가 그대로 `--- PASS`다(AC-010). (f) 변이 본문 — 한 로케일 누락, 질문 도구 이름 포함, 포인터 누락, 카드 id 포함, 한쪽 리더 공지만 보유 — 은 각각 점검 함수에서 거부된다. (g) SessionStart 핸들러 수준에서 칸반 리더와 팩토리 리더 각각, 에이전트용 영어 사본(`additionalContext`)과 운영자용 현지어 사본(`systemMessage`, ko 한 경로 이상) 모두에 문장이 실린다.
- Verify: `go test -count=1 -v -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/` — 새 테스트의 `--- PASS` 줄이 12개 이상(여덟 조합 + 레인·동반 부재 + 소스 스캔 + 변이 다섯 이상 + 핸들러 수준)이고 `--- FAIL`이 0줄. 빈 selector의 `ok … [no tests to run]`은 통과가 아니다(RED-6에 그 출력이 있다).
- Verify: `grep -c AskUserQuestion internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go` → 네 파일 모두 `:0`(exit 1, 줄 순서는 인자 순서와 다를 수 있다)
- Verify: `` grep -cF 'auto-semantics.md` §9.2' internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory_i18n.go `` → 두 파일 모두 `:4`(로케일 값 하나씩. 주석에 같은 토큰을 되풀이하지 않는다)
- 변이 탐침(`verification-completeness.md` §2): 한 로케일의 문장을 지운 공지는 (a)와 grep `:4`에서 실패하고, 질문 도구 이름을 넣은 문장은 (b)·(d)에서 실패하고, 포인터를 뺀 문장은 (a)와 grep에서 실패하고, 한쪽 리더 공지에만 문장을 둔 구현은 (a)에서 실패한다. 통과할 수 있는 변이: 포인터를 둔 채 뜻이 정본과 반대인 문장(어휘적 한계, `spec.md` G-7) — sync 감사가 네 로케일 문장을 읽어 막는다.

## §Evidence Ledger (RED-now 관측, 트리 `c50da9c2f`, 브랜치 `WT-batch-approval-gate`)

각 항목은 명령(단일 호출) · 원문 stdout · 종료 코드 · 빨간 이유를 가진다. 종료 코드는 별도 필드로 적었다(단일 호출 형식에서는 `; echo $?`를 못 쓰므로). 측정은 플랜 단계에서 이 에이전트가 직접 실행했다.

### RED-1 — AC-001: 정본 절이 없다

- 명령 A: `grep -c "^### 9.2 The batch gate summary" .claude/rules/moai/workflow/auto-semantics.md`
  - stdout: `0` · exit: `1`
- 명령 B: `grep -c "^### 9.2 The batch gate summary" internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`
  - stdout: `0` · exit: `1`
- 빨간 이유: `auto-semantics.md`는 §9.1(`:173`)에서 끝나는 §9 뒤에 §10(`:183`)이 곧바로 이어진다. §9.2는 없다.
- 초록 경로: M2가 두 파일에 제목을 쓰면 둘 다 `1`.

### RED-2 — AC-002~007: 가드 테스트가 없다

- 명령: `ls internal/template/batch_gate_summary_doctrine_test.go`
  - stdout: `ls: internal/template/batch_gate_summary_doctrine_test.go: No such file or directory` · exit: `1`
- 빨간 이유: 테스트 파일이 없다. 존재하지 않는 테스트를 `-run`으로 고르면 "테스트 0개 실행, ok"가 되어 *공허하게 통과*하므로(`verification-completeness.md` §2.1 근거 사례) RED 칸은 `go test`가 아니라 파일 존재 여부로 잡았다.
- 초록 경로: M1이 파일을 만들고(정본 절이 없으므로 처음엔 `FAIL`), M2가 절을 써서 `--- PASS`로 뒤집는다. 통과 출력의 하위 테스트 집합이 비어 있지 않음을 AC-007이 요구한다.

### RED-3 — AC-008: 기준선 산출물이 없다

- 명령: `ls .moai/reports/t1344/baseline-gate-rounds.md`
  - stdout: `ls: .moai/reports/t1344/baseline-gate-rounds.md: No such file or directory` · exit: `1`
- 빨간 이유: 기준선이 아직 측정되지 않았다(G-1). `.moai/reports/<card-id>/` 아래 산출물은 추적 대상이다(`git ls-files .moai/reports`에 `…/verdict.md` 항목 확인).
- 초록 경로: M0가 파일을 만들고 단독 커밋한다.

### RED-4 — AC-009: 허용 목록의 .go 변경 일곱 줄이 아직 없다

- 명령: `git diff --name-only develop...HEAD -- '*.go'`
  - stdout: (비어 있음) · exit: `0` (0.2.0에서 트리 `c50da9c2f`에 다시 쟀다. 이 트리에서 `git merge-base develop HEAD`는 `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`, 즉 HEAD 자신이다)
- 빨간 이유: 기대값은 *정확히 일곱 줄*이다. 빈 출력은 일곱 파일이 아직 없다는 뜻이다. 빈 출력은 "Go가 안 바뀐" 상태와도 같으므로, 이 기준은 특정 일곱 줄이 생겼음을 요구해서 변경이 없는 빈 실행을 통과시키지 않는다.
- 초록 경로: M1이 템플릿 가드 테스트 한 줄을, M5가 제품 Go 넷과 훅 테스트 둘을 더해 일곱 줄이 된다. 허용 목록 밖의 `.go` 파일은 줄을 늘려 실패한다. 허용 파일 *안*의 범위 초과 변경은 이 명령이 못 잡는다 — AC-009 읽기 단계(`spec.md` G-8).

### RED-5 — AC-014: 낡은 표현이 아직 있고 §9.1·§9.2 인용이 없다

모두 트리 `c50da9c2f`에서 이 실행이 직접 쟀다(라이브 사본. 미러 사본은 `spec-assembly.md`가 동일, `moai.md`의 편집 대상 줄 `:144`·`:240`이 동일하다 — `plan.md` E15).

- 명령 5a: `grep -c "stays MANDATORY" .claude/skills/moai/workflows/plan/spec-assembly.md`
  - stdout: `1` · exit: `0`
- 명령 5b: `grep -c "does NOT substitute for the gate" .claude/skills/moai/workflows/plan/spec-assembly.md`
  - stdout: `1` · exit: `0`
- 명령 5c: `grep -c "never bypasses it" .claude/skills/moai/workflows/moai.md`
  - stdout: `1` · exit: `0`
- 명령 5d: `grep -ci "score-independent" .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md .claude/skills/moai/workflows/plan/spec-assembly.md`
  - stdout(줄 순서는 인자 순서와 다르게 찍혔다): `internal/template/templates/.claude/skills/moai/workflows/moai.md:2` · `.claude/skills/moai/workflows/moai.md:2` · `.claude/skills/moai/workflows/plan/spec-assembly.md:1` · exit: `0`
- 명령 5e: `grep -c "§9.2" .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/moai.md`
  - stdout: `.claude/skills/moai/workflows/plan/spec-assembly.md:0` · `.claude/skills/moai/workflows/moai.md:0` · exit: `1`
- 명령 5f: `grep -c "§9.1" .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/moai.md`
  - stdout: `.claude/skills/moai/workflows/plan/spec-assembly.md:0` · `.claude/skills/moai/workflows/moai.md:0` · exit: `1`
- 빨간 이유: 5a–5d는 기대값 `0`인데 낡은 문구가 각각 있고, 5e·5f는 기대값 1 이상인데 두 문서가 정본 절을 한 번도 인용하지 않는다.
- 초록 경로: M4가 정합하면 5a–5d는 `0`(grep -c는 0일 때 exit 1이므로 통과 판독은 stdout), 5e·5f는 두 문서 모두 1 이상.

### RED-6 — AC-015: 리더 공지 문장도 그 테스트도 없다

- 명령 6a: `ls internal/hook/session_start_leader_gate_notice_test.go`
  - stdout: `ls: internal/hook/session_start_leader_gate_notice_test.go: No such file or directory` · exit: `1`
- 명령 6b: `` grep -cF 'auto-semantics.md` §9.2' internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory_i18n.go ``
  - stdout: `internal/hook/session_start_kanban_i18n.go:0` · `internal/hook/session_start_factory_i18n.go:0` · exit: `1`
- 명령 6c: `go test -count=1 -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/`
  - stdout: `ok  	github.com/modu-ai/moai-adk/internal/hook	0.390s [no tests to run]` · exit: `0`
- 빨간 이유: 6a·6b가 RED 신호다 — 테스트 파일이 없고 두 로케일 표 어디에도 포인터 토큰이 없다. 6c는 *왜 RED 칸을 `go test`로 잡지 않았는가*의 증거다: 존재하지 않는 테스트를 `-run`으로 고르면 "테스트 0개 실행, ok, exit 0"이 되어 공허하게 통과한다(`verification-completeness.md` §2.1). 그래서 AC-015의 통과 판독은 `--- PASS` 줄의 수가 12 이상일 때만이다.
- 초록 경로: M5가 테스트 파일을 먼저 만들고(포인터 부재로 `FAIL`을 관측), 필드와 네 로케일 값을 더해 `--- PASS`로 뒤집는다. 6b는 두 파일 각 `:4`가 된다.

### GREEN 기준선 (regression-guard가 공허하지 않음을 보이는 변경 전 관측, 트리 `c50da9c2f`)

| 가드 | 명령 | 관측 |
|---|---|---|
| 상시 로딩 예산 | `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/` | `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)` · `--- PASS` · exit 0 |
| 미러 동일성 | `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/` | `--- PASS` · 하위 테스트 9개 · exit 0 |
| LOC 한도 | `go test -count=1 -v -run '^(TestSubSkillLOCCeiling\|TestEntryRouterLOCCeiling)$' ./internal/skills/` | 두 테스트 `--- PASS` · exit 0 |
| 보존 구간·스펙 조립 | `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment\|TestAutoRankMirrorParity\|TestSpecAssembly_RewrittenToCLIPath)$' ./internal/cli/` | 세 테스트 `--- PASS` · exit 0 |
| 리더 공지 고정 테스트(0.2.0) | `plan.md` "추가 측정" E17의 23개 이름 명령(`./internal/hook/`) | `--- PASS` 63줄 · `--- FAIL` 0줄 · `=== RUN` 63줄 · `ok  github.com/modu-ai/moai-adk/internal/hook  6.364s` · exit 0 |
| 알림 소스 질문 도구 토큰(0.2.0) | `grep -c AskUserQuestion internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go` | 네 파일 모두 `:0` · exit 1(적중 없음) — 줄 순서는 인자 순서와 다르게 찍혔다 |

AC-010의 나머지 테스트(`TestImplementationKickoffApprovalPreservedBeforeGoal`, `TestTemplateNoInternalContentLeak`, `TestContractModeEmitterSites`, `TestSpecAssembly_NoNewInternalTokens`)의 변경 전 기준선은 이번 플랜에서 돌리지 않았다 — Gap. run-phase Pre-flight(§C)에서 먼저 측정한다.

## §D.1 Edge cases

- **행이 하나뿐**: 둘 이상일 때만 요약이다. 하나면 개별 질문이다(정본 절 적용 대상 문장).
- **모두 차단 행**: 승인 질문이 없다. 차단 사유를 보고하고 끝낸다.
- **질문 채널이 없는 하니스**: 같은 요약이 blocker 보고서로 나간다(`spec.md` G-6).
- **요약 직후 새 행 도착**: 이미 나간 승인은 그 행을 덮지 않는다(REQ-BGS-005).
- **PASS 기준 직상 점수**: 여유가 작은 행은 반대 증거로 올라가 앞쪽에 놓인다(REQ-BGS-004·007).
- **시나리오 기준(AC-011~013)은 기계 점검이 불가능하다**: 감사가 정본 절을 읽고 도출 가능성을 판정한다. 이 한계는 §G-4에서 공시했다.
- **리더 공지가 나가지 않는 원천**: resume·clear·compact·fork에서는 리더 공지가 방출되지 않는다(기존 `…ForSource` 게이트). 문장이 새 방출 경로를 만들지 않는다(AC-015는 이 동작을 바꾸지 않는다).
- **todo 비활성**: 칸반 리더 공지에 `moai todo`가 없어야 한다. 문장은 이를 건드리지 않는다(AC-015 (b)).
- **알 수 없는 로케일**: `kanbanMessagesFor`·`factoryMessagesFor`가 영어로 폴백하므로 문장도 영어로 나간다(포인터 토큰은 같다).
- **레인·동반 세션**: 문장이 없다(AC-015 (c)). 레인은 카드 1장만 쥔다(REQ-BGS-001).

## §D.2 Quality gates

- 새·변경 Go 파일(테스트 셋 + 제품 넷): `golangci-lint`는 변경 범위만, CI 버전 기준으로(`gofmt -l` 출력 없음, `go vet` 포함). 전체 스위트는 CI에 맡긴다. `internal/hook` 패키지 전체는 분 단위 스위트이므로 레인에서는 이름 선택만 돌린다.
- `moai spec lint SPEC-AUTONOMY-BATCH-GATE-001` — 트리에서 빌드한 바이너리로, 판정 빌드 커밋을 함께 기록(VCI §2.2).
- 미러 동일성, 중립성, 예산, LOC, 보존 구간 가드 전부 GREEN(AC-010).
- 정본 절에 카드 id·SPEC ID·날짜·"176" 결과 수치가 없다.

## §D.3 Definition of Done

- AC-001~010·AC-014·AC-015 PASS(명령 + 원문 출력 + 트리 SHA), AC-011~013 감사 읽기 통과. AC-014와 AC-015는 Q4·Q5 범위 판정(2026-10-02)으로 무조건이다.
- 기준선 커밋이 정본 절 커밋의 조상(AC-008).
- `progress.md` §E.2에 RED 원문, 변이 거부 출력, 예산·미러·LOC 출력, 사전 차이 줄(`kanban-dispatch.md:177`, `moai.md`)을 건드리지 않았다는 확인이 있다.
- 결과 수치(176→10~20)를 어디에도 단언하지 않았다.

## §D.4 Traceability (REQ → AC)

| REQ | AC |
|---|---|
| REQ-BGS-001, 002, 003 | AC-006, AC-013 |
| REQ-BGS-004, 005, 006 | AC-002, AC-011, AC-012 |
| REQ-BGS-007, 008, 009 | AC-003, AC-011, AC-012 |
| REQ-BGS-010, 012 | AC-004 |
| REQ-BGS-011 | AC-005, AC-011 |
| REQ-BGS-013 | AC-001, AC-009, AC-010 |
| REQ-BGS-014 | AC-014 |
| REQ-BGS-015 | AC-008 |
| REQ-BGS-016 | AC-015, AC-009, AC-010 |
| (전체 반증 가능성) | AC-007 |
