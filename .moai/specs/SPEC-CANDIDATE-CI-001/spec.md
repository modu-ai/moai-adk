---
id: SPEC-CANDIDATE-CI-001
title: "Pre-landing candidate CI — merge-tree candidate commits pushed to ci/<card> with a green-gated landing check"
version: "0.1.0"
status: in-progress
created: 2026-10-09
updated: 2026-10-10
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli, internal/factory, internal/config, internal/homestate, .github/workflows"
lifecycle: spec-anchored
tags: "ci, candidate-branch, merge-tree, factory, integration-window, guard-bundle, race-split, flaky-retry"
card: t1478
tier: L
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-09 | manager-spec | Initial authoring (card t1478, Tier L, operator-approved 2026-10-03). Requirements REQ-CCI-001..017 honoring the two REQ anchors the codebase pre-allocated at HEAD db0c514d3: REQ-CCI-004 (card branch/tree resolution, internal/factory/integration_merge_step.go:252,675) and REQ-CCI-011 (shared landing check, internal/factory/integration_merge_step.go:104, internal/cli/integration_merge.go:96-99) |

---

## §A. Background (Context)

This repository's factory lands card work through an integration window: a lane acquires
the window (`moai integration acquire`), the merge step gates and merges (`moai
integration merge --card`), the card reaches `merged-local`, and the leader batch-pushes
the integration branch (`lead_push_threshold`, .moai/config/sections/git-strategy.yaml:35).
Remote CI on the integration branch is the authoritative verdict surface
(.claude/rules/local/gitflow-lane-protocol.md §4). The lane is today forbidden from pushing
its worktree branch and from requesting CI — CI runs only when the leader pushes the
integration branch (gitflow-lane-protocol.md §2 EXCLUDED clause, §4).

The operator's investigation C (card t1478, operator-approved 2026-10-03) measured 24
consecutive non-green runs on the integration-branch CI, of which **zero were pure
regressions** — the failure causes were guard drift (12 kinds) and platform drift. Under
the current model those reds are only discovered AFTER a batch push to the integration
branch, so a drift red invalidates a whole batch instead of one candidate.

Two forward-wiring anchors for this SPEC already exist in the tree at HEAD db0c514d3:

1. `candidateCIEnabled(root)` — a constant-false placeholder that owns the config key name
   `workflow.candidate_ci.enabled` and whose @MX:UPGRADE names this card
   (internal/cli/integration_merge.go:175-189).
2. `MergeStepSeams.LandingCheck` — the merge step's gate 5 seam, documented as
   "SPEC-CANDIDATE-CI-001's REQ-CCI-011 shared check", currently wired by the verb to a
   LOUD refusal placeholder (internal/factory/integration_merge_step.go:104-106,
   internal/cli/integration_merge.go:90-99).

The operator-approved doctrine change embedded in the card: the §4.1 rule "WT push · CI
요청 금지" (lanes never push; lanes never request CI) is REPLACED for one narrow path —
the candidate path this SPEC defines. The local-only rule file
`.claude/rules/local/gitflow-lane-protocol.md` is amended in the sync phase (REQ-CCI-016).

## §B. Problem Statement

A card's work reaches the integration branch without any pre-landing full-suite verdict.
The first CI run that sees the card's bytes is the integration branch's own run, after the
merge — so a red discovered there costs a batch, not a card. The 24-run non-green streak
(investigation C) is dominated by two drift classes that a per-card pre-landing run would
have caught cheaply:

- **Guard drift** — guards keyed to source coordinates or registries red-flag unrelated
  edits (12 kinds measured).
- **Platform drift** — changes that build on the author's OS but not on linux/windows.

Additionally, the current race jobs are a 2-way name-prefix split whose internal/cli legs
run ~834s/~947s against a 35m per-binary timeout with a measured +43% same-class variance
(.github/workflows/ci.yml:251-257, 291-297), and one ~369s internal/cli test (investigation
C figure) dominates its shard.

## §C. User Story

As the factory operator, I want each card to produce a disposable merge-tree candidate
commit pushed to `ci/<card>` BEFORE it can land, so that the full CI suite judges the card
in isolation; the integration branch advances only on a green candidate; drift-prone guard
failures are grouped into one attributable check; and the card is held automatically when
its candidate is red — so that integration-branch CI returns to being a confirmation
surface instead of the first line of defense.

