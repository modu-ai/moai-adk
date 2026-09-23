# progress.md — SPEC-FACTORY-WORKER-NAMING-001 (card t1085)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-22
- plan_audit: iter-1 PASS 0.9375 (commit 5a96704e2) → D1/D2/D6 fix (commit 08f22109c) → iter-2 confirm PASS
- artifacts: spec.md, plan.md, acceptance.md, research.md (card-mandated Tier M + research addendum), this skeleton
- baseline: HEAD 3f3ffbb57, branch WT-worker-rename

## §E.2 Run-phase Evidence

### M1 — GTD todo naming-axis closure record (2026-09-22, HEAD cb74ebd58 tree)

- AC-001 PASS — the closure note `gtd-todo-naming-closure.md` (committed 08f22109c) was read this run; all four required elements present: (1) operator decision dated 2026-09-22, `moai todo` keeps its name (lines 3-4); (2) 8-tree GTD family closed as investigation-only (line 8); (3) t855 zero-work-commits note (line 9); (4) t1084 named as disposal owner (line 10). No SPEC status transitions performed. Baseline: this run, this tree, HEAD cb74ebd58.
- AC-008 (post-M1 leg) PASS — `$ go test ./internal/cli/ -run TestGTD -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 4.934s` (this run, this tree, HEAD cb74ebd58).
- M2 gate re-measurement at M1 close: `git ls-tree develop --name-only internal/factorymsg/ | wc -l` → `0` (factorymsg NOT in develop); t1074 worktree files unchanged (mtime 17:17). Gate CLOSED — M2-M4 halt per REQ-003; resume on gate re-measurement after t1074 lands.

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~10 files (factory.go, factory_slots.go, i18n, tests, doc twins); domains 3 (go source, hook i18n, rule-doc twins + templates); file mix go+md; concurrency benefit LOW (coding-heavy rename).
- Mode evaluation: direct — no (multi-file code change); fanout — no (coding-heavy, Anthropic caveat); sweep — no (semantic rename across coupled surfaces, not uniform-mechanical ≥30 files); agent-team — not requested.
- Decision: **serial** (one manager-develop spawn per milestone batch: M2-M3 inventory+rename, then M4 decision).
- Justification: coding-heavy work on interdependent surfaces (token, help text, i18n 4-locale lockstep, tests, doc twins) — sequential edits in one tree by one writer; the t1074 gate already serializes the whole rename. Boundary case: none.

### M2 gate re-measurement (2026-09-22, post-18:18 — this session, branch WT-worker-rename, HEAD eace7849f)

Both gate observations re-measured fresh from this session against the develop ref (attributed baseline: `git rev-parse --short develop` → **a08972a28**, measured 2026-09-22 after 18:18):

1. develop SHA (the tree the gate read):
   - `$ git rev-parse --short develop` → `a08972a28`
2. factorymsg presence in develop:
   - `$ git ls-tree develop --name-only internal/factorymsg/ | wc -l` → `0`
3. t1074 SPEC status read from develop:
   - `$ git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md` → `fatal: path '.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md' does not exist in 'develop'` (exit 128 — path absent from the develop tree)

**Conclusion**: M2 gate remains **CLOSED** (factorymsg absent from develop AND SPEC-FACTORY-MIXED-HOOK-001 SPEC absent from develop). M2-M4 **halt per REQ-003**. Resume condition: t1074 (SPEC-FACTORY-MIXED-HOOK-001) lands in the local `develop` ref — observable as `git ls-tree develop --name-only internal/factorymsg/` returning non-empty AND `git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md` succeeding with frontmatter `status` ∈ {implemented, completed}.

### M2 — gate OPEN + old-token inventory (2026-09-23, branch WT-worker-rename, HEAD 861510fb6)

Kickoff: operator Implementation Kickoff Approval relayed by the lead 2026-09-23. All commands below ran in this session against this tree; outputs verbatim.

**Step 1 — gate (REQ-002 / AC-002)**

```
$ git rev-parse --short HEAD                                   → 861510fb6
$ git rev-parse --short develop                                → 5d2d2b780   (first read; later reads: 08113ff0f — develop is moving under concurrent lane merges)
$ git rev-list --count --left-right develop...HEAD             → 3	0         (first read; later 18	0)
$ git ls-tree develop --name-only internal/factorymsg/ | wc -l → 4
$ git ls-tree develop --name-only internal/factorymsg/
internal/factorymsg/launch_pending_rollback_test.go
internal/factorymsg/roster_test.go
internal/factorymsg/store.go
internal/factorymsg/store_test.go
$ git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md | grep '^status:' → status: completed
```

