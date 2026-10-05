# SPEC-UPDATE-MIRROR-HEAL-001 — Acceptance Criteria

Version 0.5.0 (post-run pre-sync repair): AC-UMH-014's AC-UMH-013 mutant row ordered a write to
**this** repository's `.agents/`, contradicting REQ-UMH-009 / C-3 in the same document — the mutant
corpus is now an isolated temporary directory, with the superseded wording, the soundness argument,
and the un-established residue kept beside it. §D.0 gains rules 6 (a compile-failure red is not a
guard catch) and 7 (a snapshot witness needs `mtime`), both from run-phase findings.
Prior: 0.4.0 (plan-audit iteration 3 repair, R3-1: AC-UMH-013's `find_rc` measured `sort`'s status,
not `find`'s, and its `echo` was never redirected into the artifact — `find`/`sort` split, both
statuses captured and written, and §D.0 gains rule 5 so the trap is not re-dug).
Prior: 0.3.0 (plan-audit iteration 2 repair: N1 AC-UMH-015 re-pointed at the new REQ-UMH-010,
N2 mutant-list recount + false sharing-claim retracted, N3 AC-UMH-013 recipe made portable and
its exit codes asserted).
Prior: 0.2.0 (iteration 1: D1 vacuous-guard removal + mutant-list coverage audit, D2 boundary
pinning, D3 version-fixture matrix).

Card t520 · every criterion is a command plus an expected output. Go criteria run under
`t.TempDir()` only; none of them reads, writes, or deletes this repository's `.agents/`
(REQ-UMH-009 / C-3).

Naming: `TestUpdateMirrorHeal_*` in `internal/cli`. Where a criterion names a test, the verifying
command is `go test ./internal/cli/ -run <name> -v` and the expected output is a `--- PASS: <name>`
line; the criterion body states what the test must assert, so a vacuously-green test does not
satisfy it.

## §D.0 Evidence-shape rules for this matrix

1. **No git-based evidence about `.agents/`.** Measured on this tree:
   `git check-ignore -v .agents/skills/foo` → `.gitignore:133:.agents/`, rc `0`; control
   `git check-ignore -v internal/cli/update.go` → rc `1`. `.agents/` is gitignored, so any
   `git diff -- .agents` returns empty **even when files were created** — git cannot witness this
   axis. Filesystem snapshots are used instead. (The wider `.gitignore` divergence is card t522's;
   this SPEC only stops relying on git here.)
2. **Every absence-as-PASS criterion carries a control.** A command whose empty output means PASS
   is paired with a second selector, in the same corpus, that must be non-empty — otherwise the
   empty result cannot be distinguished from a selector that never matched anything.
3. **Every absence guard appears in AC-UMH-014's mutant list.** §D.2 records the coverage audit that
   keeps that list from silently omitting a guard. The audit **counts the matrix rows**; it does not
   restate a remembered count (0.2.0 asserted 8 where the matrix marked 9).
4. **No `-printf`, no `2>/dev/null`, and the command's own exit status is asserted.** Measured on
   this machine (darwin): `/usr/bin/find . -maxdepth 0 -printf 'x\n'` → `find: -printf: unknown
   primary or operator`, rc `1`; control `/usr/bin/find . -maxdepth 0 -print` → `.`, rc `0`. So
   `-printf` is a GNU-only primary absent here, and a suppressed stderr would turn that failure into
   two empty snapshots and a PASS — the same vacuity class as §D.0 rule 1.
5. **`$?` after a pipeline is the LAST command's status, not the interesting one.** A recipe that
   asserts a command's exit status must run that command **unpiped** and capture `$?` immediately.
   Measured in a scratch dir (darwin): `find <missing> -print | LC_ALL=C sort > out; echo $?` → `0`
   with `out` empty, while the split form `find <missing> -print > raw; echo $?` → `1`. An assertion
   must also be **redirected into the artifact**; a value echoed to the terminal is not evidence a
   later reader can check. This is the fourth instance of the same family in this card
   (§D.0 rule 1 → rule 4 → this rule): a guard whose command silently did not measure what it names.