## §D. Requirements (GEARS)

Requirement IDs REQ-CCI-NNN. The `CCI` mnemonic is shared with
SPEC-V3R6-CLI-CONFIG-INTEGRITY-001 (its REQ-CCI-001/002/006/007); all REQ-CCI citations in
the candidate-CI scope refer to THIS SPEC, and cross-SPEC citations stay SPEC-scoped (see
§H). REQ-CCI-004 and REQ-CCI-011 are pinned to the meanings the committed code anchors
already give them.

### REQ-CCI-001 — Candidate verb (Ubiquitous)

The `moai integration` command family shall register a `candidate` subcommand alongside
its existing seven subcommands (status, acquire, release, policy, remeasure, merge,
preflight — internal/cli/integration.go:269), invocable as
`moai integration candidate --card <id>`.

### REQ-CCI-002 — Candidate commit construction (Ubiquitous + event-detected)

The candidate verb shall construct the merge candidate with `git merge-tree --write-tree`
followed by `git commit-tree`, producing a candidate commit whose first parent is the
integration branch tip and whose second parent is the card's pinned SHA — the exact
would-be merge commit — reusing the merge-tree discipline internal/factorylane/merge.go:37-46
already establishes.

When `git merge-tree` reports a conflict, the candidate verb shall refuse with a message
naming the card and the conflict, write no candidate branch, and record no candidate
commit — conflict resolution remains the owning lane's duty
(gitflow-lane-protocol.md §5).

### REQ-CCI-003 — Candidate branch (Ubiquitous)

The candidate verb shall push the candidate commit to the remote branch named
`ci/<card-id>`, where `<card-id>` is the factory card id validated against the project's
card-id shape; a re-candidate for the same card shall replace that branch's tip in place
(candidate branches are disposable per-card artifacts, never protected refs).

### REQ-CCI-004 — Card branch/tree resolution contract (Ubiquitous — pre-allocated anchor)

The candidate verb shall resolve the card's worktree to its `WT-*` branch and pin ONE
commit SHA from that branch before any candidate construction — the same resolution
contract internal/factory/integration_merge_step.go:252 and :675 cite as "REQ-CCI-004's
contract" for `resolveCardBranch` — so the candidate, the merge step, and the record all
name the same pinned SHA. The verb shall refuse a card whose worktree is missing, detached,
or carries no `WT-*` branch.

### REQ-CCI-005 — Candidate record (Ubiquitous)

The candidate verb shall write a candidate record into the primary checkout's state store
(the store shared with the integration window record, internal/cli/integration.go:12-15)
keyed by (card id, pinned SHA) and carrying at minimum: the candidate commit SHA, the
integration branch name and tip it was built against, the candidate branch name, the run
identity once CI starts, the verdict (`pending` | `green` | `red`), and the timestamps of
push and verdict observation.

### REQ-CCI-006 — Capability gate (Where)

Where the config key `workflow.candidate_ci.enabled` is absent or false, the candidate
verb shall refuse with a message naming the key, and the merge step's landing check shall
remain the absent no-op seam it is today (internal/cli/integration_merge.go:186-189,
internal/factory/integration_merge_step.go:303-309).

Where the key is true, the candidate verb shall operate and
`candidateCIEnabled` (internal/cli/integration_merge.go:186) shall read the real key,
replacing the constant-false placeholder its @MX:UPGRADE names.

The key shall be defined in BOTH `.moai/config/sections/workflow.yaml` AND its template
source `internal/template/templates/.moai/config/sections/workflow.yaml` (Template-First;
the Go struct side per internal/config/types.go Workflow struct, yaml key
`candidate_ci`), defaulting to false.

### REQ-CCI-007 — CI trigger (Ubiquitous)

The CI workflow (.github/workflows/ci.yml:16-25) shall extend its `push` trigger with the
`ci/**` branch pattern, so a candidate push runs the workflow; candidate runs shall carry
the same required check set as integration-branch push runs (detect, lint, test, build,
test-integration, plus the guard bundle per REQ-CCI-008), with the paths-filter detect job
still free to fast-skip a docs-only candidate via the shared filter
(.github/workflows/ci.yml:44-62).

### REQ-CCI-008 — Guard bundle job (Ubiquitous)

