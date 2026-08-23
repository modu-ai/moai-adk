# SPEC Review Report: SPEC-DEPLOY-RESULT-WIRE-001

Iteration: 2/2 (Tier M ceiling — `.moai/config/sections/harness.yaml:77` `plan_audit_tier_ceilings.M: 2`)
Verdict: **FAIL**
Overall Score: **0.750** (Tier M PASS 임계값 **0.80** — `spec-workflow.md:140`)

작성자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정 대상은 디스크의 artifact 뿐이다. iter-1 보고서는 **결함 목록의 출처로만** 읽었고, 그 보고서의 실측값은 하나도 승계하지 않았다 — 아래 모든 수치는 이번 감사에서 다시 쟀다.

## 0. 고정 상태 (audit anchor)

```
$ git log --oneline -1
75e7ce381 docs(spec): revise SPEC-DEPLOY-RESULT-WIRE-001 to v0.2.0 (card t176)
$ git status --short
(빈 출력)
$ git branch --show-current
WT-deploy-result-wire
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t176
```

감사 중 SPEC artifact 편집 0건, 임시 파일 생성 0건.

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-DRW-[0-9]*' spec.md | sort -u` = `REQ-DRW-001..010`. 순차, 결번 0, 중복 0, zero-padding 일관.
- **[PASS] MP-2 GEARS 형식 준수** — 요구사항 계층 10건 전건 대응(`spec.md:124-133`). 001/002/003 event-driven(`…할 때(When) … shall`), 004/008/010 ubiquitous(+`shall not` 결합), 005 where(`구현하지 않는 경우(Where)`), 006/007 unwanted(`shall not`), 009 state-driven+unwanted(`갖기 전까지(While) … shall not`). **판정은 요구사항 계층에 대해서만** 내렸다 — `acceptance.md` 의 AC 9건은 검증 계층이고 Given-When-Then 이 정상 형식이므로 Group 4 에서 별도 채점했다.
- **[PASS] MP-3 YAML frontmatter 유효성** — `spec.md:2-14`. 정본 12필드 전건 존재·타입 적합(`version: "0.2.0"` 인용 semver, `created`/`updated` ISO, `priority: P2`, `lifecycle: spec-anchored`, `tags` CSV 문자열). 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `tier: M` 은 허용 선택 필드.
- **[N/A] MP-4 §22 언어 중립성** — 단일 언어(Go) 저장소 내부 코드 대상. `internal/template/templates/` 아래 신규 파일 없음(`spec.md:163` 자체 명시, 실측 일치). 자동 PASS.
- **[PASS] MP-5 D7 교차 SPEC 조정** — 본문 참조 SPEC 1건(`SPEC-CODEX-SKILLS-CANONICAL-001`), `status: implemented`. retired/superseded/archived 아님 → 조정 의무 미발생. BLOCKING 0건.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -c syscall spec.md` = **0**. 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-DEPLOY-RESULT-WIRE-001/` = 0건.

must-pass 실패 0건. FAIL 은 아래 rubric 점수(0.750 < 0.80)와 blocking 결함에서 나온다.

## Category Scores

| 차원 | 점수 | Rubric Band | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 요구사항 10건 자체는 단일 해석. 다만 문서 내부 자기모순 1건(§A.6 `:205` ↔ §D `:196`, N1)과 인용 off-by-one 1건(N3)이 남아, "실측 정확성" 을 권위 근거로 쓰는 문서의 신뢰 표면을 깎는다. |
| Completeness | 1.00 | 1.0 | HISTORY(`spec.md:19`) · §A 기준선 · §B 결정 · §C 요구사항 · §D 범위 밖(`### Out of Scope — …` H3 **5개**, 각 `-` 항목 보유) · §E 비기능 · §F 교차참조. frontmatter 완전. acceptance 매트릭스·심각도·추적성·DoD·전방점검 전건 존재. |
| Testability | 0.50 | 0.50 | AC 9건 중 **3건**이 판정 불능 또는 판단 개입: AC-DRW-004 2번 팔이 명세된 올바른 구현에서 **붉어진다**(N2, blocking), AC-DRW-003 의 2번 위증 검사가 자기 Given 에서 **재현 불가**(N4, blocking), AC-DRW-009 1번 단언의 검증 수단 미지정(N5). |
| Traceability | 1.00 | 1.0 | 양방향 실측. REQ-DRW-001..010 전건이 최소 1개 AC 에 덮이고(003 은 AC-004 2번 팔 + AC-006 공동), AC-DRW-001..009 전건이 실재 REQ 를 지목한다. 고아 AC 0건, 미덮인 REQ 0건. 매트릭스(`acceptance.md:9-17`)와 본문 일치. |

