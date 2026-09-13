# SPEC Review Report: SPEC-INIT-QUIET-WIZARD-001
Iteration: 3/3 (최종)
Verdict: FAIL
Overall Score: 0.87

작성자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정 근거는 SPEC 산출물 5종(spec.md v0.1.4, plan.md, acceptance.md, design.md, research.md), 보조 산출물 progress.md·spec-compact.md, 워크트리 트리(HEAD `120436f58`)뿐이다.

감사 경로: Claude 단독. `grep -rn audit_model .moai/config/sections/` 는 매치가 없다(exit 0, 출력 없음). 1·2회차와 같으므로 MCP 교차 백엔드는 호출하지 않았다.

이번 회차는 Retry Loop Contract 에 따라 2회차 결함 차분(D9~D16)과 D1~D8 회귀 확인, v0.1.4 가 새로 쓴 명령 블록을 대상으로 했다. 점수는 0.86 에서 0.87 로 올랐으므로 STOP 신호는 내지 않는다. Tier L 기준 0.85 는 넘지만, D9 의 결함 부류가 다른 축으로 남아 blocking 결함 D17 이 되었으므로 FAIL 이다. 마지막 회차이므로 아래 권고에 PASS-with-debt 판단 자료를 따로 적었다. 결정은 리드 몫이다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L78-93` 에 `REQ-IQW-001`~`016` 이 있다. `grep -oE '^- \*\*REQ-IQW-[0-9]+' spec.md | sort | uniq -c` 결과 16개 모두 1회씩이고, 빈 번호가 없으며 3자리 채움이 일정하다.
- [PASS] MP-2 EARS/GEARS 형식 — **요구사항 계층(spec.md REQ)** 으로 판정했다. 16개 모두 GEARS 패턴과 구조가 맞는다. v0.1.4 가 고친 REQ-IQW-011(`L88`)은 여전히 Ubiquitous 다("…실행되어야 한다(shall)"). 덧붙은 순서 근거 문장은 설명절이며 형식을 깨지 않는다. `acceptance.md` 의 Given-When-Then 은 검증 계층이므로 이 기준으로 판정하지 않았다.
- [PASS] MP-3 YAML frontmatter: `spec.md:L2-13` 에 12필드가 모두 있다. `version: "0.1.4"`(따옴표 semver), `created`/`updated: 2026-09-11`, `priority: P1`, `phase: "v3.2.0 target"`(금지값 아님), `lifecycle: spec-anchored`, `tags` 는 쉼표 문자열이다. 거부 별칭은 없다. `moai spec lint .moai/specs/SPEC-INIT-QUIET-WIZARD-001` → `✓ No findings — all SPEC documents are valid`, `lint-exit=0`. 판정 빌드는 아래 Gaps 참조. frontmatter 판정은 파일을 직접 읽은 결과이며 lint 에 기대지 않는다.
- [N/A] MP-4 언어 중립성: `module: "internal/cli/wizard, internal/cli"`(`L11`). §7(`L154-166`) 변경 대상에 `internal/template/templates/**` 가 없다. Go 구현 내부 전용이다.
- [PASS] MP-5 D7 교차 SPEC: D7 검증 명령을 모든 산출물(`*.md`)에 돌렸다. 외부 참조 9건이 모두 존재한다. 8건은 `completed`, `SPEC-V3R5-INIT-WIZARD-EXPANSION-001` 은 `implemented` 다(자기 자신은 `draft`). retired·superseded·archived 는 없다. BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: 7개 파일 모두 `grep -c syscall` = 0 이다. 자동 PASS.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` 는 매치가 없다(`mp7-exit=1`).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75–1.0 | D12 가 해소됐다(`spec.md:L106` §4.2 머리말에 대상 범위 문장). 남은 모호성은 둘이다. REQ-IQW-012(`L89`)는 "각 테스트는 §4.2 점검표를 코드에 갖춘다"고 적는데, §4.2 머리말은 새 게이트 테스트에 일부 항목만 요구한다(D20). `acceptance.md:L436` 은 AC-005 가 "세 파일만 본다"는, v0.1.4 이전 전제를 그대로 적고 있다(D18). |
| Completeness | 0.95 | 1.0 | HISTORY `L18-24`(v0.1.4 기록 포함), §1 배경 `L26`, §2 범위 `L39`, §3 요구 `L76`, §4 제약 `L95`, §8 Exclusions `L168` 에 `### Out of Scope — …` H3 8개와 각 `-` 항목이 있다. Tier L 산출물 5종이 있다. §5 완료 정의(`acceptance.md:L515`)에 v0.1.4 가 새로 둔 뮤턴트 E·F 기록이 빠져 있다(D19, 선택). |
| Testability | 0.80 | 0.75–1.0 | D10·D13·D14·D15·D16 은 해소됐다. D10 은 감사자가 다시 쟀다: 기준 추출물 4줄, 현재 트리 0줄에 `body-diff-exit=1`(옳은 이유의 RED-now), 필드 누락 모의는 `mutE-exit=1`, 값 반전 모의도 1, 같은 파일끼리는 `identity-exit=0`. 그러나 AC-005 는 여전히 REQ-012 보다 좁다. 뮤턴트 탐침에서 요구를 어기면서 판정은 통과하는 변형을 다시 쓸 수 있다(D17). |
| Traceability | 0.90 | 0.75–1.0 | 추적표 `acceptance.md:L20-38` 에서 REQ 16개가 모두 AC 에 닿고, AC 17건(`grep -c '^### AC-IQW-'` = 17)이 모두 실재 REQ 를 가리킨다. REQ-011 의 두 절은 이제 짝이 있다. 기본값 절은 AC-004 본문 보존 관측(`L102, L107`)에, 착지 순서 절은 REQ 안에 명시된 구조적 근거(`spec.md:L88`)에 닿는다. REQ-012 의 "본문을 다시 쓰는" 부류는 AC-005 대상 목록에 들어갈 길이 없다(D17). |

집계는 조화평균(0.85, 0.95, 0.80, 0.90) ≈ **0.87** 이다. Tier L 기준 0.85 를 넘지만 blocking 결함 D17 이 남아 FAIL 이다.

## Defects Found (structured defect-list)

2회차 결함의 처분은 아래 회귀 확인에 있다. 여기에는 이번 회차에 남은 결함과 새로 찾은 결함만 적는다.

D17. AC005-SWEEP-NARROWER-THAN-REQ012 — acceptance.md:L161-L191 (특히 L163, L169-L175), spec.md:L89 — Severity: minor — Class: **blocking** — Required fix: 문구만 고치면 된다(아래).
  - 무엇이 문제인가: REQ-IQW-012 의 의무 단위는 **테스트**다. 대상은 "이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트"이고, 요구 문장은 "각 테스트는 §4.2 점검표를 코드에 갖춘다"이다. AC-005 의 Given(`L161`)도 "새로 쓰거나 본문을 다시 쓴 … 파일"을 대상으로 적는다. 그런데 Then 의 대상 목록(`L163`, 명령 `L169-L175`)은 **이 카드가 추가한 파일**(`--diff-filter=A` 커밋분 + `ls-files --others` 미추적분) ∪ 계획한 세 파일로만 이루어진다.
  - 그 결과 "본문을 다시 쓰는" 부류는 구조적으로 대상 목록에 들어갈 수 없다. 이것은 AC-005 안의 Given 과 Then 이 어긋나는 내부 모순이다.
  - 2회차 뮤턴트(네 번째 새 파일)는 이제 잡힌다. 그러나 같은 결함 부류의 뮤턴트를 다른 축으로 다시 쓸 수 있다.
    - (i) **기존 파일 안의 새 테스트**: 기존 파일(예: `init_workflow_wiring_test.go`)에 `TestRunInit_QuietWizardUnsetResolvesToDefaults` 를 새로 두고 기존 헬퍼 `runInitForAutonomyAtHomeCapturingOut(` 을 부른다. 그 파일은 추가분이 아니므로 대상 목록에 들지 않는다. 금지 토큰 검사는 통과하고, AC-006 의 `-run` 은 파일과 무관하게 이름으로 골라 PASS 한다. REQ-012 는 어긴다.
    - 기존 파일에는 이미 금지 토큰이 정당하게 들어 있어 파일 단위 검사를 걸 수도 없다. 감사자 재측정 기준으로 금지 토큰 파일 목록에 `init_agent_wizard_test.go`, `init_autonomy_wiring_test.go`, `init_workflow_wiring_test.go` 가 들어 있다.
    - (ii) **본문을 다시 쓴 기존 테스트**: (i)과 같은 이유로 대상 목록에 들 수 없다. plan.md §H(`L178-L185`)는 기존 테스트에 필드 참조 제거만 계획한다. 계획 밖으로 나가도 AC 가 그것을 보지 못한다.
    - (iii) **stage 했지만 커밋하지 않은 새 파일**: `git ls-files --others` 는 index 에 올라간 파일을 내지 않고, `git diff develop...HEAD` 는 커밋만 비교한다. 따라서 `git add` 만 한 네 번째 파일은 두 목록 모두에서 빠진다. 이것은 git 동작의 판독이며 실행하지 않았다(Gaps). `L163` 의 "커밋분과 미커밋분" 주장은 이 상태를 덮지 못한다.
    - (iv) **판별식 밖의 래퍼 이름**: 새 래퍼 `runQuietInit(` 를 부르는 네 번째 파일은 `runInit[A-Za-z]*\(` 에 걸리지 않는다. `L191` 이 이 한계를 스스로 적고 수작업 규율에 맡겼으므로, 이것만으로는 blocking 사유로 보지 않는다.
  - 실해 범위: 실제 홈 쓰기 누출은 AC-015 슬롯 지문이 별도로 덮는다. 따라서 이 결함의 해악은 REQ-012 코드 점검표의 미준수가 판정을 통과하는 데까지다.
  - 수정(문구만, REQ 는 그대로 둔다):
    1. 대상 목록을 "카드가 추가한 최상위 `internal/cli` 테스트 파일 전부"로 넓힌다. 호출자 교집합은 뺀다. 이렇게 하면 (iv)도 함께 닫힌다.
       - pathspec 은 하위 패키지를 빼야 한다. 감사자 측정에서 `git ls-files -- 'internal/cli/*_test.go'` 는 `internal/cli/wizard/…`, `internal/cli/harness/…` 까지 잡는다. 따라서 `':(glob)internal/cli/*_test.go'` 를 쓰거나 `':!internal/cli/*/**'` 를 더한다.
       - 커밋 전 stage 분을 위해 `git diff --cached --name-only --diff-filter=A -- <pathspec>` 한 줄을 추가분에 합친다(iii).
    2. 기존 파일에 새 테스트 함수가 생기지 않았음을 따로 판정한다(i·ii).
       - 명령: `git diff -U0 --diff-filter=M develop...HEAD -- ':(glob)internal/cli/*_test.go' | command grep -c '^+func Test'` → 0 을 기대하고, 커밋 전분은 `git diff -U0 HEAD -- ':(glob)internal/cli/*_test.go' | command grep -c '^+func Test'` 로 잰다.
       - 필드 참조 제거와 테스트 삭제는 `+func Test` 줄을 만들지 않는다.
       - 과거에 기존 테스트 파일에 테스트 함수를 추가한 커밋 하나를 기지 RED 입력으로 plan 시점에 기록한다.
       - §4.2 머리말에 "이 SPEC 이 새로 쓰는 init 실행 테스트는 새 파일에만 둔다"는 한 문장을 둬, 이 판정이 재는 대상을 명시한다.

