# SPEC Review Report: SPEC-AC-COLLECTOR-ANCHOR-001 (카드 t528)

- Iteration: 1/2 (Tier M ceiling)
- **Verdict: FAIL** — must-pass firewall (MP-7) + blocking 결함 6건
- Overall Score: **0.80** (Tier M PASS 임계 0.80 — 점수는 임계를 만족하나 must-pass 실패는 점수로 상쇄되지 않는다)
- 감사 트리: `.claude/worktrees/t528`, 브랜치 `WT-ac-collector-anchor`, HEAD `52f863f36` (감사 시점 재판독)
- 입력: `spec.md` + `plan.md` + `acceptance.md` (Tier M 3종) + `.moai/reports/t528/measurement-20260908.md`
- Reasoning context ignored per M1 Context Isolation. 카드 배차문의 서술은 감사 입력이 아니라 감사 대상 질문으로만 소비했다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `spec.md:110-120`에 `REQ-ACA-001-001` ~ `-011`이 결번·중복 없이 연속. 제로패딩 일관.
- **[PASS] MP-2 GEARS 형식 준수 (requirement layer)** — 판정 대상은 `spec.md`의 `REQ-XXX` 층이며 `acceptance.md`의 Given-When-Then은 검증층이므로 여기서 벌하지 않았다(M3 § Scope). `-001` ubiquitous(`…해야 한다(shall)`), `-003`/`-004` **Where**(`spec.md:112-113`), `-009` **When**(`spec.md:118`), `-006`/`-011` unwanted 형(`shall not`, `spec.md:115,120`). 한국어 렌더이나 패턴 어휘가 명시돼 있어 수용.
- **[PASS] MP-3 프론트매터 유효성** — `spec.md:2-14`에 canonical 12필드 전부 present, snake_case alias 0건, `version: "0.1.0"` quoted, `phase: "v3.2.0 target"`은 금지된 lifecycle 토큰이 아니다. `tier: M`은 optional 필드로 적법.
- **[N/A] MP-4 언어 중립성** — Go 단일 언어 프로젝트 내부 패키지(`internal/spec`) 한정 SPEC. 16개 프로그래밍 언어를 다루지 않으므로 해당 없음(자동 통과).
- **[PASS] MP-5 D7 교차-SPEC 정합** — 실행: `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' … | sort -u` → `SPEC-ARTIFACT-STATELESS-001` / `SPEC-CLIFIX-CONCURRENCY-001` / `SPEC-COVERAGE-RULE-SCOPE-001` 3건 + 자기참조. 셋 다 존재하며 `status=completed`. retired/superseded/archived 0건 → BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall'` = 0/0/0 (spec·plan·acceptance). 자동 통과.
- **[FAIL] MP-7 clarification gate** — 실행: `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-AC-COLLECTOR-ANCHOR-001/`

  ```
  plan.md:35:**[NEEDS CLARIFICATION: 마지막 분절을 숫자로 고정하는 것이 관측된 1160개 줄 전부를 덮는가]**
  ```

  미해소 마커가 audit 시점에 살아 있다. 이 마커는 형식 위반에 그치지 않는다 — 아래 D2가 보이듯 **형제 문서는 같은 결정을 이미 확정된 것으로 취급하고 있어**, 두 산출물이 「열린 결정」과 「닫힌 결정」으로 갈라져 있다.

---

## Category Scores (rubric-anchored)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분의 REQ가 단일 해석이나, §B.1의 열린 결정 vs `acceptance.md:17,103`의 확정 서술이 구현자에게 상반된 지시를 준다(D2) |
| Completeness | 0.80 | 0.75~1.0 | 전 필수 절 present, `### Out of Scope — <topic>` H3 4개 + 불릿 ✓, Gaps/Residual-risk 절 별도 존재. 다만 blast radius 지도가 `lint.go:636`과 CLI 치명 경로를 누락(D4/D8) |
| Testability | 0.65 | 0.50~0.75 | 판정 명령이 이 트리에서 공허하게 통과하는 AC 1건(D1), 계측 대상 신호에 도달하지 못하는 AC 절 1건(D3), 과수용을 재는 AC 0건(D5) |
| Traceability | 1.00 | 1.0 | REQ 11건 전부 ≥1 AC 대응(`acceptance.md:29-43`), 존재하지 않는 REQ를 인용하는 AC 0건, `plan.md` §C·§G의 AC 인용 3건이 전부 실재 AC로 해소됨 |

