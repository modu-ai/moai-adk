# t555 — `moai todo` 미등록 동사 + id 조합의 add 유출 (GH #1654)

- 카드: t555 · 클래스 B (plan 생략, run → sync)
- 브랜치: `WT-todo-verb-guard` · 워크트리 `.claude/worktrees/t555`
- base: 로컬 develop `d060e0d13`
- 레인: lane-7

---

## 1. Claim

1. **전수 판별 완료.** 가드가 어느 조합에서 발화하고 어느 조합에서 유출하는지를 12(첫 토큰) × 5(둘째 토큰) = 60칸 행렬로 실측했다. 유출의 뿌리는 가드의 **두 판별식이 각각 자신이 보호하는 표면보다 좁다**는 것이다.
2. **첫째 판별식(첫 토큰)이 좁았다.** `^[a-z][a-z-]{1,15}$` 는 소문자 철자 하나만 동사로 인정했다. 실측된 유출: `Show t401`·`SHOW t401`(대소문자), `show2 t401`(인접키 오타), `show_it t401`(구분자), `s t401`(약어 — 1글자 하한에 걸림), `supercalifragilis t401`(길이 상한). 여섯 형태 모두 **실재하는 카드 id 를 겨누고도** 조용히 카드가 됐다.
3. **둘째 판별식(카드 주소)이 좁았다.** 동사 계층 자신은 맨숫자를 카드 주소로 받는다 — `normalizeTodoRef` 가 `401` → `t401` 로 정규화한다(`internal/cli/todo.go`). 그런데 가드는 `t<n>` 만 주소로 봤다. 그래서 `show 401` 이 유출됐다. **가드가 자신이 지키는 표면의 주소 문법과 어긋나 있었다.**
4. **정크 카드 실사례는 운영 큐에 없다.** 76장 전수(export) 결과 해당 형태 0건.
5. **수리 후 유출 0.** 「미등록 첫 토큰 × 동사가 실제로 받는 주소」 조합 전부가 거부된다. 남은 통과 행은 **동사 계층 자신도 주소로 받지 않는** 형태뿐이다.

---

## 2. Evidence

### 2.1 정크 카드 전수 감사 (카드 [HARD] 항목)

`moai todo list --json` — 기본 20행 렌더가 아닌 전수 export.

```
{"last_seq": 595, "n": 76}
```

동사+id 형태 grep — 0건:

```
$ jq -r '.items[] | "\(.id)\t\(.state)\t\(.text[0:80])"' queue.json \
    | grep -Ei '^\S+\s+\S+\s+[a-z-]{2,16} t[0-9]+'
(출력 없음)
```

짧은 텍스트(정크 카드의 특징) 전수 — 0건:

```
$ jq -r '.items[] | "\(.id)\t\(.state)\t\(.text)"' queue.json \
    | awk -F'\t' 'length($3) < 60'
(출력 없음)
```

76장 모두 60자 이상의 서술형 카드다. **삭제 대상 없음** — 따라서 운영자 결재 사안도 발생하지 않는다.

부수 관측(이 카드 소관 아님, §5 참조): 운영 큐의 실체는 `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` 이며, 체크아웃 안의 `.moai/state/kanban/backlog.json`(1장, 2026-09-08)은 낡은 잔재다.

### 2.2 유출 조합 전수 행렬 — 수리 전

측정 도구: `internal/cli/todo_verb_leak_test.go`(초기에는 probe 형태), 격리 fixture `todoFixture` + `runTodo` 내부의 `liveTodoQueueRootReason` 가드 경유. **운영 큐에 도달할 수 없는 경로다.**

`REFUSED` = 가드 발화 · `LEAK` = 카드 생성 · `NO-CARD` = 등록된 동사가 서브커맨드로 라우팅됨