6. **A red produced by a compile failure is NOT a catch by the guard.** When a mutant is applied by
   deleting or bypassing a gate, the change can leave a variable unused (or a symbol undefined), and
   the package then fails to build. The test reports red without any assertion having executed, so
   the probe establishes nothing about the guard's discriminating power. A mutant must keep the
   package compiling — e.g. `if !mirrorRepairGateOpen(stamp) && false {` rather than removing the
   call — so the red is produced by the assertion. Where the first form of a mutant produced a
   compile-level red, **record that form** alongside the corrected one: it is the shape most easily
   mis-scored as a genuine catch. (Instances from this card's run phase are recorded by the
   implementer in `progress.md` §E.2 — the AC-UMH-003 mutant, and the compile-level RED-1 of
   AC-UMH-001..009 / 015..017 for which an extra out-of-list mutant supplied assertion-level red
   after the fact. Both are limits of that execution, not closed questions.)
7. **A snapshot witness must include `mtime`, or it cannot see a same-bytes rewrite.** Path + size
   alone are blind to a file rewritten with identical content, which is exactly what AC-UMH-006's
   mutant does. Go-side snapshots use `ModTime().UnixNano()` for portability.

## §D Acceptance matrix

| AC | Covers | Layer | Absence guard? |
|---|---|---|---|
| AC-UMH-001 | repair case (Path A) | REQ-UMH-001 | no |
| AC-UMH-002 | repair case (Path B) | REQ-UMH-004 | no |
| AC-UMH-003 | no-op: stamp below constant | REQ-UMH-002 | **yes** |
| AC-UMH-004 | no-op: stamp absent | REQ-UMH-002 | **yes** |
| AC-UMH-005 | early return preserved | REQ-UMH-003 | no |
| AC-UMH-006 | healthy project is a no-op | REQ-UMH-005 | **yes** |
| AC-UMH-007 | non-symlink occupant untouched | REQ-UMH-005 | **yes** |
| AC-UMH-008 | scope: template-shipped skills only | REQ-UMH-007 | **yes** |
| AC-UMH-009 | fail-open | REQ-UMH-006 | no |
| AC-UMH-010 | gate constant is grounded | §3.5 K4 | no |
| AC-UMH-011 | doctor guidance reconciled | REQ-UMH-008 | **yes** |
| AC-UMH-012 | doctor stays read-only | C-4 | **yes** |
| AC-UMH-013 | repository `.agents/` untouched | REQ-UMH-009 | **yes** |
| AC-UMH-014 | mutation probe over every absence guard | verification integrity | no |
| AC-UMH-015 | boundary: stamped-but-partial, Path A | §3.6 / REQ-UMH-010 | **yes** |
| AC-UMH-016 | boundary: stamped-but-partial, Path B | §3.6 | no |
| AC-UMH-017 | version-string degradation matrix | §3.5 / D3 | no |

---

**AC-UMH-001 — Path A restoration (precondition built by deploying, not by hand-writing a stamp).**
*Given* a project under `t.TempDir()` created by actually running the deploy path — so its
`.moai/config/sections/system.yaml` stamp and its `.agents/skills` entries are both products of the
same run — whose `.agents` directory is then removed;
*When* the update path runs at version match (the branch returning `skipped == true`);
*Then* `.agents/skills` exists again and holds a mirror entry for every template-shipped skill
present under `.claude/skills`, each resolving to its canonical directory.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_RestoresPathA -v` → `--- PASS`.
The test asserts the restored entry-name set **equals** the set captured before deletion (set
equality, not "non-empty") and that `os.Stat` succeeds on each entry (no dangling links). The
fixture MUST NOT write the stamp by hand — hand-writing it is what let 0.2.0 bypass the §3.6
question entirely.

**AC-UMH-002 — Path B restoration.**
*Given* the AC-UMH-001 fixture state;
*When* the version-matched update path runs;
*Then* all 16 published artifacts `.agents/skills/moai-<command>/SKILL.md` exist again with bytes
equal to the corresponding embedded template files.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_RestoresPathB -v` → `--- PASS`.
The expected 16 names are derived from the production set, never hand-typed; bytes are compared, so
a created-but-empty file fails.