D18. AC016-STALE-AC005-PREMISE — acceptance.md:L436 — "AC-IQW-005 의 `t.Parallel()` 검사는 `internal/cli` 파일 세 개만 보므로" 는 v0.1.3 까지의 전제다. v0.1.4 에서 AC-005 는 검증 시점 스윕으로 바뀌었다. 결론("`internal/core/project` 게이트 테스트를 덮지 못한다")은 여전히 참이지만 근거 문장이 낡았다(`verification-completeness.md` §3 교차 계층 개정 스윕 누락). — Severity: minor — Class: optional — Required fix: "AC-IQW-005 의 검사는 `internal/cli` 테스트 파일만 보므로" 로 고친다.

D19. COMPLETION-DEF-MISSING-MUTANT-E-F — acceptance.md:L515 — §5 완료 정의가 요구하는 뮤턴트 증거 목록은 "AC-IQW-004 배선 제거 뮤턴트, AC-IQW-007b 코드 뮤턴트, AC-IQW-016 뮤턴트 C1·C2·D" 다. v0.1.4 가 기록 의무로 새로 둔 AC-004 뮤턴트 E(`L110`)와 AC-005 뮤턴트 F(`L193-L197`)가 빠져 있다. plan.md M6(`L95-L99`)도 AC-007b 뮤턴트만 적는다. AC 본문이 의무를 지므로 판정 자체는 막히지 않는다. — Severity: minor — Class: optional — Required fix: `L515` 와 plan.md M6 에 뮤턴트 E·F 를 더한다.

