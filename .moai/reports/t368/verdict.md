# t368 판정서 — J5 배너 단언이 로고 복원 이전 계약을 들고 있던 건

- 카드: t368
- 워크트리: `.claude/worktrees/t368` · 브랜치 `WT-e2e-j5-banner-contract`
- 베이스: 로컬 develop 흡수 후 `083d7b790`
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t368`

---

## 1. 카드 전제 검증 — 전량 재현됨

카드 값은 `main 48239c7dc` 기준이었고, 아래는 **내 트리(`083d7b790`)에서 다시 잰 것**이다.

### 지금 배너 (수리 전, `NO_COLOR=1 moai --help`)

```
awk '/USAGE/{exit} {n++} END{print n+0}' j5-help.out   →  16
```

구성도 카드 서술과 같다 — 로고 6줄(1-6) + 빈 줄 2 + 설명 8줄 = 16. 로고 6줄 중
5줄이 `U+2588`(블록)을 담고, 6번째 줄은 박스드로잉 하단(`U+255A U+2550 U+255D`)이라 블록 문자가 없다.

### 두 단언은 실제로 실패한다

| 줄 | 단언 | 수리 전 판정 |
|---|---|---|
| 218 | `! grep -Eq` 블록/ASCII-art 패턴 → "no large ASCII-art logo" | 패턴이 1-5행에 매치되므로 부정 단언 **실패** |
| 220 | `BANNER_LINES -le 12` | 16 > 12 이므로 **실패** |

### 마지막 초록의 실체

카드가 인용한 `e2e/.runs/tux3-20260721/asserts.log:34-35` 를 직독했다.
이 경로는 gitignore 대상이라 워크트리에 없고 **primary 체크아웃에만** 있어 거기서 읽었다:

```
PASS  J5 no large ASCII-art logo
PASS  J5 compact banner (<=12 lines before USAGE, got 9)
```

배너 9줄 · 로고 없음 — 카드 서술과 정확히 일치한다. 즉 두 단언은 로고가 없던 시절의
계약이고, 로고 복원(`b1ea545e2`, PR #1145) 이후 상시 red 였다.

---

## 2. 갱신 기준 — REQ-TUXIU-055 / AC-TUXIU-024

`spec.md:104` 와 `acceptance.md:110,219` 가 계약을 명시한다. 로고는 이제 **의도된 것**이고,
계약은 세 갈래다:

1. 명시적 루트 help(`moai --help` / `-h` / `help`, len==1) → 로고 **있음**
2. 서브커맨드 help(`moai help init` / `moai init --help`) → 로고 **없음**
3. 무인자 `moai` → 로고 **정확히 1회** (이중 출력 가드)

SPEC status 는 `completed` 이고 복원 커밋은 `b1ea545e2`
("feat(SPEC-CLI-TUX-INIT-UPDATE-001): modernize moai init/update TUI + restore MoAI-ADK logo (#1145)").

### 분업 — e2e 가 맡을 절반

2·3번(음의 방향)은 `internal/cli/fang_roothelp_test.go` 의 `TestIsRootHelpArgs` 가
**6개 arg 형태 전수**로 이미 덮는다(빈 벡터 / `--help` / `-h` / `help` / `help init` / `init --help`).
다만 그것은 **술어**를 덮지 렌더 결과를 덮지 않는다 — 단위 테스트는 로고가 실제로
찍히는지 볼 수 없다. 그 절반이 J5 의 몫이므로, J5 에는 1번(렌더 확인)만 넣고
2·3번을 e2e 로 중복시키지 않았다.

---

## 3. 수리 — 가드를 갱신했지 끄지 않았다

`e2e/cli/tux3_journeys.sh` 두 단언.

**218 — 부호를 뒤집었다.** "로고가 없어야 한다" → "복원된 로고가 루트 help 에 있어야 한다".
단언을 지우지 않은 이유: 지우면 로고가 통째로 사라져도 J5 가 아무 말을 안 한다.

**220 — 예산을 살렸다.** 배너 전체 상한을 12→16(또는 그 이상)으로 올리는 것은
가드를 **갱신**하는 것이 아니라 **은퇴**시키는 것이다. 로고 6줄은 자체 SSOT 가드를 가진
고정 자산이므로, 원래의 12줄 압축 예산은 그것이 여전히 의미를 갖는 곳 —
**로고가 아닌 산문** — 에 그대로 남겼다.

```
BANNER_LINES=$(awk '/USAGE/{exit} {n++} END{print n+0}' ...)
LOGO_LINES=$(awk '/USAGE/{exit} /<block>|<boxdraw>/{n++} END{print n+0}' ...)
PROSE_LINES=$((BANNER_LINES - LOGO_LINES))
[ "$PROSE_LINES" -le 12 ]
```

현재 값: 16 - 6 = **10 ≤ 12**. 로고 복원 전 산문이 9줄이었으므로 예산은 사실상 그대로다.

계수식은 멀티바이트 문자클래스 대신 **교대**(`/A|B/`) 형태를 썼다. 둘 다 이 트리에서 6을
냈지만(실측), BSD awk 의 멀티바이트 문자클래스는 이식성이 불확실하다.

---

## 4. 검증

### 전체 journey — 실제 실행

```
MOAI_E2E_BIN=<내가 빌드한 바이너리> MOAI_E2E_RUN_DIR=<scratch> bash e2e/cli/tux3_journeys.sh
```

```
[J1] PASS  [J1b] PASS  [J2] PASS  [J3] PASS  [J4] PASS  [J5] PASS  [J6] PASS
```

J5 단언 **10/10 통과** (증거: `.moai/reports/t368/asserts-after.log`):

```
PASS  J5 restored logo present on root help (REQ-TUXIU-055)
PASS  J5 compact banner prose (<=12 non-logo lines before USAGE, got 10 of 16)
```

카드는 "나머지 8개 통과"라 했고, 이제 10개 전부 통과한다.

### 뮤테이션 — 새 단언이 공허하지 않음

입력을 변조해 각 단언이 실제로 FAIL 하는지 봤다.

| 뮤턴트 | 조작 | 관측 |
|---|---|---|
| M1 | 로고 줄 제거 (2026-07-21 형태로 되돌림) | 존재 단언 `grep -q` rc=**1** → FAIL |
| M2 | 로고 유지 + 산문 6줄 추가 | BANNER=22, LOGO=6 → PROSE=**16 > 12** → FAIL |

두 방향 모두 잡힌다. M1 은 "로고를 없애면 잡히는가", M2 는 "가드를 은퇴시키지 않았는가"를
각각 묻는다 — 한쪽만 보면 부호만 뒤집고 예산을 버린 것과 구별되지 않는다.

### 기타

- `bash -n e2e/cli/tux3_journeys.sh` → 무출력 (문법 OK)
- Go 코드 변경 없음 — 이 카드는 e2e 스크립트 1파일만 건드린다

---

## 5. Gaps / 잔여 위험

- **Gaps**: `shellcheck` 가 이 머신에 없어(`command not found`) 정적 검사는 `bash -n` 까지다.
  CI 판정은 읽지 않았다(레인은 CI 를 직접 요청하지 않는다 — 판독은 리드 몫).
  darwin 에서만 돌렸고 linux/windows 는 관측하지 않았다.
- **잔여 위험**: `LOGO_LINES` 계수식은 로고가 블록/박스드로잉 문자를 쓴다는 사실에 의존한다.
  로고 아트가 다른 문자 집합으로 바뀌면 계수가 0이 되고, 그때는 218(존재 단언)이 먼저
  FAIL 하므로 조용히 통과하지는 않는다 — 즉 실패 모드가 침묵이 아니라 red 다.
- **범위 밖**: 서브커맨드 help 의 로고 부재와 무인자 이중 출력 가드는 e2e 에 넣지 않았다
  (§2 분업). 그쪽이 깨지면 `TestIsRootHelpArgs` 가 잡는다.
