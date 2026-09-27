# progress — SPEC-ROLE-NAMING-CODE-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1256 · branch `WT-role-naming-code` · plan base `e62c3e183`
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L) + this file
- Evidence: `.moai/reports/t1256/census.md` (reproducible via `census.py`), `.moai/reports/t1256/conflicts.md`
- v0.1.0 (`6fe67c674`): REQ 20 · AC 20.
- v0.2.0 revision (2026-09-26): operator answers relayed by the leader applied — `lane` canonical, legacy spellings rejected (no aliases), persisted values write-new with the run-boundary rule, role-marker guard `lane`-only, lane self-dispatch help text, homonym qualifiers, `manager-lead` kept, t1193 demoted to a recorded dependency, new ordering t1242 → t1245 → t1256 run → t1240 → t1257 (t1193 excluded pending decision), design §3 zh note scoped to the code layer.
  - Rewritten REQ: 001, 003, 004, 005, 007, 009, 010, 011, 012, 013, 014, 016, 017, 018, 019, 020. New REQ: 021, 022, 023. Unchanged: 002, 006, 008, 015.
  - Rewritten AC: 002, 003, 004, 005, 007, 008, 009, 011, 013, 014, 015, 016, 018, 019, 020. New AC: 021, 022, 023. Unchanged: 001, 006, 010, 012, 017.
  - REQ count 23 · AC count 23 (AC-RNC-001..023, contiguous); every REQ maps to ≥1 AC (acceptance.md §C).
- v0.2.0 items to confirm at Kickoff: O0 derived `lead` rejection, O1 env names kept, O2 run-boundary persisted-data rule; O7 open. **Superseded by v0.3.0 below.**
- Plan-audit iteration 1: FAIL 0.77 (`.moai/reports/t1256/plan-audit-iter1.md`, audited at `d0770b9cc`) — MP-7 (O7 open) + blocking D2–D7, D9.
- v0.3.0 revision (2026-09-26):
  - Operator answers given directly in the lane window recorded in plan.md §B as RESOLVED: O7 rename role-sense Go identifiers; O0 `lead`/`lead-<suffix>` rejection confirmed; O1 env var names kept; O2 live legacy records refused with retire-and-relaunch guidance. No clarification marker remains in plan.md or research.md.
  - Rewritten REQ: 001 (D6), 003/005 (case variants, D16), 004 (`<n>` ≥ 1), 009 (detection carve-out), 011 (D18), 012 (D4 — equality assertion is the one coupling), 018 (D2/D5/D15), 019 (D1/D3), 022 (D7 — factory-run scope and membership basis). New REQ: 024 (D9 — legacy-only run retire), 025 (D7/D8 — kanban registry, legacy board role declaration, SessionStart session-record writer). REQ-number placement note added (D12).
  - Rewritten AC: 002, 003, 004, 005, 006, 008, 011, 012, 013 (D13 fixed string), 014, 016 (D5), 018, 019, 022, 023 (D14 fixed population command). New AC: 024, 025. Word-boundary rule stated once at the top of acceptance.md.
  - plan.md: §B all resolved; §C.3 sibling-SPEC re-check with plan-time branch @ SHA (D10); §E remeasure scope adds `./internal/spec` (plus `./internal/homestate`, `./internal/web`); M1 names the SessionStart session-record writer and the retire fallback (D8, D9); M6 records the partial supersession of SPEC-FACTORY-WORKER-NAMING-001 at sync (D11); R9 added.
  - design.md D2/D4/D7/D8 and §2/§4 updated; research.md §3 splits board role declarations from session records (D8).
  - Evidence hygiene: stray nested copy `.moai/reports/t1256/.moai/` removed (D17). No auditor helper scripts (`recount.py`, `sample.py`) found untracked in the worktree.
  - REQ count 25 · AC count 25 (AC-RNC-001..025, contiguous) — both at the Tier L ceiling of 25, neither over; every REQ maps to ≥1 AC (acceptance.md §C).