| 첫 토큰 \ 둘째 토큰 | `t401` | `401` | `T401` | `t401x` | `card` |
|---|---|---|---|---|---|
| `list` (등록) | NO-CARD† | NO-CARD† | NO-CARD† | NO-CARD† | NO-CARD† |
| `done` (등록) | NO-CARD† | NO-CARD† | NO-CARD† | NO-CARD† | NO-CARD† |
| `pr` (등록) | NO-CARD | NO-CARD | NO-CARD | NO-CARD | NO-CARD |
| `show` | REFUSED | **LEAK** | LEAK | LEAK | LEAK |
| `pick` | REFUSED | **LEAK** | LEAK | LEAK | LEAK |
| `un-pick` | REFUSED | **LEAK** | LEAK | LEAK | LEAK |
| `show2` | **LEAK** | **LEAK** | LEAK | LEAK | LEAK |
| `show_it` | **LEAK** | **LEAK** | LEAK | LEAK | LEAK |
| `Show` | **LEAK** | **LEAK** | LEAK | LEAK | LEAK |
| `SHOW` | **LEAK** | **LEAK** | LEAK | LEAK | LEAK |
| `s` | **LEAK** | **LEAK** | LEAK | LEAK | LEAK |
| `supercalifragilis` | **LEAK** | **LEAK** | LEAK | LEAK | LEAK |

† 동사 자신이 실패했을 뿐(빈 큐에 그런 카드 없음) 카드는 생기지 않았다 — 대조군의 요지가 이것이다.

굵은 칸(18개)이 이 카드가 닫은 유출이다. 굵지 않은 `LEAK` 는 §2.4의 이유로 의도적으로 카드로 남긴다.

맨숫자가 주소라는 근거는 대조군이 직접 낸다 — `done 401` 의 오류 메시지:

```
mutation refused: no backlog item t401
```

`401` 을 넣었는데 `t401` 을 찾았다. 반면 `done T401` 은 `no backlog item tT401` 을 냈다 — `T401` 은 동사도 주소로 받지 않는다.

### 2.3 유출 조합 전수 행렬 — 최종 수리 후 (재측정)

이 표는 **F1 재수리를 포함한 최종 코드**에서 다시 측정한 것이다. 1차 수리 직후에도 한 번 쟀지만, 그 값은 §2.5 의 재수리 이전 트리의 것이라 귀속이 어긋난다 — 재수리가 3토큰 이상에서만 발동하므로 2토큰 행이 바뀔 수 없다는 것은 **논증이지 측정이 아니었다.** 아래는 측정이다.

| 첫 토큰 \ 둘째 토큰 | `t401` | `401` | `T401` | `t401x` | `card` |
|---|---|---|---|---|---|
| `list`·`done` (등록) | REFUSED† | REFUSED† | REFUSED† | REFUSED† | REFUSED† |
| `pr` (등록) | NO-CARD | NO-CARD | NO-CARD | NO-CARD | NO-CARD |
| 미등록 9종 전부 | REFUSED | REFUSED | 카드 | 카드 | 카드 |

† 동사 자신의 오류(빈 큐에 그런 카드 없음)이지 fallthrough 가 아니다 — 카드는 생기지 않았다.

**「미등록 첫 토큰 × 동사가 받는 주소」 18칸이 전부 REFUSED 로 전환됐고, 대조군 15칸은 불변이며, 1차 측정과 셀 단위로 동일하다.**

**셀렉터 검증 (공허한 초록 방어)**: `-run` 패턴이 아무것도 매치하지 않으면 0초에 초록이 나온다. 그래서 서브테스트 통과 수를 세어 기대값과 대조했다 — 첫 토큰 12종 × 둘째 토큰 5종 = **60** 이 기대값이고, 실측 `--- PASS` 60건 / `FAIL` 0건 / `GOTEST_EXIT=0`.

**시간 값 부재 확인**: 경합 중 실행이므로, 이 행렬의 어느 칸도 시간 값을 재지 않는다는 것을 기계적으로 확인했다 — probe 파일과 그것이 부르는 두 헬퍼(`runTodo`·`todoFixture`)에 `time.`/`Deadline`/`Timeout`/`Sleep`/`Since` 적중 0건, 같은 패턴의 양성 대조는 1건 적중. 결과 값만 재므로 경합은 이 표를 느리게 할 뿐 뒤집지 못한다(§6).

### 2.4 남긴 통과 행의 근거

`T401`·`t401x`·`card` 는 **동사 계층 자신이 주소로 받지 않는다**(위 `done` 대조 측정). 이들까지 거부하면 가드가 자신이 지키는 표면보다 넓어져, 이번엔 반대 방향의 같은 종류 결함이 된다. 경계를 「동사가 공표한 주소 문법」에 정확히 맞춘 것이 이 수리의 판별식이다.

### 2.5 독립 감사가 잡은 잔여 결함 — arity 조건이 틀렸다

