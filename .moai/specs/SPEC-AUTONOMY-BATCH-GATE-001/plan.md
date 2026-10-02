# SPEC-AUTONOMY-BATCH-GATE-001 — Plan

> 이 문서는 구현 계획이다. 요구사항은 `spec.md`, 검증은 `acceptance.md`, 열린 결정은 `decision-index.md`가 소유한다. 시간 추정 없이 우선순위와 순서만 쓴다.

## §A Context

- 에픽 소속 없음(단독 SPEC). 형제: `SPEC-LANE-STALL-WATCHDOG-001`(`auto-semantics.md` 저작), `SPEC-AUTONOMY-KICKOFF-CALIB-001`, `SPEC-AUTONOMY-GATE-REWIRE-001`, `SPEC-AUTONOMY-ESCALATION-001`.
- 위치: 워크트리 `.moai/worktrees/t1344`, 브랜치 `WT-batch-approval-gate`, HEAD `c50da9c2f`(= 로컬 `develop` 끝). Tier M: `spec.md` + `plan.md` + `acceptance.md` (+ `spec-compact.md`, `decision-index.md`, `progress.md`).
- 변경 성격(0.2.0): **문서(규칙·스킬) + 리더 공지 문장(제품 Go 파일 4개: `internal/hook/session_start_{kanban,factory}{,_i18n}.go`) + 테스트 Go 파일 3개**(훅 2, 템플릿 1). 제품 Go 변경은 리더 공지 문자열에 한정된다(`spec.md` §B.7). 0.1.0의 "제품 Go 변경 없음"은 Q5 판정(2026-10-02)으로 철회됐다.
- Tier 재측정: 계획된 변경 파일이 미러를 따로 세면 18개, 쌍을 한 파일로 세면 13개다(`spec.md` §A.5). `tier:`는 바꾸지 않았고 판정은 오케스트레이터 소관이다.
- 요약: 같은 게이트 행에서 운영자 결정을 기다리는 카드들을 한 번에 보여 주고 질문 하나로 승인받는 *제시 형식*을 `auto-semantics.md` §9.2로 정본화하고, 반대 증거 의무와 fail-closed 제외 규칙을 못 박는다. 효과의 크기는 미측정이므로(G-1) 결과 수치는 약속하지 않는다.

## Evidence and premise verification

플랜 단계에서 직접 확인한 근거. 상태: VERIFIED(이번 실행에서 관측), CORRECTED(전제가 틀려 정정), UNVERIFIED(관측하지 못함).

