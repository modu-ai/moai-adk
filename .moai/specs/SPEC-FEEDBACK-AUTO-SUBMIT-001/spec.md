---
id: SPEC-FEEDBACK-AUTO-SUBMIT-001
title: "자동 피드백 전송 — 동의 게이트 · 전송 전 마스킹 스크러버 · 취약점 분류"
version: 0.2.0
status: draft
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: High
phase: "v3.1.3"
module: "internal/feedback"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "feedback, masking, security, config, wizard, web-console, template-first"
related_specs: [SPEC-TODO-ENABLE-FLAG-001, SPEC-WEBCONF-SIMPLIFY-001, SPEC-INVOCATION-MODEL-001]
---

# SPEC-FEEDBACK-AUTO-SUBMIT-001

## §A Problem / Motivation

`/moai feedback`은 사용자가 발견한 문제를 공개 저장소(`modu-ai/moai-adk`)의 GitHub issue로 등록한다. 운영자 지시(2026-08-22, 카드 t170)는 자동 등록 옵션(`feedback.auto_submit`, 기본 off)과, 공개 채널 전송에 걸리는 보안 조항 3종(마스킹 / 취약점 제외 / 로그·큐잉)을 요구한다.

카드의 나머지 한 축인 **todo 기본 사용 설정은 별도 SPEC으로 분리**됐다 — `SPEC-TODO-ENABLE-FLAG-001`. 분리 사유: 초안의 AC가 32개로 Tier M 상한(16)과 Tier L 상한(25)을 모두 넘었고, 이는 예산을 완화할 게 아니라 쪼개라는 신호이기 때문이다(`spec-workflow.md` § SPEC Complexity Tier).

**Tier L 판단 근거**: 신규 패키지 `internal/feedback` 6파일 + `internal/cli`·`internal/config`·`internal/settings`·`internal/web`·스킬 2사본·템플릿·문서 4로케일 — 편집 파일 15개 초과, 신규 LOC 1000 이상 추정.

### §A.1 카드 전제 3건이 조사로 반증됨 — 정정 기록 (이 SPEC 소관분)

계획 단계 조사(읽기 전용 렌즈 4종, `.moai/reports/t170/`)에서 카드 전제 4건이 반증됐다. 그중 셋이 이 SPEC 소관이다(넷째 P4는 `SPEC-TODO-ENABLE-FLAG-001` 소관). 카드만 읽은 구현자가 같은 전제를 재도입하는 것을 막기 위해 명시한다.

**정정 P1 — "auto_submit=true가 제출 확인 질의를 생략한다"는 전제는 거짓. 그 질의는 존재하지 않는다.**
`.claude/skills/moai/workflows/feedback.md`의 `AskUserQuestion` 호출은 `:52`(필드 수집)과 `:156`(제출 이후)뿐이고 그 사이에는 없다. 영문 문서도 "Once you answer, a GitHub issue is created automatically"(`docs-site/content/en/utility-commands/moai-feedback.md:179`)라고 명시한다. **오늘의 제출은 이미 무조건적**이다. 이 SPEC은 질의를 생략하는 것이 아니라 **카드가 있다고 가정한 확인 게이트를 신설**한다(결정 D1).

**정정 P2 — "마스킹을 기계화한다"의 대상 경로에 Go 코드가 없다.**
피드백 경로 전체가 산문(오케스트레이터가 따르는 스킬 본문)이다. `feedback.md:118`의 `gh issue create`는 코드가 아니라 지시문이며 플래그 집합조차 리포지토리 어디에도 문자열로 존재하지 않는다. 오늘 상태에서는 **어떤 규칙도 기계적으로 강제할 수 없다**. 따라서 스크러버를 Go CLI로 신설해 변환 자체를 테스트 가능한 코드로 만든다(결정 D2). 강제력의 한계는 §E.3에 정직하게 적는다.

