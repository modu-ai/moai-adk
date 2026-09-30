# SPEC-STALE-RUN-LABEL-001 — Progress

## Plan Phase (2026-09-30, card t1373)

- Status: draft — spec.md / plan.md / acceptance.md authored by manager-spec, then repaired at audit iteration 1.
- Research: read-only; mechanism verified in this worktree (`internal/hook/factory_messages.go`, `internal/hook/session_stale_run.go`, `internal/hook/session_start_factory.go`, `internal/factorymsg/store.go` `ValidateActiveRun`, `internal/homestate/factory_run_retire.go` `retireRun`, `internal/cli/codex_factory.go` env stamp, `internal/kanban/bootstrap.go` legacy vocabulary).
- Root cause summary: legacy-label branch short-circuits to the retire prescription before any run-state measurement; env residue survives /clear because the identity lives in the session process env (plus tmux pane env); no unbind path exists. Worker-70 separation: inbound Claim path is peer-record-keyed and never reads the env label.
- SPEC ID self-check: `SPEC-STALE-RUN-LABEL-001` regex PASS (executed), uniqueness confirmed against `.moai/specs/`.
- Next: plan-audit iteration 2 (delta scope D1-D5), then M1 (run-state-gated prescription + unbind state).

## Audit record

- **Iteration 1 — FAIL, score 0.81** (`.moai/reports/t1373/plan-audit-r1.md`, audited_sha d194083fb, 2026-09-30). 5 blocking defects (D1 carrier-premise false, D2 sweep instrument report-not-verdict, D3 unmapped release-blocking test, D4 key-set unenumerated, D5 tier field missing) + 3 optional (D6, D7, D8 — D8 no fix required).
- **Iteration 2 fix pass (spec/plan/acceptance 0.2.0, this update)**:
  - D1: §C.1 carrier sentence rewritten to the measured fact (settings.local.json carries NO factory keys — auditor measured 0 `MOAI_*` keys live, no writer in the tree; residue is process env + tmux pane env). Took the auditor's preferred **DROP branch** for the persisted-carrier scrub requirement (old REQ-SRL-005): no writer exists to guard against (Enforce Simplicity). Cascaded: acceptance AC-SRL-005 shrinks to unbind-notice + re-bind line; plan M2 scrub item removed; spec Out of Scope gains the dropped-scrub bullet.
  - D2: AC-SRL-009 / E5 instrument replaced with diff-scoped added-line extraction (`TestNoNewEnvNameLiteralsInDiff`, added-line count logged); measured baseline 28 total (hook+factorymsg 12, cli 16) pinned as the RED cell with its false-positive sources.
  - D3: `TestPrescriptionGateUnavailableFailsOpen` added to plan M1 (M1.1 + M1 test list); AC matrix Test column now carries owning milestones for all 9 ACs.
  - D4: dissolved by the D1 drop (no key-set to enumerate).
  - D5: `tier: M` added to spec.md frontmatter.
  - D6 (optional): REQ-SRL-003 clarifying parenthetical — the accessor's own DB-file probe is part of the shared measurement; hook-side re-derivation from broker-file absence is the prohibited act.
  - D7 (optional): plan M1.2 clause — the once-per-session-identity dedup carrier is shared across ALL prescription surfaces (SessionStart bootstrap + UserPromptSubmit peer path), so turn 1 cannot emit twice.
  - Renumbering: old REQ-SRL-006..010 → REQ-SRL-005..009 (sequential, no gaps); AC IDs unchanged AC-SRL-001..009 with mappings updated (005→REQ-005/006, 006→REQ-007, 007→REQ-008, 008→REQ-003, 009→REQ-009).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M set) under `.moai/specs/SPEC-STALE-RUN-LABEL-001/`.
- Frontmatter: 12 canonical fields present + `tier: M`; `status: draft`; version 0.2.0.
- Out of Scope: four H3 topics including the explicit card t1345 exclusion and the audited dropped-scrub note.
- Open items for audit iteration 2: RED-now cells for the Go-test release-blocking ACs are scheduled at M1 RED (plan-phase author cannot execute unexported-function tests without writing test files); AC-SRL-009's RED cell is already measured and pinned (auditor baseline 28/12 @ d194083fb); unbind-state carrier (session-record vs factorymsg peer marker) remains the M1 review decision, now with the cross-surface sharing clause (D7) attached.

