# acceptance.md — SPEC-CANDIDATE-CI-001

Acceptance criteria are machine-verifiable: each carries the exact command or observable
that proves it. Release-blocking criteria carry a RED-now observation on tree
db0c514d3 (measured 2026-10-09) per the two-cell discipline
(.claude/rules/moai/development/verification-completeness.md §2): the RED cell states why
it is red, and the green path names the milestone that flips it.

## Requirement traceability

| AC | maps REQ-… | AC | maps REQ-… |
|----|----------------|----|----------------|
| AC-CCI-001-1 | maps REQ-CCI-001 | AC-CCI-009-2 | maps REQ-CCI-009 |
| AC-CCI-001-2 | maps REQ-CCI-001 | AC-CCI-009-3 | maps REQ-CCI-009 |
| AC-CCI-002-1 | maps REQ-CCI-002 | AC-CCI-010-1 | maps REQ-CCI-010 |
| AC-CCI-002-2 | maps REQ-CCI-002, REQ-CCI-017 | AC-CCI-011-1 | maps REQ-CCI-011 |
| AC-CCI-003-1 | maps REQ-CCI-003 | AC-CCI-011-2 | maps REQ-CCI-011 |
| AC-CCI-004-1 | maps REQ-CCI-004 | AC-CCI-012-1 | maps REQ-CCI-012 |
| AC-CCI-005-1 | maps REQ-CCI-005 | AC-CCI-013-1 | maps REQ-CCI-013 |
| AC-CCI-006-1 | maps REQ-CCI-006 | AC-CCI-014-1 | maps REQ-CCI-014 |
| AC-CCI-006-2 | maps REQ-CCI-006 | AC-CCI-015-1 | maps REQ-CCI-015 |
| AC-CCI-006-3 | maps REQ-CCI-006 | AC-CCI-016-1 | maps REQ-CCI-016 |
| AC-CCI-007-1 | maps REQ-CCI-007 | AC-CCI-007-2 | maps REQ-CCI-007 |
| AC-CCI-008-1 | maps REQ-CCI-008 | AC-CCI-008-2 | maps REQ-CCI-008 |
| AC-CCI-009-1 | maps REQ-CCI-009 | (REQ-CCI-017 rides AC-CCI-002-2's refusal family) | maps REQ-CCI-017 |

---

## AC-CCI-001-1 — Candidate verb registered (release-blocking)

- Given: the repository built from a tree containing M2.
- When: `go run ./cmd/moai integration candidate --help` runs.
- Then: exit 0 AND the output contains a candidate-specific usage line
  (e.g. `candidate --card <id>`), distinct from the parent command's help.
- RED-now: on db0c514d3 the same invocation exits 0 but prints the PARENT command's help
  ("Hold and release the release-integration window.") — the only "candidate" match is
  remeasure's Short text; no candidate usage line exists. Red because the verb is absent,
  not because the command fails. Flips at M2.

## AC-CCI-001-2 — Parent subcommand list intact (regression guard)

- When: `go run ./cmd/moai integration --help` runs.
- Then: the subcommand list contains status, acquire, release, policy, remeasure, merge,
  preflight AND candidate (eight entries; the list at
  internal/cli/integration.go:269 gains exactly one member).

## AC-CCI-002-1 — Candidate commit two-parent shape (release-blocking; commit-graph witnessed)

- Given: a temporary origin repo (the internal/factory temp-origin test pattern) with an
  integration branch at tip T and a card branch pinned at SHA P, where P descends from the
  base of T.
- When: the candidate verb runs for the card.
- Then: `git cat-file -p <candidate-sha>` in the temp clone shows exactly two `parent`
  lines, and the set of parents equals {T, P}. The commit graph is the witness — no
  record or log line substitutes.
- Ordering clause: the candidate commit does not exist before the verb runs; the record's
  candidate SHA and the pushed branch tip MUST be the same object
  (`git rev-parse refs/heads/ci/<card>` == record candidate SHA).

## AC-CCI-002-2 — Construction refusals: conflict, push failure, invalid id (release-blocking; maps REQ-CCI-002 and REQ-CCI-017)

- Given: card branch and integration branch touching the same lines.
- When: the candidate verb runs.
- Then: exit non-zero, the message names the card and the conflict, `git ls-remote origin
  refs/heads/ci/<card>` shows no candidate branch, and no candidate record row exists for
  the card.
- And (REQ-CCI-017 family): a push failure (test-double remote refusal) exits non-zero,
  names the stage reached, and leaves any previously recorded verdict untouched; an
  invalid card id refuses BEFORE any remote call observable in the test double.

## AC-CCI-003-1 — Candidate branch replace-in-place (release-blocking)

- Given: a card with an existing candidate push at `ci/<card>`.
- When: the card gains new commits and the candidate verb runs again.
- Then: `git ls-remote origin refs/heads/ci/<card>` reports the NEW candidate SHA (the
  branch tip moved in place), and the previous candidate's record verdict remains
  readable but superseded by the new record keyed to the new pinned SHA.

## AC-CCI-004-1 — Card resolution contract (release-blocking)

- Given: the merge step's resolver as the contract holder
  (internal/factory/integration_merge_step.go:677 `resolveCardBranch`).
- When: `go test ./internal/factory/ -run '^(TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepPreMergeCausesReleaseWithDistinctCodes)$' -count=1` runs after M2
  (the EXISTING contract tests carrying the resolver and gate-order contract today,
  internal/factory/integration_merge_step_test.go:210 and :320) plus the M2 family
  `go test ./internal/cli/ -run '^TestIntegrationCandidate$' -count=1`.
- Then: the candidate verb's resolution shares the resolver (or its extracted form):
  card worktree → `WT-*` branch → one pinned SHA; detached HEAD, missing worktree, and
  non-WT branch each refuse with a message naming the card. RED-now: the candidate verb
  has no resolver call site (the verb itself is absent, AC-CCI-001-1's RED); the resolver
  contract tests exist and pass today — the NEW assertions are the candidate verb's use
  of them, observed in the M2 test family.

## AC-CCI-005-1 — Candidate record shape (release-blocking)

- When: a candidate is pushed for a card.
- Then: the primary checkout's state store contains a record whose key is (card id,
  pinned SHA) and whose fields include candidate SHA, integration branch, integration
  tip, candidate branch, verdict (initially `pending`), and push timestamp. Proven by the
  store's read-back test in `go test ./internal/factory/ -run '^TestCandidateRecord$' -count=1` (the run phase
creates `TestCandidateRecord`).

