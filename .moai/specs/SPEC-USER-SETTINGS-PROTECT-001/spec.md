---
id: SPEC-USER-SETTINGS-PROTECT-001
title: "User settings protection: init keeps user permissions, tests stay out of the operator home, clean --home covers run and db"
version: "0.1.0"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/config/toolpolicy, internal/core/project, internal/cli, internal/homestate, internal/escalation, internal/contract/receipt, internal/factory, internal/factorymsg, internal/web"
lifecycle: spec-anchored
tags: "settings, permissions, init, test-isolation, home-sandbox, codex, clean-home, card-3.2-0-1"
tier: M
---

# SPEC-USER-SETTINGS-PROTECT-001 — User Settings Protection (card 3.2-0-1)

## HISTORY

- 2026-10-10: v0.1.0 created by manager-spec. Card t1630 (backlog 3.2-0-1 "사용자 설정 보호", priority P0, change class C, design change), plan phase only. Dispatched by factory lane-18 (run tmnboq). Absorbed cards in scope: t1567 and t1594. Measured evidence and gaps: plan.md §A. Operator decisions the artifacts depend on: decision-index.md rows Q1-Q7.

## 0. Tier basis (provisional)

Tier M is kept by the orchestrator's decision (lane-18, ladder ⑤). Ruling d-20261010T081810Z-49e5 (board, kind=ruling) confirms option (b): Tier M kept at 16 files, with the Q2 template file in this SPEC and the basis recorded in this section. The file band of 5 to 15 files is a guide, not a gate. The affected-file count is 15 without the Q2 template file and 16 with it (plan.md §A.4 Basis); the 16th file is a one-line template edit. That decision overrides the tier-up trigger, which would otherwise require Tier L or a split before run. The run-phase reach measurement (plan.md §A.4 G-9) stays a recorded input and is reported in progress.md. The requirement count is 13 (ceiling 16) and the acceptance-criterion count is 13 (ceiling 16). The Tier judgment is normally a Socratic question in spec-assembly; this SPEC runs without a user channel, so the lane records the judgment in progress.md §G.

Q5 adds no file: the PROJECT-scope change lives in `internal/config/toolpolicy/tier_render.go`, which is already counted as item 1.

## 1. Background and Premise

Card 3.2-0-1 asks for three properties. First, `moai init` leaves the user's existing settings permissions intact. Second, tests never touch the operator's home directory. Third, `moai clean --home` covers the run and db categories. Absorbed card t1567 (permissions.defaultMode in the template) is in scope, and its template item follows decided Q2 (the template carries the default, pinned to `"default"` by the lane in progress.md §G). Absorbed card t1594 (user assets and deployment residue) is blocked on its card text (M8) and has no requirement in this revision.

The measured state at tree 2aab5f797 (commands and verbatim output in plan.md §A.2):

- S-1 (card item a) — init drops user permissions. The default init path rewrites the USER-scope settings `permissions` object to `defaultMode` only. The user's `allow`, `ask`, and `deny` lists are removed. Keys the writer does not model (for example `additionalDirectories`) and sibling keys (`env`) survive. Verified by two probes (plan.md E-3, E-4). The contract this violates is already committed: SPEC-INIT-WIZARD-REPAIR-001 §4 states that the distributed-default write changes exactly the `permissions.defaultMode` key and preserves allow, deny, and ask verbatim, and its v0.1.1 HISTORY row pins that condition (plan.md E-23; decision-index.md Q8, DECIDED). The defect is therefore a regression against a completed contract, not a new design question.
- S-2 (card item b) — one live Codex test is not isolated. The live audit fixture already sets CODEX_HOME to a temporary directory. The live review-gate test sets neither CODEX_HOME nor HOME.
- S-3 (card item c) — most test binaries do not block the real home. Of the 38 `TestMain(m *testing.M)` entry points, 36 contain no MOAI_HOME or HOME redirection. Only `internal/cli` and `internal/hook` redirect MOAI_HOME.
- S-4 (card item d) — the store roots follow MOAI_HOME. `homestate.RunProjectDir` and the receipt and escalation `StoreDir` resolvers resolve under MOAI_HOME, which only the cli and hook binaries redirect.
- S-5 (card item e) — `clean --home` has no run or db category. Its scan covers projects, debug, releases, logs, and backups.
- S-6 (t1567) — the settings template ships no `permissions.defaultMode`. The t1567 card text is not in the tree (Gap G-1).
- S-7 (t1594) — the residue inventory for t1594 is not in the tree (Gap G-2). Its scope item is blocked.

