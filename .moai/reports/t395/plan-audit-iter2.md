# SPEC Review Report (iter-2): SPEC-BACKLOG-JSON-DISCLOSURE-001

카드: t395 · 트리 `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `36ca74c3d`
Iteration: **2/2** (Tier M ceiling — 마지막 반복)
**Verdict: PASS**
**Overall Score: 0.94** (iter-1 0.85 → **+0.09**, 단조 증가 — 점수 퇴행 STOP 조건 미해당)
Tier M PASS threshold 0.80 · ⏭️ skip-eligible

iter-1 보고서: `.moai/reports/t395/plan-audit.md` (덮어쓰지 않음)

## 트리 확인

```
$ git rev-parse HEAD
36ca74c3d818ee3d24a2ce903acbee6a99b7209b
$ git log --oneline -3
36ca74c3d docs(SPEC-BACKLOG-JSON-DISCLOSURE-001): iter-2 — close eight audit defects by wording (t395)
a09f65056 docs(t395): plan-audit PASS-WITH-DEBT 0.85, and correct this lane's own R5 citation
ea97eec22 docs(SPEC-BACKLOG-JSON-DISCLOSURE-001): plan-phase — ...
```

HEAD = 못박힌 `36ca74c3d` 와 일치. 감사 중 이동 없음. iter-1 은 `ea97eec22` 대상이었고,
델타는 `ea97eec22..36ca74c3d` 의 SPEC 4파일 + `reader-surfaces.md` 정정이다.

이번 라운드의 모든 판정은 **기준문의 명령을 실제로 실행해서** 냈다. 리드의 재측정을 인용하지
않았고, 저자의 서술을 근거로 삼지 않았다.

---

## Must-Pass Results (전면 재확인 — 문구 수정이 불변식을 깰 수 있으므로)

- **[PASS] MP-1** — `REQ-BJD-001`..`014` 14개, 결번·중복 0. `grep -o "REQ-BJD-[0-9]*" | sort -u | wc -l` → 14.
- **[PASS] MP-2** — GEARS 레이블 14개 전수 재판독. **REQ-BJD-003 이 `(Capability gate)/Where` →
  `(State-driven)/While` 로 정정**됐고(D6), 나머지 13개는 불변. 요구 레이어 기준 판정.
  비형식 문장 0건. REQ-BJD-002 에 붙은 breadth 해설 문단은 요구문 자체가 아니라 주석이며
  Event-driven 문장은 온전하다.
- **[PASS] MP-3** — 정본 12필드 + `tier` 전부 존재, 타입 정합. `version` 만 `"0.1.0"` → `"0.2.0"` 변경.
  거부 alias 0건.
- **[N/A] MP-4** — 단일 언어 SPEC. 불변.
- **[PASS] MP-5** — D7 참조 SPEC 4건 불변, 상태 불변(retired/superseded/archived 0).
- **[PASS] MP-6** — `grep -rc syscall` → 4파일 전부 0.
- **[PASS] MP-7** — `grep -rn 'NEEDS CLARIFICATION'` → 0건. 판단은 §"clarification gate 미배치" 절 참조.

**Must-Pass 7/7 통과.**

---

## Category Scores (단조 델타)

| Dimension | iter-1 | iter-2 | Δ | 근거 |
|-----------|--------|--------|---|------|
| Clarity | 0.80 | **0.90** | +0.10 | breadth 3중 분열 해소(세 표면 전부 "read"), 사이트/파일 단위 정정, §A.2 귀속 교체. 잔여는 "read command" 동사 집합 미열거. |
| Completeness | 0.90 | **0.95** | +0.05 | AC-BJD-015 커버리지 공백 폐쇄, §A.2 에 대조군 명시, plan §D.1 신설, §G 안티패턴 4건 추가. 잔여는 D9. |
| Testability | 0.70 | **0.90** | +0.20 | 이번 라운드 최대 이동. AC-BJD-007 실행 가능(실측 확인), AC-BJD-010 Given 자기증거화, AC-BJD-012/015 가 **미수리 트리에서 실제로 실패함을 실행으로 확인**. 잔여는 D10/D11. |
| Traceability | 1.00 | **1.00** | 0 | 표 14행 불변, AC 16개 전부 참조, 고아 0·미커버 0. 재계수 완료. |

집계 = (0.90 + 0.95 + 0.90 + 1.00) / 4 = **0.9375 → 0.94**

---

## 리드가 지목한 6개 압박 지점

### 1. D5 — AC-BJD-015 의 2-grep 이 공허하지 않은가 · **공허하지 않다 (실행 확인)**

리드의 우려는 정확한 형태였다 — "둘 다 요구하는 2-grep 이라도, 두 번째 grep 의 기대값이
미수리 트리가 이미 만족하는 것이면 여전히 공허하다." 그래서 리드의 재측정을 인용하지 않고
**기준문의 두 명령을 미수리 트리에서 그대로 실행**했다.

```
$ grep -rn 'state/todo/backlog\.json' internal/template/templates/ | wc -l
       4          ← 기대값 "exactly one match (todo-queue-storage.md:55)" · 4 ≠ 1 → FAIL