Gate **OPEN** (factorymsg present on develop AND SPEC-FACTORY-MIXED-HOOK-001 `completed`). The t1074 merge `861510fb6` is HEAD itself. develop drift since kickoff (`861510fb6..08113ff0f`, 18 commits: t1086/t1087/t1090/t1091) touches `internal/web/*_test.go`, `internal/mission/governance_receipt_jev_test.go`, `internal/harness/rosterguard/registry.go`, `.claude/rules/moai/workflow/worktree-integration.md`, `.claude/rules/local/*`, and other SPEC/report dirs — `git diff --name-only HEAD develop` shares **zero** paths with this SPEC's edit set. Absorption is deferred to the integration window (lane protocol), not done here.

**Step 2 — inventory (REQ-007 input)**. Raw grep outputs: `.moai/reports/t1085/run-verdict.md` § M2 inventory. Per surface class:

| # | Surface class | Files (count of `lane-` / agent-shape¹ lines) | Disposition input |
|---|---|---|---|
| 1 | Factory CLI production (help, usage error, parse, desugar) | cc.go 10/4 · glm.go 9/0 · factory.go 18/11 · codex_factory.go 2/5 · kanban.go 5/0 | user-facing → rename to worker; old forms parse as aliases |
| 2 | Label vocabulary (kanban) | bootstrap.go 8/7 · factory_slots.go 2/0 · record.go 2/0 · role.go 1/0 | label producer → `worker-<n>`; parser accepts legacy `lane-<n>`/`agent-<n>` |
| 3 | Hook factory surfaces + i18n | session_start_factory.go 4/0 · session_start_factory_i18n.go 8/8 (4 locales × 2 fields) · session_start_kanban_i18n.go 4/0 (`nameChoices`, 4 locales) · session_start_record.go / factory_messages.go 0/0 (consume the label via the kanban parsers) | 4-locale lockstep rename of notation |
| 4 | **factorymsg (persisted, per-run broker.db)** | store.go 0/2 — `if p.Slot == "agent"` allocates `fmt.Sprintf("agent-%d", n)` (store.go:320-322); `peers.slot` is an opaque TEXT key; production callers (hook `registerFactoryHookPeer`, cli `registerFactoryLaunchPending`) pass the already-claimed env label (`MOAI_FACTORY_WORKER`), never the bare sentinel — only tests (store_test.go:459, factory_mixed_test.go:116) reach it | **schema unchanged**; rename changes only newly-written values. Compat: allocator treats an existing `agent-<n>`/`lane-<n>`/`worker-<n>` row as taking number n; old rows stay readable (opaque strings) |
| 5 | **factory.db `workers.label` (persisted, project-scoped)** | factory_slots.go `ClaimFactoryWorkerName` / `FactoryFreeSlots`; rows hold `lane-<n>` / `agent-<n>` from pre-rename launchers, reaped by pid liveness | schema unchanged; compat: a live legacy-labelled row blocks its number for the canonical `worker-<n>` claim |
| 6 | **kanban session record (persisted JSON)** | `Record.Role` value `"lane"` (`kanban.RoleLane`, role.go:42) + `Record.Lane` int, read by web | **KEEP** — internal persisted role key, not user notation; renaming would change on-disk record format |
| 7 | web console | factory_lanes.go 2 (`SplitFactoryLaneLabel` ×2) | consumes the kanban parser; widening it to worker+legacy keeps legacy rows visible |
| 8 | Tests (six named) | factory_test.go 86 lane-/2 · codex_factory_test.go 2/2 · goal_mission_test.go 13 (`lane-10` owner label) · gtd_compat_test.go 1 (`--lane lane-10`) · goal_blocked_question_regression_test.go 1 + doctor_jev_test.go 1 (prose "lane-question routing") | rename notation; prose compounds kept |
| 9 | Tests (other factory) | session_start_factory_test.go 27 · session_start_record_test.go 7 · factory_slots_test.go 15 · factory_label_test.go 4 · web factory_lanes_test.go 12 · factory_lane_section_test.go 9 · factorymsg/*_test + hook/cli factorymsg tests (`agent-1`, `agent-2` slot fixtures) | update where they assert produced labels; legacy-input cases retained as alias coverage |
| 10 | Rule-doc twins | local kanban-dispatch.md 2 (L211 "lane-local" adj., L266 notation) · detail 3 (L32/L182 notation, L176 "lane-local" adj.); template mirrors identical (sha `ea596b32d163` / `6607f4238a55` pairwise) | rename notation lines; keep "lane-local" (prose adjective, per-line) |
| 11 | Out of this SPEC's write scope (public/user docs, agents) | docs-site 4 locales × (factory-mode 8 + launchers 2 + kanban-mode 2) = 48 · README ×4 = 12 · CHANGELOG 2 · manager-lead.md ×2 + `.codex` toml 1 — all advertise `-f lane-<n>` / `-f agent` | sync-phase / separate card; **published docs teach the old forms → strong keep-alias evidence** |
| 12 | Unrelated homographs | integration-lock / slot-lease tests using `lane-2` as an arbitrary holder name (kanban 26+9+7+6+5+3+2, cli 6+6+5+7+4…), `agent-memory`, `--agent-name`, `agent-model` | not factory labels — untouched |
| 13 | Operator-side scripts outside the repo | not measurable from this tree | informs keep-alias; out of write reach |

¹ agent-shape = lines matching `-f agent` · `agent-<` · `agent-%d` · `agent lane` · `factoryAgentRole` · `Slot == "agent"`.

`-f agent` repo-wide: 45 lines (8 i18n, 6 factory.go, 3 cc.go, 2 codex_factory.go, 1 bootstrap.go, tests 7, SPEC dirs 18). Doc/README/CHANGELOG/agent sweep (`-f lane` · `--name lane` · `lane-<` · `-f agent`): 65 lines across 20 files.

Persisted-format verdict (constraint 2): **no blocker** — neither SQLite schema nor the record JSON shape changes; only newly-written label values change, and every reader of those values gains a legacy-accepting parse (compat read path + tests in M3).

### M3 — vocabulary rename (2026-09-23, tree = M2 commit eb105cca2 + M3 working changes)

**Design (one numbering, canonical label, legacy read).** Every label the product now PRODUCES is `worker-<n>` (`kanban.FactoryLaneLabel`, prefix constant `factoryLaneRole = "worker"`). The two pre-rename shapes — `lane-<n>` (numbered form) and `agent-<n>` (role-token form) — are READ everywhere a label is parsed, and share the one worker numbering: a live claim in any shape occupies its number. Helpers: `IsLegacyFactoryLabel`, `CanonicalFactoryLabel`, `NextFactoryWorkerNumber` (replaces `NextFactoryAgentNumber`; highest live claim across all shapes + 1). Behaviour change to note: the former separate `agent-<n>` and `lane-<n>` sequences are now ONE sequence (`agent-3` and `lane-3` can no longer coexist under the new binary; `TestNextFactoryWorkerNumber` pins it).

**M3a code** — `internal/kanban/{bootstrap,factory_slots,record,role}.go`; `internal/cli/{factory,codex_factory,codex_launcher,cc,glm,kanban}.go`; `internal/factorymsg/store.go` (bare `worker` sentinel + legacy bare `agent` → `worker-<n>`; a number is taken when a `worker-`/`agent-`/`lane-<n>` slot row exists). `ClaimFactoryWorkerName` canonicalizes the requested label and counts live legacy rows as taken; `FactoryFreeSlots` likewise. `resolveFactoryWorkerName` is the single deprecation-hint site (legacy spelling → `factory: <hint>; launching as worker-<n>` on stderr). `kanban.RoleLane` keeps its persisted value `"lane"` (record format unchanged; comment says why).

**M3b i18n (4-locale lockstep)** — `session_start_factory_i18n.go` `leadManual` + `entryGuide` in en/ko/ja/zh: `worker-1..worker-%d`, `` `moai %[2]s -f worker` `` taking the next free `worker-<n>`, `` `moai %[2]s -f worker-<n>` `` to pin a number. `session_start_kanban_i18n.go` `nameChoices` (4 locales): `` `lane-N` `` → `` `worker-N` ``. Prose word "lane"/"레인"/"レーン"/"泳道" kept — it names the slot; only notation changed (per-field disposition below).

**M3c tests** — updated to worker vocabulary: `factory_test.go`, `codex_factory_test.go`, `codex_factory_helper_test.go`, `factory_mixed_test.go`, `goal_mission_test.go` (`lane-10`→`worker-10`), `gtd_compat_test.go` (`--lane lane-10`→`worker-10`), `factory_operational_live_test.go`, `kanban/factory_label_test.go`, `hook/session_start_factory_test.go`, `session_start_factory_provider_test.go`, `lane_spawn_authority_test.go`, `session_start_kanban_i18n_test.go`. `goal_blocked_question_regression_test.go` / `doctor_jev_test.go` hits are the prose compound "lane-question routing" (comment text, not a label) → kept. Legacy forms retained as explicit alias coverage (never deleted): `-f agent` / `-f=agent` / `-f lane-2` parse cases, codex `-f agent` / `-f=lane-3` / `-f agent-2`, `-f lane-5` end-to-end runCC case, the legacy registry/broker/env rows in the new tests.

**M3d doc twins (judged per file)** — all four carried identical text at M2 (pairwise sha `ea596b32d163` / `6607f4238a55`); each notation line was judged and rewritten, each "lane-local" (hyphenated adjective, L211 / detail L176) judged prose and kept. After: pairs still byte-identical (`3bf1e8efab8d` / `2bfebca1d4e5`) because the judged edits coincide, not by copy. Template text carries no SPEC id / card id / date / SHA. Stale-but-out-of-scope residue recorded, not fixed: both twins still say `moai cc -f <N>` (the numeric count form was retired earlier).

**New RED→GREEN tests** — `kanban/factory_worker_label_test.go` (6), `factorymsg/worker_slot_test.go` (2), `cli/factory_worker_naming_test.go` (6), `hook/session_start_factory_worker_test.go` (3). RED evidence (verbatim, pre-GREEN):

```
# kanban — stubbed new symbols, HEAD eb105cca2 + test file only
--- FAIL: TestFactoryLaneLabelProducesWorkerNotation (0.00s)
    factory_worker_label_test.go:14: FactoryLaneLabel(3) = "lane-3", want "worker-3"
--- FAIL: TestNextFactoryWorkerNumberSpansLegacyShapes (0.00s)
    factory_worker_label_test.go:81: NextFactoryWorkerNumber = 0, want 5
--- FAIL: TestSplitFactoryLaneLabelAcceptsCanonicalAndLegacyLane (0.00s)
    factory_worker_label_test.go:27: SplitFactoryLaneLabel("worker-1") = (0, false), want (1, true)
--- FAIL: TestFactoryFreeSlotsHonoursLegacyLiveClaims (0.36s)
    factory_worker_label_test.go:144: FactoryFreeSlots = [1 2 3 4], want [1 4]
# factorymsg
--- FAIL: TestRegisterPeerAllocatesWorkerSlots (0.05s)
    worker_slot_test.go:23: sentinel "worker" allocated "worker", want "worker-1"
    worker_slot_test.go:23: sentinel "agent" allocated "agent-1", want "worker-2"
--- FAIL: TestRegisterPeerLegacySlotRowsStayReadableAndBlockTheirNumber (0.04s)
    worker_slot_test.go:47: worker allocated "worker" beside a legacy agent-1 row, want worker-2
# cli — the three existing-symbol tests run on the baseline tree 861510fb6
--- FAIL: TestResolveFactoryWorkerNameCanonicalizesLegacyWithHint (0.32s)
    zz_red_probe_test.go:45: resolve lane-4 = ("lane-4", <nil>), want worker-4
--- FAIL: TestFactoryFlagUsageErrorAdvertisesWorkerForms (0.00s)
--- FAIL: TestLauncherHelpAdvertisesWorkerVocabulary (0.00s)
    zz_red_probe_test.go:74: cc help still advertises "-f agent"
# hook — 44 assertion lines across the 4 locales, e.g.
--- FAIL: TestKanbanNameChoicesUseWorkerNotation (0.00s)
--- FAIL: TestFactoryGuideTeachesWorkerFormsInEveryLocale (0.00s)
    session_start_factory_worker_test.go:33: en leadManual still teaches "-f agent":
```

The remaining three cli tests (`TestParseFactoryFlagWorkerVocabulary`, `TestParseLauncherEntryDesugarsWorkerLabel`, `TestStripCodexFactoryFlagWorkerVocabulary`) reference new fields/signatures; their RED was a compile failure (`p.WorkerRole undefined`, `undefined: factoryWorkerRoleToken`), recorded as such — not an assertion-level red.

### M4 — compatibility decision (REQ-007 / REQ-008), from the M2 inventory

| Old form | Decision | Measured basis (M2 inventory row) | Behaviour now |
|---|---|---|---|
| `-f agent` | **KEEP-ALIAS** | row 11: published docs teach it (docs-site 4 locales, README ×4, CHANGELOG, `manager-lead.md` ×2 + emitted `.toml`); row 13: operator scripts unmeasurable; row 1: it was the t1074 join token on every door | parses exactly like `-f worker`; desugars to the legacy `agent-<n>` label so the claim canonicalizes it to `worker-<n>` and prints `` `-f agent` (and the agent-<n> label) is a deprecated spelling — use `-f worker` `` |
| `-f lane-<n>` | **KEEP-ALIAS** | row 11: docs-site factory-mode/launchers pages ×4 locales teach `-f lane-<n>`; row 1: the lead notice printed `moai cc -f lane-<i>` launch lines until this change, so live operator muscle memory and pasted lines exist | parses as `-f worker-<n>`; the typed label is canonicalized at the claim with hint `` `-f lane-<n>` / `--name lane-<n>` is a deprecated spelling — use `-f worker-<n>` `` |
| `--name lane-<n>` (with `-f`/`-k`) | **KEEP-ALIAS** | row 1 (`-k N --name lane-<i>` documented in cc/glm help until this change) + row 5 (registry rows under `lane-<n>` from running launchers) | recognised as a worker label (factory branch selected); claimed as `worker-<n>`; same hint; `replaceNamedLabel` carries the canonical label into the backend argv |

No removal is proposed in this SPEC. Argument recorded for the lead (not acted on): the aliases can be retired once the out-of-scope public surfaces (row 11) stop teaching them and one release has shipped the hint — the hint text is the retirement channel. Commit-order check for AC-007: no commit on this branch removes an old form; the M2 evidence commit `eb105cca2` precedes every rename commit.

**AC-010 exception list** (every residual hit of `-f agent` in `internal/` and of `lane-` in `internal/cli/factory.go` + `internal/hook/session_start_factory_i18n.go` is one of these): (a) legacy-alias doc comments in `factory.go` L70/L84-85/L94-96/L218/L231-233/L324 and `codex_factory.go` L20-21, `kanban/bootstrap.go` legacy-shape comments; (b) the two deprecation-hint strings `factory.go` L471/L473 (they must name the old spelling to retire it); (c) keep-alias test cases in `factory_test.go`, `codex_factory_test.go`, `factory_worker_naming_test.go`, `session_start_factory_worker_test.go` (the latter two assert ABSENCE of the old forms from taught surfaces). `session_start_factory_i18n.go`: **0** `lane-` hits.

### M3/M4 verification (this run; M3 commit f8c472d2b; env scrubbed of the session's own MOAI_FACTORY_*/MOAI_KANBAN_* per the kanban-dispatch env-isolated form — the agent session itself runs as `MOAI_FACTORY_WORKER=agent-35`, which otherwise leaks into launcher tests)