**Overall = 조화평균(0.75, 1.00, 0.50, 1.00) = 4 / (1.3333 + 1 + 2 + 1) = 4 / 5.3333 = 0.750.** Tier M 임계값 0.80 미달.

> **판정이 갈리는 지점을 밝힌다.** Testability 를 0.75 로 매기면 총점은 0.857 이 되어 PASS 다. 0.50 으로 매긴 근거는 N2 다 — 이것은 "정밀하지 않은 판정" 이 아니라 **틀린 판정**이다(올바른 구현을 붉게 만든다). rubric 에 그 밴드가 따로 없으므로, 판정이 자기 존재 이유를 수행하지 못하는 상태를 0.50 밴드("판단 개입이 필요한 AC 다수")로 읽었다. 감사 규율의 "의심스러우면 FAIL" 도 같은 방향이다.

> **점수 추이**: iter-1 0.75 → iter-2 0.750. **하락이 아니므로 STOP 신호는 발하지 않는다.** 다만 결함 7건 중 5건을 실제로 닫고도 총점이 제자리인 것은, 개정이 blocking 결함 2건(N2·N4)을 **새로 들여왔기** 때문이다.

## 1. iter-1 결함 7건 — 닫힘 판정

판정 기준: **그 결함이 지목한 실패 상황에서 artifact 가 이제 붉어지는가.** 그럴듯한 재작성만으로는 닫히지 않는다.

| ID | 판정 | 재도출 근거 |
|---|---|---|
| **D1** (critical) | **닫힘** | 세 귀결 전부 확인. ① 사실 주장: §A.4 표(`spec.md:65`)가 clean-reinstall 을 "**stdout**" 으로 다시 쓰고 nil 분기의 프로덕션 미도달을 명시, 결론(`:68`)이 "stderr 는 init 하나, stdout 위험은 둘" 로 교체 — **내 재측정과 일치**(§2). ② plan M3(`plan.md:64`)이 `out` 이 아니라 신설 `CleanReinstallOptions.ErrOut` 주입으로 재배선하고 [HARD] 로 iter-1 지시의 위반성을 기록. ③ AC-DRW-003 Given(`acceptance.md:45`)이 "`CleanReinstallOptions.Out` 에 **stdout 을 주입한 채로**(프로덕션 호출부 `update.go:425`·`:627` 이 실제로 하는 것과 동일)" 로 프로덕션 배선에 고정되고, `:52` [HARD] 가 미주입 구성 판정을 금지. **요구된 "Given 이 프로덕션 배선에 고정" 조건 충족.** |
| **D2** (major) | **닫힘 — 그 과정에서 N2 를 들여왔다** | REQ-DRW-007(`spec.md:130`)이 "통지 **전체**(복사 폴백 요약과 `failed` 출력을 모두 포함)" 로 확장되고 REQ-DRW-003(`:126`)이 `failed` 를 "개수 + 최대 3건" 으로 상한. §B.D4(`:114`)가 모순을 명시 정정. AC-DRW-004 에 `failed` 팔 추가(`acceptance.md:65`). 모순 자체는 해소 — 다만 추가된 팔의 N 값이 상한과 충돌한다(N2). |
| **D3** (major) | **닫힘** | AC-DRW-003 제목·Given 이 "stdout 위험 **두 호출부**" 로 각각 결박되고(`acceptance.md:43,45`), Then 이 "두 호출부 **각각에 대해**" 두 단언을 요구(`:47-50`). `:54` [HARD] 가 "1번만 쓰지 않는다 — 두 스트림 모두에 쓰는 구현이 통과한다" 로 iter-1 D3 이 지목한 회피 형태를 봉쇄. |
| **D4** (minor) | **미해결 (부분 정정)** | §A.6(`spec.md:78`)은 `:205` 로 정정됐다. **그러나 iter-1 D4 가 지목한 두 위치 중 두 번째(§D)는 그대로다** — `spec.md:139` 이 여전히 `skill_mirror.go:196`. HISTORY(`:23`)는 "`196` → **`205`**(재측정)" 을 완료로 적고 있어, **HISTORY 가 문서 자신에 의해 반증되는 완료 주장**을 담는다. → **N1**. |
| **D5** (minor) | **닫힘 (실질) — 부수 결함 N6 동반** | REQ-DRW-010(`spec.md:133`) + AC-DRW-009(`acceptance.md:122-133`) 신설, plan R1(`plan.md:14`)·M3 닫힘조건(`:66`)이 참조. seam 기본값이 프로덕션 배포기이자 `ResultDeployer` 를 만족함을 단언한다. 다만 plan §E·§H 의 AC 열거가 `AC-DRW-001..008` 로 스테일이라 **신설 가드 AC 가 자가검증 목록에서 빠져 있다**(N6). |
| **D6** (minor) | **닫힘 — 실측 정확** | §A.7 신설(`spec.md:80-90`). 세 행 인용 전건 재측정: `skill_mirror.go:217` `cannot replace stale link: %v`, `:226` `cannot create %s: %v`, `:243` `symlink and copy both failed: %v …` — **3/3 일치**. `Mode: MirrorModeFailed` 는 `:216`/`:225`/`:242`, 문구는 그 다음 줄이므로 인용 대상(문구 행)이 정확하다. `:205` 가 `MirrorModeSkipped` 라는 부기(`:90`)도 실측 일치. |
| **D7** (minor) | **형식은 균일, 유효성은 아니다** | AC 9건 **전건**에 `**위증 검사**` 절 존재(`acceptance.md:31,41,56,71,85,97,110,120,133` — 9/9). 그러나 AC-DRW-003 의 2번 위증 검사는 자기 Given 에서 재현할 수 없다(N4). **규칙은 균일하게 서술됐고, 균일하게 검증되지는 않았다.** |