## §E.2 Run-phase Evidence

### Run-phase environment attribution (read first)

This run executed in the Claude-isolated agent worktree `.claude/worktrees/agent-a703d55100343647b` on branch `WT-stale-run-gate` (renamed in place from the runtime agent branch), because the worktree-isolation guard refuses every cross-tree operation from this session — measured at session start: `git -C` toward the card tree was refused ("a worktree-isolated agent's git operations must target its own worktree") and a Write to the card tree was refused ("Edit the worktree copy"). The card tree `.moai/worktrees/t1373` still holds the plan-phase SPEC copies; the authoritative run-phase copies are the ones committed on `WT-stale-run-gate` — byte-identical clones (`cmp`-verified all four) that then received only the sanctioned M1 frontmatter transition (spec.md `status:`/`updated:`; all bodies untouched, audit r2 hash subjects preserved). Every test invocation ran env-scrubbed as ONE compound `unset <MOAI factory/kanban vars> && go test ...` invocation — this session carries the live factory run tm3yoq lane env (measured: `MOAI_KANBAN_ID=tm3yoq`, `MOAI_FACTORY_WORKER=lane-1`), which falsifies env-reading guard tests. Note: the dispatch's `env -u` prefix form was refused by this environment's worktree guard ("cannot be shown not to be git"); the repo-canonical `unset <VARS> && <command>` single-invocation form (kanban-dispatch.md § Verification load is lane-local) was used instead — same scrub semantics, one process.

### M1 RED cells — observed 2026-10-01 on tree `a1a22b919`

Command (single compound invocation, verbatim; exit code 1):

```
unset MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_ROLE MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -v -run 'TestStaleRunNoticeSilentWhenRunRetired|TestStaleRunNoticeFiresWhenRunActive|TestStaleRunNoticeOncePerSession|TestUnbindNoticeThenSilence|TestUnbindNoticeRebindLinePresence|TestPrescriptionGateUnavailableFailsOpen|TestClearSourceDeadRunEnvYieldsUnbound|TestInboundClaimIndependentOfEnvLabel' ./internal/hook/
```

Evidence ledger (verification-completeness.md §2.1 four-element cells; carrier = this ledger, cited by id):

