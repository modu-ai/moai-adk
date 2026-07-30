# SPEC-CLIFIX-HYGIENE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §3/§4 hygiene rows + §5 P4. Status: draft. Pending plan-audit. Sequenced last in the CLIFIX series (P0→P4).
- 2026-07-30: v0.2.0 rescope after plan-audit iter-1 FAIL (0.58, D1-D11 stale anchors). Re-verified every anchor against current worktree (HEAD = origin/main, post P0-P3 merge). Dropped REQ-HYG-001-007 (deepMerge3Way — already fixed in `internal/merge/`) and REQ-HYG-001-008 (token 0600 — superseded by F1 no-persist redesign); renumbered old 009→007, old 010→008; refreshed all line/symbol anchors; added `### Out of Scope — Broader i18n sweep` (~24 deferred Hangul files). Status: draft — NOT audit-ready (pending re-audit).

## §E.2 Run-phase Evidence

### M1 — PRESERVE net + ANALYZE inventories (2026-07-30)

**M1 scope (per plan.md §F M1): characterization tests + goldens for update
flows + Hangul/threshold/env-literal inventories frozen as fixtures. NO
production `.go` edits.**

#### Characterization net added

- File: `internal/cli/update_hygiene_characterization_test.go` (new, ~230 LOC).
- Surface covered (per plan.md §F M1: "dry-run plan path AND settings sync/merge path"):
  - **Settings sync/merge path** — `cleanLegacyHooks` (update.go:1759): 4 characterization tests pinning the exact 10-pattern legacy set (golden), the per-pattern removal contract, the non-legacy-hook preservation contract, and the mixed-group surgical-delete contract. The no-hooks-section fast-path short-circuit is also pinned.
  - **Dry-run plan path** — `runAgencyMigrationAdapter` (update.go:1247): 2 characterization tests pinning the ErrMigrateNoSource → return-nil suppression contract (no .agency source must not fail update) and the dryRun=true read-only guarantee (no `.moai` mutation when dryRun is set).
- PASS evidence (verbatim, on unmodified code):
  - `go test ./internal/cli/ -run 'CleanLegacyHooks|RunAgencyMigrationAdapter' -count=1 -v` → `PASS` / `ok github.com/modu-ai/moai-adk/internal/cli 0.739s` (6 new tests + pre-existing `cleanLegacyHooks` tests all green).

#### ANALYZE caller-graph verdicts (dead-code inventory)

Frozen at `.moai/specs/SPEC-CLIFIX-HYGIENE-001/inventories/dead-code-caller-graph.md`. Per-symbol verdict:

| Symbol | Verdict | Production callers | Notes |
|---|---|---:|---|
| `buildGLMEnvVars` (glm.go:917) | **DEAD (M5-deletable)** | 0 | test-only callers in glm_team_test.go + coverage_improvement_test.go; no tag/reflection/registration gate |
| `ttyConfirmer` (branch_protection.go:39-55) | **DEFERRED** | 0 | self-attested `SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (paired with ttyConfirmer)` — M5 must resolve the pairing first; live confirmer is `yesConfirmer` |
| `worktree_validation.go` (whole file) | **DEAD-in-prod, DEFERRED-pending-wire-up-decision** | 0 | file header + `@MX:NOTE` self-attest "wire-up planned in follow-up SPEC"; exported but `internal/cli` blocks external import; M5 surfaces to orchestrator before deleting |

PRESERVE-KEEP (confirmed LIVE — NOT deletable): `acquireUpdateLock` (update.go:264 caller, P0-wired), `cleanStaleLock` (update_cleanup.go:59 caller, P0-wired), `scanDeprecatedPaths` (4 callers in update_clean_install.go), `cleanup_old_backups`.

Realistic M5 net-LOC delta walked back honestly (vs the audit's ~500-line figure): **−60 to −320 lines**, contingent on the two DEFERRED resolutions.

#### Inventories frozen as fixtures (diffable for M3/M4/M5)

| Fixture | Path | M-milestone consumer |
|---|---|---|
| Dead-code caller-graph | `inventories/dead-code-caller-graph.md` | M5 (deletion) |
| Hangul strings snapshot (6 in-scope files, 73 lines across 5 files; doctor.go already clean = drift finding recorded) | `inventories/hangul-strings.snapshot.md` | M4 (English UI strings) |
| Threshold-literal snapshot (THREE `[]int{1,3,5,10}` sites + TWO `30*time.Second` dispatcher timeouts — original "×2" audit undercount corrected) | `inventories/threshold-literals.snapshot.md` | M3 (defaults.go extraction) |
| GLM env inject/clear sites snapshot (11 sites across glm.go/glm_tools.go/launcher.go) | `inventories/glm-env-sites.snapshot.md` | M3 (envkeys.go + glmEnvVarSet) |

#### AC setup vs completion (E1 matrix — M1 is the safety net)

M1 does NOT complete any AC by itself; it SETS UP the ACs the later milestones
will satisfy. M1's load-bearing claim is "characterization tests PASS on
unmodified code" (the PRESERVE guarantee), verified verbatim above.

| AC | M1 status | Milestone that completes it |
|---|---|---|
| AC-HYG-001-001 | SET UP (characterization net exists; M5 must keep it byte-identical post-split) | M5 |
| AC-HYG-001-002 | SET UP (caller-graph inventory frozen; M5 performs the deletion) | M5 |
| AC-HYG-001-003 | SET UP (env-site inventory frozen) | M3 |
| AC-HYG-001-004 | SET UP (threshold-literal inventory frozen) | M3 |
| AC-HYG-001-005 | SET UP (Hangul inventory frozen; doctor.go already-clean drift recorded) | M4 |
| AC-HYG-001-006 | not started (M2 repro-first) | M2 |
| AC-HYG-001-007 | not started (M2 repro-first) | M2 |
| AC-HYG-001-008 | not started (M2 repro-first) | M2 |

#### M1 production-code edit count: **0** (PRESERVE + ANALYZE only — per plan.md §F M1 scope)

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Date: 2026-07-30 (run-phase entry, post Implementation Kickoff Approval — user approved semi-autonomous per-milestone verification)
- Input parameters: tier=L; scope≈12-15 files (internal/cli update cluster + wizard + internal/worktree + read-only internal/merge); domain count=1 (CLI hygiene, single Go module); language mix=Go; concurrency benefit=LOW (coding-heavy refactoring + defect fixes, not parallelizable research).
- Mode evaluation:
  - trivial — not selected (multi-file refactoring + defect fixes)
  - background — not selected (write-capable, not read-only)
  - agent-team — RETIRED (not selectable)
  - parallel — not selected (coding-heavy single-domain; Anthropic coding-task parallelism caveat)
  - workflow — not selected (semantic refactoring + multi-rule defect fixes, not a single ≥30-file uniform mechanical transform)
  - sub-agent — selected
- Decision: sub-agent (Mode 5)
- Justification: B6 is coding-heavy, single-domain CLI hygiene work (DDD characterization-first decomposition + reproduction-first TDD defect correction). Per Anthropic's coding-task parallelism caveat, sequential per-milestone sub-agent delegation is the safe default for coding work; the work is semantic/multi-rule (not a uniform mechanical transform), so Mode 6 workflow does not apply. Mode 5 with the Tier L Section A-E delegation template, one milestone per spawn, orchestrator trust-but-verify (build/test/lint/AC) between milestones.
- Note: worktree base = origin/main 2f449e189; origin/main has since advanced to cc1986cff (+2, multi-session race) — update-branch at run-PR time.