**요약: 닫힘 5건(D1·D2·D3·D5·D6), 미해결 1건(D4 → N1), 부분 1건(D7 → N4).**

## 2. 재측정 원장 (iter-1 에서 아무것도 승계하지 않았다)

| 주장 | 출처 | 실측 명령 / 관측 | 판정 |
|---|---|---|---|
| `out := cmd.OutOrStdout()` @ `update.go:154` | §A.4 | `grep -n "out :=" internal/cli/update.go` → `154`, `696`, `783` | 일치 (범위 내 유효 정의는 `:154` 하나) |
| 프로덕션 `CleanReinstallOptions{` 2곳 | §A.4 | `grep -rn "CleanReinstallOptions{" internal/ \| grep -v _test` → `update.go:418`, `update.go:624` | 일치 |
| `Out: out` @ `:425` / `:627` | §A.4 | `grep -n "Out:" internal/cli/update.go` → `425: Out: out`, `627: Out: out` | 일치 |
| 호출 위치 `:418` / **`:625`** | `spec.md:65` | 실측 **624** (`if _, runErr := runCleanReinstall(planCtx, …`) | **불일치 — off-by-one (N3)** |
| `emitDryRunReinstallPlan` 호출 `:362` | §A.4 | `grep -n "emitDryRunReinstallPlan("` → 호출 `362`, 정의 `592` | 일치 |
| `nil ⇒ os.Stderr` @ `update_clean_install.go:55-56,138-140` | §A.4 | `:56 Out io.Writer`(주석 `:55`), `:138 out := opts.Out`, `:139-140 if out == nil { out = os.Stderr }` | 일치 |
| 오귀속 문구 행 = **205** | §A.6 | `grep -n "non-symlink entry already exists" skill_mirror.go` → `205` | 일치 (§D 의 `196` 은 불일치 — N1) |
| `failed` 문구 3곳 `:217`/`:226`/`:243` | §A.7 | 위 D6 행 | 3/3 일치 |
| 미러 대상 스킬 34건 | §A.6·AC-004 | `ls -d internal/template/templates/.claude/skills/*/ \| wc -l` → **34** | 일치 |
| `update_template_sync.go:130` 인라인 배포기 / `:323` `deployer.Deploy(` | plan R1 | `130: deployer := template.NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)`, `323: … deployer.Deploy(ctx, …)` | 일치 |
| `NewDeployerWithRendererAndForceUpdate` 가 `ResultDeployer` 를 만족 | AC-009 2번 | `deployer.go:129` 시그니처 + `deployer.go:74` `var _ ResultDeployer = (*deployer)(nil)` 컴파일 시점 단언 | **성립** — AC-009 2번 단언은 현재 코드에서 참이며 회귀 가드로 유효 |
| 선행 SPEC `status` | §F | `grep -m1 '^status:' SPEC-CODEX-SKILLS-CANONICAL-001/spec.md` → `implemented` | 일치 |

