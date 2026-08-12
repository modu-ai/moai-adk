---
id: SPEC-CONFIG-TIER-PERSIST-001
title: "config resolution and persistence: explicit falsey values must win their tier, the local tier must be reachable, every config write must be atomic and mode-preserving, and a malformed section must never be silently written back as defaults"
version: "0.2.0"
status: draft
created: 2026-07-31
updated: 2026-08-12
author: manager-spec
priority: P0
phase: "v3.0.2"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "config, merge-tier, priority, atomic-write, file-mode, loader, gitignore, update, data-integrity"
related_specs: [SPEC-UPDATE-YAML-PRESERVE-001, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001, SPEC-V3R2-RT-005, SPEC-V3R6-UPDATE-NOISE-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001]
depends_on: [SPEC-UPDATE-DATA-SURVIVAL-001]
---

# SPEC-CONFIG-TIER-PERSIST-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Four-lens audit of `moai update` / `.moai/config`; Epic SPEC 3 of 6. Findings F1-F8 each reproduced by a runnable probe against HEAD `d5336214e`. |
| 0.2.0 | 2026-07-31 | Plan-audit revision (D1-D10 + D15). REQ-CTP-006 re-scoped from an `iota`-only reorder to a three-site reorder after `Priority()` was measured to have zero non-test consumers. REQ-CTP-033's reversal of `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007` declared explicitly in the new §K. REQ-CTP-013/014 marked already-landed at `b9fc75016`. REQ-CTP-005 split into two non-contradictory clauses. F1's file-mode evidence re-attributed to the primary checkout. |

## §A Problem / Motivation

The `.moai/config` subsystem has two independent halves that were audited together and found to
disagree with their own documentation in five places and to lose data in three.

**The resolution half** — `internal/config/merge.go` `MergeAll` plus the `Source` enum in
`internal/config/source.go` — implements an 8-tier priority system. Two defects make it
unable to express the thing a configuration system exists to express.

`MergeAll` walks tiers in priority order and skips any value that is the zero value for its type
(`internal/config/merge.go:149-152`, `if isZero(value) { continue }`). The consequence is that
nothing can ever be turned **off** by a higher tier: an explicit `false`, `0`, `""`, or empty list
loses to whatever the next tier down says, and ultimately to `SrcBuiltin`. Reproduced against HEAD
`d5336214e` with `SrcProject{enabled:false, count:0, name:""}` layered over
`SrcBuiltin{enabled:true, count:42, name:"default"}`:

```
workflow.branch_guard.enabled    => true ok=true source=builtin
x.count                          => 42 ok=true source=builtin
x.name                           => "default" ok=true source=builtin
```

An organisation policy tier that sets `enabled: false` is therefore silently ignored — and worse,
the `PolicyOverrideRejected` enforcement at `merge.go:155-162` is nested inside
`if winningValue != nil`, so when the policy value is falsey there is no winning value, no
rejection is raised, and `strict_mode` provides no protection at all for exactly the policies most
likely to be restrictive.

The second resolution defect is an ordering contradiction. `internal/config/source.go:26-28`
documents `SrcLocal` as "local overrides (not committed to git)" pointing at
`.claude/settings.local.json` and `.moai/config/local/*.yaml`, but every one of the three places
that expresses tier order puts `SrcLocal` *after* `SrcProject`. An override that cannot override is
not an override. Reproduced:

```
k => from-project (source=project) [project=2 local=3]
```

Which of the three places actually *decides* the outcome is not obvious, and the first draft of this
SPEC got it wrong. `Priority()` (`internal/config/source.go:73`) returns `int(s)`, the enum integer —
but it is read by nothing outside tests. Measured on the baseline tree:

```
$ grep -rn '\.Priority()' --include='*.go' internal/ | grep -v '_test.go'
(no output, exit 1)
```

Resolution order is decided instead by two independent hand-written literal slices:
`AllSources()` (`internal/config/source.go:103-114`), which `MergeAll` walks at
`internal/config/merge.go:137`, and a second `tiers` slice inside
`internal/permission/resolver.go:225-234`. Both list `SrcProject` before `SrcLocal`. Correcting the
`iota` block alone would therefore change no resolution anywhere while breaking
`TestAllSources`; REQ-CTP-006 is scoped to all three sites accordingly.

Compounding it, the local tier is unreachable from the typed configuration path entirely.
`CLAUDE.local.md` §22.9 instructs maintainers to opt into the branch guard by creating
`.moai/config/local/workflow.yaml`, but the branch guard reads through
`hook.ConfigProvider.Get()` → `config.Loader.Load(configDir)`, and
`grep -n "config/local\|localDir\|SrcLocal" internal/config/loader.go` returns **zero matches**.
The 8-tier `resolver` (`internal/config/resolver.go:386-390`, `loadLocalTier`) and the typed
`Loader` are two systems that never meet. A probe writing `sections/workflow.yaml`=false and
`local/workflow.yaml`=true observed `BranchGuard.Enabled = false` — the documented opt-in is inert.

**The persistence half** loses data in three distinct ways.

`ConfigManager.atomicWrite` (`internal/config/manager.go:420-437`) creates its temp file with
`os.CreateTemp`, which uses mode 0600, and then `os.Rename`s it over the target. `os.Rename` does
not inherit the destination's mode, so every `Save()` permanently narrows the file from 0644 to
0600. `defs.FilePerm` (`internal/defs/perms.go:11`) is never applied on this path. The narrowing is
observable in the **primary checkout** at `/Users/goos/MoAI/moai-adk-go` — four files sit at
`-rw-------` while the git index records `100644` for all of them, and all four are exactly
`Save()`'s `saveSection` targets:

```
$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/config/sections/*.yaml | awk '{print $1, $NF}'
...
-rw-------@ .../sections/git-convention.yaml
-rw-------@ .../sections/git-strategy.yaml
-rw-------@ .../sections/language.yaml
-rw-------@ .../sections/user.yaml
```

**Where this evidence does and does not reproduce.** The four narrowed files are visible only in the
primary checkout. In the plan-authoring worktree
(`.claude/worktrees/epic-update-config`, branch `plan/epic-update-config-audit`) the same command
reports **all 32 section files at `-rw-r--r--`**, with zero narrowed. That is expected rather than
contradictory: git records only the executable bit, so a `git worktree add` materialises every file
at the checkout umask and cannot carry a working-tree `chmod` across. The narrowing is a property of
the tree in which `Save()` actually ran, not of the commit.

Two consequences bind the rest of this SPEC. First, the file-mode observation is attributed to the
primary checkout and **not** to the baseline tree, so `acceptance.md` §A clause 3's "recorded from
this tree" does not cover it. Second, any acceptance criterion phrased as "the four narrowed files in
this repository read `-rw-r--r--`" is trivially true in a worktree and has no discriminating power;
AC-CTP-022 therefore asserts the migration against a `t.TempDir()` fixture, and the primary-checkout
widening is recorded as a manual follow-up rather than an AC. The *mechanism* is independently
confirmed and is not affected: a direct probe observed a pre-existing `-rw-r--r--` target become
`-rw-------` after one `atomicWrite` round trip.

The correct pattern already exists in-repo: `internal/settings/yamlpatch/yamlpatch.go:202` calls
`os.Chmod(tmpName, info.Mode().Perm())` before renaming. `manager.go:161` carries an `@MX:REASON`
declaring atomic write critical to data integrity — yet five of the six writers that touch
`.moai/config/**` use bare `os.WriteFile`, which truncates before writing:
`internal/cli/update/backup/restore.go:105,128,142,145`,
`internal/cli/update/merge/merge.go:84`, `internal/core/project/initializer.go:371-450`, and
`internal/cli/harness.go:390` (which additionally hardcodes `0o644` instead of `defs.FilePerm`).

That non-atomicity is dangerous specifically because of the third persistence defect. Every one of
the fourteen section loaders in `internal/config/loader.go` swallows a load failure into compiled
defaults (`loader.go:121-131` and thirteen siblings, all `slog.Warn(... "using defaults")`). A hook
reading configuration while `moai update` truncates a section file observes a malformed file, gets
builtin defaults for that turn, and surfaces nothing the operator can see. If a
`ConfigManager.Save()` follows in the same process, those builtin defaults are then written back
over the user's real configuration — a silent, permanent read-modify-write destruction of user
data. Narrowed file modes make the same failure reachable from a different direction: a container
or CI step running as a different UID can no longer read a 0600 section file, and the resulting
read failure is swallowed identically.

Two further defects were found on the update path. `MergeGitignoreFile`
(`internal/cli/update/merge/merge.go:41-85`) classifies every backup line absent from the *new*
template as user-authored, so a pattern the template **drops** is migrated into the
`# User Custom Patterns` section and stays there forever; duplicates are preserved verbatim because
nothing dedupes. The header is emitted as a comment and skipped as a comment on the next pass, so
the file does not grow without bound — but the reclassification is permanent. Reproduced with a
template that drops `DS_Store` between updates:

```
after update2 (108 bytes):
node_modules/

# User Custom Patterns (preserved by moai update)
DS_Store          <- template-dropped pattern, now permanently user-authored
mysecret.env
build/
build/
build/            <- three copies preserved
```