$ grep -rn 'moai/todo/<project-key>/backlog\.json' internal/template/templates/
internal/template/templates/.claude/skills/moai/workflows/todo.md:21:...`~/.moai/todo/<project-key>/backlog.json`
   count=1        ← 기대값 "zero matches" · 1 ≠ 0 → FAIL
```

**두 번째 grep 의 기대값(0건)을 미수리 트리는 만족하지 못한다** — 1건을 반환한다. 즉
`todo.md:21` 을 안 고치면 grep 2 가 실패하고, "Both commands are required" 이므로 기준 전체가
실패한다. 리드가 경계한 공허 형태가 아니다.

grep 1 도 독립적으로 실패하므로(4 vs 1), 두 명령은 각각 비공허하고 합쳐서 4개 사이트 + 대조군
1개를 빠짐없이 덮는다. iter-1 D5 는 닫혔다.

### 2. AC-BJD-012 — 로컬측 쌍둥이 · **내가 iter-1 에서 놓쳤다. 명시한다.**

먼저 인정할 것부터. **iter-1 감사는 D5 를 템플릿 미러(AC-BJD-015)에서만 잡고, 같은 구조적
공백이 로컬 `.claude/` 측(AC-BJD-012)에도 있다는 것을 놓쳤다.** iter-1 보고서의 D5 는
`internal/template/templates/` 만 다뤘고, AC-BJD-012 의 당시 문구("searched for
`state/todo/backlog.json` and for `~/.moai/todo/<project-key>/backlog.json`")를 통과시켰다.
그 문구는 두 패턴을 **산문으로 나열**했을 뿐 실행 명령이 아니었으므로, D2 에 적용한 것과
같은 잣대라면 그때 걸렸어야 한다. 저자가 지적 없이 스스로 수리한 표면이다.

새 기준으로서 감사한 결과 — 미수리 트리에서 세 형태 전부 실패한다:

```
$ grep -rn 'state/todo/backlog\.json' .claude/skills/ | wc -l
       3          ← 기대 0 → FAIL   (SKILL.md:170 · todo.md:17 · foreman SKILL.md:95)
$ grep -rn 'moai/todo/<project-key>/backlog\.json' .claude/skills/ | wc -l
       1          ← 기대 0 → FAIL   (todo.md:21)
$ grep -rn -e 'state/todo/backlog\.json' -e 'moai/todo/<project-key>/backlog\.json' \
    .claude/skills/ | grep -v -e 'backlog\.db' -e 'export' | wc -l
       4          ← 완화형 기대 0 → FAIL
```

세 번째가 특히 중요하다. 완화 조항("생존 매치가 `backlog.db` 또는 `export` 를 포함한 줄에
있으면 허용")이 **기준을 무력화하지 않는지**를 본 것인데, 미수리 4개 사이트 중 어느 줄도 그
두 토큰을 포함하지 않아 전부 필터를 통과해 살아남는다. 완화는 정당한 수리(`"여기 backlog.json
은 export 다"` 같은 라벨링 문장)를 허용하기 위한 것이고, 결함 상태를 통과시키지 않는다.

다만 완화형에는 좁은 잔여 구멍이 있다 → D10.

### 3. D4 — 내 권고가 기각됐다 · **기각이 옳다. 내 권고가 D4 가 막으려던 공허 그 자체였다.**

양보하지 않고 근거로 판단했고, 결론은 저자와 리드가 맞다는 것이다.

