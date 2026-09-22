# acceptance.md — SPEC-AGENTS-WORKTREE-ROW-001

## §A AC-REQ traceability

| AC | REQ | Verification surface |
|----|-----|----------------------|
| AC-AWR-001 | REQ-AWR-001 | machine grep on the root contract (post-edit) |
| AC-AWR-002 | REQ-AWR-002 | machine grep on the template mirror (post-edit) |
| AC-AWR-003 | REQ-AWR-003 | machine grep on the mirror Verbs table (post-edit) |
| AC-AWR-004 | REQ-AWR-004 | scope diff + untouched-file greps |
| AC-AWR-005 | REQ-AWR-005 | `make build` exit code + regeneration check |

## §D AC Matrix — Given / When / Then

### AC-AWR-001 — the root contract carries the worktree-entry row

**Given** the landing tree is `WT-cross-harness-row` at or beyond the plan-phase commit, with
`AGENTS.md` at its capability table.
**When** the run-phase lands M1 and the guard runs:
`grep -c '^| worktree-entry |' AGENTS.md`
**Then** the command exits 0 and prints exactly `1`; the row's line contains the substring
`moai codex -w` and the phrase `never creates one`; and the three pre-existing rows remain
present at their plan-phase counts — `question-channel` = 1, `task-list` = 1, `design-sync` = 1
(the same counts measured pre-edit). These are presence counts: they prove the rows were not
removed, not that their cell text is untouched — content identity of pre-existing rows is
carried by plan.md §G's no-rewording constraint, checked at run-phase diff review.

RED-now cell (measured 2026-09-22, `cd99336bf`): the guard command printed `0` (exit 1). Green
path: run-phase M1 flips it to `1` — no other change this work makes affects this grep.

### AC-AWR-002 — the mirror's capability table carries the row independently

**Given** the template mirror `internal/template/templates/AGENTS.md.tmpl` is edited in M2(a).
**When** the guard runs: `grep -c '^| worktree-entry |' internal/template/templates/AGENTS.md.tmpl`
**Then** it prints exactly `1`. Byte-identity with the root file is NOT asserted — the two
copies are an intentional fork (spec §D), so this AC judges only the row's presence and shape in
this copy.

RED-now cell (measured 2026-09-22, `cd99336bf`): `0`. Green path: M2(a).

### AC-AWR-003 — the Verbs table registers `moai codex` with its resolve-only limit

**Given** the mirror's `## 11. moai CLI Verbs` table is edited in M2(b).
**When** the guard runs: `grep -c '^| \`moai codex\`' internal/template/templates/AGENTS.md.tmpl`
**Then** it prints exactly `1`; the row's text names both the launch verb (`cli`) and the
readout verb (`status`), and carries `never creates` (the resolve-only limit); and the nine
pre-existing verb rows remain present at their plan-phase counts (`moai init` = 1 and the
table's row count unchanged at 9) — presence counts, carrying plan.md §G's no-rewording
constraint the same way as AC-AWR-001.

RED-now cell (measured 2026-09-22, `cd99336bf`): `0`. Green path: M2(b).

### AC-AWR-004 — the card touches nothing outside its three files

**Given** the plan-phase and run-phase commits exist on the branch, and the t1072 fence files
exist at BOTH copies — dogfood (`.claude/rules/moai/workflow/worktree-integration.md`,
`.claude/rules/moai/workflow/session-handoff-examples.md`) and template-mirror
(`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`,
`internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md`) —
verified at plan-audit iter-1.
**When** the scope diff runs:
`git diff --name-only <base>..HEAD | grep -v '^\.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/' | grep -v '^AGENTS\.md$' | grep -v '^internal/template/templates/AGENTS\.md\.tmpl$' | wc -l`
**Then** it prints `0`; both fence-file copies show no diff in the range
(`git log --oneline <base>..HEAD -- <full path>` empty, per copy); and no line of
`internal/cli/codex_launcher.go` differs in the range.

**Adoption cells** (verification-completeness.md §2): this AC is **green-now by construction** —
at plan-phase close (`fdaaa27ef`) the commit range touches only the SPEC directory, so the
filtered residual is 0 without this work constraining anything. Its red IS constructible and
MUST be observed once at run-phase before the green is trusted: after the M1/M2 contract
commits land, weaken the filter by one exclusion
(`git diff --name-only <base>..HEAD | grep -v '^\.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/' | wc -l`)
and observe `2` (the two contract files) — then observe the strict filter print `0`. The
weakened-filter `2` is the red this instrument can show; without it the strict `0` is an
empty-sweep pass.

### AC-AWR-005 — the build recompiles against the committed template [regression-guard]

**Classification: regression-guard / process gate** (verification-completeness.md §2.1). The
red is unconstructible for this change class: no step of the build chain observes
`AGENTS.md.tmpl` drift — `agents-emit-check`/`commands-emit-check` cover the `.codex/` and
`.agents/` families only (`Makefile:34`, measured at plan-audit iter-1), and `//go:embed`
embeds committed bytes without parsing. Per §2.1 this AC is NOT recorded as a proof-pass; its
run-phase result is recorded as a process-gate discharge.

**Given** M2 edited a file under `internal/template/templates/`.
**When** `make build` runs (M3) and the exit code is captured.
**Then** the command exits `0` — the binary was recompiled against the committed template
(the Template-First cycle's regeneration step, per REQ-AWR-005). No claim is made beyond this:
nothing in this build output distinguishes a build that compiled the new rows from one that
did not, and this AC does not assert that it does.

## §D.1 Edge cases

- **Duplicate-row mutant**: a run-phase that appends the row twice satisfies a `>= 1` check; the
  `== 1` count in AC-AWR-001/002/003 kills this mutant ("exactly one" per REQ wording).
- **Comment-hit mutant**: wording the row inside a prose paragraph instead of the table
  satisfies a substring grep but fails the `^| worktree-entry |` row-anchored pattern — the
  anchor's leading pipe is the discriminator against the C6 comment-hit class.
- **Wrong-column mutant**: placing the `moai codex` form in column 3 (the absence-path column)
  would still pass the row anchor but contradicts spec §C.1's wording decision; caught at
  plan-audit wording review, not by the guard — recorded here so the reviewer looks.

## §D.2 Quality gates

- Lint: `go run ./cmd/moai spec lint SPEC-AGENTS-WORKTREE-ROW-001` exit 0 (E4).
- Build: `make build` exit 0 (E3).
- Neutrality: the new rows must not carry internal SPEC ids, commit SHAs, or macOS-biased paths
  (§25 content classes — the Verbs row and the capability row are distributed content).

## §D.3 Definition of Done

All five ACs PASS with verbatim outputs recorded in progress.md §E.1's plan-phase evidence;
the single run-phase commit carries the card trailers; nothing is pushed (lead owns push).
