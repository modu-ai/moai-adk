# SPEC-CLAUDELOCAL-PUSH-MODEL-001 — 수용 기준

> 검증 층. 요구(GEARS)는 `spec.md §B` 에 있고 여기에 다시 적지 않는다.
> 각 AC 는 Given-When-Then 이고, 이진 판정 가능해야 한다.

## 공통 전제 (모든 AC 앞에 선다)

```bash
CARD_BASE=$(git merge-base origin/develop HEAD)   # 범위의 왼쪽 끝 — 읽는 시점에 재유도
```

리터럴 `bce6d7e08` 은 날짜 붙은 앵커이지 범위의 왼쪽 끝이 아니다.
부재 주장은 `/usr/bin/grep` 으로 잰다(셸 `grep` 은 조용히 건너뛰는 ugrep 래퍼).
어떤 대조군이 0 을 내면 **「변경 없음」이 아니라 「측정 불가」**로 보고한다.

`BACKUP` 은 PRIMARY 체크아웃에 있다 — 이 워크트리가 아니다:

```bash
BACKUP=/Users/goos/MoAI/moai-adk-go/.moai/reports/t531/CLAUDE.local.md.primary-uncommitted-backup-20260908
```

---

## AC-CLPM-001 — 흡수 대상이 로컬 develop 으로 정정된다

**Given** 이 워크트리의 `CLAUDE.local.md` §4.1 레인 창 절차가 base 시점에 `git merge origin/develop`
흡수를 지시하고 있고,
**When** run-phase 수리가 끝난 뒤 §4.1 절을 절 구분자로 잘라 흡수 지시를 재고,
**Then** 흡수 대상이 **로컬 `develop`** 임이 문면에 나타나고, `git merge origin/develop` 형태의
흡수 지시는 남아 있지 않다.

```bash
sed -n '/^### §4.1/,/^## 5\./p' CLAUDE.local.md > /tmp/s41-after.md
/usr/bin/grep -c 'merge origin/develop' /tmp/s41-after.md      # 기대 0
git show "$CARD_BASE":CLAUDE.local.md | sed -n '/^### §4.1/,/^## 5\./p' \
  | /usr/bin/grep -c 'merge origin/develop'                     # 대조군: >=1 이어야 판별력 성립
```

**판정**: after 0 **그리고** base ≥1. base 가 0 이면 프로브가 죽은 것이므로 「측정 불가」로 보고한다.

---

## AC-CLPM-002 — 정본 선택 규칙이 문면에 있다

**Given** 사본이 갈렸을 때 어느 쪽을 믿을지가 다음 push-model 변경에서 다시 문제가 되고,
**When** 수리된 `CLAUDE.local.md` 에서 선택 규칙 문장을 찾으면,
**Then** **「레인이 분기하는 트리가 지배한다 = `develop`」** 이 문장으로 있고, 날짜·최신성으로
판별하라는 서술은 없다.

```bash
/usr/bin/grep -n '분기하는 트리' CLAUDE.local.md    # 기대 >=1 행
```

**판정**: 적중 ≥1, 그리고 그 문장이 `develop` 을 지배 사본으로 지목한다(사람이 읽어 확인).

---

## AC-CLPM-003 — ` M CLAUDE.local.md` 가 의도된 상태로 기록된다

**Given** primary 가 `main` 에 체크아웃돼 있고 main 커밋본이 폐기 모델이라 워킹본이 develop 판인
한 영원히 modified 로 읽히며,
**When** 수리된 문서에서 그 표식을 다루는 절을 찾으면,
**Then** ` M CLAUDE.local.md` 가 **의도된 상태**임이 적혀 있고, 되돌림 명령이 **금지 문맥 안에서만**
등장한다 — 문자열 존재가 아니라 그 문장이 금지인지 지시인지를 판정한다.

> **왜 문자열 존재로는 부족한가**: 「절대 `git restore CLAUDE.local.md` 하지 마라」와
> 「정리하려면 `git restore CLAUDE.local.md` 를 실행하라」는 문자열 프로브에서 **똑같이 통과한다**.
> 후자는 이 카드가 막으려는 바로 그 문장이다. 「같은 절 안에 있다」는 판정문도 이 둘을 못 가른다.
> 따라서 AC 를 만족시키는 가장 싼 경로가 거짓 초록이 되지 않도록, 아래 팔 2·3 이 창 안의
> **어휘**를 본다.