산술 평균 0.80. Tier M 임계(0.80)에 걸치나, MP-7 및 blocking 결함이 점수와 무관하게 FAIL을 강제한다.

---

## Defects Found

### D1 — AC-ACA-001-001의 판정 명령이 이 트리에서 공허하게 통과한다
- 위치: `acceptance.md:65-71` (`§D.1`), 근거 보고서 `measurement-20260908.md:7,47`
- Severity: **critical** · Class: **blocking**
- **Claim**: AC-001은 「기준선을 만든 것과 **같은 프로브를 같은 명령으로** 다시 태우면」을 판정 조건으로 삼는데, 그 프로브는 트리에 존재하지 않는다.
- **Evidence** (이 트리, HEAD `52f863f36`):
  ```
  $ ls internal/spec/zz_t528_probe_test.go
  ls: internal/spec/zz_t528_probe_test.go: No such file or directory

  $ go test ./internal/spec/ -run TestT528Probe -count=1 -timeout 600s
  ok  github.com/modu-ai/moai-adk/internal/spec  0.516s [no tests to run]
  EXIT=0
  ```
- 이것은 `verification-completeness.md` §1.1이 이름 붙인 **빈 집합 위의 초록**이다. 셀렉터가 0건을 고르고도 `ok` + exit 0을 낸다. AC-001을 문자 그대로 집행하면 아무것도 재지 않고 PASS가 난다.
- 파생 문제: 프로브가 없으므로 **216 / 1160 / 100 / 18을 제3자가 기록만 보고 재유도할 수 없다.** 「AC 선언 줄(불릿)」의 판별식이 삭제된 코드 안에만 있었고, 1차→2차 정정(`measurement-20260908.md:116-127`)이 그 판별식 변경만으로 수치가 크게 움직였음을 스스로 증명한다. run 단계가 프로브를 재저작하면 needle 정의가 달라질 수 있고, 그러면 새 수치는 216/1160과 비교 가능하지 않다 — AC-001이 요구하는 「기준선과 나란히」가 성립하지 않는다.
- 필요한 수정: 프로브를 **커밋 가능한 형태로 트리에 복원**(또는 판별식을 산문으로 완전 명세)하고, AC-001에 「이 명령이 최소 1개 테스트를 선택했음」을 확인하는 절을 추가한다. `[no tests to run]` 토큰 부재 확인이 가장 싼 형태다.

### D2 — `plan.md` §E의 마일스톤 순서가 AC-ACA-001-002의 RED-now 대조군 규율과 모순된다
- 위치: `plan.md:96-97` (M1→M2) vs `acceptance.md:93` (`§D.2` RED-now)
- Severity: **critical** · Class: **blocking**
- **Claim**: AC-002는 「대조군을 넓힘 **이후에** 만들면 그것은 대조군이 아니라 사후 기록이다」라고 못박으면서, 같은 문장에서 그 산출물을 **M2 착수 시점**에 만들라고 지시한다. `plan.md` §E는 M1이 「앵커를 구현하고 코퍼스에 태운다」이므로, M2 착수 시점의 트리에는 이미 넓힌 파서가 들어 있다. 문서가 금지한 바로 그 순서를 문서 자신이 지시한다.
- **Evidence**:
  - `plan.md:96` — `**M1** | §B의 네 축을 반영한 앵커를 parseSingleACLine에 구현하고 … 재측정한다`
  - `plan.md:97` — `**M2** | 회귀 대조군을 세운다 — 현재 파싱되는 18개 파일이 여전히 같은 결과를 내는지`
  - `acceptance.md:93` — `**RED-now**: 기준선의 18개 파일 목록과 각 파일의 AC 집합을 M2 착수 시점에 산출물로 먼저 남긴다 … 대조군을 넓힘 **이후에** 만들면 그것은 대조군이 아니라 사후 기록이다.`
