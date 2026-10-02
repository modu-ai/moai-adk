---
id: SPEC-AUTONOMY-BATCH-GATE-001
title: "Research — 근거 요약과 전제 검증"
version: "0.3.0"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
tier: L
---

# SPEC-AUTONOMY-BATCH-GATE-001 — Research

> 이 문서는 근거 요약, 전제 검증, 측정 원문을 소유한다(0.3.0에서 `plan.md`의 "Evidence and premise verification" 절을 옮겼다). 요구사항은 `spec.md`, 설계 결정은 `design.md`, 구현 순서는 `plan.md`가 정본이고 여기서는 되풀이하지 않는다. 상태 표기: VERIFIED(이번 실행에서 관측), CORRECTED(전제가 틀려 정정), UNVERIFIED(관측하지 못함).

## R1 범위와 출처 도구

- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1344`, 브랜치 `WT-batch-approval-gate`. 0.1.0·0.2.0 측정은 HEAD `c50da9c2f`, 0.3.0 재측정은 HEAD `72e09d27b0f7f8140bb11223c0d44a36d44984bb`(`c50da9c2f` 위의 SPEC 전용 커밋 하나; 소스 파일은 같다).
- 판정 빌드(VCI §2.2): 첫 빌드 `go build -o <scratch>/moai-t1344 ./cmd/moai`의 `go version -m` VCS 스탬프는 `vcs.revision=c8f245c2c9a58083518f0b7cbebff0542848e033`, `vcs.modified=true`다. `git merge-base --is-ancestor c8f245c2c9a5 HEAD`는 exit 1이라 스탬프는 HEAD의 조상이 아니며 판정 빌드를 증명하지 못한다(감사 1회차와 E16이 기록한 것과 같은 스탬프). 그래서 `go build -ldflags="-X github.com/modu-ai/moai-adk/pkg/version.Commit=72e09d27b" -o <scratch>/moai-72e09d27b ./cmd/moai`로 다시 빌드했고 `moai-72e09d27b version`의 출력은 `v3.1.3   72e09d27b   built unknown`이다. 이 두 번째 바이너리가 판정 도구이며 판정 빌드 커밋 `72e09d27b`은 트리 HEAD와 같다. 설치된 `moai`(`v3.2.0-rc.23`, `d194083fb`)와 MCP 서버(같은 커밋)는 HEAD보다 오래되어 쓰지 않았다.

## R2 근거 요약과 전제 검증

| ID | 주장 | 상태 | 근거 (명령/위치 + 관측) |
|---|---|---|---|
| E1 | 제안서 P4 `:38`(크기 S, "176→10~20"), 게이트 라운드 행 `:12`(약 176, T4 5파일 155), 병목 #1 `:22`, 리스크 `:54` | VERIFIED | `Read /Users/goos/MoAI/moai-adk-go/.moai/reports/autonomy-bottleneck-proposal-20260929.md`로 전문 열람 |
| E1b | 그 보고서는 이 워크트리에 없다(미추적) | VERIFIED | `ls .moai/worktrees/t1344/.moai/reports/autonomy-bottleneck-proposal-20260929.md` → `No such file or directory` |
| E2 | `auto-semantics.md`가 보고서보다 뒤에 도입됐다 | VERIFIED(날짜) / 가정(도입 커밋) | `git log --oneline -1 --format='%h %ad' --date=short 147c25d77` → `147c25d77 2026-09-30`. 이 커밋이 파일을 *도입*했다는 점은 형제 SPEC `§A.4`("NEW doctrine rule")를 따랐고 diff로 확인하지 않았다 |
| E2b | §9 목록(`:152-171`), §9.1(`:173-181`), keep-set 문구(`:154-158`) | VERIFIED | `Read .claude/rules/moai/workflow/auto-semantics.md` 전문(0.3.0에서 `:100-210`을 다시 읽었다) |
| E2c | "t1318 owner conf 0.38"은 근거 불명 | UNVERIFIED | 이 수치의 출처는 제안서 `:54` 자기 자신뿐. primary의 `.moai/reports/t1318/`(파일: `kickoff-20261001.md`, `plan-20261001.md` 등)에서 `grep -rIn "0\.38"` → 적중 없음(양성 대조는 돌리지 않았다 — 약한 부재 증거). 다른 보고서의 `0.38` 적중은 열어 보지 않았다. 3등급 *규칙*만 쓴다 |
| E3 | 3등급 = 리드 3등급 독트린(절대 위임 불가). 로컬 문서이며 배포되지 않음 | VERIFIED | `sed -n 14,34p .moai/docs/jev-local-operations.md`(3등급 정의 `:20-30`). 템플릿 쪽 적중 없음: `grep -rIln "ASK-OPERATOR\|route.sh\|3-grade\|3등급" internal/template/templates .claude/rules/moai .claude/skills` → 출력 없음. `ls scripts/jev/route.sh` → `No such file or directory`(exit 1); `scripts/jev/`에는 `triage.py`, `test_triage.py`만 있다 |
| E3b | 배포되는 대응물 | VERIFIED | `kanban-dispatch.md:106`("The deputy never holds a power of consequence" — 목록은 R4.5에 원문 대조); 템플릿 미러에도 `:106` 존재(`grep -n` 확인) |
| E4 | 상충 표현 위치 | VERIFIED | `spec-assembly.md:202`("stays MANDATORY"), `moai.md:144,240`(`grep -n`), 대조: `run.md:127-139`, `kanban-dispatch.md:187`, `orchestration-mode-selection.md:16,18` |
| E4b | 형제 SPEC 개정 목록이 두 파일을 다루지 않았다 | VERIFIED | `grep -n "spec-assembly" .moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md` → 적중 없음(exit 1); `grep -n "workflows/moai.md" …` → 적중 없음(exit 1). 양성 대조: 같은 파일에서 `grep -n "kanban-dispatch.md"`는 `:108`, `:169` 등에서 적중. 즉 개정 행 A–N과 "검토 후 변경 불필요" 목록 어디에도 두 파일명이 없다 |
| E5 | 용어 충돌 | VERIFIED | `kanban-dispatch.md:31`("`/moai:todo --auto` is the operator's batch approval"), `:102,177`(배치=디스패치/릴리스 배치), `gtd.md:334`, `auto-semantics.md:169` |
| E6 | 재사용할 서식과 결손 | VERIFIED | §10 `:183-199`(결정 기록 3필드, 반대 증거 필드 없음), §7 `:113-132`(권한 게이트 불변식), §11 결정 보드, §14 대기 기록 |
| E7 | `factory decide` 다중 카드, 단일 gate/choice, 기록 없음 | VERIFIED | `sed -n 1465,1495p internal/cli/factory_card.go`(순회 루프 `:1469-1491`); `mcp_factory_card.go:70-79,205-217`(카드 1장 단위, `factoryLaneRefusal()` 거부) |
| E8 | 주입 문장에 승인 언급 없음 | VERIFIED | `grep -n "AskUserQuestion\|Kickoff\|plan-audit\|approval" internal/hook/session_start_factory_i18n.go internal/hook/session_start_kanban_i18n.go internal/hook/lane_spawn_authority.go` → 적중은 `lane_spawn_authority.go:10` 주석("approval") 하나. `grep -rn "AskUserQuestion" internal/hook \| grep -v _test.go`에는 `post_tool.go:292`, `user_decision_capture.go` 주석·상수 적중이 있다(코드 호출 아님). 단 `internal/hook/CLAUDE.md:13`은 이 종류의 grep이 0줄이어야 한다고 쓰므로 그 문구는 이 트리에서 거짓이다 — E13 |
| E9 | LOC 한도 | VERIFIED | `wc -l`: `run.md` 199, `spec-assembly.md` 597, `kanban-dispatch.md` 199, `auto-semantics.md` 256. `go test -count=1 -v -run '^(TestSubSkillLOCCeiling\|TestEntryRouterLOCCeiling)$' ./internal/skills/` → 두 테스트 `--- PASS`, `ok` |
| E9b | 상시 로딩 예산 여유 | **CORRECTED** | 코드 주석은 "168 tokens of 77,600"이라 쓰지만 `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/` → `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)`, `--- PASS`. 주석이 낡았다(이 SPEC의 몫이 아님). 그래도 상시 로딩 파일 증가는 `rule-authoring.md` 1,000바이트 기준을 따른다 |
| E9c | 고정 가드의 현재 상태(GREEN 기준선) | VERIFIED | `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/` → `--- PASS`(하위 테스트 9개); `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment\|TestAutoRankMirrorParity\|TestSpecAssembly_RewrittenToCLIPath)$' ./internal/cli/` → 3개 `--- PASS`. 이들은 플랜 트리(HEAD `c50da9c2f`)의 *변경 전* 측정이다. 감사 1회차가 같은 이름들의 존재를 `go test -list`로 독립 확인했다(재실행은 `TestAlwaysLoadedTokenBudget`와 훅 23개 이름만) |
| E9d | 미러 현황 | VERIFIED | `diff -q`: `auto-semantics.md`·`run.md`·`spec-assembly.md` 동일; `kanban-dispatch.md`는 `:177`에서, `moai.md`는 `:215,245,253,283-284`에서 이미 다름(사전 존재 — 건드리지 않는다). 0.3.0: `wc -c`로 `kanban-dispatch.md` 라이브 26807바이트·미러 26485바이트(R5) |
| E9e | Frozen 구역 | VERIFIED | `grep -c "ZONE:Frozen"`: `auto-semantics.md`·`kanban-dispatch.md`·`run.md`·`spec-assembly.md`·`moai.md` 전부 0. `zone-registry.md`에 이 파일들의 항목 없음. 단 `orchestration-mode-selection.md:18`에는 `[ZONE:Frozen]`이 있으므로 **편집 대상에서 제외**한다 |
| E10 | VCI §2.3 순서 | 반영 | REQ-BGS-015·AC-008: 기준선 산출물은 변경 커밋보다 앞선 별도 커밋. 검증은 커밋 그래프로(`git merge-base --is-ancestor`). 0.3.0: 기준선 위치는 추적되는 SPEC 디렉터리로 옮겼다(R4.1) |
| E11 | 게이트 라운드 임시 계수 | VERIFIED(범위 한정) | R3.0 |
| E12 | 리더 공지 둘의 위치와 구조 | VERIFIED | `Read`: `session_start_factory.go`(팩토리 리더 `factoryLeaderNotice` `:185`, 규율 블록 (e) `:222`, 레인 공지 `factoryLaneNotice` `:264`, 레인 규칙 `:96-130`), `session_start_kanban.go`(칸반 리더 `kanbanLeaderNotice` `:116`, 문맥 블록 (e) `:177-194`, 팩토리 환경 가드 `:50-52`, 동반 공지 `:232`), `session_start_factory_i18n.go`(구조체 `:35-67`, en `:74-124`, ko `:125-169`, ja `:170-214`, zh `:215-258`), `session_start_kanban_i18n.go`(구조체 `:37-51`, en `:60-83`, ko `:84-105`, ja `:106-127`, zh `:128-149`). 방출은 startup(빈 원천 포함)만: `kanbanBootstrapNoticeForSource` `:91-96`, `factoryBootstrapNoticeForSource` `:77-82`. 두 공지는 에이전트용 영어 사본(additionalContext)과 운영자용 현지어 사본(systemMessage)으로 나간다(`session_start_kanban.go:17-19`). 두 공지 모두 지금은 승인·Kickoff 문구가 없다(E8) |
| E12b | 두 리더 공지 모두에 싣는 근거 | VERIFIED(읽기) | (1) 두 공지는 배타적이다: `session_start_kanban.go:50-52`가 팩토리 환경이면 빈 문자열을 내고 `TestKanbanNoticeSuppressedUnderFactoryEnv`(`session_start_factory_test.go:234-244`)가 고정한다. 칸반 쪽에만 두면 팩토리 리더는 문장을 받지 못한다. (2) `kanban-dispatch.md` Scope(`:11`)는 칸반 리더를, "Factory Mode" 절(`:179`)과 Deputy dispatch surface(`:102`, "the `-k`/`-f` leader session")는 두 모드의 리더를 가리킨다. (3) `agent-common-protocol.md` § User Interaction Boundary: 레인은 자기 카드의 질문 채널을 팩토리 리더를 통해 쥔다. (4) 레인·동반이 받지 않는 근거는 REQ-BGS-001(카드 1장)이다. 이 판정은 읽기이며 기계 점검이 아니다 |
| E13 | `internal/hook/CLAUDE.md:13`의 "0 matches" 가드 | **CORRECTED**(문구가 거짓) | 그 파일은 `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v _test.go`가 0줄이어야 한다고 쓴다. 이 트리에서 같은 명령은 **22줄**을 낸다(주석 21 + `pre_tool.go:834`의 도구 이름 문자열 비교 1). 알림 소스 4파일의 토큰 수는 각 0이다. 패키지 전역 grep을 지키는 테스트는 찾지 못했다. 있는 것은 파일 범위 가드 하나다: `TestNoUserInteraction`(`handoff_inject_test.go:418-436`)이 `handoff_inject.go`·`handoff_inject_render.go` 두 파일을 주석 줄을 빼고 스캔한다. 알림 소스 4파일은 어떤 가드에도 걸려 있지 않다(`grep -rln AskUserQuestion internal/hook --include='*_test.go'`의 8개 파일 중 이 둘과 `escalation_m5_test.go`·`session_start_deferred_optout_guard_test.go`·`askuser_observer_test.go`를 읽어 확인했고, 나머지 `protocol_test.go`·`user_decision_capture*_test.go`·`handoff/persist_test.go`는 열어 읽지 않았다 — UNVERIFIED). 실제 경계는 "호출 금지"이며, AC-015는 `TestNoUserInteraction`과 같은 파일 범위·주석 제외 형태로 4파일을 점검한다. 감사 1회차 독립 확인: 같은 grep이 24줄(제안서의 파이프 형태는 줄 자체에 `_test.go`가 든 두 줄을 떨궈 22줄이 된다)이며 알림 소스 4파일은 각 0이다 — 결함이 아니다 |
| E14 | `workflows/moai.md`의 줄 한도 | **CORRECTED**(전제 정정) | 이 파일에는 줄 수 테스트가 걸려 있지 않다. `TestSubSkillLOCCeiling`은 `workflows/` 최상위 `.md`를 건너뛰고(`internal/skills/workflow_split_test.go:104-110`, 주석 "Top-level .md file (entry router) — not a sub-skill, skip"), `TestEntryRouterLOCCeiling`의 대상은 `run.md`·`sync.md`·`project.md`·`plan.md` 넷뿐이다(`:143`). 두 테스트 모두 라이브 트리만 읽는다(`:88`, `:141`). `wc -l` → 라이브 284, 미러 282. 이 파일에 걸린 고정은 contract-mode 블록 `contract-pipeline-gates`·`contract-merged-round`(`internal/template/contract_mode_blocks_test.go:54`), Kickoff 분류 E(`contract_mode_guided_test.go:662`), 고아 `manager-tdd`/`manager-ddd` 참조 스캔 계열의 목록 항목(`agent_frontmatter_audit_test.go:411`). 따라서 "여유"는 정의되지 않는다. M4는 자기 제약으로 두 사본의 줄 수 불변(284/282)을 건다 |
| E15 | `workflows/moai.md` 라이브↔미러 사전 차이 | VERIFIED | 두 사본의 `diff` 헌크 헤더는 `215c215`, `245c245`, `253d252`, `283,284c282` 넷이다(exit 1, R3.1). 편집 대상 `:144`와 `:240`은 두 사본에서 같은 줄 번호에 같은 문장이고(`grep -n "Score-independent\|score-independent\|exactly once per pipeline"`가 두 사본에서 같은 두 줄을 낸다) 사전 차이 헌크 밖이다. `spec-assembly.md`는 두 사본이 동일하다(597줄 / 597줄, E9d) |
| E16 | 도구 출처 (VCI §2.2) | VERIFIED | 0.2.0: 이전 실행의 스크래치 바이너리 `moai-tree`는 `go version -m`이 `vcs.revision=c8f245c2c9a58083…`, `vcs.modified=true`를 보였고 그 커밋은 HEAD의 조상이 아니다(`git merge-base --is-ancestor c8f245c2c9a5 HEAD` → exit 1). 그래서 `-X …/pkg/version.Commit=c50da9c2f`로 다시 빌드했다: `moai-c50da9c2f version` → `v3.2.0-tree   c50da9c2f   built unknown`. 0.3.0: 같은 절차를 `72e09d27b`로 반복했다(R1). 설치본은 `v3.2.0-rc.23`(`d194083fb`)로 HEAD보다 오래됐다 |
| E17 | 리더 공지 테스트의 GREEN 기준선 | VERIFIED | R3.1의 명령: `--- PASS` 63줄, `--- FAIL` 0줄, `=== RUN` 63줄, `ok  github.com/modu-ai/moai-adk/internal/hook  6.364s`, exit 0, 트리 `c50da9c2f`. 이 63개는 새 문장이 들어오기 *전* 값이다. 감사 1회차 독립 재실행(트리 `72e09d27b`): 63/0/63, `ok … 11.305s` |

