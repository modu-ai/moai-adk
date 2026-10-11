# SPEC-ZONE-SHELL-CD-TRACKING-001 — progress

Card: t1584 (security P1, operator-approved expansion axis, Class C) · Lane: lane-25 · Branch: WT-zone-cd-deny · Phase: plan
Parent: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 (completed, cross-reference only) · Predecessor: SPEC-ZONE-SHELL-PARSING-001 (completed, cd class separated at v0.4.0)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-09
artifacts: spec.md + plan.md + acceptance.md + progress.md + decision-index.md (Tier M set + decision gate artifact)
note: the audit-ready flip of this line is owned by the lane after the independent plan-audit verdict lands (`.moai/reports/t1584/`); §E.2–§E.4 below are phase-owned placeholders (manager-develop / manager-docs) and MUST NOT be populated at plan phase.

## §E.2 Run-phase Evidence

Scope of this invocation: M1 steps 1–5 only (RED-first reproduction, then the fixes K1, K3, K2). M2 and M3 were not run. Raw verbatim captures (machine-local, gitignored): `.moai/reports/t1584/red-reproduction-m1.md`, `.moai/reports/t1584/green-m1.md`, and the raw files `green-m1-matrix-raw.txt`, `green-m1-family-raw.txt`, `green-m1-zone-surface-raw.txt` in the same directory.

**Commits (branch WT-zone-cd-deny).** `9262172f8` test(hook): reproduce cd-tracking RED shapes for SPEC-ZONE-SHELL-CD-TRACKING-001 (t1584) — the RED reproduction only. `6f4a56af8` fix(hook): track cd destinations in the protected-zone shell guard (t1584) — K1, K3, K2.

**Home of the RED reproduction.** New file `internal/hook/protected_zone_shell_cd_repro_test.go`, test `TestProtectedZoneShellCdTrackingRepro`, three subtests (`shape1_dashdash_relative`, `shape2_absolute_in_project`, `shape3_dashdash_hyphen`). Reason: the in-project absolute shape needs the fixture root, which the matrix runner's static cell command cannot carry (plan M1 step 1 fallback). The landed matrix file is unchanged in M1.

**RED capture (E2 and E8) at `9262172f8`, the state before the fix commit.** One invocation per shape, command `go test ./internal/hook/ -run '^TestProtectedZoneShellCdTrackingRepro$/^<subtest>$' -count=1 -v -timeout 5m`. Exit code `1` for each shape. Verbatim stdout:

Shape 1 — `cd -- zone_dir && rm a.log`:

```text
=== RUN   TestProtectedZoneShellCdTrackingRepro
=== RUN   TestProtectedZoneShellCdTrackingRepro/shape1_dashdash_relative
    protected_zone_shell_cd_repro_test.go:63: cmd="cd -- zone_dir && rm a.log" decision="allow" reason=""
    protected_zone_shell_cd_repro_test.go:65: "cd -- zone_dir && rm a.log": decision="allow" reason="", want deny — the cd destination was not tracked (the bare -- separator is counted as a second cd word, so zoneNextCwd resets the directory set)
--- FAIL: TestProtectedZoneShellCdTrackingRepro (0.01s)
    --- FAIL: TestProtectedZoneShellCdTrackingRepro/shape1_dashdash_relative (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.444s
FAIL
```

Shape 2 — `cd <fixture-root>/zone_dir && rm a.log` (the fixture root is the physical temp path; elided below only in the duplicated path, full text in `red-reproduction-m1.md`):

```text
=== RUN   TestProtectedZoneShellCdTrackingRepro
=== RUN   TestProtectedZoneShellCdTrackingRepro/shape2_absolute_in_project
    protected_zone_shell_cd_repro_test.go:63: cmd="cd /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestProtectedZoneShellCdTrackingReproshape2_absolute_in_project170180304/001/zone_dir && rm a.log" decision="allow" reason=""
    protected_zone_shell_cd_repro_test.go:65: "cd /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestProtectedZoneShellCdTrackingReproshape2_absolute_in_project170180304/001/zone_dir && rm a.log": decision="allow" reason="", want deny — the cd destination was not tracked (an in-project absolute destination takes zoneNextCwd's absolute reset, so the directory set degenerates to the root)
--- FAIL: TestProtectedZoneShellCdTrackingRepro (0.00s)
    --- FAIL: TestProtectedZoneShellCdTrackingRepro/shape2_absolute_in_project (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.405s
FAIL
```