## 2. Requirements (GEARS)

### Permission preservation (S-1, S-6)

- REQ-001: The `moai init` USER-scope settings writer shall preserve every existing `permissions.allow`, `permissions.ask`, and `permissions.deny` entry and every other key of the existing `permissions` object, changing only `permissions.defaultMode`. This restores the condition pinned by SPEC-INIT-WIZARD-REPAIR-001 §4 (decision-index.md Q8, DECIDED).
- REQ-002: When `moai init` writes the USER-scope settings file and the resolved defaultMode already matches the value in the file, the writer shall not rewrite the `permissions` region, so that its bytes are identical before and after the init run.
- REQ-003: The `moai init` default path and automatic path shall keep an existing USER-scope `permissions.defaultMode` value and write that key only when it is absent (decision Q1, DECIDED).
- REQ-004: The settings template `.claude/settings.json.tmpl` shall carry a `permissions.defaultMode` value of `"default"`; a fresh `moai init` shall write that key only where it is absent, and `moai update` shall not modify it (decision Q2, DECIDED; the value is pinned in progress.md §G).
- REQ-005: Where a tool-policy document is present and the policy bundle is applied, the PROJECT-scope writer shall keep user-added `permissions.allow` entries by set union with the regenerated managed block. The writer keeps the last generated managed allow list in `.moai/state/tool-policy/managed-allow.json` as `last_generated`. User-added entries are the existing allow entries absent from `last_generated`, and the result is the regenerated list union the user-added entries. With no such record, every existing entry is kept, so nothing is removed. Only the user removes a user-added entry (decision Q5, DECIDED; detection pinned in progress.md §G).

### Test isolation (S-2, S-3, S-4)

- REQ-006: Every Go test package whose test binary reaches a home-resolving function shall install a MOAI_HOME sandbox, a directory owned by that test binary, in its `TestMain` before `m.Run()` is called.
- REQ-007: While a Go test binary of a package that reaches a home-resolving function runs, the process HOME environment variable shall point at a test-owned directory, so that no path under the operator's real home (`.moai`, `.claude`, `.codex`, `.agents`) is written by that binary. Excluded: packages whose test files reference no home-resolving function, because their binaries resolve no home path (plan.md E-9, a name-based measure; G-9).
- REQ-008: While a Go test binary runs under its MOAI_HOME sandbox, `homestate.RunProjectDir` and the receipt and escalation `StoreDir` resolvers shall return paths under the MOAI_HOME sandbox root.
- REQ-009: Every live Codex test shall set CODEX_HOME to a test-owned directory before it starts any `codex` child process, including `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey`.
- REQ-010: When a test package's `TestMain` omits the sandbox that REQ-006 requires, the repository test run shall fail with a named finding. The mutant procedure that proves this is in acceptance.md AC-010.

### clean --home (S-5)

- REQ-011: `moai clean --home` shall include the `run` and `db` categories in its candidate scan. A `run` item is a delete candidate only when no live record (active run, lease, unclosed card) references it. A `db` item is listed, and is deleted only with the explicit confirmation flag `--force`, which is the record's `--yes` (decision Q3, DECIDED; the flag is resolved by ruling d-20261010T081810Z-49e5).
- REQ-012: `moai clean --home` shall keep dry-run as its default, shall delete only allowlisted categories under `--force`, and shall keep the carve-out predicate unchanged. The allowlist includes the `run` and `db` categories, so that `--force` deletes run and db candidates. Run items are candidates only under the REQ-011 rule, and `--force` deletes only candidates. Without `--force`, run and db items are listed only (decision Q3; progress.md §G, G-23).

