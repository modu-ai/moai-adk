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






## §E.3 Run-phase Audit-Ready Signal

- run_status: M3 complete (M1 + M2 + M3 of the code milestones; M4 marker flip, M5 identifier rename remain) — cli full suite green with coverage (`ok ... 1064.433s coverage: 84.0%`), hook full suite green (351.118s; cover 86.6%), kanban green (86.4%), web green (74.7%), statusline green, lint 0 issues (golangci v2.1.6), both builds exit 0
- run_complete_at: 2026-09-27
- run_commit_sha: f0ebcc82b
- RED evidence: M3 `.moai/reports/t1256/raw/red-m3-{hook,web,cli}.txt` (11 verbatim pre-implementation failures across the notice-table, legacy-leader view-model, doctor, and homonym tests); M2 `.moai/reports/t1256/raw/red-m2.txt` (27 verbatim pre-implementation refusal failures); M1 `.moai/reports/t1256/raw/red-m1.txt` (5 assertion REDs in internal/cli + compile-RED for the new kanban/hook/factorymsg APIs)
- coverage: kanban 86.4 (base 86.5) · hook 86.6 (base 86.6) · factorymsg 81.5 (base 81.5) · web 74.7 (base 74.7) · cli 84.0 (equal to M1/M2; merge-base 미측정 — Gap, §E.2)
- notes:
  - M1 transient: the t1245 AC-AP-018 kanban pin limb (`internal/kanban/factory_label_pin_test.go`) is pinned to the M1 state — prefix `lane`, legacy prefixes detection-only — and carries an M4 tripwire; the full three-way equality (marker value == CLI token == prefix) is restored at M4 when `config.FactoryRoleWorker` flips to `lane` (constant untouched by M1 per delegation §C).
  - `-f worker` / `-f agent` role tokens still PARSE at M1 (minimal compile adaptation; desugars to canonical `lane-<n>` labels) — their dedicated rejection wording is M2 (REQ-RNC-003/-005/-007). Legacy LABELS on the input path are already refused by the claim naming the canonical `lane-<n>`.
  - The i18n message tables (`session_start_factory_i18n.go`, `session_start_kanban_i18n.go` nameChoices) received the minimal worker→lane token swap forced by the label machinery; the full four-locale native rewrite stays M3 (REQ-RNC-015).


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
