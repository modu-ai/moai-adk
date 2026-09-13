# SPEC Review Report: SPEC-DOCS-TABCOUNT-DRIFT-001 (card t530)

Iteration: 3/3 (delta re-audit of D1-D4)
Verdict: **PASS**
Overall Score: **0.8375** (arithmetic) / **0.8352** (harmonic) — Tier M threshold 0.80
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t530`, branch `WT-web-tab-docs`, HEAD `8336f0ac7`, `git merge-base develop HEAD` = `1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09`

Auditor note: 저자의 추론 맥락은 참조하지 않았다. 배차 프롬프트가 운반한 측정치는 **확인할 주장으로** 다뤘고, 전부 이 트리에서 직접 재실행했다(M1 Context Isolation). `mcp__moai__spec_audit` 는 `project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t530` 로 호출했다.

점수 추이 0.66 → 0.775 → **0.8375**. 회귀 없음 — STOP 신호 없음.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `grep -oE '^### REQ-TCD-[0-9]+' spec.md` → `REQ-TCD-001 … 009`, 9개 연속, 결번·중복 없음, zero-padding 일관.
- **[PASS] MP-2 GEARS 형식** (**요구사항 층** `REQ-XXX` 에 대해서만 판정 — `acceptance.md` 의 Given/When/Then 본문은 검증 층이므로 Group 4 에서 채점했다) — `REQ-TCD-005` 는 D10 수리로 본문이 바뀌었으므로 재확인했다: `spec.md:145` `When the guard scans a document, it shall restrict its scan to an explicit 12-file list…` — Event-driven 패턴 유지. 나머지 8개는 iter2 판정에서 본문 무변경.
- **[PASS] MP-3 YAML frontmatter** — `spec.md:2-14` 12 정규 필드 전부 존재, `version: "0.3.0"`, `status: draft`, `tier: M`. snake_case 별칭 없음. `plan.md`·`acceptance.md`·`progress.md` 의 `status:` 는 0 (아래 구조 불변식 표).
- **[N/A] MP-4 언어 중립성** — 이 저장소 자신의 README·docs-site 를 대상으로 하는 SPEC 이며 템플릿 바인딩 콘텐츠가 아니다. 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC 조정** — 참조 2건 모두 해소: `SPEC-WEB-CODEX-PANEL-001` → `status: completed`, `SPEC-PRECOMMIT-GATE-SCOPE-001` → `status: completed`. retired/superseded/archived 없음 ⇒ 조정 의무 없음. `mcp__moai__spec_audit` → `modern_era_clean: 1`, 소견은 INFO `EraAutoDetected` (`H-5 (modern phase or created date)`) 하나뿐.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -rc 'syscall'` 4 아티팩트 + 열거 산출물 전부 `0`. 자동 통과.
- **[PASS] MP-7 해소 게이트** — 게이트가 묶는 표면(`plan.md`, `research.md`)에 미해소 마커 0. `research.md` 는 Tier M 이라 부재. 마커 토큰이 잡히는 곳은 **이전 감사 보고서 2본뿐**(`plan-audit.md` 3, `plan-audit-iter2.md` 1)이며 게이트 범위 밖이다. 이 보고서는 그 토큰을 리터럴로 적지 않는다(해소 선언문 안의 토큰도 grep 게이트에 걸린다는 이 저장소의 기록된 교훈).

must-pass 실패 없음.

---

## Category Scores (rubric-anchored)

| Dimension | iter2 | iter3 | Rubric Band | Evidence |
|-----------|------|-------|-------------|----------|
| Clarity | 0.75 | **0.85** | 0.75-1.0 | D4·D9·D8 수리로 산출물 내부 모순과 `§1.1` 자기 모순이 사라졌다(아래 D4·D9 증거). 감점은 신규 소견 **N1** — `AC-TCD-012` 의 서술된 의도(`acceptance.md:329` "**가드의 낱말 축**이 실제로 걸리는지를 본다")와 기계적 구속(`:343` 셸 재현 명령 5)이 서로 다른 대상을 가리킨다. |
| Completeness | 0.80 | **0.85** | 0.75-1.0 | 전 섹션 존재, `### Out of Scope — ` H3 ×5 각각 구체 불릿. AC 12 / RG 3 / REQ 9 (Tier M 상한 16, 독립 적용 — 여유 있음). 감점은 **N3** — `spec.md §7` 잔여 위험이 수사 클래스 **유지 비용**은 적었으나 **수사 토큰에 낱말 경계가 없다는 축**은 적지 않았다(아래 측정). |
| Testability | 0.70 | **0.80** | 0.75-1.0 | 가장 크게 오른 축이다. `AC-TCD-012` 의 green path 를 **시뮬레이션으로 실행**해 6행 → **0행** 을 관측했고, `AC-TCD-006 V4` 변이가 실제로 잡히는 것도 실행으로 확인했다(둘 다 아래 증거). `AC-TCD-011` 의 B군 삭제 변이 구멍은 추출 범위 한정으로 닫혔다(행 32 / 경로 12 를 같은 영역에서 계수). 감점은 **N1**(가드 낱말 클래스 변이가 열려 있음)과 **N2**(§2.1 4요소 형식이 12개 중 1개에서만 완비). |
| Traceability | 0.85 | **0.85** | 0.75-1.0 | REQ↔AC 매핑을 양방향으로 재확인했다: REQ-TCD-001~009 전부 최소 1개 AC/RG 에 대응, 모든 AC 가 실재 REQ 를 지목. iter2 D8 의 존재하지 않는 AC id 인용 2건 모두 정정됨(아래). `moai spec lint` 는 이 SPEC 에 대해 `CoverageIncomplete` 를 **한 건도** 내지 않았다 — 린트 자신의 커버리지 검사도 같은 결론이다. 감점 유지: **N1** 이 traceability 결함이기도 하다 — `REQ-TCD-005` 의 "both digit and spelled-out numerals" 조항을 구속하는 AC 가 없다. |

