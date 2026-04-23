---
id: SPEC-V3-OUT-001
title: Output Style Contract Alignment with Claude Code Schema
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: Wave 4 SPEC writer
priority: P2 Medium
phase: "v3.0.0 — Phase 6b — Tier 2 Polish"
module: "internal/output/, .claude/output-styles/, outputStyles loader"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - gm#165
  - gm#166
  - gm#167
  - gm#168
  - gm#169
  - gm#175
  - gm#176
  - gm#177
  - gm#178
  - gm#179
  - gm#180
  - gm#181
  - gm#182
related_theme: "Theme 7 — Output Style Contract"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "output, style, frontmatter, cc-compat, rendering, tier2, v3"
---

# SPEC-V3-OUT-001: Output Style Contract Alignment with Claude Code Schema

## HISTORY

| Version | Date       | Author | Description                                   |
|---------|------------|--------|-----------------------------------------------|
| 0.1.0   | 2026-04-22 | Wave 4 | Initial SPEC draft per v3-master §3.7 / §8.5  |

---

## 1. Goal (목적)

moai-adk-go가 Claude Code의 네이티브 output-style 스키마와 호환되는 YAML frontmatter 형식을 채택하고, 파일 시스템 자동 탐색 경로(`.claude/output-styles/` 및 `~/.claude/output-styles/`)를 구현하며, 구조화된 출력(diff, validation errors, progress, OSC-8 file links, StatusIcon)을 CC의 Ink 렌더러가 해석 가능한 형태로 산출한다. **moai의 기존 MoAI output-style은 대체하지 않는다 — CC 스키마와 공존할 수 있도록 정렬만 수행한다.** moai 고유 메타데이터는 `moai:` 프리픽스 네임스페이스로 분리하여 충돌을 방지한다.

### 1.1 배경

