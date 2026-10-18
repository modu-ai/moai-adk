# SPEC-FACTORY-LEADER-ADOPT-CORE-001 — Acceptance Criteria

Every AC is verifiable without a live factory run and without a terminal: instruments are Go tests on `t.TempDir()` factory databases, with in-memory buffers standing in for stderr and the launch seam (`installFactoryLaunchSeam`, `internal/cli/factory_test.go:756-778` — it already captures `EnvFactoryLeadAddr` — or the `driveFactoryEntry` pattern, `internal/cli/codex_factory_retire_test.go:460-474`) standing in for the session. The Go-test form is one compound invocation with the environment scrubbed (a lane-stamped environment falsifies tests that read it):

```bash
unset MOAI_FACTORY_WORKERS MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_NAME MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test <one package> -run '^(<TestName1>|<TestName2>)$' -count=1 -v
```

**Swept-count rule.** A Go-test AC passes only when the `-v` output shows one top-level `--- PASS` line (no leading spaces) per named test; `[no tests to run]`, a PASS count below the named count, or a missing name is a FAIL. The `-run` pattern anchors every alternation branch.

**Two-cell discipline.** Each AC carries a RED-now cell (the ledger entry that shows it red on the pre-implementation tree, and why) and a green-path cell (the milestone that flips it). The RED-now evidence is the same for every new Go-test AC of a package: the named test does not exist, so the sweep is empty. That is red for the stated reason that the criterion asserts a behavior that does not exist yet, and it is not vacuous because the swept-count rule turns an empty sweep into a FAIL. A behavioral RED (test present, assertion failing against the old code) is observed by manager-develop at each milestone's RED step and recorded in `progress.md` §E.2.

**Mutant probe.** Each AC names the mutants that satisfy a weaker reading and must fail the AC. **No seam-injected default exists in this SPEC** — there is no interactivity detection (deferred to the follow-up card), so the predecessor's D15 class (a one-directional default) has no surface here; the one input seam (`classifyNamedRun`, plan M2 item 5) is exercised in both directions by AC-FLAC-003 (live-owner takeover proceeds) and AC-FLAC-004(c) (owner-change aborts).

## Evidence ledger (measured on tree `2178966bb90b28576e8a4c3bce340947b83c7cda`, this run; Go tests judged by this tree's Go toolchain, `moai` judged by the installed build `ee981a932`, a verified ancestor of HEAD)

