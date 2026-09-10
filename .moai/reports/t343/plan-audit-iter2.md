# SPEC Review Report: SPEC-RED-NOW-THRESHOLD-001

Iteration: 2/2 (Tier M 상한 = 2 — `.moai/config/sections/harness.yaml:75-78` `plan_audit_tier_ceilings` `M: 2`)
Verdict: **FAIL**
Overall Score: **0.800** (iter1 0.75 대비 **+0.050**)
Tier: **M** (`spec.md:14` `tier: M`)
적용 임계: **0.80** — 출처 `.claude/rules/moai/workflow/spec-workflow.md:138-141` (Tier 표, "plan-auditor PASS threshold" 열) 및 같은 파일 `:329-330`
감사 트리: `WT-red-now-threshold@a6bbbf82b` (`git rev-parse HEAD` → `a6bbbf82b0bc3d46426750520cac7dfe8365c771`)

Reasoning context ignored per M1 Context Isolation. 배차문이 전달한 저자 측 주장(가중 없음·약화 없음·3개 회귀가드 유지)은 근거로 채택하지 않고 전부 이 트리에서 직접 측정했다.

**FAIL의 근거는 점수 축이 아니다.** 집계 0.800은 임계 0.80을 충족한다. FAIL은 **critical blocking 결함 1건(N1)** 에서 나온다 — D1의 구조적 수정(명령을 표 셀 밖 ledger로 이동)이 D4의 수정(L1을 class 무관으로 확대)의 적용 범위를 우회시켜, **이 문서 자신에 L1을 걸면 검사되는 명령이 0개**가 된다. iteration 1의 머리 소견("자기적용이 참이되 공허하다")이 다른 경로로 재발했다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -n "^\*\*REQ-RNT-" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/spec.md` → 14행, `REQ-RNT-001`~`REQ-RNT-014` 연속, 3자리 padding 일관, 결번 0 / 중복 0 (`spec.md:151,158,160,162,164,166,168,170,172,174,176,178,180,182`).
- **[PASS] MP-2 GEARS/EARS format compliance** — **요구 계층(`REQ-XXX`)에 대해서만 판정**. 14개 전부 GEARS: Where(001), Ubiquitous(002·004·007·008·009·010·012·014), Event-driven(005·013), 그리고 003·006·011은 Where/Ubiquitous 합성절. `grep -nE "^\*\*REQ-RNT-[0-9]+\*\* — " spec.md | grep -vE "— (The|Where|When|While)"` → 3건(`:162,168,170`)이 남지만 전부 `<artifact> shall <response>` 형태의 Ubiquitous(일반화 subject)라 감점 없음. `acceptance.md`의 Given-When-Then은 검증 계층이므로 여기서 판정하지 않았다(M3 § Scope). 판정 계층 명시: **requirement layer**.
- **[PASS] MP-3 YAML frontmatter validity** — 정본 12필드 전부 존재(`spec.md:2-13`), `version: "0.2.0"` quoted, `created`/`updated` ISO, `priority: P1`, `lifecycle: spec-anchored`. `grep -c "created_at:\|updated_at:\|labels:\|spec_id:" spec.md` → `0` (exit 1) — 거부 별칭 없음. **D9 반영 확인**: `module: internal/spec`(`spec.md:11`)로 path-like 요건 충족.
- **[PASS] MP-4 language neutrality** — 배포 트리 반입분은 `plan-auditor.md` 미러의 MP-8 절뿐이며, Go 테스트와 fixture는 REQ-RNT-012가 배포 트리 밖으로 못박는다. 미러 반입 절에 언어별 도구명 없음. `internal/spec/`·`make build`는 전부 비배포 범위.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -oE "SPEC-([A-Z][A-Z0-9]+-)+[0-9]+" spec.md | sort -u` → 4건. `SPEC-TODO-SQLITE-001`: `grep "^status:" .moai/specs/SPEC-TODO-SQLITE-001/spec.md` → `status: completed` (retired/superseded/archived 아님 → D7-4 미발화). `SPEC-TODO-LANDING-STATE-001`: `ls .moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md` → `No such file or directory` (exit 1) → D7-5 SHOULD이나 `spec.md:92-103`이 "unlanded, `.claude/worktrees/t331/`에만 존재, 파일 의존 없음"을 명시 → 해소. `SPEC-AUTH-001`은 미러의 illustrative placeholder 인용(`spec.md:176`) → 해소. **BLOCKING 없음.**
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0` (exit 1). D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-RED-NOW-THRESHOLD-001/` → 매치 없음(exit 1). `plan.md` §F M1의 열린 결정 2건은 (a)/(b) 모두 SETTLED로 닫혔다(`plan.md:96-124`) → **D10 해소**.

