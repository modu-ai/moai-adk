---
id: SPEC-HARNESS-EVOLVE-002
title: "Curator Editable Surfaces — Loop 2 (write layer) of the self-evolving harness"
version: "0.1.0"
status: completed
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness, internal/harness/curator, internal/template, internal/template/templates"
lifecycle: spec-anchored
tags: "harness-evolve-epic, curator, managed-block, learned-workflow, snapshot-rollback, lineage, 3-zone, recall-contract, template-first"
era: V3R6
tier: L
depends_on: [SPEC-HARNESS-EVOLVE-001]
---

# SPEC-HARNESS-EVOLVE-002 — Acceptance Criteria

> Counterpart to `spec.md` (SSOT for requirements) and `plan.md` (SSOT for
> implementation approach). This document owns the machine-verifiable
> acceptance criteria matrix. Every REQ-HEV2-XXX in spec.md maps to at least
> one AC-HEV2-XXX here; cross-file registrations (template mirror,
> `mergeSectionBased` extension, `LineageEntry` extension, snapshot manifest
> entry) are pinned as SEPARATE baseline-0 ACs per the reachability
> discipline (inherited from EVOLVE-001 +
> `feedback_ac_token_presence_not_reachability`).

## §A. AC Matrix Mapping (REQ → AC)

