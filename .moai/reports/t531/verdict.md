# t531 / SPEC-CLAUDELOCAL-PUSH-MODEL-001 — sync-audit 판정

- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t531`
- 브랜치 / HEAD: `WT-claudelocal-push-model` / `25474b8a5`
- `CARD_BASE`: 읽는 시점에 재유도 — `git merge-base origin/develop HEAD` = `bce6d7e083208097960c88deac11c1365ad900bc`
- 판정 도구: `moai` v3.2.0-rc.1 `list-744-g91d25bc61` (built 2026-09-07T20:47:45Z)
- 모든 부재 주장은 `/usr/bin/grep` 으로 쟀다. 워킹트리는 감사 전후 모두 `git status --porcelain` 무출력.

## 총평 — FAIL

AC-CLPM-001~008 을 레인 매트릭스를 소비하지 않고 전부 재실행했다. **8/8 PASS 이고 대조군은 모두 살아 있다.**
뮤턴트 M-003 도 독립적으로 두 벌 만들어 돌렸고 둘 다 AC-CLPM-003 을 FAIL 시켰다 — 프로브는 죽지 않았다.

그럼에도 FAIL 이다. 이 카드는 **어떤 AC 도 보지 않는 자리에서 두 개의 결함**을 산출물에 남겼고,
둘 다 다음 레인이 **실행하는 문장**에 있다. 카드 자신이 이미 같은 모양의 결함을 하나 냈다
(`status` 가 `draft` 인 채 8/8 초록을 통과했다 — 어떤 AC 도 그 필드를 보지 않기 때문이다).
그것이 우연이 아니라 이 AC 집합의 구조적 사각이라는 것이 이번 감사의 결론이다.

`CLAUDE.local.md` 는 매 세션 전량 로드되는 문서다. 여기 실린 틀린 지시는 조용히 전파된다.

## 차원 점수

| 차원 | 점수 | 판정 | 증거 |
|---|---|---|---|
| Functionality (40%) | 72/100 | **FAIL** (must-pass) | AC 8/8 재실행 PASS(대조군 001=2, 005=3, 007=5, 003 창=5행) + 뮤턴트 2종 FAIL-as-required. 그러나 F1·F2 가 산출물의 실행 문장에 남았다 |
| Security (25%) | 90/100 | PASS | 카드 diff 시크릿 스캔 0 적중(`api key\|secret\|password\|token\|PRIVATE KEY\|ghp_\|sk-`), `internal/template/` 변경 0(로컬 전용 파일 유출 없음), 코드 표면 0 |
| Craft (20%) | 80/100 | PASS | `moai spec lint` 뮤턴트로 도구 생존 확인(`StatusValueInvalid` 발화), SHA placeholder backfill 2건 정직, `4362ba52f` 자기 정정. 감점: F3·F4·F6 |
| Consistency (15%) | 82/100 | PASS | Conventional Commits + `t531` 8/8 커밋, CHANGELOG 영문 파일 관례 준수·중복 0, §E 섹션 맵 보존. 감점: F4 커밋 subject, F5 |

가중 합 79.6/100. 별개로 **must-pass 방화벽**(Functionality)이 발화하므로 총평은 FAIL 이다.

## AC 재실행 결과 (레인 매트릭스 미소비)

| AC | 측정값 | 대조군 | 판정 |
|---|---|---|---|
| 001 | §4.1 슬라이스 `merge origin/develop` **0** (파일 전체도 0) | base 슬라이스 **2** (49행 bullet, 64행 bash) | PASS |
| 002 | `분기하는 트리` 13·15행 적중, 15행이 `develop` 을 지배 사본으로 지목 | 존재형 | PASS |
| 003 | 팔1 = 2 · 창 = 5행 · 팔2 = 2 · 팔3 = **0** | 뮤턴트 2종 모두 FAIL (아래) | PASS |
| 004 | `폐기`+`main` 동일 행 **3건**, 27행이 인용 금지 지목 | 존재형 | PASS |
| 005 | 수리본 **2**건(390행 리드 일괄 서술, 421행 리드 블록 안 명령 — 둘 다 리드 절차 안) | 백업 **3**건 | PASS |
| 006 | `미커밋` 19·21행, 21행이 「인용 일반의 규칙」 명시 | 존재형 | PASS |
| 007 | `.go` **0** | 전체 변경 **8**파일 | PASS |
| 008 | sha256 `23f8427589b739c705c8b17def0409c187af92ce5f27ee8bb87037a8376c9a7a` 일치 · **660**줄 | — | PASS |

**0 을 낸 대조군은 없다.** 「측정 불가」로 보고할 항목 없음.

### 뮤턴트 M-003 — 독립 재실행 (원본 무변경, `/tmp` 사본에서만)

| 뮤턴트 | 변이 | 팔2 | 팔3 | AC-003 |
|---|---|---|---|---|
| 원본 | — | 2 | 0 | PASS |
| A (35행만 지시로 뒤집음) | 금지 1줄 → `정리하려면 … 를 실행하라` | 1 | **1** | **FAIL** |
| B (33·35·37행 전면 뒤집음) | 창 전체를 지시 문면으로 | **0** | **1** | **FAIL** |

레인이 기록한 것은 B 형(2→0, 0→1)이고 내 재현과 일치한다. 프로브는 살아 있다.
감사 종료 후 `git status --porcelain` 무출력 — 추적 파일은 변이되지 않았다.

## 레인이 미리 표시한 4건 — 재측정 결과

**1. 40,000자 상한 초과 — 귀속은 성립, 변명은 미성립.** 세 좌표를 python3 `len()`(UTF-8 문자)로 독립 재측정했다:
base `bce6d7e08` **40,085** / after **41,769** / 델타 **+1,684**. 레인 수치와 정확히 일치하고, 85자 base 초과가
카드보다 앞선 것도 사실이다. 그러나 `char_budget.attribution` 의 「줄이는 것은 요구를 지우는 것이지
만족시키는 것이 아니다」는 성립하지 않는다 — `coding-standards.md:44-48` 이 지시하는 것은 **삭제가 아니라 이관**
(paths-scoped rule / `.moai/docs/` + 산문 포인터 / 매 세션 불필요한 내용 정리)이고, 이 파일 자신이 §18-27 에서
이미 그 사다리를 쓴다. 대안 이관을 재본 흔적이 없다. 또한 이 카드는 초과폭을 **85자 → 1,769자로 20배** 키웠다.
CI 가드는 없고(`.github/workflows/`·`internal/` 전수 검색), 규칙은 "should" 이며 파일은 유지자 전용이므로 blocking 은 아니다. → **F6**

**2. §E.4 `ownership_transition_lint: verdict: unmeasured` — 전부 검증됐고, 정직하다.** 판정 바이너리
`91d25bc61` 의 블롭을 직접 열어 확인했다: `defaultRules()` 에 `&OwnershipTransitionRule{}` 등록됨,
`lint_ownership.go` 에 `if rec.AuthoredByAgent == "" { return nil }` 실재. 블롭 해시 대응도 일치
(`91d25bc61`→`ef598d5c7c…`, `7ad9f8534`→`a53e9e2410…`). 이 브랜치 8커밋 전부 트레일러 0.
`moai spec lint --json` = `[]`. **결함이 아니라 올바르게 보존된 간극이다.** 다만 이 초록이 공허하지 않음을
따로 확인했다 — `/tmp` 사본에 `status: bogus-not-a-status` 를 주입하니 `ERROR StatusValueInvalid` 를 냈다.
린터 자체는 살아 있고, 침묵하는 것은 소유권 규칙 하나뿐이다.

**3. `draft → in-progress` 지각 착지 — 사실이며, 기록은 모범적이다.** 커밋별 frontmatter 추적:
`4af14489a`(실제 구현 커밋) `status: draft` → `9cf0d625f` 여전히 `draft` → `b9571a231`(chore) 에서 전이.
`spec-frontmatter-schema.md:99` 이 규정한 것은 「manager-develop, M1 커밋 시작, subject `fix|feat(SPEC-{ID}): M1 …`」
이므로 시점과 subject 둘 다 벗어났다. 다만 `b9571a231` 커밋 본문이 이중 근본원인(배차문의 과협 범위 지정 /
에이전트의 침묵한 충돌 해소)과 「AC 가 `status` 를 보지 않아 8/8 초록이 유지됐다」는 발견 경로까지 기록했고,
`run-phase-evidence.md:103` 에도 남아 있다. 이력 재작성 없이 기록한 것은 이 저장소의 확립된 대응이다.
→ **F4, optional**

**4. §G 간극 4건과 §A.1 `[추론]` 표지 — 열려 있고 약화되지 않았다.** `spec.md §G` 의 1~4번이 원문 그대로
있고, §A.1 의 `**[추론]**` 표지도 그대로다. `progress.md` 는 `open_gaps_preserved: 4` 를 §E.3·§E.4 두 곳에
각각 기록했다. §F 기계 가드 범위 밖 선언, 백업 보존-무분석 선언도 그대로다. **결함 없음.**

## Findings

- **F1** [medium] [**blocking**] `CLAUDE.local.md:227` (기존 §2.3) ↔ `CLAUDE.local.md:29-37` (신설 §0.4) —
  §2.3 의 `**[HARD]** update 실행 후 매번 검증한다. 전제: 실행 **전** 추적 파일 수정이 0이어야 diff 귀속이 가능하다`
  는 primary 체크아웃에서 도는 절차인데, 같은 파일의 신설 §0.4 가 그 체크아웃의 ` M CLAUDE.local.md` 를
  **영구·의도된 상태**로 못박고 `git restore CLAUDE.local.md` 를 금지했다. **그 전제는 이제 원리상 성립할 수 없다.**
  §2.3 을 문자 그대로 따르는 사람은 「추적 파일 수정 0」을 만들려고 하고, 그 유일한 경로가 §0.4 가 막은 restore 다 —
  REQ-CLPM-004 가 겨눈 바로 그 행위로 밀어낸다. 실측: primary 체크아웃 `main`/`7ad9f8534`,
  `git status --porcelain -- CLAUDE.local.md` = ` M CLAUDE.local.md`. 어떤 AC 도 §2.3 을 보지 않는다.
  **Required fix**: §2.3:227 의 전제에 예외를 명시한다 — 「추적 파일 수정이 §0.4 의 `CLAUDE.local.md` 하나뿐이어야」
  로 고치고 §0.4 를 교차참조한다.

- **F2** [medium] [**blocking**] `CLAUDE.local.md:389,404-405` ↔ `.claude/rules/local/gitflow-lane-protocol.md:120` —
  정정된 §4.1 은 「로컬 `develop` 이 `origin/develop` 보다 **앞설** 수 있다」한 방향만 다룬다. 형제 독트린은
  반대 방향을 명시한다: 「다른 레인의 병합이 `origin/develop`에 올라가면 통합 워크트리의 로컬 `develop`은 **뒤처진다**」
  (§11 이 그 갱신 절차를 소유한다, `:120`·`:135`). §4.1 어디에도 흡수 전에 로컬 `develop` 을 최신화하라는 절이 없다.
  **스테일한 로컬 develop 을 흡수한 레인은 원격 착지분이 빠진 베이스에서 재측정한다 — 이 카드가 고친 결함의 거울상이다.**
  현재 실측은 `git rev-list --count --left-right origin/develop...develop` = `0 0` 이라 발현 중은 아니다.
  즉 활성 파손이 아니라 **정정의 미완결**이다. 이 결함은 §H 가 선언하고 수행하지 않은 정합성 확인(F3)에서 나왔다.
  **Required fix**: §4.1 흡수 단계 앞에 한 절 추가 — `gitflow-lane-protocol.md §11` 로 로컬 `develop` 을 먼저 최신화
  (`git rev-list --count --left-right origin/develop...develop` 가 `0 0`)한 뒤 흡수한다.

- **F3** [low] [optional] `spec.md:205`, `plan.md:136` — §H 가 `.claude/rules/local/gitflow-lane-protocol.md` 를
  「정정 후 정합성 확인 대상」으로 선언했으나, 어느 밀스톤도 소유하지 않고 어느 AC 도 덮지 않으며
  `progress.md`·`run-phase-evidence.md` 어디에도 수행 기록이 없다(전수 grep 적중 0). 내가 대신 수행했다:
  `:135` 의 `git merge origin/develop` 은 **다른 작업**(develop 워크트리를 원격에서 갱신)이고 방향이 옳다 —
  결과는 깨끗하다. 그러나 같은 확인에서 F2 가 나왔다.
  **Required fix**: 확인 수행 사실과 결과를 `progress.md §E.4` 에 기록하거나, §H 의 선언을 내린다.

- **F4** [medium] [optional] `4af14489a`(구현 커밋)이 `status: draft` 로 착지했고 전이는 `b9571a231` 의
  `chore(...)` 커밋이 운반했다 — `spec-frontmatter-schema.md:99` 의 「M1 커밋 시작 + `fix|feat(SPEC-{ID}): M1 …`」
  에서 시점·subject 둘 다 벗어난다. 근본원인·발견경로가 커밋 본문과 `run-phase-evidence.md:103` 에 기록됐고
  이력을 재작성하지 않았다. 완결성을 위해 보고하며, 처분은 재량이다.

- **F5** [low] [optional] [**선행 결함 — 이 카드가 만든 것이 아니다**] `CLAUDE.local.md:404` —
  `git -C <카드워크트리> merge develop` 은 앞선 줄들이 독자를 `.claude/worktrees/develop` 에 세운 블록 안에 있어
  **교차 트리 `git -C`** 다. `gitflow-lane-protocol.md:39` 는 그 형태를 worktree-session 가드가 거부한다고 명시한다.
  base 에도 같은 형태(`git -C <카드워크트리> merge origin/develop`)가 있었으므로 선행 결함이고, 이 SPEC 의
  선언된 범위 밖이다. 다만 이 카드가 **바로 그 줄을 편집하면서** 형태를 그대로 뒀다는 점은 기록해 둔다.
  **Required fix(채택 시)**: 별도 카드 — `EnterWorktree(<카드워크트리>)` + 평문 `git merge develop` 로 교체.

- **F6** [low] [optional] `progress.md §E.4 char_budget.attribution` — 세 좌표(40,085 / 41,769 / +1,684)는
  독립 재측정으로 정확하나, 「줄이는 것은 요구를 지우는 것」이라는 필연성 주장은 근거가 없다.
  `coding-standards.md:44-48` 은 삭제가 아니라 이관 사다리를 제시하고, 대안 이관은 재보지 않았다.
  `ruling:` 줄이 이미 리드 판정을 운반하므로, 과잉인 것은 `attribution:` 산문이다.
  **Required fix**: 이관 후보를 하나 재보고 왜 안 되는지 기록하거나, 필연성 주장을 빼고 「리드 판정으로 수용한 초과」로 재서술한다.

## 산출물 문면 검증 (AC 밖)

CHANGELOG 와 §E.4 가 §0·§4.1 에 대해 주장한 것을 문면과 대조했다 — **모두 사실이다**:

- §0.1 의 판별식 문장 + 날짜/최신성 명시적 기각 — 실재(13·15·17행).
- §0.2 의 `git show <ref>:<path>` 검사법 — 실재(23행).
- §0.3 의 「main 커밋본 = 폐기된 제3의 모델」 — **실측으로 참**: `7ad9f8534:CLAUDE.local.md` 의 §4.1 제목은
  `### §4.1 로컬 통합 레인 (develop)` (현행과 다름), `push origin develop` **0건**, 278행에
  「**`develop`은 원격에 올리지 않는다.** push 금지, upstream 설정 금지」.
