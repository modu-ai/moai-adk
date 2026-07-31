# SPEC-CONFIG-TIER-PERSIST-001 — Implementation Plan

## §A Context

Epic SPEC 3 of 6 from a four-lens audit of `moai update` and `.moai/config`. This SPEC owns the
**config resolution and persistence layer**: which tier wins, which writers preserve what, and
whether a write is atomic. It does not own the YAML merge engine internals.

Baseline tree: HEAD `d5336214e`, branch `plan/epic-update-config-audit` (merged with `origin/main`). Every `file:line` in `spec.md` was re-verified
against this tree; drift is recorded in §B.

Two adjacent packages are in scope:

- `internal/config/` — `merge.go`, `source.go`, `loader.go`, `manager.go`, `resolver.go`
- `internal/cli/update/` — `backup/restore.go`, `merge/merge.go`

Plus `internal/permission/resolver.go`, which carries its own literal tier-order slice and is
therefore a first-class edit site for REQ-CTP-006 rather than a downstream consumer, and two
single-line touches: `internal/cli/harness.go` and `internal/core/project/initializer.go`.

The repository `.gitignore` is **not** in scope: its `.moai/config/local/` entry landed at
`b9fc75016`, before the baseline (see D-2).

## §B Known Issues and Verified Drift

Every finding in the audit brief was re-verified. Three references drifted.

Every citation below was **re-measured** during the plan-audit revision. The v0.1.0 table contained
one entry that "corrected" a correct citation into an incorrect one; it is reverted here.

| Ref as briefed | Actual | Note |
|---|---|---|
| `internal/config/manager.go:420-438` (`atomicWrite`) | `:420-438` | **Briefed correctly.** `func` at `:420`, last statement (`return os.Rename`) at `:437`, closing brace at `:438`. v0.1.0 recorded `:420-437` by counting to the last statement rather than the closing brace; both spans name the same function, and the brace-inclusive form is used here for consistency with the other entries. |
| `merge.go:200-224` (`isZero`) | `:200-224` | **Briefed correctly — v0.1.0's "correction" to `:196-220` was wrong and is reverted.** Measured: `grep -n 'func isZero' internal/config/merge.go` → `200:func isZero(v any) bool {`, closing brace at `:224`. The `:196-220` span would have included the tail of `MergeAll` and cut off `isZero`'s `default:` branch. |
| `merge.go:74` (`Priority()`) | `internal/config/source.go:73` | **Wrong file.** `Priority()` lives in `source.go`, not `merge.go`. Confirmed. |
| `merge.go:144` (`valueExists` computed) | `internal/config/merge.go:143` | `:143` is `value, valueExists := tierData[key]`; `:144` is the `if !valueExists {` that consumes it. |
| `initializer.go:475` (`writeReportConfig` writes with `defs.FilePerm`) | func at `:468`, `os.WriteFile` at `:476` | `:475` is the `reportPath :=` assignment. Cite `:476` for the write. |
| `yamlpatch.go:181-209` (`atomicWrite`) | `:182-211` | `func` at `:182`, closing brace at `:211`. The `os.Chmod` citation at `:202` is exact. |
| `restore.go:121-135` / `:139-145` | `:121-134` / `:139-145` | Off-by-one on the 3-way block; the asymmetry is exactly as described. |
| `.moai/config/sections/report.yaml` at 0600 | **now 0644** in both trees | See below, and see `spec.md` §A on why no section file reads 0600 in this worktree. |

All confirmed exactly as briefed: `internal/defs/perms.go:11`,
`internal/settings/yamlpatch/yamlpatch.go:202`, `internal/config/manager.go:161` (`@MX:REASON`),
`internal/config/merge.go:149-152` / `:155-162`, `internal/config/source.go:26-28`,
zero grep matches for `config/local|localDir|SrcLocal` in `internal/config/loader.go`,
`internal/config/loader.go:121-131`, `internal/cli/update/merge/merge.go:41-85`,
`internal/cli/update/backup/restore.go:105,128,142,145`,
`internal/core/project/initializer.go:371-450`, `internal/cli/harness.go:390`.

