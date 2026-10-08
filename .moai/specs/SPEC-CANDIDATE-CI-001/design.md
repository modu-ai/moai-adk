# design.md — SPEC-CANDIDATE-CI-001

Design decisions D1-D9. Each names the decision, the alternatives considered, and the
reason the chosen shape wins. Behavior-level only; file/function naming lives in plan.md.

---

## D1 — The candidate commit is a TRUE two-parent merge commit

Decision: `git merge-tree --write-tree <tip> <pinned>` → `git commit-tree <tree> -p <tip>
-p <pinned>`.

Alternatives: (a) push the card branch itself to `ci/<card>` (fast, but CI then tests the
card in ISOLATION, not the merge result — platform/guard drift against the integration
tip goes unseen, defeating the card's evidence that integration CI reds were
0-pure-regression); (b) a real throwaway clone merging on the server side (needs a
worktree and a push of the merge — heavier, and the merge already exists locally as a
dry-run primitive).

Reason: the card's evidence (investigation C) is about the INTEGRATED result. A two-parent
candidate commit is byte-identical in tree to what the real merge would produce, so its
CI verdict transfers to the merge step — which is exactly what the pre-allocated landing
check needs to gate on. The merge-tree discipline already exists in-tree
(internal/factorylane/merge.go:37-46, :272-277).

## D2 — Candidate verdict transfers by pinned-SHA identity, not by trust

Decision: the landing check requires record.pinnedSHA == merge step's pinned SHA, record
verdict green, and the recorded candidate commit still descending from the pinned SHA.

Reason: a card's branch moves; an old green for an ancestor SHA must not admit the new
tip. Ancestry is re-verifiable locally with git inside the merge step (cheap, no network),
so the check does not need to trust the record's freshness claims beyond its verdict.

## D3 — Verdict observation is an explicit read, not a push callback

Decision: a gh-based read (the local CI-watch read surface) fills the record's verdict;
nothing writes back from CI.

Alternatives: (a) CI writes the verdict back (needs a write token in CI — a new secret
surface, rejected); (b) the candidate verb polls synchronously until the run finishes
(blocks the lane for the full CI duration — rejected); (c) a separate scheduled poller
(more moving parts than any consumer needs).

Reason: the consumer that needs the verdict (the merge gate) runs on the lane/leader side,
where gh reads already exist. The record carries `pending` until an explicit observation
runs, and the landing check treats `pending` as not-green — fail-closed by default.

## D4 — The candidate verb is window-free

Decision: candidate construction touches no integration-window state.

Reason: the window serializes MUTATION of the integration branch. The candidate is a
read-only probe from the integration side (merge-tree) plus a per-card disposable branch
push; serializing it through the window would queue candidates behind merges and destroy
the card's cheap pre-landing signal. The landing check remains inside the merge step's
existing gate order — the serialized moment stays the merge.

## D5 — Integration branch naming is resolved, never literal

Decision: every reference in code and workflow reads the integration branch through the
existing resolver (config git-flow `develop_branch` / flow-scoped integration target).

Reason: the repository is mid-cutover (git-strategy.yaml:9 still `workflow: git-flow`
with the comment deferring the flip to SPEC-GITHUB-FLOW-DEFAULT-001 M5, lines 28-34). The
card text says "develop" because it was approved under git-flow; the SPEC's requirement
is "the integration branch", and the resolver is what makes that survive the cutover.

## D6 — Guard bundle defaults to gating (preserves current behavior)

Decision: `guard_bundle_required` defaults true; the bundle is one named check either way.

Reason: today these guards ride the ordinary required suite, so a drift red blocks.
Bundling changes ATTRIBUTION (one check naming guard drift) and must not silently change
GATING. The published default rule: the option that preserves current behavior wins. The
operator may flip the key after observing drift rates — that is a policy decision with
data, not this SPEC's default.

## D7 — Race split is a measured partition, and the repair precedes the split

Decision: four internal/cli shards + rosterguard standalone + remainder, selectors formed
from `go test -list` measurement; the ~369s test is repaired BEFORE final selector
balancing.

Reason: balancing selectors around a dominant single test bakes the dominance in. The
existing split's own history (ci.yml:251-257: measured 826/834.5s vs 4264/947s, chosen by
measurement) is the precedent: measure, then partition, and keep the ceiling-headroom
comment discipline (ci.yml:291-297).

## D8 — Exactly-one recorded retry, registry-gated

Decision: a committed registry of known-flaky test names, each entry carrying its
evidence citation; the runner retries a failed registry test exactly once, records it.

Alternatives: blanket retry-N on any failure (masks regressions — rejected, and the card
says "known flaky"); root-cause-fix-everything-first (the flaky-stabilize series owns
that; the registry is the honest interim that records instead of hides).

Reason: the card's own wording — "알려진 flaky 자동 1회 재시도(기록)" — fixes both the
count (1) and the record (기록). The evidence-citation requirement imports the
SPEC-CI-FLAKY-STABILIZE-001 §A discipline (verbatim CI log or it does not enter).

## D9 — REQ-CCI mnemonic collision: adopt, disclose, scope

Decision: SPEC-CANDIDATE-CI-001 keeps the REQ-CCI numbering the committed anchors already
cite (REQ-CCI-004 resolution, REQ-CCI-011 landing check), disclosed against
SPEC-V3R6-CLI-CONFIG-INTEGRITY-001's REQ-CCI-001/002/006/007.

Alternatives: a fresh prefix (REQ-CDCI-*) — cleaner greps, but it dangles two committed
code anchors that name REQ-CCI-011 and REQ-CCI-004 explicitly (internal/factory/
integration_merge_step.go:104, :252/:675), leaving them pointing at a REQ id no SPEC
owns.

Reason: the anchors are already in the tree, cited by SPEC ID. Honoring them costs one
disclosure paragraph; renaming costs a follow-up edit to committed factory code and a
review round. All citations stay SPEC-scoped, which is the convention anyway.

## D10 — The red-candidate hold is per-card record state, never the shared window policy

Decision: a red candidate holds ONLY its owning card, enforced at two per-card points —
the landing check (already keyed to card + pinned SHA) and a card-aware acquire
precondition in the verb layer, in the settings-drift precondition's position
(internal/cli/integration.go:427-436). The shared `IntegrationWindowPolicy` is never
written by a candidate verdict.

Alternatives: (a) flipping the shared policy to hold on a red candidate (the original
REQ-CCI-012 shape) — REJECTED on measurement: the policy-hold refusal is card-blind
(internal/factory/integration_lock.go:416 rejects every acquisition regardless of
`want.Card`; TestAcquireUnderHoldRefusesNamingReason and
TestAcquireWaitUnderHoldOnEmptyWindowMustEnqueueNotGrant both PASS = global hold is a
blanket refusal), so one card's red candidate would freeze cards holding green
candidates — colliding with AGENTS.local.md:197 (락은 병합을 직렬화하는 장치이지 수리를
직렬화하는 장치가 아니다); (b) extending the factory hold branch to discriminate on
`want.Card` — workable (the field already travels, integration.go:460) but it couples
the window lock to candidate-record reads inside a serialized mutation; kept as the
sanctioned fallback, not the seam.

Reason: the candidate record IS the per-card state — keyed (card, pinned SHA), it
answers "is THIS card held" without touching anything shared. The precondition seam
reuses the one card-aware pre-record-mutation shape the acquire verb already has, so
the window record, the window policy, and the queue are untouched by candidate verdicts.
