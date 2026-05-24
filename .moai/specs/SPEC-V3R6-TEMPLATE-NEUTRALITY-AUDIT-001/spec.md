---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — dev incident/path/refs sanitization for 16-language template distribution"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, neutrality, distribution, ci-guard"
tier: L
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Template Neutrality Audit

## §1 Goal

`internal/template/templates/` 하위 138개 unique file에 잔존하는 **개발자-local 또는 dev-incident 흔적**(macOS-biased absolute paths / V3R[0-9] dev version refs / 2026-05-XX dates / feedback_·memory refs / CLAUDE.local.md refs / PR #N / commit hash)을 **범용적(template-grade)** 컨텐츠로 정정한다. 16-language 사용자 프로젝트에 배포되는 모든 template 파일이 다음 4개 neutrality 기준을 만족해야 한다:

1. **No absolute path bias** — `/Users/...`, `/home/...` 등 OS-specific 경로 placeholder 금지 → `$HOME/...` 또는 `~/...` 통일
2. **No dev-incident traceability leakage** — V3R[0-9] dev version sigils, specific commit hashes, specific PR #N references → allow-list 정의 후 잔존은 doctrine 인용 경우만 허용
3. **No maintainer-personal information** — author 필드 `GOOS Kim` → `Author Name` (이미 fc47f31a7 Critical fix 완료, 본 SPEC은 재발 방지 CI guard 추가)
4. **No local-only file refs** — `CLAUDE.local.md` 참조는 template 배포 부적합 (사용자 프로젝트에는 존재하지 않음)

**False positive** identification: Go `GOOS=windows`, `GOOS=linux`, `GOOS=darwin` cross-compile environment variable은 보존(audit script exclude).

## §2 Non-Goals

본 SPEC은 다음을 **수행하지 않는다**:

### §2.1 Out of Scope — Template neutrality audit boundary

- Sprint 1 in-progress SPEC directories (`SPEC-V3R6-AGENT-MODEL-ROUTING-001/`, `SPEC-V3R6-HOOK-ASYNC-EXPAND-001/`, `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/`, `SPEC-V3R6-PROMPT-CACHE-001/`) 직접 수정 — 각 SPEC 개별 plan/run 흐름에서 처리
- `docs-site/content/{en,ko,ja,zh}/book/` untracked 잔존 디렉토리 정리 — 별도 docs-site cleanup SPEC
- `internal/hook/.moai/` working dir leak 정리 — git-strategy cleanup SPEC
- `.moai/specs/.moai/` symlink/dup 정리 — workspace hygiene SPEC
- `settings.local.json`의 maintainer-specific 값 (e.g., `teammateMode`, `defaultMode`) — runtime-managed, §2 [HARD] settings.local.json Separation
- `EXCL-CCE-001` 패턴의 user-facing string literal (Korean nodejs install URLs, GitLab descriptions, BODP literal `(권장)`) — translation 불필요
- Code 파일 (`internal/*.go`, `pkg/*.go`)의 dev-incident 주석 — 본 SPEC scope는 `internal/template/templates/` 한정 (Go code dev-incident 정리는 SPEC-V3R6-CODE-COMMENTS-EN-001 또는 별도 SPEC)
- Template Go template variable rendering 정정 (`{{.GoBinPath}}` 등 `internal/template/render_*.go`의 Go template logic은 무관)

## §3 Background

### 3.1 발견 경위

2026-05-23 commit `fc47f31a7`에서 Critical 4 violations을 즉시 fix하였음:
- `internal/template/templates/.claude/skills/moai/SKILL.md` 2 lines: `/Users/goos/MoAI/moai-adk-go/...` → `${CLAUDE_SKILL_DIR}/...`
- `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` 2 lines: `author: GOOS Kim` → `author: Author Name`
- 신규 `internal/template/templates/.claude/rules/moai/development/sprint-round-naming.md` SSOT (V3R6 / PR #819 refs 제거된 generic version)

그러나 본 세션 추가 audit 결과:
- **C1 macOS-biased path placeholder NEW 4 files / 8 lines** (`worktree-integration.md`, `run/context-loading.md`, `moai-foundation-cc/examples.md`, `moai-workflow-loop/examples.md`) — 모두 `/Users/user/...` 또는 `/Users/john/...` generic placeholder이나 macOS-only path syntax
- **C2 V3R[0-9] refs 70 files** — 일부는 정당한 doctrine 인용 (`decisions/lsp-client-choice.md` V3R5 결정 기록), 다수는 incident traceability 노출
- **C3 2026-05-XX dates 32 files** — incident date allow-list 외 generalize
- **C4 feedback_/memory refs 9 files** — rule/doctrine 인용 정상, 외 generalize
- **C5 CLAUDE.local.md refs 10 files** — local-only file ref는 template에 부적합
- **C6 PR #N refs 3 files** — 특정 PR 번호는 template 부적합
- **C7 commit hash refs 2 files** — 특정 hash는 template 부적합
- **C8 False positive** `GOOS=(linux|windows|darwin)` Go env var 4 hits / 3 files — PRESERVE

총 **138 unique files** (Critical NEW 4 + Medium deferred 134). Tier L mass-migration SPEC scope.

### 3.2 Template-First Rule 부작용

`.claude/rules/moai/development/template-first-rule.md`는 모든 신규 rule이 `internal/template/templates/` 동시 미러를 의무화한다. 그 결과 **dev-specific incident rule** (예: V3R6 round-splitting lessons, sprint-round naming PR #819 history)이 user-facing template에 자동 배포될 위험. 본 SPEC은:
- Audit script (`internal/template/template_neutrality_audit_test.go`)로 재발 차단
- CI workflow (`template-neutrality-check.yaml`)으로 PR-level enforcement
- Template-First Rule guideline 보강 — acceptable content range 명문화

### 3.3 분류 매트릭스 필요성

70 V3R refs / 34 dates / 15 memory refs는 일괄 removal하면 안 됨. 각 카테고리에 **case-by-case classify**가 필요:
- **PRESERVE (allow-list)**: rule SSOT 인용, decision rationale, doctrine citation (예: "Per session-handoff.md, the 5 triggers are ..." — V3R5 ID로 등록된 HARD clause는 보존)
- **GENERALIZE**: incident-specific 표현을 패턴 표현으로 변경 (예: "2026-04-25 incident" → "the 2026 stream-stall incident" 또는 "a prior stream-stall incident")
- **REMOVE**: 단순 dev-history 흔적 (예: "fc47f31a7로 fix됨" → 제거)

이 매트릭스는 `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md`에 single source로 기록되며, M2~M5에서 manager-develop이 참조한다.

## §4 EARS GEARS Requirements

다음 13개 REQ는 GEARS notation (Where / When / While / If-Then) self-dogfooding 원칙을 따른다. EARS-only "The system shall …" 패턴 사용 시에도 GEARS-compatible context 선언과 결합한다.

### REQ-TNA-001 — macOS-biased path placeholder removal (C1)

**Where** the template tree (`internal/template/templates/**`) contains an absolute path placeholder beginning with `/Users/`, `/home/`, or any OS-specific user home prefix, **the system shall** replace the prefix with `$HOME` or `~` while preserving the remaining path segments, file role context, and surrounding example narrative.

Affected baseline (verified 2026-05-23): 4 files / 8 lines
- `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` — lines 260, 261, 270, 273
- `internal/template/templates/.claude/skills/moai/workflows/run/context-loading.md` — lines 177, 185, 188
- `internal/template/templates/.claude/skills/moai-foundation-cc/references/examples.md` — line 646
- `internal/template/templates/.claude/skills/moai-workflow-loop/references/examples.md` — line 396

### REQ-TNA-002 — V3R[0-9] dev version refs classification (C2)

**Where** the template tree contains a substring matching the regex `V3R[0-9]`, **the system shall** classify each occurrence per the migration matrix into PRESERVE (rule SSOT citation / decision record), GENERALIZE (incident pattern), or REMOVE (dev-history trace). **If** the occurrence is classified as PRESERVE, **then** the file shall be added to the allow-list defined in `migration-matrix.md` §C2 with explicit rationale; otherwise the occurrence shall be transformed per the classified action.

Baseline (verified 2026-05-23): 70 files containing `V3R[0-9]` substring. Examples include `.moai/decisions/lsp-client-choice.md` (V3R5 decision record — PRESERVE), `.moai/config/sections/git-strategy.yaml.tmpl` (V3R[0-9] mention — case-by-case), etc.

### REQ-TNA-003 — Date refs classification (C3)

**Where** the template tree contains an ISO date matching `2026-0[5-9]-[0-9]{2}`, **the system shall** classify each occurrence per the migration matrix. **If** the date is tied to an incident referenced in active doctrine (e.g., `feedback_large_spec_wave_split.md` cross-reference), **then** the file shall be added to the C3 allow-list; **otherwise** the date shall be generalized to month-year granularity or removed.

Baseline (verified 2026-05-23): 32 files. Allow-list candidates: rule files referencing canonical incident dates (e.g., the 2026-04-25 stream-stall incident, the 2026-05-09 model-specific threshold revision).

### REQ-TNA-004 — feedback_/memory refs classification (C4)

**Where** the template tree contains a substring matching `feedback_` or `memory\.md`, **the system shall** classify each occurrence. **If** the reference cites a rule SSOT or doctrine pattern (e.g., "Lessons Protocol writes to memory `lessons.md`"), **then** PRESERVE; **otherwise** the reference shall be removed or replaced with a generic "auto-memory" phrasing.

Baseline (verified 2026-05-23): 9 files.

### REQ-TNA-005 — CLAUDE.local.md refs handling (C5)

**Where** the template tree references `CLAUDE.local.md` directly, **the system shall** remove the reference because `CLAUDE.local.md` is documented as a maintainer-only local file (per `CLAUDE.local.md` self-description). **If** the surrounding context discusses local-only configuration in the abstract, **then** the reference shall be replaced with a generic statement (e.g., "local override file" or "machine-specific configuration").

Baseline (verified 2026-05-23): 10 files.

### REQ-TNA-006 — PR #N refs removal (C6)

**Where** the template tree contains a substring matching `PR #[0-9]+`, **the system shall** remove the specific PR number. **If** the surrounding context discusses a rule precedent that requires historical evidence, **then** the PR ref shall be replaced with a generic phrase (e.g., "a prior round of the same rule" or "the originating PR" without a specific number).

Baseline (verified 2026-05-23): 3 files. Examples include `manager-develop-prompt-template.md`, `sync/delivery.md`, `scripts/ci-mirror/cross-compile.sh`.

### REQ-TNA-007 — Commit hash refs removal (C7)

**Where** the template tree contains a substring matching `[a-f0-9]{7,40}` that resolves to a specific commit hash (verified via git context, not coincidental hex strings such as color codes or SHA test fixtures), **the system shall** remove the hash. **If** the surrounding context cites a doctrinal decision rooted in a commit, **then** the hash shall be replaced with a doctrine name or rule citation.

Baseline (verified 2026-05-23): ~2 files (post-deduplication of false-positive hex matches in color codes / test fixtures).

### REQ-TNA-008 — False positive exclusion (`GOOS=...` Go env var) (C8)

**Where** the template tree contains a substring matching `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` in the context of Go cross-compilation (e.g., shell scripts, Go build commands), **the system shall** preserve the occurrence unchanged. The audit script (REQ-TNA-009) shall exclude this pattern from violation reports.

Baseline (verified 2026-05-23): 4 hits / 3 files
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md`
- `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md`
- `internal/template/templates/scripts/ci-mirror/cross-compile.sh`

### REQ-TNA-009 — Audit script (`template_neutrality_audit_test.go`)

**Where** `internal/template/` is the Go test discovery root, **the system shall** provide a new test file `internal/template/template_neutrality_audit_test.go` containing the function `TestTemplateNeutralityAudit`. **When** invoked via `go test ./internal/template/...`, the test shall scan `internal/template/templates/**` and emit a single FAIL with a structured violation report when any of categories C1, C5, C6, C7 (binary patterns) finds a hit outside its allow-list, and a WARN log line (not a FAIL) when C2, C3, C4 (case-by-case patterns) find a hit outside the migration-matrix allow-list. **If** the C8 false-positive pattern (`GOOS=(linux|windows|darwin)`) matches, **then** the hit shall be excluded from all reports.

### REQ-TNA-010 — CI guard workflow (PR touching template/)

**Where** a GitHub PR modifies any file under `internal/template/templates/**`, **the system shall** trigger the `template-neutrality-check.yaml` workflow which runs `TestTemplateNeutralityAudit` and surfaces failures as a required status check. **If** the workflow detects a C1/C5/C6/C7 violation, **then** the PR check shall fail; **if** only C2/C3/C4 WARN-level findings are emitted, **then** the check shall pass with annotations.

### REQ-TNA-011 — Migration matrix (`migration-matrix.md`)

**Where** the SPEC directory is `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/`, **the system shall** include a file `migration-matrix.md` containing one section per category (C1–C8) with: (a) the regex / detection rule, (b) the action policy (PRESERVE / GENERALIZE / REMOVE), (c) the allow-list (file paths + rationale), (d) the baseline count, and (e) the post-fix expected count. **If** a category has zero allow-list entries, **then** the allow-list section shall explicitly state "Empty (no exceptions)".

### REQ-TNA-012 — Template-First Rule guideline reinforcement

**Where** the canonical Template-First Rule is documented (currently in `CLAUDE.local.md` §2 and any equivalent rule file), **the system shall** add an `Acceptable Content Range for Templates` subsection enumerating: (a) what is acceptable in template content (rule SSOT citations, doctrine name references, GENERIC examples), (b) what is rejected (specific PR / commit / date / V3R refs, absolute paths, maintainer-personal names, local-only file refs). **If** a future rule introduces template-bound content, **then** the rule author shall verify against this guideline before committing.

### REQ-TNA-013 — Documentation for future template additions

**Where** a contributor authors a new file destined for `internal/template/templates/`, **the system shall** provide a discoverable checklist (referenced from `CLAUDE.local.md` and from REQ-TNA-012's subsection) that the contributor follows. **If** the checklist verification fails (e.g., the new file references a specific PR #), **then** the contributor shall remediate before opening the PR; the CI guard (REQ-TNA-010) shall act as the safety net catching missed manual verifications.

## §5 Acceptance Criteria reference

See `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/acceptance.md` for the 13 binary AC scenarios (AC-TNA-001 through AC-TNA-013) and the Given-When-Then test scenarios.
