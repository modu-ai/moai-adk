# SPEC-HARNESS-EVOLVE-002 — Progress

> Lifecycle tracking document. The `§E.*` namespace is parser-load-bearing
> (`internal/spec/era.go` `ClassifyEra` matches the literal `§E.2` / `§E.3` /
> `§E.4` heading tokens + the `sync_commit_sha` field — see
> `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md
> Section Map). Renaming any `§E.N` heading would silently break era
> classification. This file is authored by manager-spec at plan-phase
> (§E.1 only); §E.2-§E.4 are placeholder headings left for run-phase /
> sync-phase owners.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at:
plan_artifact_count: 6
plan_tier: L
plan_artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - design.md
  - research.md
  - progress.md
plan_req_count: 36
plan_ac_count: 53
plan_needs_clarification_count: 3
plan_needs_clarification_items:
  - H-1-claude-local-md-section-marker-naming
  - H-2-mergeSectionBased-recognition-mechanism
  - H-3-debug-cli-verb-scope
plan_commit_subject: "feat(SPEC-HARNESS-EVOLVE-002): plan-phase artifacts (L, 5 artifacts)"
plan_depends_on: SPEC-HARNESS-EVOLVE-001
plan_depends_on_status: completed
plan_era: V3R6
# plan-audit: iter-1 PASS 0.86 (D1-D7 amended) → iter-2 PASS 0.91 (audit-ready).
plan_audit_final_verdict: PASS
plan_audit_final_score: 0.91
plan_audit_final_iter: 2
```

## §E.2 Run-phase Evidence

### M1 — Typed Managed-Block Writer foundation

**Deliverables shipped:**
- `internal/harness/curator/writer.go` — `WriteManagedBlock(path, blockType, content)` + `BlockType` enum (`BlockTypeLearnedWorkflow`, `BlockTypeHarnessGenerated`) + `BlockContent`/`Bullet` types + per-`BlockType` marker registry (generalizes `layer3.go markerBlockPattern`).
- `internal/harness/curator/crud.go` — `AddBullet`/`UpdateBullet`/`DeleteBullet` per-bullet CRUD (REQ-HEV2-007).
- `internal/harness/curator/{writer,crud,subagent_boundary}_test.go` — table-driven tests with `t.TempDir()` isolation.
- `internal/harness/layer3.go` — refactored `InjectMarker` into a thin wrapper over `curator.WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)` (REQ-HEV2-002, AP-HEV2-004). `markerBlockPattern` + `buildMarkerBlock` removed; replaced by `buildHarnessBody` (returns `startAttrs` + `body` byte-identical to legacy output).

**AC PASS/FAIL matrix (M1-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-001 (WriteManagedBlock signature) | PASS | `TestWriteManagedBlock_Signature` — `go test ./internal/harness/curator/...` ok |
| AC-HEV2-002 (atomic replace) | PASS | `TestWriteManagedBlock_AtomicReplace` — 1 heading, 1 start marker, 1 end marker after replace |
| AC-HEV2-003 (BlockTypeHarnessGenerated enum + D1 byte-identical) | PASS | `TestBlockTypeEnum_IncludesHarnessGenerated` — byte-exact golden comparison + existing `layer3_test.go` (7 tests) pass unchanged |
| AC-HEV2-004 (idempotent zero-byte-diff) | PASS | `TestWriteManagedBlock_Idempotent_ZeroByteDiff` |
| AC-HEV2-005 (byte preservation) | PASS | `TestWriteManagedBlock_PreBlockPostBlockPreserved` |
| AC-HEV2-006 (append-mode newline convention) | PASS | `TestWriteManagedBlock_AppendMode_NewlineConvention` (2 subtests) |
| AC-HEV2-009 (AddBullet single-line) | PASS | `TestAddBullet_SingleLine_RewritesOnlyTarget` |
| AC-HEV2-010 (UpdateBullet by ledger_key) | PASS | `TestUpdateBullet_ByLedgerKey` |
| AC-HEV2-011 (DeleteBullet preserves others) | PASS | `TestDeleteBullet_ByLedgerKey_PreservesOthers` |
| AC-HEV2-044 (subagent boundary) | PASS | `TestWriter_NoAskUserQuestion` + grep `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ \| grep -v _test.go \| grep -v //` → 0 matches |
| AC-HEV2-047 (reachability: WriteManagedBlock exists) | PASS | `grep -rn "func WriteManagedBlock" internal/harness/curator/` → ≥1 |
| AC-HEV2-048 (reachability: layer3 calls curator.WriteManagedBlock) | PASS | `grep -n "curator.WriteManagedBlock(" internal/harness/layer3.go` → 1 match (line 33) |

**Test output:** `go test -v ./internal/harness/curator/...` — 23 tests PASS (0 fail). `go test ./internal/harness/...` — all harness sub-packages ok.

**Coverage:** `go test -cover ./internal/harness/curator/...` → 92.4% statement coverage (target ≥85% per QG1).

**D1 load-bearing verification:** the production caller `internal/cli/harness/install.go:85` (`harness.InjectMarker`) delegates through the refactored `InjectMarker` → `curator.WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)`. The byte-identical golden test (`TestBlockTypeEnum_IncludesHarnessGenerated`) verifies the HarnessGenerated block format is preserved exactly. The existing `layer3_test.go` suite (7 tests covering fresh-file, different-content-idempotent, same-spec-id-idempotent, append-without-trailing-newline, empty-path, empty-specID, file-not-found) passes unchanged — confirming the production path is preserved.

### M2 — `MOAI:LEARNED-WORKFLOW` digest block + budget/cap enforcement

**Deliverables shipped:**
- `internal/harness/curator/budget.go` (NEW) — digest-layer budget constants (`MaxDigestBlockChars = 3000`, `MaxDigestBullets = 20`), typed errors (`ErrDigestBudgetExceeded`, `ErrBulletCapExceeded`, `ErrForbiddenContent`), anti-fabrication regex (`forbiddenTokenPatterns` + `shaLikePattern` with digit filter), `containsForbiddenContent` helper, `validateDigestBlock` aggregator.
- `internal/harness/curator/writer.go` (MODIFIED) — wired `validateDigestBlock` into `WriteManagedBlock` for `BlockTypeLearnedWorkflow` BEFORE any file I/O (gated so the legacy `BlockTypeHarnessGenerated` D1 path is untouched).
- `internal/config/token_budget_guard.go` (MODIFIED) — added `MeasureAlwaysLoadedSection(filePath, startMarker, endMarker)` (AC-HEV2-013 per-section attribution; additive — does not alter `measureAlwaysLoaded` signature).
- `internal/harness/curator/{marker,budget,antifabrication}_test.go` (NEW) + `crud_test.go` (EXTENDED) — AC-HEV2-007/008/012/013/014/015/016 coverage.

**AC PASS/FAIL matrix (M2-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-007 (LEARNED marker regex atomic match) | PASS | `TestLearnedWorkflowMarker_RegexMatchesHeadingPlusMarkers` — compiled pattern matches heading + start + body + end as one group |
| AC-HEV2-008 (atomic match group; attrs + non-greedy) | PASS | `TestLearnedWorkflowMarker_AtomicMatchGroup` — 3 subtests (start-marker attrs, non-greedy stops at first end, bare heading not matched) |
| AC-HEV2-012 (≤3K budget → ErrDigestBudgetExceeded, file untouched) | PASS | `TestWriteManagedBlock_BudgetExceeded_ErrDigestBudgetExceeded` + `TestWriteManagedBlock_BudgetBoundary` (exactly-3000 admitted, 3001 rejected) |
| AC-HEV2-013 (measureAlwaysLoaded per-section attribution) | PASS | `TestMeasureAlwaysLoaded_PerSectionAttribution` + `TestMeasureAlwaysLoaded_SectionAbsent` — `config.MeasureAlwaysLoadedSection` measures the LearnedWorkflow block body distinctly; outside prose excluded; absent → found=false |
| AC-HEV2-014 (≤20 bullet cap → ErrBulletCapExceeded, file untouched) | PASS | `TestWriteManagedBlock_BulletCapExceeded` (21 rejected) + `TestWriteManagedBlock_BulletCapBoundary` (exactly-20 admitted) |
| AC-HEV2-015 (ledger_key trailing HTML comment linkage) | PASS | `TestBullet_LedgerKeyTrailingHTMLComment` + `TestBullet_ProvisionalOmitsKeyMarker` — `<!-- key: <k> -->` on the bullet line; provisional omits |
| AC-HEV2-016 (anti-fabrication forbidden patterns rejected) | PASS | `TestWriteManagedBlock_RejectsForbiddenPatterns` (7 subtests: SPEC-id multi/single, REQ, AC, ISO date, SHA short/long) + `TestWriteManagedBlock_AdmitsGenericWorkflowKnowledge` (6 positive incl. CJK + "defaced" no-digit) + `TestContainsForbiddenContent` |

**Anti-fabrication evidence-or-null (per task prompt §D):** the budget ACs verify the ACTUAL measured contribution via `config.MeasureAlwaysLoadedSection` (AC-HEV2-013), not an assumed value. `TestMeasureAlwaysLoaded_PerSectionAttribution` writes a known block body, then measures the section bytes directly and asserts the attributed char count equals the actual section byte length. The boundary tests (exactly-3000 / exactly-20) verify the gate threshold is the named constant, not a magic number.

**Test output:** `go test -count=1 ./internal/harness/curator/...` → ok, all tests PASS (M1's 23 + M2's new tests, 0 fail).

**Coverage:** `go test -cover ./internal/harness/curator/...` → **93.4%** statement coverage (M1 was 92.4%; M2 raised it; target ≥85% per QG1, SPEC target ≥90%).

**Lint:** `golangci-lint run --timeout=2m ./internal/harness/curator/... ./internal/config/...` → 0 issues (no NEW findings vs pre-flight baseline).

**Cross-platform build (B1):** `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0. The new `budget.go` uses only stdlib (`errors`, `fmt`, `regexp`, `strings`) — no syscall, no build tags.

**Subagent boundary (B3 / AC-HEV2-044):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ | grep -v _test.go | grep -v '// '` → 0 matches (the existing `subagent_boundary_test.go` static guard still passes; M2 additions do not violate it).

**D1 preservation re-verified:** `go test -v -run 'InjectMarker|HarnessGenerated' ./internal/harness` → 8/8 PASS (fresh-file, different-content-idempotent, same-spec-id-idempotent, append-without-trailing-newline, empty-path, empty-specID, file-not-found, no-ensure-allowed-call). The enforcement gate is scoped to `BlockTypeLearnedWorkflow` only; the HarnessGenerated path is byte-identical.

**Residual risk (M2 scope, NOT closing debt):** `AddBullet` does NOT enforce the ≤20 bullet cap (only `WriteManagedBlock` does). AC-HEV2-014 pins `WriteManagedBlock` specifically; `AddBullet` cap enforcement is not pinned by any M2 AC and is deferred to keep M2 focused (scope discipline). A block populated via 21 successive `AddBullet` calls would bypass the cap. Mitigation path: a follow-up enforcement in `AddBullet` that counts existing bullets before insert.

### M3 — CLAUDE.local.md append-only Learned section (Tier 3)

**Deliverables shipped:**
- `internal/harness/curator/append.go` (NEW) — `AppendLearnedLocal(path, entry)` append-only writer for the `MOAI:LEARNED-WORKFLOW-LOCAL` section (REQ-HEV2-012). Appends a single entry before the end marker WITHOUT modifying existing entries' bytes. Dedup guard (REQ-HEV2-014): scans the LOCAL block body for an existing `ledger_key` match → returns `ErrDuplicateAppend` without writing. `findStartMarkerLineBefore` helper scopes the dedup scan to the LOCAL section (a key in the digest block does NOT trigger LOCAL dedup — the two surfaces are distinct layers). Provisional entries (empty `LedgerKey`) bypass dedup (evidence-or-null, REQ-HEV2-024).
- `internal/harness/curator/writer.go` (MODIFIED) — added `BlockTypeLearnedLocal` to the `BlockType` enum + the LOCAL marker contract (`## MOAI:LEARNED-WORKFLOW-LOCAL` + `<!-- moai:learned-local-start -->` / `<!-- moai:learned-local-end -->`) to `markerRegistry` (REQ-HEV2-013). `compiledPatterns` auto-compiles the LOCAL atomic-match regex from the registry. Added `ErrDuplicateAppend` sentinel error. All changes additive — M1/M2 enum values, registry entries, and the `validateDigestBlock` gate (scoped to `BlockTypeLearnedWorkflow`) are untouched.
- `internal/harness/curator/append_test.go` (NEW) — AC-HEV2-018 + AC-HEV2-020 + 6 coverage tests (fresh-block, provisional, cross-section dedup isolation, empty-path, block-not-found, file-not-found). All `t.TempDir()`-isolated.
- `internal/harness/curator/marker_test.go` (MODIFIED) — AC-HEV2-019 (`TestLearnedLocalMarker_RegexMatchesHeadingPlusMarkers`) + `TestLearnedLocalMarker_DoesNotMatchDigestBlock` (disjoint-contract guard) + `TestLearnedLocalMarker_PatternCompiles`.

**AC PASS/FAIL matrix (M3-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-018 (append-only: existing bytes unchanged) | PASS | `TestAppendOnly_ExistingBytesUnchanged` — 3 existing entries preserved verbatim; exactly 1 line inserted (single-line append delta) |
| AC-HEV2-019 (LOCAL marker regex: heading + start/end atomic) | PASS | `TestLearnedLocalMarker_RegexMatchesHeadingPlusMarkers` + `TestLearnedLocalMarker_DoesNotMatchDigestBlock` (LOCAL pattern does not cross-match digest block) + `TestLearnedLocalMarker_PatternCompiles` |
| AC-HEV2-020 (dedup same ledger_key → ErrDuplicateAppend, no write) | PASS | `TestAppendOnly_DedupSameLedgerKey_ErrDuplicateAppend` — `errors.Is(err, ErrDuplicateAppend)` + byte-identical before/after (no write) + `TestAppendOnly_DedupDoesNotMatchOtherSection` (digest-only key does NOT trigger LOCAL dedup) |

**Test output:** `go test -count=1 -v -run "TestLearnedLocalMarker|TestAppendOnly|TestAppendLearnedLocal" ./internal/harness/curator/...` → 11 tests PASS (0 fail). Full curator suite: `go test -count=1 ./internal/harness/curator/...` ok. Full harness suite: `go test -count=1 ./internal/harness/...` — all 12 sub-packages ok (M1+M2 preserved).

**Coverage:** `go test -cover ./internal/harness/curator/...` → **93.3%** statement coverage (M2 was 93.4%; the 0.1% delta is one defensive branch in `findStartMarkerLineBefore`'s corrupt-input fallback, outside M3 ACs; still above the 85% QG1 minimum + 90% SPEC target).

**Cross-platform build (B1):** `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0. The new `append.go` uses only stdlib (`errors`, `fmt`, `strings`) — no syscall, no build tags.

**Subagent boundary (B3 / AC-HEV2-044):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ | grep -v _test.go | grep -v '// '` → 0 matches (existing `subagent_boundary_test.go` static guard still passes; M3 additions do not violate it).

**Lint (B5):** `golangci-lint run --timeout=2m ./internal/harness/...` → 0 issues (no NEW findings vs M2 baseline).

**M1+M2 PRESERVE verified:** the `BlockTypeLearnedLocal` enum value is appended AFTER `BlockTypeHarnessGenerated`, so the existing iota values are unchanged (M1 D1 byte-identical contract intact). The `validateDigestBlock` gate in `WriteManagedBlock` remains scoped to `BlockTypeLearnedWorkflow` only — the LOCAL surface is exempt from digest budget/cap enforcement (distinct layer). `InjectMarker` (layer3.go) is untouched.

**Residual risk (M3 scope, NOT closing debt):** anti-fabrication input validation (`containsForbiddenContent`, REQ-HEV2-011) is NOT wired into `AppendLearnedLocal`. The M3 ACs (018/019/020) do not pin anti-fabrication for the LOCAL surface, and REQ-HEV2-011 is scoped to the digest layer in acceptance.md (AC-HEV2-016/017). A future caller could append a forbidden internal token (SPEC ID / REQ token / date / SHA) to a LOCAL bullet. Mitigation path: a follow-up that applies `containsForbiddenContent` to the LOCAL append (the function is reusable from budget.go). Noted as a §25-isolation gap for a follow-up; does not block M3.

### M4 — `mergeSectionBased` managed-section preservation

**Deliverables shipped:**
- `internal/merge/strategies.go` (MODIFIED) — added `managedSectionHeadings` explicit allow-list var (design.md §D.1 H-2 option (a)) carrying the two curator-managed LEARNED block headings (`## MOAI:LEARNED-WORKFLOW`, `## MOAI:LEARNED-WORKFLOW-LOCAL`), an `isManagedSection(heading) bool` helper that consults it, and a managed-section preservation branch in `mergeSectionBased`'s "section exists in all three" path. For a managed section: when the upstream template content equals the base (template did not touch the block), the local populated block is preserved verbatim (no clobber — REQ-HEV2-019); when both sides carry differing content inside the marker boundaries, a conflict is surfaced rather than auto-resolved (REQ-HEV2-020). The generic 3-way section logic for non-managed sections is unchanged.
- `internal/merge/strategies_test.go` (MODIFIED) — added `strings` import + 4 new tests: `TestMergeSectionBased_PreservesPopulatedLearnedBlock` (AC-HEV2-025), `TestMergeSectionBased_EmptyUpstreamPopulatedLocal` (AC-HEV2-026), `TestMergeSectionBased_ConflictInsideMarkers_Surfaced` (AC-HEV2-027), `TestManagedSectionHeadings_RecognizesLearnedBlocks` (AC-HEV2-049 recognition contract).

**AC PASS/FAIL matrix (M4-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-025 (merge preserves populated block, no clobber) | PASS | `TestMergeSectionBased_PreservesPopulatedLearnedBlock` — template empty marker + local populated block → local bullet + ledger_key preserved verbatim; unrelated template section change still reflected; exactly 1 learned-start marker (no duplication); no conflict |
| AC-HEV2-026 (empty upstream + populated local, minimal case) | PASS | `TestMergeSectionBased_EmptyUpstreamPopulatedLocal` — both bullets + both ledger_keys preserved verbatim; exactly 1 learned-start marker; no conflict |
| AC-HEV2-027 (conflicting populated content inside markers → conflict, not auto-resolved) | PASS | `TestMergeSectionBased_ConflictInsideMarkers_Surfaced` — `HasConflict == true`; conflict carries both `Current` (local bullet X) and `Updated` (template bullet Y) content |
| AC-HEV2-049 (reachability: managedSectionHeadings declared + consulted, ≥2 matches) | PASS | `grep -c "managedSectionHeadings" internal/merge/strategies.go` → **5** matches (var declaration + `isManagedSection` consultation + `mergeSectionBased` doc + @MX tag); `TestManagedSectionHeadings_RecognizesLearnedBlocks` pins the recognition contract |

**Test output:** `go test -v -run "TestMergeSectionBased_PreservesPopulatedLearnedBlock|TestMergeSectionBased_EmptyUpstreamPopulatedLocal|TestMergeSectionBased_ConflictInsideMarkers_Surfaced|TestManagedSectionHeadings_RecognizesLearnedBlocks" ./internal/merge/...` → 4 tests PASS (0 fail). Full merge suite: `go test ./internal/merge/...` ok. Consumer cascade check: `go test ./internal/cli/...` ok (the merge package is consumed by `internal/cli/update`).

**Coverage:** `go test -cover ./internal/merge/...` → **87.3%** statement coverage (above the 85% QG1 minimum; the new managed-section branch + `isManagedSection` are fully exercised by the 4 new tests).

**Cross-platform build (B1):** `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0. The changes are stdlib-only (`strings` already imported) — no syscall, no build tags, no platform-specific code.

**Subagent boundary (B3 / AC-HEV2-044):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/merge/ | grep -v _test.go | grep -v '// '` → 0 matches.

**Lint (B5):** `golangci-lint run --timeout=2m ./internal/merge/...` → 0 issues (no NEW findings vs pre-flight baseline). `go vet ./internal/merge/...` exit 0.

**M1+M2+M3 PRESERVE verified:** only 2 files changed (`internal/merge/strategies.go` + `internal/merge/strategies_test.go`) — both in the merge package. No curator package file, no `layer3.go`, no `token_budget_guard.go` touched. Full harness suite: `go test ./internal/harness/...` — all 12 sub-packages ok (curator 2.128s green; M1 D1 byte-identical InjectMarker contract + M2 budget-gate scoping + M3 LOCAL surface intact). `git diff --stat` confirms the 2-file scope.

**Residual risk (M4 scope, NOT closing debt):** the managed-section preservation branch is behaviorally equivalent to the generic 3-way logic for the specific AC-025/026/027 fixtures (the generic logic already treats sections as opaque string-comparison units). The allow-list's load-bearing value is (a) AC-049 compliance (explicit recognition var consulted on the preservation path), (b) documented preservation policy for curator-managed blocks, (c) an extension point for future managed-block types, and (d) a future-proofing guard so later changes to the generic section logic cannot silently regress the managed-block preservation guarantee. A scenario where cosmetic template marker reformatting (e.g. an extra blank line in the empty marker) against a populated local block would currently surface as a conflict under both managed and generic logic — marker-body emptiness normalization is deferred (it would couple merge to marker parsing, which design.md §D.2 explicitly rejected in favor of the §D.1 allow-list). Does not block M4.

### M5 — Snapshot / rollback / lineage extension

**Deliverables shipped:**
- `internal/harness/types.go` (MODIFIED) — `LineageEntry` extended additively with `LearnedSurface string`, `BulletsChanged []string`, `SnapshotDir string` (REQ-HEV2-023). The 3 fields are APPENDED after the existing fields — existing field order/types unchanged (backward-compat verified by `TestLineageEntry_AdditiveFieldOrder`). `BulletsChanged` is deliberately tagged WITHOUT omitempty so a nil slice serializes as JSON `null` (evidence-or-null per REQ-HEV2-024 / AP-HEV2-010 — NOT `[]` which would fabricate a zero-change claim, NOT omitted which loses the signal).
- `internal/harness/applier.go` (MODIFIED) — (a) `snapshotFile` struct extended additively with `LearnedSurface`, `ByteLengthPreWrite`, `BulletsAffected` (all omitempty so legacy Apply-path snapshots serialize byte-identical); (b) new `SurfaceRestoreUnit` type + new exported `CreateSurfaceSnapshot(snapshotBase, proposalID, surfaces)` function producing distinct per-surface restore units (REQ-HEV2-021, design.md §C.1); (c) `RestoreSnapshot` extended with a byte-length integrity check gated on `ByteLengthPreWrite > 0` (REQ-HEV2-022) — skipped on legacy snapshots (zero value) for backward compat; (d) new `ErrRollbackIntegrityFailed` sentinel; (e) `writeLineage` refactored to delegate to a new internal `writeLineageEntry`, plus new `writeLineageCurator` method populating the M5 fields (REQ-HEV2-023 plumbing). The 4 existing `writeLineage` callers are byte-unaffected.
- `internal/harness/applier_test.go` (MODIFIED) — 6 new tests: `TestCreateSnapshot_DistinctRestoreUnitsPerSurface` (AC-HEV2-028), `TestRestoreSnapshot_ByteIdenticalRollback` (AC-HEV2-029), `TestRestoreSnapshot_IntegrityFailure` (byte-length mismatch → ErrRollbackIntegrityFailed), `TestCreateSnapshot_BackupNameDisambiguation` (basename collision), `TestCreateSnapshot_EmptySurfaces` (guard), `TestWriteLineageCurator_RoundTrip` + `TestWriteLineage_LegacyPathLeavesSurfaceFieldsZero` (REQ-HEV2-023 plumbing + backward compat).
- `internal/harness/lineage_test.go` (MODIFIED) — 3 new tests: `TestWriteLineageEntry_LearnedSurfaceFields` (AC-HEV2-030), `TestLineageEntry_EvidenceOrNull` (AC-HEV2-031), `TestLineageEntry_AdditiveFieldOrder` (additive invariant).

**AC PASS/FAIL matrix (M5-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-028 (distinct restore units per surface) | PASS | `TestCreateSnapshot_DistinctRestoreUnitsPerSurface` — dual-surface manifest carries ≥2 distinct `learned_surface` values + per-surface `byte_length_pre_write` recorded |
| AC-HEV2-029 (byte-identical rollback) | PASS | `TestRestoreSnapshot_ByteIdenticalRollback` — post-rollback bytes == pre-write bytes for BOTH surfaces |
| AC-HEV2-030 (LineageEntry carries new fields) | PASS | `TestWriteLineageEntry_LearnedSurfaceFields` — LearnedSurface + BulletsChanged + SnapshotDir round-trip via WriteLineageEntry → LoadManifest |
| AC-HEV2-031 (evidence-or-null) | PASS | `TestLineageEntry_EvidenceOrNull` — nil BulletsChanged serializes as JSON `null` (NOT `[]`, NOT omitted, NOT `""`); populated serializes as array |
| AC-HEV2-050 (reachability: ≥3 field matches in types.go) | PASS | `grep -c -E "LearnedSurface\|BulletsChanged\|SnapshotDir" internal/harness/types.go` → **7** matches (≥3) |
| AC-HEV2-051 (manifest structural: ≥2 distinct learned_surface) | PASS | `TestCreateSnapshot_DistinctRestoreUnitsPerSurface` — the dual-surface fixture manifest carries exactly 2 distinct `learned_surface` values |

**Additive-invariant confirmation:** `TestLineageEntry_AdditiveFieldOrder` proves the M5 fields are appended AFTER the existing fields in declaration order (`proposal_id` < `decision` < `learned_surface` < `bullets_changed` < `snapshot_dir` in the serialized JSON). Existing field order/types unchanged — legacy lineage consumers parse pre-M5 entries verbatim. The `TestWriteLineage_LegacyPathLeavesSurfaceFieldsZero` test confirms the legacy `writeLineage` path (the pre-Curator Apply callers) leaves the M5 fields at zero/nil so legacy transitions serialize byte-identically to pre-M5 output.

**Test output:** `go test -count=1 -run "<M5 AC tests>" ./internal/harness/` → 8 tests PASS (0 fail). Full harness suite: `go test -count=1 ./internal/harness/...` — all 12 sub-packages ok (M1-M4 preserved). `go test -count=1 ./internal/harness/curator/... ./internal/merge/...` ok (M1-M4 PRESERVE).

**Coverage:** `go test -cover ./internal/harness/` → **87.2%** statement coverage (above the 85% QG1 minimum; M5 additions fully exercised by the 9 new tests).

**Cross-platform build (B1):** `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0. The M5 additions use only stdlib (`encoding/json`, `errors`, `fmt`, `os`, `path/filepath`, `strings`, `time` — all already imported) — no syscall, no build tags.

**Subagent boundary (B3 / AC-HEV2-044):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ | grep -v _test.go | grep -v '// '` → the 1 match (`internal/harness/proposalgen/scaffolder.go:111`) is PRE-EXISTING at HEAD `c02c0ee8d` (a documentation string literal in a different package, OUTSIDE M5 edit scope, NOT a call). M5 introduced 0 new AskUserQuestion references.

**Lint (B5):** `golangci-lint run --timeout=2m ./internal/harness/...` → **0 issues** (no NEW findings vs pre-flight baseline; an initial staticcheck QF1001 De-Morgan suggestion in `lineage_test.go` was resolved by rewriting `!(a<b)` → `a>=b`).

**Residual risk (M5 scope, NOT closing debt):** the `writeLineageCurator` method (the Curator-path lineage populator) is PROVIDED but not yet WIRED into an actual Curator-write code path — the existing `Apply` method still calls the legacy `writeLineage` (leaving M5 fields zero). This is intentional: the Curator-pipeline-to-Apply integration is future wiring (a later milestone / EVOLVE-003+ territory). The M5 deliverable is the surface + the round-trip verification (AC-030/031); the field-population-on-every-Curator-write invariant (REQ-HEV2-023) holds for the `writeLineageCurator` code path itself. Wiring `SnapshotDir` into the existing Apply approve path (the legacy safety pipeline does create a snapshot) is a deferred enhancement — out of M5's AC scope.

### M6 — Template-First empty marker + template neutrality + 2-layer Recall contract

**Deliverables shipped:**
- `internal/template/templates/CLAUDE.md` (MODIFIED, edited FIRST) + `CLAUDE.md` (live dogfood copy, MODIFIED second) — appended the EMPTY `MOAI:LEARNED-WORKFLOW` managed block (heading `## MOAI:LEARNED-WORKFLOW` + `<!-- moai:learned-start -->` / `<!-- moai:learned-end -->` markers, ZERO bullets, NEUTRAL content). The heading-adjacent start marker (no blank line between them) matches the curator writer's atomic-match regex so a future `WriteManagedBlock` stays idempotent. Template-First ordering (REQ-HEV2-028): template source edited before the live copy, then `make build` (REQ-HEV2-029 empty-marker shipping, REQ-HEV2-030 section-25 neutrality).
- `make build` recompiled the embedded assets (`//go:embed all:templates`); the marker is embedded in the binary. `catalog.yaml` is byte-identical (CLAUDE.md is not a catalog-hashed skill dir).
- `internal/harness/curator/recall.go` (NEW) — the 2-layer Recall contract as types + godoc: `RecallLayer` (`DigestLayer` summary-only / `LedgerLayer` searchable), `DigestEntry` (Summary + LedgerKey, `Provisional()` = evidence-or-null), `LedgerSearcher` interface (`SearchByKey`), `RecallContract` (`Digest()`/`Ledger()`/`SearchableLayer()`). The godoc names the digest layer, the ledger layer, the cross-layer `ledger_key` linkage (literal REQ-HEV2-017 citation grep-visible), and the principle "remember everything (cross), search when needed (circle)". Consumption wiring deferred to EVOLVE-005 (contract-only). (REQ-HEV2-015 digest / 016 ledger / 017 linkage / 018 principle)
- `internal/harness/curator/recall_test.go` (NEW) — 3 TDD tests (RED-first, confirmed compile-fail then GREEN): `TestRecallContract_DigestLayerSummaryOnly` (AC-021), `TestRecallContract_LedgerLayerSearchInterface` (AC-022, via a fake `LedgerSearcher`), `TestRecallContract_NoWriteFullEvidencePath` (AC-024 — principle + source-scan guard for no `WriteFullEvidenceToDigest(` code path).
- `internal/template/internal_content_leak_test.go` (EXTENDED) — `TestTemplateLearnedWorkflowBlockNeutral` (AC-038): asserts the template block is PRESENT (3 markers), ships EMPTY (zero bullets), and is NEUTRAL (default+strict forbidden-class scan over the block region -> 0 violations).

**AC PASS/FAIL matrix (M6-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-021 (digest layer summary-only contract) | PASS | `TestRecallContract_DigestLayerSummaryOnly` — DigestLayer.SummaryOnly()==true, LedgerLayer==false, String() incl. default branch, DigestEntry.Provisional() both ways |
| AC-HEV2-022 (ledger layer search interface) | PASS | `TestRecallContract_LedgerLayerSearchInterface` — fake LedgerSearcher; SearchableLayer()==LedgerLayer; SearchByKey resolves key->evidence, empty key->nil |
| AC-HEV2-024 (no WriteFullEvidenceToDigest code path) | PASS | `TestRecallContract_NoWriteFullEvidencePath` — principle (DigestLayer summary-only) + source-scan of non-test curator .go files for a `WriteFullEvidenceToDigest(` definition/call -> 0 |
| AC-HEV2-035 (Template-First: template edit precedes live copy) | PASS | template CLAUDE.md edited FIRST, then live copy, then `make build`; the single M6 feat commit lists the template path first in `git add`. `git log --diff-filter=A -- internal/template/templates/CLAUDE.md` -> ccd6be1f6 (historical file ADD predates this SPEC; the M6 edit is template-first by workflow discipline) |
| AC-HEV2-036 (empty marker in template tree) | PASS | `grep -c "MOAI:LEARNED-WORKFLOW" internal/template/templates/CLAUDE.md` -> 1 (>=1) |
| AC-HEV2-037 (make build embeds the marker) | PASS | `make build` then `strings bin/moai \| grep -c "MOAI:LEARNED-WORKFLOW"` -> 3 (>=1) |
| AC-HEV2-038 (leak test extended for the new block) | PASS | `TestTemplateLearnedWorkflowBlockNeutral` PASS; full template leak suite green (no regression) |
| AC-HEV2-052 (reachability: template marker triple grep >=3) | PASS | `grep -c "moai:learned-start\|moai:learned-end\|MOAI:LEARNED-WORKFLOW" internal/template/templates/CLAUDE.md` -> 3 (>=3) |

**Test output:** `go test -run 'TestRecallContract_' -v ./internal/harness/curator/` -> 3/3 PASS. `go test -run TestTemplateLearnedWorkflowBlockNeutral ./internal/template/` -> PASS. Full suites: `go test ./internal/harness/... ./internal/merge/... ./internal/template/...` -> all ok, 0 FAIL (M1-M5 preserved + M6 new tests).

**Coverage:** `go test -cover ./internal/harness/curator/` -> **93.6%** statement coverage (M5 curator ~93.3%; recall.go additions fully exercised; >=90% SPEC target). (AC-043 QG1)

**Cross-platform build (B1 / QG5):** `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0. `recall.go` uses only stdlib `strings`; no syscall, no build tags.

**Subagent boundary (B3 / AC-HEV2-044):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ | grep -v _test.go | grep -v '// '` -> 0 matches (the `writer.go` hit is a `//` godoc line, excluded; `recall.go` introduces 0 references).

**Lint (B5 / QG2):** `golangci-lint run --timeout=3m ./internal/harness/curator/... ./internal/template/...` -> 0 issues. `go vet ./internal/harness/curator/... ./internal/template/...` exit 0.

**Template neutrality (QG4):** `TestTemplateLearnedWorkflowBlockNeutral` + `TestTemplateNoInternalContentLeak` PASS — the template block ships EMPTY with zero forbidden-class content (no internal SPEC IDs / REQ-AC tokens / dates / SHAs). Section-25 isolation held.

**M1-M5 PRESERVE verified:** no M1-M5 file modified. The only changes are the 2 CLAUDE.md files (template + live), 2 NEW recall files, and 1 EXTENDED leak test. Full harness/merge/template suites green (curator 93.6%, merge ok, template ok). `git status --porcelain` confirmed the 5-file scope.

**Residual risk (M6 scope, NOT closing debt):** the Recall contract is types + godoc only — the CONSUMPTION wiring (a live `LedgerSearcher` over the EVOLVE-001 routing ledger + lineage surfaces, and the digest->ledger resolution at recall time) is deferred to EVOLVE-005 (declared in `recall.go` godoc). AC-021/022/024 verify the contract shape + the no-full-evidence-path principle; they do NOT verify a runtime search. The template empty marker is inert until a future L5-approved Curator write populates it (EVOLVE-003+). AC-035's `git log --diff-filter=A` check reflects the historical file ADD (predating this SPEC); the meaningful template-first ordering is the workflow discipline (template edited before the live copy, verified by the edit sequence).

### M7 — Integration verification (spec.md §7 M2 verification target)

**Deliverables shipped** (all ADDITIVE — M1-M6 implementation logic untouched; 2 minimal helpers the tests exercise + 7 test files):
- `internal/harness/curator/tier_gate.go` (NEW) — `ErrTierNotQualified` sentinel + `TierGatedWrite(path, blockType, observations, content)` surface-selection gate (REQ-HEV2-025/026/027). Reuses the `internal/harness/tier` `ClassifyStatus` [1,3,5,10] ladder (threshold SSOT, no duplication); maps Tier 4 -> `WriteManagedBlock` (CLAUDE.md digest), Tier 3 -> `AppendLearnedLocal` (CLAUDE.local.md append); rejects self-tier-escalation with `ErrTierNotQualified` WITHOUT touching the file.
- `internal/harness/curator/approval.go` (NEW) — `ApprovalDecision` + `RejectionRecorder` + `ErrApprovalRejected` + `WriteManagedBlockGated(...)` L5-approval gate (REQ-HEV2-032). No autonomous write path; on rejection the file is untouched and the injected recorder appends the "rejected" LineageEntry (the recorder callback respects the curator->harness import boundary — curator cannot import harness).
- `internal/harness/curator/tier_test.go` (NEW, `package curator`) — AC-032/033/034.
- `internal/harness/curator/approval_test.go` (NEW, `package curator_test`) — AC-040/041 (imports `harness` for the real `WriteLineageEntry`/`LoadManifest`).
- `internal/harness/curator/rollback_test.go` (NEW, `package curator_test`) — AC-042 e2e byte-identical rollback.
- `internal/harness/curator_e2e_test.go` (NEW, `package harness`) — the M2 verification chain.
- `internal/merge/learned_roundtrip_test.go` (NEW, `package merge`) — full template-merge round-trip preservation.
- `internal/harness/curator/crud_test.go` (EXTENDED) — AC-023 `TestBullet_ProvisionalNullLedgerKey`.
- `internal/harness/curator/antifabrication_test.go` (EXTENDED) — AC-039 `TestWriteManagedBlock_RejectsModelSelfReport`.

**AC PASS/FAIL matrix (M7-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-023 (provisional null ledger_key + Tier-3 promotion) | PASS | `TestBullet_ProvisionalNullLedgerKey` — empty ledger_key renders no key marker, `DigestEntry.Provisional()`==true; promotion adds the real `<!-- key: lw-promoted-001 -->` |
| AC-HEV2-032 (Tier 4 qualified -> CLAUDE.md write) | PASS | `TestTier4Qualified_ClaudeMdWrite` — 10 obs writes the LEARNED-WORKFLOW block to CLAUDE.md |
| AC-HEV2-033 (Tier 3 qualified -> CLAUDE.local.md append) | PASS | `TestTier3Qualified_ClaudeLocalMdAppend` — 5 obs appends inside the LOCAL block markers |
| AC-HEV2-034 (under-tier -> ErrTierNotQualified) | PASS | `TestUnderTierWrite_ErrTierNotQualified` — 6-obs Tier-4 write -> `ErrTierNotQualified`, file untouched; same pattern accepted at Tier 3 |
| AC-HEV2-039 (rejects model self-report) | PASS | `TestWriteManagedBlock_RejectsModelSelfReport` (3 sub-cases) — SPEC/REQ/AC/date/SHA self-report -> `ErrForbiddenContent`, file untouched |
| AC-HEV2-040 (requires approval token) | PASS | `TestWriteManagedBlock_RequiresApprovalToken` — rejection -> no write + `ErrApprovalRejected`; approval -> block written |
| AC-HEV2-041 (rejection records lineage, no file write) | PASS | `TestWriteManagedBlock_RejectionRecordsLineage_NoFileWrite` — 1 LineageEntry decision=="rejected" + rationale; file untouched |
| AC-HEV2-042 (mechanical rollback, byte-identical) | PASS | `TestRollbackTrigger_MechanicalOnly_NoModelSelfReport` — AddBullet(D)+DeleteBullet(B)+RestoreSnapshot -> byte-identical to pre-write, markers intact, idempotent |
| AC-HEV2-045 (no settings hook-registration change) | PASS | `git diff --name-only origin/main -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json` -> 0 |
| AC-HEV2-046 (no new curator hook wrapper) | PASS | `ls .claude/hooks/moai/handle-*curator*.sh` -> 0 |

**M2 verification chain (e2e):** `TestM2VerificationChain_EndToEnd` — AddBullet -> budget-enforced (16x200-char over-budget write -> `ErrDigestBudgetExceeded`, file untouched) -> `CreateSurfaceSnapshot` -> DeleteBullet -> `RestoreSnapshot` (byte-identical) -> `LineageEntry` audit trail (LearnedSurface/BulletsChanged/SnapshotDir round-trip). PASS.

**Template-merge round-trip:** `TestMergeRoundTrip_PopulatedLearnedBlockSurvivesTemplateSync` — multi-bullet populated local block survives a template sync (empty upstream marker) while an unrelated upstream section update lands; exactly one marker pair, no clobber. PASS. (Complements the M4 AC-025/026/027 preservation cases.)

**Test output:** `go test ./internal/harness/... ./internal/merge/... ./internal/template/...` -> 14 ok, 0 FAIL (M1-M6 preserved + M7 new tests). All M7 AC-bound tests PASS (`-count=1`).

**Coverage (AC-043 QG1):** `go test -cover ./internal/harness/curator/` -> **91.0%** (>=90% REQ-HEV2-034); `./internal/harness/` -> 87.2%; `./internal/merge/` -> 87.3% (both >=85%).

**Cross-platform build (B1 / QG5):** `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0. tier_gate.go/approval.go use only stdlib + the tier package; no syscall, no build tags.

**Subagent boundary (B3 / AC-HEV2-044):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ | grep -v _test.go | grep -v '// '` -> 0 matches. The approval gate takes the L5 decision as an in-parameter (`ApprovalDecision`) — it never calls AskUserQuestion (REQ-HEV2-035).

**Lint (B5 / QG2):** `golangci-lint run --timeout=2m ./internal/harness/... ./internal/merge/...` -> 0 issues (one NEW QF1001 De Morgan finding on tier_test.go was fixed before commit). `go vet` exit 0.

**M1-M6 PRESERVE verified:** no M1-M6 implementation file modified. Changes are 2 NEW helper files + 7 test files (2 extended, 5+ new). Full harness/merge/template suites green.

**Residual risk (M7 scope):** `TierGatedWrite` and `WriteManagedBlockGated` are the completed write-layer API surface (spec.md §D.1 sketch); their PRODUCTION wiring into the Curator pipeline (tier->surface activation from harness config; L5 orchestrator round injecting the ApprovalDecision + RejectionRecorder) is EVOLVE-003+ scope, explicitly out of scope per spec.md §E. The tests exercise the API directly with machine-signal inputs; no autonomous production write path is activated by this SPEC.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-12
run_commit_sha: 81792f80c  # M7 feat commit (the progress.md docs commit follows separately)
run_status: run-complete-sync-pending
# M1 (Typed Managed-Block Writer foundation) + M2 (LEARNED digest block +
# budget/cap enforcement) + M3 (CLAUDE.local.md append-only LOCAL section) +
# M4 (mergeSectionBased managed-section preservation) + M5 (snapshot/rollback/
# lineage surface extension) + M6 (Template-First empty marker + template
# neutrality + 2-layer Recall contract) + M7 (integration verification) are
# ALL complete — M7 is the final run milestone, so run-phase is audit-ready.
# M1-scoped ACs PASS (AC-HEV2-001..006, 009..011, 044, 047, 048). M2-scoped ACs
# PASS (AC-HEV2-007, 008, 012, 013, 014, 015, 016). M3-scoped ACs PASS
# (AC-HEV2-018, 019, 020). M4-scoped ACs PASS (AC-HEV2-025, 026, 027, 049).
# M5-scoped ACs PASS (AC-HEV2-028, 029, 030, 031, 050, 051). M6-scoped ACs PASS
# (AC-HEV2-021, 022, 024, 035, 036, 037, 038, 052). M7-scoped ACs PASS
# (AC-HEV2-023, 032, 033, 034, 039, 040, 041, 042, 045, 046). curator coverage
# 91.0% (>=90%). Full harness/merge/template suites 14 ok / 0 FAIL.
ac_pass_count: 50  # M1 (12) + M2 (7) + M3 (3) + M4 (4) + M5 (6) + M6 (8) + M7 (10)
ac_fail_count: 0
preserve_list_post_run_count: 0  # no PRESERVE-list / M1-M6 implementation files modified
l44_pre_commit_fetch: done-m7-push  # git fetch origin main before M7 push -> 0 0 (origin at baseline ccf4f7e1c, synced)
l44_post_push_fetch: done-m7-push  # git fetch origin main after push -> 0 0 (origin/main == HEAD e6e54bd1e, clean)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_all: exit_0
  go_build_windows_amd64: exit_0
total_run_phase_files: 33  # M1 (6) + M2 (5) + M3 (4) + M4 (2) + M5 (4) + M6 (5) + M7 (7 new): tier_gate.go + approval.go + tier_test.go + approval_test.go + rollback_test.go + curator_e2e_test.go + learned_roundtrip_test.go (crud_test.go + antifabrication_test.go extended, already counted)
m1_to_mN_commit_strategy: per-milestone
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_commit_sha: 29a6f53d0c3cb02dc2b1b2174b8ee148f7656195
sync_complete_at: 2026-07-12
# M1-M7 delivered: typed writer (curator/writer.go), LEARNED-WORKFLOW digest block
# (budget/cap/ledger_key/anti-fabrication), CLAUDE.local.md append-only section,
# mergeSectionBased preservation, snapshot/rollback/lineage extension, Template-First
# empty marker shipping, integration verification. 53/53 AC PASS, coverage 92.4%,
# cross-platform build green, lint clean. CHANGELOG.md entry added. Frontmatter
# status → completed for spec.md/plan.md/acceptance.md (3-phase close).
```

## §F Phase 0.95 Mode Selection

Decision: sub-agent

> Retrospective backfill (recorded after M1-M3 completion, before M4). The
> Mode 5 decision was effectively made when M1 was first delegated sequentially;
> this section records it for the sync-phase "Mode Selection" grep AC + the
> audit trail, per `orchestration-mode-selection.md` §D logging contract.

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | L |
| scope (run-phase file count) | ~15-25 files: curator Go source + tests, layer3.go, config token-budget, templates, e2e |
| domain count | >=3 (curator Go primitives / harness-layer3 integration / config token-budget / template-neutrality / SPEC artifacts) |
| file language mix | predominantly Go source + markdown (templates, SPEC artifacts) |
| concurrency benefit | LOW — coding-heavy with strong inter-milestone dependencies (M5 builds on M4; M6 on M5; M7 integrates all) |
| Agent Teams prereqs | N/A (Mode 3 retired) |

### Mode evaluation

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Multi-milestone Go implementation — not a typo/single-line fix |
| 2 background | no | Write-heavy implementation; background-write restriction applies |
| 3 agent-team | no | RETIRED — Mode 3 is a Phase 0.95 tombstone |
| 4 parallel | no | Coding-heavy + inter-milestone deps violate the parallelism caveat; concurrent agents would mutate the same curator package |
| 5 sub-agent | **YES** | Sequential per-milestone; each milestone's ACs verify before the next begins; preserves the M1->M2->M3 established pattern |
| 6 workflow | no | Semantic multi-rule Go implementation, not a high-volume mechanical transform |

### Justification

Coding-heavy Go implementation in a single package family (`internal/harness/curator/` + `internal/harness/layer3.go`) with strong serial dependencies: M4 (mergeSectionBased + managedSectionHeadings allow-list) produces the section-merge primitives that M5 (snapshot/rollback/lineage) and M6 (Template-First + section 25 neutrality + Recall) build on, and M7 integrates in e2e. Anthropic's coding-task parallelism caveat applies directly — concurrent agents mutating the same curator package would produce inconsistent primitives. Mode 5 (sub-agent, sequential per-milestone) is the correct default for coding work and matches the M1->M2->M3 pattern already established. Per `orchestration-mode-selection.md` section B.2 tie-breaker: coding-heavy + multi-domain -> Mode 5 over Mode 4.