- Plan-audit iteration 2: FAIL 0.845 (`.moai/reports/t1256/plan-audit-iter2.md`, audited at `5102a69e9`) — blocking N1 (AC-RNC-008 required a role-declaration write no production path performs) + optional N2–N8.
- v0.3.1 revision (2026-09-26), no REQ or AC added (both stay at 25):
  - N1 (blocking): measured at `5102a69e9` over production `.go` files — `DeclareRole` 0 callers; `ResolveDeclaredRole` called only at `internal/kanban/board_store.go:191`; `WriteBoardState` only at `board_store.go:363` (inside `TransitionIntoRunOpts`), which is called only at `board_store.go:344` (inside `TransitionIntoRun`), which has 0 callers; `RecoverBoard` 0. AC-RNC-008 drops the declaration from the launch-written records and verifies the read side with a test-written `DeclareRole` declaration; REQ-RNC-025 / AC-RNC-025 board clause names the legacy role `lead` instead of a relaunch remedy and states it governs guard behavior only; plan R1 rewritten with the measured call chain; design.md §2/§4 wording aligned. No launcher declaration write added.
  - N2: AC-RNC-008 lists each record with its expected value in a table.
  - N3: `lead` matched case-insensitively (`(?i)\blead\b`) in REQ-RNC-018 and the acceptance word-boundary rule (current-tree measurement: case-insensitive and case-sensitive string-literal counts both 29, so no existing `Lead`/`LEAD` hit). The `MOAI_*` exclusion cannot produce a match under word-boundary semantics (`_` is a word character — `perl -ne 'print if /\blead\b/i'` prints nothing for `MOAI_KANBAN_LEAD_ADDR`), so AC-RNC-018's control is now paired: token-only string passes, the same token plus a free-standing `lead` fails (a real hit); a `Lead` sentence-start mutation added.
  - N4: AC-RNC-023 population command strips Go comments (`sed -E 's#[[:space:]]+//.*$##'` then re-grep); reproduced at `5102a69e9` → 2 rows (`internal/cli/factory.go:544`, `internal/cli/kanban.go:711`), the `internal/statusline/types.go:250` comment row gone.
  - N5: REQ-RNC-022 states an empty-`run_id` legacy claim belongs to no run, is not refused, is not recognized as a lane, and is stale once dead.
  - N6: AC-RNC-019 adds the M5 census re-run recording remaining identifier-internal role-sense rows, each tagged with its exclusion, untagged count 0.
  - N7: allowlist entries bound to file + exact literal; line numbers are recorded only (REQ-RNC-018, AC-RNC-018).
  - N8: AC-RNC-025 SessionStart case split into (a) label trigger with no record and (b) record trigger with a `leader` label.
- Plan-audit iteration 3 (final, Tier L ceiling): FAIL 0.89 (`.moai/reports/t1256/plan-audit-iter3.md`, audited at `7c2b4d528`) — mandatory gates 7/7, blocking P1 only (plan.md M5 and R5 still described the allowlist as file:line-bound), optional P2–P4.
- Post-audit fixes (2026-09-26), no REQ or AC added (both stay at 25):
  - P1: plan.md M5 and R5 now state allowlist entries are bound to file + exact literal, the line number recorded only — consistent with REQ-RNC-018 and AC-RNC-018.
  - P2: census script committed at `.moai/specs/SPEC-ROLE-NAMING-CODE-001/census.py` with an `--out` argument (default output under the ignored `.moai/reports/t1256/raw/`); output byte-identical to the report copy on the current tree (16608 rows). AC-RNC-019, plan.md §C pre-flight step 4, and research.md §1 reference the tracked path.
- Accepted debt (operator override, not fixed at plan close):
  - P3: REQ-RNC-022's empty-`run_id` legacy-claim clause has no dedicated AC-RNC-022 case. Run phase should add the Given/When/Then (live legacy claim with `run_id=''`, owner `worker-5` → factory leader launch proceeds, claim not counted as a lane) when implementing AC-RNC-022.
  - P4: AC-RNC-025(a) requires "no session record file is created", which REQ-RNC-025 implies ("instead of re-deriving") but does not state; wording alignment deferred.
- Plan-audit final state: iteration 3 FAIL 0.89 → operator override to **PASS-WITH-DEBT** after the P1 fix, decided by the operator in the lane window on 2026-09-26; no re-audit.
- Implementation Kickoff: approved by the operator on 2026-09-26, progression mode semi-autonomous.
- Status: plan closed; frontmatter `status: draft` left for manager-develop's `draft → in-progress` at run entry.

## §F Phase 4 Mode Selection

