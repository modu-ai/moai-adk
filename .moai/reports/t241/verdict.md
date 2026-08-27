# t241 — 예측 장부 판정 (SPEC-VERIFICATION-COMPLETENESS-001 §A.5)

- 카드: t241 (재범위 — 규칙 저작이 아니라 **착지한 규칙의 예측 판정**)
- 워크트리: `.claude/worktrees/t241`, 브랜치 `WT-check-must-fail`
- 측정 트리: `origin/develop` = `d34a789a4`
- 판정일: 2026-08-28

## Claim

`SPEC-VERIFICATION-COMPLETENESS-001` `plan.md §A.5` 예측 장부의 6행(VC-1~VC-6)에 `verified` 열을 신설하고 판정했다. **여섯 행 모두 `false`다.**

규칙(`verification-completeness.md`, 착지 `7f5b6a947`, 2026-08-25 13:05:04 +0900)이 이름 붙인 여섯 실패 형태는 착지 이후에도 **전부 재발했다**. 다만 재발이 곧 규칙의 무용은 아니다 — 아래 §해석을 볼 것.

## Evidence

### 코퍼스 확정과 그 정정

착지 이후 착지한 감사 산출물을 열거했다:

```
git log --since="2026-08-25 13:05" --name-only --format="" -- ".moai/reports/*audit*" ".moai/reports/*verdict*"
→ 49 unique paths
```

**이 열거는 틀렸고, 두 수집 레인이 서로 독립적으로 잡아냈다.** `--since` 는 committer date 를 본다:

```
git log -1 --format="%h authored=%ad committed=%cd" --date=iso eff98eabc
→ eff98eabc authored=2026-08-25 04:05:26 +0900 committed=2026-08-27 17:40:19 +0900
```

`.moai/reports/t228/plan-audit-iter1..iter5.md` 5건은 규칙 착지보다 **9시간 앞서 작성**됐는데 08-27 리베이스로 코퍼스에 딸려 들어왔다. 더 중요한 것은 `plan.md §A.4` 가 **바로 그 t228 감사를 이 규칙들의 원천 관측으로 지목**한다는 점이다(`:97` VC-4 ← t228 iter2 E1/E2, `:99` VC-5 ← iter3, `:100` VC-6 ← iter4 N1). 규칙의 입력을 규칙의 성적표로 세면 순환 논증이 된다. **5건 제외, 유효 코퍼스 44건.**

경계 1건: `.moai/reports/t205/verdict.md` 는 author date 가 `2026-08-25 13:05:48` — 규칙 착지 **44초 뒤**라 그 세션이 규칙을 읽었을 수 없다. 어느 검색어에도 걸리지 않아 판정에 영향 없음. 은닉하지 않고 기록한다.

### 행별 증거 (원문 인용, 인용 6건은 판정자가 직접 재확인)

**VC-1 — 신규 3건, 게이트 통과 2건.**

- `t301/plan-audit.md:159` — `**D8. AC-FSD-005와 AC-FSD-009 첫 블록이 report-not-verdict 형태다** ... Class: blocking`. 규칙이 08-25 에 도입한 용어를 그대로 쓴다. `plan-audit-iter2.md:179` 에서 폐쇄 → **차단됨**.
- `t301/plan-audit-final.md:209` (직접 확인) — `| N6 | AC-FSD-007의 "같은 문장" 은 눈 확인 잔존 | **run-phase 반입 가능.** iter-3 판정 유지 |` → **기계 판정자 부재를 명시적으로 수용하고 채택.**
- `t303/sync-audit.md:131` — `Vacuous-green. The cited selector -run 'TestInternalContentLeak|TestRuleTemplateMirror|TestCommandsAudit' names three tests; ... neither exists anywhere under internal/. ... Carried into the sync closure unreviewed.` → **sync 종결까지 도달.** 이 건은 판정자 본인이 만든 결함이며(카드 t303 §E.2), 자기 결함을 자기 장부에 계수한다.

**VC-2 — 신규 7건, 채택까지 생존 1건.**

t298 ×2, t301 ×3, t316 ×1, t235 ×1. 6건은 감사 루프 안에서 닫혔다. 생존 1건은 `t301/plan-audit-final.md:210` (직접 확인):

> `| N7 | 전량 어휘족을 버린 우회 표현 mutant 생존 | **run-phase 반입 가능.** grep으로 임의 의역까지 닫는 것은 원리적으로 불가 |`

**VC-3 — 확정 신규 1건, 모호 2건.**

`t301/plan-audit-final.md:114` (직접 확인):

> `**명령 1이 앞에 있는 한 명령 3은 빨간불이 될 수 없다.** 저자가 AC 본문에 적은 "... 둘 다를 잡는 유일한 기계 장치" 는 소스에 비추어 **거짓**이다.`

