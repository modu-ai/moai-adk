# SPEC-BINLAG-KEYGUARD-001 — 수락 기준

카드 **t479**. 문서 수준 트리 핀: 아래 「기준선」 관측은 워크트리 `.claude/worktrees/t479`,
HEAD **`93fb36344`**(로컬 `develop @ a825183dd` 흡수 병합 커밋)에서 얻었다.
개별 항목이 자기 SHA를 갖지 않으면 이 문서 핀이 구속한다.

읽는 법: 이 파일의 항목은 **Given-When-Then**이다. GEARS 문장은 요구 계층(`spec.md` §2)에 있고
여기에 사본을 두지 않는다.

**[HARD] 새 가드의 확정 이름은 `TestBinaryLag_AllowlistKeysAreLiveNames`다.** 아래 레시피들이 이
이름을 문자 그대로 쓰므로, 이름이 달라지면 레시피가 아무것도 잡지 못한 채 초록이 된다.
이름을 바꾸려면 이 파일의 모든 레시피를 같은 커밋에서 함께 고친다.

**[HARD] 판정식** (REQ-BLKG-005 준수): 종료코드 **0을 통과로 읽는** 판정 줄은 **비공허성 단언과
짝지어질 때에만** 쓴다. 짝 없는 단독 사용은 금지다. 종료코드 **≠ 0을 실패로 읽는** 판정은
공허한 경우가 실패 쪽으로 떨어지므로 이 제약 밖이다 — **단, 실패의 정체를 함께 요구할 때에만**.
`≠ 0`은 빌드 실패·setup 실패(`[setup failed]`)·패닉도 포함하므로, 그것만 읽으면 「가드가 잡았다」와
「애초에 빌드가 안 됐다」가 합쳐진다. 이 문서에서 `≠ 0`을 읽는 세 자리(A 분류)는 모두 실패 메시지
본문이나 파일:줄을 함께 요구하므로 그 구멍이 레시피에는 없다 — 좁힌 것은 문장 쪽이다.

**예외 조항은 폐기했다.** 이전 판의 이 자리에는 「`go test`의 종료코드는 러너 자신의 판정이라
못 찾음과 실패가 섞이지 않는다」는 문장이 있었다. **그 문장은 실측으로 거짓이다** — 트리
`93fb36344`, 새 가드 부재 상태에서:

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.704s [no tests to run]
    (exit 0)

「하나도 안 돌았다」가 「전부 통과했다」와 같은 신호를 낸다. 러너의 종료코드도 안전하지 않다.
근거 규칙: `.claude/rules/moai/development/verification-completeness.md` §1.1 —
*"A pass whose swept set is empty asserts nothing"*, go의 `[no tests to run]` 토큰이 가장 값싼 증거.
잘못 적었던 문장을 지우지 않고 남기는 것은, 다음 사람이 같은 예외를 다시 만들지 않게 하기 위해서다.

**`git merge-base --is-ancestor`만 종료코드로 읽되, 단서가 붙는다.** 그 명령은 stdout이 비어 있어
종료코드가 유일한 출력이지만, **유효한 ref에 한해서만** 0/1이 조상 여부를 뜻한다. 오타 ref는
`fatal:`을 stderr로 내고 `128`을 돌려준다(실측, 트리 `93fb36344`):

    $ git merge-base --is-ancestor deadbee9 develop
    fatal: Not a valid object name deadbee9
    (exit 128)

따라서 이 명령을 읽을 때는 **0 / 1 / 그 밖**을 갈라, 그 밖이면 판정이 아니라 ref 오류로 처분한다.

**[HARD] diff base는 `93fb36344`로 명시 핀하고 two-dot을 쓴다.** `origin/develop`을 base로 쓰면
안 된다 — 그 ref는 `25a3212a9`에 멈춰 있고 로컬 `develop`보다 130 커밋 뒤이며(레인은 develop을
push하지 않으므로 구조적으로 그렇다), three-dot은 merge-base가 `25a3212a9`로 내려간다.
그러면 t466·t477이 바꾼 파일이 이 카드 몫으로 집계돼 **run-phase가 한 글자도 고치지 않은 상태에서도
항목이 거짓 실패**한다. 실측(HEAD `93fb36344`):

    $ git diff --name-only origin/develop...HEAD -- internal/ pkg/ cmd/ | wc -l
    24        ← 세 점(three-dot) + 스테일 ref: 남의 카드 24개가 섞인다
    $ git diff --name-only 93fb36344..HEAD -- internal/ pkg/ cmd/ | wc -l
    0         ← 두 점(two-dot) + 명시 핀: 이 카드의 실제 델타

---

## §C 증거 원장

