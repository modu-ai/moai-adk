# SPEC Review Report: SPEC-CODEX-MIRROR-DOCTOR-001

Iteration: 2/2 (Tier M ceiling — 마지막 회차)
Verdict: **PASS**
Overall Score: **0.8875** (Tier M PASS threshold 0.80)

감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t498`, 브랜치 `WT-codex-mirror-doctor`, HEAD `dcb3ba0c7`.
대상: `spec.md` 0.2.0, `acceptance.md` 0.2.0, `plan.md` 0.1.0(불변 주장).
저자 추론 맥락은 M1 Context Isolation 에 따라 무시했다.

이 감사는 **델타 범위**다: iteration 1(`.moai/reports/t498/plan-audit.md`, FAIL 0.775)이 열거한 D1-D10 의 종결 여부와, 수리가 새로 심은 것이 있는지만 본다. 수리가 건드린 표면은 전부 다시 쟀고, 수리 주장은 하나도 그대로 받지 않았다 — 전부 이 트리에서 독립 측정했다.

**PASS 는 무결이 아니다.** 아래 F1(major, blocking)은 구현 착수 **전에** 한 줄 고쳐야 하며, 그대로 두면 AC-CMD-014 는 어떤 구현으로도 통과하지 못한다. 회차가 남지 않았으므로 이 수정은 재감사 대상이 아니라 run-phase 진입 조건으로 넘긴다 — 근거는 § 판정 근거 절에 적었다.

---

## Must-Pass Results (수리가 건드린 표면 전수 재측정)

- **[PASS] MP-1 REQ number consistency** — `grep -oE '^REQ-CMD-[0-9]{3}' spec.md | sort | uniq -c` → REQ-CMD-001…011 각 1회, 결번·중복 0. 수리가 REQ 를 추가·삭제하지 않았다(11건 유지).
- **[PASS] MP-2 GEARS format compliance (요구 계층)** — 수리가 고친 것은 REQ-CMD-006/007 두 건뿐이다. 둘 다 주절이 `Where the project declares Codex wiring, while the mirror directory exists, the check shall …` 로 Where/While 복합 GEARS 를 유지한다(spec.md:74-78, :82-86). 폐기 대상 `If … then` 0건, 규범문 내 `should`/`may` 0건(`grep -nE '\bshould\b|\bmay\b|^If .*then' spec.md` 의 유일한 히트는 spec.md:119 §3 표의 설명 산문 "a project **may** legitimately carry" — 규범 텍스트가 아니다). 판정은 `spec.md` 의 `REQ-XXX` 요구 계층에 대해 내렸고, `acceptance.md` 의 Given-When-Then 은 검증 계층 정상 형식이므로 여기서 감점하지 않았다.
  관측: 두 요구의 **둘째 문장**은 서술법이다("the count is not computed and not reported"). 모달이 아니지만 주절이 GEARS 를 만족하므로 MP-2 는 통과다. 아래 F3 에 optional 로만 적어 둔다.
- **[PASS] MP-3 YAML frontmatter validity** — 정본 12필드 전부 존재(spec.md:2-13) + 선택 `tier: M`. `version: "0.2.0"` 인용 semver, `updated: 2026-09-07`. 거부 대상 snake_case 별칭 0건.
  기계 검증 + **비공허성 대조군**은 아래 § 기계 도구 출력 참조. 대조군은 0.2.0 사본에서 새로 세웠다 — 옛 다이제스트를 재사용하지 않았다.
- **[N/A] MP-4 Section 22 language neutrality** — 단일 언어(Go) 범위. 수리가 이 축을 건드리지 않았다. N/A 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → 자기 참조 1건뿐. 수리가 외부 SPEC 참조를 추가하지 않았다. BLOCKING 0.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` → spec.md `0`, plan.md `0`, acceptance.md `0`. D8-4 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/` → 무출력(EXIT=1). `plan.md` 가 존재하므로 N/A 가 아닌 실측 PASS.

must-pass 방화벽 전부 통과. 이 PASS 는 방화벽 + 루브릭 점수 양쪽에서 나온 것이다.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75-1.0 사이 | 요구 계층의 모호성이 사라졌다: REQ-CMD-006/007 이 배선 전제를 얻어 plan.md:63 과 일치하고(D4 종결), §6(spec.md:174-179)이 §3 row 3 의 조건부 서술과 일치한다(D3 종결). 남은 것은 검증 계층 2건뿐 — AC-CMD-014 절 3 은 층을 잘못 짚었고(F1), AC-CMD-013 둘째 하위 사례는 두 가지로 읽힌다(F2). |
| Completeness | 1.00 | 1.0 | HISTORY(:19, 0.2.0 행 추가), Context(:26), Requirements(:46), 분류 근거(:110), 지시문(:121), Exclusions(:137), Constraints(:168) 전부 존재. `### Out of Scope — <topic>` H3 4개(spec.md:139,146,153,160) 각각 구체 `-` 불릿 보유. frontmatter 12필드 완비. AC 15건 ≤ plan.md §B AC 상한 16. |
| Testability | 0.75 | 0.75 | D1 이 실측으로 닫혔다 — 새 심링크-루프 픽스처를 그대로 옮겨 돌렸더니 PASS, 정리 위험 0(아래 프로브). D2 도 닫혔다 — 파일 수준 swept-count 게이트가 새 테스트명을 지목하는 **모든** AC 를 덮고, 비공허성 측정을 내가 재현했다. weasel word 0건. 남은 결함은 1건 — AC-CMD-014 절 3 은 작성된 그대로면 어떤 구현으로도 통과하지 못한다(F1). 루브릭 0.75("한 AC 가 정확히 이진 판정은 아니나 최소한의 해석으로 측정 가능")에 정확히 앉는다. |
| Traceability | 1.00 | 1.0 | 양방향 기계 검증: AC 표제 15개 ↔ `^REQ: ` 줄 15개(각 표제 바로 아래), REQ-CMD-001…011 전부 최소 1회 인용, 존재하지 않는 REQ 를 가리키는 AC 0건. REQ-CMD-001 은 AC-CMD-014 가, REQ-CMD-011 은 AC-CMD-015 가 덮는다(D6 종결). AC-CMD-011/012 만 `REQ: —` 이며 §6 Constraints 소관임을 스스로 밝힌다 — 고아가 아니라 명시적 무주(無主)다. |

