# SPEC-UPDATE-REPAIR-RC30-001 — Implementation Plan

Milestones are ordered by decision reversibility: the decisions most likely to change come first (the profile-reach decision and the shared-root data model), then user-visible output, then mechanical message fixes. Priority labels are used instead of time estimates.

## A. Context

- Tree of the RED cells: `8fa0dc24b` (code paths equal main `2aab5f797`). Main = `2aab5f797`. Commits after `00ab5bbb1` changed `internal/cli/migrate_project_assets.go` (`f9a58fa4e`, `00c1b4b3f`); the code paths are unchanged only through `00ab5bbb1` (`git diff --name-only 8fa0dc24b 00ab5bbb1 -- internal cmd` is empty).
- Absorbed refs (recorded at M1, not before plan-audit): `WT-user-asset-bundle` (t1591) and `WT-update-migration-fixes` (t1578), absorbed at run start; the absorption-time tip is recorded in progress.md section G (Absorption Record).
- Installed `moai` binary (`4f8aba061`) is not evidence for this tree. Any build used for a measurement is built from this tree and invoked by path.
- Existing SPECs whose decisions bind this plan: SPEC-USER-ASSET-INSTALL-001 (P6, REQ-002), SPEC-CLI-TUX-INIT-UPDATE-001 (REQ-TUXIU-004, -012 to -016, -020 to -022, -041, -054), SPEC-V3R6-UPDATE-NOISE-001 (`--verbose`, REQ-UN-005, REQ-UN-010), SPEC-V3R6-UPDATE-PROGRESS-001 (`tui.ProgressLine`), SPEC-UPDATE-MIGRATION-001 (REQ-UPM-030 to -032, § H out-of-scope notes for t1509 and t1527), SPEC-UPDATE-YAML-PRESERVE-001 (comment preservation).

## B. Milestones

### B.1 Milestone M0 — RED capture (High)

Before any change: capture the RED-now cell of every criterion in `acceptance.md` on tree `8fa0dc24b` (code paths equal main `2aab5f797`), as single-invocation commands with verbatim stdout, exit code, and tree SHA. Measure the runnable profile-visibility command form (AC-RC30-01, moved to SPEC B) and confirm the `make tui-snapshot` target. Record each result in `progress.md` § E.2 (run-phase evidence). Re-observe every RED cell on HEAD at M1, before absorption; the result replaces the `8fa0dc24b` cells (recorded as a Gap in `progress.md` section E.2).

Exit condition: every criterion has a RED cell or a recorded reason it cannot be captured (Gap).

### B.2 Milestone M1 — Absorption and overlapping items (High, first run-stage milestone)

Runs after plan-audit PASS.

1. Absorb in the order below. Each branch is merged with `git merge --no-commit --no-ff`; plain `git merge` (commit on success) is not used. (1) Absorb `WT-user-asset-bundle` (t1591). (2) Write the TDD RED test for the removal switch on the staged tree. (3) Add `removalEnabled` (default false) around the removal call (its line is recorded at absorption). (4) Commit that merge. (5) Absorb `WT-update-migration-fixes` (t1578) the same way, then commit it. Do not rebase. Route (a) is chosen for the removal call: each branch's content stays whole, with no hunk selection, and the switch is added inside the first merge before its commit; route (b) would select hunks from one file that carries both the predicate fix and the removal call, which breaks the whole-branch merge. 1b. AC-RC30-02, the judging criterion of the switch, is moved to SPEC B; the switch is inert by default (`removalEnabled` false) and has no criterion in this SPEC. Record each absorption-time tip in `progress.md` section G (Absorption Record).
2. Measure every range from `CARD_BASE=$(git merge-base HEAD <absorbed-tip>)`, not from a literal SHA. Re-absorb if a branch moves.
3. Resolve conflicts in `internal/userassets/install.go` (both branches) and `internal/cli/update.go` (t1578) by the rule that the branch's own tests keep passing.
4. Implement the BLOCKED-ON items in this order:
   - U-01 is BLOCKED-ON t1591 (operator decision Q2 is decided as option (c), d-20261010T060848Z-dc3b; decision-index.md). The removal switch of step 1b is inert by default in SPEC A; turning it on belongs to SPEC B (AC-RC30-01 moved there). The reach mechanism follows option (c).
   - U-09 (non-catalog message), U-03 (bundle announce), U-12 (kept assets). U-02 and U-10 moved to SPEC B with AC-RC30-03 and AC-RC30-11 (split, d-20261009T235135Z-5e35).
   - U-11 (user-root downgrade refusal, both branches).
   - U-17 (one banner) and U-21 (profile sync resolution, t1578).
   - U-31 moved to SPEC B; the repair stays with card t1578 (leader correction).