| REQ | AC(s) | Severity | Verification kind |
|-----|-------|----------|-------------------|
| REQ-HEV2-001 (typed writer API) | AC-HEV2-001, AC-HEV2-002 | MUST | Go test (signature existence + behavioral) |
| REQ-HEV2-002 (BlockType enum) | AC-HEV2-003 | MUST | Go test (enum values + InjectMarker backward-compat) |
| REQ-HEV2-003 (idempotent) | AC-HEV2-004 | MUST | Go test (byte-diff = 0 on identical re-apply) |
| REQ-HEV2-004 (byte preservation) | AC-HEV2-005 | MUST | Go test (pre-block + post-block bytes unchanged) |
| REQ-HEV2-005 (append-mode) | AC-HEV2-006 | MUST | Go test (fresh-block append + newline convention) |
| REQ-HEV2-006 (LEARNED marker contract) | AC-HEV2-007, AC-HEV2-008 | MUST | Go test (marker regex + heading atomic match) |
| REQ-HEV2-007 (bullet CRUD) | AC-HEV2-009, AC-HEV2-010, AC-HEV2-011 | MUST | Go test (AddBullet / UpdateBullet / DeleteBullet single-line) |
| REQ-HEV2-008 (≤3K budget) | AC-HEV2-012, AC-HEV2-013 | MUST | Go test (measureAlwaysLoaded reuse + ErrDigestBudgetExceeded) |
| REQ-HEV2-009 (≤20-bullet cap) | AC-HEV2-014 | MUST | Go test (ErrBulletCapExceeded on 21st bullet) |
| REQ-HEV2-010 (ledger_key linkage) | AC-HEV2-015 | MUST | Go test (key present in trailing HTML comment) |
| REQ-HEV2-011 (anti-fabrication) | AC-HEV2-016, AC-HEV2-017 | MUST | Go test (forbidden-pattern rejection) + grep (template neutrality) |
| REQ-HEV2-012 (append-only) | AC-HEV2-018 | MUST | Go test (existing entry bytes unchanged on append) |
| REQ-HEV2-013 (LOCAL marker contract) | AC-HEV2-019 | MUST | Go test (marker regex for LOCAL variant) |
| REQ-HEV2-014 (dedup append) | AC-HEV2-020 | MUST | Go test (ErrDuplicateAppend on same ledger_key) |
| REQ-HEV2-015 (digest layer contract) | AC-HEV2-021 | MUST | Go test (godoc + type contract) |
| REQ-HEV2-016 (ledger layer contract) | AC-HEV2-022 | MUST | Go test (godoc + type contract) |
| REQ-HEV2-017 (cross-layer linkage) | AC-HEV2-023 | MUST | Go test (null ledger_key → provisional marker) |
| REQ-HEV2-018 (principle codification) | AC-HEV2-024 | MUST | Go test (no WriteFullEvidenceToDigest code path) |
| REQ-HEV2-019 (merge preservation) | AC-HEV2-025, AC-HEV2-026 | MUST | Go test (mergeSectionBased preserves populated block) |
| REQ-HEV2-020 (no silent clobber) | AC-HEV2-027 | MUST | Go test (conflict surfaced, not auto-resolved) |
| REQ-HEV2-021 (snapshot extension) | AC-HEV2-028 | MUST | Go test (distinct restore units for each surface) |
| REQ-HEV2-022 (byte-identical rollback) | AC-HEV2-029 | MUST | Go test (post-rollback bytes == pre-write bytes) |
| REQ-HEV2-023 (lineage recording) | AC-HEV2-030 | MUST | Go test (LineageEntry carries required fields) |
| REQ-HEV2-024 (evidence-or-null) | AC-HEV2-031 | MUST | Go test (null in evidence-ref absent fields) |
| REQ-HEV2-025 (Tier 4 gate) | AC-HEV2-032 | MUST | Go test (10-observation pattern → CLAUDE.md write) |
| REQ-HEV2-026 (Tier 3 gate) | AC-HEV2-033 | MUST | Go test (5-observation pattern → CLAUDE.local.md append) |
| REQ-HEV2-027 (no self-tier-escalation) | AC-HEV2-034 | MUST | Go test (ErrTierNotQualified on under-tier write) |
| REQ-HEV2-028 (Template-First) | AC-HEV2-035 | MUST | grep (template edit precedes live copy) |
| REQ-HEV2-029 (empty-marker shipping) | AC-HEV2-036, AC-HEV2-037 | MUST | grep (empty marker in template tree + make build embeds it) |
| REQ-HEV2-030 (template isolation) | AC-HEV2-038 | MUST | grep (no SPEC IDs / REQ tokens / dates / SHAs in template block) |
| REQ-HEV2-031 (machine-signal-only) | AC-HEV2-039 | MUST | Go test (writer rejects free-text model prose) |
| REQ-HEV2-032 (L5 approval gate) | AC-HEV2-040, AC-HEV2-041 | MUST | Go test (approval-token required + rejection → lineage + no write) |
| REQ-HEV2-033 (mechanical rollback trigger) | AC-HEV2-042 | MUST | Go test (no model-self-report entry point) |
| REQ-HEV2-034 (coverage + quality) | AC-HEV2-043 | MUST | `go test -cover` ≥ 90% + golangci-lint clean |
| REQ-HEV2-035 (subagent boundary) | AC-HEV2-044 | MUST | grep (no AskUserQuestion in curator/) |
| REQ-HEV2-036 (no new hook surface) | AC-HEV2-045, AC-HEV2-046 | MUST | grep (settings.json hook registration unchanged + no new wrapper) |

**Cross-file reachability ACs** (per `feedback_ac_token_presence_not_reachability`):

| AC | Target | Baseline | Post-implementation |
|----|--------|----------|---------------------|
| AC-HEV2-047 | `internal/harness/curator/writer.go` exists with `WriteManagedBlock` exported | 0 matches | ≥ 1 match |
| AC-HEV2-048 | `internal/harness/layer3.go InjectMarker` calls `curator.WriteManagedBlock` | 0 matches | ≥ 1 match (backward-compat wrapper) |
| AC-HEV2-049 | `internal/merge/strategies.go` defines the `managedSectionHeadings` allow-list var (design.md §D.1 H-2 option (a)) AND `mergeSectionBased` consults it on the preservation path | 0 matches | ≥ 2 matches (1 var declaration + ≥1 consultation in mergeSectionBased) |
| AC-HEV2-050 | `internal/harness/types.go LineageEntry` carries `LearnedSurface` + `BulletsChanged` + `SnapshotDir` fields | 0 matches | ≥ 3 matches (one per field) |
| AC-HEV2-051 | snapshot manifest JSON (written by `internal/harness/applier.go createSnapshot`) carries ≥2 distinct `learned_surface` values when both CLAUDE.md + CLAUDE.local.md are snapshotted — per-surface restore units (design.md §C.1 manifest entry shape) | 0 distinct `learned_surface` values (manifest schema has no `learned_surface` field today) | ≥ 2 distinct `learned_surface` values in a dual-surface snapshot fixture |
| AC-HEV2-052 | `internal/template/templates/CLAUDE.md` carries the empty `MOAI:LEARNED-WORKFLOW` marker | 0 matches | heading + start marker + end marker (≥ 3 matches) |
| AC-HEV2-053 | **(meta-traceability check — NOT a reachability AC; relabeled per plan-audit iter-1 D7)** `.moai/specs/SPEC-HARNESS-EVOLVE-002/spec.md` REQ→AC matrix covers all 36 REQs | (this file) | 100% coverage |

