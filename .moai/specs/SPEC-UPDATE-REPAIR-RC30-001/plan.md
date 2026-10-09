# SPEC-UPDATE-REPAIR-RC30-001 — Implementation Plan

Milestones are ordered by decision reversibility: the decisions most likely to change come first (the profile-reach decision and the shared-root data model), then user-visible output, then mechanical message fixes. Priority labels are used instead of time estimates.

## A. Context

- Tree at measurement: `2aab5f797b75983e132af451da68f69e3426557b` (`WT-moai-update-repair`).
- Absorbed refs (recorded at M1, not before plan-audit): `WT-user-asset-bundle` (t1591, `cae7e28fe` at measurement), `WT-update-migration-fixes` (t1578, `317ed31b6` at measurement).
- Installed `moai` binary (`4f8aba061`) is not evidence for this tree. Any build used for a measurement is built from this tree and invoked by path.
- Existing SPECs whose decisions bind this plan: SPEC-USER-ASSET-INSTALL-001 (P6, REQ-002), SPEC-CLI-TUX-INIT-UPDATE-001 (REQ-TUXIU-004, -012 to -016, -020 to -022, -041, -054), SPEC-V3R6-UPDATE-NOISE-001 (`--verbose`, REQ-UN-005, REQ-UN-010), SPEC-V3R6-UPDATE-PROGRESS-001 (`tui.ProgressLine`), SPEC-UPDATE-MIGRATION-001 (REQ-UPM-030 to -032, § H out-of-scope notes for t1509 and t1527), SPEC-UPDATE-YAML-PRESERVE-001 (comment preservation).

## B. Milestones

### B.1 Milestone M0 — RED capture (High)

Before any change: capture the RED-now cell of every criterion in `acceptance.md` on tree `2aab5f797`, as single-invocation commands with verbatim stdout, exit code, and tree SHA. Measure the runnable profile-visibility command form (AC-RC30-01) and confirm the `make tui-snapshot` target. Record each result in `progress.md` § E.2 (run-phase evidence).

Exit condition: every criterion has a RED cell or a recorded reason it cannot be captured (Gap).

### B.2 Milestone M1 — Absorption and overlapping items (High, first run-stage milestone)

Runs after plan-audit PASS.

1. Absorb `WT-user-asset-bundle` (t1591) and `WT-update-migration-fixes` (t1578) into this tree with plain `git merge`. Do not rebase. Record both tip SHAs in `progress.md` § G Absorption Record.
2. Measure every range from `CARD_BASE=$(git merge-base HEAD <absorbed-tip>)`, not from a literal SHA. Re-absorb if a branch moves.
3. Resolve conflicts in `internal/userassets/install.go` (both branches) and `internal/cli/update.go` (t1578) by the rule that the branch's own tests keep passing.
4. Implement the BLOCKED-ON items in this order:
   - U-01 is DECISION-PENDING (operator decision Q2; decision-index.md). It is not a BLOCKED-ON item. The removal gate that REQ-RC30-003 requires does not depend on Q2 and is implemented in M1. The reach mechanism waits for the Q2 verdict.
   - U-02 and U-10 (migration backup and emptied-directory removal), U-09 (non-catalog message), U-03 (bundle announce), U-12 (kept assets).
   - U-11 (user-root downgrade refusal, both branches).
   - U-17 (one banner) and U-21 (profile sync resolution, t1578).
   - U-31 (integrity probe, codex-only harness condition, t1578): the codex-only fixture test is written first (RED), then the harness-conditioned fix (GREEN).

Exit condition: the absorbed refs are recorded; AC-RC30-02, -03, -04, -10, -11, -12, -13, -18, -21 pass or have a recorded blocker.

### B.3 Milestone M2 — Shared-root data model and profile reach (High; decision-dependent)

1. Record the § D.1 status table with runnable evidence for the four roots and `~/.moai`. Each Gap cell is replaced with the command and output that closes it.
2. Run the D.3 init path check: build this tree, invoke the built binary by path with `HOME` set to a scratch directory, and list which roots `init` writes.
3. Implement the profile-reach mechanism only after Q2 is answered. Until then, AC-RC30-01 is the gate. No project-side copy is deleted (REQ-RC30-003).

Exit condition: the D.1 table has no Gap cell that the milestone could close; AC-RC30-01 is measured.

### B.4 Milestone M3 — Config data model and restore (High; not gated on M1)

Config save and restore are data-model changes. They do not overlap either branch by file.

1. `ConfigManager.Save` round-trip: unknown keys, `deployment_mode`, and comments survive (U-05, AC-RC30-06).
2. Record-less projects with a migration archive get a visible warning (U-06, AC-RC30-07).
3. `--restore` applies the backed-up `settings.json`, lists uncovered paths, and counts retention per run (U-07, AC-RC30-08).
4. `gate.yaml` comment loss and the `llm.yaml` re-marshal trigger are investigated with their tests first (U-23, U-24, AC-RC30-22, -23). Fixes follow the named cause.