## R3 측정 원문

### R3.0 임시 측정 (기준선 아님)

목적: 지표 M-1을 *어떻게* 셀 수 있는지, 그리고 제안서의 176이 단일 명령으로 재현되지 않는다는 점을 보인다.

```bash
cd ~/.moai/claude-profiles/moai-adk/projects && find . -name '*.jsonl' -newermt 2026-09-26 -path '*moai-adk-go*' -print0 | xargs -0 grep -h '"type":"tool_use","id":"[^"]*","name":"AskUserQuestion"' | grep -o '"timestamp":"20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]' | sort | uniq -c
```

관측 출력(0.1.0 플랜 실행, 이 머신, 대상 파일 1020개 — `find ... | wc -l` → `1020`):

```text
   8 "timestamp":"2026-09-24
   9 "timestamp":"2026-09-25
  70 "timestamp":"2026-09-26
  23 "timestamp":"2026-09-27
  13 "timestamp":"2026-09-28
  18 "timestamp":"2026-09-29
  17 "timestamp":"2026-09-30
  27 "timestamp":"2026-10-01
   4 "timestamp":"2026-10-02
```

읽는 법과 한계:
- 09-26~09-29 합 = 124(제안서의 176과 다름). 09-30 = 17, 10-01 = 27(자율 전환 도입 이후). 어느 쪽이든 *모든* `AskUserQuestion` 호출이라 게이트와 명확화가 섞여 있다 — 감소 여부를 판단할 근거가 못 된다.
- 파일 범위는 `-newermt 2026-09-26` 수정시각 필터이고 날짜는 이벤트 시각이다. 세션 이어받기·분기 사본이 같은 호출을 두 번 담을 수 있어 과대 계수가 가능하다. 서브에이전트 전사본이 별도 파일이면 포함 여부도 미확인이다.
- 파이프라인 전체의 종료 코드는 기록하지 않았다(Gap). 같은 계수를 프로젝트 디렉터리 하나에만 걸었을 때는 일별 값이 위와 달랐고(예: 09-26이 33), 디렉터리별 반복문에 수정시각 구간 필터를 붙인 변형은 합계 0을 냈다(원인은 확인하지 않았다 — 구간 필터가 이어받은 세션을 걸렀을 가능성은 가설일 뿐이다). 즉 이 명령은 *범위 정의에 민감*하다. 그래서 M0는 범위를 먼저 정의해야 한다.
- 전사본 저장소는 프로젝트 디렉터리 406개(워크트리 디렉터리 336개)로 쪼개져 있다(`ls | wc -l`, `ls | grep -c worktrees`).
- 0.3.0에서 이 명령은 다시 돌리지 않았다. 위 표는 0.1.0 실행의 관측이며 오늘 트리에 대한 baseline이 아니다.

