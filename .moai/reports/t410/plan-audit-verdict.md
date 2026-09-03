# SPEC-DRIFT-CLOSE-BODY-001 — plan-audit 판정 (t410)

Iteration: 1/1 (Tier S 상한 — `harness.yaml` `plan_audit_tier_ceilings` S=1)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.80** (Tier S PASS 임계 0.75 — `spec-workflow.md` § SPEC Complexity Tier)

> Reasoning context ignored per M1 Context Isolation. 리드 지시문이 옮긴 "ALREADY VERIFIED" 2건은 재도출하지 않고 교차확인만 했다(둘 다 옳다 — §부록 A).

## 감사 대상 고정 (pin)

**감사 중 레인이 아티팩트를 편집했다.** 판정은 아래 고정본에 대한 것이며, 그 뒤의 편집은 이 판정의 범위 밖이다.

| 항목 | 값 |
|---|---|
| 워크트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t410` |
| 트리 HEAD | `4e4607abe` (감사 개시 시 측정) |
| 고정 시각 | 2026-09-03T06:27:42Z |
| `spec.md` sha256 | `ff4489f3140aaf514d2f9cbc3e4bd6e39c81a12f5c6b07ccb69364d03f46ae96` |
| `plan.md` sha256 | `9b69231aa6625b17bed8a7cdebe1cbf234ea43a3780497a6fbee65a4fae2243a` |
| `progress.md` sha256 | `782121d3028f6e4c9564120f00fc8bfeb1aa5b89d0ee9867ca1968abce362a0c` |

감사 개시 시점의 판본에는 **차단급 결함이 하나 더 있었다** — REQ-DCB-003이 후보 게이트로 `closeInfixMatch`를 요구했는데, 그 술어가 이 카드의 확정 대상 `e979a4d13`(`chore(SPEC group C): Mx-phase close`)을 떨어뜨려 AC-DCB-001/004가 구조적으로 실패하는 상태였다. 감사 중 레인이 §5.4를 추가하며 스스로 이를 기각·수정했고, 고정본에는 남아 있지 않다. 독립 재현은 §부록 B에 남긴다.

## Must-Pass 결과 (M5 방화벽)

| # | 기준 | 판정 | 근거 |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | **PASS** | `grep -c '^\*\*REQ-DCB-00'` → 7. REQ-DCB-001~007 연속, 결번·중복·패딩 불일치 없음 |
| MP-2 | GEARS 형식 (요구 계층만) | **PASS** | spec.md:84~96. 001 `~해야 한다(shall)` Ubiquitous / 002 `While ~ 동안` State-driven / 003 `Where ~ 때에 한해` / 004·005 `~해서는 안 된다(shall not)` Unwanted / 006 Ubiquitous / 007 `When ~ 하면` Event-driven. 판정은 `REQ-XXX` 요구 계층에 대해서만 했다 — §3의 Given-When-Then은 검증 계층(AC)의 정규 형식이므로 Group 4에서 채점 |
| MP-3 | 프론트매터 12필드 | **PASS** | spec.md:2-14. 12필드 전부 정규명으로 존재(+ 선택 `tier: S`). snake_case 별칭 0건. `phase: "v3.2.0 target"` — 금지된 lifecycle 토큰 아님 |
| MP-4 | §22 프로그래밍-언어 중립성 | **N/A** | 대상이 `internal/spec` Go 단일 언어. 16개 언어 툴링을 다루지 않음 |
| MP-5 | D7 cross-SPEC 정합 | **PASS** | 참조 SPEC-ID 20개 전수 조회 → `retired`/`superseded`/`archived` 0건. NOT-FOUND 3건(`SPEC-FIX-ALPHA-001`/`SPEC-HIER-BETA-001`/`SPEC-DEP-GAMMA-001`)은 AC 픽스처용 합성 ID로 의도된 것이므로 D7-5 SHOULD를 발부하지 않는다 |
| MP-6 | D8 cross-platform 규율 | **PASS** | `grep -c 'syscall' spec.md` → 0. D8-4 자동 통과 |
| MP-7 | clarification 게이트 | **PASS** | `grep -rn 'NEEDS CLARIFICATION'` → rc 1 (0건). `research.md`는 Tier S라 부재 |

## 차원 점수 (rubric 기준)

| 차원 | 점수 | Band | 근거 |
|---|---|---|---|
| Clarity | 0.85 | 0.75~1.0 | 대부분 단일 해석. 모양 B의 스캔 의미(D2)와 REQ-DCB-002의 오류 경로(D3)에서 실제 해석 여지가 남는다 |
| Completeness | 0.85 | 0.75~1.0 | HISTORY/배경/요구/AC/범위밖 전부 존재, `### Out of Scope` H3 5개 각각 `-` 불릿 보유. Tier 근거 박약(D5), AC-DCB-002 Given 불완전(D6) |
| Testability | 0.75 | 0.75 | 강한 축: §5.3 실측 10줄 픽스처 + 뮤테이션 2종 + 공허방지 4개. 약한 축: AC-DCB-007 1항의 술어가 깨져 있고(D1) 모양 B의 음성 케이스가 없다(D2) |
| Traceability | 0.75 | 0.75 | REQ-DCB-005에 대응 AC가 없다(D7). 나머지 6개는 직접 대응 |