| ID | 주장 | 상태 | 근거 (명령/위치 + 관측) |
|---|---|---|---|
| E1 | 제안서 P4 `:38`(크기 S, "176→10~20"), 게이트 라운드 행 `:12`(약 176, T4 5파일 155), 병목 #1 `:22`, 리스크 `:54` | VERIFIED | `Read /Users/goos/MoAI/moai-adk-go/.moai/reports/autonomy-bottleneck-proposal-20260929.md`로 전문 열람 |
| E1b | 그 보고서는 이 워크트리에 없다(미추적) | VERIFIED | `ls .moai/worktrees/t1344/.moai/reports/autonomy-bottleneck-proposal-20260929.md` → `No such file or directory` |
| E2 | `auto-semantics.md`가 보고서보다 뒤에 도입됐다 | VERIFIED(날짜) / 가정(도입 커밋) | `git log --oneline -1 --format='%h %ad' --date=short 147c25d77` → `147c25d77 2026-09-30`. 이 커밋이 파일을 *도입*했다는 점은 형제 SPEC `§A.4`("NEW doctrine rule")를 따랐고 diff로 확인하지 않았다 |
| E2b | §9 목록(`:152-171`), §9.1(`:173-181`), keep-set 문구(`:154-158`) | VERIFIED | `Read .claude/rules/moai/workflow/auto-semantics.md` 전문 |
| E2c | "t1318 owner conf 0.38"은 근거 불명 | UNVERIFIED | 이 수치의 출처는 제안서 `:54` 자기 자신뿐. primary의 `.moai/reports/t1318/`(파일: `kickoff-20261001.md`, `plan-20261001.md` 등)에서 `grep -rIn "0\.38"` → 적중 없음(양성 대조는 돌리지 않았다 — 약한 부재 증거). 다른 보고서의 `0.38` 적중은 열어 보지 않았다. 3등급 *규칙*만 쓴다 |
| E3 | 3등급 = 리드 3등급 독트린(절대 위임 불가). 로컬 문서이며 배포되지 않음 | VERIFIED | `sed -n 14,34p .moai/docs/jev-local-operations.md`(3등급 정의 `:20-30`). 템플릿 쪽 적중 없음: `grep -rIln "ASK-OPERATOR\|route.sh\|3-grade\|3등급" internal/template/templates .claude/rules/moai .claude/skills` → 출력 없음. `ls scripts/jev/route.sh` → `No such file or directory`(exit 1); `scripts/jev/`에는 `triage.py`, `test_triage.py`만 있다 |
| E3b | 배포되는 대응물 | VERIFIED | `kanban-dispatch.md:106`("The deputy never holds a power of consequence" — 최종 PASS/FAIL, `LEAD-MERGE-APPROVED`, 운영자 게이트, `moai gtd` 변경, CodeRabbit 슬롯 대기); 템플릿 미러에도 `:106` 존재(`grep -n` 확인) |
| E4 | 상충 표현 위치 | VERIFIED | `spec-assembly.md:202`("stays MANDATORY"), `moai.md:144,240`(`grep -n`), 대조: `run.md:127-139`, `kanban-dispatch.md:187`, `orchestration-mode-selection.md:16,18` |
| E4b | 형제 SPEC 개정 목록이 두 파일을 다루지 않았다 | VERIFIED | `grep -n "spec-assembly" .moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md` → 적중 없음(exit 1); `grep -n "workflows/moai.md" …` → 적중 없음(exit 1). 양성 대조: 같은 파일에서 `grep -n "kanban-dispatch.md"`는 `:108`, `:169` 등에서 적중. 즉 개정 행 A–N과 "검토 후 변경 불필요" 목록 어디에도 두 파일명이 없다 |
| E5 | 용어 충돌 | VERIFIED | `kanban-dispatch.md:31`("`/moai:todo --auto` is the operator's batch approval"), `:102,177`(배치=디스패치/릴리스 배치), `gtd.md:334`, `auto-semantics.md:169` |
| E6 | 재사용할 서식과 결손 | VERIFIED | §10 `:183-199`(결정 기록 3필드, 반대 증거 필드 없음), §7 `:113-132`(권한 게이트 불변식), §11 결정 보드, §14 대기 기록 |
| E7 | `factory decide` 다중 카드, 단일 gate/choice, 기록 없음 | VERIFIED | `sed -n 1465,1495p internal/cli/factory_card.go`(순회 루프 `:1469-1491`); `mcp_factory_card.go:70-79,205-217`(카드 1장 단위, `factoryLaneRefusal()` 거부) |
| E8 | 주입 문장에 승인 언급 없음 | VERIFIED | `grep -n "AskUserQuestion\|Kickoff\|plan-audit\|approval" internal/hook/session_start_factory_i18n.go internal/hook/session_start_kanban_i18n.go internal/hook/lane_spawn_authority.go` → 적중은 `lane_spawn_authority.go:10` 주석("approval") 하나. `grep -rn "AskUserQuestion" internal/hook \| grep -v _test.go`에는 `post_tool.go:292`, `user_decision_capture.go` 주석·상수 적중이 있다(코드 호출 아님). 단 `internal/hook/CLAUDE.md:13`은 이 종류의 grep이 0줄이어야 한다고 쓰므로 그 문구는 이 트리에서 거짓이다 — 0.2.0에서 재측정, E13 |
| E9 | LOC 한도 | VERIFIED | `wc -l`: `run.md` 199, `spec-assembly.md` 597, `kanban-dispatch.md` 199, `auto-semantics.md` 256. `go test -count=1 -v -run '^(TestSubSkillLOCCeiling\|TestEntryRouterLOCCeiling)$' ./internal/skills/` → 두 테스트 `--- PASS`, `ok` |
| E9b | 상시 로딩 예산 여유 | **CORRECTED** | 코드 주석은 "168 tokens of 77,600"이라 쓰지만 `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/` → `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)`, `--- PASS`. 주석이 낡았다(이 SPEC의 몫이 아님). 그래도 상시 로딩 파일 증가는 `rule-authoring.md` 1,000바이트 기준을 따른다 |
| E9c | 고정 가드의 현재 상태(GREEN 기준선) | VERIFIED | `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/` → `--- PASS`(하위 테스트 9개); `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment\|TestAutoRankMirrorParity\|TestSpecAssembly_RewrittenToCLIPath)$' ./internal/cli/` → 3개 `--- PASS`. 이들은 이번 플랜 트리(HEAD `c50da9c2f`)에서 *변경 전* 측정이다 |
| E9d | 미러 현황 | VERIFIED | `diff -q`: `auto-semantics.md`·`run.md`·`spec-assembly.md` 동일; `kanban-dispatch.md`는 `:177`에서, `moai.md`는 `:215,245,253,283-284`에서 이미 다름(사전 존재 — 건드리지 않는다) |
| E9e | Frozen 구역 | VERIFIED | `grep -c "ZONE:Frozen"`: `auto-semantics.md`·`kanban-dispatch.md`·`run.md`·`spec-assembly.md`·`moai.md` 전부 0. `zone-registry.md`에 이 파일들의 항목 없음. 단 `orchestration-mode-selection.md:18`에는 `[ZONE:Frozen]`이 있으므로 **편집 대상에서 제외**한다 |
| E10 | VCI §2.3 순서 | 반영 | REQ-BGS-015·AC-008: 기준선 산출물은 변경 커밋보다 앞선 별도 커밋. 검증은 커밋 그래프로(`git merge-base --is-ancestor`) |
| E11 | 게이트 라운드 임시 계수 | VERIFIED(범위 한정) | 아래 "임시 측정" |
| E12 | 리더 공지 둘의 위치와 구조 (0.2.0) | VERIFIED | `Read`: `session_start_factory.go`(팩토리 리더 `factoryLeaderNotice` `:185`, 규율 블록 (e) `:222`, 레인 공지 `factoryLaneNotice` `:264`, 레인 규칙 `:96-130`), `session_start_kanban.go`(칸반 리더 `kanbanLeaderNotice` `:116`, 문맥 블록 (e) `:177-194`, 팩토리 환경 가드 `:50-52`, 동반 공지 `:232`), `session_start_factory_i18n.go`(구조체 `:35-67`, en `:74-124`, ko `:125-169`, ja `:170-214`, zh `:215-258`), `session_start_kanban_i18n.go`(구조체 `:37-51`, en `:60-83`, ko `:84-105`, ja `:106-127`, zh `:128-149`). 방출은 startup(빈 원천 포함)만: `kanbanBootstrapNoticeForSource` `:91-96`, `factoryBootstrapNoticeForSource` `:77-82`. 두 공지는 에이전트용 영어 사본(additionalContext)과 운영자용 현지어 사본(systemMessage)으로 나간다(`session_start_kanban.go:9-21`). 두 공지 모두 지금은 승인·Kickoff 문구가 없다(E8) |
| E12b | 두 리더 공지 모두에 싣는 근거 | VERIFIED(읽기) | (1) 두 공지는 배타적이다: `session_start_kanban.go:50-52`가 팩토리 환경이면 빈 문자열을 내고 `TestKanbanNoticeSuppressedUnderFactoryEnv`(`session_start_factory_test.go:234-244`)가 고정한다. 칸반 쪽에만 두면 팩토리 리더는 문장을 받지 못한다. (2) `kanban-dispatch.md` Scope(`:11`)는 칸반 리더를, "Factory Mode" 절(`:179`)과 Deputy dispatch surface(`:102`, "the `-k`/`-f` leader session")는 두 모드의 리더를 가리킨다. (3) `agent-common-protocol.md` § User Interaction Boundary: 레인은 자기 카드의 질문 채널을 팩토리 리더를 통해 쥔다. (4) 레인·동반이 받지 않는 근거는 REQ-BGS-001(카드 1장)이다. 이 판정은 읽기이며 기계 점검이 아니다 |
| E13 | `internal/hook/CLAUDE.md:13`의 "0 matches" 가드 | **CORRECTED**(문구가 거짓) | 그 파일은 `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v _test.go`가 0줄이어야 한다고 쓴다. 이 트리에서 같은 명령은 **22줄**을 낸다(주석 21 + `pre_tool.go:834`의 도구 이름 문자열 비교 1). 알림 소스 4파일의 토큰 수는 각 0이다. 패키지 전역 grep을 지키는 테스트는 찾지 못했다. 있는 것은 파일 범위 가드 하나다: `TestNoUserInteraction`(`handoff_inject_test.go:418-436`)이 `handoff_inject.go`·`handoff_inject_render.go` 두 파일을 주석 줄을 빼고 스캔한다. 알림 소스 4파일은 어떤 가드에도 걸려 있지 않다(`grep -rln AskUserQuestion internal/hook --include='*_test.go'`의 8개 파일 중 이 둘과 `escalation_m5_test.go`·`session_start_deferred_optout_guard_test.go`·`askuser_observer_test.go`를 읽어 확인했고, 나머지 `protocol_test.go`·`user_decision_capture*_test.go`·`handoff/persist_test.go`는 열어 읽지 않았다 — UNVERIFIED). 실제 경계는 "호출 금지"이며, AC-015는 `TestNoUserInteraction`과 같은 파일 범위·주석 제외 형태로 4파일을 점검한다. 아래 "추가 측정" 참조 |
| E14 | `workflows/moai.md`의 줄 한도 (0.2.0) | **CORRECTED**(전제 정정) | 이 파일에는 줄 수 테스트가 걸려 있지 않다. `TestSubSkillLOCCeiling`은 `workflows/` 최상위 `.md`를 건너뛰고(`internal/skills/workflow_split_test.go:104-110`, 주석 "Top-level .md file (entry router) — not a sub-skill, skip"), `TestEntryRouterLOCCeiling`의 대상은 `run.md`·`sync.md`·`project.md`·`plan.md` 넷뿐이다(`:143`). 두 테스트 모두 라이브 트리만 읽는다(`:88`, `:141`). `wc -l` → 라이브 284, 미러 282. 이 파일에 걸린 고정은 contract-mode 블록 `contract-pipeline-gates`·`contract-merged-round`(`internal/template/contract_mode_blocks_test.go:54`), Kickoff 분류 E(`contract_mode_guided_test.go:662`), 고아 `manager-tdd`/`manager-ddd` 참조 스캔 계열의 목록 항목(`agent_frontmatter_audit_test.go:411`). 따라서 "여유"는 정의되지 않는다. M4는 자기 제약으로 두 사본의 줄 수 불변(284/282)을 건다 |
| E15 | `workflows/moai.md` 라이브↔미러 사전 차이 (0.2.0) | VERIFIED | 두 사본의 `diff` 헌크 헤더는 `215c215`, `245c245`, `253d252`, `283,284c282` 넷이다(exit 1, 아래 "추가 측정"). 편집 대상 `:144`와 `:240`은 두 사본에서 같은 줄 번호에 같은 문장이고(`grep -n "Score-independent\|score-independent\|exactly once per pipeline"`가 두 사본에서 같은 두 줄을 낸다) 사전 차이 헌크 밖이다. `spec-assembly.md`는 두 사본이 동일하다(597줄 / 597줄, E9d) |
| E16 | 도구 출처 (VCI §2.2, 0.2.0) | VERIFIED | 이전 실행의 스크래치 바이너리 `moai-tree`는 `go version -m`이 `vcs.revision=c8f245c2c9a58083…`, `vcs.modified=true`를 보였고 그 커밋은 HEAD의 조상이 아니다(`git merge-base --is-ancestor c8f245c2c9a5 HEAD` → exit 1). 그래서 판정 도구로 쓰지 않고 `-X …/pkg/version.Commit=c50da9c2f`로 다시 빌드했다: `moai-c50da9c2f version` → `v3.2.0-tree   c50da9c2f   built unknown`. 설치본은 `v3.2.0-rc.23`(`d194083fb`)로 HEAD보다 오래됐다 |
| E17 | 리더 공지 테스트의 GREEN 기준선 (0.2.0) | VERIFIED | 아래 "추가 측정"의 명령: `--- PASS` 63줄, `--- FAIL` 0줄, `=== RUN` 63줄, `ok  github.com/modu-ai/moai-adk/internal/hook  6.364s`, exit 0, 트리 `c50da9c2f`. 이 63개는 새 문장이 들어오기 *전* 값이다 |

