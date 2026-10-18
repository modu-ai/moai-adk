# Acceptance: SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

Document-level tree pin (binds every criterion that carries no pin of its own): `fb749d372e61a6b10c53091362c76b4412827427`, except the scope-fence commands G-15 to G-15d and the sub-criteria cells of the 2026-10-19 amendment (spec.md HISTORY 0.1.1), which carry the measurement HEAD `9fc118b44ec917f298809968f4034c26af4f0336` named in §E. Verification layer: Given-When-Then entries; the requirement layer is spec.md §B (GEARS).

## §A — Authoring rules applied

1. **Plain commands.** Every command is one plain pipeline, one `git` per line, no `$()`, no process substitution, no shell variables.
2. **Classification.** Every criterion is **release-blocking**: each carries its own RED-now cell (§E) with command, verbatim stdout, exit code, and pin, re-executed in this revision, plus a green-path cell naming the milestone that flips it. The Definition of Done gates all 16. No criterion of the original 16 is a regression-guard. The four sub-criteria added by the 2026-10-19 amendment (AC-ISB-004a, AC-ISB-006a, AC-ISB-006b, AC-ISB-006c) are **regression-guards**, not release-blocking: their RED-now cannot be re-executed on the current tree (the pre-repair code is history; the only record of the failing run is the repair commit message), so under the undecidable disposition of verification-completeness §2.1 they lose release-blocking eligibility and are not recorded as a pass by this amendment. They sit under logical criteria AC-ISB-004 and AC-ISB-006 and add no identifier to the Tier M budget of 16.
3. **What the plan-phase RED is.** At the pin no gate code and no gate test exists, so a `go test` run cannot be a RED (it sweeps nothing and exits 0). Each criterion's plan-phase RED cell is instead its own read-only command that fails today on the committed tree and reproduces: a search for the criterion's own planned test (empty output, exit 1), or a search or count over the artifact the work must create or change. The behavioural RED (the test exists and fails against a gate that lacks the behaviour) is established in the run phase by the committed RED commit that precedes the GREEN commit in the commit graph (verification-claim-integrity §2.3); the plan-phase cell says only "not yet observable: the test is absent".
4. **Swept count.** Every Go-test command in §D names its tests; a green counts only when `-v` shows each named test `--- PASS`.

## §B — Criteria matrix

| ID | Requirement | Class | RED-now (ledger id, §E) | Green path (§D command, milestone) |
|----|-------------|-------|--------------------------|--------------------------------------|
| AC-ISB-001 | REQ-ISB-001 | release-blocking | E-12a | G-01, M3 |
| AC-ISB-002 | REQ-ISB-002 | release-blocking | E-12b | G-02, M3 |
| AC-ISB-003 | REQ-ISB-003 | release-blocking | E-12c | G-03, M2 |
| AC-ISB-004 | REQ-ISB-004 | release-blocking | E-12d | G-04, M2 |
| AC-ISB-005 | REQ-ISB-005 | release-blocking | E-12e | G-05, M2 |
| AC-ISB-006 | REQ-ISB-006 | release-blocking | E-09 | G-06, G-06b, G-06c, M2 |
| AC-ISB-007 | REQ-ISB-007 | release-blocking | E-12f, E-01, E-02 | G-07, M1 |
| AC-ISB-008 | REQ-ISB-008 | release-blocking | E-12g | G-08, M3 |
| AC-ISB-009 | REQ-ISB-009 | release-blocking | E-12h | G-09, G-09b, M2, M3 |
| AC-ISB-010 | REQ-ISB-010 | release-blocking | E-12i | G-10, G-10b, G-10c, M2, M3 |
| AC-ISB-011 | REQ-ISB-011 | release-blocking | E-12j, E-05 | G-11, M3 |
| AC-ISB-012 | REQ-ISB-012 | release-blocking | E-02, E-04, E-08 | G-12, G-12b, M1, M4 |
| AC-ISB-013 | REQ-ISB-013 | release-blocking | E-03, E-06 | G-13, G-13b, M1 |
| AC-ISB-014 | REQ-ISB-014 | release-blocking | E-10, E-12k | G-14, G-14b, M3 |
| AC-ISB-015 | REQ-ISB-015 | release-blocking | E-11 | G-15, G-15b, G-15c, G-15d, M5 |
| AC-ISB-016 | REQ-ISB-016 | release-blocking | E-12l | G-16, M3 |

Sub-criteria added 2026-10-19 (regression-guards, counted inside AC-ISB-004 and AC-ISB-006; RED-now not re-executable, see §E):

