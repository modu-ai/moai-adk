# Acceptance — SPEC-TODO-HOME-TEMP-GUARD-001

측정 기준 트리: `.claude/worktrees/t536` @ `412c8cb14` (`WT-home-fallback`). 아래 RED는 **예정 형태**다 — 채택 판정(RED 실측)은 plan.md §F의 해당 마일스톤이 수행하고, 실측 출력 전문을 `progress.md` §E.2에 남긴다.

판본: **0.1.4**(spec.md HISTORY와 같은 판본). 0.1.4의 변경은 AC-THG-001 갈래 (c) 추가(D22), AC-THG-003 매트릭스 행의 REQ-THG-009 반영(D21), DoD의 영향 테스트 8→14건과 C행 판정식 교체(D18·D19)다. **AC 개수는 8로 불변이다.**

등급 용어: **blocking** = RED-now 셀 + green path 셀을 갖춘 릴리스 게이트. **blocking (RED 관측 전)** = 예상 RED 형태 + 코드 근거를 기록해두고 실측 RED는 run-phase가 관측하는 것. **invariant-guard** = 기준 트리에서 이미 GREEN이라 RED-now 셀을 가질 수 없는 불변 방어(release-blocking 자격 없음, 기록도 pass로 남기지 않는다).

테스트 이름은 **제안**이며 run-phase가 최종 확정한다. 모든 테스트는 `t.TempDir()` + `stubHome`(canary HOME) 안에서만 판정한다 — 실제 HOME을 측정 대상으로 삼지 않는다(plan.md D10).

---

## §D AC 매트릭스

| AC | 요구사항 | RED (현재 트리) | GREEN (목표) | 등급 |
|---|---|---|---|---|
| AC-THG-001 | REQ-THG-001 | **(a)** 로컬 큐 없음: 비git 임시 base → canary `HOME/.moai/todo/<key>` 루트로 해석. **(b)** 로컬 `backlog.json` 있음: 위에 더해 adopting 경로가 `MkdirAll`에 도달해 canary HOME 아래 디렉터가 **생성됨** (M2 RED 관측 예정) | (a) 홈 루트가 아니라 **`base`** 로 해석 + `BacklogPathForRoot(반환 루트)`가 로컬 큐의 canonical 경로와 일치. (b) 추가로 그 경로의 파일이 **stat 가능**(픽스처 큐가 실제로 읽힌다) + canary HOME 아래 디렉터 0 | blocking (RED 관측 전) |
| AC-THG-002 | REQ-THG-002, 003 | 미해석/해석 두 철자 중 한쪽에서 판별 실패 (M1 RED 관측 예정) | `/var/folders/...`와 `/private/var/folders/...` 모두 임시로 분류 | blocking (RED 관측 전) |
| AC-THG-003 | REQ-THG-005, REQ-THG-009 | (범위 정확성 — 현재는 홈 폴백이 동작) | 비임시 비git base는 홈 폴백 **유지** | blocking |
| AC-THG-004 | REQ-THG-004 | 정규화 실패 입력에서 임시로 오분류 가능 (M1 RED 관측 예정) | "임시 아님" 보고 (fail-open) | blocking (RED 관측 전) |
| AC-THG-005 | REQ-THG-006, 007 | (안내 부재 — 현재는 조용히 홈 큐 생성) | 명령 경로: 안내 관측 + 홈 디렉터 0. 순수 경로: 무쓰기·무에러·무안내 | blocking (RED 관측 전) |
| AC-THG-006 | REQ-THG-002 | 생문자열 접두사 구현이면 `/tmpfoo`를 임시로 오분류 (M1 RED 관측 예정) | 구성요소 경계에서 비임시로 분류 | blocking (RED 관측 전) |
| AC-THG-007 | REQ-THG-001 (판별 증거) | (뮤턴트 오염은 **AC-THG-001 (b) 픽스처에서** 재현 가능 — 그것이 증거) | 판별식 무력화 뮤턴트가 (b) 픽스처에서 canary HOME 오염을 재현하고 (a)·(b) 단언이 모두 FAIL | invariant-guard |
| AC-THG-008 | REQ-THG-007, 008 | (불변 대상 — 현재 상태가 기준) | 키 유도·git 가지·순수성 불변 + `~/.moai/todo` 계수 불변 | invariant-guard |