- 실질 피해: 「before」 집합을 M1 이후에 만들면 넓힌 파서로 잰 값이 되어 무회귀 판정이 **자기 자신과의 비교**가 된다(항상 초록). D1과 결합하면 더 나쁘다 — before를 다시 재려면 프로브도, 이전 파서도 둘 다 복원해야 한다.
- 부수 확인: 18개 파일의 **목록 자체가 기록에 없다.** `measurement-20260908.md:55-57`은 POSITIVE-NEEDLE 3줄만 남겼고 나머지 15개는 수만 있다. 대조군의 구성원이 기록되지 않았으므로 재현 불가.
- 필요한 수정: M0(또는 M1 착수 전)에 18개 파일 목록 + 파일별 루트 AC 집합을 산출물로 먼저 고정하도록 §E를 재배열하고, AC-002의 RED-now 절을 그 순서로 정정한다.

### D3 — AC-ACA-001-010의 `ValidateDepth` / `DuplicateAcceptanceID` 계측이 신호가 없는 경로에서 이뤄진다
- 위치: `acceptance.md:204` (`§D.10` 셋째 **And**), `spec.md:169,178`, `plan.md:128`
- Severity: **critical** · Class: **blocking**
- **Claim**: SPEC은 「`doc.Criteria`가 커지면 `ValidateDepth` / `DuplicateAcceptanceID`가 새로 발화해 착지 시 코퍼스를 붉게 만들 수 있다」를 최대 잔여 위험 2순위로 놓고, AC-010이 이를 **lint 실행에서** 관측하라고 지시한다. 그런데 lint 경로는 그 오류를 **버린다**.
- **Evidence**:
  ```
  internal/spec/lint.go:636:  criteria, _ := ParseAcceptanceCriteria(body, false)
  internal/spec/lint.go:637:  doc.Criteria = criteria
  ```
  두 번째 반환값(`ValidateDepth` 오류와 `DuplicateAcceptanceID`를 담는 슬라이스)이 `_`로 폐기된다. 즉 `moai spec lint`를 아무리 돌려도 이 두 오류는 **0으로 관측된다** — 실제로 발화하든 안 하든. AC-010의 해당 절은 항상 「발화 없음」을 산출하는, 판정 불가능한 계측이다.
- **정정된 위험 소재**: 진짜 붉어지는 자리는 CLI다.
  ```
  internal/cli/spec_view.go:73-86
      criteria, parseErrors := spec.ParseAcceptanceCriteria(...)
      switch e := err.(type) {
      case *spec.DanglingRequirementReference: … warning
      case *spec.MissingRequirementMapping:    … warning
      default:
          return fmt.Errorf("parse error: %w", err)   // ← 치명
      }
  ```
  `DuplicateAcceptanceID`와 깊이 초과는 `default`로 떨어져 **명령 전체를 실패시킨다.** 넓힘이 어떤 SPEC에서 중복 ID를 새로 만들면, 그 SPEC의 `moai spec view --acceptance`는 기준선의 `No acceptance criteria found`(양성이지만 무해)에서 **하드 에러**로 바뀐다 — 이 카드가 고치겠다고 지목한 바로 그 명령이 기준선보다 나빠진다. 어떤 산출물도 이 방향을 계측하지 않는다.
- 필요한 수정: AC-010의 셋째 **And**를 lint 경로가 아니라 `parseAcceptanceCriteriaInternal`의 오류 슬라이스를 직접 읽는 프로브(또는 코퍼스 전체에 대한 `spec view` 실행)로 재지정하고, 「기준선에서 blind였던 SPEC이 넓힘 이후 **에러로** 바뀌지 않는다」를 AC-012에 negative 절로 추가한다.