산술: (0.80 + 1.00 + 0.75 + 1.00) / 4 = 3.55 / 4 = **0.8875**. Tier M 임계 0.80 초과.

iteration 1 대비: 0.775 → 0.8875 (**+0.1125**). 점수 회귀 없음 — STOP 조항 미발동.

---

## Regression Check (iteration 1 결함 델타)

| # | 판정 | 근거(이 트리에서 직접 측정) |
|---|---|---|
| D1 | **RESOLVED** | AC-CMD-008(acceptance.md:168-198)이 chmod 를 버리고 이웃 함수의 심링크-루프 관용구로 다시 쓰였다. 세 픽스처 의무가 실제로 실려 있다: Windows skip(:183), `os.Symlink` 실패 시 `t.Skipf`(:184), 호출 **전** 대조 단언(:185-187). 그리고 새 픽스처 자체의 정리 위험을 **프로브로 결판냈다** — AC 문구를 그대로 옮긴 테스트가 PASS 했고, 같은 하니스에 넣은 chmod 대조군은 여전히 FAIL 했다(아래 verbatim). 이웃이 쓴다는 이유로 안전하다고 가정하지 않았다. 덤으로 `--- SKIP:` 은 통과가 아니라는 조항(:197-198)까지 붙어 skip-as-green 구멍도 함께 막혔다. |
| D2 | **RESOLVED** | 파일 수준 게이트(acceptance.md:16-41)가 "아래 모든 `Decided by`"에 바인딩된다. 비공허성 측정을 내가 다시 돌려 **바이트 수준으로 재현**했다(EXIT=0, `[no tests to run]`, `grep -c` → `0`). 게이트 도달 범위도 전수 확인: 새 테스트명을 지목하는 AC 9건 전부에 `swept-count gate applies` 가 붙어 있고(:54,109,122,135,148,165,196,211,224,276,299), 특히 **AC-CMD-010(읽기 전용 경계)**·**AC-CMD-001(un-nagging 회귀)** 둘 다 덮인다 — AC-CMD-001 에는 §2 RED-now 셀(명령·verbatim·exit·tree SHA `dcb3ba0c7`)까지 붙었다. 맨 `-run` 판정자로 남은 AC 는 0건. 기존 테스트만 지목하는 AC(002, 014 후반)는 그 테스트의 실재를 확인했다(`TestDoctorGolden_{Light,Dark,NoColor}` :161/:188/:215, `TestDoctor_CheckCount` :250). |
| D3 | **RESOLVED** | spec.md:174-179 가 "where a copy-mode mirror is the normal case" 단정을 버리고 조건절로 바뀌었다: "Where symlink creation is unavailable (…) a copy-mode mirror is the expected materialization — its frequency is **unmeasured**, and root-cause.md Gaps records the Windows copy-fallback path as unexercised." §3 row 3(spec.md:118)과 표현이 일치하고, 미측정 전제를 사실로 적은 문장이 사라졌다. |
| D4 | **RESOLVED** | REQ-CMD-006(spec.md:74-78)·007(:82-86)이 `Where the project declares Codex wiring,` 전제를 얻었고, 미배선 시 "count is not computed and not reported, whatever the mirror's state" 를 plan.md §D M2 인용과 함께 명시한다. plan.md:63("Call it only in the `wired` branch")과 정면 일치 — 두 문서가 이제 같은 것을 말한다. 미결이던 조합(미배선 + 미러 **존재**)에 판정자도 생겼다: AC-CMD-007 하위 사례 (b)(acceptance.md:155-162)가 유효 심링크·실디렉터리·dangling 을 모두 갖춘 미러를 두고도 아무 출력이 없음을 요구한다. |
| D5 | **RESOLVED** | AC-CMD-013(acceptance.md:251-276)이 신설돼 tail-drop 참여를 판정한다. 주 절의 도달 가능성을 계산으로 확인했다: 기존 stale 요약은 `$CODEX_HOME/config.toml: 1 stale skill entry` = **44 runes**(display 는 상수 — doctor_codex.go:461-463), 미러 요약 최소형 `.agents/skills mirror absent — run moai update --templates-only --force --yes` = **77 runes**. 결합 77+2+44 = **123 > 113** 이므로 픽스처는 실제로 만들 수 있다. 둘째 하위 사례만 흔들린다 → F2. |
| D6 | **RESOLVED** | 양방향 기계 검증 통과(Traceability 행 참조). 추가로 AC-CMD-001 이 기존 `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`(:547, 실재 확인)를 **대체하지 않고 보완**한다는 관계를 명시했다(acceptance.md:56-60) — iteration 1 이 지적한 불명 상태가 해소됐다. |
| D7 | **OPEN (optional, 의도적 보류)** | REQ-CMD-001/011 은 여전히 배치와 심볼명을 요구한다. iteration 1 의 권고대로 손대지 않았다. 판정 불변: 배치를 못박을 정당한 이유(un-nagging 불변식)가 있으므로 optional. |
| D8 | **OPEN (optional, 의도적 보류)** | `mirrorSkillsRelDir` / `mirrorLinkTarget` 는 여전히 package `template` 의 미노출 식별자이고, doctor 는 리터럴을 복제한다. 후속 카드 후보로 남긴다. |
| D9 | **RESOLVED (권고 밖 추가 수리)** | AC-CMD-002 둘째 절이 base SHA 에 고정됐다: `git diff --name-only ace1c5440 -- internal/cli/testdata/`(acceptance.md:92). 그 SHA 가 이 HEAD 의 조상임을 확인했다(`git merge-base --is-ancestor ace1c5440 HEAD` → EXIT=0, `ace1c5440fad4d2a3787b0331059d4f3c4149d15`). |
| D10 | **RESOLVED (권고 밖 추가 수리)** | Definition of Done(acceptance.md:335-338)이 `.moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/` 을 명시하고 "which includes the run-phase `progress.md` §E.2/§E.3 updates — those are expected, not a scope breach" 까지 적는다. |

