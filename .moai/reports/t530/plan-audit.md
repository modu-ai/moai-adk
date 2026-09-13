# SPEC Review Report: SPEC-DOCS-TABCOUNT-DRIFT-001 (카드 t530)

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.66** (Tier M PASS 기준 0.80)

측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, HEAD `c205eeeff`, base `1d150a27d`
(`origin/develop` 는 이 시점에 `1d150a27d` 와 동일)
작성자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정은 아티팩트와 트리 실측만으로 내렸다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 정합** — `spec.md` L89~L140 에 REQ-TCD-001 ~ 009 가 빠짐·중복 없이 3자리 패딩으로 연속한다. `grep -c '^### REQ-'` = `9`.
- **[PASS] MP-2 GEARS 준수 (요구층)** — 9개 REQ 전부가 다섯 GEARS 패턴 중 하나에 든다: Ubiquitous(001/002/006), Where(003/004), Event-driven(005), State-driven(007), Unwanted(008), Unwanted+Ubiquitous 복합(009). 예: L126-129 `While any one of the ko / en / ja / zh copies of a page is changed, the other three copies of that page shall be changed in the same commit.` H3 제목이 한국어인 것은 표제일 뿐 규범 문장은 영문 GEARS 다. **판정 층 명시**: 이 판정은 `spec.md` 의 `REQ-XXX` 요구층에만 적용했다. `acceptance.md` 의 Given-When-Then 은 검증층의 정상 형식이므로 여기서 감점하지 않았다(Group 4 에서 별도 채점).
- **[PASS] MP-3 YAML frontmatter** — `spec.md` L2~L14 에 12개 정본 필드가 모두 있고 타입이 맞는다(`id`/`title`/`version:"0.1.0"`/`status: draft`/`created`/`updated`/`author`/`priority: P2`/`phase:"v3.1.4 target"`/`module`/`lifecycle`/`tags`). snake_case 별칭 없음.
- **[N/A] MP-4 언어 중립성** — 대상이 README·docs-site 문서와 `internal/web` Go 테스트 하나뿐인 단일 언어 범위다. 16개 프로그래밍 언어 도구 얘기가 아니므로 해당 없음.
- **[PASS] MP-5 D7 교차 SPEC** — 본문이 참조하는 `SPEC-WEB-CODEX-PANEL-001`, `SPEC-PRECOMMIT-GATE-SCOPE-001` 둘 다 `.moai/specs/` 에 존재하고 `status: completed` 다. retired/superseded/archived 없음 ⇒ BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -c 'syscall' spec.md` = `0`. 자동 통과.
- **[FAIL] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md` 가 `plan.md:106` 에서 적중한다. 문장 자체는 "이 결정으로 해소되었고 남겨두지 않는다" 는 **해소 기록**이지 미결 마커가 아니다. 그러나 MP-7 의 검증 동사는 grep 이고, grep 은 의미를 읽지 않는다. 게이트가 기계적으로 붉다.
  **필요 수정**: L106 의 리터럴 대괄호 토큰을 없앤다 — 예) `` `[NEEDS CLARIFICATION: llm-tab-label]` 마커는 `` → `llm-tab-label 미결 항목은`.

---

## Category Scores

| 차원 | 점수 | 루브릭 대역 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ 문장은 단일 해석이나, REQ-TCD-005 의 "인접(adjacent)" 이 요구층에서 수치화되지 않았고 그 조작적 정의(`[^.。]{0,4}`)가 `acceptance.md` 에만 있으며 실측으로 틀렸다(D1) |
| Completeness | 0.80 | 0.75~1.0 경계 | 필수 섹션 전부 존재, `### Out of Scope — <주제>` H3 5개 각각 `-` 불릿 보유(spec.md L194~217). 감점: 형제 아티팩트 statelessness 위반(D7), 열거 산출물의 산문 4자리가 행으로 세어지지 않음(D6) |
| **Testability** | **0.35** | 0.25~0.50 | 10개 AC 중 **6개가 결함** — 대상 자리를 놓치거나(D1), 도달 불가능하거나(D2), 공허한 초록이거나(D4, D5), 변이체를 통과시킨다(D3) |
| Traceability | 0.75 | 0.75 | REQ-TCD-001~009 전부 AC 보유. 단 AC-TCD-010 은 §D 매트릭스가 스스로 검증 대상 REQ 를 `—` 로 적은 **고아 AC** 다(acceptance.md L34) |

