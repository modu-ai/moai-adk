# SPEC-BINLAG-KEYGUARD-001 — 진행 기록

- 카드: **t479**
- 워크트리: `.claude/worktrees/t479` · 브랜치 `WT-binarylag-key-guard`
- base 핀: **`93fb36344`** — 로컬 `develop @ a825183dd` 흡수 병합 커밋 (부모: `25a3212a9` + `a825183dd`)
- 단계: plan (run 미착수). 1차 FAIL 0.86 → D1/D2/D3 + D4/D5/D6/D8 · 2차 FAIL 0.87 → R1/R2 + R4/R5 · **최종 수정 → R6 + 판정식 정정 + AC-BLKG-009 신설**
- 3차(최종) 감사 **PASS · 0.90** (Tier M 임계 0.80) — R1/R2/R4/R5/R6 전부 RESOLVED.
  **부채를 안고 통과**이며 진입 조건 2건(T2 형제 스윕 · T1 과소열거)을 이 판에서 닫았다. T3/R3도 함께.
- 운영자 결정: 최종 수정 후 **감사 없이 run 진입**

---

## §E.1 Plan-phase Audit-Ready Signal

### Claim

plan-phase 산출물 4종(`spec.md` / `plan.md` / `acceptance.md` / `progress.md`)을 작성하고,
1차 감사 FAIL 0.86의 D1/D2/D3 + D4/D5/D6/D8, 2차 감사 FAIL 0.87의 R1/R2 + R4/R5를 반영했다.
**2차 라운드에서 철회한 것이 있다**: 1차 수정에서 신설한 「`go test` 종료코드 예외」가 실측으로
거짓이어서, 그 조항을 폐기하고 REQ-BLKG-005 자체를 다시 썼다(자세히는 아래 Evidence).
요구 7건(REQ-BLKG-001 … 007), 수락 8건(AC-BLKG-001 … 008) — 수는 유지되고 REQ-BLKG-003이
**양방향**으로 넓어졌다. 수리 방식은 **(a) 허용목록 키-모양 가드**이며 기각안 (b)와 사유는
`plan.md` §B에 있다. 변경 범위는 `internal/cli/binary_lag_test.go` 한 파일이고 프로덕션 코드는
건드리지 않는다. 전제(t466·t477)는 **착지했고 흡수도 끝났다**.

### Evidence

이 워크트리, HEAD **`93fb36344`**에서 실행한 명령과 그 출력 그대로.

    $ git rev-parse --short HEAD
    93fb36344

    $ git log -1 --format='%h %p %s' 93fb36344
    93fb36344 25a3212a9 a825183dd Merge develop a825183dd into WT-binarylag-key-guard (t479): absorb t466 Hook Delivery check + t477 allowlist entry

**전제 착지 — ref가 갈린다 (D2):**

    $ git merge-base --is-ancestor 4e91bf6a9 origin/develop; echo $?
    1          ← origin/develop 기준: 조상 아님
    $ git merge-base --is-ancestor 4e91bf6a9 develop; echo $?
    0          ← 로컬 develop 기준: 조상
    $ git rev-list --count origin/develop..develop
    130
    $ git merge-base --is-ancestor a825183dd HEAD; echo $?
    0          ← 흡수한 develop tip이 이 HEAD의 조상

**리터럴 체크는 이제 실재한다 — 매치 수 1:**

    $ grep -c 'Hook Delivery' internal/cli/doctor.go
    1

**diff base 선택이 24 대 0을 가른다 (D1):**

    $ git diff --name-only origin/develop...HEAD -- internal/ pkg/ cmd/ | wc -l
    24
    $ git diff --name-only 93fb36344..HEAD -- internal/ pkg/ cmd/ | wc -l
    0

**기존 가드는 흡수 후에도 GREEN:**

    $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s; echo "exit=$?"
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.947s
    exit=0

**허용목록의 현재 내용 — 두 모양이 나란히 있다** (`internal/cli/binary_lag_test.go:194`):

    var namesAddedAfterBaseline = map[string]bool{
    	"hookWiringCheckName": true,
    	`"Hook Delivery"`:     true,
    }

**AC-BLKG-008(2)의 공허성 실증 (D3)** — 함수가 없는데 요구값 `0`이 나온다:

    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -c
    0
    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -l
    0

