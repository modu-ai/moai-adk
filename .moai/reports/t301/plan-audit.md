# SPEC 감사 보고서: SPEC-FULL-SUITE-DOCTRINE-001

Iteration: 1/2 (Tier M 선언 기준 상한. 단, `tier:` 프론트매터 부재로 기계적 해석은 Tier L=3 — D7 참조)
Verdict: **FAIL**
Overall Score: **0.57** (조화평균; Tier M 기준선 0.80, Tier L 기준선 0.85 — 어느 쪽으로 읽어도 미달)

측정 트리: 워크트리 `.claude/worktrees/t301`, HEAD `d29b8942e`, branch `WT-full-suite-doctrine`.
저자 추론 맥락은 M1 Context Isolation에 따라 무시했다 (Reasoning context ignored per M1 Context Isolation). 판정 근거는 SPEC 산출물 4종과 이 트리에서 직접 실행한 명령 출력뿐이다.

---

## Must-Pass 결과

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-FSD-[0-9]*' spec.md | sort | uniq -c` 실행 결과 REQ-FSD-001~010이 각 1회, 결번·중복 없음, 3자리 패딩 일관. 근거: `spec.md:71-92`.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 층 판정)** — 판정 대상은 `spec.md §B`의 `REQ-XXX` 항목 10개다. 검증 층(`acceptance.md`의 Given-When-Then)은 이 기준으로 재지 않았다. 001/003/004/005 Ubiquitous, 002/007/009 Unwanted(`shall not`), 006/010 Event-driven(`When …, … shall …`), 008 Capability gate(`Where …`). 근거: `spec.md:71-92`.
  - 경계 1건(차단 아님): REQ-FSD-010은 `its duration shall be recorded` — 주어가 행위자가 아닌 산출물(duration)의 수동형이다. GEARS의 일반화된 `<subject>`가 artifact를 허용하므로 통과시키되, 기록 주체가 문면에 없다는 점은 D5와 함께 읽어야 한다.
- **[PASS] MP-3 YAML 프론트매터 유효성** — 정본 12필드가 모두 존재하고 snake_case 별칭(`created_at`/`updated_at`/`labels`/`spec_id`)은 0건. `phase: "v3.1.4 target"` 은 금지된 라이프사이클 토큰(plan/run/sync/mx)이 아니다. 근거: `spec.md:2-13`.
  - 참고(이 SPEC의 결함 아님): `progress.md:16` 의 ID 자가 점검이 사용한 정규식 `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` 는 실제 구현(`internal/spec/lint.go:715` `specIDPattern`)과 **문자 그대로 일치**한다. 자가 점검은 올바른 계측기를 썼다. 단일 세그먼트로 적힌 쪽은 규칙 문서(`spec-frontmatter-schema.md` 필드표)이며 그쪽이 구현에 뒤처져 있다 — 별도 소관.