D20. REQ012-VS-4.2-GATE-TEST-SCOPE — spec.md:L89, L106 — REQ-IQW-012 는 이 SPEC 이 새로 쓰는 init 실행 테스트 전부에 "§4.2 점검표를 코드에 갖춘다"를 요구한다. §4.2 머리말은 새 `internal/core/project` 게이트 테스트에 `MOAI_HOME` 우회·스파이·원복만 요구한다. 더 구체적인 §4.2 를 따르는 쪽으로 합리적 엔지니어는 같은 해석에 이르므로 실해는 작다. 다만 REQ 문장이 개정되지 않아 두 계층이 문자 그대로는 다르게 말한다. — Severity: minor — Class: optional — Required fix: REQ-012 에 "코드 점검표의 적용 범위는 §4.2 머리말을 따른다"는 한 구절을 둔다.

## Regression Check (Iteration 2+)

### 2회차 결함 D9~D16

| ID | 처분 | 근거 |
|---|---|---|
| D9 | **PARTIALLY RESOLVED** | 2회차가 적시한 뮤턴트(계획 밖 네 번째 새 파일)는 이제 잡힌다. `L169-L181` 스윕이 미추적 추가분 ∩ 호출 파일을 대상 목록에 넣고, 도달 대조군(`L177-L178`)과 금지 토큰 교집합(`L179-L181`)으로 판정한다. 기존 헬퍼 호출이 금지 토큰에 들어갔고, 뮤턴트 F 기록 의무가 생겼다(`L193-L197`). 감사자 재측정(HEAD `120436f58`, 이 세션): 커밋 추가분 출력 없음, 미추적 추가분 출력 없음, 호출 파일 15개(`coverage_improvement_test.go` … `update_mode_test.go`, 모두 최상위). 이 값들은 `L201` 의 plan 시점 대조군 "추가분 0, 호출 파일 15" 와 일치한다. 그러나 결함 부류 — AC-005 대상 목록이 REQ-012 의 대상보다 좁음 — 는 다른 축(기존 파일 안의 새 테스트, 본문 재작성, stage 만 한 파일)으로 남았다 → D17. |
| D10 | **RESOLVED** | AC-004 에 본문 보존 관측이 생겼다(`L102, L107, L119-L123, L136-L137`). 감사자 재측정: `git show 120436f58:…initializer.go` 추출 → `4` 줄. 기준 본문(`initializer.go:677-685` 에 해당)을 직접 읽은 결과와 옵션 4줄이 같다. 현재 트리 추출 → `0` 줄, `body-diff-exit=1`(`defaultConfigureShellEnv` 부재 — 옳은 이유의 RED-now, M1 이 뒤집음). `PreferLoginShell` 을 뺀 모의 → diff `4d3 / < 		PreferLoginShell:        true,`, `mutE-exit=1`. `true`→`false` 값 반전 모의 → `mutG-value-flip-exit=1`. 동일 파일 → `identity-exit=0`. design.md §4.2(`L110-L119`)가 `defaultConfigureShellEnv` 를 `initializer.go` 안의 최상위 `func` 로 두므로 `^func defaultConfigureShellEnv` 추출 범위가 맞는다. 형태가 달라지면 추출이 비거나 넘쳐 diff 가 1 이 되는 쪽(시끄러운 실패)으로 기운다. 남는 뮤턴트(옵션 리터럴은 두고 `Configure` 호출 방식·로거·조기 반환을 바꿈)는 `L107` 이 스스로 한계로 적었고, 2회차 Residual-risk 가 같은 한계를 이미 받아들였다. |
| D11 | **RESOLVED** | `spec.md:L88` 이 순서 절을 유지하면서 "스파이를 끼우는 테스트는 시접 변수가 없으면 컴파일되지 않으므로 이 순서는 구조적으로 강제되며, 그래서 인수 기준에서 따로 재지 않는다"는 근거를 적었다. 2회차가 제시한 두 수정 가운데 하나다. |
| D12 | **RESOLVED** | `spec.md:L106` §4.2 머리말: "대상은 `internal/cli` 의 `runInit` 실행 테스트이며, `internal/core/project` 게이트 테스트는 `MOAI_HOME` 우회·스파이·원복만 갖추고 실제 홈은 §4.3 슬롯 지문이 덮는다". REQ 문장과의 문자 차이는 선택 결함 D20 으로 따로 적었다. |
| D13 | **RESOLVED** | `acceptance.md:L451`: "대입 줄 수가 뮤턴트 전보다 1 줄어들고 … 이 뮤턴트의 판정은 실행 RED 로 한다". 명령 주석 `L470` 도 적용 전후 `-c` 비교를 지시한다. |
| D14 | **RESOLVED** | `acceptance.md:L127-L133` 에 뮤턴트 B 실행·FAIL 계수, 원복 뒤 PASS·`no tests to run` 계수가 있다. `L141` 기대값과 명령이 대응한다. |
| D15 | **RESOLVED** | `acceptance.md:L417-L418` 이 슬롯 행에 고정됐다(`^\| SLOT-.*home-diff-exit=0`, `…=[1-9]`). 지문 출력을 슬롯 행 밖(펜스 블록 등)에 두면 셋째 수가 줄어 삼자 불일치로 실패한다. 공허 통과가 아니라 시끄러운 실패 쪽이다. |
| D16 | **RESOLVED** | `acceptance.md:L93` 에 RED/GREEN 대조가 기록됐다. 감사자 재측정(이 세션): `git diff --quiet 93182d137...develop -- internal/cli/update_wizard.go` → `three-dot-known-red-exit=1`, `…/wizard/questions.go` → `three-dot-known-green-exit=0`, `git merge-base --is-ancestor c4990eea7 develop` → `0`, `git rev-parse develop` → `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`(`L93` 인용과 일치), `git merge-base develop HEAD` → `93182d137159c4facbf87c66dea3fd69160a6b8b`. |