Shape 3 — `cd -- -zone && rm a.log` against a `-zone/` fixture:

```text
=== RUN   TestProtectedZoneShellCdTrackingRepro
=== RUN   TestProtectedZoneShellCdTrackingRepro/shape3_dashdash_hyphen
    protected_zone_shell_cd_repro_test.go:63: cmd="cd -- -zone && rm a.log" decision="allow" reason=""
    protected_zone_shell_cd_repro_test.go:65: "cd -- -zone && rm a.log": decision="allow" reason="", want deny — the cd destination was not tracked (a hyphen-leading operand after -- is read as an option and zoneNextCwd resets the directory set)
--- FAIL: TestProtectedZoneShellCdTrackingRepro (0.00s)
    --- FAIL: TestProtectedZoneShellCdTrackingRepro/shape3_dashdash_hyphen (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.464s
FAIL
```

Each shape is red for the stated reason: its subtest runs (a `=== RUN` line), and its verdict is the directory-set reset (`allow`, empty reason).

**E1 — verdict per shape.** Before (`9262172f8`): `decision="allow" reason=""` for all three shapes. After (`6f4a56af8`): `decision="deny"` with category `probe_zone` and path `zone_dir/a.log` (shapes 1 and 2) and `-zone/a.log` (shape 3). GREEN per-shape runs at `6f4a56af8` exit `0` each (verbatim blocks in `green-m1.md`).

**E3 — landed matrix at `6f4a56af8`.** `go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v -timeout 5m`, exit `0`. Count keys: `=== RUN` 61 (the parent plus 60 cells); `--- PASS` 61; `--- FAIL` 0; lines containing `cd_track_` 0; lines containing `cd_` 10 (the five landed cells, unchanged).

**AC-ZSCD-005 family and zone surface at `6f4a56af8`.** `go test ./internal/hook/ -run '^TestProtectedZone$' -count=1 -v -timeout 5m`: exit `0`, `=== RUN` 21, `--- PASS` 21, `--- FAIL` 0. The zone surface (the landed family, matrix, parsing-bypass, fast-path, invalid-manifest, and backslash tests, plus the new reproduction): exit `0`, `=== RUN` 97, `--- PASS` 97, `--- FAIL` 0.

**E4, E5, gofmt at `6f4a56af8`.** `go vet ./internal/hook/`: exit `0`, no output. `GOOS=windows GOARCH=amd64 go build ./internal/hook/`: exit `0`. `gofmt -l` on both changed files: no output.

**E6 — working tree and history.** `git status --short`: ` M .moai/specs/SPEC-ZONE-SHELL-CD-TRACKING-001/progress.md` only (this section; unstaged; the engine-written lines at the end are untouched). `git log --oneline -4`: `6f4a56af8`, `9262172f8`, `bf9532147`, `b02d2741f`.

**E7 — preserved-behavior proof.** `git diff bf9532147 6f4a56af8 -- internal/hook/protected_zone_shell_matrix_test.go`: empty (zero changed lines). `git diff --stat bf9532147 6f4a56af8` over every landed zone test file: empty. Files changed in the run: `internal/hook/protected_zone_shell.go` (the guard) and `internal/hook/protected_zone_shell_cd_repro_test.go` (new).

**B2 pre-scan.** Every hit is in `.moai/reports/t1584/b2-prescan-m1.txt` (41 lines, verbatim). None is in the zone surface. The hits are retired-agent and retired-event code (`subagent_start.go`, `agent_start_test.go`, `retired_events.go`, `audit_test.go`, `stale_run_gate_test.go`, the `session_start_*` tests, `factory_bind_cache_test.go`) and unrelated "superseded" text (`branch_guard.go`, `closure_push_test.go`, `dangerous_removal_test.go`, `handoff/*`, `quality/*`).