가중 평균 (0.75+0.80+0.35+0.75)/4 = **0.6625 → 0.66**. Tier M 기준 0.80 미달.

---

## Defects Found

### D1. AC-TCD-002 가 A군 4자리 중 영어 1자리를 **구조적으로 못 본다** — 이 카드의 대표 결함이 자기 인수 기준을 빠져나간다

- **위치**: `acceptance.md:49-51` (정규식), 대상 `docs-site/content/en/cli-reference/web.md:53`
- **Severity: critical — Class: blocking**
- **무엇이 틀렸나**: AC-TCD-002 의 정규식은 수사와 탭 명사 사이에 **최대 4글자**만 허용한다(`[^.。]{0,4}`). 영어 A2 자리의 실물은

  ```
  | `/settings` | GET | The nine settings tabs. `?tab=` selects the tab, ... |
  ```

  `nine` 과 `tabs` 사이는 ` settings ` = **10글자**다. 실측:

  ```
  $ grep -nE '(9|nine|九|아홉)[^.。]{0,4}(tabs?|개 탭|タブ|个标签页)' docs-site/content/{ko,en,ja,zh}/cli-reference/web.md
  zh/cli-reference/web.md:53: ...设置九个标签页...
  ko/cli-reference/web.md:53: ...설정 9개 탭...
  ja/cli-reference/web.md:53: ...設定 9 タブ...
  (en 없음)
  ```

  즉 **ko/ja/zh 세 자리만 고치고 영어의 틀린 `nine` 을 그대로 두어도 AC-TCD-002 는 출력 없이 `rc=1` 로 PASS 한다.** 이것은 억지 변이체가 아니라 M2 를 성실히 수행해도 일어나는 기본 경로다. REQ-TCD-002(네 자리 전부)와 REQ-TCD-007(4로케일 동시성)을 검증층에서 동시에 무력화한다.
- **더 넓은 함의**: `plan.md §A.3` 은 이 간격을 **더 좁히겠다**고 적었다("사이에 허용하는 글자는 소수(공백·개·-·の·个·가지 정도)로 좁힌다"). 그대로 구현하면 영어 누락은 악화된다. 낱말 오탐이 두려워 간격을 좁히면 수식어가 낀 영어 어순을 통째로 잃는다 — 이 카드가 없애려는 결함(같은 사실을 자리마다 따로 세다 어긋남)을 한 층 위에서 그대로 재생산한 것이다.
- **필요 수정**: (a) 수사-명사 인접 창을 어순 수식어가 들어갈 만큼 넓히고, 대신 **탭 명사 쪽을 `settings tabs`/`설정 탭` 같은 한정어로 좁혀** 오탐을 억제한다. (b) 역순 표기(`탭 14개`, `标签页共 14 个`)를 정규식이 아예 보지 않는다는 점도 같이 닫는다. (c) 수정한 정규식이 A1~A4 **네 자리 전부**를 적중함을 M2 이전에 보이고 그 출력을 `progress.md §E.2` 에 남긴다.

### D2. AC-TCD-003 의 기대값 `0건` 은 현 범위 선언과 **양립 불가능**하다 — 미열거 오탐이 화이트리스트 안에 있다

- **위치**: `acceptance.md:56-71`, 적중 대상 `docs-site/content/ja/advanced/moai-web-console.md:149`
- **Severity: critical — Class: blocking**
- **무엇이 틀렸나**: AC-TCD-003 의 스윕을 이 트리에서 실행하면 20행이 나온다. 그런데 **그 20행의 집합이 열거 산출물의 20자리와 다르다.** 실측 차분:
  - 스윕에 **없는** 열거 자리: `en/cli-reference/web.md:53` (A2 — D1 의 원인)
  - 열거에 **없는** 스윕 적중: `ja/advanced/moai-web-console.md:149`

  ```
  $ sed -n '149p' docs-site/content/ja/advanced/moai-web-console.md
  codex の設定はもともと監査タブと MCP タブに分かれていました。…「2 つのタブを行き来し」…
  ```

  이 줄은 **codex 패널의 존재 이유를 설명하는 산문**이며, 세는 대상은 설정 탭 총수가 아니라 "감사 탭과 MCP 탭 둘"이다. `spec.md §4`·`§6` 이 codex 패널 반경을 명시적으로 범위 밖으로 선언했으므로, 이 줄은 **고칠 수 없고 고쳐서도 안 되는데 AC 는 0건을 요구한다.** ko/en/zh 대응 줄에는 수사가 없어 ja 단독 자리다(실측: 다른 세 로케일 무적중).
