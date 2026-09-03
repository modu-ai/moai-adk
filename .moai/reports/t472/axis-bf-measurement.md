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

---

# 두-귀속 제목 개체 수 (리드 [HARD] 2026-09-03)

리드가 「한 제목이 서로 다른 두 카드의 귀속 형태를 갖는 경우」를 Kickoff 전 판정 대상에서
빼되 **개체 수를 먼저 재라**고 지시했다. 그리고 내가 든 예 두 개가 그 형태가 아니라고
정정했다 — **리드가 옳다**:

- `chore(catalog): … t443 jurisdiction (t461)` — 스코프는 `catalog`(id 아님). 귀속 형태는
  말미 괄호 `(t461)` 하나뿐이고 `t443` 은 자유 텍스트 언급이다.
- `docs(t263): … behind t216` — 귀속 형태는 스코프 id `t263` 하나뿐. `t216` 은 말미 괄호가
  아니다.

두 예는 **MUT-SUBJECT-ONLY 를 무너뜨리는 근거로는 여전히 유효**하다(제목 안에 있으나
귀속 위치가 아니다). 다만 두-귀속 케이스의 사례는 아니었다.

## 측정

코퍼스: `refs/remotes/origin/develop` = `7835148d3`(fetch 후 재확인 — 리드가 말한
`1b9c02991` 은 **로컬** develop 이라 원격 코퍼스는 불변), 제목 5,837행.

| 형태 | 패턴 | 건수 |
|---|---|---|
| 스코프 id | `^[a-z]+\(tNNN\)!?:` | 290 |
| 말미 괄호 id | `\((card )?tNNN\)$` | 869 |
| 머지 제목 | `^Merge card tNNN` | 5 |
| **둘 다 가진 제목** | 위 1 ∧ 2 | **34** |
| ├ 두 자리 id 동일 | `docs(t371): … (t371)` 형태 | **34** |
| └ **두 자리 id 상이 = 모호 케이스** | | **0** |

공허하지 않음: 네 피연산자 모두 비어 있지 않고, 교집합(34)도 비어 있지 않다. 0 은 실제 0이다.

**측정 중 잡은 함정 1건**: 이 환경의 `grep` 은 ugrep 이고 `-E` 에서 백레퍼런스를 지원하지
않는다(`invalid escape`, rc=2). 대조군(`t1 t1` / `t1 t2`)을 먼저 태우지 않았다면 오류 출력을
「모호 케이스 0건」으로 오독했을 것이다. 최종 측정은 `-P` 로 했고 대조군이 통과했다.

## 처분 — 판정 없이 결정론적 타이브레이크 1줄

관측 0건이므로 가설적 케이스다. 설계로 다투지 않고 타이브레이크를 문서화하되, 그 근거로
「관측 0건」과 위 표를 적는다. 방향은 **말미 괄호 우선, 스코프 차순**이고 코퍼스가 그것을
지지한다:

- 말미 괄호가 3배 흔하다(869 vs 290).
- 스코프는 카드 id 가 아닌 경우가 실재한다(`chore(catalog):`, `chore(reports):`) — 즉
  스코프는 카드 운반자가 **아닐 수 있는** 자리다.
- `AGENTS.md` §3 이 커밋 메시지의 카드 id 운반을 의무로 두는데, 실사용에서 그 자리가
  말미 괄호다.

N ≥ 1 이 되면 그때 실례를 들고 리드 판정을 받는다. 표본 0 위에서 미리 다투지 않는다.

---

# D1 독립 재현 (plan-audit iter-1 이후, lane-3)

감사가 D1 을 BLOCKING 으로 냈다. 감사 판정을 그대로 싣지 않고 이 트리에서 다시 쟀다.
코퍼스: `origin/develop 7835148d3` 제목 5,837행.

## 재현 결과 — 감사와 일치하고, 규모는 더 크다

`REQ-TLA-002` 형태 3 이 "머지 제목이 카드를 이름한다"로 적혀 있어 **occurrence test** 다.
그 술어가 실제로 무엇을 잡는지:

| 측정 | 건수 |
|---|---|
| 형태 3 원문 `^Merge card tNNN` (귀속 형태) | **5** |
| 머지 제목 중 카드 id 를 어디든 담은 것 (occurrence test 의 범위) | **146** |
| 그중 형태 2(말미 괄호)로 이미 잡히는 것 | 91 |
| **형태 2·3 어느 쪽도 아닌데 occurrence test 는 잡는 것** | **50** |
| 그중 흡수 방향(`into WT-`) 머지 | **26** (감사 수치와 일치) |

귀속 형태 5건 대비 occurrence test 는 146건 — 29배다.

## 왜 위험한가 — 표본이 스스로 말한다

    Merge branch 'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)
    Merge branch 'develop' into WT-inbox-drain-gap (absorb t280, lane-15 window; includes t239 merge e79c010b8)