| ID | Requirement | Class | RED-now (ledger id, §E) | Green path (§D command) |
|----|-------------|-------|--------------------------|--------------------------|
| AC-ISB-004a | REQ-ISB-004, REQ-ISB-005 | regression-guard | E-14a | G-17, G-17b |
| AC-ISB-006a | REQ-ISB-006 | regression-guard | E-14b | G-18 |
| AC-ISB-006b | REQ-ISB-006 | regression-guard | E-14c | G-19 |
| AC-ISB-006c | REQ-ISB-006 | regression-guard | E-14d | G-20 |

Coverage declarations: AC-ISB-001 maps REQ-ISB-001; AC-ISB-002 maps REQ-ISB-002; AC-ISB-003 maps REQ-ISB-003; AC-ISB-004 maps REQ-ISB-004; AC-ISB-005 maps REQ-ISB-005; AC-ISB-006 maps REQ-ISB-006; AC-ISB-007 maps REQ-ISB-007; AC-ISB-008 maps REQ-ISB-008; AC-ISB-009 maps REQ-ISB-009; AC-ISB-010 maps REQ-ISB-010; AC-ISB-011 maps REQ-ISB-011; AC-ISB-012 maps REQ-ISB-012; AC-ISB-013 maps REQ-ISB-013; AC-ISB-014 maps REQ-ISB-014; AC-ISB-015 maps REQ-ISB-015; AC-ISB-016 maps REQ-ISB-016; AC-ISB-004a maps REQ-ISB-004 and REQ-ISB-005; AC-ISB-006a, AC-ISB-006b and AC-ISB-006c map REQ-ISB-006.

## §C — Scenarios (Given-When-Then)

### AC-ISB-001 — acquire refuses before touching the window

**Given** the gate enabled, K = 0, a fixture repository whose local `main` has one first-parent landing after the `Base` stored in a valid re-measure record keyed to the card tree, and a window record file with known bytes.
**When** `moai integration acquire --card <id> --wait` runs from the card worktree.
**Then** it exits 16, prints a line beginning `BASE_FRESHNESS_VIOLATION:`, the window record bytes are identical to before, and the queue holds no ticket.
**Mutant probe.** A gate placed after `AcquireIntegrationWindow` fails the byte-identity clause; one placed after the `--wait` branch fails the empty-queue clause.

### AC-ISB-002 — complete refuses before acquiring

**Given** the AC-ISB-001 fixture, a merge-ready card whose integration tip is NOT an adoption (REQ-ISB-016), and a caller that does not hold the window.
**When** `moai factory complete` runs for the card.
**Then** it exits 16 with the sentinel; the window record, the integration branch tip, and the card state are unchanged.

### AC-ISB-003 — landings are first-parent, measured from the stored base

**Given** the gate enabled, and a fixture `main` of N = 3 integration merges, each merging a side branch of 4 commits, after the `Base` of a valid record.
**When** the gate measures the card.
**Then** landings behind is 3, not 15, and a record whose `Base` is not an ancestor of the tip is reported stale at every K.
**Mutant probe.** A count without `--first-parent` returns 15; a count from the live merge-base instead of `rec.Base` differs in AC-ISB-010's content-free-absorb case.

### AC-ISB-004 — the K boundary and the message content

**Given** the gate enabled and landings behind equal to K, then K + 1, with K = 0 and with K = 2.
**When** the gate evaluates each.
**Then** equal to K admits; K + 1 refuses; the refusal text contains the numeric landings behind, the numeric K, the literal key `workflow.base_freshness_gate.max_behind_landings`, the short tip, and the words `remeasure` and `retry`.
**Mutant probe.** A `>=` comparison refuses at equality; a constant K fails the K = 2 run.

### AC-ISB-004a — the remedy names `remeasure --branch <branch>` for an explicit acquire `--branch` (regression-guard)

