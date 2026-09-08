# t533 — mutant evidence for the AC-CGM-001 read-only guard

Card t533 · SPEC-CODEX-GHOST-SKILLS-MEASURE-001 · worktree `.claude/worktrees/t533` · HEAD `6a46c0edbe2dec6014c685184aa2cf9dc346cc59` · tree `4b07b9eb68e98f6ef8a2a861e854431a330cc27d`

All three mutants ran in an **isolated directory outside `~/.codex`**, created and destroyed inside this run. Zero writes reached `~/.codex`; the residue check after the suite returned `0` and `config.toml` mtime stayed `1788771295`.

Isolated path used: `…/scratchpad/t533-mutant/dir` (session scratchpad, removed at the end).

## Why three mutants and not one

The guard AC-CGM-001 declares has **two probes**, and `Then` requires both to be silent:

| Probe | Selector | Catches |
|---|---|---|
| P1 entry list | `/bin/ls -A ~/.codex \| sort` | a name appearing or disappearing |
| P2 mtime | `find ~/.codex -maxdepth 1 -mindepth 1 -exec stat -f '%N %m' {} +` | a file modified **in place** (name unchanged) |

Plan-audit round 3 (N5) found that only P1 had ever been exercised — and P2 is the only probe that catches prune's actual write shape (`os.WriteFile(cfgPath, pruned, 0o600)`, `codex_skills_prune.go:210`), which rewrites `config.toml` under its existing name. An unexercised absence-as-PASS probe is this card's own `plan.md` D-2 violation.

## Mutant 1 — a new entry appears

```
touch "$D/dir/.mutant"
```

| Probe | diff output | rc |
|---|---|---|
| P1 entry | `0a1` / `> .mutant` | 1 (RED — caught) |
| P2 mtime | `0a1` / `> …/dir/.mutant 1788790028` | 1 (RED — caught) |

Both probes catch a new entry. This is the mutant the plan already carried.

## Mutant 2 — in-place modification (the N5 debt)

Baseline mtimes were pinned into the past with `touch -t 202601010000` so the change is provably distinct rather than dependent on wall-clock timing.

```
printf 'in-place content change\n' >> "$D/dir/a"     # name unchanged
```

Observed: `a` moved `1767193200` → `1788790185`.

| Probe | diff output | rc | reading |
|---|---|---|---|
| P1 entry | *(none)* | 0 | **blind, as predicted** — the name set did not change |
| P2 mtime | `1c1` `< …/dir/a 1767193200` / `> …/dir/a 1788790185` | 1 | **RED — caught** |

**N5 is closed.** P2 turns red on exactly the event P1 cannot see, so the two probes are proven to catch different events and the `Then`-requires-both wording is now justified rather than assumed.

## Mutant 3 — a `..`-prefixed name (the N6 debt)

```
touch "$D/dir/..codex-global-state.json.tmp-t533probe"
```

| Selector | diff output | rc | reading |
|---|---|---|---|
| legacy glob pair `DIR/* DIR/.[!.]*` | *(none)* | 0 | **BLIND — the measured hole** |
| replacement `find -maxdepth 1 -mindepth 1` | `0a1` `> …/..codex-global-state.json.tmp-t533probe …` | 1 | **RED — caught** |
| entry probe `/bin/ls -A` | lists it (`grep -c` → 1) | — | sees it |

**This is an uncaught-mutant record for the selector the AC currently names**, kept rather than deleted: it draws the legacy selector's boundary. The hole is real on the live tree too — see below.

## The N6 hole, measured on the real `~/.codex` (read-only)

```
/bin/ls -A ~/.codex | wc -l                                        → 72
stat -f '%N %m' ~/.codex/* ~/.codex/.[!.]* | wc -l                 → 70   (legacy)
find ~/.codex -maxdepth 1 -mindepth 1 -exec stat -f '%N %m' {} +   → 72   (replacement)
```

The two entries invisible to the legacy glob pair:

```
..codex-global-state.json.tmp-1780394894609-a08f5abb-b93c-462a-996a-3885457c3e0e
..codex-global-state.json.tmp-1786321411778-e43bc617-aa1d-48d9-92d3-9c43b4597095
```

Neither glob can match a name beginning with `..`: `DIR/*` excludes dot-names, and `DIR/.[!.]*` requires the second character not to be a dot. The baseline snapshots in this directory therefore use the `find` form and hold **72** rows.

**Impact stated without inflation.** The threat this card actually guards against is still caught either way: a write by the layer-2 verb creates a new backup file (P1 catches it) and rewrites `config.toml`, whose name carries no dot at all and is inside the legacy glob. The defect is a mismatch between the AC's declared scope ("everything under `~/.codex`") and the selector's real coverage — not a live escape route for the write under discussion.

## A second, independent defect in the legacy selector, found while reproducing it

Under **zsh**, a glob with no matches aborts the whole command line. In the isolated directory, before any dot-file existed:

```
stat -f '%N %m' "$D/dir"/* "$D/dir"/.[!.]* 2>/dev/null | sort > glob.before
→ zsh: no matches found: …/dir/.[!.]*
→ glob.before: 0 lines
```

`2>/dev/null` does not suppress it — the failure happens in the shell before `stat` is executed, and the redirection still creates an **empty** file. In an absence-as-PASS check an empty selector output is indistinguishable from "nothing changed". The `find` replacement has no such failure mode. (Reproducing the legacy selector for mutant 3 required `setopt nonomatch`.)

## Residual risk this evidence does **not** remove

- `stat -f '%m'` is **whole-second**. An in-place write landing in the same second as the baseline capture would leave P2 silent. Measured: `%Fm` on the same file yields `1788790185.324442879`, so a nanosecond-resolution probe is available. This run therefore also captured `codex-mtimes-frac.before` with `%Fm` alongside the AC-named `%m` artifact.
- The mutants prove the selectors' **discriminating power**, not that no write occurred. That claim rests on the `before`/`after` diffs, recorded separately.
