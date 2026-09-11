# Plan — SPEC-TODO-HOME-TEMP-GUARD-001

## §A 맥락

카드 **t536** (Class C · **Tier M** — 0.1.4에서 S→M). 선행 SPEC: `SPEC-STATE-ANCHOR-001` §5 — 미결 설계 결정을 표면화하고 결정하지 않은 채 닫았다. 본 SPEC이 그 미구현 절반이다. 작업 트리 `.claude/worktrees/t536`, 브랜치 `WT-home-fallback`, base `412c8cb14`.

### Tier 판정 — **M** (0.1.4, 리드 판정)

| 축 | 값 |
|---|---|
| REQ | **9** (REQ-THG-001..009) — Tier M 상한 16 |
| AC | 8 (AC-THG-001..008) — Tier M 상한 16, 불변 |
| 생산 파일 | `internal/kanban/todo_root.go` + 신규 판별식 파일 1 + CLI 안내 1행(`internal/cli/todo.go` 계열) |
| 영향 테스트 | **14건 / 4패키지**(§C.1 — 0.1.4에서 8→14) |
| 산출물 | spec.md · plan.md · acceptance.md (Tier M 3파일 — 이미 그 집합이다) |

- ~~**S 유지 근거**: 변경의 실질은 폴백 가지 한 곳에 술어 하나를 삽입하는 것이다.~~ [0.1.3까지의 논거 — 보존] 코드 변경의 크기로는 여전히 참이나, **Tier를 정하는 것은 코드 변경만이 아니다**: 요구사항 9개, 영향 테스트 14건 / 4패키지, 교차-SPEC 판정 2건이 붙는다. 마일스톤 3개·접촉 패키지 2개는 그대로다.
- **M으로 커질 신호**: 정규화 규칙을 공유 위치로 **추출**해야 한다는 판단이 서면 접촉 패키지가 3개(+`internal/cli` 리팩터)로 늘고, 그 시점에 Tier 재판정을 blocker로 보고한다. 본 계획의 기본 가정은 추출 없이 `internal/kanban` 안에 규칙을 두는 것이다.
- **산출물은 Tier M 집합과 이미 일치한다** — spec.md + plan.md + acceptance.md 3파일. S→M 전환으로 새로 만들 파일은 없다(`design.md` / `research.md`는 Tier L 소관이며 만들지 않는다). Tier S였을 때 `acceptance.md`를 별도로 둔 것은 카드 지시였고, 그 선택이 지금 전환 비용을 0으로 만들었다.
- **plan-audit PASS 임계가 `0.75` → `0.80`으로 올라간다**(Tier M). 이후 회차의 plan-audit 상한도 Tier M 규정을 따른다.
- **Tier 재판정 이력 — 0.1.3은 S를 유지했고, 0.1.4에서 리드가 뒤집었다.** ⑴ 0.1.3(3차 D15 파생): 판별식 주입 이음매가 요구사항으로 확정되면서 REQ가 9개가 될 수 있었고(Tier S 상한 8), 새 요구사항을 만드는 대신 **REQ-THG-002에 접었다**. 그때도 「접기를 거부하면 REQ 9 → Tier M」이라는 갈림길을 spec.md §8에 함께 적었다 — 조용히 9개를 싣지 않기 위해서였다. ⑵ **0.1.4(리드 판정 2026-09-08): 반론 채택 — 접힘 해제, `REQ-THG-009` 분리, Tier M.** 근거는 spec.md §8에 있고 요지는 하나다: 접힌 이음매는 **주어가 다른 요구사항의 종속절**이라 후속 편집이 **어떤 AC의 저항도 없이** 지울 수 있는데, 지워지면 가드가 시험 불가 상태로 출하된다. 관측 가능성을 보장하려고 존재하는 요구사항은 독립적으로 지목 가능해야 한다. 현재: **REQ 9 / AC 8 / Tier M**.

### 감사 점수 이력 — 하락은 산출물의 악화가 아니다

`0.75`(1차) → `0.875`(2차) → `0.8125`(3차) → 4차는 **부분 감사라 종합 점수를 산출하지 않았다**(D14·D16·D15 세 축 한정, 감사자 명시). 부분 감사 수치를 전수 감사 수치와 비교하는 것은 성립하지 않으므로 STOP 조항의 비교 대상도 아니다. 3차의 하락은 **측정 면적의 확대**다: 2차가 잰 모든 축에서 0.1.2는 개선되거나 유지됐고(3차 §1 전수 표), 하락은 2차가 **재지 않은 축**(가드가 기존 테스트에 미치는 영향 D15, 가드와 no-home 가지의 교차 D16, 뮤턴트 기제의 실행 반증 D14)을 3차가 처음 재면서 Completeness·Clarity가 내려간 결과다. **세 회차에 걸쳐 지속된 결함은 0건**이다 — 매 회차 지적이 실제로 닫혔고 신규 결함은 새로 측정된 면적에서 나왔다. 정체(stagnation) 신호가 아니다. **4차도 같은 모양이다**: D14·D16은 닫혔고, D15 축의 미닫힘(D18·D19)은 0.1.3이 **새로 연 표면**(전수 열거와 그 판정식)에서 나왔다 — 0.1.2 이전에는 그 표면이 존재하지 않았으므로 잴 수도 없었다. 다만 이 회차가 드러낸 것이 하나 더 있다: **두 결함 모두 판독 도구가 결론을 정한 사례였다**(spec.md §8 관측). 새 면적을 여는 것만으로는 부족하고 그것을 재는 도구도 함께 점검해야 한다.

### 개발 방법론

**TDD (RED-GREEN-REFACTOR)**. 기존 동작의 조건부 변경이고, 판별식은 순수 함수라 테스트가 싸다. 순서 구속은 §F에 있다.

## §B 알려진 결함 (이미 측정됨 — 재론 금지)

1. 비git base가 임시 디렉터일 때 `~/.moai/todo/<key>/`에 고아 큐가 생긴다 (`internal/kanban/todo_root.go:112` `homeTodoQueueRoot`).
2. 도달 가능성은 3갈래로 확인됐다 (spec.md §1.1) — 가드는 죽은 코드가 아니다.
3. `os.TempDir()`의 미해석 형태와 `EvalSymlinks`를 거친 호출자 경로는 철자가 다르다 (spec.md §4) — 순진한 접두사 비교는 조용히 실패한다.
4. 테스트 축 유입은 t422(`e7a078970`)가 이미 끊었다 — 본 카드는 생산 축만 닫는다.

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈추고 보고)