### D4 — §4의 「분리」 주장이, 그 주장을 뒷받침하려 인용한 출처에 의해 반박된다
- 위치: `spec.md:33`, `spec.md:140-141` / 출처 `internal/spec/lint_coverage_sibling.go:29-32`
- Severity: **major** · Class: **blocking**
- **Claim**: SPEC은 「섹션 스코핑이 산문 가드이고, 항목 문법을 넓히는 일은 그 가드를 약화시키지 않는다」고 두 번 단언하며(`spec.md:33`, `:141`), 그 근거로 `lint_coverage_sibling.go` 머리주석을 인용한다. 그러나 그 주석은 **정확히 반대**를 말한다.
- **Evidence** (출처 verbatim, `internal/spec/lint_coverage_sibling.go:29-32`):
  ```
  // ParseAcceptanceCriteria, which is scoped twice over: findACSectionStart needs
  // an `##` heading containing "acceptance", and parseSingleACLine needs the
  // `AC-…:` colon form. BOTH scopings exist because spec.md is a mixed document
  // in which prose must not be read as AC.
  ```
  산문 가드의 주체로 명시된 둘 중 **하나가 바로 이 카드가 넓히려는 콜론 형식**이다. SPEC은 이 문장을 그대로 인용해 놓고(`spec.md:140`), 인용문이 부정하는 결론을 그 다음 줄에서 단언한다.
- 왜 blocking인가: 이것은 문구 다툼이 아니라 **권고의 전제**다(`verification-claim-integrity.md` §1.1 surface 4). 「가드가 약화되지 않는다」가 성립하면 잔여 위험은 뮤턴트로 충분하고, 성립하지 않으면 **과수용을 실측할 의무**가 생긴다 — 그리고 D5가 보이듯 그 의무를 지는 AC가 없다.
- 구체적 반례(구성): AC 절 **안**의 산문 불릿
  ```
  - AC-OGR-003 (RETIRED) — 이 항목은 SPEC-X로 이관됐다.
  ```
  현행 앵커는 거절한다(콜론 없음). §B.3+§B.4를 반영한 넓힌 앵커는 id `AC-OGR-003` → 괄호 한정어 `(RETIRED)` 건너뜀 → 구분자 `—` → content=산문으로 **수용한다.** 섹션 스코핑은 이 줄을 막지 못한다 — 절 안에 있기 때문이다. 코퍼스 실측으로는 이 정확한 모양을 찾지 못했으나(`/usr/bin/grep -rEn`로 em-dash AC 불릿 19파일을 전수 열람했고 전부 진짜 선언이었다), **반례의 부재는 가드의 존재가 아니다** — 넓힘 이후 새로 들어오는 944줄은 아직 아무도 열어보지 않았다.
- 필요한 수정: §4의 「약화시키지 않는다」를 **「콜론 형식이 담당하던 산문 가드의 일부를 의도적으로 포기하며, 그 대가를 D5의 실측으로 상환한다」**로 정정한다. 인용을 유지하려면 인용문의 내용과 결론이 일치해야 한다.

### D5 — 과수용(false positive)을 재는 AC가 없다 — `plan.md` §H가 스스로 이름 붙인 반패턴이 미봉쇄
- 위치: `plan.md:139` (§H 첫 항목) vs `acceptance.md:29-43` (AC 매트릭스 전체)
- Severity: **major** · Class: **blocking**
- **Claim**: `plan.md:139`는 「넓힌 파서가 **더 많이 읽는다**는 사실만 관측하고 **무엇을 잘못 읽는지**는 재지 않는 것」을 첫 번째 반패턴으로 적는다. 그런데 13개 AC 중 그것을 재는 AC가 없다.
  - AC-001은 **회수량과 잔여 거절**만 잰다 — 새로 수용된 줄의 성질은 대상이 아니다.
  - AC-003의 부정 케이스는 `AC-FOO-BAR` 한 건뿐이며, 이는 「숫자 꼬리 규칙」의 단위 테스트이지 코퍼스 과수용 측정이 아니다.
  - 뮤턴트(AC-011)는 **구현이 자기 테스트에 대해** 갖는 경계를 그린다. 「코퍼스의 실제 산문을 AC로 읽는가」는 그리지 못한다 — 뮤턴트는 기능을 **제거**하는 방향이고, 과수용은 기능이 **작동할 때** 일어난다.
- **Evidence**: `acceptance.md:29-43` 매트릭스 13행 전수 판독. 「새로 수용된 줄 중 N건을 표본 열람했다」류 절이 어느 AC에도 없음. `acceptance.md:274` Edge Cases는 절 **밖** 180건만 다루고, 절 **안**의 새 수용분은 언급하지 않는다.
- 필요한 수정: AC 1건 추가 — 「M1 재측정에서 **새로 수용된** 줄 중 무작위 표본 N건(최소 30)을 열람하고, 선언이 아닌 것(산문 인용·폐기 표시·교차참조)의 건수와 줄을 기록한다. 0이 아니면 억제하지 말고 적는다.」 판정은 크기가 아니라 기록의 완결성으로 두면 AC-001과 같은 등급 체계에 맞는다.

### D6 — baseline attribution: 핀 SHA가 측정 집합을 고정하지 못한다
- 위치: `measurement-20260908.md:85-86`, `acceptance.md:54`
- Severity: **major** · Class: **blocking**
- **Claim**: 문서 수준 좌표 핀은 `52f863f36`이지만, 측정 대상 코퍼스는 미추적 파일을 포함하므로 **같은 SHA에서 값이 이미 달라졌다.**
- **Evidence** (HEAD 불변, 감사 시점 `git rev-parse --short HEAD` = `52f863f36`):
  ```
  $ find .moai/specs -name spec.md | wc -l
       807
  ```
  보고서의 `spec.md=806`과 1건 어긋난다. 원인은 자명하다 — 이 SPEC 자신의 `spec.md`가 측정 이후 그 트리에 추가됐다. 즉 분모가 커밋 상태로 결정되지 않는다. `verification-completeness.md` §4가 요구하는 것은 「불변 주장을 고정 SHA에 핀」인데, 여기서는 **핀이 있어도 대상 집합이 움직인다.** 재측정 때 806↔807 차이를 「넓힘의 효과」로 오독할 여지가 그대로 남는다.
- 부수: `acceptance.md:55`가 스스로 인정하듯 **종료 코드가 기록되지 않았다.** 정직한 자기신고이며 이 점은 가점 요소지만, release-blocking AC 2건이 기대는 RED-now 셀이 §2.1 4요소 중 하나를 결여한 상태로 남아 있다는 사실 자체는 변하지 않는다.
- 필요한 수정: 재측정 시 (a) 코퍼스 집합을 명령 출력으로 함께 고정(`find … | wc -l` 값과 `git status --short .moai/specs | wc -l`), (b) 종료 코드 별도 필드, (c) 측정 트리 SHA. AC-001이 (b)(c)는 이미 요구하므로 (a)만 추가하면 된다.

### D7 — 뮤턴트 6이 자기가 겨냥한 절을 행사하지 못한다
- 위치: `plan.md:77`, `acceptance.md:147`
- Severity: **minor** · Class: **optional**
- **Claim**: 뮤턴트 6은 「`findACSectionStart`를 항상 0 반환으로」이고 AC-006이 잡을 것으로 사전 선언돼 있다. 그러나 `return 0`은 「문서 전체가 AC 절」을 만들지 않는다 — `extractACLines`가 **첫 `##` 헤딩에서 break** 하므로, 대부분의 `spec.md`에서 수집은 문서 앞머리에서 즉시 끝난다.
- **Evidence**:
  ```
  internal/spec/parser.go:83-85
      if strings.HasPrefix(trimmed, "##") { break }
  ```
  결과적으로 이 뮤턴트는 AC-006의 **첫째 절**(절 안 선언이 수집됨)을 깨서 적발된다. **둘째 절**(절 밖 선언이 수집되지 않음)은 행사되지 않는다. 즉 「절 밖 줄을 수집하는 구현」을 이 뮤턴트가 대표하지 못한다.