- **왜 red-now 가 아니라 red-forever 인가**: AC-TCD-003 은 지금 붉고, M3 가 20자리를 지워도 이 한 줄 때문에 **여전히 붉다**. `acceptance.md:69-71` 의 RED-now 주석은 "20행이 나오며 M3 가 그것을 0으로 만든다"고 적었는데, 그 20행 안에 M3 가 건드릴 수 없는 줄이 섞여 있다는 사실을 모른다. `verification-completeness.md §2` 용어로 **wrong-reason red** 이며, 초기 관측이 붉다는 이유로 impossible 방향이 그대로 통과한 사례다.
- **가드 판별력에 대한 판정**: `plan.md §A.4` 는 두 장치(파일 화이트리스트 + 수·탭명사 인접)를 함께 두면 오탐이 잡힌다고 주장한다. ja:149 는 **두 장치가 동시에 실패하는 반례**다 — 화이트리스트 안의 파일이고, 수사와 탭 명사가 붙어 있다. 지명된 오탐 3자리(`README.md:422` `nine`=1, `README.md:418` `Eleven ref skills`=1, `plugins.md:100` `4 タブ`=1)는 실측상 모두 제대로 배제되지만, 그것은 두 장치가 **다루기 쉬운 방향**만 막는다는 뜻이다.
- **`spec.md §7` 잔여 위험 고백에 대한 판정**: 정직하되 **불완전**하다. §7 이 인정한 잔여는 "화이트리스트 **밖** 새 문서"라는 한 방향뿐이고, 실제로 이 카드를 막을 방향 — "화이트리스트 **안**의 정당한 비-설정탭 계수" — 은 언급이 없다. 숨긴 것은 아니나 더 큰 구멍을 보지 못했다.
- **필요 수정**: 셋 중 하나를 골라 명시한다. (a) ja:149 를 열거 §7 오탐표에 추가하고 AC-TCD-003 의 기대값을 `0건` → `허용 목록과 정확히 일치(1건, ja:149)` 로 바꾼다 — 그러면 `plan.md §B` 의 "20자리 전부 removable ⇒ 허용 목록 없음" 결론도 함께 정정해야 한다. (b) 정규식에 설정탭 한정어를 넣어(D1 수정과 같은 축) ja:149 를 구조적으로 배제한다. (c) ja:149 의 범위 밖 선언을 철회하고 고친다(권장하지 않음 — t509 경계를 다시 다투게 된다).

### D3. AC-TCD-006 / AC-TCD-008 을 통과하면서 REQ-TCD-004 / 006 을 위반하는 변이체가 **작성 가능**하다

