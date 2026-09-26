# SPEC-INSTRUCTION-FILES-UNIFY-001 — design

> **Scope note (v0.3.0).** The migration verb, the Codex fallback advisory, the
> `moai update` / `moai doctor` advisories, the docs-site rewrite, and this repository's own
> `CLAUDE.local.md` migration moved to `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` (card t1259).
> This document's §C read-order analysis and §D budget analysis are shared context both SPECs
> read; the sections are kept here intact rather than duplicated, and §C.1 of plan.md names
> the one function the two SPECs both edit.

## §A The decision this design still has to make

The operator-approved design document settled the three-file structure. One question inside
it was answered by a recommendation the M0 measurement has since refuted, so it is
re-opened here: **how does a session whose cwd is a worktree reach the user's local
instructions, which live in the primary checkout?**

### A.1 What was measured

From `.moai/reports/t1243/m0/verdict.md`, against Claude Code `2.1.283`, headless, one
throwaway repository per probe. The deciding commands and verbatim output are in that
file's Evidence section and are not restated here.

| Finding | Status | Consequence |
|---|---|---|
| M0-1 — an `@import` with an absent target is silently skipped (exit 0, empty stderr, the directive survives as literal text) | measured | Safe for a user with no `AGENTS.local.md`, but a broken import produces no signal. Drives REQ-IFU-023. |
| M0-2 — an `@import` of a gitignored in-project file loads normally, no approval window | measured | `AGENTS.local.md` being gitignored is not an obstacle. |
| M0-3 — an `@import` whose resolved target is **outside the project directory** is silently NOT loaded (symlink and bare absolute path; default and `bypassPermissions`; target inside and outside `$HOME`; and in the worktree geometry) | measured, 6 probe rows | **The approved design's Q5 symlink recommendation is refuted.** The in-project symlink control (P3) DID load, so symlinks as such work — escaping the project directory is the discriminator. |
| P7 line 3 — a repository-root `CLAUDE.local.md` loads in a session whose cwd is a nested directory | **synthetic-fixture observation, UNCONFIRMED** | See A.2. Measured in a plain `git init` repository with an ordinary nested directory, **not a real linked git worktree**. |

### A.2 [HARD] The ancestor-discovery observation is not a premise

Per the lead's ruling (2026-09-26), P7 line 3 **MUST NOT be used as the premise of a
proposed solution** until it is re-measured in a real linked git worktree during the run
phase. Wherever it appears in this SPEC it is labelled *synthetic-fixture observation,
unconfirmed*. The re-measurement is M1 (plan.md) and carries its own criterion
(AC-IFU-021).

Two things make the distinction load-bearing rather than pedantic. A real linked worktree
has `.git` as a **file** rather than a directory, which is exactly the kind of detail a
root-discovery walk keys on. And the observation names `CLAUDE.local.md` — the filename
this SPEC retires — not `AGENTS.local.md`.

### A.3 [HARD] If the mechanism is real, it is also a defect

The lead judges P7 line 3 to be very likely the same phenomenon as open card **t1219 item
(1)**: a worktree session loading **two** copies of `CLAUDE.local.md` — the worktree copy
at 44.4k characters and the primary copy at 39.3k characters, with differing content, for
roughly 30k extra tokens.

If they are one phenomenon, then *the mechanism that would make local instructions reach a
worktree is the same mechanism that double-loads them*. A proposal that leans on ancestor
discovery without saying what happens to the duplicate is incomplete, so this design says
it explicitly:

- **Deduplication is out of scope for this SPEC** (spec.md §D). It belongs to t1219, which
  is already open against the phenomenon.
- **This SPEC must not make the duplicate worse.** Adding `AGENTS.local.md` alongside a
  still-present `CLAUDE.local.md` would give a worktree session up to four local-instruction
  loads. That is what the no-coexistence invariant (REQ-IFU-010) prevents, and it is the
  reason that invariant is blocking rather than cosmetic. **As of v0.3.0 that invariant is
  `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`'s** (card t1259) — the obligation is unchanged and its
  criteria travelled with it; this SPEC's dependency on it is a cross-SPEC one, recorded in
  plan.md §C.1.