첫 수리는 맨숫자를 「invocation 의 나머지 전부일 때만」 주소로 봤다(`len(args) == 2`). 독립 감사(sync-auditor, 읽기 전용)가 **이 조건이 실재하는 verb 문법 셋을 통째로 빠뜨린다**는 것을 소스로 확정했다.

세 verb 는 **위치 인자를 2개** 받는다:

| verb | 시그니처 | 위치 |
|---|---|---|
| `relate` | `relate <a> <b>` · `ExactArgs(2)` · 양쪽 다 `normalizeTodoRef` | `internal/cli/todo_relate.go:36,38` |
| `drop` | `drop <n> <reason>` · `ExactArgs(2)` | `internal/cli/todo_drop.go:68,83` |
| `edit` | `edit <n> <text>` · `ExactArgs(2)` | `internal/cli/todo_edit_move.go:45,55` |

따라서 이들의 오타 호출은 **3토큰 이상**이 되어 arity 조건을 그대로 빠져나간다:

| 입력 | 1차 수리 후 | 같은 뜻의 `t<n>` 형태 |
|---|---|---|
| `relat 401 402` | 카드 생성 | `relat t401 t402` → 거부 |
| `drp 401 stale` | 카드 생성 | `drp t401 stale` → 거부 |
| `edt 401 new title` | 카드 생성 | `edt t401 new title` → 거부 |

감사는 arity 조건이 지키려던 반례에 실측 근거가 없다는 것도 함께 쟀다 — 라이브 큐 307건(items 76 + archived 230)에서 `<단어> <숫자> …` 형태 **0건**, `<단어> t<숫자>` 형태 **0건**(grep 은 양성 대조로 검증됨).

**수리**: arity 조건을 판별식으로 삼은 것이 잘못이었다. `drp 401 stale`(오타 verb)과 `fix 3 flaky tests`(카드)는 **모양이 같다** — 단어, 숫자, 단어. 어떤 모양 검사도 둘을 가르지 못한다. 가르는 것은 첫 토큰이다: `drp` 는 등록된 verb 에서 한 글자 거리이고 `fix` 는 아니다. 그래서 맨숫자는 두 경우에만 주소로 읽는다 — **나머지 전부이거나, 첫 토큰이 등록된 verb 의 near-miss 이거나**. 판별은 명령 트리에서 유도하므로(`todoVerbNames`) 나중에 verb 가 추가돼도 따로 손댈 곳이 없다.

부수로 감사의 F2 도 반영했다: `^\d+$` → `^[1-9][0-9]*$`. 큐는 선행 0 이 붙은 id 를 발급하지 않으므로(`fmt.Sprintf("t%d", …)`), 종전 정규식은 오히려 verb 문법보다 **넓었다**.

감사가 확인하고 깨끗하다고 판정한 것들: 다른 fallthrough 표면 없음(부모 명령 중 카드를 만드는 것은 `todo.go` 하나), 웹 콘솔은 읽기 전용이라 같은 결함이 존재할 표면 자체가 없음, `moai todo add` 직접 경로 무변경, 카드 ref 를 해석하는 verb 전수에서 주소 문법이 `t<n>`/`<n>` 두 형태뿐임, t69 fallthrough 보존, 템플릿 미러 의무 없음, 크로스플랫폼·SUBAGENT BOUNDARY 무영향.

### 2.6 뮤테이션 — 공허한 초록이 아님을 확인

다섯 뮤턴트를 넣고 새 테스트가 죽는지 실측했다.

| 뮤턴트 | 되돌린 것 | 죽은 하위 테스트 |
|---|---|---|
| A | 첫 토큰 판별식을 옛 `^[a-z][a-z-]{1,15}$` 로 | 7건 — `capitalized`·`uppercase`·`digit_typo`·`underscore`·`single_char`·`long_word`·`bare_number_capitalized` |
| B | 맨숫자 주소 인정을 비활성 | 2건 — `bare_number`·`bare_number_capitalized` |
| C | **과잉 확장**: 첫 토큰 `^.+$` + arity 조건 제거 | 3건 — `StillAddsNonAddresses/fix_3_flaky_tests`·`/재현_t401`, 그리고 **기존** `TestTodoNaturalLanguageCardsSurviveVerbGuard/fix_3_flaky_tests` |
| D | near-miss 갈래 제거(arity 만 남김) | 4건 — `near_miss_relate`·`near_miss_drop`·`near_miss_edit`·`near_miss_prefix` |
| E | **과잉 확장**: near-miss 를 항상 참으로 | 3건 — `StillAddsNonAddresses/fix_3_flaky_tests`·`/epic_7_planning`, 그리고 기존 t69 테스트 |

