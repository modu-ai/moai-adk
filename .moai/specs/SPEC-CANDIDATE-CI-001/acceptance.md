# acceptance.md — SPEC-CANDIDATE-CI-001

Acceptance criteria are machine-verifiable: each carries the exact command or observable
that proves it. ALL 19 release-blocking criteria carry a four-element RED-now observation
(command, verbatim stdout, exit code, tree SHA) per the two-cell discipline
(.claude/rules/moai/development/verification-completeness.md §2). Five cells were
measured at authoring on tree db0c514d3 (AC-CCI-001-1, 006-1, 007-1, 011-1, 014-1); the
remaining fourteen were measured 2026-10-09 on tree c47aeda2d and live in the
§ RED-now evidence ledger at the end of this file, cited by id — the Go tree at
c47aeda2d is code-identical to db0c514d3 (commits ec8180a38, 09380a9aa, c47aeda2d touch
only .moai/specs/ artifacts). Each RED cell states why it is red, and the green path
names the milestone that flips it.

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
- RED-now (ledger EV-CCI-A, tree c47aeda2d): command
  `go run ./cmd/moai integration candidate --card t1478` → stdout
  `ERROR` / `Unknown flag: --card.` / `Try --help for usage.` / `exit status 1` (error-box
  padding elided), exit code 1 — the verb does not exist, so no candidate commit can be
  constructed and no two-parent object exists to inspect. Flips at M2.

## AC-CCI-002-2 — Construction refusals: conflict, push failure, invalid id (release-blocking; maps REQ-CCI-002 and REQ-CCI-017)

- Given: card branch and integration branch touching the same lines.
- When: the candidate verb runs.
- Then: exit non-zero, the message names the card and the conflict, `git ls-remote origin
  refs/heads/ci/<card>` shows no candidate branch, and no candidate record row exists for
  the card.
- RED-now (ledger EV-CCI-A, tree c47aeda2d): command
  `go run ./cmd/moai integration candidate --card t1478` → stdout
  `ERROR` / `Unknown flag: --card.` / `Try --help for usage.` / `exit status 1` (box
  padding elided), exit code 1 — with no verb there is no conflict path whose refusal
  could be observed. Flips at M2.
- And (REQ-CCI-017 family): a push failure (test-double remote refusal) exits non-zero,
  names the stage reached, and leaves any previously recorded verdict untouched; an
  invalid card id refuses BEFORE any remote call observable in the test double.

## AC-CCI-003-1 — Candidate branch replace-in-place (release-blocking)

- Given: a card with an existing candidate push at `ci/<card>`.
- When: the card gains new commits and the candidate verb runs again.
- Then: `git ls-remote origin refs/heads/ci/<card>` reports the NEW candidate SHA (the
  branch tip moved in place), and the previous candidate's record verdict remains
  readable but superseded by the new record keyed to the new pinned SHA.
- RED-now (ledger EV-CCI-A, tree c47aeda2d): command
  `go run ./cmd/moai integration candidate --card t1478` → stdout
  `ERROR` / `Unknown flag: --card.` / `Try --help for usage.` / `exit status 1` (box
  padding elided), exit code 1 — no verb, so no `ci/<card>` branch exists to replace.
  Flips at M2.

## AC-CCI-004-1 — Card resolution contract (release-blocking)

- Given: the merge step's resolver as the contract holder
  (internal/factory/integration_merge_step.go:677 `resolveCardBranch`).
- When: `go test ./internal/factory/ -run '^(TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepPreMergeCausesReleaseWithDistinctCodes)$' -count=1` runs after M2
  (the EXISTING contract tests carrying the resolver and gate-order contract today,
  internal/factory/integration_merge_step_test.go:210 and :320) plus the M2 family
  `go test ./internal/cli/ -run '^TestIntegrationCandidate$' -count=1`.
- Then: the candidate verb's resolution shares the resolver (or its extracted form):
  card worktree → `WT-*` branch → one pinned SHA; detached HEAD, missing worktree, and
  non-WT branch each refuse with a message naming the card.