- **위치**: `acceptance.md:93-98`(006), `acceptance.md:112-117`(008)
- **Severity: major — Class: blocking**
- **변이체 A (AC-006 대상)**: `README.md` 의 이름 목록만 읽고 나머지 11개 파일을 전혀 스캔하지 않는 가드. AC-TCD-006 의 변이 시험은 **`README.md` 한 자리(`Audit`→`Audits`)** 만 깨뜨리므로 이 가드도 FAIL→PASS 를 정상 재현한다. AC-TCD-004(PASS 행 존재)·AC-TCD-005(오탐 침묵)·AC-TCD-008 도 모두 통과한다. 그러나 REQ-TCD-004 가 요구하는 "REQ-TCD-003 이후 남은 수·이름 자리" 의 대다수(`advanced/moai-web-console.md` 8자리 + 로케일 3본 README)는 지켜지지 않는다.
- **변이체 B (AC-008 대상)**: `TestDocsTabContract` 에 `names` 서브테스트를 **만들지 않는다**. `go test -run 'TestDocsTabContract/names' -v` 는 부모 테스트를 실행하고 `--- PASS: TestDocsTabContract` 를 출력한다. AC-TCD-008 의 Then 절은 "`--- PASS` 가 나오고 FAIL 이 없다" 이므로 **서브테스트가 존재하지 않는 채로 통과한다.** AC 가 스스로 경계한 빈 선택자 함정(`acceptance.md:78-79`)이 008 에서는 닫히지 않았다.
- **필요 수정**: (a) AC-TCD-006 의 변이 지점을 **파일 부류 × 로케일** 로 넓힌다 — 최소 `README.<locale>.md` 1자리 + `advanced/moai-web-console.md` 1자리, 서로 다른 로케일. (b) 가드가 실제로 읽은 파일 수를 출력하고 그것이 선언된 12와 같은지 단정하는 AC 를 추가한다(쓸어담은 집합의 크기를 판정 전에 센다). (c) AC-TCD-008 의 Then 을 리터럴 `--- PASS: TestDocsTabContract/names` 존재로 바꾼다.

### D4. AC-TCD-007 은 지금 공허하게 초록이고, README 4본을 **전혀 보지 않으며**, 움직이는 ref 로 잰다

- **위치**: `acceptance.md:100-110`
- **Severity: major — Class: blocking**
- **세 결함이 겹친다**:
  1. **공허**: 실측 `git diff --name-only origin/develop... -- docs-site/content | wc -l` = `0`. 상대 경로 종류 0 × 4 == 0 이므로 **손대지 않은 트리에서 이미 통과한다.**
  2. **범위 누락**: 판정 대상이 `-- docs-site/content` 뿐이다. C군 8자리(README ×4 :414, :750)와 D군 README 4자리는 **이 AC 밖**이다. "한 로케일만 고치는 것은 새 드리프트" 라는 REQ-TCD-007 의 핵심 주장이 32자리 중 12자리에 대해 기계적으로 검사되지 않는다.
  3. **움직이는 ref**: `origin/develop...` 는 고정 SHA 가 아니다. 이 저장소는 리드가 develop 을 일괄 push 하므로 카드 수명 중 반드시 전진한다. 전진하면 남의 변경이 이 AC 의 분모에 섞여 판정이 스스로 뒤집힌다.
- **필요 수정**: `README.md README.ko.md README.ja.md README.zh.md` 를 판정 대상에 넣고, 기준을 읽는 시점의 `git merge-base origin/develop HEAD` 또는 고정 base SHA `1d150a27d` 로 바꾼다. 그리고 **대조군을 붙인다** — 변경 파일 수가 0 이면 "패리티 유지" 가 아니라 "측정 불가" 로 보고한다.

### D5. AC-TCD-009 / AC-TCD-010 은 도착 시점에 이미 초록인 회귀 가드이지 인수 기준이 아니다

- **위치**: `acceptance.md:119-134`
- **Severity: minor — Class: blocking** (분류 표기 문제라 수정은 가볍다)
- **AC-TCD-009**: 실측 `git diff --name-only origin/develop... -- assets/images/` 출력 없음, 지금 통과한다. 뒤에 붙인 `echo "rc=$?"` 는 `git diff` 의 종료코드(성공=0)를 찍을 뿐 무매치를 뜻하지 않는다 — 판정 근거가 되지 못하는 잉여 필드다. D4 와 같은 움직이는-ref 문제도 함께 안는다.
- **AC-TCD-010**: 실측 `cd docs-site && hugo --gc --minify 2>&1 | grep -iE 'warn|error'` → 출력 없음, `rc=1`. **손대지 않은 트리에서 이미 통과한다.**
- **필요 수정**: 두 AC 를 **회귀 가드**로 명시 분류하고, "이 카드의 작업이 무엇을 초록으로 바꾸는가" 를 요구하는 목록과 섞지 않는다. AC-TCD-009 의 `echo "rc=$?"` 는 제거하거나 `git diff --quiet --exit-code` 로 바꿔 종료코드가 실제 판정을 운반하게 한다. 기준 ref 를 `1d150a27d` 로 고정한다.