A·B·D 는 「막아야 할 것을 막는가」를, C·E 는 「막지 말아야 할 것을 막지 않는가」를 각각 반증 가능하게 만든다. C·E 가 기존 t69 테스트까지 죽인 것은 과잉 확장이 실재 위험이었다는 뜻이다.

D 를 처음 돌렸을 때 `near_miss_prefix`(`don 401`)만 살아남았다 — 2토큰이라 arity 갈래가 대신 잡았기 때문이다. 그 케이스를 3토큰(`don 401 now`)으로 바꿔 prefix 규칙 자체를 겨누게 한 뒤 네 건이 모두 죽었다. **뮤턴트 하나가 테스트 케이스 하나의 도달 불가를 드러낸 것**이며, 그대로 뒀으면 prefix 규칙은 무검증으로 남았을 것이다.

감사의 F4(항진 단언)도 함께 고쳤다: 토큰을 `strconv.Quote` 로 감싸 대조한다. 종전 `strings.Contains(combined, "s")` 는 어떤 영문 메시지에도 무조건 성립해 아무것도 단언하지 않았다.

### 2.6 게이트

```
go vet ./internal/cli/            → exit 0
GOOS=windows GOARCH=amd64 go build ./...  → exit 0
golangci-lint run ./internal/cli/...      → 0 issues, exit 0
gofmt -l (변경 2파일)                      → 출력 없음
go test ./internal/cli/ -run TestTodo      → ok  314.914s
go test ./internal/cli/... (전체)          → §6 참조
```

---

## 3. Baseline-attribution

- 트리: 워크트리 `.claude/worktrees/t555`, 브랜치 `WT-todo-verb-guard`, HEAD `d060e0d13`(로컬 develop 과 동일 — `git merge develop --ff-only` 로 fast-forward)
- 변경 파일 2개: `internal/cli/todo.go`(수정), `internal/cli/todo_verb_leak_test.go`(신규)
- 위 수치는 전부 이 트리에서, 이번 실행으로 측정했다. 다른 트리·다른 시점에서 가져온 값은 없다.
- 정크 카드 감사만 예외적으로 **운영 큐**를 읽었다(읽기 전용 export). 쓰기는 전혀 없었다.

---

## 4. Gaps — 관측하지 않은 것

- **동사 계층의 주소 문법 전수는 재지 않았다.** `normalizeTodoRef` 한 함수와 `done` 의 실동작으로 `t<n>`/`<n>` 두 형태를 확정했을 뿐, 각 동사가 그 함수를 실제로 경유하는지는 동사별로 세지 않았다.
- **행렬의 둘째 토큰 클래스는 5종이다.** `t0`(스토어가 거부하는 id), 음수, 매우 큰 수, 유니코드 숫자는 재지 않았다.
- **3토큰 이상은 전수가 아니다.** near-miss 갈래를 넣은 뒤 여섯 칸(`show t401 please`·`relat 401 402`·`drp 401 stale`·`edt 401 new title`·`don 401 now` 거부, `fix 3 flaky tests`·`epic 7 planning` 통과)을 쟀을 뿐, 60칸 행렬을 3토큰으로 다시 돌리지는 않았다.
- **near-miss 판별의 오탐률을 사전에 재지 않았다.** `fix`·`epic`·`재현` 세 반례만 확인했고, 등록된 14개 verb 각각에 대해 한 글자 거리 안에 드는 평범한 영단어의 집합은 열거하지 않았다.
- **커버리지를 측정하지 않았다**(`go test -cover` 미실행).
- **다른 표면은 보지 않았다.** 웹 콘솔·`moai todo add` 직접 경로는 이 카드의 범위 밖이다.
- **CI 판정은 아직 없다.** 로컬 초록은 조기 신호일 뿐이며, darwin/windows 매트릭스는 develop push 가 일으키는 실행이 낸다.

---

## 5. Residual-risk