| 명령 | 기대값 |
|---|---|
| `git rev-parse --short HEAD` + `git branch --show-current` | `412c8cb14` 계열 + `WT-home-fallback` |
| `/usr/bin/grep -c 'ResolveTodoQueueRoot' internal/cli/todo.go internal/web/todo_queue_read.go` | 두 파일 모두 ≥1 (소비자 2곳 불변) |
| `go test ./internal/kanban/... -count=1` | ok (수리가 깨뜨리면 안 되는 baseline) |
| `go test ./internal/kanban/ -run TestResolveTodoQueueRoot_FallbackNoGit -count=1` | PASS — 아래 **영향 테스트 전수 표**의 1행이다. 종전 판본은 이 1건만 이름 불렀다(plan-audit 3차 D15) |
| `find ~/.moai/todo -maxdepth 1 -type d \| wc -l` | 344 (자기 자신 포함 — 343 + 1). run-phase 종료 시 **동일**해야 한다(REQ-THG-008) |
| `/usr/bin/grep -c 'func Resolve' internal/cli/todo.go` | **0** — CLI의 래퍼는 소문자 `resolveTodoQueueRoot`(`:71`)다. 0.1.3의 열거 grep이 이것을 못 본 것이 D18의 기제다. **값이 0이 아니면** 래퍼가 개명된 것이므로 §C.1 열거를 다시 돌린다 |
| `/usr/bin/grep -rlEi 'resolveTodoQueueRoot\|BacklogPathForRoot' internal/ --include='*.go' \| wc -l` | **23** (§C.1 축 1). 다르면 §C.1 표를 재유도한다 |
| `/usr/bin/grep -rlE 'newTodoCmd\|newTodoStore\|readTodoQueue\|todoBodyFor\|liveTodoQueueRootReason' internal/ --include='*_test.go' \| wc -l` | **10** (§C.1 축 2). 다르면 §C.1 표를 재유도한다 |

> **344의 귀속 (plan-audit 3차 Gap — 감사자가 재측정하지 못했다).** 워크트리 격리 가드가 실 HOME을 겨눈 `find`를 거절해 **3차 감사는 이 값을 재측정하지 못했다 — Gap이다.** 대신 리드가 primary 체크아웃에서 343을 실측했고(+1은 `find`가 세는 자기 자신), 그 인용이 이 값의 현재 근거다. run-phase는 §C 진입 시 이 행을 **자기 트리에서 다시 잰다** — 값이 다르면 멈추고 보고한다(fail-safe 방향: 그 사이 개수가 변했다면 §C가 진입에서 멈춘다).

> **[HARD] 마지막 행의 측정 자세**: 실제 HOME은 **개수만 세고** 쓰지 않는다. 모든 판정 측정은 canary HOME(`t.TempDir()` + `stubHome`) 안에서 한다. 측정 행위가 오염을 만들면 안 된다.

### §C.1 가드가 깨뜨리는 기존 테스트 — 전수 (plan-audit 3차 D15 → **4차 D18에서 재열거, 8건 → 14건**)

**측정 방법 — 두 축. 0.1.3의 방법은 한 축이었고, 그 축이 결론을 정했다.**

| 축 | 명령 (`/usr/bin/grep`) | 왜 필요한가 |
|---|---|---|
| 축 1 — 심볼, **대소문자 무관** | `/usr/bin/grep -rlEi 'resolveTodoQueueRoot\|BacklogPathForRoot' internal/ --include='*.go'` → **23 파일**(그중 테스트 15) | 0.1.3이 인용한 `/usr/bin/grep -rn 'ResolveTodoQueueRoot' … --include='*_test.go'`는 **대문자 `R`로 시작**하므로 CLI의 소문자 래퍼 `resolveTodoQueueRoot()`(`internal/cli/todo.go:71`)를 부르는 테스트를 **원리상 볼 수 없다**. 실측: `/usr/bin/grep -c 'func Resolve' internal/cli/todo.go` → **0** |
| 축 2 — 간접 소비자 경유 | `/usr/bin/grep -rlE 'newTodoCmd\|newTodoStore\|readTodoQueue\|todoBodyFor\|liveTodoQueueRootReason' internal/ --include='*_test.go'` → **10 파일** | 심볼을 전혀 적지 않고 명령/콘솔 진입점을 통해 리졸버에 도달하는 테스트가 있다. `internal/cli/todo_axisa_guard_test.go`는 **축 1에 전혀 잡히지 않는다** |

두 축의 합집합은 테스트 파일 **22건**이다. 그중 리졸버에 실제로 도달하는 것만 남기고, 도달하되 **비git base가 아니어서 가드에 닿지 않는 것**을 다시 뺀다.

**제외 기준(제외한 것도 적는다 — D20).**

| 제외 대상 | 제외 이유 |
|---|---|
| `internal/kanban/{backlog_counts,factory_slots,state_dir,state_dir_lock,todo_root_convention}_test.go`, `internal/statusline/{backlog_sqlite,landed}_test.go`, `internal/cli/{todo_export,todo_json_disclosure}_test.go` | `BacklogPathForRoot`를 **명시적 루트**에 대해 부를 뿐 리졸버를 지나지 않는다 |
| `internal/cli/{todo,todo_relate,todo_composed_upgrade,todo_flag_independence}_test.go` | 픽스처가 `initGitRepo(t, root)`로 **git 저장소**를 만든다 → `primaryCheckoutRoot`가 답하므로 가드에 도달하지 않는다 |
| `internal/cli/{todo_help_store,todo_surface,todo_lazy_landedref}_test.go` | `newTodoCmd()`를 **구조 검사**(Long 문자열·verb 목록·플래그)로만 쓰고 실행하지 않는다 |
| `internal/web/todo_queue_read_test.go` | 심볼 목록 테스트(문자열 `"kanban.ResolveTodoQueueRoot"`)이지 리졸버 호출이 아니다 |
| `internal/web/todo_section_test.go:81`·`:133`·`:155` | 비git 임시 루트로 리졸버를 지나지만 단언 대상이 **렌더 결과**라 가드 전후로 의미가 변하지 않는다 |

**A. 확실히 깨진다 (6건 — 0.1.3의 4건 + D18의 2건)** — base가 `t.TempDir()`(= 임시 기원 ⇒ 가드 발화)이고 홈 루트 또는 홈 오염을 단언한다.

