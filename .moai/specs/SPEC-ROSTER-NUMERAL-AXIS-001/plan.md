# SPEC-ROSTER-NUMERAL-AXIS-001 — Implementation Plan

frontmatter-less sibling artifact.
SPEC: `.moai/specs/SPEC-ROSTER-NUMERAL-AXIS-001/spec.md` · card t930 · branch
`WT-numeral-roster-guard` · base `690dfe369`

## §A Context

- Guard package: `internal/harness/rosterguard` (`axis.go`, `check.go`, `registry.go`,
  `rosterguard_test.go`) — card t922.
- Reuse target: `internal/web/docs_tab_contract_test.go`, the `N1 — numeral sweep + allowlist`
  layer (`tabNoun`, `adjacency`, `digitNumeralRe`, `wordNumeralRe`, `ordinalPrefixRe`,
  `allowRule`, `docsTabAllowRules`, `docsTabAllowlist`).
- Roster owner: `template.ProfileMatrixAgents()` — 13 names
  (`internal/template/profile_matrix.go:234-248`, read in this tree).
- Existing hand-registered count-only rows: `product-md-profile-matrix-size`,
  `tech-md-profile-matrix-size` (`registry.go`), both `SweepUnreachable` +
  `KnownStale{DeclaredCount: 11}`, both naming this card as the owner of the general closure.
- Development mode: `.moai/config/sections/quality.yaml` → `development_mode: tdd`,
  `test_coverage_target: 85`.
- Baseline measured in this tree while planning:
  `go test -count=1 ./internal/harness/rosterguard/...` → `ok … 1.096s`.

## §B Known Issues

| # | Issue | Evidence |
|---|-------|----------|
| B1 | `Sweep()` is enumeration-keyed; a count-only claim is structurally unreachable | `registry.go` count-only block comment; `tech.md:17` names zero agents |
| B2 | Two live sites claim 11 where the roster carries 13 | `product.md:151`, `tech.md:17`, `profileMatrixAgentOrder` (13) |
| B3 | Roughly 5 further stale sites use the `11-agent catalog` wording, none registered | `moai-foundation-core/{SKILL.md,modules/INDEX.md,modules/delegation-advanced.md,modules/delegation-implementation.md,modules/token-optimization.md}`, `moai-foundation-quality/SKILL.md`, `model-policy.md` + template mirrors |
| B4 | Legitimate historical citations ("the then-8-agent catalog", "17->8 agent catalog") sit in the same noun class | §E Out of Scope — historical citations |
| B5 | False positives measured in the lead's probe (`172); MoAI retained agents`, `05-25), the agent catalog`, `4 - the retained agent`, rosterguard's own self-describing files) | plan-phase probe output |

## §C Pre-flight (run-phase entry)

- [ ] Re-derive the digit-axis and word-axis populations IN-RUN against the tree being changed,
      and the residual arithmetic with them (hits, `ClaimCount`-discharged, residual); never
      reuse the §A plan-phase figures as the run baseline (REQ-RNA-012, AC-RNA-014).
- [ ] Confirm `go test -count=1 ./internal/harness/rosterguard/...` is green before any edit,
      and land that baseline record as its own commit BEFORE the implementation commit
      (`verification-claim-integrity.md` §2.3 — the commit graph is the only sequencing
      witness).
- [ ] Confirm `Sweep()` / `CheckSite()` / the two existing tests still pass unchanged after
      each milestone (PRESERVE surface).

## §D Constraints

Carried from spec.md §C. Operational additions:

- Verification scope: `./internal/harness/rosterguard/...` plus any package actually touched.
  No full local suite.
- Commit discipline: every commit on this branch names `card t930`; Conventional Commits.
- Plan phase commits nothing outside `.moai/specs/SPEC-ROSTER-NUMERAL-AXIS-001/`.

## §E Milestones

Ordered by decision-reversibility: the shape of the reported set comes first, the mechanical
registry expansion last.

### M1 — Discharge model (the decision most likely to change)

The D2 rule, expressed in the type system before any regexp exists.

- Add the numeral-exempt declaration to `axis.go` beside `SweepUnreachable` and `Staleness` —
  same shape as those two: a declaration with a mandatory non-empty reason, never an inference.
- Implement discharge: `ClaimCount` row for the path, or a numeral-exempt declaration.
  A `ClaimMembership`-only row does NOT discharge (REQ-RNA-004).
- Empty/whitespace reason → a finding reporting the declaration as incomplete (REQ-RNA-005).
- RED first: table-driven cases for each discharge path and each refusal.

### M2 — Noun class, adjacency and numeral selection

The D1 decision, plus the two selection rules that decide what a hit MEANS.

- Layer scope per REQ-RNA-013: inherit `sweepSkipPrefixes` + `sweepSkipFiles` from
  `check.go` and add `.moai/research/`, stated in one place and cited rather than re-listed.
- Noun class per D1 with ASCII word boundaries (`tabNoun` precedent).
- Adjacency window adopted from the reuse target (12 characters) — re-measured in-run rather
  than assumed, and recorded with the measurement in the code comment.
- Nearest-preceding-numeral selection with a regression case pinning
  `CLAUDE.md 4 (13 retained agents)` to 13 (REQ-RNA-003).
- Selector neutralisation for `one of the N …` and `N are retained agents` (REQ-RNA-007), by
  the `ordinalPrefixRe` mechanism.

