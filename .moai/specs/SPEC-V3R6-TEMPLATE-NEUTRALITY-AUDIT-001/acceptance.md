---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Acceptance Criteria"
version: "0.1.2"
status: in-progress
created: 2026-05-23
updated: 2026-05-30
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, acceptance, ci-guard, verification"
tier: L
related_specs: [SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Acceptance Criteria

## §1 Binary Acceptance Criteria

11 active binary AC scenarios + 2 deferred markers. Each active AC has an explicit verification command and expected outcome. AC failure = SPEC NOT done. AC-TNA-003 (C3) and AC-TNA-007 (C7) are **DEFERRED** to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 per the v0.1.1 rescope (see spec.md §3.3 partition); their numbers are preserved for traceability contiguity and they emit no verification command here.

> **Pre-existing package RED (v0.1.2, out of scope — see spec.md §3.4)**: at run-phase baseline (HEAD `1162b0de8`) the `internal/template` Go package is already RED with 13 pre-existing failing test functions unrelated to this SPEC (notably `TestTemplateNoInternalContentLeak`, `TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestManagerDevelopActiveAgentPresent`/`TestManagerDevelopIsActiveAgent`, `TestEmbeddedTemplates_AgentDefinitions`, `TestAgentFrontmatterAudit`, `TestBuilderSkillPathStructure`, `TestTemplateAgentsStructure`, `TestSettingsTemplateHookEventCount`, `TestContractSchemaVerification`, `TestBackwardCompatibility`, `TestContractAssertionsNaturalLanguage`). **This SPEC does NOT fix them.** All AC commands that invoke `go test` (AC-TNA-008) MUST use the **isolated** `-run TestTemplateNeutralityAudit` form — a package-wide green (`go test ./internal/template/...`) is NOT a precondition and MUST NOT block this SPEC's closure. A separate cleanup SPEC (provisional `SPEC-V3R6-TEMPLATE-PACKAGE-RED-CLEANUP-001`, Tier M) is recommended to own the package-wide green. Mirror-drift caveat (F3): when editing the 4 byte-parity files (`manager-develop-prompt-template.md`, `manager-spec.md`, `spec-workflow.md`, `manager-git.md`), run-phase manager-develop MUST keep template ↔ `.claude/` mirror parity (edit both sides, or verify the `.claude/` side already matches the intended generic form).

### Traceability Matrix (AC → REQ)

| AC | Satisfies REQ | Verification mode | Status |
|----|---------------|-------------------|--------|
| AC-TNA-001 | REQ-TNA-001 | Binary grep | Active |
| AC-TNA-002 | REQ-TNA-002 | Awk allow-list count + perl negative-lookbehind grep (bare-narrative) | Active |
| AC-TNA-003 | REQ-TNA-003 | (none) | **DEFERRED → ISOLATION** |
| AC-TNA-004 | REQ-TNA-004 | Awk allow-list count + grep | Active |
| AC-TNA-005 | REQ-TNA-005 | Awk allow-list count + grep | Active |
| AC-TNA-006 | REQ-TNA-006 | Binary grep | Active |
| AC-TNA-007 | REQ-TNA-007 | (none) | **DEFERRED → ISOLATION** |
| AC-TNA-008 | REQ-TNA-009 | Go test cross-platform | Active |
| AC-TNA-009 | REQ-TNA-010 | Workflow file existence + YAML parse + run | Active |
| AC-TNA-010 | REQ-TNA-011 | Matrix section header count | Active |
| AC-TNA-011 | REQ-TNA-008 | False positive preservation | Active |
| AC-TNA-012 | REQ-TNA-012 | Guideline subsection grep | Active |
| AC-TNA-013 | REQ-TNA-013 + M2-M5 smoke | `moai init` clean run + checklist grep | Active |

### AC-TNA-001 — C1 macOS-bias absolute path removal

**Given** the template tree at `internal/template/templates/`
**When** the auditor executes `grep -rln '/Users/' internal/template/templates/`
**Then** the command shall output 0 lines.

**Verification command**:
```bash
test "$(grep -rln '/Users/' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')" = "0"
```

**Baseline (2026-05-23)**: 4 files / 8 lines (worktree-integration.md / context-loading.md / moai-foundation-cc-examples.md / moai-workflow-loop-examples.md)
**Post-fix expected**: 0

### AC-TNA-002 — C2 V3R[0-9] **bare-narrative** files ≤ allow-list count

**Given** the migration matrix at `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md` defines an allow-list of file paths permitted to contain **bare-narrative** `V3R[0-9]` sigils (NOT ID-embedded `SPEC-V3R…` / `CONST-V3R…` / `REQ-V3R…`)
**When** the auditor counts files matching the **bare-narrative** V3R regex (a `V3R[0-9]` NOT immediately preceded by `[A-Za-z0-9-]`)
**Then** the count shall be equal to or less than the allow-list size.

**Verification command** (bare-narrative detection via perl PCRE negative-lookbehind; allow-list count via the corrected non-self-terminating awk flag-based extractor per plan-audit iter-1 D2 fix):
```bash
actual=$(grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')
allowlist=$(awk '/^### C2 /{f=1;next} /^### C[0-9] /{f=0} f' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$actual" -le "$allowlist"
```

> **Bare-narrative narrowing (v0.1.2 M3 blocker resolution)**: the broad `grep -rln 'V3R[0-9]'` form matched 341 occurrences, of which 299 (88%) are ID-embedded substrings inside `SPEC-V3R…` / `CONST-V3R…` / `REQ-V3R…` identifiers — a domain owned by the sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001's `C1-spec-id` leak-test class. Counting those 299 would force this SPEC to sanitize SPEC-IDs (ISOLATION's job — forbidden cross-SPEC scope bleed), making the broad ≤18 target unachievable. The corrected verification uses `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'` (perl PCRE negative-lookbehind) which matches ONLY the 7 bare-narrative files. The Go audit script (REQ-TNA-009) implements the same exclusion as a two-pass approach (RE2 lacks lookbehind) — see spec.md REQ-TNA-002 § "Two-pass detection approach".