**B3 boundary.** Plan-filtered check (non-test files, comment lines removed): one line, `internal/hook/pre_tool.go:856` (`if input.ToolName == "AskUserQuestion" {`). It is pre-existing, unchanged by this run, and compares a tool name rather than invoking the tool.

**Decisions and plan deviations (for review).**
1. K2 lexical rule. The plan's "exact-prefix strip" is realized by `zoneLexicalRel`, the REQ-SIPZ-006 function the file-tool path already uses. It folds ASCII case on a segment boundary. Reason: one lexical semantics across both guards, and macOS paths are case-insensitive. No pinned cell depends on case. Reviewer to confirm.
2. K3 locus. The post-separator spelling `./<word>` is applied in the cd branch, where the separator is known. `zoneNextCwd` keeps the option reset for words before the separator and for the bare `-`.
3. K2 plumbing: `zoneWalker` carries the project root (set in `checkProtectedZoneShell`); `zoneNextCwd` takes it as a parameter.
4. The landed matrix is unchanged in M1. AC-ZSCD-004 (M2) still needs the `cd_track_` group in the matrix.

**Residual risk and decision needed (not changed in M1).**
- Option before the operand: `cd -P zone_dir && rm a.log` and `cd -P -- zone_dir && rm a.log` answer allow (observed in the edge probe), while bash moves into the zone. This is the pre-existing option reset (plan §B K3, option territory); K1 does not touch it. REQ-ZSCD-001's literal wording ("the word after the separator as cd's single directory operand") can be read to cover `cd -P -- dir`. Operator decision needed: track the operand after the known cd options (-L, -P, -e, -@), or keep the reset and record it in the Out of Scope accepted-under-match list.
- Lexical `..` and symlinks: `path.Clean` does not consult the filesystem. A symlink component followed by `..` is judged lexically, as the relative cd class already is. Not probed; by code reading.
- Symlinked project-root spelling (for example `/var` against `/private/var` on macOS): the lexical strip sees it as outside the root and keeps the reset. Not probed; by code reading.

**Gaps.** golangci-lint not run (M3). The full hook package suite not run (M3, under a slot lease). The M3 batch not run (AC-ZSCD-006). The M2 `cd_track_` group not added (AC-ZSCD-004). The Pre-Edit and Pre-Spawn sync checks not run (isolated worktree, exempt; no network operation).

**M2 — cd-class regression group (AC-ZSCD-004, REQ-ZSCD-005).** Commit `ec1eec308` `test(hook): add cd-tracking regression group to the protected-zone matrix (t1584)`, on pre-run HEAD `6f4a56af8`. Full record: `.moai/reports/t1584/green-m2.md`.