The CI workflow shall define ONE grouped job — the guard bundle — that runs the three
drift-prone guard families as a single named check:

- source-scan guards (exemplar: the destructive-target registry check that statically
  parses Go source, internal/cli/update_destructive_registry.go:18-21);
- line-key/coordinate guards (exemplars: the AST-parsing guards that record 1-based line
  numbers, internal/cli/codex_launcher_guards_test.go:301,
  internal/cli/vocabulary_guard_test.go:20);
- the census guard registry (exemplar: scripts/ci-census/test-census.sh with its
  fixture check census-check.sh, riding today's race jobs at
  .github/workflows/ci.yml:299 and :358).

Whether the guard bundle gates the candidate verdict shall be governed by the config key
`workflow.candidate_ci.guard_bundle_required`, defaulting to true — the default preserves
today's behavior where these guards gate (they ride the ordinary test suite), and the
operator may flip it only after observing guard-drift rates.

### REQ-CCI-009 — Race split (Ubiquitous)

The CI workflow's race jobs shall partition the suite so that:

- internal/cli is split across FOUR race shards (today: 2, ci.yml:258 and :317);
- internal/harness/rosterguard runs as its own standalone race shard;
- the shard selectors form a partition: every test belongs to exactly one shard, with no
  overlap and no gap (machine-checkable from `go test -list`);
- the single longest internal/cli race test (measured ~369s, investigation C) is repaired
  so no single test dominates its shard — the in-tree corroboration for this family is the
  documented "internal/cli ~379s/70%" measurement (.github/workflows/ci.yml:281);
- every shard's runtime keeps headroom over its measured last-green figure, continuing the
  ceiling discipline the race steps document (ci.yml:291-297).

### REQ-CCI-010 — Verdict observation (Event-driven)

When a candidate run completes on `ci/<card>`, the candidate record for that card shall
carry the run's verdict — `green` when the required check set passed, `red` otherwise —
together with the run identity and observation timestamp; the observation is performed by
an explicit read of the CI run state (the same gh-based read surface the local CI-watch
protocols use), never assumed from the push alone. A run's verdict may be recorded for a
candidate only when the observed run's head SHA equals the record's candidate commit SHA
AND the run's ref equals the record's candidate branch; a completed run matching neither
— a late-arriving run from a superseded candidate on the same `ci/<card>` ref — shall be
discarded and never recorded as the current candidate's verdict. The record's integration
target (the branch name and tip it was built against, REQ-CCI-005) is part of the
verdict's meaning: the landing check (REQ-CCI-011) shall admit a merge only when that
target matches the merge's ACTUAL destination — the record's integration branch equals
the merge step's resolved integration branch, and the candidate commit's first parent
(the recorded tip) equals that branch's current tip. A verdict verified against one
target is evidence about that target's merge only; this SPEC does not transfer it across
targets.

### REQ-CCI-011 — The shared landing check (Ubiquitous — pre-allocated anchor)

The merge step's gate 5 (internal/factory/integration_merge_step.go:303-309) shall wire
its `LandingCheck` seam (declared :104-106 as this SPEC's "REQ-CCI-011 shared check") to
the real implementation, and the implementation shall admit a merge only when a candidate
record exists whose pinned SHA equals the merge step's pinned SHA, whose candidate commit
still descends from that pinned SHA, and whose verdict is `green`. A red, missing, or
stale candidate shall refuse the merge with cause 5 (`MergeExitLandingRefused`,
internal/factory/integration_merge_step.go:43). The gate is UNAVOIDABLE: the step gates
on `seams.LandingCheck != nil` (:305), so BOTH of its call sites — the integration merge
verb (internal/cli/integration_merge.go:100) and the factory complete self-issued merge
path (internal/cli/factory_card.go:1977-1988, whose seams today carry only `ReadCard`) —
shall wire the same shared check, and the step shall refuse (fail closed) when the key
is enabled and no LandingCheck is wired, or when the candidate's recorded integration
target does not match the actual merge destination — a branch mismatch refuses outright
(a candidate verified against another target never admits), and a target tip advanced
past the candidate's first parent voids the verification (re-candidate required, the
same re-measure-and-re-acquire discipline as cause 2,
integration_merge_step.go:283-288) — so no current or future caller merges without the
gate or across a target mismatch.

