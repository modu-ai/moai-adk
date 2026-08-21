---
id: SPEC-FEEDBACK-AUTO-SUBMIT-001
title: "자동 피드백 전송 — 동의 게이트 · 전송 전 마스킹 스크러버 · 취약점 분류"
version: "0.4.0"
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
- `moai feedback scrub` — 마스킹 변환 + 취약점 분류 + 판정 출력을 담당하는 신규 Go 명령과 그 뒤의 `internal/feedback` 패키지. 스크럽 대상은 **제목과 본문 둘 다**(REQ-3).
- `internal/sandbox/env.go` — 환경변수 이름 어휘 접근자 `DefaultEnvDenyList()` 신설(REQ-6). 이 SPEC이 다른 패키지를 편집하는 유일한 지점이다.
- 마스킹 로그(`.moai/logs/feedback-mask.log`, `0o600`)와 전송 실패 재시도 큐(`.moai/state/feedback/queue.json`).
- `.claude/skills/moai/workflows/feedback.md`(+ 템플릿 미러)의 제출 전 확인 게이트 + 스크러버 경유 [HARD] 조항.
- `moai init` 마법사 확인 질문 **1개**(자동 전송 동의)와 그 4로케일 번역.
- 웹 콘솔 토글 노출 — `feedback` 섹션을 `RouteSeam`으로 재개방(§D 결정 D5, 확정).
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

`moai feedback scrub`이 호출되면, 명령은 **제출될 모든 텍스트 입력** — 이슈 제목과 본문 — 에 마스킹 변환(REQ-4·5·6)과 취약점 분류(REQ-7)를 적용하고, 마스킹된 제목·본문과 기계 판독 가능한 판정을 표준 출력으로 반환해야 한다.

**[HARD] 제목은 본문과 같은 통제를 받는다(D1).** 제목은 `gh issue create`에 **별도로 조립되는 입력**이며(`.claude/skills/moai/workflows/feedback.md:84,102`), 사용자 자유 텍스트다. 본문만 스크럽하면 이 SPEC의 유일한 보안 통제가 선언된 범위에서 새고, 확인 게이트(REQ-2)가 마스킹 본문만 보여주므로 **제목 속 시크릿은 유일한 사람 체크포인트에서도 보이지 않는다**. 따라서 제목은 스크럽 대상이며, 마스킹된 제목만 제출·표시된다.

- 입력 표면: 본문은 표준 입력, 제목은 `--title` 인자. 파이프 관용구(`… | moai feedback scrub --title "<제목>"`)를 유지하기 위한 분리이며, 두 입력 모두 같은 파이프라인을 통과한다.
- 판정은 최소 `verdict`(`ok` | `blocked`), `title`(마스킹 제목), `body`(마스킹 본문), `findings`(항목별 종류·건수·위치), `reason`(blocked일 때의 사유)을 포함한다.
- `findings`의 각 항목은 어디에서 나왔는지(`title` | `body`)를 담아야 한다 — 확인 게이트가 "제목에서 1건 가려졌다"를 사용자에게 말할 수 있어야 하기 때문이다.
- 스크럽이 완료되면 종료 코드 0으로 종료한다. 도구 자체가 실패한 경우 0이 아닌 코드로 종료하며, 이때 호출자는 제출해서는 안 된다(shall not) — 제출 경로는 fail-closed다. 로깅·큐잉 경로의 fail-open(REQ-8·9)과는 다른 축이다.
- `findings`는 **마스킹된 값 자체를 담아서는 안 된다**(shall not). 종류·건수·위치만 담는다.

**프로젝트 루트 해석(D5)**: 스크러버는 REQ-8·9의 경로를 쓰기 위해 프로젝트 루트를 알아야 한다. 루트는 **현재 작업 디렉터리에서 `.moai/` 마커를 만날 때까지 상향 탐색**해 해석하며, `--root <path>` 인자가 주어지면 그 값이 탐색을 대체한다(테스트가 `t.TempDir()`를 루트로 주입하는 경로가 이것이다). 마커를 찾지 못하고 인자도 없으면 로그·큐 쓰기를 건너뛰되(fail-open, REQ-8·9) 스크럽 자체는 정상 완료한다 — 루트 부재가 마스킹을 막아서는 안 된다(shall not).

### REQ-4 — 시크릿 패턴 집합: 기존 정책 재사용 + 1건 확장

