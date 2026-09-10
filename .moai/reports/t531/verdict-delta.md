# t531 델타 재감사 — F1/F2 폐쇄 판정

> 별도 파일로 쓴 이유: `verdict.md` 는 개정 커밋 `05b7914a9` 가 **근거로 인용한 기록**이다.
> 인용된 파일에 덧붙이면 인용 대상이 바뀐다. 원본은 바이트 그대로 두고 델타만 여기 남긴다.

- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t531` (`git rev-parse --show-toplevel` 로 첫 측정과 같은 배치에서 확인)
- 브랜치 / HEAD: `WT-claudelocal-push-model` / `71839912e`
- `CARD_BASE` = `bce6d7e083208097960c88deac11c1365ad900bc` (`git merge-base origin/develop HEAD`, 읽는 시점에 재유도)
- 범위: F1 · F2 폐쇄 여부, 개정 전이의 실재, 부수 파손. **AC 전수 스윕은 하지 않았다**(선행 감사가 재실행해 확정)
- 모든 부재·계수 측정은 `/usr/bin/grep`

---

## Claim

1. **F1 폐쇄 — 성립.** §2.3 의 전제가 §0.4 가 규정한 ` M CLAUDE.local.md` 한 건을 기대 baseline 으로 지목하고, primary 에서 0 이 되지 않음을 명시하며, 다른 파일이 함께 수정된 경우의 행동을 따로 준다. 문자 그대로 따르는 독자가 §0.4 가 금지한 되돌림으로 떠밀리지 않는다. 귀속 지시는 조언으로 약화되지 않았다 — 계수 레시피는 그대로 남아 있고 주석이 baseline 값만 정정한다.
2. **F2 폐쇄 — 성립.** 앞섬·뒤처짐 두 방향이 함께 서술되고, 흡수 전 최신화가 지시되며, 판정식과 갱신 경로는 복사 대신 `.claude/rules/local/gitflow-lane-protocol.md` §11 을 가리킨다. 포인터는 실재하고 주장한 내용을 실제로 담는다. 산문 불릿과 레시피 주석이 같은 §11 을 가리킨다.
3. **개정 전이 — 실재하고 승인 경로 A 와 일치.** `completed → in-progress`, `0.1.0 → 0.1.1`, 자기참조 `amendment_of`, 직전 completed SHA `b7344d957` 기록. 요구 **정의** 7건 전후 동일, `acceptance.md` diff 0, §G 간극 4건·`[추론]` 2건 무변.
4. **세 취약 프로브 — 전부 살아 있다.** AC-001 after 0 / base 대조군 2, AC-003 창 5줄·팔2 2·팔3 0 에 뮤턴트 M-003 이 요구대로 죽인다(팔2 2→0, 팔3 0→1), AC-005 정확히 2건이 모두 리드 일괄 블록 안.
5. **[새 결함] 개정이 형제 산출물 둘을 옛 상태에 남겼다.** `CHANGELOG.md:375` 는 이 SPEC 을 `in-progress → completed` 로 단언하고 흡수 정정 근거를 **앞섬 한 방향으로만** 서술한다. `progress.md` 는 sync 종결 상태와 수리 이전 수치를 그대로 들고 있다(AC-003 arm1 `2`, 실측 `3`; AC-007 대조군 `5`, 실측 `9`). 개정의 선언된 범위(「`CLAUDE.local.md` §2.3 · §4.1 **본문만**」)가 이 둘을 **명시적으로 배제**하므로, 개정문을 문자 그대로 읽는 레인은 이 상태를 그대로 두고 재종결한다 — F1 과 같은 모양의 문자-독해 함정이다.

## Evidence

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t531
$ git rev-parse HEAD
71839912e9ae0688db2eb1ed2fd145ee4b7d2c89
$ git branch --show-current
WT-claudelocal-push-model
$ git merge-base origin/develop HEAD
bce6d7e083208097960c88deac11c1365ad900bc
```

**F1 — 산출 문면 (`CLAUDE.local.md:227`, `:230`)**

