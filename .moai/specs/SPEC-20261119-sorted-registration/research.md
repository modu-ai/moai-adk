# C-75 plan: measured before/after on 10+ historical conflicting merges

Date: 2026-11-19 · tree: 061177768c (origin/v2026.11.0) · method below. Scratch scripts live in
the worktree's ignored `target/` (find_conflicted_merges.sh, simulate_sorted.py) and are not
committed.

## Method

1. Walk the last 400 merge commits of `origin/v2026.11.0`; for each, run
   `git merge-tree --write-tree --name-only <p1> <p2>` on its parents and keep the merges whose
   parents still conflict today in `src/main.rs`, `src/lib.rs` or `docs/rust-product-path.md`
   (12 found; the card asked for ten).
2. **before** = conflict-marker count of `git merge-file -p --diff-algorithm=histogram` over the
   pair's three versions (ours=p1, base=merge-base, theirs=p2) of the conflicted file.
3. **after(sorted)** = the same merge after transforming both sides' file to the proposed form:
   table rows sorted by row (docs), contiguous `pub mod` lines sorted (lib.rs), variant blocks
   sorted (main.rs).
4. **after(rows removed)** = the same merge after deleting the committed table rows on both
   sides (the "generated, not committed" design: the doc keeps markers and a pointer, the table
   is emitted at docs-build time, like `ops_registry.rs` in `OUT_DIR` since #2617).

Caveat: a merge-tree replay judges the pair on today's tree — intervening merges can make a
pair conflict now that merged cleanly then, and the other way round. The numbers compare
designs on the same tree, which is what the plan needs.

## Measured pairs

| merge | file | before | after(sorted) | after(rows removed) |
| --- | --- | --- | --- | --- |
| 430095e84b9d | src/lib.rs | 1 | 1 | — |
| 59d16b68f774 | src/main.rs | 1 | 1 | — |
| 82bcc208bf46 | docs/rust-product-path.md | 1 | 0 | 0 |
| c4426155f971 | docs/rust-product-path.md | 1 | 1 | 1 |
| 64837aa2ca2d | src/lib.rs | 1 | 1 | — |
| 2f2bb9f38b1e | docs/rust-product-path.md | 1 | 0 | 0 |
| 1f94a74210c9 | docs/rust-product-path.md | 1 | 1 | 0 |
| 6394155352cd | docs/rust-product-path.md | 1 | 1 | 0 |
| fb945ff2dcd8 | docs/rust-product-path.md | 1 | 1 | 0 |
| bf04be357d78 | docs/rust-product-path.md | 1 | 1 | 0 |
| 97ec516e781a | docs/rust-product-path.md | 1 | 1 | 0 |
| d7fb43d224f7 | src/lib.rs | 1 | 1 | — |

Totals: docs table pairs 8 — sorted helps 2, rows-removed helps 7 (the one that stays,
c4426155f971, edits other regions of the doc too). lib.rs 3 and main.rs 1 — sorting helps none.

## Conclusions

- **C-1 (recommended): the operation table must not be committed.** `docs/rust-product-path.md`
  keeps `<!-- ops-table-begin/end -->` markers and a one-line pointer ("generated:
  `tools/gen_ops_table.py`"); the rows move into their per-operation fragments
  (`docs/design-review/appendix/operations/<word>.md`) as a structured block, and the generator
  assembles the table for the docs build and for a CI check. This removes 7 of 8 measured table
  conflicts and is the same pattern the operation registry already uses for code (#2617).
  Cost: a one-off migration of ~150 rows into fragments (scripted, one PR), and the fragments
  gain a required section (REQ in the follow-up SPEC; malformed fragments refuse the build).
- **C-2: sorting `src/lib.rs` and `src/main.rs` lists is not supported by measurement.** All
  four non-docs pairs keep their conflict under the sorted transform: the historical conflicts
  are adjacent-line edits in one region, which sorting does not separate. Doing it would churn
  `--help` goldens (`all*.txt`) for no measured gain — skip.
- **C-3: what already works stays.** New `src/ops/<word>.rs` files already merge without a
  shared-line change (build.rs registry, #2617); nothing to do there.

## Follow-up slices (after the leader picks C-1's fragment format)

1. Fragment schema + scripted row migration (one PR, mechanical, generator refuses a fragment
   missing the section).
2. `tools/gen_ops_table.py` (deterministic assembly, `--out` for docs builds, `--check` for CI)
   + marker swap in the doc + doc-update flow sentence.
3. Pins: unit test for determinism and refusal, docs-build step, the word-scan that keeps the
   committed doc free of `| ` rows between the markers.