- **[N/A] MP-4 §22 언어 중립성** — 이 SPEC은 다중 언어 툴링을 다루지 않는다. 기존 Go 전용 지시문 1개를 교정하는 단일 언어 범위 작업이고, 16개 언어 중 일부만 열거하는 서술을 새로 만들지 않는다. 관련 관찰은 D11(optional)로 분리.
- **[PASS] MP-5 D7 교차 SPEC 정합** — 검증 동사 실행 가능. `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → `SPEC-FULL-SUITE-DOCTRINE-001` 자기 참조 1건뿐이고 해당 디렉터리는 존재한다. retired/superseded/archived 참조 0건 → BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall' .moai/specs/SPEC-FULL-SUITE-DOCTRINE-001/*.md` → 4개 파일 모두 `0`. D8-4에 따라 자동 통과.
- **[PASS] MP-7 clarification 게이트** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-FULL-SUITE-DOCTRINE-001/` → 매치 없음(rc=1). `research.md` 는 부재(Tier M 산출물 집합)이므로 그 축은 N/A.

Must-Pass 7항 전부 통과다. **FAIL 판정은 M5 방화벽이 아니라 집계 점수와 차단성 결함에서 나온다.**

---

## 항목별 점수 (루브릭 앵커)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | REQ-FSD-006이 문자 그대로 이행되면 REQ-FSD-009를 깨뜨린다(D1). AC-FSD-006의 "이 AC를 좁힌다"가 4개 파일 행 전부인지 ACPR 한정인지 불명(D2). `acceptance.md:107`, `spec.md:82`, `spec.md:88` |
| Completeness | 0.75 | 0.75 | 필수 절 전부 존재. `### Out of Scope — <주제>` H3 4개에 각각 `-` 불릿 존재(`spec.md:110-132`). 프론트매터 12필드 완비. 감점 사유는 `tier:` 부재 1건(D7) |
| Testability | 0.50 | 0.50 | 이진 판정 불가 AC 다수: AC-FSD-005는 "눈으로 확인"에 판정을 맡기고(`acceptance.md:89`), AC-FSD-007도 "눈 확인"(`:119`), AC-FSD-009 첫 블록은 명령과 Then이 서로 다른 것을 잰다(`:147-154`), AC-FSD-006은 저자 스스로 틀렸을 수 있다고 적은 기대치를 그대로 싣는다(`:107`) |
| Traceability | 0.60 | 0.50↔0.75 | `§D.2` 표는 형식상 완전(REQ 10개 전부 AC 보유, 고아 AC 없음). 그러나 REQ-FSD-001·004의 **긍정 절**과 REQ-FSD-005의 **시간 추정치 절**은 대응 AC가 그 내용을 전혀 재지 않는 공허한 매핑이다(D3·D5). 10개 중 3개가 실질 미커버 |

조화평균: `4 / (1/0.50 + 1/0.75 + 1/0.50 + 1/0.60)` = `4 / 7.000` = **0.571**.

---

## 발견된 결함

**D1. REQ-FSD-006이 배포 템플릿에 리포 지역 브랜치명을 박는다 — `spec.md:82` — Severity: critical — Class: blocking**

REQ-FSD-006의 문면은 완료 보고가 "the full-suite verdict is delegated to `origin/develop` CI" 임을 말하도록 **에이전트 정의에 요구**한다. 그런데 REQ-FSD-008은 그 파일(`manager-develop.md`)의 수리가 `internal/template/templates/` 미러에 착지해야 한다고 요구하고, REQ-FSD-009는 미러된 문면에 리포 지역 내용이 들어가서는 안 된다고 요구한다. 세 요구사항을 동시에 만족시킬 수 없다.

측정: `grep -rln 'origin/develop' internal/template/templates/` → **출력 없음**. `origin/develop` 은 현재 배포 템플릿 전체에 0회 등장한다. 이 SPEC이 그 첫 사례를 만든다. `develop` 통합 브랜치 모델은 로컬 전용 규칙 `.claude/rules/local/gitflow-lane-protocol.md` 소관이고, 그 디렉터리에는 템플릿 미러가 없다(`ls internal/template/templates/.claude/rules/local/` → No such file or directory). 배포판 `spec-workflow.md`는 여전히 `main` 기준 Route A/B를 서술한다. 하류 사용자에게 `origin/develop` 은 존재하지 않는 브랜치다.

AC-FSD-009는 이 위반을 **잡지 못한다** — 패턴 집합(`SPEC-`/`CLAUDE.local`/`2026-`/`load 413`/`/Users/`)에 브랜치명이 없다.

내부 불일치도 함께 있다: `plan.md:57` (M1.3)은 상위 계약의 일반 근거를 모델로 삼고 "지역 사정은 한 글자도 넣지 않는다"고 적는다. 상위 계약 원문(`AGENTS.md:117-118`, 이 트리에서 재확인)은 브랜치를 명명하지 않는다 — "then push and let CI run the full suite". plan.md가 옳고 REQ-FSD-006이 그르다.

**필요한 수정**: REQ-FSD-006에서 `origin/develop` 을 제거하고 판정 주체를 브랜치 중립적으로 재작성한다(예: "the CI run on the integration branch after push"). `origin/develop` 이라는 지역 사실은 `spec.md §A.3`의 배경 서술에만 남긴다 — spec.md는 배포되지 않는다. 동시에 AC-FSD-009의 패턴 집합에 `origin/develop` 을 추가한다.

---