### 1회차 결함 D1~D8 — 회귀 없음

- D1 — 유지. AC-004 실행 관측 구조가 그대로다(`acceptance.md:L95-L157`). 본문 보존 관측이 더해졌을 뿐 주·게이트·보조 관측과 배선 뮤턴트 A·B 는 바뀌지 않았다.
- D2 — 유지. `spec.md:L89` REQ-012 8항목 한정과 "8항목 밖은 대상 아님" 문장이 있다.
- D3 — 유지. 스윕 확인 규칙(`acceptance.md:L15`)이 있고, 새로 쓴 명령 줄(AC-004 `L131-L133`, AC-005 `L167-L168`)에도 PASS 계수와 `no tests to run` 0 판정이 붙어 있다.
- D4 — 유지. REQ-IQW-015(`spec.md:L92`) ↔ AC-008(`acceptance.md:L30`).
- D5 — 유지. `spec.md:L148` 의 `workflow.audit` 하위 키 서술.
- D6 — 유지. `spec.md:L80` 의 파일시스템 루트 예외절.
- D7 — 유지. REQ-011 은 Ubiquitous, REQ-009 는 AC-006 으로 절차를 넘기고(`L86`), REQ-014/016 은 나뉘어 있다.
- D8 — 유지. AC-015 삼자 일치 `acceptance.md:L394, L421`.