---

## §D.1 AC 상세

### AC-THG-001 — 임시 기원은 홈 큐를 얻지 않는다

Given이 **두 갈래**다. 갈래를 나누지 않으면 아래 RED 정정이 보여주듯 And절이 공허하게 통과하고, AC-THG-007의 뮤턴트까지 무의미해진다.

**(a) 로컬 큐 없는 갈래 — 반환 루트를 판정한다**

- **Given** git 저장소가 아닌 임시 디렉터(`t.TempDir()`)를 base로 하고, base에 project-local `backlog.json`이 **없으며**, canary HOME이 스텁된 해석 요청이 있고,
- **When** `ResolveTodoQueueRoot`(순수) 및 `ResolveTodoQueueRootAdopting`(명령)이 호출되면,
- **Then** 어느 쪽도 `canaryHOME/.moai/todo/<key>` 를 반환하지 않고, **`base`를 반환한다**(REQ-THG-001, plan.md D12).
- **And** 반환 루트는 **루트 계층에서 성립한다**: `BacklogPathForRoot(반환 루트)`가 `base/.moai/state/todo/backlog.json` — 즉 이 프로젝트의 로컬 큐가 놓이는 canonical 경로 — 와 같다. (이 갈래에는 파일이 없으므로 stat이 아니라 **경로 동일성**으로 판정한다. 파일 존재까지 묻는 것은 (b)다.)

**(b) 로컬 큐 있는 갈래 — 오염 경로를 실제로 밟고, 그 큐가 반환 루트로 실제로 읽히는지를 판정한다**

- **Given** 위와 같되 base에 project-local `backlog.json`이 **존재하고**(픽스처는 `BacklogPathForRoot(base)`에 쓴다 — 기존 `todo_root_test.go:65-67` `seedLocalQueue`와 같은 경로. 그래야 `adoptLocalTodoQueue`가 `:172-174`의 조기 반환을 넘어 `:175`의 `MkdirAll`에 도달한다),
- **When** `ResolveTodoQueueRootAdopting`이 호출되면,
- **Then** 반환 루트는 (a)와 같고,
- **And** **[판정의 핵심] 그 반환 루트로 큐가 실제로 읽힌다**: `os.Stat(BacklogPathForRoot(반환 루트))`가 성공하고, 그 파일이 픽스처가 세운 바로 그 `backlog.json`이다(내용 또는 inode 동일성으로 확인).
- **And** `canaryHOME/.moai/todo` 아래에 어떤 디렉터도 생성되지 않으며, 로컬 `backlog.json`은 제자리에 남아 있다(이관되지 않는다).

**(c) 교차 갈래 — 「임시 기원 ∧ 홈 해석 불가」에서도 반환은 `base`다 (0.1.4, D22)**

- **And** 갈래 (a)의 픽스처에서 canary HOME 스텁 대신 **`HomeDirFn`이 오류를 반환하도록** 주입해도 반환 루트는 여전히 **`base`**이며, `BacklogPathForRoot(반환 루트)`는 (a)와 같은 로컬 canonical 경로를 가리킨다 — `base/.moai/state/todo`(= `todo_root.go:121`의 계층 어긋난 형태)가 **아니다**. 이 절은 REQ-THG-001이 교차에 대해 고정한 값을 산출 AC 없이 두지 않기 위한 것이다(plan-audit 4차 D22). AC 개수는 늘지 않는다 — AC-THG-001의 And 한 줄이다.
- **RED**: 현재 트리에서 그 교차는 `:121`을 거쳐 `base/.moai/state/todo`를 반환한다(`todo_root_test.go:204-217`가 지금 그것을 단언한다 — §C.1 B행, 의도적 갱신 대상). 즉 이 And절은 **오늘 RED다**.