- 필요한 수정: 뮤턴트 6을 「`findACSectionStart` → `return 0` **그리고** `extractACLines`의 `##` break 제거」로 바꾸거나, break 제거만 주입하는 별도 뮤턴트를 추가한다. 미적발이면 그 사실을 경계 기록으로 남기는 규율은 그대로 유효하다.

### D8 — 소비자 지도가 파싱 지점을 누락한다
- 위치: `spec.md:97-101`, `measurement-20260908.md:87-89`
- Severity: **minor** · Class: **optional**
- **Claim**: SPEC은 프로덕션 소비자를 `lint.go:915`와 `spec_view.go:72` 둘로 특정한다. 독립 조사 결과 **호출 지점은 맞으나 지도는 불완전**하다.
- **Evidence**:
  ```
  $ grep -rn --include='*.go' 'ParseAcceptanceCriteria' . | grep -v _test.go
  internal/spec/lint.go:636:      criteria, _ := ParseAcceptanceCriteria(body, false)
  internal/cli/spec_view.go:72:   criteria, parseErrors := spec.ParseAcceptanceCriteria(...)
  ```
  `lint.go:636`이 지도에 없다. 이 줄이 D3의 근거이므로 누락이 무해하지 않다.
  나머지 도달성 확인(배차문 질문 3 대응): `buildTree`·`autoWrapSingle`·`DuplicateAcceptanceID`·`ValidateDepth`는 `parseAcceptanceCriteriaInternal` 안에서 전부 도달하며(`parser.go:43-48,116-181`), `CheckDanglingReferences`는 **프로덕션 호출자가 0건**이다(정의만 존재) — 넓힘이 그 함수에 닿지 않는다. `ACIDInvalid`가 프로덕션 규칙으로 존재하지 않는다는 `spec.md:102`의 주장은 **참**이다(유일한 등장이 `lint_coverage_sibling.go:18`의 주석).
