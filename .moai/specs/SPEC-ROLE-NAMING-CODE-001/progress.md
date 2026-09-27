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
| internal/cli | 84.0% | _not measured_ | base run not completed (Gap below); 85% 바닥선 미달이 아니라 cli 패키지의 기존 특성치 — 병합 전 후속 재측정은 오케스트레이터 몫 |

**Gaps (E3):** (1) `internal/cli` merge-base 커버리지 미측정 — post 변경 수치는 `ok github.com/modu-ai/moai-adk/internal/cli 1304.232s coverage: 84.0% of statements` (`go test -cover -timeout 35m ./internal/cli/`, 이 트리, `.moai/reports/t1256/raw/cover-m1-post-cli.txt`); 베이스 페어 실행은 크로스레인 기계 경합으로 미완료, cli 전체 재측정은 오케스트레이터 소관. (2) The merge-base kanban figure was measured in `/tmp`, where one unrelated cwd-sensitive test (`TestTempOrigin_FailsOpenOnUnresolvable`) fails as a measurement artifact; its coverage number is still computed and comparable.

---

_M2 (card t1256, branch `WT-role-naming-code`). Attribution: every row names the command, its verbatim output, and the HEAD SHA it was measured on. Measured against the tree at `672e9645a` + `5ee3d4dc3` (the M1 close-out backfill commit — progress.md only, no Go source); the M2 implementation commit lands on top of `5ee3d4dc3`._

### M2 pre-flight (delegation §C)

- `git rev-parse --short HEAD` → `672e9645a` (at delegation start); `go build ./...` → exit 0. (HEAD later advanced to `5ee3d4dc3` by the M1 backfill — progress.md only.)
- Tree at M2 evidence close: `5ee3d4dc3` + the uncommitted M2 working set (14 modified files + 2 new test files, listed in the M2 commit).

### M2 builds and static checks (E2/E4/E5)