> **이 단언이 무엇을 막는가 (plan-audit 2차 D10 — 지적이 옳다).** 0.1.1의 Then절은 반환 루트의 **문자열 동일성**만 물었다. 그래서 계층이 한 단계 어긋난 값(`base/.moai/state/todo`)을 사양으로 못박아도 — 소비자가 그 위에 `BacklogPathForRoot`를 덧붙여 `base/.moai/state/todo/.moai/state/todo/backlog.json`을 읽고 **존재하는 큐를 못 보게 되는데도** — 두 갈래가 모두 **GREEN으로 통과했다**(D9). 문자열을 묻는 대신 「운영자의 카드가 사는 큐가 이 루트로 도달되는가」를 물으면 그 오류가 RED로 잡힌다. 같은 성질을 고정하는 선례가 이 리포에 이미 있다: `todo_root_contract_test.go:22-26`이 `os.Stat(BacklogPathForRoot(root))`로 "consumers read where the queue is"를 단언한다.

판정 명령(제안): `go test ./internal/kanban/ -run TestTodoQueueRoot_TempOriginRefusesHomeQueue -count=1` (두 갈래를 서브테스트로)

**RED (M2 관측 예정)**: 현재 `todo_root.go:64-69`가 `primaryCheckoutRoot` 실패 시 `fallbackTodoQueueRoot` → `homeTodoQueueRoot`(`:112`)로 무조건 떨어져 **홈 루트가 반환된다** — 이것이 (a)의 RED다. adopting 경로가 `MkdirAll`(`:175`)에 도달하는 것은 **base에 project-local `backlog.json`이 있을 때뿐이다**: `adoptLocalTodoQueue`는 `:172-174`의 `if _, err := os.Stat(local); err != nil { return }`로 조기 반환한다(직접 판독). 이것이 (b)의 RED다. **읽힘 단언의 RED도 같은 사실에서 나온다**: 반환 루트가 홈 루트이므로 `BacklogPathForRoot(반환 루트)`는 `canaryHOME/.moai/todo/<key>/.moai/state/todo/backlog.json`을 가리키고, (a)에서는 로컬 canonical 경로와 불일치, (b)에서는 픽스처가 `base` 아래에 쓴 파일이 거기 없다.

> **정정 기록 (plan-audit D2 — 지적이 옳다).** 종전 판본은 갈래를 나누지 않은 채 "adopting 경로는 `MkdirAll`까지 간다(`:175`)"고 **무조건** 적었다. 그 문장은 자기 Given(로컬 큐를 세우지 않는 `t.TempDir()`) 아래에서 **거짓**이고, 귀결로 And절("canary HOME 아래 디렉터 0")이 **가드 없는 현재 트리에서도 이미 참**이라 그 절반이 공허하게 통과했다. 갈래 (b)가 그 공허를 닫는다 — `MkdirAll` 경로가 실제로 도달 가능해져야 「디렉터 0」이 가드의 결과가 된다.

**GREEN (M2)**: 두 갈래 모두 PASS.

### AC-THG-002 — 심링크 두 철자가 등가로 분류된다 (판별식의 핵심)

- **Given** 임시 루트의 **미해석** 철자(`os.TempDir()` = 이 기계에서 `/var/folders/kt/.../T/`)로 만든 base와, 같은 위치의 **해석된** 철자(`filepath.EvalSymlinks` 결과 = `/private/var/folders/...`)로 만든 base가 있고,
- **When** 판별식이 각각을 분류하면,
- **Then** 둘 다 임시 기원으로 보고되고, 보고된 이유(매치된 임시 루트)가 같은 루트를 이름 부른다.

판정 명령(제안): `go test ./internal/kanban/ -run TestTempOrigin_SymlinkSpellingEquivalence -count=1`

**RED (M1 관측 예정)**: 한쪽만 정규화하는 구현(또는 `strings.HasPrefix(abs, os.TempDir())`)에서 해석된 철자가 비임시로 분류된다. 근거: 이 기계 실측 — `readlink /var → private/var`, `readlink /tmp → private/tmp`, `TMPDIR=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/`(미해석). 그리고 `TodoQueueProjectKey`(`todo_root.go:197-204`)가 `filepath.Abs`만 쓰므로 리졸버에 도달하는 철자는 호출자에 달려 있다.