`mcp__moai__spec_audit`(project_root = 이 워크트리) → drift finding 1건, `EraAutoDetected` **INFO**뿐. 차단 없음.

---

## MP-8 자기적용 (이 SPEC이 제안하는 기준을 이 SPEC에 그대로 적용)

**iteration 1의 결과를 재사용하지 않았다.** ledger가 신설되고 명령이 전부 재작성되었으므로 11개 전부 이 트리에서 원문 그대로 재실행했다. 실행 형태는 `<명령>; echo "exit: $?"` — 명령 자체에는 파이프·`&&`·`;`·서브셸이 없다(원문 바이트 확인: `sed -n '30,78p' acceptance.md | grep '^E-'`, 백슬래시 0개).

### L2 재실행 — ledger E-01..E-11 (11/11 재현)

| id | 재실행한 명령 | 관측 stdout | exit | ledger 주장과 |
|----|---------------|-------------|------|----------------|
| E-01 | `grep -c "RED-now cell content" .claude/rules/moai/development/verification-completeness.md` | `0` | 1 | 일치 |
| E-02 | `grep -c -e tense -e mood -e counterfactual -e "future.sense" …verification-completeness.md …plan-auditor.md` | `…verification-completeness.md:0` / `…plan-auditor.md:0` | 1 | 일치 |
| E-03 | `grep -c "regression-guard" …verification-completeness.md` | `0` | 1 | 일치 |
| E-04 | `grep -c "MP-8" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| E-05 | `ls internal/spec/red_now_cell_test.go` | (stdout 공백) / stderr `ls: …: No such file or directory` | 1 | 일치 |
| E-06 | `grep -c "AC-6:" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| E-07 | `grep -c "MOAI-REDNOW-BEGIN" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| E-08 | `ls internal/spec/testdata/red_now/` | (stdout 공백) / stderr `ls: …: No such file or directory` | 1 | 일치 |
| E-09 | `grep -c "MOAI-REDNOW-BEGIN" internal/template/templates/.claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| E-10 | `grep -nE "SPEC-[A-Z]+-[0-9]{3}" internal/template/templates/.claude/agents/moai/plan-auditor.md` | `446:…SPEC-AUTH-001/…` / `464:…SPEC-AUTH-001/…` | 0 | 일치 |
| E-11 | `grep -rl "red_now" internal/template/templates/` | (공백) | 1 | 일치 |

### release-blocking 12건의 인용 RED 매핑 (12/12 재현)

AC-001→E-01 · AC-003→E-03 · AC-004→E-04 · AC-005→E-05 · AC-006→E-04 · AC-007→E-06 · AC-008→E-07 · AC-009a→E-08 · AC-009b→E-08 · AC-010→E-09 · AC-013→E-05 · AC-014→E-05. **인용된 12건 모두 non-zero exit의 RED이며 전부 재현**된다. exit 0인 E-10·E-11을 RED로 인용한 release-blocking 셀은 없다(확인함).

### §D.0.1 divergence probe 재현

scratch fixture 2개(`fixA`= 판별어 `mood` 심음, `fixB`= 없음)에 대해 직접 실행:

```
$ grep -cE "tense\|mood\|counterfactual\|future.sense" <fixA>   → 0    exit 1   ← 구 형태: 거짓 green
$ grep -c -e tense -e mood -e counterfactual -e "future.sense" <fixA> → 1  exit 0   ← E-02 형태: 참 red
$ grep -c -e tense -e mood -e counterfactual -e "future.sense" <fixB> → 0  exit 1   ← E-02 형태: 참 green
```

§D.0.1(`acceptance.md:89-97`)이 적은 것과 **정확히 일치**한다. E-02의 비공허성은 저자 주장이 아니라 재현된 관측이다.

### D1 구조적 수정의 격리 검증

- `grep -nF 'grep' acceptance.md` → 구 파손 형태(`grep -cE "tense\|mood…"`)는 **85·91행에만** 존재하며 두 행 모두 §D.0.1(divergence 기록) 안이다. 셀(106-120행)은 어느 것도 인용하지 않는다 → **의도된 문서화이며 잔존 결함 아님. 격리 확인.**
- `grep -n 'wc -l' acceptance.md` → 매치 없음(exit 1). D2의 잔재 없음.
- 원문 바이트 기준 11개 명령 전부 이스케이프 없이 그대로 실행됨 → **GFM 셀 위험은 재배치가 아니라 제거되었다**(단, 아래 N1 참조 — 제거의 부작용이 L1의 사각을 만들었다).

---

## D1..D10 폐쇄 표