- 추가 관측(미검증 잔여): `buildTree:136-141`은 중복 ID를 만나면 `continue`로 **그 줄을 버린다.** 넓힘으로 산문 불릿이 먼저 들어오고 진짜 선언이 뒤에 오면, 진짜 선언이 조용히 사라진다. 어떤 AC도 이 방향을 다루지 않는다.

### D9 — 등급 분류가 3건을 누락하고, 무회귀·불변 AC는 no-op 트리에서도 통과한다
- 위치: `acceptance.md:23-25` vs 매트릭스 `:29-43`
- Severity: **minor** · Class: **optional**
- 등급 절이 `-012`/`-013`을 분류하지 않는다(`-001`/`-002` release-blocking, `-003~-009` regression-guard, `-010`/`-011` 계측 의무까지만).
- AC-002(무회귀)·AC-006(스코핑 불변)·AC-007(`lint.go` 바이트 동일)·AC-013(패키지 초록)은 **아무 변경도 하지 않은 트리에서 전부 통과한다.** 이는 결함이 아니라 가드의 성질이고 `acceptance.md:24`가 「없는 RED를 있다고 적지 않는다」로 정직하게 처리했다 — `verification-completeness.md` §2 준수 사례로 기록한다. 다만 매트릭스가 이들을 must-pass로 표기하므로 「must-pass 12건 통과」를 실질 진척으로 오독할 여지가 있다. 등급 열을 매트릭스에 직접 넣는 편이 싸다.

### D10 — `.1` 형 수치 하위 접미가 §B.2 논의에 없다
- 위치: `plan.md:37-39` (§B.2)
- Severity: **minor** · Class: **optional**
- 코퍼스에 `AC-LCLN-004.1` 형태가 실재한다(`.moai/specs/SPEC-V3R5-LINT-CLEAN-001/spec.md:125`). 현행 접미 문법은 소문자 알파(`.a` / `.a.i`)만 받으므로 거절되며, `hasIDSuffix`가 `strings.Contains(id,".")`인 이상 만약 수용되면 `autoWrapSingle` 분기가 뒤집혀 **트리 모양이 바뀐다**(`parser.go:177,185`). §B.2는 「접미 의미론 보존」만 말하고 이 미수용 형태의 처분을 정하지 않는다. AC-001의 잔여 분해가 흡수할 수는 있으나 명시하는 편이 낫다.

---