### Ordering sites re-measured (the basis for REQ-CTP-006's rescope)

```
$ grep -rn '\.Priority()' --include='*.go' internal/ | grep -v '_test.go'
(no output, exit 1)
```

`Priority()` is read only by `internal/config/source_test.go:49,98,99`. The three sites that actually
express tier order are `internal/config/source.go` `iota` (`:16-44`), `AllSources()` (`:103-114`), and
`internal/permission/resolver.go` `tiers` (`:225-234`). `internal/config/merge.go:137` walks
`AllSources()`. See `spec.md` §A and REQ-CTP-006.

### The `report.yaml` open question — resolved

The brief could not attribute `report.yaml`'s 0600 mode to a known writer and flagged it as an open
question. It is resolved, in two parts.

First, the observation itself drifted **during this authoring session**. An early `ls -la` showed
`-rw-------` for `report.yaml`; a `stat` a few minutes later showed `-rw-r--r--`, with the mtime
unchanged at `Jul 27 17:31`. An unchanged mtime with a changed mode is a `chmod`, not a rewrite —
some concurrent actor in that checkout widened it. The narrowed set is therefore now four
files, not five — **and it is observable only in the primary checkout**
(`/Users/goos/MoAI/moai-adk-go`), not in this worktree:

```
$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/config/sections/*.yaml | awk '{print $1, $NF}'
-rw-------@ .../git-convention.yaml
-rw-------@ .../git-strategy.yaml
-rw-------@ .../language.yaml
-rw-------@ .../user.yaml
(28 others at -rw-r--r--)

$ ls -la .moai/config/sections/*.yaml | awk '{print $1, $NF}'   # this worktree
(all 32 at -rw-r--r--)
```

A `git worktree add` materialises files at the checkout umask — git records only the executable bit —
so a working-tree `chmod` cannot cross into a worktree. The four narrowed files are a property of the
tree in which `Save()` ran. `spec.md` §A carries the consequence for AC design; the short form is
that no AC may assert on this repository's own section-file modes.

Second, and more usefully, all four remaining files are `ConfigManager.Save()`'s `saveSection`
targets, and `report.yaml` is not. `Save()` writes exactly six sections — `user`, `language`,
`quality`, `git-convention`, `git-strategy`, `llm`. The two absent from the narrowed set,
`quality.yaml` and `llm.yaml`, carry later mtimes (`Jul 27 20:09`, `Jul 29 19:06`) than the 0600
cluster, consistent with a subsequent `moai update` restore rewriting them at `defs.FilePerm` and
widening them back.

`report.yaml` has exactly two writers, and neither can narrow:
`initializer.writeReportConfig` (`internal/core/project/initializer.go:468`, writing at `:476`) uses
`defs.FilePerm`, and the settings seam routes through `yamlpatch.atomicWrite`
(`internal/settings/yamlpatch/yamlpatch.go:182-211`), which `os.Stat`s the target and `os.Chmod`s
the temp file to the target's mode before renaming. `internal/config/audit_registry.go:76` confirms
`report` is "settings-seam only — not in the Loader.Load chain".

Conclusion: F1's mechanism is fully attributed to `ConfigManager.atomicWrite`. `report.yaml` was
never evidence for it. REQ-CTP-025's migration still covers `report.yaml` because it is scoped by
directory, not by writer.

### Reproduction probes

All three probes below were run against `d5336214e` and are the regression fixtures for §F.

- **F2/F3** — a temporary `internal/config/zzz_probe_test.go` calling `MergeAll` directly produced
  the `source=builtin` and `from-project` outputs quoted in `spec.md` §A. Removed after running.
- **F7** — a temporary `internal/cli/update/merge/zzz_probe_test.go` driving two successive
  `MergeGitignoreFile` calls with a template that drops `DS_Store` produced the migration and the
  three surviving `build/` duplicates quoted in `spec.md` §A. Removed after running.