- RED-now (ledger EV-CCI-B, tree c47aeda2d): command
  `go test -list '^TestIntegrationCandidate$' ./internal/cli/` → stdout
  `ok  github.com/modu-ai/moai-adk/internal/cli	1.403s` (zero test names listed), exit
  code 0 — the M2 test family that would observe the verb's resolution does not exist,
  so the NEW property (the verb shares the resolver) has no starting observation beyond
  this absence; the existing resolver contract tests (TestMergeStep* family) pass today
  and pin the contract the verb will share. Flips at M2.

## AC-CCI-005-1 — Candidate record shape (release-blocking)

- When: a candidate is pushed for a card.
- Then: the primary checkout's state store contains a record whose key is (card id,
  pinned SHA) and whose fields include candidate SHA, integration branch, integration
  tip, candidate branch, verdict (initially `pending`), and push timestamp. Proven by the
  store's read-back test in `go test ./internal/factory/ -run '^TestCandidateRecord$' -count=1` (the run phase
creates `TestCandidateRecord`).
- RED-now (ledger EV-CCI-C, tree c47aeda2d): command
  `go test -list '^TestCandidateRecord$' ./internal/factory/` → stdout
  `ok  github.com/modu-ai/moai-adk/internal/factory	0.442s` (zero test names listed),
  exit code 0 — the record type, its store, and its read-back test do not exist. Flips
  at M1.

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
- RED-now (ledger EV-CCI-D, tree c47aeda2d): command
  `grep -c candidate_ci .moai/config/sections/workflow.yaml` → stdout `0`, exit code 1 —
  the key is absent from the live config (and its template mirror), so no gated refusal
  behavior exists to observe. Flips at M1.

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
- RED-now (ledger EV-CCI-E, tree c47aeda2d): command
  `grep -c guard-bundle .github/workflows/ci.yml` → stdout `0`, exit code 1 — no
  guard-bundle job exists in the workflow. Flips at M5.

## AC-CCI-008-2 — Guard bundle gating default preserves today's behavior

- When: the candidate verdict is computed with `guard_bundle_required: true`.
- Then: a red guard bundle reds the candidate verdict (same gating as today, where these
  guards ride the ordinary suite); with the key false, a red bundle is visible and
  recorded but does not red the verdict. Proven by the verdict-collection test in M5's
  test family.

## AC-CCI-009-1 — Race shard partition exactness (release-blocking)

- When: the partition check runs after M6 (the milestone lands the exact one-command form;
  shape: `go test -list` over ./... diffed against the union and intersection of the four
  internal/cli shard selectors, the rosterguard selector, and the remainder leg — the
  remainder leg's set computed by explicit set difference from the full list, never as
  `-list -skip`: measured 2026-10-09, `-list` honors `-run` but ignores `-skip`
  (internal/factory: `-list '.*' -skip '^TestMergeStep'` still lists all 22
  `TestMergeStep*` names while `-list '^TestMergeStep'` lists exactly those 22); see
  plan.md M5's recorded measurement and set-difference form).
- Then: union == full test set, pairwise intersections empty. RED-now (ledger EV-CCI-F,
  tree c47aeda2d): command `grep -c "name: Race Test" .github/workflows/ci.yml` → stdout
  `2`, exit code 0 — the split is the 2-way form; the 4-way internal/cli split and the
  standalone rosterguard shard are absent. Flips at M6.

## AC-CCI-009-2 — Standalone rosterguard shard

- When: the race job list is read from ci.yml after M6.
- Then: a job exists whose selector covers internal/harness/rosterguard and no other
  package's tests (rosterguard package at internal/harness/rosterguard/, 4 test files at
  HEAD — numeral_test.go, numeral_rederivation_test.go, root_test.go, rosterguard_test.go).

## AC-CCI-009-3 — Longest-test repair (release-blocking; ordering-baseline)

- Given: the M6 re-measure identifying the dominant internal/cli race test (card figure
  ~369s; in-tree corroboration "internal/cli ~379s/70%" ci.yml:281).
- When: `go test -json -race -count=1 <the identified test's package selector>` runs
  after the repair.
- Then: the identified test's duration is below the threshold M6 records from its own
  pre-repair measurement. Ordering clause: the pre-repair timing artifact (the RED cell)
  lands in its own commit BEFORE the repair commit — the commit graph is the only
  sequencing witness (verification-claim-integrity.md §2.3).
