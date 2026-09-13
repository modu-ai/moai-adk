# SPEC Review Report: SPEC-INIT-TUX-I18N-001

- 카드: t586 · 감사 단계: plan · Iteration: 2/3 (Tier L 상한 3, `harness.plan_audit_tier_ceilings` L=3)
- 감사 대상 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `9989aa5124d9529a868d417806306fdf3773382c` (착수 시 `git rev-parse HEAD` 로 일치 확인, 보고서 작성 뒤 다시 확인)
- 대상 산출물 (Tier L 입력 5종 + 기록): `.moai/specs/SPEC-INIT-TUX-I18N-001/{spec.md,plan.md,acceptance.md,design.md,research.md,progress.md}` (v0.2.1, REQ 18 · AC 22)
- 이전 감사: `.moai/reports/t586/plan-audit.md` (iteration 1, FAIL 0.67, D1~D17). 재현 근거: `.moai/reports/t586/verdict.md`
- 작성자 추론 맥락은 M1 격리 원칙에 따라 배제했다. 호출문의 리드 판정 Q1~Q5 와 결정 D1~D6 은 판정 기준으로만 썼다.
- 앞선 iteration-2 시도는 API 속도 제한으로 아무것도 읽기 전에 중단됐고 결과가 없어 회차로 세지 않았다.

**Verdict: FAIL**
**Overall Score: 0.84** (네 차원 조화평균, Tier L 통과 기준 0.85)

must-pass 7개는 모두 통과했다. 1회차 결함 D1~D17 은 17건 모두 증거 수준에서 고쳐졌다. FAIL 을 가르는 것은 다시 쓴 pty 판정 계약에서 새로 생긴 결함 두 건이다(N1, N2). 둘 다 "공허 초록"이 아니라 **수리가 옳아도 FAIL 이 나는** 방향의 결함이라, MUST AC 가 이진 판정이 되지 못한다. 고칠 범위는 좁다(아래 § Recommendation). 점수는 0.67 → 0.84 로 올랐으므로 STOP 신호는 없다.

---

## Must-Pass Results

| # | 결과 | 판정 층 / 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | PASS | 요구 층. 정의 줄 `spec.md:169-198` 에 `REQ-ITI-001`~`REQ-ITI-018` 이 빠짐·중복 없이 3자리 패딩으로 있다(E-6). |
| MP-2 GEARS 형식 | PASS | **요구 층만 채점**했다. AC 의 Given-When-Then 은 검증 층 형식이라 여기서 보지 않았다. 18개 REQ 가 모두 정식 패턴 이름을 단다: Event-driven `:169,178,185,193`(모두 `When …, … shall`), State-driven `:198`(`While …`), Unwanted behavior `:187`(`shall not contain`), 나머지 Ubiquitous. 1회차 D16 의 `Event-detected` 라벨과 데이터 조건 `Where` 는 사라졌다. `REQ-ITI-017` 머리의 `Once …` 절은 형식 위반이 아니라 optional 표기 문제로 분류했다(N7). |
| MP-3 YAML frontmatter | PASS | `spec.md:2-14` 에 12개 필드가 모두 있다. `version: "0.2.1"`(따옴표 semver), `status: draft`, `created`/`updated: 2026-09-11`, `priority: P1`, `lifecycle: spec-anchored`, `tags` 쉼표 문자열. 거부 별칭(`created_at`·`labels`·`spec_id`) 없음. `tier: L`, `related_specs` 는 선택 필드. 트리 빌드 lint: 오류 0(E-5). |
| MP-4 언어 중립성 | N/A | 대상은 `internal/cli`·`internal/cli/wizard` Go 코드이고 `internal/template/templates/**` 는 바뀌지 않는다(`spec.md:161`). 16개 프로그래밍 언어 도구 열거 대상이 아니다. |
| MP-5 D7 교차 SPEC | PASS | 본문이 참조하는 SPEC 11개 가운데 `.moai/specs/` 에 있는 9개가 모두 `status: completed` 다. retired·superseded·archived 참조가 없어 BLOCKING 은 없다. `SPEC-INIT-QUIET-WIZARD-001` 은 이 트리에 없다(D7-5 SHOULD, N9) — SPEC 스스로 t583 이 미커밋 plan 단계라고 밝힌다(`spec.md:142`). |
| MP-6 D8 크로스 플랫폼 | PASS | `grep -c syscall` 결과 다섯 산출물 모두 0(E-6). D8-4 자동 PASS. |
| MP-7 명확화 게이트 | PASS | `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md spec.md acceptance.md design.md` → 출력 없음, 종료 1(E-6). `plan.md:138` "남은 확인 필요 표식은 없다". |

---

## Category Scores (0.0-1.0, rubric-anchored)

| 차원 | 점수 | 기준 구간 | 근거 |
|---|---|---|---|
| Clarity | 0.90 | 0.75~1.0 (1.0 쪽) | 요구 대부분이 해석 하나로 닫힌다. Q1~Q5 가 규범 텍스트에 판별 가능한 형태로 들어갔다(`spec.md:169-170,185-186,194`). 남은 흐림은 표기 수준이다: `REQ-ITI-017` 의 `Once …` 전제절(N7), tmux 부재를 `plan.md:37` 은 "FAIL 이 아니라 Gap", `acceptance.md:26` P2 는 "t.Fatal(FAIL)" 로 적는 표현 차이(N8), `spec.md:203` 의 catppuccin 서술 오류(N6). |
| Completeness | 0.90 | 0.75~1.0 (1.0 쪽) | 필수 절(HISTORY `spec.md:22`, 배경 §A, 요구 §B, 제외 범위 `### Out of Scope — …` 6개 `spec.md:213-239`), frontmatter, Tier L 산출물 5종과 `progress.md` §E.1~§E.4 뼈대가 모두 있다. `research.md` §13 이 미확인 항목을 정직하게 적고 V-a~V-d 로 넘겼다. 감점은 사실 서술 한 곳(N6)과 D4 연기 항목이 AC-ITI-003 선결로 연결되지 않은 점(N2 의 문서 측면)이다. |
| Testability | 0.70 | 0.50~0.75 (0.75 쪽) | 대부분의 AC 는 양의 존재 조건·대조군·뮤턴트를 갖춘 이진 판정이다. 그러나 MUST AC 두 개가 수리와 무관한 이유로 FAIL 한다: 실제 HOME 매니페스트 비교 대상 `~/.moai/` 는 파일 43,918개·8.4G 이고 5분 사이 63개가 다른 세션에 의해 바뀐다(N1, AC-ITI-003·020). AC-ITI-003 은 1회차 pty 캡처에서 사라진 것으로 측정된 스테퍼 줄을 판정 조건으로 쓴다(N2). 그 밖에 REQ-ITI-010 의 "every positive guard" 가 AC 뮤턴트로 절반만 확인되고(N3), 자체 [HARD] 규칙이 요구하는 대조군이 세 곳에 없다(N4). |
| Traceability | 0.90 | 0.75~1.0 (1.0 쪽) | 모든 REQ 에 AC 가 있고 고아 AC 가 없다(`acceptance.md:160`). 트리 빌드 lint 커버리지 경고 0 이며, 뮤턴트 두 개로 수집기가 이 형식을 실제로 읽음을 확인했다 — 끝에 붙인 미매핑 REQ, 그리고 문서 맨 끝 AC 두 개(019·020)의 매핑 제거(E-5). 감점은 REQ-ITI-010 한 절의 부분 커버(N3). |

