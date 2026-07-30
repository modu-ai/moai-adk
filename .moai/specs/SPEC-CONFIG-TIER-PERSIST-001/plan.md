# SPEC-CONFIG-TIER-PERSIST-001 — Implementation Plan

## §A Context

Epic SPEC 3 of 6 from a four-lens audit of `moai update` and `.moai/config`. This SPEC owns the
**config resolution and persistence layer**: which tier wins, which writers preserve what, and
whether a write is atomic. It does not own the YAML merge engine internals.

Baseline tree: HEAD `1d4e4f7da`, branch `main`. Every `file:line` in `spec.md` was re-verified
against this tree; drift is recorded in §B.

Two adjacent packages are in scope:

- `internal/config/` — `merge.go`, `source.go`, `loader.go`, `manager.go`, `resolver.go`
- `internal/cli/update/` — `backup/restore.go`, `merge/merge.go`

Plus three single-line touches: `internal/cli/harness.go`, `internal/core/project/initializer.go`,
and the repository `.gitignore`.

## §B Known Issues and Verified Drift

Every finding in the audit brief was re-verified. Three references drifted.

| Ref as briefed | Actual | Note |
|---|---|---|
| `internal/config/manager.go:420-438` (`atomicWrite`) | `:420-437` | Function starts at `:420`; body ends `:437`. Behaviour as described. |
| `merge.go:200-224` (`isZero`) | `internal/config/merge.go:196-220` | Four-line offset. Behaviour as described. |
| `merge.go:74` (`Priority()`) | `internal/config/source.go:73` | **Wrong file.** `Priority()` lives in `source.go`, not `merge.go`. |
| `restore.go:121-135` / `:139-145` | `:121-134` / `:139-145` | Off-by-one on the 3-way block; the asymmetry is exactly as described. |
| `.moai/config/sections/report.yaml` at 0600 | **now 0644** | See below. |

All confirmed exactly as briefed: `internal/defs/perms.go:11`,
`internal/settings/yamlpatch/yamlpatch.go:202`, `internal/config/manager.go:161` (`@MX:REASON`),
`internal/config/merge.go:144` / `:149-152` / `:155-162`, `internal/config/source.go:26-28`,
zero grep matches for `config/local|localDir|SrcLocal` in `internal/config/loader.go`,
`internal/config/loader.go:121-131`, `internal/cli/update/merge/merge.go:41-85`,
`internal/cli/update/backup/restore.go:105,128,142,145`,
`internal/core/project/initializer.go:371-450`, `internal/cli/harness.go:390`.

### The `report.yaml` open question — resolved

The brief could not attribute `report.yaml`'s 0600 mode to a known writer and flagged it as an open
question. It is resolved, in two parts.

First, the observation itself drifted **during this authoring session**. An early `ls -la` showed
`-rw-------` for `report.yaml`; a `stat` a few minutes later showed `-rw-r--r--`, with the mtime
unchanged at `Jul 27 17:31`. An unchanged mtime with a changed mode is a `chmod`, not a rewrite —
some concurrent actor in this shared checkout widened it. The narrowed set is therefore now four
files, not five:

```
-rw------- .moai/config/sections/git-convention.yaml
-rw------- .moai/config/sections/git-strategy.yaml
-rw------- .moai/config/sections/language.yaml
-rw------- .moai/config/sections/user.yaml
```

Second, and more usefully, all four remaining files are `ConfigManager.Save()`'s `saveSection`
targets, and `report.yaml` is not. `Save()` writes exactly six sections — `user`, `language`,
`quality`, `git-convention`, `git-strategy`, `llm`. The two absent from the narrowed set,
`quality.yaml` and `llm.yaml`, carry later mtimes (`Jul 27 20:09`, `Jul 29 19:06`) than the 0600
cluster, consistent with a subsequent `moai update` restore rewriting them at `defs.FilePerm` and
widening them back.

`report.yaml` has exactly two writers, and neither can narrow:
`initializer.writeReportConfig` (`internal/core/project/initializer.go:475`) writes with
`defs.FilePerm`, and the settings seam routes through `yamlpatch.atomicWrite`
(`internal/settings/yamlpatch/yamlpatch.go:181-209`), which `os.Stat`s the target and `os.Chmod`s
the temp file to the target's mode before renaming. `internal/config/audit_registry.go:76` confirms
`report` is "settings-seam only — not in the Loader.Load chain".

Conclusion: F1's mechanism is fully attributed to `ConfigManager.atomicWrite`. `report.yaml` was
never evidence for it. REQ-CTP-025's migration still covers `report.yaml` because it is scoped by
directory, not by writer.

### Reproduction probes

All three probes below were run against `1d4e4f7da` and are the regression fixtures for §F.

- **F2/F3** — a temporary `internal/config/zzz_probe_test.go` calling `MergeAll` directly produced
  the `source=builtin` and `from-project` outputs quoted in `spec.md` §A. Removed after running.
