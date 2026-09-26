# SPEC-INSTRUCTION-FILES-UNIFY-001 — implementation plan

## §A Context

Card t1243, worktree `.claude/worktrees/t1243`, branch `WT-instruction-files`, base
develop `553e224f3`. Tier L, class C. Plan phase only; the run phase is blocked (§C).

## §B Blocking dependencies and baseline

**[HARD] t1175 (rules diet) must land on develop before the run phase starts.** t1175 is
rewriting the always-loaded rule tree, which is the same surface this SPEC's `AGENTS.md`
reconciliation touches. Running both concurrently produces a merge whose combined state
neither card measured.

**The baseline is re-measured after t1175 lands.** Every figure in research.md §B was read
against `553e224f3` and is attributed to that tree only. After t1175 lands, the byte counts,
section counts, and line numbers are re-read before any milestone begins — they are the
inputs to REQ-IFU-018 and REQ-IFU-019, and a carried-over figure would make the cap check
unattributed.

`origin/develop` has already moved 8 commits ahead of this worktree's base (card t1229,
test isolation). No scope overlap was observed. Absorption happens at the integration
window, not now.

## §C Sequencing risk — SPEC A

The approved design names SPEC A (codex factory retirement) as overlapping this work at
`AGENTS.md` §3 and §8, and recommends A landing first with B absorbing it. **SPEC A does
not exist as a card.** Neither order is assumed here:

- If A lands first, M4 (the `AGENTS.md` body rewrite) absorbs its §3/§8 changes and the
  reconciliation is done once.
- If B lands first, A rewrites §3/§8 again on top of the reconciled body, and the `-f`
  removal from §8 that this SPEC performs may be re-touched.
- If they land concurrently, the §3/§8 region conflicts. This is the case to avoid, and
  the lead's dispatch decision is what prevents it.

Surfacing this is the plan's job; choosing is the lead's.

## §D Constraints

- **Template-First.** `internal/template/templates/` changes first, then `make build`, then
  the root copies. Never the reverse.
- **Two mirrors, two commands.** Every deployed-file check runs against both paths with
  separate exit codes (design.md §E).
- **No sweep-staging.** Explicit pathspec only; `git status --short` re-read immediately
  before staging.
- **The lane does not push.** Integration is a lead-granted window; push is the lead's
  batch.

## §E Milestones

Ordered by decision-reversibility: the unsettled design decision first, then the mechanical
code changes, then the body rewrite that depends on the decision, then documentation, and
this repository's own migration last.

### M1 — settle the two unmeasured questions (measurement, no code)

Two independent measurements, both recorded in `progress.md` §E.2 with command and verbatim
output **whichever way they come out**. A negative result is a result; neither is retried
until it produces the convenient answer.

**M1a — real-worktree ancestor discovery (AC-IFU-021).** Answers research.md Q1 and Q2. The
M0 P7 line-3 observation is a synthetic-fixture observation and, per the lead's ruling, may
not serve as a premise — this measurement is what settles it. Method mirrors the M0 probes
(headless `claude -p`, distinct sentinels, throwaway repository) with one change that is the
whole point: the tree is created by a real `git worktree add`, so `.git` in it is a file
rather than a directory. Outcome routes per design.md §A.5.

If discovery IS confirmed, the finding is handed to card **t1219** with its evidence — the
mechanism that carries local instructions into a worktree is the same one that double-loads
them (design.md §A.3). This SPEC does not resolve the duplicate.

**M1b — Codex discovery of `AGENTS.local.md` (AC-IFU-022).** Answers research.md Q5. The
criterion cannot be "run Codex and see that it works": the failure mode is silent tail
truncation, so a passing session proves nothing. Tail sentinels in both `AGENTS.md` and
`AGENTS.local.md` make truncation observable. A `LOCAL_HEAD` hit blocks — it means the
content is counted twice and the design's budget arithmetic is wrong.

Gate: M2 does not start until both measurements' evidence is on disk. M1a's outcome does not
gate M2 (design.md §A.5: every branch leaves the primary-checkout shape unchanged); M1b's
does, because a positive result changes the byte budget M4 has to fit inside.

### M2 — Codex read order, guard set, learner target

Three independent, mechanically small changes, each with its test:

- `codex_launcher.go:125` iteration order + fallback advisory (REQ-IFU-006, -007, -008).
- `pre_tool.go:1231` frozen set gains two basenames (REQ-IFU-013).
- `curator/dispatch.go:43` Tier-3 path (REQ-IFU-014).

Invert `codex_contract_link_test.go:202` and the `codex_local_instructions_test.go`
read-order assertions here, not later — leaving them for M4 would put the tree red across
two milestones.

### M3 — the migration verb and the advisories

`moai migrate local-instructions` (REQ-IFU-009, -010), the `moai update` advisory
(REQ-IFU-011), and the `moai doctor` advisory (REQ-IFU-012). The no-coexistence invariant
is the load-bearing part: Claude reads `CLAUDE.local.md` on its own, so leaving the original
in place double-loads the same content.

