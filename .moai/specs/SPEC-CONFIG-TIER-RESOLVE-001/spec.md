---
id: SPEC-CONFIG-TIER-RESOLVE-001
title: "config tier resolution: explicit falsey values must win their tier, SrcLocal must resolve above SrcProject at all three ordering sites, and the local tier must be reachable for both enable and disable"
version: "0.1.0"
status: draft
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P0
phase: "v3.0.2"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "config, merge-tier, priority, falsey-wins, local-tier, SrcLocal, resolver, reachability"
related_specs: [SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-ATOMIC-WRITE-001]
---

# SPEC-CONFIG-TIER-RESOLVE-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-12 | Initial draft. Carved out of the over-large parent `SPEC-CONFIG-TIER-PERSIST-001` (35-REQ, Tier-M-exceeding) as **slice (a)** of the 3-way split recorded in the parent's §L Split Branches. Re-namespaces the parent's §D.1 "Tier merge semantics" (REQ-CTP-001..005b) and §D.2 "Tier ordering and local-tier reachability" (REQ-CTP-006..014) into child REQ-CTR-001..014, preserving the parent's exact semantic content verbatim — including the measured facts the parent already paid for (e.g. `Priority()` has zero non-test consumers; `resolver.go` carries its own literal `tiers` slice; `isZero` disposition is a caller-count decision, not a blanket delete/keep). Slice (b) (atomic, mode-preserving writes) is already CLOSED as `SPEC-CONFIG-ATOMIC-WRITE-001`; slice (c) (malformed-section writeback, gitignore merge) remains resident in the parent. |

## Split Lineage

This SPEC is **slice (a)** of the 3-way split of `SPEC-CONFIG-TIER-PERSIST-001`, recorded in that SPEC's §L Split Branches section. The split carving preserves the parent's exact tier-resolution semantics; the child REQ namespace (`REQ-CTR-*`) is a clean re-namespace of the parent's slice-(a) REQ namespace (`REQ-CTP-001..014`) and does NOT cross-reference parent REQ-CTP IDs in the body — the mapping is documented here for traceability only.

| Child REQ | Parent REQ | Parent section |
|-----------|-----------|----------------|
| REQ-CTR-001 | REQ-CTP-001 | §D.1 Tier merge semantics |
| REQ-CTR-002 | REQ-CTP-002 | §D.1 |
| REQ-CTR-003 | REQ-CTP-003 | §D.1 |
| REQ-CTR-004 | REQ-CTP-004 | §D.1 |
| REQ-CTR-005a | REQ-CTP-005a | §D.1 |
| REQ-CTR-005b | REQ-CTP-005b | §D.1 |
| REQ-CTR-006 | REQ-CTP-006 | §D.2 Tier ordering and local-tier reachability |
| REQ-CTR-007 | REQ-CTP-007 | §D.2 |
| REQ-CTR-008 | REQ-CTP-008 | §D.2 |
| REQ-CTR-009 | REQ-CTP-009 | §D.2 |
| REQ-CTR-010 | REQ-CTP-010 | §D.2 |
| REQ-CTR-011 | REQ-CTP-011 | §D.2 |
| REQ-CTR-012 | REQ-CTP-012 | §D.2 |
| REQ-CTR-013 | REQ-CTP-013 | §D.2 (already-satisfied regression guard) |
| REQ-CTR-014 | REQ-CTP-014 | §D.2 (already-satisfied regression guard) |

The parent's `depends_on: [SPEC-UPDATE-DATA-SURVIVAL-001]` does NOT carry over to this child: slice-(a) tier resolution is independent of the update data-survival concern (the parent's `depends_on` was driven by slice (c) — backup coverage / write atomicity — not by tier resolution). `related_specs:` carries a bi-directional link to the parent and to the slice-(b) sibling for traceability.

## §A Problem / Motivation