## 3. 일반화된 규칙은 균일하게 적용됐는가 — **서술은 균일, 검증은 아니다**

작성자가 D1 에서 끌어낸 규칙은 두 문장이다(`spec.md:22`): **(a) 판정은 프로덕션 배선에서 이뤄진다, (b) 각 AC 는 어떤 잘못된 구현에서 붉어지는지를 본문에 적는다.**

- **(b) 형식 적용 — 균일하다.** 9/9 AC 가 `**위증 검사**` 절을 갖는다(실측). plan 에도 AP-9·AP-10 으로 안티패턴화됐다(`plan.md:88-89`).
- **(a) 적용 — 균일하지 않다.** AC-DRW-003 은 Given 을 프로덕션 배선에 고정했으나(D1 닫힘), **그 배선의 한 층 위**(`update.go:425`·`:627` 의 옵션 리터럴)는 어느 AC 도 보지 않는다. 그래서 AC-DRW-003 이 스스로 적은 2번 위증 검사 — "`ErrOut` 을 추가했으나 **호출부에서 주입하지 않아** nil 로 남는 구현" — 이 그 AC 의 Given 안에서 재현될 수 없다(N4). 규칙이 파생 결함 D5·D7 에는 적용됐고, **자기 자신이 태어난 D1 의 인접 층에는 적용되지 않았다.**
- **(b) 유효성 — 균일하지 않다.** 같은 이유로 위증 검사 9건 중 1건이 자기 Given 에서 실행 불가능한 주장이다.

## Defects Found

### N1 — §D 가 §A.6 의 재측정값을 반영하지 않아 문서가 자기모순이다 (iter-1 D4 미해결)

- **위치**: `spec.md:139` (§D "Out of Scope — 오귀속 문구 자체의 수정")
- **Severity: minor — Class: blocking**

§A.6(`spec.md:78`)은 `skill_mirror.go:205` 로 정정하면서 sync-audit F4 의 `:196` 이 오류였음을 명시한다. 그런데 §D(`:139`)는 여전히 ``skill_mirror.go:196`` 을 지목한다. iter-1 D4 는 두 위치(§A.6, §D)를 지목했고 한 곳만 고쳐졌다.

blocking 으로 분류하는 이유는 오탈자여서가 아니다. **HISTORY(`spec.md:23`)가 "`196` → `205`(재측정)" 을 완료로 적고 있고, 같은 문서 `:139` 가 그 주장을 반증한다.** 실측 정확성을 자기 권위의 근거로 삼는 문서에서 이 형태는 관측되지 않은 완료 주장이다(`verification-claim-integrity.md` §1.1 surface 1).

**Required fix**: `spec.md:139` 의 `:196` 을 `:205` 로 정정한다. (SPEC 전체 `grep -n '196'` 3건 중 나머지 2건은 정정 경위 서술이므로 정당하다 — 고칠 것은 `:139` 하나다.)

### N2 — AC-DRW-004 2번 팔이 **명세된 올바른 구현에서 붉어진다** (개정이 들여온 결함)

- **위치**: `acceptance.md:65` (2번 팔), `acceptance.md:12` (매트릭스 행) — 충돌 상대 `spec.md:126`(REQ-DRW-003), `plan.md:49`(M1)
- **Severity: major — Class: blocking**