1. **첫 토큰 확장이 실제 카드를 삼킬 수 있다.** 이제 `revert t401` 같은 「진짜로 그렇게 읽히는」 2단어 카드가 거부된다. 다만 이는 t203 이 이미 받아들인 거래를 넓힌 것이고, 탈출구 `moai todo add "<text>"` 는 그대로다. 비ASCII 첫 토큰은 산문으로 남겨 한국어 카드가 걸리지 않게 했다.
2. **2토큰 「단어 + 숫자」 카드가 거부된다.** `moai todo Windows 11`·`moai todo phase 2` 는 이제 오류다(arity 갈래). 감사 실측으로 307건 중 해당 형태 0건이었고, 에이전트 경로는 스킬 문서가 `moai todo add "<description>"` 을 명시적으로 지시하므로 자동화 회귀는 없다. 수용한 비용이다.
3. **near-miss 판별은 한 글자 거리 + 3자 이상 prefix 로 고정돼 있다.** 두 글자 이상 틀린 오타(`dorp` 는 잡히지만 `dropp` 는 prefix 로 잡히고, `drrp` 는 둘 다 아님)는 3토큰 이상일 때 여전히 카드가 된다. 임계를 2로 넓히면 평범한 단어를 verb 로 오인하기 시작하므로 넓히지 않았다.
4. **`internal/harness` 에 같은 성격의 비공개 edit-distance 구현이 이미 있다.** 패키지 경계를 넘어 export 하는 것이 이 한 번의 호출보다 넓은 변경이라 판단해 20줄짜리 bounded 버전을 지역에 뒀다. 두 곳이 갈라질 여지는 남는다.
5. **부수 관측 — 큐 파편화. 기존 카드 `t542` 소관이다.** 리드 확인 결과 `t542`("~/.moai/todo 의 orphan 큐 디렉터리 343개(11MB) 정리 — 유입은 이미 막혔고 남은 것은 잔재다", 2026-09-07 발행, queued)가 이미 있다. 이 카드가 보태는 것 둘은 t542 문안에 없다: (a) **운영 큐의 실체는 `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db`** 이며 sqlite 엔진 위에 있다 — 체크아웃 안 `.moai/state/kanban/backlog.json` 을 읽으면 운영 큐가 아닌 것을 읽는다, (b) 그 체크아웃 안 파일(1장, 2026-09-08) 자체도 **잔재 후보**다. **원인 미확립**을 그대로 남긴다 — 유입 경로를 모른 채 「잔재로 보인다」로 닫으면 t542 가 청소만 하고 재발을 막지 못한다.
6. **부수 관측 — `moai goal` 의 형제 결함. 카드 `t605` 로 발행됐다.** `internal/cli/goal.go` 의 `Args: cobra.ArbitraryArgs` + `RunE` 가 `len(args) > 0` 이면 곧장 arm 한다. 즉 `moai goal statuss` 는 오류가 아니라 `"statuss"` 라는 조건으로 goal 을 무장한다. **결과는 정크 카드보다 무겁다** — 정크 카드는 큐에 한 줄 남기고 끝나지만, 무장된 goal 은 Stop 훅이 매 턴엔드를 천장까지 차단하고 조건이 오타이므로 영원히 충족되지 않는다. t605 문안에 [HARD] 두 가지가 들어갔다(리드): t556 의 arm 시점 관문이 이 경로도 막는지 먼저 볼 것, 그리고 조건은 자유 텍스트이므로 「동사처럼 생긴 것」을 전부 거절하면 정당한 조건을 버린다는 과잉 수리 주의.

---

## 6. 최종 스위트 — 동시 실행 오염 기록

[HARD] **이 실행은 단독 실행이 아니었다.** 리드가 `ps` 부모 체인 추적으로 실측한 바(2026-09-10 14:35), 같은 머신에서 `internal/cli` 전체 스위트가 **둘** 동시에 돌고 있었다:

```
pid 40952  go test ./internal/cli/ -count=1 -timeout 20m   (경과 17:44)  ← lane-6
pid 69925  go test ./internal/cli/ -timeout 3600s          (경과 12:58)  ← lane-7 (이 카드)
```

load average 13~18. 경위는 직렬화 통지가 이 레인에 누락된 것이고, 리드가 자기 잘못으로 귀속했다. 리드 지시에 따라 실행은 죽이지 않고 끝까지 뒀다.

**판정 해석 규칙 — 비대칭이다.**

| 결과 | 유효한가 | 이유 |
|---|---|---|
| 초록 | **유효** | 경합은 타임아웃과 플레이크를 만들지, 거짓 통과를 만들지 않는다 |
| 빨강 | **무효 — 판정 없음** | 동시 실행 두 건이 있었다는 사실이 「내 변경 탓인가 부하 탓인가」를 가르지 못하게 한다 |