- §0.4 의 전제 — **실측으로 참**: primary 는 `main`/`7ad9f8534` 이고 ` M CLAUDE.local.md` 가 실제로 떠 있다.
- §4.1 정정 2곳 모두 착지(hunk 3개, 삭제 2줄이 정확히 두 흡수 지시). `refs/heads/develop` = `91d25bc61` 실재하므로
  `git merge develop` 은 실행 가능하다 — 문자열 치환이 아니라 동작하는 정정이다.
- §E.4 self-test A/B/C 재실행: CHANGELOG 중복 0 · AC 수 `acceptance.md` 기준 8 = CHANGELOG 주장 8 · 호출 경로 4/4 실재.
- SHA placeholder 규약: `b7344d957` 는 `sync_commit_sha: pending-backfill`, `25474b8a5` 가 `b7344d957` 로 채움.
  `4af14489a` 는 `run_commit_sha: pending-backfill`, `9cf0d625f` 가 `4af14489a` 로 채움. 규약 준수.

**§0 이 파일의 기존 내용과 충돌하는가** — F1 하나가 나왔다. 그 외 `git restore` 출현 3곳(`:233` `^ D` 삭제 복구,
`:242`·`:245` `git-strategy.yaml` 재적용)은 대상이 달라 충돌하지 않는다.

