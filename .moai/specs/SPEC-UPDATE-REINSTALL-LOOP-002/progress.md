# SPEC-UPDATE-REINSTALL-LOOP-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

### Baseline

- **Code baseline** `d5336214e`; **worktree HEAD** `5468a4afc` (branch `plan/epic-update-config-audit`), a descendant of `d5336214e` that changes SPEC documents only.
- `git merge-base --is-ancestor d5336214e HEAD` → exit `0`. `git diff --name-only d5336214e...HEAD -- 'internal/cli/*.go' | wc -l` → `0`, confirming no Go source differs between the two.
- Every artifact that names a baseline names `5468a4afc`, citing `d5336214e` only as the code baseline it inherits.

### v0.1.0 (initial plan-phase authoring)

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- Every claimed `file:line` in the delegation brief was independently re-verified. All matched; see plan.md §C for the verification table.
- One additional defect was found during verification and folded into scope: clean-reinstall Step 4 never calls `backupDeprecatedPaths` (spec.md §A Defect 5, plan.md §B).
- The version matrix and the 9-entry preserve-root intersection were reproduced by an executed probe, not carried over from the brief.
- Defect 3 decision recorded and justified in plan.md §A D2 (narrow the override's consequence; do not defer).

### v0.2.0 (plan-audit revision — D1 through D5)

Plan-audit verdict on v0.1.0: **FAIL, 0.77** against the Tier M threshold 0.80; Must-Pass 7/7 passed; baseline integrity found sound (no unobserved-baseline defects). D1-D5 resolved below; D6 and D7-a deferred to the next audit.

- **D1** — `REQ-RIL2-003` narrowed to `v`/`V`-prefixed major-2 forms, so it no longer contradicts `NFR-RIL2-001`; the constraint takes no exception. `AC-RIL2-003` rebuilt on a **residue-free** fixture with `2.5.0`, `V2.5.0`, and `abc` in its input set. spec.md §A gained the residue-free widening table; §G risk row 2 rewritten.
- **D2** — plan.md §E M4's first option (relocate the dry-run early return) dropped as a confirmed defect and recorded as a rejected alternative with its evidence. `--dry-run` must never mutate; the fix hoists detection above the branch so the already-implemented non-mutating renderer at `update_clean_install.go:186-198` becomes reachable.
- **D3** — plan.md §E M3's `:384-390` corrected to `:406-412`, with the condition token given as the durable anchor.
- **D4** — `AC-RIL2-015` replaced: the tautological package-relative count became a fixed-literal assertion on `[dry-run] total:` and `use a worktree for isolation`, both read from the current sources.
- **D5** — acceptance.md §C.2 split into Batch A (compiles pre-fix, `--- FAIL` required) and Batch B (build failure is a valid falsification, with the undefined symbol named in stderr); `<pre-fix-commit>` bound to the run-phase base recorded in §E.2 below; new §C.3 gives `AC-RIL2-003` its own mutation-based falsification.
- **Coverage gaps closed** — `AC-RIL2-019` binds `REQ-RIL2-021`; `AC-RIL2-020` binds `NFR-RIL2-004`, diff-scoped after a package-wide draft was run and rejected (it fails on five pre-existing unformatted test files and one pre-existing camelCase filename).
- **Additional finding, not in the audit's D-list** — plan.md §A D1 step 4 read "unparseable → Signal 1 positive", ambiguous between "file unparseable" and "major digits unparseable". Under the second reading a residue-free `abc` moves `IsV2` false→true, an `NFR-RIL2-001` violation of the same class as D1. Step 4 and `REQ-RIL2-004` / `REQ-RIL2-006` now state the file/value distinction explicitly.

### Observed at revision time (2026-07-31, HEAD `5468a4afc`)

- `moai spec lint` → `0 error(s), 62 warning(s)`; `moai spec lint | grep -c 'SPEC-UPDATE-REINSTALL-LOOP-002'` → `0`. Zero findings for this SPEC; the 62 warnings all belong to other, grandfathered SPECs. This replaces the v0.1.0 line that asserted the output "recorded at plan-phase handoff" without recording it.
- `grep -n 'fpErr == nil && !fingerprint.IsV2 && isMoAIProject' internal/cli/update.go` → `406:` (D3 evidence).
- `grep -rn 'func TestUpdateDryRun' internal/cli/*_test.go` → no matches; `go test ./internal/cli/ -run 'TestUpdateDryRun' -v | grep -c -- '--- PASS'` → `0` (D4 evidence: the retired AC compared `0` against a tree-derived expectation of `0`).
- Residue-free classification probe (temporary, removed after use): `2.5.0` → `Signal 1=false, IsV2=false`; `V2.5.0` → `Signal 1=false, IsV2=false`; `abc` → `Signal 1=false, IsV2=false`; `v2.5.0` → `Signal 1=true, IsV2=true` (D1 evidence).
- Residue-carrying probe, run against **the v0.1.0 seven-row matrix** (not the revised nine-row table now in acceptance.md §B AC-RIL2-001): `IsV2=true` for all six non-v3-confirmed rows of that seven-row table, confirming the v0.1.0 matrix fixture cannot discriminate a Signal-1 change. Under the pre-change literal `"v3."` rule only `v3.0.1` was v3-confirmed there, so six of seven rows were non-v3-confirmed — **six is the accurate historical count for the v0.1.0 table and is deliberately retained.** The nine-row table has four such rows; that separate figure lives in acceptance.md §B.

### v0.3.0 (plan-audit iter3 revision — D12, D9, D15)

Plan-audit verdict on v0.2.0: **PASS, 0.88** against the Tier M threshold 0.80; Must-Pass 7/7. Three residual defects closed; no code changed (documentation-only revision).

- **D12 (major) — `AC-RIL2-014` command (a) was satisfiable by a comment alone.** The literal-match grep ran over the whole test file, and grep cannot distinguish code from prose, so a fixture whose `permissions.deny` array is empty and whose only occurrence of a `retiredV2DenyEntries` literal sat in a comment satisfied the check. This is the same defect class the sibling `SPEC-UPDATE-DATA-SURVIVAL-001` removed from its own AC-UDS-M5 command (b). Closed on two levels: (i) command (a) now strips leading-comment lines (`grep -v '^[[:space:]]*//' | grep -cE …`), verified to print `0`/exit 1 on a comment-only fixture and `1`/exit 0 on a genuinely-seeded one; (ii) the AC now **requires `TestUpdateDryRun_ZeroMutation` to assert at runtime** that the written `.claude/settings.json` parses and its `permissions.deny` array contains one of the twelve literals. The runtime assertion is the durable fix — the comment-stripping grep alone is still defeated by a *trailing* comment on a code line (observed: prints `1`, exit 0), and that residual gap is stated explicitly in the AC body. The §D Definition of Done clause was rewritten to reference all three checks rather than "both halves".
- **D9 (minor) — the two "six non-v3-confirmed rows" occurrences are NOT the same defect, and were handled differently.**
  - `acceptance.md:37` was genuinely wrong and is fixed. Its referent is the revised **nine-row** AC-RIL2-001 table, whose `V3VersionConfirmed = false` rows are `""`, `v2.5.0`, `2.5.0`, `V2.5.0` — **four**, not six. Four is the number that makes the sentence's own argument true, since exactly those four hold `IsV2 = true` on both sides of the change; the other four v3-confirmed rows (`3.0.1`, `V3.0.1`, `v4.0.0`, `4.0.0`) flip `true → false`.
  - `progress.md:37` (above, in the v0.2.0 observation block) was **correct as written and its number was deliberately NOT changed.** Its referent is explicitly the v0.1.0 **seven-row** matrix, where the pre-change literal `"v3."` rule left only `v3.0.1` v3-confirmed — so six of seven rows were non-v3-confirmed. **The iter3 audit classified this as a second occurrence of the stale figure; that classification is rejected.** Changing it to "four" would have corrupted an accurate historical observation by re-scoping it to a table it never described. The line was instead edited to name its referent (the v0.1.0 seven-row table) unambiguously, so no future reader or auditor mistakes it for the current nine-row table.
- **D15 (minor) — the sibling co-edit constraint is now recorded in this SPEC.** `SPEC-UPDATE-DOC-DRIFT-001/progress.md` §E.1 ("M1 versus E1") carried a constraint this SPEC did not: the two SPECs must not edit the `--dry-run` branch of `internal/cli/update.go` concurrently, and if the sibling lands first this SPEC's M4 becomes a no-op verification rather than an implementation. **Recorded in `plan.md` §E M4** (extending the existing consistency note) rather than here, because it changes *what M4 does* — degrading it from "implement the hoist" to "verify the hoist and record that no change was required" — and is therefore a plan constraint, not merely a sequencing note. The sibling's own claim that it settles `--dry-run` the same way was independently confirmed at `SPEC-UPDATE-DOC-DRIFT-001/spec.md:373` (option B selected) and `:376-378` (early return retained), and those line anchors are now cited in the plan.md note.

## §E.2 Run-phase Evidence

_<pending run-phase>_

- `pre_fix_commit:` _<pending — capture `git rev-parse HEAD` at run-phase entry, before M1's first implementation commit. This SHA binds acceptance.md §C.2 Batch A / Batch B and AC-RIL2-020.>_

### Epic run order (depends_on sequencing)

`SPEC-UPDATE-REINSTALL-LOOP-002` declares `related_specs`, not `depends_on`, so its own run-phase `depends_on` pre-flight is trivially satisfied. The sibling Epic SPECs are all `status: draft`, which is the expected state for an Epic whose members have not yet run — it is a sequencing fact, not a per-SPEC defect.

Intended order within the Epic, so that any dependency is satisfied by sequencing rather than by an `--ignore-deps` override:

1. `SPEC-UPDATE-REINSTALL-LOOP-002` (this SPEC) — the detection and clean-reinstall correctness base.
2. `SPEC-UPDATE-YAML-PRESERVE-001` — disjoint scope (YAML merge fidelity); may run in parallel.
3. Remaining Epic siblings, in the order the Epic entry point assigns.

Where a sibling later adds a `depends_on:` entry naming this SPEC, that entry is satisfied only when this SPEC reaches `status: completed` — the strict fulfilment definition. No sibling should be entered on `--ignore-deps` to work around draft status.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