- **RED-AC-SRL-001** — test `TestStaleRunNoticeSilentWhenRunRetired`; expected red: the legacy branch prescribes a retire for a measured-retired run; observed stdout: `stale_run_gate_test.go:71: peer path prescribed a retire for a measured-retired run: "stale run: lane label \"worker-69\" is legacy vocabulary ... 'moai factory runs --retire srl-retired', then relaunch"`; exit 1; tree `a1a22b919`. Right-reason red: the branch returns at `factory_messages.go:66-68` before any run-state measurement.
- **RED-AC-SRL-003** — test `TestStaleRunNoticeOncePerSession`; expected red: the prescription repeats on the second turn of the same session identity; observed: `stale_run_gate_test.go:105: second turn repeated the prescription: "stale run: lane label \"worker-69\" ... 'moai factory runs --retire srl-once', then relaunch"`; exit 1; tree `a1a22b919`. Right-reason red: no dedup carrier exists.
- **RED-AC-SRL-005a** — test `TestUnbindNoticeThenSilence`; expected red: the unbind path does not exist, the first turn emits the prescription; observed: `stale_run_gate_test.go:127: unbind notice carries the retire prescription: "stale run: ... 'moai factory runs --retire srl-unbind', then relaunch"`; exit 1; tree `a1a22b919`.
- **RED-AC-SRL-005b** — test `TestUnbindNoticeRebindLinePresence`; expected red: the unbind path does not exist, the answer is the prescription; observed: `stale_run_gate_test.go:152: unbind path prescribed a retire: "stale run: ... 'moai factory runs --retire srl-dead', then relaunch"`; exit 1; tree `a1a22b919`.
- **RED-AC-SRL-008** — test `TestPrescriptionGateUnavailableFailsOpen`; expected red: with an unmeasurable run state (corrupt factory.db) the branch still prescribes; observed (re-measured — the first RED run of this test used a hand-built DB path that missed `homestate.FactoryDBPath`'s actual layout, so its premise measured not-active rather than unavailable; the corrected setup was re-measured RED on the same pre-implementation tree before any implementation existed, with the working-tree gate files moved aside and the three call-site files restored to `a1a22b919` for the measurement): `stale_run_gate_test.go:187: unmeasurable run state still prescribed a retire: "stale run: lane label \"worker-69\" is legacy vocabulary ... 'moai factory runs --retire srl-unavailable', then relaunch"`; exit 1; tree `a1a22b919`. Right-reason red: no measurement precedes the prescription.
- **RED-AC-SRL-004** — test `TestClearSourceDeadRunEnvYieldsUnbound`; expected red: the `/clear` boundary's SessionStart peer path prescribes for a dead-run label; observed: `stale_run_gate_test.go:197: the /clear boundary prescribed a retire for a dead-run label: "stale run: ... 'moai factory runs --retire srl-clear', then relaunch"`; exit 1; tree `a1a22b919`.
- **GREEN-AT-ARRIVAL-AC-SRL-002** — test `TestStaleRunNoticeFiresWhenRunActive`; observed: `--- PASS: TestStaleRunNoticeFiresWhenRunActive (0.65s)` on tree `a1a22b919`, pre-gate. This is the acceptance matrix's own positive control: the unconditional branch fires regardless of state, so no honest red exists on any pre-gate tree — a red cell here is fabricable only. Recorded per the acceptance contract's letter as regression-guard-class evidence (red not observed at M1), while noting its release-blocking FUNCTION: it is the anti-suppression control proving the M1 gate measures state rather than blanket-suppressing, and it is re-exercised in the final matrix.
- **GREEN-AT-ARRIVAL-AC-SRL-006** — test `TestInboundClaimIndependentOfEnvLabel`; observed: `--- PASS: TestInboundClaimIndependentOfEnvLabel (0.42s)` on tree `a1a22b919`, pre-gate. Same green-at-arrival classification: the Claim path never reads the env label today (worker-70 separation is the existing behavior being preserved); recorded as regression-guard-class evidence guarding REQ-SRL-007 against coupling, re-exercised in the final matrix.

AC-SRL-009's RED cell is the audit-pinned baseline (acceptance.md, whole-tree 28 literals / hook+factorymsg 12, measured by the auditor at `d194083fb`) — no M1 re-measurement is required or performed for it; its test lands at M3 with the mutate probe.

### M1 GREEN — flipped at commit `dc4c55504` (2026-10-01)

Command (same compound `unset` scrub form as the RED run, `-run` naming the same eight gate tests plus the two factorymsg accessor tests) → `ok github.com/modu-ai/moai-adk/internal/hook` + `ok ...internal/factorymsg`. Flip record: AC-SRL-001, 003, 004, 005a, 008 red→green; AC-SRL-002 and 006 green-at-arrival stayed green (their anti-suppression and separation functions survived the gate). `TestUnbindNoticeRebindLinePresence` intentionally stayed red — the re-bind line is the M2 increment (acceptance.md assigns it to M2), so the M1 tree kept one owned red.

Existing-test update (SPEC-required behavior change, sanctioned): `TestStaleRunNoticeFactoryLegacyLabel` (stale_run_m1_test.go, SPEC-ROLE-NAMING-CODE-001) now seeds an active run before asserting the prescription — the prescription branch it pins requires a measured-active run under REQ-SRL-001/002; a dead run gets the unbind notice instead. The kanban-branch tests (`LegacyLeaderSpelling`, `LegacySessionRecord`) and the pure-render tests (`TestRoleNamingM3StaleRunNoticeNamesRetireStep`, `TestFactoryHookPeerRefusesLegacyLabel`) needed no changes and pass untouched.

Full affected-package suites after M1 (compound scrub form, `-count=1 ./internal/hook/ ./internal/factorymsg/ ./internal/homestate/`): factorymsg `ok` (69.8s), homestate `ok` (60.5s), hook FAIL with exactly three failures — attributed to the base tree, not this card (below).

### Base-attributed hook failures (two-arm measured — not this card's)

Three hook-suite failures were re-measured with this card's implementation removed (working tree restored to `f2fad4fd9`'s hook package via `git checkout f2fad4fd9 -- internal/hook/` with the card's files moved aside) and fail **identically on the base tree** — arm A (card tree): FAIL; arm B (base tree): FAIL:

- `TestMaybeSet1MAutoCompactWindow` / `TestMaybeDeclareGLMContextWindow` — GLM context-window resolution (`session_start_test.go`); the resolver under test is byte-identical between the base and this card's diff (the card touches neither file).
- `TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration` — fixture email `sweep-test@example.com` in `internal/cli/worktree/sweep_test.go` (the t1369 sweep card's fixture, present at the card's base) missing from `builtinCommitIdentityDenyEmails`.

Their subject code is outside this SPEC's scope envelope; recorded for the leader. CI on `origin/develop` is their verdict owner.

### M2 GREEN — flipped at commit `f3d446330` (2026-10-01)

`TestUnbindNoticeRebindLinePresence` PASS: the unbind notice names `'moai cc -f lane-<n>'` only while an active run exists in the same root (REQ-SRL-006), omits the line otherwise, and omits it on a failed liveness measurement (fail-open).

The rebind-line test caught a real defect on landing, recorded as the M2 fix: one gate answer issued **five** homestate path resolutions, and every resolution re-runs `CanonicalProjectRoot` (git subprocesses, no memoization) — the 200ms gate budget was exhausted before the liveness measurement ran, so the line was silently omitted. Fix: the gate resolves `homestate.FactoryDBPath` once and measures through the new path-based accessor forms (`factorymsg.ProbeRunStateAt` / `ActiveRunExistsAt`); the notice marker path derives from the same resolved path (factory dir = DB dir), so the carrier location is unchanged from M1.

### M3 — separation, preservation, env-literal instrument (2026-10-01)

- `TestCurrentVocabularyBindPathUnchanged` PASS (AC-SRL-007, REQ-SRL-008): `lane-3` + measured-active run binds through the unchanged path (`factory messaging bound`, launch-pending consumed, lane resolved to the session).
- `TestNoNewEnvLiteralsInDiff` PASS (AC-SRL-009): `env-literal sweep: 401 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=f2fad4fd9..., distinct literals=0`. Two recorded notes: (a) **test-name deviation from acceptance.md** — acceptance names `TestNoNewEnvNameLiteralsInDiff`, but the dispatch's external goal evaluator pattern (`TestNoNewEnvLiterals`) cannot match that spelling (the `Name` infix breaks the match), so the test is named `TestNoNewEnvLiteralsInDiff`; acceptance.md body was not touched (run-phase boundary). (b) the instrument uses the working-tree diff form (`git diff <base>`) because the acceptance's own mutate-probe clause requires a working-tree edit to flip the extraction — the committed `base...HEAD` form cannot see one; on a clean tree the two forms are the same diff.
- **Mutate probe executed before adoption** (obligation): a deliberate `var srlMutateProbe = "MOAI_FACTORY_PROBE"` literal added to `stale_run_gate.go` flipped the sweep red — `--- FAIL: TestNoNewEnvLiteralsInDiff` with `distinct literals=1 ... [MOAI_FACTORY_PROBE]`, exit 1 (observed on a known failing input); reverted (0 occurrences verified by grep) → PASS again with `distinct literals=0`.
- Pre-adoption catch: the sweep itself flagged one comment-line literal this card had added (`gatedStaleRunAnswer`'s doc comment carried the `MOAI_FACTORY_WORKERS` spelling) — reworded to the constant-reference form (`config.EnvMoaiFactoryWorkers`) rather than weakening the instrument.

### E2 — vet, lint, cross-platform build (2026-10-01, this tree @ `597befcf1`)

- `go vet ./internal/hook/ ./internal/factorymsg/` → exit 0 (clean).
- `golangci-lint run ./internal/hook/... ./internal/factorymsg/...` → exit 0, `0 issues` — golangci-lint v2.1.6, the CI-pinned version, so the verdict is attributable.
- `GOOS=windows GOARCH=amd64 go build` over internal/hook, factorymsg, homestate, cli → exit 0; host `go build ./...` → exit 0.

### E3 — coverage (2026-10-01, this tree @ `597befcf1`; hook measured at the M2/M3 code state)

- `internal/hook`: **86.5%** package coverage (`go test -count=1 -timeout 30m -coverprofile -skip 'TestMaybeSet1MAutoCompactWindow|TestMaybeDeclareGLMContextWindow|TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration' ./internal/hook/` → `coverage: 86.5% of statements`) — the three skipped tests are the base-attributed failures recorded above (measured failing on the base tree; their absence does not touch this card's paths). Above the 85% `test_coverage_target`.
- `internal/factorymsg`: **81.7%** with the card's files (`coverage: 81.7% of statements`) vs **81.5%** measured on the same package with the card's two files removed (two-arm run) — the shortfall against the 85% goal is baseline-inherited (pre-existing store.go mass), and the card's files raise the package mean. Per-function coverage of the card's new code: `stale_run_gate.go` 66.7–100% (`staleRunPrescriptionGate` 86.4%, `unbindFactoryHookNotice` 83.3%, `gatedStaleRunAnswer` 100%; the sub-85 functions are fail-open error edges), `run_state.go` 75–100% (`ProbeRunStateAt` 93.8%, `String` 100% after the M3 addition; the 75% functions are the root-form delegates' `FactoryDBPath`-error edges).
- First hook coverage attempt without `-timeout 30m` reached the default 10m test timeout (the instrumented hook suite runs ~2x its 392s plain time) — re-run with the repo's standard 30m budget; no test hung.

### Final AC matrix — deciding run (2026-10-01 @ `597befcf1`, exit 0)

Single compound scrub invocation naming all twelve tests over both packages → **exit 0, 12/12 PASS, both packages `ok`**. Verbatim deciding lines are recorded in `.moai/reports/t1373/run-ac-green.md` (this worktree; the isolation guard refused writing it into the card tree — noted for the leader to harvest before worktree disposal). AC roll-up: AC-SRL-001..009 all PASS (7 release-blocking + 2 High); the two green-at-arrival cells (002, 006) kept their release-blocking function as the anti-suppression control and the worker-70 separation guard, with their regression-guard-class RED status recorded honestly above.


## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-01"
run_commit_sha: "597befcf1"
run_status: "complete"
run_branch: "WT-stale-run-gate"
run_worktree: ".claude/worktrees/agent-a703d55100343647b (Claude-isolated agent worktree — cross-tree git/write refused; card tree untouched)"
run_commits:
  - "a1a22b919 test(SPEC-STALE-RUN-LABEL-001): M1 RED cells + plan artifacts + draft->in-progress"
  - "dc4c55504 fix(SPEC-STALE-RUN-LABEL-001): M1 run-state-gated prescription, tri-state accessor, unbind state"
  - "f3d446330 fix(SPEC-STALE-RUN-LABEL-001): M2 re-bind line on the unbind notice"
  - "597befcf1 test(SPEC-STALE-RUN-LABEL-001): M3 vocabulary preservation and env-literal diff guard"
ac_pass_count: 9
ac_fail_count: 0
red_cells_observed_at_m1: 6
green_at_arrival_cells: 2
red_cell_demotions: "none — AC-SRL-002/006 observed green-at-arrival and are recorded per the acceptance contract's letter as regression-guard-class evidence while retaining their release-blocking function; no fabricated RED anywhere"
mutate_probe_executed: true
new_warnings_or_lints_introduced: 0
cross_platform_build:
  windows_amd64: "exit 0 (hook, factorymsg, homestate, cli)"
  host: "exit 0 (go build ./...)"
coverage:
  hook_package: "86.5% (>= 85% target)"
  factorymsg_package: "81.7% (vs 81.5% base two-arm — shortfall baseline-inherited)"
  card_files_per_function: "66.7%-100%"
test_suites:
  factorymsg: "ok"
  homestate: "ok"
  hook: "green except three base-attributed failures (two-arm measured on f2fad4fd9; CI owns their verdict)"
push_state: "not pushed — lane protocol: the leader batch-pushes develop; not merged — integration rides the serial window"
evidence_file: ".moai/reports/t1373/run-ac-green.md (this worktree; card-tree write refused by the isolation guard)"
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