**정정 P3 — "init 안내 문구는 영어(16언어 중립성)"는 이 표면에 적용되지 않는다.**
마법사 프롬프트는 템플릿이 아니라 Go 소스(`internal/cli/wizard/questions.go`)에 있고, 번역은 `internal/cli/wizard/translations.go`에 ko/ja/zh 3로케일이 이미 존재한다. 템플릿 중립성 규율(CLAUDE.local.md §2.1)은 `internal/template/templates/` 산문을 대상으로 하므로 이 문자열에 닿지 않는다. 영어 단독으로 새 질문을 추가하면 `TestWizardQuestionTranslationCompleteness`(`internal/cli/wizard/translations_completeness_test.go:89`)가 **실패한다**. 두 언어 집합은 서로 다르다 — 템플릿의 16개는 프로그래밍 언어, 마법사의 4개는 대화 언어다(결정 D4).

### §A.2 운영자 결정 (재론 금지)

| ID | 해소한 전제 | 결정 |
|---|---|---|
| D1 | P1 | 카드가 있다고 가정한 **확인 게이트를 신설**한다. 제출 직전 마스킹된 본문을 보여주는 `AskUserQuestion` 라운드를 추가하고, `auto_submit: false`(기본)면 묻고 `true`면 건너뛴다. |
| D2 | P2 | **Go CLI 스크러버가 강제 지점**이다. `moai feedback scrub`이 조립된 본문을 받아 마스킹 본문 + 기계 판독 가능한 판정을 돌려주고, 스킬 본문은 [HARD] 조항으로 이 명령을 경유하도록 구속된다. |
| D4 | P3 | **기존 로컬라이제이션 관례를 따른다.** 이 SPEC이 추가하는 마법사 질문 1개는 en/ko/ja/zh 번역을 함께 싣는다. 공개 채널 게시 동의는 사용자가 자기 언어로 읽어야 할 종류의 질문이다. |

D3(todo 런타임 표면 한정)은 `SPEC-TODO-ENABLE-FLAG-001` 소관이다.

## §B Scope

**In Scope**

- `feedback.auto_submit`(bool, 기본 `false`) 설정 키와 그 접근자.
- `moai feedback scrub` — 마스킹 변환 + 취약점 분류 + 판정 출력을 담당하는 신규 Go 명령과 그 뒤의 `internal/feedback` 패키지.
- 마스킹 로그(`.moai/logs/feedback-mask.log`, `0o600`)와 전송 실패 재시도 큐(`.moai/state/feedback/queue.json`).
- `.claude/skills/moai/workflows/feedback.md`(+ 템플릿 미러)의 제출 전 확인 게이트 + 스크러버 경유 [HARD] 조항.
- `moai init` 마법사 확인 질문 **1개**(자동 전송 동의)와 그 4로케일 번역.
- 웹 콘솔 토글 노출(§D 결정 D5의 선택지에 따름).
- `feedback.auto_submit`의 템플릿 미러 + 키 인벤토리 등록 + `make build`.

**Out of Scope**

### Out of Scope — todo 기능 설정
- `workflow.todo.enabled` 키, todo 런타임 표면 억제, todo 마법사 질문, 그 웹 토글은 전부 `SPEC-TODO-ENABLE-FLAG-001` 소관이다.
- 두 SPEC이 같은 파일(`internal/cli/wizard/questions.go` 외)을 건드리는 사실과 그 충돌 회피 조건은 §E.1에 있다.

### Out of Scope — 피드백 경로의 기존 동작 변경
- 중복 검색(`feedback.md:71`, `gh issue list --search`)은 그대로 둔다. 이 호출이 제출 결정 이전에 이미 제목 키워드를 GitHub로 내보낸다는 사실은 §E.3에 잔여 위험으로 기록한다.
- 라벨 매핑(bug / enhancement / question)과 이슈 본문 템플릿의 구성 자체는 변경하지 않는다.
- `feedback.repository` 키와 그 접근자(`FeedbackRepository()`)는 손대지 않는다.