조화평균 = 4 / (1/0.90 + 1/0.90 + 1/0.70 + 1/0.90) = 4 / 4.762 = **0.840**.

---

## Regression Check — iteration 1 결함 D1~D17 처분

| ID | 1회차 결함 | 처분 | 근거 (이번 실행에서 다시 잰 것) |
|---|---|---|---|
| D1 | `[NEEDS CLARIFICATION]` 4건 | **Fixed** | 다섯 산출물 grep 종료 1(E-6). Q1~Q5 판정표 `plan.md:128-136`, 각 판정이 REQ·AC 에 반영됨(아래 § 리드 판정 점검). |
| D2 | SPEC-CLI-TUI-MODERNIZE-001 계약 충돌 | **Fixed** | `related_specs` 에 포함(`spec.md:15`). §A.6.1 표가 뒤집히는 절 4개(REQ-TUIM-040·045, AC-TUIM-026·029)와 대체 보장을 적는다(`spec.md:129-134`). REQ-ITI-009 가 요구 층에서 대체를 선언하고(`:180`), 대체 보장은 AC-ITI-011 이 시험한다(`acceptance.md:96-99`, 대조군 `var wizardIsDark` 1줄 · 참/거짓 강제 골든). 인수 문장이 있다(`spec.md:138`). 원문 대조: `SPEC-CLI-TUI-MODERNIZE-001/spec.md:117,122`, `acceptance.md:116,119` 가 인용과 일치(E-4). REQ-TUIM-041 을 뒤집지 않는 이유도 적었다(`spec.md:136`). |
| D3 | 소스 스캔 가드 3건만 조사 | **Fixed** | `git grep -n 'ReadFile("profile_setup.go")' -- 'internal/cli/*_test.go'` → 9줄. 줄마다 감싸는 함수 이름이 §A.5 표 S1~S9 와 일치(E-1). 재조준 표 `design.md:158-168`, 비공허 뮤턴트 AC-ITI-010(3)(4). 잔여 커버리지 문제는 새 결함 N3 으로 분리했다. |
| D4 | `schemaSelectOptions` v1 타입·`ErrUserAborted` 누락 | **Fixed** | §A.4 "시그니처가 바뀌는 표면" 표(`spec.md:82-87`). 호출 재측정: 테스트 호출 6곳이 3개 파일(`profile_setup_nested_test.go:107`, `profile_setup_projectconfig_test.go:160`, `profile_setup_schema_options_test.go:49,68,92,108`), 정의 `profile_setup.go:124` 가 `[]huh.Option[string]` 반환(E-1). `ErrUserAborted` 는 `profile_setup.go:338,445`(E-1). |
| D5 | REQ 옵션 출처와 `model_policy` 예외 불일치 | **Fixed** | REQ-ITI-004 가 예외를 명시(`spec.md:175`). AC-ITI-005 (2) 가 동기화 결과(`user.yaml`·`language.yaml` 다섯 키)를 판정(`acceptance.md:70`). 인용 함수 실재: `template.ValidModelPolicies` `internal/template/model_policy.go:32`, `settings.FieldOptionDefs` `internal/settings/accessors.go:89`(E-7). |
| D6 | 보존 동작 4건 누락 | **Fixed** | REQ-ITI-005 (6)~(9)(`spec.md:176`), AC-ITI-006 사례 6~9(`acceptance.md:80-83`, 바이트 비교·키 부재·사전 선택값). |
| D7 | 존재하지 않는 명령 수준 테스트에 기댐 | **Fixed** | AC-ITI-007 이 이음새 주입 새 테스트로 두 진입 × (취소/오류/성공) × 이름 인자, 정리 이음새 1회와 clean-exit 인자, 진입 순서를 판정(`acceptance.md:84`). 현재 코드의 정리 호출 형태와도 일치(`research.md:227`). |
| D8 | pty 공허 초록 경로 네 가지 | **Fixed** (새 결함 N1·N5 파생) | (a) 양의 존재 P6 · (b) 기준 문자열 대기와 기한 초과 FAIL P5 · (c) 고유 이름·`t.Cleanup` 정확 이름 종료·센티널 생존·강제 실패/기한 초과 자기 검증 P7, AC-ITI-020 · (d) 임시 HOME·임시 cwd P4, 감시 경로 열거와 sha256 P8 — 요구한 수정은 모두 들어갔다(`acceptance.md:21-33,123-128`). 다만 (d) 를 구현한 방식이 새 결함 두 건을 낳았다(N1 감시 범위, N5 환경 격리). |
| D9 | REQ 대비 AC 부분 커버 | **Fixed** | AC-ITI-013 은 네 표면 × 네 로케일(다운그레이드 확인창 포함), AC-ITI-015 는 다운그레이드 확인창 + 도달성을 단정한 확인형 픽스처, AC-ITI-016 은 프로필 위저드 모든 그룹. 표면 출처표 `acceptance.md:39-49`. |
| D10 | 미결 답이 규범에 조용히 들어감 | **Fixed** | Q3 판별 사례 (a) 프로젝트 `ja` + 프로필 `ko` → `ja` 와 순서 뒤집기 뮤턴트(`acceptance.md:104,108`). Q1 은 질문 출현 횟수 1과 `runProfileSetup` 복원 뮤턴트(`:56`). Q2·Q4 는 AC-ITI-013·009. |
| D11 | `progress.md` 부재 | **Fixed** | `progress.md` 에 §E.1~§E.4 제목이 있고 §E.1 만 채워져 있다(`progress.md:3,32,36,40`). MCP `spec_audit` 도 era 를 V3R6 으로 판정(E-5). |
| D12 | AC 검색 범위·`go mod tidy` 미판정 | **Fixed** | AC-ITI-004 (1) 이 추적 비테스트 Go 파일 전체를 범위로 하고(`internal`·`cmd`·`pkg` 밖 12개 명시), (4) tidy 뒤 diff 없음, (5) Windows 크로스 빌드. 12개 재측정: `.moai/reports/` 5 · `.moai/scripts/` 2 · `scripts/` 5(E-1). v1 importer 5개 재측정 일치(E-1). 부수 서술 오류는 N6. |
| D13 | AC-ITI-009 예외 목록 없음 | **Fixed** | 닫힌 예외 목록 X(`acceptance.md:86-89`). 함수 이름으로 닫은 두 항목 실재: `template.ModelAliasPickerValues` `model_policy.go:139`, `models.ValidDevelopmentModes` `pkg/models/config.go:18`(E-7). |
| D14 | RED 원장 zsh 글롭 실패 | **Fixed** | 원장이 경로 인자를 인용한 `git grep` 으로 바뀌고 원문 출력과 판정 규칙("출력 0줄 = GREEN")이 있다(`acceptance.md:164-218`). L1 원문은 이번 재측정과 같다(E-1). |
| D15 | Tier 규모 초과 | **Fixed** | Tier L 로 올리고 `design.md`·`research.md` 를 보탰다(`plan.md:9-18`, `spec.md:14`). |
| D16 | GEARS 라벨 | **Fixed** | REQ-ITI-002 Ubiquitous(`spec.md:170`), 데이터 조건은 `When`(REQ-ITI-016 `:193`). |
| D17 | 절 번호·RED 마일스톤 | **Fixed** | `acceptance.md` 절이 §A·§B·§C·§D 로 이어지고, §D.1 에 AC 별 RED 마일스톤과 선결 열이 있다(`:130-154`). |

