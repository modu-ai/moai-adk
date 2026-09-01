---
id: SPEC-STATUSLINE-PROFILE-RESPECT-001
title: "Progress — statusline opt-out honored end-to-end + subtree profile resolution"
version: "0.2.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "progress"
lifecycle: spec-anchored
tags: "progress, statusline, forge, opt-out, launch-ledger"
tier: M
---

# Progress — SPEC-STATUSLINE-PROFILE-RESPECT-001

## §A. Current Phase

**plan-phase complete; kickoff gate PASSED (2026-08-27)** — decisions D1-D5
folded, artifacts at v0.2.0. Status: draft (the `draft → in-progress`
transition belongs to manager-develop on the first run-phase commit). Run
phase may begin.

## §B. Artifact Status

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/spec.md` | draft v0.2.0 (11 REQs; REQ-009 DEFERRED) |
| plan.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/plan.md` | draft v0.2.0 (M0-M4, M6, M7; M5 DEFERRED; 0 open markers) |
| acceptance.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/acceptance.md` | draft v0.2.0 (11 ACs; AC-009 DEFERRED; seam placement pinned §D) |
| progress.md | this file | §E skeleton emitted |

## §C. Milestone Tracker

> Table reconciled during sync phase (2026-08-27). Every row below read `pending`
> while M1-M4 and M7 had in fact landed — the tracker was never updated after the
> run-phase commits. Statuses and evidence are now taken from `git show --stat
> 3abde7053..1fb1ab78a` on this tree, not from the run-phase narrative.

| Milestone | Scope | Status | Evidence |
|-----------|-------|--------|----------|
| M1 | Profile subtree resolution, READ path (REQ-006/007/008) | **landed** | `62485c918` — `internal/profile/profile.go` (+46) + `subtree_resolve_test.go` (+170) |
| M2 | Segment-gated refresh spawn (REQ-001/003) | **landed** | `5a193fa4c` — `builder.go` (+8/-1), `github.go` (+32), `forge_spawn_gate_test.go` (+274) |
| M3 | Explicit-override spawn early-out (REQ-002) | **landed** | `5a193fa4c` (same commit as M2 — the two gates were authored together) |
| M4 | Display-honesty characterization (REQ-005) | **landed** | `d615bf374` — `forge_pair_states_test.go` (+115) |
| M5 | Write-side ledger normalization (REQ-009) | **DEFERRED** → card **t297** (kickoff D1) | — (no run-phase test by design) |
| M6 | Test seams + verification closure (REQ-010/011) | **satisfied — no separate commit** | AC-010 / AC-011 are grep-review criteria discharged against the test files added by M1/M2/M4; no distinct M6 commit exists, and none was expected |
| M7 | Operational edit: `forge: none` in this repo's statusline.yaml (NOT code) | **landed** | `0a92b22f9` — `.moai/config/sections/statusline.yaml` (+9). Operational effect is deployment-gated: see §E.4 |

## §D. Blockers / Open Decisions

None open — kickoff gate passed 2026-08-27. Resolutions on record:

- D1/D2 — ancestor-walk READ-path only (`ResolveLaunchProfileForProject` miss
  path); write-side normalization (REQ-009/AC-009/M5) DEFERRED to follow-up
  card "launch-ledger write-side subtree normalization".
- D2(policy) — anonymous-session default-off REJECTED; explicit config only.
- New — operational `forge: none` edit accepted → plan.md M7 (§2.3
  update-wipe caveat noted inline).
- D3 — spawn-counter seam pinned: post-freshness, pre-`isSelfInvocable`
  (acceptance.md §D preamble).
- D4 — AC-005 fixture wording corrected (absent/corrupt cache ⇒ Available=false).
- D5 — REQ-004 scoped to the recognized-override branch (none→github).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex check: executed 2026-08-27, output `PASS` for
  `SPEC-STATUSLINE-PROFILE-RESPECT-001` (pattern
  `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`); no existing SPEC directory collision.
- Frontmatter: 12 canonical fields present in spec.md per
  `.claude/rules/moai/development/spec-frontmatter-schema.md` (tier: M added as
  optional field; status: draft — the only transition manager-spec performs).
- Out of Scope: 5 `### Out of Scope — <topic>` H3 sub-headings, each with `-`
  bullets (satisfies `OutOfScopeRule`).
- Requirements notation: GEARS (capability-gate `Where`, state-driven `While`,
  event-driven `When`, unwanted `shall not`, ubiquitous). No `IF/THEN`.