`verification-completeness.md` §2.1이 허용하는 두 carrier 중 **원장** 쪽이다. 표 칸을 쓰지 않는 이유는
그 절이 적은 그대로다 — 표 칸은 줄바꿈과 셸 메타문자를 뭉개므로 **verbatim stdout을 보존하지 못한다**.
아래 각 항목은 네 요소(명령 · verbatim stdout · 종료코드 · 트리 SHA)를 함께 담고, 수락 기준이 id로 인용한다.

**모두 트리 SHA `93fb36344`, 워크트리 `.claude/worktrees/t479`에서 측정했다.**

### EV-BLKG-001 — 새 가드 부재 상태의 RED-now (AC-BLKG-001이 인용)

- **명령** (단일 호출):

      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s -v

- **verbatim stdout**:

      testing: warning: no tests to run
      PASS
      ok  	github.com/modu-ai/moai-adk/internal/cli	1.276s [no tests to run]

- **종료코드**: `0`
- **트리 SHA**: `93fb36344`
- **파생 계수**: `[no tests to run]` → **1** (1단 위반) · 뒤 공백형 `--- PASS: <이름> ` → **0** (2단 위반)
- 경과 시간(`1.276s`)은 실행마다 달라지는 유일한 필드다. 재현 시 그 자리만 다를 수 있으며,
  판정에 쓰이는 두 토큰은 달라지지 않는다.

### EV-BLKG-002 — 대조군: 실재하고 실제로 통과하는 테스트 (AC-BLKG-001이 인용)

- **명령** (단일 호출):

      go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s -v

- **verbatim stdout**:

      === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
      --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.06s)
      PASS
      ok  	github.com/modu-ai/moai-adk/internal/cli	0.990s

- **종료코드**: `0`
- **트리 SHA**: `93fb36344`
- **파생 계수**: `[no tests to run]` → **0** · 뒤 공백형 → **1** · **`$` 앵커형 → 0**(상시 0)
- EV-BLKG-001과 **종료코드가 같다**(`0`). 이 한 쌍이 「종료코드는 판정에 쓸 수 없다」의 증거다.

### EV-BLKG-003 — SKIP 경로: 1단만으로는 막지 못한다 (AC-BLKG-001 · AC-BLKG-004가 인용)

- **명령** (단일 호출):

      go test ./internal/cli/ -run TestHookCommandFlushesLastHandlerEntry -count=1 -short -timeout 600s -v

- **verbatim stdout**:

      === RUN   TestHookCommandFlushesLastHandlerEntry
          hook_flush_test.go:81: builds the moai binary; skipped under -short
      --- SKIP: TestHookCommandFlushesLastHandlerEntry (0.00s)
      PASS
      ok  	github.com/modu-ai/moai-adk/internal/cli	0.708s

- **종료코드**: `0`
- **트리 SHA**: `93fb36344`
- **파생 계수**: `[no tests to run]` → **0**(1단을 통과해 버린다) · 뒤 공백형 → **0**(2단이 잡는다)
- 이 표본은 `binary_lag_test.go`가 아니라 같은 패키지의 `-short` 스킵 테스트다. `t.Skipf` 경로를
  **편집 없이** 관측하려고 골랐으며, 스킵의 출력 모양은 어느 테스트든 같다.
  `binary_lag_test.go:202`의 `t.Skipf`가 이 카드에서 실재하는 그 경로다.

---

## §D 수락 기준 매트릭스

| ID | 요구 | 분류 | 방향 |
|---|---|---|---|
| AC-BLKG-001 | REQ-BLKG-001 | 릴리스 차단 | 키 실재 단언 |
| AC-BLKG-002 | REQ-BLKG-002 | 릴리스 차단 | 무력 키 적발 |
| AC-BLKG-003 | REQ-BLKG-003 | 릴리스 차단 | 실패가 원인을 지목 — **양방향** |
| AC-BLKG-004 | REQ-BLKG-004 | 릴리스 차단 | 비공허성 |
| AC-BLKG-005 | REQ-BLKG-002 | 릴리스 차단 | **뮤턴트 m1** |
| AC-BLKG-006 | REQ-BLKG-003 | 릴리스 차단 | **뮤턴트 m2 / m2′** |
| AC-BLKG-007 | REQ-BLKG-006 | 회귀 가드 | 폭발 반경 |
| AC-BLKG-008 | REQ-BLKG-005, REQ-BLKG-007 | 회귀 가드 | 판정식·독립성 |
| AC-BLKG-009 | REQ-BLKG-004, REQ-BLKG-005 | 릴리스 차단 | **판정식 자신의 RED 관측** |

실질 판정 대상 **9건** 전부가 run-phase의 판정 대상이다. **자명 충족으로 강등되는 항목은 없다.**