정체 탐지: 세 회차 모두에 같은 모양으로 남은 결함은 없다. D17 은 2회차 D9 와 같은 부류이며, 2회차 수정 제안 자체가 파일 단위였기 때문에 남은 것이다. 저자가 진전을 내지 않은 경우가 아니다.

## 새 표면 검토 (v0.1.4 명령 블록, verification-completeness §1.1·§2)

| 표면 | §1.1 빈 스윕 | §2 두 칸 (RED-now / green path) | §2 뮤턴트 탐침 |
|---|---|---|---|
| AC-004 본문 보존 관측 (`L119-L123`, `L136-L137`) | 기준 추출물 `wc -l` ≥1 대조군(`L122, L141`). 감사자 재측정 4 | RED-now 0줄·`body-diff-exit=1`, 이유 명시(`L147`). green path M1. 감사자 재측정 일치 | 필드 누락·값 반전은 잡힘(감사자 모의 exit 1). 호출 방식 변경은 못 잡음(`L107` 선언, 잔여 위험) |
| AC-004 뮤턴트 B·원복 (`L127-L134`) | 원복 줄에 PASS 계수·`no tests to run` 0. 뮤턴트 줄은 FAIL 계수 1(0 이면 선택자 실패 가능) | 판정 대상이 기록 증거라 RED-now 칸 해당 없음 | 흠잡을 뮤턴트 없음 |
| AC-005 스윕 (`L169-L181`) | 대상 목록 수 ≥3, 도달 대조군 = 목록 수(`L186`). plan 시점 도달 0 → RED(`L203`). 기지 입력으로 형식 확인(`L206`) | RED-now 옳은 이유(파일 부재). green path M1·M2 | **D17** — 기존 파일 안의 새 테스트, 본문 재작성, stage 만 한 새 파일이 대상 목록 밖. 래퍼 이름은 선언된 한계 |
| AC-015 마감 계수 (`L415-L418`) | 세 수 ≥1 | 판정 대상이 기록 증거라 RED-now 칸 해당 없음 | 슬롯 행 고정으로 서술문 채움 뮤턴트 차단(D15) |
| AC-016 C1 기대 (`L451`) | 변화 없음 | 변화 없음 | 실행 RED 로 판정 이동, 텍스트 가정 제거(D13) |