첫 줄은 t387·t386 **둘 다**, 둘째 줄은 t280·t239 **둘 다** 담는다. 전자는 의존 메모이고
후자는 흡수 기록이다 — 어느 쪽도 귀속이 아니다. 이 카드가 축 F 에서 기각한 바로 그 형태
(본문 언급을 귀속으로 오독)가 **SPEC 자신의 요구 안에** 들어 있다.

`MUT-MERGE-ANY-TOKEN`(형태 3만 occurrence 로 구현한 뮤턴트)이 AC-TLA-001..007 을 온전히
통과한다는 감사 지적도 성립한다. 오늘의 7/9 를 움직이지 않으므로 **조용히 착지한다** —
그것이 이 결함을 BLOCKING 으로 두어야 하는 이유다.

## 처방 방향 (SPEC 소관)

형태 3 을 귀속 형태로 좁힌다. `^Merge card tNNN` 은 귀속이고, `Merge branch 'X' into Y (…)`
의 괄호 안은 형태 2 가 이미 판정한다. 흡수 방향 머지는 어느 형태로도 귀속이 아니다 —
그 머지의 주체는 브랜치이지 카드가 아니기 때문이다.

## 프로세스 결함 1건 (내 잘못, 기록)

감사가 도는 동안 이 워크트리에 커밋 2건(`a72fb378c`, `161554e4a`)을 얹었다. 감사 중인
워크트리는 작성자가 하나여야 한다(`agent-common-protocol.md` § Background Agent Execution).
감사가 이를 "foreign writes to an actively audited worktree"로 지적했고 옳다. 두 커밋 다
근거 문서라 판정을 뒤집을 성질은 아니지만, 감사가 읽은 트리와 지금 트리가 다르다는 사실
자체가 결함이다. 다음 감사에서는 창을 닫고 기다린다.

---

# plan-audit iter-2 검증 (lane-3)

판정 PASS-WITH-DEBT **0.875** (임계 0.80, iter-1 대비 **+0.0625** 단조 상승), iter-1 결함 9건
전부 VERIFIED FIXED. 신규 5건(D1·D2·D3 major). 감사 주장 두 건을 직접 확인했다.

## 감사가 틀린 것 1건 — 감사 예산은 소진되지 않았다

감사가 "Tier M 상한 2회를 소진했으므로 D1·D3 수정은 추가 감사 라운드로 확인할 수 없다"며
선택지를 둘로 좁혔다(미감사 수용 / 부채 수용). **규칙과 다르다.**

`.claude/rules/moai/workflow/spec-workflow.md:160`:

> Maximum 3 plan-auditor iterations per SPEC plan-phase; after iter3, escalate via
> PASS-with-debt OR scope-reduction OR explicit user override.

상한은 **3회**이고 2회를 썼으므로 **iter-3 이 남아 있다.** 같은 줄의 STOP 조건도 살펴야
하는데 — `iter(N+1)` 점수가 `iter(N)`보다 낮으면 STOP — 이번은 0.8125 → 0.875 로 **올랐다**.
즉 계속 진행이 규칙상 인가된 경로다. 감사가 제시한 두 갈래는 실제보다 좁았다.

## 감사가 맞은 것 — D1 공허성 확인

`t412` 가 `acceptance.md` 에 **0회** 등장한다. 형태 3b 는 오직 t412 를 회수하려고 존재하는데,
그것을 단언하는 기준이 없다. 즉 **형태 3b 를 통째로 빼도 12개 기준이 전부 초록이다.**
이 카드가 축 F 에서 기각한 것과 같은 계열 — 실패를 관측한 적 없는 검사 — 이 이번엔 SPEC 의
수정본 안에 들어왔다.

## D2 확인 — 그리고 양쪽이 놓친 벡터 1건

저자의 "과소계수는 정확히 1건(t250)"은 재현되지 않는다는 감사 지적이 맞다. 표본을 직접 보니
형태가 여러 갈래다:

    Merge branch 'worktree-t40': t40 — moai update observability (3 quiet failures)
    merge: t36 — in-run calibrated latency bounds … — absorbs t2
    test(timing): … (card t36, absorbs t2)
    Merge branch 'WT-t250-followup' into develop — card t279 AC-GF-022 ordering deviation record

**양쪽이 명명하지 않은 벡터**: 마지막 줄에서 `WT-t250-followup` 은 **브랜치 이름**이 t250 을
운반하는데 이 커밋의 카드는 **t279** 다. 브랜치 이름은 제목 안에 있으므로 occurrence test 는
t250 을 착지로 읽는다. 형태 3b 는 말미 괄호를 요구하므로 발화하지 않고, 비귀속 규칙은 대상이
`develop` 이라 발화하지 않는다 — 즉 현행 §A.4 에서 이 제목은 어느 형태에도 안 걸리지만,
**판별식을 조금이라도 느슨하게 잡으면 곧바로 거짓 양성이 된다.** 수정 라운드에 넣는다.