**Given** the gate enabled, K = 0, and three refusal kinds in turn (landings behind, stale base, no record), each produced once with the gate input marked as an explicit `--branch` and once without; and a CLI fixture in which `moai integration acquire --card <id> --branch alternate` is refused because `alternate` is one landing past the recorded base, next to the same refusal without `--branch`.
**When** the pure gate evaluates each (G-17) and the CLI path runs both `acquire` forms (G-17b).
**Then** with the explicit branch every remedy ends with `moai integration remeasure --branch <branch>` (landings behind and stale base: `Absorb local <branch> into the card branch, re-run moai integration remeasure --branch <branch>, then retry.`; no record: `Run moai integration remeasure --branch <branch> after the last absorb, then retry.`); without it the three texts are byte-identical to REQ-ISB-004 and REQ-ISB-005 as first written and contain no `--branch`; each refusal stays one line.
**Mutant probe.** A gate that always names `--branch` fails the without-flag clause; one that never does fails the explicit clause; one that names the configured branch instead of the explicit one fails the `alternate` CLI case.
**Tests.** `TestBaseFreshnessRemedyNamesExplicitBranch` (`internal/factory/base_freshness_test.go`), `TestAcquireRefusalRemedyNamesTheExplicitBranch` and `TestAcquireRefusalRemedyUnchangedWithoutExplicitBranch` (`internal/cli/integration_base_freshness_test.go`).

### AC-ISB-005 — the record is required whatever L is

**Given** the gate enabled, L = 0, and four record states for the current HEAD tree: absent; present for a different (pre-absorb) tree only; present for this tree with `ExitCode` 7; present and valid.
**When** the gate evaluates each (and again at K = 2).
**Then** the first three refuse with the record message (the third names the invalid cause); the fourth admits.
**Mutant probe.** A gate that checks only L admits the first three; one that keys the record by the merge-base instead of the HEAD tree admits the second.

### AC-ISB-006 — reuse, not a parallel mechanism

**Given** the tree after M2.
**When** the reuse greps run (G-06, G-06b, G-06c).
**Then** exactly one `RemeasureRecord` type exists, the gate source calls `ValidateRemeasureRecord` and `ReadRemeasureRecord`, and the gate source contains no git `merge` invocation.
**Mutant probe.** A gate defining its own record struct raises the type count to 2 (E-07 shows the count command sees the real type today).

### AC-ISB-006a — `moai integration remeasure --branch` resolves its target as `acquire` does; no flag changes nothing (regression-guard)

**Given** the gate enabled, K = 0, a git-flow configuration whose integration target is `main`, a branch `alternate` one landing past `main`, and a card that has absorbed `alternate`.
**When** `moai integration remeasure -- true` runs without the flag, then `acquire --card <id> --branch alternate`, then `moai integration remeasure --branch alternate -- true`, then the same `acquire` (G-18).
**Then** the record written without the flag names `main`'s tip as its `Base`, and the first `acquire` is refused as `1 landings behind`; the record written with `--branch alternate` names `alternate`'s tip, and the second `acquire` admits and records the window against `alternate`.
**Mutant probe.** A remeasure that ignores `--branch` leaves the second `acquire` refused; one that resolves `--branch` differently from `acquire` records a different base; one that changes the no-flag default fails the first clause.
**Test.** `TestRemeasureBranchFlagMeasuresTheExplicitAcquireTarget` (`internal/cli/integration_base_freshness_test.go`).

### AC-ISB-006b — remeasure and acquire read the integration target from the same root (regression-guard)

**Given** a primary checkout whose git-strategy configuration names `main` as the integration target, a linked card worktree whose own configuration names `develop` (a branch that stayed at the pre-landing base), the card branched from `main` after a landing, and `CLAUDE_PROJECT_DIR` naming the linked worktree for both verbs.
**When** `moai integration remeasure -- true` runs, then `acquire --card <id>` (G-19).
**Then** the record is stored at the common root and its `Base` is `main`'s tip, and the `acquire` admits and records the window at the common root.
**Mutant probe.** A remeasure that reads the target through the environment variable as given records `develop`'s tip and the `acquire` is then refused after a successful remeasure.
**Test.** `TestRemeasureAndAcquireAgreeOnIntegrationTargetAcrossLinkedRoot` (`internal/cli/integration_base_freshness_test.go`).

### AC-ISB-006c — remeasure's git reads scrub the repo-scoping environment (regression-guard)

**Given** `GIT_DIR` and `GIT_WORK_TREE` in the process environment naming another repository, and a card worktree one commit past an absorbed landing.
**When** `RunRemeasure` runs for the card (G-20).
**Then** the stored record is keyed by the card's own HEAD tree and based on the card's own merge-base, and the gate reading that tree finds a valid record.
**Mutant probe.** Any of `gitHeadSHA`, `gitMergeBase` or `headTreeSHA` run without the scrub keys the record by the other repository's tree and the gate reports no valid record for a candidate that was just remeasured.
**Test.** `TestRemeasureKeysRecordByOwnTreeUnderInheritedGitEnv` (`internal/factory/integration_remeasure_gitenv_test.go`).