Aggregate: 산술 평균 **0.8375**, 조화 평균 **0.8352**. 둘 다 Tier M 임계 0.80 을 넘는다.

---

## D1-D4 per-finding verdict

### D1 (critical, blocking) — `AC-TCD-010` 이동 ref → **REPAIRED**

`acceptance.md:259-300` 이 왼쪽 끝을 읽는 시점 `git merge-base develop HEAD` 로 바꾸고 pathspec 을 열거 12파일로 좁혔다.

```
$ git merge-base develop HEAD
1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09

$ git diff --name-only 1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09 -- <12 files>
(빈 출력)          ← 대조군 0 = 측정 불가 = FAIL (RED-now, 의도된 상태)

$ git diff --name-only 1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09 develop -- <12 files>
README.ja.md
README.ko.md
README.md
README.zh.md
docs-site/content/en/advanced/moai-web-console.md
docs-site/content/ja/advanced/moai-web-console.md
docs-site/content/ko/advanced/moai-web-console.md
docs-site/content/zh/advanced/moai-web-console.md
                   ← 8파일 (종전 넓은 pathspec 은 155파일이었다)
```

세 가지를 요구했고 셋 다 확인했다.

1. **대조군이 명시되어 있다.** `acceptance.md:266-267` 은 0 을 "패리티 유지" 가 아니라 **"측정 불가 = FAIL"** 로 판정하고, `:297` 의 Then 이 `첫 수가 1 이상이고(0 이면 측정 불가 = FAIL)` 로 기계적으로 못 박는다. iter2 가 요구한 형태 그대로다.
2. **한계가 기록되어 있다.** `acceptance.md:299-300` — `병합 뒤에는 이 판정식을 쓰지 않는다 — 병합 후 merge-base 는 카드 tip 자신이 되어 범위가 비고 공허하게 통과한다(gitflow-lane-protocol.md §8 의 한계 조항)`. 병합 후 근거를 트리 동일성으로 대체한다는 점까지 적었다. 감사가 요구한 바로 그 문장이다.
3. **붉은 이유가 바르게 바뀌었다.** 종전에는 남의 커밋 155파일이 분모에 들어가 green path 가 "누군가 무관한 파일을 고친다" 를 지나갔다(`verification-completeness.md` §2 가 실격시키는 모양). 지금은 대조군 0 이 **이 카드가 아직 문서를 고치지 않았기 때문**이며, M2~M4 가 그것을 뒤집는다. 좁힌 pathspec 의 develop churn 8파일은 흡수 시 merge-base 가 전진해 사라지는 값이고, 문서가 그 대비를 명시적으로 적어 두었다(`:279-280`).

`moai spec lint` 의 자체 `MovingRefUnpinned` 규칙도 이 SPEC 에 대해 0건이다(프로젝트 전체로는 116건 존재하는 살아 있는 규칙이다) — 이동 ref 수리가 린트 축에서도 깨끗하다는 독립 신호다.

### D11 (minor, optional) — `RG-TCD-001` 같은 끝점 → **REPAIRED**

`acceptance.md:356` 이 `git diff --quiet "$(git merge-base develop HEAD)" -- assets/images/` 로 바뀌었다. 재실행: 종료코드 `0`(무변경). `:357-358` 이 `echo "rc=$?"` 를 붙이지 말라고 명시한 것도 맞다 — 그 형태는 `git diff` 의 성공 여부를 찍을 뿐 무매치를 뜻하지 않는다.

### D2 + D3 (major, blocking) — 낱말 축 → **REPAIRED, 실행으로 확인**

이 축은 주장을 읽는 것으로 끝내지 않고 **세 번 실행**했다.