> **D2 fix rationale (preserved)**: the awk allow-list count uses a flag-based block extractor (set flag on the `### C2 ` header + skip via `next`, clear on the next `### C<digit> ` header, print while flag set), NOT the original self-terminating range expression `awk '/^### C2 /,/^### C[0-9]+/'` which extracted 0 bullets because the range-END pattern matched the range-START header on the same line. Verified 2026-05-30: returns the C2 file-path bullet count (the post-narrow allow-list, computable, non-zero).

**Baseline (orchestrator-measured 2026-05-30 at HEAD `1162b0de8`, bare-narrative grep)**: **7 bare-narrative files** (replaces the stale 70/73 broad-regex counts). The broad `V3R[0-9]`=341 hits; 299 are ID-embedded (ISOLATION-owned, NOT a C2 target). Point-in-time; run-phase M3 re-measures the bare-narrative set and reduces it to the allow-list before this AC passes.
**Post-fix expected**: ≤ allow-list size (6 PRESERVE: zone-registry namespace decision record + CONST section headers; manager-spec.md SPEC-ID decomposition self-check example; the 4 harness `V3R4 Self-Evolving` authoritative-SPEC decision-record citations — but `moai-harness-learner` + `moai/SKILL.md` + `harness.md` + `moai-meta-harness` count as the harness PRESERVE group; `manager-develop-prompt-template.md` is GENERALIZE/REMOVE). M3 refines the C2 allow-list and reduces `actual` to satisfy `actual <= allowlist`.

### AC-TNA-003 — C3 dates — **[DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]**

**[DEFERRED]** The generic ISO-date class (`2026-0[5-9]`) is enforced by the shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 via its `internal/template/internal_content_leak_test.go` strict-tier `S1-internal-date` class (`\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b`, opt-in via `MOAI_TEMPLATE_LEAK_STRICT=1`). This SPEC emits NO verification command for C3 — re-scanning here would create a second, divergent date allow-list in the same `internal/template/` Go package. The AC number is preserved for traceability contiguity. Verification of the date class is owned by the leak test:

```bash
# Date-class enforcement (owned by ISOLATION-001, NOT this SPEC):
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak
```

**Status**: deferred — no NEUTRALITY-owned baseline or post-fix target.