### R3.1 추가 측정 (0.2.0, 트리 `c50da9c2f`, 브랜치 `WT-batch-approval-gate`)

E13 — 패키지 전역 가드 문구 대 실제 출력. 명령과 관측 출력(22줄, 파이프라인 exit 0):

```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go
```

```text
internal/hook/post_tool.go:292:	// Capture AskUserQuestion user decisions into the preference memory layer
internal/hook/user_decision_capture.go:6:// When a PostToolUse event names the AskUserQuestion tool, this subpipeline
internal/hook/user_decision_capture.go:42:// INVOKING the AskUserQuestion API or mcp__askuser__* tools; this subpipeline
internal/hook/user_decision_capture.go:110:	// schema tolerance). AskUserQuestion tool_result payload shape varies
internal/hook/user_decision_capture.go:183:// PostToolUse result payload) and tool_input (the original AskUserQuestion
internal/hook/session_guard.go:26:// calls AskUserQuestion, and tolerates every error (missing registry file, parse
internal/hook/session_ledger.go:20:// (REQ-SEG-008, ≤5s). AskUserQuestion / mcp__askuser 미호출 (REQ-SEG-007, C-HRA-008).
internal/hook/pre_tool.go:27:// pattern matching without AskUserQuestion interaction.
internal/hook/pre_tool.go:32:// [HARD] No AskUserQuestion calls. Deny reason is emitted as sentinel string.
internal/hook/pre_tool.go:828:	// AskUserQuestion observation. The third branch of the same kind as the two
internal/hook/pre_tool.go:834:	if input.ToolName == "AskUserQuestion" {
internal/hook/askuser_observer.go:25:// structured row per observed AskUserQuestion issuance.
internal/hook/askuser_observer.go:34:// AskUserObservationRecord is one JSONL row per AskUserQuestion issuance.
internal/hook/askuser_observer.go:42:// AskUserQuestion payload carries no type tag, and whether a report preceded
internal/hook/askuser_observer.go:99:// askUserPayload is the parsed shape of an AskUserQuestion tool_input. Only
internal/hook/askuser_observer.go:168:// observeQuestionChannel is the PreToolUse(AskUserQuestion) entry point. It is
internal/hook/subagent_stop.go:103:	// [HARD] This package does not call AskUserQuestion. Errors emit to slog only.
internal/hook/subagent_stop.go:276:// [HARD] No AskUserQuestion call — this package is a subagent-level hook handler.
internal/hook/handoff_inject.go:11:// AskUserQuestion (subagent boundary, REQ-AUTORESUME-016).
internal/hook/handoff_inject.go:44:// @MX:REASON: registry에 3번째 SessionStart 핸들러로 등록 (deps.go). 중복 주입은 factory.db의 status='pending' 조건부 UPDATE가 막는다. claim-then-inject 순서(claim 성공 후 주입)는 유지한다. DB 오류는 fail-open이다. manual mode는 stale이어도 pure no-op (REQ-009 vs REQ-019 모순 해소). AskUserQuestion 미호출 (C-HRA-008).
internal/hook/security/guardian.go:12:// invokes AskUserQuestion or emits a prose question (REQ-SG-040); a block rides
internal/hook/handoff/persist.go:17:// does not invoke AskUserQuestion or write to stdout/stderr in any
```