세 문서를 함께 읽으면 산술이 맞지 않는다.

- REQ-DRW-003(`spec.md:126`): `failed` 는 "**개수**와 그중 **최대 3건의 경고 문구**".
- plan M1(`plan.md:49`): "`MirrorModeFailed` **개수 요약 1줄 + 예시 경고 최대 3건**".
- AC-DRW-004 2번 팔(`acceptance.md:65`): "`MirrorModeFailed` **2개** 실행과 **34개** 실행의 통지 줄 수가 **같다**".

명세대로 구현하면 — N=2: 요약 1줄 + 예시 **2**건 = **3줄**. N=34: 요약 1줄 + 예시 **3**건 = **4줄**. **3 ≠ 4 → 올바른 구현이 이 팔에서 붉어진다.** 상한이 3인데 하한 표본을 3보다 **작게**(2) 잡은 것이 원인이다. 복사 팔은 요약 1줄 고정이라 이 문제가 없어(2 vs 34 모두 1줄), 1번 팔을 그대로 복제한 것이 함정이 됐다.

빠져나갈 구멍은 하나뿐인데 그것도 결함이다: 예시 경고 여러 건을 **한 줄에 이어 붙이면** 양쪽 모두 2줄로 같아진다. 그러면 이 판정은 아무도 명시하지 않은 **출력 서식 결정을 암묵적으로 강제**하는 셈이고, `spec.md:157`("통지 문자열의 정확한 어휘 … 는 run-phase 판단")의 범위 결정과 충돌한다.

우회도 봉쇄돼 있다: `acceptance.md:67` [HARD] 가 상한 형태("34줄보다 작다")를 금지하고, `:69` [HARD] 가 2번 팔 제거를 금지한다. 즉 run-phase 구현자는 **REQ-DRW-003 의 상한을 어기거나, AC-DRW-004 를 붉힌 채로 가거나** 둘 중 하나에 몰린다. 이 계보가 세 번 제거한 「자기 실패에서 통과하는 판정」의 **역상** — 성공에서 실패하는 판정 — 이며, 실행이 아니라 **plan 시점에 산술로 도출된다**.

**Required fix**: 2번 팔의 하한 표본을 상한보다 **크게** 잡는다 — 예 "`MirrorModeFailed` **4개** 실행과 **34개** 실행의 줄 수가 같다"(둘 다 1 + 3 = 4줄). 3건 상한 자체를 판정해야 한다면 **별도 팔**로 분리한다(예: "N=2 에서 문구 2건, N=34 에서 문구 3건" — 이는 줄 수 동일성 팔이 아니라 상한 팔이다). AC-DRW-006 2번 단언(`:94`, "1건 이상 3건 이하")은 이미 상한과 정합하므로 그대로 둔다.

### N3 — §A.4 의 clean-reinstall 두 번째 호출 위치가 off-by-one (개정이 들여온 결함)

- **위치**: `spec.md:65` — "`update.go:627` `Out: out`(호출 **`:625`**, …)"
- **Severity: minor — Class: optional**

실측: `grep -n "runCleanReinstall(planCtx" internal/cli/update.go` → **624**. `:625` 는 옵션 리터럴의 `DryRun: true` 행이다. 같은 문장의 나머지 인용(`:627`, `:362`, `:154`, `:418`, `:425`)은 전부 정확하다.

D1 을 닫으며 새로 쓴 문장이므로 iter-1 에는 없던 값이다. §A.4 는 바로 아래(`:70`)에서 "[HARD] … 스트림 판정은 **주입값 실측**으로 한다" 를 선언하는 자리라, 그 자리의 인용 오차는 문서의 자기 기준에 걸린다.

**Required fix**: `:625` → `:624`.

### N4 — AC-DRW-003 의 2번 위증 검사가 자기 Given 에서 재현 불가능하다 (개정이 들여온 결함)

- **위치**: `acceptance.md:56` (위증 검사) — 관련 Given `acceptance.md:45`, plan M3 `plan.md:64`
- **Severity: major — Class: blocking**

AC-DRW-003 의 위증 검사 두 번째 문장:

> `ErrOut` 을 추가했으나 **호출부에서 주입하지 않아** nil 로 남는 구현 — 1번 단언에서 붉어져야 한다.