### REQ-CCI-012 — Red-candidate per-card hold (Event-driven)

When the candidate verdict for a card is observed red, the hold shall be PER-CARD state
carried by the card's own candidate record — never a mutation of the shared integration
window policy. Enforcement:

- The owning card's merge shall refuse (REQ-CCI-011's landing check — per-card by
  construction, keyed to the card and its pinned SHA).
- The owning card's window acquisition shall refuse while its candidate reads red: the
  acquire path shall run a card-aware candidate precondition, in the same position and
  shape as the existing settings-drift precondition that runs before any window-record
  mutation (internal/cli/integration.go:427-436), and its refusal shall name the card,
  the red verdict, and the pinned SHA — writing neither the window record nor the window
  policy.
- The shared `IntegrationWindowPolicy` hold (the card-blind refusal at
  internal/factory/integration_lock.go:416, which rejects every acquisition regardless of
  `want.Card`; measured green by TestAcquireUnderHoldRefusesNamingReason and
  TestAcquireWaitUnderHoldOnEmptyWindowMustEnqueueNotGrant, both PASS) shall remain
  reserved for existing integration failures (the post-merge holds,
  integration_merge_step.go:575-582). A candidate verdict shall never write it: a global
  hold would freeze cards whose candidates are green, colliding with the serialization
  discipline AGENTS.local.md:197 (락은 병합을 직렬화하는 장치이지 수리를 직렬화하는
  장치가 아니다).
- A re-candidate for the card producing a green verdict clears the per-card hold.

Seam note (measured): the card argument already reaches the acquire path
(`want.Card`, internal/cli/integration.go:460), so the precondition needs NO
`AcquireIntegrationWindow` signature extension. Where the run phase instead extends the
factory hold branch itself to discriminate on `want.Card`, that extension is the
sanctioned fallback — the requirement is the per-card observable, not the seam.

### REQ-CCI-013 — Known-flaky single retry (Ubiquitous)

The CI workflow shall maintain a committed registry of known-flaky test names, and When a
registry-listed test fails in a candidate or integration run, the runner shall re-run that
test exactly once and record the retry and its outcome in the job's output summary. A
non-registry failure shall never retry; a registry test that fails twice shall fail the
run. Every registry entry shall cite the evidence that earned it (the
SPEC-CI-FLAKY-STABILIZE series precedent, .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/spec.md
§A: failures admitted only from verbatim CI logs).

### REQ-CCI-014 — Concurrency policy (Ubiquitous)

The CI workflow's concurrency shall never cancel an in-progress run on the integration
branch's ref (today `cancel-in-progress: true` applies to every ref,
.github/workflows/ci.yml:35-37 — the integration branch is the exception), while runs on
a `ci/**` ref shall cancel the previous in-progress run of that same ref (a re-candidate
supersedes its predecessor).

### REQ-CCI-015 — Cross-platform vet pre-verification (Ubiquitous)

Candidate runs shall carry the linux and windows vet legs — the existing build job's
`go vet ./...` cross-compile step (.github/workflows/ci.yml:577-582) shall run on
candidate pushes — so platform drift reds a candidate instead of the integration branch.

### REQ-CCI-016 — Doctrine amendment (Ubiquitous, sync-phase delivery)

The candidate path shall REPLACE the "WT push · CI 요청 금지" prohibition for this one
path: a lane's public surface extends to pushing `ci/<card>` candidate branches via the
candidate verb, while the integration branch push remains the leader's batch act
(AGENTS.local.md §4.1 discipline 2-3; gitflow-lane-protocol.md §4, itself drift-flagged
by AGENTS.local.md:203). The sync phase shall amend the canonical integration-chain
section `AGENTS.local.md` §4.1 (AGENTS.local.md:175) and the drift-flagged develop-era
rule `.claude/rules/local/gitflow-lane-protocol.md` §2/§4 (local-only by its own header,
line 10 — never mirrored to internal/template/templates/), recording the operator
approval of 2026-10-03. The retired CLAUDE.local.md (AGENTS.local.md §0.3, lines 27-29)
is neither read nor amended.

### REQ-CCI-017 — Candidate failure modes (Event-detected)