스크러버는 시크릿 탐지 패턴으로 `hook.DefaultSecurityPolicy().SensitiveContentPatterns`(`internal/hook/pre_tool.go:262-273`)를 재사용해야 한다. 이 집합은 이미 export돼 있고 대소문자 무시로 컴파일되며 `security.extra_sensitive_content_patterns`로 확장 가능하다.

여기에 `AIza[0-9A-Za-z_-]{35}`(Google API key)를 추가해야 한다 — `.moai/astgrep-rules/security/credentials.yml`에는 있고 `hook`의 목록에는 없다(D9: 리포에 세 번째 Go 패턴 목록 `internal/github/workflow/validator.go:155`가 있으나 워크플로 파일 검증용이라 재사용 대상이 아니다 — `research.md` §2에 조사됨-미채택으로 기록). 두 집합은 서로 포함관계가 아니므로 합집합을 취한다.

**[HARD] REQ-4 하위 조항 — 탐지기→재작성기 비대칭: 치환 span 규칙**

**탐지기로 건전한 패턴이 재작성기로도 건전하지는 않다(D1b).** 원본 집합은 `MatchString` 불리언 판정에 쓰이므로 "마커의 존재가 민감성을 증명한다"로 충분하지만, 매칭된 span을 그대로 치환하면 두 방향으로 무너진다.

1. **과소 마스킹(치명적)** — `-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----` 와 `-----BEGIN\s+CERTIFICATE-----`(`internal/hook/pre_tool.go:263-264`)는 **헤더 마커만** 매칭한다. span을 치환하면 헤더만 지워지고 **개인키 본문 전체가 공개 이슈로 실려 나간다** — 이 SPEC이 낼 수 있는 최악의 결과다.
2. **과잉 마스킹** — `compilePatterns`(`internal/hook/pre_tool.go:104`)가 `(?i)`를 붙이므로 `AKIA[0-9A-Z]{16}` 이 대소문자 무시가 된다. 권한 프롬프트에서는 무해하지만 재작성기에서는 사용자 산문 속 평범한 소문자 연속을 먹어치울 수 있다.

따라서 스크러버는:

- 매칭된 span을 **패턴이 시크릿 자체를 앵커할 때만** 치환해야 한다(shall).
- **마커 앵커 패턴**(`-----BEGIN …`)은 매칭 span이 아니라 **블록 종료자(`-----END … -----`)까지** 마스킹해야 한다(shall). 종료자를 찾지 못하면 입력 끝까지 마스킹한다 — 잘린 키 블록을 통과시켜서는 안 된다(shall not).
- 대소문자 민감성이 시크릿 형태의 일부인 패턴(`AKIA…` 등)은 재작성 시 **대소문자를 구별해** 적용해야 한다(shall). 원본 집합의 `(?i)` 를 그대로 재작성에 쓰는 것은 금지다(shall not).

관측: AC-F-024(개인키 블록 전체 부재), AC-F-008(소문자 산문 과잉 마스킹 없음).

치환 출력 형태는 기존 값 단위 마스커 3종(`internal/github/secret.go:144`, `internal/cli/glm.go:454`, `internal/cli/glm_tools.go:992`) 중 하나를 채택해 **통일**해야 하며, 네 번째 형태를 만들어서는 안 된다(shall not).

### REQ-5 — 절대 홈 경로 축약

본문에 사용자 홈 디렉터리로 시작하는 절대 경로가 포함된 경우, 스크러버는 해당 접두사를 `~`로 축약해야 한다.

홈 해석은 `paths.Home()`(`internal/paths/paths.go:49`)을 경유해야 한다 — 오버라이드된 `HOME`이 다른 소비자와 보조를 맞추도록 하는 계약이 `paths.go:8`에 명시돼 있다. 리포에 `~` 축약 헬퍼는 존재하지 않으므로 신규 작성이다.

### REQ-6 — 환경변수 값 마스킹

본문에 민감 환경변수의 값이 산문 형태로 포함된 경우, 스크러버는 그 값을 마스킹해야 한다.

민감 변수 이름 어휘는 `internal/sandbox`의 denylist와 `security.sandbox.env_scrub_extra`에서 가져온다. `ScrubEnv` 자체는 `KEY=VALUE` 슬라이스에서 변수를 제거하는 함수라 산문 속 값을 마스킹할 수 없으므로, 이름 어휘만 재사용하고 변환은 신규 작성한다.