---

### AC-BLKG-001 — 허용목록의 모든 키가 현재 트리에서 실제로 추출된다

**Given** `namesAddedAfterBaseline`에 키가 하나 이상 있고,
**When** 새 가드를 실행하면,
**Then** 모든 키가 현재 `internal/cli/doctor.go`에서 `checkNamesFromSource`가 만든 집합의 원소임이 확인되고 테스트가 통과한다.

- 판정 명령 — **단일 호출**(파이프·리다이렉션·체이닝 없음; `verification-completeness.md` §2.1):

      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s -v

- **판정식은 그 호출의 stdout에서 두 매치 수를 읽는다. 종료코드는 기록만 하고 판정에 쓰지 않는다.**
  §1.1이 빈 sweep 판정을 **다른 모든 신호보다 먼저** 내리라 하므로 순서가 고정이다:

  | 순서 | 계수 대상 | 요구값 | 못 맞추면 |
  |---|---|---|---|
  | **1단 (먼저)** | `[no tests to run]` | **0** | 빈 sweep — **판정 불가**. 원인(함수 부재 / 이름 불일치)을 밝혀 고친 뒤 다시 잰다 |
  | **2단** | `^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames ` (**뒤 공백**으로 끝남) | **정확히 1** | 선택·실행·통과 중 하나가 성립하지 않았다. SKIP도 여기서 0으로 떨어진다 |

  계수는 위 호출의 stdout을 담은 파일에 대해 `grep -c <패턴> <파일>`로 낸다. **종료코드는 읽지 않는다.**

- **[HARD] 패턴 끝의 공백은 하중을 진다. `$`로 조이면 안 된다.** go는 통과 줄 끝에 ` (0.06s)`를
  붙이므로 `$` 앵커는 **어떤 통과에서도 매치하지 않는다** — 상시 0이며, 그것을 「통과 = 1」로 쓰면
  **가드가 완벽히 작동해도 항상 실패**로 읽힌다. 실측(트리 `93fb36344`, 실제로 통과하는 테스트):

      $ grep -e '--- PASS' OUT | cat -e
      --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.06s)$
      $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged$' OUT
      0
      $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged ' OUT
      1

- **[HARD] 1단만으로는 SKIP을 막지 못한다 — 그래서 2단이 함께 필요하다.** 테스트가 존재하되
  `t.Skipf`로 건너뛰면 `[no tests to run]`은 찍히지 않고 `ok` + exit 0만 나온다. 그 경로는 이 파일에
  실재한다(`internal/cli/binary_lag_test.go:202`). 실측(같은 트리, `-short`로 실제 스킵되는 표본):

      $ go test ./internal/cli/ -run TestHookCommandFlushesLastHandlerEntry -count=1 -short -timeout 600s -v
      === RUN   TestHookCommandFlushesLastHandlerEntry
          hook_flush_test.go:81: builds the moai binary; skipped under -short
      --- SKIP: TestHookCommandFlushesLastHandlerEntry (0.00s)
      PASS
      ok  	github.com/modu-ai/moai-adk/internal/cli	0.708s
      (exit 0)

      [no tests to run] 계수 → 0      ← 1단을 통과해 버린다
      뒤 공백형 PASS 계수  → 0      ← 2단이 잡는다

- **RED-now**: 증거 원장 **EV-BLKG-001**을 인용한다(§2.1의 두 허용 carrier 중 원장 쪽 — 표 칸은
  줄바꿈과 셸 메타문자를 뭉개 verbatim을 보존하지 못하므로 쓰지 않는다).

  **읽는 법**: 새 가드가 존재하지 않는 지금, 그 실행은 종료코드 `0`과 `ok`와 `PASS`를 **전부** 낸다.
  그러므로 「부재 자체가 RED-now다」는 **거짓이며**, 이전 판의 그 주장을 여기서 철회한다.
  지금 이 항목을 RED로 만드는 것은 종료코드가 아니라 1단 계수(`[no tests to run]` → 1)이며,
  그래서 이 항목은 **판정 불가**로 떨어진다 — 통과로 세어지지 않는다.

  **대조군**은 EV-BLKG-002다. 두 원장의 차이가 판별식이다: 뒤 공백형 `--- PASS: <이름> ` 줄은
  실재하고 실제로 통과할 때만 나오고, `[no tests to run]`은 빈 sweep일 때만 나온다.
  **두 실행 모두 종료코드는 0이다** — 그래서 종료코드는 판정에 쓸 수 없다.