### D6. AC-TCD-001 은 열거의 완전성을 세지 못한다 — 산문 4자리를 지워도 통과한다

- **위치**: `acceptance.md:41-42`, `tab-count-sites.md:108-110`
- **Severity: minor — Class: optional**
- **무엇이 틀렸나**: AC-TCD-001 은 `grep -c '^| [ABCD][0-9]'` = 28 이상을 요구한다(실측 `28`). 그런데 D군의 산문 4자리(`advanced/moai-web-console.md:165` × 4로케일)는 표 행이 아니라 §6 말미 산문으로만 기록돼 이 계수에 들지 않는다. `acceptance.md §D.2` 4항과 `plan.md` M4 는 "D군 12자리" 를 요구하는데(실측 12행 확인), AC 는 8자리만 센다. **§6 말미 문단을 통째로 지워도 AC-TCD-001 은 28 로 통과한다** — REQ-TCD-001 의 exhaustive 를 검증하지 못한다.
- **필요 수정**: 산문 4자리를 D9~D12 표 행으로 승격하고 기대값을 32 로 올린다. 그러면 REQ-TCD-001 이 요구한 로케일 컬럼도 함께 갖춰진다.

### D7. `plan.md` / `acceptance.md` 가 `status:` 를 들고 있다 — 아티팩트 statelessness 위반

- **위치**: `plan.md:5`, `acceptance.md:5` (둘 다 `status: draft`)
- **Severity: minor — Class: blocking**
- **무엇이 틀렸나**: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness 는 `plan.md`·`acceptance.md` 가 status 축에서 무상태여야 한다고 정한다 — SPEC 의 생애 상태는 `spec.md` 한 곳에만 산다. 두 파일이 사본을 들고 있으면 sync 단계에서 `spec.md` 만 전이될 때 조용히 갈라진다.
- **필요 수정**: 두 파일 frontmatter 에서 `status:` 줄 제거. 나머지 필드는 건드리지 않는다.

### D8. "서로 독립인 세 측정" 은 두 측정과 그 둘의 비교다

- **위치**: `spec.md:45-51`, `tab-count-sites.md:14-20`
- **Severity: minor — Class: optional**
- **무엇이 틀렸나**: 측정 1(`schemaform.go` 의 `LabelKey:` 계수 = **14**, 재현 확인)과 측정 2(`tab_layout_test.go` 의 `wantTabOrder` 리터럴 id **14개**, 재현 확인)는 서로 다른 파일이지만, 측정 3(`TestConsoleTabsOrder`, 재현 확인 `ok`)은 **측정 1과 측정 2가 같은지를 비교하는 테스트 그 자체**다(`tab_layout_test.go:32-37` 이 `len(tabs) != len(wantTabOrder)` 를 단정한다). 독립한 세 번째 관측이 아니라 앞 둘의 함수다. 더구나 `wantTabOrder` 는 `consoleTabs()` 를 미러하도록 작성된 리터럴이라 측정 2의 독립성도 약하다. 실질적으로 원천 1개 + 미러 1개다.
- **다만 이 축의 본질은 통과한다**: SPEC 은 14를 상수로 박지 않는다. `spec.md:53-55` 가 명시적으로 "14 는 이 base 에서의 실측치이지 상수가 아니다" 라고 적고, 재도출 명령을 §1.1 표와 `tab-count-sites.md` § 재현 명령에 남겼다. 다음 탭 추가가 이 문서를 거짓으로 만들지 않는다.
- **필요 수정**: "서로 독립인 세 측정" → "두 원천(구현·테스트 리터럴)과 그 둘의 일치를 확인하는 테스트" 로 문구를 낮춘다.

### D9. 낱말 표기 커버리지가 현재 탭 수 부근에 고정돼 있다

- **위치**: `acceptance.md:61`, `plan.md:48`
- **Severity: minor — Class: optional**
- **무엇이 틀렸나**: 수사 대안열이 `nine|ten|eleven|twelve|thirteen|fourteen|fifteen|九|十四|열네|아홉` 이다. 탭이 16개가 되면 `sixteen`·`열여섯`·`十六` 중 아무것도 잡히지 않는다. 이 SPEC 자신이 11→14 드리프트를 사례로 들었으므로, 창을 현재 값 주변에 고정하는 것은 같은 실패를 예약하는 일이다.
- **필요 수정**: 숫자는 `[0-9]+` 로 이미 열려 있으니 낱말 축도 로케일별 수사 문자류로 열거나, 최소한 대안열을 테스트 안의 표로 고정하고 "탭 수가 바뀌면 이 표를 갱신한다" 는 의무를 가드 실패 메시지가 말하게 한다.