### Out of Scope — 기존 마스킹 소비자 확대
- `internal/telemetry`의 훅 트레이스, 기존 `.moai/logs/*` 로그 등 다른 출력 표면에 스크러버를 적용하는 작업은 범위 밖이다.
- `hook`의 `sensitiveContentPatterns` 원본에 `AIza`를 추가해 Write/Edit **deny** 판정을 넓히는 것도 범위 밖이다 — 이 SPEC은 스크러버 쪽 합집합으로 한정한다.

## §C Requirements (GEARS)

### REQ-1 — `feedback.auto_submit` 설정 키, 기본 OFF

`feedback.auto_submit`이 설정에 없거나 `false`인 경우, 피드백 워크플로는 제출 전 확인 게이트(REQ-2)를 거쳐야 하며 확인 없이 이슈를 생성해서는 안 된다(shall not).

키는 `.moai/config/sections/feedback.yaml`에 상주하고, Go 측은 `FeedbackConfig`(`internal/config/types.go:1310-1314`)에 필드를 추가하며 기본값 상수는 `internal/config/defaults.go`에 둔다. 부분 오버라이드 계약(`loader.go:281`이 채워진 기본값으로 wrapper를 seed)이 그대로 적용되므로 키를 생략한 YAML은 컴파일 기본값을 유지한다.

### REQ-2 — 제출 전 확인 게이트 신설

본문 조립과 스크러버 통과가 끝난 시점에 `auto_submit`이 `false`인 경우, 워크플로는 **마스킹된 본문 전문**을 사용자에게 제시하는 `AskUserQuestion` 라운드를 1회 수행해야 한다. 사용자가 제출을 선택하지 않으면 `gh issue create`를 실행해서는 안 된다(shall not).

`auto_submit`이 `true`인 경우 이 라운드는 생략된다. 이 게이트는 오늘 존재하지 않는 신규 표면이다(§A.1 정정 P1).

### REQ-3 — `moai feedback scrub` — 마스킹 변환의 단일 강제 지점

`moai feedback scrub`이 호출되면, 명령은 표준 입력으로 받은 본문에 마스킹 변환(REQ-4·5·6)과 취약점 분류(REQ-7)를 적용하고 마스킹된 본문과 기계 판독 가능한 판정을 표준 출력으로 반환해야 한다.

- 판정은 최소 `verdict`(`ok` | `blocked`), `body`(마스킹 본문), `findings`(항목별 종류·건수), `reason`(blocked일 때의 사유)을 포함한다.
- 스크럽이 완료되면 종료 코드 0으로 종료한다. 도구 자체가 실패한 경우 0이 아닌 코드로 종료하며, 이때 호출자는 제출해서는 안 된다(shall not) — 제출 경로는 fail-closed다. 로깅·큐잉 경로의 fail-open(REQ-8·9)과는 다른 축이다.
- `findings`는 **마스킹된 값 자체를 담아서는 안 된다**(shall not). 종류와 건수만 담는다.

### REQ-4 — 시크릿 패턴 집합: 기존 정책 재사용 + 1건 확장

스크러버는 시크릿 탐지 패턴으로 `hook.DefaultSecurityPolicy().SensitiveContentPatterns`(`internal/hook/pre_tool.go:262-273`)를 재사용해야 한다. 이 집합은 이미 export돼 있고 대소문자 무시로 컴파일되며 `security.extra_sensitive_content_patterns`로 확장 가능하다.

여기에 `AIza[0-9A-Za-z_-]{35}`(Google API key)를 추가해야 한다 — `.moai/astgrep-rules/security/credentials.yml`에는 있고 Go 목록에는 없다. 두 집합은 서로 포함관계가 아니므로 합집합을 취한다.

치환 출력 형태는 기존 값 단위 마스커 3종(`internal/github/secret.go:144`, `internal/cli/glm.go:454`, `internal/cli/glm_tools.go:992`) 중 하나를 채택해 **통일**해야 하며, 네 번째 형태를 만들어서는 안 된다(shall not).

