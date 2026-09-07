# t497 — 공정 결함 기록

## PD-1 — 감사 진행 중 트리에 두 번째 작성자가 생겼다 (레인 자책)

**무슨 일**: `plan-auditor` 를 HEAD `46ea34137` 로 배차해 놓고, 감사가 도는 동안
레인이 같은 워크트리에 `53886ba40 docs(t497): resolve the 3-vs-4 binding-row gap`
를 커밋했다. 감사 종료 시점에 감사자가 이를 발견하고 D0 으로 보고했다.

**규율**: `agent-common-protocol.md` § Background Agent Execution —
"While a worktree is being actively audited, it has exactly one writer.
The audit window runs from the opening measurement to the landed verdict."
감사 창은 열려 있었고, 레인이 그 창 안에서 썼다.

**귀속**: 외부 세션이 아니라 **레인 본인**이다. 리드에게 보고한다.

**영향 판정**: 감사자가 `git show --stat` 로 그 커밋이 `measurement.md` 하나만
건드렸음을 확인했고, 레인도 재현한다:
```
$ git show --stat --oneline 53886ba40
53886ba40 docs(t497): resolve the 3-vs-4 binding-row gap from in-tree evidence
 .moai/reports/t497/measurement.md
```
SPEC 산출물 4본은 미접촉이므로 **어떤 판정도 무효화되지 않았다.** 그러나
"영향이 없었다" 는 사후 확인이지 면책이 아니다 — 다음 감사 창에서는 창이 닫힐
때까지 쓰지 않는다.

**재발 방지**: 감사 배차 후 창이 닫히기 전까지 레인의 쓰기는 워크트리 **밖**
(scratchpad) 에 두고, 창이 닫힌 뒤 트리로 옮긴다.

**아이러니로 남기는 사실**: 그 커밋이 담은 발견(부재 3건 = 실린 3행)은
감사자가 **독립적으로 같은 결론**에 도달했고, 게다가 더 강한 경로를 찾았다 —
4번째 행이 `cross-session-messaging` 이며 후보 표 원본이
`.moai/reports/t196/csn003-table-4row.txt`(373 B, SPEC §B.D7 의 "4행 373 B" 와
정확히 일치) 로 커밋돼 있다. 레인이 창을 어기고 쓴 것이 결과적으로도 불필요했다.