**(a) 적중 집합 — 재현 명령 5 축어 실행 (HEAD `8336f0ac7`)**

```
docs-site/content/en/cli-reference/web.md:53          (A2)
docs-site/content/zh/cli-reference/web.md:53          (A4)
docs-site/content/en/advanced/moai-web-console.md:128 (B4)
README.md:414                                          (C1)
README.ko.md:414                                       (C2)
README.zh.md:414                                       (C4)
--- 6
```

원시 스윕 **8행**. 배차가 운반한 "6행 = A2·A4·B4·C1·C2·C4, 허용되지 않은 오탐 0" 은 참이다.

**(b) 파일 범위 위험 — 전역 `grep -v 'codex'` 로 바꾼 같은 파이프라인**

```
docs-site/content/zh/cli-reference/web.md:53
docs-site/content/en/cli-reference/web.md:53
docs-site/content/en/advanced/moai-web-console.md:128
--- 3
```

README 3행(C1·C2·C4)이 조용히 사라진다. 저자가 스스로 찾아 [HARD] 로 못 박은 위험이며, 실재한다.

네 자리 배치를 확인했고 **닫혔다고 판정한다**: `spec.md:145-161`(REQ-TCD-005 본문 + `The file scope is load-bearing` 문단), `plan.md:64-67`([HARD] 표시 + 실측), `acceptance.md:229-230`(`규칙에 파일 범위가 없으면 FAIL`), `acceptance.md:346-347`(AC-TCD-012 의 And 절이 `grep -vE` 패턴에 파일 경로 포함을 요구), 그리고 `tab-count-sites.md:143-147`·`:233-234`. 네 곳 중 **둘은 인수 기준의 FAIL 조건**이라 산문이 아니라 판정이다.

**(c) green path — 카드 자신의 작업만으로 초록에 닿는가 (시뮬레이션)**

트리를 건드리지 않기 위해 12파일을 스크래치로 복사하고, `plan.md §B` 의 처분대로 낱말 표기 6자리만 최소 구절 편집했다(`The nine settings tabs.` → `The settings tabs.`, `设置九个标签页` → `设置标签页`, `unfolds fourteen tabs below it` → `unfolds the tabs below it`, `splits into fourteen tabs:` → `splits into the tabs:`, ` 열네 개 탭으로 나뉜다` → ` 탭으로 나뉜다`, `分成十四个标签页` → `分成标签页`). 그 트리에서 재현 명령 5 를 다시 돌렸다:

```
(빈 출력)
--- 0
```

**`AC-TCD-012` 는 도달 가능하다.** 다른 자리를 고치지 않고 이 카드의 작업만으로 6 → 0 이 된다 — wrong-reason red 도 red-forever 도 아니다. 이것이 D1 에서 요구했던 두 셀 규율(`verification-completeness.md` §2)을 낱말 축에서 실제로 충족한 증거다.

**(d) `AC-TCD-006 V4` 변이 — 축이 살아 있는가**

위 초록 트리에 V4 를 넣었다(`unfolds the tabs below it` → `unfolds fourteen settings tabs below it`):

```
docs-site/content/en/advanced/moai-web-console.md:128:Choosing Settings in the rail unfolds fourteen settings tabs below
--- 1
```

**iter2 가 D2 로 성공시킨 바로 그 변이체가 이제 잡힌다.** 리터럴 층 0 / 숫자 스윕 0 이던 자리를 낱말 축이 단독으로 잡는다.

**D3(전제 철회)** 도 확인했다: `plan.md:92-110` 이 이전 판의 근거를 명시적으로 철회하고(`이전 판의 근거를 철회한다 — 낱말 축을 닫아 둘 실측 근거는 없었다`), 단계별 표(8 → 7 → 6, 허용되지 않은 오탐 0)로 대체했다. `spec.md:273-279` 의 §7 잔여 위험도 "오탐" 이 아니라 **수사 클래스 유지 비용**으로 다시 쓰였다. `verification-claim-integrity.md` §1.1 surface 4(권고의 전제도 결함 주장과 같은 증거 부담을 진다)가 요구한 형태다.

### D4 (major, blocking) — 과장 서술 → **REPAIRED (전 아티팩트 훑음)**

두 줄만 보지 않고 아티팩트 집합 전체를 훑었다.

```
$ grep -rn '독립인 세 측정|서로 독립인 세|세 측정|three independent' <spec dir> tab-count-sites.md count-literals.txt
progress.md:10:  "서로 독립인 세 측정" 이 아니다(`spec.md §1.1`).
spec.md:28:      | 0.1.0 | … **[0.2.0 정정: "서로 독립인 세 측정" 은 과장이었다 — …]**
spec.md:49:      **정확히 말하면 원천 1 + 미러 1 + … — "서로 독립인 세 측정" 이 아니다.**
tab-count-sites.md:16:  **독립인 세 측정이 아니다.** …
tab-count-sites.md:197: # 1) 기준값 — 원천 1 + 미러 1 + 둘의 일치 테스트 1 (독립인 세 측정이 아니다)
```