집계 0.80 ≥ Tier S 임계 0.75, must-pass 7/7 → **PASS 자격은 성립한다.** 다만 아래 blocking 3건(D1·D2·D7)이 전부 **M1 착수 전에 값싸게 닫히고 M1 이후에는 비싸진다** — 술어가 M1에서 확정되기 때문이다(plan.md §F 자체 서술). 그래서 PASS가 아니라 PASS-WITH-DEBT다.

## Defects Found

**D1. AC-DCB-007 1항의 판정 술어가 깨졌다 — 플랫폼에 따라 공허하거나 과대매칭** — `spec.md:164` — Severity: **major** — Class: **blocking**

`^[-+].*^status:` 는 한 패턴 안에 `^`가 둘이다. 이 머신에서 실측했다:

```
$ grep -nE '^[-+].*^status:' <pinned spec.md>
169:- 근거: 이 정정은 `status:`를 바꾸지 않으므로 amendment 절차 대상이 아니다(...)
```

`-`로 시작하고 본문 어딘가에 `status:`를 포함한 산문 줄이 매치됐다 — 즉 이 grep에서 중간 `^`는 빈 문자열로 소거되고 술어는 사실상 `^[-+].*status:` 다(**과대매칭**: diff 어디서든 `status:`를 언급하는 줄이 있으면 FAIL). GNU grep ERE에서는 중간 `^`가 만족 불가능한 앵커라 **항상 0건 → 항상 통과**(공허). 어느 쪽도 의도한 "변경된 `status:` 프론트매터 줄이 없다"가 아니다. M4가 추가할 HISTORY 행이 "status" 문자열을 담으면 BSD 쪽에서 거짓 FAIL이 난다.

이것은 `verification-completeness.md` §1.1이 이름 붙인 **report-not-verdict**의 한 형태이고, 동시에 §1.2(b)의 "붉게 만드는 입력" 미기재다.

**Required fix**: 술어를 `git diff -U0 -- .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md | grep -cE '^[-+]status:'` → 기대 0 으로 교체하고, **RED를 실측한다** — 해당 파일의 `status:` 줄을 일시 편집해 카운트가 1이 되는 것을 보고 되돌린 축자 출력을 원장에 남긴다. 붉은 것을 본 적 없는 초록은 아무것도 주장하지 않는다.

---

**D2. 모양 B의 스캔 의미가 미정의이고, 픽스처에 모양 B 음성 케이스가 없다** — `spec.md:§5.3 표 / 필수 픽스처 10줄` — Severity: **major** — Class: **blocking**

§5.3은 모양 B를 "표지를 걷어낸 줄이 그 자체로 conventional-commit subject → 기존 체인에 그대로 먹인다"로 정의한다. 두 가지가 안 적혀 있다: (a) 체인이 `completed`를 낼 때에만 close 선언인가, (b) 첫 모양-B 줄에서 판정이 끝나는가 아니면 자격 있는 줄을 만날 때까지 계속 훑는가.

실측 반례가 있다. `7beda68a5`의 본문은 `SPEC-WORKTREE-ENTRY-STRATEGY-001`을 명명하는 모양-B 줄을 **7개** 담는다:

```
$ git show -s --format=%b 7beda68a5 | grep -n 'WORKTREE-ENTRY'
1:  * fix(SPEC-WORKTREE-ENTRY-STRATEGY-001): M1 web auto-toggles default OFF (...)
15: * feat(SPEC-WORKTREE-ENTRY-STRATEGY-001): M3a launcher L2 absolute-path resolver for -w flag
43: * docs(SPEC-WORKTREE-ENTRY-STRATEGY-001): populate progress.md §E.2/§E.3 run-phase evidence (Round 1)
53: * docs(SPEC-WORKTREE-ENTRY-STRATEGY-001): M2-M6 doc-alignment (...)
108:* chore(SPEC-WORKTREE-ENTRY-STRATEGY-001): backfill Round 2 run_commit_sha 1201680b3 (...)
117:* docs(SPEC-WORKTREE-ENTRY-STRATEGY-001): sync-phase artifacts — 3-phase close (CHANGELOG + completed transition)
123:* chore(SPEC-WORKTREE-ENTRY-STRATEGY-001): backfill sync_commit_sha 14027ffec (...)
```

`ClassifyPRTitle`(`transitions.go:149`)은 close-infix를 prefix 루프보다 **먼저** 본다. 따라서 117번만 `completed`를 내고, 1·15·43·53은 비-`completed`, 108·123은 `shouldSkipCommitTitle`이 삼킨다(`isSPECIDScopedChore` + `backfill` + close-infix 없음). **첫 모양-B 줄에서 멈추는 구현은 이 SPEC의 진짜 close를 놓친다.** 그리고 이 SPEC은 `SPEC-WORKTREE-ENTRY-STRATEGY-001`을 조사의 TIGHT 목록에 담고 있으므로 가상의 경우가 아니다.

10줄 픽스처는 이 경우를 하나도 담지 않는다 — 3·4번이 둘 다 자격을 갖춘 모양 B라서, "첫 모양-B 줄에서 반환한다"는 뮤턴트가 10/10을 통과한다. `verification-completeness.md` §2의 mutant probe가 겨누는 정확한 형태다.

**Required fix**: (1) REQ-DCB-003/§5.3에 "모양 B는 `ClassifyPRTitle`이 `completed`를 낼 때에만 close 선언이며, 스캔은 자격 있는 줄을 만날 때까지 본문 전체를 훑는다"를 명시. (2) 픽스처에 11번 줄을 추가: `* fix(SPEC-WORKTREE-ENTRY-STRATEGY-001): M1 web auto-toggles default OFF (AutoCleanup+AutoMerge true→false)` (출처 `7beda68a5`, 기대 = 언급). (3) AC-DCB-002의 뮤테이션 의무에 세 번째 뮤턴트를 넣는다 — "첫 모양-B 줄에서 반환"이 11번 줄에서 반드시 실패해야 한다.

---

**D3. REQ-DCB-005에 대응 AC가 없다 — 그리고 모양 B가 그 보장을 구조적에서 조건부로 바꿨다** — `spec.md:92 (REQ-DCB-005), §5.1` — Severity: **major** — Class: **blocking**

§5.1은 "fallback 경로의 출력은 `completed` 하나뿐이므로 본문에서 `draft`나 `implemented`를 주워 오는 일이 애초에 불가능하다"고 REQ-DCB-005를 **구조적으로** 보장한다고 적는다. 그 논거는 §5.3이 모양 B를 도입하기 전에는 옳았다. 지금은 본문 줄이 `ClassifyPRTitle`에 먹여지고, 그 함수는 `implemented`·`in-progress`·`draft`를 반환한다(`transitions.go:132-177`). 보장은 이제 "구현이 비-`completed` 결과를 버린다"에 의존하며 — 이는 D2가 지적한 바로 그 이음매다 — **그것을 검증하는 AC가 하나도 없다.**

**Required fix**: AC-DCB-003에 (d) 케이스를 추가하거나 AC-DCB-002에 케이스를 붙인다 — "frontmatter가 `completed`이고 본문에 대상 SPEC의 비-close 모양-B 줄만 있을 때, git 함의 상태가 1차 워크의 값 그대로 남는다". D2의 11번 픽스처가 그대로 이 AC의 입력이 된다(두 결함이 한 수리를 공유한다는 것이 각각이 실재한다는 신호다).

---

**D4. REQ-DCB-002의 FALLBACK-ONLY 조건이 대상 코드의 한 갈래에서 도달 불가능하다** — `spec.md:86` — Severity: **minor** — Class: **blocking**