```
**[HARD] update 실행 후 매번 검증한다.** 전제: 실행 **전** 추적 파일 수정이 **§0.4 가 규정한 ` M CLAUDE.local.md` 한 건뿐**이어야 diff 귀속이 가능하다. primary 체크아웃에서 그 표식은 영구적이므로 **0 이 되는 일은 없다** — 0 을 전제로 읽고 그 한 건을 없애려 들면 §0.4 가 막은 회귀로 떠밀린다. 다른 파일이 함께 수정돼 있으면 그때는 귀속이 불가능하니, update 전에 그쪽을 먼저 정리한다.
git status --porcelain | grep -v '^??' | wc -l        # 실제 변경 수 — primary 의 baseline 은 0 이 아니라 1(§0.4)
```

§0.4 (`:29`-`:39`) 와 함께 읽었을 때: §0.4 는 표식이 설계상 상시임을 선언하고 `git restore CLAUDE.local.md` 를 금지한다. §2.3 은 그 한 건을 기대값으로 못박고 **다른 파일**만 정리 대상으로 남긴다. 두 절이 같은 행동을 지시한다.

**F2 — 산출 문면 (`CLAUDE.local.md:389`, `:404`-`:410`)**

```
389: … 본인 워크트리에서 `git merge develop` 흡수(대상은 **로컬** `develop` — 원격이 아니다. 흡수 **전에** 그 로컬 develop 이 최신인지부터 본다 — 판정식과 갱신 경로는 `.claude/rules/local/gitflow-lane-protocol.md` §11) → …
404: # 흡수 전에 로컬 develop 을 먼저 최신화한다. 판정식(ref 비교)과 갱신 경로는
405: # `.claude/rules/local/gitflow-lane-protocol.md` §11 이 소유한다 — 여기 복사하지 않는다(두 벌이 되면 갈라진다).
407: # 어긋나는 방향은 둘이고, 둘 다 같은 결함을 낸다.
408: #   앞설 때: … 원격을 흡수하면 그 착지분이 빠진 베이스에서 재측정한다.
409: #   뒤처질 때: … 최신화 없이 로컬을 흡수하면 낡은 베이스에서 재측정한다.
410: # 거울상이므로 한쪽만 막으면 다른 쪽으로 새어 나간다.
```

포인터 실측 — `§11` 은 실재하고 주장한 두 가지를 모두 담는다:

```
$ /usr/bin/grep -n '^## 11' .claude/rules/local/gitflow-lane-protocol.md
118:## 11. develop 갱신 — 병합 후 로컬 develop 재생성 (SPEC-RC-TESTBED-001)
```
§11 본문: `[HARD] 판정 기준은 origin/develop 과의 ref 비교` + `git rev-list --count --left-right origin/develop...develop`("0 0" = 최신), `[HARD] 갱신 경로는 BranchGuard-안전 경로 하나뿐 — 통합 워크트리로의 런처 진입`.
그 파일은 이 카드가 건드리지 않았고 `origin/develop` 사본과 동일하다:

```
$ git diff --stat origin/develop -- .claude/rules/local/gitflow-lane-protocol.md
(출력 없음 = 동일)
$ git show origin/develop:.claude/rules/local/gitflow-lane-protocol.md | /usr/bin/grep -n '^## 11'
118:## 11. develop 갱신 — 병합 후 로컬 develop 재생성 (SPEC-RC-TESTBED-001)
```

`뒤처` 1건 / `앞설` 1건 — 개정 근거가 지목한 「뒤처 0건」이 해소됐다.

**세 취약 프로브 (전부 이 트리에서 직접 재실행)**

```
# AC-CLPM-001
sed -n '/^### §4.1/,/^## 5\./p' CLAUDE.local.md > /tmp/s41-after.md   # 102줄
/usr/bin/grep -c 'merge origin/develop' /tmp/s41-after.md            → 0
/usr/bin/grep -c 'merge origin/develop' CLAUDE.local.md              → 0      (파일 전역)
git show $CARD_BASE:CLAUDE.local.md | sed -n '/^### §4.1/,/^## 5\./p' | /usr/bin/grep -c 'merge origin/develop'  → 2  (대조군 살아 있음)
git show $CARD_BASE:CLAUDE.local.md | /usr/bin/grep -c 'merge origin/develop'                                    → 2

# AC-CLPM-003
/usr/bin/grep -n 'git restore CLAUDE.local.md' CLAUDE.local.md       → 35 (정확히 1건)
wc -l < /tmp/ac003-window.txt                                        → 5
팔1 /usr/bin/grep -c 'M CLAUDE.local.md'                             → 3
팔2 '하지 마라|하지 않는다|금지|회귀|되돌리면'                        → 2
팔3 '실행하라|실행한다|정리하려면|하면 된다|권장'                      → 0
창 내용: 33 `> **[HARD] 이 표식은 정리 대상이 아니다.**` / 35 `> **\`git restore CLAUDE.local.md\` 를 실행하지 마라** — … 회귀다.` / 37 `> 되돌리면 …`