남은 5건이 **전부 부정문이거나 정정 주석이 붙은 보존된 HISTORY 행**이다. iter2 가 지목한 `tab-count-sites.md:14` 는 이제 `:14-19` 에서 `spec.md §1.1` 과 같은 문장·같은 뜻으로 다시 쓰였고, 파일 안 `:14` 와 `:167` 의 모순이 사라졌다.

부수 지목도 확인:
- **D9** `spec.md:48` "원천 두 개" → `grep -rn '원천 두 개'` 무매치. `:49` 와 `:52` 가 일치한다.
- **D8** 존재하지 않는 AC id 2건 → `tab-count-sites.md:194` 는 `AC-TCD-011`(경로 완전성, 올바름), `plan.md:127` 은 `AC-TCD-005`(swept 계수, 올바름). 둘 다 정정됨. `AC-TCD-012` 는 이제 실재하기도 한다.
- **D10** `REQ-TCD-005` 의 행 번호 하드코딩 → 본문이 구절 식별로 바뀌었다(`spec.md:151` `Negative cases are named by phrase, never by line number`). REQ 본문에 bare 행 번호 인용 무매치.
- **D5** `AC-TCD-011` B군 삭제 변이 → 추출 범위를 `^| [ABCD][0-9]` 로 한정하고 B군 경로 축약을 전체 경로로 되돌렸다. 실측: 행 `32`, 경로 `12`, 경로 실재 검사 무출력. 이제 표 하나가 사라지면 두 수치가 함께 떨어진다.
- **D6** `swept` 카운터 → `acceptance.md:175-179` 가 [HARD] 로 "읽기 성공 **직후** `n++`, 읽기 실패는 `t.Fatal`" 을 규정하고 `len(targetFiles)` 구현을 명시적으로 실격시킨다.
- **D7** `allowed` 단위 → `allowed rules 1` + `allowed lines 16` 으로 분리. 실측 재확인: `grep -ci codex` 4로케일 각 `4`, 합계 **16**. 수치가 맞다.

---

## 신규 소견 (수리가 도입했거나 감사가 새로 잰 것)

**N1. 가드의 **낱말 클래스**를 구속하는 인수 기준이 없다 — `REQ-TCD-005` 의 조항이 검증 층에 대응 항목을 갖지 못했다** — `acceptance.md:58,327-347` vs `spec.md:145-147` — Severity: **major** — Class: **blocking**

`REQ-TCD-005` 는 가드가 `both digit and spelled-out numerals` 를 매치하라고 **명시한다**. 그런데 그 조항을 기계적으로 구속하는 AC 가 없다.

- `AC-TCD-012` 는 D 매트릭스(`:58`)와 도입 문단(`:329`)에서 "**가드의** 낱말 축 적중" 을 본다고 서술하지만, When(`:343`)이 실행하는 것은 `tab-count-sites.md § 재현 명령 5` — **셸 파이프라인**이지 Go 가드가 아니다. 서술된 의도와 기계적 구속이 서로 다른 대상을 가리킨다.
- `AC-TCD-006 V4` 는 가드를 구속하지만 **변이 한 건**(`fourteen`)뿐이다.

변이체가 쓰인다: 낱말 클래스를 `{fourteen}` 하나로만 구현한 Go 가드는 V1·V2(이름) · V3(숫자) · V4(`fourteen`) 를 전부 통과하고, `AC-TCD-012`(셸) 도 통과하며, `AC-TCD-005`·`007`·`008`·`009` 도 통과한다. 그러면서 `nine` · `열네` · `十四` · `九个` 를 놓쳐 **열거된 낱말 표기 6자리 중 5자리가 가드 밖에 남는다** — 이 카드가 "지난 청소가 실패한 원인" 으로 지목한 바로 그 부류다(`spec.md:89-90`). `verification-completeness.md` §2 의 mutant probe 가 "기준을 만족시키면서 요구사항을 위반하는 변이체를 쓸 수 있으면 그 기준은 채택하기에 얕다" 고 말하는 형태다.

**Required fix (한 문장이면 된다)**: `AC-TCD-012` 의 When 을 둘로 만든다 — 셸 재현 명령 5 **그리고** 가드 자신의 낱말 축 적중 집합을 출력시켜(예: `word-axis hits 6` 토큰) M2 이전 트리에서 둘이 **같은 6행**임을 단정한다. 또는 `AC-TCD-006` 에 V5·V6(`nine`, `열네` 또는 `十四`)를 추가해 클래스의 로케일 폭을 변이로 구속한다. 전자가 싸고 정확하다.

