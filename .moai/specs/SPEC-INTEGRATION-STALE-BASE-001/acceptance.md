# Acceptance: SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

Document-level tree pin (binds every criterion that carries no pin of its own): `fb749d372e61a6b10c53091362c76b4412827427`. Verification layer: Given-When-Then entries; the requirement layer is spec.md §B (GEARS).

## §A — Authoring rules applied

1. **Plain commands.** Every command is one plain pipeline, one `git` per line, no `$()`, no process substitution, no shell variables (the worktree-session guard refuses those shapes).
2. **Two cells per criterion.** Each criterion carries a RED-now cell (an evidence-ledger entry in §E, measured on the pinned tree) and a green-path cell (which milestone flips it and what the passing output becomes). A criterion whose RED-now would be an empty sweep is labelled **Ordered**: its Go test does not exist at the pin, so the pin-tree run is vacuous (ledger E-00) and the real RED is the failing run M1 commits before M2 makes it green.
3. **Swept count.** Every Go-test command in §D selects named tests; a green counts only when `-v` shows each named test `--- PASS` (zero tests run is not a pass).

## §B — Criteria matrix

| ID | Requirement | Class | Verification (ledger id, §E) | Flips at |
|----|-------------|-------|------------------------------|----------|
| AC-ISB-001 | REQ-ISB-001 | Ordered (after M1) | E-00, G-01 | M3 |
| AC-ISB-002 | REQ-ISB-002 | Ordered (after M1) | E-00, G-02 | M3 |
| AC-ISB-003 | REQ-ISB-003 | Ordered (after M1) | E-00, G-03 | M2 |
| AC-ISB-004 | REQ-ISB-004 | Ordered (after M1) | E-00, G-04 | M2 |
| AC-ISB-005 | REQ-ISB-005 | Ordered (after M1) | E-00, G-05 | M2 |
| AC-ISB-006 | REQ-ISB-006 | Static | E-07, G-06 | M2 |
| AC-ISB-007 | REQ-ISB-007 | Static plus Ordered | E-01, E-02, G-07 | M1 |
| AC-ISB-008 | REQ-ISB-008 | Ordered (after M1) | E-00, G-08 | M2 |
| AC-ISB-009 | REQ-ISB-009 | Ordered (after M1) | E-00, G-09 | M2 |
| AC-ISB-010 | REQ-ISB-010 | Ordered (after M1) | E-00, G-10 | M3 |
| AC-ISB-011 | REQ-ISB-011 | Static plus Ordered | E-05, G-11 | M3 |
| AC-ISB-012 | REQ-ISB-012 | Static | E-04, E-02, G-12 | M1, M4 |
| AC-ISB-013 | REQ-ISB-013 | Static plus Ordered | E-03, E-06, G-13 | M1 |
| AC-ISB-014 | REQ-ISB-014 | Static plus Ordered | E-07, G-14 | M2 |
| AC-ISB-015 | REQ-ISB-015 | Static | G-15 | M5 |

Coverage declarations (each criterion maps its requirement): AC-ISB-001 maps REQ-ISB-001; AC-ISB-002 maps REQ-ISB-002; AC-ISB-003 maps REQ-ISB-003; AC-ISB-004 maps REQ-ISB-004; AC-ISB-005 maps REQ-ISB-005; AC-ISB-006 maps REQ-ISB-006; AC-ISB-007 maps REQ-ISB-007; AC-ISB-008 maps REQ-ISB-008; AC-ISB-009 maps REQ-ISB-009; AC-ISB-010 maps REQ-ISB-010; AC-ISB-011 maps REQ-ISB-011; AC-ISB-012 maps REQ-ISB-012; AC-ISB-013 maps REQ-ISB-013; AC-ISB-014 maps REQ-ISB-014; AC-ISB-015 maps REQ-ISB-015.

## §C — Scenarios (Given-When-Then)

### AC-ISB-001 — acquire refuses before touching the window