처분 합계: **Fixed 17 · Partially 0 · Not fixed 0.** 세 회차 내내 바뀌지 않은 결함이 없어 정체(stagnation) 표지는 없다.

---

## 리드 판정과 조건 점검

| 항목 | 결론 | 근거 |
|---|---|---|
| Q1 init 은 프로필 위저드를 부르지 않고 대화 언어는 한 번 | 인코딩됨·판별 가능 | REQ-ITI-001·002 `spec.md:169-170`. AC-ITI-001 이 stdin×플래그 조합 전부에서 이음새 호출 0 과 `No profile found`·`runProfileSetup(` 부재를 대조군과 함께 단정. AC-ITI-002 는 질문 목록에서 `conversation_language` 1회·인덱스 0 과 복원 뮤턴트. 현재 코드 좌표 재확인: 확인창 `init.go:647-663`, `update.go:174-192`, 위저드 이음새 `init.go:704`(E-7). |
| Q2 키별 짧은 라벨, 문장형 삭제 | 인코딩됨·판별 가능 | REQ-ITI-012 `spec.md:186`, 라벨 원천 표 `design.md:112-125`. AC-ITI-013 의 `git grep -n -E 'HelpSelect\|HelpInput' -- '*.go'` 0줄 + 같은 형태 대조군 `ConfirmYes\|ConfirmNo` 1줄 이상. 현재 참조 좌표 `translations.go:18-19,542-564`, `wizard_test.go:283-298,1256-1260` 재확인(E-7). huh v2 기본 라벨 원문 `keymap.go:107-183` 과 표 동작 이름이 일치(E-3). |
| Q3 프로젝트 → 활성 프로필 → 영어, 프로젝트 밖 AC | 인코딩됨·판별 가능 | REQ-ITI-011 `spec.md:185`. AC-ITI-012 사례 (b)(c) 가 프로젝트 밖, (a) 가 우선순위 판별, 뮤턴트 요구. 호출부 `update_version.go:325-331` 재확인(E-7). |
| Q4 흡수된 프로필 위저드도 같은 단계 표시 형식 | 인코딩됨·판별 가능 | REQ-ITI-008, AC-ITI-009(점 개수 합 = N, `<k> / <N>` 끝, init 대조군). `tui.Stepper` 는 total 개의 `●`/`○` 와 ` k / total` 을 그린다(`internal/tui/status.go:128-156`, E-3) — 규칙이 실제 출력 모양과 맞는다. |
| Q5 `agent_wiring`·`autonomy_tier` → `Agents & Autonomy`, 3→2 페이지 | 인코딩됨 | REQ-ITI-017 `spec.md:194`, `design.md` §9, `plan.md:136`. |
| Q5(a) 페이지 수와 분모의 독립 판정 | 성립 | `buildFormGroups` 는 조건 없는 연속 질문 중 `Group` 이 같은 것만 묶고(`wizard.go:176-189`), 스테퍼 분모는 보이는 질문 수다(`:236-239`). AC-ITI-018 뮤턴트(`autonomy_tier` → `Autonomy`)는 그룹 3·질문 4 → 018 FAIL, 021 은 그룹 머리 k=1,3,4 가 모두 `/ 4` 로 끝나 PASS. AC-ITI-021 뮤턴트(`user_name` 뒤 조건 없는 입력형 `Basic` 질문)는 그룹 2·질문 5 → 021 FAIL, 018 은 "첫 그룹에 두 질문이 있음"을 포함 관계로 읽으면 PASS. 두 뮤턴트는 각각 한 성질만 깬다. 018 의 "같은 뮤턴트에서 021 PASS 관측" 요구가 포함 관계 해석을 강제한다. |
| Q5(b) `questions.go` Group 편집이 게이트 뒤 | 성립 | `plan.md:115`(M7, 게이트 뒤) · §D 제약 `:44`. t583 확정 범위가 "남는 두 리터럴의 내용과 그룹 라벨"을 비접촉으로 둔다(t583 워크트리 `plan.md` §F.1:106, E-8). |
| Q5(c) 번역 키 없음 판단 | 재측정으로 확인 | `research.md` §6.1 명령 5개를 이 트리에서 다시 돌려 출력이 한 줄도 다르지 않았다(E-2). AC-ITI-022 의 grep(`[^A-Za-z0-9_.]`)은 wizard 비테스트 파일에서 `wizard.go:158,159,183,186,195,204,206` 을 내며, 모두 `huh.Group` 타입 이름이거나 `q.Group` 묶기 비교·대입이다(E-2). `\b` 점검: AC 어디에도 `\b` 가 없다. 증거 줄 `research.md:142` 에 옛 `\.Group\b` 형태가 남아 있으나 `:146` 이 공허 가능성을 명시하고 경계 없는 형태로 다시 쟀다. 재현해 보니 공허는 `git grep -E` 에서만 나타났다(`git grep -n -E 'q\.Group\b'` 종료 1, 셸의 ugrep 래퍼 `grep -n -E` 는 2줄 종료 0, E-2). 0건 grep 대조군은 AC-ITI-001·013·022 에 있고, AC-ITI-004 (2)(3)·011 (1) 에는 없다(N4). |
| V-a~V-d 선결 표시 | 성립 | §D.1 선결 열(`acceptance.md:132-154`)과 `plan.md:66-70` 이 일치한다. V-b → 012·013, V-c → 009, V-d → 020(2)(3), V-a 는 AC 없음. 표시 안 된 의존을 찾아봤다. AC-ITI-008 의 로케일 재렌더는 제목·옵션 함수가 결과 구조체가 아니라 `locale` 포인터에 묶여 있어(`wizard.go:309-320` `TitleFunc(…, locale)`, `OptionsFunc(optionsFn, locale)`) V-c 에 기대지 않는다. 설계 결정(`design.md` §4·§7)은 V 항목이 거짓으로 닫히면 리드에게 올린다는 규칙 아래 있어, 검증된 사실로 행세하지 않는다. 다만 V 목록 밖의 미확립 사실이 AC 하나를 받친다(N2). |
| D2 SPEC-CLI-TUI-MODERNIZE-001 인수 | 성립 | 위 D2 행. |
| t583 충돌 회피 전제 | 성립 | §A.7 표(`spec.md:144-151`)가 t583 워크트리의 `SPEC-INIT-QUIET-WIZARD-001/plan.md` §F.1(`:100-110`)과 내용이 같다. t583 은 여전히 미커밋이다(`git log --all -- .moai/specs/SPEC-INIT-QUIET-WIZARD-001` 출력 없음, t583 브랜치 HEAD `120436f58`, E-8). 이 트리의 좌표 `wizard.go:158,236,242,487`, `questions.go:490-491,513-514` 를 재확인했다(E-2, E-7). 다섯 파일을 건드리는 M4~M8 은 모두 게이트 뒤이고, 게이트에서 재측정한다(`plan.md:39,90-92`, `spec.md:157`). M1~M3 은 새 파일·`update_version.go`·`profile_setup.go`·테스트만 건드린다. |
| 판정 결정 D1~D6 준수 | 성립 | D1 확인창 제거(REQ-ITI-001), D2 흡수(REQ-ITI-003~010), D3 확인창 안쪽 빈 줄 유지(`spec.md:213-216`), D4 연기(`:218-220`), D5 설명 열 정렬(REQ-ITI-016), D6 pty 수리 판정 + 골든 회귀(`acceptance.md:5-11`). 결정 자체는 다시 열지 않았다. |