Model the verb on `migrate_agency_*` — the existing precedent for a move-plus-backup verb.

### M4 — the `AGENTS.md` body rewrite and the `CLAUDE.md` thinning

Depends on M1 (the worktree paragraph the body must state) and on the §C sequencing
decision.

- Reconcile root (8 sections) against template (12) onto one **section set** (REQ-IFU-018);
  decide whether the four template-only reference sections belong in a contract
  (research.md Q3), under both ceilings.
- **[HARD] Do not touch the mirror's filename.** It is `AGENTS.md.tmpl`, and the suffix is
  what keeps it out of Codex's filename discovery. Card t925 renamed it after measuring the
  failure: chain sum 33,738 / 32,768, the mirror's last section dropped and the preceding
  one cut mid table row, no warning, exit 0. Reconciliation means the section set, never the
  name (design.md §D.2, AC-IFU-004).
- Content is NOT reconciled: the two mirrors diverge by 46 intentional lines, so the check
  is a section-set diff (AC-IFU-008), not a byte diff.
- Update the touched contract constants: `codexLinkAgentsDirective` and
  `codexCreatedClaudeBody` (`codex_contract.go:38`, `:46-47`) — the created stub must emit
  the new two-import shape or it will not satisfy AC-IFU-002 (design.md §C).
- Correct the budget statement (REQ-IFU-017); keep the truncation statement.
- Drop `-f` from §8; fix the §3 "Codex lanes" / `-w` inconsistency.
- Thin `CLAUDE.md` to import + mechanism layer + import (REQ-IFU-002).
- Add the C5 neutrality allowlist change (REQ-IFU-015).
- `make build`, then both mirrors, then **both** ceiling checks: per-file 24,576
  (AC-IFU-005) and nested-sum 32,768 (AC-IFU-006). They are different limits with different
  owners and neither substitutes for the other. Raising `project_doc_max_bytes` is not an
  available remedy — the override is silently ignored until the user is `trusted`, and a
  distributed user's first session is untrusted by construction.

### M5 — docs-site, four locales

Six pages × four locales (REQ-IFU-020). Korean is the canonical source per the project's
i18n rules; en/ja/zh derive. Same-PR obligation applies — all four locales or none.

### M6 — this repository's own migration (operator-gated)

**[HARD] Proceeds only after explicit operator confirmation in this lane.** It is the
riskiest step: 61,908 bytes of live maintainer doctrine, git-tracked despite being
gitignored, moving to a file under 40,000 characters with operational procedure split into
`.moai/docs/`.

- Split procedure out, leave rules behind (REQ-IFU-021).
- Rewrite §0 to name `AGENTS.local.md` (REQ-IFU-022) — §0 is the discriminator for which
  copy is canonical, and it names the filename directly.
- Verify M1's answer actually holds for this repository's own worktrees before relying on
  it; the lanes read this file.

## §F Self-verification

Per-milestone: the affected packages only (`go test ./internal/<pkg>/...`), never
`go test ./...` locally. The full-suite verdict is CI's, on the PR head, in a clean
environment.

At close: AC-IFU-024 plus a separate re-run of every two-mirror criterion.

## §G Anti-patterns

- Grepping one mirror and reporting the tree clean.
- Reading the presence of `@AGENTS.local.md` in `CLAUDE.md` as evidence the import
  resolved. M0-1 forbids this; AC-IFU-020 is the correct assertion.
- Re-running M1a or M1b until it produces the convenient answer.
- Treating the M0 P7 line-3 observation as settled. It is a synthetic-fixture observation
  and AC-IFU-021 exists because it is not evidence for either answer.
- "Tidying up" the `.tmpl` suffix on the template mirror, or adding a file named `AGENTS.md`
  under `internal/template/templates/`. This is the exact shape of the defect t925 fixed,
  and prose did not stop it the first time — AC-IFU-004 does.
- Reading a completed Codex session as evidence the instructions loaded whole. Tail
  truncation is silent; only the tail sentinel decides.
- Resolving the worktree duplicate-load here. It is t1219's, and this SPEC's obligation is
  only not to make it worse (REQ-IFU-010).
- Letting `moai update` do the migration "since it is already touching the tree".
- Starting M6 without the operator's confirmation because the earlier milestones went well.

## §H Cross-references

- `.moai/reports/t1243/m0/verdict.md` — the measurement this plan's M1 extends.
- design.md §A — the option analysis M1 resolves.
- design.md §D.2 — the `.tmpl` invariant and the t925 measurement behind it.
- design.md §A.3 / card **t1219** item (1) — the worktree duplicate load, scoped out.
- `internal/config/token_budget_guard.go:101` — `CodexContractByteCeiling`.
- `.moai/docs/gitflow-integration-chain.md` — the integration window this lane uses.