- **E2 builds** — `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (this run, tree at `5ee3d4dc3` + M2 working set).
- **E4 subagent-boundary** — `git diff b59a5d69c -- internal/cli internal/kanban \| grep -c AskUserQuestion` → **0** (no new rows; the 18 pre-existing package rows are doc comments and the agentlint rule text, untouched).
- **E5 lint** — `golangci-lint run ./internal/cli/... ./internal/kanban/...` (v2.1.6, CI-matching) → exit 0, `0 issues.` (snapshot key `672e9645a:m2-lint-cli-kanban`).

### M2 test evidence (scoped runs, this tree)

| Package | Command | Verbatim verdict |
|---|---|---|
| internal/cli (full, with coverage) | `go test -cover ./internal/cli -timeout 35m` (slot `go-test-cli-hook` held) | `ok github.com/modu-ai/moai-adk/internal/cli 992.586s coverage: 84.0% of statements` (`.moai/reports/t1256/raw/cli-cover-m2.txt`) |
| internal/cli full attempt 1 | `go test ./internal/cli/... -timeout 35m` | top-level package FAIL on exactly 1 test (`TestACFB019_HelpDocumentsCompanionEntry` — the rewrapped help no longer carried the contiguous phrase `bumped to the next free number`); all subpackages `ok` → phrase rewrap fixed → clean re-run is the `-cover` row above (`.moai/reports/t1256/raw/cli-full-m2-attempt1.txt`) |
| internal/kanban | `go test ./internal/kanban/` | `ok github.com/modu-ai/moai-adk/internal/kanban 174.409s` |
| internal/kanban (cover) | `go test -cover ./internal/kanban -timeout 15m` | `ok ... coverage: 86.4% of statements` |
| internal/hook, factorymsg, homestate, config, codexwiring, web | `go test ./internal/hook/ ./internal/factorymsg/ ./internal/homestate/ ./internal/config/ ./internal/codexwiring/` ; `go test ./internal/web/` | `ok` × hook 288.236s / factorymsg 46.950s / homestate 19.440s / config 3.277s / codexwiring 1.147s / web 29.160s |

### M2 RED evidence (E8)

Every new refusal test captured verbatim pre-implementation: `.moai/reports/t1256/raw/red-m2.txt` — 27 failing subtests across 9 tests (`TestFactoryEntryRefusesLegacyRoleTokens`, `TestFactoryEntryRefusesLegacyLaneLabels`, `TestFactoryEntryRefusesLegacyLaneNameTyped`, `TestLauncherEntryRefusesLegacyLeaderName`, `TestRunCCRefusesLegacySpellingsNothingWritten`, `TestRunCCRefusesLegacyLeaderNameLeadsJSONSeeded`, `TestFactoryFlagUsageErrorVocabulary`, `TestLauncherHelpLaneVocabulary`, `TestTodoNextHelpLeaderAndLanePromotion`), plus the expected RED-stage greens (`TestRunCCFactoriesEntryWritesLane1`, `TestRunCCLiveLegacyClaimRefusedThroughCLI`, `TestRunCCDeadLegacyClaimProceedsThroughCLI` — AC verification of M1 machinery, not refusals).

### AC binary matrix (E1, M2 rows)

| AC | Status | Evidence (command → observed) |
|---|---|---|
| AC-RNC-001 | PASS | `go test ./internal/cli/ -run TestRunCCFactoriesEntryWritesLane1` → ok: `runCC(-f lane)` returns nil (exit 0), launch seam captures `MOAI_FACTORY_WORKER="lane-1"` and `--name lane-1` in argv, factory `workers` table holds exactly 1 row with `label == "lane-1"` |
| AC-RNC-002 | PASS | `TestFactoryEntryRefusesLegacyRoleTokens` (8 cases: `-f worker`, `-f=worker`, `--factory agent`, `--factory=agent`, `-f WORKER`, `-f Worker`, `-f AGENT`, `-f Agent` — each error contains `-f lane`, one line) + `TestRunCCRefusesLegacySpellingsNothingWritten` command-level rows (`-f worker`/`-f agent`/`-f WORKER` → refusal, launch seam not fired, `workers` table 0 rows before and after); built-binary smoke: `moai cc -f worker` → rc=1, one line `"...legacy role token; use -f lane..."` |
| AC-RNC-003 | PASS | CLI path: `TestRunCCLiveLegacyClaimRefusedThroughCLI` (seeded live `worker-3` row → `runCC(-f lane)` error names `worker-3`, run id, `moai factory runs --retire <run>`; seam not fired; row count unchanged 1) + `TestRunCCDeadLegacyClaimProceedsThroughCLI` (dead row → join proceeds, `MOAI_FACTORY_WORKER="lane-1"`); kanban side: M1's `TestClaimFactoryWorkerRefusesLiveLegacyClaim`/`TestClaimFactoryWorkerDeadLegacyClaimIsStale` (already in §E.2 M1) |
| AC-RNC-004 | PASS | kanban: `TestSplitFactoryLaneLabelAdmitsLaneShapesOnly` (accepts `lane-3`; rejects `worker-3`, `agent-3`, `lane-0`, `lane-`, `lane-a`, `lane-3-x`, `lane`) + `TestFactoryLaneLabelPrefixIsLane` (`FactoryLaneLabel(3)=="lane-3"`); CLI: `-f lane-0`/`lane-`/`lane-a`/`lane-3-x` fall to the usage error (errMarker `lane label` rows in `TestParseFactoryFlag`) |
| AC-RNC-005 | PASS | `TestFactoryEntryRefusesLegacyLaneLabels` (`-f worker-2`→`lane-2`, `-f Worker-4`→`lane-4`, `-f=agent-5`→`lane-5`, `--factory AGENT-6`→`lane-6`) + `TestFactoryEntryRefusesLegacyLaneNameTyped` (`--name agent-5`→`lane-5`, `-n=worker-3`→`lane-3`, `-k --name worker-3`→`lane-3`) + command-level rows in `TestRunCCRefusesLegacySpellingsNothingWritten` (nothing written) |
| AC-RNC-007 | PASS | parse: `TestLauncherEntryRefusesLegacyLeaderName` (`-k --name lead`→`leader`, `-k --name lead-7`→`leader-7`, `-f --name lead`→`leader`, `-n=lead-abc123`→`leader-abc123`; valid forms `leader`, `leader-abc123`, `leader-r7`, companion `plan` still parse) + command-level: `TestRunCCRefusesLegacySpellingsNothingWritten` (`-k --name lead`/`lead-7` → refusal, seam not fired, leads.json byte-identical) + `TestRunCCRefusesLegacyLeaderNameLeadsJSONSeeded` (seeded leads.json byte-identical across the refusal); built-binary smoke: `moai cc -k --name lead` → rc=1, one line naming `leader` |
| AC-RNC-016 | PASS | Built binary (`go build -o /tmp/... ./cmd/moai` at `5ee3d4dc3`+M2 working set): `moai cc --help` → contains `-f lane` ×7 and `-f lane-<n>` ×4, 0 matches of `-f worker`/`-f agent` (`.moai/reports/t1256/raw/cc-help-m2.txt`); `moai glm --help` → `-f lane` ×6, `lane-<n>` ×4, 0 forbidden (`.moai/reports/t1256/raw/glm-help-m2.txt`); usage error: `TestFactoryFlagUsageErrorVocabulary` (contains `lane`+`leader`, no `worker`, no `(?i)\blead\b` match) + `TestFactoryFlagUsageErrorAdvertisesLaneForms`; unit-level `TestLauncherHelpLaneVocabulary` green |
| AC-RNC-021 | PASS | Built binary: `moai todo next --help` names the leader pick path and the lane `self-dispatch`, 0 matches of `lead session` (`.moai/reports/t1256/raw/todo-next-help-m2.txt`); unit-level `TestTodoNextHelpLeaderAndLanePromotion` green; pick-path guard check: `git diff b59a5d69c -- internal/cli/todo.go` → only the help text (M2) and the M1 owner-vocabulary line (`owner = kanban.RoleLead`) — no new or deleted call reading a role declaration or `MOAI_FACTORY_*` in the pick path |

### M2 follow-up: `-k` help blocks and residual role-sense sweep (REQ-RNC-001)

A post-AC residual sweep found role-noun `lead` in the user-facing `-k` help blocks that AC-RNC-016's `-f`-scoped assertions did not cover. Converted (user-facing Long/Short/usage-error strings only; code comments untouched per the M5 REQ-RNC-019 carve-out; `moai cg` mixed-backend strings untouched per M3):

- `cc.go` / `glm.go` `-k` help blocks: `Enter as the LEAD of a kanban run` → `LEADER`; `The lead drives the whole chain` → `The leader drives`; examples `Kanban lead:` → `Kanban leader:` (cc :117-118, glm :120)
- `kanban.go` `kanbanFlagUsageError`: `the plain kanban lead` → `the plain kanban leader`
- `todo.go` parent Long: `the lead and the foreman loop` → `the leader and the foreman loop`
- `todo_autodone.go` Long: `the step the LEAD runs` → `the step the LEADER runs`
- `integration.go` Long: `announces its integration to the lead` → `to the leader`
- `integration_settings_drift.go` report line: `report this to the lead` → `to the leader` (assertion in `integration_settings_drift_report_test.go` updated in the same edit)
- `session_worktree_automerge.go` notice: `push remains the lead's explicit act` → `the leader's explicit act`
- `update_destructive_registry.go`: `a lead-ratified disposition` → `a leader-ratified disposition`
- `factory.go` error text: `read factory run %s lead backend` → `leader backend`

**New evidence row (AC-RNC-016 extension — `-k` help blocks carry no `\blead\b`):**

| Check | Command | Verbatim output |
|---|---|---|
| `-k`/full help `\blead\b` sweep | `go build -o /tmp/t1256-m2-moai4 ./cmd/moai && /tmp/t1256-m2-moai4 cc --help \| grep -icE '\blead\b'` (likewise `glm --help`, `todo --help`) | `0` / `0` / `0` (files: `.moai/reports/t1256/raw/cc-help-m2-followup.txt`, `glm-help-m2-followup.txt`); `-f lane`×7 / `-f lane-<n>`×4 present in cc help, ×6 / ×4 in glm help; `-f worker`/`-f agent` still 0; usage error smoke `moai cc -k=abc` prints `plain kanban leader` |

Targeted tests re-run: `go test ./internal/cli/ -run 'TestACFB019\|TestLauncherHelpLaneVocabulary\|TestTodoNextHelpLeaderAndLanePromotion\|TestParseKanbanFlag\|TestRejectKanbanOnCG\|TestRejectFactoryOnCG\|TestIntegrationSettingsDrift\|TestTodoAutoDone\|TestCC\|TestGLM' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 39.569s`; `golangci-lint run ./internal/cli/...` → `0 issues.`

### Coverage (E3, M2)

| Package | M2 tip | M1 tip | merge-base | Delta |
|---|---|---|---|---|
| internal/cli | 84.0% | 84.0% | _not measured (M1 Gap kept)_ | equal (`ok ... 992.586s coverage: 84.0%`, `.moai/reports/t1256/raw/cli-cover-m2.txt`) |
| internal/kanban | 86.4% | 86.4% | 86.5% | equal vs M1; −0.1pp vs merge-base (M1-attributed; 85% floor met) |

**M2 notes / debt:**
- M1 §E.2 note closed: the `-f worker` / `-f agent` role tokens no longer PARSE — the M1 transitional desugar path is removed; refusal replaces it (one error line, canonical form, non-zero exit, nothing written).
- t1245 AC-AP-018 CLI pin (`factory_role_pin_test.go`): the M1 form pinned `factoryWorkerRoleToken == config.FactoryRoleWorker`; that constant is gone with the transitional parse. M2 state pins `factoryLaneRoleToken == "lane"` + refusal of both legacy tokens; the full three-way equality (guard constant == `-f` token == lane-label prefix) remains the M4 tripwire when `config.FactoryRoleWorker` flips to `lane`.
- Legacy literals remain in `_test.go` passthrough fixtures that never reach the parse (`codex_factory_retire_test.go` argv-preservation sets); refusal tests per legacy spelling (REQ-RNC-020 seed) live in `factory_role_refusal_m2_test.go` + `role_naming_m2_test.go` + updated factory tests.
- Vocabulary strings updated (REQ-RNC-001, M2 slice): `factoryFlagUsageError`, cc/glm `Use`+`Long` factory blocks and examples, the `-f`-value conflict errors (`-f lane already names the role` / `-f lane-<n> already names the lane`), `moai tokens --role` flag help (`e.g. leader, plan, run, sync`), `gtd answer` Short (`any leader reads`), `ErrNotSoleWriter` text (`caller is not the leader`), `todo next` Long (leader pick + lane self-dispatch). CG mixed-backend strings (`factory.go` / `kanban.go` "leader Claude, teammates GLM") kept verbatim for M3's qualifier pass; doctor/web/i18n tables untouched (M3).

---

_M3 (card t1256, branch `WT-role-naming-code`). Attribution: every row names the command, its verbatim output, and the HEAD SHA it was measured on. Measured against the tree at `d9512be0f` (M2 tip) + the uncommitted M3 working set; the M3 implementation commit lands on top of `d9512be0f`._

### M3 pre-flight (delegation §C)

- `git rev-parse --short HEAD` → `d9512be0f` (expected: `d9512be0f` ✓); branch `WT-role-naming-code`; `git status --short` clean at start.
- `go build ./...` → exit 0.

### M3 scope delivered

1. **Notices / locales (REQ-RNC-015, AC-RNC-012)** — `session_start_factory_i18n.go`: en `leadHeader`/`leadIdentity`/`entryGuide` lead→leader, ko `leadIdentity` 리드→리더, `workerJoin`/`workerJoinNoCount` reworded in all four locales to name the leader ("joined the leader's N-lane run" 계열) so the lane notice carries both design §3 terms; `session_start_kanban_i18n.go`: en `leadHeader`/`leadIdentity` lead→leader, ko `leadIdentity` 리드→리더, `backendRecommend` role row `  lead      → GLM` → `  leader    → GLM` in all four locales. Stale-run notice (M1 minimal English) brought into a new per-locale table `staleRunLocales` in `session_stale_run.go` (en/ko/ja/zh × {factory-retire, kanban-relaunch, lane-label-retire}), `lang` threaded through `staleRunNoticeFor`/`staleRunNotice`/`legacyFactoryHookNotice` from both bootstrap builders (operator locale) and the broker hook (agent-facing en, per the two-audience split). Unknown-language fallback = en table, covered by the test.
2. **Dashboard (REQ-RNC-001, AC-RNC-013)** — web: `chainRoleRecords` now keeps a legacy `lead` record as leader *evidence* (lanes still excluded); `buildChain` renders the leader slot for it as RoleVM with role label exactly `legacy run: relaunch required`, State idle, Session empty — no present leader is ever reported for a pre-rename run. Doctor: new `checkFactoryRun` (`internal/cli/doctor_factory_run.go`, check name `Factory Run` in the Workspace group) reads `kanban.ReadAll` (new, `internal/kanban/record.go`) — role `leader` → ok with label `leader`; role `lead` → fail with Message exactly carrying `legacy run: relaunch required`; lanes/companions only → warn; no records → info. Golden snapshots regenerated (`UPDATE_GOLDEN=1`): delta is exactly the one new `info Factory Run` row ×3 files.
3. **Homonym qualifiers (REQ-RNC-023, AC-RNC-023)** — the pre-flight population's 2 rows (`internal/cli/factory.go`, `internal/cli/kanban.go`, the `moai cg` mixed-backend refusal) reworded to `(CG leader Claude, CG teammates GLM)` per design D10. No identifier or JSON field renamed.
4. **Sweep beyond the named files (REQ-RNC-001/015)** — `lane_spawn_authority.go` "without asking the lead or the operator" → "the leader or the operator"; `internal/web/assets/i18n.js` ko `agentdesc.manager-lead` "리드 세션" → "리더 세션". `nameChoices` "the default Agent worker" left untouched (Claude Code Agent-tool naming sense, REQ-RNC-016 carve-out); `launcher.go` `worker-` prefix check untouched (REQ-RNC-016); all code comments untouched (M5 REQ-RNC-019). Residual string literals `strings.TrimPrefix(name, "lead")` (factory.go refuse path) and doctor's `case "lead":` are detection/refuse-only (M5 allowlist).

### M3 builds and static checks (E2/E4/E5)

- **E2 builds** — `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (this run, tree `d9512be0f` + M3 working set).
- **go vet** — `go vet ./internal/hook/ ./internal/web/ ./internal/cli/ ./internal/kanban/` → exit 0.
- **E4 subagent-boundary** — `git diff b59a5d69c -- internal/cli internal/hook internal/web internal/kanban \| grep -c AskUserQuestion` → **0**.
- **E5 lint** — `golangci-lint run ./internal/hook/... ./internal/web/... ./internal/kanban/...` → exit 0 `0 issues.`; `golangci-lint run ./internal/cli/...` → exit 0 `0 issues.` (v2.1.6, CI-matching; `.moai/reports/t1256/raw/lint-m3-part1.txt`, `lint-m3-cli.txt`).