**GREEN (M1)**: 두 철자 모두 PASS.

> **리눅스 셀**: `/tmp`가 실디렉터이고 `/private/tmp`가 없는 환경에서 이 테스트가 어떻게 판정되는지는 이 기계에서 측정하지 않았다(spec.md §8). CI linux 매트릭스의 결과를 §E.2에 인용한다 — **로컬 macOS PASS를 리눅스 판정으로 재사용하지 않는다.**

### AC-THG-003 — 비임시 비git 프로젝트는 홈 폴백을 유지한다 (범위의 정확성)

- **Given** 임시 루트 **밖**에 있고 git 저장소가 **아닌** base와 스텁된 canary HOME이 있고,
  - **[픽스처 구성 — 종전 예시는 성립하지 않는다]** 종전 판본은 예시로 "canary HOME 아래의 일반 디렉터"를 들었으나, canary HOME 자체가 `t.TempDir()`이므로 그 아래 디렉터도 **임시 루트 아래**다 — 가드가 발화해 이 Given이 자기 모순이 된다. 그리고 이 리포의 [HARD] 테스트 격리 규율(모든 테스트 임시 디렉터는 `t.TempDir()`)은 "그냥 비임시 디렉터를 쓴다"는 도피처를 **허용하지 않는다**. 그래서 픽스처는 REQ-THG-002가 요구하는 **임시 루트 집합 주입 이음매**를 통해 구성한다: base는 `t.TempDir()` 그대로 두고(격리 규율 준수), 그 실행 동안 임시 루트 집합을 base를 포함하지 않는 값으로 스텁해 「비임시 비git base」를 선언한다. `HomeDirFn`/`stubHome`이 canary HOME을 만드는 방식과 같다.
  - **이 AC가 곧 이음매(REQ-THG-009)의 검증이다**: 이음매가 없으면 이 픽스처가 존재할 수 없고, REQ-THG-005는 산출 AC를 잃는다(spec.md §8 판정 기록). 0.1.4에서 이음매가 `REQ-THG-002`에서 **`REQ-THG-009`로 분리**됐으므로, 이 AC는 **REQ-THG-005와 REQ-THG-009 둘 다**를 겨눈다(§D 매트릭스 행, D21).
  - **And** 픽스처는 자기 base에 대해 **판별식의 판정을 직접 단언한다**: `TempOriginReason(base)`가 `isTemp == false`를 보고한다. 이것이 이음매가 실제로 들었다는 **긍정 단언**이며, 스텁이 듣지 않으면 그 자리에서 RED가 된다(§F M2 C행 판정식과 같은 형태 — D19).
- **When** `ResolveTodoQueueRoot` / `ResolveTodoQueueRootAdopting`이 호출되면,
- **Then** 종전대로 `canaryHOME/.moai/todo/<TodoQueueProjectKey(base)>`로 해석된다 — 문서화된 "exactly one queue" 가용성이 유지된다.
- **And** adopting 경로의 adopt-not-shadow 이관도 종전대로 동작한다.

판정 명령(제안): `go test ./internal/kanban/ -run TestTodoQueueRoot_NonTempNonGitKeepsHomeFallback -count=1`

**RED**: 없음 — 이것은 **불변 단언이자 범위 가드**다. 기각된 넓은 판독(「모든 비git base 거부」)을 구현하면 이 테스트가 FAIL한다. 그 의미에서 AC-THG-003은 M2 구현의 오확대를 잡는 blocking 게이트다.

**GREEN (M2)**: 테스트 PASS.

### AC-THG-004 — 판별 불가는 "임시 아님"이다 (fail-open)

- **Given** 정규화가 성립하지 않는 입력(존재하지 않는 경로, 접근 불가 경로, 빈 문자열)이 있고,
- **When** 판별식이 호출되면,
- **Then** `isTemp == false`가 보고되고, 홈 폴백은 종전대로 동작한다 — 어떤 입력에서도 판별식이 panic하거나 에러를 전파하지 않는다.