| 테스트 | 좌표 | 축 | 깨지는 단언 | 처분 |
|---|---|---|---|---|
| `TestResolveTodoQueueRoot_FallbackNoGit` | `internal/kanban/todo_root_test.go:98` | 1 | 홈 루트 단언 실패 | **보존 이관** — 원 의도(키 유도 형태 단언)를 이음매(REQ-THG-009) 픽스처로 옮긴다 |
| `TestResolveTodoQueueRoot_PopulatedFallbackWins` | `internal/kanban/todo_root_test.go:196` | 1 | `got != fallbackRoot` — 가드가 `base`를 반환 | **보존 이관** — 읽기 우선순위(D-2) 단언이므로 비임시 픽스처에서 계속 밟는다 |
| `TestResolveTodoQueueRootAdopting_AdoptsLocalQueue` | `internal/kanban/todo_root_test.go:229`·`:233` | 1 | 반환 루트 = 홈 루트 실패 + 「로컬 파일이 사라졌다」 실패 | **보존 이관 (필수)** — `SPEC-WEB-TODO-QUEUE-001` **AC-WTQ-008** 산출 테스트(spec.md §8 판정: 철회 아님, 전제 정합) |
| `TestResolveTodoQueueRoot_FallbackNoGit` (CLI 미러) | `internal/cli/todo_queue_root_test.go:117-129` | 1 | 홈 루트 단언 실패. `dir := t.TempDir()` + `CLAUDE_PROJECT_DIR=dir` | **보존 이관** — **두 번째 패키지** |
| **`TestTodoQueue_FallbackAdoptsExistingLocalQueue`** (D18 신규) | `internal/cli/todo_queue_root_test.go:153` | 1 (소문자 래퍼 — 종전 grep 사각) | `root != want`(= `home/.moai/todo/<key>`) 실패. 이어서 이관된 3항목·상태·`last_seq=7` 단언까지 무너진다 | **보존 이관** — 주석이 스스로 `[HARD] verification 3 in code form` / "ADOPTED — never shadowed"라 적는 **adopt-not-shadow 단언**이다. A행 3번과 같은 부류이므로 같은 처분 |
| **`TestGuardBypassMutant_ObserveHomePollution`** (D18 신규) | `internal/cli/todo_axisa_guard_test.go:150` | 2 (심볼 없음 — 축 1에 전혀 안 잡힘) | 오염이 **없으면** `t.Fatalf`. 가드가 하는 일이 정확히 그 오염을 없애는 것이므로 확실히 깨진다 | **가지 이행 (보존도 갱신도 아닌 제3의 처분)** — `SPEC-STATE-ANCHOR-001` **AC-SA-011** 산출 테스트다. REQ-SA-011이 이미 **두 번째 가지**(「오염을 못 낸 뮤턴트는 그 사실과 가드 경계를 보고하라」)를 담고 있으므로, 가드 착지는 그 요구사항의 **위반이 아니라 문서화된 가지로의 이행**이다. 아래 D행 참조 |

**B. 배치 의존 → 이제 의도적 갱신 (1건)**

| 테스트 | 좌표 | 왜 | 처분 |
|---|---|---|---|
| `TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing` | `internal/kanban/todo_root_test.go:204-217` | `t.TempDir()` base + `HomeDirFn` 오류 = 「임시 기원 ∧ 홈 해석 불가」 교차를 정확히 밟는다 | **의도적 갱신** — REQ-THG-001이 교차의 반환을 `base`로 고정했으므로(D16), `dir/.moai/state/todo`가 아니라 `dir`를 단언하도록 갱신된다. 갱신은 **결정의 이행**이다. 이 갱신을 산출 AC로 덮는 것이 `acceptance.md` AC-THG-001 갈래 (c)다(D22) |

**C. 통과하지만 공허해진다 (7건 — 0.1.3의 5건 + D18의 2건, 4개 패키지)** — 가드가 `base`를 돌려주는 덕에 단언은 만족되지만, **그 테스트들이 쓰인 이유인 홈 폴백·adopt 가지를 더는 밟지 않는다.**

| 테스트 | 좌표 | 축 | 원래 밟던 가지 | 소관 AC |
|---|---|---|---|---|
| `TestResolveTodoQueueRoot_PureFallbackWritesNothing` | `internal/kanban/todo_root_test.go:125` | 1 | 순수 폴백의 무쓰기 | AC-WTQ-006 |
| `TestResolveTodoQueueRoot_ReadThroughToProjectLocal` | `internal/kanban/todo_root_test.go:156` | 1 | read-through(D-2) | AC-WTQ-007 |
| `TestAdoptionLandsWhereConsumersRead` | `internal/kanban/todo_root_contract_test.go:14` | 1 | adopt 목적지 경로 계약 | — (「카드가 조용히 사라진」 사고의 **회귀 가드**) |
| `TestAdoptingAndPureResolversAgreeWhenAdoptionFails` | `internal/kanban/todo_root_contract_test.go:40` | 1 | adopt 실패 시 두 리졸버 합치 | REQ-WTQ-005 |
| `TestTodoSectionReadsThroughToProjectLocalQueue` | `internal/web/todo_section_test.go:194` | 2 | 콘솔 측 read-through | AC-WTQ-007 (**세 번째 패키지**) |
| **`TestTodoQueueRootGuard_SilentOnHomeFallbackFixture`** (D18 신규) | `internal/cli/todo_queue_root_test.go:298` | 2 | 그 이름 그대로 **홈 폴백 루트**에서 축-A 가드가 침묵함 | — (`queueRootInsideTemp`의 두-철자 처리 가드) |
| **`TestAxisACanaryHomeSweep_TodoFamily`** (D18 신규 — 감사자 제안) | `internal/cli/todo_axisa_guard_test.go:69` | 2 | canary HOME 0-오염이 `runTodo` 게이트의 증거 | AC-SA-010 (`MOAI_AXIS_A_CANARY_SWEEP=1` 게이트) |

> **C행 6번의 공허화 기제 (판독 근거).** 그 테스트는 비git `t.TempDir()` + 스텁 홈에서 `liveTodoQueueRootReason() == ""`를 단언한다. 그 함수(`todo_queue_root_test.go:226-233`)는 `queueRootInsideTemp(resolveTodoQueueRoot())`가 참이면 `""`를 돌려준다. 가드 착지 후 반환값은 `dir`(= `t.TempDir()` 아래)이므로 여전히 참 → **PASS는 유지되지만, 그 테스트가 겨눈 「홈 폴백 루트」라는 대상이 픽스처에서 사라진다.**
>
> **C행 7번**은 기본 실행에서 `t.Skip`되지만(환경변수 게이트), 게이트를 켜면 0-오염이 **가드에 의해 구조적으로 보장**되므로 `runTodo` 게이트의 증거이기를 그친다.

> **`queueRootInsideTemp`는 M1 판별식의 선례다 (부수 관측, 요구사항 아님).** `internal/cli/todo_queue_root_test.go:236-253`이 이미 **같은 규칙**을 구현한다: 양쪽에 `EvalSymlinks`(존재 시), `filepath.Rel` 기반 구성요소 포함, 형제 접두사 면역(`/tmp/x` vs `/tmp/xy`). REQ-THG-003이 고정한 규칙과 같다. 다만 그것은 `_test.go` 안의 비노출 심볼이라 임포트할 수 없으므로 **재사용이 아니라 선례로만** 쓴다 — 이로써 같은 규칙의 구현이 트리에 셋(launcher `resolveSymlinks` · 이것 · M1)이 되는데, 통합 여부는 본 카드의 범위가 아니다(spec.md §8 마지막 항목과 같은 처분).

**D. 교차-SPEC 처분 — `SPEC-STATE-ANCHOR-001` AC-SA-011 (A행 6번의 상세)**

