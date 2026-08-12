---
id: SPEC-CONFIG-SECTION-INTEGRITY-001
title: "config section integrity: a malformed section must never be silently written back as defaults, MergeGitignoreFile must preserve user patterns across template updates, and a 3-way→2-way fallback must be visible on first occurrence"
version: "0.1.0"
status: draft
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P0
phase: "v3.0.2"
module: "internal/config, internal/cli/update"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "config, malformed-section, writeback-guard, gitignore, merge, fallback-visibility, data-integrity"
related_specs: [SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-ATOMIC-WRITE-001, SPEC-CONFIG-TIER-RESOLVE-001, SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001, SPEC-V3R6-UPDATE-NOISE-001]
---

# SPEC-CONFIG-SECTION-INTEGRITY-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-12 | Initial draft. Carved out of the over-large parent `SPEC-CONFIG-TIER-PERSIST-001` (35-REQ, Tier-M-exceeding) as **slice (c)** of the 3-way split recorded in the parent's §L Split Branches. Re-namespaces the parent's §D.3 "Malformed-configuration contract" (REQ-CTP-015..020), §D.5 ".gitignore merge correctness" (REQ-CTP-028..032), and §D.6 "Fallback visibility" (REQ-CTP-033..035) into child REQ-CSI-001..014, preserving the parent's exact semantic content verbatim. Migrates the parent's §K Reconciliation with `SPEC-V3R6-UPDATE-NOISE-001` into this child's §K verbatim — REQ-CSI-012's declared supersession of `REQ-UN-007`'s silent-first-occurrence clause is unlandable without that reconciliation content in place. Slice (a) (tier resolution) was authored as the sibling `SPEC-CONFIG-TIER-RESOLVE-001`; slice (b) (atomic, mode-preserving writes) is already CLOSED as `SPEC-CONFIG-ATOMIC-WRITE-001`. |

## Split Lineage

This SPEC is **slice (c)** of the 3-way split of `SPEC-CONFIG-TIER-PERSIST-001`, recorded in that SPEC's §L Split Branches section. The split carving preserves the parent's exact slice-(c) semantics; the child REQ namespace (`REQ-CSI-*`) is a clean re-namespace of the parent's slice-(c) REQ namespace (`REQ-CTP-015..020`, `REQ-CTP-028..032`, `REQ-CTP-033..035`) and does NOT cross-reference parent REQ-CTP IDs in the body — the mapping is documented here for traceability only.

| Child REQ | Parent REQ | Parent section |
|-----------|-----------|----------------|
| REQ-CSI-001 | REQ-CTP-015 | §D.3 Malformed-configuration contract |
| REQ-CSI-002 | REQ-CTP-016 | §D.3 |
| REQ-CSI-003 | REQ-CTP-017 | §D.3 |
| REQ-CSI-004 | REQ-CTP-018 | §D.3 |
| REQ-CSI-005 | REQ-CTP-019 | §D.3 |
| REQ-CSI-006 | REQ-CTP-020 | §D.3 |
| REQ-CSI-007 | REQ-CTP-028 | §D.5 .gitignore merge correctness |
| REQ-CSI-008 | REQ-CTP-029 | §D.5 |
| REQ-CSI-009 | REQ-CTP-030 | §D.5 |
| REQ-CSI-010 | REQ-CTP-031 | §D.5 |
| REQ-CSI-011 | REQ-CTP-032 | §D.5 |
| REQ-CSI-012 | REQ-CTP-033 | §D.6 Fallback visibility (supersedes `REQ-UN-007` — see §K) |
| REQ-CSI-013 | REQ-CTP-034 | §D.6 |
| REQ-CSI-014 | REQ-CTP-035 | §D.6 |

The parent's `depends_on: [SPEC-UPDATE-DATA-SURVIVAL-001]` carries over to this child: slice (c) owns `.moai/config/**` writers in `internal/cli/update/backup/restore.go` and `internal/cli/update/merge/merge.go`, whose atomicity and backup coverage is the update data-survival concern. `related_specs:` additionally carries a bi-directional link to the parent, to the slice-(a) and slice-(b) siblings, and to `SPEC-V3R6-UPDATE-NOISE-001` (the implemented SPEC whose `REQ-UN-007` silent-first-occurrence clause this child's REQ-CSI-012 declares a supersession of — see §K).

## §A Problem / Motivation

The `.moai/config` subsystem loses user data in three distinct ways. All three were identified and reproduced by the parent SPEC's four-lens audit; this child owns their fix end-to-end under one invariant: **bad data must never be silently written back as defaults, and silent fallback must be visible.**

