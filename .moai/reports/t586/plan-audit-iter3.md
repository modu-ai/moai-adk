# SPEC Review Report: SPEC-INIT-TUX-I18N-001

- 카드: t586 · 감사 단계: plan · Iteration: 3/3 (Tier L 상한 3, 마지막 회차)
- 감사 대상 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `02227fa1224eedc44bdeeb7050f2fbcdfff24fab` — 착수 시, 감사 도중, 보고서 작성 직후 세 번 `git rev-parse HEAD` 로 같은 값을 확인했다.
- 대상 산출물 (Tier L 입력 5종 + 기록): `.moai/specs/SPEC-INIT-TUX-I18N-001/{spec.md,plan.md,acceptance.md,design.md,research.md,progress.md}` (v0.2.2, REQ 18 · AC 22). `git status --short -- .moai/specs/SPEC-INIT-TUX-I18N-001` 출력 없음(커밋본과 워킹 사본 동일).
- 이전 감사: `plan-audit.md` (1회차, FAIL 0.67, D1~D17), `plan-audit-iter2.md` (2회차, FAIL 0.84, N1~N10). 재현 근거: `verdict.md`.
- 작성자 추론 맥락은 M1 격리 원칙에 따라 배제했다. 호출문의 리드 지시(N1~N10 점검 항목)는 판정 기준으로만 썼다.
- 판정 빌드: 이 트리에서 `go build -o <scratchpad>/moai-t586 ./cmd/moai` 로 만든 바이너리(ldflags 없음 — `version` 출력 `v3.1.3 none built unknown`), 테스트 바이너리는 `go test -c -o <scratchpad>/cli.test ./internal/cli`. 두 빌드 모두 트리 HEAD `02227fa12` 에서 만들었다. 설치본 `moai` 는 쓰지 않았다.

**Verdict: FAIL**
**Overall Score: 0.90** (네 차원 조화평균 0.896, Tier L 통과 기준 0.85)

점수는 기준을 넘는다. 그러나 FAIL 이다. 2회차 결함 **N3 가 증거 수준에서 해소되지 않았다**: AC-ITI-010 (4) 에 새로 넣은 S2 양성 절 뮤턴트(저장 경로 파일에서 `persistProjectConfig` 호출 삭제)는 설계가 유지하기로 한 S2 가드를 **실패시키지 못한다**. 같은 파일에 함수 정의 `func persistProjectConfig`(`profile_setup.go:160`)가 남아 가드의 문자열 검사가 계속 참이기 때문이다. 실제로 테스트 바이너리를 돌려 관측했다(E-4: S2 뮤턴트에서 `--- PASS`, 대신 S9 가 `--- FAIL`). 재시도 계약상 이전 회차 결함이 남으면 점수와 무관하게 FAIL 이다. 새 blocking 결함은 이 한 건(F1)뿐이고 고칠 범위는 한두 줄이다. 나머지 N1·N2·N4·N5 blocking 과 N6~N10 optional 은 모두 해소됐다.

3회차가 마지막이므로 아래 § Recommendation 끝에 사용자 개입 선택지를 적었다. 점수는 0.67 → 0.84 → 0.90 으로 올랐으므로 STOP 신호(점수 역행)는 없다.

---

## Must-Pass Results

