---
id: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
title: "Local-instruction migration — CLAUDE.local.md to AGENTS.local.md, with advisories and docs"
version: "0.2.0"
status: draft
priority: P1
phase: "v3.3.0 target"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
module: "internal/cli, docs-site/content"
lifecycle: spec-anchored
tags: "instruction-files, codex, migration, advisory, docs-site"
tier: L
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-26 | **Created by carve from `SPEC-INSTRUCTION-FILES-UNIFY-001` at commit `1140bcd1d`, by operator decision (via the lead, 2026-09-26).** Nine requirements — `REQ-IFU-007~012` and `REQ-IFU-020~022`, everything in that SPEC touching a **user-owned file** — and the seven acceptance criteria covering them are transferred here **verbatim**. Nothing was dropped and nothing was renumbered: the ids keep their `IFU` infix so every existing cross-reference, traceability row, and audit citation still resolves. Card **t1259** owns this SPEC's plan phase; the artifacts here exist to preserve the authored clauses and give that card real coordinates, and are **not a completed plan phase**. Repairs applied during the carve, both from the plan-audit of `1140bcd1d`: D3 (`REQ-IFU-021` now names which `CLAUDE.local.md` copy migrates, and states one unit throughout) and the frontmatter schema (canonical 12 fields, so this SPEC parses from birth). |
| 0.1.1 | 2026-09-26 | **Plan-audit iter-2 repairs (subject `653e53572`).** D14: `AC-IFU-007`'s two clauses were jointly unsatisfiable at their own boundary — a 4,381-character reduction from the measured 44,381 lands on exactly 40,000 and fails `< 40000`; floor corrected to 4,382, and the **same off-by-one repaired in `REQ-IFU-021` here**, which the audit did not name. D9 class (found by the cross-SPEC sweep, not in the audit's B2 note): `AC-IFU-011` arrived from the carve with the head-only pattern `'^TestCodexLocalInstructions'`, which matches 22 pre-existing sibling tests and made the criterion unfailable; both-end anchoring plus the asserted-line delimiter applied, the rule stated at the head of `acceptance.md`, and the advisory half declared as a test this SPEC creates. D12 class (also found by the cross-SPEC sweep, also absent from the audit's B2 note): sibling-SPEC `AC-` tokens in prose made this file's AC counter read **10** live criteria against 7 declared — `[REF]`-marked, counter now reports 7. Typo `discharegable` → `dischargeable`. Still **not a completed plan phase** — t1259 owns that. |
| 0.2.0 | 2026-09-26 | **Plan phase completed by card t1259.** Tier raised **M → L** on the measured file-count axis (35 files affected; the Tier L threshold is >15) — `design.md` and `research.md` are added accordingly. `REQ-IFU-010` split into two clauses, resolving the inherited `AC-IFU-014` contradiction (design.md §A). `REQ-IFU-020`'s `memory` page disambiguated: it resolved to **two** files per locale, and the criterion named neither. `REQ-IFU-021`'s before-value **re-measured in this tree: 44,740 characters, not the carve's 44,381** — and the fixed reduction floor is replaced with a derived one, because a hard-coded floor drifts with the file it measures. Both recorded debts unfolded (`AC-IFU-029`, `AC-IFU-030`), `AC-IFU-015` promoted to blocking, whole-change CI criterion `AC-IFU-031` authored. |

> **Plan phase complete (v0.2.0, card t1259).** Every item the carve left open is now either
> applied or explicitly deferred with a reason; `progress.md` §E.1 is the itemised record. The
> Tier judgment is taken (L, on measured file count), the milestones are re-sequenced for this
> SPEC's own dependency order, both recorded debts are unfolded, and the parent `research.md`
> Q4 is answered. What remains open by design: no plan-audit has run against this SPEC yet, and
> the `SPEC-INSTRUCTION-FILES-UNIFY-001` M2 sequencing dependency (plan.md §B) is unchanged.
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
  paths. Two clauses, because the unambiguous and the ambiguous case have different correct
  outcomes:
  - **REQ-IFU-010a (resolution).** Where exactly one of the two files is present, the migration
    verb shall leave exactly one of the two in place.
  - **REQ-IFU-010b (refusal).** Where **both** files are present, the migration verb shall
    refuse — exiting non-zero, modifying neither file, and naming the coexistence as the
    reason. It shall not choose between them.
- **REQ-IFU-011** — `moai update` shall not move, rename, or delete either local
  instruction file; where both exist, it shall emit an advisory only.
- **REQ-IFU-012** — `moai doctor` shall report the same advisory as `moai update` when both
  local instruction files exist or when only `CLAUDE.local.md` exists.

### C.3 Documentation and this repository

- **REQ-IFU-020** — The six docs-site pages named below shall describe the three-file
  structure in all four locales (ko, en, ja, zh), naming each by its **path** rather than its
  stem:
  `advanced/claude-md-guide.md`, `advanced/codex-dual-harness.md`,
  `advanced/harness-learning.md`, `claude-code/context-memory/memory.md`,
  `getting-started/quickstart.md`, `cli-reference/update.md`.
- **REQ-IFU-021** — This repository's own `CLAUDE.local.md` — specifically **the copy
  committed on the `develop` branch**, which that file's own §0.1 declares canonical (the tree
  the lanes branch from) — shall migrate to `AGENTS.local.md` at **under 40,000 characters**
  (`wc -m`, i.e. at most 39,999), with operational procedure relocated to `.moai/docs/`. The
  pre-migration value shall be **re-measured at the milestone** rather than carried from this
  document.