**Home and fixture root.** The group sits in the landed matrix (`zoneParsingMatrixCells`, decision Q3 default). The in-project absolute cell carries a `{ROOT}` placeholder, and `TestProtectedZoneShellParsingMatrix` substitutes the fixture root after the handler is built (one added statement, the M1 repro's `filepath.Join(root, …)` approach applied in the runner). No landed line changed: `git diff --numstat 6f4a56af8 HEAD -- internal/hook/protected_zone_shell_matrix_test.go` = `19 0`.

**Cells (6, all deny, no allow control).** `cd_track_dashdash_rm`, `cd_track_dashdash_cp`, `cd_track_dashdash_mv` (REQ-ZSCD-001, `cd -- zone_dir && <verb> …`); `cd_track_dashdash_hyphen_rm` (REQ-ZSCD-001, `-zone/` fixture); `cd_track_absolute_in_project_rm` (REQ-ZSCD-002); `cd_track_dot_slash_spelling_pin` (REQ-ZSCD-001, `sub/-zone/` fixture, the `./<word>` spelling from a tracked non-root cwd). The relative preserve shape is not repeated: `bare_cd_tracking_control` (line 127) is its one count, and it fails under a relative-tracking-disabled mutant.

**RED on the pre-fix guard (`bf9532147`).** 6 of 6 `cd_track_` cells FAIL, each with `decision="allow" reason=""` (the directory-set reset).

**Run (E2).** `go test ./internal/hook/ -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v -timeout 5m` at `ec1eec308`: exit 0; `=== RUN` 67, `--- PASS` 67, `--- FAIL` 0; `cd_track_` executed 6 (6 RUN, 6 PASS). Raw output: `.moai/reports/t1584/green-m2-matrix-raw.txt`.

**Mutation evidence (E7; scratch trees, not committed).** K1 reversed: 5 `cd_track_` FAIL. K3 spelling dropped: 2 FAIL. K2 reversed: 1 FAIL. Non-root join dropped: the pin FAILs and the landed `recursion_unbounded_loop` FAILs. Relative tracking disabled: 6 `cd_track_` FAIL and 3 landed FAIL. Fix tree: 0 FAIL.

**E4, E5.** `go vet ./internal/hook/` exit 0. `GOOS=windows GOARCH=amd64 go build ./internal/hook/` exit 0.

**Gaps.** golangci-lint, the full hook suite, and the scrubbed batch not run (M3). The M1 residual risk 6 that the run brief cites is not on disk; the spelling pin was built from the spelling's observable effect. A Windows test-binary compile not run. §E.3 left at the M1 state, per the instruction.

**Residual risk.** A pre-cd world that already covers a zone masks a cell (the new cells avoid this). The symlinked project-root spelling stays an M1 residual. No allow control, so over-denial is not caught here. The M1 repro file repeats the three shapes; B10 keeps it.

**Recorded deviation (leader ruling d-20261010T005950Z-2a72):** K2 uses zoneLexicalRel (segment boundary, ASCII case-insensitive, REQ-SIPZ-006 semantics) instead of the plan's exact-prefix strip; it denies a superset, the safe direction; the SPEC text is not edited; the sync audit judges it.

**Evidence backfill (leader ruling d-20261010T005950Z-2a72).** After kickoff, the EV-ZSCD-002 header in acceptance.md line 222 changed from "TO BE CAPTURED at M1 step 1" to "CAPTURED at M1 step 1", with the pointers to red-reproduction-m1.md and green-m1.md. No requirement or criterion text changed. The plan artifact hash moved from 3feeac35a8abca958356556dc241bd7e29a78fc19d658a9e9e60610c3729e6fc (round-7 verdict) to b011988ad4ab3c97f1f7f6ca7b9a0f0b691bfb69231b8266f619fa1fdd1e2447 (audit_cache compute_hash at 01:14Z). The change is permitted by the ruling and is recorded here.

### M3 — scoped verification batch (run record)

Tree: HEAD `ec1eec308` (branch `WT-zone-cd-deny`), unchanged through the run; no commit was made in this run. Full record: `.moai/reports/t1584/green-m3.md` (gitignored). Raw captures in the same directory: `green-m3-test-raw.txt`, `green-m3-lint-head.txt`, `green-m3-lint-baseline.txt`, `green-m3-matrix-raw.txt`.

- **Slot lease.** `moai slot acquire --resource hook-suite-t1584 --max-duration 35m` → `slot hook-suite-t1584 acquired by 34078ca9-9e22-4f4c-86ff-8219d6280689 until 2026-10-10T01:51:48Z` (exit 0). Release → `slot hook-suite-t1584 released (was 34078ca9-9e22-4f4c-86ff-8219d6280689)` (exit 0).
- **Test batch, exit 0.** `unset MOAI_KANBAN_ID MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -timeout 30m -count=1` → verbatim `ok  	github.com/modu-ai/moai-adk/internal/hook	298.839s`. Failing tests: none. Passing-test count: not observed, because the batch is non-verbose (record, Gap G1).
- **go vet, exit 0.** No output.
- **golangci-lint, exit 0.** HEAD: `0 issues.` Baseline bf9532147 (`git archive` export; `.golangci.yml` identical between the two): `0 issues.` NEW findings: none. No fix commit was needed.
- **Windows cross-build, exit 0.** `GOOS=windows GOARCH=amd64 go build ./internal/hook/`.
- **AC-ZSCD-004 observation (matrix, verbose, scoped run, exit 0).** `=== RUN` 67, `--- PASS` 67, `--- FAIL` 0. The six `cd_track_` cells each executed and printed `--- PASS` (names in the record). A PASS is a deny verdict in category `probe_zone`: the cells use the deny-probe want value, which the runner asserts at `internal/hook/protected_zone_shell_matrix_test.go` lines 246-249. This matches the M2 run.
- **AC-ZSCD-006 observation.** All batch items above exit 0, with no NEW findings against the M1-entry baseline. No acceptance cell was flipped in this run; the flip is for the lane to request from the leader.

**Follow-up (leader ruling d-20261010T013712Z-06cc).** The AC-ZSCD-004 and AC-ZSCD-006 green-path cells were flipped as an evidence backfill, as in ruling 2a72(3): manager-spec changed those two cells and pointed them at the captured records (green-m2-matrix-raw.txt; green-m3-test-raw.txt, green-m3-lint-head.txt, green-m3-lint-baseline.txt, green-m3-matrix-raw.txt, green-m3.md). No requirement or criterion text changed. The acceptance plan artifact hash moved from b011988ad4ab3c97f1f7f6ca7b9a0f0b691bfb69231b8266f619fa1fdd1e2447 to df6c20dd9c5009391859cf3e6ba7ebcc06ba6084d01f20e35773336ca6b6ed8e (audit_cache compute_hash, measured after the flip).

**M1 reproduction file removed (ruling 06cc(b)).** internal/hook/protected_zone_shell_cd_repro_test.go is removed in this commit. Its three shapes are already covered by the M2 cells cd_track_dashdash_rm, cd_track_absolute_in_project_rm, and cd_track_dashdash_hyphen_rm (same command and fixture), so no matrix cell was added. Its RED stays witnessed at 9262172f8. Checks after the removal: go vet ./internal/hook/ exit 0; scoped matrix (go test -run '^TestProtectedZoneShellParsingMatrix$' -count=1 -v) RUN 67, PASS 67, FAIL 0, six cd_track_ cells PASS (green-m4-matrix-raw.txt; ok github.com/modu-ai/moai-adk/internal/hook 1.896s); golangci-lint run ./internal/hook/... exit 0, output "0 issues." (green-m4-lint.txt). The full package suite was not re-run after the removal; the full-suite verdict is CI's.

**Section placement (ruling 06cc(d)).** The M3 run record moved here from §E.3, so the verbatim M3 outputs that the SPEC cites now sit in this section. §E.3 was set to run_status complete and audit_ready true by manager-develop, the section owner, on the leader's order (run_commit_sha 8089304e6), in the commit that also records the in-progress status.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
audit_ready: true
run_scope: "M1-M3 complete: RED 9262172f8, fix 6f4a56af8, cd_track_ group ec1eec308; M3 scoped batch recorded in section E.2; cell backfill, repro removal, and record move in 8089304e6"
run_complete_at: 2026-10-10
m1_complete_at: 2026-10-10
m1_red_commit: 9262172f8
m1_fix_commit: 6f4a56af8
run_commit_sha: 8089304e6
ac_pass_count: 6
ac_pass_list: AC-ZSCD-001, AC-ZSCD-002, AC-ZSCD-003, AC-ZSCD-004, AC-ZSCD-005, AC-ZSCD-006
ac_pending: none
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not run (isolated card worktree; the pre-edit sync check is exempt there)
l44_post_push_fetch: not applicable (no push in this invocation)
new_warnings_or_lints_introduced: none from go vet or gofmt; golangci-lint 0 issues (M3 at HEAD; scoped run after the repro removal)
cross_platform_build:
  windows_amd64_hook_build: pass (exit 0 at 6f4a56af8)
total_run_phase_files: 2
m1_to_mN_commit_strategy: commit 1 RED test only (9262172f8), commit 2 fix K1/K3/K2 (6f4a56af8), commit 3 M2 cd_track_ group (ec1eec308), commit 4 cell backfill, repro removal, and record move (8089304e6); M3 ran without a commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
audit_ready: true
sync_commit_sha: 430d9f5a28ffc4903f66017987c9a44ad36c796c
sync_scope: spec.md status completed in the sync commit (in-progress to implemented to completed, one commit); section E.4 sync signal; no CHANGELOG or docs-site change
sync_audit_backend: codex_audit (target baseBranch, project_root = card tree)
```

**Lifecycle.** The status chain in-progress to implemented to completed is recorded in the sync commit, per the Status Transition Ownership Matrix.

**Gaps (carried to the completion report).**
- 삭제 후 패키지 전체 스위트 미재실행, 전체 판정은 CI
- cd -P 옵션형 under-match

**Deviation (carried to the completion report).** The draft-to-in-progress transition was not recorded at the first run-phase commit (M1). It is recorded late by its owner, manager-develop, in commit 5e464c950, on the leader's order.

## Run-phase delta record (card t1584, 5d7c case 2)

- Absorb: merge commit e906b4f60 (parents 93915b71f = card side, 478285ab4 = local main; merge base 2aab5f797), pushed to origin 02:50Z. CI on e906b4f60: run 38105865348, conclusion success, verified by gh run view JSON.
- Case applied: (2). The absorb resolved a code conflict (5d7c section 2). Conflicted path: internal/hook/protected_zone_shell.go (the only conflicted path). Resolution: this card's K1, K2, and K3 changes re-applied on main's version.
- Delta audit of e906b4f60 (sync-auditor, .moai/reports/t1584/delta-audit-e906b4f60.md, 02:53Z): verdict FAIL. F1 (High, blocking): a deletion after a multi-word cd was judged from the root. F2 (High, debt): an option word before the operand.
- F1 disposition: fixed in this commit (audit fix 2a). A cd statement with two or more literal operand words, with or without the separator, keeps the reset and also tracks its first operand as a possible directory, unless an option word precedes that operand. Scope note: REQ-ZSCD-001 had been written for the separator form only. It now covers the no-separator shape too, because the audit's F1 includes `cd zone_dir extra && rm a.log`. That widening is a lane judgment, recorded in the card wait record, and open to leader review.
- F2 disposition: accepted as a known limitation of this card, per leader ruling d-20261010T005950 item 1. Option-bearing forms, such as cd -P dir and cd -P -- dir, keep the reset and are not added to this card. An earlier draft folded this form into the union; it was reverted before this commit. No test pins the option form as allow, because the SPEC warns against freezing a bypass behind a green gate. E5 evidence (bash 3.2.57 probes in the delta audit): cd -P /usr moves to /usr; cd /usr /tmp moves to /usr.
- Matrix: TestProtectedZoneShellParsingMatrix, 68 cells (66 landed, plus cd_track_dashdash_second_separator_rm and cd_track_multi_word_rm, both DENY). RED on e906b4f60: the two DENY shapes failed with decision="allow" (lane scratch file red-matrix.txt). GREEN on this tree: 68 cells, 0 FAIL. The non-zone multi-word allow control moved out of the matrix into TestProtectedZoneShellCdAllowDispositions (REQ-ZSCD-005).
- Full internal/hook package (go test -count=1 -timeout 40m, env scrubbed): ok, 310.978 s, 0 FAIL. Lane-local run on this tree (output scratch file full-package-final.txt).
- SPEC: the false premise (bash rejects a two-operand cd and the shell stays put) is corrected in REQ-ZSCD-001, REQ-ZSCD-005, plan K1, plan M2, and acceptance section D.2. Bash 3.2.57 moves into the first operand. No requirement or acceptance ID changed. Frontmatter: spec.md 0.2.2 (2026-10-11), plan.md 0.2.1 (2026-10-11), acceptance.md 0.2.1 (2026-10-11); status completed unchanged. The plan-artifact hash changes with these edits (evidence backfill, see section E.4).
- Known documentation gap, not in this change: the wording "to be captured" for EV-ZSCD-002 remains at spec.md line 64 and acceptance.md line 467. It predates this delta, and acceptance.md line 222 records the capture. Recorded for the completion report.
- Pending at this commit, reported in the completion report: the CI run on the pushed head, the codex audit (audit_multi) on the final SHA, and the sync-auditor delta scoped to F1 and F2.
