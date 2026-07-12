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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at:
run_commit_sha: pending-backfill-m3
run_status: m3-complete-m4-m7-pending
# M1 (Typed Managed-Block Writer foundation) + M2 (LEARNED digest block +
# budget/cap enforcement) + M3 (CLAUDE.local.md append-only LOCAL section) are
# complete. M1-scoped ACs PASS (AC-HEV2-001..006, 009..011, 044, 047, 048).
# M2-scoped ACs PASS (AC-HEV2-007, 008, 012, 013, 014, 015, 016). M3-scoped
# ACs PASS (AC-HEV2-018, 019, 020). Coverage 93.3%. Remaining milestones
# M4-M7 NOT started — run_status is NOT audit-ready.
ac_pass_count: 22  # M1 (12) + M2 (7) + M3 (3)
ac_fail_count: 0
preserve_list_post_run_count: 0  # no PRESERVE-list files modified
l44_pre_commit_fetch: pending-m3-push
l44_post_push_fetch: pending-m3-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_all: exit_0
  go_build_windows_amd64: exit_0
total_run_phase_files: 15  # M1 (6) + M2 (5) + M3 (4): append.go + append_test.go + writer.go mod + marker_test.go mod
m1_to_mN_commit_strategy: per-milestone
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
