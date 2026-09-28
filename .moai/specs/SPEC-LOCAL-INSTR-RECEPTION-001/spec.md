---
id: SPEC-LOCAL-INSTR-RECEPTION-001
title: "Worktree reception gate for AGENTS.local.md, then this repository's local-instruction migration"
version: "0.1.0"
status: draft
created: 2026-09-28
updated: 2026-09-28
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "AGENTS.local.md, CLAUDE.local.md, .moai/docs, internal/template/templates"
lifecycle: spec-anchored
tags: "instruction-files, worktree, reception, migration, repo-local"
tier: L
related_specs: [SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001]
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-28 | **Created by card t1290 as the M3 split of `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` (plan-audit iter4 finding D2, severity major, class blocking — report `.moai/reports/t1259/plan-audit-iter4.md` in the t1259 worktree).** The split relationship is a **move**: this SPEC takes over the parent's `REQ-IFU-021` and `REQ-IFU-022` (this repository's own `CLAUDE.local.md` → `AGENTS.local.md` migration), and their two acceptance criteria (`AC-IFU-007`, `AC-IFU-024`). The ids keep their `IFU` infix — the carve precedent from `1140bcd1d` — so every existing cross-reference and audit citation still resolves. The parent retains its M1 (migration verb), M2 (Codex fallback advisory), and M4 (docs-site); its owning card t1259 owes the corresponding excision edit (recorded in this SPEC's plan.md §B). What the move adds and the parent never had: the **reception gate** — `REQ-LIR-001` through `REQ-LIR-004`, a [HARD] first milestone proving that a worktree session of this repository demonstrably receives `AGENTS.local.md` content before any content moves, because the landed contract states the opposite for the geometry the parent would have created (`AGENTS.md:262`, and the parent-of-parent's measured `NOT_PRESENT`, `SPEC-INSTRUCTION-FILES-UNIFY-001/progress.md` §M1a). Executing the parent's M3 as written would have stripped maintainer doctrine from every lane with all ten parent criteria green. Autonomous kickoff per `CLAUDE.local.md` §31; card-inherent design choices take the recommended option and record reasoning in design.md. |

---

## §A Context

`SPEC-INSTRUCTION-FILES-UNIFY-001` built the three-file instruction structure; its child
`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` owns the user-owned-file half, including this repository's
own migration (`REQ-IFU-021`/`REQ-IFU-022`). Card t1259's plan-audit iter4 found that migration
unimplementable as specified: the landed contract (`AGENTS.md:262`) states that in a linked
worktree the `@AGENTS.local.md` import is skipped silently, and the parent-of-parent's real
nested-worktree measurement (§M1a, `claude 2.1.283`, `LOCAL_AGENTS_TOKEN = NOT_PRESENT`)
confirms it for the gitignored geometry. This repository's lanes **are** worktree sessions, and
today they receive the maintainer doctrine through `CLAUDE.local.md`'s directory-ancestor walk.
Migrating the file without first proving a worktree reception path would delete that doctrine
from every lane.

This SPEC therefore owns two things in a fixed order:

1. **The reception gate** — prove, by measured token probe on a real worktree in this
   repository's geometry, that a worktree session receives `AGENTS.local.md` content — before
   any content moves.
2. **The migration itself** — the moved `REQ-IFU-021`/`REQ-IFU-022` scope, executed only behind
   that gate.

The reception mechanism (force-tracking the migrated file, with `.worktreeinclude` as the
fallback) is decided in design.md §A/§C with the measurements each choice rests on.

---

## §B Goal

Migrate this repository's maintainer local instructions to `AGENTS.local.md` **without any lane
losing them**: first prove the channel that delivers the file into every worktree session, then
move the content through that proven channel.

---

## §C Requirements (GEARS)

> **Id conventions.** `REQ-LIR-*` are this SPEC's own reception-gate requirements. `REQ-IFU-021`
> and `REQ-IFU-022` are **moved** from `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` and keep their ids
> so existing cross-references resolve; their bodies carry the additions the reception gate and
> the iter4 D1 finding require, recorded in HISTORY at the version that adds them.

### C.1 Reception guarantee (the [HARD] first milestone)

- **REQ-LIR-001** — When this repository's maintainer local-instruction content lives in
  `AGENTS.local.md`, a Claude Code session entered in a real linked worktree of this repository
  shall load that content, for each supported entry path: `moai cc -w`, `claude --worktree`, and
  `EnterWorktree` re-entry. Reception is proven only by the token probe defined in
  acceptance.md (`AC-LIR-001`~`003`), run from that session's working directory, with the
  tracked-contract token as positive control in the same run.

- **REQ-LIR-002** — When the content lives in `AGENTS.local.md`, a `moai codex -w` child session
  shall receive that content through the launcher's `developer_instructions` path, proven by the
  Codex-side probe (`AC-LIR-004`). The mechanism claim — that a `-w` child reads the original
  project root (`AGENTS.md` §8) — shall be verified by measurement, not assumed.