- **If M1 confirms the mechanism, the finding is handed to t1219** with its evidence, and
  this SPEC's worktree leg stays contingent on t1219's resolution rather than claiming it.

### A.4 The options

**Option 1 — accept that worktrees lose local instructions.** A worktree session reads
`AGENTS.md` and the Claude mechanism layer, but not the user's local file. Measured basis:
M0-3, all six rows. Cost: a worktree lane loses the maintainer's local rules — in this
repository, the git-flow lane protocol and the integration-window discipline live there.

**Option 2 — keep a Claude-discovered filename for the local file.** Rejected: it requires
keeping `CLAUDE.local.md`, which is the name this SPEC retires, and it rests on the same
unconfirmed observation as Option 3 without Option 3's benign failure mode.

**Option 3 (proposed, contingent) — `@import` in the primary checkout; in a worktree, the
import silently no-ops and the local file is reached, *if at all*, by whatever discovery
mechanism M1 establishes.**

### A.5 What is proposed, and what happens if it fails

**Option 3, explicitly contingent on M1.**

The reasoning is not that ancestor discovery works — that is unconfirmed and may not. It
is that Option 3's failure mode is *already measured to be benign*. In a worktree the
`@AGENTS.local.md` import does not resolve (M0-3) and is silently skipped (M0-1), costing
nothing. So Option 3 does not bet on the unconfirmed fact; it only benefits if M1 confirms
it.

**Named contingency — what the SPEC does in each M1 outcome:**

| M1 result | Consequence |
|---|---|
| Discovery confirmed in a real worktree, for `AGENTS.local.md` | The worktree leg is real. **The finding and its criterion TRANSFER to card t1219 or a successor SPEC; neither lands in this SPEC's criterion set.** This SPEC does not claim the leg as resolved. |
| Discovery confirmed for `CLAUDE.local.md` but NOT for `AGENTS.local.md` | The worktree leg does not exist under the new filename. Option 3 degrades to Option 1; recorded as a known limitation in progress.md, not retried. |
| Discovery not reproduced in a real worktree at all | P7 line 3 was a synthetic-fixture artifact. Option 3 degrades to Option 1, and t1219's cause is elsewhere — that finding is reported to t1219 too. |

In every branch the primary-checkout behaviour is unchanged and no deployed file needs a
different shape. That is the property that makes it safe to proceed to M2 before the
worktree question is settled.

**[HARD] The confirmed branch transfers; it does not add here.** §A.3 establishes that if
the mechanism is real, the worktree leg is entangled with the t1219 duplicate load — the
thing that carries local instructions into a worktree is the thing that double-loads them.
A criterion in THIS SPEC asserting the leg works, while t1219 holds an open defect against
the same mechanism, would be asserting half a behaviour. The leg is therefore t1219's to
assert once it disposes of the duplicate, and this SPEC's criterion set is closed to it.

This is a scope boundary, not a capacity workaround. It would hold at any criterion count.

**And the transfer is itself asserted.** The obligation used to rest on prose in three files
plus the run-phase agent remembering it; acceptance.md §D.3 now carries it as a conditional
Definition-of-Done item requiring the handoff be recorded in `progress.md` §E.2.

---

## §B The import-resolution gate

M0-1 has the longest reach: an `@import` that does not resolve produces exit 0 and empty
stderr, and the directive survives as literal text. Every signal a reader would use to
detect the failure is absent, and the symptom — "the instructions were ignored" — is the
hardest shape to diagnose.

REQ-IFU-023 therefore requires a check that the import **resolved**, not that the directive
is present. Grepping `CLAUDE.md` for `@AGENTS.local.md` establishes only the second.
AC-IFU-019 states the mechanical form: a sentinel in a fixture `AGENTS.local.md` must
appear in what the session actually loaded.

The same reasoning drives AC-IFU-022 (§D below): a passing session proves nothing when the
failure mode is silent truncation.

---

## §C Read-order change (Codex), and the constants it touches

