# SPEC-INSTRUCTION-FILES-UNIFY-001 — research

> **Scope note (v0.3.0).** The migration verb, the Codex fallback advisory, the `moai update` /
> `moai doctor` advisories, the docs-site rewrite, and this repository's own `CLAUDE.local.md`
> migration moved to `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` (card t1259). This survey is shared
> context both SPECs read and is kept intact rather than duplicated.
>
> **Source locations are cited by symbol.** `origin/develop` is roughly 93 commits ahead of this
> worktree's base and has already moved one of them (card t1224 shifted `frozenInstructionFiles`
> by two lines while leaving the symbol intact). Every figure below is attributed to base develop
> `553e224f3` on 2026-09-26 and is re-measured after t1175 lands (plan.md §B).

## §A Measured facts (M0)

The full evidence-bearing record is `.moai/reports/t1243/m0/verdict.md`, produced against
Claude Code `2.1.283` in headless mode, one throwaway git repository per probe, one
sentinel token per file. It is not restated here; design.md §A.1 carries the four findings
in summary form and this section carries only what is not in either.

The verdict's own Gaps section lists what was NOT observed. Those gaps are carried forward
verbatim as unestablished, not upgraded:

- **P7 line 3 (ancestor discovery) is a synthetic-fixture observation, UNCONFIRMED.** It was
  made in a plain `git init` repository with an ordinary nested directory, not a real linked
  git worktree, and it names `CLAUDE.local.md` rather than `AGENTS.local.md`. It may not be
  used as the premise of a proposed solution (lead ruling, 2026-09-26).
- Interactive mode was not measured. `bypassPermissions` producing the same skip is
  evidence against an approval gate, not proof of its absence.
- Writing through an in-project symlink was not measured; P3 establishes reading only.
- Codex-side reading was not measured at all. Every probe was Claude Code. The Codex
  `developer_instructions` path is asserted from source reading.
- The exact containment predicate was not derived. "Outside the project directory" is the
  observed discriminator, not a quoted rule.
- The `@import` depth limit (documented as 4 hops) was not exercised; all probes were
  depth 1. Behavior under nested imports crossing the boundary mid-chain is unknown.

## §B Source-tree survey (this worktree, base develop `553e224f3`)

Read-only observations made while authoring this SPEC.

| Surface | Observed state |
|---|---|
| `AGENTS.md` (root) | 16,441 bytes, 8 `##` sections |
| `internal/template/templates/AGENTS.md.tmpl` | 19,177 bytes, 12 `##` sections — the four extra are Hook Event Coverage, Configuration Map, moai CLI Verbs, Status Line Tokens. **The `.tmpl` suffix is an invariant, not an accident** — see §C.2 |
| `internal/config/token_budget_guard.go` — `CodexContractByteCeiling` | `CodexContractByteCeiling = 24576`, the per-file ceiling; `contractDocuments` follows the mirror to its renamed path |
| `CLAUDE.md` (root) and template | 19,553 bytes, byte-identical sizes; already carries `@AGENTS.md` |
| `CLAUDE.local.md` | 61,908 bytes / 44,381 characters on `develop` (the copy its own §0.1 declares canonical), git-tracked despite being listed in `.gitignore` |
| `AGENTS.local.md` | absent from the tree; already gitignored in both `.gitignore` and `internal/template/templates/.gitignore` |
| `internal/cli/codex_contract.go` — `codexClaudeLocalName` / `codexLocalInstructionName` | both constants already exist; the comment already calls `AGENTS.local.md` the Codex-only local input |
| `internal/cli/codex_launcher.go` — local-instruction loop | iterates `{codexClaudeLocalName, codexLocalInstructionName}` — `CLAUDE.local.md` first |
| `internal/hook/pre_tool.go` — `frozenInstructionFiles` | `frozenInstructionFiles = []string{"CLAUDE.md", "CLAUDE.local.md"}`; basename match |
| `internal/harness/curator/dispatch.go` — Tier-3 entry | Tier 3 → `CLAUDE.local.md`, append-only, `BlockTypeLearnedLocal` |
| `internal/cli/codex_contract_link_test.go` — the `@AGENTS.local.md` imports assertion | asserts `executing @AGENTS.local.md imports = 0` |
| `internal/cli/` migrate verbs | only `migrate_agency_*`; no `migrate local-instructions` exists |
| `docs-site/content/` | four locale trees: `en`, `ja`, `ko`, `zh` |

Three consequences worth naming:

1. **Much of the Codex side is already built.** The constant, the gitignore entries, and
   the contract's own framing of `AGENTS.local.md` are in place. REQ-IFU-006 is an
   iteration-order change, not new plumbing.