| # | 판정 | 판정을 결정한 명령/판독 |
|---|------|--------------------------|
| **D1** 이스케이프 파이프 공허 명령 | **closed** | `grep -nF 'grep' acceptance.md` → 구 형태는 §D.0.1(85·91행)에만; 셀 인용 0. E-02 원문 재실행 → `…:0`/`…:0` exit 1. divergence probe 3케이스 재현 |
| **D2** 실행 불가 `\| wc -l` | **closed** | `grep -n 'wc -l' acceptance.md` → exit 1. E-11 원문 재실행 → 공백 exit 1. `spec.md:154` verbatim = raw file bytes 정의 확인 |
| **D3** AC-RNT-011의 부정직한 "being preserved" | **closed** | `acceptance.md:117` — "green today **only because the MP-8 clause does not exist yet** (E-09 → `0`)" — AC-RNT-012와 동일 형태로 정렬 |
| **D4** 강등이 명령 검증의 탈출구 | **partially closed** | `spec.md:170` REQ-RNT-008이 class 무관으로 확대된 것은 사실(M-4 차단). **그러나 범위 술어가 "any cell"이라 fenced ledger에 닿지 않는다** → N1 참조 |
| **D5** token-insertion mutant 5건 | **partially closed** | `acceptance.md:106,108,109,111,112` — 5건 전부 span-scoped 명명 subtest로 이관됨(파일 전역 grep 소멸 확인). **그러나 span 내부 삽입으로 여전히 만족 가능** → N3 |
| **D6** MP-8 신뢰 경계 미정의 | **closed** | `spec.md:153`(read-only single invocation 정의) + `spec.md:180` REQ-RNT-013(거부·비-pass·리포 규율 우선) + `spec.md:129-137` 노출 반전 문단. 잔여 위험 1건은 N5(저자도 미종결로 선언) |
| **D7** continued-firing axis 2 부재 | **partially closed** | `spec.md:182` REQ-RNT-014 + `acceptance.md:186-216` 2축 분리. MP-8 축은 실제로 닫힘. Go 테스트 자신의 비실행은 미커버로 선언 → 판정은 N4 |
| **D8** 트리 핀 없는 줄번호 | **closed** | `grep -n "446\|464" spec.md` 근거: `spec.md:176`에 줄번호 없음; 위치는 `acceptance.md:117` + ledger E-10(문서 핀 `a6bbbf82b` 상속)으로 이동 |
| **D9** `module` non-path-like | **closed** | `spec.md:11` → `module: internal/spec` |
| **D10** 마커 없는 열린 결정 | **closed** | `plan.md:96-124` (a)/(b) 모두 SETTLED. `grep -rn '\[NEEDS CLARIFICATION' <spec dir>` → exit 1 |

**저자가 닫혔다고 주장했으나 내가 재현하지 못한 항목: 없음.** 거짓 완료 주장은 관측되지 않았다. D4·D5·D7의 "partially"는 저자 주장의 허위가 아니라 **수정이 닿지 못한 잔여 표면**이다(D7은 저자 스스로 미커버를 명시했다).

---

## Monotonicity 판정

**측정 한계 먼저.** SPEC 디렉터리는 미추적(`git status --short` → `?? .moai/specs/SPEC-RED-NOW-THRESHOLD-001/`)이라 iteration-1 아티팩트의 바이트가 남아 있지 않다. 따라서 전문 대조는 불가능하며, 대조 기준은 iter1 보고서가 인용한 발췌·표·계수(`plan-audit-iter1.md`)다. 이 한계는 Gaps에 기록한다.

측정 결과:

| 항목 | iter1 | iter2 (이 트리에서 측정) | 판정 |
|---|---|---|---|
| REQ 수 | 12 | `grep -c "^\*\*REQ-RNT-" spec.md` → **14** | +2 (013·014 신설) |
| AC 수 | 13 | `grep -c "^| \*\*AC-RNT-" acceptance.md` → **15** | +2 (013·014 신설) |
| release-blocking | 10 | 표 행 직접 판독 → **12** (`:106,108,109,110,111,112,113,114,115,116,119,120`) | +2, 기존 10건 전원 유지 |
| regression-guard | 3 | `grep -c "| \*\*regression-guard\*\*"` → **3** (`:107` AC-002, `:117` AC-011, `:118` AC-012) | **동일 3건, 승격 0** |