### REQ-5 — 절대 홈 경로 축약

본문에 사용자 홈 디렉터리로 시작하는 절대 경로가 포함된 경우, 스크러버는 해당 접두사를 `~`로 축약해야 한다.

홈 해석은 `paths.Home()`(`internal/paths/paths.go:49`)을 경유해야 한다 — 오버라이드된 `HOME`이 다른 소비자와 보조를 맞추도록 하는 계약이 `paths.go:8`에 명시돼 있다. 리포에 `~` 축약 헬퍼는 존재하지 않으므로 신규 작성이다.

### REQ-6 — 환경변수 값 마스킹

본문에 민감 환경변수의 값이 산문 형태로 포함된 경우, 스크러버는 그 값을 마스킹해야 한다.

민감 변수 이름 어휘는 `sandbox.defaultDenyList`(`internal/sandbox/env.go:31-37`)와 `security.sandbox.env_scrub_extra`에서 가져온다. `ScrubEnv` 자체는 `KEY=VALUE` 슬라이스에서 변수를 제거하는 함수라 산문 속 값을 마스킹할 수 없으므로, 이름 어휘만 재사용하고 변환은 신규 작성한다.

### REQ-7 — 보안 취약점 내용 분류 시 자동 제출 거부

본문이 보안 취약점 보고로 분류되는 경우, 스크러버는 `verdict: blocked`를 반환해야 하고 워크플로는 `auto_submit` 값과 무관하게 `gh issue create`를 실행해서는 안 된다(shall not).

거부 메시지는 `SECURITY.md`가 이미 담고 있는 인간용 규칙을 인용해야 한다 — "Do NOT open a public GitHub issue for security vulnerabilities" 및 GitHub Security Advisories 경로(`https://github.com/modu-ai/moai-adk/security/advisories/new`). 새 정책을 발명해서는 안 된다(shall not).

분류는 **마스킹 이전 원문**을 대상으로 수행해야 한다 — 마스킹이 분류 신호(시크릿 패턴 적중)를 지우기 때문이다(design.md §3 참조). 분류기·CVE/CWE 탐지기·어휘는 리포에 전혀 없으므로 신규 작성이다.

### REQ-8 — 마스킹 로그의 로컬 잔존 (투명성)

스크러버가 마스킹을 수행한 경우, 무엇이 마스킹됐는지를 `.moai/logs/feedback-mask.log`에 append해야 한다.

- 파일 권한은 `0o600`이다(`internal/hook/failure_observer.go:156`의 관례). 주제가 시크릿 인접이므로 `internal/config/log.go`의 `0o644`를 따르지 않는다.
- 항목은 시각(RFC3339), 종류, 건수를 담고 **마스킹된 원문 값은 담지 않는다**(shall not).
- 쓰기 실패는 fail-open이다 — `slog.Warn`으로 강등하고 스크럽을 중단시키지 않는다.

### REQ-9 — 전송 실패 시 로컬 큐잉

이슈 생성이 실패한 경우, 워크플로는 마스킹된 본문과 메타데이터를 로컬 큐에 적재해야 하며 조용히 폐기해서는 안 된다(shall not). 이후 같은 항목의 전송이 성공하면 큐에서 제거해야 한다.

큐는 `internal/kanban.BacklogStore`(`internal/kanban/backlog_store.go`) 형태를 따른다 — 단일 JSON 파일, 형제 lock 파일, `Mutate()` 읽기-수정-쓰기, `internal/atomicfile.Replace` 경유 원자적 교체. 순수 append-only JSONL은 충분하지 않다(성공 시 삭제가 필요하다).

### REQ-10 — 스킬 본문 구속 조항

`.claude/skills/moai/workflows/feedback.md`와 그 템플릿 미러는 이슈 본문을 `moai feedback scrub`에 통과시킨 뒤 그 출력만을 제출하도록 지시하는 [HARD] 조항을 담아야 한다.