Wave 1.4 §6.1-§6.7 및 Wave 1.5 §7에 따르면 Claude Code는 `outputStyles/loadOutputStylesDir.ts`를 통해 frontmatter 4개 표준 키(`name`, `description`, `keep-coding-instructions`, `force-for-plugin`)를 인식한다. moai-adk는 현재 `.moai/output-styles/` 하위에 자체 format을 사용하며, CC 스키마와의 충돌 가능성(Wave 1.4 §8.5 risk)이 존재한다. 또한 moai는 plain-text diff/errors/progress를 emit하여 CC의 StructuredDiff/ValidationErrorsList/ProgressBar 컴포넌트가 활용되지 못한다(gm#175-#182).

### 1.2 비목표 (Non-Goals)

- moai의 MoAI output-style 기본 제공 스타일의 **철회나 대체** (alignment only)
- Ink TUI, React 컴포넌트의 moai 측 재구현 (CC 측 렌더링에 위임)
- CC의 `Explanatory` / `Learning` built-in style의 moai 번들링 (CC runtime 측 책임)
- `/output-style` 슬래시 커맨드 재구현 (gm#125 deferred, out of scope)
- Plugin `force-for-plugin` 자동 활성화 로직 (SPEC-V3-PLG-001 scope 밖에서는 warn만)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/output-styles/` (project) 및 `~/.claude/output-styles/` (user) 디렉터리 자동 탐색
- CC 호환 YAML frontmatter 파서 (name, description, keep-coding-instructions, force-for-plugin)
- `moai:` 프리픽스 네임스페이스를 통한 moai 확장 메타데이터 (spec_id, version 등)
- `internal/output/` 패키지 신설: diff, errors, progress, code fence, OSC-8 links, StatusIcon 렌더러
- moai-default style의 CC frontmatter 스키마 변환
- `.claude/output-styles/moai-default.md` 템플릿 (이미 moai가 제공하는 스타일의 재포맷)
- TTY 감지 기반 ANSI color / OSC-8 fallback
- `moai doctor output-style` 서브커맨드 (frontmatter 검증)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- CC의 `outputStyles/loadOutputStylesDir.ts` 전체 재구현 — moai는 자체 파서만 제공하고 렌더링은 CC에 위임한다
- moai-adk-go 측 Ink/React 렌더러 — Go는 terminal escape sequences만 emit하고 CC가 파싱한다
- 기존 MoAI output-style SKILL.md 포맷의 **제거** — 공존 보장 (§2.2 Exclusion)
- 사용자가 작성한 custom output-style 파일의 자동 마이그레이션 — M06 이후 future migration으로 deferred
- OSC-8 미지원 터미널용 폴백 렌더링의 시각적 디자인 개선 — plain absolute path만 제공
- Windows `cmd.exe` 레거시 콘솔에서의 OSC-8 / StatusIcon 색상 보장 — Windows Terminal 이상만 지원

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/output/` 신설 패키지
- Claude Code v2.1.111+ (StructuredDiff, ValidationErrorsList, ProgressBar, StatusIcon 컴포넌트 가용)
- YAML 파서: `gopkg.in/yaml.v3 v3.0.1` (기존 의존성)
- TTY 감지: `github.com/mattn/go-isatty v0.0.21` (기존 의존성)
- 대상 OS: macOS, Linux (OSC-8 기본 지원), Windows (Windows Terminal 1.19+ 기준)
- 영향 디렉터리:
  - 신설: `internal/output/`, `.claude/output-styles/` (local + template)
  - 수정: `internal/cli/doctor.go` (output-style subcommand), `internal/template/templates/.claude/output-styles/`
- 외부 레퍼런스: Wave 1.4 §6.1-§6.7, §8.1-§8.5, Wave 1.5 §7.1-§7.5, master §3.7

---

## 4. Assumptions (가정)

- Claude Code의 outputStyles frontmatter 키(`name`, `description`, `keep-coding-instructions`, `force-for-plugin`)는 v3.0 개발 기간 동안 안정적으로 유지된다 (Wave 1.5 §7.2 evidence).
- 사용자는 프로젝트 레벨(`.claude/output-styles/`)과 사용자 레벨(`~/.claude/output-styles/`) 중 적어도 하나를 선호한다.
- `diff --git` 형식을 emit하면 CC의 StructuredDiff 렌더러가 정상적으로 파싱한다 (Wave 1.4 §8.4).
- OSC-8 하이퍼링크는 macOS Terminal.app, iTerm2, Windows Terminal 1.19+, 대부분의 Linux 터미널에서 지원된다 (`mattn/go-isatty` TTY 감지로 충분).
- StatusIcon 유니코드 글리프(✓ ✗ ⚠ ℹ ○ …)는 UTF-8 로케일에서 렌더링 가능하다.
- moai-default output-style은 CC의 frontmatter 스키마 변환 후에도 현재의 MoAI identity(SPEC-First DDD, TRUST 5, EARS)를 보존한다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-OUT-001-001**
The `internal/output/` 패키지 **shall** 단일 `Renderer` 타입을 통해 diff / validation errors / progress / file links / status icons / code blocks 렌더링을 제공하며, TTY 감지에 기반한 ANSI color on/off를 자동 전환한다.

**REQ-OUT-001-002**
The moai-adk-go **shall** `.claude/output-styles/*.md`(project scope)와 `~/.claude/output-styles/*.md`(user scope) 디렉터리를 자동 탐색하며, project scope가 user scope를 override한다 (CC `loadOutputStylesDir.ts:14-24` 호환).

**REQ-OUT-001-003**
The output-style YAML frontmatter 파서 **shall** CC 표준 키 4개(`name: string`, `description: string`, `keep-coding-instructions: bool|string`, `force-for-plugin: bool`)를 필수로 인식하고, 알 수 없는 루트 키는 경고 없이 무시한다 (forward-compat).

**REQ-OUT-001-004**
The output-style YAML frontmatter 파서 **shall** `moai:` 프리픽스 네임스페이스 하위의 moai 고유 메타데이터(예: `moai.spec_id`, `moai.version`, `moai.harness_level`)를 별도 객체로 보존하며, CC 표준 키와 충돌하지 않도록 한다.

**REQ-OUT-001-005**
The `diff` 렌더러 **shall** 모든 file modification 출력 시 `diff --git a/<path> b/<path>` 형식과 unified context를 emit하여 CC의 StructuredDiff 컴포넌트가 해석 가능한 출력을 생성한다 (Wave 1.4 §8.4).

**REQ-OUT-001-006**
The `validation error` 렌더러 **shall** 에러 리스트를 YAML 또는 JSON lines 형식으로 emit하며, 각 엔트리는 최소 4개 필드를 포함한다: `severity` (error|warning|info), `path`, `line`, `message`. 선택적으로 `suggestion` 필드를 포함할 수 있다.

**REQ-OUT-001-007**
The `code fence` 렌더러 **shall** 모든 code block 출력 시 triple-backtick과 language identifier를 명시한다 (예: ```` ```go ````, ```` ```yaml ````). 언어를 알 수 없는 경우 ```` ```text ```` 를 사용한다.

**REQ-OUT-001-008**
The `StatusIcon` 렌더러 **shall** 다음 6개 semantic glyph을 제공한다: ✓ (success), ✗ (error), ⚠ (warning), ℹ (info), ○ (pending), … (loading). TTY 환경에서는 ANSI color (green/red/yellow/cyan/gray/magenta)를 추가한다.

**REQ-OUT-001-009**
The `FilePath` 렌더러 **shall** 절대 경로와 line number를 받아 OSC-8 hyperlink (`\x1b]8;;file://<path>#L<N>\x1b\\<text>\x1b]8;;\x1b\\`)를 생성한다. TTY 미감지 시 `<path>:<line>` plain fallback을 emit한다.

### 5.2 Event-Driven Requirements

**REQ-OUT-001-010**
**When** moai가 긴 작업(10초 이상 예상)을 시작하면, the `Progress` 렌더러 **shall** 주기적으로 `Progress: N/M — <message>` 라인을 stderr에 emit한다. step number는 0-indexed이며 `M`은 총 단계 수다.

**REQ-OUT-001-011**
**When** 사용자가 `moai doctor output-style`를 실행하면, the subcommand **shall** `.claude/output-styles/` 및 `~/.claude/output-styles/` 하위의 모든 `.md` 파일을 로드하여 frontmatter를 검증하고, 누락된 필수 키 또는 타입 오류를 REQ-OUT-001-006 형식의 validation error로 출력한다.

**REQ-OUT-001-012**
**When** output-style frontmatter에 `force-for-plugin: true`가 설정되어 있으나 해당 스타일이 plugin capability가 아닐 때, the 파서 **shall** stderr에 경고 "force-for-plugin is valid only for plugin-sourced output styles; ignored"를 emit하고 해당 플래그를 false로 취급한다 (Wave 1.5 §7.2 호환).

**REQ-OUT-001-013**
**When** `internal/output/Renderer.Diff()`가 빈 old 또는 new 파일(새 파일 생성 / 삭제)을 받으면, the 렌더러 **shall** `diff --git a/<path> b/<path>` 헤더에 `new file mode` 또는 `deleted file mode`를 포함하여 Git-compatible 형식을 유지한다.

### 5.3 State-Driven Requirements

**REQ-OUT-001-014**
**While** stdout이 TTY가 아닌 상태(파이프, 리디렉션)에서, the `Renderer` **shall** ANSI color escape sequence를 생략하고, OSC-8 hyperlink 대신 plain `<path>:<line>` 형식을 emit한다.

**REQ-OUT-001-015**
**While** 환경 변수 `NO_COLOR`가 설정된 상태(빈 문자열 제외)에서, the `Renderer` **shall** TTY 감지 결과와 무관하게 모든 ANSI color escape을 생략한다 (no-color.org 표준 준수).

**REQ-OUT-001-016**
**While** harness level이 `thorough`인 상태에서, the Progress 렌더러 **shall** 각 단계마다 하위 milestone을 포함한 상세 진행률을 emit한다 (harness level이 `minimal`에서는 최종 요약만 emit).

### 5.4 Optional Requirements

**REQ-OUT-001-017**
**Where** 사용자가 `.claude/output-styles/`에 moai-default.md 외 custom style을 추가한 환경에서, the 시스템 **shall** 해당 custom style의 frontmatter와 body를 CC와 동일한 방식으로 로드하고, `moai:` prefix 메타데이터가 있으면 추가 파싱한다.

**REQ-OUT-001-018**
**Where** 터미널이 OSC-8 hyperlinks를 지원하지 않는 환경(TERM=dumb, CI 로그 등)에서, the FilePath 렌더러 **shall** plain text 폴백을 제공하며 동작 실패하지 않는다.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-OUT-001-019 (Unwanted Behavior)**
**If** output-style frontmatter의 `name` 필드가 빈 문자열이거나 누락되면, **then** the 파서 **shall** 해당 스타일 파일의 basename(확장자 제외)을 name으로 fallback 하고, 가능한 경우 stderr에 정보 로그를 emit한다 (CC `loadOutputStylesDir.ts:38` 동등 동작).

**REQ-OUT-001-020 (Unwanted Behavior)**
**If** output-style frontmatter에 `keep-coding-instructions` 값이 `true`/`false`/문자열 `"true"`/`"false"` 외의 타입이면, **then** the 파서 **shall** 해당 플래그를 `undefined`(unset)로 취급하고 CC runtime 기본값(false)을 따른다.

**REQ-OUT-001-021 (Complex: State + Event)**
**While** moai가 CI 환경(환경변수 `CI=true`)에서 실행 중이고, **when** 긴 작업 Progress가 시작되면, the Renderer **shall** human-readable Progress 라인 대신 JSON lines 형식(`{"type":"progress","step":N,"total":M,"message":"..."}`)을 stderr에 emit하여 CI 로그 파서가 해석 가능하도록 한다.

**REQ-OUT-001-022 (Unwanted Behavior)**
**If** `moai:` 네임스페이스 하위 메타데이터가 YAML 스펙을 위반하거나 무한 중첩(>5 level)된 경우, **then** the 파서 **shall** 해당 하위 트리만 무시하고 나머지 frontmatter를 계속 파싱하며, 경고를 stderr에 emit한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-OUT-001-01**: Given `.claude/output-styles/moai-default.md` 파일이 CC 호환 frontmatter를 갖춘 상태 When style_loader가 로드 Then 4개 표준 키(`name`, `description`, `keep-coding-instructions`, `force-for-plugin`) 파싱이 성공하고 `moai:` 네임스페이스 메타데이터도 별도 객체로 보존됨 (maps REQ-OUT-001-002, REQ-OUT-001-003, REQ-OUT-001-004).
- **AC-OUT-001-02**: Given file modification 렌더링 요청 When `internal/output/Renderer.Diff()` 호출 Then `diff --git a/<path> b/<path>` 형식과 unified context를 갖춘 Git-compatible 출력이 생성됨 (maps REQ-OUT-001-005, REQ-OUT-001-013).
- **AC-OUT-001-03**: Given validation errors 리스트 입력 When `internal/output/Renderer.ValidationError()` 호출 Then `severity`, `path`, `line`, `message` 4개 필드를 포함한 YAML 또는 JSON lines 출력이 생성됨 (maps REQ-OUT-001-006).
- **AC-OUT-001-04**: Given `.claude/output-styles/` 내 잘못된 frontmatter 파일 존재 When `moai doctor output-style` 실행 Then 누락 필수 키 또는 타입 오류를 validation error 형식으로 리포트 (maps REQ-OUT-001-011).
- **AC-OUT-001-05**: Given stdout이 TTY인 상태 When StatusIcon 렌더링 Then ANSI color escape가 포함된 colored glyph 출력; non-TTY (pipe/redirect)에서는 plain glyph만 출력 (maps REQ-OUT-001-008, REQ-OUT-001-014).
- **AC-OUT-001-06**: Given 환경변수 `NO_COLOR=1` 설정된 상태 When Renderer 호출 Then TTY 감지 결과와 무관하게 모든 ANSI color escape이 생략됨 (maps REQ-OUT-001-015).
- **AC-OUT-001-07**: Given 기존 MoAI output-style의 SPEC-First DDD / TRUST 5 reference 콘텐츠 When CC frontmatter 스키마로 변환 Then 텍스트 일치율 >95% 검증 통과 (maps REQ-OUT-001-001).
- **AC-OUT-001-08**: Given `CI=true` 환경에서 긴 작업 Progress 시작 When Renderer.Progress() 호출 Then human-readable 라인 대신 JSON lines 형식 `{"type":"progress","step":N,"total":M,"message":"..."}`이 stderr에 emit됨 (maps REQ-OUT-001-021).
- **AC-OUT-001-09**: Given Go code block 출력 요청 + 언어 식별자 "go" When `code fence` 렌더러 호출 Then 결과가 triple-backtick + "go" 식별자로 감싸지고 언어 미상일 경우 "text" 사용됨 (maps REQ-OUT-001-007).
- **AC-OUT-001-10**: Given 절대 경로 `/foo/bar.go`와 line 42를 전달한 상태 When TTY 환경에서 FilePath 렌더러 호출 Then OSC-8 hyperlink escape sequence `\x1b]8;;file:///foo/bar.go#L42\x1b\\` 형태로 emit됨 (maps REQ-OUT-001-009).
- **AC-OUT-001-11**: Given 10초 이상 예상되는 작업 시작 상태 When Progress 렌더러 호출 Then 주기적으로 `Progress: N/M — <message>` 라인이 stderr에 emit되고 step number가 0-indexed임 (maps REQ-OUT-001-010).
- **AC-OUT-001-12**: Given non-plugin style 파일에 `force-for-plugin: true` 설정된 상태 When 파서 로드 Then stderr에 "force-for-plugin is valid only for plugin-sourced output styles; ignored" 경고 emit 및 플래그 false로 취급됨 (maps REQ-OUT-001-012).
- **AC-OUT-001-13**: Given harness level이 `thorough`인 상태 When 긴 작업 Progress 호출 Then 각 단계별 하위 milestone이 상세 출력되며, harness가 `minimal`일 때는 최종 요약만 emit됨 (maps REQ-OUT-001-016).
- **AC-OUT-001-14**: Given 사용자가 `.claude/output-styles/custom.md`에 moai 확장 메타데이터를 작성한 상태 When style_loader 로드 Then 4개 표준 키 + `moai:` prefix 메타데이터 모두 파싱되어 별도 객체로 보존됨 (maps REQ-OUT-001-017).
- **AC-OUT-001-15**: Given `TERM=dumb` 환경 상태 When FilePath 렌더러 호출 Then plain text `<path>:<line>` 폴백이 emit되고 OSC-8 escape는 사용되지 않음, 렌더링 실패 없음 (maps REQ-OUT-001-018).
- **AC-OUT-001-16**: Given output-style 파일에 `name` 필드가 빈 문자열 또는 누락 상태 When 파서 실행 Then basename(확장자 제외)을 name으로 fallback하고 stderr에 정보 로그 emit (maps REQ-OUT-001-019).
- **AC-OUT-001-17**: Given `keep-coding-instructions: 42` 등 허용되지 않는 타입 상태 When 파서 실행 Then 플래그를 undefined로 취급하고 CC runtime 기본값(false)을 따름 (maps REQ-OUT-001-020).
- **AC-OUT-001-18**: Given `moai:` 네임스페이스 하위 메타데이터가 YAML 스펙 위반 또는 6-level 이상 중첩 상태 When 파서 실행 Then 해당 하위 트리만 무시되고 나머지 frontmatter 파싱 계속됨, 경고 stderr emit (maps REQ-OUT-001-022).

---

## 7. Constraints (제약)

- moai-adk-go의 9-direct-dep 정책 준수: 신규 외부 의존성 금지. 모든 diff/YAML 파싱은 기존 stdlib + `yaml.v3` + `go-isatty`로 구현.
- 템플릿 우선 원칙(CLAUDE.local.md §2): `internal/template/templates/.claude/output-styles/moai-default.md`를 먼저 배치 후 `make build`.
- 언어 중립성(CLAUDE.local.md §15): output-style 템플릿은 특정 언어에 편향되지 않는다.
- 기존 MoAI output-style의 functional equivalence 보장 (content 변환, 의미 보존).
- CC 표준 키와 `moai:` 네임스페이스는 YAML 루트에서 **반드시 분리**되어야 한다 (충돌 금지).
- TTY/NO_COLOR 감지 로직은 이미 moai가 statusline 렌더링에 사용하는 패턴을 재사용한다.
- Progress stderr 출력 경로: stdout은 JSON response 전용(SPEC-HOOK-001 REQ-HOOK-013 원칙 준수).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| CC outputStyles frontmatter 키가 v3 개발 도중 변경 | 파서 호환성 깨짐 | Wave 1.5 §7.2 증거를 기반으로 고정했으며, `moai doctor output-style`이 unknown key를 경고로 처리하도록 forward-compat 설계 |
| 기존 moai-default 스타일의 의미 손실 | UX 저하 | 변환 후 regression 검증: 기존 테스트 픽스처를 재사용하고 텍스트 일치율 >95% 검증 |
| TTY 감지 오판 (nested shell, SSH) | color escape sequences가 로그에 누적 | `NO_COLOR` 환경 변수 지원 + CI 변수 감지로 이중 safeguard |
| OSC-8 미지원 터미널에서 텍스트 깨짐 | UX 저하 | REQ-OUT-001-018의 plain fallback이 필수 요구사항 |
| `moai:` 네임스페이스가 CC의 향후 버전에서 충돌 | 호환성 깨짐 | v3.0 이후 CC 변경사항 모니터링, 필요 시 `x-moai-*` 이름 공간으로 재배치 (하위 호환) |
| Windows 레거시 콘솔에서 UTF-8 glyph 깨짐 | StatusIcon 깨짐 | Windows Terminal 1.19+ 요구사항 명시 (§3 Environment). Legacy cmd.exe는 지원 대상 외 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- 없음 (본 SPEC은 Phase 6b polish로, Phase 1 foundation SPEC에 하드 의존하지 않음)

### 9.2 Blocks

- SPEC-V3-PLG-001 (Plugin system v1) — Plugin이 output-style capability를 갖는 경우 본 SPEC의 `force-for-plugin` 경로 논리를 기반으로 확장. **v3.0 scope에서는 outputStyles capability는 Plugin v1에서 제외**이므로 현재는 warn만 emit.

### 9.3 Related

- SPEC-V3-SCH-001 (Formal config schemas) — 비차단. 본 SPEC은 config.yaml의 `design.yaml` 스키마와 독립적이나, output rendering 옵션이 design.yaml에 추가될 가능성에 대비해 스키마 등록 경로에 접근 필요.
- SPEC-V3-HOOKS-001 (Hook Protocol v2) — Hook `systemMessage` 출력은 StatusIcon 스키마와 일관성 유지 권장.
- SPEC-V3-MIG-001 / MIG-002 (Migration framework) — v3.0 최초 실행 시 기존 `.moai/output-styles/` (구 MoAI format) 파일을 `.claude/output-styles/`로 마이그레이션하는 단계는 M06(optional, post-v3.0) 후속에서 수행.

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 요구사항 ID는 `plan.md` 마일스톤(Wave 5 작성)과 §6 Acceptance Criteria 시나리오로 역참조된다.
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-OUT-001:REQ-OUT-001-<NNN>` 주석 부착.
- 총 REQ 개수: 22개 (Ubiquitous 9, Event-Driven 4, State-Driven 3, Optional 2, Complex 4).
- 예상 코드 구현 경로:
  - `internal/output/renderer.go` (REQ-OUT-001-001, 014, 015)
  - `internal/output/diff.go` (REQ-OUT-001-005, 013)
  - `internal/output/errors.go` (REQ-OUT-001-006, 011)
  - `internal/output/progress.go` (REQ-OUT-001-010, 016, 021)
  - `internal/output/status_icon.go` (REQ-OUT-001-008)
  - `internal/output/file_link.go` (REQ-OUT-001-009, 018)
  - `internal/output/style_loader.go` (REQ-OUT-001-002, 003, 004, 017, 019, 020, 022)
  - `internal/cli/doctor_output_style.go` (REQ-OUT-001-011)
  - `.claude/output-styles/moai-default.md` + `internal/template/templates/.claude/output-styles/moai-default.md`
  - `.claude/output-styles/moai-default_test.go` (frontmatter validation integration test)
- Gap matrix 추적: gm#165-#169 (4-source merge, built-ins, frontmatter parsing, keep-coding-instructions, force-for-plugin), gm#175-#182 (diff, errors, progress, code fences, icons, links, progress bar, schema compat).
- v3-master §3.7 Theme 7 및 §8.5 SPEC-V3-OUT-001 (single SPEC spans UI/UX theme).

---

End of SPEC.