판정 명령(제안): `go test ./internal/kanban/ -run TestTempOrigin_FailsOpenOnUnresolvable -count=1`

**RED (M1 관측 예정)**: fail-closed 구현(정규화 실패 시 임시로 간주)에서 비임시 입력이 홈 폴백을 잃는다.

**GREEN (M1)**: 테스트 PASS.

### AC-THG-005 — 거부는 조용한 성공이 아니다 (그리고 콘솔은 조용하다)

- **Given** 비git 임시 base에서 `moai todo` 명령 경로가 큐 루트를 해석하고,
- **When** 가드가 홈 큐를 거부하면,
- **Then** 안내가 관측되고, 그 안내는 ⑴ **매치된 임시 루트를 이름 부르며**, ⑵ **계속 쓰는 대체 루트(`base`)를 이름 부르고**, ⑶ canary HOME 아래 디렉터는 0개다.
- **And** 명령은 **계속 진행한다** — 종료 코드는 0이고, 대체 루트의 큐에 대해 정상 동작한다(U1 = (b), plan.md §D).
- **And** 같은 입력에서 순수 경로(`ResolveTodoQueueRoot`, 콘솔이 임포트하는 것)는 **아무것도 쓰지 않고 에러도 안내도 내지 않는다** — 페이지 렌더가 백로그를 건드리지 않는다는 `SPEC-WEB-TODO-QUEUE-001`의 불변이 그대로다.

판정 명령(제안): `go test ./internal/kanban/ ./internal/cli/ -run 'TestTempOriginGuidance|TestTodoQueueRoot_PureGuardIsSilent' -count=1`

**RED (M2 관측 예정)**: 현재는 안내 없이 홈 큐가 생긴다(조용한 성공).

**GREEN (M2)**: 두 단언 모두 PASS. **판정 대상은 Then절의 ⑴⑵⑶ 세 항목 전부이며, GREEN이라고 해서 축소되지 않는다.**

**U1 두-형태 표현의 가지치기 (정정 이력 — 지우지 않고 남긴다).** 종료 형태는 이제 **(b) 안내 후 계속**으로 결정됐다(운영자, 2026-09-07; plan.md §D). 그래서 이 AC는 (b) 하나만 판정한다.
- 0.1.0의 ⑵는 "대신 **해석된** 루트를 이름 부른다"였고, (a)에서는 그런 루트가 존재하지 않으므로 **(b) 편향**이었다. 중립성은 GREEN 주석이 판정 대상을 3항 → 2항으로 **사후 축소**함으로써만 성립했다 — 같은 AC 안에 판정 대상이 두 개였다(plan-audit 1차 D3, 옳은 지적).
- 0.1.1이 ⑵를 양쪽에 지시대상을 갖는 선언형으로 고쳐 그 축소를 없앴다. **그 두-형태 표현은 U1이 열려 있던 동안 옳았다.**
- 0.1.2가 결정된 형태로 좁혔다 — 표현이 틀려서가 아니라 **떠받치던 미결이 닫혔기 때문**이다. 판정 대상은 여전히 Then절 전문이며 GREEN이라고 축소되지 않는다.

### AC-THG-006 — 경계는 구성요소 단위다

- **Given** 임시 루트의 **형제** 경로(`/tmpfoo` 형태 — 임시 루트 문자열을 접두사로 갖지만 그 아래가 아닌 경로)가 base이고,
- **When** 판별식이 분류하면,
- **Then** 비임시로 보고된다.
- **And** 임시 루트 **자기 자신**과 그 **하위** 경로는 임시로 보고된다.

판정 명령(제안): `go test ./internal/kanban/ -run TestTempOrigin_ComponentBoundary -count=1`

**RED (M1 관측 예정)**: `strings.HasPrefix` 구현에서 `/tmpfoo`가 임시로 오분류된다.

**GREEN (M1)**: 테스트 PASS.

### AC-THG-007 — 가드가 잡는 것의 뮤턴트 증명