### AC-TNA-004 — C4 feedback_/memory refs ≤ allow-list count — **[KEPT — NEUTRALITY-unique]**

**Given** the migration matrix defines a feedback_/memory ref allow-list
**When** the auditor counts files matching the regex `feedback_|memory\.md`
**Then** the count shall be equal to or less than the allow-list size.

**Verification command** (corrected non-self-terminating awk per plan-audit iter-1 D2 fix — same flag-based extractor as AC-TNA-002):
```bash
actual=$(grep -rln 'feedback_\|memory\.md' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')
allowlist=$(awk '/^### C4 /{f=1;next} /^### C[0-9] /{f=0} f' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$actual" -le "$allowlist"
```

> **Kept (not deferred)**: the shipped `internal_content_leak_test.go` does NOT enforce the `feedback_` / `memory.md` substring reference class (default OR strict). Its C5 class enforces only memory *paths* (`~/.claude/projects/-Users-` / `.moai/backups/agent-archive-`), a disjoint pattern. Deferring C4 would silently drop enforcement, so C4 remains NEUTRALITY-owned (verified 2026-05-30). The corrected awk returns the C4 allow-list bullet count (non-zero, computable).

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: 9 files. Point-in-time; run-phase M3 re-measures before fixing.
**Post-fix expected**: ≤ allow-list size (M3 refines the C4 allow-list and reduces `actual` to satisfy `actual <= allowlist`).

### AC-TNA-005 — C5 CLAUDE.local.md refs = 0 (binary; allow-list empty)

**Given** `CLAUDE.local.md` is documented as a maintainer-only local file
**When** the auditor counts template files referencing `CLAUDE.local.md`
**Then** the count shall be 0 (C5 is a Binary FAIL class with an empty allow-list).

**Verification command** (C5 allow-list is "Empty (no exceptions)" per migration-matrix.md §C5, so the binary form is the canonical check; the awk form below confirms the matrix declares an empty allow-list):
```bash
test "$(grep -rln 'CLAUDE\.local\.md' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')" = "0"
# Matrix declares empty allow-list (corrected awk returns 0 bullets for C5):
allowlist=$(awk '/^### C5 /{f=1;next} /^### C[0-9] /{f=0} f' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$allowlist" = "0"
```

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: 3 files (was 10 at 2026-05-23; −7 partial prior cleanup). Point-in-time; run-phase M3 re-measures before fixing.
**Post-fix expected**: 0 files.

### AC-TNA-006 — C6 PR #N refs = 0

**Given** specific PR numbers are unsuitable for template distribution
**When** the auditor counts template files matching `PR #[0-9]+`
**Then** the count shall be 0.

**Verification command**:
```bash
test "$(grep -rln 'PR #[0-9]\+' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')" = "0"
```

**Baseline (re-measured 2026-05-30 at HEAD `ecda4ef04`)**: 3 files (unchanged). Point-in-time; run-phase M3 re-measures before fixing.
**Post-fix expected**: 0

### AC-TNA-007 — C7 commit hash — **[DEFERRED to SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001]**

**[DEFERRED]** The commit-hash class is enforced by the shipped sibling SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 via its `internal/template/internal_content_leak_test.go` strict-tier `S2-short-sha-sentence-final` class (`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`, opt-in via `MOAI_TEMPLATE_LEAK_STRICT=1`). This is a deliberately more conservative detector than the original NEUTRALITY proposal (broad `[a-f0-9]{7,40}` which matched 45 FP files including color codes and SHA test fixtures). This SPEC emits NO verification command for C7 — and crucially does NOT depend on a `TestTemplateNeutralityAudit/C7_commit_hash` subtest (the original AC referenced a test that did not exist). The AC number is preserved for traceability contiguity. Verification is owned by the leak test:

```bash
# Commit-hash-class enforcement (owned by ISOLATION-001, NOT this SPEC):
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak
```

**Status**: deferred — no NEUTRALITY-owned baseline or post-fix target. (Resolves plan-audit iter-1 D3: the original AC depended on a not-yet-existing test and an FP-saturated regex with no written discrimination rule.)

### AC-TNA-008 — Audit Go test PASS on darwin + linux + windows