- **F1** — two separate observations, deliberately kept apart. (a) The *mechanism*: a probe against
  `d5336214e` observed a pre-existing `-rw-r--r--` target become `-rw-------` after one `atomicWrite`
  round trip — this reproduces anywhere and is the regression fixture for AC-CTP-020. (b) The
  *field evidence*: `ls -la` plus `git ls-files -s` over the **primary checkout's**
  `.moai/config/sections/`, showing four files at `-rw-------` against a git index recording `100644`
  for every one. Observation (b) does **not** reproduce in this worktree (see above) and is therefore
  not an AC baseline.

## §C Pre-flight

Before M1:

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — this checkout
   is shared and demonstrably concurrent (see the `report.yaml` chmod above). Do not begin without
   a clean divergence read.
2. `go test ./internal/config/... ./internal/cli/update/... -count=1` — record the green baseline.
3. `golangci-lint run --timeout=2m` — record the lint baseline.
4. Confirm `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) status, since it is `depends_on`. Two distinct
   checks ride on this step:
   - **Gate.** The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
     `status: completed`. E2 is `draft`, as is every SPEC in this Epic, so entering `/moai run` on
     this SPEC today raises the 3-option wait / override / abort blocker. **The dependency is
     satisfied by sequencing, not by `--ignore-deps`** — the Epic run order in `progress.md` §E.1 is
     the mechanism, and a run-phase agent may not set that flag on its own. If E2 slips and starting
     this SPEC early becomes necessary, the correct move is to run M1-M3 and M5-M6 (which carry no
     E2 edge) and hold M4 open until E2 closes; that is an orchestrator scope decision surfaced via
     `AskUserQuestion`, not a flag.
   - **Content.** If E2 has landed its shared-write work, M4 below becomes a consumer rather than an
     author — check before writing.
5. Re-measure the three ordering sites before M2, because REQ-CTP-006's scope depends on them:
   `grep -rn '\.Priority()' --include='*.go' internal/ | grep -v '_test.go'` (expect no output),
   and confirm `AllSources()` and `internal/permission/resolver.go`'s `tiers` both still list
   `SrcProject` before `SrcLocal`.

## §D Constraints and Decision Records

### D-1 — `SrcLocal` outranks `SrcProject` (REQ-CTP-006): reorder the enum, do not correct the doc

**Decision: reorder.**

The doc comment at `source.go:26-28` is not aspirational prose; it names two concrete file
locations — `.claude/settings.local.json` and `.moai/config/local/*.yaml` — and both are, by their
own governing doctrine, override surfaces. `CLAUDE.local.md` §22 records that
`.claude/settings.local.json` is Claude Code's own highest-priority settings scope, above project
settings; `CLAUDE.local.md` §22.9 records `.moai/config/local/workflow.yaml` as the maintainer
opt-in mechanism. An override tier that loses to the tier it is supposed to override is not a
naming problem, it is a broken tier.

Correcting the documentation instead would encode the defect as intent and leave
`.moai/config/local/` with no mechanism to do the one thing it exists to do. It would also make
REQ-CTP-010 pointless: wiring a local tier into the `Loader` that cannot outrank the sections
directory delivers nothing.

**What "reorder the enum" actually means — the v0.1.0 scope was wrong.** This plan originally said
"move `SrcLocal` above `SrcProject` in the `iota` block" and treated that as the whole change. It is
not, and executing it alone would have produced a change that fixes nothing and breaks a test.
Measured on the baseline tree:

```
$ grep -rn '\.Priority()' --include='*.go' internal/ | grep -v '_test.go'
(no output, exit 1)
```

`Priority()` (`source.go:73`) returns `int(s)` and has **zero non-test consumers**. Tier order is
carried instead by two hand-written literal slices — `AllSources()` (`source.go:103-114`), walked by
`MergeAll` at `merge.go:137`, and `tiers` inside `internal/permission/resolver.go:225-234`. Both are
symbolic (`SrcPolicy, SrcUser, SrcProject, SrcLocal, …`), so renumbering the `iota` block does not
move them.

The consequences of the three candidate scopes, traced through the two existing guards
(`source_test.go:89-102 TestAllSources`, asserting `AllSources()[i].Priority() == i`; and
`:104-119 TestSourceOrdering`, asserting `AllSources()` against a symbolic literal):

| Scope | `MergeAll` order | Permission order | `TestAllSources` | `TestSourceOrdering` |
|---|---|---|---|---|
| `iota` only (v0.1.0 plan) | unchanged — AC-CTP-005 **fails** | unchanged — AC-CTP-007 **fails** | **breaks** (`[2]` is `SrcProject`, `Priority()`=3) | passes (symbolic) |
| `iota` + `AllSources()` | fixed | unchanged — AC-CTP-007 **fails** | passes (invariant restored) | **fails — literal needs updating** |
| all three (REQ-CTP-006) | fixed | fixed | passes, no edit needed | **fails — literal needs updating** |

So the deliverable is: reorder all three sites in one commit, and update **`TestSourceOrdering`'s
`expectedOrder` literal only**. `TestAllSources` needs no edit, because reordering the `iota` block
and `AllSources()` together preserves `AllSources()[i].Priority() == i` — the plan-audit's suggested
remedy ("update both existing tests") over-corrects on that point.

**Against the enum's fan-in.** `source.go:5` records `fan_in=71`. That number argues for care, not
for inaction — and the measurement above reframes the risk: the realistic failure mode is
**under-reach** (the edit never touches the resolution path), not over-reach. REQ-CTP-006 names the
three sites so under-reach is not possible by omission; AC-CTP-005 and AC-CTP-007 assert merge and
permission *outcomes* rather than ordinals, so a partial reorder fails them; NFR-CTP-005 keeps it a
standalone revertable commit.

The remaining hazard is any code that persists a `Source` as an integer. That must be grepped for
before landing, and is a milestone task, not an assumption — see AC-CTP-008, whose measured baseline
is six matches, all safe.

**Explicitly not decided here:** `SrcUser` currently outranks `SrcProject`, which is the opposite of
what most project-scoped tools do. No lens produced evidence of harm, so `spec.md` §C excludes it.
Re-opening that pair would double the blast radius for no evidenced gain.

### D-2 — `.moai/config/local/` (REQ-CTP-010): wire it into the `Loader`, ordered after the zero-value fix

**Decision: wire it, and only after M1.**

`CLAUDE.local.md` §22.9 documents the branch-guard opt-in as creating
`.moai/config/local/workflow.yaml`. The branch guard reads `Workflow.BranchGuard.Enabled` through
`hook.ConfigProvider.Get()` → `config.Loader.Load(configDir)`, and the `Loader` has no local-tier
code path at all. The documented instruction is inert. The alternative — correcting the doctrine to
say "there is no typed local override" — leaves maintainers of shared checkouts with no way to
enable a guard that was deliberately shipped default-off *for them* to opt into. That is a worse
outcome than the wiring cost.

**The ordering is the load-bearing part.** REQ-CTP-012 requires M1 (zero-value semantics) to land
first. Wiring the local tier while `isZero` still skips falsey values produces a directory in which
`enabled: true` works and `enabled: false` does nothing — an opt-in with no opt-out. That asymmetry
is worse than the current state, because the current state is uniformly inert and therefore
uniformly diagnosable.

**The gitignore coupling — already closed before this SPEC was drafted.** v0.1.0 of this plan stated
that `.moai/config/local/` is not gitignored and that `CLAUDE.local.md` §22.9 still carries a
`BLOCKER` note, and derived two M2 work items from that premise. Both halves of the premise are
false on the baseline tree:

```
$ git check-ignore -v .moai/config/local/workflow.yaml
.gitignore:183:.moai/config/local/	.moai/config/local/workflow.yaml     (exit 0)

$ grep -c 'BLOCKER (gitignore)' CLAUDE.local.md
0

$ git merge-base --is-ancestor b9fc75016 d5336214e; echo $?
0
```

The entry landed at `b9fc75016`, an ancestor of the baseline, and §22.9 already reads
"gitignore (해결됨)". The hazard M2 was supposed to close is closed. REQ-CTP-013 and REQ-CTP-014
therefore carry **no work item**; they survive only as regression guards (AC-CTP-012, AC-CTP-013)
that fail if the entry or the corrected note is later removed. Leaving them as work items would let
run-phase claim two already-green ACs as evidence of progress and dilute the remaining verification —
which is precisely why they are demoted here rather than silently kept.

What still holds from the original reasoning is the *coupling direction*: M2 is what makes the
directory load-bearing, so if the gitignore entry ever regressed, M2 would be the milestone that
turns the regression into a live hazard. That is what AC-CTP-012 guards.

### D-3 — malformed configuration (REQ-CTP-015 through REQ-CTP-020): asymmetric by caller, plus a save guard

**Decision: absent stays silent; malformed is a CLI error and a hook advisory; and `Save()` refuses
to persist a section that failed to load.**

Three distinct states are currently collapsed into one. A section file that is *absent* is a normal
greenfield condition. A section file that is *malformed* is a defect. A section that loaded
*cleanly* is the happy path. All fourteen loaders in `internal/config/loader.go` treat the first two
identically — `slog.Warn(..., "using defaults")` and continue.

*Absent stays silent* because a fresh project legitimately has no `sections/` entries yet, and
`loader.go:42` already warns once for the missing directory.

*Malformed is a CLI error* because a CLI invocation is a foreground, operator-attended action with
an exit code; silently substituting defaults there produces a command that reports success while
having read none of the operator's configuration.

*Malformed is a hook advisory, not a hook failure*, because hooks are fail-open by established
norm. `.claude/rules/moai/workflow/main-checkout-branch-guard.md` states the branch guard "fires
ONLY on positive evidence" and that any uncertainty "falls through to allow, writes an advisory to
stderr". A hook that aborts a session over a config parse error is a worse failure than one that
proceeds on defaults. What must change is visibility: `slog.Warn` is not operator-visible in a hook
context, so REQ-CTP-018 adds a stderr advisory alongside it.

**The save guard is the actual data-loss fix.** REQ-CTP-019 exists because F5 and F6 compose into
something neither is alone. A process reads a section file that `moai update` has truncated
mid-write, swallows the parse error, holds compiled defaults in `cfg`, and then calls
`ConfigManager.Save()` — which serialises those defaults and writes them over the user's real file.
The user's configuration is gone, permanently, with no error anywhere. Fixing atomicity (M4) closes
the window that produces the truncated read; the save guard closes the amplification. Both are
needed, because a malformed file can also arrive by hand-editing, by a merge conflict marker, or by
disk corruption — none of which M4 addresses.

REQ-CTP-020 scopes the refusal to the failed section alone, so one bad file does not block
persisting five good ones.

### D-4 — the shared write helper (REQ-CTP-021 through REQ-CTP-024): boundary with `SPEC-UPDATE-DATA-SURVIVAL-001`

E2 owns *that the bytes reach disk before a destructive step* — backup coverage, the recovery
manifest, the restore entry point. It explicitly does not specify a write mechanism. Its §C states
"Merge-tier semantics, dead `.moai/config` keys, CI gate additions, and documentation drift" belong
to Epic SPECs 3 through 6, and its YAML exclusion says it "cares that the bytes reach disk before
deletion, not how they are later merged".

This SPEC owns *how the write itself behaves*: atomic, and mode-preserving. The helper introduced
in M4 is the mechanism E2's writes will use. Concretely, M4 must not duplicate an atomic-write
helper if E2 has already landed one — §C pre-flight step 4 checks this, and the M4 task list starts
with "consume or author", not "author".

The implementation pattern is not novel: `internal/settings/yamlpatch/yamlpatch.go:182-211` already
does exactly this. The helper should be lifted from that shape, with one addition — `yamlpatch`
`os.Stat`s the target and errors when it is absent, which is correct for a patch operation but
wrong for a create. REQ-CTP-022 requires the shared helper to fall back to `defs.FilePerm` on a
missing target rather than erroring.

**Why the migration (REQ-CTP-025) is separate from the fix.** A `Stat`-based preservation fix reads
the *current* mode and preserves it. Applied to a file already at 0600, it preserves 0600 forever.
The four narrowed files in this repository — and every equivalent file in every user project that
has ever run a `Save()` — stay narrow unless something widens them. REQ-CTP-026 bounds the
migration to widening-toward-`defs.FilePerm`, inside `.moai/config/` only.

### D-5 — `.gitignore` merge (REQ-CTP-028 through REQ-CTP-032): parse the header, keep a fallback

The current heuristic — "any backup line absent from the new template is user-authored" — is
correct only when the template never removes a pattern. When it does, the removed pattern is
reclassified as user-authored on the next update and becomes permanent. The header the merger
itself writes is the missing information: everything below `# User Custom Patterns` was user
content on the previous pass, and everything above it was template content.

REQ-CTP-032 keeps the old heuristic as a fallback for backups written before the header convention,
because a header-only parse would classify an entire pre-header backup as template content and
discard every user pattern in it. That is a data-loss regression introduced by a data-loss fix, so
the fallback is not optional.

### D-6 — fallback visibility (REQ-CTP-033 through REQ-CTP-035)

The asymmetry is inverted: the *more* serious failure is the *quieter* one. A 3-way failure means
the merge lost provenance and fell back to a weaker algorithm; a 2-way failure means the merge gave
up entirely and restored the backup. The first is silent until three consecutive occurrences; the
second prints immediately.

A third path is quieter than both. When `os.ReadFile(basePath)` errors — the base file is absent —
control falls to 2-way with neither a `recordFallback` call nor a warning. REQ-CTP-034 separates
that case, because "the base is missing" and "the merge failed" have different causes and different
fixes.

REQ-CTP-035 preserves the noise-suppression ledger's purpose. The ledger exists to stop *repetition*
becoming noise; suppressing a *first* occurrence is not noise suppression, it is concealment.

R9 accepts the consequence: while `SPEC-UPDATE-YAML-PRESERVE-001` is unlanded, `quality.yaml` takes
this path on every update, so operators will start seeing an advisory they did not see before. That
is correct. They have been silently receiving 2-way merges while believing they received 3-way.

**This is a reversal of an implemented sibling requirement, and v0.1.0 did not say so.** The silence
REQ-CTP-033 removes is not an accident of the code — it is `REQ-UN-007` of
`SPEC-V3R6-UPDATE-NOISE-001` (`status: implemented`), realised as
`const fallbackAdvisoryThreshold = 3` (`internal/cli/update_noise.go:37`) and cited by name in the
call site's own comment (`// REQ-UN-007/008/010: …` immediately above the `recordFallback` call in
`restore.go`). v0.1.0 named neither the SPEC nor the requirement anywhere in this directory. The
reversal is now declared in `spec.md` §K, which also records the rejected alternative
(`REQ-UN-010`'s `--verbose` escape hatch) and the `REQ-UN-*` clauses that survive intact — the
counter, the 3-strike advisory, the success reset, and `--verbose` itself. M6 must not land without
that section, and if the reversal is declined at the kickoff gate, REQ-CTP-033/AC-CTP-031 are
withdrawn together while REQ-CTP-034 (absent-base advisory) proceeds — it conflicts with nothing.

## §E Self-Verification

Before declaring any milestone complete:

- The milestone's ACs in `acceptance.md` each ran, and each `-run` AC's output carried its
  `--- PASS: <exact name>` line. An `ok ... [no tests to run]` is a failure.
- Each new guard's falsification in `acceptance.md` §C ran and produced the expected `--- FAIL`.
- `go test ./... -count=1` is green.
- `golangci-lint run --timeout=2m` is clean.
- `git diff --stat` touches no path under `internal/template/templates/`.
- No test file references `t.Setenv` with an `OTEL_` variable.

## §F Milestones

Ordered by decision-reversibility: the decisions most likely to change sit first, so review
attention lands on them while they are still cheap to revise. M1 through M3 are behavioural changes
to load-bearing semantics; M4 through M6 are progressively more mechanical.

### M1 — Tier merge semantics (REQ-CTP-001 through REQ-CTP-005b) — Priority High

The highest-blast-radius change in the SPEC and the one most likely to be revised under review.

1. **Enumerate the blast radius before writing code.** Grep every shipped
   `.moai/config/sections/*.yaml` and `internal/config/defaults.go` for falsey values — `false`,
   `0`, `""`, `[]`, `{}` — and produce a table of every key whose effective value changes once
   falsey values start winning. This table is a deliverable, not a footnote; R1 and R2 are
   mitigated by it or not at all.
2. Split "key present" from "value is zero" in the `MergeAll` tier walk. `valueExists` is already
   computed at `merge.go:143`; the fix is to stop discarding it at `:149-152`.
3. Verify the `PolicyOverrideRejected` path at `:155-162` becomes reachable for a falsey policy
   value — this is the security-relevant half of the change and needs its own assertion.
4. Resolve `isZero`'s fate per REQ-CTP-005a/b: keep it together with any caller that genuinely needs
   zero-detection; remove it if none remains. Record the decision and its caller count — AC-CTP-037
   asserts the two states are consistent, so an unreferenced surviving helper fails.
5. Land as a single revertable commit (NFR-CTP-005).

### M2 — Tier ordering and local-tier reachability (REQ-CTP-006 through REQ-CTP-014) — Priority High

Blocked by M1 per REQ-CTP-012.

1. Grep for any persistence of a `Source` as an integer — a stored ordinal would silently change
   meaning. This is the one unmitigated hazard in D-1 and must be checked before the reorder
   (AC-CTP-008; baseline is six matches, all safe).
2. **Reorder all three ordering sites in one commit** (REQ-CTP-006): the `iota` block in
   `internal/config/source.go`, the `AllSources()` literal at `source.go:103-114`, and the `tiers`
   literal at `internal/permission/resolver.go:225-234`. Reordering only the `iota` block changes no
   resolution and breaks `TestAllSources` — see D-1 for the traced comparison.
3. Update `TestSourceOrdering`'s `expectedOrder` literal (`source_test.go:104-119`) to the new order
   in the same commit and keep it (REQ-CTP-008). Do **not** author a duplicate ordering guard, and do
   **not** replace the literal with an enum-derived sequence. `TestAllSources` (`:89-102`) needs no
   edit.
4. Assert `internal/permission/resolver.go` resolution under the new order (REQ-CTP-009). Its `tiers`
   slice is independent of the enum, so this assertion fails if step 2 reordered only two sites.
5. Add local-tier reading to `internal/config/loader.go`, applied above `sections/`
   (REQ-CTP-010/011).

> Steps 6 and 7 of the v0.1.0 plan — adding `.moai/config/local/` to `.gitignore` and clearing the
> `CLAUDE.local.md` §22.9 `BLOCKER` note — are **removed**. Both landed at `b9fc75016`, an ancestor of
> the baseline; see D-2 for the measurements. REQ-CTP-013/014 survive as regression guards only.

### M3 — Malformed-configuration contract (REQ-CTP-015 through REQ-CTP-020) — Priority High

1. Separate absent from malformed across all fourteen loaders in `internal/config/loader.go`.
2. Record failed sections on the `Loader` alongside `loadedSections`.
3. CLI callers surface a returned error (REQ-CTP-017); hook callers emit a stderr advisory and
   continue (REQ-CTP-018).
4. Add the `ConfigManager.Save` refusal for failed sections (REQ-CTP-019/020) — the data-loss fix.

### M4 — Atomic, mode-preserving writes (REQ-CTP-021 through REQ-CTP-027) — Priority High

1. **First**: check whether `SPEC-UPDATE-DATA-SURVIVAL-001` has landed a shared write helper. If
   so, consume it and add mode preservation there rather than authoring a second one.
2. Author or extend the helper, modelled on `yamlpatch.atomicWrite` but with REQ-CTP-022's
   create-path fallback to `defs.FilePerm`.
3. Route all six writers through it (REQ-CTP-023).
4. Replace the `0o644` literal at `internal/cli/harness.go:390` with `defs.FilePerm`
   (REQ-CTP-024).
5. Add the widening migration (REQ-CTP-025/026).
6. Add the bare-`os.WriteFile` guard (REQ-CTP-027).

### M5 — `.gitignore` merge correctness (REQ-CTP-028 through REQ-CTP-032) — Priority Medium

Mechanical relative to M1-M4; the header parse and the dedupe are local to one function.

1. Parse `# User Custom Patterns` as a section boundary.
2. Dedupe user patterns, first occurrence wins.
3. Keep the not-in-template heuristic as the pre-header fallback (REQ-CTP-032).
4. Assert idempotence (REQ-CTP-031).

### M6 — Fallback visibility (REQ-CTP-033 through REQ-CTP-035) — Priority Medium

The smallest change in the SPEC by line count, and the largest by cross-SPEC consequence: it reverses
`REQ-UN-007` of the implemented `SPEC-V3R6-UPDATE-NOISE-001`.

0. **Gate.** Confirm `spec.md` §K is present and that the reversal was accepted at Implementation
   Kickoff Approval. If it was declined, drop step 1 and AC-CTP-031 and land steps 2-3 only.
1. Emit a first-occurrence advisory on 3-way merge failure. The commit body must state that it
   supersedes `REQ-UN-007`'s silent-first-occurrence clause and cite §K.
2. Emit a distinguishable advisory when the base file is absent. This conflicts with nothing — the
   path today emits neither a `recordFallback` call nor a warning.
3. Confirm the ledger still governs repetition rather than first occurrence, and that the
   `fallback_count` increment (`REQ-UN-007`), the 3-strike advisory (`REQ-UN-008`), the success reset
   (`REQ-UN-009`), and `--verbose` (`REQ-UN-010`) all still behave as `SPEC-V3R6-UPDATE-NOISE-001`
   specifies. Only the silent-first-occurrence clause is superseded.
4. Do **not** edit `SPEC-V3R6-UPDATE-NOISE-001`'s own artifacts (§K.3).

## §G Anti-Patterns

- **Deleting `isZero` reflexively.** It may have a legitimate caller elsewhere. REQ-CTP-005a/b
  require a decision conditioned on caller survival, not a deletion.
- **Reordering only the `iota` block.** It changes no resolution — `Priority()` has zero non-test
  consumers — and breaks `TestAllSources`. All three sites, one commit (D-1).
- **Authoring a second ordering guard.** `TestSourceOrdering` already pins the literal sequence and
  is green; the obligation is to update its `expectedOrder`, not to duplicate it (REQ-CTP-008).
- **Treating REQ-CTP-013/014 as work.** Both landed at `b9fc75016`. Claiming their ACs as evidence of
  M2 progress dilutes the milestone's real verification (D-2).
- **Landing REQ-CTP-033 without `spec.md` §K.** It silently reverses an implemented sibling
  requirement (D-6).
- **Asserting on this repository's own section-file modes.** They read 0644 in every worktree; such
  an AC passes trivially and proves nothing (`spec.md` §A).
- **Reordering the enum without grepping for persisted ordinals.** M2 step 1 exists for this.
- **Landing M2 before M1.** Produces an opt-in with no opt-out (R4).
- **Fixing mode preservation without the migration.** A `Stat`-based fix preserves 0600 forever on
  every already-narrowed file (D-4).
- **Parsing the `.gitignore` header without the pre-header fallback.** Discards every user pattern
  in an old backup (R8).
- **Suppressing the new 3-way advisory because it is noisy.** The noise is `quality.yaml` failing
  3-way on every update. That is the signal (R9).
- **Using `git stash` for falsification.** This checkout is shared and demonstrably concurrent —
  the `report.yaml` chmod during authoring is the proof. Use `go test -overlay` or a scratch
  worktree driven by `go -C`.
- **Writing fixtures anywhere but `t.TempDir()`.** NFR-CTP-001.
- **Editing `internal/template/templates/**`.** NFR-CTP-004.

## §H Cross-References

- `spec.md` §I — amendment requests to `SPEC-UPDATE-YAML-PRESERVE-001`.
- `spec.md` §J — evidence carried forward for `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`.
- `spec.md` §K — the declared reversal of `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007` (D-6, M6).
- `progress.md` §E.1 — the Epic run order that satisfies `depends_on` by sequencing (§C step 4).
- `acceptance.md` §C — the falsification procedures.
- `.moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/plan.md` — the E2 boundary referenced in D-4.
- `CLAUDE.local.md` §22.9 — the opt-in doctrine amended by REQ-CTP-014.