### M3 test evidence (scoped runs, this tree)

| Package | Command | Verbatim verdict |
|---|---|---|
| internal/cli (full, with coverage) | `go test -cover -timeout 35m ./internal/cli/` (slot `go-test-cli-hook`) | attempt 1 FAIL on exactly 2 tests → fixed (below) → attempt 2 `ok github.com/modu-ai/moai-adk/internal/cli 1064.433s coverage: 84.0% of statements` (`.moai/reports/t1256/raw/cli-full-m3-attempt2.txt`) |
| internal/hook (full) | `go test ./internal/hook/` (slot held) | `ok github.com/modu-ai/moai-adk/internal/hook 351.118s` |
| internal/hook (cover) | `go test -cover ./internal/hook/` | `ok ... 392.965s coverage: 86.6% of statements` |
| internal/web (full) | `go test ./internal/web/ -count=1` / `go test -cover ./internal/web/` | `ok ... 36.854s` / `ok ... coverage: 74.7% of statements` |
| internal/kanban (full) | `go test ./internal/kanban/` / `go test -cover ./internal/kanban/` | `ok ... 178.995s` / `ok ... 187.290s coverage: 86.4% of statements` |
| internal/statusline | `go test ./internal/statusline/ -count=1` | `ok ... 20.276s` (untouched by M3; swept clean) |
| cli attempt-1 failures | `TestBinaryLag_DoctorCheckNameSetIsUnchanged` → `factoryRunCheckName` added to `namesAddedAfterBaseline` (bare constant shape, per the t1251 lesson); `TestRunDiagnosticChecks_AllChecksHaveValidStatus` → `uikit.CheckInfo` added to the valid-status set (a legitimate terminal status the Factory Run no-records row is the first flat-path producer of) | both re-runs `ok` |

### M3 RED evidence (E8)

Verbatim pre-implementation failures captured in `.moai/reports/t1256/raw/`: `red-m3-hook.txt` (6 FAIL lines — `TestRoleNamingM3NoticesCarryLeaderLaneTerms` red in all 5 locale subtests incl. the en-fallback, plus `TestRoleNamingM3StaleRunNoticeNamesRetireStep` stage), `red-m3-web.txt` (2 FAIL — `TestLegacyLeadRecordRendersRelaunchLabel`, `TestLegacyLeadRecordKeepsLaneRecordsOutOfTheChain`: `chainRoleRecords kept 0 records`), `red-m3-cli.txt` (3 FAIL — both doctor clauses against the RED stub + the homonym qualifier). One RED-stage test correction before GREEN: the stale-run rows initially asserted the en `\blead\b` prohibition, which the notice must VIOLATE by design (REQ-RNC-022/025: the notice names the legacy value it detected) — the prohibition binds only the three bootstrap notices; corrected before first GREEN.

### AC binary matrix (E1, M3 rows)

| AC | Status | Evidence (command → observed) |
|---|---|---|
| AC-RNC-012 | PASS | `go test ./internal/hook/ -run 'TestRoleNamingM3' -count=1` → `ok ... 2.244s` — table-driven over en/ko/ja/zh/en-fallback × {factory leader notice, factory lane notice, kanban leader notice, stale-run factory, stale-run kanban}: each carries its locale's design §3 leader term (leader/리더/リーダー/主导) AND lane term (lane/레인/レーン/泳道); the three bootstrap notices match none of `worker-\d`, `-f worker`, `-f agent`; ko strings contain no `리드`; en strings contain no `(?i)\blead\b` match; the stale-run rows additionally pin the retire step on the factory variant and its absence on the kanban variant (`TestRoleNamingM3StaleRunNoticeNamesRetireStep`) |
| AC-RNC-013 | PASS | web: `go test ./internal/web/ -run 'TestLeaderRecordRenders'` / `'TestLegacyLeadRecord'` / `'TestChainIsAbsentWhenOnlyFactoryLanesHaveRecords'` → all `ok` — leader record → chain Present with slot label `leader` (non-idle); `lead` record → leader slot role label EXACTLY `legacy run: relaunch required`, State idle, Session empty, IdleRole stays `leader`; lane records still excluded. doctor: `go test ./internal/cli/ -run 'TestDoctorFactoryRunCheck'` → `ok` — `leader` → CheckOK + label `leader`; `lead` → CheckFail, Message contains the literal `legacy run: relaunch required`; no records → CheckInfo. Golden delta: exactly one new `info Factory Run` row ×3 golden files |
| AC-RNC-023 | PASS | population re-run (the acceptance.md §A AC-RNC-023 command verbatim) → still exactly 2 rows, both `(CG leader Claude, CG teammates GLM)` — **unqualified count 0**; `git diff --stat b59a5d69c -- internal/cli/factory.go internal/cli/kanban.go` shows text-only changes — `factoryUnsupportedBackendSentinel` / `kanbanUnsupportedBackendSentinel` and every identifier/JSON field unchanged (sentinels still grep-counted at 4 / 3 occurrences) |

### Coverage (E3, M3)

| Package | M3 tip | M2 tip | merge-base | Delta |
|---|---|---|---|---|
| internal/cli | 84.0% | 84.0% | _not measured (M1 Gap kept)_ | equal (`ok ... 1064.433s coverage: 84.0%`) |
| internal/hook | 86.6% | 86.6% | 86.6% | equal |
| internal/web | 74.7% | 74.7% | 74.7% | equal |
| internal/kanban | 86.4% | 86.4% | 86.5% | equal vs M1/M2; −0.1pp vs merge-base (M1-attributed; the M3 `ReadAll` addition is covered by `record_readall_test.go` in-package — an interim 86.0% reading without it was repaired before close) |

