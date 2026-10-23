# SPEC-CARD-AGE-ALARM-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_complete_at: 2026-10-17T20:47:40Z
plan_status: audit-ready
final_verdict: PASS (plan-audit-iter3, rule-C scoped delta, overall_score 0.93, receipt rcpt-15b53a319b1f4ef3f7603208, audited_sha 7ef9b8588b760ef0e696ee8fb303711f992319ea)
open_decisions: none
open_conditions: none

## §E.2 Run-phase Evidence

Tree: worktree `.moai/worktrees/t1838`, branch `WT-10-17-g1`, HEAD `2a1a7f62e`. Every command below ran in this run (2026-10-18, manager-develop evidence delegation) against that HEAD unless a line says otherwise. Judging toolchain: `go` as installed via PATH (direct `go test` / `go build` / `go vet` invocations); no moai binary was used as a measurement tool.

### Commits (E6)

| SHA | Role |
|---|---|
| `3b4f6ee31` | M1 RED — card_age_alarm config tests |
| `d5f2b9b56` | M1 GREEN — card_age_alarm config keys |
| `019a3ab03` | M2 RED — age measurement tests |
| `54b08b753` | M2 GREEN — read-only age measurement |
| `8670542c6` | M3 RED — status wiring tests |
| `2a1a7f62e` | M3 GREEN — status wiring and display |
| this evidence commit | §E.2/§E.3 recording — self-referential; resolves to `git rev-parse --short HEAD` on `WT-10-17-g1` immediately after it lands (D3 placeholder philosophy, single-commit close, no backfill) |

Run-phase file set (`git diff --name-only 3b4f6ee31^..HEAD`, 12 files): `.moai/specs/SPEC-CARD-AGE-ALARM-001/spec.md`, `internal/cli/factory_age_alarm.go`, `internal/cli/factory_age_alarm_test.go`, `internal/cli/factory_card.go`, `internal/config/defaults.go`, `internal/config/loader_card_age_alarm.go`, `internal/config/loader_card_age_alarm_test.go`, `internal/config/testdata/shipped_key_inventory.yaml`, `internal/config/types.go`, `internal/homestate/card_age.go`, `internal/homestate/card_age_test.go`, `internal/template/templates/.moai/config/sections/workflow.yaml` (9 Go files among them). This evidence commit additionally carries the lane's §F Phase 4 Mode Selection record, which was already present (uncommitted) in the working tree at delegation start — same file, coherent §F content, committed here rather than left dirty.

No push — lane worktree; the factory leader batch-pushes card branches.

### E8 — RED disposition (honesty record)

The lane session that landed M1-M3 died before recording the verbatim RED outputs; no historical RED stdout exists and none is fabricated. Falsifiability is restored two ways:

1. Commit-graph witness: each milestone's RED commit precedes its GREEN commit — `3b4f6ee31` → `d5f2b9b56`, `019a3ab03` → `54b08b753`, `8670542c6` → `2a1a7f62e`, in `git log --oneline` order on this branch. Each pair is two separate commits, so the ordering claim is witnessed by the graph, not by prose.
2. Live mutant probes (both mutants are the ones acceptance.md names; protocol: apply mutant → observe the named test FAIL → `git restore` the file → re-observe GREEN; all on HEAD `2a1a7f62e`):
   - Heartbeat-conflation mutant — `internal/homestate/card_age.go`, `out.Activity.At` admitted to the no-progress basis in `ComputeCardAges`. `go test -count=1 -run 'TestCardAgeSignals' ./internal/homestate` →
     ```
     --- FAIL: TestCardAgeSignals (4.55s)
         card_age_test.go:274: no-progress after renewal = {Known:true At:2026-09-26 18:00:00 +0000 UTC Age:1h0m0s}, want still 9h25m0s — activity is not progress
     FAIL
     FAIL	github.com/modu-ai/moai-adk/internal/homestate	4.976s
     ```
     exit=1. Restored (`git status --short` back to only progress.md); the full 7-test homestate batch re-observed `ok github.com/modu-ai/moai-adk/internal/homestate 22.269s`, exit=0.
   - Append-event mutant — `internal/cli/factory_age_alarm.go`, one `INSERT INTO events(run_id,kind,payload_json,created_at)` inside `computeFactoryAgeAlarms` after the events read. `go test -count=1 -run 'TestAgeAlarmNoWrites' ./internal/cli` →
     ```
     --- FAIL: TestAgeAlarmNoWrites (2.71s)
         factory_age_alarm_test.go:241: the alarm computation wrote:
             before below:1:2026-09-26T08:30:00Z gate-alert:1:2026-09-24T00:00:00Z gate-warn:1:2026-09-25T21:00:00Z stall-warn:1:2026-09-26T05:00:00Z events=0
             after  below:1:2026-09-26T08:30:00Z gate-alert:1:2026-09-24T00:00:00Z gate-warn:1:2026-09-25T21:00:00Z stall-warn:1:2026-09-26T05:00:00Z events=2
     FAIL
     FAIL	github.com/modu-ai/moai-adk/internal/cli	4.199s
     ```
     exit=1 (events 0 → 2: one appended row per status invocation). Restored; the full cli batch re-observed `ok github.com/modu-ai/moai-adk/internal/cli 14.516s`, exit=0.

