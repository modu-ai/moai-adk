---
id: SPEC-V3R6-CODE-COMMENTS-EN-001
title: "Mass migration of Korean comments to English in Go source code"
version: "0.2.0"
status: completed
created: 2026-05-23
updated: 2026-05-30
author: manager-spec
priority: Medium
phase: "v3.0.0"
module: "internal/"
lifecycle: spec-anchored
tags: "code-quality, comments, internationalization, en-migration, mass-migration"
tier: L
issue_number: null
depends_on: []
related_specs: []
---

# SPEC-V3R6-CODE-COMMENTS-EN-001 — Go 소스 코드 한국어 주석 영어 마이그레이션

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-23 | manager-spec | 초안 작성 — Tier L mass migration SPEC, 7-wave 분할, AskUserQuestion 4-결정 반영 (test files 포함 / 문자열 리터럴 보존 / Agent-based 실행 / @MX 태그 영어화) |
| 0.1.1 | 2026-05-23 | manager-spec | Cobra exception (OQ-CCE-001 사용자 결정 2026-05-22 Option B) 본문 반영 — EXCL-CCE-001 예외 명시 (8 entries placeholder), AC-CCE-010 default 변경 검토 |
| 0.2.0 | 2026-05-23 | manager-spec | plan-auditor iter 1 REVISE 7-fix 반영: (1) Frontmatter B4 canonical schema (12-field, `created`/`updated`/`tags`, `lifecycle: spec-anchored` 추가, `labels` 제거) — 5 artifacts 모두 적용; (2) AC-CCE-004 Cobra exception 반영 (Option α — PRE_COUNT − 14 == POST_COUNT, baseline 2186 cached); (3) EXCL-CCE-001 exception count 8 → **14 정정** (verified inventory: mx.go 2 + glm_tools.go 4 + migration.go 8); (4) REQ-CCE-005 + REQ-CCE-008 carve-out 추가 (Cobra 예외 명시); (5) Wave 4 재정의 — pkg/ + ALL remaining internal/* non-test (39 files, 9+25+23+39=96 reconcile); Wave 7 expansion 명시 (~83 files, 50+38+83=171 reconcile); (6) AC-CCE-002 multi-line block comment perl `-0777` slurp scanner 채택 (Edge Case 1 RESOLVED); (7) AC-CCE-004 stash-based pre-count 폐기, research.md §2.2 cached baseline 2186 reference 채택. Cross-reference: research.md §2.2 + §6.1 expanded enumeration. |

---

## 1. Purpose

오픈소스 프로젝트 `moai-adk-go`의 Go 소스 코드 (`internal/`, `cmd/`, `pkg/`)에 누적된 **한국어 주석을 영어로 일괄 마이그레이션**하여 `.moai/config/sections/language.yaml`의 `code_comments: en` 정책 (CLAUDE.local.md §3 "All code, comments, godoc in English")을 강제한다.

본 SPEC은 **comments + godoc + @MX tag descriptions** 영어화에 한정하며, **문자열 리터럴 (`"..."`, `` `...` `` 내부 한국어) 및 SPEC-ID/REQ-ID 식별자는 보존**한다.

### 1.1 배경

- 본 프로젝트는 OSS 공개 저장소 — 영어권 사용자가 코드를 읽음
- 정책 (`.moai/config/sections/language.yaml`): `code_comments: en` 명시
- 현재 상태 (2026-05-23 인벤토리):
  - **96 non-test files** / **171 test files** = **267 files**
  - **4,246 comment lines** 한국어 포함 (`//.*[가-힣]`)
  - **2,186 string literal lines** 한국어 포함 (`".*[가-힣].*"`) — **본 SPEC 범위 외 (보존)**
  - **139 @MX tag** 한국어 설명 (`@MX:NOTE/WARN/REASON/ANCHOR/TODO`)

### 1.2 범위 (User-confirmed via AskUserQuestion 2026-05-23)

| 결정 항목 | 확정 사항 |
|----------|----------|
| Test files | **INCLUDED** — 171 `*_test.go` files 모두 in-scope |
| String literals | **PRESERVED** — `"..."` / `` `...` `` 내부 한국어 보존 (user-facing 메시지 가능성) |
| Execution method | **Agent-based batch translation** — `manager-develop` per-file Read → Translate → Edit cycle (NOT sed/script) |
| @MX tag descriptions | **ENGLISHIFIED** — `@MX:NOTE/WARN/REASON/ANCHOR/TODO` 한국어 설명 → 영어 |

---

## 2. Stakeholders & Audience

- **Primary**: OSS contributors / external code reviewers (영어권)
- **Secondary**: MoAI-ADK 메인테이너 (정책 강제 + 일관성)
- **Tertiary**: Agent runtime (코드 컨텍스트 영어 일관성 → 추론 효율)

---

## 3. EARS Requirements

### 3.1 Ubiquitous Requirements (항상 활성)

**REQ-CCE-001** (Ubiquitous): The system shall ensure all `//` line comments and `/* */` block comments in Go source files under `internal/`, `cmd/`, `pkg/` (including `*_test.go`) are written in English.

**REQ-CCE-002** (Ubiquitous): The system shall ensure all `@MX:NOTE`, `@MX:WARN`, `@MX:REASON`, `@MX:ANCHOR`, `@MX:TODO` tag descriptions are written in English regardless of file type (test or non-test).

**REQ-CCE-003** (Ubiquitous): The system shall ensure all godoc comments (immediately preceding exported types, functions, variables, constants) are written in English.

**REQ-CCE-008** (Ubiquitous): The migration shall not change any non-comment Go syntax — function signatures, types, imports, control flow, and string literal contents remain byte-identical, **with the documented N=14 Cobra exception (REQ-CCE-005 carve-out) as the sole permitted divergence**.

### 3.2 Event-Driven Requirements

**REQ-CCE-004** (Event-Driven): WHEN a comment references a SPEC-ID (e.g., `SPEC-V3R6-XXX-001`), REQ-ID (e.g., `REQ-CCE-001`), or AC-ID (e.g., `AC-CCE-001`), the migration shall preserve the identifier verbatim without translation.

### 3.3 State-Driven Requirements

**REQ-CCE-005** (State-Driven): WHILE migrating a file, string literal contents (text inside `"..."` and `` `...` ``) shall remain byte-identical to the pre-migration state, **except for the N=14 Cobra command `Use:/Short:/Long:/Example:` fields enumerated in EXCL-CCE-001 exception (research.md §6.1)** which are Englishified per OQ-CCE-001 user decision (2026-05-22) and verified by AC-CCE-010.

### 3.4 Optional Requirements

**REQ-CCE-006** (Optional): WHERE comments contain technical terms (Go keywords like `goroutine`, `defer`, `select`; library names like `cobra`, `gh CLI`; error codes like `EPERM`, `ENOENT`), the migration may preserve such terms verbatim within otherwise-English sentences.

### 3.5 Unwanted Requirements

**REQ-CCE-007** (Unwanted): IF a comment is ambiguous between Korean particles and code identifiers (e.g., 변수명에 한국어 조사 결합), the migration shall prefer clarity over literal translation and record the judgment basis in the commit message body.

---

## 4. Acceptance Summary

| AC ID | Description | Verification |
|-------|-------------|--------------|
| AC-CCE-001 | Go 주석 한국어 grep 0 matches | `grep -rn '//.*[가-힣]' internal/ cmd/ pkg/ --include="*.go"` |
| AC-CCE-002 | Block comment 한국어 0 matches | `grep -rn '/\*.*[가-힣].*\*/' internal/ cmd/ pkg/ --include="*.go"` |
| AC-CCE-003 | @MX tag 한국어 0 matches | `grep -rn '@MX:[A-Z]*[: ].*[가-힣]' internal/ cmd/ pkg/ --include="*.go"` |
| AC-CCE-004 | String literal 한국어 count 보존 (~2,186 lines) | before/after stash diff |
| AC-CCE-005 | `go build ./...` exit 0 | cross-package build |
| AC-CCE-006 | Cross-platform build PASS | `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| AC-CCE-007 | `go test ./...` baseline 보존 | NEW failures = 0 (baseline residual 문서화) |
| AC-CCE-008 | golangci-lint NEW issues = 0 | `golangci-lint run --timeout=2m` |
| AC-CCE-009 | Sample manual review (5 files) semantic 보존 | random file sampling + reviewer eval |
| AC-CCE-010 | Cobra command 설명 영어 | `Use:/Short:/Long:` 필드 영어 |
| AC-CCE-011 | SPEC-ID/REQ-ID 식별자 verbatim 보존 | `grep -rn '\(SPEC\|REQ\|AC\)-[A-Z0-9-]\+' --include="*.go"` 카운트 동일 |
| AC-CCE-012 | git diff scope 제한 (`*.go` only) | `git diff --stat` 결과 `.go` 확장자만 |