| # | 결과 | 판정 층 / 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | PASS | 요구 층. `spec.md:172-201` 에 `REQ-ITI-001`~`REQ-ITI-018` 이 빠짐·중복 없이 3자리로 있다. 트리 빌드 lint 오류 0(E-5). |
| MP-2 GEARS 형식 | PASS | **요구 층만 채점**했다. AC 의 Given-When-Then 은 검증 층 형식이라 여기서 보지 않았다. Event-driven `:172,181,188,196`(`When …, … shall`), State-driven `:197,201`(`While …`), Unwanted behavior `:190`(`shall not contain`), 나머지 Ubiquitous. 2회차 N7 의 `Once …` 절은 사라졌고 `REQ-ITI-017` 라벨은 여는 절 `While` 과 맞는다. 정식 패턴 선택에 대한 의견은 F4(optional). |
| MP-3 YAML frontmatter | PASS | `spec.md:2-14` 에 12개 필드가 모두 있다: `id`, `title`, `version: "0.2.2"`(따옴표 semver), `status: draft`, `created`/`updated: 2026-09-11`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` 쉼표 문자열. 거부 별칭 없음. `tier: L`, `related_specs` 는 선택 필드. |
| MP-4 언어 중립성 | N/A | 대상은 `internal/cli`·`internal/cli/wizard` Go 코드이고 `internal/template/templates/**` 는 바뀌지 않는다(`spec.md:164`). |
| MP-5 D7 교차 SPEC | PASS | 참조 SPEC 11개(자기 자신 포함) 가운데 `.moai/specs/` 에 있는 9개 관련 SPEC 이 모두 `status: completed`. retired·superseded·archived 참조 없음. `SPEC-INIT-QUIET-WIZARD-001` 부재(D7-5 SHOULD)는 SPEC 이 스스로 기록했다(`spec.md:145`, `research.md:284`, `plan.md:38`) — E-6. |
| MP-6 D8 크로스 플랫폼 | PASS | `grep -c syscall spec.md` → `0`. D8-4 자동 PASS. |
| MP-7 명확화 게이트 | PASS | `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, 종료 1. `plan.md:138` "남은 확인 필요 표식은 없다". |

---

## Category Scores (0.0-1.0, rubric-anchored)

| 차원 | 점수 | 기준 구간 | 근거 |
|---|---|---|---|
| Clarity | 0.90 | 0.75~1.0 (1.0 쪽) | 규범 텍스트가 해석 하나로 닫힌다. 2회차의 표현 흐림(N6 catppuccin, N7 `Once`, N8 tmux 부재)은 모두 고쳐졌다(`spec.md:206`, `:197`, `plan.md:37`). 남은 흐림: `REQ-ITI-017` 의 조건이 런타임 상태가 아니라 트리 내용이라 `While` 보다 `Where` 가 정식에 가깝다(F4), `acceptance.md:49` 감시 목록 머리가 "update 흐름"까지 덮는다고 적지만 update 경로의 홈 쓰기 하나를 빠뜨렸다(F3). |
| Completeness | 0.95 | 0.75~1.0 (1.0 쪽) | 필수 절(HISTORY `spec.md:22`, §A 배경, §B 요구, `### Out of Scope — …` 6개 `spec.md:216-242`), frontmatter 12필드, Tier L 산출물 5종과 `progress.md` §E.1~§E.4 가 있다. `research.md` §14~§16 이 N1·N4·N5 의 도출 명령과 출력을 담았다. 감점: `research.md:9` 의 재현 명령 기록이 문서 자신 때문에 지금은 재현되지 않는다(F2). |
| Testability | 0.80 | 0.75~1.0 (0.75 쪽) | N1(감시 목록·양성 대조군), N2(스테퍼 조건 제거), N4(대조군 3곳), N5(실효 환경 관측)로 MUST AC 두 개(003·020)가 이진 판정이 됐다. 그러나 MUST AC-ITI-010 (4) 의 S2 절은 설계대로 구현하면 **반드시 FAIL** 하는 관측 조건이다(F1, E-4 실측). |
| Traceability | 0.95 | 0.75~1.0 (1.0 쪽) | 모든 REQ 에 AC 가 있고 고아 AC 가 없다(`acceptance.md:188`). 트리 빌드 lint 커버리지 경고 0 이고, 미매핑 REQ 뮤턴트로 수집기가 이 형식을 실제로 읽음을 다시 확인했다(E-5). 감점: `REQ-ITI-010` "every positive guard shall fail when its asserted construct is removed" 가 S2 에 대해서는 AC 로 실제로 확인되지 않는다(F1). |

조화평균 = 4 / (1/0.90 + 1/0.95 + 1/0.80 + 1/0.95) = 4 / 4.4664 = **0.896**.

---

## N1~N10 처분 (2회차 결함)

| 결함 | 분류(2회차) | 처분 | 증거 |
|---|---|---|---|
| N1 실제 HOME 매니페스트 비이진 | major·blocking | **해소** | 트리 매니페스트가 코드 도출 감시 목록 W1~W6 으로 바뀌었다(`acceptance.md:32` P8, `:49-58`). 런타임 하위 트리 제외를 명시했다(`:60`). AC-ITI-020 (4) 에 가짜 HOME 양성 대조군 다섯 경우가 있고, (ii) 가 감시 파일 한 바이트 변경으로 FAIL 을 요구한다(`:155`). `research.md` §14 도출 grep 을 이 트리에서 다시 돌려 11줄·종료 0 이 기록과 같다(E-1). 추가 탐색에서 pty 사례가 부르는 경로에 목록 밖 실제 HOME 쓰기는 찾지 못했다(E-2). update 경로의 체크포인트 쓰기 하나는 목록에 없으나 pty 사례가 update 를 부르지 않아 판정에 영향이 없다(F3, optional). |
| N2 AC-003 스테퍼 전제 | major·blocking | **해소** | AC-ITI-003 이 기준 문자열과 옵션 줄 4개로 도달성을 단정하고 단계 표시 줄을 판정하지 않는다고 명시한다(`acceptance.md:84`, `:68`). 문서 전체에서 `1 / 4` 문자열은 AC-003 의 위임 문장(`:84`) 한 곳뿐이고 판정 조건이 아니다. 분모는 AC-ITI-021 이 `<k> / 4` 로 판정한다(`:145`). |
| N3 양성 가드 뮤턴트 커버 | minor·blocking | **미해소(부분)** | AC-ITI-010 (4) 에 S2 양성 절·S3·S4·S6 뮤턴트가 더해졌고 REQ-ITI-010 은 좁혀지지 않았다(`acceptance.md:122`, `spec.md:184`). S3·S4·S9 의 기준 문자열은 `profile_setup.go` 에 한 번씩만 있어 뮤턴트가 가드를 죽인다(E-3, S3 실측 FAIL). **S2 는 아니다**: 가드는 주석 아닌 줄에 `persistProjectConfig` 가 있는지만 보는데(`profile_setup_nested_test.go:39-41`), 호출(`:520`)을 지워도 정의(`:160`)가 남는다. 실측으로 S2 뮤턴트에서 S2 는 `--- PASS`, S9 가 `--- FAIL` 이었다(E-4). → F1 |
| N4 부재 단정 대조군 | minor·blocking | **해소** | AC-ITI-004 (2) 대조군 `git grep -l '"charm.land/huh/v2"' -- '*_test.go'`(`acceptance.md:90`), (3) `grep -c 'charm.land/huh/v2 ' go.mod` → `1`(`:91`), AC-ITI-011 (1) 한 호출 안의 대조군 경로 `internal/cli/wizard/wizard.go`(`:124`). 측정 원문은 `research.md` §16. |
| N5 자식 환경 격리 | minor·blocking | **해소** | P4 가 변수마다 `-e VAR=value` 로 넘기고 부모 쪽 정리로 추론하지 않는다고 정한다(`acceptance.md:28`). 정리 목록 `:35-45`(HOME·MOAI_HOME 임시, CLAUDE_CONFIG_DIR 빈 값, `MOAI_KANBAN` 접두 9개 빈 값). 실효 환경 관측 네 단정(`:47`)을 AC-ITI-003(`:84`)과 AC-ITI-020 (1)(`:156`)이 요구한다. `spec.md:209` 테스트 격리, `design.md:176,180`, AC-ITI-012 인프로세스 `t.Setenv` 와 읽기 확인(`acceptance.md:135`). `internal/config/envkeys.go` 의 `"MOAI_KANBAN` 상수는 9개이고 목록과 이름이 모두 같다. envkeys 밖에 `"MOAI_KANBAN…"` 리터럴은 없다(E-6). |
| N6 catppuccin 서술 | minor·optional | **해소** | `spec.md:206` 이 huh v2 도 요구해 남는다고 고쳤다. `go mod graph` 에 `charm.land/huh/v2@v2.0.3 github.com/catppuccin/go@v0.2.0`, 모듈 캐시 `theme.go:6` 이 import 한다(E-7). |
| N7 REQ-017 `Once` 절 | minor·optional | **해소(의견 있음)** | `spec.md:197` 이 `(State-driven): **While** …` 로 라벨과 여는 절이 맞는다. 조건이 트리 내용(정적 구성)이라 GEARS 정식 선택은 `Where` 쪽이 더 맞다 — F4(optional). MP-2 판정에는 영향 없음. |
| N8 tmux 부재 표현 | minor·optional | **해소** | `plan.md:37` "pty 테스트는 FAIL … SKIP 이나 Gap 으로 바꿔 적지 않으며", P2 `acceptance.md:26` `t.Fatal`(FAIL), AC-ITI-019 (b) `:150`. 세 곳이 FAIL 로 통일됐다. |
| N9 t583 SPEC 부재 | minor·optional | **해소** | `spec.md:145`, `research.md:284`(대조군 포함), `plan.md:38` 게이트 재확인. 이 트리에서도 부재 확인(E-6). |
| N10 세션 목록 전역 비교 | minor·optional | **해소** | P7 이 이름 생성 함수 하나와 접두 한정 비교를 정한다(`acceptance.md:31`). AC-ITI-019 (a) 가 `moai-ptycap-` 접두 집합만 비교한다(`:150`). `design.md:176` "세션 이름은 이름 생성 함수 하나만 만든다". |

집계: 해소 9 (blocking 4 · optional 5), 미해소 1 (N3, blocking).

---

## Regression Check

| 대상 | 결과 | 증거 |
|---|---|---|
| D1~D17 (1회차) 전반 | 회귀 없음 | REQ 18·AC 22 번호 연속, 추적 표 `acceptance.md:188`, lint 0/0(E-5). |
| AC-018/021/022 단일 성질 뮤턴트 | 유지 | `acceptance.md:144`(그룹 재분할 뮤턴트, 021 PASS 동시 관측), `:145`(분모 뮤턴트, 018 PASS 동시 관측, 골든은 증거로 세지 않음), `:146`(`.Title(pending[0].Group)` 뮤턴트). |
| V-a~V-d 선결 표시 (Q5 조건) | 유지 | `plan.md:66-70`, `acceptance.md` §D.1 `:160-182`. |
| SPEC-CLI-TUI-MODERNIZE-001 계약 인수 | 유지 | `spec.md:126-139`(뒤집는 절 4개, 대체 보장, 인수 문장). |
| t583 게이트 위치 | 유지 | `spec.md:160` [HARD], `plan.md:44`, `:90-92`, M4~M8 게이트 뒤. |
| 리드 판정 Q1~Q5 | 유지 | `plan.md:130-136`. |
| 부재 단정마다 대조군 | 유지 | AC-001 `:82`(두 쌍), AC-004 `:89-91`, AC-011 `:124-125`, AC-013 `:136`, AC-022 `:146`. |
| ERE 와 `\b` | 판정 명령에 없음 | `grep -rn -E '\\b'` 적중 2줄은 `research.md:143`(옛 조사 문단)과 `:147`(그 형태가 공허하다는 정정 기록)뿐이다. AC·plan 명령에는 없다. 옛 문단은 F5(optional). |
| 흡수로 바뀐 파일의 인용 좌표 | 유효 | `18144b7ac..02227fa12` 제품 코드 차이 56개 파일, `538b56f19..02227fa12` 는 0개. SPEC 이 인용하는 CLI 파일 중 바뀐 것은 `update.go`(−35/+17, 483행 이후 삭제), `update_tux.go`, `update_wizard.go`. SPEC 의 `update.go` 좌표 12개(`:172-192`, `:174`, `:178`, `:179`, `:185`, `:187`, `:205`, `:790`, `:923`, `:931-958`, `:940`, `:940-943`)와 `init.go`·`update_version.go`·`launcher.go:1094-1110` 좌표를 HEAD 에서 읽어 모두 일치했다(E-8). 기준 문자열이 같은 `toolpolicy/types.go`·`i18n.js` 는 파일명 충돌일 뿐 SPEC 인용 대상이 아니다. |

---

## Defects Found (structured defect-list)

F1. N3-S2-MUTANT-SURVIVES — acceptance.md:L120, L122; design.md:L161 (§10 S2); spec.md:L184 — AC-ITI-010 (4) 는 "저장 경로 파일에서 `persistProjectConfig` 호출을 지우면 S2(양성 절) … 실패", "뮤턴트마다 실패 출력에 해당 가드의 테스트 이름이 있음을 확인" 을 요구한다. 그런데 `design.md` §10 은 S2 를 "스캔 유지 … 기준 문자열 `persistProjectConfig` 는 `profile_setup.go` 에서" 로 유지하고, AC-ITI-010 (2) 는 기준 문자열을 "저장 함수 이름" 으로 정한다. 현재 S2 는 주석 아닌 줄에 `persistProjectConfig` 문자열이 있는지만 보는데(`profile_setup_nested_test.go:39-41`), 이 파일에는 함수 정의 `func persistProjectConfig(`(`profile_setup.go:160`)가 있어 호출(`:520`)을 지워도 문자열이 남는다. 테스트 바이너리로 실측한 결과 호출만 지운 사본에서 `TestTUINestedConfigNoParallelWriter` 는 `--- PASS`, 같은 뮤턴트를 잡은 것은 S9 `TestWizardCarriesStoredSegmentsIntoPrefs` 였다(E-4). 같은 방법이 S3 뮤턴트는 죽여(`--- FAIL`) 방법 자체는 유효하다. 설계대로 구현하면 MUST AC 한 절이 반드시 FAIL 하고, REQ-ITI-010 "every positive guard shall fail when its asserted construct is removed" 가 S2 에 대해 성립하지 않는다. 2회차 N3 의 요구("뮤턴트마다 그 가드가 실패")를 S2 에서 채우지 못했다. — Severity: major — Class: blocking — Required fix: 둘 중 하나. (a) `design.md` §10 S2 의 기준 문자열을 호출 형태로 바꾼다(예: `persistProjectConfig(cwd,` 또는 정의 줄을 뺀 호출 수 ≥1 단정). 그에 맞춰 AC-ITI-010 (2) 의 "저장 함수 이름" 을 "저장 함수 호출" 로 고친다. (b) S2 양성 절을 저장 행동 테스트로 대체한다고 적고, AC-ITI-010 (4) 의 S2 뮤턴트 기대 실패 가드를 그 대체 테스트 이름으로 바꾼다. 어느 쪽이든 plan.md M5 "소스 스캔 가드 9건을 design.md §10 대로 재조준" 문장은 그대로 둘 수 있다.

F2. RESEARCH-SELF-INVALIDATING-RECORD — research.md:L9 — "`update_wizard.go`·`update_tux.go` 도 바뀌었으나 SPEC 문서에 두 파일 이름이 없다(`grep -n -E 'update_wizard|update_tux' .moai/specs/SPEC-INIT-TUX-I18N-001/*.md` 종료 1)" 라고 적었다. 하지만 이 문장 자체가 두 이름을 담아, 기록된 명령을 지금 돌리면 `research.md:9` 한 줄이 나오고 종료 0 이다(E-8). 판독 결론(인용 좌표가 두 파일에 없음)은 여전히 참이다. — Severity: minor — Class: optional — Required fix: 명령에 `':!research.md'` 류 제외를 붙이거나, "측정 당시 종료 1, 이 문장 추가 뒤에는 이 줄 자체가 적중" 이라고 적는다.

F3. WATCHLIST-UPDATE-FLOW-CLAIM — acceptance.md:L49, L60; research.md:L288-330 — 감시 목록 머리는 "이 SPEC 의 pty 사례가 부르는 흐름(init, update, 프로필 위저드, 다운그레이드 확인창)이 실제 HOME 에서 쓸 수 있는 파일" 이라고 적는다. 그런데 `moai update` 의 잔여 정리 경로(`update_residue_cleanup.go:85` → `update.go:858` `runAgencyMigrationAdapter`)는 `migrate_agency.go:188-194` `checkpointPath` 로 `H/.moai/.migrate-tx-<id>.json`(또는 `MOAI_HOME` 아래)을 쓴다. 이 경로는 W1~W6 에도, 제외 목록에도 없다. §14 도출 grep 패턴에 `paths.Home()`·`runAgencyMigrationAdapter` 가 없어서 빠졌다. 어떤 pty 사례도 `moai update` 흐름을 실행하지 않고(AC-015 (a) 는 확인창 헬퍼만), 자식 HOME·MOAI_HOME 이 임시 경로라 판정에는 영향이 없다. — Severity: minor — Class: optional — Required fix: 머리 문장을 실제 pty 사례가 부르는 경로(init 첫 화면·프로필 위저드·확인창 헬퍼)로 좁히거나, W 목록에 `M/.migrate-tx-*.json` 글롭을 더한다.

F4. REQ017-WHILE-VS-WHERE — spec.md:L197 — `REQ-ITI-017` 의 조건 "While the tree carries the card t583 init question set" 은 실행 중에 바뀌는 시스템 상태가 아니라 소스 트리의 정적 내용이다. GEARS 는 `Where` 를 정적 구성·기능 게이트로 정의하므로 정식 선택은 `Where` 에 가깝다. `While` 도 다섯 패턴 중 하나이고 라벨과 여는 절이 맞아 MP-2 는 PASS 다. — Severity: minor — Class: optional — Required fix: `(Where): **Where** the init question set is the card t583 set, …` 로 바꾸거나 그대로 둔다.

F5. RESEARCH-STALE-BACKSLASH-B-PARAGRAPH — research.md:L143 — 옛 문단이 `grep -rn '\.Group\b'` 결과로 "q.Group 을 읽는 곳은 :183, :186 뿐" 이라고 단정한다. 바로 아래 §6.1(`:147`)이 그 형태가 공허할 수 있음을 밝히고 경계 없는 패턴과 대조군으로 같은 결론을 다시 쟀다. 판정 명령이 아니라 판독 결과에는 영향이 없다. — Severity: minor — Class: optional — Required fix: `:143` 문단 앞에 "§6.1 로 대체됨" 표식을 단다.

---

## Recommendation

**FAIL (3회차, 마지막).** blocking 결함은 F1 한 건이고 2회차 N3 의 남은 부분이다. 고칠 곳은 manager-spec 소관 문서 두 곳이다.

1. `design.md:161` (§10 표 S2 행): "기준 문자열 `persistProjectConfig`" 를 호출 형태(`persistProjectConfig(cwd,` 등 정의 줄에 걸리지 않는 문자열)로 바꾸거나, S2 양성 절을 저장 행동 테스트로 대체한다고 적는다.
2. `acceptance.md:120` AC-ITI-010 (2) 의 기준 문자열 설명("저장 함수 이름")과 `:122` (4) 의 S2 뮤턴트 기대를 1 의 선택에 맞춘다. 수정 뒤에는 E-4 방식(가드가 읽는 파일만 바꾼 사본에서 테스트 바이너리 실행)으로 S2 뮤턴트가 S2 이름으로 실패함을 한 번 관측해 두면 run 단계의 AC-010 (4) 판정이 막히지 않는다.
3. optional(F2~F5)은 오케스트레이터 재량이다.

재시도 상한(Tier L 3회)에 도달했으므로 오케스트레이터는 사용자에게 다음 셋을 올린다(순서는 권고가 아니다):

- **PASS-with-debt** — 현재 상태를 받아들이고 F1 을 run 단계 첫 작업(M5 가드 재조준 전)의 선결 수정으로 기록한 뒤 진행한다. 남는 위험: 기록이 누락되면 run 단계 AC-ITI-010 (4) 가 FAIL 로 막힌다.
- **범위 축소 후 재진입** — F1 두 줄을 고친 뒤 결함 차분만 재감사한다.
- **사용자 명시 연장** — 4회차 감사를 허용한다.

---

## Evidence

### E-1 감시 목록 도출 재실행 (`research.md` §14)

```
$ git grep -n -E 'profile\.WritePreferences\(|homestate\.EnsureProjectLayout\(|ApplyAutonomyTierBundle\(|ensureGlobalSettingsEnv\(|globalMoaiHooksDir\(|runShellEnvConfig\(|configureShellEnv\(|RecordLastUsedProfile\(|paths\.(UserSettingsFile|UserConfigSectionsDir|ProfilesDir|StateDir|CacheDir|ReleasesDir|WorktreesDir|GlmEnvFile)\(' -- internal/cli/init.go internal/cli/update.go internal/cli/update_version.go internal/cli/profile_setup.go internal/cli/profile.go internal/core/project/initializer.go ':!*_test.go'
internal/cli/init.go:877:	if err := homestate.EnsureProjectLayout(opts.ProjectRoot); err != nil {
internal/cli/init.go:890:		if tierErr := project.ApplyAutonomyTierBundle(
internal/cli/init.go:964:	if err := ensureGlobalSettingsEnv(); err != nil {
internal/cli/profile_setup.go:489:	if err := profile.WritePreferences(profileName, prefs); err != nil {
internal/cli/update.go:205:		return runShellEnvConfig(cmd)
internal/cli/update.go:790:func runShellEnvConfig(cmd *cobra.Command) error {
internal/cli/update.go:923:func globalMoaiHooksDir(homeDir string) string {
internal/cli/update.go:931:func ensureGlobalSettingsEnv() error {
internal/cli/update.go:940:	globalHooksDir := globalMoaiHooksDir(homeDir)
internal/core/project/initializer.go:334:		if shellResult, err := i.configureShellEnv(); err != nil {
internal/core/project/initializer.go:677:func (i *projectInitializer) configureShellEnv() (*shell.ConfigResult, error) {
derive_exit=0
```

기록(`research.md:293-305`)과 11줄이 모두 같다.

### E-2 감시 목록 밖 홈 쓰기 탐색

```
$ git grep -n -E 'os\.UserHomeDir\(|paths\.Home\(|userHomeDirFn\(|paths\.MoaiHome\(|homestate\.|GetBaseDir\(|os\.WriteFile\(|os\.MkdirAll\(|os\.Rename\(|os\.RemoveAll\(|os\.Remove\(|os\.Create\(|os\.OpenFile\(' -- internal/cli/init.go internal/cli/update.go internal/cli/update_version.go internal/cli/profile_setup.go internal/cli/profile.go internal/cli/init_update_notice.go internal/cli/update_wizard.go internal/cli/update_tux.go ':!*_test.go'
internal/cli/init.go:578:		if err := os.MkdirAll(rootFlag, 0755); err != nil {
internal/cli/init.go:877:	if err := homestate.EnsureProjectLayout(opts.ProjectRoot); err != nil {
internal/cli/init.go:889:	if homeDir, homeErr := userHomeDirFn(); homeErr == nil {
internal/cli/profile.go:69:// global profile dir (profile.GetBaseDir()), and states EXPLICITLY that the
internal/cli/profile.go:80:		wtPath, profile.GetBaseDir())
internal/cli/update.go:859:	homeDir, err := paths.Home()
internal/cli/update.go:932:	homeDir, err := userHomeDirFn()
internal/cli/update.go:942:		_ = os.RemoveAll(globalHooksDir)
exit=0
$ git grep -n -E 'user\.Current\(|os\.UserConfigDir\(|os\.UserCacheDir\(|XDG_(CONFIG|CACHE|STATE|DATA)_HOME' -- 'internal/*.go' 'pkg/*.go' ':!*_test.go'
(출력 없음)
$ grep -n 'homeDir' internal/cli/migrate_agency*.go | grep -v _test
internal/cli/migrate_agency.go:143:	homeDir     string
internal/cli/migrate_agency.go:186:// follows the override (REQ-MHP-001); otherwise the injected homeDir seam
internal/cli/migrate_agency.go:193:	return filepath.Join(r.homeDir, defs.MoAIDir, name)
internal/cli/migrate_agency.go:720:	homeDir, err := paths.Home()
internal/cli/migrate_agency.go:727:		homeDir:     homeDir,
$ grep -n 'runAgencyMigrationAdapter(' internal/cli/*.go | grep -v _test
internal/cli/update.go:858:func runAgencyMigrationAdapter(projectRoot string, dryRun, force bool, out io.Writer) error {
internal/cli/update_residue_cleanup.go:85:		if migrateErr := runAgencyMigrationAdapter(projectRoot, dryRun, force, out); migrateErr != nil {
```

판독: `update.go:859` 는 에이전시 이관 체크포인트 경로(F3). `initializer.go:394` `os.UserHomeDir()` 는 `detectGoBinPath` 읽기용이다. `WritePreferences`(`preferences.go:166`)는 기본·이름 프로필 `preferences.yaml` 하나만 쓴다(`MkdirAll` + `WriteFile`). `selectConfigFile`(`shell/detect.go:128-182`)은 W5 의 여섯 파일과 PowerShell 경로만 돌려준다. `loadSessionWorktreeConfig` 는 `os.Getwd()`(`session_worktree.go:717`)로 루트를 푼다. HOME 을 무시하고 홈을 푸는 호출(`user.Current` 등)은 제품 코드에 없으므로, 자식 HOME·MOAI_HOME 을 임시로 둔 상태에서 실제 HOME 에 닿는 길은 정리 실패뿐이다. 정리 실패는 실효 환경 관측이 잡는다.

`CLAUDE_PROJECT_DIR` 를 읽는 비테스트 코드는 hook·session·gate·navigator 등에 있고 init/update/profile_setup/session_worktree 에는 없다(`git grep -n -E 'CLAUDE_PROJECT_DIR|EnvClaudeProjectDir' -- 'internal/*.go' ':!*_test.go' ':!internal/hook/*'`).

### E-3 S2·S3·S4·S9 기준 문자열의 등장 위치

```
$ grep -rn 'persistProjectConfig' internal/cli --include='*.go' | grep -v '_test\.go:'
internal/cli/profile_setup.go:152:// persistProjectConfig writes the selected development_mode + git_convention
internal/cli/profile_setup.go:160:func persistProjectConfig(projectRoot, devMode, convention string) error {
internal/cli/profile_setup.go:227:// (persistProjectConfig writes .moai/config/sections/*.yaml). Read-only
internal/cli/profile_setup.go:243:	// persistProjectConfig). Mirrors the M2 init / M3 web wiring.
internal/cli/profile_setup.go:518:			// persistProjectConfig writes only non-empty values (EC-1), so the stored
internal/cli/profile_setup.go:520:			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
$ sed -n 27,78p internal/cli/profile_setup_nested_test.go | grep -n -E 'Contains|persist'
13:	if !strings.Contains(codeLines, "persistProjectConfig") {
$ grep -n -E 'permissionMode == defaultPermissionMode|EmptyLabelFor\("model_policy"\)' internal/cli/*.go | grep -v _test
internal/cli/profile_setup.go:405:					huh.NewOption(settings.EmptyLabelFor("model_policy"), ""),
internal/cli/profile_setup.go:456:	if permissionMode == defaultPermissionMode {
$ grep -n -E 'StatuslineSegments|persistProjectConfig\(cwd' internal/cli/profile_setup.go
463:	// StatuslineSegments carries the profile's STORED map through untouched. The
479:		StatuslineSegments: existingPrefs.StatuslineSegments,
508:			syncPrefs.StatuslineSegments = nil
520:			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
```

`nonCommentLines`(`profile_setup_nested_test.go:46-56`)는 `//` 로 시작하는 줄만 뺀다. `:160` 정의 줄은 남는다.

### E-4 S2 뮤턴트 실측 (F1)

가드는 실행 시점 작업 디렉터리의 `profile_setup.go` 를 읽는다. 제품 트리는 건드리지 않고 스크래치 디렉터리 셋에 사본을 두었다.

```
$ go test -c -o <scratch>/cli.test ./internal/cli
testbuild_exit=0
$ diff internal/cli/profile_setup.go <scratch>/g-s2mut/profile_setup.go
520c520
< 			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
---
> 			if err := error(nil); err != nil {
$ diff internal/cli/profile_setup.go <scratch>/g-s3mut/profile_setup.go
456c456
< 	if permissionMode == defaultPermissionMode {
---
> 	if permissionMode == "acceptEdits" {
```

각 디렉터리에서 `<scratch>/cli.test -test.run <이름> -test.v -test.count=1 > <파일> 2>&1` 로 실행:

```
g-ctl/s2.txt:=== RUN   TestTUINestedConfigNoParallelWriter
g-ctl/s2.txt:--- PASS: TestTUINestedConfigNoParallelWriter (0.00s)             exit=0
g-s2mut/s2.txt:=== RUN   TestTUINestedConfigNoParallelWriter
g-s2mut/s2.txt:--- PASS: TestTUINestedConfigNoParallelWriter (0.00s)           exit=0   ← 뮤턴트 생존
g-s3mut/s2.txt:=== RUN   TestTUINestedConfigNoParallelWriter
g-s3mut/s2.txt:--- PASS: TestTUINestedConfigNoParallelWriter (0.00s)           exit=0
g-ctl/s3.txt:=== RUN   TestPermissionModeNormalizeAcceptEdits
g-ctl/s3.txt:--- PASS: TestPermissionModeNormalizeAcceptEdits (0.00s)          exit=0
g-s3mut/s3.txt:=== RUN   TestPermissionModeNormalizeAcceptEdits
g-s3mut/s3.txt:    profile_setup_nested_test.go:86: profile_setup.go must normalize acceptEdits permission mode to empty (REQ-WC10-014)
g-s3mut/s3.txt:--- FAIL: TestPermissionModeNormalizeAcceptEdits (0.00s)          exit=1   ← 방법 양성 대조
g-ctl/s9.txt:=== RUN   TestWizardCarriesStoredSegmentsIntoPrefs
g-ctl/s9.txt:--- PASS: TestWizardCarriesStoredSegmentsIntoPrefs (0.00s)         exit=0
g-s2mut/s9.txt:=== RUN   TestWizardCarriesStoredSegmentsIntoPrefs
g-s2mut/s9.txt:    profile_setup_removed_questions_test.go:131: persistProjectConfig must be called with an empty convention so git-convention.yaml is preserved
g-s2mut/s9.txt:--- FAIL: TestWizardCarriesStoredSegmentsIntoPrefs (0.00s)         exit=1   ← 잡은 것은 S9
```

선택된 테스트 수는 파일마다 `=== RUN` 1줄로 0 이 아니다.

### E-5 SPEC lint (판정 빌드 = 트리 빌드)와 수집기 비공허성

```
$ <scratch>/moai-t586 spec lint .moai/specs/SPEC-INIT-TUX-I18N-001      (cwd = 워크트리 루트)
INFO      OwnershipTransitionUnmeasured  .moai/specs/SPEC-INIT-TUX-I18N-001/spec.md  1  …
0 error(s), 0 warning(s)
lint_exit=0
```

대조군과 뮤턴트(스크래치 트리에 SPEC 복사, git 없음):

```
ctl: 0 error(s), 0 warning(s)                                        ctl_exit=0
mut: (acceptance.md 에서 AC-ITI-017 의 `maps REQ-ITI-016` → `maps REQ-ITI-015`, §D.2 의 `REQ-ITI-016→AC-ITI-017 · ` 제거; grep -c 'REQ-ITI-016' → 0)
WARNING   CoverageIncomplete  .moai/specs/SPEC-INIT-TUX-I18N-001/spec.md  181  REQ REQ-ITI-016 is not referenced by any AC
0 error(s), 1 warning(s)                                             mut_exit=0
```

`mcp__moai__spec_audit`(project_root = 워크트리, filter SPEC-INIT-TUX-I18N-001): `modern_era_clean: 1`, finding 은 `EraAutoDetected` INFO 하나.

### E-6 D7 상태·환경 키·t583 부재

```
SPEC-CLI-TUI-MODERNIZE-001 status: completed
SPEC-CLI-TUX-V3-002 status: completed
SPEC-CLI-TUX-INIT-UPDATE-001 status: completed
SPEC-CLI-WIZARD-RESTRUCTURE-001 status: completed
SPEC-INIT-WIZARD-REPAIR-001 status: completed
SPEC-INIT-HARNESS-PROMPT-001 status: completed
SPEC-WEB-CONSOLE-002 status: completed
SPEC-WEB-CONSOLE-003 status: completed
SPEC-I18N-GOVERNANCE-001 status: completed
SPEC-INIT-QUIET-WIZARD-001 MISSING
$ grep -c -E '"MOAI_KANBAN' internal/config/envkeys.go
9
$ grep -n -E '"MOAI_KANBAN' internal/config/envkeys.go
182 MOAI_KANBAN · 187 MOAI_KANBAN_SPEC · 195 MOAI_KANBAN_ID · 205 MOAI_KANBAN_LABEL · 216 MOAI_KANBAN_SETTINGS_INJECTED · 222 MOAI_KANBAN_LEAD_ADDR · 235 MOAI_KANBAN_BACKEND · 246 MOAI_KANBAN_CARD · 263 MOAI_KANBAN_LEAD_NAME
$ git grep -n -E '"MOAI_KANBAN[A-Z_]*"' -- '*.go' ':!*_test.go' ':!internal/config/envkeys.go'
other_literal_exit=1
$ grep -rn -E 'Getenv\((config\.)?EnvMoaiKanban[A-Za-z]*\)' internal --include='*.go' | grep -v -c '_test\.go:'
20
```

(목록 9개 이름이 `acceptance.md:42` 와 같다. `Getenv` 줄 수 20 은 `research.md:353` 기록과 같다.)

### E-7 모듈 그래프

```
$ go mod graph | grep -E 'github.com/catppuccin/go'
github.com/modu-ai/moai-adk github.com/catppuccin/go@v0.3.0
charm.land/huh/v2@v2.0.3 github.com/catppuccin/go@v0.2.0
github.com/charmbracelet/huh@v1.0.0 github.com/catppuccin/go@v0.3.0
graph_exit=0
$ grep -n catppuccin "$(go env GOMODCACHE)/charm.land/huh/v2@v2.0.3/theme.go"
6:	catppuccin "github.com/catppuccin/go"
```

### E-8 흡수 뒤 인용 좌표 대조

```
$ git diff --name-only 18144b7aca714ea8924363b1eab4640cf101c6d0 02227fa1224eedc44bdeeb7050f2fbcdfff24fab -- internal cmd pkg go.mod go.sum | wc -l
56
$ git diff --name-only 538b56f1923c7b72e8dcb8379d55d05e4fadb1c5 02227fa1224eedc44bdeeb7050f2fbcdfff24fab -- internal cmd pkg go.mod go.sum | wc -l
0
$ git diff --stat 18144b7ac 02227fa12 -- internal/cli/update.go internal/cli/update_wizard.go internal/cli/update_tux.go internal/cli/init.go internal/cli/update_version.go internal/cli/profile_setup.go internal/cli/wizard
 internal/cli/update.go        | 33 +++------------------------------
 internal/cli/update_tux.go    |  2 +-
 internal/cli/update_wizard.go | 17 +++++++++++++----
```

`update.go` 변경 훙크는 `@@ -483,16 +483,9 @@`(사전 스냅숏 삭제)와 `@@ -555,26 +548,6 @@`(레거시 스킬 보관 블록 삭제)다. `update_tux.go` 는 복구 명령 문구, `update_wizard.go` 는 `system.yaml` 읽기·쓰기 오류 반환이다. SPEC 이 인용하는 확인창·흐름 구간과 겹치지 않는다.

HEAD 에서 읽은 좌표: `update.go:174` `if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) {`, `:178` `confirm := huh.NewConfirm().`, `:179` 제목 `"No profile found. …"`, `:185` `…WithTheme(moaiHuhTheme())`, `:187` `runProfileSetup(cmd, nil)`, `:205` `return runShellEnvConfig(cmd)`, `:790` `func runShellEnvConfig`, `:923` `func globalMoaiHooksDir`, `:931` `func ensureGlobalSettingsEnv`, `:940` `globalHooksDir := …`, `:942` `os.RemoveAll`. `init.go:648` 조건, `:651` 제목, `:657` 테마, `:659` 호출, `:704` `runWizardFn`, `:707` `"Initialization cancelled."`, `:877` `EnsureProjectLayout`, `:889-897` `ApplyAutonomyTierBundle`, `:964` `ensureGlobalSettingsEnv`. `update_version.go:325` 조건, `:327` `huh.NewConfirm()`, `:331` `WithTheme(moaiHuhTheme())`. `launcher.go:1094-1110` 은 `runProfileSetup` 정규화 블록을 가리키는 주석. `wizard/questions.go` 함수 시작 줄(48·162·268·296·307·318·329·349), `wizard.go`(35·158·236·242·397·467·487), `translations.go:540` `var uiStrings` 모두 `spec.md:143`·`research.md` §6 과 같다. v1 importer 5개(`huh_theme.go`, `init.go`, `profile_setup.go`, `update.go`, `update_version.go`)와 v2 대조군 `wizard/wizard.go` 도 §D.3 L1 과 같다.

```
$ grep -n -E 'update_wizard|update_tux' spec.md plan.md acceptance.md design.md research.md progress.md
.moai/specs/SPEC-INIT-TUX-I18N-001/research.md:9:- 3회차 개정 측정 트리: HEAD `538b56f19…` …
exit=0
```

(F2 근거.)

---

## Baseline-attribution

- 트리: 워크트리 `.claude/worktrees/t586`, HEAD `02227fa1224eedc44bdeeb7050f2fbcdfff24fab`, SPEC 디렉터리 워킹 사본 = 커밋본. 모든 명령은 이 감사 실행에서 이 트리에 대해 돌렸다.
- 판정 빌드: `moai-t586`, `cli.test` 모두 이 트리 HEAD 에서 스크래치 경로로 빌드(`build_exit=0`, `testbuild_exit=0`). 설치본 `moai` 와 MCP 서버 빌드(v3.2.0-rc.5 `84fa4ece4`)는 lint 판정에 쓰지 않았다. MCP `spec_audit` 결과(E-5 끝)만 그 서버 빌드에서 왔고, 판정에는 보조 정보로만 썼다.
- 이전 회차 수치(0.67, 0.84)는 각 보고서 파일의 기록을 인용했고 다시 재지 않았다.

## Gaps

- pty 하네스, 골든, V-a~V-d 는 run 단계 산출물이라 실행하지 않았다(아직 코드가 없다).
- F1 의 S2 뮤턴트는 **현재** `profile_setup.go` 사본으로 관측했다. 흡수 뒤 S2 가 읽을 파일 구성은 설계 문서(§10)로만 판단했다. 설계가 `persistProjectConfig` 를 `profile_setup.go` 에 둔다고 적으므로(§2.2 "저장 이하 그대로") 같은 결과가 날 것으로 보지만 실행으로 확인하지는 않았다.
- 감시 목록 누락 탐색(E-2)은 흐름 파일과 직접 호출 대상까지만 읽었다. 간접 호출 전체 그래프는 걷지 않았다.
- 교차 모델 감사(`mcp__moai__audit_multi`·codex·GLM)는 호출문이 요구하지 않아 돌리지 않았다.
- AC-ITI-004 (3) 대조군 `grep -c 'charm.land/huh/v2 ' go.mod` → `1` 은 `research.md` §16 기록과 `research.md:201` go.mod 발췌로 판단했고, 이 실행에서 따로 돌리지 않았다.

## Residual-risk

- 이 레인 세션 환경에는 정리 목록 밖의 레인 변수(`MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, `MOAI_AUTONOMY_TIER`, `MOAI_SESSION_PID`, `MOAI_CONFIG_SOURCE`)가 있다(`env | cut -d= -f1` 이름만 확인). 읽는 코드는 `glm_task.go`·`todo.go`·`launcher_blockcap_infinite.go`·`factory.go`·`kanban.go`·`config/autonomy.go` 이고, `config.AutonomyTier()` 를 부르는 `internal/cli` 비테스트 코드는 없어 init·프로필 경로에 닿지 않는다고 판단했다. 제품이 나중에 이 변수를 init·프로필 경로에서 읽게 되면 자식 환경 정리 목록이 그것을 놓친다.
- `research.md` §14 의 감시 목록은 현재 호출 구조에서 도출했다. 흡수(t583) 뒤 새 홈 쓰기가 생기면 `plan.md` §C 6 의 재도출이 유일한 방어다.
- tmux 자식 실효 환경은 이번에도 실행으로 재지 않았다(설계상 run 단계 AC 가 관측한다).

---

## Iteration History

| 회차 | 트리 | 판정 | 점수 | 핵심 |
|---|---|---|---|---|
| 1 | `18144b7ac` 계열 | FAIL | 0.67 | D1~D17 |
| 2 | `9989aa512` | FAIL | 0.84 | N1(HOME 매니페스트 비이진)·N2(AC-003 스테퍼 전제) blocking major, N3~N5 blocking minor, N6~N10 optional |
| 3 | `02227fa12` | FAIL | 0.90 | N1·N2·N4·N5·N6~N10 해소. N3 의 S2 절 미해소(F1, 실측). optional F2~F5 |

정체(세 회차 연속 같은 결함) 표식 대상은 없다. N3 는 2회차에 처음 나왔고 3회차에 부분 해소됐다.

---

Provenance: plan-auditor (iteration 3, 독립 감사) · 트리 `02227fa1224eedc44bdeeb7050f2fbcdfff24fab` · 판정 빌드 = 트리 빌드(`go build ./cmd/moai`, `go test -c ./internal/cli`, 스크래치 경로) · 2026-09-11 · 작성자 추론 맥락 배제(M1)