REQ-DCB-002는 "1차 워크가 `completed`도 terminal도 내지 못한 동안" 조회한다고 적는다. 오류 반환은 그 서술을 만족한다. 그러나 `DetectDrift`는 `inMemImpliedStatus`가 오류를 내면 ① 게이트에 닿기 전에 `continue` 한다:

```go
// internal/spec/drift.go:216-220
gitStatus, err := inMemImpliedStatus(commits, a.specID)
if err != nil {
    continue                      // ← ① 블록(:238)에 도달하지 않는다
}
```

창 전체가 분류 불가능한 SPEC은 fallback을 영영 못 받는다. REQ-DCB-001의 무조건 서술도 같은 간극을 물려받는다. **효과는 무해하다**(그런 SPEC은 표에서 통째로 빠지므로 DRIFT 오탐을 만들지 않는다) — 그래서 severity가 minor다. 문제는 요구가 계획된 배치보다 넓게 주장한다는 것이다.

**Required fix**: REQ-DCB-002를 "1차 워크가 상태를 **반환했고** 그것이 `completed`도 terminal도 아닌 동안"으로 좁히고, 오류 경로(창 소진)를 §4에 범위 밖 한 줄로 명시한다.

---

**D5. §5 후보표와 §5.2에 §5.4가 뒤집은 서술이 남아 있다** — `spec.md:§5 표 행 B, §5.2` — Severity: **minor** — Class: **blocking**

§5.4는 `closeInfixMatch`를 **측정으로 기각**했는데, §5 후보표의 행 B는 여전히 "코드에 이미 close-infix 개념(`closeInfixMatch`)과 FALLBACK-ONLY 조회 골격이 있다"를 B 채택의 근거로 든다. 그리고 §5.2는 "REQ-DCB-004의 줄 선두 술어"를 반대방향 오류의 방벽으로 지목하는데, 그 술어는 지금 REQ-DCB-003 소관이고 "줄 선두"는 두 모양 중 A 하나뿐이다.

`verification-completeness.md` §3(cross-layer revision sweep)이 이름 붙인 형태다 — 개정이 시작된 절에서 끝나지 않았다.

**Required fix**: 행 B의 근거를 FALLBACK-ONLY 골격 재사용만 남기고 `closeInfixMatch` 언급을 뺀다(또는 "§5.4에서 기각"을 명시). §5.2의 포인터를 `REQ-DCB-003 / §5.3 모양 A`로 고친다.

---

**D6. Tier S 근거가 예측뿐이고, "Tier를 올리지 말고 범위를 줄인다" 규칙에 정의된 동작이 없다** — `plan.md:20-27` — Severity: **minor** — Class: **optional**

예상 4파일/~210 LOC는 계획 자신이 요구하는 산출물을 빠뜨린다 — `.moai/reports/t410/run-evidence.md`(§E 원장 의무), `progress.md` §E.2/§E.3, 그리고 이 SPEC 자신의 `spec.md`/`plan.md`. 게다가 §5.4가 후보 창 게이트를 run-phase 결정으로 넘겼으므로 술어 헬퍼가 한 파일 더 붙을 수 있다. 그리고 "넘으면 범위를 줄인다"는 초과가 **범위 안 작업에서 나온 경우** 수행 가능한 동작이 없다 — 열거된 파일은 전부 어떤 AC가 요구한다.

**Required fix**: 파일 수의 모집단을 "소스 파일(원장·progress 제외)"로 명시하고, 초과 시 동작을 "원장에 기록하고 Tier 재판정을 리드에게 blocker로 올린다"로 바꾼다.

---

**D7. AC-DCB-002의 Given이 픽스처의 frontmatter 상태를 안 적는다** — `spec.md:115` — Severity: **minor** — Class: **optional**

언급 반례가 fallback에 **도달하려면** `SPEC-DEP-GAMMA-001`의 frontmatter가 `completed`여야 하고 1차 워크가 비-terminal 비-`completed`를 내야 한다(REQ-DCB-002). Given은 둘 다 안 적는다. 다만 뮤테이션 의무가 이 공허를 잡아낸다(픽스처가 애초에 도달하지 못하면 `strings.Contains` 뮤턴트도 실패하지 않고, AC 본문이 그 경우 "술어 또는 픽스처가 잘못된 것"이라고 이미 적어 두었다). 자기교정형이라 minor·optional.

**Required fix**: Given에 "frontmatter가 `completed`이고 1차 워크가 비-terminal 비-`completed`를 낸다"를 추가.