- Open clarification markers: 0 — the single marker (plan.md M5, decision D1)
  was resolved at the 2026-08-27 kickoff gate and replaced by a recorded
  verbatim decision digest; AC-009/REQ-009 carry the DEFERRED outcome for
  traceability.

## §E.2 Run-phase Evidence

Run phase executed 2026-08-27 in worktree `.claude/worktrees/t293`, branch
`WT-statusline-profile`, base `3abde7053`. Commits: `62485c918` (M1),
`5a193fa4c` (M2+M3), `d615bf374` (M4), plus the M7/§E.2 closure commit. All
evidence below is (this run, this tree) against the post-M4 HEAD unless noted.

### AC matrix

| AC | Status | Test | Evidence |
|----|--------|------|----------|
| AC-001 | PASS | `TestBuilder_SegmentGateSuppressesPairAndSpawn` | `go test ./internal/statusline/ ./internal/profile/ -count=1 -v -run '...'` → exit 0, 12 top-level PASS / 0 FAIL (`ac-matrix.txt`) |
| AC-002 | PASS | `TestMaybeRefreshGitHubCounts_ExplicitNoForgeSpawnsNothing` (+`TestMaybeRefreshGitHubCounts_UnsetOverrideStillSpawns`) | same run, exit 0 |
| AC-003 | PASS | `TestMaybeRefreshGitHubCounts_NoConfigSpawnsOnce`, `TestBuilder_NilSegmentsPreservesSpawn` | same run, exit 0 |
| AC-004 | PASS | `TestForgeOptOut_TwoWayRevert` (paired with AC-002 in same run) | same run, exit 0 |
| AC-005 | PASS | `TestRender_ForgePairFourStateContract` (6 subtests) | same run, exit 0 |
| AC-006 | PASS | `TestResolveLaunchProfileForProject_SubtreeWorktreeResolves` | same run, exit 0 |
| AC-007 | PASS | `TestResolveLaunchProfileForProject_DeepestRegisteredAncestorWins` | same run, exit 0 |
| AC-008 | PASS | `TestResolveLaunchProfileForProject_LexicalPrefixIsNotAnAncestor` | same run, exit 0 |
| AC-009 | DEFERRED (kickoff D1) | — | no run-phase test by design; follow-up card "launch-ledger write-side subtree normalization" |
| AC-010 | PASS | grep review | `grep -n 'exec.Command\|gh \|git remote'` over the 3 new test files → only comment-line hits; zero exec targets (`ac-matrix.txt` sibling greps) |
| AC-011 | PASS | grep review | new code adds no env-var reads; pre-existing refs use `config.EnvClaudeConfigDir` / `config.EnvNoProfileFallback`; paths via `filepath` APIs |

### TDD RED evidence (pre-GREEN, `--count=1`)

- Statusline gates: `go test ./internal/statusline/ -run 'ExplicitNoForgeSpawnsNothing|SegmentGateSuppressesPairAndSpawn|TwoWayRevert'` → exit 1;
  verbatim `spawn attempts = 1, want 0 — an explicit no-forge override must
  gate the refresh child` and `gated render spawn attempts = 1, want 0`
  (`red-statusline.txt`).
- Profile walk: `go test ./internal/profile/ -run 'Subtree|Deepest|Lexical|Stale|ExactMatchBeatsSubtree'` → exit 1;
  verbatim `subtree session resolved to "", want "alpha"` (`red-profile.txt`).
  RED for the right stated reason: gates/walk absent, not fixture breakage
  (AC-008 / miss-fallthrough / exact-match guards were green at RED and stayed
  green — the two-cell pair).

### Closure gates

- `go test ./internal/statusline/... ./internal/profile/... -count=1` → exit 0
  (`final-suite.txt`; both `ok`).
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`final-winbuild.txt`).
- `go vet ./internal/statusline/... ./internal/profile/... ./internal/cli/` → exit 0.
- `gofmt -l` over the 6 touched files → empty.
- Coverage (touched packages): statusline 90.6%; profile 84.3% package /
  `lookupSubtreeProjectKey` 88.9% and `ResolveLaunchProfileForProject` 100.0%
  per-function (`coverage.txt`, `profile-cover.out`).
- M7 operational read-back: `.moai/config/sections/statusline.yaml` line 12
  reads `forge: "none"`; `git status internal/template/` clean (no mirror).