### 임시 측정 (기준선 아님)

목적: 지표 M-1을 *어떻게* 셀 수 있는지, 그리고 제안서의 176이 단일 명령으로 재현되지 않는다는 점을 보인다.

```bash
cd ~/.moai/claude-profiles/moai-adk/projects && find . -name '*.jsonl' -newermt 2026-09-26 -path '*moai-adk-go*' -print0 | xargs -0 grep -h '"type":"tool_use","id":"[^"]*","name":"AskUserQuestion"' | grep -o '"timestamp":"20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]' | sort | uniq -c
```

관측 출력(이번 실행, 이 머신, 대상 파일 1020개 — `find ... | wc -l` → `1020`):

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

### 추가 측정 (0.2.0, 트리 `c50da9c2f`, 브랜치 `WT-batch-approval-gate`)

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

E16 — 폐기한 바이너리의 VCS 스탬프와 다시 빌드한 바이너리:

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

## §B Known Issues (이 SPEC에 해당하는 항목만)

- **B3 서브에이전트 경계**: 훅은 질문 도구를 호출하지 않는다. 이 트리에서 패키지 전역 grep 가드 문구는 거짓이다(E13). 리더 공지 문장(REQ-BGS-016)은 질문 도구 이름을 담지 않고, 알림 소스 4파일의 `AskUserQuestion` 토큰 수는 0을 유지한다(AC-015 — `TestNoUserInteraction`과 같은 파일 범위·주석 제외 형태).
- **B4 frontmatter**: `created:`/`updated:`/`tags:` 정본 12필드, `phase`는 릴리스 목표 라벨.
- **B6 spec-lint 제목 규약**: `### Out of Scope — <topic>` H3 + `-` 항목.
- **B8/B10 작업 트리 위생·PRESERVE**: 아래 §D 보존 목록 밖은 건드리지 않는다.
- **B13(형제 SPEC)**: 상시 로딩 파일 편집은 마일스톤 끝에 모아서 한다(캐시 지시문 3). 이 SPEC의 상시 로딩 편집은 `kanban-dispatch.md` 한 곳이다.

## §C Pre-flight (run-phase 시작 시 실행)