## 리드가 지목한 7개 축에 대한 응답

| # | 축 | 판정 |
|---|---|---|
| 1 | 물려받은 거짓 전제 | **깨끗하다.** §1.3이 t382 원인 ①의 반증을 명시하고, 어떤 REQ/AC도 그 원인 위에 서 있지 않다. §5 후보 A의 기각 근거는 반증을 올바르게 뒤집어 쓴다. 잔재는 D5(문구)뿐 |
| 2 | 수의 함정 | **깨끗하다.** AC-DCB-005가 특정 수치를 기대값으로 삼는 것을 명시 금지하고, §5.3이 5·6·7번을 TIGHT 안의 **측정된 오탐**으로 못박는다. 독립 확인: 리드가 지목한 `SPEC-AUTONOMY-TIERS-001` 외에 `80dea9684`→`SPEC-INTERNAL-TEST-002`("follow-up ... -002" 언급), `fb8aff006`→`SPEC-WORKTREE-BRANCH-GUARD-001`·`-OPTIN-001`("Depends on ... (both completed)")도 오탐이다 — TIGHT 15에는 최소 **4건**이 들어 있다. SPEC은 이미 셋을 픽스처로 담았다. 건수는 AC-DCB-005의 결과로 정해진다 |
| 3 | 공허성 | **부분 실패.** AC-001(선행 in-progress 단언)·002(뮤테이션 2종)·005(해제 0건이면 실패)·006(셀렉터 0매치는 초록 아님)은 실질적이다. AC-DCB-005는 0건에서 실제로 실패한다. AC-DCB-002의 명명된 뮤턴트는 실제로 죽는다(`strings.Contains` 뮤턴트는 `depends_on` 줄로 GAMMA를 무죄방면 → 케이스 실패). **AC-DCB-007 1항은 실패**(D1), 그리고 모양 B에는 죽지 않는 뮤턴트가 있다(D2) |
| 4 | 측정 환경 | **깨끗하다.** §6이 203을 모집단에서 배제하고 총계로 성패를 판정하는 것을 금지한다. AC-DCB-004가 브랜치를 `main`으로 못박는다. 독립 확인: `refs/heads/main` = `7ad9f8534`, `git merge-base --is-ancestor e979a4d13 main` → rc 0. 즉 확정 대상의 close는 판정 브랜치에 있고, AC-DCB-004의 기대 출력은 브랜치 선택으로 무효화되지 않는다 |
| 5 | 넓어진 노출의 인정(§5.2) | **인정 자체는 충분하고, 기계적 경계도 실재한다** — REQ-DCB-003의 두 모양 술어 + AC-DCB-002의 뮤턴트 2종 + 실측 10줄 픽스처가 그 경계다. 다만 §5.2가 가리키는 포인터가 낡았다(D5). 별도의 경계를 더 요구할 근거는 없다 |
| 6 | Tier S 정당화 | **약하다** — D6. 예측은 근거가 얇고, "범위를 줄인다"는 강제 가능한 절차가 아니다 |
| 7 | 범위 규율 | **의도는 깨끗하다.** AC-DCB-007이 변경을 HISTORY 1행 + `version:`/`updated:`로 한정하고, §4에 t382 재판정 금지가 독립 항목으로 있으며, `SPEC-ERA-H3-NARROWING-001`은 `status: completed` v0.5.0 그대로다. 다만 **그 한정을 기계적으로 확인하는 술어가 D1으로 깨져 있다** — 지금 상태에서 범위 규율은 주장이지 검증이 아니다 |

## Recommendation

M1 착수 **전에** 아래 셋을 닫는다. 전부 plan 계층 편집이고, M1이 술어를 확정한 뒤에는 되돌리기가 비싸진다(plan.md §F 자신의 순서 논거).

1. **D1** — AC-DCB-007 1항 술어를 `git diff -U0 -- <t382 spec.md> | grep -cE '^[-+]status:'` → 0 으로 교체 + RED 실측(일시 편집 → 1 → 되돌림)을 원장 의무로 명시.
2. **D2 + D3** — 한 수리다. REQ-DCB-003/§5.3에 모양 B의 자격 조건(`ClassifyPRTitle == completed`)과 전수 스캔 의미를 명시하고, 픽스처에 11번 줄(`* fix(SPEC-WORKTREE-ENTRY-STRATEGY-001): M1 ...`, 출처 `7beda68a5`, 기대 = 언급)을 추가하고, AC-DCB-003(또는 002)에 "본문에 비-close 모양-B 줄만 있으면 1차 워크 값 그대로"를 넣어 REQ-DCB-005를 덮는다. 뮤테이션 의무에 "첫 모양-B 줄에서 반환" 뮤턴트를 추가한다.