**[HARD] 재사용 기제를 명시한다(D3).** `internal/sandbox/env.go:32`의 `defaultDenyList`는 **unexported**라 `internal/feedback`이 import할 수 없고, 그 표면의 유일한 export는 이 SPEC이 이미 배제한 `ScrubEnv`(`:51`)다. 따라서 `internal/sandbox`에 접근자 `func DefaultEnvDenyList() []string`를 신설하고 스크러버는 그것을 호출해야 한다(shall). 이름 목록을 `internal/feedback`으로 **복사해서는 안 된다**(shall not) — 복사는 두 목록의 드리프트를 보장하며 AP-4가 패턴 집합에 대해 금지한 것과 같은 실패다.

`internal/sandbox/env.go`는 이 SPEC의 편집 대상이며(§B In Scope, plan.md M2), 접근자는 `defaultDenyList`를 그대로 반환한다(사본을 반환해 호출자가 원본을 변경하지 못하게 한다).

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

**[HARD] 기존 로컬 초안 경로와의 관계(D4).** `.claude/skills/moai/workflows/feedback.md:36-44`에 이미 [HARD] 블록이 있어, `gh` **인증 실패·레이트리밋** 시 마스킹되지 않은 제목+본문을 `.moai/state/feedback-draft-<timestamp>.md`에 쓰고 "어떤 초안도 폐기되지 않는다"로 닫는다. 이 SPEC은 그 경로를 **대체하지 않고 실패 부류로 분리한다** — 두 표면은 서로 다른 시점의 서로 다른 실패를 담당한다.

| 실패 시점 | 담당 표면 | 내용 |
|---|---|---|
| **제출 이전** — `gh auth status` 실패, 레이트리밋 등 제출을 시작조차 못한 경우 | 기존 초안(`feedback-draft-<ts>.md`) | 스크럽 이전 원문(제목+본문). 사용자가 손으로 이어서 쓰는 작업본이다. |
| **제출 시도 이후** — `gh issue create` 자체가 실패한 경우 | 신규 재시도 큐(`queue.json`) | **마스킹된** 제목+본문. 재전송 대상이므로 스크럽을 통과한 형태만 담는다. |

즉 `gh issue create` 실패는 **큐가 쓰인다**. 초안 경로는 제출 전 단계의 작업 손실 방지이고 큐는 제출 후 단계의 재시도이며, 같은 실패에 둘 다 쓰이는 경우는 없다. 스킬 본문(REQ-10)은 이 분기를 명시해야 한다(shall).

**주의**: 초안 파일은 스크럽 이전 원문을 담으므로 **로컬 전용이며 어떤 경로로도 전송되지 않는다** — 이는 기존 동작이고 이 SPEC이 바꾸지 않지만, 두 파일이 같은 `.moai/state/` 아래 공존하므로 재전송 코드가 초안 파일을 큐로 오인해서는 안 된다(shall not).

### REQ-10 — 스킬 본문 구속 조항

`.claude/skills/moai/workflows/feedback.md`와 그 템플릿 미러는 **이슈 제목과 본문을** `moai feedback scrub`에 통과시킨 뒤 그 출력만을 제출하도록 지시하는 [HARD] 조항을 담아야 한다.

이 조항은 본문의 verbatim 보존 규칙(`feedback.md:104`)에 대한 **명시적 예외**로 기술해야 한다 — 암묵적 예외로 두어서는 안 된다(shall not).

**[HARD] 산문 표면이 지는 의무는 넷이며, 넷 다 관측 대상이다(N1).** 이 SPEC의 통제는 스스로 인정하듯 규약 수준 강제(§E.3)이고, 제목까지 범위를 넓히면서 스킬 본문이 지는 [HARD] 의무가 넷으로 늘었다. 본문이 그중 하나를 빠뜨려도 나머지 AC가 전부 통과하는 상태를 만들지 않기 위해, 넷을 **두 사본 각각에서** 기계적으로 관측한다(AC-F-019):

1. **제목 전달** — 스크러버 호출에 `--title` 이 실려 있을 것. 빠지면 제목이 마스킹 없이 `gh issue create`에 닿는다(D1이 지목한 그 경로).
2. **3문장 fail-closed 조항** — 종료 코드 ≠ 0 / `verdict != ok`(필드 부재·파싱 불가 포함) / 60초 무응답(`design.md` §9).
3. **실패 부류 분기** — `gh issue create` 실패는 큐, `gh auth status` 실패·레이트리밋은 기존 초안 경로(REQ-9 D4).
4. **`conversation_language` 라벨 의무** — 확인 게이트의 옵션 라벨과 findings 요약(`design.md` §7 D11).