- **판정은 spec.md §8이 소유한다** — 네 조각(§5의 위임 조항 · §5:115의 「어느 REQ에도 영향 없음」 · REQ-SA-011의 두 번째 가지 · §5:115가 예고한 후속 카드)이 같은 방향을 가리킨다. 철회가 아니다.
- **그 SPEC의 산출물은 한 글자도 수정하지 않는다.** `status: completed`도 그대로다. 그 SPEC `acceptance.md:23`의 증거 칸("뮤턴트 오염은 지금도 재현 가능 — 그것이 증거")은 **t536 이전 트리에 시각 고정된** 기재이며, 그 사실을 본 카드 문서가 적는 것으로 처분이 끝난다.
- **테스트의 처분**: 가지 1(오염 관측)에서 **가지 2(오염 부재를 발견으로 보고)**로 옮긴다. `t.Fatalf`가 죽는 대신, 오염 부재를 관측하고 **경계 보고를 산출물로 남기도록** 갱신한다. 그 산출물의 명세는 §F M3가 정한다.
- **이 갱신은 요구사항이 지시한 것이다** — 픽스처 편의가 아니다. 그 구분을 §E.2 인용에서 명시한다.

**전수 주장의 귀속 (D20 — 「전수」는 방법으로 재유도되어야 한다).** 위 14건(A 6 · B 1 · C 7)은 각 행의 「축」 칸이 어느 명령에서 나왔는지를 밝히고, 제외한 후보는 제외 기준 표에 있다. 0.1.3의 「8건」은 인용된 방법으로 재유도되지 않았다 — 그 grep이 못 보는 항목이 목록에 없었고(축 1의 대소문자), 동시에 그 grep의 산출물이 아닌 항목(`todo_section_test.go:194`)이 실려 있었다. **결과가 방법보다 좁으면서 동시에 넓었다.**

**실행으로 확인하지 않았다** — 가드가 아직 없으므로 원리상 불가능하다. 판독 유도다. 구현이 술어를 더 좁은 지점에 넣으면 파급 집합이 줄 수 있고, 그 경우 **줄었다는 사실을 §E.2에 기록한다**.

**[HARD] 공허화는 무파괴 기준으로 원리상 보이지 않는다.** 공허하게 통과하는 테스트는 **실패하지 않는다** — 그래서 「미갱신 테스트 무실패」는 C행 7건에 대해 아무것도 주장하지 못한다. 볼 수 있는 판정식은 §F M2가 정한다(0.1.3의 판정식은 D19에서 폐기됐다).

## §D 구속 조건 (재논의 금지)

- **D1 — 범위는 임시-디렉터 기원 한정이다.** 「모든 비git base 거부」는 기각된 판독이다(spec.md §3). 이 경계를 넓히는 구현은 요구사항 위반이다(REQ-THG-005).
- **D2 — 정규화 규칙은 고정이다.** 존재하면 `EvalSymlinks`, 없으면 어휘적 `Clean`, **양쪽 모두에** 적용(REQ-THG-003). 근거는 `internal/cli/launcher.go:456-485`의 문서화된 GOOS 결정성이며, 이 규칙을 바꾸는 것은 요구사항 변경이다.
- **D3 — 비교는 경로 구성요소 단위다.** 생문자열 접두사 금지(REQ-THG-002).
- **D4 — 불확실은 fail-open이다.** 정규화 실패·판별 불가 → "임시 아님"(REQ-THG-004). 거짓 양성이 거짓 음성보다 비싸다는 비대칭이 근거다.
- **D5 — 순수성 불변.** `ResolveTodoQueueRoot`은 어느 가지에서도 쓰지 않는다. 가드 도입이 콘솔 경로에 쓰기·에러·로그를 만들면 실패다(REQ-THG-007).
- **D6 — 키 유도 불변.** `TodoQueueProjectKey`를 건드리지 않는다 — 바꾸면 기존 홈 큐들이 새 키를 얻어 도달 불가가 된다.
- **D7 — 정리 금지.** `~/.moai/todo/` 아래 어떤 것도 삭제·이동하지 않는다(t542 소관, REQ-THG-008).
- **D8 — 접촉 패키지는 `internal/kanban` + `internal/cli`(안내 1행)뿐이다.** `internal/statusline`·`internal/config`·`internal/hook`·`internal/session`·`internal/stateanchor` 변경 금지.
- **D9 — 부재 가드의 판별은 뮤턴트다.** 「임시 기원에서 홈 디렉터 0개」는 가드 없이도 관측될 수 있다. 판별식을 무력화한 뮤턴트가 오염을 재현하는 것(AC-THG-007)이 채택 요건이며, **못 잡은 뮤턴트도 보고한다** — 그것이 가드의 경계다.
- **D10 — 테스트 격리는 `t.TempDir()` + `stubHome`뿐이다.** 개발자 실제 HOME에 쓰는 테스트는 없다.
- **D11 — 로컬 검증 범위는 패키지 단위다.** `internal/kanban`, `internal/cli`(todo 서브셋), `internal/web`. `go test ./...` 로컬 전체 실행 금지 — 전 패키지·리눅스 판정은 CI 몫이다.
- **D12 — 거부 시의 대체 루트는 `base`다(미결 아님).** `ResolveTodoQueueRoot`은 `string` 하나를 반드시 돌려주므로(`todo_root.go:64`) 가드가 발화해도 무언가는 반환된다. 그 무언가는 **launch base 자체**다 — `fallbackTodoQueueRoot`의 read-through가 이미 `:148`에서 `return base`로 돌려주는 값이고, 파일 머리말(`:22-24`)이 그것을 "PROJECT-LOCAL root"라 부른다(REQ-THG-001).
  - **판정 규칙은 값이 아니라 성질이다**: 대체 루트는 `BacklogPathForRoot(<반환값>)`이 **프로젝트의 실제 로컬 큐 파일을 가리키는** 루트여야 한다. 구현이 이 성질을 지키면 판정 통과이고, 값을 어떻게 유도하든 계층 오류가 요구사항 수준에서 표현되지 않는다.
  - **[HARD] 술어의 배치 순서도 구속된다 (plan-audit 3차 D16).** 「임시 기원」과 「홈 해석 불가」는 **배타적이지 않다** — 비git 임시 base에서 `HomeDirFn()`이 오류를 내면 둘 다 성립한다. 그 교차에서 **`base`가 이긴다**(REQ-THG-001). 귀결: 판별식은 `fallbackTodoQueueRoot`/`homeTodoQueueRoot`의 **`ok` 검사보다 먼저** 평가되어야 한다. 뒤에 두면 반환값이 `:121`의 `resolveStateDir(base,false)` = `base/.moai/state/todo`가 되어 — **본 카드가 본뜨지 않기로 한 바로 그 계층 어긋난 형태를 배치 순서만으로 다시 집어 든다.** 이것은 스타일이 아니라 결과를 바꾸는 선택이므로 구현자가 고르지 않는다. 관측 지점: `internal/kanban/todo_root_test.go:204-217`(§C.1 B행, 의도적 갱신).
  - `""` 반환은 **금지**다: 콘솔(`internal/web/todo_queue_read.go:33-38`)은 반환값을 `vm.Root`에 그대로 싣고 `BacklogPathForRoot(root)`로 읽으므로, 빈 문자열은 렌더에 빈 루트를 표시하고 프로세스 CWD 상대의 엉뚱한 경로를 읽는다 — 거부가 아니라 **조용한 오작동**이다.
  - **정정 기록 (plan-audit 2차 D9 — 지적이 옳다. 종전 문장을 지우지 않고 여기 남긴다).** 0.1.1 판본은 대체 루트를 "project-local **state directory**" = `resolveStateDir(base, false)`(`todo_root.go:121`의 형태)로 고정하고, `base`는 「덜 정밀하다」며 **기각했다. 그 판단은 뒤집혔다.** `resolveStateDir(base,false)`는 이미 `base/.moai/state/todo`이고 소비자는 반환값에 `BacklogPathForRoot`를 **덧붙이므로**(콘솔 `internal/web/todo_queue_read.go:33-35`, CLI `internal/cli/todo.go:57`·`:72`), 그 값을 루트로 돌려주면 읽히는 경로가 `base/.moai/state/todo/.moai/state/todo/backlog.json`이 된다 — 아무도 쓰지 않는 경로다. 운영자의 실제 로컬 큐는 `base/.moai/state/todo/backlog.json`에 있고 그것이 AC-THG-001 (b)가 세우는 픽스처다. 즉 명령과 콘솔이 **존재하는 큐를 못 보고 빈 큐를 렌더한다** — 바로 이 D12가 `""`를 금지하며 든 근거("거부가 아니라 조용한 오작동")와 **같은 실패 부류**이며, 그 논거가 종전 선택도 함께 반박한다. 「덜 정밀」이라는 기각 사유는 계층을 거꾸로 읽은 것이다: 계층이 맞는 값은 `base`이고, 종전 선택은 이 파일에서 **유일하게 계층이 어긋난 반환**(`:121`)을 본떴다. 그 어긋남 자체는 기존 코드의 것이며 **본 카드에서 고치지 않는다**(spec.md §8 관측 항목 — 트리거가 다르고 별도 카드 후보다).
