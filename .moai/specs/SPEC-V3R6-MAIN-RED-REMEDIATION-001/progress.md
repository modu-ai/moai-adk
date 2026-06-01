---
id: SPEC-V3R6-MAIN-RED-REMEDIATION-001
artifact: progress
version: "0.2.0"
created: 2026-05-30
updated: 2026-05-30
status: completed
---

# Progress Tracking — SPEC-V3R6-MAIN-RED-REMEDIATION-001

This file tracks run-phase implementation progress for SPEC-V3R6-MAIN-RED-REMEDIATION-001.
It carries M1-M4 milestone evidence (per plan.md §F) and the audit-ready signals
consumed by sync-phase manager-docs.

## §A.0 Pre-flight Verification

Pre-flight checks executed per plan.md §C at run-phase entry on 2026-05-30T12:00:00Z:

| Check | Command | Result |
|-------|---------|--------|
| Baseline test suite | `go test ./internal/template/...` | FAIL (13 parent tests RED; 4 groups identified) |
| Lint baseline | `golangci-lint run` | PASS (0 issues) |
| Plan-phase commit | `git log --oneline -1` | plan-phase commit (manager-spec) |
| Multi-session race | `git fetch && git rev-list --count --left-right origin/main...HEAD` | `0 0` (clean) |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (PASS) |

**Orchestration mode** (Phase 0.95): **Mode 5 — Sub-Agent Sequential** per
`.claude/rules/moai/workflow/orchestration-mode-selection.md` § Mode Catalog.

**Mode Selection** rationale: SPEC scope is test-correction + template-mirror-alignment only (no source code generation) — 4 parent tests correcion groups in single `internal/template` package domain. Per tier classification, scope < 1000 LOC, < 5 files modified (test corrections + template lint trigger fix). Tier M (300-1000 LOC) scope + single domain = Mode 5 sub-agent sequential per Anthropic 2026 Finding A4 caveat.

## §E.5 Mx-phase Audit-Ready Signal (2026-06-02)

```yaml
mx_complete_at: 2026-06-02
mx_status: skip-justified
mx_commit_sha: pending_close_backfill
mx_tag_count: 0
mx_skip_justified: true
mx_verdict: SKIP-JUSTIFIED
mx_evidence: |
  Tier S test-only SPEC with 0 production .go modifications (9 _test.go + 11 md + 29 other files). 
  No @MX:ANCHOR candidates (no new functions added). 
  No @MX:WARN candidates (no goroutines, no complexity ≥15, no state mutation).
  No new business rules or TODO items. 
  Skip judgment per mx-tag-protocol.md §a: file-exclusion criteria + tag-necessity rubric both satisfied.
```

## §E.2 Sync-phase Audit-Ready Signal

```yaml
sync_started_at: "2026-05-30T14:32:00Z"
sync_commit_sha: "(this commit)"
status: completed
```

Run-phase completion evidence (all 4 milestone groups):

### G1 — agents-layout test-correction (9 tests fixed)

**Completed**: Run-phase milestone M1-M4 (test correction per plan.md §F.1-F.4)
- `TestAgentFrontmatterAudit` — FLAT `moai/` layout expectation corrected
- `TestEmbeddedTemplates_AgentDefinitions` — retired `core/` / `meta/` / `expert/` expectations removed
- `TestManagerDevelopActiveAgentPresent` — 7 retained agents correctly identified in FLAT layout
- 6 additional agents-layout tests corrected to FLAT canonical layout

Verification:
```
$ go test -run TestAgent ./internal/template/...
ok  internal/template  0.5s
```

### G4 — hook-count expected-value update

**Completed**: `TestSettingsTemplateHookEventCount` hook count constant updated 20 → 21

Verification: `PreCommit` event presence confirmed in rendered settings.json template + test constant updated + HISTORY comment added.

### G3 — internal-content leak sanitization (30→0 violations)

**Completed**: 30 internal-content leak sites sanitized via 2 pathways:
- 29 sites: prose-substituted (SPEC-ID / REQ-ID / AC-ID / archive-date tokens replaced per §25 dictionary)
- 1 site: `pedagogicalAllowlist` path corrected (core/ → moai/)

Verification: `TestTemplateNoInternalContentLeak` returns 0 occurrence count.

### G2 — mirror-drift resolution + per-file policy

**Completed**: 8 mirror-drift files resolved per design-decision policy (plan.md §G):
- 7 files (rule files): byte-parity allowlist maintenance
- 1 file (manager-spec.md): per-file policy — excluded from byte-parity allowlist, covered by leak test instead (§25 sanitization precedence)

Verification: `TestRuleTemplateMirrorDrift` + `TestLateBranchTemplateMirror` PASS.

## §E.3 Run-phase status field

```yaml
run_started_at: "2026-05-30T12:00:00Z"
run_completed_at: "2026-05-30T14:32:00Z"
status: completed
m1_status: implemented
m2_status: implemented
m3_status: implemented
m4_status: implemented
```

## §E.4 Sync-phase audit-ready signal (extended)

**Manager-docs sync-phase completion verification**:

| Check | Status | Evidence |
|-------|--------|----------|
| spec.md frontmatter status transition | ✓ PASS | `status: in-progress → implemented`, `version: 0.1.1 → 0.2.0`, `updated: 2026-05-30` |
| progress.md §E.2 population | ✓ PASS | Sync-phase audit-ready signal populated (this section) |
| CHANGELOG.md entry | ✓ PASS | New entry appended under [Unreleased] |
| spec.md/plan.md/acceptance.md body untouched | ✓ PASS | Only frontmatter status + version + updated fields touched |
| All 4 groups GREEN gate | ✓ PASS | `go test ./internal/template/...` 0 fail + cross-platform build exit 0 |

## §E.5 Mx-phase audit-ready signal

Mx-phase ownership: orchestrator-direct (post-sync). Manager-docs sync-phase does not populate this section.

mx_commit_sha: (this commit)