여기서 "호출부" 는 `update.go:425`·`:627` 의 `CleanReinstallOptions` 리터럴이다(plan M3 `plan.md:64` 가 그 두 곳에 `ErrOut: cmd.ErrOrStderr()` 주입을 지시한다). 그런데 AC-DRW-003 의 Given(`:45`)은 `CleanReinstallOptions.Out` 에 값을 주입한 상태를 구성한다 — 즉 **테스트가 옵션 구조체를 직접 만들어 `runCleanReinstall` 을 호출**한다. 테스트가 stderr 를 캡처하려면 그 구조체에 `ErrOut` 을 **스스로 주입해야 하므로**, "호출부가 주입을 빠뜨린 상태" 는 이 Given 안에서 **결코 재현되지 않는다**. 그 실패는 `runCleanReinstall` 보다 한 층 위, `runUpdate`/`emitDryRunReinstallPlan` 의 리터럴에 산다.

귀결이 D1 과 같은 형태다. plan M3 은 `ErrOut` nil 을 `os.Stderr` 로 떨어뜨리라고 지시하므로, 주입 누락은 **프로덕션에서 조용히 기본값 분기로 흘러간다** — 그리고 §A.4 [HARD](`spec.md:70`)·AP-9(`plan.md:88`)가 바로 그 "기본값에 기댄 상태" 를 금지한다. 결과: 모든 AC 가 초록인 채로 제품이 문서가 금지한 구성으로 동작할 수 있다. 결함의 **영향**은 D1 보다 가볍지만(nil→`os.Stderr` 라 REQ-DRW-004 자체는 우연히 지켜진다) **형태는 동일**하고, 문서가 세운 규칙에 정면으로 걸린다.

AC-DRW-009 도 이 구멍을 덮지 않는다 — 그 AC 가 지키는 것은 update **template sync** 의 배포기 주입 seam 기본값이지 clean-reinstall 의 `ErrOut` 주입이 아니다(`acceptance.md:124`).

**Required fix** (택1):
- AC-DRW-009 에 팔을 하나 더한다: "`update.go:418`·`:624` 의 `CleanReinstallOptions` 리터럴이 `ErrOut` 을 주입한다" 를 프로덕션 리터럴 자체에 대한 단언으로 세운다(예: `runUpdate` 경로를 구동하거나, 옵션을 만드는 함수를 추출해 그 반환값을 단언).
- 또는 AC-DRW-003 의 2번 위증 검사에서 재현 불가능한 그 문장을 **삭제하고**, `ErrOut` 미주입 위험을 §D.6 전방 점검이 아니라 REQ 수준(REQ-DRW-010 의 확장)으로 올린다.
- 어느 쪽도 아니라면 최소한 그 문장을 "이 AC 로는 검출되지 않는다" 로 **정직하게 낮춘다** — 지금 문장은 검출된다고 주장한다.

### N5 — AC-DRW-009 1번 단언의 검증 수단이 지정되지 않아 약한 구현이 통과할 수 있다

- **위치**: `acceptance.md:128`
- **Severity: minor — Class: optional**

1번 단언은 "배포기가 `template.NewDeployerWithRendererAndForceUpdate` 가 만드는 프로덕션 배포기와 **같은 구성**이다(임베드 FS · renderer · forceUpdate=true)" 다. 세 속성을 지명한 것은 좋으나, `internal/template.deployer` 의 해당 필드는 **전부 비공개**이고 테스트는 `internal/cli` 에 산다. 실제 판정은 `reflect.DeepEqual` 비교가 되기 쉬운데, renderer 가 함수 필드를 가지면 `DeepEqual` 은 항상 거짓이 되어 **판정 불능**이고, 회피해서 `reflect.TypeOf` 만 비교하면 `forceUpdate=false` 나 잘못된 FS 로 만든 배포기도 **통과한다**.

2번 단언(`ResultDeployer` 만족)은 타입 단언 하나로 확정 판정 가능하며 실측상 참이다(`deployer.go:74`). 즉 이 AC 의 절반은 견고하고 절반은 수단 미지정이다.