- **Given** M2의 임시-기원 테스트(AC-THG-001)가 **두 갈래 모두** GREEN인 상태에서 판별식을 상수 `false` 반환으로 무력화한 뮤턴트가 있고,
- **When** 같은 테스트를 canary HOME에서 실행하면 — **오염 재현은 갈래 (b)(로컬 `backlog.json`이 있는 픽스처)가 담당한다**,
- **Then** 두 갈래가 **모두 FAIL**한다: (a)는 반환 루트가 `canaryHOME/.moai/todo/<key>`로 되돌아가 실패하고, (b)는 그에 더해 `adoptLocalTodoQueue`가 `:172-174`를 통과해 `:175`의 `MkdirAll`을 실행하므로 canary HOME 아래 홈 큐 오염이 **실제로 재현되어** 실패한다 — 0-오염이 우연이 아니라 판별식의 결과라는 증거.
- **And** 갈래 (a)만으로는 이 증거가 성립하지 않는다: 로컬 큐가 없으면 뮤턴트를 걸어도 `MkdirAll`에 도달하지 않아 디렉터가 생기지 않는다. 뮤턴트가 잡히는 것은 (a)에서는 **반환 루트**를 통해서이고, **오염 재현**을 통해서는 (b)에서만이다.
- **And** **0.1.2에서 추가된 읽힘 단언(D10)의 판별 범위 — 갈래마다 다르다.** (a)에서는 FAIL 사유를 **하나 더한다**: 뮤턴트 아래 반환 루트가 `canaryHOME/.moai/todo/<key>`로 되돌아가므로 `BacklogPathForRoot(반환 루트)`는 홈 아래 경로가 되고, 로컬 큐의 canonical 경로와 **일치하지 않는다**. 그러나 **(b)에서는 뮤턴트에 공허하게 통과한다** — 아래 정정 기록 참조. 그리고 두 갈래 모두에서 **D9 계열의 계층 오류**(반환 루트가 한 계층 어긋난 값일 때)는 제대로 RED로 잡는다. 즉 이 단언의 값은 (a)의 추가 FAIL 사유와 양 갈래의 계층 방어이지, (b)의 뮤턴트 판별이 아니다.

- **And** 무력화해도 잡히지 않는 뮤턴트가 있으면 그것도 함께 보고한다 — 그것이 가드의 경계다(plan.md D9). **보고의 형태는 산출물이다**: `.moai/reports/t536/guard-boundary.md`(plan.md §F M3). 이것은 `SPEC-STATE-ANCHOR-001` REQ-SA-011의 두 번째 가지(「오염을 만들지 못한 뮤턴트는 그 사실과 가드 경계를 **보고하라**」)가 요구하는 보고이기도 하다 — 커밋 메시지나 세션 로그가 아니라 카드 증거 경로의 파일이어야 한다(spec.md §8 AC-SA-011 판정).

