---
id: SPEC-V3-SKL-001
title: "Skill Frontmatter Enhancements (context, paths, skills array)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 3 Agent Runtime v2"
module: "internal/config/schema/, internal/template/deployer.go"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - gm#87
  - gm#88
  - gm#89
  - gm#90
  - gm#99
  - gm#100
  - gm#101
  - gm#104
  - gm#105
related_theme: "Theme 4 — Agent Frontmatter Expansion (skill parallel)"
breaking: false
lifecycle: spec-anchored
tags: "skill, frontmatter, paths, context-fork, effort, v3"
---

# SPEC-V3-SKL-001: Skill Frontmatter Enhancements (context, paths, skills array)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial draft (Wave 4 SPEC writer)   |

---

## 1. Goal (목적)

moai skill frontmatter를 Claude Code의 v2 skill schema와 동등하게 확장한다. 현재 moai skill은 `name`, `description`, `allowed-tools`만 널리 사용되며(findings-wave1-moai-current.md §7.3), `paths:`는 coding-standards.md에 문서화되어 있으나 실제로는 활용되지 않는다. CC는 다음 핵심 필드를 지원한다(findings-wave1-hooks-commands.md §14.5 및 gap matrix):

- `context: inline | fork` — inline은 현재 호출자 context에 전개; fork는 sub-agent로 격리 실행 (gm#87)
- `paths:` — glob 기반 조건부 활성화 (gm#89; 문서화되어 있으나 미사용)
- `skills:` — skill 간 의존성 / 선행 로드 (기존 CSV string 형태에서 YAML array로 통일)
- `effort:` — Skill별 Opus 4.7 effort override (gm#90)
- `${CLAUDE_SKILL_DIR}`, `${CLAUDE_SESSION_ID}` 본문 치환 (gm#104, gm#105)
- `$ARGUMENTS[N]`, `$N`, `$name` args 치환 (gm#99, gm#100, gm#101)

본 SPEC은 v3.0에서 `paths`, `skills (array)`, `effort`, 그리고 본문/인자 치환 4종(`${CLAUDE_SKILL_DIR}`, `${CLAUDE_SESSION_ID}`, `$ARGUMENTS[N]`, `$name`)만 구현한다. `context: fork`는 v3.0 범위에서는 **스키마 필드만 추가**하고 실제 fork 실행은 SPEC-V3-SKL-002에서 수행한다.

### 1.1 배경

moai는 50개 스킬을 템플릿으로 배포하며(findings-wave1-moai-current.md §7.2), Progressive Disclosure Level 2(CLAUDE.md §13)로 ~5K tokens/skill 본문을 관리한다. 현재 frontmatter는:

```yaml
---
name: moai-workflow-spec
description: SPEC Workflow Management
allowed-tools: Bash(go:*), Read, Glob, Grep, Edit, Write
---
```

CC의 frontmatter는 위에 더해 `paths:` (glob 조건부 load), `skills:` (의존 skill array), `effort:` (Opus 4.7 override), `context:` (inline vs fork 실행) 등을 지원. 특히 `paths`는 skill이 파일 위치에 따라 자동 활성화되도록 하여 token waste를 방지한다.

### 1.2 Non-Goals

- `context: fork` 의 실제 실행 로직 (SPEC-V3-SKL-002에서 처리)
- 동적 skill discovery (walk up nested `.claude/skills/`; gm#106) — SPEC-V3-SKL-002
- Realpath canonicalization dedup (gm#107) — SPEC-V3-SKL-002
- Brace expansion in paths (gm#98) — v3.1+
- `disableModelInvocation`, `user-invocable`, `hide-from-slash-command-tool` (gm#91-93) — v3.1+
- `version:` 필드 (gm#94) — v3.1+
- `isSensitive: true` (gm#95) — v3.1+
- `immediate: true` (gm#96) — v3.1+
- YAML special-char auto-quoting (gm#97) — v3.1+
- `!cmd`/`!block` 본문 shell 실행 (gm#103) — v3.1+

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/config/schema/skill.go` 신설: `SkillFrontmatter` struct + validator/v10 태그
- `internal/template/deployer.go`의 skill frontmatter 파서 확장: 신규 필드 파싱
- `internal/core/skill/path_matcher.go`: `paths:` glob 매칭 (gitignore-style)
- `internal/core/skill/dep_loader.go`: `skills:` 배열 의존성 선행 로드 (topological)
- `internal/core/skill/effort.go`: `effort:` 해석 및 Opus 4.7 request에 주입
- `internal/core/skill/body_substitution.go`: `${CLAUDE_SKILL_DIR}`, `${CLAUDE_SESSION_ID}` 치환
- `internal/core/skill/args_substitution.go`: `$ARGUMENTS`, `$ARGUMENTS[N]`, `$N`, `$name` 치환
- 50개 기존 스킬 frontmatter v3 호환 검증 (warning-only; non-blocking)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- `context: fork` 실제 실행 (SPEC-V3-SKL-002)
- 동적 skill discovery (SPEC-V3-SKL-002)
- Brace expansion (`src/*.{ts,tsx}`)
- 9개 minor frontmatter fields (disableModelInvocation, user-invocable, version, isSensitive, immediate, hide-from-slash-command-tool, keep-coding-instructions는 output-style 전용, force-for-plugin는 output-style 전용)
- 본문 inline shell execution (\`!cmd\`)
- 50개 스킬 일괄 frontmatter 업그레이드 (opt-in 방식; 사용자가 필요시 추가)
- Skill usage telemetry
- Skill deprecation markers

---

## 3. Environment (환경)

- 런타임: moai-adk-go v3.0.0+ (Go 1.23+)
- Claude Code v2.1.111+ (Opus 4.7 지원)
- 의존: SPEC-V3-SCH-001 (validator/v10)
- 영향 디렉터리: `internal/config/schema/`, `internal/core/skill/`, `internal/template/deployer.go`, `.claude/skills/`
- 현 skill 수: 50 (template) / 47 (local) — SPEC-V3-SKL-002에서 별도 드리프트 해소
- OS 동등성: macOS / Linux / Windows
- 참조: `findings-wave1-hooks-commands.md` §6.7, §7, §8, §11, §14.5

---

## 4. Assumptions (가정)

- SPEC-V3-SCH-001이 먼저 머지되어 validator/v10이 도입되어 있다.
- 50개 기존 스킬 frontmatter는 v3 스키마에서 0 error로 파싱된다(추가 필드는 optional).
- `paths` glob은 gitignore 스타일(`**`, `*`, `!`, leading slash)을 지원하는 널리 쓰이는 pattern이다.
- `skills:` 의존성은 non-cyclic이다 (cycle detection은 본 SPEC 범위).
- `${CLAUDE_SKILL_DIR}`는 skill의 절대 디렉터리 경로이며, Windows에서는 forward-slash로 정규화된다 (CC 동일 정책).
- `$ARGUMENTS[N]`은 0-indexed다 (CC 동일).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-SKL-001-001 (Ubiquitous) — schema struct**
The `internal/config/schema/skill.go` **shall** define `SkillFrontmatter` struct with at least: `Name`, `Description`, `AllowedTools`, `Paths`, `Skills`, `Effort`, `Context`, `Arguments`.

**REQ-SKL-001-002 (Ubiquitous) — validator tags**
The `SkillFrontmatter` struct **shall** enforce validator/v10 rules:
- `Context` ∈ `{"", "inline", "fork"}` (empty treated as `inline`)
- `Effort` ∈ `{"", "low", "medium", "high", "xhigh", "max"}`
- `Paths` each entry must be valid glob pattern
- `Skills` each entry must match `^[a-z0-9-]+$` slug

**REQ-SKL-001-003 (Ubiquitous) — paths CSV or YAML array**
The frontmatter parser **shall** accept `paths:` in YAML array form only (not CSV), in line with CC schema (findings §14.5).

**REQ-SKL-001-004 (Ubiquitous) — skills dependency order**
The skill loader **shall** load skills listed in `skills:` before loading the current skill's body (topological order).

### 5.2 Event-Driven (이벤트 기반)

**REQ-SKL-001-005 (Event-Driven) — paths 조건부 활성화**
**When** a file path is touched AND a skill declares `paths:` with at least one matching glob, the skill loader **shall** activate the skill for the current context.

**REQ-SKL-001-006 (Event-Driven) — effort override**
**When** a skill body is included in an Opus 4.7 agent's prompt AND the skill declares `effort:`, the agent runtime **shall** override the agent's default `effort` value with the skill's value for that turn.

**REQ-SKL-001-007 (Event-Driven) — body substitution**
**When** a skill body is rendered, the renderer **shall** substitute:
- `${CLAUDE_SKILL_DIR}` → skill's absolute dir path (forward-slash normalized on Windows)
- `${CLAUDE_SESSION_ID}` → current session id from env

**REQ-SKL-001-008 (Event-Driven) — args substitution**
**When** a skill body contains `$ARGUMENTS` / `$ARGUMENTS[N]` / `$N` / `$name` (where `name` appears in the skill's `arguments:` frontmatter declaration), the renderer **shall** substitute them with the corresponding positional or named argument value.

### 5.3 State-Driven (상태 기반)

**REQ-SKL-001-009 (State-Driven) — fork schema placeholder**
**While** a skill declares `context: fork` in v3.0, the loader **shall** accept the field as valid schema BUT **shall** execute the skill as `inline` and emit a one-time warning that fork execution requires SPEC-V3-SKL-002 to land.

**REQ-SKL-001-010 (State-Driven) — cycle detection**
**While** the dependency graph implied by `skills:` arrays contains a cycle, the loader **shall** reject the cycle with error `SKL_DEPENDENCY_CYCLE` and list the cycle path.

### 5.4 Optional (선택)

**REQ-SKL-001-011 (Optional) — inherit effort from agent**
**Where** a skill does not declare `effort:`, the agent's declared `effort` **shall** be used unchanged.

**REQ-SKL-001-012 (Optional) — progressive type hint**
**Where** a skill declares `arguments: "name1 name2"` and a user types `/skill-name partial`, the CLI **may** surface a typeahead hint for `name2`. This is an OPT-IN UX enhancement; not required for v3.0 correctness.

### 5.5 Unwanted Behavior

**REQ-SKL-001-013 (Unwanted Behavior) — invalid glob**
**If** a `paths:` entry contains an invalid glob pattern (e.g., unclosed bracket), **then** the parser **shall** reject the skill with error `SKL_INVALID_PATHS_GLOB` referencing the exact offending entry.

**REQ-SKL-001-014 (Unwanted Behavior) — unknown named arg**
**If** a skill body references `$name` where `name` is not declared in `arguments:`, **then** the renderer **shall** leave the token literal AND emit a warning to stderr.

**REQ-SKL-001-015 (Unwanted Behavior) — recursive skill dependency via fork skill**
**If** a skill with `context: fork` declares a dependency on another `context: fork` skill, **then** the loader **shall** reject the combination in v3.0 with error `SKL_NESTED_FORK_UNSUPPORTED` (fork-in-fork is a v3.1+ concern).

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-SKL-001-01**: 50개 기존 스킬 frontmatter 모두 v3 schema에서 0 error parse
- **AC-SKL-001-02**: `paths: ["**/*.go"]` 스킬이 Go 파일 터치 시 자동 활성화 (로그 검증)
- **AC-SKL-001-03**: skill A가 `skills: [B, C]` 선언 시 B, C 본문이 A 본문보다 먼저 로드됨
- **AC-SKL-001-04**: `effort: xhigh` 선언된 스킬이 agent 호출 시 xhigh 값 전달 확인
- **AC-SKL-001-05**: `${CLAUDE_SKILL_DIR}` 치환이 Windows에서 forward-slash 정규화됨
- **AC-SKL-001-06**: `$ARGUMENTS[0]` / `$0` 표기 모두 첫 번째 인자 값으로 치환
- **AC-SKL-001-07**: A→B→A 순환 의존성 → `SKL_DEPENDENCY_CYCLE` 반환, path A→B→A 로그
- **AC-SKL-001-08**: `paths: ["[broken"]` → `SKL_INVALID_PATHS_GLOB`
- **AC-SKL-001-09**: `context: fork` 선언 시 v3.0에서는 inline 실행 + warning (REQ-SKL-001-009)
- **AC-SKL-001-10**: `go test ./internal/core/skill/...` 전체 통과, coverage ≥ 85%

---

## 7. Constraints (제약)

- [HARD] 50개 기존 스킬 frontmatter는 변경하지 않는다 (opt-in via new fields; validation warning only).
- [HARD] `context: fork` 실 실행은 SPEC-V3-SKL-002로 분리 (v3.0 scope limit).
- [HARD] 하드코딩 금지: glob library, substitution pattern 등은 `internal/core/skill/constants.go`에 정의.
- [HARD] 신규 의존 금지: gitignore glob은 표준 라이브러리 또는 기존 `github.com/bmatcuk/doublestar` (해당 없으면 inline 구현).
- [HARD] 16개 언어 중립성: `paths:` glob은 특정 언어에 편중된 기본값 없음.
- [HARD] Windows path normalization: `${CLAUDE_SKILL_DIR}` 포함 모든 substitution은 forward-slash.
- [HARD] Progressive Disclosure Level 2 유지 (CLAUDE.md §13): SKILL.md ≤ 500 lines, 본문 ~5K tokens.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크                                                              | 영향   | 완화                                                                            |
|---------------------------------------------------------------------|--------|---------------------------------------------------------------------------------|
| `paths:` glob 라이브러리 선택이 CC와 semantic 차이                   | Medium | doublestar 사용(CC의 minimatch와 유사 semantic); 차이 케이스는 문서화           |
| Dependency cycle이 50개 스킬 중 잠재적으로 존재                      | Medium | REQ-SKL-001-010 cycle detection; CI에서 50개 스킬 그래프 검증                   |
| `effort:` override로 의도치 않은 token cost 증가                     | Medium | 기존 50개 스킬에서는 미사용; new skill이 선언 시 반드시 문서에 근거 명시        |
| `${CLAUDE_SKILL_DIR}` Windows path 오류                              | Low    | forward-slash 정규화 + Windows CI test (CLAUDE.local.md §13)                    |
| `$N` shorthand가 기존 $1 (bash substitution)과 혼동                  | Low    | skill body는 markdown; $1은 CC 치환 전용                                        |
| `context: fork` 선언된 스킬이 v3.0에서 inline 실행되어 기대치 불일치 | Low    | REQ-SKL-001-009 명시적 warning; 문서에 clear migration note                     |
| Topological sort의 large graph 성능                                  | Very Low | 50개 스킬 규모에서는 무시; O(V+E) 표준 Kahn's algorithm                          |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** (Formal config schemas): validator/v10 선행.

### 9.2 Blocks

- **SPEC-V3-SKL-002** (`context: fork` 실행 + 동적 discovery): 본 SPEC의 schema 필드 선언에 의존.

### 9.3 Related

- **SPEC-V3-AGT-001** (Agent frontmatter): agent의 `skills:` 배열과 skill의 `skills:` 배열은 별개이나 parsing 규칙 일관.
- **SPEC-HOOKS-005** (sibling writer — CLAUDE_ENV_FILE mechanism): `${CLAUDE_SKILL_DIR}` 치환과 CLAUDE_ENV_FILE은 별개 경로.

---

## 10. Traceability (추적성)

- 총 REQ 개수: 15 (Ubiquitous 4, Event-Driven 4, State-Driven 2, Optional 2, Unwanted Behavior 3)
- 예상 AC 개수: 10
- 관련 Wave 1 근거:
  - findings-wave1-hooks-commands.md §6.7 (paths conditional activation)
  - findings-wave1-hooks-commands.md §7 (frontmatterParser.ts YAML quoting)
  - findings-wave1-hooks-commands.md §8 (argument substitution)
  - findings-wave1-hooks-commands.md §9 (body substitution)
  - findings-wave1-hooks-commands.md §11 (context: inline vs fork)
  - findings-wave1-hooks-commands.md §14.5 (skill frontmatter extras)
  - findings-wave1-moai-current.md §7.3 (Current skill frontmatter fields)
  - master-v3 §3 Theme 4 (agent/skill parallel)
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-SKL-001:REQ-SKL-001-NNN` 주석 부착
- 코드 구현 예상 경로:
  - `internal/config/schema/skill.go` (REQ-SKL-001-001, 002, 003)
  - `internal/core/skill/path_matcher.go` (REQ-SKL-001-005, 013)
  - `internal/core/skill/dep_loader.go` (REQ-SKL-001-004, 010)
  - `internal/core/skill/effort.go` (REQ-SKL-001-006, 011)
  - `internal/core/skill/body_substitution.go` (REQ-SKL-001-007)
  - `internal/core/skill/args_substitution.go` (REQ-SKL-001-008, 014)
  - `internal/core/skill/fork_placeholder.go` (REQ-SKL-001-009, 015)
  - `internal/config/schema/skill_test.go` (AC-SKL-001-01, 08)
  - `internal/core/skill/*_test.go` (AC-SKL-001-02..09)

---

End of SPEC.