## AC-CCI-006-1 — Gate default false (release-blocking)

- When: `grep -A2 "candidate_ci:" .moai/config/sections/workflow.yaml
  internal/template/templates/.moai/config/sections/workflow.yaml` runs.
- Then: both files define `candidate_ci:` with `enabled: false` and
  `guard_bundle_required: true`.
- RED-now: `grep -c "candidate_ci"` returns 0 rows across both files on db0c514d3.
  Flips at M1.

## AC-CCI-006-2 — Config symmetry audit stays green

- When: `go test ./internal/config/ -run '^(TestAuditLoaderCompleteness|TestWorkflowConfigFields|TestNewDefaultWorkflowConfig|TestNewDefaultWorkflowConfigNestedDefaults)$' -count=1` runs
  (EXISTING exact test names, measured via `go test -list`).
- Then: exit 0 (the struct side of `candidate_ci` matches the YAML on both mirrors).

## AC-CCI-006-3 — Disabled gate refuses the verb (release-blocking)

- Given: default config (key false).
- When: `moai integration candidate --card t0000` runs against any remote.
- Then: exit non-zero, the message names `workflow.candidate_ci.enabled`, and no remote
  call is made (no network observable in the test double).

## AC-CCI-007-1 — ci/** trigger present (release-blocking)

- When: `git grep -n "ci/\*\*" .github/workflows/ci.yml` runs after M3.
- Then: at least one match inside the `on: push:` block
  (ci.yml:16-18 today carries only `[main]`; RED-now: zero `ci/**` rows — flips at M3),
  and the workflow parses clean under the repo's validator.