### GREEN, per AC (E1) — this run, HEAD `2a1a7f62e`

Command (config): `go test -v -count=1 -run 'TestLoadCardAgeAlarm|TestCardAgeAlarmBounds' ./internal/config` → `--- PASS: TestLoadCardAgeAlarm (0.02s)` with 4 passing subtests (no-file / no-block / enabled-only / all-keys), `--- PASS: TestCardAgeAlarmBounds (0.00s)` with 10 passing subtests, `ok github.com/modu-ai/moai-adk/internal/config 0.352s`, exit=0.

Command (homestate): `go test -v -count=1 -run 'TestGateWaitAge|TestCardAgeSignals|TestNoProgressAge|TestNoProgressAgeCreationAnchor|TestAgeAnchorsFutureTimestamps|TestStageCheckpointSameStageReEntry|TestAgeAlarmCoverage' ./internal/homestate` →
```
--- PASS: TestGateWaitAge (6.36s)
    --- PASS: TestGateWaitAge/kickoff-entry-event-not-updated-at (1.01s)
    --- PASS: TestGateWaitAge/mid-merge-expiry-anchors-on-lease-expired-event (1.62s)
    --- PASS: TestGateWaitAge/no-events-row-falls-back-to-updated-at (1.65s)
    --- PASS: TestGateWaitAge/needs-decision-and-blocked-are-gate-states (2.08s)
--- PASS: TestCardAgeSignals (1.74s)
--- PASS: TestNoProgressAge (1.27s)
--- PASS: TestNoProgressAgeCreationAnchor (2.86s)
--- PASS: TestAgeAnchorsFutureTimestamps (3.82s)
--- PASS: TestStageCheckpointSameStageReEntry (1.60s)
--- PASS: TestAgeAlarmCoverage (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/homestate	18.099s
```
exit=0 (TestNoProgressAgeCreationAnchor carried 3 passing subtests, TestAgeAnchorsFutureTimestamps 4).

Command (cli): `go test -v -count=1 -run 'TestFactoryStatusAgeAlarm|TestFactoryStatusDisabledUnchanged|TestAgeAlarmNoWrites' ./internal/cli` →
```
--- PASS: TestFactoryStatusDisabledUnchanged (1.12s)
--- PASS: TestFactoryStatusAgeAlarmText (1.08s)
--- PASS: TestFactoryStatusAgeAlarmJSON (1.25s)
--- PASS: TestFactoryStatusAgeAlarmRefusedKey (0.82s)
--- PASS: TestAgeAlarmNoWrites (2.19s)
--- PASS: TestFactoryStatusAgeAlarmFutureCommit (1.63s)
ok  	github.com/modu-ai/moai-adk/internal/cli	9.457s
```
exit=0. The `TestFactoryStatusAgeAlarm` selector matches the Text/JSON/RefusedKey/FutureCommit variants by unanchored regex, so this batch swept 6 named tests, not 3.

Sweep-count proof (`go test <pkg> -list '<same pattern>'`; none printed `[no tests to run]`, each listing ended `ok <pkg>`): config 2 names (TestLoadCardAgeAlarm, TestCardAgeAlarmBounds); homestate 7 names (the seven above); cli 6 names (the six above).

AC flips of the authoring greps (acceptance.md RED cells observed 0/empty on `7faf75fdf`):
- `grep -c "card_age_alarm" internal/config/defaults.go` → `2`, exit=0 (AC-CAA-001 flip).
- `grep -rln "CardAgeAlarm" internal/config/` → 5 files (`types.go`, `defaults.go`, `loader_card_age_alarm.go`, `loader_card_age_alarm_test.go`, `testdata/shipped_key_inventory.yaml`), exit=0 (AC-CAA-002 flip).
- `grep -c "card_age_alarm" internal/template/templates/.moai/config/sections/workflow.yaml` → `2`, exit=0 (AC-CAA-002 flip).

| AC | Status | Evidence (this run, HEAD `2a1a7f62e`) |
|---|---|---|
| AC-CAA-001 | PASS | `TestLoadCardAgeAlarm` PASS (4 subtests); defaults.go grep = 2 |
| AC-CAA-002 | PASS | `TestCardAgeAlarmBounds` PASS (10 subtests); both greps ≥ 1; degrade-not-refuse on the status surface = `TestFactoryStatusAgeAlarmRefusedKey` PASS (exactly 1 warning line, card rows survive) |
| AC-CAA-003 | PASS | `TestGateWaitAge` PASS (4 subtests incl. the updated_at fallback) |
| AC-CAA-004 | PASS | `TestCardAgeSignals`, `TestNoProgressAge`, `TestNoProgressAgeCreationAnchor`, `TestAgeAnchorsFutureTimestamps`, `TestStageCheckpointSameStageReEntry`, `TestFactoryStatusAgeAlarmFutureCommit` all PASS; the heartbeat-conflation mutant probe FAILs as acceptance.md requires |
| AC-CAA-005 | PASS | `TestFactoryStatusAgeAlarmJSON` PASS |
| AC-CAA-006 | PASS | `TestFactoryStatusAgeAlarmText` PASS |
| AC-CAA-007 | PASS (regression guard) | `TestFactoryStatusDisabledUnchanged` PASS |
| AC-CAA-008 | PASS (regression guard) | `TestAgeAlarmNoWrites` PASS; the append-event mutant probe FAILs as acceptance.md requires |
| AC-CAA-009 | PASS | `TestAgeAlarmCoverage` PASS |