---

## Defects Found (structured defect-list)

N1. PTY-HOME-WATCH-NONBINARY — acceptance.md:L32 (P8), L57 (AC-ITI-003), L128 (AC-ITI-020); design.md:L179 — 실제 HOME 무기록 판정이 `~/.moai/` 와 `~/.claude/` **전체 트리**의 경로·크기·sha256 매니페스트를 실행 전후로 비교한다. 이 기계에서 `~/.moai` 는 파일 43,918개·8.4G 이고, 감사 도중 5분 사이 63개 파일이 이 SPEC 과 무관한 세션들 때문에 바뀌었다(세션 transcript `.jsonl`, `db/…/todo/backlog.db`, `claude-profiles/*/.claude.json` 백업, `run/*/…lock` — E-9). 감사 세션 자신의 transcript 도 그 트리 안에 있다. 제품 프로필 경로(`profile.GetBaseDir` → `~/.moai/claude-profiles/`, `internal/profile/profile.go:55-65`)와 런타임 세션 기록이 같은 하위 트리를 쓴다. 따라서 "매니페스트가 실행 전후 같다"는 하네스가 옳아도 사실상 늘 FAIL 이고, 수리를 판정하지 못한다. MUST AC 두 개(003·020)가 이진 판정이 아니다. 매 실행마다 8.4G 해시를 두 번 뜨는 비용도 따른다. — Severity: major — Class: blocking — Required fix: 감시 대상을 트리가 아니라 **제품이 실제로 쓸 수 있는 파일 목록**으로 바꾼다. 예: 실제 HOME 기준으로 계산한 프로필 `preferences.yaml` 경로들(`profile.GetBaseDir()` 아래 `*/preferences.yaml`), `MOAI_HOME` 계열 상태 경로, `~/.claude/settings.json`, 셸 rc 파일들. t583 이 쓰는 8항목 목록(t583 `plan.md:62`)과 같은 방식이다. 런타임 기록 하위 트리(`claude-profiles/*/projects`, `sessions`, `backups`, `debug`, `db`, `run`, `logs`, `state`, `worktrees`)는 명시적으로 뺀다. 감시 목록에 파일 하나를 심고 바꿨을 때 비교가 FAIL 함을 보이는 양성 대조군을 AC-ITI-020 에 더한다.

N2. AC003-STEPPER-PTY-PREMISE — acceptance.md:L41 (§C AC-ITI-003 도달성), L57; spec.md:L218-220; plan.md:L126 — AC-ITI-003(pty, MUST)은 "첫 캡처에 스테퍼 줄이 끝이 `1 / 4` 로 존재"를 판정 조건과 도달성 단정으로 쓴다. 그런데 이 SPEC 의 기준선인 80×30 tmux pty 캡처에서 init 첫 페이지의 스테퍼 줄은 **보이지 않았다**(`.moai/reports/t586/tty-wizard-firstpage-ko.txt` 1~2행 공백, 3행부터 제목; `verdict.md:118` "80×30 pty 에서 보이지 않음(원인 미확립)", E-10). 이 SPEC 은 그 현상(D4)의 원인과 수리를 제외 범위로 두고(`spec.md:218-220`), REQ-ITI-008 을 바로 그 이유로 뷰 문자열 기준으로 돌렸다. 그러면서도 pty AC 에는 같은 줄을 요구한다. t583 뒤 질문 수가 바뀌어 줄이 돌아올 수도 있지만 측정되지 않았고, `plan.md` §G 의 D4 재측정은 AC-ITI-003 의 선결로 연결돼 있지 않다. D4 가 남으면 AC-ITI-003 은 제외 범위 결함 때문에 FAIL 하고, run 단계는 범위 밖 수리를 강요받는다. V-a~V-d 밖에 있는 미확립 사실이 MUST AC 를 받치는 경우다. — Severity: major — Class: blocking — Required fix: 둘 중 하나. (a) AC-ITI-003 의 pty 판정에서 스테퍼 줄 조건을 빼고, 도달성은 기준 문자열 `Select conversation language` 와 옵션 줄 4개로 단정한다. `1 / 4` 는 뷰 문자열 판정(AC-ITI-021)에 맡긴다. (b) AC-ITI-003 의 §D.1 선결 열에 "D4 게이트 뒤 재측정에서 80×30 pty 에 스테퍼 줄이 보임"을 걸고, 보이지 않으면 판정 대신 리드에게 올린다고 적는다. (a) 가 더 단순하다.

N3. REQ010-POSITIVE-GUARD-COVERAGE — spec.md:L181; acceptance.md:L94-95 — REQ-ITI-010 은 "every positive guard shall fail when its asserted construct is removed" 를 요구한다. §A.5 기준 양성 성질을 가진 가드는 S1·S2(양성 절 `persistProjectConfig`)·S3·S4·S5·S6·S9 로 7개다(`spec.md:101-109`). AC-ITI-010 (4) 가 관측을 요구하는 양성 뮤턴트는 S1·S5·S9 세 개뿐이다. S2 양성 절·S3(정규화)·S4(스키마 빈 라벨)·S6(질문 id 집합 10개) 에는 "성질 제거 시 실패" 관측이 없다. (2) 의 빈 파일 뮤턴트는 기준 문자열 존재만 증명할 뿐 성질 제거를 증명하지 않는다. 문서가 스스로 세운 기준을 AC 가 절반만 확인한다. — Severity: minor — Class: blocking — Required fix: AC-ITI-010 (4) 에 나머지 네 개의 성질 제거 뮤턴트를 더한다(저장 경로 파일에서 `persistProjectConfig` 호출 삭제 → S2, 정규화 비교 삭제 → S3, `EmptyLabelFor("model_policy")` 를 리터럴로 바꿈 → S4, 프로필 질문 세트에서 질문 하나 삭제 → S6). 아니면 REQ-ITI-010 의 "every" 를 AC 가 실제로 덮는 가드로 좁힌다.

N4. ZERO-COUNT-WITHOUT-CONTROL — acceptance.md:L63, L64, L97 — `acceptance.md:19` [HARD] 는 "부재를 단정하는 AC 는 같은 명령 형태의 양성 대조군이 출력 1줄 이상을 내는 것을 함께 보인다"고 정한다. 세 곳이 이를 어긴다. AC-ITI-004 (2) `git grep -l '"github.com/charmbracelet/huh"' -- '*_test.go'` → 0줄(대조군 없음. (1) 의 대조군은 경로 명세가 다르다), AC-ITI-004 (3) `grep -c 'github.com/charmbracelet/huh ' go.mod` → `0`(대조군 없음), AC-ITI-011 (1) `git ls-files internal/cli/huh_theme.go internal/cli/huh_theme_test.go` → 0줄(없는 경로에도 출력 0·종료 0 이라 경로 오타나 잘못된 작업 디렉터리에서 공허 통과). — Severity: minor — Class: blocking — Required fix: 각각 같은 형태의 대조군을 붙인다. 예: `git grep -l '"testing"' -- '*_test.go'` 1줄 이상, `grep -c 'charm.land/huh/v2 ' go.mod` → `1`, `git ls-files internal/cli/wizard/wizard.go` → 1줄.