- **D14 — 임시 루트 집합은 주입 이음매를 통해 도달한다(요구사항 — 0.1.4에서 REQ-THG-002에서 분리된 **REQ-THG-009**).** 분리 근거는 spec.md §8의 리드 판정(2026-09-08)이며, 이음매가 **독립적으로 지목 가능한 요구사항**이어야 후속 편집이 그것을 조용히 지울 수 없다. 이음매는 `internal/cli`·`internal/web` 테스트에서도 닿아야 하므로 **exported**다. 하드코딩된 리터럴 집합으로 끝내지 않는다. 근거는 spec.md §8의 판정 기록이며 요약하면 하나다: 이 리포의 [HARD] 격리 규율(`t.TempDir()`) 아래에서는 이음매가 없으면 **비임시 비git base 픽스처를 만들 수 없고**, 그러면 AC-THG-003이 판정 불가가 되고 §C.1 C행 5건의 비임시 사본도 만들 수 없다. 선례는 같은 파일의 `HomeDirFn`(`todo_root.go:39-46`) + `stubHome`(`todo_root_test.go:55-61`)이다. **이음매를 생략하고 「테스트에서 비임시 디렉터를 쓰면 된다」로 대체하는 것은 요구사항 위반이다** — 그 디렉터를 만들 인가된 수단이 없다.
- **D15 — 영향 집합은 §C.1의 14건 전수다(0.1.4에서 8→14, D18).** M2가 「깨진 것만 고친다」로 좁히지 않는다. 특히 §C.1 C행 7건(공허화)은 **실패하지 않으므로 무파괴 검사로 보이지 않는다** — M2 종료 조건이 그것을 따로 본다(§F). 두 형제 SPEC의 산출물은 **어느 쪽도 수정하지 않는다**: `SPEC-WEB-TODO-QUEUE-001`의 AC 산출 테스트 3건(AC-WTQ-006/007/008)은 **보존**, `SPEC-STATE-ANCHOR-001`의 AC-SA-011 산출 테스트는 **REQ-SA-011 두 번째 가지로의 이행**(§C.1 D행)이다.
- **D13 — `/var/tmp`는 루트 집합에 넣지 않는다.** 근거와 재검토 조건은 spec.md §8. 「이왕이면 넓게」로 추가하지 않는다 — 거짓 양성 사전확률이 가장 높은 루트이고, 측정된 오염 기원도 아니다.

### 결정된 항목 (종전 미결 — run-phase가 고르지 않는다)

- **U1 — 거부 시 CLI의 형태: (b) 안내 후 대체 루트(`base`, D12)로 계속한다. 결정됨 — 운영자 선택 2026-09-07(리드 경유).** run-phase는 이 형태를 구현할 뿐이며 다시 고르지 않는다. 종전에 이 자리에 걸려 있던 blocker 보고 의무는 **해제된다**(그 의무는 미결이었기 때문에 있었다).
  - **선택 근거(운영자 제시)**: ⑴ 오염 방지는 홈을 건드리지 않는 것으로 이미 달성되므로 비영 종료(a)가 추가로 막는 것이 없다. ⑵ (b)가 D12의 대체 루트 결정과 정합한다. ⑶ 콘솔은 종료 코드가 없고 어느 쪽이든 루트를 받으므로, (a)는 **같은 거부에 대해 CLI와 콘솔의 행동을 갈라놓는다** — 소비자 둘이 서로 다른 것을 본다. ⑷ 임시 디렉터에서 도는 기존 테스트·스크립트가 (b)에서 계속 동작한다.
  - **U1과 대체 루트는 별개 축이었고, 지금은 같은 값에서 만난다.** 순수 경로(콘솔)가 받는 루트와 명령 경로가 계속 쓰는 루트가 같다.
  - **두-형태 표현의 가지치기**: REQ-THG-006과 AC-THG-005는 0.1.1에서 (a)·(b) 양쪽에 지시대상을 갖도록 쓰였다. 그 표현은 **U1이 열려 있던 동안 의도적이었고 옳았다** — 지금 결정이 내려져 불필요해졌으므로 결정된 형태로 좁힌다. 틀려서 지우는 것이 아니다.
  - **종전 판본의 과장 정정(보존)**: 이 자리에 있던 "AC-THG-005는 두 형태 모두에서 판정 가능하게 쓰여 있다(홈 디렉터 0 + 안내 관측)"는 주장은 그 시점 AC에 대해 참이 아니었다 — Then절은 3항인데 GREEN 주석만 2항으로 축소해 중립성을 만들었다. 0.1.1이 Then절 자체를 고쳤고, 0.1.2가 결정된 형태로 좁혔다.

## §E 자가 검증

마일스톤 종료마다 §C 표의 해당 행과 acceptance.md §D.1의 판정 명령을 재측정하고 **실제 출력을 그대로** `progress.md` §E.2에 인용한다. 요약 문장("전부 통과")은 증거가 아니다. 각 항목은 명령 + 관측 출력 + baseline attribution `(this run, this tree)` + HEAD SHA의 사인조를 갖춘다.

## §F 마일스톤