**N2. RED-now 셀이 `verification-completeness.md` §2.1 의 4요소 중 2요소만 운반한다 (12개 중 11개)** — `acceptance.md:45-58, 75-347` — Severity: major — Class: **optional**

§2.1 은 release-blocking 기준의 RED-now 셀에 **명령 · 그 명령의 축어 stdout · 종료코드(별도 필드) · 측정 트리 SHA** 네 가지를 함께 요구한다. 이 문서는 명령과 SHA 는 어디에나 있다(문서 수준 pin `acceptance.md:20-21` 이 §2.1 이 허용하는 형태로 모든 기준에 구속된다 — 올바른 사용이다). 그러나 **축어 stdout 과 별도 종료코드 필드를 모두 갖춘 것은 `AC-TCD-004` 하나뿐**이다(`:146-152`). 나머지 11개는 `실측: 6행 적중` 처럼 **요약된 수치**를 적는다.

§2.1 은 "세 요소만 운반하는 셀은 미채택" 이라고 쓴다. 다만 **보고하되 차단으로 올리지 않는** 이유를 밝힌다: §2.1 의 undecidable disposition 은 "인용된 RED 를 현재 트리에서 재실행할 수 없을 때" 발동하는데, **나는 이 감사에서 그 셀들을 실제로 재실행했고 전부 재현했다**(AC-001 4행, AC-002 20행, AC-003 12행, AC-004 축어 일치, AC-007 16줄, AC-010 대조군 0, AC-011 32/12, AC-012 6행). 즉 §2.1 이 지키려는 실질(RED 가 실재하고 재실행 가능하다)은 충족되어 있고, 미비한 것은 기록 형식이다. 또 §2.1 이 허용하는 대체 운반체(외부 evidence-ledger 를 id 로 인용)에 `AC-TCD-012` 가 이미 근접해 있다(`tab-count-sites.md § 재현 명령 5` 인용).

**Required fix**: `progress.md §E.2` 가 이미 "RED-now 출력과 green 출력을 **모두** 축어로 남긴다" 를 DoD 2번으로 요구하므로(`acceptance.md:395`), run-phase 가 그 자리를 채우면 형식이 완비된다. 추가 수리를 plan-phase 에 요구하지 않는다.

**N3. 수사 토큰에 낱말 경계가 없고, 그 축이 잔여 위험에 적혀 있지 않다** — `plan.md:84-90`, `spec.md:273-279` — Severity: minor — Class: **optional**

가드 정규식은 **탭 명사**에는 낱말 경계를 붙이지만(`\btabs?\b` — `selectable` 을 닫는 장치, 실측 확인) **수사 토큰에는 붙이지 않는다**. 한국어 단음절 수사(`한`·`두`·`세`·`네`·`열`)는 평범한 낱말의 부분 문자열이다. 측정:

```
1  <- 세부 설정 탭을 연다        (세부의 '세')
1  <- 네트워크 관련 탭            (네트워크의 '네')
1  <- 열기 버튼이 있는 탭         (열기의 '열')
1  <- phone settings tab         (phone 의 'one')
0  <- Everyone can open the tab
```

**오늘 이 트리에서는 오탐 0 이 사실이다**(위 D2 (a) 에서 직접 확인했다) — 지금 결함이 아니라 앞으로의 노출이다. 다만 `spec.md §7` 의 잔여 위험 문단은 **"새 표기가 들어오면 클래스를 손봐야 한다"**(누락 방향)만 적고 **"평범한 낱말이 수사로 읽힌다"**(오탐 방향)는 적지 않았다. `plan.md §A.4` 의 장치 표도 낱말 경계를 탭 명사에만 귀속시킨다.

**Required fix**: `spec.md §7` 의 낱말-축 불릿에 한 문장을 더한다 — 한국어·영어 수사 토큰은 낱말 경계가 없어 `세부`/`네트워크`/`열기`/`phone` 류가 미래 문장에서 걸릴 수 있고, 그때 대응은 허용 목록 확대가 아니라 수사 토큰에 경계를 붙이는 것이라고.

**N4. 재현 명령 5 의 2단계 정규식이 1단계의 부분집합이 아니다 — 수사 클래스가 조용히 좁아진다** — `tab-count-sites.md:224-229` — Severity: minor — Class: **optional**

파이프라인 1단계는 `seventeen|eighteen|nineteen|twenty` 와 `열한|열두|열세|열다섯|열여섯` 를 포함하는데, `sed` 뒤의 2단계 재필터는 그것들을 빼고 다시 쓴다. 대부분은 우연히 구제된다(`seventeen` ⊃ `seven`, `nineteen` ⊃ `nine`, `열두` ⊃ `열`). 구제되지 않는 것이 있다:

```
stage1=1 stage2=0  <- twenty tabs
stage1=1 stage2=1  <- seventeen tabs
stage1=1 stage2=1  <- 열두 개 탭
```

`twenty tabs` 는 1단계를 통과하고 2단계에서 **말없이 버려진다**. 서수 접두 제외는 `sed` 가 수행하므로 2단계 정규식이 1단계와 달라야 할 이유가 없다 — 두 벌을 손으로 적으면서 갈라진 것으로 보인다.

**Required fix**: 2단계 정규식을 1단계와 **동일한 문자열**로 맞춘다(`sed` 가 이미 `第X` 로 바꿔 놓았으므로 같은 패턴을 그대로 재사용해도 서수는 걸리지 않는다). 산출물 한 곳의 수정이고, Go 가드가 이 파이프라인을 옮겨 적을 때 같은 비대칭을 물려받지 않게 한다.

**N5. `(none) → draft` 전이는 여전히 귀속되지 않는다 — 트레일러가 전이를 운반하지 않는 커밋에 붙었다** — 커밋 `8336f0ac7` vs `c205eeeff` — Severity: minor — Class: **optional** — 정보성

배차가 운반한 주장 두 가지를 각각 확인했고, 하나는 참, 하나는 결론이 다르다.

*참인 쪽* — 소비자 정규식은 실제로 매치한다. `internal/spec/lint_ownership.go:256` 의 `var authoredByAgentLine = regexp.MustCompile(\`(?mi)^\s*Authored-By-Agent:\s*(\S+)\s*$\`)` 를 소스에서 확인했고, `git log -1 --format='%b' 8336f0ac7` 의 28번째 줄이 그 패턴에 걸린다. git 자신의 트레일러 파서는 `🗿 MoAI` 서명 줄 때문에 보지 못한다(`%(trailers:key=...)` 빈 출력) — 배차의 서술대로다.

*결론이 다른 쪽* — **그럼에도 D12 는 해소되지 않았다.** 트레일러가 붙은 `8336f0ac7` 은 `status:` 를 바꾸지 않는다(바뀐 것은 `version: "0.2.0"` → `"0.3.0"` 뿐). `(none) → draft` 전이를 운반하는 커밋은 `c205eeeff` 하나이고 거기엔 트레일러가 없다. 린트를 돌려 확인했다:

```
$ moai spec lint --json   (exit 0)
{
  "file": ".../SPEC-DOCS-TABCOUNT-DRIFT-001/spec.md", "line": 1,
  "severity": "info", "code": "OwnershipTransitionUnmeasured",
  "message": "SPEC SPEC-DOCS-TABCOUNT-DRIFT-001 transition \"(none)\" → \"draft\" expected owner
   \"manager-spec\" but commit c205eeeff1793157c7d99163998e286988f14a92 (…) has no
   Authored-By-Agent trailer — ownership transition unmeasured"
}
```

이 SPEC 을 지목하는 린트 행은 **이 한 건뿐이고 INFO** 이며, `spec-frontmatter-schema.md` § OwnershipTransitionRule 이 규정하듯 `--strict` 도 이것을 에러로 올리지 않는다. 전이 커밋은 이미 착지했으므로 이 판에서 고칠 것이 없다 — **부채로 기록한다**. 다음 plan-phase 커밋은 트레일러를 **전이를 운반하는 커밋 자신**에 붙인다.

---

## 채택 가능성 판정 — 가드가 아직 없는데 인수 기준을 채택해도 되는가

배차가 명시적으로 물은 항목이다. `internal/web/docs_tab_contract_test.go` 는 존재하지 않는다(`ls` 확인). 판정: **`AC-TCD-004`~`009` 와 `AC-TCD-012` 는 plan-phase 에서 채택 가능하다.** 근거는 셋이다.

1. **`verification-completeness.md` §2 가 요구하는 것은 "지금 붉고, 그 이유가 바르며, 이 작업이 뒤집을 수 있다" 이지 "가드가 이미 있다" 가 아니다.** `AC-TCD-004` 의 RED-now 는 축어로 적혀 있고 내가 재실행해 재현했다(`testing: warning: no tests to run` / `PASS` / `ok … [no tests to run]`). 나머지의 RED-now 인 "가드 부재" 는 실제로 관측되는 조건이고, 그 이유가 명시되어 있다.
2. **낱말 축의 인수 판정은 가드에 의존하지 않는다.** `AC-TCD-012` 의 When 은 셸 파이프라인이므로 Go 구현 여부와 무관하게 실행 가능하고, 그 green path 를 나는 시뮬레이션으로 도달시켰다(6 → 0). 즉 이 카드의 가장 새로운 인수 주장은 **미작성 가드에 얹혀 있지 않다**.
3. **§2.1 의 undecidable disposition 은 발동하지 않는다.** 그 조항은 "인용된 RED 를 현재 트리에서 재실행할 수 없을 때" 를 위한 것이고, 여기 셀들은 전부 재실행되어 재현됐다.

