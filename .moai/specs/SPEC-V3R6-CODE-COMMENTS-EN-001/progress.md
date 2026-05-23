---
id: SPEC-V3R6-CODE-COMMENTS-EN-001
type: progress
version: "0.3.0"
created: 2026-05-23
updated: 2026-05-23
---

# Progress — SPEC-V3R6-CODE-COMMENTS-EN-001

## Status

- **Phase**: sync (Wave 7 complete — full run phase delivered on origin/main)
- **plan_complete_at**: 2026-05-23
- **plan_status**: audit-ready
- **plan_auditor_invoked**: true (PASS aggregate 0.87 iter 2, Tier L threshold 0.85, margin +0.02)
- **run_phase_complete_at**: 2026-05-23
- **run_status**: delivered (Wave 7-2 Test C + Wave 7 partial cumulative, 64 test files total, commits b35ca8b96 + ed064a6f2)
- **sync_phase_complete_at**: 2026-05-23
- **sync_status**: complete (CHANGELOG.md entry added, direct push to origin/main per Hybrid Trunk Tier S docs sync)
- **next_phase**: archive (SPEC-V3R6-CODE-COMMENTS-EN-001 complete — no follow-up waves pending)

## Wave 1 — Foundation (COMPLETE 2026-05-23)

**Scope**: 9 non-test Go files (config=4, core=1, hook=1, spec=3).

**Files modified**:
- `internal/config/defaults.go` (2 comment regions)
- `internal/config/errors.go` (9 sentinel error docblocks)
- `internal/config/loader.go` (LoadHarnessConfig + detectSchemaDrift Korean documentation translated)
- `internal/config/types.go` (HarnessConfig + sub-structs + ContextConfig/InterviewConfig/DesignConfig docblocks + harnessFileWrapper / 4 MIG-003 wrappers + ralphFileWrapper)
- `internal/core/project/initializer.go` (1 godoc line: REQ-V3R2-RT-007-001)
- `internal/hook/session_start.go` (1 comment block: SPEC-V3R2-RT-007 REQ-020 silent migration)
- `internal/spec/drift.go` (gitLogWindowSize + getGitImpliedStatus walker + commitMatchesSPECID + shouldSkipCommitTitle + inline scanner comments — full file Korean → English)
- `internal/spec/lint.go` (ExtractFrontmatter + HarnessLevel godoc + FrontmatterSchemaRule + terminalStatusEnum + StatusGitConsistencyRule inline comment)
- `internal/spec/status.go` (1 comment line preserving Korean string literal `"상태"`)

**Wave 1 AC matrix (in-scope subset)**:

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CCE-001 | PASS | `grep -rn '//.*[가-힣]' <wave-1 scope> --include="*.go"` | 0 matches |
| AC-CCE-002 | PASS | `grep -rnE '/\*.*[가-힣]' <wave-1 scope> --include="*.go"` | 0 matches |
| AC-CCE-003 | PASS | `grep -rn '@MX:[A-Z]*[: ].*[가-힣]' <wave-1 scope> --include="*.go"` | 0 matches |
| AC-CCE-004 (Wave 1 scope) | PASS | string literal Korean count pre=8 / post=8 (status.go format markers + lint.go OrphanBCID error msg, all EXCL-CCE-001 preserved) | byte-identical |
| AC-CCE-005 | PASS | `go build ./...` | exit 0 |
| AC-CCE-006 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| AC-CCE-007 (Wave 1 scope) | PASS | `go test ./internal/{config,core/project,hook,spec}/...` NEW failures vs baseline | NEW=0; pre-existing baseline `TestAuditRegistrationParity` + `TestAuditThreeWaySync` (WorktreeCreate unwire residual per commit a3239d3de, EXCL-CCE-008 preserved) |
| AC-CCE-008 (Wave 1 scope) | PASS | `golangci-lint run --timeout=2m ./internal/config/... ./internal/core/project/... ./internal/hook/... ./internal/spec/...` | `0 issues.` |
| AC-CCE-011 (Wave 1 scope) | PASS | identifier-preservation grep — all SPEC-/REQ-/AC- IDs preserved verbatim in translated comments | verbatim |

**Cross-platform builds**: PASS both darwin and windows/amd64.

**C-HRA-008 subagent boundary** (`internal/hook/session_start.go` touched): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/session_start.go | grep -v '_test.go' | grep -v '// '` → 0 matches.

**Lint baseline preservation**: NEW issues = 0 in Wave 1 scope. Wave 1 file gofmt drift on `internal/config/types.go`, `internal/spec/drift.go`, `internal/spec/lint.go` is **pre-existing baseline** (verified via stash-test on HEAD before Wave 1 edits) — out of scope per REQ-CCE-008 (no non-comment Go syntax/whitespace changes).

**String literal preservation evidence** (8 surviving Korean string literals in Wave 1 scope, all under EXCL-CCE-001):
- `internal/spec/status.go:86,91,172,177,209,241,248` — `"| 상태 |"`, `"- **상태**:"`, `"상태"` (Korean SPEC table/markdown format detection markers — functional behavior)
- `internal/spec/lint.go:756` — OrphanBCID error message (REQ-CCE-005 + OQ-CCE-002 default preservation)

## Artifacts (5/5 complete)

| File | Lines | Status |
|------|-------|--------|
| spec.md | ~210 | DRAFT |
| plan.md | ~310 | DRAFT |
| acceptance.md | ~430 | DRAFT |
| design.md | ~410 | DRAFT |
| research.md | ~480 | DRAFT |

## Frontmatter validation

All 9 required canonical fields present in spec.md:

- [x] `id: SPEC-V3R6-CODE-COMMENTS-EN-001` (matches regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`)
- [x] `version: "0.1.0"` (quoted)
- [x] `status: draft` (enum)
- [x] `created_at: 2026-05-23` (ISO date)
- [x] `updated_at: 2026-05-23` (ISO date)
- [x] `author: manager-spec`
- [x] `priority: Medium` (Title-case)
- [x] `labels: [code-quality, comments, internationalization, en-migration, mass-migration]` (YAML array)
- [x] `issue_number: null`

