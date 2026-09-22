---
id: SPEC-AGENTS-WORKTREE-ROW-001
title: "register the worktree entry capability in the cross-harness contract"
version: "0.1.0"
status: draft
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.2.1 target"
module: internal/template/templates
lifecycle: spec-anchored
tier: M
tags: "agents-md, worktree, codex-launcher, cross-harness, documentation, t1071"
---

# SPEC-AGENTS-WORKTREE-ROW-001 — the cross-harness contract never mentions the worktree entry it already has

## HISTORY

- 2026-09-22 · v0.1.0 · manager-spec · Authored from card t1071 (dispatch 2026-09-22). Every
  measured statement is attributed to a Claim number recorded in this SPEC's §A and re-measured
  in the landing tree (worktree t1071, HEAD `cd99336bf`) with positive controls. The card's
  investigator coordinates are treated as leads, not evidence; where they disagree with the
  re-measured tree, the re-measured coordinates below are the ones this SPEC carries.

---

## §A Background — what is measured, not assumed

Card t1071 dispatches on three investigator citations and requires this card to re-measure them
before writing (card constraint (b)). The re-measurements below were run in the landing tree on
2026-09-22. Each carries its command, its observed output, and a positive control where the claim
is about absence — a zero must be proven to be a real zero, not a broken instrument.

**C1 — the root capability table has no worktree row.** `AGENTS.md:18-25` (read in the landing
tree): the paragraph at `:18-19` introduces "Capability bindings … a row exists only where a
harness driving this contract lacks the capability"; the table at `:21-22` declares three columns
(`Capability` / `Claude implementation` / `If this harness lacks it`); the body carries exactly
three rows at `:23-25` — `question-channel`, `task-list`, `design-sync`. No row mentions a
worktree. Negative control: `grep -c '^| worktree-entry |' AGENTS.md` → `0`. Positive control:
`grep -c '^| question-channel |' AGENTS.md` → `1` (the row-shape pattern fires on the table as
it stands, so the same pattern measuring zero for worktree rows is evidence of absence, not of a
broken instrument).

**C2 — the worktree section lists only Claude launchers.** `AGENTS.md:106-111` (read): §3
opens "Work inside a worktree, entered through the launcher (`moai cc -w <name>`, `moai cc -w
<name> --spawn` for a new window, `EnterWorktree(<path>)` to re-enter)". The Codex launcher form
is absent from the enumeration. This section is the dispatch's `:108-111` citation; the rows
match.

**C3 — the template mirror's capability table has the same gap.**
`internal/template/templates/AGENTS.md.tmpl:18-25` (read): the same three-column table with the
same three rows. The mirror is an intentional fork (card constraint (c)) — this copy was judged
on its own measurement, and it independently lacks the row. Negative control:
`grep -c '^| worktree-entry |' internal/template/templates/AGENTS.md.tmpl` → `0`.

**C4 — the template's moai CLI Verbs table omits `moai codex`.**
`internal/template/templates/AGENTS.md.tmpl:291-305` (read): `## 11. moai CLI Verbs` at `:291`;
table header `:293-294`; nine verb rows at `:295-303` (`moai init`, `moai update`, `moai tool
enable codex`, `moai hook`, `moai doctor`, `moai worktree`, `moai cc` / `moai glm` / `moai gpt`,
`moai migrate cg`, `moai version`). No `moai codex` row. Negative control:
`grep -c '^| \`moai codex\`' internal/template/templates/AGENTS.md.tmpl` → `0`; positive control:
`grep -c '^| \`moai init' internal/template/templates/AGENTS.md.tmpl` → `1`. The dispatch's
`:293-302` citation is off by one row at the bottom (the table closes at `:303`, and the closing
pointer line "Run `moai --help` for the generated, current command surface." is `:305`); the
re-measured range governs.

**C5 — `moai codex -w` exists and resolves, and never creates.** Read
`internal/cli/codex_launcher.go` in the landing tree:

- the command registers at `:324` (`Use: "codex [cli | status | app]"`) inside the launch group
  (`GroupID: "launch"` at `:359`), with `DisableFlagParsing: true` at `:360` — the launcher
  itself strips and consumes `-w`;
- `resolveCodexWorktreeDir` at `:272-299` turns the `-w` value into a directory: the
  L2-absolute form is validated by the same rule `moai cc` applies
  (`resolveWorktreeL2Path`, `:279`), the L1 name form is joined under
  `.claude/worktrees` (`:286`), and an absent directory is a diagnostic — `os.Stat` fails into
  the error at `:289-297` whose message says the launcher "resolves an existing worktree and
  never creates one";
- the same limit is stated in the named diagnostics (`:47` usage line carrying `-w <worktree>`;
  `:52` `codexWorktreeValueDiag` — "moai codex resolves an existing worktree and never creates
  one") and in the command help text (`:341-342` — "launch in an EXISTING worktree instead of
  the project root; this never creates one").

**C6 — the dispatch's `:16-17` citation is a comment, not a declaration.** Read
`codex_launcher.go:16-19`: the lines are part of the file's package-doc comment block ("`-w` is
consumed HERE and never forwarded…"). The card's own positive-control warning (constraint (b) —
"a grep hit inside a comment or a table is not the declaration line") applies to the investigator
citation itself: the load-bearing declaration anchors for this SPEC are `:324` (verb
registration) and `:272-299` (the resolve path), both code, both read in this run.