`worktree-t40` 은 D4(=`WT-` 접두 규약을 SPEC 이 소유하지 않는다)의 실증이기도 하다 — 이
코퍼스 자신이 구 접두사를 쓴다.

`(card t36, absorbs t2)` 는 괄호군 안에 **두 카드**가 있는 사례다. 앞서 "두-귀속 0건"으로
측정한 것은 *스코프 id ∧ 말미 괄호 id* 조합이었고, 이것은 *한 괄호군 안의 두 토큰*이라
다른 축이다 — §D 타이브레이크가 이 축도 덮는지 확인이 필요하다.

## 다음

iter-3 이 남아 있고 점수가 올랐으므로, D1·D2·D3 수정 후 iter-3 으로 확인한다. 부채로
받아들이지 않는다.

---

# plan-audit iter-3 검증 (lane-3) — D3-1 재현

판정 PASS-WITH-DEBT **0.8375**. iter-2 대비 **−0.0375 하락**이라 `spec-workflow.md:160` 의
STOP 조건(점수 하락)이 발화했고, 감사 상한 3회도 소진됐다. Kickoff 준비: **아니오**.

## D3-1 재현 — 감사가 맞고, 범위는 더 넓다

형태 2 의 `)$` 앵커가 **PR 참조가 뒤에 붙은 자기 자신의 모양**을 놓친다:

    grep -cE '\([^()]*t[0-9]+[^()]*\) \(#[0-9]+\)$' <develop 제목>  → 43
    같은 명령, origin/main 제목                                      → 43

표본:

    docs(SPEC-CI-FLAKE-SERIES-001): sync-phase — … (t278) (#1670)
    SPEC-TEAMMATE-REVIVAL-GUARD-001: … (t267) (#1667)
    feat: lead-session deputy — … (SPEC-LEAD-DEBOTTLENECK-001, t283) (#1664)

**세 가지가 겹쳐 blocking 이 맞다.**

1. `plan.md` §D 의 유일한 관용 근거("잔여 7건은 전부 구세대라 영향 없음")가 **거짓**이다 —
   위 표본은 #1657~#1670, 이 저장소에서 **가장 최근** 병합들이다.
2. 잔여 수치가 **세 번째로** 틀렸다: 1 → 19/7 → 최소 46. 매 라운드 커졌다.
3. **같은 형태가 `origin/main` 에 43건 있고, M1 이 걷는 ref 가 바로 그것이다.** 즉 M1 착지
   시점에 곧바로 드러나는 모집단이다.

이 카드가 다루는 결함 계열이 **세 번째로** 자기 수정본 안에서 재발했다 — 형태 3b 공허(iter-2),
형태 3a 공허(수정 중 자체 발견), 그리고 이번엔 앵커가 조용히 놓치는 모양이다. 판별식을 손으로
열거하는 접근 자체가 이 실패를 반복 생성한다는 신호로 읽는다.

## 감사가 스스로 정정한 것 (기록)

iter-2 의 D3 "76건 vs 0건 창" 주장을 감사 자신이 **철회**했다. 저자의 `origin/main` 측정
(4,457 / 101 / 0 / 0)을 재현하고 "내 iter-2 창 주장은 틀렸으며 대체 결함이 건전하다"고 적었다.
구성 픽스처 `release/v9` 도 "허용 가능한 정도가 아니라 **필요하다**"로 판정을 바꿨다 —
실제 제목 중 하드코딩판을 구분하는 것이 없기 때문이다.

또한 저자가 감사 제안(말미 괄호군 첫 토큰 채택)을 기각한 것을 **지지**했고, "이 형태만 잡는
id 수" 열이 형식적 재진술이 아니라 **실제 관문**이라고 판정했다(형태 3b 공허를 그것이 잡아냈다).

## 나머지 신규 결함

D3-2 `t311`(`(closes t311)`) 미분류 · D3-3 병합 수 414 vs 실측 661 · D3-4 "9개" 목록이 11개를
출력하고 둘은 대상이 아니라 소스 · D3-5 `plan.md` §D 가 여섯 번째 형태 기각 규칙을 SPEC 자신의
실무보다 강하게 진술.

## 잔여 위험 (감사·저자·레인 모두 못 닫음)

46 은 하한이지 총계가 아니다. 89건 중 분류된 것이 19건, 이번에 43건이 추가로 드러났으나
어느 쪽도 전수 분류가 아니다. 39개 id 의 라이브 큐 상태도 미측정이다(38개가 backlog.db
두 테이블 어디에도 없음 — 알려진 저장소 분열과 정합).
