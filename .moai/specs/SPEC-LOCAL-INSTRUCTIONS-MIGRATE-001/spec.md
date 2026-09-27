---
id: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
title: "Local-instruction migration — CLAUDE.local.md to AGENTS.local.md, with advisories and docs"
version: "0.3.0"
status: draft
priority: P1
phase: "v3.3.0 target"
created: 2026-09-26
updated: 2026-09-28
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
| 0.2.1 | 2026-09-26 | **plan-audit iter1 repairs (PASS-WITH-DEBT 0.85, report `.moai/reports/t1259/plan-audit-iter1.md` at `b8fb023e8`).** Four blocking defects, all in the acceptance layer, all repaired. **D1** — `AC-IFU-023`'s locale-parity clause asserted equality over a set already unequal (`claude-md-guide.md` ko=10 vs 18; `quickstart.md` ko=14 vs 10); equality replaced by **equal-delta against a recorded baseline**, which keeps the locales-in-step property without pulling an unscoped twelve-section restructure into M4. **D2** — `AC-IFU-029` passed before any work; `REQ-IFU-008` is now declared a **preservation requirement** here and the criterion a **regression guard** with its baseline recorded, AND extended to the new fallback path so it has a red-able half. **D3** — `AC-IFU-031` cited a PR head this lane never produces; re-sited on the `origin/develop` run carrying the lane's merge SHA. **D4** — `AC-IFU-024` used `grep -c` for a per-sentence assertion; now `grep -n -C1`, judged per line. Optional **D5** (stdout-vs-stderr rationale) and **D6** (`REQ-IFU-012` leading `When`) also applied. |
| 0.2.2 | 2026-09-26 | **plan-audit iter2 repairs (PASS-WITH-DEBT 0.92, report `.moai/reports/t1259/plan-audit-iter2.md` at `28476f1a9`).** Five of six iter1 repairs confirmed; two items remained. **D3-residual** — `AC-IFU-031` still asserted a docs-site build no workflow performs (`grep -rn 'hugo\|vercel' .github/workflows/` → nothing across 19 files). The whole clause is now grounded against the workflow files once, as the audit asked: `CI` and `spec-lint` named as the checks that actually run on a `develop` push, the docs-site build clause **dropped** rather than re-sited onto an unmeasured Vercel behaviour, and `docs i18n parity check` named in its true weight — it fires on this head but is **advisory, `strict=false`** by construction, so it is read, never gated. **N1** — the v0.2.1 alternation silenced this file's own `no tests to run` guard; split into two commands, and the marker is documented as the discriminator since `go test` exits `0` on a selector matching nothing. |
| 0.2.3 | 2026-09-26 | **plan-audit iter3 repairs (FAIL 0.88 — the audit ceiling; report `.moai/reports/t1259/plan-audit-iter3.md` at `8d73a2a88`).** N1 and the D3 reasoning were both confirmed repaired; the FAIL is that the D3 repair **stopped at the criterion body**. `grep -rn "PR head"` found four live normative sites it never reached — `acceptance.md` §D.3 and `plan.md` §D/§E ×2 — so two sections of one file gave different close instructions for the same criterion and `plan.md` agreed with the superseded one. All four now name the `origin/develop` head carrying the lane's merge SHA. §D.3 additionally gains the four close-time duties the earlier repairs established but never propagated out of the notes that produced them: the `no tests to run` marker read, a named home for recording the docs-i18n log reading, the close-time re-read of the two decaying external readings, and a handover pointer to `SPEC-V3R3-DOCS-PARITY-001` for the docs-parity residual this SPEC does not own. **N2** — two miscited figures corrected: `strict=false` is at `docs-i18n-check.yml:75`, and there are **19** workflow files, not 20. |
| 0.2.4 | 2026-09-26 | **Repair of a defect introduced by the v0.2.3 commit itself.** `da0df2113` shipped `progress.md` with the iter3 block duplicated (113 lines), carrying a second copy of the iter2 verification section and a surviving copy of the brittle-count table v0.2.3 had just replaced — so the file both argued against the count and retained it. Cause: a span-replacing edit whose end anchor was not unique resolved *before* its start anchor, re-emitting the span instead of removing it. No content change to `spec.md`, `plan.md`, or `acceptance.md`; the v0.2.3 repairs are unaffected. Found by the lead's close grep, not by the no-regression trio — all three were green over the duplicated file, because none reads prose structure. |
| 0.2.5 | 2026-09-28 | **Dispatch-time premise re-measurement repairs (card t1259; readings in `.moai/reports/t1259/premise-20260928.md`).** Three stale statements repaired, no requirement or criterion changed. **(1)** `REQ-IFU-008`'s preamble citation `codex_launcher.go:138` → **`:137`** on `develop` = `origin/develop` = `37dc766b9`. The old figure did not merely drift: `git show 5ba87003f:internal/cli/codex_launcher.go \| grep -n 'source: %s'` → `136`, so `:138` was already wrong at the commit it was attributed to. Corrected in `spec.md` §C.1 and `acceptance.md` §D.2; the historical records in `progress.md` and `research.md` are kept and annotated, together with the other drifted `codex_launcher.go` / `codex_contract.go` lines (producer `:123`→`:124`, call site `:825`→`:813`, loop `:125`→`:126`, name constants `:33-34`→`:35-36`, install hint `:804`→`:792`, worktree error `:842`→`:830`). **(2)** `REQ-IFU-021` / `AC-IFU-007` before-value re-measured: `git show develop:CLAUDE.local.md \| wc -m` → **45,810** (44,740 at v0.2.0). The derived floor still yields a satisfiable, non-vacuous criterion: `after <= 39,999` demands a reduction of at least 5,811 characters (~12.7%). The v0.2.0 notes keep their 44,740 as history and now carry a dated pointer. **(3)** Both run-phase blocking dependencies read `status: completed` on `origin/develop` (`SPEC-INSTRUCTION-FILES-UNIFY-001`, `SPEC-ALWAYS-LOADED-DIET-002`), and the parent's loop now iterates `codexLocalInstructionName` first — recorded in `plan.md` §B and `progress.md` §E.1. Status stays `draft`. |
| 0.3.0 | 2026-09-28 | **Scope reduction after plan-audit iter4 (FAIL 0.84, report `.moai/reports/t1259/plan-audit-iter4.md` at `a73c78d1a`); the lead's tier-3 decision, not re-opened here.** Milestone M3 — `REQ-IFU-021`, `REQ-IFU-022`, and the two criteria that verified only them (`AC-IFU-007`, `AC-IFU-024`) — is **transferred to card t1290**, whose prerequisite is a design guaranteeing that worktree sessions receive `AGENTS.local.md` (`AGENTS.md:262` states they currently do not; iter4 D2). Moved to §D Out of Scope with that reason; the verbatim requirement and criterion text stays readable at `4441cf1a6`. **D1** is moot here because both criteria it named left with M3, and it is carried to t1290 with a demonstrated committed-tree predicate (`.moai/reports/t1259/d1-mutant.md`). **D3** — the self-contradicting status paragraph under this table is rewritten to the current state. **N2** — `docs-i18n-check.yml:71-74` → `:71-75` in `AC-IFU-031`. Requirement ids and the remaining milestone ids are kept stable (M1, M2, M4; no renumber). Criteria 10 → **8**; Tier stays **L** (24 docs-site files alone exceed the >15 threshold). Status stays `draft`. |

