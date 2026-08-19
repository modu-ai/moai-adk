# t85 review verdict — factory lead loop

- Reviewer: lead session (operator-confirmed direct verdict; t96 lead-verified precedent)
- Card: t85 · Worktree: `.claude/worktrees/t85` · Branch: `WT-t85` @ `59dd62d02` (base `7588f0529`)
- Delta reviewed: `7588f0529..59dd62d02` (18 files, +852/−158)
- Lens: `--deep` (multi-package logic + operator-visible notice surface)
- SPEC: SPEC-FACTORY-WORKER-FANOUT-001 v1.1.0 (REQ-FF-001 amend reviewed — spec text matches implementation exactly)
- Note: verdict file lives in the lead's release-v311 tree (worktree isolation blocks writing into the lane's tree); the lane receives the verdict by cross-session message.

## Verdict: PASS (1st pass — superseded as integration trigger by the FINAL re-review below)

## 1. Claims reviewed (run-lane report vs this review's direct reads)

| # | Claim | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Commit 59dd62d02, unpushed, on WT-t85 | `git log` + `git show-ref refs/heads/WT-t85` = `59dd62d0279e41611816a08c20762b9407b3cf6e` | PASS |
| 2 | 18 files +852/−158 | `git diff --stat 7588f0529..59dd62d02` — exact match | PASS |
| 3 | Registry cluster move cli→kanban (rename, single implementation) | `factory_slots.go` new (6 exported symbols); cli keeps thin delegates; old private bodies removed — no duplicate implementation survives | PASS |
| 4 | Bare `-f` → default 8 | `factory.go` parseFactoryFlag: no value → `workers = config.DefaultFactoryWorkers`; `defaults.go:376 const DefaultFactoryWorkers = 8`; supplied invalid count → error naming forms + default (pin phrase "requires a worker count of 1 or more" retained, extended) | PASS |
| 5 | `QueuedCount` shared shape | `BacklogStore.QueuedCount()` + `BacklogPathForRoot` + `QueuedBacklogCountForRoot` in `internal/kanban/backlog_store.go`; hook/cli both consume it | PASS |
| 6 | Lead notice loop block, 4 locales | `session_start_factory_i18n.go` — full field set (`leadLoop`/`leadQueueSummary`/`leadFreeSlots`/`leadSlotsNone`) at en:60-72, ko:85-95, ja:108-118, zh:131-140; protocol tokens verbatim across locales | PASS |
| 7 | Staggered spawn + no-model-override [HARD] codified | `leadLoop` text names staggered activation ("started producing output" token) and `ANTHROPIC_DEFAULT_*_MODEL` no-override — matches cache-aware-execution directive 2 and the GLM slot-mapping constraint | PASS |
| 8 | glm_task factory guard | `glm_task.go:159+`: explicit `model` param + `MOAI_FACTORY_WORKERS` set → override ignored in favor of SSOT resolve, `result.Note` states so. Correct branch shape (empty model already routes SSOT) | PASS |
| 9 | RED→GREEN | RED captures in run evidence (loop-token test FAIL + undefined-symbol build FAIL) + `TestFactoryLeadNoticeCarriesLoopDiscipline` exists @ `session_start_factory_test.go:159` | PASS |
| 10 | t94/t96 surfaces excluded | delta file list (18) contains no t94 rules files, no t96 skill/foreman files | PASS |
| 11 | Root threading | `session_start.go`: `factoryRoot = input.ProjectDir || input.CWD`, fail-open empty | PASS |
| 12 | Cross-platform | `factory_alive_{unix,windows}.go` moved as pair; lane evidence: `GOOS=windows GOARCH=amd64` build+vet exit 0 (not re-run by reviewer — see Gaps) | PASS-attributed |

## 2. Evidence (this review's commands)

All reads against the shared object store from the release-v311 worktree (worktree lock on `.claude/worktrees/t85` prevents session entry; object-level reads are equivalent for diff/grep/show):

- `git log --oneline -3 59dd62d02` · `git show-ref --verify refs/heads/WT-t85`
- `git diff --stat / --name-only 7588f0529..59dd62d02`
- `git diff 7588f0529..59dd62d02 -- <kanban core / cli / hook / config paths>` (full read of factory_slots.go, backlog_store.go additions, factory.go, glm_task.go guard, defaults.go, cc.go/glm.go help, session_start_factory.go, session_start_kanban.go, session_start.go, SPEC spec.md amend)
- `git grep` @59dd62d02: old-symbol audit (internal/cli), help "8" pin locations, `DefaultFactoryWorkers` test pin
- `git show 59dd62d02:internal/hook/session_start_factory_i18n.go` (locale field audit)
- Read: `.moai/reports/t94/t85-run-evidence.md` (5-section, RED captures + GREEN outputs + orchestrator re-run section, gaps declared)

## 3. Baseline attribution

- Code reads: tree @ `59dd62d02` (WT-t85 ref verified). SPEC amend read at same commit.
- Test-execution outputs: NOT re-executed by this reviewer — consumed from the implementing lane's attributed evidence (commands + verbatim outputs, same tree). Rationale: (a) WT-t85 is session-locked and worktree isolation blocks cross-tree execution; (b) lane-local load discipline — full-suite judgment belongs to CI on the integrated release head. The reviewer re-verified every statically checkable claim directly (table §1).

## 4. Gaps

- Test/lint/vet outputs are lane-attributed, not reviewer-re-executed (see §3). CI on the integrated release branch is the final full-suite judge.
- glm_task env-inheritance final runtime link (claude→mcp-server stdio child) verified in code chain + production notice corroboration only — carried from lane gap.
- `factory_slots_test.go` (162 lines) coverage accepted from lane's 11-subtest golden PASS; reviewer read symbols, not every subtest body.