```text
E1  command : unset <scrub list> && go test ./internal/homestate -run '^(TestFindOrphanRunsRequiresDeadStampAndCards|TestFindOrphanRunsExcludesNonQualifyingRuns|TestFindOrphanRunsCountsNonTerminalCardsOnly)$' -count=1 -v
    stdout  : testing: warning: no tests to run
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/homestate	0.399s [no tests to run]
    exit    : 0
    verdict : 0 of 3 named tests swept -> FAIL under the swept-count rule
E2  command : unset <scrub list> && go test ./internal/cli -run '^(TestLeaderStartOrphanRefusalPrecedesRunWrite|TestLeaderStartOrphanGuidanceRendersFacts|TestLeaderStartOrphanGating|TestLeaderStartOrphanTouchesNoOtherRun|TestLeaderStartOrphanGLMMirrorsCC|TestLeaderStartOrphanTimesRenderInKST|TestExplicitFactoryRunAdoptTakesOverAtomically|TestExplicitFactoryRunAdoptRefusals|TestLeaderAdoptGuidanceDocsMatchAnchors)$' -count=1 -v
    stdout  : testing: warning: no tests to run
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/cli	1.403s [no tests to run]
    exit    : 0
    verdict : 0 of 9 named tests swept -> FAIL
E3  command : unset <scrub list> && go test ./internal/cli -run '^(TestAdoptResumedLeaderRunAdoptsSingleCandidate|TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped|TestAdoptResumedLeaderRunWarnsOnMultipleCandidates|TestAdoptResumedLeaderRunMintsWhenNoCandidates|TestAdoptResumedLeaderRunWarnsCodexLedCandidate|TestStrictResumeCarrierSeen|TestLeaderRunIDWasMinted|TestFactorySkillAssertionsMatchBehavior|TestFactoryLeadRefusesCodexLedRun|TestFactoryJoinAndLeadAcceptNonCodexRun)$' -count=1 -v
    stdout  : --- PASS: TestFactoryJoinAndLeadAcceptNonCodexRun (8.05s)
              --- PASS: TestFactoryLeadRefusesCodexLedRun (2.29s)
              --- PASS: TestAdoptResumedLeaderRunAdoptsSingleCandidate (2.38s)
              --- PASS: TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped (0.48s)
              --- PASS: TestAdoptResumedLeaderRunWarnsOnMultipleCandidates (0.73s)
              --- PASS: TestAdoptResumedLeaderRunMintsWhenNoCandidates (0.58s)
              --- PASS: TestAdoptResumedLeaderRunWarnsCodexLedCandidate (1.56s)
              --- PASS: TestStrictResumeCarrierSeen (0.00s)
              --- PASS: TestLeaderRunIDWasMinted (0.00s)
              --- PASS: TestFactorySkillAssertionsMatchBehavior (15.61s)
              ok  github.com/modu-ai/moai-adk/internal/cli	33.047s
    exit    : 0
    verdict : 10 of 10 swept, GREEN baseline — regression guards: the resume adoption's outputs, the explicit path's codex refusal and today's lead takeover, the skill document test
E4  command : unset <scrub list> && go test ./internal/homestate -run '^(TestAdoptRunRefusesChangedStamp|TestLiveRecordsTerminalSetMatchesSharedDefinition|TestReconcilePreservesDeadOwnerRunWithLiveRecords|TestReconcilePreservesDeadOwnerWithOnlyWorkers|TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly|TestFindResumableRunAdoptsSingleDeadOwnerWithLiveRecords|TestFindResumableRunSkipsLiveOwner|TestFindResumableRunSkipsRunWithoutLiveRecords|TestFindResumableRunReportsMultipleCandidates)$' -count=1 -v
    stdout  : --- PASS: TestFindResumableRunAdoptsSingleDeadOwnerWithLiveRecords (1.13s)
              --- PASS: TestFindResumableRunSkipsLiveOwner (0.75s)
              --- PASS: TestFindResumableRunSkipsRunWithoutLiveRecords (0.80s)
              --- PASS: TestFindResumableRunReportsMultipleCandidates (0.73s)
              --- PASS: TestAdoptRunRefusesChangedStamp (0.46s)
              --- PASS: TestLiveRecordsTerminalSetMatchesSharedDefinition (0.00s)
              --- PASS: TestReconcilePreservesDeadOwnerRunWithLiveRecords (0.44s)
              --- PASS: TestReconcilePreservesDeadOwnerWithOnlyWorkers (0.52s)
              --- PASS: TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly (0.66s)
              ok  github.com/modu-ai/moai-adk/internal/homestate	5.759s
    exit    : 0
    verdict : 9 of 9 swept, GREEN baseline — the adoption write's CAS refusal, the shared terminal set, the reconcile preservation semantics
E5  command : grep -c 'leader-start/orphan' docs-site/content/en/advanced/factory-mode.md docs-site/content/ko/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md README.md README.ko.md README.ja.md README.zh.md .claude/skills/moai/workflows/factory.md internal/template/templates/.claude/skills/moai/workflows/factory.md
    stdout  : <all ten files>:0
    exit    : 0
    verdict : the anchor span is absent from all ten guidance files -> RED for AC-FLAC-010
    note    : grep -n for the four banned README substrings ('retired automatically when the next lane joins' / '자동으로 회수되므로' / '次にレーンが合流した時点で自動的に retire されるので' / '会在下一条泳道加入时自动退役') each matched at line 70 of its README on this tree — the sentences M4 must REPLACE are present
E6  command : grep -n '^## ' docs-site/content/en/advanced/factory-mode.md | wc -l   (and the same for ko/ja/zh)
    stdout  : 9 for each of the four locales (headings verified by reading; the equality AC asserts locale-locale equality, never a fixed count)
    exit    : 0
    verdict : context only — the four locales agree on 9 level-2 headings before the change
```