이 조항은 본문의 verbatim 보존 규칙(`feedback.md:104`)에 대한 **명시적 예외**로 기술해야 한다 — 암묵적 예외로 두어서는 안 된다(shall not).

### REQ-11 — `moai init` 마법사 동의 질문 1개

`moai init`이 대화형으로 실행되는 경우, 마법사는 "Quality & Workflow" 그룹에 자동 전송 동의 확인 질문 1개(`feedback_auto_submit`, 기본 `false`)를 제시해야 한다.

- 질문 정의는 `Page3Questions`(`internal/cli/wizard/questions.go`)에 추가한다. `DefaultQuestions`에 넣어서는 안 된다(shall not) — `TestQuestionOrder`(`wizard/questions_test.go:101`)가 5개로 고정하고 있다.
- en/ko/ja/zh 번역을 `internal/cli/wizard/translations.go`에 함께 실어야 한다(§A.1 정정 P3).
- 답변은 `saveBoolAnswer`(`wizard.go:459`) → `WizardResult` → `applyWizardPage3ToOpts`(`internal/cli/init.go:185`) 경로로 포착하고, 파일 기록은 `WritePhase1Configs`(`internal/core/project/initializer_expansion.go:30`)에서 `yamlpatch.PatchFile`로 수행해야 한다.
- 비대화형(`--non-interactive`, CI, TTY 부재)에서는 마법사가 실행되지 않으므로 컴파일 기본값이 그대로 적용된다.

### REQ-12 — 웹 콘솔 토글 노출

웹 설정 화면에서 `feedback.auto_submit`을 토글할 수 있어야 한다. 노출 방식은 §D 결정 D5의 두 선택지 중 착수 승인 시점에 확정한다. 어느 쪽이든 4로케일 i18n 키를 등록해야 한다(`internal/web/schema_label_test.go:96`).

### REQ-13 — Template-First 미러

`feedback.auto_submit`이 추가되면 `internal/template/templates/.moai/config/sections/feedback.yaml`에 같은 키를 미러하고 `make build`를 실행해야 한다. `internal/config/testdata/shipped_key_inventory.yaml`에도 항목을 등록해야 한다 — 미등록 시 `TestShippedConfigKeysHaveReaders`(`internal/config/shipped_key_reader_test.go:70`)가 실패한다.

템플릿 주석은 중립을 유지해야 한다(CLAUDE.local.md §2.1 — SPEC ID·REQ 토큰 금지).

## §D Evidence (조사에서 관측된 사실)

전부 `.moai/reports/t170/` 렌즈 보고서에서 인용하며 각 항목이 `file:line`을 동반한다. 재사용/신규 판정의 전체 표는 `research.md`에 있다.

| 사실 | 근거 |
|---|---|
| 피드백 경로의 `AskUserQuestion`은 `:52`, `:156`, `:178` 세 곳뿐 | `lens-feedback.md` §1 |
| `gh issue create`의 플래그 집합이 리포 어디에도 문자열로 없음 | `lens-feedback.md` §1 |
| `auto_submit`/`AutoSubmit` 전 리포 grep 0건 — 신규 키 | `lens-feedback.md` §3 |
| `FeedbackRepository()` 프로덕션 호출자 0건 — 설정 키가 Go에서 읽히지 않는 선례 | `lens-feedback.md` §3 |
| 리포에 텍스트 재작성형 `Redact*` 함수 0건 | `lens-masking.md` 헤드라인 |
| `~` 축약 헬퍼 0건 (grep 결과 2건은 모두 반대 방향) | `lens-masking.md` §4 |
| 취약점 분류기·CVE/CWE 탐지기·어휘 0건 | `lens-masking.md` §3 |
| 실패한 외부 전송을 재시도용으로 적재하는 스풀 0건 | `lens-masking.md` §5 |
| 마법사 번역 완전성 테스트가 영어 단독 신규 질문에서 실패 | `lens-init.md` §5 |
| `feedback` 섹션이 `RouteExcluded`이며 두 테스트가 그것을 고정 | `lens-web-todo.md` §A.3 |