## 감사 질문별 응답 (배차문 1-7)

1. **기준선은 귀속 가능한가** — **아니다.** D1·D6. 프로브가 삭제돼 판별식이 기록에 없고, 종료 코드가 없으며, 핀 SHA가 코퍼스 집합을 고정하지 못한다(같은 SHA에서 806→807). 216/1160/100/18은 **Claim이지 재유도 가능한 Evidence가 아니다.** 18개 대조군의 구성원 목록도 부재.
2. **산문 가드는 온전한가** — **분리는 부분적으로 합리화다.** D4. 인용된 출처가 콜론 형식을 산문 가드의 절반으로 지목한다. 반례는 구성 가능하며(`- AC-X-001 (RETIRED) — 산문`), 섹션 스코핑이 막지 못한다. 다만 코퍼스 em-dash AC 불릿 19파일 전수 열람에서는 실제 오탐 줄을 찾지 못했다 — **이는 안전의 증거가 아니라 미측정의 표시다**(새로 들어올 944줄은 아직 열어보지 않았다).
3. **blast radius는 완전한가** — **불완전.** D8. `lint.go:636` 누락, 그리고 `spec_view.go:80-86`의 `default → 치명 에러` 경로가 어디에도 없다. `ValidateDepth`/`Duplicate`는 lint 경로에서 폐기되고 CLI 경로에서 치명이 된다 — SPEC은 정반대로 배치했다(D3). `CheckDanglingReferences`는 프로덕션 호출자 0.
4. **AC는 반증 가능한가** — 대체로 그렇다. 예외 3건: AC-001(공허한 셀렉터, D1), AC-010 셋째 절(신호 없는 경로, D3), 과수용 무측정(D5). AC-002는 **무증가 대조군을 갖춘 유일한 AC**이며 설계가 훌륭하다 — 순서만 D2로 깨져 있다.
5. **뮤턴트는 작도인가 주장인가** — **대체로 작도.** 사전 선언 + 미적발 보존 규율(`plan.md:65-66`, `acceptance.md:226-232`)은 이 프로젝트 규율의 모범 사례다. 다만 뮤턴트 6은 겨냥한 절을 행사하지 못하고(D7), 집합 전체가 「기능 제거」 방향이라 **과수용 축을 그리지 못한다**(D5).
6. **범위 규율** — **준수.** 180건은 `spec.md:143-147`에서 관측으로만 기록되고 어떤 AC도 회수 대상으로 삼지 않는다(오히려 AC-006이 반대 방향을 요구). `plan.md` 어디에서도 `lint.go`·`internal/cli`·코퍼스 재저작에 손을 뻗지 않는다. M5의 `moai spec view` 실행은 소비이지 편집이 아니다. `internal/spec/lint.go` 바이트 동일성이 AC-007로 판정 대상화돼 있다.
7. **내부 정합성** — 추적성 만점(REQ 11/11 대응, 고아 AC 0, 교차문서 AC 인용 3건 전부 해소). 수치 정합 통과: 모든 인용 수치가 측정 보고서에 존재하며, `–`(U+2013)에 실측 근거가 없다는 사실을 `acceptance.md:190`이 **스스로 밝힌다**(모범). 시간 추정 0건. **불합치 1건**: §B.1의 열린 결정 vs `acceptance.md:17,103`의 확정 취급(D2/MP-7).

---

## Recommendation

FAIL. 아래 순서로 수정한 뒤 iteration 2를 요청한다. **1~3이 must-fix**이고, 4~6은 blocking이지만 병렬 처리 가능하다.