When the candidate push fails (authentication, remote refusal, network), the candidate
verb shall report the failure naming the stage reached, shall leave any previously
recorded candidate verdict untouched, and shall exit non-zero — a failed push never reads
as a pending-or-green candidate. When the card id fails validation, the verb shall refuse
before any remote call.

## §E. Non-Functional Constraints

- The candidate verb performs NO integration-window mutation: candidate construction is
  window-free (the merge-tree probe is read-only on the integration side), so candidate
  traffic never serializes through the window queue.
- The merge step's window-hold discipline is unchanged: the landing check runs at gate 5,
  before the collision probe and the merge, inside the existing gate order
  (integration_merge_step.go:171-568); no gate reordering.
- The candidate branch namespace `ci/*` is never a branch-protection subject; nothing in
  this SPEC touches branch protection.
- All new Go code follows TRUST 5 (85%+ coverage target on the run phase; the test families
  are named per milestone in plan.md §F).
- No time estimates anywhere in these artifacts (constitution §Time Estimation).

## §F. Acceptance Criteria

The acceptance criteria live in acceptance.md (Tier L), AC-CCI-001-1 .. AC-CCI-016-1,
each carrying the exact command or observable that proves it, and each release-blocking
criterion carrying its RED-now observation on tree db0c514d3 per the two-cell discipline
(.claude/rules/moai/development/verification-completeness.md §2).

## §G. Out of Scope

### Out of Scope — The reserved pushed → ci-green card edge

- The F1 card-state edge `pushed → ci-green` is a RESERVED edge whose admission belongs to
  the F3 controller's CI verdict reader (internal/homestate/card_transition.go:294-295
  refuses it naming F3). This SPEC does not wire that edge; the candidate record and the
  landing check are the verdict surfaces this SPEC owns. Wiring the reserved edge is F3's
  scope (card t1241 lineage).

### Out of Scope — Leader batch-push machinery

- The `lead_push_threshold` count trigger and the batch green-conditional remain
  doctrine-owned (gitflow-lane-protocol.md §4, git-strategy.yaml:23-35). This SPEC adds no
  new push path for the integration branch.

### Out of Scope — Flaky root-cause repair beyond the named test

- Root-causing flaky tests beyond the single ~369s internal/cli test of REQ-CCI-009 stays
  with the SPEC-CI-FLAKY-STABILIZE series. The registry of REQ-CCI-013 records known
  flaky tests; it does not excuse their repair.

### Out of Scope — Template distribution of the candidate CI

- The repo's own ci.yml is dev-only infrastructure (the workflow template directory
  carries only label-sync.yml, internal/template/templates/.github/workflows/). The
  candidate CI workflow is NOT distributed to user projects. The config KEY ships through
  the workflow.yaml template mirror (REQ-CCI-006) and is inert at its false default.

### Out of Scope — Codex lane delivery edges

- The Codex merge-edge refusal (REQ-SD-025, internal/cli/integration_merge.go:27-34) is
  unchanged. The candidate verb is lane-agnostic, but nothing in this SPEC reopens Codex
  delivery paths.

## §H. Cross-References

- SPEC-MERGE-WINDOW-QUEUE-001 — the merge step and its 13-cause gate order this SPEC's
  landing check plugs into (REQ-MWQ-017/018/019).
- SPEC-V3R6-CLI-CONFIG-INTEGRITY-001 — the OTHER owner of the `REQ-CCI` mnemonic
  (REQ-CCI-001/002/006/007). The collision is disclosed in §D; every citation stays
  SPEC-scoped. Grep hazard: `grep -rn "REQ-CCI"` matches both families — always read the
  citing file's SPEC context.
- SPEC-FACTORY-LANE-AUTONOMY-001 — the merge-tree condition triple
  (internal/factorylane/merge.go) whose conflict-probe discipline REQ-CCI-002 reuses.
- SPEC-CI-FLAKY-STABILIZE-001/002/003 — the flaky-evidence precedent REQ-CCI-013 adopts.
- SPEC-UPDATE-MIGRATION-001 (card t1547) — the card immediately preceding this one in the
  factory-automation arc; no artifact dependency.
- .claude/rules/local/gitflow-lane-protocol.md — §2/§4 the candidate path amends
  (REQ-CCI-016, sync phase).
- .claude/rules/local/ci-watch-protocol.md — the existing gh-based CI read surface
  REQ-CCI-010's verdict observation reuses.