### REQ-11 — `moai init` 마법사 동의 질문 1개

`moai init`이 대화형으로 실행되는 경우, 마법사는 "Quality & Workflow" 그룹에 자동 전송 동의 확인 질문 1개(`feedback_auto_submit`, 기본 `false`)를 제시해야 한다.

- 질문 정의는 `Page3Questions`(`internal/cli/wizard/questions.go`)에 추가한다. `DefaultQuestions`에 넣어서는 안 된다(shall not) — `TestQuestionOrder`(`wizard/questions_test.go:101`)가 5개로 고정하고 있다.
- en/ko/ja/zh 번역을 `internal/cli/wizard/translations.go`에 함께 실어야 한다(§A.1 정정 P3).
- 답변은 `saveBoolAnswer`(`wizard.go:459`) → `WizardResult` → `applyWizardPage3ToOpts`(`internal/cli/init.go:185`) 경로로 포착하고, 파일 기록은 `WritePhase1Configs`(`internal/core/project/initializer_expansion.go:30`)에서 `yamlpatch.PatchFile`로 수행해야 한다.
- 비대화형(`--non-interactive`, CI, TTY 부재)에서는 마법사가 실행되지 않으므로 컴파일 기본값이 그대로 적용된다.

### REQ-12 — 웹 콘솔 토글 노출

사용자가 웹 설정 화면을 열면, 웹 콘솔은 `feedback.auto_submit`을 불리언 토글로 렌더하고 그 변경을 `.moai/config/sections/feedback.yaml`에 영속해야 한다.

이를 위해 `feedback` 섹션은 `RouteExcluded`에서 `RouteSeam`으로 재개방돼야 한다(shall) — §D 결정 D5의 **선택지 A로 확정**했다(D8: 요구를 다시 쓰게 되는 결정을 산문으로 남겨 두지 않는다). 편집 지점은 넷이다: `internal/settings/schema_sections.go`(필드), `internal/settings/sectionroute.go`(라우트 + `ExcludedSections()` 제거), `internal/web/schemaform.go`(탭 + 패널). 4로케일 i18n 키를 등록해야 한다 — 누락 시 `TestI18nKeySetParity`(`internal/web/schema_label_test.go:74`)가 실패한다.

이 재개방은 **SPEC-WEBCONF-SIMPLIFY-001 M3의 기록된 결정을 되돌린다.** 그 결정을 고정하는 두 테스트(`internal/settings/sectionroute_test.go:27`, `internal/web/scope_contract_test.go:79`)의 기대값을 명시적으로 갱신해야 하며, 갱신 사실을 커밋 메시지에 반전으로 기록해야 한다(shall). 조용히 고칠 테스트가 아니다.

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

### 결정 D5 — 웹 노출 경로 — **해소됨: 선택지 A**

iter1에서 열어 두었던 결정이며, iter2에서 **선택지 A로 확정하고 선택지 B를 삭제했다**(감사 D8). 요구를 다시 쓰게 되는 결정을 산문으로만 남기면 기계적 clarification 게이트가 아무것도 보지 못하고, REQ-12가 "노출 방식은 나중에 정한다"는 유예를 담아 GEARS 요구가 되지 못했기 때문이다.

**확정: `feedback` 섹션을 `RouteExcluded` → `RouteSeam` 으로 재개방한다.** 편집 4곳(schema field + `sectionRoutes` 항목 + `consoleTabs` 항목 + `schemaSectionMetas` 패널) + i18n 키. 부수 효과로 `lens-web-todo.md` §A.3의 잠재 불일치(위조 POST가 `feedback.repository`를 seam 쓰기로 밀어 하드 에러를 내는 현상)가 해소된다.

**근거**: 카드가 키의 집(`feedback.yaml`)과 웹 토글을 **둘 다** 명시했다. 기각한 대안(토글을 `workflow.yaml` 아래 두기)은 편집 1줄로 싸지만 키가 `workflow.feedback.auto_submit`이 되어 카드가 명시한 키의 집을 포기해야 했다. 둘 중 하나를 포기하는 선택지보다 둘 다 지키는 쪽을 택했다.