### AC-ISB-007 — K lives in configuration

**Given** the template yaml and `NewDefaultConfig`.
**When** the config tests run (G-07).
**Then** the template block holds `max_behind_landings: 0` and `enabled: false`; `NewDefaultConfig` holds the same; a section file setting `max_behind_landings: 3` loads as 3; the gate source holds no decimal literal compared against landings behind.
**Mutant probe.** A gate reading the default instead of the loaded value fails the K = 3 case.

### AC-ISB-008 — gate off: measure and advise, never refuse

**Given** the AC-ISB-001 fixture with `enabled: false`.
**When** `acquire --card <id>` runs.
**Then** it exits 0, records the window, prints one stderr line beginning `base-freshness (advisory, gate off): would refuse`, and prints no sentinel.

### AC-ISB-009 — uncertainty admits with one cause line; not-applicable admits silently

**Given** the gate enabled and the tree otherwise stale, in two groups. Uncertain (applicable call): unreadable configuration; K = -1; integration branch missing or unresolvable; a git failure injected through the runner seam. Not applicable: no `--card`; cwd on the integration branch; detached HEAD.
**When** the pure gate evaluates each (G-09) and the CLI path runs `acquire` for each in a fixture repository (G-09b).
**Then** every uncertain case exits 0, records the window, and writes exactly one stderr line naming its cause (the line count is asserted equal to 1); every not-applicable case exits 0, records the window, and writes no gate line; none prints the sentinel.
**Mutant probe.** A gate treating a git failure as stale refuses the injected case; a gate printing for not-applicable fails the silent clause; a pure-package test alone would pass a CLI that drops the line, which G-09b catches.

### AC-ISB-010 — K = 0 agrees with merge cause 2; K above 0 never relaxes it

**Given** the gate enabled and a valid record for the card tree. Case A (content-free absorb): the absorb changed no content, so the tree and record are unchanged, and the tip moved one landing past the stored `Base`. Case B: K = 3 and a candidate 2 landings behind.
**When** the gate evaluates, and `moai integration merge --card <id>` runs while the window is held.
**Then** Case A: at K = 0 the gate refuses and the merge step exits 2 (`MergeExitBaseMoved`) on the same fixture; after `moai integration remeasure` rewrites the record the gate admits and the merge proceeds. Case B: `acquire` admits and the merge exits 2 and releases the window. The existing merge-step tests still pass unchanged (G-10b).
**Mutant probe.** A gate using the live merge-base admits Case A and fails; a K that relaxes cause 2 turns the exit-2 clause red.

### AC-ISB-011 — sentinel and exit code

**Given** any refusal of AC-ISB-001, AC-ISB-004, AC-ISB-005.
**When** the process exit status and first output token are read.
**Then** the first token is `BASE_FRESHNESS_VIOLATION:`, the exit status is 16, and no code in 1-15 or 20 is returned by the gate.

### AC-ISB-012 — guidance pair and template neutrality

**Given** the tree after M4.
**When** the doc and template greps run (G-12, G-12b).
**Then** the gates companion carries the section, the template mirror is byte-identical (`diff -q` silent, exit 0), and the template yaml block has no card id, SPEC id, date, or `enabled: true` within the block.

### AC-ISB-013 — registered everywhere the config surface is registered

**Given** the tree after M1.
**When** the inventory grep, the schema constant, and the reader guard run (G-13, G-13b).
**Then** both paths are in `shipped_key_inventory.yaml`, `configCacheSchemaVersion` is at least 13, and `TestShippedConfigKeysHaveReaders` passes with its swept count above zero.

### AC-ISB-014 — no verdict read, no record written

**Given** the gate source files.
**When** the import greps and the writes-nothing test run (G-14, G-14b).
**Then** neither gate file references `homestate` or `card_evidence`; `TestBaseFreshnessGateWritesNothing` shows the record directory, the window record, and the factory database file bytes unchanged across a refusing run.
**Milestone.** G-14 names `internal/cli/integration_base_freshness.go`, which M3 creates, so AC-ISB-014 flips at M3 (before M3 the grep exits 2, files absent). **Positive control.** G-06b prints at least 1, so the greps target the right files.

### AC-ISB-015 — scope fence