### 순서 구속 — 가장 바뀔 결정이 앞이다

판별식의 **모양**(정규화 규칙·루트 집합·비교 방식·fail 방향)이 이 카드에서 가장 바뀔 가능성이 큰 결정이다. 따라서 M1이 판별식을 순수 함수로 먼저 못박고, 배선(M2)과 뮤턴트·배치 검증(M3)이 뒤따른다. 배선을 먼저 하면 판별식 수정이 매번 배선을 흔든다.

### M1 — 임시-기원 판별식 (순수 함수 + 테스트)

- `internal/kanban`에 판별식을 둔다. 제안 형태: `func TempOriginReason(base string) (reason string, isTemp bool)` — 분류 결과와 **매치된 임시 루트**를 함께 돌려준다(안내 문구가 이유를 이름 부를 수 있어야 하므로 bool만으로는 부족하다).
- 정규화: 존재→`EvalSymlinks`, 비존재→`Clean`, 양쪽 적용(D2). 임시 루트 집합: `os.TempDir()`, `/tmp`, `/var/folders`(D3 비교).
- **임시 루트 집합은 패키지 수준 주입 이음매를 통해 도달한다**(**REQ-THG-009**, D14 — 0.1.4에서 REQ-THG-002로부터 분리). 형태는 `HomeDirFn`의 것과 같은 급이면 된다 — 패키지 수준 함수 변수 + 테스트 헬퍼. **이것이 M1에 있는 이유**: 이 이음매가 없으면 M2의 AC-THG-003 픽스처도, §C.1 C행 **7건**의 비임시 사본도 만들 수 없고, C행 판정식(`TempOriginReason(base)`의 직접 단언, D19)도 세울 수 없다. **`internal/cli`·`internal/web` 테스트에서도 닿아야 하므로 exported다**(REQ-THG-009). 판별식의 모양과 함께 가장 먼저 못박는다(§F 순서 구속).
- RED→GREEN: AC-THG-002(두 철자 등가), AC-THG-006(`/tmpfoo` 경계), AC-THG-004(정규화 실패 시 fail-open).
- **부작용 0**: 판별식은 `Stat`/`Lstat` 외의 파일시스템 접촉을 하지 않는다.

**M1 종료 조건**: AC-THG-002·004·006 GREEN + **임시 루트 집합 이음매(REQ-THG-009)가 테스트에서 실제로 스텁된 출력**(그것 없이는 M2의 픽스처가 서지 않는다) + **이음매가 호출 시점에 읽힌다는 증거**(스텁을 건 뒤 `TempOriginReason` 판정이 바뀌는 것 — init 시점에 값으로 복사하는 구현이면 스텁이 조용히 무시되고, 그 형태가 M2 C행 판정식을 통째로 공허하게 만든다). `go test ./internal/kanban/... -count=1` 무파괴(M1은 아직 배선하지 않으므로 §C.1의 8건 어느 것도 접촉하지 않는다).

### M2 — 폴백 가지 배선 + 안내