**대가는 명시적으로 치른다**: 이 재개방은 SPEC-WEBCONF-SIMPLIFY-001 M3의 기록된 결정을 되돌린다. 그 결정을 고정하는 두 테스트를 명시적으로 갱신하고 커밋 메시지에 반전으로 기록한다(REQ-12). 운영자가 이 반전을 원치 않는다면 착수 승인 시점에 되돌릴 수 있으며, 그 경우 REQ-1·REQ-12와 이 절을 함께 개정한다 — 그것이 SPEC 개정 사안이라는 사실 자체가 이 결정을 산문 유예로 두지 않는 이유다.

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

[HARD] **양쪽 모두 같은 파일의 서로 다른 항목만 추가한다.** 기존 항목의 재배치·재작성·서식 변경을 하지 않는다.

**[HARD] 텍스트 충돌은 예외가 아니라 예상되는 결과다 — 해소 규칙.** "다른 항목만 추가"는 *의미* 충돌을 막을 뿐 *텍스트* 충돌을 막지 못한다. 두 SPEC은 같은 구조체 리터럴의 같은 위치(`questions.go` Page3 "Quality & Workflow" 그룹, `types.go`, `wizard.go` `saveBoolAnswer`, `translations.go` 3개 로케일 블록)에 인접 삽입하며, 인접 줄 삽입은 삽입 항목이 서로 달라도 git에서 충돌한다. 게다가 이 저장소는 전 티어가 PR 경로이므로(`.claude/rules/local/repo-local-pr-policy.md`) 두 SPEC은 같은 9개 파일을 건드리는 **동시 PR 2건**으로 착지한다.

1. **두 번째로 착지하는 쪽이 충돌 해소 소유자다.** 먼저 착지한 쪽은 자기 PR을 고치지 않는다.
2. **해소는 항상 양쪽 항목을 모두 보존한다.** 어느 한쪽 질문·필드·case·번역쌍을 버리는 해소는 금지다 — 버려진 쪽은 개수 고정 테스트가 아니라 **번역 완전성 테스트**에서 뒤늦게 드러난다.
3. **해소 중 기존 항목을 재배치하지 않는다.** 충돌 구간 밖의 줄은 손대지 않으며, 신규 두 항목의 상대 순서는 임의로 정하되 그 선택을 커밋 본문에 한 줄로 남긴다.
4. **해소 후 마법사 테스트를 다시 실행한다** — `TestQuestionOrder`(5개), `TestReconfigureQuestionsOrder`(12개), 번역 완전성 테스트가 두 질문이 공존하는 트리에서 통과함을 관측한다. `-v` 출력에서 `TestFeedbackAutoSubmitQuestion` 과 `TestTodoEnabledQuestion` 이 **둘 다 RUN 으로 찍히는지** 확인한다. 이 재실행이 해소가 끝났다는 유일한 근거다.
5. **해소가 4를 통과하지 못하면 되돌리고 리드에게 블로커로 보고한다.** 테스트를 고쳐 통과시키지 않는다.

이 규칙은 형제 SPEC(`SPEC-TODO-ENABLE-FLAG-001` §E.1)과 **같은 내용이다.** 두 문서가 같은 충돌 상황을 다르게 서술하면 규율이 아니라 모순이며, 나중에 착지하는 사람이 어느 쪽을 읽었느냐에 따라 다른 기대를 갖는다.

**`depends_on` 미기재 — 트레이드오프 기록**: 이것은 "의존이 없다"는 관찰이 아니라 **선택**이다. `depends_on`을 선언했다면 Phase 1 Depends_on Pre-flight 가 형제 SPEC이 `completed` 될 때까지 run-phase 진입을 막아 두 SPEC을 **직렬화**했을 것이고, 위의 공유 파일 위험 9종이 통째로 사라졌을 것이다. 그 대신 **동시성을 택했다** — 두 SPEC은 같은 릴리즈(v3.1.3) 배치에 실리고, 직렬화하면 뒤엣것이 앞엣것의 전 사이클(plan→run→sync)을 기다린다. 값은 위 해소 규칙으로 치른다. 기능 축에서 의존이 없다는 것(각자 독립 키·독립 표면)은 사실이지만 그것만으로는 생략을 정당화하지 못한다 — 생략이 사는 것은 동시성이고 치르는 것은 병합 충돌이다. 뒤집으려면(직렬화로 전환하려면) 양쪽 SPEC에 `depends_on`을 추가하고 양쪽 §E.1을 함께 고친다.