Exit condition: the absorbed refs are recorded; the removal call is inert by default (judged by AC-RC30-02, moved to SPEC B); AC-RC30-04, -10, -12, -13, -18, -21 pass or have a recorded blocker.

### B.3 Milestone M2 — Shared-root data model and profile reach (High; decision-dependent)

1. Record the § D.1 status table with runnable evidence for the four roots and `~/.moai`. Each Gap cell is replaced with the command and output that closes it.
2. Run the D.3 init path check: build this tree, invoke the built binary by path with `HOME` set to a scratch directory, and list which roots `init` writes.
3. Moved to SPEC B (split, d-20261009T235135Z-5e35): the adoption of AC-RC30-01 is no longer a step of this SPEC. Until SPEC B adopts it, the removal switch stays off, no project-side copy is deleted, and no removal code path is turned on (REQ-RC30-003).
4. Moved to SPEC B: turning the switch on, U-02 (backup before removal, AC-RC30-03), and U-10 (emptied-directory removal, AC-RC30-11) belong to SPEC B. Nothing in SPEC A enables them.
5. Implement the profile-reach mechanism under option (c) (U-01); Q2 is decided.

Exit condition: AC-RC30-01, -03 and -11, and items U-02 and U-10, are moved to SPEC B; no criterion of this SPEC is owned by M2; the D.1 table has no Gap cell that the milestone could close.

### B.4 Milestone M3 — Config data model and restore (High; not gated on M1)

Config save and restore are data-model changes. They do not overlap either branch by file.

1. `ConfigManager.Save` round-trip: unknown keys, `deployment_mode`, and comments survive (U-05, AC-RC30-06).
2. Record-less projects with a migration archive get a visible warning (U-06, AC-RC30-07).
3. The user-facing restore does not write the live `.claude/settings.json` and lists it as a not-restored path with a manual copy instruction; it lists uncovered paths and counts retention per run (U-07, AC-RC30-08).
4. `gate.yaml` comment loss and the `llm.yaml` re-marshal trigger are investigated with their tests first (U-23, U-24, AC-RC30-22, -23). Fixes follow the named cause.

Exit condition: AC-RC30-06, -07, -08 pass; AC-RC30-22 and -23 are recorded as blocked with the cause named, per the Definition of Done.

### B.5 Milestone M4 — Output TUX and report wording (High; user-visible; not gated on M1)