**정체(stagnation) 없음.** 3회 연속 그대로인 결함 0건. 미해결로 넘어온 것은 iteration 1 이 스스로 optional 로 분류하고 보류를 권고한 D7·D8 둘뿐이며, 이는 미수리가 아니라 합의된 처분이다.

---

## 요청된 두 건의 판단

### ① manager-spec 이 내 D2 처방의 일부를 거절한 것 — **거절이 옳다. 내 처방이 약한 쪽이었다.**

iteration 1 의 D2 Required fix 는 두 형태를 선택지로 제시했다: (a) `--- PASS: <name>` 이 정확히 1건 존재할 것, 또는 (b) `[no tests to run]` 이 **없을** 것. manager-spec 은 (b)를 거부하고 (a)를 구속형으로 삼았다.

거절 근거가 성립한다. `[no tests to run]` 의 **부재**는 "의도한 테스트가 돌았다"를 함의하지 않는다 — 정규식 `-run` 이 다른 테스트를 하나라도 쓸어 담으면 그 토큰은 나타나지 않고, 서브테스트가 필터링된 부모 매치에서도 나타나지 않는다. 즉 (b)는 **엉뚱한 테스트를 쓸고도 초록**이 될 수 있다. 정확한 이름의 `--- PASS:` 를 세는 양성 형태는 다른 테스트로 만족될 수 없으므로 엄격히 강하다.

