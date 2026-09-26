# SPEC-MCP-WORKTREE-UNTRACKED-001 — Plan

Card: t1202 | Issue: #1716 | Branch: WT-worktree-moai-root | Base: origin/develop `df526c9a9` | Tier: M

## §A Context

The `project_root` validator rejects a linked worktree of a repository that
keeps `.moai/` untracked, and omitting the input silently acts on the primary
checkout. Premise measured LIVE (`.moai/reports/t1202/verdict.md`). Design (a)
— validator-side acceptance plus a tree/`.moai` root split — is recommended in
`spec.md` §2.3; design (b) is not planned.

## §B Known Issues

- `internal/cli/codex_audit_launch.go` already carries a same-repository check
  (git common dir + `git worktree list --porcelain`) for a different purpose.
  Reuse its predicate shape rather than inventing a second one; do not change its
  behavior.
- `internal/core/git/checkout.go` `ResolveGitDirs` / `IsPrimaryCheckout` already
  resolve git dir and common dir with an older-git fallback. Prefer reusing them.
- Which handlers perform `.moai`-relative reads is not yet inventoried; M1
  produces the list before any handler changes. Do not assume the list from the
  tool names alone.

## §C Open Questions for the operator (Kickoff)

1. [NEEDS CLARIFICATION: `.moai` root for state tools] REQ-MWU-007 routes
   `.moai` reads to the primary checkout. Alternative: the state tools
   (`spec_*`, `verify_*`) reject such a worktree with a message pointing at the
   primary, and only tree tools accept it. Recommendation: route to the primary
   (keeps one `.moai`, matches how the repository is actually organized).
2. [NEEDS CLARIFICATION: snapshot writes] Under REQ-MWU-007, `verify_snapshot`
   with `command` writes into the primary's `.moai`. Snapshot keys are HEAD SHAs,
   so a worktree's snapshot does not collide with the primary's — is writing
   there acceptable?
3. [NEEDS CLARIFICATION: registration requirement] REQ-MWU-002 requires the tree
   to appear in `git worktree list`. A tree whose admin entry was pruned is
   therefore rejected. Accept this strictness?
4. Design (b) is recommended against. Should a backlog card still record it as a
   rejected alternative, or is the `spec.md` §2 record sufficient?
5. `moai worktree new` branched this card's tree from `main` rather than
   `develop` (same class as t1159). Separate card?

## §D Constraints

- Reject-never-fall-back stays binding (spec.md §4).
- Scoped verification only: `go test ./internal/cli/ -run '<new tests>'`, the
  existing `mcp_project_root` tests, `./internal/template/...` for the mirror and
  neutrality guards. No local full-suite run.
- Tests that set `GIT_*` via `t.Setenv` are non-parallel.
- Template-first for the rule file: edit the template mirror and the local copy
  together, then `make build`.

## §E Self-Verification (run-phase exit)

- E1: AC-MWU-001..010 matrix with verbatim command output.
- E2: RED commit precedes the fix commit (AC-MWU-001; `verification-claim-integrity.md` §2.3).
- E3: `diff` of the two rule copies exits 0.
- E4: `go vet ./internal/cli/` and `golangci-lint run ./internal/cli/...` clean.

## §F Milestones (ordered by decision-reversibility)

### M1 — Priority High — Decide the data shape and inventory `.moai` reads

- Fix the resolver's result shape: a tree root plus a `.moai` root, and the
  provenance fields that expose both (REQ-MWU-007/008). This is the decision
  most likely to change and must be settled before code.
- Inventory, per tool in the `project_root` family, every `.moai`-relative read
  it performs on the call's behalf (catalogue, snapshots, config sections).
  Record the list in progress.md §E.2.

### M2 — Priority High — RED reproduction, committed alone

- Add the reproduction test (AC-MWU-001) and the diff-collection test
  (AC-MWU-002) against the unchanged validator; capture the failing output;
  commit before any fix (baseline-first ordering).

### M3 — Priority High — Validator: linked-worktree acceptance

- Implement REQ-MWU-002..006 on the no-`.moai` path only: canonical top-level
  check, linked-worktree check, registration check, primary `.moai` check,
  scrubbed git environment, fail-closed on any git error. Error text names the
  failed condition.

### M4 — Priority Medium — Wire the `.moai` root into the state reads

- Route each read found in M1 to the `.moai` root; tree operations keep the tree
  root. Extend the `_root` provenance block.

### M5 — Priority Medium — Descriptions and rule doc

- Update the shared `project_root` input description text.
- Update § The `project_root` input in both copies of `moai-mcp-tools.md`
  (add the linked-worktree row to the situation table; one short paragraph on
  the tree/`.moai` split). Neutral wording — no SPEC IDs, card ids, dates.
- `make build`.

### M6 — Priority Low — Verification batch

- Run the E1–E4 batch in one turn; record evidence in progress.md §E.2.

## §G Anti-Patterns

- Accepting any directory inside the repository (subdirectories) — only a
  worktree top level qualifies.
- Falling back to the primary when git inspection errors — that is the silent
  wrong-tree defect in a new place.
- Reading the worktree's absent `.moai` and returning an empty catalogue as
  success.
- Comparing un-canonicalized paths against `git worktree list` output.

## §H Cross-References

- `internal/cli/mcp_project_root.go` — validator and resolvers
- `internal/cli/codex_audit_launch.go` — existing same-repository predicate
- `internal/core/git/checkout.go` — git dir / common dir resolution
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input (+ template mirror)
- SPEC-MCP-WORKTREE-ROOT-001 — origin of the `project_root` input
- `.moai/reports/t1202/verdict.md` — premise measurement