0. Measured values first (U-08, AC-RC30-09): remove the literal at `update_archive.go` lines 282 and 367 (E-01), and print the no-op outcome from measured counts. U-27 [MOVED to the follow-up placement F1] consumes these values, so this step precedes it. The same change updates the characterization expectation at `internal/cli/tuxiu_characterization_test.go:201`, which pins the literal and `Updated 26 files` (E-22). 1. Step progress [MOVED to the follow-up placement F1]: N of M, a bar, one ✓ ! ✗ line per step, in place on a TTY, plain lines otherwise (U-26; reuse `tui.ProgressLine` and `tui.Progress`).
2. One-screen closing summary with the five counts, one action line only when needed, the backup location, and one restore command (U-27). [MOVED to the follow-up placement F1]
3. Conflicts and warnings above the summary; the measured outcome replaces the pre-merge risk class (U-28). [MOVED to the follow-up placement F1]
4. Narrow-terminal and fallback behavior (U-29) [MOVED to the follow-up placement F1]. The threshold is a named constant in `internal/tui` defaults.
5. File lists collapse to counts; the full list goes behind `--verbose` or to a log path (U-16). `--verbose` keeps the NOISE-001 meaning.
6. Locale mechanism read, then key parity (U-30; REQ-RC30-019) [MOVED to the follow-up placement F1]: first read how the update output reaches locale strings (E-21 shows the renderer files read no locale setting; the read is recorded in research.md), then make the ko, en, ja, and zh tables define the same keys. The parity check is judged by the locale part of AC-RC30-25 (`TestUpdateLocaleKeyParity`) and runs in this step, after the read.
7. Golden snapshots are regenerated only through `make tui-snapshot`, on a reviewed diff. [MOVED to the follow-up placement F1]

Exit condition: AC-RC30-09 and AC-RC30-17 pass. AC-RC30-25 [MOVED to the follow-up placement F1] is not a condition of this SPEC.

### B.6 Milestone M5 — Measured messages and counts (Medium; mechanical)

U-08 (literal removal and the measured no-op outcome) moved to M4, step 0 of § B.5; make the preview deployer equal the run deployer (U-18, AC-RC30-19); skip decision on the template tree hash (U-13, AC-RC30-14); stamp skew warning (U-14, AC-RC30-15).

Exit condition: AC-RC30-14, -15, -19 pass.

### B.7 Milestone M6 — Investigations (Medium)

Each GAP-INVESTIGATE item is closed by its named evidence, with its command and output in research.md: U-15 (AC-RC30-16), U-19 and U-20 (AC-RC30-20), U-25 (AC-RC30-24). U-30 [MOVED to the follow-up placement F1] has no M6 work: its locale mechanism is read in M4 step 6, before the parity check (§ B.5; Q1 is resolved from disk, decision-index.md Q1).

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
- Absorption of both branches happens at M1, not at landing, so the merge-ready state includes their content.

## F. Risks

| Risk | Effect | Mitigation |
|------|--------|-----------|
| The profile-visibility premise (project-side copies are seen by profile sessions) is unverified | A removal gate built on a false premise | AC-RC30-01 moved to SPEC B; no copy is deleted and no removal path is enabled in SPEC A |
| REQ-002 of SPEC-USER-ASSET-INSTALL-001 (spec.md line 255) may bar a profile install | U-01 cannot use a reach mechanism that writes into a profile directory | Q2 is decided as option (c); U-01 is BLOCKED-ON t1591 |
| Both branches change `internal/userassets/install.go` | Merge conflict at M1 | Plain merge; conflicts resolved with the branch tests kept green |
| Branch tips move after absorption | Measurements go stale | Re-absorb on movement; ranges are measured from the merge-base each time |
| Golden churn in the TUX output | Review load | Golden updates only through `make tui-snapshot` on a reviewed diff |
| "16 locales" versus "four locale strings" | Wrong scope for U-30 | Q1 is resolved from disk (decision-index.md Q1); U-30 [MOVED to the follow-up placement F1] is judged by the locale part of AC-RC30-25, after its mechanism is read in M4 step 6 |
| Track overlap with t1621 and t1622 | Scope creep into factory code | Named under Dependencies in spec.md § H; not items |

## G. Dependencies

Listed in spec.md § H. The plan adds no dependency beyond those.

## H. Out of Scope (plan view)

Out of scope: the factory slug rule (t1621), the factory inbox error (t1622), init-side presentation (not in this item register; queue overlap with t1527 recorded with disposition undecided in spec.md § F), and the mo.ai.kr operator recovery (audit § C).

## 범위 부록