**Given** the run-phase HEAD.
**When** the changed-file listing runs against the merge-base with the absorbed local `main` (G-15, G-15b, G-15c, G-15d). The plan-pin form `git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD` is withdrawn: after four absorbs it prints 204 paths in the tree outside `.moai`, almost all of them `main`'s own commits.
**Then** every path G-15 lists is in plan.md §C's allow-list (18 paths, the three audit-repair paths and `CHANGELOG.md` included), G-15d lists only the live `workflow.yaml`, the SPEC directory and `.moai/reports/t1834/`, the pathspec check over neighbour-owned paths (G-15b) prints nothing, and the control G-15c prints the gate path. The measurement is valid before the card lands only (plan.md G15).

### AC-ISB-016 — adoption is not refused by the gate (no-window regression)

**Given** the gate enabled, K = 0, a card merged earlier through `moai integration merge --card` (the integration tip is a merge commit whose second parent is the card branch's current tip, and a valid record keys the tip's tree), and a caller that does NOT hold the window.
**When** `moai factory complete` runs.
**Then** it completes through the adoption path exactly as without the gate (card state `merged-local`, exit 0, no `BASE_FRESHNESS_VIOLATION:` line); a second fixture in which the card gained a commit after the merge is not an adoption and the gate evaluates normally.
**Mutant probe.** A gate with no adoption exemption refuses this fixture (the merge commit is a first-parent landing and `record.Base != tip`), failing the first clause.

## §D — Green-path commands (run at the milestone named in §B)

```
G-01  go test ./internal/cli/ -run '^TestIntegrationAcquireRefusesStaleBase$' -v -count=1
G-02  go test ./internal/cli/ -run '^TestFactoryCompleteRefusesStaleBaseBeforeAcquire$' -v -count=1
G-03  go test ./internal/factory/ -run '^TestLandingsBehindCountsFirstParent$' -v -count=1
G-04  go test ./internal/factory/ -run '^TestBaseFreshnessKBoundary$' -v -count=1
G-05  go test ./internal/factory/ -run '^TestBaseFreshnessRequiresValidRecordForCurrentTree$' -v -count=1
G-06  grep -rn 'type RemeasureRecord struct' internal --include='*.go'
G-06b grep -c 'ValidateRemeasureRecord' internal/factory/base_freshness.go
G-06c grep -n '"merge"' internal/factory/base_freshness.go
G-07  go test ./internal/config/ -run '^(TestWorkflowBaseFreshnessGateDefaults|TestWorkflowBaseFreshnessGateLoaderRoundTrip|TestWorkflowBaseFreshnessGateTemplateParity)$' -v -count=1
G-08  go test ./internal/cli/ -run '^TestBaseFreshnessGateOffAdvisesOnly$' -v -count=1
G-09  go test ./internal/factory/ -run '^TestBaseFreshnessFailOpenCauses$' -v -count=1
G-09b go test ./internal/cli/ -run '^TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine$' -v -count=1
G-10  go test ./internal/factory/ -run '^TestBaseFreshnessK0MatchesMergeCause2$' -v -count=1
G-10b go test ./internal/factory/ -run '^(TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestMergeStepMidstepBranchTipMoveRefusesMerge)$' -v -count=1
G-10c go test ./internal/cli/ -run '^TestBaseFreshnessKDoesNotRelaxMergeStep$' -v -count=1
G-11  go test ./internal/cli/ -run '^TestBaseFreshnessRefusalSentinelAndExitCode$' -v -count=1
G-12  grep -c 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md
G-12b diff -q .claude/rules/moai/workflow/factory-dispatch-gates.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
G-13  grep -c 'workflow.base_freshness_gate' internal/config/testdata/shipped_key_inventory.yaml
G-13b go test ./internal/config/ -run '^TestShippedConfigKeysHaveReaders$' -v -count=1
G-14  grep -n -E 'homestate|card_evidence' internal/factory/base_freshness.go internal/cli/integration_base_freshness.go
G-14b go test ./internal/factory/ -run '^TestBaseFreshnessGateWritesNothing$' -v -count=1
G-15  git diff --name-only main...HEAD -- . ':!.moai'
G-15b git diff --name-only main...HEAD -- AGENTS.local.md .claude/rules/local internal/cli/session_worktree.go internal/cli/local_main_resync.go internal/cli/integration_merge.go internal/factory/integration_merge_step.go
G-15c git diff --name-only main...HEAD -- internal/factory/base_freshness.go
G-15d git diff --name-only main...HEAD -- .moai/config .moai/reports .moai/specs
G-16  go test ./internal/cli/ -run '^TestFactoryCompleteAdoptionNotRefusedByBaseFreshness$' -v -count=1
G-17  go test ./internal/factory/ -run '^TestBaseFreshnessRemedyNamesExplicitBranch$' -v -count=1
G-17b go test ./internal/cli/ -run '^(TestAcquireRefusalRemedyNamesTheExplicitBranch|TestAcquireRefusalRemedyUnchangedWithoutExplicitBranch)$' -v -count=1
G-18  go test ./internal/cli/ -run '^TestRemeasureBranchFlagMeasuresTheExplicitAcquireTarget$' -v -count=1
G-19  go test ./internal/cli/ -run '^TestRemeasureAndAcquireAgreeOnIntegrationTargetAcrossLinkedRoot$' -v -count=1
G-20  go test ./internal/factory/ -run '^TestRemeasureKeysRecordByOwnTreeUnderInheritedGitEnv$' -v -count=1
```