```bash
# 팔 1 — 표식 절 존재
/usr/bin/grep -n 'M CLAUDE.local.md' CLAUDE.local.md                     # 기대 >=1

# 되돌림 명령 적중 줄 ± 2줄 = 5줄 창을 잘라 낸다 (창 크기 5, 고정)
/usr/bin/grep -n -B2 -A2 'git restore CLAUDE.local.md' CLAUDE.local.md > /tmp/ac003-window.txt
wc -l < /tmp/ac003-window.txt                                            # 대조군: >=1 이어야 창이 산다

# 팔 2 — 금지 어휘가 그 5줄 창 안에 있는가 (양성 팔)
/usr/bin/grep -cE '하지 마라|하지 않는다|금지|회귀|되돌리면' /tmp/ac003-window.txt   # 기대 >=1

# 팔 3 — 지시 어휘가 그 5줄 창 안에 있는가 (음성 팔 — 뒤집힘 탐지)
/usr/bin/grep -cE '실행하라|실행한다|정리하려면|하면 된다|권장' /tmp/ac003-window.txt  # 기대 0
```

**판정**: 팔 1 ≥1 **그리고** 팔 2 ≥1 **그리고** 팔 3 = 0.
창 대조군(`wc -l`)이 0 이면 `git restore` 적중 자체가 없는 것이므로 팔 2·3 은 **공허**하다 —
그때는 PASS 로 세지 않고 「측정 불가」로 보고한다(빈 창에서 팔 3 = 0 은 공짜로 성립한다).

### [HARD] 뮤턴트 M-003 — 이 AC 의 유일한 판별 증거

팔 3 은 부재 형태이므로, **그것이 실제로 뒤집힘을 잡는지는 뮤턴트로만 증명된다.**
run-phase 에서 아래를 **실제로 돌리고 출력을 인용한다.**

```bash
# 원본을 건드리지 않는다 — /tmp 사본에서만 변이시킨다 (공유 트리 뮤테이션 아님)
cp CLAUDE.local.md /tmp/ac003-mutant.md
# 금지 → 지시로 뒤집는다 (실제 문면의 금지 어휘에 맞춰 치환한다)
#   예: "절대 … 하지 마라 — 폐기 모델로 회귀한다"  →  "정리하려면 … 를 실행하라"
$EDITOR /tmp/ac003-mutant.md    # 또는 sed 로 해당 줄 치환

# 같은 3팔 프로브를 뮤턴트에 돌린다
/usr/bin/grep -n -B2 -A2 'git restore CLAUDE.local.md' /tmp/ac003-mutant.md > /tmp/ac003-mutant-window.txt
/usr/bin/grep -cE '하지 마라|하지 않는다|금지|회귀|되돌리면' /tmp/ac003-mutant-window.txt   # 뮤턴트 기대 0
/usr/bin/grep -cE '실행하라|실행한다|정리하려면|하면 된다|권장' /tmp/ac003-mutant-window.txt  # 뮤턴트 기대 >=1
```

**뮤턴트 판정**: 뮤턴트에서 **AC-CLPM-003 이 반드시 FAIL** 해야 한다 — 팔 2 가 0 으로 떨어지거나
팔 3 이 ≥1 로 올라가거나, 둘 다. 뮤턴트가 여전히 PASS 하면 **프로브가 죽은 것**이고,
AC-CLPM-003 은 PASS 로 세지 않는다.

뮤턴트 파일은 판정 후 폐기한다(`rm /tmp/ac003-mutant*.txt /tmp/ac003-mutant.md`).
원본 `CLAUDE.local.md` 는 이 절차에서 **변경되지 않는다**.

---

## AC-CLPM-004 — main 커밋본이 폐기 모델임이 기록된다