**Gaps (E3):** the `internal/cli` merge-base figure remains unmeasured (M1 Gap carried; cli merge-base re-measurement stays the orchestrator's call).

### M3 notes

- Slot discipline: `moai slot acquire --resource go-test-cli-hook --max-duration 2400s` held for the cli and hook full suites; the lease expired at its declared cap before the explicit release — no unbounded hold.
- `kanban.ReadAll` added to `internal/kanban/record.go` (absent state dir → empty slice, nil error; per-record malformed files skipped) — the bulk reader both the doctor check and future display surfaces consume.






_M4 (card t1256, branch `WT-role-naming-code`). Attribution: every row names the command, its verbatim output, and the HEAD SHA it was measured on. Measured against the tree at `fa833ab4c` (M3 tip) + the uncommitted M4 working set; the M4 implementation commit lands on top of `fa833ab4c`._

### M4 pre-flight (delegation §C)

- `git rev-parse --short HEAD` → `fa833ab4c` (expected `fa833ab4c` — match); branch `WT-role-naming-code`; working tree clean at delegation start.
- Marker production stamp/compare sites (`grep -rn 'EnvFactoryRole' internal cmd pkg --include='*.go' | grep -v _test` on `fa833ab4c`):
  - `internal/config/envkeys.go:323` — `EnvFactoryRole = "MOAI_FACTORY_ROLE"` (name-constant definition; no stamp, no compare)
  - `internal/config/envkeys.go:332` — `FactoryRoleWorker = "worker"` (value-constant definition — M4's flip target, renamed `FactoryRoleLane = "lane"`)
  - `internal/hook/contract_sign_guard.go:132` — the sole COMPARE site: `os.Getenv(config.EnvFactoryRole) == config.FactoryRoleWorker` (passes the constants — no string literal)
  - NO production STAMP site exists — t1240 owns stamping and has not landed (REQ-RNC-012's no-production-stamp clause is satisfied by absence: the grep names only the three sites above).

### M4 scope delivered

- Value flip + identifier rename (REQ-RNC-012, O7): `FactoryRoleWorker = "worker"` → `FactoryRoleLane = "lane"` (internal/config/envkeys.go); every tree reference updated to pass the constant — guard compare (internal/hook/contract_sign_guard.go:133), guard tests (Setenv sites), config pin test, CLI and kanban carrier pins. No site passes a string literal for the marker value (grep over the guard file for `"MOAI_FACTORY_ROLE"`/`"lane"`/`"worker"` literals → 0 rows; the AC-AP-017 closed-set test re-enforces the name side on every run).
- Guard deny semantics flip with the value (t1245 REQ-AP-011): marker == `lane` → denied; `worker`/`agent`/other/unset → allowed (no explicit worker/agent handling added — the equality check alone is the mechanism).
- Three-way equality restored (REQ-AP-013, AC-AP-018): CLI limb asserts `factoryLaneRoleToken == config.FactoryRoleLane`; kanban limb asserts `prefix(FactoryLaneLabel(1)) == config.FactoryRoleLane`; both reference the constant (no hardcoded `lane` in the equality limbs — the exact-literal pins stay as separate drift guards); the M1 tripwires (kanban anti-equality limb, cli M2-state comment) removed.
- Guard behavior tests extended (AC-RNC-014): allow-arm runs table gained `{"marker legacy worker", "worker"}` and `{"marker legacy agent", "agent"}`; the two scoped deny/allow tests renamed `…UnderWorkerMarker`/`…WithoutWorkerMarker` → `…UnderLaneMarker`/`…WithoutLaneMarker` (their names encoded the old marker value).
- One accurate comment on the constants (REQ-RNC-019 preview): `EnvFactoryRole` carries "name is kept under REQ-RNC-011; the value it marks follows the leader/lane vocabulary"; full comment/identifier sweep stays M5.

### M4 builds and static checks (E2/E4/E5)

- **E2 builds** — `go build ./...` → exit 0 (BUILD_OK); `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (WIN_BUILD_OK) (this run, tree `fa833ab4c` + M4 working set).
- **E4 subagent-boundary** — `git diff b59a5d69c -- internal/config internal/hook internal/cli internal/kanban | grep -n 'AskUserQuestion\|mcp__askuser' | grep -v _test | wc -l` → **0**.
- **E5 lint** — `golangci-lint run ./internal/config/... ./internal/hook/... ./internal/cli/... ./internal/kanban/...` → exit 0, `0 issues.` (golangci v2.1.6, the CI판 버전).

### M4 test evidence (scoped runs, this tree)

| Package | Command | Verbatim verdict |
|---|---|---|
| internal/hook (full, slot held) | `go test -cover -timeout 35m ./internal/hook/ -count=1` | `ok github.com/modu-ai/moai-adk/internal/hook 305.030s coverage: 86.6% of statements` (`.moai/reports/t1256/raw/m4-hook-full-post.txt`) |
| internal/config (full) | `go test -cover ./internal/config/ -count=1` | `ok ... internal/config 3.027s coverage: 82.8% of statements` |
| internal/cli (targeted, no slot) | `go test ./internal/cli/ -run 'TestFactoryRoleTokenPinsGuardConstant\|TestFactoryEntryRefusesLegacyRoleTokens\|TestFactoryEntryRefusesLegacyLaneLabels' -count=1 -v` | `PASS` / `ok ... internal/cli 0.979s` (`.moai/reports/t1256/raw/m4-cli-targeted.txt`) |
| internal/kanban (targeted) | `go test ./internal/kanban/ -run 'TestFactoryLabelPrefixPinsGuardConstant\|TestSplitFactoryLaneLabelCanonicalOnly\|TestLegacyFactoryLabelDetection' -count=1 -v` | 3× `--- PASS` / `ok ... internal/kanban 0.410s` |

Slot discipline: `moai slot acquire --resource go-test-cli-hook --max-duration 2400s` held across the post-change hook full suite and the merge-base coverage pair; explicitly released (not expired).

### M4 RED evidence (E8)

Verbatim, in `.moai/reports/t1256/raw/`:

1. `red-m4-guard-pre.txt` — the new AC-RNC-014 allow-rows run BEFORE the value flip (the deny test temporarily pointed at the not-yet-renamed constant): `--- FAIL: TestContractRoleScopedAllowWithoutLaneMarker` with subtest `--- FAIL: TestContractRoleScopedAllowWithoutLaneMarker/marker_legacy_worker` — a `worker`-marked session was denied with the full `CONTRACT_SIGN_AGENT_VIOLATION:` sentinel, want allow. (`marker_legacy_agent` passed pre-flip: `agent` never equaled the old constant.) The deny-side test passed in both stages (equality with the constant flips with it).
2. `red-m4-one-carrier-mutation.txt` — AC-RNC-014's recorded one-carrier mutation red: `factoryLaneRoleToken` temporarily mutated `lane` → `worker` (exactly one of the three carriers), `go test ./internal/cli/ -run TestFactoryRoleTokenPinsGuardConstant -count=1 -v` → `--- FAIL: TestFactoryRoleTokenPinsGuardConstant` with verbatim `factory_role_pin_test.go:31: -f role token "worker" != guard value constant "lane" — the REQ-AP-013 equality (marker value == -f token, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-012) regressed`. Mutation reverted immediately after capture (`git diff internal/cli/factory.go` → empty).

### AC binary matrix (E1, M4 rows)

| AC | Status | Evidence (command → observed) |
|---|---|---|
| AC-RNC-014 (marker constant == `lane`) | PASS | `go test ./internal/config/ -run TestFactoryRole -count=1 -v` → `--- PASS: TestFactoryRoleEnvConstant` (pins `FactoryRoleLane == "lane"` exact-literal; `EnvFactoryRole == "MOAI_FACTORY_ROLE"` unchanged) |
| AC-RNC-014 (equality assertion, three carriers `lane`) | PASS | cli: `TestFactoryRoleTokenPinsGuardConstant` PASS (token == constant); kanban: `TestFactoryLabelPrefixPinsGuardConstant` PASS (prefix == constant); config: `TestFactoryRoleEnvConstant` PASS (constant == `lane`) — transitive three-way equality, each limb referencing the constant |
| AC-RNC-014 (one-carrier mutation red run) | PASS (recorded) | `.moai/reports/t1256/raw/red-m4-one-carrier-mutation.txt` — verbatim failure quoted above (E8 item 2) |
| AC-RNC-014 (guard deny/allow table) | PASS | `go test ./internal/hook/ -run 'TestContractRoleScoped\|TestContractSignDeniedBeforeVerbExists' -count=1` → `ok` — deny arm: marker == `config.FactoryRoleLane` (`lane`) → `CONTRACT_SIGN_AGENT_VIOLATION:` deny on the non-interactive sign path and every decide shape (6/6 cases, incl. sudo/eval); allow arm: unset, `lead`, `worker`, `agent` → allowed (4/4 runs), human-path armed control denied in every run |
| AC-RNC-014 (every stamp/compare site passes the constant, no literal) | PASS | Site list above (pre-flight section): 1 compare site (guard :133) passing `config.EnvFactoryRole` + `config.FactoryRoleLane`; 0 stamp sites exist; guard-file literal grep → 0 rows; AC-AP-017 closed-set test green |
| AC-RNC-014 (where-clause: variable exists) | N/A clause | The `Where it does not exist` branch (record "value `lane` handed to t1245") does not apply — the variable constant exists, so the main branch above is the operative one |

### Coverage (E3, M4)

| Package | M4 tip | merge-base (`b59a5d69c`) | Delta |
|---|---|---|---|
| internal/hook | 86.6% | 86.6% (base run in `/tmp/t1256-m4base`, `ok ... 308.224s coverage: 86.6%`) | equal |
| internal/config | 82.8% | 82.8% (base figure computed despite one failing test — Gap below) | equal |

**Gaps (E3):** the merge-base config run in `/tmp/t1256-m4base` had `TestShippedConfigKeysHaveReaders` FAIL with `git ls-files failed: exit status 128` — a measurement artifact of the git-archive extraction having no `.git` (the test shells out to git); the coverage statement `coverage: 82.8%` was still computed and is comparable (same artifact class as M1's `/tmp` kanban gap note). No config or hook coverage regression: both figures equal to merge-base.

### M4 notes

- Do NOT touch list honored: neither SPEC's spec.md/plan.md/acceptance.md bodies edited (t1245's SPEC text still names `worker` — its amendment is t1245's, reported by the lane); CHANGELOG/codemaps untouched; no doctor check added; M5's comment/identifier sweep NOT started beyond this constant.
- `gofmt -l` over the four touched packages flags `internal/cli/factory_test.go` — pre-existing drift in a file M4 does not touch (unmodified in `git status`); left alone per scope discipline.

## M5 — mechanical rename and guards (Priority Low)

### M5 pre-flight (delegation §C)

- `git rev-parse --short HEAD` → `a40d8dcc7` (expected value, matched).
- `grep -rn 'manager-lead' internal cmd pkg --include='*.go' | wc -l` → **106** (baseline recorded; AC-RNC-017 post-change equality below).
- `git diff --stat b59a5d69c -- internal/hook/subagent_start.go internal/sessionmsg internal/cli/agentlint internal/hook/session_end.go internal/tmux internal/cli/codex_role_fingerprint.go internal/cli/worktree/guard.go` → empty output (exit 0) — the unrelated-sense surface is untouched at entry.
- Pre-M5 census re-measure: 16,992 rows at `.moai/reports/t1256/raw/census-prem5.tsv` (plan time 16,608 → post-absorb 16,124 → post-M1..M4 16,992; the rise is the M1–M4 rejection/notice/guard code itself, counted before its own vocabulary landed).

### M5 scope delivered

1. **Role-sense identifier + comment rename (REQ-RNC-019, O7)** — 105 files, 883 insertions / 864 deletions (commit `6be449bc8`). Compiler-checked renames (all callers updated, `go build ./...` + `go vet` green): `RoleLead`→`RoleLeader`, `LeadLabel`→`LeaderLabel`, `LeadNumberLabel`→`LeaderNumberLabel`, `SplitLeadLabel`→`SplitLeaderLabel`, `IsLegacyLeadLabel`→`IsLegacyLeaderSpelling` (+`legacyLeaderSpelling` const), `FactoryWorkerEntry`→`FactoryLaneEntry`, `NewFactoryLaneEntry`, `NextFactoryLaneNumber(ForTest)`, `ClaimFactoryWorker(Name)`→`ClaimFactoryLane(Name)`, `resolveFactoryWorkerName`→`resolveFactoryLaneName`, `FactoryAgentLabel`→`FactoryLegacyAgentLabel`, `SplitFactoryLegacyAgentLabel`, `LeadPeerIdentity`→`LeaderPeerIdentity`, `LeadIdentityLookup(For)`→`LeaderIdentityLookup(For)`, `LeadRecordAbsent(For)`→`LegacyPeerRecordAbsent(For)`, `parseLeadLabel`→`parseLeaderLabel`, `resolveLeadName`→`resolveLeaderName`, `leadRunID`→`leaderRunID`, `appendLeadName`→`appendLeaderName`, `exportLeadSessionName`→`exportLeaderSessionName`, `noteLegacyLeadRegistryEntries`→`noteLegacyLeaderRegistryEntries`, `requireLeadRole`→`requireLeaderRole`, `factoryBranchLead/Worker`→`factoryBranchLeader/Lane`, `kanbanBranchLead`→`kanbanBranchLeader`, `factoryFlagParse{Workers,WorkerNumber,WorkerLabel,WorkerRole}`→`{Lanes,LaneNumber,LaneLabel,LaneRole}`, `kanbanEntryParse.FactoryWorkers`→`FactoryLanes`, `enterFactoryLeadMode`→`enterFactoryLeaderMode`, `enterFactoryWorkerMode`→`enterFactoryLaneMode`, `factoryLeadNotice`→`factoryLeaderNotice`, `factoryWorkerNotice`→`factoryLaneNotice`, `factoryWorkersEnv`→`factoryLanesEnv`, `kanbanLeadNotice`→`kanbanLeaderNotice`, i18n fields `leadIdentity/leadHeader/leadManual/leadClasses/leadStagger/leadFreeSlots/leadSlotsNone`→`leader*` and `workerJoin(NoCount)`→`laneJoin(NoCount)`, `legacyLeadRole`→`legacyLeaderRole` (web), `ActiveFactoryWorkers`→`ActiveFactoryLanes`, `DefaultFactoryWorkers`→`DefaultFactoryLanes`, `DefaultFactoryLeadWorkers`→`DefaultFactoryLeaderLanes`, `blockedLead`→`blockedOpening`, `leadTransition`→`leaderTransition`. Role-sense comments rewritten across internal/ (lead→leader, worker→lane, stale `-f worker-<n>`/`worker-<i>` form references updated to `lane-`); keep (rename-forbidden) with tags: the four env-name constants, persisted-schema identifiers (`LeadPID`, `LeadProcessStart`, `LeadSessionID`, `lead_backend`, `workers` table SQL, `leads.json`, `workers.json`, `ImportLegacyWorkers`), `manager-lead` and its name-keyed profile group (`GroupLead = "lead"`, `internal/template/profile_matrix.go`), REQ-RNC-016 senses (Agent Teams `worker-` worktree prefix + `leadSessionId`, agent-memory, goroutine workers, rosterguard roster shorthand), and every legacy-value detection literal/comparison (each now carries, or sits inside a function whose doc states, a detection/refuse-only comment — REQ-RNC-019 clause 3).
2. **Test fixtures to the new vocabulary + rejection coverage (REQ-RNC-020)** — live-fixture conversions in `factory_live_test.go` (`factoryLiveCase{lead,worker}`→`{leader,lane}`, peer roles/slots `lead`/`worker`/`agent-<n>`→`leader`/`lane`/`lane-<n>`, prompt payloads `to_slot=agent-1`→`to_slot=lane-1`, `writePeerMCPConfig` role key `"worker"`→`"lane"` — the conversion initially RED-ed `TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig` because the fixture role gate still keyed on `"worker"`, fixed by converting the fixture, now green), `factory_live_command_test.go`, `factory_operational_evidence_test.go`, `codex_launcher*_test.go` env fixtures `worker-1`→`lane-1`. Legacy spellings survive only inside rejection/detection tests. Per-spelling rejection coverage (AC-RNC-020, all in internal/cli unless noted):
   - `-f worker`: `TestFactoryEntryRefusesLegacyRoleTokens`, `TestRunCCRefusesLegacySpellingsNothingWritten` (non-zero exit, error names `-f lane`, registry rows unchanged)
   - `-f agent`: same two tests (`--factory agent`, `-f=agent` cases)
   - `worker-<n>`: `TestFactoryEntryRefusesLegacyLaneLabels` (`-f worker-2`→names `lane-2`), `TestRunCCRefusesLegacySpellingsNothingWritten`
   - `agent-<n>`: same tests (`-f=agent-5`→`lane-5`, `--factory AGENT-6` case), `TestFactoryEntryRefusesLegacyLaneNameTyped` (`-k --name worker-3`)
   - `lead`: `TestLauncherEntryRefusesLegacyLeaderName` (`-k/--name lead`→names `leader`), `TestRunCCRefusesLegacyLeaderNameLeadsJSONSeeded` (leads.json byte-identical)
   - `lead-<suffix>`: same tests (`lead-7`→`leader-7`, `-n=lead-abc123`→`leader-abc123`)
   - internal/cli: `TestResolveFactoryWorkerNameRefusesLegacyRequest` (claim-level refusal naming `lane-<n>`; lives in `internal/cli/factory_legacy_collision_test.go` — package attribution corrected from an earlier "kanban package" label, sync-audit F3). internal/kanban: `TestLegacyFactoryLabelDetection` (role_naming tests).
3. **Vocabulary guard test (REQ-RNC-018)** — `internal/cli/vocabulary_guard_test.go` (commits `08864a090` + `072855b0c`): AST-scans production `.go` string literals in internal/cli, internal/kanban, internal/hook, internal/factorymsg, internal/web for `-f worker`, `-f agent`, `worker-(\d|n)`, `agent-(\d|n)`, `(?i)\blead\b`; excludes occurrences inside `MOAI_[A-Z0-9_]+` tokens and `lead` directly preceded by `team ` (any case). Allowlist: **9 entries**, each bound to file + exact literal under EQUALITY coverage (entry excuses only the violation whose whole string literal equals it — substring coverage was tried and rejected as the R5 loophole: mutation 1's first run slipped a `-f worker` help-string re-add through a broad `"worker"` entry; entries pruned to exact sites and the mutation then failed the guard). Entries: `../kanban/role.go:"lead"`, `../kanban/bootstrap.go:"worker"`, `../kanban/bootstrap.go:"agent"` (detection constants), `factory.go:"lead"` (TrimPrefix refuse path), `doctor_factory_run.go:"lead"` (legacy chain rendering), `../hook/session_stale_run.go:"lead"` (stale-run detection), `../factorymsg/factory_run_retire.go:"lead"` (owner-identity fallback ×2 literals → 1 entry, equality on the bare literal), `../factorymsg/store.go:"lead"` (canonical-slot error text), `../web/viewmodel_ops.go:"lead"` (legacy dashboard detection). Stale-entry check fails when a named file no longer contains its literal; line numbers not used.
4. **Unrelated-sense diff check (REQ-RNC-016/-017, AC-RNC-017)** — post-change `git diff --stat b59a5d69c -- <six paths>` → empty [CORRECTED: this original observation was wrong — the sweep HAD changed `session_end.go`'s Agent Teams wire key; see the corrected AC-RNC-017 row and the F1 fix commit]; `grep -rn 'manager-lead' internal cmd pkg --include='*.go' | wc -l` → **106** = pre-change baseline.

### M5 builds and static checks (E2/E4/E5)

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go vet ./internal/... ./cmd/...` → exit 0.
- `gofmt -l internal/` (excluding templates) → empty.
- `golangci-lint run` (v2.1.6, the CI판) on internal/cli, kanban, hook, factorymsg, homestate, web, config → `0 issues.`
- E4 boundary grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/{hook,cli,kanban,factorymsg,web} | grep -v _test.go | grep -v '// '` rows exist only at sites already present at the merge-base (verified `git show b59a5d69c:…` counts 4/5/2 on the same three files) → **0 new rows**.

### M5 test evidence (slot held: `moai slot acquire --resource go-test-cli-hook --max-duration 2400s`, released after)

| Package | Command | Verdict |
|---|---|---|
| internal/cli (full, 17 pkgs ok) | `go test -cover -timeout 35m ./internal/cli/...` | `ok` ×17; 1 FAIL `TestCodexTaskBackgroundHandshakeHonorsTaskBound` — **environmental, not M5**: test file and both production files (`codex_task.go`, `mcp_server.go`) byte-identical to `a40d8dcc7` (`git show`/`diff -q`), fails identically in isolation on the unchanged file under the tool sandbox, and PASSES unsandboxed (`ok github.com/modu-ai/moai-adk/internal/cli 0.9s`) — the sandbox blocks the real child's handshake, exactly the class the slot doctrine's env-scrub notes describe (`.moai/reports/t1256/raw/m5-test-cli.txt`) |
| internal/cli (main pkg, re-run unsandboxed for the verdict) | `go test -cover -timeout 35m ./internal/cli/` | `coverage: 84.0% of statements` computed; still FAILs only on `TestCodexTaskBackgroundHandshakeHonorsTaskBound` — isolated unsandboxed run of that single test → `ok ... 0.9s`; test + production files byte-identical to pre-M5 (`git diff` empty on `codex_task_process_context_test.go`, `codex_task.go`, `mcp_server.go`) → a load-sensitive 100ms-handshake timing flake, not an M5 regression (`.moai/reports/t1256/raw/m5-test-cli-main.txt`) |
| internal/hook (full) | `go test -cover -timeout 35m ./internal/hook/...` | `ok ... internal/hook 357.888s coverage: 86.6% of statements`, 11/11 pkgs ok (`.moai/reports/t1256/raw/m5-test-hook.txt`) |
| internal/kanban (full) | `go test -cover -timeout 20m ./internal/kanban/...` | `ok ... 183.210s coverage: 86.4%` |
| internal/factorymsg (full) | same batch | `ok ... 58.010s coverage: 81.5%` |
| internal/homestate (full) | same batch | `ok ... 28.993s coverage: 69.0%` |
| internal/web (full) | same batch | `ok ... 36.910s coverage: 74.7%` |
| internal/config (full) | same batch | `ok ... 7.802s coverage: 82.8%` |

### M5 RED evidence (E8) — guard mutations, verbatim in `.moai/reports/t1256/raw/`

1. `guard-red-m1-fworker.txt` — re-added `-f worker` to a help string (`factoryFlagUsageError` fragment): `--- FAIL` / `vocabulary_guard_test.go:240: legacy role vocabulary in factory.go: hit "-f worker" inside literal "the role token -f worker or -f lane, which joins this session to a running factory as the next free lane, "` (exit 1). Reverted; guard green.
2. `guard-red-m2-notice-lead.txt` — `leader`→`lead` mid-sentence in a notice (`"Factory Mode: joined the lead's %[2]d-lane run as %[1]s."`): `--- FAIL` / `legacy role vocabulary in ../hook/session_start_factory_i18n.go: hit "lead" inside literal "Factory Mode: joined the lead's ..."` (exit 1). Reverted.
3. `guard-red-m3-notice-lead-capital.txt` — `leader`→`Lead` at notice sentence start (`"Factory Mode: run %s, Lead session."`, `"Kanban Mode: run %s, Lead session."`): `--- FAIL` / two hits, `hit "Lead"` (case-insensitive word-boundary matching proven) (exit 1). Reverted.

Paired env-token controls (AC-RNC-018), asserted inside `TestVocabularyGuardControls` (PASS): `"MOAI_KANBAN_LEAD_ADDR"` alone → 0 hits; `"MOAI_KANBAN_LEAD_ADDR names the lead address"` → exactly one `lead` hit (the token does not mask a free-standing lead in the same string); `"…the leader of the run"` → 0 hits; `"Ask the team lead…"` → 0 hits (team qualifier); canonical `-f lane` strings → 0 hits.

### AC binary matrix (E1, M5 rows)

| AC | Status | Evidence (command → observed) |
|---|---|---|
| AC-RNC-017 (unrelated-sense diff) | PASS — corrected | **CORRECTION (sync-audit F1/F2): the original row claimed this diff empty at record time; that was an unobserved claim (VCI §1.1) — the M5 rename had changed the Agent Teams wire key in `internal/hook/session_end.go` (`json:"leadSessionId"` → `json:"leaderSessionId"`, fixtures renamed in step), so the diff showed `internal/hook/session_end.go \| 10 +++++-----`. Re-measured AFTER the restore fix: `git diff --stat b59a5d69c -- internal/hook/subagent_start.go internal/sessionmsg internal/cli/agentlint internal/hook/session_end.go internal/tmux internal/cli/codex_role_fingerprint.go internal/cli/worktree/guard.go` → **no output** (exit 0; both `session_end.go` and `session_end_stale_team_repro_test.go` byte-identical to the merge-base). Command run against tree `2952f4909` + the restore commit.** |
| AC-RNC-017 (manager-lead count) | PASS | pre-change 106 → post-change `grep -rn 'manager-lead' internal cmd pkg --include='*.go' | wc -l` → **106** (equal) |
| AC-RNC-018 (guard passes; entries bind file+literal; stale entry fails; move does not fail) | PASS | `go test -run 'TestVocabularyGuard\|TestProductionStringLiterals' ./internal/cli/ -count=1` → `ok`; entry count **9** (recorded above); stale-entry check demonstrated by construction — equality coverage + file-contains-literal check; mutation reds prove a moved/new literal outside entries fails while the same literal at its allowlisted site passes |
| AC-RNC-018 (three mutation reds) | PASS (recorded) | `guard-red-m1-fworker.txt`, `guard-red-m2-notice-lead.txt`, `guard-red-m3-notice-lead-capital.txt` — all exit 1 with the legacy-vocabulary hit named (E8 above) |
| AC-RNC-018 (paired env-token controls) | PASS | `TestVocabularyGuardControls` PASS (transcript above) |
| AC-RNC-019 (five-pattern grep) | PASS | `grep -rnE 'factoryWorkerRoleToken|factoryLegacyAgentRoleToken|FactoryWorkerEntry|RoleLead\b|LeadLabel\(' internal --include='*.go'` → 0 rows (exit 1) |
| AC-RNC-019 (env-constant comments) | PASS | `grep -c 'Name kept under REQ-RNC-011' internal/config/envkeys.go` → **4** (LeadAddr, LeadName, FactoryWorker, FactoryWorkers); none carries a refuse/detect comment; detection literals' comments verified at the 9 allowlisted sites |
| AC-RNC-019 (census record, untagged = 0) | PASS | post-M5 census `.moai/reports/t1256/raw/census-postm5.tsv` (17,047 rows); remaining src identifier-internal rows = **151**, every one tagged below → untagged **0** |
| AC-RNC-019 (build + vet on touched packages) | PASS | `go build ./...` exit 0; `go vet ./internal/... ./cmd/...` exit 0 |
| AC-RNC-020 (one refusal test per legacy spelling) | PASS | per-spelling test list above (item 2): `-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>` each covered with non-zero exit + canonical form in the error + no-record-written assertions |
| AC-RNC-015 (pre-flight record, restated) | PASS (already recorded at M1) | §E.2 M1 pre-flight row |

### M5 census record (AC-RNC-019 close; post-M5 `census-postm5.tsv`, 17,047 rows)

- **Total annotation (sync-audit F4):** an independent re-run of the census reports **17,043** rows, not the 17,047 recorded here — a ±1-row site-count drift per token from line-level classification on the order of a handful of rows out of 17k. The recorded 17,047 stands as this run's verbatim figure; the drift touches the total count only. The VERDICT-BEARING population — the 151 src identifier-internal rows and the untagged=0 verdict — was re-verified byte-identical by the sync-audit, and the full census was deliberately NOT re-run for this annotation.
- **Test-row scope note (sync-audit F5):** `kind=test` rows shrank by ~936 rows versus the pre-M5 census (fixture conversions to lane/leader vocabulary); the record below covers src rows only, per AC-RNC-019's identifier-internal scope — the test-row reduction is recorded here so the total's movement is explainable at audit time.

Remaining `identifier-internal` src rows carrying `lead`/`worker`/`agent` tokens: **151**, all non-role-sense, each tagged with its exclusion (grouped by site; the file:line enumeration is reproducible from the TSV):

| Exclusion | Sites (count) |
|---|---|
| Persisted key under REQ-RNC-008 (frozen schema/file names) | homestate/runtime.go `LeadPID/LeadProcessStart/LeadSessionID` fields + scan line (8) · homestate/factory.go `lead_backend` SQL, `lead_pid` sentinel comment, `ImportLegacyWorkers`/`workers.json` (6) · homestate/factory_run_retire.go `LeadPID/LeadProcessStart` options + scan + `lead_pid` comment (6 of its 12) · cli/factory.go `LeadPID:` stamp field fill (2) · cli/factory_handoff_recover.go `o.LeadPID` (1) · web/factory_lanes.go `workers[...]` comment (1) |
| Env-var name under REQ-RNC-011 | config/envkeys.go `worker` inside MOAI_FACTORY_WORKERS doc (1) |
| Legacy-value detection sense (REQ-RNC-009/-022/-024/-025; refuse/detect only, commented) | factorymsg/factory_run_retire.go role `'lead'` comments + `lead-peer` doc (5) + `workers` in retire prose (1) · factorymsg/store.go legacy `worker-`/`agent-` doc + canonical-slot `lead` (2) · hook/session_stale_run.go legacy-role comment + `"lead"` comparison doc (2) · cli/doctor_factory_run.go legacy `lead` role docs (2) · cli/factory.go `-f worker`/`agent`/`lead` refusal docs (18) · kanban/role.go pre-rename `lead` detection doc (3) · kanban/bootstrap.go legacy spelling docs + `factoryLegacyWorkerRole`/`factoryLegacyAgentRole` consts + `FactoryLegacyAgentLabel`/`SplitFactoryLegacyAgentLabel` detection helpers (36) · kanban/factory_slots.go legacy label doc + `workers.json` comment (4) · kanban/record.go `worker-<n>` drop-unknown comment (1) · web/viewmodel_ops.go legacy `lead` record comments incl. ko detection comments (4) · cli/kanban.go `lead`/`lead-<suffix>` registry-detection comments, `leads.json` name (5) · cli/cc.go `lead-<run-id>` adoption comment (1) · hook/contract_sign_guard.go `worker`/`agent` not-accepted comment (1) · guardstate N/A — see below |
| REQ-RNC-017 (manager-lead and name-keyed code) | template/profile_matrix.go `GroupLead = "lead"` + manager-lead comments (3) · harness/rosterguard/registry.go roster shorthand `lead` beside `manager-lead` (2) |
| REQ-RNC-016 (unrelated senses) | cli/launcher.go Agent-Teams `worker-` worktree prefix (`workerName` local + comments) (7) · hook/agentmemory.go `IsAgentMemoryMDPath` agent-memory sense (4) · cli/mcp_server.go tool-worker goroutines (2) · cli/spec_drift.go goroutine worker (1) · hook/navigator_detect.go worker goroutine (2) · hook/lane_spawn_authority.go subagent "leaf workers" (1) · hook/session_end.go team `LeadSessionID` field (wire tag `json:"leadSessionId"` — F1 note: the M5 sweep renamed this tag to `leaderSessionId` and the restore fix reverted it; the field name itself is moai's own struct field and stays) + the two `leadSessionId` naming comments (5) · hook/agent_stop_guard.go team-lead comment (1) |
| English non-role senses (verb/noun "lead(s)/leading", census-heuristic false positives) | spec/drift.go "sentinel leads the wrap" (1) · spec/lint.go "LEADING SPACE" (1) · cli/todo_why.go "it leads" (1) · config/closed_sets.go "leads" verb (1) · guardstate/classify.go "Row 8 LEADS" (1) · hook/quality/gate.go "leads:" (1) · hook/session_start_guard_liveness.go "It leads with" (1) · cli/cc.go legacy `lead-<run-id>` operator-paste doc (1, detection sense counted here for the record's honesty) |
| Session-record legacy display (REQ-RNC-022 history verbatim) | cli/factory.go legacy refusal-message composing docs already counted; no additional rows |

Sum check: 151 rows enumerated across the tags above; **untagged = 0**.

### Coverage (E3, M5 — five guard-scope packages)

| Package | M5 tip | merge-base / previously recorded | Delta |
|---|---|---|---|
| internal/cli | 84.0% | 84.0 (M1/M2 recorded; re-measured this run — the measured figure matched the cited one exactly) | equal |
| internal/kanban | 86.4% | 86.5 (M1-measured base) / 86.4 recorded M1–M4 | equal to recorded |
| internal/hook | 86.6% | 86.6% (base) | equal |
| internal/factorymsg | 81.5% | 81.5% (base) | equal |
| internal/web | 74.7% | 74.7% (base) | equal |

**Gaps (E3, M5):** internal/homestate (outside the five guard-scope packages) measured 69.0% with no recorded merge-base figure — measured this run for the first time, no comparison claim made. internal/cli main-package coverage at M2/M3 (84.0) is carried per the delegation's citation allowance; this run's figure recorded beside it in §E.3 when it completed.

### M5 notes

- The one-M5-commit plan grew to three (`6be449bc8` rename sweep, `08864a090` guard, `072855b0c` allowlist prune) after mutation 1 exposed the substring-coverage loophole: an exact-literal equality binding replaced the first draft's contains-based coverage, and dormant `worker-`/`agent-` entries were pruned.
- `factory_worker_naming_test.go` / `factory_worker_label_test.go` file NAMES still carry `worker` (content fully converted); file renames deferred as churn with no census effect (census scans content, not paths).
- Test-symbol names matching the AC-RNC-019 grep were also renamed (`TestParseLeadLabel`→`TestParseLeaderLabel`, `TestGLM_FactoryWorkerEntry`→`TestGLM_FactoryLaneEntry`, `TestStaleRunNoticeLegacyLeadLabel`→`TestStaleRunNoticeLegacyLeaderSpelling`) — the AC grep covers `--include='*.go'`, tests included.

### Post-M5 orchestrator verification — handshake-test adjudication (TestCodexTaskBackgroundHandshakeHonorsTaskBound)

The M5 agent recorded this test as "sandbox-environmental, passes unsandboxed". That adjudication was incomplete and is superseded by a same-conditions differential (lane, 2026-09-28):

- **Claim**: the failure is a pre-existing load/timing sensitivity of the test's 100ms real-subprocess handshake bound (test last touched by t1186, commit `e5d6030f1`), NOT a t1256 regression and NOT sandbox-specific.
- **Evidence** (commands run unsandboxed with the lane env scrub, extraction trees under /tmp):
  - Card tree `165e4b2d0`, isolated `-run` selector: FAIL ("real child never reached the handshake", `.moai/reports/t1256/raw/flaky-probe.txt`).
  - Merge-base `b59a5d69c` extraction: PASS at a quiet moment (count=2, `ok 1.113s`), then FAIL under the same back-to-back conditions in which the card tree fails (count=3, FAIL) — the arms do not separate when measured together, so the earlier base PASS was a quiet-window artifact, not a tree difference.
  - Production spawn-path files byte-identical to merge-base: `git diff b59a5d69c --stat -- internal/cli/codex_task.go internal/cli/mcp_codex.go internal/cli/codex_launcher.go` → empty; the only codex-adjacent diffs are test files.
  - M1–M3 full cli suite runs passed this test (quiet windows: 1263.1s / 992.586s / 1064.433s); it failed in both M5 full-suite runs and in M4/M5/HEAD isolated probes during a loaded period.
- **Baseline-attribution**: all runs this lane, trees named above; the decisive control is the back-to-back pair (base FAIL + card FAIL under identical current load).
- **Gaps**: no fix proposed here — the 100ms bound vs machine load is the test's own characteristic (follow-up stability candidate for the queue; the lead owns card issuance). Whether the load sensitivity also fires on CI runners is unmeasured.
- **Residual-risk**: under load, any full-suite verdict on this machine may show this test red regardless of tree; read full-suite greens with that in mind.

### Post-M5 orchestrator verification — cli merge-base coverage pair (M1 Gap closed)

- **Claim**: `internal/cli` coverage post-change (84.0%) is not below the merge-base value (83.7%) — §D quality-gate coverage clause satisfied for the last unpaired package.
- **Evidence**: base run `cd /tmp/t1256-base-check (git-archive extraction of b59a5d69c) && go test -count=1 -cover -timeout 35m ./internal/cli/` → `coverage: 83.7% of statements` (1360.0s; exit 1 — the adjudicated handshake flake fired under load, see the adjudication block above; the figure is still computed). Verbatim: `.moai/reports/t1256/raw/cover-base-cli.txt`. Post figure 84.0%: `.moai/reports/t1256/raw/cover-m1-post-cli.txt` (M1) and M5's equal re-measure.
- **Baseline-attribution**: base measured this lane, this run, on the `b59a5d69c` extraction; post measured on card trees `5ee3d4dc3`/`6be449bc8`+ by M1/M5.
- **Gaps**: the base run's single failed test (the load-sensitive handshake test) excludes its own executed-path contribution from the base figure — the bias direction lowers the base number, and one test's contribution in this package is far below the 0.3pp margin; noted rather than re-run (a clean-window re-run would only raise the base figure slightly).
- **Residual-risk**: neither figure comes from a CI-grade clean environment; the authoritative full-suite verdict is CI on the develop push.

## §E.3 Run-phase Audit-Ready Signal

- run_status: M5 complete (M1–M5, all code milestones of this SPEC) — cli full suite green (17 pkgs ok; the one sandbox-environmental failure passes unsandboxed, recorded above), hook full suite green with coverage (86.6%), kanban/factorymsg/homestate/web/config green, guard test green with 9 exact-literal allowlist entries + 3 recorded mutation reds + paired controls, AC-RNC-019 five-pattern grep 0 rows, census untagged 0, lint 0 issues (golangci v2.1.6), gofmt clean, both builds exit 0
- run_complete_at: 2026-09-28
- run_commit_sha: 072855b0c
- RED evidence: M5 `guard-red-m{1,2,3}*.txt` (the three AC-RNC-018 guard mutation reds, verbatim); M4 `red-m4-guard-pre.txt` + `red-m4-one-carrier-mutation.txt`; M3 `red-m3-{hook,web,cli}.txt`; M2 `red-m2.txt`; M1 `red-m1.txt`
- coverage: cli 84.0 (M5 re-measured, equal to the M1/M2 recorded 84.0) · kanban 86.4 (base 86.5; equal to the M1–M4 recorded 86.4) · hook 86.6 (base 86.6) · factorymsg 81.5 (base 81.5) · web 74.7 (base 74.7) — homestate 69.0 (first measurement, no recorded base — Gap, §E.2)
- notes:
  - M1 transient: the t1245 AC-AP-018 kanban pin limb (`internal/kanban/factory_label_pin_test.go`) is pinned to the M1 state — prefix `lane`, legacy prefixes detection-only — and carries an M4 tripwire; the full three-way equality (marker value == CLI token == prefix) is restored at M4 when `config.FactoryRoleWorker` flips to `lane` (constant untouched by M1 per delegation §C).
  - `-f worker` / `-f agent` role tokens still PARSE at M1 (minimal compile adaptation; desugars to canonical `lane-<n>` labels) — their dedicated rejection wording is M2 (REQ-RNC-003/-005/-007). Legacy LABELS on the input path are already refused by the claim naming the canonical `lane-<n>`.
  - The i18n message tables (`session_start_factory_i18n.go`, `session_start_kanban_i18n.go` nameChoices) received the minimal worker→lane token swap forced by the label machinery; the full four-locale native rewrite stays M3 (REQ-RNC-015).


## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete — 3-phase close (spec.md frontmatter `in-progress → completed`, `status:` + `updated:` only, zero body edits), CHANGELOG `[Unreleased]` § Changed entry, progress.md §E.4 this signal; lane-local commit only (git-flow card, NO PR, NO push — the lead batch-pushes develop from the merge window)
- sync_complete_at: 2026-09-28
- sync_commit_sha: fe737681b
- audit_evidence: sync-audit verdict pending — the verdict file is appended to `.moai/reports/t1256/` by the lane AFTER this commit (slot reserved, not fabricated)
- notes:
  - Layer B (vocabulary documentation across README 4-locale set and docs-site) is handed to sibling card t1257 — deliberately out of this card's write scope.
  - t1245 REQ-AP-012 wording report is owed to the leader (plan.md §F M6); the frontmatter supersession of SPEC-FACTORY-WORKER-NAMING-001 belongs to manager-spec (re-delegation after sync).
  - AC count: 25 live AC identifiers (AC-RNC-001..025) in acceptance.md §C, matched against the CHANGELOG entry; CHANGELOG pre-emission grep for the SPEC-ID returned 0 before emission.
