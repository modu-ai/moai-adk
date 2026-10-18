# decision-index.md — SPEC-PRECOMMIT-AFFECTED-SCOPE-001

Authored at plan phase per `interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`). One row per decision surfaced during assembly that the operator has not settled. Labels: `DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER`. No row carries a recommendation; a `Default:` line is the published Default rule applied, not a preference. Q1 and Q5 are product-level and were RULED by the leader under operator delegation (board d-20261018T005931Z-f264); Q2 and Q3 are implementation-level and carry the applied default (Q3 endorsed by the leader); Q4 stays EVIDENCE-NEEDED for a run-phase measurement.

### Q1: May the pre-commit gate carry a configured list of packages it leaves out of the executed test set and reports as deferred to CI?

Label: FOUNDER
Class: product-level
Authority anchor: (none — the leader's dispatch states two constraints, no skips and no full `internal/cli` suite, but is not in the admissible authority register; ruling 35e9 is a board record not pinned here)
Why unresolved: The two constraints collide in this repository. `internal/cli` is a direct dependent of `internal/hook/quality` and of most packages through its own imports, it carries 1022 `_test.go` files (spec.md F6), and the gate's 120 s test budget must not be enlarged. A commit that selects it either runs a suite that times out, or leaves it out of the executed set. Leaving it out is a visible, configured deferral to CI (AGENTS.md §4), but it is also a test left unrun by the gate, which the "no test skips" instruction may or may not cover.
Options: (A) ship the mechanism (REQ-PAS-012, AC-PAS-012, M4) with an empty default; this repository's value is set at handoff. (B) ship no mechanism; a commit that selects `internal/cli` fails by timeout until a separate card designs name-based selection.
Default: (B) (rule: the option that preserves current behavior)
Alternate: (A)
Operator verdict: RULED 2026-10-18 leader (tn1p4e, operator delegation) board d-20261018T005931Z-f264: (B) — no configurable deferral mechanism (a deferred_packages list, even empty by default, is a standing skip path and the operator forbids test skips); a commit whose affected set includes internal/cli fails by budget overrun, the 120 s budget stays unchanged, and a pre-commit -run selection for internal/cli is a follow-up card recorded by the leader, outside this SPEC. REQ-PAS-012, AC-PAS-012 and plan M4 are withdrawn.

### Q2: When a commit changes go.mod, go.sum, go.work, go.work.sum or the build-tags file, what test scope does the gate use?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none)
Why unresolved: The dependency graph itself changed, so per-package attribution is not available from the staged file list alone.
Options: full-module scope (today's `./...`) with the trigger named in the summary; or widening to the packages that import a dependency named in the changed `go.mod`/`go.sum` lines.
Default: full-module scope (rule: the option that preserves current behavior)
Alternate: widen to importers of the changed dependencies
Operator verdict: DEFAULT-APPLIED 2026-10-18T00:55Z manager-spec (plan-phase author, claude-sonnet-5-5)

### Q3: How is the suppression-policy defect on the Go fixture removed — by narrowing the files the check examines, or by redefining marker/REASON ordering?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — SPEC-UTIL-002's adjacency rule is cited in code comments, but its SPEC body is not pinned here as an anchor)
Why unresolved: Reordering the fixture is ruled out by observation (a marker-first variant makes `sg scan` report `go-interface-empty-not-any`, so `TestRuleSeed` `fixtures-scan` would fail). The remaining options differ in how much policy they touch.
Options: skip `testdata` path segments in the check's file walk; or let the checker accept `@MX:REASON` on either side of the marker.
Default: skip `testdata` path segments (rule: the option with the smaller user-visible surface)
Alternate: accept REASON on either side of the marker
Operator verdict: DEFAULT-APPLIED 2026-10-18T00:55Z manager-spec (plan-phase author, claude-sonnet-5-5)

### Q4: Do the heavy gate's project-wide vet, lint and typecheck steps finish within their unchanged budgets (30 / 60 / 300 s) on this repository in the pre-commit context?

Label: EVIDENCE-NEEDED
Why unresolved: No measurement was taken in this plan phase (a project-wide vet over 176 packages on a shared machine is itself the load AGENTS.md §4 warns about). The result decides whether re-enabling hooks can work at all without enlarging a budget, which the dispatch forbids. It gates the handoff runbook (plan.md § E, pre-condition 3), not any code in this SPEC.
Operator verdict:

### Q5: When, and by which mechanism, are hooks switched back on in the shared repository?

Label: FOUNDER
Class: product-level
Authority anchor: (none)
Why unresolved: The switch is a write to the shared repository configuration that affects every session (spec.md F5). The dispatch makes it a post-landing leader/operator decision; this SPEC only supplies the runbook. The row exists so the decision is recorded as owed, not so that it blocks implementation.
Options: unset `core.hooksPath` in the shared configuration; enable per-worktree configuration; leave hooks off and keep the per-invocation `git -c core.hooksPath=...` override for rehearsal only.
Default: leave hooks off (rule: the option that preserves current behavior)
Alternate: unset `core.hooksPath` in the shared configuration
Operator verdict: RULED 2026-10-18 leader (tn1p4e, operator delegation) board d-20261018T005931Z-f264: hooks stay OFF in this card; plan.md § E keeps the runbook with the three options (unset shared core.hooksPath / per-worktree config / keep off) and the pre-push 5 s sleep revival note; the actual re-enable is a post-landing operator decision.