Optional fields included: `tier: L`, `phase: v3.0.0`, `module: internal/`, `tags`, `depends_on: []`, `related_specs: []`.

## AC count + traceability

- AC count: **12 binary ACs** (within 8-12 target)
- REQ ↔ AC traceability: **100%** (8 REQs all covered by at least 1 AC, see acceptance.md §5)
- Edge cases: 5 documented (acceptance.md §3)
- DoD items: 7 (acceptance.md §4)

## Wave-split summary

| Wave | Scope | Files | ~Lines | Priority |
|------|-------|-------|--------|----------|
| 1 | Foundation (config/core/hook/spec) | 9 | ~250 | Critical |
| 2 | CLI surface | 25 | ~1000 | High (OQ-CCE-001 Cobra blocker) |
| 3 | Harness + migration | 23 | ~750 | High |
| 4 | pkg/ + remaining non-test | 3 | ~75 | Low |
| 5 | Test A (cli + template) | 50 | ~1250 | Medium |
| 6 | Test B (harness + lsp + hook) | 38 | ~750 | Medium |
| 7 | Test C (remaining test) | 15 | ~300 | Low |

**Total**: 267 files, ~4,375 lines, 7 independent PRs

## Open questions

| ID | Decision | User input required? |
|----|----------|----------------------|
| OQ-CCE-001 | Cobra `Use:/Short:/Long:` Korean (8 cases found) | YES at Wave 2 entry — default Option B (영어화) recommended |
| OQ-CCE-002 | `errors.New("...")` Korean preservation | NO (default acceptable, follow-up SPEC) |
| OQ-CCE-003 | Log message Korean preservation | NO (default acceptable, follow-up SPEC) |
| OQ-CCE-004 | main 직진 vs feat-branch | NO (default feat-branch per CLAUDE.local.md §23 Tier L) |

## Blockers

None at plan-phase exit.

## Self-verification deliverables (plan-phase)

- **E1 SPEC artifacts**:
  - spec.md (210 lines): EARS 8 REQs, 5 EXCL Out-of-Scope (h3 sub-section per lessons #B6), HISTORY, 7 risks
  - plan.md (310 lines): 7-wave decomposition, Section A-E template reference (MANDATORY Tier L)
  - acceptance.md (430 lines): 12 binary ACs, REQ↔AC matrix, edge cases, DoD, verification scripts
  - design.md (410 lines): Translation methodology, Agent batch strategy, anti-patterns
  - research.md (480 lines): Korean inventory, identifier baseline (1733/1444/830), 10 sample patterns
- **E2 Frontmatter validation**: 9 canonical fields PASS (see above)
- **E3 AC traceability**: 100% (8 REQs × 12 ACs, all REQ has ≥1 AC link)
- **E4 spec-lint**: Not run locally (recommend CI). Heading check: `### 5.1 Out of Scope` (h3 per lessons #B6) present in spec.md.
- **E5 plan-auditor**: Not invoked (Tier L threshold 0.85 — orchestrator decision; recommended via parallel call)
- **E6 Wave-split**: 7 waves documented in plan.md §3 with file ownership and Justification
- **E7 Open questions**: 4 OQs documented, 1 (OQ-CCE-001) becomes BLOCKER at Wave 2 entry

## Next action

```
# Option 1: Invoke plan-auditor (recommended at Tier L)
Use the plan-auditor subagent to audit SPEC-V3R6-CODE-COMMENTS-EN-001 at Tier L threshold 0.85.

# Option 2: Proceed to run-phase (after audit PASS or skip)
/moai run SPEC-V3R6-CODE-COMMENTS-EN-001
# → Wave 1 entry: manager-develop Section A-E delegation,
#                 scope: internal/{config,core,hook,spec}/**/*.go (9 files, ~250 Korean lines)
```

## Notes for run-phase

1. **Section A-E delegation MANDATORY** for Tier L. Use `.claude/rules/moai/development/manager-develop-prompt-template.md` 5-section structure.
2. **Per-Wave PR strategy** — sequential merge recommended (Wave N PR merged before Wave N+1 entry).
3. **Baseline capture** at Wave 1 entry: lint (~22 issues), test (HARNESS-RENAME-001 cascade 3 FAILs + others, EXCL-CCE-008).
4. **Wave 2 BLOCKER** at OQ-CCE-001 — Cobra string literal vs AC-CCE-010 conflict. AskUserQuestion required for Option A (preserve) vs Option B (translate, recommended).
5. **String literal preservation** is critical (REQ-CCE-005 / AC-CCE-004) — REQ-CCE-008 byte-identity → no go.mod / no template / no config drift.
6. **Identifier preservation** is critical (REQ-CCE-004 / AC-CCE-011) — pre-Wave baseline: 1733 SPEC-IDs / 1444 REQ-IDs / 830 AC-IDs.

---

Version: 0.1.0
Plan-phase complete: 2026-05-23