다만 **바로 그 분리가 N1 을 만든다**: 인수 판정이 셸에 얹히고 가드 구속이 V4 한 건뿐이라, 가드의 낱말 클래스 폭 자체는 어느 기준에도 걸려 있지 않다. 그러니 "미작성 가드가 채택을 막는가" 의 답은 **아니오**이고, "미작성 가드가 무언가를 열어 두는가" 의 답은 **예 — N1** 이다. 둘은 다른 질문이며, 후자가 이 감사가 남기는 유일한 blocking 소견이다.

---

## Defects Found (구조화 목록)

D1. N1 — `acceptance.md:58,329,343` / `spec.md:145-147` — 가드 낱말 **클래스**를 구속하는 AC 부재; `{fourteen}` 만 구현한 가드가 전 기준을 통과하며 6자리 중 5자리를 놓친다 — Severity: major — Class: **blocking** — Required fix: `AC-TCD-012` 에 가드 자신의 낱말 축 적중 집합(6행)을 셸과 대조하는 And 절을 추가하거나, `AC-TCD-006` 에 `nine`·`열네`(또는 `十四`) 변이를 추가한다.
D2. N2 — `acceptance.md` 전 AC — §2.1 RED-now 4요소 중 축어 stdout·별도 종료코드가 `AC-TCD-004` 외 11개에서 누락 — Severity: major — Class: optional — Required fix: run-phase 가 DoD 2번(`:395`)에 따라 `progress.md §E.2` 를 축어로 채우면 완비된다. plan-phase 추가 수리 불요.
D3. N3 — `spec.md:273-279`, `plan.md:84-90` — 수사 토큰의 낱말 경계 부재가 잔여 위험에 미기재(실측: `세부`/`네트워크`/`열기`/`phone` 적중) — Severity: minor — Class: optional — Required fix: §7 낱말-축 불릿에 오탐 방향 한 문장 추가.
D4. N4 — `tab-count-sites.md:224-229` — 파이프라인 2단계 정규식이 1단계의 부분집합이 아님(`twenty tabs` 소실) — Severity: minor — Class: optional — Required fix: 2단계 패턴을 1단계와 동일 문자열로 통일.
D5. N5 — 커밋 `c205eeeff` — `(none) → draft` 전이 미귀속(트레일러가 `8336f0ac7` 에 붙었으나 그 커밋은 전이를 운반하지 않음) — Severity: minor — Class: optional — 정보성 — Required fix: 이 카드에는 없음. 다음 plan-phase 커밋은 트레일러를 전이 커밋 자신에 붙인다.

---

## Regression Check (iteration 2 defects)

| iter2 | 판정 | 증거 |
|---|---|---|
| D1 `AC-TCD-010` 이동 ref (critical) | **RESOLVED** | merge-base 끝점 + 12파일 pathspec + 대조군 + 병합 후 한계 조항, 넷 다 확인 |
| D2 낱말 표기 변이체 (major) | **RESOLVED** | V4 변이가 실행으로 잡힘(1행); green path 6 → 0 도달 확인 |
| D3 `§A.4` 전제 미뒷받침 (major) | **RESOLVED** | 근거 명시적 철회 + 단계별 재측정(8→7→6, 오탐 0) 재현 |
| D4 산출물 과장 (major) | **RESOLVED** | 전 아티팩트 훑음 — 잔존 5건 전부 부정문/보존 HISTORY |
| D5 `AC-TCD-011` 변이 (minor) | **RESOLVED** | 추출 범위 한정 + B군 전체 경로 복원; 행 32 / 경로 12 |
| D6 `swept` 카운터 (minor) | **RESOLVED** | [HARD] 읽기 성공 후 증가 + `t.Fatal` 규정 |
| D7 `allowed` 단위 (minor) | **RESOLVED** | `rules 1` + `lines 16` 분리, 실측 16 재확인 |
| D8 존재하지 않는 AC id (minor) | **RESOLVED** | 두 인용 모두 정정 |
| D9 `§1.1` 자기 모순 (minor) | **RESOLVED** | `원천 두 개` 무매치 |
| D10 `REQ-TCD-005` 행 번호 (minor) | **RESOLVED** | 구절 식별로 교체 |
| D11 `RG-TCD-001` 고정 base (minor) | **RESOLVED** | merge-base 끝점, rc=0 |
| D12 트레일러 (minor, 정보성) | **UNRESOLVED (부채)** | N5 — 전이 커밋에 붙지 않아 린트가 여전히 INFO 를 낸다 |

세 반복을 통틀어 바뀌지 않고 살아남은 결함은 없다 — stagnation 신호 없음.

---

## 확인하지 못한 것 (Gaps)