## Gaps — 관측하지 않은 것

- **Go 테스트 스위트 미실행.** 카드 지시로 금지됐고 이 카드의 `.go` 변경은 0 이다. 판정으로 세지 않는다.
- **`moai spec lint` 전 spec 트리 스윕 미완.** 백그라운드 실행이 시간 내에 끝나지 않았다(exit 1). 대신
  `/tmp` 사본 뮤턴트로 린터 생존만 확인했다 — 다른 SPEC 의 상태는 재지 않았다.
- **백업 파일 내용 미분석.** `spec.md §F` 범위 밖. sha256 과 줄 수, `push origin develop` 건수(3)만 쟀다.
- **`.moai/reports/t531/premise-remeasure.md` 의 수치를 원천에서 재유도하지 않았다.** 세 변종 표는 그 파일에서
  왔고, 나는 그중 변종 1(main 블롭)과 변종 3 방향만 독립 확인했다. 변종 2 의 원 관측은 재현하지 않았다(§G-2 와 같은 자리).
- **2026-09-07 사건 자체 미관측.** §G-3 이 열어 둔 그대로다.
- **워크트리 40+ 전수 미수행.** §G-1 그대로. 다른 워크트리의 `CLAUDE.local.md` 워킹 사본은 재지 않았다.
- **F2 의 발현을 실제로 만들어 보지 않았다.** 로컬 develop 을 뒤처지게 한 뒤 흡수해 베이스가 어긋나는 것을
  재현하지는 않았다 — 형제 독트린의 명시 서술과 현재 `0 0` 실측에 근거한 구조 판정이다.