**AC-UMH-003 — no-op when the stamp is below the constant.**
*Given* a project under `t.TempDir()` whose recorded `template_version` is below the
mirror-introducing constant, with no `.agents/` directory;
*When* the update path runs;
*Then* `.agents` does not exist afterwards and nothing under the project changed.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_NoCreateBelowStamp -v` → `--- PASS`.
The test asserts `os.Stat(root/".agents")` returns `fs.ErrNotExist` **and** that a recursive
pre/post filesystem snapshot (relative path + size + mode + symlink target) is identical. Control:
the same snapshot helper, run over a project the pass *does* repair, must report a difference —
proving the helper can see a change at all.

**AC-UMH-004 — no-op when no stamp exists.**
*Given* a directory under `t.TempDir()` with no `.moai/config/sections/system.yaml` (so
`GetProjectConfigVersion` yields `"0.0.0"`);
*When* the update path runs;
*Then* nothing under `.agents/` is created and the pre/post snapshot is identical.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_NoCreateWithoutStamp -v` → `--- PASS`,
with the same control as AC-UMH-003.

**AC-UMH-005 — the early return is preserved.**
*Given* the implementation as landed;
*When* `internal/cli/update_template_sync.go` is read and a version-matched run is executed;
*Then* the version-match branch still returns `(true, nil)` before `runTemplateSyncWithReporter`,
`runTemplateSyncWithProgress` still has signature `(skipped bool, err error)`, and the executed run
performs no template write while still repairing the mirror.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_EarlyReturnPreserved -v` → `--- PASS`.
The behavioral half (no template write + mirror repaired) is what distinguishes "repair beside the
optimization" from "optimization removed"; the source-read half alone would pass on a tree where
the guard exists but is unreachable.

**AC-UMH-006 — healthy project: no-op.**
*Given* a stamped project whose `.agents/skills` mirror is complete and correct;
*When* the version-matched update path runs;
*Then* the pre/post snapshot of `.agents/` is identical, symlink targets included.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_HealthyIsNoop -v` → `--- PASS`.
The snapshot witness MUST include `ModTime().UnixNano()` per §D.0 rule 7: with path + size only,
this guard is blind to a file rewritten with identical bytes — which is precisely what its mutant
does, so without `mtime` the guard is vacuous rather than merely weak.

**AC-UMH-007 — a non-symlink occupant is left untouched.**
*Given* a stamped project where `.agents/skills/<skill>` is a real directory (or file) holding user
bytes;
*When* the version-matched update path runs;
*Then* those bytes are unchanged, the path is still not a symlink, and the run reports the skip
rather than failing.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_SkipsForeignOccupant -v` → `--- PASS`.

**AC-UMH-008 — repair scope is template-shipped skills only.**
*Given* a stamped project carrying an extra locally-authored `.claude/skills/local-only-skill/`
that no template ships, with `.agents` removed;
*When* the version-matched update path runs;
*Then* `.agents/skills/local-only-skill` does **not** exist, while the template-shipped entries do.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_ScopeExcludesLocalSkills -v` → `--- PASS`.
The positive half (template-shipped entries present) is the control: it prevents a pass that
mirrors nothing at all from satisfying the negative half.

**AC-UMH-009 — fail-open.**
*Given* a stamped project in which mirror creation is made to fail (injected symlink/copy seam);
*When* the version-matched update path runs;
*Then* the update exits 0, a warning naming the failure reaches stderr, and no error propagates.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_FailOpen -v` → `--- PASS`.

**AC-UMH-010 — the gate constant is grounded.**
*Given* the mirror-introducing version constant;
*When* its declaration is read;
*Then* its value is `3.1.3` and its doc comment cites both the release record
(`CHANGELOG.md` `[3.1.3] - 2026-08-24`) and the introducing commit (`9c94c6b7a`).
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_GateConstantGrounded -v` → `--- PASS`,
and by hand `grep -n "3.1.3" <constant file>` → the declaration line plus its citation comment
(non-empty expected output, so no control is needed).

