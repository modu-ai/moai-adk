# t531 — run-phase 반출 증거

> 인용 대상은 이 파일이다. `.moai/state/verify/` 및 `/tmp` 산출물은 machine-local scratch 이고
> gitignored 이므로 어떤 clone·CI 러너·다른 머신에도 닿지 않는다. 여기 반출되지 않은 자료는
> 판정 근거로 인용하지 않으며, 손실은 § 반출하지 않은 것에 이름으로만 남긴다.

## 측정 좌표

| 항목 | 값 |
|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t531` |
| git-common-dir | `/Users/goos/MoAI/moai-adk-go/.git` |
| 브랜치 | `WT-claudelocal-push-model` |
| `CARD_BASE` | `bce6d7e083208097960c88deac11c1365ad900bc` (읽는 시점에 `git merge-base origin/develop HEAD` 로 재유도) |
| 측정 시점 HEAD | `a833b5658` (AC 배치) · `b9571a231` (lint 배치) |
| 측정자 | lane-6 (manager-develop 보고와 **독립**으로 재측정) |

부재·건수 주장은 전부 `/usr/bin/grep` 으로 쟀다. 이 셸의 `grep` 은 조용히 건너뛰는 ugrep 래퍼이므로
셸 `grep` 의 0 은 부재의 증거가 되지 못한다.

---

## 1. 뮤턴트 M-003 — AC-CLPM-003 의 유일한 판별 증거

AC-CLPM-003 의 팔 3 은 부재 형태(`실행하라|실행한다|정리하려면|하면 된다|권장` 이 창 안에 0)다.
부재 형태는 기능이 없을 때 저절로 만족하므로, **그 팔이 실제로 뒤집힘을 잡는지는 뮤턴트로만 증명된다.**

원본은 건드리지 않고 `/tmp` 사본에서만 금지 문면을 지시 문면으로 치환했다.

### 뮤턴트 창 (실물)

```
33-> **[HARD] 이 표식은 정리 대상이다.**
34->
35:> 정리하려면 `git restore CLAUDE.local.md` 를 실행하라.
36->
37-> 표식이 사라지고 워킹 사본이 main 판으로 맞춰진다.
```

### 3팔 비교

| 팔 | 프로브 | 원본 | 뮤턴트 | 기대 |
|---|---|---|---|---|
| 1 | `grep -c 'M CLAUDE.local.md'` | 2 | — | ≥1 |
| 창 대조군 | `wc -l` of `grep -n -B2 -A2 'git restore CLAUDE.local.md'` | 5 | 5 | ≥1 (0 이면 공허) |
| 2 (양성) | `grep -cE '하지 마라\|하지 않는다\|금지\|회귀\|되돌리면'` | **2** | **0** | 원본 ≥1 |
| 3 (음성) | `grep -cE '실행하라\|실행한다\|정리하려면\|하면 된다\|권장'` | **0** | **1** | 원본 0 |

**판정**: 뮤턴트에서 팔 2 가 2→0 으로 떨어지고 팔 3 이 0→1 로 올랐다 — **AC-CLPM-003 이 뮤턴트에서
FAIL 한다.** 프로브가 살아 있다.

### 공유 트리 안전 증거

뮤테이션은 트리를 고의로 임시 훼손하는 행위이므로, 되돌린 뒤 동일성을 보이는 것이 그 창의 유일한
안전 증거다.

```
$ shasum -a 256 CLAUDE.local.md      # 절차 전
de5f4d0fb642832478818d44c5b61674a63d7bc1da89311509cdb36ef4e12596
$ shasum -a 256 CLAUDE.local.md      # 절차 후
de5f4d0fb642832478818d44c5b61674a63d7bc1da89311509cdb36ef4e12596
$ ls /tmp/lane6-mut*                 # 뮤턴트 파일 폐기 확인
(eval):13: no matches found: /tmp/lane6-mut*
```

원본은 이 절차에서 변경되지 않았고, 뮤턴트 사본은 남지 않았다.

---

## 2. `CLAUDE.local.md` 문자 예산 — 귀속을 세워 둔다

sync-audit 이 이 초과를 보게 되므로, 초과가 **어디서 왔는지**를 미리 귀속한다.

```
$ git show "$CARD_BASE":CLAUDE.local.md | wc -m
40085
$ wc -m < CLAUDE.local.md
41769
```

| 항목 | 값 |
|---|---|
| 상한 (`coding-standards.md § File Size Limits`) | 40,000자 |
| base(`bce6d7e08`) 시점 | **40,085자** — 상한 대비 **+85 초과** |
| 수리 후 | **41,769자** |
| 이 카드의 기여 | **+1,684자** (§0 신설) |

**두 가지가 이 측정으로 선다:**

1. **base 초과는 이 카드 이전에 존재했다.** 카드가 만든 위반이 아니다. 독립 관측이 하나 더 있다 —
   세션 시작 리마인더가 primary 체크아웃에서도 `40085` 를 발화한다.
2. **감축은 AC 를 지우지 않고는 불가능하다.** 추가된 1,684자는 AC-CLPM-002(`분기하는 트리`),
   AC-CLPM-003(` M ` 표식 절 + 금지 문맥 안의 되돌림 명령), AC-CLPM-004(main 폐기 기록),
   AC-CLPM-006(미커밋 인용 금지)이 요구하는 문면이다. 줄이려면 요구를 없애야 한다.

리드 판정(2026-09-08): 이 카드의 블로커가 아니다. 파일 분할은 후속 카드 후보이며, 그 방향은
path-scoped 룰 또는 `.moai/docs/` 이관 — 다만 **§0 은 항상 로드돼야 하므로 이관 대상이 아니다.**

---

## 3. 소유권 전이 lint — 추측을 실측으로 바꾼 기록

manager-develop 이 열린 위험으로 제기했다: `draft → in-progress` 전이가 매트릭스가 규정한
`M1` 커밋이 아니라 세 번째 커밋(`b9571a231`, `chore(...)`)에 실렸으므로 `OwnershipTransitionInvalid`
소견이 뜰 수 있다. 추측으로 두지 않고 쟀다.

```
$ moai spec lint .moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/spec.md
✓ No findings — all SPEC documents are valid
EXIT=0