## Gaps (관측하지 않은 것)

- `go test`, `moai init`, `moai update`, 빌드는 실행하지 않았다(리드 지시). 스파이 호출 횟수, `-count=2` 재진입, 뮤턴트 D panic, 뮤턴트 A·B·C1·C2 의 RED 는 모두 판독이다.
- **lint 판정 빌드 좌표(귀속 한계, 재론하지 않음)**: `command -v moai` → `/Users/goos/go/bin/moai`, 배너 `moai-adk v3.2.0-rc.7`. 이 설치본의 커밋 `ed71054d3`(progress.md:14 기록, dirty)은 HEAD `120436f58` 과 어느 방향으로도 조상 관계가 없다. 이 세션에서 다시 쟀다: `git merge-base --is-ancestor ed71054d3 HEAD` → 1, `git merge-base --is-ancestor HEAD ed71054d3` → 1. 따라서 `lint-exit=0` 은 이 트리로 만든 빌드의 판정이 아니다.
- D17 (iii) — "stage 한 새 파일은 `git ls-files --others` 와 `git diff develop...HEAD` 어느 쪽에도 나오지 않는다"는 git 동작 판독이다. index 를 바꾸는 실험은 읽기 전용 제약 때문에 하지 않았다.
- D17 (i) — 뮤턴트 파일을 실제로 만들어 스윕을 돌리지 않았다(트리 쓰기 금지). 판정은 명령 구성의 판독과, 감사자가 잰 기존 파일의 금지 토큰 목록에 근거한다.
- D9 해소 판정의 "네 번째 새 파일은 잡힌다" 역시 명령 구성 판독이다. 형식이 실제로 빨개질 수 있다는 기지 입력 관측은 저자 기록(`L206`)이며, 감사자가 같은 교집합을 다시 만들어 보지는 않았다.
- design.md 가 인용한 `internal/cli/init.go` 행번호는 이번 회차에 다시 읽지 않았다. `internal/` 는 이 워크트리에서 수정되지 않았다(gitStatus 스냅숏: 미추적 두 디렉터리뿐).
- MCP 교차 백엔드는 설정 부재로 호출하지 않았다.

## Residual-risk

