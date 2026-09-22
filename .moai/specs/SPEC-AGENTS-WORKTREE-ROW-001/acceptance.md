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
`moai codex -w` and the phrase `never creates one`; and the three pre-existing rows
(`question-channel`, `task-list`, `design-sync`) are unchanged (verified by
`grep -c '^| question-channel |' AGENTS.md` still printing `1`).

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
**Then** it prints exactly `1`; the row's text contains `never creates` (the resolve-only limit)
and names the readout verb (`status`); and the nine pre-existing verb rows are unchanged
(verified by `grep -c '^| \`moai init' internal/template/templates/AGENTS.md.tmpl` still
printing `1`).

RED-now cell (measured 2026-09-22, `cd99336bf`): `0`. Green path: M2(b).

### AC-AWR-004 — the card touches nothing outside its two files

**Given** the plan-phase and run-phase commits exist on the branch.
**When** the scope diff runs:
`git diff --name-only <base>..HEAD | grep -v '^\.moai/specs/SPEC-AGENTS-WORKTREE-ROW-001/' | grep -v '^AGENTS\.md$' | grep -v '^internal/template/templates/AGENTS\.md\.tmpl$' | wc -l`
**Then** it prints `0`; and `worktree-integration.md` / `session-handoff-examples.md` show no
diff in the range (`git log --oneline <base>..HEAD -- <each path>` empty for both); and no line
of `internal/cli/codex_launcher.go` differs in the range.

### AC-AWR-005 — the embedded copy is regenerated with the source

**Given** M2 edited a file under `internal/template/templates/`.
**When** `make build` runs (M3) and the E3 measurement captures its exit code.
**Then** the command exits `0`; the `agents-emit-check` stage inside it reports no drift; and a
second run of `make build` is a no-op (exit 0, no rewrites reported) — proving the committed
template and the embedded copy are in step, not that the first run silently failed.

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

All five ACs PASS with verbatim outputs recorded in progress.md §E.2's plan-phase evidence;
the single run-phase commit carries the card trailers; nothing is pushed (lead owns push).