## AC-CCI-007-2 — Docs-only candidate fast-skip preserved

- Given: a candidate push whose diff matches no `go_code` filter input.
- When: the candidate run executes.
- Then: the detect job reports go_code false and the skip-marker emits the required
  check names (the parity contract, ci.yml:379-384), so the candidate verdict is green
  without running the Go suite.

## AC-CCI-008-1 — Guard bundle job exists with three families (release-blocking)

- When: `git grep -n "guard-bundle" .github/workflows/ci.yml` runs after M5.
- Then: one job id groups steps invoking the three families (static-parse source-scan
  guards, the line-key guard tests, and scripts/ci-census/test-census.sh +
  census-check.sh), and `bash scripts/ci-census/census-check.sh` exits 0.

## AC-CCI-008-2 — Guard bundle gating default preserves today's behavior

- When: the candidate verdict is computed with `guard_bundle_required: true`.
- Then: a red guard bundle reds the candidate verdict (same gating as today, where these
  guards ride the ordinary suite); with the key false, a red bundle is visible and
  recorded but does not red the verdict. Proven by the verdict-collection test in M5's
  test family.

## AC-CCI-009-1 — Race shard partition exactness (release-blocking)

- When: the partition check runs after M6 (the milestone lands the exact one-command form;
  shape: `go test -list` over ./... diffed against the union and intersection of the four
  internal/cli shard selectors, the rosterguard selector, and the remainder leg).
- Then: union == full test set, pairwise intersections empty. RED-now: the current 2-way
  split (`-run '^(TestFactory|TestInstall|TestTodo|TestRun)'` and its negation,
  ci.yml:298/:357) IS a partition — the NEW properties are the 4-way internal/cli split
  and the standalone rosterguard shard, absent today. Flips at M6.

## AC-CCI-009-2 — Standalone rosterguard shard

- When: the race job list is read from ci.yml after M6.
- Then: a job exists whose selector covers internal/harness/rosterguard and no other
  package's tests (rosterguard package at internal/harness/rosterguard/, 5 test files at
  HEAD).

## AC-CCI-009-3 — Longest-test repair (release-blocking; ordering-baseline)

- Given: the M6 re-measure identifying the dominant internal/cli race test (card figure
  ~369s; in-tree corroboration "internal/cli ~379s/70%" ci.yml:281).
- When: `go test -json -race -count=1 <the identified test's package selector>` runs
  after the repair.
- Then: the identified test's duration is below the threshold M6 records from its own
  pre-repair measurement. Ordering clause: the pre-repair timing artifact (the RED cell)
  lands in its own commit BEFORE the repair commit — the commit graph is the only
  sequencing witness (verification-claim-integrity.md §2.3).

## AC-CCI-010-1 — Verdict recorded from a real read (release-blocking)

- Given: a completed candidate run.
- When: the verdict observation path runs (M4's gh-based read).
- Then: the record's verdict is `green` or `red` (never left `pending` by a completed
  run), and the record carries the run identity + observation timestamp. Proven in the
  test family with a scripted run-state double; no live-CI dependency in unit tests.

## AC-CCI-011-1 — Landing gate admits green only (release-blocking)

- Given: a merge step at gate 5 with pinned SHA P.
- When: the candidate record for (card, P) is `green`.
- Then: `RunMergeStep` proceeds past gate 5 to the collision probe.
- RED-now: today the seam is wired to a LOUD refusal placeholder
  (internal/cli/integration_merge.go:95-99) — with the key absent the seam is nil
  (no-op); the NEW behavior only exists after M4. Flips at M4.

## AC-CCI-011-2 — Red/missing/stale refuse with cause 5 (release-blocking)

- When: the record verdict is `red`, or no record exists for (card, P), or the record's
  candidate commit no longer descends from P.
- Then: `MergeExitCode(err)` == 5 (`MergeExitLandingRefused`,
  internal/factory/integration_merge_step.go:43) and the window releases (pre-merge
  cause class).

## AC-CCI-012-1 — Red candidate auto-holds (release-blocking)