- 금지 토큰 검사는 텍스트 기반이다. 상수로 우회한 `t.Setenv(envHome, …)` 나 `t` 가 아닌 수신자 이름(`tb.Parallel()`)은 잡지 못한다. 병렬 쪽은 `t.Setenv` 충돌 panic(AC-016)이 실행으로 받치지만, HOME 상수 우회는 코드 리뷰에 기댄다.
- D10 수정은 옵션 리터럴 보존만 잰다. `defaultConfigureShellEnv` 가 리터럴은 남긴 채 로거·구성기·조기 반환을 바꾸는 회귀는 스파이 관측과 포인터 비교를 모두 통과한다(`L107` 선언).
- `L191` 의 판별식 이름 의존은 run 단계가 래퍼를 새로 둘 때 규율로만 지켜진다. D17 수정 1(추가분 전부를 대상으로)을 택하면 사라진다.
- 셸 설정 줄이 이미 있는 머신에서는 슬롯 지문 불변이 누출 부재의 강한 증거가 못 된다(`L431`, 저자 인정).

## Recommendation

must-pass 7개는 모두 통과했다. 점수는 0.79 → 0.86 → 0.87 로 계속 올랐고, 2회차 결함 8건 중 7건이 해소됐다. FAIL 사유는 blocking 결함 한 건(D17)이다. 이번이 마지막 회차이므로 Retry Loop Contract 에 따라 리드가 사용자에게 선택지를 올린다.

### (a) 남은 blocking 결함

1. **D17** — AC-005 대상 목록이 REQ-012 의 단위(테스트)보다 좁다(파일 추가분만 봄). 기존 파일 안의 새 테스트, 본문을 다시 쓴 기존 테스트, stage 만 한 새 파일이 판정을 통과한다. AC-005 안에서도 Given("본문을 다시 쓴")과 Then(추가분만)이 어긋난다.

### (b) 문구만으로 고칠 수 있는가

- **D17 — 예, 문구만으로 고칠 수 있다.** REQ 는 그대로 두고 `acceptance.md` AC-005 의 Then 과 명령 블록만 바꾸면 된다.
  - 수정 1 — 추가분 전부를 대상으로 삼는다. 최상위로 한정한 pathspec `':(glob)internal/cli/*_test.go'` 를 쓰고, `--cached --diff-filter=A` 줄을 더하며, 호출자 교집합은 뺀다.
  - 수정 2 — 수정된 기존 파일의 `^+func Test` 계수가 0 인지 판정한다. 기지 RED 입력을 plan 시점에 한 번 기록한다.
  - 선택 — §4.2 머리말에 "새 init 실행 테스트는 새 파일에만 둔다" 한 문장을 둔다.
- D18·D19·D20(선택) 도 모두 한두 구절의 문구 수정이다.

### (c) PASS-with-debt 적격성 (Retry Loop Contract)

- **적격 판단 자료**:
  - must-pass 전부 PASS.
  - 점수 0.87 로 Tier L 기준 0.85 이상이며, 회차마다 올랐다(회귀 없음).
  - 남은 blocking 결함은 1건이고 severity 는 minor, 문구 수정만 필요하다.
  - 해악은 REQ-012 코드 점검표 미준수가 판정을 통과하는 데 그친다. 실제 홈 쓰기 누출은 독립 관측인 AC-015 슬롯 지문과 REQ-016 멈춤 규칙이 계속 덮는다.
- **부채로 넘길 경우 조건**:
  - run 단계 위임 프롬프트(Section D)에 "새 init 실행 테스트는 새 파일에만 두며, 기존 테스트 파일에는 필드 참조 제거 외 테스트 함수를 추가하지 않는다"를 제약으로 명시한다.
  - M6 마감 시 `git diff -U0 --diff-filter=M develop...HEAD -- ':(glob)internal/cli/*_test.go' | command grep -c '^+func Test'` = 0 을 progress.md §E.2 에 기록하는 것을 부채 상환 조건으로 둔다.
- **대안**:
  - 문구 수정 후 확인 재감사는 4회차가 되므로 명시적 사용자 연장이 필요하다. 그 경우 재감사 범위는 D17 차분 하나로 한정할 수 있다.
  - 범위 축소는 결함 성격상 필요하지 않다.
- 결정과 사용자 질의는 리드 몫이다. Implementation Kickoff Approval 은 이 판정과 무관하게 필수다.