## §B. Severity Definitions

- **MUST**: failure blocks merge. All MUST ACs MUST PASS.
- **SHOULD**: failure emits a warning; merge proceeds with debt recorded.
- **NICE**: failure emits an info note.

All ACs in §A are MUST unless otherwise stated.

## §C. Traceability (REQ → AC → test file)

Every AC maps to a concrete Go test file + test name. The test files live
alongside the implementation per Go convention (`_test.go` co-located):

| AC | Test file | Test name (sketch) |
|----|-----------|---------------------|
| AC-HEV2-001 | `internal/harness/curator/writer_test.go` | `TestWriteManagedBlock_Signature` |
| AC-HEV2-002 | `internal/harness/curator/writer_test.go` | `TestWriteManagedBlock_AtomicReplace` |
| AC-HEV2-003 | `internal/harness/curator/writer_test.go` | `TestBlockTypeEnum_IncludesHarnessGenerated` |
| AC-HEV2-004 | `internal/harness/curator/writer_test.go` | `TestWriteManagedBlock_Idempotent_ZeroByteDiff` |
| AC-HEV2-005 | `internal/harness/curator/writer_test.go` | `TestWriteManagedBlock_PreBlockPostBlockPreserved` |
| AC-HEV2-006 | `internal/harness/curator/writer_test.go` | `TestWriteManagedBlock_AppendMode_NewlineConvention` |
| AC-HEV2-007 | `internal/harness/curator/marker_test.go` | `TestLearnedWorkflowMarker_RegexMatchesHeadingPlusMarkers` |
| AC-HEV2-008 | `internal/harness/curator/marker_test.go` | `TestLearnedWorkflowMarker_AtomicMatchGroup` |
| AC-HEV2-009 | `internal/harness/curator/crud_test.go` | `TestAddBullet_SingleLine_RewritesOnlyTarget` |
| AC-HEV2-010 | `internal/harness/curator/crud_test.go` | `TestUpdateBullet_ByLedgerKey` |
| AC-HEV2-011 | `internal/harness/curator/crud_test.go` | `TestDeleteBullet_ByLedgerKey_PreservesOthers` |
| AC-HEV2-012 | `internal/harness/curator/budget_test.go` | `TestWriteManagedBlock_BudgetExceeded_ErrDigestBudgetExceeded` |
| AC-HEV2-013 | `internal/harness/curator/budget_test.go` | `TestMeasureAlwaysLoaded_PerSectionAttribution` |
| AC-HEV2-014 | `internal/harness/curator/budget_test.go` | `TestWriteManagedBlock_BulletCapExceeded` |
| AC-HEV2-015 | `internal/harness/curator/crud_test.go` | `TestBullet_LedgerKeyTrailingHTMLComment` |
| AC-HEV2-016 | `internal/harness/curator/antifabrication_test.go` | `TestWriteManagedBlock_RejectsForbiddenPatterns` |
| AC-HEV2-017 | (template-tree grep) | `internal/template/internal_content_leak_test.go` extension |
| AC-HEV2-018 | `internal/harness/curator/append_test.go` | `TestAppendOnly_ExistingBytesUnchanged` |
| AC-HEV2-019 | `internal/harness/curator/marker_test.go` | `TestLearnedLocalMarker_RegexMatchesHeadingPlusMarkers` |
| AC-HEV2-020 | `internal/harness/curator/append_test.go` | `TestAppendOnly_DedupSameLedgerKey_ErrDuplicateAppend` |
| AC-HEV2-021 | `internal/harness/curator/recall_test.go` | `TestRecallContract_DigestLayerSummaryOnly` |
| AC-HEV2-022 | `internal/harness/curator/recall_test.go` | `TestRecallContract_LedgerLayerSearchInterface` |
| AC-HEV2-023 | `internal/harness/curator/crud_test.go` | `TestBullet_ProvisionalNullLedgerKey` |
| AC-HEV2-024 | `internal/harness/curator/recall_test.go` | `TestRecallContract_NoWriteFullEvidencePath` |
| AC-HEV2-025 | `internal/merge/strategies_test.go` | `TestMergeSectionBased_PreservesPopulatedLearnedBlock` |
| AC-HEV2-026 | `internal/merge/strategies_test.go` | `TestMergeSectionBased_EmptyUpstreamPopulatedLocal` |
| AC-HEV2-027 | `internal/merge/strategies_test.go` | `TestMergeSectionBased_ConflictInsideMarkers_Surfaced` |
| AC-HEV2-028 | `internal/harness/applier_test.go` | `TestCreateSnapshot_DistinctRestoreUnitsPerSurface` |
| AC-HEV2-029 | `internal/harness/applier_test.go` | `TestRestoreSnapshot_ByteIdenticalRollback` |
| AC-HEV2-030 | `internal/harness/lineage_test.go` | `TestWriteLineageEntry_LearnedSurfaceFields` |
| AC-HEV2-031 | `internal/harness/lineage_test.go` | `TestLineageEntry_EvidenceOrNull` |
| AC-HEV2-032 | `internal/harness/curator/tier_test.go` | `TestTier4Qualified_ClaudeMdWrite` |
| AC-HEV2-033 | `internal/harness/curator/tier_test.go` | `TestTier3Qualified_ClaudeLocalMdAppend` |
| AC-HEV2-034 | `internal/harness/curator/tier_test.go` | `TestUnderTierWrite_ErrTierNotQualified` |
| AC-HEV2-035 | (grep across commits) | `git log --diff-filter=A -- internal/template/templates/CLAUDE.md` ordering |
| AC-HEV2-036 | (template-tree grep) | `grep -c "MOAI:LEARNED-WORKFLOW" internal/template/templates/CLAUDE.md` ≥ 1 |
| AC-HEV2-037 | (build verification) | `make build` then `strings moai | grep "MOAI:LEARNED-WORKFLOW"` ≥ 1 |
| AC-HEV2-038 | (template-tree grep) | `internal/template/internal_content_leak_test.go` extended for the new block |
| AC-HEV2-039 | `internal/harness/curator/antifabrication_test.go` | `TestWriteManagedBlock_RejectsModelSelfReport` |
| AC-HEV2-040 | `internal/harness/curator/approval_test.go` | `TestWriteManagedBlock_RequiresApprovalToken` |
| AC-HEV2-041 | `internal/harness/curator/approval_test.go` | `TestWriteManagedBlock_RejectionRecordsLineage_NoFileWrite` |
| AC-HEV2-042 | `internal/harness/curator/rollback_test.go` | `TestRollbackTrigger_MechanicalOnly_NoModelSelfReport` |
| AC-HEV2-043 | (coverage + lint) | `go test -cover ./internal/harness/...` ≥ 90%; `golangci-lint run` exit 0 |
| AC-HEV2-044 | (subagent-boundary grep) | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ \| grep -v "_test.go" \| grep -v "// "` → 0 |
| AC-HEV2-045 | (settings.json diff) | `git diff main -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` → no hook-registration changes attributable to this SPEC |
| AC-HEV2-046 | (hook-wrapper grep) | `ls .claude/hooks/moai/handle-*curator*.sh 2>/dev/null \| wc -l` → 0 (no new wrapper) |
| AC-HEV2-047 | (reachability grep) | `grep -rn "func WriteManagedBlock" internal/harness/curator/` → ≥ 1 match |
| AC-HEV2-048 | (reachability grep) | `grep -n "curator.WriteManagedBlock(" internal/harness/layer3.go` → ≥ 1 match (the literal call site with opening paren — discriminates the actual invocation from a type-reference comment; per plan-audit iter-1 D4 the original 2-token compound grep was acceptable but is tightened here since it is a one-line discriminating change) |
| AC-HEV2-049 | (reachability grep) | `grep -n "managedSectionHeadings" internal/merge/strategies.go` → ≥ 2 matches (the `var managedSectionHeadings = []string{...}` declaration + its consultation inside `mergeSectionBased`'s preservation path — discriminating per design.md §D.1 H-2 option (a) explicit allow-list; a single dead comment mentioning the symbol once does NOT satisfy the ≥2 threshold) |
| AC-HEV2-050 | (reachability grep) | `grep -E "LearnedSurface|BulletsChanged|SnapshotDir" internal/harness/types.go` → ≥ 3 matches |
| AC-HEV2-051 | (manifest-structure assertion, driven by Go test `TestCreateSnapshot_DistinctRestoreUnitsPerSurface` AC-HEV2-028) | on a fixture dual-surface snapshot manifest: `jq -r '.Files[].learned_surface' <manifest> \| sort -u \| wc -l` → ≥ 2 (structural proof of per-surface restore units — NOT a substring grep that a single source comment satisfies; the manifest is a runtime JSON artifact per design.md §C.1) |
| AC-HEV2-052 | (template-tree grep) | `grep -c "moai:learned-start\|moai:learned-end\|MOAI:LEARNED-WORKFLOW" internal/template/templates/CLAUDE.md` → ≥ 3 |
| AC-HEV2-053 | (self-coverage — **meta-traceability check, NOT a reachability AC; relabeled per plan-audit iter-1 D7**) | this matrix covers all 36 REQs in spec.md §C (100%) |

## §D. Given-When-Then Scenarios (key behavioral ACs)

### Scenario 1 — Idempotent re-application (AC-HEV2-004)

```gherkin
Given a CLAUDE.md file with a populated MOAI:LEARNED-WORKFLOW block
 And the block contains 5 bullets with distinct ledger_keys