```bash
git branch --show-current                    # WT-batch-approval-gate
git rev-parse --short HEAD                   # 계획 마감 이후 HEAD 기록
go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/
go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/
go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/
go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/
go test -count=1 -v -run '^(TestImplementationKickoffApprovalPreservedBeforeGoal|TestTemplateNoInternalContentLeak|TestContractModeEmitterSites)$' ./internal/template/
# 리더 공지 GREEN 기준선 — 명령 전문은 "추가 측정" E17 (23개 이름, ./internal/hook/)
```

기준선이 전부 `--- PASS`여야 시작한다. 하나라도 빨가면 이 SPEC 이전의 문제이므로 blocker로 보고한다. 플랜 단계에서 측정한 기준선: 예산·미러·LOC·보존 구간(E9~E9c)과 리더 공지 63개(E17). `internal/template` 쪽 세 테스트(`TestImplementationKickoffApprovalPreservedBeforeGoal`, `TestTemplateNoInternalContentLeak`, `TestContractModeEmitterSites`)와 `TestSpecAssembly_NoNewInternalTokens`의 변경 전 기준선은 이번에도 돌리지 않았다 — Gap. run-phase에서 첫 명령으로 측정한다. `internal/hook` 패키지 전체는 분 단위 스위트이므로 레인 로컬에서는 위 이름 선택만 돌리고 전체는 CI에 맡긴다(전체를 돌려야 하면 `moai slot acquire`로 자원 임대를 먼저 잡는다 — `gitflow-lane-protocol.md` §8).

## §D Constraints

1. **Template-First, 바이트 동일 미러.** `.claude/...`와 `internal/template/templates/.claude/...`를 함께 편집하고 `make build`로 임베드한다. 편집 대상 파일 중 `kanban-dispatch.md`(`:177`)와 `moai.md`(`:215,245,253,283-284`)는 라이브↔미러가 이미 다르다 — 그 줄들은 그대로 둔다.
2. **언어.** 새 독트린 문장은 영어(지침 문서 언어 정책). SPEC 서술만 한국어.
3. **템플릿 중립성.** 미러 복사본에 카드 id(`tNNN`), SPEC ID, 날짜를 쓰지 않는다(`TestTemplateNoInternalContentLeak`). 형제 SPEC도 같은 방식으로 개념 토큰만 썼다.
4. **줄·예산 한도.** `run.md` 199/200줄, `spec-assembly.md` 597/600줄 — 새 줄을 추가하지 않고 *기존 줄을 바꿔 쓴다*(`TestEntryRouterLOCCeiling`, `TestSubSkillLOCCeiling`). 상시 로딩 `kanban-dispatch.md`는 순증가를 1,000바이트 미만으로 하고(`rule-authoring.md`), 측정 여유는 13,373토큰(E9b). `workflows/moai.md`(라이브 284 / 미러 282줄)에는 줄 수 테스트가 걸려 있지 않다(E14) — 그래도 M4는 두 사본의 줄 수를 바꾸지 않는다(`:144`·`:240`은 한 줄을 한 줄로 바꿔 쓴다).
5. **고정 문구.** `TestImplementationKickoffApprovalPreservedBeforeGoal`: `run.md`는 "Implementation Kickoff Approval", "AskUserQuestion", `/goal`을 이 순서로 유지하고 "regardless of"·"score-independent" 문구도 유지. `TestSpecAssembly_RewrittenToCLIPath`: `spec-assembly.md`는 `[HARD] The Implementation Kickoff Approval`, `moai plan render-html`, `Fail-open`을 유지. `TestSpecAssembly_NoNewInternalTokens`: 금지 토큰 추가 금지.
6. **보존 구간.** `kanban-dispatch.md`의 "Promotion is the operator's act, always." → "The self-dispatch lane exception." 구간은 라이브·미러 바이트 동일을 유지하고, "never picks for the operator"·"queue ADMISSION stays the operator's" 금지 문구를 지킨다(`TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`). 이 구간에 `:31`의 "batch approval" 문장이 들어 있으므로 *그 줄은 편집 금지*다.
7. **계약 모드 문구.** `TestJevAmendmentLinkage`(`internal/contract/kickoff/activation_test.go:208`)가 MCP 카탈로그·workflow.yaml의 contract-mode Kickoff/Jev 예외 문자열 제거를 막는다 — 건드리지 않는다. `TestContractModeEmitterSites`(`internal/template/contract_mode_guided_test.go:695`)도 같다.
8. **Frozen 구역 보존.** `orchestration-mode-selection.md:18`의 `[ZONE:Frozen]` 조항은 편집하지 않는다. 이 SPEC이 만지는 어떤 조항도 `zone-registry.md`에 항목이 없으므로(E9e) `moai constitution amend`는 필요 없다.
9. **명령 형식.** 워크트리 가드 때문에 검증 명령은 단순 명령·리터럴 경로로 쓴다(`worktree-integration.md` § Refused Commands).
10. **도구 출처(VCI §2.2).** 이 플랜의 `go test`는 트리에서 직접 빌드한 결과다. `moai` 설치본(`v3.2.0-rc.23`, 커밋 `d194083fb`)은 트리 HEAD보다 오래됐으므로 `moai spec lint`는 트리에서 빌드한 바이너리로 돌리고 판정 빌드 커밋을 함께 적는다. 0.2.0: 이전 실행의 스크래치 바이너리는 VCS 스탬프가 HEAD와 달라(E16) 쓰지 않았고, `-X pkg/version.Commit=c50da9c2f`로 다시 빌드한 `moai-c50da9c2f`가 이번 판정 도구다.
11. **리더 공지 Go 제약(0.2.0).** 새 문장은 기존 테스트가 이미 고정한 다음을 깨지 않는다. (a) 블록 구조: 칸반 공지는 정확히 5개 블록이고 3번째 블록은 명령 3줄뿐이다(`TestKanbanLeadNoticeBlockLayout`) → 문장은 *기존 블록 안의 한 줄*로 합치고 새 블록을 만들지 않는다. (b) 역할 용어: 팩토리·칸반 리더 공지의 en에는 단어 `lead`(대소문자 무시, 단어 경계)가 없고, ko에는 `리드`가 없으며, `worker-<n>`·`-f worker`·`-f agent`가 없고, 두 공지는 로케일별 leader·lane 용어를 계속 담는다(`TestRoleNamingM3NoticesCarryLeaderLaneTerms`). (c) todo 비활성 시 칸반 공지에 `moai todo`가 없다(`TestSessionStartKanbanRespectsTodoDisabled`) → 문장에 `moai todo`를 쓰지 않는다. (d) SPEC 미설정 시 `SPEC-` 부분 문자열이 없고(`TestKanbanLeadNoticeOmitsSPECWhenUnset`), 대소문자 무시 `epic`이 없다(`TestKanbanLeadNoticeBacklogSummaryCountsQueuedOnly`). (e) 팩토리 공지에 `moai todo list`·`.moai/state/kanban/backlog.json`·`poll the backlog queue`·`No model override`·`ANTHROPIC_DEFAULT_*_MODEL`이 없다(`TestFactoryLeadNoticeCarriesDispatchDiscipline`). (f) 새 필드는 형식 문자열이 아니다(`%s`·`%d` 동사 없음). (g) 질문 도구 이름 없음(B3). (h) 카드 id·SPEC ID·날짜 없음, 네 로케일 동시 갱신(`TestKanbanLocalesCoverEveryField`, `TestKanbanNoticePreservesProtocolTokensInEveryLocale`). (i) 포인터 토큰 `.claude/rules/moai/workflow/auto-semantics.md` §9.2는 네 로케일에서 번역하지 않는 주소다(기존 명령 토큰과 같은 취급).
12. **`workflows/moai.md` 편집 방식(0.2.0).** 라이브와 미러는 `:215`·`:245`·`:253`·`:283-284`에서 이미 다르다(E15). M4는 각 사본에 같은 `old_string` 치환을 *따로* 적용하고, 한쪽을 다른 쪽으로 복사하거나 재동기화하지 않는다. 편집 뒤 `diff`의 헌크 헤더가 편집 전과 같은 넷이어야 한다(AC-014).

