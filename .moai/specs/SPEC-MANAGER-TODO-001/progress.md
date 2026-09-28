# SPEC-MANAGER-TODO-001 — progress.md

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase status: complete (artifacts authored 2026-09-29 by manager-spec, worktree `WT-manager-todo-agent`, base develop `b7ff456b7`).

### Plan-phase evidence (commands run this session + verbatim output)

1. **SPEC ID pre-write self-check** (HARD protocol):

```
$ ID="SPEC-MANAGER-TODO-001" && [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

2. **Reference-sweep baseline** (research.md §B; REQ-MT-005 checklist):

```
$ git grep -n -i "mission.governor" -- . ':(exclude).moai/reports' ':(exclude).moai/specs' | awk -F: '{print $1":"$2}' | wc -l
     262
$ git grep -i "mission.governor" -- . | wc -l
     373
```

Live-surface baseline: 262 hits / 105 unique files (hit count: the sweep piped to `wc -l` = 262; unique-file aggregation: same command piped through `cut -d: -f1 | sort -u | wc -l` = 105 — the iter-1 draft's "95 files" was an under-count, corrected per audit finding D7), enumerated per file:line in research.md §B. The 111-hit remainder under `.moai/reports/**` + `.moai/specs/**` is historical and dispositioned out of scope.

3. **Three-copy confirmation**:

```
$ ls .claude/agents/moai/ | grep mission-governor
mission-governor.md
$ ls internal/template/templates/.codex/agents/moai/ | grep mission-governor
mission-governor.toml
```

(C2 mirror `internal/template/templates/.claude/agents/moai/mission-governor.md` confirmed present in the same sweep.)

4. **Todo state vocabulary + store**:

```
$ grep -n "backlog.db" internal/cli/todo.go | head -1
197:		Long: `Operate the kanban backlog queue at ~/.moai/db/<project-key>/todo/backlog.db.
```

Current states measured: picked / queued / done (`hold` absent) — the pickup predicate is written in positive state vocabulary (REQ-MT-010) so the future `hold` state is skipped without SPEC revision.

5. **Codex read-only role roster** (conflict-report source):

```
$ grep -n "mission-governor" internal/cli/codex_audit_mcp.go | head -1
199:			mcp.WithDescription("Start a role whose permission contract is read-only (plan-auditor, sync-auditor, mission-governor, super-advisor) ...
$ grep -n "mission-governor: read-only" internal/template/agentemit/agents-codex.yaml
83:      mission-governor: read-only
489:    mission-governor: read-only
```

6. **internal/web measured zero**:

```
$ git grep -n -i "mission.governor" -- internal/web/
(no output — zero hits)
```

### Phase-1 SKIP rationale

Cross-reference: plan.md §I DP1. The plan-phase discovery round (Socratic/clarify, workflow phase 1) is satisfied by the approval proxy — the operator-approved card plus the lead's explicit plan-commencement directive dated 2026-09-29 — recorded verbatim in plan.md §I; no AskUserQuestion channel exists inside a lane, so no clarification round was run. The one open decision (DP2, roster disposition) was escalated, not guessed — it was carried as a clarification marker through iteration 1 and is now RESOLVED by the lead (2026-09-29, rename-registered, zero removals; plan.md §I DP2, research.md §C, design D-3). Zero `[NEEDS CLARIFICATION]` markers remain in the artifact set (measured post-repair: grep returns nothing).

### Decision Point 1 disposition

DP1 (plan-phase entry): approval proxy = operator-approved card t1306 + lead's explicit "plan 착수" directive dated 2026-09-29. Recorded as such; run-phase entry still requires the standing Implementation Kickoff gate.

### Audit-ready signal

- plan_complete_at: 2026-09-29T04:42:18+0900
- plan_status: audit-ready

Iteration 2 (delta re-audit): **PASS 0.89** @ `b912842df` — Tier L threshold 0.85 met, zero blocking defects; D1–D11 + D9/DP2 all RESOLVED against the `b912842df` diff (+95/−52, 6 files); regression checks hold (MP-1/2/3/5, AC 22, REQ→AC 18/18, lint 0 error/0 warning/1 INFO reproduced). Report `.moai/reports/t1306/plan-audit-iter2.md` (line 1: `레인 백엔드: glm (glm-5.3-flash)`). Non-blocking notes carried to run phase: D12a-c (wording), D13 (AC-MT-015 lease grep — narrow to acquisition identifiers if the D-11 boundary prose becomes code comments), D14 (plan M3 step 2 C2-mirror mention), D15 (transcription corrected in this close-out).

### Plan-phase artifact set (Tier L)

- `spec.md` — 12-field frontmatter, 18 REQs (GEARS), Out of Scope (5 H3 topics)
- `plan.md` — milestones M1-M5, pre-flight, self-verification, decision points DP1-DP3 (DP2 resolved at repair round)

### Iteration 1 verdict + repair round (2026-09-29)

**Iter-1 verdict**: FAIL 0.80 (Tier L threshold 0.85) + 3 blocking defects; report `.moai/reports/t1306/plan-audit-iter1.md` (measured against `e11110c60`). Repair round authorized by the lane.

**DP2 resolution applied** (lead decision 2026-09-29, option a — rename-registered, never removed): plan.md §I DP2 carries the verbatim decision + the four surfaces (`codex_audit_mcp.go:199`, `agents-codex.yaml:83,489`, catalogue rule row, fingerprint prose); plan.md §B item 1 rewritten; M2 step 3 gate reads resolved; design D-3 rewritten (removal disposition superseded); research.md §C + §G updated; acceptance.md DoD item 3 updated; spec.md REQ-MT-006/017 reworded to zero-removals. Zero `[NEEDS CLARIFICATION]` markers remain:

```
$ grep -rn "NEEDS CLARIFICATION" .moai/specs/SPEC-MANAGER-TODO-001/
(no output — exit 1)
```

**Defect repairs**:

- D1 (blocking): REQ-MT-005 + AC-MT-005 now carry the third allowed disposition **RENAME-at-next-regen** with a mechanical codemaps arm (count equals baseline 8, disposition literal present in the §B table, D-6 conditional-flip arm); research.md §G's "either is consistent" claim corrected.
- D2 (blocking): REQ-MT-003's sweep-refresh clause removed (property owned by REQ-MT-005/AC-MT-005 at M2 closure); AC-MT-003 relabeled to verify the deletion clause it claims.
- D3 (blocking): design D-11 added — processing means = `moai-kanban-foreman` pattern (pick → isolated in-session Agent() worker → completion judged on disk evidence read → done; failure path unpicks to `queued` with labelled non-finding; explicit no-factory-run/lease/slot boundary); REQ-MT-007 reworded onto that contract; AC-MT-009 extended with the evidence-absent failure arm and the entry-point arm (D6).
- D4: REQ-MT-009 active prohibition form; REQ-MT-012 `Where`→`When`; REQ-MT-014 capability-gate phrasing.
- D5: AC-MT-001 body checks made mechanical (keyword counts + sub-role heading grep); AC-MT-014 pins the D-8 canonical string; AC-MT-015 converted to counted grep (0 hits, exit 1) + lease-dir snapshot; AC-MT-017 converted to call-site consumption grep + poisoned-value behavior test.
- D6: AC-MT-009 entry-point arm (alias routing check + `--auto` in help output).
- D7: "95 files" corrected to 105 unique files with the aggregation command, re-measured this session before the edit (262 hits / 105 files) — research.md §B + progress.md above.
- D8: plan.md M4 verify row corrected to REQ-MT-014..017 / AC-MT-017..019.
- D10: plan.md §B rewritten (marker removed; acceptance.md citation restated accurately).
- D11: AC-MT-011 non-cache arm added (revived-owner re-measurement between consecutive decisions).

**Post-repair lint**:

```
$ moai spec lint SPEC-MANAGER-TODO-001
INFO  OwnershipTransitionUnmeasured  spec.md:1 — transition "(none)" → "draft" expected owner "manager-spec" but commit e11110c60 has no Authored-By-Agent trailer — ownership transition unmeasured
0 error(s), 0 warning(s)
```

Close-out re-measurement 2026-09-29T04:42:18+0900, exit 0. The pre-repair paste above (`✓ No findings`) was the authoring-time output — the INFO row appeared once the ownership rule could measure the landed transition commit; D15 (iter2) corrected this transcription.

**Repair commit**: `b912842df`

- `acceptance.md` — 22 ACs (AC-MT-001..022), all mechanical; edge cases; quality gates; REQ→AC traceability map (complete, 18/18)
- `research.md` — measured current state, 262-hit baseline enumeration, contract-conflict report (resolved), t1240 seam
- `design.md` — decisions D-1..D-11

### Gate disposition — run entry (2026-09-29T05:26:19+0900)

운영자 승인(리드 전달 2026-09-29): served_model_gate 해제 + GLM 판정 채택.

- Changed key: `workflow.served_model_gate.enabled` → `false` — 본 카드 트리 `.moai/config/sections/workflow.yaml`만 (템플릿 원본은 shipped default-false 유지).
- Adopted verdict: plan-audit iter2 **PASS 0.89** @ `b912842df` (서빙 glm-5.3-flash — 리포트 첫 줄 명시분).
- 기존 거부 영수증 파일: 삭제하지 않고 보존(본 세션은 영수증 파일을 건드린 적 없음).
- Run entry 재개: Plan Audit Gate skip-eligible — ① verdict PASS 0.89 ≥ Tier L 0.85 ② plan-artifact hash unchanged since verdict(close-out `ecb12d147`는 progress.md만 변경 — hash 대상 아님) ③ verdict 측정 트리 `b912842df`. Kickoff: §31 자율(운영자 승인 카드 + 리드 run 지시 2026-09-29).
- t1240 재독(리드 ① 조건, 05:26 측정): 브랜치 tip `7714fcd9e` — M4(런처)·M5(codex relaunch) 추가 착지 확인, `internal/cli/todo.go`는 M3(`ac7b1b87a`) 이후 미변경 → 대폭 재작성 신호 없음, 병렬 진행 확정. develop `a62a05764` 흡수 예정(`b7ff456b7..develop`의 todo.go 델타 0건 관측).

## §E.2 Run-phase Evidence

Run-phase lane: `WT-manager-todo-agent`, card t1306, cycle_type=tdd, serial (progress.md §F). Evidence per milestone: command + verbatim decisive output, measured against this tree in this run.

### Pre-flight (2026-09-29, HEAD `02ce44220`)

- `git branch --show-current` → `WT-manager-todo-agent`; `git rev-parse --short HEAD` → `02ce44220` (develop `a62a05764` absorbed, clean tree).
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `golangci-lint run --timeout=2m` → `0 issues.` (baseline clean).
- `make agents-emit-check` exit 0 on base.
- B2 retired-SPEC pre-scan (`grep -rn 'Retired\|superseded' internal/cli internal/kanban internal/web internal/mission internal/harness`): hits are CG-retirement / model-key-strip / harness-verb prose — no conflict with this SPEC's scope.
- M2-start reference re-sweep: `git grep -n -i "mission.governor" -- . ':(exclude).moai/reports' ':(exclude).moai/specs'` → **262 hits / 105 unique files — identical to the plan-phase baseline** (no decay in the absorb). Codemaps arm: `git grep -c -i mission.governor -- .moai/project/codemaps/` → data-flow.md 1 + docs-truth.md 6 = **7** (plan-time figure 8; delta explained by t1305 codemaps refresh8 regenerating the files in the absorb — the re-measured 7 is the operative baseline per plan §C item 1).

### M1 — Agent rename/repurpose (with AC-MT-001..003)

- C1 `​.claude/agents/moai/manager-todo.md` authored + C2 mirror hand-edited; C1+C2 `mission-governor.md` deleted; C3 regenerated via `make agents-emit`; orphaned `mission-governor.toml` (source gone) removed — deletion, not a content hand-edit; `make agents-emit-check` exit 0 after.
- AC-MT-001 mechanical: `grep -c "^name: manager-todo"` = 1 each; `grep -E "^(model|effort):"` exit 1; tools CSV (`awk` → `CSV_OK`); body keywords `todo-queue`=1 (first pass 0 — description carried the capital-T form; body fixed), `jev`=4, `dispatch`=6, `^## .*[Ss]ub-[Rr]ole`=1; `permissionMode: acceptEdits` (D-1: queue role needs more than plan; read-only discipline moved to the sub-role prompt contract).
- AC-MT-003: `ls` on all three `mission-governor` paths → `No such file or directory` ×3.
- AC-MT-004 verified at the M1+M2 boundary (Go surfaces are M2 scope) — see below.

### M2 — Reference sweep refresh (with AC-MT-004..008)

- Mechanical rename applied across 29 enumerated Go/yaml surfaces (81 occurrences) + rules/mirrors + CLAUDE.md pair + AGENTS.md.tmpl + goal.md pair + catalogue rule pair.
- delegationmap fixture: `mission_governor_undesignated.jsonl` → `manager_todo_undesignated.jsonl` (RENAME chosen over FROZEN: freezing leaves the fixture's role outside the renamed retainedCatalog, and the analyzer tests fail either way — rename+filename+tests together is the only green state). Research §B "RENAME-or-FROZEN" resolved to RENAME.
- `rosterguard`/`axis.go` historical narratives reworded to name the role without asserting the old name as current (t917 line, profile.go creation line).
- `governance_receipt.go` issuer literal renamed (`"mission-governor"` → `"manager-todo"`) with sealer/validator consistency; receipt schema, path convention, and the `--governor-receipt` flag name unchanged (D-9). Research §C's "receipt binds only to schema" claim was imprecise — the issuer literal is code; corrected here.
- **D-6 conditional flip FIRED**: rosterguard's `TestRegisteredSitesMatchTheirDeclaredAxis` failed on the stale codemaps (`docs-truth-catalog … absent from the site: manager-todo`). Per design D-6 the disposition flipped to in-milestone regeneration: docs-truth.md (header hand-update note, drift note, §1 table row 12, class breakdown, file-mapping row) and data-flow.md (receipt-issuer line) updated in place — docs-truth.md's own header instructs hand-updating it at each regeneration; codemaps grep now returns **0** (was 7).
- README 4-locale roster rows pulled forward from M5 (rosterguard readme-* sites failed on the same guard): goal `--auto` prose + roster table row updated in en/ko/ja/zh with section parity; M5 retains the verification duty.
- `TestManagerTodoJudgmentSubRoleBoundary` (rewritten from the former frontmatter read-only test, per D-1) green; rosterguard suite green (17.782s).
- AC-MT-004: `grep -rn "mission-governor" .claude/agents/ CLAUDE.md internal/template/retained_agents.go internal/template/catalog.yaml` → exit 1 (zero hits); `manager-todo` present in each; `go test ./internal/template/... ./internal/harness/rosterguard/... ./internal/harness/delegationmap/...` exit 0.
- Post-sweep live grep: 131 hits remain = docs-site 4-locale 124 (M5 scope) + FROZEN `internal/cli/testdata/codex-rollouts-t1171/**` 12 + HIST (CHANGELOG.md 3, repo-root `reports/moai-dual-harness-*` 2) — all three dispositions recorded in research.md §B.

### M1+M2 commit-granularity decision (plan DP3)

Measured: the interim M1-only state cannot go test-green (rosterguard/delegationmap/mission/template tests read the agent file set, the catalog, and the receipt issuer). Per plan §F M1 risk note and DP3, M1 and M2 land as ONE commit.

### M3 — `/moai:todo --auto` serial mode (+M4 Jev boundary, combined commit — deviation noted)

- RED evidence captured before GREEN: `go vet ./internal/cli/` on the test-first tree → `undefined: autoLiveness / autoRegistryEntry / autoPickTargets / runAutoCycle / autoOptions` (planned API absent — the tests fail because the implementation does not exist).
- GREEN: `internal/cli/todo_auto.go` (pickup D-5 positive vocabulary; two-channel liveness D-4 with per-decision re-measurement; serial D-11 cycle with evidence-read completion; unpick failure path with labelled non-finding; per-card /clear guidance; display-only Jev consultation with degraded labelled non-finding), wired as `--auto` / `--auto-wait` on the todo parent command tree (gtd spelling shares it — one implementation, both entry points).
- `TestTodoAuto*` suite green (10 tests: selection table incl. hold-shaped unknown state, liveness 3-scenario + non-cache arm, no-takeover, serial cycle incl. failure arm, guidance-per-card, no-factory-lease, empty-queue zero-target, Jev poisoned-value + degraded, command-surface entry point).
- Workflow text: gtd.md § `--auto` — the serial batch consumption (both copies) carries the D-8 canonical sentence; kanban-dispatch.md § Entry into the board is an operator act carries the reconciliation clause (both copies).
- Guard reconciliation: `TestTodoAutoDone_CloseSurfaceExclusivity` allowlist extended to the third close surface (`todo_auto.go` = `auto`) with the D-11 rationale recorded in the test; `TestProductionStringLiteralsUseLeaderLaneVocabulary` hits fixed by rewriting the two consultation strings off the legacy role vocabulary.
- FROZEN-fixture seam: `codex_role_contract_test.go` maps the t1171 fixtures' historical session label to the current role name AFTER the fingerprint derive (dual-label `role`/`fixtureRole`) — the frozen fixtures and the frozen expectation table keep their recorded names; the mapping lives in the test. The one `mission-governor` literal this introduces is the historical-label reference the frozen-fixture seam requires, recorded as disposition **HIST** in the §B reconciliation.
- Absorb: local develop moved mid-run (t1326 surface-guard declaration `a59988ad8`, t1300 model-docs sweep) — `TestTodoVerbSurfaceZeroDelta` was failing on the stale base; develop absorbed into the card branch before M5 (merge `a94c273fb`, no conflicts).

### M5 — docs-site 4-locale + verification

- Post-absorb re-sweep: docs-site hits reduced 124 → 69 by the absorbed t1300 model-docs sweep (model-policy.md, profile-matrix.md, faq.md, cli.md hits already gone); README hits already 0 (pulled forward in M2).
- All 69 remaining docs-site hits updated across en/ja/ko/zh: judgment-role prose reframed to the manager-todo sub-role framing (agent-guide table + 2 paragraphs per locale, no-haiku-3tier verdict passages, moai-goal receipt + proposer prose, moai-gtd proposals line, sub-agents catalog enumeration, what-is-moai-adk table, introduction, claude-md-guide); simple hits (roster tables, mermaid tree diagrams, file-tree listings) token-swapped. Residual: **0** (`git grep -i mission.governor -- docs-site README.*` → 0).
- 4-locale parity: all 8 touched pages exist in all 4 locales; heading-count parity verified equal in en/ja/zh for every page. Pre-existing divergence measured at the card base and NOT introduced by this card: ko agent-guide 22 vs 37 headings, ko claude-md-guide 21 vs 46, ja/zh sub-agents 14 vs 23 (identical counts at base `02ce44220`; this card changes prose lines only, never headings) — recorded as a pre-existing baseline Gap, out of this SPEC's scope.
- Hugo build: `hugo --quiet` exit 0, sitemap emitted (warning-free).
- Template neutrality: forbidden-class spot grep over `git diff 02ce44220..HEAD -- internal/template/templates/` added lines → 0 hits (no card ids, no internal dates, no commit SHAs, no macOS-bias paths, no local-instruction references).
- Post-absorb gates: `make build` exit 0, `make agents-emit-check` exit 0, `make commands-emit-check` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0; template/harness/mission/graph suites 21 packages ok.
- Lint final: `golangci-lint run` (4 changed package groups) → `0 issues.` after fixing 19 NEW errcheck findings (unchecked `fmt.Fprint*` in todo_auto.go — blank-assigned per the file family's convention).
- Final cli suite on the merged tree: `go test -timeout 30m -count=1 -cover ./internal/cli/...` → only `TestLocalInstructions_UpdateDoctorPreserveFile` failed (pre-existing doctor Constitution Registry DRIFT baseline; see verdict Gaps). All rename/auto/roster/surface/guidance/liveness tests green.
- Coverage: internal/mission 88.1%, rosterguard 93.8%, delegationmap 87.7%, internal/template 83.8% (pre-existing package baseline; new/changed surfaces measured separately), internal/cli main package 84.0% (pre-existing package baseline). New `todo_auto.go` function coverage from the focused suite: consultation 85.7%, ownerAlive 94.7%, pickup 82.4%, cycle 81.4%, guidance/directive/writer paths 83–100%; the uncovered remainder is the production `lsof`/registry process wrappers, which are seam-covered in behavior tests.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29T06:20:00+0900
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 22
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (worktree lane; no primary-checkout fetch permitted)
l44_post_push_fetch: not-run (lane does not push; lead batch-pushes develop)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: pass
cross_platform_build.windows: pass
total_run_phase_files: 77
m1_to_m5_commit_strategy: M1+M2 combined (DP3 measured interim red), M3+M4 combined (shared todo_auto.go surface), M5 evidence/docs
```

Notes: AC-MT-016 PASS is conditional on the recorded pre-existing baseline row (`TestLocalInstructions_UpdateDoctorPreserveFile`, Constitution Registry DRIFT predating this card). Evidence file: `.moai/reports/t1306/verdict.md` (machine-local record; decisive lines transcribed there).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

## §F Phase 4 Mode Selection

Inputs: tier L; run-phase scope ≈ 25-40 files (agent defs C1/C2, codex_audit roster + emit, todo CLI + command + skill, reference sweep surfaces, docs-site ~15 pages × 4 locales, README × 4, tests across internal/{cli,template,kanban}); domains ≥ 3 (Go, markdown/docs, yaml/templates); language mix Go + markdown + yaml; concurrency benefit LOW (coding-heavy — Anthropic coding-task parallelism caveat).

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | semantic multi-file implementation |
| serial | **selected** | coding-heavy multi-domain — sequential manager-develop per milestone |
| fanout | no | research-heavy only; this work is coding-heavy |
| sweep | no | not mechanical-uniform; cross-file invariants (emit, roster, catalogue) |

Decision: serial

Justification: coding-heavy work carrying cross-file invariants (agents-emit regeneration, roster/catalogue coupling, template neutrality) sequences safely; one manager-develop spawn drives milestones M1–M5 with the Section A-E delegation template. Parallel lane work (t1240) proceeds in its own tree; the todo.go overlap resolves in the serial merge window per the lead's ① decision.
