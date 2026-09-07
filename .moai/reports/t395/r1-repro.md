# t395 R1 — foreman 큐 감시가 죽어 있음 (실행 재현)

트리: `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `d2d2a7fdb`
리드 [HARD] 지시("코드 판독으로 끝내지 말고 실제로 돌려 관측할 것")에 대한 응답.
`reader-surfaces.md` R1 의 Gap 을 닫는다.

## Claim

`.claude/skills/moai-kanban-foreman/SKILL.md:95` 의 persistent Monitor 는 마이그레이션된
프로젝트에서 **큐가 실제로 바뀌어도 이벤트를 내지 않는다.** 코드 판독이 아니라 관측이다.

## Evidence

### 격리 (라이브 큐 무손상)

fixture 프로젝트 루트: `/private/tmp/.../scratchpad/r1fix` (git 메타데이터 없음)
→ `ResolveTodoQueueRoot` 의 홈 폴백으로 큐가
`/Users/goos/.moai/todo/r1fix-497dc824/.moai/state/todo/` 에 생성됨.

```
$ ls -la /Users/goos/.moai/todo/r1fix-497dc824/.moai/state/todo/
-rw-r--r--  40960 Sep  2 08:08 backlog.db
-rw-r--r--     59 Sep  2 08:08 backlog.lock
```

`backlog.json` 없음 — 마이그레이션 완료 후(및 신규 설치)의 정상 레이아웃, 즉 R1 Case A.

라이브 큐 무손상 확인: `select id,text from items where text like '%fixture card%'` → 0행
(`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`).

### A. 감시 스크립트가 분기하는 값의 결정적 측정

스크립트는 `[ -f "$f" ]` 와 `cksum "$f"` 두 값에만 분기한다. 실제 큐 변경을 사이에 두고 잰다:

```
BEFORE  json: No such file or directory
BEFORE  db  : 1125950968 40960 .../backlog.db

$ moai todo add "fixture card two — queue mutation under watch"   → t2 2

AFTER   json: No such file or directory
AFTER   db  : 1188524958 40960 .../backlog.db
```

큐는 확실히 바뀌었고(db cksum 1125950968 → 1188524958, 카드 1→2장), `backlog.json` 은
그 사이 계속 부재다. 따라서 스크립트의 `cur` 은 두 반복 모두 `missing` 으로 동일하고
`[ "$cur" != "$last" ]` 는 영원히 거짓이다 — `backlog changed` 가 나올 경로가 없다.

### B. 스크립트를 원문 그대로 Monitor 로 실행 (관측)

foreman 이 쓰는 것과 같은 도구(Monitor)에, SKILL.md:95 의 루프를 문자 그대로 넣었다.
유일한 변경은 `f=` 의 상대경로를 fixture 큐의 절대경로로 바꾼 것뿐이다(상대경로가 해석되는
바로 그 경로).

| 감시 | 대상 | 창 | 창 안에서 일어난 일 | 결과 |
|---|---|---|---|---|
| JSON-WATCH (원문) | `backlog.json` | 45s | `moai todo add` → t3 (카드 2→3장) | **이벤트 0건**, timeout kill |
| DB-WATCH (대조군) | `backlog.db` | 45s | 같은 변경 | **`DB-WATCH backlog changed` 발화** |

대조군이 같은 창·같은 변경에서 발화했으므로 JSON-WATCH 의 침묵은 **하네스가 루프를 안 돌린
탓이 아니다.** 감시 대상이 틀린 탓이다.

## Baseline-attribution

이번 실행에서 격리 fixture 큐에 대해 측정. 바이너리는 `/Users/goos/go/bin/moai` (v3.1.2,
2026-09-01 06:37 빌드) — SQLite 큐 저장소(08/27 착지) 이후 빌드다.
스크립트 원문은 워크트리 `d2d2a7fdb` 의 `.claude/skills/moai-kanban-foreman/SKILL.md:95`.

## Gaps (미검증)

- Case B(**스테일 json 이 존재**하는 이 저장소의 현재 상태)는 Monitor 로 재현하지 않았다.
  A 의 논증으로는 cksum 이 고정이라 같은 결론이지만, 관측은 Case A 만 했다.
- foreman 반복 전체(`moai-kanban-foreman` 스킬의 나머지 단계)를 돌리지는 않았다 — 감시 단계만
  격리해 돌렸다.
- 정본을 `backlog.db` 로 바꾸면 낫는지는 대조군이 시사할 뿐, 수리안의 검증은 run-phase 몫이다.

## Residual-risk

- 대조군이 보인 것은 "db 를 보면 발화한다"까지다. db 는 WAL 을 쓰므로 변경이 즉시 메인 파일
  cksum 에 반영되지 않는 구간이 있을 수 있다 — 수리안이 cksum 을 그대로 db 로 돌리는 것이라면
  그 지점을 run-phase 에서 확인해야 한다(이번 관측에서는 반영됐다).
- fixture 는 홈 폴백 루트라 primary 체크아웃 경로와 완전히 같지는 않다. 감시 스크립트는 경로만
  보므로 결론에 영향이 없다고 판단했으나, 이는 판단이지 측정이 아니다.

## 정리

fixture 큐(`~/.moai/todo/r1fix-497dc824/`)와 scratchpad(`.../scratchpad/r1fix`)는 이 문서를
남긴 뒤 제거한다.