## §E Self-Verification (run-phase 완료 보고에 담을 항목)

`manager-develop-prompt-template.md` §E의 E1–E8을 따른다. 이 SPEC에 특화된 항목:
- E1 AC 이진 PASS/FAIL 행렬(`acceptance.md` 기준, 명령 + 원문 출력).
- RED 원문: 가드 테스트가 정본 절 부재 상태에서 실패하는 출력(E8, TDD).
- 변이 증거: 앵커를 뺀 변이 본문 각각의 거부 출력(AC-007).
- 상시 로딩 `wc -c` 전/후와 `TestAlwaysLoadedTokenBudget` 출력.
- 미러 동일성(`TestRuleTemplateMirrorDrift`)과 LOC 한도 출력.
- 리더 공지(M5): AC-015의 RED 원문(파일 부재·포인터 토큰 0), GREEN 출력(하위 테스트 `--- PASS` 줄 수), 변이 거부 출력, 알림 소스 4파일의 토큰 수 0, 기존 고정 테스트 63개(E17)의 재실행 결과, `git diff --numstat`(AC-009 읽기 단계).
- `workflows/moai.md`(M4): 편집 뒤 `diff` 헌크 헤더 넷과 `wc -l`(284/282) 출력(AC-014).

## §F Milestones

우선순위 표기만 쓴다. 번호는 *실행 순서*이고, 변경 가능성이 큰 순서(검토 우선순위)는 아래 표다.

| 검토 우선순위 | 마일스톤 | 이유 |
|---|---|---|
| 1 (가장 바뀔 가능성 큼) | M2 정본 절 | 사용자 대면 흐름(요약 보고·단일 질문·pull-out)과 반대 증거 형식 — 운영자 판정(Q1·Q2·Q3·Q7)에 직접 걸린다 |
| 2 | M5 리더 공지 문장 | 세션 시작 때 리더와 운영자가 읽는 문구(네 로케일). 명칭(Q10)이나 정본 위치가 바뀌면 같이 바뀐다. 실행은 M2 뒤 |
| 3 | M4 낡은 표현 정합 | 운영자가 범위 안으로 판정(Q4, 2026-10-02). 문구가 §9.1·§9.2에 종속된다. 실행은 M2 뒤 |
| 4 | M3 포인터 | 정본 위치(Q6)가 정해지면 기계적 |
| 5 (가장 기계적) | M0·M1 | 기준선 기록과 테스트 — 결정에 덜 민감 |

### M0 — 기준선 기록 (Priority High, 선행 필수)

- 범위를 *먼저* 정의한다: 어느 전사본 디렉터리(프로젝트 디렉터리 406개 중 어느 접두사), 어느 기간, 게이트 질문 분류 방식(질문 문구가 §9 게이트 행 이름을 포함하는가), 중복 처리.
- 산출물 `.moai/reports/t1344/baseline-gate-rounds.md`: 명령 원문 + 관측 출력 + 분류 방식 + 한계. **이 파일만 담은 단독 커밋**으로 남긴다(VCI §2.3 — 기준선과 변경이 같은 커밋이면 순서를 커밋 그래프로 증명할 수 없다).
- 종료 조건: AC-008의 명령이 통과한다(파일 존재 + 기준선 커밋이 정본 절 커밋의 조상).

### M1 — 가드 테스트, RED 먼저 (Priority High)

- 새 파일 `internal/template/batch_gate_summary_doctrine_test.go`(테스트 전용). 정본 절을 템플릿 복사본에서 잘라 내 앵커 표(아래 불변식)를 점검하는 검사 함수를 두고, 표 구동 하위 테스트로 (a) 실제 절은 위반 0, (b) 앵커를 하나씩 뺀 변이 본문은 *해당* 불변식 위반을 보고함을 확인한다.
- RED: 정본 절이 아직 없으므로 첫 실행은 실패해야 한다(원문 출력을 progress §E.2에 보존).
- 불변식 앵커 후보(문구는 M2에서 확정, 표에서 파생): 같은 게이트 행 키, 행 필드 목록(`counter_refs=` 포함), 질문 하나·승인 범위, pull-out, `none`+`searched=`, 행별 결정 기록, keep-set 세 범주, 권한 예약 목록, 차단 조건(FAIL·INCONCLUSIVE·부재·해시 변경·열린 차단), contract 모드 제외, "176" 결과 수치 부재.

### M2 — 정본 절 `auto-semantics.md` §9.2 (Priority High)