- RED-now (ledger EV-CCI-G, tree c47aeda2d): command
  `grep -c "379s" .github/workflows/ci.yml` → stdout `2`, exit code 0 — the committed
  dominance measurement ("internal/cli ~379s/70%", ci.yml:281) stands with no repair
  artifact in-tree; the investigation-C per-test identity is not in the tree either
  (research.md R11), so the repair has not happened. Flips at M6.

## AC-CCI-010-1 — Verdict recorded from a real read (release-blocking)

- Given: a completed candidate run.
- When: the verdict observation path runs (M4's gh-based read).
- Then: the record's verdict is `green` or `red` (never left `pending` by a completed
  run), and the record carries the run identity + observation timestamp. VERDICT-SHA
  BINDING: a run's verdict may be recorded for a candidate only when the observed run's
  head SHA equals the record's candidate commit SHA AND the run's ref equals the
  record's candidate branch — a completed run matching neither (a late-arriving run
  from a superseded candidate, same `ci/<card>` ref) is discarded, never recorded as
  the current candidate's verdict. Verification expression (the stale-run case): after
  a re-candidate produces candidate SHA C2, an observed green for the prior candidate
  C1 on the same ref writes NOTHING to C2's record — C2's verdict stays `pending`.
  The record's integration target (branch + tip) is the binding the landing check
  compares the merge against (AC-CCI-011-2's target-mismatch case) — a verdict without
  a matching target is not admissible evidence for any merge.
  Proven in the test family with a scripted run-state double; no live-CI dependency in
  unit tests.
- RED-now (ledger EV-CCI-H, tree c47aeda2d): command
  `go test -list '^TestCandidateVerdict$' ./internal/factory/` → stdout
  `ok  github.com/modu-ai/moai-adk/internal/factory	0.226s` (zero test names listed),
  exit code 0 — no verdict-observation path exists to bind. Flips at M4.

## AC-CCI-011-1 — Landing gate admits green only (release-blocking; BOTH call sites)

- Given: a merge step at gate 5 with pinned SHA P.
- When: the candidate record for (card, P) is `green`.
- Then: `RunMergeStep` proceeds past gate 5 to the collision probe — at BOTH call sites:
  the integration merge verb (internal/cli/integration_merge.go:100) AND the factory
  complete self-issued merge path (internal/cli/factory_card.go:1977, whose seams
  today carry only `ReadCard` — integration_merge_step.go:305 gates on
  `seams.LandingCheck != nil`, so an unwired call site skips the gate silently).
  Fail-closed: with the key enabled, a call site that wires no LandingCheck refuses
  rather than merging unchecked.
- RED-now: today the verb's seam is wired to a LOUD refusal placeholder
  (internal/cli/integration_merge.go:95-99) — with the key absent the seam is nil
  (no-op); the complete call site wires no seam at all; the NEW behavior only exists
  after M4. Flips at M4.

## AC-CCI-011-2 — Red/missing/stale/target-mismatch refuse with cause 5 (release-blocking; BOTH call sites)

- When: the record verdict is `red`, or no record exists for (card, P), or the record's
  candidate commit no longer descends from P, or the record's integration target does
  not match the merge destination — the record's integration branch differs from the
  merge's resolved branch (e.g. the candidate was built on the config target while the
  window record names `--branch <other>`), or the target's current tip has advanced
  past the candidate commit's first parent (verification-completeness: the verified
  tree is no longer the tree about to be merged).
- Then: `MergeExitCode(err)` == 5 (`MergeExitLandingRefused`,
  internal/factory/integration_merge_step.go:43) and the window releases (pre-merge
  cause class) — observed at BOTH call sites, `integration merge --card` and
  `factory complete`, the self-issued path included. A BRANCH MISMATCH refuses outright
  (a candidate verified against another target never admits, whatever its verdict); a
  TIP ADVANCE refuses with re-candidate guidance (the cause-2 re-measure-and-re-acquire
  discipline). The target-mismatch refusals are proven by
  `TestLandingCheckRefusesTargetMismatch` in M4's family (the run phase creates it).
- RED-now (ledger EV-CCI-I, tree c47aeda2d): command
  `go test -list '^TestLandingCheck$' ./internal/factory/` → stdout
  `ok  github.com/modu-ai/moai-adk/internal/factory	0.193s` (zero test names listed),
  exit code 0 — no landing-check test exists, and no production path can raise cause 5
  today (the only wiring is the placeholder that never activates). Flips at M4.