### Ordering (cross-cutting)

- REQ-013: The run phase shall not start before the landing commits of cards t1619, t1578, and t1591 are ancestors of the run base tip. The acceptance criteria that depend on the init writer shall be executed after t1619 lands.

## 3. Acceptance Criteria (summary)

The authoritative matrix lives in acceptance.md (AC-001 to AC-013, 13 criteria, each naming its verifying command). Coverage map: REQ-001 → AC-001; REQ-002 → AC-002; REQ-003 → AC-003; REQ-004 → AC-004; REQ-005 → AC-005; REQ-006 → AC-006; REQ-007 → AC-007; REQ-008 → AC-008; REQ-009 → AC-009; REQ-010 → AC-010; REQ-011 → AC-011; REQ-012 → AC-012; REQ-013 → AC-013.

The completion judgements from the card are carried as criteria. The permissions-section diff across `moai init` is 0 when the existing defaultMode already matches the resolved tier default (AC-002). The differing case is decided by Q1 (the existing value is kept) and covered by AC-003 part B; the lists survive in both cases (AC-003 part A). The entry count under `~/.moai/run` increases by 0 after one test run (AC-007).

## 4. Constraints

- C1: No test in the suite writes under the operator's real home. Measurements that would show such a write are run against a sandboxed HOME or a throwaway account, never against the operator's home.
- C2: The init writer changes only the settings file it targets and writes atomically (the existing atomic-file path is retained).
- C3: The `phase:` field carries a release label. Lifecycle tokens are not used.
- C4: Every change lands test-first. A RED observation on the pre-implementation tree is recorded before the GREEN change (verification-completeness rule, §2).
- C5: The fix changes no gate semantics. The autonomy-tier gates and the sandbox-proof gate keep their current behaviour.
- C6: Verification evidence is attributed to the tree it measured, with the command and verbatim output (verification-claim-integrity §2).

## 5. Open Decisions

