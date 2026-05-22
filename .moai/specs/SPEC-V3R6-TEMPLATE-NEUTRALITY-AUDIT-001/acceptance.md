---
id: SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001
title: "Template Neutrality Audit — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template-system, audit, acceptance, ci-guard, verification"
tier: L
---

# SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 — Acceptance Criteria

## §1 Binary Acceptance Criteria

13 binary AC scenarios. Each AC has an explicit verification command and expected outcome. AC failure = SPEC NOT done.

### Traceability Matrix (AC → REQ)

| AC | Satisfies REQ | Verification mode |
|----|---------------|-------------------|
| AC-TNA-001 | REQ-TNA-001 | Binary grep |
| AC-TNA-002 | REQ-TNA-002 | Awk range + grep |
| AC-TNA-003 | REQ-TNA-003 | Awk range + grep |
| AC-TNA-004 | REQ-TNA-004 | Awk range + grep |
| AC-TNA-005 | REQ-TNA-005 | Awk range + grep |
| AC-TNA-006 | REQ-TNA-006 | Binary grep |
| AC-TNA-007 | REQ-TNA-007 | Go test subtest |
| AC-TNA-008 | REQ-TNA-009 | Go test cross-platform |
| AC-TNA-009 | REQ-TNA-010 | Workflow file existence + YAML parse + run |
| AC-TNA-010 | REQ-TNA-011 | Matrix section header count |
| AC-TNA-011 | REQ-TNA-008 | False positive preservation |
| AC-TNA-012 | REQ-TNA-012 | Guideline subsection grep |
| AC-TNA-013 | REQ-TNA-013 + M2-M5 smoke | `moai init` clean run + checklist grep |

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

### AC-TNA-002 — C2 V3R[0-9] refs ≤ allow-list count

**Given** the migration matrix at `.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md` defines an allow-list of file paths permitted to contain `V3R[0-9]` substrings
**When** the auditor counts files matching the V3R regex
**Then** the count shall be equal to or less than the allow-list size.

**Verification command** (read allow-list count `N` from migration-matrix.md §C2):
```bash
actual=$(grep -rln 'V3R[0-9]' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')
allowlist=$(awk '/^### C2 /,/^### C[0-9]+/' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$actual" -le "$allowlist"
```

**Baseline (2026-05-23)**: 70 files
**Post-fix expected**: ≤ allow-list size (determined in M1)

### AC-TNA-003 — C3 2026-05-XX dates ≤ allow-list count

**Given** the migration matrix defines a date allow-list (canonical incident dates)
**When** the auditor counts files matching the date regex
**Then** the count shall be equal to or less than the allow-list size.

**Verification command**:
```bash
actual=$(grep -rln '2026-0[5-9]' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')
allowlist=$(awk '/^### C3 /,/^### C[0-9]+/' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$actual" -le "$allowlist"
```

**Baseline (2026-05-23)**: 32 files
**Post-fix expected**: ≤ allow-list size

### AC-TNA-004 — C4 feedback_/memory refs ≤ allow-list count

**Given** the migration matrix defines a feedback_/memory ref allow-list
**When** the auditor counts files matching the regex `feedback_|memory\.md`
**Then** the count shall be equal to or less than the allow-list size.

**Verification command**:
```bash
actual=$(grep -rln 'feedback_\|memory\.md' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')
allowlist=$(awk '/^### C4 /,/^### C[0-9]+/' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$actual" -le "$allowlist"
```

**Baseline (2026-05-23)**: 9 files
**Post-fix expected**: ≤ allow-list size

### AC-TNA-005 — C5 CLAUDE.local.md refs ≤ allow-list count (preferably 0)

**Given** `CLAUDE.local.md` is documented as a maintainer-only local file
**When** the auditor counts template files referencing `CLAUDE.local.md`
**Then** the count shall be 0 unless an allow-list entry justifies the reference.

**Verification command**:
```bash
actual=$(grep -rln 'CLAUDE\.local\.md' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')
allowlist=$(awk '/^### C5 /,/^### C[0-9]+/' .moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/migration-matrix.md | grep -c '^- ')
test "$actual" -le "$allowlist"
```

**Baseline (2026-05-23)**: 10 files
**Post-fix expected**: 0 (preferred) or ≤ allow-list size

### AC-TNA-006 — C6 PR #N refs = 0

**Given** specific PR numbers are unsuitable for template distribution
**When** the auditor counts template files matching `PR #[0-9]+`
**Then** the count shall be 0.

**Verification command**:
```bash
test "$(grep -rln 'PR #[0-9]\+' internal/template/templates/ 2>/dev/null | wc -l | tr -d ' ')" = "0"
```

**Baseline (2026-05-23)**: 3 files
**Post-fix expected**: 0

### AC-TNA-007 — C7 Commit hash refs ≤ allow-list count

**Given** specific commit hashes are unsuitable for template distribution
**When** the auditor counts template files matching commit hash patterns (7-40 hex chars) excluding known non-commit hex usages (color codes, SHA test fixtures registered in allow-list)
**Then** the count shall be ≤ allow-list size.

**Verification command** (audit script handles allow-list logic; manual grep approximation):
```bash
go test ./internal/template/... -run TestTemplateNeutralityAudit/C7_commit_hash
```

**Baseline (2026-05-23)**: ~2 files (post-dedup of false positives)
**Post-fix expected**: ≤ allow-list size (typically 0)

### AC-TNA-008 — Audit Go test PASS on darwin + linux + windows

**Given** the new audit script `internal/template/template_neutrality_audit_test.go` is implemented per REQ-TNA-009
**When** `go test ./internal/template/... -run TestTemplateNeutralityAudit` is executed on darwin, linux, and windows runners
**Then** all three platforms shall report PASS (exit code 0).

**Verification command** (per platform):
```bash
go test ./internal/template/... -run TestTemplateNeutralityAudit -v
```

**Expected**: 0 FAIL, optional WARN logs for C2/C3/C4 within allow-list.

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

- `internal/template/template_neutrality_audit_test.go` :: `TestTemplateNeutralityAudit`
  - Subtest `C1_macos_bias`: regex `/Users/`, expect 0 hits post-M2 (binary FAIL category)
  - Subtest `C2_v3r_refs`: regex `V3R[0-9]`, WARN if > allow-list (advisory category)
  - Subtest `C3_dates`: regex `2026-0[5-9]-[0-9]{2}`, WARN if > allow-list
  - Subtest `C4_feedback_memory`: regex `feedback_|memory\.md`, WARN if > allow-list
  - Subtest `C5_claude_local`: regex `CLAUDE\.local\.md`, FAIL if > allow-list (binary)
  - Subtest `C6_pr_refs`: regex `PR #[0-9]+`, FAIL if > 0 (binary)
  - Subtest `C7_commit_hash`: regex `[a-f0-9]{7,40}` filtered by allow-list, FAIL if > 0
  - Subtest `C8_false_positive`: regex `GOOS=(linux|windows|darwin)` MUST be preserved, audit MUST NOT emit violations on these hits
  - Cross-platform PASS on darwin / linux / windows runners

### Integration Tests (M5 deliverable)

- `.github/workflows/template-neutrality-check.yaml` workflow validation:
  - On PR touching `internal/template/templates/**` → workflow triggered
  - On PR not touching template/ → workflow skipped (or no-op)
  - On C1/C5/C6/C7 violation → workflow fails, PR status check fails
  - On C2/C3/C4 WARN only → workflow passes with annotations

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