**Given** a fixture repository whose local `main` has one first-parent landing after a card branch's merge-base, a valid re-measure record keyed to the card tree, the gate enabled, K = 0, and a window record file with known bytes.
**When** `moai integration acquire --card <id> --wait` runs from the card worktree.
**Then** it exits 16, prints a line beginning `BASE_FRESHNESS_VIOLATION:`, the window record bytes are identical to before, and the queue holds no ticket.
**Mutant probe.** A gate placed after `AcquireIntegrationWindow` passes the exit-code clause and fails the byte-identity clause; a gate placed after the `--wait` branch fails the empty-queue clause.
**Green path.** M3 flips `TestIntegrationAcquireRefusesStaleBase` from FAIL to PASS (command G-01).

### AC-ISB-002 — complete refuses before acquiring

**Given** the AC-ISB-001 fixture, a merge-ready card, and a caller that does not hold the window.
**When** `moai factory complete` runs for the card.
**Then** it exits 16 with the sentinel, the window record is unchanged, the integration branch tip is unchanged, and the card state is unchanged.
**Green path.** M3 flips `TestFactoryCompleteRefusesStaleBaseBeforeAcquire` (G-02).

### AC-ISB-003 — landings are first-parent

**Given** a fixture `main` of N = 3 integration merges, each merging a side branch of 4 commits.
**When** the gate measures a card branch whose merge-base is the commit before the first merge.
**Then** landings behind is 3, not 15 and not 3 times the side-branch size.
**Mutant probe.** A count without `--first-parent` returns 15 and fails.
**Green path.** M2 flips `TestLandingsBehindCountsFirstParent` (G-03).

### AC-ISB-004 — the K boundary and the message content

**Given** landings behind equal to K, then K + 1.
**When** the gate evaluates each, with K = 0 and again with K = 2.
**Then** equal to K admits; K + 1 refuses; the refusal text contains the numeric landings behind, the numeric K, the literal key `workflow.base_freshness_gate.max_behind_landings`, the short tip, and the words `remeasure` and `retry`.
**Mutant probe.** A `>=` comparison refuses at equality and fails the first clause; a constant K fails the K = 2 run.
**Green path.** M2 flips `TestBaseFreshnessKBoundary` (G-04).

### AC-ISB-005 — the record is required whatever L is

**Given** L = 0 and four record states for the current HEAD tree: absent; present for a different (pre-absorb) tree only; present for this tree with `ExitCode` 7; present and valid.
**When** the gate evaluates each.
**Then** the first three refuse with the record message (the third names the invalid cause from `ValidateRemeasureRecord`); the fourth admits.
**Mutant probe.** A gate that checks only L admits the first three; a gate that reads the record keyed by the merge-base instead of the HEAD tree admits the second.
**Green path.** M2 flips `TestBaseFreshnessRequiresValidRecordForCurrentTree` (G-05).

### AC-ISB-006 — reuse, not a parallel mechanism

**Given** the tree after M2.
**When** the reuse greps run (G-06).
**Then** exactly one `RemeasureRecord` type exists, the gate source calls both `ReadRemeasureRecord` and `ValidateRemeasureRecord`, and the gate source contains no git `merge` invocation.
**Mutant probe.** A gate that defines its own record struct raises the type count to 2; the positive control (E-07) shows the count command sees the real type today.

### AC-ISB-007 — K lives in configuration

**Given** the template yaml and `NewDefaultConfig`.
**When** the config tests run (G-07) and the key greps run (E-01, E-02).
**Then** the template block holds `max_behind_landings: 0` and `enabled: false`; `NewDefaultConfig` holds the same values; a section file setting `max_behind_landings: 3` loads as 3; the gate source contains no decimal literal compared against landings behind.
**Mutant probe.** A gate that reads the default instead of the loaded value passes the default case and fails the K = 3 load case.

### AC-ISB-008 — gate off: measure and advise, never refuse

**Given** the AC-ISB-001 fixture with `enabled: false`.
**When** `acquire --card <id>` runs.
**Then** it exits 0, records the window, and prints one stderr line beginning `base-freshness (advisory, gate off): would refuse`; no sentinel line is printed.
**Green path.** M2/M3 flip `TestBaseFreshnessGateOffAdvisesOnly` (G-08).

### AC-ISB-009 — uncertainty admits