The `.moai/config` resolution subsystem — `internal/config/merge.go` `MergeAll` plus the `Source` enum in `internal/config/source.go` — implements an 8-tier priority system. Two independent defects make it unable to express the thing a configuration system exists to express, and a third makes the documented local-tier opt-in inert. All three were identified and reproduced by the parent SPEC's four-lens audit; this child owns their fix end-to-end.

**Defect 1 — falsey values cannot win their tier.** `MergeAll` walks tiers in priority order and skips any value that is the zero value for its type (`internal/config/merge.go:149-152`, `if isZero(value) { continue }`). The consequence is that nothing can ever be turned **off** by a higher tier: an explicit `false`, `0`, `""`, or empty list loses to whatever the next tier down says, and ultimately to `SrcBuiltin`. Reproduced (parent SPEC §A, against HEAD `d5336214e`) with `SrcProject{enabled:false, count:0, name:""}` layered over `SrcBuiltin{enabled:true, count:42, name:"default"}`:

```
workflow.branch_guard.enabled    => true ok=true source=builtin
x.count                          => 42 ok=true source=builtin
x.name                           => "default" ok=true source=builtin
```

An organisation policy tier that sets `enabled: false` is therefore silently ignored — and worse, the `PolicyOverrideRejected` enforcement at `merge.go:155-162` is nested inside `if winningValue != nil`, so when the policy value is falsey there is no winning value, no rejection is raised, and `strict_mode` provides no protection at all for exactly the policies most likely to be restrictive.

**Defect 2 — `SrcLocal` is ordered below `SrcProject` at every site that expresses tier order.** `internal/config/source.go:26-28` documents `SrcLocal` as "local overrides (not committed to git)" pointing at `.claude/settings.local.json` and `.moai/config/local/*.yaml`, but every one of the three places that expresses tier order puts `SrcLocal` *after* `SrcProject`. An override that cannot override is not an override. Reproduced:

```
k => from-project (source=project) [project=2 local=3]
```

Which of the three places actually *decides* the outcome is not obvious, and the parent SPEC's v0.1.0 got it wrong. `Priority()` (`internal/config/source.go:73`) returns `int(s)`, the enum integer — but it is read by nothing outside tests. Measured on the baseline tree:

```
$ grep -rn '\.Priority()' --include='*.go' internal/ | grep -v '_test.go'
(no output, exit 1)
```

Resolution order is decided instead by two independent hand-written literal slices: `AllSources()` (`internal/config/source.go:103-114`), which `MergeAll` walks at `internal/config/merge.go:137`, and a second `tiers` slice inside `internal/permission/resolver.go:225-234`. Both list `SrcProject` before `SrcLocal`. Correcting the `iota` block alone would therefore change no resolution anywhere while breaking `TestAllSources`; REQ-CTR-006 is scoped to all three sites accordingly.

**Defect 3 — the typed `Loader` never consults the local tier.** `CLAUDE.local.md` §22.9 instructs maintainers to opt into the branch guard by creating `.moai/config/local/workflow.yaml`, but the branch guard reads through `hook.ConfigProvider.Get()` → `config.Loader.Load(configDir)`, and `grep -n "config/local\|localDir\|SrcLocal" internal/config/loader.go` returns **zero matches**. The 8-tier `resolver` (`internal/config/resolver.go:386-390`, `loadLocalTier`) and the typed `Loader` are two systems that never meet. A probe writing `sections/workflow.yaml`=false and `local/workflow.yaml`=true observed `BranchGuard.Enabled = false` — the documented opt-in is inert.

These three defects compose. Wiring the local tier (Defect 3) while falsey values are still skipped (Defect 1) would produce an opt-in that cannot opt out — a local tier that can only turn things on is a new asymmetry and must not ship. REQ-CTR-012 therefore orders the milestones so the merge-semantics fix lands before the local-tier wiring.

## §B Goals

- An explicit falsey value set by a higher tier wins its tier, and `strict_mode` policy rejection becomes reachable for falsey policy values.
- The `SrcLocal` tier means what its own documentation says, and is reachable from the typed `Loader` path so the documented `.moai/config/local/` opt-in actually works.
- The local directory can both enable AND disable a setting; a local tier that can only turn things on does not ship.

