# Acceptance: SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

Document-level tree pin (binds every criterion that carries no pin of its own): `fb749d372e61a6b10c53091362c76b4412827427`. Verification layer: Given-When-Then entries; the requirement layer is spec.md §B (GEARS).

## §A — Authoring rules applied

1. **Plain commands.** Every command is one plain pipeline, one `git` per line, no `$()`, no process substitution, no shell variables.
2. **Classification.** Every criterion is **release-blocking**: each carries a RED-now cell observed on the pinned tree (§E) with command, verbatim stdout, exit code, and pin, plus a green-path cell naming the milestone that flips it. The Definition of Done gates all 16. No criterion is a regression-guard, because every RED below was re-executed.
3. **What a RED is here.** Where the artifact under test does not exist at the pin, the observed RED is the **structural RED**: a throwaway probe test naming the planned symbol was written, run, and deleted, and the verbatim compile failure is recorded (E-00a, E-00b, E-00c). A zero-test sweep is not used as a RED (`[no tests to run]`, exit 0, is a vacuous pass).
4. **Swept count.** Every Go-test command in §D names its tests; a green counts only when `-v` shows each named test `--- PASS`.

## §B — Criteria matrix

| ID | Requirement | Class | RED-now (ledger id, §E) | Green path (§D command, milestone) |
|----|-------------|-------|--------------------------|--------------------------------------|
| AC-ISB-001 | REQ-ISB-001 | release-blocking | E-00b | G-01, M3 |
| AC-ISB-002 | REQ-ISB-002 | release-blocking | E-00b | G-02, M3 |
| AC-ISB-003 | REQ-ISB-003 | release-blocking | E-00a | G-03, M2 |
| AC-ISB-004 | REQ-ISB-004 | release-blocking | E-00a | G-04, M2 |
| AC-ISB-005 | REQ-ISB-005 | release-blocking | E-00a | G-05, M2 |
| AC-ISB-006 | REQ-ISB-006 | release-blocking | E-09 | G-06, G-06b, G-06c, M2 |
| AC-ISB-007 | REQ-ISB-007 | release-blocking | E-00c, E-01, E-02 | G-07, M1 |
| AC-ISB-008 | REQ-ISB-008 | release-blocking | E-00b | G-08, M3 |
| AC-ISB-009 | REQ-ISB-009 | release-blocking | E-00a, E-00b | G-09, G-09b, M3 |
| AC-ISB-010 | REQ-ISB-010 | release-blocking | E-00a, E-00b | G-10, G-10b, G-10c, M3 |
| AC-ISB-011 | REQ-ISB-011 | release-blocking | E-00b, E-05 | G-11, M3 |
| AC-ISB-012 | REQ-ISB-012 | release-blocking | E-02, E-04, E-08 | G-12, G-12b, M1, M4 |
| AC-ISB-013 | REQ-ISB-013 | release-blocking | E-03, E-06 | G-13, G-13b, M1 |
| AC-ISB-014 | REQ-ISB-014 | release-blocking | E-00a, E-10 | G-14, G-14b, M2 |
| AC-ISB-015 | REQ-ISB-015 | release-blocking | E-11 | G-15, G-15b, G-15c, M5 |
| AC-ISB-016 | REQ-ISB-016 | release-blocking | E-00b | G-16, M3 |

Coverage declarations: AC-ISB-001 maps REQ-ISB-001; AC-ISB-002 maps REQ-ISB-002; AC-ISB-003 maps REQ-ISB-003; AC-ISB-004 maps REQ-ISB-004; AC-ISB-005 maps REQ-ISB-005; AC-ISB-006 maps REQ-ISB-006; AC-ISB-007 maps REQ-ISB-007; AC-ISB-008 maps REQ-ISB-008; AC-ISB-009 maps REQ-ISB-009; AC-ISB-010 maps REQ-ISB-010; AC-ISB-011 maps REQ-ISB-011; AC-ISB-012 maps REQ-ISB-012; AC-ISB-013 maps REQ-ISB-013; AC-ISB-014 maps REQ-ISB-014; AC-ISB-015 maps REQ-ISB-015; AC-ISB-016 maps REQ-ISB-016.

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

**Given** the gate enabled and the tree otherwise stale, in two groups. Uncertain (applicable call): unreadable configuration; K = -1; integration branch missing; a git failure injected through the runner seam. Not applicable: no `--card`; cwd on the integration branch; detached HEAD.
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
**Positive control.** G-06b prints at least 1, so the greps target the right files.