**Given** 다음 사람이 `main` 사본을 정본으로 인용하는 것이 2026-09-07 실패의 재현이고,
**When** 수리된 문서에서 main 커밋본을 다루는 문장을 찾으면,
**Then** `main` 의 커밋본이 **폐기된 모델**(develop 을 원격에 올리지 않는 제3의 모델)이라는 사실이
적혀 있다.

```bash
/usr/bin/grep -n '폐기' CLAUDE.local.md | /usr/bin/grep -i 'main'   # 기대 >=1
```

**판정**: 적중 ≥1, 그리고 그 문장이 main 사본을 인용 금지 대상으로 지목한다.

---

## AC-CLPM-005 — 레인-push 지시가 재도입되지 않는다 (판별 대조군 동반)

**Given** 변종 2 의 지문은 「레인이 창 안에서 `git push origin develop`」이고,
**When** 수리 후 문서에서 그 지시를 찾고 **같은 프로브를 백업(변종 2)에도 돌리면**,
**Then** 수리본에서는 레인-push 지시가 없고, 백업에서는 있다 — 프로브가 살아 있음이 성립한다.

```bash
/usr/bin/grep -c 'push origin develop' CLAUDE.local.md   # 기대: 리드 일괄 절차 안의 것만
/usr/bin/grep -c 'push origin develop' "$BACKUP"         # 대조군: 3 (변종 2 의 지문)
```

**판정**: 수리본의 모든 적중이 **리드 일괄 절차 블록 안**에 있고(각 적중의 문맥을 읽어 확인),
백업 대조군이 3 이다. 대조군이 3 이 아니면 백업이 바뀐 것이므로 「측정 불가」로 보고한다.

---

## AC-CLPM-006 — 미커밋 워킹 사본의 정본 참칭 금지가 일반 규칙으로 있다

**Given** 축 1 의 결함이 특정 파일 사고가 아니라 인용 규율의 구멍이고,
**When** 수리된 문서에서 인용 규율 문장을 찾으면,
**Then** **「이력에 커밋된 적 없는 워킹 사본은 정본으로 인용될 수 없다」** 취지의 일반 규칙이
있다.

```bash
/usr/bin/grep -n '미커밋' CLAUDE.local.md    # 기대 >=1
```

**판정**: 적중 ≥1, 그리고 그 문장이 특정 파일이 아니라 **인용 일반**을 구속한다.

---

## AC-CLPM-007 — Go 코드 0 변경

**Given** 이 SPEC 이 문서 수리이고,
**When** 카드 범위의 diff 를 재면,
**Then** `.go` 파일 변경이 0 이다.

```bash
CARD_BASE=$(git merge-base origin/develop HEAD)
git diff --name-only "$CARD_BASE"..HEAD -- '*.go' | wc -l      # 기대 0
git diff --name-only "$CARD_BASE"..HEAD | wc -l                # 대조군: >=1 이어야 범위가 산다
```

**판정**: `.go` 0 **그리고** 전체 변경 ≥1. 전체가 0 이면 범위가 비었으므로 「측정 불가」다.

---

## AC-CLPM-008 — 백업이 무결하게 보존된다 (분석하지 않는다)

**Given** 처분의 되돌림 가능성이 백업 한 파일에 걸려 있고,
**When** 백업의 서명을 재면,
**Then** sha256 이 `23f8427589b739c705c8b17def0409c187af92ce5f27ee8bb87037a8376c9a7a` 이고
줄 수가 660 이며, **내용은 분석되지 않는다**.

```bash
shasum -a 256 "$BACKUP"
wc -l < "$BACKUP"      # 기대 660
```

**판정**: sha256 일치 **그리고** 660줄. 이 AC 는 존재·무결성만 판정한다 —
diff, 요약, 부분 복원 제안은 §F 범위 밖이다.

---

## Definition of Done

- AC-CLPM-001 ~ 008 전부 PASS, 각 AC 의 대조군이 판별력을 보였다.
- `spec.md §F` 에 기계 가드가 범위 밖으로, 백업이 보존-무분석으로 기록돼 있다.
- `spec.md §G` 의 간극 4건이 닫히지 않은 채로 남아 있다.
- 커밋 메시지가 `t531` 을 운반한다. push 하지 않았고 run-phase 에 진입하지 않았다.
