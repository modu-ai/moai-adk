# SPEC-UPDATE-REPAIR-RC30-001 — Implementation Plan

Milestones are ordered by decision reversibility: the decisions most likely to change come first (the profile-reach decision and the shared-root data model), then user-visible output, then mechanical message fixes. Priority labels are used instead of time estimates.

## A. Context

- Tree of the RED cells: `6f95917a1` (revision 0.1.4), on which both branches are absorbed (merges `ff2bc876d` and `6f95917a1`). The revision 0.1.3 cells were pinned to `8fa0dc24b` (code paths equal main `2aab5f797` through `00ab5bbb1`); the drift is recorded in acceptance.md. Main = `2aab5f797`. Commits after `00ab5bbb1` changed `internal/cli/migrate_project_assets.go` (`f9a58fa4e`, `00c1b4b3f`).
- Absorbed refs: `WT-user-asset-bundle` (t1591, tip `a372a984c`, merge `ff2bc876d`) and `WT-update-migration-fixes` (t1578, tip `e73a7cbf5`, merge `6f95917a1`), recorded in progress.md § G (Absorption Record).
- Installed `moai` binary (`4f8aba061`) is not evidence for this tree. Any build used for a measurement is built from this tree and invoked by path.
- Existing SPECs whose decisions bind this plan: SPEC-USER-ASSET-INSTALL-001 (P6, REQ-002), SPEC-CLI-TUX-INIT-UPDATE-001 (REQ-TUXIU-004, -012 to -016, -020 to -022, -041, -054), SPEC-V3R6-UPDATE-NOISE-001 (`--verbose`, REQ-UN-005, REQ-UN-010), SPEC-V3R6-UPDATE-PROGRESS-001 (`tui.ProgressLine`), SPEC-UPDATE-MIGRATION-001 (REQ-UPM-030 to -032, § H out-of-scope notes for t1509 and t1527), SPEC-UPDATE-YAML-PRESERVE-001 (comment preservation).

## B. Milestones

### B.1 Milestone M0 — RED capture (High)

Done for tree `6f95917a1` in revision 0.1.4: every criterion's RED cell is re-executed and recorded in acceptance.md (RED ledger, E-01 to E-25), with the line drift from `8fa0dc24b`. The runnable form of AC-RC30-01 is moved to card t1668; the `make tui-snapshot` target is confirmed (E-09). The first step of M1 re-observes every RED cell on the then-current HEAD (§ B.2 step 2), and each result is recorded in `progress.md` § E.2 (run-phase evidence).

Exit condition: every criterion has a RED cell or a recorded reason it cannot be captured (Gap).

### B.2 Milestone M1 — Overlapping items and the removal switch (High, first run-stage milestone)

Runs after the plan-phase gate: the plan audit is recorded as BYPASSED by operator instruction (progress.md § E.1), not as PASS.

1. Absorption is done (spec.md § G.2): both branches are merged at `ff2bc876d` and `6f95917a1`. No absorption step remains in this milestone. If a branch moves, merge it again and record its new tip in `progress.md` § G.
2. Re-observe every RED cell (E-01 to E-25, acceptance.md) on the then-current HEAD before any implementation step, and record the result, with any line drift, in `progress.md` § E.2. A re-observation that differs from the revision 0.1.4 cells is a blocker report, not a silent update.
3. Add the removal switch: `removalEnabled` (default false) around the removal call at `internal/cli/migrate_project_assets.go` (line 174 at `6f95917a1`, E-04). The switch is not on the tree yet. Its TDD RED test and its judging criterion (AC-RC30-02) belong to card t1668; this step adds the switch inert and adds no criterion.
4. Implement the BLOCKED-ON items in this order:
   - U-09 (non-catalog message), U-03 (bundle announce), U-12 (kept assets). U-02 and U-10 are moved to card t1668 with AC-RC30-03 and AC-RC30-11 (split, d-20261009T235135Z-5e35).
   - U-11 (user-root downgrade refusal).
   - U-17 (one banner) and U-21 (profile sync resolution, t1578).
   - U-31 is moved to card t1668; the repair stays with card t1578 (leader correction).
   - U-01 is implemented in M2 (§ B.3), not here.