## AC-CCI-012-1 — Red candidate holds the OWNING card only (release-blocking; per-card)

- When: a candidate verdict is observed red for card A.
- Then:
  (a) card A's `integration acquire --card A` refuses naming card A, the red verdict,
  and the pinned SHA, with the window record AND the window policy byte-unchanged
  (the shared hold is never written by a candidate verdict);
  (b) NON-INTERFERENCE: a second card B whose candidate is green acquires the window and
  merges unaffected in the same state — one card's red must not freeze another card's
  green (AGENTS.local.md:197: 락은 병합을 직렬화하는 장치이지 수리를 직렬화하는 장치가
  아니다);
  (c) card A's own merge attempt refuses with cause 5 (AC-CCI-011-2);
  (d) a re-candidate for card A producing a green verdict clears the hold.
- RED-now (ledger EV-CCI-J, tree c47aeda2d): command
  `go test -list '^TestCandidateAcquirePrecondition$' ./internal/cli/` → stdout
  `ok  github.com/modu-ai/moai-adk/internal/cli	1.006s` (zero test names listed), exit
  code 0 — no per-card acquire precondition exists; the acquire path has no
  candidate-aware refusal today. Flips at M4.
- Why not the shared policy: the policy-hold refusal is card-blind by measurement —
  internal/factory/integration_lock.go:416 refuses every acquisition regardless of
  `want.Card`, and TestAcquireUnderHoldRefusesNamingReason +
  TestAcquireWaitUnderHoldOnEmptyWindowMustEnqueueNotGrant (both PASS) pin that
  blanket-refusal shape. Proven by `TestCandidateAcquirePrecondition` in M4's family
  (the run phase creates it following the `acquireSettingsDriftPrecondition` test
  shape).

## AC-CCI-013-1 — Single recorded retry (release-blocking)

- Given: the flaky registry contains a test T (with its evidence citation).
- When: T fails in a candidate run and the retry wrapper re-runs it.
- Then: exactly one retry executes; the job summary records "retried: T, attempt 2,
  outcome pass|fail"; a second failure fails the run; a NON-registry failure performs
  zero retries. Proven by the wrapper's fixture-stream test (test-census.sh precedent)
  plus one observed CI run carrying a registry retry.
- RED-now (ledger EV-CCI-K, tree c47aeda2d): command
  `ls scripts/ci/flaky-registry.txt` → stdout empty, stderr
  `ls: scripts/ci/flaky-registry.txt: No such file or directory`, exit code 1 — an
  empty stdout with a non-zero exit is a complete observation (the file, the registry,
  and the retry wrapper do not exist). Flips at M6.

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
- RED-now (ledger EV-CCI-L, tree c47aeda2d): command
  `grep -c "ci/\*\*" .github/workflows/ci.yml` → stdout `0`, exit code 1 — no `ci/**`
  trigger exists, so no candidate run (and no vet legs on one) can occur. Flips at M3
  (trigger) with the vet-legs observation completing at M5.

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

---

## RED-now evidence ledger

Fourteen entries measured 2026-10-09 on tree **c47aeda2d** (worktree
.moai/worktrees/t1478). The Go tree at c47aeda2d is code-identical to the plan-phase
baseline db0c514d3 — commits ec8180a38, 09380a9aa, c47aeda2d touch only `.moai/specs/`
artifacts — so every entry is equally a db0c514d3 observation. Each entry carries the
four elements: the command (single invocation), its verbatim stdout, its exit code, and
the tree SHA. Entries whose command produced no stdout say so; an empty stdout with a
non-zero exit is a complete observation (verification-completeness.md §2.1).

### EV-CCI-A — serves AC-CCI-002-1, 002-2, 003-1

