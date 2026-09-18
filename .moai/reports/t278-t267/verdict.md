# verdict — card t278+t267 잔여 체리픽

두 브랜치(`WT-ci-flake-series` / `WT-taskstop-name-reclaim`)의 미병합 커밋 중 develop 에
아직 반영되지 않은 내용만 골라내라는 배차에 대한 판정서.

- 측정 시점: 2026-09-18
- 판정 트리: `.claude/worktrees/t278-t267` (base `d9951c50e` = `origin/develop`)
- 작업 브랜치: `WT-flake-report-ledger`

## 결론 한 줄

**양쪽 다 체리픽할 것이 없다.** t267 은 전부 develop 에 반영돼 있고, t278 은 SPEC·Go 산출물이
전부 반영돼 있으며 남은 `.moai/reports/t278/` 9 파일은 **운영자가 2026-09-14 에 일부러 추적
해제한 대상**이라 다시 올리는 것이 회귀다. 두 브랜치 마감을 요청한다.

---

## Claim

1. **t267 (`WT-taskstop-name-reclaim`) — 체리픽할 것 없음. 전부 중복.** 미병합 14 커밋 중
   순수 작업 13건의 산출물은 모두 develop 에 반영돼 있다.
2. **t278 (`WT-ci-flake-series`) — 체리픽할 것 없음.** 미병합 13 커밋 중 순수 작업 12건의
   산출물 가운데 `.moai/specs/SPEC-CI-FLAKE-SERIES-001/**` 4종과 Go 변경은 전부 develop 에
   반영돼 있고, 유일한 미반영분인 `.moai/reports/t278/` 9 파일은 **PR #1666(`379b310a6`)으로
   main 에 착지했다가 커밋 `ba60eb6d5`(2026-09-14, "chore(repo): untrack .moai/reports —
   local-only artifacts")로 운영자 지시에 따라 추적 해제된 것**이다. 부재가 누락이 아니라
   결정이다.
3. **두 브랜치의 미병합 커밋에는 각각 정확히 1개의 머지 커밋이 들어 있다** — 부모 수 2로
   확인. 통째 병합 금지 판정의 근거가 실측으로 성립한다.
4. **`internal/sessionmsg/stoprule_test.go` 의 브랜치-develop 차이는 t278 의 미반영분이
   아니라 develop 이 앞선 것이다.** 옮기면 카드 t253 을 되돌린다.
5. **이 판정 과정에서 나 자신이 한 번 틀렸고, 그 사실을 기록한다.** 9 파일을 "끊긴 인용 경로를
   되살리는 전치"로 판단해 실제로 스테이징했다가, 추적 해제 이력을 확인하고 되돌렸다. 기전:
   `git check-ignore` 를 **스테이징 이후에** 물었고, 그 명령은 인덱스에 있는 경로를 보고하지
   않는다. 판정 근거로 쓴 "무시되지 않는다"는 출력은 실은 "이미 인덱스에 있다"였다.

## Evidence

### E1 — 미병합 커밋 전수와 부모 수 (머지 판별은 제목이 아니라 부모 수)

```
$ git log --format='%h | parents=%p | %s' develop..WT-ci-flake-series
2aec09c27 | parents=03cf93cc2 968ed2acb | Merge origin/main into WT-ci-flake-series — resolve sync PR conflicts (t278)
03cf93cc2 | parents=2cbf3078c | docs(SPEC-CI-FLAKE-SERIES-001): post-merge ledger first accrual N=2 (t278)
... (부모 1개인 순수 작업 커밋 11건 생략, d1289c5db 까지)
$ git rev-list --count develop..WT-ci-flake-series
13

$ git log --format='%h | parents=%p | %s' develop..WT-taskstop-name-reclaim
052830c0a | parents=c572fd8a9 da791eb0a | chore(t267): merge origin/main, resolve CHANGELOG [Unreleased] keeping both sides
c572fd8a9 | parents=90f7ad3e3 | docs(SPEC-TEAMMATE-REVIVAL-GUARD-001): §E.4 cites §E.3 run-chain closure note instead of restating (t267)
... (부모 1개인 순수 작업 커밋 12건 생략, 6b7adb026 까지)
$ git rev-list --count develop..WT-taskstop-name-reclaim
14
```

부모 2개 = 머지: t278 은 `2aec09c27` 1건, t267 은 `052830c0a` 1건. 나머지는 전부 부모 1개.

### E2 — 중복 판정은 커밋 id 가 아니라 내용으로

`git cherry -v develop <branch>` 는 양쪽 모두 25 커밋 전부를 `+`(상류에 등가 patch 없음)로
보고한다. 즉 **patch-id 기준으로는 아무것도 중복이 아니다.** 하지만 patch-id 는 중간 단계
diff 를 비교하므로, 최종 내용이 다른 경로(PR squash)로 착지한 경우를 잡지 못한다. 실제로
t278 본체는 PR #1666 의 squash 커밋 `379b310a6` 으로 착지했다. 그래서 경로별 최종 내용으로
다시 판정했다.

```
$ for f in <t278 touched paths>; do git diff --numstat develop WT-ci-flake-series -- "$f"; done
.moai/specs/SPEC-CI-FLAKE-SERIES-001/spec.md                 IDENTICAL
.moai/specs/SPEC-CI-FLAKE-SERIES-001/plan.md                 IDENTICAL
.moai/specs/SPEC-CI-FLAKE-SERIES-001/acceptance.md           IDENTICAL
.moai/specs/SPEC-CI-FLAKE-SERIES-001/progress.md             IDENTICAL
CHANGELOG.md                                                 1+/438-
internal/hook/config_change_test.go                          IDENTICAL
internal/sessionmsg/stoprule_test.go                         1+/6-
internal/sessionmsg/store_test.go                            IDENTICAL
internal/timing/paired_asym_test.go                          IDENTICAL
internal/timing/timing.go                                    IDENTICAL

$ git diff --stat develop WT-ci-flake-series -- .moai/reports/t278
 9 files changed, 685 insertions(+)      ← develop 에 전무 (사유는 E3)
```

```
$ for f in <t267 touched paths, 23개>; do git diff --numstat develop WT-taskstop-name-reclaim -- "$f"; done
.moai/docs/agent-stop-audit-correlation.md                       IDENTICAL
.moai/specs/SPEC-TEAMMATE-REVIVAL-GUARD-001/ (6종 전부)           IDENTICAL
internal/cli/hook.go                                             IDENTICAL
internal/config/workflow_agent_stop_guard_test.go                IDENTICAL
internal/hook/agent_model_guard.go                               IDENTICAL
internal/hook/agent_stop_guard.go                                IDENTICAL
internal/hook/agent_stop_guard_test.go                           IDENTICAL
.claude/rules/.../cross-session-messaging.md                     4+/21-
.claude/settings.json                                            170+/210-
.moai/config/sections/workflow.yaml                              3+/21-
CHANGELOG.md                                                     1+/440-
internal/config/cache.go                                         2+/4-
internal/config/defaults.go                                      17+/200-
internal/config/types.go                                         35+/257-
internal/hook/pre_tool.go                                        15+/136-
internal/hook/session_end.go                                     49+/222-
internal/template/templates/.claude/.../cross-session-messaging.md 4+/21-
internal/template/templates/.claude/settings.json.tmpl           10+/26-
```

차이가 남은 8 파일은 전부 `삽입 ≪ 삭제` 형태 — develop 이 앞선 것이다. t267 고유 산출물이
develop 에 실제로 살아 있는지는 마커로 직접 확인했다.

```
$ git show develop:internal/config/cache.go | grep -n "configCacheSchemaVersion ="
27:const configCacheSchemaVersion = 4          ← 브랜치는 2. develop 이 더 진행됨
$ git show develop:internal/config/types.go    | grep -c "AgentStopGuard"      → 4
$ git show develop:internal/config/defaults.go | grep -c "AgentStopGuard"      → 1
$ git show develop:.moai/config/sections/workflow.yaml | grep -A1 agent_stop_guard
181:    agent_stop_guard:
182-        enabled: true
$ git show develop:.claude/rules/moai/workflow/cross-session-messaging.md | grep -c STOPPED_TEAMMATE_VIOLATION  → 2
$ git show develop:internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md | grep -c STOPPED_TEAMMATE_VIOLATION  → 2
$ git grep -n "ClearAgentStops" develop -- internal/
develop:internal/hook/agent_stop_guard.go:474  (선언)
develop:internal/hook/agent_stop_guard.go:479  (함수)
develop:internal/hook/agent_stop_guard_test.go:832,837,841,842
develop:internal/hook/session_end.go:111       ← 세션 종료 시 정리 호출, 착지 확인
```

`session_end.go` 는 한 번 오판했다. 처음 `grep -i "agent-stops\|StopRegistry\|ClearSessionStops"`
로 0행을 얻어 "미반영"으로 읽었는데, 실제 심볼은 `ClearAgentStops` 였다. 트리 전체 `git grep`
으로 다시 재서 develop:111 에 살아 있음을 확인했다. **0행은 부재가 아니라 내 패턴이 틀렸다는
뜻일 수 있다.**

### E3 — t278 보고서 9 파일은 "누락"이 아니라 "추적 해제 결정"

```
$ git log --oneline --all -- .moai/reports/t278/forensics.md
ba60eb6d5 chore(repo): untrack .moai/reports — local-only artifacts
379b310a6 fix: CI flake 3종 계열 수리 — poller TOCTOU·AND-gate·p95 (t278) (#1666)
0239bddd4 docs(SPEC-CI-FLAKE-SERIES-001): M1 investigation artifacts (t278)

$ git show --stat --format='' ba60eb6d5 -- .moai/reports/t278
 .moai/reports/t278/forensics.md                 |  80 ------
 .moai/reports/t278/plan-audit-iter1.md          |  87 ------
 .moai/reports/t278/pr-body.md                   | 107 ------
 .moai/reports/t278/refetch-jobs.sh              |  15 ---
 .moai/reports/t278/reproduction-rate.md         |  93 ------
 .moai/reports/t278/series-analysis.md           |  96 ------
 .moai/reports/t278/sweep-attempts.sh            |  58 ---
 .moai/reports/t278/sync-audit.md                | 104 ------
 .moai/reports/t278/timing-statistic-decision.md |  45 ---
 9 files changed, 685 deletions(-)

$ sed -n '226,232p' .gitignore
# Card/audit reports are local-only artifacts (operator directive 2026-09-14):
...
.moai/reports/*
!.moai/reports/plan-audit/
.moai/reports/plan-audit/*
!.moai/reports/plan-audit/.gitkeep
```

세 이력이 한 줄로 읽힌다: 카드가 보고서를 만들고(`0239bddd4`), PR #1666 이 main 에
착지시키고(`379b310a6`), 운영자가 2026-09-14 에 `.moai/reports` 전체를 추적 해제하며 같은
9 파일을 지웠다(`ba60eb6d5`). develop 은 main 을 포함하므로 지금 부재한 것이 정상이다.
다시 올리면 `git add -f` 로 ignore 를 뚫어야 하고, 그것이 곧 운영자 결정의 회귀다.

### E4 — stoprule_test.go 차이의 방향 (옮기면 회귀)

```
$ git diff develop WT-ci-flake-series -- internal/sessionmsg/stoprule_test.go
-	// total sits under config.DefaultSessionMsgMaxPending (64): ... (card t253)
-	const total = 60
+	const total = 100

$ git log -1 --format='%h %ad %s' --date=short develop -- internal/sessionmsg/stoprule_test.go
c5374c4ec 2026-09-02 feat(sessionmsg): cap pending mailbox depth at DefaultSessionMsgMaxPending (card t253)
$ git log -1 --format='%h %ad %s' --date=short WT-ci-flake-series -- internal/sessionmsg/stoprule_test.go
def99739d 2026-08-27 fix(SPEC-CI-FLAKE-SERIES-001): M2 stop-rule re-poll closes poller TOCTOU (t278)
```

브랜치의 `total = 100` 이 옛 상태다. develop 은 t253 이 도입한 깊이 상한(64) 때문에 60 으로
낮춰 놓았고 그 이유를 주석으로 적어 놓았다.

### E5 — 잘못 실행한 전치와 그 되돌림 (실행 기록)

```
$ git checkout 03cf93cc2 -- .moai/reports/t278     # 9 파일 스테이징 (오판)
$ git add .moai/reports/t278-t267/verdict.md
The following paths are ignored by one of your .gitignore files:
.moai/reports/t278-t267                            ← 여기서 ignore 규칙을 처음 봤다
$ git check-ignore -v .moai/reports/t278/forensics.md
(t278 NOT ignored)                                 ← 인덱스에 있으므로 보고되지 않은 것

$ git rm -r --cached --quiet .moai/reports/t278 && rm -rf .moai/reports/t278
$ git status --porcelain | wc -l
0                                                  ← 트리 원상 복구 확인
```

## Baseline-attribution

이번 판정의 모든 비교는 이 트리에서, 이 시점에 실행한 명령의 출력이다.

| 대상 | 값 | 확인 명령 |
|---|---|---|
| base / `origin/develop` | `d9951c50e06d2eee7b078b5e8cf3267e4db6db20` | `git rev-parse HEAD` · `git rev-parse origin/develop` (동일) |
| develop 동기 상태 | `0 0` | `git rev-list --count --left-right origin/develop...develop` (fetch 직후) |
| 저장소 기본 브랜치 | `origin/develop` | `git symbolic-ref --short refs/remotes/origin/HEAD` |
| 판정 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t278-t267` | `git rev-parse --show-toplevel` |
| 작업 브랜치 | `WT-flake-report-ledger` | `git branch --show-current` |
| t278 원본 브랜치 tip | `2aec09c27` | `git log develop..WT-ci-flake-series` |
| t267 원본 브랜치 tip | `052830c0a` | `git log develop..WT-taskstop-name-reclaim` |
| 추적 해제 커밋 | `ba60eb6d5` (2026-09-14) | `git log -S'Card/audit reports are local-only artifacts' -- .gitignore` |
| t278 본체 착지 커밋 | `379b310a6` (PR #1666) | `git log --all -- .moai/reports/t278/forensics.md` |

배차가 준 기준값(develop `d9951c50e`, t278 미병합 13 / t267 미병합 14)은 이 트리에서 그대로
재측정돼 일치했다. 원본 두 브랜치는 읽기만 했고 건드리지 않았다.

## Gaps

- **t267 의 8개 diverged 파일을 줄 단위로 전수 대조하지 않았다.** 고유 산출물 마커(E2)와
  `삽입 ≪ 삭제` 형태로 develop-앞섬을 판정했다. 그 8 파일 안에 t267 이 넣었고 develop 이
  나중에 지운 줄이 있는지는 관측하지 않았다.
- **끊긴 인용 경로를 해소하지 않았다.** develop 에 이미 착지한 `acceptance.md`(7곳)·
  `plan.md`(9곳)·`progress.md`(4곳)·`CHANGELOG.md`(1곳)이 `.moai/reports/t278/*` 를 증거로
  인용하는데 그 경로가 저장소 안에서 해소되지 않는다. 이것이 추적 해제 결정의 알려진
  대가인지, 아니면 별도로 처분할 사안인지는 **운영자 판정 사항**이라 손대지 않았다
  (Residual-risk 1).
- **AC-CFS-007 관측 창을 닫지 않았다.** SPEC 은 `status: implemented` 이고 창이 열려 있다.
  그 원장이 추적 해제된 `reproduction-rate.md` 라는 점은 위 인용 문제와 같은 사안이다.
- **Go 테스트를 돌리지 않았다.** 이 브랜치는 Go·템플릿·훅 표면을 한 줄도 건드리지 않았다
  (`git diff --name-only develop..HEAD` = verdict.md 1건). 영향받는 패키지가 없어 세정
  검증을 돌릴 대상이 없다.
- **원격 착지·CI 판정은 관측되지 않았다.** push 는 리드 일괄이며 이 세션은 하지 않았다.

## Residual-risk

- **인용 경로 부재가 실제로 문제일 가능성.** `.moai/reports/**` 를 추적 해제한 결정과, 그
  경로를 증거로 인용하는 SPEC 문서가 저장소에 남아 있는 상태는 서로 긴장 관계다. 해소 방향이
  둘 있다 — 인용을 걷어내거나(문서 수정), 증거를 예외로 되살리거나(`git add -f`). 어느 쪽도
  이 카드의 범위가 아니고 운영자 결정이라 판단해 보고만 한다. **다만 t547·t675·t802·t877·
  t897 의 `verdict.md`(+t675 의 red/run 증거)는 추적 해제 이후에도 예외로 추적되고 있어**,
  "증거를 예외로 되살린다"는 선택지에 선례가 있다.
- **이 판정서 자체가 ignore 규칙을 뚫고 들어간다.** `.moai/reports/*` 에 걸리므로
  `git add -f` 로 넣었다. 위 5개 카드 verdict 의 선례를 따른 것이고, 그 선례가 잘못이라면
  이 파일도 같이 빠져야 한다.
- **t267 두 브랜치 마감은 리드 판정 사항이다.** 이 판정서는 "전부 중복"을 증거와 함께 적었을
  뿐이고, 브랜치·워크트리 처분은 하지 않았다.

---

## 처분 요청

| 항목 | 요청 |
|---|---|
| t267 (`WT-taskstop-name-reclaim`) | **전부 중복 — 마감 요청.** 옮길 것이 없다 |
| t278 (`WT-ci-flake-series`) | **옮길 것 없음 — 마감 요청.** 유일한 미반영분이 운영자 추적 해제 대상(`ba60eb6d5`) |
| 별건 보고 | t278 인용 경로 18곳 + 미결 AC-CFS-007 원장이 저장소에서 해소되지 않음 → 운영자 판정 필요 |
| 이 브랜치 | `WT-flake-report-ledger` — 판정서 1 커밋뿐. 병합 여부는 리드 판단 |
| 워크트리 | `.claude/worktrees/t278-t267` · `.claude/worktrees/t278` · `.claude/worktrees/t267` 모두 **보존** |
