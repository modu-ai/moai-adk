# decision-index.md — SPEC-CI-VERDICT-INTEGRITY-001

Stateless on the status axis — the SPEC lifecycle lives in `spec.md` alone.
Authored at plan phase (card t1534). `interview.decision_gate: on` (`.moai/config/sections/interview.yaml`).
Amended at plan-phase amendment 2 (iteration-2): the original DEFAULT-APPLIED
mirror answer for Q1 was withdrawn — wrong-by-construction against the trigger
facts (ledger E21).

### Q1: Which context list should the SSoT's `release/*` branch key carry after the correction?

Label: FOUNDER

Class: implementation-level

Why it surfaced: the card directs the SSoT correction ("required-checks.yml 실제
발행 이름으로 정정") without saying what the `release/*` key carries. The original
draft answered "mirror the corrected `main` list" via the published-Default rule
(the file historically mirrored main). That answer is **withdrawn as
wrong-by-construction**: every PR trigger in the three producing workflows
restricts to `branches: [main]` (ci.yml:20, codeql.yml:7,
release-pr-multi-os.yml:14 — ledger E21), so under the mirror a protection apply
to a release branch would require checks that release-targeting PRs never
publish, which the M3 missing-required-check handling keeps pending forever —
blocked merges by construction.

Resolution (recorded reason): `release/*` carries only what release-targeting
PRs actually publish. Under the current triggers that set is **empty** — no
corrected context (nor any other Lint/Test/Build/Gate context) publishes on a
release-targeting PR, so the truthful list is an empty `contexts:` list. The
`main` key carries the full corrected list and is this card's only protection
apply target (the repo's protected branch is `main`; release/* branches are not
protection targets in this flow). An empty release/* list is honest for the
watch-loop classifier too: a release-base PR today publishes no required
checks, and saying so is the truth — the multi-OS coverage gap for release-base
PRs belongs to the release-pipeline scope, not to phantom contexts.

Boundary (explicitly out): expanding the PR triggers to `release/*` (or to any
release-base branch) is the release-pipeline trigger/coverage scope owned by
card t1536. This SPEC records the boundary and does not expand into it; if
t1536 lands trigger expansion, the release/* list is revisited then, against
what the expanded triggers actually publish.

Authority anchor: none reachable in the committed register at authoring time —
routed FOUNDER (never relabeled). The resolution above is the operator-side
amendment-2 directive (coordinator, card t1534) applied to the measured
trigger facts (E21); the committed artifacts carrying it are this file,
plan.md §F (M2), and acceptance.md AC-CI-008.

Operator verdict: OPERATOR-SETTLED 2026-10-06 coordinator amendment-2 directive
(card t1534): release/* = published-only list (empty under current triggers);
main-only apply; trigger expansion out via t1536.