- 이 트리의 허용목록은 키 2개(`hookWiringCheckName` 상수 등록 · `` `"Hook Delivery"` `` 리터럴 등록)를
  담고 있으므로, 이 항목은 **두 모양 모두**에 대해 실재를 단언한다.

---

### AC-BLKG-002 — 아무것도 매치하지 않는 키는 실패로 드러난다

**Given** 허용목록에 현재 추출 집합의 어느 원소와도 일치하지 않는 키가 있고,
**When** 새 가드를 실행하면,
**Then** 테스트가 실패하며 실패 메시지가 그 키를 이름으로 지목하고, 그 키가 어떤 이름도 허용하지 못한다는 사실을 말한다.

- 판정 방식: AC-BLKG-005(m1) / AC-BLKG-006(m2)의 뮤턴트 실행으로 관측한다.
- 통과 조건: 실패 메시지 본문에 문제의 키 문자열이 그대로 등장한다.

---

### AC-BLKG-003 — 실패가 따옴표 실수의 **방향**을 지목한다 (양방향)

**Given** 실패한 키 `k`의 따옴표 변형 중 하나가 추출 집합에 존재하고,
**When** 새 가드가 그 키에서 실패하면,
**Then** 실패 메시지는 **어느 방향의 실수인지** 지목하고 그 방향에 맞는 교정을 제시한다.

| 관측 | 메시지가 말해야 할 것 |
|---|---|
| `k`가 bare identifier이고 `"` + k + `"`가 집합에 있다 | 따옴표가 **벗겨졌다** → backtick raw string으로 적어라 |
| `k`가 따옴표에 감싸여 있고 그 알맹이가 집합에 있다 | 따옴표를 **덧붙였다** → 평범한 문자열 키로 적어라 |

- 판정 방식: m2(탈락)와 m2′(덧붙임) 두 뮤턴트의 실패 메시지 본문을 각각 인용해,
  **서로 다른 방향**을 지목하는지 읽는다. 두 뮤턴트가 같은 메시지를 내면 실패다 —
  방향을 못 가리는 진단은 「어딘가 틀렸다」와 다를 바 없다.
- 거울상을 포함하는 이유: 트리 `93fb36344`의 허용목록은 두 모양을 나란히 담고 있어, 다음 사람은
  **어느 쪽을 베끼든** 틀릴 수 있다. 한 방향만 막는 가드는 나머지 절반에서 침묵한다.
- 이 항목이 카드의 실제 불만이다. 실패하는 것만으로는 부족하고, **왜 실패했는지가 화면에 있어야** 한다.

---

### AC-BLKG-004 — 검사 대상이 0이면 초록을 주장하지 않는다

**Given** 허용목록에 실질 키가 하나도 없고,
**When** 새 가드를 실행하면,
**Then** 테스트는 통과하지 않는다 — 검사 대상 0의 초록은 아무것도 주장하지 않기 때문이다.

- 판정 명령(임시 뮤턴트: 허용목록 본문을 빈 맵으로):

      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s

- 통과 조건: 종료코드 ≠ 0, 그리고 실패 메시지가 「허용목록이 비어 있어 이 가드가 아무것도 검사하지 못한다」는 취지를 말한다.
- **함께 기록할 것**: 같은 뮤턴트에서 기존 가드 `TestBinaryLag_DoctorCheckNameSetIsUnchanged`도
  RED가 되는지 실행해 결과를 남긴다.

      go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s -v

  기대: 허용목록이 비면 `hookWiringCheckName`과 `` `"Hook Delivery"` ``가 더 이상 면제되지 않으므로
  기존 가드는 RED여야 한다.
- **[HARD] 초록이 보이면 세 갈래로 갈라 처분한다.** 「초록이면 blocker」로 뭉뚱그리면 오탐이 된다 —
  아래 두 갈래는 가드의 침묵이 아니라 **측정이 성립하지 않은** 경우다:

  | stdout에서 읽히는 것 | 무슨 일 | 처분 |
  |---|---|---|
  | `--- SKIP: TestBinaryLag_DoctorCheckNameSetIsUnchanged` | baseline blob(`22f90b1c7`)을 `git show`로 못 읽어 `t.Skipf`로 빠졌다 | 측정 미성립. **판정 불가**로 적고, blob 가용성을 복구한 뒤 다시 잰다 |
  | `[no tests to run]` | 셀렉터가 아무것도 고르지 못했다(이름 오타 등) | 측정 미성립. **판정 불가**로 적고 명령을 고쳐 다시 잰다 |
  | 뒤 공백형 `^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged ` 계수 = 1 | 진짜로 통과했다 | **이것만이 발견이다.** 기존 가드가 어딘가에서 침묵하고 있다는 뜻이므로 blocker로 보고한다 |

  **[HARD] 세 갈래는 두 매치 수로 가른다** — `[no tests to run]` 계수와 뒤 공백형 PASS 계수.
  `$`로 조인 패턴은 상시 0이라 진짜 PASS까지 미성립으로 오분류한다(EV-BLKG-002).
  SKIP은 `[no tests to run]`을 찍지 않으므로 1단만으로는 걸러지지 않는다(EV-BLKG-003) —
  뒤 공백형 계수가 0으로 떨어지는 것이 SKIP을 잡는 신호다.

  이 갈래가 필요한 이유: 이 환경에서는 baseline blob이 가용하고 두 이름 모두 baseline에 없으므로
  진짜 PASS는 **도달 불가**로 예상된다. 즉 초록이 보이면 대개 위 두 미성립 중 하나이며,
  그것을 blocker로 올리면 **없는 결함을 만든 것**이 된다.