# 뮤턴트 M-003 — /tmp 사본에서만 (추적 파일 무변)
orig sha before/after: a0919df2c7e7e82a541838d3d607f55cf6c1d976213785260b6b2e92aeb5c5fa  (동일, ORIGINAL UNCHANGED)
뮤턴트 팔2 → 0   (기대 0)
뮤턴트 팔3 → 1   (기대 >=1)
⇒ 뮤턴트에서 AC-CLPM-003 은 요구대로 FAIL 한다. 프로브는 죽지 않았다. 뮤턴트 사본 폐기 완료.

# AC-CLPM-005
/usr/bin/grep -n 'push origin develop' CLAUDE.local.md → 2건
  390 — `[HARD] WT 브랜치 push·CI 직접 요청 금지` 불릿, 「리드가 … 일괄로 실행하는」 문맥
  426 — `**[HARD] 리드 develop 일괄 push (2026-09-02)**` 블록의 코드펜스(`# 리드 — 통합 워크트리…, 창 밖에서`)
  ⇒ 둘 다 리드 일괄 절차 안. 레인-push 지시 0.
```

**개정 전이**

```
$ git show 05b7914a9 -- …/spec.md   →  -version "0.1.0" / +version "0.1.1", -status: completed / +status: in-progress,
                                        +amendment_of: SPEC-CLAUDELOCAL-PUSH-MODEL-001, + ## Amendments (직전 SHA b7344d957 · 근거 · 범위)
$ git cat-file -t b7344d957 → commit ; 제목 `docs(SPEC-CLAUDELOCAL-PUSH-MODEL-001): sync-phase artifacts (t531)`
  그 커밋의 spec.md 훙크: `-status: in-progress` / `+status: completed`   ⇒ 인용된 직전 completed SHA 는 실재하고 맞다
$ 요구 정의(줄 앵커 `^\*\*REQ-CLPM-[0-9]{3} `): HEAD 7, d17dcf2be 7      ⇒ 무변
$ REQ-CLPM 언급 전수 8줄 = 정의 7 + §Amendments 의 인용 1 (REQ-CLPM-004)   ⇒ 앵커가 놓친 정의 없음
$ git diff d17dcf2be HEAD -- …/acceptance.md   → 출력 없음 (AC 무변)
$ spec.md 훙크 헤더: @@ -1,10 +1,11 @@ / @@ -21,6 +22,36 @@   ⇒ 변경은 frontmatter + HISTORY/Amendments 두 곳뿐
$ §G 간극 4건 그대로, `[추론]` 2건 그대로 (diff 에 해당 줄 0)
```

**새 결함의 실측**

```
$ git diff --name-only d17dcf2be..HEAD
.moai/specs/SPEC-CLAUDELOCAL-PUSH-MODEL-001/spec.md
CLAUDE.local.md                       ⇒ CHANGELOG.md · progress.md 는 델타에서 손대지 않았다

$ CHANGELOG.md:375 (발췌)
  "… the local `develop` can be ahead of `origin/develop` … the reason is now carried inline beside the command."
  "This sync commit carries the 3-phase close (`spec.md` frontmatter `in-progress → completed` …)"
  ⇒ 전자는 F2 가 지적한 한-방향 서술 그대로, 후자는 HEAD 의 `status: in-progress` 와 정면 충돌

$ progress.md
  222:  AC-CLPM-003: {… arm1: 2 …}      실측 3   (수리가 §2.3 에 ` M CLAUDE.local.md` 한 줄을 더했다)
  227:  AC-CLPM-007: {… control: 5}      실측 9
  306:  spec.md: "in-progress → completed — status 와 updated 두 키만"   ⇒ 지금은 역방향 전이가 실재
