# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — implementation plan

> **Plan phase completed by card t1259 (v0.2.0).** The milestone list below is re-sequenced for
> this SPEC's own dependency order and is no longer the transferred one. The Tier is raised to
> **L** on measured file count, so `design.md` and `research.md` join the artifact set.

## §A Context

Carved from `SPEC-INSTRUCTION-FILES-UNIFY-001` at commit `1140bcd1d` by operator decision
(2026-09-26). Scope: 9 requirements (11 clauses, `REQ-IFU-010` carrying two) and 10 acceptance criteria — the
user-owned-file half of the instruction-file unification.

**Tier L, confirmed at v0.2.0 against the measured file count.** The carve's provisional `M`
rested on the requirement count; the binding axis is files affected, and the Tier L threshold is
`> 15`. Measured in this worktree 2026-09-26: the docs-site milestone alone touches **24** files
(6 pages × 4 locales, each resolved by `find docs-site/content -name '<page>.md'`), before
`internal/cli` (launcher, contract, the new verb and its tests, `update`, `doctor`),
`CLAUDE.local.md` → `AGENTS.local.md`, and the `.moai/docs/` relocations — roughly 35 in total.
The LOC axis agrees: the verb is modelled on `migrate_agency.go`, which is 25,790 bytes with a
30,580-byte test file beside it.

Two consequences follow and are not optional: the artifact set becomes **5 files** (`design.md`
and `research.md` added, authored at v0.2.0 and deliberately thin — they carry this SPEC's own
decisions and cross-reference the parent for shared context rather than duplicating it), and the
plan-auditor PASS threshold rises from 0.80 to **0.85**.

Work location is card t1259's own worktree, created from `develop` per the lane protocol. The
parent SPEC's card is t1243.

## §B Blocking dependencies

**[HARD] `SPEC-INSTRUCTION-FILES-UNIFY-001` M2 must land before this SPEC's M2 (the
fallback-advisory work) starts.** M1 is unblocked by it, which is why M1 now runs first. Both SPECs edit the same function — the local-instruction loop in
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

## §D Milestones (re-sequenced at v0.2.0)

Ordered by decision-reversibility and by dependency, not by the parent SPEC's numbering. The
mapping from the carve's transferred list, so nothing is lost by the renumber:

| New | Was | Moved because |
|---|---|---|
| **M1** — migration verb + advisories | M2 | Nothing depends on it; `M2` and `M3` both depend on it. It was second only because the parent authored it second. |
| **M2** — Codex fallback advisory | M1 | Blocked on the parent SPEC's M2 landing (§B). Sequencing it first would have stalled the whole card behind another card's merge. |
| **M3** — this repository's own migration | M4 | Unchanged in position, but now explicitly downstream of M1: the migration is performed **with** the verb M1 builds, which is the dogfood and is what proves the verb on a real 44,740-character file rather than a fixture. |
| **M4** — docs-site, four locales | M3 | Moved last. It documents the verb's name, its refusal behaviour, and the advisory text — all of which M1 and M2 settle. Writing 24 files against a design that has not landed is the expensive way to discover a rename. |

### M1 — the migration verb and the advisories

`moai migrate local-instructions` (`REQ-IFU-009`, `REQ-IFU-010a`, `REQ-IFU-010b`), the
`moai update` advisory (`REQ-IFU-011`), and the `moai doctor` advisory (`REQ-IFU-012`).
Criteria: `AC-IFU-013`, `AC-IFU-014`, `AC-IFU-015`, `AC-IFU-030`.

The no-coexistence invariant is the load-bearing part: Claude reads `CLAUDE.local.md` on its own,
so leaving the original in place double-loads the same content. **The refusal branch
(`REQ-IFU-010b`) is written first**, before the happy path — it is the clause that protects a
user-authored file, and writing it second is how it ends up as an afterthought on a verb that
already works.

Model the verb on `migrate_agency_*`: `internal/cli/migrate_agency.go` is the existing
move-plus-backup precedent, with a separate idempotency test file
(`migrate_agency_idempotent_test.go`) worth mirroring — re-running the verb after a successful
migration must be a clean no-op, not a second backup.

### M2 — the Codex fallback branch and its advisory

`codex_launcher.go`'s local-instruction loop emits the fallback advisory naming
`moai migrate local-instructions`, and the provenance preamble is asserted to carry the literal
filename actually read (`REQ-IFU-007`, `REQ-IFU-008`; `AC-IFU-011`, `AC-IFU-029`).