- 원복 후 두 가드 모두 GREEN임을 AC-BLKG-001과 같은 2단 판정으로 재확인한다.

---

### AC-BLKG-005 — 뮤턴트 m1: 단언을 제거하면 가드가 죽는다

**Given** 새 가드의 1차 단언(키 실재 검사)을 제거한 뮤턴트가 있고,
**When** AC-BLKG-002의 무력 키를 심은 상태로 테스트를 실행하면,
**Then** 뮤턴트에서는 통과하고 원본에서는 실패한다 — 즉 그 단언이 하중을 진다.

- 절차: (1) 무력 키를 허용목록에 추가 → 원본에서 RED 관측 → (2) 단언 제거 뮤턴트 → GREEN 관측 →
  (3) 둘 다 원복 → GREEN 재확인.
- **[HARD] (2)와 (3)의 GREEN은 AC-BLKG-001과 **같은 두 매치 수**로 읽는다.** 단언을 제거한 뮤턴트에서
  종료코드 0만 보고 GREEN이라 적으면, 「이름이 어긋나 안 돌았다」와 「뮤턴트가 진짜로 통과했다」가
  구별되지 않는다 — EV-BLKG-001과 EV-BLKG-002가 그 두 신호(종료코드 `0`)가 동일함을 보인다.
  따라서 (2)·(3)은 `-v`로 실행해 stdout에서 **순서대로**:

  1. `[no tests to run]` 계수 = **0** (아니면 **판정 불가** — 뮤턴트 절차 전체를 다시 돌린다)
  2. `^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames ` (**뒤 공백**) 계수 = **정확히 1**

  `$`로 조인 패턴은 쓰지 않는다 — 통과 줄 끝의 ` (0.00s)` 때문에 상시 0이다(EV-BLKG-002).
  2단이 SKIP까지 잡는다(EV-BLKG-003).
- (1)의 RED는 종료코드 ≠ 0을 실패로 읽는 방향이므로 공허한 경우가 실패로 떨어져 안전하다.
- 통과 조건: 세 단계의 stdout(해당 판별 줄 포함)과 실패한 파일:줄을 모두 기록한다.

---

### AC-BLKG-006 — 뮤턴트 m2 / m2′: 두 방향의 따옴표 실수가 각각 RED이고, 메시지가 방향을 말한다

**Given** 허용목록 엔트리를 틀린 모양으로 적은 뮤턴트 둘이 있고 —
**m2**: 리터럴 등록 체크(`"Hook Delivery"`)의 키를 **따옴표 없이** `"Hook Delivery": true`로,
**m2′**: 상수 등록 체크(`hookWiringCheckName`)의 키를 **따옴표를 덧붙여** `` `"hookWiringCheckName"`: true ``로 —
**When** 각각에 대해 새 가드를 실행하면,
**Then** 둘 다 실패하고, 각 실패 메시지가 AC-BLKG-003 표의 **서로 다른 행**에 해당하는 진단을 낸다.

- 판정 명령(각 뮤턴트마다):

      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s

- 통과 조건: 두 실행 모두 종료코드 ≠ 0이고, 두 메시지가 **다른** 교정을 제시한다. 각 실패의
  파일:줄과 메시지 본문을 그대로 기록한다.
- **원복 후 GREEN은 AC-BLKG-001과 같은 두 매치 수로 읽는다** — `[no tests to run]` 계수 = **0**을
  **먼저**, 그다음 뒤 공백형 `^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames ` 계수 = **정확히 1**.
  `$`로 조인 패턴은 상시 0이라 쓰지 않는다(EV-BLKG-002). 종료코드는 판정에 쓰지 않는다.
- m2가 **이 카드의 핵심 증거**다. m1만으로는 「엔트리가 있으면 통과」까지만 보이고,
  「엔트리의 *모양*이 틀리면 없는 것과 같다」는 보이지 않는다(t477 `b2f98f6aa`의 관측과 같은 구조).