승인 사슬이 이 건의 핵심이다 — 그 체크를 처방한 것은 **한 회차 앞선 감사 자신**(`plan-audit-iter3.md:232`)이고, 저자가 `acceptance.md` v0.4.0 에 채택했으며, 다음 회차가 구조적 항상-녹색으로 판정해 뒤집었다. 게이트 결과는 FAIL 이라 run 에는 들어가지 않았다.

**VC-4 — 신규 5건, 채택까지 생존 1건.**

생존 건은 `t306/plan-audit.md:25` (직접 확인):

> `AC-014's RED-now is defined counterfactually ...; AC-013 concedes its red is "in the future sense" — honest, but neither observes a red on tree d29b8942e.`

이 SPEC 은 그 상태로 **0.94 로 통과**했다. 두 칸 규율의 RED-now 칸이 관측 없이 채택된 것이며, 지적은 됐으나 차단 사유로 분류되지 않았다.

**VC-5 — 신규 1건, 착지 전 수리.**

`t298/plan-audit-iter2.md:86` (직접 확인):

> `**N1. REQ-OVERCLAIM-SURVIVES-D4-REPAIR** ... the D4 repair bounded the plan and the risk section but left the requirement itself mandating the forbidden claim.`

**iter-1 수리가 만들어낸 결함**이다. 수리 전에는 두 표면이 같이 틀렸고(일관되게 틀림), 수리가 한쪽만 고쳐 요구 층과 계획 층의 살아 있는 모순으로 바꿨다. 규칙 §3 이 이름 붙인 최악의 형태다. 착지본을 직접 확인한 결과 수리됐다 — `spec.md:191` 이 `restored session-anchored liveness guarantee` 로 읽히며, 문제의 `restored serialization guarantee` 문면은 남아 있지 않다.

**VC-6 — 신규 2건, 통과 1건.**

`t272/verdict.md:127` (직접 확인):

> `Command: git diff --stat origin/main -- .claude/skills/moai-domain-svg-infographic/ ...`
> `Result: empty output (no diff)`

불변 주장을 움직이는 ref 에 대고 세웠고, 그 문서 어디에도 SHA 핀이 없어 **오늘 재현이 불가능하다.** 규칙 §4 가 예시로 든 형태 그대로이며, 이 verdict 는 착지했다.

t274 는 감사 중 브랜치가 움직여 첫 측정이 오염됐고, 감사가 그것을 탐지해 판정에 쓰이는 수치를 전량 고정 SHA 로 재측정했다(`sync-audit-verdict.md:53`). **형태는 발생했고 규칙의 처방이 작동한 경우**로, 통과가 아니라 자가교정으로 계수한다.

면제 판별식은 지켜졌다 — t292 는 카드 주제 자체가 `origin/main` 의 고아 스탬프라 움직이는 ref 유지가 옳고, 고정하면 주장이 무너진다. 위반으로 계수하지 않았다.

## Baseline-attribution

- 규칙 착지: `7f5b6a947`, author == committer == `2026-08-25 13:05:04 +0900`.
- 유효 코퍼스: 위 열거 49건에서 t228 5건 제외 = 44건, author date `2026-08-26 01:04` ~ `2026-08-27 23:43`.
- 수집: 읽기 전용 3레인(VC-1/3, VC-2/4, VC-5/6)이 같은 코퍼스를 서로 다른 렌즈로 훑음. 각 레인은 판정을 반환하지 않고 인용만 반환했으며, 판정은 이 문서가 한다.
- 위 인용 6건(`t301` ×3, `t306`, `t298`, `t272`)은 판정자가 `sed -n` 으로 원문을 직접 재판독했다.

## 해석 — false 가 뜻하는 것과 뜻하지 않는 것

여섯 행이 모두 false 이지만, **규칙이 아무 일도 하지 않았다는 뜻은 아니다.** 같은 코퍼스가 반대 방향 증거도 낸다: 감사 산출물들이 `verification-completeness.md` 를 파일명과 절 번호로 인용하고, 뮤턴트를 이름으로 지목해 죽이며, 비공허성을 계수로 증명한다(`t303/sync-audit.md:43` `non-vacuity: 911 shipped keys, 974 inventory entries, 344 struct fields`). t316 은 뮤턴트 8개를 열거해 7개를 죽였고 탈출 2건을 차단 결함으로 올렸다.

읽어야 할 그림은 이것이다 — **감사 층은 규칙을 흡수했고, 저작 층은 아직 아니다.** 그리고 감사가 100% 는 아니라서, 라운드마다 몇 건이 채택까지 살아남는다:

| 행 | 발생 | 채택까지 생존 |
|---|---:|---:|
| VC-1 | 3 | **2** |
| VC-2 | 7 | **1** |
| VC-3 | 1 (+모호 2) | 0 (게이트 FAIL) |
| VC-4 | 5 | **1** |
| VC-5 | 1 | 0 (착지 전 수리) |
| VC-6 | 2 | **1** |