$ moai spec lint … | /usr/bin/grep -c 'OwnershipTransition'
0
```

### [정정 2026-09-08] 위 0 을 「위험 부재」로 읽으면 안 된다

이 절은 처음에 「룰 문자열이 바이너리에 6행 적중하므로 0 은 죽은 프로브가 아니다 → 제기된 위험은
발생하지 않았다」로 적혀 있었다. **그 추론은 성립하지 않는다.** sync-phase 에서 manager-docs 가
「문자열 존재는 *실행되어 통과함* 과 *이 입력에서 돌지 않음* 을 가르지 못한다」를 짚었고, 갈라 보니
후자에 가까웠다. 아래가 재측정이다.

**먼저 어느 빌드가 판정했는지부터** (tool-provenance — 좌표는 둘이다):

```
$ moai version
v3.2.0-rc.1   list-744-g91d25bc61   built 2026-09-07T20:47:45Z
$ git -C <primary> rev-parse --short HEAD
7ad9f8534
$ git rev-parse 7ad9f8534:internal/spec/lint_ownership.go
a53e9e24106a12315c7d536ab3005b4157ead2c7
$ git rev-parse 91d25bc61:internal/spec/lint_ownership.go
ef598d5c7c54bfbc88b60ee0d4fd09ab41821a3e
```

판정한 바이너리는 `91d25bc61`(develop)에서 빌드됐는데 primary 체크아웃은 main `7ad9f8534` 다.
두 `lint_ownership.go` 블롭은 **다르다**. 첫 판독은 판정하지 않은 트리에서 읽은 것이었으므로,
아래 인용은 **판정한 커밋의 블롭**에서 다시 뜬 것이다(실제 차이는 1줄이었지만, 그건 확인한 뒤에야
알 수 있는 사실이다).

**룰은 등록돼 있고 실행된다** — `91d25bc61:internal/spec/lint.go:139` 의 `defaultRules()` 에
`&OwnershipTransitionRule{}` 가 있다.

**그런데 트레일러가 없으면 조용히 건너뛴다** — `91d25bc61:internal/spec/lint_ownership.go:414`:

```go
if rec.AuthoredByAgent == "" {
    return nil
}
```

이 카드의 커밋에는 그 트레일러가 없다:

```
$ git log --format='%h |%(trailers:key=Authored-By-Agent,valueonly)|' -5
132e752bf ||
b9571a231 ||
9cf0d625f ||
4af14489a ||
a833b5658 ||
```

**정정된 판정**: 룰은 돌았고, `Authored-By-Agent:` 트레일러가 없어 **말없이 건너뛰었다**. 따라서
`0 findings` 는 「전이가 소유권 검사를 통과했다」가 아니라 **「이 검사가 이 커밋들에 적용되지
않는다」**를 뜻한다. manager-develop 이 제기한 위험은 이 도구로 **확인되지도 반증되지도 않았다 —
미측정이다.** git 이력 판독 실패(`OwnershipTransitionUnreachable`)도 아니다: 그 경로였다면 Info
소견이 났을 텐데, `--json` 출력이 `[]` 였다.

**남는 채무 (둘)**:
1. 전이가 `M1` 커밋이 아니라 `chore(...)` 커밋(`b9571a231`)에 실렸다. 이력을 다시 쓰지 않는다.
2. **이 저장소의 커밋이 `Authored-By-Agent:` 트레일러를 달지 않으므로 `OwnershipTransitionRule` 은
   구조적으로 무음이다.** 이 카드의 범위 밖이며, 별개 관측으로 남긴다 — 이 룰의 초록을 소유권
   준수의 근거로 인용하는 모든 자리가 같은 공백 위에 서 있다.

---

## 4. AC 매트릭스 — 레인 독립 재측정

manager-develop 의 보고와 **독립으로** 같은 배치를 돌렸고 전 항목이 일치했다.

| AC | 측정 | 대조군 | 판정 |
|---|---|---|---|
| 001 | §4.1 슬라이스 `merge origin/develop` after **0** (파일 전체도 0) | base 슬라이스 **2** | PASS |
| 002 | `분기하는 트리` 적중 **2** (13행 제목, 15행 본문) | 존재형 — 대조군 없음 | PASS |
| 003 | 팔1 **2** · 창 **5행** · 팔2 **2** · 팔3 **0** | 뮤턴트 M-003 (§1) | PASS |
| 004 | `폐기`+`main` **한 줄 동시** 적중 **3** (25·27·31행) | 존재형 — 대조군 없음 | PASS |
| 005 | 수리본 `push origin develop` **2** (390행·421행, 둘 다 리드 일괄 절차 안) | 백업(변종 2) **3** | PASS |
| 006 | `미커밋` 적중 **2** (19행 제목, 21행 본문) | 존재형 — 대조군 없음 | PASS |
| 007 | `.go` 변경 **0** | 범위 전체 변경 **5** | PASS |
| 008 | sha256 일치 · **660줄** | 백업 `push origin develop` **3** | PASS |

AC-CLPM-008 백업 서명:
`23f8427589b739c705c8b17def0409c187af92ce5f27ee8bb87037a8376c9a7a`
(경로는 **primary 체크아웃**의 `.moai/reports/t531/CLAUDE.local.md.primary-uncommitted-backup-20260908`
— 이 워크트리가 아니다. 내용은 열지 않았다: `spec.md §F` 범위 밖.)

---

## 5. 반출하지 않은 것 — 이름만 부르고 근거로 쓰지 않는다

- **`/tmp/lane6-*` 중간 산출물** — 뮤턴트 사본, 창 절편, §4.1 슬라이스. 절차 종료 시 폐기했다.
  위 §1 의 표와 인용문이 그 자리를 대신하며, 원본 파일은 그 자체로 재현 가능하다.
- **manager-develop 의 전체 세션 트랜스크립트** — 반출하지 않았다. 그 보고의 수치는 레인이
  독립으로 재측정해 §4 에 세웠으므로, 트랜스크립트를 근거로 인용할 필요가 없다.
- **`moai spec audit` 출력** — 돌리지 않았다. lint 만 돌렸다.

## 6. 열어 둔 간극 (`spec.md §G`, 닫지 않는 것이 계획이다)

1. 워크트리 40+ 전수 미측정 — 「세 변종」은 읽은 세 좌표에 대한 진술이지 열거가 아니다.
2. 미커밋 편집본의 저자·시점 불명 — 커밋되지 않았으므로 이력에 없다.
3. 2026-09-07 사건 미재현 — 카드 본문의 서술이며 여기서 측정한 것이 아니다.
4. 「변종 2 가 당시 인용됐다」는 **[추론]** — 텍스트 일치에서 나왔고 그 배차를 관측한 것이 아니다.

## 7. 이 라운드에서 잡은 결함 (AC 밖)

`status` 가 두 run 커밋 착지 후에도 `draft` 였다. **AC-CLPM-001~008 어느 것도 `status` 필드를 보지
않으므로**, 8/8 초록이 이 결함을 가렸다. sync 직전에 프론트매터를 따로 연 것이 유일한 검출 경로였다.

원인은 두 층이고, 한 층으로 정리하지 않는다:

- **레인(프롬프트 층, 지배적 원인)** — 위임문이 `spec.md` 를 명시적으로 금지했고
  (「You own progress.md §E.2 and §E.3 only」), 매트릭스가 허용하는 프론트매터
  `status:`/`updated:` 예외를 적지 않았다.
- **manager-develop(에이전트 층)** — 상시 계약이 이 전이를 자기 의무로 규정하므로 충돌을 볼 위치에
  있었으나, 블로커를 올리지 않고 조용히 프롬프트 쪽으로 해소했다.

수리: `b9571a231`. 이력을 다시 쓰지 않았다.