Evidence dir: `.moai/state/verify/t293-dev/` — `m0-baseline.txt`,
`red-statusline.txt`, `red-profile.txt`, `green-m1m3.txt`, `green-m1m3-v2.txt`,
`m4-fourstate.txt`, `coverage.txt`, `profile-cover.out`, `final-suite.txt`,
`final-winbuild.txt`, `ac-matrix.txt`.

### Gaps

- Profile package-level coverage 84.3% is below the 85% bar; the shortfall is
  pre-existing uncovered functions (Delete, GetCurrentName paths) outside this
  SPEC's scope. The SPEC's own additions are covered at 88.9% / 100%.
  - **[Correction — sync audit F4, 2026-08-27]** The conclusion holds but one cited
    name is wrong: `GetCurrentName` measures **100.0%**, not uncovered. The sync
    auditor re-ran `go tool cover -func=` on this tree and found 13 sub-85%
    functions, all pre-existing — `IsValidPermissionMode` 0.0%, `IsValidProfileName`
    0.0%, `EnsureDir` 70.0%, `GetBaseDir` 71.4%, `WritePreferences` 72.7%,
    `launchCandidateIsUsable` 75.0%, `Delete` 76.9%, `RecordLastUsedProfileForProject`
    77.5%, `syncStatusline` 78.3%, among others. The SPEC's own additions are
    confirmed at `lookupSubtreeProjectKey` 88.9% / `ResolveLaunchProfileForProject`
    100.0%. Evidence: `.moai/reports/t293/sync-audit.md` §2.7. The run-phase sentence
    above is left intact as lane-2's record; this line supersedes its function list.
- No integration run of the real statusline binary against the repo-local
  `forge: none` key (M7 is an operational edit; its code path is covered by
  unit tests against identical fixtures).


## §E.3 Run-phase Audit-Ready Signal

> **Provenance — read, not performed.** The run phase was executed by a different
> session (lane-2) on 2026-08-27 and its commits landed before this signal was
> written. The sync-phase actor did NOT run the run-phase cycle; it read lane-2's
> committed evidence and the independent sync auditor's re-execution, and records
> below only what one of those two actually observed. Every row names which.

```yaml
run_complete_at: 2026-08-27
run_performed_by: lane-2          # NOT the sync-phase actor
run_commit_shas: [62485c918, 5a193fa4c, d615bf374, 0a92b22f9]   # M1, M2+M3, M4, M7
run_base: 3abde7053
run_status: complete
ac_pass_count: 10                 # AC-001..008, AC-010, AC-011
ac_deferred_count: 1              # AC-009 — kickoff decision D1, follow-up card t297
ac_fail_count: 0
milestones_landed: [M1, M2, M3, M4, M7]
milestones_deferred: [M5]         # -> t297
milestones_without_own_commit: [M6]   # grep-review criteria, discharged by M1/M2/M4 test files
new_warnings_or_lints_introduced: 0
cross_platform_build:
  performed: true
  by: lane-2 (build) + sync auditor (build AND vet)
  result: GOOS=windows GOARCH=amd64 go build ./... -> rc 0; the auditor additionally
    ran GOOS=windows go vet on both packages, which compiles the _test.go files that
    a bare build never touches
coverage:
  statusline: 90.6%
  profile: 84.3%                  # below the 85% bar; attribution verified, see below
  new_functions: lookupSubtreeProjectKey 88.9% / ResolveLaunchProfileForProject 100.0%
evidence_dir: .moai/reports/t293/evidence/   # 11 files, recovered in the sync commit
```

### Independent re-execution by the sync auditor (this is what makes the run signal audit-ready)

lane-2's own AC matrix is a claim. The following were re-measured cold by the sync
auditor against this tree, and are the reason this signal is marked audit-ready
rather than merely reported:

- **Mutant probe, 3/3 killed.** Reverting each gate condition in turn made the
  corresponding tests fail with verbatim output (`spawn attempts = 1, want 0`;
  `spawn attempts = 2/3/4, want 0`; `subtree session resolved to "", want "alpha"`).
  A test that still passed with the fix removed would have proven nothing — none did.
  No mutant residue: `git diff --stat` empty after restoration.
- **Selector honesty.** The cited `-run` regexes matched a non-zero number of tests —
  3 and 4 top-level in `internal/statusline`, 8 in `internal/profile`. A zero-match
  selector prints `ok` and asserts nothing; that failure mode is excluded by count.