## §C Scope Exclusions

### Out of Scope — atomic, mode-preserving writes

- The shared atomic-write helper, the `os.CreateTemp` mode-narrowing fix, the hardcoded-`0o644` remediation, and the mode-widening migration. All owned by the slice-(b) child `SPEC-CONFIG-ATOMIC-WRITE-001` (status: closed). This child owns *which tier wins when sources disagree*; the slice-(b) child owns *how the result is written to disk*. The two slices are entirely orthogonal.

### Out of Scope — malformed-configuration contract

- Distinguishing absent vs malformed section files, refusing to write compiled defaults back over a user's malformed-but-recoverable file, and the CLI/hook split for surfacing malformed sections. Owned by slice (c), still resident in the parent `SPEC-CONFIG-TIER-PERSIST-001` §D.3. This child's merge-semantics change (REQ-CTR-001..005b) touches `MergeAll`'s value-selection path, not its error-handling path; the two are separable and slice (c) proceeds independently.

### Out of Scope — `.gitignore` merge correctness

- `MergeGitignoreFile` header parsing, template-dropped-pattern migration, deduplication. Owned by slice (c), resident in the parent §D.5. This child touches no `internal/cli/update/merge/` code.

### Out of Scope — fallback visibility

- 3-way merge failure advisory, absent-base advisory, and the declared reversal of `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007`. Owned by slice (c), resident in the parent §D.6 and reconciled in the parent §K. This child has no update-path footprint.

### Out of Scope — the SrcUser / SrcProject relative order

- `SrcUser` currently outranks `SrcProject`, which is unusual for a project-scoped tool. No audit lens produced evidence of a defect arising from it, and no requirement here depends on it. Re-litigating that pair is deliberately excluded so the enum change stays minimal and its blast radius stays assessable. Recorded as an observation, not a finding.

### Out of Scope — template content

- Any edit under `internal/template/templates/**`. That tree is distributed across 16 languages and must carry no SPEC ID, REQ token, internal date, or commit SHA. All fixtures live in `_test.go` files created with `t.TempDir()`.

### Out of Scope — dead config keys, CI gates, documentation drift

- Owned by sibling Epic SPECs 4 through 6 (per the parent §C).

## §D Requirements (GEARS)

### D.1 Tier merge semantics

- **REQ-CTR-001** — `MergeAll` shall distinguish *a key being present in a tier* from *that key's value being the zero value for its type*, and shall select the first tier in priority order in which the key is present, regardless of whether that value is falsey.

- **REQ-CTR-002** — **When** a higher-priority tier supplies an explicit `false`, `0`, `""`, or empty collection for a key, `MergeAll` shall record that value as the winning value with that tier as its `Provenance.Source`.

- **REQ-CTR-003** — **While** `policy.strict_mode` is enabled, `MergeAll` shall raise `PolicyOverrideRejected` for a lower tier attempting to override a policy-designated key whose policy value is falsey, on the same terms as a non-falsey policy value.

- **REQ-CTR-004** — `MergeAll` shall continue to omit from its output any key that is absent from every tier; absence and falsey-presence shall remain distinguishable in the merged result.

- **REQ-CTR-005a** — **Where** `isZero` retains at least one caller after this change whose semantics genuinely require zero-detection, the helper shall be retained together with that caller.

- **REQ-CTR-005b** — **Where** `isZero` retains no caller after this change, the helper shall be removed rather than left unreferenced.

  > REQ-CTR-005a/b together resolve the single-REQ self-contradiction that a naively split SPEC can reintroduce (a clause that simultaneously forbids and mandates deletion — "shall not be deleted … it shall be removed"). The intent is a disposition decision conditioned on caller survival; the two clauses state that without self-contradiction. Covered by AC-CTR-006.

### D.2 Tier ordering and local-tier reachability