When WriteManagedBlock is called with content byte-identical to the existing block
Then the file's bytes after the call are byte-identical to the file's bytes before
 And no diff is produced
 And the call exits 0 with no error
```

### Scenario 2 — Budget enforcement (AC-HEV2-012)

```gherkin
Given a CLAUDE.md file with an empty MOAI:LEARNED-WORKFLOW block
 And the proposed write carries bullets totaling 3,500 characters
When WriteManagedBlock is called with the proposed content
Then the writer returns ErrDigestBudgetExceeded
 And the file is NOT touched (bytes unchanged)
 And the lineage records the rejected proposal with outcome "rejected"
```

### Scenario 3 — Bullet CRUD single-line rewrite (AC-HEV2-009)

```gherkin
Given a CLAUDE.md file with a MOAI:LEARNED-WORKFLOW block containing bullets [A, B, C, D]
 And each bullet carries a distinct ledger_key
When AddBullet is called with a new bullet E (ledger_key "E")
Then the block now contains [A, B, C, D, E]
 And the bytes of bullets A, B, C, D are unchanged
 And only a single new line is inserted (the E bullet line)
```

### Scenario 4 — `mergeSectionBased` preservation (AC-HEV2-025)

```gherkin
Given a template CLAUDE.md with an EMPTY MOAI:LEARNED-WORKFLOW marker
 And a local CLAUDE.md with a POPULATED MOAI:LEARNED-WORKFLOW block (5 bullets)