- **The gate stops the spawn, not merely the render.** Both gates return before the
  spawn point — one inside `maybeRefreshGitHubCounts` (override), one at its call
  site (segment). This is the criterion that matters for issue #1675: suppressing
  only the render would have left the rate-limit path open while reading green.
- **Coverage attribution.** The 84.3% shortfall is pre-existing and outside this
  SPEC's additions — confirmed by `go tool cover -func=`. One cited function name in
  §E.2 was wrong; corrected in place above (F4).

### §E.3 Gaps — what the run phase did NOT observe

- **No integration run of the real statusline binary.** Neither lane-2 nor the auditor
  executed `moai statusline` and counted actual child processes; the spawn-suppression
  verdict rests on unit-level spawn counters and on reading the call path, not on
  observing a process that did not start.
- **`__no_such_profile__` origin unresolved.** The sentinel is absent from this
  repository; which upper process emits it was not traced. This is the accepted Gap
  recorded in spec.md §F, unchanged.
- **`launch.yaml` exercised through fixtures only.** Real-ledger shapes (case
  variants, `SameFile` aliases, dead worktree paths) were not run against the
  subtree walk.
- **Full suite not run**, per the standing prohibition on local full-suite runs. Only
  `./internal/statusline/...`, `./internal/profile/...` and a `./internal/cli/` vet
  were exercised; the cross-package verdict belongs to CI on origin/develop.

## §E.4 Sync-phase Audit-Ready Signal

> **Why this closes at `implemented` and not `completed`:** the independent audit
> returned **PASS-WITH-DEBT**, and the 3-phase close contract admits `completed`
> only on a clean PASS. `completed` becomes available once the open debt (F3, F5)
> is discharged or explicitly accepted by the operator. The grade follows what the
> audit returned, not how severe the remaining findings are — 0 blocking findings
> does not convert PASS-WITH-DEBT into PASS.

```yaml
sync_complete_at: 2026-08-27
sync_commit_sha: 2114ed981   # this SPEC's sync commit; backfilled by its immediate successor
sync_status: complete
spec_status_transition: in-progress -> implemented
transition_rationale: >-
  NOT taken to `completed`. The independent audit returned PASS-WITH-DEBT, and the
  3-phase close contract admits `completed` only on a clean PASS (the same
  precondition `moai spec close` enforces). Precedent in this repository: the
  CI-flake series SPEC landed PASS-WITH-DEBT and its sync commit carried
  `in-progress -> implemented` only. The `completed` transition is available once
  the debt below is either discharged or explicitly accepted by the operator.
independent_sync_audit:
  verdict: PASS-WITH-DEBT
  weighted_harmonic_mean: 86.7
  tier: M
  threshold: 80
  dimensions: {functionality: 88, security: 92, craft: 78, consistency: 88}
  must_pass_firewall: functionality PASS (88), security PASS (92)
  findings: 8   # HIGH 1, MEDIUM 2, LOW 4, INFO 1 — blocking 0
  report: .moai/reports/t293/sync-audit.md
  performed_by: sync-auditor, cold start, different session from the run phase
issue_1675_closure:
  code_layer: closed        # mutant-verified: both gates return before the spawn point
  operational_layer: not_yet_in_force   # deployment-gated, see below
deferred: {ac: AC-009, milestone: M5, follow_up_card: t297}
docs_impact: none           # measured, see below
changelog_entry: added under [Unreleased] -> Fixed
```

### What "#1675 closed" does and does not mean

**Code layer — closed.** The criterion is not "the segment config value is right"
but "the gh polling child is never spawned". Both opt-out levers return before the
spawn point — the override gate inside `maybeRefreshGitHubCounts`, the segment gate
at its call site — and the sync auditor confirmed this by mutation: reverting each
gate made the corresponding tests fail with verbatim spawn-count output. A fix that
suppressed only the render would have left the 429 path open while reading green;
that shape is excluded by measurement, not by inspection.

**Operational layer — not yet in force.** A worktree session resolves `boardRoot`
from `worktree.original_cwd`, i.e. the **primary checkout**, and the primary's
`.moai/config/sections/statusline.yaml` does not yet carry a `forge` key
(`grep -n 'forge' <primary>/.moai/config/sections/statusline.yaml` → rc 1; 33 lines,
pre-merge original). Only this worktree's copy carries `forge: "none"` at line 12.
The opt-out therefore takes effect when the merge reaches the primary working tree —
a deployment state, not a code defect. lane-2's `verdict.md` claimed the polling had
already stopped; that sentence is corrected in place (F2), with the original wording
preserved so the correction is auditable.