- Input parameters: tier L · scope ≈ census 88 production files / 1,261 occurrences (post-absorb 16,124 census rows) · domains 6+ (kanban, cli, hook, factorymsg, homestate, web, config, codexwiring) · file mix Go + i18n string tables · concurrency benefit LOW (coding-heavy: rename with rejection/stale-record behavior changes and per-package tests) · Agent Teams prereqs: not requested.
- Mode evaluation: `direct` not selected (Tier L, multi-file, behavior-bearing) · `serial` **selected** (coding-heavy; one manager-develop milestone spawn at a time, M1→M5 ordered by change likelihood) · `fanout` not selected (coding-heavy, write-capable — violates one-writer discipline) · `sweep` not selected (multi-rule semantic change, not one uniform mechanical transform; the mechanical identifier rename is compiler-checked and stays inside serial M5) · `agent-team` not selected (no operator request).
- Decision: serial.
- Justification: the rename carries behavior changes (rejection paths, run-boundary refusal, SessionStart writer guard, retire fallback) whose tests must land before the rename in the same packages — sequential milestone delegation with scoped per-package verification is the safe shape; high-volume mechanical transform criteria for sweep do not hold across milestones (uniform rule only within M5, which rides the serial chain anyway). Implementation Kickoff Approval: approved by the operator on 2026-09-26 (§E.1); progression mode semi-autonomous (operator policy CLAUDE.local.md §31).

## §E.2 Run-phase Evidence

_M1 (card t1256, branch `WT-role-naming-code`). Attribution: every row names the command, its verbatim output, and the HEAD SHA it was measured on._

### Pre-flight (REQ-RNC-014 / AC-RNC-015)

| Check | Observed (measured by the orchestrator on `d39a1dc09`) |
|---|---|
| develop SHA | `b59a5d69c`; `git rev-list --count --left-right origin/develop...develop` → `0 0` |
| t1242 deletion landed | `git cat-file -e` on `internal/cli/codex_factory.go` / `codex_kanban.go` → 128 (absent on develop) → **landed** |
| t1245 constants | `EnvFactoryRole = "MOAI_FACTORY_ROLE"` at internal/config/envkeys.go:323; `FactoryRoleWorker = "worker"` at :332 → **M4 branch selected** (constants untouched by M1) |
| t1193 state | `git merge-base --is-ancestor WT-factory-broker-safety develop` → exit 1 → **not landed**; six overlap files clear to edit |
| Census post-absorb | 16,124 rows at `.moai/reports/t1256/raw/census-postabsorb.tsv` (plan time: 16,608) |
| AC-RNC-023 population | 2 rows: `internal/cli/factory.go:547`, `internal/cli/kanban.go:711` (both `moai cg runs a mixed backend (leader Claude, teammates GLM)`) — M3's |
| Characterization | five packages ok on `d39a1dc09` — `.moai/reports/t1256/raw/preflight-characterization.txt`; `internal/cli` ok 988.881s with `-timeout 35m` — `.moai/reports/t1256/raw/preflight-characterization-cli.txt` |

### Freeze guards

- **AC-RNC-010 (REQ-RNC-008)** — `git merge-base develop HEAD` → `b59a5d69c`; `git diff b59a5d69c -- internal/homestate internal/factorymsg \| grep -cE '^[+-].*(CREATE TABLE\|ALTER TABLE\|CREATE INDEX\|CREATE UNIQUE INDEX)'` → **0** (this run, this tree).
- **AC-RNC-011 first clause (REQ-RNC-011)** — `grep -rnE '"MOAI_(KANBAN_LEADER\|FACTORY_LANE)' internal --include='*.go'` → **0 rows**; `git diff b59a5d69c -- internal/codexwiring/ internal/config/envkeys.go` → **empty** (env name constants AND `mcpServerEnvVarsValue` byte-identical to the merge base).

### Build and boundary