- **REQ-LIR-003** — The probe shall be demonstrated to fail: When the probe runs in a geometry
  that lacks the file (the pre-migration tree, and a worktree created from a base commit without
  the file), the probe shall return `NOT_PRESENT` for the local token while the positive control
  reads present in the same run (`AC-LIR-005`). A probe whose failure has never been observed
  proves nothing.

- **REQ-LIR-004** — While the criteria of REQ-LIR-001 and REQ-LIR-002 are unproven, the
  migration steps of this SPEC shall not be executed. This is the [HARD] ordering the iter4
  audit demanded and the parent's plan bullet lacked: content moves only into a channel proven
  to deliver it.

### C.2 The migration (moved from SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001)

- **REQ-IFU-021** — This repository's own `CLAUDE.local.md` — specifically **the copy committed
  on the `develop` branch**, which that file's own §0.1 declares canonical — shall migrate to
  `AGENTS.local.md` at **under 40,000 characters** (`wc -m`, i.e. at most 39,999), with
  operational procedure relocated to `.moai/docs/`. The pre-migration value shall be
  **re-measured at the milestone** rather than carried from any document. The migrated file
  shall be **present in the committed tree** — force-added past `.gitignore` — because every
  measurement this requirement feeds reads the committed tree, not a working copy. This
  requirement shall be executed only after REQ-LIR-004's gate is satisfied.

- **REQ-IFU-022** — The migrated `AGENTS.local.md` §0 shall name `AGENTS.local.md` as the file
  whose canonical copy it discriminates, read from the committed tree.

- **REQ-LIR-005** — After the migration, the committed tree shall carry exactly one local
  instruction file: `git ls-files` shall list `AGENTS.local.md` and shall not list
  `CLAUDE.local.md`.

### C.3 Documentation reconciliation

- **REQ-LIR-006** — When the reception mechanism lands, the documentation stating the
  worktree-reception negative (`AGENTS.md` §8's import-skip sentence and `CLAUDE.md` §18) shall
  be updated to match the measured behavior. The edit is Template-First: the template copies
  (`AGENTS.md.tmpl`, template `CLAUDE.md`) change first, then `make build`, then the root
  copies, carrying no card ids, SPEC ids, or internal dates.

---

## §D Out of Scope

This SPEC deliberately does not build the following.

### Out of Scope — the migration verb and the Codex advisory

- The `moai migrate local-instructions` verb itself, the `moai update`/`moai doctor` advisories,
  the Codex fallback-advisory work, and the no-coexistence invariant's guard behavior. Those
  remain with `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` (its `REQ-IFU-007`~`REQ-IFU-012`,
  `REQ-IFU-020`). This SPEC **uses** the verb it builds; it does not build it.

### Out of Scope — docs-site rewrite

- The 24-file docs-site rewrite across four locales (`REQ-IFU-020`). It stays with the parent.

### Out of Scope — user projects and template distribution of the file

- Migrating any user project's local instructions, and shipping `AGENTS.local.md` itself in the
  template tree. The file is user-owned; the distributed template keeps it out. What this SPEC
  may touch in the template tree is only the **reception plumbing** (design.md §C fallback) and
  the contract wording of `REQ-LIR-006`, both of which are product-neutral.

### Out of Scope — the worktree duplicate-load

- Deduplicating the local-instruction load in a worktree session (card t1219 item (1)). This
  SPEC's reception work must not make the duplicate worse, and its outcome changes what that
  card can assume, but the dedup itself is not here.

### Out of Scope — rule substance

- Rewriting what any relocated maintainer rule says. The migration splits procedure into
  `.moai/docs/`; it does not re-decide doctrine.

---

## §E Cross-references

- `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` — the parent this SPEC's §C.2 is moved from; owns the
  verb this SPEC's migration runs.
- `SPEC-INSTRUCTION-FILES-UNIFY-001/progress.md` §M1a/§M1b — the real-worktree reception
  measurements (negative for `AGENTS.local.md`, positive for the `CLAUDE.local.md` walk) this
  SPEC's gate is built from.
- `.moai/reports/t1259/plan-audit-iter4.md` — the D2 blocking finding that forced this split,
  and the D1 committed-tree finding whose fix is folded into `REQ-IFU-021`.
- `AGENTS.md` §8 and `:262`; `CLAUDE.md` §18 and `:171` — the contract statements this SPEC
  measures against and later reconciles.
- `.worktreeinclude` (root + `internal/template/templates/`) — the existing native copy
  mechanism evaluated as fallback in design.md §C.
- `internal/cli/codex_contract.go` (`codexLocalInstructionName`,
  `codexClaudeLocalName`) and `internal/cli/codex_launcher.go` — the Codex read path.
- `.claude/rules/moai/workflow/worktree-integration.md` § `.worktreeinclude` — the documented
  coverage claim (creation-time, launcher and isolation paths).