알림 소스 4파일의 토큰 수(단일 명령, exit 1 — 적중 없음. 줄 순서는 인자 순서와 다르게 찍혔다):

```bash
grep -c AskUserQuestion internal/hook/session_start_kanban.go internal/hook/session_start_kanban_i18n.go internal/hook/session_start_factory.go internal/hook/session_start_factory_i18n.go
```

```text
internal/hook/session_start_factory_i18n.go:0
internal/hook/session_start_kanban_i18n.go:0
internal/hook/session_start_factory.go:0
internal/hook/session_start_kanban.go:0
```

E14·E15 — `workflows/moai.md`. 줄 수와 줄 한도 테스트의 대상 목록(`internal/skills/workflow_split_test.go:143-144`):

```bash
wc -l .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md
```

```text
     284 .claude/skills/moai/workflows/moai.md
     282 internal/template/templates/.claude/skills/moai/workflows/moai.md
     566 total
```

```text
	entryRouters := []string{"run.md", "sync.md", "project.md", "plan.md"}
	const maxLOC = 200
```

사전 차이(편집 전, exit 1) — 헌크 헤더만 옮긴다:

```bash
diff .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md
```

```text
215c215
245c245
253d252
283,284c282
```

E16 — 폐기한 바이너리의 VCS 스탬프와 다시 빌드한 바이너리(0.2.0):