- **삭제된 criterion 없음**: iter1의 AC-001·003·004·005·006·007·008·009a·009b·010·002·011·012 전 13건이 그대로 존재한다(행 번호 위와 같음).
- **강등/승격 없음**: regression-guard 3건은 iter1과 동일한 AC-002·011·012이며, 게이트를 부풀리기 위한 승격도 없다.
- **약화된 criterion 없음**: iter1 D5 표가 인용한 5건의 green path는 `grep -c <token> <file> ≥1` + 판정 불가 산문이었다. 현재는 파일 전역 grep이 사라지고(§D.2 `acceptance.md:170` 및 셀 본문 확인) span 한정 명명 subtest로 대체됐다 — **범위가 좁아졌으므로 강화**다. AC-002의 green path는 파손 명령에서 재현 가능한 E-02로 교체(강화), AC-011은 문구만 정직화(실질 동일), AC-012는 파이프 제거(강화).
- **회귀 가드 3건 유지 확인**: 위 표.

**판정: 점수 상승은 정당하다.** 기준을 제거하거나 무르게 해서 얻은 상승이 아니며, 상승분은 (a) 명령 11개가 실제로 재현 가능해진 것, (b) green path 5건의 범위 축소, (c) REQ 2건 신설로 D6·D7의 공백을 요구 계층에 올린 것에서 나온다. **단조성 위반 없음.**

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | "the command"·"verbatim"·문서 핀 상속이 전부 정의됨(`spec.md:153-156`), 줄번호 제거(`:176`). **감점**: `spec.md:158` REQ-RNT-002가 요소를 **3개**(command / output block / pinned SHA)로 열거해 `spec.md:151` REQ-RNT-001의 **4개**와 정면 충돌(N2); `plan.md:87`이 여전히 "three-element RED-cell obligation"(N2); `acceptance.md:216`의 Out-of-Scope 상호참조가 실제로 그 배제를 담지 않는 절을 가리킴(N6) |
| Completeness | 0.90 | 1.0 (감점 적용) | HISTORY·WHY·WHAT·HOW·REQUIREMENTS·ACCEPTANCE·Out of Scope 전부 존재; `### Out of Scope —` H3 **5개**(`spec.md:203,208,214,221,226`) 각각 `-` 불릿 보유; frontmatter 12필드 완비. §D.3이 iter1에 없던 axis 2를 답한다(`acceptance.md:202-216`). **감점**: 저자가 근거로 삼는 배제("repository-wide test-inventory guard")가 `spec.md` Out of Scope 어디에도 없음(N6) |
| Testability | 0.65 | 0.50~0.75 사이, 0.75에서 하향 | ledger 11개 전부 원문 재현(위 표), weasel word 0(`grep -niE "appropriate\|adequate\|reasonable\|proper" acceptance.md` → 1건은 "property" 오탐), 15개 전부 Given-When-Then. **하향 근거**: L1의 범위 술어가 이 문서의 명령 전부를 사각에 둠(N1, critical); M-3이 span 내부에서 생존(N3); `acceptance.md:108,111`이 명령이 판정 못 하는 산문 접속절을 여전히 보유; `acceptance.md:142`의 인용 명령 출력이 재현되지 않음(N7) |
| Traceability | 0.90 | 1.0 (감점 적용) | `acceptance.md:124-139` 매핑표 14 REQ ↔ 15 AC, 고아 AC 0, 미커버 REQ 0, REQ-RNT-009만 2건(009a·009b)으로 분기. Tier M 16/16 예산 이내. **감점**: 같은 절(`:142`)의 계수 인용 명령이 자기 참조로 13을 내는데 12로 적힘(N7) |

**Aggregate = (0.75 + 0.90 + 0.65 + 0.90) / 4 = 0.800** — Tier M 임계 0.80을 **충족**(iter1 0.75 → **+0.050**).

**그럼에도 Verdict는 FAIL이다.** 점수 축은 통과하고, MP-1~MP-7도 전부 통과한다. FAIL은 아래 N1(critical, blocking) 단독으로 성립한다.

---

## Defects Found (이번 개정이 도입/잔존시킨 것)

**N1. L1의 범위 술어("any cell")가 이 SPEC 자신의 명령 11개 전부를 사각에 둔다 — 자기적용이 다시 공허하다** — `spec.md:170` (REQ-RNT-008), `spec.md:113` (§B L1 행), `acceptance.md:19-28` (§D.0), `acceptance.md:231-233` (§D.5) — Severity: **critical** — Class: **blocking**

D1의 수정은 "**표 셀 안에 명령을 두지 않는다**"였다(`acceptance.md:23-24`: "no command appears in a table cell at all. Cells cite a ledger id"). D4의 수정은 "**셀에 명령이 있으면** class 무관으로 형식을 검사한다"였다(`spec.md:170`: "Where any **cell** of an audited `acceptance.md` carries a command …"). 두 수정은 서로를 무효화한다.

측정:

```
$ grep -n "ledger" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/spec.md
(매치 없음 — exit 1)
```

요구 계층에는 ledger라는 개념 자체가 없다. `spec.md:113`의 L1 정의도 "Does a **cell** carry a command …"로 셀 한정이다. 따라서 REQ-RNT-008의 class-무관 L1을 이 문서에 걸면 **검사 대상 명령은 0개**다 — 명령이 전부 fenced ledger 안에 있기 때문이다.

그 결과 `acceptance.md:231-233`(§D.5)의 문장 — "L1 checks every command regardless of class, so **this document's own regression-guard cells are inside the gate** rather than outside it — the iteration-1 hole that made the previous self-application claim true but empty" — 는 **이 개정본에서 거짓**이다. 그 셀들은 이제 명령을 담고 있지 않으므로 게이트 안에 있는 것이 아니라 게이트가 볼 것이 없다.

이는 mutant M-4(class laundering)의 사촌이다. 저자가 M-4를 닫은 방식(class 축 제거)은 옳았으나, 같은 회피가 **구조 축**으로 열려 있다: 명령을 셀 밖으로 옮기면 L1을 벗어난다. 그리고 이 SPEC 자신이 그 회피를 §D.0에서 **권장 패턴으로 제도화**한다.

**Required fix**: REQ-RNT-008(및 `spec.md:113` L1 행)의 범위 술어를 "cell"에서 "`acceptance.md`가 인용하는 모든 명령 — 셀 안이든 fenced evidence ledger 안이든"으로 바꾼다. 한 절의 단어 교체이며 새 기구가 필요 없다. 동시에 §D.5의 자기적용 문장을 새 범위에 맞춰 다시 쓴다.

---

**N2. 요소 개수가 4와 3으로 갈린다 — 규범 텍스트 2곳** — `spec.md:158` (REQ-RNT-002), `plan.md:87` (M1 본문) — Severity: **major** — Class: **blocking**

`spec.md:151` REQ-RNT-001은 **4요소**(command / verbatim stdout / exit code / tree SHA)를 요구한다. 그러나:

- `spec.md:158` REQ-RNT-002: "a structural test over cell content (presence of **a command, an output block, and a pinned SHA**)" — **3요소**, exit code 누락. L1 구조 테스트를 REQ-RNT-002에서 읽는 구현자는 exit code를 검사하지 않는 3요소 체커를 만든다.
- `plan.md:87` M1: "add the **three-element** RED-cell obligation" — 그 절을 실제로 착지시키는 마일스톤 본문이 3요소로 지시한다. (같은 파일 `:106-109`는 4요소로 정확히 적혀 있어 파일 내부에서도 갈린다.)

배차문이 지목한 "3/4 불일치는 Clarity 결함"에 정확히 해당하며, 두 곳 모두 산문 해설이 아니라 **지시 텍스트**다.

**Required fix**: `spec.md:158`의 괄호 열거에 exit code를 추가하고, `plan.md:87`을 "four-element"로 고친다.

---

**N3. M-3(token insertion)은 span 내부에서 생존한다 — 그리고 `acceptance.md:106`이 가지지 않은 면역을 주장한다** — `acceptance.md:106,108,109,111` — Severity: **major** — Class: **blocking**

D5의 수정은 실질적 개선이다(파일 전역 grep 소멸, 범위가 482행 → 절 단위로 축소). 그러나 술어의 **종류**는 바뀌지 않았다: 여전히 토큰 존재 검사이고, 무대만 좁아졌다. `verification-completeness.md` §2 span은 실측 41행(`:120-160`)이며, 그 안에 주석이나 무관한 문장을 한 줄 넣으면 토큰 검사는 통과한다.

그런데 `acceptance.md:106` AC-RNT-001의 green path는 이렇게 적는다:

> "**A token pasted into a comment does not satisfy it** — the assertion runs inside the extracted §2 span, not over the file."

**주석이 span 안에 있으면 만족된다.** 문장은 참이 아니라 조건부로만 참이며, 조건을 적지 않았다. 이 SPEC은 정확히 이런 형태의 과잉 주장을 잡으려고 존재한다.

추가로 두 셀은 아직 명령/테스트가 판정하지 못하는 산문 접속절을 남겼다:
- `acceptance.md:108`: "and that the demotion is stated **as the disposition rather than as an option**" — 어떤 술어가 이것을 판정하는지 명시 없음.
- `acceptance.md:111`: "both an `N/A` token and **a stated-reason obligation**" — 후자는 토큰이 아니다.