### 결정 D5 — 웹 노출 경로 (착수 승인 시점 확정)

카드는 키의 집을 `.moai/config/sections/feedback.yaml`로 명시했고 이 SPEC은 그것을 따른다(REQ-1). 열려 있는 것은 **웹 콘솔 노출 방식**이다.

**선택지 A (기본안) — `feedback` 섹션을 `RouteSeam`으로 재개방.**
비용: 편집 4곳(schema field + `sectionRoutes` 항목 + `consoleTabs` 항목 + `schemaSectionMetas` 패널) + i18n 키. 대가: **SPEC-WEBCONF-SIMPLIFY-001 M3의 기록된 결정을 되돌린다**. 그 결정을 강제하는 두 테스트(`internal/settings/sectionroute_test.go:27`, `internal/web/scope_contract_test.go:79`)를 명시적으로 갱신해야 한다 — 조용히 고칠 테스트가 아니라 뒤집을 결정이다. 부수 효과로 `lens-web-todo.md` §A.3의 잠재 불일치(위조 POST가 `feedback.repository`를 seam 쓰기로 밀어 하드 에러를 내는 현상)가 해소된다. 카드 요구("web 설정 화면에서 토글 가능")를 문자 그대로 충족한다.

**선택지 B (더 싼 대안) — 토글을 `workflow.yaml` 아래 두고 `feedback` 라우트는 닫아 둔다.**
비용: 편집 1줄. 대가: 키가 `workflow.feedback.auto_submit`이 되어 카드 본문의 명시(`feedback.yaml`에 `auto_submit`)와 어긋나고, REQ-1의 키 집이 바뀌므로 SPEC 개정이 동반된다.

**권고**: A. 카드가 키의 집과 웹 토글을 모두 명시했고 B는 둘 중 하나를 포기해야 하기 때문이다.

## §E Constraints / Non-Goals

### §E.1 Hard 제약 — 형제 SPEC과 공유하는 파일

`SPEC-TODO-ENABLE-FLAG-001`과 이 SPEC은 **같은 파일 6종을 동시에 건드린다**. run-phase가 알아야 할 실제 충돌 위험이다.

| 공유 파일 | 이 SPEC이 추가하는 항목 | 형제 SPEC이 추가하는 항목 |
|---|---|---|
| `internal/cli/wizard/questions.go` | `feedback_auto_submit` 질문 1개 | `todo_enabled` 질문 1개 |
| `internal/cli/wizard/types.go` | `WizardResult` 필드 1개 | 필드 1개 |
| `internal/cli/wizard/wizard.go` | `saveBoolAnswer` case 1개 | case 1개 |
| `internal/cli/wizard/translations.go` | ko/ja/zh 3블록 × 1쌍 | 3블록 × 1쌍 |
| `internal/cli/init.go` | `applyWizardPage3ToOpts` 대입 1줄 | 대입 1줄 |
| `internal/core/project/initializer_expansion.go` | `feedback.yaml` writer | `workflow.yaml` writer |
| `internal/settings/schema_sections.go` | 필드 1줄 | 필드 1줄 |
| `internal/web/assets/i18n.js` | 4로케일 × 1쌍 | 4로케일 × 1쌍 |
| `internal/config/testdata/shipped_key_inventory.yaml` | 항목 1개 | 항목 1개 |

[HARD] **양쪽 모두 같은 파일의 서로 다른 항목만 추가한다.** 기존 항목의 재배치·재작성·서식 변경을 하지 않으며, 두 SPEC 중 어느 쪽이 먼저 착지하든 나중 것이 텍스트 충돌 없이 얹혀야 한다. 두 번째로 착지하는 쪽은 병합 후 마법사 개수 고정 테스트(`TestQuestionOrder` 5개, `TestReconfigureQuestions` 12개)와 번역 완전성 테스트를 **다시** 돌려 두 질문이 함께 통과함을 확인한다.