**SQLite 의미론.** `wal_autocheckpoint` 은 연결(connection) 속성이다 —
`sqlite3_wal_autocheckpoint(D,N)` 은 "database connection D" 에 대해 설정하며, 데이터베이스
파일에 **저장되지 않는다.** 따라서 fixture 를 만들 때 어떤 연결에서 그 pragma 를 걸어도,
그 값은 디스크에 남지 않고 다음 프로세스가 여는 연결로 전달되지 않는다.

**이 트리에서의 실측.**

```
$ grep -rn 'wal_autocheckpoint' internal/
NO MATCH in internal/
```

그리고 `moai todo` 가 여는 연결의 DSN 은 `backlogDSN`(`internal/kanban/backlog_sqlite.go:226-238`)
가 만드는데, 붙는 pragma 는 `busy_timeout`, `journal_mode(WAL)`, `_txlock=immediate` 셋뿐이다.
**autocheckpoint 은 건드리지 않으므로 SQLite 기본값이 적용된다.**

**결론.** iter-1 이 권고한 "`PRAGMA wal_autocheckpoint=0` 으로 자동 checkpoint 를 끈 fixture"
는 **`moai todo` 프로세스의 연결에 아무 영향을 주지 못한다.** 그 fixture 는 설정된 것처럼
읽히지만 실제로는 아무것도 제어하지 않는다 — 저자의 표현대로 *built* 이 아니라 *believed* 다.
그리고 그것은 **D4 가 닫으려던 공허 부류와 정확히 같은 부류**다: 성립하지 않은 Given 위에서
초록을 받는 것. iter-1 의 권고는 자기가 제기한 결함의 한 사례였다. 기록해 둔다.

(부수적으로, `moai todo` 는 단명 프로세스라 종료 시 마지막 연결이 닫히며 checkpoint 후 WAL 이
회수될 수 있다. `r1-repro.md` 대조군에서 `backlog.db` cksum 이 즉시 움직인 것과 정합한다.
`plan.md` §B 가 이를 **미측정 후보 설명**으로 정확히 표기했다 — 단정하지 않았다.)

**대체 Given 은 구성 가능한가 · 그렇다 (적어도 후보 (a)는).**

새 Given: "`backlog.db-wal` 이 존재하고 크기 > 0 **이며** `cksum backlog.db` 가 변경 전 값과
동일" — 둘 다 **감시를 걸기 전에** 관측.

- 후보 (a)(두 번째 연결로 열린 read snapshot 유지)는 원리상 성립한다. checkpoint 는 살아 있는
  독자의 mark 보다 앞선 WAL 프레임을 회수하지 못하고, 다른 연결이 DB 를 잡고 있는 동안에는
  WAL 파일이 삭제되지 않는다. 따라서 `moai todo` 의 종료 시 checkpoint 도 WAL 을 접지 못한다.
  칸반 레인이 동시에 큐를 잡는 실제 운영 형태와도 닮아 있어, 인공적 상태만은 아니다.
- 후보 (b)(직접 연결로 변경 + 연결 유지)는 성립하지만 `moai todo` 의 쓰기 경로를 우회한다 → D12.

핵심은 **어느 기법도 추론으로 인정되지 않는다**는 점이다: 기준문이 "the technique is not
credited on its reasoning" 라고 못박고, 위 두 관측값만이 성립 여부를 판정한다. 이것이 옳은
전도(轉倒)다 — "이렇게 하면 상태가 만들어진다"(믿을 수 있음)에서 "이 상태가 성립했음을
관측했다"(잴 수밖에 없음)로 옮겼다.

**성립 못 한 Given 이 정말 Gap 인가 · 그렇다 (4개 표면 일치).**

① AC-BJD-010 본문 "Establishing the Given is part of the criterion, not a preliminary … the
result is recorded as a **Gap** — never as a pass" · ② `plan.md` §G 신설 안티패턴 "Passing
AC-BJD-010 without its Given ever holding" · ③ `plan.md` M2 "do not report a pass whose Given
never held" · ④ §D.7 DoD "Any criterion that could not be met is recorded in `progress.md`
§E.2 as a Gap … never dropped". 네 곳이 같은 말을 한다. 우회 문구 없음. **D4 는 닫혔다.**

### 4. D3 — 세 표면 일치 · 바닥의 독립 구현 가능성 · 나중 확대의 무모순 · **셋 다 성립**