**Blocked on the parent SPEC's M2 iteration-order change (§B)** — this is the one hard ordering
constraint the card carries across SPEC boundaries. It depends on M1 only for the advisory's
wording: the advisory names the verb, so the verb's final spelling must be settled first.

The parent's `research.md` Q4 — whether the fallback branch has a diagnostic surface — **is
answered** (`research.md` §A, `design.md` §B). It is no longer this milestone's first task.

### M3 — this repository's own migration (operator-gated)

**[HARD] Proceeds only after explicit operator confirmation in the lane**, recorded with the turn
it arrived in. It is the riskiest step: live maintainer doctrine, git-tracked despite being
gitignored, and read by every lane. `REQ-IFU-021`, `REQ-IFU-022`; `AC-IFU-007`, `AC-IFU-024`.

- The migrating copy is the **`develop`-committed** one, per that file's own §0.1 discriminant —
  not the primary checkout's working copy, which §0.2 forbids citing and §0.4 documents as
  permanently modified by design.
- **Re-measure the before-value at this milestone.** It was 44,381 at the carve and **44,740**
  when t1259 measured it; it will have moved again. The bound that does not drift is
  `after <= 39,999` (`AC-IFU-007` v0.2.0 note).
- Split operational procedure out into `.moai/docs/`, leave the rules behind. Do not re-decide
  what any relocated rule says (spec.md §D).
- Rewrite §0 to name `AGENTS.local.md` (`AC-IFU-024`) — §0 is the discriminator for which copy is
  canonical, and it names the filename directly.
- Perform the migration **with the M1 verb**, not by hand. A hand-migration proves nothing about
  the verb and forfeits the one real-file test this card can run.
- Verify the parent SPEC's M1a worktree answer holds for this repository's own worktrees before
  relying on it; the lanes read this file.

### M4 — docs-site, four locales

Six page **paths** × four locales = 24 files (`REQ-IFU-020`; `AC-IFU-023`). Korean is the
canonical source per the project's i18n rules; en/ja/zh derive. Same-PR obligation: all four
locales or none.

**[HARD] The `memory` page is `claude-code/context-memory/memory.md`, not
`cli-reference/memory.md`.** Both exist in all four locales and the requirement previously named
only the stem. The former carries the CLAUDE.md discussion (28 matches in the ko copy); the
latter is the `moai memory` CLI reference and carries none. Editing the wrong one would satisfy a
stem-based grep while leaving the documented structure untouched.

### Close

`AC-IFU-031` (whole-change CI on the PR head) is read after all four milestones land. It is a
close gate, not a milestone gate.
## §E Self-verification

Per-milestone: the affected packages only (`go test ./internal/<pkg>/...`), never
`go test ./...` locally. The full-suite verdict is CI's, on the PR head, in a clean environment.

At close: the `acceptance.md` §D.2 traceability diff command, plus a separate re-run of every
two-mirror criterion.

Closed at v0.2.0: this SPEC now carries its own whole-change assertion, `AC-IFU-031`, read from
the PR head's CI run. The parent's `AC-IFU-025` stayed with the parent because it asserts that
SPEC's always-loaded budget clause, which this SPEC does not own.

## §F Anti-patterns

- Letting `moai update` do the migration "since it is already touching the tree".
- Starting M3 without the operator's confirmation because the earlier milestones went well.
- Hand-editing this repository's `CLAUDE.local.md` in M3 instead of running the M1 verb on it.
- Citing an absolute before-value for `CLAUDE.local.md` from this document rather than
  re-measuring. It moved 359 characters between the carve and the plan phase.
- Editing `cli-reference/memory.md` in M4 because the stem matched.
- Migrating the primary checkout's *working* copy of `CLAUDE.local.md` because it is the one
  `wc -m` reaches without a `git show`. §0.2 forbids citing it, and it already satisfies the cap
  — so measuring it discharges `AC-IFU-007` without doing any work.
- Mixing bytes and characters when reporting the reduction. The cap is characters (`wc -m`).
- Reading a `go test` exit code as evidence the test ran. `-run` on a pattern matching nothing
  exits `0` and prints `no tests to run`.
- Writing the advisory into the launcher loop before the parent SPEC's M2 order change has
  landed (§B).
- Emitting the deprecation advisory into the `developer_instructions` payload. It is operator
  diagnostics, not model context (design.md §B).
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