- When: a candidate verdict is observed red.
- Then: the integration window policy record reads hold, its reason names the card and
  the red verdict, and a subsequent `integration merge --card <id>` is refused by the
  landing gate (AC-CCI-011-2) until a re-candidate for the same pinned SHA reads green.
  Proven by the policy-writer test in M4's family; the policy-record shape is
  `CompletePostMergeHold`'s (integration_merge_step.go:575-582).

## AC-CCI-013-1 — Single recorded retry (release-blocking)

- Given: the flaky registry contains a test T (with its evidence citation).
- When: T fails in a candidate run and the retry wrapper re-runs it.
- Then: exactly one retry executes; the job summary records "retried: T, attempt 2,
  outcome pass|fail"; a second failure fails the run; a NON-registry failure performs
  zero retries. Proven by the wrapper's fixture-stream test (test-census.sh precedent)
  plus one observed CI run carrying a registry retry.

## AC-CCI-014-1 — Concurrency policy (release-blocking)

- When: the concurrency block is read from ci.yml after M3.
- Then: the integration branch ref (resolved per config — `develop` today,
  git-strategy.yaml:16) evaluates cancel-in-progress false; `ci/**` refs evaluate true
  (supersede). Proven by the YAML + one observed pair: a re-candidate cancels the prior
  `ci/<card>` run, and an integration-branch push never cancels a running integration
  run. RED-now: ci.yml:35-37 cancels every ref. Flips at M3.

## AC-CCI-015-1 — Candidate vet legs (release-blocking)

- When: a candidate run executes after M5.
- Then: the run's job graph includes the build job's vet step for linux and windows
  (`go vet ./...` under GOOS matrix, ci.yml:577-582), observed in the run's job list.

## AC-CCI-016-1 — Doctrine amendment landed (sync-phase gate)

- When: the sync commit is read
  (`git show <sync-sha> -- AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md`).
- Then: `AGENTS.local.md` §4.1 (the canonical integration-chain section,
  AGENTS.local.md:175) carries the candidate-path exception naming
  `moai integration candidate` and the 2026-10-03 operator approval;
  gitflow-lane-protocol.md §2/§4 gains the same exception with its develop-era drift
  notice retained (AGENTS.local.md:203 already flags that file's develop text as drift);
  the integration-branch leader-batch rule stands in both. The retired CLAUDE.local.md
  is untouched (AGENTS.local.md §0.3 — a discarded model, never a citation target). The
  commit-graph order is the witness: the amendment commit is the sync phase's, after the
  run-phase landing commits.

---

## Edge cases

- Card id shapes: a card id that is not a valid branch-name fragment refuses before any
  remote call (REQ-CCI-017).
- Integration tip moved between candidate push and merge attempt: the record's stored tip
  no longer equals the current tip → stale → refuse (re-candidate required); the merge
  step's own cause-2 (base moved) remains the second net.
- Two candidates for the same card at different pinned SHAs: the record lookup keys on
  the pinned SHA; only the record matching the merge step's pinned SHA is consulted.
- Candidate run for a docs-only diff: detect fast-skips (AC-CCI-007-2) — the verdict is
  green with the skip visible in the summary.
- Remote without push permission for ci/**: REQ-CCI-017's failure reporting; nothing
  claims pending.

## Quality gates (run phase)

- TRUST 5: gofmt/golangci-lint clean; 85%+ coverage on internal/factory candidate-record
  and landing-check code; internal/cli verb tests cover gate, conflict, push-failure,
  and validation paths.
- The 13-cause merge gate order is untouched: the full existing merge-step test family
  passes unmodified except where M4 replaces the placeholder wiring.
- Every workflow edit parses clean and keeps required-check NAME parity
  (test-skip-marker contract).

## Definition of Done

1. All release-blocking ACs above PASS with verbatim command outputs recorded in §E.2.
2. The ordering-sensitive baselines (AC-CCI-009-3's pre-repair timing) sit in commits
   that precede the changes they measure.
3. sync-phase M7 (doctrine amendment) landed per AC-CCI-016-1.
4. `go test ./internal/cli/ ./internal/factory/ ./internal/config/ -count=1` green; CI on
   the integration branch green with the new jobs present.