- 위치: §9.1 바로 뒤. 제목 `### 9.2 The batch gate summary`. 분량 목표는 3KB 안팎(상한 없음, 단순 유지).
- 내용 개요(최종 문구는 run-phase가 작성):
  1. 적용 대상: 운영자 형태로 대기 중인 행이 같은 게이트 행에서 둘 이상.
  2. 묶는 주체(리더)·키(같은 게이트 행)·스냅숏·준비된 카드 붙잡기 금지·동일 판정 주체 1회 질문.
  3. 보고서 서식: 헤더 줄 + 행 표(카드 | SPEC | 판정(점수/기준/여유) | 해시 | 기록 | counter_refs). 반대 증거 행 우선.
  4. 질문 하나: 승인 범위 문장("approval covers the listed approvable rows only"), pull-out, 권고 라벨은 현행 `recommendation_mode` 규약 그대로.
  5. `counter_refs=<paths+ids>` / `counter_refs=none searched=<what>`; 원천 목록.
  6. 승인된 행마다 §10 기록 + `counter_refs=` + `ladder_path`에 배치 표시. 기록 없는 행은 미승인.
  7. 요약 밖: keep-set, 권한 예약, 차단 행, contract 모드.
  8. 하니스 주: 질문 채널이 없으면 같은 요약이 blocker 보고서로 나간다.
  9. 구별 문장: `--auto` 배치 권한 부여와 별개.
- §10에 한 문장 추가: 요약으로 승인된 결정은 `counter_refs=`를 싣는다. §9.1 끝에 §9.2 포인터 한 줄.
- 템플릿 먼저 → 라이브 동일 내용 → `make build`. 미러 동일성 확인.

### M3 — 포인터 (Priority Medium)

- `kanban-dispatch.md` Boundaries "No gate bypass." 불릿 안에서 문장을 다듬어 §9.2를 가리킨다(상시 로딩, 순증가 1,000바이트 미만, 보존 구간 `:29-33` 밖). 미러 동시 편집, `:177` 사전 차이는 유지.
- `run.md:137`의 기존 문장에서 "§9.1" 인용을 "§9.1–9.2"로 바꾸는 식의 *줄 수 불변* 편집만. 고정 문구·순서 유지.
- 검증: 예산·LOC·AutoRank·Kickoff 보존 테스트 전부.

### M4 — 낡은 표현 정합 (Priority Medium, 무조건 — Q4 범위 안, 2026-10-02; **M2 뒤에 실행**)

- 선행: M2 커밋이 브랜치에 있어야 한다(다시 쓴 문구가 §9.1·§9.2를 가리킨다).
- `spec-assembly.md:202-208`(7줄)을 같은 줄 수로 다시 쓴다. 보존: 첫 줄의 `[HARD] The Implementation Kickoff Approval` 시작 문구(`TestSpecAssembly_RewrittenToCLIPath`가 템플릿 사본에서 고정), `moai plan render-html` 단계(`:197`), `Fail-open` 문단(`:210` 이후), HTML 보고서가 게이트를 *대체하지 않고 보강만 한다*는 취지와 `(권장)`/pull 라벨 설명. 정정: "stays MANDATORY and score-independent"와 "A plan-auditor PASS or a high skip-eligible score does NOT substitute for the gate."를 `auto-semantics.md` §9.1(기본은 자율 전환, keep-set은 운영자 답)에 맞게 쓰고 §9.2를 가리킨다. 597줄 불변(한도 600, `TestSubSkillLOCCeiling`은 라이브 사본만 읽는다). 새 내부 토큰(SPEC ID·REQ/AC 토큰) 추가 금지(`TestSpecAssembly_NoNewInternalTokens`).
- `workflows/moai.md:144`(파이프라인 게이트 2)와 `:240`(Step 11.3)을 각각 *한 줄을 한 줄로* 바꿔 쓴다. 정정 대상: `:144`의 "Score-independent: a plan-auditor PASS or skip-eligible score never bypasses it"와 `:240`의 "score-independent". 같은 줄의 보존 서술(게이트를 한 파이프라인 진입당 한 번 제시, 병합 라운드가 질문 둘을 한 호출에 싣는 문장, 파생 완료 조건이 run 진입을 허락하지 않는다는 문장)은 유지한다. 두 줄의 레이블 `HUMAN GATE`는 keep-set 형태에서만 문자 그대로 참이다 — 함께 정정할지는 run-phase가 §9.1 문구와 대조해 정하고 이유를 progress §E.2에 남긴다(정정 범위를 임의로 넓히지 않는다). 뒤따르는 contract-mode 블록(`contract-pipeline-gates`, `contract-merged-round`)은 건드리지 않는다(`contract_mode_blocks_test.go:54`가 고정). 적용은 두 사본에 *각각* 같은 `old_string`으로(제약 12). 줄 수 284/282 불변, 편집 뒤 `diff` 헌크 헤더가 `215c215`·`245c245`·`253d252`·`283,284c282` 넷 그대로(AC-014).
- 읽고 판정(편집 없음, 0.2.0에서 읽음): `.claude/skills/moai/SKILL.md:166`("Progression mode … the gate stays mandatory in both modes")과 `:372`("A derived completion condition NEVER authorizes autonomous run-phase entry — Implementation Kickoff Approval remains mandatory at the plan→run boundary")는 *게이트가 충족돼야 한다*는 뜻으로 읽히며 §9.1(게이트는 자율 형태 또는 운영자 형태로 충족)과 충돌하지 않는다고 판정했다. 판정이지 기계 점검이 아니다. `:170`은 contract-mode 블록 `contract-signing-router`의 첫 줄이다. SKILL.md는 Q4가 이름 붙인 파일이 아니고 두 사본이 이미 다르므로(`diff -q` exit 1) 편집 대상에서 뺀다.
- 검증: `TestSpecAssembly_RewrittenToCLIPath`, `TestSpecAssembly_NoNewInternalTokens`, `TestSubSkillLOCCeiling`, `TestImplementationKickoffApprovalPreservedBeforeGoal`(run.md만 읽으므로 영향이 없음을 확인), `TestContractModeEmitterSites`와 contract-mode 블록 테스트, 새 문구에 `manager-tdd`·`manager-ddd` 금지(고아 참조 스캔), AC-014의 grep·`diff`·`wc -l`.

### M5 — 리더 공지 문장 (Priority Medium, 무조건 — Q5 범위 안, 2026-10-02; **M2 뒤에 실행**)