N5. PTY-CHILD-ENV-ISOLATION — acceptance.md:L28 (P4); design.md:L176; spec.md:L206 — P4 는 자식 프로세스의 `HOME` 과 `TERM` 만 정하고, 테스트 격리 제약은 HOME 만 언급한다. 그러나 제품은 HOME 밖의 환경 변수로도 홈 쪽 경로와 프로필을 푼다. `MOAI_HOME` 을 읽는 비테스트 파일은 `internal/paths/paths.go:69`, `internal/homestate/paths.go:68`, `internal/kanban/todo_root.go` 등 7곳이고, 현재 프로필 이름은 `CLAUDE_CONFIG_DIR` 로 정해진다(`internal/profile/profile.go:95`). 이 감사 환경에서 `CLAUDE_CONFIG_DIR=/Users/goos/.moai/claude-profiles/moai-adk` 가 설정돼 있고 `MOAI_HOME` 은 비어 있다(E-9). 부모 환경에 `MOAI_HOME` 이 실제 경로로 잡혀 있으면 임시 HOME 을 줘도 자식이 실제 상태 경로에 쓸 수 있다. 인프로세스 골든(AC-ITI-012 의 "활성 프로필" 사례)도 `CLAUDE_CONFIG_DIR` 을 통제하지 않으면 기대 로케일이 환경 따라 달라진다. tmux 세션이 서버 환경을 물려받는다는 부분은 이번에 실행으로 확인하지 않았다(추론). — Severity: minor — Class: blocking — Required fix: P4 와 `spec.md` §C 테스트 격리에 자식 환경 정리 목록을 넣는다(`MOAI_HOME` 은 임시 경로로, `CLAUDE_CONFIG_DIR` 은 비우거나 임시 경로로, `MOAI_KANBAN*` 은 비움). AC-ITI-012 의 활성 프로필 준비가 `CLAUDE_CONFIG_DIR` 을 명시적으로 정한다고 적는다. AC-ITI-020 의 "자식 HOME·작업 디렉터리가 실제와 다름을 출력" 절에 `MOAI_HOME`·`CLAUDE_CONFIG_DIR` 도 더한다.

N6. SPEC-C-CATPPUCCIN-CLAIM — spec.md:L203 — §C 는 `github.com/catppuccin/go v0.3.0` 을 huh v1 "그것만 끌어오던 간접 의존성"으로 적는다. 그러나 huh v2 도 catppuccin 을 import 한다(`charm.land/huh/v2@v2.0.3/theme.go:6`). 모듈 그래프에도 `charm.land/huh/v2@v2.0.3 github.com/catppuccin/go@v0.2.0` 이 있다(E-3). v1 import 가 0 이 된 뒤 tidy 해도 catppuccin 은 남는다. bubbletea v1.3.10·bubbles v1.0.0 은 그래프상 v1 경로로만 들어온다(`go mod why -m` 이 huh v1 경로만 보고). 이 서술은 허용 문구라 AC-ITI-004 (4)(tidy 뒤 diff 없음)의 판정에는 영향이 없다. — Severity: minor — Class: optional — Required fix: catppuccin 을 목록에서 빼거나 "huh v2 도 필요로 해 남는다"고 고친다.

N7. REQ017-ONCE-CLAUSE — spec.md:L194 — REQ-ITI-017 은 Ubiquitous 라벨인데 `Once the card t583 question set is absorbed,` 로 시작한다. `Once` 는 GEARS 수식어(Where/While/When)가 아니고, 흡수 게이트는 이미 §A.7 [HARD] 와 plan 이 관리한다. — Severity: minor — Class: optional — Required fix: 전제절을 지우고 "The init question set shall place …" 로 쓴다(게이트는 §A.7 에 둔다).

N8. TMUX-ABSENT-WORDING — plan.md:L37; acceptance.md:L26, L123 — tmux 부재 시 결과를 `plan.md` §C 4 는 "FAIL 이 아니라 실행 불가 Gap 으로 보고하고 run 을 멈춘다", P2·AC-ITI-019 (b) 는 "t.Fatal(FAIL)" 이라 적는다. 앞은 판정 기록의 분류, 뒤는 테스트 결과라 모순은 아니지만, 읽는 사람이 둘을 같은 층으로 읽을 수 있다. — Severity: minor — Class: optional — Required fix: plan §C 4 를 "테스트는 FAIL 로 끝나고(P2), E1 표에는 PASS 가 아닌 실행 불가 Gap 으로 적는다"로 맞춘다.

N9. D7-REF-NOT-FOUND — spec.md:L142; research.md:L185 — `SPEC-INIT-QUIET-WIZARD-001` 이 이 트리의 `.moai/specs/` 에 없다(D7-5 SHOULD). 문서 스스로 t583 이 미커밋 plan 단계라고 밝히므로 오타가 아니다. — Severity: minor — Class: optional — Required fix: 없음. 게이트 재측정 때 병합된 SPEC 경로를 progress 에 기록한다.

N10. AC019A-GLOBAL-SESSION-LIST — acceptance.md:L123 — AC-ITI-019 (a) 는 실행 전후 `tmux list-sessions` 전체 출력이 같기를 요구한다. 같은 기계의 다른 tmux 사용자(운영자 세션, `moai cg` 패널 등)가 그 사이 세션을 만들거나 닫으면 하네스와 무관하게 FAIL 한다. 감사 시점에는 tmux 서버가 떠 있지 않았다(E-9). — Severity: minor — Class: optional — Required fix: 비교를 `moai-ptycap-` 접두 세션으로 한정한다(AC-ITI-020 (1) 과 같은 방식).

(blocking: N1·N2(major), N3·N4·N5(minor). optional: N6·N7·N8·N9·N10.)

---

## Recommendation

FAIL 이다. must-pass 는 모두 통과했고 1회차 결함은 전부 닫혔으므로, 3회차는 아래 결함 차이만 좁혀 재감사한다.

1. **N1** — `acceptance.md` §B P8 과 `design.md` §11 "HOME 감시"의 감시 대상을 제품이 쓸 수 있는 파일 목록으로 바꾸고 런타임 기록 하위 트리를 뺀다. AC-ITI-020 에 심은 쓰기를 잡아내는 양성 대조군을 더한다. AC-ITI-003·020 의 "매니페스트 같음" 문장은 새 목록을 가리키게 한다.
2. **N2** — AC-ITI-003 의 pty 판정과 §C 도달성 표에서 `1 / 4` 스테퍼 줄 조건을 빼고 기준 문자열 + 옵션 줄 4개로 단정한다. 아니면 D4 재측정을 AC-ITI-003 의 선결로 건다.
3. **N3** — AC-ITI-010 (4) 에 S2 양성 절·S3·S4·S6 성질 제거 뮤턴트를 더한다.
4. **N4** — AC-ITI-004 (2)(3), AC-ITI-011 (1) 에 같은 형태의 대조군을 붙인다.
5. **N5** — P4·`spec.md` §C·AC-ITI-012·AC-ITI-020 에 `MOAI_HOME`·`CLAUDE_CONFIG_DIR`·`MOAI_KANBAN*` 통제를 적는다.

