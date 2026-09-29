# plan.md — SPEC-FACTORY-LANE-AUTONOMY-001 (card t1338)

## §A. Context

Umbrella SPEC completing factory-lane autonomy on top of 4 boundary cards. Baseline: worktree
`.moai/worktrees/t1338`, branch `WT-lane-autonomy-umbrella`, base develop `145c3d98c`. 16 REQ
(spec.md §B), 17 AC (acceptance.md). Development mode: tdd (`quality.yaml`).
Reconnaissance: `research.md` R1-R5 + R-D1 (carried explorer evidence; not re-derived).

## §B. Known Issues

- **B1**: t1240/t1241 SPEC dirs absent from develop's tree → no `depends_on` (design.md D8); M0
  gate enforces the predecessor instead.
- **B2**: t1241's condition triple is card-text/plan normative only until its SPEC lands → treated
  as an interface; re-pinned at M3 pre-flight (design.md D3, research.md §R5).
- **B3**: `BacklogItem` has no priority/exec-axis field (`internal/kanban/backlog_store.go:83`) →
  consumer is metadata-optional by construction (D2, REQ-FLA-007).
- **B4**: messaging failure is quiet today (`cross-session-messaging.md` § Availability
  constraints; no detection API in `internal/sessionmsg/`) → the Go probe of D1 is the fix surface.
- **B5**: `moai todo add "-f lane …"` registration failure (R-D1) — routed elsewhere, tracked here
  only so the lead does not lose it.
- **B6**: an idle notice is NOT completion evidence (`cross-session-messaging.md` § An idle
  notice); the fallback foreman must judge completion by the evidence-read contract only.

## §C. Pre-flight (all verified this run unless noted)

- [x] SPEC-ID regex self-check: `PASS: SPEC-FACTORY-LANE-AUTONOMY-001` (verbatim Bash output).
- [x] ID uniqueness: no `SPEC-FACTORY-LANE-AUTONOMY-001` among 984 SPEC dirs in this tree.
- [x] `SPEC-MANAGER-TODO-001` + `SPEC-FACTORY-RECORD-001` present; `SPEC-FACTORY-SELF-DISPATCH-001`
      + `SPEC-FACTORY-CONTROLLER-001` absent (basis for D8).
- [x] `.moai/reports/t1338/` evidence path exists.
- [ ] Run-phase entry gate M0 (below) — NOT yet satisfied; verified at run-phase entry, not here.

## §D. Constraints

Card constraints verbatim-mapped: queue stays the channel (REQ-FLA-016); operator-gate reduction
scope explicit (§ Operator-Gate Reduction Table below); fallback-transition observability
(REQ-FLA-003); run-entry predecessor t1240 merged (M0); umbrella non-overlap (spec.md §D/§F).
No time estimates anywhere — priority labels and phase ordering only.

## § Operator-Gate Reduction Table (card-required named section)

| Operator involvement | Today | Under this SPEC | Verdict |
|---|---|---|---|
| Integration-window grant (lead names the holder) | Lead grants/announces each window | Lane-direct merge: lane acquires the window itself after verifying the condition triple (REQ-FLA-009); the recorded hold still serializes (REQ-FLA-010) | **SHRINKS** — human coordination → mechanical check |
| Worktree disposal approval | Lead/operator doctrine judgment per disposal | Machine-gated `--auto` disposal after the origin-landing check (REQ-FLA-012/014) | **SHRINKS** — but refusal-before-landing (REQ-FLA-015) and both existing guards (REQ-FLA-013) unchanged |
| Dispatch messaging (lead nudges lane) | Lead dispatches each instruction | Lane self-service `--auto` pickup under declared fallback (REQ-FLA-001/002) | **SHRINKS** — only under a declared, logged fallback |
| Card admission (operator picks / standing source) | Operator-owned | Unchanged — no auto-admission; the queue stays the channel (REQ-FLA-016) | **NEVER shrinks** |
| Implementation Kickoff Approval | Mandatory human gate (plan→run) | Unchanged — applies to run-phase entry of any card a lane self-picks | **NEVER shrinks** |
| Irreversible/destructive confirmations | Operator | Unchanged | **NEVER shrinks** |
| CI verdict | Lead reads CI | Unchanged — CI-green NOT in the disposal machine check (design.md D4) | **NEVER shrinks** |

## §E. Self-Verification (plan-phase)