> **정정 기록 (plan-audit 3차 D14 — 지적이 옳다. 종전 문장을 지우지 않고 여기 남긴다).** 0.1.2 판본은 이 자리에 「(b)에서는 그 경로가 **stat 불가**(픽스처는 `base` 아래에 썼고, 게다가 `Rename`으로 이동됐다)라 FAIL한다」고 적었다. **거꾸로다.** 인과가 반대다 — `adoptLocalTodoQueue`의 `Rename` 목적지가 바로 `BacklogPathForRoot(fallbackRoot)`이므로(`todo_root.go:168` `target := BacklogPathForRoot(fallbackRoot)`, `:179` `os.Rename(local, target)`), `Rename`은 stat 실패의 이유가 아니라 **stat 성공의 이유**다. 따라서 뮤턴트 아래 (b)에서 `os.Stat(BacklogPathForRoot(반환 루트))`는 **성공하고**, 읽힘 단언은 거기서 **공허하게 통과한다**.
>
> **실행 증거(재측정 불필요 — 이 판독은 실행으로 반증됐다).** 가드가 아직 없는 현재 트리의 동작이 곧 「판별식 상수 `false`」 뮤턴트와 같은 상태이고, 갈래 (b)의 픽스처를 쓰는 기존 테스트 둘이 그 상태에서 통과한다:
>
> ```
> $ go test ./internal/kanban/ -run 'TestAdoptionLandsWhereConsumersRead|TestResolveTodoQueueRootAdopting_AdoptsLocalQueue' -count=1
> --- PASS: TestAdoptionLandsWhereConsumersRead (0.06s)
> --- PASS: TestResolveTodoQueueRootAdopting_AdoptsLocalQueue (0.06s)
> ```
>
> `TestAdoptionLandsWhereConsumersRead`(`todo_root_contract_test.go:22-26`)는 `os.Stat(BacklogPathForRoot(반환 루트))`가 **성공**해야만 통과한다. 통과했다.
>
> **결론과 논거를 갈라 적는다.** ⑴ **결론은 살아남는다** — 뮤턴트는 여전히 양 갈래에서 FAIL한다. ⑵ **기재된 논거는 거짓이었다** — (b)의 FAIL은 읽힘 단언에서 오지 않는다. 둘을 섞어 「결론이 맞으니 됐다」로 넘기지 않는다: 이 문장을 읽고 뮤턴트 관측을 설계한 구현자는 stat 실패를 기대하다 성공을 보고 「테스트가 틀렸다」로 오진하거나, 반대로 그 단언 하나로 (b)의 뮤턴트 포착을 확인했다고 기록할 수 있다.
>
> **(b)의 실제 FAIL 사유는 셋이다**(위 Then절이 이미 그것들을 담고 있다): ⑴ 반환 루트가 `base`가 아니라 `canaryHOME/.moai/todo/<key>`다, ⑵ canary HOME 아래 디렉터 계수가 0이 아니다(`MkdirAll`이 돌았다), ⑶ 로컬 `backlog.json`이 제자리에 없다(`Rename`으로 이동됐다).
>
> **이것이 이 SPEC에서 세 번째 공허한 초록이다**(D2 · D10 · D14). 반복 자체에 대한 관측은 spec.md §8에 있다.

판정 절차: 뮤턴트 적용 → 테스트 실행 → 출력 전문 `progress.md` §E.2 인용 → 뮤턴트 되돌림 확인(`git diff` 무변경).

**등급 근거**: 부재 가드(「오염이 생기지 않는다」)는 기능이 없어도 저절로 만족되므로 RED-now로 채택 판정할 수 없다 — invariant-guard로 기록하고, 판별 증거는 뮤턴트다.

### AC-THG-008 — 건드리지 않은 것들이 그대로다

- **Given** M3 종료 시점의 트리에서,
- **When** 아래를 측정하면,
- **Then** 모두 기준값과 같다:
  - `TodoQueueProjectKey`의 유도가 불변(테스트: 같은 base에 대해 수리 전후 같은 키) — 기존 홈 큐가 도달 불가가 되지 않는다.
  - git 가지 불변: 저장소 안(임시 루트 아래의 저장소 포함)의 base는 `primaryCheckoutRoot`로 해석되고 가드에 **도달하지 않는다**.
  - `ResolveTodoQueueRoot`의 순수성: 어느 가지에서도 `MkdirAll`/`Rename`/`WriteFile` 없음(기존 `TestResolveTodoQueueRoot_PureFallbackWritesNothing` 계열 통과).
  - `git diff 412c8cb14..HEAD -- internal/statusline/ internal/config/ internal/hook/ internal/session/ internal/stateanchor/` 무변경.
  - `find ~/.moai/todo -maxdepth 1 -type d | wc -l` 이 plan.md §C 기준값(344)과 동일 — 고아 343개는 증거로 보존된다(t542 소관).

판정 명령(제안): `go test ./internal/kanban/ -run 'TestTodoQueueRoot_GitBranchUnreachedByGuard|TestResolveTodoQueueRoot_PureFallbackWritesNothing' -count=1` + 위 두 셸 측정.

---

## §D.2 Definition of Done