상세 검증 절차는 [acceptance.md](./acceptance.md) 참조.

---

## 5. Out of Scope

### 5.1 Out of Scope

**EXCL-CCE-001**: 문자열 리터럴 (`"..."` 및 `` `...` ``) 내부의 한국어는 보존한다. 사용자 노출 메시지일 가능성 (CLI output, error string, log message) 존재.

> **EXCL-CCE-001 예외 (OQ-CCE-001 사용자 결정 2026-05-22)**: Cobra command 정의의 `Use:`, `Short:`, `Long:`, `Example:` 필드는 string literal이지만 **영어화 대상**. 이유: `moai <cmd> --help` 출력은 오픈소스 CLI 사용자가 보는 documentation surface이며, `code_comments: en` 정책과 일관성 유지. AC-CCE-010이 본 예외를 검증한다. **14개 항목 (verified 2026-05-23: internal/cli/mx.go 2개 (lines 12,13), internal/cli/migration.go 8개 (lines 17,18,36,37,67,68,122,123), internal/cli/glm_tools.go 4개 (lines 67,68,87,95))**가 영향 받음. 상세 enumeration은 research.md §6.1 참조.

**EXCL-CCE-002**: `.moai/specs/`, `.moai/research/`, `.moai/design/`, `.moai/docs/` 등 `.moai/` 내 Markdown 한국어 문서는 보존한다 (`documentation: ko` 정책).