- **REQ-CTR-006** — Tier order shall place `SrcLocal` above `SrcProject` at **all three** sites that express it, in a single commit: the `iota` block in `internal/config/source.go` (so that `SrcLocal.Priority() < SrcProject.Priority()`), the `AllSources()` literal slice (`internal/config/source.go:103-114`, which `internal/config/merge.go:137` walks), and the `tiers` literal slice in `internal/permission/resolver.go:225-234`. Reordering fewer than all three shall be treated as an incomplete change.

  > The `iota` block alone does not decide resolution. `Priority()` (`source.go:73`) returns `int(s)` and has zero non-test consumers on the baseline tree, so an `iota`-only reorder leaves both literal slices — and therefore both `MergeAll` and the permission resolver — resolving in the old order, while breaking `TestAllSources`'s `AllSources()[i].Priority() == i` invariant. Reordering the `iota` block and `AllSources()` **together** restores that invariant, so `TestAllSources` needs no edit; `TestSourceOrdering`'s `expectedOrder` literal does (REQ-CTR-008).

- **REQ-CTR-007** — The relative order of every other pair of tiers shall be unchanged, and `SrcPolicy` shall remain the numerically highest-priority tier.

- **REQ-CTR-008** — The existing `TestSourceOrdering` guard (`internal/config/source_test.go:104-119`) already pins the complete tier ordering against a literal, hand-written sequence. Its `expectedOrder` literal shall be updated to the new order **in the same commit** as REQ-CTR-006 and shall be retained; it shall not be deleted, weakened, or replaced by an enum-derived comparison.

  > The guard already exists and is green — the observed `expectedOrder` literal is `[SrcPolicy, SrcUser, SrcProject, SrcLocal, SrcPlugin, SrcSkill, SrcSession, SrcBuiltin]`. The real obligation is to update its literal rather than to author a duplicate. `TestAllSources` (`:89-102`) is a separate, complementary invariant (`AllSources()[i].Priority() == i`) and needs no edit under REQ-CTR-006's three-site reorder.

- **REQ-CTR-009** — `internal/permission/resolver.go` shall be shown to resolve permissions correctly under the new ordering. It does **not** iterate the `Source` enum: it carries its own literal `tiers` slice at `:225-234`, so REQ-CTR-006's reorder reaches it only because that slice is one of the three named sites, and the assertion is therefore load-bearing rather than confirmatory.

- **REQ-CTR-010** — The typed `Loader` shall read `.moai/config/local/*.yaml` and apply it above `.moai/config/sections/*.yaml`, so that a value set in the local directory overrides the same value set in the sections directory.

- **REQ-CTR-011** — **When** `.moai/config/local/` is absent, the `Loader` shall behave exactly as it does today, with no warning and no error.

- **REQ-CTR-012** — The `Loader`'s local-tier read shall be applied **after** REQ-CTR-001 is in force, so that the local directory can both enable and disable a setting. A local tier that can only turn things on is a new asymmetry and shall not be shipped.

- **REQ-CTR-013** — **ALREADY SATISFIED at `b9fc75016`; retained as a regression guard only.** `.moai/config/local/` shall remain present in the repository `.gitignore`, so that a maintainer-local opt-in cannot be committed and thereby flip a distributed default. No milestone work item follows from this requirement; AC-CTR-013 asserts the entry has not been removed.

  > Measured (parent SPEC §D.2): `git check-ignore -v .moai/config/local/workflow.yaml` exits 0 and prints `.gitignore:183:.moai/config/local/`. The commit that added it, `b9fc75016`, is an ancestor of the `d5336214e` baseline (`git merge-base --is-ancestor b9fc75016 d5336214e` → exit 0), so the work was already complete when the parent SPEC was drafted.

- **REQ-CTR-014** — **ALREADY SATISFIED; retained as a regression guard only.** `CLAUDE.local.md` §22.9 shall continue to carry no `BLOCKER (gitignore)` note and shall continue to state that the documented opt-in is effective. No milestone work item follows from this requirement; AC-CTR-014 asserts the note has not been reintroduced.

  > Measured (parent SPEC §D.2): `grep -c 'BLOCKER (gitignore)' CLAUDE.local.md` → `0`. §22.9 already reads "**gitignore (해결됨)**: `.moai/config/local/` 디렉터리는 이제 `.gitignore`에 등록되었다".

