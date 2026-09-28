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

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

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