optional(N6~N10)은 같은 개정에서 함께 고치면 싸지만, 판정을 다시 여는 조건은 아니다.

---

## Evidence

모든 명령은 이번 실행에서 워크트리 `.claude/worktrees/t586`, HEAD `9989aa512` 위에서 돌렸다. 긴 출력은 줄였고, 줄인 곳은 표시했다.

### E-1 1회차 재측정 (가드·호출·importer·범위)

```
$ git grep -n 'ReadFile("profile_setup.go")' -- 'internal/cli/*_test.go'
internal/cli/profile_setup_model_policy_test.go:38
internal/cli/profile_setup_nested_test.go:27
internal/cli/profile_setup_nested_test.go:80
internal/cli/profile_setup_nested_test.go:117
internal/cli/profile_setup_projectconfig_test.go:145
internal/cli/profile_setup_removed_questions_test.go:61
internal/cli/profile_setup_removed_questions_test.go:87
internal/cli/profile_setup_removed_questions_test.go:116
internal/cli/schema_bridge_test.go:121          (9줄, 내용 열은 줄였다)
$ grep -n '^func Test' <다섯 파일>   (감싸는 함수만 발췌)
model_policy_test.go:37 TestProfileSetup_ModelPolicySelectPresent
nested_test.go:25 TestTUINestedConfigNoParallelWriter · :62 TestPermissionModeNormalizeAcceptEdits · :98 TestTUIEmptyLabelsSchemaSourced
projectconfig_test.go:143 TestProfileSetupConstructsProjectSelects
removed_questions_test.go:59 TestWizardOmitsRemovedQuestions · :85 TestWizardWritesNoStatuslineTheme · :114 TestWizardCarriesStoredSegmentsIntoPrefs
schema_bridge_test.go:109 TestTUIRendersSchemaFieldSet
$ git grep -n 'schemaSelectOptions' -- 'internal/cli/*.go'
profile_setup.go:124: func schemaSelectOptions(t profileSetupText, field string, withEmpty bool) []huh.Option[string] {
profile_setup.go:395,414,423,439 (비테스트 호출) · profile_setup_nested_test.go:107 · profile_setup_projectconfig_test.go:160 · profile_setup_schema_options_test.go:49,68,92,108
(주석 줄 profile_setup.go:108,384, projectconfig_test.go:154 제외)
$ git grep -l '"github.com/charmbracelet/huh"' -- '*.go' ':!*_test.go'
internal/cli/huh_theme.go
internal/cli/init.go
internal/cli/profile_setup.go
internal/cli/update.go
internal/cli/update_version.go
$ git grep -l '"github.com/charmbracelet/huh"' -- '*_test.go'
(출력 없음)
$ git ls-files -- '*.go'   (internal·cmd·pkg 밖 비테스트만 셈)
.moai/reports/: t114/measure_tool.go, t539/ctxsweep/main.go, t540/lab/ts.go, t582/repro/main.go, t95/budget_breakdown.go (5)
.moai/scripts/: lint-skip-cleanup.go, status-drift-cleanup.go (2)
scripts/: convert-nextra-to-hextra/main.go, docs-version-snapshot/main.go, i18n-validator/{diff,lockset,main}.go (5)
$ git grep -n 'ErrUserAborted' -- '*.go'
internal/cli/profile_setup.go:338 · :445 · internal/cli/wizard/unified_form_test.go:269,270 · internal/cli/wizard/wizard.go:136
```

참고: 첫 시도에서 `for` 루프·변수 치환이 섞인 복합 명령은 워크트리 세션 가드가 거부했다. 위 결과는 단일 명령으로 나눠 다시 잰 값이다.

### E-2 그룹 라벨 렌더 경로 (`research.md` §6.1 재실행, AC-ITI-022 grep, `\b` 재현)

```
$ git grep -n -E '\.Group([^A-Za-z0-9_]|$)' -- '*.go' ':!*_test.go'
(research.md:149-168 과 같은 19줄: huh_theme.go:107,108 · model.go:148,153 · root.go:121-123 · wizard.go:158,159,183,186,195,204,206,587,588 · lsp aggregator.go:47 · manager.go:43,44) exit=0
$ git grep -n -E 'q\.Group([^A-Za-z0-9_]|$)' -- 'internal/cli/wizard/wizard.go'
wizard.go:183 · wizard.go:186  exit=0
$ git grep -n -F 'Quality & Workflow' -- internal/cli/wizard/translations.go
exit=1
$ git grep -c -F 'ConfirmYes' -- internal/cli/wizard/translations.go
internal/cli/wizard/translations.go:6  exit=0
$ grep -n 'huh.NewGroup(fields' internal/cli/wizard/wizard.go
172:		groups = append(groups, huh.NewGroup(fields...))
$ git grep -n -E '\.Group([^A-Za-z0-9_.]|$)' -- 'internal/cli/wizard/*.go' ':!*_test.go'
wizard.go:158 (func buildFormGroups … []*huh.Group) · :159 (var groups []*huh.Group) · :183 (q.Group != pendingLabel) · :186 (pendingLabel = q.Group) · :195 (… *huh.Group) · :204 (주석 "creates a huh.Group") · :206 (… *huh.Group)  exit=0
$ git grep -n -F 'Agents & Autonomy' -- internal/cli/wizard/translations.go
exit=1
$ git grep -n -E 'q\.Group\b' -- internal/cli/wizard/wizard.go
exit=1                      ← q.Group 이 있는 트리에서 빈 출력(공허)
$ grep -n -E 'q\.Group\b' internal/cli/wizard/wizard.go
183: … · 186: …  ERE_exit=0 ← 셸의 grep 은 ugrep 래퍼 함수(`type grep` → "shell function")라 경계로 해석
$ grep -n -F '\b' spec.md plan.md acceptance.md design.md research.md progress.md
research.md:142 (옛 `\.Group\b` 서술) · research.md:146 (그 형태가 공허할 수 있다는 정정)  — AC 에는 없음
```

### E-3 모듈 그래프·huh v2 원문