```text
	build	vcs=git
	build	vcs.revision=c8f245c2c9a58083518f0b7cbebff0542848e033
	build	vcs.time=2026-09-29T06:19:47Z
	build	vcs.modified=true
```

```text
 v3.2.0-tree   c50da9c2f   built unknown
```

E17 — 리더 공지 테스트 GREEN 기준선(23개 이름, 새 문장 *전*). 명령(출력은 파일로 돌려 `--- PASS` 63줄 · `--- FAIL` 0줄 · `=== RUN` 63줄을 센 값):

```bash
go test -count=1 -v -run '^(TestKanbanLocalesCoverEveryField|TestKanbanNoticePreservesProtocolTokensInEveryLocale|TestKanbanLeadNoticeBlockLayout|TestKanbanLeadNoticeFullContent|TestKanbanLeadNoticeOmitsSPECWhenUnset|TestKanbanLeadNoticeNamesTheLeadSession|TestKanbanLeadNoticeBacklogSummaryCountsQueuedOnly|TestKanbanLeadNoticeBacklogSummaryEveryLocale|TestSessionStartKanbanRespectsTodoDisabled|TestSessionStartKanbanChannelsCarryTheirOwnLanguage|TestKanbanRecommendationTableMatchesLaunchLines|TestKanbanCompanionNoticeCarriesSpawnAuthority|TestKanbanNoticeSuppressedUnderFactoryEnv|TestFactoryWorkerNoticeCarriesSpawnAuthority|TestFactoryWorkerNoticeNamesLabel|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide|TestFactoryLeadNoticeCarriesDispatchDiscipline|TestFactoryLeadNoticeDispatchDisciplineKorean|TestFactoryLeadNoticeAllSlotsClaimed|TestFactoryBootstrapNoticeStartupOnly|TestFactoryGuideTeachesLaneFormsInEveryLocale|TestSD_AC019_NextCardRuleInjection|TestRoleNamingM3NoticesCarryLeaderLaneTerms)$' ./internal/hook/
```

```text
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	6.364s
```

