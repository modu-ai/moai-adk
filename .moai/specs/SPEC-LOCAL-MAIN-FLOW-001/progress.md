# SPEC-LOCAL-MAIN-FLOW-001 — Progress

Plan-phase skeleton. The four §E headings are the canonical markers, in this order. Each phase writes only its own section.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
audit_ready: true
```

## §E.2 Run-phase Evidence

Range: `0106da184..af8199f8f` (`git log --format='%h %s'`, read at HEAD `af8199f8f` before this commit). Excluded: the plan-phase tip `0106da184` and the absorb merge `e32f69c46`. Subjects are verbatim; each group is in chronological order.

**M1** (REQ-LMF-001, 002, 003, 004, 006, and the primary re-sync of REQ-LMF-014; `eb3c477c0` is a one-line R2_D correction outside the M1 list)
- `fe2362f5d` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for primary-merge ignored-content shapes
- `8526e72de` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for integration surface, configuration, and verb
- `c6dab38f9` test(SPEC-LOCAL-MAIN-FLOW-001): repair merge RED tests to observe the step
- `dfee70154` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for the primary re-sync with compile-only stubs
- `e0f132a56` feat(SPEC-LOCAL-MAIN-FLOW-001): dirty-primary guidance and post-merge symbolic HEAD hold
- `ccfa83598` feat(SPEC-LOCAL-MAIN-FLOW-001): local-main integration config key, shipped off
- `237432a1a` feat(SPEC-LOCAL-MAIN-FLOW-001): primary integration surface behind the local_main_integration gate
- `d0d7901a5` feat(SPEC-LOCAL-MAIN-FLOW-001): factory complete lands the card through the primary when the gate is on
- `a65a4f179` feat(SPEC-LOCAL-MAIN-FLOW-001): primary re-sync and the integration resync verb
- `eb3c477c0` test(SPEC-LOCAL-MAIN-FLOW-001): case-insensitive provisioning match in TestR2_D (card t1616)
- `7fe19c387` test(SPEC-LOCAL-MAIN-FLOW-001): holder-calls-resync keeps its window (card t1616)
- `cbbae93de` fix(SPEC-LOCAL-MAIN-FLOW-001): resync keeps a window the caller already held (card t1616)

**M2** (REQ-LMF-005, separate worktree surface)
- `5db413a58` test(SPEC-LOCAL-MAIN-FLOW-001): RED tests for the separate-worktree overlap decision (card t1616)
- `c08605b44` feat(SPEC-LOCAL-MAIN-FLOW-001): separate-worktree surface uses the B5 overlap decision (card t1616)

**M3** (REQ-LMF-008, 009; M3a `5e202adf1` lands before M3b `1236348ac`)
- `5e202adf1` chore(SPEC-LOCAL-MAIN-FLOW-001): auto_merge off before develop_branch moves to main (card t1616)
- `1236348ac` chore(SPEC-LOCAL-MAIN-FLOW-001): local-main integration on, develop_branch main (card t1616)

**M4** (REQ-LMF-010 to 014)
- `9c7a56fe0` docs(SPEC-LOCAL-MAIN-FLOW-001): AGENTS.md branch wording from the template (card t1616)
- `80ecc6145` docs(SPEC-LOCAL-MAIN-FLOW-001): supersede markers in AGENTS.local.md (card t1616)
- `c556cf063` docs(SPEC-LOCAL-MAIN-FLOW-001): AGENTS.local.md §4.0 default flow, local-main integration (card t1616)
- `34739c6ef` docs(SPEC-LOCAL-MAIN-FLOW-001): gitflow-lane-protocol.md local-main flow (card t1616)
- `363e6ff83` docs(SPEC-LOCAL-MAIN-FLOW-001): repo-local-pr-policy.md local-main flow (card t1616)
- `dea179c6b` docs(SPEC-LOCAL-MAIN-FLOW-001): gitflow-integration-chain.md local-main flow (card t1616)
- `af8199f8f` docs(SPEC-LOCAL-MAIN-FLOW-001): hns-release-specialist.md local-main flow (card t1616)

**No milestone** (artifact commits: measurement pins and operator verdicts)
- `7642093a0` docs(SPEC-LOCAL-MAIN-FLOW-001): measurement pins against 0106da184 and the audit-ready signal
- `9ca36a428` docs(SPEC-LOCAL-MAIN-FLOW-001): operator verdicts 3fbd in decision-index (card t1616)

**M5**: satisfied by the AGENTS.local.md §4.0 section committed in `c556cf063` (plan §C); no separate commit.

**M6** (REQ-LMF-015): no commit. Specification only (plan §G.2 and §G.3); no code, no hook, and no measurement procedure run in this card. Counts observed in this commit against plan.md: `^### G\.[23] Gap` 2, `Run-phase procedure` 2, `^Open:` 2. M7 is operator-gated and not required.

**Verification at af8199f8f**: taken from the lane progress record, section 64 (`.moai/reports/t1616/progress.md`, gitignored); the command outputs are in the evidence files `.moai/reports/t1616/verify-*-af8199f8f.txt`. This commit did not re-run them.
- V2 13/13 PASS; V3 3/3 PASS; V4 43/43 PASS; V5 9/9 PASS; V6 `make build` exit 0 with the tree still clean; V8 exit 0; V11 empty diff.
- V1 is red as raw: one gofmt violation in `internal/factory/integration_remeasure_test.go`, a file the card did not change (Gap 9).
- V7 is not run at af8199f8f; it is measured at the final state after sync. V10 is not observed (Gap 13).

**Gaps** (numbers and titles; detail in lane record §63 items 1-12 and §64 item 13)
1. §4.0 adds +3,332 characters to an always-loaded file
2. draft → in-progress missing at M1; applied at this run-close commit
3. REQ-LMF-014 cites REQ-LMF-011 inconsistently
4. Bodies of `34739c6ef`, `363e6ff83`, `dea179c6b` carry no card id
5. Trailer layout: `card:` and `Authored-By-Agent:` are not parsed as git trailers (leader kept the layout)
6. decision-index.md header and plan §H still list Q1 and Q2 as open
7. §11 re-sync decision formula not checked against the code
8. lane-protocol §6 target rule for card done; done is blocked until t1621
9. V1 pre-existing gofmt violation in `internal/factory/integration_remeasure_test.go`
10. Process deviation: three git commands chained with `;` in one read-only Bash call
11. OQ-7 dependent markers still open (lane-protocol line 153; hns-release-specialist.md, 18 lines)
12. plan §E slot label is wrong for the two path-scoped rule files
13. V10 (PreToolUse hook verdict on a temporary primary) not observed

## §E.3 Run-phase Audit-Ready Signal

```yaml
audit_ready: true
```

Status transition: draft → in-progress at this run-close commit (Gap; ruling d-20261010T062901Z-6d3d, item 2).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
audit_ready: true
sync_commit_sha: pending-backfill
```

Status transition: in-progress → implemented at this sync commit. The implemented → completed transition and the sync_commit_sha backfill are pending (Gap; ruling ca0c, item 3), to be settled at the landing-resume ruling.