$ 대조: git show d17dcf2be:CLAUDE.local.md | /usr/bin/grep -c 'M CLAUDE.local.md' → 2   (arm1 2→3 의 귀속)
$ git diff --name-only $CARD_BASE..HEAD | wc -l → 9 ;  … -- '*.go' | wc -l → 0
```

## Baseline-attribution

- 모든 「after」 수치: 이 실행, 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t531`), HEAD `71839912e` 의 워킹 사본에서 위 명령으로 측정.
- 모든 「before/대조군」 수치: 같은 실행에서 `git show <ref>:<path>` 로 `bce6d7e08`(CARD_BASE) 및 `d17dcf2be`(직전 판정 커밋) 블롭을 직독. 선행 감사 보고서의 수치를 옮겨 적지 않았다.
- 뮤턴트: `/tmp/ac003-mutant.md` 사본만 변이. 원본 `CLAUDE.local.md` 의 sha256 을 변이 전후로 재어 동일(`a0919df2…`)함을 확인했고, `git status --porcelain` 은 비어 있었다.
- 문자 수: `wc -m` 로 HEAD 42,318 / CARD_BASE 40,085 — 범위 밖 지시에 따라 판정 근거로 쓰지 않고 문맥으로만 기록한다.

## Gaps (관측하지 않은 것)

- **AC 전수 스윕**: AC-CLPM-001·003·005 와 002/004/006/007 의 계수만 재고, AC-CLPM-008(백업 sha256·660줄)은 **재지 않았다** — 백업 파일 접근이 이 감사의 범위 밖으로 지정됐다. AC-005 의 백업 대조군(기대 3) 역시 같은 이유로 재지 않았고, 선행 감사의 관측에 의존한다.
- **`origin/develop` CI**: 범위 밖 지시에 따라 보지 않았다.
- **§11 이 지정한 갱신 경로의 실행 가능성**: `EnterWorktree` → `git merge origin/develop` 경로를 실제로 밟아 보지 않았다. 문면 대조만 했다.
- **Go 테스트**: 이 카드에 Go 변경이 0이므로 돌리지 않았다(`.go` diff 0 은 측정했다).
- **다른 워크트리 40+**: `CLAUDE.local.md` 사본의 전수 대조는 여전히 §G-1 로 열려 있고 여기서도 재지 않았다.
- **primary 체크아웃**: 지시대로 건드리지 않았다. 「primary 에서 baseline 이 1」은 §2.3 의 **규범 진술**이며, 이 감사에서 primary 의 실제 `git status` 를 재어 확인하지 않았다.

## Residual-risk

- **§11 의 「갱신 창과 병합 창이 겹치지 않게 한다」와 새 §4.1 배치의 관계.** 새 지시는 갱신을 `moai integration acquire` **이후**, 즉 창 안에 넣는다. 같은 락 보유자 한 명만 도는 한 두 레인이 겹치지 않으므로 §11 의 의도는 지켜지지만, §11 문면을 「갱신은 별도 창에서」로 읽는 독자와는 갈라진다. 이 감사는 전자로 읽었다 — 확신도 중간.
- **산문 불릿(`:389`)은 「최신인지부터 **본다**」, 레시피(`:404`)는 「먼저 **최신화한다**」.** 하나는 판정, 하나는 실행이다. 둘 다 §11 을 가리키므로 실질 결과는 같지만, 커밋 메시지의 「prose bullet carries the same clause as the recipe so the two cannot drift apart」는 문자 그대로는 성립하지 않는다.
- **`git -C <카드워크트리> merge develop` 의 가드 위험은 이 카드가 만든 것이 아니다.** 절차 블록은 `moai cc -w develop` 로 통합 워크트리에 들어간 뒤 카드 워크트리를 `git -C` 로 조작하는데, 형제 독트린 §2 는 그 반대 방향의 교차-트리 `git -C` 를 가드가 거부한다고 못박는다. 이 형태는 base 에도 있었으므로 델타의 결함으로 세지 않는다 — 별도 카드 소관으로 남긴다.
- 문자 수 42,318 / 상한 40,000(초과분은 base 40,085 에서 이어진 것, t568 소관).

---

## 발견 (F 번호는 이 델타 감사의 자체 번호)