### §E.2 Hard 제약 — 그 밖

- 값 마스킹 출력 형태를 네 번째로 만들지 않는다(REQ-4).
- 마스킹 로그와 `findings` 어디에도 원문 값을 넣지 않는다(REQ-3, REQ-8).
- fail-open / fail-closed 축 분리: 로깅·큐잉은 fail-open, 제출 경로는 fail-closed.
- 로컬 검증은 패키지 스코프(`go test ./internal/<pkg>/...`)로만. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).
- 신규 테스트는 `t.TempDir()` 사용, 병렬 테스트에 `t.Setenv("HOME", ...)` 금지.

### §E.3 잔여 위험 — "기계화"의 한계를 과대 주장하지 않는다

**변환은 테스트 가능한 Go지만, 강제는 규약 수준이다.** `moai feedback scrub`은 마스킹 변환 자체를 결정적이고 테스트 가능하게 만든다. 그러나 제출을 수행하는 것은 여전히 스킬 본문을 따르는 오케스트레이터이며, **자기 지시문을 무시하는 세션은 스크러버를 우회할 수 있다**. 샌드박스가 아니라 기계적 변환 위에 얹은 규약 강제(enforcement-by-convention)다. 카드의 "기계화" 표현을 그대로 받아 "이제 마스킹이 강제된다"고 적는 것은 미검증 주장이다.

같은 한계가 REQ-2의 확인 게이트에도 적용된다 — `AskUserQuestion`은 오케스트레이터 전용 채널이라 Go에서 호출할 수 없다. 우회 폭을 좁히려면 `gh issue create` 자체를 Go 명령으로 감싸야 하며, 그 경로는 후속 카드 후보다(design.md §6).

**두 번째 잔여 위험**: 중복 검색(`feedback.md:71`)이 제출 결정 이전에 **스크럽되지 않은** 제목 유래 키워드를 GitHub로 보낸다. REQ-3이 제목을 스크럽 대상에 넣었으므로 *제출* 경로의 제목 누출은 닫혔지만, 이 검색 호출은 스크럽 이전 시점에 일어나고 이 SPEC은 그 호출을 변경하지 않으므로 위험이 남는다. 닫으려면 검색을 스크럽 이후로 옮기거나 마스킹된 제목으로 검색해야 하며, 그것은 기존 동작 변경이라 후속 카드다(design.md §10).

**세 번째 잔여 위험**: 취약점 분류기는 신규 작성이며 선례가 없다. 오탐과 미탐이 양방향으로 가능하다. 초기 임계값은 미탐보다 오탐 쪽으로 보수적으로 잡고, 오탐 시 사용자가 수동 경로를 안내받도록 한다.

## §F Cross-References

- 형제 SPEC: `SPEC-TODO-ENABLE-FLAG-001` (카드 t170의 todo 축)
- 설계: `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/design.md`
- 조사: `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/research.md`
- 근거 렌즈: `.moai/reports/t170/lens-{feedback,masking,init,web-todo}.md`
- `SECURITY.md` — REQ-7 거부 메시지가 인용할 정책 원문
- 관련 SPEC: SPEC-WEBCONF-SIMPLIFY-001(결정 D5 선택지 A가 되돌리는 대상), SPEC-INVOCATION-MODEL-001(`feedback.repository`의 유래)

## §G HISTORY