**D2. AC-FSD-006이 틀렸을 수 있다고 스스로 적은 기대치를 싣고, 금지된 소유권 교차를 지시한다 — `acceptance.md:104-107` — Severity: critical — Class: blocking**

두 개의 결함이 한 항목에 겹쳐 있다.

(가) 사후 기대치 `:0` 이 저자 자신의 분석상 달성 불가능할 수 있다. 이 트리에서 재측정한 3히트의 실체:

```
:22  go test ./... > /tmp/moai-verify/1-go-test.log ...     ← 배치 1번 항목 (범위 안)
:65  Turn 1: go test ./...     → wait for completion ...    ← "직렬 검증 안티패턴" 예시 블록
:76  - Commands that depend on each other (e.g., `make build` before `go test ./...`)  ← "언제 직렬로" 예시
```

`:65`·`:76`은 처방이 아니라 **예시 산문**이다. 전자는 하지 말라는 패턴을 보여주고, 후자는 의존 관계 예시다. 전량 실행을 처방하지 않는다. 따라서 `:0`을 그대로 요구하면 범위 밖 산문을 뜯어고쳐야 한다. `verification-completeness.md §2`의 표현으로는 "도착 시점에 RED이고 구현 후에도 RED인" 방향 — 이 작업이 만들 수 있는 어떤 변경으로도 뒤집히지 않는다.

(나) 완화 절차가 금지된 행위를 지시한다. `:107`은 "구현자는 … 이 AC를 **1번 항목 줄 한정으로 좁히고** 그 판단을 progress.md에 남긴다"고 적는다. AC를 좁히는 것은 `acceptance.md` 본문 편집이며, `manager-develop`은 `spec.md`/`plan.md`/`acceptance.md` 본문 수정이 **금지**돼 있다(`spec-frontmatter-schema.md § Forbidden ownership crossings`). run-phase에서 이 판단이 필요해지면 manager-develop은 blocker 보고를 반환하고 오케스트레이터가 manager-spec에 재위임해야 한다. 즉 이 지시를 따르면 규율 위반이고, 따르지 않으면 AC가 막힌다.

**질문 3에 대한 판정: 미루는 것은 허용되지 않는다.** 미룰 값이 없기 때문이다 — 두 히트의 성격은 이미 이 트리에서 읽을 수 있고(위 3줄), 저자도 `:106`에서 정확히 그렇게 분류해 놓았다. 판단에 필요한 정보가 전부 모여 있는데 판단만 뒤로 미루면, 뒤에서 그 판단을 내릴 수 있는 행위자가 없다.

**필요한 수정**: plan-phase에서 지금 좁힌다. ACPR 행을 배치 블록 한정 판정으로 재작성한다(예: 배치 코드블록의 `# 1.` ~ `# 2.` 구간만 잘라낸 뒤 그 안에서 카운트 → 기대 0, 사전 baseline 1). `verification-batch-pattern.md` 행은 전체 파일 카운트를 유지한다(사전 baseline 1, 전량 호출이 Group A 행에만 있음 — 이 트리에서 확인). 좁혀도 커버리지는 잃지 않는다: `:22`의 호출은 배치 블록 안에 있으므로 여전히 잡힌다.

---

**D3. 긍정 요구사항에 대응하는 AC가 하나도 없다 — 동의어 재작성 mutant가 전 항목을 통과한다 — `acceptance.md:20-107`, `spec.md:71,77` — Severity: major — Class: blocking**

AC-FSD-001~006은 전부 **부재(absence)** 판정이다. REQ-FSD-001("shall scope run-phase test execution to the packages the change can affect")과 REQ-FSD-004("shall prescribe change-scoped test execution")는 **존재** 요구인데, 이를 재는 AC가 없다. `§D.2` 는 이 둘을 AC-001/002/003/004/006에 매핑하지만 그 AC들은 문자열이 사라졌는지만 본다.

Mutant probe(작성 가능함을 확인): `:92`를 `Step 5 runs the complete suite regardless of project size.` 로 바꾸면 —

- `always runs the full suite` 불일치 → AC-001/002 통과
- `COMPLETE test suite` 는 대소문자 구분 grep이라 `complete suite` 와 불일치 → AC-001/002 통과
- `full suite, coverage` 불일치 → AC-003 통과