2. **The mirror's filename is load-bearing.** The template contract is `AGENTS.md.tmpl`, not `AGENTS.md`. Any reconciliation work must reconcile the section set and leave the filename alone (§C.2).
3. **`frozenInstructionFiles` matches on basename**, so adding two strings is sufficient —
   no path normalization work is implied by REQ-IFU-013.

## §C Upstream facts taken as established

### C.1 From the approved design document

Not re-derived here:

- `AGENTS.override.md` must not be used: Codex finds it ahead of `AGENTS.md` in the same
  folder, so it can shadow the deployed contract, and Claude does not read it at all.
- Claude's `claude-md-and-agents-md` project-instructions setting is unusable for this
  purpose: ignored in project and local settings, honored only in user and managed
  settings, and it does not cover local files.
- The instruction budget is charged against project files only (from Codex's `agents_md.rs`
  on main). The current `AGENTS.md` preamble claim that a personal `~/.codex/AGENTS.md` is
  consumed first and narrows the project budget is therefore wrong and is corrected by
  REQ-IFU-017. The adjacent claim — that overflow is silently truncated from the tail — is
  correct and stays.

### C.2 From card t925 (commit `703598937`, 2026-09-18) — the `.tmpl` rename

Prior measurement, taken as established:

- The mirror was renamed from `AGENTS.md` to `AGENTS.md.tmpl` **specifically to take it out
  of Codex's filename-based discovery.**
- Before the rename the nested-chain guard was red at **33,738 / 32,768 bytes**. Running
  Codex inside `internal/template/templates/` merged the root contract with the mirror,
  **dropped the mirror's last section entirely, and cut the preceding section mid table
  row** — no warning, exit 0, stderr empty.
- **32,768 is Codex's measured `project_doc_max_bytes` default.** Raising it is forbidden as
  a diet substitute (REQ-AMC-018): the project-scope override takes effect only once the
  user registers `trust_level = "trusted"` and is **silently ignored** until then, and a
  distributed user's first session is untrusted by construction.
- Codex discovers project instructions **by filename** and exposes no exclusion key — its
  config carries only `project_doc_max_bytes` and `fallback_filenames`. Renaming was the
  only available lever.
- The deployer strips the `.tmpl` suffix, so a user project still receives `AGENTS.md`.
- Root and mirror are edited together and **diverge intentionally** — which is why AC-IFU-008
  compares section sets rather than content. The t925-era figure of 46 lines is retired: it does
  not reproduce against this tree (measured 2026-09-26 against `553e224f3` — 57 template-only,
  17 root-only lines; spec.md §C.4).

### C.3 From card t1219 item (1) — the worktree duplicate load

A worktree session loads **two** copies of `CLAUDE.local.md`: the worktree copy at 44.4k
characters and the primary copy at 39.3k characters, with differing content, costing
roughly 30k extra tokens. The lead judges this very likely the same phenomenon as the M0
P7 line-3 observation. Scoped out of this SPEC (spec.md §D); cross-referenced because the
mechanism that would carry local instructions into a worktree is the same one that
double-loads them.

## §D Open questions for the run phase

| # | Question | Why it cannot be answered now |
|---|---|---|
| Q1 | Does Claude Code discover `AGENTS.local.md` by ancestor walk, as it appears to discover `CLAUDE.local.md`? | The M0 P7 line-3 observation names `CLAUDE.local.md` only. |
| Q2 | Does that discovery hold in a **real linked git worktree**, where `.git` is a file rather than a directory? | P7 used a synthetic nested directory inside a plain `git init` repository. Per the lead's ruling the observation is UNCONFIRMED and may not serve as a premise. |
| Q3 | Which of the four template-only `AGENTS.md` sections belong in a harness-neutral contract, under both the 24,576-byte per-file ceiling and the 32,768-byte nested sum? | A content judgment the reconciliation milestone makes; not a measurement. |
| Q4 | Does the Codex launcher's fallback branch have a surface on which to emit the advisory without polluting `developer_instructions`? | Requires reading the launcher's diagnostic path, which is run-phase work. |
| Q5 | Does **Codex** discover `AGENTS.local.md` by filename? | Unmeasured. It is not literally `AGENTS.md`, so it probably falls outside discovery — but `fallback_filenames` is configurable, and the design assumes the content arrives via `developer_instructions`, which is exempt from the 32,768 budget. If discovery picks it up, the content is counted twice and the budget arithmetic is wrong. |

Q1 and Q2 together gate the worktree leg (design.md §A.5) and are the run phase's M1,
held by AC-IFU-021. Q5 is held by AC-IFU-022 and is the reason that criterion needs a
tail sentinel: the failure mode is silent truncation, so a session that merely completes
proves nothing.