그 다음, 값싼 순서로:

3. **D4** — REQ-DCB-002를 "상태를 반환했고" 로 좁히고 오류 경로를 §4에 한 줄.
4. **D5** — §5 행 B 근거에서 `closeInfixMatch` 제거(또는 "§5.4에서 기각" 명시), §5.2 포인터를 REQ-DCB-003/§5.3 모양 A로.
5. **D6** — Tier 파일 모집단 명시 + 초과 시 동작을 blocker 상신으로.
6. **D7** — AC-DCB-002 Given에 frontmatter/1차 워크 전제 추가.

D4~D7은 run-phase 중 원장에 기록하며 처리해도 무방하다(부채로 수용 가능). D1~D3은 아니다.

**M3 전수 대조를 위한 참고 (판정 아님, 입력도 아님).** 감사 중 관측한 것 — TIGHT 15의 출처 커밋 6개 중 `e979a4d13`(3행)과 `2f449e189`(4행)과 `7beda68a5`(3행)은 §5.3 두 모양 중 하나에 걸리고, `80dea9684`(1행)·`fb8aff006`(2행)·`a83934d55`(1행)은 언급이라 걸리지 않는다. 이 관측을 **기대값으로 삼지 말 것** — AC-DCB-005가 명시적으로 금지하며, 실제 건수는 그 대조의 결과다.

---

## 부록 A — "ALREADY VERIFIED" 2건 교차확인

1. **FALLBACK-ONLY combined-scope 슬롯의 gate (a)가 파생 scope-prefix를 요구해 `chore(SPEC group C)`를 놓친다** — **옳다.** `drift.go:545-552`의 `hasScopePrefix`는 `chore(spec-<prefix>)` / `chore(spec-<prefix>:` / `docs(...)` 네 형태만 인정하고, `deriveScopePrefix("SPEC-V3R6-SESSION-HANDOFF-AUTO-001")`는 마지막 `-AUTO-001`을 걷어내 `SPEC-V3R6-SESSION-HANDOFF`를 낸다. `SPEC group C`는 어떤 ID에서도 파생되지 않는다.
2. **`moai spec lint`가 "No findings"** — 재도출하지 않았다. 구조적 방증만 확인했다: 정규 12필드 전부 존재(별칭 0), `### Out of Scope` H3 5개가 각각 `-` 불릿 보유(`OutOfScopeRule` 관례 충족).

## 부록 B — 감사 개시 판본의 차단급 결함 (고정본에서는 해소됨)

레인의 §5.4와 독립적으로 같은 결함을 측정했다. `closeInfixMatch`(`transitions.go:88-92`)가 인정하는 리터럴은 `3-phase close` / `4-phase close` / `mx-phase audit-ready` 셋뿐이고, 소문자화한 실측 subject에 대해:

```
NO-MATCH  chore(spec group c): mx-phase close (status implemented→completed, 2026-06-02)   ← 확정 대상
NO-MATCH  close out 2 specs with 3-phase lifecycle completion (doc-only) (#1210)
MATCH     docs(specs): batch sync-phase close — 5 b-grade specs (3-phase close) (#1240)
MATCH     docs(spec-internal-test-001): sync-phase artifacts + 3-phase close
MATCH     feat(spec-hierarchical-team-001): ... (tier m, 3-phase close) (#1394)
```

부수 확인: `closeInfixMatch`는 `drift.go:450`(`shouldSkipCommitTitle` D5 가드), `drift.go:530`(combined-scope gate b), `transitions.go:149`(`ClassifyPRTitle`) 세 곳이 공유한다. 따라서 그 상수 집합을 넓히는 것은 1차 워크와 기존 fallback의 동작을 함께 바꾸는 일이라 **REQ-DCB-006 위반**이다 — §5.4가 "별개 변경이라 이 SPEC 범위 밖"이라고 적은 판단은 옳다.

---

## 5절 증거 보고 (`verification-claim-integrity.md` §3)