전량 실행 지시는 **그대로 살아 있는데** MUST 3항이 전부 초록이다. `§D.1`이 이 세 항목을 "부분 수리 탐지 — 대표 mutant 방어"라고 적은 주장은 이 mutant 앞에서 성립하지 않는다.

**필요한 수정**: 긍정 AC를 추가한다. 사전 0히트가 확보되는 형태여야 한다 — 예컨대 `manager-develop.md` 로컬·템플릿 양쪽에서 변경 범위 스코프를 명명하는 문면의 존재를 세는 grep(현재 트리에서 0히트임을 먼저 실측). REQ-FSD-004에도 동일하게 배치 1번 항목이 스코프 형태 호출을 담고 있음을 재는 긍정 AC가 필요하다.

---

**D4. REQ-FSD-003(dangling conditional 금지)에 기계 판정이 없고, plan.md M1의 권고와 충돌한다 — `spec.md:73`, `plan.md:55`, `acceptance.md:199` — Severity: major — Class: blocking**

**질문 4에 대한 판정: 막지 못한다.**

`§D.2`는 REQ-FSD-003을 "AC-FSD-001, 002 (+ 눈 확인)"에 매핑한다. AC-001/002는 `otherwise the full suite` 문자열의 부재만 증명한다. 그 줄을 통째로 지워도, 문법이 깨진 채 남겨도 똑같이 통과한다. 실제로 `:126`을 재측정하면:

```
3. **Verify behavior**: run tests — targeted when `ddd` LARGE_SCALE, otherwise the full suite (memory guard: ...)
```

여기서 `otherwise the full suite` 만 제거하면 `run tests — targeted when ddd LARGE_SCALE,` 로 끝난다 — 문자 그대로 dangling이다. AC-001은 이것을 초록으로 통과시킨다.

권고의 정합성도 문제다. `plan.md:55`는 "감지 로직 자체를 삭제하는 것은 범위 밖으로 보고, 전량 주장만 제거하는 최소 편집을 권장한다"면서 같은 문장에서 LARGE_SCALE이 "변별력을 잃었으니"라고 적는다. `:92`를 재측정하면 LARGE_SCALE의 유일한 귀결은 `switches PRESERVE/IMPROVE to targeted test execution` 이다. 새 독트린에서는 **모든** 실행이 targeted이므로, 감지를 남기면 아무것도 가르지 않는 판별자가 남는다. `otherwise` 라는 단어가 사라져도 실질은 dangling이다 — REQ-FSD-003이 금지하려던 바로 그것이다.

**필요한 수정**: (a) REQ-FSD-003에 기계 판정 AC를 붙인다. 최소한 `:126`·`:92` 대체 문면을 문자열로 고정해 존재를 재는 형태(사전 0히트 확보 후). (b) plan.md M1.1의 권고를 REQ-FSD-003과 정합시킨다 — 감지 로직을 남길 것인지, 남긴다면 그것이 무엇을 가르는지를 M1에서 문면으로 확정하고, 가르는 것이 없다면 그 문장에서 감지 언급까지 정리한다. 둘 중 어느 쪽이든 plan-phase에서 정해져야 한다(D2와 같은 이유: run-phase 행위자는 acceptance.md를 고칠 수 없다).

---

**D5. REQ-FSD-005의 두 번째 절(시간 추정치)에 대응 AC가 없다 — `spec.md:78`, `acceptance.md:79-89` — Severity: major — Class: blocking**

REQ-FSD-005는 두 가지를 요구한다: (1) Group A 행이 변경 범위 스코프 실행을 명명할 것, (2) "its time estimate shall not assert a figure this SPEC did not measure". `§D.2`는 AC-005·006에 매핑하지만 —

- AC-005의 Then: "그 줄에 `go test ./...` 문자열이 없다" — 시간 추정치 무관.
- AC-006: 전량 호출 문자열 카운트 — 시간 추정치 무관.

재측정한 대상 행(`verification-batch-pattern.md:45`, 로컬·템플릿 동일):

```
| A. Functional | `go test ./...`, coverage | 30-120 s |
```

