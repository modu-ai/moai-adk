# t2121 progress record (Class B, no SPEC, cycle_type=tdd)

## What changed
- `internal/factorylane/merge_specless.go`
  - `checkSyncAuditVerdictFile` takes the integration ref (`MergeTripleInput.Develop`, already used by the other merge conditions; no new config key). `merge.go` passes `in.Develop`.
  - New `absorbOnlyChange`: when audited_sha is an ancestor of HEAD but the evidence-dir-only path check fails, accept iff (a) every first-parent commit in audited..HEAD is either a merge whose extra parents are all ancestors of the integration ref, or touches only `.moai/reports/<card>/`; and (b) the card diff against its merge-base with the integration ref, excluding the evidence dir, is the same at audited_sha and at HEAD. Comparison is `git diff -U0 --no-renames` with `index` lines dropped and `@@` positions blanked (blob ids and line numbers legitimately shift on absorb; the removed/added lines are what the audit read). Unresolvable integration ref, merge-base failure, or diff failure fails closed.
  - Evidence file: `verdict.md` when present, else latest `sync-audit*.md` (highest `-iter<N>`, no suffix = 1). Own small helper: the existing iter pickers live in other packages (closure, homestate, runtime) and are unexported or plan-audit-specific; importing them risks cycles.
  - Refusal text now names both `verdict.md` and `sync-audit*.md` and the absorb allowance.
- `internal/factorylane/merge_specless_test.go`: `TestMergeTripleSpecLessRefusalText` pins the exact refusal text; updated to the new wording (required consequence of "mention both names").
- New `internal/factorylane/merge_specless_absorb_t2121_test.go` (real temp repos): clean absorb passes; absorb then evidence; late code change after absorb refused; late code change without absorb refused; evil merge editing a card-owned file refused; evil merge editing a path outside the card diff refused; merge of a non-main branch refused; unresolvable integration ref refused; audited==HEAD and evidence-only regressions pass; sync-audit.md only; iter2 beats base; iter10 beats iter2; verdict.md wins; absent names both.

## Commits
- RED e6f5ad9c51 (red-output.txt: 2 test funcs failing, exit 1)
- GREEN be6e18347b

## Commands run and results
- `go test ./internal/factorylane/ -run T2121 -count=1` before fix: FAIL (see red-output.txt); after fix: `ok ... 90.808s`
- `go test ./internal/factorylane/ -count=1` (timeout 570): `ok  github.com/modu-ai/moai-adk/internal/factorylane 110.934s`
- `go test ./internal/cli/ -run 'MergeReady|SpecLess|Specless|FactoryMerge' -count=1`: `ok  github.com/modu-ai/moai-adk/internal/cli 95.634s`
- `go vet ./internal/factorylane/`: exit 0, no output. `gofmt -l internal/factorylane`: empty.

## Gaps
- Full `./internal/cli/` suite not run (focused pattern only); no repo-wide run; CI owns the repository-wide verdict (PENDING).
- Tool-provenance: tests ran via `go test` from this tree (built from tree); the installed `moai` binary was not used.
- No `moai factory` verb run, per instruction.

## Residual risk
- Conservative false refusals: if main touched lines adjacent to the card's hunks, -U0 patches stay equal, but a main change overlapping the card's own file region changes the merged file content and the card diff may differ, forcing re-audit (fail-closed direction).
- Hunk-position blanking means a pure line reordering by an evil merge inside a card file would be undetected only if the removed/added line multiset order is unchanged; content edits are detected.
- An absorb of a main tip newer than the one audited is accepted; an absorb of an older main commit also passes (parents only need to be ancestors of the current ref).

## Repair round (sync-audit FAIL 0.70 @7723481e7e: F1, F2)

### What changed
- `internal/factorylane/merge_specless.go`
  - F1: `cardPatch` diff now runs with `--no-textconv --binary` (plus the existing `--no-ext-diff`), so binary and `-diff` file content is part of the compared patch even though `index` lines are still dropped.
  - F2: for every first-parent absorb merge `M` in audited..HEAD, new `remergeMismatch` runs `git merge-tree --write-tree p1 p2` and requires the first output line to equal `M^{tree}`. Refused: octopus merge (more than 2 parents), re-merge conflict (exit 1, "resolved by hand"), tree mismatch, and any other merge-tree failure (fail closed; the refusal text names the git >= 2.38 requirement). The card-diff equality check stays as the second layer.
  - `@@` blanking kept: mutation test (blanking disabled) fails `P10_clean_absorb_with_shifted_line_numbers_passes`, so it is still needed. Recorded in the `absorbOnlyChange` comment.
- New `internal/factorylane/merge_specless_repair_t2121_test.go` (real temp repos): P1 binary swap, P7 `-diff` text edit, P2b moved hunk, octopus, conflict-resolved merge (all refuse); P10 clean absorb with main shifting line numbers (passes, also audit F3 regression); unsupported-git runner (merge-tree exit 129) refuses with "2.38".
- F5 (`sync-audit-verdict*.md`) skipped: not trivially safe, the pattern's iteration pick would need a new ordering rule; the failure direction is a false refusal (fail closed). Not done.

### Commits
- RED fa64aa0e35 (`.moai/reports/t2121/red-repair-output.txt`: 5 subtests + unsupported-git test FAIL, P10 PASS, exit 1)
- GREEN 58dd8cb496

### Commands and results
- `go test ./internal/factorylane/ -run T2121AbsorbRepair` before fix: `FAIL ... 21.971s`; after fix: all 7 PASS.
- `go test ./internal/factorylane/ -count=1` (timeout 580): `ok  github.com/modu-ai/moai-adk/internal/factorylane	115.134s`
- `go test ./internal/cli/ -run MergeReady`: `ok ... 57.494s`; `-run SpecLess`: `ok ... 48.046s`; `-run Specless`: `ok ... 1.388s [no tests to run]`; `-run FactoryMerge`: `ok ... 25.484s`
- `go vet ./internal/factorylane/`: exit 0. `gofmt -l internal/factorylane`: empty.

### Gaps
- The worktree-isolation guard refused the single `-run 'MergeReady|SpecLess|Specless|FactoryMerge'` command (quoted `|` pattern, also with `\|`); the same four patterns were run as four separate commands instead (substitution, not the original one-line form).
- CI runner git version unchecked (`merge-tree --write-tree` needs >= 2.38; local `git version 2.54.0`); older git refuses every absorb (fail closed).
- Full `./internal/cli/` and repo-wide suites not run; CI owns that verdict (PENDING). Tests ran from this tree via `go test`, not the installed `moai`.
- F6/F7 (outside the diff) untouched.

### Residual risk
- A clean auto-merge whose semantics interact between main's and the card's changes in one file is not detectable by text-level checks.