- **F1 (선행 감사) — 폐쇄.** 근거: 위 F1 문면 + §0.4 대조 독해.
- **F2 (선행 감사) — 폐쇄.** 근거: 양방향 서술 + §11 포인터 실재·내용 일치 + `뒤처` 1건.
- **F3 [Medium] [blocking] `CHANGELOG.md:375` · `progress.md`** — 개정이 SPEC 을 `in-progress` 로 되돌렸는데 두 형제 산출물은 종결 상태와 수리 이전 수치를 그대로 단언한다(CHANGELOG: 「`in-progress → completed`」 + 흡수 근거 앞섬 한 방향; progress.md: AC-003 `arm1: 2`(실측 3), AC-007 `control: 5`(실측 9), 전이 기록 「in-progress → completed」). 개정의 선언 범위가 「`CLAUDE.local.md` §2.3 · §4.1 **본문만**」이라 이 둘을 배제하므로, 개정문을 문자 그대로 읽는 레인은 이 상태로 재종결한다 — F1 과 동형의 문자-독해 함정. **요구 수리**: 지금 CHANGELOG/progress 를 고치라는 뜻이 아니라, 개정의 범위 절에 **재종결 시 갱신할 산출물**(CHANGELOG 항목의 상태 문장 + 흡수 근거 서술, progress.md §E 재측정)을 한 줄로 명시할 것. 확신도 높음(전부 실측).
- **F4 [Low] [optional] `CLAUDE.local.md:227`** — §0.4 는 표식의 영구성을 「primary 가 `main` 에 체크아웃돼 있는 동안」으로 **조건 지어** 선언하는데, §2.3 의 새 문장은 그 조건절을 떨어뜨리고 「primary 체크아웃에서 그 표식은 영구적」이라고 무조건으로 쓴다. primary 가 develop 계열에 체크아웃되면 baseline 은 0 이고 「0 이 되는 일은 없다」가 거짓이 된다. **요구 수리**: 조건절 복원(「primary 가 `main` 에 있는 동안」). 확신도 높음(문면 대조).
- **F5 [Low] [optional] `CLAUDE.local.md:227` 끝문장** — 「다른 파일이 함께 수정돼 있으면 … update 전에 그쪽을 먼저 정리한다」가 **안전한 경로를 지목하지 않는다**. 공유 primary 체크아웃에서 「정리」의 가장 짧은 문자적 독해는 `git restore`/`git checkout --` 이고, 그것은 다른 세션의 미커밋 작업을 파괴하며 상위 독트린이 primary 에서 금지한 행위다. **요구 수리**: 「커밋하거나 워크트리로 격리한다 — primary 에서 `git restore` 로 지우지 않는다」 정도의 한정. 확신도 중간(문자적 독해 기반, 실제 오작동은 관측하지 않았다).
- **F6 [Low] [optional] `CLAUDE.local.md:389` vs `:404`** — 판정(「본다」)과 실행(「최신화한다」)의 어긋남. 커밋 메시지의 비드리프트 주장과 불일치. 실질 위험 없음.
- **F7 [Low] [optional] `CLAUDE.local.md:381` vs `:389`/`:405`** — 같은 절이 같은 문서의 같은 절을 **이름**(「**develop 갱신** 절」)과 **번호**(§11)라는 두 형태로 가리킨다. 번호 포인터는 절 번호가 바뀌면 조용히 어긋나는 취약한 쪽이고, 파일에는 이미 이름 포인터가 있었다. **요구 수리**: 새 포인터도 이름 형태로 통일하거나, 번호에 절 이름을 병기.

## 델타 판정

**F1 · F2 — CLOSED.** 두 결함 모두 산출 문면에서 해소됐고, 폐쇄를 지탱하는 세 프로브(AC-001/003/005)는 대조군과 뮤턴트로 살아 있음이 이 트리에서 재확인됐다. 개정 전이는 실재하며 승인 경로 A 와 일치하고, 요구·인수 기준·§G 간극·`[추론]` 표지는 하나도 움직이지 않았다.

**전체 판정: FAIL (blocking F3 1건).** 수리 자체는 깨진 것이 없다 — 파손은 개정 커밋이 형제 산출물 둘을 옛 상태에 남기고, 그 둘을 자기 범위에서 명시적으로 배제한 데서 나온다. F3 의 최소 수리는 한 줄(개정 범위 절에 재종결 산출물 명시)이며, 그것이 들어가면 이 카드는 통과한다. F4~F7 은 선택 사항으로 보고하며, 자동으로 수리 대상에 넣지 않는다.