```
$ go mod why -m github.com/charmbracelet/huh github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles github.com/catppuccin/go github.com/charmbracelet/lipgloss
huh:        internal/cli → github.com/charmbracelet/huh
bubbletea:  internal/cli → huh → bubbletea
bubbles:    internal/cli → huh → bubbles/filepicker
catppuccin: internal/cli → huh → catppuccin/go
lipgloss:   internal/cli → lipgloss      exit=0
$ go mod graph | grep -E ' (…bubbletea@v1|…bubbles@v1|…catppuccin/go@|…lipgloss@v1)' | grep -v '^github.com/charmbracelet/huh@'   (발췌)
charm.land/huh/v2@v2.0.3 github.com/catppuccin/go@v0.2.0
github.com/charmbracelet/glamour@v1.0.0 github.com/charmbracelet/lipgloss@v1.1.1-0.20250404203927-76690c660834
$ grep -rln 'catppuccin' …/charm.land/huh/v2@v2.0.3/ --include='*.go'
…/huh/v2@v2.0.3/theme.go
$ grep -n 'catppuccin' …/huh/v2@v2.0.3/theme.go
6:	catppuccin "github.com/catppuccin/go"
$ git grep -l '"github.com/charmbracelet/lipgloss"' -- '*.go'   (비테스트 발췌)
internal/cli/agentlint/agent_lint.go · github.go · help.go · uikit/{banner,render,styles}.go · version.go · wizard/styles.go · worktree/render.go · internal/statusline/{gradient,renderer,theme}.go  → huh_theme.go 외 소비자 있음(spec.md:203 서술과 일치)
$ sed -n 105,185p …/huh/v2@v2.0.3/keymap.go   (발췌)
Input: complete · back · next · submit
Select: back · select · submit · up · down · left(disabled) · right(disabled) · filter · set filter(disabled) · clear filter(disabled) · ½ page up · ½ page down · go to start · go to end
Confirm: back · next · submit · toggle · Yes(y) · No(n)
$ sed -n 115,156p internal/tui/status.go   (발췌)
func Stepper(current, total int, th *Theme) string { … for i := 1; i <= total; i++ { … "●" / "○" } … label " " + itoa(current) + " / " + itoa(total) }
```

### E-4 교차 SPEC 원문

```
SPEC-CLI-TUI-MODERNIZE-001/spec.md:117 REQ-TUIM-040 … shall remain two separate factories …
SPEC-CLI-TUI-MODERNIZE-001/spec.md:122 REQ-TUIM-045 … existing package-level indirection variables …
SPEC-CLI-TUI-MODERNIZE-001/acceptance.md:116 AC-TUIM-026 … grep -n "func moaiHuhStyles" internal/cli/huh_theme.go → 1 hit …
SPEC-CLI-TUI-MODERNIZE-001/acceptance.md:119 AC-TUIM-029 … grep -n "var huhThemeIsDark" internal/cli/huh_theme.go → 1 hit …
관련 SPEC status: CLI-TUI-MODERNIZE-001 · CLI-TUX-V3-002 · CLI-TUX-INIT-UPDATE-001 · CLI-WIZARD-RESTRUCTURE-001 · INIT-WIZARD-REPAIR-001 · INIT-HARNESS-PROMPT-001 · WEB-CONSOLE-002 · WEB-CONSOLE-003 · I18N-GOVERNANCE-001 → 모두 completed; INIT-QUIET-WIZARD-001 → NOT FOUND
삭제 표면(huh_theme·moaiHuh*·HelpSelect·"No profile found")을 계약으로 드는 다른 SPEC: SPEC-CLI-TUI-MODERNIZE-001 하나(spec·plan·progress·design·research 서술, 판정 절은 AC-TUIM-026·029)
```

### E-5 SPEC lint (판정 빌드 = 트리 빌드) 와 수집기 비공허성

```
$ go build -o <scratchpad>/moai-tree ./cmd/moai     → BUILD_EXIT=0   (tree HEAD=9989aa5124d9529a868d417806306fdf3773382c)
$ <scratchpad>/moai-tree version                     → v3.1.3 none built unknown   (ldflags 없이 이 트리에서 빌드)
$ <scratchpad>/moai-tree spec lint SPEC-INIT-TUX-I18N-001
INFO  OwnershipTransitionUnmeasured  spec.md 1  … commit d0ec7921f… has no Authored-By-Agent trailer …
0 error(s), 0 warning(s)   LINT_EXIT=0
뮤턴트 A (사본 spec.md 끝에 AC 없는 REQ-ITI-019 추가):
WARNING  CoverageIncomplete  …/mutA/spec.md  235  REQ REQ-ITI-019 is not referenced by any AC   0 error(s), 1 warning(s)
뮤턴트 B (사본 acceptance.md 에서 AC-ITI-019·020 의 maps 와 §D.2 의 REQ-ITI-018 을 REQ-ITI-099 로 바꿈, 남은 REQ-ITI-018 언급 1곳):
WARNING  CoverageIncomplete  …/mutB/spec.md  183  REQ REQ-ITI-018 is not referenced by any AC   0 error(s), 1 warning(s)
MCP spec_audit (project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t586, filter SPEC-INIT-TUX-I18N-001):
{"total_specs":1,"modern_era_clean":1,"drift_findings":[{"era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO"}]}
```

→ 수집기는 문서 맨 끝 목록형 AC 의 `(maps REQ-…)` 까지 읽는다. 원본의 경고 0 은 실제 커버리지 판정이다.

### E-6 구조 점검

```
$ grep -rn '\[NEEDS CLARIFICATION' plan.md research.md spec.md acceptance.md design.md   → exit=1
$ grep -c syscall spec.md plan.md acceptance.md design.md research.md                    → 0 0 0 0 0
REQ 정의 줄: spec.md:169,170,174,175,176,177,178,179,180,181,185,186,187,191,192,193,194,198 → REQ-ITI-001..018
AC 정의 줄: acceptance.md:55,56,57,61,68,74,84,85,90,91,96,103,109,110,114,115,116,117,118(021),119(022),123(019),124(020) → AC-ITI-001..022 (22개)
$ git diff --stat 18144b7aca714ea8924363b1eab4640cf101c6d0 HEAD -- internal cmd pkg go.mod go.sum  → (출력 없음) exit=0
```

### E-7 인용 좌표 대조

```
update_version.go:325  if !assumeYes && isatty.IsTerminal(os.Stdin.Fd()) && isVersionDowngrade(…)
update_version.go:327-331  huh.NewConfirm()… WithTheme(moaiHuhTheme())
init.go:647-663  "Auto-prompt profile setup" 블록 · init.go:704 runWizardFn(rootFlag, opts.ConvLang, opts.UserName) · :707 "Initialization cancelled."
update.go:174-192  같은 확인창 사본
init.go:83  Bool("non-interactive", …)
questions.go:66  Title: "Select conversation language"
questions.go:490-491  ID "agent_wiring", Group "Quality & Workflow" · :513-514 ID "autonomy_tier", Group "Autonomy"
translations.go:18-19 HelpSelect/HelpInput 필드 · :540 var uiStrings · :542-564 로케일 값
wizard_test.go:283-298, :1256-1260 HelpSelect/HelpInput 참조
wizard.go:135 mapFormErr · :290 key = opt.Label + " - " + opt.Desc · :487 buildConfirmField · :525 var wizardIsDark
config_helpers.go:18 func ReadLocaleFromProject
template/model_policy.go:32 ValidModelPolicies · :139 ModelAliasPickerValues · pkg/models/config.go:18 ValidDevelopmentModes
settings/accessors.go:89 FieldOptionDefs · :108 EmptyLabelFor
agent_wiring_question_test.go:87-88 그룹 단정 "Quality & Workflow"
```

(`wizard.go` 의 `var wizardIsDark` 는 525 줄이다. SPEC 은 이 줄 번호를 인용하지 않고 grep 으로 판정하므로 결함이 아니다.)

### E-8 t583 상태와 확정 범위