## 5. Residual risks

- **PID reuse**: a dead worker's pid reused by an unrelated live process keeps that slot blocked until the process exits. Pre-existing v1 registry semantics (kill -0 probe), unchanged by t85; prune mitigates the accumulate case, not the reuse case.
- **Duplicate liveness probes**: `FactoryFreeSlots` probes each taken slot twice (prune pass + per-slot re-check) — ≤2N `kill(0)` calls, negligible at N≤8; noted, not a defect.
- **Help "8" hardcode** (`cc.go:91`, `glm.go:106` example lines): intentional (golden %-escape constraint); drift risk mitigated by the constant pin test (`factory_test.go:430` asserts `DefaultFactoryWorkers == 8`). If the default ever changes, both help strings need a manual touch.
- **`cc.go` -k example still shows 4-phase "plan->run->verify->sync"** — pre-existing branch content, out of t85 scope; t97 (D1) is the card that cleans this phrasing.
- Factory notice data lines are point-in-time at session start (queue count / free slots go stale until the lead re-polls) — by design; the loop block instructs polling.

---

# FINAL — rework re-review (the integration trigger)

- Rework commit: `a34494665` (`refactor(factory): retire -f entry flag — unify factory entry on -k N`)
- Delta reviewed: `9794f293d..a34494665` (13 files, +629/−566). Base `9794f293d` verified as merge of origin/release/v3.1.1 into WT-t85; prior commits `59dd62d02`/`4c5994237` verified intact in ancestry (no amend). WT-t85 ref = `a34494665be7…`.
- Scope: operator §6 final correction — -f retirement, -k N unification, card-class routing codified (4 locales), t96 overlap removal, t94 file update, SPEC v1.2.0.

## Verdict: PASS — FINAL (t85 + t94 may integrate; t97 collision resolved per the recorded principle)

## F.1 Claims reviewed (rework report vs this review's direct reads)

| # | Claim | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Commit/stat/base/ancestry | `git log`, `merge-base --is-ancestor` ×2, `diff --stat`, `show-ref` — all exact | PASS |
| 2 | -f fully retired | `parseFactoryFlag` + flag constants + `rejectConflictingModes` deleted; `rejectRetiredFactoryFlag` new — exact-token match (`-f`, `--factory`, `-f=`, `--factory=`), `--` stops the scan, error names the new entry | PASS |
| 3 | -k N unified parse | `kanbanEntryParse` read in full: numeric discriminator (Atoi; n<1 → usage error naming every shape), `=`-joined form for count only, worker-shape bare-`-k` post-pass defaults to `config.DefaultFactoryWorkers` (count-less factory LEAD provably nonexistent — bare -k is the kanban lead), `--` discipline coherent with `parseFactoryWorkerLabel` | PASS |
| 4 | 16-case truth table test | `TestParseKanbanFlagUnifiedEntry` present (`factory_test.go`) | PASS |
| 5 | Card-class routing codified (4 locales) | i18n diff read: `leadClasses` (A/B wholesale vs C serial 3-stage) at en/ko/ja read in full; field-count grep = 16 (4 fields × 4 locales, zh covered) | PASS |
| 6 | t96 dedup | polling sentence → foreman handoff line (matches t96's reserved seam: queue polled by operator/foreman, factory routes PICKED cards); `leadQueueSummary` removed (its consumer was the polling directive); free-slot line KEPT (stagger consumes it) | PASS |
| 7 | Dead helpers deleted / live surfaces kept | `factoryFreeSlots`/`factoryQueuedCards` gone from internal/cli; `kanban.FactoryFreeSlots`/`QueuedBacklogCountForRoot` still consumed by hook (grep: 3 usages); glm_task guard unchanged | PASS |
| 8 | Stagger fan-out-only + workflow exclusion | `leadStagger` names `CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS` exclusion | PASS |
| 9 | SPEC v1.2.0 | title/GENEALOGY/REQ-FF-001/002/004 + Amendments read — matches implementation exactly (incl. the worker-entry-default-8 rationale) | PASS |
| 10 | t94 file update + mirror parity | §C.4 one-line `-k <N>` update read; local↔template shasum identical (`30a970ae8…` both) | PASS |
| 11 | cg rejection split | `rejectKanbanOnCG` skips FactoryEnabled shapes; `rejectFactoryOnCG` answers factory shapes via the unified parse — order documented | PASS |
| 12 | Golden updates | `TestFactoryGenealogyInHelp` markers RETIRED/`-k <N>`/absence assertion verified in test source | PASS |

## F.2 Gaps

- Rework-round verification outputs are inline-message-only (no persisted rework evidence file) — same shape as round 1; compensated by this review's direct reads of every statically checkable claim. Snapshot `HEAD:a34494665` recorded by the lane.
- Test outputs lane-attributed (build OK; cli `-run 'Kanban|Factory|GLMTask|TestParse'` ok 1.425s; kanban ok 24.541s; hook ok 0.559s; template suite ok 34.460s incl. mirror+neutrality; vet/lint 0; GOOS=windows build 0).

## F.3 Residual risks

- `-k=SPEC-X` (`=`-joined non-numeric) now errors — a stricter surface than the old no-`=`-form kanban entry; documented in REQ-FF-001.
- `-k worker-2` (non-numeric positional) still reads as a kanban SPEC id — pre-existing tolerance class, not a regression.
- t97 integration collision (cc/glm/factory/kanban.go + kanban i18n) — resolve per the t97 verdict principle: flag semantics follow this rework's `-k N` shape; the D1 three-role direction is common to both.
- The 1st-pass residual notes that still apply: PID-reuse slot occupancy (v1 semantics), duplicate liveness probes (negligible). The queue-count staleness note is OBSOLETE — the queue-count line was removed in this rework.