**Required fix**: `:106`의 면역 주장을 "span 밖 삽입은 만족시키지 못한다"로 정확히 축소하고, `:108`·`:111`의 두 접속절을 판정 가능한 토큰/구조 술어로 바꾸거나(예: 부정어 + `pass`의 인접 패턴) 잔여 한계로 명시한다.

---

**N4. §D.3 axis 2의 "정직한 공백"은 정당한 경계다 — 단, 그 경계를 뒷받침한다고 인용한 절이 존재하지 않는다** — `acceptance.md:211-216` — Severity: **minor** — Class: **blocking** (판정은 아래)

**판정: 면책 문구를 두른 §1.3 결함이 아니라, 정당한 도구 경계다.** 근거 셋:

1. `verification-completeness.md:105-107`의 §1.3 시험("내일 이게 멈추면 내가 보는 것 중 무엇이 달라지는가")을 **MP-8**에 걸면 이번 개정은 실제로 답을 만들었다 — REQ-RNT-014가 Go 테스트에 MP-8 span + 보고 템플릿 행을 둘 다 단언시키고, AC-RNT-014가 그 실패를 **관측**하도록 요구한다. iter1에서 비어 있던 축이 닫혔다.
2. 남은 축(Go 테스트 자신의 비실행)을 닫는 기구는 리포지토리 전역 테스트 인벤토리 가드이며, 이는 이 SPEC이 재사용하는 span-extraction seam으로는 만들 수 없는 **다른 종류의 기구**다. Enforce Simplicity 관점에서 한 SPEC이 짊어질 범위가 아니다.
3. §1.3이 금지하는 것은 "비실행이 성공과 구별되지 않는 검사"를 **답한 것처럼 두는 것**이다. 여기서는 구별 불가능성을 명시적으로 선언했으므로, 독자가 속지 않는다. 면책이 결함을 덮은 것이 아니라 결함의 위치를 공표했다.

**그러나 결함이 하나 붙어 있다.** `acceptance.md:216`은 그 배제의 근거로 "spec.md § Out of Scope — adjacent cards"를 인용하는데, 측정 결과 그 절(`spec.md:221-224`)은 t341·t344·t345만 담고 있고 test-inventory guard는 언급하지 않는다. 즉 배제가 요구 계층에 존재하지 않으며 인용은 끊어져 있다.

**Required fix**: `spec.md` §D에 `### Out of Scope — repository-wide test-inventory guard` H3를 신설하고 한 줄 불릿으로 배제를 명시한 뒤, `acceptance.md:216`이 그것을 가리키게 한다. (N6과 같은 수정으로 함께 닫힌다.)

---

**N5. 형식을 지키면서도 임의로 비싼 명령이 가능하다 — 이 SPEC 안에서 닫아야 한다** — `spec.md:153` (REQ-RNT-001 form), `spec.md:180` (REQ-RNT-013) — Severity: **minor** — Class: **blocking**

저자가 잔여 위험으로 공표한 항목이다. 판정을 요청받았으므로 명시한다: **후속 카드가 아니라 이 SPEC 소관이다.**

근거는 비용이다. REQ-RNT-013은 이미 (a) 실행 실패, (b) 형식 미충족 거부, (c) 금지된 연산이라는 **처분 열거 구조**를 갖고 있고, 세 경우 모두 처분이 동일하다(실행 안 함 / pass 아님 / REQ-RNT-003 강등). "정해진 시간 안에 끝나지 않는 경우"를 네 번째 항목으로 추가하는 것은 **열거에 한 구를 더하는 일**이지 새 기구를 세우는 일이 아니다. 반면 후속 카드로 미루면, 그 사이 MP-8은 저자가 제어하는 문자열로 감사를 무기한 정지시킬 수 있는 경로를 열어둔 채 착지한다 — 그리고 그 정지는 실패 신호를 내지 않으므로 §1.3형 결함이 된다.

**Required fix**: REQ-RNT-013의 조건 열거에 "또는 정해진 실행 상한 안에 완료되지 않는" 한 구를 추가한다.

---

**N6. `spec.md` Out of Scope에 test-inventory guard 배제가 없다** — `spec.md:221-224` vs `acceptance.md:216` — Severity: **minor** — Class: **blocking** — (N4와 동일 수정으로 닫힘)

`grep -n "test-inventory" spec.md acceptance.md` → `acceptance.md:215`만 매치. 요구 계층에는 그 배제가 존재하지 않는다.

---

**N7. §D.1의 계수 인용 명령이 자기 참조로 13을 내는데 12로 적혀 있다** — `acceptance.md:142` — Severity: **major** — Class: **blocking**

측정:

```
$ grep -c "| release-blocking |" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md
13
exit: 0
```

