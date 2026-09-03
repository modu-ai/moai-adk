# t472 — 축 B~F 실측 (plan 근거)

측정 트리: `.claude/worktrees/t472`, HEAD `a6517e868`.
판정 바이너리: 이 트리에서 빌드(`go build ./cmd/moai`, rc=0). 설치본 아님 — VCI §2.2.
측정자: lane-3. 리드가 건넨 t237 사례를 인용하지 않고 이 트리에서 다시 쟀다.

## 축 F (신규) — 착지 판정이 제목 귀속과 본문 언급을 구분하지 않는다

### 기제

`internal/kanban/prlink_landed.go:LandedGrepArgs` 가 만드는 argv:

    git log <ref> --perl-regexp --grep=\btNNN\b --oneline

`--grep` 은 **커밋 메시지 전체**를 훑는다. 따라서 다른 카드의 커밋이 본문에서 이 카드를
언급하기만 해도 `landed` 가 된다. 같은 파일의 주석이 이 위험을 이미 알고 있다 —
"a card's first matching commit may be another card's report commit that merely mentions it" —
그러나 그 인식은 **SHA 노출을 막는 데만** 쓰였고 매칭 자체는 막지 않았다.

### 재현 (t237, origin/main = 현재 판정 ref)

    git log origin/main --perl-regexp --grep='\bt237\b' --oneline
      539349c5b docs(t230): t230 sync-audit evidence ... (#1649)
      32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)

두 건 다 **t230 커밋**이다. `32d2221fa` 본문에 "the card about to change the hook body is
t237/#1641. Re-measured: the issue is OPEN" 이 있다 — 착지가 아니라 *미착지라는 서술*이
착지 근거로 읽혔다. 제목에 t237 이 있는 커밋은 0건.

트리 바이너리 판정: `moai todo pr t237` → `landed`. **거짓 양성 확정.**

### 규모 — 현재 ref(origin/main)

전 큐 57행에 대해 `todo pr` 이 내는 `landed` 는 2건(t237, t312)이고 **둘 다 거짓 양성**이다
(제목 귀속 각 0건, 본문 언급 각 2건). 현재 판정의 정밀도 = 0/2.

### 규모 — 축 A 가 고쳐질 경우(develop)

이 값이 축 A 판정에 직접 걸린다. 살아있는 카드 31장을 develop 에 물으면:

| | 카드 | 수 |
|---|---|---|
| `landed` 로 판정될 것 | t204 t216 t237 t315 t359 t401 t436 t440 t443 | 9 |
| 참 양성(제목 귀속 있음) | t401 t440 | 2 |
| 거짓 양성 | t204 t216 t237 t315 t359 t436 t443 | 7 |

대조군(공허하지 않음): develop 제목 귀속 id 291개, 메시지 언급 id 379개 — 두 피연산자 모두
비어 있지 않다.

### [HARD] 순서 결론 — F 가 A 보다 먼저다

축 A(ref 를 develop 로 발효)만 먼저 착지하면, 지금 잠자는 결함이 **9장 중 7장의 잘못된
닫기**로 깨어난다. A 는 증상을 고치는 것이 아니라 거짓 양성의 모집단을 2장에서 9장으로
키운다. F 를 먼저 세우고 A 를 그 위에 얹어야 한다.

### 순진한 수리는 불충분하다 (뮤턴트 2종)

"제목만 보면 된다"는 처방은 거짓 양성 7장 중 2장을 못 잡는다:

    docs(t263): die-at-exit reproduced (0/5) — remedy sequenced behind t216
    chore(catalog): revert sync-auditor hash to develop value — t443 jurisdiction (t461)

둘 다 **다른 카드의 제목이 이 카드를 언급**한 형태다. 판별식은 등장 위치가 아니라
**귀속 위치**여야 한다 — conventional-commit scope(`docs(t440):`), 말미 괄호(`(t401)`),
머지 제목(`Merge card t440 (...)`). 참 양성 2장은 정확히 이 세 형태로 잡힌다.

## 축 B — 폴백은 조용하다. 단, 카드가 지목한 곳이 아니다

카드는 "unknown 도 카드를 닫는다"를 문제로 적었으나 코드는 이미 정직하다:

- `todo done` 은 플래그 없이도 `done <id> landing=unknown` 을 찍는다 — "가드가 돌지 않았다"가
  바이트로 남는다(`todo.go:505`).
- `--require-landed` 는 not-landed 에만 거절하고 unknown 에는 stderr 고지 후 진행한다
  (`todoRequireLanded`, 비대칭이 의도적이며 주석에 근거가 적혀 있다).

조용한 것은 **ref 폴백**이다. `LandedRefFor` 는 설정이 비면 아무 고지 없이 `origin/main` 을
쓴다. 판정 시점 출력 `done <id> landing=landed` 는 **어느 ref 가 답했는지 담지 않는다.**
help 문안은 ref 를 이름하지만 help 를 부르지 않는 운영자는 보지 못한다. 이 리포에서는
그 폴백이 곧 오답이므로(통합은 develop, 질문은 origin/main) 축 B 는 실재한다 —
다만 처방은 "unknown 거절"이 아니라 **판정 줄이 ref 를 운반하고, 폴백을 밟았을 때 고지**다.

## 축 C — 착지 증거 컬럼 부재 (카드 서술 참)

    CREATE TABLE archived_items (
      seq INTEGER PRIMARY KEY, id TEXT NOT NULL UNIQUE, text TEXT NOT NULL,
      added_at TEXT NOT NULL, spec_id TEXT, state TEXT NOT NULL, position INTEGER NOT NULL
    );

착지 SHA·판정 ref·판정 시각 어느 것도 없다. `done` 이 찍는 `landing=` 한 줄은 stdout 에만
남고 저장되지 않는다. 아카이브 126행 전부가 근거 없이 닫혀 있다.

주의 — 이 축은 **t359(착지 증거 컬럼 스키마)와 같은 표면**이다. t359 는 picked 상태이고
plan-audit iter-1 재설계 2건을 안고 있다. 중복 발행하지 말 것.

## 축 D — 카드↔SPEC 연결 부재 (카드 서술보다 넓다)

    items:          57행 중 spec_id 공란 57 (100%)
    archived_items: 126행 중 spec_id 공란 126 (100%)

카드는 "53장 전부"라 적었으나 현재 57장이고, **아카이브까지 포함하면 183장 전부**다.
`todo next <n> --spec <SPEC-ID>` 로 붙이는 경로가 존재하는데도 실사용이 0이다 — 스키마
결함이 아니라 운용 결함이다. 카드가 "이 카드에서 수리하지 않음"으로 명시했으므로 범위 밖.

## 축 E — 탐지 표면은 이미 있다. 없는 것은 신뢰성과 주기다

`moai todo pr` 이 전 큐에 대해 착지 열을 낸다(57행 판독, 5-값 outcome, 쓰기 0). 카드가
"조회 서브커맨드가 후보"라 적었으나 **이미 배포돼 있다.** 남는 공백은 둘:

1. 그 열의 판정이 축 F 때문에 신뢰할 수 없다 — 지금 내는 landed 2건이 둘 다 거짓이다.
2. 아무도 주기적으로 돌리지 않는다. 적체는 리드의 수동 스윕으로만 발견됐다.

따라서 축 E 는 "표면 신설"이 아니라 **F 착지 후 운용 배선**이다. 새 서브커맨드를 만드는
구현은 이 축의 대표 뮤턴트다.

## 미검증 / 잔여 위험

- `todo pr` 의 `landed` 는 열린 PR 이 없을 때만 나온다(t204 는 `linked #1685`). 따라서
  위 "0/2 정밀도"는 landed 로 표시된 모집단에 대한 것이지, 언급-매칭 전체가 아니다.
- 제목 귀속 프록시는 괄호 관례 + bare-token 두 형태로 쟀다. 세 번째 관례가 있다면
  참 양성이 과소계수됐을 수 있다 — 다만 그 방향의 오차는 위 결론(F 선행)을 약화시키지 않는다.
- 축 A 의 primary 체크아웃 파킹 원인은 `premise-recheck.md` 의 미검증 항목 그대로다.
- 종전 `premise-recheck.md` 의 잔여위험 절에 "소비자를 아직 세지 않았다"가 남아 있으나
  같은 문서 앞 절에서 4곳을 이미 셌다 — 그 줄은 스테일이다.

---

# 축 A+B 합병 — 「착지의 정의」 (리드 판정 2026-09-03 반영)