## R4 플랜 감사 1회차 반영 중 새로 확인한 전제 (0.3.0, 트리 `72e09d27b`)

### R4.1 기준선 경로의 추적 여부 (결함 D1)

감사 보고서는 `.moai/reports/t1344/baseline-gate-rounds.md`가 `.gitignore:235`에 걸린다고 봤다. 이번 개정이 두 경로와 양성 대조를 같은 트리에서 다시 돌렸다(각 명령은 단일 호출, 종료 코드는 별도 필드).

| 명령 | stdout | exit | 읽는 법 |
|---|---|---|---|
| `git check-ignore -v .moai/reports/t1344/baseline-gate-rounds.md` | `.gitignore:235:.moai/reports/*	.moai/reports/t1344/baseline-gate-rounds.md` | 0 | 옛 위치는 무시된다 |
| `git check-ignore -v .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md` | (비어 있음) | 1 | 새 위치는 무시되지 않는다 |
| `git check-ignore -v .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/plan.md` | (비어 있음) | 1 | 양성 대조 ①: 추적되는 SPEC 파일은 무시되지 않는다 |
| `git check-ignore -v .moai/reports/t1344/plan-audit-iter1.md` | `.gitignore:235:.moai/reports/*	.moai/reports/t1344/plan-audit-iter1.md` | 0 | 양성 대조 ②: 이 보고서 경로의 기존 파일은 같은 규칙으로 무시된다 |
| `git ls-files .moai/reports/t1386 .moai/reports/t1381` | (비어 있음) | 0 | 최근 카드의 보고서는 추적되지 않는다(`git ls-files`는 무적중에도 exit 0) |
| `git ls-files .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001` | `acceptance.md`·`decision-index.md`·`plan.md`·`progress.md`·`spec-compact.md`·`spec.md` | 0 | SPEC 디렉터리는 추적된다 |

해석: `.gitignore:233-246`의 주석이 운영자 지시(2026-09-14, 보고서는 로컬 전용)와 철회된 verdict 예외("Do not re-add it")를 기록한다. `.moai/reports/` 아래에 추적되는 파일(`git ls-files -c .moai/reports --format="%(path)" -- ':(glob).moai/reports/*/verdict.md'`가 5개 이상의 `…/verdict.md`를 내고 `.moai/reports/plan-audit/`에 스캐폴드와 옛 검토 파일이 있다)이 있다는 0.2.0의 관측은 사실이지만, 그것은 지시 이전에 강제 추가된 옛 항목이 인덱스에 남은 것이다. "보고서 아래 산출물은 추적 대상이다"는 최근 카드에는 거짓이다. 그래서 기준선을 추적되는 `.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`로 옮긴다(오케스트레이터 판정 R2).

### R4.2 §9.1 문면과 PASS-WITH-DEBT·BYPASSED의 침묵 (결함 D3)

- `grep -n "PASS-WITH-DEBT\|BYPASSED" .claude/rules/moai/workflow/auto-semantics.md` → 적중 없음(출력 비어 있음). §9.1(`:175-176`): "the independent plan-audit verdict is PASS (FAIL and INCONCLUSIVE stay hard blocks — the §7 authority-gate invariant)". §7(`:126-132`): "only a positive verdict proceeds".
- `grep -n "PASS-WITH-DEBT" .claude/agents/moai/plan-auditor.md` → `:203` `verdict: <PASS|PASS-WITH-DEBT|FAIL>`, `:231` `AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL> spec=…`. 감사 표면은 PASS-WITH-DEBT를 낸다.
- `spec-workflow.md`의 Plan Audit Gate 생략 계약: "Verdict is `PASS` (NOT FAIL, NOT INCONCLUSIVE, NOT BYPASSED)", Verdicts 표에 `BYPASSED`("User passed `--skip-audit` …").
- 결론(문면에서 나온 사실): §9.1은 PASS를 요구하고 FAIL·INCONCLUSIVE만 강한 차단으로 명명한다. PASS-WITH-DEBT와 BYPASSED는 §9.1이 말하지 않는다. 따라서 §9.1 문면의 PASS가 아니며 REQ-BGS-011은 이를 차단으로 다룬다.

### R4.3 해시 대상 집합 (결함 D1·D6)

`grep -n "planArtifactNames" -A 12 internal/runtime/audit_cache.go` → `:89-96`: `"acceptance.md"`, `"design.md"`, `"plan.md"`, `"research.md"`, `"spec.md"`, `"tasks.md"` 여섯 이름. `ComputeHash`(`:120-132`)는 없는 파일을 건너뛴다. 따라서 `baseline-gate-rounds.md`, `spec-compact.md`, `decision-index.md`, `progress.md`는 해시 대상이 아니고 기준선 파일을 SPEC 디렉터리에 더해도 캐시된 감사 판정이 무효가 되지 않는다. 반대로 `design.md`와 `research.md`는 Tier L의 해시 대상이다.

### R4.4 Kickoff 문구 형제 파일 정독 (결함 D17)