**Defect 1 — a malformed section file is silently converted into compiled defaults that are then written back over the user's data.** Every one of the fourteen section loaders in `internal/config/loader.go` swallows a load failure into compiled defaults (`loader.go:121-131` and thirteen siblings, all `slog.Warn(... "using defaults")`). A hook reading configuration while `moai update` truncates a section file observes a malformed file, gets builtin defaults for that turn, and surfaces nothing the operator can see. If a `ConfigManager.Save()` follows in the same process, those builtin defaults are then written back over the user's real configuration — a silent, permanent read-modify-write destruction of user data. The malformed file is indistinguishable from an absent file in today's loader: both return early after `slog.Warn` and neither records the failure.

Reproduced (parent SPEC §A): a truncated `user.yaml` causes a hook to read builtin defaults silently, and a subsequent `Save()` overwrites the recoverable malformed file with those defaults. The narrowing-files defect from slice (b) makes the same failure reachable from a different direction: a container or CI step running as a different UID can no longer read a 0600 section file, and the resulting read failure is swallowed identically.

**Defect 2 — `MergeGitignoreFile` reclassifies template-dropped patterns as user-authored and preserves duplicates verbatim.** `MergeGitignoreFile` (`internal/cli/update/merge/merge.go:58`) classifies every backup line absent from the *new* template as user-authored, so a pattern the template **drops** between releases is migrated into the `# User Custom Patterns` section and stays there forever. Duplicates are preserved verbatim because nothing dedupes. The header is emitted as a comment and skipped as a comment on the next pass, so the file does not grow without bound — but the reclassification is permanent. Reproduced with a template that drops `DS_Store` between updates (parent SPEC §A):

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