- **전제는 충족됐다**: 트리 `93fb36344`에 리터럴 등록 체크가 실재한다
  (`grep -c 'Hook Delivery' internal/cli/doctor.go` → `1`). 착지 전이었다면 이 항목은 판정 불가였고,
  그 경우 **판정 불가를 통과로 적지 않는다**는 규율이 적용됐을 것이다.

---

### AC-BLKG-007 — 폭발 반경은 테스트 파일 하나다

**Given** 이 SPEC의 run-phase가 끝났고,
**When** 흡수 병합 커밋 `93fb36344` 대비 diff를 세면,
**Then** 변경 파일은 `internal/cli/binary_lag_test.go` 하나이며, 프로덕션 코드 변경은 0이다.

- 판정 명령 — base **명시 핀** + **two-dot**:

      git diff --name-only 93fb36344..HEAD -- internal/ pkg/ cmd/ | wc -l
      git diff --stat      93fb36344..HEAD -- internal/ pkg/ cmd/

- 판정식: 첫 명령의 **파일 수**. 편집 전 `0`, 편집 후 정확히 `1`이며 그 1개가
  `internal/cli/binary_lag_test.go`여야 한다. (SPEC 문서 경로는 pathspec 밖이라 계수에 섞이지 않는다.)
- **[HARD] `origin/develop`을 base로 쓰지 말 것. three-dot을 쓰지 말 것.**
  `origin/develop`은 `25a3212a9`에 멈춰 있고(레인은 develop을 push하지 않는다) three-dot은
  merge-base를 거기까지 끌어내리므로, 두 경우 모두 t466·t477이 바꾼 24개 파일이 이 카드 몫으로
  집계된다. 그 상태에서는 run-phase가 **한 글자도 고치지 않아도** 이 항목이 거짓 실패한다.
  머리말의 실측 두 줄(24 vs 0)이 그 차이다.
- run-phase가 develop을 다시 흡수하면, 이 base 핀을 **그때의 새 병합 커밋 SHA로 갱신**한다.
  갱신하지 않으면 새로 흡수한 남의 카드 파일이 다시 계수에 섞인다.

---

### AC-BLKG-008 — 판정식은 매치 수이고, 새 가드는 과거 blob에 의존하지 않는다

**Given** 이 SPEC의 모든 검증 레시피와 새 가드 구현이 있고,
**When** 그것들을 읽으면,
**Then** (1) 어떤 판정도 `grep`의 종료 상태를 판정식으로 쓰지 않고, (2) 새 가드는 `git show`/`lagBaselineSHA`를 참조하지 않아 얕은 클론에서도 `Skip` 없이 판정된다.

#### (2)의 판정 — **2단 레시피**. 1단이 성립할 때에만 2단을 읽는다

**1단 — 비공허성 선단언. 검사 대상이 실재하는가:**

    awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -l

판정식: 출력이 **0보다 커야** 한다. `0`이면 함수가 없거나 이름이 어긋났거나 awk 범위가 빗나간
것이며, 그 경우 이 항목은 **판정 불가**다 — 통과로 적지 않고 원인(함수 부재 / 이름 불일치 /
범위 오류)을 밝혀 고친 뒤 다시 잰다.

**2단 — 1단이 0보다 클 때에만. baseline 참조가 0인가:**

    awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | grep -c 'lagBaselineSHA\|git show'

판정식: 출력이 `0`이면 통과. 종료코드는 읽지 않는다 — `grep -c`는 매치 0에서 stdout `0` +
exit `1`을 내므로, 종료코드를 게이트로 쓰면 판정이 정반대로 뒤집힌다.

**왜 2단인가 (실측)**: 트리 `93fb36344`에서 함수가 아직 없는데도 1단·2단을 합친 한 줄짜리 레시피는
요구값 `0`을 그대로 냈다.

    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -c
    0

빈 입력에 `grep -c`를 물리면 언제나 `0`이다. **검사 대상이 0인데 초록** — 이 카드가 막으려는
결함과 정확히 같은 모양이며, 함수명이 바뀌거나 awk 범위가 어긋나도 똑같이 초록이 된다.
1단이 그 침묵을 깬다.

#### (1)의 판정 — **열거 레시피**. 위반 수만이 아니라 분류한 모집단도 기록한다

**위반의 정의**(REQ-BLKG-005과 같은 문장): *비공허성 단언과 짝지어지지 않은 채,
종료코드 0을 통과로 읽는 판정 줄.* 종료코드 ≠ 0을 실패로 읽는 줄은 위반이 아니다 —
공허한 경우가 실패 쪽으로 떨어지기 때문이다. 증거로 종료코드를 **기록만** 하는 줄도 위반이 아니다.