**EXCL-CCE-003**: `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `CHANGELOG.md` 등 root-level multi-locale README는 보존한다.

**EXCL-CCE-004**: `docs-site/content/{ko,en,ja,zh}/**` 4-locale Hugo docs는 보존한다.

**EXCL-CCE-005**: `.claude/agents/`, `.claude/skills/`, `.claude/rules/`, `.claude/commands/`, `.claude/output-styles/` 디렉토리는 이미 영어 정책 적용 — 본 SPEC에서 다루지 않는다.

**EXCL-CCE-006**: `internal/template/templates/.claude/` mirror도 이미 영어 — 본 SPEC 범위 외.

**EXCL-CCE-007**: 테스트 파일의 `t.Errorf` / `t.Fatalf` / `t.Logf` Korean message strings — string literal로 EXCL-CCE-001에 포섭됨.

**EXCL-CCE-008**: 사전 존재 baseline test failures (HARNESS-RENAME-001 cascade, WorktreeCreate hook V3R6 unwire 등 잔존)는 본 SPEC에서 introduced 되지 않음 — pre-existing baseline으로 문서화한다.

**EXCL-CCE-009**: SPEC-ID, REQ-ID, AC-ID, MEMO-ID 등 식별자 토큰은 verbatim 보존 (REQ-CCE-004).

**EXCL-CCE-010**: 한국어 인용 (user input quotation, error string variable interpolation) — 코드 컨텍스트 내 한국어 변수/문자열 참조는 보존.

---

## 6. Constraints

### 6.1 Technical Constraints

- **C-CCE-001**: Translation 결과는 **semantic preservation** 우선 — 직역보다 의미 보존. Agent (LLM) 기반 batch translation 의무.
- **C-CCE-002**: SPEC-ID/REQ-ID/AC-ID identifier patterns (`SPEC-[A-Z0-9-]+`, `REQ-[A-Z]+-\d+`, `AC-[A-Z]+-\d+`) verbatim 보존.
- **C-CCE-003**: Cross-platform build 영향 0 (Windows/Linux/macOS) — REQ-CCE-008 byte-identity 보장.
- **C-CCE-004**: @MX:ANCHOR fan_in invariant 보존 — `@MX:REASON fan_in >= N` 같은 contract 표현은 영어로 동등 표현.

### 6.2 Operational Constraints

- **C-CCE-005**: Working tree hygiene — `.moai/harness/usage-log.jsonl`, `.moai/state/`, 기타 runtime-managed files 변경 금지.
- **C-CCE-006**: 7-wave 분할 — 각 Wave별 independent PR 권장 (large diff 회피, review 용이성). [plan.md](./plan.md) §3 참조.
- **C-CCE-007**: Template-First Rule — `internal/template/templates/.claude/` mirror는 이미 영어 정책이므로 별도 작업 불필요. 본 SPEC은 source code only.

### 6.3 Quality Gate Constraints

- **C-CCE-008**: 본 SPEC 본문 작성 시 Korean — `.moai/config/sections/language.yaml` `documentation: ko` 정책 준수. 영어화 대상은 **Go source code**에 한정.
- **C-CCE-009**: Section A-E delegation template (`.claude/rules/moai/development/manager-develop-prompt-template.md`) Tier L MANDATORY 적용.

---

## 7. Risks and Mitigation

| Risk | Severity | Mitigation |
|------|---------|------------|
| R-CCE-001: Semantic drift (의미 손실) | High | Agent-based translation + sample manual review (AC-CCE-009) |
| R-CCE-002: Identifier 오번역 (SPEC-ID 등) | Critical | REQ-CCE-004 verbatim 규칙 + AC-CCE-011 count 검증 |
| R-CCE-003: String literal 우발 변경 | High | REQ-CCE-005 byte-identity + AC-CCE-004 before/after diff |
| R-CCE-004: 코드 byte-corruption (탭/스페이스 변경) | Medium | Edit tool 사용 (sed/awk 금지) + AC-CCE-005 build pass |
| R-CCE-005: Wave 간 conflict (parallel work) | Low | 7-wave 직렬 진행 권장, 각 Wave PR merged 후 다음 Wave 진입 |
| R-CCE-006: 대량 diff로 review 부담 | Medium | 7-wave 분할로 PR 크기 제한 (~30-40 files/PR) |
| R-CCE-007: Test failure cascade | Medium | EXCL-CCE-008 baseline 문서화 + stash-test before/after 비교 |

---

## 8. Dependencies

- **No upstream SPEC blockers** (independent migration)
- **Concurrent baseline**: HARNESS-RENAME-001 cascade (3 pre-existing FAIL) — EXCL-CCE-008로 포섭
- **Recommended sequence**: Wave 1 (foundation) → Wave 2-4 (non-test) → Wave 5-7 (test files)

---

## 9. References

- `.moai/config/sections/language.yaml` — `code_comments: en` 정책
- `CLAUDE.local.md` §3 — "All code, comments, godoc in English"
- `CLAUDE.md` §5 — MX Tag Integration (`@MX:` tag types)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — MX tag canonical
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier L delegation template
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Frontmatter canonical
- Project memory: `lessons #B4 Frontmatter`, `lessons #B6 spec-lint heading`, `lessons #B8 working tree hygiene`

---

## 10. Glossary

- **Comment**: `//` line comment 또는 `/* */` block comment (string literal 외부)
- **Godoc**: 노출 식별자 (exported type/func/var/const) 바로 위 line comment (Go doc tool 인식)
- **@MX tag**: MoAI MX annotation 시스템 — `@MX:NOTE/WARN/REASON/ANCHOR/TODO` (`.claude/rules/moai/workflow/mx-tag-protocol.md`)
- **String literal**: `"..."` interpreted 또는 `` `...` `` raw string (Go syntax, byte-identical 보존 대상)
- **Identifier token**: SPEC-ID (`SPEC-V3R6-XXX-001`), REQ-ID (`REQ-CCE-001`), AC-ID (`AC-CCE-001`), MEMO-ID, error code (verbatim 보존 대상)
- **Wave**: 본 SPEC plan.md에서 정의된 7-stage 분할 단위 (각 stage = independent PR 권장)

---

Version: 0.2.0
Status: draft
Tier: L
Total scope: 267 Go files (96 non-test + 171 test), 4,246 comment lines, 139 @MX tags
Wave reconciliation: 9 + 25 + 23 + 39 = 96 non-test ; 50 + 38 + 83 = 171 test ; sum = 267
String literal Korean baseline: 2,186 (Cobra exception delta: -14 → expected post: 2,172)