- [x] All 16 REQ use canonical GEARS modality (Ubiquitous/When/While/Where/Unwanted `shall not`);
      no `IF/THEN` legacy form.
- [x] Out of Scope: `### Out of Scope —` H3 sub-headings with `-` bullets (OutOfScopeRule).
- [x] Frontmatter: 12 canonical fields + `tier: L`, `card`, `related_specs`; no snake_case aliases;
      `phase: "v3.2.0 target"` (not a lifecycle token).
- [x] AC sub-IDs Given-When-Then, binary-testable, ≤25.
- [x] Every quantitative claim carries a research.md citation.
- [x] Zero open clarification markers — all 8 design decisions resolved from card + recon.

## §F. Milestones (priority-ordered by decision reversibility — most-likely-to-change first)

### Milestone Gate M0 — Run-entry predecessor (P1, blocking)

Entry test (mechanical): t1240's develop merge confirmed — SPEC-FACTORY-SELF-DISPATCH-001 present
in the develop tree at run entry, or `git merge-base --is-ancestor d43e50bb3 develop` succeeds on
the integration branch. Rationale for gate form: design.md D8. M1-M5 do not start before M0 holds.

### M1 — Messaging-fallback detection + self-service switch (P1 — highest design volatility)

REQ-FLA-001..005. The Go probe (D1), the no-response bound (REQ-FLA-002), the transition-event log
(REQ-FLA-003), and the resume obligations wired onto F1's
`factory handoff recover-resume`/`abandon-lane` + t1241's stall interface (D6). Highest
change-likelihood: new public surface (probe API shape), new doctrine wiring. TDD: probe
predicate tests first (available / unavailable / no-response), then transition-log contract.

### M2 — Classified pickup consumption (P1 — interface boundary to t1332)

REQ-FLA-006..008. Consumer logic per D2: sequential-exclusive / parallel-concurrent rules;
absent-metadata fallback. Gated: if t1332's SPEC does not exist at M2 entry, land the consumer
behind the Where-fallback default with the declared minimal consumption interface documented —
never the producer schema. Change-likelihood high: the consumed schema is a foreign interface.

### M3 — Lane-direct merge conditions (P2)

REQ-FLA-009..011. Pre-flight: re-pin the condition triple from t1241's landed SPEC text (B2).
Implement the lane-executed check sequence over `moai integration acquire/release`; no new
serialization mechanism. Change-likelihood medium: the triple is already normative text.

### M4 — Post-push disposal machine check + `--auto` (P2)

REQ-FLA-012..015. Extend `worktree done` (D4): fetch + rev-list origin-landing precondition,
refusal messaging, `--auto` allowance; L1 + anchored-session guards byte-unchanged in behavior.
Change-likelihood lower: additive precondition on an existing, guarded verb.

### M5 — Cross-fragment integration + observability hardening (P3 — mechanical, last)

Cross-fragment tests (fallback lane × sequential exclusivity; disposal refusal × unpushed merge),
transition-log query surface polish, docs. Purely mechanical; deferred to the bottom per the
reversibility ordering.

Milestone order M1 → M2 → M3 → M4 → M5 after M0; no time estimates (priority + phase ordering
only).

## §G. Anti-Patterns

- Do NOT build F3 M3 controller machinery, t1332's producer schema, or changes to the `--auto`
  foreman's evidence-read contract — all three are boundary-card surfaces (spec.md §F).
- Do NOT treat a messaging idle notice, or a lane's claim of a passed triple, as completion
  evidence — judge by the recorded checks (B6; verification-claim-integrity).
- Do NOT add a second serialization point for merges; the `moai integration` window is the only
  one (REQ-FLA-011).
- Do NOT dispose a worktree before the origin-landing machine check confirms (REQ-FLA-015) —
  including "the branch looks merged" text judgments.
- Do NOT re-add `depends_on` entries for branch-resident SPECs (D8).
- Do NOT let the fallback switch bypass Implementation Kickoff Approval for a self-picked card's
  run phase (reduction table: NEVER shrinks).

## §H. Cross-References

- spec.md §D (interface boundaries) · acceptance.md (AC matrix) · design.md D1-D8 · research.md
  R1-R5, R-D1 · `.claude/rules/moai/workflow/kanban-dispatch.md` (queue-as-channel, disposal
  prohibition) · `.moai/docs/gitflow-integration-chain.md` (integration window procedure) ·
  `CLAUDE.local.md` §4.1 (GitFlow chain, lead push-batch, CI judgment).