`30-120 s` 를 그대로 두어도 두 AC 모두 통과한다. REQ-FSD-005의 절반이 무방비다. 이 SPEC이 §D에서 "시간 추정치를 실측으로 대체하는 것은 범위 밖, 근거 없는 수치를 주장하지 않게 다듬는 것까지가 범위"라고 명시적으로 범위 안에 넣은 항목이라 더 그렇다.

**필요한 수정**: Group A 행에서 `30-120 s` 셀이 측정되지 않은 수치를 주장하지 않음을 재는 AC를 추가한다(예: 해당 행에 대한 정규식으로 `[0-9]+-[0-9]+ s` 부재 판정, 사전 baseline 1히트 실측).

---

**D6. AC-FSD-009의 판정 명령이 staged 변경을 보지 못해 공허 통과가 가능하다 — `acceptance.md:159-163` — Severity: major — Class: blocking**

판정 명령은 `git diff -- internal/template/templates/` 의 추가 줄(`^+`)에 패턴을 거는 형태다. 이 형태는 **unstaged** 변경만 본다. 구현자가 `git add` 이후(정상적인 M5 커밋 흐름 직전)에 이 AC를 실행하면 diff가 비고, 파이프 뒤 grep은 무출력이며, "Then 무출력"이 충족된다 — 중립성을 위반하는 줄이 실제로 추가돼 있어도 통과한다. 커밋 이후 실행하면 더욱 그렇다.

`verification-completeness.md §1.1` 각주의 형태 그대로다: 아무것도 훑지 않은 스윕도 초록을 찍는다.

**필요한 수정**: 비교 기준을 `HEAD`(커밋 이후라면 분기 base)로 고정해 staged/committed 변경까지 보게 한다. 추가로 훑은 줄 수를 함께 세어 0줄 스윕이 통과로 읽히지 않게 한다. D1에 따라 패턴 집합에 `origin/develop` 도 추가한다.

---

**D7. `tier:` 프론트매터가 없어 선언된 Tier M이 어떤 기계 소비자에게도 전달되지 않는다 — `spec.md:2-13` vs `plan.md:11`, `progress.md:5` — Severity: major — Class: blocking**

`grep -n '^tier:' spec.md` → 매치 없음(rc=1). plan.md와 progress.md는 "Tier: M"을 선언한다.

`spec-workflow.md § SPEC Complexity Tier`: "`tier:` 부재 시 Tier L로 취급(하위 호환)". 결과적으로 —

- plan-auditor PASS 기준선: 선언 0.80 → 기계 해석 0.85
- 재감사 상한: 선언 2회 → 기계 해석 3회
- 요구 산출물 집합: 3종 → 5종(`design.md`·`research.md` 요구, 둘 다 부재)
- `internal/runtime/audit_cache.go` `ComputeHash` 주체 집합 해석이 달라진다

이 감사 자체가 이 모호성의 영향을 받는다 — 어느 기준선으로 재는지가 프론트매터 한 줄에 달려 있다. (이번 판정은 0.571이라 두 기준선 어느 쪽으로도 미달이므로 결론은 바뀌지 않지만, 그것은 우연이지 설계가 아니다.)

**필요한 수정**: `spec.md` 프론트매터에 `tier: M` 을 추가한다.

---

**D8. AC-FSD-005와 AC-FSD-009 첫 블록이 report-not-verdict 형태다 — `acceptance.md:84-89`, `:146-154` — Severity: minor — Class: blocking**

AC-FSD-005의 명령은 `grep -n 'A. Functional'` 로 해당 행을 **출력**할 뿐인데 Then은 "그 줄에 `go test ./...` 문자열이 없다"를 요구한다. 명령의 종료 코드는 행이 존재하기만 하면 0이며, Then이 묻는 것과 무관하다. 판정은 "눈으로 확인"과 AC-006에 위임된다 — 그런데 AC-006은 D2에 따라 좁혀질 예정이므로, 좁힘 범위가 `verification-batch-pattern.md` 까지 포함되면 이 행의 기계 판정자가 사라진다.

AC-FSD-009 첫 블록도 같다: 명령은 파일 전체를 grep하고 `rc` 를 찍는데, Then은 "이번 수리가 **추가한** 줄 중에 매치가 없다"이다. 명령이 그 명제를 만들지 않는다.