| D-number | item (short name) | corresponding REQ/AC IDs in this SPEC | note |
|---|---|---|---|
| D1 | Shared-root assets vs. profile config directory | REQ-RC30-001, REQ-RC30-002, REQ-RC30-003 (owned by SPEC B; cross-reference); AC-RC30-27 (new, decision d-20261010T084752Z-d4be); AC-RC30-01, AC-RC30-02 (moved to SPEC B) | U-01; Q2 decided as option (c); BLOCKED-ON t1591; judged by AC-RC30-27 |
| D2 | Migration deletes files before the backup step | REQ-RC30-004; AC-RC30-03 (moved to SPEC B) | U-02; moved to SPEC B |
| D3 | Plugin-mode projects receive core assets only | REQ-RC30-005; AC-RC30-04 | U-03; BLOCKED-ON WT-user-asset-bundle (t1591) |
| D4 | Renamed skill unremovable and labelled as opt-in | REQ-RC30-005; AC-RC30-10 | U-09; BLOCKED-ON WT-user-asset-bundle (t1591); AC-RC30-10 covers the non-catalog message only; no REQ or AC here covers the rename/retired map |
| D5 | Emptied skill directories left behind | REQ-RC30-004; AC-RC30-11 (moved to SPEC B) | U-10; moved to SPEC B; no active criterion in this SPEC |
| D6 | Kept optional agents and .agents/skills copies never refreshed | REQ-RC30-001; AC-RC30-13 | U-12; BLOCKED-ON WT-user-asset-bundle (t1591) |
| D7 | Typed save of llm.yaml drops deployment_mode, alias keys and comments | REQ-RC30-007; AC-RC30-06 | U-05; IN-NOW |
| D8 | Record-less project silently re-stamped as local | REQ-RC30-008; AC-RC30-07 | U-06; IN-NOW; follows from D7 |
| D9 | No rollback: two roots, partial coverage, keep-5 eviction, unbounded namespace backups | REQ-RC30-009, REQ-RC30-010; AC-RC30-08 | U-07; IN-NOW; the run identity is the .moai-run marker (design.md section 5) |
| D10 | Shared-root downgrade, string-equal skip, unreliable stamp | REQ-RC30-012, REQ-RC30-022, REQ-RC30-023; AC-RC30-12, AC-RC30-14, AC-RC30-15 | U-11 (BLOCKED-ON t1578 and t1591); U-13 (IN-NOW); U-14 (IN-NOW) |
| D11 | permissions.ask emptied when the template retires the key | REQ-RC30-024; AC-RC30-16 | U-15; GAP-INVESTIGATE; audit status Candidate; cause inferred, not reproduced in this tree |
| D12 | Hardcoded no-op literals in the update output | REQ-RC30-011, REQ-RC30-017; AC-RC30-09 | U-08; IN-NOW |
| D13 | 369 preserved lines printed with no paging | REQ-RC30-015; AC-RC30-17 | U-16; IN-NOW |
| D14 | Double version banner; preview 266 versus run 423 files | REQ-RC30-013, REQ-RC30-011; AC-RC30-18, AC-RC30-19 | U-17 (banner; BLOCKED-ON t1578); U-18 (file count; IN-NOW) |
| D15 | No bundle listing or discovery | REQ-RC30-006; AC-RC30-05 | U-04; IN-NOW |
| D16 | 83 rule references and 17 commands name a non-deployed moai skill path | REQ-RC30-025; AC-RC30-20 | U-19; GAP-INVESTIGATE; AC-RC30-20 is an investigation with a recorded result |
| D17 | docs-site update page lists the agents path as an update target | REQ-RC30-025; AC-RC30-20 | U-20; GAP-INVESTIGATE; the audit line is not reproduced on this tree |
| D18 | Profile sync resolves the profile from the environment variable | REQ-RC30-002; AC-RC30-21 | U-21; BLOCKED-ON WT-update-migration-fixes (t1578) |
| D19 | Legacy root deny specifiers not normalized (`normalizeLegacyRootDenySpecifiers`) | run 목록 추가, 감사 보고서 최소 수리 | U-22; tracked for traceability only, owned by t1578; no REQ or AC in this SPEC |
