# SPEC Review Report: SPEC-SPEC-LINT-ID-ARG-001

Iteration: 1/1 (Tier S 천장 — `harness.plan_audit_tier_ceilings` S=1)
Verdict: **FAIL**
Overall Score: **0.56** (적용 임계값: **Tier S = 0.75**, `spec-workflow.md` § SPEC Complexity Tier)

M1 Context Isolation: 저자의 추론 맥락은 무시했다. 판정 입력은 이 트리의 파일뿐이다 —
`spec.md` + `plan.md` + `progress.md`(Tier S 입력 계약), 형제 SPEC과 그 판정문(비중첩 판별 목적 한정),
그리고 아래에 인용한 Go 소스.

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t518`, 브랜치 `WT-spec-lint-axes`,
base `0b1e27877`. 종료 코드는 모두 파이프 없이 읽었다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-SLI-001`..`008` 연속, 결번·중복 0, 자릿수 일정(`spec.md:111-118`).
- **[PASS] MP-2 GEARS 형식 (요구사항 층 한정)** — 8건 모두 다섯 패턴 중 하나의 형태를 갖춘다.
  판정 대상은 `spec.md §C`의 `REQ-XXX` 항목이며, `§D`의 Given-When-Then은 검증 층이므로 여기서 채점하지 않았다.
  다만 REQ-SLI-003의 `Where` 라벨은 GEARS 의미(capability gate / feature flag / static config)와 어긋난다 — D-12.
- **[PASS] MP-3 프론트매터** — 정본 12필드 전부 존재. `phase: "v3.2.0 target"`은 릴리스 라벨이며 금지된 단계 토큰이 아니다.
  기계 확인: `moai spec lint .moai/specs/SPEC-SPEC-LINT-ID-ARG-001/spec.md` → rc=0, `0 error(s), 8 warning(s)`, `FrontmatterInvalid` 0건.
- **[N/A] MP-4 언어 중립성** — 단일 언어(Go CLI) 범위 SPEC. 16개 프로그래밍 언어 도구 표면을 주장하지 않는다.
- **[PASS] MP-5 D7 교차 SPEC** — 본문이 지명한 SPEC 4건 전부 실재하고 retired/superseded/archived 없음:
  `SPEC-SEC-HARDEN-002` completed · `SPEC-CLIFIX-CONTRACT-001` completed · `SPEC-SPEC-LINT-BLIND-AXES-001` draft ·
  `SPEC-AC-COUNT-DISCRIMINATOR-001` completed. (`SPEC-FIX-001/002`·`SPEC-NOPE-999`는 픽스처 이름이지 참조가 아니다.)
- **[PASS] MP-6 D8 크로스플랫폼** — SPEC 디렉터리 전체에 `syscall` 0건(`grep -rn syscall` rc=1). 자동 통과.
- **[PASS] MP-7 해명 게이트** — `grep -rn '\[NEEDS CLARIFICATION'` 0건(rc=1). plan.md의 미해결 결정(D1·D2)은
  이 표기 규약을 쓰지 않았으나, 규약상 마커 부재는 부재다.

**must-pass 실패는 없다.** FAIL은 아래 루브릭 점수(0.56 < 0.75)에서 나온다.

---

## Category Scores (rubric-anchored)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | 요구·기준·계획 사이 모순 2건(D-1, D-4). 성실한 구현자가 의도와 다르게 구현한다 |
| Completeness | 0.75 | 0.75 | 절 구성·프론트매터 완비, `### Out of Scope — <주제>` H3 6개에 구체 불릿. 다만 REQ 하나가 AC 없이 남는다(D-5) |
| Testability | 0.50 | 0.50 | AC 9건 중 1건은 기술된 대로 실행 불가(D-1), 1건은 no-op으로 만족(D-7), 3건은 뮤턴트 가드 부재(D-8) |
| Traceability | 0.50 | 0.50 | AC→REQ 명시 인용 **0건**(기계 확인: `CoverageIncomplete` 8/8). 읽어서 7/8은 유도되지만 문서에 쓰여 있지 않고, 1건(007)은 유도조차 안 된다 |