실측으로 세 표면 확인:

```
spec.md:132       - **REQ-BJD-002** (Event-driven) — **When** a `moai todo` **read** command runs
acceptance.md:24  When a `moai todo` **read** command is invoked against it,
plan.md:143       - Emit the disclosure from the `moai todo` **read** surface, on stderr, ...
```

세 곳 전부 "read". iter-1 의 3중 분열(invocation / read command / at minimum)은 해소됐다.

**바닥의 독립 구현 가능성**: 읽기 표면은 그 자체로 완결된 구현 단위다. 쓰기 동사에 대한
결정 없이도 M3 를 끝까지 수행할 수 있고, AC-BJD-002/003/004/005 전부 읽기 표면만으로 판정된다.
차단 없음.

**나중 확대가 모순을 만드는가**: 만들지 않는다. REQ-BJD-002 는 "read command 가 돌면 → 낸다"
형태라 확대는 **트리거 집합의 추가**이지 read 케이스의 부정이 아니다. AC-BJD-002 도 invocation
집합이 늘 뿐이다. AC-BJD-003(stdout 동일성)·AC-BJD-004(json 없으면 무발화)·AC-BJD-005(파일
무손상)는 breadth 와 직교한다. 어떤 기준도 재작성되지 않는다 — 저자의 주장대로다.

잔여: "read command" 가 어떤 동사 집합인지 열거되지 않았다. `plan.md` M3 가 "Report which read
verbs carry it" 로 run-phase 보고 의무를 지웠으므로 은폐되지는 않는다. 결함으로 올리지 않는다.

### 5. D2 — 조인 명령이 주장하는 바를 실제로 단언하는가 · **한다 (실행 확인)**

내가 iter-1 에 제시한 패턴이 기각된 것도 타당하다 — 그 패턴은 doc 주석과 호출부까지 9줄을
잡아 "정의가 한 곳"이라는 명제를 세우지 못한다. 저자의 조인 형태를 실행했다:

```
$ grep -c 'func Inspect.*Backlog.*\(Store\|Vouch\)' internal/kanban/*.go | grep -v ':0$'
internal/kanban/backlog_archive_vouch.go:1
        → 비영(非零) 파일 단 하나, 합계 1. 기대값과 일치.
          _test.go 는 호출만 하고 `func Inspect` 를 정의하지 않아 잡히지 않는다(확인).

$ grep -rn 'BacklogStore\(SQLite\|LegacyJSON\|None\) *=' internal/kanban/ | grep -v _test.go
internal/kanban/backlog_archive_vouch.go:30 · :32 · :34
        → 전부 그 파일. 사용부(`BacklogStoreNone}` · `case kanban.BacklogStoreLegacyJSON:`)는
          뒤에 `=` 가 없어 잡히지 않는다 — 정의만 세는 것이 맞다.

$ grep -n 'type BacklogArchiveVouch struct' -A 8 internal/kanban/backlog_archive_vouch.go
41:type BacklogArchiveVouch struct { … }  (exit=0)
        → macOS BSD grep 에서 피연산자 뒤 옵션 형태가 정상 동작. 이식성 문제 없음.
```

세 명령 모두 이 트리에서 실행되며 기술된 결과를 낸다. `grep -c` 가 다중 파일에서 `file:count`
형태로 출력한다는 점까지 기준문이 감안해 "summed count … and the sole non-zero file" 로 쓴 것도
정확하다.

한계는 남는다 → D11. 다만 "저장소 정체를 보고하는 함수가 유일하다"는 **전칭 부정을 grep 으로
완전히 기계화하는 것은 원리적으로 불가능**하다. 기준을 실행 가능·이진으로 만들라는 것이 D2 의
요구였고 그것은 충족됐다. 그 이상을 요구하는 것은 과잉이다.

### 6. clarification-gate 마커 미배치 · **판단이 옳다. 동의한다.**

MP-7 이 존재하는 이유는 **미해결 질문이 조용히 run-phase 로 흘러드는 것**을 막기 위해서다.
여기서는 세 조건이 모두 다르다: ① 바닥이 규범적이고 완결적이라 run 이 **완전히** 진행된다
② 미결 부분은 확대이지 미결정 전제가 아니다 ③ 은폐가 없다 — `spec.md` REQ-BJD-002 본문,
`plan.md` §D.1, `progress.md` §E.1 세 곳이 "escalated, unanswered" 를 명시한다. 마커는
"답 없이는 못 간다"는 뜻인데 갈 수 있으므로, 배치하면 근거 없는 hard-gate 가 된다.
Implementation Kickoff Approval 에서 운영자가 이 결정을 보게 되는 경로도 이미 열려 있다.