방법: `grep -rIl -i "kickoff" .claude/skills .claude/rules .claude/agents .moai/docs` → 38개 파일. 파일마다 `grep -n -i "kickoff" <파일>`로 줄을 뽑아 읽었다(본문 정독이 아니라 적중 줄과 그 문장의 읽기이며, 분류는 **읽기 판정이지 기계 점검이 아니다**). 분류 기준:
- **영향(A)**: 카드별 운영자 답이 필요하다고 현재형으로 단정하거나 감사 PASS·생략 가능 점수가 운영자 질문을 대체하지 못한다고 단정하는 문장.
- **표지만(B)**: "HUMAN GATE"·"human gate" 표지를 쓰지만 운영자 답의 필요를 단정하지 않는 문장(조건 순서·선행 조건 인용).
- **무관(C)**: 다른 게이트이거나, 이미 §9.1을 인용하거나, 게이트가 *충족돼야 한다*만 말해 자율·운영자 두 형태 모두에서 참인 문장.

| 분류 | 파일:줄 | 읽은 내용 |
|---|---|---|
| A | `.claude/agents/harness/workflow-specialist.md:72-77` | "a HARD human gate … MUST obtain explicit approval before `/moai run`. This gate is NOT bypassed by a skip-eligible plan-auditor verdict (≥0.90)" — 카드별 운영자 답 단정이며 폐기된 `≥0.90` 술어도 인용 |
| A | `.claude/skills/hns-moaiadk-patterns/SKILL.md:150` | "Add a SPEC" 절차의 3단계 "Implementation Kickoff Approval human gate (orchestrator runs `AskUserQuestion`)" |
| A | `.claude/rules/moai/workflow/session-handoff-examples.md:237` | "remains mandatory regardless of the seeded mode … the orchestration shape the user is subsequently asked to approve" |
| A | `.claude/skills/moai/workflows/project/doc-generation.md:384,392` | pipeline 이어받기 가지가 "emits that gate and stops for the operator's answer; it never selects, answers, pre-fills, or defaults it"; `:361,371`도 "emitting the Implementation Kickoff Approval gate" (프로젝트 이어받기 설정의 문맥이라 §9.1 keep-set 읽기와의 충돌 여부는 판정이다) |
| B | `.claude/skills/moai/workflows/plan.md:53,73` | `:53` 표지 "(plan→run HUMAN GATE)", `:73` "proceeds only after all clarifications are resolved"(조건 순서) |
| B | `.claude/skills/moai-workflow-spec/SKILL.md:173,191` | `:173` 표지 "(plan→run HUMAN GATE)", `:191` "(mandatory human gate) proceeds only after … resolved" |
| B | `.claude/skills/moai/workflows/design.md:48,52` | "the plan→run human gate is cleared", "ahead of the human gate" |
| B | `.claude/rules/moai/workflow/session-handoff-examples.md:242,250` | `:242` 표지 "(plan→run HUMAN GATE) remains mandatory", `:250` "the … human gate remains required" |
| B | `.claude/skills/moai/SKILL.md:166,372` | "the gate stays mandatory in both modes", "Kickoff Approval remains mandatory at the plan→run boundary" — 게이트가 충족돼야 한다는 뜻으로 읽힌다(0.2.0 판정 유지). 두 사본이 이미 달라 손대지 않는다 |
| C | `.claude/skills/moai/workflows/goal.md:192,221,267,273,294`; `.claude/rules/moai/workflow/goal-directive.md:23,25,28,33`; `goal-directive-detail.md:50,66-70`; `session-handoff.md:60,64,73,118`; `spec-workflow.md:71,73,316,344,346,359`; `cadence-bridge.md:26,107`; `askuser-protocol.md:111,160`; `orchestration-mode-selection.md` | 이미 `auto-semantics.md` §9.1을 인용한다(기본 자율, keep-set은 운영자 형태) 또는 선행 조건 인용 |
| C | `.claude/skills/moai/workflows/factory.md:57`; `run/mode-orchestration.md:88`; `run/task-decomposition.md:133`; `run/phase-execution.md:265,273,497`; `.claude/agents/moai/manager-develop.md:58,166`; `.claude/agents/moai/manager-design.md:23`; `.moai/docs/session-handoff-appendix.md:25,68` | 게이트가 통과됐다는 선행 조건 인용. 자율·운영자 두 형태 모두에서 참 |
| C | `.claude/skills/moai/workflows/e2e.md:49,240,242,333,346`; `harness-build-entry.md:116` | 다른 게이트(autofix 루프 승인, 하니스 Builder 승인)이고 plan→run Kickoff가 아니다 |
| C | `.claude/rules/moai/workflow/dynamic-workflows.md:58,99,108`; `archived-agent-rejection.md:155`; `coding-standards.md:146`; `moai-mcp-tools-catalogue.md:139,233` | 게이트가 실행 *앞서* 결정된다는 선행 조건 또는 유추 인용(`coding-standards.md`), contract 모드의 `llm+jev` 교차 확인(`moai-mcp-tools-catalogue.md`) |
| C | `.claude/rules/moai/workflow/contract-autonomy.md`; `.moai/docs/jev-local-operations.md:32`; `.claude/skills/moai-kanban-foreman/SKILL.md:73`; `.claude/agents/moai/plan-auditor.md:148` | contract 모드 정본, 로컬 3등급 문서, 포어맨이 대신 답하지 않는 판단 목록(운영자 형태를 대신 답하지 않는다는 뜻이라 양립), 감사 MP-7 |
| (이 SPEC의 편집 대상) | `spec-assembly.md:202-208`; `workflows/moai.md:144,240`; `kanban-dispatch.md`; `run.md`; `auto-semantics.md` | REQ-BGS-013·014 |