Expected green: every Go-test command prints `--- PASS: <name>` for each named test and `ok`; G-06 one line; G-06b at least 1; G-06c nothing (exit 1); G-12 at least 1; G-12b nothing (exit 0); G-13 prints 2; G-14 nothing (exit 1); G-15 prints exactly the 18 allow-listed paths of plan.md §C (`.claude/rules/moai/workflow/factory-dispatch-gates.md`, `CHANGELOG.md`, `internal/cli/factory_card.go`, `internal/cli/integration.go`, `internal/cli/integration_base_freshness.go`, `internal/cli/integration_base_freshness_test.go`, `internal/cli/integration_remeasure.go`, `internal/config/cache.go`, `internal/config/defaults.go`, `internal/config/testdata/shipped_key_inventory.yaml`, `internal/config/types.go`, `internal/config/workflow_base_freshness_gate_test.go`, `internal/factory/base_freshness.go`, `internal/factory/base_freshness_test.go`, `internal/factory/integration_remeasure.go`, `internal/factory/integration_remeasure_gitenv_test.go`, `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md`, `internal/template/templates/.moai/config/sections/workflow.yaml`) and no other; G-15b nothing; G-15c prints the path; G-15d prints `.moai/config/sections/workflow.yaml`, the files of `.moai/reports/t1834/`, and the files of `.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/`. G-15b no longer names `internal/factory/integration_remeasure.go`: that file moved from neighbour-owned to allow-listed with the audit repair (iter9 G1).

## §E — Evidence ledger (RED-now, re-executed at HEAD `da051947fc792072e90c53834807400e9ad8ca9a`; no `internal/` file differs from the plan pin `fb749d372e61a6b10c53091362c76b4412827427`)

Every cell below was run in this revision; stdout is verbatim. The probe-file cells of the previous revision (E-00a/b/c) are withdrawn: they did not reproduce on a committed tree.