**필요한 수정**: AC-005를 자기완결 판정으로 재작성한다(해당 행만 추출한 뒤 그 안에서 전량 호출 문자열을 카운트 → 0). AC-009는 첫 블록을 삭제하고 두 번째 블록(수정된 형태, D6)만 남긴다.

---

**D9. AC-FSD-007의 패턴이 VCI 방어물로 쓰기에 너무 얕다 — `acceptance.md:115-121` — Severity: major — Class: blocking**

**질문 5에 대한 판정: 충분히 강하지 않다.**

패턴은 `grep -c -i -e 'delegated' -e 'pending'` 이다. 사전 baseline은 실측 `0`(두 파일 모두) — RED-now는 올바르게 확보돼 있다. 이 점은 이 SPEC의 강점이다.

그러나 사후 `≥1` 이 증명하는 것은 "두 단어 중 하나가 파일 어딘가에 생겼다"뿐이다. 대소문자 무시 `pending` 은 산문에서 흔한 단어이고, 이 SPEC 밖의 후속 편집으로도 우연히 만족될 수 있다. AC가 실제로 요구해야 하는 세 가지 —

1. 그 문장이 **STEP 5 완료 보고 지시 안에** 있을 것
2. 판정 **주체**를 명명할 것
3. 보고 시점 상태가 **미결**임을 말할 것

— 는 전부 "눈 확인"에 남아 있다. 이것이 B3(조용한 삭제)를 막는 **유일한** 장치라고 `plan.md:56`이 적었는데, 그 유일한 장치가 이진 판정이 아니다.

부수적으로: 재측정한 STEP 5 블록(`manager-develop.md:130-137`)에는 "Generate the completion report" 항목이 존재하므로(`:136`), 문장을 심을 자리는 실재한다. 앵커는 있다.

**필요한 수정**: 위치를 고정하는 판정으로 바꾼다 — STEP 5 블록을 `awk` 로 잘라낸 뒤 그 안에서만 세거나, 대체 문면을 문자열로 고정해 존재를 잰다. 판정 주체 명명은 D1의 중립화 결정과 함께 확정한 뒤 그 문자열로 재야 한다(브랜치명이 아닌 중립 문면).

---

**D10. AC-FSD-011은 결함이 아니다 — 정직하게 서술돼 있다 — `acceptance.md:171-178` — Severity: (없음) — Class: (해당 없음)**

**질문 7에 대한 판정: 정직하고, 지연 처리가 옳다.**

`:177`은 49분 40초가 "리드에게서 물려받은 주장이며 이 SPEC의 어떤 세션에서도 재측정하지 않았다"고 명시하고, 재측정하려면 C4가 금지한 행위가 필요하다는 사실까지 적는다. `:178`은 "사후 값만 실측이며 사전 값은 인용" — 대조 실험이 아님을 스스로 선언하고 "개선폭을 정량 주장으로 내세우지 않는다"고 못박는다. `§D.1`은 SHOULD(지연), `§D.3`은 "미결을 통과로 적지 않는다".

이것은 회피가 아니라 `verification-claim-integrity.md §2`가 요구하는 귀속 규율의 모범 사례다. 물려받은 수치를 baseline인 척 쓰지 않았고, 측정할 수 없는 이유를 명시했고, 미결을 미결로 남겼다. 유일한 부수 효과는 REQ-FSD-010이 이 SPEC 종결 시점에 미검증으로 남는다는 것인데, 그것도 `plan.md:82`에 선언돼 있다.

---

**D11. 배포 대체 문면의 프로그래밍 언어 중립성이 다뤄지지 않았다 — `spec.md:77`, `spec.md:124-128` — Severity: minor — Class: optional**

REQ-FSD-004는 "rather than an unconditional `go test ./...` invocation" 이라고 적는다. 수리 대상 3개 파일은 전부 배포 템플릿이고, 배포 사용자는 16개 프로그래밍 언어를 쓴다. 대체 문면이 Go 전용 형태로 굳으면 기존 Go 편향이 그대로 이월된다.