### D10. AC-TCD-005 의 오탐 검사는 자리를 고정하지 않는다

- **위치**: `acceptance.md:86`
- **Severity: minor — Class: optional**
- **무엇이 틀렸나**: `grep -c 'nine' README.md` >= 1 은 `README.md` 어디든 `nine` 이 하나 있으면 만족한다. 지켜야 할 자리(`:422` 의 SVG 형태 수)가 사라지고 엉뚱한 곳에 `nine` 이 생겨도 통과한다.
- **필요 수정**: 행 번호 또는 구절(`nine forms`)까지 단정한다.

### D11. 열거 산출물의 최신성을 요구하는 DoD 3항에 검증 동사가 없다

- **위치**: `acceptance.md:148` (§D.2 3항)
- **Severity: minor — Class: optional**
- **무엇이 틀렸나**: 본 카드는 편집으로 행 번호를 움직인다. 그런데 이 의무에는 명령도 기대 출력도 없다 — 산문 단정이고, `acceptance.md:18` 이 스스로 "산문 단정은 인수 근거가 아니다" 라고 적었다.
- **필요 수정**: 열거표의 각 `파일:행` 이 실재하고 표 내용과 맞는지 훑는 한 줄 명령을 AC 로 세우거나, 3항을 DoD 에서 내린다.

---

## 잘 된 부분 (근거 포함 — 판정을 상쇄하지 않는다)

- **기준값 재도출**: 카드가 들고 온 수를 옮기지 않고 이 트리에서 다시 쟀고, 명령이 남아 있어 그대로 재현했다(`14`, `14`, `ok`). 14를 상수로 박지 않은 것이 이 SPEC 의 중심 미덕이다(`spec.md:53-55`).
- **§1.2 이름 드리프트 논증**: 성립한다. 20자리 전부 removable 이므로 수를 지우면 정보는 이름이 진다는 연역이 맞고(`plan.md §B`), D군 12자리가 그 이동의 실물이라는 관측도 실측으로 확인했다(12행). 범위 확대의 합리화가 아니다.
- **스크린샷 배제**: 세 근거(산출물 종류 차이·절차 부재가 별개 결함·가시성)가 구체적이고, REQ-TCD-009 + AC-TCD-009 로 "안 건드렸음" 이 기계 검증 가능하며, 후속 카드 2장을 명시했다. 정당한 배제다.
- **llm 탭 이름 갈래 1 결정**: 근거가 실측으로 뒷받침된다 — `i18n.js:229/1096/1852/2608` 의 4로케일 값(`GLM Settings`/`GLM 설정`/`GLM設定`/`GLM设置`)과 `schemaform.go` 의 `Baseline: "GLM Settings"` 가 일치하므로 "코드 안에서 무엇이 현재 의도인가" 에 모호함이 없다는 `plan.md §C` 진술을 확인했다. 반대 갈래 처리를 `spec.md §7` 에 남긴 것도 적절하다.
- **빈 선택자 함정 인지**: `acceptance.md:78-79` 가 `-run` 이 없는 이름에 대해 조용히 `ok` 를 낸다는 사실을 알고 AC-TCD-004 에 `--- PASS` 행 존재 확인을 붙였다(같은 방어가 008 에는 불완전 — D3).

---

## Recommendation

Tier M 기준 0.80 에 0.66 으로 미달하고, MP-7 이 기계적으로 붉다. **FAIL.** 아래 순서로 고친다 — 1~4·5·6이 blocking, 나머지는 리드 재량.