- 선행: M2 커밋이 브랜치에 있어야 한다(포인터가 가리킬 정본 절이 먼저 있어야 한다). 훅 Go 파일은 세션 로딩 파일이 아니므로 `cache-aware-execution.md` 지시문 3의 "상시 로딩 파일 편집은 끝으로" 제약은 이 마일스톤에 걸리지 않는다. M3·M4와의 순서는 자유이고 M2에만 의존한다.
- **두 리더 공지에 모두 싣는다**(근거 E12b). 레인·동반 공지는 받지 않는다.
- 제품 Go 대상, 정확히 넷:
  - `internal/hook/session_start_kanban_i18n.go`: 구조체(`:37-51`)에 문장 필드 하나(후보 이름 `leaderGateSummary`), 네 로케일 값(en `:60-83`, ko `:84-105`, ja `:106-127`, zh `:128-149`).
  - `internal/hook/session_start_kanban.go`: `kanbanLeaderNotice`(`:116-197`)의 문맥 블록 (e)(`:177-194`)에서 설정 줄 뒤에 그 필드를 한 줄로 합친다. 새 블록 없음.
  - `internal/hook/session_start_factory_i18n.go`: 구조체(`:35-67`)에 같은 필드, 네 로케일 값(en `:74-124`, ko `:125-169`, ja `:170-214`, zh `:215-258`).
  - `internal/hook/session_start_factory.go`: `factoryLeaderNotice`(`:185-249`)의 규율 블록 (e)(`:222`)의 `strings.Join` 목록에 그 필드를 더한다.
- 테스트 Go 대상: 기존 `session_start_kanban_i18n_test.go`의 `TestKanbanLocalesCoverEveryField` 필드 표(`:27-41`)에 새 필드를 덧붙인다(약화 없음). 새 `internal/hook/session_start_leader_gate_notice_test.go`: 두 리더 공지 × 네 로케일을 렌더해 검사 함수에 넣는다(포인터 토큰 존재; 금지 토큰 부재 — 질문 도구 이름, `SPEC-`, 카드 id 꼴, ISO 날짜, `moai todo`, 단어 `lead`, `리드`, `epic`). 레인·동반 공지에 포인터가 없음을 확인한다. 변이 본문(로케일 하나 누락, 질문 도구 이름 포함, 포인터 누락, 카드 id 포함, 한쪽 리더 공지만 보유)이 각각 거부됨을 보인다. 알림 소스 4파일의 주석 제외 토큰 스캔(`TestNoUserInteraction` 형태).
- 문장의 모양(최종 문구는 run-phase가 쓴다): 한 문장. 무엇 — 같은 운영자 형태 게이트에서 대기 중인 카드가 둘 이상이면 카드별 질문이 아니라 배치 게이트 요약 하나로 제시한다. 어디 — `.claude/rules/moai/workflow/auto-semantics.md` §9.2. 서식·제외·반대 증거 의무는 *적지 않는다*: 정본을 가리킬 뿐이다(REQ-BGS-016 "restate 금지").
- 순서: RED 먼저. 테스트 파일을 먼저 만들어 실패(포인터 부재)를 관측하고 원문을 progress §E.2에 남긴 뒤, 필드와 값을 더해 GREEN으로 뒤집는다. 제품 Go와 테스트는 한 커밋이어도 된다(VCI §2.3은 기준선에만 걸린다).
- 비용 공시: 리더 세션의 `startup`에서만 문장 한 줄이 늘고 비리더 세션은 0바이트다. 렌더 길이의 전/후는 run-phase에서 공시한다(플랜 단계에서는 측정하지 않았다 — UNVERIFIED).
- 검증: E17 명령 전체 PASS(63개 + 새 하위 테스트), 새 테스트 PASS(하위 테스트 `--- PASS` 12개 이상), 4파일 토큰 수 0, 변경 파일에 `gofmt -l`·`go vet`·CI 버전 golangci-lint.

## §G Anti-Patterns

- 요약을 *승인 수단*으로 오해해 행별 증거 판독을 건너뛰는 것 — 요약은 제시 형식이다. 행마다 §9.1 증거(판정 PASS·해시 불변·차단 없음)를 리더가 읽는다.
- 대기 카드를 모으려고 준비된 카드를 지연시키는 것(REQ-BGS-002).
- keep-set·권한 예약 행을 "일괄이 편하니까" 넣는 것(REQ-BGS-010).
- `none`만 적고 검색 내용을 빼는 것(REQ-BGS-008).
- 기준선과 변경을 한 커밋에 담는 것(VCI §2.3).
- 새 줄을 늘려 `run.md`·`spec-assembly.md` 한도를 넘기는 것.
- 리더 공지 문장에 정본의 서식·제외·반대 증거 규칙을 옮겨 적는 것(REQ-BGS-016 "restate 금지") — 두 벌이 되는 순간 갈라진다.
- 리더 공지 문장을 레인·동반 공지나 레인 규칙 필드에 넣는 것, 또는 한쪽 리더 공지에만 넣는 것(E12b).
- 공지에 새 블록을 만들어 `TestKanbanLeadNoticeBlockLayout`의 5블록 구조를 깨는 것.
- `moai.md`의 라이브와 미러를 맞추려고 한쪽을 복사하거나 재동기화해 사전 차이(`:215`·`:245`·`:253`·`:283-284`)를 건드리는 것.
- 패키지 전역 `internal/hook` grep을 통과 기준으로 쓰는 것 — 이 트리에서 그 가드 문구는 거짓이다(E13).

## §H Cross-References