Finally, the more serious of the two merge failure paths is the quieter one. In
`internal/cli/update/backup/restore.go`, a **3-way** merge failure calls `recordFallback` and falls
through silently (`restore.go:131-134`) — the noise-suppression ledger stays quiet until three
consecutive failures — while a **2-way** failure prints an immediate stderr warning
(`restore.go:139-141`). A third path is quieter still: when the base file is absent,
`os.ReadFile(basePath)` errors and control falls to 2-way with neither a `recordFallback` call nor
a warning. This matters today, not hypothetically: `quality.yaml` takes the 3-way-fails path on
every update (recorded in `SPEC-UPDATE-YAML-PRESERVE-001`'s `audit.md`), so users are silently
receiving 2-way merges while believing they receive 3-way.

## §B Goals

- An explicit falsey value set by a higher tier wins its tier, and `strict_mode` policy rejection
  becomes reachable for falsey policy values.
- The `SrcLocal` tier means what its own documentation says, and is reachable from the typed
  `Loader` path so the documented `.moai/config/local/` opt-in actually works.
- Every write into `.moai/config/**` goes through one atomic, mode-preserving helper, and the four
  already-narrowed files are widened back.
- A malformed section file is never silently converted into compiled defaults that are then written
  back over the user's data.
- `.gitignore` merging stops reclassifying template-dropped patterns as user-authored and stops
  preserving duplicates.
- The 3-way merge fallback is at least as visible as the 2-way fallback.

## §C Scope Exclusions

### Out of Scope — the YAML merge engine internals

- `DeepMerge3Way`, `ValuesEqual`, `systemFields`, sequence handling, and the old-only-key drop
  inside `internal/cli/update/backup/merge.go`. All owned by `SPEC-UPDATE-YAML-PRESERVE-001`.
  This SPEC records four **amendment requests** against that SPEC in §I with evidence, and
  specifies no fix for any of them.

### Out of Scope — 3-way merge base provenance

- The fact that the 3-way merge base is derived from the NEW template, so `base ≡ new`. Owned by
  the deferred `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`. §J carries forward the concrete
  release-tag evidence for that SPEC so it is not lost; this SPEC changes no base-selection code.

### Out of Scope — backup coverage and the update failure contract

- Whether a backup exists on disk before a destructive step, the recovery manifest, and the restore
  entry point. Owned by `SPEC-UPDATE-DATA-SURVIVAL-001`, which this SPEC declares in `depends_on`.
  The boundary is precise: E2 owns *that the bytes reach disk before deletion*; this SPEC owns
  *that the write itself is atomic and preserves mode*. The shared helper introduced here is the
  mechanism E2's writes will use; E2 specifies no write mechanism of its own.

### Out of Scope — v2 detection and deprecated-path handling

- Semver-normalised v2/v3 detection, `defs.DeprecatedPaths` membership, and residue cleanup. Owned
  by `SPEC-UPDATE-REINSTALL-LOOP-002`.

### Out of Scope — the SrcUser / SrcProject relative order

- `SrcUser` currently outranks `SrcProject`, which is unusual for a project-scoped tool. No audit
  lens produced evidence of a defect arising from it, and no requirement here depends on it.
  Re-litigating that pair is deliberately excluded so the enum change stays minimal and its blast
  radius stays assessable. Recorded as an observation, not a finding.

### Out of Scope — template content

- Any edit under `internal/template/templates/**`. That tree is distributed across 16 languages and
  must carry no SPEC ID, REQ token, internal date, or commit SHA. All fixtures live in `_test.go`
  files created with `t.TempDir()`.

### Out of Scope — dead config keys, CI gates, documentation drift

- Owned by Epic SPECs 4 through 6.

## §D Requirements (GEARS)

### D.1 Tier merge semantics

- **REQ-CTP-001** — `MergeAll` shall distinguish *a key being present in a tier* from *that key's
  value being the zero value for its type*, and shall select the first tier in priority order in
  which the key is present, regardless of whether that value is falsey.

- **REQ-CTP-002** — **When** a higher-priority tier supplies an explicit `false`, `0`, `""`, or
  empty collection for a key, `MergeAll` shall record that value as the winning value with that
  tier as its `Provenance.Source`.

- **REQ-CTP-003** — **While** `policy.strict_mode` is enabled, `MergeAll` shall raise
  `PolicyOverrideRejected` for a lower tier attempting to override a policy-designated key whose
  policy value is falsey, on the same terms as a non-falsey policy value.

- **REQ-CTP-004** — `MergeAll` shall continue to omit from its output any key that is absent from
  every tier; absence and falsey-presence shall remain distinguishable in the merged result.

- **REQ-CTP-005a** — **Where** `isZero` retains at least one caller after this change whose semantics
  genuinely require zero-detection, the helper shall be retained together with that caller.

- **REQ-CTP-005b** — **Where** `isZero` retains no caller after this change, the helper shall be
  removed rather than left unreferenced.

  > REQ-CTP-005a/b replace the single REQ-CTP-005 of v0.1.0, whose text simultaneously forbade and
  > mandated deletion ("shall not be deleted … it shall be removed"). The intent was always a
  > disposition decision conditioned on caller survival; the two clauses state that without
  > self-contradiction. Covered by AC-CTP-037.

### D.2 Tier ordering and local-tier reachability

- **REQ-CTP-006** — Tier order shall place `SrcLocal` above `SrcProject` at **all three** sites that
  express it, in a single commit: the `iota` block in `internal/config/source.go` (so that
  `SrcLocal.Priority() < SrcProject.Priority()`), the `AllSources()` literal slice
  (`internal/config/source.go:103-114`, which `internal/config/merge.go:137` walks), and the `tiers`
  literal slice in `internal/permission/resolver.go:225-234`. Reordering fewer than all three shall
  be treated as an incomplete change.

  > The `iota` block alone does not decide resolution. `Priority()` (`source.go:73`) returns
  > `int(s)` and has zero non-test consumers on the baseline tree, so an `iota`-only reorder leaves
  > both literal slices — and therefore both `MergeAll` and the permission resolver — resolving in
  > the old order, while breaking `TestAllSources`'s `AllSources()[i].Priority() == i` invariant.
  > Reordering the `iota` block and `AllSources()` **together** restores that invariant, so
  > `TestAllSources` needs no edit; `TestSourceOrdering`'s `expectedOrder` literal does (REQ-CTP-008).

- **REQ-CTP-007** — The relative order of every other pair of tiers shall be unchanged, and
  `SrcPolicy` shall remain the numerically highest-priority tier.

- **REQ-CTP-008** — The existing `TestSourceOrdering` guard (`internal/config/source_test.go:104-119`)
  already pins the complete tier ordering against a literal, hand-written sequence. Its
  `expectedOrder` literal shall be updated to the new order **in the same commit** as REQ-CTP-006 and
  shall be retained; it shall not be deleted, weakened, or replaced by an enum-derived comparison.

  > v0.1.0 recorded this as "add a guard". The guard already exists and is green — the observed
  > `expectedOrder` literal is `[SrcPolicy, SrcUser, SrcProject, SrcLocal, SrcPlugin, SrcSkill,
  > SrcSession, SrcBuiltin]`. The real obligation is to update its literal rather than to author a
  > duplicate. `TestAllSources` (`:89-102`) is a separate, complementary invariant
  > (`AllSources()[i].Priority() == i`) and needs no edit under REQ-CTP-006's three-site reorder.

- **REQ-CTP-009** — `internal/permission/resolver.go` shall be shown to resolve permissions
  correctly under the new ordering. It does **not** iterate the `Source` enum: it carries its own
  literal `tiers` slice at `:225-234`, so REQ-CTP-006's reorder reaches it only because that slice is
  one of the three named sites, and the assertion is therefore load-bearing rather than confirmatory.

- **REQ-CTP-010** — The typed `Loader` shall read `.moai/config/local/*.yaml` and apply it above
  `.moai/config/sections/*.yaml`, so that a value set in the local directory overrides the same
  value set in the sections directory.

- **REQ-CTP-011** — **When** `.moai/config/local/` is absent, the `Loader` shall behave exactly as
  it does today, with no warning and no error.

- **REQ-CTP-012** — The `Loader`'s local-tier read shall be applied **after** REQ-CTP-001 is in
  force, so that the local directory can both enable and disable a setting. A local tier that can
  only turn things on is a new asymmetry and shall not be shipped.

- **REQ-CTP-013** — **ALREADY SATISFIED at `b9fc75016`; retained as a regression guard only.**
  `.moai/config/local/` shall remain present in the repository `.gitignore`, so that a
  maintainer-local opt-in cannot be committed and thereby flip a distributed default. No M2 work item
  follows from this requirement; AC-CTP-012 asserts the entry has not been removed.

  > Measured: `git check-ignore -v .moai/config/local/workflow.yaml` exits 0 and prints
  > `.gitignore:183:.moai/config/local/`. The commit that added it, `b9fc75016`, is an ancestor of the
  > `d5336214e` baseline (`git merge-base --is-ancestor b9fc75016 d5336214e` → exit 0), so the work
  > was already complete when this SPEC was drafted.

- **REQ-CTP-014** — **ALREADY SATISFIED; retained as a regression guard only.** `CLAUDE.local.md`
  §22.9 shall continue to carry no `BLOCKER (gitignore)` note and shall continue to state that the
  documented opt-in is effective. No M2 work item follows from this requirement; AC-CTP-013 asserts
  the note has not been reintroduced.

  > Measured: `grep -c 'BLOCKER (gitignore)' CLAUDE.local.md` → `0`. §22.9 already reads
  > "**gitignore (해결됨)**: `.moai/config/local/` 디렉터리는 이제 `.gitignore`에 등록되었다".

### D.3 Malformed-configuration contract

- **REQ-CTP-015** — The section loaders shall distinguish an **absent** section file from a
  **malformed** one. An absent file shall continue to yield compiled defaults silently.

- **REQ-CTP-016** — **When** a section file is present but fails to parse, the `Loader` shall
  record that section as failed, and shall not mark it loaded.

- **REQ-CTP-017** — **Where** the caller is a CLI command, a malformed section file shall surface
  as a returned error naming the file and the parse failure, and the command shall exit non-zero.

- **REQ-CTP-018** — **Where** the caller is a hook, a malformed section file shall not abort the
  hook; the hook shall continue with compiled defaults for that section and shall emit an operator-
  visible advisory on stderr naming the file, in addition to the existing `slog.Warn`. This
  preserves the fail-open norm that `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  establishes for hooks.

- **REQ-CTP-019** — **When** any section failed to parse during the most recent `Load`,
  `ConfigManager.Save` shall refuse to persist that section and shall return an error naming it,
  rather than writing compiled defaults over the user's file.

- **REQ-CTP-020** — The refusal in REQ-CTP-019 shall be scoped to the failed section alone;
  sections that loaded cleanly shall still be persisted.

### D.4 Atomic, mode-preserving writes

- **REQ-CTP-021** — A single shared helper shall perform every write into `.moai/config/**` using
  temp-file-plus-rename, and shall set the temp file's mode to the pre-existing target's mode
  before renaming.

- **REQ-CTP-022** — **When** the target file does not yet exist, the shared helper shall create it
  with `defs.FilePerm`, never with `os.CreateTemp`'s 0600 default.

- **REQ-CTP-023** — `ConfigManager.atomicWrite`, `backup.RestoreMoaiConfig`,
  `merge.MergeGitignoreFile`, `merge.MergeUserFiles`, `core/project/initializer`, and
  `cli/harness.go` shall each route their `.moai/config/**` writes through the shared helper.

- **REQ-CTP-024** — `internal/cli/harness.go:390` shall use `defs.FilePerm` rather than a
  hardcoded `0o644` literal, per the repository's hardcoding-prevention rule.

- **REQ-CTP-025** — A migration step shall widen any `.moai/config/sections/*.yaml` file whose mode
  is narrower than `defs.FilePerm` back to `defs.FilePerm`, because a `Stat`-based preservation fix
  alone perpetuates every file already narrowed.

- **REQ-CTP-026** — The migration in REQ-CTP-025 shall only widen; it shall never narrow a file the
  operator deliberately restricted beyond `defs.FilePerm`, and shall never alter a file outside
  `.moai/config/`.

- **REQ-CTP-027** — A guard shall assert that no writer under `internal/config/`,
  `internal/cli/update/`, or `internal/core/project/` reaches `.moai/config/**` via bare
  `os.WriteFile`, so a future writer cannot silently reintroduce the truncate-then-write window.

### D.5 `.gitignore` merge correctness

- **REQ-CTP-028** — `MergeGitignoreFile` shall parse the `# User Custom Patterns` header as a real
  section boundary, and shall treat only lines below that header in the backup as user-authored.

- **REQ-CTP-029** — **When** the new template drops a pattern that the previous template contained
  and the user never added, `MergeGitignoreFile` shall not migrate that pattern into the user
  section.

- **REQ-CTP-030** — `MergeGitignoreFile` shall emit each user pattern at most once, preserving the
  first occurrence's original text and discarding subsequent exact duplicates.

- **REQ-CTP-031** — `MergeGitignoreFile` shall remain idempotent: merging a file with its own
  previous output shall produce byte-identical content.

- **REQ-CTP-032** — **When** a backup predates the header convention and therefore contains no
  `# User Custom Patterns` header, `MergeGitignoreFile` shall fall back to the current
  not-in-template heuristic rather than discarding the user's patterns.

### D.6 Fallback visibility

- **REQ-CTP-033** — **When** a 3-way merge fails and the restore path falls back to 2-way, the
  operator shall receive a stderr advisory naming the file, on the first occurrence rather than
  after three. **This supersedes the silent-first-occurrence clause of
  `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007`, which is `status: implemented`.** The reversal, the
  alternative that was rejected, and the parts of `REQ-UN-007` that survive are declared in §K.
  Landing REQ-CTP-033 without the §K reconciliation is prohibited.

- **REQ-CTP-034** — **When** the 3-way base file is absent and control falls through to 2-way
  without a merge attempt, the operator shall receive an advisory distinguishable from the
  merge-failure advisory of REQ-CTP-033.

- **REQ-CTP-035** — The advisory of REQ-CTP-033 shall be at least as prominent as the existing
  2-way failure warning at `restore.go:139-141`; the noise-suppression ledger shall continue to
  govern *repetition*, not *first occurrence*.

## §E Non-Functional Constraints

- **NFR-CTP-001** — Every test creates its temporary directories with `t.TempDir()` and touches no
  path outside them. No test reads or writes the operator's real `~/.moai`, `~/.claude`, or the
  repository's own `.moai/config/`.
- **NFR-CTP-002** — No test sets an `OTEL_*` environment variable, per the parallel-test data-race
  hazard recorded in `CLAUDE.local.md` §2.
- **NFR-CTP-003** — Go conventions hold: `snake_case.go` file names, `fmt.Errorf("...: %w", err)`
  error wrapping, English comments and godoc.
- **NFR-CTP-004** — No edit under `internal/template/templates/**`.
- **NFR-CTP-005** — The tier-ordering change is a single commit that touches the enum and its
  guards only, so it can be reverted independently of the merge-semantics change.
- **NFR-CTP-006** — The mode-widening migration is idempotent and safe to run on a tree where it
  has already run.

## §F Success Criteria

- An explicit `false` from `SrcProject` beats `true` from `SrcBuiltin`, verified by a test that
  reproduces the §A probe and now observes `source=project`.
- `SrcLocal.Priority() < SrcProject.Priority()`, and `.moai/config/local/workflow.yaml` setting
  `branch_guard.enabled: true` is observed through `Loader.Load` when
  `sections/workflow.yaml` says `false` — and the reverse also holds.
- A `Save()`-then-`stat` round trip in a `t.TempDir()` fixture preserves 0644, and the migration
  widens a `0600` fixture file to `0644`. (The four narrowed files in the primary checkout are a
  manual follow-up, not a criterion — see §A on why that observation does not reproduce in a
  worktree.)
- A truncated section file causes a CLI error and a hook advisory, and a subsequent `Save()`
  refuses to overwrite that section.
- The §A `.gitignore` probe, re-run after the fix, shows `DS_Store` absent from the user section
  and exactly one `build/`.
- A 3-way merge failure prints an advisory on its first occurrence.
- `go test ./...` and `golangci-lint run` are clean.

## §G Risk Register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------:|--------|------------|
| R1 | A previously-ignored `false` starts winning and disables a feature that users depend on being on | High | High | Enumerate every falsey value present in any shipped `sections/*.yaml` and in `defaults.go` before landing REQ-CTP-001; the enumeration is a milestone deliverable, not a footnote. Ship the merge-semantics change alone in one revertable commit. |
| R2 | A previously-ignored `""` starts winning and blanks a string that had a working default | Medium | High | Same enumeration as R1; empty strings are the subtlest case because a blank name or path often fails far from the config layer. |
| R3 | The reorder **fails to reach** the production path: `Priority()` has zero non-test consumers, so an `iota`-only edit changes no resolution while breaking `TestAllSources` | High | High | REQ-CTP-006 names all three ordering sites explicitly (`iota`, `AllSources()`, `resolver.go` `tiers`); AC-CTP-005 asserts the merge outcome rather than the ordinal alone; AC-CTP-007 asserts permission resolution directly. The v0.1.0 framing of this risk — ripple through `fan_in ≈ 71` — had the direction backwards; the measured hazard is under-reach, not over-reach. |
| R4 | Wiring the local tier while zero-values are still skipped produces an opt-in that cannot opt out | Medium | Medium | REQ-CTP-012 orders the milestones so this cannot happen. |
| R5 | `.moai/config/local/` reaches a public commit and flips a distributed default | Low | High | Already mitigated before this SPEC: the `.gitignore` entry landed at `b9fc75016`, an ancestor of the baseline. REQ-CTP-013 degenerates to a regression guard (AC-CTP-012) that fails if the entry is removed. |
| R6 | Turning a malformed config into a CLI error breaks a workflow that currently limps along on defaults | Medium | Medium | REQ-CTP-015 keeps *absent* silent; only *malformed* is loud. Hooks stay fail-open per REQ-CTP-018. |
| R7 | The mode-widening migration widens a file an operator deliberately restricted | Low | Medium | REQ-CTP-026 scopes it to `.moai/config/` and to widening-toward-`defs.FilePerm` only; the migration is announced in its output. |
| R8 | Header-based `.gitignore` parsing discards patterns from a pre-header backup | Medium | High | REQ-CTP-032 requires the heuristic fallback when no header is present. |
| R9 | Making the 3-way fallback loud produces noise on every update while `quality.yaml` still fails 3-way — **and reverses `REQ-UN-007` of the implemented `SPEC-V3R6-UPDATE-NOISE-001`, whose whole purpose was to suppress exactly this output** | High | Medium | Accepted and intentional, but only as a declared supersession: §K states the reversal, records the rejected `--verbose` alternative, and preserves the counter and 3-strike advisory that `REQ-UN-007`/`REQ-UN-008` also own. `SPEC-UPDATE-YAML-PRESERVE-001` removes the underlying cause; until then operators should see it. REQ-CTP-035 keeps the ledger governing repetition. |

## §I Amendment Requests to `SPEC-UPDATE-YAML-PRESERVE-001`

These are findings that belong to the sibling SPEC's ownership of
`internal/cli/update/backup/merge.go`. They are recorded here as **requests with evidence** so its
post-FAIL revision can absorb them. This SPEC specifies no fix for any of them.

### A1 — sequences are atomic, so one user edit freezes a whole list

`DeepMerge3Way` does not treat lists as mergeable. Observed with
`new=[go,python,rust]`, `old=[go,python,mine]`, `base=[go,python]` → result `[go,python,mine]`; the
template's new `rust` is lost. The real consequence is that a user who adds a single `exclude:`
pattern to `mx.yaml` never receives any future template exclusion for that key.
`SPEC-UPDATE-YAML-PRESERVE-001`'s `acceptance.md:154-156` currently pins this behaviour as
intended. The request is not to reverse the decision but to **re-confirm it against this evidence**,
because the loss is silent and the pinned rationale predates the observation.

### A2 — `ValuesEqual` compares `fmt.Sprintf("%v")` strings, collapsing types

At `internal/cli/update/backup/merge.go:103-111`. Observed to report equal for `(1,"1")`,
`(true,"true")`, `(1.0,1)`, `([]any{1,2},"[1 2]")`, and `(map[string]any{"a":1},"map[a:1]")`. Since
equality drives the "the user did not change this, so take the template value" branch, a type
collision silently discards a user's explicit choice.

### A3 — `systemFields` applies at every nesting depth

The set is declared inside the function body (`merge.go:47-50`) and re-applied on each recursion
(`merge.go:79`), so **any** key named `version` at any depth takes the template value. Nested
`version` keys exist in shipped templates: `delegation.yaml:18`, `mx.yaml:4`, `design.yaml:7`. The
`design.version` case is the concerning one — it may legitimately be the user's own design-system
version rather than a MoAI system field.

### A4 — the old-only-key drop recurs at nested depth

`REQ-UYP-006` says "a key present in the user's document and absent from the new template" without
naming a scope, and cites only a top-level comment as evidence. Observed nested loss:
`new=s:{x:1}`, `old=s:{x:1,mine:9}`, `base=s:{x:1}` → `s.mine` gone. The request is for the
requirement text to say "at every nesting depth" and for the acceptance fixture to include a nested
case.

## §J Evidence carried forward for `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`

That SPEC is deferred and not yet authored. It owns the fact that the 3-way merge base is derived
from the NEW template, making `base ≡ new`. An audit lens ran the real merge with actual
release-tag bytes and identified the concretely affected keys between `v3.0.0-rc8` and `v3.0.1`.
Recorded here so it survives:

| file | key | rc8 | v3.0.1 | what an upgrading rc8 user actually gets |
|---|---|---|---|---|
| `security.yaml` | `strict_mode` | false | true | **false** — never arrives |
| `cache.yaml` | `enabled` | false | true | **false** |
| `harness.yaml` | effort minimal / standard / thorough | medium / high / xhigh | low / medium / high | **medium / high / xhigh** |

Newly-**added** keys do arrive, because the `!oldExists` branch handles them. Only **changed
defaults** are stuck. The net effect is that a security default raised to `true` reaches new
installations only. Note the interaction with this SPEC: `security.strict_mode` and `cache.enabled`
are both booleans whose stuck value is `false`, which is also the value REQ-CTP-001 makes winnable
— the two SPECs touch the same keys from opposite directions and their fixtures should not be
assumed independent.

## §K Reconciliation with `SPEC-V3R6-UPDATE-NOISE-001` (declared reversal)

`SPEC-V3R6-UPDATE-NOISE-001` is `status: implemented`. Its `REQ-UN-007` reads, verbatim:

> `mergeYAML3Way` 가 실패하여 2-way fallback 으로 전환할 때, 시스템은 merge-history ledger 의
> `fallback_count` 를 1 증가시켜야 한다. `fallback_count` 가 3 미만이면 stderr 로 어떤 출력도
> emit 하지 않는다 (silent fallback).

REQ-CTP-033 commands the opposite of that second sentence. The implementation carries the constant
(`internal/cli/update_noise.go:37`, `const fallbackAdvisoryThreshold = 3`) and the call site cites the
requirement by name (`internal/cli/update/backup/restore.go`, the comment
`// REQ-UN-007/008/010: emit advisory or legacy warning per noise-suppression policy.` immediately
above the `recordFallback(projectRoot, relPath, false, os.Stderr)` call). v0.1.0 of this SPEC named
neither the sibling SPEC nor the requirement anywhere — `grep -rn 'UPDATE-NOISE\|REQ-UN-'` over this
SPEC directory returned no matches. Reversing an implemented sibling requirement in silence is the
defect this section exists to close.

### K.1 — the alternative that was rejected

`REQ-UN-010` already provides a diagnostic escape hatch: `--verbose` bypasses the silent branch and
restores the per-occurrence `Warning: 3-way merge failed for <relPath>, falling back to 2-way`
message. Satisfying this SPEC's F8 concern through `--verbose` alone would require no reversal at all.

It is rejected because the two requirements answer different questions. `--verbose` serves an operator
who **already suspects** something is wrong and is opting into detail. F8's finding is that the
operator does not suspect anything: `quality.yaml` takes the 3-way-fails path on *every* update, and
the operator's belief that they are receiving a 3-way merge is never contradicted. A flag that must be
passed by someone who already knows to pass it cannot correct a belief held by someone who does not.
The information has to arrive unasked at least once.

### K.2 — what is superseded, and what survives

| `REQ-UN-*` clause | Disposition under REQ-CTP-033 |
|---|---|
| `REQ-UN-007` — increment `fallback_count` on every 3-way failure | **Survives unchanged.** The counter is not the noise. |
| `REQ-UN-007` — emit nothing to stderr while `fallback_count < 3` | **Superseded.** The first occurrence becomes visible. |
| `REQ-UN-008` — advisory at the 3-strike threshold | **Survives.** REQ-CTP-035 keeps the ledger governing repetition; the threshold advisory still marks sustained failure. |
| `REQ-UN-009` — reset the counter on 3-way success | **Survives unchanged.** |
| `REQ-UN-010` — `--verbose` restores the legacy per-occurrence warning | **Survives.** It becomes redundant with the new first-occurrence advisory for the *first* failure and remains the only source of per-occurrence output for failures 2 and beyond. |

### K.3 — obligations this places on the run phase

1. The commit implementing REQ-CTP-033 shall state in its body that it supersedes `REQ-UN-007`'s
   silent-first-occurrence clause and shall reference this section.
2. `SPEC-V3R6-UPDATE-NOISE-001`'s own artifacts are **not** edited by this SPEC. Recording the
   supersession in that SPEC is a sync-phase concern for its owner, surfaced here as a cross-SPEC
   note rather than performed unilaterally.
3. If the reversal is rejected during Implementation Kickoff Approval, REQ-CTP-033 and AC-CTP-031
   are withdrawn together and M6 reduces to REQ-CTP-034 (the absent-base advisory), which conflicts
   with nothing — the absent-base path today emits neither a `recordFallback` call nor a warning, so
   making it visible adds a signal rather than reversing a suppression.

## §H Cross-References

- `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/` — YAML merge engine internals; §I amendment requests.
- `.moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/` — backup coverage and the update failure contract;
  `depends_on`. Its §C explicitly assigns "merge-tier semantics" to Epic SPECs 3 through 6.
- `.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/` — v2 detection, deprecated-path handling.
- `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` — deferred; §J evidence.
- `SPEC-V3R2-RT-005` — the 8-tier system this SPEC corrects; `source.go:5` `@MX:ANCHOR` records
  `fan_in=71`.
- `.moai/specs/SPEC-V3R6-UPDATE-NOISE-001/` — `status: implemented`. Its `REQ-UN-007`
  silent-first-occurrence clause is superseded by REQ-CTP-033; see §K for the declared reversal,
  the rejected `--verbose` alternative, and the clauses that survive.
- `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001` — the branch-guard opt-in whose documented mechanism
  REQ-CTP-010 makes functional.
- `CLAUDE.local.md` §22.9 — the maintainer opt-in doctrine. Its former `BLOCKER (gitignore)` note is
  already cleared (REQ-CTP-014 is a regression guard, not a work item).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — the hook fail-open norm that
  REQ-CTP-018 preserves.

## §L Split Branches

> **Split status: complete — all three slices extracted** (2026-08-12). The parent SPEC's `status:`
> remains `draft` — the split does NOT transition the parent's lifecycle state. Each slice ships as an
> independent child SPEC; the parent is retained as the lineage carrier and the §G/§K
> risk-and-reconciliation provenance of record.

`SPEC-CONFIG-TIER-PERSIST-001` was an over-large (35-REQ, Tier-M-exceeding) SPEC. Per the 3-way
split recommendation, the parent is divided into three independently-shippable slices. All three
slices have now been extracted into independent child SPECs; the subsections below record each
slice, the child SPEC that owns it, and the parent REQs it carries.

### Extracted — slice (b) → `SPEC-CONFIG-ATOMIC-WRITE-001`

| Attribute | Value |
|-----------|-------|
| **Date** | 2026-08-12 |
| **Child SPEC** | `SPEC-CONFIG-ATOMIC-WRITE-001` (status: draft, Tier M, era V3R6) |
| **Slice** | (b) — atomic, mode-preserving writes for `.moai/config` + CLI persistence |
| **Rationale** | Independently shippable. Carved out of a 35-REQ over-large SPEC per the 3-way split recommendation; the atomic-write + mode-preservation invariant is self-contained, has no dependency on tier resolution (slice a) or on malformed-section writeback / gitignore merge (slice c), and is the highest-priority shippable slice (live mode-narrowing defect at `manager.go:420-438` + non-atomic hardcoded write at `harness.go:390`). |
| **REQ mapping** | Parent `REQ-CTP-021/022/023/024/027` → child `REQ-CAW-001..007` (atomicity, mode preservation, shared helper, call-site remediation, no-literal rule, temp cleanup, regression guard). Parent `REQ-CTP-025/026` (mode-widening migration) extracted into follow-up sibling `SPEC-CONFIG-MODE-MIGRATE-001` (status: draft, Tier S, commit `4c2b998ea`, 2026-08-12) — carries `REQ-MIG-001/002` (dry-run-first widening + only-widen/never-narrow scope); the child's scope is the one-time migration, complementing this slice's write-path invariant. |
| **Lineage carrier** | Child `spec.md` `related_specs: [SPEC-CONFIG-TIER-PERSIST-001]` + this section (bi-directional link). |

### Extracted — slice (a) → `SPEC-CONFIG-TIER-RESOLVE-001`

| Attribute | Value |
|-----------|-------|
| **Date** | 2026-08-12 |
| **Child SPEC** | `SPEC-CONFIG-TIER-RESOLVE-001` (status: draft, Tier M, era V3R6) |
| **Slice** | (a) — tier resolution: explicit-falsey-wins-tier semantics, `SrcLocal` above `SrcProject` at all three ordering sites, typed-Loader / resolver disconnect, local-tier reachability |
| **Rationale** | Independently shippable. Carved out of the 35-REQ parent per the 3-way split. Slice (a) is entirely orthogonal to the write path (slices b/c) — it concerns *which* tier wins when sources disagree, not *how* the result is written to disk — so it proceeds in parallel with the write-path slices without conflict. |
| **REQ mapping** | Parent `REQ-CTP-001..014` (§D.1 tier merge semantics + §D.2 tier ordering / local-tier reachability, including the two already-satisfied regression guards 013/014) → child `REQ-CTR-001..014`. The three parent-measured facts carry verbatim: `Priority()` has zero non-test consumers (an iota-only reorder is inert); `resolver.go:225-234` carries its own literal `tiers` slice (the REQ-CTR-009 assertion is load-bearing, not confirmatory); the `isZero` disposition is a caller-count decision, not a blanket delete. |
| **Lineage carrier** | Bi-directional: child `related_specs: [SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-ATOMIC-WRITE-001]` + this section. |

### Extracted — slice (c) → `SPEC-CONFIG-SECTION-INTEGRITY-001`

| Attribute | Value |
|-----------|-------|
| **Date** | 2026-08-12 |
| **Child SPEC** | `SPEC-CONFIG-SECTION-INTEGRITY-001` (status: draft, Tier M, era V3R6) |
| **Slice** | (c) — malformed-section writeback-as-defaults prevention, `MergeGitignoreFile` header-based correctness, dropped-pattern migration, 3-way→2-way fallback visibility |
| **Rationale** | Independently shippable. Carries the parent's §K reconciliation (REQ-CTP-033's declared supersession of `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007`) verbatim into the child §K — landing the fallback-visibility REQ without that reconciliation is prohibited. The malformed/gitignore writers MAY adopt the slice-(b) shared atomic-write helper (API locked at slice-(b) `plan.md` M1) without re-deriving it. |
| **REQ mapping** | Parent `REQ-CTP-015..020` (§D.3 malformed-configuration contract) + `REQ-CTP-028..032` (§D.5 .gitignore merge correctness) + `REQ-CTP-033..035` (§D.6 fallback visibility) → child `REQ-CSI-001..014`. The §K reconciliation content (REQ-UN-007 supersession, rejected `--verbose` alternative, surviving REQ-UN-* clauses, commit-body obligation) is migrated into the child §K. |
| **Lineage carrier** | Bi-directional: child `related_specs: [SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-ATOMIC-WRITE-001, SPEC-CONFIG-TIER-RESOLVE-001, SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001, SPEC-V3R6-UPDATE-NOISE-001]` + this section. |

### Deferred

| Slice / REQ | Disposition |
|-------------|-------------|
| `REQ-CTP-025/026` (mode-widening migration) | **Extracted** into child `SPEC-CONFIG-MODE-MIGRATE-001` (status: draft, Tier S, commit `4c2b998ea`, 2026-08-12) — carries `REQ-MIG-001/002` (dry-run-first widening + only-widen/never-narrow scope). No longer deferred. |

No slice remains resident in the parent — all three primary slices (a)/(b)/(c) plus the deferred mode-migration slice (d) have been extracted into independent
child SPECs. The parent is retained as the lineage carrier and the §G/§K risk-and-reconciliation
provenance of record; its own `status:` stays `draft`.