**Defect 3 — the more serious of the two merge failure paths is the quieter one.** In `internal/cli/update/backup/restore.go`, a **3-way** merge failure calls `recordFallback` and falls through silently (`restore.go:131-134`) — the noise-suppression ledger stays quiet until three consecutive failures — while a **2-way** failure prints an immediate stderr warning (`restore.go:139-141`). A third path is quieter still: when the base file is absent, `os.ReadFile(basePath)` errors and control falls to 2-way with neither a `recordFallback` call nor a warning. This matters today, not hypothetically: `quality.yaml` takes the 3-way-fails path on every update (recorded in `SPEC-UPDATE-YAML-PRESERVE-001`'s `audit.md`), so users are silently receiving 2-way merges while believing they receive 3-way.

## §B Goals

- A malformed section file is never silently converted into compiled defaults that are then written back over the user's data; the writeback refusal is scoped to the failed section alone so one bad file does not block five good ones.
- A hook stays fail-open on a malformed section but emits an operator-visible advisory; a CLI command surfaces the malformed section as an error and exits non-zero. Absent files continue to yield compiled defaults silently.
- `.gitignore` merging stops reclassifying template-dropped patterns as user-authored, stops preserving duplicates, and remains idempotent and safe for pre-header backups.
- The 3-way merge fallback is at least as visible as the 2-way fallback on its first occurrence; the noise-suppression ledger continues to govern repetition, not first occurrence.

## §C Scope Exclusions

### Out of Scope — tier resolution, falsey-wins, local-tier reachability

- `MergeAll` value-selection (zero-skip), `SrcLocal` ordering across the three sites, the typed-Loader / resolver disconnect, and the branch-guard local-tier opt-in. All owned by the slice-(a) sibling `SPEC-CONFIG-TIER-RESOLVE-001`. This child's malformed-section handling touches the loader's error-handling path, not its value-selection path; the two are separable.

### Out of Scope — atomic, mode-preserving writes

- The shared atomic-write helper, the `os.CreateTemp` mode-narrowing fix, the hardcoded-`0o644` remediation, the no-bare-`os.WriteFile` guard, and the mode-widening migration. All owned by the slice-(b) sibling `SPEC-CONFIG-ATOMIC-WRITE-001` (status: closed). This child's writers in `internal/cli/update/backup/restore.go` and `internal/cli/update/merge/merge.go` MAY adopt the slice-(b) helper for their own `.moai/config/**` writes; the helper's API is locked at slice-(b) `plan.md` M1 so this child is unblocked. This child owns *which bytes get written and whether the write is refused*; the slice-(b) child owns *how the write reaches disk atomically*.

### Out of Scope — the YAML merge engine internals

- `DeepMerge3Way`, `ValuesEqual`, `systemFields`, sequence handling, and the old-only-key drop inside `internal/cli/update/backup/merge.go`. All owned by `SPEC-UPDATE-YAML-PRESERVE-001`. The parent SPEC records four **amendment requests** against that SPEC (parent §I, A1-A4) with evidence; this child inherits none of that scope and specifies no fix for any of them. REQ-CSI-012/013/014 govern only the *visibility* of the fallback path that calls into that engine, not the engine itself.

### Out of Scope — 3-way merge base provenance

- The fact that the 3-way merge base is derived from the NEW template, so `base ≡ new`. Owned by the deferred `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`. The parent SPEC's §J carries forward the concrete release-tag evidence for that SPEC; this child changes no base-selection code.

### Out of Scope — backup coverage and the update failure contract

- Whether a backup exists on disk before a destructive step, the recovery manifest, and the restore entry point. Owned by `SPEC-UPDATE-DATA-SURVIVAL-001`, which the parent (and this child via `related_specs`) declares in its dependency footprint. The boundary is precise: that SPEC owns *that the bytes reach disk before deletion*; slice (b) owns *that the write itself is atomic and preserves mode*; this child owns *that a malformed section is not written back as defaults and that the silent fallback is visible*.

### Out of Scope — v2 detection and deprecated-path handling

- Semver-normalised v2/v3 detection, `defs.DeprecatedPaths` membership, and residue cleanup. Owned by `SPEC-UPDATE-REINSTALL-LOOP-002`.

### Out of Scope — template content

- Any edit under `internal/template/templates/**`. That tree is distributed across 16 languages and must carry no SPEC ID, REQ token, internal date, or commit SHA. All fixtures live in `_test.go` files created with `t.TempDir()`.

### Out of Scope — dead config keys, CI gates, documentation drift

- Owned by sibling Epic SPECs 4 through 6 (per the parent §C).

## §D Requirements (GEARS)

### D.1 Malformed-configuration contract

- **REQ-CSI-001** — The section loaders shall distinguish an **absent** section file from a **malformed** one. An absent file shall continue to yield compiled defaults silently.

- **REQ-CSI-002** — **When** a section file is present but fails to parse, the `Loader` shall record that section as failed, and shall not mark it loaded.

- **REQ-CSI-003** — **Where** the caller is a CLI command, a malformed section file shall surface as a returned error naming the file and the parse failure, and the command shall exit non-zero.

- **REQ-CSI-004** — **Where** the caller is a hook, a malformed section file shall not abort the hook; the hook shall continue with compiled defaults for that section and shall emit an operator-visible advisory on stderr naming the file, in addition to the existing `slog.Warn`. This preserves the fail-open norm that `.claude/rules/moai/workflow/main-checkout-branch-guard.md` establishes for hooks.

- **REQ-CSI-005** — **When** any section failed to parse during the most recent `Load`, `ConfigManager.Save` shall refuse to persist that section and shall return an error naming it, rather than writing compiled defaults over the user's file.

- **REQ-CSI-006** — The refusal in REQ-CSI-005 shall be scoped to the failed section alone; sections that loaded cleanly shall still be persisted.

### D.2 `.gitignore` merge correctness

- **REQ-CSI-007** — `MergeGitignoreFile` shall parse the `# User Custom Patterns` header as a real section boundary, and shall treat only lines below that header in the backup as user-authored.

- **REQ-CSI-008** — **When** the new template drops a pattern that the previous template contained and the user never added, `MergeGitignoreFile` shall not migrate that pattern into the user section.

- **REQ-CSI-009** — `MergeGitignoreFile` shall emit each user pattern at most once, preserving the first occurrence's original text and discarding subsequent exact duplicates.

- **REQ-CSI-010** — `MergeGitignoreFile` shall remain idempotent: merging a file with its own previous output shall produce byte-identical content.

- **REQ-CSI-011** — **When** a backup predates the header convention and therefore contains no `# User Custom Patterns` header, `MergeGitignoreFile` shall fall back to the current not-in-template heuristic rather than discarding the user's patterns.

### D.3 Fallback visibility

- **REQ-CSI-012** — **When** a 3-way merge fails and the restore path falls back to 2-way, the operator shall receive a stderr advisory naming the file, on the first occurrence rather than after three. **This supersedes the silent-first-occurrence clause of `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007`, which is `status: implemented`.** The reversal, the alternative that was rejected, and the parts of `REQ-UN-007` that survive are declared in §K. Landing REQ-CSI-012 without the §K reconciliation is prohibited.

- **REQ-CSI-013** — **When** the 3-way base file is absent and control falls through to 2-way without a merge attempt, the operator shall receive an advisory distinguishable from the merge-failure advisory of REQ-CSI-012.

- **REQ-CSI-014** — The advisory of REQ-CSI-012 shall be at least as prominent as the existing 2-way failure warning at `restore.go:139-141`; the noise-suppression ledger shall continue to govern *repetition*, not *first occurrence*.

## §E Non-Functional Constraints

- **NFR-CSI-001** — Every test creates its temporary directories with `t.TempDir()` and touches no path outside them. No test reads or writes the operator's real `~/.moai`, `~/.claude`, or the repository's own `.moai/config/`.
- **NFR-CSI-002** — No test sets an `OTEL_*` environment variable, per the parallel-test data-race hazard recorded in `CLAUDE.local.md` §2.
- **NFR-CSI-003** — Go conventions hold: `snake_case.go` file names, `fmt.Errorf("...: %w", err)` error wrapping, English comments and godoc.
- **NFR-CSI-004** — No edit under `internal/template/templates/**`.
- **NFR-CSI-005** — The malformed-section writeback refusal (M1) ships in one revertable commit, so a workflow that currently limps along on defaults can be restored to the prior behaviour without unwinding the gitignore (M2) or fallback-visibility (M3) changes.
- **NFR-CSI-006** — The declared supersession of `REQ-UN-007` (§K) is explicit in the implementing commit body, not silent.

## §F Success Criteria

- A truncated section file causes a CLI error and a hook advisory, and a subsequent `Save()` refuses to overwrite that section.
- The refusal in the preceding bullet is scoped: with `user.yaml` malformed and `language.yaml` clean, a `Save()` that errors on `user.yaml` still wrote the updated `language.yaml`.
- The §A `.gitignore` probe, re-run after the fix, shows `DS_Store` absent from the user section and exactly one `build/`.
- A 3-way merge failure prints an advisory on its first occurrence, and a follow-up repeated failure does not produce an unbounded advisory stream.
- An absent 3-way base file emits a distinct advisory, distinguishable from the merge-failure advisory.
- `go test ./...` and `golangci-lint run` are clean.

## §G Risk Register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------:|--------|------------|
| R6 | Turning a malformed config into a CLI error breaks a workflow that currently limps along on defaults | Medium | Medium | REQ-CSI-001 keeps *absent* silent; only *malformed* is loud. Hooks stay fail-open per REQ-CSI-004. |
| R8 | Header-based `.gitignore` parsing discards patterns from a pre-header backup | Medium | High | REQ-CSI-011 requires the heuristic fallback when no header is present. |
| R9 | Making the 3-way fallback loud produces noise on every update while `quality.yaml` still fails 3-way — **and reverses `REQ-UN-007` of the implemented `SPEC-V3R6-UPDATE-NOISE-001`, whose whole purpose was to suppress exactly this output** | High | Medium | Accepted and intentional, but only as a declared supersession: §K states the reversal, records the rejected `--verbose` alternative, and preserves the counter and 3-strike advisory that `REQ-UN-007`/`REQ-UN-008` also own. `SPEC-UPDATE-YAML-PRESERVE-001` removes the underlying cause; until then operators should see it. REQ-CSI-014 keeps the ledger governing repetition. |

## §K Reconciliation with `SPEC-V3R6-UPDATE-NOISE-001` (declared reversal)

> Migrated verbatim from the parent SPEC's §K. REQ-CSI-012 (the re-namespaced REQ-CTP-033) commands the reversal; this section is the single load-bearing non-obvious content that makes REQ-CSI-012 landable. Without this reconciliation content in place in THIS child SPEC, REQ-CSI-012 is UNLANDABLE.

`SPEC-V3R6-UPDATE-NOISE-001` is `status: implemented`. Its `REQ-UN-007` reads, verbatim:

> `mergeYAML3Way` 가 실패하여 2-way fallback 으로 전환할 때, 시스템은 merge-history ledger 의
> `fallback_count` 를 1 증가시켜야 한다. `fallback_count` 가 3 미만이면 stderr 로 어떤 출력도
> emit 하지 않는다 (silent fallback).

REQ-CSI-012 commands the opposite of that second sentence. The implementation carries the constant (`internal/cli/update_noise.go:37`, `const fallbackAdvisoryThreshold = 3`) and the call site cites the requirement by name (`internal/cli/update/backup/restore.go`, the comment `// REQ-UN-007/008/010: emit advisory or legacy warning per noise-suppression policy.` immediately above the `recordFallback(projectRoot, relPath, false, os.Stderr)` call). The parent SPEC's v0.1.0 named neither the sibling SPEC nor the requirement anywhere — `grep -rn 'UPDATE-NOISE\|REQ-UN-'` over the parent SPEC directory returned no matches. Reversing an implemented sibling requirement in silence is the defect this section exists to close.

### K.1 — the alternative that was rejected

`REQ-UN-010` already provides a diagnostic escape hatch: `--verbose` bypasses the silent branch and restores the per-occurrence `Warning: 3-way merge failed for <relPath>, falling back to 2-way` message. Satisfying this SPEC's fallback-visibility concern through `--verbose` alone would require no reversal at all.

It is rejected because the two requirements answer different questions. `--verbose` serves an operator who **already suspects** something is wrong and is opting into detail. The §A finding is that the operator does not suspect anything: `quality.yaml` takes the 3-way-fails path on *every* update, and the operator's belief that they are receiving a 3-way merge is never contradicted. A flag that must be passed by someone who already knows to pass it cannot correct a belief held by someone who does not. The information has to arrive unasked at least once.

### K.2 — what is superseded, and what survives

| `REQ-UN-*` clause | Disposition under REQ-CSI-012 |
|---|---|
| `REQ-UN-007` — increment `fallback_count` on every 3-way failure | **Survives unchanged.** The counter is not the noise. |
| `REQ-UN-007` — emit nothing to stderr while `fallback_count < 3` | **Superseded.** The first occurrence becomes visible. |
| `REQ-UN-008` — advisory at the 3-strike threshold | **Survives.** REQ-CSI-014 keeps the ledger governing repetition; the threshold advisory still marks sustained failure. |
| `REQ-UN-009` — reset the counter on 3-way success | **Survives unchanged.** |
| `REQ-UN-010` — `--verbose` restores the legacy per-occurrence warning | **Survives.** It becomes redundant with the new first-occurrence advisory for the *first* failure and remains the only source of per-occurrence output for failures 2 and beyond. |

### K.3 — obligations this places on the run phase

1. The commit implementing REQ-CSI-012 shall state in its body that it supersedes `REQ-UN-007`'s silent-first-occurrence clause and shall reference this section.
2. `SPEC-V3R6-UPDATE-NOISE-001`'s own artifacts are **not** edited by this SPEC. Recording the supersession in that SPEC is a sync-phase concern for its owner, surfaced here as a cross-SPEC note rather than performed unilaterally.
3. If the reversal is rejected during Implementation Kickoff Approval, REQ-CSI-012 and AC-CSI-012 are withdrawn together and M3 reduces to REQ-CSI-013 (the absent-base advisory), which conflicts with nothing — the absent-base path today emits neither a `recordFallback` call nor a warning, so making it visible adds a signal rather than reversing a suppression.

## §H Cross-References

- `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/` — the parent SPEC. This child's full lineage and REQ mapping live in the parent's §L Split Branches and in this child's § Split Lineage above.
- `.moai/specs/SPEC-CONFIG-ATOMIC-WRITE-001/` — slice-(b) sibling (status: closed). Owns the atomic, mode-preserving write helper. This child's `.moai/config/**` writers MAY adopt that helper; the helper's API is locked at the sibling's `plan.md` M1.
- `.moai/specs/SPEC-CONFIG-TIER-RESOLVE-001/` — slice-(a) sibling. Owns tier resolution; orthogonal to this child's section-integrity concern.
- `.moai/specs/SPEC-UPDATE-YAML-PRESERVE-001/` — YAML merge engine internals. The parent records four amendment requests (§I A1-A4) against it; this child specifies no fix for any of them. REQ-CSI-012/013/014 govern only the *visibility* of the fallback path that calls into that engine.
- `.moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/` — backup coverage and the update failure contract; this child's writers depend on its backup-coverage guarantee.
- `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` — deferred; owns the 3-way base provenance (`base ≡ new`).
- `.moai/specs/SPEC-V3R6-UPDATE-NOISE-001/` — `status: implemented`. Its `REQ-UN-007` silent-first-occurrence clause is superseded by REQ-CSI-012; see §K for the declared reversal, the rejected `--verbose` alternative, and the clauses that survive.
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — the hook fail-open norm that REQ-CSI-004 preserves.
