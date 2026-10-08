---
id: SPEC-CANDIDATE-CI-001
title: "Pre-landing candidate CI — implementation plan"
version: "0.1.0"
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
card: t1478
---

# plan.md — SPEC-CANDIDATE-CI-001 (card t1478, Tier L)

## §A. Context

The card lands a pre-landing candidate CI path for the factory: `moai integration
candidate` builds the would-be merge commit with `git merge-tree --write-tree` +
`git commit-tree`, pushes it to `ci/<card>`, and the CI workflow (trigger extended with
`ci/**`) judges it. The merge step's pre-allocated gate-5 seam
(`MergeStepSeams.LandingCheck`) then admits a merge only on a green candidate for the
same pinned SHA. Two anchors are already in the tree at HEAD db0c514d3:
`candidateCIEnabled` (internal/cli/integration_merge.go:186-189, constant-false
placeholder owning the key name `workflow.candidate_ci.enabled`) and the LandingCheck
seam declaration (internal/factory/integration_merge_step.go:104-106). The card also
restructures the CI workflow itself: guard bundle job, race 4+1 split with the ~369s test
repaired, known-flaky single retry, integration-branch cancel-in-progress exception, and
linux/windows vet on candidates.

Milestone order below follows decision-reversibility: the config key and record schema
(the decisions most likely to be revised by review) first, the verb and gate next, the
workflow YAML last (each YAML change is a single revert).

## §B. Known Issues

- B1: The `CCI` REQ mnemonic collides with SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
  (internal/cli/update.go:113). Disclosed in spec.md §D/§H; not fixable here without
  dangling the committed anchors.
- B2: The ~369s test figure is operator-provided (investigation C, card text); the tree
  carries the corroborating "internal/cli ~379s/70%" measurement
  (.github/workflows/ci.yml:281) but no per-test 369s record. M6 re-measures before
  repairing; the repair target is whatever the re-measure names as the dominant single
  test.
- B3: The 12 guard-drift kinds are investigation-C counts, not an in-tree registry. M5
  bundles the three families the card names; enumerating the exact 12 is repair work the
  bundle makes attributable, not a deliverable of this SPEC.
- B4: `go run ./cmd/moai integration candidate --help` on db0c514d3 exits 0 printing the
  PARENT command's help (cobra fall-through; measured 2026-10-09). AC-CCI-001-1's RED-now
  is therefore "no candidate usage line in the help output", not a non-zero exit.
- B5: The `pushed → ci-green` reserved edge refuses with "owned by F3"
  (internal/homestate/card_transition.go:294-295). This SPEC leaves that refusal intact
  (spec.md §G); if the plan-auditor reads the auto-hold as requiring that edge, the answer
  is no — the auto-hold is the window policy hold (REQ-CCI-012), not a card-state
  transition.