## §E Non-Functional Constraints

- **NFR-CTR-001** — Every test creates its temporary directories with `t.TempDir()` and touches no path outside them. No test reads or writes the operator's real `~/.moai`, `~/.claude`, or the repository's own `.moai/config/`.
- **NFR-CTR-002** — No test sets an `OTEL_*` environment variable, per the parallel-test data-race hazard recorded in `CLAUDE.local.md` §2.
- **NFR-CTR-003** — Go conventions hold: `snake_case.go` file names, `fmt.Errorf("...: %w", err)` error wrapping, English comments and godoc.
- **NFR-CTR-004** — No edit under `internal/template/templates/**`.
- **NFR-CTR-005** — The tier-ordering change is a single commit that touches the enum and its guards only, so it can be reverted independently of the merge-semantics change.
- **NFR-CTR-006** — The merge-semantics change ships in one revertable commit, so a previously-ignored false that disables a depended-on feature can be reverted without unwinding the local-tier wiring (M3) or the ordering change (M2).

## §F Success Criteria

- An explicit `false` from `SrcProject` beats `true` from `SrcBuiltin`, verified by a test that reproduces the §A probe and now observes `source=project`.
- `SrcLocal.Priority() < SrcProject.Priority()`, AND `.moai/config/local/workflow.yaml` setting `branch_guard.enabled: true` is observed through `Loader.Load` when `sections/workflow.yaml` says `false` — AND the reverse also holds (local can turn a setting OFF).
- `go test ./...` and `golangci-lint run` are clean.

## §G Risk Register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------:|--------|------------|
| R1 | A previously-ignored `false` starts winning and disables a feature that users depend on being on | High | High | Enumerate every falsey value present in any shipped `sections/*.yaml` and in `internal/config/defaults.go` before landing REQ-CTR-001; the enumeration is a milestone deliverable (M0), not a footnote. Ship the merge-semantics change alone in one revertable commit (NFR-CTR-006) so it can be reverted without unwinding M2/M3. |
| R2 | A previously-ignored `""` starts winning and blanks a string that had a working default | Medium | High | Same enumeration as R1; empty strings are the subtlest case because a blank name or path often fails far from the config layer. |
| R3 | The reorder **fails to reach** the production path: `Priority()` has zero non-test consumers, so an `iota`-only edit changes no resolution while breaking `TestAllSources` | High | High | REQ-CTR-006 names all three ordering sites explicitly (`iota`, `AllSources()`, `resolver.go` `tiers`); AC-CTR-005 asserts the merge outcome rather than the ordinal alone; AC-CTR-007 asserts permission resolution directly. The measured hazard is under-reach, not over-reach. |
| R4 | Wiring the local tier while zero-values are still skipped produces an opt-in that cannot opt out | Medium | Medium | REQ-CTR-012 orders the milestones so this cannot happen: M1 (merge semantics) lands before M3 (local-tier wiring). AC-CTR-010 asserts the disable direction explicitly. |

## §H Cross-References

- `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/` — the parent SPEC. This child's full lineage and REQ mapping live in the parent's §L Split Branches and in this child's § Split Lineage above.
- `.moai/specs/SPEC-CONFIG-ATOMIC-WRITE-001/` — slice-(b) sibling (status: closed). Owns the atomic, mode-preserving write helper. This child's tier resolution is orthogonal to write mechanics; the two slices proceed independently.
- `SPEC-V3R2-RT-005` — the 8-tier system this SPEC corrects; `source.go:5` `@MX:ANCHOR` records `fan_in=71`.
- `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001` — the branch-guard opt-in whose documented mechanism REQ-CTR-010 makes functional.
- `CLAUDE.local.md` §22.9 — the maintainer opt-in doctrine. Its former `BLOCKER (gitignore)` note is already cleared (REQ-CTR-014 is a regression guard, not a work item).