> **Plan phase state (v0.3.0, card t1259).** Four plan-audits have run against this SPEC —
> iter1 PASS-WITH-DEBT 0.85, iter2 PASS-WITH-DEBT 0.92, iter3 (scoped) FAIL 0.88, iter4
> (dispatch-time) FAIL 0.84 — and every blocking defect from all four is either repaired or
> transferred with its requirement (v0.2.1 - v0.3.0; itemised in `progress.md` §E.1). Both run-phase
> dependencies are **completed** on `origin/develop` (`SPEC-INSTRUCTION-FILES-UNIFY-001`,
> `SPEC-ALWAYS-LOADED-DIET-002`; plan.md §B). iter4 found that the landed parent turned M3's premise
> into a known negative — a worktree session does not receive `AGENTS.local.md` (`AGENTS.md:262`) —
> so M3 (`REQ-IFU-021`, `REQ-IFU-022`) is **split out to card t1290** (§D). What remains — M1 verb
> and advisories, M2 launcher advisory, M4 docs — is implementable on the current tree.
>
> The iter3 close condition still holds as history: `grep -rn "PR head"` over this directory is
> judged per hit, and the passing condition is **no live normative hit** — NOT "no output", since
> every repair record must quote the wording it removed (`progress.md` §E.1, iter3 section).

---

## §A Context

`SPEC-INSTRUCTION-FILES-UNIFY-001` establishes the three-file instruction structure — a
harness-neutral `AGENTS.md`, a thin per-harness `CLAUDE.md`, and a user-owned
`AGENTS.local.md` — together with the guards and byte ceilings that hold it.

This SPEC owns the other half: everything that touches a file the **user** owns. Concretely,
the Codex-side `CLAUDE.local.md` fallback and its deprecation advisory, the explicit migration
verb, the no-coexistence invariant, the `moai update` / `moai doctor` advisories, and the
docs-site rewrite across four locales. (This repository's own `CLAUDE.local.md` migration was
carved out of this SPEC at v0.3.0 and belongs to card t1290 — §D.)

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
harnesses read, with the transition documented in all four locales.

---

## §C Requirements (GEARS)

> Ids are transferred verbatim from `SPEC-INSTRUCTION-FILES-UNIFY-001` and retain the `IFU`
> infix. The gaps (`001~006`, `013~019`, `023~025`) are the carve's footprint: those
> requirements remain with the parent SPEC. `021~022` are a second, later gap: they left this
> SPEC for card t1290 at v0.3.0 (§D). Renumbering would break every cross-reference and the
> audit history that already cites these ids by name.