이것은 단순한 취향 차가 아니다. D2 가 지적한 결함이 바로 "판정 절차가 미존재를 초록과 구분하지 못한다"였는데, 내가 내놓은 선택지 (b)는 **같은 종류의 공허를 다른 문으로 다시 들여놓는 형태**였다. 감사자의 처방이 공허 위험을 품고 있었고, 저자가 그것을 잡아냈다. 기록에 그렇게 남긴다.

(수리본은 (b)를 폐기하지 않고 **추가 필요조건**으로 유지한다 — acceptance.md:23-26 이 `[no tests to run]` 포함 시 무조건 RED 라고 못박는다. 양성 계수를 구속형으로 두고 음성 토큰을 보조로 남긴 이 배치가 두 형태 중 최선이다.)

### ② AC-CMD-013 의 도달 가능성 — **주 절은 도달 가능(측정으로 확인). 둘째 하위 사례만 흔들린다.**

manager-spec 이 "그 문자열을 구성하지 않았다"고 밝힌 부분을 내가 대신 쟀다.

- **주 절(결합 113 초과)**: 도달 가능하다. stale 요약 44 runes + 미러 요약 최소 77 runes + 구분자 2 = 123 > 113. display 가 상수라 stale 쪽 길이는 거의 고정이고, 미러 요약은 AC-CMD-003 이 `.agents/skills` 와 `moai update --templates-only --force --yes` 를 **둘 다** 담도록 요구하므로 하한이 눌려 있다. 게다가 배선 프로젝트에서는 `config.toml missing …` 등 다른 finding 도 얼마든지 함께 세울 수 있어 픽스처 여유가 크다. 이 하위 사례는 지어질 수 있다.
- **둘째 하위 사례(미러 finding 단독으로 상한 초과)**: `checkCodexWiring` 경유로는 **지어질 수 없다**. 그런 상태를 만들려면 구현이 113 runes 를 넘는 미러 요약을 내야 하는데, REQ-CMD-009 와 AC-CMD-009 가 정확히 그것을 금지한다. 같은 패키지이므로 `joinCodexSummaries` 를 합성 finding 으로 직접 호출하면 지어지지만, AC 는 둘 중 어느 경로인지 말하지 않는다. AC 스스로 "should be unreachable … this sub-case exists to make that unreachability *observed*" 라고 적어 구현자에게 경고는 해 두었고, 지목한 함수·줄(`doctor_codex.go:228-248`)이 직접 호출 독법을 강하게 시사한다 — 그래서 **불가능이 아니라 모호**로 판정한다(F2, optional).

---

## Defects Found (structured defect-list — iteration 2 신규분)

F1. **AC-CMD-014 절 3 은 층을 잘못 짚었고, 작성된 그대로면 어떤 구현으로도 통과하지 못한다** — `.moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/acceptance.md:291-292` — "`check.Detail` is empty of mirror text when the same call is made with `verbose=false` — Detail is a `--verbose`-only register". 뒷문장은 **렌더 층에서는 참**이다(`internal/cli/doctor_render.go:134-137` — `if verbose { … if c.Detail != "" { … } }`). 그러나 앞문장이 단언하는 대상은 `DiagnosticCheck.Detail` **필드**이고, 그 필드는 verbose 와 무관하게 채워진다: `checkCodexWiring` 은 `check.Detail = joinCodexDetails(problems, extraDetail)` 을 verbose 게이트 없이 실행한다(`internal/cli/doctor_codex.go:196, 200`). 추론이 아니라 실측이다 — 기존 테스트 `TestCheckCodexWiring_StaleHomeSkillsReported` 가 `checkCodexWiring(…, false)` 로 호출한 뒤 Detail 이 finding 전문을 담고 있음을 단언하며, 이 트리에서 지금 통과한다(`--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)`, doctor_codex_test.go:198·:221). AC 의 Given 은 배선 프로젝트 + 미러 부재 = REQ-CMD-004 의 **finding** 이므로 그 detail 은 `problems` 를 타고 반드시 `check.Detail` 에 들어간다. 절 3 을 만족시키려면 미러 finding 의 detail 만 verbose 로 걸러야 하는데, 그것은 같은 AC 절 2("appears in **both** registers per the file's `codexFinding` convention")와 REQ-CMD-001("existing `codexFinding` two-register shape")을 거스른다. 즉 **한 AC 안에서 절 2 와 절 3 이 서로를 배제한다.** — Severity: **major** — Class: **blocking** — Required fix: 절 3 의 단언 대상을 렌더 층으로 옮긴다 — 예: "`renderDoctorGroups` 를 `verbose=false` 로 렌더한 출력에 미러 detail 문구가 나타나지 않는다(필드는 채워지되 표시되지 않는다 — `doctor_render.go:134-137`)". 또는 절 3 을 삭제한다. **어느 쪽이든 한 줄 편집이며, 절 1·2 와 AC-CMD-014 의 REQ-CMD-001 커버리지는 그대로 유지된다.**