**정책 층에는 기계적 탐지기가 없다.** 가장 선명한 증거는 판정자 자신의 F2 다 — 규칙은 08-25 에 착지했고, F2 는 08-27 에, 그것도 규칙 `paths:` 범위 **안**인 `.moai/specs/**/progress.md` 에서 났다. 커버리지 공백이 아니다. 규칙이 닿는 자리에서, 규칙이 이름 붙인 형태가, 이틀 뒤에 다시 났다.

## 후속 카드 후보 (발행은 운영자 승인 — 리드 처리)

| # | 축 | 무엇 | t333 겹침 |
|---|---|---|---|
| C1 | 선택자 계수 | `-run` 정규식이 고른 케이스 수를 세지 않으면 초록이 무의미하다. 기계 층 후보: 증거 행에 `-run` 이 있는데 매치 수 기록이 없으면 경고하는 훅/린트. VC-1 의 3건 중 2건이 이 형태 | 없음 — 발현이 아니라 저작 시점 |
| C2 | 미고정 불변 주장 | verdict·progress 증거 행이 `origin/main`·`origin/develop` 를 불변/PRESERVE 주장에 쓰면서 SHA 핀이 없으면 경고. **면제 판별식(주제가 mainline 인 provenance 서술)을 반드시 구현해야 오탐이 안 난다** — t292 가 그 사례 | 없음 |
| C3 | RED 미관측 채택 | 두 칸 중 RED-now 가 "counterfactual"·"future sense" 로만 채워진 AC 를 감사가 차단 사유로 올리도록 임계 조정. t306 이 0.94 로 통과한 형태 | 없음 |
| C4 | 장부 지표 결함 | 예측을 "감사 지적 0건" 으로 쓰면 규칙이 잘 들을수록 성적이 나빠 보인다. 다음 장부는 **채택까지 생존 건수**를 예측으로 쓸 것 | **겹침** — t333 이 발현 기대치를 설계할 때 같은 함정 |
| C5 | 규칙의 발현 자체 | 이 판정이 보여준 것은 정책 층 규칙에 발현 관측이 없다는 것이다. 규칙이 인용되는지·적용되는지를 무엇이 확인하는가 | **t333 로 넘김 후보** — t333 (a)(b)(c) 축과 같은 질문 |

## Gaps (관측하지 않은 것)

- **SPEC 본문(`spec.md`/`plan.md`/`acceptance.md`/`progress.md`)을 전수 훑지 않았다.** 세 레인 모두 보고서 코퍼스를 정본으로 삼았고, 감사가 표면화하지 않은 인스턴스는 이 판정에 잡히지 않는다. SPEC 디렉터리 다수가 git 미추적이라 날짜 귀속도 불가능했다.
- **탐지가 grep 형태다.** 두 행(VC-5·VC-6)은 산출물이 **조용히** 저지를 수 있는 결함인데, 이 방법은 감사가 이름 붙였거나 인식 가능한 명령 형태를 쓴 것만 찾는다. `origin/` 없이 브랜치 이름이나 `HEAD~N` 로 표현된 미고정 주장은 놓친다.
- **VC-1 의 건수는 장부 문면 해석에 달려 있다.** "신규 체크 승인물" 이 규칙 이후 작성됐으나 게이트에서 차단된 것을 포함하면 3, 통과한 것만 세면 2다. 이 문서는 발생 3 / 생존 2 로 갈라 적어 해석에 의존하지 않게 했다.
- **채택 여부를 최종 산출물로 확인한 것은 VC-5 뿐이다.** 나머지 행의 "수리됨" 은 감사관 자신의 후속 회차 진술에 근거한다.
- **리드가 전한 lane-15 의 t322 관측**은 산출물이 primary 체크아웃에 있어 직접 읽어 1차로 승격했다(`t322/plan-audit.md:314` 공허 AC, `plan-audit-iter2.md:260` 수리 확인). 다만 t322 는 코퍼스 열거에 잡히지 않았다(미추적) — 판정 숫자에는 넣지 않고 해석의 방증으로만 썼다.

## Residual-risk

- 이 판정 자체가 정책 층 산물이라 **자기가 진단한 병에 걸릴 수 있다.** 여섯 행을 false 로 적었지만, 그 판정을 기계적으로 재현할 장치는 없다. 다음 라운드가 이 표를 재측정 없이 인용하면 그것이 VC-6 이 금지하는 형태다 — 재인용 시 코퍼스를 다시 열거할 것.
- 유효 코퍼스가 44건, 관측 창이 이틀이다. 라운드 하나의 표본이며 추세가 아니다.
- 감사 층이 흡수했다는 판단은 **감사가 스스로 남긴 기록**에 근거한다. 감사가 놓친 것은 이 방법으로 보이지 않는다.
