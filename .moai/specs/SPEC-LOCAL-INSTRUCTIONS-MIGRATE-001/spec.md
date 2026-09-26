---
id: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
title: "Local-instruction migration — CLAUDE.local.md to AGENTS.local.md, with advisories and docs"
version: "0.1.0"
status: draft
priority: P1
phase: "v3.3.0 target"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
module: "internal/cli, docs-site/content"
lifecycle: spec-anchored
tags: "instruction-files, codex, migration, advisory, docs-site"
tier: M
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-26 | **Created by carve from `SPEC-INSTRUCTION-FILES-UNIFY-001` at commit `1140bcd1d`, by operator decision (via the lead, 2026-09-26).** Nine requirements — `REQ-IFU-007~012` and `REQ-IFU-020~022`, everything in that SPEC touching a **user-owned file** — and the seven acceptance criteria covering them are transferred here **verbatim**. Nothing was dropped and nothing was renumbered: the ids keep their `IFU` infix so every existing cross-reference, traceability row, and audit citation still resolves. Card **t1259** owns this SPEC's plan phase; the artifacts here exist to preserve the authored clauses and give that card real coordinates, and are **not a completed plan phase**. Repairs applied during the carve, both from the plan-audit of `1140bcd1d`: D3 (`REQ-IFU-021` now names which `CLAUDE.local.md` copy migrates, and states one unit throughout) and the frontmatter schema (canonical 12 fields, so this SPEC parses from birth). |

> **[HARD] What is NOT done here, and is t1259's to do.** This SPEC's plan phase is unfinished by
> design. Specifically open: the Tier judgment is provisional (`tier: M` on 9 requirements);
> `plan.md` carries the transferred milestones but no re-sequenced milestone plan; the two
> **recorded debts** in `acceptance.md` (`AC-IFU-011`, `AC-IFU-015`) are preserved as written and
> their unfold decision is t1259's — the Tier L ceiling that forced each fold no longer binds
> here (Tier M, 16/16, currently 9 requirements and 7 criteria), so the headroom exists, but
> spending it is a plan-phase decision this carve deliberately does not take; and no plan-audit
> has run against this SPEC.

---

## §A Context

`SPEC-INSTRUCTION-FILES-UNIFY-001` establishes the three-file instruction structure — a
harness-neutral `AGENTS.md`, a thin per-harness `CLAUDE.md`, and a user-owned
`AGENTS.local.md` — together with the guards and byte ceilings that hold it.

This SPEC owns the other half: everything that touches a file the **user** owns. Concretely,
the Codex-side `CLAUDE.local.md` fallback and its deprecation advisory, the explicit migration
verb, the no-coexistence invariant, the `moai update` / `moai doctor` advisories, the docs-site
rewrite across four locales, and this repository's own `CLAUDE.local.md` migration.

The discriminator is exactly that — **whether the work touches a user-owned file**. It is what
keeps the one milestone requiring explicit operator confirmation out of the branch that carries
the mechanical guards, and it is why the migration verb, which must never run implicitly, lives
apart from the guards that run on every build.

**The carve's cause was arithmetic.** The plan-audit of `1140bcd1d` found two requirements in
the parent SPEC covered by no acceptance criterion; the fix needed two new criteria, and that
SPEC stood at exactly 25 requirements and 25 criteria — both Tier L ceilings, with no tier above
L. Splitting was the only remedy that did not either leave a requirement untested or relax a
budget the tier rule forbids relaxing.

---

## §B Goal

Retire `CLAUDE.local.md` as a live read path — through an explicit operator-invoked migration,
never an implicit one — leaving exactly one user-owned local instruction file that both
harnesses read, with the transition documented in all four locales and performed on this
repository's own copy.

---

## §C Requirements (GEARS)

> Ids are transferred verbatim from `SPEC-INSTRUCTION-FILES-UNIFY-001` and retain the `IFU`
> infix. The gaps (`001~006`, `013~019`, `023~025`) are the carve's footprint: those
> requirements remain with the parent SPEC. Renumbering would break every cross-reference and
> the audit history that already cites these ids by name.

### C.1 Codex fallback and provenance

- **REQ-IFU-007** — Where only `CLAUDE.local.md` exists, the `moai codex` launcher shall
  read it as a fallback and shall emit a deprecation advisory naming the migration command.
- **REQ-IFU-008** — The Codex provenance preamble shall name the literal filename of the
  file it actually read, never a normalized or substituted name.

### C.2 Migration

- **REQ-IFU-009** — The system shall provide an explicit migration verb
  (`moai migrate local-instructions`) that moves `CLAUDE.local.md` content to
  `AGENTS.local.md` and relocates the original to a backup directory.
- **REQ-IFU-010** — `AGENTS.local.md` and `CLAUDE.local.md` shall not coexist as live read
  paths; the migration verb shall leave exactly one of the two in place.
