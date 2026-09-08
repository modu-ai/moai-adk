# SPEC-UPDATE-HOOK-DELIVERY-001 — Research (code evidence)

> All measurements below were verified against THIS tree: worktree `.claude/worktrees/t466`, branch `WT-update-hook-delivery`, HEAD `d592b0551`. Line numbers are pinned to that SHA.

## §A The defect chain (verified verbatim)

**Fact 1 — base derivation is map-only recursive.** `internal/cli/update/merge/base.go:112-131` `pruneToShared`:
- recurses into nested `map[string]any` only (`:124`: `if updatedIsMap && currentIsMap { pruned[key] = pruneToShared(...) }`);
- copies non-map values wholesale from the template side (`:128`: `pruned[key] = updatedVal`);
- excludes keys absent from the user's file from the base (`:116-121`).

Hook event keys in `.claude/settings.json` hold ARRAYS of hook entries — they take the `:128` wholesale-copy path, so the base's array is always the template's array, never a per-entry history.

**Fact 2 — the "only user changed" misclassification.** `internal/merge/strategies.go:427-429` (verbatim):
```go
case baseChanged && !updChanged:
	// Only user changed.
	result[key] = curVal
```
`valuesEqual` (`strategies.go:685-692`) JSON-marshals both operands and compares strings — arrays deep-compare. Consequence chain, strictly: base array == template array (Fact 1) ⇒ template's added entry makes `updChanged=false`; user's array differs from base ⇒ `baseChanged=true` ⇒ case at `:427` fires ⇒ user's array kept ⇒ template entry dropped, silently.

**Fact 3 — no detection surface.** `internal/cli/doctor.go:768-783` `checkHooksConfig` does exactly one `os.Stat` on the `.claude/hooks/` directory (`:771`) and never opens `.claude/settings.json`. No doctor check or update-output path compares template hook entries against the user's file.

**Scope confirmation — add-blind, not delete-blind** (derived from the same code, logically strict):
- User DELETES an entry: base still equals template ⇒ `baseChanged=true, updChanged=false` ⇒ `:427` keeps the user's (shortened) array ⇒ deletion correctly preserved. This protective behavior is a REQ-UHD-002 invariant.
- Template ADDS a whole event key: absent from user file ⇒ excluded from base (`:116-121`) ⇒ merge classifies template-introduced ⇒ delivered. Preserved as REQ-UHD-001.
- Template ADDS/MUTATES inside a carried event key's array: the defect surface — dropped silently (Facts 1+2).

## §B Wire-up evidence

`internal/cli/update/merge/merge.go:301` lists `settings.json` among the high-risk core config files on the merge path — the settings merge is a deliberate, protected surface, which is why a targeted (not generic) fix is required: a generic `pruneToShared` recursion into arrays would resurrect user-deleted entries everywhere, violating REQ-UHD-002 across all config, not just hooks.

## §C Test-coverage gap (measured)

`internal/template/settings_test.go` carries 31 `func Test*` declarations (counted via `grep -c '^func Test'` on this tree) — all template-shape tests. Zero tests cover update-time delivery of hook entries into an existing project's settings.json. The regression class this SPEC guards is untested today; M2's characterization tests are net-new coverage, not duplication.

## §D Related surfaces (context, not in scope)

- **t461 boundary**: `installPreCommitHookOptional` (the `.git/hooks/pre-commit` direct-write path) is a DIFFERENT delivery axis owned by card t461, in run on another lane. This SPEC's evidence does not touch it; AC-UHD-012 enforces the boundary mechanically.
- **SPEC-UPDATE-YAML-PRESERVE-001**: sibling update-merge preservation work (YAML comments/order for `.moai/config/`) — established that the merge path's lossy-map intermediate is a known hazard family; this SPEC is the JSON/settings member of that family.
- **SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001**: base-snapshot mechanics of the update merge — the mechanism that produces the base `pruneToShared` narrows.
- **t216 investigation report**: NOT accessible from this tree (lives in another lane's worktree); not cited as a source. This SPEC's evidence stands entirely on §A-§C above.
- **CLAUDE.local.md §2.3**: `CleanMoaiManagedPaths` wipes `.moai/state/` and `.moai/config` wholesale on update — binding constraint on any sidecar-state design (design.md §B scheme A3).

## §G Unresolved at plan-phase

- Claude Code's own settings.json rewrite durability (bears on in-file bookkeeping, design.md §F Q2) — unmeasured; a run-phase measurement task if Option A with in-file state is selected.
- golangci-lint baseline on this tree — deferred to M1 pre-flight (plan.md §C).