문서는 "`grep -c "| release-blocking |"` → 12"라고 적는다. 13이 나오는 이유는 **142행 자신**이 그 패턴을 이스케이프 없이 담고 있어 스스로에게 매치되기 때문이다(`grep -n` 으로 매치 13행 확인: 표 행 12개 + `:142`). 형제 인용 두 개는 정확하다 — `grep -c "^| \*\*AC-RNT-"` → **15** (일치), `grep -c "| \*\*regression-guard\*\*"` → **3** (일치, 이쪽은 패턴에 `\*` 이스케이프가 있어 자기 매치가 일어나지 않는다). 즉 **세 인용 중 하나만 자기 참조에 걸린 비대칭**이다.

사실로서의 계수(release-blocking 12건)는 옳다 — 표 행을 직접 판독해 확인했다. 틀린 것은 **인용된 명령의 관측 출력**이다. 이 SPEC의 중심 주장이 "인용된 명령은 적힌 출력을 재현해야 한다"이므로, 자기 문서의 계수 절에서 재현 실패가 나온 것은 iteration 1의 D1/D2와 같은 계열의 결함이다. 저자가 스스로 AC 계수 오기(2→3)를 측정으로 잡아낸 것과 대조된다 — 측정 습관은 작동했으나 이 한 줄에는 닿지 않았다.

**Required fix**: 패턴을 자기 매치되지 않는 형태로 바꾸거나(예: `grep -c "^| \*\*AC-RNT-[0-9]*\*\* | release-blocking"`) 관측 출력을 `13`으로 정정하고 그중 1건이 인용 행 자신임을 적는다. 그리고 세 인용을 ledger로 옮겨 §D.0의 규율 아래 둔다.

---

**N8. `acceptance.md:116`이 §D.0의 "표 셀에 명령을 두지 않는다"를 스스로 어긴다(경미)** — `acceptance.md:116` — Severity: **minor** — Class: **optional**

AC-RNT-010의 green path 셀은 "`make build` exits 0"을 담는다. 형식(read-only는 아니나 단일 호출, 파이프 없음)에는 문제가 없고 green path이므로 RED 의무 밖이지만, §D.0(`:23-24`)의 절대 문장 "no command appears in a table cell at all"은 이 셀 때문에 참이 아니다. 진술을 "인용된 RED 명령은 표 셀에 두지 않는다"로 좁히면 해소된다.

---

**N9. 세 criterion(AC-005·013·014)이 하나의 파일 부재(E-05)를 RED으로 공유한다** — `acceptance.md:110,119,120` — Severity: **minor** — Class: **optional**

`ls internal/spec/red_now_cell_test.go` 하나의 부재가 세 criterion의 RED을 동시에 만든다. M3가 파일을 만드는 순간 세 RED이 함께 뒤집히며, 그 파일이 execution-discipline subtest나 liveness subtest를 실제로 담고 있는지는 RED 축이 판정하지 않는다(green path의 subtest 이름이 그 몫을 진다). rule §2의 "RED은 옳은 이유로 붉어야 한다"에 대해 다소 성긴 형태다. green path가 보완하므로 optional로 둔다.

---

## Recommendation

**Verdict: FAIL** — 단, iteration 1과 성질이 다르다. 점수 축(0.800 ≥ 0.80)과 must-pass 축(MP-1~MP-7 전부 PASS)은 **둘 다 통과**한다. FAIL은 **N1 단독**으로 성립한다: L1의 범위 술어가 이 문서의 명령 전부를 사각에 두어, `acceptance.md` §D.5가 명시적으로 주장하는 자기적용 문장이 거짓이 된다. 이 SPEC의 존재 이유가 "선언된 검사가 실제로 무언가를 검사하는가"이므로, 그 검사가 자기 문서에서 0건을 검사하는 상태로 착지시키는 것은 부채로 이월할 성질이 아니다.

### 수정 순서 (전부 단어/절 수준, 새 기구 0개)

1. **N1** — REQ-RNT-008 + `spec.md:113` L1 행의 범위를 "cell" → "`acceptance.md`가 인용하는 모든 명령(셀·fenced ledger 포함)"으로. §D.5 문장 동반 수정. (critical)
2. **N2** — `spec.md:158`에 exit code 추가, `plan.md:87` "three-element" → "four-element".
3. **N7** — `acceptance.md:142` 계수 인용 3건을 ledger로 이관하고 자기 매치 없는 패턴으로 재작성.
4. **N3** — `:106`의 면역 주장 축소, `:108`·`:111`의 산문 접속절 술어화.
5. **N5** — REQ-RNT-013에 실행 상한 한 구 추가.
6. **N4·N6** — `spec.md` §D에 test-inventory guard 배제 H3 신설, `acceptance.md:216` 참조 정정.
7. **N8·N9** — 문구 정정(optional).