**열거 대상 문서 3개**: `acceptance.md` · `plan.md` · `spec.md` (`progress.md`는 기록 문서이지
레시피 문서가 아니므로 제외).

**열거 절차**: 세 문서에서 **판정에 종료코드를 읽는 줄**을 전부 찾아 `file:line`으로 적고,
각 줄을 아래 셋 중 하나로 분류한다. 위반은 C뿐이다.

**[HARD] 분류는 줄의 *역할*로 한다 — 낱말 매치로 하지 않는다.** 「종료코드」라는 낱말이 든 줄을
훑어 몇 개 문구를 빼는 거친 필터는 금지다. 그 필터는 **판정하지 않는 줄을 위반으로 올린다** —
증거의 종류를 나열한 표 칸, 기록 의무를 지시하는 산문, 규칙 자체를 서술한 문장이 전부 걸린다.
각 줄을 열고 「이 줄이 통과/실패를 **결정하는가**」를 읽어 판단한다.

> 이것은 가정이 아니라 관측이다. 2차 감사 회차에서 낱말 기반 훑기가 실제로 오탐 2건을 냈다 —
> 증거 종류를 나열한 표 칸과 기록 의무를 지시한 산문이 판정 줄로 계수됐고, **총계가 우연히
> 맞아떨어져(5 = 5)** 검산을 통과할 뻔했다. 총계 일치는 목록 일치가 아니다.
> 그래서 결과를 수가 아니라 **`file:line` 목록**으로 남긴다 — 다음 사람이 각 줄을 열어
> 역할을 직접 확인할 수 있어야 하고, 그때에만 오탐이 드러난다.

| 분류 | 뜻 | 위반? |
|---|---|---|
| **A** | 종료코드 ≠ 0을 실패로 읽는다 | 아니오 (안전 방향) |
| **B** | `git merge-base --is-ancestor`의 0/1을 읽되, **0 / 1 / 그 밖** 3분기 단서가 붙어 있다 | 아니오 (머리말 참조) |
| **C** | 비공허성 단언 없이 종료코드 0을 통과로 읽는다 | **예 — 위반** |

**[HARD] 기대값은 `C = 0` 하나뿐이다.** 모집단과 A/B 건수는 기대값이 **아니며**, 판정할 때마다
세 문서를 처음부터 끝까지 다시 읽어 열거해 기록한다. 좌표에 기대값을 매어 두면 문서 한 줄만
고쳐도 부패한다 — 관측된 사실이다: 2차 판 좌표 `:62 / :110 / :147`이 3차에서 전부 밀렸고,
3차 판 좌표 `:243 / :290 / :307`도 이번 편집으로 다시 밀렸다.

**전회 열거 결과 (비권위적 seed — 기대값이 아니다.** 트리 `93fb36344`, 최종 수정판**)**:

    A: acceptance.md:246  (AC-BLKG-004 통과조건)
    A: acceptance.md:293  (AC-BLKG-005 (1)단계, 안전 방향임을 본문에 명시)
    A: acceptance.md:310  (AC-BLKG-006 통과조건)
    B: plan.md:74, plan.md:77  (§C 착지 판정, --is-ancestor)
    C: (없음)
    → 모집단 5 · A 3 · B 2 · C 0

**`spec.md`가 다시 0건인 이유 — 명시 판정**: 감사가 지적한 `spec.md` §4의 그 줄은 **C 위반이
맞았다**(비공허성 짝 없이 exit 0을 「기준선 GREEN」으로 읽었다). 이번 수정에서 그 줄을 두 매치 수
판독으로 **바꿔 없앴으므로**, 지금은 종료코드를 판정에 읽는 줄이 `spec.md`에 없다.
즉 이번의 0건은 **놓쳐서 0이 아니라 고쳐서 0**이며, 이 구분을 적어 두지 않으면 다음 열거가
같은 0을 보고 「전과 같다」고 넘어간다.

seed는 **다시 열거할 때의 출발점**일 뿐이며, seed와 재열거 결과가 다르면 **재열거 쪽이 옳다.**
seed를 그대로 옮겨 적고 「C 0건」이라 쓰는 것은 측정이 아니다.

**판정식**: C = 0이면 통과. 다만 **모집단이 0이면 판정 불가**다 — 아무것도 분류하지 못한 채 얻은
「위반 0」은 이 카드가 막으려는 바로 그 공허한 초록이다. 결과에는 세 수(모집단·A/B·C)를 모두 적는다.