**AC-UMH-011 — doctor guidance reconciled.**
*Given* the repair landed;
*When* the mirror-absent detail text is rendered;
*Then* it no longer asserts that a routine `moai update` on a version-matched project cannot
restore the mirror.
Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring|TestUpdateMirrorHeal_DoctorGuidance' -v`
→ all `--- PASS`; plus
`grep -c "does not restore it" internal/cli/doctor_codex.go` → `0`
with control `grep -c "mirror absent" internal/cli/doctor_codex.go` → non-zero (so the `0` is a
measurement, not a selector that matches nothing in that file).

**AC-UMH-012 — the doctor stays read-only.**
*Given* the changes as landed;
*When* `inspectSkillMirror` and its callees are read;
*Then* they contain no create, write, remove, or symlink call.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_DoctorRemainsReadOnly -v` → `--- PASS`.
The test asserts absence of `os.Mkdir|os.MkdirAll|os.WriteFile|os.Remove|os.Symlink|os.Create`
within the doctor mirror functions' source span, and asserts the **same selector matches non-zero
times inside the repair pass's own source span** — the control that proves the selector works.

**AC-UMH-013 — this repository's `.agents/` is untouched (portable filesystem snapshot, not git).**
*Given* the full run-phase work;
*When* the repository's `.agents/` state is captured before the first edit and again at the end;
*Then* its existence status is unchanged and, where it exists, its entry listing (paths plus
symlink targets) is identical — with **every capture command's exit status asserted**, so a command
that failed to run cannot present as an empty snapshot.

Recipe, run once as `<phase>` = `before` and once as `after`. POSIX primaries only (`-print`,
`-type`); no `-printf`, no `2>/dev/null`.

1. Existence, recorded rather than inferred:
   `test -e .agents; echo "agents_exists_rc=$?" >> .moai/reports/t520/agents-<phase>.txt`
   (`0` = present, `1` = absent). On this tree the expected value is `1`.
2. Only when step 1 printed `0`, the listing — `find` and `sort` are **separate commands**, so the
   status captured is `find`'s own:
   ```sh
   find .agents -print > .moai/reports/t520/agents-<phase>.raw
   find_rc=$?
   echo "find_rc=$find_rc" >> .moai/reports/t520/agents-<phase>.txt
   LC_ALL=C sort .moai/reports/t520/agents-<phase>.raw >> .moai/reports/t520/agents-<phase>.txt
   find .agents -type l -print > .moai/reports/t520/agents-<phase>.links
   link_rc=$?
   echo "link_rc=$link_rc" >> .moai/reports/t520/agents-<phase>.txt
   while read -r l; do printf '%s -> %s\n' "$l" "$(readlink "$l")"; done \
     < .moai/reports/t520/agents-<phase>.links | LC_ALL=C sort >> .moai/reports/t520/agents-<phase>.txt
   ```
   `find_rc` and `link_rc` MUST both be `0`; a non-zero value fails the criterion instead of being
   swallowed, and both are **written into the artifact** (`>>`), so the asserted value is the value a
   later reader can check.

   **Why not the one-liner.** `find … | LC_ALL=C sort > out` followed by `$?` reads the status of
   `sort`, not of `find`. Measured in a scratch dir on this machine: the piped form against a missing
   path printed `old_form_rc=0` while producing an empty file, whereas the split form above printed
   `find_rc_absent=1` (missing path) and `find_rc_present=0` (existing path, 2 lines) — both
   directions shown, so the assertion is known not to be always-true or always-false.
   `${PIPESTATUS[@]}` is deliberately **not** used: it is bash-only, and this recipe is meant to run
   under `/bin/sh` as well.
3. Compare: `diff .moai/reports/t520/agents-before.txt .moai/reports/t520/agents-after.txt` → no
   output, exit 0.

**Control (mandatory, same machine, same recipe).** Run step 2's `find` recipe against
`.claude/skills` — a directory known to be non-empty — and assert its line count is greater than
zero. This proves the recipe *produces output at all* on this platform, so an empty `.agents`
snapshot is a measurement rather than a silently-failed command. `test ! -e .agents` alone is not
accepted as that control: it measures the directory, not the command (iteration-2 finding N3).

Explicitly rejected evidence: `git diff --name-only <base>..HEAD -- .agents`, which returns empty
regardless because `.agents/` is gitignored (§D.0 rule 1).

**AC-UMH-014 — every absence guard is probed by mutation.**
*Given* AC-UMH-003, 004, 006, 007, 008, 011, 012, 013, 015 — the nine rows the §D matrix marks
`yes` — each assert that something does **not** happen, and an unimplemented or inert feature satisfies such a guard trivially;
*When* each is probed with a mutant that makes the forbidden thing happen;
*Then* each guard flips red, and every mutant is reverted afterwards.

Mutant per guard — the list is derived from the "Absence guard?" column of §D, not hand-picked:

| Guard | Mutant |
|---|---|
| AC-UMH-003 | make the pass ignore the version gate, creating `.agents` on a below-stamp project |
| AC-UMH-004 | treat the `"0.0.0"` degradation as satisfying the gate |
| AC-UMH-006 | make the pass rewrite published files unconditionally instead of restore-missing-only |
| AC-UMH-007 | make the pass remove and replace a non-symlink occupant |
| AC-UMH-008 | widen the scope set to every directory under `.claude/skills` (S1) |
| AC-UMH-011 | re-introduce the "does not restore it" phrasing in the doctor detail |
| AC-UMH-012 | add a write call inside the doctor mirror span |
| AC-UMH-013 | in an **isolated temporary directory** carrying a `.agents/skills/` tree (never this repository — see the note below), create `.agents/skills/mutant-probe`, then run the criterion's snapshot recipe before and after — the `diff` MUST report a difference and exit non-zero (this is the mutant the 0.2.0 git-based form could not catch, and the reason the guard was re-written) |
| AC-UMH-015 | remove the REQ-UMH-010 target-existence filter — on the empty-skills fixture the pass must then create dangling entries, flipping AC-UMH-015 red |

Verify: per mutant `go test ./internal/cli/ -run <guard> -v` → `--- FAIL` while mutated (for
AC-UMH-013, a non-empty `diff` instead), and after revert
`grep -rn 'MUTANT\|mutant' internal/cli` → no output, with the full guard set green again.
Control for that revert check: the same selector run while a mutant is in place must be non-empty.
Any mutant that is **not** caught is recorded in `progress.md` §E.2 as a fact about the guard's
boundary — never deleted, never quietly reworded.

> **AC-UMH-013 mutant corpus — superseded wording (run phase, deviation D-1).** The row previously
> read "create `.agents/skills/mutant-probe` in **this** repository". That instruction contradicted
> REQ-UMH-009 / C-3 — and the dispatch's [HARD] restatement of it — inside the same document: the
> guard forbids writing this repository's `.agents/`, and the mutant ordered exactly that write.
> C-3 governs, so the row now names an isolated temporary directory.
>
> **Why the substitution is sound rather than a dodge**: an absence guard's mutant needs a corpus of
> the *same kind* as the one the guard observes — a directory tree with a `.agents/skills/` layout —
> and nothing about the probe requires that corpus to be **this** repository. The recipe under test
> is filesystem-generic. The implementer reached the same conclusion independently, ran the mutant
> in `/tmp` (exercising the `readlink` half with a symlink entry, plus an unmutated re-snapshot
> control), and recorded it as deviation D-1 rather than silently substituting — that disposition is
> endorsed here, and the wording is what changed, not the execution.
>
> **Gap, carried and not closed**: it is **NOT established** that this recipe behaves identically
> against this repository's own `.agents/` tree, because that tree does not exist here. The
> substitution is sound for the mutant's purpose and is not evidence about the real corpus.

**AC-UMH-015 — boundary: stamped-but-partial project, Path A.**
*Given* a project constructed to hold the state §3.6 accepts as residual risk — `system.yaml`
stamped ≥ the constant, but no mirror and an empty or absent `.claude/skills` (a deploy that wrote
the stamp inside the walk and then failed);
*When* the version-matched update path runs;
*Then* **zero** Path A mirror entries are created, because REQ-UMH-010's target-existence filter
drops every candidate whose `.claude/skills/<name>` is absent.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_PartialDeployPathA -v` → `--- PASS`.
The test asserts the count is zero **and** that no entry under `.agents/skills` is a symlink whose
`os.Stat` fails — because the failure mode here is not "too many entries" but *dangling* ones.
This criterion verifies a requirement on new work, not an inherited property: the unmodified
producer never stats `srcDir` on the symlink path (`skill_mirror.go:206` vs `:247`) and
`os.Symlink` (`:168`) succeeds against a missing target, so without REQ-UMH-010 this fixture yields
16 dangling links reported as `MirrorModeSymlink`.