**Required fix**: 1번 단언에 판정 수단을 한 줄 못박는다 — 예 "seam 기본 생성자가 `template.NewDeployerWithRendererAndForceUpdate` 를 **호출한다**는 것을 그 함수를 감싼 패키지 수준 변수로 관측한다", 또는 "`forceUpdate=true` 는 배포기의 관측 가능한 동작(기존 파일 덮어쓰기)으로 판정한다".

### N6 — plan §E·§H 의 AC 열거가 스테일이라 신설 가드 AC 가 자가검증에서 빠진다 (개정이 들여온 결함)

- **위치**: `plan.md:37` (§E 자가 검증), `plan.md:94` (§H 교차 참조)
- **Severity: minor — Class: blocking**

두 곳 모두 `AC-DRW-001..008` 로 적혀 있다. 판정 기준은 이제 **9건**이다(`acceptance.md:19`, `:137`).

§E 는 run-phase 자가검증 산출물의 정의다. 그 목록을 그대로 따르면 **AC-DRW-009 가 PASS/FAIL 매트릭스에서 누락**된다 — 그리고 AC-DRW-009 는 `acceptance.md:131` 이 스스로 적었듯 "**다른 모든 AC 는 무언가를 주입한 구성에서 판정하므로, 주입하지 않은 프로덕션 구성을 보는 것은 오직 이 AC 뿐**" 인 가드다. iter-1 D5 를 닫으려고 만든 가드가 자가검증에서 빠지면 D5 는 실질적으로 다시 열린다. M3 닫힘 조건(`plan.md:66`)에는 들어 있어 전면 누락은 아니므로 minor 로 두되, class 는 blocking 이다.

**Required fix**: 두 곳을 `AC-DRW-001..009` 로 정정한다.

## 4. 다른 문서의 주장과 충돌하는 것 (별도 요청 항목)

1. **본 SPEC ↔ 본 SPEC — 충돌 (N1).** §A.6 `:205` vs §D `:196`. HISTORY 는 정정 완료를 주장한다.
2. **본 SPEC ↔ 본 SPEC — 충돌 (N2).** REQ-DRW-003 의 3건 상한 vs AC-DRW-004 2번 팔의 N=2/N=34 동일성.
3. **본 SPEC ↔ 본 SPEC — 규율 위반 (N4).** §A.4 [HARD] / AP-9 가 금지한 "기본값에 기댄 상태" 를 `ErrOut` 미주입 경로가 어느 AC 에도 걸리지 않은 채 남긴다.
4. **plan ↔ acceptance — 불일치 (N6).** `AC-DRW-001..008` vs 실재 9건.
5. **본 SPEC ↔ 실제 코드 — 불일치 1건 (N3).** 호출 위치 `:625` vs 실측 `624`. **그 외 §A.2·§A.3·§A.4·§A.6·§A.7 의 코드 인용은 전건 재측정 결과 정확하다**(§2 원장).
6. **본 SPEC ↔ t81 sync-audit — 올바르게 처리했다.** sync-audit F4 의 `:196` 이 오류임을 §A.6 이 명시적으로 뒤집는다. 원본 감사의 결함을 상속하지 않고 정정한 형태이므로 강점이다 — 다만 §D 에서만 상속이 남았다(N1).
7. **선행 SPEC 과의 충돌: 없음.** `REQ-CSC-005`(배포기 내부 직접 출력 금지)를 §D 가 인용·존중, `AC-CSC-006`(반환값 기준 판정)을 §B.D2 가 보존, `AC-CSC-010`(seam 토글 불변식) 회귀 요구를 §E·plan M4·§D.5 가 유지. 참조 SPEC `status: implemented` 로 조정 의무 없음.
8. **iter-1 감사와의 충돌: 없음.** iter-1 이 §A.3 을 전건 검증하고 init 범위 결정을 지지한 판정은 유지된다 — 재론하지 않았고, v0.2.0 도 §B.D1 을 그대로 두었다(`spec.md:24` 가 그 사실을 명시).
9. **`34건`, `ResultDeployer` 컴파일 단언, 선행 SPEC status — 전부 재검증됨** (§2 원장).

## Recommendation

