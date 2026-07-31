# SPEC-UPDATE-LEGACY-SKILL-LIST-001 — Implementation Plan

Sections are ordered by decision-reversibility: the choices most likely to change on review come first, mechanical steps last.

---

## §A Decisions requiring review first

### D1 — This SPEC adopts NO Epic ordinal, and the siblings are not renumbered

This SPEC is a **member of the `moai update` / `.moai/config` four-lens audit Epic, joined after the initially-numbered set**. It carries no "Epic SPEC N of M" line. Three measured facts drive that:

**(i) The existing "of 6" numbering is already inconsistent with the actual roster.** Six SPECs self-number 1..6 with no gaps:

```
$ for s in <the seven>; do grep -om1 'Epic SPEC [0-9] of [0-9]' .moai/specs/$s/spec.md; done
SPEC-UPDATE-REINSTALL-LOOP-002   Epic SPEC 1 of 6
SPEC-UPDATE-DATA-SURVIVAL-001    Epic SPEC 2 of 6
SPEC-CONFIG-TIER-PERSIST-001     Epic SPEC 3 of 6
SPEC-CONFIG-KEY-HONESTY-001      Epic SPEC 4 of 6
SPEC-UPDATE-CI-GUARD-001         Epic SPEC 5 of 6
SPEC-UPDATE-DOC-DRIFT-001        Epic SPEC 6 of 6
SPEC-UPDATE-YAML-PRESERVE-001    (no ordinal line)
```

But `SPEC-UPDATE-YAML-PRESERVE-001` **is** an Epic member, named as such in two sibling artifacts:

```
$ grep -n 'remaining Epic SPECs' .moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/progress.md
40: | 3+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-CONFIG-TIER-PERSIST-001`, …