```
E-12a cmd:    grep -rl 'func TestIntegrationAcquireRefusesStaleBase(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12b cmd:    grep -rl 'func TestFactoryCompleteRefusesStaleBaseBeforeAcquire(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12c cmd:    grep -rl 'func TestLandingsBehindCountsFirstParent(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12d cmd:    grep -rl 'func TestBaseFreshnessKBoundary(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12e cmd:    grep -rl 'func TestBaseFreshnessRequiresValidRecordForCurrentTree(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12f cmd:    grep -rl 'func TestWorkflowBaseFreshnessGateDefaults(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12g cmd:    grep -rl 'func TestBaseFreshnessGateOffAdvisesOnly(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12h cmd:    grep -rl 'func TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12i cmd:    grep -rl 'func TestBaseFreshnessK0MatchesMergeCause2(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12j cmd:    grep -rl 'func TestBaseFreshnessRefusalSentinelAndExitCode(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12k cmd:    grep -rl 'func TestBaseFreshnessGateWritesNothing(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-12l cmd:    grep -rl 'func TestFactoryCompleteAdoptionNotRefusedByBaseFreshness(' internal --include='*_test.go'
      stdout: (empty)
      exit:   1
      why:    not yet observable: the criterion's own planned test is absent. The behavioural RED is the
              run-phase RED commit that precedes the GREEN commit (verification-claim-integrity §2.3).
E-13  cmd:    go doc ./internal/factory EvaluateBaseFreshness
      stdout: doc: no symbol EvaluateBaseFreshness in package ./internal/factory
      exit:   1
      why:    the gate symbol the factory-package criteria need does not exist yet (corroborates E-12c to E-12e, E-12i, E-12k).

E-01  cmd:    grep -c 'base_freshness_gate' internal/config/types.go
      stdout: 0
      exit:   1
E-02  cmd:    grep -c 'base_freshness_gate' internal/template/templates/.moai/config/sections/workflow.yaml
      stdout: 0
      exit:   1
E-03  cmd:    grep -c 'workflow.base_freshness_gate' internal/config/testdata/shipped_key_inventory.yaml
      stdout: 0
      exit:   1
E-04  cmd:    grep -c -i 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md
      stdout: 0
      exit:   1
E-05  cmd:    grep -rl 'BASE_FRESHNESS_VIOLATION' internal --include='*.go'
      stdout: (empty)
      exit:   1
E-06  cmd:    grep -n 'configCacheSchemaVersion = ' internal/config/cache.go
      stdout: 50:const configCacheSchemaVersion = 12
      exit:   0
      why:    RED: the criterion demands at least 13.
E-07  cmd:    grep -rn 'type RemeasureRecord struct' internal --include='*.go'
      stdout: internal/factory/integration_remeasure.go:45:type RemeasureRecord struct {
      exit:   0
      why:    positive control only (the count command sees the one real type); not an AC's RED.
E-08  cmd:    diff -q .claude/rules/moai/workflow/factory-dispatch-gates.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
      stdout: (empty)
      exit:   0
      why:    baseline: the pair is byte-identical today.

E-09  cmd:    grep -c 'ValidateRemeasureRecord' internal/factory/base_freshness.go
      stdout: ugrep: warning: internal/factory/base_freshness.go: No such file or directory
      exit:   2
      why:    RED for AC-ISB-006: the green path demands a count of at least 1 (exit 0); exit 2 here.

E-10  cmd:    grep -n -E 'homestate|card_evidence' internal/factory/base_freshness.go internal/cli/integration_base_freshness.go
      stdout: ugrep: warning: internal/factory/base_freshness.go: No such file or directory
              ugrep: warning: internal/cli/integration_base_freshness.go: No such file or directory
      exit:   2
      why:    RED for AC-ISB-014: the green path is "no match, files present" (exit 1); exit 2 means the
              files do not exist, so the clean grep cannot yet be read as a pass.

E-11  cmd:    git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- internal/factory/base_freshness.go
      stdout: (empty)
      exit:   0
      why:    RED for AC-ISB-015: the control G-15c must print the gate path; today it prints nothing, so
              the fence check cannot be read (an empty control makes a clean G-15 vacuous).
```

### §E.1 — Amendment cells (2026-10-19, measured at HEAD `9fc118b44ec917f298809968f4034c26af4f0336`)

The scope-fence commands below were run in this amendment against that HEAD; stdout is verbatim. The sub-criteria cells E-14a to E-14d are not RED observations: their RED-now cannot be re-executed (the pre-repair code is history), so they record where the failing run is claimed and that this amendment did not observe it. No Go command was run in this amendment, so G-17 to G-20 (and every other Go-test command) are unobserved here; the names were checked to exist by `grep` only.