Exit condition: AC-RC30-06, -07, -08 pass; -22 and -23 are RED with the cause named.

### B.5 Milestone M4 — Output TUX and report wording (High; user-visible; not gated on M1)

0. Measured values first (U-08, AC-RC30-09): remove the literal at `update_archive.go` lines 282 and 367 (E-01), and print the no-op outcome from measured counts. U-27 consumes these values, so this step precedes it. The same change updates the characterization expectation at `internal/cli/tuxiu_characterization_test.go:201`, which pins the literal and `Updated 26 files` (E-22). 1. Step progress: N of M, a bar, one ✓ ! ✗ line per step, in place on a TTY, plain lines otherwise (U-26; reuse `tui.ProgressLine` and `tui.Progress`).
2. One-screen closing summary with the five counts, one action line only when needed, the backup location, and one restore command (U-27).
3. Conflicts and warnings above the summary; the measured outcome replaces the pre-merge risk class (U-28).
4. Narrow-terminal and fallback behavior (U-29). The threshold is a named constant in `internal/tui` defaults.
5. File lists collapse to counts; the full list goes behind `--verbose` or to a log path (U-16). `--verbose` keeps the NOISE-001 meaning.
6. Locale key parity (U-30; REQ-RC30-019): the locale tables ko, en, ja, and zh define the same keys, judged by the locale part of AC-RC30-25 (`TestUpdateLocaleKeyParity`).
7. Golden snapshots are regenerated only through `make tui-snapshot`, on a reviewed diff.

Exit condition: AC-RC30-09, -17, -25 pass; `make tui-snapshot` output is reviewed.

### B.6 Milestone M5 — Measured messages and counts (Medium; mechanical)

U-08 (literal removal and the measured no-op outcome) moved to M4, step 0 of § B.5; make the preview deployer equal the run deployer (U-18, AC-RC30-19); skip decision on the template tree hash (U-13, AC-RC30-14); stamp skew warning (U-14, AC-RC30-15).

Exit condition: AC-RC30-14, -15, -19 pass.

### B.7 Milestone M6 — Investigations (Medium)

Each GAP-INVESTIGATE item is closed by its named evidence, with its command and output in research.md: U-15 (AC-RC30-16), U-19 and U-20 (AC-RC30-20), U-25 (AC-RC30-24). U-30 waits for the locale-mechanism read (Q1 is resolved from disk; decision-index.md Q1).

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
- Secured: no new file writes outside the project and the shared roots. The profile directory is not written (REQ-002 of SPEC-USER-ASSET-INSTALL-001, unless Q2 says otherwise).
- Trackable: conventional commits with the card id, as the lane commits (this planning phase commits nothing).

## E. Landing (merge-ready only)

- Order: t1591, then t1578, then t1619. This card ends at merge-ready.
- No card merges until t1616 lands. The landing face after t1616 is local main.
- This plan names no remote push, no develop command, and no merge into main as a landing path.
- Absorption of both branches happens at M1, not at landing, so the merge-ready state includes their content.

## F. Risks

| Risk | Effect | Mitigation |
|------|--------|-----------|
| The profile-visibility premise (project-side copies are seen by profile sessions) is unverified | A removal gate built on a false premise | AC-RC30-01 is measured before any removal; no copy is deleted until it passes |
| REQ-002 of SPEC-USER-ASSET-INSTALL-001 (spec.md line 255) may bar a profile install | U-01 cannot use a reach mechanism that writes into a profile directory | Q2 is raised; U-01 is DECISION-PENDING (Q2), not BLOCKED-ON any branch |
| Both branches change `internal/userassets/install.go` | Merge conflict at M1 | Plain merge; conflicts resolved with the branch tests kept green |
| Branch tips move after absorption | Measurements go stale | Re-absorb on movement; ranges are measured from the merge-base each time |
| Golden churn in the TUX output | Review load | Golden updates only through `make tui-snapshot` on a reviewed diff |
| "16 locales" versus "four locale strings" | Wrong scope for U-30 | Q1 is resolved from disk (decision-index.md Q1); U-30 has no criterion until its locale mechanism is read |
| Track overlap with t1621 and t1622 | Scope creep into factory code | Named under Dependencies in spec.md § H; not items |

## G. Dependencies

Listed in spec.md § H. The plan adds no dependency beyond those.

## H. Out of Scope (plan view)

Out of scope: the factory slug rule (t1621), the factory inbox error (t1622), init-side presentation (not in this item register; queue overlap with t1527 recorded with disposition undecided in spec.md § F), and the mo.ai.kr operator recovery (audit § C).