The local-instruction loop in `internal/cli/codex_launcher.go` ranges over
`[]string{codexClaudeLocalName, codexLocalInstructionName}` — `CLAUDE.local.md` first. Both
constants already exist in `codex_contract.go`, and the contract comment already
calls `AGENTS.local.md` the Codex-only local input.

**This SPEC's change is the iteration order alone (REQ-IFU-006).** The deprecation advisory on
the fallback branch is `REQ-IFU-007`, and **as of v0.3.0 that requirement is
`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`'s** (card t1259) — the obligation is unchanged and its
criterion travelled with it; this SPEC's dependency on it is a cross-SPEC one, recorded in
plan.md §C.1. The two lanes edit the same function, which is why §C.1 exists and why M2
(plan.md §E) states the advisory is not written here. Annotated in the §A.3 style above.

The provenance preamble keeps the literal filename of the file read — **`REQ-IFU-008`, also
the sibling's** (same annotation, same reason): a preamble naming `AGENTS.local.md` while
`CLAUDE.local.md` was read makes the fallback invisible in exactly the situation the advisory
exists to surface. It is described here because §C is shared context both SPECs read (§ above),
not because this SPEC implements it.

Two further constants in the same file are touched by the new structure and are named here
so they are not discovered late:

- `codexLinkAgentsDirective` (`codex_contract.go`) — the `@AGENTS.md` link the contract
  writes into `CLAUDE.md`. Under REQ-IFU-002 the link block becomes two imports around a
  mechanism layer, so what this constant represents changes.
- `codexCreatedAgentsBody` / `codexCreatedClaudeBody` (`codex_contract.go`) — the
  stub bodies created when a file is missing. `codexCreatedClaudeBody` currently emits the
  single `@AGENTS.md` link; it must emit the new two-import shape or the created stub will
  not satisfy AC-IFU-002.

---

## §D Codex budget: two distinct ceilings, and one unmeasured input

### D.1 The two ceilings

They are different limits with different owners and must never be conflated:

| Ceiling | Value | Owner | Scope |
|---|---|---|---|
| Per-file contract ceiling | 24,576 B | `CodexContractByteCeiling`, declared in `internal/config/token_budget_guard.go` | one contract document |
| Nested-sum discovery budget | 32,768 B | Codex's measured `project_doc_max_bytes` default | the sum of every file Codex discovers by filename in the chain |

Both get their own criterion (AC-IFU-005, AC-IFU-006). The approved design document
mentions only the first.

A **third** limit sits alongside them and is neither: the always-loaded instruction surface
guarded by `TestAlwaysLoadedTokenBudget` (`internal/config/token_budget_guard_test.go`). Root
`AGENTS.md` is inside that surface, and this SPEC moves the total from two directions
(REQ-IFU-002 thins `CLAUDE.md`; REQ-IFU-018 may fold template-only sections in), while card
t1175 concurrently retunes the same budget. Neither ceiling above would catch an overrun
there, so `AC-IFU-025` asserts that guard by name.

[HARD] A criterion invoking any of the three asserts `--- PASS: <symbol>` under `-v`, never
exit `0` alone: `go test -run` exits `0` when its pattern matches nothing. The plan-audit of
`1140bcd1d` found the nested-sum guard — the sole enforcement of REQ-IFU-025 — named as
`TestNestedChainBudget`, a pattern matching no test, and therefore unable to fail.

**[HARD] Raising `project_doc_max_bytes` is not an available remedy.** Per REQ-AMC-018, the
project-scope override takes effect only once the user registers `trust_level = "trusted"`,
and is ignored **silently** until then — and a distributed user's first session is untrusted
by construction. Diet is the only lever.

### D.2 [HARD] The `.tmpl` suffix is an invariant, not an oddity

The template mirror is `internal/template/templates/AGENTS.md.tmpl`, **not**
`internal/template/templates/AGENTS.md`. The suffix is load-bearing.

Card **t925** (commit `703598937`, 2026-09-18) renamed it precisely to take it out of
Codex's filename-based discovery. Prior measurement from that commit, taken as established
and not re-derived:

- Before the rename the nested-chain guard was red at **33,738 / 32,768 bytes**. Running
  Codex inside `internal/template/templates/` merged the root contract with the mirror,
  **dropped the mirror's last section entirely, and cut the preceding section mid table
  row** — no warning, exit 0, stderr empty.
- Codex discovers project instructions **by filename** and exposes no exclusion key; its
  config carries only `project_doc_max_bytes` and `fallback_filenames`. Renaming was the
  only available lever.
- The deployer strips the `.tmpl` suffix, so a user project still receives a file named
  `AGENTS.md`.
- `contractDocuments` follows the mirror to its renamed path, so the 24,576-byte per-file
  ceiling still binds. The nested-sum guard deliberately names no path and walks for the
  filename Codex keys on.
- Root and mirror are edited together and **diverge intentionally**. The t925-era figure of 46
  lines does not reproduce against this tree: measured 2026-09-26 against base develop
  `553e224f3`, `diff AGENTS.md internal/template/templates/AGENTS.md.tmpl` reports 57
  template-only (`^>`) and 17 root-only (`^<`) lines. Retired rather than carried forward; what
  it supports holds at any value (spec.md §C.4).

**Therefore REQ-IFU-018 ("reconcile root and template onto one shape") reconciles the
SECTION SET, never the filename.** Restoring the mirror to a Codex-discovered name reverses
t925 and reintroduces the silent mid-table-row truncation.

This is stated explicitly rather than left implicit because the failure it guards against
is precisely a later reader tidying up an extension that looks accidental. The invariant is
held by REQ-IFU-024 and mechanically by AC-IFU-004 — prose alone would not have stopped the
original defect either.

Note also that the intentional divergence means AC-IFU-008 compares the `## `
section set, not file content. A byte-level diff would fail by design.

### D.3 Unmeasured: does Codex discover `AGENTS.local.md`?

Adding `AGENTS.local.md` to a user project adds another instruction file to the tree.
**Whether Codex discovers it by filename is UNMEASURED.** It is not literally `AGENTS.md`,
so it probably falls outside discovery — but `fallback_filenames` is configurable, and the
whole design assumes Codex receives that content through `developer_instructions`, which is
exempt from the 32,768 budget.

If discovery does pick it up, the content is counted **twice** and the budget arithmetic in
the approved design is wrong.

[HARD] This is a run-phase measurement with a named command (AC-IFU-022), not an
assumption. And the criterion cannot be "run Codex and see that it works": **the failure
mode is silent tail truncation, so a passing session proves nothing.** The measurement
places a sentinel at the very end of the assembled instruction text and asserts its
presence, making truncation observable rather than inferred from the absence of an error.

---

## §E Test-assertion inversions

Two existing assertions encode the pre-unification rule and invert together:

- `codex_contract_link_test.go`'s link assertion currently states `executing @AGENTS.local.md imports = 0`. Under
  REQ-IFU-002 the deployed `CLAUDE.md` imports it, so the assertion becomes: `CLAUDE.md`
  imports `AGENTS.local.md`; `AGENTS.md` does not. **The `AGENTS.md`-side half is
  load-bearing and must not be dropped** — the neutral contract importing a local file
  would make it non-neutral, and would also feed a discovered file into the nested-sum
  budget.
- `codex_local_instructions_test.go` read-order assertions follow §C.

---

## §F Template-First and the two-mirror rule

Template first, then `make build`, then the root copies.

[HARD] A grep proving one mirror clean establishes nothing about the other. Each mirror is
a separate command with a separate exit code. Every criterion touching a deployed file
names both paths explicitly — and for `AGENTS.md` the two paths have **different
filenames** (§D.2), which is itself a way the one-mirror mistake gets made.

Current asymmetry the reconciliation resolves: root `AGENTS.md` = 16,441 bytes / 8 `## `
sections; `internal/template/templates/AGENTS.md.tmpl` = 19,177 bytes / 12 sections. The
four template-only sections are Hook Event Coverage, Configuration Map, moai CLI Verbs, and
Status Line Tokens. Reconciliation must decide whether those belong in a contract at all —
they are reference material, not obligations — under both ceilings from §D.1.