**좌표(트리 `93fb36344`에서 재측정, D4):**

    $ grep -n 'func checkNamesFromSource\|func exprSource\|var namesAddedAfterBaseline\|func TestBinaryLag_DoctorCheckNameSetIsUnchanged\|namesAddedAfterBaseline\[name\]\|lagBaselineSHA =' internal/cli/binary_lag_test.go
    24:const lagBaselineSHA = "22f90b1c7"
    118:func checkNamesFromSource(t *testing.T, src []byte) map[string]bool {
    162:func exprSource(fset *token.FileSet, src []byte, e ast.Expr) string {
    194:var namesAddedAfterBaseline = map[string]bool{
    199:func TestBinaryLag_DoctorCheckNameSetIsUnchanged(t *testing.T) {
    213:		if beforeNames[name] || namesAddedAfterBaseline[name] {

`:184/:188/:202`(최초 작성 시 `25a3212a9` 좌표)는 병합으로 밀려 `:194/:199/:213`이 됐다.

**2차 감사 R1 — 러너 종료코드가 안전하다는 주장은 거짓이다** (트리 `93fb36344`, 새 가드 부재):

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.704s [no tests to run]
    (exit 0)

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s -v
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.084s [no tests to run]
    (exit 0)

`-v`의 맨 아래 `PASS` 한 줄까지 나온다는 것이 특히 중요하다 — `PASS`만 보는 판별식도 공허하다.

**대조군** — 판별식이 공허하지 않음을 보인다(실재하는 테스트 이름, 같은 트리):

    $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s -v
    === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.06s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.990s
    (exit 0)

두 출력 모두 종료코드 0이며, 갈리는 것은 stdout의 두 토큰뿐이다:
`--- PASS: <이름>`(실재할 때만) / `[no tests to run]`(0매치일 때만). 이것이 새 판정식의 근거다.

**R5 — `--is-ancestor`의 단서** (같은 트리):

    $ git merge-base --is-ancestor deadbee9 develop
    fatal: Not a valid object name deadbee9
    (exit 128)

유효 ref에 한해서만 0/1이 조상 여부를 뜻한다. 그래서 0 / 1 / 그 밖 3분기로 읽도록 고쳤다.

**R2-3 — AC-BLKG-008(1) 열거를 실제로 수행해 기대값을 확정했다** (수정 후 트리):
모집단 **5건** — A(≠0을 실패로) 3건 `acceptance.md:165, 203, 220` · B(`--is-ancestor` 3분기) 2건
`plan.md:74, 77` · **C(위반) 0건**. 각 좌표는 `sed -n '<n>p'`로 해당 줄을 실제로 출력해 확인했다.

**낱말 매치가 오탐을 낸다 — 관측 사례** (2차 감사 회차, 리드가 스스로 정정):
「종료코드」가 든 줄을 훑고 몇 문구만 빼는 필터가 `plan.md:122`(증거 종류를 나열한 표 칸)과
`plan.md:151`(기록 의무 산문)을 판정 줄로 계수했다. 실제 판정 줄은 `plan.md:69, 72`
(`--is-ancestor`)였다. **총계가 우연히 5로 같아** 검산을 통과할 뻔했다 — 총계 일치는 목록 일치가
아니다. 이 관측을 근거로 AC-BLKG-008(1)에 [HARD] 「분류는 줄의 *역할*로, 낱말 매치 금지」 조항과
`file:line` 목록 보존 의무를 넣었다.

**멤버십이 리드의 목록과 다른 것은 정상이다** — 리드의 좌표는 수정 *전* 판이고 내 것은 *후* 판이다.
`acceptance.md:62`(구 AC-BLKG-001 판정 줄)는 R1 수정으로 **위반 자체가 사라졌고**,
`acceptance.md:203`(AC-BLKG-005 (1)단계 안전 방향 명시)은 이번에 **새로 생긴 A 분류**다.
수정 전후로 총계가 둘 다 5인 것은 우연이며, 그래서 수를 맞춰 보는 것으로 검산하지 않았다.
현 트리의 `plan.md:122 / :151`은 편집으로 밀려 각각 빈 줄과 `**M4 …**`이며, 이것이 리드의 좌표가
수정 전 판임을 독립적으로 뒷받침한다.

**최종 회차 — 판정식 정정 3건, 전부 이 트리에서 직접 쟀다:**

① **`$` 앵커는 진짜 통과에서도 상시 0이다.** 실재하고 실제로 통과하는 테스트로:

    $ grep -e '--- PASS' OUT | cat -e
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.06s)$
    $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged$' OUT
    0          ← 통과했는데 0
    $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged ' OUT
    1          ← 뒤 공백형

`cat -e`로 줄 끝을 확인해 원인을 특정했다: go가 ` (0.06s)`를 붙인다. 따라서 `$` 형태를
「통과 = 1」로 쓰면 **가드가 완벽히 작동해도 항상 실패**로 읽힌다.

② **1단(`[no tests to run]` = 0)만으로는 SKIP을 못 막는다.** `t.Skipf` 경로를 편집 없이 관측하려고
같은 패키지의 `-short` 스킵 테스트를 표본으로 골라 돌렸다:

    $ go test ./internal/cli/ -run TestHookCommandFlushesLastHandlerEntry -count=1 -short -timeout 600s -v
    --- SKIP: TestHookCommandFlushesLastHandlerEntry (0.00s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.708s
    (exit 0)
    [no tests to run] 계수 → 0      ← 1단을 통과해 버린다
    뒤 공백형 PASS 계수  → 0      ← 2단이 잡는다

이 카드에서 실재하는 그 경로: `internal/cli/binary_lag_test.go:202`의 `t.Skipf`.

③ **대상 이름으로도 두 계수를 직접 확인했다** (새 가드 부재 상태):

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s -v
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.276s [no tests to run]
    (exit 0)
    [no tests to run] 계수 → 1
    뒤 공백형 PASS 계수  → 0

세 실행 **모두 종료코드 0**이다. 그것이 종료코드를 판정에서 뺀 근거이며, 세 원장
(EV-BLKG-001/002/003)이 그 쌍을 보존한다.

**AC-BLKG-008(1) 재열거 — 최종판.** 이전 좌표를 재사용하지 않고 세 문서를 처음부터 끝까지 다시
읽어 열거했다. 각 좌표는 `sed -n '<n>p'`로 출력해 **역할을** 확인했다(낱말 매치가 아니라).

    A: acceptance.md:246  (AC-BLKG-004 통과조건)
    A: acceptance.md:293  (AC-BLKG-005 (1)단계)
    A: acceptance.md:310  (AC-BLKG-006 통과조건)
    B: plan.md:74, plan.md:77  (§C 착지 판정, --is-ancestor)
    C: (없음)
    → 모집단 5 · A 3 · B 2 · C 0

**최종 감사 T1 — 이전 열거는 과소열거였다.** `spec.md` §4가 `; echo "exit=$?"`의 `0`만 읽고
「기준선은 여전히 GREEN」이라 판정하고 있었다. **C 위반이 맞다**(비공허성 짝 없음). 이번에 그 줄을
두 매치 수 판독으로 바꿔 없앴으므로 지금의 `spec.md` 0건은 **놓쳐서 0이 아니라 고쳐서 0**이다.
바뀐 뒤의 실측:

    [no tests to run] 계수 → 0
    뒤 공백형 PASS 계수  → 1

**기록해 둘 아이러니**: [HARD] 「낱말 매치 금지」 조항을 만들어 낸 **바로 그 열거**가 낱말 없는
줄들을 놓쳤다. 앞 회차의 오탐(판정하지 않는 줄을 위반으로 올림)과 이번 누락(판정하는 줄을 못 봄)은
**같은 원인의 양 방향**이다. 그래서 열거는 훑기가 아니라 정독이며, 열거 대상 문서가 목록에 있는데
그 문서 출신이 0건이면 그 자체를 의심 신호로 삼도록 항목에 적었다.

**좌표는 seed로 강등했다** — 기대값은 `C = 0` 하나뿐이다. 근거는 관측이다: 2차 판 좌표
`:62 / :110 / :147`이 3차에서 밀렸고, 3차 판 `:243 / :290 / :307`이 이번에 또 밀렸다.
두 번 연속 부패한 값을 기대값으로 두는 것은 측정이 아니다.

**T2 형제 표면 스윕** — 초록을 읽는 자리를 문서 전체에서 훑어 네 곳에 두 매치 수를 붙였다:
`acceptance.md` AC-BLKG-006 원복 · §D.1 기존 가드 · `plan.md` §E 표 「기존 가드 무변경」 ·
`plan.md` §F M3 원복. 감사가 지적한 것은 앞의 둘이며, **뒤의 둘은 같은 클래스를 훑다가 찾았다** —
이 카드의 반복 패턴(한 자리를 고치고 형제를 안 훑음)을 이번엔 형제 클래스 전체로 닫았다.

**SPEC ID 정규식 자가검사:**

    $ ID="SPEC-BINLAG-KEYGUARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
    PASS

### Baseline-attribution

위 측정 전부 **이 실행에서, 이 트리(`.claude/worktrees/t479` @ `93fb36344`)에 대해** 냈다.
1차 감사가 보고한 값(24 / 0 / 130 / ancestor 1·0 / awk 0바이트)과 2차 감사가 보고한 값
(0매치 `ok` + exit 0)을 **그대로 옮겨 적지 않고 각각 다시 실행해** 같은 값을 관측했다.
`-v` 출력, 대조군, `--is-ancestor` 오타 ref, 열거 좌표 5건은 감사가 주지 않은 값이며
이 세션이 새로 쟀다. 다른 트리·다른 시점에서 옮겨온 값은 없다.

세 가지는 **건네받은 값**이며 그렇게 표기한다:

- 카드 t477(`WT-binarylag-allowlist`, `b2f98f6aa`)의 뮤턴트 2건(m1 → `:215` FAIL, m2 → `:216` FAIL) —
  **이 세션이 재현하지 않았다.** 그 창의 판정을 전제로 인용할 뿐이며, AC-BLKG-005/006은 run-phase가
  자기 트리에서 다시 재도록 요구한다.
- 흡수를 수행한 주체는 리드이며, 이 세션은 그 결과(HEAD·부모·조상 관계)를 사후에 관측했을 뿐이다.
- plan-audit의 점수 0.86과 그 판정 근거는 `.moai/reports/t479/plan-audit.md` 소관이다.

### Gaps

명시적으로 **관측하지 않은** 것들.

- **새 가드의 동작 일체**: 아직 코드가 없다. AC-BLKG-001 … 006의 GREEN/RED는 전부 미관측이다.
  m2′(따옴표 덧붙임)는 이번에 새로 요구한 방향이며, 어느 쪽도 아직 실행되지 않았다.
- **기존 가드가 빈 허용목록에서 RED가 되는지**(D8) — 뮤턴트를 심지 않았으므로 미관측이다.
  run-phase의 AC-BLKG-004에서 실제로 잰다.
- **t466 / t477 자체의 정확성**: 재검증하지 않았다. 범위 밖(`spec.md` §6).
- **전체 스위트**: `go test ./...`를 돌리지 않았다. 로컬 전체 스위트는 금지이며 판정은 CI 몫이다.
- **재흡수 유보 1건 (의도적 결정)**: 로컬 `develop`은 흡수한 `a825183dd` 이후에도 움직였다.
  그럼에도 **재흡수하지 않기로 했다** — 새 가드의 입력은 `doctor.go`의 세 레지스트리와 허용목록
  둘뿐이고 그 구간에서 두 파일이 무변경이므로, 재흡수는 판정 대상을 바꾸지 않으면서 `93fb36344`
  핀만 무효화해 1차 D1과 같은 손해(남의 카드 파일이 diff 계수에 섞임)를 자초한다.
  **남는 위험**: 최종 병합 창에서의 **의미 충돌**은 이 판단이 배제하지 못한다 — 파일이 겹치지
  않아도 다른 카드가 `doctor.go`에 체크를 더하면 허용목록과의 관계가 달라진다.
  따라서 **병합 창에서 병합 트리 재측정이 필요하며**, 그때 AC-BLKG-007의 base 핀도 함께 갱신한다.

### Residual-risk

- **base 핀이 낡을 위험**: AC-BLKG-007의 판정은 `93fb36344`라는 **고정 SHA**에 매여 있다.
  run-phase가 develop을 다시 흡수하면 이 핀은 즉시 낡고, 갱신하지 않으면 새로 들어온 남의 카드
  파일이 다시 계수에 섞인다. acceptance.md에 갱신 의무를 적어 뒀으나, 적어 둔 것이 지켜지는지는
  이 항목이 보증하지 못한다.
- **양방향 진단의 취약성**: AC-BLKG-003의 두 분기는 「따옴표 변형 중 하나가 집합에 있다」는 조건에
  얹혀 있다. 등록 방식이 나중에 한쪽으로 통일되면 한 분기가 조용히 도달 불가가 될 수 있다 —
  가드는 고장 나도 침묵한다.
- **[HARD] 같은 함정이 이 카드에서 세 번 나왔다. 네 번째가 없다는 보장은 없다.**

  | 회차 | 공허했던 것 | 어떻게 드러났나 |
  |---|---|---|
  | 1차 D3 | `awk` 빈 추출에 `grep -c`가 요구값 `0`을 냄 | **실행** |
  | 2차 R1 | `-run` 0매치가 `ok` + exit 0을 냄 | **실행** |
  | 최종 | `$` 앵커가 진짜 통과에서도 상시 0 | **실행** |

  **셋 다 실행해 보고 나서야 드러났고, 읽어서 드러난 적은 한 번도 없다.** 앞의 둘은 수정 라운드가
  심은 것이고, 두 번째는 첫 번째를 고치며 쓴 예외 조항 자체가 원인이었다. 세 번째는 문서에
  들어가기 전 회신 단계에서 잡혔다.

  그래서 **AC-BLKG-009**를 신설했다 — run-phase는 판정식을 적용한 뒤 **존재하지 않는 이름으로
  한 번 돌려 판정식이 실제로 실패하는지 확인해야 한다.** 이 의무를 문서 본문이 아니라 수락 기준에
  둔 이유는 하나다: 문서에만 적힌 의무는 수행되지 않는다.

  다만 AC-BLKG-009도 **이미 아는 축(이름 축)만** 검사한다. 아직 이름 붙지 않은 네 번째 공허 형태는
  이 항목도 잡지 못한다.

- **[HARD] 이번 판이 또 결함을 심었다면 잡아낼 감사 라운드가 없다.** 운영자는 이 사실을 알고
  「최종 수정 1회 → 마지막 감사 → 판정과 무관하게 run 진입」을 택했다. 즉 남은 검증 지점은
  마지막 감사 한 번과 run-phase의 실행뿐이며, run-phase가 AC-BLKG-009를 실제로 돌리는 것이
  이 카드에 남은 **마지막 기계적 안전망**이다.
- **행 번호 좌표의 부패**: AC-BLKG-008(1)의 기대값이 `file:line` 5개에 매여 있다. 문서를 한 줄만
  고쳐도 밀리며, 밀린 좌표로 「C 0건」을 적으면 그것이 곧 공허한 초록이다. 항목에 재열거 의무를
  적었으나 지켜지는지는 이 항목이 보증하지 못한다.
- **범위 번짐**: 등록 방식을 통일하고 싶은 유혹이 프로덕션 코드로 번질 수 있다. `spec.md` §6이
  이를 금하고, AC-BLKG-007이 diff 파일 수로 기계적으로 판정한다.

---

## §E.2 Run-phase Evidence

### R0 — 착수 전 기록: 채택안과 기각안 (코드 편집 전에 적는다)

이 절은 run-phase가 **첫 코드 편집을 하기 전에** 적었다. 결정은 plan 시점에 내려진 것이며
여기서 다시 다투지 않고 기록만 한다.

**채택 — (a) 허용목록 키-모양 가드.** `namesAddedAfterBaseline`의 모든 키가, 현재
`internal/cli/doctor.go`에서 `checkNamesFromSource`가 실제로 만들어 내는 이름인지 단언하는
새 가드를 더한다. 키 하나가 아무것도 매치하지 못하면 그 키는 증명 가능하게 무력이며, 가드가
그것을 실패로 알린다.

**기각 — (b) `exprSource` 정규화(두 등록 방식이 한 모양을 내게 하기).**
기각 사유는 「더 위험해서」가 아니라 **실행 불가능해서**다. 기존 가드
`TestBinaryLag_DoctorCheckNameSetIsUnchanged`의 `beforeNames`는 현재 트리가 아니라
**과거 blob**(`lagBaselineSHA = "22f90b1c7"`)을 `git show`로 읽어 파싱한다. 그 시점에는
오늘 존재하는 상수가 아예 없을 수 있다. 정규화를 넣으면 before 집합과 after 집합의 의미가
갈라지고, 개선하려던 그 가드가 먼저 깨진다.

### R1 — Mode Selection 및 착수 상태

Phase 4 모드는 `serial`이며 근거는 §F에 적었다.

**착수 시점 트리 상태:**

    $ git rev-parse --short HEAD
    93fb36344
    $ git branch --show-current
    WT-binarylag-key-guard

### R2 — AC-BLKG-009 먼저: 판정식 자신이 실패할 수 있음을 실행으로 보였다

완료 정의가 요구하는 대로 **다른 어떤 초록보다 먼저** 돌렸다.

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames_NoSuchName -count=1 -timeout 600s -v
    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.237s [no tests to run]
    (exit 0)

    $ grep -c '\[no tests to run\]' ac009.out
    1
    $ grep -c '^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames_NoSuchName ' ac009.out
    0

1단 계수 **1**(빈 sweep) · 2단 계수 **0**. 기대와 같다. 이 결과는 판정식에서 **판정 불가**로
처분되며 통과로 세지 않는다. 종료코드는 `0`이었고 — 그것이 종료코드를 판정에서 뺀 이유다.
**이 항목이 통과했으므로 아래의 초록들을 읽을 수 있다.**

### R3 — 구현: `TestBinaryLag_AllowlistKeysAreLiveNames`

`internal/cli/binary_lag_test.go` 한 파일에 가드를 더했다. 입력은 현재 `doctor.go` 하나뿐이며,
과거 blob도 `git show`도 참조하지 않는다.

- **키 실재**: 허용목록의 모든 키가 `checkNamesFromSource`의 추출 집합 원소인지 단언한다.
- **비공허성**: 실질 키가 0이면 `t.Fatal`로 「아무것도 검사하지 못했다」고 말하며 실패한다.
- **양방향 진단**: 실패한 키의 따옴표 변형이 집합에 있으면 **어느 방향**의 실수인지
  (ADDED / STRIPPED) 지목하고 그에 맞는 교정을 제시한다. 어느 쪽도 아니면 「어떤 이름도
  매치하지 않는다」고 말한다.

### R4 — RED: 무력 키를 심으면 실패한다 (AC-BLKG-002 · AC-BLKG-005 (1)단계)

허용목록에 `"NoSuchCheckEver": true`를 심었다.

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    --- FAIL: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
        binary_lag_test.go:260: allowlist key "NoSuchCheckEver" matches no registered check name in doctor.go at all, in either quoting shape; it allows nothing and is provably inert. Remove it, or correct it to a name checkNamesFromSource extracts.
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.579s
    (exit 1)

실패 메시지 본문에 문제의 키 문자열 `NoSuchCheckEver`가 그대로 등장하고, 실패 위치가
`binary_lag_test.go:260`으로 특정된다. `≠ 0`을 실패로 읽되 **실패의 정체**(메시지 본문 + 파일:줄)를
함께 얻었으므로 빌드 실패·setup 실패와 갈린다.

### R5 — 뮤턴트 m1: 단언을 제거하면 가드가 죽는다 (AC-BLKG-005)

무력 키를 심은 채 키 실재 단언(`for key := range …` 루프)을 제거했다.

**첫 시도는 컴파일되지 않았다** — `live`가 미사용이 되어 `binary_lag_test.go:227:2: declared and
not used: live`로 빌드 실패했다. 빌드 실패는 뮤턴트의 GREEN이 아니므로 `_ = live`를 넣어
컴파일되는 뮤턴트로 고쳐 다시 돌렸다. (기록해 두는 이유: 그 상태의 `exit=1`을 「뮤턴트도 RED」로
읽었다면 m1이 거짓으로 통과했을 것이다.)

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s -v
    === RUN   TestBinaryLag_AllowlistKeysAreLiveNames
    --- PASS: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.183s
    (exit 0)

    [no tests to run] 계수 → 0      ← 1단 통과
    뒤 공백형 PASS 계수  → 1      ← 2단 통과

뮤턴트 GREEN · 원본 RED(R4). 그 단언이 **하중을 진다.**

원복(뮤턴트 + 무력 키 둘 다) 후 재확인: `[no tests to run]` 계수 **0** · 뒤 공백형 PASS 계수 **1**.

### R6 — 뮤턴트 m2 / m2′: 두 방향이 각각 RED이고, 메시지가 방향을 말한다 (AC-BLKG-003 · 006)

**m2 — 리터럴 등록 체크의 따옴표를 떨어뜨림** (backtick raw string → 평범한 문자열):

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    --- FAIL: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
        binary_lag_test.go:257: allowlist key "Hook Delivery" matches no registered check name, but "\"Hook Delivery\"" does: the quotes were STRIPPED. doctor.go registers this check as a string literal and checkNamesFromSource keeps those quotes as characters, so write the entry as a backtick raw string: `"Hook Delivery"`.
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.155s
    (exit 1)

**m2′ — 상수 등록 체크에 따옴표를 덧붙임** (평범한 문자열 → backtick raw string):

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    --- FAIL: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
        binary_lag_test.go:253: allowlist key "\"hookWiringCheckName\"" matches no registered check name, but "hookWiringCheckName" does: the quotes were ADDED. doctor.go registers this check through a constant identifier, so write the entry as a plain string key "hookWiringCheckName".
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.373s
    (exit 1)

**두 메시지는 다르다.** 하나는 `STRIPPED` + 「backtick raw string으로 적어라」, 다른 하나는
`ADDED` + 「따옴표 없는 평범한 문자열 키로 적어라」. 실패 줄도 갈린다(`:257` vs `:253`).

**이 카드의 결함 모양이 여기서 네 번째로 나왔다 — 실행으로만 드러났다.** m2의 **첫 실행**에서
메시지가 이렇게 나왔다:

    binary_lag_test.go:254: allowlist key "Hook Delivery" matches no registered check name, but "Hook Delivery" does: ...

키를 `%q`로, 추출 집합의 원소를 `%s`로 렌더한 결과 **두 피연산자가 글자 그대로 같아졌다** —
`%q`의 `Hook Delivery`도 `"Hook Delivery"`, `%s`의 `"Hook Delivery"`도 `"Hook Delivery"`다.
방향을 말한다면서 정작 두 대상을 구별하지 못하는 메시지였다. 두 피연산자를 **모두 `%q`로**
렌더하도록 고쳤고(`"\"Hook Delivery\""`), 그 근거를 코드 주석에 남겼다.
**읽어서는 보이지 않았고 돌려 보고서야 보였다** — §Residual-risk가 경고한 네 번째 회차다.

원복 후 재확인: `[no tests to run]` 계수 **0** · 뒤 공백형 PASS 계수 **1**.

### R7 — AC-BLKG-004: 검사 대상이 0이면 초록을 주장하지 않는다

허용목록 본문을 빈 맵(`map[string]bool{}`)으로 바꿨다.

**새 가드:**

    $ go test ./internal/cli/ -run TestBinaryLag_AllowlistKeysAreLiveNames -count=1 -timeout 600s
    --- FAIL: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
        binary_lag_test.go:232: namesAddedAfterBaseline holds no substantive key, so this guard checked nothing; a green over an empty subject asserts nothing about key shape
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.153s
    (exit 1)

**함께 기록할 것 — 같은 뮤턴트에서 기존 가드:**

    $ go test ./internal/cli/ -run TestBinaryLag_DoctorCheckNameSetIsUnchanged -count=1 -timeout 600s -v
    === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
        binary_lag_test.go:283: this SPEC added doctor check name hookWiringCheckName; REQ-BLV-009 rewires the existing "Binary Freshness" item and registers no new name
        binary_lag_test.go:283: this SPEC added doctor check name "Hook Delivery"; REQ-BLV-009 rewires the existing "Binary Freshness" item and registers no new name
    --- FAIL: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.07s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.162s
    (exit 1)

    [no tests to run] 계수                    → 0
    뒤 공백형 --- PASS: <이름> 계수            → 0
    --- SKIP: <이름> 계수                      → 0

**세 갈래 처분**: 초록이 아니라 **RED**다. `[no tests to run]` 0이므로 셀렉터는 매치했고,
`--- SKIP:` 0이므로 baseline blob도 읽혔다. 즉 측정은 성립했고 결과가 RED다 —
기대대로이며 **blocker 없음**. (초록이었다면 blocker 보고 대상이었다.)

원복 후 두 가드 모두 GREEN(R8).

### R8 — 원복 후 최종 GREEN (AC-BLKG-001 · §D.1)

    $ go test ./internal/cli/ -run TestBinaryLag -count=1 -timeout 600s -v
    --- PASS: TestBinaryLag_OneSeamServesBothSurfaces (0.09s)
    --- PASS: TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero (0.00s)
    --- PASS: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.07s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.154s
    (exit 0)

    [no tests to run] 계수                                        → 0   ← 1단 (먼저)
    ^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames  계수      → 1   ← 2단
    ^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged  계수  → 1   ← 2단 (기존 가드)

허용목록은 원본 그대로이며 두 등록 모양을 나란히 담고 있다. 즉 이 항목은 **두 모양 모두**에
대해 실재를 단언했다.

### R9 — AC-BLKG-008 (2): 새 가드는 과거 blob에 의존하지 않는다

**1단 — 비공허성 선단언 (0보다 커야 한다):**

    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | wc -l
    47

**2단 — baseline 참조 계수 (0이어야 한다):**

    $ awk '/^func TestBinaryLag_AllowlistKeysAreLiveNames/,/^}/' internal/cli/binary_lag_test.go | grep -c 'lagBaselineSHA\|git show'
    0

1단이 47 > 0으로 성립했으므로 2단을 읽을 수 있고, 2단이 0이므로 통과다.
`grep -c`의 종료코드는 읽지 않았다(매치 0에서 exit 1을 내므로 게이트로 쓰면 뒤집힌다).

### R10 — AC-BLKG-008 (1): 열거 (재열거, seed를 옮겨 적지 않았다)

**[HARD] 열거 도중 세 문서가 이 세션 밖에서 바뀐 것을 관측했다.** run 착수 시 읽은
`acceptance.md`는 470줄이었고, 열거 시점에는 **489줄**이었다(`spec.md`에 0.5.0 항목이 생기고,
좌표가 **기대값에서 seed로 강등**됐다). SPEC 산출물은 아직 untracked라 상태 조회에 수정으로
잡히지 않으므로, 파일을 다시 읽지 않았다면 낡은 판을 근거로 열거했을 것이다.
따라서 아래 열거는 **현재 판**(acceptance.md 489줄 / plan.md 178줄 / spec.md 237줄) 기준이다.

두 축으로 훑은 뒤 각 줄을 열어 **역할**로 분류했다(낱말 매치 금지 조항 준수):
① 종료코드 토큰 축(`종료코드` / `exit` / `$?` / `종료 상태`) — 후보 58줄,
② 판정 자리 축(`통과 조건` / `판정식` / `판정 명령` / `판정 방식`) — 후보 31줄.

**모집단 5 · A 3 · B 2 · C 0**

    A: acceptance.md:246  (AC-BLKG-004 통과조건 — ≠0 + 실패 메시지 취지 요구)
    A: acceptance.md:293  (AC-BLKG-005 (1)단계, 안전 방향임을 본문에 명시)
    A: acceptance.md:310  (AC-BLKG-006 통과조건 — ≠0 + 두 메시지가 다를 것)
    B: plan.md:74, plan.md:77  (§C 착지 판정, --is-ancestor, 0/1/그 밖 3분기 단서 있음)
    C: (없음)

제외한 것과 사유(대표):
`acceptance.md:77 / :96 / :115`는 증거 원장의 **기록** 필드, `:152 / :160 / :314 / :366`은
「종료코드를 판정에 쓰지 않는다」는 **금지** 서술, `:382 / :383 / :404 / :406`은 **규칙 정의**,
`plan.md:31-33 / :158 / :166`은 규칙·기록 의무, `spec.md:130 / :171`의 `(exit 0)`은
verbatim transcript 안의 기록이다. 어느 것도 통과/실패를 **결정하지 않는다.**

**`spec.md` 출신 0건 — 의심 신호로 다뤘다.** [HARD] 조항이 요구하는 대로 그냥 넘기지 않고
§4를 직접 읽어 확인했다: 종전 C 위반이던 「exit 0만 읽고 기준선 GREEN」 판정 줄이
두 매치 수 판독(`spec.md:173-176`)으로 **대체돼 사라졌다**. 즉 **놓쳐서 0이 아니라 고쳐서 0**이며,
이는 `acceptance.md:423-424`의 서술과 일치한다.

**seed와 결과가 같았다.** 다만 이것은 seed를 옮겨 적어서가 아니라 다시 열거해 같은 값을 얻은
것이며, 좌표 5개를 각각 출력해 역할을 읽어 확인했다.

### R11 — AC-BLKG-007: 폭발 반경은 테스트 파일 하나다

**편집 전** (base 핀 `93fb36344`, two-dot) — 대상 pathspec `internal/ pkg/ cmd/`:

    파일 수 → 0

**커밋 후** (같은 base 핀, 같은 two-dot, 같은 pathspec):

    파일 수 → 1
    internal/cli/binary_lag_test.go
    (stat) internal/cli/binary_lag_test.go | 70 +++++++++  1 file changed, 70 insertions(+)

파일 **1개**, 그리고 그 1개가 `internal/cli/binary_lag_test.go`다. 프로덕션 코드 변경 **0**.
base는 `93fb36344`로 명시 핀했고 **two-dot**을 썼다(three-dot이나 `origin/develop`을 쓰면
남의 카드 24개가 섞인다). develop을 재흡수하지 않았으므로 이 핀은 갱신 대상이 아니다.

### R12 — 빌드 · vet · lint · 패키지 전체

    $ go build ./internal/cli/
    (출력 없음, exit 0)

    $ go vet ./internal/cli/...
    (출력 없음, exit 0)

    $ golangci-lint run --timeout=5m ./internal/cli/...
    0 issues.
    (exit 0)

    $ go test ./internal/cli/ -count=1 -timeout 600s
    ok  	github.com/modu-ai/moai-adk/internal/cli	501.612s
    (exit 0)

전체 스위트(`go test ./...`)는 **돌리지 않았다** — 로컬 전체 스위트는 금지이며 판정은 CI 몫이다.

### R13 — 커밋 상태

    729060e63 test(SPEC-BINLAG-KEYGUARD-001): M1 guard allowlist keys are live check names (t479)

브랜치 `WT-binarylag-key-guard`, base `93fb36344` 위의 커밋 1개.
커밋 메시지에 카드 id **t479**가 들어 있다.
**push하지 않았다** — 공개는 리드의 일괄 행위다.

### R14 — 스테일 판독 · §E 충돌 사후 확인 (리드 지적에 대한 응답)

리드가 「spawn 이후 SPEC 4개가 다시 쓰였다」고 알려 왔다. 두 가지를 확인했다.

**① 낡은 판을 읽었는가 — 그렇다. 다만 열거 전에 스스로 발견해 재판독했다.**

세션 착수 시 읽은 `acceptance.md`는 470줄이었고 열거 시점에는 489줄이었다(R10에 기록).
좌표가 `:243/:290/:307` → `:246/:293/:310`으로 밀린 것도 R10의 열거가 현재 판을 대상으로
수행됐음을 보인다. 지금 다시 세 좌표를 출력해 대조했다:

    $ sed -n '246p' acceptance.md
    - 통과 조건: 종료코드 ≠ 0, 그리고 실패 메시지가 「허용목록이 비어 있어 이 가드가 아무것도 검사하지 못한다」는 취지를 말한다.
    $ sed -n '293p' acceptance.md
    - (1)의 RED는 종료코드 ≠ 0을 실패로 읽는 방향이므로 공허한 경우가 실패로 떨어져 안전하다.
    $ sed -n '310p' acceptance.md
    - 통과 조건: 두 실행 모두 종료코드 ≠ 0이고, 두 메시지가 **다른** 교정을 제시한다. 각 실패의

R10이 A로 분류한 세 줄과 **일치한다.** 즉 열거는 현재 판 기준이며, seed를 옮겨 적은 것이 아니다.

**낡은 판을 읽은 것이 구현에 영향을 주지 않은 이유**: 바뀌지 않은 것들(가드 이름, 두 매치 수
판정식, AC-BLKG-009 선행 의무, 범위 1파일, push 금지)이 구현을 규정하는 전부였고, 바뀐 것들
(T1/T2/T3/R3 + 좌표 강등)은 **판정과 보고 방식**에 걸리는데 그 셋 모두 이 회차가 충족한다:

- **T3**(≠ 0에 실패의 정체를 함께 요구) — 뮤턴트 3종 전부 파일:줄 + 메시지 본문으로 기록했다
  (m1 `:260` / m2 `:257` / m2′ `:253`, AC-BLKG-004는 `:232`).
- **T2**(초록 읽기 형제 자리에 두 매치 수) — §D.1의 「기존 가드도 함께 GREEN」을 두 매치 수로
  읽었다(R8: `[no tests to run]` 0 · 뒤 공백형 `DoctorCheckNameSetIsUnchanged` 1).
  AC-BLKG-006 원복 초록도 같은 두 수로 읽었다(R6 말미).
- **R3**(뮤턴트 3건) — m1 / m2 / m2′ 3건을 돌렸다.
- **좌표 seed 강등** — R10에서 재열거했다.

**② §E 충돌 — 없다. 양쪽 다 온전하다.**

    $ grep -n '^## §' progress.md
    13:## §E.1 Plan-phase Audit-Ready Signal
    294:## §E.2 Run-phase Evidence
    589:## §E.2.1 Run-phase 5-Section 보고
    646:## §E.3 Run-phase Audit-Ready Signal
    670:## §E.4 Sync-phase Audit-Ready Signal
    676:## §F Phase 4 Mode Selection

§E.1(spec-t479 소관)에 **0.5.0 회차의 내용이 들어 있다** — 즉 내가 이어 쓴 대상은 그쪽의
**최종판**이었다:

    $ sed -n '13,293p' progress.md | grep -n 'T1\|seed\|T2'
    181:**최종 감사 T1 — 이전 열거는 과소열거였다.** …
    194:**좌표는 seed로 강등했다** — 기대값은 `C = 0` 하나뿐이다. …
    198:**T2 형제 표면 스윕** — 초록을 읽는 자리를 문서 전체에서 훑어 네 곳에 두 매치 수를 붙였다:

§E.1은 Residual-risk 마지막 항목까지 온전히 끝나고(`progress.md:286-290`), 그 뒤에 §E.2가 온다.
내가 쓴 것은 §E.2 / §E.2.1 / §E.3 / §F뿐이며 §E.1은 한 글자도 건드리지 않았다.

**`internal/cli/binary_lag_test.go`도 무충돌** — 이 파일의 유일한 작성자는 이 세션이며,
커밋 `729060e63` 하나가 전부다(AC-BLKG-007에서 파일 수 1로 기계 판정).

**남는 위험**: 위 두 확인은 **현재 시점의 관측**이다. 리드가 「작성자는 이제 당신 하나」라고
알려 왔으므로 이후 재판독 의무는 종료하되, 병합 창에서 병합 트리 재측정이 필요하다는 점은
그대로다(§Residual-risk).

### AC 매트릭스

| AC | 판정 | 근거 |
|---|---|---|
| AC-BLKG-001 | **PASS** | R8 — 1단 0 / 2단 1. 두 등록 모양 모두에 대해 실재 단언 |
| AC-BLKG-002 | **PASS** | R4 — 무력 키가 이름으로 지목되고 「어떤 이름도 허용하지 못한다」고 말한다 |
| AC-BLKG-003 | **PASS** | R6 — 두 뮤턴트가 STRIPPED / ADDED로 **다른** 방향과 **다른** 교정을 제시 |
| AC-BLKG-004 | **PASS** | R7 — 빈 허용목록에서 exit 1 + 「아무것도 검사하지 못했다」. 기존 가드는 RED(측정 성립) |
| AC-BLKG-005 | **PASS** | R4+R5 — 원본 RED / 뮤턴트 GREEN(0·1) / 원복 GREEN(0·1) |
| AC-BLKG-006 | **PASS** | R6 — m2 `:257` STRIPPED · m2′ `:253` ADDED, 원복 GREEN(0·1) |
| AC-BLKG-007 | **PASS** | R11 — two-dot 파일 수 1, 그 1개가 대상 테스트 파일 |
| AC-BLKG-008 | **PASS** | R9 (1단 47 > 0, 2단 0) + R10 (모집단 5 · A 3 · B 2 · **C 0**) |
| AC-BLKG-009 | **PASS** | R2 — 존재하지 않는 이름에서 1단 1 / 2단 0. **가장 먼저** 돌렸다 |

판정 불가 항목 **없음**. 통과 9 / 실패 0.

---

## §E.2.1 Run-phase 5-Section 보고

### Claim

`internal/cli/binary_lag_test.go`에 가드 `TestBinaryLag_AllowlistKeysAreLiveNames`를 더해,
`namesAddedAfterBaseline`의 모든 키가 현재 `doctor.go`에서 실제로 추출되는 이름임을 기계가
판정하게 했다. 실패는 따옴표 실수의 **방향**을 지목하고 그에 맞는 교정을 제시한다.
허용목록이 비면 통과를 주장하지 않는다. AC 9건 전부 PASS이며 프로덕션 코드는 변경하지 않았다.

### Evidence

R2 … R13. 각 항목이 자기 명령과 verbatim 출력을 갖는다.

### Baseline-attribution

전부 **이 실행에서, 이 트리**(`.claude/worktrees/t479`)에 대해 냈다.
R2 · R11(편집 전) · R9(1단)은 HEAD **`93fb36344`**에서, R4 … R8 · R12는 워킹트리 상태에서,
R10 · R11(커밋 후) · R13은 HEAD **`729060e63`**에서 측정했다.
plan-phase가 기록한 값을 옮겨 적은 것은 없다 — RED-now(EV-BLKG-001)와 같은 모양의 관측도
R2에서 새 이름으로 다시 실행해 얻었다.

**건네받은 값 1건**: 카드 t477(`b2f98f6aa`)의 뮤턴트 2건은 이 세션이 재현하지 않았다.
이 카드의 뮤턴트 3건(m1 / m2 / m2′)은 그것과 다른 집합이며 전부 이 트리에서 직접 돌렸다.

### Gaps

명시적으로 **관측하지 않은** 것.

- **전체 스위트**(`go test ./...`): 돌리지 않았다. 로컬 전체 스위트는 금지이며 판정은 CI 몫이다.
  `internal/cli` 패키지 전체는 돌렸다(R12).
- **크로스플랫폼 빌드**(`GOOS=windows` 등): 돌리지 않았다. 변경이 테스트 파일 한 개이고
  플랫폼 의존 코드가 없으나, **안 쟀다는 사실 자체는 그대로 남는다.**
- **`internal/cli` 밖의 패키지**: 건드리지 않았으므로 돌리지 않았다.
- **t466 / t477 자체의 정확성**: 재검증하지 않았다. 범위 밖.
- **커버리지 수치**: 재지 않았다. 이 카드는 테스트만 더하며 프로덕션 코드 커버리지 목표를
  대상으로 삼지 않는다.
- **병합 트리에서의 재측정**: develop을 재흡수하지 않았으므로 병합 후 상태는 미관측이다.

### Residual-risk

- **양방향 진단의 도달 가능성**: 두 분기는 허용목록에 두 등록 모양이 공존하는 데 기대고 있다.
  나중에 등록 방식이 한쪽으로 통일되면 한 분기가 조용히 도달 불가가 된다 — 가드는 고장 나도
  침묵한다. 지금은 두 뮤턴트가 각 분기를 실제로 밟았으므로 **현재는** 둘 다 살아 있다.
- **`default` 분기의 뮤턴트는 1종뿐이다**: R4의 무력 키가 그 분기를 밟았지만, 「따옴표 변형도
  아니고 완전 오타도 아닌」 제3의 모양은 시험하지 않았다.
- **AC-BLKG-008(1)의 좌표 부패**: 이번 판에서 좌표가 seed로 강등돼 위험이 줄었으나,
  재열거를 실제로 수행하는지는 이 항목이 보증하지 못한다. 이번 회차는 수행했고
  **문서가 세션 도중 바뀌어 있었다**(R10) — 그 자체가 위험이 실재함의 증거다.
- **병합 창의 의미 충돌**: 다른 카드가 `doctor.go`에 체크를 더하면 허용목록과의 관계가 달라진다.
  파일이 겹치지 않아도 새 가드가 RED가 될 수 있다. 병합 트리에서 재측정이 필요하며,
  그때 AC-BLKG-007의 base 핀도 함께 갱신한다.
- **네 번째 공허 형태가 실제로 나왔다**(R6). `%q`/`%s` 혼용으로 두 피연산자가 같은 글자가 된
  것은 AC-BLKG-009가 검사하는 「이름 축」 밖의 형태이며, **판정식 검증이 아니라 뮤턴트 실행이**
  잡았다. 다섯 번째가 없다는 보장은 여전히 없다.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-06
run_commit_sha: 729060e63
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
ac_undecidable_count: 0
preserve_list_post_run_count: 0        # PRESERVE 목록 위반 0
l44_pre_commit_fetch: not-performed    # 레인 로컬 커밋만 수행, push 없음
l44_post_push_fetch: not-applicable    # push하지 않았다
new_warnings_or_lints_introduced: 0    # golangci-lint ./internal/cli/... -> 0 issues
cross_platform_build:
  darwin_amd64: not-measured           # 테스트 파일 1개 변경, 플랫폼 의존 코드 없음
  windows_amd64: not-measured
  note: "안 쟀다. 판정은 CI 매트릭스 몫."
total_run_phase_files: 1               # internal/cli/binary_lag_test.go
m1_to_mN_commit_strategy: single-commit # M1 하나로 종결 (729060e63)
```


---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-06
sync_commit_sha: 72e6e380f              # backfill 완료 (sync 커밋 자신은 자기 해시를 인용할 수 없다)
sync_status: complete
b12_self_test_a: pass                    # 사전 중복 grep: grep -c 'SPEC-BINLAG-KEYGUARD-001' CHANGELOG.md -> 0 (append 전)
b12_self_test_b: pass                    # AC 수 일치: acceptance.md 고유 AC-ID 9개 == CHANGELOG 인용 9
b12_self_test_c: pass                    # CHANGELOG가 주장한 파일 경로 5개 전부 ls 확인
changelog_entry_position: "[Unreleased] / ### Added 최상단 (CHANGELOG.md:12)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"    # updated: 2026-09-06 (이미 동일 날짜)
  plan_md: not-applicable                # status: 필드 없음
  acceptance_md: not-applicable          # status: 필드 없음
  progress_md: not-applicable            # frontmatter 없음 (본문 머리말만)
canary_compliance_check:
  applicable: false                      # 이 SPEC은 자기 sync가 시험할 forward-looking 정책을 정의하지 않는다
verification_rerun_at_sync:
  command_1: "go test ./internal/cli/ -run TestBinaryLag -count=1 -timeout 600s -v"
  verdict_1: "[no tests to run] 계수 0 / '--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames ' 계수 1 / SKIP 0 / FAIL 0"
  command_2: "go vet ./internal/cli/..."
  verdict_2: "exit 0, 출력 0바이트"
  command_3: "git diff --stat 93fb36344..HEAD -- internal/ pkg/ cmd/"
  verdict_3: "internal/cli/binary_lag_test.go | 70 ++ — 소스 변경 파일 1개"
docs_surface_touched: none               # 내부 테스트 가드. docs-site / README / locale 파일 0
template_tree_touched: none              # internal/template/templates/ 변경 0 — make build 불필요
pushed: false                            # 레인은 push하지 않는다. 공개는 리드의 일괄 행위
card_evidence: .moai/reports/t479/verdict.md
```

---

## §F Phase 4 Mode Selection

**입력 파라미터**

| 항목 | 값 |
|---|---|
| tier | M |
| 범위(파일 수) | 1 (`internal/cli/binary_lag_test.go`) |
| 도메인 수 | 1 (Go 테스트 코드) |
| 파일 언어 구성 | 100% Go |
| 병렬화 이득 | 낮음 — 코딩 작업이며 독립 분해 대상이 없다 |

**모드 평가**

| 모드 | 선택 | 사유 |
|---|---|---|
| `direct` | 아니오 | 오타·1줄 수정이 아니다. 뮤턴트 3종과 판정식 검증을 포함하는 실질 구현이다 |
| `serial` | **예** | 파일 하나에 대한 코딩 작업. 기본 fallback이며 이 카드에 정확히 맞는다 |
| `fanout` | 아니오 | 도메인 1개, 파일 1개. 펼칠 대상이 없다 |
| `sweep` | 아니오 | 기계적 대량 변환(~30 파일)이 아니다 |

**Decision: serial**

**근거.** 변경 대상이 단일 파일의 단일 테스트 함수이고, 뮤턴트 3종(m1 / m2 / m2′)은 같은 파일을
순차로 고쳤다 되돌리는 절차라 병렬로 나눌 수 없다 — 두 뮤턴트를 동시에 심으면 서로의 판정을
오염시킨다. 또한 이 프로젝트의 규칙은 코딩 작업에 대해 순차 sub-agent를 기본으로 두며
(Anthropic의 coding-task parallelism 단서), 이 카드는 그 기본에서 벗어날 이유가 없다.
경계 사례 없음 — 어느 임계값에도 근접하지 않는다.