빨강이면 재측정 순번(lane-6 → lane-2 → lane-4 뒤)을 받아 다시 잰다. 이 절은 결과가 무엇이든 남긴다 — 초록이어도 단독 측정이 아니었다는 사실 자체가 기록될 값이다.

### 선행 실행 1건 (폐기)

이보다 앞서 `-timeout 900s` 로 돌린 실행이 906.522초에서 타임아웃 패닉으로 죽었다. 그 실행은 두 가지 이유로 근거가 못 된다: (1) 명령에 `| tail -20` 을 걸어 실패 상세를 잘라버려 원인을 귀속할 수 없고, (2) F1 재수리 **이전**에 컴파일된 코드를 쟀다. 폐기하고 다시 잰 것이 위 실행이다.

### 결과 — 빨강, 그러나 이 카드의 판정은 아니다

```
FAIL  github.com/modu-ai/moai-adk/internal/cli  1329.785s   GOTEST_EXIT=1
--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (8.55s)
--- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (132.76s)
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (3.64s)
--- FAIL: TestAuditLagUsesBinlagSeam (0.81s)
```

실패 4건이 지목하는 파일은 이 카드가 건드리지 않은 둘이다:

- `home_state_coverage_test.go:520` — `audited production file changed after coverage tip: internal/cli/launcher.go`
- `mcp_build_identity_test.go:643,648` — baseline hit `home_state_coverage.go:243`·`:251` MISSING, NEW ancestry hit `:245`·`:253`

이 카드의 변경은 `internal/cli/todo.go`(정규식 판별식 둘 + 헬퍼 둘)와 신규 테스트 파일뿐이다.

#### 대조군 — 「무관해 보인다」를 측정으로 바꾼다

무관해 보인다는 것은 귀속이 아니다. 그래서 **이 카드의 변경을 전부 걷어낸 트리**(`git show HEAD:internal/cli/todo.go` 로 원본 복원 + 신규 테스트 파일 제거, `git status` 로 변경 0 확인)에서 같은 4건을 돌렸다:

```
FAIL  github.com/modu-ai/moai-adk/internal/cli  144.302s   GOTEST_EXIT=1
--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (3.26s)
--- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (135.53s)
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (3.59s)
--- FAIL: TestAuditLagUsesBinlagSeam (0.92s)
```

**집합이 완주 실행의 실패 집합과 정확히 같다.** 변경이 0인 트리에서 같은 4건이 같은 메시지로 실패하므로, 이 4건은 develop tip 의 선재 적색이며 이 카드의 판정이 아니다.

#### 귀속의 독립 확인 4경로

같은 결론에 네 경로가 독립으로 도달했다(리드 취합):

| 경로 | 방법 | 성질 |
|---|---|---|
| lane-2 | 자기 수리를 되돌린 베이스라인에서 바이트 동일 메시지 재현 | 측정 |
| lane-6 | 지목 파일들이 develop↔HEAD blob 동일 | 유도 |
| lane-7 (이 카드) | 완주 실행의 실패 집합이 기지 적색 4건과 일치 | 측정 |
| lane-7 (이 카드) | 변경 0 트리에서 같은 4건 재현 | 측정 |

**원인**(리드 확인): 홈 상태 이전(t591/t592, PR #1699 `d8e9896e5`)이 `home_state_coverage.go` 를, 바이너리 경로 pin(t593, `6010d5d82`)이 `launcher.go` 를 마지막으로 바꿨고 둘 다 develop 조상이다. binlag 스윕의 baseline 이 그 줄 이동을 따라가지 못한 형태다. 소관은 **t600**.

#### 이 실행이 남긴 절차 결함 (리드 귀속)

대조군 실행은 lane-6 의 기준선 완주 실행과 겹쳤다. `RunsBoundedFocusedSuite` 가 내부에서 자식 `go test` 를 띄우는 시간 값 형태라 오염원이 된다. 리드가 절차 부재로 귀속했다 — 레인에게 「지금 슬롯이 누구 것인지」를 보여주는 표면이 없어, 통지만으로는 같은 일이 반복된다. 새 절차: **`internal/cli` 는 결과 값이든 시간 값이든 리드의 명시적 슬롯 승인 회신을 받고 시작한다**(알림만으로 시작하지 않는다). 다른 패키지는 종전대로.