결과: 영향(A) 네 파일, 표지만(B) 다섯 파일(`session-handoff-examples.md`는 A와 B 모두에 줄이 있다). 이들은 이 SPEC이 고치지 않으며 범위 확대는 decision-index Q12로 남았다. 양성 대조: 같은 방식의 `grep -n -i "kickoff"`가 이 SPEC의 편집 대상 `moai.md:144,240`과 `spec-assembly.md:202`를 적중시킨다.

### R4.5 `kanban-dispatch.md:106` 목록 원문 (결함 D4)

`grep -n "The deputy never holds" .claude/rules/moai/workflow/kanban-dispatch.md` → `:106`: "**The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done` (`moai gtd` mutations), CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader session." — 여섯 항목: 최종 PASS/FAIL 판정, 최종 병합 승인, 운영자 게이트, 카드 발행과 `done`(큐 변경), CodeRabbit 슬롯 대기 판정, 세션 간 분쟁 조정. "user-facing behavior changes"는 이 줄에 없다. 그 구절은 로컬 문서 `.moai/docs/jev-local-operations.md:30`에만 있고 출하되지 않는다(E3).

### R4.6 질문 채널 상한 (결함 D9)

`askuser-protocol.md` Structural Constraints: "Maximum 4 questions per `AskUserQuestion` call", "Maximum 4 options per question". Free-form Circumvention Prohibition: "`AskUserQuestion` automatically appends an 'Other' option to every question set". 따라서 질문 하나에 행마다 선택지를 두는 방식은 행이 4개를 넘으면 성립하지 않는다.

### R4.7 에이전트용 사본은 구성상 영어 (결함 D11)

`session_start_kanban.go:17-19`: "Each audience also reads its own language: the agent-facing copy is rendered with langEnglish and the operator-facing copy with the configured conversation_language." `session_start_kanban_i18n.go:17-26`: `agent_prompt_language` (English)은 에이전트가 읽는 것에, `conversation_language`는 운영자가 읽는 것에; `langEnglish`는 모든 에이전트용 사본의 언어.

## R5 계획 시점 재계수 (결함 D6·D15, 트리 `72e09d27b`)

| 명령(단일 호출) | 출력 | exit |
|---|---|---|
| `git ls-files <기존 파일 15개 경로>` — 규칙·스킬 문서 10, 제품 Go 4, `session_start_kanban_i18n_test.go` 1 | 15줄 | 0 |
| `ls internal/template/batch_gate_summary_doctrine_test.go internal/hook/session_start_leader_gate_notice_test.go` | `ls: …session_start_leader_gate_notice_test.go: No such file or directory` · `ls: …batch_gate_summary_doctrine_test.go: No such file or directory` | 1 |
| `git diff --name-only develop...HEAD -- . ':(exclude).moai/specs' ':(exclude).moai/reports'` | (비어 있음) | 0 |
| `git diff --name-only develop...HEAD` (대조: 제외 없음) | `.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/` 아래 여섯 줄 | 0 |
| `git cat-file -s c50da9c2f:.claude/rules/moai/workflow/kanban-dispatch.md` | `26807` | 0 |
| `git cat-file -s c50da9c2f:internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `26485` | 0 |
| `git cat-file -s c50da9c2f:.claude/skills/moai/workflows/run.md` | `20235` | 0 |
| `wc -c` 위 세 파일의 현재 값 | `26807`, `26485`, `20235`(미러 `run.md`도 `20235`) | 0 |

읽는 법: 기존 15 + 신규 2 = 열거 17. 변경 전 트리에서 재계수 명령은 비고, 제외 없는 대조는 SPEC 산출물을 낸다(양성 대조로 명령 형태가 발화함을 보인다). 바이트 기준값은 고정 SHA `c50da9c2f` 기준이며 `kanban-dispatch.md`는 라이브와 미러가 이미 다르다(E9d).

## R6 출처 조사의 세 결과

이 SPEC의 근거가 된 조사에서 설계를 가장 크게 바꾼 세 결과다. 본문은 `spec.md` §A.2가 소유하고 여기서는 요약과 포인터만 둔다.

1. **"176"은 재현되지 않고 게이트 한정도 아니다** (`spec.md` §A.2 #1·§E G-2, R3.0). 제안서의 숫자는 3일·20세션 합산이고 이 머신의 전사본 계수는 모든 `AskUserQuestion`의 일별 합이어서 어느 쪽도 기준선이 아니다. 그래서 결과 수치를 단언하지 않고 M0 기준선을 먼저 둔다.
2. **"카드 5장 직렬 대기"는 Kickoff가 아니라 하나의 `served_model_gate` 판정이었다** (`spec.md` §A.2 #2). 같은 판정이 여러 카드의 공통 결정 주체일 때 한 번 묻고 카드를 전부 적는 규칙(REQ-BGS-003)이 거기서 왔다.
3. **리더 공지는 둘이고 배타적이며, 제안서의 "3등급"은 로컬 독트린이다** (`spec.md` §A.2 #8·#9). 두 리더 공지에 모두 싣는 이유(REQ-BGS-016·017)와 예약 목록을 출하된 문서의 표현(keep-set 정의, `kanban-dispatch.md:106`)으로 한정하는 이유(REQ-BGS-010)가 여기서 왔다.

"세 결과"를 위 셋으로 읽은 것은 이 문서 작성자의 해석이다. 요청문은 셋을 이름으로 지정하지 않았다.