```
$ go vet ./internal/cli/... ./internal/kanban/... ./internal/hook/... ./internal/factorymsg/...   → exit 0 (no output)
$ golangci-lint run ./internal/kanban/... ./internal/hook/... ./internal/factorymsg/...           → 16 issues: errcheck: 16
  baseline tree 861510fb6 (git archive, same command)                                            → 16 issues: errcheck: 16   (same files; 0 new)
$ golangci-lint run ./internal/cli/...                                                            → 25 issues: errcheck: 25
  baseline tree 861510fb6                                                                          → 25 issues: errcheck: 25   (identical per-file distribution; factory.go 1 = the pre-existing db.Close in recordFactoryRunStart; 0 new)
$ go test ./internal/kanban/... ./internal/factorymsg/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	175.220s
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	2.787s
$ go test ./internal/hook/ ./internal/factorymsg/ -count=1
--- FAIL: TestFactoryHookBenchmarkBudget (0.00s)                 [inherited: fails identically on 861510fb6 — "MOAI_FACTORY_BENCH=1 is required"]
--- FAIL: TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift     [load-timing: 295ms vs 50ms budget at load ~40; re-run alone → ok 0.923s]
FAIL	github.com/modu-ai/moai-adk/internal/hook	217.578s
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	3.052s
$ go test -timeout 25m ./internal/cli/... -count=1       (compiled ~09:07, before the last three test-file edits below)
--- FAIL: TestGLM_FactoryWorkerEntry (0.19s)             [inherited: 861510fb6 fails identically — "runGLM(-f lane-3): NO_ACTIVE_FACTORY"]
--- FAIL: TestSessionPIDStamp_NotSetFromHooks (0.06s)    [inherited: 861510fb6 fails identically — flags internal/hook/factory_messages_test.go]
--- FAIL: TestACFB019_HelpDocumentsCompanionEntry (0.00s) [MINE: help rewrap split "bumped to the next free number"; fixed in cc.go/glm.go, 861510fb6 passes it]
FAIL	github.com/modu-ai/moai-adk/internal/cli	1261.375s   (17 sub-packages ok; not a timeout)
$ go test ./internal/cli/ -run 'TestGLM_FactoryWorkerEntry|TestSessionPIDStamp_NotSetFromHooks|TestACFB019_HelpDocumentsCompanionEntry|...|Help' -count=1   (after the fix)
--- FAIL: TestGLM_FactoryWorkerEntry     [inherited, unchanged]
--- FAIL: TestSessionPIDStamp_NotSetFromHooks   [inherited, unchanged]
(TestACFB019 and every other Help test pass)
$ go test ./internal/cli/ -run 'TestGTD|GoalMission|Mission|Operational|WorkerVocabulary|WorkerNaming|CanonicalizesLegacy|DesugarsWorker|UsageErrorAdvertises|LauncherHelp' -count=1 -v   → 43 PASS, 2 SKIP (live operational tests, env-gated)
$ go test ./internal/cli/ -run TestGTD -count=1 -v | grep -c '^--- PASS'   → 8
$ go test ./internal/cli/ -run TestGTD -count=1                             → ok  	github.com/modu-ai/moai-adk/internal/cli	4.738s
$ go test ./internal/template/ -count=1 -v | grep '^--- FAIL'   → --- FAIL: TestRuleTemplateMirrorDrift  [inherited: worktree-integration.md drift, repaired on develop by t1086 b64726ebb; the kanban-dispatch twins are not flagged]
$ go build ./... && GOOS=windows GOARCH=amd64 go build ./...     → build-ok / win-build-ok
$ make build   (after the template-mirror edits)                 → exit 0; tail: "catalog.yaml updated successfully (13408 bytes)" + go build … -o bin/moai ./cmd/moai
$ make embed-check                                               → Pass 1  Warn 0  Fail 0
embedded-FS probe (temporary test, removed): kanban-dispatch.md worker-1..worker-N=1 lane-1..lane-N=0; kanban-dispatch-detail.md worker-1..worker-N=2 lane-1..lane-N=0 → PASS
```