F2. **AC-CMD-013 둘째 하위 사례가 두 경로로 읽히고, AC 는 어느 쪽인지 말하지 않는다** — `acceptance.md:268-273` — 미러 finding 이 유일한 finding이면서 그 요약이 단독으로 상한을 넘는 상태는 `checkCodexWiring` 경유로 지을 수 없다(REQ-CMD-009·AC-CMD-009 가 금지). `joinCodexSummaries` 를 합성 `codexFinding` 으로 직접 호출하면 지어지지만, 그때 판정 대상은 "미러 finding" 이 아니라 조인 함수다. 구현자가 전자를 택하면 AC 는 미충족으로 남고 DoD("AC-CMD-001 through AC-CMD-015 all pass")가 걸린다. AC 가 지목한 함수·줄 인용이 후자를 시사하므로 실무상 막히지는 않을 것이다. — Severity: **minor** — Class: **optional** — Required fix: (선택) 한 절을 덧붙여 경로를 못박는다 — "이 하위 사례는 `joinCodexSummaries` 를 합성 finding 으로 직접 호출해 판정한다(같은 패키지); `checkCodexWiring` 경유로는 AC-CMD-009 때문에 도달 불가하며, 그 도달 불가 자체가 AC-CMD-009 의 단언이다."

F3. **REQ-CMD-006/007 의 둘째 문장이 서술법이라 요구가 아니라 설명으로 읽힌다** — `spec.md:76-78`, `spec.md:85-86` — "the count is not computed and not reported" 는 `shall`/`shall not` 이 아니다. 주절이 GEARS 를 만족하므로 MP-2 는 통과하고, AC-CMD-007 하위 사례 (b)가 이 조항을 실제로 판정하므로 실행 위험도 없다. 순전히 형식 일관성 문제다. — Severity: **minor** — Class: **optional** — Required fix: (선택) "the check **shall not** compute or report the count" 로 바꾼다.

**이월(iteration 1 에서 열린 채 유지)**

D7. REQ-CMD-001/011 이 행동이 아니라 배치·심볼명을 요구 — `spec.md:48-51`, `spec.md:104-108` — Severity: minor — Class: **optional** — 판정 불변, 보류 정당.
D8. doctor 가 `.agents/skills` 경로 리터럴을 복제하며 생산자와의 결속을 못박는 요구가 없음 — `spec.md:66` ↔ `internal/template/skill_mirror.go:52` — Severity: minor — Class: **optional** — 후속 카드 후보.

집계: 총 5건 — blocking **1건**(major 1), optional 4건(전부 minor). critical 0건(must-pass 위반 없음).
iteration 1 대비: blocking 6 → 1, 총 10 → 5.

---

## 기계 도구 출력 (verbatim)

### `moai spec lint` — 대상 SPEC 0.2.0

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md
✓ No findings — all SPEC documents are valid
EXIT=0
```

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md --json
[]
EXIT=0
```

### 위 초록의 비공허성 대조군 (0.2.0 사본에서 새로 수립)

대조군은 불변을 주장하는 장치이므로 iteration 1 의 결과를 재사용하지 않고, 변경된 파일에서 다시 세웠다.