```
$ git worktree list | grep -i -E 't583|quiet'
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t583  120436f58 [WT-init-quiet-wizard] locked
$ git log --all --oneline -5 -- .moai/specs/SPEC-INIT-QUIET-WIZARD-001   → (출력 없음) exit=0
$ git show-ref --verify refs/heads/WT-init-quiet-wizard   → 120436f58af4986def4eab17e504ad359c688abe
$ git merge-base --is-ancestor 120436f58 HEAD             → exit=1 (서로 다른 develop 흡수 기준, 게이트에서 흡수 예정)
t583 워크트리 SPEC-INIT-QUIET-WIZARD-001/plan.md (미커밋, 수정 시각 Sep 11 10:49) §F.1:100-110 — questions.go·wizard.go·types.go·translations.go 행이 spec.md §A.7 표와 같다. "남는 두 리터럴(agent_wiring·autonomy_tier)의 내용과 그룹 라벨" 비접촉(:106), "548 이후 UI 문자열 표" 비접촉(:109).
```

### E-9 실제 HOME 규모·변동과 환경

```
$ find /Users/goos/.moai -type f | wc -l ; du -sh /Users/goos/.moai
   43918
8.4G	/Users/goos/.moai
$ find /Users/goos/.moai -type f -mmin -5 | wc -l
      63
$ find /Users/goos/.moai -type f -mmin -3   (발췌)
~/.moai/db/mo.ai.kr-840b641b/todo/backlog.db · ~/.moai/run/001-…/home-state-migration.json.lock ·
~/.moai/claude-profiles/moai-adk/.claude.json · …/backups/.claude.json.backup.1789091670997 ·
~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t586/b5f354f2-….jsonl (이 감사 세션 자신) ·
…/claude-worktrees-t583/286e9356-….jsonl · …/sessions/14540.json 외
$ find /Users/goos/.claude -type f -mmin -3   → (출력 없음)
$ sed -n 52,66p internal/profile/profile.go
// GetBaseDir returns ~/.moai/claude-profiles/.  … home, err := os.UserHomeDir() … return filepath.Join(home, profilesDir)
$ git grep -l 'EnvHome\b\|"MOAI_HOME"' -- 'internal/*.go' ':!*_test.go'
internal/cli/migrate_agency.go · internal/config/envkeys.go · internal/glmcred/glmcred.go · internal/homestate/paths.go · internal/kanban/state_dir.go · internal/kanban/todo_root.go · internal/paths/paths.go
$ printenv MOAI_HOME            → (없음) exit=1
$ printenv CLAUDE_CONFIG_DIR    → /Users/goos/.moai/claude-profiles/moai-adk
$ tmux list-sessions -F '#{session_name}'  → no server running on /private/tmp/tmux-501/default  exit=1
```

### E-10 D4 기준선 캡처

```
$ cat -n .moai/reports/t586/tty-wizard-firstpage-ko.txt   (80×30 pty, 18문항 트리)
 1	(빈 줄)
 2	(빈 줄)
 3	┃ 대화 언어 선택
 …
30	↑ up • ↓ down • / filter • enter select
→ 스테퍼 줄(● ○ … 1 / N) 없음. verdict.md:118 "80×30 pty 에서 보이지 않음 (원인 미확립)".
```

---

## Baseline-attribution

- 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `9989aa5124d9529a868d417806306fdf3773382c`. 제품 코드는 `18144b7ac` 와 같다(`git diff --stat … -- internal cmd pkg go.mod go.sum` 출력 없음).
- 판정 빌드: lint 는 이 HEAD 에서 `go build -o <scratchpad>/moai-tree ./cmd/moai` 로 만든 바이너리를 경로로 불렀다(버전 표기 `v3.1.3 none`, ldflags 없음). 설치된 `moai` 는 쓰지 않았다. MCP 서버 빌드는 `v3.2.0-rc.5 (commit 84fa4ece4)` 이고 `spec_audit` 한 건에만 썼다(era 판정 INFO, 판정 근거로 쓰지 않음).
- huh 원문: 모듈 캐시 `charm.land/huh/v2@v2.0.3`.
- t583 범위: t583 워크트리의 미커밋 `plan.md` 를 읽기 전용으로 읽었다.

## Gaps

- `go mod tidy` 는 실행하지 않았다. 작업 트리의 `go.mod` 를 바꾸기 때문이다. tidy 결과는 `go mod why`/`go mod graph` 로 추론했다(N6 은 그래프와 import 원문 근거).
- 뮤턴트 판정(AC-ITI-018·021 의 상호 독립)은 코드 읽기로 추론했다. 테스트를 짜서 돌리지 않았다. 흡수 트리가 아직 없어 돌릴 수도 없다.
- tmux 세션이 서버 환경을 물려받는지는 실행으로 확인하지 않았다(N5 의 해당 부분은 추론).
- `~/.claude` 는 3분 창에서 변동 0 이었다. 더 긴 창이나 다른 시간대의 변동은 재지 않았다.
- codex·glm 교차 감사 백엔드는 부르지 않았다. 호출문이 `audit_model` 을 지정하지 않았다.
- V-a~V-d 자체의 참·거짓은 run 단계 소관이라 판정하지 않았다.

## Residual-risk

- t583 이 확정 범위 밖을 고치면 §A.7 좌표와 AC-ITI-022 의 grep 허용 줄 목록이 달라질 수 있다. 게이트 재측정이 이를 잡도록 설계돼 있다.
- AC-ITI-018 의 "첫 그룹에 두 질문" 을 등가 비교로 구현하면 AC-ITI-021 뮤턴트에서도 018 이 FAIL 해 독립성 관측이 깨진다. AC 의 상대 PASS 관측 요구가 포함 비교를 강제하지만, 구현자가 문장만 보고 등가 비교를 택할 여지가 남는다.
- 번역 키를 두지 않는 판단은 "라벨을 그리는 경로가 없다"는 현재 코드에 기댄다. 흡수 뒤 새 프로필 빌더가 다른 표현으로 라벨을 읽으면 grep 이 놓칠 수 있다. 센티널 렌더 판정이 그 경우를 덮는다.

## Operational Notes (unverified)

- `inferred` — N1 을 고칠 때 감시 목록을 "제품 경로 함수가 실제 HOME 기준으로 돌려주는 값"에서 계산하면 목록이 코드와 함께 움직인다. 규칙: 경로를 손으로 열거하면 제품 경로가 바뀔 때 목록이 낡는다. 측정 방법: `profile.GetBaseDir()`·`paths` 계열 함수가 돌려주는 경로를 테스트 출력에 찍어 목록과 대조한다.
- `measured` — 이 기계에서 셸의 `grep` 은 ugrep 래퍼 함수라 `-E` 와 `\b` 조합이 `git grep -E` 와 다르게 동작한다(E-2 두 명령). 판정 명령은 계속 `git grep` 으로 쓰고, `\b` 는 쓰지 않는다.

---

Provenance: plan-auditor (iteration 2/3, Tier L) · 감사 트리 HEAD `9989aa5124d9529a868d417806306fdf3773382c` · 판정 빌드 `go build ./cmd/moai` @ 같은 HEAD (scratchpad `moai-tree`, `v3.1.3 none`) · MCP 서버 빌드 `v3.2.0-rc.5 84fa4ece4`(era INFO 한 건만) · 작성일 2026-09-11 · 이 파일은 감사자가 직접 썼다(릴레이 아님).