### 상한 도달 — 운영자 에스컬레이션 필요

**Tier M 상한은 2회이며 이번이 iteration 2다**(`harness.yaml:77` `M: 2`). 추가 반복은 상한 밖이므로, 리드는 다음 3택을 운영자에게 제시해야 한다:

1. **PASS-with-debt** — 집계 0.800이 임계를 충족하고 MP-1~MP-7이 전부 통과하므로 방어 가능한 선택이다. **다만 N1은 부채로 이월하지 말 것을 권고한다** — 단어 교체 한 건이고, 이월하면 이 SPEC이 금지하려는 공허한 검사가 이 SPEC 자신으로 착지한다. N2·N7까지 3건만 고치고 나머지(N3~N9)를 run-phase 부채로 넘기는 절충이 비용 대비 가장 낫다.
2. **범위 축소** — 권고하지 않는다. 결함이 범위 과대에서 오지 않았다.
3. **운영자 override로 iteration 3 연장** — N1~N7이 전부 단어/절 수준이라 1회 개정으로 닫힐 가능성이 높다.

### 이번 개정에서 공격했으나 버틴 것 (반증 시도 기록)

- ledger 11개 명령 전부 원문 바이트에서 그대로 실행되며 stdout·exit가 ledger 기재와 일치한다 — **재배치가 아니라 실제 제거**였다.
- §D.0.1의 divergence probe는 3케이스 전부 내 fixture에서 재현됐다 — E-02의 비공허성은 주장이 아니라 관측이다.
- 구 파손 형태는 §D.0.1(85·91행)에만 존재하고 어떤 셀도 인용하지 않는다 — 격리 성립.
- regression-guard 3건이 그대로 유지되고 승격이 0건이다 — 게이트 부풀리기 없음.
- release-blocking 12건이 인용한 RED은 전부 non-zero exit이고 exit 0인 ledger 항목(E-10·E-11)을 RED으로 인용한 셀은 없다.
- MP-1~MP-7 전부 통과하며, `mcp__moai__spec_audit`은 INFO 1건만 낸다.

---

## Gaps — 이 감사에서 관측하지 **않은** 것

- **iteration-1 아티팩트의 바이트 대조를 하지 못했다.** SPEC 디렉터리가 미추적이라 이전 판이 남아 있지 않다(`git log --oneline -- <spec dir>` → 출력 없음; `git status --short` → `??`). 단조성 판정은 iter1 보고서가 인용한 발췌·계수에 근거하며, 계수·클래스·green path 형태에 대해서는 충분하지만 **전문 diff는 불가능했다**.
- **`.moai/reports/t343/red-now-premeasurement.md`를 열지 않았다.** 저자 측 추론 산출물이므로 M1에 따라 감사 입력에서 제외했다.
- **`.claude/agents/moai/plan-auditor.md`와 `verification-completeness.md`의 미러 전문 대조(`diff`)를 하지 않았다.** `plan.md` §C의 pre-flight 주장(미러가 다르다 / 같다)은 재측정하지 않았다 — 이번 감사의 어떤 판정도 그 값에 기대지 않는다.
- **M3 Go 테스트를 실행하지 않았다.** 아직 존재하지 않는다(E-05). green path의 subtest 명세는 형식으로만 판정했다.
- **706개 SPEC 디렉터리 전수 sweep을 하지 않았다.** iter1과 동일하게 미해소.
- **N3의 span 내부 삽입 뮤테이션을 실제 파일에 심어 관측하지는 않았다.** §2 span의 크기(41행, `verification-completeness.md:120-160`)와 술어의 형태(토큰 존재)로부터 도출한 판정이며, MP-8 span은 아직 존재하지 않아 실측 대상이 없다. 이 항목은 관측이 아니라 구조 추론임을 명시한다.

## Residual risk

- 집계 0.800은 임계와 **정확히 같다**. 채점 4개 중 2개(Completeness 0.90 / Traceability 0.90)가 밴드 1.0에서 감점한 값이고 Testability 0.65는 밴드 사이 값이다 — 판단 폭이 있는 수치이며, 다른 감사자가 ±0.05 내에서 다르게 잴 수 있다. **이 보고서의 FAIL은 그 폭에 의존하지 않는다**(N1 단독으로 성립).
- N1을 고치면 L1의 검사 대상이 넓어지므로, 그 시점에 ledger 11개 명령이 새 형식 검사를 실제로 통과하는지 재확인이 필요하다(현재 원문상 전부 read-only·단일 호출이므로 통과할 것으로 보이나, 이는 예측이지 측정이 아니다).
