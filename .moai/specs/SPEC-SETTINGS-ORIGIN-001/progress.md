# progress.md — SPEC-SETTINGS-ORIGIN-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-05
artifacts: spec.md + plan.md + acceptance.md (Tier S + additive acceptance.md per t487 dispatch) + progress.md — v0.2.0 (plan-audit PASS 0.80, fixes D1-D4+D7 applied; verdict: .moai/reports/t487/plan-audit.md)
baseline: worktree t487, base 25a3212a9, branch WT-settings-origin

## §E.2 Run-phase Evidence

Run phase completed 2026-09-06 (read-only investigation; production code changes: 0, per SPEC §5). Terminal artifact: `.moai/reports/t487/verdict.md` (5-section evidence-bearing format). Raw sweep data: `.moai/reports/t487/sweep/` (sweep.tsv, blob-check.tsv, tree-list.txt, tree-heads.txt, run-sweep.sh, verify-blobs.sh).

| AC | Status | Verification command (representative) | Observed |
|---|---|---|---|
| AC-001 inventory completeness | PASS | `rg -n 'settings\.json' internal pkg cmd` + defs-constant reverse trace + repo-root second-pattern sweep | 7 writer paths with file:line (verdict E3); cross-check surfaced 0 misses; scripts/CI/Makefile/hooks write redirections: 0 |
| AC-002 shape reconciliation | PASS | per-writer serialization-order reading (deploy render / map-marshal / toolpolicy splice) | every writer emits template order, alphabetical order, or byte-preserving splice — none produces the dirty order (verdict E3 table) |
| AC-003 sweep distribution | PASS | `run-sweep.sh` (79 trees, shape-based, cross-tree git 0) + `verify-blobs.sh` (HEAD-blob comparison, own-tree `git show` only) | 79 enumerated = 79 checked; dirty-family hit: t334 (md5 4f455d94…, HEAD blob 437834679… mismatch); dirty md5 b669… exact match: 0 trees; 78/79 CLEAN_MATCHES_HEAD |
| AC-004 conditional external inflow | PASS | branch taken: M4 executed (Q1 found no producer — verdict C1) | CC-rewrite order measured alphabetical on 4 sibling projects (dirty ≠ sorted); disk binaries 8-era sampled — none embeds dirty-order template; sibling checkouts 13 measured — 0 match; npm TS predecessor: NOT FETCHABLE in this spawn — flagged for lane (verdict Gaps 1) |
| AC-005 single recommendation | PASS | — | exactly one: pre-merge `.claude/settings.json` drift assertion in the card-worktree merge window (`--no-optional-locks` form), preserve+report on hit (verdict §Q3) |
| AC-006 verdict artifact | PASS | `ls .moai/reports/t487/verdict.md` | exists; all 5 sections present |
| AC-007 evidence integrity | PASS | — | every quantitative/negative claim carries command+output (verdict Evidence P0-E7); sweep NOFILE count is a measured 0, not absent output |
| AC-008 baseline reuse | PASS | — | t480's six eliminations cited only (0 re-executions); §1 fingerprint used as hypotheses H1-H4 only (verdict E7) |

Key findings (detail + verbatim evidence in verdict.md):

1. **Second dirty-family instance discovered — t334 worktree is dirty NOW** (md5 `4f455d9425a396d38c202f2614bc918f` vs HEAD blob `437834679fcc2435761e264cec2ffc15`): same 11-key dirty order, missing 3 keys, narrow matcher, sto=1, but ask=`[]` (thinner than the t452 variant's `[sudo]`). Byte-preserved — recorded only, never modified.
2. H4 resolved: the writer class is **systematic** (2 instances / 6 days / distinct trees), writing **mid-session** (t334 file mtime 3h after worktree birth) — creation-time replication refuted.
3. Q1: no repository code path can produce the dirty shape (7 writers inventoried, cross-check clean). H2 (old render/binary) eliminated across all committed states (template 93-commit history, tracked 64-commit history, release tags v3.0.0/v3.1.0/v3.1.2, TS-era, sampled on-disk binaries). H1 (standard CC rewrite) eliminated on the order axis locally (CC rewrites alphabetically). H3 (hand/AI re-assembly) is the surviving best-supported hypothesis — process name unrecoverable per t485 C4.
4. Q3: one process-level recommendation (pre-merge drift assertion), t485-C4-compliant.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-06
run_commit_sha: 19422aa6e
run_status: run-complete (all 8 AC PASS by inspection of verdict.md; evidence-only close per SPEC §5 — no code, no test gate applicable)
ac_pass_count: 8
ac_fail_count: 0
evidence_path: .moai/reports/t487/verdict.md
sweep_data_path: .moai/reports/t487/sweep/
measurement_scope: 79 worktrees (shape sweep + HEAD-blob comparison, read-only) · repo writer inventory (internal/pkg/cmd/hooks/scripts/CI/Makefile) · template/tracked key-order history (tags + 8-of-64 tracked commits + 8-of-93 .tmpl commits sampled across all eras) · on-disk binary embedded-template sampling (8 eras) · sibling-project settings shapes (13 projects)
constraints_honored: cross-tree git 0 (guard) · byte-preserve (t334 dirty recorded, untouched) · no push · no worktree disposal · no background load · zero production code changes

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-06
sync_commit_sha: "c7da0d522"
sync_status: completed
b12_self_test_a: n/a — no CHANGELOG emission (see disposition below)
b12_self_test_b: n/a — no AC-count comparison applicable without a CHANGELOG entry
b12_self_test_c: n/a — no file paths claimed in a CHANGELOG entry

**Sync summary** — Sync scope: spec.md frontmatter transition only (`status: in-progress → implemented → completed` merged close + `updated: 2026-09-06` refresh; no body content of spec.md / plan.md / acceptance.md touched) plus this §E.4 authoring, on the single sync commit. CHANGELOG disposition: N/A — no CHANGELOG.md entry emitted. This SPEC is a Tier S read-only investigation with zero production code changes (SPEC §5); the run output is analysis + evidence files under `.moai/reports/t487/` (verdict.md, sweep/, preserved-copies/), which is a local-only surface that ships to no distributed template, binary, or docs-site consumer. With no shipped product surface changed, there is no user-visible entry to record; no sync-gate rule mandates a CHANGELOG entry for investigation-only SPECs with no product-surface delta. Verdict evidence path: `.moai/reports/t487/verdict.md` (5 sections + lane addendum L1-L3). The `sync_commit_sha` field above carries the sanctioned `pending-backfill-sync` placeholder (D3 pattern — a commit cannot cite its own SHA) and is backfilled with the real sync commit SHA in the immediately following backfill commit.