### C.1 Codex fallback and provenance

- **REQ-IFU-007** — Where only `CLAUDE.local.md` exists, the `moai codex` launcher shall
  read it as a fallback and shall emit a deprecation advisory naming the migration command.
- **REQ-IFU-008** — The Codex provenance preamble shall name the literal filename of the
  file it actually read, never a normalized or substituted name.

  > **[HARD] This is a preservation requirement, not new behaviour.** It is already satisfied at
  > `internal/cli/codex_launcher.go:137`, inside `codexLocalDeveloperInstructionArgs`
  > (`fmt.Fprintf(&payload, "<!-- source: %s -->\n", name)`), measured 2026-09-28 on
  > `develop` = `origin/develop` = `37dc766b9` with `grep -n 'source: %s' internal/cli/codex_launcher.go`. (v0.2.4 and
  > earlier cited `:138` "at `5ba87003f`"; that commit reads `:136`, so the figure was wrong when
  > written. The symbol is the durable anchor; the line number is re-read before use.) `AC-IFU-029` is therefore labelled a **regression guard** rather than
  > a verification, with its passing baseline recorded, and is extended to the new
  > fallback-advisory path — the advisory is emitted from the same launch path that builds the
  > payload, so the change most likely to break this preamble is this SPEC's own.

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
- **REQ-IFU-012** — When both local instruction files exist, or when only `CLAUDE.local.md`
  exists, `moai doctor` shall report the same advisory as `moai update`.

### C.3 Documentation

- **REQ-IFU-020** — The six docs-site pages named below shall describe the three-file
  structure in all four locales (ko, en, ja, zh), naming each by its **path** rather than its
  stem:
  `advanced/claude-md-guide.md`, `advanced/codex-dual-harness.md`,
  `advanced/harness-learning.md`, `claude-code/context-memory/memory.md`,
  `getting-started/quickstart.md`, `cli-reference/update.md`.

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
  no-coexistence invariant bounds the set of local-instruction files the verb leaves in any
  project to one. Whether a worktree session receives that one file at all is a separate
  question — the landed contract says it does not (`AGENTS.md:262`) — and it is card t1290's
  prerequisite, not this SPEC's (next section).

### Out of Scope — this repository's own migration (transferred to card t1290)

- The migration of this repository's own `CLAUDE.local.md` to `AGENTS.local.md` —
  requirements `REQ-IFU-021` (the `develop`-committed copy, under 40,000 characters by `wc -m`,
  procedure relocated to `.moai/docs/`) and `REQ-IFU-022` (the migrated §0 names
  `AGENTS.local.md`), with the two criteria that verified only them, `AC-IFU-007` and
  `AC-IFU-024`, and plan milestone M3. **Transferred to card t1290** by the lead's decision on
  plan-audit iter4. Their verbatim text, including the D3 "which copy, which unit" note and the
  v0.2.0 / v0.2.5 before-value readings, is preserved at `4441cf1a6` (`spec.md` §C.3,
  `acceptance.md` §D.3, `plan.md` M3).
- **Why it left (iter4 D2).** This repository's lanes are worktree sessions, and the landed
  parent contract states that a worktree session does not receive `AGENTS.local.md`
  (`AGENTS.md:262`, measured again at v0.3.0). Migrating the file lanes read today into one they
  would not receive strips the maintainer doctrine from every lane while every criterion stays
  green. t1290's prerequisite is therefore a design that guarantees worktree reception first.
- **What t1290 inherits (iter4 D1).** Both transferred criteria read a gitignored working copy
  (`.gitignore:275` ignores `/AGENTS.local.md`), so a migration that never `git add -f`s the new
  file passes them. The committed-tree predicate that kills that mutant — and its demonstrated
  mutant/control runs — is recorded in `.moai/reports/t1259/d1-mutant.md` for t1290 to adopt.
- **Constraints that travel with it.** Relocating procedure out of the document does not
  re-decide what any rule says; occurrences of the old filename that mark it as retired are
  expected to remain after the rename.

---

## §E Cross-references

- `SPEC-INSTRUCTION-FILES-UNIFY-001` — the parent SPEC this was carved from; holds the
  contract, the guards, and the ceilings.
- `.moai/reports/t1243/plan-audit-iter1.md` — the audit of `1140bcd1d` whose D2 arithmetic
  forced the carve and whose D3 finding was repaired in `REQ-IFU-021` (transferred to card t1290
  at v0.3.0).
- `.moai/reports/t1243/m0/verdict.md` — the M0 measurement the parent SPEC's design rests on.
- `internal/cli/codex_contract.go`, `internal/cli/codex_launcher.go` — the fallback branch and
  the provenance preamble.
- `internal/cli/` `migrate_agency_*` — the existing move-plus-backup verb precedent.
- `CLAUDE.local.md` §0 — the canonical-copy discriminant the transferred REQ-IFU-021 names (card
  t1290).