- Command: `go run ./cmd/moai integration candidate --card t1478`
- stdout (verbatim; the renderer's blank-box padding lines elided):
  ```
  ERROR
  Unknown flag: --card.

    Try --help for usage.

  exit status 1
  ```
- Exit code: 1
- Tree: c47aeda2d
- Why red: the candidate verb does not exist — the invocation dies on an unknown flag
  before any candidate behavior; no candidate commit, branch, or conflict path exists.

### EV-CCI-B — serves AC-CCI-004-1

- Command: `go test -list '^TestIntegrationCandidate$' ./internal/cli/`
- stdout (verbatim):
  ```
  ok  	github.com/modu-ai/moai-adk/internal/cli	1.403s
  ```
  (zero test names listed above the `ok` line)
- Exit code: 0
- Tree: c47aeda2d
- Why red: the M2 test family that would observe the resolution is absent — the swept
  set is empty and the absence IS the observation.

### EV-CCI-C — serves AC-CCI-005-1

- Command: `go test -list '^TestCandidateRecord$' ./internal/factory/`
- stdout (verbatim):
  ```
  ok  	github.com/modu-ai/moai-adk/internal/factory	0.442s
  ```
  (zero test names listed)
- Exit code: 0
- Tree: c47aeda2d
- Why red: no candidate-record type, store, or test exists.

### EV-CCI-D — serves AC-CCI-006-3

- Command: `grep -c candidate_ci .moai/config/sections/workflow.yaml`
- stdout (verbatim): `0`
- Exit code: 1
- Tree: c47aeda2d
- Why red: the key is absent from the live config (mirror likewise, AC-CCI-006-1's
  measurement), so no gated refusal exists.

### EV-CCI-E — serves AC-CCI-008-1

- Command: `grep -c guard-bundle .github/workflows/ci.yml`
- stdout (verbatim): `0`
- Exit code: 1
- Tree: c47aeda2d
- Why red: no guard-bundle job exists in the workflow.

### EV-CCI-F — serves AC-CCI-009-1

- Command: `grep -c "name: Race Test" .github/workflows/ci.yml`
- stdout (verbatim): `2`
- Exit code: 0
- Tree: c47aeda2d
- Why red: exactly the 2-way race split exists (test-race-1/test-race-2); the 4-way
  internal/cli split and the standalone rosterguard shard do not.

### EV-CCI-G — serves AC-CCI-009-3

- Command: `grep -c "379s" .github/workflows/ci.yml`
- stdout (verbatim): `2`
- Exit code: 0
- Tree: c47aeda2d
- Why red: the committed dominance measurement ("internal/cli ~379s/70%", ci.yml:281
  and its duplicate comment) stands with no repair artifact in-tree; the
  investigation-C per-test identity is not in the tree (research.md R11).

### EV-CCI-H — serves AC-CCI-010-1

- Command: `go test -list '^TestCandidateVerdict$' ./internal/factory/`
- stdout (verbatim):
  ```
  ok  	github.com/modu-ai/moai-adk/internal/factory	0.226s
  ```
  (zero test names listed)
- Exit code: 0
- Tree: c47aeda2d
- Why red: no verdict-observation path exists to bind to the candidate SHA.

### EV-CCI-I — serves AC-CCI-011-2

- Command: `go test -list '^TestLandingCheck$' ./internal/factory/`
- stdout (verbatim):
  ```
  ok  	github.com/modu-ai/moai-adk/internal/factory	0.193s
  ```
  (zero test names listed)
- Exit code: 0
- Tree: c47aeda2d
- Why red: no landing-check test exists, and no production path can raise cause 5 —
  the only LandingCheck wiring is the placeholder that never activates (constant-false
  key, internal/cli/integration_merge.go:186-189).

### EV-CCI-J — serves AC-CCI-012-1

- Command: `go test -list '^TestCandidateAcquirePrecondition$' ./internal/cli/`
- stdout (verbatim):
  ```
  ok  	github.com/modu-ai/moai-adk/internal/cli	1.006s
  ```
  (zero test names listed)
- Exit code: 0
- Tree: c47aeda2d
- Why red: no candidate-aware acquire precondition exists — the acquire path has no
  per-card refusal today.

### EV-CCI-K — serves AC-CCI-013-1

- Command: `ls scripts/ci/flaky-registry.txt`
- stdout (verbatim): (empty)
- stderr (verbatim): `ls: scripts/ci/flaky-registry.txt: No such file or directory`
- Exit code: 1
- Tree: c47aeda2d
- Why red: the registry file, the registry, and the retry wrapper do not exist.

### EV-CCI-L — serves AC-CCI-015-1

- Command: `grep -c "ci/\*\*" .github/workflows/ci.yml`
- stdout (verbatim): `0`
- Exit code: 1
- Tree: c47aeda2d
- Why red: no `ci/**` trigger exists — no candidate run can occur, so no vet legs on
  one are observable.