- **2026-08-22** v0.4.0 — plan-audit iter2 **FAIL 0.84**(Tier L 임계 0.85, must-pass 7/7 통과) 대응. Tier L 재시도 상한 회차이므로 **블로킹 2건만** 고치고 선택 4건(N3~N6)은 의도적으로 손대지 않았다 — 상한 회차에 비블로킹 편집으로 회귀를 만들지 않기 위함이다. iter2가 닫은 것들(제목 채널·탐지기 비대칭·D2의 실패 부류 차단·강제 정직성 언어)은 후퇴시키지 않았다.
  - **N1**(블로킹) — 제목 의무가 REQ-3·REQ-10·plan M6 **세 곳에 적혀 있는데 관측 수단이 없었다**. AC-F-019가 `grep -c 'moai feedback scrub'` 하나만 봤고, `--title` 언급은 CLI를 테스트하는 AC-F-003 안에만 있었다 — 그래서 **본문만 파이프하는 스킬 본문이 24개 AC를 전부 통과하고 제목이 마스킹 없이 제출되는 상태**가 가능했다(D1이 지목한 그 경로). 두 곳을 고쳤다: (a) REQ-10을 `이슈 제목과 본문을`로 수정하고, 산문 표면이 지는 [HARD] 의무가 넷(제목 전달 / 3문장 fail-closed / 실패 부류 분기 / `conversation_language` 라벨)이며 넷 다 관측 대상임을 명문화. (b) AC-F-019를 **두 사본 각각에서 5개 grep**을 도는 형태로 확장 — 새 AC 없이 예산 24/25 유지. 같은 성김에 덮여 있던 iter2 추가 3건(3문장 조항·D4 분기·라벨 언어)도 함께 관측 안으로 들어왔다.
  - **N2**(블로킹) — **AC-F-013이 어느 마일스톤에도 없었다.** iter2의 D7 수정(마일스톤 AC 범위 재매핑)이 떨어뜨린 것이다: 범위 합집합이 24개 중 23개를 덮고 F-013만 빠졌다. F-013은 **분류가 마스킹 이전 원문을 본다**는 순서 가드이고 그 역순은 `design.md` §3이 "조용한 미탐"이라 부른 것이라, 마일스톤 단위로 일하는 구현자가 한 번도 돌리지 않을 참이었다. M3(분류기) Exit를 `F-008, F-012, F-013`으로 고치고 빠뜨리지 말라는 근거를 함께 적었다.
  - **형제 §E.1 정렬**(리드 지시, 감사관 잔여 위험 3번) — 이 SPEC의 §E.1이 iter1 문구("충돌 없이 얹혀야 한다")로 남아 있어, 형제 `SPEC-TODO-ENABLE-FLAG-001`이 확립한 5조 해소 규칙과 **같은 상황을 다르게 서술**하고 있었다. 두 번째 착지자가 이쪽 문서를 읽으면 "충돌은 일어나지 않아야 한다"는 낡은 기대를 받는다. 형제와 동일한 5조 규칙(두 번째 착지자가 해소 소유자 / 양쪽 항목 보존 / 재배치 금지 / 재실행이 유일한 근거 / 실패 시 되돌리고 블로커 보고)과 `depends_on` 트레이드오프 기록을 그대로 반영했다.
  - 미처리(선택, 상한 회차 판단): N3 편집 지점 넷/파일 셋 표현, N4 `design.md §6`→`§10` 오참조, N5 REQ-6의 `그대로 반환`↔`사본을 반환` 자기모순(plan.md M2가 사본으로 해소), N6 끝까지-마스킹 폴백의 과잉 마스킹 비용이 §E.3에 없음. 넷 다 블로킹이 아니며 후속 개정 대상이다.