**Given** each cause in turn: unreadable configuration; K = -1; integration branch missing; a git failure injected through the runner seam; no `--card`; cwd on the integration branch; detached HEAD.
**When** the gate evaluates with the gate enabled and the tree otherwise stale.
**Then** every case admits (exit 0, window recorded) and prints one line naming its cause; none prints the sentinel.
**Mutant probe.** A gate that treats a git failure as stale refuses the injected-failure case and fails.
**Green path.** M2 flips `TestBaseFreshnessFailOpenCauses` (G-09), all seven subtests named in `-v`.

### AC-ISB-010 — the merge step stays exact

**Given** K = 3 and a candidate 2 landings behind with a valid record whose `Base` is the old tip.
**When** `acquire` runs, then `moai integration merge --card <id>` runs while the window is held.
**Then** `acquire` admits; the merge exits 2 (`MergeExitBaseMoved`) and releases the window. The existing merge-step tests still pass unchanged (G-10).
**Mutant probe.** A change that lets K relax cause 2 turns the exit-2 clause green-to-red.

### AC-ISB-011 — sentinel and exit code

**Given** any refusal of AC-ISB-001, AC-ISB-004, AC-ISB-005.
**When** the process exit status and first output byte run are read.
**Then** the first printed token is `BASE_FRESHNESS_VIOLATION:`, the exit status is 16, and no merge exit code in 1-15 or 20 is returned by the gate.
**Green path.** M3 flips `TestBaseFreshnessRefusalSentinelAndExitCode` (G-11); the source grep E-05 turns from no hit to one hit.

### AC-ISB-012 — guidance pair and template neutrality

**Given** the tree after M4.
**When** the doc and template greps run (G-12).
**Then** the gates companion carries the new section, its template mirror is byte-identical (`diff -q` silent, exit 0), and the template yaml block has no card id, SPEC id, date, or `enabled: true` within the block.

### AC-ISB-013 — registered everywhere the config surface is registered

**Given** the tree after M1.
**When** the inventory grep, the schema constant, and the reader guard run (G-13).
**Then** both paths are in `shipped_key_inventory.yaml`, `configCacheSchemaVersion` is at least 13, and `TestShippedConfigKeysHaveReaders` passes with its swept count above zero.

### AC-ISB-014 — no verdict read, no record written

**Given** the gate source files.
**When** the import greps and the writes-nothing test run (G-14).
**Then** neither gate file references `homestate`, `card_evidence`, or any `Write` of a record, window, or ticket; `TestBaseFreshnessGateWritesNothing` shows the record directory, the window record, and the factory database file bytes unchanged across a refusing run.
**Positive control.** `grep -c ReadRemeasureRecord` on the gate file prints at least 1, so the grep targets the right file.

### AC-ISB-015 — scope fence

**Given** the run-phase HEAD.
**When** the changed-file listing runs against the plan pin (G-15).
**Then** every listed path is in plan.md §C's allow-list, and the pathspec check over the neighbour-owned paths prints nothing; the control run over a path this work does change prints at least one line.

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
G-10  go test ./internal/cli/ -run '^TestBaseFreshnessKDoesNotRelaxMergeStep$' -v -count=1
G-10b go test ./internal/factory/ -run '^(TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestMergeStepMidstepBranchTipMoveRefusesMerge)$' -v -count=1
G-11  go test ./internal/cli/ -run '^TestBaseFreshnessRefusalSentinelAndExitCode$' -v -count=1
G-12  grep -c 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md
G-12b diff -q .claude/rules/moai/workflow/factory-dispatch-gates.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
G-13  grep -c 'workflow.base_freshness_gate' internal/config/testdata/shipped_key_inventory.yaml
G-13b go test ./internal/config/ -run '^TestShippedConfigKeysHaveReaders$' -v -count=1
G-14  grep -n 'homestate' internal/factory/base_freshness.go internal/cli/integration_base_freshness.go
G-14b go test ./internal/factory/ -run '^TestBaseFreshnessGateWritesNothing$' -v -count=1
G-15  git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD
G-15b git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- AGENTS.local.md .claude/rules/local internal/cli/session_worktree.go internal/cli/local_main_resync.go internal/cli/integration_merge.go internal/factory/integration_merge_step.go internal/factory/integration_remeasure.go
G-15c git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- internal/factory/base_freshness.go
```

Expected green outputs: G-01 to G-05, G-07 to G-11, G-13b, G-14b print `--- PASS: <name>` for every named test (swept count equals the number named) and `ok`; G-06 prints exactly one line; G-06b prints at least 1; G-06c prints nothing (exit 1); G-12 prints at least 1; G-12b prints nothing (exit 0); G-13 prints 2; G-14 prints nothing (exit 1); G-15 lists only allow-listed paths; G-15b prints nothing; G-15c prints the path (the control).

## §E — Evidence ledger (RED-now, measured at tree `fb749d372e61a6b10c53091362c76b4412827427`, this authoring run)

```
E-00  cmd:    go test ./internal/factory/ -run '^TestBaseFreshnessKBoundary$' -count=1
      stdout: ok  	github.com/modu-ai/moai-adk/internal/factory	0.304s [no tests to run]
      exit:   0
      why:    VACUOUS, not a RED. No gate test exists at the pin, so the selector sweeps nothing and exits 0.
              Every criterion labelled "Ordered" takes its real RED from the failing run M1 commits (the test
              exists and fails because the gate function does not exist yet).