리드가 축 A 를 「설정 한 줄」에서 설계 질문으로 재서술하도록 판정했다. 릴리스 대기는
기각됐다(시점이 t204 배포 게이트에 종속 + 적체가 릴리스 주기마다 재발).

## Q1 — 소비자 4곳이 원하는 의미 (리드 [HARD] 실측 요구)

| 소비자 | 위치 | 이 값으로 무엇을 하는가 | 원하는 의미 |
|---|---|---|---|
| 카드 워크트리 base | `session_worktree.go:215` | 새 카드 트리를 어느 브랜치에서 팔지 | **통합**(develop) — 새 작업은 통합 팁에서 시작한다 |
| doctor 점검 | `doctor_worktree_base.go:43` | 설정값 vs `origin/HEAD` 정합 보고 | 소비자 1을 따라감 → **통합** |
| 착지 ref | `prlink_landed.go:75` | 카드가 착지했는가 | **통합** — 카드 수명주기는 통합에서 끝난다 |
| SessionStart 재정렬 | `hook/worktree_base_branch.go:156` | `origin/HEAD` 를 설정값에 맞춰 **쓴다** | 소비자 1의 집행부 → **통합** |

**결과: 네 소비자 모두 통합(develop) 의미를 원한다. 릴리스(main) 의미를 원하는 소비자는
없다.** 한 값이 넷을 다 답할 수 있다 — 키 이름(`worktree_base_branch`)부터가 통합-base
이름이고, 착지 판정이 그것을 빌려 쓴 것은 의미 과적이나 방향은 어긋나지 않는다.

단, 네 번째는 조회가 아니라 **저장소 공유 ref 쓰기**다(`git remote set-head`). 이 설정은
불활성이 아니며, `git log origin` 같은 git 자체 동작까지 따라 움직인다. 「기제 변경 0」이
성립하지 않는 두 번째 이유가 여기다.

## Q2 — 진짜 막힌 곳은 의미가 아니라 **읽는 자리**다

의미에 이견이 없으므로 남는 것은 해석 경로다. 실측:

    git symbolic-ref refs/remotes/origin/HEAD   → refs/remotes/origin/develop
    primary .git/HEAD                            → ref: refs/heads/main
    primary git-strategy.yaml:7                  → worktree_base_branch: ""

**저장소의 git 수준 답은 이미 `develop` 이다.** 그런데 착지 판정은 primary 체크아웃의
설정 파일을 읽고, 그 파일이 공란이므로 `DefaultLandedRef` 하드코딩 `origin/main` 으로
떨어진다. 즉 판정은 저장소가 이미 아는 답을 무시하고 상수를 쓴다.

## 설계 선택지 (SPEC 이 판정할 것)

| | 안 | 성질 |
|---|---|---|
| A1 | 릴리스 대기 | **기각됨**(리드 판정) — 적체가 릴리스 주기마다 재발 |
| A2 | primary 설정에 develop 기입 | 커밋 불가(primary 는 main 위) + 다음 `moai update` 가 `.moai/config` 를 통째 재배포해 되돌린다(CLAUDE.local.md §2.3) |
| A3 | **폴백 사슬을 설정 → `origin/HEAD` → `origin/main` 3단으로** | 이 리포에서 설정 편집·릴리스 대기 없이 즉시 develop 을 묻는다. 하드코딩 상수보다 저장소의 실제 기록을 우선하는 것이라 `CLAUDE.local.md §14`(하드코딩 금지)와도 같은 방향 |
| A4 | 착지 ref 만 워크트리 설정에서 읽기 | `todo.go:81-90` 의 "한 저장소 한 큐" 근거를 뒤집는다 — 근거를 세워 반박하지 않는 한 채택 불가 |

A3 가 유력하나 판정은 SPEC 소관이다. **A3 도 축 F 를 대체하지 않는다** — 오히려 A3 가
착지하면 거짓 양성 모집단이 2장에서 9장으로 커지므로 F 가 선행이다.

## 부수 관측 (리드 전달, 범위 침범 없이 기록만)

`moai todo done t278` → `done t278 landing=unknown`. 플래그 없이 닫으면 판정이 아예
돌지 않고, 그 `unknown` 조차 stdout 한 줄로만 남고 저장되지 않는다(축 C·t359 인접).