When mergeSectionBased is called with (base=template, current=local, updated=template-new)
 Then the merge result preserves the local populated block verbatim
 And no clobber occurs
 And the merge exits without conflict
```

### Scenario 5 — Byte-identical rollback (AC-HEV2-029)

```gherkin
Given a CLAUDE.md file with a MOAI:LEARNED-WORKFLOW block containing bullets [A, B, C]
 And a snapshot has been taken recording the file's bytes
When AddBullet adds bullet D, then DeleteBullet deletes bullet B
 And RestoreSnapshot is called with the snapshot directory
Then the file's bytes are byte-identical to the pre-write state (bullets [A, B, C])
 And the byte-length check passes
 And the marker block is not orphaned
```

### Scenario 6 — L5 approval gate rejection (AC-HEV2-041)

```gherkin
Given a Curator proposal to write bullet X to the MOAI:LEARNED-WORKFLOW block
 And the orchestrator's AskUserQuestion round returns a rejection
When the writer receives the rejection token
Then the writer does NOT touch the file (bytes unchanged)
 And the writer appends a LineageEntry with outcome "rejected"
 And the LineageEntry carries the rejection rationale (audit trail)
```

### Scenario 7 — Tier-differentiated write path (AC-HEV2-032, AC-HEV2-033, AC-HEV2-034)

```gherkin
Given a pattern P3 with 5 observations (Tier 3 qualified, NOT Tier 4)
 And a pattern P4 with 10 observations (Tier 4 qualified)