`plan.md` §D.1 의 괄호 문단(리터럴 토큰을 일부러 안 쓴 이유를 밝힌 부분)도 짚어둔다. 자동
검사의 리터럴 매칭을 피했다고 서술한 것은 형식상 회피처럼 보일 수 있으나, **그 회피 사실
자체를 문서에 적었고 미결 결정을 세 곳에 노출했으므로** 은폐가 아니다. 내 MP-7 검사가 실제로
리터럴 grep 이라는 서술도 사실이다. 투명하게 밝힌 오탐 회피로 판단한다 — 수용.

---

## Regression Check (iter-1 defect delta)

| # | iter-1 결함 | 판정 | 증거 |
|---|---|---|---|
| D1 | AC-BJD-008 재바인딩이 단일 `f=` 전제 | **RESOLVED** | "rebound only so that **the queue directory** its watch target **or targets** resolve against is the fixture's" + AC-BJD-010 다중 target 허용과의 상호무효화를 명시적으로 설명. |
| D2 | AC-BJD-007 기계 검증 불가 | **RESOLVED** | 3개 명령으로 대체, 전부 실행해 기대 결과 확인(위 §5). 잔여 한계는 D11 로 이월(원리적, optional). |
| D3 | 동사 범위 3중 불일치 | **RESOLVED** | 세 표면 전부 "read" 실측 확인. 미결 확대는 `plan.md` §D.1 에 공개 기록, 무모순 확인. |
| D4 | AC-BJD-010 Given 성립 기법 부재 | **RESOLVED** | Given 이 자기증거화(2개 관측값), 미성립 = Gap 이 4개 표면에서 일치. 내 원 권고는 기각이 옳음(위 §3). |
| D5 | AC-BJD-015 단일 regex 실명(失明) | **RESOLVED** | 2-grep 열거형, 미수리 트리에서 **둘 다 실패함을 실행으로 확인**(4≠1, 1≠0). |
| D6 | 사이트 4 vs 파일 4 · REQ-BJD-003 레이블 | **RESOLVED** | AC-BJD-016 "the **three** mirrored files" + 3파일 명시 열거. REQ-BJD-003 `(State-driven)/While` 실측 확인. |
| D7 | §A.2 의 R5 인용이 주장 미뒷받침 | **RESOLVED** | §A.2 가 2개 grep 을 직접 싣고, `reader-surfaces.md` 도 정정 블록 + 로컬↔템플릿 4행 표로 교체됨(원 결론은 유지, 귀속만 교체 — 올바른 처리). |
| D8 | "exactly one" 이 REQ-TAQ-013 줄과 경합 | **RESOLVED** | AC-BJD-002 에 archive-tables Given + "line **introduced by this SPEC**" 이중 한정, `plan.md` M3 에 fixture 조건 추가. AC-BJD-004 미전파는 D9 로 신규(optional). |

**8/8 RESOLVED. 정체(stagnation) 결함 0건 — 세 반복에 걸쳐 변하지 않은 결함 없음.**

---

## Defects Found (신규 · 전부 optional class)

**D9** — D8 의 수리가 형제 기준 AC-BJD-004 로 전파되지 않음 — **Severity: minor · Class: optional**
D8 은 AC-BJD-002 에 두 가지 방어를 넣었다: archive-tables Given 과 "line introduced by this
SPEC" 한정. AC-BJD-004 는 Given 을 독립적으로 재서술("Given a queue directory containing
`backlog.db` and no `backlog.json`")하므로 archive-tables 조건을 **상속하지 않고**, Then 도
"no **disclosure line** is emitted on either stream" 로 한정어가 없다. archive 테이블이 없는
fixture 에서는 기존 REQ-TAQ-013 줄이 발화하므로, D8 이 AC-BJD-002 에서 제거한 바로 그 모호성이
AC-BJD-004 에 남는다.
**Required fix**: AC-BJD-004 의 Given 에 "whose `backlog.db` has its archive tables present" 를
추가하고, Then 을 "no line introduced by this SPEC is emitted on either stream" 로 한정.