### M3 — Word axis (structural, population 0)

- `wordNumeralRe` covering `eleven|twelve|thirteen` and the locale forms the reuse target
  already lists, alternatives ordered longest-first (Go regexp is leftmost-FIRST).
- A positive control on synthetic input proves the regexp fires, so a live count of 0 is
  distinguishable from a dead regexp (REQ-RNA-006).
- The code comment states the axis as "correct when it turns on later", not as catching
  something now, and records the standing maintenance cost of the locale class.

### M4 — Output, anti-vacuity and the control probe

- Hits printed one line per hit via `fmt.Println`, anchored at column 0 — `t.Logf` indents and
  prefixes with `file:line`, which breaks an anchored grep (REQ-RNA-008; the reuse target
  states this choice and its reason).
- Zero hits → fail as a measurement failure (REQ-RNA-009), matching `Sweep()`'s own clause.
- Control probe feeding input that MUST violate, observed to fire
  (`TestGuardFiresOnDeliberatelyWrongInput` precedent). A silent control probe is a dead guard:
  run it with `-v` and confirm it executed, then run the package without `-run`.

### M5 — Registry expansion (mechanical)

- Add a proper `ClaimCount` row for each newly-reached stale site: its own `CountPattern`
  (exactly-one-match rule) and a `KnownStale` marker with `DeclaredCount`, reason and follow-up.
- **Consequence, stated plainly and corrected against measurement.** Measured by the
  dispatching lead in this tree at HEAD `6abcc85fa` with its own `registry.go` parse:

  | Quantity | Value |
  |---|---|
  | Hit paths | 63 (62 digit-axis + `workflow-specialist.md`, word-only) |
  | Registry paths carrying a `ClaimCount` row | 20 |
  | Hits already discharged by such a row | 17 |
  | **Residual needing a new row or an exempt declaration** | **46** |

  46, not the "roughly 15" this plan asserted in its first draft. That figure is the
  deliverable's real cost, and it is what decision D3 (mirror derivation, spec.md §D) responds
  to: deriving each `internal/template/templates/` row from its local counterpart via a
  `mirrorOf()` helper — the `readmeSite()` precedent in the same file — cuts what a reviewer
  must read to roughly 26 without changing the noun class or the guard's reach.
- [HARD] "Roughly 26" is a projection, not a measurement: the mirror pairs have not been counted
  exhaustively. Run-phase re-derives the exact post-derivation row count in-run (REQ-RNA-012)
  and records it as the observed cost.
- Historical citations and measured false positives are handled per §E of spec.md: excluded by
  mechanism where the mechanism covers them, otherwise carried as an exempt declaration with a
  reason a reviewer can disagree with.
- Repair of the stale prose itself is NOT done here.

### M6 — Doc comments

- Package doc comment gains a short section naming the second axis and why it exists, in the
  register `check.go`'s head comment already uses, plus the `mirrorOf()` helper's own doc
  comment stating the derivation reason the way `readmeSite()` does.
- Template-First is **conditional here and is expected not to fire**: M5 edits `registry.go`
  only, and prose repair is out of scope (spec.md §E), so no `.claude/` ↔
  `internal/template/templates/.claude/` pair is touched by construction. The clause is
  recorded so that IF a pair is touched it is edited on both sides in the same commit — not as
  work to go find.

## §F Risks

| Risk | Shape | Mitigation |
|---|---|---|
| Allowlist becomes the subject matter | Each exempt row is prose a reviewer must read; too many and the guard reports its own exemptions | The measured residual is 46, not a small number — this risk is REAL at the adopted noun class. D3 (mirror derivation) is the mitigation: it cuts reviewer-read rows to roughly 26 without narrowing reach. An exempt row still needs a reason, which prices it |
| Adjacency window tuned to make the tree green | A window chosen to silence hits is a guard fitted to today's prose | The window is adopted from a measured precedent and re-measured in-run; any change from 12 is recorded with the measurement that motivated it |
| Word axis reads as dead code | Population 0 invites deletion by a later reader | Positive control + a comment stating the axis is a structural hole, not a current catch |
| Historical citations mass-exempted | Blanket exemption of the "then-N" shape would also absorb a future stale claim written in that tense | Prefer a mechanism (tense/selector neutralisation) over an exempt list; where a list is used, it is per path with a reason |
| Double counting against the template mirror | Roughly half the 62 files are `internal/template/templates/` mirrors | Rows are per path, as in the existing registry; the mirror is a real site and gets its own row |

## §G Anti-Patterns

- Counting hits instead of printing them — a guard implementing one word passes every
  shell-level criterion while missing most sites.
- Discharging a count hit on a membership row (the D2 rejected alternative).
- Lowering `SweepThreshold` as a cheaper substitute — it cannot reach a zero-name file.
- Repairing the prose in this card and calling the hole closed; the hole is the missing layer,
  not the two stale numbers.
- Treating a green run with an empty hit set as success.

## §H Cross-References

- `internal/web/docs_tab_contract_test.go` — the reuse target and its stated decisions
- `internal/harness/rosterguard/{axis.go,check.go,registry.go,rosterguard_test.go}` — t922
- `internal/template/profile_matrix.go` — `ProfileMatrixAgents()`, the roster owner
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.1, §2.3 — exemption-marker
  precedent and baseline-first ordering

🗿 MoAI