- AC 8건 전수 판정, 각 판정의 명령 + 실제 출력 + HEAD SHA가 `progress.md` §E.2에 있다.
- AC-THG-001의 픽스처가 **두 갈래((a) 로컬 큐 없음 / (b) 로컬 `backlog.json` 있음)로 실재**하고, (b)가 `adoptLocalTodoQueue`의 `MkdirAll`(`todo_root.go:175`) 경로를 실제로 밟는다는 것이 RED 실측 출력으로 보인다.
- 뮤턴트 관측 기록(AC-THG-007)이 있고 **(b) 픽스처에서 오염이 재현됐음**이 출력으로 보이며, 못 잡은 뮤턴트가 있으면 그것도 기록돼 있다.
- **영향 테스트 14건 전수**(plan.md §C.1 영향 표 — 0.1.4에서 8→14로 정정, D18)의 처분이 이행됐음이 보인다 — 1건이 아니라 전수다:
  - **확실히 깨지는 6건**(`FallbackNoGit` 2개 패키지 · `PopulatedFallbackWins` · `AdoptsLocalQueue`) + **배치 의존 1건**(`HomeUnresolvableWritesNothing` — REQ-THG-001의 교차 결정으로 이제 의도적 갱신): 각각 「보존 이관」인지 「의도적 갱신」인지가 plan.md §F M2의 처분대로 이행되고, 보존 대상의 원 단언이 실제로 살아 있음이 출력으로 보인다.
  - **공허화 7건**: 각각에 대해 비임시 픽스처 사본이 존재하고, **그 사본이 자기 base에 대해 판별식의 「임시 아님」 판정을 직접 단언한다**(`TempOriginReason(base)` → `isTemp == false`)는 것이 출력으로 보인다. **「미갱신 테스트 무실패」로는 이 7건을 볼 수 없다** — 공허하게 통과하는 테스트는 실패하지 않기 때문이다. 0.1.3의 대체 판정식(「사본의 단언 대상이 홈 루트인 것」)도 **7건 중 1건에만 성립했으므로 폐기됐다**(plan-audit 4차 D19; 건별 결과는 plan.md §F M2).
  - **판정식의 RED 조건이 실연됐다**: 사본 1건에서 이음매 스텁을 고의로 제거한 상태의 실행 출력이 §E.2에 있고, 그 출력에서 새 단언이 **실제로 FAIL**한다. 실패할 수 없는 판정식은 판정식이 아니며, 이 SPEC은 이미 공허한 판정식을 두 번 실었다(D14 → D19).
  - `SPEC-WEB-TODO-QUEUE-001`의 AC 산출 테스트 3건(AC-WTQ-006/007/008)이 **보존**됐고 그 SPEC의 산출물은 **한 글자도 수정되지 않았다**(spec.md §8 판정 기록).
  - `SPEC-STATE-ANCHOR-001`의 **AC-SA-011 산출 테스트**(`internal/cli/todo_axisa_guard_test.go:150`)가 REQ-SA-011의 **두 번째 가지로 이행**됐고, 그 SPEC의 산출물도 **한 글자도 수정되지 않았다**(spec.md §8 판정 기록). 그 이행의 산출물인 **가드 경계 보고**(`.moai/reports/t536/guard-boundary.md`)가 존재하고, plan.md §F M3가 정한 세 항목(어느 뮤턴트인가 · 어느 경계를 드러내는가 · 오염 부재가 왜 가드의 작동이지 뮤턴트의 실패가 아닌가)을 담는다.
- **반환 루트가 루트 계층에서 성립함이 관측으로 보인다**: (b)에서 `BacklogPathForRoot(반환 루트)`가 픽스처 `backlog.json`을 가리키고 stat이 성공한 출력이 §E.2에 있다(D10 — 문자열 동일성만으로는 부족하다).
- U1은 **plan.md §D에서 이미 (b)로 결정됐다**. run-phase는 그 형태를 구현하고, 구현이 결정과 일치함(종료 코드 0 + 안내가 대체 루트를 이름 부름)을 §E.2에 인용한다 — 선택을 다시 기록하는 것이 아니다.
- `go test ./internal/kanban/... ./internal/web/... -count=1` 통과, CI가 linux 셀을 포함해 초록.
- `~/.moai/todo` 계수 불변, 접촉 금지 패키지 diff 0.