Reading note: E3 and E4 are baselines for behavior this SPEC must not change (the resume adoption, the lane-join preservation of orphan runs, today's explicit-path outcomes, the skill document test); they are GREEN today and are release-blocking only as regression guards. E3's `TestFactoryJoinAndLeadAcceptNonCodexRun` and `TestFactoryLeadRefusesCodexLedRun` pin the exact surface M2 re-routes (plan M2, "Watch the pinning tests").

## Coverage

| AC | Requirement | Proving instrument | Green path |
|---|---|---|---|
| AC-FLAC-001 | REQ-FLAC-001 | E1 names (3 homestate tests) | M1 |
| AC-FLAC-002 | REQ-FLAC-002 | E2 name `RefusalPrecedesRunWrite` | M3 |
| AC-FLAC-003 | REQ-FLAC-003 | E2 names `ExplicitFactoryRunAdoptTakesOverAtomically` | M2 |
| AC-FLAC-004 | REQ-FLAC-004 | E2 name `ExplicitFactoryRunAdoptRefusals` | M2 |
| AC-FLAC-005 | REQ-FLAC-005 | E2 name `GuidanceRendersFacts` | M3 |
| AC-FLAC-006 | REQ-FLAC-006 | E2 name `Gating` plus E3 unchanged | M3 |
| AC-FLAC-007 | REQ-FLAC-007 | E2 name `TouchesNoOtherRun` | M3 |
| AC-FLAC-008 | REQ-FLAC-008 | E2 name `GLMMirrorsCC` | M3 |
| AC-FLAC-009 | REQ-FLAC-009 | E2 name `TimesRenderInKST` | M3 |
| AC-FLAC-010 | REQ-FLAC-010 | E2 name `TestLeaderAdoptGuidanceDocsMatchAnchors` plus E3's skill test | M4 |

## Acceptance criteria

### AC-FLAC-001 (maps REQ-FLAC-001)

- **Given** a sandbox factory database and a `ReconcileOptions` value the homestate test builds itself (no `factorymsg` import, which would invert the import direction): a per-pid `Classify` function (pid 101 dead, 102 dead, 103 dead, 104 live, 106 indeterminate, 107 dead, 108 dead, 42 dead) and a `Fallback` that returns `(42, "peer-start", true)` for run E only. Runs, each recorded with `mustRecordRun` and cards added with `mustInsertCard` (the fixtures of `internal/homestate/factory_run_adopt_test.go`): A, pid 101, two non-terminal cards (states `run` and `blocked`) and one `done`; A2, pid 108, one non-terminal card (state `blocked`) and a distinct `created_at` — a second qualifying orphan; B, pid 102, only `done`, `failed` and `abandoned` cards; C, pid 103, one workers row (`mustInsertWorker`) and no cards; D, pid 104 (classified live), one non-terminal card; E, `LeadPID` 0 and an empty start so the owner resolves through the Fallback (basis peer, classified dead), one non-terminal card; F, pid 106 (classified indeterminate), one non-terminal card; G, pid 107 (classified dead), one non-terminal card, status `retired`; H, `LeadPID` 0 and an empty start with NO Fallback hit, its run row and its one non-terminal card seeded with timestamps before the injected boot time, and the `BootTime` and `LegacyPeerRecordAbsent` seams set (`internal/homestate/factory_run_retire.go:96-101`) so the row classifies dead through the boot proof (`BasisBoot`, `:266-274`) — the excluded boot-basis shape.
- **When** `FindOrphanRuns` runs with those options
- **Then** it returns exactly runs A and A2 with non-terminal card counts of 2 and 1 respectively, returns none of B to H, and writes nothing: a content snapshot (every row of `runs`, `cards`, `events` and `workers`, ordered) taken before equals the one taken after.
- **Instrument**: E1, three top-level PASS lines. **RED-now**: E1 (names absent). **Green path**: M1.
- **Mutants that must fail** (each detectable because the excluded run carries a non-terminal card): dropping the dead-classification test returns D and F; dropping the stamp-basis test returns E; an admit rule widened to `Basis != BasisPeer` (instead of `Basis == BasisStamp`) returns H; counting workers as live records returns C; counting every card gives A a count of 3; dropping the card test returns B and C; a finder with no status filter returns the retired run G; a finder that returns only the first qualifying orphan omits A2 (the "every" quantifier); a finder that updates or deletes a `workers` row fails the content snapshot.

### AC-FLAC-002 (maps REQ-FLAC-002)

- **Given** one orphan run (the AC-FLAC-001 run A fixture), a sandbox with the `runs` and `events` rows snapshotted, the launch seam (`installFactoryLaunchSeam`) counting session launches, a buffer standing for the cobra command's error stream, and the cobra command's input set to a counting reader that records every `Read` invocation
- **When** the `moai cc -f` entry runs with no resume option and no `--factory-run`
- **Then** the entry returns a non-nil error (a non-zero exit); at every point observable from the test the `runs` row count equals its pre-launch value and no event names the minted id (the refusal precedes the run write); the launch seam was never called (no session started); the counting reader recorded exactly zero `Read` calls (no path read the command's input).
- **Instrument**: E2 name `TestLeaderStartOrphanRefusalPrecedesRunWrite`. **RED-now**: E2. **Green path**: M3.
- **Mutants that must fail**: a gate placed after `recordLeaderRunStart` records the minted run (the post-condition on `runs`/events fails); a gate that returns a nil error exits zero; a gate that falls through to the session calls the seam; a gate that prompts (or reads input in any form through the command) trips the counting reader — an empty reader would return n=0/EOF and prove nothing, which is why the counter, not consumption bytes, is the instrument.

### AC-FLAC-003 (maps REQ-FLAC-003)

- **Given** one active run with two cards, `lane_capacity` seeded to 7, owned by a dead stamp (`LeadPID 99999999`, fingerprint `gone-before-resume`), its `lead_backend` seeded to the glm backend (a cross-backend takeover: the `cc` launcher adopting a glm-led run — the codex-led shape is AC-FLAC-004(a)'s refusal), and the launch seam capturing the environment
- **When** the `moai cc -f --factory-run <run id>` entry runs
- **Then** the seam captured `EnvFactoryRunID` equal to the named id and `EnvFactoryLeadAddr` equal to `factory.FactoryLeaderSocketPath(<run id>)`; the run row's owner stamp equals the launching process's pid and fingerprint; the run row's `lead_backend` equals `claude` — the adopting launcher's backend, written by the adoption write's own transaction, not left at the seeded `glm`; exactly one `run.resumed` event was appended with basis `leader-start-adoption`; no `run.started` event exists for the run and the `runs` row count is unchanged (no second row, no re-activation write); both card rows are byte-identical to their pre-launch snapshot; `lane_capacity` is still 7.
- **And** the same launch over a run whose owner classifies LIVE (the row carries the test process's own pid and true fingerprint, and the claude backend — the seed that AC-FLAC-008's glm leg flips): **then** the takeover succeeds through the same write with the same invariants (one `run.resumed`, no `run.started`, address republished, stamp now the launcher's, `lead_backend` the launcher's) — the input seam's proceed direction.
- **Instrument**: E2 name `TestExplicitFactoryRunAdoptTakesOverAtomically`. **RED-now**: E2. **Green path**: M2.
- **Mutants that must fail**: adopting through `RecordRun` (today's unconditional upsert) appends `run.started`, clobbers `lane_capacity`, and fails both; a takeover that skips the `lead_backend` write leaves each row at its seed (`glm` on the dead-stamp fixture, `claude` on the live-owner fixture — the seeds are crossed so each launcher leg of AC-FLAC-008 discriminates); skipping the address republish leaves `EnvFactoryLeadAddr` at the minted id's socket (the captured-addr assertion fails); a compare-and-set against a stale or hard-coded stamp fails the live-owner scenario (the write misses, the launch errors, the scenario's success assertion fails).

### AC-FLAC-004 (maps REQ-FLAC-004)

- **Given** an active dead-stamp run with one non-terminal card, in three variants: (a) the run is led by codex (`lead_backend` set to the codex backend — the fixture pattern of `TestAdoptResumedLeaderRunWarnsCodexLedCandidate`); (b) the run holds a live legacy peer record (the broker-database fixture of `TestEnterSelectedFactoryRunRefusesLiveLegacyPeer`, `internal/cli/factory_run_boundary_m1_test.go:57`); (c) the `classifyNamedRun` seam (plan M2 item 5) is overridden to restamp the run row's `lead_pid` and `lead_process_start` to a different value before returning the classification — the competing writer inside the launch's classify-to-write window
- **When** the `moai cc -f --factory-run <run id>` entry runs in each variant
- **Then** each returns a non-nil error whose text names the refusal; the launch seam was never called; no new `runs` row exists and no event was appended; the run row is byte-identical to the snapshot taken right after the competing restamp in variant (c) — the competing writer's stamp survives — and to the pre-launch snapshot in (a) and (b).
- **Instrument**: E2 name `TestExplicitFactoryRunAdoptRefusals`. **RED-now**: E2. **Green path**: M2 (variants a/b are pinned GREEN today by E3's `TestFactoryLeadRefusesCodexLedRun`; the AC adds the nothing-minted/no-event assertions and variant c).
- **Mutants that must fail**: a refusal that falls through to the mint leaves a second `runs` row; a compare-and-set that drops the owner-stamp `WHERE` clause succeeds in variant (c) and overwrites the competing writer's stamp (the byte-identical assertion fails).

### AC-FLAC-005 (maps REQ-FLAC-005)

- **Given** two orphan runs — the AC-FLAC-001 A and A2 fixtures (distinct run ids, dead owners, distinct lead pids, distinct parseable `created_at` values, non-terminal card counts 2 and 1) — and a buffer standing for the cobra command's error stream
- **When** the `moai cc -f` entry runs, and again the `moai glm -f` entry
- **Then** each returns a non-nil error; the error stream carries exactly one guidance block (the block's header line occurs exactly once) naming EACH orphan run — one fact line per orphan (its run id, the word `dead`, its lead pid, its non-terminal card count, its start time) and one deliberate adopt command per orphan (`moai cc -f --factory-run <first run id>` and `moai cc -f --factory-run <second run id>`, respectively the two `moai glm -f` forms); the `runs` row count equals its pre-launch value; no event was appended; the launch seam was never called; both orphan rows are unchanged.
- **Instrument**: E2 name `TestLeaderStartOrphanGuidanceRendersFacts`. **RED-now**: E2. **Green path**: M3.
- **Mutants that must fail**: a path that mints after the guidance leaves a second `runs` row; a path that returns a nil error exits zero; a renderer that lists only the first orphan omits the second run's fact line and adopt command (the "each" quantifier); a block naming the other launcher's command (or the hard-coded `moai cc -f` of the resume warn line, `internal/cli/factory_run_adopt.go:79`) fails the glm command-string assertion; a renderer that emits the block twice fails the header-count assertion.

### AC-FLAC-006 (maps REQ-FLAC-006)

- **Given** a table of launches over the orphan fixture, each with the guidance output and the refusal instrumented: (1) `--resume <id>`; (2) the attached form `-r<id>`; (3) `--factory-run <orphan id>`; (4) a non-empty `MOAI_KANBAN_ID`; (5) a pasted `leader-<run-id>` name; and, with no resume option, (6) no orphan because the only dead-stamp run holds only terminal cards; (7) a workers-only dead-stamp run; (8) a live-owner run with a non-terminal card; (9) a dead-stamp run holding a non-terminal card whose status is `retired`; (10) no runs at all; (11) `-c <id>` (not part of the resume vocabulary, `internal/cli/lane_resume.go:36-37`) over the orphan fixture; (12) `-cr <id>` — an r-carrying short cluster that `scanResumeArgs`' `isShortClusterCarryingR` fires on (`lane_resume.go:74-76`) but `strictResumeCarrierSeen` deliberately does not (no `-r` prefix); (13) `--resume <id>` over the TWO-orphan fixture (the AC-FLAC-001 runs A and A2); and (14) `--factory-run <run A id>` over the TWO-orphan fixture
- **When** each launch runs
- **Then** rows 1 to 10 write no guidance block, refuse for no orphan-related reason, and produce the outcome the same launch produces on the base tree (rows 1-2 take the resume adoption path — E3 pins their outputs; row 3 adopts through M2's write — AC-FLAC-003; rows 4-5 proceed as today with their deliberate id; rows 6-10 mint and start the session as today); rows 11 and 12 are each refused with the guidance block (a `-c` or `-cr` start over an orphan is a fresh mint over an orphan); row 13 writes no guidance block and refuses for no orphan-related reason — with two stamp-dead candidates the resume adoption admits neither (`FindResumableRun`'s single-candidate rule, `internal/homestate/factory_run_adopt.go:92-95`), warns, and mints exactly as E3's `TestAdoptResumedLeaderRunWarnsOnMultipleCandidates` pins, so BOTH orphans survive the resume decision and a gate evaluated after it still sees two; row 14 adopts run A through M2's write (AC-FLAC-003's outcome) while A2 stays dead-stamped and byte-unchanged, with no guidance block and no refusal.
- **Instrument**: E2 name `TestLeaderStartOrphanGating`, plus E3 unchanged. **RED-now**: E2. **Green path**: M3.
- **Mutants that must fail**: a gate that omits the resume test (`!strictResumeCarrierSeen`) refuses row 13 — both orphans survive the resume path's warn-and-adopt-nothing, so the gate sees them and misfires (on rows 1-2's single orphan the resume adoption consumes the orphan before the gate and masks the mutant, which is why the catch lives on the two-orphan row); a gate that omits the `--factory-run` test refuses row 14 — run A's adoption leaves A2 dead-stamped and visible (row 3's single orphan is consumed the same way, so that catch too lives on the two-orphan row); a gate that omits the fresh-mint test refuses rows 4 and 5; a finder with no status filter refuses row 9; a finder that counts workers refuses row 7; a gate that treats `-c` itself as a resume carrier exempts row 11; a gate widened to the broader `scanResumeArgs` vocabulary (whose `isShortClusterCarryingR` fires on the `r` inside `-cr`) exempts row 12 — the over-refusal / under-refusal pair of the predecessor's D14 class, both directions named.

### AC-FLAC-007 (maps REQ-FLAC-007)

- **Given** a database with the orphan run, a second dead-stamp run holding only terminal cards, a live-owner run, and a retired run; and three launches, each from a fresh copy of the database: the bare `-f` refusal over the orphan, the deliberate `--factory-run` adopt of the orphan, and a no-orphan `-f` mint (empty database)
- **When** each launch runs
- **Then** after each, every run row other than the adopted one (adopt launch) or the minted one (mint launch) is byte-identical to its pre-launch snapshot, and after the refusal every row is byte-identical; for the lost-compare-and-set adopt (AC-FLAC-004(c)) the comparison snapshot is the one taken right after the competing restamp and the restamped orphan keeps the competing writer's stamp; in every case no `run.retired` event exists anywhere in the database; and a bare `FindOrphanRuns` call on the pre-launch database leaves a full content snapshot byte-identical (the finder itself writes nothing).
- **Instrument**: E2 name `TestLeaderStartOrphanTouchesNoOtherRun`. **RED-now**: E2. **Green path**: M3 (the finder half lands with M1 and is asserted there too).
- **Mutants that must fail**: a gate that calls `ReconcileActiveRuns` retires the terminal-only run (a `run.retired` event appears); an adopt that restamps the live-owner run while adopting the orphan fails that row's byte-identical assertion; a finder that writes fails its own content snapshot.

### AC-FLAC-008 (maps REQ-FLAC-008)

- **Given** the AC-FLAC-003 fixtures with crossed seeds (the dead-stamp run seeded `glm` — flipped to `claude` by the cc takeover; the live-owner run seeded `claude` — flipped to `glm` by the glm takeover; each leg adopts a run whose seed is the OTHER launcher's backend, so neither leg's backend equality can hold without the write), AC-FLAC-004(a) and AC-FLAC-005 fixtures
- **When** each is driven once through the `moai cc -f` entry and once through the `moai glm -f` entry
- **Then** the run states, the events and the error outcomes are identical across the two entries except for the backend recorded on the runs row (`claude` after the cc takeover of the glm-seeded run, `glm` after the glm takeover of the claude-seeded run — each leg's expectation is the other launcher's backend, so the difference is observable and non-vacuous on BOTH legs); the glm guidance and refusal text equals the cc text after replacing the entry command `moai glm -f` with `moai cc -f`; each entry's text contains its own entry command and not the other's.
- **Instrument**: E2 name `TestLeaderStartOrphanGLMMirrorsCC`. **RED-now**: E2. **Green path**: M3.
- **Mutants that must fail**: a gate wired into `cc.go` only leaves the `glm` entry minting (row counts diverge); an adoption write that pins one launcher's backend, or omits the write on one launcher's path, leaves the seed on at least one leg (each leg's seed is the other launcher's backend, so the per-launcher omission is caught); guidance text that copies the hard-coded `moai cc -f` of the resume warn line fails the glm command-string assertion.

### AC-FLAC-009 (maps REQ-FLAC-009)

- **Given** an orphan run whose `created_at` is `2026-10-17T13:39:10.123456789Z`, plus a second whose `created_at` is the unparseable string `not-a-time`
- **When** the guidance block renders
- **Then** the first renders `2026-10-17 22:39:10 KST` and the second renders `not-a-time` unchanged.
- **Instrument**: E2 name `TestLeaderStartOrphanTimesRenderInKST`. **RED-now**: E2. **Green path**: M3.
- **Mutants that must fail**: a UTC rendering prints `13:39:10`; a rendering in `time.Local` fails because the test sets `time.Local` to a fixed zone other than UTC and KST (`UTC-5`) for its duration.

### AC-FLAC-010 (maps REQ-FLAC-010)

- **Given** the ten guidance files of ledger E5 and two anchor lines, each a backticked span built by the test from the same constants the guidance renderer uses: `leader-start/orphan: refused; no run created; adopt with --factory-run <run-id>` and `lane-join/dead-owner-with-cards-or-workers: kept`
- **When** `TestLeaderAdoptGuidanceDocsMatchAnchors` reads them
- **Then** each of the ten files contains each anchor span exactly once; the unqualified legacy sentence is absent from the four READMEs (the substrings `is retired automatically when the next lane joins`, `자동으로 회수되므로`, `次にレーンが合流した時点で自動的に retire されるので` and `会在下一条泳道加入时自动退役` do not occur in `README.md`, `README.ko.md`, `README.ja.md` and `README.zh.md` respectively); no file contains an anchor span with a different value for either key (for example `leader-start/orphan: mint`, `leader-start/orphan: adopt automatically`, `lane-join/dead-owner-with-cards-or-workers: retired`); the four docs-site pages have the same number of level-2 headings as one another and both anchors sit inside one level-2 section of each; the template skill and its local twin are byte-identical; and `TestFactorySkillAssertionsMatchBehavior` still passes (E3). The pre-change heading count (E6) is context only: no fixed count is asserted.
- **Instrument**: `go test ./internal/cli -run '^(TestLeaderAdoptGuidanceDocsMatchAnchors|TestFactorySkillAssertionsMatchBehavior)$' -count=1 -v`, two top-level PASS lines. **RED-now**: E2 (test absent) and E5 (anchor 0 occurrences in all ten files). **Green path**: M4.
- **Mutants that must fail**: a document that carries one anchor and states the opposite for the other (a mint instead of a refusal, a retirement despite cards) fails the value assertion; a document missing an anchor fails the presence assertion; a README that keeps the legacy sentence beside correct anchors fails the substring ban; a locale page missing the new section fails the heading-equality and single-section assertions; a skill span `moai cc -f --factory-run X` added to the prose fails `TestFactorySkillAssertionsMatchBehavior`. Prose outside the anchor lines is reviewed, not machine-judged.

## Edge cases

- An orphan run whose cards are all in a state outside the terminal set but never leased (`picked`, `blocked`, `needs-decision`) counts as an orphan (the terminal set is `{done, failed, abandoned}`); covered by the count in AC-FLAC-001.
- A `-c` (continue) launch over an orphan is refused with guidance like a bare start — it is not a resume carrier (AC-FLAC-006 row 11).
- `moai codex -f` is refused before the leader branch (entry-table row, replayed by E3); nothing here adds a codex leader path.
- An interactive terminal start over an orphan is refused with the same guidance — the refusal reads no stream (spec.md Terms, DP-3); no AC contradicts this, and the follow-up chooser card is the surface that will change it.

## Definition of Done

- AC-FLAC-001 to AC-FLAC-010 pass under the swept-count rule on the final tree, with their verbatim output in `progress.md` §E.2.
- E3 and E4 are still green on the final tree.
- `moai spec lint --strict` over the SPEC reports no findings, naming the judging build.
- The twin skill files are byte-identical and `make build` leaves no tracked-file change.
- The four-locale verify recipe of `hns-oss-docs-verify` passes.