**Claim** — 고정본 SPEC-DRIFT-CLOSE-BODY-001은 must-pass 7/7 통과, 집계 0.80(Tier S 임계 0.75 이상), blocking 결함 3건 + optional 4건을 안은 PASS-WITH-DEBT다.

**Evidence** — 이 문서 본문의 각 결함 항목이 명령과 축자 출력을 운반한다. 주요 측정: `grep -nE '^[-+].*^status:' <pinned spec.md>` → 169행 매치(D1) · `git show -s --format=%b 7beda68a5 | grep -n 'WORKTREE-ENTRY'` → 7행(D2) · `git merge-base --is-ancestor e979a4d13 main` → rc 0(축 4) · 참조 SPEC-ID 20개 status 전수 조회(MP-5) · `grep -c 'syscall'` → 0(MP-6) · `grep -rn 'NEEDS CLARIFICATION'` → rc 1(MP-7).

**Baseline-attribution** — 전부 이 실행에서, 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t410`, HEAD `4e4607abe`)에 대해, 위 pin 표의 sha256 3개가 가리키는 아티팩트 판본에 대해 측정했다. 코드 인용은 워크트리의 `internal/spec/{drift.go,drift_index.go,transitions.go}` 현재 상태다.

**Gaps** — 관측하지 않은 것: (a) `moai spec lint`를 직접 실행하지 않았다(리드가 재도출 금지, 구조적 방증만 확인 — 부록 A). (b) `moai spec drift --no-cache`를 실행하지 않았다 — 코퍼스 AC의 사전 상태는 `r3-drift-row.log`에 의존했고 이 트리에서 재측정하지 않았다(그 재측정은 plan.md §C의 run-phase 의무다). (c) TIGHT 15 전 행의 본문을 읽지 않았다 — 6개 출처 커밋 중 5개만 읽었다(`2f449e189`·`7beda68a5`·`80dea9684`·`fb8aff006`·`e979a4d13`; `a83934d55`는 SPEC의 인용을 받아들였다). (d) 뮤턴트를 실제로 실행하지 않았다 — 죽는지 여부는 코드 판독에 의한 추론이다. (e) `moai spec drift`의 캐시 동작을 직접 확인하지 않았다.

**Residual-risk** — (1) 레인이 감사 중 아티팩트를 편집하고 있었다. 이 판정은 pin 이후 편집에 대해 아무것도 주장하지 않으며, 편집이 계속됐다면 D1~D7 일부가 이미 닫혔을 수 있다 — 수리 여부는 pin 이후 판본에서 다시 봐야 한다. (2) D2의 "첫 줄에서 멈추는 구현" 뮤턴트가 실제로 통과하는지는 실행이 아니라 판독으로 판정했다 — M1의 픽스처가 서면 실측으로 확정된다. (3) grep의 중간 `^` 처리는 플랫폼 의존이며 이 머신에서만 실측했다 — GNU grep 쪽 거동(공허)은 문서 지식에 의한 추론이다. D1의 처방은 두 거동 모두에서 옳다. (4) Tier S 상한이 iteration 1이므로 이 판정을 뒤집을 두 번째 라운드는 없다 — 이의가 있으면 리드가 범위 축소 또는 명시적 상한 연장을 결정한다.

---

## 부록 C — pin 이후 관측된 드리프트 (판정 발행 직전)

판정 작성 중 `spec.md`가 다시 바뀌었다. 재측정:

| 파일 | pin sha256 | 발행 직전 sha256 | 변경 |
|---|---|---|---|
| `spec.md` | `ff4489f3…` | `7b31a226…` | **바뀜** — §6.1 신설(15행 추가) |
| `plan.md` | `9b69231a…` | `9b69231a…` | 동일 |
| `progress.md` | `782121d3…` | `782121d3…` | 동일 |

추가된 §6.1 "TIGHT 15는 조치 대상 목록이 아니다 — 오탐 4건이 측정됐다"는 이 감사가 축 2에서 독립적으로 측정한 오탐 4건(`a83934d55`/`fb8aff006`×2/`80dea9684`)과 **일치한다**. 두 관측이 수렴했다.

**이 추가는 D1~D7 중 어느 것도 닫지 않는다** — 전부 §2·§3·§5·plan.md §A 소관이고, §6.1은 §6에 붙었다. 따라서 위 판정과 부채 목록은 발행 시점 판본(`7b31a226…`)에도 그대로 유효하다. 그 이후의 편집에 대해서는 아무것도 주장하지 않는다.