Inherited-failure attribution method: the baseline tree `861510fb6` was exported with `git archive` into a scratch directory and the same test selectors were run there with the same scrubbed environment.

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready (M1-M4 complete)
- run_complete_at: 2026-09-23
- run_commit_sha: f8c472d2b (M3 code commit); M2 evidence eb105cca2; this M4 record commit on top is docs-only (`.moai/specs/SPEC-FACTORY-WORKER-NAMING-001/progress.md`, `spec.md` `updated:`)
- ac_pass_count: 10 / ac_fail_count: 0 (AC-009 PASS-WITH-DEBT: two inherited `internal/cli` failures and one inherited `internal/hook` failure, all reproduced on the baseline tree 861510fb6)
- new_warnings_or_lints_introduced: 0 (golangci-lint 16 → 16 and 25 → 25 against the baseline tree; go vet exit 0)
- cross_platform_build: darwin `go build ./...` ok; `GOOS=windows GOARCH=amd64 go build ./...` ok
- total_run_phase_files: 34 in the M3 commit (30 modified + 4 new test files)
- m1_to_mN_commit_strategy: M1 (eace7849f + close) → M2 gate+inventory evidence commit eb105cca2 (precedes every rename edit) → M3 rename+alias commit f8c472d2b → M4 decision/evidence record commit (this)
- push: none (lane protocol — the lead batch-pushes develop)
- develop drift: branch is behind local develop (861510fb6 → 08113ff0f at last read, 18 commits, zero shared paths); absorb at the integration window