1. **MP-7 해소** — `plan.md:35`의 `[NEEDS CLARIFICATION]`을 닫는다. 「마지막 분절 숫자 고정」은 이미 `acceptance.md:17,103`이 확정으로 취급하고 AC-003의 부정 케이스가 그 위에 서 있으므로, **결정을 확정으로 승격하고 82건 잔여 분해를 「결정 재검토 트리거」로 격하**하는 것이 실질에 맞다. 운영자 판단이 필요하면 orchestrator가 `AskUserQuestion`으로 묻는다 — 이 감사자는 묻지 않는다.
2. **D1 — 프로브를 트리에 복원**하고 AC-001에 empty-sweep 가드(`[no tests to run]` 부재 확인)를 추가한다. 프로브를 커밋하지 않겠다면 needle 판별식을 산문으로 완전 명세해 제3자 재유도를 가능케 한다.
3. **D2 — §E를 M0(대조군 고정) → M1(구현) → M2(비교)로 재배열**하고, AC-002의 RED-now 절을 그 순서로 정정한다. 18개 파일 목록 + 파일별 루트 AC 집합을 `.moai/reports/t528/` 산출물로 먼저 남긴다.
4. **D3 — AC-010 셋째 And의 계측 지점을 재지정**한다(`parseAcceptanceCriteriaInternal` 오류 슬라이스 직독 또는 코퍼스 `spec view` 실행). 아울러 AC-012에 negative 절 추가: 「기준선 blind SPEC이 넓힘 이후 `parse error:`로 바뀌지 않는다」.
5. **D4 — `spec.md:33,141`의 「약화시키지 않는다」를 정정**한다. 인용문이 부정하는 결론을 인용문 옆에 두지 않는다.
6. **D5 — 과수용 실측 AC 1건 추가**(새 수용분 표본 ≥30 열람, 비선언 건수·줄 기록, 0이 아니면 억제 금지). 판정은 기록의 완결성으로 둔다.

optional(D7~D10)은 iteration 2 착수 전에 함께 처리하면 싸지만, 이 감사의 verdict를 좌우하지 않는다. **긴 optional 목록으로 FAIL을 만들지 않았다** — FAIL은 MP-7과 D1~D6이 만든다.

칭찬은 근거와 함께 남긴다: 등급 분류 절(`acceptance.md:19-25`)이 「없는 RED를 있다고 적지 않는다」로 release-blocking을 2건으로 좁힌 것, 종료 코드 결여를 스스로 신고한 것(`:55`), `–`에 실측이 없음을 밝힌 것(`:190`), 1차 분류 정정이 **간극을 넓히는** 방향임에도 기록으로 남긴 것(`spec.md:27-29`) — 네 곳 모두 이 프로젝트의 증거 규율을 정확히 실행한 사례다. 이 SPEC의 결함은 정직성의 결함이 아니라 **계측 지점과 순서의 결함**이다.

---

## Gaps — 이 감사가 관측하지 않은 것

- **넓힌 앵커의 실제 정규식을 보지 못했다.** §B는 방향만 정하고 패턴을 쓰지 않았으므로, D4의 반례는 §B.3+§B.4 기술을 그대로 따랐을 때의 **구성**이지 실행 관측이 아니다.
- **944개 거절 줄을 표본 열람하지 않았다.** D5의 근거는 「그것을 재는 AC가 없다」는 문서 관측이며, 실제 오탐률은 이 감사도 모른다.
- **`CoverageRule` finding 수를 직접 세지 않았다.** 방향 예측(감소)의 검증은 run 단계 몫이며, 이 감사는 AC-010의 계측 지점 오류만 판정했다.
- **코퍼스 전체에 대한 `moai spec view` 실행을 하지 않았다.** D3의 CLI 치명 경로는 코드 판독(`spec_view.go:80-86`)에 근거하며, 현재 코퍼스에서 실제로 몇 건이 그 경로에 빠지는지는 재지 않았다.

## Residual-risk

- 이 감사는 **문서와 코드**를 읽었고 **구현을 보지 않았다.** run 단계 구현이 §B의 서술과 다르게 넓히면 D4·D5의 크기가 달라진다.
- D6의 806↔807 불일치는 원인을 자명하게 특정했으나(자기 SPEC 추가), 다른 미추적 SPEC 디렉터리가 측정 이후 더 들어왔을 가능성은 배제하지 않았다.
- MP-7을 「형식 위반」으로만 읽고 마커만 지우면, D2가 드러낸 실질 모순(열린 결정 vs 확정 취급)은 그대로 남는다. 마커 삭제는 수정이 아니다.