**AC-UMH-016 — boundary: stamped-but-partial project, Path B.**
*Given* the same fixture;
*When* the version-matched update path runs;
*Then* the 16 published `SKILL.md` files **are** created, and the run reports it.
Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_PartialDeployPathB -v` → `--- PASS`.
This is the accepted residual risk of §3.6 made explicit: the behavior is asserted, not assumed, so
"accepted" does not decay into "unverified". A future decision to exclude this population changes
this criterion deliberately rather than discovering the behavior by accident.

**AC-UMH-017 — version-string degradation is asserted, not emergent.**
*Given* the gate compares the recorded stamp with `compareVersionLoose`
(`internal/cli/update_version.go:233`, which strips a `go-v`/`v` prefix via `stripVersionPrefix`
and truncates at `-`/`+` in `parseSemverLoose`);
*When* the gate is evaluated against each fixture below;
*Then* the outcome is as stated.

| Fixture `template_version` | Expected gate | Why |
|---|---|---|
| `3.1.3` | open | equal to the constant |
| `v3.1.3` | open | `v` prefix stripped — this repository's own stamp is `v3.1.3` |
| `3.2.0-rc.0` | open | pre-release suffix truncated → `[3,2,0]` |
| `3.1.2` | closed | below the constant |
| `dev` | closed | parses to `[0,0,0]` |
| (field absent) | closed | `GetProjectConfigVersion` returns `"0.0.0"` |

Verify: `go test ./internal/cli/ -run TestUpdateMirrorHeal_VersionMatrix -v` → `--- PASS`, one
sub-test per row. These outcomes are today an emergent property of `compareVersionLoose`; this
criterion converts them into asserted properties, so a future comparator change cannot silently
close or open the gate.

## §D.1 Definition of Done

All 17 criteria PASS; `go test ./internal/cli/... ./internal/template/...` green; `go vet ./...`
clean; evidence exported to `.moai/reports/t520/` before any verdict cites it; the full-suite
verdict left to CI on the develop push (no local `go test ./...`).

## §D.2 Mutant-list coverage audit

The 0.1.0 list omitted AC-UMH-013 — the guard that could not fail — so the SPEC's own
vacuity brake did not reach the one criterion that most needed it. The corrective is structural, not
a single addition: **the mutant list in AC-UMH-014 is derived from the "Absence guard?" column of
the §D matrix, and the two MUST be checked against each other whenever a criterion is added.**
**Retraction (iteration 2, N2).** The 0.2.0 audit asserted "the matrix marks 8 absence guards" and
exempted AC-UMH-015 on the ground that it shares AC-UMH-008's mutant. **Both statements were false**,
and the second was the more dangerous:

- The count was wrong by inspection — the matrix marked **9** rows `yes` (015 included). The audit
  restated a remembered number instead of counting its own table, which is the failure mode the
  audit exists to prevent.
- The sharing claim inverted the direction. The two guards sit on **opposite sides of the same
  filter**: widening the scope set (S1) still yields zero entries on AC-UMH-015's empty-skills
  fixture, so **015 stays green**; removing the target-existence filter still excludes
  `local-only-skill` from AC-UMH-008's fixture, so **008 stays green**. Neither mutant breaks both,
  so no sharing exists and the exemption removed real coverage.

**Audit of this revision, counted rather than recalled.** Matrix rows marked `yes`: 003, 004, 006,
007, 008, 011, 012, 013, 015 → **9**. Mutant rows in AC-UMH-014: **9**, same identifiers. No
exemptions claimed. Any future sharing claim must name the single mutant and argue that **both**
guards go red under it — the test the 0.2.0 claim would have failed.