- **REQ-IFU-022** — The migrated `AGENTS.local.md` §0 shall name `AGENTS.local.md` as the
  file whose canonical copy it discriminates.

> **[HARD] Which copy, and in which unit — the D3 repair.** Both halves of this were ambiguous
> in the parent SPEC, and the ambiguity was not incidental: that file's §0 exists precisely
> because its copies diverge.
>
> **The copy.** §0.1 declares the canonical copy to be the one on the tree the lanes branch
> from, currently `develop`; §0.2 states that an uncommitted working copy may never be cited as
> canonical; §0.3 records the copy committed on `main` as a retired third model. The migration
> target is therefore the `develop`-committed copy, and nothing else. (The plan-audit measured a
> third value, 39,258, from the primary checkout's *working* copy — which §0.2 excludes from
> citation and §0.4 documents as permanently modified by design. That reading is not the
> canonical copy, and this SPEC does not rest on it.)
>
> **The unit is characters throughout** — the parent SPEC mixed 61,908 **bytes** with a
> character count and a 40,000 cap, so the measuring command is stated: `wc -m`.
>
> **[HARD] v0.2.0 — the before-value moved, and the fixed floor is withdrawn.** Re-measured in
> the t1259 worktree, 2026-09-26: `git show origin/develop:CLAUDE.local.md | wc -m` → **44,740**
> (`git show develop:CLAUDE.local.md | wc -m` agrees), not the **44,381** the carve recorded. The
> canonical copy is a live maintainer document that other cards keep editing, so **any** absolute
> before-value written into a SPEC is stale the moment a sibling card lands. The v0.1.1 repair
> corrected the arithmetic of a constant that should not have been a constant: a floor of "at
> least 4,382" is now wrong by 359 characters, and would have been read as authoritative.
>
> The floor is therefore **derived, not fixed**: the binding condition is `after <= 39999`, and
> the reduction is whatever `before - after` turns out to be, with `before` re-measured at the
> milestone by the command named above. Both values are still recorded — that obligation is what
> stops the criterion being discharged against an already-compliant copy — but neither is
> predicted here. At 44,740 the reduction is ~4,741 characters, roughly 11%; that figure is an
> expectation, not a bound.

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