- **가드가 존재하지 않는다.** `AC-TCD-005`~`009` 의 변이 내성은 명세 독해 + 표적 grep 이고, 실제 Go 가드에 대해 붉음을 실행한 적이 없다. N1 이 이 공백에서 나왔다.
- **`AC-TCD-012` 의 green path 는 시뮬레이션이다.** 감사는 카드 트리를 변경하지 않으므로 12파일을 스크래치로 복사해 편집했다. 6 → 0 은 그 사본에서의 관측이며, run-phase 가 같은 처분을 실제 트리에 적용했을 때의 재현은 M2~M3 가 확인한다.
- **`RG-TCD-002` 미실행.** `hugo` 빌드를 돌리지 않았다. 무경고 주장은 아티팩트에서 옮겨 온 것이지 이 감사가 잰 값이 아니다.
- **D1 의 병합 후 거동은 추론이다.** 병합을 수행하지 않았다. "흡수 후 merge-base 가 흡수한 develop 커밋으로 전진한다" 는 git 의 정의상 참이지만, 병합 트리에서 실측한 것은 아니다.
- **N3 의 오탐은 합성 문자열로 쟀다.** 실제 미래 문서에서 그 표현이 나타날 확률은 재지 않았다 — 오늘 이 트리의 오탐이 0 이라는 사실은 별도로 실측했다.

## 잔여 위험 (Residual-risk)

- **N1 이 run-phase 에서 실현될 수 있다.** 가드를 쓰는 사람이 `REQ-TCD-005` 의 "both digit and spelled-out" 조항을 읽지 않고 V4 만 보고 클래스를 좁히면, 카드가 자기 주제를 절반만 닫은 채 초록으로 마감된다. 요구사항이 명시적이라는 것이 유일한 방어선이고, 기계적 방어선은 없다.
- **허용 규칙 1개가 16줄을 면제한다.** `AC-TCD-007` 이 그 표면의 **크기**를 보고하지만 **내용**은 보지 않는다. 허용된 줄 안에 진짜 설정 탭 계수가 섞이면 가드가 침묵한다 — SPEC 이 §7 에 정직하게 적어 둔 값이다.
- **N4 의 비대칭이 Go 가드로 복제될 수 있다.** 구현자가 재현 명령 5 를 옮겨 적으면 두 벌 패턴을 함께 가져갈 소지가 있다.

---

## Recommendation

**PASS**, 0.8375 / 임계 0.80. must-pass 7개 중 실패 0(MP-4 는 N/A). 점수는 세 반복에 걸쳐 단조 상승했고(0.66 → 0.775 → 0.8375), iter2 의 blocking 4건은 전부 해소됐으며, 그 중 셋은 **문서를 읽어서가 아니라 명령을 실행해서** 확인했다.

이번 반복은 감사가 기대할 수 있는 최선의 모양에 가깝다 — 저자가 자기 수리에서 스스로 위험(전역 `grep -v 'codex'`)을 찾아 [HARD] 로 네 곳에 박았고, 나는 그 위험이 실재함(6 → 3)과 방어가 유효함을 각각 실행으로 확인했다.

**iteration 3 이 상한이므로 네 번째 반복을 제안하지 않는다.** 남은 blocking 소견 N1 에 대해 증거가 지지하는 처분은 셋 중 **PASS-with-debt 도, scope reduction 도, operator override 도 아닌 네 번째 길** 이다 — 정확히 말하면 이들 셋은 "임계를 못 넘었을 때" 를 위한 선택지이고, 여기서는 임계를 넘었다. N1 은 **run-phase M1 의 입구 작업**으로 라우팅하는 것이 맞다:

1. **N1 을 M1 착수 조건으로 붙인다 (권장).** 고칠 것은 인수 기준 한 문장이고, 그 문장이 구속하는 대상(가드)은 M1 이 지금 쓰려는 바로 그 파일이다. 가드를 쓰기 **전에** 기준을 확정하는 것이 순서상 자연스럽고, 별도 plan 반복의 비용을 치르지 않는다. `manager-spec` 이 `acceptance.md` 한 줄을 고치고(인수 층 소유권), M1 은 그 기준을 보고 클래스를 쓴다.
2. **그대로 진행하고 N1 을 부채로 기록한다.** `REQ-TCD-005` 가 명시적이므로 구현자가 요구사항을 읽으면 결함은 발생하지 않는다. 다만 방어가 사람의 독해뿐이라는 점을 `progress.md` 에 남긴다.

선택지 2 를 취하더라도 카드는 정상적으로 진행 가능하다. N3·N4·N5 는 선택 사항이며, N4 는 산출물 한 곳 수정이라 M1 에 함께 접는 비용이 사실상 0 이다. 판단은 운영자의 것이고, 이 보고서는 그 판단이 가능하도록 근거를 남기는 데까지가 역할이다.