**D10** — AC-BJD-012 완화 조항이 부분문자열 판정이라 단언과 언급을 구분하지 못함 — **Severity: minor · Class: optional**
완화형은 `grep -v -e 'backlog\.db' -e 'export'` 로 생존 매치를 거른다. 그런데 이는 **줄에 그
토큰이 있는지**만 보므로, `` State lives at `.moai/state/todo/backlog.json`; see also
`backlog.db`. `` 같은 잘못된 수리가 필터에 걸려 사라지고 통과한다 — 그 줄은 여전히 JSON 을
상태 위치로 **단언**하는데도. 기준문 산문("What must not survive is the **assertion**")이
기계형과 어긋난다. AC-BJD-013 도 리터럴로 읽으면 그 줄이 `backlog.db` 를 포함하므로 못 잡는다
(사람이 "상태 위치 문장"으로 읽으면 잡힌다).
발생하려면 상당히 이상한 수리가 필요하므로 optional 로 둔다.
**Required fix**: 완화 조항에 "생존 매치는 그 줄의 **상태 위치 단언 대상**이 `backlog.db` 여야
한다"는 판독 규칙을 덧붙이거나, AC-BJD-013 의 Then 을 "각 사이트의 상태 위치 문장이 명명하는
경로가 `backlog.db` 로 끝날 것"으로 좁힌다.

**D11** — AC-BJD-007 의 전칭 부정이 명명 규약에 의존 — **Severity: minor · Class: optional**
`grep -c 'func Inspect.*Backlog.*\(Store\|Vouch\)'` 는 **그 이름 형태를 가진 함수가 하나**임을
단언하지, "저장소 정체를 보고하는 함수가 하나"임을 단언하지 않는다. `func ProbeQueueLayoutIdentity`
같은 이름의 두 번째 inspector 는 두 명령 어디에도 걸리지 않는다(2번 명령도 **같은 이름의 상수**
재정의만 잡는다).
이는 grep 의 원리적 한계이며 iter-1 대비 명백한 개선이므로 **수정 요구가 아니라 잔여 위험
기록**으로 둔다. 실질 방어는 세 번째 명령(새 사실이 `BacklogArchiveVouch` 의 **필드**여야 함)이
REQ-BJD-006 의 긍정 절반을 못박는 데서 나온다.
**Required fix (선택)**: `plan.md` §G 에 "두 번째 inspector 는 이름을 달리해도 REQ-BJD-006
위반" 한 줄을 추가해, 명령이 못 잡는 축을 리뷰 시야에 남긴다.