### Quality batch (E2, E3, E4, E5)

- `go vet ./internal/config ./internal/homestate ./internal/cli` → no output, exit=0.
- `GOOS=windows GOARCH=amd64 go build ./internal/config ./internal/homestate ./internal/cli` → no output, exit=0 (scoped 3-package form; the run had launched and completed before the slot-release directive arrived — see Gaps).
- E4 boundary: `grep -rn 'AskUserQuestion' internal/homestate internal/cli/factory_age_alarm.go internal/cli/factory_age_alarm_test.go internal/cli/factory_card.go internal/config/defaults.go` → no output, exit=1 (0 matches across the whole homestate package and this SPEC's cli/config files). For the record: the delegation's whole-package form (`grep -rn 'AskUserQuestion' internal/config internal/homestate internal/cli --include='*.go' | grep -v _test.go`) prints ~80 pre-existing comment/doc mentions, all in files outside this SPEC's 12-file diff (harness.go, mcp_*.go, agentlint, ...) — 0 matches is not obtainable from that form on this tree, so the scoped probe above is the E4 verdict for this card.
- E3 coverage (scoped-coverage discipline — `-run`-limited `-cover`; package-wide suites deferred to CI under the load rule): config `ok ... coverage: 1.5% of statements`; homestate (7-name pattern) `coverage: 20.1% of statements`; cli (3-name pattern) `coverage: 4.8% of statements` (a second scoped run for the coverprofile measured 4.9% — run-to-run variance in the package-wide denominator). Per-file on the three new files (same scoped runs, `-coverprofile` + `go tool cover -func`): `internal/config/loader_card_age_alarm.go` — LoadCardAgeAlarm 80.0%, ResolveCardAgeAlarm 100.0%, parseAlarmBand 100.0%; `internal/homestate/card_age.go` — ComputeCardAges 92.9%, LoadCardEvents 76.2%, creationEvent 75.0%, all other functions 100.0%; `internal/cli/factory_age_alarm.go` — computeFactoryAgeAlarms 91.7%, bandAgeAlarm 100.0%, ageAlarmSignal 100.0%, writeFactoryAgeAlarmText 100.0%, factoryAgeAlarmCommitAt 77.8%.
- E5 lint: `golangci-lint run --timeout=5m ./internal/config/... ./internal/homestate/... ./internal/cli/...` → `Error: parallel golangci-lint is running`, exit=3 (another session's golangci-lint held the lock). The go-test-heavy slot was then released by the lane per the leader directive (t1858 landing remeasure priority, 2026-10-18T01:56Z) and was not re-acquired, so no retry was made. DEFERRED TO CI; new-vs-baseline lint classification for the three packages is unobserved in this run.
- `gofmt -l` on the 9 run-phase Go files → empty list, exit=0.

### Gaps (not observed)

- Historical verbatim RED outputs of M1-M3: lost (the lane session died before recording them). Falsifiability restored via the commit-graph witness and the two live mutant probes above; the loss itself remains the Gap.
- golangci-lint on the three packages: not observed (concurrent-lock exit 3, then the slot was released for the t1858 landing remeasure — leader directive 2026-10-18T01:56Z). Deferred to CI.
- Package-wide suites, `-race`, full-coverage runs, and repo-wide `go build ./...` / `GOOS=windows go build ./...`: deferred to CI under the load rule (scoped `-run` tests and the scoped 3-package windows build were run instead).
- Windows EXECUTION of the tests: unobserved (cross build only).
- CI on the integration branch has not run for this tree: the repository-wide test verdict is PENDING at the time of this record and belongs to that CI run.
- The evidence-commit SHA cannot appear inside the commit that creates it; it is recorded in §E.3 as "this evidence commit" and resolves to the branch HEAD after landing.

### Residual risk

- The M1 GREEN commit `d5f2b9b56` carried an in-flight test repair (the template-parity case parses short-form durations — a test adjusted on the GREEN commit, disclosed in that commit's body). The M1 RED→GREEN pair therefore does not witness the final test text's red state; that criterion is carried by the authoring-time RED cell on `7faf75fdf` (grep = 0) and this run's live GREEN.
- The M1-M3 commits carry no `Authored-By-Agent:` trailer, so the ownership lint reads the `draft → in-progress` transition as `OwnershipTransitionUnmeasured` (Info severity). History is not rewritten; this evidence commit carries the trailer.
- Scoped coverage percentages use package-wide denominators under `-run` filters; the per-file numbers on the three new files are the load-rule-sanctioned substitute, and CI owns the package-wide figure.
- Load-rule deferrals (lint, package-wide suites / race / full coverage) mean a NEW lint finding or a suite-level regression outside the swept tests would first surface in CI, not in this record.

### Sync-audit iter1 repair (F1-F5)

Context: the manager-develop session that authored the F1-F5 fixes died on a rate limit (HTTP 429) after committing the RED regression tests (`169f6abc7`) but before committing the fixes or writing evidence. The fixes survived as uncommitted working-tree changes (4 files, +111/−16). A successor session (this record's author) verified each diff against the verdict's fix prescriptions (`.moai/reports/t1838/sync-audit-iter1.md`), re-ran the scoped verification GREEN, and committed the fixes; the RED→GREEN ordering is witnessed by the commit graph (`169f6abc7` RED, fixes in the following commit — verification-claim-integrity §2.3).

RED attribution: this repair session did NOT observe the historical RED run live. The RED state is attributed via the commit graph — `169f6abc7` ("fix(SPEC-CARD-AGE-ALARM-001): sync-audit F1-F5 RED regression tests") precedes the fix commit, and that commit's own body records the live RED observations: F1 commit=unknown under inherited GIT_DIR, F2 marker file created by the foreign gpg helper (trap armed in both repos so the marker leg also fails the env-scrub-without-flag mutant), F3 LoadFactoryLease reverted to 15m0s/5m0s with no key-naming warning. F4/F5 carry no RED state by design — they are pins of already-correct behavior per the verdict's own wording (169f6abc7's message).

- **F1 (BLOCKING, env isolation)** — fix: `factoryAgeAlarmCommitAt` sets `cmd.Env = append(gitenv.Env(), "GIT_TERMINAL_PROMPT=0")` (internal/cli/factory_age_alarm.go:154), the house pattern of internal/core/git/manager.go:331-333; an inherited GIT_DIR/GIT_WORK_TREE can no longer outrank `-C`. GREEN (this run, tree `169f6abc7` + working fixes): `go test -v -count=1 -run 'TestFactoryStatusAgeAlarm|TestFactoryStatusDisabledUnchanged|TestAgeAlarmNoWrites|TestFactoryAgeAlarmGitEnvIsolation' ./internal/cli` → `--- PASS: TestFactoryAgeAlarmGitEnvIsolation (2.95s)` among 8 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 22.364s`.
- **F2 (helper-free read)** — fix: the invocation is `git -C <path> -c log.showSignature=false log -1 --format=%cI` (internal/cli/factory_age_alarm.go:150); no repo-configured gpg helper executes during the status read. GREEN: same cli run — the `TestFactoryAgeAlarmGitEnvIsolation` marker leg (trap armed in both target and foreign repos, premise-controlled: markers proven to fire during setup) passed; no `the gpg helper ... executed` failure.
- **F3 (alarm-block decode isolation)** — fix: `LoadCardAgeAlarm` decodes only the `card_age_alarm` block via `cardAgeAlarmFile` (internal/config/loader_card_age_alarm.go:29,54); `LoadFactoryLease` decodes only the `factory_lease` block via `factoryLeaseFile` (internal/config/loader_factory_lease.go:12,39) — together, an alarm-key YAML type error degrades the alarm alone (compiled defaults, enabled=false, exactly one key-naming warning) and the lease keys keep their written values. GREEN: `go test -v -count=1 -run 'TestLoadCardAgeAlarm|TestCardAgeAlarmBounds|TestCardAgeAlarmTypeErrorIsolated' ./internal/config` → `--- PASS: TestCardAgeAlarmTypeErrorIsolated (0.01s)` among all PASS, `ok github.com/modu-ai/moai-adk/internal/config 0.345s`; `TestLoadCardAgeAlarm` re-observed separately → 4 subtests PASS, `ok github.com/modu-ai/moai-adk/internal/config 0.274s`.
- **F4 (JSON refused-key pin)** — fix: the JSON leg inside `TestFactoryStatusAgeAlarmRefusedKey` asserts `age_alarm_warning` names `workflow.card_age_alarm.gate_warn_after` and no `age_alarms` key renders (internal/cli/factory_age_alarm_test.go:221-241). GREEN: same cli run — `--- PASS: TestFactoryStatusAgeAlarmRefusedKey (2.37s)`.
- **F5 (warn-boundary pin)** — fix: `TestFactoryStatusAgeAlarmWarnBoundary` pins age == warn_after → band=warn row and one-second-below → no row (internal/cli/factory_age_alarm_test.go:246-275), matching the `sig.Age < warn` admission in `bandAgeAlarm`. GREEN: same cli run — `--- PASS: TestFactoryStatusAgeAlarmWarnBoundary (1.52s)`.

Hygiene on the repaired tree: `go vet ./internal/config ./internal/cli` → exit 0; `gofmt -l` on the four touched files → empty, exit 0; homestate confirmation (unchanged by this repair): `go test -v -count=1 -run 'TestGateWaitAge|TestCardAgeSignals|TestNoProgressAge|TestNoProgressAgeCreationAnchor|TestAgeAnchorsFutureTimestamps|TestStageCheckpointSameStageReEntry|TestAgeAlarmCoverage' ./internal/homestate` → 7 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/homestate 21.875s`.

Gaps: the historical RED run's verbatim output was not re-executed by this session (attributed via the commit graph above). The lint / package-wide / `-race` / full-coverage / Windows-execution / integration-CI deferrals of the base §E.2 carry over unchanged — CI on the integration branch remains the owner of the repository-wide verdict, PENDING at this record's time.

### Sync-audit iter2 repair (N1-N2)

Context: the iter2 required gate (codex) returned two adjacent-surface residuals on the F1-F5-repaired tree (`.moai/reports/t1838/sync-audit-iter2.md`); this repair scopes to exactly that delta. Both RED states were observed LIVE this run (unlike iter1's lost historical REDs), and the RED→GREEN ordering is witnessed by the commit graph: RED `3f246f054` precedes the fix commit (verification-claim-integrity §2.3).

- **N1 (GIT_NO_LAZY_FETCH)** — RED (live, this run, RED test committed on tree `b612304bf5`): `go test -count=1 -run 'TestFactoryAgeAlarmGitEnvIsolation' -v ./internal/cli` → `--- FAIL: TestFactoryAgeAlarmGitEnvIsolation (3.42s)` with `factory_age_alarm_test.go:623: the ext helper of the promisor worktree executed during the commit read (N1: git lazily fetched through remote.origin.url): .../lazy-fetch-ran exists`. The leg seeds a t.TempDir promisor repo: one dated commit, the object store moved aside (the cheap honest missing-head-object mechanism — no network, deterministic), then `remote.origin.url=ext::<helper>` + `protocol.ext.allow=always` + `extensions.partialClone=origin` in the repo config; the helper appends to a marker file and exits 1. Premise control per the F2 pattern: the helper is executed once during setup and the marker proven before its absence is read as isolation (the helper exits 1 by design, so the marker — not the exit code — is the premise). Pin seam: the FULL integration leg, not the minimal env-builder seam the delegation prompt offered as fallback — the full setup measured deterministic and discriminating on this machine's git (2.54.0 Apple; probed with plain `git log` on the same repo shape before writing the test: mutant = 3 ext invocations, exit 128, marker written; pinned = immediate `bad object HEAD`, no helper run), so the pin proves behavior rather than construction. Fix: `cmd.Env = append(gitenv.Env(), "GIT_TERMINAL_PROMPT=0", "GIT_NO_LAZY_FETCH=1")` (internal/cli/factory_age_alarm.go:162). GREEN (this run): same command → `--- PASS: TestFactoryAgeAlarmGitEnvIsolation (3.10s)` among 8 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 20.510s`.
- **N2 (isolated decode for the other workflow consumers)** — RED (live, this run): `go test -count=1 -run 'TestQuotaSlotSurviveAlarmTypeError' -v ./internal/config` → `--- FAIL: TestQuotaSlotSurviveAlarmTypeError (0.00s)` with the verdict's repro end-state verbatim: `LoadQuotaGate = {Enabled:false FiveHourHoldPct:90 SevenDayHoldPct:95 ReleaseMarginPct:5 MaxAge:30m0s}`, `LoadQuotaScanBound = 128, want the written 17`, `LoadSlotLeaseDefaultMaxDuration = "30m", want the written 47m`. Fix: `quotaGateFile` (decodes only the quota_gate block) now backs LoadQuotaGate / LoadQuotaScanBound, seeded with `NewDefaultWorkflowConfig().QuotaGate` so absent keys keep their defaults (internal/config/loader_quota_gate.go:73,82,102); `slotLeaseFile` (decodes only the slot_lease block) backs LoadSlotLeaseDefaultMaxDuration (internal/config/loader_slot_lease.go:22,32) — the cardAgeAlarmFile/factoryLeaseFile pattern. Own-block decode errors keep each loader's existing silent-default degrade semantics; other blocks' errors no longer reach them. GREEN (this run): `--- PASS: TestQuotaSlotSurviveAlarmTypeError (0.00s)` within the scoped batch `go test -count=1 -run 'TestLoadCardAgeAlarm|TestCardAgeAlarmBounds|TestCardAgeAlarmTypeErrorIsolated|TestQuotaSlotSurviveAlarmTypeError|TestQAS_AC007_ConfigDefaultsMirrorTemplate|TestQWR_AC013_MaxScanDirsConfigKey|TestLoadSlotLeaseDefaultMaxDuration' -v ./internal/config` → exit 0, 7 top-level `--- PASS`, 0 `--- FAIL`, `ok github.com/modu-ai/moai-adk/internal/config 0.414s`; `go test -count=1 -list 'TestQuotaSlotSurviveAlarmTypeError' ./internal/config` lists the name (swept set non-empty).

Hygiene on the repaired tree: `go vet ./internal/config ./internal/homestate ./internal/cli` → exit 0; `gofmt -l` on the five touched files → empty, exit 0; homestate confirmation (unchanged by this repair): the 7-name scoped run → 7 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/homestate 26.938s`.

Gaps: the lint / package-wide / `-race` / full-coverage / Windows-execution / integration-CI deferrals carry over unchanged from the base §E.2 — CI on the integration branch remains the owner of the repository-wide verdict, PENDING at this record's time. The lazy-fetch leg's discrimination was measured on this machine's git (2.54.0 Apple); a deployment whose git ignores repo-local `protocol.ext.allow` does not weaken the fix — the env pin, not the repo config, is what closes the vector (the pin fails the read before any transport is selected).

### Sync-audit iter3 repair (N3)

Context: the iter3 required gate (codex) returned one P1 on the N1 repair delta itself — the version bracket of N1's own pin (`.moai/reports/t1838/sync-audit-iter3.md` § New finding): `GIT_NO_LAZY_FETCH` is first honored in git 2.46.0 (source-verified across v2.43.0..master: `promisor-remote.c` carries no env check before 2.46.0), so on older git the `ext::` transport armed by `remote.origin.url` + `protocol.ext.allow=always` stays reachable. The auditor-prescribed mechanism for the fix: `is_transport_allowed` consults the `GIT_ALLOW_PROTOCOL` env allow-list FIRST — before any per-protocol repo config (`transport.c` @ v2.43.0:1070-1076, master:1095-1109) — so an EMPTY env value (an empty allow-list) blocks every transport including `ext::` regardless of the attacker's repo config, on every supported git version. This repair scopes to exactly that delta: one env token + one pin test. The RED→GREEN ordering is witnessed by the commit graph: RED `656975cbc` precedes the fix commit (verification-claim-integrity §2.3).

- **N3 (version-independent transport block)** — RED observed LIVE this run. Seam: the **env-map seam**, not the promisor marker leg — the auditor's explicit instruction, because on git ≥ 2.46.0 the lazy-fetch pin already blocks the marker leg's vector, leaving the new pin unobservable there. The env construction was extracted (behavior-neutral, no new token in the RED commit) into the package-level helper `ageAlarmChildEnv()` (internal/cli/factory_age_alarm.go, called from `factoryAgeAlarmCommitAt`), and `TestFactoryAgeAlarmChildEnvProtocolPin` (internal/cli/factory_age_alarm_test.go) asserts the helper's constructed env as a map: `GIT_ALLOW_PROTOCOL` present with an EMPTY value, plus regression assertions for the two earlier pins (`GIT_NO_LAZY_FETCH=1`, `GIT_TERMINAL_PROMPT=0`). RED (live, this run, on tree `656975cbc`): `go test -count=1 -run 'TestFactoryAgeAlarmChildEnvProtocolPin' -v ./internal/cli` → `--- FAIL: TestFactoryAgeAlarmChildEnvProtocolPin (0.00s)` with `factory_age_alarm_test.go:668: GIT_ALLOW_PROTOCOL absent from the child env's git keys [GIT_ASKPASS GIT_CONFIG_COUNT GIT_CONFIG_GLOBAL GIT_CONFIG_KEY_0 GIT_CONFIG_KEY_1 GIT_CONFIG_VALUE_0 GIT_CONFIG_VALUE_1 GIT_EDITOR GIT_NO_LAZY_FETCH GIT_TERMINAL_PROMPT]: the version-independent transport block (N3) is missing — on git < 2.46.0 the ext:: transport of remote.origin.url stays reachable`. Fix: `ageAlarmChildEnv` returns `append(gitenv.Env(), "GIT_TERMINAL_PROMPT=0", "GIT_NO_LAZY_FETCH=1", "GIT_ALLOW_PROTOCOL=")` (internal/cli/factory_age_alarm.go), with both doc comments updated to name the two pins and their coverage (GIT_NO_LAZY_FETCH = lazy-fetch guard, git ≥ 2.46.0 only; GIT_ALLOW_PROTOCOL = the version-independent transport block, env consulted before repo config). GREEN (this run): same command → `--- PASS: TestFactoryAgeAlarmChildEnvProtocolPin (0.00s)`, `ok github.com/modu-ai/moai-adk/internal/cli 1.771s`.

Hygiene on the repaired tree: `go test -v -count=1 -run 'TestFactoryStatusAgeAlarm|TestFactoryStatusDisabledUnchanged|TestAgeAlarmNoWrites|TestFactoryAgeAlarmGitEnvIsolation|TestFactoryAgeAlarmChildEnvProtocolPin' ./internal/cli` → exit 0, 9 top-level `--- PASS`, 0 `--- FAIL`, `ok github.com/modu-ai/moai-adk/internal/cli 11.726s`; `go test -count=1 -list 'TestFactoryAgeAlarmChildEnvProtocolPin' ./internal/cli` lists the name (swept set non-empty); homestate confirmation (unchanged by this repair): the 7-name scoped run → 7 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/homestate 15.568s`; `go vet ./internal/cli` → exit 0; `gofmt -l` on the two touched Go files → empty.

Gaps: the lint / package-wide / `-race` / full-coverage / Windows-execution / integration-CI deferrals carry over unchanged from the base §E.2 — CI on the integration branch remains the owner of the repository-wide verdict, PENDING at this record's time. The version bracket itself (guard absent < 2.46.0, env-before-config precedence) is accepted on the auditor's tagged-source verification (iter3 verdict), not re-executed here — no pre-2.46.0 git binary exists on this machine to observe; the env-map pin is version-independent by construction (an env allow-list is consulted before repo config wherever `is_transport_allowed` exists). The delegation prompt's optional reinforcement (one env assertion inside the promisor marker leg) was skipped: that leg drives the real `factory status` binary through `runFactory` and its assertions observe output lines and marker files only — the child env of the internal helper is not exposed on that surface, so the assertion would have required expanding the leg's seam rather than extending it.

### Card-review repair (R1-R2)

Context: the pre-landing advisory `codex_review` (scope=card, base 7faf75fdf) returned two findings; the lane mechanically confirmed R1 by running the CI test and observing FAIL. This repair scopes to exactly that delta and appends to the §E.2 repair history after iter3. RED→GREEN ordering is witnessed by the commit graph: RED `595a9410b` (the R2 rejection test alone, tree otherwise pre-fix) precedes the fix commit `9405da696` (verification-claim-integrity §2.3).

- **R1 (template internal-identifier leak — CONFIRMED CI blocker)** — mechanism: the shipped template's `card_age_alarm` comment cited the internal SPEC identifier, which `TestTemplateNoInternalContentLeak` class C1 (general `SPEC-<DOMAIN>-NNN`) forbids anywhere in the distribution tree. Fix location: `internal/template/templates/.moai/config/sections/workflow.yaml:29-30` — the parenthetical `(SPEC-CARD-AGE-ALARM-001)` removed; the functional description (display-only age alarms on gate-waiting and stalled factory cards, unmeasured bands, ships OFF) is unchanged, and `TestCardAgeAlarmTemplateCarriesBlock`'s "unmeasured" requirement still holds. RED re-observed live this run on tree `bb7ac820a` (the lane's own FAIL preceded it with the same match): `go test -count=1 -run 'TestTemplateNoInternalContentLeak' ./internal/template` → `--- FAIL: TestTemplateNoInternalContentLeak (5.54s)` with `internal_content_leak_test.go:1653: [1] templates/.moai/config/sections/workflow.yaml | class=C1-spec-id-prefix | match=SPEC-CARD-AGE-ALARM-001`. GREEN (this run, HEAD `9405da696`): same command → `ok github.com/modu-ai/moai-adk/internal/template 3.115s`. No new test authored — the rejection test already existed and was red for the right stated reason (the only C1 match in the tree was this comment). Scope note: the repo's local dogfood twin `.moai/config/sections/workflow.yaml` carries no SPEC id in this comment (verified by grep) and is not enrolled in any byte-parity allowlist for this file, so the template-only edit introduces no mirror drift.

- **R2 (shared workflow loader isolation — same class as audited N2)** — mechanism: `loadWorkflowSection` (the shared wrapper every non-isolated workflow consumer reads through; entry function `Loader.Load` → `(*Loader).loadWorkflowSection`, internal/config/loader.go:236) decoded the whole `WorkflowConfig`, so an alarm-key YAML type error (e.g. `gate_warn_after: [6h]`) failed the whole decode and reverted every workflow setting to defaults — including `local_main_integration.enabled: true`, which would refuse legitimate integration merges. Fix location: internal/config/loader.go:236-256 (recovery branch in `loadWorkflowSection`) and internal/config/loader.go:276-311 (new `decodeWorkflowWithoutAlarm`) — on decode failure the file is re-decoded with the `card_age_alarm` block stripped from the workflow mapping (the cardAgeAlarmFile / factoryLeaseFile isolation pattern), the stripped decode keeps every other key's written value, and the removed block is decoded in isolation: on its own type error the alarm degrades to the seeded compiled default with the established one-warning convention naming `workflow.card_age_alarm`. Tolerance boundary: any OTHER key's type error still fails the stripped pass and keeps the original warn-and-default fallback — the tolerance is not widened beyond the alarm block; the tier-token rejection (REQ-TIER-009) runs on the recovered wrapper unchanged. RED observed LIVE this run on tree `595a9410b` (pre-fix shared path): `go test -count=1 -run 'TestSharedWorkflowSurviveAlarmTypeError' ./internal/config` → `--- FAIL: TestSharedWorkflowSurviveAlarmTypeError (0.01s)` with `loader_card_age_alarm_test.go:268: Workflow.ExecutionMode = "auto", want the written loop`, `:271: Workflow.LocalMainIntegration.Enabled = false, want the written true`, `:274: Workflow.FactoryLease = {Duration:15m0s RenewInterval:5m0s}, want the written 45m/10m`, and `:289: the warning does not name the key path workflow.card_age_alarm` (the observed warning was the generic `failed to load workflow config, using defaults` — the whole-decode failure mechanism itself). GREEN (this run): same command → `--- PASS: TestSharedWorkflowSurviveAlarmTypeError (0.00s)`, `ok github.com/modu-ai/moai-adk/internal/config 0.164s`, with the test's internal assertions verifying exactly one WARN naming `workflow.card_age_alarm`.

Hygiene on the repaired tree (HEAD `9405da696`, all commands this run): `go test -count=1 -run 'TestLoadCardAgeAlarm|TestCardAgeAlarmBounds|TestCardAgeAlarmTypeErrorIsolated|TestQuotaSlotSurviveAlarmTypeError|TestSharedWorkflowSurviveAlarmTypeError' ./internal/config` → `ok github.com/modu-ai/moai-adk/internal/config 0.372s`; `go test -count=1 -run 'TestFactoryStatusAgeAlarm|TestFactoryStatusDisabledUnchanged|TestAgeAlarmNoWrites|TestFactoryAgeAlarmGitEnvIsolation|TestFactoryAgeAlarmChildEnvProtocolPin' ./internal/cli` → `ok github.com/modu-ai/moai-adk/internal/cli 16.338s`; `go test -count=1 ./internal/config` (change-scoped package regression for the shared loader) → `ok github.com/modu-ai/moai-adk/internal/config 13.064s`; `go test -list 'TestSharedWorkflow.*' ./internal/config` lists the new name (swept set non-empty); `go vet ./internal/config ./internal/template` → exit 0; `gofmt -l` on the touched Go files → empty. Commits: RED `595a9410b` (1 file, +58), GREEN `9405da696` (2 files, +71/−3). Working tree clean after both commits.

Gaps: the lint / package-wide `-race` / full-coverage / Windows-execution / integration-CI deferrals carry over unchanged from the base §E.2 — CI on the integration branch remains the owner of the repository-wide verdict, PENDING at this record's time. R1's RED is the pre-existing test (no new rejection test authored — none applies to a comment-token removal beyond the existing C1 sweep). The merged-shape recovery path (alarm type error + tier-token violation in one file) now REJECTS via the tier check where the pre-fix whole-decode failure warned-and-defaulted — the documented contract ("only the tier surface rejects") is the post-fix behavior; the corner combination itself is not separately pinned by a test.

### Rule-C scale-down: R2 reverted (leader d-20261018T085543Z-cda5)

The sync audit FAILED at iter5 on the R2 code (`.moai/reports/t1838/sync-audit-iter5.md`, F1 alias/merge bypass + F2 silent dup-key accept); the leader's ruling (board `d-20261018T085543Z-cda5`) denied a further repair round and ordered R2 reverted via a NEW revert commit (no history rewrite), keeping only R1. Restored to `bb7ac820ae` contents (measured: `git diff bb7ac820ae --stat` after restore = exactly 2 files, template workflow.yaml + this progress.md): `internal/config/loader.go` (the `decodeWorkflowWithoutAlarm` recovery path) and `internal/config/loader_card_age_alarm_test.go` (the `TestSharedWorkflowSurviveAlarmTypeError` test), via `git checkout bb7ac820ae --`. Kept: R1 (the template `workflow.yaml` SPEC-identifier removal), its GREEN verbatim this run: `go test -count=1 -run 'TestTemplateNoInternalContentLeak' ./internal/template` → `ok github.com/modu-ai/moai-adk/internal/template 1.439s`, exit 0. Zero-sweep proof the R2 test is gone — this is the EXPECTED emptiness, verbatim: `go test -count=1 -run 'TestSharedWorkflowSurviveAlarmTypeError' ./internal/config` → `ok github.com/modu-ai/moai-adk/internal/config 0.351s [no tests to run]`, exit 0. Post-restore scoped batches (all this run, captured on the restored working tree — the two Go files byte-identical to what the revert commit carries, differing only by this progress.md addendum; exit 0 each): config 5-name → 5 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/config 0.307s`; cli 5-name → 9 top-level `--- PASS`, 0 `--- FAIL`, `ok github.com/modu-ai/moai-adk/internal/cli 19.897s`; homestate 7-name (unchanged by this revert) → 7 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/homestate 18.262s`; `go vet ./internal/config` → exit 0; `gofmt -l` on the two restored files → empty. R2's original point (shared-decode isolation) and iter5 F1/F2 are recorded as follow-up candidates (leader item 9891) — not repaired in this card.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-18
run_commits: 3b4f6ee31 + d5f2b9b56 (M1 RED+GREEN), 019a3ab03 + 54b08b753 (M2 RED+GREEN), 8670542c6 + 2a1a7f62e (M3 RED+GREEN), plus this evidence commit (HEAD of WT-10-17-g1 immediately after it lands)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-18
sync_commit_sha: 55778cc398c995560ea38fdbfc534a5412bf20b5

## §F Phase 4 Mode Selection

- Input parameters: tier M · scope 6 files (3 new, 3 touched) · domains 1 (Go backend chain `internal/config` → `internal/homestate` → `internal/cli` + 1 template yaml) · language mix Go + yaml · concurrency benefit LOW (coding-heavy, fixed milestone order)
- Mode evaluation: `direct` not selected (specialist domain — manager-develop per Status Transition Ownership Matrix) · `serial` **selected** · `fanout` not selected (coding-heavy per the coding-task parallelism caveat; M2 anchors feed the M3 renderer — inter-milestone dependency) · `sweep` not selected (6 files, not a mechanical bulk transform)
- Decision: serial
- Justification: plan §D fixes M1→M2→M3 order with each milestone's RED landing before its implementation; later milestones consume earlier ones' anchors (§C symbols), so sequential single-specialist delegation is the safe shape. Recorded 2026-10-18 by lane-33 on takeover: M1-M3 had already landed under lane-7's session in exactly this serial shape (one specialist chain, one milestone at a time); the remaining run work — AC verification + §E.2/§E.3 evidence — is one further manager-develop delegation.

## §G Override and Refusal Record

- 2026-10-17T20:45:42Z SPEC-CARD-AGE-ALARM-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1838/.moai/reports/t1838/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1838/.moai/reports/t1838/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1838/.moai/reports/t1838/plan-audit-iter3.md
- 2026-10-17T20:50:31Z SPEC-CARD-AGE-ALARM-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1838/.moai/reports/t1838/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1838/.moai/reports/t1838/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1838/.moai/reports/t1838/plan-audit-iter3.md