- B6: CI runs triggered from `ci/**` push events carry no `head_ref`; any job step
  branching on `startsWith(github.head_ref, ...)` (ci.yml:261, :320) reads false on
  candidates — the release/* skip conditions stay harmless, but M3 must re-read each
  `if:` after adding the trigger.

## §C. Pre-flight

- [ ] Confirm HEAD is the integration base this plan pinned: `git rev-parse --short HEAD`
      → db0c514d3 (re-read at run start; never trust this file's value).
- [ ] `go build ./... && go test ./internal/cli/ ./internal/factory/ ./internal/config/ -count=1`
      green before the first edit (baseline attribution).
- [ ] `git grep -n "candidateCIEnabled" internal/` → exactly the placeholder and its two
      call sites (integration_merge.go:95, :186) before M1 replaces it.
- [ ] Confirm the active integration branch resolves as the config says:
      `go run ./cmd/moai integration status` (git-strategy.yaml:9 still `workflow:
      git-flow`, `develop_branch: develop` — the code path is the resolver
      internal/cli/integration.go:280-282, never a literal `develop`).

## §D. Constraints

- [HARD] Template-First for the config key: `.moai/config/sections/workflow.yaml` AND
  `internal/template/templates/.moai/config/sections/workflow.yaml` change together; the
  struct side must satisfy the existing YAML-symmetry audit
  (internal/config/audit_struct_yaml_symmetry_test.go).
- [HARD] No integration-window gate reordering; the landing check joins at gate 5
  exactly where the seam already sits (integration_merge_step.go:303-309).
- [HARD] The candidate verb never mutates the integration window record and never merges.
- [HARD] ci.yml edits keep the required-check NAME sets stable (the test-skip-marker
  parity contract, ci.yml:379-384) unless a milestone explicitly renames a check.
- [HARD] No `git push --force` to any protected ref; candidate branch replacement is
  scoped to `ci/**` refs only (AGENTS.md destructive-primitive rule).
- Every workflow YAML change passes the repo's workflow parse validation (yamllint /
  actionlint, whichever the repo's tooling carries; verify the exact validator in M3
  pre-flight, do not assume).

## §E. Self-Verification

Each milestone's verification command is run and its output observed before the milestone
is reported complete; the §E.2 run-phase evidence table records command + verbatim output
per the evidence-bearing report format. Coverage claims name the package and are
re-measured in the run phase, not carried from this plan.

## §F. Milestones

Priority labels: all milestones P1 unless marked; ordering is by dependency and
decision-reversibility, never by time.

### M1 — Config gate + candidate record schema (P1, decision-bearing first)

Files touched:
- `.moai/config/sections/workflow.yaml` — add `candidate_ci:` block (`enabled: false`,
  `guard_bundle_required: true`) following the section's existing key-comment style.
- `internal/template/templates/.moai/config/sections/workflow.yaml` — mirror (Template-First).
- `internal/config/types.go` — `CandidateCIConfig` struct + `yaml:"candidate_ci"` on the
  Workflow struct; defaults in `internal/config/defaults.go`.
- `internal/factory/candidate_record.go` (new) — the candidate record type + store
  (read/write/lookup keyed card+pinned SHA) in the primary checkout's state store, the
  same root resolution the integration lock uses.
- `internal/cli/integration_merge.go` — replace the constant-false `candidateCIEnabled`
  body with the real config read (the @MX:UPGRADE this landing fulfills); keep the verb's
  LandingCheck wiring a LOUD refusal until M4 replaces it.

Test families: `go test ./internal/config/...` (loader + YAML symmetry audits),
`go test ./internal/factory/ -run '^TestCandidateRecord$'` (record round-trip, key lookup,
absent-key = refuse — the run phase creates `TestCandidateRecord` in
internal/factory/candidate_record_test.go), `go test ./internal/cli/ -run '^TestCandidateCIEnabled$'`
(the run phase creates `TestCandidateCIEnabled` replacing the placeholder's tests).
The config-side gate-flag precedent names the intended new tests:
`TestWorkflowCandidateCIDefaultFalse` / `TestWorkflowCandidateCILoaderRoundTrip`, after
the existing `TestWorkflowSettingsDriftGateDefaultFalse` /
`TestWorkflowSettingsDriftGateLoaderRoundTrip` pair.

Verification command:
`go test ./internal/config/ -run '^(TestAuditLoaderCompleteness|TestWorkflowConfigFields|TestNewDefaultWorkflowConfig|TestNewDefaultWorkflowConfigNestedDefaults)$' -count=1 && go test ./internal/factory/ -run '^TestCandidateRecord$' -count=1 && go vet ./internal/cli/`
(every branch is an EXISTING exact test name, measured via `go test -list`:
TestAuditLoaderCompleteness at internal/config/audit_loader_completeness_test.go:58;
TestWorkflowConfigFields, TestNewDefaultWorkflowConfig, and
TestNewDefaultWorkflowConfigNestedDefaults in the same package).

### M2 — The candidate verb (P1)

Files touched:
- `internal/cli/integration_candidate.go` (new) — the verb: config gate → card read →
  REQ-CCI-004 resolution (reuse/extend the merge step's resolver contract) →
  `git merge-tree --write-tree` → conflict refusal → `git commit-tree -p <tip> -p <pinned>`
  → push to `ci/<card>` (replace-in-place) → candidate record write (verdict `pending`).
- `internal/cli/integration.go` — register the subcommand in the existing AddCommand list
  (:269).
- `internal/factory/candidate_record.go` — push-failure path leaves a prior verdict
  untouched (REQ-CCI-017).

Test families: `go test ./internal/cli/ -run '^TestIntegrationCandidate$'` (verb gating,
conflict refusal, id validation, push-failure path), reusing the package's temporary-origin
test helper (internal/factory/temp_origin.go pattern) for a fake remote; the merge-tree
two-parent shape asserted from `git cat-file -p` in the temp repo.

Verification command:
`go test ./internal/cli/ -run '^TestIntegrationCandidate$' -count=1 -v` plus one manual
temp-clone smoke: `moai integration candidate --card <test-card>` against a scratch remote.

### M3 — CI trigger + concurrency policy (P1)

Files touched:
- `.github/workflows/ci.yml` — `push.branches` gains `ci/**` (:16-18); concurrency block
  becomes ref-conditional: integration branch ref never cancelled, `ci/**` refs supersede
  (:35-37); re-read every `if: startsWith(github.head_ref, ...)` for candidate pushes (B6).

Test families: workflow parse validation (the repo's YAML/Actions validator — confirm the
exact tool in pre-flight); `go test` untouched. The observable is one real `ci/*` push run
after this milestone lands on the integration branch.

Verification command:
`yamllint -s .github/workflows/ci.yml` (or the repo's parse guard equivalent) plus
`git grep -n "ci/\*\*" .github/workflows/ci.yml` showing the trigger and concurrency rows.

### M4 — Landing gate + red per-card hold (P1)

Files touched:
- `internal/cli/integration_merge.go` — replace the LOUD-refusal LandingCheck placeholder
  (:95-99) with the real check: read the candidate record for (card, pinned SHA), require
  green + ancestry; refuse cause 5 otherwise.
- `internal/cli/factory_card.go` — wire the SAME shared LandingCheck at the complete
  call site (:1977-1988, the self-issued merge path): its seams today carry ONLY
  `ReadCard`, and the step gates on `seams.LandingCheck != nil`
  (integration_merge_step.go:305), so without this wiring `factory complete` merges with
  no landing gate even after the key is enabled (audited plan-auditor defect D2).
- `internal/factory/integration_merge_step.go` — the fail-closed guard in gate 5: when
  the key is enabled and the seam is nil, refuse (cause-5 class) instead of merging
  unchecked — no caller, current or future, omits the gate.
- `internal/factory/candidate_record.go` — verdict observation update path (REQ-CCI-010):
  read CI run state via the gh surface, write verdict + run identity + observed-at;
  enforce the verdict-SHA binding — run head SHA == record candidate SHA and run ref ==
  record candidate branch, else the observation is discarded.
- `internal/cli/integration.go` (acquire verb) — the card-aware candidate precondition
  (REQ-CCI-012's per-card hold): when the ACQUIRING card's candidate record reads red,
  refuse naming card + verdict + pinned SHA BEFORE any window-record mutation — the
  settings-drift precondition's position and shape (integration.go:427-436). NO shared
  `IntegrationWindowPolicy` write: the policy hold is card-blind by measurement
  (integration_lock.go:416 refuses every acquire regardless of `want.Card`;
  TestAcquireUnderHoldRefusesNamingReason and
  TestAcquireWaitUnderHoldOnEmptyWindowMustEnqueueNotGrant both PASS), so a global hold
  would freeze green-candidate cards against AGENTS.local.md:197. The signature needs NO
  extension — `want.Card` already travels; extending the factory hold branch on
  `want.Card` is the sanctioned fallback only.

Test families: `go test ./internal/factory/ -run '^(TestLandingCheck|TestCandidateVerdict)$'`
(green admits past gate 5; red/missing/stale refuse with MergeExitLandingRefused; the
stale-run observation — an older green arriving for a superseded candidate on the same
ref — writes nothing; the run phase creates `TestLandingCheck` and `TestCandidateVerdict`
beside the existing family); `go test ./internal/cli/ -run '^TestCompleteRefusesRedCandidate$'`
(the complete call site refuses a red/missing candidate with cause 5 when the key is
true — the run phase creates `TestCompleteRefusesRedCandidate` in the factory_card test
family); `go test ./internal/cli/ -run '^TestCandidateAcquirePrecondition$'` (the per-card
hold: card A red → A's acquire refused with record+policy byte-unchanged; card B green
acquires and merges unaffected in the same state; a green re-candidate clears the hold —
the run phase creates `TestCandidateAcquirePrecondition` following the
`acquireSettingsDriftPrecondition` test shape); `go test ./internal/factory/ -run '^(TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepPreMergeCausesReleaseWithDistinctCodes)$'`
(regression: the 13-cause gate order unchanged, exit codes 1-15 stable — the two EXISTING
contract tests at internal/factory/integration_merge_step_test.go:210 and :320; the full
`TestMergeStep*` family of 22 runs in the ordinary suite).

Verification command:
`go test ./internal/factory/ ./internal/cli/ -run '^(TestLandingCheck|TestCandidateVerdict|TestCandidateAcquirePrecondition|TestCompleteRefusesRedCandidate|TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestIntegrationCandidate)$' -count=1`

### M5 — Guard bundle job, guard/ordinary separation, candidate vet legs (P2)

Files touched:
- `internal/cli/` — the three guard families' test membership (source-scan static-parse
  guards, line-key guards) is recorded as an EXPLICIT selector alternation, measured at
  run start (`go test -list`) and kept in the workflow step. Separation mechanism:
  SELECTOR-based, not build tags — the tree's established test-partition convention is
  `-run`/`-skip` (the race split at ci.yml:298/:357); the cli test build tags that exist
  are OS/backend conditionals, not a family convention. The required `test` job's
  `go test ./...` gains the complementary `-skip '<guard-selector>'`. Without this
  separation the bundle is cosmetic: a guard failure still reddens the ordinary test
  job, so `guard_bundle_required: false` could not mean anything (AC-CCI-008-2). The
  two scopes must PARTITION: no guard test in both jobs (silent duplicate), none in
  neither (silent gap — the empty-sweep hazard,
  verification-completeness.md §1.1). The census family needs no go-test separation —
  it is shell (`bash scripts/ci-census/census-check.sh`) and moves to the bundle as-is.
- `.github/workflows/ci.yml` — new `guard-bundle` job running ONLY the three families
  (the guard go-test selector + census-check.sh) as one named check; the
  `test` job gains the complementary skip; the verdict-collecting step implements
  `workflow.candidate_ci.guard_bundle_required` at the WORKFLOW layer (a required-check
  name is static — the key governs whether the bundle's result is admitted into the
  candidate verdict).
- Confirm the build job's vet matrix (:577-582) runs on candidate pushes (trigger-only
  change; assert in the same milestone's verification).

Test families: the guard families themselves (they run inside the bundle — the bundle
selector's `-list` count is the family count); the census fixture check.

Verification commands:
- Bundle scope runs guards only: `go test -list '<guard-selector>' ./internal/cli/`
  names exactly the family set (count N), and
  `go test -run '^<guard-selector>$' -count=1 ./internal/cli/` is green with every RUN
  line inside the selector.
- Ordinary scope runs without guards: `go test -skip '<guard-selector>' -count=1 ./internal/cli/`
  green, and `go test -skip '<guard-selector>' -count=1 -v ./internal/cli/ | grep -c "=== RUN <sample-guard-name>"`
  reads 0 — a guard name absent from the ordinary scope's output.
- MEASURED (2026-10-09, this tree — the `-list`/`-skip` interaction): 
  `go test -list '.*' -skip '^TestMergeStep' ./internal/factory/` STILL lists all 22
  `TestMergeStep*` names, while `go test -list '^TestMergeStep' ./internal/factory/`
  lists exactly those 22 — **`-list` honors `-run` but IGNORES `-skip`**. The ordinary
  set is therefore NEVER computed as `-list -skip`; it is the explicit set difference
  below.
- Partition exactness by explicit set difference (ordinary = full − guards):
  ```
  go test -list '.*' ./internal/cli/ | grep '^Test' | sort > /tmp/m5-all.txt
  go test -list '<guard-selector>' ./internal/cli/ | grep '^Test' | sort > /tmp/m5-guard.txt
  comm -23 /tmp/m5-all.txt /tmp/m5-guard.txt > /tmp/m5-ordinary.txt        # full − guards
  comm -12 /tmp/m5-ordinary.txt /tmp/m5-guard.txt | wc -l                  # 0 — ordinary ∩ guard empty (guards ⊆ all, so the operands are the derived ordinary set and the guard set)
  sort -u /tmp/m5-ordinary.txt /tmp/m5-guard.txt | diff - /tmp/m5-all.txt  # no output — union == full
  ```
  and the ordinary-scope RUN's observed test set matches `/tmp/m5-ordinary.txt` (the
  `-v` `=== RUN` names contain no guard name — the grep -c 0 check above is its quick
  form).
- Census fixture: `bash scripts/ci-census/census-check.sh` exits 0.
- Workflow: `git grep -n "guard-bundle" .github/workflows/ci.yml` shows the job and the
  test job's complementary skip.

### M6 — Race split restructure + flaky single retry (P2)

Files touched:
- `.github/workflows/ci.yml` — replace the 2-way race split (:258-374) with four
  internal/cli shards + one rosterguard standalone shard + remainder leg, selectors formed
  as a partition measured from `go test -list`; keep the ceiling-headroom discipline in
  each step comment (ci.yml:291-297 style).
- The ~369s internal/cli test (B2: re-measure first) — repair per its measured cause.
- `scripts/ci/flaky-registry.txt` (new) + the retry wrapper the race/test jobs call —
  exactly-one recorded retry for registry tests (REQ-CCI-013).
- `.github/test-input-filters.yml` — register BOTH new paths under `go_code`:
  `- 'scripts/ci/**'` (the flaky registry + retry wrapper), following the census
  precedent already at :63 (`- 'scripts/ci-census/**'` — same rationale: the test/race
  jobs shell out to them) with an evidence comment in the file's per-root style naming
  the consumer steps. WITHOUT the registration, a diff touching only the wrapper or the
  registry reads go_code=false, the skip-marker stub satisfies the required check
  (ci.yml:376-384), and the retry logic lands over a path CI never exercised.

Test families: the repartitioned shards (each exercised in CI), the repaired test and its
package suite, the retry wrapper unit-checked against a fixture stream
(scripts/ci-census/test-census.sh precedent).

Verification command:
partition check — `go test -list '.*' ./... | sort` diffed against the union of shard
selectors (the milestone records the exact one-command form it lands);
`bash scripts/ci/census-check.sh`-style fixture run for the retry wrapper;
filter registration — `grep -n "scripts/ci" .github/test-input-filters.yml` shows the
new `scripts/ci/**` pattern;
script-only-change observance — after landing, push a candidate whose ONLY diff is a
comment edit inside `scripts/ci/flaky-registry.txt` and read the run's job list: the
detect job must report go_code=true and the test/race jobs must EXECUTE (not
skip-marker) — the retry logic's CI surface exercised by its own file's change.

### M7 (sync phase, not a run milestone) — Doctrine amendment

The canonical amendment target is `AGENTS.local.md` §4.1 (`### §4.1 통합 체인 (main)`,
AGENTS.local.md:175 — the post-cutover integration-chain canon: card branches from
`main`, card PRs via the serial integration window, leader-batch push, CI on
`origin/main` as the verdict): it gains the candidate-path exception (REQ-CCI-016)
recording the 2026-10-03 operator approval. The develop-era procedure docs that
AGENTS.local.md:203 already flags as drift — `.claude/rules/local/gitflow-lane-protocol.md`
§2/§4, `.moai/docs/git-workflow-doctrine.md`, `.moai/docs/git-local-workflow-doctrine.md`,
`.moai/docs/gitflow-integration-chain.md` — gain the same candidate-path exception WITH
their develop-era drift notice retained (폐기 표시 유지). CLAUDE.local.md is RETIRED
(AGENTS.local.md §0.3, lines 27-29: a discarded model, never a citation target) — it is
neither read nor amended. Owned by manager-docs in sync; listed here so the landing is
not judged incomplete without it (AGENTS.local.md §4.1 is the doctrine the lanes
actually read, per its own §0.4 canonical-copy registry).

## §G. Anti-Patterns

- Do NOT resolve the integration branch by literal name (`develop`/`main`) anywhere — the
  resolver is the contract (internal/cli/integration.go:280-282); the GitHub Flow cutover
  must not silently break the candidate path.
- Do NOT let the candidate verb acquire or touch the integration window.
- Do NOT widen the §4.1 amendment beyond the candidate path — integration-branch push
  stays leader-batched.
- Do NOT retry non-registry test failures (masks real regressions; the card allows exactly
  one recorded retry for KNOWN flaky tests).
- Do NOT add a `paths:` filter that would skip ci.yml changes from CI on candidates —
  a workflow-only candidate still must run (B6 family).
- Do NOT delete or repurpose the existing race-job census steps when re-sharding; the
  census rides every shard.

## §H. Cross-References

- spec.md §D (requirements), acceptance.md (AC matrix), design.md (D1-D9 decisions),
  research.md (file:line evidence for every existing-behavior claim).
- internal/factory/integration_merge_step.go (the 13-cause gate order), internal/cli/integration.go
  (the verb family), .github/workflows/ci.yml (the workflow this card restructures),
  .claude/rules/local/gitflow-lane-protocol.md (the doctrine this card amends).