**D12** — AC-BJD-010 의 Given 에서 "`moai todo` 쓰기 경로" 한정이 사라짐 — **Severity: minor · Class: optional**
0.1.0 의 Given 은 "a fixture queue mutated **through the normal `moai todo` write path**" 였다.
0.2.0 은 "a mutation has been committed and not yet checkpointed, established by a technique the
run-phase names" 로, 쓰기 경로 한정이 **삭제**됐다. 후보 (b)(직접 연결로 변경)를 허용하기 위한
것으로 보이며, 감시가 누가 썼는지에 무관하므로 기준의 목적(감시가 지연된 쓰기를 보는가)에는
타당한 완화다. 다만 결과적으로 **`moai todo` 자신의 쓰기가 지연 상태를 만들 수 있는지는 더 이상
주장되지 않는다.**
연결된 미확정 가지: run-phase 가 "`moai todo` 쓰기는 종료 시 항상 WAL 을 접는다"를 측정하면
REQ-BJD-010 이 지키는 상태는 단독 프로세스 경로에서 도달 불가가 된다. 그 경우 REQ-BJD-010 을
규범으로 유지할지 여부를 어떤 산출물도 말하지 않는다(측정 **기록** 의무는 "the measured
checkpoint behaviour is recorded either way" 로 이미 있음). 동시 독자가 있는 칸반 운영 형태에서는
지연이 실재할 가능성이 높아 optional 로 둔다.
**Required fix (선택)**: AC-BJD-010 에 한 줄 — "지연이 `moai todo` 단독 경로로 도달 불가로
측정되면, 그 사실을 `progress.md` §E.2 에 적고 REQ-BJD-010 의 규범 유지 여부를 리드에게
올린다".

---

## 불변식 재확인 (문구 수정이 깨뜨렸는지)

- **REQ/AC 계수**: REQ 14 · AC 16 — 불변. 상한(16/16) 내.
- **Traceability**: §D.6 표 14행 불변. AC-BJD-001..016 전부 참조, 고아 0 · 미커버 REQ 0. 재계수 완료.
- **GEARS**: 14개 전수 재판독. REQ-BJD-003 만 변경(개선). 나머지 불변.
- **범위 규율**: §B 전체를 금지 축 4개(삭제/청소 · 엔진·마이그레이션·downgrade · 스키마 ·
  quarantine/State D) 어휘로 훑었다. 유일한 히트는 REQ-BJD-005 인데 그것은 그 행위들을
  **금지하는** 문장이다 — 드리프트가 아니라 가드. §C 4개 Out of Scope H3 불변, t359 위임 불변.
  iter-2 가 추가한 `plan.md` §G 항목 중 "**Editing `.moai/docs/todo-queue-storage.md:20` or
  `:55`**"는 오히려 범위를 **좁히는** 방향의 신설 가드다.
- **Tier M**: 불변. 접촉 파일·마일스톤 구조가 그대로이고, iter-2 는 문구만 바꿨다.
  `design.md`/`research.md` 요구할 아키텍처 표면은 여전히 없다.

---

## Recommendation

**PASS · 0.94.** Must-Pass 7/7, 집계가 Tier M 임계 0.80 을 크게 상회하며 iter-1 대비 단조 증가
(+0.09)라 점수 퇴행 STOP 조건에 해당하지 않는다. **blocking class 생존 결함 0건.** 신규 4건은
전부 optional·minor 이고, M6 재사용 절제 원칙상 이것들로 반복을 더 돌릴 이유가 없다 —
optional 결함의 목록 길이는 그 자체로 FAIL 근거가 되지 않는다.

**반복 상한에 대해 (리드 요청 사항).** Tier M ceiling 은 2 이고 이번이 2회차다. **PASS 이므로
상한은 무의미해졌다** — 운영자에게 "상한을 넘겨 반복할지" 물을 필요가 없고, 그런 결정 자체가
발생하지 않는다. 카드는 Implementation Kickoff Approval 로 바로 갈 수 있다. (FAIL 이었다면
PASS-with-debt / 범위 축소 / 명시적 override 3지 선택을 운영자에게 올려야 했을 것이다.)

D9~D12 는 run-phase 진입을 막지 않는다. 다만 **D9 는 한 문장이고 D8 수리의 누락된 형제**이므로,
킥오프 전 처리 비용이 사실상 0 이다. 나머지 셋은 run-phase 가 그 지점에 도달했을 때 판단해도
늦지 않다 — 특히 D12 는 run-phase 의 실측이 나와야 결정할 수 있는 성격이다.

**iter-1 감사 자체에 대한 기록 두 건.**

1. **D4 의 내 권고는 틀렸고, 틀린 방식이 D4 가 막으려던 공허과 같았다.**
   `PRAGMA wal_autocheckpoint=0` 은 연결 단위라 `moai todo` 의 연결에 닿지 않으며
   (`grep -rn 'wal_autocheckpoint' internal/` → 0건, `backlogDSN` 이 걸지 않음), 그 fixture 는
   구성된 것처럼 읽히지만 아무것도 제어하지 않는다. 저자의 기각과 리드의 판단이 옳다.
   결함을 지적하는 쪽도 같은 공허에 빠질 수 있다는 사례로 남긴다.
2. **AC-BJD-012 의 같은 구조적 공백을 iter-1 이 놓쳤다.** D5 를 템플릿 미러에서만 잡고 로컬
   `.claude/` 쪽 쌍둥이를 통과시켰다. 저자가 지적 없이 스스로 찾아 같은 2-grep 형태로 고쳤고,
   이번에 새 기준으로 감사해 비공허함을 실행으로 확인했다.

SPEC 의 강점은 iter-1 판단에서 달라지지 않았고 오히려 굳었다. 지적을 문구로 흡수하면서
**두 건은 근거를 들어 되받았고 그 되받음이 옳았다.** 감사자가 틀렸을 때 그것을 증거로
반박하는 산출물은, 지적을 전부 수용하는 산출물보다 신뢰도가 높다.