- **REQ-IFU-011** — `moai update` shall not move, rename, or delete either local
  instruction file; where both exist, it shall emit an advisory only.
- **REQ-IFU-012** — `moai doctor` shall report the same advisory as `moai update` when both
  local instruction files exist or when only `CLAUDE.local.md` exists.

### C.3 Documentation and this repository

- **REQ-IFU-020** — The docs-site pages `claude-md-guide`, `codex-dual-harness`,
  `harness-learning`, `memory`, `quickstart`, and `update` shall describe the three-file
  structure in all four locales (ko, en, ja, zh).
- **REQ-IFU-021** — This repository's own `CLAUDE.local.md` — specifically **the copy
  committed on the `develop` branch**, which that file's own §0.1 declares canonical (the tree
  the lanes branch from) — shall migrate to `AGENTS.local.md` at **under 40,000 characters**,
  with operational procedure relocated to `.moai/docs/`.
- **REQ-IFU-022** — The migrated `AGENTS.local.md` §0 shall name `AGENTS.local.md` as the
  file whose canonical copy it discriminates.

> **[HARD] Which copy, and in which unit — the D3 repair.** Both halves of this were ambiguous
> in the parent SPEC, and the ambiguity was not incidental: that file's §0 exists precisely
> because its copies diverge.
>
> **The copy.** §0.1 declares the canonical copy to be the one on the tree the lanes branch
> from, currently `develop`; §0.2 states that an uncommitted working copy may never be cited as
> canonical; §0.3 records the copy committed on `main` as a retired third model. The migration
> target is therefore the `develop`-committed copy, and nothing else. Measured 2026-09-26:
> `git show origin/develop:CLAUDE.local.md | wc -m` → **44,381**, identical to this worktree's
> copy. (The plan-audit measured a third value, 39,258, from the primary checkout's *working*
> copy — which §0.2 excludes from citation and §0.4 documents as permanently modified by
> design. That reading is not the canonical copy, and this SPEC does not rest on it.)
>
> **The unit is characters throughout.** The parent SPEC mixed 61,908 **bytes** with a 44,381
> **character** count and a 40,000 cap, which made the reduction read as ~35% when against the
> canonical copy it is ~10%. The measuring command is `wc -m`, and the required reduction is
> 44,381 → under 40,000 characters, i.e. at least 4,381 characters. The canonical copy does
> **not** already satisfy the cap.

---

## §D Out of Scope

This SPEC deliberately does not build the following.

### Out of Scope — the contract and guard half

- The `AGENTS.md` contract body and its reconciliation, the `CLAUDE.md` thinning, the Codex
  read-**order** change, the frozen-instruction guard set, the curator Tier-3 target, the
  template neutrality allowlist, the two byte ceilings, and the import-resolution gate. Those
  are `SPEC-INSTRUCTION-FILES-UNIFY-001`'s (`REQ-IFU-001~006`, `013~019`, `023~025`). The one
  place the two SPECs touch the same function is the launcher's local-instruction loop: that
  SPEC changes its iteration order, this one adds the advisory on its fallback branch, and the
  two edits are ordered rather than concurrent.

### Out of Scope — automation

- Automatic renaming of `CLAUDE.local.md` by `moai update`, `moai init`, or any hook.
  Migration is an explicit operator-invoked verb only — that is the substance of REQ-IFU-011,
  not a stylistic preference.
- Automatic restoration or repair of a modified local instruction file. The file may carry
  runtime-written values, so an automatic restore is itself a destructive act.

### Out of Scope — the worktree duplicate-load

- Deduplicating the local-instruction load in a worktree session. Open card **t1219 item (1)**
  covers it. What this SPEC owes is only not making the duplicate worse: REQ-IFU-010's
  no-coexistence invariant is what keeps a worktree session from carrying up to four
  local-instruction loads.

### Out of Scope — adjacent work

- Rewriting the substance of any maintainer rule while relocating it. REQ-IFU-021 splits
  procedure out of a document; it does not re-decide what the document says.
- Retiring the `CLAUDE.local.md` filename from historical references. REQ-IFU-022 requires the
  migrated §0 to name the new file; occurrences of the old name that mark it as retired are
  expected to remain.

---

## §E Cross-references

- `SPEC-INSTRUCTION-FILES-UNIFY-001` — the parent SPEC this was carved from; holds the
  contract, the guards, and the ceilings.
- `.moai/reports/t1243/plan-audit-iter1.md` — the audit of `1140bcd1d` whose D2 arithmetic
  forced the carve and whose D3 finding this SPEC repairs.
- `.moai/reports/t1243/m0/verdict.md` — the M0 measurement the parent SPEC's design rests on.
- `internal/cli/codex_contract.go`, `internal/cli/codex_launcher.go` — the fallback branch and
  the provenance preamble.
- `internal/cli/` `migrate_agency_*` — the existing move-plus-backup verb precedent.
- `CLAUDE.local.md` §0 — the canonical-copy discriminant REQ-IFU-021 names.