- `homeTodoQueueRoot`/`fallbackTodoQueueRoot` 가지에 술어를 삽입한다. 임시 기원이면 홈 루트를 반환하지도 만들지도 않고, **`base`를 대신 반환한다**(REQ-THG-001, D12). 판정 규칙은 성질이다 — `BacklogPathForRoot(반환값)`이 프로젝트의 실제 로컬 큐를 가리켜야 한다. **`resolveStateDir(base, false)`(`todo_root.go:121`의 형태)를 본뜨지 않는다**: 그 값은 이미 상태 디렉터라 소비자가 한 계층을 더 붙인다(D12 정정 기록).
- `ResolveTodoQueueRootAdopting`: 임시 기원이면 `adoptLocalTodoQueue`를 **호출하지 않는다**(홈 target 아래로 `MkdirAll`이 도는 유일한 경로 — 단 그 `MkdirAll`은 base에 로컬 `backlog.json`이 있을 때만 도달한다, `todo_root.go:172-175`).
- CLI 안내(REQ-THG-006): 매치된 임시 루트와 **계속 쓰는 대체 루트(`base`)를 이름 부르고**, 실행은 계속한다(비영 종료 아님). **U1은 §D에서 이미 (b)로 결정됐다** — M2는 구현할 뿐 고르지 않고, 종전의 blocker 보고 의무는 해제됐다.
- 순수 경로는 조용하다(D5): 콘솔은 안내를 내지 않고, 쓰지 않으며, 에러도 내지 않는다.
- RED→GREEN: AC-THG-001, AC-THG-003(비임시 비git 유지 — 범위의 정확성), AC-THG-005.
- **무파괴 판정 — 「미갱신 테스트 무실패」는 폐기한다 (plan-audit 3차 D15).** 그 판정식은 §C.1이 재기 전의 것이고, **C행 5건을 원리상 볼 수 없다**(공허하게 통과하는 테스트는 실패하지 않는다). 대신 §C.1의 세 갈래에 각각 다른 판정식을 둔다:
  - **A행 6건 (확실히 깨짐) — 5건은 보존 이관, 1건은 가지 이행.** 6번째(`TestGuardBypassMutant_ObserveHomePollution`)는 보존도 갱신도 아닌 **§C.1 D행의 가지 이행**이며 그 산출물은 M3의 경계 보고다. 나머지 5건 — **보존 이관.** 원 단언을 **지우지 않고** D14 이음매로 만든 비임시 비git 픽스처로 옮긴다. 판정식: 옮긴 테스트가 PASS하고, 그 단언이 **여전히 홈 루트를 단언한다**(옮긴 뒤에도 홈 폴백 가지를 밟는다는 뜻). 단언을 새 동작에 맞춰 고쳐 쓰는 것은 **금지** — 그것이 AC-WTQ-008을 조용히 철회하는 경로다(D15).
  - **B행 1건 (`HomeUnresolvableWritesNothing`) — 의도적 갱신.** REQ-THG-001의 교차 결정(D12 순서 구속)을 반영해 `dir`를 단언하도록 고친다. **갱신 사실과 근거를 §E.2에 인용한다** — 결정의 이행이지 편의가 아니다.
  - **C행 6건 (공허화) — 비임시 사본으로 원 가지를 계속 밟게 한다.** 판정식은 무실패가 **아니다**. 원본은 **그대로 둔다**(가드 아래 동작의 회귀 가드로 계속 유효하다).
  - **C행 7행(`AxisACanaryHomeSweep_TodoFamily`) — 수용된 손실.** 사본을 구성할 수 없다(프로세스 경계). 처분 전문과 그 근거는 §C.1 D19 표 아래 **「7행 — 수용된 손실」** 절에 있다. 공허화하는 것은 여전히 **7건**이고(그래서 :122·:143의 계수는 7이 맞다), 사본으로 처분되는 것이 **6건**이다 — 두 수는 서로 다른 것을 센다.
    - **판정식 (0.1.4, D19)**: 각 비임시 사본은 자기 base에 대해 **판별식의 판정을 직접 단언한다** — `TempOriginReason(base)`가 `isTemp == false`를 보고한다(REQ-THG-009의 이음매로 임시 루트 집합을 스텁한 상태에서). **원 단언은 지우지 않고 그대로 병기한다.** 그러면 판정은 「사본이 PASS **하고** 그 사본 안의 판별식 단언이 비임시를 보고했다」가 된다.
    - **왜 관측 대상을 반환값에서 판별식으로 옮겼는가.** 0.1.3의 최소 형태(「사본의 단언 대상이 홈 루트(또는 adopt 목적지)인 것 — PASS 자체가 그 가지를 밟았다는 증거」)는 **7건 중 1건에만 성립한다**. 건별 실측(plan-audit 4차 D19가 본문 전수 판독으로 산출, 본 판본이 재확인):

      | C행 사본 | 원 단언 대상 (실측) | 0.1.3 판정식 | 0.1.4 판정식 |
      |---|---|---|---|
      | `PureFallbackWritesNothing` (`todo_root_test.go:125`) | 로컬 mtime 불변 + `os.Stat(fallbackRoot)`가 **`IsNotExist`** (`:147-149`) | **불성립** — 홈 루트가 나오지만 **부재 단언**이라 가드 발화 후에도 그대로 참 | **성립** — 긍정 단언이므로 스텁이 안 들면 그 자리에서 RED |
      | `ReadThroughToProjectLocal` (`:156`) | `got == dir` (`:164-166`) | **불성립** — read-through는 정의상 project-local 루트를 돌려주고, 그 `dir`이 **가드의 반환값과 같다**(두 상태를 PASS가 구별 못 함) | **성립** — 같은 이유 |
      | `AdoptionLandsWhereConsumersRead` (`todo_root_contract_test.go:14`) | `os.Stat(BacklogPathForRoot(root))` 성공 (`:23`) | **성립** — 7건 중 **유일** | **성립** (두 겹: 판별식 단언 + 홈 루트 동일성) |
      | `AdoptingAndPureResolversAgreeWhenAdoptionFails` (`:40`) | `adopting == pure` + stat 성공 (`:61-67`) | **불성립** — 픽스처가 adopt를 **고의로 실패**시키므로 두 리졸버 모두 read-through로 `proj`를 돌려주고, `proj`가 다시 가드의 반환값과 같다 | **성립** |
      | `TodoSectionReadsThroughToProjectLocalQueue` (`internal/web/todo_section_test.go:194`) | 렌더 3행 + mtime 불변 + `fallbackRoot` **부재** (`:207-220`) | **불성립** — 1·2행의 결함을 겸한다(부재 단언 + 콘솔은 애초에 루트를 반환하지 않는다) | **성립** — 단, `internal/web`에서 `kanban.TempOriginReason`이 닿아야 하므로 **REQ-THG-009의 exported 요건이 여기서 구속된다** |
      | `SilentOnHomeFallbackFixture` (`internal/cli/todo_queue_root_test.go:298`) | `liveTodoQueueRootReason() == ""` | **불성립** — 단언 대상이 침묵(부재형)이고 가드 전후로 참이다 | **성립** |
      | `AxisACanaryHomeSweep_TodoFamily` (`internal/cli/todo_axisa_guard_test.go:69`) | 자식 실행의 `--- PASS` 계수 > 0 + canary HOME 아래 엔트리 0 | **불성립** — 0-오염이 **부재 단언**이다 | **수용된 손실** — 아래 참조 |

      **7행 — 수용된 손실 (운영자 결재 2026-09-08, 5차 감사 D23)**

      이 테스트는 `exec.Command("go","test",…)`(`todo_axisa_guard_test.go:109`)로 **자식 프로세스**를 판정한다. `REQ-THG-009`의 이음매는 **패키지 수준 변수**(§D12와 같은 급)이고 자식 프로세스는 그것을 물려받지 않는다. 따라서 「비임시 사본」은 **구성 불가**다 — 사본을 만들 수 없거나, 부모 쪽에서 쓴 `TempOriginReason` 단언이 **자식 픽스처에 대해 아무것도 말하지 않는다**(후자가 공허다).

      처분은 둘이다. ① 이 테스트는 **`runTodo` 게이트의 증거이기를 그친다** — 가드 착지 후 canary HOME 0-오염은 게이트가 아니라 리졸버 층 가드가 만든 것이므로, 그 관측만으로는 게이트를 세우지 못한다. ② 그 커버리지는 **M3 경계 보고 항목 2**(`.moai/reports/t536/guard-boundary.md` — t422의 `runTodo` 게이트와 t536의 리졸버 층 가드가 서로 다른 층이고 후자가 전자의 우회를 덮는다)로 간다. 그 항목이 이미 정확히 이 층 경계를 다룬다.

      **왜 하필 여기서 또 났는지를 함께 남긴다.** 이것은 이 SPEC이 세 번 값을 치른 계열(D2·D10·D14)의 **네 번째**이고, **하필 그 계열을 막으려고 쓴 조항 안에** 들어 있었다. 손실만 적고 이 문장을 빼면, 다음에 이음매를 설계하는 사람이 같은 자리에서 다시 지불한다 — 배워야 할 것은 **패키지 수준 변수는 `exec.Command` 자식에 건너가지 않는다**는 것이고, 그래서 프로세스 경계를 넘는 판정 대상에는 프로세스-내 이음매가 닿지 않는다는 것이다.

      기각된 대안: 이음매를 환경변수로도 선언 가능하게 넓히는 안. 프로덕션 코드가 env를 읽게 만드는 것은 **이 카드가 다루는 것과 다른 축의 설계 결정**이고 자기 몫의 검증을 요구한다(`queueRootInsideTemp` 통합을 범위 밖으로 둔 것과 같은 이유).

      > 이 항목은 **6차 수리가 아니라 5차 감사 판정의 반영**이다. 새 설계 결정이 없고 감사가 처방까지 준 것을 그대로 적었으므로 재감사를 요구하지 않는다. 회차 상한을 넘긴 것이 아니다.

      즉 **6/7에서 0.1.3의 판정식이 성립하지 않는다.** 그 6건의 사본은 이음매 스텁이 실패해 가드가 발화해도 **그대로 초록**이다 — 3차가 폐기시킨 「미갱신 테스트 무실패」와 판별력이 같다. 이 SPEC이 이미 세 번 대가를 치른 부류(D2·D10·D14)의 네 번째가 예약된 자리였다.
    - **[HARD] 새 판정식이 RED가 될 수 있음을 M2가 실연한다.** 실패할 수 없는 판정식은 판정식이 아니고, 공허한 판정식을 공허한 판정식으로 바꾸는 것이 이 SPEC이 지금 두 번째로 하려던 일이다. **RED 조건**: 사본 1건에서 **이음매 스텁을 제거**하면(또는 구현이 집합을 호출 시점이 아니라 init 시점에 값으로 복사해 이음매가 사실상 안 듣게 되면) 그 base는 `t.TempDir()` 아래이므로 `TempOriginReason(base)`가 `isTemp == true`를 보고하고 **새 단언이 그 줄에서 FAIL한다**. M2 종료 조건은 그 실행의 출력 전문을 `progress.md` §E.2에 남길 것을 요구한다 — 이것은 픽스처에 대한 뮤턴트이므로 기계적으로 확인 가능하고, **M1만 착지해도 관측된다**(가드 배선을 기다리지 않는다).