**[HARD] 이 열거는 실제로 과소열거를 낸 적이 있다 — 아이러니를 기록해 둔다.**
바로 위 [HARD] 「낱말 매치 금지」 조항을 만들어 낸 그 열거가, 정작 **낱말이 없는 줄을 놓쳤다.**
최종 감사가 멤버십 검산으로 잡은 것:

| 놓친 줄 | 왜 안 보였나 | 처분 |
|---|---|---|
| `spec.md` §4의 `; echo "exit=$?"` → 「기준선 GREEN」 판정 | 열거 대상에 `spec.md`가 명시돼 있었는데도 모집단에 **`spec.md` 출신이 0건**이었다 | C 위반. 두 매치 수로 고쳤다 |
| AC-BLKG-006 원복 초록 · §D.1 「함께 GREEN」 | 「종료코드」라는 낱말이 없어 눈에 띄지 않았다 | 초록 읽기 자리. 두 매치 수를 붙였다 |

**오탐과 누락은 같은 원인의 양 방향이다.** 앞 회차에서 낱말 훑기가 판정하지 않는 줄을 위반으로
올렸고(오탐), 이번엔 같은 훑기가 판정하는 줄을 못 봤다(누락). 그래서 **훑기가 아니라 정독**이며,
열거 대상 문서가 목록에 있는데 그 문서 출신이 0건이면 **그 자체를 의심 신호로 삼는다.**

---

### AC-BLKG-009 — 판정식 자체가 실패할 수 있음을 실행으로 보인다

**Given** AC-BLKG-001 / 004 / 005가 쓰는 두 매치 수 판정식이 확정됐고,
**When** 그 판정식을 **존재하지 않는 테스트 이름**으로 한 번 돌리면,
**Then** 판정식이 실제로 실패한다 — 1단 계수가 1(빈 sweep)이거나 2단 계수가 0으로 떨어진다.

- 판정 명령(예: 이름에 `_NoSuchName`을 붙여):

      go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames_NoSuchName -count=1 -timeout 600s -v

  그 stdout에 대해 두 계수를 낸다. 기대: `[no tests to run]` → **1**, 뒤 공백형 PASS → **0**.
- 통과 조건: 두 계수가 위와 같고, 그 결과가 판정식에서 **판정 불가/실패**로 처리됨을 기록한다.
  두 계수가 통과값(0 / 1)으로 나오면 **판정식이 공허한 것이며**, 판정식을 고친 뒤 다시 돌린다.

- **왜 이 항목이 따로 있나 — 같은 함정이 이 카드에서 세 번 나왔다:**

  | 회차 | 공허했던 것 | 어떻게 드러났나 |
  |---|---|---|
  | 1차 D3 | `awk` 범위가 빈 추출을 내는데 `grep -c`가 요구값 `0`을 냄 | **실행** |
  | 2차 R1 | `-run` 0매치가 `ok` + exit 0을 냄 | **실행** |
  | 3차 | `$` 앵커가 진짜 통과에서도 상시 0 | **실행** |

  **셋 다 실행해 보고 나서야 드러났고, 읽어서 드러난 적은 한 번도 없다.** 그러므로 판정식을
  눈으로 검토하는 것으로는 이 항목을 대신할 수 없다 — 반드시 돌려야 한다.
  이 의무를 문서 본문이 아니라 **수락 기준**에 둔 이유가 이것이다. 문서에만 적힌 의무는 수행되지 않는다.

---

## §D.1 완료 정의 (Definition of Done)

- AC-BLKG-001 … AC-BLKG-009 전부 PASS이며, 각 항목이 **자기 명령과 출력**을 갖는다.
- **AC-BLKG-009가 먼저 통과해야 나머지의 GREEN을 읽을 수 있다** — 판정식이 실패할 수 있음을
  보이지 못한 상태의 초록은 해석되지 않은 출력이다.
- 뮤턴트 3종(m1 / m2 / m2′)의 RED가 각각 파일:줄과 메시지 본문으로 기록돼 있다.
- 기존 가드 `TestBinaryLag_DoctorCheckNameSetIsUnchanged`도 함께 GREEN — **이 초록도 두 매치 수로
  읽는다**: `[no tests to run]` 계수 = **0**을 먼저, 그다음 뒤 공백형
  `^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged ` 계수 = **정확히 1**.
  이 가드에는 `t.Skipf` 경로가 실재하므로(`binary_lag_test.go:202`) 2단이 특히 필요하다 — SKIP은
  1단을 통과해 버린다(EV-BLKG-003).
- `progress.md` §E.2 / §E.3에 명령·출력·종료코드가 그대로 남아 있다.
- 판정 불가 항목(전제 미착지 등)이 있다면 **통과로 합산하지 않고** 사유를 명시한다.