- **2026-08-22** v0.3.0 — plan-audit iter1 **FAIL 0.75**(Tier L 임계 0.85) + **MP-2 FAIL** 대응. 감사관이 "leaking scope boundary 를 가진 좋은 SPEC이지 약한 SPEC이 아니다"로 판정했고, 강제 주장의 정직성(§E.3 · design.md §1 · plan.md AP-12)은 유지 판정을 받아 **후퇴시키지 않았다**. 처리 내역:
  - **D1**(치명, 보안) — 이슈 **제목**이 스크러버를 통과하지 않고 `gh issue create`에 닿는 경로를 REQ-3이 열어 두고 있었다(제목은 `feedback.md:84,102`에서 별도 조립되는 입력이고, 확인 게이트도 본문만 보여주므로 유일한 사람 체크포인트에서도 보이지 않았다). **범위 밖 선언이 아니라 스크럽 대상 확장을 택했다** — REQ-3이 `--title` 인자를 받고 `title` 필드를 돌려주며, `findings`에 위치(`title`|`body`)가 추가된다. design.md §1/§4/§7과 AC-F-003·F-006도 함께 갱신.
  - **D1b**(치명, 보안) — 탐지기의 패턴 집합을 재작성기로 쓰는 것이 두 방향으로 불건전함을 REQ-4 하위 조항으로 명문화. **과소 마스킹**: `-----BEGIN … PRIVATE KEY-----`(`pre_tool.go:263-264`)는 헤더 마커만 매칭하므로 span 치환은 헤더만 지우고 키 본문을 공개 이슈로 보낸다 → **블록 종료자까지** 마스킹(종료자 부재 시 입력 끝까지). **과잉 마스킹**: `compilePatterns`(`:104`)의 `(?i)` 때문에 `AKIA[0-9A-Z]{16}`이 대소문자 무시가 된다 → 재작성 시 대소문자 구별 적용. 관측은 AC-F-024(개인키 블록 전체 부재) 신설 + AC-F-008에 소문자 `akia…` 산문 케이스 추가.
  - **D2**(블로킹) — AC-F-023의 `-run` 선택자 2개가 실재 테스트와 매칭되지 않았다(`TestSchemaLabel`·`TestSectionRoute` 둘 다 부재). `TestI18nKeySetParity`, `TestRouteForSectionTable|TestExcludedSectionsAllRejected`로 교체하고 `-v` 출력 줄 검사를 요구. acceptance.md 서두가 금지한 것을 자기가 어긴 지점이었다.
  - **D3**(블로킹) — REQ-6이 unexported `defaultDenyList`(`internal/sandbox/env.go:32`)의 재사용을 지시했다. 기제를 명시: `internal/sandbox`에 `DefaultEnvDenyList()` 접근자 신설(§B In Scope + plan M2), 복사 금지.
  - **MP-2 / D8**(블로킹) — REQ-12를 주어 있는 GEARS로 재작성("사용자가 웹 설정 화면을 열면, 웹 콘솔은 … 렌더하고 … 영속해야 한다"). 함께 **§D 결정 D5를 선택지 A로 확정하고 선택지 B를 삭제** — 요구를 다시 쓰게 되는 결정을 산문 유예로 두면 기계적 게이트가 아무것도 보지 못하기 때문이다. `[NEEDS CLARIFICATION]` 마커 대신 해소를 택했다.
  - **D4**(블로킹) — REQ-9의 재시도 큐가 기존 초안 경로(`feedback.md:36-44`)와 충돌 없이 공존하도록 **실패 부류로 분리**: 제출 이전 실패(auth·레이트리밋) → 기존 초안(스크럽 이전 원문), 제출 시도 이후 실패(`gh issue create`) → 신규 큐(마스킹본). 스킬 본문이 이 분기를 명시하도록 REQ-10 연계.
  - **D7**(블로킹) — plan.md 마일스톤 3개의 AC 범위 인용 정정(M4 → F-015~F-018, M6 → F-019~F-022, M7 → F-023).
  - **D5**(블로킹) — 스크러버의 프로젝트 루트 해석 규칙을 REQ-3에 명시(`.moai/` 마커 상향 탐색, `--root` 인자가 대체, 미해석 시 로그·큐 생략하되 스크럽은 완료) + design.md §9 반영.
  - **D6**(블로킹) — design.md §9에 두 행 추가(타임아웃 상한 → fail-closed, 파싱 불가 stdout → fail-closed), 스킬 [HARD] 조항을 3문장으로, `paths.Home()` 행의 부정확한 서술 정정(`os.UserHomeDir()` 폴백으로 대개 성공).
  - **D11**(블로킹) — design.md §7 게이트 옵션 3개와 findings 요약을 `conversation_language`로 낸다는 의무 추가(템플릿 미러는 영어 라벨).
  - **D9 · D10 · D12**(선택) — `Go 목록에는 없다` → `hook의 목록에는 없다`로 한정하고 `validator.go:155`를 research.md §2에 조사됨-미채택으로 기록 / `~` grep 개수 2 → 3 정정(`internal/shell/config.go:222`) / `version` 인용.
  - AC 23 → **24**(상한 25). D1b가 요구한 두 관측 중 하나(소문자 과잉 마스킹)는 이미 오탐 대조 AC인 AC-F-008에 케이스로 접었다 — 관측은 유지하면서 예산을 지키기 위함이며, 판정 해상도는 단언이 분리돼 있어 떨어지지 않는다.
- **2026-08-22** v0.2.0 — AC 예산 초과(32 > Tier L 상한 25)로 SPEC을 둘로 분리. todo 축을 `SPEC-TODO-ENABLE-FLAG-001`로 이관하고 이 SPEC은 Tier L(5종)로 재작성. AC 23개.
- **2026-08-22** v0.1.0 — 최초 초안(plan-phase). 카드 t170 전제 4건이 읽기 전용 렌즈 4종에서 반증돼 정정으로 기록, 운영자 결정 D1~D4 반영.