산술 평균 = (0.50 + 0.75 + 0.50 + 0.50) / 4 = **0.5625 → 0.56**.

---

## Defects Found

### D-1 — REQ-SLI-006과 AC-SLI-006은 어떤 구현으로도 도달할 수 없다 (요구는 no-op으로 만족되고, AC는 올바른 구현에서 FAIL 난다)
- **위치**: `spec.md:116`(REQ-SLI-006) · `spec.md:147`(AC-SLI-006) · `plan.md §B D1` · `internal/cli/specid/specid.go:26-40` · `internal/spec/lint.go:952` · `internal/cli/spec_status.go:18`
- **설명**: `ValidateSpecID`가 거부하는 것은 정확히 셋뿐이다 — 절대 경로(`:28`), `..` 포함(`:32`), 경로 구분자 `/` `\`(`:36`).
  그런데 D1 판별식은 이 셋을 **판별 단계에서 이미 전부 배제한다**.
  - 엄격 정규식 `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`(`lint.go:952`)는 `.`도 `/`도 문자 클래스에 없어 `SPEC-../../etc`에 매치되지 않는다.
  - 같은 패키지의 느슨한 정규식 `SPEC-[A-Z0-9-]+-[0-9]+`(`spec_status.go:18`)도 같은 이유로 매치되지 않는다.
  - 게다가 plan.md D1 권고는 "경로 신호가 있으면 무조건 경로"를 **선행 배제**로 둔다. `/`는 경로 신호다.
  - 귀결 (a): `SPEC-../../etc`는 ID가 아니라 **경로**로 분류되어 `linter.Lint`에 그대로 넘어간다 →
    `ParseFailure` 1건 · rc=1 · **findings 표에 그 경로가 나타난다**. AC-SLI-006의 Then 세 항목이 전부 반대로 나온다.
    즉 D1을 올바르게 구현할수록 AC-SLI-006은 FAIL이다.
  - 귀결 (b): ID로 분류된 인자는 정의상 `..`도 `/`도 없으므로 `ValidateSpecID`는 **언제나 nil을 반환한다**.
    REQ-SLI-006은 호출만 하면 만족되고, 그 호출을 지우는 뮤턴트조차 RED를 만들지 못한다.
- **Severity: critical — Class: blocking**
- **필요한 수리**: 둘 중 하나를 고르고 근거를 남긴다. (i) `ValidateSpecID`를 **판별식보다 앞에** 두고
  (형제 `view`가 `spec_view.go:46`에서 하는 그대로) AC-SLI-006을 "`SPEC-../../etc` → rc=3"으로 다시 쓴다 — 이때 D1의
  선행 배제 순서가 바뀌므로 plan.md §B에 그 변경을 기록한다. (ii) 이 호출을 도달 불가능한 심층 방어로 명시하고
  AC-SLI-006에서 거부 단언을 빼되, 그 사실("이 자리에서 거부될 수 있는 입력은 존재하지 않는다")을 SPEC에 적는다.
  어느 쪽이든 지금처럼 남겨서는 안 된다 — 지금은 요구도 기준도 실패할 수 없다.

### D-2 — 판별식 정규식의 출처가 미해결로 남았지만 답은 이미 트리에 있고, 대체안은 같은 패키지의 동명 심볼과 충돌한다
- **위치**: `plan.md §B D1` · `plan.md §D` · `internal/spec/lint.go:952` · `internal/cli/spec_status.go:17-18`
- **설명**: plan.md는 "`internal/spec`의 `specIDPattern`과 같은 문자열이어야 한다 … import 재사용이 가능한지 먼저 확인"으로
  남겼다. **지금 측정 가능하다**: `lint.go:952`의 `specIDPattern`은 소문자 시작 = 비공개이므로 `internal/cli`에서 import할 수 없고,
  공개로 바꾸는 것은 `internal/spec` 편집이라 REQ-SLI-008이 금지한다. 즉 사본 경로가 유일한 선택지인데,
  **`internal/cli` 패키지에는 이미 `specIDPattern`이 있다**(`spec_status.go:18`, `SPEC-[A-Z0-9-]+-[0-9]+`) — 같은 이름, 다른 뜻,
  더 느슨하다. 앵커가 없어 부분 문자열로 매치되고(`dir/SPEC-A-001.md` 안의 `SPEC-A-001`), 3자리 제약이 없으며(`SPEC-A-1` 통과),
  하이픈 연속도 통과한다(`SPEC--1`). SPEC도 plan도 이 충돌을 **한 번도 언급하지 않는다.**
  귀결: run-phase 행위자가 "이 패키지에 이미 있는 `specIDPattern`"을 재사용하면 과다 수용 판별식이 되고,
  그것은 plan.md 자신이 "이쪽이 더 나쁘다"고 이름 붙인 방향(경로를 ID로 오독)이다.
- **Severity: critical — Class: blocking**
- **필요한 수리**: (1) `spec_status.go:18`의 동명 심볼을 plan.md D1에 명시하고 새 심볼 이름을 지정한다(예 `specIDArgPattern`).
  (2) import 불가라는 측정 결과를 §D 사전 점검의 "확인할 것"이 아니라 **확정 사실**로 옮긴다.
  (3) 사본과 원본의 어긋남 감지 테스트가 무엇을 비교하는지(문자열 리터럴 대조) 적는다.

### D-3 — §G #4의 반사실이 틀렸다. 명명한 두 수리로는 자문 경고 8건이 하나도 사라지지 않는다
- **위치**: `spec.md:216` · `spec.md:77`
- **설명**: §G #4는 "없앨 수는 있었다(AC ID를 `AC-SLI-001-01` 꼴로, 제목에 영어 `acceptance`를 넣으면 된다)"고 적는다.
  **측정하면 반대다.** 실제 8건은 전부 `CoverageIncomplete` — `REQ REQ-SLI-00N is not referenced by any AC`이고,
  이 규칙의 covered 집합은 `collectAllREQIDs`(`internal/spec/lint.go:720-735`)가 `ac.RequirementIDs`에서 모은다.
  그 필드를 채우는 것은 `ExtractRequirementMappings`(`internal/spec/ears.go:123-130`)이며, AC 본문에 명시적
  `maps REQ-…` 토큰을 요구한다. **이 SPEC의 어떤 AC도 REQ를 인용하지 않는다**(§D 블록 grep 0매치).
  따라서 제목을 영어로 바꾸고 AC ID 문법을 맞춰 절과 ID가 파싱되더라도 `RequirementIDs`는 비어 있고 8건은 그대로 남는다.
  거부 자체는 옳다(초록을 사서 증거를 지우지 않겠다는 판단에 동의한다). 틀린 것은 **기전**이며, 그 대가는 다음 행위자다 —
  §G #4를 읽고 "고쳐 보는" 사람은 경고가 그대로인 것을 보고 파서 진단이 틀렸다고 결론 낼 수 있다.
- **Severity: major — Class: blocking**
- **필요한 수리**: §G #4에 세 번째 원인(AC가 REQ를 인용하지 않음 — 파서와 무관한, 저자 통제 하의 원인)을 추가하고,
  "제목·ID 문법만 고쳐도 8건은 남는다"를 명시한다. 원문을 지우지 말고 정정을 나란히 붙인다.

### D-4 — REQ-SLI-001의 기준 디렉터리가 §B의 관용구 제약과 모순되고, 고치려던 그 불일치를 새 자리에 다시 만든다
- **위치**: `spec.md:111`(REQ-SLI-001) · `spec.md:102`(§B 제약) · `internal/cli/spec_lint.go:150-156` · `internal/cli/spec_view.go:51-57`
- **설명**: §B는 "형제 `close`/`status`/`view`가 이미 세운 관용구이며, 새 관용구를 발명하지 않는다"고 못박는다.
  그런데 REQ-SLI-001은 기준 디렉터리를 `detectBaseDir(cwd)`로 SHALL-고정한다. 측정하면 이 둘은 같은 관용구가 아니다.
  - `detectBaseDir`(`spec_lint.go:150-156`): `<cwd>/.moai/specs`가 있으면 그것, 없으면 **`cwd` 그대로**.
  - `view`(`spec_view.go:51-57`): `findProjectRootFn()`로 **프로젝트 루트를 거슬러 찾은 뒤** `<root>/.moai/specs/<ID>/spec.md`.
  귀결: 프로젝트 하위 디렉터리에서 `moai spec view SPEC-X`는 성공하고 `moai spec lint SPEC-X`는 새 rc=3으로 실패한다.
  이 SPEC이 없애려는 결함(형제끼리 어긋난 계약)이 **다른 축에서 새로 생긴다**. "새 탐색 규칙을 도입하지 않는다"는
  근거도 성립하지 않는다 — `findProjectRoot`는 새 규칙이 아니라 형제들이 이미 쓰는 규칙이다.
- **Severity: major — Class: blocking**
- **필요한 수리**: (i) 기준을 `findProjectRoot`로 바꾸고 REQ-SLI-001을 그렇게 다시 쓰거나,
  (ii) cwd 기준을 유지하되 하위 디렉터리에서 형제와 갈라진다는 사실을 §E 또는 §G에 **수용된 부채로 명시**한다.
  어느 쪽이든 하위 디렉터리 동작을 재는 AC를 하나 붙인다.

### D-5 — REQ-SLI-007에 대응하는 AC가 없다 (요구가 마일스톤 산출물로만 실현된다)
- **위치**: `spec.md:117`(REQ-SLI-007) · `plan.md §C M4` · `spec.md §D.2`
- **설명**: REQ→AC 대응을 전수로 읽으면 001→AC-001a/001b, 002→AC-001b, 003→AC-002, 004→AC-004/005,
  005→AC-003, 006→AC-006, 008→AC-008이고 **007만 어디에도 없다.** 도움말(`Use`/`Long`)이 받아들이는 인자 모양을
  전부 밝혀야 한다는 요구는 plan.md §C M4의 "도움말 diff"라는 산출물로만 존재하며, §D.2 완료 정의에도 없다.
  체크리스트 항목은 이진 판정 기준이 아니다 — 형제 SPEC 판정문의 결함 부류 (e)와 같은 모양이다.
- **Severity: major — Class: blocking**
- **필요한 수리**: `Use` 문자열과 `Long` 본문이 ID·경로·디렉터리 세 모양을 모두 명시함을 단언하는 AC를 추가하고,
  뮤턴트(도움말 편집 되돌리기 → RED)를 함께 적는다.

### D-6 — AC와 REQ 사이에 명시적 인용이 하나도 없다
- **위치**: `spec.md §D.1` 전체
- **설명**: 기계 확인 — `moai spec lint .moai/specs/SPEC-SPEC-LINT-ID-ARG-001/spec.md` → rc=0,
  `CoverageIncomplete` **8건**(REQ-SLI-001..008 전부, 보고 라인 `spec.md:96-103`). 읽어서 유도되는 대응은 7/8이지만
  문서에는 쓰여 있지 않고, 유도는 다음 행위자마다 다시 해야 한다. D-3과 원인은 겹치되 수리는 다르다 —
  D-3은 산문의 사실관계를, 이것은 AC 본문을 고친다.
- **Severity: major — Class: blocking**
- **필요한 수리**: 각 AC에 자기 REQ를 인용한다. 코퍼스 파서가 요구하는 형태(`maps REQ-SLI-00N`)를 쓰면 경고도 함께 닫히고,
  쓰지 않기로 한다면 그 선택을 §G에 명시한다(§G #4 수리와 함께 처리).

### D-7 — AC-SLI-008(스스로 blocking이라 부른 반경 격리 기준)이 기계 판정 불가이고 no-op으로 만족된다
- **위치**: `spec.md:149-150`
- **설명**: "변경 diff … `internal/spec/` 아래 파일이 0개이고, 새로 도입된 finding 코드가 0개"라고만 적혀 있다.
  명령이 없고, **비교 기준(base ref)이 없으며**, "새 finding 코드 0개"를 재는 방법이 없다.
  수리 전 트리에서 diff는 비어 있으므로 지금 이 AC는 **초록이다** — RED-now가 없고, SPEC 자신의 §D.0-4
  ("부재를 단언하는 AC는 단독으로 아무것도 증명하지 않는다")를 이 AC가 스스로 위반한다. 짝이 되는 출현 AC도 없다.
  브리핑이 "기계 판정 가능한가"를 물은 자리이고, 답은 **약속이지 판정식이 아니다**.
- **Severity: major — Class: blocking**
- **필요한 수리**: base를 SHA로 고정하고 명령을 적는다 — `git diff --name-only 0b1e27877..HEAD`의 출력에서
  `internal/spec/` 항목이 0개임을 보이고(파이프 안에서 세지 말고 출력을 그대로 인용할 것),
  finding 코드 집합을 재는 두 번째 명령을 별도로 적는다.
  뮤턴트도 적는다: `internal/spec` 아래 파일 하나를 건드리면 RED.

### D-8 — 출현·동등성을 단언하는 AC 3건에 뮤턴트 가드가 없다 (SPEC 자신의 §D.0-3 위반)
- **위치**: `spec.md:144`(AC-SLI-004) · `spec.md:145-146`(AC-SLI-005) · `spec.md:148`(AC-SLI-007)
- **설명**: §D.0-3은 "출현을 단언하는 AC에는 뮤턴트 가드가 붙는다"이고 §D.2 완료 정의는 그 목록에 **005를 명시한다**.
  그런데 AC-SLI-005 본문에는 가드가 없다(있는 것은 "왜 필요한가"). AC-SLI-007(혼합 인자, `FILE` 집합 동등)과
  AC-SLI-004(수리 전후 분포 동등)도 없다. AC-SLI-004는 추가로 **채취 창**이 명시돼 있지 않다 — 수리 전 산출물은
  M2가 착지하기 전에 잡아야 하며 지금은 plan.md §C M1이 암시할 뿐이다.
- **Severity: minor — Class: blocking**
- **필요한 수리**: 004·005·007에 각각 RED를 만드는 뮤턴트를 적고, 004의 수리 전 산출물 채취 시점을 M1로 못박는다.

### D-9 — 선택지 3을 기각한 근거가 이 트리에서 확인되지 않는다 (결론은 옳고 논거가 틀렸다)
- **위치**: `spec.md:92`
- **설명**: "경로 형태는 … 훅·CI·`--json` 소비자가 이미 쓴다"를 기각 근거로 든다.
  `.github/` · `.claude/hooks/` · `scripts/` · `Makefile` · `.moai/config/`를 훑으면 기계적 소비자는 **단 하나**,
  `.github/workflows/spec-lint.yml:58`의 `go run ./cmd/moai spec lint --strict`이며 **인자를 하나도 넘기지 않는다.**
  결론(경로 형태를 깨지 않는다)은 독립적으로 옳다 — 경로 형태가 이 SPEC의 대조군이라는 근거는 검증된다.
  틀린 것은 인용한 소비자 사실이다. 배포판 사용자 프로젝트의 CI를 뜻한 것이라면 그것은 이 트리에서 검증 불가이므로
  미검증으로 표기해야 한다.
- **Severity: minor — Class: blocking**
- **필요한 수리**: 근거를 "대조군이므로 움직이면 어떤 측정도 귀속되지 않는다"로 바꾸고, 소비자 주장을 유지하려면
  "이 트리 밖(배포 사용자) — 미검증"으로 라벨한다.

### D-10 — 코퍼스 유래 수치가 상수로 박혔고, 이 SPEC 자신의 저작이 그것을 이미 움직였다
- **위치**: `spec.md:180`
- **설명**: §E가 "인자 없이 돌리는 전체 코퍼스 스캔의 경고 총량(오늘 4,346건)"을 확정값 모양으로 적는다.
  **같은 워크트리에서 지금 재측정하면 4,366건이다**(`moai spec lint` 인자 없음, rc=0, `0 error(s), 4366 warning(s)`).
  증분 +20은 정확히 귀속된다 — 이 SPEC의 `CoverageIncomplete` 8건 + 형제 SPEC의 12건. 즉 형제 판정문 D-2가 형제 문서에서
  실측한 것과 **같은 축의 결함이 이 문서에도 있다.** 범위 밖 산문 수치이고 어떤 AC도 그 위에 서 있지 않으므로 minor로 둔다.
  다만 §E는 이 수치를 근거로 "재계수는 형제 SPEC 소관"이라는 경계를 그으므로, 낡은 수치가 경계 판단의 입력이 된다.
- **Severity: minor — Class: blocking**
- **필요한 수리**: 수치에 측정 명령과 측정 시점을 붙이고 "재유도값"임을 명시한다(확정값처럼 읽히지 않게).

### D-11 — AC 개수 셈법이 문서 안에서 갈린다
- **위치**: `spec.md:217`(§G #5) · `spec.md §D.1`
- **설명**: §G #5는 "AC 8개 … Tier S 상한과 정확히 같다"고 적지만 §D.1의 식별자는 **9개**다
  (001a, 001b, 002, 003, 004, 005, 006, 007, 008). 8이 되는 읽기(001a+001b를 한 기준의 두 항목으로 셈)는
  타당하지만 문서 어디에도 그 규약이 적혀 있지 않다. 상한이 걸린 수치이므로 셈법이 애매하면 안 된다.
- **Severity: minor — Class: optional**
- **필요한 수리**: §G #5에 셈법 한 줄("001a/001b는 한 기준의 두 항목") 추가.

### D-12 — REQ-SLI-003의 GEARS 라벨이 의미와 어긋난다
- **위치**: `spec.md:113`
- **설명**: `(Where — 인자가 SPEC 디렉터리를 가리키는 경우)`로 적혀 있다. GEARS의 `Where`는 capability gate /
  feature flag / static config를 가리키며, "인자가 어떤 모양인 경우"는 실행 시 입력 조건이므로 `When`이다.
  형태 자체는 다섯 패턴 중 하나에 부합하므로 MP-2는 통과시켰다.
- **Severity: minor — Class: optional**
- **필요한 수리**: 라벨을 `When`으로 바꾼다.

### D-13 — 소유자 없는 후속 결함의 현재 상태가 낡았다
- **위치**: `spec.md:79` · `spec.md:191` · `spec.md:215`
- **설명**: "어느 카드도 소유하지 않는다 … 별도 후속 카드를 발행해야 한다"로 남아 있으나, 리드는 이미 t528을 발행했다.
  저작 시점 기준으로는 정확했으므로 저자의 결함이 아니다. 다만 다음 독자가 같은 판단을 다시 유도하게 된다.
  (t528 발행 사실 자체는 이 트리에서 재확인하지 않았다 — 브리핑 진술로만 안다. Gaps 참조.)
- **Severity: minor — Class: optional**
- **필요한 수리**: 후속 카드 id를 한 줄로 지명한다.

---

## 확인했고 결함이 아닌 것 (기록 — 다시 유도하지 않기 위해)

- **종료 코드 3은 실제로 비어 있고, 계약 파손이 발생할 수 없다.** 브리핑이 "가장 일어나기 쉬운 결함"으로 지목한 항목이다.
  - 저장소 내 유일한 기계적 소비자는 `.github/workflows/spec-lint.yml:58`이고 **인자가 없다** → 해석기가 손대지 않는 경로다
    (SPEC §E가 "해석기는 인자가 주어진 경로에서만 동작한다"로 이미 못박았다).
  - ID 형태와 디렉터리 형태는 **오늘 이미 고장 나 있다**(rc=1 `ParseFailure`, 증거 `.moai/reports/t518/id-arg-control-pair.txt` [1][3]).
    고장 난 경로의 종료 코드에 의존하는 소비자는 존재할 수 없다. 따라서 rc 1→3 전환은 어떤 기존 소비자도 깨지 않는다.
  - 경로 형태는 REQ-SLI-004가 동결한다.
  - `internal/cli`에서 코드 3을 내는 지점은 `spec_lint.go:49` 하나뿐이며, `exitcode_contract_test.go:58`이
    `--json --sarif` 경우를 고정하고 있고 이 변경이 건드리지 않는다.
- **형제 SPEC과의 파일 반경은 실제로 겹치지 않는다.** 형제의 `spec.md`/`plan.md`가 지명하는 파일은
  `internal/spec/lint_req_widen.go`(4회)와 `internal/spec/lint.go`(3회)뿐이고, 이 SPEC은 `internal/cli/*`만 지명한다.
  다만 경계는 **스쳤다** — 판별식 정규식의 SSOT가 `internal/spec`에 있고 그것을 쓰려면 그 파일을 편집해야 한다.
  이 SPEC은 편집을 금지(REQ-SLI-008)하고 사본으로 우회하므로 경계는 지켜지되, 그 대가가 D-2다.
- **범위 확대(파서 결함) 침범 없음.** `internal/spec/parser.go`를 고치겠다는 주장이 없고,
  §E가 명시적으로 범위 밖에 두며, plan.md 어떤 마일스톤도 `internal/spec`을 산출물로 갖지 않는다.
- **자문 경고 8건을 남긴 거부는 SPEC에 기록되어 있다**(§A 부수 발견 + §E 마지막 절 + §G #4). 브리핑이 물은
  "기록인가 구두 주장인가"의 답은 **기록**이다. 기록의 기전이 틀렸을 뿐이다(D-3).
- **종료 코드 규율 준수.** rc를 단언에 쓰는 자리는 AC-SLI-003 하나이고, 그 자리에서 `ParseFailure` 0건 +
  시도 경로 문자열 포함 + rc=3 세 판별식을 함께 쓰며 "파이프 없이 판독"을 본문에 적었다.
- **대조쌍은 이 SPEC의 가장 강한 부분이다.** 001a/001b(해석 vs 동등성), 003/005(부재 ID vs 부재 경로 — 판별식이
  두 모양을 실제로 가르는지 재는 반대편)가 진짜 쌍이고, 002는 001b와 비교하며, 004/005가 대조군 무변형을
  실재·부재 양쪽에서 잡는다. AC-SLI-005의 존재 이유 서술은 특히 정확하다.
- **파서 사각지대 진술은 재현된다** — `parser.go:67`의 `strings.Contains(strings.ToLower(trimmed), "acceptance")`,
  `parser.go:218`의 `^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z]…)?)\s*:\s*`. 둘 다 SPEC 서술과 일치한다.

---

## 검사하지 않은 것 (Gaps — 명시)

- **t528 카드의 실재를 큐에서 확인하지 않았다.** D-13은 브리핑 진술에 의존한다.
- **Go 테스트를 하나도 실행하지 않았다.** 브리핑의 read-only 제약을 지켰다. `internal/cli` 현행 GREEN 여부
  (plan.md §D가 baseline으로 인용하겠다고 한 것)는 미측정이다.
- **`findProjectRootFn`의 구현(어디까지 거슬러 올라가는지)을 읽지 않았다.** D-4는 `view`가 cwd가 아닌 루트 기준이라는
  사실에만 의존하며, 그 사실은 `spec_view.go:51-57`에서 직접 읽었다.
- **D-10의 4,366 재측정은 `--strict` 없이 돌렸다.** 원본 baseline(`baseline-lint.txt` 꼬리 `0 error(s), 4346 warning(s)`)과
  같은 출력 모양이라 비교 가능하다고 판단했으나, 두 실행의 플래그가 동일함을 직접 확인하지는 않았다.
  귀속(+20 = 12 + 8)은 형제 판정문 D-2의 독립 측정과 일치한다.
- **형제 SPEC의 본문은 반경·비중첩 판정에 필요한 범위(파일 지명, status)만 읽었다.** 형제의 품질은 재판정하지 않았다.
- **배포 사용자 프로젝트의 CI/훅 소비자는 원리적으로 확인 불가하다**(D-9의 미검증 부분).
- **뮤턴트를 실제로 심어 보지 않았다.** D-1(b)·D-7의 "뮤턴트가 RED를 만들지 못한다"는 코드를 읽어 도출한 것이며,
  실행으로 확인하지 않았다 — 다만 근거가 정규식과 `ValidateSpecID` 본문이라 실행 없이도 결정적이다.

---

## Recommendation (수리 순서 — 결함 델타로 재감사 가능)

1. **D-1을 먼저 정한다.** 판별식 순서(ValidateSpecID를 앞에 둘 것인가)가 걸려 있어 다른 수리보다 앞선다.
   여기서 무엇을 고르든 plan.md §B D1의 순서 서술이 함께 바뀐다.
2. **D-2**를 plan.md D1에 반영한다 — 동명 심볼 명시 + import 불가 확정 + 새 이름 지정. run-phase에 남기면
   가장 나쁜 방향(과다 수용)으로 조용히 해결될 수 있는 갈림길이다.
3. **D-4**를 정한다(기준 디렉터리). 결정이 REQ-SLI-001의 SHALL 문장을 바꾼다.
4. **D-5·D-6**을 함께 처리한다 — AC에 REQ 인용을 붙이면서 REQ-SLI-007용 AC를 신설한다.
5. **D-7·D-8** — AC-SLI-008에 base SHA와 명령을, 004·005·007에 뮤턴트를 붙인다.
6. **D-3·D-9·D-10** — 산문의 사실관계를 정정한다. 원문을 지우지 말고 정정을 나란히 붙인다
   (D-3과 D-9는 둘 다 결론이 옳고 논거만 틀렸다).
7. D-11·D-12·D-13은 optional. 오케스트레이터 재량.

Tier S 천장이 1회이므로 이 판정 뒤의 재감사는 **위 델타 범위로 한정**하며, 재감사 여부와 범위는
오케스트레이터가 정한다. 판정 권한은 이 에이전트에 있고, 낮은 천장이 오케스트레이터의 자체 평가로
감사 판정을 대신하는 것을 허용하지 않는다.

---

## Residual-risk

- D-1의 수리 (ii)(도달 불가 심층 방어로 명시)를 고르면 경로 순회 방어가 **문서상으로만** 존재하게 된다.
  판별식이 나중에 느슨해지면(예: 소문자 ID 허용) 그때 비로소 도달 가능해지는데, 그 시점에 이 AC는 이미 약화돼 있다.
- D-4를 (ii)(cwd 기준 유지)로 고르면 rc=3이 "그런 SPEC 없음"과 "여기서는 안 보임"을 다시 뭉갠다 —
  이 SPEC이 §A에서 결함으로 지목한 [1]/[4] 미구별과 같은 모양의 축소판이다.
- 이 SPEC의 8건 자문 경고는 의도적으로 남는다. 코퍼스 전역 재계수(형제 SPEC 소관)가 이 8건을 포함해 세면
  두 카드의 수치가 서로 오염될 수 있다 — 형제 판정문 D-2가 같은 축을 이미 지적했다.
- 판정 근거 대부분이 정적 판독(정규식·함수 본문·호출 순서)이다. run-phase가 D1의 순서를 plan.md 권고와 다르게
  구현하면 D-1의 귀결 (a)는 성립하지 않을 수 있다 — 그 경우에도 귀결 (b)(요구의 공허함)는 남는다.