**Given** the new audit script `internal/template/template_neutrality_audit_test.go` is implemented per REQ-TNA-009 (scoped to kept classes C1/C2/C4/C5/C6/C8; C3/C7 NOT scanned — owned by the leak test)
**When** the **isolated** `go test ./internal/template/... -run TestTemplateNeutralityAudit` is executed on darwin, linux, and windows runners
**Then** all three platforms shall report PASS (exit code 0) **for this test function in isolation** (the package-wide `go test ./internal/template/...` remains RED due to 13 pre-existing failures unrelated to this SPEC — see §1 package-RED note + spec.md §3.4; that package-wide RED MUST NOT block this AC).

**Verification command** (per platform — isolated `-run` form only):
```bash
go test ./internal/template/... -run TestTemplateNeutralityAudit -v
```

**Expected**: 0 FAIL for `TestTemplateNeutralityAudit` (and its C1/C2/C4/C5/C6/C8 subtests), optional WARN logs for C2 (bare-narrative)/C4 within allow-list. The audit script's pattern set is **disjoint** from `internal_content_leak_test.go` (no class enforced by both files). The C2 subtest implements the bare-narrative two-pass exclusion (RE2 lacks lookbehind) and its file set MUST equal `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'`.

### AC-TNA-009 — CI workflow `template-neutrality-check.yaml` triggers on template/ PR

**Given** a GitHub PR modifies files under `internal/template/templates/**`
**When** GitHub Actions evaluates `.github/workflows/template-neutrality-check.yaml`
**Then** the workflow shall be triggered and the `template-neutrality` job shall run.

**Verification command (pre-merge local check — M5 deliverable existence)**:
```bash
test -f .github/workflows/template-neutrality-check.yaml
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/template-neutrality-check.yaml'))"
```

**Verification command (post-merge, observe a representative PR)**:
```bash
gh run list --workflow=template-neutrality-check.yaml --limit 1 --json conclusion,status
```

**Expected**: pre-merge: workflow file exists AND YAML parses cleanly (exit 0). post-merge: `conclusion: success` for PRs without violations.

### AC-TNA-010 — Migration matrix complete (all 8 categories documented)

**Given** the SPEC requires explicit classification per category
**When** the auditor reads `migration-matrix.md`
**Then** the file shall contain exactly 8 sections (C1–C8), each with (a) detection regex, (b) action policy, (c) allow-list (or "Empty (no exceptions)"), (d) baseline count, (e) post-fix expected count.

**Verification command**:
```bash
matrix=.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md
count=$(grep -cE '^### C[1-8] ' "$matrix")
test "$count" = "8"
```

### AC-TNA-011 — False positive 0 for `GOOS=(linux|windows|darwin)` Go env var

**Given** `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` is a legitimate Go cross-compile env var (C8 false positive class)
**When** the audit script runs against `internal/template/templates/**`
**Then** the audit shall NOT emit any violation for these 4 hits / 3 files.

**Verification command** (positive test — confirm preservation; regex matches REQ-TNA-008 narrow form):
```bash
test "$(grep -rln 'GOOS=\(linux\|windows\|darwin\|freebsd\|openbsd\|netbsd\)' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')" = "3"
```

**Expected**: 3 files preserve `GOOS=` substring AND audit script returns PASS on those files.

### AC-TNA-012 — Template-First Rule guideline updated

**Given** `CLAUDE.local.md` §2 documents the Template-First Rule
**When** the auditor inspects `CLAUDE.local.md` for a subsection on acceptable template content
**Then** the subsection "Acceptable Content Range for Templates" (or equivalent heading) shall exist and enumerate: (a) acceptable categories, (b) rejected categories.

**Verification command**:
```bash
grep -q 'Acceptable Content Range\|template-acceptable-content' CLAUDE.local.md
```

### AC-TNA-013 — `moai init` clean run

**Given** the post-fix template tree is rebuilt via `make build`
**When** `moai init /tmp/test-tna-$(date +%s)` is executed
**Then** the command shall exit 0 and the deployed project shall contain no `/Users/`, no `PR #N`, and no removed feedback_/memory refs (other than allow-list entries).

**Verification command** (M2-M5 outcome smoke + REQ-TNA-013 contributor checklist discoverability):
```bash
make build
target=/tmp/test-tna-$(date +%s)
moai init "$target" --quiet
test "$(grep -rln '/Users/' "$target/.claude/" "$target/.moai/" 2>/dev/null | grep -v 'settings.local.json' | wc -l | tr -d ' ')" = "0"
# Additional check: contributor checklist discoverability (REQ-TNA-013)
grep -q 'Acceptable Content Range\|template-acceptable-content\|contributor-checklist\|Pre-PR Verification' CLAUDE.local.md
```

## §2 Test Plan

### Unit Tests (M5 deliverable)

- `internal/template/template_neutrality_audit_test.go` :: `TestTemplateNeutralityAudit` (scoped to kept classes; C3/C7 NOT scanned — owned by `internal_content_leak_test.go`)
  - Subtest `C1_macos_bias`: regex `/Users/`, expect 0 hits post-M2 (binary FAIL category)
  - Subtest `C2_v3r_bare_narrative`: bare-narrative `V3R[0-9]` via **two-pass exclusion** (Pass 1 `\bV3R[0-9]`; Pass 2 drop `SPEC-V3R`/`CONST-V3R`/`REQ-V3R` ID-embedded + preceding-rune `[A-Za-z0-9-]`), WARN if > allow-list (advisory category). MUST equal `grep -rlP '(?<![A-Za-z0-9-])V3R[0-9]'`. MUST NOT match the 299 ID-embedded `SPEC-V3R…`/`CONST-V3R…`/`REQ-V3R…` substrings (those are ISOLATION-001's `C1-spec-id` domain).
  - Subtest `C4_feedback_memory`: regex `feedback_|memory\.md`, WARN if > allow-list (advisory; NEUTRALITY-unique, not covered by leak test)
  - Subtest `C5_claude_local`: regex `CLAUDE\.local\.md`, FAIL if > allow-list (binary, empty allow-list)
  - Subtest `C6_pr_refs`: regex `PR #[0-9]+`, FAIL if > 0 (binary)
  - Subtest `C8_false_positive`: regex `GOOS=(linux|windows|darwin|freebsd|openbsd|netbsd)` MUST be preserved, audit MUST NOT emit violations on these hits
  - **NOT present**: `C3_dates` and `C7_commit_hash` subtests — these classes are owned by `internal_content_leak_test.go` strict-tier `S1-internal-date` / `S2-short-sha-sentence-final` (deferred per v0.1.1 rescope; avoids dual-allow-list drift in the `internal/template/` package)
  - Cross-platform PASS on darwin / linux / windows runners

### Integration Tests (M5 deliverable)

- `.github/workflows/template-neutrality-check.yaml` workflow validation:
  - On PR touching `internal/template/templates/**` → workflow triggered
  - On PR not touching template/ → workflow skipped (or no-op)
  - On C1/C5/C6 violation (kept binary classes) → workflow fails, PR status check fails
  - On C2/C4 WARN only → workflow passes with annotations
  - C3/C7 are out of this workflow's scope (owned by the leak-test gate; see spec.md REQ-TNA-009 SCOPE note)

### Manual Verification (M6 chore)

- AC-TNA-013 `moai init` clean run on a fresh `/tmp/` directory + contributor checklist discoverability check

### Out of Scope

본 acceptance phase가 명시적으로 **수행하지 않는** 검증:

- Sprint 1 in-progress SPEC directories의 frontmatter / content 검증
- `docs-site/content/{en,ko,ja,zh}/book/` 4-locale parity 자세히 (touch 없으면 vacant)
- `settings.local.json` 변경 검증 (runtime-managed 제외)
- Go code files (`internal/*.go`, `pkg/*.go`) sanitization 검증 — 별도 SPEC scope
- Template Go template rendering logic 정정 검증
- User-facing string literals preserved per EXCL-CCE-001 검증 (별도 SPEC)
- `make build` 후 `internal/template/embedded.go` regeneration 정합성 (`internal/template/embedded_test.go`이 별도 검증)
- 시간 예측 / 일정 추정 (HARD rule: never use time predictions)