### AC-ISB-015 — scope fence

**Given** the run-phase HEAD.
**When** the changed-file listing runs against the plan pin (G-15, G-15b, G-15c).
**Then** every listed path is in plan.md §C's allow-list (the SPEC directory and `.moai/reports/t1834/` included), the pathspec check over neighbour-owned paths prints nothing, and the control G-15c prints the gate path.

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
G-15  git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD
G-15b git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- AGENTS.local.md .claude/rules/local internal/cli/session_worktree.go internal/cli/local_main_resync.go internal/cli/integration_merge.go internal/factory/integration_merge_step.go internal/factory/integration_remeasure.go
G-15c git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- internal/factory/base_freshness.go
G-16  go test ./internal/cli/ -run '^TestFactoryCompleteAdoptionNotRefusedByBaseFreshness$' -v -count=1
```

Expected green: every Go-test command prints `--- PASS: <name>` for each named test and `ok`; G-06 one line; G-06b at least 1; G-06c nothing (exit 1); G-12 at least 1; G-12b nothing (exit 0); G-13 prints 2; G-14 nothing (exit 1); G-15 only allow-listed paths; G-15b nothing; G-15c prints the path.

## §E — Evidence ledger (RED-now, tree `fb749d372e61a6b10c53091362c76b4412827427`, this authoring run)

```
E-00a cmd:    go test ./internal/factory/ -run '^TestBaseFreshnessKBoundary$' -count=1
      probe:  a throwaway internal/factory/zz_probe_test.go containing
              func TestBaseFreshnessKBoundary(t *testing.T) { _ = EvaluateBaseFreshness }
              (written, run, and deleted; git status shows no leftover)
      stdout: # github.com/modu-ai/moai-adk/internal/factory [github.com/modu-ai/moai-adk/internal/factory.test]
              internal/factory/zz_probe_test.go:5:53: undefined: EvaluateBaseFreshness
              FAIL	github.com/modu-ai/moai-adk/internal/factory [build failed]
              FAIL
      exit:   1
      why:    structural RED: the planned gate symbol does not exist. It is red for the right reason (the
              work defines the symbol) and flips only when M2 defines it and the real tests pass.

E-00b cmd:    go test ./internal/cli/ -run '^TestIntegrationAcquireRefusesStaleBase$' -count=1
      probe:  throwaway internal/cli/zz_probe_test.go referencing acquireBaseFreshnessPrecondition
      stdout: internal/cli/zz_probe_test.go:5:...: undefined: acquireBaseFreshnessPrecondition
              FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]   (same shape; one probe line recorded)
      exit:   1
      why:    structural RED for the CLI adapter symbol; AC-ISB-016's own guard behaviour is green today
              (no gate exists), so its observed RED is the missing test and adapter, and its failing
              direction is the mutant of §C (a gate without the exemption).

E-00c cmd:    go test ./internal/config/ -run '^TestWorkflowBaseFreshnessGateDefaults$' -count=1
      probe:  throwaway internal/config/zz_probe_test.go referencing NewDefaultConfig().Workflow.BaseFreshnessGate
      stdout: ... undefined (type WorkflowConfig has no field or method BaseFreshnessGate)
              FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
      exit:   1

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

## §F — Edge cases and quality gates

- Edge: a content-free absorb leaves the tree and record valid but moves the tip past the stored `Base`; the gate refuses at K = 0 until the lane re-runs `moai integration remeasure` (AC-ISB-010 Case A), exactly as merge cause 2 would.
- Edge: `max_behind_landings` set high never removes the record requirement (AC-ISB-005 at K = 2).
- Edge: a leader acquiring without `--card` is not applicable and silent (AC-ISB-009).
- Edge: an already-merged card adopted by `factory complete` is exempt (AC-ISB-016).
- Quality gates: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/factory/... ./internal/cli/... ./internal/config/...` reports no new issue; `make build` exit 0 with the embed regeneration committed; `moai spec lint` reports no error for this SPEC.

## §G — Definition of Done

All 16 criteria PASS with their commands run in the run phase and the verbatim output carried into the run-phase verdict file; every criterion shows its committed RED before its GREEN; the allow-list check (G-15) is clean with its control (G-15c) non-empty; the merge-step regression run (G-10b) is green with a swept count above zero; the sync phase moves the guidance pair into the template tree before the repository copy (Template-First).