When the Curator proposes writes for P3 and P4
Then P3's proposal targets the CLAUDE.local.md append surface (Tier 3)
 And P4's proposal targets the CLAUDE.md managed block surface (Tier 4)
 And a 6-observation pattern P_under neither reaches Tier 4 nor triggers ErrTierNotQualified when written to Tier 3
 And direct Tier 4 write attempt for P_under returns ErrTierNotQualified
```

### Scenario 8 — Anti-fabrication input validation (AC-HEV2-016)

```gherkin
Given a proposed bullet with text "Per SPEC-HARNESS-EVOLVE-001 REQ-HEV-006, the ledger records..."
When WriteManagedBlock evaluates the proposed content
Then the writer rejects the bullet (matches the forbidden SPEC-ID + REQ-token pattern)
 And the rejection is recorded in lineage with rationale "anti-fabrication: internal SPEC ID / REQ token in bullet text"
 And the file is NOT touched
```

### Scenario 9 — Cross-layer linkage null on provisional bullet (AC-HEV2-023)

```gherkin
Given a Tier 1 observation (single occurrence, no aggregated evidence yet)
 And the Curator emits a provisional digest bullet for it
When the bullet is written
Then the bullet carries ledger_key "null" (or empty with Provisional=true)
 And the bullet is marked provisional in the LineageEntry
 And the next-tier promotion (Tier 3) replaces the provisional marker with a real ledger_key
