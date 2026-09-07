# SPEC-UPDATE-MIRROR-HEAL-001 — Progress

Card t520 · branch `WT-update-mirror-heal` · base `origin/develop` = `0b1e27877`
Evidence path: `.moai/reports/t520/`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md` (0.5.0), `plan.md`, `acceptance.md` (0.4.0), this file. Tier M, Class B.
- Plan-audit iteration 3: **PASS 0.91** (0.81 → 0.84 → 0.91, no regression). Follow-ups closed:
  **R3-1** AC-UMH-013's `find_rc` measured `sort`'s status through a pipeline and its `echo` was
  never redirected — `find`/`sort` split, `find_rc` + `link_rc` captured and written into the
  artifact, verified in both directions on this machine (missing path → `1`, existing path → `0`;
  the old piped form → `0` with an empty file), and §D.0 gained rule 5 so the family
  (D1 → N3 → R3-1) is named rather than re-dug. `${PIPESTATUS[@]}` rejected as bash-only.
  **R3-3** REQ-UMH-010's coordinates demoted to a dated reference with a re-measure command.
  **R3-2** ordering left as-is by deliberate judgment (a final-stage renumber is what produced N2).
- **Post-run, pre-sync repair (`acceptance.md` 0.5.0).** The run surfaced a SPEC-internal
  contradiction: AC-UMH-014's AC-UMH-013 mutant row ordered a write to **this** repository's
  `.agents/`, which REQ-UMH-009 / C-3 forbid. The mutant corpus is now an isolated temporary
  directory; the superseded wording is retained with the soundness argument (an absence guard's
  mutant needs a corpus of the same *kind*, not this specific repository) and with the residue kept
  open — it is **not** established that the recipe behaves identically on this repository's own
  `.agents/` tree, which does not exist here. Deviation **D-1** already records the run-phase
  disposition in §E.2 (verified present; not edited — §E.2 is run-phase owned).
  §D.0 also gained two rules from run findings: **rule 6** a compile-failure red is not a catch by
  the guard (with the AC-UMH-003 first-form mutant and the compile-level RED-1 of
  AC-UMH-001..009 / 015..017 named as limits of that execution), and **rule 7** a snapshot witness
  must carry `ModTime().UnixNano()` or it cannot see a same-bytes rewrite (AC-UMH-006).
- Run-phase notes recorded in `plan.md` M3/M4: dangling is a Path A-only hazard; a deployer-option
  seam shape would also change deploy-time behavior and must be chosen deliberately.
- Plan-audit iteration 2: FAIL 0.84 (no regression; the three iteration-1 blockers were confirmed
  closed, and the new findings were introduced by that repair). All five closed: **N1** the Path A
  bound was described as an existing producer property when the producer has no such check —
  promoted to REQ-UMH-010 and the conflicting `plan.md` M3 sentence rewritten (two-step derive +
  filter); **N2** mutant coverage recounted (9 guards, 9 rows) and the false AC-008/AC-015 sharing
  claim retracted with its inversion argued; **N3** AC-UMH-013's `-printf` recipe replaced with
  POSIX primaries, `2>/dev/null` removed, exit statuses asserted, control run against a known
  non-empty directory; **N4** K4 narrowed to the pre-feature population with the superseded wording
  retained; **N5** the fixture-reachability gap recorded below.
- Plan-audit iteration 1: FAIL 0.81, 3 blocking findings; all three repaired in `spec.md` 0.3.0 /
  `acceptance.md` 0.2.0 (D1 vacuous git guard + mutant-list coverage audit §D.2, D2 §3.6 branch
  analysis with the boundary pinned by AC-UMH-015/016, D3 AC-UMH-017 version matrix, D4 §4 premise
  correction). Both externally-supplied D2 arguments are recorded in §3.6; the decision rests on
  two measured grounds (stuck-forever control flow, Path A self-limitation), not on either.
- SPEC ID regex self-check executed: `[[ "SPEC-UPDATE-MIRROR-HEAL-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- ID uniqueness: `ls .moai/specs | grep -i -e UPDATE-MIRROR -e MIRROR-HEAL` → no output.
- Open decisions resolved in-SPEC: existence gate (`spec.md` §3.5, R5 deploy-version stamp; R1/R2/R3/R4 rejected with measurements) and repair scope (`spec.md` §4, S2).
- Accepted residual risk (explicit, and pinned by AC): a stamped-but-partial project receives the 16
  Path B files it never held (`spec.md` §3.6). Path A stays at zero entries there by construction.
- **Carried gap (was report-only until iteration 2 — N5).** AC-UMH-015/016 need a *genuinely
  partial* deploy fixture (stamp written inside the walk, walk then failed). Whether that state is
  reachable through existing test seams or needs a new injection point is unsettled, and it is the
  one place the §3.6 boundary pin could prove expensive. Recorded here because a boundary that lives
  only in a report is not an approved boundary.
- **Budget signal (not a defect).** 17 acceptance criteria against the Tier M guide of 16. Coverage
  was not cut to fit. The one consolidation candidate identified is AC-UMH-003 + AC-UMH-004 (both
  gate-closed no-ops over the same fixture family, differing only in the stamp value); it is left
  unmerged in this round because renumbering the mutant table during a final repair risks more than
  the row it saves.
- Gaps carried into run phase: the constant's home package (`plan.md` M1), the seam's shape (`plan.md` M3), and t498's unrecorded scratch-project preconditions (`plan.md` B3).

## §E.2 Run-phase Evidence

### Decisions taken in run phase (the two `plan.md` gaps, resolved)

- **M1 — the constant's home: `internal/template`.** `MirrorIntroducedVersion = "3.1.3"` is declared
  in `internal/template/skill_mirror_repair.go`, next to the producer whose behavior it describes
  rather than next to the CLI gate that reads it. Consequence accepted and recorded: AC-UMH-010's
  citation check reads `../template/skill_mirror_repair.go` from the `internal/cli` test.
- **M3 — the seam's shape: a package-level function, NOT a `DeployerOption`.**
  `template.RepairSkillMirror(fsys, projectRoot) *MirrorRepairResult`. This is the deliberate answer
  to `plan.md` M3 run-phase note 2: a deployer option would ALSO be live on the deploy-time path,
  where `Deploy` already runs the producer, so that shape would change deploy behavior as a side
  effect of a repair feature. **Deploy-side effect of the chosen shape: none.** `deployer.go` is
  unmodified, and `Deploy`'s call to `mirrorSkills` is untouched — the blast radius is the repair
  pass alone. Per-entry semantics still reuse `mirrorOneSkill`, so REQ-UMH-005 (non-symlink
  occupancy) and the idempotent already-correct branch come from the producer and cannot drift
  from it.
- **REQ-UMH-010 is Path A only** (`plan.md` M3 run-phase note 1, honoured). The target-existence
  filter sits in `repairPathA` above the shared producer. Path B carries no such filter: those are
  ordinary template writes that never touch the mirror producer, so the assumed shared failure mode
  does not exist there.

### Files changed

| File | Change |
|---|---|
| `internal/template/skill_mirror_repair.go` | NEW — `MirrorIntroducedVersion`, `MirrorRepairResult`, `PublishedSkillNames`, `RepairSkillMirror` |
| `internal/cli/update_mirror_heal.go` | NEW — `mirrorRepairGateOpen`, `repairSkillMirrorBestEffortAt`, `repairSkillMirrorBestEffort` |
| `internal/cli/update_mirror_heal_test.go` | NEW — AC-UMH-001..012, 015..017 |
| `internal/cli/update_mirror_heal_wiring_test.go` | NEW — REQ-UMH-001 call-site reachability guard |
| `internal/cli/update.go` | call site added beside `refreshCodexWiringBestEffort`, before the `syncSkipped` return |
| `internal/cli/doctor_codex.go` | M5 — mirror-absent detail text reconciled (REQ-UMH-008) |

No file under `internal/template/templates/**` changed, so `make build` is not owed (M6 — checked,
not assumed: `git status --porcelain -- internal/template/templates` → empty, with the control
`git status --porcelain -- internal/template` → `?? internal/template/skill_mirror_repair.go`, so
the empty result is a measurement rather than a selector that matches nothing).

### RED evidence (test-first, verbatim)

**RED-1 — before any implementation existed.** `go vet ./internal/cli/` with only the test file
present:

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_mirror_heal_test.go:169:2: undefined: repairSkillMirrorBestEffortAt
internal/cli/update_mirror_heal_test.go:182:32: undefined: template.PublishedSkillNames
internal/cli/update_mirror_heal_test.go:519:14: undefined: template.MirrorIntroducedVersion
internal/cli/update_mirror_heal_test.go:520:95: undefined: template.MirrorIntroducedVersion
internal/cli/update_mirror_heal_test.go:728:14: undefined: mirrorRepairGateOpen
internal/cli/update_mirror_heal_test.go:742:6: undefined: mirrorRepairGateOpen
```

**RED-2 — assertion-level, with the seam present but M5 not yet done:**

```
--- FAIL: TestUpdateMirrorHeal_GateConstantGrounded (0.00s)
    the constant's doc comment does not cite "CHANGELOG.md" — the value would be a guess on the record
--- FAIL: TestUpdateMirrorHeal_DoctorGuidance (0.00s)
    the doctor still asserts a routine update cannot restore the mirror — false after this card (REQ-UMH-008)
```

**Stated plainly (a limit of this run, not a claim about it).** RED-1 is a COMPILE-level red for
AC-UMH-001..009 and 015..017: the symbols did not exist, so those tests could not have passed, but
no assertion-level red was captured for them at that moment. The gap is closed after the fact by
the `EXTRA-noop` mutant below, which neuters `RepairSkillMirror` on the finished tree and shows
each positive criterion going red on its own assertion.

### Per-AC matrix

Command form for a Go criterion: `go test ./internal/cli/ -run <name> -count=1 -v`.
Full green run: `go test ./internal/cli/ -run TestUpdateMirrorHeal -count=1` → `ok
github.com/modu-ai/moai-adk/internal/cli 2.850s` (16/16 `--- PASS`).

| AC | Verifying command | Actual output | Status |
|---|---|---|---|
| AC-UMH-001 | `-run TestUpdateMirrorHeal_RestoresPathA` | `--- PASS: TestUpdateMirrorHeal_RestoresPathA (1.35s)` | PASS |
| AC-UMH-002 | `-run TestUpdateMirrorHeal_RestoresPathB` | `--- PASS: TestUpdateMirrorHeal_RestoresPathB (1.04s)` | PASS |
| AC-UMH-003 | `-run TestUpdateMirrorHeal_NoCreateBelowStamp` | `--- PASS: … (0.42s)` | PASS |
| AC-UMH-004 | `-run TestUpdateMirrorHeal_NoCreateWithoutStamp` | `--- PASS: … (0.36s)` | PASS |
| AC-UMH-005 | `-run TestUpdateMirrorHeal_EarlyReturnPreserved` | `--- PASS: … (0.33s)` | PASS |
| AC-UMH-006 | `-run TestUpdateMirrorHeal_HealthyIsNoop` | `--- PASS: … (0.33s)` | PASS |
| AC-UMH-007 | `-run TestUpdateMirrorHeal_SkipsForeignOccupant` | `--- PASS: … (0.34s)` | PASS |
| AC-UMH-008 | `-run TestUpdateMirrorHeal_ScopeExcludesLocalSkills` | `--- PASS: … (0.40s)` | PASS |
| AC-UMH-009 | `-run TestUpdateMirrorHeal_FailOpen` | `--- PASS: … (0.33s)` | PASS |
| AC-UMH-010 | `-run TestUpdateMirrorHeal_GateConstantGrounded` + `grep -n "3.1.3" internal/template/skill_mirror_repair.go` | `--- PASS`; grep → `39:… "[3.1.3] - 2026-08-24" …` and `48:const MirrorIntroducedVersion = "3.1.3"` | PASS |
| AC-UMH-011 | `-run TestUpdateMirrorHeal_DoctorGuidance`; `grep -c 'does not restore it' internal/cli/doctor_codex.go`; control `grep -c 'mirror absent' …` | `--- PASS`; `0`; control `1` (non-zero → the `0` is a measurement) | PASS |
| AC-UMH-012 | `-run TestUpdateMirrorHeal_DoctorRemainsReadOnly` | `--- PASS` (doctor span 0 write calls; control: repair-pass source non-zero) | PASS |
| AC-UMH-013 | snapshot recipe, `before` (pre-first-edit) vs `after` | both `agents_exists_rc=1`; `diff … → diff_exit=0`, no output; control `find .claude/skills -print` → `control_find_rc=0`, `control_line_count=427` | PASS |
| AC-UMH-014 | `.moai/reports/t520/mutants.py` + `ac013-mutant-probe.sh` | 9/9 mutants caught — table below | PASS |
| AC-UMH-015 | `-run TestUpdateMirrorHeal_PartialDeployPathA` | `--- PASS: … (0.01s)` | PASS |
| AC-UMH-016 | `-run TestUpdateMirrorHeal_PartialDeployPathB` | `--- PASS: … (0.02s)` | PASS |
| AC-UMH-017 | `-run TestUpdateMirrorHeal_VersionMatrix` | `--- PASS` over 6 sub-tests (`3.1.3`, `v3.1.3`, `3.2.0-rc.0`, `3.1.2`, `dev`, `absent`) | PASS |

Additional guard beyond the matrix (reachability, REQ-UMH-001):
`-run TestUpdateMirrorHeal_WiredBeforeSyncSkippedReturn` → `--- PASS`. Without it, every criterion
above drives the seam directly and none would notice the call site being dropped from `runUpdate`.

### Mutant results (AC-UMH-014) — caught / not caught

Applied with MUST-REPLACE semantics (`.moai/reports/t520/mutants.py` aborts if a needle is absent
or ambiguous), so a mutant that silently failed to apply cannot masquerade as "not caught".
Verbatim run output: `.moai/reports/t520/mutants-output.txt`.

| Guard | Mutant | Caught | Evidence |
|---|---|---|---|
| AC-UMH-003 | ignore the version gate | YES | `.agents exists after a below-stamp run (err=<nil>)` |
| AC-UMH-004 | treat `"0.0.0"` as satisfying the gate | YES | `.agents exists after an unstamped run (err=<nil>)` |
| AC-UMH-006 | rewrite published files unconditionally | YES | `repair changed a healthy mirror; diff: […mtime…]` |
| AC-UMH-007 | remove and replace a non-symlink occupant | YES | `user file gone after repair: … MINE.md: no such file or directory` |
| AC-UMH-008 | widen the scope set to every dir under `.claude/skills` (S1) | YES | `the repair mirrored a locally-authored skill no deploy produced` |
| AC-UMH-011 | re-introduce the "does not restore it" phrasing | YES | `the doctor still asserts a routine update cannot restore the mirror` |
| AC-UMH-012 | add a write call inside the doctor mirror span | YES | `the doctor mirror span contains 1 write call(s)` |
| AC-UMH-013 | create `.agents/skills/mutant-probe`, then diff the snapshots | YES (substituted corpus — see deviation D-1) | `5a6 > .agents/skills/mutant-probe` |
| AC-UMH-015 | remove the REQ-UMH-010 target-existence filter | YES | `the repair created 34 Path A entries on a project with no .claude/skills` / `34 dangling entries` |
| EXTRA — neuter the whole pass (not in the AC-014 list) | `RepairSkillMirror` returns immediately | YES | `repair restored no entries at all`; `published artifact not restored: …` — the assertion-level RED for the positive criteria |

**Mutants NOT caught: none.** Recorded as a fact rather than an absence: all nine listed mutants
plus the extra probe flipped their guard red.

Two boundary facts about HOW two of them were caught, recorded rather than smoothed:

- **AC-UMH-003's first mutant form (`if false {`) caught the guard by COMPILE FAILURE**, not by the
  assertion — bypassing the gate left `stamp` unused. A compile failure proves nothing about the
  guard, so the mutant was re-run in a compiling form (`if !mirrorRepairGateOpen(stamp) && false {`)
  and the guard then flipped red on its own assertion. The first form is recorded because it is
  exactly the shape that would have been mis-scored as a real catch.
- **AC-UMH-006 is only a real guard because the snapshot witness records `mtime`.** Path+size alone
  cannot see a file rewritten with identical bytes, which is precisely what the mutant does. The
  witness was strengthened during run phase for this reason (portable `ModTime().UnixNano()`, no
  `syscall.Stat_t`, so the guard still compiles on windows).

### Deviations

- **D-1 — AC-UMH-013's mutant was run against `/tmp`, not this repository.** `acceptance.md`
  AC-UMH-014 specifies "create `.agents/skills/mutant-probe` in **this** repository, then run the
  criterion's snapshot diff". That instruction contradicts REQ-UMH-009 / C-3, which forbids this
  card from writing this repository's `.agents/` state. C-3 was treated as controlling, so the same
  recipe was exercised against a scratch corpus of matching shape
  (`.moai/reports/t520/ac013-mutant-probe.sh`), including a symlink entry so the `readlink` half of
  the recipe is exercised, plus an unmutated re-snapshot control proving the recipe is stable when
  nothing changes. What the mutant tests — the recipe's ability to report a difference — is
  corpus-independent; what is NOT established is that the recipe behaves identically on this
  repository's own `.agents/` tree, which does not exist here (`agents_exists_rc=1`). **This is a
  SPEC-internal contradiction, surfaced rather than resolved unilaterally.**

### Inherited RED (not caused by this card)

`go test ./internal/template/ -run TestManifestHashFormat` fails with `CATALOG_HASH_UNSTABLE` on
four entries (`manager-develop`, `manager-lead`, `manager-design`, `e2e-tester`). Attribution:
`git status --porcelain -- .claude/agents internal/template/catalog.yaml` → empty (all inputs to
that test are byte-identical to HEAD), with the control
`git status --porcelain -- internal/cli/doctor_codex.go` → ` M internal/cli/doctor_codex.go`, so
the empty result is a measurement. The failure therefore pre-dates this branch's work. No repair
attempted — it is outside this SPEC's scope envelope (L46).

### Contention observation (a fact about the measurement, not about the code)

A post-commit re-run of `go test ./internal/cli/ -count=1 -timeout 900s` reported
`FAIL github.com/modu-ai/moai-adk/internal/cli 901.894s` — the wall-clock timeout, not a test
assertion. Attributed rather than assumed: `uptime` at that moment read
`load averages: 21.82 13.60 12.40` with `ps aux | grep -c '[g]o test'` → `6` concurrent test
processes from other lanes. Re-measured at a larger budget:
`go test ./internal/cli/ -count=1 -timeout 1800s` → `ok github.com/modu-ai/moai-adk/internal/cli
700.131s`, `exit=0`. The same package had already completed inside the 900s budget in an earlier,
quieter window. Recorded because a timeout is indistinguishable from a failure in the exit code,
and reading it as either without attribution would be an unobserved claim in one direction or the
other. It is also the exact hazard `CLAUDE.local.md` §4/§6 names: this measurement measured the
machine.

### Verification-tooling friction (candidate `/moai:feedback`)

The `Write` tool's ast-grep hook rejects the project's own REQUIRED errcheck idiom. Rule
`go-error-ignored-blank` (`.moai/astgrep-rules/go/error-handling.yml:9`, pattern
`$_, $ERR = $FUNC($$$ARGS)`) has no comment-awareness, so `_, _ = fmt.Fprintf(...)` is refused as an
ERROR — while `.golangci.yml:35` sets `errcheck.check-blank: false`, which makes that blank-assign
form the required way to write a deliberately-unchecked console write. The idiom occurs 885 times
in `internal/cli` already. `internal/cli/update_mirror_heal.go` was therefore written via a Bash
heredoc, with the rationale recorded inline at both sites.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: dbb9a53a4        # M1 implementation commit (this SPEC's only implementation commit)
run_status: complete
ac_pass_count: 17
ac_fail_count: 0
ac_pass_with_debt_count: 0
mutants_listed: 9
mutants_caught: 9
mutants_not_caught: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0   # golangci-lint run ./internal/cli/... ./internal/template/... -> "0 issues."
go_vet: clean                          # go vet ./internal/cli/... ./internal/template/... -> exit 0, no output
cross_platform_build:
  darwin_arm64: pass                   # go test ./internal/cli/... -> all ok
  windows_amd64_build: pass            # GOOS=windows GOARCH=amd64 go build ./... -> exit 0
  windows_amd64_vet: pass              # GOOS=windows GOARCH=amd64 go vet ./internal/cli/... -> exit 0 (test files compile)
scoped_test_run:
  command: go test ./internal/cli/... ./internal/template/... -count=1 -timeout 900s
  result: all packages ok EXCEPT internal/template TestManifestHashFormat (inherited RED, attributed above)
full_local_suite_run: false            # CLAUDE.local.md §4/§6 — the full-suite verdict is CI's
total_run_phase_files: 6
template_source_changed: false         # make build not owed (M6, checked)
m1_to_mN_commit_strategy: two commits on WT-update-mirror-heal, both naming card t520 (implementation+status, then evidence)
branch_pushed: false                   # lane does not push its branch and does not request CI
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: 7a16df6605ac7958891473090358939b31d89c9d   # the sync commit; written as pending-backfill-sync there (a commit cannot cite its own hash) and backfilled here
sync_status: complete
b12_self_test_a: pass                    # grep -c 'SPEC-UPDATE-MIRROR-HEAL-001' CHANGELOG.md -> 0 before emission
b12_self_test_b: pass                    # distinct AC ids in acceptance.md -> 17; the entry states 17 (AC-UMH-001..017)
b12_self_test_c: pass                    # every path cited in the entry verified with ls; 12/12 present, plus internal/cli/init.go
changelog_entry_position: "[Unreleased] / ### Added, first bullet (CHANGELOG.md:12)"
frontmatter_status_transitions:
  spec.md: in-progress -> completed      # `status` + `updated` only
  plan.md: none                          # no frontmatter block; stateless on this axis
  acceptance.md: none                    # no frontmatter block
  progress.md: none                      # no frontmatter block
docs_surface_decision: changelog-only
docs_rationale: >-
  No user-facing documentation surface carries a claim this change falsifies. The docs-site
  `moai update` and Codex Dual Harness pages describe the mirror as a deploy-time artifact and
  make no statement about a version-matched update failing to restore it (a repo-wide grep over
  `docs-site/content` for the `--templates-only --force` redeploy advice, and for the doctor's
  `mirror absent` / `Codex Wiring` strings, both return no output), so the 4-locale obligation
  was not triggered and no locale file was touched. The one user-visible string that DID assert
  the old behavior lives in the binary, not the docs, and it was corrected in run phase
  (`internal/cli/doctor_codex.go`, REQ-UMH-008 / AC-UMH-011).
sync_phase_files_changed: 3              # CHANGELOG.md, spec.md frontmatter, this file
uncommitted_pre_sync_work_folded_in: true
uncommitted_pre_sync_work_note: >-
  Three lane-authored, uncommitted files were present when sync began and are carried in this
  same commit rather than left behind (a worktree merge takes commits only): `acceptance.md`
  0.5.0 (the D-1 contradiction repair, already narrated in §E.1), the matching §E.1 bullet in
  this file, and a lint-only loop-variable rename in `.moai/reports/t520/mutants.py`.
  manager-docs authored none of them and edited no `acceptance.md` body content.
carried_gaps:
  - partial-deploy fixture reachability through existing test seams is unsettled (§E.1, N5)
  - AC-UMH-013's recipe is NOT established against this repository's own `.agents/` tree, which
    does not exist here; the mutant ran against an isolated /tmp corpus (deviation D-1)
  - RED-1 for AC-UMH-001..009 / 015..017 was compile-level, not assertion-level; assertion-level
    red came after the fact from an out-of-list mutant (a limit of that execution)
  - inherited RED `TestManifestHashFormat` (CATALOG_HASH_UNSTABLE, 4 entries) is not this card's,
    was not repaired, and is not claimed to pass; attribution re-checked in sync phase with a
    valid control (the porcelain status of `.claude/agents` + `internal/template/catalog.yaml` is
    empty, while the same selector under `--ignored` reports the known ignored residue)
  - full local suite not run (lane load discipline); the full-suite verdict is CI's
observations_not_actioned:
  - >-
    Stray runtime state exists inside the SPEC directory at
    `.moai/specs/SPEC-UPDATE-MIRROR-HEAL-001/.moai/state/` — untracked and ignored
    (`.gitignore:228`), authored by neither this card's author nor its implementer. NOT deleted
    (an untracked delete has no safety net) and NOT this card's to fix; worth a separate
    feedback item.
sync_phase_verification:
  - "diff --stat against the merge-base 0b1e27877 -> 18 files, 3476 insertions, 1 deletion (merge-base, NOT origin/develop, which has moved since this branch was cut)"
  - "go build ./internal/... -> exit 0"
  - "grep -c 'SPEC-UPDATE-MIRROR-HEAL-001' CHANGELOG.md -> 1 after emission"
```