blocking 4건(N1·N2·N4·N6)을 닫아야 한다. 그중 **N2 와 N4 가 실질**이고, N1·N6 은 각 1줄 정정이다.

1. **N2 (최우선)** — AC-DRW-004 2번 팔의 하한 표본을 `2` → `4`(상한 3보다 크게)로 바꾼다. 3건 상한 판정이 필요하면 별도 팔로 분리한다. 이 한 줄이 없으면 run-phase 구현자는 REQ-DRW-003 위반과 AC 붉음 사이에서 선택을 강요당한다.
2. **N4** — `update.go:418`·`:624` 의 옵션 리터럴이 `ErrOut` 을 주입한다는 사실을 보는 팔을 AC-DRW-009 에 추가한다(또는 AC-DRW-003 의 그 위증 검사 문장을 검출 불가로 정직하게 낮춘다). D1 의 형태가 한 층 위에 남아 있다.
3. **N1** — `spec.md:139` 의 `:196` → `:205`.
4. **N6** — `plan.md:37`·`:94` 의 `AC-DRW-001..008` → `..009`.
5. **N3·N5 (optional)** — `:625` → `:624`; AC-DRW-009 1번 단언에 판정 수단 1줄.
6. **유지할 것** — init 범위 결정(§B.D1), D1 정정 전체(§A.4 표·결론·plan M3 재배선·AC-DRW-003 Given 고정), §A.7 의 3-of-3 실측, 9/9 위증 검사 체계. 재작업 대상이 아니다.

**Tier M 반복 상한(2/2) 도달.** 이 감사는 마지막 반복이므로 오케스트레이터는 다음 셋 중 하나를 사용자에게 물어야 한다: (a) **PASS-with-debt** — N1·N3·N6 은 각 1줄, N2 는 숫자 하나(`2` → `4`)이므로 즉시 정정 가능하고, 정정 후 남는 실질 부채는 N4·N5 다. (b) **범위 축소** — 해당 없음(범위는 이미 최소). (c) **상한 연장** — iter-3 로 N2·N4 만 재확인.

정정 규모(숫자 2개 + 행 번호 2개 + AC 팔 1개)에 비해 반복 1회의 비용이 크므로 감사관의 권고는 **(a) 정정 후 PASS-with-debt** 다 — 다만 그 정정이 실제로 들어갔는지는 **읽어서 확인**해야 한다. 이번 iter 의 D4 가 정확히 그 확인 없이 통과할 뻔한 사례다.

## 감사 방법 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

- **Claim**: 위 판정 전건.
- **Evidence**: 읽기 전용 `grep -n` / `sed -n` / `ls | wc -l` / `git log` / `git status`. 명령과 관측값은 §2 원장에 행 단위로 기록.
- **Baseline-attribution**: 전 측정이 `75e7ce381`, clean tree, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t176` 에서 수행. iter-1 보고서의 수치는 **하나도 인용하지 않았다** — 동일 결론에 이른 항목도 이번에 다시 쟀다.
- **Gaps (관측하지 않은 것)**: Go 테스트 미실행(plan-phase, 구현 부재). `go vet`·`golangci-lint` 미실행. 실제 Windows/권한 없는 환경의 폴백 재현 미수행. N2 의 줄 수 산술은 **명세 텍스트로부터의 도출**이지 실행 관측이 아니다(구현이 없어 관측 자체가 불가능하다). N5 의 `reflect.DeepEqual` 취약성은 renderer 의 함수 필드 보유 여부를 실측하지 않은 채 일반 위험으로 적었다.
- **Residual-risk**: N4 의 "테스트가 옵션 구조체를 직접 만든다" 는 AC 문언(`acceptance.md:45`)에서 도출한 것이며, run-phase 가 `runUpdate` 전체를 구동하는 방식으로 AC-DRW-003 을 구현하면 그 위증 검사가 성립할 수 있다 — 그 경우 N4 는 소멸한다. 다만 현재 AC 문언은 그 구동을 요구하지 않으므로 명세 상태로는 결함이다. N2 의 유일한 회피(예시 경고를 한 줄에 잇기)가 run-phase 에서 선택되면 그 팔은 통과하지만, 그때는 명시되지 않은 서식 결정이 판정에 의해 강제된 것이다.