$ grep -n 'YAML-PRESERVE' .moai/specs/SPEC-UPDATE-DOC-DRIFT-001/progress.md
121: | 5+ | `SPEC-UPDATE-CI-GUARD-001` (E5), `SPEC-UPDATE-YAML-PRESERVE-001`, **this SPEC** (E6) | …
```

The roster is therefore **7 members, 6 of them self-numbered "of 6"**. An "Epic SPEC 7 of 7" line here would be wrong on both counts — wrong ordinal (YAML-PRESERVE is unnumbered, so no position 7 is free by construction) and wrong denominator relative to every sibling. Dropping the ordinal removes the question rather than answering it incorrectly.

**(ii) Renumbering six sibling artifacts is outside this SPEC's scope.** This SPEC's scope is the `legacySkillIDs` correction, a guard test, three git deletions, and one loop refactor (§E M1-M4). Editing six other SPECs' HISTORY tables is not in it. That is the whole of the argument — it is a scope statement, and it needs no supporting mechanism.

**Decision.** No ordinal in this SPEC; no edit to any sibling. Renumbering the Epic — if ever wanted — belongs to whichever change lands last across it, and would have to reconcile YAML-PRESERVE's missing ordinal at the same time. Out of scope here.

> **Two retracted rationales — recorded because the retraction is the useful artifact.** This decision has been correct since v0.1.0; the *reasons* given for it were wrong twice, each time because a mechanism was supplied for a decision that never needed one.
>
> | Version | Rationale offered | How it failed |
> |---------|-------------------|---------------|
> | v0.1.0 | "The six siblings are mid plan-audit revision, in flight on `plan/epic-update-config-audit`; editing from a different branch would collide." | **False premise.** `git diff --stat main plan/epic-update-config-audit` is empty and `git rev-list --count --left-right main...plan/epic-update-config-audit` → `0	17` — the branch is fully merged and both this SPEC and the siblings live on `main`. No collision was possible. |
> | v0.2.0 | "Editing a sibling's `spec.md` changes the `planArtifactNames` hash, fails skip-eligibility condition 3, and forces a full Phase 1 re-audit on each of six — a real mechanical cost." | **True mechanism, false consequence.** `planArtifactNames` is real (`internal/runtime/audit_cache.go:63-68`, hashing `{acceptance.md, plan.md, spec.md, tasks.md}`) — but (a) skip-eligibility requires score ≥ 0.90 (`spec-workflow.md:319`, condition 2 of four) and the six siblings score 0.81-0.88, so **none is skip-eligible** and each already re-executes Phase 1 on its next run regardless; condition 3 is moot. And (b) `internal/runtime/audit_cache.go:73-74` states verbatim: "Each GateConfig holds its own InMemoryCache instance, so cache entries do not persist across separate /moai run invocations" — there is no cross-invocation cached PASS to invalidate. The claimed cost is **zero**. |
>
> The shared failure mode is not carelessness about facts — both citations resolved, and the v0.2.0 one was verified down to the function. It is **reaching for a causal mechanism to justify a decision that stands on its own**, and then verifying the mechanism's *existence* without verifying its *implication*. That is the same claim-integrity failure class this Epic exists to fix (`verification-claim-integrity.md` §2: a claim is attributed to a command that was run AND an output that was observed — "the function exists" is not the same observation as "the consequence follows").
>
> The rule this SPEC now applies: **when a decision is a scope statement, state the scope and stop.** No third causal claim about branches, caches, or re-audit costs is offered, and none should be added later.

> **Observation, scoped out.** `SPEC-UPDATE-DATA-SURVIVAL-001/progress.md:22` cites `report: .moai/reports/plan-audit/SPEC-UPDATE-DATA-SURVIVAL-001.md` for its iteration-1 verdict. That path does not resolve (`test -f` → missing); the file that exists is `…-epic-update-config-iter2.md`, a different iteration. A cited evidence path that does not resolve is the same claim-integrity defect class this Epic exists to fix (`verification-claim-integrity.md` §2 — a claim must be attributable to an observed artifact). Recorded here so it is not lost; correcting another SPEC's artifact is out of this SPEC's scope.
>
> **Measurement trap worth recording.** Both the plan-audit report and the revision brief reached opposite conclusions about whether those iteration-2 reports exist. The discriminator is that this shell aliases `ls` to long format, so `ls <dir> | grep -E "^SPEC-"` matches nothing — every line begins with `-rw-r--r--@`. A glob (`ls .moai/reports/plan-audit/*epic-update-config-iter2.md`) resolves it. The same alias produced a garbled skill-directory diff during initial authoring. Use `find … -exec basename` or a glob, never `ls | grep '^name'`.

### D2 — Where the guard test lives, and what it derives its truth from

Two viable homes:

| Option | Placement | Trade-off |
|--------|-----------|-----------|
| **A (chosen)** | New file `internal/cli/update_archive_guard_test.go`, package `cli` | `legacySkillIDs` is an unexported `cli` identifier, so the guard must live in `package cli` to read it directly. `internal/template` is already imported by `internal/cli/doctor_skills.go`, so no new dependency edge appears (NFR-LSL-005). |
| B (rejected) | `internal/template`, asserting against a copy of the list | Would require exporting the list or duplicating it — a second copy is precisely the drift source this SPEC exists to remove. |

The guard derives the live set from `template.EmbeddedMoaiSkillNames()` at test time (REQ-LSL-009). It must **not** hard-code a skill inventory: a hard-coded set would be a third copy that can itself go stale, reproducing the defect one layer up.

Note on the helper's set boundary: `EmbeddedMoaiSkillNames()` matches the prefix `moai-` (trailing dash significant), so it returns 30 names and deliberately excludes the bare `moai` unified skill directory. Every `legacySkillIDs` entry carries the `moai-` prefix, so the exclusion is harmless here — but the guard must not be re-written later to compare against a raw `fs.ReadDir` listing, which would include `moai` and change the set size from 30 to 31 for no benefit.

### D3 — What "graceful degradation" means for the guard (REQ-LSL-007)

`EmbeddedMoaiSkillNames()` returns `([]string, error)` and its doc comment says callers **must treat an empty derived set as "manifest unavailable" and degrade gracefully rather than** mis-classify. Two failure shapes, one response:

- non-nil `error` → `t.Skip` with the error text
- `nil` error but zero names → `t.Skip` with a "manifest empty" message

`t.Skip` is chosen over `t.Fatal` because a read failure of the embedded FS is an environment fault, not a list defect; failing would produce a red suite that misdirects the reader to the wrong file. `t.Skip` is chosen over silently passing because a silent pass is a vacuous green — the exact failure mode this SPEC's §A Defect 5 documents.

The degradation path needs its own coverage. Because the real embedded FS always reads successfully in a compiled binary, the degradation branch cannot be reached by calling the helper. The implementation therefore factors the comparison into a small pure function over `(legacyIDs []string, embedded []string, err error)` and the test drives that function with synthetic inputs. Keeping the pure function unexported and test-adjacent avoids widening the package API for a test.

### D4 — Aggregate-error shape for `archiveLegacySkills` (REQ-LSL-013 … REQ-LSL-016)

The current signature is `func archiveLegacySkills(projectRoot string, out io.Writer, force bool) (int, error)`. It stays. What changes is the loop body:

- per-entry failures append to a slice of `(id, err)` instead of returning
- the loop always reaches the `total:` summary emission
- after the loop, `nil` is returned when the slice is empty; otherwise a single error naming every failed ID

Chosen construction: `errors.Join` over per-ID wrapped errors, so each individual `%w` chain is preserved and callers can still `errors.As` a `*MigrateError` out of it. A hand-rolled concatenated string would discard the wrapped `ARCHIVE_DRIFT` codes that the force-path tests inspect.

Three return sites inside the loop must all be converted, not just the obvious one: the two drift-backup errors (`create drift backup parent for %s`, `backup drift archive for %s`) and the archive error (`archive %s`). Converting only the last would leave the abort behaviour reachable through the `--force` path.

Call-site impact, verified: the only production caller is the "Post-sync steps" block in `internal/cli/update.go`, which does `if _, archiveErr := archiveLegacySkills(...); archiveErr != nil { … CheckLine("warn", …) }` and continues regardless. An aggregate error is strictly more informative there and needs no call-site change.

### D5 — `.gitignore` for the archive tree: follow-up, not scope

`.gitignore` currently has no `archive` rule (`grep -n archive .gitignore` → no match, exit 1). Adding one would be a blunt instrument: a bare `.moai/archive/` rule does not un-track the four already-tracked genuine files, so the tree would end up half-tracked and half-ignored — a state harder to reason about than either extreme.

**Decision: out of scope.** M3 removes the three wrong files from the index and leaves the four genuine ones tracked exactly as they are. Whether the archive tree should be ignored wholesale is a separate question about user-data-in-git policy, and it interacts with the downstream-cleanup follow-up (spec.md §C). Recorded as a follow-up so the omission is a decision, not an oversight.

---

## §B Milestone numbering (mapping to the original triage)

The delegating triage numbered the archive-loop work **M5**. This plan uses sequential numbering:

| This plan | Original triage | Work |
|-----------|-----------------|------|
| M1 | M1 | Correct `legacySkillIDs` |
| M2 | M2 | Cross-check guard test |
| M3 | M3 | Remove the three wrong archive files from git |
| **M4** | **M5** | Non-aborting archive loop |

There is no gap in scope — the triage's numbering simply skipped 4.

---

## §C Verification of every claim carried into this plan

Every `file:line`, count, and command in the delegation brief was re-run before being written into an artifact. Results:

| Claim | Verified | Note |
|-------|----------|------|
| `legacySkillIDs` holds 16 entries | yes | block-scoped `awk` + `grep -c '"moai-'` → `16` |
| The three IDs exist in the template tree | yes | `SKILL.md` present for all three |
| Catalog registration lines 157 / 162 / 220 | yes | `grep -n` reproduced all six lines (name + path per skill) |
| Reference counts 46 / 25 / 17 | **corrected** | measured 56 / 30 / 26 files (68 / 30 / 34 occurrences). Conclusion unchanged. |
| 31 template = 31 catalog = 31 local | yes, with precision | local is 30 `moai-*` **plus** the bare `moai` directory = 31. `find`-based diff shows the only difference is the bare `moai`, which a `moai-*` glob excludes. |
| Clean-reinstall globs `.claude/skills/moai*` | yes | `isGlob: true` target in the `targets` slice |
| `.moai/archive/` not in the clean list; only `.moai/config` removed | yes | read the full target slice and the `os.RemoveAll(configDir)` block |
| Archive runs after redeploy, in Post-sync steps | yes | read the `tui.Section("Post-sync steps", …)` block in `update.go` |
| Live vs archive md5 differ for all three | yes | md5 table in spec.md §A Defect 2 |
| Git timeline `74bae50f4` / `ec0e9e257` / `697a6e2c7` | yes | dates 2026-04-27 / 04-27 / 04-28 confirmed |
| `git log -S"moai-domain-backend" -- update_archive.go` → 1 commit | yes | `ec0e9e257` only |
| Six test files reference the list, all self-referential | yes | all six read the list and seed `t.TempDir()` fixtures from it |
| `update_archive_force_test.go` uses `[0]`, `[1]`, `[2]` | **extended** | it also uses `legacySkillIDs[3]`. `update_idempotency_test.go` uses `[:8]`; `update_skip_sync_test.go` uses `[0]`. |
| `EmbeddedMoaiSkillNames()` contract | yes | read verbatim; already consumed by `internal/cli/doctor_skills.go` |
| Intersection is exactly 3 | yes | temporary in-package probe: `legacySkillIDs=16 embedded=30 overlap=3 [backend frontend database]`; probe file removed, tree clean |
| 7 tracked archive files, 7 directories | yes | `git ls-files` → 7; `find -type f` → 7 |
| Committed by `9373e558f` | yes | `git log -1` on one of the paths |
| `.gitignore` has no archive rule | yes | `grep -n archive .gitignore` exits 1 |
| Deleted legacy body had `references/` | yes | `git ls-tree 74bae50f4^` → `SKILL.md`, `references/examples.md`, `references/reference.md` |
| `restoreSkill` `--force` does `os.RemoveAll(targetDir)` | yes | read the function |
| Prior SPEC declared the list out of scope | yes | `spec.md:76`, `plan.md:44`, `acceptance.md:268` |
| Baseline suite is green pre-fix | yes | `go test ./internal/cli/ -run 'TestArchive\|TestRestoreSkill\|TestSkipSyncNoArchive' -count=1` → `ok … 0.776s` |

**Not in the brief, found during verification** (folded into spec.md §A as Defect 4): `internal/template/skills_removal_test.go` was already narrowed from 16 to 9 entries with an explicit "still exist in the template tree" comment. The revival was noticed in the template package and the CLI list was never brought along. This is the strongest single piece of evidence that the list is an un-propagated correction rather than an intentional design.

---

## §D Constraints

1. `development_mode: tdd` — each milestone writes its failing test first, observes the failure, then makes it pass.
2. Every filesystem fixture uses `t.TempDir()` (NFR-LSL-001, CLAUDE.local.md §6).
3. No file under `internal/template/templates/` is touched (NFR-LSL-002).
4. The `archive: ` and `total: ` output literals are preserved (REQ-LSL-017) — an existing test asserts on both.
5. Only files inside `internal/cli/`, plus the three archive-file deletions, are modified.
6. Positional test indices must stay in range: the largest index used is `legacySkillIDs[3]` and the largest slice is `[:8]`; a 13-entry list satisfies both.

---

## §E Milestones (execution order)

### M1 — Correct `legacySkillIDs` (REQ-LSL-001 … REQ-LSL-004)

Order within M1 matters: the guard from M2 must not be written first, or M1's own removal cannot be observed as the thing that turns it green.

1. Delete the three entries from the `var legacySkillIDs = []string{` block in `internal/cli/update_archive.go`, leaving 13 in their existing relative order.
2. Update the doc comment: it currently reads "lists the 16 skill IDs removed in BC-V3R3-007". It must state 13 and record why three were withdrawn (revived by `697a6e2c7` on 2026-04-28; live template skills, so archiving them fed a permanent drift loop).
3. Run the six pre-existing test files and confirm green. Expected outcome, from the index analysis above: no failure. Every one of them derives its fixtures from the list, and every positional access stays in range. Should any fail, adjust the *test*, never the list.
4. Decide the `…_All16Skills` / `…_All16RoundTrip` rename. These names become misnomers at 13 entries. A rename is behaviour-neutral and cheap; leaving them is also defensible since the bodies iterate the list dynamically. Whichever is chosen, record it — AC-LSL-006 accepts either, but requires the choice to be observable.

### M2 — Cross-check guard (REQ-LSL-005 … REQ-LSL-009)

1. Write **`internal/cli/update_archive_guard_test.go`** (filename pinned — three ACs reference it by path) with `TestLegacySkillIDsNotEmbedded`, plus the pure comparison function D3 describes and its degradation subtests.

   **Required test shape** (pinned, because the AC counts depend on it): the production assertion lives **directly in the parent function body**, and the two degradation cases are subtests named `manifest_error` and `manifest_empty`. Go's `-v` output prints a `--- PASS` line for the parent *and* for every passing subtest, so a third `production` subtest would make the parent-plus-subtest PASS count 2 where the ACs expect 1. The ACs additionally anchor their greps to `^--- PASS:` / `^--- FAIL:` so the count survives a later refactor into subtests — but the shape is pinned anyway so the two defences are independent.
2. **Falsify it before trusting it, without touching the working tree.** Use `go test -overlay` with a mutated copy of `update_archive.go` that re-adds one live ID; observe `--- FAIL` and confirm the message names that ID. The source file is never modified, so no revert step is needed and no fix can be destroyed. Procedure and post-conditions: acceptance.md §C.1. A guard that has only ever been observed passing has not been shown to catch anything (REQ-LSL-008).
3. Confirm the degradation subtests skip — not pass, not fail — for both the error and the empty-set inputs.

> **No commit precondition.** The overlay-based falsification in step 2 replaces file content only at compile time, so M1 does **not** need to be committed before M2 begins. This is the reason the overlay form was chosen over a `git restore --source=<sha>` form: it removes the ordering constraint instead of documenting one.

### M3 — Remove the three wrong archive files from git (REQ-LSL-010 … REQ-LSL-012)

1. `git rm` the three `SKILL.md` paths under `.moai/archive/skills/v2.16/moai-domain-{backend,database,frontend}/`.
2. Verify `git ls-files .moai/archive/skills/v2.16` now lists exactly 4 files, all under the four genuine directories.
3. Verify the four genuine files are byte-unchanged (`git diff --stat` shows no modification to them).
4. Do not add a `.gitignore` rule — see §A D5.

### M4 — Non-aborting archive loop (REQ-LSL-013 … REQ-LSL-017)

1. Write the failing test first, in **`internal/cli/update_archive_continue_test.go`** (filename pinned — three ACs reference it by path): a fixture where one entry cannot be archived and at least one *later* entry can, asserting that the later entry **is** archived and that `total:` is emitted.

   Injecting a per-entry failure without a `t.Skip` on Windows needs care. Preferred mechanism: seed the archive **destination** path for one ID as a regular file rather than a directory, so `archiveSkill`'s `MkdirAll` fails on a portable `ENOTDIR`-class error. A permission-based fixture (`chmod 0o000`) is the fallback but is skipped on Windows and is a no-op for root, so it is the weaker choice.
2. Convert all three in-loop `return archived, …` sites to accumulate.

   **`continue` semantics are explicit, not implied.** After recording a failure, control MUST `continue` to the next entry — it must NOT fall through into the remaining per-entry work for that same entry. This matters because the two drift-backup sites sit *above* `archiveSkill` in the loop body: a literal `return → append` substitution without a `continue` would fall through into `os.Rename` and `archiveSkill` for an entry whose backup parent could not be created. That is latent rather than immediately breaking (the subsequent `os.Rename` also fails, so `TestArchiveForce/force_with_drift_backup_failure_preserves_original` still passes either way), which is exactly why it must be stated rather than left to the existing tests to catch.
3. Emit `total:` unconditionally after the loop; return `errors.Join(...)` of the wrapped per-ID errors, or `nil` when none failed.
4. Confirm the `archive: ` / `total: ` literals are unchanged and the existing output assertions still pass.
5. Confirm `archiveVersion` is still `"v2.16"` — spec.md §B Goal 5 and NFR-LSL-003 require the constant and the directory scheme to be untouched, and this milestone is the only one that edits the file the constant lives in (AC-LSL-015).

---

## §F Anti-patterns to avoid

- **Editing a test to match the list instead of the list to match reality.** The six existing tests pass today *because* they are self-referential. Making the new guard self-referential too would produce a seventh vacuous test.
- **Hard-coding the live skill set in the guard.** A third inventory copy re-creates the defect one layer up. Derive it from the embedded manifest.
- **Trusting an unfalsified guard.** M2 step 2 is not optional. "It passes" is not evidence it can fail.
- **Using `-run` selectors that match nothing.** A `-run` naming a nonexistent test exits 0. Every AC that runs a test first greps that the test name exists.
- **Silently passing on manifest-unavailable.** Skip, do not pass.
- **Converting only the last in-loop return in M4.** The two drift-backup returns above it are equally abortive and are reachable via `--force`.
- **Deleting the four genuine archive entries "for consistency".** They are the only legitimate content in that tree.
- **Renumbering the sibling Epic SPECs.** See §A D1.

---

## §G Cross-references

- spec.md §A Defect 1-7 — the verified defect set
- spec.md §C — the four out-of-scope items, including the `restore-skill` follow-up
- acceptance.md §B — the AC set and each AC's observed pre-fix baseline
- acceptance.md §C — the falsification procedure for the guard and for the archive loop
- `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001` — the prior SPEC whose `--force` path is a symptom workaround for this defect
