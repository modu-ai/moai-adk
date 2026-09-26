# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — implementation plan

> **[HARD] This plan is a carve artifact, not a completed plan phase.** It preserves the
> milestones the parent SPEC authored for this scope and names what remains open. Card **t1259**
> owns completing it: re-sequencing the milestones for this SPEC's own dependency order,
> confirming the Tier judgment, running the plan-audit, and taking the two debt-unfold decisions
> `acceptance.md` records. Do not read the milestone list below as a finished sequence.

## §A Context

Carved from `SPEC-INSTRUCTION-FILES-UNIFY-001` at commit `1140bcd1d` by operator decision
(2026-09-26). Scope: 9 requirements, 7 acceptance criteria — the user-owned-file half of the
instruction-file unification. Provisional Tier M.

Work location is card t1259's own worktree, created from `develop` per the lane protocol. The
parent SPEC's card is t1243.

## §B Blocking dependencies

**[HARD] `SPEC-INSTRUCTION-FILES-UNIFY-001` M2 must land before this SPEC's fallback-advisory
work starts.** Both SPECs edit the same function — the local-instruction loop in
`internal/cli/codex_launcher.go`. That SPEC changes its **iteration order** (REQ-IFU-006); this
one adds the **deprecation advisory on its fallback branch** (REQ-IFU-007). The two edits are
compatible but not independent: this SPEC's advisory work builds on the reordered loop.
Concurrent lanes against that one function is the case to avoid, and the lead's dispatch
decision is what prevents it.

**[HARD] t1175 (rules diet) must land on develop before the run phase starts** — inherited from
the parent SPEC, because the docs-site milestone and the `CLAUDE.local.md` relocation both touch
surfaces t1175 is rewriting.

**Source locations are cited by symbol, not by line.** `origin/develop` is roughly 93 commits
ahead of the carve base and has already moved one location the parent SPEC cited (card t1224
shifted `frozenInstructionFiles` while leaving the symbol intact). Every figure in the parent's
`research.md` §B is attributed to base develop `553e224f3` on 2026-09-26 and is re-measured
before any milestone begins.

## §C Constraints

- **Migration is never implicit.** No hook, no `moai update`, no `moai init` may perform the
  rename. This is REQ-IFU-011's substance, not a style preference.
- **The repository's own migration is operator-gated.** See M4.
- **Template-First.** `internal/template/templates/` changes first, then `make build`, then the
  root copies. Never the reverse.
- **Two mirrors, two commands.** Every deployed-file check runs against both paths with separate
  exit codes.
- **Every `go test` assertion carries `-v` and a `--- PASS:` read.** Exit `0` alone is not
  evidence a test ran.
- **The lane does not push.** Integration is a lead-granted window; push is the lead's batch.

## §D Milestones (transferred; sequence open)

Transferred from the parent SPEC's M3 / M5 / M6 and renumbered as M1-M4 for this SPEC only. The
parent's M1 (measurements), M2 (read order, guards, learner target) and M3 (contract body) stay
with it.

### M1 — the Codex fallback branch and its advisory

`codex_launcher.go`'s local-instruction loop gains the fallback advisory naming
`moai migrate local-instructions`, and the provenance preamble is asserted to carry the literal
filename actually read (REQ-IFU-007, REQ-IFU-008; `AC-IFU-011`). Builds on the parent SPEC's M2
iteration-order change (§B).

Open for t1259: research.md Q4 of the parent SPEC — whether the fallback branch has a
diagnostic surface on which to emit the advisory without polluting `developer_instructions` — is
unanswered and is this milestone's first task.

### M2 — the migration verb and the advisories

`moai migrate local-instructions` (REQ-IFU-009, REQ-IFU-010), the `moai update` advisory
(REQ-IFU-011), and the `moai doctor` advisory (REQ-IFU-012). The no-coexistence invariant is the
load-bearing part: Claude reads `CLAUDE.local.md` on its own, so leaving the original in place
double-loads the same content.

Model the verb on `migrate_agency_*` — the existing precedent for a move-plus-backup verb.

### M3 — docs-site, four locales

Six pages × four locales (REQ-IFU-020; `AC-IFU-023`). Korean is the canonical source per the
project's i18n rules; en/ja/zh derive. Same-PR obligation applies — all four locales or none.

### M4 — this repository's own migration (operator-gated)

**[HARD] Proceeds only after explicit operator confirmation in the lane.** It is the riskiest
step: live maintainer doctrine, git-tracked despite being gitignored, moving to a file under
40,000 characters with operational procedure split into `.moai/docs/`.

- The migrating copy is the **`develop`-committed** one, per that file's own §0.1 discriminant
  — not the primary checkout's working copy, which §0.2 forbids citing and §0.4 documents as
  permanently modified by design. Measured 2026-09-26: 44,381 characters (`wc -m`), so the
  required reduction is at least 4,381 characters, roughly 10%.
- Split procedure out, leave rules behind (REQ-IFU-021; `AC-IFU-007`).
- Rewrite §0 to name `AGENTS.local.md` (REQ-IFU-022; `AC-IFU-024`) — §0 is the discriminator for
  which copy is canonical, and it names the filename directly.
- Verify the parent SPEC's M1a worktree answer actually holds for this repository's own
  worktrees before relying on it; the lanes read this file.

## §E Self-verification

Per-milestone: the affected packages only (`go test ./internal/<pkg>/...`), never
`go test ./...` locally. The full-suite verdict is CI's, on the PR head, in a clean environment.

At close: the `acceptance.md` §D.2 traceability diff command, plus a separate re-run of every
two-mirror criterion.

Open for t1259: this SPEC has no whole-change CI criterion of its own (the parent's
`AC-IFU-025` stayed with the parent, since it asserts that SPEC's always-loaded budget clause).

## §F Anti-patterns

- Letting `moai update` do the migration "since it is already touching the tree".
- Starting M4 without the operator's confirmation because the earlier milestones went well.
- Migrating the primary checkout's *working* copy of `CLAUDE.local.md` because it is the one
  `wc -m` reaches without a `git show`. §0.2 forbids citing it, and it already satisfies the cap
  — so measuring it discharges `AC-IFU-007` without doing any work.
- Mixing bytes and characters when reporting the reduction. The cap is characters (`wc -m`).
- Reading a `go test` exit code as evidence the test ran. `-run` on a pattern matching nothing
  exits `0` and prints `no tests to run`.
- Writing the advisory into the launcher loop before the parent SPEC's M2 order change has
  landed (§B).
- Restoring or repairing a user's modified local instruction file automatically. It may carry
  runtime-written values, so the restore is itself destructive.
- Resolving the worktree duplicate-load here. It is card t1219's; this SPEC's obligation is only
  not to make it worse (REQ-IFU-010).

## §G Cross-references

- `.moai/specs/SPEC-INSTRUCTION-FILES-UNIFY-001/` — the parent SPEC; its `design.md` §C
  (read-order analysis) and `research.md` §B (source survey) are shared context this SPEC reads
  rather than duplicates.
- `.moai/reports/t1243/plan-audit-iter1.md` — the audit whose D2 arithmetic forced the carve and
  whose D3 finding is repaired in `spec.md` REQ-IFU-021 and `acceptance.md` AC-IFU-007.
- `.moai/reports/t1243/m0/verdict.md` — the M0 measurement the parent SPEC's design rests on.
- `CLAUDE.local.md` §0 — the canonical-copy discriminant M4 obeys.
- `.claude/rules/local/gitflow-lane-protocol.md` — the lane and integration-window discipline.
