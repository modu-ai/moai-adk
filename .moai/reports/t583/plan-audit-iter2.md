# SPEC Review Report: SPEC-INIT-QUIET-WIZARD-001
Iteration: 2/3
Verdict: FAIL
Overall Score: 0.86

작성자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정 근거는 SPEC 산출물 5종(spec.md v0.1.3, plan.md, acceptance.md, design.md, research.md), 보조 산출물 progress.md·spec-compact.md, 워크트리 트리(HEAD `120436f58`)뿐이다. `git status --short` 출력은 `?? .moai/reports/t583/`, `?? .moai/specs/SPEC-INIT-QUIET-WIZARD-001/` 두 줄이고 `internal/` 은 수정되지 않았다.

감사 경로: Claude 단독. 1회차 판정을 옮겨 쓰지 않고 다시 쟀다. `.moai/config/sections/` 전체에서 `audit_model` 은 0건이다. `workflow.yaml:20` 의 `audit:` 블록에는 `codex`·`glm` 모델 핀만 있고 `model:` 키가 없다. 컴파일 기본값은 `internal/config/defaults.go:966-973` `Audit: AuditConfig{Model: AuditModelClaude, …}` 다. 그래서 MCP 교차 백엔드는 호출하지 않았다.

점수는 1회차 0.79 에서 0.86 으로 올랐다. 점수 하락이 없으므로 STOP 신호는 내지 않는다. 점수는 Tier L 기준 0.85 를 넘지만, 아래 blocking 결함 2건(D9·D10)이 남아 FAIL 이다. 둘 다 `verification-completeness.md` §2 뮤턴트 탐침에 걸린다. 요구를 어기면서 인수 기준은 통과하는 뮤턴트를 실제로 쓸 수 있다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L77-92` 에 `REQ-IQW-001` 부터 `REQ-IQW-016` 까지 16개가 있다. 3자리 채움이 일정하고 빈 번호와 중복이 없다. Tier L 상한 25 안이다.
- [PASS] MP-2 EARS/GEARS 형식 — 요구사항 계층(`spec.md` 의 REQ)으로 판정했다. 16개 모두 GEARS 패턴과 구조가 맞는다.
  - Ubiquitous: 001, 004, 008, 009, 011
  - Unwanted `shall not`: 002, 012, 013
  - Event-driven `When`: 003, 010, 014, 015, 016
  - State-driven `While`: 005, 006
  - Where: 007
  - 1회차 D7 에서 지적한 REQ-011 의 `Where` 오용은 Ubiquitous 로 바뀌었다(`L87`).
  - Given-When-Then 은 `acceptance.md` 의 AC 계층에만 있으므로 이 기준으로 판정하지 않았다.
- [PASS] MP-3 YAML frontmatter: `spec.md:L2-13` 에 정규 12필드가 모두 있다. `version: "0.1.3"`(따옴표 semver), `created`/`updated: 2026-09-11`, `priority: P1`, `phase: "v3.2.0 target"`(금지값 아님), `lifecycle: spec-anchored`, `tags` 는 쉼표 문자열이다. `moai spec lint .moai/specs/SPEC-INIT-QUIET-WIZARD-001` 출력은 `✓ No findings — all SPEC documents are valid`, `lint-exit=0` 이다.
  - 판정 빌드 좌표(VCI §2.2): lint 는 `command -v moai` 가 가리키는 설치본 `/Users/goos/go/bin/moai` 로 돌았다. 버전은 `v3.2.0-rc.7`, `…-ged71054d3-dirty` 다.
  - `git merge-base --is-ancestor ed71054d3 HEAD` 와 그 역방향 모두 종료 코드 1 이다. 설치본은 트리의 조상이 아니므로 뒤처진 빌드는 아니다. 다만 트리와 갈라진 dirty 빌드라는 점을 잔여 위험에 적었다.
- [N/A] MP-4 언어 중립성: `module: "internal/cli/wizard, internal/cli"`(`spec.md:L11`)이다. §7 변경 대상(`L155-165`)에 `internal/template/templates/**` 가 없다. Go 구현 내부 전용이다.
- [PASS] MP-5 D7 교차 SPEC: 5개 산출물에서 뽑은 외부 참조 9건이 모두 존재한다.
  - completed: CLI-WIZARD-RESTRUCTURE-001, FEEDBACK-AUTO-SUBMIT-001, INIT-HARNESS-PROMPT-001, INIT-WIZARD-REPAIR-001, MCP-DEFAULT-ON-001, MOAI-MCP-SERVER-001, PROJECT-CONTINUATION-KEY-001, TODO-ENABLE-FLAG-001
  - implemented: V3R5-INIT-WIZARD-EXPANSION-001
  - retired·superseded·archived 는 없다. 부분 대체 관계는 `spec.md:L131-141` §5 표가 조정한다. BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: 7개 파일 모두 `grep -c syscall` = 0 이다. 자동 PASS.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` 는 매치가 없다(`mp7-exit=1`). `plan.md:L231-234` §L 은 "남은 확인 항목 없음" 이다.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75–1.0 | 1회차의 모호성 세 곳을 모두 정리했다: REQ-012 범위 `L88`, §6 `audit:` 서술 `L147`, REQ-011 형식 `L87`. 남은 모호성은 하나다. `internal/core/project` 게이트 테스트도 실제 `Init` 을 도는데, 이것이 REQ-012 가 말하는 "init 실행 테스트"(8항목 코드 대조 의무)에 드는지 spec 이 말하지 않는다. plan.md M1 `L61` 이 그 테스트의 의무를 따로 적어 두어, 합리적인 엔지니어라면 같은 쪽으로 해석한다(D12). |
| Completeness | 0.95 | 1.0 | HISTORY `spec.md:L18-23`(v0.1.2·v0.1.3 개정 기록 포함), 배경 §1 `L25`, 범위 §2 `L38`, 요구사항 §3 `L75`, 제약 §4 `L94`, Exclusions §8 `L167` 이 모두 있다. `### Out of Scope — …` H3 8개에 각각 `-` 항목이 있다. Tier L 산출물 5종이 있고, 미측정 목록은 `plan.md:L217-229` 이다. |
| Testability | 0.80 | 0.75–1.0 | 1회차 D1(관측 불가)은 실행 관측으로, D3(빈 스윕)는 전 선택자의 PASS 줄 수와 `no tests to run` 판정으로, D8(약한 마감 판정)은 삼자 일치로 해소됐다. 새 표면인 AC-016·003·005 는 대체로 두 칸 채택 규율을 지켰다. 다만 뮤턴트 탐침에 걸리는 곳이 둘(D9·D10) 있고, 기대값이 구현 모양을 가정하는 곳이 셋(D13·D14·D15) 있다. |
| Traceability | 0.85 | 0.75–1.0 | 추적표 `acceptance.md:L20-38` 에서 REQ 16개가 모두 AC 에 닿고, 모든 AC 가 실재 REQ 를 가리킨다. AC-008 은 새 REQ-015 에 매핑됐다(D4 해소). 다만 REQ-011 의 두 절, 곧 "운영 기본값 = 현재 기록 동작"과 "먼저 착지"를 검증하는 AC 가 없다(D10·D11). |

집계는 조화평균(0.85, 0.95, 0.80, 0.85) ≈ **0.86** 이다. Tier L 기준 0.85 를 넘지만, blocking 결함 D9·D10 이 남아 있어 FAIL 이다.

## Defects Found (structured defect-list)

1회차 결함 D1~D8 은 모두 해소됐다(아래 회귀 확인 참조). 아래는 이번 회차에 새로 찾은 결함이다.

D9. AC005-FIXED-LIST-ADMITS-MUTANT — acceptance.md:L143-145, L150 — AC-IQW-005 의 금지 토큰 검사(`t.Setenv("HOME"`, `t.Parallel()`)와 새 존재·도달 대조군이 고정 목록 세 파일만 본다. REQ-IQW-012(`spec.md:L88`)의 코드 의무는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트 **전부**에 걸린다. — Severity: minor — Class: blocking — Required fix: 아래 참조.
  - 쓸 수 있는 뮤턴트: run 단계가 새 실행 테스트를 네 번째 파일(예: `init_quiet_wizard_mcp_test.go`)로 나누고, 그 파일에서 기존 HOME 헬퍼 `runInitForAutonomyAtHomeCapturingOut` 을 재사용하거나 `t.Parallel()` 을 쓴다. 이 뮤턴트에서 AC-005 세 대조군은 모두 3 이고 금지 토큰 검사는 `exit=1` 이라 통과한다. 그런데 REQ-012 와 §4.2 1·6항은 어긴다. AC-016 은 시접 대입 파일만 스윕하므로 이 파일을 잡지 못한다.
  - 목록을 늘리는 일은 L150 이 수작업 규율로 맡겼다. 하지만 v0.1.3 HISTORY(`spec.md:L23`)는 바로 이 이유로 AC-016 을 "고정 파일 목록이 아니라" 동적 스윕으로 바꿨다. 한 SPEC 안에서 같은 약점을 한쪽은 고치고 다른 쪽은 남긴 것이다.
  - 수정: AC-005 의 대상 목록을 검증 시점에 다시 뽑는다. 예를 들어 카드가 추가한 `internal/cli/*_test.go`(커밋분 `git diff --name-only --diff-filter=A develop...HEAD -- 'internal/cli/*_test.go'` + 미커밋분 `git ls-files --others --exclude-standard -- 'internal/cli/*_test.go'`) 가운데 `runInit(` 을 부르는 파일에 계획한 세 파일을 합친다. 금지 토큰 검사는 그 목록 전체에 걸고, 목록 수 하한은 3 으로 둔다.
  - 계획한 세 파일 중 `init_home_guard_test.go` 는 `runInit` 을 직접 부르지 않을 수 있으므로 하한 대조군에 따로 넣는다.

D10. REQ011-DEFAULT-BODY-UNVERIFIED — spec.md:L87, acceptance.md:L101, L105, design.md:L126, L130 — REQ-IQW-011 은 "시접의 운영 기본값은 현재의 셸 설정 기록 동작이어서 운영 동작은 바뀌지 않으며"를 요구한다. 이 절을 재는 관측은 AC-004 보조 관측, 곧 `ConfigureShellEnvFn` 기본값과 `defaultConfigureShellEnv` 의 `reflect` 포인터 비교 하나뿐이다. 이 비교는 변수가 그 함수를 가리키는지만 보고, 그 함수 본문이 옮기기 전 `configureShellEnv` 본문(`internal/core/project/initializer.go:677-685`, HEAD `120436f58` 에서 재확인)과 같은지는 보지 않는다. — Severity: minor — Class: blocking — Required fix: 아래 참조.
  - 쓸 수 있는 뮤턴트: `defaultConfigureShellEnv` 로 옮기면서 `PreferLoginShell: true` 를 빠뜨리거나 `AddGoBinPath` 를 false 로 둔다.
    - 이 뮤턴트에서 스파이 관측(AC-004 주·게이트), 포인터 비교, 배선 뮤턴트 A·B, AC-016 은 모두 통과한다. 모든 사용자의 `moai init` 셸 설정 동작은 바뀐다.
    - 이 옵션을 단언하는 기존 테스트도 없다. `git grep -nE 'PreferLoginShell|AddGoBinPath|configureShellEnv' 120436f58 -- internal` 결과에서 `internal/core/project` 쪽 테스트 파일은 0건이고, 걸린 곳은 운영 코드와 `internal/shell/env_test.go` 뿐이다.
  - 금지 뮤턴트 규율(`plan.md §D`) 때문에 실제 기록을 실행해 볼 수는 없다. 그러나 여기서 재야 할 것은 도달이 아니라 본문 보존이다. 따라서 spec §4.1(`L101`)이 금지한 "도달 증거로서의 소스 문자열 검사"에 해당하지 않는다.
  - 수정: AC-IQW-004 또는 새 절에 AC-010 방식의 본문 비교를 둔다.
    - 기준 쪽: `git show 120436f58:internal/core/project/initializer.go` 에서 `shell.ConfigOptions{` 부터 닫는 `})` 까지 추출한다.
    - 현재 쪽: HEAD 의 `defaultConfigureShellEnv` 에서 같은 범위를 추출한다.
    - 판정: 두 추출물의 `diff` 종료 코드가 0 이고, 기준 추출물이 비어 있지 않다(plan 시점 대조군 줄 수를 기록).

D11. REQ011-ORDER-CLAUSE-DISCLAIMED — spec.md:L87, acceptance.md:L130, plan.md:L54 — REQ-IQW-011 은 "이 시접은 어떤 init 실행 테스트보다 먼저 착지해야 한다"를 규범으로 둔다. 그런데 AC-004(`L130`)와 plan §F(`L54`)는 이 순서를 "AC 증거로 쓰지 않는다"고 명시한다. 요구 절 하나에 검증이 없다. 스파이를 쓰는 새 테스트는 시접 없이 컴파일되지 않으므로 순서가 구조적으로 강제된다는 점에서 실질 위험은 낮다. — Severity: minor — Class: optional — Required fix: 이 절을 REQ 에서 빼 plan §F 의 공정 규율로만 두거나, REQ 에 "스파이를 쓰는 테스트는 시접 없이 컴파일되지 않아 순서가 구조적으로 강제된다"는 근거를 적는다.

D12. GATE-TEST-SCOPE-AMBIGUITY — spec.md:L88, L103-112, plan.md:L61 — REQ-012 와 §4.2 는 "이 SPEC 이 새로 쓰는 init 실행 테스트"에 8항목 코드 대조와 자기 가드를 요구한다. 그런데 새 `internal/core/project` 게이트 테스트(`initializer_shell_seam_test.go`)도 실제 `Init` 을 실행한다. 그 테스트가 이 의무 대상인지 spec 본문이 정하지 않는다. §4.2 2항의 `userHomeDirFn` 은 그 패키지에서 접근할 수 없다. plan M1 은 그 테스트에 8항목 대조를 주지 않는다. — Severity: minor — Class: optional — Required fix: §4.2 머리말에 "대상은 `internal/cli` 의 `runInit` 실행 테스트이며, `internal/core/project` 게이트 테스트는 `MOAI_HOME` 우회·스파이·원복만 갖추고 실제 홈은 §4.3 슬롯 지문이 덮는다"는 한 문장을 둔다.

D13. AC016-C1-LINE-COUNT-ASSUMPTION — acceptance.md:L403 — 뮤턴트 C1 은 되돌리는 대입을 지운 뒤 "그 파일의 대입 줄 수가 1 이 되어 텍스트 관측이 RED" 라고 기대한다. 이 기대는 게이트 테스트 파일에 바꿔 끼우기 1줄과 되돌리기 1줄만 있다고 가정한다. `SkipShellConfig=false/true` 를 하위 테스트마다 바꿔 끼우면 줄 수가 3 이상으로 남아 텍스트 관측이 RED 가 되지 않는다. 실행 관측(`-count=2`)은 여전히 RED 이므로 판정 자체는 유효하다. — Severity: minor — Class: optional — Required fix: 기대값을 "되돌리는 대입을 지운 뒤 줄 수가 뮤턴트 전보다 1 줄어듦"으로 고치거나, 텍스트 RED 기대를 빼고 실행 RED 만 기대한다.

D14. AC004-MUTANT-B-COMMAND-ABSENT — acceptance.md:L116-119, L122 — 명령 블록에는 뮤턴트 A 출력 파일(`ac004-mutant-a.txt`) 실행과 FAIL 계수만 있다. 기대값은 "뮤턴트 A·B 파일은 각각" 을 판정한다. 뮤턴트 B 의 출력 파일 이름, 원복 뒤 PASS 재실행 명령도 적혀 있지 않다. — Severity: minor — Class: optional — Required fix: AC-007b(`L210-221`)처럼 뮤턴트 B 실행·FAIL 계수와 원복 후 PASS·`no tests to run` 계수 줄을 명령 블록에 적는다.

D15. AC015-UNANCHORED-RECORD-COUNT — acceptance.md:L369-370 — 마감 판정 `command grep -c 'home-diff-exit=0' progress.md` 는 슬롯 행이 아닌 줄도 센다. 예를 들어 §E.2 서술문이나 뮤턴트 기록에 그 문자열이 들어간 줄이다. 슬롯 하나가 기록을 빠뜨려도 서술문 한 줄이 그 수를 메워 삼자 일치가 성립하는 뮤턴트를 쓸 수 있다. 현재 progress.md 의 해당 문자열 수는 0 이다(`grep -c` 결과). — Severity: minor — Class: optional — Required fix: 두 계수를 슬롯 행으로 고정한다(`command grep -cE '^\| SLOT-.*home-diff-exit=0'`, `command grep -cE '^\| SLOT-.*home-diff-exit=[1-9]'`).

D16. AC003-REVERSE-DIRECTION-UNRECORDED — acceptance.md:L93 — AC-003 은 새 세 점 형식이 실제로 빨개지는 방향(카드가 파일을 바꿨을 때 `…-card-diff-exit=1`)을 "plan 단계에서 실행하지 않은 판독" 으로 남겼다(§1.1 관측된 실패). 감사자가 같은 형식을 이미 알고 있는 입력에 대어 재 보았다.
  - 빨간 입력: `git diff --quiet 93182d137...develop -- internal/cli/update_wizard.go` → `three-dot-known-red-exit=1`. 로컬 develop 에 든 t587 `c4990eea7` 때문이다. 조상 관계는 `git merge-base --is-ancestor c4990eea7 develop` = 0 으로 확인했다.
  - 초록 입력: 같은 범위의 `internal/cli/wizard/questions.go` → `three-dot-known-green-exit=0`.
  - 형식은 옳다. SPEC 에 기록만 없다.
  - Severity: minor — Class: optional — Required fix: 위 두 명령과 출력을 AC-003 의 plan 시점 측정 문단에 RED/GREEN 대조로 옮겨 적는다.

## Regression Check (Iteration 2+)

1회차 결함:

- D1 (AC-004 관측 불가) — **RESOLVED**. 보고서가 제안한 두 수정 (a)·(b) 와는 다른 설계로 해소됐다. 결함의 핵심은 "`runInit` 이 기본 게이트를 셸 설정 단계까지 실제로 전달함을 관측할 수 없다"였고, 새 설계는 이것을 실행으로 관측한다.
  - 설계: `internal/core/project` 의 내보낸 함수 변수 시접 `ConfigureShellEnvFn` 하나(`spec.md:L87`, `design.md:L103-130`, `plan.md:L15, L20, L58-65`).
    - Step 6 이 이 변수를 거쳐 호출한다.
    - 테스트는 스파이를 끼운 실제 `runInit` 에서 호출 정확히 1회를 본다(`acceptance.md:L99, L103`).
    - `runInit` 은 `SkipShellConfig` 를 켜지 않으므로 1회는 기본 게이트가 전달됐다는 뜻이고, 0회는 게이트가 켜졌거나 호출이 사라졌다는 뜻이다.
    - 게이트 테스트가 `true` 0회, `false` 1회를 본다(`L100, L104`).
    - 배선 제거 뮤턴트 A(게이트 반전)·B(호출 삭제)의 RED 를 기록한다(`L107`).
  - 1회차가 짚은 부수 문제 해소:
    - 기본값 전달을 실제 기록으로 관측하면 `~/.zshenv` 에 쓰는 문제는 스파이가 기록 함수를 대신해 사라졌다(`design.md:L128`).
    - "셸 설정 파일 불변" 간접 관측은 증거에서 뺐고, 뺀 이유를 적었다(`acceptance.md:L129`).
    - A=B 에도 0 을 내는 조상 판정도 증거에서 뺐다(`L130`).
    - 시접 우회 뮤턴트는 실제 홈에 쓰므로 금지했다(`L124`, `plan.md:L37`).
  - 트리 대조:
    - `initializer.go:333-347`(Step 6 `if !opts.SkipShellConfig` → `i.configureShellEnv()`)과 `:677-685`(`shell.NewEnvConfigurator(i.logger).Configure(...)` 옵션 4개)가 design §4.1·§4.2 서술과 같다.
    - `git grep -c 'ConfigureShellEnvFn' 120436f58 -- internal` → `exit=1` 이다. 시접이 아직 없으므로 RED-now 이며, 빨간 이유도 맞다.
    - `git grep -nE '^var [A-Za-z]+Fn = ' 120436f58 -- internal/core/project ':!*_test.go'` → `exit=1` 로, design.md `L127` 의 "첫 사례" 서술과 맞다.
  - 남는 틈은 새 결함 D10(기본값 본문 보존 미검증)과 D11(착지 순서 절)로 따로 올렸다.
- D2 (REQ-012 범위 초과) — **RESOLVED**. `spec.md:L88` 이 REQ-012 를 §4.2 5항의 8항목으로 좁혔고, "8항목 밖의 홈 경로는 이 요구의 대상이 아니며 plan.md §K 미측정으로 남는다"는 문장을 REQ 안에 두었다. `plan.md:L226` §K 에 미추적 함수 4개가 적혀 있다.
- D3 (빈 스윕 미방어) — **RESOLVED**.
  - 규칙이 `acceptance.md:L15` 에 명문화됐다.
  - 새 테스트를 겨누는 모든 선택자에 `-v`, `--- PASS: <이름> ` 계수, `no tests to run` 0 판정이 붙었다: AC-001 `L50-51`, AC-002 `L64-65`, AC-004 `L111-115`, AC-005 `L141-142`, AC-006 `L184-185`, AC-007a `L198-199`, AC-007b 원복 `L219-220`, AC-008 `L234-235`, AC-009 `L248-250`, AC-011 `L283-284`, AC-016 `L417-421`.
  - 뮤턴트 실행은 FAIL 줄 0 이면 무효로 본다(`L224`).
  - 트리 대조: `find internal/cli/wizard internal/core/project -mindepth 1 -type d` 출력이 없다. 하위 패키지가 없으므로, `./…/...` 선택이 하위 패키지에서 `[no tests to run]` 을 찍어 판정을 엉뚱한 이유로 빨갛게 만들 위험은 없다.
- D4 (AC-008 이 REQ 보다 강함) — **RESOLVED**. REQ-IQW-015(`spec.md:L91`, Event-driven)가 섹션 파일 5개 바이트 동일을 요구하고, AC-008 이 그 REQ 에 매핑됐다(`acceptance.md:L30`). §6 이 REQ-015 를 가리킨다(`spec.md:L145, L150`).
- D5 (§6 `audit:` 서술 오류) — **RESOLVED**. `spec.md:L147` 이 "`workflow.audit` 아래 `model`·`gates` … `audit:` 키 자체는 남는다"로 고쳤다. `acceptance.md:L230` 은 "`audit:` 키 자체의 부재를 단언하지 않는다"로 고쳤다. 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml:85-91` 에 `audit: codex: model/effort, glm: model/effort` 가 실제로 있음을 재확인했다.
- D6 (REQ-003 절대형 vs 알려진 예외) — **RESOLVED**. `spec.md:L79` 에 파일시스템 루트 `project_name` 예외절과 근거가 들어갔다. 근거 `internal/cli/wizard/questions.go:50-53`(`.`·`/`·`\` → `my-project`)를 재확인했다. `acceptance.md:L454` §3 이 이 예외절과 일치한다.
- D7 (요구 계층의 방법 누출, optional) — **RESOLVED**. REQ-011 은 형식을 design.md §4 로 넘겼다(`L87`). REQ-009 는 절차를 AC-006 으로 넘겼다(`L85`). REQ-014 는 채집(014 `L90`)과 멈춤(016 `L92`)으로 나뉘었다.
- D8 (AC-015 약한 마감 판정, optional) — **RESOLVED**. 선언 슬롯 행 수 = `after.out` 파일 수 = `home-diff-exit=0` 기록 수이고, 셋 모두 1 이상이며 0 이 아닌 기록은 없어야 한다(`acceptance.md:L346, L366-373`). 슬롯 행 고정이 빠진 계수 방식은 새 선택 결함 D15 로 따로 적었다.

정체 탐지: 세 회차 연속으로 같은 모양으로 남은 결함은 없다.

## 새 표면 검토 (verification-completeness §1.1·§2)

| 표면 | §1.1 빈 스윕 | §2 두 칸 채택 (RED-now / green path) | §2 뮤턴트 탐침 |
|---|---|---|---|
| AC-IQW-016 시접 대입 테스트 | 대입 파일 총수 ≥2, `internal/cli` 쪽 ≥1, `internal/core/project` 쪽 ≥1 하한과 `sweep-exit`(`L393, L408-411, L433`). PASS 줄 계수·`no tests to run` 0(`L417-421`) | RED-now: 대입 파일 0, `git grep -c 'ConfigureShellEnvFn' 120436f58 -- internal` → `exit=1`(감사자 재측정 일치). green path: M1 이 대입 파일을 만든다. 패턴이 대입을 잡는지 대조: `userHomeDirFn` 대입 파일 8개, `update_home_seam_test.go:2`(감사자 재측정 일치) | 병렬: 대입 파일 밖에서 `t.Parallel` 을 넣는 뮤턴트 D 를 `t.Setenv` 충돌 panic 이 잡는다. panic 문자열 `testing: test using t.Setenv, t.Chdir, or cryptotest.SetGlobalRandom can not use t.Parallel` 은 go1.26.8 `src/testing/testing.go:1752` 에 있고, `Parallel` 은 `:1765-1766`, `checkParallel` 은 `:1830-1838` 이다(감사자 판독, `go.mod` `go 1.26.8`). 원복: 뮤턴트 C1·C2 를 `-count=2` 재진입 실행이 잡는다. 남는 틈은 `-count=2` 가 `t.Cleanup` 형식을 보지 않는다는 점인데, AC 가 `L445` 에 스스로 적었다. 결함: D13(텍스트 RED 기대의 가정) |
| AC-IQW-003 카드 범위 diff | 대조군 `git diff --name-only develop...HEAD \| wc -l` 이 0 이면 "측정 불가"(`L91`) | RED-now 칸은 보존 주장이라 원리상 초록이다. 대신 대조군 0 = 측정 불가를 기록했다. 감사자 재측정: `git merge-base develop HEAD` → `93182d137159c4facbf87c66dea3fd69160a6b8b`, 대조군 `0`, `mergebase-exit=0`, 옛 핀 형식 `pinned-vs-develop-exit=1`(`progress.md:L39-47` 과 일치). green path: 첫 run 커밋이 대조군을 1 이상으로 만든다. 병합 뒤에는 쓸 수 없다는 한계를 명시했다(`acceptance.md:L11`) | 흡수 뒤에도 merge-base 가 흡수한 develop 에 머물러 다른 카드의 변경이 섞이지 않는다. 형식의 RED 방향은 감사자가 알려진 입력으로 관측했다(D16). 흠잡을 뮤턴트는 찾지 못했다 |
| AC-IQW-005 존재·도달 대조군 (`--untracked`) | 존재 대조군(`git ls-files --cached --others --exclude-standard`)과 도달 대조군(`git grep --untracked -l '^package cli'`)이 목록 길이 3 이어야 한다. 모자라면 금지 토큰 검사의 `exit=1` 은 판정이 아니다(`L150`) | RED-now: 두 대조군 0, 이유 명시(`L154-157`). `--untracked` 필요성 대조: 미커밋 SPEC 디렉터리에서 `--untracked` 없음 → `tracked-only-exit=1`, 있음 → 파일 7개(감사자 재측정 일치). green path: M1·M2 가 파일을 만든다 | 목록 밖 새 실행 테스트 파일이라는 뮤턴트를 쓸 수 있다 → **D9** |

## Gaps (관측하지 않은 것)

- 컴파일·`go test`·`moai init`·`moai update`·빌드는 실행하지 않았다(리드 지시).
  - 따라서 뮤턴트 D 의 panic 이 실제로 나는지, `-count=2` 재진입이 C1·C2 를 실제로 빨갛게 하는지, 스파이 호출이 정확히 1회인지는 모두 판독이다.
  - panic 문자열과 발생 조건은 go1.26.8 소스 판독으로만 확인했다.
- AC 명령이 이 세션에서 실제로 실행되는 형태인지는 일부만 쟀다. `git diff --name-only develop...HEAD | wc -l`, `git grep --untracked`, `git diff --quiet A...B -- <path>` 형식은 이 워크트리 세션에서 가드에 거부되지 않고 돌았다. 지문 명령(`AC-015 L353`)과 `go test` 줄은 실행하지 않았다.
- `internal/core/project` 게이트 테스트가 실제 `Init` 을 도는 동안 셸 설정 외의 실제 홈 경로에 쓰는지는 추적하지 않았다. SPEC 의 §K 공백과 같다.
- `.moai/reports/t583/verdict.md` 의 F1 재현·F4 반증 측정값은 다시 재지 않았다.
- design.md 가 인용한 `internal/cli/init.go` 행번호(`:592-617`, `:832-833`, `:867`, `:999-1011`)는 이번 회차에 다시 읽지 않았다. 1회차 재측정에서 일치했고, 트리가 그 뒤 바뀌지 않았다(`internal/` 수정 없음).

## Residual-risk

- MP-3 의 lint 무결함은 트리와 갈라진 dirty 설치본(`ed71054d3-dirty`, HEAD `120436f58` 과 어느 방향으로도 조상 관계 없음)의 판정이다. 트리에만 있는 lint 규칙은 돌지 않았을 수 있다. frontmatter 12필드는 파일을 직접 읽어 따로 확인했으므로 MP-3 판정 자체는 lint 에 기대지 않는다.
- D10 을 소스 본문 비교로 고쳐도, 옮긴 옵션 구조체와 운영 `Configure` 호출 사이의 다른 변경(예: 로거 인자)은 잡지 못한다. 그 부분은 코드 리뷰가 받친다.
- AC-016 의 "`t.Cleanup` 으로" 형식 보장은 같은 파일의 다른 `t.Cleanup`(지문 재채집 등)으로도 충족된다. AC 가 스스로 인정한 잔여 위험(`acceptance.md:L445`)이며 코드 리뷰가 받친다.
- 셸 설정 줄이 이미 있는 머신에서는 슬롯 지문이 불변으로 나와도 누출 부재를 강하게 증명하지 못한다. SPEC 이 `acceptance.md:L383`, `plan.md:L229` 에서 인정했다.

## Recommendation

must-pass 7개는 모두 통과했고 점수는 0.79 에서 0.86 으로 올라 Tier L 기준을 넘었다. 1회차 결함 8건도 모두 해소됐다. 남은 FAIL 사유는 blocking 두 건이다. 3회차(마지막) 재감사는 아래 결함 차분만 대상으로 한다.

1. **D9** — `acceptance.md:L143-150` AC-IQW-005 의 대상 목록을 검증 시점 스윕으로 바꾼다: 카드가 추가한 `internal/cli/*_test.go`(커밋분 + 미커밋분) 가운데 `runInit(` 을 부르는 파일 ∪ 계획한 세 파일. 금지 토큰 검사를 그 목록 전체에 걸고, 목록 수 하한 3 을 대조군으로 둔다.
2. **D10** — REQ-IQW-011 의 "운영 기본값 = 현재 기록 동작" 절을 재는 본문 비교를 둔다. `git show 120436f58:internal/core/project/initializer.go` 의 `shell.ConfigOptions{…}` 블록과 HEAD `defaultConfigureShellEnv` 의 같은 블록을 추출해 `diff` 한다. 기준 추출물이 비어 있지 않다는 대조군과 plan 시점 줄 수를 함께 기록한다.
3. D11~D16 은 선택 사항이며 오케스트레이터 재량이다. 이 중 D14(뮤턴트 B 명령 누락)와 D16(RED 방향 기록)은 비용이 가장 적다.