### AC status matrix (M1-M4)

| AC | Status | Verification | Actual output (verbatim / pointer) |
|----|--------|--------------|------------------------------------|
| AC-001 | PASS | closure note read (M1) | see M1 row history above — unchanged this run |
| AC-002 | PASS | `git ls-tree develop --name-only internal/factorymsg/ \| wc -l` + `git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md \| grep '^status:'` | `4` / `status: completed` (§E.2 M2 Step 1) |
| AC-003 | PASS (gate-pass branch) | gate-pass evidence committed before the first M3 edit | M2 evidence commit eb105cca2 is the parent of the M3 commit f8c472d2b (`git log --oneline`) |
| AC-004 | PASS | `grep -n 'factoryWorkerRoleToken = \|factoryLegacyAgentRoleToken = ' internal/cli/factory.go`; `grep -c -e '-f agent' -e 'lane-' internal/cli/cc.go internal/cli/glm.go`; `TestLauncherHelpAdvertisesWorkerVocabulary`, `TestFactoryFlagUsageErrorAdvertisesWorkerForms` | `58: factoryWorkerRoleToken = "worker"` / `63: factoryLegacyAgentRoleToken = "agent"` (alias, M4 list); cc.go `0`, glm.go `0`; both tests PASS (RED on 861510fb6 recorded in §E.2 M3) |
| AC-005 | PASS | `grep -c 'lane-' internal/hook/session_start_factory_i18n.go`; `grep -c 'worker-1\.\.worker-'` same file; `TestFactoryGuideTeachesWorkerFormsInEveryLocale` (en/ko/ja/zh) | `0`; `4` (one per locale block); test PASS |
| AC-006 | PASS | per-file `grep -c 'lane-[0-9N<]'` on the 4 twins; `make build`; `make embed-check`; embedded-FS probe | `0 / 0 / 0 / 0` notation hits (remaining `lane-` = "lane-local" prose, per-line kept); make build exit 0; embed-check Pass 1 Fail 0; embedded twins carry `worker-1..worker-N`, zero `lane-1..lane-N` |
| AC-007 | PASS | M4 decision table vs M2 inventory; commit order | every old form has a KEEP-ALIAS decision citing its inventory row; no commit removes an old form |
| AC-008 | PASS | `go test ./internal/cli/ -run TestGTD -count=1` (post-M4 leg) | `ok  	github.com/modu-ai/moai-adk/internal/cli	4.738s` (8 tests PASS) |
| AC-009 | PASS-WITH-DEBT | kanban/factorymsg/hook/cli package runs (§E.2 M3/M4 verification) | kanban ok 175.220s; factorymsg ok; hook FAIL only on inherited `TestFactoryHookBenchmarkBudget` (+ one load-timing test that passes alone); cli FAIL only on inherited `TestGLM_FactoryWorkerEntry` + `TestSessionPIDStamp_NotSetFromHooks` — all three fail identically on 861510fb6; no test file constructs a `lane-`/`agent-<n>` join label except the M4 keep-alias cases |
| AC-010 | PASS | `grep -rn -- '-f agent' internal/`; `grep -n 'lane-' internal/cli/factory.go internal/hook/session_start_factory_i18n.go` | every hit is on the M4 exception list (§E.2 M4); i18n file 0 |

- debt carried to the lead: (1) `TestGLM_FactoryWorkerEntry` needs an active-run fixture since t1074's `NO_ACTIVE_FACTORY` contract; (2) `TestSessionPIDStamp_NotSetFromHooks` flags `internal/hook/factory_messages_test.go` (t1074); (3) `TestFactoryHookBenchmarkBudget` fails unless `MOAI_FACTORY_BENCH=1`; (4) public docs (docs-site ×4 locales, README ×4, CHANGELOG, `manager-lead.md` + emitted `.toml`) still teach `-f agent` / `-f lane-<n>` — out of this SPEC's write scope; (5) both kanban-dispatch twins still say `moai cc -f <N>` (retired count form, pre-existing).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
