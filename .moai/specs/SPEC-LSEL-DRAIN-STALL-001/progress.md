# SPEC-LSEL-DRAIN-STALL-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-25
spec_id: SPEC-LSEL-DRAIN-STALL-001
tier: M
version: 0.1.0
artifact_set: [spec.md, plan.md, acceptance.md, progress.md]
baseline_sha: a739d04b4
worktree: .claude/worktrees/t259
branch: WT-lsel-drain-stall
red_now_observations:
  wrapper_absent: "test -f .claude/skills/hns-lsel-curator/session_drain.sh → rc=1 (2026-08-25, worktree a739d04b4)"
  live_offset_gap: "wc -l → 4204행 vs drain-offset.json offset 629 (2026-08-25, primary 체크아웃 live state — RED for AC-LDS-010 predicate)"
  offset_frozen_since: "2026-08-04T05:21:03Z (21일) — clusters.json 같은 날짜"
  sessionstart_wiring_absent: "settings.json SessionStart 항목 = 1(session-attribution); settings.json+settings.local.json hooks에 lsel/backlog grep → 0 matches"
  loop_recipe_false_claim: "grep -c 'The recipe runs drain.sh' lsel-drain-loop.js → 1 (12행) vs 본문 45-49행 console.log뿐"
  dead_anchor: "grep -rn '§28' .claude/skills/hns-lsel-curator/ → 1건 (backlog_check.sh:50); live CLAUDE.local.md(35,355B)에 §28/LSEL 0회"
  backlog_composition_dry_run: "3,563행 delta — noise 3,095(86.9%), singleton 6, 후보 ≈13-15, top tool_failure:Read:UnknownFailure(146) — drain.sh 건재 rc=0"
id_precheck: "SPEC-LSEL-DRAIN-STALL-001 → bash regex → PASS (2026-08-25)"
```

계획 산출 4건 완료(v0.1.0, 2026-08-25). 카드 t259 (Class B). 원인 확정: **트리거 부재** — `/loop` 스케줄이 session-scoped로 소멸 + 레시피 본문이 실행 없는 console.log 묶음(헤더 주장과 불일치). Tier M 확정(AC 12 > Tier S 상한 8; tracked-코드/로컬-적용 이중 검증 표면). Milestone 2개: M1 wrapper+신호+위생, M2 일괄 드레인+검증(mutant guard)+로컬 인도물 적용. plan-audit 대기.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