```
E-15a cmd:    git diff --name-only main...HEAD -- . ':!.moai'
      stdout: .claude/rules/moai/workflow/factory-dispatch-gates.md
              CHANGELOG.md
              internal/cli/factory_card.go
              internal/cli/integration.go
              internal/cli/integration_base_freshness.go
              internal/cli/integration_base_freshness_test.go
              internal/cli/integration_remeasure.go
              internal/config/cache.go
              internal/config/defaults.go
              internal/config/testdata/shipped_key_inventory.yaml
              internal/config/types.go
              internal/config/workflow_base_freshness_gate_test.go
              internal/factory/base_freshness.go
              internal/factory/base_freshness_test.go
              internal/factory/integration_remeasure.go
              internal/factory/integration_remeasure_gitenv_test.go
              internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
              internal/template/templates/.moai/config/sections/workflow.yaml
      exit:   0
      why:    G-15 at 9fc118b44e: 18 paths, all in plan.md §C. merge-base with local main is 0ddf5ca48bed7aaeeb68ed4fc686bcfa098a82b1.
E-15b cmd:    git diff --name-only main...HEAD -- AGENTS.local.md .claude/rules/local internal/cli/session_worktree.go internal/cli/local_main_resync.go internal/cli/integration_merge.go internal/factory/integration_merge_step.go
      stdout: (empty)
      exit:   0
      why:    G-15b at 9fc118b44e: no neighbour-owned path changed.
E-15c cmd:    git diff --name-only main...HEAD -- internal/factory/base_freshness.go
      stdout: internal/factory/base_freshness.go
      exit:   0
      why:    G-15c control: non-empty, so the empty G-15b is readable as a clean result.
E-15d cmd:    git diff --name-only main...HEAD -- .moai/config .moai/reports .moai/specs
      stdout: .moai/config/sections/workflow.yaml
              .moai/reports/t1834/card-review.md
              .moai/reports/t1834/plan-audit-iter1.md
              .moai/reports/t1834/plan-audit-iter2.md
              .moai/reports/t1834/plan-audit-iter3.md
              .moai/reports/t1834/plan-audit-iter4.md
              .moai/reports/t1834/sync-audit-iter5.md
              .moai/reports/t1834/sync-audit-iter6.md
              .moai/reports/t1834/sync-audit.md
              .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/acceptance.md
              .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/decision-index.md
              .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/plan.md
              .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/progress.md
              .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/spec.md
      exit:   0
      why:    G-15d at 9fc118b44e (before the amendment commit): only the live workflow.yaml, the SPEC directory and .moai/reports/t1834/.
E-14a cmd:    (not re-executable) the RED of TestAcquireRefusalRemedyNamesTheExplicitBranch and TestBaseFreshnessRemedyNamesExplicitBranch on the tree before commit 9fc118b44e
      stdout: (none observed in this amendment)
      exit:   (none)
      why:    regression-guard: the commit message of 9fc118b44e states that the remedy told a lane to re-run a bare remeasure; asserted there, not observed here.
E-14b cmd:    (not re-executable) the RED of TestRemeasureBranchFlagMeasuresTheExplicitAcquireTarget on the tree before commit f932d66dee
      stdout: (none observed in this amendment)
      exit:   (none)
      why:    regression-guard: the test comment records that remeasure had no --branch flag, so the explicit-target acquire stayed refused; asserted there, not observed here.
E-14c cmd:    (not re-executable) the RED of TestRemeasureAndAcquireAgreeOnIntegrationTargetAcrossLinkedRoot on the tree before commit e81bf9429e
      stdout: (none observed in this amendment)
      exit:   (none)
      why:    regression-guard: the commit message of e81bf9429e states the test failed before the fix (BASE_FRESHNESS_VIOLATION after a successful remeasure); asserted there, not observed here.
E-14d cmd:    (not re-executable) the RED of TestRemeasureKeysRecordByOwnTreeUnderInheritedGitEnv on the tree before commit 087a270e48
      stdout: (none observed in this amendment)
      exit:   (none)
      why:    regression-guard: the test comment records the record keyed by the decoy repository's tree under an inherited GIT_DIR; asserted there, not observed here.
```

## §F — Edge cases and quality gates

- Edge: a content-free absorb leaves the tree and record valid but moves the tip past the stored `Base`; the gate refuses at K = 0 until the lane re-runs `moai integration remeasure` (AC-ISB-010 Case A), exactly as merge cause 2 would.
- Edge: `max_behind_landings` set high never removes the record requirement (AC-ISB-005 at K = 2).
- Edge: a leader acquiring without `--card` is not applicable and silent (AC-ISB-009).
- Edge: an already-merged card adopted by `factory complete` is exempt (AC-ISB-016).
- Edge: a refusal of `acquire --branch <other>` names `remeasure --branch <other>` (AC-ISB-004a); a remeasure run without the flag still records against the configured target (AC-ISB-006a).
- Edge: a process environment carrying `GIT_DIR` / `GIT_WORK_TREE` for another repository does not re-key the re-measure record (AC-ISB-006c).
- Quality gates: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/factory/... ./internal/cli/... ./internal/config/...` reports no new issue; `make build` exit 0 with the embed regeneration committed; `moai spec lint` reports no error for this SPEC.

## §G — Definition of Done

All 16 criteria PASS with their commands run in the run phase and the verbatim output carried into the run-phase verdict file; every criterion shows its committed RED before its GREEN; the allow-list check (G-15) lists only the 18 allow-listed paths, with G-15b empty, its control (G-15c) non-empty and G-15d limited to the live configuration, the SPEC directory and `.moai/reports/t1834/`; the four regression-guard sub-criteria (AC-ISB-004a, AC-ISB-006a, AC-ISB-006b, AC-ISB-006c) pass at G-17 to G-20 with every named test shown `--- PASS`, as a guard and not as a release gate; the merge-step regression run (G-10b) is green with a swept count above zero; the run-phase milestone M4 edits the guidance pair in the template tree first and then the repository copy (Template-First).