Exit condition: the absorbed refs are recorded (spec.md § G.2); the removal switch is present and inert by default (no criterion here; AC-RC30-02 is card t1668's); AC-RC30-04, -10, -12, -13, -18 and -21 pass or have a recorded blocker.

### B.3 Milestone M2 — Shared-root data model, profile reach, and the keep behavior (High; decision-dependent)

1. Record the § D.1 status table with runnable evidence for the four roots and `~/.moai`. Each Gap cell is replaced with the command and output that closes it.
2. Run the D.3 init path check: build this tree, invoke the built binary by path with `HOME` set to a scratch directory, and list which roots `init` writes.
3. Moved to card t1668 (split, d-20261009T235135Z-5e35): the adoption of AC-RC30-01 is no longer a step of this SPEC. Until card t1668 adopts it, the removal switch stays off, no project-side copy is deleted, and no removal code path is turned on (REQ-RC30-003).
4. Moved to card t1668: turning the switch on, U-02 (backup before removal, AC-RC30-03), and U-10 (emptied-directory removal, AC-RC30-11) belong to card t1668. Nothing in SPEC A enables them.
5. Implement the profile-reach mechanism under option (c) (U-01): the keep field `FileEntry.KeptBy`, its constant `manifest.KeptByProfileReach`, the keep behavior in the migration, and the keep marking. This milestone owns AC-RC30-27, because it is the first milestone that introduces the keep field, the constant, and the keep behavior.
6. AC-RC30-27 behavioral RED: after an API stub that defines `FileEntry.KeptBy` and `manifest.KeptByProfileReach` but gives them no keep behavior, run the criterion and record the four-element RED cell (command, verbatim stdout, exit status, tree SHA) in `progress.md` § E.2. The build-red observation E-25 is not this RED cell.
7. AC-RC30-28 reach judgment: on a profile-mapped project, run `moai update`, then `moai doctor`, with the binary built from this tree and invoked by path. Record the doctor output in `progress.md` § E.2 (Gap G-28 until recorded).

Exit condition: AC-RC30-27 passes (keep half only; the removal half is card t1668's); AC-RC30-28 is recorded (regression-guard until the profile-session run is recorded); AC-RC30-01, -03 and -11, and items U-02 and U-10, are moved to card t1668; the D.1 table has no Gap cell that this milestone could close.

### B.4 Milestone M3 — Config data model, restore, and bundle listing (High; not gated on M1)

Config save and restore are data-model changes. They do not overlap either branch by file.

1. `ConfigManager.Save` round-trip: unknown keys, `deployment_mode`, and comments survive (U-05, AC-RC30-06). The field the round-trip test sets before Save is Gap G-06.
2. Record-less projects with a migration archive get a visible warning (U-06, AC-RC30-07).
3. The user-facing restore does not write the live `.claude/settings.json` and lists it as a not-restored path with a manual copy instruction; it lists uncovered paths and counts retention per run (U-07, AC-RC30-08).
4. `gate.yaml` comment loss and the `llm.yaml` re-marshal trigger are investigated with their tests first (U-23, U-24, AC-RC30-22, -23). Fixes follow the named cause.
5. `moai bundle list` prints every catalog bundle with its skill and agent counts, and `moai bundle add --help` names the valid bundles (U-04, AC-RC30-05). This step is the owner of U-04.

Exit condition: AC-RC30-05, -06, -07 and -08 pass; AC-RC30-22 and -23 are recorded as blocked with the cause named, per the Definition of Done.

### B.5 Milestone M4 — Output TUX and report wording (High; user-visible; not gated on M1)

0. Measured values first (U-08, AC-RC30-09): remove the literal at `update_archive.go` lines 282 and 367 (E-01), and print the no-op outcome from measured counts. U-27 [MOVED to card t1669] consumes these values, so this step precedes it. The same change updates the characterization expectation at `internal/cli/tuxiu_characterization_test.go:201`, which pins the literal and `Updated 26 files` (E-22). 1. Step progress [MOVED to card t1669]: N of M, a bar, one ✓ ! ✗ line per step, in place on a TTY, plain lines otherwise (U-26; reuse `tui.ProgressLine` and `tui.Progress`).
2. One-screen closing summary with the five counts, one action line only when needed, the backup location, and one restore command (U-27). [MOVED to card t1669]
3. Conflicts and warnings above the summary; the measured outcome replaces the pre-merge risk class (U-28). [MOVED to card t1669]
4. Narrow-terminal and fallback behavior (U-29) [MOVED to card t1669]. The threshold is a named constant in `internal/tui` defaults, set to 60 columns (decision Q7, applied default).
5. File lists collapse to counts; the full list goes behind `--verbose` or to a log path (U-16). `--verbose` keeps the NOISE-001 meaning.
6. Locale mechanism read, then key parity (U-30; REQ-RC30-019) [MOVED to card t1669]: first read how the update output reaches locale strings (E-21 shows the renderer files read no locale setting; the read is recorded in research.md), then make the ko, en, ja, and zh tables define the same keys. The parity check is judged by the locale part of AC-RC30-25 (`TestUpdateLocaleKeyParity`) and runs in this step, after the read.
7. Golden snapshots are regenerated only through `make tui-snapshot`, on a reviewed diff. [MOVED to card t1669]

Exit condition: AC-RC30-09 and AC-RC30-17 pass. AC-RC30-25 [MOVED to card t1669] is not a condition of this SPEC.

### B.6 Milestone M5 — Measured messages and counts (Medium; mechanical)

U-08 (literal removal and the measured no-op outcome) moved to M4, step 0 of § B.5; make the preview deployer equal the run deployer (U-18, AC-RC30-19); skip decision on the template tree hash (U-13, AC-RC30-14); stamp skew warning (U-14, AC-RC30-15).

Exit condition: AC-RC30-14, -15, -19 pass.

### B.7 Milestone M6 — Investigations (Medium)

Each GAP-INVESTIGATE item is closed by its named evidence, with its command and output in research.md: U-15 (AC-RC30-16), U-19 and U-20 (AC-RC30-20). U-25 (AC-RC30-24) is deferred to the operator's rc.30 field run per ruling d-20261010T223350Z-7f6e and is not closed by evidence here. U-30 [MOVED to card t1669] has no M6 work: its locale mechanism is read in M4 step 6, before the parity check (§ B.5; Q1 is resolved from disk, decision-index.md Q1).

### B.8 Milestone M7 — Verification and merge-ready (High)

1. Run `go test` for each touched package with `-count=1 -v`, and record the `--- PASS` lines.
2. Measure coverage with `go test -cover` on each touched package and record the figure with its command. The target is 85% or more per package.
3. Run `make tui-snapshot` and confirm the golden diff is the reviewed one.
4. State the card as merge-ready. Do not merge. The landing sequence is set in § E.

## C. Technical approach

- The four shared roots (`internal/userassets/paths.go:36-43`) remain the install source. The migration gains a backup-before-removal step, and its removal is gated by the profile-visibility check.
- Output is one renderer in `internal/cli/update_tux.go`. It takes measured values only. It uses the existing `tui` glyph set (✓, !, ✗ per REQ-TUXIU-004), and it reuses `tui.Progress` and `tui.ProgressLine`.
- Config writes stop round-tripping through the typed struct alone. Unknown keys and comments are preserved (U-05).
- Thresholds (narrow width) and retention counts live in named constants, not inline literals.

## D. Quality (TRUST 5)

- Tested: each touched package reaches 85% or more statement coverage (measured in M7). The packages are `internal/cli`, `internal/cli/update`, `internal/cli/update/report`, `internal/config`, `internal/userassets`, `internal/backup`, and `internal/tui`.
- Readable: no new identifiers in the output that do not match existing style. Comments follow `code_comments: en`.
- Unified: `gofmt` and the project linter (`golangci-lint`) run on touched packages.
- Secured: no new file writes outside the project and the shared roots. The profile directory is not written (REQ-002 of SPEC-USER-ASSET-INSTALL-001); option (c) writes nothing there.
- Trackable: conventional commits with the card id, as the lane commits (this planning phase commits nothing).

## E. Landing (merge-ready only)

- Order: t1591, then t1578, then t1619. This card ends at merge-ready.
- No card merges until t1616 lands. The landing face after t1616 is local main.
- This plan names no remote push, no develop command, and no merge into main as a landing path.
- Absorption of both branches is done (spec.md § G.2; merges `ff2bc876d` and `6f95917a1`), so the merge-ready state includes their content. A later branch move is absorbed again before landing.

## F. Risks

| Risk | Effect | Mitigation |
|------|--------|-----------|
| The profile-visibility premise (project-side copies are seen by profile sessions) is unverified | A removal gate built on a false premise | AC-RC30-01 moved to card t1668; no copy is deleted and no removal path is enabled in SPEC A |
| REQ-002 of SPEC-USER-ASSET-INSTALL-001 (spec.md line 255) may bar a profile install | U-01 cannot use a reach mechanism that writes into a profile directory | Q2 is decided as option (c); U-01 is BLOCKED-ON t1591 (absorbed; spec.md § G.2) |
| Both branches change `internal/userassets/install.go` | Merge conflict at absorption (done) | Absorbed (spec.md § G.2); conflicts resolved with the branch tests kept green |
| Branch tips move after absorption | Measurements go stale | Re-absorb on movement; ranges are measured from the merge-base each time |
| Golden churn in the TUX output | Review load | Golden updates only through `make tui-snapshot` on a reviewed diff |
| "16 locales" versus "four locale strings" | Wrong scope for U-30 | Q1 is resolved from disk (decision-index.md Q1); U-30 [MOVED to card t1669] is judged by the locale part of AC-RC30-25, after its mechanism is read in M4 step 6 |
| Track overlap with t1621 and t1622 | Scope creep into factory code | Named under Dependencies in spec.md § H; not items |

## G. Dependencies

Listed in spec.md § H. The plan adds no dependency beyond those.

## H. Out of Scope (plan view)

Out of scope: the factory slug rule (t1621), the factory inbox error (t1622), init-side presentation (not in this item register; queue overlap with t1527 recorded with disposition undecided in spec.md § F), and the mo.ai.kr operator recovery (audit § C).

## 범위 부록

| D-number | item (short name) | corresponding REQ/AC IDs in this SPEC | note |
|---|---|---|---|
| D1 | Shared-root assets vs. profile config directory | REQ-RC30-001, REQ-RC30-002, REQ-RC30-003 (owned by card t1668; cross-reference); AC-RC30-27 (keep, new, decision d-20261010T084752Z-d4be); AC-RC30-28 (reach, new in 0.1.4); AC-RC30-01, AC-RC30-02 (moved to card t1668) | U-01; Q2 decided as option (c); BLOCKED-ON t1591 (absorbed); judged by AC-RC30-27 and AC-RC30-28 |
| D2 | Migration deletes files before the backup step | REQ-RC30-004; AC-RC30-03 (moved to card t1668) | U-02; moved to card t1668 |
| D3 | Plugin-mode projects receive core assets only | REQ-RC30-005; AC-RC30-04 | U-03; BLOCKED-ON WT-user-asset-bundle (t1591, absorbed) |
| D4 | Renamed skill unremovable and labelled as opt-in | REQ-RC30-005; AC-RC30-10 | U-09; BLOCKED-ON WT-user-asset-bundle (t1591, absorbed); AC-RC30-10 covers the non-catalog message only; no REQ or AC here covers the rename/retired map |
| D5 | Emptied skill directories left behind | REQ-RC30-004; AC-RC30-11 (moved to card t1668) | U-10; moved to card t1668; no active criterion in this SPEC |
| D6 | Kept optional agents and .agents/skills copies never refreshed | REQ-RC30-001; AC-RC30-13 | U-12; BLOCKED-ON WT-user-asset-bundle (t1591, absorbed) |
| D7 | Typed save of llm.yaml drops deployment_mode, alias keys and comments | REQ-RC30-007; AC-RC30-06 | U-05; IN-NOW |
| D8 | Record-less project silently re-stamped as local | REQ-RC30-008; AC-RC30-07 | U-06; IN-NOW; follows from D7 |
| D9 | No rollback: two roots, partial coverage, keep-5 eviction, unbounded namespace backups | REQ-RC30-009, REQ-RC30-010; AC-RC30-08 | U-07; IN-NOW; the run identity is the .moai-run marker (design.md section 5) |
| D10 | Shared-root downgrade, string-equal skip, unreliable stamp | REQ-RC30-012, REQ-RC30-022, REQ-RC30-023; AC-RC30-12, AC-RC30-14, AC-RC30-15 | U-11 (BLOCKED-ON t1578 and t1591, both absorbed); U-13 (IN-NOW); U-14 (IN-NOW) |
| D11 | permissions.ask emptied when the template retires the key | REQ-RC30-024; AC-RC30-16 | U-15; GAP-INVESTIGATE; audit status Candidate; cause inferred, not reproduced in this tree |
| D12 | Hardcoded no-op literals in the update output | REQ-RC30-011, REQ-RC30-017; AC-RC30-09 | U-08; IN-NOW |
| D13 | 369 preserved lines printed with no paging | REQ-RC30-015; AC-RC30-17 | U-16; IN-NOW |
| D14 | Double version banner; preview 266 versus run 423 files | REQ-RC30-013, REQ-RC30-011; AC-RC30-18, AC-RC30-19 | U-17 (banner; BLOCKED-ON t1578, absorbed); U-18 (file count; IN-NOW) |
| D15 | No bundle listing or discovery | REQ-RC30-006; AC-RC30-05 | U-04; IN-NOW; step 5 of § B.4 (M3) |
| D16 | 83 rule references and 17 commands name a non-deployed moai skill path | REQ-RC30-025; AC-RC30-20 | U-19; GAP-INVESTIGATE; AC-RC30-20 is an investigation with a recorded result |
| D17 | docs-site update page lists the agents path as an update target | REQ-RC30-025; AC-RC30-20 | U-20; GAP-INVESTIGATE; the audit line is not reproduced on this tree |
| D18 | Profile sync resolves the profile from the environment variable | REQ-RC30-002; AC-RC30-21, AC-RC30-28 | U-21; BLOCKED-ON WT-update-migration-fixes (t1578, absorbed) |
| D19 | Legacy root deny specifiers not normalized (`normalizeLegacyRootDenySpecifiers`) | run 목록 추가, 감사 보고서 최소 수리 | U-22; tracked for traceability only, owned by t1578; no REQ or AC in this SPEC |