**The issue title mis-attributed the cause.** Not "the fallback profile ignores the
config" but "the opt-out key was never in the file the renderer reads, and would not
have gated the spawn even if it had been". spec.md §A.1 refuted this during plan
phase; the audit re-confirmed it independently.

### Documentation impact — measured, not assumed

No documentation change is required, and the basis is a measurement rather than a
judgement: the change adds no user-facing command, flag, or configuration key. The
`forge` and `segments.github` keys it gates are pre-existing and already documented;
what changed is that they now bind the refresh spawn as well as the render. The one
user-visible surface — the four-state forge-pair display contract — was characterized
(pinned) rather than altered, so no rendered output changed shape. The M7 edit is an
operational change to this repository's own config and is deliberately NOT mirrored
into `internal/template/templates/` (a private workflow setting must not ship to the
16 supported programming languages); `git status internal/template/` was clean at
run phase and the neutrality guard stayed green.

### Debt carried into `implemented` (the reason this is not `completed`)

| ID | Severity | Item | Disposition |
|---|---|---|---|
| F1 | HIGH | Run-phase evidence lived only at gitignored `.moai/state/verify/t293-dev/`, so §E.2's citation did not resolve for anyone reading the merged tree | **Closed by this commit** — 11 files recovered byte-identically (`diff -r` rc 0) into tracked `.moai/reports/t293/evidence/` |
| F2 | MEDIUM | `verdict.md` claimed polling had already stopped pre-merge | **Closed by this commit** — corrected in place, original wording preserved |
| F3 | MEDIUM | The two levers read different roots: the segment lever reads the worktree's `statusline.yaml`, the forge lever reads the primary's. An operator editing the worktree copy gets silence, not an error | **Open** — needs a follow-up card: either align the roots or document the asymmetry in code and SPEC. Not fixable inside this SPEC's stated scope |
| F4 | LOW | §E.2 attributed the coverage shortfall partly to `GetCurrentName`, which measures 100.0% | **Closed by this commit** — superseding line added with the measured 13-function list |
| F5 | LOW | `internal/profile` package coverage 84.3% < 85% bar | **Open** — pre-existing, outside this SPEC (verified). Adding tests to `IsValidProfileName` / `IsValidPermissionMode` (both 0.0%) would clear the bar; separate card |
| F6 | LOW | AC-001/AC-003 wording says "config fixture" while the tests inject the segment map directly | **No action** — the yaml→map path is already covered by `internal/cli/statusline_test.go`; wording only |
| F7 | LOW | An unrelated checkout nested inside a registered project now inherits that project's profile (hence its credential dir) | **No action now** — the intended behaviour's flip side; re-evaluate if vendored sub-repositories appear |
| F8 | INFO | AC-009 / M5 deferral | Not a defect — self-contained read-path delivery, cost recorded, follow-up card **t297** queued |

### §E.4 Gaps — what sync phase did NOT observe

- **CI status on `origin/develop` was not read.** Both the run phase and this audit
  judged the tree, not the pipeline.
- **No integration render.** `moai statusline` was never executed against the
  repo-local key; the operational-layer finding rests on tracing which file each
  lever reads, not on counting child processes at runtime.
- **CodeRabbit review state not read** — the branch merged without a card PR under
  the git-flow transition, so there is no PR review surface to read.
- **No full-suite run**, per the standing local prohibition.

### §E.4 Residual risk

- **The lever asymmetry fails silently (F3).** An operator who writes `forge: none`
  into a worktree's `statusline.yaml` will see no error and no effect. The two files
  currently converge on the same committed content, which hides it; editing one in a
  worktree reproduces it immediately.
- **The M7 key is volatile.** `.moai/config` is a root `moai update` wipes wholesale
  (`CleanMoaiManagedPaths`), and the in-file warning comment is read by humans, not
  by the deleting code. The opt-out can vanish on the next update.
- **Ledger growth continues (deferred REQ-009).** Only the read path was fixed, so
  `projects[]` rows keep accumulating per worktree. Exact-match still outranks the
  subtree walk, so the present risk is low.
- **The `0/0` symptom was routed around, not fixed.** The reported "11 issues / 8 PRs
  shown as 0/0" is a stale cache held honestly; this SPEC pinned that contract rather
  than changing it. Here the pair disappears under `forge: none`, so the symptom is
  not observed — in another repository it reproduces unchanged.