**`depends_on` 미기재 근거**: 두 SPEC 사이에 기능 의존이 없다 — 각자 독립된 설정 키를 읽고 독립된 표면을 바꾸며, 한쪽이 없어도 다른 쪽이 동작한다. 남는 것은 텍스트 충돌 위험뿐이고 그것은 순서 의존이 아니라 병합 규율(위 [HARD])로 다룬다.

### §E.2 Hard 제약 — 그 밖

- 값 마스킹 출력 형태를 네 번째로 만들지 않는다(REQ-4).
- 마스킹 로그와 `findings` 어디에도 원문 값을 넣지 않는다(REQ-3, REQ-8).
- fail-open / fail-closed 축 분리: 로깅·큐잉은 fail-open, 제출 경로는 fail-closed.
- 로컬 검증은 패키지 스코프(`go test ./internal/<pkg>/...`)로만. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).
- 신규 테스트는 `t.TempDir()` 사용, 병렬 테스트에 `t.Setenv("HOME", ...)` 금지.

### §E.3 잔여 위험 — "기계화"의 한계를 과대 주장하지 않는다

**변환은 테스트 가능한 Go지만, 강제는 규약 수준이다.** `moai feedback scrub`은 마스킹 변환 자체를 결정적이고 테스트 가능하게 만든다. 그러나 제출을 수행하는 것은 여전히 스킬 본문을 따르는 오케스트레이터이며, **자기 지시문을 무시하는 세션은 스크러버를 우회할 수 있다**. 샌드박스가 아니라 기계적 변환 위에 얹은 규약 강제(enforcement-by-convention)다. 카드의 "기계화" 표현을 그대로 받아 "이제 마스킹이 강제된다"고 적는 것은 미검증 주장이다.

같은 한계가 REQ-2의 확인 게이트에도 적용된다 — `AskUserQuestion`은 오케스트레이터 전용 채널이라 Go에서 호출할 수 없다. 우회 폭을 좁히려면 `gh issue create` 자체를 Go 명령으로 감싸야 하며, 그 경로는 후속 카드 후보다(design.md §6).

**두 번째 잔여 위험**: 중복 검색(`feedback.md:71`)이 제출 결정 이전에 제목 유래 키워드를 GitHub로 보낸다. 제목이 시크릿을 담고 있다면 확인 게이트가 뜨기 전에 이미 유출된다.

**세 번째 잔여 위험**: 취약점 분류기는 신규 작성이며 선례가 없다. 오탐과 미탐이 양방향으로 가능하다. 초기 임계값은 미탐보다 오탐 쪽으로 보수적으로 잡고, 오탐 시 사용자가 수동 경로를 안내받도록 한다.

## §F Cross-References

- 형제 SPEC: `SPEC-TODO-ENABLE-FLAG-001` (카드 t170의 todo 축)
- 설계: `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/design.md`
- 조사: `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/research.md`
- 근거 렌즈: `.moai/reports/t170/lens-{feedback,masking,init,web-todo}.md`
- `SECURITY.md` — REQ-7 거부 메시지가 인용할 정책 원문
- 관련 SPEC: SPEC-WEBCONF-SIMPLIFY-001(결정 D5 선택지 A가 되돌리는 대상), SPEC-INVOCATION-MODEL-001(`feedback.repository`의 유래)

## §G HISTORY

- **2026-08-22** v0.2.0 — AC 예산 초과(32 > Tier L 상한 25)로 SPEC을 둘로 분리. todo 축을 `SPEC-TODO-ENABLE-FLAG-001`로 이관하고 이 SPEC은 Tier L(5종)로 재작성. AC 23개.
- **2026-08-22** v0.1.0 — 최초 초안(plan-phase). 카드 t170 전제 4건이 읽기 전용 렌즈 4종에서 반증돼 정정으로 기록, 운영자 결정 D1~D4 반영.