- **E2 builds** — `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (this run, tree at M1 tip; SHA recorded in §E.3).
- **E4 subagent-boundary** — `grep -rn 'AskUserQuestion' internal/kanban internal/hook internal/factorymsg internal/homestate | grep -v _test.go | grep -v '// '` → **1 row**: `internal/hook/pre_tool.go:756: if input.ToolName == "AskUserQuestion" {`. Pre-existing at the merge base (`git diff b59a5d69c -- internal/hook/pre_tool.go` → empty): a ToolName comparison, not an invocation. **0 new findings.**
- **E5 lint** — `golangci-lint run ./internal/kanban/... ./internal/hook/... ./internal/factorymsg/... ./internal/cli/... ./internal/web/...` → exit 0, `0 issues.` (`.moai/reports/t1256/raw/lint-m1.txt`).

### M1 test evidence (scoped runs, this tree)

| Package | Command | Verbatim verdict |
|---|---|---|
| internal/kanban | `go test ./internal/kanban/` | `ok github.com/modu-ai/moai-adk/internal/kanban 174.906s` |
| internal/hook | `go test ./internal/hook/` | `ok github.com/modu-ai/moai-adk/internal/hook 289.691s` |
| internal/factorymsg | `go test ./internal/factorymsg/` | `ok github.com/modu-ai/moai-adk/internal/factorymsg 47.400s` |
| internal/homestate | `go test ./internal/homestate/ ./internal/web/` | `ok ... internal/homestate 18.904s` / `ok ... internal/web 29.587s` |
| internal/cli (full) | `go test -timeout 35m ./internal/cli/` (slot `go-test-cli-hook` held) | attempt 1 FAIL (10 vocabulary-fixture tests) → fixtures updated → attempt 2 `ok github.com/modu-ai/moai-adk/internal/cli 1263.126s` (`.moai/reports/t1256/raw/cli-full-m1-attempt2.txt`) |
| internal/config, internal/codexwiring | `go test ./internal/config/ ./internal/codexwiring/` | both `ok` (freeze surfaces) |

### AC binary matrix (E1)

| AC | Status | Evidence (command → observed) |
|---|---|---|
| AC-RNC-006 | PASS | `go test ./internal/kanban/ -run 'TestRoleLeadValueIsLeader\|TestSplitLeadLabelLeaderForms'` → ok (`RoleLead=="leader"`, `LeadNumberLabel(2)=="leader-2"`, `SplitLeadLabel("leader-r7")→("r7",true)`, digits-bump edge `leader-2` covered); cli full-suite attempt 2 green covers the launcher composing these labels |
| AC-RNC-008 | PASS | kanban M1 tests: claim writes registry label `lane-1` (`TestClaimFactoryWorkerWritesLaneLabel`); broker writer emits `leader`/`lane`/`lane-1` (hook `factory_messages.go` now `kanban.RoleLead`/`kanban.RoleLane`, `TestFactoryLauncherRegistersLaunchPendingPeers` green); board guard admits `leader`/refuses `lead` (`TestBoardGuardAdmitsLeaderDeclaration`/`TestBoardGuardRefusesLegacyLeadDeclaration`); run-retire reads a `leader` peer (`TestLeadPeerIdentityReadsLeaderAndLegacyLead`); SessionStart writes role `leader` (`TestKanbanRoleFromEnvNewVocabulary`); card owner writer `kanban.RoleLead` (`TestFactoryCardOwnerWriterUsesLeaderConstant`); no written value equals `lead`/`worker`/`agent` (factorymsg `TestSendDeliversNewVocabularyRefusesLegacy` + web `ChainRoles` swap) |
| AC-RNC-009 | PASS | factorymsg: `TestResolveLaneRefusesLegacySlots`, `TestRegisterPeerBareLaneNumbersLaneOnly`, `TestSendDeliversNewVocabularyRefusesLegacy`; hook: `TestKanbanRoleFromEnvReadsOnlyLaneLabels`, `TestKanbanRoleFromEnvLegacyLabelsNotRecognized`; kanban: `TestSplitFactoryLaneLabelCanonicalOnly`, `TestLegacyFactoryLabelDetection` |
| AC-RNC-010 | PASS | schema-freeze filter over merge-base diff → **0 lines** (above); writers verified in AC-RNC-008 rows |
| AC-RNC-011 | PASS | freeze checks above (0 grep rows; envkeys+codexwiring diff empty); launch env values verified by `TestCC_FactoryEntryThroughRunCC` (`MOAI_FACTORY_WORKER="lane-2"`) and `TestLeadNameArgs_InjectsWhenUnnamed` (`--name leader`) |
| AC-RNC-022 | PASS | cli: `TestEnterSelectedFactoryRunRefusesLiveLegacyPeer` (message names `worker-2`, `runR`, `moai factory runs --retire runR`; peer+run rows untouched), `TestEnterSelectedFactoryRunDeadLegacyPeerProceeds`, `TestEnterSelectedFactoryRunLiveLegacyLeaderPeerRefuses`; hook: `TestStaleRunNoticeFactoryLegacyLabel`; factorymsg: `TestLiveLegacyPeerDetectsLiveLegacyRow`; **P3 debt**: `TestClaimFactoryWorkerIgnoresLegacyClaimOfNoRun` + `TestClaimFactoryWorkerAutoIgnoresLegacyRows` (run_id='' live legacy claim → launch proceeds, not counted as a lane) |
| AC-RNC-024 | PASS | factorymsg: `TestLeadPeerIdentityReadsLeaderAndLegacyLead` (leader peer primarily; legacy `lead` peer as identity evidence only); cli/homestate retire path: `TestEnterSelectedFactoryRunDeadLegacyPeerProceeds` + base homestate suite ok (retirable accepts `OwnerDead` unchanged); no other reader treats the `lead` peer as leader (AC-RNC-009/022 tests) |
| AC-RNC-025 | PASS | hook: `TestStaleRunNoticeLegacyLeadLabel` ((a) + debt P4: no record file created), `TestStaleRunNoticeLegacySessionRecord` ((b) record byte-identical), `TestFactoryHookPeerRefusesLegacyLabel`, `TestResolveLeadNameNoticesLegacyLeadEntry` (cli: live `lead` entry → one notice naming it + relaunch, launch proceeds as `leader`, entry untouched); kanban: `TestBoardGuardRefusesLegacyLeadDeclaration`/`TestBoardGuardAdmitsLeaderDeclaration` (test-written `DeclareRole`; `ErrNotSoleWriter` names `lead`; declaration byte-identical) |

### Coverage (E3)

`go test -cover` figures, this tree at M1 tip vs the same command on the merge-base tree (`git archive b59a5d69c` extracted to `/tmp/t1256-base`):

| Package | M1 tip | merge-base | Delta |
|---|---|---|---|
| internal/kanban | 86.4% | 86.5% | −0.1pp (the run deleted the superseded alias-collision test file; the replaced refusal paths are covered by the new M1 tests; 85% floor met) |
| internal/hook | 86.6% | 86.6% | equal |
| internal/factorymsg | 81.5% | 81.5% | equal |
| internal/web | 74.7% | 74.7% | equal |
| internal/cli | _see Gap below_ | — | — |

**Gaps (E3):** (1) `internal/cli` paired coverage did not complete — the full cli `-cover` runs were launched twice and did not finish under heavy cross-lane machine load (the delegated close-out makes the full cli suite the orchestrator's re-measure surface; a cli `-cover` figure is recorded there). (2) The merge-base kanban figure was measured in `/tmp`, where one unrelated cwd-sensitive test (`TestTempOrigin_FailsOpenOnUnresolvable`) fails as a measurement artifact; its coverage number is still computed and comparable.





## §E.3 Run-phase Audit-Ready Signal

- run_status: M1 complete (kanban / hook / factorymsg / homestate / web / config / codexwiring scoped suites green; full `internal/cli` suite green on attempt 2; lint 0 issues; both builds exit 0)
- run_complete_at: 2026-09-27
- run_commit_sha: pending-backfill-m1
- RED evidence: `.moai/reports/t1256/raw/red-m1.txt` (verbatim pre-implementation failures: 5 assertion REDs in internal/cli + compile-RED for the new kanban/hook/factorymsg APIs)
- coverage: kanban 86.4 (base 86.5) · hook 86.6 (base 86.6) · factorymsg 81.5 (base 81.5) · web 74.7 (base 74.7) — cli pair not completed under machine contention (Gap, see §E.2)
- notes:
  - M1 transient: the t1245 AC-AP-018 kanban pin limb (`internal/kanban/factory_label_pin_test.go`) is pinned to the M1 state — prefix `lane`, legacy prefixes detection-only — and carries an M4 tripwire; the full three-way equality (marker value == CLI token == prefix) is restored at M4 when `config.FactoryRoleWorker` flips to `lane` (constant untouched by M1 per delegation §C).
  - `-f worker` / `-f agent` role tokens still PARSE at M1 (minimal compile adaptation; desugars to canonical `lane-<n>` labels) — their dedicated rejection wording is M2 (REQ-RNC-003/-005/-007). Legacy LABELS on the input path are already refused by the claim naming the canonical `lane-<n>`.
  - The i18n message tables (`session_start_factory_i18n.go`, `session_start_kanban_i18n.go` nameChoices) received the minimal worker→lane token swap forced by the label machinery; the full four-locale native rewrite stays M3 (REQ-RNC-015).


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