다만 이 편향은 **이 SPEC이 만든 것이 아니라** 이미 그 파일들에 있던 것이고(`# 1. Full test suite (Go)`), `spec.md §D`는 "이 4개 파일 밖의 전량-스위트 지시 전수 조사는 하지 않는다"로 범위를 좁혔다. 범위 규율상 이 SPEC에 언어 중립화를 강제하는 것은 과잉이다. optional로 분류한다 — 오케스트레이터 재량. 다루기로 한다면 M1의 문면 확정에서 한 문장으로 흡수 가능하다.

---

**D12. 이 SPEC의 REQ 표기가 `moai spec lint` 의 REQ 파서에 보이지 않는다 — `spec.md:71-92` — Severity: minor — Class: optional**

`internal/spec/lint.go:447` `reqLinePattern` 은 `- REQ-XXX-NNN-NNN: text` 형태만 매치한다. 이 SPEC의 항목은 `- **REQ-FSD-001** (Ubiquitous) — The …` 이므로 `doc.REQs` 는 빈 슬라이스가 되고, `:646`의 REQ ID 검사는 아예 실행되지 않는다.

이 SPEC의 결함이라기보다 리포 전반의 관례와 린트 파서 사이 드리프트다(현대 SPEC 다수가 `REQ-XXX-NNN` 3자리 단일형을 쓴다). 다만 **결과는 이 SPEC에 실재한다**: REQ/AC 관련 기계 린트가 이 SPEC에 대해 공허하게 통과하므로, D3·D5가 지적한 커버리지 공백을 잡아줄 안전망이 없다. 위 수동 감사가 유일한 검사였다는 사실을 기록해 둔다.

---

## 질문별 요약 판정

| # | 질문 | 판정 |
|---|---|---|
| 1 | 부분 수리 탐지 | **부분적으로만.** 4개 지점 각각에 전용 패턴이 있어 "한 곳만 고치기" 형태는 전부 잡힌다(AC-001이 S1·S2·S3, AC-003이 S4). 로컬/템플릿 비대칭도 AC-001+002+008이 잡는다. 그러나 **동의어 재작성 mutant**는 전부 통과한다 — D3 |
| 2 | 0히트 baseline 유효성 | **전 항목 실측 확인.** 부재형 AC(001~006)의 사전 baseline이 전부 비영(아래 표). 유일한 0 baseline은 AC-007인데 그것은 **존재형** AC이므로 0이 올바른 RED-now다. 형식 불량 패턴 없음. **다만 AC-006은 baseline이 아니라 사후 기대치가 틀렸다** — D2 |
| 3 | AC-FSD-006 과잉 차단 | **미루면 안 된다.** 판단 재료가 이미 다 있고(`:65`·`:76`은 예시 산문), 미룬 판단을 내릴 수 있는 행위자가 run-phase에 없다(소유권 교차 금지) — D2 |
| 4 | dangling conditional | **REQ-FSD-003은 막지 못한다.** 기계 판정 부재 + plan.md M1 권고가 "변별력 잃은 판별자를 남긴다"는 형태로 REQ와 충돌 — D4 |
| 5 | 대체 문면의 VCI 적합성 | **충분히 강하지 않다.** AC-007은 단어 2개 존재만 잰다. 위치·주체·미결 상태는 전부 눈 확인 — D9. 게다가 REQ-FSD-006이 지정한 주체 문면 자체가 중립성 위반 — D1 |
| 6 | 템플릿 중립성 | **올바르게 강제하지 못한다.** 판정 명령이 staged 변경을 못 본다(D6) + 이 SPEC이 실제로 도입할 지역 내용(`origin/develop`)을 패턴 집합이 잡지 못한다(D1) |
| 7 | 지연 AC 정직성 | **정직하고 지연이 옳다** — D10 |

### baseline 실측표 (이 트리, HEAD `d29b8942e`)