## Residual-risk

- **F1·F2 는 매 세션 전량 로드되는 파일에 있다.** 두 결함 모두 문장 한 절씩이면 닫히지만, 닫히기 전까지는
  이 문서를 읽는 모든 레인에 전파된다. 특히 F2 는 실패가 조용하다 — 스테일한 베이스에서 낸 초록은 초록으로 보인다.
- **AC 집합의 사각이 닫히지 않았다.** 이 카드에서만 세 건(`status` 필드, F1, F2)이 「어떤 AC 도 보지 않는 자리」에서
  나왔다. 후속 카드가 같은 AC 양식을 복제하면 같은 사각을 물려받는다.
- **`OwnershipTransitionRule` 의 구조적 침묵**(§E.4 `residual_2`)은 이 저장소 전역이다. 그 규칙의 초록을
  소유권 준수 근거로 인용하는 모든 자리가 같은 간극 위에 있다. 이 카드 범위 밖이지만 살아 있다.
- **문자 상한은 이 카드 이후 더 나빠진다.** 후속 카드가 §0 옆에 같은 논리로 절을 더하면 초과폭은 누적된다.
  기계 가드가 없으므로 누적을 멈추는 것은 사람의 판단뿐이다.

---

판정: **FAIL** — F1·F2 두 blocking 을 닫고 재감사한다. 재감사는 열거된 결함 델타로 한정한다.
AC-CLPM-001~008 은 재실행 결과 유효하므로 다시 돌리지 않아도 된다.