1. **D1** — AC-TCD-002 정규식이 `en/cli-reference/web.md:53` 의 `The nine settings tabs` 를 적중하게 고치고, A1~A4 네 자리 전부 적중을 M2 이전에 출력으로 보인다. `plan.md §A.3` 의 "간격을 더 좁힌다" 방침도 함께 정정한다.
2. **D2** — `ja/advanced/moai-web-console.md:149` 를 열거 §7 오탐표에 넣고, AC-TCD-003 기대값을 `0건` 에서 `허용 목록 정확 일치` 로 바꾸거나 정규식으로 구조 배제한다. 어느 쪽이든 `plan.md §B` 의 "허용 목록 없음" 결론과 `spec.md §7` 잔여 위험을 같은 변경에서 갱신한다.
3. **D3** — AC-TCD-006 변이 지점을 파일 부류 × 로케일로 넓히고, 가드가 읽은 파일 수 == 12 를 단정하는 AC 를 추가하며, AC-TCD-008 의 Then 을 리터럴 `--- PASS: TestDocsTabContract/names` 로 고친다.
4. **D4** — AC-TCD-007 에 README 4본을 넣고, 기준을 고정 SHA(또는 읽는 시점 merge-base)로 바꾸고, 변경 0 을 "측정 불가" 로 보고하는 대조군을 붙인다.
5. **MP-7** — `plan.md:106` 의 리터럴 `[NEEDS CLARIFICATION: …]` 토큰을 문장에서 없앤다.
6. **D7** — `plan.md`·`acceptance.md` frontmatter 의 `status:` 줄 제거.
7. **D5** — AC-TCD-009/010 을 회귀 가드로 명시 분류, `echo "rc=$?"` 정리, 009 기준 ref 고정.
8. **D6** — 산문 4자리를 D9~D12 행으로 승격, AC-TCD-001 기대값 32.
9. **D8** — "서로 독립인 세 측정" 문구 하향.
10. **D9 / D10 / D11** — 수사 대안열 개방, AC-TCD-005 자리 고정, DoD 3항에 검증 동사 부여 또는 하향.

재감사는 위 열거된 결함 델타에 한정한다(Tier M 상한 2회 중 1회 남음).

---

## 검증하지 못한 것 (Gap — 조용히 통과시키지 않았다)

- **가드 자체의 동작**: `internal/web/docs_tab_contract_test.go` 는 아직 없다. AC-TCD-004/006/008 은 실행으로 판정하지 못했고, 변이체 작성 가능성이라는 **설계 수준 논증**으로만 평가했다. 가드가 작성된 뒤 D3 의 변이체 A·B 를 실제로 써서 통과하는지 확인해야 결론이 확정된다.
- **hugo 경고 0 의 인과**: 현재 무경고는 실측했으나, 문서 수정 후에도 무경고인지는 수정이 없어 확인 불가.
- **`origin/develop` 의 미래 이동**: 측정 시점에 `origin/develop` == `1d150a27d` == base 다. D4/D5 의 움직이는-ref 위험은 구조적 판단이며 이 시점에는 아직 발현하지 않았다.
- **다른 로케일의 미열거 오탐**: `ja:149` 를 찾은 방식(AC-003 스윕과 열거표의 집합 차분)은 **현재 정규식이 보는 범위 안에서만** 완전하다. 정규식이 D1/D9 대로 넓어지면 새 미열거 적중이 더 나올 수 있고, 그것은 넓힌 뒤 다시 세어야 한다.

---

## 잔여 위험

- D1 과 D2 는 같은 뿌리(정규식 하나가 "무엇이 설정 탭 계수인가" 를 단독 판별한다)에서 갈라져 나왔다. 둘을 따로 고치면 한쪽 수정이 다른 쪽을 되살린다 — 인접 창을 넓히면(D1) 화이트리스트 안 오탐이 늘고(D2), 좁히면 반대가 된다. **한 번의 변경에서 두 방향을 같은 케이스 표로 고정**하지 않으면 이 카드는 자기가 고치려는 병을 정규식 안에서 재생산한다.
- `plan.md §B` 의 "20자리 전부 removable ⇒ must-stay 없음 ⇒ 허용 목록 없음" 은 여러 결정이 매달린 전제다. D2 의 처분이 (a) 로 가면 이 전제가 무너지고 N1 의 단정 형태가 바뀐다. 전제를 바꾸는 수정임을 인지하지 못한 채 AC 만 손대면 층 간 모순이 남는다.