```
$ sed '/^module:/d' .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md > .moai/cache/audit2-probe/mutant.md
$ ~/go/bin/moai spec lint .moai/cache/audit2-probe/mutant.md
SEVERITY  CODE                FILE                                LINE  MESSAGE
--------  ----                ----                                ----  -------
ERROR     FrontmatterInvalid  .moai/cache/audit2-probe/mutant.md  1     Frontmatter required field missing: module

1 error(s), 0 warning(s)
EXIT=1
```

### `mcp__moai__spec_audit` (project_root = 이 워크트리)

```json
{"audited_at":"2026-09-07T04:01:00.285109Z","total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[{"spec_id":"SPEC-CODEX-MIRROR-DOCTOR-001","era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}]}
```

### `mcp__moai__spec_drift` (project_root = 이 워크트리)

전체 출력 93,353자 — 루트 귀속과 sweep 규모를 먼저 확인한 뒤 해당 행만 추출했다.

```
$ jq -c '._root, .total_specs, .modern_clean' <tool-result>
{"dir":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t498","source":"param"}
776
511
```

```json
{"spec_id":"SPEC-CODEX-MIRROR-DOCTOR-001","era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}
```

기계 도구가 낸 유일한 항목은 INFO 수준 `EraAutoDetected` 하나 — 결함이 아니라 `era:` 명시 override 부재 알림이다. drift 0, ERROR 0, WARNING 0. iteration 1 과 동일하며, 수리가 새 lint/drift 항목을 만들지 않았다.

---

## 직접 측정한 검증 (수리 주장을 받지 않고 다시 잰 것)

### D1 — AC-CMD-008 심링크-루프 픽스처의 정리 위험 결판

AC 문구를 그대로 옮긴 프로브(`.moai/cache/audit2-probe/loop_test.go`, 중첩 `go.mod` 로 부모 모듈에서 격리)에 **chmod 대조군을 함께** 넣었다 — 하니스가 정리 위험을 실제로 잡는다는 것을 먼저 보여야 새 픽스처의 PASS 가 의미를 갖기 때문이다.

```
$ go test -C .moai/cache/audit2-probe ./... -count=1 -v
=== RUN   TestACCMD008Fixture
    loop_test.go:37: ReadDir err (AC requires non-nil, non-ENOENT): open /var/folders/…/001/.agents/skills: too many levels of symbolic links
--- PASS: TestACCMD008Fixture (0.00s)
=== RUN   TestChmodZeroWithEntriesControl
    testing.go:1464: TempDir RemoveAll cleanup: openfdat /var/folders/…/001/.agents/skills: permission denied
--- FAIL: TestChmodZeroWithEntriesControl (0.00s)
FAIL
FAIL	probe	0.384s
EXIT=1
```

읽히는 것 세 가지: (1) 새 픽스처는 `t.TempDir()` 정리를 통과한다 — `RemoveAll` 은 심링크를 따라가지 않고 unlink 하므로 루프가 남기는 위험이 없다. (2) AC 가 요구하는 대조 단언이 실제로 성립한다 — `os.ReadDir` 이 ELOOP(`too many levels of symbolic links`)을 내며 이는 non-nil 이고 `fs.ErrNotExist` 가 아니다. (3) 같은 하니스에서 옛 chmod 픽스처는 **여전히 FAIL** 한다 — 이 프로브가 무엇이든 통과시키는 장치가 아님이 확인된다. iteration 1 의 D1 은 실측으로 닫혔다.

### D2 — 비공허성 측정 재현

```
$ go test ./internal/cli/ -run TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.849s [no tests to run]
EXIT=0
$ grep -c -- '--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow' <output>
0
```

acceptance.md:31-38 의 주장과 동일하다(`ok` 시각만 0.690s → 0.849s 로 다르며 이는 측정 노이즈다). tree SHA 귀속도 맞다 — 그 파일이 인용한 `dcb3ba0c7` 이 현재 HEAD 다.