**M2 종료 조건**: AC-THG-001(갈래 (a)·(b)·(c) 전부)·003·005 GREEN + **AC-THG-008**(순수성·키 유도 불변) GREEN + **§C.1 14건 전수의 처분 이행**(A행 6건 — 5건 보존 이관 PASS + 1건은 §C.1 D행의 가지 이행 · B행 1건 갱신 + 근거 인용 · C행 **6건** 사본이 판별식 단언을 들고 PASS + **7행 손실 기록**이 §C.1에 실려 있을 것) + **새 판정식의 RED 실연 출력** + `SPEC-WEB-TODO-QUEUE-001` **및** `SPEC-STATE-ANCHOR-001` 산출물 diff 0.

> 종전 판본은 이 자리에 `AC-THG-007`을 적었다. 007은 **뮤턴트 증명**(M3 소관)이고, 순수성·키 유도 불변은 008이다 — 괄호 설명이 가리키는 것이 008이므로 번호를 008로 정정한다. 그대로 두면 M2가 M3 소관인 뮤턴트 관측을 게이트로 요구하거나(§F 순서 구속 위반), 반대로 불변 방어가 M2에서 판정되지 않고 새어 나간다. `plan.md` §F M3의 007 인용은 뮤턴트를 가리키므로 옳다.

### M3 — 뮤턴트 증명 + 배치 검증

- **뮤턴트**(AC-THG-007 판별 증거, D9): 판별식을 상수 `false` 반환으로 무력화한 뒤 M2의 임시-기원 테스트를 돌려 canary HOME 아래 오염이 **재현되는 것**을 관측한다. 못 잡은 뮤턴트가 있으면 함께 보고한다.
- **[HARD] 경계 보고는 산출물이다 — `.moai/reports/t536/guard-boundary.md`.** 「보고한다」를 커밋 메시지나 세션 로그로 해석하지 않는다: `SPEC-STATE-ANCHOR-001` REQ-SA-011이 요구하는 것은 「오염을 만들지 못한 뮤턴트를 **그 사실과 그것이 드러내는 가드 경계와 함께** 보고」하는 것이고, 그 요구는 본 카드가 그 요구사항의 두 번째 가지를 이행하는 순간 본 카드에 걸린다(§C.1 D행, spec.md §8). 파일이어야 하는 이유는 단순하다 — 세션이 끝나면 로그는 읽히지 않고, 카드 증거 경로는 감사자가 읽는 자리다.
  - **담을 것 (셋, 전부 필수)**:
    1. **어느 뮤턴트인가** — 뮤턴트의 정확한 형태(무엇을 무엇으로 바꿨는가)와 그것을 적용해 실행한 명령·출력 전문.
    2. **어느 경계를 드러내는가** — 그 뮤턴트가 오염을 내지 못하는 이유. 여기서는 `runTodo`의 `liveTodoQueueRootReason` 게이트를 우회해도 **리졸버 층의 임시-기원 가드가 위에 있어** 홈 루트가 애초에 반환되지 않는다는 것 — 즉 t422 게이트(테스트 축)와 t536 가드(생산 축)가 **서로 다른 층**이고, 후자가 전자의 우회를 덮는다.
    3. **오염 부재가 왜 가드의 작동이지 뮤턴트의 실패가 아닌가** — 이것이 보고의 핵심이며, 근거는 대조군이다: 같은 뮤턴트를 **가드 이전 트리**(`412c8cb14`)에서 실행하면 오염이 재현된다(그 SPEC의 M4 증거가 그것이다). 두 실행의 차이가 가드 하나임을 보인다. 이 대조가 없으면 「뮤턴트가 그냥 안 돌았다」와 구별되지 않는다.
  - **쓰는 시점은 run-phase다.** plan-phase는 명세만 한다 — 이 파일을 지금 만들지 않는다.
- **배치**: §C 전 행 재측정 + AC 매트릭스 전수 판정, 출력 전문 §E.2.
- `git diff 412c8cb14..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/` 무변경 확인(D8).
- `find ~/.moai/todo -maxdepth 1 -type d | wc -l` 이 §C 값과 동일함을 확인(REQ-THG-008).

**M3 종료 조건**: AC 8건 전수 판정 + 뮤턴트 관측 기록 + **`.moai/reports/t536/guard-boundary.md` 존재 및 세 항목 충족** + D8/D7 diff·계수 확인 + 두 형제 SPEC(`SPEC-WEB-TODO-QUEUE-001`·`SPEC-STATE-ANCHOR-001`) 산출물 diff 0.

## §G 안티패턴 (이 카드에서 특히 하기 쉬운 것들)

- **접두사 한 줄로 끝내기** — `strings.HasPrefix(abs, os.TempDir())`. 이 기계에서 조용히 발화하지 않는 형태다(spec.md §4).
- **정규화를 한쪽에만 적용** — 호출자 경로만 해석하고 임시 루트는 원형으로 두면 같은 불일치가 반대 방향으로 재현된다.
- **범위 확대** — "비git이면 어차피 임시나 마찬가지"라는 단순화. 기각된 판독이다(D1).
- **fail-closed로 뒤집기** — 정규화 실패 시 "임시로 간주"가 안전해 보이지만, 진짜 프로젝트의 큐를 회수하는 방향이다(D4).
- **고아 정리를 곁들이기** — "이왕 온 김에" 343개를 지우는 것. t542의 소관이고 증거다(D7).
- **키 유도 손보기** — 정규화한 경로로 `TodoQueueProjectKey`를 다시 유도하면 기존 홈 큐가 전부 도달 불가가 된다(D6).

## §H 상호 참조

- `SPEC-STATE-ANCHOR-001` §5(그 파일 108-116행) — 본 SPEC이 이행하는 미결 결정이며, `:115`가 「본 결정의 구현은 별도 후속 카드로 발행하는 것이 적절하다」로 **본 카드를 예고한다**. §1.2가 그 SPEC의 「표본 2건」을 343개로 정정한다. 그 SPEC의 REQ-SA-011 / AC-SA-011에 대한 교차 판정은 spec.md §8, 테스트 처분은 §C.1 D행.
- `SPEC-WEB-TODO-QUEUE-001` — 순수/adopting 분할과 read-through 결정(D-2)을 도입한 SPEC. 본 SPEC은 그 분할을 보존한다(D5).
- 카드 **t542** — 기존 고아 정리(범위 밖).
- 카드 **t537** — `stateanchor.Resolve`의 `project_dir` 미검증(다른 파일·다른 카드, 접점 없음).
- `internal/cli/launcher.go:456-485` — 정규화 규칙의 근거 형태와 그 GOOS 논거.