decision-index.md carries the rows. Q1 (existing defaultMode kept), Q2 (template defaultMode, written only when absent), Q3 (candidate rule for run and db), and Q5 (user-added allow entries kept) are DECIDED by the pinned board record `board:d-20261010T073547Z-07ae#9c93e47809f9`. Q8 (preservation of the user's lists on the distributed-default write) is DECIDED by a committed HISTORY row and is not open. Q4 (profile-scoped USER write) and Q6 (init --force on project settings) are EVIDENCE-NEEDED and close at M2 without blocking run. Q7 (sandbox mechanism) is FOUNDER, implementation-level, and open: no default is applied, so the autonomous Kickoff waits for an operator verdict. The earlier wording that applied a default to Q7 at plan close is withdrawn. Literals the record leaves open are settled by the lane in progress.md §G: the template value is `"default"` (G-17) and the managed-block detection is diff-based (G-20). The record's `--yes` is the existing `--force` flag (ruling d-20261010T081810Z-49e5; G-19).

## 6. Non-goals and Out of Scope

### Out of Scope — install design reconstruction

- The phrase "설치 설계 §9" does not appear in any committed document of this tree (plan.md G-4). This SPEC does not reconstruct or cite its contents.
- The t1594 residue inventory is blocked on its card text (M8; plan.md G-2, §B B4). It has no requirement in this revision. This is a blocked disposition, not an out-of-scope one.

### Out of Scope — writer attribution and drift detection

- Attributing the writer of dirty tracked project `.claude/settings.json` copies stays with SPEC-SETTINGS-ORIGIN-001.
- The merge-window drift predicate stays as SPEC-PREMERGE-SETTINGS-DRIFT-001 defines it. This SPEC does not change it.

### Out of Scope — user-folder assets and update namespaces

- The common skill and agent install and the project payload contract (SPEC-USER-ASSET-INSTALL-001) are out of scope.
- User-owned namespace protection on `moai update` (SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001) is out of scope.

### Out of Scope — profile routing and operator-home measurement

- Routing the USER-scope write to an active `CLAUDE_CONFIG_DIR` profile is out of scope unless decision Q4 resolves to it.
- Measuring the real `~/.moai/run` entry count is a run-phase measurement outside this plan.

## 7. Covering-SPEC cross-check

Each entry states what the SPEC covers, what it does not cover, and the delta this card adds. Status is read from the SPEC's own frontmatter; whether each behaviour is live today is checked in plan.md §A.2 (E-5, E-6, E-19).

1. SPEC-USER-ASSET-INSTALL-001 (status completed, tier L).
   - Covers: per-user folder install of common skills and agents; the project payload reduced to default settings, AGENTS.md, the lock file, and the project-only harness (REQ-005); the init first-install trigger (REQ-024). Live: `internal/cli/init.go` calls `ensureUserAssetsLocked` after the executor (E-5).
   - Does not cover: the permissions content of any settings file, including the USER-scope file the autonomy bundle writes. No REQ names `permissions`.
   - Delta: REQ-001 to REQ-005 of this SPEC govern the permissions content. The per-profile rule (REQ-002 of USER-ASSET) is not changed.
2. SPEC-SETTINGS-ORIGIN-001 (status completed, tier S, read-only investigation).
   - Covers: the attribution question for dirty tracked project `.claude/settings.json` copies in worktrees t452 and t334. §5 excludes production code changes.
   - Does not cover: any writer fix, and the USER-scope file.
   - Delta: none to its requirements. This SPEC does not establish that the init writer produces the dirty shape recorded in its §1; that link is unmeasured (plan.md G-7). Its verdict artifact is untracked and is not cited here.
3. SPEC-PREMERGE-SETTINGS-DRIFT-001 (status completed, tier M, re-closed after later code landed).
   - Covers: merge-window detection of drift in the project `.claude/settings.json` through `git --no-optional-locks status --porcelain -- .claude/settings.json` (REQ-PSD-001, REQ-PSD-004); preserve-copy; no automatic restore (REQ-PSD-006).
   - Does not cover: the USER-scope `~/.claude/settings.json`; prevention at the writer. REQ-PSD-004 fixes the detector path to the project file.
   - Delta: this SPEC prevents loss at the init writer. The PREMERGE detector is unchanged.
4. SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (status implemented, not completed; tier M).
   - Covers: `moai update` preservation of user-owned namespaces (`.claude/skills/my-harness-*`, `.claude/agents/harness/`, `.moai/harness/`) (REQ-UNP-001 to REQ-UNP-010). Live: the `my-harness` token appears in `internal/cli/update_preserve_inventory.go` and `internal/cli/update/plan/plan.go`, not in `internal/cli/update.go` as its module field states (E-19).
   - Does not cover: settings permissions, and the init path.
   - Delta: none. Its progress record has no close marker (no sync_commit_sha, plan_status, or §E heading), so its behaviour is treated as unverified beyond the grep.

## 8. Dependencies

- Ordering (REQ-013): the t1619 work (commit 8108eb256, subject naming card t1619, SPEC-UPDATE-REPAIR-RC30-001), the t1578 work (commit e73a7cbf5, SPEC-UPDATE-MIGRATION-FIX-001 sync close), and the t1591 work (commit a372a984c, SPEC-USERASSET-DEPLOY-GUARD-001 sync backfill) are not ancestors of tree 2aab5f797 at plan time (E-14). The landing SHAs for the run base are named by the leader (G-13).
- Configuration: `.moai/config/sections/interview.yaml` sets `decision_gate: on`, so decision-index.md accompanies this SPEC (E-17).
