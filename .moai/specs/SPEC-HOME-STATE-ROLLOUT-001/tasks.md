# tasks.md — SPEC-HOME-STATE-ROLLOUT-001

Card: `t592`
Tier: `L`
Cycle: `tdd`

## Plan phase

- [x] `t592` scope and exclusions captured.
- [x] Relevant home-state, Todo, session, Factory, MCP, handoff, launcher, and clean surfaces researched.
- [x] SPEC ID regex and uniqueness pre-write checks passed.
- [x] Tier L artifact set authored.
- [x] Plan audit iteration 1 D1–D6 findings incorporated into SPEC version 0.2.0.
- [x] Plan audit iteration 2 D1/D2/D5 findings incorporated into SPEC version 0.3.0.
- [x] Independent plan re-audit passed at Tier L threshold (PASS 1.00, blocking finding 0).
- [x] User implementation kickoff approval recorded from the upstream `t592` delegation in this turn; this does not mark implementation or live apply complete.

## Run phase

- [x] M1: Capture verbatim RED for the new migration/barrier API.
- [x] M1: Complete shared cross-process barrier, two fresh censuses, nonce gate, and verified recovery/rollback.
- [x] M2: Capture verbatim RED for schema v2, expiry/reclaim, token ABA, and duplicate boundary.
- [x] M2: Implement resume handoff lease/reclaim and token-bound finish CAS.
- [x] M2: Keep invalid legacy rows claimed, then recover through token-bound pending/failed CAS.
- [x] M3: Capture verbatim RED for global lease registry and lifecycle API.
- [x] M3: Complete Windows child-process handoff wiring and cross-build proof.
- [x] Bind fresh changed-surface coverage into `validateLivePreApply` with low/zero/error/tamper rejection.
- [x] Verify AC-HSR-025 unknown-owner recovery requires zero-active census.
- [x] Reach the 85% changed-line coverage gate (exact mechanical diff-line result: 1198/1408 = 85.085%; prior static allowlist result invalidated).
- [x] F6-R2: fail closed on worktree inventory command failure, empty/malformed porcelain, unavailable root, and canonical mismatch; preserve linked registry and PID dedup behavior.
- [x] F14: same live PID in primary and linked registries counts once and produces stable `1:0:0` fingerprint; dedup-removal mutant rejected.
- [x] Re-measure after F6-R2: 1214/1425 = 85.193%.
- [x] Final independent sync-audit: PASS 100/100, findings 0; live rollout remains pending.
- [x] Fix post-commit clean-tree coverage resolution and F15 child recursion guard; verify clean, merge, later unrelated, dirty, stale, duplicate, non-descendant, and all-child guard fixtures; exact coverage 1197/1408 = 85.014%.
- [ ] Run independent sync re-audit for the post-commit coverage remediation.

## Live rollout

- [ ] Re-read canonical project root, project key, source, target, and all active-runtime registries.
- [ ] Confirm fresh source integrity/count; do not reuse the plan-time `75` as current evidence.
- [ ] Run pre-apply AC-001..021/023..025 validators at current HEAD, verify determinate zero-active census, then consume the in-memory one-shot nonce before any mutation.
- [ ] Prove bare `--apply` plus stale, tampered, forged, zero-test and replay authorization fail while marker/backup/target inventories remain unchanged.
- [ ] After authorization, install the exclusive admission marker first; verify new SessionStart/Factory/MCP admission is blocked before backup or data apply.
- [ ] Create and verify backup as the first data mutation, bind its ID/hash to the migration, then apply data.
- [ ] After apply only, run AC-022 against AC-001..021/023..025 plus readback; append AC-022 result afterward and let sync audit close all-25 completeness.
- [ ] Verify source/target logical parity and integrity; preserve original and backup.
- [ ] Confirm no global forced clean ran.

## Deferred follow-ups

- [ ] Search runtime producer, rebuild, and invalidation — out of scope.
- [ ] Receiver-side exactly-once handoff dedupe — out of scope.