E-01  cmd:    grep -c 'base_freshness_gate' internal/config/types.go
      stdout: 0
      exit:   1
      why:    RED for AC-ISB-007 and AC-ISB-013: the config struct field does not exist yet.

E-02  cmd:    grep -c 'base_freshness_gate' internal/template/templates/.moai/config/sections/workflow.yaml
      stdout: 0
      exit:   1
      why:    RED for AC-ISB-007 and AC-ISB-012: the template block does not exist yet.

E-03  cmd:    grep -c 'workflow.base_freshness_gate' internal/config/testdata/shipped_key_inventory.yaml
      stdout: 0
      exit:   1
      why:    RED for AC-ISB-013: the inventory has no entry for either key.

E-04  cmd:    grep -c -i 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md
      stdout: 0
      exit:   1
      why:    RED for AC-ISB-012: the guidance section does not exist yet.

E-05  cmd:    grep -rl 'BASE_FRESHNESS_VIOLATION' internal --include='*.go'
      stdout: (empty)
      exit:   1
      why:    RED for AC-ISB-011: the sentinel appears in no Go source yet.

E-06  cmd:    grep -n 'configCacheSchemaVersion = ' internal/config/cache.go
      stdout: 50:const configCacheSchemaVersion = 12
      exit:   0
      why:    RED for AC-ISB-013: the constant is 12, and the criterion demands at least 13.

E-07  cmd:    grep -rn 'type RemeasureRecord struct' internal --include='*.go'
      stdout: internal/factory/integration_remeasure.go:45:type RemeasureRecord struct {
      exit:   0
      why:    POSITIVE CONTROL for AC-ISB-006: the count command sees the one real type today (1 line). It is
              green at the pin by design; the criterion's RED is a mutant that adds a second type (count 2).

E-08  cmd:    diff -q .claude/rules/moai/workflow/factory-dispatch-gates.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
      stdout: (empty)
      exit:   0
      why:    BASELINE for AC-ISB-012: the pair is byte-identical today, so a silent diff after M4 means the
              pair was edited together, and a printed line means it was not.
```

## §F — Edge cases and quality gates

- Edge: a card absorbs `main` and the absorb changes no content (identical tree). The record for that tree stays valid and L is 0 after the absorb; the gate admits (AC-ISB-005 fourth state).
- Edge: `max_behind_landings` set to a large value never makes the record requirement disappear (AC-ISB-005 runs at K = 2 as well).
- Edge: a card with no `--card` flag (the leader acquiring for a batch) is `not-applicable` (AC-ISB-009).
- Quality gates: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (the gate shells out to `git` only); `golangci-lint run --timeout=2m ./internal/factory/... ./internal/cli/... ./internal/config/...` reports no new issue; `make build` exit 0 with the embed regeneration committed; `moai spec lint` reports no error for this SPEC.

## §G — Definition of Done

All 15 criteria PASS with their commands run in the run phase and the verbatim output carried into the run-phase verdict file; every `Ordered` criterion shows its committed RED before its GREEN; the allow-list check (G-15) is clean; the merge-step regression run (G-10b) is green with a swept count above zero; the sync phase moves the guidance pair into the template tree before the repository copy (Template-First).