### F1 — Detail 이 verbose=false 에서도 채워진다는 실측

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v
=== RUN   TestCheckCodexWiring_StaleHomeSkillsReported
--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.684s
```

이 테스트는 `checkCodexWiring(wireProjectForDoctor(t), false)`(:198)로 호출한 뒤 `codexDetailText(check)` 가 finding 전문을 담고 있음을 단언한다(:221). 지금 통과한다 → Detail 필드는 verbose=false 에서도 채워진다.

### AC-CMD-015 대조군 실재 확인

```
$ grep -cE 'os\.UserHomeDir\(|exec\.LookPath\(|os\.Getenv\("HOME"\)' internal/cli/doctor.go
4
```

acceptance.md:319-320 이 주장한 `4` 와 일치한다. 이 AC 의 grep 은 대조군을 갖췄고, 그 대조군이 실제로 4를 낸다.

### 인용된 판정자·헬퍼 실재 전수 확인

`TestDoctorGolden_Light`(doctor_golden_test.go:161), `_Dark`(:188), `_NoColor`(:215), `TestDoctor_CheckCount`(:250), `TestCheckCodexWiring_RenderedPanelStaysInBand`(doctor_codex_test.go:437), `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`(:547), `TestCheckCodexWiring_IndeterminateStatNotMissing`(:335), 헬퍼 `wireProjectForDoctor`(:26)·`writeCodexHomeConfig`(:95)·`absentSkillPath`(:132) — 전부 실재. AC 가 기존 자산으로 지목한 것 중 없는 것은 하나도 없다.

### base SHA 고정 확인

```
$ git merge-base --is-ancestor ace1c5440 HEAD ; echo $?
0
$ git rev-parse ace1c5440
ace1c5440fad4d2a3787b0331059d4f3c4149d15
```

---

## 판정 근거 — blocking 1건을 안고 왜 PASS 인가

솔직하게 적는다. F1 을 blocking·major 로 등급한 것과 PASS 판정은 긴장 관계에 있고, 그 긴장을 감추지 않는다.

- **방화벽**: MP-1 ~ MP-7 전부 통과. 미통과 0건. 방화벽발 FAIL 은 없다.
- **점수**: 0.8875 ≥ Tier M 임계 0.80. FAIL 로 뒤집으려면 Clarity+Testability 합이 1.20 미만, 즉 평균 0.60 이어야 하는데 그것은 루브릭 0.50 대역("다수의 AC 가 판단 개입을 요구")이다. AC 15건 중 13건이 깨끗하고 swept-count 게이트가 전 AC 를 덮는 이 문서를 그 대역에 놓는 것은 측정과 어긋난다 — **점수를 깎아 FAIL 을 만들어내는 것**이며, 그것은 M6 이 금지하는 행위다.
- **이월 결함**: iteration 1 의 미해결분 자동 FAIL 조항은 발동하지 않는다. D1-D6 전부 종결이고, 열린 D7·D8 은 iteration 1 이 스스로 optional 로 분류하고 보류를 권고한 항목이다.
- **동일 기준 검증**: "마지막 회차라서 봐준 것인가"를 자문했다. 이것이 iteration 1 이었더라도 0.8875 · 방화벽 전통과 · blocking 1건이면 PASS 였다. iteration 1 의 FAIL 은 blocking 6건에 더해 **측정으로 확정된 깨진 픽스처 1건과 AC 9건에 걸친 계통적 공허-초록 노출**이 있었기 때문이다. 그 둘 다 지금은 없다. 델타는 실재한다.
- **F1 의 실제 대가**: run-phase 에서 AC-CMD-014 를 쓰는 순간 드러나며, 렌더 층으로 단언을 옮기는 한 줄 편집으로 닫힌다. 설계도, 읽기 전용 경계도, 추적성도 건드리지 않는다. 이 한 줄을 위해 운영자 에스컬레이션(PASS-with-debt / 범위 축소 / override 3지선다)을 여는 것은 비용이 편익을 넘는다.

**따라서 PASS 이되, F1 은 run-phase 진입 조건이다.** 감사자 권한으로 명시한다: F1 을 고치지 않은 채 AC-CMD-014 를 판정하면 그 AC 는 통과할 수 없고, 통과했다는 보고는 공허한 초록이다.

---

## Recommendation

착수 전에 다음 **한 건**을 고친다.

1. **F1 — acceptance.md:291-292 절 3 을 렌더 층으로 옮기거나 삭제한다.** 필드가 아니라 `renderDoctorGroups(…, verbose=false)` 출력에 미러 detail 문구가 없음을 단언하도록 바꾼다. 이 SPEC 의 AC 중 유일하게 "작성된 그대로면 통과 불가"인 항목이며, iteration 1 의 D1 과 같은 결함 종류다.

권고이되 강제하지 않는 것(전부 optional, 안 고쳐도 구현에 지장 없음):

2. F2 — AC-CMD-013 둘째 하위 사례의 판정 경로를 한 절로 못박는다(`joinCodexSummaries` 직접 호출).
3. F3 — REQ-CMD-006/007 둘째 문장을 `shall not` 형으로 바꾼다.
4. D8 — `mirrorSkillsRelDir` / `mirrorLinkTarget` 경로 드리프트는 이 SPEC 을 완벽히 구현해도 남는다. 후속 카드로 발행할 것을 권한다.

**F2·F3·D7·D8 을 이유로 재수정 라운드를 돌리지 말 것.** 회차가 남지 않았고, 이 넷은 전부 문구·후속 소관이며, 라운드를 하나 더 돌리는 것 자체가 새 결함을 심는다(수리 라운드마다 새 결함이 들어오는 것은 이 카드에서 이미 한 번 관측됐다 — 0.2.0 이 F1 을 들여왔다).

---

## Gaps (이 감사가 관측하지 **않은** 것)

- **`plan.md` 불변 주장을 커밋 대조로 검증하지 못했다.** SPEC 디렉터리 전체가 미추적(`git status --short` → `?? .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/`)이라 diff 기준이 없다. 대신 iteration 1 이 인용한 좌표에서 내용이 그대로임을 확인했다(plan.md:61-64 의 문구가 iteration 1 인용과 일치, frontmatter `version: "0.1.0"`). 이것은 **정황 일치이지 바이트 동일성 증명이 아니다.**
- `go test ./internal/cli/... -count=1` 전체 베이스라인은 이번에도 **측정하지 않았다**(CLAUDE.local.md §4 부하 규율). 개별 `-run` 3회만 돌렸다. AC-CMD-011 의 세 다리 중 테스트 다리의 착수 시점 도달 가능성은 여전히 **미확정**이며, acceptance.md:237-240 이 이 사실을 스스로 적어 두었다.
- `golangci-lint` / `go vet` / windows cross-build 는 이번 회차에 **다시 재지 않았다.** iteration 1 이 `26b77973e` 에서 잰 값이며, acceptance.md:237 이 그 트리에 귀속시켜 인용한다. 이 회차의 HEAD(`dcb3ba0c7`)는 그 위에 문서 커밋 1건만 얹은 것이라 코드 변경이 없지만, **이 트리에서 다시 재지는 않았다.**
- 미러 검사기가 실제로 낼 요약 문자열은 구현 전이라 알 수 없다. D5 도달 가능성 계산은 SPEC 문구에서 재구성한 **후보** 문자열(최소형 77 runes)로 했으며, 구현된 실제 문자열의 폭은 미측정이다.
- MCP 서버 빌드(`v3.2.0-rc.0 / e79c010b8`)는 이 워크트리 HEAD 보다 앞선 커밋에서 기동돼 있다. `_root.source: param` 으로 읽힌 트리가 이 워크트리임은 확인했으나, 감사 로직 자체는 그 빌드의 것이다.
- 프로브 산출물(`.moai/cache/audit2-probe/`)은 gitignore 대상이라 커밋되지 않으며, 중첩 `go.mod` 로 부모 모듈의 `./...` 에서 격리된다. 증거 경로로 남겨 두었고, 폐기해도 무방하다.

## Residual risk (관측한 것에도 불구하고 여전히 틀릴 수 있는 것)

- **F1 이 조용히 우회될 수 있다.** 구현자가 절 3 을 만족시키려고 미러 detail 만 verbose 로 거르는 길을 택하면 AC 는 초록이 되지만 파일의 균일성이 깨지고 `--json` 소비자가 verbose 에 따라 다른 모양을 보게 된다. 이 보고서가 그 길을 명시적으로 배제했으나, 배제는 문장이지 기계가 아니다.
- **swept-count 게이트는 규약이지 실행기가 아니다.** 구현자가 `-v` 없이 돌리고 exit 0 만 인용하면 게이트는 아무것도 막지 못한다. 이 위험은 감사가 아니라 §E 증거 귀속에서만 잡힌다 — iteration 1 의 잔여 위험이 그대로 남는다.
- D8 의 경로 드리프트는 이 SPEC 을 완벽히 구현해도 남는다. 생산자가 `mirrorSkillsRelDir` 를 바꾸는 순간 doctor 는 조용히 잘못된 "미러 없음" finding 을 내며, 어떤 AC 도 그것을 잡지 못한다.
- `plan.md` 불변이 정황으로만 확인됐으므로, 만약 인용 좌표 **밖**이 수정됐다면 이 감사는 그것을 보지 못했다.