| AC | 명령 요지 | SPEC 기재 baseline | 실측 | 일치 |
|---|---|---|---|---|
| AC-FSD-001 | 4패턴 `grep -c` (로컬 agent) | 3 | 3 | ✓ |
| AC-FSD-002 | 4패턴 `grep -c` (템플릿 agent) | 3 | 3 | ✓ |
| AC-FSD-003 | `full suite, coverage` | 각 1 | 각 1 | ✓ |
| AC-FSD-004 | `Full test suite` | 각 1, rc=0 | 각 1, rc=0 | ✓ |
| AC-FSD-005 | `A. Functional` | (미기재) | 각 1행 존재 | — |
| AC-FSD-006 | 전량 호출 문자열 카운트 | ACPR 각 3 / VBP 각 1 | ACPR 각 3 / VBP 각 1 | ✓ |
| AC-FSD-007 | `-i delegated\|pending` | 각 0 | 각 0 | ✓ |
| AC-FSD-008 | 3쌍 `diff` | 1줄 / 0 / 0 | `1a2 > isolation: worktree` / 무출력 / 무출력 | ✓ |

`spec.md §A.1` 의 4개 지점 좌표도 재측정해 전부 일치 확인: 로컬 `:92`·`:126`·`:132`·`:135`, 템플릿 `:93`·`:127`·`:133`·`:136`(일괄 +1). `spec.md §A.2` 의 "로컬·템플릿 byte 동일" 주장도 두 쌍 모두 `diff` rc=0 으로 확인. 상위 계약 인용 좌표도 재확인: `AGENTS.md:117-118`, `kanban-dispatch.md:195`.

---

## 권고 (manager-spec 재작업 항목, 우선순위 순)

1. **REQ-FSD-006에서 `origin/develop` 제거**, 판정 주체를 브랜치 중립 문면으로 재작성. 지역 사실은 `spec.md §A.3` 배경에만 남긴다. AC-FSD-009 패턴 집합에 `origin/develop` 추가. (D1)
2. **AC-FSD-006을 지금 좁힌다** — ACPR은 배치 블록 한정 판정으로, VBP는 전체 파일 카운트 유지. run-phase 위임 문장을 삭제한다. (D2)
3. **긍정 AC 2개 추가** — REQ-FSD-001(변경 범위 스코프 문면 존재), REQ-FSD-004(배치 1번 항목의 스코프 형태 호출 존재). 각각 이 트리에서 사전 0히트를 실측해 함께 기재. (D3)
4. **REQ-FSD-003에 기계 판정 AC 부여** + `plan.md:55` 권고를 REQ와 정합시킨다. LARGE_SCALE 감지의 존치 여부를 M1이 아니라 SPEC 문면에서 결정한다. (D4)
5. **REQ-FSD-005 시간 추정치 절에 AC 부여.** (D5)
6. **AC-FSD-009 판정 명령의 비교 기준을 `HEAD` 로 고정**하고 첫 블록 삭제. (D6, D8)
7. **`spec.md` 프론트매터에 `tier: M` 추가.** (D7)
8. **AC-FSD-007을 위치 고정 판정으로 재작성** — STEP 5 블록 범위 안에서만 판정. (D9, D8)
9. (optional) 대체 문면의 프로그래밍 언어 중립성 판단을 M1에 흡수할지 결정. (D11)

수정 후 재감사는 위 열거된 결함 델타로 범위를 좁혀 수행한다(iteration 2). Tier M 선언 기준 상한은 2회이므로, iteration 2에서도 FAIL이면 scope-reduction 또는 PASS-with-debt를 운영자에게 제시해야 한다.

---

## Gaps (관측하지 않은 것)

- 대체 문면 자체를 읽지 않았다 — 아직 존재하지 않는다(`plan.md` M1이 run-phase 산출로 남겨둠). D1·D3·D9의 판정은 **요구사항과 AC의 문면**에 대한 것이지 아직 쓰이지 않은 대체 텍스트에 대한 것이 아니다.
- 전체 테스트 스위트를 포함해 어떤 테스트도 실행하지 않았다(지시 제약 및 `spec.md §C` C4).
- `.claude/skills/` 층의 같은 형태 지시(`spec.md §D`가 범위 밖으로 선언)는 세지 않았다. 그 선언이 옳은지는 판정하지 않았다.
- 어떤 SPEC 산출물도 수정하지 않았다.

**감사 산출물 경로**: `/Users/goos/MoAI/moai-adk-go/.moai/reports/t301/plan-audit.md`
**측정 트리**: `.claude/worktrees/t301` @ `d29b8942e`