- `.claude/rules/moai/workflow/auto-semantics.md` §7, §9, §9.1, §10, §11 — 정본 기반
- `.claude/rules/moai/workflow/kanban-dispatch.md` — 리더 책임, 보존 구간, Boundaries
- `.claude/rules/moai/core/askuser-protocol.md` § Report-Before-Ask Gate, § Recommendation mode
- `.claude/rules/moai/core/verification-claim-integrity.md` §1, §2.3
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, §2.1, §3(교차 계층 개정 스윕 — 0.2.0 개정이 따랐다)
- `internal/hook/session_start_{kanban,factory}{,_i18n}.go` — 리더 공지 조립과 로케일 표(M5); `internal/hook/CLAUDE.md` § Conventions(서브에이전트 경계 문구, E13)
- `.claude/rules/moai/development/rule-authoring.md` — 비호출 세션 비용 공시 형식(리더 공지 문장의 R-3에 준용)
- 형제: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md` §A.7

## 표면 목록 (D6 형제 점검 결과)

점검 방법: `grep -rIl -i "kickoff" .claude/skills .claude/rules .claude/agents .moai/docs`(적중 39개 파일)와, 대상 파일 줄 단위 재독.

| 분류 | 파일 | 처분 |
|---|---|---|
| **편집(항상)** | `.claude/rules/moai/workflow/auto-semantics.md` (+미러) | §9.2 신설, §9.1 끝·§10 한 줄 |
| **편집(항상)** | `.claude/rules/moai/workflow/kanban-dispatch.md` (+미러) | Boundaries 불릿 포인터(상시 로딩, 보존 구간 밖) |
| **편집(항상)** | `.claude/skills/moai/workflows/run.md` (+미러) | `:137` 줄 수 불변 포인터 |
| **편집(항상)** | `internal/template/batch_gate_summary_doctrine_test.go` | 신규 테스트 전용 파일 |
| **편집(항상, Q4 범위 안)** | `.claude/skills/moai/workflows/plan/spec-assembly.md:202-208` (+미러) | 낡은 표현 정합(M4) |
| **편집(항상, Q4 범위 안)** | `.claude/skills/moai/workflows/moai.md:144,240` (+미러, 사전 차이 줄 미접촉) | 낡은 표현 정합(M4) |
| **편집(항상, Q5 범위 안)** | `internal/hook/session_start_kanban_i18n.go`, `session_start_kanban.go`, `session_start_factory_i18n.go`, `session_start_factory.go` | 리더 공지 문장 하나, 두 리더 공지 × 4 로케일(M5) |
| **편집(항상, Q5 범위 안)** | `internal/hook/session_start_kanban_i18n_test.go`(덧붙임), `internal/hook/session_start_leader_gate_notice_test.go`(신규) | 필드 빈값 표 + 문장 규칙 점검(M5) |
| **읽고 변경 없음(0.2.0에서 읽음)** | `.claude/skills/moai/SKILL.md:166,372` | §9.1과 충돌하지 않는다고 판정(M4 참조); 미러가 이미 달라 손대지 않는다 |
| **읽고 변경 없음(이번에 줄 단위로 읽음)** | `.claude/skills/moai/workflows/gtd.md`(Kickoff 언급 없음, `:334`는 큐 소비 용어), `kanban-dispatch-detail.md`·`kanban-dispatch-mechanics.md`(Kickoff 언급 없음 — 위 grep 목록에 없음), `contract-autonomy.md`(계약 모드는 서명이 게이트 → REQ-BGS-012), `.claude/skills/moai/workflows/{plan,factory,goal}.md`·`run/{phase-execution,mode-orchestration,task-decomposition}.md`(Kickoff를 *선행 조건*으로 인용), `moai-kanban-foreman/SKILL.md:73`(포어맨이 답하지 않는 판단 목록 — 큐 진입 보호), `cadence-bridge.md`(형제가 개정), `plan-auditor.md`·`manager-develop.md`(선행 조건 인용) | 변경 없음 |
| **보존(편집 금지)** | `orchestration-mode-selection.md`(`:18` Frozen), `askuser-protocol.md`(상시 로딩, 이미 Report-Before-Ask 예외에 Kickoff 언급) | 그대로 |
| **grep 적중, 개별 정독 안 함** | `askuser-protocol.md` 외 `core/moai-mcp-tools-catalogue.md`, `development/coding-standards.md`, `workflow/{archived-agent-rejection,dynamic-workflows,goal-directive,goal-directive-detail,session-handoff,session-handoff-examples,spec-workflow}.md`, `workflows/{design,e2e,harness-build-entry,project/doc-generation}.md`, `.claude/agents/{harness/workflow-specialist,moai/manager-design}.md`, `hns-moaiadk-patterns/SKILL.md`, `moai-workflow-spec/SKILL.md`, `.moai/docs/session-handoff-appendix.md` | **UNVERIFIED 변경 불필요** — M3에서 재독하고 이탈은 progress §E.2에 기록 |
| **로컬 전용, 비배포** | `.moai/docs/jev-local-operations.md`, `.claude/rules/local/*` | 편집 없음(3등급 독트린의 집은 그대로) |

## Go 작업 판단 (D5/E7)

- **제품 Go: 리더 공지 문자열에 한정, 파일 넷**(0.2.0, Q5 범위 안). `internal/hook/session_start_kanban.go`·`session_start_kanban_i18n.go`·`session_start_factory.go`·`session_start_factory_i18n.go` — 로케일 표에 문장 필드 하나와 네 로케일 값, 조립부에서 기존 블록 안에 한 줄 합치기뿐이다(M5). 새 블록·분기·환경 변수·호출은 없다. 두 리더 공지 모두에 싣고(E12b) 레인·동반 공지에는 싣지 않는다.
- **그 밖의 제품 Go: 없음.** 이유: (1) 다중 카드 적용은 `moai factory decide <card>...`가 이미 한다 — 요약 승인 뒤 승인 집합을 인자로 넘기면 되고, 빼낸 행은 인자에서 뺀다. 한 호출은 gate/choice 하나만 받으므로 같은 게이트 행끼리만 묶는 REQ-BGS-001과 맞는다. (2) 결정 기록은 줄 기록이라 CLI 없이 쓴다(형제 SPEC과 같은 선). (3) `factory_decide` MCP는 카드 1장 단위이고 레인 세션에 거부되지만 리더는 레인이 아니므로 카드별 호출이 가능하다. 배치 기구 자체의 Go 강제(Q8)는 열린 채 범위 밖이다.
- **잔여 위험**: 강제 장치가 없어 오케스트레이터 규율에 기댄다(G-3). 자동 기록·검증이 필요한지는 Q8. 알림 문장은 리더에게 규칙의 존재를 상기시킬 뿐 준수를 강제하지 않는다.
- **테스트 전용 Go 3파일**(템플릿 가드 1, 훅 신규 1, 훅 기존 덧붙임 1)은 제품 동작이 아니라 독트린 불변식과 알림 문장 규칙의 반증 가능 점검이다(B.7). 허용 파일 밖의 `.go` 변경은 AC-009가 막는다.

## 요약 승인의 적용 방식 (문서 수준)

- factory 실행: 승인된 행 집합 → `moai factory decide <card>... --gate kickoff --choice approve`(카드별 호출 또는 다중 인자) → 카드마다 §10 줄 기록. 빼낸 행은 개별 질문.
- 단독 오케스트레이션: 승인된 SPEC마다 run-phase 진입을 허락하고 같은 기록을 남긴다.
- 결정 기록은 디스크 결정 보드(§11, 홈 표면)에 줄로 쓴다.