```

## §E. Edge Cases

- **E1 — Empty file (zero bytes)**: `WriteManagedBlock` on a zero-byte file
  appends the block as the file's first content, with a leading newline to
  respect the file's existing (empty) trailing-newline convention.
- **E2 — File without trailing newline**: appending the block inserts a
  separating newline first (no concatenation of the last existing line with
  the new heading).
- **E3 — Multiple consecutive blank lines pre-block**: preserved verbatim
  (byte-preservation, REQ-HEV2-004).
- **E4 — Block marker orphaned (start marker present, end marker absent)**:
  the writer treats this as a corrupt block and refuses the write with
  `ErrMarkerCorrupt` (does NOT attempt recovery).
- **E5 — Two blocks of the same type (should never happen, but defensive)**:
  the writer treats this as a corrupt state and refuses with
  `ErrDuplicateBlock`.
- **E6 — Block at end of file with no trailing newline**: the writer
  preserves the no-trailing-newline convention when idempotent (REQ-HEV2-003).
- **E7 — Unicode bullet text (Korean / Japanese / Chinese)**: the
  anti-fabrication regex is locale-neutral — it matches only the forbidden
  patterns (SPEC IDs, REQ tokens, dates, SHAs) and does NOT reject CJK
  prose. (The CLAUDE.md block may carry localized workflow knowledge per
  the project's `code_comments` setting.)
- **E8 — Concurrent writer invocation**: the writer uses `O_APPEND|O_CREATE|
  O_WRONLY` single-write semantics for the lineage jsonl (inherited from
  EVOLVE-001 REQ-HEV-007); the CLAUDE.md / CLAUDE.local.md write path uses
  a tmp-file + atomic rename pattern to avoid partial writes under
  concurrent sessions. The Pre-Spawn Sync Check discipline mitigates the
  multi-session race at the orchestrator layer.
- **E9 — Snapshot directory absent at rollback time**: `RestoreSnapshot`
  returns a typed error (snapshot missing — manual recovery required);
  the lineage records the failed rollback attempt.
- **E10 — mergeSectionBased with conflicting upstream AND local marker
  changes**: surfaced as a standard merge conflict (REQ-HEV2-020), NOT
  auto-resolved.

## §F. Quality Gates

- **QG1 — Coverage**: `go test -cover ./internal/harness/curator/... ./internal/harness/...`
  ≥ 90% statement coverage on new + extended packages (REQ-HEV2-034).
- **QG2 — Lint**: `golangci-lint run --timeout=2m` exit 0; no NEW findings
  vs the pre-flight baseline (plan.md §C step 3).
- **QG3 — Subagent boundary**: AC-HEV2-044 grep yields 0 matches.
- **QG4 — Template neutrality**: `internal/template/internal_content_leak_test.go`
  + `.github/workflows/template-neutrality-check.yaml` PASS; no internal SPEC
  IDs / REQ tokens / dates / SHAs in the template block (AC-HEV2-038).
- **QG5 — Cross-platform build**: `GOOS=windows GOARCH=amd64 go build ./...`
  exit 0 (B1).
- **QG6 — spec-lint**: `moai spec lint .moai/specs/SPEC-HARNESS-EVOLVE-002/`
  exit 0 (no FrontmatterInvalid, no MissingExclusions, no EARS/GEARS violations);
  `moai spec lint --strict` exit 0 (no LegacyEARSKeyword in NEW SPEC).

## §G. Definition of Done

- [ ] All 36 REQ-HEV2-XXX in spec.md §C have at least one PASS AC.
- [ ] All MUST ACs in §A PASS.
- [ ] All 6 cross-file reachability ACs (AC-HEV2-047..052) + 1
      meta-traceability cross-check (AC-HEV2-053) PASS. (Iter-1 D5/D7
      reconciliation: the original "9" was an arithmetic error — the range
      047..053 is 7 ACs total; D7 relabeled 053 as a meta-traceability
      check, leaving 6 genuine reachability ACs 047..052.)
- [ ] All 9 Given-When-Then scenarios in §D PASS.
- [ ] All 10 edge cases in §E handled (test-covered or explicitly documented
      as out-of-scope for this SPEC's milestone).
- [ ] All 6 quality gates in §F PASS.
- [ ] The implementation commit history follows Conventional Commits with
      `feat(SPEC-HARNESS-EVOLVE-002): M{N} ...` subjects.
- [ ] The 3 NEEDS-CLARIFICATION items (H-1 marker naming, H-2 mergeSectionBased
      recognition, H-3 debug CLI verb scope) are resolved before
      Implementation Kickoff Approval OR explicitly deferred with rationale.
- [ ] The `progress.md` §E.2 / §E.3 are populated by manager-develop at
      run-phase (this agent authors only §E.1 at plan-phase).
- [ ] No PRESERVE-list file (plan.md §A.5) is modified.
- [ ] No working-tree-unrelated file is committed (specific-path `git add`
      discipline).
- [ ] EVOLVE-001's routing-ledger writer is untouched (read-only
      consumption only).