- **F7** — a temporary `internal/cli/update/merge/zzz_probe_test.go` driving two successive
  `MergeGitignoreFile` calls with a template that drops `DS_Store` produced the migration and the
  three surviving `build/` duplicates quoted in `spec.md` §A. Removed after running.
- **F1** — `ls -la .moai/config/sections/` plus `git ls-files -s .moai/config/sections/`, showing
  four files at `-rw-------` against a git index recording `100644` for every one.

## §C Pre-flight

Before M1:

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — this checkout
   is shared and demonstrably concurrent (see the `report.yaml` chmod above). Do not begin without
   a clean divergence read.
2. `go test ./internal/config/... ./internal/cli/update/... -count=1` — record the green baseline.
3. `golangci-lint run --timeout=2m` — record the lint baseline.
4. Confirm `SPEC-UPDATE-DATA-SURVIVAL-001` status, since it is `depends_on`. If E2 has landed its
   shared-write work, M4 below becomes a consumer rather than an author — check before writing.

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

**Against the enum's fan-in.** `source.go:5` records `fan_in=71`. That number argues for care, not
for inaction. The change is narrow — it moves one constant within an `iota` block and shifts the
integer value of `SrcPlugin`, `SrcSkill`, `SrcSession`, and `SrcBuiltin` by nothing (only `SrcLocal`
and `SrcProject` swap). Three mitigations carry it: REQ-CTP-008 pins the full ordering in an
explicit guard so a future reorder fails loudly; REQ-CTP-009 requires
`internal/permission/resolver.go` — which iterates the same enum at `:230` — to be asserted
directly rather than assumed; NFR-CTP-005 keeps it a standalone revertable commit.

The one real hazard is any code that persists a `Source` as an integer. That must be grepped for
before landing, and is a milestone task, not an assumption.

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

**The gitignore coupling.** `CLAUDE.local.md` §22.9 already carries a `BLOCKER` note: the directory
is not gitignored, so a maintainer creating the opt-in file would commit it to a public repository
and flip the distributed default to `true`. Today that hazard is theoretical because the directory
is inert. M2 makes it load-bearing, so M2 must add the gitignore entry in the same milestone
(REQ-CTP-013) and clear the note (REQ-CTP-014). Splitting them would open a window in which the
hazard is real and undocumented.

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

The implementation pattern is not novel: `internal/settings/yamlpatch/yamlpatch.go:181-209` already
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

### M1 — Tier merge semantics (REQ-CTP-001 through REQ-CTP-005) — Priority High

The highest-blast-radius change in the SPEC and the one most likely to be revised under review.

1. **Enumerate the blast radius before writing code.** Grep every shipped
   `.moai/config/sections/*.yaml` and `internal/config/defaults.go` for falsey values — `false`,
   `0`, `""`, `[]`, `{}` — and produce a table of every key whose effective value changes once
   falsey values start winning. This table is a deliverable, not a footnote; R1 and R2 are
   mitigated by it or not at all.
2. Split "key present" from "value is zero" in the `MergeAll` tier walk. `valueExists` is already
   computed at `merge.go:144`; the fix is to stop discarding it at `:149-152`.
3. Verify the `PolicyOverrideRejected` path at `:155-162` becomes reachable for a falsey policy
   value — this is the security-relevant half of the change and needs its own assertion.
4. Resolve `isZero`'s fate per REQ-CTP-005: keep it where a caller genuinely needs zero-detection,
   delete it if none remains.
5. Land as a single revertable commit (NFR-CTP-005).

### M2 — Tier ordering and local-tier reachability (REQ-CTP-006 through REQ-CTP-014) — Priority High

Blocked by M1 per REQ-CTP-012.

1. Grep for any persistence of a `Source` as an integer — a stored ordinal would silently change
   meaning. This is the one unmitigated hazard in D-1 and must be checked before the reorder.
2. Move `SrcLocal` above `SrcProject` in the `iota` block.
3. Add the ordering guard (REQ-CTP-008) as an explicit sequence, not a computed one.
4. Assert `internal/permission/resolver.go` resolution under the new order (REQ-CTP-009).
5. Add local-tier reading to `internal/config/loader.go`, applied above `sections/`
   (REQ-CTP-010/011).
6. Add `.moai/config/local/` to `.gitignore` (REQ-CTP-013) **in this milestone**.
7. Amend `CLAUDE.local.md` §22.9 to drop the `BLOCKER` note (REQ-CTP-014).

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

The smallest change in the SPEC: three advisory emissions in `restore.go`.

1. Emit a first-occurrence advisory on 3-way merge failure.
2. Emit a distinguishable advisory when the base file is absent.
3. Confirm the ledger still governs repetition rather than first occurrence.

## §G Anti-Patterns

- **Deleting `isZero` reflexively.** It may have a legitimate caller elsewhere. REQ-CTP-005 requires
  a decision, not a deletion.
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
- `acceptance.md` §C — the falsification procedures.
- `.moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/plan.md` — the E2 boundary referenced in D-4.
- `CLAUDE.local.md` §22.9 — the opt-in doctrine amended by REQ-CTP-014.