**C7 — the root AGENTS.md carries no Verbs table.** `grep -n 'moai CLI Verbs' AGENTS.md` → no
matches. The verb table exists only in the template mirror, so the verb-registration fix touches
one file, not two.

**C8 — the command is referenced in prose while missing from the table.**
`internal/template/templates/AGENTS.md.tmpl:270` states "`AGENTS.local.md` is Codex-only and
uncommitted. `moai codex` reads it from the project root…". The prose knows the command exists;
the Verbs table does not list it. This is the asymmetry the registration fixes.

### A.1 The one-line shape of the gap

`moai codex -w <worktree>` exists and works — but no doctrine row mentions it. Entry is
possible; creation is impossible (`:52`, `:289-297`). Without that limit stated where readers
look for it, the next reader of the cross-harness contract assumes generation works — and the
launcher's own refusal message is the first place anyone finds out.

---

## §B Why — what breaks if this stays unregistered

A maintainer driving a Codex lane reads §3's launcher enumeration, sees only `moai cc`, and
concludes cross-harness worktree entry does not exist. The false conclusions then fork in two
directions, and both cost more than the missing row: some sessions work in the shared checkout
when they should have isolated (the exact hazard AGENTS.md §2 and the worktree rules exist to
prevent), and other sessions attempt creation through the launcher and hit the diagnostic.
The contract's job is to state the boundary before the reader tests it.

The Verbs table omission is cheaper but the same shape: the table is the one-stop inventory of
the CLI surface, and a real command absent from it makes the table's completeness unverifiable
by reading alone.

---

## §C Requirements (GEARS)

- **REQ-AWR-001** (ubiquitous) — The root `AGENTS.md` capability-binding table (`AGENTS.md:21-25`)
  shall carry exactly one `worktree-entry` row, and the row's text shall state, in one line, that
  entry is possible through `moai codex -w <worktree>` while creation is impossible — the
  launcher resolves an existing tree and never creates one. The row's shape shall match the table's
  three-column form so the machine guard `grep -c '^| worktree-entry |' AGENTS.md` returns `1`.

- **REQ-AWR-002** (ubiquitous) — The template mirror's capability-binding table
  (`internal/template/templates/AGENTS.md.tmpl:21-25`) shall carry the same one-line
  `worktree-entry` row. The two copies are an intentional fork (no byte-identity invariant), so
  this requirement is judged against the mirror's own table, independently of REQ-AWR-001's
  outcome.

- **REQ-AWR-003** (ubiquitous) — The template mirror's `## 11. moai CLI Verbs` table
  (`internal/template/templates/AGENTS.md.tmpl:293-303`) shall carry exactly one row whose first
  cell is `` `moai codex` ``, and the row's purpose text shall name the launch/readout verbs and
  state the resolve-only limit of `-w <worktree>` in the same line. The guard
  `grep -c '^| \`moai codex\`' internal/template/templates/AGENTS.md.tmpl` shall return `1`.

- **REQ-AWR-004** (unwanted) — The card shall not change Go source, shall not touch
  `worktree-integration.md` or `session-handoff-examples.md` (card t1072's files), and shall not
  alter any launcher behavior, flag, diagnostic string, or threshold. The changes are prose rows
  in exactly two files.

- **REQ-AWR-005** (event-driven) — When any file under `internal/template/templates/` is edited,
  the embedded copy shall be regenerated with `make build` before the change is judged complete,
  so the deployed binary and the committed template do not diverge (Template-First cycle; the
  `agents-emit-check` inside `make build` is read-only and does not discharge this — the build
  itself is the regeneration act).

### C.1 Wording decisions taken (recorded for the plan-auditor)

- The row's capability token is `worktree-entry` (lowercase, hyphenated, matching the existing
  rows' `kebab-case` style). The machine guard keys on `^| worktree-entry |`, which is
  unambiguous against the existing rows and against the §3 prose that already uses the words
  "worktree" freely.
- The row keeps the existing table's semantic: column 2 names the Claude-side launcher forms
  (`moai cc -w`), column 3 names the fallback for a harness without the capability — and the
  Codex launcher's presence is stated inside column 2's scope in the same line, because the gap
  is that the *contract* does not name the Codex form anywhere. Placing the Codex form in
  column 3 would misstate it as an absence-path rather than a present capability.
- REQ-AWR-001 says "exactly one" and REQ-AWR-003 says "exactly one" so the guards are
  falsifiable in both directions: a duplicate row is as much a defect as an absent one.

---

## §D Out of Scope

### Out of Scope — sibling doctrine files (card t1072)

- `worktree-integration.md` — its § Terminology Glossary carries the L1/L2 tier vocabulary this
  card references but does not amend.
- `session-handoff-examples.md` — its Block 0 worktree-anchored resume forms are t1072's surface.
- Neither file is read-and-fixed as a side effect even where this card's re-measurement touches
  their subject matter.

### Out of Scope — launcher behavior

- No change to `internal/cli/codex_launcher.go` or any Go source: flag handling, the
  resolve-don't-create diagnostics (`:47`, `:52`, `:289-297`), and the verb router are all
  frozen by REQ-AWR-004.
- No new `moai worktree` verb, no worktree-creation capability for the Codex launcher — the
  resolve-only limit is being *documented*, not changed.

### Out of Scope — byte-identity between the copies

- The root `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` are an intentional fork:
  the root copy is the maintainer dogfood contract, the mirror is the neutral distributable
  layer. Each copy is edited and judged separately; byte-identity is neither claimed nor
  restored by this card.

---
